package main

import (
	"github.com/lab259/rlog/v2"
	"os"
)

func main() {
	stdOutLogger, err := rlog.NewLogger(rlog.Config{})
	if err != nil {
		panic(err)
	}
	stdOutLogger.SetOutput(os.Stdout)
	stdErrLogger, err := rlog.NewLogger(rlog.Config{})
	if err != nil {
		panic(err)
	}
	stdErrLogger.SetOutput(os.Stderr)
	stdOutLogger.Info("This log line goes to the STDOUT")
	stdErrLogger.Info("This log line goes to the STDERR")
}
