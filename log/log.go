package log

import (
	"os"
)

const (
	//Base log levels
	LogLevelINFO = 1 << iota
	LogLevelWARN
	LogLevelERR
	LogLevelDEBUG

	//Common log level combinations
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

//SetDefaultLogger sets the default logger... Yeah really.
//This is pretty much the only thing that isn't safe to run
//in multiple go routines. You should call SetDefaultLogger
//in your init or at the beginning of your main function.
func SetDefaultLogger(l *Logger) {
	defaultLogger = l
	Info = defaultLogger.Info
	Warn = defaultLogger.Warn
	Err = defaultLogger.Err
	Debug = defaultLogger.Debug
}

//LogFormatter is the interface containing all the print methods needed for a typical Logger.
type LogFormatter interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})

	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

//Logger is a struct containing a LogFormatter for each of the four basic log levels.
type Logger struct {
	Info  LogFormatter
	Warn  LogFormatter
	Err   LogFormatter
	Debug LogFormatter
}
