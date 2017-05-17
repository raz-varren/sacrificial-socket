package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

type jLog struct {
	Timestamp time.Time `json:"ts"`
	File      string    `json:"file"`
	Line      int       `json:"line"`
	Level     string    `json:"level"`
	Fatal     bool      `json:"fatal"`
	Msg       string    `json:"msg"`
}

func (jl *jLog) Marshal() []byte {
	data, err := json.Marshal(jl)
	if err != nil {
		return []byte(fmt.Sprintf(`{"Error":"%s", "Line": %d, "File":"%s"}`, err, jl.Line, jl.File))
	}
	return data
}

type jsonLogFmt struct {
	out             io.Writer
	logLevel        int
	logLevelAllowed int
}

func (j *jsonLogFmt) canLog() bool {
	return (j.logLevel & j.logLevelAllowed) == j.logLevelAllowed
}

func (j *jsonLogFmt) writeJLog(msg string, fatal bool) {
	jl := newJLog(msg, LogLevelMap[j.logLevelAllowed], fatal, 3)
	data := jl.Marshal()
	j.out.Write(append(data, '\n'))
}

func newJLog(msg, level string, fatal bool, callDepth int) *jLog {
	jl := &jLog{
		Timestamp: time.Now(),
		Level:     level,
		Fatal:     fatal,
		Msg:       msg,
	}
	_, file, line, ok := runtime.Caller(callDepth)

	jl.File = file
	jl.Line = line

	if !ok {
		jl.File = "???"
		jl.Line = -1
	}

	return jl
}

func (j *jsonLogFmt) Print(v ...interface{}) {
	if !j.canLog() {
		return
	}
	msg, fatal := fmt.Sprint(v...), false
	msg = strings.TrimRight(msg, "\n")
	j.writeJLog(msg, fatal)
}

func (j *jsonLogFmt) Printf(format string, v ...interface{}) {
	if !j.canLog() {
		return
	}

	msg, fatal := fmt.Sprintf(format, v...), false
	msg = strings.TrimRight(msg, "\n")

	j.writeJLog(msg, fatal)
}

func (j *jsonLogFmt) Println(v ...interface{}) {
	if !j.canLog() {
		return
	}

	msg, fatal := fmt.Sprintln(v...), false
	msg = strings.TrimRight(msg, "\n")

	j.writeJLog(msg, fatal)
}

func (j *jsonLogFmt) Fatal(v ...interface{}) {
	msg, fatal := fmt.Sprint(v...), true
	msg = strings.TrimRight(msg, "\n")

	j.writeJLog(msg, fatal)
	os.Exit(1)
}

func (j *jsonLogFmt) Fatalf(format string, v ...interface{}) {
	msg, fatal := fmt.Sprintf(format, v...), true
	msg = strings.TrimRight(msg, "\n")

	j.writeJLog(msg, fatal)
	os.Exit(1)
}

func (j *jsonLogFmt) Fatalln(v ...interface{}) {
	msg, fatal := fmt.Sprintln(v...), true
	msg = strings.TrimRight(msg, "\n")

	j.writeJLog(msg, fatal)
	os.Exit(1)
}

func NewJSONLogger(out io.Writer, logLevel int) *Logger {
	info := &jsonLogFmt{
		out:             out,
		logLevel:        logLevel,
		logLevelAllowed: LogLevelINFO,
	}

	warn := &jsonLogFmt{
		out:             out,
		logLevel:        logLevel,
		logLevelAllowed: LogLevelWARN,
	}

	err := &jsonLogFmt{
		out:             out,
		logLevel:        logLevel,
		logLevelAllowed: LogLevelERR,
	}

	debug := &jsonLogFmt{
		out:             out,
		logLevel:        logLevel,
		logLevelAllowed: LogLevelDEBUG,
	}

	return &Logger{
		Info:  info,
		Warn:  warn,
		Err:   err,
		Debug: debug,
	}
}
