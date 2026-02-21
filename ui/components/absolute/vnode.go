// Package absolute provides Fiber-first absolute positioning component.
package absolute

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Position Types (VNode layer - declarative only)
// =============================================================================

// PositionValue represents a position value (absolute or relative).
type PositionValue interface {
	isPositionValue()
	Resolve(containerSize int) int
}

// AbsolutePos is a fixed position in cells.
type AbsolutePos int

func (a AbsolutePos) isPositionValue() {}
func (a AbsolutePos) Resolve(_ int) int { return int(a) }

// RelativePos is a percentage (0-100).
type RelativePos int

func (r RelativePos) isPositionValue() {}
func (r RelativePos) Resolve(containerSize int) int { return containerSize * int(r) / 100 }

// =============================================================================
// Anchor Types - Aliases for layout.Anchor
// =============================================================================

// Anchor is an alias for layout.Anchor for positioning alignment.
// This ensures type compatibility across the layout engine.
type Anchor = layout.Anchor

// Anchor constants - aliases for layout.Anchor values
const (
	AnchorTopLeft     Anchor = layout.AnchorTopLeft
	AnchorTop         Anchor = layout.AnchorTop
	AnchorTopRight    Anchor = layout.AnchorTopRight
	AnchorLeft        Anchor = layout.AnchorLeft
	AnchorCenter      Anchor = layout.AnchorCenter
	AnchorRight       Anchor = layout.AnchorRight
	AnchorBottomLeft  Anchor = layout.AnchorBottomLeft
	AnchorBottom      Anchor = layout.AnchorBottom
	AnchorBottomRight Anchor = layout.AnchorBottomRight
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the absolute positioning component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Child ===
	child rtui.VNode

	// === Position Props ===
	left   PositionValue
	top    PositionValue
	right  PositionValue
	bottom PositionValue
	anchor Anchor

	// === Sizing Props ===
	width  int // 0 = auto
	height int // 0 = auto
	zIndex int // stacking order
	flex   int // flex factor

	// === Style ===
	style style.Style
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructors
// =============================================================================

// New creates a new Absolute VNode.
func New(child rtui.VNode) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("absolute"),
		child:        child,
		anchor:       AnchorTopLeft,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (a *VNode) Key() string {
	return a.key
}

// SetKey sets the component key - returns VNode for chaining.
func (a *VNode) SetKey(key string) rtui.VNode {
	a.key = key
	return a
}

// Tag returns the tag name.
func (a *VNode) Tag() string {
	return "absolute"
}

// Style returns the visual style.
func (a *VNode) Style() style.Style {
	return a.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (a *VNode) SetStyle(st style.Style) rtui.VNode {
	a.style = st
	return a
}

// Children returns child nodes.
func (a *VNode) Children() []rtui.VNode {
	if a.child == nil {
		return nil
	}
	return []rtui.VNode{a.child}
}

// SetChildren sets child nodes - returns VNode for chaining.
func (a *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	if len(children) > 0 {
		a.child = children[0]
	}
	return a
}

// GetLayer returns the rendering layer.
func (a *VNode) GetLayer() rtui.Layer {
	return rtui.LayerOverlay
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (a *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return a
}

// Props returns the node properties.
func (a *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":    a.key,
		"child":  a.child,
		"left":   a.left,
		"top":    a.top,
		"right":  a.right,
		"bottom": a.bottom,
		"anchor": a.anchor,
		"width":  a.width,
		"height": a.height,
		"zIndex": a.zIndex,
		"flex":   a.flex,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (a *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		a.key = v
	}
	if v, ok := p["child"].(rtui.VNode); ok {
		a.child = v
	}
	if v, ok := p["left"].(PositionValue); ok {
		a.left = v
	}
	if v, ok := p["top"].(PositionValue); ok {
		a.top = v
	}
	if v, ok := p["right"].(PositionValue); ok {
		a.right = v
	}
	if v, ok := p["bottom"].(PositionValue); ok {
		a.bottom = v
	}
	if v, ok := p["anchor"].(Anchor); ok {
		a.anchor = v
	}
	if v, ok := p["width"].(int); ok {
		a.width = v
	}
	if v, ok := p["height"].(int); ok {
		a.height = v
	}
	if v, ok := p["zIndex"].(int); ok {
		a.zIndex = v
	}
	if v, ok := p["flex"].(int); ok {
		a.flex = v
	}
	return a
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new AbsoluteInstance from this VNode description.
func (a *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		"key":    a.key,
		"child":  a.child,
		"left":   a.left,
		"top":    a.top,
		"right":  a.right,
		"bottom": a.bottom,
		"anchor": a.anchor,
		"width":  a.width,
		"height": a.height,
		"zIndex": a.zIndex,
		"flex":   a.flex,
		"style":  a.style,
	})
}

// =============================================================================
// Builder Methods - Fluent API
// =============================================================================

// SetChild sets the child node.
func (a *VNode) SetChild(child rtui.VNode) *VNode {
	a.child = child
	return a
}

// SetLeft sets the left position.
func (a *VNode) SetLeft(pos PositionValue) *VNode {
	a.left = pos
	return a
}

// SetTop sets the top position.
func (a *VNode) SetTop(pos PositionValue) *VNode {
	a.top = pos
	return a
}

// SetRight sets the right position.
func (a *VNode) SetRight(pos PositionValue) *VNode {
	a.right = pos
	return a
}

// SetBottom sets the bottom position.
func (a *VNode) SetBottom(pos PositionValue) *VNode {
	a.bottom = pos
	return a
}

// SetAnchor sets the anchor point.
func (a *VNode) SetAnchor(anchor Anchor) *VNode {
	a.anchor = anchor
	return a
}

// SetWidth sets the explicit width.
func (a *VNode) SetWidth(width int) *VNode {
	a.width = width
	return a
}

// SetHeight sets the explicit height.
func (a *VNode) SetHeight(height int) *VNode {
	a.height = height
	return a
}

// SetZIndex sets the z-index (stacking order).
func (a *VNode) SetZIndex(z int) *VNode {
	a.zIndex = z
	return a
}

// SetFlex sets the flex factor.
func (a *VNode) SetFlex(flex int) *VNode {
	a.flex = flex
	return a
}

// =============================================================================
// Convenience Builder Methods
// =============================================================================

// Position sets left and top positions.
func (a *VNode) Position(left, top PositionValue) *VNode {
	a.left = left
	a.top = top
	return a
}

// Size sets width and height.
func (a *VNode) Size(width, height int) *VNode {
	a.width = width
	a.height = height
	return a
}

// CenterAnchor sets anchor to center.
func (a *VNode) CenterAnchor() *VNode {
	a.anchor = AnchorCenter
	return a
}

// TopLeftAnchor sets anchor to top-left.
func (a *VNode) TopLeftAnchor() *VNode {
	a.anchor = AnchorTopLeft
	return a
}

// =============================================================================
// Props Accessors
// =============================================================================

// Child returns the child node.
func (a *VNode) Child() rtui.VNode {
	return a.child
}

// LeftPos returns the left position.
func (a *VNode) LeftPos() PositionValue {
	return a.left
}

// TopPos returns the top position.
func (a *VNode) TopPos() PositionValue {
	return a.top
}

// RightPos returns the right position.
func (a *VNode) RightPos() PositionValue {
	return a.right
}

// BottomPos returns the bottom position.
func (a *VNode) BottomPos() PositionValue {
	return a.bottom
}

// AnchorPoint returns the anchor point.
func (a *VNode) AnchorPoint() Anchor {
	return a.anchor
}

// AbsWidth returns the explicit width.
func (a *VNode) AbsWidth() int {
	return a.width
}

// AbsHeight returns the explicit height.
func (a *VNode) AbsHeight() int {
	return a.height
}

// ZIndex returns the z-index.
func (a *VNode) ZIndex() int {
	return a.zIndex
}

// Flex returns the flex factor.
func (a *VNode) Flex() int {
	return a.flex
}

// =============================================================================
// Position Calculation
// =============================================================================

// CalculatePosition calculates the actual x, y position based on container size.
func (a *VNode) CalculatePosition(containerWidth, containerHeight int) (int, int) {
	x := 0
	y := 0

	// Calculate X position
	if a.left != nil {
		x = a.left.Resolve(containerWidth)
	} else if a.right != nil {
		rightPos := a.right.Resolve(containerWidth)
		x = containerWidth - rightPos
	}

	// Calculate Y position
	if a.top != nil {
		y = a.top.Resolve(containerHeight)
	} else if a.bottom != nil {
		bottomPos := a.bottom.Resolve(containerHeight)
		y = containerHeight - bottomPos
	}

	// Adjust based on anchor
	childWidth := a.width
	if childWidth == 0 {
		childWidth = 20 // default
	}

	childHeight := a.height
	if childHeight == 0 {
		childHeight = 1 // default
	}

	switch a.anchor {
	case AnchorTop, AnchorTopLeft:
		// No adjustment needed
	case AnchorTopRight:
		x = x - childWidth
	case AnchorLeft:
		// No adjustment needed
	case AnchorCenter:
		x = x - childWidth/2
		y = y - childHeight/2
	case AnchorRight:
		x = x - childWidth
	case AnchorBottom:
		y = y - childHeight
	case AnchorBottomLeft:
		y = y - childHeight
	case AnchorBottomRight:
		x = x - childWidth
		y = y - childHeight
	}

	return x, y
}

// =============================================================================
// Layout Info
// =============================================================================

// GetLayoutInfo returns layout information for the layout engine.
func (a *VNode) GetLayoutInfo() rtui.LayoutInfo {
	return rtui.LayoutInfo{
		Flex: a.flex,
	}
}

// MeasureConstraints returns the constraints for Measure.
func (a *VNode) MeasureConstraints(c layout.Constraints) layout.Size {
	inst := a.CreateInstance()
	if measurable, ok := inst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		return measurable.Measure(c)
	}
	return layout.Size{Width: c.MinWidth, Height: c.MinHeight}
}
