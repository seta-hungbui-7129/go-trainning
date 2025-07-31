package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	WithContext(ctx context.Context) Logger
	WithFields(fields ...Field) Logger
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// LogrusLogger is a logrus-based implementation of Logger
type LogrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// NewLogger creates a new structured logger
func NewLogger(level string, format string, output io.Writer) Logger {
	logger := logrus.New()
	
	// Set output
	if output != nil {
		logger.SetOutput(output)
	} else {
		logger.SetOutput(os.Stdout)
	}
	
	// Set level
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
	
	// Set format
	if format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}
	
	return &LogrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}
}

func (l *LogrusLogger) fieldsToLogrus(fields []Field) logrus.Fields {
	logrusFields := make(logrus.Fields)
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}
	return logrusFields
}

func (l *LogrusLogger) Debug(msg string, fields ...Field) {
	if len(fields) > 0 {
		l.entry.WithFields(l.fieldsToLogrus(fields)).Debug(msg)
	} else {
		l.entry.Debug(msg)
	}
}

func (l *LogrusLogger) Info(msg string, fields ...Field) {
	if len(fields) > 0 {
		l.entry.WithFields(l.fieldsToLogrus(fields)).Info(msg)
	} else {
		l.entry.Info(msg)
	}
}

func (l *LogrusLogger) Warn(msg string, fields ...Field) {
	if len(fields) > 0 {
		l.entry.WithFields(l.fieldsToLogrus(fields)).Warn(msg)
	} else {
		l.entry.Warn(msg)
	}
}

func (l *LogrusLogger) Error(msg string, fields ...Field) {
	if len(fields) > 0 {
		l.entry.WithFields(l.fieldsToLogrus(fields)).Error(msg)
	} else {
		l.entry.Error(msg)
	}
}

func (l *LogrusLogger) Fatal(msg string, fields ...Field) {
	if len(fields) > 0 {
		l.entry.WithFields(l.fieldsToLogrus(fields)).Fatal(msg)
	} else {
		l.entry.Fatal(msg)
	}
}

func (l *LogrusLogger) WithContext(ctx context.Context) Logger {
	return &LogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithContext(ctx),
	}
}

func (l *LogrusLogger) WithFields(fields ...Field) Logger {
	return &LogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(l.fieldsToLogrus(fields)),
	}
}

// Global logger instance
var globalLogger Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(level string, format string, output io.Writer) {
	globalLogger = NewLogger(level, format, output)
}

// GetLogger returns the global logger instance
func GetLogger() Logger {
	if globalLogger == nil {
		globalLogger = NewLogger("info", "json", nil)
	}
	return globalLogger
}

// Convenience functions using global logger
func Debug(msg string, fields ...Field) {
	GetLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	GetLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetLogger().Warn(msg, fields...)
}

func ErrorLog(msg string, fields ...Field) {
	GetLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	GetLogger().Fatal(msg, fields...)
}
