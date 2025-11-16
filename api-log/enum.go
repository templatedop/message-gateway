package log

import (
	"strings"

	"github.com/rs/zerolog"
)

func FetchLogLevel(level string) zerolog.Level {
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "no-level":
		return zerolog.NoLevel
	case "disabled":
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

type logOutputWriter int

const (
	stdoutOutputWriter logOutputWriter = iota
	noopOutputWriter
	testOutputWriter
	consoleOutputWriter
)

func (l logOutputWriter) String() string {
	switch l {
	case noopOutputWriter:
		return noop
	case testOutputWriter:
		return test
	case consoleOutputWriter:
		return console
	default:
		return stdout
	}
}

// FetchLogOutputWriter returns a [LogOutputWriter] for a given value.
func FetchLogOutputWriter(l string) logOutputWriter {
	switch strings.ToLower(l) {
	case noop:
		return noopOutputWriter
	case test:
		return testOutputWriter
	case console:
		return consoleOutputWriter
	default:
		return stdoutOutputWriter
	}
}
