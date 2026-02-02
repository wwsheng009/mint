package basic

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Divider Component
// =============================================================================

// DividerStyle defines the visual style of a divider
type DividerStyle int

const (
	DividerSolid   DividerStyle = iota // ───────────
	DividerDashed                      // - - - - - -
	DividerDotted                      // ·· ·· ·· ··
	DividerDouble                      // ═══════════
)

// DividerVNode represents a divider component
type DividerVNode struct {
	*ui.ElementVNode
	text         string
	dividerStyle DividerStyle
	thickness    int
}

// NewDivider creates a new divider
func NewDivider() *DividerVNode {
	return &DividerVNode{
		ElementVNode:  ui.NewElement("divider"),
		text:          "",
		dividerStyle:  DividerSolid,
		thickness:     1,
	}
}

// Divider creates a new divider node
func Divider() ui.VNode {
	return NewDivider()
}

// Builder pattern
type DividerBuilderType struct {
	node *DividerVNode
}

// DividerBuilder creates a new divider builder
func DividerBuilder() *DividerBuilderType {
	return &DividerBuilderType{node: NewDivider()}
}

// Build returns the divider ui.VNode
func (b *DividerBuilderType) Build() ui.VNode {
	return b.node
}

// Text sets the divider text (centered label)
func (b *DividerBuilderType) Text(text string) *DividerBuilderType {
	b.node.SetText(text)
	return b
}

// Style sets the divider style
func (b *DividerBuilderType) Style(style DividerStyle) *DividerBuilderType {
	b.node.SetDividerStyle(style)
	return b
}

// Thickness sets the divider thickness
func (b *DividerBuilderType) Thickness(thickness int) *DividerBuilderType {
	b.node.SetThickness(thickness)
	return b
}

// Key sets the key for diffing
func (b *DividerBuilderType) Key(key string) *DividerBuilderType {
	b.node.SetKey(key)
	return b
}

// Getters
func (d *DividerVNode) Text() string         { return d.text }
func (d *DividerVNode) DividerStyle() DividerStyle { return d.dividerStyle }
func (d *DividerVNode) Thickness() int      { return d.thickness }

// Setters
func (d *DividerVNode) SetText(text string)               { d.text = text }
func (d *DividerVNode) SetDividerStyle(style DividerStyle) { d.dividerStyle = style }
func (d *DividerVNode) SetThickness(thickness int)         { d.thickness = thickness }

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
func (d *DividerVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if d == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	width := constraints.MaxWidth
	if width <= 0 {
		width = 80 // Default width
	}

	height := d.thickness
	if height < 1 {
		height = 1
	}

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface
func (d *DividerVNode) Paint(x, y int) []paint.DrawCmd {
	if d == nil {
		return nil
	}

	style := d.Style()

	// Get width from style or use default
	width := 80
	if style.Width > 0 {
		width = style.Width
	}

	var dividerStr string
	switch d.dividerStyle {
	case DividerSolid:
		dividerStr = strings.Repeat("─", width)
	case DividerDashed:
		dividerStr = strings.Repeat("─ ", width/2)
	case DividerDotted:
		dividerStr = strings.Repeat("· ", width/2)
	case DividerDouble:
		dividerStr = strings.Repeat("═", width)
	default:
		dividerStr = strings.Repeat("─", width)
	}

	// If there's text, insert it in the middle
	if d.text != "" {
		textLen := utf8.RuneCountInString(d.text)
		if textLen < width {
			padding := (width - textLen) / 2
			leftPart := dividerStr[:padding]
			rightPart := dividerStr[padding+textLen:]
			if len(rightPart) > len(dividerStr) {
				rightPart = ""
			}
			dividerStr = leftPart + d.text + rightPart
		}
	}

	return []paint.DrawCmd{
		paint.NewTextCmd(x, y, dividerStr, style),
	}
}
