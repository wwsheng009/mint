package inspector

import (
	"os"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestInspectorRenderOverlayPositioning verifies that RenderOverlay sets position props
func TestInspectorRenderOverlayPositioning(t *testing.T) {
	inspector := NewStandaloneInspector()
	inspector.Enable()
	inspector.ToggleVisibility()

	// Set custom position
	inspector.SetFloatingPosition(100, 10)

	// Render overlay
	overlay := inspector.RenderOverlay()

	if overlay == nil {
		t.Fatal("RenderOverlay returned nil")
	}

	// Check layer is set
	if overlay.GetLayer() != rtui.LayerInspector {
		t.Errorf("Expected LayerInspector, got %v", overlay.GetLayer())
	}

	// Check position props
	props := overlay.Props()
	if props == nil {
		t.Fatal("Overlay props are nil")
	}

	x, hasX := props["x"].(int)
	y, hasY := props["y"].(int)

	if !hasX {
		t.Error("Position prop 'x' not found")
	} else if x != 100 {
		t.Errorf("Expected x=100, got x=%d", x)
	}

	if !hasY {
		t.Error("Position prop 'y' not found")
	} else if y != 10 {
		t.Errorf("Expected y=10, got y=%d", y)
	}

	t.Logf("✅ Inspector overlay has correct position props: x=%d, y=%d", x, y)
}

// TestInspectorDefaultPosition verifies default positioning
func TestInspectorDefaultPosition(t *testing.T) {
	os.Setenv("TUI_DEBUG_INSPECTOR", "true")
	defer os.Setenv("TUI_DEBUG_INSPECTOR", "false")

	inspector := NewStandaloneInspector()
	inspector.Enable()
	inspector.ToggleVisibility()

	// Render overlay (should use default position from NewStandaloneInspector)
	overlay := inspector.RenderOverlay()

	if overlay == nil {
		t.Fatal("RenderOverlay returned nil")
	}

	// Check default position
	props := overlay.Props()
	if props == nil {
		t.Fatal("Overlay props are nil")
	}

	x, hasX := props["x"].(int)
	y, hasY := props["y"].(int)

	if !hasX || !hasY {
		t.Fatal("Position props not set")
	}

	// Default position from NewStandaloneInspector is floatX=80, floatY=5
	if x != 80 {
		t.Errorf("Expected default x=80, got x=%d", x)
	}
	if y != 5 {
		t.Errorf("Expected default y=5, got y=%d", y)
	}

	t.Logf("✅ Inspector overlay has default position: x=%d, y=%d", x, y)
}
