package collapse

import (
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

// Instance is the runtime entity for Collapse components.
type Instance struct {
	key                  string
	componentID          string
	parent               rtui.ComponentInstance
	childInstances       []rtui.ComponentInstance
	items                []Item
	accordion            bool
	activeKeys           []string
	activeKeysControlled bool
	disabled             bool
	bordered             bool
	ghost                bool
	width                int
	collapseStyle        style.Style
	headerStyle          style.Style
	activeHeaderStyle    style.Style
	contentStyle         style.Style
	changeIntent         intent.Intent
	changeIntentField    intent.FieldIntent
	intentEmitter        func(intent.Intent)
	dirty                bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
	_ rtui.TreeNode                = (*Instance)(nil)
	_ rtui.TreeContainer           = (*Instance)(nil)
	_ intent.IntentHandler         = (*Instance)(nil)
	_ intent.TreeComponent         = (*Instance)(nil)
)

// NewInstance creates a new Collapse instance.
func NewInstance(props rtui.Props) *Instance {
	activeKeysControlled := proputil.GetBool(props, propActiveKeysControl, false)
	activeKeys := getStringsProp(props, propInitialActiveKeys, nil)
	if activeKeysControlled {
		activeKeys = getStringsProp(props, propActiveKeys, activeKeys)
	}
	inst := &Instance{
		key:                  proputil.GetString(props, propKey, ""),
		componentID:          proputil.GetString(props, propComponentID, ""),
		items:                normalizeItems(getItemsProp(props)),
		accordion:            proputil.GetBool(props, propAccordion, false),
		activeKeys:           cloneStrings(activeKeys),
		activeKeysControlled: activeKeysControlled,
		disabled:             proputil.GetBool(props, propDisabled, false),
		bordered:             proputil.GetBool(props, propBordered, true),
		ghost:                proputil.GetBool(props, propGhost, false),
		width:                proputil.GetInt(props, propWidth, 0),
		collapseStyle:        proputil.GetStyle(props, propStyle, style.Style{}),
		headerStyle:          proputil.GetStyle(props, propHeaderStyle, style.Style{}),
		activeHeaderStyle:    proputil.GetStyle(props, propActiveHeaderStyle, style.Style{}),
		contentStyle:         proputil.GetStyle(props, propContentStyle, style.Style{}),
		changeIntent:         proputil.GetIntent(props, propChangeIntent, nil),
		changeIntentField:    getFieldIntentProp(props, propChangeIntentField),
		dirty:                true,
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
	oldItems := cloneItems(inst.items)
	oldActiveKeys := cloneStrings(inst.activeKeys)
	oldAccordion := inst.accordion
	oldControlled := inst.activeKeysControlled
	oldDisabled := inst.disabled
	oldBordered := inst.bordered
	oldGhost := inst.ghost
	oldWidth := inst.width
	oldStyle := inst.collapseStyle
	oldHeaderStyle := inst.headerStyle
	oldActiveHeaderStyle := inst.activeHeaderStyle
	oldContentStyle := inst.contentStyle
	oldChangeIntent := inst.changeIntent
	oldChangeIntentField := inst.changeIntentField
	oldComponentID := inst.componentID

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.componentID = proputil.GetString(props, propComponentID, inst.componentID)
	inst.items = normalizeItems(getItemsProp(props))
	inst.accordion = proputil.GetBool(props, propAccordion, inst.accordion)
	nextControlled := proputil.GetBool(props, propActiveKeysControl, inst.activeKeysControlled)
	if nextControlled {
		inst.activeKeys = getStringsProp(props, propActiveKeys, inst.activeKeys)
	} else if oldControlled && !nextControlled {
		inst.activeKeys = getStringsProp(props, propInitialActiveKeys, inst.activeKeys)
	}
	inst.activeKeysControlled = nextControlled
	inst.disabled = proputil.GetBool(props, propDisabled, inst.disabled)
	inst.bordered = proputil.GetBool(props, propBordered, inst.bordered)
	inst.ghost = proputil.GetBool(props, propGhost, inst.ghost)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.collapseStyle = proputil.GetStyle(props, propStyle, inst.collapseStyle)
	inst.headerStyle = proputil.GetStyle(props, propHeaderStyle, inst.headerStyle)
	inst.activeHeaderStyle = proputil.GetStyle(props, propActiveHeaderStyle, inst.activeHeaderStyle)
	inst.contentStyle = proputil.GetStyle(props, propContentStyle, inst.contentStyle)
	inst.changeIntent = proputil.GetIntent(props, propChangeIntent, inst.changeIntent)
	inst.changeIntentField = getFieldIntentProp(props, propChangeIntentField)
	inst.normalize()

	changed := !itemsEqual(oldItems, inst.items) ||
		!stringSlicesEqual(oldActiveKeys, inst.activeKeys) ||
		oldAccordion != inst.accordion ||
		oldControlled != inst.activeKeysControlled ||
		oldDisabled != inst.disabled ||
		oldBordered != inst.bordered ||
		oldGhost != inst.ghost ||
		oldWidth != inst.width ||
		oldStyle != inst.collapseStyle ||
		oldHeaderStyle != inst.headerStyle ||
		oldActiveHeaderStyle != inst.activeHeaderStyle ||
		oldContentStyle != inst.contentStyle ||
		oldChangeIntent != inst.changeIntent ||
		oldChangeIntentField != inst.changeIntentField ||
		oldComponentID != inst.componentID
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propAccordion:         inst.accordion,
		propActiveHeaderStyle: inst.activeHeaderStyle,
		propActiveKeys:        cloneStrings(inst.activeKeys),
		propActiveKeysControl: inst.activeKeysControlled,
		propBordered:          inst.bordered,
		propChangeIntent:      inst.changeIntent,
		propChangeIntentField: inst.changeIntentField,
		propComponentID:       inst.componentID,
		propContentStyle:      inst.contentStyle,
		propDisabled:          inst.disabled,
		propGhost:             inst.ghost,
		propHeaderStyle:       inst.headerStyle,
		propItems:             cloneItems(inst.items),
		propKey:               inst.key,
		propStyle:             inst.collapseStyle,
		propWidth:             inst.width,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }

func (inst *Instance) IsDirty() bool { return inst.dirty }

func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

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

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	if len(inst.items) == 0 {
		return nil
	}

	panelNodes := make([]rtui.VNode, 0, len(inst.items))
	for index, item := range inst.items {
		panelNodes = append(panelNodes, inst.buildPanelVNode(index, item))
	}

	root := rtui.VStackBuilder(panelNodes...).
		Gap(inst.rootGap()).
		AlignCross(rtui.AlignStart)
	if inst.width > 0 {
		root.Width(inst.width)
	}
	if !inst.collapseStyle.IsEmpty() {
		root.SetStyleProps(inst.collapseStyle)
	}

	rootNode := root.Build()
	rootNode.SetKey("collapse-root")
	return []rtui.VNode{rootNode}
}

func (inst *Instance) HandleIntent(i intent.Intent) bool {
	if !intent.ShouldHandleIntentWithID(inst.componentID, i) {
		return false
	}

	switch value := i.(type) {
	case CollapseToggleIntent:
		return inst.toggleByIntent(value)
	default:
		return false
	}
}

// ActiveKeys returns the current expanded keys.
func (inst *Instance) ActiveKeys() []string {
	return cloneStrings(inst.activeKeys)
}

// IsExpanded reports whether the given key is expanded.
func (inst *Instance) IsExpanded(key string) bool {
	return containsString(inst.activeKeys, key)
}

func (inst *Instance) buildPanelVNode(index int, item Item) rtui.VNode {
	expanded := inst.IsExpanded(item.Key)
	headerButton := button.NewBuilder(inst.headerLabel(item, expanded)).
		Key("collapse-header-button-" + item.Key).
		Small().
		TextAlign(rtui.AlignStart).
		FocusStyle(button.FocusStyleBold).
		Style(inst.resolveHeaderStyle(expanded, item.Disabled)).
		Disabled(inst.disabled || item.Disabled).
		OnPress(inst.toggleIntent(index, item)).
		Build()

	headerRowChildren := []rtui.VNode{rtui.Flex(headerButton, 1)}
	if item.Extra != nil {
		headerRowChildren = append(headerRowChildren, item.Extra)
	}
	headerRowBuilder := rtui.HStackBuilder(headerRowChildren...).
		Gap(1).
		AlignCross(rtui.AlignCenter)
	headerRow := headerRowBuilder.Build()
	headerRow.SetKey("collapse-header-row-" + item.Key)

	children := []rtui.VNode{headerRow}
	if expanded && item.Content != nil {
		children = append(children, inst.buildBodyVNode(item))
	}

	panelBuilder := rtui.VStackBuilder(children...).
		Gap(0).
		AlignCross(rtui.AlignStart)
	if inst.width > 0 {
		panelBuilder.Width(inst.width)
	}
	if inst.bordered && !inst.ghost {
		panelBuilder.SingleBorder()
		panelBuilder.SetBorderColor(inst.resolveBorderColor(expanded, item.Disabled))
	}
	panelNode := panelBuilder.Build()
	panelNode.SetKey("collapse-panel-" + item.Key)
	return panelNode
}

func (inst *Instance) buildBodyVNode(item Item) rtui.VNode {
	bodyBuilder := rtui.HStackBuilder(
		textcomp.New("  ").Foreground(theme.Muted()),
		rtui.Flex(item.Content, 1),
	).
		Gap(0).
		AlignCross(rtui.AlignStart)
	if !inst.contentStyle.IsEmpty() {
		bodyBuilder.SetStyleProps(inst.contentStyle)
	}
	bodyNode := bodyBuilder.Build()
	bodyNode.SetKey("collapse-body-" + item.Key)
	return bodyNode
}

func (inst *Instance) headerLabel(item Item, expanded bool) string {
	header := strings.TrimSpace(item.Header)
	if header == "" {
		header = item.Key
	}
	indicator := "▶"
	if expanded {
		indicator = "▼"
	}
	return indicator + " " + header
}

func (inst *Instance) resolveHeaderStyle(expanded, disabled bool) style.Style {
	base := style.NewStyle().Merge(inst.collapseStyle).Merge(inst.headerStyle)
	switch {
	case inst.disabled || disabled:
		return base.Foreground(theme.DisabledFG())
	case expanded:
		active := style.NewStyle().Foreground(theme.Primary()).Bold(true)
		return base.Merge(active).Merge(inst.activeHeaderStyle)
	default:
		return base.Foreground(theme.Text())
	}
}

func (inst *Instance) resolveBorderColor(expanded, disabled bool) style.Color {
	switch {
	case inst.disabled || disabled:
		return theme.DisabledFG()
	case expanded:
		return theme.Primary()
	default:
		return theme.Muted()
	}
}

func (inst *Instance) toggleIntent(index int, item Item) CollapseToggleIntent {
	if inst.componentID != "" {
		return ToggleWithID(inst.componentID, item.Key, item.Header, index)
	}
	return Toggle(item.Key, item.Header, index)
}

func (inst *Instance) toggleByIntent(toggle CollapseToggleIntent) bool {
	index, item, ok := inst.resolveItem(toggle.ItemKey, toggle.Index)
	if !ok {
		return false
	}
	if inst.disabled || item.Disabled {
		return false
	}

	next, expanded := inst.nextActiveKeys(item.Key)
	if stringSlicesEqual(next, inst.activeKeys) {
		return false
	}

	inst.activeKeys = next
	inst.dirty = true
	inst.emitChange(index, item, expanded)
	return true
}

func (inst *Instance) resolveItem(itemKey string, index int) (int, Item, bool) {
	if itemKey != "" {
		for itemIndex, item := range inst.items {
			if item.Key == itemKey {
				return itemIndex, item, true
			}
		}
	}
	if index >= 0 && index < len(inst.items) {
		return index, inst.items[index], true
	}
	return -1, Item{}, false
}

func (inst *Instance) nextActiveKeys(targetKey string) ([]string, bool) {
	if inst.accordion {
		if containsString(inst.activeKeys, targetKey) {
			return nil, false
		}
		return []string{targetKey}, true
	}

	nextSet := make(map[string]struct{}, len(inst.activeKeys)+1)
	for _, key := range inst.activeKeys {
		nextSet[key] = struct{}{}
	}

	_, expanded := nextSet[targetKey]
	if expanded {
		delete(nextSet, targetKey)
	} else {
		nextSet[targetKey] = struct{}{}
	}

	result := make([]string, 0, len(nextSet))
	for _, item := range inst.items {
		if _, ok := nextSet[item.Key]; ok {
			result = append(result, item.Key)
		}
	}
	return result, !expanded
}

func (inst *Instance) emitChange(index int, item Item, expanded bool) {
	if inst.intentEmitter == nil {
		return
	}

	if inst.componentID != "" {
		inst.intentEmitter(ChangeWithID(inst.componentID, inst.activeKeys, item.Key, item.Header, expanded, index, inst.accordion))
	} else {
		inst.intentEmitter(Change(inst.activeKeys, item.Key, item.Header, expanded, index, inst.accordion))
	}
	if inst.changeIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: strings.Join(inst.activeKeys, ","),
		})
	}
	if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
}

func (inst *Instance) normalize() {
	inst.items = normalizeItems(inst.items)
	inst.activeKeys = normalizeActiveKeys(inst.items, inst.activeKeys, inst.accordion)
	if inst.ghost {
		inst.bordered = false
	}
}

func (inst *Instance) rootGap() int {
	if inst.bordered && !inst.ghost {
		return 0
	}
	return 1
}

func getItemsProp(props rtui.Props) []Item {
	if items, ok := props[propItems].([]Item); ok {
		return cloneItems(items)
	}
	return nil
}

func getStringsProp(props rtui.Props, key string, def []string) []string {
	if value, ok := props[key]; ok {
		if values, ok := value.([]string); ok {
			return cloneStrings(values)
		}
	}
	return cloneStrings(def)
}

func getFieldIntentProp(props rtui.Props, key string) intent.FieldIntent {
	if value, ok := props[key]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	return nil
}

func normalizeActiveKeys(items []Item, requested []string, accordion bool) []string {
	if len(items) == 0 || len(requested) == 0 {
		return nil
	}

	allowed := make(map[string]struct{}, len(requested))
	for _, key := range requested {
		if key == "" {
			continue
		}
		allowed[key] = struct{}{}
	}

	if accordion {
		for _, item := range items {
			if _, ok := allowed[item.Key]; ok {
				return []string{item.Key}
			}
		}
		return nil
	}

	result := make([]string, 0, len(allowed))
	for _, item := range items {
		if _, ok := allowed[item.Key]; ok {
			result = append(result, item.Key)
		}
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func itemsEqual(a, b []Item) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func stringSlicesEqual(a, b []string) bool {
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
