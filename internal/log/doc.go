// Package log provides structured logging for the mint TUI framework.
//
// # Environment Variables
//
// Logging is controlled via environment variables:
//
//	TUI_DEBUG=true          - Enable all debug logging
//	TUI_DEBUG_FOCUS=true    - Enable focus-related logging
//	TUI_DEBUG_RECONCILER=true - Enable reconciler logging
//	TUI_DEBUG_RENDER=true   - Enable render logging
//	TUI_DEBUG_KEYS=true     - Enable key event logging
//	TUI_DEBUG_UI=true       - Enable UI component logging
//
// # Usage
//
//	import "github.com/wwsheng009/mint/internal/log"
//
//	log.FocusLogger.Debug("Focus moved to index %d", index)
//	log.ReconcilerLogger.Debug("Starting reconciliation")
//
// # Global Loggers
//
// The package provides pre-configured loggers for different domains:
//
//	log.FocusLogger      - Focus management events
//	log.ReconcilerLogger - Reconciler operations
//	log.RenderLogger     - Rendering operations
//	log.KeyLogger        - Keyboard events
//	log.UILogger         - UI component events
//	log.ButtonLogger     - Button-specific events
//
// # Thread Safety
//
// All loggers are thread-safe and can be used from goroutines.
package log
