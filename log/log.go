package log

import (
	"io/ioutil"
	"log"
	"os"
)

var Access *log.Logger
var Debug *log.Logger
var Error *log.Logger
var Info *log.Logger

func init() {
	Access = log.New(os.Stdout, "access ", log.LstdFlags)
	Debug = log.New(ioutil.Discard, "debug  ", log.LstdFlags|log.Llongfile)
	Error = log.New(os.Stderr, "error  ", log.LstdFlags|log.Llongfile)
	Info = log.New(os.Stderr, "info   ", log.LstdFlags)
}

func EnableDebug() {
	Debug = log.New(os.Stdout, "debug  ", log.LstdFlags|log.Llongfile)
}
