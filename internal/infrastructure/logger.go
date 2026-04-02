package infrastructure

import (
	"log"
	"os"
)

// Logger provides structured logging interface
type Logger struct {
	info  *log.Logger
	error *log.Logger
	debug *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debug: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs an informational message
func (l *Logger) Info(msg string) {
	l.info.Println(msg)
}

// Infof logs a formatted informational message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.info.Printf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.error.Println(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.error.Printf(format, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.debug.Println(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.debug.Printf(format, args...)
}
