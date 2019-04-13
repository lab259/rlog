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

func clr(file *os.File, value ...color.Attribute) Color {
	c := color.New(value...)
	if file != nil && terminal.IsTerminal(int(file.Fd())) {
		c.EnableColor()
	} else {
		c.DisableColor()
	}
	return c.Sprint
}

func newClrs(file *os.File) map[Level]Color {
	return map[Level]Color{
		levelNone:  fmt.Sprint,
		levelCrit:  clr(file, color.BgRed, color.FgWhite),
		levelErr:   clr(file, color.FgRed),
		levelWarn:  clr(file, color.FgYellow),
		levelInfo:  clr(file, color.FgCyan),
		levelDebug: clr(file, color.FgMagenta),
		levelTrace: clr(file, color.FgHiBlack),
	}
}

func NewDefaultFormatter(file *os.File) *defaultFormatter {
	return &defaultFormatter{
		Colors: newClrs(file),
		Width:  100,
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
