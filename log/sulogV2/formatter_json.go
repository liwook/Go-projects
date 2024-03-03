package sulog

import (
	"strconv"
	"time"

	"github.com/bytedance/sonic"
)

type JsonFormatter struct {
	DisableTimestamp bool
	TimestampFormat  string
}

func (f *JsonFormatter) Format(e *Entry) error {
	data := make(Fields, len(e.Data)+5)
	prefixFieldClashes(e.Data)
	for k, v := range e.Data {
		data[k] = v
	}
	if !f.DisableTimestamp {
		if f.TimestampFormat == "" {
			f.TimestampFormat = time.RFC3339
		}
		data[KeyTime] = e.Time.Format(f.TimestampFormat)
	}

	data[KeyLevel] = LevelName[e.Level]

	if e.File != "" {
		data[KeyFile] = e.File + ":" + strconv.Itoa(e.Line)
		data[KeyFunc] = e.Func
	}

	data[KeyMsg] = e.Message

	return sonic.ConfigDefault.NewEncoder(e.Buffer).Encode(data)
}
