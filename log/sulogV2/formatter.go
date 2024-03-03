package sulog

import "time"

type fieldKey string

const (
	defaultTimestampFormat = time.RFC3339
	KeyMsg                 = "msg"
	KeyLevel               = "level"
	KeyTime                = "time"
	KeyFunc                = "func"
	KeyFile                = "file"
)

type Formatter interface {
	Format(entry *Entry) error
}

// FieldMap allows customization of the key names for default fields.
type FieldMap map[fieldKey]string

func prefixFieldClashes(data Fields) {
	for k, v := range data {
		switch k {
		case KeyMsg:
			data["fields."+KeyMsg] = v
			delete(data, KeyMsg)
		case KeyLevel:
			data["fields."+KeyLevel] = v
			delete(data, KeyLevel)
		case KeyFunc:
			data["fields."+KeyFunc] = v
			delete(data, KeyFunc)
		case KeyTime:
			data["fields."+KeyTime] = v
			delete(data, KeyTime)
		case KeyFile:
			data["fields."+KeyFile] = v
			delete(data, KeyFile)
		}
	}
}
