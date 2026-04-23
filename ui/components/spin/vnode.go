package spin

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

const (
	propKey      = "key"
	propSpinning = "spinning"
	propTip      = "tip"
	propSize     = "size"
	propDelay    = "delay"
	propStyle    = "style"
)

// =============================================================================
// Size
// =============================================================================

// Size defines the visual size of the spinner.
type Size int

const (
	SizeSmall   Size = iota // compact spinner
	SizeDefault             // standard spinner
	SizeLarge               // large spinner
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the immutable description of a Spin component.
type VNode struct {
	*rtui.ElementVNode

	key       string
	spinning  bool
	tip       string
	size      Size
	delay     int // milliseconds before showing spinner
	spinStyle style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Spin VNode (spinning by default).
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("spin"),
		spinning:     true,
		size:         SizeDefault,
	}
}

// =============================================================================
// rtui.VNode Interface
// =============================================================================

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "spin" }
func (v *VNode) Style() style.Style                           { return v.spinStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.spinStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:      v.key,
		propSpinning: v.spinning,
		propTip:      v.tip,
		propSize:     v.size,
		propDelay:    v.delay,
		propStyle:    v.spinStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if s, ok := props[propKey].(string); ok {
		v.key = s
	}
	if b, ok := props[propSpinning].(bool); ok {
		v.spinning = b
	}
	if s, ok := props[propTip].(string); ok {
		v.tip = s
	}
	if s, ok := props[propSize].(Size); ok {
		v.size = s
	}
	if d, ok := props[propDelay].(int); ok {
		v.delay = d
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.spinStyle = s
	}
	return v
}

// =============================================================================
// rtui.InstanceFactory Interface
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// NewInstance creates a typed *Instance from this VNode.
func (v *VNode) NewInstance() *Instance {
	return NewInstance(v.Props())
}

// =============================================================================
// Typed Getters
// =============================================================================

func (v *VNode) Spinning() bool         { return v.spinning }
func (v *VNode) Tip() string            { return v.tip }
func (v *VNode) Size() Size             { return v.size }
func (v *VNode) Delay() int             { return v.delay }
func (v *VNode) SpinStyle() style.Style { return v.spinStyle }

// =============================================================================
// Typed Setters (return *VNode for chaining)
// =============================================================================

func (v *VNode) SetSpinning(s bool) *VNode      { v.spinning = s; return v }
func (v *VNode) SetTip(tip string) *VNode        { v.tip = tip; return v }
func (v *VNode) SetSize(s Size) *VNode           { v.size = s; return v }
func (v *VNode) SetDelay(ms int) *VNode          { v.delay = ms; return v }
func (v *VNode) SetSpinStyle(s style.Style) *VNode { v.spinStyle = s; return v }

// Size helper methods
func (v *VNode) Small() *VNode   { v.size = SizeSmall; return v }
func (v *VNode) Default() *VNode { v.size = SizeDefault; return v }
func (v *VNode) Large() *VNode   { v.size = SizeLarge; return v }
