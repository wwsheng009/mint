package alert

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

// Instance is the runtime entity for Alert components.
type Instance struct {
	key         string
	alertType   AlertType
	title       string
	message     string
	closable    bool
	closeIntent interface{}
	alertStyle  style.Style
	bounds      [4]int
	dirty       bool
}

var (
	_ rtui.ComponentInstance                               = (*Instance)(nil)
	_ rtui.PaintableInstance                              = (*Instance)(nil)
	_ interface{ Measure(layout.Constraints) layout.Size } = (*Instance)(nil)
)

// NewInstance creates a new Alert Instance from props.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:         proputil.GetString(props, propKey, ""),
		alertType:   getAlertTypeProp(props, AlertInfo),
		title:       proputil.GetString(props, propTitle, ""),
		message:     proputil.GetString(props, propMessage, ""),
		closable:    proputil.GetBool(props, propClosable, false),
		closeIntent: props[propCloseIntent],
		alertStyle:  proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:       true,
	}
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string                      { return inst.key }
func (inst *Instance) SetKey(key string)                 { inst.key = key }
func (inst *Instance) IsDirty() bool                    { return inst.dirty }
func (inst *Instance) MarkClean()                        { inst.dirty = false }
func (inst *Instance) MarkDirty()                        { inst.dirty = true }
func (inst *Instance) Destroy()                          {}
func (inst *Instance) OnMount()                          {}
func (inst *Instance) OnUnmount()                        {}
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	inst.alertType = getAlertTypeProp(props, AlertInfo)
	inst.title = proputil.GetString(props, propTitle, "")
	inst.message = proputil.GetString(props, propMessage, "")
	inst.closable = proputil.GetBool(props, propClosable, false)
	if ci, ok := props[propCloseIntent]; ok {
		inst.closeIntent = ci
	}
	inst.alertStyle = proputil.GetStyle(props, propStyle, style.Style{})
	inst.dirty = true
	return true
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:         inst.key,
		propAlertType:   inst.alertType,
		propTitle:       inst.title,
		propMessage:     inst.message,
		propClosable:    inst.closable,
		propCloseIntent: inst.closeIntent,
		propStyle:       inst.alertStyle,
	}
}

// =============================================================================
// Measure
// =============================================================================

// Measure returns the size needed to render the alert.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := constraints.MaxWidth
	if width <= 0 {
		width = 40
	}
	height := 1 // message line
	if inst.title != "" {
		height++ // title line
	}
	if inst.closable {
		height++ // close hint line
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
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

// Paint renders the alert as draw commands.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd
	w := inst.bounds[2]
	if w == 0 {
		w = 40
	}

	s := inst.alertStyle
	switch inst.alertType {
	case AlertSuccess:
		s = s.Foreground(theme.BG()).Background(style.Color("green")).Bold(true)
	case AlertWarning:
		s = s.Foreground(theme.BG()).Background(style.Color("yellow")).Bold(true)
	case AlertError:
		s = s.Foreground(theme.BG()).Background(style.Color("red")).Bold(true)
	default: // AlertInfo
		s = s.Foreground(theme.BG()).Background(style.Color("blue")).Bold(true)
	}

	row := y

	// Title row
	if inst.title != "" {
		titleText := fmt.Sprintf(" %-*s", w-1, inst.title)
		if len(titleText) > w {
			titleText = titleText[:w]
		}
		cmds = append(cmds, paint.DrawCmd{X: x, Y: row, Text: titleText, Style: s})
		row++
	}

	// Message row
	bodyStyle := inst.alertStyle.Foreground(theme.Foreground())
	msgText := fmt.Sprintf(" %-*s", w-1, inst.message)
	if len(msgText) > w {
		msgText = msgText[:w]
	}
	cmds = append(cmds, paint.DrawCmd{X: x, Y: row, Text: msgText, Style: bodyStyle})
	row++

	// Close hint row
	if inst.closable {
		hintStyle := inst.alertStyle.Foreground(theme.Muted())
		cmds = append(cmds, paint.DrawCmd{X: x, Y: row, Text: " [press x to close]", Style: hintStyle})
	}

	return cmds
}
