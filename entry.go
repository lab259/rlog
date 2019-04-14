package rlog

type EntryCallerInfo struct {
	PID          int
	GID          uint64
	FileName     string
	Line         int
	FunctionName string
}

type Entry struct {
	Time        string
	CallerInfo  EntryCallerInfo
	Level       Level
	TraceLevel  int
	FieldsCache string
	Fields      FieldsArr
	Message     string
}

func (entry *Entry) Reset() {
	entry.FieldsCache = ""
	entry.Message = ""
}
