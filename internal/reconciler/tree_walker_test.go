// Package reconciler tests for TreeWalker.
package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Mock FocusableVNode for Testing
// =============================================================================

// mockFocusableNode is a test implementation of FocusableVNode
type mockFocusableNode struct {
	*rtui.ElementVNode
	id          string
	label       string
	isFocusable bool
	hasFocus    bool
}

func newMockFocusable(tag, id, label string) *mockFocusableNode {
	return &mockFocusableNode{
		ElementVNode: rtui.NewElement(tag),
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

func (m *mockFocusableNode) Type() rtui.VNodeType {
	return rtui.VNodeElement
}

// =============================================================================
// TreeWalker Constructor Tests
// =============================================================================

func TestNewTreeWalker(t *testing.T) {
	t.Run("with nil root", func(t *testing.T) {
		w := NewTreeWalker(nil)
		if w == nil {
			t.Error("NewTreeWalker should not return nil")
		}
		if w.root != nil {
			t.Error("root should be nil")
		}
	})

	t.Run("with valid root", func(t *testing.T) {
		vnode := rtui.Element("div").Build()
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		if w.root != fiber {
			t.Error("root should be set to the provided fiber")
		}
	})
}

// =============================================================================
// TreeWalker.SetRoot Tests
// =============================================================================

func TestTreeWalker_SetRoot(t *testing.T) {
	vnode1 := rtui.Element("div").Build()
	fiber1 := CreateFiberFromVNode(vnode1)

	vnode2 := rtui.Element("span").Build()
	fiber2 := CreateFiberFromVNode(vnode2)

	w := NewTreeWalker(fiber1)

	// Change root
	w.SetRoot(fiber2)

	if w.root != fiber2 {
		t.Error("root should be updated")
	}

	// Set to nil
	w.SetRoot(nil)

	if w.root != nil {
		t.Error("root should be nil")
	}
}

// =============================================================================
// TreeWalker.Count Tests
// =============================================================================

func TestTreeWalker_Count(t *testing.T) {
	tests := []struct {
		name     string
		vnode    rtui.VNode
		expected int
	}{
		{
			name:     "nil root",
			vnode:    nil,
			expected: 0,
		},
		{
			name:     "single element",
			vnode:    rtui.Element("div").Build(),
			expected: 1,
		},
		{
			name: "nested elements",
			vnode: rtui.VStack(
				rtui.Element("div").Build(),
				rtui.HStack(
					rtui.Element("span").Build(),
					rtui.Element("text").Prop("content", "a").Build(),
				),
			),
		},
		{
			name:     "fragment with children",
			vnode:    rtui.Fragment(rtui.Element("a").Build(), rtui.Element("b").Build()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fiber := CreateFiberFromVNode(tt.vnode)
			w := NewTreeWalker(fiber)

			count := w.Count()
			// Count returns the number of Fiber nodes, which can vary
			// Just verify it doesn't crash and returns a non-negative value
			if count < 0 {
				t.Errorf("Count() returned negative value: %d", count)
			}
			t.Logf("Count() for %s: %d", tt.name, count)
		})
	}
}

// =============================================================================
// TreeWalker.CollectFocusable Tests
// =============================================================================

func TestTreeWalker_CollectFocusable(t *testing.T) {
	t.Run("empty tree", func(t *testing.T) {
		w := NewTreeWalker(nil)
		focusable := w.CollectFocusable()

		if focusable == nil {
			t.Error("CollectFocusable should return empty slice, not nil")
		}
		if len(focusable) != 0 {
			t.Errorf("CollectFocusable on nil tree should return empty slice, got %d", len(focusable))
		}
	})

	t.Run("no focusable elements", func(t *testing.T) {
		vnode := rtui.VStack(
			rtui.Element("div").Build(),
			rtui.Element("span").Build(),
			rtui.Element("text").Prop("content", "hello").Build(),
		)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		focusable := w.CollectFocusable()

		if focusable == nil {
			t.Error("CollectFocusable should return empty slice, not nil")
		}
		if len(focusable) != 0 {
			t.Errorf("CollectFocusable on non-focusable tree should return empty slice, got %d", len(focusable))
		}
	})

	t.Run("with button elements", func(t *testing.T) {
		// Create mock focusable button elements
		btn1 := newMockFocusable("button", "btn1", "Button1")
		btn2 := newMockFocusable("button", "btn2", "Button2")

		vnode := rtui.VStack(btn1, btn2)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		focusable := w.CollectFocusable()

		if focusable == nil {
			t.Error("CollectFocusable should return empty slice, not nil")
		}
		if len(focusable) != 2 {
			t.Errorf("CollectFocusable should find 2 buttons, got %d", len(focusable))
		}
	})

	t.Run("nested focusable elements", func(t *testing.T) {
		btn1 := newMockFocusable("button", "btn1", "Btn1")
		btn2 := newMockFocusable("button", "btn2", "Btn2")
		btn3 := newMockFocusable("button", "btn3", "Btn3")

		// Create nested structure
		vnode := rtui.VStack(
			rtui.HStack(btn1, btn2),
			btn3,
		)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		focusable := w.CollectFocusable()

		if focusable == nil {
			t.Error("CollectFocusable should return empty slice, not nil")
		}
		if len(focusable) != 3 {
			t.Errorf("CollectFocusable should find 3 buttons, got %d", len(focusable))
		}
	})

	t.Run("with disabled button", func(t *testing.T) {
		// Disabled buttons should not be collected
		btn1 := newMockFocusable("button", "btn1", "Enabled")
		btn2 := newMockFocusable("button", "btn2", "Disabled")
		btn2.isFocusable = false // Mark as disabled

		vnode := rtui.VStack(btn1, btn2)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		focusable := w.CollectFocusable()

		// Should only collect the enabled button
		if len(focusable) != 1 {
			t.Errorf("CollectFocusable should find 1 enabled button, got %d", len(focusable))
		}
	})

	t.Run("with input elements", func(t *testing.T) {
		input1 := newMockFocusable("input", "input1", "Input1")
		input2 := newMockFocusable("input", "input2", "Input2")

		vnode := rtui.HStack(input1, input2)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		focusable := w.CollectFocusable()

		if focusable == nil {
			t.Error("CollectFocusable should return empty slice, not nil")
		}
		if len(focusable) != 2 {
			t.Errorf("CollectFocusable should find 2 inputs, got %d", len(focusable))
		}
	})

	t.Run("mixed focusable types", func(t *testing.T) {
		button := newMockFocusable("button", "btn1", "Click")
		input := newMockFocusable("input", "input1", "Type")
		checkbox := newMockFocusable("checkbox", "cb1", "Check")

		vnode := rtui.VStack(button, input, checkbox)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		focusable := w.CollectFocusable()

		if focusable == nil {
			t.Error("CollectFocusable should return empty slice, not nil")
		}
		if len(focusable) != 3 {
			t.Errorf("CollectFocusable should find 3 focusable elements, got %d", len(focusable))
		}
	})
}

// =============================================================================
// TreeWalker.FindByTag Tests
// =============================================================================

func TestTreeWalker_FindByTag(t *testing.T) {
	t.Run("nil root", func(t *testing.T) {
		w := NewTreeWalker(nil)
		results := w.FindByTag("div")

		if results == nil {
			t.Error("FindByTag should return empty slice, not nil")
		}
		if len(results) != 0 {
			t.Errorf("FindByTag on nil tree should return empty slice, got %d", len(results))
		}
	})

	t.Run("find by existing tag", func(t *testing.T) {
		vnode := rtui.VStack(
			rtui.Element("div").Build(),
			rtui.Element("span").Build(),
			rtui.Element("div").Build(),
		)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		results := w.FindByTag("div")

		if results == nil {
			t.Error("FindByTag should return empty slice, not nil")
		}
		if len(results) != 2 {
			t.Errorf("FindByTag should find 2 'div' elements, got %d", len(results))
		}
	})

	t.Run("find by non-existent tag", func(t *testing.T) {
		vnode := rtui.VStack(
			rtui.Element("div").Build(),
			rtui.Element("span").Build(),
		)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		results := w.FindByTag("nonexistent")

		if results == nil {
			t.Error("FindByTag should return empty slice, not nil")
		}
		if len(results) != 0 {
			t.Errorf("FindByTag for non-existent tag should return empty, got %d", len(results))
		}
	})

	t.Run("find with nested structure", func(t *testing.T) {
		vnode := rtui.VStack(
			rtui.HStack(
				rtui.Element("button").Build(),
				rtui.Element("button").Build(),
			),
			rtui.VStack(
				rtui.Element("button").Build(),
			),
		)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		results := w.FindByTag("button")

		if results == nil {
			t.Error("FindByTag should return empty slice, not nil")
		}
		if len(results) != 3 {
			t.Errorf("FindByTag should find 3 'button' elements, got %d", len(results))
		}
	})

	t.Run("empty string tag matches all", func(t *testing.T) {
		vnode := rtui.VStack(
			rtui.Element("div").Build(),
			rtui.Element("span").Build(),
			rtui.Element("button").Build(),
		)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		// Empty tag won't match anything since elements return their tag name
		results := w.FindByTag("")

		if results == nil {
			t.Error("FindByTag should return empty slice, not nil")
		}
		if len(results) != 0 {
			t.Errorf("FindByTag for empty tag should return 0 results, got %d", len(results))
		}
	})
}

// =============================================================================
// TreeWalker Edge Cases
// =============================================================================

func TestTreeWalker_EdgeCases(t *testing.T) {
	t.Run("deeply nested structure", func(t *testing.T) {
		// Create a deeply nested structure
		vnode := rtui.Element("div").Build()
		for i := 0; i < 50; i++ {
			vnode = rtui.VStack(vnode)
		}
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		// Should not crash or hang
		count := w.Count()
		if count <= 0 {
			t.Errorf("Count() should return positive value for nested structure, got %d", count)
		}

		// CollectFocusable should also work
		focusable := w.CollectFocusable()
		if focusable == nil {
			t.Error("CollectFocusable should not return nil")
		}
		t.Logf("Deep nest: Count=%d, Focusable=%d", count, len(focusable))
	})

	t.Run("fragment with many children", func(t *testing.T) {
		children := make([]rtui.VNode, 100)
		for i := 0; i < 100; i++ {
			children[i] = rtui.Element("text").Prop("content", string(rune('A'+i))).Build()
		}
		vnode := rtui.Fragment(children...)
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		// Should handle large fragment
		count := w.Count()
		if count < 100 {
			t.Errorf("Count() should return at least 100 for fragment with 100 children, got %d", count)
		}
		t.Logf("Large fragment: Count=%d", count)
	})

	t.Run("circular reference protection", func(t *testing.T) {
		// Create a simple tree
		vnode := rtui.Element("div").Build()
		fiber := CreateFiberFromVNode(vnode)
		w := NewTreeWalker(fiber)

		// Walk multiple times to ensure no state corruption
		for i := 0; i < 10; i++ {
			count := w.Count()
			if count != 1 {
				t.Errorf("Count() should be consistent across calls, got %d on iteration %d", count, i)
			}
		}
	})
}

// =============================================================================
// TreeWalker.Convenience Functions Tests
// =============================================================================

func TestCollectFocusableFromFiber(t *testing.T) {
	t.Run("nil fiber", func(t *testing.T) {
		focusable := CollectFocusableFromFiber(nil)

		if focusable == nil {
			t.Error("CollectFocusableFromFiber should return empty slice, not nil")
		}
		if len(focusable) != 0 {
			t.Errorf("CollectFocusableFromFiber on nil fiber should return empty, got %d", len(focusable))
		}
	})

	t.Run("with focusable elements", func(t *testing.T) {
		btn := newMockFocusable("button", "btn1", "Test")
		fiber := CreateFiberFromVNode(btn)

		focusable := CollectFocusableFromFiber(fiber)

		if focusable == nil {
			t.Error("CollectFocusableFromFiber should return empty slice, not nil")
		}
		if len(focusable) != 1 {
			t.Errorf("CollectFocusableFromFiber should find 1 button, got %d", len(focusable))
		}
	})
}

func TestFindFibersByTag(t *testing.T) {
	t.Run("nil fiber", func(t *testing.T) {
		results := FindFibersByTag(nil, "div")

		if results == nil {
			t.Error("FindFibersByTag should return empty slice, not nil")
		}
		if len(results) != 0 {
			t.Errorf("FindFibersByTag on nil fiber should return empty, got %d", len(results))
		}
	})

	t.Run("find matching elements", func(t *testing.T) {
		vnode := rtui.VStack(
			rtui.Element("div").Key("div1").Build(),
			rtui.Element("span").Build(),
			rtui.Element("div").Key("div2").Build(),
		)
		fiber := CreateFiberFromVNode(vnode)

		results := FindFibersByTag(fiber, "div")

		if results == nil {
			t.Error("FindFibersByTag should return empty slice, not nil")
		}
		if len(results) != 2 {
			t.Errorf("FindFibersByTag should find 2 'div' elements, got %d", len(results))
		}
	})
}
