package log

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// isEnvEnabled checks if an environment variable is set to a truthy value
func isEnvEnabled(value string) bool {
	if value == "" {
		return false
	}
	// Support multiple truthy values: "true", "1", "yes", "on"
	normalized := strings.ToLower(strings.TrimSpace(value))
	return normalized == "true" || normalized == "1" || normalized == "yes" || normalized == "on"
}

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
	mu           sync.Mutex
	prefix       string
	enabled      bool
	category     string
	customEnvVar string // Optional custom environment variable name
}

// NewLogger creates a new logger with the given prefix and category.
// The category is used to check the TUI_DEBUG_<category> environment variable.
func NewLogger(prefix, category string) *Logger {
	return NewLoggerWithEnv(prefix, category, "")
}

// NewLoggerWithEnv creates a new logger with a custom environment variable name.
// If customEnvVar is empty, uses the default TUI_DEBUG_<category> format.
func NewLoggerWithEnv(prefix, category, customEnvVar string) *Logger {
	l := &Logger{
		prefix:       prefix,
		category:     category,
		customEnvVar: customEnvVar,
	}
	l.checkEnabled()
	return l
}

// checkEnabled checks if debug logging is enabled for this logger's category
func (l *Logger) checkEnabled() {
	if isEnvEnabled(os.Getenv("TUI_DEBUG_ALL")) {
		l.enabled = true
		return
	}
	if l.category == "" {
		l.enabled = isEnvEnabled(os.Getenv("TUI_DEBUG"))
		return
	}

	// Use custom env var if provided
	envVar := l.customEnvVar
	if envVar == "" {
		// Check category-specific debug flag (e.g., TUI_DEBUG_FOCUS)
		envVar = "TUI_DEBUG_" + l.category
	}
	l.enabled = isEnvEnabled(os.Getenv(envVar)) || isEnvEnabled(os.Getenv("TUI_DEBUG"))
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
		if getLogFile() != nil {
			writeToFile(msg)
		}
	case OutputConsole:
		fmt.Fprintln(os.Stderr, msg)
	case OutputBoth:
		// Output to stderr
		fmt.Fprintln(os.Stderr, msg)
		// Output to file
		if getLogFile() != nil {
			writeToFile(msg)
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

	// RenderLogger logs render-related messages (enabled via TUI_DEBUG_RENDER)
	RenderLogger = NewLogger("Render", "RENDER")

	// KeyLogger logs key event messages (enabled via TUI_DEBUG_KEY)
	KeyLogger = NewLogger("KeyEvent", "KEY")

	// EventLogger logs event-related messages (enabled via TUI_DEBUG_EVENTS)
	EventLogger = NewLogger("Event", "EVENTS")

	WinLogger = NewLogger("Windows", "WIN")
	// LinuxLogger logs Linux-specific messages (enabled via TUI_DEBUG_LINUX)
	LinuxLogger = NewLogger("Linux", "LINUX")

	InspectorLogger = NewLogger("Inspector", "INSPECTOR")

	LayoutLogger = NewLogger("Layout", "LAYOUT")

	//Layer debug
	LayerLogger = NewLogger("Layer", "LAYER")

	EngineLogger = NewLogger("Engine", "ENGINE")
	// UILogger logs UI-related messages (enabled via TUI_DEBUG_UI)
	UILogger = NewLogger("UI", "UI")

	FiberLogger = NewLogger("Fiber", "FIBER")

	// ButtonLogger logs button-specific messages (enabled via TUI_DEBUG_BUTTON)
	ButtonLogger = NewLogger("Button", "BUTTON")

	// HitMapLogger logs hit map debugging messages (enabled via TUI_DEBUG_HITMAP)
	HitMapLogger = NewLogger("HitMap", "HITMAP")

	// BorderLogger logs border debugging messages (enabled via TUI_DEBUG_BORDER)
	BorderLogger = NewLogger("Border", "BORDER")

	// PipelineLogger logs pipeline debugging messages (enabled via TUI_DEBUG_PIPELINE)
	PipelineLogger = NewLogger("Pipeline", "PIPELINE")

	// PaintLogger logs paint debugging messages (enabled via TUI_DEBUG_PAINT)
	PaintLogger = NewLogger("Paint", "PAINT")

	// WrapLogger logs wrap layout debugging messages (enabled via TUI_DEBUG_WRAP)
	WrapLogger = NewLogger("Wrap", "WRAP")

	// PumpLogger logs pump debugging messages (enabled via TUI_DEBUG_PUMP)
	PumpLogger = NewLogger("Pump", "PUMP")

	// FormLogger logs form debugging messages (enabled via TUI_DEBUG_FORM)
	FormLogger = NewLogger("Form", "FORM")

	// CursorLogger logs cursor debugging messages (enabled via TUI_DEBUG_CURSOR)
	CursorLogger = NewLogger("Cursor", "CURSOR")

	// InputLogger logs text input debugging messages (enabled via TUI_DEBUG_INPUT)
	InputLogger = NewLogger("Input", "INPUT")

	// RenderingLogger logs rendering debugging messages (enabled via TUI_DEBUG_RENDERING)
	RenderingLogger = NewLogger("Rendering", "RENDERING")

	// ValidationLogger logs validation debugging messages (enabled via TUI_DEBUG_VALIDATION)
	ValidationLogger = NewLogger("Validation", "VALIDATION")

	ActionLogger = NewLogger("Action", "ACTION")
	IntentLogger = NewLogger("Intent", "INTENT")

	PlatFormLogger = NewLogger("PlatForm", "PLATFORM")

	TempLogger = NewLogger("Temp", "TEMP")
)

// SetAllEnabled sets the enabled state for all global loggers
func SetAllEnabled(enabled bool) {
	FocusLogger.SetEnabled(enabled)
	RenderLogger.SetEnabled(enabled)
	KeyLogger.SetEnabled(enabled)
	EventLogger.SetEnabled(enabled)
	WinLogger.SetEnabled(enabled)
	LinuxLogger.SetEnabled(enabled)
	InspectorLogger.SetEnabled(enabled)
	LayoutLogger.SetEnabled(enabled)
	LayerLogger.SetEnabled(enabled)
	EngineLogger.SetEnabled(enabled)
	UILogger.SetEnabled(enabled)
	FiberLogger.SetEnabled(enabled)
	ButtonLogger.SetEnabled(enabled)
	HitMapLogger.SetEnabled(enabled)
	BorderLogger.SetEnabled(enabled)
	PipelineLogger.SetEnabled(enabled)
	PaintLogger.SetEnabled(enabled)
	WrapLogger.SetEnabled(enabled)
	PumpLogger.SetEnabled(enabled)
	FormLogger.SetEnabled(enabled)
	CursorLogger.SetEnabled(enabled)
	InputLogger.SetEnabled(enabled)
	RenderingLogger.SetEnabled(enabled)
	ValidationLogger.SetEnabled(enabled)
	ActionLogger.SetEnabled(enabled)
	IntentLogger.SetEnabled(enabled)
	PlatFormLogger.SetEnabled(enabled)
	TempLogger.SetEnabled(enabled)
	defaultLogger.SetEnabled(enabled)
}

// =============================================================================
// Default Package-Level Functions
// =============================================================================

// defaultLogger is the default logger package-level functions use
var defaultLogger = NewLogger("", "DEFAULT")

// SetDefaultPrefix sets the prefix for the default logger
func SetDefaultPrefix(prefix string) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.prefix = prefix
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...any) {
	defaultLogger.Debug(format, args...)
}

// Info logs an info message using the default logger
func Info(format string, args ...any) {
	defaultLogger.Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...any) {
	defaultLogger.Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...any) {
	defaultLogger.Error(format, args...)
}

// Fatal logs a fatal message and exits the program
func Fatal(format string, args ...any) {
	defaultLogger.Error(format, args...)
	os.Exit(1)
}

// Print logs a message at Info level using the default logger
func Print(format string, args ...any) {
	defaultLogger.Info(format, args...)
}

// Printf is an alias for Info for compatibility with standard library
func Printf(format string, args ...any) {
	defaultLogger.Info(format, args...)
}

// Println logs a message at Info level with a newline using the default logger
func Println(args ...any) {
	msg := fmt.Sprint(args...)
	defaultLogger.Info("%s", msg)
}
