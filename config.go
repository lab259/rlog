package rlog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// rlogConfig captures the entire configuration of rlog, as supplied by a user
// via environment variables and/or config files. This still requires checking
// and translation into more easily used config items. All values therefore are
// stored as simple strings here.
type rlogConfig struct {
	logLevel        string // What log level. String, since filters are allowed
	traceLevel      string // What trace level. String, since filters are allowed
	logTimeFormat   string // The time format spec for date/time stamps in output
	logFile         string // Name of logfile
	confFile        string // Name of config file
	logStream       string // Name of logstream: stdout, stderr or NONE
	logNoTime       string // Flag to determine if date/time is logged at all
	showCallerInfo  string // Flag to determine if caller info is logged
	showGoroutineID string // Flag to determine if goroute ID shows in caller info
	confCheckInterv string // Interval in seconds for checking config file
}

// LoadFromEnv loads the configuration from the env variables.
//
// If the prefix is empty, it uses the RLOG prefix.
func (config *rlogConfig) LoadFromEnv(prefix string) {
	if prefix == "" {
		prefix = "RLOG"
	}
	// Read the initial configuration from the environment variables
	*config = rlogConfig{
		logLevel:        os.Getenv(fmt.Sprintf("%s_LOG_LEVEL", prefix)),
		traceLevel:      os.Getenv(fmt.Sprintf("%s_TRACE_LEVEL", prefix)),
		logTimeFormat:   os.Getenv(fmt.Sprintf("%s_TIME_FORMAT", prefix)),
		logFile:         os.Getenv(fmt.Sprintf("%s_LOG_FILE", prefix)),
		confFile:        os.Getenv(fmt.Sprintf("%s_CONF_FILE", prefix)),
		logStream:       strings.ToUpper(os.Getenv(fmt.Sprintf("%s_LOG_STREAM", prefix))),
		logNoTime:       os.Getenv(fmt.Sprintf("%s_LOG_NOTIME", prefix)),
		showCallerInfo:  os.Getenv(fmt.Sprintf("%s_CALLER_INFO", prefix)),
		showGoroutineID: os.Getenv(fmt.Sprintf("%s_GOROUTINE_ID", prefix)),
		confCheckInterv: os.Getenv(fmt.Sprintf("%s_CONF_CHECK_INTERVAL", prefix)),
	}
}

// We keep a copy of what was supplied via environment variables, since we will
// consult this every time we read from a config file. This allows us to
// determine which values take precedence.
var configFromEnvVars rlogConfig

// updateConfigFromFile reads a configuration from the specified config file.
// It merges the supplied config with the new values.
func updateConfigFromFile(config *rlogConfig) {
	lastConfigFileCheck = time.Now()

	settingConfFile = config.confFile
	// If no config file was specified we will default to a known location.
	if settingConfFile == "" {
		execName := filepath.Base(os.Args[0])
		settingConfFile = fmt.Sprintf("/etc/rlog/%s.conf", execName)
	}

	// Scan over the config file, line by line
	file, err := os.Open(settingConfFile)
	if err != nil {
		// Any error while attempting to open the logfile ignored. In many
		// cases there won't even be a config file, so we should not produce
		// any noise.
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		i++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) == 0 {
			continue
		}
		if len(tokens) != 2 {
			rlogIssue("Malformed line in config file %s:%d. Ignored.",
				settingConfFile, i)
			continue
		}
		name := strings.TrimSpace(tokens[0])
		val := strings.TrimSpace(tokens[1])

		// If the name starts with a '!' then it should overwrite whatever we
		// currently have in the config already.
		priority := false
		if name[0] == '!' {
			priority = true
			name = name[1:]
		}

		switch name {
		case "RLOG_LOG_LEVEL":
			config.logLevel = updateIfNeeded(config.logLevel, val, priority)
		case "RLOG_TRACE_LEVEL":
			config.traceLevel = updateIfNeeded(config.traceLevel, val, priority)
		case "RLOG_TIME_FORMAT":
			config.logTimeFormat = updateIfNeeded(config.logTimeFormat, val, priority)
		case "RLOG_LOG_FILE":
			config.logFile = updateIfNeeded(config.logFile, val, priority)
		case "RLOG_LOG_STREAM":
			val = strings.ToUpper(val)
			config.logStream = updateIfNeeded(config.logStream, val, priority)
		case "RLOG_LOG_NOTIME":
			config.logNoTime = updateIfNeeded(config.logNoTime, val, priority)
		case "RLOG_CALLER_INFO":
			config.showCallerInfo = updateIfNeeded(config.showCallerInfo, val, priority)
		case "RLOG_GOROUTINE_ID":
			config.showGoroutineID = updateIfNeeded(config.showGoroutineID, val, priority)
		default:
			rlogIssue("Unknown or illegal setting name in config file %s:%d. Ignored.",
				settingConfFile, i)
		}
	}
}

// configFromEnv extracts settings for our logger from environment variables.
func configFromEnv() rlogConfig {
	// Read the initial configuration from the environment variables
	var config rlogConfig
	config.LoadFromEnv(" RLOG")
	return config
}
