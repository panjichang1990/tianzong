package tzlog

import (
	"fmt"
)

type ILog interface {
	Debug(format string, v ...interface{})
	Warning(format string, v ...interface{})
	Error(format string, v ...interface{})
	Info(format string, v ...interface{})
}

var logDriver ILog = &fmtLog{}

type fmtLog struct {
}

func (fl *fmtLog) Debug(format string, v ...interface{}) {
	fmt.Println(fmt.Sprintf(format, v...))
}

func (fl *fmtLog) Warning(format string, v ...interface{}) {
	fmt.Println(fmt.Sprintf(format, v...))
}
func (fl *fmtLog) Error(format string, v ...interface{}) {
	fmt.Println(fmt.Sprintf(format, v...))
}
func (fl *fmtLog) Info(format string, v ...interface{}) {
	fmt.Println(fmt.Sprintf(format, v...))
}

func SetLogger(logger ILog) {
	logDriver = logger
}

func D(tpl string, args ...interface{}) {
	logDriver.Debug(tpl, args...)

}

func I(tpl string, args ...interface{}) {
	logDriver.Info(tpl, args...)
}

func E(tpl string, args ...interface{}) {
	logDriver.Error(tpl, args...)
}

func W(tpl string, args ...interface{}) {
	logDriver.Warning(tpl, args...)
}
