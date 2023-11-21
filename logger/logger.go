package logger

type Logger interface {
	Fatalf(string, ...interface{})
	Fatal(...interface{})
	Panicf(string, ...interface{})
	Panic(...interface{})
	Errorf(string, ...interface{})
	Error(...interface{})
	Warnf(string, ...interface{})
	Warn(...interface{})
	Infof(string, ...interface{})
	Info(...interface{})
	Debugf(string, ...interface{})
	Debug(...interface{})
}
