package rlog

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type TextFormatter struct{}

var (
	textFormatterDatePrefix         = []byte(`date="`)
	textFormatterLevelPrefix        = []byte(`level=`)
	textFormatterMessagePrefix      = []byte(`msg="`)
	textFormatterSeparator          = byte(' ')
	textFormatterQuoteWithSeparator = []byte(`" `)
	textFormatterQuote              = byte('"')
	textFormatterQuoteArr           = []byte(`"`)
	textFormatterQuoteEscaped       = []byte(`\\"`)
	textFormatterLineEnding         = byte('\n')
	outputPool                      = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 512)
		},
	}
)

func (formatter *TextFormatter) Format(entry *Entry) []byte {
	output := AcquireOutput()

	if entry.Time != "" {
		output = append(output, textFormatterDatePrefix...)
		output = append(output, entry.Time...)
		output = append(output, textFormatterQuoteWithSeparator...)
	}
	output = append(output, textFormatterLevelPrefix...)
	levelBytes := entry.Level.Bytes()
	lw := levelWidth - len(levelBytes)
	output = append(output, levelBytes...)
	if entry.Level == levelTrace && entry.TraceLevel > notATrace {
		s := strconv.Itoa(entry.TraceLevel)
		lw -= 2 + len(s)
		output = append(output, '(')
		output = append(output, s...)
		output = append(output, ')')
	}

	if entry.FieldsCache != "" {
		output = append(output, textFormatterSeparator)
		output = append(output, entry.FieldsCache...)
	}

	if entry.Message != "" {
		output = append(output, textFormatterSeparator)
		output = append(output, textFormatterMessagePrefix...)
		output = append(output, bytes.Replace([]byte(entry.Message), textFormatterQuoteArr, textFormatterQuoteEscaped, -1)...)
	}
	return append(output, textFormatterQuote, textFormatterLineEnding)
}

func (formatter *TextFormatter) FormatField(key string, data interface{}) string {
	s := fmt.Sprint(data)
	if !(strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) && strings.ContainsAny(s, `" `) {
		replacer := strings.NewReplacer(`"`, `\"`, "\\", "\\\\")
		return fmt.Sprintf(`%s="%s"`, key, replacer.Replace(s))
	}
	return fmt.Sprintf(`%s=%s`, key, s)
}

func (formatter *TextFormatter) FormatFields(fields FieldsArr) string {
	s := make([]string, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			key = fmt.Sprint(key)
		}
		s[i/2] = formatter.FormatField(key, fields[i+1])
	}
	return strings.Join(s, formatter.Separator())
}

func (formatter *TextFormatter) Separator() string {
	return " "
}

var defaultTextFormatter TextFormatter
