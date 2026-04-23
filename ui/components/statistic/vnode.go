package statistic

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propBordered         = "bordered"
	propDecimalSeparator = "decimalSeparator"
	propExtra            = "extra"
	propGroupSeparator   = "groupSeparator"
	propKey              = "key"
	propLoading          = "loading"
	propPrecision        = "precision"
	propPrefix           = "prefix"
	propPrefixStyle      = "prefixStyle"
	propStatisticStyle   = "style"
	propSuffix           = "suffix"
	propSuffixStyle      = "suffixStyle"
	propTitle            = "title"
	propTitleStyle       = "titleStyle"
	propTrend            = "trend"
	propTrendStyle       = "trendStyle"
	propValue            = "value"
	propValueStyle       = "valueStyle"
	propWidth            = "width"
)

// Trend controls optional trend indicator rendering.
type Trend int

const (
	TrendNone Trend = iota
	TrendUp
	TrendDown
)

// VNode is the declarative description of a Statistic component.
type VNode struct {
	*rtui.ElementVNode

	key              string
	title            string
	value            interface{}
	prefix           string
	suffix           string
	extra            rtui.VNode
	precision        int
	groupSeparator   string
	decimalSeparator string
	loading          bool
	bordered         bool
	trend            Trend
	width            int
	rootStyle        style.Style
	titleStyle       style.Style
	valueStyle       style.Style
	prefixStyle      style.Style
	suffixStyle      style.Style
	trendStyle       style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Statistic VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:     rtui.NewElement("statistic"),
		precision:        -1,
		groupSeparator:   ",",
		decimalSeparator: ".",
		bordered:         false,
		trend:            TrendNone,
		rootStyle:        style.Style{},
		titleStyle:       style.Style{},
		valueStyle:       style.Style{},
		prefixStyle:      style.Style{},
		suffixStyle:      style.Style{},
		trendStyle:       style.Style{},
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "statistic" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propBordered:         v.bordered,
		propDecimalSeparator: v.decimalSeparator,
		propExtra:            v.extra,
		propGroupSeparator:   v.groupSeparator,
		propKey:              v.key,
		propLoading:          v.loading,
		propPrecision:        v.precision,
		propPrefix:           v.prefix,
		propPrefixStyle:      v.prefixStyle,
		propStatisticStyle:   v.rootStyle,
		propSuffix:           v.suffix,
		propSuffixStyle:      v.suffixStyle,
		propTitle:            v.title,
		propTitleStyle:       v.titleStyle,
		propTrend:            v.trend,
		propTrendStyle:       v.trendStyle,
		propValue:            v.value,
		propValueStyle:       v.valueStyle,
		propWidth:            v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if value, ok := props[propValue]; ok {
		v.value = value
	}
	if prefix, ok := props[propPrefix].(string); ok {
		v.prefix = prefix
	}
	if suffix, ok := props[propSuffix].(string); ok {
		v.suffix = suffix
	}
	if extra, ok := props[propExtra].(rtui.VNode); ok {
		v.extra = extra
	}
	if precision, ok := props[propPrecision].(int); ok {
		v.precision = precision
	}
	if groupSeparator, ok := props[propGroupSeparator].(string); ok {
		v.groupSeparator = groupSeparator
	}
	if decimalSeparator, ok := props[propDecimalSeparator].(string); ok {
		v.decimalSeparator = decimalSeparator
	}
	if loading, ok := props[propLoading].(bool); ok {
		v.loading = loading
	}
	if bordered, ok := props[propBordered].(bool); ok {
		v.bordered = bordered
	}
	if trend, ok := props[propTrend].(Trend); ok {
		v.trend = trend
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if s, ok := props[propStatisticStyle].(style.Style); ok {
		v.rootStyle = s
	}
	if s, ok := props[propTitleStyle].(style.Style); ok {
		v.titleStyle = s
	}
	if s, ok := props[propValueStyle].(style.Style); ok {
		v.valueStyle = s
	}
	if s, ok := props[propPrefixStyle].(style.Style); ok {
		v.prefixStyle = s
	}
	if s, ok := props[propSuffixStyle].(style.Style); ok {
		v.suffixStyle = s
	}
	if s, ok := props[propTrendStyle].(style.Style); ok {
		v.trendStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// SetTitle sets the title text.
func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

// SetValue sets the statistic value.
func (v *VNode) SetValue(value interface{}) *VNode {
	v.value = value
	return v
}

// SetPrefix sets the prefix text.
func (v *VNode) SetPrefix(prefix string) *VNode {
	v.prefix = prefix
	return v
}

// SetSuffix sets the suffix text.
func (v *VNode) SetSuffix(suffix string) *VNode {
	v.suffix = suffix
	return v
}

// SetExtra sets the optional extra node.
func (v *VNode) SetExtra(extra rtui.VNode) *VNode {
	v.extra = extra
	return v
}

// SetPrecision sets numeric precision. Use -1 for raw formatting.
func (v *VNode) SetPrecision(precision int) *VNode {
	v.precision = precision
	return v
}

// SetGroupSeparator sets the integer grouping separator.
func (v *VNode) SetGroupSeparator(separator string) *VNode {
	v.groupSeparator = separator
	return v
}

// SetDecimalSeparator sets the decimal separator.
func (v *VNode) SetDecimalSeparator(separator string) *VNode {
	v.decimalSeparator = separator
	return v
}

// SetLoading toggles loading state.
func (v *VNode) SetLoading(loading bool) *VNode {
	v.loading = loading
	return v
}

// SetBordered toggles bordered style.
func (v *VNode) SetBordered(bordered bool) *VNode {
	v.bordered = bordered
	return v
}

// SetTrend sets the trend indicator.
func (v *VNode) SetTrend(trend Trend) *VNode {
	v.trend = trend
	return v
}

// SetWidth sets the preferred width.
func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

// SetTitleStyle sets title style.
func (v *VNode) SetTitleStyle(s style.Style) *VNode {
	v.titleStyle = s
	return v
}

// SetValueStyle sets value style.
func (v *VNode) SetValueStyle(s style.Style) *VNode {
	v.valueStyle = s
	return v
}

// SetPrefixStyle sets prefix style.
func (v *VNode) SetPrefixStyle(s style.Style) *VNode {
	v.prefixStyle = s
	return v
}

// SetSuffixStyle sets suffix style.
func (v *VNode) SetSuffixStyle(s style.Style) *VNode {
	v.suffixStyle = s
	return v
}

// SetTrendStyle sets trend indicator style.
func (v *VNode) SetTrendStyle(s style.Style) *VNode {
	v.trendStyle = s
	return v
}
