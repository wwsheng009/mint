// Test for Inspector overlay positioning
package layer

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestInspectorPositioning verifies that Inspector overlay is positioned at specified coordinates
func TestInspectorPositioning(t *testing.T) {
	// Create a manager
	mgr := NewManager()

	// Create an Inspector overlay with explicit position
	// Using VStack as content to avoid text node complexity
	overlay := rtui.Bordered().
		Width(80).
		Height(25).
		Child(rtui.VStack()).
		Build()

	// Set layer and position props
	overlay.SetLayer(rtui.LayerInspector)
	overlay.SetProps(rtui.Props{
		"x": 80, // Position at x=80
		"y": 5,  // Position at y=5
	})

	// Create a layer node
	layerNode := &LayerNode{
		Content: overlay,
		Visible: true,
	}

	// Create layout engine
	engine := compute.NewEngine()

	// Layout with full-screen constraints
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	// Layout the layer
	layout, err := mgr.layoutLayer(layerNode, rtui.LayerInspector, constraints, engine, nil)
	if err != nil {
		t.Fatalf("layoutLayer failed: %v", err)
	}

	if layout.Root == nil {
		t.Fatal("layout.Root is nil")
	}

	// Verify the overlay is positioned at the specified coordinates
	if layout.Root.Box.X != 80 {
		t.Errorf("Expected X=80, got X=%d", layout.Root.Box.X)
	}
	if layout.Root.Box.Y != 5 {
		t.Errorf("Expected Y=5, got Y=%d", layout.Root.Box.Y)
	}

	// Note: Size may vary based on content, we're testing positioning here
	t.Logf("✅ Inspector overlay positioned at (%d, %d) with size %dx%d",
		layout.Root.Box.X, layout.Root.Box.Y,
		layout.Root.Box.Width, layout.Root.Box.Height)
}

// TestInspectorPositioningWithoutProps verifies default positioning when no props are set
func TestInspectorPositioningWithoutProps(t *testing.T) {
	// Create a manager
	mgr := NewManager()

	// Create an Inspector overlay WITHOUT position props
	overlay := rtui.Bordered().
		Width(80).
		Height(25).
		Child(rtui.VStack()).
		Build()

	// Set layer only (no position props)
	overlay.SetLayer(rtui.LayerInspector)

	// Create a layer node
	layerNode := &LayerNode{
		Content: overlay,
		Visible: true,
	}

	// Create layout engine
	engine := compute.NewEngine()

	// Layout with full-screen constraints
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	// Layout the layer
	layout, err := mgr.layoutLayer(layerNode, rtui.LayerInspector, constraints, engine, nil)
	if err != nil {
		t.Fatalf("layoutLayer failed: %v", err)
	}

	if layout.Root == nil {
		t.Fatal("layout.Root is nil")
	}

	// Verify the overlay has the expected size
	if layout.Root.Box.Width != 80 {
		t.Errorf("Expected width 80, got %d", layout.Root.Box.Width)
	}
	if layout.Root.Box.Height != 25 {
		t.Errorf("Expected height 25, got %d", layout.Root.Box.Height)
	}

	// Without position props, should default to (0, 0)
	if layout.Root.Box.X != 0 {
		t.Errorf("Expected default X=0, got X=%d", layout.Root.Box.X)
	}
	if layout.Root.Box.Y != 0 {
		t.Errorf("Expected default Y=0, got Y=%d", layout.Root.Box.Y)
	}

	t.Logf("✅ Inspector overlay with default position at (%d, %d)",
		layout.Root.Box.X, layout.Root.Box.Y)
}

// TestInspectorPositioningEdgeCases tests various edge cases
func TestInspectorPositioningEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		x         int
		y         int
		width     int
		height    int
		expectedX int
		expectedY int
	}{
		{"Top-left corner", 0, 0, 80, 25, 0, 0},
		{"Top-right corner", 100, 0, 80, 25, 100, 0},
		{"Bottom-left", 0, 30, 80, 25, 0, 30},
		{"Center", 40, 15, 80, 25, 40, 15},
		{"Negative coordinates (should clamp to 0)", -10, -5, 80, 25, 0, 0}, // positionInspector clamps to 0,0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a manager
			mgr := NewManager()

			// Create overlay with specified position
			overlay := rtui.Bordered().
				Width(tt.width).
				Height(tt.height).
				Child(rtui.VStack()).
				Build()

			overlay.SetLayer(rtui.LayerInspector)
			overlay.SetProps(rtui.Props{
				"x": tt.x,
				"y": tt.y,
			})

			layerNode := &LayerNode{
				Content: overlay,
				Visible: true,
			}

			engine := compute.NewEngine()
			constraints := runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  120,
				MinHeight: 0,
				MaxHeight: 40,
			}

			layout, err := mgr.layoutLayer(layerNode, rtui.LayerInspector, constraints, engine, nil)
			if err != nil {
				t.Fatalf("layoutLayer failed: %v", err)
			}

			if layout.Root.Box.X != tt.expectedX {
				t.Errorf("Expected X=%d, got X=%d", tt.expectedX, layout.Root.Box.X)
			}
			if layout.Root.Box.Y != tt.expectedY {
				t.Errorf("Expected Y=%d, got Y=%d", tt.expectedY, layout.Root.Box.Y)
			}

			t.Logf("✅ %s: positioned at (%d, %d)",
				tt.name, layout.Root.Box.X, layout.Root.Box.Y)
		})
	}
}
