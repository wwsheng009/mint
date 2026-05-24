package progress

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propKey           = "key"
	propIndeterminate = "indeterminate"
	propLabel         = "label"
	propMax           = "max"
	propStatus        = "status"
	propShowPercent   = "showPercent"
	propStyle         = "style"
	propType          = "type"
	propValue         = "value"
	propWidth         = "width"
)

// =============================================================================
// Progress Type / Status
// =============================================================================

// Type controls the visual shape of the progress component.
type Type int

const (
	TypeLine Type = iota
	TypeCircle
	TypeDashboard
	TypeBlock
)

// Status controls the semantic state of the progress component.
type Status int

const (
	StatusNormal Status = iota
	StatusSuccess
	StatusException
	StatusActive
	StatusWarning
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the progress bar description.
type VNode struct {
	*rtui.ElementVNode

	key           string
	label         string
	style         style.Style
	width         int
	value, max    int
	indeterminate bool
	progressType  Type
	status        Status
	showPercent   bool
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Progress VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("progress"),
		progressType: TypeLine,
		status:       StatusNormal,
		max:          100,
		width:        30,
		showPercent:  true,
	}
}

// =============================================================================
// rtui.VNode Interface
// =============================================================================

func (p *VNode) Key() string                                  { return p.key }
func (p *VNode) SetKey(key string) rtui.VNode                 { p.key = key; return p }
func (p *VNode) Tag() string                                  { return "progress" }
func (p *VNode) Style() style.Style                           { return p.style }
func (p *VNode) SetStyle(s style.Style) rtui.VNode            { p.style = s; return p }
func (p *VNode) Children() []rtui.VNode                       { return nil }
func (p *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return p }
func (p *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (p *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return p }

func (p *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:           p.key,
		propIndeterminate: p.indeterminate,
		propLabel:         p.label,
		propStyle:         p.style,
		propWidth:         p.width,
		propValue:         p.value,
		propMax:           p.max,
		propType:          p.progressType,
		propStatus:        p.status,
		propShowPercent:   p.showPercent,
	}
}

func (p *VNode) SetProps(props rtui.Props) rtui.VNode {
	if v, ok := props[propKey].(string); ok {
		p.key = v
	}
	if v, ok := props[propIndeterminate].(bool); ok {
		p.indeterminate = v
	}
	if v, ok := props[propLabel].(string); ok {
		p.label = v
	}
	if v, ok := props[propStyle].(style.Style); ok {
		p.style = v
	}
	if v, ok := props[propWidth].(int); ok {
		p.width = v
	}
	if v, ok := props[propValue].(int); ok {
		p.value = v
	}
	if v, ok := props[propMax].(int); ok {
		p.max = v
	}
	if v, ok := props[propType].(Type); ok {
		p.progressType = v
	}
	if v, ok := props[propStatus].(Status); ok {
		p.status = v
	}
	if v, ok := props[propShowPercent].(bool); ok {
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

func (p *VNode) SetLabel(label string) *VNode   { p.label = label; return p }
func (p *VNode) SetWidth(width int) *VNode      { p.width = width; return p }
func (p *VNode) SetValue(value int) *VNode      { p.value = value; return p }
func (p *VNode) SetMax(max int) *VNode          { p.max = max; return p }
func (p *VNode) SetIndeterminate(v bool) *VNode { p.indeterminate = v; return p }
func (p *VNode) SetShowPercent(v bool) *VNode   { p.showPercent = v; return p }
func (p *VNode) SetType(t Type) *VNode          { p.progressType = t; return p }
func (p *VNode) SetStatus(status Status) *VNode { p.status = status; return p }

func (p *VNode) Line() *VNode {
	p.progressType = TypeLine
	return p
}

func (p *VNode) Circle() *VNode {
	p.progressType = TypeCircle
	return p
}

func (p *VNode) Dashboard() *VNode {
	p.progressType = TypeDashboard
	return p
}

func (p *VNode) Block() *VNode {
	p.progressType = TypeBlock
	return p
}

func (p *VNode) Normal() *VNode {
	p.status = StatusNormal
	return p
}

func (p *VNode) Success() *VNode {
	p.status = StatusSuccess
	return p
}

func (p *VNode) Exception() *VNode {
	p.status = StatusException
	return p
}

func (p *VNode) Active() *VNode {
	p.status = StatusActive
	return p
}

func (p *VNode) Warning() *VNode {
	p.status = StatusWarning
	return p
}

func (p *VNode) State(state string) *VNode {
	p.status = StatusForState(state)
	return p
}

func (p *VNode) Indeterminate() *VNode {
	p.indeterminate = true
	p.status = StatusActive
	return p
}

func (p *VNode) Determinate() *VNode {
	p.indeterminate = false
	return p
}

// =============================================================================
// Props Accessors
// =============================================================================

func (p *VNode) Label() string         { return p.label }
func (p *VNode) Width() int            { return p.width }
func (p *VNode) Value() int            { return p.value }
func (p *VNode) Max() int              { return p.max }
func (p *VNode) IsIndeterminate() bool { return p.indeterminate }
func (p *VNode) ProgressType() Type    { return p.progressType }
func (p *VNode) Status() Status        { return p.status }
func (p *VNode) ShowPercent() bool     { return p.showPercent }
func (p *VNode) Percent() int {
	if p.max == 0 {
		return 0
	}
	return (p.value * 100) / p.max
}
