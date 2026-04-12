package logger

import "log"

type Interface interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type stdLogger struct{}

func New() Interface { return &stdLogger{} }

func (stdLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (stdLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}
