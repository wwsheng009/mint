package cursor

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"time"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Instance is the standalone runtime cursor component.
type Instance struct {
	key     string
	config  Config
	visible bool
	model   *Model
	bounds  [4]int
	dirty   bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ rtui.TickableInstance  = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new cursor instance from props.
func NewInstance(props rtui.Props) *Instance {
	cfg := getConfigProp(props, "config", DefaultConfig())
	inst := &Instance{
		key:     proputil.GetString(props, "key", ""),
		config:  cfg,
		visible: proputil.GetBool(props, "visible", true),
		model:   NewModel(cfg),
		dirty:   true,
	}
	inst.model.SetVisible(inst.visible)
	return inst
}

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}
func (inst *Instance) Destroy()   {}
func (inst *Instance) OnMount()   {}
func (inst *Instance) OnUnmount() {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	changed := false

	if key := proputil.GetString(props, "key", inst.key); key != inst.key {
		inst.key = key
		changed = true
	}

	cfg := getConfigProp(props, "config", inst.config)
	if cfg != inst.config {
		inst.config = cfg
		if inst.model.SetConfig(cfg) {
			changed = true
		}
	}

	visible := proputil.GetBool(props, "visible", inst.visible)
	if visible != inst.visible {
		inst.visible = visible
		if inst.model.SetVisible(visible) {
			changed = true
		}
	}

	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":     inst.key,
		"config":  inst.config,
		"visible": inst.visible,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if cmd, ok := inst.model.DrawCmd(x, y, "", style.Style{}); ok {
		return []paint.DrawCmd{cmd}
	}
	return nil
}

func (inst *Instance) WantsTick() bool {
	return inst.model.WantsTick()
}

func (inst *Instance) Tick(now time.Time) bool {
	if inst.model.Tick(now) {
		inst.dirty = true
		return true
	}
	return false
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	return layout.Size{
		Width:  constraints.ConstrainWidth(1),
		Height: constraints.ConstrainHeight(1),
	}
}

func (inst *Instance) ClearDirty() {
	inst.dirty = false
}

func getConfigProp(props rtui.Props, key string, def Config) Config {
	if value, ok := props[key]; ok {
		if cfg, ok := value.(Config); ok {
			return NormalizeConfig(cfg)
		}
	}
	return NormalizeConfig(def)
}
