package rlog

import (
	"time"
)

type EntryCallerInfo struct {
	PID          int
	GID          uint64
	FileName     string
	Line         int
	FunctionName string
}

type Entry struct {
	Time       time.Time
	CallerInfo EntryCallerInfo
	Level      Level
	TraceLevel int
	Fields     string
	Message    string
}

func (entry *Entry) Reset() {
	entry.Fields = ""
	entry.Message = ""
}
