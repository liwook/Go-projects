package sulog

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Entry struct {
	logger *logger
	Buffer *bytes.Buffer
	// DataMap map[string]any //为了日志是json格式使用的

	Data    Fields //保存WithField中的数据
	Level   Level
	Time    time.Time
	File    string
	Line    int
	Func    string
	Message string
}

func entry(logger *logger) *Entry {
	return &Entry{
		logger: logger,
		// Buffer:  new(bytes.Buffer),
		// DataMap: make(map[string]any, 5)
		Data: make(map[string]any, 5),
	}
}

func (e *Entry) Log(depth int, level Level, args ...any) {
	if e.Level > level { //日志等级不符合
		return
	}
	e.log(depth, level, fmt.Sprint(args...))
}

func (e *Entry) Logf(depth int, level Level, format string, args ...any) {
	if e.Level > level { //日志等级不符合
		return
	}
	e.log(depth, level, fmt.Sprintf(format, args...))
}

func (e *Entry) log(depth int, level Level, msg string) {
	e.Time = time.Now()
	e.Level = level
	e.Message = msg

	if !e.logger.opt.disableCaller {
		if pc, file, line, ok := runtime.Caller(depth); !ok {
			e.File = "???"
			e.Func = "???"
		} else {
			e.File, e.Line, e.Func = file, line, runtime.FuncForPC(pc).Name()
			e.Func = e.Func[strings.LastIndex(e.Func, "/")+1:]
		}
	}

	e.fireHooks()
	bufPool := e.logger.bufferPool
	buffer := bufPool.Get().(*bytes.Buffer)
	e.Buffer = buffer
	defer func() {
		e.Buffer = nil
		buffer.Reset()
		bufPool.Put(buffer)
	}()

	e.write()
}

func (e *Entry) write() {
	e.logger.mu.Lock()
	defer e.logger.mu.Unlock()
	e.logger.opt.formatter.Format(e)
	e.logger.opt.output.Write(e.Buffer.Bytes())
}

// withField
// Add a single field to the Entry.
func (entry *Entry) WithField(key string, value any) *Entry {
	return entry.WithFields(Fields{key: value})
}

// Add a map of fields to the Entry.
func (entry *Entry) WithFields(fields Fields) *Entry {
	data := make(Fields, len(entry.Data)+len(fields))
	//为了可以这样使用sulog.WithField("name","li").WithField("addr","zhong").Info(1)
	for k, v := range entry.Data {
		data[k] = v
	}

	for k, v := range fields {
		isErrField := false
		if t := reflect.TypeOf(v); t != nil {
			//如果value类型是函数类型，是不符合要求的
			if t.Kind() == reflect.Func || (t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Func) {
				isErrField = true
			}
		}

		if isErrField {
			tmp := fmt.Sprintf("can not add field %q", k)
			fmt.Println(tmp)
		} else {
			data[k] = v
		}
	}

	return &Entry{logger: entry.logger, Data: data}
}

// entry打印方法
func (e *Entry) Debug(args ...any) {
	e.Log(3, DebugLevel, args...)
}

func (e *Entry) Info(args ...any) {
	e.Log(3, InfoLevel, args...)
}

func (e *Entry) Warn(args ...any) {
	e.Log(3, WarnLevel, args...)
}

func (e *Entry) Error(args ...any) {
	e.Log(3, ErrorLevel, args...)
}

func (e *Entry) Panic(args ...any) {
	e.Log(3, PanicLevel, args...)
	panic(fmt.Sprint(args...))
}

func (e *Entry) Fatal(args ...any) {
	e.Log(3, FatalLevel, args...)
	os.Exit(1)
}

// 自定义格式
func (e *Entry) Debugf(format string, args ...any) {
	e.Logf(3, DebugLevel, format, args...)
}

func (e *Entry) Infof(format string, args ...any) {
	e.Logf(3, InfoLevel, format, args...)
}
func (e *Entry) Warnf(format string, args ...any) {
	e.Logf(3, WarnLevel, format, args...)
}

func (e *Entry) Errorf(format string, args ...any) {
	e.Logf(3, ErrorLevel, format, args...)
}

func (e *Entry) Panicf(format string, args ...any) {
	e.Logf(3, PanicLevel, format, args...)
	panic(fmt.Sprintf(format, args...))
}

func (e *Entry) Fatalf(format string, args ...any) {
	e.Logf(3, FatalLevel, format, args...)
	os.Exit(1)
}

// hooks
func (entry *Entry) fireHooks() {
	var tmpHooks LevelHooks
	entry.logger.mu.Lock()
	tmpHooks = make(LevelHooks, len(entry.logger.Hooks))
	//进行拷贝
	for k, v := range entry.logger.Hooks {
		tmpHooks[k] = v
	}
	entry.logger.mu.Unlock()

	err := tmpHooks.Fire(entry.Level, entry) //执行hook
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fire hook: %v\n", err)
	}
}
