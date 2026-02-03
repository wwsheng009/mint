// Package log provides structured logging for the mint TUI framework.
// It provides domain-specific loggers for focus, reconciler, and other components
// with support for debug mode via environment variables.
package log

import (
	"fmt"
	"os"
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

// Logger is a structured logger that writes to stderr
type Logger struct {
	mu       sync.Mutex
	prefix   string
	enabled  bool
	category string
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
	fmt.Fprintln(os.Stderr, msg)
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
