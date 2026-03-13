package cursor

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"time"
)

// VNode is the declarative description of a standalone cursor.
// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propConfig = "config"
	propKey = "key"
	propVisible = "visible"
)

type VNode struct {
	*rtui.ElementVNode
	key     string
	config  Config
	visible bool
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new cursor vnode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("cursor"),
		config:       DefaultConfig(),
		visible:      true,
	}
}

func (v *VNode) Key() string                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode { v.key = key; return v }
func (v *VNode) Tag() string                  { return "cursor" }
func (v *VNode) Style() style.Style           { return v.config.Style }
func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.config.Style = s
	return v
}
func (v *VNode) Children() []rtui.VNode { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	return v
}
func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return v
}

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:     v.key,
		propConfig:  v.config,
		propVisible: v.visible,
	}
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if key, ok := p[propKey].(string); ok {
		v.key = key
	}
	if cfg, ok := p[propConfig].(Config); ok {
		v.config = NormalizeConfig(cfg)
	}
	if visible, ok := p[propVisible].(bool); ok {
		v.visible = visible
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetConfig(cfg Config) *VNode {
	v.config = NormalizeConfig(cfg)
	return v
}

func (v *VNode) SetShape(shape Shape) *VNode {
	v.config.Shape = shape
	return v
}

func (v *VNode) SetTheme(theme ThemeRole) *VNode {
	v.config.Theme = theme
	return v
}

func (v *VNode) SetGlyph(glyph string) *VNode {
	v.config.Glyph = glyph
	return v
}

func (v *VNode) SetBlink(blink bool) *VNode {
	v.config.Blink = blink
	return v
}

func (v *VNode) SetBlinkInterval(interval time.Duration) *VNode {
	v.config.BlinkInterval = interval
	return v
}

func (v *VNode) SetVisible(visible bool) *VNode {
	v.visible = visible
	return v
}

func (v *VNode) Config() Config {
	return v.config
}

func (v *VNode) Visible() bool {
	return v.visible
}

func (v *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{Border: layout.Border{Style: layout.BorderNone}}
}
