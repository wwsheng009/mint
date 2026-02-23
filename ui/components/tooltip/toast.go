package tooltip

import (
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
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
	visible bool
	timer   *time.Timer
	bounds  [4]int // x, y, w, h
	dirty   bool

	// === Timestamps ===
	createdAt time.Time
	expireAt  time.Time
}

// Ensure ToastInstance implements required interfaces
var (
	_ rtui.ComponentInstance = (*ToastInstance)(nil)
	_ rtui.PaintableInstance = (*ToastInstance)(nil)
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
		key:           getStringProp(props, "key", ""),
		title:         getStringProp(props, "title", ""),
		message:       getStringProp(props, "message", ""),
		toastType:     getToastTypeProp(props, ToastInfo),
		toastDuration: duration,
		toastStyle:    getStyleProp(props),
		closeIntent:   props["closeIntent"],
		padding:       getPaddingProp(props),
		visible:       true,
		dirty:         true,
		createdAt:     time.Now(),
		expireAt:      time.Now().Add(duration),
	}

	// Start auto-hide timer
	inst.startTimer()

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
	if inst.timer != nil {
		inst.timer.Stop()
	}
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

	inst.title = getStringProp(props, "title", inst.title)
	inst.message = getStringProp(props, "message", inst.message)
	inst.toastType = getToastTypeProp(props, inst.toastType)
	inst.toastDuration = getDurationProp(props, inst.toastDuration)
	inst.toastStyle = getStyleProp(props)
	inst.closeIntent = props["closeIntent"]
	inst.padding = getPaddingProp(props)

	visible := getBoolProp(props, "visible", inst.visible)
	if visible != inst.visible {
		inst.visible = visible
		inst.dirty = true
		if visible {
			inst.startTimer()
		} else {
			inst.stopTimer()
		}
	}

	changed := oldTitle != inst.title ||
		oldMessage != inst.message ||
		oldType != inst.toastType

	if changed {
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
	if !inst.visible {
		inst.visible = true
		inst.dirty = true
		inst.startTimer()
	}
}

// Hide hides the toast and stops the timer.
func (inst *ToastInstance) Hide() {
	inst.visible = false
	inst.dirty = true
	inst.stopTimer()
}

// IsExpired returns true if the toast has exceeded its duration.
func (inst *ToastInstance) IsExpired() bool {
	return time.Now().After(inst.expireAt)
}

// Refresh resets the expire time for the toast.
func (inst *ToastInstance) Refresh() {
	inst.expireAt = time.Now().Add(inst.toastDuration)
	inst.stopTimer()
	inst.startTimer()
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
// Timer Management
// =============================================================================

func (inst *ToastInstance) startTimer() {
	inst.stopTimer()
	inst.timer = time.AfterFunc(inst.toastDuration, func() {
		inst.Hide()
	})
}

func (inst *ToastInstance) stopTimer() {
	if inst.timer != nil {
		inst.timer.Stop()
		inst.timer = nil
	}
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

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getPaddingProp(props rtui.Props) [4]int {
	if v, ok := props["padding"]; ok {
		if p, ok := v.([4]int); ok {
			return p
		}
	}
	return [4]int{}
}

func getToastDurationProp(props rtui.Props, def time.Duration) time.Duration {
	if v, ok := props["duration"]; ok {
		if d, ok := v.(time.Duration); ok {
			return d
		}
	}
	return def
}
