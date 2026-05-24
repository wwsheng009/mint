package transfer

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"github.com/wwsheng009/mint/ui/components/list"
)

// Instance is the runtime entity for Transfer components.
type Instance struct {
	key                  string
	componentID          string
	parent               rtui.ComponentInstance
	childInstances       []rtui.ComponentInstance
	items                []Item
	titles               [2]string
	operations           [2]string
	searchable           bool
	searchControlled     bool
	searchPlaceholders   [2]string
	sourceSearch         string
	targetSearch         string
	targetKeys           []string
	targetKeysControlled bool
	selectedSourceKeys   []string
	selectedTargetKeys   []string
	listWidth            int
	listHeight           int
	width                int
	rootStyle            style.Style
	changeIntent         runtimeintent.Intent
	changeIntentField    runtimeintent.FieldIntent
	formID               string
	sourceListVersion    int
	targetListVersion    int
	dirty                bool
	intentEmitter        func(runtimeintent.Intent)
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
	_ rtui.TreeNode                = (*Instance)(nil)
	_ rtui.TreeContainer           = (*Instance)(nil)
	_ runtimeintent.IntentHandler  = (*Instance)(nil)
	_ runtimeintent.TreeComponent  = (*Instance)(nil)
)

// NewInstance creates a new Transfer instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:                  proputil.GetString(props, propKey, ""),
		componentID:          proputil.GetString(props, propComponentID, ""),
		items:                getItemsProp(props),
		titles:               getTitlesProp(props, defaultTitles),
		operations:           getOperationsProp(props, defaultOperations),
		searchable:           proputil.GetBool(props, propSearchable, false),
		searchControlled:     proputil.GetBool(props, propSearchControlled, false),
		searchPlaceholders:   getSearchPlaceholdersProp(props, defaultSearchPlaceholders),
		sourceSearch:         strings.TrimSpace(proputil.GetString(props, propSourceSearch, "")),
		targetSearch:         strings.TrimSpace(proputil.GetString(props, propTargetSearch, "")),
		targetKeys:           getStringSliceProp(props, propTargetKeys, nil),
		targetKeysControlled: proputil.GetBool(props, propTargetKeysControlled, false),
		listWidth:            proputil.GetInt(props, propListWidth, defaultListWidth),
		listHeight:           proputil.GetInt(props, propListHeight, defaultListHeight),
		width:                proputil.GetInt(props, propWidth, 0),
		rootStyle:            proputil.GetStyle(props, propStyle, style.Style{}),
		changeIntent:         proputil.GetIntent(props, propChangeIntent, nil),
		changeIntentField:    getFieldIntentProp(props, propChangeIntent),
		formID:               proputil.GetString(props, propFormID, ""),
		dirty:                true,
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
	oldItems := cloneItems(inst.items)
	oldTitles := inst.titles
	oldOperations := inst.operations
	oldSearchable := inst.searchable
	oldSearchControlled := inst.searchControlled
	oldSearchPlaceholders := inst.searchPlaceholders
	oldSourceSearch := inst.sourceSearch
	oldTargetSearch := inst.targetSearch
	oldTargetKeys := append([]string(nil), inst.targetKeys...)
	oldControlled := inst.targetKeysControlled
	oldListWidth := inst.listWidth
	oldListHeight := inst.listHeight
	oldWidth := inst.width
	oldStyle := inst.rootStyle
	oldChangeIntent := inst.changeIntent
	oldFormID := inst.formID
	oldSelectedSource := append([]string(nil), inst.selectedSourceKeys...)
	oldSelectedTarget := append([]string(nil), inst.selectedTargetKeys...)

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.componentID = proputil.GetString(props, propComponentID, inst.componentID)
	inst.items = getItemsPropOr(props, inst.items)
	inst.titles = getTitlesProp(props, inst.titles)
	inst.operations = getOperationsProp(props, inst.operations)
	inst.searchable = proputil.GetBool(props, propSearchable, inst.searchable)
	nextSearchControlled := proputil.GetBool(props, propSearchControlled, inst.searchControlled)
	if nextSearchControlled || inst.searchControlled != nextSearchControlled {
		inst.sourceSearch = strings.TrimSpace(proputil.GetString(props, propSourceSearch, inst.sourceSearch))
		inst.targetSearch = strings.TrimSpace(proputil.GetString(props, propTargetSearch, inst.targetSearch))
	}
	inst.searchControlled = nextSearchControlled
	inst.searchPlaceholders = getSearchPlaceholdersProp(props, inst.searchPlaceholders)
	nextControlled := proputil.GetBool(props, propTargetKeysControlled, inst.targetKeysControlled)
	if nextControlled {
		inst.targetKeys = getStringSliceProp(props, propTargetKeys, inst.targetKeys)
	} else if inst.targetKeysControlled && !nextControlled {
		inst.targetKeys = getStringSliceProp(props, propTargetKeys, inst.targetKeys)
	}
	inst.targetKeysControlled = nextControlled
	inst.listWidth = proputil.GetInt(props, propListWidth, inst.listWidth)
	inst.listHeight = proputil.GetInt(props, propListHeight, inst.listHeight)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.rootStyle = proputil.GetStyle(props, propStyle, inst.rootStyle)
	inst.changeIntent = proputil.GetIntent(props, propChangeIntent, inst.changeIntent)
	inst.changeIntentField = getFieldIntentProp(props, propChangeIntent)
	inst.formID = proputil.GetString(props, propFormID, inst.formID)
	inst.normalize()

	changed := !reflect.DeepEqual(oldItems, inst.items) ||
		oldTitles != inst.titles ||
		oldOperations != inst.operations ||
		oldSearchable != inst.searchable ||
		oldSearchControlled != inst.searchControlled ||
		oldSearchPlaceholders != inst.searchPlaceholders ||
		oldSourceSearch != inst.sourceSearch ||
		oldTargetSearch != inst.targetSearch ||
		!equalStrings(oldTargetKeys, inst.targetKeys) ||
		oldControlled != inst.targetKeysControlled ||
		oldListWidth != inst.listWidth ||
		oldListHeight != inst.listHeight ||
		oldWidth != inst.width ||
		oldStyle != inst.rootStyle ||
		!reflect.DeepEqual(oldChangeIntent, inst.changeIntent) ||
		oldFormID != inst.formID ||
		!equalStrings(oldSelectedSource, inst.selectedSourceKeys) ||
		!equalStrings(oldSelectedTarget, inst.selectedTargetKeys)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propChangeIntent:         inst.changeIntent,
		propComponentID:          inst.componentID,
		propFormID:               inst.formID,
		propItems:                cloneItems(inst.items),
		propKey:                  inst.key,
		propListHeight:           inst.listHeight,
		propListWidth:            inst.listWidth,
		propOperations:           inst.operations,
		propSearchable:           inst.searchable,
		propSearchControlled:     inst.searchControlled,
		propSearchPlaceholders:   inst.searchPlaceholders,
		propSourceSearch:         inst.sourceSearch,
		propStyle:                inst.rootStyle,
		propTargetKeys:           append([]string(nil), inst.targetKeys...),
		propTargetKeysControlled: inst.targetKeysControlled,
		propTargetSearch:         inst.targetSearch,
		propTitles:               inst.titles,
		propWidth:                inst.width,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }

func (inst *Instance) IsDirty() bool { return inst.dirty }

func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) SetIntentEmitter(fn func(runtimeintent.Intent)) { inst.intentEmitter = fn }

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	root := rtui.HStackBuilder(
		inst.buildListNode(inst.titles[0], inst.sourceItems(), inst.sourceVisibleItems(), inst.selectedSourceKeys, inst.sourceListID(), inst.sourceListKey(), inst.sourceSearchField(), inst.sourceSearch, inst.searchPlaceholders[0]),
		inst.buildOperationsNode(),
		inst.buildListNode(inst.titles[1], inst.targetItems(), inst.targetVisibleItems(), inst.selectedTargetKeys, inst.targetListID(), inst.targetListKey(), inst.targetSearchField(), inst.targetSearch, inst.searchPlaceholders[1]),
	).Gap(1).AlignCross(rtui.AlignStart)
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

func (inst *Instance) HandleIntent(i runtimeintent.Intent) bool {
	if selection, ok := i.(list.SelectionChangeIntent); ok {
		return inst.handleSelectionChange(selection)
	}
	if search, ok := i.(SearchChangeIntent); ok {
		return inst.handleSearchIntent(search)
	}
	if change, ok := i.(runtimeintent.FieldChangeIntent); ok {
		return inst.handleSearchChange(change)
	}
	if !runtimeintent.ShouldHandleIntentWithID(inst.componentID, i) {
		return false
	}
	move, ok := i.(MoveIntent)
	if !ok {
		return false
	}
	return inst.move(move.Direction)
}

func (inst *Instance) handleSelectionChange(selection list.SelectionChangeIntent) bool {
	switch selection.ComponentID {
	case inst.sourceListID():
		keys, rejected := selectedKeysFromIndices(inst.sourceVisibleItems(), selection.CheckedIndices)
		if rejected {
			inst.sourceListVersion++
		}
		if equalStrings(inst.selectedSourceKeys, keys) && !rejected {
			return false
		}
		inst.selectedSourceKeys = keys
		inst.dirty = true
		return true
	case inst.targetListID():
		keys, rejected := selectedKeysFromIndices(inst.targetVisibleItems(), selection.CheckedIndices)
		if rejected {
			inst.targetListVersion++
		}
		if equalStrings(inst.selectedTargetKeys, keys) && !rejected {
			return false
		}
		inst.selectedTargetKeys = keys
		inst.dirty = true
		return true
	default:
		return false
	}
}

func (inst *Instance) handleSearchChange(change runtimeintent.FieldChangeIntent) bool {
	if change.Field != inst.sourceSearchField() && change.Field != inst.targetSearchField() {
		return false
	}
	side := SearchSideSource
	if change.Field == inst.targetSearchField() {
		side = SearchSideTarget
	}
	return inst.applySearch(side, change.Value)
}

func (inst *Instance) handleSearchIntent(search SearchChangeIntent) bool {
	if !runtimeintent.ShouldHandleIntentWithID(inst.componentID, search) {
		return false
	}
	return inst.applySearch(search.Side, search.Value)
}

func (inst *Instance) applySearch(side SearchSide, rawValue string) bool {
	value := strings.TrimSpace(rawValue)
	switch side {
	case SearchSideSource:
		if inst.sourceSearch == value {
			return false
		}
		inst.sourceSearch = value
		inst.selectedSourceKeys = normalizeKeysForItems(inst.sourceVisibleItems(), inst.selectedSourceKeys)
		inst.sourceListVersion++
	case SearchSideTarget:
		if inst.targetSearch == value {
			return false
		}
		inst.targetSearch = value
		inst.selectedTargetKeys = normalizeKeysForItems(inst.targetVisibleItems(), inst.selectedTargetKeys)
		inst.targetListVersion++
	default:
		return false
	}
	inst.dirty = true
	return true
}

func (inst *Instance) move(direction MoveDirection) bool {
	moved := inst.movableKeys(direction)
	if len(moved) == 0 {
		return false
	}

	switch direction {
	case MoveDirectionToTarget:
		inst.targetKeys = addKeys(inst.items, inst.targetKeys, moved)
		inst.selectedSourceKeys = nil
		inst.selectedTargetKeys = removeKeys(inst.selectedTargetKeys, moved)
	case MoveDirectionToSource:
		inst.targetKeys = removeTargetKeys(inst.targetKeys, moved)
		inst.selectedTargetKeys = nil
		inst.selectedSourceKeys = removeKeys(inst.selectedSourceKeys, moved)
	default:
		return false
	}

	inst.normalize()
	inst.dirty = true
	inst.emitChange(direction, moved)
	return true
}

func (inst *Instance) buildListNode(title string, allItems, visibleItems []Item, selectedKeys []string, componentID, key, searchField, searchValue, searchPlaceholder string) rtui.VNode {
	rowStyleFn := func(index int, row string) style.Style {
		if index < 0 || index >= len(visibleItems) {
			return style.Style{}
		}
		if visibleItems[index].Disabled {
			return style.NewStyle().Foreground(theme.DisabledFG()).Background(theme.Surface())
		}
		return style.Style{}
	}

	listNode := list.NewBuilder().
		Key(key).
		ComponentID(componentID).
		Header(titleWithCount(title, len(visibleItems), len(allItems))).
		Rows(renderRows(visibleItems)).
		MultiSelect().
		CheckedIndices(checkedIndices(visibleItems, selectedKeys)...).
		ViewportHeight(inst.listHeight).
		RowStyleFn(rowStyleFn).
		Build()

	children := []rtui.VNode{listNode}
	if inst.searchable {
		searchNode := input.NewBuilder().
			Key(searchField).
			SetID(searchField).
			Search().
			Placeholder(searchPlaceholder).
			Value(searchValue).
			Width(inst.listWidth).
			OnChange(SearchChangeIntent{ComponentID: inst.componentID, Side: searchSideFromField(inst, searchField)}).
			Build()
		children = append([]rtui.VNode{searchNode}, children...)
	}

	wrapper := rtui.VStackBuilder(children...).Gap(0).AlignCross(rtui.AlignStart).Stretch()
	if inst.listWidth > 0 {
		wrapper.Width(inst.listWidth)
	}
	node := wrapper.Build()
	node.SetKey(key + "-wrapper")
	return node
}

func (inst *Instance) buildOperationsNode() rtui.VNode {
	moveToTargetDisabled := len(inst.movableKeys(MoveDirectionToTarget)) == 0
	moveToSourceDisabled := len(inst.movableKeys(MoveDirectionToSource)) == 0

	toTarget := button.NewBuilder(inst.operations[0]).
		Key(inst.baseKey("to-target")).
		Small().
		OnPress(MoveToTargetWithID(inst.componentID)).
		Disabled(moveToTargetDisabled).
		Build()
	toSource := button.NewBuilder(inst.operations[1]).
		Key(inst.baseKey("to-source")).
		Small().
		OnPress(MoveToSourceWithID(inst.componentID)).
		Disabled(moveToSourceDisabled).
		Build()

	node := rtui.VStackBuilder(toTarget, toSource).Gap(1).AlignCross(rtui.AlignCenter).Build()
	node.SetKey(inst.baseKey("operations"))
	return node
}

func (inst *Instance) normalize() {
	inst.items = cloneItems(inst.items)
	if inst.titles[0] == "" {
		inst.titles[0] = defaultTitles[0]
	}
	if inst.titles[1] == "" {
		inst.titles[1] = defaultTitles[1]
	}
	if inst.operations[0] == "" {
		inst.operations[0] = defaultOperations[0]
	}
	if inst.operations[1] == "" {
		inst.operations[1] = defaultOperations[1]
	}
	if inst.searchPlaceholders[0] == "" {
		inst.searchPlaceholders[0] = defaultSearchPlaceholders[0]
	}
	if inst.searchPlaceholders[1] == "" {
		inst.searchPlaceholders[1] = defaultSearchPlaceholders[1]
	}
	if inst.listWidth <= 0 {
		inst.listWidth = defaultListWidth
	}
	if inst.listHeight <= 0 {
		inst.listHeight = defaultListHeight
	}
	if inst.width < 0 {
		inst.width = 0
	}
	inst.targetKeys = normalizeTargetKeys(inst.items, inst.targetKeys)
	inst.selectedSourceKeys = normalizeKeysForItems(inst.sourceItems(), inst.selectedSourceKeys)
	inst.selectedTargetKeys = normalizeKeysForItems(inst.targetItems(), inst.selectedTargetKeys)
}

func (inst *Instance) emitChange(direction MoveDirection, movedKeys []string) {
	if inst.intentEmitter == nil {
		return
	}

	sourceKeys := sourceKeys(inst.items, inst.targetKeys)
	targetKeys := append([]string(nil), inst.targetKeys...)
	inst.intentEmitter(ChangeWithID(inst.componentID, direction, movedKeys, sourceKeys, targetKeys))

	value := strings.Join(targetKeys, defaultValueSep)
	if inst.formID != "" && inst.changeIntentField != nil {
		runtimeintent.Emit(inst, form.FieldChange(inst.formID, inst.changeIntentField.GetField(), value, true))
		return
	}
	if inst.changeIntentField != nil {
		inst.intentEmitter(runtimeintent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: value,
		})
		return
	}
	if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
}

func (inst *Instance) sourceItems() []Item {
	targetSet := makeKeySet(inst.targetKeys)
	items := make([]Item, 0, len(inst.items))
	for _, item := range inst.items {
		if _, exists := targetSet[item.Key]; exists {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (inst *Instance) targetItems() []Item {
	targetSet := makeKeySet(inst.targetKeys)
	items := make([]Item, 0, len(inst.targetKeys))
	for _, item := range inst.items {
		if _, exists := targetSet[item.Key]; exists {
			items = append(items, item)
		}
	}
	return items
}

func (inst *Instance) sourceVisibleItems() []Item {
	return filterItems(inst.sourceItems(), inst.sourceSearch)
}

func (inst *Instance) targetVisibleItems() []Item {
	return filterItems(inst.targetItems(), inst.targetSearch)
}

func (inst *Instance) movableKeys(direction MoveDirection) []string {
	switch direction {
	case MoveDirectionToTarget:
		return normalizeKeysForItems(inst.sourceVisibleItems(), inst.selectedSourceKeys)
	case MoveDirectionToSource:
		return normalizeKeysForItems(inst.targetVisibleItems(), inst.selectedTargetKeys)
	default:
		return nil
	}
}

func (inst *Instance) sourceListID() string { return inst.baseKey("source") }

func (inst *Instance) targetListID() string { return inst.baseKey("target") }

func (inst *Instance) sourceSearchField() string { return inst.baseKey("source-search") }

func (inst *Instance) targetSearchField() string { return inst.baseKey("target-search") }

func searchSideFromField(inst *Instance, field string) SearchSide {
	if field == inst.targetSearchField() {
		return SearchSideTarget
	}
	return SearchSideSource
}

func (inst *Instance) sourceListKey() string {
	return inst.baseKey("source-list-" + strconv.Itoa(inst.sourceListVersion))
}

func (inst *Instance) targetListKey() string {
	return inst.baseKey("target-list-" + strconv.Itoa(inst.targetListVersion))
}

func (inst *Instance) rootKey() string { return inst.baseKey("root") }

func (inst *Instance) baseKey(suffix string) string {
	base := inst.componentID
	if base == "" {
		base = inst.key
	}
	if base == "" {
		base = "transfer"
	}
	return base + "-" + suffix
}

func titleWithCount(title string, visibleCount, totalCount int) string {
	if totalCount < 0 {
		totalCount = 0
	}
	if visibleCount < 0 {
		visibleCount = 0
	}
	if visibleCount == totalCount {
		return title + " (" + strconv.Itoa(visibleCount) + ")"
	}
	return title + " (" + strconv.Itoa(visibleCount) + "/" + strconv.Itoa(totalCount) + ")"
}

func renderRows(items []Item) []string {
	if len(items) == 0 {
		return nil
	}
	rows := make([]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, renderRow(item))
	}
	return rows
}

func renderRow(item Item) string {
	title := strings.TrimSpace(item.Title)
	if title == "" {
		title = item.Key
	}
	description := strings.TrimSpace(item.Description)
	if description == "" {
		return title
	}
	return title + " - " + description
}

func filterItems(items []Item, query string) []Item {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" || len(items) == 0 {
		return items
	}
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return items
	}
	filtered := make([]Item, 0, len(items))
	for _, item := range items {
		haystack := strings.ToLower(strings.Join([]string{item.Key, item.Title, item.Description}, " "))
		matched := true
		for _, term := range terms {
			if !strings.Contains(haystack, term) {
				matched = false
				break
			}
		}
		if matched {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func selectedKeysFromIndices(items []Item, indices []int) ([]string, bool) {
	if len(indices) == 0 {
		return nil, false
	}
	keys := make([]string, 0, len(indices))
	seen := make(map[string]struct{}, len(indices))
	rejected := false
	for _, index := range indices {
		if index < 0 || index >= len(items) {
			rejected = true
			continue
		}
		item := items[index]
		if item.Disabled {
			rejected = true
			continue
		}
		if _, exists := seen[item.Key]; exists {
			rejected = true
			continue
		}
		seen[item.Key] = struct{}{}
		keys = append(keys, item.Key)
	}
	if len(keys) == 0 {
		return nil, rejected
	}
	return keys, rejected
}

func checkedIndices(items []Item, selectedKeys []string) []int {
	if len(items) == 0 || len(selectedKeys) == 0 {
		return nil
	}
	selectedSet := makeKeySet(selectedKeys)
	indices := make([]int, 0, len(selectedKeys))
	for index, item := range items {
		if _, exists := selectedSet[item.Key]; exists {
			indices = append(indices, index)
		}
	}
	if len(indices) == 0 {
		return nil
	}
	return indices
}

func normalizeKeysForItems(items []Item, keys []string) []string {
	if len(items) == 0 || len(keys) == 0 {
		return nil
	}
	allowed := make(map[string]Item, len(items))
	for _, item := range items {
		allowed[item.Key] = item
	}
	normalized := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		item, exists := allowed[key]
		if !exists || item.Disabled {
			continue
		}
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizeTargetKeys(items []Item, keys []string) []string {
	if len(items) == 0 || len(keys) == 0 {
		return nil
	}
	allowed := makeKeySet(keys)
	normalized := make([]string, 0, len(keys))
	for _, item := range items {
		if _, exists := allowed[item.Key]; exists {
			normalized = append(normalized, item.Key)
		}
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func sourceKeys(items []Item, targetKeys []string) []string {
	targetSet := makeKeySet(targetKeys)
	keys := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := targetSet[item.Key]; exists {
			continue
		}
		keys = append(keys, item.Key)
	}
	if len(keys) == 0 {
		return nil
	}
	return keys
}

func addKeys(items []Item, targetKeys, movedKeys []string) []string {
	set := makeKeySet(targetKeys)
	for _, key := range movedKeys {
		set[key] = struct{}{}
	}
	keys := make([]string, 0, len(set))
	for _, item := range items {
		if _, exists := set[item.Key]; exists {
			keys = append(keys, item.Key)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	return keys
}

func removeTargetKeys(targetKeys, movedKeys []string) []string {
	if len(targetKeys) == 0 {
		return nil
	}
	removeSet := makeKeySet(movedKeys)
	next := make([]string, 0, len(targetKeys))
	for _, key := range targetKeys {
		if _, exists := removeSet[key]; exists {
			continue
		}
		next = append(next, key)
	}
	if len(next) == 0 {
		return nil
	}
	return next
}

func removeKeys(keys, removed []string) []string {
	if len(keys) == 0 {
		return nil
	}
	removeSet := makeKeySet(removed)
	next := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, exists := removeSet[key]; exists {
			continue
		}
		next = append(next, key)
	}
	if len(next) == 0 {
		return nil
	}
	return next
}

func makeKeySet(keys []string) map[string]struct{} {
	set := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		set[key] = struct{}{}
	}
	return set
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

func getStringSliceProp(props rtui.Props, key string, def []string) []string {
	if value, ok := props[key]; ok {
		if keys, ok := value.([]string); ok {
			return append([]string(nil), keys...)
		}
	}
	return append([]string(nil), def...)
}

func getTitlesProp(props rtui.Props, def [2]string) [2]string {
	if value, ok := props[propTitles]; ok {
		if titles, ok := value.([2]string); ok {
			return titles
		}
	}
	return def
}

func getOperationsProp(props rtui.Props, def [2]string) [2]string {
	if value, ok := props[propOperations]; ok {
		if operations, ok := value.([2]string); ok {
			return operations
		}
	}
	return def
}

func getSearchPlaceholdersProp(props rtui.Props, def [2]string) [2]string {
	if value, ok := props[propSearchPlaceholders]; ok {
		if placeholders, ok := value.([2]string); ok {
			return placeholders
		}
	}
	return def
}

func getFieldIntentProp(props rtui.Props, key string) runtimeintent.FieldIntent {
	if value, ok := props[key]; ok {
		if fieldIntent, ok := value.(runtimeintent.FieldIntent); ok {
			return fieldIntent
		}
	}
	return nil
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
