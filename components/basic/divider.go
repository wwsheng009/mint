package basic

import (
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
