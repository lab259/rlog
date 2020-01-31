package rlog

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {
	Describe("WithFields", func() {
		When("using defaultFormatter", func() {
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
				Expect(strings.TrimSpace(buff.String())).To(Equal(`INFO[00000] this is a INFO
INFO[00000] this is in a sublogger                                       var1=value1`))
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
				Expect(strings.TrimSpace(buff.String())).To(Equal(`INFO[00000] this is a INFO
INFO[00000] this is in a sublogger                                       var1=value1 var2=value2`))
			})
		})

		When("using TextFormatter", func() {
			It("should aggregate fields in a sublogger", func() {
				logger, err := NewLogger(Config{
					Formatter:  "text",
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
				Expect(strings.TrimSpace(buff.String())).To(Equal(`level=INFO msg="this is a INFO"
level=INFO var1=value1 msg="this is in a sublogger"`))
			})

			It("should aggregate fields in sequence", func() {
				logger, err := NewLogger(Config{
					Formatter:  "text",
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
				Expect(strings.TrimSpace(buff.String())).To(Equal(`level=INFO msg="this is a INFO"
level=INFO var1=value1 var2=value2 msg="this is in a sublogger"`))
			})
		})
	})

	Describe("WithPrefix", func() {
		It("should add a prefix for non format functions", func() {
			logger, err := NewLogger(Config{
				LogNoTime:  true,
				TraceLevel: "10",
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Info("this is a INFO")

			sublogger := logger.WithPrefix("prefix1").WithField("var1", "value1")
			sublogger.Info("this is in a sublogger")

			msgs := strings.Split(strings.TrimSpace(buff.String()), "\n")
			Expect(msgs).To(HaveLen(2))
			Expect(msgs[0]).To(Equal(`INFO[00000] this is a INFO`))
			Expect(msgs[1]).To(Equal(`INFO[00000] prefix1this is in a sublogger                                var1=value1`))
		})

		It("should add a prefix for non format functions", func() {
			logger, err := NewLogger(Config{
				LogNoTime:  true,
				TraceLevel: "10",
			})
			buff := bytes.NewBuffer(nil)
			logger.SetOutput(buff)
			Expect(err).ToNot(HaveOccurred())
			logger.Infof("this is a INFO: %d", 1)

			sublogger := logger.WithPrefix("prefix1").WithField("var1", "value1")
			sublogger.Infof("this is in a sublogger: %d", 2)

			msgs := strings.Split(strings.TrimSpace(buff.String()), "\n")
			Expect(msgs).To(HaveLen(2))
			Expect(msgs[0]).To(Equal(`INFO[00000] this is a INFO: 1`))
			Expect(msgs[1]).To(Equal(`INFO[00000] prefix1this is in a sublogger: 2                             var1=value1`))
		})
	})
})
