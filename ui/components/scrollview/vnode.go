package scrollview

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the scrollview description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Content ===
	child rtui.VNode // Content to scroll

	// === Visual Props ===
	style style.Style

	// === Layout Props ===
	width  int // viewport width (0 = auto)
	height int // viewport height (0 = auto, shows all content)

	// === Scroll Props ===
	scrollOffset int // current scroll position
	showBorder   bool
	showIndicator bool // show scroll position indicator

	// === Box Model (via interface) ===
	rtui.BoxModelMixin
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new ScrollView VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:  rtui.NewElement("scrollview"),
		showIndicator: true,
	}
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
	return "scrollview"
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
// ScrollView handles its own content painting, so returns nil here.
// The child is stored internally for content extraction during Measure/Paint.
func (v *VNode) Children() []rtui.VNode {
	// Return nil - ScrollView paints its own content
	// The child VNode is only used for content extraction, not direct painting
	return nil
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
		"key":           v.key,
		"style":         v.style,
		"width":         v.width,
		"height":        v.height,
		"scrollOffset":  v.scrollOffset,
		"showBorder":    v.showBorder,
		"showIndicator": v.showIndicator,
		"child":         v.child,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p["key"].(string); ok {
		v.key = val
	}
	if val, ok := p["style"].(style.Style); ok {
		v.style = val
	}
	if val, ok := p["width"].(int); ok {
		v.width = val
	}
	if val, ok := p["height"].(int); ok {
		v.height = val
	}
	if val, ok := p["scrollOffset"].(int); ok {
		v.scrollOffset = val
	}
	if val, ok := p["showBorder"].(bool); ok {
		v.showBorder = val
	}
	if val, ok := p["showIndicator"].(bool); ok {
		v.showIndicator = val
	}
	if val, ok := p["child"].(rtui.VNode); ok {
		v.child = val
	}
	return v
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new ScrollView Instance from this VNode description.
func (v *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		"style":         v.style,
		"width":         v.width,
		"height":        v.height,
		"scrollOffset":  v.scrollOffset,
		"showBorder":    v.showBorder,
		"showIndicator": v.showIndicator,
		"child":         v.child,
	}
	return NewInstance(props)
}

// =============================================================================
// BoxModel Implementation
// =============================================================================

// GetBorder returns the border configuration.
func (v *VNode) GetBorder() layout.Border {
	if v.showBorder {
		return layout.Border{
			Style: layout.BorderSingle,
			Width: 1,
		}
	}
	return layout.Border{}
}

// GetMargin returns margin (scrollview has no margin by default).
func (v *VNode) GetMargin() layout.Margin {
	return layout.Margin{}
}

// GetPadding returns padding (scrollview has no padding by default).
func (v *VNode) GetPadding() layout.Padding {
	return layout.Padding{}
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetChild sets the child content.
func (v *VNode) SetChild(child rtui.VNode) *VNode {
	v.child = child
	return v
}

// SetWidth sets the viewport width.
func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

// SetHeight sets the viewport height.
func (v *VNode) SetHeight(height int) *VNode {
	v.height = height
	return v
}

// SetScrollOffset sets the scroll position.
func (v *VNode) SetScrollOffset(offset int) *VNode {
	v.scrollOffset = offset
	return v
}

// SetShowBorder sets whether to show border.
func (v *VNode) SetShowBorder(show bool) *VNode {
	v.showBorder = show
	return v
}

// SetShowIndicator sets whether to show scroll indicator.
func (v *VNode) SetShowIndicator(show bool) *VNode {
	v.showIndicator = show
	return v
}

// SetStyleProps sets the visual style.
func (v *VNode) SetStyleProps(s style.Style) *VNode {
	v.style = s
	return v
}

// =============================================================================
// Props Accessors
// =============================================================================

// Child returns the child content.
func (v *VNode) Child() rtui.VNode {
	return v.child
}

// Width returns the viewport width.
func (v *VNode) Width() int {
	return v.width
}

// Height returns the viewport height.
func (v *VNode) Height() int {
	return v.height
}

// ScrollOffset returns the current scroll position.
func (v *VNode) ScrollOffset() int {
	return v.scrollOffset
}

// ShowBorder returns whether border is shown.
func (v *VNode) ShowBorder() bool {
	return v.showBorder
}

// ShowIndicator returns whether scroll indicator is shown.
func (v *VNode) ShowIndicator() bool {
	return v.showIndicator
}

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the ScrollView VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: ScrollView uses BoxModelMixin for padding/margin, and optional border.
func (v *VNode) GetBoxModel() layout.BoxModel {
	var border layout.Border
	if v.showBorder {
		border = layout.NewBorder(layout.BorderSingle)
	} else {
		border = layout.Border{Style: layout.BorderNone}
	}

	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   v.BoxModelMixin.Padding()[3],
			Right:  v.BoxModelMixin.Padding()[1],
			Top:    v.BoxModelMixin.Padding()[0],
			Bottom: v.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   v.BoxModelMixin.Margin()[3],
			Right:  v.BoxModelMixin.Margin()[1],
			Top:    v.BoxModelMixin.Margin()[0],
			Bottom: v.BoxModelMixin.Margin()[2],
		},
		Border: border,
	}
}
