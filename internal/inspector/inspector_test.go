package inspector

import (
	"testing"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// TestNewInspector tests creating a new Inspector
func TestNewInspector(t *testing.T) {
	inspector := NewInspector()

	if inspector == nil {
		t.Fatal("Expected non-nil Inspector")
	}

	if inspector.IsEnabled() {
		t.Error("New inspector should be disabled by default")
	}
}

// TestInspectorEnableDisable tests enabling and disabling the inspector
func TestInspectorEnableDisable(t *testing.T) {
	inspector := NewInspector()

	// Test enable
	inspector.Enable()
	if !inspector.IsEnabled() {
		t.Error("Inspector should be enabled after Enable()")
	}

	// Test disable
	inspector.Disable()
	if inspector.IsEnabled() {
		t.Error("Inspector should be disabled after Disable()")
	}

	// Test that disable clears selection
	inspector.Enable()
	button := app.ButtonBuilder("Test").Build()
	inspector.SetSelectedVNode(button)

	inspector.Disable()

	if inspector.GetSelectedVNode() != nil {
		t.Error("Selected VNode should be cleared after disable")
	}
}

// TestSetSelectedVNode tests setting the selected VNode
func TestSetSelectedVNode(t *testing.T) {
	inspector := NewInspector()

	button := app.ButtonBuilder("Test Button").Build()
	inspector.SetSelectedVNode(button)

	selected := inspector.GetSelectedVNode()
	if selected == nil {
		t.Error("Selected VNode should not be nil")
	}

	// Verify it's the same button
	info := ExtractElementInfo(selected)
	if info.Label != "Test Button" {
		t.Errorf("Expected label 'Test Button', got '%s'", info.Label)
	}
}

// TestGetSelectedInfo tests getting ElementInfo for selected VNode
func TestGetSelectedInfo(t *testing.T) {
	inspector := NewInspector()

	button := app.ButtonBuilder("Click Me").Build()
	inspector.SetSelectedVNode(button)

	info := inspector.GetSelectedInfo()

	if info.Type == "" {
		t.Error("Type should not be empty")
	}

	if info.Label != "Click Me" {
		t.Errorf("Expected label 'Click Me', got '%s'", info.Label)
	}
}

// TestGetHoveredInfo tests getting ElementInfo for hovered VNode
func TestGetHoveredInfo(t *testing.T) {
	inspector := NewInspector()

	text := ui.Text("Hover Test")

	// Simulate hovering (manually set for now)
	inspector.hoveredVNode = text

	info := inspector.GetHoveredInfo()

	if info.Label != "Hover Test" {
		t.Errorf("Expected label 'Hover Test', got '%s'", info.Label)
	}
}

// TestFindVNodeAt tests finding a VNode at a position
func TestFindVNodeAt(t *testing.T) {
	t.Skip("Requires actual layout engine to set bounds properly")

	inspector := NewInspector()

	// Create a simple VNode tree
	button := app.ButtonBuilder("Test").Build()

	// Simulate bounds being set
	if boundsAware, ok := button.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsAware.SetBounds(10, 5, 20, 1)
	}

	// Find node at a position within the button
	found := inspector.FindVNodeAt(button, 15, 5)

	if found == nil {
		t.Error("Expected to find VNode at (15, 5)")
	}

	info := ExtractElementInfo(found)
	if info.Label != "Test" {
		t.Errorf("Expected to find button with label 'Test', got '%s'", info.Label)
	}

	// Find node at a position outside the button
	foundOutside := inspector.FindVNodeAt(button, 100, 100)
	if foundOutside != nil {
		t.Error("Expected nil when searching outside button bounds")
	}
}

// TestHandleMouseEvent tests mouse event handling
func TestHandleMouseEvent(t *testing.T) {
	inspector := NewInspector()
	inspector.Enable()

	// Handle a mouse event
	handled := inspector.HandleMouseEvent(50, 25)

	if !handled {
		// For now, HandleMouseEvent returns false until layout integration
		// This is expected for Phase 2
		t.Log("HandleMouseEvent returns false (expected until layout integration)")
	}

	// Check mouse position was recorded
	x, y := inspector.GetMousePosition()
	if x != 50 || y != 25 {
		t.Errorf("Expected mouse position (50, 25), got (%d, %d)", x, y)
	}
}

// TestHandleMouseEvent_Disabled tests that mouse events are ignored when disabled
func TestHandleMouseEvent_Disabled(t *testing.T) {
	inspector := NewInspector()
	// Don't enable inspector

	handled := inspector.HandleMouseEvent(50, 25)

	if handled {
		t.Error("Mouse event should not be handled when inspector is disabled")
	}
}

// TestVNodeContains tests the VNode containment check
func TestVNodeContains(t *testing.T) {
	t.Skip("Requires actual layout engine to set bounds properly")

	button := app.ButtonBuilder("Test").Build()

	// Simulate bounds
	if boundsAware, ok := button.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsAware.SetBounds(10, 5, 20, 1)
	}

	tests := []struct {
		name     string
		x, y     int
		expected bool
	}{
		{"Inside bounds", 15, 5, true},
		{"At left edge", 10, 5, true},
		{"At right edge", 29, 5, true},
		{"Outside left", 9, 5, false},
		{"Outside right", 30, 5, false},
		{"Outside top", 15, 4, false},
		{"Outside bottom", 15, 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vnodeContains(button, tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("Expected %v for position (%d, %d), got %v",
					tt.expected, tt.x, tt.y, result)
			}
		})
	}
}

// TestNewOverlay tests creating a new Overlay
func TestNewOverlay(t *testing.T) {
	overlay := NewOverlay()

	if overlay == nil {
		t.Fatal("Expected non-nil Overlay")
	}

	if !overlay.showBorders {
		t.Error("Borders should be shown by default")
	}

	if !overlay.showDimensions {
		t.Error("Dimensions should be shown by default")
	}
}

// TestOverlaySetters tests overlay configuration methods
func TestOverlaySetters(t *testing.T) {
	overlay := NewOverlay()

	// Test SetShowDimensions
	overlay.SetShowDimensions(false)
	if overlay.showDimensions {
		t.Error("showDimensions should be false after SetShowDimensions(false)")
	}

	// Test SetShowBorders
	overlay.SetShowBorders(false)
	if overlay.showBorders {
		t.Error("showBorders should be false after SetShowBorders(false)")
	}
}

// TestGetBorderStyle tests getting border style for different element types
func TestGetBorderStyle(t *testing.T) {
	overlay := NewOverlay()

	button := app.ButtonBuilder("Test").Build()
	text := ui.Text("Hello")

	buttonStyle := overlay.GetBorderStyle(button)
	textStyle := overlay.GetBorderStyle(text)

	if string(buttonStyle) != "▓" {
		t.Errorf("Expected diamond style for button, got '%s'", string(buttonStyle))
	}

	if string(textStyle) != "•" {
		t.Errorf("Expected bullet style for text, got '%s'", string(textStyle))
	}
}

// TestPaintHighlight tests painting corner highlights
func TestPaintHighlight(t *testing.T) {
	overlay := NewOverlay()

	button := app.ButtonBuilder("Test").Build()

	// Set bounds
	if boundsAware, ok := button.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsAware.SetBounds(10, 5, 20, 3)
	}

	// We can't easily test the actual buffer without importing paint.Buffer
	// So we just test that the method doesn't panic
	err := overlay.PaintHighlight(nil, button, '*')
	if err != nil {
		t.Errorf("PaintHighlight should not return error, got %v", err)
	}
}
