package sulog

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"time"
)

type Entry struct {
	logger  *logger
	Buffer  *bytes.Buffer
	DataMap map[string]any //为了日志是json格式使用的
	Level   Level
	Time    time.Time
	File    string
	Line    int
	Func    string
	Message string
}

func entry(logger *logger) *Entry {
	return &Entry{
		logger:  logger,
		Buffer:  new(bytes.Buffer),
		DataMap: make(map[string]any, 5)}
}

func (e *Entry) Log(level Level, args ...any) {
	e.log(level, fmt.Sprint(args...))
}

func (e *Entry) Logf(level Level, format string, args ...any) {
	e.log(level, fmt.Sprintf(format, args...))
}

func (e *Entry) log(level Level, msg string) {
	e.Time = time.Now()
	e.Level = level
	e.Message = msg

	if !e.logger.opt.disableCaller {
		if pc, file, line, ok := runtime.Caller(4); !ok {
			e.File = "???"
			e.Func = "???"
		} else {
			e.File, e.Line, e.Func = file, line, runtime.FuncForPC(pc).Name()
			e.Func = e.Func[strings.LastIndex(e.Func, "/")+1:]
		}
	}

	e.write()
}

func (e *Entry) write() {
	e.logger.mu.Lock()
	defer e.logger.mu.Unlock()
	e.logger.opt.formatter.Format(e)

	e.logger.opt.output.Write(e.Buffer.Bytes())
}
