package logger

type LoggerFieldType map[string]any

type Logger interface {
	Info(msg string, fields LoggerFieldType)
	Error(msg string, err error, fields LoggerFieldType)
	Fatal(msg string, fields LoggerFieldType)
	Warn(msg string, fields LoggerFieldType)
	Debug(msg string, fields LoggerFieldType)
}
