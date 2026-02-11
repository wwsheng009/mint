// Package log provides structured logging for the mint TUI framework.
// It provides domain-specific loggers for focus, reconciler, and other components
// with support for debug mode via environment variables.
//
// Log output destination is controlled by TUI_LOG_OUTPUT environment variable:
// - "file" or "": output to file only (default)
// - "console": output to console only
// - "both": output to both file and console
//
// Default log file path: logs/application.log
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// LevelDebug is for detailed debugging information
	LevelDebug LogLevel = iota
	// LevelInfo is for general informational messages
	LevelInfo
	// LevelWarn is for warning messages
	LevelWarn
	// LevelError is for error messages
	LevelError
)

// LogOutput represents the logging output destination
type LogOutput int

const (
	// OutputFile logs only to file
	OutputFile LogOutput = iota
	// OutputConsole logs only to console (stderr)
	OutputConsole
	// OutputBoth logs to both file and console
	OutputBoth
)

// Logger is a structured logger that can write to console and/or file
type Logger struct {
	mu       sync.Mutex
	prefix   string
	enabled  bool
	category string
	file     *os.File
}

// NewLogger creates a new logger with the given prefix and category.
// The category is used to check the TUI_DEBUG_<category> environment variable.
func NewLogger(prefix, category string) *Logger {
	l := &Logger{
		prefix:   prefix,
		category: category,
	}
	l.checkEnabled()
	return l
}

// checkEnabled checks if debug logging is enabled for this logger's category
func (l *Logger) checkEnabled() {
	if l.category == "" {
		l.enabled = os.Getenv("TUI_DEBUG") == "true"
		return
	}
	// Check category-specific debug flag (e.g., TUI_DEBUG_FOCUS)
	envVar := "TUI_DEBUG_" + l.category
	l.enabled = os.Getenv(envVar) == "true" || os.Getenv("TUI_DEBUG") == "true"
}

// Enabled returns whether this logger is currently enabled
func (l *Logger) Enabled() bool {
	return l.enabled
}

// SetEnabled sets whether this logger is enabled
func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// Debug logs a debug message if debugging is enabled
func (l *Logger) Debug(format string, args ...any) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...any) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...any) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...any) {
	l.log(LevelError, format, args...)
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.enabled && level == LevelDebug {
		return
	}

	msg := fmt.Sprintf(format, args...)
	if l.prefix != "" {
		msg = "[" + l.prefix + "] " + msg
	}

	// Output based on configuration
	switch getLogOutput() {
	case OutputFile:
		if logFile := getLogFile(); logFile != nil {
			fmt.Fprintln(logFile, msg)
			logFile.Sync()
		}
	case OutputConsole:
		fmt.Fprintln(os.Stderr, msg)
	case OutputBoth:
		// Output to stderr
		fmt.Fprintln(os.Stderr, msg)
		// Output to file
		if logFile := getLogFile(); logFile != nil {
			fmt.Fprintln(logFile, msg)
			logFile.Sync()
		}
	}
}

// Focus logs a focus-related debug message
func (l *Logger) Focus(format string, args ...any) {
	l.Debug("[Focus] "+format, args...)
}

// Reconciler logs a reconciler-related debug message
func (l *Logger) Reconciler(format string, args ...any) {
	l.Debug("[Reconciler] "+format, args...)
}

// Render logs a render-related debug message
func (l *Logger) Render(format string, args ...any) {
	l.Debug("[Render] "+format, args...)
}

// Key logs a key event debug message
func (l *Logger) Key(format string, args ...any) {
	l.Debug("[Key] "+format, args...)
}

// =============================================================================
// Global Loggers
// =============================================================================

var (
	// FocusLogger logs focus-related messages (enabled via TUI_DEBUG_FOCUS)
	FocusLogger = NewLogger("FocusManager", "FOCUS")

	// ReconcilerLogger logs reconciler-related messages (enabled via TUI_DEBUG_RECONCILER)
	ReconcilerLogger = NewLogger("Reconciler", "RECONCILER")

	// RenderLogger logs render-related messages (enabled via TUI_DEBUG_RENDER)
	RenderLogger = NewLogger("Render", "RENDER")

	// KeyLogger logs key event messages (enabled via TUI_DEBUG_KEYS)
	KeyLogger = NewLogger("KeyEvent", "KEYS")

	// EventLogger logs event-related messages (enabled via TUI_DEBUG_EVENTS)
	EventLogger = NewLogger("Event", "EVENTS")

	WinLogger = NewLogger("Windows", "WIN")

	InspectorLogger = NewLogger("Inspector", "INSPECTOR")

	EngineLogger = NewLogger("Engine", "ENGINE")
	// UILogger logs UI-related messages (enabled via TUI_DEBUG_UI)
	UILogger = NewLogger("UI", "UI")

	// ButtonLogger logs button-specific messages (enabled via TUI_DEBUG_BUTTON)
	ButtonLogger = NewLogger("Button", "BUTTON")
)

// SetAllEnabled sets the enabled state for all global loggers
func SetAllEnabled(enabled bool) {
	FocusLogger.SetEnabled(enabled)
	ReconcilerLogger.SetEnabled(enabled)
	RenderLogger.SetEnabled(enabled)
	KeyLogger.SetEnabled(enabled)
	UILogger.SetEnabled(enabled)
	ButtonLogger.SetEnabled(enabled)
}

// =============================================================================
// Global Log File Initialization
// =============================================================================

var (
	globalLogFile *os.File
	logFileOnce   sync.Once
	logOutput     LogOutput
	outputOnce    sync.Once
)

// getLogOutput returns the configured log output destination
// Controlled by TUI_LOG_OUTPUT environment variable:
// - "file" or "": output to file only (default)
// - "console": output to console only
// - "both": output to both file and console
func getLogOutput() LogOutput {
	outputOnce.Do(func() {
		switch os.Getenv("TUI_LOG_OUTPUT") {
		case "console":
			logOutput = OutputConsole
		case "both":
			logOutput = OutputBoth
		case "file", "":
			logOutput = OutputFile
		default:
			// Default to file output
			logOutput = OutputFile
		}
	})
	return logOutput
}

// initLogFile initializes the global log file
func initLogFile() {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "[Logger] Failed to create logs directory: %v\n", err)
		return
	}

	// Open log file in append mode
	logPath := filepath.Join(logDir, "application.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Logger] Failed to open log file: %v\n", err)
		return
	}

	globalLogFile = file
}

// getLogFile returns the global log file, initializing it if necessary
func getLogFile() *os.File {
	logFileOnce.Do(initLogFile)
	return globalLogFile
}

// CloseLogFile closes the global log file
func CloseLogFile() error {
	if globalLogFile != nil {
		err := globalLogFile.Close()
		globalLogFile = nil
		return err
	}
	return nil
}

// FlushLogFile flushes the global log file to disk
func FlushLogFile() error {
	if globalLogFile != nil {
		return globalLogFile.Sync()
	}
	return nil
}
