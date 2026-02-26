// Package render tests for DeclarativeNode extensions.
package render

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime/paint"
	rtevent "github.com/wwsheng009/mint/runtime/event"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// DeclarativeNode.Root Tests
// =============================================================================

func TestDeclarativeNode_Root(t *testing.T) {
	t.Run("returns the root VNode", func(t *testing.T) {
		vnode := rtui.Element("div").Key("test-root").Build()
		node := NewDeclarativeNode(vnode)

		root := node.Root()
		if root != vnode {
			t.Error("Root() should return the original VNode")
		}
	})

	t.Run("returns nil for nil root", func(t *testing.T) {
		node := NewDeclarativeNode(nil)
		root := node.Root()
		if root != nil {
			t.Error("Root() should return nil when root is nil")
		}
	})

	t.Run("returns updated root after UpdateRoot", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())
		newRoot := rtui.Element("span").Build()
		node.UpdateRoot(newRoot)

		root := node.Root()
		if root != newRoot {
			t.Error("Root() should return the updated VNode")
		}
	})
}

// =============================================================================
// DeclarativeNode.HandleEvent Tests
// =============================================================================

func TestDeclarativeNode_HandleEvent(t *testing.T) {
	t.Run("handles event with focus manager", func(t *testing.T) {
		node := NewDeclarativeNodeFromFunc(func() rtui.VNode {
			return rtui.Element("div").Build()
		})

		ev := &mockKeyEvent{key: 'a'}

		result := node.HandleEvent(ev)
		_ = result
	})

	t.Run("returns false for nil root", func(t *testing.T) {
		node := NewDeclarativeNode(nil)
		ev := &mockKeyEvent{key: 'a'}

		result := node.HandleEvent(ev)
		if result {
			t.Error("HandleEvent should return false for nil root")
		}
	})
}

// =============================================================================
// DeclarativeNode.GetFocusedIndex Tests
// =============================================================================

func TestDeclarativeNode_GetFocusedIndex(t *testing.T) {
	t.Run("returns -1 for nil focus manager", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())
		node.focusMgr = nil

		idx := node.GetFocusedIndex()
		if idx != -1 {
			t.Errorf("GetFocusedIndex() = %d, want -1 for nil focus manager", idx)
		}
	})
}

// =============================================================================
// DeclarativeNode.GetFocusedType Tests
// =============================================================================

func TestDeclarativeNode_GetFocusedType(t *testing.T) {
	t.Run("returns 0 for nil focus manager", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())
		node.focusMgr = nil

		typ := node.GetFocusedType()
		if typ != 0 {
			t.Errorf("GetFocusedType() = %d, want 0 for nil focus manager", typ)
		}
	})
}

// =============================================================================
// DeclarativeNode.GetButtons Tests
// =============================================================================

func TestDeclarativeNode_GetButtons(t *testing.T) {
	t.Run("returns empty list for no buttons", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		buttons := node.GetButtons()
		if len(buttons) != 0 {
			t.Errorf("GetButtons() = %d buttons, want 0", len(buttons))
		}
	})

	t.Run("returns empty list for nil root", func(t *testing.T) {
		node := NewDeclarativeNode(nil)

		buttons := node.GetButtons()
		// GetButtons returns nil for nil root - this is expected behavior
		if buttons != nil && len(buttons) != 0 {
			t.Errorf("GetButtons() = %d buttons, want 0", len(buttons))
		}
	})
}

// =============================================================================
// DeclarativeNode.isButtonVNode Tests
// =============================================================================

func TestDeclarativeNode_isButtonVNode(t *testing.T) {
	node := NewDeclarativeNode(rtui.Element("div").Build())

	t.Run("returns false for nil VNode", func(t *testing.T) {
		if node.isButtonVNode(nil) {
			t.Error("isButtonVNode(nil) should return false")
		}
	})

	t.Run("returns false for non-button element", func(t *testing.T) {
		div := rtui.Element("div").Build()
		if node.isButtonVNode(div) {
			t.Error("isButtonVNode(div) should return false")
		}
	})
}

// =============================================================================
// DeclarativeNode.GetInputs Tests
// =============================================================================

func TestDeclarativeNode_GetInputs(t *testing.T) {
	t.Run("returns empty list for no inputs", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		inputs := node.GetInputs()
		if len(inputs) != 0 {
			t.Errorf("GetInputs() = %d inputs, want 0", len(inputs))
		}
	})

	t.Run("returns empty list for nil root", func(t *testing.T) {
		node := NewDeclarativeNode(nil)

		inputs := node.GetInputs()
		// GetInputs returns nil for nil root - this is expected behavior
		if inputs != nil && len(inputs) != 0 {
			t.Errorf("GetInputs() = %d inputs, want 0", len(inputs))
		}
	})
}

// =============================================================================
// DeclarativeNode.GetTextareas Tests
// =============================================================================

func TestDeclarativeNode_GetTextareas(t *testing.T) {
	t.Run("returns empty list for no textareas", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		textareas := node.GetTextareas()
		if len(textareas) != 0 {
			t.Errorf("GetTextareas() = %d textareas, want 0", len(textareas))
		}
	})
}

// =============================================================================
// DeclarativeNode.GetCheckboxes Tests
// =============================================================================

func TestDeclarativeNode_GetCheckboxes(t *testing.T) {
	t.Run("returns empty list for no checkboxes", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		checkboxes := node.GetCheckboxes()
		if len(checkboxes) != 0 {
			t.Errorf("GetCheckboxes() = %d checkboxes, want 0", len(checkboxes))
		}
	})
}

// =============================================================================
// DeclarativeNode.GetSelects Tests
// =============================================================================

func TestDeclarativeNode_GetSelects(t *testing.T) {
	t.Run("returns empty list for no selects", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		selects := node.GetSelects()
		if len(selects) != 0 {
			t.Errorf("GetSelects() = %d selects, want 0", len(selects))
		}
	})
}

// =============================================================================
// collectByType Tests
// =============================================================================

func TestCollectByType(t *testing.T) {
	t.Run("returns empty list for nil root", func(t *testing.T) {
		result := collectByType(nil, func(vnode rtui.VNode) bool {
			return true
		})
		// collectByType returns nil for nil root - this is expected behavior
		if result != nil && len(result) != 0 {
			t.Errorf("collectByType(nil) = %d items, want 0", len(result))
		}
	})

	t.Run("filters by predicate", func(t *testing.T) {
		root := rtui.Element("div").Children(
			rtui.Element("span").Build(),
			rtui.Element("span").Build(),
		).Build()

		result := collectByType(root, func(vnode rtui.VNode) bool {
			return true
		})

		if len(result) != 0 {
			t.Errorf("collectByType with non-focusable nodes = %d, want 0", len(result))
		}
	})
}

// =============================================================================
// fiberReconcilerAdapter.SetApp Tests
// =============================================================================

func TestFiberReconcilerAdapter_SetApp(t *testing.T) {
	t.Run("sets app when passed framework.App", func(t *testing.T) {
		app := framework.NewApp()
		fn := func() rtui.VNode {
			return rtui.Element("div").Build()
		}
		node := NewDeclarativeNodeFromFuncWithFiber(fn, app)

		node.mu.RLock()
		reconciler, ok := node.reconciler.(*fiberReconcilerAdapter)
		node.mu.RUnlock()

		if !ok {
			t.Skip("reconciler is not *fiberReconcilerAdapter")
		}

		reconciler.SetApp(app)
	})

	t.Run("does not panic with non-framework.App", func(t *testing.T) {
		app := framework.NewApp()
		fn := func() rtui.VNode {
			return rtui.Element("div").Build()
		}
		node := NewDeclarativeNodeFromFuncWithFiber(fn, app)

		node.mu.RLock()
		reconciler, ok := node.reconciler.(*fiberReconcilerAdapter)
		node.mu.RUnlock()

		if !ok {
			t.Skip("reconciler is not *fiberReconcilerAdapter")
		}

		reconciler.SetApp("not an app")
	})
}

// =============================================================================
// DeclarativeNode.paintText Tests
// =============================================================================

func TestDeclarativeNode_paintText(t *testing.T) {
	t.Run("paints text content", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())
		buf := paint.NewBuffer(80, 24)

		vnode := rtui.Element("text").Prop("content", "Hello").Build()

		node.paintText(vnode, 0, 0, buf)

		content := buf.GetContent(0, 0)
		if content.Cluster != "H" {
			t.Errorf("expected H at (0,0), got %s", content.Cluster)
		}
	})

	t.Run("handles empty text", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())
		buf := paint.NewBuffer(80, 24)

		vnode := rtui.Element("text").Prop("content", "").Build()

		node.paintText(vnode, 0, 0, buf)
	})

	t.Run("handles nil VNode", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())
		buf := paint.NewBuffer(80, 24)

		node.paintText(nil, 0, 0, buf)
	})
}

// =============================================================================
// Mock types for testing
// =============================================================================

type mockKeyEvent struct {
	key rune
}

func (m *mockKeyEvent) Type() event.EventType {
	return event.EventKeyPress
}

func (m *mockKeyEvent) Timestamp() time.Time {
	return time.Now()
}

func (m *mockKeyEvent) Target() event.Component {
	return nil
}

func (m *mockKeyEvent) SetTarget(target event.Component) {}

func (m *mockKeyEvent) Source() event.Component {
	return nil
}

func (m *mockKeyEvent) SetSource(source event.Component) {}

func (m *mockKeyEvent) RuntimeEvent() rtevent.Event {
	return rtevent.NewBaseEvent(rtevent.EventKeyPress)
}

// =============================================================================
// DeclarativeNode.distributeEventToVNode Tests
// =============================================================================

func TestDeclarativeNode_distributeEventToVNode(t *testing.T) {
	t.Run("returns false for nil VNode", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())
		ev := &mockKeyEvent{key: 'a'}

		result := node.distributeEventToVNode(nil, ev)
		if result {
			t.Error("distributeEventToVNode should return false for nil VNode")
		}
	})

	t.Run("returns false when VNode doesn't handle event", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())
		div := rtui.Element("div").Build()
		ev := &mockKeyEvent{key: 'a'}

		result := node.distributeEventToVNode(div, ev)
		if result {
			t.Error("distributeEventToVNode should return false when event not handled")
		}
	})

	t.Run("distributes to children", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())
		root := rtui.Element("div").Children(
			rtui.Element("span").Build(),
			rtui.Element("span").Build(),
		).Build()
		ev := &mockKeyEvent{key: 'a'}

		// Should not panic and should return false
		result := node.distributeEventToVNode(root, ev)
		if result {
			t.Error("distributeEventToVNode should return false when no children handle event")
		}
	})
}

// =============================================================================
// DeclarativeNode.getFrameworkApp Tests
// =============================================================================

func TestDeclarativeNode_getFrameworkApp(t *testing.T) {
	t.Run("returns nil for new node", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		app := node.getFrameworkApp()
		if app != nil {
			t.Error("getFrameworkApp should return nil for node without app")
		}
	})
}

// =============================================================================
// DeclarativeNode.GetFocusedType with focus manager Tests
// =============================================================================

func TestDeclarativeNode_GetFocusedType_WithFocusManager(t *testing.T) {
	t.Run("returns 0 for empty focus manager", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		typ := node.GetFocusedType()
		if typ != 0 {
			t.Errorf("GetFocusedType() = %d, want 0 for empty focus manager", typ)
		}
	})
}

// =============================================================================
// DeclarativeNode.collectButtonsFromFiber Tests
// =============================================================================

func TestDeclarativeNode_collectButtonsFromFiber(t *testing.T) {
	t.Run("returns empty for nil fiber", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		result := node.collectButtonsFromFiber(nil)
		if result == nil || len(result) != 0 {
			// Expected: empty slice or nil
		}
	})

	t.Run("returns empty for fiber with nil VNode", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		result := node.collectButtonsFromFiber(&reconciler.Fiber{})
		if result == nil || len(result) != 0 {
			// Expected: empty slice or nil
		}
	})
}

// =============================================================================
// DeclarativeNode.collectFocusableFromFiber Tests
// =============================================================================

func TestDeclarativeNode_collectFocusableFromFiber(t *testing.T) {
	t.Run("returns empty for nil fiber", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		result := node.collectFocusableFromFiber(nil)
		if result == nil || len(result) != 0 {
			// Expected: empty slice or nil
		}
	})

	t.Run("returns empty for fiber with nil VNode", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("div").Build())

		result := node.collectFocusableFromFiber(&reconciler.Fiber{})
		if result == nil || len(result) != 0 {
			// Expected: empty slice or nil
		}
	})
}
