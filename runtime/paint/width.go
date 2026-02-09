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
// IMPORTANT: This is the original implementation maintained for compatibility
// with the layout engine. For TUI-character-safe alternatives, see
// CellWidthOfRune() and SanitizeForTerminal().
func StringWidth(text string) int {
	// Check if this is a single character
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

// ============================================================================
// TUI ALTERNATIVE WIDTH FUNCTIONS (for safe character handling)
// ============================================================================

// CellWidthOfRune returns the terminal cell width of a single rune.
//
// Use this for layout calculations that need to account for multi-rune emojis.
// For most TUI rendering, use StringWidth() instead for consistency.
func CellWidthOfRune(r rune) int {
	return runewidth.RuneWidth(r)
}

// CellWidthOfString returns the total number of cells a string will occupy.
//
// This method calculates width by summing individual rune widths, which
// correctly handles multi-rune emojis (🖼️) that occupy multiple cells.
//
// Example:
//
//	"🖼️"   = U+1F5BC + U+FE0F = 2 runes = 2 cells (each width 1)
//	"📦"   = U+1F4E6           = 1 rune  = 2 cells (wide char)
//	"↑↓"   = 2 unicode arrows = 2 cells
//	"text" = 4 ASCII chars     = 4 cells
func CellWidthOfString(s string) int {
	w := 0
	for _, r := range s {
		w += CellWidthOfRune(r)
	}
	return w
}

// ============================================================================
// TUI CHARACTER SAFETY
// ============================================================================

// SanitizeForTerminal removes characters that are unsafe for TUI rendering.
//
// This function implements the TUI Character Safety Layer as described in
// docs/TUI_BUFFER_FIX.md. It removes:
//
// 1. Variation Selectors (U+FE0F-U+FE0F) - causes multi-rune emojis
// 2. Zero Width Joiners (U+200D) - causes ZWJ sequences
// 3. Combining Marks - modifies preceding character
// 4. Control characters (except \n, \t)
//
// Usage: All UI text should pass through this function before writing to Buffer
//
// Example:
//
//	icon = "🖼️"
//	safe := SanitizeForTerminal(icon) → "🖼" (VS16 removed)
func SanitizeForTerminal(s string) string {
	var out []rune

	for _, r := range s {
		// Filter: Variation Selector 16 (VS16)
		if r == 0xFE0F {
			continue
		}

		// Filter: Zero Width Joiner (ZWJ)
		if r == 0x200D {
			continue
		}

		// Filter: Combining Marks
		// Mn = Mark, Nonspacing (accents, diacritics, etc.)
		if (r >= 0x0300 && r <= 0x036F) ||
			(r >= 0x1AB0 && r <= 0x1AFF) ||
			(r >= 0x1DC0 && r <= 0x1DFF) ||
			(r >= 0xFE20 && r <= 0xFE2F) {
			continue
		}

		// Filter: Control characters (keep newline, tab)
		if r < 32 && r != '\n' && r != '\t' && r != '\r' {
			continue
		}

		out = append(out, r)
	}

	return string(out)
}

// TUI SAFETY WARNING:
//
// ┌─────────────────────────────────────────────────────────┐
// │  DO NOT use multi-rune emojis in TUI                    │
// │  Forbidden examples: 🖼️ ⚙️ ✏️ ☑️ 👨‍👩‍👧‍👦           │
// │                                                          │
// │  Use single-rune alternatives instead:                  │
// │  🖼️ → 🎨  ⚙️ → ⚙  ✏️ → ✏  ☑️ → ☑                │
// │                                                          │
// │  See: docs/TUI_BUFFER_FIX.md                          │
// └─────────────────────────────────────────────────────────┘
