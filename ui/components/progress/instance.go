package progress

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"fmt"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Progress components.
type Instance struct {
	key         string
	label       string
	progressStyle style.Style
	width       int
	value, max  int
	showPercent bool
	bounds      [4]int
	dirty       bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface{ Measure(layout.Constraints) layout.Size } = (*Instance)(nil)
)

// NewInstance creates a new ProgressInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:         proputil.GetString(props, "key", ""),
		label:       proputil.GetString(props, "label", ""),
		progressStyle: proputil.GetStyle(props, "style", style.Style{}),
		width:       proputil.GetInt(props, "width", 30),
		value:       proputil.GetInt(props, "value", 0),
		max:         proputil.GetInt(props, "max", 100),
		showPercent: proputil.GetBool(props, "showPercent", true),
		dirty:       true,
	}
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()          {}
func (inst *Instance) OnMount()          {}
func (inst *Instance) OnUnmount()        {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldValue := inst.value
	oldWidth := inst.width

	inst.label = proputil.GetString(props, "label", inst.label)
	inst.progressStyle = proputil.GetStyle(props, "style", style.Style{})
	inst.width = proputil.GetInt(props, "width", inst.width)
	inst.value = proputil.GetInt(props, "value", inst.value)
	inst.max = proputil.GetInt(props, "max", inst.max)
	inst.showPercent = proputil.GetBool(props, "showPercent", inst.showPercent)

	changed := oldValue != inst.value || oldWidth != inst.width
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:   inst.key,
		propValue: inst.value,
		propMax:   inst.max,
	}
}

func (inst *Instance) MarkDirty()    { inst.dirty = true }
func (inst *Instance) IsDirty() bool { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

// =============================================================================
// PaintableInstance Interface
// =============================================================================

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	s := inst.resolveStyle()
	var cmds []paint.DrawCmd

	width := inst.width
	if width < 10 {
		width = 10
	}

	percent := inst.Percent()

	// Build progress bar: [======>     ]
	barWidth := width - 2
	filledCount := (percent * barWidth) / 100

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filledCount {
			bar += "="
		} else if i == filledCount {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	cmds = append(cmds, paint.DrawCmd{X: x, Y: y, Text: bar, Style: s})

	// Draw percentage or label
	if inst.showPercent || inst.label != "" {
		var labelText string
		if inst.label != "" && inst.showPercent {
			labelText = fmt.Sprintf("%s: %d%%", inst.label, percent)
		} else if inst.label != "" {
			labelText = inst.label
		} else {
			labelText = fmt.Sprintf("%d%%", percent)
		}
		cmds = append(cmds, paint.DrawCmd{X: x, Y: y + 1, Text: labelText, Style: s})
	}

	return cmds
}

func (inst *Instance) resolveStyle() style.Style {
	s := inst.progressStyle

	if s.FG == "" {
		s = s.Foreground(theme.Primary())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}

	return s
}

// =============================================================================
// Progress-specific Methods
// =============================================================================

func (inst *Instance) SetValue(value int) {
	if value < 0 {
		value = 0
	}
	if value > inst.max {
		value = inst.max
	}
	if inst.value != value {
		inst.value = value
		inst.dirty = true
	}
}

func (inst *Instance) GetValue() int { return inst.value }
func (inst *Instance) GetMax() int   { return inst.max }

func (inst *Instance) Percent() int {
	if inst.max == 0 {
		return 0
	}
	return (inst.value * 100) / inst.max
}

// =============================================================================
// Measurable Interface
// =============================================================================

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	width := inst.width
	if width < 10 {
		width = 10
	}

	height := 1
	if inst.showPercent || inst.label != "" {
		height = 2
	}

	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

