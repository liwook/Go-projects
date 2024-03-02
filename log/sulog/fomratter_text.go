package sulog

import (
	"fmt"
	"time"
)

type TextFormatter struct {
	DisableTimestamp bool
	TimestampFormat  string
}

// 格式是： 时间 日志等级 文件:所在行号 函数名称 日志内容
func (f *TextFormatter) Format(e *Entry) error {
	if !f.DisableTimestamp {
		if f.TimestampFormat == "" {
			f.TimestampFormat = time.RFC3339
		}
		e.Buffer.WriteString(fmt.Sprintf("%s %s", e.Time.Format(f.TimestampFormat), LevelNameMapping[e.Level]))
	} else {
		e.Buffer.WriteString(LevelNameMapping[e.Level])
	}
	if e.File != "" {
		short := e.File
		for i := len(e.File) - 1; i > 0; i-- {
			if e.File[i] == '/' {
				short = e.File[i+1:]
				break
			}
		}

		e.Buffer.WriteString(fmt.Sprintf(" %s:%d %s", short, e.Line, e.Func))
	}
	e.Buffer.WriteString(" ")

	e.Buffer.WriteString(e.Message)
	e.Buffer.WriteByte('\n')

	return nil
}
