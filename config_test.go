package rlog

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"time"
)

var _ = Describe("Config", func() {
	It("should load the log level from the stream", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_LOG_LEVEL=CRITICAL")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
		Expect(config.LogLevel).To(Equal("CRITICAL"))
	})

	It("should load the trace level from the stream", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_TRACE_LEVEL=10")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
		Expect(config.TraceLevel).To(Equal("10"))
	})

	It("should load the time format from the stream", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_TIME_FORMAT=unix")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
		Expect(config.logTimeFormat).To(Equal("unix"))
	})

	It("should load the log file from the stream", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_LOG_FILE=log")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
		Expect(config.LogFile).To(Equal("log"))
	})

	It("should load the log stream from the stream", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_LOG_STREAM=stdout")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
		Expect(config.LogStream).To(Equal("STDOUT"))
	})

	It("should load the no time flag from the stream", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_LOG_NOTIME=true")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
		Expect(config.LogNoTime).To(BeTrue())
	})

	It("should load the caller info from the stream", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_CALLER_INFO=true")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
		Expect(config.ShowCallerInfo).To(BeTrue())
	})

	It("should load the go routine id from the stream", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_GOROUTINE_ID=true")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
		Expect(config.ShowGoroutineID).To(BeTrue())
	})

	It("should not fail when finding a unknown variable", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_UNKNOWN_VARIABLE=anyvalue")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
	})

	It("should fail parsing an malformed file", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_LOG_LEVEL=DEBUG")
		fmt.Fprintln(buff, "RLOG_GOROUTINE_ID")

		var config Config
		err := config.loadFromStream(buff)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("malformed line at line 2"))
	})

	It("should parse multiple variables in the same file", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_LOG_LEVEL=DEBUG")
		fmt.Fprintln(buff, "")
		fmt.Fprintln(buff, "RLOG_TRACE_LEVEL=10")
		fmt.Fprintln(buff, "RLOG_GOROUTINE_ID=true")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
		Expect(config.LogLevel).To(Equal("DEBUG"))
		Expect(config.TraceLevel).To(Equal("10"))
		Expect(config.ShowGoroutineID).To(BeTrue())
	})

	It("should parse multiple variables in the same file with comments", func() {
		buff := bytes.NewBuffer(nil)
		fmt.Fprintln(buff, "RLOG_LOG_LEVEL=DEBUG")
		fmt.Fprintln(buff, "# RLOG_TRACE_LEVEL=10")
		fmt.Fprintln(buff, "# RLOG_GOROUTINE_ID=true")

		var config Config
		Expect(config.loadFromStream(buff)).To(Succeed())
		Expect(config.LogLevel).To(Equal("DEBUG"))
		Expect(config.TraceLevel).To(BeEmpty())
		Expect(config.ShowGoroutineID).To(BeFalse())
	})

	It("should parse multiple variables from a file", func() {
		tmpFileName := fmt.Sprintf("rlog_config_%d", time.Now().Unix())
		tmpFile, err := ioutil.TempFile(os.TempDir(), tmpFileName)
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(tmpFile.Name())

		fmt.Fprintln(tmpFile, "RLOG_LOG_LEVEL=DEBUG")
		fmt.Fprintln(tmpFile, "RLOG_TRACE_LEVEL=10")
		fmt.Fprintln(tmpFile, "RLOG_GOROUTINE_ID=true")
		tmpFile.Close()

		var config Config
		Expect(config.LoadFromFile(tmpFile.Name())).To(Succeed())
		Expect(config.LogLevel).To(Equal("DEBUG"))
		Expect(config.TraceLevel).To(Equal("10"))
		Expect(config.ShowGoroutineID).To(BeTrue())
	})

	It("should fail loading the configuration from a non existent file", func() {
		var config Config
		err := config.LoadFromFile("any-non-existing-file")
		Expect(err).To(HaveOccurred())
		Expect(os.IsNotExist(err)).To(BeTrue())
	})
})
