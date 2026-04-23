package bulletchart

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating bullet chart VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new bullet chart builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

func (b *Builder) Label(label string) *Builder {
	b.node.SetLabel(label)
	return b
}

func (b *Builder) Value(value int) *Builder {
	b.node.SetValue(value)
	return b
}

func (b *Builder) Target(target int) *Builder {
	b.node.SetTarget(target)
	return b
}

func (b *Builder) Max(max int) *Builder {
	b.node.SetMax(max)
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) ShowTarget(show bool) *Builder {
	b.node.SetShowTarget(show)
	return b
}

func (b *Builder) ShowValueText(show bool) *Builder {
	b.node.SetShowValueText(show)
	return b
}

func (b *Builder) ValueLabelMode(mode ValueLabelMode) *Builder {
	b.node.SetValueLabelMode(mode)
	return b
}

func (b *Builder) Direction(direction Direction) *Builder {
	b.node.SetDirection(direction)
	return b
}

func (b *Builder) HigherIsBetter() *Builder {
	b.node.SetDirection(DirectionHigherBetter)
	return b
}

func (b *Builder) LowerIsBetter() *Builder {
	b.node.SetDirection(DirectionLowerBetter)
	return b
}

func (b *Builder) NeutralDirection() *Builder {
	b.node.SetDirection(DirectionNeutral)
	return b
}

func (b *Builder) InlineValueLabel() *Builder {
	b.node.SetValueLabelMode(ValueLabelModeInline)
	return b
}

func (b *Builder) BelowValueLabel() *Builder {
	b.node.SetValueLabelMode(ValueLabelModeBelow)
	return b
}

func (b *Builder) AutoValueLabel() *Builder {
	b.node.SetValueLabelMode(ValueLabelModeAuto)
	return b
}

func (b *Builder) QualitativeRanges(ranges ...QualitativeRange) *Builder {
	b.node.SetQualitativeRanges(ranges)
	return b
}

func (b *Builder) TargetMarkerRune(marker rune) *Builder {
	b.node.SetTargetMarkerRune(marker)
	return b
}

func (b *Builder) TargetMarkerStyle(s style.Style) *Builder {
	b.node.SetTargetMarkerStyle(s)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetChartStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildTyped() *VNode {
	return b.node
}
