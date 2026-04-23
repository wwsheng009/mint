package empty

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// defaultImage is the ASCII art shown when no custom image is provided.
const defaultImage = "  ( ∅ )"

// Instance is the runtime entity for Empty components.
type Instance struct {
	key         string
	description string
	image       string
	emptyStyle  style.Style
	bounds      [4]int
	dirty       bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new Empty Instance from props.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:         proputil.GetString(props, propKey, ""),
		description: proputil.GetString(props, propDescription, "No Data"),
		image:       proputil.GetString(props, propImage, ""),
		emptyStyle:  proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:       true,
	}
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string                        { return inst.key }
func (inst *Instance) SetKey(key string)                  { inst.key = key }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) MarkClean()                         { inst.dirty = false }
func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) Destroy()                           {}
func (inst *Instance) OnMount()                           {}
func (inst *Instance) OnUnmount()                         {}
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.description = proputil.GetString(props, propDescription, "No Data")
	inst.image = proputil.GetString(props, propImage, "")
	inst.emptyStyle = proputil.GetStyle(props, propStyle, style.Style{})
	inst.dirty = true
	return true
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:         inst.key,
		propDescription: inst.description,
		propImage:       inst.image,
		propStyle:       inst.emptyStyle,
	}
}

// =============================================================================
// Measure
// =============================================================================

// Measure returns the size needed to render the empty state.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	img := inst.image
	if img == "" {
		img = defaultImage
	}
	width := len(img)
	if len(inst.description) > width {
		width = len(inst.description)
	}
	height := 2 // image row + description row
	if constraints.MaxWidth > 0 && width > constraints.MaxWidth {
		width = constraints.MaxWidth
	}
	return layout.Size{Width: width, Height: height}
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// SetBounds sets the render bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// Paint renders the empty state as draw commands.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd

	s := inst.emptyStyle
	s = s.Foreground(theme.Muted()).Bold(false)

	img := inst.image
	if img == "" {
		img = defaultImage
	}

	// Center image
	w := inst.bounds[2]
	imgX := x
	if w > len(img) {
		imgX = x + (w-len(img))/2
	}
	cmds = append(cmds, paint.DrawCmd{
		X:     imgX,
		Y:     y,
		Text:  img,
		Style: s,
	})

	// Center description
	desc := inst.description
	descX := x
	if w > len(desc) {
		descX = x + (w-len(desc))/2
	}
	cmds = append(cmds, paint.DrawCmd{
		X:     descX,
		Y:     y + 1,
		Text:  fmt.Sprintf("%s", desc),
		Style: s,
	})

	return cmds
}
