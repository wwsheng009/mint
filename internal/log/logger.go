// Package log provides structured logging for the mint TUI framework.
// It provides domain-specific loggers for focus, reconciler, and other components
// with support for debug mode via environment variables.
//
// Log output destination is controlled by TUI_LOG_OUTPUT environment variable:
// - "file" or "": output to file only (default)
// - "console": output to console only
// - "both": output to both file and console
//
// Log rotation is enabled by default and controlled by environment variables:
// - TUI_LOG_MAX_SIZE: maximum size per log file (e.g., "100M", "50K", default "100M")
// - TUI_LOG_MAX_FILES: maximum number of log files to keep (default 10)
// - TUI_LOG_COMPRESS: compress old log files ("true"/"false", default "true")
//
// Default log file path: logs/application.log
package log

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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

	// ReconcilerLogger logs reconciler-related messages (enabled via TUI_DEBUG_RECONCILER)
	ReconcilerLogger = NewLogger("Reconciler", "RECONCILER")

	// RenderLogger logs render-related messages (enabled via TUI_DEBUG_RENDER)
	RenderLogger = NewLogger("Render", "RENDER")

	// KeyLogger logs key event messages (enabled via TUI_DEBUG_KEYS)
	KeyLogger = NewLogger("KeyEvent", "KEYS")

	// EventLogger logs event-related messages (enabled via TUI_DEBUG_EVENTS)
	EventLogger = NewLogger("Event", "EVENTS")

	WinLogger = NewLogger("Windows", "WIN")
	// LinuxLogger logs Linux-specific messages (enabled via TUI_DEBUG_LINUX)
	LinuxLogger = NewLogger("Linux", "LINUX")

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

	// Log rotation configuration
	rotationConfig *RotationConfig
	rotationOnce   sync.Once
)

// RotationConfig holds log rotation settings
type RotationConfig struct {
	MaxSize   int64  // Maximum size in bytes before rotation
	MaxFiles  int    // Maximum number of log files to keep
	Compress  bool   // Whether to compress old log files
	Enable    bool   // Whether rotation is enabled
}

// parseSize parses a size string like "100M", "50K", "1G" into bytes
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	var multipliers = map[rune]int64{
		'K': 1024,
		'M': 1024 * 1024,
		'G': 1024 * 1024 * 1024,
	}

	var size int64
	var multiplier int64 = 1

	if len(sizeStr) > 1 {
		lastChar := rune(sizeStr[len(sizeStr)-1])
		if mult, ok := multipliers[lastChar]; ok {
			multiplier = mult
			sizeStr = sizeStr[:len(sizeStr)-1]
		}
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return size * multiplier, nil
}

// getRotationConfig returns the log rotation configuration
// Controlled by environment variables:
// - TUI_LOG_MAX_SIZE: maximum size per log file (e.g., "100M", "50K", default "100M")
// - TUI_LOG_MAX_FILES: maximum number of log files to keep (default 10)
// - TUI_LOG_COMPRESS: compress old log files ("true"/"false", default "true")
func getRotationConfig() *RotationConfig {
	rotationOnce.Do(func() {
		maxSizeStr := os.Getenv("TUI_LOG_MAX_SIZE")
		if maxSizeStr == "" {
			maxSizeStr = "100M"
		}

		maxSize, err := parseSize(maxSizeStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[Logger] Invalid TUI_LOG_MAX_SIZE, using default 100M: %v\n", err)
			maxSize, _ = parseSize("100M")
		}

		maxFilesStr := os.Getenv("TUI_LOG_MAX_FILES")
		maxFiles := 10
		if maxFilesStr != "" {
			if n, err := strconv.Atoi(maxFilesStr); err == nil && n > 0 {
				maxFiles = n
			}
		}

		compressStr := os.Getenv("TUI_LOG_COMPRESS")
		compress := true
		if compressStr != "" {
			compress = strings.ToLower(compressStr) == "true"
		}

		rotationConfig = &RotationConfig{
			MaxSize:  maxSize,
			MaxFiles: maxFiles,
			Compress: compress,
			Enable:   true,
		}
	})
	return rotationConfig
}

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

// =============================================================================
// Log Rotation
// =============================================================================

// rotateLog performs log rotation if needed
func rotateLog() error {
	if globalLogFile == nil {
		return nil
	}

	config := getRotationConfig()
	if !config.Enable {
		return nil
	}

	// Get current file size
	fileInfo, err := globalLogFile.Stat()
	if err != nil {
		return err
	}

	// Check if rotation is needed
	if fileInfo.Size() < config.MaxSize {
		return nil
	}

	// Close current file
	oldPath := filepath.Join("logs", "application.log")
	if err := globalLogFile.Close(); err != nil {
		return err
	}

	// Generate rotated file name with timestamp
	timestamp := time.Now().Format("20060102_150405")
	rotatedName := fmt.Sprintf("application_%s.log", timestamp)
	rotatedPath := filepath.Join("logs", rotatedName)

	// Rename current log file
	if err := os.Rename(oldPath, rotatedPath); err != nil {
		return err
	}

	// Compress the rotated file if enabled
	if config.Compress {
		go compressLogFile(rotatedPath)
	}

	// Open new log file
	file, err := os.OpenFile(oldPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	globalLogFile = file

	// Clean up old log files
	go cleanOldLogs()

	return nil
}

// compressLogFile compresses a log file using gzip
func compressLogFile(filePath string) {
	gzipPath := filePath + ".gz"

	// Open source file
	src, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer src.Close()

	// Create gzip file
	dst, err := os.Create(gzipPath)
	if err != nil {
		return
	}
	defer dst.Close()

	// Create gzip writer
	writer := gzip.NewWriter(dst)
	defer writer.Close()

	// Copy content
	if _, err := io.Copy(writer, src); err != nil {
		return
	}

	// Remove original file
	os.Remove(filePath)
}

// cleanOldLogs removes old log files beyond the max limit
func cleanOldLogs() {
	config := getRotationConfig()
	if config.MaxFiles < 1 {
		return
	}

	logDir := "logs"
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}

	// Collect log files (both .log and .log.gz)
	var logFiles []os.DirEntry
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "application") &&
			(strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".log.gz")) &&
			name != "application.log" {
			logFiles = append(logFiles, entry)
		}
	}

	// If we have more than MaxFiles, remove oldest
	if len(logFiles) > config.MaxFiles {
		// Sort by modification time (oldest first)
		for i := 0; i < len(logFiles); i++ {
			for j := i + 1; j < len(logFiles); j++ {
				fileInfo1, _ := logFiles[i].Info()
				fileInfo2, _ := logFiles[j].Info()
				if fileInfo1.ModTime().After(fileInfo2.ModTime()) {
					logFiles[i], logFiles[j] = logFiles[j], logFiles[i]
				}
			}
		}

		// Remove excess files
		for i := 0; i < len(logFiles)-config.MaxFiles; i++ {
			filePath := filepath.Join(logDir, logFiles[i].Name())
			os.Remove(filePath)
		}
	}
}

// writeToFile writes a message to the log file with rotation check
func writeToFile(msg string) {
	if globalLogFile == nil {
		return
	}

	// Check and rotate if needed
	rotateLog()

	// Write message
	fmt.Fprintln(globalLogFile, msg)
	globalLogFile.Sync()
}

// =============================================================================
// Public Rotation API
// =============================================================================

// RotateLog manually triggers log rotation
func RotateLog() error {
	return rotateLog()
}

// GetRotationConfig returns the current rotation configuration
func GetRotationConfig() *RotationConfig {
	return getRotationConfig()
}

// CleanOldLogs manually triggers cleanup of old log files
func CleanOldLogs() {
	cleanOldLogs()
}

// GetLogFiles returns information about all log files
func GetLogFiles() ([]string, error) {
	logDir := "logs"
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "application") &&
			(strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".log.gz")) {
			fileInfo, _ := entry.Info()
			info := fmt.Sprintf("%s (%.2f MB, modified: %s)",
				name,
				float64(fileInfo.Size())/1024/1024,
				fileInfo.ModTime().Format("2006-01-02 15:04:05"))
			files = append(files, info)
		}
	}

	return files, nil
}
