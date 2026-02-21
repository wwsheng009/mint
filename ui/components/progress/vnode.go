package progress

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the progress bar description.
type VNode struct {
	*rtui.ElementVNode

	key         string
	label       string
	style       style.Style
	width       int
	value, max  int
	showPercent bool
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Progress VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("progress"),
		max:          100,
		width:        30,
		showPercent:  true,
	}
}

// =============================================================================
// rtui.VNode Interface
// =============================================================================

func (p *VNode) Key() string                    { return p.key }
func (p *VNode) SetKey(key string) rtui.VNode   { p.key = key; return p }
func (p *VNode) Tag() string                    { return "progress" }
func (p *VNode) Style() style.Style             { return p.style }
func (p *VNode) SetStyle(s style.Style) rtui.VNode { p.style = s; return p }
func (p *VNode) Children() []rtui.VNode         { return nil }
func (p *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return p }
func (p *VNode) GetLayer() rtui.Layer           { return rtui.LayerBase }
func (p *VNode) SetLayer(l rtui.Layer) rtui.VNode { return p }

func (p *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":         p.key,
		"label":       p.label,
		"style":       p.style,
		"width":       p.width,
		"value":       p.value,
		"max":         p.max,
		"showPercent": p.showPercent,
	}
}

func (p *VNode) SetProps(props rtui.Props) rtui.VNode {
	if v, ok := props["key"].(string); ok {
		p.key = v
	}
	if v, ok := props["label"].(string); ok {
		p.label = v
	}
	if v, ok := props["style"].(style.Style); ok {
		p.style = v
	}
	if v, ok := props["width"].(int); ok {
		p.width = v
	}
	if v, ok := props["value"].(int); ok {
		p.value = v
	}
	if v, ok := props["max"].(int); ok {
		p.max = v
	}
	if v, ok := props["showPercent"].(bool); ok {
		p.showPercent = v
	}
	return p
}

// =============================================================================
// InstanceFactory
// =============================================================================

func (p *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(p.Props())
}

// =============================================================================
// Builder Methods
// =============================================================================

func (p *VNode) SetLabel(label string) *VNode      { p.label = label; return p }
func (p *VNode) SetWidth(width int) *VNode         { p.width = width; return p }
func (p *VNode) SetValue(value int) *VNode         { p.value = value; return p }
func (p *VNode) SetMax(max int) *VNode             { p.max = max; return p }
func (p *VNode) SetShowPercent(v bool) *VNode      { p.showPercent = v; return p }

// =============================================================================
// Props Accessors
// =============================================================================

func (p *VNode) Label() string       { return p.label }
func (p *VNode) Width() int          { return p.width }
func (p *VNode) Value() int          { return p.value }
func (p *VNode) Max() int            { return p.max }
func (p *VNode) ShowPercent() bool   { return p.showPercent }
func (p *VNode) Percent() int {
	if p.max == 0 {
		return 0
	}
	return (p.value * 100) / p.max
}
