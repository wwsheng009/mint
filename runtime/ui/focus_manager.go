package ui

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/internal/log"
)

// VNodeFocusManager manages keyboard focus within a VNode tree.
// It handles Tab navigation and distributes keyboard events to the focused element.
type VNodeFocusManager struct {
	focusable  []FocusableVNode // All focusable nodes in the tree
	current    int              // Index of currently focused node, -1 if none
	onNavigate func(from, to FocusableVNode) // Callback when focus changes
}

// NewVNodeFocusManager creates a new focus manager.
func NewVNodeFocusManager() *VNodeFocusManager {
	return &VNodeFocusManager{
		focusable: []FocusableVNode{},
		current:   -1,
	}
}

// SetFocusable updates the list of focusable nodes.
// It attempts to preserve focus by ID across re-renders.
// If IDs are duplicated, it preserves focus by index position.
func (m *VNodeFocusManager) SetFocusable(nodes []FocusableVNode) {
	// Save current focus ID and index
	currentID := ""
	currentIndexBefore := m.current
	if m.current >= 0 && m.current < len(m.focusable) {
		currentID = m.focusable[m.current].GetFocusID()
		log.FocusLogger.IfEnabled().Debug("SetFocusable: saving currentID=%s from index %d", currentID, m.current)
	}

	m.focusable = nodes

	// Try to restore focus by ID and index
	m.current = -1
	if currentID != "" && currentIndexBefore >= 0 {
		// First, try to preserve focus by index if the node at that index has the same ID
		// This handles the case where multiple nodes have the same ID (e.g., buttons without keys)
		if currentIndexBefore < len(m.focusable) {
			nodeID := m.focusable[currentIndexBefore].GetFocusID()
			if nodeID == currentID {
				// Same ID at same index - preserve focus by index
				m.current = currentIndexBefore
				m.focusable[m.current].SetFocus(true)
				log.FocusLogger.IfEnabled().Debug("SetFocusable: preserved focus at index %d by position", m.current)
			} else {
				// Different ID - search by ID
				for i, node := range m.focusable {
					nodeID := node.GetFocusID()
					log.FocusLogger.IfEnabled().Debug("SetFocusable: checking node %d, ID=%s", i, nodeID)
					if nodeID == currentID {
						m.current = i
						node.SetFocus(true)
						log.FocusLogger.IfEnabled().Debug("SetFocusable: restored focus to index %d by ID=%s", i, nodeID)
						break
					}
				}
			}
		} else {
			// Previous index out of range - search by ID
			for i, node := range m.focusable {
				nodeID := node.GetFocusID()
				log.FocusLogger.IfEnabled().Debug("SetFocusable: checking node %d, ID=%s", i, nodeID)
				if nodeID == currentID {
					m.current = i
					node.SetFocus(true)
					log.FocusLogger.IfEnabled().Debug("SetFocusable: restored focus to index %d by ID=%s", i, nodeID)
					break
				}
			}
		}
	}

	log.FocusLogger.IfEnabled().Debug("SetFocusable: before=%d, after=%d, nodes=%d", currentIndexBefore, m.current, len(nodes))

	// If no focus and there are focusable nodes, focus the first one
	if m.current < 0 && len(m.focusable) > 0 {
		m.FocusFirst()
	}
}

// UpdateFocusableList directly updates the focusable list without changing focus state.
// This is used internally (e.g., by the reconciler) when the list needs to be refreshed
// but the focus index should be preserved as-is.
func (m *VNodeFocusManager) UpdateFocusableList(nodes []FocusableVNode) {
		m.focusable = nodes
}

// FocusNext moves focus to the next focusable node.
// Wraps around to the first node when at the end.
func (m *VNodeFocusManager) FocusNext() bool {
	log.FocusLogger.IfEnabled().Debug("FocusNext current=%d, len(focusable)=%d", m.current, len(m.focusable))
	if len(m.focusable) == 0 {
		return false
	}

	old := m.current
	m.current = (m.current + 1) % len(m.focusable)

	log.FocusLogger.IfEnabled().Debug("FocusNext old=%d, new=%d", old, m.current)

	m.updateFocusState(old, m.current)

	if m.onNavigate != nil {
		m.onNavigate(getOrNil(m.focusable, old), m.focusable[m.current])
	}

	return true
}

// FocusPrev moves focus to the previous focusable node.
// Wraps around to the last node when at the beginning.
func (m *VNodeFocusManager) FocusPrev() bool {
	if len(m.focusable) == 0 {
		return false
	}

	old := m.current
	m.current = m.current - 1
	if m.current < 0 {
		m.current = len(m.focusable) - 1
	}

	m.updateFocusState(old, m.current)

	if m.onNavigate != nil {
		m.onNavigate(getOrNil(m.focusable, old), m.focusable[m.current])
	}

	return true
}

// FocusFirst focuses the first focusable node.
func (m *VNodeFocusManager) FocusFirst() bool {
	if len(m.focusable) == 0 {
		return false
	}

	m.current = 0
	m.focusable[0].SetFocus(true)

	if m.onNavigate != nil {
		m.onNavigate(nil, m.focusable[0])
	}

	return true
}

// FocusLast focuses the last focusable node.
func (m *VNodeFocusManager) FocusLast() bool {
	if len(m.focusable) == 0 {
		return false
	}

	m.current = len(m.focusable) - 1
	m.focusable[m.current].SetFocus(true)

	if m.onNavigate != nil {
		m.onNavigate(nil, m.focusable[m.current])
	}

	return true
}

// GetCurrent returns the currently focused node, or nil if none.
func (m *VNodeFocusManager) GetCurrent() FocusableVNode {
	if m.current < 0 || m.current >= len(m.focusable) {
		return nil
	}
	return m.focusable[m.current]
}

// HandleEvent handles keyboard events for focus navigation.
// Returns (handled, shouldRender) tuple.
func (m *VNodeFocusManager) HandleEvent(ev event.Event) (handled bool, shouldRender bool) {
	keyEvent, ok := ev.(*event.KeyEvent)
	if !ok {
		return false, false
	}

	// Debug: Log all key events
	modStr := ""
	if keyEvent.Key.Alt {
		modStr += "Alt+"
	}
	if keyEvent.Key.Ctrl {
		modStr += "Ctrl+"
	}
	if keyEvent.Key.Shift {
		modStr += "Shift+"
	}
	if keyEvent.Special != 0 { // 0 = KeyUnknown
		log.KeyLogger.Debug("SpecialKey: %s (value=%d) Modifiers: %s",
			keyEvent.Special, keyEvent.Special, modStr)
	} else if keyEvent.Key.Rune > 0 {
		log.KeyLogger.Debug("Rune: %c (0x%X) Modifiers: %s",
			keyEvent.Key.Rune, keyEvent.Key.Rune, modStr)
	}

	// Tab - navigate to next
	if keyEvent.Special == event.KeyTab {
		if keyEvent.Modifiers == event.ModShift {
			return m.FocusPrev(), true
		}
		return m.FocusNext(), true
	}

	return false, false
}

// DispatchToFocused dispatches an event to the currently focused node.
// Returns true if the event was handled.
func (m *VNodeFocusManager) DispatchToFocused(ev event.Event) bool {
	if m.current < 0 || m.current >= len(m.focusable) {
		return false
	}

	// Try to dispatch to focused component if it implements event.Component
	if component, ok := m.focusable[m.current].(interface{ HandleEvent(event.Event) bool }); ok {
		return component.HandleEvent(ev)
	}

	return false
}

// SetOnNavigate sets a callback for focus navigation changes.
func (m *VNodeFocusManager) SetOnNavigate(fn func(from, to FocusableVNode)) {
	m.onNavigate = fn
}

// Count returns the number of focusable nodes.
func (m *VNodeFocusManager) Count() int {
	return len(m.focusable)
}

// GetFocusable returns the list of focusable nodes.
func (m *VNodeFocusManager) GetFocusable() []FocusableVNode {
	return m.focusable
}

// CurrentIndex returns the index of the currently focused node.
// Returns -1 if no node is focused.
func (m *VNodeFocusManager) CurrentIndex() int {
	return m.current
}

// updateFocusState updates the focus state of nodes when focus changes.
func (m *VNodeFocusManager) updateFocusState(oldIndex, newIndex int) {
	// Remove focus from old node
	if oldIndex >= 0 && oldIndex < len(m.focusable) {
		m.focusable[oldIndex].SetFocus(false)
	}
	// Add focus to new node
	if newIndex >= 0 && newIndex < len(m.focusable) {
		m.focusable[newIndex].SetFocus(true)
	}
}

// getOrNil returns the element at index or zero value if out of bounds.
func getOrNil[T any](slice []T, index int) T {
	if index < 0 || index >= len(slice) {
		var zero T
		return zero
	}
	return slice[index]
}

// =============================================================================
// Focus Collection Utilities
// =============================================================================

// CollectFocusable collects all focusable VNodes from a VNode tree.
func CollectFocusable(root VNode) []FocusableVNode {
	return collectFocusableRecursive(root)
}

// CollectFocusableInLayer collects only focusable VNodes from a specific layer.
// When a modal is open, this ensures Tab navigation is limited to modal elements.
func CollectFocusableInLayer(root VNode, layer Layer) []FocusableVNode {
	return collectFocusableInLayerRecursive(root, layer)
}

// collectFocusableInLayerRecursive recursively collects focusable nodes in a specific layer.
// When collecting for modal layer, we need to track whether we're inside a modal subtree.
func collectFocusableInLayerRecursive(node VNode, targetLayer Layer) []FocusableVNode {
	return collectFocusableInLayerWithParent(node, targetLayer, false)
}

// collectFocusableInLayerWithParent recursively collects focusable nodes with parent layer tracking.
// insideTargetLayer is true when we're traversing inside a subtree that belongs to the target layer.
func collectFocusableInLayerWithParent(node VNode, targetLayer Layer, insideTargetLayer bool) []FocusableVNode {
	var result []FocusableVNode

	if node == nil {
		return result
	}

	// Check this node's layer
	nodeLayer := node.GetLayer()

	// Update insideTargetLayer flag
	// If we're not inside target layer but this node IS the target layer, enter it
	if !insideTargetLayer && nodeLayer == targetLayer {
		insideTargetLayer = true
	}
	// If we encounter a different non-base layer while inside target layer, exit
	if insideTargetLayer && nodeLayer != LayerBase && nodeLayer != targetLayer {
		// We've exited the target layer subtree
		return result
	}

	// For collecting modal focusables, we only include nodes that are inside a modal subtree
	if insideTargetLayer {
		// Check if current node is focusable
		if focusable, ok := node.(FocusableVNode); ok {
			if focusable.IsFocusable() {
				result = append(result, focusable)
			}
		}
	}

	// Recursively check children
	for _, child := range node.Children() {
		childFocusable := collectFocusableInLayerWithParent(child, targetLayer, insideTargetLayer)
		result = append(result, childFocusable...)
	}

	return result
}

// HasModalInTree checks if the VNode tree contains any modal layer nodes.
// This is used to determine if focus should be trapped in the modal.
func HasModalInTree(root VNode) bool {
	return hasModalInTreeRecursive(root)
}

// hasModalInTreeRecursive recursively checks for modal nodes.
func hasModalInTreeRecursive(node VNode) bool {
	if node == nil {
		return false
	}

	if node.GetLayer() == LayerModal {
		return true
	}

	for _, child := range node.Children() {
		if hasModalInTreeRecursive(child) {
			return true
		}
	}

	return false
}

// collectFocusableRecursive recursively collects focusable nodes.
func collectFocusableRecursive(node VNode) []FocusableVNode {
	var result []FocusableVNode

	if node == nil {
		return result
	}

	// Check if current node is focusable
	if focusable, ok := node.(FocusableVNode); ok {
		if focusable.IsFocusable() {
			result = append(result, focusable)
		}
	}

	// Recursively check children
	for _, child := range node.Children() {
		childFocusable := collectFocusableRecursive(child)
		result = append(result, childFocusable...)
	}

	return result
}

// FindFocusableByID finds a focusable node by its ID.
func FindFocusableByID(root VNode, id string) FocusableVNode {
	if root == nil {
		return nil
	}

	// Check current node
	if focusable, ok := root.(FocusableVNode); ok {
		if focusable.GetFocusID() == id {
			return focusable
		}
	}

	// Check children
	for _, child := range root.Children() {
		if found := FindFocusableByID(child, id); found != nil {
			return found
		}
	}

	return nil
}

// SetFocusByID sets focus to a node with the given ID.
func (m *VNodeFocusManager) SetFocusByID(id string) bool {
	for i, node := range m.focusable {
		if node.GetFocusID() == id {
			old := m.current
			m.current = i
			m.updateFocusState(old, m.current)
			if m.onNavigate != nil {
				m.onNavigate(getOrNil(m.focusable, old), node)
			}
			return true
		}
	}
	return false
}

// SetFocusByIndex sets focus to a node at the given index.
// Returns true if the index is valid, false otherwise.
func (m *VNodeFocusManager) SetFocusByIndex(index int) bool {
	if index < 0 || index >= len(m.focusable) {
		return false
	}
	old := m.current
	m.current = index
	m.updateFocusState(old, m.current)
	if m.onNavigate != nil {
		m.onNavigate(getOrNil(m.focusable, old), m.focusable[index])
	}
	return true
}

// HasFocus returns whether the given node currently has focus.
// This provides a single source of truth for focus state - components
// should call this method instead of storing their own hasFocus field.
func (m *VNodeFocusManager) HasFocus(node FocusableVNode) bool {
	if m.current < 0 || m.current >= len(m.focusable) {
		return false
	}
	return m.focusable[m.current] == node
}

// GetFocusedNode returns the currently focused node, or nil if none.
func (m *VNodeFocusManager) GetFocusedNode() FocusableVNode {
	if m.current < 0 || m.current >= len(m.focusable) {
		return nil
	}
	return m.focusable[m.current]
}

// DebugString returns a debug string representation of the focus manager.
func (m *VNodeFocusManager) DebugString() string {
	if len(m.focusable) == 0 {
		return "VNodeFocusManager{no focusable nodes}"
	}

	currentID := "none"
	if m.current >= 0 && m.current < len(m.focusable) {
		currentID = m.focusable[m.current].GetFocusID()
	}

	return fmt.Sprintf("VNodeFocusManager{count=%d, current=%d, currentID=%s}",
		len(m.focusable), m.current, currentID)
}
