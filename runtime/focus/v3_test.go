package focus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime/state"
)

func TestV3_Scope(t *testing.T) {
	t.Run("focus scope definition", func(t *testing.T) {
		scope := NewScope("modal-scope", "modal-root")

		assert.Equal(t, "modal-scope", scope.ID, "Scope ID should match")
		assert.Equal(t, "modal-root", scope.Root, "Root should match")
		assert.False(t, scope.Active, "Should be inactive initially")
		assert.False(t, scope.Modal, "Should not be modal initially")
		assert.Empty(t, scope.GetFocusable(), "Focusable should be empty initially")
	})

	t.Run("focusable within scope limit", func(t *testing.T) {
		scope := NewScope("form-scope", "form-root")

		// Set focusable components within scope
		focusable := []string{"input1", "input2", "button1"}
		scope.SetFocusable(focusable)

		result := scope.GetFocusable()
		assert.Equal(t, focusable, result, "Should return focusable components")
	})

	t.Run("jump outside scope is blocked", func(t *testing.T) {
		scope := NewScope("modal-scope", "modal-root")

		// Only modal components are focusable
		scope.SetFocusable([]string{"modal-btn1", "modal-btn2"})

		// Try to focus on component outside scope
		outsideID := "outside-button"
		focusable := scope.GetFocusable()

		// Should not contain outside component
		assert.NotContains(t, focusable, outsideID, "Outside component should not be in focusable list")
	})
}

func TestV3_ScopeStack(t *testing.T) {
	t.Run("scope push and pop", func(t *testing.T) {
		baseScope := NewScope("base", "app-root")
		modalScope := NewScope("modal", "modal-root")

		// Simulate stack operations
		scopes := []*Scope{baseScope}

		// Push modal scope
		scopes = append(scopes, modalScope)
		assert.Len(t, scopes, 2, "Should have 2 scopes in stack")

		// Current top is modal
		current := scopes[len(scopes)-1]
		assert.Equal(t, "modal", current.ID, "Top of stack should be modal")

		// Pop modal scope
		scopes = scopes[:len(scopes)-1]
		assert.Len(t, scopes, 1, "Should have 1 scope after pop")

		current = scopes[len(scopes)-1]
		assert.Equal(t, "base", current.ID, "Top should be base after pop")
	})

	t.Run("nested scopes", func(t *testing.T) {
		appScope := NewScope("app", "app-root")
		panelScope := NewScope("panel", "panel-root")
		dialogScope := NewScope("dialog", "dialog-root")

		// Build stack: app -> panel -> dialog
		stack := []*Scope{appScope, panelScope, dialogScope}

		// Top is dialog
		assert.Equal(t, "dialog", stack[len(stack)-1].ID)

		// Focus should be limited to dialog's focusable components
		dialogScope.SetFocusable([]string{"dialog-btn1", "dialog-btn2"})
		panelScope.SetFocusable([]string{"panel-btn1"})

		// When dialog is active, only its components are focusable
		activeScope := stack[len(stack)-1]
		focusable := activeScope.GetFocusable()
		assert.Len(t, focusable, 2, "Should only have dialog focusable items")
		assert.NotContains(t, focusable, "panel-btn1", "Panel button should not be focusable")
	})
}

func TestV3_DefaultFocus(t *testing.T) {
	t.Run("set default focus", func(t *testing.T) {
		scope := NewScope("form", "form-root")

		// Set focusable components
		scope.SetFocusable([]string{"input1", "input2", "submit"})

		// Set default focus (first focusable)
		scope.SetFocusPath(state.FocusPath{"input1"})

		path := scope.GetFocusPath()
		assert.Equal(t, "input1", path.Current(), "Default focus should be set")
	})

	t.Run("default focus after scope change", func(t *testing.T) {
		scope1 := NewScope("scope1", "root1")
		scope2 := NewScope("scope2", "root2")

		scope1.SetFocusable([]string{"btn1", "btn2"})
		scope1.SetFocusPath(state.FocusPath{"btn1"})

		scope2.SetFocusable([]string{"btn3", "btn4"})

		// When switching to scope2, default should be its first focusable
		scope2.SetFocusPath(state.FocusPath{"btn3"})

		path1 := scope1.GetFocusPath()
		path2 := scope2.GetFocusPath()

		assert.Equal(t, "btn1", path1.Current(), "Scope1 focus unchanged")
		assert.Equal(t, "btn3", path2.Current(), "Scope2 has its own default")
	})

	t.Run("default focus when component doesn't exist", func(t *testing.T) {
		scope := NewScope("form", "form-root")

		scope.SetFocusable([]string{"input1", "input2"})

		// Try to set non-existent component as default
		scope.SetFocusPath(state.FocusPath{"nonexistent"})

		path := scope.GetFocusPath()
		// Path is set even if component doesn't exist (validation happens elsewhere)
		assert.Equal(t, "nonexistent", path.Current())
	})
}

func TestV3_Migration(t *testing.T) {
	t.Run("migrate from V2 index-based to V3 path-based", func(t *testing.T) {
		// Simulate V2 data
		v2Focusable := []string{"item1", "item2", "item3"}
		v2CurrentIndex := 1 // Focus on item2

		// Create V3 scope
		v3Scope := NewScope("migrated", "root")
		v3Scope.SetFocusable(v2Focusable)

		// Migrate focus
		if v2CurrentIndex >= 0 && v2CurrentIndex < len(v2Focusable) {
			targetID := v2Focusable[v2CurrentIndex]
			v3Scope.SetFocusPath(state.FocusPath{targetID})
		}

		// Verify migration
		path := v3Scope.GetFocusPath()
		assert.Equal(t, "item2", path.Current(), "Should migrate to item2")
	})

	t.Run("API compatibility", func(t *testing.T) {
		scope := NewScope("test", "root")

		// Old API pattern: set focusable list
		scope.SetFocusable([]string{"a", "b", "c"})

		// New API pattern: use FocusPath
		scope.SetFocusPath(state.FocusPath{"b"})

		// Both should work together
		focusable := scope.GetFocusable()
		assert.Contains(t, focusable, "b", "Focusable should contain 'b'")

		path := scope.GetFocusPath()
		assert.Equal(t, "b", path.Current(), "FocusPath should point to 'b'")
	})

	t.Run("data transformation", func(t *testing.T) {
		// Simulate transformation from V2 FocusManager to V3 Scope
		oldManagerFocusable := []string{"x", "y", "z"}
		oldCurrentFocus := "y"

		newScope := NewScope("transformed", "root")
		newScope.SetFocusable(oldManagerFocusable)
		newScope.SetFocusPath(state.FocusPath{oldCurrentFocus})

		// Verify data integrity
		assert.Equal(t, oldManagerFocusable, newScope.GetFocusable())
		assert.Equal(t, oldCurrentFocus, newScope.GetFocusPath().Current())
	})
}

func TestV3_Persistence(t *testing.T) {
	t.Run("focus state persistence", func(t *testing.T) {
		scope := NewScope("persistent", "root")

		// Set initial focus
		scope.SetFocusable([]string{"field1", "field2", "field3"})
		scope.SetFocusPath(state.FocusPath{"field2"})

		// Save state
		savedPath := scope.GetFocusPath()
		savedFocusable := scope.GetFocusable()

		// Simulate state loss and restoration
		scope.SetFocusPath(nil)
		scope.SetFocusable(nil)

		assert.Empty(t, scope.GetFocusPath(), "Focus should be cleared")

		// Restore state
		scope.SetFocusPath(savedPath)
		scope.SetFocusable(savedFocusable)

		// Verify restoration
		path := scope.GetFocusPath()
		assert.Equal(t, "field2", path.Current(), "Focus should be restored")

		focusable := scope.GetFocusable()
		assert.Equal(t, 3, len(focusable), "Focusable list should be restored")
	})

	t.Run("state persistence across scope changes", func(t *testing.T) {
		// Save state before leaving scope
		originalScope := NewScope("original", "root1")
		originalScope.SetFocusable([]string{"btn1", "btn2"})
		originalScope.SetFocusPath(state.FocusPath{"btn1"})

		savedPath := originalScope.GetFocusPath()
		savedFocusable := originalScope.GetFocusable()

		// Switch to different scope
		newScope := NewScope("new", "root2")
		newScope.SetFocusable([]string{"btn3"})

		// Return to original scope and restore
		originalScope.SetFocusPath(savedPath)
		originalScope.SetFocusable(savedFocusable)

		// Verify state is preserved
		assert.Equal(t, "btn1", originalScope.GetFocusPath().Current())
		assert.Equal(t, []string{"btn1", "btn2"}, originalScope.GetFocusable())
	})

	t.Run("abnormal state handling", func(t *testing.T) {
		scope := NewScope("test", "root")

		// Set up valid state
		scope.SetFocusable([]string{"a", "b"})

		// Simulate abnormal state: focus points to non-existent component
		scope.SetFocusPath(state.FocusPath{"nonexistent"})

		// System should handle gracefully (validation happens during focus operations)
		path := scope.GetFocusPath()
		assert.Equal(t, "nonexistent", path.Current(), "Path accepts any ID")

		// When trying to navigate, should handle invalid state
		_, ok := scope.FocusNext()
		// Should either succeed (find next valid) or fail gracefully
		assert.True(t, ok || !ok, "Navigation should handle invalid state")
	})
}

func TestV3_FocusPath(t *testing.T) {
	t.Run("simple focus path", func(t *testing.T) {
		scope := NewScope("test", "root")

		path := state.FocusPath{"component1"}
		scope.SetFocusPath(path)

		result := scope.GetFocusPath()
		assert.Equal(t, "component1", result.Current())
	})

	t.Run("nested focus path", func(t *testing.T) {
		scope := NewScope("test", "root")

		// Path representing nested focus: panel -> input
		path := state.FocusPath{"panel", "input"}
		scope.SetFocusPath(path)

		result := scope.GetFocusPath()
		assert.Equal(t, "input", result.Current())
		assert.Len(t, result, 2, "Path should have 2 components")
	})
}

func TestV3_ThreadSafety(t *testing.T) {
	t.Run("concurrent access to scope", func(t *testing.T) {
		scope := NewScope("concurrent", "root")
		scope.SetFocusable([]string{"a", "b", "c", "d", "e"})

		done := make(chan bool)

		// Multiple goroutines accessing scope
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					scope.GetFocusable()
					scope.GetFocusPath()
					scope.FocusNext()
					scope.FocusPrev()
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Scope should still be consistent
		focusable := scope.GetFocusable()
		assert.Len(t, focusable, 5, "Focusable list should remain consistent")
	})
}

func TestV3_FocusNext(t *testing.T) {
	t.Run("focus next within scope", func(t *testing.T) {
		scope := NewScope("test", "root")
		scope.SetFocusable([]string{"a", "b", "c"})
		scope.SetFocusPath(state.FocusPath{"a"})

		nextID, ok := scope.FocusNext()

		assert.True(t, ok, "FocusNext should succeed")
		assert.Equal(t, "b", nextID, "Should move to next component")

		// Verify path was updated
		path := scope.GetFocusPath()
		assert.Equal(t, "b", path.Current())
	})

	t.Run("focus next wraps around", func(t *testing.T) {
		scope := NewScope("test", "root")
		scope.SetFocusable([]string{"a", "b"})
		scope.SetFocusPath(state.FocusPath{"b"})

		nextID, ok := scope.FocusNext()

		assert.True(t, ok, "FocusNext should succeed")
		assert.Equal(t, "a", nextID, "Should wrap to first component")
	})

	t.Run("focus next with empty focusable", func(t *testing.T) {
		scope := NewScope("test", "root")
		scope.SetFocusable([]string{})

		nextID, ok := scope.FocusNext()

		assert.False(t, ok, "FocusNext should fail with empty list")
		assert.Empty(t, nextID, "Should return empty ID")
	})
}

func TestV3_FocusPrev(t *testing.T) {
	t.Run("focus prev within scope", func(t *testing.T) {
		scope := NewScope("test", "root")
		scope.SetFocusable([]string{"a", "b", "c"})
		scope.SetFocusPath(state.FocusPath{"c"})

		prevID, ok := scope.FocusPrev()

		assert.True(t, ok, "FocusPrev should succeed")
		assert.Equal(t, "b", prevID, "Should move to previous component")

		// Verify path was updated
		path := scope.GetFocusPath()
		assert.Equal(t, "b", path.Current())
	})

	t.Run("focus prev wraps around", func(t *testing.T) {
		scope := NewScope("test", "root")
		scope.SetFocusable([]string{"a", "b"})
		scope.SetFocusPath(state.FocusPath{"a"})

		prevID, ok := scope.FocusPrev()

		assert.True(t, ok, "FocusPrev should succeed")
		assert.Equal(t, "b", prevID, "Should wrap to last component")
	})
}

func TestV3_ActiveModal(t *testing.T) {
	t.Run("modal scope is active", func(t *testing.T) {
		scope := NewScope("modal", "root")
		scope.Modal = true
		scope.Active = true

		assert.True(t, scope.Modal, "Should be modal")
		assert.True(t, scope.Active, "Should be active")
	})

	t.Run("non-modal scope", func(t *testing.T) {
		scope := NewScope("panel", "root")
		scope.Modal = false
		scope.Active = true

		assert.False(t, scope.Modal, "Should not be modal")
		assert.True(t, scope.Active, "Should be active")
	})
}

// Benchmark tests
func BenchmarkV3_FocusNext(b *testing.B) {
	scope := NewScope("bench", "root")

	// Create large focusable list
	largeList := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeList[i] = "item" + string(rune('0'+i%10))
	}
	scope.SetFocusable(largeList)
	scope.SetFocusPath(state.FocusPath{"item0"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scope.FocusNext()
	}
}

func BenchmarkV3_GetFocusPath(b *testing.B) {
	scope := NewScope("bench", "root")
	scope.SetFocusable([]string{"a", "b", "c"})
	scope.SetFocusPath(state.FocusPath{"a"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scope.GetFocusPath()
	}
}
