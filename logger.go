// logger.go
package main

import (
	"fmt"
	"log"
	"os"
)

type logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

var globalLogger *logger

func InitializeLogger() {
	// file options (Llongfile Lshortfile) are nice but useless for this wrapper
	// as we'd end up with `logger.go : line`
	flags := log.Ldate | log.Ltime | log.Lmicroseconds
	globalLogger = &logger{
		infoLogger:  log.New(os.Stdout, "", flags),
		errorLogger: log.New(os.Stderr, "", flags),
	}
}

func maybeInitializeLogger() {
	if globalLogger == nil {
		InitializeLogger()
	}
}

func formatLogMessage(prefix, format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s", prefix, fmt.Sprintf(format, args...))
}

func LogInfo(msg string, args ...interface{}) {
	// we'll allow the initialization to be overlooked
	maybeInitializeLogger()
	formattedMessage := formatLogMessage("INFO: ", msg, args...)
	globalLogger.infoLogger.Printf(formattedMessage)
}

func LogError(msg string, args ...interface{}) {
	maybeInitializeLogger()
	formattedMessage := formatLogMessage("ERROR: ", msg, args...)
	globalLogger.errorLogger.Printf(formattedMessage)
}
