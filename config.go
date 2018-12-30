package rlog

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// rlogConfig captures the entire configuration of rlog, as supplied by a user
// via environment variables and/or config files. This still requires checking
// and translation into more easily used config items. All values therefore are
// stored as simple strings here.
type Config struct {
	// What log level. String, since filters are allowed
	LogLevel string
	// What trace level. String, since filters are allowed
	TraceLevel string
	// The time format spec for date/time stamps in output
	logTimeFormat string
	// Name of logfile
	LogFile string
	// Name of config file
	confFile string
	// Name of logstream: stdout, stderr or NONE
	LogStream string
	// Flag to determine if date/time is logged at all
	LogNoTime bool
	// CallerInfo is a flag to determine if caller info is logged
	ShowCallerInfo bool
	// Flag to determine if goroute ID shows in caller info
	ShowGoroutineID bool
	// Interval in seconds for checking config file
	confCheckInterv string
}

// LoadFromEnv loads the configuration from the env variables.
//
// If the prefix is empty, it uses the RLOG prefix.
func (config *Config) LoadFromEnv(prefix string) {
	if prefix == "" {
		prefix = "RLOG"
	}
	// Read the initial configuration from the environment variables
	*config = Config{
		LogLevel:        os.Getenv(fmt.Sprintf("%s_LOG_LEVEL", prefix)),
		TraceLevel:      os.Getenv(fmt.Sprintf("%s_TRACE_LEVEL", prefix)),
		logTimeFormat:   os.Getenv(fmt.Sprintf("%s_TIME_FORMAT", prefix)),
		LogFile:         os.Getenv(fmt.Sprintf("%s_LOG_FILE", prefix)),
		confFile:        os.Getenv(fmt.Sprintf("%s_CONF_FILE", prefix)),
		LogStream:       strings.ToUpper(os.Getenv(fmt.Sprintf("%s_LOG_STREAM", prefix))),
		LogNoTime:       isTrueBoolString(os.Getenv(fmt.Sprintf("%s_LOG_NOTIME", prefix))),
		ShowCallerInfo:  isTrueBoolString(os.Getenv(fmt.Sprintf("%s_CALLER_INFO", prefix))),
		ShowGoroutineID: isTrueBoolString(os.Getenv(fmt.Sprintf("%s_GOROUTINE_ID", prefix))),
		confCheckInterv: os.Getenv(fmt.Sprintf("%s_CONF_CHECK_INTERVAL", prefix)),
	}
}

func (config *Config) loadFromStream(stream io.Reader) error {
	scanner := bufio.NewScanner(stream)
	lineN := 0
	for scanner.Scan() {
		lineN++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) != 2 {
			return fmt.Errorf("malformed line at line %d", lineN)
		}
		name := strings.TrimSpace(tokens[0])
		val := strings.TrimSpace(tokens[1])

		switch name {
		case "RLOG_LOG_LEVEL":
			config.LogLevel = val
		case "RLOG_TRACE_LEVEL":
			config.TraceLevel = val
		case "RLOG_TIME_FORMAT":
			config.logTimeFormat = val
		case "RLOG_LOG_FILE":
			config.LogFile = val
		case "RLOG_LOG_STREAM":
			val = strings.ToUpper(val)
			config.LogStream = val
		case "RLOG_LOG_NOTIME":
			config.LogNoTime = isTrueBoolString(val)
		case "RLOG_CALLER_INFO":
			config.ShowCallerInfo = isTrueBoolString(val)
		case "RLOG_GOROUTINE_ID":
			config.ShowGoroutineID = isTrueBoolString(val)
		default:
			rlogIssue("Unknown or illegal setting name in config file %d. Ignored.", lineN)
		}
	}
	return nil
}

// LoadFromFile load the configuration from a file.
func (config *Config) LoadFromFile(fileName string) error {
	// Scan over the config file, line by line
	file, err := os.Open(fileName)
	if err != nil {
		// Any error while attempting to open the logfile ignored. In many
		// cases there won't even be a config file, so we should not produce
		// any noise.
		return err
	}
	defer file.Close()

	return config.loadFromStream(file)
}

// We keep a copy of what was supplied via environment variables, since we will
// consult this every time we read from a config file. This allows us to
// determine which values take precedence.
var configFromEnvVars Config

// configFromEnv extracts settings for our logger from environment variables.
func configFromEnv() Config {
	// Read the initial configuration from the environment variables
	var config Config
	config.LoadFromEnv(" RLOG")
	return config
}
