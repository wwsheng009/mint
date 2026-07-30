package timer

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for Timer VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a Timer builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

func (b *Builder) Label(label string) *Builder {
	b.node.SetLabel(label)
	return b
}

func (b *Builder) Mode(mode Mode) *Builder {
	b.node.SetMode(mode)
	return b
}

func (b *Builder) Elapsed() *Builder {
	b.node.Elapsed()
	return b
}

func (b *Builder) Countdown(duration time.Duration) *Builder {
	b.node.Countdown(duration)
	return b
}

func (b *Builder) Until(deadline time.Time) *Builder {
	b.node.Until(deadline)
	return b
}

func (b *Builder) Deadline(deadline time.Time) *Builder {
	b.node.SetDeadline(deadline)
	b.node.SetMode(ModeCountdown)
	return b
}

func (b *Builder) Duration(duration time.Duration) *Builder {
	b.node.SetDuration(duration)
	return b
}

func (b *Builder) StartedAt(startedAt time.Time) *Builder {
	b.node.SetStartedAt(startedAt)
	return b
}

func (b *Builder) StartTime(startedAt time.Time) *Builder {
	return b.StartedAt(startedAt)
}

func (b *Builder) Now(now time.Time) *Builder {
	b.node.SetNow(now)
	return b
}

func (b *Builder) Live(live bool) *Builder {
	b.node.SetLive(live)
	return b
}

func (b *Builder) Static() *Builder {
	b.node.Static()
	return b
}

func (b *Builder) Realtime() *Builder {
	b.node.Realtime()
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) ShowProgress(show bool) *Builder {
	b.node.SetShowProgress(show)
	return b
}

func (b *Builder) Progress(show bool) *Builder {
	b.node.Progress(show)
	return b
}

func (b *Builder) ProgressWidth(width int) *Builder {
	b.node.SetProgressWidth(width)
	return b
}

func (b *Builder) ProgressGlyphStyle(glyphStyle ProgressGlyphStyle) *Builder {
	b.node.SetProgressGlyphStyle(glyphStyle)
	return b
}

func (b *Builder) UnicodeProgress() *Builder {
	b.node.UnicodeProgress()
	return b
}

func (b *Builder) ASCIIProgress() *Builder {
	b.node.ASCIIProgress()
	return b
}

func (b *Builder) ExpiredText(text string) *Builder {
	b.node.SetExpiredText(text)
	return b
}

func (b *Builder) WarningBelow(duration time.Duration) *Builder {
	b.node.SetWarningBelow(duration)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetTimerStyle(s)
	return b
}

func (b *Builder) WarningStyle(s style.Style) *Builder {
	b.node.SetWarningStyle(s)
	return b
}

func (b *Builder) ExpiredStyle(s style.Style) *Builder {
	b.node.SetExpiredStyle(s)
	return b
}

func (b *Builder) ProgressStyle(s style.Style) *Builder {
	b.node.SetProgressStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildTyped() *VNode {
	return b.node
}

func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.node.Props())
}

// Timer creates a new Timer VNode.
func Timer() *VNode {
	return New()
}

// NewTimer creates a new Timer VNode.
func NewTimer() *VNode {
	return New()
}
