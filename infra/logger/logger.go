package logger

import (
	"log"
	"os"
)

// Logger represents the application logger
type Logger struct {
	*log.Logger
}

// New creates a new logger instance
func New(level string) *Logger {
	logger := log.New(os.Stdout, "VISTARA-AI: ", log.LstdFlags|log.Lshortfile)
	
	return &Logger{
		Logger: logger,
	}
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.Printf("[INFO] %s", msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.Printf("[ERROR] %s", msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.Printf("[WARN] %s", msg)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.Printf("[DEBUG] %s", msg)
}
