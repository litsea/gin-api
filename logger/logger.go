package logger

import (
	"log/slog"
)

var (
	log Logger
)

func init() {
	// disable default
	log = &defaultLogger{}
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type defaultLogger struct {
	enable bool
}

func (l *defaultLogger) Debug(msg string, args ...any) {
	if !l.enable {
		return
	}

	slog.Debug(msg, args...)
}

func (l *defaultLogger) Info(msg string, args ...any) {
	if !l.enable {
		return
	}

	slog.Info(msg, args...)
}

func (l *defaultLogger) Warn(msg string, args ...any) {
	if !l.enable {
		return
	}

	slog.Warn(msg, args...)
}

func (l *defaultLogger) Error(msg string, args ...any) {
	if !l.enable {
		return
	}

	slog.Error(msg, args...)
}

func SetLogger(l Logger) {
	log = l
}

func SetDefaultLogger() {
	log = &defaultLogger{enable: true}
}

func Debug(msg string, args ...any) {
	log.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	log.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	log.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	log.Error(msg, args...)
}
