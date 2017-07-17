// Copyright 2017 Publit Sweden AB. All rights reserved.

// Publit API logger. Handles logging for internals in the PublitAPI SDKs.
// Has info and debug levels only.
// No error level logs because all actual errors propagate to the implementation.
// The debug level is a handled error (handled by logging).
package APILog

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

// Log output
//
// Default set as ioutil.Discard. An effective way to turn of logging.
var LogOutput io.Writer = ioutil.Discard

// Set Levels to log
//
// Default set to both info and debug.
var OutputLevel LogLevel = LEVEL_INFO | LEVEL_DEBUG

// Flags for logger
//
// Uses same flags as log.SetFlags().
var LogFlags int = log.Lshortfile

// LogLevel Bitmask
//
type LogLevel uint32

// Log Json Format
//
// Indicates if logs should be formatted to json.
//
// LogJsonFormat = true outputs:
//  file.go:1: {"level":"INFO","message":"\"Some informational message.\"","timestamp":"2017-07-14 16:12:03.425Z"}
// LogJsonFormat = false outputs:
//  file.go:1: [INFO] Some informational message.
var LogJsonFormat bool = true

// Only using two LogLevels: info and debug.
const (
	LEVEL_INFO LogLevel = 1 << iota
	LEVEL_DEBUG
)

// Log message headers.
const (
	LEVEL_STRING_INFO  = "info"
	LEVEL_STRING_DEBUG = "debug"
)

// APILog struct.
type APILog struct {
	L *log.Logger
}

// Creates new APILog with set log.logger.
func New() *APILog {
	logger := log.New(LogOutput, "", LogFlags)
	return &APILog{logger}
}

// Logs message.
func (a APILog) log(logHeader string, message interface{}, level LogLevel) {
	logMessage := ""
	if LogJsonFormat {
		logMessage = formatJSONLog(logHeader, message)
	} else {
		logMessage = fmt.Sprintf("[%s]: %v", strings.ToUpper(logHeader), message)
	}

	if OutputLevel.HasLevel(level) {
		a.L.Println(logMessage)
	}
}

// JsonLogMessage struct.
type jsonLogMessage struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// FormatJsonLog. Formats log message to json format.
func formatJSONLog(logHeader string, message interface{}) string {
	jm := jsonLogMessage{
		Level:     strings.ToUpper(logHeader),
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
	}

	m, err := json.Marshal(message)
	if err != nil {
		jm.Message = fmt.Sprint(message)
	} else {
		jm.Message = string(m)
	}

	str, err := json.Marshal(jm)
	if err != nil {
		fmt.Println(err)
	}

	return string(str)
}

// Creates debug log.
func (a APILog) Debug(message interface{}) {
	a.log(LEVEL_STRING_DEBUG, message, LEVEL_DEBUG)
}

// Creates info log.
func (a APILog) Info(message interface{}) {
	a.log(LEVEL_STRING_INFO, message, LEVEL_INFO)
}

// Checks if LogLevel flag is set. For bitmasking.
func (l LogLevel) HasLevel(level LogLevel) bool {
	return l&level != 0
}
