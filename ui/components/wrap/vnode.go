// Package wrap provides Fiber-first Wrap layout component.
// Wrap is a layout container that wraps children to new rows when they exceed the container width.
// Similar to CSS flex-wrap: wrap.
package wrap

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types - Use rtui types for consistency
// =============================================================================

// Align aliases for convenience
type Align = rtui.Align

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

// VNode is the wrap layout component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Layout Props ===
	gap         int   // spacing between items in the same row
	rowGap      int   // spacing between rows (0 = use gap)
	align       Align // main-axis alignment for each row
	width       int   // container width for wrap calculation
	padding     [4]int // top, right, bottom, left
	fillWidth   bool  // stretch each row to fill container width
	fillHeight  bool  // stretch wrap to fill parent height

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
// Constructor
// =============================================================================

// New creates a new Wrap VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("wrap"),
		gap:          1,             // default gap
		rowGap:       0,             // default row gap (use gap)
		align:        AlignStart,    // default alignment
		width:        80,            // default width
		padding:      [4]int{0, 0, 0, 0},
		borderStyle:  "none",        // 默认无边框
		borderLabel:  "",
		children:     nil,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (w *VNode) Key() string {
	return w.key
}

// SetKey sets the component key - returns VNode for chaining.
func (w *VNode) SetKey(key string) rtui.VNode {
	w.key = key
	return w
}

// Tag returns the tag name.
func (w *VNode) Tag() string {
	return "wrap"
}

// Style returns the visual style.
func (w *VNode) Style() style.Style {
	return w.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (w *VNode) SetStyle(s style.Style) rtui.VNode {
	w.style = s
	return w
}

// Children returns child nodes.
func (w *VNode) Children() []rtui.VNode {
	return w.children
}

// SetChildren sets child nodes - returns VNode for chaining.
func (w *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	w.children = children
	return w
}

// GetLayer returns the rendering layer.
func (w *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (w *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return w
}

// Props returns the node properties.
func (w *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":        w.key,
		"gap":        w.gap,
		"rowGap":     w.rowGap,
		"align":      w.align,
		"width":      w.width,
		"padding":    w.padding,
		"fillWidth":  w.fillWidth,
		"fillHeight": w.fillHeight,
		"borderStyle": w.borderStyle,  // ✨ 边框样式
		"label":      w.borderLabel,  // ✨ 边框标签
		"children":   w.children,
		"style":      w.style,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (w *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		w.key = v
	}
	if v, ok := p["gap"].(int); ok {
		w.gap = v
	}
	if v, ok := p["rowGap"].(int); ok {
		w.rowGap = v
	}
	if v, ok := p["align"].(Align); ok {
		w.align = v
	}
	if v, ok := p["width"].(int); ok {
		w.width = v
	}
	if v, ok := p["padding"].([4]int); ok {
		w.padding = v
	}
	if v, ok := p["fillWidth"].(bool); ok {
		w.fillWidth = v
	}
	if v, ok := p["fillHeight"].(bool); ok {
		w.fillHeight = v
	}
	// ✨ 边框属性
	if v, ok := p["borderStyle"].(string); ok {
		w.borderStyle = v
	}
	if v, ok := p["label"].(string); ok {
		w.borderLabel = v
	}
	if v, ok := p["children"].([]rtui.VNode); ok {
		w.children = v
	}
	if v, ok := p["style"].(style.Style); ok {
		w.style = v
	}
	return w
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new WrapInstance from this VNode description.
func (w *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		"key":        w.key,
		"gap":        w.gap,
		"rowGap":     w.rowGap,
		"align":      w.align,
		"width":      w.width,
		"padding":    w.padding,
		"fillWidth":  w.fillWidth,
		"fillHeight": w.fillHeight,
		"borderStyle": w.borderStyle,  // ✨ 边框样式
		"label":      w.borderLabel,  // ✨ 边框标签
		"children":   w.children,
		"style":      w.style,
	})
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetGap sets the spacing between items in the same row.
func (w *VNode) SetGap(gap int) *VNode {
	w.gap = gap
	return w
}

// SetRowGap sets the spacing between rows (0 = use gap value).
func (w *VNode) SetRowGap(rowGap int) *VNode {
	w.rowGap = rowGap
	return w
}

// SetAlign sets the main-axis alignment for each row.
func (w *VNode) SetAlign(a Align) *VNode {
	w.align = a
	return w
}

// SetWidth sets the container width for wrap calculation.
func (w *VNode) SetWidth(width int) *VNode {
	w.width = width
	return w
}

// SetPadding sets the padding (top, right, bottom, left).
func (w *VNode) SetPadding(top, right, bottom, left int) *VNode {
	w.padding = [4]int{top, right, bottom, left}
	return w
}

// SetFillWidth makes each row stretch to fill the container width.
func (w *VNode) SetFillWidth(fill bool) *VNode {
	w.fillWidth = fill
	return w
}

// SetFillHeight makes the wrap container stretch to fill parent's height.
func (w *VNode) SetFillHeight(fill bool) *VNode {
	w.fillHeight = fill
	return w
}

// SetChildrenList sets the children.
func (w *VNode) SetChildrenList(children []rtui.VNode) *VNode {
	w.children = children
	return w
}

// AddChild appends a child to the children list.
func (w *VNode) AddChild(child rtui.VNode) *VNode {
	w.children = append(w.children, child)
	return w
}

// SetStyleProps sets the visual style.
func (w *VNode) SetStyleProps(s style.Style) *VNode {
	w.style = s
	return w
}

// =============================================================================
// Props Accessors (for Instance creation)
// =============================================================================

// Gap returns the spacing between items.
func (w *VNode) Gap() int {
	return w.gap
}

// RowGap returns the spacing between rows.
func (w *VNode) RowGap() int {
	return w.rowGap
}

// Align returns the main-axis alignment.
func (w *VNode) Align() Align {
	return w.align
}

// Width returns the container width.
func (w *VNode) Width() int {
	return w.width
}

// Padding returns the padding [top, right, bottom, left].
func (w *VNode) Padding() [4]int {
	return w.padding
}

// FillWidth returns whether rows should fill container width.
func (w *VNode) FillWidth() bool {
	return w.fillWidth
}

// FillHeight returns whether wrap should fill parent height.
func (w *VNode) FillHeight() bool {
	return w.fillHeight
}

// =============================================================================
// Layout Info (for flex layout)
// =============================================================================

// GetLayoutInfo returns layout information for the layout engine.
func (w *VNode) GetLayoutInfo() rtui.LayoutInfo {
	flex := 0
	if w.fillHeight {
		flex = 1
	}
	return rtui.LayoutInfo{
		Flex: flex,
	}
}

// MeasureConstraints returns the constraints for Measure.
func (w *VNode) MeasureConstraints(c layout.Constraints) layout.Size {
	inst := w.CreateInstance()
	if measurable, ok := inst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		return measurable.Measure(c)
	}
	return layout.Size{Width: c.MinWidth, Height: c.MinHeight}
}

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the Wrap VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: Border is treated as a container property (方案 A), not as a separate component.
func (w *VNode) GetBoxModel() layout.BoxModel {
	boxModel := layout.BoxModel{}

	// Padding from wrap properties
	boxModel.Padding = layout.Padding{
		Left:   w.padding[3],
		Right:  w.padding[1],
		Top:    w.padding[0],
		Bottom: w.padding[2],
	}

	// Border from wrap properties (container-level border, 方案 A)
	if w.borderStyle != "none" && w.borderStyle != "" {
		var borderStyle layout.BorderStyle
		switch w.borderStyle {
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
		if w.borderLabel != "" {
			boxModel.Border = layout.NewBorderWithLabel(borderStyle, w.borderLabel)
		} else {
			boxModel.Border = layout.NewBorder(borderStyle)
		}
	}

	// Note: Margin is not currently supported on Wrap VNode
	// If needed, it can be added as a property in the future

	return boxModel
}

// =============================================================================
// ✨ Border Builder Methods (方案 A - 边框作为容器属性)
// =============================================================================

// Border sets the border style and label.
func (w *VNode) Border(style string, label string) *VNode {
	w.borderStyle = style
	w.borderLabel = label
	return w
}

// Bordered sets border with specified style (no label).
func (w *VNode) Bordered(style string) *VNode {
	return w.Border(style, "")
}

// NoBorder removes border.
func (w *VNode) NoBorder() *VNode {
	return w.Border("none", "")
}

// SingleBorder sets single line border with optional label.
func (w *VNode) SingleBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return w.Border("single", lbl)
}

// DoubleBorder sets double line border with optional label.
func (w *VNode) DoubleBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return w.Border("double", lbl)
}

// RoundedBorder sets rounded border with optional label.
func (w *VNode) RoundedBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return w.Border("rounded", lbl)
}

// DashedBorder sets dashed border with optional label.
func (w *VNode) DashedBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return w.Border("dashed", lbl)
}

// BorderLabel sets only the border label (keeps current style).
func (w *VNode) BorderLabel(label string) *VNode {
	w.borderLabel = label
	return w
}
