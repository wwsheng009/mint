// Package inspector provides tests for framework integration
package inspector

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/render"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestFrameworkIntegration tests that Inspector integrates with framework via hooks
func TestFrameworkIntegration(t *testing.T) {
	// Create a hook manager (simulating framework)
	hooks := render.NewHookManager()

	// Create and enable inspector
	inspector := NewStandaloneInspector()
	inspector.Enable()
	inspector.ToggleVisibility()

	// Register inspector hook (simulating framework.SetInspector)
	RegisterInspector(inspector, hooks)

	// Verify hook was registered
	if hooks.VNodeHookCount() != 1 {
		t.Fatalf("Expected 1 hook registered, got %d", hooks.VNodeHookCount())
	}

	// Apply hooks to a test VNode (simulating render)
	testVNode := rtui.NewElement("app")
	result := hooks.ApplyVNodeHooks(testVNode)

	// Should return Fragment with inspector overlay
	fragment, ok := result.(*rtui.FragmentVNode)
	if !ok {
		t.Fatalf("Expected FragmentVNode, got %T", result)
	}

	children := fragment.Children()
	if len(children) != 2 {
		t.Fatalf("Expected 2 children in Fragment, got %d", len(children))
	}

	// Verify inspector overlay has LayerInspector
	inspectorOverlay := children[1]
	if inspectorOverlay.GetLayer() != rtui.LayerInspector {
		t.Errorf("Expected LayerInspector (%d), got %d",
			rtui.LayerInspector, inspectorOverlay.GetLayer())
	}

	// Verify inspector is positioned
	props := inspectorOverlay.Props()
	if props == nil {
		t.Fatal("Inspector overlay should have props")
	}

	if _, hasX := props["x"].(int); !hasX {
		t.Error("Inspector should have x position prop")
	}

	if _, hasY := props["y"].(int); !hasY {
		t.Error("Inspector should have y position prop")
	}

	if _, hasWidth := props["width"].(int); !hasWidth {
		t.Error("Inspector should have width prop")
	}

	if _, hasHeight := props["height"].(int); !hasHeight {
		t.Error("Inspector should have height prop")
	}

	t.Log("✅ Framework integration successful")
}

// TestHookRegistrarInterface tests that Inspector implements HookRegistrar
func TestHookRegistrarInterface(t *testing.T) {
	inspector := NewStandaloneInspector()

	// Verify Inspector implements HookRegistrar interface
	var _ interface{ RegisterWithHookManager(interface{}) } = inspector

	hooks := render.NewHookManager()

	// Call RegisterWithHookManager (should not panic)
	inspector.RegisterWithHookManager(hooks)

	if hooks.VNodeHookCount() != 1 {
		t.Errorf("Expected 1 hook registered, got %d", hooks.VNodeHookCount())
	}
}
