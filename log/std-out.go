package log

import (
	"bytes"
	"fmt"
	"io"
	goLog "log"
	"os"
)

const (
	colorLogDepth = 4
)

var (
	ansiColorPallet = map[string][]byte{
		"none":         []byte("\x1b[0m"),
		"black":        []byte("\x1b[0;30m"),
		"red":          []byte("\x1b[0;31m"),
		"green":        []byte("\x1b[0;32m"),
		"orange":       []byte("\x1b[0;33m"),
		"blue":         []byte("\x1b[0;34m"),
		"purple":       []byte("\x1b[0;35m"),
		"cyan":         []byte("\x1b[0;36m"),
		"light-gray":   []byte("\x1b[0;37m"),
		"dark-gray":    []byte("\x1b[1;30m"),
		"light-red":    []byte("\x1b[1;31m"),
		"light-green":  []byte("\x1b[1;32m"),
		"yellow":       []byte("\x1b[1;33m"),
		"light-blue":   []byte("\x1b[1;34m"),
		"light-purple": []byte("\x1b[1;35m"),
		"light-cyan":   []byte("\x1b[1;36m"),
		"white":        []byte("\x1b[1;37m"),
	}

	logColors = map[int][]byte{
		LogLevelINFO:  ansiColorPallet["white"],
		LogLevelWARN:  ansiColorPallet["orange"],
		LogLevelERR:   ansiColorPallet["red"],
		LogLevelDEBUG: ansiColorPallet["light-blue"],
	}
)

//NewColorLogger creates a new *Logger that outputs different log levels as different ANSI colors
func NewColorLogger(out io.Writer, logLevel int) *Logger {
	return newLogger(out, logLevel, true)
}

//NewLogger creates a new *Logger without ANSI colors
func NewLogger(out io.Writer, logLevel int) *Logger {
	return newLogger(out, logLevel, false)
}

func newLogger(out io.Writer, logLevel int, color bool) *Logger {
	logger := &Logger{
		Info:  newColorLogFmt(out, logLevel, LogLevelINFO, color),
		Warn:  newColorLogFmt(out, logLevel, LogLevelWARN, color),
		Err:   newColorLogFmt(out, logLevel, LogLevelERR, color),
		Debug: newColorLogFmt(out, logLevel, LogLevelDEBUG, color),
	}

	return logger
}

type colorLogFmt struct {
	out      io.Writer
	logger   *goLog.Logger
	logLevel int
	fmtLevel int
	color    bool
}

func newColorLogFmt(out io.Writer, logLevel, fmtLevel int, color bool) *colorLogFmt {
	clf := &colorLogFmt{
		logLevel: logLevel,
		fmtLevel: fmtLevel,
		out:      out,
		color:    color,
	}
	lFlags := goLog.LstdFlags

	if fmtLevel != LogLevelINFO {
		lFlags |= goLog.Lshortfile
	}
	clf.logger = goLog.New(out, padStr(LogLevelMap[fmtLevel], 6), lFlags)

	return clf
}

func (lf *colorLogFmt) canLog() bool {
	return (lf.logLevel & lf.fmtLevel) == lf.fmtLevel
}

func (lf *colorLogFmt) doOutput(pType, format string, v ...interface{}) {
	switch pType {
	case "Print", "Fatal":
		lf.logger.Output(colorLogDepth, fmt.Sprint(v...))
	case "Printf", "Fatalf":
		lf.logger.Output(colorLogDepth, fmt.Sprintf(format, v...))
	case "Println", "Fatalln":
		lf.logger.Output(colorLogDepth, fmt.Sprintln(v...))
	default:
		lf.logger.Output(colorLogDepth, fmt.Sprint(v...))
	}
}

func (lf *colorLogFmt) doPrint(pType, format string, v ...interface{}) {
	switch pType {
	case "Fatal", "Fatalf", "Fatalln":
		lf.setErrColor()
		lf.doOutput(pType, format, v...)
		lf.dropColor()
		os.Exit(1)
	case "Print", "Printf", "Println":
		if !lf.canLog() {
			return
		}
		lf.setFmtColor()
		lf.doOutput(pType, format, v...)
		lf.dropColor()
	}
}

//Fatal logs don't care what the logLevel is. They print using the
//LogLevelERR color (if using the color logger) and exit with a satus of 1
func (lf *colorLogFmt) Fatal(v ...interface{}) {
	lf.doPrint("Fatal", "", v...)
}

func (lf *colorLogFmt) Fatalf(format string, v ...interface{}) {
	lf.doPrint("Fatalf", format, v...)
}

func (lf *colorLogFmt) Fatalln(v ...interface{}) {
	lf.doPrint("Fatalln", "", v...)
}

//Print operates the same as fmt.Print but with a log prefix
func (lf *colorLogFmt) Print(v ...interface{}) {
	lf.doPrint("Print", "", v...)
}

//Printf operates the same as fmt.Printf but with a log prefix
func (lf *colorLogFmt) Printf(format string, v ...interface{}) {
	lf.doPrint("Printf", format, v...)
}

//Println operates the same as fmt.Println but with a log prefix
func (lf *colorLogFmt) Println(v ...interface{}) {
	lf.doPrint("Println", "", v...)
}

func (lf *colorLogFmt) setColor(color []byte) {
	if !lf.color {
		return
	}

	lf.out.Write(color)
}

func (lf *colorLogFmt) setErrColor() {
	lf.setColor(logColors[LogLevelERR])
}

func (lf *colorLogFmt) setFmtColor() {
	lf.setColor(logColors[lf.fmtLevel])
}

func (lf *colorLogFmt) dropColor() {
	lf.setColor(ansiColorPallet["none"])
}

func padStr(s string, length int) string {
	b := bytes.NewBuffer(nil)
	b.WriteString(s)
	for {
		if b.Len() >= length {
			return b.String()
		}
		b.WriteString(" ")
	}
}
