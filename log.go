package rest

import "context"

type Logger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

func Nop(context.Context) Logger {
	return nop{}
}

type nop struct{}

func (no nop) Debugf(format string, v ...interface{}) {
}

func (no nop) Infof(format string, v ...interface{}) {
}

func (no nop) Warnf(format string, v ...interface{}) {
}

func (no nop) Errorf(format string, v ...interface{}) {
}
