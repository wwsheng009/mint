package treeview

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

// =============================================================================
// Helper Methods
// =============================================================================

// getVisibleNodes returns nodes that should be visible based on expand state.
func (inst *Instance) getVisibleNodes() []TreeNode {
	visible, _ := inst.visibleEntries()
	nodes := make([]TreeNode, 0, len(visible))
	for _, entry := range visible {
		nodes = append(nodes, entry.Node)
	}
	return nodes
}

// visibleEntries returns visible nodes with metadata plus a map of node index -> visible index.
func (inst *Instance) visibleEntries() ([]nodeEntry, []int) {
	if !inst.cacheDirty && inst.visibleCache != nil {
		return inst.visibleCache, inst.visibleIndexCache
	}

	entries := inst.buildNodeEntries()
	visible := make([]nodeEntry, 0, len(entries))
	visibleIndex := make([]int, len(entries))
	for i := range visibleIndex {
		visibleIndex[i] = -1
	}

	query := strings.TrimSpace(inst.searchQuery)
	filterActive := query != ""
	match := make([]bool, len(entries))
	include := make([]bool, len(entries))
	if filterActive {
		stack := make([]int, 0, 8)
		for _, entry := range entries {
			for len(stack) > 0 && entries[stack[len(stack)-1]].Depth >= entry.Depth {
				stack = stack[:len(stack)-1]
			}
			if inst.nodeMatches(entry.Node, query) {
				match[entry.Index] = true
				include[entry.Index] = true
				for _, ancestor := range stack {
					include[ancestor] = true
				}
			}
			if entry.HasChildren {
				stack = append(stack, entry.Index)
			}
		}
	}

	type stackEntry struct {
		depth    int
		expanded bool
		visible  bool
	}

	stack := make([]stackEntry, 0, 8)
	for _, entry := range entries {
		for len(stack) > 0 && stack[len(stack)-1].depth >= entry.Depth {
			stack = stack[:len(stack)-1]
		}

		isVisible := true
		if len(stack) > 0 {
			top := stack[len(stack)-1]
			isVisible = top.visible && top.expanded
		}

		if filterActive && !include[entry.Index] {
			isVisible = false
		}

		if isVisible {
			entry.Match = filterActive && match[entry.Index]
			visibleIndex[entry.Index] = len(visible)
			visible = append(visible, entry)
		}

		if entry.HasChildren {
			expanded := inst.expandState[entry.Key]
			if filterActive && include[entry.Index] {
				expanded = true
			}
			stack = append(stack, stackEntry{
				depth:    entry.Depth,
				expanded: expanded,
				visible:  isVisible,
			})
		}
	}

	inst.entryCache = entries
	inst.visibleCache = visible
	inst.visibleIndexCache = visibleIndex
	inst.cacheDirty = false
	return visible, visibleIndex
}

// buildNodeEntries prepares metadata for each node.
func (inst *Instance) buildNodeEntries() []nodeEntry {
	entries := make([]nodeEntry, len(inst.nodes))
	for i, node := range inst.nodes {
		depth := nodeDepth(node)
		hasDescendants := false
		if i+1 < len(inst.nodes) {
			nextDepth := nodeDepth(inst.nodes[i+1])
			if nextDepth > depth {
				hasDescendants = true
			}
		}
		hasChildren := hasDescendants
		if node.Lazy {
			hasChildren = true
		}
		key := nodeKey(node, i)
		if hasDescendants || !node.Lazy {
			delete(inst.lazyRequested, key)
		}
		entries[i] = nodeEntry{
			Index:          i,
			Node:           node,
			Depth:          depth,
			HasChildren:    hasChildren,
			HasDescendants: hasDescendants,
			Key:            key,
		}
	}
	return entries
}

func (inst *Instance) insertLazyChildren(parent nodeEntry, children []TreeNode, replace bool) bool {
	if parent.Index < 0 || parent.Index >= len(inst.nodes) {
		return false
	}
	if replace {
		inst.removeDescendantsAt(parent.Index)
	} else if inst.hasDescendantsAt(parent.Index) {
		return false
	}

	normalized := inst.normalizeLazyChildren(parent, children)
	insertAt := parent.Index + 1
	for insertAt < len(inst.nodes) && nodeDepth(inst.nodes[insertAt]) > parent.Depth {
		insertAt++
	}
	if len(normalized) > 0 {
		inst.nodes = append(inst.nodes[:insertAt], append(normalized, inst.nodes[insertAt:]...)...)
	}
	changed := len(normalized) > 0 || replace || inst.nodes[parent.Index].Lazy || inst.nodes[parent.Index].Loading || inst.nodes[parent.Index].LoadError != ""
	inst.nodes[parent.Index].Lazy = false
	inst.nodes[parent.Index].Loading = false
	inst.nodes[parent.Index].LoadError = ""
	if inst.lazyInsertions != nil {
		if len(normalized) == 0 {
			delete(inst.lazyInsertions, parent.Key)
		} else {
			inst.lazyInsertions[parent.Key] = normalized
		}
	}
	if changed {
		inst.invalidateCache()
	}
	return changed
}

func (inst *Instance) normalizeLazyChildren(parent nodeEntry, children []TreeNode) []TreeNode {
	if len(children) == 0 {
		return nil
	}
	baseIndent := (parent.Depth + 1) * indentStep
	parentIndent := parent.Depth * indentStep
	normalized := make([]TreeNode, 0, len(children))
	for _, child := range children {
		c := child
		if c.Indent <= parentIndent {
			c.Indent = baseIndent + c.Indent
		} else if c.Indent < baseIndent {
			c.Indent = baseIndent
		}
		if c.Path == "" {
			segment := strings.TrimSuffix(c.Content, "/")
			if segment == "" {
				segment = c.Content
			}
			if parent.Node.Path != "" {
				c.Path = parent.Node.Path + "/" + segment
			} else {
				c.Path = segment
			}
		}
		normalized = append(normalized, c)
	}
	return normalized
}

func (inst *Instance) reapplyLazyInsertions() bool {
	if len(inst.lazyInsertions) == 0 {
		return false
	}
	changed := false
	for key, children := range inst.lazyInsertions {
		parentIndex := inst.findNodeIndexByKey(key)
		if parentIndex < 0 || parentIndex >= len(inst.nodes) {
			continue
		}
		if inst.hasDescendantsAt(parentIndex) {
			continue
		}
		parent := nodeEntry{
			Index: parentIndex,
			Node:  inst.nodes[parentIndex],
			Depth: nodeDepth(inst.nodes[parentIndex]),
			Key:   key,
		}
		if inst.insertLazyChildren(parent, children, false) {
			changed = true
		}
	}
	return changed
}

func (inst *Instance) removeDescendantsAt(index int) bool {
	if index < 0 || index >= len(inst.nodes)-1 {
		return false
	}
	depth := nodeDepth(inst.nodes[index])
	end := index + 1
	for end < len(inst.nodes) && nodeDepth(inst.nodes[end]) > depth {
		end++
	}
	if end == index+1 {
		return false
	}
	inst.nodes = append(inst.nodes[:index+1], inst.nodes[end:]...)
	return true
}

func (inst *Instance) hasDescendantsAt(index int) bool {
	if index < 0 || index >= len(inst.nodes)-1 {
		return false
	}
	return nodeDepth(inst.nodes[index+1]) > nodeDepth(inst.nodes[index])
}

func (inst *Instance) findNodeIndexByKey(key string) int {
	for i, node := range inst.nodes {
		if nodeKey(node, i) == key {
			return i
		}
	}
	return -1
}

func (inst *Instance) resolveNodeIndex(nodeIndex int, path string, nodeID int) int {
	if nodeIndex >= 0 && nodeIndex < len(inst.nodes) {
		return nodeIndex
	}
	if path != "" {
		if idx := inst.findNodeIndexByPath(path); idx >= 0 {
			return idx
		}
	}
	if nodeID != 0 {
		if idx := inst.findNodeIndexByID(nodeID); idx >= 0 {
			return idx
		}
	}
	return -1
}

func (inst *Instance) applyLazyLoadResult(nodeIndex int, children []TreeNode, replace bool) bool {
	if nodeIndex < 0 || nodeIndex >= len(inst.nodes) {
		return false
	}
	parent := nodeEntry{
		Index: nodeIndex,
		Node:  inst.nodes[nodeIndex],
		Depth: nodeDepth(inst.nodes[nodeIndex]),
		Key:   nodeKey(inst.nodes[nodeIndex], nodeIndex),
	}
	if !inst.insertLazyChildren(parent, children, replace) {
		return false
	}
	inst.normalizeSelectionAndScroll()
	inst.dirty = true
	return true
}

func (inst *Instance) applyLazyLoadFailure(nodeIndex int, err string) bool {
	if nodeIndex < 0 || nodeIndex >= len(inst.nodes) {
		return false
	}
	if !inst.nodes[nodeIndex].Loading && inst.nodes[nodeIndex].LoadError == err {
		return false
	}
	inst.nodes[nodeIndex].Loading = false
	inst.nodes[nodeIndex].LoadError = err
	inst.invalidateCache()
	inst.normalizeSelectionAndScroll()
	inst.dirty = true
	return true
}

func (inst *Instance) nodeMatches(node TreeNode, query string) bool {
	if query == "" {
		return false
	}
	if inst.searchFn != nil {
		return inst.searchFn(node, query)
	}
	lower := strings.ToLower(query)
	content := strings.ToLower(node.Content)
	path := strings.ToLower(node.Path)
	if strings.Contains(content, lower) || (path != "" && strings.Contains(path, lower)) {
		return true
	}
	if node.LoadError != "" && strings.Contains(strings.ToLower(node.LoadError), lower) {
		return true
	}
	return false
}

func (inst *Instance) parentVisibleIndex(visible []nodeEntry, index int) int {
	if index <= 0 || index >= len(visible) {
		return -1
	}
	currentDepth := visible[index].Depth
	if currentDepth == 0 {
		return -1
	}
	for i := index - 1; i >= 0; i-- {
		if visible[i].Depth < currentDepth {
			return i
		}
	}
	return -1
}

func (inst *Instance) firstChildVisibleIndex(visible []nodeEntry, index int) int {
	if index < 0 || index >= len(visible)-1 {
		return -1
	}
	currentDepth := visible[index].Depth
	next := visible[index+1]
	if next.Depth > currentDepth {
		return index + 1
	}
	return -1
}

func (inst *Instance) findNodeIndexByPath(path string) int {
	if path == "" {
		return -1
	}
	for i, node := range inst.nodes {
		if node.Path == path {
			return i
		}
	}
	return -1
}

func (inst *Instance) findNodeIndexByID(id int) int {
	if id == 0 {
		return -1
	}
	for i, node := range inst.nodes {
		if node.NodeID == id {
			return i
		}
	}
	return -1
}

// isExpanded checks if a node is expanded.
func (inst *Instance) isExpanded(nodeIndex int) bool {
	if nodeIndex < 0 || nodeIndex >= len(inst.nodes) {
		return false
	}
	key := nodeKey(inst.nodes[nodeIndex], nodeIndex)
	return inst.expandState[key]
}

func (inst *Instance) calculateContentWidth(visible []nodeEntry) int {
	maxWidth := 1
	for _, entry := range visible {
		prefix, icon, content := inst.lineParts(entry)
		lineWidth := paint.StringWidth(prefix + icon + content)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	return maxWidth
}

func (inst *Instance) chromeHeight() int {
	if inst.showBorder {
		return 2 + inst.statsHeight()
	}
	return inst.statsHeight()
}

func (inst *Instance) statsHeight() int {
	if inst.showSearchStats {
		return 1
	}
	return 0
}

func (inst *Instance) searchStatsLine() string {
	query := strings.TrimSpace(inst.searchQuery)
	total, selected := inst.GetMatchStats()
	if query == "" {
		return "Search: --"
	}
	return fmt.Sprintf("Search: %q %d/%d", query, selected, total)
}

func (inst *Instance) desiredViewportHeight(visibleCount int) int {
	viewSize := inst.viewportHeight
	if viewSize < 1 {
		viewSize = visibleCount
	}
	if visibleCount > 0 && viewSize > visibleCount {
		viewSize = visibleCount
	}
	if viewSize < 1 {
		viewSize = 1
	}
	return viewSize
}

func (inst *Instance) effectiveViewportHeight(visibleCount int) int {
	viewSize := inst.desiredViewportHeight(visibleCount)
	if inst.bounds[3] > 0 {
		available := inst.bounds[3] - inst.chromeHeight()
		if available < 1 {
			available = 1
		}
		if viewSize > available {
			viewSize = available
		}
	}
	if visibleCount > 0 && viewSize > visibleCount {
		viewSize = visibleCount
	}
	if viewSize < 1 {
		viewSize = 1
	}
	return viewSize
}

func (inst *Instance) visibleViewport(visibleCount int) scrollutil.VerticalViewport {
	viewSize := inst.effectiveViewportHeight(visibleCount)
	return scrollutil.NewVerticalViewport(visibleCount, viewSize, inst.scrollOffset)
}

func (inst *Instance) normalizeSelectionAndScroll() {
	visible, _ := inst.visibleEntries()
	visibleCount := len(visible)
	if visibleCount == 0 {
		inst.selectedIndex = -1
		inst.scrollOffset = 0
		return
	}

	query := strings.TrimSpace(inst.searchQuery)
	if query == "" {
		inst.autoSelectMatch = false
	}
	if query != "" && !inst.selectedIndexControlled {
		firstMatch := -1
		for i, entry := range visible {
			if entry.Match {
				firstMatch = i
				break
			}
		}
		if firstMatch >= 0 {
			if inst.autoSelectMatch || inst.selectedIndex < 0 || inst.selectedIndex >= visibleCount || !visible[inst.selectedIndex].Match {
				inst.selectedIndex = firstMatch
			}
		} else if inst.autoSelectMatch {
			inst.selectedIndex = -1
		}
		inst.autoSelectMatch = false
	}
	if inst.selectedIndex >= visibleCount {
		inst.selectedIndex = visibleCount - 1
	}
	if inst.selectedIndex < -1 {
		inst.selectedIndex = -1
	}
	viewport := inst.visibleViewport(visibleCount)
	if inst.selectedIndex >= 0 {
		if viewport.EnsureVisible(inst.selectedIndex) {
			inst.scrollOffset = viewport.Offset
		}
	} else {
		inst.scrollOffset = viewport.Offset
	}

	inst.updateSearchStats()
}

func (inst *Instance) normalizeCheckedKeys() {
	if inst.selectionMode == SelectionNone || len(inst.nodes) == 0 {
		inst.checkedKeys = nil
		inst.selectionAnchorKey = ""
		return
	}
	normalized := make(map[string]bool)
	for i, node := range inst.nodes {
		key := nodeKey(node, i)
		if inst.checkedKeys != nil && inst.checkedKeys[key] {
			normalized[key] = true
		}
	}
	if inst.selectionMode == SelectionSingle && len(normalized) > 1 {
		trimmed := make(map[string]bool, 1)
		for i, node := range inst.nodes {
			key := nodeKey(node, i)
			if normalized[key] {
				trimmed[key] = true
				break
			}
		}
		normalized = trimmed
	}
	inst.checkedKeys = normalized
	if inst.selectionAnchorKey != "" && inst.findNodeIndexByKey(inst.selectionAnchorKey) < 0 {
		inst.selectionAnchorKey = ""
	}
}

func (inst *Instance) emitCheckedSelectionChanged() {
	if inst.intentEmitter == nil {
		return
	}

	keys, indices, paths, nodeIDs := inst.checkedSnapshot()
	if inst.componentID != "" {
		inst.emitOptionalGlobalIntent(SelectionChange(inst.componentID, inst.selectionMode, keys, indices, paths, nodeIDs))
	}

	if inst.selectionIntentField != nil {
		value := strings.Join(paths, ",")
		if value == "" {
			value = strings.Join(keys, ",")
		}
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.selectionIntentField.GetField(),
			Value: value,
		})
		return
	}
	if inst.selectionIntent != nil {
		inst.intentEmitter(inst.selectionIntent)
	}
}

func (inst *Instance) checkedSnapshot() (keys []string, indices []int, paths []string, nodeIDs []int) {
	if inst.checkedKeys == nil {
		return nil, nil, nil, nil
	}
	keys = make([]string, 0, len(inst.checkedKeys))
	indices = make([]int, 0, len(inst.checkedKeys))
	paths = make([]string, 0, len(inst.checkedKeys))
	nodeIDs = make([]int, 0, len(inst.checkedKeys))
	for i, node := range inst.nodes {
		key := nodeKey(node, i)
		if inst.checkedKeys[key] {
			keys = append(keys, key)
			indices = append(indices, i)
			if node.Path != "" {
				paths = append(paths, node.Path)
			}
			if node.NodeID != 0 {
				nodeIDs = append(nodeIDs, node.NodeID)
			}
		}
	}
	return keys, indices, paths, nodeIDs
}

func (inst *Instance) selectVisibleIndex(target int, visible []nodeEntry, emitSelect bool) bool {
	if len(visible) == 0 {
		inst.selectedIndex = -1
		inst.scrollOffset = 0
		return false
	}
	if target < 0 {
		target = 0
	}
	if target >= len(visible) {
		target = len(visible) - 1
	}

	viewport := inst.visibleViewport(len(visible))
	oldScroll := inst.scrollOffset
	changed := inst.selectedIndex != target
	inst.selectedIndex = target
	if viewport.EnsureVisible(target) {
		if inst.scrollOffset != viewport.Offset {
			changed = true
		}
		inst.scrollOffset = viewport.Offset
	} else if inst.scrollOffset != viewport.Offset {
		changed = true
		inst.scrollOffset = viewport.Offset
	}

	if changed {
		inst.dirty = true
		if emitSelect {
			inst.emitNodeSelect(visible[target].Index)
		}
		if inst.scrollOffset != oldScroll {
			inst.emitScroll(inst.scrollOffset-oldScroll, viewport.ViewSize, len(visible))
		}
	}
	return changed
}

func (inst *Instance) scrollBy(delta int) bool {
	visible, _ := inst.visibleEntries()
	viewport := inst.visibleViewport(len(visible))
	oldScroll := inst.scrollOffset
	if !viewport.ScrollBy(delta) {
		return false
	}
	inst.scrollOffset = viewport.Offset
	inst.dirty = true
	inst.emitScroll(inst.scrollOffset-oldScroll, viewport.ViewSize, len(visible))
	return true
}

func (inst *Instance) toggleChecked(entry nodeEntry) bool {
	if inst.selectionMode == SelectionNone {
		return false
	}
	changed := false
	switch inst.selectionMode {
	case SelectionSingle:
		changed = !inst.isChecked(entry) || len(inst.checkedKeys) != 1
		inst.checkedKeys = map[string]bool{entry.Key: true}
	case SelectionMultiple:
		if inst.checkedKeys == nil {
			inst.checkedKeys = make(map[string]bool)
		}
		targetChecked := !inst.checkedKeys[entry.Key]
		changed = inst.setCheckedSubtree(entry.Index, targetChecked)
	}

	if changed {
		inst.selectionAnchorKey = entry.Key
		inst.normalizeCheckedKeys()
		inst.dirty = true
		inst.emitCheckedSelectionChanged()
	}
	return changed
}

func (inst *Instance) setCheckedSubtree(index int, checked bool) bool {
	if inst.selectionMode != SelectionMultiple {
		return false
	}
	if index < 0 || index >= len(inst.nodes) {
		return false
	}
	if inst.checkedKeys == nil {
		inst.checkedKeys = make(map[string]bool)
	}
	changed := false
	depth := nodeDepth(inst.nodes[index])
	for i := index; i < len(inst.nodes); i++ {
		if i > index && nodeDepth(inst.nodes[i]) <= depth {
			break
		}
		key := nodeKey(inst.nodes[i], i)
		if checked {
			if !inst.checkedKeys[key] {
				inst.checkedKeys[key] = true
				changed = true
			}
			continue
		}
		if inst.checkedKeys[key] {
			delete(inst.checkedKeys, key)
			changed = true
		}
	}
	return changed
}

func (inst *Instance) toggleCheckedSelection(act *action.Action) bool {
	if inst.selectionMode == SelectionNone {
		return false
	}
	if _, matchesOnly, ok := inst.selectionInvertFromAction(act); ok {
		return inst.invertCheckedSelection(matchesOnly)
	}
	visible, visibleIndex := inst.visibleEntries()
	entry, ok := inst.entryFromAction(act, visible, visibleIndex)
	if !ok {
		if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
			if len(visible) == 0 {
				return false
			}
			entry = visible[0]
		} else {
			entry = visible[inst.selectedIndex]
		}
	}
	return inst.toggleChecked(entry)
}

func (inst *Instance) deselectByAction(act *action.Action) bool {
	if inst.selectionMode == SelectionNone {
		return false
	}
	visible, visibleIndex := inst.visibleEntries()
	entry, ok := inst.entryFromAction(act, visible, visibleIndex)
	if !ok {
		if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
			return false
		}
		entry = visible[inst.selectedIndex]
	}
	if inst.checkedKeys == nil || !inst.checkedKeys[entry.Key] {
		return false
	}
	delete(inst.checkedKeys, entry.Key)
	inst.selectionAnchorKey = entry.Key
	inst.normalizeCheckedKeys()
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) selectRangeByAction(act *action.Action) bool {
	if inst.selectionMode != SelectionMultiple {
		return false
	}
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	start, end, ok := inst.rangeFromAction(act, len(visible))
	if !ok {
		anchor, anchorOK := inst.selectionAnchorVisibleIndex(visible)
		if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
			return false
		}
		if !anchorOK {
			anchor = inst.selectedIndex
			inst.selectionAnchorKey = visible[anchor].Key
		}
		start, end, ok = anchor, inst.selectedIndex, true
	}
	if start > end {
		start, end = end, start
	}
	if inst.checkedKeys == nil {
		inst.checkedKeys = make(map[string]bool)
	}
	for i := start; i <= end; i++ {
		inst.checkedKeys[visible[i].Key] = true
	}
	inst.selectionAnchorKey = visible[start].Key
	inst.normalizeCheckedKeys()
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) selectAllByAction(act *action.Action) bool {
	if inst.selectionMode != SelectionMultiple {
		return false
	}
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	filterActive := inst.selectionScopeMatchesOnly(act, strings.TrimSpace(inst.searchQuery) != "")
	checked := make(map[string]bool, len(visible))
	for _, entry := range visible {
		if filterActive && !entry.Match {
			continue
		}
		checked[entry.Key] = true
	}
	inst.checkedKeys = checked
	if inst.selectedIndex >= 0 && inst.selectedIndex < len(visible) {
		inst.selectionAnchorKey = visible[inst.selectedIndex].Key
	}
	inst.normalizeCheckedKeys()
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) clearCheckedSelectionByAction(act *action.Action) bool {
	if len(inst.checkedKeys) == 0 {
		return false
	}
	if act != nil {
		if scoped, matchesOnly, ok := inst.selectionClearScopeFromAction(act); ok {
			if scoped {
				return inst.clearCheckedScope(matchesOnly)
			}
		}
	}
	inst.checkedKeys = nil
	inst.selectionAnchorKey = ""
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) clearCheckedScope(matchesOnly bool) bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	changed := false
	for _, entry := range visible {
		if matchesOnly && !entry.Match {
			continue
		}
		if inst.checkedKeys[entry.Key] {
			delete(inst.checkedKeys, entry.Key)
			changed = true
		}
	}
	if !changed {
		return false
	}
	inst.normalizeCheckedKeys()
	if len(inst.checkedKeys) == 0 {
		inst.selectionAnchorKey = ""
	}
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) invertCheckedSelection(matchesOnly bool) bool {
	if inst.selectionMode != SelectionMultiple {
		return false
	}
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	if inst.checkedKeys == nil {
		inst.checkedKeys = make(map[string]bool)
	}
	changed := false
	for _, entry := range visible {
		if matchesOnly && !entry.Match {
			continue
		}
		if inst.checkedKeys[entry.Key] {
			delete(inst.checkedKeys, entry.Key)
		} else {
			inst.checkedKeys[entry.Key] = true
		}
		changed = true
	}
	if !changed {
		return false
	}
	if inst.selectedIndex >= 0 && inst.selectedIndex < len(visible) {
		inst.selectionAnchorKey = visible[inst.selectedIndex].Key
	}
	inst.normalizeCheckedKeys()
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) selectionAnchorVisibleIndex(visible []nodeEntry) (int, bool) {
	if inst.selectionAnchorKey == "" {
		return -1, false
	}
	for i, entry := range visible {
		if entry.Key == inst.selectionAnchorKey {
			return i, true
		}
	}
	return -1, false
}

func (inst *Instance) selectionInvertFromAction(act *action.Action) (invert bool, matchesOnly bool, ok bool) {
	if act == nil {
		return false, false, false
	}
	def := strings.TrimSpace(inst.searchQuery) != ""
	switch payload := act.Payload.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(payload)) {
		case "invert":
			return true, def, true
		case "invert_visible":
			return true, false, true
		case "invert_matches":
			return true, true, true
		}
	case map[string]string:
		mode := strings.ToLower(strings.TrimSpace(payload["mode"]))
		if mode == "" {
			mode = strings.ToLower(strings.TrimSpace(payload["op"]))
		}
		if mode == "invert" {
			return true, inst.selectionScopeMatchesOnly(act, def), true
		}
	case map[string]interface{}:
		mode := ""
		if raw, exists := payload["mode"]; exists {
			if value, ok := raw.(string); ok {
				mode = strings.ToLower(strings.TrimSpace(value))
			}
		}
		if mode == "" {
			if raw, exists := payload["op"]; exists {
				if value, ok := raw.(string); ok {
					mode = strings.ToLower(strings.TrimSpace(value))
				}
			}
		}
		if mode == "invert" {
			return true, inst.selectionScopeMatchesOnly(act, def), true
		}
	}
	return false, false, false
}

func (inst *Instance) selectionScopeMatchesOnly(act *action.Action, def bool) bool {
	if act == nil {
		return def
	}
	switch payload := act.Payload.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(payload)) {
		case "visible", "all", "all_visible", "invert_visible", "clear_visible":
			return false
		case "matches", "matched", "search", "search_matches", "invert_matches", "clear_matches":
			return true
		}
	case map[string]string:
		if scope := strings.ToLower(strings.TrimSpace(payload["scope"])); scope != "" {
			return scope == "matches" || scope == "matched" || scope == "search"
		}
	case map[string]interface{}:
		if raw, exists := payload["scope"]; exists {
			if scope, ok := raw.(string); ok {
				scope = strings.ToLower(strings.TrimSpace(scope))
				return scope == "matches" || scope == "matched" || scope == "search"
			}
		}
	}
	return def
}

func (inst *Instance) selectionClearScopeFromAction(act *action.Action) (scoped bool, matchesOnly bool, ok bool) {
	if act == nil {
		return false, false, false
	}
	switch payload := act.Payload.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(payload)) {
		case "visible", "clear_visible":
			return true, false, true
		case "matches", "matched", "clear_matches":
			return true, true, true
		}
	case map[string]string:
		mode := strings.ToLower(strings.TrimSpace(payload["mode"]))
		if mode == "" {
			mode = strings.ToLower(strings.TrimSpace(payload["op"]))
		}
		if mode == "clear" {
			return true, inst.selectionScopeMatchesOnly(act, false), true
		}
	case map[string]interface{}:
		mode := ""
		if raw, exists := payload["mode"]; exists {
			if value, ok := raw.(string); ok {
				mode = strings.ToLower(strings.TrimSpace(value))
			}
		}
		if mode == "" {
			if raw, exists := payload["op"]; exists {
				if value, ok := raw.(string); ok {
					mode = strings.ToLower(strings.TrimSpace(value))
				}
			}
		}
		if mode == "clear" {
			return true, inst.selectionScopeMatchesOnly(act, false), true
		}
	}
	return false, false, false
}

