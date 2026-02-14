// Integration test for Inspector overlay positioning
package layer

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestInspectorOverlayFullFlow tests the complete flow from VNode to positioned layout
func TestInspectorOverlayFullFlow(t *testing.T) {
	// Simulate what happens in main.go
	inspectorOverlay := rtui.Bordered().
		Width(80).
		Height(25).
		Child(rtui.VStack()).
		Build()

	// Set layer and position (this is what StandaloneInspector.RenderOverlay does)
	inspectorOverlay.SetLayer(rtui.LayerInspector)
	inspectorOverlay.SetProps(rtui.Props{
		"x": 80,
		"y": 5,
	})

	// Create a layer node (this is what LayerManager.Collect does)
	layerNode := &LayerNode{
		Content: inspectorOverlay,
		Visible: true,
	}

	// Create layout engine and manager
	engine := compute.NewEngine()
	manager := NewManager()

	// Simulate full-screen constraints (120x40 terminal)
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	// Layout the layer (this is what LayerManager.layoutLayer does)
	layout, err := manager.layoutLayer(layerNode, rtui.LayerInspector, constraints, engine, nil)
	if err != nil {
		t.Fatalf("layoutLayer failed: %v", err)
	}

	if layout.Root == nil {
		t.Fatal("layout.Root is nil")
	}

	// Verify the overlay has the correct size
	if layout.Root.Box.Width != 80 {
		t.Errorf("Expected width 80, got %d", layout.Root.Box.Width)
	}
	if layout.Root.Box.Height != 25 {
		t.Errorf("Expected height 25, got %d", layout.Root.Box.Height)
	}

	// Verify the overlay is positioned at the specified coordinates
	if layout.Root.Box.X != 80 {
		t.Errorf("Expected X=80, got X=%d", layout.Root.Box.X)
	}
	if layout.Root.Box.Y != 5 {
		t.Errorf("Expected Y=5, got Y=%d", layout.Root.Box.Y)
	}

	t.Logf("✅ Inspector overlay correctly positioned at (%d, %d) with size %dx%d",
		layout.Root.Box.X, layout.Root.Box.Y,
		layout.Root.Box.Width, layout.Root.Box.Height)
}

// TestInspectorOverlayInVStack tests that VStack doesn't affect Inspector positioning
func TestInspectorOverlayInVStack(t *testing.T) {
	// Create app content
	appContent := rtui.VStack(
		rtui.NewElement("text"),
		rtui.NewElement("text"),
		rtui.NewElement("text"),
	)

	// Create Inspector overlay with explicit position
	inspectorOverlay := rtui.Bordered().
		Width(80).
		Height(25).
		Child(rtui.NewElement("text")).
		Build()

	inspectorOverlay.SetLayer(rtui.LayerInspector)
	inspectorOverlay.SetProps(rtui.Props{
		"x": 80,
		"y": 5,
	})

	// Put both in VStack (this is what main.go does)
	root := rtui.VStack(
		appContent,
		inspectorOverlay,
	)

	// Create layer manager and extract layer nodes
	manager := NewManager()
	manager.collector.Collect(root)

	// Verify that Inspector was collected
	hasInspector := manager.HasInspector()
	if !hasInspector {
		t.Error("LayerManager should have found Inspector layer")
	}

	// Get Inspector nodes
	inspectorNodes := manager.GetInspectorNodes()
	if len(inspectorNodes) == 0 {
		t.Fatal("No inspector nodes found")
	}

	// Layout the inspector layer
	engine := compute.NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	layout, err := manager.layoutLayer(inspectorNodes[0], rtui.LayerInspector, constraints, engine, nil)
	if err != nil {
		t.Fatalf("layoutLayer failed: %v", err)
	}

	if layout.Root == nil {
		t.Fatal("layout.Root is nil")
	}

	// Verify position is NOT affected by VStack
	if layout.Root.Box.X != 80 {
		t.Errorf("Expected X=80 (independent of VStack), got X=%d", layout.Root.Box.X)
	}
	if layout.Root.Box.Y != 5 {
		t.Errorf("Expected Y=5 (independent of VStack), got Y=%d", layout.Root.Box.Y)
	}

	t.Logf("✅ Inspector positioning independent of VStack: (%d, %d)",
		layout.Root.Box.X, layout.Root.Box.Y)
}

// TestInspectorOverlayZOrder verifies that Inspector is the highest layer
func TestInspectorOverlayZOrder(t *testing.T) {
	// Create overlay
	overlay := rtui.Bordered().
		Width(80).
		Height(25).
		Child(rtui.NewElement("text")).
		Build()

	overlay.SetLayer(rtui.LayerInspector)
	overlay.SetProps(rtui.Props{
		"x": 80,
		"y": 5,
	})

	// Check z-index
	zIndex := overlay.GetLayer().ZIndex()
	expectedZIndex := 4 // LayerInspector should be 4 (base=0, overlay=1, modal=2, tooltip=3, inspector=4)

	if zIndex != expectedZIndex {
		t.Errorf("Expected z-index %d, got %d", expectedZIndex, zIndex)
	}

	t.Logf("✅ Inspector has highest z-index: %d", zIndex)
}

// TestInspectorPositionVariations tests different Inspector positions
func TestInspectorPositionVariations(t *testing.T) {
	positions := []struct {
		name      string
		x         int
		y         int
		width     int
		height    int
		expectedX int
		expectedY int
	}{
		{"Top-left corner", 0, 0, 60, 20, 0, 0},
		{"Top-right corner", 100, 0, 80, 25, 100, 0},
		{"Bottom-left", 0, 30, 60, 20, 0, 30},
		{"Default position", 80, 5, 80, 25, 80, 5},
		{"Custom center", 40, 10, 70, 22, 40, 10},
	}

	for _, tt := range positions {
		t.Run(tt.name, func(t *testing.T) {
			overlay := rtui.Bordered().
				Width(tt.width).
				Height(tt.height).
				Child(rtui.NewElement("text")).
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

			manager := NewManager()
			engine := compute.NewEngine()
			constraints := runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  120,
				MinHeight: 0,
				MaxHeight: 40,
			}

			layout, err := manager.layoutLayer(layerNode, rtui.LayerInspector, constraints, engine, nil)
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
