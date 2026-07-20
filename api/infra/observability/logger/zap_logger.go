package logger

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

var globalLogger Logger

type zapLogger struct {
	logger *zap.SugaredLogger
}

func Initialize() Logger {
	l, _ := zap.NewProduction()
	globalLogger = &zapLogger{
		logger: l.Sugar(),
	}
	return globalLogger
}

func GetLogger() Logger {
	if globalLogger == nil {
		return Initialize()
	}
	return globalLogger
}

func FromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(loggerKey{}).(Logger); ok {
		return l
	}
	return GetLogger()
}

func ToContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func (l *zapLogger) Info(msg string, fields LoggerFieldType) {
	l.logger.Infow(msg, l.mapToSlice(fields)...)
}

func (l *zapLogger) Error(msg string, err error, fields LoggerFieldType) {
	f := l.mapToSlice(fields)
	if err != nil {
		f = append(f, "error", err)
	}
	l.logger.Errorw(msg, f...)
}

func (l *zapLogger) Fatal(msg string, fields LoggerFieldType) {
	l.logger.Fatalw(msg, l.mapToSlice(fields)...)
}

func (l *zapLogger) Warn(msg string, fields LoggerFieldType) {
	l.logger.Warnw(msg, l.mapToSlice(fields)...)
}

func (l *zapLogger) Debug(msg string, fields LoggerFieldType) {
	l.logger.Debugw(msg, l.mapToSlice(fields)...)
}

func (l *zapLogger) mapToSlice(fields LoggerFieldType) []any {
	if fields == nil {
		return nil
	}
	slice := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		slice = append(slice, k, v)
	}
	return slice
}
