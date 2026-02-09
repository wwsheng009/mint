package render

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestVNodeHookBasic tests basic hook functionality
func TestVNodeHookBasic(t *testing.T) {
	hm := NewHookManager()

	// Create a base VNode
	base := rtui.NewElement("base")

	// Register a hook that adds a prop
	hm.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		props := vnode.Props()
		if props == nil {
			vnode.SetProps(make(rtui.Props))
		}
		vnode.Props().Set("hooked", true)
		return vnode
	})

	// Apply hook
	result := hm.ApplyVNodeHooks(base)

	// Verify hook was applied
	if result.Props()["hooked"] != true {
		t.Error("Hook was not applied")
	}
}

// TestMultipleHooks tests that multiple hooks are applied in order
func TestMultipleHooks(t *testing.T) {
	hm := NewHookManager()

	base := rtui.NewElement("base")

	// Register multiple hooks
	hm.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		props := vnode.Props()
		if props == nil {
			vnode.SetProps(make(rtui.Props))
		}
		vnode.Props().Set("hook1", true)
		return vnode
	})

	hm.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		props := vnode.Props()
		if props == nil {
			vnode.SetProps(make(rtui.Props))
		}
		vnode.Props().Set("hook2", true)
		return vnode
	})

	result := hm.ApplyVNodeHooks(base)

	// Both hooks should be applied
	if result.Props()["hook1"] != true {
		t.Error("Hook1 was not applied")
	}
	if result.Props()["hook2"] != true {
		t.Error("Hook2 was not applied")
	}
}

// TestLIFOOrder tests that hooks are applied in LIFO order
func TestLIFOOrder(t *testing.T) {
	hm := NewHookManager()

	base := rtui.NewElement("base")

	callOrder := []int{}

	// Register hooks in order 1, 2, 3
	hm.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		callOrder = append(callOrder, 1)
		return vnode
	})

	hm.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		callOrder = append(callOrder, 2)
		return vnode
	})

	hm.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		callOrder = append(callOrder, 3)
		return vnode
	})

	// Apply hooks
	hm.ApplyVNodeHooks(base)

	// Should be called in reverse order: 3, 2, 1
	expected := []int{3, 2, 1}
	if len(callOrder) != len(expected) {
		t.Fatalf("Expected %d calls, got %d", len(expected), len(callOrder))
	}

	for i, v := range expected {
		if callOrder[i] != v {
			t.Errorf("Call order[%d] = %d, want %d", i, callOrder[i], v)
		}
	}
}

// TestFragmentWrapping tests that hooks can wrap VNodes in Fragments
func TestFragmentWrapping(t *testing.T) {
	hm := NewHookManager()

	base := rtui.NewElement("base")
	overlay := rtui.NewElement("overlay")

	// Register a hook that wraps in Fragment
	hm.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		return rtui.Fragment(vnode, overlay)
	})

	result := hm.ApplyVNodeHooks(base)

	// Result should be a Fragment
	fragment, ok := result.(*rtui.FragmentVNode)
	if !ok {
		t.Fatalf("Expected FragmentVNode, got %T", result)
	}

	// Fragment should have 2 children
	children := fragment.Children()
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

// TestNilHookManager tests that nil HookManager works safely
func TestNilHookManager(t *testing.T) {
	var hm *HookManager
	base := rtui.NewElement("base")

	// Should not panic
	result := hm.ApplyVNodeHooks(base)

	if result != base {
		t.Error("Nil HookManager should return original VNode")
	}

	count := hm.VNodeHookCount()
	if count != 0 {
		t.Errorf("Nil HookManager should return 0 hooks, got %d", count)
	}
}

// TestClearHooks tests clearing all hooks
func TestClearHooks(t *testing.T) {
	hm := NewHookManager()

	base := rtui.NewElement("base")

	// Register some hooks
	hm.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		return vnode
	})

	hm.RegisterVNodeHook(func(vnode rtui.VNode) rtui.VNode {
		return vnode
	})

	if hm.VNodeHookCount() != 2 {
		t.Errorf("Expected 2 hooks, got %d", hm.VNodeHookCount())
	}

	// Clear hooks
	hm.ClearVNodeHooks()

	if hm.VNodeHookCount() != 0 {
		t.Errorf("Expected 0 hooks after clear, got %d", hm.VNodeHookCount())
	}

	// Apply hooks after clear - should return original
	result := hm.ApplyVNodeHooks(base)
	if result != base {
		t.Error("Should return original VNode after clearing hooks")
	}
}

// TestNilHook tests that nil hooks are ignored
func TestNilHook(t *testing.T) {
	hm := NewHookManager()

	// Register nil hook - should be ignored
	hm.RegisterVNodeHook(nil)

	if hm.VNodeHookCount() != 0 {
		t.Errorf("Nil hook should not be registered, got %d", hm.VNodeHookCount())
	}
}
