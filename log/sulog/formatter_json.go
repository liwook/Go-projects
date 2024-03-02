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
	if !f.DisableTimestamp {
		if f.TimestampFormat == "" {
			f.TimestampFormat = time.RFC3339
		}
		e.DataMap["time"] = e.Time.Format(f.TimestampFormat)
	}

	e.DataMap["level"] = LevelNameMapping[e.Level]

	if e.File != "" {
		e.DataMap["file"] = e.File + ":" + strconv.Itoa(e.Line)
		e.DataMap["func"] = e.Func
	}

	e.DataMap["message"] = e.Message

	return sonic.ConfigDefault.NewEncoder(e.Buffer).Encode(e.DataMap)
}
