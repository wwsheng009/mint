// Package devtools provides the logger for DevTools.
//
// P1-4: 实现轻量级日志系统，支持可配置级别和环形缓冲区
package devtools

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// LogLevel represents the severity level of a log message.
type LogLevel int

const (
	// LevelDebug represents debug messages.
	LevelDebug LogLevel = iota
	// LevelInfo represents informational messages.
	LevelInfo
	// LevelWarn represents warning messages.
	LevelWarn
	// LevelError represents error messages.
	LevelError
	// LevelNone disables all logging.
	LevelNone
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelNone:
		return "NONE"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a single log entry.
type LogEntry struct {
	Level     LogLevel
	Timestamp time.Time
	Message   string
	Fields    map[string]interface{}
}

// Logger is a lightweight logger for DevTools.
// P1-4: 使用环形缓冲区减少内存分配
type Logger struct {
	// Atomic state
	level   atomic.Int32 // Current log level
	enabled atomic.Bool  // Enable/disable flag

	// Ring buffer for recent entries
	buffer    []LogEntry
	bufferPos int
	bufferMu  sync.RWMutex
	bufferCap int

	// Output (optional)
	outputMu sync.Mutex
	output   OutputFunc

	// Statistics
	stats loggerCounters
}

// OutputFunc is a function that writes log entries.
type OutputFunc func(entry *LogEntry)

type loggerCounters struct {
	LogsEmitted atomic.Uint64
	LogsDropped atomic.Uint64
}

// LoggerStats contains a point-in-time logger statistics snapshot.
type LoggerStats struct {
	LogsEmitted  uint64
	LogsDropped  uint64
	BufferSize   int
	CurrentLevel LogLevel
}

// NewLogger creates a new logger with the specified buffer size.
func NewLogger(bufferSize int) *Logger {
	// Ensure buffer size is a power of 2
	if bufferSize&(bufferSize-1) != 0 {
		bufferSize = 1 << (32 - nlz32(uint32(bufferSize)))
	}

	l := &Logger{
		buffer:    make([]LogEntry, bufferSize),
		bufferCap: bufferSize,
		bufferPos: 0,
		output:    nil, // No output by default
	}
	l.level.Store(int32(LevelInfo))
	l.enabled.Store(false)
	return l
}

// Enable enables the logger.
func (l *Logger) Enable() {
	l.enabled.Store(true)
}

// Disable disables the logger.
func (l *Logger) Disable() {
	l.enabled.Store(false)
}

// IsEnabled returns true if the logger is enabled.
func (l *Logger) IsEnabled() bool {
	return l.enabled.Load()
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level LogLevel) {
	l.level.Store(int32(level))
}

// GetLevel returns the current log level.
func (l *Logger) GetLevel() LogLevel {
	return LogLevel(l.level.Load())
}

// SetOutput sets the output function for log entries.
func (l *Logger) SetOutput(fn OutputFunc) {
	l.outputMu.Lock()
	defer l.outputMu.Unlock()
	l.output = fn
}

// log logs a message at the specified level.
// P1-4: 无锁快速路径，环形缓冲区写入
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	// Fast path: disabled or level too low
	if !l.enabled.Load() {
		return
	}
	if level < l.GetLevel() {
		return
	}

	// Create log entry
	entry := LogEntry{
		Level:     level,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(format, args...),
		Fields:    nil,
	}

	// Write to ring buffer
	l.bufferMu.Lock()
	l.buffer[l.bufferPos] = entry
	l.bufferPos = (l.bufferPos + 1) & (l.bufferCap - 1)
	l.bufferMu.Unlock()

	// Update stats
	l.stats.LogsEmitted.Add(1)

	// Call output function if set (without holding buffer lock)
	l.outputMu.Lock()
	out := l.output
	l.outputMu.Unlock()

	if out != nil {
		// Run output in separate goroutine to avoid blocking
		go out(&entry)
	}
}

// Logf logs a formatted message at the specified level.
func (l *Logger) Logf(level LogLevel, format string, args ...interface{}) {
	l.log(level, format, args...)
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an informational message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// WithFields logs a message with additional fields.
func (l *Logger) WithFields(level LogLevel, fields map[string]interface{}, format string, args ...interface{}) {
	// Fast path: disabled or level too low
	if !l.enabled.Load() {
		return
	}
	if level < l.GetLevel() {
		return
	}

	entry := LogEntry{
		Level:     level,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(format, args...),
		Fields:    fields,
	}

	l.bufferMu.Lock()
	l.buffer[l.bufferPos] = entry
	l.bufferPos = (l.bufferPos + 1) & (l.bufferCap - 1)
	l.bufferMu.Unlock()

	l.stats.LogsEmitted.Add(1)
}

// GetRecentEntries returns the most recent log entries.
func (l *Logger) GetRecentEntries(maxEntries int) []LogEntry {
	l.bufferMu.RLock()
	defer l.bufferMu.RUnlock()

	if maxEntries > l.bufferCap {
		maxEntries = l.bufferCap
	}

	entries := make([]LogEntry, maxEntries)

	// Read entries in reverse order (most recent first)
	for i := 0; i < maxEntries; i++ {
		pos := (l.bufferPos - 1 - i) & (l.bufferCap - 1)
		if l.buffer[pos].Timestamp.IsZero() {
			// Empty slot
			break
		}
		entries[i] = l.buffer[pos]
	}

	return entries
}

// GetAllEntries returns all log entries in the buffer.
func (l *Logger) GetAllEntries() []LogEntry {
	l.bufferMu.RLock()
	defer l.bufferMu.RUnlock()

	// Count non-empty entries
	count := 0
	for i := 0; i < l.bufferCap; i++ {
		if !l.buffer[i].Timestamp.IsZero() {
			count++
		}
	}

	entries := make([]LogEntry, count)
	idx := 0

	// Read in chronological order
	startPos := l.bufferPos
	for i := 0; i < l.bufferCap && idx < count; i++ {
		pos := (startPos + i) & (l.bufferCap - 1)
		if !l.buffer[pos].Timestamp.IsZero() {
			entries[idx] = l.buffer[pos]
			idx++
		}
	}

	return entries
}

// Clear clears all log entries.
func (l *Logger) Clear() {
	l.bufferMu.Lock()
	defer l.bufferMu.Unlock()

	for i := range l.buffer {
		l.buffer[i] = LogEntry{}
	}
	l.bufferPos = 0
}

// GetStats returns the current logger statistics.
func (l *Logger) GetStats() LoggerStats {
	return LoggerStats{
		LogsEmitted:  l.stats.LogsEmitted.Load(),
		LogsDropped:  l.stats.LogsDropped.Load(),
		BufferSize:   l.bufferCap,
		CurrentLevel: l.GetLevel(),
	}
}

// SetStdoutOutput sets stdout as the output destination.
func (l *Logger) SetStdoutOutput() {
	l.SetOutput(func(entry *LogEntry) {
		fmt.Fprintf(os.Stdout, "[%s] %s %s\n",
			entry.Timestamp.Format("15:04:05.000"),
			entry.Level.String(),
			entry.Message)
	})
}

// SetStderrOutput sets stderr as the output destination.
func (l *Logger) SetStderrOutput() {
	l.SetOutput(func(entry *LogEntry) {
		fmt.Fprintf(os.Stderr, "[%s] %s %s\n",
			entry.Timestamp.Format("15:04:05.000"),
			entry.Level.String(),
			entry.Message)
	})
}

// Flush flushes any pending log entries.
// Currently a no-op since logs are written immediately,
// but provided for interface compatibility.
func (l *Logger) Flush() {
	// No-op for immediate write mode
}

// Default logger instance
var defaultLogger = NewLogger(256)

// Default logger functions for convenience

// SetDefaultLogLevel sets the log level for the default logger.
func SetDefaultLogLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetDefaultLoggerOutput sets the output for the default logger.
func SetDefaultLoggerOutput(fn OutputFunc) {
	defaultLogger.SetOutput(fn)
}

// EnableDefaultLogger enables the default logger.
func EnableDefaultLogger() {
	defaultLogger.Enable()
}

// DisableDefaultLogger disables the default logger.
func DisableDefaultLogger() {
	defaultLogger.Disable()
}

// LogDebug logs a debug message to the default logger.
func LogDebug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// LogInfo logs an info message to the default logger.
func LogInfo(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// LogWarn logs a warning message to the default logger.
func LogWarn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// LogError logs an error message to the default logger.
func LogError(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// GetDefaultLogger returns the default logger instance.
func GetDefaultLogger() *Logger {
	return defaultLogger
}
