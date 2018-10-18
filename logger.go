package rlog

import "fmt"

// The logger is a cheap struct that adds the ContextInfo right after the
// logging messages for all logging scopes.
type Logger struct {
	ContextInfo []interface{}
}

// NewLogger creates a new instance of `Logger` the `ContextInfo` member
// initialized.
func NewLogger(info ... interface{}) *Logger {
	return &Logger{
		ContextInfo: info,
	}
}

// Trace is for low level tracing of activities. It takes an additional 'level'
// parameter. The RLOG_TRACE_LEVEL variable is used to determine which levels
// of trace message are output: Every message with a level lower or equal to
// what is specified in RLOG_TRACE_LEVEL. If RLOG_TRACE_LEVEL is not defined at
// all then no trace messages are printed.
func (logger *Logger) Trace(traceLevel int, a ...interface{}) {
	// There are possibly many trace messages. If trace logging isn't enabled
	// then we want to get out of here as quickly as possible.
	initMutex.RLock()
	defer initMutex.RUnlock()
	if len(traceFilterSpec.filters) > 0 {
		prefixAddition := fmt.Sprintf("(%d)", traceLevel)
		basicLog(levelTrace, traceLevel, true, logger.ContextInfo, "", prefixAddition, a...)
	}
}

// Tracef prints trace messages, with formatting.
func (logger *Logger) Tracef(traceLevel int, format string, a ...interface{}) {
	// There are possibly many trace messages. If trace logging isn't enabled
	// then we want to get out of here as quickly as possible.
	initMutex.RLock()
	defer initMutex.RUnlock()
	if len(traceFilterSpec.filters) > 0 {
		prefixAddition := fmt.Sprintf("(%d)", traceLevel)
		basicLog(levelTrace, traceLevel, true, logger.ContextInfo, format, prefixAddition, a...)
	}
}

// Debug prints a message if RLOG_LEVEL is set to DEBUG.
func (logger *Logger) Debug(a ...interface{}) {
	basicLog(levelDebug, notATrace, false, logger.ContextInfo, "", "", a...)
}

// Debugf prints a message if RLOG_LEVEL is set to DEBUG, with formatting.
func (logger *Logger) Debugf(format string, a ...interface{}) {
	basicLog(levelDebug, notATrace, false, logger.ContextInfo, format, "", a...)
}

// Info prints a message if RLOG_LEVEL is set to INFO or lower.
func (logger *Logger) Info(a ...interface{}) {
	basicLog(levelInfo, notATrace, false, logger.ContextInfo, "", "", a...)
}

// Infof prints a message if RLOG_LEVEL is set to INFO or lower, with
// formatting.
func (logger *Logger) Infof(format string, a ...interface{}) {
	basicLog(levelInfo, notATrace, false, logger.ContextInfo, format, "", a...)
}

// Println prints a message if RLOG_LEVEL is set to INFO or lower.
// Println shouldn't be used except for backward compatibility
// with standard log package, directly using Info is preferred way.
func (logger *Logger) Println(a ...interface{}) {
	basicLog(levelInfo, notATrace, false, logger.ContextInfo, "", "", a...)
}

// Printf prints a message if RLOG_LEVEL is set to INFO or lower, with
// formatting.
// Printf shouldn't be used except for backward compatibility
// with standard log package, directly using Infof is preferred way.
func (logger *Logger) Printf(format string, a ...interface{}) {
	basicLog(levelInfo, notATrace, false, logger.ContextInfo, format, "", a...)
}

// Warn prints a message if RLOG_LEVEL is set to WARN or lower.
func (logger *Logger) Warn(a ...interface{}) {
	basicLog(levelWarn, notATrace, false, logger.ContextInfo, "", "", a...)
}

// Warnf prints a message if RLOG_LEVEL is set to WARN or lower, with
// formatting.
func (logger *Logger) Warnf(format string, a ...interface{}) {
	basicLog(levelWarn, notATrace, false, logger.ContextInfo, format, "", a...)
}

// Error prints a message if RLOG_LEVEL is set to ERROR or lower.
func (logger *Logger) Error(a ...interface{}) {
	basicLog(levelErr, notATrace, false, logger.ContextInfo, "", "", a...)
}

// Errorf prints a message if RLOG_LEVEL is set to ERROR or lower, with
// formatting.
func (logger *Logger) Errorf(format string, a ...interface{}) {
	basicLog(levelErr, notATrace, false, logger.ContextInfo, format, "", a...)
}

// Critical prints a message if RLOG_LEVEL is set to CRITICAL or lower.
func (logger *Logger) Critical(a ...interface{}) {
	basicLog(levelCrit, notATrace, false, logger.ContextInfo, "", "", a...)
}

// Criticalf prints a message if RLOG_LEVEL is set to CRITICAL or lower, with
// formatting.
func (logger *Logger) Criticalf(format string, a ...interface{}) {
	basicLog(levelCrit, notATrace, false, logger.ContextInfo, format, "", a...)
}
