// logger.go
package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

var globalLogger *logger

func InitializeLogger() {
	globalLogger = &logger{
		infoLogger:  log.New(os.Stdout, "", 0),
		errorLogger: log.New(os.Stderr, "", 0),
	}
}

func maybeInitializeLogger() {
	if globalLogger == nil {
		InitializeLogger()
	}
}

func formatLogMessage(prefix, format string, args ...interface{}) string {
	now := time.Now()
	timestamp := now.Format("2006-01-02 15:04:05.000") // Format with milliseconds
	return fmt.Sprintf("%s %s%s", timestamp, prefix, fmt.Sprintf(format, args...))
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
