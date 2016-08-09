package log

import (
	l "log"
	"os"
)

var Err = l.New(os.Stderr, "ERROR: ", l.Ldate|l.Ltime|l.Lshortfile)
var Info = l.New(os.Stdout, "INFO:  ", l.Ldate|l.Ltime)

func CheckFatal(err error) {
	if err != nil {
		Err.Output(2, err.Error())
		os.Exit(1)
	}
}
