package breadcrumb

import (
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

type segment struct {
	text  string
	style style.Style
	width int
}

// Instance is the runtime entity for Breadcrumb components.
type Instance struct {
	key             string
	items           []Item
	separator       string
	maxWidth        int
	breadcrumbStyle style.Style
	itemStyle       style.Style
	currentStyle    style.Style
	separatorStyle  style.Style
	bounds          [4]int
	dirty           bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new Breadcrumb instance from props.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:             proputil.GetString(props, propKey, ""),
		items:           getItemsProp(props),
		separator:       getSeparatorProp(props),
		maxWidth:        maxInt(0, proputil.GetInt(props, propMaxWidth, 0)),
		breadcrumbStyle: proputil.GetStyle(props, propStyle, style.Style{}),
		itemStyle:       proputil.GetStyle(props, propItemStyle, style.Style{}),
		currentStyle:    proputil.GetStyle(props, propCurrentStyle, style.Style{}),
		separatorStyle:  proputil.GetStyle(props, propSeparatorStyle, style.Style{}),
		dirty:           true,
	}
}

func (inst *Instance) Key() string { return inst.key }

func (inst *Instance) SetKey(key string) { inst.key = key }

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) Destroy() {}

func (inst *Instance) OnMount() {}

func (inst *Instance) OnUnmount() {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldItems := inst.items
	oldSeparator := inst.separator
	oldMaxWidth := inst.maxWidth
	oldStyle := inst.breadcrumbStyle
	oldItemStyle := inst.itemStyle
	oldCurrentStyle := inst.currentStyle
	oldSeparatorStyle := inst.separatorStyle

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.items = getItemsProp(props)
	inst.separator = getSeparatorProp(props)
	inst.maxWidth = maxInt(0, proputil.GetInt(props, propMaxWidth, inst.maxWidth))
	inst.breadcrumbStyle = proputil.GetStyle(props, propStyle, inst.breadcrumbStyle)
	inst.itemStyle = proputil.GetStyle(props, propItemStyle, inst.itemStyle)
	inst.currentStyle = proputil.GetStyle(props, propCurrentStyle, inst.currentStyle)
	inst.separatorStyle = proputil.GetStyle(props, propSeparatorStyle, inst.separatorStyle)

	changed := !itemsEqual(oldItems, inst.items) ||
		oldSeparator != inst.separator ||
		oldMaxWidth != inst.maxWidth ||
		oldStyle != inst.breadcrumbStyle ||
		oldItemStyle != inst.itemStyle ||
		oldCurrentStyle != inst.currentStyle ||
		oldSeparatorStyle != inst.separatorStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:            inst.key,
		propItems:          cloneItems(inst.items),
		propSeparator:      inst.separator,
		propMaxWidth:       inst.maxWidth,
		propStyle:          inst.breadcrumbStyle,
		propItemStyle:      inst.itemStyle,
		propCurrentStyle:   inst.currentStyle,
		propSeparatorStyle: inst.separatorStyle,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }

func (inst *Instance) IsDirty() bool { return inst.dirty }

func (inst *Instance) MarkClean() { inst.dirty = false }

func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := inst.displayWidth(inst.fullSegments())
	if inst.maxWidth > 0 && (width == 0 || inst.maxWidth < width) {
		width = inst.maxWidth
	}
	width = constraints.ConstrainWidth(width)
	return layout.Size{Width: width, Height: constraints.ConstrainHeight(1)}
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	width := inst.bounds[2]
	if inst.maxWidth > 0 && (width <= 0 || inst.maxWidth < width) {
		width = inst.maxWidth
	}
	segments := inst.visibleSegments(width)
	if len(segments) == 0 {
		return nil
	}

	cmds := make([]paint.DrawCmd, 0, len(segments))
	cursor := x
	for _, seg := range segments {
		if seg.text == "" {
			continue
		}
		cmds = append(cmds, paint.DrawCmd{
			X:     cursor,
			Y:     y,
			Text:  seg.text,
			Style: seg.style,
		})
		cursor += seg.width
	}
	return cmds
}

func (inst *Instance) fullSegments() []segment {
	if len(inst.items) == 0 {
		return nil
	}

	baseStyle := inst.breadcrumbStyle
	itemStyle := baseStyle.Merge(style.NewStyle().Foreground(theme.Link()))
	itemStyle = itemStyle.Merge(inst.itemStyle)
	currentStyle := baseStyle.Merge(style.NewStyle().Foreground(theme.Foreground()).Bold(true))
	currentStyle = currentStyle.Merge(inst.currentStyle)
	separatorStyle := baseStyle.Merge(style.NewStyle().Foreground(theme.Muted()))
	separatorStyle = separatorStyle.Merge(inst.separatorStyle)

	currentIndex := inst.currentIndex()
	separator := getSeparatorFallback(inst.separator)
	separatorWidth := paint.StringWidth(separator)

	segments := make([]segment, 0, len(inst.items)*2-1)
	for i, item := range inst.items {
		if i > 0 {
			segments = append(segments, segment{
				text:  separator,
				style: separatorStyle,
				width: separatorWidth,
			})
		}

		text := item.Label
		if item.Icon != "" {
			text = item.Icon + " " + text
		}
		segStyle := itemStyle
		if i == currentIndex {
			segStyle = currentStyle
		}
		segments = append(segments, segment{
			text:  text,
			style: segStyle,
			width: paint.StringWidth(text),
		})
	}

	return segments
}

func (inst *Instance) currentIndex() int {
	if len(inst.items) == 0 {
		return -1
	}
	currentIndex := len(inst.items) - 1
	for i, item := range inst.items {
		if item.Current {
			currentIndex = i
		}
	}
	return currentIndex
}

func (inst *Instance) visibleSegments(width int) []segment {
	full := inst.fullSegments()
	if len(full) == 0 {
		return nil
	}
	if width <= 0 {
		return full
	}
	if inst.displayWidth(full) <= width {
		return full
	}

	itemsOnly := inst.itemSegments()
	if len(itemsOnly) == 0 {
		return nil
	}

	last := itemsOnly[len(itemsOnly)-1]
	if last.width >= width {
		return []segment{{
			text:  truncateWithEllipsis(last.text, width),
			style: last.style,
			width: width,
		}}
	}

	separator := segment{
		text:  getSeparatorFallback(inst.separator),
		style: inst.resolvedSeparatorStyle(),
		width: paint.StringWidth(getSeparatorFallback(inst.separator)),
	}
	ellipsis := segment{
		text:  "…",
		style: inst.resolvedSeparatorStyle(),
		width: paint.StringWidth("…"),
	}

	if last.width+ellipsis.width+separator.width > width {
		return []segment{last}
	}

	selected := []segment{last}
	used := last.width
	start := len(itemsOnly) - 1

	for i := len(itemsOnly) - 2; i >= 0; i-- {
		extra := itemsOnly[i].width + separator.width
		prefixReserve := 0
		if i > 0 {
			prefixReserve = ellipsis.width + separator.width
		}
		if used+extra+prefixReserve > width {
			break
		}
		selected = append([]segment{itemsOnly[i]}, selected...)
		used += extra
		start = i
	}

	if start == 0 {
		return interleaveItems(selected, separator)
	}

	collapsed := []segment{ellipsis, separator}
	collapsed = append(collapsed, interleaveItems(selected, separator)...)
	return collapsed
}

func (inst *Instance) itemSegments() []segment {
	full := inst.fullSegments()
	if len(full) == 0 {
		return nil
	}
	items := make([]segment, 0, (len(full)+1)/2)
	for i, seg := range full {
		if i%2 == 0 {
			items = append(items, seg)
		}
	}
	return items
}

func (inst *Instance) resolvedSeparatorStyle() style.Style {
	baseStyle := inst.breadcrumbStyle.Merge(style.NewStyle().Foreground(theme.Muted()))
	return baseStyle.Merge(inst.separatorStyle)
}

func (inst *Instance) displayWidth(segments []segment) int {
	width := 0
	for _, seg := range segments {
		width += seg.width
	}
	return width
}

func interleaveItems(items []segment, separator segment) []segment {
	if len(items) == 0 {
		return nil
	}
	segments := make([]segment, 0, len(items)*2-1)
	for i, item := range items {
		if i > 0 {
			segments = append(segments, separator)
		}
		segments = append(segments, item)
	}
	return segments
}

func itemsEqual(a, b []Item) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func getItemsProp(props rtui.Props) []Item {
	if items, ok := props[propItems].([]Item); ok {
		return cloneItems(items)
	}
	return nil
}

func getSeparatorProp(props rtui.Props) string {
	return getSeparatorFallback(proputil.GetString(props, propSeparator, " / "))
}

func getSeparatorFallback(separator string) string {
	if strings.TrimSpace(separator) == "" {
		return " / "
	}
	return separator
}

func truncateWithEllipsis(content string, width int) string {
	const ellipsis = "…"
	if width <= 0 {
		return ""
	}
	if paint.StringWidth(content) <= width {
		return content
	}
	ellipsisWidth := paint.StringWidth(ellipsis)
	if width <= ellipsisWidth {
		return ellipsis
	}
	trimmed := strings.TrimRight(truncateByDisplayWidth(content, width-ellipsisWidth), " ")
	if trimmed == "" {
		return ellipsis
	}
	return trimmed + ellipsis
}

func truncateByDisplayWidth(content string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(content)
	result := make([]rune, 0, len(runes))
	currentWidth := 0
	for _, r := range runes {
		runeWidth := paint.RuneWidth(r)
		if currentWidth+runeWidth > width {
			break
		}
		result = append(result, r)
		currentWidth += runeWidth
	}
	return string(result)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
