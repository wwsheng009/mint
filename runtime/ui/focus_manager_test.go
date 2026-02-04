package ui

import (
	"testing"

	"github.com/wwsheng009/mint/framework/event"
)

// =============================================================================
// Mock FocusableVNode for Testing
// =============================================================================

// mockFocusableNode is a test implementation of FocusableVNode
type mockFocusableNode struct {
	*ElementVNode
	id          string
	label       string
	isFocusable bool
	hasFocus    bool
}

func newMockFocusable(id, label string) *mockFocusableNode {
	return &mockFocusableNode{
		ElementVNode: NewElement("mock"),
		id:           id,
		label:        label,
		isFocusable:  true,
		hasFocus:     false,
	}
}

func (m *mockFocusableNode) SetFocus(hasFocus bool) {
	m.hasFocus = hasFocus
}

func (m *mockFocusableNode) IsFocusable() bool {
	return m.isFocusable
}

func (m *mockFocusableNode) GetFocusID() string {
	return m.id
}

func (m *mockFocusableNode) Label() string {
	return m.label
}

func (m *mockFocusableNode) HasFocus() bool {
	return m.hasFocus
}

// =============================================================================
// Focus Manager Edge Case Tests
// =============================================================================

// TestVNodeFocusManager_EmptyList tests navigation with empty focusable list
func TestVNodeFocusManager_EmptyList(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T, *VNodeFocusManager)
	}{
		{
			name: "FocusNext on empty list",
			fn: func(t *testing.T, m *VNodeFocusManager) {
				if got := m.FocusNext(); got {
					t.Error("FocusNext() on empty list should return false")
				}
				if m.GetCurrent() != nil {
					t.Error("GetCurrent() should return nil for empty list")
				}
			},
		},
		{
			name: "FocusPrev on empty list",
			fn: func(t *testing.T, m *VNodeFocusManager) {
				if got := m.FocusPrev(); got {
					t.Error("FocusPrev() on empty list should return false")
				}
			},
		},
		{
			name: "FocusFirst on empty list",
			fn: func(t *testing.T, m *VNodeFocusManager) {
				if got := m.FocusFirst(); got {
					t.Error("FocusFirst() on empty list should return false")
				}
			},
		},
		{
			name: "FocusLast on empty list",
			fn: func(t *testing.T, m *VNodeFocusManager) {
				if got := m.FocusLast(); got {
					t.Error("FocusLast() on empty list should return false")
				}
			},
		},
		{
			name: "SetFocusByIndex on empty list",
			fn: func(t *testing.T, m *VNodeFocusManager) {
				if got := m.SetFocusByIndex(0); got {
					t.Error("SetFocusByIndex(0) on empty list should return false")
				}
			},
		},
		{
			name: "SetFocusByID on empty list",
			fn: func(t *testing.T, m *VNodeFocusManager) {
				if got := m.SetFocusByID("any-id"); got {
					t.Error("SetFocusByID() on empty list should return false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewVNodeFocusManager()
			tt.fn(t, m)
		})
	}
}

// TestVNodeFocusManager_SingleElement tests navigation with single focusable element
func TestVNodeFocusManager_SingleElement(t *testing.T) {
	m := NewVNodeFocusManager()
	node := newMockFocusable("btn1", "Button 1")

	m.SetFocusable([]FocusableVNode{node})

	// Should auto-focus first element
	if m.CurrentIndex() != 0 {
		t.Errorf("CurrentIndex() = %d, want 0", m.CurrentIndex())
	}
	if !node.HasFocus() {
		t.Error("First element should be auto-focused")
	}

	// FocusNext should wrap to same element
	if !m.FocusNext() {
		t.Error("FocusNext() should return true")
	}
	if m.CurrentIndex() != 0 {
		t.Errorf("After FocusNext, CurrentIndex() = %d, want 0", m.CurrentIndex())
	}

	// FocusPrev should also wrap to same element
	if !m.FocusPrev() {
		t.Error("FocusPrev() should return true")
	}
	if m.CurrentIndex() != 0 {
		t.Errorf("After FocusPrev, CurrentIndex() = %d, want 0", m.CurrentIndex())
	}
}

// TestVNodeFocusManager_TabNavigation tests Tab key navigation
func TestVNodeFocusManager_TabNavigation(t *testing.T) {
	m := NewVNodeFocusManager()
	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
		newMockFocusable("btn3", "Button 3"),
	}

	m.SetFocusable(nodes)

	// Initial focus on first
	if m.CurrentIndex() != 0 {
		t.Errorf("Initial CurrentIndex() = %d, want 0", m.CurrentIndex())
	}

	// Tab moves forward
	tabEvent := &event.KeyEvent{Special: event.KeyTab, Modifiers: 0}
	handled, _ := m.HandleEvent(tabEvent)
	if !handled {
		t.Error("Tab key should be handled")
	}
	if m.CurrentIndex() != 1 {
		t.Errorf("After Tab, CurrentIndex() = %d, want 1", m.CurrentIndex())
	}

	// Tab wraps around
	tabEvent2 := &event.KeyEvent{Special: event.KeyTab, Modifiers: 0}
	m.HandleEvent(tabEvent2) // Move to index 2
	tabEvent3 := &event.KeyEvent{Special: event.KeyTab, Modifiers: 0}
	m.HandleEvent(tabEvent3) // Wrap to 0
	if m.CurrentIndex() != 0 {
		t.Errorf("After wrapping Tab, CurrentIndex() = %d, want 0", m.CurrentIndex())
	}

	// Shift+Tab moves backward
	shiftTabEvent := &event.KeyEvent{Special: event.KeyTab, Modifiers: event.ModShift}
	handled, _ = m.HandleEvent(shiftTabEvent)
	if !handled {
		t.Error("Shift+Tab key should be handled")
	}
	if m.CurrentIndex() != 2 {
		t.Errorf("After Shift+Tab, CurrentIndex() = %d, want 2", m.CurrentIndex())
	}
}

// TestVNodeFocusManager_FocusPreservation_SameID tests focus preservation with same ID
func TestVNodeFocusManager_FocusPreservation_SameID(t *testing.T) {
	m := NewVNodeFocusManager()

	// Initial list with focus on second element
	nodes1 := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
		newMockFocusable("btn3", "Button 3"),
	}
	m.SetFocusable(nodes1)
	m.SetFocusByIndex(1) // Focus btn2

	if m.CurrentIndex() != 1 {
		t.Errorf("Before update, CurrentIndex() = %d, want 1", m.CurrentIndex())
	}

	// Re-render with same IDs at same positions
	nodes2 := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"), // Same ID at same index
		newMockFocusable("btn3", "Button 3"),
	}
	m.SetFocusable(nodes2)

	// Focus should be preserved at same index
	if m.CurrentIndex() != 1 {
		t.Errorf("After update with same ID, CurrentIndex() = %d, want 1", m.CurrentIndex())
	}
	if !nodes2[1].(*mockFocusableNode).HasFocus() {
		t.Error("Second element should still have focus")
	}
}

// TestVNodeFocusManager_FocusPreservation_DifferentIndex tests focus preservation when ID moves
func TestVNodeFocusManager_FocusPreservation_DifferentIndex(t *testing.T) {
	m := NewVNodeFocusManager()

	// Initial list with focus on second element
	nodes1 := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
		newMockFocusable("btn3", "Button 3"),
	}
	m.SetFocusable(nodes1)
	m.SetFocusByIndex(1) // Focus btn2
	focusedID := nodes1[1].GetFocusID()

	// Re-render with btn2 moved to different position
	nodes2 := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn3", "Button 3"),
		newMockFocusable("btn2", "Button 2"), // btn2 moved to index 2
	}
	m.SetFocusable(nodes2)

	// Focus should follow the ID
	if m.CurrentIndex() != 2 {
		t.Errorf("After ID moved, CurrentIndex() = %d, want 2", m.CurrentIndex())
	}
	if nodes2[m.CurrentIndex()].GetFocusID() != focusedID {
		t.Errorf("Focused ID = %s, want %s", nodes2[m.CurrentIndex()].GetFocusID(), focusedID)
	}
}

// TestVNodeFocusManager_FocusPreservation_DuplicateIDs tests behavior with duplicate IDs
func TestVNodeFocusManager_FocusPreservation_DuplicateIDs(t *testing.T) {
	m := NewVNodeFocusManager()

	// List with duplicate IDs (e.g., multiple buttons without keys)
	nodes1 := []FocusableVNode{
		newMockFocusable("button", "Button 1"),
		newMockFocusable("button", "Button 2"),
		newMockFocusable("button", "Button 3"),
	}
	m.SetFocusable(nodes1)

	// Focus on middle element
	m.SetFocusByIndex(1)

	// Re-render with same duplicate IDs - should preserve by position
	nodes2 := []FocusableVNode{
		newMockFocusable("button", "Button 1"),
		newMockFocusable("button", "Button 2"),
		newMockFocusable("button", "Button 3"),
	}
	m.SetFocusable(nodes2)

	// Should preserve by index position since IDs are the same
	if m.CurrentIndex() != 1 {
		t.Errorf("With duplicate IDs, CurrentIndex() = %d, want 1", m.CurrentIndex())
	}
}

// TestVNodeFocusManager_FocusPreservation_IDNotFound tests when focused ID is removed
func TestVNodeFocusManager_FocusPreservation_IDNotFound(t *testing.T) {
	m := NewVNodeFocusManager()

	// Initial list
	nodes1 := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
		newMockFocusable("btn3", "Button 3"),
	}
	m.SetFocusable(nodes1)
	m.SetFocusByIndex(1) // Focus btn2

	// Re-render with btn2 removed
	nodes2 := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn3", "Button 3"),
	}
	m.SetFocusable(nodes2)

	// Focus should reset to first element (since focused ID not found)
	if m.CurrentIndex() != 0 {
		t.Errorf("When focused ID removed, CurrentIndex() = %d, want 0", m.CurrentIndex())
	}
	if !nodes2[0].(*mockFocusableNode).HasFocus() {
		t.Error("First element should be focused when previous focus is removed")
	}
}

// TestVNodeFocusManager_FocusPreservation_EmptyToNonEmpty tests transition from empty to non-empty
func TestVNodeFocusManager_FocusPreservation_EmptyToNonEmpty(t *testing.T) {
	m := NewVNodeFocusManager()

	// Start with empty list
	m.SetFocusable([]FocusableVNode{})
	if m.CurrentIndex() != -1 {
		t.Errorf("With empty list, CurrentIndex() = %d, want -1", m.CurrentIndex())
	}

	// Add elements
	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
	}
	m.SetFocusable(nodes)

	// Should auto-focus first element
	if m.CurrentIndex() != 0 {
		t.Errorf("After adding elements, CurrentIndex() = %d, want 0", m.CurrentIndex())
	}
}

// TestVNodeFocusManager_FocusPreservation_NonEmptyToEmpty tests transition from non-empty to empty
func TestVNodeFocusManager_FocusPreservation_NonEmptyToEmpty(t *testing.T) {
	m := NewVNodeFocusManager()

	// Start with elements
	nodes1 := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
	}
	m.SetFocusable(nodes1)
	m.SetFocusByIndex(1)

	// Clear all elements
	m.SetFocusable([]FocusableVNode{})

	if m.CurrentIndex() != -1 {
		t.Errorf("After clearing, CurrentIndex() = %d, want -1", m.CurrentIndex())
	}
}

// TestVNodeFocusManager_UnfocusableNode tests handling of nodes that become unfocusable
func TestVNodeFocusManager_UnfocusableNode(t *testing.T) {
	m := NewVNodeFocusManager()

	// Create nodes where second becomes unfocusable
	btn1 := newMockFocusable("btn1", "Button 1")
	btn2 := newMockFocusable("btn2", "Button 2")
	btn3 := newMockFocusable("btn3", "Button 3")

	// Make middle node unfocusable
	btn2.isFocusable = false

	nodes := []FocusableVNode{btn1, btn2, btn3}
	m.SetFocusable(nodes)

	// Should skip unfocusable node and focus first
	if m.CurrentIndex() != 0 {
		t.Errorf("With unfocusable node, CurrentIndex() = %d, want 0", m.CurrentIndex())
	}

	// CollectFocusable should filter out unfocusable nodes
	root := VStack(btn1, btn2, btn3)
	focusable := CollectFocusable(root)
	if len(focusable) != 2 {
		t.Errorf("CollectFocusable() should filter unfocusable, got %d, want 2", len(focusable))
	}
}

// TestVNodeFocusManager_SetFocusByIndex_OutOfRange tests SetFocusByIndex with invalid indices
func TestVNodeFocusManager_SetFocusByIndex_OutOfRange(t *testing.T) {
	m := NewVNodeFocusManager()
	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
	}
	m.SetFocusable(nodes)

	tests := []struct {
		name  string
		index int
		want  bool
	}{
		{"negative index", -1, false},
		{"index at count", 2, false},
		{"index beyond count", 5, false},
		{"valid index 0", 0, true},
		{"valid index 1", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.SetFocusByIndex(tt.index); got != tt.want {
				t.Errorf("SetFocusByIndex(%d) = %v, want %v", tt.index, got, tt.want)
			}
		})
	}
}

// TestVNodeFocusManager_SetFocusByID tests SetFocusByID functionality
func TestVNodeFocusManager_SetFocusByID(t *testing.T) {
	m := NewVNodeFocusManager()
	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
		newMockFocusable("btn3", "Button 3"),
	}
	m.SetFocusable(nodes)

	// Find and focus by ID
	if !m.SetFocusByID("btn2") {
		t.Error("SetFocusByID(btn2) should return true")
	}
	if m.CurrentIndex() != 1 {
		t.Errorf("After SetFocusByID(btn2), CurrentIndex() = %d, want 1", m.CurrentIndex())
	}

	// Non-existent ID
	if m.SetFocusByID("nonexistent") {
		t.Error("SetFocusByID(nonexistent) should return false")
	}
	// Current index should not change
	if m.CurrentIndex() != 1 {
		t.Errorf("After failed SetFocusByID, CurrentIndex() = %d, want 1", m.CurrentIndex())
	}

	// Empty ID
	if m.SetFocusByID("") {
		t.Error("SetFocusByID(empty) should return false")
	}
}

// TestVNodeFocusManager_HasFocus tests HasFocus method
func TestVNodeFocusManager_HasFocus(t *testing.T) {
	m := NewVNodeFocusManager()
	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
	}
	m.SetFocusable(nodes)

	// First node has focus
	if !m.HasFocus(nodes[0]) {
		t.Error("HasFocus(first node) should return true")
	}
	if m.HasFocus(nodes[1]) {
		t.Error("HasFocus(second node) should return false")
	}

	// Focus second node
	m.SetFocusByIndex(1)

	if m.HasFocus(nodes[0]) {
		t.Error("After change, HasFocus(first node) should return false")
	}
	if !m.HasFocus(nodes[1]) {
		t.Error("After change, HasFocus(second node) should return true")
	}

	// Nil node
	if m.HasFocus(nil) {
		t.Error("HasFocus(nil) should return false")
	}
}

// TestVNodeFocusManager_OnNavigateCallback tests the onNavigate callback
func TestVNodeFocusManager_OnNavigateCallback(t *testing.T) {
	m := NewVNodeFocusManager()
	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
		newMockFocusable("btn3", "Button 3"),
	}
	m.SetFocusable(nodes)

	var from, to FocusableVNode
	callbackCalled := false

	m.SetOnNavigate(func(f, t FocusableVNode) {
		callbackCalled = true
		from = f
		to = t
	})

	// Navigate from index 0 to 1
	m.FocusNext()

	if !callbackCalled {
		t.Error("onNavigate callback should be called")
	}
	if from != nodes[0] {
		t.Error("from should be btn1")
	}
	if to != nodes[1] {
		t.Error("to should be btn2")
	}
}

// TestVNodeFocusManager_UpdateFocusableList tests UpdateFocusableList (doesn't change focus)
func TestVNodeFocusManager_UpdateFocusableList(t *testing.T) {
	m := NewVNodeFocusManager()
	nodes1 := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
	}
	m.SetFocusable(nodes1)
	m.SetFocusByIndex(1) // Focus btn2

	// UpdateFocusableList doesn't try to preserve focus
	nodes2 := []FocusableVNode{
		newMockFocusable("btn3", "Button 3"),
		newMockFocusable("btn4", "Button 4"),
	}
	m.UpdateFocusableList(nodes2)

	// Index is preserved but may be out of bounds
	if m.CurrentIndex() != 1 {
		t.Errorf("UpdateFocusableList preserves index, got %d, want 1", m.CurrentIndex())
	}

	// Note: This is different from SetFocusable which tries to restore focus
}

// TestVNodeFocusManager_DispatchToFocused tests event dispatching to focused node
func TestVNodeFocusManager_DispatchToFocused(t *testing.T) {
	m := NewVNodeFocusManager()

	// Create nodes with HandleEvent
	btn1Handled := false
	btn2Handled := false

	nodes := []FocusableVNode{
		&mockFocusableWithEvents{
			mockFocusableNode: newMockFocusable("btn1", "Button 1"),
			handleEvent: func(ev event.Event) bool {
				btn1Handled = true
				return true
			},
		},
		&mockFocusableWithEvents{
			mockFocusableNode: newMockFocusable("btn2", "Button 2"),
			handleEvent: func(ev event.Event) bool {
				btn2Handled = true
				return true
			},
		},
	}
	m.SetFocusable(nodes)

	// Focus second node (index 1)
	m.SetFocusByIndex(1)

	// Dispatch event to focused node (btn2)
	testEvent := &event.KeyEvent{Key: event.Key{Rune: 'a'}}
	handled := m.DispatchToFocused(testEvent)

	if !handled {
		t.Error("DispatchToFocused should return true when event is handled")
	}
	if !btn2Handled {
		t.Error("Focused node's (btn2) HandleEvent should be called")
	}
	if btn1Handled {
		t.Error("Non-focused node's (btn1) HandleEvent should NOT be called")
	}
}

// mockFocusableWithEvents is a mock that implements HandleEvent
type mockFocusableWithEvents struct {
	*mockFocusableNode
	handleEvent func(event.Event) bool
}

func (m *mockFocusableWithEvents) HandleEvent(ev event.Event) bool {
	if m.handleEvent != nil {
		return m.handleEvent(ev)
	}
	return false
}

// TestVNodeFocusManager_NonKeyEvent tests that non-key events are not handled
func TestVNodeFocusManager_NonKeyEvent(t *testing.T) {
	m := NewVNodeFocusManager()
	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
	}
	m.SetFocusable(nodes)

	// Create a non-key event (using a struct that doesn't match KeyEvent)
	nonKeyEvent := &struct{ event.Event }{}

	handled, shouldRender := m.HandleEvent(nonKeyEvent)
	if handled {
		t.Error("Non-key event should not be handled")
	}
	if shouldRender {
		t.Error("Non-key event should not trigger render")
	}
}

// TestVNodeFocusManager_GetCurrent tests GetCurrent method
func TestVNodeFocusManager_GetCurrent(t *testing.T) {
	m := NewVNodeFocusManager()

	// Empty list
	if m.GetCurrent() != nil {
		t.Error("GetCurrent() on empty list should return nil")
	}

	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
	}
	m.SetFocusable(nodes)

	// Should return first node (auto-focused)
	current := m.GetCurrent()
	if current == nil {
		t.Fatal("GetCurrent() should return first node")
	}
	if current.GetFocusID() != "btn1" {
		t.Errorf("GetCurrent().GetFocusID() = %s, want btn1", current.GetFocusID())
	}

	// After navigating
	m.FocusNext()
	current = m.GetCurrent()
	if current.GetFocusID() != "btn2" {
		t.Errorf("GetCurrent().GetFocusID() after FocusNext = %s, want btn2", current.GetFocusID())
	}
}

// TestVNodeFocusManager_Count tests Count method
func TestVNodeFocusManager_Count(t *testing.T) {
	m := NewVNodeFocusManager()

	if m.Count() != 0 {
		t.Errorf("Count() on empty list = %d, want 0", m.Count())
	}

	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
		newMockFocusable("btn3", "Button 3"),
	}
	m.SetFocusable(nodes)

	if m.Count() != 3 {
		t.Errorf("Count() = %d, want 3", m.Count())
	}
}

// TestVNodeFocusManager_FocusFirstLast tests FocusFirst and FocusLast
func TestVNodeFocusManager_FocusFirstLast(t *testing.T) {
	m := NewVNodeFocusManager()
	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
		newMockFocusable("btn3", "Button 3"),
		newMockFocusable("btn4", "Button 4"),
	}
	m.SetFocusable(nodes)

	// Move to middle
	m.SetFocusByIndex(2)
	if m.CurrentIndex() != 2 {
		t.Errorf("After SetFocusByIndex(2), CurrentIndex() = %d", m.CurrentIndex())
	}

	// FocusFirst
	if !m.FocusFirst() {
		t.Error("FocusFirst() should return true")
	}
	if m.CurrentIndex() != 0 {
		t.Errorf("After FocusFirst(), CurrentIndex() = %d, want 0", m.CurrentIndex())
	}

	// Move to middle again
	m.SetFocusByIndex(2)

	// FocusLast
	if !m.FocusLast() {
		t.Error("FocusLast() should return true")
	}
	if m.CurrentIndex() != 3 {
		t.Errorf("After FocusLast(), CurrentIndex() = %d, want 3", m.CurrentIndex())
	}
}

// =============================================================================
// CollectFocusable Tests
// =============================================================================

// TestCollectFocusable_NilRoot tests collecting from nil root
func TestCollectFocusable_NilRoot(t *testing.T) {
	result := CollectFocusable(nil)
	if result != nil {
		t.Errorf("CollectFocusable(nil) should return nil slice, got %v", result)
	}
}

// TestCollectFocusable_EmptyTree tests collecting from tree with no focusable nodes
func TestCollectFocusable_EmptyTree(t *testing.T) {
	root := VStack() // Empty VStack
	result := CollectFocusable(root)
	if len(result) != 0 {
		t.Errorf("CollectFocusable(empty tree) should return empty slice, got %d elements", len(result))
	}
}

// TestCollectFocusable_NestedStructure tests collecting from nested VNode tree
func TestCollectFocusable_NestedStructure(t *testing.T) {
	// Create nested structure with focusable nodes at different levels
	btn1 := newMockFocusable("btn1", "Button 1")
	btn2 := newMockFocusable("btn2", "Button 2")

	root := VStack(
		HStack(
			Element("text").Prop("content", "Label").Build(),
			btn1,
		),
		VStack(
			btn2,
		),
	)

	result := CollectFocusable(root)
	if len(result) != 2 {
		t.Errorf("CollectFocusable(nested) = %d elements, want 2", len(result))
	}
}

// TestFindFocusableByID tests finding focusable nodes by ID
func TestFindFocusableByID(t *testing.T) {
	btn1 := newMockFocusable("btn1", "Button 1")
	btn2 := newMockFocusable("btn2", "Button 2")
	btn3 := newMockFocusable("btn3", "Button 3")

	root := VStack(btn1, btn2, btn3)

	// Find existing
	found := FindFocusableByID(root, "btn2")
	if found == nil {
		t.Fatal("FindFocusableByID(btn2) should find node")
	}
	if found.GetFocusID() != "btn2" {
		t.Errorf("Found node ID = %s, want btn2", found.GetFocusID())
	}

	// Find non-existing
	notFound := FindFocusableByID(root, "nonexistent")
	if notFound != nil {
		t.Error("FindFocusableByID(nonexistent) should return nil")
	}

	// Find in nil root
	nilResult := FindFocusableByID(nil, "btn1")
	if nilResult != nil {
		t.Error("FindFocusableByID(nil, ...) should return nil")
	}
}

// TestVNodeFocusManager_DebugString tests DebugString output
func TestVNodeFocusManager_DebugString(t *testing.T) {
	m := NewVNodeFocusManager()

	// Empty manager
	s := m.DebugString()
	if s == "" {
		t.Error("DebugString() should not be empty")
	}

	// With nodes
	nodes := []FocusableVNode{
		newMockFocusable("btn1", "Button 1"),
		newMockFocusable("btn2", "Button 2"),
	}
	m.SetFocusable(nodes)

	s = m.DebugString()
	t.Logf("DebugString: %s", s)
	// Just verify it contains relevant info
	if s == "" {
		t.Error("DebugString() should not be empty with nodes")
	}
}
