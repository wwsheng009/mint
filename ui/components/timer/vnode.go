// Package timer provides a Fiber-first timer and countdown display component.
package timer

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propDuration      = "duration"
	propExpiredStyle  = "expiredStyle"
	propExpiredText   = "expiredText"
	propKey           = "key"
	propLabel         = "label"
	propLive          = "live"
	propMode          = "mode"
	propNow           = "now"
	propProgressStyle = "progressStyle"
	propProgressWidth = "progressWidth"
	propShowProgress  = "showProgress"
	propStartedAt     = "startedAt"
	propStyle         = "style"
	propDeadline      = "deadline"
	propWarningBelow  = "warningBelow"
	propWarningStyle  = "warningStyle"
	propWidth         = "width"
)

// Mode controls how the timer value is interpreted.
type Mode int

const (
	// ModeElapsed renders elapsed time since StartedAt.
	ModeElapsed Mode = iota
	// ModeCountdown renders time remaining until Deadline or StartedAt+Duration.
	ModeCountdown
)

// VNode is the immutable description of a Timer component.
type VNode struct {
	*rtui.ElementVNode

	key           string
	label         string
	mode          Mode
	duration      time.Duration
	startedAt     time.Time
	deadline      time.Time
	now           time.Time
	live          bool
	width         int
	showProgress  bool
	progressWidth int
	expiredText   string
	warningBelow  time.Duration
	timerStyle    style.Style
	warningStyle  style.Style
	expiredStyle  style.Style
	progressStyle style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Timer VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:  rtui.NewElement("timer"),
		mode:          ModeElapsed,
		live:          true,
		progressWidth: 12,
		expiredText:   "00:00",
		warningBelow:  10 * time.Second,
		showProgress:  false,
		progressStyle: style.Style{},
		timerStyle:    style.Style{},
		warningStyle:  style.Style{},
		expiredStyle:  style.Style{},
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "timer" }

func (v *VNode) Style() style.Style { return v.timerStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.timerStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propDuration:      v.duration,
		propExpiredStyle:  v.expiredStyle,
		propExpiredText:   v.expiredText,
		propKey:           v.key,
		propLabel:         v.label,
		propLive:          v.live,
		propMode:          v.mode,
		propNow:           v.now,
		propProgressStyle: v.progressStyle,
		propProgressWidth: v.progressWidth,
		propShowProgress:  v.showProgress,
		propStartedAt:     v.startedAt,
		propStyle:         v.timerStyle,
		propDeadline:      v.deadline,
		propWarningBelow:  v.warningBelow,
		propWarningStyle:  v.warningStyle,
		propWidth:         v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if label, ok := props[propLabel].(string); ok {
		v.label = label
	}
	if mode, ok := props[propMode].(Mode); ok {
		v.mode = mode
	}
	if duration, ok := props[propDuration].(time.Duration); ok {
		v.duration = normalizeDuration(duration)
	}
	if startedAt, ok := props[propStartedAt].(time.Time); ok {
		v.startedAt = startedAt
	}
	if deadline, ok := props[propDeadline].(time.Time); ok {
		v.deadline = deadline
	}
	if now, ok := props[propNow].(time.Time); ok {
		v.now = now
	}
	if live, ok := props[propLive].(bool); ok {
		v.live = live
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = maxInt(0, width)
	}
	if showProgress, ok := props[propShowProgress].(bool); ok {
		v.showProgress = showProgress
	}
	if progressWidth, ok := props[propProgressWidth].(int); ok {
		v.progressWidth = maxInt(0, progressWidth)
	}
	if expiredText, ok := props[propExpiredText].(string); ok {
		v.expiredText = expiredText
	}
	if warningBelow, ok := props[propWarningBelow].(time.Duration); ok {
		v.warningBelow = normalizeDuration(warningBelow)
	}
	if timerStyle, ok := props[propStyle].(style.Style); ok {
		v.timerStyle = timerStyle
	}
	if warningStyle, ok := props[propWarningStyle].(style.Style); ok {
		v.warningStyle = warningStyle
	}
	if expiredStyle, ok := props[propExpiredStyle].(style.Style); ok {
		v.expiredStyle = expiredStyle
	}
	if progressStyle, ok := props[propProgressStyle].(style.Style); ok {
		v.progressStyle = progressStyle
	}
	return v
}

// CreateInstance implements rtui.InstanceFactory.
func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetLabel(label string) *VNode {
	v.label = label
	return v
}

func (v *VNode) SetMode(mode Mode) *VNode {
	v.mode = mode
	return v
}

func (v *VNode) SetDuration(duration time.Duration) *VNode {
	v.duration = normalizeDuration(duration)
	return v
}

func (v *VNode) SetStartedAt(startedAt time.Time) *VNode {
	v.startedAt = startedAt
	return v
}

func (v *VNode) SetDeadline(deadline time.Time) *VNode {
	v.deadline = deadline
	return v
}

func (v *VNode) SetNow(now time.Time) *VNode {
	v.now = now
	return v
}

func (v *VNode) SetLive(live bool) *VNode {
	v.live = live
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = maxInt(0, width)
	return v
}

func (v *VNode) SetShowProgress(show bool) *VNode {
	v.showProgress = show
	return v
}

func (v *VNode) SetProgressWidth(width int) *VNode {
	v.progressWidth = maxInt(0, width)
	return v
}

func (v *VNode) SetExpiredText(text string) *VNode {
	v.expiredText = text
	return v
}

func (v *VNode) SetWarningBelow(duration time.Duration) *VNode {
	v.warningBelow = normalizeDuration(duration)
	return v
}

func (v *VNode) SetTimerStyle(s style.Style) *VNode {
	v.timerStyle = s
	return v
}

func (v *VNode) SetWarningStyle(s style.Style) *VNode {
	v.warningStyle = s
	return v
}

func (v *VNode) SetExpiredStyle(s style.Style) *VNode {
	v.expiredStyle = s
	return v
}

func (v *VNode) SetProgressStyle(s style.Style) *VNode {
	v.progressStyle = s
	return v
}

func (v *VNode) Elapsed() *VNode {
	v.mode = ModeElapsed
	return v
}

func (v *VNode) Countdown(duration time.Duration) *VNode {
	v.mode = ModeCountdown
	v.duration = normalizeDuration(duration)
	return v
}

func (v *VNode) Until(deadline time.Time) *VNode {
	v.mode = ModeCountdown
	v.deadline = deadline
	return v
}

func (v *VNode) Static() *VNode {
	v.live = false
	return v
}

func (v *VNode) Realtime() *VNode {
	v.live = true
	return v
}

func (v *VNode) Progress(show bool) *VNode {
	v.showProgress = show
	return v
}

func (v *VNode) Label() string               { return v.label }
func (v *VNode) Mode() Mode                  { return v.mode }
func (v *VNode) Duration() time.Duration     { return v.duration }
func (v *VNode) StartedAt() time.Time        { return v.startedAt }
func (v *VNode) Deadline() time.Time         { return v.deadline }
func (v *VNode) Now() time.Time              { return v.now }
func (v *VNode) Live() bool                  { return v.live }
func (v *VNode) Width() int                  { return v.width }
func (v *VNode) ShowProgress() bool          { return v.showProgress }
func (v *VNode) ProgressWidth() int          { return v.progressWidth }
func (v *VNode) ExpiredText() string         { return v.expiredText }
func (v *VNode) WarningBelow() time.Duration { return v.warningBelow }
func (v *VNode) TimerStyle() style.Style     { return v.timerStyle }
func (v *VNode) WarningStyle() style.Style   { return v.warningStyle }
func (v *VNode) ExpiredStyle() style.Style   { return v.expiredStyle }
func (v *VNode) ProgressStyle() style.Style  { return v.progressStyle }
func (v *VNode) NewInstance() *Instance      { return NewInstance(v.Props()) }
