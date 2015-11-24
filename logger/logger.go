package logger

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	FATAL
)

var (
	levelStrings = [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
)

func (l Level) String() string {
	if l < 0 || int(l) > len(levelStrings) {
		return "UNKNOWN"
	}
	return levelStrings[int(l)]
}

var (
	LogBufferLength = 32
)

type LogRecord struct {
	Level   Level
	Created time.Time
	Source  string // 消息源
	Message string // 消息体
}

type LogWriter interface {
	LogWrite(rec *LogRecord)
	Close()
}

type Filter struct {
	Level Level
	LogWriter
}

type Logger map[string]*Filter

func NewDefaultLogger(lvl Level) Logger {
	return Logger{
		"stdout": &Filter{lvl, NewConsoleLogWriter()},
	}
}

func (log Logger) Close() {
	for name, filt := range log {
		filt.Close()
		delete(log, name)
	}
}

func (log Logger) AddFilter(name string, lvl Level, writer LogWriter) Logger {
	log[name] = &Filter{lvl, writer}
	return log
}

func (log Logger) intLogf(lvl Level, format string, args ...interface{}) {
	skip := true

	for _, filt := range log {
		if lvl >= filt.Level {
			skip = false
			break
		}
	}
	if skip {
		return
	}

	pc, _, lineno, ok := runtime.Caller(2)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
	}

	msg := format

	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Source:  src,
		Message: msg,
	}

	for _, filt := range log {
		if lvl < filt.Level {
			continue
		}
		filt.LogWrite(rec)
	}
}

func (log Logger) intLogc(lvl Level, closure func() string) {
	skip := true

	for _, filt := range log {
		if lvl > filt.Level {
			skip = false
			break
		}
	}
	if skip {
		return
	}

	pc, _, lineno, ok := runtime.Caller(2)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
	}

	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Source:  src,
		Message: closure(),
	}

	for _, filt := range log {
		if lvl < filt.Level {
			continue
		}
		filt.LogWrite(rec)
	}
}

func (log Logger) Log(lvl Level, source, message string) {
	skip := true

	for _, filt := range log {
		if lvl >= filt.Level {
			skip = false
			break
		}
	}
	if skip {
		return
	}

	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Source:  source,
		Message: message,
	}

	for _, filt := range log {
		if lvl < filt.Level {
			continue
		}
		filt.LogWrite(rec)
	}
}

func (log Logger) Logf(lvl Level, format string, args ...interface{}) {
	log.intLogf(lvl, format, args...)
}

func (log Logger) Logc(lvl Level, closure func() string) {
	log.intLogc(lvl, closure)
}

func (log Logger) Debug(arg0 interface{}, args ...interface{}) {
	const (
		lvl = DEBUG
	)
	switch first := arg0.(type) {
	case string:
		log.intLogf(lvl, first, args...)
	case func() string:
		log.intLogc(lvl, first)
	default:
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (log Logger) Info(arg0 interface{}, args ...interface{}) {
	const (
		lvl = INFO
	)
	switch first := arg0.(type) {
	case string:
		log.intLogf(lvl, first, args...)
	case func() string:
		log.intLogc(lvl, first)
	default:
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (log Logger) Warn(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = WARNING
	)
	var msg string
	switch first := arg0.(type) {
	case string:
		msg = fmt.Sprintf(first, args...)
	case func() string:
		msg = first()
	default:
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	log.intLogf(lvl, msg)
	return errors.New(msg)
}

func (log Logger) Error(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = ERROR
	)
	var msg string
	switch first := arg0.(type) {
	case string:
		msg = fmt.Sprintf(first, args...)
	case func() string:
		msg = first()
	default:
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	log.intLogf(lvl, msg)
	return errors.New(msg)
}

func (log Logger) Fatal(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = FATAL
	)
	var msg string
	switch first := arg0.(type) {
	case string:
		msg = fmt.Sprintf(first, args...)
	case func() string:
		msg = first()
	default:
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	log.intLogf(lvl, msg)
	return errors.New(msg)
}
