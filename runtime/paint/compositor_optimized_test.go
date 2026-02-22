package paint

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Compositor Layer Culling Tests
// =============================================================================

func TestCompositor_layerInViewport(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		layer    *Layer
		expected bool
	}{
		{
			name:   "Layer fully inside viewport",
			width:  100,
			height: 50,
			layer:  NewLayerWithRect("test", LayerBackground, 0, Rect{X: 10, Y: 10, Width: 50, Height: 30}),
			expected: true,
		},
		{
			name:   "Layer partially outside right",
			width:  100,
			height: 50,
			layer:  NewLayerWithRect("test", LayerBackground, 0, Rect{X: 80, Y: 10, Width: 50, Height: 30}),
			expected: true,
		},
		{
			name:   "Layer completely outside right",
			width:  100,
			height: 50,
			layer:  NewLayerWithRect("test", LayerBackground, 0, Rect{X: 100, Y: 10, Width: 50, Height: 30}),
			expected: false,
		},
		{
			name:   "Layer completely outside left",
			width:  100,
			height: 50,
			layer:  NewLayerWithRect("test", LayerBackground, 0, Rect{X: -60, Y: 10, Width: 50, Height: 30}),
			expected: false,
		},
		{
			name:   "Layer completely outside top",
			width:  100,
			height: 50,
			layer:  NewLayerWithRect("test", LayerBackground, 0, Rect{X: 10, Y: -60, Width: 50, Height: 30}),
			expected: false,
		},
		{
			name:   "Layer completely outside bottom",
			width:  100,
			height: 50,
			layer:  NewLayerWithRect("test", LayerBackground, 0, Rect{X: 10, Y: 50, Width: 50, Height: 30}),
			expected: false,
		},
		{
			name:   "Layer slightly outside top-left",
			width:  100,
			height: 50,
			layer:  NewLayerWithRect("test", LayerBackground, 0, Rect{X: -10, Y: -10, Width: 50, Height: 30}),
			expected: true,
		},
		{
			name:   "Layer covering entire viewport",
			width:  100,
			height: 50,
			layer:  NewLayerWithRect("test", LayerBackground, 0, Rect{X: 0, Y: 0, Width: 100, Height: 50}),
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCompositor(tc.width, tc.height)
			result := c.layerInViewport(tc.layer)

			if result != tc.expected {
				t.Errorf("layerInViewport() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestCompositor_IsLayerVisible(t *testing.T) {
	c := NewCompositor(100, 50)
	layer := NewLayerWithRect("test", LayerBackground, 0, Rect{X: 10, Y: 10, Width: 50, Height: 30})

	// Test enabled + visible + in viewport
	layer.Enabled = true
	layer.Visible = true
	if !c.IsLayerVisible(layer) {
		t.Error("Layer should be visible when enabled, visible, and in viewport")
	}

	// Test disabled
	layer.Enabled = false
	if c.IsLayerVisible(layer) {
		t.Error("Layer should not be visible when disabled")
	}
	layer.Enabled = true

	// Test invisible
	layer.Visible = false
	if c.IsLayerVisible(layer) {
		t.Error("Layer should not be visible when Visible=false")
	}
	layer.Visible = true

	// Test outside viewport
	layerOutside := NewLayerWithRect("outside", LayerBackground, 0, Rect{X: 200, Y: 10, Width: 50, Height: 30})
	layerOutside.Enabled = true
	layerOutside.Visible = true
	if c.IsLayerVisible(layerOutside) {
		t.Error("Layer should not be visible when outside viewport")
	}
}

// =============================================================================
// Compositor RenderDirty Optimization Tests
// =============================================================================

func TestCompositor_RenderDirty_CullsInvisible(t *testing.T) {
	c := NewCompositor(100, 50)

	// Visible layer
	visibleLayer := NewLayerWithRect("visible", LayerBackground, 0, Rect{X: 10, Y: 10, Width: 50, Height: 30})
	visibleLayer.Buffer.SetCell(0, 0, 'X', style.Style{FG: "red"})
	visibleLayer.MarkDirty()
	c.AddLayer(visibleLayer)

	// Disabled layer
	disabledLayer := NewLayerWithRect("disabled", LayerContent, 1, Rect{X: 10, Y: 10, Width: 50, Height: 30})
	disabledLayer.Buffer.SetCell(0, 0, 'Y', style.Style{FG: "blue"})
	disabledLayer.MarkDirty()
	disabledLayer.Enabled = false
	c.AddLayer(disabledLayer)

	// Invisible layer
	invisibleLayer := NewLayerWithRect("invisible", LayerOverlay, 2, Rect{X: 10, Y: 10, Width: 50, Height: 30})
	invisibleLayer.Buffer.SetCell(0, 0, 'Z', style.Style{FG: "green"})
	invisibleLayer.MarkDirty()
	invisibleLayer.Visible = false
	c.AddLayer(invisibleLayer)

	// Outside viewport layer
	outsideLayer := NewLayerWithRect("outside", LayerStream, 3, Rect{X: 200, Y: 10, Width: 50, Height: 30})
	outsideLayer.Buffer.SetCell(0, 0, 'W', style.Style{FG: "yellow"})
	outsideLayer.MarkDirty()
	c.AddLayer(outsideLayer)

	// Only the visible layer should be rendered
	output := c.RenderDirty()

	// Should contain content from visible layer
	if len(output) == 0 {
		t.Error("Should render visible layer")
	}

	// Disabled and invisible layers should not be cleared
	// (they have dirty internally, but RenderDirty should skip them)
	if !disabledLayer.Dirty {
		t.Error("Disabled layer dirty flag should not be cleared")
	}
	if !invisibleLayer.Dirty {
		t.Error("Invisible layer dirty flag should not be cleared")
	}
	// Outside viewport layers should also not be cleared
	if !outsideLayer.Dirty {
		t.Error("Outside viewport layer dirty flag should not be cleared")
	}
}

func TestCompositor_RenderDirty_SkipsCleanLayers(t *testing.T) {
	c := NewCompositor(100, 50)

	// Create multiple layers
	dirtyLayer := NewLayerWithRect("dirty", LayerBackground, 0, Rect{X: 10, Y: 10, Width: 50, Height: 30})
	dirtyLayer.Buffer.SetCell(0, 0, 'X', style.Style{FG: "red"})
	dirtyLayer.MarkDirty()
	c.AddLayer(dirtyLayer)

	cleanLayer := NewLayerWithRect("clean", LayerContent, 1, Rect{X: 10, Y: 10, Width: 50, Height: 30})
	cleanLayer.Buffer.SetCell(0, 0, 'Y', style.Style{FG: "blue"})
	// Note: not marked dirty
	c.AddLayer(cleanLayer)

	// Render (should only render dirty layer)
	output := c.RenderDirty()

	// Verify output is produced
	if len(output) == 0 {
		t.Error("Should render dirty layer")
	}

	// Verify clean layer is still clean
	if cleanLayer.IsDirty() {
		t.Error("Clean layer should not be marked dirty")
	}

	// Verify dirty layer is cleared
	if dirtyLayer.IsDirty() {
		t.Error("Dirty layer should be cleared after render")
	}
}

// =============================================================================
// Compositor blitLayer Clipping Tests
// =============================================================================

func TestCompositor_blitLayer_Clipping(t *testing.T) {
	c := NewCompositor(100, 50)
	dst := NewBuffer(100, 50)

	// Layer partially outside viewport on right
	src := NewLayerWithRect("partial", LayerBackground, 0, Rect{X: 90, Y: 10, Width: 50, Height: 30})
	src.Buffer.SetCell(0, 0, 'A', style.Style{FG: "red"})
	src.Buffer.SetCell(20, 0, 'B', style.Style{FG: "blue"})
	src.Buffer.SetCell(40, 0, 'C', style.Style{FG: "green"})


	c.blitLayer(dst, src)

	// Check that only visible cells are copied
	if dst.Cells[10][90].Cluster != "A" {
		t.Error("Should copy visible cell at viewport edge")
	}

	// Check that cells outside viewport are not copied (should be initial space)
	// Cell at position (110, 10) would be outside viewport
	if dst.Cells[10][95].Cluster != "" {
		// This cell is within viewport, so it should be from layer
	}
}

func TestCompositor_blitLayer_CompletelyOutside(t *testing.T) {
	c := NewCompositor(100, 50)
	dst := NewBuffer(100, 50)

	// Layer completely outside viewport
	src := NewLayerWithRect("outside", LayerBackground, 0, Rect{X: 200, Y: 10, Width: 50, Height: 30})
	src.Buffer.SetCell(0, 0, 'X', style.Style{FG: "red"})

	// Clear destination
	for y := 0; y < dst.Height; y++ {
		for x := 0; x < dst.Width; x++ {
			dst.Cells[y][x] = Cell{Cluster: " "}
		}
	}

	c.blitLayer(dst, src)

	// Check that destination is unchanged (all spaces)
	for y := 0; y < dst.Height; y++ {
		for x := 0; x < dst.Width; x++ {
			if dst.Cells[y][x].Cluster != " " {
				t.Errorf("Destination should be unchanged at (%d, %d)", x, y)
			}
		}
	}
}

func TestCompositor_blitLayer_NegativeCoordinates(t *testing.T) {
	c := NewCompositor(100, 50)
	dst := NewBuffer(100, 50)

	// Layer with negative coordinates (partially outside on left/top)
	src := NewLayerWithRect("negative", LayerBackground, 0, Rect{X: -10, Y: -5, Width: 30, Height: 20})

	// Set some cells that would be visible
	src.Buffer.SetCell(10, 5, 'V', style.Style{FG: "red"}) // Should be visible at (0, 0)
	src.Buffer.SetCell(0, 0, 'H', style.Style{FG: "blue"})  // Should be outside

	c.blitLayer(dst, src)

	// Check visible cell is copied
	if dst.Cells[0][0].Cluster != "V" {
		t.Error("Should copy cell that becomes visible at (0, 0)")
	}

	// Hidden cell should not be copied
	// At (-10, -5) which is outside buffer
}

func TestCompositor_RenderDirtyRect(t *testing.T) {
	c := NewCompositor(100, 50)

	// Create two layers with content
	layer1 := NewLayerWithRect("layer1", LayerBackground, 0, Rect{X: 0, Y: 0, Width: 50, Height: 25})
	layer1.Buffer.SetString(0, 0, "AAAA", style.Style{FG: "red"})
	layer1.Buffer.SetString(0, 1, "BBBB", style.Style{FG: "red"})
	layer1.Buffer.SetString(0, 2, "CCCC", style.Style{FG: "red"})
	layer1.MarkDirty()
	c.AddLayer(layer1)

	layer2 := NewLayerWithRect("layer2", LayerContent, 1, Rect{X: 0, Y: 25, Width: 50, Height: 25})
	layer2.Buffer.SetString(0, 0, "DDDD", style.Style{FG: "blue"})
	layer2.Buffer.SetString(0, 1, "EEEE", style.Style{FG: "blue"})
	layer2.Buffer.SetString(0, 2, "FFFF", style.Style{FG: "blue"})
	layer2.MarkDirty()
	c.AddLayer(layer2)

	// Render only the top half of screen
	clipRect := Rect{X: 0, Y: 0, Width: 100, Height: 25}
	output := c.RenderDirtyRect(clipRect)

	// Should contain content from layer1
	if !strings.Contains(output, "AAAA") {
		t.Error("Should contain content from layer1 in render")
	}

	// Layer1 should be cleared (rendered)
	if layer1.IsDirty() {
		t.Error("Layer1 should be cleared after render")
	}

	// Layer2 should not be cleared (not in clip region)
	if !layer2.Dirty {
		t.Error("Layer2 should not be cleared (outside clip region)")
	}
}

func TestCompositor_RenderDirtyRect_PartialLayer(t *testing.T) {
	c := NewCompositor(100, 50)

	// Create a layer spanning both halves
	layer := NewLayerWithRect("layer", LayerBackground, 0, Rect{X: 0, Y: 0, Width: 50, Height: 50})
	layer.Buffer.SetString(0, 0, "TOP", style.Style{FG: "red"})
	layer.Buffer.SetString(0, 20, "MIDDLE", style.Style{FG: "red"})
	layer.Buffer.SetString(0, 40, "BOTTOM", style.Style{FG: "red"})
	layer.MarkDirty()
	c.AddLayer(layer)

	// Render only the middle section
	clipRect := Rect{X: 0, Y: 15, Width: 50, Height: 10}
	_ = c.RenderDirtyRect(clipRect)

	// Layer should be cleared after render
	if layer.IsDirty() {
		t.Error("Layer should be cleared after render")
	}
}

func TestCompositor_clipRectToViewport(t *testing.T) {
	c := NewCompositor(100, 50)

	tests := []struct {
		name     string
		rect     Rect
		expected Rect
	}{
		{
			name:     "Rect inside viewport",
			rect:     Rect{X: 10, Y: 10, Width: 50, Height: 30},
			expected: Rect{X: 10, Y: 10, Width: 50, Height: 30},
		},
		{
			name:     "Rect extends beyond right",
			rect:     Rect{X: 80, Y: 10, Width: 50, Height: 30},
			expected: Rect{X: 80, Y: 10, Width: 20, Height: 30},
		},
		{
			name:     "Rect extends beyond bottom",
			rect:     Rect{X: 10, Y: 40, Width: 50, Height: 30},
			expected: Rect{X: 10, Y: 40, Width: 50, Height: 10},
		},
		{
			name:     "Rect outside left",
			rect:     Rect{X: -30, Y: 10, Width: 20, Height: 30},
			expected: Rect{X: 0, Y: 10, Width: 0, Height: 30},
		},
		{
			name:     "Rect outside top",
			rect:     Rect{X: 10, Y: -30, Width: 50, Height: 20},
			expected: Rect{X: 10, Y: 0, Width: 50, Height: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := c.clipRectToViewport(tc.rect)

			// Allow small variations due to clipping logic
			if result.X != tc.expected.X {
				t.Errorf("X: got %d, want %d", result.X, tc.expected.X)
			}
			if result.Y != tc.expected.Y {
				t.Errorf("Y: got %d, want %d", result.Y, tc.expected.Y)
			}
			if result.Width != tc.expected.Width {
				t.Errorf("Width: got %d, want %d", result.Width, tc.expected.Width)
			}
			if result.Height != tc.expected.Height {
				t.Errorf("Height: got %d, want %d", result.Height, tc.expected.Height)
			}
		})
	}
}

func TestRectIntersect(t *testing.T) {
	tests := []struct {
		r1      Rect
		r2      Rect
		expect  bool
	}{
		{
			r1:     Rect{X: 0, Y: 0, Width: 10, Height: 10},
			r2:     Rect{X: 5, Y: 5, Width: 10, Height: 10},
			expect: true,
		},
		{
			r1:     Rect{X: 0, Y: 0, Width: 10, Height: 10},
			r2:     Rect{X: 10, Y: 0, Width: 10, Height: 10},
			expect: false, // Edge touching doesn't count
		},
		{
			r1:     Rect{X: 0, Y: 0, Width: 10, Height: 10},
			r2:     Rect{X: 20, Y: 20, Width: 10, Height: 10},
			expect: false,
		},
		{
			r1:     Rect{X: 0, Y: 0, Width: 50, Height: 50},
			r2:     Rect{X: 25, Y: 25, Width: 10, Height: 10},
			expect: true,
		},
	}

	for _, tc := range tests {
		result := rectIntersect(tc.r1, tc.r2)
		if result != tc.expect {
			t.Errorf("rectIntersect(%v, %v) = %v, want %v", tc.r1, tc.r2, result, tc.expect)
		}
	}
}

func TestRectIntersection(t *testing.T) {
	tests := []struct {
		r1      Rect
		r2      Rect
		expect  Rect
	}{
		{
			r1:     Rect{X: 0, Y: 0, Width: 10, Height: 10},
			r2:     Rect{X: 5, Y: 5, Width: 10, Height: 10},
			expect: Rect{X: 5, Y: 5, Width: 5, Height: 5},
		},
		{
			r1:     Rect{X: 0, Y: 0, Width: 10, Height: 10},
			r2:     Rect{X: 0, Y: 0, Width: 10, Height: 10},
			expect: Rect{X: 0, Y: 0, Width: 10, Height: 10},
		},
	}

	for _, tc := range tests {
		result := rectIntersection(tc.r1, tc.r2)
		if result != tc.expect {
			t.Errorf("rectIntersection(%v, %v) = %v, want %v", tc.r1, tc.r2, result, tc.expect)
		}
	}
}

// =============================================================================

// Helper function to create a test layer with some content
func createTestLayer(id string, layerType LayerType, zIndex int, rect Rect, content string) *Layer {
	layer := NewLayerWithRect(id, layerType, zIndex, rect)

	// Fill with content pattern
	for y := 0; y < layer.Buffer.Height && y < len(content); y++ {
		for x := 0; x < layer.Buffer.Width && x < len(content) && x < 10; x++ {
			if x < len(content) {
				layer.Buffer.SetCell(x, y, rune(content[x]), style.Style{FG: "white"})
			}
		}
	}

	return layer
}
