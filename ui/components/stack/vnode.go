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

	// === Border Props (方案 A - 边框作为容器属性) ===
	borderStyle  string // "none", "single", "double", "rounded", "dashed"
	borderLabel  string // Optional label displayed on top border

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
		borderStyle:  "none", // 默认无边框
		borderLabel:  "",
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
		"borderStyle":  s.borderStyle,  // ✨ 边框样式
		"label":        s.borderLabel,  // ✨ 边框标签
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
	// ✨ 边框属性
	if v, ok := p["borderStyle"].(string); ok {
		s.borderStyle = v
	}
	if v, ok := p["label"].(string); ok {
		s.borderLabel = v
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
		"borderStyle":  s.borderStyle,  // ✨ 边框样式
		"label":        s.borderLabel,  // ✨ 边框标签
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

// ✨ Border Builder Methods (方案 A - 边框作为容器属性) =============================================================================

// Border sets the border style and label.
func (s *VNode) Border(style string, label string) *VNode {
	s.borderStyle = style
	s.borderLabel = label
	return s
}

// Bordered sets border with specified style (no label).
func (s *VNode) Bordered(style string) *VNode {
	return s.Border(style, "")
}

// NoBorder removes border.
func (s *VNode) NoBorder() *VNode {
	return s.Border("none", "")
}

// SingleBorder sets single line border with optional label.
func (s *VNode) SingleBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return s.Border("single", lbl)
}

// DoubleBorder sets double line border with optional label.
func (s *VNode) DoubleBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return s.Border("double", lbl)
}

// RoundedBorder sets rounded border with optional label.
func (s *VNode) RoundedBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return s.Border("rounded", lbl)
}

// DashedBorder sets dashed border with optional label.
func (s *VNode) DashedBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return s.Border("dashed", lbl)
}

// BorderLabel sets only the border label (keeps current style).
func (s *VNode) BorderLabel(label string) *VNode {
	s.borderLabel = label
	return s
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

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the Stack VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: Border is treated as a container property (方案 A), not as a separate component.
func (s *VNode) GetBoxModel() layout.BoxModel {
	boxModel := layout.BoxModel{}

	// Padding from stack properties
	boxModel.Padding = layout.Padding{
		Left:   s.padding[3],
		Right:  s.padding[1],
		Top:    s.padding[0],
		Bottom: s.padding[2],
	}

	// Border from stack properties (container-level border, 方案 A)
	if s.borderStyle != "none" && s.borderStyle != "" {
		var borderStyle layout.BorderStyle
		switch s.borderStyle {
		case "double":
			borderStyle = layout.BorderDouble
		case "rounded":
			borderStyle = layout.BorderRounded
		case "dashed":
			borderStyle = layout.BorderDashed
		case "single":
			borderStyle = layout.BorderSingle
		default:
			borderStyle = layout.BorderNone
		}

		// If label is set, use NewBorderWithLabel; otherwise use NewBorder
		if s.borderLabel != "" {
			boxModel.Border = layout.NewBorderWithLabel(borderStyle, s.borderLabel)
		} else {
			boxModel.Border = layout.NewBorder(borderStyle)
		}
	}

	// Note: Margin is not currently supported on Stack VNode
	// If needed, it can be added as a property in the future

	return boxModel
}
