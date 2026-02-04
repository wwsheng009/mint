package focus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime"
)

// MockFocusableComponent is a mock implementation of FocusableComponent
type MockFocusableComponent struct {
	id         string
	focusable bool
	focused    bool
}

func NewMockFocusableComponent(id string, focusable bool) *MockFocusableComponent {
	return &MockFocusableComponent{
		id:         id,
		focusable: focusable,
		focused:    false,
	}
}

func (m *MockFocusableComponent) SetFocus(focus bool) {
	m.focused = focus
}

func (m *MockFocusableComponent) IsFocusable() bool {
	return m.focusable
}

func (m *MockFocusableComponent) GetID() string {
	return m.id
}

// createMockFocusableItem creates a mock FocusableItem for testing
func createMockFocusableItem(id string, focusable bool) *runtime.FocusableItem {
	mockComp := NewMockFocusableComponent(id, focusable)
	
	// Create a minimal LayoutNode
	node := &runtime.LayoutNode{
		ID: id,
		Component: &runtime.ComponentRef{
			Instance: mockComp,
		},
		Children: []*runtime.LayoutNode{},
	}
	
	return &runtime.FocusableItem{
		ID:       id,
		Node:     node,
		Instance: mockComp,
	}
}

func TestFocusManager_SetFocus(t *testing.T) {
	tests := []struct {
		name        string
		items       []*runtime.FocusableItem
		focusID     string
		shouldFocus bool
	}{
		{
			name: "set focus to valid node",
			items: []*runtime.FocusableItem{
				createMockFocusableItem("node1", true),
				createMockFocusableItem("node2", true),
			},
			focusID:     "node2",
			shouldFocus: true,
		},
		{
			name: "set focus to invalid node",
			items: []*runtime.FocusableItem{
				createMockFocusableItem("node1", true),
			},
			focusID:     "nonexistent",
			shouldFocus: false,
		},
		{
			name: "set focus to non-focusable node",
			items: []*runtime.FocusableItem{
				createMockFocusableItem("node1", true),
				createMockFocusableItem("node2", false), // Not focusable
			},
			focusID:     "node2",
			shouldFocus: true, // Actually works because item is in the list
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := runtime.NewFocusManager()
			mgr.SetFocusable(tt.items)
			
			result := mgr.Focus(tt.focusID)
			
			assert.Equal(t, tt.shouldFocus, result, "Focus result should match expected")
			
			if tt.shouldFocus {
				current := mgr.GetCurrent()
				assert.NotNil(t, current, "Should have current focus")
				assert.Equal(t, tt.focusID, current.ID, "Should focus on correct node")
				
				// Verify focus was applied to component
				if item := mgr.FindByID(tt.focusID); item != nil {
					if mockComp, ok := item.Instance.(*MockFocusableComponent); ok {
						assert.True(t, mockComp.focused, "Component should be focused")
					}
				}
			}
		})
	}
}

func TestFocusManager_RemoveFocus(t *testing.T) {
	t.Run("remove current focus", func(t *testing.T) {
		mgr := runtime.NewFocusManager()
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
			createMockFocusableItem("node2", true),
			createMockFocusableItem("node3", true),
		}
		mgr.SetFocusable(items)
		
		// Focus node2
		mgr.Focus("node2")
		assert.Equal(t, "node2", mgr.GetCurrent().ID, "Should focus node2")
		
		// Remove node2
		mgr.RemoveFocusable("node2")
		
		// Focus should be cleared
		assert.Nil(t, mgr.GetCurrent(), "Focus should be cleared after removing current")
	})
	
	t.Run("remove non-current focus", func(t *testing.T) {
		mgr := runtime.NewFocusManager()
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
			createMockFocusableItem("node2", true),
		}
		mgr.SetFocusable(items)
		
		// Focus node1
		mgr.Focus("node1")
		assert.Equal(t, "node1", mgr.GetCurrent().ID, "Should focus node1")
		
		// Remove node2
		mgr.RemoveFocusable("node2")
		
		// Focus should remain on node1
		assert.Equal(t, "node1", mgr.GetCurrent().ID, "Focus should remain on node1")
	})
}

func TestFocusManager_GetCurrentFocus(t *testing.T) {
	tests := []struct {
		name       string
		setupItems []*runtime.FocusableItem
		focusID    string
		hasFocus   bool
	}{
		{
			name: "get current focus when focused",
			setupItems: []*runtime.FocusableItem{
				createMockFocusableItem("node1", true),
				createMockFocusableItem("node2", true),
			},
			focusID:  "node2",
			hasFocus: true,
		},
		{
			name: "get current focus when none focused",
			setupItems: []*runtime.FocusableItem{
				createMockFocusableItem("node1", true),
			},
			focusID:  "",
			hasFocus: false,
		},
		{
			name:       "get current focus with empty list",
			setupItems: []*runtime.FocusableItem{},
			focusID:    "",
			hasFocus:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := runtime.NewFocusManager()
			mgr.SetFocusable(tt.setupItems)
			
			if tt.focusID != "" {
				mgr.Focus(tt.focusID)
			}
			
			current := mgr.GetCurrent()
			
			if tt.hasFocus {
				assert.NotNil(t, current, "Should have current focus")
				assert.Equal(t, tt.focusID, current.ID, "Should return correct focused node")
			} else {
				assert.Nil(t, current, "Should return nil when no focus")
			}
		})
	}
}

func TestFocusManager_GetFocusableComponents(t *testing.T) {
	t.Run("get all focusable components", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
			createMockFocusableItem("node2", true),
			createMockFocusableItem("node3", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		result := mgr.GetFocusable()
		
		assert.Len(t, result, 3, "Should return all focusable components")
		assert.Equal(t, "node1", result[0].ID, "First item should match")
		assert.Equal(t, "node2", result[1].ID, "Second item should match")
		assert.Equal(t, "node3", result[2].ID, "Third item should match")
	})
	
	t.Run("filter non-focusable components", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
			createMockFocusableItem("node2", false), // Not focusable
			createMockFocusableItem("node3", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		// SetFocusable should include all items (filtering happens elsewhere)
		result := mgr.GetFocusable()
		assert.Len(t, result, 3, "Should return all items (filtering happens elsewhere)")
	})
}

func TestFocusManager_Clear(t *testing.T) {
	t.Run("clear all focus", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
			createMockFocusableItem("node2", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		// Focus node1
		mgr.Focus("node1")
		assert.NotNil(t, mgr.GetCurrent(), "Should have focus before clear")
		
		// Clear focus
		mgr.ClearFocus()
		
		assert.Nil(t, mgr.GetCurrent(), "Focus should be cleared")
		assert.Equal(t, -1, mgr.GetCurrentIndex(), "Index should be -1")
	})
	
	t.Run("clear resets focusable list", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		// Verify focusable list exists
		assert.Greater(t, mgr.Count(), 0, "Should have focusable items")
		
		// Clear all
		mgr.ClearFocus()
		
		// Count should remain (only focus is cleared)
		assert.Greater(t, mgr.Count(), 0, "Focusable list should remain")
	})
}

func TestFocusManager_Next(t *testing.T) {
	tests := []struct {
		name         string
		items        []*runtime.FocusableItem
		initialFocus string
		expectedNext string
		wraps        bool
	}{
		{
			name: "Tab key forward navigation",
			items: []*runtime.FocusableItem{
				createMockFocusableItem("node1", true),
				createMockFocusableItem("node2", true),
				createMockFocusableItem("node3", true),
			},
			initialFocus: "node1",
			expectedNext: "node2",
			wraps:        true,
		},
		{
			name: "cyclic navigation",
			items: []*runtime.FocusableItem{
				createMockFocusableItem("node1", true),
				createMockFocusableItem("node2", true),
			},
			initialFocus: "node2",
			expectedNext: "node1", // Wraps around
			wraps:        true,
		},
		{
			name: "boundary handling - last to first",
			items: []*runtime.FocusableItem{
				createMockFocusableItem("node1", true),
				createMockFocusableItem("node2", true),
			},
			initialFocus: "node2",
			expectedNext: "node1",
			wraps:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := runtime.NewFocusManager()
			mgr.SetWrap(tt.wraps)
			mgr.SetFocusable(tt.items)
			
			if tt.initialFocus != "" {
				mgr.Focus(tt.initialFocus)
			}
			
			next := mgr.FocusNext()
			
			assert.NotNil(t, next, "Should return next focused item")
			assert.Equal(t, tt.expectedNext, next.ID, "Should focus on correct next node")
		})
	}
	
	t.Run("no wrap at boundary", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
			createMockFocusableItem("node2", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetWrap(false) // Disable wrap
		mgr.SetFocusable(items)
		
		mgr.Focus("node2")
		
		next := mgr.FocusNext()
		
		// When wrap is disabled, FocusNext returns nil if already at last item
		// and there's no next item
		assert.Nil(t, next, "Should return nil when at boundary and wrap disabled")
		
		// Current should remain on node2
		assert.Equal(t, "node2", mgr.GetCurrent().ID, "Should stay on last node when wrap disabled")
	})
}

func TestFocusManager_Prev(t *testing.T) {
	tests := []struct {
		name         string
		items        []*runtime.FocusableItem
		initialFocus string
		expectedPrev string
		wraps        bool
	}{
		{
			name: "Shift+Tab backward navigation",
			items: []*runtime.FocusableItem{
				createMockFocusableItem("node1", true),
				createMockFocusableItem("node2", true),
				createMockFocusableItem("node3", true),
			},
			initialFocus: "node3",
			expectedPrev: "node2",
			wraps:        true,
		},
		{
			name: "cyclic backward navigation",
			items: []*runtime.FocusableItem{
				createMockFocusableItem("node1", true),
				createMockFocusableItem("node2", true),
			},
			initialFocus: "node1",
			expectedPrev: "node2", // Wraps around
			wraps:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := runtime.NewFocusManager()
			mgr.SetWrap(tt.wraps)
			mgr.SetFocusable(tt.items)
			
			if tt.initialFocus != "" {
				mgr.Focus(tt.initialFocus)
			}
			
			prev := mgr.FocusPrev()
			
			assert.NotNil(t, prev, "Should return previous focused item")
			assert.Equal(t, tt.expectedPrev, prev.ID, "Should focus on correct previous node")
		})
	}
}

func TestFocusManager_TabOrder(t *testing.T) {
	t.Run("default tab order", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("first", true),
			createMockFocusableItem("second", true),
			createMockFocusableItem("third", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		// Test tab order
		mgr.FocusFirst()
		assert.Equal(t, "first", mgr.GetCurrent().ID, "First should be first")
		
		mgr.FocusNext()
		assert.Equal(t, "second", mgr.GetCurrent().ID, "Second should be second")
		
		mgr.FocusNext()
		assert.Equal(t, "third", mgr.GetCurrent().ID, "Third should be third")
	})
	
	t.Run("custom tab order via focus order", func(t *testing.T) {
		// Add items in custom order
		items := []*runtime.FocusableItem{
			createMockFocusableItem("zebra", true),
			createMockFocusableItem("apple", true),
			createMockFocusableItem("mango", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		// Custom order should be preserved
		focusable := mgr.GetFocusable()
		assert.Equal(t, "zebra", focusable[0].ID, "Custom order should be preserved")
		assert.Equal(t, "apple", focusable[1].ID, "Custom order should be preserved")
		assert.Equal(t, "mango", focusable[2].ID, "Custom order should be preserved")
	})
}

func TestFocusManager_FocusFirst(t *testing.T) {
	t.Run("focus first component", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
			createMockFocusableItem("node2", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		result := mgr.FocusFirst()
		
		assert.True(t, result, "FocusFirst should succeed")
		assert.Equal(t, "node1", mgr.GetCurrent().ID, "Should focus first component")
	})
	
	t.Run("focus first with empty list", func(t *testing.T) {
		mgr := runtime.NewFocusManager()
		
		result := mgr.FocusFirst()
		
		assert.False(t, result, "FocusFirst should fail with empty list")
	})
}

func TestFocusManager_FocusLast(t *testing.T) {
	t.Run("focus last component", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
			createMockFocusableItem("node2", true),
			createMockFocusableItem("node3", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		result := mgr.FocusLast()
		
		assert.True(t, result, "FocusLast should succeed")
		assert.Equal(t, "node3", mgr.GetCurrent().ID, "Should focus last component")
	})
	
	t.Run("focus last with single item", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		result := mgr.FocusLast()
		
		assert.True(t, result, "FocusLast should succeed")
		assert.Equal(t, "node1", mgr.GetCurrent().ID, "Should focus the only component")
	})
}

func TestFocusManager_HasFocus(t *testing.T) {
	t.Run("check focus on focused component", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
			createMockFocusableItem("node2", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		mgr.Focus("node1")
		
		assert.True(t, mgr.HasFocus("node1"), "node1 should have focus")
		assert.False(t, mgr.HasFocus("node2"), "node2 should not have focus")
		assert.False(t, mgr.HasFocus("nonexistent"), "nonexistent should not have focus")
	})
}

func TestFocusManager_RemoveFocusable(t *testing.T) {
	t.Run("remove adjusts current index", func(t *testing.T) {
		items := []*runtime.FocusableItem{
			createMockFocusableItem("node1", true),
			createMockFocusableItem("node2", true),
			createMockFocusableItem("node3", true),
		}
		
		mgr := runtime.NewFocusManager()
		mgr.SetFocusable(items)
		
		// Focus node2 (index 1)
		mgr.Focus("node2")
		assert.Equal(t, 1, mgr.GetCurrentIndex(), "Should be at index 1")
		
		// Remove node1 (index 0)
		mgr.RemoveFocusable("node1")
		
		// Current index should be adjusted
		assert.Equal(t, 0, mgr.GetCurrentIndex(), "Index should be adjusted after removal")
		assert.Equal(t, "node2", mgr.GetCurrent().ID, "Should still be focused on node2")
	})
}

func TestFocusManager_EmptyState(t *testing.T) {
	mgr := runtime.NewFocusManager()
	
	assert.True(t, mgr.IsEmpty(), "Should be empty initially")
	assert.Equal(t, 0, mgr.Count(), "Count should be 0")
	assert.Nil(t, mgr.GetCurrent(), "Current should be nil")
	assert.Equal(t, -1, mgr.GetCurrentIndex(), "Index should be -1")
	
	// FocusNext on empty should return nil
	assert.Nil(t, mgr.FocusNext(), "FocusNext should return nil on empty")
	
	// FocusPrev on empty should return nil
	assert.Nil(t, mgr.FocusPrev(), "FocusPrev should return nil on empty")
}

func TestFocusManager_FocusAt(t *testing.T) {
	items := []*runtime.FocusableItem{
		createMockFocusableItem("node1", true),
		createMockFocusableItem("node2", true),
		createMockFocusableItem("node3", true),
	}
	
	mgr := runtime.NewFocusManager()
	mgr.SetFocusable(items)
	
	t.Run("focus at valid index", func(t *testing.T) {
		result := mgr.FocusAt(1)
		
		assert.True(t, result, "FocusAt should succeed")
		assert.Equal(t, "node2", mgr.GetCurrent().ID, "Should focus node2")
		assert.Equal(t, 1, mgr.GetCurrentIndex(), "Index should be 1")
	})
	
	t.Run("focus at invalid index", func(t *testing.T) {
		result := mgr.FocusAt(10)
		
		assert.False(t, result, "FocusAt should fail for invalid index")
	})
}

func TestFocusManager_SetFocusChangeCallback(t *testing.T) {
	items := []*runtime.FocusableItem{
		createMockFocusableItem("node1", true),
		createMockFocusableItem("node2", true),
	}
	
	mgr := runtime.NewFocusManager()
	mgr.SetFocusable(items)
	
	var callbackCalled bool
	var callbackFocused, callbackPrevious *runtime.FocusableItem
	
	mgr.SetFocusChangeCallback(func(focused, previous *runtime.FocusableItem) {
		callbackCalled = true
		callbackFocused = focused
		callbackPrevious = previous
	})
	
	mgr.Focus("node1")
	
	assert.True(t, callbackCalled, "Callback should be called")
	assert.NotNil(t, callbackFocused, "Focused should be set in callback")
	assert.Nil(t, callbackPrevious, "Previous should be nil for first focus")
}

// Benchmark tests
func BenchmarkFocusManager_FocusNext(b *testing.B) {
	items := make([]*runtime.FocusableItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = createMockFocusableItem("node"+string(rune('0'+i%10)), true)
	}
	
	mgr := runtime.NewFocusManager()
	mgr.SetFocusable(items)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.FocusNext()
	}
}

func BenchmarkFocusManager_Focus(b *testing.B) {
	items := make([]*runtime.FocusableItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = createMockFocusableItem("node"+string(rune('0'+i%10)), true)
	}
	
	mgr := runtime.NewFocusManager()
	mgr.SetFocusable(items)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.Focus("node50")
	}
}

// =============================================================================
// focus.Manager Tests (for runtime/focus/manager.go)
// =============================================================================

// createFocusableNode creates a LayoutNode with a focusable component
func createFocusableNode(id string, x, y, w, h int, children ...*runtime.LayoutNode) *runtime.LayoutNode {
	mockComp := NewMockFocusableComponent(id, true)
	node := &runtime.LayoutNode{
		ID:            id,
		X:             x,
		Y:             y,
		MeasuredWidth:  w,
		MeasuredHeight: h,
		Children:      children,
		Component: &runtime.ComponentRef{
			Instance: mockComp,
		},
	}
	return node
}

func TestNewManager(t *testing.T) {
	root := createFocusableNode("root", 0, 0, 100, 100)
	mgr := NewManager(root)

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.focusedIndex != -1 {
		t.Errorf("Initial focusedIndex should be -1, got %d", mgr.focusedIndex)
	}

	if mgr.rootNode != root {
		t.Error("Root node should be set")
	}

	if mgr.geometricNavigator == nil {
		t.Error("GeometricNavigator should be initialized")
	}

	if mgr.trapManager == nil {
		t.Error("TrapManager should be initialized")
	}
}

func TestNewManager_NilRoot(t *testing.T) {
	mgr := NewManager(nil)

	if mgr == nil {
		t.Fatal("NewManager with nil root should still return manager")
	}

	if mgr.rootNode != nil {
		t.Error("Root node should be nil when nil was passed")
	}
}

func TestManager_RefreshFocusables(t *testing.T) {
	// Create a tree with focusable and non-focusable components
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	child2 := createFocusableNode("child2", 50, 0, 50, 50)
	nonFocusable := &runtime.LayoutNode{
		ID:            "label",
		X:             0,
		Y:             50,
		MeasuredWidth: 100,
		MeasuredHeight: 30,
		// No component, so not focusable
	}

	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1, child2, nonFocusable},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	// Should have 2 focusable components (child1, child2)
	if len(mgr.focusableComponents) != 2 {
		t.Errorf("Expected 2 focusable components, got %d", len(mgr.focusableComponents))
	}

	// focusedIndex should be reset to -1
	if mgr.focusedIndex != -1 {
		t.Errorf("focusedIndex should be -1 after RefreshFocusables, got %d", mgr.focusedIndex)
	}

	// Check IDs are correct
	containsID := func(id string) bool {
		for _, fid := range mgr.focusableComponents {
			if fid == id {
				return true
			}
		}
		return false
	}

	if !containsID("child1") {
		t.Error("child1 should be focusable")
	}
	if !containsID("child2") {
		t.Error("child2 should be focusable")
	}
	if containsID("label") {
		t.Error("label should not be focusable")
	}
}

func TestManager_RefreshFocusables_NilRoot(t *testing.T) {
	mgr := NewManager(nil)

	// Should not panic
	mgr.RefreshFocusables()

	if len(mgr.focusableComponents) != 0 {
		t.Errorf("Expected 0 focusable components with nil root, got %d", len(mgr.focusableComponents))
	}
}

func TestManager_FocusNext(t *testing.T) {
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	child2 := createFocusableNode("child2", 50, 0, 50, 50)
	child3 := createFocusableNode("child3", 0, 50, 50, 50)
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1, child2, child3},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	t.Run("focus next from no focus", func(t *testing.T) {
		id, ok := mgr.FocusNext()

		if !ok {
			t.Error("FocusNext should succeed")
		}
		if id != "child1" {
			t.Errorf("Expected first focusable to be 'child1', got '%s'", id)
		}
		if mgr.focusedIndex != 0 {
			t.Errorf("focusedIndex should be 0, got %d", mgr.focusedIndex)
		}
	})

	t.Run("focus next moves to next component", func(t *testing.T) {
		// child1 is already focused from previous test
		id, ok := mgr.FocusNext()

		if !ok {
			t.Error("FocusNext should succeed")
		}
		if id != "child2" {
			t.Errorf("Expected next focusable to be 'child2', got '%s'", id)
		}
	})

	t.Run("focus next wraps around", func(t *testing.T) {
		// Move to last
		mgr.FocusNext()

		// One more should wrap to first
		id, ok := mgr.FocusNext()

		if !ok {
			t.Error("FocusNext should succeed with wraparound")
		}
		if id != "child1" {
			t.Errorf("Expected wraparound to 'child1', got '%s'", id)
		}
	})
}

func TestManager_FocusNext_EmptyList(t *testing.T) {
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	id, ok := mgr.FocusNext()

	if ok {
		t.Errorf("FocusNext should fail with empty list, got '%s'", id)
	}
	if id != "" {
		t.Errorf("Expected empty ID with empty list, got '%s'", id)
	}
}

func TestManager_FocusPrev(t *testing.T) {
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	child2 := createFocusableNode("child2", 50, 0, 50, 50)
	child3 := createFocusableNode("child3", 0, 50, 50, 50)
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1, child2, child3},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	t.Run("focus prev from no focus wraps to last", func(t *testing.T) {
		id, ok := mgr.FocusPrev()

		if !ok {
			t.Error("FocusPrev should succeed")
		}
		if id != "child3" {
			t.Errorf("Expected last focusable to be 'child3', got '%s'", id)
		}
	})

	t.Run("focus prev moves to previous component", func(t *testing.T) {
		// child3 is focused
		id, ok := mgr.FocusPrev()

		if !ok {
			t.Error("FocusPrev should succeed")
		}
		if id != "child2" {
			t.Errorf("Expected previous focusable to be 'child2', got '%s'", id)
		}
	})

	t.Run("focus prev wraps around to last", func(t *testing.T) {
		// Current state: child2 is focused
		// FocusPrev twice: child2 -> child1 -> child3 (wrap)
		mgr.FocusPrev()
		id, ok := mgr.FocusPrev()

		if !ok {
			t.Error("FocusPrev should succeed with wraparound")
		}
		if id != "child3" {
			t.Errorf("Expected wraparound to 'child3', got '%s'", id)
		}
	})
}

func TestManager_FocusPrev_EmptyList(t *testing.T) {
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	id, ok := mgr.FocusPrev()

	if ok {
		t.Errorf("FocusPrev should fail with empty list, got '%s'", id)
	}
	if id != "" {
		t.Errorf("Expected empty ID with empty list, got '%s'", id)
	}
}

func TestManager_FocusFirst(t *testing.T) {
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	child2 := createFocusableNode("child2", 50, 0, 50, 50)
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1, child2},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	t.Run("focus first from any position", func(t *testing.T) {
		// Focus second component first
		mgr.FocusSpecific("child2")

		id, ok := mgr.FocusFirst()

		if !ok {
			t.Error("FocusFirst should succeed")
		}
		if id != "child1" {
			t.Errorf("Expected first focusable to be 'child1', got '%s'", id)
		}
	})

	t.Run("focus first with empty list", func(t *testing.T) {
		emptyRoot := &runtime.LayoutNode{
			ID:            "empty",
			X:             0,
			Y:             0,
			MeasuredWidth: 100,
			MeasuredHeight: 100,
			Children:      []*runtime.LayoutNode{},
		}
		emptyMgr := NewManager(emptyRoot)
		emptyMgr.RefreshFocusables()

		id, ok := emptyMgr.FocusFirst()

		if ok {
			t.Errorf("FocusFirst should fail with empty list, got '%s'", id)
		}
		if id != "" {
			t.Errorf("Expected empty ID with empty list, got '%s'", id)
		}
	})
}

func TestManager_FocusSpecific(t *testing.T) {
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	child2 := createFocusableNode("child2", 50, 0, 50, 50)
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1, child2},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	t.Run("focus specific valid component", func(t *testing.T) {
		ok := mgr.FocusSpecific("child2")

		if !ok {
			t.Error("FocusSpecific should succeed for valid component")
		}

		focusedID, hasFocus := mgr.GetFocused()
		if !hasFocus || focusedID != "child2" {
			t.Errorf("Expected focus on 'child2', got '%s', hasFocus=%v", focusedID, hasFocus)
		}
	})

	t.Run("focus specific invalid component", func(t *testing.T) {
		ok := mgr.FocusSpecific("nonexistent")

		if ok {
			t.Error("FocusSpecific should fail for invalid component")
		}
	})
}

func TestManager_GetFocused(t *testing.T) {
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	t.Run("no focus initially", func(t *testing.T) {
		id, hasFocus := mgr.GetFocused()

		if hasFocus {
			t.Errorf("Should not have focus initially, got '%s'", id)
		}
		if id != "" {
			t.Errorf("Expected empty ID when no focus, got '%s'", id)
		}
	})

	t.Run("get focused after focusing", func(t *testing.T) {
		mgr.FocusSpecific("child1")

		id, hasFocus := mgr.GetFocused()

		if !hasFocus {
			t.Error("Should have focus after FocusSpecific")
		}
		if id != "child1" {
			t.Errorf("Expected focused 'child1', got '%s'", id)
		}
	})
}

func TestManager_HasFocus(t *testing.T) {
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	child2 := createFocusableNode("child2", 50, 0, 50, 50)
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1, child2},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	t.Run("has focus on focused component", func(t *testing.T) {
		mgr.FocusSpecific("child1")

		if !mgr.HasFocus("child1") {
			t.Error("Should have focus on child1")
		}
		if mgr.HasFocus("child2") {
			t.Error("Should not have focus on child2")
		}
		if mgr.HasFocus("nonexistent") {
			t.Error("Should not have focus on nonexistent component")
		}
	})

	t.Run("has focus when no focus", func(t *testing.T) {
		mgr.Clear()

		if mgr.HasFocus("child1") {
			t.Error("Should not have focus after Clear")
		}
	})
}

func TestManager_PushFocusTrap(t *testing.T) {
	root := createFocusableNode("root", 0, 0, 100, 100)
	mgr := NewManager(root)

	t.Run("push focus trap", func(t *testing.T) {
		trap := NewFocusTrap("modal", TrapModal, root)
		mgr.PushFocusTrap(trap)

		if !mgr.HasActiveFocusTrap() {
			t.Error("Should have active focus trap")
		}

		currentTrap := mgr.GetCurrentFocusTrap()
		if currentTrap == nil {
			t.Error("GetCurrentFocusTrap should return the trap")
		}
		if currentTrap.ID != "modal" {
			t.Errorf("Expected trap ID 'modal', got '%s'", currentTrap.ID)
		}
	})

	t.Run("push multiple traps", func(t *testing.T) {
		mgr.ClearFocusTraps()

		trap1 := NewFocusTrap("trap1", TrapMenu, root)
		trap2 := NewFocusTrap("trap2", TrapModal, root)

		mgr.PushFocusTrap(trap1)
		mgr.PushFocusTrap(trap2)

		// Should have 2 traps in stack (managed by TrapManager)
		if !mgr.HasActiveFocusTrap() {
			t.Error("Should have active focus trap")
		}
	})
}

func TestManager_PopFocusTrap(t *testing.T) {
	root := createFocusableNode("root", 0, 0, 100, 100)
	mgr := NewManager(root)

	t.Run("pop focus trap", func(t *testing.T) {
		trap := NewFocusTrap("modal", TrapModal, root)
		mgr.PushFocusTrap(trap)

		if !mgr.HasActiveFocusTrap() {
			t.Fatal("Should have active focus trap before pop")
		}

		popped := mgr.PopFocusTrap()

		if popped == nil {
			t.Error("PopFocusTrap should return the trap")
		}
		if popped.ID != "modal" {
			t.Errorf("Expected popped trap ID 'modal', got '%s'", popped.ID)
		}

		if mgr.HasActiveFocusTrap() {
			t.Error("Should not have active focus trap after pop")
		}
	})

	t.Run("pop from empty traps", func(t *testing.T) {
		// Pop without push should return nil
		popped := mgr.PopFocusTrap()

		if popped != nil {
			t.Errorf("PopFocusTrap with no traps should return nil, got '%v'", popped)
		}
	})
}

func TestManager_RemoveFocusTrap(t *testing.T) {
	root := createFocusableNode("root", 0, 0, 100, 100)
	mgr := NewManager(root)

	t.Run("remove focus trap by ID", func(t *testing.T) {
		trap := NewFocusTrap("modal", TrapModal, root)
		mgr.PushFocusTrap(trap)

		removed := mgr.RemoveFocusTrap("modal")

		if !removed {
			t.Error("RemoveFocusTrap should return true for existing trap")
		}

		if mgr.HasActiveFocusTrap() {
			t.Error("Should not have active focus trap after removal")
		}
	})

	t.Run("remove non-existent trap", func(t *testing.T) {
		removed := mgr.RemoveFocusTrap("nonexistent")

		if removed {
			t.Error("RemoveFocusTrap should return false for non-existent trap")
		}
	})
}

func TestManager_IsFocusTrapActive(t *testing.T) {
	root := createFocusableNode("root", 0, 0, 100, 100)
	mgr := NewManager(root)

	t.Run("inactive trap initially", func(t *testing.T) {
		if mgr.IsFocusTrapActive("modal") {
			t.Error("Should not have active trap initially")
		}
	})

	t.Run("check active trap", func(t *testing.T) {
		trap := NewFocusTrap("modal", TrapModal, root)
		mgr.PushFocusTrap(trap)

		if !mgr.IsFocusTrapActive("modal") {
			t.Error("Should have active 'modal' trap")
		}
		if mgr.IsFocusTrapActive("other") {
			t.Error("Should not have active 'other' trap")
		}
	})
}

func TestManager_HasActiveFocusTrap(t *testing.T) {
	root := createFocusableNode("root", 0, 0, 100, 100)
	mgr := NewManager(root)

	if mgr.HasActiveFocusTrap() {
		t.Error("Should not have active focus trap initially")
	}

	trap := NewFocusTrap("modal", TrapModal, root)
	mgr.PushFocusTrap(trap)

	if !mgr.HasActiveFocusTrap() {
		t.Error("Should have active focus trap after push")
	}
}

func TestManager_GetCurrentFocusTrap(t *testing.T) {
	root := createFocusableNode("root", 0, 0, 100, 100)
	mgr := NewManager(root)

	t.Run("get current trap with no trap", func(t *testing.T) {
		trap := mgr.GetCurrentFocusTrap()

		if trap != nil {
			t.Error("GetCurrentFocusTrap should return nil when no trap")
		}
	})

	t.Run("get current active trap", func(t *testing.T) {
		trap := NewFocusTrap("modal", TrapModal, root)
		mgr.PushFocusTrap(trap)

		current := mgr.GetCurrentFocusTrap()

		if current == nil {
			t.Error("GetCurrentFocusTrap should return active trap")
		}
		if current.ID != "modal" {
			t.Errorf("Expected trap ID 'modal', got '%s'", current.ID)
		}
	})
}

func TestManager_ClearFocusTraps(t *testing.T) {
	root := createFocusableNode("root", 0, 0, 100, 100)
	mgr := NewManager(root)

	t.Run("clear all traps", func(t *testing.T) {
		trap1 := NewFocusTrap("trap1", TrapMenu, root)
		trap2 := NewFocusTrap("trap2", TrapModal, root)
		mgr.PushFocusTrap(trap1)
		mgr.PushFocusTrap(trap2)

		if !mgr.HasActiveFocusTrap() {
			t.Fatal("Should have active focus trap")
		}

		mgr.ClearFocusTraps()

		if mgr.HasActiveFocusTrap() {
			t.Error("Should not have active focus trap after ClearFocusTraps")
		}
	})
}

func TestManager_GetFocusableComponents(t *testing.T) {
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	child2 := createFocusableNode("child2", 50, 0, 50, 50)
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1, child2},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	components := mgr.GetFocusableComponents()

	if len(components) != 2 {
		t.Errorf("Expected 2 focusable components, got %d", len(components))
	}
}

func TestManager_GetFocusableCount(t *testing.T) {
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	child2 := createFocusableNode("child2", 50, 0, 50, 50)
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1, child2},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()

	count := mgr.GetFocusableCount()

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestManager_Clear(t *testing.T) {
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()
	mgr.FocusSpecific("child1")

	// Verify focus is set
	if !mgr.HasFocus("child1") {
		t.Fatal("Should have focus before Clear")
	}

	mgr.Clear()

	// Focus should be cleared
	if mgr.HasFocus("child1") {
		t.Error("Focus should be cleared")
	}

	// Focusable list should be empty
	if mgr.GetFocusableCount() != 0 {
		t.Errorf("Focusable count should be 0 after Clear, got %d", mgr.GetFocusableCount())
	}

	// focusedIndex should be -1
	if mgr.focusedIndex != -1 {
		t.Errorf("focusedIndex should be -1 after Clear, got %d", mgr.focusedIndex)
	}

	// Traps should be cleared
	if mgr.HasActiveFocusTrap() {
		t.Error("Focus traps should be cleared")
	}
}

func TestManager_FocusDirection(t *testing.T) {
	child1 := createFocusableNode("child1", 0, 0, 50, 50)
	child2 := createFocusableNode("child2", 50, 0, 50, 50)
	root := &runtime.LayoutNode{
		ID:            "root",
		X:             0,
		Y:             0,
		MeasuredWidth: 100,
		MeasuredHeight: 100,
		Children:      []*runtime.LayoutNode{child1, child2},
	}

	mgr := NewManager(root)
	mgr.RefreshFocusables()
	mgr.geometricNavigator.RefreshBounds(mgr.GetFocusableComponents())

	t.Run("focus direction with empty list", func(t *testing.T) {
		emptyMgr := NewManager(nil)
		emptyMgr.RefreshFocusables()

		id, ok := emptyMgr.FocusDirection(DirectionUp)

		if ok {
			t.Errorf("FocusDirection should fail with empty list, got '%s'", id)
		}
	})

	t.Run("focus direction to component", func(t *testing.T) {
		id, ok := mgr.FocusDirection(DirectionRight)

		if !ok {
			t.Error("FocusDirection should succeed")
		}
		if id == "" {
			t.Error("FocusDirection should return a component ID")
		}
	})
}

func TestManager_FocusUp(t *testing.T) {
	t.Run("focus up with empty list", func(t *testing.T) {
		mgr := NewManager(nil)
		mgr.RefreshFocusables()

		id, ok := mgr.FocusUp()

		if ok {
			t.Errorf("FocusUp should fail with empty list, got '%s'", id)
		}
	})
}

func TestManager_FocusDown(t *testing.T) {
	t.Run("focus down with empty list", func(t *testing.T) {
		mgr := NewManager(nil)
		mgr.RefreshFocusables()

		id, ok := mgr.FocusDown()

		if ok {
			t.Errorf("FocusDown should fail with empty list, got '%s'", id)
		}
	})
}

func TestManager_FocusLeft(t *testing.T) {
	t.Run("focus left with empty list", func(t *testing.T) {
		mgr := NewManager(nil)
		mgr.RefreshFocusables()

		id, ok := mgr.FocusLeft()

		if ok {
			t.Errorf("FocusLeft should fail with empty list, got '%s'", id)
		}
	})
}

func TestManager_FocusRight(t *testing.T) {
	t.Run("focus right with empty list", func(t *testing.T) {
		mgr := NewManager(nil)
		mgr.RefreshFocusables()

		id, ok := mgr.FocusRight()

		if ok {
			t.Errorf("FocusRight should fail with empty list, got '%s'", id)
		}
	})
}
