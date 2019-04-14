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
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jamillosantos/macchiato"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRLog(t *testing.T) {
	log.SetOutput(GinkgoWriter)
	if os.Getenv("ENV") == "" {
		err := os.Setenv("ENV", "test")
		if err != nil {
			panic(err)
		}
	}
	RegisterFailHandler(Fail)
	macchiato.RunSpecs(t, "RLog Test Suite")
}

var logfile string

// These two flags are used to quickly change behaviour of our tests, so that
// we can manually check and test things during development of those tests.
// The settings here reflect the correct behaviour for normal test runs.
var removeLogfile = true
var fixedLogfileName = false

// setup is called at the start of each test and prepares a new log file. It
// also returns a new configuration, as it may have been supplied by the user
// in environment variables, which can be used by this test.
func setup() Config {
	if fixedLogfileName {
		logfile = "/tmp/rlog-test.log"
	} else {
		logfile = fmt.Sprintf("/tmp/rlog-test-%d.log", time.Now().UnixNano())
	}

	// If there's a logfile with that name already, remove it so that our tests
	// always start from scratch.
	os.Remove(logfile)

	// Provide a default config, which can be used or modified by the tests
	return Config{
		LogLevel:       "",
		TraceLevel:     "",
		logTimeFormat:  "",
		confFile:       "",
		LogFile:        logfile,
		LogStream:      "NONE",
		LogNoTime:      true,
		ShowCallerInfo: false,
	}
}

// cleanup is called at the end of each test.
func cleanup() {
	if removeLogfile {
		os.Remove(logfile)
	}
}

// OverrideEnv stores the current state of the environment variables, sets the
// new values according with the passed params, calls the callback and finally
// restores the previous state.
func OverrideEnv(vars map[string]string, callback func() error) error {
	current := make(map[string]string)
	for k, v := range vars {
		current[k] = os.Getenv(k)
		os.Setenv(k, v)
	}
	err := callback()
	for k, v := range current {
		os.Setenv(k, v)
	}
	return err
}

// fileMatch compares entries in the logfile with expected entries provided as
// a list of strings (one for each line). If a timeLayout string is provided
// then we will assume the first part of the line is a timestamp, which we will
// check to see if it's correctly formatted according to the specified time
// layout.
func fileMatch(t *testing.T, checkLines []string, timeLayout string) {
	// We need to know how many characters at the start of the line we should
	// assume to belong to the time stamp. The formatted time stamp can
	// actually be of different length than the time layout string, because
	// actual timezone names can have more characters than the TZ specified in
	// the layout. So we create the current time in the specified layout, which
	// should be very similar to the timestamps in the log lines.
	currentSampleTimestamp := time.Now().Format(timeLayout)
	timeStampLen := len(currentSampleTimestamp)

	// Scan over the logfile, line by line and compare to the lines provided in
	// checkLines.
	file, err := os.Open(logfile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		line := scanner.Text()
		// Process and strip off the time stamp at the start, if a time layout
		// string was provided.
		if timeLayout != "" {
			dateTime := line[:timeStampLen]
			line = line[timeStampLen+1:]
			_, err := time.Parse(timeLayout, dateTime)
			if err != nil {
				t.Fatalf("Incorrect date/time format.\nSHOULD: %s\nIS:     %s\n", timeLayout, dateTime)
			}
		}
		t.Logf("\n-- fileMatch: SHOULD %s\n              IS     %s\n",
			checkLines[i], line)
		if i >= len(checkLines) {
			t.Fatal("Not enough lines provided in checkLines.")
		}
		if line != checkLines[i] {
			t.Fatalf("Log line %d does not match check line.\nSHOULD: %s\nIS:     %s\n", i, checkLines[i], line)
		}
		i++
	}
	if len(checkLines) > i {
		t.Fatalf("Only %d of %d checklines found in output file.", i, len(checkLines))
	}
	if i == 0 {
		t.Fatal("No input scanned")
	}
}

// ---------- Tests -----------

var _ = Describe("RLog Test Suite", func() {

	BeforeEach(func() {
		setup()
	})

	AfterEach(func() {
		cleanup()
	})

	Describe("Levels", func() {
		It("should place a INFO with date", func() {
			logger, err := NewLogger(Config{})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Critical("this is a CRITICAL")
			Expect(strings.TrimSpace(buff.String())).To(HavePrefix(`date="`))
			Expect(strings.TrimSpace(buff.String())).To(HaveSuffix(`level="CRITICAL" msg="this is a CRITICAL"`))
		})

		It("should place a CRITICAL line", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Critical("this is a CRITICAL")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="CRITICAL" msg="this is a CRITICAL"`))
		})

		It("should place a CRITICAL line with format", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Criticalf("this is a CRITICAL with format %s", "enabled")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="CRITICAL" msg="this is a CRITICAL with format enabled"`))
		})

		It("should ignore CRITICAL when level is NONE", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
				LogLevel:  "NONE",
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Critical("this is a CRITICAL")
			Expect(strings.TrimSpace(buff.String())).To(BeEmpty())
		})

		It("should place a ERROR line", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Error("this is a ERROR")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="ERROR" msg="this is a ERROR"`))
		})

		It("should place a ERROR line with format", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Errorf("this is a ERROR with format %s", "enabled")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="ERROR" msg="this is a ERROR with format enabled"`))
		})

		It("should ingore ERROR when level is CRITICAL", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
				LogLevel:  "CRITICAL",
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Error("this is a ERROR")
			Expect(strings.TrimSpace(buff.String())).To(BeEmpty())
		})

		It("should place a WARN line", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Warn("this is a WARN")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="WARN" msg="this is a WARN"`))
		})

		It("should place a WARN line with format", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Warnf("this is a WARN with format %s", "enabled")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="WARN" msg="this is a WARN with format enabled"`))
		})

		It("should ingore WARN when level is CRITICAL", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
				LogLevel:  "WARN",
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Warn("this is a WARN")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="WARN" msg="this is a WARN"`))
		})

		It("should place a INFO line", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Info("this is a INFO")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="INFO" msg="this is a INFO"`))
		})

		It("should place a INFO line with format", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Infof("this is a INFO with format %s", "enabled")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="INFO" msg="this is a INFO with format enabled"`))
		})

		It("should ingore INFO when level is WARN", func() {
			logger, err := NewLogger(Config{
				LogNoTime: true,
				LogLevel:  "WARN",
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Info("this is a INFO")
			Expect(strings.TrimSpace(buff.String())).To(BeEmpty())
		})

		It("should place a DEBUG line", func() {
			logger, err := NewLogger(Config{
				LogLevel:  "DEBUG",
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Debug("this is a DEBUG")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="DEBUG" msg="this is a DEBUG"`))
		})

		It("should place a DEBUG line with format", func() {
			logger, err := NewLogger(Config{
				LogLevel:  "DEBUG",
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Debugf("this is a DEBUG with format %s", "enabled")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="DEBUG" msg="this is a DEBUG with format enabled"`))
		})

		It("should ingore DEBUG when level is INFO", func() {
			logger, err := NewLogger(Config{
				LogLevel:  "INFO",
				LogNoTime: true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Debug("this is a DEBUG")
			Expect(strings.TrimSpace(buff.String())).To(BeEmpty())
		})

		It("should place a TRACE line", func() {
			logger, err := NewLogger(Config{
				LogLevel:   "DEBUG",
				TraceLevel: "10",
				LogNoTime:  true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Trace(1, "this is a TRACE")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="TRACE(1)" msg="this is a TRACE"`))
		})

		It("should place a TRACE line with format", func() {
			logger, err := NewLogger(Config{
				LogLevel:   "DEBUG",
				TraceLevel: "10",
				LogNoTime:  true,
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Tracef(1, "this is a TRACE with format %s", "enabled")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="TRACE(1)" msg="this is a TRACE with format enabled"`))
		})

		It("should ingore TRACE when trace level is greater than the set", func() {
			logger, err := NewLogger(Config{
				LogLevel:   "DEBUG",
				LogNoTime:  true,
				TraceLevel: "10",
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Trace(9, "this is a TRACE(9)")
			logger.Trace(10, "this is a TRACE(10)")
			logger.Trace(11, "this is a TRACE(11)")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="TRACE(9)" msg="this is a TRACE(9)"
level="TRACE(10)" msg="this is a TRACE(10)"`))
		})
	})
})

// writeLogfile is a small utility function for the creation of unique config
// files for these tests.
func writeLogfile(lines []string) string {
	confFile := fmt.Sprintf("/tmp/rlog-test-%d.conf", time.Now().UnixNano())
	cf, _ := os.Create(confFile)
	defer cf.Close()
	for _, l := range lines {
		cf.WriteString(l + "\n")
	}
	return confFile
}

// checkLogFilter simplifies the checking of correct log levels in the tests.
func checkLogFilter(t *testing.T, shouldPattern string, shouldLevel int) {
	f := defaultLogger.logFilterSpec.filters[0]
	if f.Pattern != shouldPattern || int(f.Level) != shouldLevel {
		t.Fatalf("Incorrect default filter '%s' / %d. Should be: '%s' / %d",
			f.Pattern, f.Level, shouldPattern, shouldLevel)
	}
}

// TestRaceConditions stress tests thread safety of rlog. Useful when running
// with the race detector flag (--race).
func TestRaceConditions(t *testing.T) {
	conf := setup()
	defer cleanup()

	for i := 0; i < 1000; i++ {
		go func(conf Config, i int) {
			for j := 0; j < 100; j++ {
				// Change behaviour and config around a little
				if j%2 == 0 {
					conf.ShowCallerInfo = true
				}
				conf.TraceLevel = strconv.Itoa(j%10 - 1) // sometimes this will be -1
				// //initialize(conf, j%3 == 0)
				// initialize(conf, false)
				Debug("Test Debug")
				Info("Test Info")
				Trace(1, "Some trace")
				Trace(2, "Some trace")
				Trace(3, "Some trace")
				Trace(4, "Some trace")
			}
		}(conf, i)
	}
}

func BenchmarkRLog(b *testing.B) {
	buff := bytes.NewBuffer(nil)
	logger, err := NewLogger(Config{
		TraceLevel: "30",
	})
	if err != nil {
		panic(err)
	}
	logger.SetOutput(buff)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		logger.Trace(1, "this is a test")
	}
}

func BenchmarkWithFields(b *testing.B) {
	buff := bytes.NewBuffer(nil)
	loggerMaster, err := NewLogger(Config{
		TraceLevel: "30",
	})
	if err != nil {
		panic(err)
	}
	loggerMaster.SetOutput(buff)
	logger := loggerMaster.WithFields(Fields{
		"var1": "value1",
		"var2": "value2",
		"var3": "value3",
		"var4": "value4",
	})
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		logger.Info("this is a test")
	}
}

func BenchmarkWithCallerInfo(b *testing.B) {
	buff := bytes.NewBuffer(nil)
	logger, err := NewLogger(Config{
		ShowCallerInfo: true,
	})
	if err != nil {
		panic(err)
	}
	logger.SetOutput(buff)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		logger.Info("this is a test")
	}
}

func BenchmarkMaps(b *testing.B) {
	b.ResetTimer()
	s := 0
	for n := 0; n < b.N; n++ {
		m := map[string]interface{}{
			fmt.Sprint("key", rand.Intn(100000)): rand.Intn(1000),
			fmt.Sprint("key", rand.Intn(100000)): rand.Intn(1000),
			fmt.Sprint("key", rand.Intn(100000)): rand.Intn(1000),
			fmt.Sprint("key", rand.Intn(100000)): rand.Intn(1000),
			fmt.Sprint("key", rand.Intn(100000)): rand.Intn(1000),
		}
		for _, v := range m {
			i, ok := v.(int)
			if ok {
				s += i
			}
		}
	}
	fmt.Println(s)
}

func BenchmarkArrays(b *testing.B) {
	b.ResetTimer()
	s := 0
	for n := 0; n < b.N; n++ {
		m := []interface{}{
			fmt.Sprint("key", rand.Intn(100000)), rand.Intn(1000),
			fmt.Sprint("key", rand.Intn(100000)), rand.Intn(1000),
			fmt.Sprint("key", rand.Intn(100000)), rand.Intn(1000),
			fmt.Sprint("key", rand.Intn(100000)), rand.Intn(1000),
			fmt.Sprint("key", rand.Intn(100000)), rand.Intn(1000),
		}
		for k := 0; k < len(m); k++ {
			i, ok := m[k+1].(int)
			if ok {
				s += i
			}
			k++
		}
	}
	fmt.Println(s)
}
