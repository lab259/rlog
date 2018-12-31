package rlog

import (
	"bytes"
	"fmt"
	"strings"
)

type LogFormatter interface {
	Format(key string, data interface{}) string
	FormatFields(Fields) string
	Separator() string
	Line(date string, levelDecoration string, callerInfo string, msg string) string
}

type TextFormatter struct{}

func (formatter *TextFormatter) Line(date string, levelDecoration string, callerInfo string, msg string) string {
	buff := bytes.NewBuffer(nil)
	if date != "" {
		buff.WriteString(fmt.Sprintf(`date="%s" `, date))
	}
	buff.WriteString(fmt.Sprintf(`level="%s" %s%s`, levelDecoration, callerInfo, msg))
	return buff.String()
}

func (formatter *TextFormatter) Format(key string, data interface{}) string {
	return fmt.Sprintf(`%s="%s"`, key, strings.Replace(fmt.Sprint(data), `"`, `\\"`, -1))
}

func (formatter *TextFormatter) FormatFields(fields Fields) string {
	s := make([]string, len(fields))
	i := 0
	for key, value := range fields {
		s[i] = formatter.Format(key, value)
		i++
	}
	return strings.Join(s, formatter.Separator())
}

func (formatter *TextFormatter) Separator() string {
	return " "
}

var defaultTextFormatter TextFormatter
