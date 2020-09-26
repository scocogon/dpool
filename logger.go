package dpool

import (
	"log"
	"os"
)

type Logger interface {
	Printf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

var dlog = log.New(os.Stdout, "", log.LstdFlags)
