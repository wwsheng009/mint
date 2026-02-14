package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

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
	logDir, _ := filepath.Abs("./logs")
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
