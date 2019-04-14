package rlog

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Formatter", func() {
	FDescribe("DefaultFormatter", func() {
		It("should format a field", func() {
			f := NewDefaultFormatter(os.Stderr)
			Expect(f.FormatField("key", "value")).To(Equal(`key=value`))
		})

		It("should format a field escaping values", func() {
			f := NewDefaultFormatter(os.Stderr)
			Expect(f.FormatField("key", `value with "quotes"`)).To(Equal(`key="value with \"quotes\""`))
		})

		It("should format fields", func() {
			f := NewDefaultFormatter(os.Stderr)
			fields := f.FormatFields(Fields{
				"field1": "value1",
				"field2": "value2",
			})
			Expect(fields).To(ContainSubstring(`field1=value1`))
			Expect(fields).To(ContainSubstring(f.Separator()))
			Expect(fields).To(ContainSubstring(`field2=value2`))
		})
	})
})
