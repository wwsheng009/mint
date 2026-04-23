package anchor

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"github.com/wwsheng009/mint/ui/components/list"
)

type flatItem struct {
	Key      string
	Title    string
	Href     string
	Disabled bool
	Depth    int
}

// Instance is the runtime entity for Anchor components.
type Instance struct {
	key                 string
	componentID         string
	parent              rtui.ComponentInstance
	childInstances      []rtui.ComponentInstance
	title               string
	items               []Item
	activeKey           string
	activeKeyControlled bool
	viewportHeight      int
	width               int
	showBorder          bool
	anchorStyle         style.Style
	currentStyle        style.Style
	changeIntent        runtimeintent.Intent
	changeIntentField   runtimeintent.FieldIntent
	formID              string
	listVersion         int
	dirty               bool
	intentEmitter       func(runtimeintent.Intent)
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
	_ rtui.TreeNode                = (*Instance)(nil)
	_ rtui.TreeContainer           = (*Instance)(nil)
	_ runtimeintent.IntentHandler  = (*Instance)(nil)
	_ runtimeintent.TreeComponent  = (*Instance)(nil)
)

// NewInstance creates a new Anchor instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:                 proputil.GetString(props, propKey, ""),
		componentID:         proputil.GetString(props, propComponentID, ""),
		title:               proputil.GetString(props, propTitle, ""),
		items:               getItemsProp(props),
		activeKey:           normalizeKey(proputil.GetString(props, propActiveKey, "")),
		activeKeyControlled: proputil.GetBool(props, propActiveKeyControlled, false),
		viewportHeight:      proputil.GetInt(props, propViewportHeight, 0),
		width:               proputil.GetInt(props, propWidth, 0),
		showBorder:          proputil.GetBool(props, propShowBorder, false),
		anchorStyle:         proputil.GetStyle(props, propStyle, style.Style{}),
		currentStyle:        proputil.GetStyle(props, propCurrentStyle, style.Style{}),
		changeIntent:        proputil.GetIntent(props, propChangeIntent, nil),
		changeIntentField:   getFieldIntentProp(props, propChangeIntent),
		formID:              proputil.GetString(props, propFormID, ""),
		dirty:               true,
	}
	inst.normalize()
	return inst
}

func (inst *Instance) Key() string { return inst.key }

func (inst *Instance) SetKey(key string) { inst.key = key }

func (inst *Instance) Parent() interface{} { return inst.parent }

func (inst *Instance) SetParent(parent rtui.ComponentInstance) { inst.parent = parent }

func (inst *Instance) Children() []rtui.ComponentInstance {
	return append([]rtui.ComponentInstance(nil), inst.childInstances...)
}

func (inst *Instance) AddChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	for index, existing := range inst.childInstances {
		if existing == child || existing.Key() == child.Key() {
			inst.childInstances[index] = child
			if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
				setter.SetParent(inst)
			}
			return
		}
	}
	inst.childInstances = append(inst.childInstances, child)
	if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
		setter.SetParent(inst)
	}
}

func (inst *Instance) RemoveChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	for index, existing := range inst.childInstances {
		if existing != child {
			continue
		}
		inst.childInstances = append(inst.childInstances[:index], inst.childInstances[index+1:]...)
		if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
			setter.SetParent(nil)
		}
		return
	}
}

func (inst *Instance) ClearChildren() {
	for _, child := range inst.childInstances {
		if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
			setter.SetParent(nil)
		}
	}
	inst.childInstances = inst.childInstances[:0]
}

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) Destroy() {}

func (inst *Instance) OnMount() {}

func (inst *Instance) OnUnmount() {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldTitle := inst.title
	oldItems := cloneItems(inst.items)
	oldActiveKey := inst.activeKey
	oldControlled := inst.activeKeyControlled
	oldViewportHeight := inst.viewportHeight
	oldWidth := inst.width
	oldShowBorder := inst.showBorder
	oldStyle := inst.anchorStyle
	oldCurrentStyle := inst.currentStyle
	oldChangeIntent := inst.changeIntent
	oldFormID := inst.formID

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.componentID = proputil.GetString(props, propComponentID, inst.componentID)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.items = getItemsPropOr(props, inst.items)
	nextControlled := proputil.GetBool(props, propActiveKeyControlled, inst.activeKeyControlled)
	if nextControlled {
		inst.activeKey = normalizeKey(proputil.GetString(props, propActiveKey, inst.activeKey))
	} else if inst.activeKeyControlled && !nextControlled {
		inst.activeKey = normalizeKey(proputil.GetString(props, propActiveKey, inst.activeKey))
	}
	inst.activeKeyControlled = nextControlled
	inst.viewportHeight = proputil.GetInt(props, propViewportHeight, inst.viewportHeight)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.showBorder = proputil.GetBool(props, propShowBorder, inst.showBorder)
	inst.anchorStyle = proputil.GetStyle(props, propStyle, inst.anchorStyle)
	inst.currentStyle = proputil.GetStyle(props, propCurrentStyle, inst.currentStyle)
	inst.changeIntent = proputil.GetIntent(props, propChangeIntent, inst.changeIntent)
	inst.changeIntentField = getFieldIntentProp(props, propChangeIntent)
	inst.formID = proputil.GetString(props, propFormID, inst.formID)
	inst.normalize()

	changed := oldTitle != inst.title ||
		!reflect.DeepEqual(oldItems, inst.items) ||
		oldActiveKey != inst.activeKey ||
		oldControlled != inst.activeKeyControlled ||
		oldViewportHeight != inst.viewportHeight ||
		oldWidth != inst.width ||
		oldShowBorder != inst.showBorder ||
		oldStyle != inst.anchorStyle ||
		oldCurrentStyle != inst.currentStyle ||
		!reflect.DeepEqual(oldChangeIntent, inst.changeIntent) ||
		oldFormID != inst.formID
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propActiveKey:           inst.activeKey,
		propActiveKeyControlled: inst.activeKeyControlled,
		propChangeIntent:        inst.changeIntent,
		propComponentID:         inst.componentID,
		propCurrentStyle:        inst.currentStyle,
		propFormID:              inst.formID,
		propItems:               cloneItems(inst.items),
		propKey:                 inst.key,
		propShowBorder:          inst.showBorder,
		propStyle:               inst.anchorStyle,
		propTitle:               inst.title,
		propViewportHeight:      inst.viewportHeight,
		propWidth:               inst.width,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }

func (inst *Instance) IsDirty() bool { return inst.dirty }

func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) SetIntentEmitter(fn func(runtimeintent.Intent)) { inst.intentEmitter = fn }

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	items := flattenItems(inst.items)
	rowStyleFn := func(index int, row string) style.Style {
		if index < 0 || index >= len(items) {
			return style.Style{}
		}
		if items[index].Disabled {
			return style.NewStyle().
				Foreground(theme.DisabledFG()).
				Background(theme.Surface()).
				Merge(inst.anchorStyle)
		}
		return inst.anchorStyle
	}

	builder := list.NewBuilder().
		Key(inst.listKey()).
		ComponentID(inst.listComponentID()).
		Header(inst.title).
		Rows(renderRows(items)).
		SelectedIndex(inst.activeIndex(items)).
		ViewportHeight(inst.effectiveViewportHeight(len(items))).
		ShowBorder(inst.showBorder).
		ShowSeparator(inst.title != "").
		RowStyle(inst.anchorStyle).
		HeaderStyle(inst.anchorStyle.Bold(true)).
		RowStyleFn(rowStyleFn)
	if !inst.currentStyle.IsEmpty() {
		builder.SelectedStyle(inst.currentStyle)
	}

	listNode := builder.Build()
	if inst.width <= 0 {
		return []rtui.VNode{listNode}
	}

	root := rtui.VStackBuilder(listNode).Gap(0).AlignCross(rtui.AlignStart).Stretch()
	root.Width(inst.width)
	node := root.Build()
	node.SetKey(inst.rootKey())
	return []rtui.VNode{node}
}

func (inst *Instance) HandleIntent(i runtimeintent.Intent) bool {
	if rowSelect, ok := i.(list.RowSelectIntent); ok && rowSelect.ComponentID == inst.listComponentID() {
		return inst.handleRowSelect(rowSelect.SelectedIndex)
	}
	if !runtimeintent.ShouldHandleIntentWithID(inst.componentID, i) {
		return false
	}
	activate, ok := i.(ActivateIntent)
	if !ok {
		return false
	}
	return inst.activateKey(normalizeKey(activate.Key))
}

func (inst *Instance) handleRowSelect(index int) bool {
	items := flattenItems(inst.items)
	if index < 0 || index >= len(items) {
		return false
	}
	item := items[index]
	if item.Disabled {
		inst.listVersion++
		inst.dirty = true
		return true
	}
	return inst.activateItem(item)
}

func (inst *Instance) activateKey(key string) bool {
	items := flattenItems(inst.items)
	for _, item := range items {
		if item.Key != key {
			continue
		}
		if item.Disabled {
			return false
		}
		return inst.activateItem(item)
	}
	return false
}

func (inst *Instance) activateItem(item flatItem) bool {
	current := inst.activeKey
	if current == item.Key {
		return false
	}
	if inst.activeKeyControlled {
		inst.listVersion++
		inst.dirty = true
	} else {
		inst.activeKey = item.Key
		inst.dirty = true
	}
	inst.emitChange(item)
	return true
}

func (inst *Instance) emitChange(item flatItem) {
	if inst.intentEmitter == nil {
		return
	}

	inst.intentEmitter(ChangeWithID(inst.componentID, item.Key, item.Href, item.Title))

	if inst.formID != "" && inst.changeIntentField != nil {
		runtimeintent.Emit(inst, form.FieldChange(inst.formID, inst.changeIntentField.GetField(), item.Key, true))
		return
	}
	if inst.changeIntentField != nil {
		inst.intentEmitter(runtimeintent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: item.Key,
		})
		return
	}
	if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
}

func (inst *Instance) normalize() {
	inst.items = cloneItems(inst.items)
	inst.activeKey = normalizeKey(inst.activeKey)
	if inst.width < 0 {
		inst.width = 0
	}
	if inst.viewportHeight < 0 {
		inst.viewportHeight = 0
	}

	if inst.activeKeyControlled {
		if !hasKey(flattenItems(inst.items), inst.activeKey) {
			inst.activeKey = ""
		}
		return
	}

	items := flattenItems(inst.items)
	if hasEnabledKey(items, inst.activeKey) {
		return
	}
	inst.activeKey = firstEnabledKey(items)
}

func (inst *Instance) effectiveViewportHeight(total int) int {
	if total <= 0 {
		return 1
	}
	if inst.viewportHeight > 0 {
		return inst.viewportHeight
	}
	return total
}

func (inst *Instance) activeIndex(items []flatItem) int {
	for index, item := range items {
		if item.Key == inst.activeKey {
			return index
		}
	}
	return -1
}

func (inst *Instance) listComponentID() string {
	base := inst.componentID
	if base == "" {
		base = inst.key
	}
	if base == "" {
		base = "anchor"
	}
	return base + "-list"
}

func (inst *Instance) listKey() string {
	return inst.listComponentID() + "-" + strconv.Itoa(inst.listVersion)
}

func (inst *Instance) rootKey() string {
	base := inst.componentID
	if base == "" {
		base = inst.key
	}
	if base == "" {
		base = "anchor"
	}
	return base + "-root"
}

func flattenItems(items []Item) []flatItem {
	if len(items) == 0 {
		return nil
	}
	flattened := make([]flatItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	var visit func([]Item, int)
	visit = func(levelItems []Item, depth int) {
		for _, item := range levelItems {
			if _, exists := seen[item.Key]; exists {
				continue
			}
			seen[item.Key] = struct{}{}
			flattened = append(flattened, flatItem{
				Key:      item.Key,
				Title:    item.Title,
				Href:     item.Href,
				Disabled: item.Disabled,
				Depth:    depth,
			})
			if len(item.Children) > 0 {
				visit(item.Children, depth+1)
			}
		}
	}
	visit(items, 0)
	return flattened
}

func renderRows(items []flatItem) []string {
	if len(items) == 0 {
		return nil
	}
	rows := make([]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, strings.Repeat("  ", item.Depth)+item.Title)
	}
	return rows
}

func normalizeKey(key string) string {
	key = strings.TrimSpace(key)
	return strings.TrimPrefix(key, "#")
}

func hasKey(items []flatItem, key string) bool {
	key = normalizeKey(key)
	if key == "" {
		return false
	}
	for _, item := range items {
		if item.Key == key {
			return true
		}
	}
	return false
}

func hasEnabledKey(items []flatItem, key string) bool {
	key = normalizeKey(key)
	if key == "" {
		return false
	}
	for _, item := range items {
		if item.Key == key && !item.Disabled {
			return true
		}
	}
	return false
}

func firstEnabledKey(items []flatItem) string {
	for _, item := range items {
		if !item.Disabled {
			return item.Key
		}
	}
	return ""
}

func getItemsProp(props rtui.Props) []Item {
	if value, ok := props[propItems]; ok {
		if items, ok := value.([]Item); ok {
			return cloneItems(items)
		}
	}
	return nil
}

func getItemsPropOr(props rtui.Props, def []Item) []Item {
	if _, ok := props[propItems]; ok {
		return getItemsProp(props)
	}
	return cloneItems(def)
}

func getFieldIntentProp(props rtui.Props, key string) runtimeintent.FieldIntent {
	if value, ok := props[key]; ok {
		if fieldIntent, ok := value.(runtimeintent.FieldIntent); ok {
			return fieldIntent
		}
	}
	return nil
}
