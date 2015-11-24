package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	Global Logger
)

func init() {
	Global = NewDefaultLogger(DEBUG)
}

func LoadConfiguration(filename string) {
	Global.LoadConfiguration(filename)
}

func AddFilter(name string, lvl Level, writer LogWriter) {
	Global.AddFilter(name, lvl, writer)
}

func Close() {
	Global.Close()
}

func Crash(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(FATAL, strings.Repeat(" %v", len(args))[1:], args...)
	}
	panic(args)
}

func Crashf(format string, args ...interface{}) {
	Global.intLogf(FATAL, format, args...)
	Global.Close()
	panic(fmt.Sprintf(format, args...))
}

func Exit(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(ERROR, strings.Repeat(" %v", len(args))[1:], args...)
	}
	Global.Close()
	os.Exit(0)
}

func Exitf(format string, args ...interface{}) {
	Global.intLogf(ERROR, format, args...)
	Global.Close()
	os.Exit(0)
}

func Stderr(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(ERROR, strings.Repeat(" %v", len(args))[1:], args...)
	}
}

func Stderrf(format string, args ...interface{}) {
	Global.intLogf(ERROR, format, args...)
}

func Stdout(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(INFO, strings.Repeat(" %v", len(args))[1:], args...)
	}
}

func Stdoutf(format string, args ...interface{}) {
	Global.intLogf(INFO, format, args...)
}

func Log(lvl Level, source, message string) {
	Global.Log(lvl, source, message)
}

func Logf(lvl Level, format string, args ...interface{}) {
	Global.intLogf(lvl, format, args...)
}

func Logc(lvl Level, closure func() string) {
	Global.intLogc(lvl, closure)
}

func Debug(arg0 interface{}, args ...interface{}) {
	const (
		lvl = DEBUG
	)
	switch first := arg0.(type) {
	case string:
		Global.intLogf(lvl, first, args...)
	case func() string:
		Global.intLogc(lvl, first)
	default:
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func Info(arg0 interface{}, args ...interface{}) {
	const (
		lvl = INFO
	)
	switch first := arg0.(type) {
	case string:
		Global.intLogf(lvl, first, args...)
	case func() string:
		Global.intLogc(lvl, first)
	default:
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func Warn(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = WARNING
	)
	switch first := arg0.(type) {
	case string:
		Global.intLogf(lvl, first, args...)
		return errors.New(fmt.Sprintf(first, args...))
	case func() string:
		str := first()
		Global.intLogf(lvl, "%s", str)
		return errors.New(str)
	default:
		Global.intLogf(lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
		return errors.New(fmt.Sprint(first) + fmt.Sprintf(strings.Repeat(" %v", len(args)), args...))
	}
	return nil
}

func Error(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = ERROR
	)
	switch first := arg0.(type) {
	case string:
		Global.intLogf(lvl, first, args...)
		return errors.New(fmt.Sprintf(first, args...))
	case func() string:
		str := first()
		Global.intLogf(lvl, "%s", str)
		return errors.New(str)
	default:
		Global.intLogf(lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
		return errors.New(fmt.Sprint(first) + fmt.Sprintf(strings.Repeat(" %v", len(args)), args...))
	}
	return nil
}

func Fatal(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = FATAL
	)
	switch first := arg0.(type) {
	case string:
		Global.intLogf(lvl, first, args...)
		return errors.New(fmt.Sprintf(first, args...))
	case func() string:
		str := first()
		Global.intLogf(lvl, "%s", str)
		return errors.New(str)
	default:
		Global.intLogf(lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
		return errors.New(fmt.Sprint(first) + fmt.Sprintf(strings.Repeat(" %v", len(args)), args...))
	}
	return nil
}
