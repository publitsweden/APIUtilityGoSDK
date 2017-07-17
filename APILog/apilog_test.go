package APILog_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/publitsweden/APIUtilityGoSDK/APILog"
	"strings"
	"testing"
	"os"
	"log"
)

func TestCanWriteLog(t *testing.T) {
	t.Run(
		"using default json format",
		func(t *testing.T) {
			tests := map[string]func(a *APILog, message interface{}){
				LEVEL_STRING_INFO: func(a *APILog, message interface{}) {
					a.Info(message)
				},
				LEVEL_STRING_DEBUG: func(a *APILog, message interface{}) {
					a.Debug(message)
				},
			}

			var b bytes.Buffer
			//Set output
			LogOutput = &b
			a := New()

			for k, v := range tests {
				t.Run(
					k,
					func(t *testing.T) {
						defer b.Reset()
						message := struct{ Name string }{Name: "some name"}

						v(a, message)

						contents := b.String()

						// Tests that the message gets parsed in json format.
						i := strings.Index(contents, "{")
						if i > -1 {
							var jsonenc interface{}
							err := json.Unmarshal([]byte(contents[i:]), &jsonenc)
							if err != nil {
								t.Error("Expected response to be able to be json unmarshalled but got error: ", err.Error())
							}
						} else {
							t.Error("Did not receive json in contents but was expecting to.")
						}
					},
				)
			}
		},
	)

	t.Run(
		"Using Go standard format",
		func(t *testing.T) {
			tests := map[string]func(a *APILog, message string){
				LEVEL_STRING_INFO: func(a *APILog, message string) {
					a.Info(message)
				},
				LEVEL_STRING_DEBUG: func(a *APILog, message string) {
					a.Debug(message)
				},
			}

			var b bytes.Buffer
			//Set output
			LogOutput = &b
			LogFlags = 0
			LogJsonFormat = false
			a := New()

			for k, v := range tests {
				t.Run(
					k,
					func(t *testing.T) {
						defer b.Reset()
						message := "some logger message."

						v(a, message)

						contents := b.String()
						expected := fmt.Sprintf("[%s]: %v\n", strings.ToUpper(k), message)

						if contents != expected {
							t.Errorf(`Log message did not have expected format. Got "%v", want "%v"`, contents, expected)
						}
					},
				)
			}
		},
	)
}

func TestCanLogByLevel(t *testing.T) {
	tests := map[LogLevel]map[string]func(a *APILog, message string) string{
		LEVEL_DEBUG: {
			LEVEL_STRING_INFO: func(a *APILog, message string) string {
				a.Info(message)
				return ""
			},
			LEVEL_STRING_DEBUG: func(a *APILog, message string) string {
				a.Debug(message)
				return fmt.Sprintf("[%s]: %v\n", strings.ToUpper(LEVEL_STRING_DEBUG), message)
			},
		},
		LEVEL_INFO: {
			LEVEL_STRING_INFO: func(a *APILog, message string) string {
				a.Info(message)
				return fmt.Sprintf("[%s]: %v\n", strings.ToUpper(LEVEL_STRING_INFO), message)
			},
			LEVEL_STRING_DEBUG: func(a *APILog, message string) string {
				a.Debug(message)
				return ""
			},
		},
	}

	var b bytes.Buffer

	LogOutput = &b
	LogJsonFormat = false
	LogFlags = 0
	a := New()

	for k, v := range tests {
		lvl := "output level set for info only."
		if k == LEVEL_DEBUG {
			lvl = "output level set for debug only."
		}
		t.Run(
			lvl,
			func(t *testing.T) {
				OutputLevel = k
				for header, callback := range v {
					message := "some " + header + " message"
					expected := callback(a, message)
					contents := b.String()

					if expected != contents {
						t.Errorf(`Output of logger did not match expected. Got "%s", want "%s"`, contents, expected)
					}
					b.Reset()
				}
			},
		)
	}
}

func ExampleNew() {
	// Create a writer
	// For real world usage it's probably more common with using something like os.Stdout
	var b bytes.Buffer

	// Assign writer to LogOutput
	LogOutput = &b

	// Create the logger
	logger := New()

	// Write an info log.
	logger.Info("Some informational message.")

	// Print the buffer.
	fmt.Println(b.String())

	//Will print something like: file.go:1: {"level":"INFO","message":"\"Some informational message.\"","timestamp":"2017-07-14 16:12:03.425Z"}
}

func ExampleAPILog_Info() {
	logJson := []bool{true, false}

	LogOutput = os.Stdout

	for _,v := range logJson {
		LogJsonFormat = v

		logger := New()
		logger.Info("Some informational message.")
	}

	//Will print something like:
	// file.go:1: {"level":"INFO","message":"\"Some informational message.\"","timestamp":"2017-07-14 16:12:03.425Z"}
	// file.go:1: [INFO] Some informational message.
}

func ExampleAPILog_Debug() {
	logJson := []bool{true, false}

	LogOutput = os.Stdout

	for _,v := range logJson {
		LogJsonFormat = v

		logger := New()
		msg := struct{Msg string}{Msg: "Some debugging message."}
		logger.Debug(msg)
	}

	//Will print something like:
	// file.go:1: {"level":"DEBUG","message":"{\"Msg\":\"Some debugging message\"}","timestamp":"2017-07-14 16:12:03.425Z"}
	// file.go:1: [DEBUG] {Some debugging message.}
}

func Example() {
	// Set output
	LogOutput = os.Stdout

	// Set levels to log. Default is LEVEL_INFO | LEVEL_DEBUG
	OutputLevel = LEVEL_INFO

	// Set flags. Default is log.Lshortfile
	LogFlags = log.Ltime | log.Llongfile

	// Create new logger
	logger := New()

	// Log
	logger.Info("Some informational message.")
}
