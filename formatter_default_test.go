package rlog

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Formatter", func() {
	Describe("DefaultFormatter", func() {
		It("should format a field", func() {
			f := NewDefaultFormatter(os.Stderr)
			str := f.formatField(&Entry{
				Level: levelTrace,
			}, "key", "value")
			Expect(str).To(ContainSubstring(`key`))
			Expect(str).To(ContainSubstring(`value`))
		})

		It("should format a field escaping values", func() {
			f := NewDefaultFormatter(os.Stderr)
			Expect(f.formatField(&Entry{
				Level: levelTrace,
			}, "key", `value with "quotes"`)).To(Equal(`key="value with \"quotes\""`))
		})

		It("should format fields", func() {
			f := NewDefaultFormatter(os.Stderr)
			fields := f.formatFields(&Entry{
				Level: levelTrace,
				Fields: FieldsArr{
					"field1", "value1",
					"field2", "value2",
				},
			})
			Expect(fields).To(ContainSubstring(`field1=value1`))
			Expect(fields).To(ContainSubstring(f.Separator()))
			Expect(fields).To(ContainSubstring(`field2=value2`))
		})
	})
})
