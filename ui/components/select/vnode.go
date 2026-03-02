package selectcomp

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Option Type
// =============================================================================

// Option represents a single option in a select.
type Option struct {
	Value string
	Label string
}

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the select description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	options []Option
	style   style.Style
	width   int

	// === Intent Props (no closures!) ===
	changeIntent intent.Intent

	// === State Props (declarative, actual state managed by Instance) ===
	selectedIndex int
	disabled      bool

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

// New creates a new Select VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:  rtui.NewElement("select"),
		options:       []Option{},
		selectedIndex: -1,
	}
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
	return "select"
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

// Children returns child nodes (select has no children).
func (s *VNode) Children() []rtui.VNode {
	return nil
}

// SetChildren is a no-op for select - returns VNode for chaining.
func (s *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
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
		"options":      s.options,
		"style":        s.style,
		"width":        s.width,
		"changeIntent": s.changeIntent,
		"selectedIndex": s.selectedIndex,
		"disabled":     s.disabled,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (s *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		s.key = v
	}
	if v, ok := p["options"].([]Option); ok {
		s.options = v
	}
	if v, ok := p["style"].(style.Style); ok {
		s.style = v
	}
	if v, ok := p["width"].(int); ok {
		s.width = v
	}
	if v, ok := p["changeIntent"].(intent.Intent); ok {
		s.changeIntent = v
	}
	if v, ok := p["selectedIndex"].(int); ok {
		s.selectedIndex = v
	}
	if v, ok := p["disabled"].(bool); ok {
		s.disabled = v
	}
	return s
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new SelectInstance from this VNode description.
func (s *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		"key":          s.key,
		"options":      s.options,
		"style":        s.style,
		"width":        s.width,
		"changeIntent": s.changeIntent,
		"selectedIndex": s.selectedIndex,
		"disabled":     s.disabled,
	}
	return NewInstance(props)
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetOptions sets the options list.
func (s *VNode) SetOptions(opts []Option) *VNode {
	s.options = opts
	return s
}

// AddOption adds a single option.
func (s *VNode) AddOption(value, label string) *VNode {
	s.options = append(s.options, Option{Value: value, Label: label})
	return s
}

// SetSelectedIndex sets the selected index.
func (s *VNode) SetSelectedIndex(idx int) *VNode {
	s.selectedIndex = idx
	return s
}

// SetDisabled sets the disabled state.
func (s *VNode) SetDisabled(disabled bool) *VNode {
	s.disabled = disabled
	return s
}

// SetWidth sets the explicit width.
func (s *VNode) SetWidth(width int) *VNode {
	s.width = width
	return s
}

// SetChangeIntent sets the change intent.
func (s *VNode) SetChangeIntent(changeIntent intent.Intent) *VNode {
	s.changeIntent = changeIntent
	return s
}

// SetStyleProps sets the visual style.
func (s *VNode) SetStyleProps(st style.Style) *VNode {
	s.style = st
	return s
}

// =============================================================================
// Props Accessors (for Instance creation)
// =============================================================================

// Options returns the options list.
func (s *VNode) Options() []Option {
	return s.options
}

// SelectedIndex returns the selected index.
func (s *VNode) SelectedIndex() int {
	return s.selectedIndex
}

// Disabled returns the disabled state.
func (s *VNode) Disabled() bool {
	return s.disabled
}

// Width returns the explicit width.
func (s *VNode) Width() int {
	return s.width
}

// ChangeIntent returns the change intent.
func (s *VNode) ChangeIntent() intent.Intent {
	return s.changeIntent
}

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the Select VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: Select uses BoxModelMixin for padding/margin, and has no border.
func (s *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   s.BoxModelMixin.Padding()[3],
			Right:  s.BoxModelMixin.Padding()[1],
			Top:    s.BoxModelMixin.Padding()[0],
			Bottom: s.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   s.BoxModelMixin.Margin()[3],
			Right:  s.BoxModelMixin.Margin()[1],
			Top:    s.BoxModelMixin.Margin()[0],
			Bottom: s.BoxModelMixin.Margin()[2],
		},
		// Select typically doesn't have a border
		Border: layout.Border{Style: layout.BorderNone},
	}
}
