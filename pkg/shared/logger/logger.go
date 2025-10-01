package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"cmp/pkg/shared/errors"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     *errors.APIError       `json:"error,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// Logger represents a structured logger
type Logger struct {
	level     LogLevel
	output    io.Writer
	fields    map[string]interface{}
	caller    bool
	requestID string
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel, output io.Writer) *Logger {
	return &Logger{
		level:  level,
		output: output,
		fields: make(map[string]interface{}),
		caller: true,
	}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := l.clone()
	newLogger.fields[key] = value
	return newLogger
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := l.clone()
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// WithRequestID adds request ID to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	newLogger := l.clone()
	newLogger.requestID = requestID
	return newLogger
}

// clone creates a copy of the logger
func (l *Logger) clone() *Logger {
	fields := make(map[string]interface{})
	for k, v := range l.fields {
		fields[k] = v
	}
	return &Logger{
		level:     l.level,
		output:    l.output,
		fields:    fields,
		caller:    l.caller,
		requestID: l.requestID,
	}
}

// getCaller returns the caller information
func (l *Logger) getCaller() string {
	if !l.caller {
		return ""
	}

	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return ""
	}

	// Get just the filename, not the full path
	parts := strings.Split(file, "/")
	filename := parts[len(parts)-1]

	return fmt.Sprintf("%s:%d", filename, line)
}

// log writes a log entry
func (l *Logger) log(level LogLevel, message string, err *errors.APIError) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    l.fields,
		Error:     err,
		Caller:    l.getCaller(),
		RequestID: l.requestID,
	}

	// Marshal to JSON
	jsonData, marshalErr := json.Marshal(entry)
	if marshalErr != nil {
		// Fallback to simple logging if JSON marshaling fails
		log.Printf("[%s] %s: %s", level.String(), message, marshalErr.Error())
		return
	}

	// Write to output
	fmt.Fprintln(l.output, string(jsonData))
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	l.log(DEBUG, message, nil)
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.log(INFO, message, nil)
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.log(WARN, message, nil)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.log(ERROR, message, nil)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string) {
	l.log(FATAL, message, nil)
	os.Exit(1)
}

// ErrorWithAPIError logs an error with API error details
func (l *Logger) ErrorWithAPIError(message string, apiErr *errors.APIError) {
	l.log(ERROR, message, apiErr)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...))
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
}

// SetCaller enables or disables caller information
func (l *Logger) SetCaller(enabled bool) {
	l.caller = enabled
}

// Global logger instance
var (
	DefaultLogger *Logger
)

// Initialize the default logger
func init() {
	DefaultLogger = NewLogger(INFO, os.Stdout)
}

// SetDefaultLogger sets the default logger
func SetDefaultLogger(logger *Logger) {
	DefaultLogger = logger
}

// Global logging functions
func Debug(message string) {
	DefaultLogger.Debug(message)
}

func Info(message string) {
	DefaultLogger.Info(message)
}

func Warn(message string) {
	DefaultLogger.Warn(message)
}

func Error(message string) {
	DefaultLogger.Error(message)
}

func Fatal(message string) {
	DefaultLogger.Fatal(message)
}

func ErrorWithAPIError(message string, apiErr *errors.APIError) {
	DefaultLogger.ErrorWithAPIError(message, apiErr)
}

func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	DefaultLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	DefaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	DefaultLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	DefaultLogger.Fatalf(format, args...)
}

// WithRequestID returns a logger with request ID context
func WithRequestID(requestID string) *Logger {
	if requestID == "" {
		return DefaultLogger
	}
	return DefaultLogger.WithField("request_id", requestID)
}
