package internal

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
)

const LOG_FILENAME = "/tmp/metarf.log"

type Logger struct {
	logger  *zerolog.Logger
	logFile *os.File
}

func NewLogger(isDebug bool) *Logger {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logLevel := zerolog.InfoLevel
	if isDebug {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// Create logfile
	logFile, err := os.OpenFile(LOG_FILENAME, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to set logging to %s", LOG_FILENAME)
	}

	logger := zerolog.New(logFile).With().Timestamp().Logger()
	if err != nil {
		log.Error().Err(err).Msg("Unable to start logging")
	}
	return &Logger{logger: &logger, logFile: logFile}
}

// Output duplicates the global logger and sets w as its output.
func (l *Logger) Output(w io.Writer) zerolog.Logger {
	return l.logger.Output(w)
}

// With creates a child logger with the field added to its context.
func (l *Logger) With() zerolog.Context {
	return l.logger.With()
}

// Level creates a child logger with the minimum accepted level set to level.
func (l *Logger) Level(level zerolog.Level) zerolog.Logger {
	return l.logger.Level(level)
}

// Sample returns a logger with the s sampler.
func (l *Logger) Sample(s zerolog.Sampler) zerolog.Logger {
	return l.logger.Sample(s)
}

// Hook returns a logger with the h Hook.
func (l *Logger) Hook(h zerolog.Hook) zerolog.Logger {
	return l.logger.Hook(h)
}

// Debug starts a new message with debug level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Info starts a new message with info level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Warn starts a new message with warn level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error starts a new message with error level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Fatal starts a new message with fatal level. The os.Exit(1) function
// is called by the Msg method.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// Panic starts a new message with panic level. The message is also sent
// to the panic function.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Panic() *zerolog.Event {
	return l.logger.Panic()
}

// WithLevel starts a new message with level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) WithLevel(level zerolog.Level) *zerolog.Event {
	return l.logger.WithLevel(level)
}

// Log starts a new message with no level. Setting zerolog.GlobalLevel to
// zerolog.Disabled will still disable events produced by this method.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Log() *zerolog.Event {
	return l.logger.Log()
}

// Print sends a log event using debug level and no extra field.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(v ...interface{}) {
	l.logger.Print(v...)
}

// Printf sends a log event using debug level and no extra field.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

// Ctx returns the Logger associated with the ctx. If no logger
// is associated, a disabled logger is returned.
func (l *Logger) Ctx(ctx context.Context) *Logger {
	return &Logger{logger: zerolog.Ctx(ctx)}
}

func (l *Logger) Close() {
	l.logFile.Close()
}
