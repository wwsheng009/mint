package toast

import (
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/animation"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

// =============================================================================
// Toast Instance - Runtime Entity
// =============================================================================

// ToastInstance is the runtime entity for Toast notification components.
type ToastInstance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	title         string
	message       string
	toastType     ToastType
	toastDuration time.Duration
	toastStyle    style.Style
	closeIntent   interface{}
	padding       [4]int

	// === Runtime State ===
	visible     bool
	expired     bool
	autoDismiss *animation.LoopDriver
	bounds      [4]int // x, y, w, h
	dirty       bool
}

// Ensure ToastInstance implements required interfaces
var (
	_ rtui.ComponentInstance = (*ToastInstance)(nil)
	_ rtui.PaintableInstance = (*ToastInstance)(nil)
	_ rtui.TickableInstance  = (*ToastInstance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*ToastInstance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewToastInstance creates a new ToastInstance from props.
func NewToastInstance(props rtui.Props) *ToastInstance {
	duration := getToastDurationProp(props, 3000*time.Millisecond)
	inst := &ToastInstance{
		key:           proputil.GetString(props, "key", ""),
		title:         proputil.GetString(props, "title", ""),
		message:       proputil.GetString(props, "message", ""),
		toastType:     getToastTypeProp(props, ToastInfo),
		toastDuration: duration,
		toastStyle:    proputil.GetStyle(props, "style", style.Style{}),
		closeIntent:   props["closeIntent"],
		padding:       getPaddingProp(props),
		visible:       true,
		dirty:         true,
	}
	inst.showAt(time.Now())

	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

// Key implements ComponentInstance.
func (inst *ToastInstance) Key() string {
	return inst.key
}

// SetKey implements ComponentInstance.
func (inst *ToastInstance) SetKey(key string) {
	inst.key = key
}

// Init implements ComponentInstance.
func (inst *ToastInstance) Init(props rtui.Props) {
	inst.SetProps(props)
}

// Destroy implements ComponentInstance.
func (inst *ToastInstance) Destroy() {
	inst.stopAutoDismiss()
}

// OnMount implements ComponentInstance.
func (inst *ToastInstance) OnMount() {
	// Timer is started in constructor
}

// OnUnmount implements ComponentInstance.
func (inst *ToastInstance) OnUnmount() {
	inst.Destroy()
}

// SetProps implements ComponentInstance.
func (inst *ToastInstance) SetProps(props rtui.Props) bool {
	oldTitle := inst.title
	oldMessage := inst.message
	oldType := inst.toastType
	oldDuration := inst.toastDuration
	oldStyle := inst.toastStyle

	inst.title = proputil.GetString(props, "title", inst.title)
	inst.message = proputil.GetString(props, "message", inst.message)
	inst.toastType = getToastTypeProp(props, inst.toastType)
	inst.toastDuration = getToastDurationProp(props, inst.toastDuration)
	inst.toastStyle = proputil.GetStyle(props, "style", style.Style{})
	inst.closeIntent = props["closeIntent"]
	inst.padding = getPaddingProp(props)

	visible := proputil.GetBool(props, "visible", inst.visible)
	changed := oldTitle != inst.title ||
		oldMessage != inst.message ||
		oldType != inst.toastType ||
		oldDuration != inst.toastDuration ||
		oldStyle != inst.toastStyle ||
		visible != inst.visible

	if !changed {
		return false
	}

	switch {
	case visible && !inst.visible:
		inst.showAt(time.Now())
	case !visible && inst.visible:
		inst.Hide()
	case oldDuration != inst.toastDuration && inst.visible:
		inst.expired = false
		inst.restartAutoDismiss(time.Now())
	default:
		inst.dirty = true
	}

	return changed
}

// GetProps implements ComponentInstance.
func (inst *ToastInstance) GetProps() rtui.Props {
	return rtui.Props{
		"key":       inst.key,
		"title":     inst.title,
		"message":   inst.message,
		"toastType": inst.toastType,
		"visible":   inst.visible,
	}
}

// MarkDirty implements ComponentInstance.
func (inst *ToastInstance) MarkDirty() {
	inst.dirty = true
}

// IsDirty implements ComponentInstance.
func (inst *ToastInstance) IsDirty() bool {
	return inst.dirty
}

// GetContext implements ComponentInstance (no hooks for Toast).
func (inst *ToastInstance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// TickableInstance Interface
// =============================================================================

func (inst *ToastInstance) WantsTick() bool {
	return inst.visible && inst.autoDismiss != nil && inst.autoDismiss.WantsTick()
}

func (inst *ToastInstance) Tick(now time.Time) bool {
	if !inst.WantsTick() {
		return false
	}
	if !inst.autoDismiss.Tick(now) {
		return false
	}
	if !inst.autoDismiss.Done() {
		return false
	}

	inst.expired = true
	inst.visible = false
	inst.stopAutoDismiss()
	inst.dirty = true
	return true
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
func (inst *ToastInstance) Paint(x, y int) []paint.DrawCmd {
	if !inst.visible {
		return nil
	}

	// Resolve toast style based on type
	toastStyle := inst.resolveStyle()

	var cmds []paint.DrawCmd

	if inst.title != "" {
		// Two-line toast: title + message
		cmds = append(cmds, paint.DrawCmd{
			X:     x + 1,
			Y:     y,
			Text:  inst.title,
			Style: toastStyle.Bold(true),
		})
		cmds = append(cmds, paint.DrawCmd{
			X:     x + 1,
			Y:     y + 1,
			Text:  inst.message,
			Style: toastStyle,
		})
	} else {
		// Single-line toast: just message
		cmds = append(cmds, paint.DrawCmd{
			X:     x + 1,
			Y:     y,
			Text:  inst.message,
			Style: toastStyle,
		})
	}

	return cmds
}

// resolveStyle resolves the visual style based on toast type.
func (inst *ToastInstance) resolveStyle() style.Style {
	s := inst.toastStyle

	// Apply type-based styling if not explicitly set
	if s.FG == "" && s.BG == "" {
		switch inst.toastType {
		case ToastSuccess:
			s = s.Foreground(theme.BG()).Background(style.Color("green")).Bold(true)
		case ToastWarning:
			s = s.Foreground(theme.BG()).Background(style.Color("yellow")).Bold(true)
		case ToastError:
			s = s.Foreground(theme.BG()).Background(style.Color("red")).Bold(true)
		case ToastInfo:
			s = s.Foreground(theme.BG()).Background(style.Color("blue")).Bold(true)
		}
	}

	return s
}

// =============================================================================
// Public Methods
// =============================================================================

// Show displays the toast and starts auto-hide timer.
func (inst *ToastInstance) Show() {
	inst.showAt(time.Now())
}

func (inst *ToastInstance) showAt(now time.Time) {
	inst.visible = true
	inst.expired = false
	inst.restartAutoDismiss(now)
	inst.dirty = true
}

// Hide hides the toast and stops the timer.
func (inst *ToastInstance) Hide() {
	inst.visible = false
	inst.dirty = true
	inst.stopAutoDismiss()
}

// IsVisible returns whether the toast is currently visible.
func (inst *ToastInstance) IsVisible() bool {
	return inst.visible
}

// IsExpired returns true if the toast has exceeded its duration.
func (inst *ToastInstance) IsExpired() bool {
	return inst.expired
}

// Refresh resets the expire time for the toast.
func (inst *ToastInstance) Refresh() {
	inst.showAt(time.Now())
}

// Message returns the toast message.
func (inst *ToastInstance) Message() string {
	return inst.message
}

// Title returns the toast title.
func (inst *ToastInstance) Title() string {
	return inst.title
}

// ToastType returns the toast type.
func (inst *ToastInstance) ToastType() ToastType {
	return inst.toastType
}

// Duration returns the toast duration.
func (inst *ToastInstance) Duration() time.Duration {
	return inst.toastDuration
}

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout.Measurable interface.
func (inst *ToastInstance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	// Calculate width based on title and message
	width := 0
	if inst.title != "" {
		titleLen := paint.StringWidth(inst.title)
		if titleLen > width {
			width = titleLen
		}
	}
	msgLen := paint.StringWidth(inst.message)
	if msgLen > width {
		width = msgLen
	}

	// Default width if empty
	if width == 0 {
		width = 20
	}

	height := 1
	if inst.title != "" {
		height = 2
	}

	// Apply padding
	horizontalPadding := inst.padding[1] + inst.padding[3] // right + left
	verticalPadding := inst.padding[0] + inst.padding[2]   // top + bottom

	width += horizontalPadding
	height += verticalPadding

	// Apply constraints
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
}

func (inst *ToastInstance) restartAutoDismiss(now time.Time) {
	inst.stopAutoDismiss()
	if inst.toastDuration <= 0 {
		return
	}
	inst.autoDismiss = animation.NewLoopDriver(animation.LoopDriverConfig{
		Duration:  inst.toastDuration,
		Cycles:    1,
		AutoStart: true,
	})
	if !now.IsZero() {
		inst.autoDismiss.Prime(now)
	}
}

func (inst *ToastInstance) stopAutoDismiss() {
	if inst.autoDismiss == nil {
		return
	}
	inst.autoDismiss.Stop()
	inst.autoDismiss = nil
}
