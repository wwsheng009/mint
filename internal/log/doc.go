// Package log provides structured logging for the mint TUI framework.
//
// # Environment Variables
//
// Logging is controlled via environment variables:
//
// ## Debug Level Control
//
//	TUI_DEBUG=true          - Enable all debug logging
//	TUI_DEBUG_FOCUS=true    - Enable focus-related logging
//	TUI_DEBUG_RECONCILER=true - Enable reconciler logging
//	TUI_DEBUG_RENDER=true   - Enable render logging
//	TUI_DEBUG_KEYS=true     - Enable key event logging
//	TUI_DEBUG_UI=true       - Enable UI component logging
//
// ## Log Output Destination
//
//	TUI_LOG_OUTPUT=file     - Output to file only (default)
//	TUI_LOG_OUTPUT=console  - Output to console only
//	TUI_LOG_OUTPUT=both     - Output to both file and console
//
// ## Log Rotation
//
//	TUI_LOG_MAX_SIZE=100M   - Maximum size per log file (default: 100M)
//	TUI_LOG_MAX_FILES=10    - Maximum number of log files to keep (default: 10)
//	TUI_LOG_COMPRESS=true   - Compress old log files using gzip (default: true)
//
// # Usage
//
//	import "github.com/wwsheng009/mint/internal/log"
//
//	// Enable debug logging
//	log.SetAllEnabled(true)
//
//	// Use domain-specific loggers
//	log.FocusLogger.Debug("Focus moved to index %d", index)
//	log.FiberLogger.Debug("Starting reconciliation")
//	log.RenderLogger.Info("Render completed in %dms", duration)
//
//	// Close log file on application exit
//	defer log.CloseLogFile()
//
// # Log File Management
//
// Logs are written to logs/application.log by default. When the file size exceeds
// TUI_LOG_MAX_SIZE, it is automatically rotated:
//
//   - Current log: logs/application.log
//   - Rotated log: logs/application_20060102_150405.log
//   - Compressed log: logs/application_20060102_150405.log.gz
//
// Old log files are automatically cleaned up, keeping only the most recent
// TUI_LOG_MAX_FILES files.
//
// # Manual Log Management
//
//	// Get current rotation configuration
//	config := log.GetRotationConfig()
//
//	// Manually trigger log rotation
//	log.RotateLog()
//
//	// Clean up old log files
//	log.CleanOldLogs()
//
//	// Get information about all log files
//	files, err := log.GetLogFiles()
//	for _, file := range files {
//	    fmt.Println(file)
//	}
//
// # Global Loggers
//
// The package provides pre-configured loggers for different domains:
//
//	log.FocusLogger      - Focus management events
//	log.FiberLogger - Reconciler operations
//	log.RenderLogger     - Rendering operations
//	log.KeyLogger        - Keyboard events
//	log.EventLogger      - Event system events
//	log.UILogger         - UI component events
//	log.ButtonLogger     - Button-specific events
//	log.WinLogger        - Window events
//	log.InspectorLogger  - Inspector events
//	log.EngineLogger     - Engine events
//
// # Thread Safety
//
// All loggers are thread-safe and can be used from goroutines.
package log
