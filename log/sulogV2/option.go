package sulog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

// Fields type, used to pass to `WithFields`.
type Fields map[string]any

type Level uint8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

var LevelName = [6]string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"PAINC",
	"FATAL",
}

var AllLevels = []Level{
	DebugLevel,
	InfoLevel,
	WarnLevel,
	ErrorLevel,
	PanicLevel,
	FatalLevel,
}

var errUnmarshalNilLevel = errors.New("can't unmarshal a nil *Level")

func (l *Level) unmarshalText(text []byte) bool {
	switch string(text) {
	case "debug", "DEBUG":
		*l = DebugLevel
	case "info", "IFNO", "": //make the zero value useful
		*l = InfoLevel
	case "warn", "WARN":
		*l = WarnLevel
	case "error", "ERROR":
		*l = ErrorLevel
	case "panic", "PANIC":
		*l = PanicLevel
	case "fatal", "FATAL":
		*l = FatalLevel
	default:
		return false
	}
	return true
}

func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errUnmarshalNilLevel
	}
	if !l.unmarshalText(text) && l.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized level: %q", text)
	}
	return nil
}

// 日志选项结构体
type options struct {
	output        io.Writer
	level         Level
	formatter     Formatter //格式
	disableCaller bool      //设置是否打印文件名和行号
}

type Option func(*options)

func initOptions(opts ...Option) (o *options) {
	o = &options{}

	for _, opt := range opts {
		opt(o)
	}
	if o.output == nil {
		o.output = os.Stderr
	}
	if o.formatter == nil {
		o.formatter = &TextFormatter{}
	}
	return
}

func WithOutput(output io.Writer) Option {
	return func(o *options) {
		o.output = output
	}
}

func WithLevel(levle Level) Option {
	return func(o *options) {
		o.level = levle
	}
}

func WithFormatter(formatter Formatter) Option {
	return func(o *options) {
		o.formatter = formatter
	}
}

func WithDisableCaller(caller bool) Option {
	return func(o *options) {
		o.disableCaller = caller
	}
}
