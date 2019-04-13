package rlog

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"

	"golang.org/x/crypto/ssh/terminal"
)

type Color func(a ...interface{}) string

type defaultFormatter struct {
	Colors map[Level]Color
	Width  uint
}

var DefaultFormatter = NewDefaultFormatter()

func clr(value ...color.Attribute) Color {
	c := color.New(value...)
	c.EnableColor()
	return c.Sprint
}

func NewDefaultFormatter() *defaultFormatter {
	return &defaultFormatter{
		Colors: map[Level]Color{
			levelNone:  fmt.Sprint,
			levelCrit:  clr(color.BgRed, color.FgWhite),
			levelErr:   clr(color.FgRed),
			levelWarn:  clr(color.FgYellow),
			levelInfo:  clr(color.FgCyan),
			levelDebug: clr(color.FgMagenta),
			levelTrace: clr(color.FgHiBlack),
		},
		Width: 100,
	}
}

func (formatter *defaultFormatter) isTTY() bool {
	return terminal.IsTerminal(int(os.Stdout.Fd()))
}

func (formatter *defaultFormatter) Color(entry *Entry) Color {
	cl, ok := formatter.Colors[entry.Level]
	if !ok {
		return fmt.Sprint
	}
	return cl
}

func (formatter *defaultFormatter) FormatField(key string, data interface{}) string {
	s := fmt.Sprint(data)
	if !(strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) && strings.ContainsAny(s, `" `) {
		replacer := strings.NewReplacer(`"`, `\"`, "\\", "\\\\")
		return fmt.Sprintf(`%s="%s"`, key, replacer.Replace(s))
	}
	return fmt.Sprintf(`%s=%s`, key, s)
}

func (formatter *defaultFormatter) FormatFields(fields Fields) string {
	s := make([]string, len(fields))
	i := 0
	for key, value := range fields {
		s[i] = formatter.FormatField(key, value)
		i++
	}
	return strings.Join(s, formatter.Separator())
}

func (formatter *defaultFormatter) Separator() string {
	return " "
}

func (formatter *defaultFormatter) Format(entry *Entry) []byte {
	output := AcquireOutput()

	if entry.Time != "" {
		output = append(output, entry.Time...)
		output = append(output, formatter.Separator()...)
	}

	levelBytes := entry.Level.Bytes()[:4]
	output = append(output, formatter.Color(entry)(string(levelBytes))...)
	trcLvl := 0
	if entry.Level == levelTrace && entry.TraceLevel > notATrace {
		trcLvl = entry.TraceLevel
	}
	output = append(output, fmt.Sprintf("[%05d]", trcLvl)...)
	if entry.Message != "" {
		output = append(output, formatter.Separator()...)
		if entry.Fields != "" {
			output = append(output, fmt.Sprintf(fmt.Sprintf("%%-%ds", formatter.Width), entry.Message)...)
		} else {
			output = append(output, entry.Message...)
		}
	}

	if entry.Fields != "" {
		output = append(output, formatter.Separator()...)
		output = append(output, entry.Fields...)
	}
	return append(output, '\n')
}
