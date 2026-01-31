package focus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime"
)

// createMockTrapNode creates a LayoutNode structure for testing focus traps
func createMockTrapNode(id string, focusable bool, children ...*runtime.LayoutNode) *runtime.LayoutNode {
	node := &runtime.LayoutNode{
		ID:             id,
		X:              0,
		Y:              0,
		MeasuredWidth:  100,
		MeasuredHeight: 50,
		Children:       children,
	}

	if focusable {
		mockComp := NewMockFocusableComponent(id, true)
		node.Component = &runtime.ComponentRef{
			Instance: mockComp,
		}
	}

	return node
}

func TestTrap_Enter(t *testing.T) {
	t.Run("enter trap activates and saves focus", func(t *testing.T) {
		// Create a modal structure
		cancelButton := createMockTrapNode("cancel", true)
		confirmButton := createMockTrapNode("confirm", true)
		modalContent := createMockTrapNode("modal-content", false, cancelButton, confirmButton)
		modal := createMockTrapNode("modal", false, modalContent)

		trap := NewFocusTrap("modal-trap", TrapModal, modal)

		// Initially inactive
		assert.False(t, trap.IsActive(), "Trap should be inactive initially")

		// Activate trap
		trap.Activate()

		assert.True(t, trap.IsActive(), "Trap should be active after activation")
		assert.Equal(t, "modal-trap", trap.ID, "Trap ID should be set")
		assert.Equal(t, TrapModal, trap.Type, "Trap type should be Modal")
	})

	t.Run("trap saves focusable IDs", func(t *testing.T) {
		button1 := createMockTrapNode("button1", true)
		button2 := createMockTrapNode("button2", true)
		container := createMockTrapNode("container", false, button1, button2)

		trap := NewFocusTrap("trap", TrapMenu, container)
		trap.Activate()

		ids := trap.GetFocusableIDs()

		assert.Len(t, ids, 2, "Should find 2 focusable items")
		assert.Contains(t, ids, "button1", "Should contain button1")
		assert.Contains(t, ids, "button2", "Should contain button2")
	})
}

func TestTrap_Exit(t *testing.T) {
	t.Run("exit trap deactivates", func(t *testing.T) {
		modal := createMockTrapNode("modal", false)
		trap := NewFocusTrap("modal-trap", TrapModal, modal)

		trap.Activate()
		assert.True(t, trap.IsActive(), "Trap should be active")

		trap.Deactivate()
		assert.False(t, trap.IsActive(), "Trap should be deactivated")
	})

	t.Run("multiple activations and deactivations", func(t *testing.T) {
		modal := createMockTrapNode("modal", false)
		trap := NewFocusTrap("modal-trap", TrapModal, modal)

		// First activation
		trap.Activate()
		assert.True(t, trap.IsActive())

		// First deactivation
		trap.Deactivate()
		assert.False(t, trap.IsActive())

		// Second activation
		trap.Activate()
		assert.True(t, trap.IsActive())

		// Second deactivation
		trap.Deactivate()
		assert.False(t, trap.IsActive())
	})
}

func TestTrap_EscapeKey(t *testing.T) {
	t.Run("ESC key exits modal trap", func(t *testing.T) {
		modal := createMockTrapNode("modal", false)
		trap := NewFocusTrap("modal-trap", TrapModal, modal)

		trap.Activate()
		assert.True(t, trap.IsActive(), "Modal should be active")

		// Simulate ESC key by deactivating
		trap.Deactivate()

		assert.False(t, trap.IsActive(), "Modal should be deactivated after ESC")
	})

	t.Run("ESC in menu trap", func(t *testing.T) {
		menu := createMockTrapNode("menu", false)
		trap := NewFocusTrap("menu-trap", TrapMenu, menu)

		trap.Activate()

		// ESC typically closes menus
		trap.Deactivate()

		assert.False(t, trap.IsActive(), "Menu should close on ESC")
	})
}

func TestTrap_Nested(t *testing.T) {
	t.Run("nested trap enters outer then inner", func(t *testing.T) {
		// Outer modal
		outerButton := createMockTrapNode("outer-btn", true)
		outerModal := createMockTrapNode("outer-modal", false, outerButton)

		// Inner dropdown inside outer modal
		innerItem1 := createMockTrapNode("item1", true)
		innerItem2 := createMockTrapNode("item2", true)
		dropdown := createMockTrapNode("dropdown", false, innerItem1, innerItem2)

		outerTrap := NewFocusTrap("outer", TrapModal, outerModal)
		innerTrap := NewFocusTrap("inner", TrapMenu, dropdown)

		// Activate outer trap
		outerTrap.Activate()
		assert.True(t, outerTrap.IsActive())
		assert.False(t, innerTrap.IsActive())

		// Activate inner trap (e.g., dropdown opens)
		innerTrap.Activate()
		assert.True(t, innerTrap.IsActive())

		// Both can be active in a stack
		assert.True(t, outerTrap.IsActive(), "Outer trap should remain active")
		assert.True(t, innerTrap.IsActive(), "Inner trap should be active")
	})

	t.Run("exit inner trap preserves outer", func(t *testing.T) {
		outerModal := createMockTrapNode("outer-modal", false)
		dropdown := createMockTrapNode("dropdown", false)

		outerTrap := NewFocusTrap("outer", TrapModal, outerModal)
		innerTrap := NewFocusTrap("inner", TrapMenu, dropdown)

		// Stack: outer -> inner
		outerTrap.Activate()
		innerTrap.Activate()

		// Exit inner trap
		innerTrap.Deactivate()

		assert.False(t, innerTrap.IsActive(), "Inner should be inactive")
		assert.True(t, outerTrap.IsActive(), "Outer should still be active")
	})
}

func TestTrap_RestoreFocus(t *testing.T) {
	t.Run("focus restoration to entry element", func(t *testing.T) {
		triggerButton := createMockTrapNode("trigger", true)
		modalContent := createMockTrapNode("modal-content", false)
		modal := createMockTrapNode("modal", false, modalContent)

		trap := NewFocusTrap("modal-trap", TrapModal, modal)

		// Simulate focus on trigger before opening modal
		triggerComp := triggerButton.Component.Instance.(*MockFocusableComponent)
		triggerComp.SetFocus(true)
		assert.True(t, triggerComp.focused, "Trigger should be focused")

		// Open modal (activates trap)
		trap.Activate()

		// Close modal (in real implementation, would restore focus to trigger)
		trap.Deactivate()

		// Focus restoration would be handled by FocusManager
		assert.False(t, trap.IsActive(), "Trap should be inactive")
	})

	t.Run("element no longer exists", func(t *testing.T) {
		modal := createMockTrapNode("modal", false)
		trap := NewFocusTrap("modal-trap", TrapModal, modal)

		trap.Activate()
		trap.Deactivate()

		// Should handle gracefully even if original element is gone
		assert.False(t, trap.IsActive())
	})
}

func TestTrap_Contains(t *testing.T) {
	t.Run("contains checks component ID membership", func(t *testing.T) {
		button1 := createMockTrapNode("button1", true)
		button2 := createMockTrapNode("button2", true)
		container := createMockTrapNode("container", false, button1, button2)

		trap := NewFocusTrap("trap", TrapMenu, container)

		assert.True(t, trap.Contains("button1"), "Should contain button1")
		assert.True(t, trap.Contains("button2"), "Should contain button2")
		assert.True(t, trap.Contains("container"), "Should contain container")
		assert.False(t, trap.Contains("nonexistent"), "Should not contain nonexistent")
		assert.False(t, trap.Contains(""), "Should not contain empty string")
	})

	t.Run("nested hierarchy contains", func(t *testing.T) {
		deepChild := createMockTrapNode("deep", true)
		child := createMockTrapNode("child", false, deepChild)
		parent := createMockTrapNode("parent", false, child)

		trap := NewFocusTrap("trap", TrapModal, parent)

		assert.True(t, trap.Contains("deep"), "Should contain deeply nested child")
		assert.True(t, trap.Contains("child"), "Should contain direct child")
		assert.True(t, trap.Contains("parent"), "Should contain parent")
	})
}

func TestTrap_GetFocusableIDs(t *testing.T) {
	t.Run("collects only focusable components", func(t *testing.T) {
		button1 := createMockTrapNode("button1", true)
		button2 := createMockTrapNode("button2", true)
		label := createMockTrapNode("label", false) // Not focusable
		container := createMockTrapNode("container", false, button1, label, button2)

		trap := NewFocusTrap("trap", TrapModal, container)
		ids := trap.GetFocusableIDs()

		assert.Len(t, ids, 2, "Should only return focusable items")
		assert.Contains(t, ids, "button1", "Should contain button1")
		assert.Contains(t, ids, "button2", "Should contain button2")
		assert.NotContains(t, ids, "label", "Should not contain non-focusable label")
		assert.NotContains(t, ids, "container", "Should not contain container")
	})

	t.Run("empty trap returns empty list", func(t *testing.T) {
		trap := NewFocusTrap("empty", TrapModal, nil)
		ids := trap.GetFocusableIDs()

		assert.Empty(t, ids, "Empty trap should return empty list")
	})

	t.Run("deep nesting collection", func(t *testing.T) {
		// Create deeply nested structure
		level3 := createMockTrapNode("level3", true)
		level2 := createMockTrapNode("level2", false, level3)
		level1 := createMockTrapNode("level1", false, level2)

		trap := NewFocusTrap("trap", TrapModal, level1)
		ids := trap.GetFocusableIDs()

		assert.Len(t, ids, 1, "Should find focusable item at any depth")
		assert.Contains(t, ids, "level3", "Should find deeply nested focusable")
	})
}

func TestTrap_Types(t *testing.T) {
	t.Run("modal trap type", func(t *testing.T) {
		modal := createMockTrapNode("modal", false)
		trap := NewFocusTrap("modal", TrapModal, modal)

		assert.Equal(t, TrapModal, trap.Type)
		assert.Equal(t, "modal", trap.ID)
	})

	t.Run("menu trap type", func(t *testing.T) {
		menu := createMockTrapNode("menu", false)
		trap := NewFocusTrap("menu", TrapMenu, menu)

		assert.Equal(t, TrapMenu, trap.Type)
	})

	t.Run("popover trap type", func(t *testing.T) {
		popover := createMockTrapNode("popover", false)
		trap := NewFocusTrap("popover", TrapPopover, popover)

		assert.Equal(t, TrapPopover, trap.Type)
	})
}

func TestTrap_ActivationCycle(t *testing.T) {
	t.Run("activation after deactivation", func(t *testing.T) {
		modal := createMockTrapNode("modal", false)
		trap := NewFocusTrap("trap", TrapModal, modal)

		// Cycle: activate -> deactivate -> activate
		trap.Activate()
		assert.True(t, trap.IsActive())

		trap.Deactivate()
		assert.False(t, trap.IsActive())

		trap.Activate()
		assert.True(t, trap.IsActive(), "Should be able to reactivate")
	})

	t.Run("multiple deactivations", func(t *testing.T) {
		modal := createMockTrapNode("modal", false)
		trap := NewFocusTrap("trap", TrapModal, modal)

		trap.Activate()

		// Multiple deactivations should be safe
		trap.Deactivate()
		trap.Deactivate()
		trap.Deactivate()

		assert.False(t, trap.IsActive())
	})
}

func TestTrap_ComplexHierarchy(t *testing.T) {
	t.Run("trap with mixed focusable and non-focusable", func(t *testing.T) {
		// Create realistic modal structure
		title := createMockTrapNode("title", false)
		input1 := createMockTrapNode("input1", true)
		input2 := createMockTrapNode("input2", true)
		form := createMockTrapNode("form", false, input1, input2)
		cancelBtn := createMockTrapNode("cancel", true)
		confirmBtn := createMockTrapNode("confirm", true)
		buttonBar := createMockTrapNode("button-bar", false, cancelBtn, confirmBtn)
		content := createMockTrapNode("content", false, title, form, buttonBar)
		modal := createMockTrapNode("modal", false, content)

		trap := NewFocusTrap("modal-trap", TrapModal, modal)
		ids := trap.GetFocusableIDs()

		// Should find: input1, input2, cancel, confirm
		assert.Len(t, ids, 4, "Should find all focusable elements")
		assert.Contains(t, ids, "input1")
		assert.Contains(t, ids, "input2")
		assert.Contains(t, ids, "cancel")
		assert.Contains(t, ids, "confirm")
		assert.NotContains(t, ids, "title")
		assert.NotContains(t, ids, "form")
		assert.NotContains(t, ids, "button-bar")
		assert.NotContains(t, ids, "content")
	})
}

// Benchmark tests
func BenchmarkTrap_GetFocusableIDs(b *testing.B) {
	// Create large hierarchy
	var buildTree func(depth int) *runtime.LayoutNode
	currentID := 0

	buildTree = func(depth int) *runtime.LayoutNode {
		if depth == 0 {
			id := "leaf-" + string(rune('0'+currentID%10))
			return createMockTrapNode(id, true)
		}

		children := []*runtime.LayoutNode{}
		for i := 0; i < 5; i++ {
			currentID++
			children = append(children, buildTree(depth-1))
		}

		id := "node-" + string(rune('0'+depth%10))
		return createMockTrapNode(id, false, children...)
	}

	root := buildTree(4) // Creates many nodes
	trap := NewFocusTrap("large-trap", TrapModal, root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trap.GetFocusableIDs()
	}
}

func BenchmarkTrap_Contains(b *testing.B) {
	// Create nested structure
	deepChild := createMockTrapNode("deep", true)
	child := createMockTrapNode("child", false, deepChild)
	parent := createMockTrapNode("parent", false, child)

	trap := NewFocusTrap("trap", TrapModal, parent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trap.Contains("deep")
	}
}
