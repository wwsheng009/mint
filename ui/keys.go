// Package ui provides keyboard key constants for sandbox testing
package ui

import "github.com/wwsheng009/mint/runtime/platform"

// =============================================================================
// Keyboard Key Constants (re-exported from runtime/platform)
// These are used for sandbox testing and simulation
// =============================================================================

const (
	// KeyEnter is the Enter key
	KeyEnter = platform.KeyEnter
	// KeyTab is the Tab key
	KeyTab = platform.KeyTab
	// KeyEscape is the Escape key
	KeyEscape = platform.KeyEscape
	// KeyBackspace is the Backspace key
	KeyBackspace = platform.KeyBackspace
	// KeyDelete is the Delete key
	KeyDelete = platform.KeyDelete
	// KeySpace is the Space key
	KeySpace = platform.KeySpace
	// KeyUp is the Up arrow key
	KeyUp = platform.KeyUp
	// KeyDown is the Down arrow key
	KeyDown = platform.KeyDown
	// KeyLeft is the Left arrow key
	KeyLeft = platform.KeyLeft
	// KeyRight is the Right arrow key
	KeyRight = platform.KeyRight
)

// SpecialKey maps platform.SpecialKey for convenience
type SpecialKey = platform.SpecialKey
