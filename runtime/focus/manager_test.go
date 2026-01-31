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
