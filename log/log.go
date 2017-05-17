package log

import (
	"os"
)

const (
	LogLevelINFO = 1 << iota
	LogLevelWARN
	LogLevelERR
	LogLevelDEBUG

	LogLevelNone = 0
	LogLevelStd  = LogLevelINFO | LogLevelWARN | LogLevelERR
	LogLevelWarn = LogLevelWARN | LogLevelERR
	LogLevelErr  = LogLevelERR
	LogLevelDbg  = LogLevelStd | LogLevelDEBUG
)

var (
	defaultLogger = NewColorLogger(os.Stdout, LogLevelDbg)

	Info  = defaultLogger.Info
	Warn  = defaultLogger.Warn
	Err   = defaultLogger.Err
	Debug = defaultLogger.Debug

	LogLevelMap = map[int]string{
		LogLevelINFO:  "INFO",
		LogLevelWARN:  "WARN",
		LogLevelERR:   "ERROR",
		LogLevelDEBUG: "DEBUG",
	}
)

//this is pretty much the only thing that isn't
//safe to run in multiple go routines, you should
//call SetDefaultLogger at the beginning of your main function
func SetDefaultLogger(l *Logger) {
	defaultLogger = l
	Info = defaultLogger.Info
	Warn = defaultLogger.Warn
	Err = defaultLogger.Err
	Debug = defaultLogger.Debug
}

type LogFormatter interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})

	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type Logger struct {
	Info  LogFormatter
	Warn  LogFormatter
	Err   LogFormatter
	Debug LogFormatter
}
