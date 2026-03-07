package paint

import (
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// tuiWidthCondition 用于非 EastAsian 环境，避免 Ambiguous 字符被错误标记为宽度 2
var tuiWidthCondition = &runewidth.Condition{
	EastAsianWidth: false,
}

// StringWidth calculates the display width of text for TUI rendering.
//
// Width is measured by grapheme cluster to avoid over-counting joined emoji
// sequences, while still forcing box-drawing runes to width 1.
func StringWidth(text string) int {
	if text == "" {
		return 0
	}

	width := 0
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		cluster := g.Str()
		runes := []rune(cluster)
		if len(runes) == 1 {
			width += RuneWidth(runes[0])
			continue
		}
		// For multi-rune clusters (emoji sequences), use conservative width
		width += tuiWidthCondition.StringWidth(cluster)
	}
	return width
}
// For Unicode Box Drawing characters (U+2500-U+257F), returns 1 instead of runewidth's 2.
func RuneWidth2(r rune) int {
	// Unicode Box Drawing block - always treat as width 1 for TUI
	if r >= 0x2500 && r <= 0x257F {
		return 1
	}
	return tuiWidthCondition.RuneWidth(r)
}
// RuneWidth calculates the display width of a single rune for TUI rendering.
//
// Strategy: Use EastAsianWidth=false to avoid ambiguous width issues,
// then explicitly handle CJK wide characters that truly need width 2.
func RuneWidth(r rune) int {
	// Unicode Box Drawing block - always treat as width 1 for TUI
	if r >= 0x2500 && r <= 0x257F {
		return 1
	}

	// CJK Unified Ideographs - these truly need width 2
	// U+4E00-U+9FFF: CJK Unified Ideographs
	// U+3400-U+4DBF: CJK Unified Ideographs Extension A
	// U+3000-U+303F: CJK Symbols and Punctuation (most are width 2)
	if (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) {
		return 2
	}

	// CJK Symbols and Punctuation (most are width 2)
	// Exception: U+3000 is ideographic space (width 2)
	if r >= 0x3000 && r <= 0x303F {
		return 2
	}

	// Fullwidth ASCII variants (U+FF00-U+FFEF)
	if r >= 0xFF01 && r <= 0xFF60 {
		return 2
	}

	// Hangul syllables (Korean)
	if r >= 0xAC00 && r <= 0xD7A3 {
		return 2
	}

	// Hangul Jamo (Korean)
	if (r >= 0x1100 && r <= 0x11FF) || (r >= 0x3130 && r <= 0x318F) {
		return 2
	}

	// Hiragana and Katakana
	if (r >= 0x3040 && r <= 0x309F) || (r >= 0x30A0 && r <= 0x30FF) {
		return 2
	}

	// For all other characters, use non-EastAsian width
	// This treats most symbols, arrows, etc. as width 1
	rw := tuiWidthCondition.RuneWidth(r)
	if rw <= 0 {
		return 1
	}
	return rw
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
