package tupliplib

import (
	"github.com/francoispqt/onelog"
	"os"
)

var logger *onelog.Logger

func init() {
	logger = onelog.New(
		os.Stdout,
		onelog.WARN|onelog.ERROR|onelog.FATAL,
	)
}

// UseLogger sets a custom Logger for the package.
func UseLogger(newLogger *onelog.Logger) {
	logger = newLogger
}
