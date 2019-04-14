package rlog

type LogFormatter interface {
	FormatField(key string, data interface{}) string
	FormatFields(FieldsArr) string
	Separator() string
	Format(entry *Entry) []byte
}

func AcquireOutput() []byte {
	return outputPool.Get().([]byte)[0:0]
}

func ReleaseOutput(data []byte) {
	outputPool.Put(data)
}
