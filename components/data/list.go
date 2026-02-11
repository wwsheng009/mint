package data

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// ListVNode represents a general-purpose list component.
//
// Unlike VirtualList (which requires []interface{} items and a renderItem callback),
// ListVNode works directly with pre-formatted string rows and a column header.
// Each row is truncated to fit within the measured width from layout constraints,
// preventing horizontal overflow in constrained containers like the Inspector panel.
//
// Features:
//   - Respects BoxConstraints from the layout engine (Measure interface)
//   - Truncates each row to the available width with "…" suffix
//   - Optional header row with distinct style
//   - Optional separator line between header and data rows
//   - Supports per-row styling via RowStyle callback
//
// Usage:
//
//	list := data.ListBuilder().
//	    Header("Z   Node            Bounds       H  C").
//	    Rows(rows).
//	    Build()
type ListVNode struct {
	*ui.ElementVNode

	header      string                        // Optional column header text
	rows        []string                      // Pre-formatted data rows
	headerStyle style.Style                   // Style for the header row
	rowStyle    style.Style                   // Default style for data rows
	rowStyleFn  func(int, string) style.Style // Per-row style callback (index, text) → style
	showSep     bool                          // Show separator line between header and rows
	sepChar     rune                          // Separator character (default '─')
	emptyText   string                        // Text shown when rows is empty
	maxRows     int                           // Maximum visible rows (0 = unlimited)

	// Measured width from layout constraints (set during Measure)
	measuredWidth int
}

// ListBuilderType provides a fluent API for building lists.
type ListBuilderType struct {
	node *ListVNode
}

// ListBuilder creates a new list builder.
func ListBuilder() *ListBuilderType {
	return &ListBuilderType{
		node: &ListVNode{
			ElementVNode: ui.NewElement("list"),
			headerStyle:  style.Style{}.Bold(true),
			rowStyle:     style.Style{},
			showSep:      true,
			sepChar:      '─',
			emptyText:    "(empty)",
			maxRows:      0,
		},
	}
}

// Header sets the column header text.
func (b *ListBuilderType) Header(h string) *ListBuilderType {
	b.node.header = h
	return b
}

// Rows sets all data rows at once.
func (b *ListBuilderType) Rows(rows []string) *ListBuilderType {
	b.node.rows = rows
	return b
}

// AddRow appends a single row.
func (b *ListBuilderType) AddRow(row string) *ListBuilderType {
	b.node.rows = append(b.node.rows, row)
	return b
}

// HeaderStyle sets the style for the header row.
func (b *ListBuilderType) HeaderStyle(s style.Style) *ListBuilderType {
	b.node.headerStyle = s
	return b
}

// RowStyle sets the default style for data rows.
func (b *ListBuilderType) RowStyle(s style.Style) *ListBuilderType {
	b.node.rowStyle = s
	return b
}

// RowStyleFn sets a per-row style callback.
// The function receives the row index and text, and returns the style to use.
func (b *ListBuilderType) RowStyleFn(fn func(int, string) style.Style) *ListBuilderType {
	b.node.rowStyleFn = fn
	return b
}

// ShowSeparator controls whether a separator line is drawn between header and rows.
func (b *ListBuilderType) ShowSeparator(show bool) *ListBuilderType {
	b.node.showSep = show
	return b
}

// SepChar sets the separator character (default '─').
func (b *ListBuilderType) SepChar(ch rune) *ListBuilderType {
	b.node.sepChar = ch
	return b
}

// EmptyText sets the text displayed when there are no rows.
func (b *ListBuilderType) EmptyText(text string) *ListBuilderType {
	b.node.emptyText = text
	return b
}

// MaxRows limits the number of visible rows (0 = show all).
func (b *ListBuilderType) MaxRows(n int) *ListBuilderType {
	b.node.maxRows = n
	return b
}

// Width sets an explicit width hint (overridden by layout constraints).
func (b *ListBuilderType) Width(w int) *ListBuilderType {
	b.node.SetProp("width", w)
	return b
}

// Style sets the base visual style.
func (b *ListBuilderType) Style(s style.Style) *ListBuilderType {
	b.node.SetStyle(s)
	return b
}

// Key sets the key for reconciliation diffing.
func (b *ListBuilderType) Key(key string) *ListBuilderType {
	b.node.SetKey(key)
	return b
}

// Build returns the list VNode.
func (b *ListBuilderType) Build() ui.VNode {
	return b.node
}

// =============================================================================
// Getters / Setters
// =============================================================================

func (l *ListVNode) Header() string                                 { return l.header }
func (l *ListVNode) Rows() []string                                 { return l.rows }
func (l *ListVNode) RowCount() int                                  { return len(l.rows) }
func (l *ListVNode) SetHeader(h string)                             { l.header = h }
func (l *ListVNode) SetRows(rows []string)                          { l.rows = rows }
func (l *ListVNode) SetRowStyleFn(fn func(int, string) style.Style) { l.rowStyleFn = fn }
func (l *ListVNode) EmptyText() string                              { return l.emptyText }

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable.
//
// It calculates the natural size of the list (widest row × total lines)
// and then clamps to the constraints provided by the layout engine.
// The measured width is cached so Paint() can truncate rows accordingly.
func (l *ListVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if l == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Calculate natural width = max rune width of all lines
	naturalWidth := 0
	updateMax := func(s string) {
		if n := utf8.RuneCountInString(s); n > naturalWidth {
			naturalWidth = n
		}
	}
	if l.header != "" {
		updateMax(l.header)
	}
	for _, r := range l.rows {
		updateMax(r)
	}
	if naturalWidth < utf8.RuneCountInString(l.emptyText) && len(l.rows) == 0 {
		naturalWidth = utf8.RuneCountInString(l.emptyText)
	}

	// Explicit width prop overrides natural width
	if w, ok := l.Props()["width"].(int); ok && w > 0 {
		naturalWidth = w
	}

	// Calculate natural height
	naturalHeight := l.visibleLineCount()
	// Apply constraints
	width := naturalWidth
	height := naturalHeight

	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if constraints.HasBoundedWidth() && width > constraints.MaxWidth {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if constraints.HasBoundedHeight() && height > constraints.MaxHeight {
		height = constraints.MaxHeight
	}

	// Cache measured width for Paint
	l.measuredWidth = width

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable.
//
// It renders the list as a series of DrawCmd, one per visible line.
// Each line is truncated to measuredWidth to prevent horizontal overflow.
func (l *ListVNode) Paint(x, y int) []paint.DrawCmd {
	if l == nil {
		return nil
	}

	// Determine effective width for truncation
	width := l.measuredWidth
	if width <= 0 {
		// Fallback: use explicit prop or default
		if w, ok := l.Props()["width"].(int); ok && w > 0 {
			width = w
		} else {
			width = 80
		}
	}

	var cmds []paint.DrawCmd
	row := 0

	// Header
	if l.header != "" {
		truncated := truncateRunes(l.header, width)
				cmds = append(cmds, paint.NewTextCmd(x, y+row, truncated, l.headerStyle))
		row++
	}

	// Separator (only if there are rows)
	if l.showSep && l.header != "" && len(l.rows) > 0 {
		sep := strings.Repeat(string(l.sepChar), width)
		cmds = append(cmds, paint.NewTextCmd(x, y+row, sep, l.Style()))
		row++
	}

	// Data rows
	if len(l.rows) == 0 {
		cmds = append(cmds, paint.NewTextCmd(x, y+row, truncateRunes(l.emptyText, width), l.Style()))
		row++
	} else {
		visibleRows := l.rows
		overflow := false
		if l.maxRows > 0 && len(visibleRows) > l.maxRows {
			visibleRows = visibleRows[:l.maxRows]
			overflow = true
		}

		for i, r := range visibleRows {
			s := l.rowStyle
			if l.rowStyleFn != nil {
				s = l.rowStyleFn(i, r)
			}
			truncated := truncateRunes(r, width)
						cmds = append(cmds, paint.NewTextCmd(x, y+row, truncated, s))
			row++
		}

		if overflow {
			moreText := truncateRunes("... (more rows)", width)
			cmds = append(cmds, paint.NewTextCmd(x, y+row, moreText, l.Style()))
			row++
		}
	}

	return cmds
}

// =============================================================================
// Internal helpers
// =============================================================================

// visibleLineCount returns the total number of lines that will be rendered.
func (l *ListVNode) visibleLineCount() int {
	n := 0
	if l.header != "" {
		n++ // header
	}
	// Only show separator if there are rows (empty list doesn't need separator)
	if l.showSep && l.header != "" && len(l.rows) > 0 {
		n++ // separator
	}
	if len(l.rows) == 0 {
		n++ // empty text
	} else {
		visible := len(l.rows)
		if l.maxRows > 0 && visible > l.maxRows {
			visible = l.maxRows
			n++ // "... (more rows)" line
		}
		n += visible
	}
	return n
}

// truncateRunes truncates a string to at most maxRunes rune-width.
// If truncated, the last character is replaced with "…".
func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		// Pad with spaces to fill the width so the background style
		// covers the full row (consistent with Table behaviour).
		if len(runes) < maxRunes {
			return s + strings.Repeat(" ", maxRunes-len(runes))
		}
		return s
	}
	if maxRunes <= 1 {
		return "…"
	}
	// Keep maxRunes-1 characters and add ellipsis
	return string(runes[:maxRunes-1]) + "…"
}
