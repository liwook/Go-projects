package sulog

import (
	"fmt"
	"os"
	"sync"
)

type logger struct {
	opt       *options
	mu        sync.Mutex
	entryPool *sync.Pool
}

var std = New()

func New(opts ...Option) *logger {
	logger := &logger{opt: initOptions(opts...)}
	logger.entryPool = &sync.Pool{New: func() any { return entry(logger) }}
	return logger
}

func StdLogger() *logger {
	return std
}

func SetOptions(opts ...Option) {
	std.SetOptions(opts...)
}

func (l *logger) SetOptions(opts ...Option) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, opt := range opts {
		opt(l.opt)
	}
}

// func Writer() io.Writer {
// 	return std
// }

// func (l *logger) Writer() io.Writer {
// 	return l
// }

// func (l *logger) Write(data []byte) (int, error) {
// 	l.entry().log(l.opt.level, FmtEmptySeparate, *(*string)(unsafe.Pointer(&data)))
// 	return 0, nil
// }

func (l *logger) entry() *Entry {
	return l.entryPool.Get().(*Entry)
}
func (l *logger) releaseEntry(e *Entry) {
	e.DataMap = map[string]any{}
	e.Line, e.File, e.Func = 0, "", ""
	e.Buffer.Reset()
	l.entryPool.Put(e)
}

func (l *logger) log(level Level, args ...any) {
	if l.opt.level > level { //日志等级不符合
		return
	}
	newEntry := l.entry()
	defer l.releaseEntry(newEntry)
	newEntry.Log(level, args...)
}

func (l *logger) logf(level Level, fomat string, args ...any) {
	if l.opt.level > level { //日志等级不符合
		return
	}
	newEntry := l.entry()
	defer l.releaseEntry(newEntry)
	newEntry.Logf(level, fomat, args...)
}

// 日志打印方法
func (l *logger) Debug(args ...any) {
	l.log(DebugLevel, args...)
}

func (l *logger) Info(args ...any) {
	l.log(InfoLevel, args...)
}

func (l *logger) Warn(args ...any) {
	l.log(WarnLevel, args...)
}

func (l *logger) Error(args ...any) {
	l.log(ErrorLevel, args...)
}

func (l *logger) Panic(args ...any) {
	l.log(PanicLevel, args...)
	panic(fmt.Sprint(args...))
}

func (l *logger) Fatal(args ...any) {
	l.log(FatalLevel, args...)
	os.Exit(1)
}

// 自定义格式
func (l *logger) Debugf(format string, args ...any) {
	l.logf(DebugLevel, format, args...)
}

func (l *logger) Infof(format string, args ...any) {
	l.logf(InfoLevel, format, args...)
}
func (l *logger) Warnf(format string, args ...any) {
	l.logf(WarnLevel, format, args...)
}

func (l *logger) Errorf(format string, args ...any) {
	l.logf(ErrorLevel, format, args...)
}

func (l *logger) Panicf(format string, args ...any) {
	l.logf(PanicLevel, format, args...)
	panic(fmt.Sprintf(format, args...))
}

func (l *logger) Fatalf(format string, args ...any) {
	l.logf(FatalLevel, format, args...)
	os.Exit(1)
}

// std logger 全局变量
func Debug(args ...any) {
	std.log(DebugLevel, args...)
}
func Info(args ...any) {
	std.log(InfoLevel, args...)
}
func Warn(args ...any) {
	std.log(WarnLevel, args...)
}
func Error(args ...any) {
	std.log(ErrorLevel, args...)
}
func Panic(args ...any) {
	std.log(PanicLevel, args...)
	panic(fmt.Sprint(args...))

}
func Fatal(args ...any) {
	std.log(FatalLevel, args...)
	os.Exit(1)
}

// 带格式的
func Debugf(format string, args ...any) {
	std.logf(DebugLevel, format, args...)
}
func Infof(format string, args ...any) {
	std.logf(InfoLevel, format, args...)
}
func Warnf(format string, args ...any) {
	std.logf(WarnLevel, format, args...)
}
func Errorf(format string, args ...any) {
	std.logf(ErrorLevel, format, args...)
}
func Panicf(format string, args ...any) {
	std.logf(PanicLevel, format, args...)
	panic(fmt.Sprintf(format, args...))
}
func Fatalf(format string, args ...any) {
	std.logf(FatalLevel, format, args...)
	os.Exit(1)
}
