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
		Width:  60,
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
	return ""
}

func (formatter *defaultFormatter) formatField(entry *Entry, key string, data interface{}) string {
	cl := formatter.Color(entry)

	s := fmt.Sprint(data)
	if !(strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) && strings.ContainsAny(s, `" `) {
		replacer := strings.NewReplacer(`"`, `\"`, "\\", "\\\\")
		return fmt.Sprintf(`%s="%s"`, cl(key), replacer.Replace(s))
	}
	return fmt.Sprintf(`%s=%s`, cl(key), s)
}

func (formatter *defaultFormatter) FormatFields(fields FieldsArr) string {
	return ""
}

func (formatter *defaultFormatter) formatFields(entry *Entry) string {
	s := make([]string, len(entry.Fields))
	for i := 0; i < len(entry.Fields); i += 2 {
		key, ok := entry.Fields[i].(string)
		if !ok {
			key = fmt.Sprint(key)
		}
		s[i/2] = formatter.formatField(entry, key, entry.Fields[i+1])
	}
	return strings.Join(s, formatter.Separator())
}

func (formatter *defaultFormatter) Separator() string {
	return " "
}

func (formatter *defaultFormatter) Format(entry *Entry) []byte {
	output := AcquireOutput()

	// If a time is defined
	if entry.Time != "" {
		output = append(output, entry.Time...)
		output = append(output, formatter.Separator()...)
	}

	// If this entry has fields
	hasFields := len(entry.Fields) > 0

	// Get the first 4 letters of the level
	levelBytes := entry.Level.Bytes()[:4]
	output = append(output, formatter.Color(entry)(string(levelBytes))...)
	trcLvl := 0
	// If is a trace ...
	if entry.Level == levelTrace && entry.TraceLevel > notATrace {
		// Define the level
		trcLvl = entry.TraceLevel
	}
	output = append(output, fmt.Sprintf("[%05d]", trcLvl)...)

	// If this entry has caller info
	hasCallerInfo := entry.CallerInfo.PID > 0

	// If this entry has GID
	hasCallerInfoGID := entry.CallerInfo.GID > 0

	if hasCallerInfo {
		output = append(output, fmt.Sprintf(" [%-5d %s:%d %s] ", entry.CallerInfo.PID, entry.CallerInfo.FileName, entry.CallerInfo.Line, entry.CallerInfo.FunctionName)...)
	}

	if hasCallerInfoGID {
		output = append(output, fmt.Sprintf("(%05d) ", entry.CallerInfo.GID)...)
	}

	// Prints message, if it is not empty
	if entry.Message != "" {
		output = append(output, formatter.Separator()...)
		if hasFields || hasCallerInfo {
			output = append(output, fmt.Sprintf(fmt.Sprintf("%%-%ds", formatter.Width), entry.Message)...)
		} else {
			output = append(output, entry.Message...)
		}
	}

	if hasFields {
		output = append(output, formatter.Separator()...)
		output = append(output, formatter.formatFields(entry)...)
	}
	return append(output, '\n')
}
