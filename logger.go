package rlog

import "fmt"

type Fields map[string]interface{}

// Logger is the interface that represents a logging unit.
type Logger interface {
	WithField(name string, value interface{}) Logger
	WithFields(fields Fields) Logger
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

// The logger is a cheap struct that adds the ContextInfo right after the
// logging messages for all logging scopes.
type loggerImpl struct {
	additionalInformation string
}

// NewLogger creates a new instance of `Logger` the `ContextInfo` member
// initialized.
func NewLogger() *loggerImpl {
	return &loggerImpl{
		additionalInformation: "",
	}
}

// Trace is for low level tracing of activities. It takes an additional 'level'
// parameter. The RLOG_TRACE_LEVEL variable is used to determine which levels
// of trace message are output: Every message with a level lower or equal to
// what is specified in RLOG_TRACE_LEVEL. If RLOG_TRACE_LEVEL is not defined at
// all then no trace messages are printed.
func (logger *loggerImpl) Trace(traceLevel int, a ...interface{}) {
	// There are possibly many trace messages. If trace logging isn't enabled
	// then we want to get out of here as quickly as possible.
	initMutex.RLock()
	defer initMutex.RUnlock()
	if len(traceFilterSpec.filters) > 0 {
		prefixAddition := fmt.Sprintf("(%d)", traceLevel)
		if logger != nil {
			basicLog(levelTrace, traceLevel, true, logger.additionalInformation, "", prefixAddition, a...)
		} else {
			basicLog(levelTrace, traceLevel, true, "", "", prefixAddition, a...)
		}
	}
}

// Tracef prints trace messages, with formatting.
func (logger *loggerImpl) Tracef(traceLevel int, format string, a ...interface{}) {
	// There are possibly many trace messages. If trace logging isn't enabled
	// then we want to get out of here as quickly as possible.
	initMutex.RLock()
	defer initMutex.RUnlock()
	if len(traceFilterSpec.filters) > 0 {
		prefixAddition := fmt.Sprintf("(%d)", traceLevel)
		if logger != nil {
			basicLog(levelTrace, traceLevel, true, logger.additionalInformation, format, prefixAddition, a...)
		} else {
			basicLog(levelTrace, traceLevel, true, "", format, prefixAddition, a...)
		}
	}
}

// Debug prints a message if RLOG_LEVEL is set to DEBUG.
func (logger *loggerImpl) Debug(a ...interface{}) {
	if logger != nil {
		basicLog(levelDebug, notATrace, false, logger.additionalInformation, "", "", a...)
	} else {
		basicLog(levelDebug, notATrace, false, "", "", "", a...)
	}
}

// Debugf prints a message if RLOG_LEVEL is set to DEBUG, with formatting.
func (logger *loggerImpl) Debugf(format string, a ...interface{}) {
	if logger != nil {
		basicLog(levelDebug, notATrace, false, logger.additionalInformation, format, "", a...)
	} else {
		basicLog(levelDebug, notATrace, false, "", format, "", a...)
	}
}

// Info prints a message if RLOG_LEVEL is set to INFO or lower.
func (logger *loggerImpl) Info(a ...interface{}) {
	if logger != nil {
		basicLog(levelInfo, notATrace, false, logger.additionalInformation, "", "", a...)
	} else {
		basicLog(levelInfo, notATrace, false, "", "", "", a...)
	}
}

// Infof prints a message if RLOG_LEVEL is set to INFO or lower, with
// formatting.
func (logger *loggerImpl) Infof(format string, a ...interface{}) {
	if logger != nil {
		basicLog(levelInfo, notATrace, false, logger.additionalInformation, format, "", a...)
	} else {
		basicLog(levelInfo, notATrace, false, "", format, "", a...)
	}
}

// Println prints a message if RLOG_LEVEL is set to INFO or lower.
// Println shouldn't be used except for backward compatibility
// with standard log package, directly using Info is preferred way.
func (logger *loggerImpl) Println(a ...interface{}) {
	if logger != nil {
		basicLog(levelInfo, notATrace, false, logger.additionalInformation, "", "", a...)
	} else {
		basicLog(levelInfo, notATrace, false, "", "", "", a...)
	}
}

// Printf prints a message if RLOG_LEVEL is set to INFO or lower, with
// formatting.
// Printf shouldn't be used except for backward compatibility
// with standard log package, directly using Infof is preferred way.
func (logger *loggerImpl) Printf(format string, a ...interface{}) {
	if logger != nil {
		basicLog(levelInfo, notATrace, false, logger.additionalInformation, format, "", a...)
	} else {
		basicLog(levelInfo, notATrace, false, "", format, "", a...)
	}
}

// Warn prints a message if RLOG_LEVEL is set to WARN or lower.
func (logger *loggerImpl) Warn(a ...interface{}) {
	if logger != nil {
		basicLog(levelWarn, notATrace, false, logger.additionalInformation, "", "", a...)
	} else {
		basicLog(levelWarn, notATrace, false, "", "", "", a...)
	}
}

// Warnf prints a message if RLOG_LEVEL is set to WARN or lower, with
// formatting.
func (logger *loggerImpl) Warnf(format string, a ...interface{}) {
	if logger != nil {
		basicLog(levelWarn, notATrace, false, logger.additionalInformation, format, "", a...)
	} else {
		basicLog(levelWarn, notATrace, false, "", format, "", a...)
	}
}

// Error prints a message if RLOG_LEVEL is set to ERROR or lower.
func (logger *loggerImpl) Error(a ...interface{}) {
	if logger != nil {
		basicLog(levelErr, notATrace, false, logger.additionalInformation, "", "", a...)
	} else {
		basicLog(levelErr, notATrace, false, "", "", "", a...)
	}
}

// Errorf prints a message if RLOG_LEVEL is set to ERROR or lower, with
// formatting.
func (logger *loggerImpl) Errorf(format string, a ...interface{}) {
	if logger != nil {
		basicLog(levelErr, notATrace, false, logger.additionalInformation, format, "", a...)
	} else {
		basicLog(levelErr, notATrace, false, "", format, "", a...)
	}
}

// Critical prints a message if RLOG_LEVEL is set to CRITICAL or lower.
func (logger *loggerImpl) Critical(a ...interface{}) {
	if logger != nil {
		basicLog(levelCrit, notATrace, false, logger.additionalInformation, "", "", a...)
	} else {
		basicLog(levelCrit, notATrace, false, "", "", "", a...)
	}
}

// Criticalf prints a message if RLOG_LEVEL is set to CRITICAL or lower, with
// formatting.
func (logger *loggerImpl) Criticalf(format string, a ...interface{}) {
	if logger != nil {
		basicLog(levelCrit, notATrace, false, logger.additionalInformation, format, "", a...)
	} else {
		basicLog(levelCrit, notATrace, false, "", format, "", a...)
	}
}
