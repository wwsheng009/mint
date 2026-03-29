package clock

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for Clock VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Clock builder.
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

func (b *Builder) Shape(shape DialShape) *Builder {
	b.node.SetShape(shape)
	return b
}

func (b *Builder) Circle() *Builder {
	b.node.Circle()
	return b
}

func (b *Builder) Ellipse() *Builder {
	b.node.Ellipse()
	return b
}

func (b *Builder) Radius(radius int) *Builder {
	b.node.SetRadius(radius)
	return b
}

func (b *Builder) RadiusX(radius int) *Builder {
	b.node.SetRadiusX(radius)
	return b
}

func (b *Builder) RadiusY(radius int) *Builder {
	b.node.SetRadiusY(radius)
	return b
}

func (b *Builder) Radii(radiusX, radiusY int) *Builder {
	b.node.SetRadii(radiusX, radiusY)
	return b
}

func (b *Builder) CellAspectX(cellAspectX float64) *Builder {
	b.node.SetCellAspectX(cellAspectX)
	return b
}

func (b *Builder) Live(live bool) *Builder {
	b.node.SetLive(live)
	return b
}

func (b *Builder) Realtime() *Builder {
	b.node.Realtime()
	return b
}

func (b *Builder) TimeValue(timeValue time.Time) *Builder {
	b.node.SetTimeValue(timeValue)
	return b
}

func (b *Builder) StaticTime(timeValue time.Time) *Builder {
	b.node.StaticTime(timeValue)
	return b
}

func (b *Builder) Location(location *time.Location) *Builder {
	b.node.SetLocation(location)
	return b
}

func (b *Builder) ShowSecondHand(show bool) *Builder {
	b.node.SetShowSecondHand(show)
	return b
}

func (b *Builder) HideSeconds() *Builder {
	b.node.HideSeconds()
	return b
}

func (b *Builder) SmoothSecond(smooth bool) *Builder {
	b.node.SetSmoothSecond(smooth)
	return b
}

func (b *Builder) ShowDigital(show bool) *Builder {
	b.node.SetShowDigital(show)
	return b
}

func (b *Builder) NumericTicks(show bool) *Builder {
	b.node.SetNumericTicks(show)
	return b
}

func (b *Builder) Preset(preset Preset) *Builder {
	b.node.SetPreset(preset)
	return b
}

func (b *Builder) Theme(theme Theme) *Builder {
	b.node.SetTheme(theme)
	return b
}

func (b *Builder) NoPreset() *Builder {
	b.node.NoPreset()
	return b
}

func (b *Builder) ClassicPreset() *Builder {
	b.node.ClassicPreset()
	return b
}

func (b *Builder) NeonPreset() *Builder {
	b.node.NeonPreset()
	return b
}

func (b *Builder) MinimalPreset() *Builder {
	b.node.MinimalPreset()
	return b
}

func (b *Builder) AlertPreset() *Builder {
	b.node.AlertPreset()
	return b
}

func (b *Builder) HandStyle(handStyle HandRenderStyle) *Builder {
	b.node.SetHandRenderStyle(handStyle)
	return b
}

func (b *Builder) ASCIIHands() *Builder {
	b.node.ASCIIHands()
	return b
}

func (b *Builder) UnicodeHands() *Builder {
	b.node.UnicodeHands()
	return b
}

func (b *Builder) DialStyle(s style.Style) *Builder {
	b.node.SetDialStyle(s)
	return b
}

func (b *Builder) TickStyle(s style.Style) *Builder {
	b.node.SetTickStyle(s)
	return b
}

func (b *Builder) CenterStyle(s style.Style) *Builder {
	b.node.SetCenterStyle(s)
	return b
}

func (b *Builder) DigitalStyle(s style.Style) *Builder {
	b.node.SetDigitalStyle(s)
	return b
}

func (b *Builder) HourHandStyle(s style.Style) *Builder {
	b.node.SetHourHandStyle(s)
	return b
}

func (b *Builder) MinuteHandStyle(s style.Style) *Builder {
	b.node.SetMinuteHandStyle(s)
	return b
}

func (b *Builder) SecondHandStyle(s style.Style) *Builder {
	b.node.SetSecondHandStyle(s)
	return b
}

func (b *Builder) HideDigital() *Builder {
	b.node.HideDigital()
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetClockStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildTyped() *VNode {
	return b.node
}

// Clock creates a new Clock VNode.
func Clock() *VNode {
	return New()
}

// NewClock creates a new Clock VNode.
func NewClock() *VNode {
	return New()
}
