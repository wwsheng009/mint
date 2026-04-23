package statistic

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Statistic VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Statistic builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

// Key sets the diff key.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// Title sets the title text.
func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
	return b
}

// Value sets the statistic value.
func (b *Builder) Value(value interface{}) *Builder {
	b.node.SetValue(value)
	return b
}

// Prefix sets the prefix text.
func (b *Builder) Prefix(prefix string) *Builder {
	b.node.SetPrefix(prefix)
	return b
}

// Suffix sets the suffix text.
func (b *Builder) Suffix(suffix string) *Builder {
	b.node.SetSuffix(suffix)
	return b
}

// Extra sets the optional extra node.
func (b *Builder) Extra(extra rtui.VNode) *Builder {
	b.node.SetExtra(extra)
	return b
}

// Precision sets numeric precision.
func (b *Builder) Precision(precision int) *Builder {
	b.node.SetPrecision(precision)
	return b
}

// GroupSeparator sets the grouping separator.
func (b *Builder) GroupSeparator(separator string) *Builder {
	b.node.SetGroupSeparator(separator)
	return b
}

// DecimalSeparator sets the decimal separator.
func (b *Builder) DecimalSeparator(separator string) *Builder {
	b.node.SetDecimalSeparator(separator)
	return b
}

// Loading toggles loading state.
func (b *Builder) Loading(loading bool) *Builder {
	b.node.SetLoading(loading)
	return b
}

// Bordered toggles bordered style.
func (b *Builder) Bordered(bordered bool) *Builder {
	b.node.SetBordered(bordered)
	return b
}

// Trend sets the trend indicator.
func (b *Builder) Trend(trend Trend) *Builder {
	b.node.SetTrend(trend)
	return b
}

// Up sets an upward trend indicator.
func (b *Builder) Up() *Builder {
	b.node.SetTrend(TrendUp)
	return b
}

// Down sets a downward trend indicator.
func (b *Builder) Down() *Builder {
	b.node.SetTrend(TrendDown)
	return b
}

// Width sets the preferred width.
func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

// Style sets the root style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// TitleStyle sets the title style.
func (b *Builder) TitleStyle(s style.Style) *Builder {
	b.node.SetTitleStyle(s)
	return b
}

// ValueStyle sets the value style.
func (b *Builder) ValueStyle(s style.Style) *Builder {
	b.node.SetValueStyle(s)
	return b
}

// PrefixStyle sets the prefix style.
func (b *Builder) PrefixStyle(s style.Style) *Builder {
	b.node.SetPrefixStyle(s)
	return b
}

// SuffixStyle sets the suffix style.
func (b *Builder) SuffixStyle(s style.Style) *Builder {
	b.node.SetSuffixStyle(s)
	return b
}

// TrendStyle sets the trend style.
func (b *Builder) TrendStyle(s style.Style) *Builder {
	b.node.SetTrendStyle(s)
	return b
}

// Build returns the configured VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildVNode returns the concrete VNode.
func (b *Builder) BuildVNode() *VNode {
	return b.node
}
