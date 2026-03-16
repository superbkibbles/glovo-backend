package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

// Init initializes the logger
func Init(env string) {
	var output io.Writer = os.Stdout

	if env == "development" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	log = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	if env == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// Get returns the logger instance
func Get() *zerolog.Logger {
	return &log
}

// Info logs an info message
func Info() *zerolog.Event {
	return log.Info()
}

// Debug logs a debug message
func Debug() *zerolog.Event {
	return log.Debug()
}

// Error logs an error message
func Error() *zerolog.Event {
	return log.Error()
}

// Warn logs a warning message
func Warn() *zerolog.Event {
	return log.Warn()
}

// Fatal logs a fatal message and exits
func Fatal() *zerolog.Event {
	return log.Fatal()
}

// WithService returns a logger with service name
func WithService(service string) zerolog.Logger {
	return log.With().Str("service", service).Logger()
}

// WithRequestID returns a logger with request ID
func WithRequestID(requestID string) zerolog.Logger {
	return log.With().Str("request_id", requestID).Logger()
}
