package notification

import (
	"fmt"
	"time"

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

// Instance is the runtime entity for Notification components.
type Instance struct {
	key              string
	notificationType NotificationType
	title            string
	message          string
	closable         bool
	closeIntent      interface{}
	duration         time.Duration
	placement        Placement
	notifyStyle      style.Style

	bounds    [4]int
	dirty     bool
	visible   bool
	shownAt   time.Time
}

var (
	_ rtui.ComponentInstance                               = (*Instance)(nil)
	_ rtui.PaintableInstance                              = (*Instance)(nil)
	_ interface{ Measure(layout.Constraints) layout.Size } = (*Instance)(nil)
)

// NewInstance creates a new Notification Instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:              proputil.GetString(props, propKey, ""),
		notificationType: getNotificationTypeProp(props, NotificationInfo),
		title:            proputil.GetString(props, propTitle, ""),
		message:          proputil.GetString(props, propMessage, ""),
		closable:         proputil.GetBool(props, propClosable, true),
		closeIntent:      props[propCloseIntent],
		duration:         getDurationProp(props, 0),
		placement:        getPlacementProp(props, PlacementTopRight),
		notifyStyle:      proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:            true,
	}
	return inst
}

// =============================================================================
// Visibility
// =============================================================================

// Show marks the notification as visible and records show time.
func (inst *Instance) Show() {
	inst.visible = true
	inst.shownAt = time.Now()
	inst.dirty = true
}

// Hide marks the notification as hidden.
func (inst *Instance) Hide() {
	inst.visible = false
	inst.dirty = true
}

// IsVisible returns whether the notification is currently visible.
func (inst *Instance) IsVisible() bool {
	return inst.visible
}

// IsExpired returns true if the notification has exceeded its duration.
// Always false when duration is 0 (persistent).
func (inst *Instance) IsExpired() bool {
	if inst.duration == 0 {
		return false
	}
	return time.Since(inst.shownAt) >= inst.duration
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

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	inst.notificationType = getNotificationTypeProp(props, NotificationInfo)
	inst.title = proputil.GetString(props, propTitle, "")
	inst.message = proputil.GetString(props, propMessage, "")
	inst.closable = proputil.GetBool(props, propClosable, true)
	if ci, ok := props[propCloseIntent]; ok {
		inst.closeIntent = ci
	}
	inst.duration = getDurationProp(props, 0)
	inst.placement = getPlacementProp(props, PlacementTopRight)
	inst.notifyStyle = proputil.GetStyle(props, propStyle, style.Style{})
	inst.dirty = true
	return true
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:              inst.key,
		propNotificationType: inst.notificationType,
		propTitle:            inst.title,
		propMessage:          inst.message,
		propClosable:         inst.closable,
		propCloseIntent:      inst.closeIntent,
		propDuration:         inst.duration,
		propPlacement:        inst.placement,
		propStyle:            inst.notifyStyle,
	}
}

// =============================================================================
// Measure
// =============================================================================

// Measure returns the size needed to render the notification.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := 40
	if constraints.MaxWidth > 0 && constraints.MaxWidth < width {
		width = constraints.MaxWidth
	}
	height := 1 // message line
	if inst.title != "" {
		height++ // title line
	}
	if inst.closable {
		height++ // close hint line
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

// Paint renders the notification as draw commands.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if !inst.visible {
		return nil
	}

	var cmds []paint.DrawCmd
	w := inst.bounds[2]
	if w == 0 {
		w = 40
	}

	// Header style varies by type
	headerStyle := inst.notifyStyle
	switch inst.notificationType {
	case NotificationSuccess:
		headerStyle = headerStyle.Foreground(theme.BG()).Background(style.Color("green")).Bold(true)
	case NotificationWarning:
		headerStyle = headerStyle.Foreground(theme.BG()).Background(style.Color("yellow")).Bold(true)
	case NotificationError:
		headerStyle = headerStyle.Foreground(theme.BG()).Background(style.Color("red")).Bold(true)
	default: // NotificationInfo
		headerStyle = headerStyle.Foreground(theme.BG()).Background(style.Color("blue")).Bold(true)
	}

	row := y

	// Title row (always shown for notifications)
	titleText := inst.title
	if titleText == "" {
		titleText = notificationTypeLabel(inst.notificationType)
	}
	titleLine := fmt.Sprintf(" %-*s", w-1, titleText)
	if len(titleLine) > w {
		titleLine = titleLine[:w]
	}
	cmds = append(cmds, paint.DrawCmd{X: x, Y: row, Text: titleLine, Style: headerStyle})
	row++

	// Message row
	bodyStyle := inst.notifyStyle.Foreground(theme.Foreground())
	msgText := fmt.Sprintf(" %-*s", w-1, inst.message)
	if len(msgText) > w {
		msgText = msgText[:w]
	}
	cmds = append(cmds, paint.DrawCmd{X: x, Y: row, Text: msgText, Style: bodyStyle})
	row++

	// Close hint row
	if inst.closable {
		hintStyle := inst.notifyStyle.Foreground(theme.Muted())
		cmds = append(cmds, paint.DrawCmd{X: x, Y: row, Text: " [press x to close]", Style: hintStyle})
	}

	return cmds
}

// notificationTypeLabel returns a default title label for the given type.
func notificationTypeLabel(t NotificationType) string {
	switch t {
	case NotificationSuccess:
		return "Success"
	case NotificationWarning:
		return "Warning"
	case NotificationError:
		return "Error"
	default:
		return "Info"
	}
}
