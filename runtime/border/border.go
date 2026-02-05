// Package border provides modular border rendering for TUI components.
// Borders are decorative and do not participate in layout calculations.
package border

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime/style"
)

// Style defines the visual style of borders
type Style int

const (
	StyleSingle Style = iota // ┌─┐ continuous line
	StyleDouble              // ╔═╗ double line
	StyleRounded             // ╭─╮ rounded corners
	StyleDashed              // +-+ ASCII style
	StyleNone                // No border
)

// Config holds border configuration
type Config struct {
	Style  Style   // Border visual style
	Color  string  // Border color name
	Label  string  // Optional title on top border
	Width  int     // Border width (currently only 1 is supported)
	PadTop int     // Top padding inside border
}

// DefaultConfig returns a default border configuration
func DefaultConfig() Config {
	return Config{
		Style:  StyleSingle,
		Color:  "blue",
		Label:  "",
		Width:  1,
		PadTop: 0,
	}
}

// Renderer handles border rendering
type Renderer struct {
	config Config
}

// New creates a new renderer with default config
func New() *Renderer {
	return &Renderer{config: DefaultConfig()}
}

// WithConfig creates a renderer with custom config
func WithConfig(cfg Config) *Renderer {
	return &Renderer{config: cfg}
}

// WithStyle sets the border style
func (r *Renderer) WithStyle(s Style) *Renderer {
	r.config.Style = s
	return r
}

// WithColor sets the border color
func (r *Renderer) WithColor(c string) *Renderer {
	r.config.Color = c
	return r
}

// WithLabel sets the border label
func (r *Renderer) WithLabel(l string) *Renderer {
	r.config.Label = l
	return r
}

// GetBorderChars returns the border characters for the current style
func (r *Renderer) GetBorderChars() (cornerTL, cornerTR, cornerBL, cornerBR, horizontal, vertical rune) {
	switch r.config.Style {
	case StyleDouble:
		return '╔', '╗', '╚', '╝', '═', '║'
	case StyleRounded:
		return '╭', '╮', '╰', '╯', '─', '│'
	case StyleDashed:
		return '+', '+', '+', '+', '-', '|'
	case StyleNone:
		return ' ', ' ', ' ', ' ', ' ', ' '
	default: // StyleSingle
		return '┌', '┐', '└', '┘', '─', '│'
	}
}

// GetBorderStyle returns the style for border rendering
func (r *Renderer) GetBorderStyle() style.Style {
	if r.config.Color == "" {
		return style.NewStyle()
	}
	return style.NewStyle().Foreground(style.Color(r.config.Color))
}

// GetLabelStyle returns the style for label rendering
func (r *Renderer) GetLabelStyle() style.Style {
	base := r.GetBorderStyle()
	return base.Bold(true)
}

// Cell is a renderable cell position
type Cell struct {
	Ch   rune
	X, Y int
	Style style.Style
}

// CellsAtPosition represents cells at a specific position
type CellsAtPosition struct {
	Cells []Cell
	X, Y  int // Base position
}

// Render generates border cells around the given content area
// The content area is defined by (x, y, width, height)
// Border is rendered OUTSIDE the content area
func (r *Renderer) Render(x, y, contentWidth, contentHeight int) []Cell {
	if r.config.Style == StyleNone {
		return nil
	}

	// Ensure non-negative dimensions
	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	var cells []Cell
	borderStyle := r.GetBorderStyle()
	labelStyle := r.GetLabelStyle()

	// Get border characters
	cornerTL, cornerTR, cornerBL, cornerBR, horizontal, vertical := r.GetBorderChars()

	// === Top border (with optional label) ===
	if r.config.Label == "" {
		// Simple top border: cornerTL + horizontal + cornerTR
		cells = append(cells, Cell{cornerTL, x, y, borderStyle})
		for i := 0; i < contentWidth; i++ {
			cells = append(cells, Cell{horizontal, x + 1 + i, y, borderStyle})
		}
		cells = append(cells, Cell{cornerTR, x + contentWidth + 1, y, borderStyle})
	} else {
		// Top border with label
		labelRunes := []rune(" " + r.config.Label + " ")
		labelWidth := len(labelRunes)
		padding := (contentWidth - labelWidth) / 2
		if padding < 0 {
			padding = 0
		}

		// Corner TL + horizontal padding
		cells = append(cells, Cell{cornerTL, x, y, borderStyle})
		for i := 0; i < padding+1; i++ {
			cells = append(cells, Cell{horizontal, x + 1 + i, y, borderStyle})
		}

		// Label
		labelX := x + 1 + padding + 1
		for i, ch := range labelRunes {
			cells = append(cells, Cell{ch, labelX + i, y, labelStyle})
		}

		// Remaining horizontal + corner TR
		remainingX := labelX + labelWidth
		remainingCount := contentWidth - padding - labelWidth + 2
		if remainingCount < 0 {
			remainingCount = 0
		}
		for i := 0; i < remainingCount; i++ {
			cells = append(cells, Cell{horizontal, remainingX + i, y, borderStyle})
		}
		cells = append(cells, Cell{cornerTR, x + contentWidth + 1, y, borderStyle})
	}

	// === Middle rows (left border + content area + right border) ===
	for row := 0; row < contentHeight; row++ {
		rowY := y + 1 + row
		// Left border
		cells = append(cells, Cell{vertical, x, rowY, borderStyle})
		// Right border
		cells = append(cells, Cell{vertical, x + contentWidth + 1, rowY, borderStyle})
	}

	// === Bottom border ===
	bottomY := y + 1 + contentHeight
	cells = append(cells, Cell{cornerBL, x, bottomY, borderStyle})
	for i := 0; i < contentWidth; i++ {
		cells = append(cells, Cell{horizontal, x + 1 + i, bottomY, borderStyle})
	}
	cells = append(cells, Cell{cornerBR, x + contentWidth + 1, bottomY, borderStyle})

	return cells
}

// Paint paints border cells to a buffer using the provided paint function
// The paint function has signature: func(x, y int, ch rune, s style.Style)
func (r *Renderer) Paint(x, y, contentWidth, contentHeight int, paintFunc func(int, int, rune, style.Style)) {
	cells := r.Render(x, y, contentWidth, contentHeight)
	for _, cell := range cells {
		paintFunc(cell.X, cell.Y, cell.Ch, cell.Style)
	}
}

// GetTotalSize returns the total size including border
// For content 5x1, returns (7, 3) - border adds 2 in each dimension
func (r *Renderer) GetTotalSize(contentWidth, contentHeight int) (width, height int) {
	if r.config.Style == StyleNone {
		return contentWidth, contentHeight
	}
	return contentWidth + 2, contentHeight + 2
}

// GetContentOffset returns the offset where content should be placed
// inside a bordered container. For border, content starts at (1, 1).
func (r *Renderer) GetContentOffset() (x, y int) {
	if r.config.Style == StyleNone {
		return 0, 0
	}
	return 1, 1
}

// String returns a string representation of the border around content
// This is useful for debugging and testing
func (r *Renderer) String(content string) string {
	if content == "" {
		content = " "
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return ""
	}

	// Find max line width
	maxWidth := 0
	for _, line := range lines {
		width := utf8.RuneCountInString(line)
		if width > maxWidth {
			maxWidth = width
		}
	}

	contentWidth := maxWidth
	_ = len(lines) // contentHeight

	// Get border characters
	cornerTL, cornerTR, cornerBL, cornerBR, horizontal, vertical := r.GetBorderChars()

	var result strings.Builder

	// Top border
	if r.config.Label == "" {
		result.WriteRune(cornerTL)
		result.WriteString(strings.Repeat(string(horizontal), contentWidth))
		result.WriteRune(cornerTR)
	} else {
		labelRunes := []rune(" " + r.config.Label + " ")
		labelWidth := len(labelRunes)
		padding := (contentWidth - labelWidth) / 2
		if padding < 0 {
			padding = 0
		}
		result.WriteRune(cornerTL)
		result.WriteString(strings.Repeat(string(horizontal), padding+1))
		result.WriteString(string(labelRunes))
		remainingWidth := contentWidth - padding - labelWidth + 2
		if remainingWidth < 0 {
			remainingWidth = 0
		}
		result.WriteString(strings.Repeat(string(horizontal), remainingWidth))
		result.WriteRune(cornerTR)
	}
	result.WriteRune('\n')

	// Content with borders
	for _, line := range lines {
		result.WriteRune(vertical)
		result.WriteString(line)
		padding := contentWidth - utf8.RuneCountInString(line)
		if padding > 0 {
			result.WriteString(strings.Repeat(" ", padding))
		}
		result.WriteRune(vertical)
		result.WriteRune('\n')
	}

	// Bottom border
	result.WriteRune(cornerBL)
	result.WriteString(strings.Repeat(string(horizontal), contentWidth))
	result.WriteRune(cornerBR)

	return result.String()
}
