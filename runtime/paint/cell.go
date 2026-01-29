package paint

import "github.com/wwsheng009/mint/runtime/style"

// Cell represents a single terminal cell with content and style.
type Cell struct {
	// Char is the rune character to display.
	Char rune

	// Style is the visual style (color, attributes) for this cell.
	// We use the framework's style definition for consistency.
	Style style.Style

	// Width is the visual width of the character (usually 1 or 2).
	Width int

	// IsContinuation marks this cell as a continuation of a wide character
	// in the previous cell. When IsContinuation is true, this cell should
	// be skipped during output and diff operations.
	IsContinuation bool
}
