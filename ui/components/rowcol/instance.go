package rowcol

import (
	"reflect"
	"strconv"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const gridColumns = 24

type rowItem struct {
	node   rtui.VNode
	span   int
	offset int
}

type rowLine struct {
	items []rowItem
	used  int
}

// RowInstance is the runtime entity for Row components.
type RowInstance struct {
	key            string
	children       []rtui.VNode
	justify        rtui.Align
	align          rtui.Align
	gutter         int
	verticalGutter int
	wrap           bool
	width          int
	rootStyle      style.Style
	dirty          bool
}

// ColInstance is the runtime entity for Col components.
type ColInstance struct {
	key       string
	span      int
	offset    int
	children  []rtui.VNode
	rootStyle style.Style
	dirty     bool
}

var (
	_ rtui.ComponentInstance       = (*RowInstance)(nil)
	_ rtui.RuntimeChildrenProvider = (*RowInstance)(nil)
	_ rtui.ComponentInstance       = (*ColInstance)(nil)
	_ rtui.RuntimeChildrenProvider = (*ColInstance)(nil)
)

// NewRowInstance creates a new Row instance.
func NewRowInstance(props rtui.Props) *RowInstance {
	inst := &RowInstance{
		key:            getStringProp(props, propKey, ""),
		children:       getChildrenProp(props),
		justify:        getAlignProp(props, propJustify, rtui.AlignStart),
		align:          getAlignProp(props, propAlign, rtui.AlignStart),
		gutter:         getIntProp(props, propGutter, 0),
		verticalGutter: getIntProp(props, propVerticalGutter, 0),
		wrap:           getBoolProp(props, propWrap, true),
		width:          getIntProp(props, propWidth, 0),
		rootStyle:      getStyleProp(props, propRowStyle, style.Style{}),
		dirty:          true,
	}
	inst.normalize()
	return inst
}

// NewColInstance creates a new Col instance.
func NewColInstance(props rtui.Props) *ColInstance {
	inst := &ColInstance{
		key:       getStringProp(props, propKey, ""),
		span:      getIntProp(props, propColSpan, 0),
		offset:    getIntProp(props, propColOffset, 0),
		children:  getChildrenProp(props),
		rootStyle: getStyleProp(props, propRowStyle, style.Style{}),
		dirty:     true,
	}
	inst.normalize()
	return inst
}

func (inst *RowInstance) Key() string           { return inst.key }
func (inst *RowInstance) SetKey(key string)     { inst.key = key }
func (inst *RowInstance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *RowInstance) Destroy()              {}
func (inst *RowInstance) OnMount()              {}
func (inst *RowInstance) OnUnmount()            {}
func (inst *RowInstance) MarkDirty()            { inst.dirty = true }
func (inst *RowInstance) IsDirty() bool         { return inst.dirty }
func (inst *RowInstance) GetContext() *rtui.ComponentContext {
	return nil
}

func (inst *ColInstance) Key() string           { return inst.key }
func (inst *ColInstance) SetKey(key string)     { inst.key = key }
func (inst *ColInstance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *ColInstance) Destroy()              {}
func (inst *ColInstance) OnMount()              {}
func (inst *ColInstance) OnUnmount()            {}
func (inst *ColInstance) MarkDirty()            { inst.dirty = true }
func (inst *ColInstance) IsDirty() bool         { return inst.dirty }
func (inst *ColInstance) GetContext() *rtui.ComponentContext {
	return nil
}

func (inst *RowInstance) SetProps(props rtui.Props) bool {
	oldChildren := append([]rtui.VNode(nil), inst.children...)
	oldJustify := inst.justify
	oldAlign := inst.align
	oldGutter := inst.gutter
	oldVerticalGutter := inst.verticalGutter
	oldWrap := inst.wrap
	oldWidth := inst.width
	oldStyle := inst.rootStyle

	inst.key = getStringProp(props, propKey, inst.key)
	inst.children = getChildrenPropWithDefault(props, inst.children)
	inst.justify = getAlignProp(props, propJustify, inst.justify)
	inst.align = getAlignProp(props, propAlign, inst.align)
	inst.gutter = getIntProp(props, propGutter, inst.gutter)
	inst.verticalGutter = getIntProp(props, propVerticalGutter, inst.verticalGutter)
	inst.wrap = getBoolProp(props, propWrap, inst.wrap)
	inst.width = getIntProp(props, propWidth, inst.width)
	inst.rootStyle = getStyleProp(props, propRowStyle, inst.rootStyle)
	inst.normalize()

	changed := oldJustify != inst.justify ||
		oldAlign != inst.align ||
		oldGutter != inst.gutter ||
		oldVerticalGutter != inst.verticalGutter ||
		oldWrap != inst.wrap ||
		oldWidth != inst.width ||
		oldStyle != inst.rootStyle ||
		!reflect.DeepEqual(oldChildren, inst.children)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *RowInstance) GetProps() rtui.Props {
	return rtui.Props{
		propAlign:          inst.align,
		propChildren:       append([]rtui.VNode(nil), inst.children...),
		propGutter:         inst.gutter,
		propJustify:        inst.justify,
		propKey:            inst.key,
		propRowStyle:       inst.rootStyle,
		propVerticalGutter: inst.verticalGutter,
		propWidth:          inst.width,
		propWrap:           inst.wrap,
	}
}

func (inst *ColInstance) SetProps(props rtui.Props) bool {
	oldChildren := append([]rtui.VNode(nil), inst.children...)
	oldSpan := inst.span
	oldOffset := inst.offset
	oldStyle := inst.rootStyle

	inst.key = getStringProp(props, propKey, inst.key)
	inst.span = getIntProp(props, propColSpan, inst.span)
	inst.offset = getIntProp(props, propColOffset, inst.offset)
	inst.children = getChildrenPropWithDefault(props, inst.children)
	inst.rootStyle = getStyleProp(props, propRowStyle, inst.rootStyle)
	inst.normalize()

	changed := oldSpan != inst.span ||
		oldOffset != inst.offset ||
		oldStyle != inst.rootStyle ||
		!reflect.DeepEqual(oldChildren, inst.children)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *ColInstance) GetProps() rtui.Props {
	return rtui.Props{
		propChildren:  append([]rtui.VNode(nil), inst.children...),
		propColOffset: inst.offset,
		propColSpan:   inst.span,
		propKey:       inst.key,
		propRowStyle:  inst.rootStyle,
	}
}

func (inst *RowInstance) RuntimeChildren() []rtui.VNode {
	items := inst.runtimeItems()
	if len(items) == 0 {
		return nil
	}

	lines := inst.splitLines(items)
	if len(lines) == 1 {
		line := inst.buildLine(lines[0], 0)
		if line == nil {
			return nil
		}
		if inst.width > 0 {
			line.SetProp("width", inst.width)
		}
		if !inst.rootStyle.IsEmpty() {
			line.SetStyle(inst.rootStyle)
		}
		line.SetKey(inst.rootKey())
		return []rtui.VNode{line}
	}

	lineNodes := make([]rtui.VNode, 0, len(lines))
	for index, line := range lines {
		node := inst.buildLine(line, index)
		if node != nil {
			lineNodes = append(lineNodes, node)
		}
	}
	if len(lineNodes) == 0 {
		return nil
	}

	root := rtui.VStackBuilder(lineNodes...).Gap(inst.verticalGutter).AlignCross(rtui.AlignStart).Stretch()
	if inst.width > 0 {
		root.Width(inst.width)
	}
	if !inst.rootStyle.IsEmpty() {
		root.SetStyleProps(inst.rootStyle)
	}
	node := root.Build()
	node.SetKey(inst.rootKey())
	return []rtui.VNode{node}
}

func (inst *ColInstance) RuntimeChildren() []rtui.VNode {
	children := filterChildren(inst.children)
	if len(children) == 0 {
		return nil
	}

	root := rtui.VStackBuilder(children...).Gap(0).AlignCross(rtui.AlignStart).Stretch()
	if !inst.rootStyle.IsEmpty() {
		root.SetStyleProps(inst.rootStyle)
	}
	node := root.Build()
	node.SetKey(inst.rootKey())
	return []rtui.VNode{node}
}

func (inst *RowInstance) runtimeItems() []rowItem {
	filtered := filterChildren(inst.children)
	if len(filtered) == 0 {
		return nil
	}

	items := make([]rowItem, 0, len(filtered))
	for _, child := range filtered {
		items = append(items, inst.resolveItem(child))
	}
	return items
}

func (inst *RowInstance) resolveItem(child rtui.VNode) rowItem {
	item := rowItem{
		node: child,
		span: gridColumns,
	}

	if col, ok := child.(*ColVNode); ok {
		span := normalizeSpan(col.span)
		offset := normalizeOffset(col.offset, span)
		item.span = span
		item.offset = offset
	}
	return item
}

func (inst *RowInstance) splitLines(items []rowItem) []rowLine {
	if len(items) == 0 {
		return nil
	}
	if !inst.wrap {
		line := rowLine{items: append([]rowItem(nil), items...)}
		for _, item := range items {
			line.used += item.offset + item.span
		}
		return []rowLine{line}
	}

	lines := make([]rowLine, 0, len(items))
	current := rowLine{}
	for _, item := range items {
		needed := item.offset + item.span
		if len(current.items) > 0 && current.used+needed > gridColumns {
			lines = append(lines, current)
			current = rowLine{}
		}
		current.items = append(current.items, item)
		current.used += needed
	}
	if len(current.items) > 0 {
		lines = append(lines, current)
	}
	return lines
}

func (inst *RowInstance) buildLine(line rowLine, lineIndex int) rtui.VNode {
	if len(line.items) == 0 {
		return nil
	}

	remaining := gridColumns - line.used
	if remaining < 0 {
		remaining = 0
	}
	leading, between, trailing := distributeRemaining(remaining, len(line.items), inst.justify)

	children := make([]rtui.VNode, 0, len(line.items)*4)
	if spacer := newFlexSpacer(leading, inst.lineKey(lineIndex, "leading")); spacer != nil {
		children = append(children, spacer)
	}

	for index, item := range line.items {
		if index > 0 {
			if gap := newGapNode(inst.gutter, inst.lineKey(lineIndex, "gap-"+strconv.Itoa(index))); gap != nil {
				children = append(children, gap)
			}
			if spacer := newFlexSpacer(between[index-1], inst.lineKey(lineIndex, "between-"+strconv.Itoa(index))); spacer != nil {
				children = append(children, spacer)
			}
		}
		if spacer := newFlexSpacer(item.offset, inst.lineKey(lineIndex, "offset-"+strconv.Itoa(index))); spacer != nil {
			children = append(children, spacer)
		}
		children = append(children, inst.wrapCol(item, lineIndex, index))
	}

	if spacer := newFlexSpacer(trailing, inst.lineKey(lineIndex, "trailing")); spacer != nil {
		children = append(children, spacer)
	}

	builder := rtui.HStackBuilder(children...).Gap(0).AlignCross(inst.align).Stretch()
	builder.FillWidth()
	node := builder.Build()
	node.SetKey(inst.lineKey(lineIndex, "root"))
	return node
}

func (inst *RowInstance) wrapCol(item rowItem, lineIndex, itemIndex int) rtui.VNode {
	builder := rtui.VStackBuilder(item.node).Gap(0).AlignCross(rtui.AlignStart).Stretch().Flex(item.span)
	node := builder.Build()
	node.SetKey(inst.lineKey(lineIndex, "col-"+strconv.Itoa(itemIndex)))
	return node
}

func (inst *RowInstance) rootKey() string {
	if inst.key == "" {
		return "row-root"
	}
	return inst.key + "-root"
}

func (inst *ColInstance) rootKey() string {
	if inst.key == "" {
		return "col-root"
	}
	return inst.key + "-root"
}

func (inst *RowInstance) lineKey(index int, suffix string) string {
	base := "row"
	if inst.key != "" {
		base = inst.key
	}
	return base + "-line-" + strconv.Itoa(index) + "-" + suffix
}

func (inst *RowInstance) normalize() {
	if inst.gutter < 0 {
		inst.gutter = 0
	}
	if inst.verticalGutter < 0 {
		inst.verticalGutter = 0
	}
	if inst.width < 0 {
		inst.width = 0
	}
}

func (inst *ColInstance) normalize() {
	inst.span = normalizeSpan(inst.span)
	inst.offset = normalizeOffset(inst.offset, inst.span)
}

func normalizeSpan(span int) int {
	if span <= 0 {
		return gridColumns
	}
	if span > gridColumns {
		return gridColumns
	}
	return span
}

func normalizeOffset(offset, span int) int {
	if offset < 0 {
		offset = 0
	}
	maxOffset := gridColumns - span
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	return offset
}

func distributeRemaining(remaining, count int, justify rtui.Align) (int, []int, int) {
	if remaining <= 0 || count <= 0 {
		return 0, make([]int, maxInt(0, count-1)), 0
	}

	switch justify {
	case rtui.AlignEnd:
		return remaining, make([]int, maxInt(0, count-1)), 0
	case rtui.AlignCenter:
		leading := remaining / 2
		return leading, make([]int, maxInt(0, count-1)), remaining - leading
	case rtui.AlignSpaceBetween:
		if count <= 1 {
			return 0, make([]int, maxInt(0, count-1)), remaining
		}
		return 0, splitUnits(remaining, count-1), 0
	case rtui.AlignSpaceAround:
		slots := splitUnits(remaining, count+1)
		leading := slots[0]
		trailing := slots[len(slots)-1]
		between := make([]int, 0, count-1)
		for index := 1; index < len(slots)-1; index++ {
			between = append(between, slots[index])
		}
		return leading, between, trailing
	default:
		return 0, make([]int, maxInt(0, count-1)), remaining
	}
}

func splitUnits(total, parts int) []int {
	if parts <= 0 {
		return nil
	}
	values := make([]int, parts)
	base := total / parts
	remainder := total % parts
	for index := 0; index < parts; index++ {
		values[index] = base
		if index < remainder {
			values[index]++
		}
	}
	return values
}

func newFlexSpacer(flex int, key string) rtui.VNode {
	if flex <= 0 {
		return nil
	}
	node := rtui.Spacer().Flex(flex).Build()
	node.SetKey(key)
	return node
}

func newGapNode(width int, key string) rtui.VNode {
	if width <= 0 {
		return nil
	}
	node := rtui.Box().Width(width).Build()
	node.SetKey(key)
	return node
}

func filterChildren(children []rtui.VNode) []rtui.VNode {
	filtered := make([]rtui.VNode, 0, len(children))
	for _, child := range children {
		if child != nil {
			filtered = append(filtered, child)
		}
	}
	return filtered
}

func getChildrenProp(props rtui.Props) []rtui.VNode {
	if value, ok := props[propChildren]; ok {
		if children, ok := value.([]rtui.VNode); ok {
			return append([]rtui.VNode(nil), children...)
		}
	}
	return nil
}

func getChildrenPropWithDefault(props rtui.Props, def []rtui.VNode) []rtui.VNode {
	if children := getChildrenProp(props); children != nil {
		return children
	}
	return append([]rtui.VNode(nil), def...)
}

func getStringProp(props rtui.Props, key, def string) string {
	if value, ok := props[key]; ok {
		if text, ok := value.(string); ok {
			return text
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if value, ok := props[key]; ok {
		if number, ok := value.(int); ok {
			return number
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if value, ok := props[key]; ok {
		if flag, ok := value.(bool); ok {
			return flag
		}
	}
	return def
}

func getAlignProp(props rtui.Props, key string, def rtui.Align) rtui.Align {
	if value, ok := props[key]; ok {
		if align, ok := value.(rtui.Align); ok {
			return align
		}
	}
	return def
}

func getStyleProp(props rtui.Props, key string, def style.Style) style.Style {
	if value, ok := props[key]; ok {
		if s, ok := value.(style.Style); ok {
			return s
		}
	}
	return def
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
