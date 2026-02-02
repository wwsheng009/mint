package ui

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/event"
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
func (m *VNodeFocusManager) SetFocusable(nodes []FocusableVNode) {
	// Save current focus ID
	currentID := ""
	if m.current >= 0 && m.current < len(m.focusable) {
		currentID = m.focusable[m.current].GetFocusID()
	}

	m.focusable = nodes

	// Try to restore focus by ID
	m.current = -1
	if currentID != "" {
		for i, node := range m.focusable {
			if node.GetFocusID() == currentID {
				m.current = i
				node.SetFocus(true)
				break
			}
		}
	}

	// If no focus and there are focusable nodes, focus the first one
	if m.current < 0 && len(m.focusable) > 0 {
		m.FocusFirst()
	}
}

// FocusNext moves focus to the next focusable node.
// Wraps around to the first node when at the end.
func (m *VNodeFocusManager) FocusNext() bool {
	if len(m.focusable) == 0 {
		return false
	}

	old := m.current
	m.current = (m.current + 1) % len(m.focusable)

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
