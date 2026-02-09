package paint

import (
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

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
		width += tuiWidthCondition.StringWidth(cluster)
	}
	return width
}

// RuneWidth calculates the display width of a single rune for TUI rendering.
//
// For Unicode Box Drawing characters (U+2500-U+257F), returns 1 instead of runewidth's 2.
func RuneWidth(r rune) int {
	// Unicode Box Drawing block - always treat as width 1 for TUI
	if r >= 0x2500 && r <= 0x257F {
		return 1
	}
	return tuiWidthCondition.RuneWidth(r)
}
