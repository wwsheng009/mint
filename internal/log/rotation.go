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

// =============================================================================
// Log Rotation
// =============================================================================

var (
	// Log rotation configuration
	rotationConfig *RotationConfig
	rotationOnce   sync.Once
)

// RotationConfig holds log rotation settings
type RotationConfig struct {
	MaxSize  int64 // Maximum size in bytes before rotation
	MaxFiles int   // Maximum number of log files to keep
	Compress bool  // Whether to compress old log files
	Enable   bool  // Whether rotation is enabled
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
