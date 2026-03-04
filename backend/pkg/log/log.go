package log

import (
	"os"

	"github.com/rs/zerolog"
)

var Log = MustInitLogger()

type Logger struct {
	infoLogger  *zerolog.Logger
	errorLogger *zerolog.Logger
	infoLevel   zerolog.Level
	errorLevel  zerolog.Level
}

func (l *Logger) Info(s string) {
	l.infoLogger.WithLevel(l.infoLevel).Caller(1).Msg(s)
}

func (l *Logger) Error(s error) {
	l.errorLogger.WithLevel(l.errorLevel).Caller(1).Msg(s.Error())
}

func MustInitLogger() *Logger {
	zerolog.TimeFieldFormat = "15:04:05 02-01-2006"

	infoLevel := zerolog.InfoLevel
	errorLevel := zerolog.ErrorLevel

	infoLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	errorLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	log := &Logger{
		infoLogger:  &infoLogger,
		errorLogger: &errorLogger,
		infoLevel:   infoLevel,
		errorLevel:  errorLevel,
	}

	return log
}
