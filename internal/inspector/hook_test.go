package inspector

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/render"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestCreateInspectorHook tests that CreateInspectorHook creates a valid hook
func TestCreateInspectorHook(t *testing.T) {
	inspector := NewStandaloneInspector()
	hook := CreateInspectorHook(inspector)

	if hook == nil {
		t.Fatal("CreateInspectorHook should return non-nil hook")
	}

	// Test with inspector not visible
	baseVNode := rtui.NewElement("base")
	result := hook(baseVNode)

	// Should return original VNode when inspector is not visible
	if result != baseVNode {
		t.Error("Hook should return original VNode when inspector is not visible")
	}
}

// TestInspectorHookWrapsInFragment tests that hook wraps VNode in Fragment when inspector is visible
func TestInspectorHookWrapsInFragment(t *testing.T) {
	inspector := NewStandaloneInspector()

	// Enable inspector
	inspector.mu.Lock()
	inspector.visible = true
	inspector.mu.Unlock()

	// Verify RenderContent works
	content := inspector.RenderContent()
	if content == nil {
		t.Fatal("RenderContent should return non-nil when inspector is visible")
	}

	t.Logf("RenderContent returned: %T", content)

	hook := CreateInspectorHook(inspector)

	baseVNode := rtui.NewElement("base")
	result := hook(baseVNode)

	t.Logf("Hook result type: %T", result)

	// Should return a Fragment
	fragment, ok := result.(*rtui.FragmentVNode)
	if !ok {
		t.Fatalf("Expected FragmentVNode, got %T", result)
	}

	// Fragment should have 2 children (base + inspector)
	children := fragment.Children()
	if len(children) != 2 {
		t.Fatalf("Expected 2 children in Fragment, got %d", len(children))
	}

	// First child should be the base VNode
	if children[0] != baseVNode {
		t.Error("First child should be the base VNode")
	}

	// Second child should be inspector overlay
	if children[1] == nil {
		t.Fatal("Second child should be inspector overlay")
	}

	// Inspector overlay should have LayerInspector set
	if children[1].GetLayer() != rtui.LayerInspector {
		t.Errorf("Inspector overlay should have LayerInspector (%d), got %d",
			rtui.LayerInspector, children[1].GetLayer())
	}
}

// TestInspectorHookSetsPosition tests that hook sets x/y position props
func TestInspectorHookSetsPosition(t *testing.T) {
	inspector := NewStandaloneInspector()

	// Enable inspector
	inspector.mu.Lock()
	inspector.visible = true
	inspector.mu.Unlock()

	hook := CreateInspectorHook(inspector)

	baseVNode := rtui.NewElement("base")
	result := hook(baseVNode)

	fragment, ok := result.(*rtui.FragmentVNode)
	if !ok {
		t.Fatalf("Expected FragmentVNode, got %T", result)
	}

	children := fragment.Children()
	if len(children) < 2 {
		t.Fatalf("Expected at least 2 children in Fragment, got %d", len(children))
	}

	inspectorOverlay := children[1]

	// Check that x and y props are set
	props := inspectorOverlay.Props()
	if props == nil {
		t.Fatal("Inspector overlay should have props")
	}

	x, hasX := props["x"].(int)
	y, hasY := props["y"].(int)

	if !hasX || !hasY {
		t.Error("Inspector overlay should have x and y props")
	}

	// Default position should be (80, 5)
	// TODO: Update to (40, 5) once screen-aware positioning is implemented
	if x != 80 || y != 5 {
		t.Errorf("Expected position (80, 5), got (%d, %d)", x, y)
	}
}

// TestInspectorHookSetsSize tests that hook sets width/height props
func TestInspectorHookSetsSize(t *testing.T) {
	inspector := NewStandaloneInspector()

	// Enable inspector
	inspector.mu.Lock()
	inspector.visible = true
	inspector.mu.Unlock()

	hook := CreateInspectorHook(inspector)

	baseVNode := rtui.NewElement("base")
	result := hook(baseVNode)

	fragment, ok := result.(*rtui.FragmentVNode)
	if !ok {
		t.Fatalf("Expected FragmentVNode, got %T", result)
	}

	children := fragment.Children()
	if len(children) < 2 {
		t.Fatalf("Expected at least 2 children in Fragment, got %d", len(children))
	}

	inspectorOverlay := children[1]

	props := inspectorOverlay.Props()
	if props == nil {
		t.Fatal("Inspector overlay should have props")
	}

	width, hasWidth := props["width"].(int)
	height, hasHeight := props["height"].(int)

	if !hasWidth || !hasHeight {
		t.Error("Inspector overlay should have width and height props")
	}

	// Default size should be 80x25
	if width != 80 || height != 25 {
		t.Errorf("Expected size 80x25, got %dx%d", width, height)
	}
}

// TestRegisterInspector tests RegisterInspector function
func TestRegisterInspector(t *testing.T) {
	inspector := NewStandaloneInspector()

	// Use real HookManager from render package
	hookManager := render.NewHookManager()

	RegisterInspector(inspector, hookManager)

	// Should have registered one hook
	if hookManager.VNodeHookCount() != 1 {
		t.Errorf("Expected 1 hook registered, got %d", hookManager.VNodeHookCount())
	}
}
