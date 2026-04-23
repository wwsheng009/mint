package pagination

import (
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

type pageItemKind string

const (
	pageItemPrev    pageItemKind = "prev"
	pageItemNext    pageItemKind = "next"
	pageItemNumber  pageItemKind = "number"
	pageItemGap     pageItemKind = "gap"
	pageItemSummary pageItemKind = "summary"
)

type pageItem struct {
	kind      pageItemKind
	label     string
	target    int
	clickable bool
	disabled  bool
	current   bool
	width     int
}

// Instance is the runtime entity for Pagination.
type Instance struct {
	key                   string
	componentID           string
	total                 int
	pageSize              int
	currentPage           int
	currentPageControlled bool
	maxButtons            int
	showTotal             bool
	disabled              bool
	paginationStyle       style.Style
	selectedStyle         style.Style
	disabledStyle         style.Style
	pageIntent            intent.Intent
	pageIntentField       intent.FieldIntent
	bounds                [4]int
	dirty                 bool
	focused               bool
	parent                rtui.ComponentInstance
	intentEmitter         func(intent.Intent)
}

var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:                   proputil.GetString(props, propKey, ""),
		componentID:           proputil.GetString(props, propComponentID, ""),
		total:                 maxInt(0, proputil.GetInt(props, propTotal, 0)),
		pageSize:              maxInt(0, proputil.GetInt(props, propPageSize, 10)),
		currentPage:           maxInt(0, proputil.GetInt(props, propCurrentPage, 0)),
		currentPageControlled: proputil.GetBool(props, propCurrentPageControlled, false),
		maxButtons:            maxInt(3, proputil.GetInt(props, propMaxButtons, 5)),
		showTotal:             proputil.GetBool(props, propShowTotal, true),
		disabled:              proputil.GetBool(props, propDisabled, false),
		paginationStyle:       proputil.GetStyle(props, propPaginationStyle, style.Style{}),
		selectedStyle:         proputil.GetStyle(props, propSelectedStyle, style.Style{}.Reverse(true).Bold(true)),
		disabledStyle:         proputil.GetStyle(props, propDisabledStyle, style.Style{}.Foreground(style.BrightBlack)),
		pageIntent:            proputil.GetIntent(props, propPageIntent, nil),
		pageIntentField:       getFieldIntentProp(props, propPageIntent),
		dirty:                 true,
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
	old := *inst
	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.componentID = proputil.GetString(props, propComponentID, inst.componentID)
	inst.total = maxInt(0, proputil.GetInt(props, propTotal, inst.total))
	inst.pageSize = maxInt(0, proputil.GetInt(props, propPageSize, inst.pageSize))
	inst.currentPage = maxInt(0, proputil.GetInt(props, propCurrentPage, inst.currentPage))
	inst.currentPageControlled = proputil.GetBool(props, propCurrentPageControlled, inst.currentPageControlled)
	inst.maxButtons = maxInt(3, proputil.GetInt(props, propMaxButtons, inst.maxButtons))
	inst.showTotal = proputil.GetBool(props, propShowTotal, inst.showTotal)
	inst.disabled = proputil.GetBool(props, propDisabled, inst.disabled)
	inst.paginationStyle = proputil.GetStyle(props, propPaginationStyle, inst.paginationStyle)
	inst.selectedStyle = proputil.GetStyle(props, propSelectedStyle, inst.selectedStyle)
	inst.disabledStyle = proputil.GetStyle(props, propDisabledStyle, inst.disabledStyle)
	inst.pageIntent = proputil.GetIntent(props, propPageIntent, inst.pageIntent)
	inst.pageIntentField = getFieldIntentProp(props, propPageIntent)
	inst.normalize()

	changed := old.key != inst.key ||
		old.componentID != inst.componentID ||
		old.total != inst.total ||
		old.pageSize != inst.pageSize ||
		old.currentPage != inst.currentPage ||
		old.currentPageControlled != inst.currentPageControlled ||
		old.maxButtons != inst.maxButtons ||
		old.showTotal != inst.showTotal ||
		old.disabled != inst.disabled ||
		old.paginationStyle != inst.paginationStyle ||
		old.selectedStyle != inst.selectedStyle ||
		old.disabledStyle != inst.disabledStyle ||
		old.pageIntent != inst.pageIntent ||
		old.pageIntentField != inst.pageIntentField
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:                   inst.key,
		propComponentID:           inst.componentID,
		propTotal:                 inst.total,
		propPageSize:              inst.pageSize,
		propCurrentPage:           inst.currentPage,
		propCurrentPageControlled: inst.currentPageControlled,
		propMaxButtons:            inst.maxButtons,
		propShowTotal:             inst.showTotal,
		propDisabled:              inst.disabled,
		propPaginationStyle:       inst.paginationStyle,
		propSelectedStyle:         inst.selectedStyle,
		propDisabledStyle:         inst.disabledStyle,
		propPageIntent:            inst.pageIntent,
		propPageIntentField:       inst.pageIntentField,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }

func (inst *Instance) IsDirty() bool { return inst.dirty }

func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Parent() interface{} { return inst.parent }

func (inst *Instance) SetParent(parent rtui.ComponentInstance) { inst.parent = parent }

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) { inst.intentEmitter = fn }

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *Instance) SetFocus(focused bool) {
	if inst.focused == focused {
		return
	}
	inst.focused = focused
	inst.dirty = true
}

func (inst *Instance) HasFocus() bool { return inst.focused }

func (inst *Instance) IsDisabled() bool { return inst.disabled }

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := inst.itemsWidth(inst.pageItems())
	width = constraints.ConstrainWidth(width)
	return layout.Size{
		Width:  width,
		Height: constraints.ConstrainHeight(1),
	}
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	items := inst.pageItems()
	if len(items) == 0 {
		return nil
	}
	cmds := make([]paint.DrawCmd, 0, len(items))
	cursor := x
	for index, item := range items {
		itemStyle := inst.paginationStyle
		if item.disabled {
			itemStyle = inst.disabledStyle
		}
		if item.current {
			itemStyle = inst.selectedStyle
		}
		if inst.focused && item.current {
			itemStyle = itemStyle.Bold(true)
		}
		cmds = append(cmds, paint.DrawCmd{
			X:     cursor,
			Y:     y,
			Text:  item.label,
			Style: itemStyle,
		})
		cursor += item.width
		if index < len(items)-1 {
			cursor += 1
		}
	}
	return cmds
}

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil || inst.disabled {
		return false
	}
	switch act.Type {
	case action.ActionNavigateLeft, action.ActionNavigatePrev:
		return inst.applyPage(inst.currentPage - 1)
	case action.ActionNavigateRight, action.ActionNavigateNext:
		return inst.applyPage(inst.currentPage + 1)
	case action.ActionNavigateHome, action.ActionNavigateFirst:
		return inst.applyPage(0)
	case action.ActionNavigateEnd, action.ActionNavigateLast:
		return inst.applyPage(inst.pageCount() - 1)
	case action.ActionNavigatePageUp:
		return inst.applyPage(inst.currentPage - inst.windowSize())
	case action.ActionNavigatePageDown:
		return inst.applyPage(inst.currentPage + inst.windowSize())
	case action.ActionClick, action.ActionSelect, action.ActionEnter:
		return inst.handleActivate(act)
	}
	return false
}

func (inst *Instance) HandleIntent(i intent.Intent) bool {
	change, ok := i.(PageChangeIntent)
	if !ok {
		return false
	}
	if inst.componentID != "" && change.ComponentID != "" && change.ComponentID != inst.componentID {
		return false
	}
	return inst.applyPage(change.ToPage)
}

func (inst *Instance) GetCurrentPage() int { return inst.currentPage }

func (inst *Instance) GetPageCount() int { return inst.pageCount() }

func (inst *Instance) normalize() {
	inst.total = maxInt(0, inst.total)
	inst.pageSize = maxInt(0, inst.pageSize)
	inst.currentPage = maxInt(0, inst.currentPage)
	inst.maxButtons = maxInt(3, inst.maxButtons)
	pageCount := inst.pageCount()
	if pageCount < 1 {
		inst.currentPage = 0
		return
	}
	if inst.currentPage >= pageCount {
		inst.currentPage = pageCount - 1
	}
}

func (inst *Instance) pageCount() int {
	if inst.pageSize <= 0 {
		return 1
	}
	if inst.total <= 0 {
		return 1
	}
	return (inst.total + inst.pageSize - 1) / inst.pageSize
}

func (inst *Instance) windowSize() int {
	if inst.maxButtons < 1 {
		return 1
	}
	return inst.maxButtons
}

func (inst *Instance) applyPage(target int) bool {
	pageCount := inst.pageCount()
	if pageCount < 1 {
		pageCount = 1
	}
	clamped := clampInt(target, 0, pageCount-1)
	if inst.currentPage == clamped {
		return false
	}
	fromPage := inst.currentPage
	inst.currentPage = clamped
	inst.dirty = true
	inst.emitPageChanged(fromPage, clamped)
	return true
}

func (inst *Instance) handleActivate(act *action.Action) bool {
	if mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg); ok && mouseMsg != nil {
		return inst.handleClick(mouseMsg.LocalX)
	}
	return false
}

func (inst *Instance) handleClick(localX int) bool {
	items := inst.pageItems()
	cursor := 0
	for index, item := range items {
		if localX >= cursor && localX < cursor+item.width {
			if item.clickable && !item.disabled {
				return inst.applyPage(item.target)
			}
			return false
		}
		cursor += item.width
		if index < len(items)-1 {
			cursor++
		}
	}
	return false
}

func (inst *Instance) emitPageChanged(fromPage, toPage int) {
	if inst.intentEmitter == nil {
		return
	}
	pageCount := inst.pageCount()
	if inst.componentID != "" {
		inst.intentEmitter(PageChangeWithID(inst.componentID, fromPage, toPage, pageCount, inst.pageSize, inst.total))
	}
	if inst.pageIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.pageIntentField.GetField(),
			Value: strconv.Itoa(toPage),
		})
	} else if inst.pageIntent != nil {
		inst.intentEmitter(inst.pageIntent)
	}
}

func (inst *Instance) pageItems() []pageItem {
	pageCount := inst.pageCount()
	if pageCount < 1 {
		pageCount = 1
	}
	current := clampInt(inst.currentPage, 0, pageCount-1)

	items := make([]pageItem, 0, pageCount+4)
	items = append(items, newPageItem(pageItemPrev, "Prev", current-1, current == 0))

	for _, page := range inst.visiblePages(pageCount, current) {
		switch {
		case page < 0:
			items = append(items, newStaticPageItem(pageItemGap, "…"))
		default:
			label := strconv.Itoa(page + 1)
			items = append(items, pageItem{
				kind:      pageItemNumber,
				label:     bracketIfCurrent(label, page == current),
				target:    page,
				clickable: page != current,
				current:   page == current,
				width:     paint.StringWidth(bracketIfCurrent(label, page == current)),
			})
		}
	}

	items = append(items, newPageItem(pageItemNext, "Next", current+1, current >= pageCount-1))
	if inst.showTotal {
		summary := strconv.Itoa(inst.total) + " items"
		items = append(items, newStaticPageItem(pageItemSummary, summary))
	}
	return items
}

func (inst *Instance) visiblePages(pageCount, current int) []int {
	if pageCount <= inst.maxButtons {
		pages := make([]int, pageCount)
		for i := 0; i < pageCount; i++ {
			pages[i] = i
		}
		return pages
	}

	window := inst.maxButtons
	if window < 3 {
		window = 3
	}
	start := current - window/2
	end := start + window - 1
	if start < 0 {
		start = 0
		end = window - 1
	}
	if end >= pageCount {
		end = pageCount - 1
		start = maxInt(0, end-window+1)
	}

	pages := make([]int, 0, window+4)
	if start > 0 {
		pages = append(pages, 0)
		if start > 1 {
			pages = append(pages, -1)
		}
	}
	for page := start; page <= end; page++ {
		pages = append(pages, page)
	}
	if end < pageCount-1 {
		if end < pageCount-2 {
			pages = append(pages, -1)
		}
		pages = append(pages, pageCount-1)
	}
	return pages
}

func (inst *Instance) itemsWidth(items []pageItem) int {
	width := 0
	for index, item := range items {
		width += item.width
		if index < len(items)-1 {
			width++
		}
	}
	return width
}

func newPageItem(kind pageItemKind, label string, target int, disabled bool) pageItem {
	return pageItem{
		kind:      kind,
		label:     label,
		target:    target,
		clickable: !disabled,
		disabled:  disabled,
		width:     paint.StringWidth(label),
	}
}

func newStaticPageItem(kind pageItemKind, label string) pageItem {
	return pageItem{
		kind:  kind,
		label: label,
		width: paint.StringWidth(label),
	}
}

func bracketIfCurrent(label string, current bool) string {
	if !current {
		return label
	}
	return "[" + label + "]"
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func getFieldIntentProp(props rtui.Props, key string) intent.FieldIntent {
	if value, ok := props[key]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	if value, ok := props[key+"Field"]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	return nil
}

func collectTexts(cmds []paint.DrawCmd) string {
	parts := make([]string, 0, len(cmds))
	for _, cmd := range cmds {
		parts = append(parts, cmd.Text)
	}
	return strings.Join(parts, " ")
}
