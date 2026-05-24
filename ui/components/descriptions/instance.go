package descriptions

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/divider"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

type packedItem struct {
	item Item
	span int
}

// Instance is the runtime entity for Descriptions components.
type Instance struct {
	key          string
	title        string
	extra        rtui.VNode
	items        []Item
	column       int
	bordered     bool
	colon        bool
	layout       Layout
	width        int
	labelWidth   int
	contentWidth int
	emptyText    string
	maskText     string
	rootStyle    style.Style
	titleStyle   style.Style
	labelStyle   style.Style
	contentStyle style.Style
	dirty        bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a new Descriptions instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          proputil.GetString(props, propKey, ""),
		title:        proputil.GetString(props, propTitle, ""),
		extra:        getVNodeProp(props, propExtra),
		items:        normalizeItems(getItemsProp(props)),
		column:       proputil.GetInt(props, propColumn, 3),
		bordered:     proputil.GetBool(props, propBordered, false),
		colon:        proputil.GetBool(props, propColon, true),
		layout:       getLayoutProp(props, LayoutHorizontal),
		width:        proputil.GetInt(props, propWidth, 0),
		labelWidth:   proputil.GetInt(props, propLabelWidth, 0),
		contentWidth: proputil.GetInt(props, propContentWidth, 0),
		emptyText:    proputil.GetString(props, propEmptyText, "-"),
		maskText:     proputil.GetString(props, propMaskText, "****"),
		rootStyle:    proputil.GetStyle(props, propStyle, style.Style{}),
		titleStyle:   proputil.GetStyle(props, propTitleStyle, style.Style{}),
		labelStyle:   proputil.GetStyle(props, propLabelStyle, style.Style{}),
		contentStyle: proputil.GetStyle(props, propContentStyle, style.Style{}),
		dirty:        true,
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
	oldExtra := inst.extra
	oldItems := cloneItems(inst.items)
	oldColumn := inst.column
	oldBordered := inst.bordered
	oldColon := inst.colon
	oldLayout := inst.layout
	oldWidth := inst.width
	oldLabelWidth := inst.labelWidth
	oldContentWidth := inst.contentWidth
	oldEmptyText := inst.emptyText
	oldMaskText := inst.maskText
	oldRootStyle := inst.rootStyle
	oldTitleStyle := inst.titleStyle
	oldLabelStyle := inst.labelStyle
	oldContentStyle := inst.contentStyle

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.extra = getVNodePropWithDefault(props, propExtra, inst.extra)
	inst.items = normalizeItems(getItemsPropWithDefault(props, inst.items))
	inst.column = proputil.GetInt(props, propColumn, inst.column)
	inst.bordered = proputil.GetBool(props, propBordered, inst.bordered)
	inst.colon = proputil.GetBool(props, propColon, inst.colon)
	inst.layout = getLayoutPropWithDefault(props, inst.layout)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.labelWidth = proputil.GetInt(props, propLabelWidth, inst.labelWidth)
	inst.contentWidth = proputil.GetInt(props, propContentWidth, inst.contentWidth)
	inst.emptyText = proputil.GetString(props, propEmptyText, inst.emptyText)
	inst.maskText = proputil.GetString(props, propMaskText, inst.maskText)
	inst.rootStyle = proputil.GetStyle(props, propStyle, inst.rootStyle)
	inst.titleStyle = proputil.GetStyle(props, propTitleStyle, inst.titleStyle)
	inst.labelStyle = proputil.GetStyle(props, propLabelStyle, inst.labelStyle)
	inst.contentStyle = proputil.GetStyle(props, propContentStyle, inst.contentStyle)
	inst.normalize()

	changed := oldTitle != inst.title ||
		oldExtra != inst.extra ||
		!reflect.DeepEqual(oldItems, inst.items) ||
		oldColumn != inst.column ||
		oldBordered != inst.bordered ||
		oldColon != inst.colon ||
		oldLayout != inst.layout ||
		oldWidth != inst.width ||
		oldLabelWidth != inst.labelWidth ||
		oldContentWidth != inst.contentWidth ||
		oldEmptyText != inst.emptyText ||
		oldMaskText != inst.maskText ||
		oldRootStyle != inst.rootStyle ||
		oldTitleStyle != inst.titleStyle ||
		oldLabelStyle != inst.labelStyle ||
		oldContentStyle != inst.contentStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propBordered:     inst.bordered,
		propColon:        inst.colon,
		propColumn:       inst.column,
		propContentStyle: inst.contentStyle,
		propContentWidth: inst.contentWidth,
		propEmptyText:    inst.emptyText,
		propExtra:        inst.extra,
		propItems:        cloneItems(inst.items),
		propKey:          inst.key,
		propLabelWidth:   inst.labelWidth,
		propLabelStyle:   inst.labelStyle,
		propLayout:       inst.layout,
		propMaskText:     inst.maskText,
		propStyle:        inst.rootStyle,
		propTitle:        inst.title,
		propTitleStyle:   inst.titleStyle,
		propWidth:        inst.width,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }

func (inst *Instance) IsDirty() bool { return inst.dirty }

func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	body := inst.buildBody()
	rootChildren := make([]rtui.VNode, 0, 2)
	if header := inst.buildHeader(); header != nil {
		rootChildren = append(rootChildren, header)
	}
	if body != nil {
		rootChildren = append(rootChildren, body)
	}
	if len(rootChildren) == 0 {
		return nil
	}

	rootBuilder := rtui.VStackBuilder(rootChildren...).
		Gap(inst.rootGap()).
		AlignCross(rtui.AlignStart)
	if inst.width > 0 {
		rootBuilder.Width(inst.width)
	}
	if !inst.rootStyle.IsEmpty() {
		rootBuilder.SetStyleProps(inst.rootStyle)
	}

	root := rootBuilder.Build()
	root.SetKey("descriptions-root")
	return []rtui.VNode{root}
}

func (inst *Instance) buildHeader() rtui.VNode {
	if strings.TrimSpace(inst.title) == "" && inst.extra == nil {
		return nil
	}

	children := make([]rtui.VNode, 0, 3)
	if strings.TrimSpace(inst.title) != "" {
		titleNode := textcomp.New(inst.title).
			Bold(true).
			SetStyleProps(style.NewStyle().
				Foreground(theme.Text()).
				Bold(true).
				Merge(inst.titleStyle),
			)
		children = append(children, titleNode)
	}
	if inst.extra != nil {
		if len(children) > 0 {
			children = append(children, rtui.Spacer().Flex(1).Build())
		}
		children = append(children, inst.extra)
	}

	header := rtui.HStackBuilder(children...).
		Gap(1).
		AlignCross(rtui.AlignCenter)
	if inst.width > 0 {
		header.Width(inst.width)
	}
	node := header.Build()
	node.SetKey("descriptions-header")
	return node
}

func (inst *Instance) buildBody() rtui.VNode {
	rows := packRows(inst.items, inst.effectiveColumns())
	if len(rows) == 0 {
		return nil
	}

	bodyChildren := make([]rtui.VNode, 0, len(rows)*2)
	for rowIndex, row := range rows {
		rowNode := inst.buildRow(rowIndex, row)
		bodyChildren = append(bodyChildren, rowNode)
		if inst.bordered && rowIndex < len(rows)-1 {
			separator := divider.NewBuilder().
				Key(fmt.Sprintf("descriptions-divider-%d", rowIndex)).
				Style(style.NewStyle().Foreground(theme.Muted())).
				Build()
			bodyChildren = append(bodyChildren, separator)
		}
	}

	body := rtui.VStackBuilder(bodyChildren...).
		Gap(0).
		AlignCross(rtui.AlignStart)
	if inst.width > 0 {
		body.Width(inst.width)
	}
	if inst.bordered {
		body.SingleBorder()
		body.SetBorderColor(theme.Muted())
	}
	node := body.Build()
	node.SetKey("descriptions-body")
	return node
}

func (inst *Instance) buildRow(rowIndex int, row []packedItem) rtui.VNode {
	columns := inst.effectiveColumns()
	used := 0
	children := make([]rtui.VNode, 0, len(row)+1)
	for cellIndex, entry := range row {
		cell := inst.buildItemCell(entry.item)
		cell.SetKey("descriptions-cell-" + entry.item.Key)
		children = append(children, rtui.Flex(cell, entry.span))
		used += entry.span
		_ = cellIndex
	}
	if remaining := columns - used; remaining > 0 {
		children = append(children, rtui.Spacer().Flex(remaining).Build())
	}

	rowBuilder := rtui.HStackBuilder(children...).
		Gap(inst.columnGap()).
		AlignCross(rtui.AlignStart)
	if inst.width > 0 {
		rowBuilder.Width(inst.width)
	}
	node := rowBuilder.Build()
	node.SetKey("descriptions-row-" + itemKey(rowIndex))
	return node
}

func (inst *Instance) buildItemCell(item Item) rtui.VNode {
	labelText := strings.TrimSpace(item.Label)
	if labelText != "" && inst.colon && inst.layout == LayoutHorizontal {
		labelText += ":"
	}

	labelNode := textcomp.New(labelText).
		SetStyleProps(style.NewStyle().
			Foreground(theme.Muted()).
			Bold(true).
			Merge(inst.labelStyle).
			Merge(item.LabelStyle),
		)
	label := rtui.VNode(labelNode)
	if width := inst.effectiveLabelWidth(item); width > 0 {
		box := rtui.Box().Width(width).Child(label)
		label = box.Build()
		label.SetKey("descriptions-label-" + item.Key)
	}

	contentNode := inst.buildContentNode(item)
	contentNode = inst.wrapContentNode(contentNode, item)
	if width := inst.effectiveContentWidth(item); width > 0 {
		box := rtui.Box().Width(width).Child(contentNode)
		contentNode = box.Build()
		contentNode.SetKey("descriptions-content-width-" + item.Key)
	}

	var cell rtui.VNode
	switch inst.layout {
	case LayoutVertical:
		builder := rtui.VStackBuilder(label, contentNode).
			Gap(0).
			AlignCross(rtui.AlignStart)
		cell = builder.Build()
	default:
		builder := rtui.HStackBuilder(label, rtui.Flex(contentNode, 1)).
			Gap(1).
			AlignCross(rtui.AlignStart)
		cell = builder.Build()
	}
	return cell
}

func (inst *Instance) buildContentNode(item Item) rtui.VNode {
	if item.Sensitive {
		return textcomp.New(inst.effectiveMaskText(item))
	}
	if item.HasValue {
		return textcomp.New(inst.formatValue(item.Value, item))
	}
	if item.Content == nil {
		return textcomp.New(inst.effectiveEmptyText(item))
	}
	return item.Content
}

func (inst *Instance) wrapContentNode(content rtui.VNode, item Item) rtui.VNode {
	merged := inst.contentStyle.Merge(item.ContentStyle)
	if merged.IsEmpty() {
		return content
	}
	wrapper := rtui.VStackBuilder(content).Gap(0).AlignCross(rtui.AlignStart)
	wrapper.SetStyleProps(merged)
	node := wrapper.Build()
	node.SetKey("descriptions-content-" + item.Key)
	return node
}

func (inst *Instance) effectiveLabelWidth(item Item) int {
	if item.LabelWidth > 0 {
		return item.LabelWidth
	}
	return inst.labelWidth
}

func (inst *Instance) effectiveContentWidth(item Item) int {
	if item.ContentWidth > 0 {
		return item.ContentWidth
	}
	return inst.contentWidth
}

func (inst *Instance) effectiveEmptyText(item Item) string {
	if item.EmptyText != "" {
		return item.EmptyText
	}
	if inst.emptyText != "" {
		return inst.emptyText
	}
	return "-"
}

func (inst *Instance) effectiveMaskText(item Item) string {
	if item.MaskText != "" {
		return item.MaskText
	}
	if inst.maskText != "" {
		return inst.maskText
	}
	return "****"
}

func (inst *Instance) formatValue(value interface{}, item Item) string {
	if value == nil {
		return inst.effectiveEmptyText(item)
	}
	text := fmt.Sprint(value)
	if strings.TrimSpace(text) == "" {
		return inst.effectiveEmptyText(item)
	}
	return text
}

func (inst *Instance) effectiveColumns() int {
	columns := inst.column
	if columns < 1 {
		columns = 1
	}
	if inst.width <= 0 {
		return columns
	}
	switch {
	case inst.width < 48:
		return 1
	case inst.width < 88 && columns > 2:
		return 2
	default:
		return columns
	}
}

func (inst *Instance) rootGap() int {
	if strings.TrimSpace(inst.title) == "" && inst.extra == nil {
		return 0
	}
	if inst.bordered {
		return 0
	}
	return 1
}

func (inst *Instance) columnGap() int {
	if inst.bordered {
		return 2
	}
	return 3
}

func (inst *Instance) normalize() {
	inst.items = normalizeItems(inst.items)
	if inst.column < 1 {
		inst.column = 1
	}
	if inst.labelWidth < 0 {
		inst.labelWidth = 0
	}
	if inst.contentWidth < 0 {
		inst.contentWidth = 0
	}
	if inst.emptyText == "" {
		inst.emptyText = "-"
	}
	if inst.maskText == "" {
		inst.maskText = "****"
	}
}

func packRows(items []Item, columns int) [][]packedItem {
	if columns < 1 {
		columns = 1
	}
	if len(items) == 0 {
		return nil
	}

	rows := make([][]packedItem, 0, len(items))
	current := make([]packedItem, 0, columns)
	used := 0
	for _, item := range items {
		span := item.Span
		if span < 1 {
			span = 1
		}
		if span > columns {
			span = columns
		}
		if used > 0 && used+span > columns {
			rows = append(rows, current)
			current = make([]packedItem, 0, columns)
			used = 0
		}
		current = append(current, packedItem{item: item, span: span})
		used += span
		if used == columns {
			rows = append(rows, current)
			current = make([]packedItem, 0, columns)
			used = 0
		}
	}
	if len(current) > 0 {
		rows = append(rows, current)
	}
	return rows
}

func getItemsProp(props rtui.Props) []Item {
	if items, ok := props[propItems].([]Item); ok {
		return cloneItems(items)
	}
	return nil
}

func getItemsPropWithDefault(props rtui.Props, def []Item) []Item {
	items := getItemsProp(props)
	if items == nil {
		return cloneItems(def)
	}
	return items
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

func getLayoutProp(props rtui.Props, def Layout) Layout {
	if value, ok := props[propLayout]; ok {
		if layout, ok := value.(Layout); ok {
			return layout
		}
	}
	return def
}

func getLayoutPropWithDefault(props rtui.Props, def Layout) Layout {
	return getLayoutProp(props, def)
}

func itemKey(index int) string {
	return fmt.Sprintf("%d", index)
}
