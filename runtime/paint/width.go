package paint

import (
	"github.com/mattn/go-runewidth"
)

// StringWidth calculates the display width of text for TUI rendering.
//
// For Unicode Box Drawing characters (U+2500-U+257F), returns 1 instead of runewidth's 2.
// This ensures correct cursor tracking and cell width calculation in TUI.
//
// For multi-character clusters (emoji, etc.) or regular text, uses runewidth.
//
// This is the recommended function for all TUI width calculations.
func StringWidth(text string) int {
	runes := []rune(text)
	if len(runes) == 1 {
		ch := runes[0]
		// Unicode Box Drawing block - always treat as width 1 for TUI
		if ch >= 0x2500 && ch <= 0x257F {
			return 1
		}
	}
	// For multi-character clusters (emoji, etc.) or regular text, use runewidth
	return runewidth.StringWidth(text)
}

// RuneWidth calculates the display width of a single rune for TUI rendering.
//
// For Unicode Box Drawing characters (U+2500-U+257F), returns 1 instead of runewidth's 2.
func RuneWidth(r rune) int {
	// Unicode Box Drawing block - always treat as width 1 for TUI
	if r >= 0x2500 && r <= 0x257F {
		return 1
	}
	return runewidth.RuneWidth(r)
}
