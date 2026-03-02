// Package border provides Fiber-first bordered container components.
package border

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types - Use runtime/layout types for consistency
// =============================================================================

// BorderStyle aliases for convenience
type BorderStyle = layout.BorderStyle

// Border aliases for convenience
type Border = layout.Border

// BorderStyle constants
const (
	BorderNone    = layout.BorderNone
	BorderSingle  = layout.BorderSingle
	BorderDouble  = layout.BorderDouble
	BorderRounded = layout.BorderRounded
	BorderDashed  = layout.BorderDashed
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the border container description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Border Props ===
	borderStyle BorderStyle
	borderColor style.Color
	borderLabel string

	// === Layout Props ===
	width  int
	height int
	flex   int

	// === Content ===
	child rtui.VNode

	// === Style ===
	style style.Style
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ paint.BorderInfo     = (*VNode)(nil)
)

// =============================================================================
// Constructors
// =============================================================================

// New creates a new border container VNode.
//
// Deprecated: Use container's border methods instead, e.g., stack.SingleBorder("Title").
// Border is now a native property of all containers (Stack, Grid, Wrap, Absolute).
// This wrapper component will be removed in a future version.
//
// Migration example:
//
//	Old: border.New().Label("Title").SetChild(content)
//	New: stack.NewVStack().SingleBorder("Title").SetChild(content)
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("bordered"),
		borderStyle:  BorderSingle,
		borderColor:  style.Color("blue"),
	}
}

// NewWithStyle creates a border with specified style.
func NewWithStyle(style BorderStyle) *VNode {
	b := New()
	b.borderStyle = style
	return b
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (v *VNode) Key() string {
	return v.key
}

// SetKey sets the component key - returns VNode for chaining.
func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// Tag returns the tag name.
func (v *VNode) Tag() string {
	return "bordered"
}

// Style returns the visual style.
func (v *VNode) Style() style.Style {
	return v.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.style = s
	return v
}

// Children returns child nodes.
func (v *VNode) Children() []rtui.VNode {
	if v.child == nil {
		return nil
	}
	return []rtui.VNode{v.child}
}

// SetChildren sets child nodes - returns VNode for chaining.
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	if len(children) > 0 {
		v.child = children[0]
	}
	return v
}

// GetLayer returns the rendering layer.
func (v *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return v
}

// Props returns the node properties.
func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":         v.key,
		"borderStyle": v.borderStyle,
		"borderColor": string(v.borderColor),
		"borderLabel": v.borderLabel,
		"label":       v.borderLabel, // ✨ 别名：label 映射到 borderLabel（边框标签）
		"width":       v.width,
		"height":      v.height,
		"flex":        v.flex,
		"child":       v.child,
		"style":       v.style,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p["key"].(string); ok {
		v.key = val
	}
	if val, ok := p["borderStyle"].(BorderStyle); ok {
		v.borderStyle = val
	}
	if val, ok := p["borderColor"].(string); ok {
		v.borderColor = style.Color(val)
	}
	if val, ok := p["borderLabel"].(string); ok {
		v.borderLabel = val
	}
	// ✨ 从 "label" 属性读取边框标签
	if val, ok := p["label"].(string); ok {
		v.borderLabel = val
	}
	if val, ok := p["width"].(int); ok {
		v.width = val
	}
	if val, ok := p["height"].(int); ok {
		v.height = val
	}
	if val, ok := p["flex"].(int); ok {
		v.flex = val
	}
	if val, ok := p["child"].(rtui.VNode); ok {
		v.child = val
	}
	if val, ok := p["style"].(style.Style); ok {
		v.style = val
	}
	return v
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new BorderInstance from this VNode description.
func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		"key":         v.key,
		"borderStyle": v.borderStyle,
		"borderColor": string(v.borderColor),
		"borderLabel": v.borderLabel,
		"label":       v.borderLabel, // ✨ 别名：label 映射到 borderLabel
		"width":       v.width,
		"height":      v.height,
		"flex":        v.flex,
		"child":       v.child,
		"style":       v.style,
	})
}

// =============================================================================
// Builder Methods - Fluent API
// =============================================================================

// SetBorderStyle sets the border style.
func (v *VNode) SetBorderStyle(s BorderStyle) *VNode {
	v.borderStyle = s
	return v
}

// SetBorderColor sets the border color.
func (v *VNode) SetBorderColor(c style.Color) *VNode {
	v.borderColor = c
	return v
}

// SetBorderLabel sets the border label.
func (v *VNode) SetBorderLabel(label string) *VNode {
	v.borderLabel = label
	return v
}

// SetWidth sets the explicit width.
func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

// SetHeight sets the explicit height.
func (v *VNode) SetHeight(height int) *VNode {
	v.height = height
	return v
}

// SetFlex sets the flex factor.
func (v *VNode) SetFlex(flex int) *VNode {
	v.flex = flex
	return v
}

// SetChild sets the child content.
func (v *VNode) SetChild(child rtui.VNode) *VNode {
	v.child = child
	return v
}

// SetStyleProps sets the visual style.
func (v *VNode) SetStyleProps(s style.Style) *VNode {
	v.style = s
	return v
}

// =============================================================================
// Convenience Methods
// =============================================================================

// Single sets border style to single line.
func (v *VNode) Single() *VNode {
	return v.SetBorderStyle(BorderSingle)
}

// Double sets border style to double line.
func (v *VNode) Double() *VNode {
	return v.SetBorderStyle(BorderDouble)
}

// Rounded sets border style to rounded corners.
func (v *VNode) Rounded() *VNode {
	return v.SetBorderStyle(BorderRounded)
}

// Dashed sets border style to dashed line.
func (v *VNode) Dashed() *VNode {
	return v.SetBorderStyle(BorderDashed)
}

// None removes the border.
func (v *VNode) None() *VNode {
	return v.SetBorderStyle(BorderNone)
}

// Label sets the border label (convenience method).
func (v *VNode) Label(label string) *VNode {
	return v.SetBorderLabel(label)
}

// Color sets the border color (convenience method).
func (v *VNode) Color(c string) *VNode {
	return v.SetBorderColor(style.Color(c))
}

// =============================================================================
// Accessors
// =============================================================================

// BorderStyle returns the border style.
func (v *VNode) BorderStyle() BorderStyle {
	return v.borderStyle
}

// BorderColor returns the border color.
func (v *VNode) BorderColor() style.Color {
	return v.borderColor
}

// BorderLabel returns the border label.
func (v *VNode) BorderLabel() string {
	return v.borderLabel
}

// Width returns the explicit width.
func (v *VNode) Width() int {
	return v.width
}

// Height returns the explicit height.
func (v *VNode) Height() int {
	return v.height
}

// Flex returns the flex factor.
func (v *VNode) Flex() int {
	return v.flex
}

// Child returns the child content.
func (v *VNode) Child() rtui.VNode {
	return v.child
}

// =============================================================================
// Layout Info (for border layout)
// =============================================================================

// GetBorder returns the border configuration for the layout engine.
func (v *VNode) GetBorder() layout.Border {
	return layout.Border{
		Style: v.borderStyle,
		Width: GetBorderWidth(v.borderStyle),
		Label: v.borderLabel,
	}
}

// GetBorderWidth returns the layout width of a border style.
// All border styles use single-cell characters (┌─┐│└┘ or ╔═╗║╚╝),
// so the layout width is always 1 character cell per side.
// The "double" visual effect is achieved with different glyphs, not wider cells.
func GetBorderWidth(s BorderStyle) int {
	switch s {
	case BorderNone:
		return 0
	default:
		// All visible borders occupy 1 character cell per side
		return 1
	}
}

// GetLayoutInfo returns layout information for the layout engine.
func (v *VNode) GetLayoutInfo() rtui.LayoutInfo {
	return rtui.LayoutInfo{
		Flex: v.flex,
	}
}

// MeasureConstraints measures with border constraints.
func (v *VNode) MeasureConstraints(c layout.Constraints) layout.Size {
	border := v.GetBorder()
	innerConstraints := layout.CalculateBorderConstraints(c, border)
	return layout.Size{
		Width:  innerConstraints.MinWidth,
		Height: innerConstraints.MinHeight,
	}
}

// =============================================================================
// paint.BorderInfo Interface Implementation
// =============================================================================

// GetBorderStyle implements paint.BorderInfo.
func (v *VNode) GetBorderStyle() paint.BorderStyle {
	switch v.borderStyle {
	case BorderSingle:
		return paint.BorderStyleSingle
	case BorderDouble:
		return paint.BorderStyleDouble
	case BorderRounded:
		return paint.BorderStyleRounded
	case BorderDashed:
		return paint.BorderStyleDashed
	default:
		return paint.BorderStyleNone
	}
}

// GetBorderColor implements paint.BorderInfo.
func (v *VNode) GetBorderColor() string {
	return string(v.borderColor)
}

// GetBorderLabel implements paint.BorderInfo.
func (v *VNode) GetBorderLabel() string {
	return v.borderLabel
}
