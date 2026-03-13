package spin

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Spin components.
type Instance struct {
	key       string
	spinning  bool
	tip       string
	size      Size
	delay     int
	spinStyle style.Style
	bounds    [4]int
	dirty     bool
	frame     int
}

var (
	_ rtui.ComponentInstance                               = (*Instance)(nil)
	_ rtui.PaintableInstance                              = (*Instance)(nil)
	_ interface{ Measure(layout.Constraints) layout.Size } = (*Instance)(nil)
)

// spinner frames for each size
var (
	framesSmall   = []string{"-", "\\", "|", "/"}
	framesDefault = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	framesLarge   = []string{"◐", "◓", "◑", "◒"}
)

// NewInstance creates a new Spin Instance from props.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:       proputil.GetString(props, propKey, ""),
		spinning:  proputil.GetBool(props, propSpinning, true),
		tip:       proputil.GetString(props, propTip, ""),
		size:      getSizeProp(props, SizeDefault),
		delay:     proputil.GetInt(props, propDelay, 0),
		spinStyle: proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:     true,
	}
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string                       { return inst.key }
func (inst *Instance) SetKey(key string)                 { inst.key = key }
func (inst *Instance) IsDirty() bool                     { return inst.dirty }
func (inst *Instance) MarkClean()                        { inst.dirty = false }
func (inst *Instance) MarkDirty()                        { inst.dirty = true }
func (inst *Instance) Destroy()                          {}
func (inst *Instance) OnMount()                          {}
func (inst *Instance) OnUnmount()                        {}
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

// TickFrame advances the spinner animation by one frame and marks dirty.
func (inst *Instance) TickFrame() {
	inst.frame++
	inst.dirty = true
}

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	inst.spinning = proputil.GetBool(props, propSpinning, true)
	inst.tip = proputil.GetString(props, propTip, "")
	inst.size = getSizeProp(props, SizeDefault)
	inst.delay = proputil.GetInt(props, propDelay, 0)
	inst.spinStyle = proputil.GetStyle(props, propStyle, style.Style{})
	inst.dirty = true
	return true
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:      inst.key,
		propSpinning: inst.spinning,
		propTip:      inst.tip,
		propSize:     inst.size,
		propDelay:    inst.delay,
		propStyle:    inst.spinStyle,
	}
}

// =============================================================================
// Measure
// =============================================================================

// Measure returns the size needed to render the spinner.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if !inst.spinning {
		return layout.Size{Width: 0, Height: 0}
	}
	width := 3 // spinner glyph + space
	switch inst.size {
	case SizeSmall:
		width = 2
	case SizeLarge:
		width = 4
	}
	height := 1
	if inst.tip != "" {
		height = 2 // spinner row + tip row
		if len(inst.tip)+1 > width {
			width = len(inst.tip) + 1
		}
	}
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

// Paint renders the spinner as draw commands.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if !inst.spinning {
		return nil
	}

	var cmds []paint.DrawCmd

	// Pick spinner frame
	var frames []string
	switch inst.size {
	case SizeSmall:
		frames = framesSmall
	case SizeLarge:
		frames = framesLarge
	default:
		frames = framesDefault
	}
	glyph := frames[inst.frame%len(frames)]

	// Style: cyan foreground on background
	s := inst.spinStyle
	s = s.Foreground(theme.BG()).Background(style.Color("cyan")).Bold(true)

	line := glyph
	if inst.tip != "" {
		line = fmt.Sprintf("%s %s", glyph, inst.tip)
	}

	cmds = append(cmds, paint.DrawCmd{
		X:     x,
		Y:     y,
		Text:  line,
		Style: s,
	})

	// Advance animation frame
	inst.frame = (inst.frame + 1) % len(frames)

	return cmds
}
