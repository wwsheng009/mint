// Package logger provides file-based logging for debugging
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	globalLogger *Logger
	once         sync.Once
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
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
	default:
		return "UNKNOWN"
	}
}

// Logger is a file-based logger
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	path     string
	level    LogLevel
	enabled  bool
	category string
}

// Init initializes the global logger
func Init(logPath string) *Logger {
	once.Do(func() {
		// Create directory if not exists
		dir := filepath.Dir(logPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
			return
		}

		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			return
		}

		globalLogger = &Logger{
			file:    file,
			path:    logPath,
			level:   DEBUG,
			enabled: true,
		}

		// Write header
		globalLogger.log(INFO, "LOGGER", "Logger initialized: %s", logPath)
	})

	return globalLogger
}

// Get returns the global logger
func Get() *Logger {
	if globalLogger == nil {
		return Init("tui_debug.log")
	}
	return globalLogger
}

// EnableCategory enables logging for a specific category
func (l *Logger) EnableCategory(category string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	newLogger := *l
	newLogger.category = category
	return &newLogger
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Enable enables the logger
func (l *Logger) Enable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = true
}

// Disable disables the logger
func (l *Logger) Disable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = false
}

// IsEnabled returns whether the logger is enabled
func (l *Logger) IsEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabled
}

// Debug logs a debug message
func (l *Logger) Debug(category string, format string, args ...interface{}) {
	l.log(DEBUG, category, format, args...)
}

// Info logs an info message
func (l *Logger) Info(category string, format string, args ...interface{}) {
	l.log(INFO, category, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(category string, format string, args ...interface{}) {
	l.log(WARN, category, format, args...)
}

// Error logs an error message
func (l *Logger) Error(category string, format string, args ...interface{}) {
	l.log(ERROR, category, format, args...)
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, category string, format string, args ...interface{}) {
	if l == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.enabled || level < l.level {
		return
	}

	if l.file == nil {
		return
	}

	// Format: [TIMESTAMP] [LEVEL] [CATEGORY] message
	timestamp := time.Now().Format("15:04:05.000")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] [%s] %s\n", timestamp, level.String(), category, message)

	l.file.WriteString(logLine)
}

// Close closes the log file
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Flush flushes the log file
func (l *Logger) Flush() error {
	if l == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Sync()
	}
	return nil
}

// Convenience functions for global logger

// Debug logs a debug message to the global logger
func Debug(category string, format string, args ...interface{}) {
	Get().Debug(category, format, args...)
}

// Info logs an info message to the global logger
func Info(category string, format string, args ...interface{}) {
	Get().Info(category, format, args...)
}

// Warn logs a warning message to the global logger
func Warn(category string, format string, args ...interface{}) {
	Get().Warn(category, format, args...)
}

// Error logs an error message to the global logger
func Error(category string, format string, args ...interface{}) {
	Get().Error(category, format, args...)
}

// InitFromEnv initializes the logger from environment variables
// TUI_DEBUG_LOG=path/to/log.log
// TUI_DEBUG_LEVEL=DEBUG|INFO|WARN|ERROR
func InitFromEnv() *Logger {
	logPath := os.Getenv("TUI_DEBUG_LOG")
	if logPath == "" {
		logPath = "tui_debug.log"
	}

	levelStr := os.Getenv("TUI_DEBUG_LEVEL")
	level := DEBUG
	switch levelStr {
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	}

	logger := Init(logPath)
	if logger != nil {
		logger.SetLevel(level)
	}
	return logger
}

// CheckEnv checks if logging is enabled via environment variable
func IsEnabled() bool {
	return os.Getenv("TUI_DEBUG_LOG") != "" || os.Getenv("TUI_DEBUG") == "true"
}

// LogIfEnabled logs a message only if logging is enabled
func LogIfEnabled(category string, format string, args ...interface{}) {
	if IsEnabled() {
		Info(category, format, args...)
	}
}
