package rlog

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Logger", func() {
	Describe("WithFields", func() {
		It("should aggregate fields in a sublogger", func() {
			logger, err := NewLogger(Config{
				LogNoTime:  true,
				TraceLevel: "10",
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Info("this is a INFO")

			sublogger := logger.WithFields(Fields{
				"var1": "value1",
			})
			sublogger.Info("this is in a sublogger")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="INFO" msg="this is a INFO"
level="INFO" var1="value1" msg="this is in a sublogger"`))
		})

		It("should aggregate fields in sequence", func() {
			logger, err := NewLogger(Config{
				LogNoTime:  true,
				TraceLevel: "10",
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Info("this is a INFO")

			sublogger := logger.
				WithField("var1", "value1").
				WithField("var2", "value2")
			sublogger.Info("this is in a sublogger")
			Expect(strings.TrimSpace(buff.String())).To(Equal(`level="INFO" msg="this is a INFO"
level="INFO" var1="value1" var2="value2" msg="this is in a sublogger"`))
		})
	})
})
