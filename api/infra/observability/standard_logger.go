package observability

import "log"

type StandardLogger struct{}

func NewLogger() Logger {
	return &StandardLogger{}
}

func (l *StandardLogger) Info(msg string, fields ...any) {
	log.Printf("INFO %s %v\n", msg, fields)
}

func (l *StandardLogger) Error(msg string, fields ...any) {
	log.Printf("ERROR %s %v\n", msg, fields)
}
