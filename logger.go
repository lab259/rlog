package rlog

import "fmt"

type Fields map[string]interface{}

type FieldsArr []interface{}

// Logger is the interface that represents a logging unit.
type Logger interface {
	WithPrefix(prefix string) Logger
	WithField(name string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithFieldsArr(fields ...interface{}) Logger
	Formatter() LogFormatter
	BasicLog(logLevel Level, traceLevel int, additionalInformation string, fields FieldsArr, format string, a ...interface{})
	Trace(level int, a ...interface{})
	Tracef(level int, format string, a ...interface{})
	Debug(a ...interface{})
	Debugf(format string, a ...interface{})
	Info(a ...interface{})
	Infof(format string, a ...interface{})
	Println(a ...interface{})
	Printf(format string, a ...interface{})
	Warn(a ...interface{})
	Warnf(format string, a ...interface{})
	Error(a ...interface{})
	Errorf(format string, a ...interface{})
	Critical(a ...interface{})
	Criticalf(format string, a ...interface{})
}

// subLogger is a cheap struct that works on top of a `Logger` for aggregation
// additional information to the entries triggered by it.
type subLogger struct {
	logger                Logger
	prefix                string
	additionalInformation string
	additionalFields      FieldsArr
}

func newSubLogger(logger Logger, fields FieldsArr) *subLogger {
	return &subLogger{
		logger:                logger,
		additionalInformation: logger.Formatter().FormatFields(fields),
		additionalFields:      fields,
	}
}

func (logger *subLogger) WithPrefix(prefix string) Logger {
	l := newSubLogger(logger, nil)
	l.prefix = prefix
	return l
}

func (logger *subLogger) WithField(name string, value interface{}) Logger {
	return newSubLogger(logger, FieldsArr{name, value})
}

func newFieldsArrFromFields(fields Fields) FieldsArr {
	r := make(FieldsArr, len(fields)*2)
	i := 0
	for k, v := range fields {
		r[i] = k
		r[i+1] = v
		i += 2
	}
	return r
}

func (logger *subLogger) WithFields(fields Fields) Logger {
	return newSubLogger(logger, newFieldsArrFromFields(fields))
}

func (logger *subLogger) WithFieldsArr(fields ...interface{}) Logger {
	return newSubLogger(logger, fields)
}

func (logger *subLogger) Formatter() LogFormatter {
	return logger.logger.Formatter()
}

func (logger *subLogger) BasicLog(logLevel Level, traceLevel int, additionalInformation string, fields FieldsArr, format string, a ...interface{}) {
	ai := logger.additionalInformation
	if len(ai) > 0 && len(additionalInformation) > 0 {
		ai = fmt.Sprint(ai, logger.Formatter().Separator(), additionalInformation)
	} else {
		ai = additionalInformation
	}
	if logger.prefix != "" {
		if format == "" {
			a = append([]interface{}{logger.prefix}, a...)
		} else {
			format = logger.prefix + format
		}
	}
	logger.logger.BasicLog(logLevel, traceLevel, ai, append(logger.additionalFields, fields...), format, a...)
}

func (logger *subLogger) internalLog(logLevel Level, traceLevel int, format string, a ...interface{}) {
	logger.logger.BasicLog(logLevel, traceLevel, logger.additionalInformation, logger.additionalFields, format, a...)
}

// Trace is for low level tracing of activities. It takes an additional 'level'
// parameter. The RLOG_TRACE_LEVEL variable is used to determine which levels
// of trace message are output: Every message with a level lower or equal to
// what is specified in RLOG_TRACE_LEVEL. If RLOG_TRACE_LEVEL is not defined at
// all then no trace messages are printed.
func (logger *subLogger) Trace(traceLevel int, a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelTrace, traceLevel, "", a...)
	} else {
		logger.internalLog(levelTrace, traceLevel, "", a...)
	}
}

// Tracef prints trace messages, with formatting.
func (logger *subLogger) Tracef(traceLevel int, format string, a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelTrace, traceLevel, format, a...)
	} else {
		logger.internalLog(levelTrace, traceLevel, format, a...)
	}
}

// Debug prints a message if RLOG_LEVEL is set to DEBUG.
func (logger *subLogger) Debug(a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelDebug, notATrace, "", a...)
	} else {
		logger.internalLog(levelDebug, notATrace, "", a...)
	}
}

// Debugf prints a message if RLOG_LEVEL is set to DEBUG, with formatting.
func (logger *subLogger) Debugf(format string, a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelDebug, notATrace, format, a...)
	} else {
		logger.internalLog(levelDebug, notATrace, format, a...)
	}
}

// Info prints a message if RLOG_LEVEL is set to INFO or lower.
func (logger *subLogger) Info(a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelInfo, notATrace, "", a...)
	} else {
		logger.internalLog(levelInfo, notATrace, "", a...)
	}
}

// Infof prints a message if RLOG_LEVEL is set to INFO or lower, with
// formatting.
func (logger *subLogger) Infof(format string, a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelInfo, notATrace, format, a...)
	} else {
		logger.internalLog(levelInfo, notATrace, format, a...)
	}
}

// Println prints a message if RLOG_LEVEL is set to INFO or lower.
// Println shouldn't be used except for backward compatibility
// with standard log package, directly using Info is preferred way.
func (logger *subLogger) Println(a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelInfo, notATrace, "", a...)
	} else {
		logger.internalLog(levelInfo, notATrace, "", a...)
	}
}

// Printf prints a message if RLOG_LEVEL is set to INFO or lower, with
// formatting.
// Printf shouldn't be used except for backward compatibility
// with standard log package, directly using Infof is preferred way.
func (logger *subLogger) Printf(format string, a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelInfo, notATrace, format, a...)
	} else {
		logger.internalLog(levelInfo, notATrace, format, a...)
	}
}

// Warn prints a message if RLOG_LEVEL is set to WARN or lower.
func (logger *subLogger) Warn(a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelWarn, notATrace, "", a...)
	} else {
		logger.internalLog(levelWarn, notATrace, "", a...)
	}
}

// Warnf prints a message if RLOG_LEVEL is set to WARN or lower, with
// formatting.
func (logger *subLogger) Warnf(format string, a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelWarn, notATrace, format, a...)
	} else {
		logger.internalLog(levelWarn, notATrace, format, a...)
	}
}

// Error prints a message if RLOG_LEVEL is set to ERROR or lower.
func (logger *subLogger) Error(a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelErr, notATrace, "", a...)
	} else {
		logger.internalLog(levelErr, notATrace, "", a...)
	}
}

// Errorf prints a message if RLOG_LEVEL is set to ERROR or lower, with
// formatting.
func (logger *subLogger) Errorf(format string, a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelErr, notATrace, format, a...)
	} else {
		logger.internalLog(levelErr, notATrace, format, a...)
	}
}

// Critical prints a message if RLOG_LEVEL is set to CRITICAL or lower.
func (logger *subLogger) Critical(a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelCrit, notATrace, "", a...)
	} else {
		logger.internalLog(levelCrit, notATrace, "", a...)
	}
}

// Criticalf prints a message if RLOG_LEVEL is set to CRITICAL or lower, with
// formatting.
func (logger *subLogger) Criticalf(format string, a ...interface{}) {
	if logger != nil {
		logger.internalLog(levelCrit, notATrace, format, a...)
	} else {
		logger.internalLog(levelCrit, notATrace, format, a...)
	}
}
