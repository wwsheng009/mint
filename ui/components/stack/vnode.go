// Package stack provides Fiber-first Stack layout components (VStack/HStack).
package stack

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types - Use rtui types for consistency
// =============================================================================

// Direction aliases for convenience
type Direction = rtui.Direction

// Align aliases for convenience
type Align = rtui.Align

// Direction constants aliases
const (
	Row    = rtui.DirectionRow
	Column = rtui.DirectionColumn
)

// Align constants aliases
const (
	AlignStart       = rtui.AlignStart
	AlignCenter      = rtui.AlignCenter
	AlignEnd         = rtui.AlignEnd
	AlignSpaceBetween = rtui.AlignSpaceBetween
	AlignSpaceAround  = rtui.AlignSpaceAround
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the stack layout component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Layout Props ===
	direction    Direction
	align        Align // main axis alignment
	crossAlign   Align // cross axis alignment
	gap          int   // spacing between children
	padding      [4]int // top, right, bottom, left
	stretchCross bool  // stretch children to fill cross axis

	// === Sizing Props ===
	width  int // explicit width (0 = auto)
	height int // explicit height (0 = auto)
	flex   int // flex factor

	// === Children ===
	children []rtui.VNode

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

// New creates a new Stack VNode with the given direction.
func New(dir Direction) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("stack"),
		direction:    dir,
		align:        AlignStart,
		crossAlign:   AlignStart,
		gap:          0,
		padding:      [4]int{0, 0, 0, 0},
		children:     nil,
	}
}

// NewHStack creates a horizontal stack (HStack).
func NewHStack() *VNode {
	return New(Row)
}

// NewVStack creates a vertical stack (VStack).
func NewVStack() *VNode {
	return New(Column)
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (s *VNode) Key() string {
	return s.key
}

// SetKey sets the component key - returns VNode for chaining.
func (s *VNode) SetKey(key string) rtui.VNode {
	s.key = key
	return s
}

// Tag returns the tag name.
func (s *VNode) Tag() string {
	if s.direction == Row {
		return "hstack"
	}
	return "vstack"
}

// Style returns the visual style.
func (s *VNode) Style() style.Style {
	return s.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (s *VNode) SetStyle(st style.Style) rtui.VNode {
	s.style = st
	return s
}

// Children returns child nodes.
func (s *VNode) Children() []rtui.VNode {
	return s.children
}

// SetChildren sets child nodes - returns VNode for chaining.
func (s *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	s.children = children
	return s
}

// GetLayer returns the rendering layer.
func (s *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (s *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return s
}

// Props returns the node properties.
func (s *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":          s.key,
		"direction":    s.direction,
		"align":        s.align,
		"crossAlign":   s.crossAlign,
		"gap":          s.gap,
		"padding":      s.padding,
		"stretchCross": s.stretchCross,
		"width":        s.width,
		"height":       s.height,
		"flex":         s.flex,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (s *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		s.key = v
	}
	if v, ok := p["direction"].(Direction); ok {
		s.direction = v
	}
	if v, ok := p["align"].(Align); ok {
		s.align = v
	}
	if v, ok := p["crossAlign"].(Align); ok {
		s.crossAlign = v
	}
	if v, ok := p["gap"].(int); ok {
		s.gap = v
	}
	if v, ok := p["padding"].([4]int); ok {
		s.padding = v
	}
	if v, ok := p["stretchCross"].(bool); ok {
		s.stretchCross = v
	}
	if v, ok := p["width"].(int); ok {
		s.width = v
	}
	if v, ok := p["height"].(int); ok {
		s.height = v
	}
	if v, ok := p["flex"].(int); ok {
		s.flex = v
	}
	if v, ok := p["children"].([]rtui.VNode); ok {
		s.children = v
	}
	return s
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new StackInstance from this VNode description.
func (s *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		"key":          s.key,
		"direction":    s.direction,
		"align":        s.align,
		"crossAlign":   s.crossAlign,
		"gap":          s.gap,
		"padding":      s.padding,
		"stretchCross": s.stretchCross,
		"width":        s.width,
		"height":       s.height,
		"flex":         s.flex,
		"children":     s.children,
		"style":        s.style,
	})
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetDirection sets the layout direction.
func (s *VNode) SetDirection(dir Direction) *VNode {
	s.direction = dir
	return s
}

// SetAlign sets the main axis alignment.
func (s *VNode) SetAlign(a Align) *VNode {
	s.align = a
	return s
}

// SetCrossAlign sets the cross axis alignment.
func (s *VNode) SetCrossAlign(a Align) *VNode {
	s.crossAlign = a
	return s
}

// SetGap sets the spacing between children.
func (s *VNode) SetGap(gap int) *VNode {
	s.gap = gap
	return s
}

// SetPadding sets the padding (top, right, bottom, left).
func (s *VNode) SetPadding(top, right, bottom, left int) *VNode {
	s.padding = [4]int{top, right, bottom, left}
	return s
}

// SetStretchCross sets whether children should stretch to fill cross axis.
func (s *VNode) SetStretchCross(stretch bool) *VNode {
	s.stretchCross = stretch
	return s
}

// SetWidth sets the explicit width.
func (s *VNode) SetWidth(width int) *VNode {
	s.width = width
	return s
}

// SetHeight sets the explicit height.
func (s *VNode) SetHeight(height int) *VNode {
	s.height = height
	return s
}

// SetFlex sets the flex factor.
func (s *VNode) SetFlex(flex int) *VNode {
	s.flex = flex
	return s
}

// SetChildrenList sets the children.
func (s *VNode) SetChildrenList(children []rtui.VNode) *VNode {
	s.children = children
	return s
}

// AddChild appends a child to the children list.
func (s *VNode) AddChild(child rtui.VNode) *VNode {
	s.children = append(s.children, child)
	return s
}

// SetStyleProps sets the visual style.
func (s *VNode) SetStyleProps(st style.Style) *VNode {
	s.style = st
	return s
}

// =============================================================================
// Convenience Builder Methods
// =============================================================================

// Horizontal sets direction to Row (HStack).
func (s *VNode) Horizontal() *VNode {
	return s.SetDirection(Row)
}

// Vertical sets direction to Column (VStack).
func (s *VNode) Vertical() *VNode {
	return s.SetDirection(Column)
}

// Stretch enables cross-axis stretching.
func (s *VNode) Stretch() *VNode {
	return s.SetStretchCross(true)
}

// Center centers children on main axis.
func (s *VNode) Center() *VNode {
	return s.SetAlign(AlignCenter)
}

// CenterCross centers children on cross axis.
func (s *VNode) CenterCross() *VNode {
	return s.SetCrossAlign(AlignCenter)
}

// =============================================================================
// Props Accessors (for Instance creation)
// =============================================================================

// Direction returns the layout direction.
func (s *VNode) Direction() Direction {
	return s.direction
}

// Align returns the main axis alignment.
func (s *VNode) Align() Align {
	return s.align
}

// CrossAlign returns the cross axis alignment.
func (s *VNode) CrossAlign() Align {
	return s.crossAlign
}

// Gap returns the spacing between children.
func (s *VNode) Gap() int {
	return s.gap
}

// Padding returns the padding [top, right, bottom, left].
func (s *VNode) Padding() [4]int {
	return s.padding
}

// StretchCross returns whether children should stretch.
func (s *VNode) StretchCross() bool {
	return s.stretchCross
}

// Width returns the explicit width.
func (s *VNode) Width() int {
	return s.width
}

// Height returns the explicit height.
func (s *VNode) Height() int {
	return s.height
}

// Flex returns the flex factor.
func (s *VNode) Flex() int {
	return s.flex
}

// =============================================================================
// Layout Info (for flex layout)
// =============================================================================

// GetLayoutInfo returns layout information for the layout engine.
func (s *VNode) GetLayoutInfo() rtui.LayoutInfo {
	return rtui.LayoutInfo{
		Flex: s.flex,
	}
}

// MeasureConstraints returns the constraints for Measure.
func (s *VNode) MeasureConstraints(c layout.Constraints) layout.Size {
	inst := s.CreateInstance()
	if measurable, ok := inst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		return measurable.Measure(c)
	}
	return layout.Size{Width: c.MinWidth, Height: c.MinHeight}
}
