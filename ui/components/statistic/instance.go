package statistic

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

// Instance is the runtime entity for Statistic components.
type Instance struct {
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
	dirty            bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a new Statistic instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:              proputil.GetString(props, propKey, ""),
		title:            proputil.GetString(props, propTitle, ""),
		value:            getValueProp(props),
		prefix:           proputil.GetString(props, propPrefix, ""),
		suffix:           proputil.GetString(props, propSuffix, ""),
		extra:            getVNodeProp(props, propExtra),
		precision:        proputil.GetInt(props, propPrecision, -1),
		groupSeparator:   proputil.GetString(props, propGroupSeparator, ","),
		decimalSeparator: proputil.GetString(props, propDecimalSeparator, "."),
		loading:          proputil.GetBool(props, propLoading, false),
		bordered:         proputil.GetBool(props, propBordered, false),
		trend:            getTrendProp(props, TrendNone),
		width:            proputil.GetInt(props, propWidth, 0),
		rootStyle:        proputil.GetStyle(props, propStatisticStyle, style.Style{}),
		titleStyle:       proputil.GetStyle(props, propTitleStyle, style.Style{}),
		valueStyle:       proputil.GetStyle(props, propValueStyle, style.Style{}),
		prefixStyle:      proputil.GetStyle(props, propPrefixStyle, style.Style{}),
		suffixStyle:      proputil.GetStyle(props, propSuffixStyle, style.Style{}),
		trendStyle:       proputil.GetStyle(props, propTrendStyle, style.Style{}),
		dirty:            true,
	}
	inst.normalize()
	return inst
}

func (inst *Instance) Key() string { return inst.key }

func (inst *Instance) SetKey(key string) { inst.key = key }

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) Destroy() {}

func (inst *Instance) OnMount() {}

func (inst *Instance) OnUnmount() {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldTitle := inst.title
	oldValue := inst.value
	oldPrefix := inst.prefix
	oldSuffix := inst.suffix
	oldExtra := inst.extra
	oldPrecision := inst.precision
	oldGroupSeparator := inst.groupSeparator
	oldDecimalSeparator := inst.decimalSeparator
	oldLoading := inst.loading
	oldBordered := inst.bordered
	oldTrend := inst.trend
	oldWidth := inst.width
	oldRootStyle := inst.rootStyle
	oldTitleStyle := inst.titleStyle
	oldValueStyle := inst.valueStyle
	oldPrefixStyle := inst.prefixStyle
	oldSuffixStyle := inst.suffixStyle
	oldTrendStyle := inst.trendStyle

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.value = getValuePropWithDefault(props, inst.value)
	inst.prefix = proputil.GetString(props, propPrefix, inst.prefix)
	inst.suffix = proputil.GetString(props, propSuffix, inst.suffix)
	inst.extra = getVNodePropWithDefault(props, propExtra, inst.extra)
	inst.precision = proputil.GetInt(props, propPrecision, inst.precision)
	inst.groupSeparator = proputil.GetString(props, propGroupSeparator, inst.groupSeparator)
	inst.decimalSeparator = proputil.GetString(props, propDecimalSeparator, inst.decimalSeparator)
	inst.loading = proputil.GetBool(props, propLoading, inst.loading)
	inst.bordered = proputil.GetBool(props, propBordered, inst.bordered)
	inst.trend = getTrendPropWithDefault(props, inst.trend)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.rootStyle = proputil.GetStyle(props, propStatisticStyle, inst.rootStyle)
	inst.titleStyle = proputil.GetStyle(props, propTitleStyle, inst.titleStyle)
	inst.valueStyle = proputil.GetStyle(props, propValueStyle, inst.valueStyle)
	inst.prefixStyle = proputil.GetStyle(props, propPrefixStyle, inst.prefixStyle)
	inst.suffixStyle = proputil.GetStyle(props, propSuffixStyle, inst.suffixStyle)
	inst.trendStyle = proputil.GetStyle(props, propTrendStyle, inst.trendStyle)
	inst.normalize()

	changed := oldTitle != inst.title ||
		!reflect.DeepEqual(oldValue, inst.value) ||
		oldPrefix != inst.prefix ||
		oldSuffix != inst.suffix ||
		!reflect.DeepEqual(oldExtra, inst.extra) ||
		oldPrecision != inst.precision ||
		oldGroupSeparator != inst.groupSeparator ||
		oldDecimalSeparator != inst.decimalSeparator ||
		oldLoading != inst.loading ||
		oldBordered != inst.bordered ||
		oldTrend != inst.trend ||
		oldWidth != inst.width ||
		oldRootStyle != inst.rootStyle ||
		oldTitleStyle != inst.titleStyle ||
		oldValueStyle != inst.valueStyle ||
		oldPrefixStyle != inst.prefixStyle ||
		oldSuffixStyle != inst.suffixStyle ||
		oldTrendStyle != inst.trendStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propBordered:         inst.bordered,
		propDecimalSeparator: inst.decimalSeparator,
		propExtra:            inst.extra,
		propGroupSeparator:   inst.groupSeparator,
		propKey:              inst.key,
		propLoading:          inst.loading,
		propPrecision:        inst.precision,
		propPrefix:           inst.prefix,
		propPrefixStyle:      inst.prefixStyle,
		propStatisticStyle:   inst.rootStyle,
		propSuffix:           inst.suffix,
		propSuffixStyle:      inst.suffixStyle,
		propTitle:            inst.title,
		propTitleStyle:       inst.titleStyle,
		propTrend:            inst.trend,
		propTrendStyle:       inst.trendStyle,
		propValue:            inst.value,
		propValueStyle:       inst.valueStyle,
		propWidth:            inst.width,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }

func (inst *Instance) IsDirty() bool { return inst.dirty }

func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	valueRow := inst.buildValueRow()
	children := make([]rtui.VNode, 0, 3)
	if strings.TrimSpace(inst.title) != "" {
		children = append(children, inst.buildTitleNode())
	}
	if valueRow != nil {
		children = append(children, valueRow)
	}
	if inst.extra != nil {
		children = append(children, inst.extra)
	}

	if len(children) == 0 {
		return nil
	}

	rootBuilder := rtui.VStackBuilder(children...).Gap(inst.rootGap()).AlignCross(rtui.AlignStart)
	if inst.width > 0 {
		rootBuilder.Width(inst.width)
	}
	if inst.bordered {
		rootBuilder.SingleBorder()
		rootBuilder.SetBorderColor(theme.Muted())
	}
	if !inst.rootStyle.IsEmpty() {
		rootBuilder.SetStyleProps(inst.rootStyle)
	}

	root := rootBuilder.Build()
	root.SetKey("statistic-root")
	return []rtui.VNode{root}
}

func (inst *Instance) buildTitleNode() rtui.VNode {
	title := textcomp.New(inst.title).
		SetStyleProps(style.NewStyle().
			Foreground(theme.Muted()).
			Merge(inst.titleStyle),
		)
	title.SetKey("statistic-title")
	return title
}

func (inst *Instance) buildValueRow() rtui.VNode {
	segments := make([]rtui.VNode, 0, 4)
	if trend := inst.buildTrendNode(); trend != nil {
		segments = append(segments, trend)
	}
	if strings.TrimSpace(inst.prefix) != "" {
		prefix := textcomp.New(inst.prefix).SetStyleProps(inst.resolvePrefixStyle())
		prefix.SetKey("statistic-prefix")
		segments = append(segments, prefix)
	}

	valueText := inst.displayValue()
	valueNode := textcomp.New(valueText).SetStyleProps(inst.resolveValueStyle())
	valueNode.SetKey("statistic-value")
	segments = append(segments, valueNode)

	if strings.TrimSpace(inst.suffix) != "" {
		suffix := textcomp.New(inst.suffix).SetStyleProps(inst.resolveSuffixStyle())
		suffix.SetKey("statistic-suffix")
		segments = append(segments, suffix)
	}

	if len(segments) == 0 {
		return nil
	}

	builder := rtui.HStackBuilder(segments...).Gap(0).AlignCross(rtui.AlignCenter)
	row := builder.Build()
	row.SetKey("statistic-value-row")
	return row
}

func (inst *Instance) buildTrendNode() rtui.VNode {
	if inst.trend == TrendNone {
		return nil
	}
	indicator := "↑"
	switch inst.trend {
	case TrendDown:
		indicator = "↓"
	}
	node := textcomp.New(indicator).SetStyleProps(inst.resolveTrendStyle())
	node.SetKey("statistic-trend")
	return node
}

func (inst *Instance) resolveValueStyle() style.Style {
	return style.NewStyle().
		Foreground(theme.Text()).
		Bold(true).
		Merge(inst.valueStyle)
}

func (inst *Instance) resolvePrefixStyle() style.Style {
	return style.NewStyle().
		Foreground(theme.Text()).
		Bold(true).
		Merge(inst.prefixStyle)
}

func (inst *Instance) resolveSuffixStyle() style.Style {
	return style.NewStyle().
		Foreground(theme.Muted()).
		Merge(inst.suffixStyle)
}

func (inst *Instance) resolveTrendStyle() style.Style {
	base := style.NewStyle()
	switch inst.trend {
	case TrendUp:
		base = base.Foreground(theme.Success()).Bold(true)
	case TrendDown:
		base = base.Foreground(theme.Error()).Bold(true)
	default:
		base = base.Foreground(theme.Muted())
	}
	return base.Merge(inst.trendStyle)
}

func (inst *Instance) displayValue() string {
	if inst.loading {
		return "..."
	}
	return formatValue(inst.value, inst.precision, inst.groupSeparator, inst.decimalSeparator)
}

func (inst *Instance) normalize() {
	if inst.precision < -1 {
		inst.precision = -1
	}
	if inst.decimalSeparator == "" {
		inst.decimalSeparator = "."
	}
}

func (inst *Instance) rootGap() int {
	if inst.extra == nil {
		return 0
	}
	return 1
}

func getValueProp(props rtui.Props) interface{} {
	if value, ok := props[propValue]; ok {
		return value
	}
	return nil
}

func getValuePropWithDefault(props rtui.Props, def interface{}) interface{} {
	if value, ok := props[propValue]; ok {
		return value
	}
	return def
}

func getVNodeProp(props rtui.Props, key string) rtui.VNode {
	if value, ok := props[key]; ok {
		if node, ok := value.(rtui.VNode); ok {
			return node
		}
	}
	return nil
}

func getVNodePropWithDefault(props rtui.Props, key string, def rtui.VNode) rtui.VNode {
	if node := getVNodeProp(props, key); node != nil {
		return node
	}
	return def
}

func getTrendProp(props rtui.Props, def Trend) Trend {
	if value, ok := props[propTrend]; ok {
		if trend, ok := value.(Trend); ok {
			return trend
		}
	}
	return def
}

func getTrendPropWithDefault(props rtui.Props, def Trend) Trend {
	return getTrendProp(props, def)
}

func formatValue(value interface{}, precision int, groupSeparator, decimalSeparator string) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case int:
		return formatNumberString(strconv.FormatInt(int64(v), 10), groupSeparator, decimalSeparator)
	case int8:
		return formatNumberString(strconv.FormatInt(int64(v), 10), groupSeparator, decimalSeparator)
	case int16:
		return formatNumberString(strconv.FormatInt(int64(v), 10), groupSeparator, decimalSeparator)
	case int32:
		return formatNumberString(strconv.FormatInt(int64(v), 10), groupSeparator, decimalSeparator)
	case int64:
		return formatNumberString(strconv.FormatInt(v, 10), groupSeparator, decimalSeparator)
	case uint:
		return formatNumberString(strconv.FormatUint(uint64(v), 10), groupSeparator, decimalSeparator)
	case uint8:
		return formatNumberString(strconv.FormatUint(uint64(v), 10), groupSeparator, decimalSeparator)
	case uint16:
		return formatNumberString(strconv.FormatUint(uint64(v), 10), groupSeparator, decimalSeparator)
	case uint32:
		return formatNumberString(strconv.FormatUint(uint64(v), 10), groupSeparator, decimalSeparator)
	case uint64:
		return formatNumberString(strconv.FormatUint(v, 10), groupSeparator, decimalSeparator)
	case float32:
		return formatFloatString(float64(v), precision, groupSeparator, decimalSeparator)
	case float64:
		return formatFloatString(v, precision, groupSeparator, decimalSeparator)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func formatFloatString(value float64, precision int, groupSeparator, decimalSeparator string) string {
	text := strconv.FormatFloat(value, 'f', precision, 64)
	return formatNumberString(text, groupSeparator, decimalSeparator)
}

func formatNumberString(text, groupSeparator, decimalSeparator string) string {
	if text == "" {
		return ""
	}

	sign := ""
	if strings.HasPrefix(text, "-") {
		sign = "-"
		text = strings.TrimPrefix(text, "-")
	}

	integerPart := text
	fractionPart := ""
	if dot := strings.IndexRune(text, '.'); dot >= 0 {
		integerPart = text[:dot]
		fractionPart = text[dot+1:]
	}

	grouped := applyGrouping(integerPart, groupSeparator)
	if fractionPart == "" {
		return sign + grouped
	}

	separator := decimalSeparator
	if separator == "" {
		separator = "."
	}
	return sign + grouped + separator + fractionPart
}

func applyGrouping(integerPart, groupSeparator string) string {
	if integerPart == "" || groupSeparator == "" {
		return integerPart
	}
	if len(integerPart) <= 3 {
		return integerPart
	}

	var parts []string
	for len(integerPart) > 3 {
		part := integerPart[len(integerPart)-3:]
		parts = append([]string{part}, parts...)
		integerPart = integerPart[:len(integerPart)-3]
	}
	if integerPart != "" {
		parts = append([]string{integerPart}, parts...)
	}
	return strings.Join(parts, groupSeparator)
}
