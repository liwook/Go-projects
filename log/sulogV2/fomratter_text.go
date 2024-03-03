package sulog

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

type TextFormatter struct {
	DisableTimestamp bool
	TimestampFormat  string
}

// 格式是： 时间 日志等级 文件:所在行号 函数名称 日志内容
func (f *TextFormatter) Format(e *Entry) error {
	prefixFieldClashes(e.Data)

	if !f.DisableTimestamp {
		if f.TimestampFormat == "" {
			f.TimestampFormat = time.RFC3339
		}
		f.appendKeyValue(e.Buffer, KeyTime, e.Time.Format(f.TimestampFormat))
	}
	f.appendKeyValue(e.Buffer, KeyLevel, LevelName[e.Level])

	if !e.logger.opt.disableCaller {
		if e.File != "" {
			short := e.File
			for i := len(e.File) - 1; i > 0; i-- {
				if e.File[i] == '/' {
					short = e.File[i+1:]
					break
				}
			}

			f.appendKeyValue(e.Buffer, KeyFunc, short)
			f.appendKeyValue(e.Buffer, KeyFile, e.File+":"+strconv.Itoa(e.Line))
		}
	}

	f.appendKeyValue(e.Buffer, KeyMsg, e.Message)
	//加上WithField()中的
	for k, v := range e.Data {
		f.appendKeyValue(e.Buffer, k, v)
	}

	e.Buffer.WriteString("\n")

	return nil
}

func (f *TextFormatter) appendKeyValue(b *bytes.Buffer, key string, value any) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
}

func (f *TextFormatter) appendValue(b *bytes.Buffer, value any) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal)) //这样就是加""符号的
	}
}
