// Copyright (c) 2016 Pani Networks
// All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package rlog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// A few constants, which are used more like flags
const (
	notATrace     = -1
	noTraceOutput = -1
)

type Level int

func (level Level) String() string {
	return levelStrings[level]
}

func (level Level) Bytes() []byte {
	return levelBytes[level]
}

// The known log levels
const (
	levelNone Level = iota
	levelCrit
	levelErr
	levelWarn
	levelInfo
	levelDebug
	levelTrace
)

// Translation map from level to string representation
var levelStrings = map[Level]string{
	levelTrace: "TRACE",
	levelDebug: "DEBUG",
	levelInfo:  "INFO",
	levelWarn:  "WARN",
	levelErr:   "ERROR",
	levelCrit:  "CRITICAL",
	levelNone:  "NONE",
}

// Translation map from level to string representation
var levelBytes = map[Level][]byte{
	levelTrace: []byte("TRACE"),
	levelDebug: []byte("DEBUG"),
	levelInfo:  []byte("INFO"),
	levelWarn:  []byte("WARN"),
	levelErr:   []byte("ERROR"),
	levelCrit:  []byte("CRITICAL"),
	levelNone:  []byte("NONE"),
}

var levelWidth = 10

// Translation from level string to number.
var levelNumbers = map[string]Level{
	"TRACE":    levelTrace,
	"DEBUG":    levelDebug,
	"INFO":     levelInfo,
	"WARN":     levelWarn,
	"ERROR":    levelErr,
	"CRITICAL": levelCrit,
	"NONE":     levelNone,
}

// filterSpec holds a list of filters. These are applied to the 'caller'
// information of a log message (calling module and file) to see if this
// message should be logged. Different log or trace levels per file can
// therefore be maintained. For log messages this is the log level, for trace
// messages this is going to be the trace level.
type filterSpec struct {
	filters              []filter
	hasAnyFilterAPattern bool
}

// filter holds filename and level to match logs against log messages.
type filter struct {
	Pattern string
	Level   Level
}

// fromString initializes filterSpec from string.
//
// Use the isTraceLevel flag to indicate whether the levels are numeric (for
// trace messages) or are level strings (for log messages).
//
// Format "<filter>,<filter>,[<filter>]..."
//     filter:
//       <pattern=level> | <level>
//     pattern:
//       shell glob to match caller file name
//     level:
//       log or trace level of the logs to enable in matched files.
//
//     Example:
//     - "RLOG_TRACE_LEVEL=3"
//       Just a global trace level of 3 for all files and modules.
//     - "RLOG_TRACE_LEVEL=client.go=1,ip*=5,3"
//       This enables trace level 1 in client.go, level 5 in all files whose
//       names start with 'ip', and level 3 for everyone else.
//     - "RLOG_LOG_LEVEL=DEBUG"
//       Global log level DEBUG for all files and modules.
//     - "RLOG_LOG_LEVEL=client.go=ERROR,INFO,ip*=WARN"
//       ERROR and higher for client.go, WARN or higher for all files whose
//       name starts with 'ip', INFO for everyone else.
func (spec *filterSpec) fromString(s string, isTraceLevels bool, globalLevelDefault Level) {
	var globalLevel Level = globalLevelDefault
	var levelToken string
	var matchToken string

	fields := strings.Split(s, ",")

	spec.hasAnyFilterAPattern = false
	for _, f := range fields {
		var filterLevel Level
		// var err error
		var ok bool

		// Tokens should contain two elements: The filename and the trace
		// level. If there is only one token then we have to assume that this
		// is the 'global' filter (without filename component).
		tokens := strings.Split(f, "=")
		if len(tokens) == 1 {
			// Global level. We'll store this one for the end, since it needs
			// to sit last in the list of filters (during evaluation in gets
			// checked last).
			matchToken = ""
			levelToken = tokens[0]
		} else if len(tokens) == 2 {
			matchToken = tokens[0]
			levelToken = tokens[1]
		} else {
			// Skip anything else that's malformed
			rlogIssue("Malformed log filter expression: '%s'", f)
			continue
		}
		if isTraceLevels {
			// The level token should contain a numeric value
			i, err := strconv.Atoi(levelToken)
			if err != nil {
				if levelToken != "" {
					rlogIssue("Trace level '%s' is not a number.", levelToken)
				}
				continue
			}
			filterLevel = Level(i)
		} else {
			// The level token should contain the name of a log level
			levelToken = strings.ToUpper(levelToken)
			filterLevel, ok = levelNumbers[levelToken]
			if !ok || filterLevel == levelTrace {
				// User not allowed to set trace log levels, so if that or
				// not a known log level then this specification will be
				// ignored.
				if levelToken != "" {
					rlogIssue("Illegal log level '%s'.", levelToken)
				}
				continue
			}

		}

		if matchToken == "" {
			// Global level just remembered for now, not yet added
			globalLevel = filterLevel
		} else {
			spec.filters = append(spec.filters, filter{matchToken, filterLevel})
			if matchToken != "" {
				spec.hasAnyFilterAPattern = true
			}
		}
	}

	// Now add the global level, so that later it will be evaluated last.
	// For trace levels we do something extra: There are possibly many trace
	// messages, but most often trace level debugging is fully disabled. We
	// want to optimize this. Therefore, a globalLevel of -1 (no trace levels)
	// isn't stored in the filter chain. If no other trace filters were defined
	// then this means the filter chain is empty, which can be tested very
	// efficiently in the top-level trace functions for an early exit.
	if !isTraceLevels || globalLevel != noTraceOutput {
		spec.filters = append(spec.filters, filter{"", globalLevel})
	}

	return
}

// matchfilters checks if given filename and trace level are accepted
// by any of the filters
func (spec *filterSpec) matchfilters(filename string, level int) bool {
	// If there are no filters then we don't match anything.
	if len(spec.filters) == 0 {
		return false
	}

	// If at least one filter matches.
	for _, filter := range spec.filters {
		if matched, loggit := filter.match(filename, level); matched {
			return loggit
		}
	}

	return false
}

// match checks if given filename and level are matched by
// this filter. Returns two bools: One to indicate whether a filename match was
// made, and the second to indicate whether the message should be logged
// (matched the level).
func (f filter) match(filename string, level int) (bool, bool) {
	var match bool
	if f.Pattern != "" {
		match, _ = filepath.Match(f.Pattern, filepath.Base(filename))
	} else {
		match = true
	}
	if match {
		return true, level <= int(f.Level)
	}

	return false, false
}

// updateIfNeeded returns a new value for an existing config item. The priority
// flag indicates whether the new value should always override the old value.
// Otherwise, the new value will not be used in case the old value is already
// set.
func updateIfNeeded(oldVal string, newVal string, priority bool) string {
	if priority || oldVal == "" {
		return newVal
	}
	return oldVal
}

// init loads configuration from the environment variables and the
// configuration file when the module is imorted.
func init() {
	var err error

	defaultLogger, err = NewLogger(configFromEnv())
	if err != nil {
		panic(err)
	}
}

// getTimeFormat returns the time format we should use for time stamps in log
// lines, or nothing if "no time logging" has been requested.
func getTimeFormat(config Config) string {
	format := ""
	if !config.LogNoTime {
		// Store the format string for date/time logging. Allowed values are
		// all the constants specified in
		// https://golang.org/src/time/format.go.
		switch strings.ToUpper(config.logTimeFormat) {
		case "ANSIC":
			format = time.ANSIC
		case "UNIXDATE":
			format = time.UnixDate
		case "RUBYDATE":
			format = time.RubyDate
		case "RFC822":
			format = time.RFC822
		case "RFC822Z":
			format = time.RFC822Z
		case "RFC1123":
			format = time.RFC1123
		case "RFC1123Z":
			format = time.RFC1123Z
		case "RFC3339":
			format = time.RFC3339
		case "RFC3339NANO":
			format = time.RFC3339Nano
		case "KITCHEN":
			format = time.Kitchen
		default:
			if config.logTimeFormat != "" {
				format = config.logTimeFormat
			} else {
				format = time.RFC3339
			}
		}
	}
	return format
}

// SetOutput re-wires the log output to a new io.Writer. By default rlog
// logs to os.Stderr, but this function can be used to direct the output
// somewhere else. If output to two destinations was specified via environment
// variables then this will change it back to just one output.
func (l *logger) SetOutput(writer io.Writer) {
	// Use the stored date/time flag settings
	l.logWriterStream = writer
	// l.logWriterStream = log.New(writer, "", 0)
	l.logWriterFile = nil
	if l.currentLogFile != nil {
		l.currentLogFile.Close()
		l.currentLogFileName = ""
	}
}

// SetOutput re-wires the log output to a new io.Writer. By default rlog
// logs to os.Stderr, but this function can be used to direct the output
// somewhere else. If output to two destinations was specified via environment
// variables then this will change it back to just one output.
func SetOutput(writer io.Writer) {
	defaultLogger.SetOutput(writer)
}

// isTrueBoolString tests a string to see if it represents a 'true' value.
// The ParseBool function unfortunately doesn't recognize 'y' or 'yes', which
// is why we added that test here as well.
func isTrueBoolString(str string) bool {
	str = strings.ToUpper(str)
	if str == "Y" || str == "YES" {
		return true
	}
	if isTrue, err := strconv.ParseBool(str); err == nil && isTrue {
		return true
	}
	return false
}

// rlogIssue is used by rlog itself to report issues or problems. This is mostly
// independent of the standard logging settings, since a problem may have
// occurred while trying to establish the standard settings. So, where can rlog
// itself report any problems? For now, we just write those out to stderr.
func rlogIssue(prefix string, a ...interface{}) {
	fmtStr := fmt.Sprintf("rlog - %s\n", prefix)
	fmt.Fprintf(os.Stderr, fmtStr, a...)
}

type logger struct {
	mutex                 sync.Mutex
	logFilterSpec         *filterSpec
	traceFilterSpec       *filterSpec
	formatter             LogFormatter
	additionalInformation string

	settingShowCallerInfo  bool   // whether we log caller info
	settingShowGoroutineID bool   // whether we show goroutine ID in caller info
	settingDateTimeFormat  string // flags for date/time output
	settingConfFile        string // config file name

	settingCheckInterval time.Duration

	logWriterStream     io.Writer   // the first writer to which output is sent
	logWriterFile       *log.Logger // the second writer to which output is sent
	lastConfigFileCheck time.Time   // when did we last check the config file
	currentLogFile      *os.File    // the logfile currently in use
	currentLogFileName  string
	logNoTime           bool
	// name of current log file
}

var defaultLogger *logger

func (l *logger) Formatter() LogFormatter {
	return l.formatter
}

func NewLogger(config Config) (*logger, error) {
	// initialize filters for trace (by default no trace output) and log levels
	// (by default INFO level).
	newTraceFilterSpec := new(filterSpec)
	newTraceFilterSpec.fromString(config.TraceLevel, true, noTraceOutput)

	newLogFilterSpec := new(filterSpec)
	newLogFilterSpec.fromString(config.LogLevel, false, levelInfo)

	var formatter LogFormatter
	switch config.Formatter {
	case "text", "":
		formatter = &TextFormatter{}
	default:
		return nil, fmt.Errorf("formatter '%s' is unknown", config.Formatter)
	}

	l := &logger{
		formatter:       formatter,
		traceFilterSpec: newTraceFilterSpec,
		logFilterSpec:   newLogFilterSpec,
	}

	var checkTime int
	checkTime, err := strconv.Atoi(config.confCheckInterv)
	if err == nil {
		l.settingCheckInterval = time.Duration(checkTime) * time.Second
	} else {
		if config.confCheckInterv != "" {
			rlogIssue("Cannot parse config check interval value '%s'. Using default.",
				config.confCheckInterv)
		}
	}
	l.settingShowCallerInfo = config.ShowCallerInfo
	l.settingShowGoroutineID = config.ShowGoroutineID

	// Evaluate the specified date/time format
	l.settingDateTimeFormat = getTimeFormat(config)

	// By default we log to stderr...
	// Evaluating whether a different log stream should be used.
	// By default (if flag is not set) we want to log date and time.
	// Note that in our log writers we disable date/time loggin, since we will
	// take care of producing this ourselves.
	if config.LogStream == "STDOUT" {
		// l.logWriterStream = log.New(os.Stdout, "", 0)
		l.logWriterStream = os.Stdout
	} else if config.LogStream == "NONE" {
		l.logWriterStream = nil
	} else {
		// l.logWriterStream = log.New(os.Stderr, "", 0)
		l.logWriterStream = os.Stderr
	}

	// ... but if requested we'll also create and/or append to a logfile
	var newLogFile *os.File
	if l.currentLogFileName != config.LogFile { // something changed
		if config.LogFile == "" {
			// no more log output to a file
			l.logWriterFile = nil
		} else {
			// Check if the logfile was changed or was set for the first
			// time. Only then do we need to open/create a new file.
			// We also do this if for some reason we don't have a log writer
			// yet.
			if l.currentLogFileName != config.LogFile || l.logWriterFile == nil {
				newLogFile, err = os.OpenFile(config.LogFile,
					os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err == nil {
					l.logWriterFile = log.New(newLogFile, "", 0)
				} else {
					rlogIssue("Unable to open log file: %s", err)
					return nil, err
				}
			}
		}

		// Close the old logfile, since we are now writing to a new file
		if l.currentLogFileName != "" {
			l.currentLogFile.Close()
			l.currentLogFileName = config.LogFile
			l.currentLogFile = newLogFile
		}
	}

	l.logNoTime = config.LogNoTime

	return l, nil
}

var (
	entryPool = sync.Pool{
		New: func() interface{} {
			return &Entry{}
		},
	}
)

// basicLog is called by all the 'level' log functions.
// It checks what is configured to be included in the log message, decorates it
// accordingly and assembles the entire line. It then uses the standard log
// package to finally output the message.
func (l *logger) BasicLog(logLevel Level, traceLevel int, additionalInformation string, format string, a ...interface{}) {
	entry := entryPool.Get().(*Entry)
	defer func() {
		entry.Reset()
		entryPool.Put(entry)
	}()

	entry.TraceLevel = traceLevel
	entry.Level = logLevel
	entry.Fields = additionalInformation
	f := format
	if f != "" {
		entry.Message = fmt.Sprintf(f, a...)
	} else {
		entry.Message = fmt.Sprint(a...)
	}
	if l.logNoTime {
		entry.Time = ""
	} else {
		entry.Time = time.Now().UTC().Format(l.settingDateTimeFormat)
	}

	if l.settingShowCallerInfo || l.settingShowGoroutineID || (l.logFilterSpec.hasAnyFilterAPattern && len(l.logFilterSpec.filters) > 0) || (l.traceFilterSpec.hasAnyFilterAPattern && len(l.traceFilterSpec.filters) > 0) {
		// Extract information about the caller of the log function, if requested.
		var callingFuncName string
		var moduleAndFileName string
		pc, fullFilePath, line, ok := runtime.Caller(2)
		if ok {
			callingFuncName = runtime.FuncForPC(pc).Name()
			// We only want to print or examine file and package name, so use the
			// last two elements of the full path. The path package deals with
			// different path formats on different systems, so we use that instead
			// of just string-split.
			dirPath, fileName := path.Split(fullFilePath)
			var moduleName string
			if dirPath != "" {
				dirPath = dirPath[:len(dirPath)-1]
				_, moduleName = path.Split(dirPath)
			}
			moduleAndFileName = moduleName + "/" + fileName
		}

		// Perform tests to see if we should log this message.
		var allowLog bool
		if traceLevel == notATrace {
			if l.logFilterSpec.matchfilters(moduleAndFileName, int(logLevel)) {
				allowLog = true
			}
		} else {
			if l.traceFilterSpec.matchfilters(moduleAndFileName, traceLevel) {
				allowLog = true
			}
		}
		if !allowLog {
			return
		}

		if l.settingShowCallerInfo {
			entry.CallerInfo.PID = os.Getpid()
			entry.CallerInfo.FileName = moduleAndFileName
			entry.CallerInfo.Line = line
			entry.CallerInfo.FunctionName = callingFuncName
			if l.settingShowGoroutineID {
				entry.CallerInfo.GID = getGID()
			}
		}
	} else if len(l.logFilterSpec.filters) > 0 || len(l.traceFilterSpec.filters) > 0 {
		// Perform tests to see if we should log this message.
		var allowLog bool
		if traceLevel == notATrace {
			if l.logFilterSpec.matchfilters("", int(logLevel)) {
				allowLog = true
			}
		} else {
			if l.traceFilterSpec.matchfilters("", traceLevel) {
				allowLog = true
			}
		}
		if !allowLog {
			return
		}
	}

	msgCapacity := 1 + len(a)
	if len(l.additionalInformation) > 0 {
		msgCapacity++
	}
	if len(additionalInformation) > 0 {
		msgCapacity++
	}

	line := l.Formatter().Format(entry)
	if l.logWriterStream != nil {
		func() {
			l.mutex.Lock()
			l.mutex.Unlock()
			l.logWriterStream.Write(line)
		}()
	}
	if l.logWriterFile != nil {
		l.logWriterFile.Print(string(line))
	}
	ReleaseOutput(line)
}

func (l *logger) WithField(name string, value interface{}) Logger {
	return newSubLogger(l, Fields{
		name: value,
	})
}

func (l *logger) WithFields(fields Fields) Logger {
	return newSubLogger(l, fields)
}

// getGID gets the current goroutine ID (algorithm from
// https://blog.sgmansfield.com/2015/12/goroutine-ids/) by
// unwinding the stack.
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// Trace is for low level tracing of activities. It takes an additional 'level'
// parameter. The RLOG_TRACE_LEVEL variable is used to determine which levels
// of trace message are output: Every message with a level lower or equal to
// what is specified in RLOG_TRACE_LEVEL. If RLOG_TRACE_LEVEL is not defined at
// all then no trace messages are printed.
func (l *logger) Trace(traceLevel int, a ...interface{}) {
	// There are possibly many trace messages. If trace logging isn't enabled
	// then we want to get out of here as quickly as possible.
	if len(l.traceFilterSpec.filters) > 0 {
		l.BasicLog(levelTrace, traceLevel, "", "", a...)
	}
}

// Tracef prints trace messages, with formatting.
func (l *logger) Tracef(traceLevel int, format string, a ...interface{}) {
	// There are possibly many trace messages. If trace logging isn't enabled
	// then we want to get out of here as quickly as possible.
	if len(l.traceFilterSpec.filters) > 0 {
		l.BasicLog(levelTrace, traceLevel, "", format, a...)
	}
}

// Debug prints a message if RLOG_LEVEL is set to DEBUG.
func (l *logger) Debug(a ...interface{}) {
	l.BasicLog(levelDebug, notATrace, "", "", a...)
}

// Debugf prints a message if RLOG_LEVEL is set to DEBUG, with formatting.
func (l *logger) Debugf(format string, a ...interface{}) {
	l.BasicLog(levelDebug, notATrace, "", format, a...)
}

// Info prints a message if RLOG_LEVEL is set to INFO or lower.
func (l *logger) Info(a ...interface{}) {
	l.BasicLog(levelInfo, notATrace, "", "", a...)
}

// Infof prints a message if RLOG_LEVEL is set to INFO or lower, with
// formatting.
func (l *logger) Infof(format string, a ...interface{}) {
	l.BasicLog(levelInfo, notATrace, "", format, a...)
}

// Println prints a message if RLOG_LEVEL is set to INFO or lower.
// Println shouldn't be used except for backward compatibility
// with standard log package, directly using Info is preferred way.
func (l *logger) Println(a ...interface{}) {
	l.BasicLog(levelInfo, notATrace, "", "", a...)
}

// Printf prints a message if RLOG_LEVEL is set to INFO or lower, with
// formatting.
// Printf shouldn't be used except for backward compatibility
// with standard log package, directly using Infof is preferred way.
func (l *logger) Printf(format string, a ...interface{}) {
	l.BasicLog(levelInfo, notATrace, "", format, a...)
}

// Warn prints a message if RLOG_LEVEL is set to WARN or lower.
func (l *logger) Warn(a ...interface{}) {
	l.BasicLog(levelWarn, notATrace, "", "", a...)
}

// Warnf prints a message if RLOG_LEVEL is set to WARN or lower, with
// formatting.
func (l *logger) Warnf(format string, a ...interface{}) {
	l.BasicLog(levelWarn, notATrace, "", format, a...)
}

// Error prints a message if RLOG_LEVEL is set to ERROR or lower.
func (l *logger) Error(a ...interface{}) {
	l.BasicLog(levelErr, notATrace, "", "", a...)
}

// Errorf prints a message if RLOG_LEVEL is set to ERROR or lower, with
// formatting.
func (l *logger) Errorf(format string, a ...interface{}) {
	l.BasicLog(levelErr, notATrace, "", format, a...)
}

// Critical prints a message if RLOG_LEVEL is set to CRITICAL or lower.
func (l *logger) Critical(a ...interface{}) {
	l.BasicLog(levelCrit, notATrace, "", "", a...)
}

// Criticalf prints a message if RLOG_LEVEL is set to CRITICAL or lower, with
// formatting.
func (l *logger) Criticalf(format string, a ...interface{}) {
	l.BasicLog(levelCrit, notATrace, "", format, a...)
}

// WithField returns a new sublogger with the new field in the context.
func WithField(name string, value interface{}) Logger {
	return defaultLogger.WithField(name, value)
}

// WithFields returns a new sublogger with the new fields in the context.
func WithFields(fields Fields) Logger {
	return defaultLogger.WithFields(fields)
}

func Trace(traceLevel int, a ...interface{}) {
	// There are possibly many trace messages. If trace logging isn't enabled
	// then we want to get out of here as quickly as possible.
	if len(defaultLogger.traceFilterSpec.filters) > 0 {
		defaultLogger.BasicLog(levelTrace, traceLevel, "", "", a...)
	}
}

// Tracef prints trace messages, with formatting.
func Tracef(traceLevel int, format string, a ...interface{}) {
	// There are possibly many trace messages. If trace logging isn't enabled
	// then we want to get out of here as quickly as possible.
	if len(defaultLogger.traceFilterSpec.filters) > 0 {
		defaultLogger.BasicLog(levelTrace, traceLevel, "", format, a...)
	}
}

// Debug prints a message if RLOG_LEVEL is set to DEBUG.
func Debug(a ...interface{}) {
	defaultLogger.BasicLog(levelDebug, notATrace, "", "", a...)
}

// Debugf prints a message if RLOG_LEVEL is set to DEBUG, with formatting.
func Debugf(format string, a ...interface{}) {
	defaultLogger.BasicLog(levelDebug, notATrace, "", format, a...)
}

// Info prints a message if RLOG_LEVEL is set to INFO or lower.
func Info(a ...interface{}) {
	defaultLogger.BasicLog(levelInfo, notATrace, "", "", a...)
}

// Infof prints a message if RLOG_LEVEL is set to INFO or lower, with
// formatting.
func Infof(format string, a ...interface{}) {
	defaultLogger.BasicLog(levelInfo, notATrace, "", "", "", fmt.Sprintf(format, a...))
}

// Println prints a message if RLOG_LEVEL is set to INFO or lower.
// Println shouldn't be used except for backward compatibility
// with standard log package, directly using Info is preferred way.
func Println(a ...interface{}) {
	defaultLogger.BasicLog(levelInfo, notATrace, "", "", a...)
}

// Printf prints a message if RLOG_LEVEL is set to INFO or lower, with
// formatting.
// Printf shouldn't be used except for backward compatibility
// with standard log package, directly using Infof is preferred way.
func Printf(format string, a ...interface{}) {
	defaultLogger.BasicLog(levelInfo, notATrace, "", format, a...)
}

// Warn prints a message if RLOG_LEVEL is set to WARN or lower.
func Warn(a ...interface{}) {
	defaultLogger.BasicLog(levelWarn, notATrace, "", "", a...)
}

// Warnf prints a message if RLOG_LEVEL is set to WARN or lower, with
// formatting.
func Warnf(format string, a ...interface{}) {
	defaultLogger.BasicLog(levelWarn, notATrace, "", format, a...)
}

// Error prints a message if RLOG_LEVEL is set to ERROR or lower.
func Error(a ...interface{}) {
	defaultLogger.BasicLog(levelErr, notATrace, "", "", a...)
}

// Errorf prints a message if RLOG_LEVEL is set to ERROR or lower, with
// formatting.
func Errorf(format string, a ...interface{}) {
	defaultLogger.BasicLog(levelErr, notATrace, "", format, a...)
}

// Critical prints a message if RLOG_LEVEL is set to CRITICAL or lower.
func Critical(a ...interface{}) {
	defaultLogger.BasicLog(levelCrit, notATrace, "", "", a...)
}

// Criticalf prints a message if RLOG_LEVEL is set to CRITICAL or lower, with
// formatting.
func Criticalf(format string, a ...interface{}) {
	defaultLogger.BasicLog(levelCrit, notATrace, "", format, a...)
}
