// Package log is the package wrapping logrus
package log

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

func init() {
	rootLogger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &rootLogger
}

type LoggerField func(c zerolog.Context) zerolog.Context

func CtxWithLogger(ctx context.Context, fields ...LoggerField) context.Context {
	l := zerolog.Ctx(ctx).With()
	for _, f := range fields {
		l = f(l)
	}
	return l.Logger().WithContext(ctx)
}

// LogField is a function that adds a field to a log entry
type LogField func(e *zerolog.Event) *zerolog.Event

// log outputs a log entry with specified level
func log(ctx context.Context, level zerolog.Level, msg string, fields ...LogField) {
	e := zerolog.Ctx(ctx).WithLevel(level)
	for _, f := range fields {
		e = f(e)
	}
	e.Msg(msg)
}

// logf outputs a formatted log entry with specified level
// Use only for debugging/diagnosing purposes
func logf(ctx context.Context, level zerolog.Level, format string, args ...interface{}) {
	zerolog.Ctx(ctx).WithLevel(level).Msgf(format, args...)
}

// Info outputs information logs
func Info(ctx context.Context, msg string, fields ...LogField) {
	log(ctx, zerolog.InfoLevel, msg, fields...)
}

// Infof outputs information logs
// Use only for debugging/diagnosing purposes
func Infof(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, zerolog.InfoLevel, format, args...)
}

// Debug outputs debug logs
func Debug(ctx context.Context, msg string, fields ...LogField) {
	log(ctx, zerolog.DebugLevel, msg, fields...)
}

// Debugf outputs debug logs with formatting
func Debugf(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, zerolog.DebugLevel, format, args...)
}

// Warn outputs warning logs
func Warn(ctx context.Context, msg string, fields ...LogField) {
	log(ctx, zerolog.WarnLevel, msg, fields...)
}

// Warnf outputs warning logs with formatting
func Warnf(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, zerolog.WarnLevel, format, args...)
}

// Error outputs error logs
func Error(ctx context.Context, msg string, fields ...LogField) {
	log(ctx, zerolog.ErrorLevel, msg, fields...)
}

// Errorf outputs error logs with formatting
func Errorf(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, zerolog.ErrorLevel, format, args...)
}
