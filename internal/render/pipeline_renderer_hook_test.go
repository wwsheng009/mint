package render

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestPipelineRendererGetHooks tests GetHooks method
func TestPipelineRendererGetHooks(t *testing.T) {
	renderer := NewPipelineRenderer()

	hooks := renderer.GetHooks()
	if hooks == nil {
		t.Error("GetHooks should return non-nil HookManager")
	}

	if hooks.VNodeHookCount() != 0 {
		t.Errorf("New renderer should have 0 hooks, got %d", hooks.VNodeHookCount())
	}

	// Verify we can register hooks through GetHooks()
	called := false
	hooks.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		called = true
		return vnode
	})

	if hooks.VNodeHookCount() != 1 {
		t.Errorf("Expected 1 hook after registration, got %d", hooks.VNodeHookCount())
	}

	// Apply hooks manually to test
	vnode := rtui.NewElement("test")
	result := hooks.ApplyVNodeHooks(vnode)

	if !called {
		t.Error("Hook should have been called")
	}

	if result == nil {
		t.Error("ApplyVNodeHooks should return non-nil VNode")
	}
}

// TestPipelineRendererHookOrder tests that hooks are applied in LIFO order
func TestPipelineRendererHookOrder(t *testing.T) {
	renderer := NewPipelineRenderer()
	hooks := renderer.GetHooks()

	order := []int{}

	hooks.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		order = append(order, 1)
		return vnode
	})

	hooks.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		order = append(order, 2)
		return vnode
	})

	hooks.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		order = append(order, 3)
		return vnode
	})

	vnode := rtui.NewElement("test")
	hooks.ApplyVNodeHooks(vnode)

	// Should be called in reverse order: 3, 2, 1
	expected := []int{3, 2, 1}
	if len(order) != len(expected) {
		t.Fatalf("Expected %d hook calls, got %d", len(expected), len(order))
	}

	for i, v := range expected {
		if order[i] != v {
			t.Errorf("Hook call order[%d] = %d, want %d", i, order[i], v)
		}
	}
}
