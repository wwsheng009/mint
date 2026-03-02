// Package layout - Portal positioning tests
package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime/types"
)

// TestParsePortalPositionConfig tests parsing of Portal positioning props
func TestParsePortalPositionConfig(t *testing.T) {
	props := map[string]interface{}{
		"position":  types.PositionFixed,
		"anchor":    types.AnchorCenter,
		"anchorId":  "button-123",
		"top":       100,
		"left":      50,
		"right":     0,
		"bottom":    0,
		"priority":  10,
	}

	config := ParsePortalPositionConfig(props, 800, 600, 400, 300)

	assert.Equal(t, types.PositionFixed, config.Position)
	assert.Equal(t, types.AnchorCenter, config.Anchor)
	assert.Equal(t, "button-123", config.AnchorID)
	assert.NotNil(t, config.Top)
	assert.Equal(t, 100, *config.Top)
	assert.NotNil(t, config.Left)
	assert.Equal(t, 50, *config.Left)
	assert.NotNil(t, config.Right)
	assert.Equal(t, 0, *config.Right)
	assert.NotNil(t, config.Bottom)
	assert.Equal(t, 0, *config.Bottom)
	assert.Equal(t, 800, config.ViewportWidth)
	assert.Equal(t, 600, config.ViewportHeight)
	assert.Equal(t, 400, config.PortalWidth)
	assert.Equal(t, 300, config.PortalHeight)
}

// TestCalculateFixedPosition_PositionFixed tests PositionFixed positioning
func TestCalculateFixedPosition_PositionFixed(t *testing.T) {
	calculator := NewPortalPositionCalculator()

	// Test: Center anchor (default modal centering)
	config := PortalPositionConfig{
		Position:      types.PositionFixed,
		Anchor:        types.AnchorCenter,
		ViewportWidth: 800,
		ViewportHeight: 600,
		PortalWidth:   400,
		PortalHeight:  300,
	}

	x, y := calculator.CalculatePosition(config)
	assert.Equal(t, 200, x) // (800-400)/2 = 200
	assert.Equal(t, 150, y) // (600-300)/2 = 150

	// Test: TopLeft with offsets
	configTopLeft := PortalPositionConfig{
		Position:      types.PositionFixed,
		Anchor:        types.AnchorTopLeft,
		Top:           ptrInt(10),
		Left:          ptrInt(20),
		ViewportWidth: 800,
		ViewportHeight: 600,
		PortalWidth:   400,
		PortalHeight:  300,
	}

	x, y = calculator.CalculatePosition(configTopLeft)
	assert.Equal(t, 20, x)
	assert.Equal(t, 10, y)

	// Test: BottomRight
	configBottomRight := PortalPositionConfig{
		Position:      types.PositionFixed,
		Anchor:        types.AnchorBottomRight,
		Right:         ptrInt(10),
		Bottom:        ptrInt(5),
		ViewportWidth: 800,
		ViewportHeight: 600,
		PortalWidth:   400,
		PortalHeight:  300,
	}

	x, y = calculator.CalculatePosition(configBottomRight)
	assert.Equal(t, 800-400-10, x)  // viewport - portalWidth - right
	assert.Equal(t, 600-300-5, y)   // viewport - portalHeight - bottom
	assert.Equal(t, 390, x)
	assert.Equal(t, 295, y)
}

// TestCalculateAnchorBasedPosition tests Anchor-based positioning (tooltips, popovers)
func TestCalculateAnchorBasedPosition(t *testing.T) {
	calculator := NewPortalPositionCalculator()

	// Test basic alignment scenarios without offsets
	pw, ph := 150, 40 // portal width/height
	ax, ay, aw, ah := 100, 100, 200, 50 // anchor position/size

	// Test TopLeft alignment
	config := PortalPositionConfig{
		Position:     types.PositionAbsolute,
		Anchor:       types.AnchorTopLeft,
		AnchorX:      ax,
		AnchorY:      ay,
		AnchorWidth:  aw,
		AnchorHeight: ah,
		PortalWidth:  pw,
		PortalHeight: ph,
	}
	x, y := calculator.CalculatePosition(config)
	assert.Equal(t, ax, x)  // portal's top-left at anchor's top-left (100)
	assert.Equal(t, ay, y)  // portal's top-left at anchor's top-left (100)

	// Test Bottom alignment (common for tooltips)
	configBottom := PortalPositionConfig{
		Position:     types.PositionAbsolute,
		Anchor:       types.AnchorBottom,
		AnchorX:      ax,
		AnchorY:      ay,
		AnchorWidth:  aw,
		AnchorHeight: ah,
		Top:          ptrInt(5),  // 5px gap
		PortalWidth:  pw,
		PortalHeight: ph,
	}
	x, y = calculator.CalculatePosition(configBottom)
	// x = ax + (aw-pw)/2 = 100 + (200-150)/2 = 125
	// y = ay + ah - ph + top = 100 + 50 - 40 + 5 = 115
	assert.Equal(t, 125, x)
	assert.Equal(t, 115, y)

	// Test Left alignment with offset
	configLeft := PortalPositionConfig{
		Position:     types.PositionAbsolute,
		Anchor:       types.AnchorLeft,
		AnchorX:      ax,
		AnchorY:      ay,
		AnchorWidth:  aw,
		AnchorHeight: ah,
		Left:         ptrInt(10),
		PortalWidth:  pw,
		PortalHeight: ph,
	}
	x, y = calculator.CalculatePosition(configLeft)
	// x = ax + left = 100 + 10 = 110
	// y = ay + (ah-ph)/2 = 100 + (50-40)/2 = 105
	assert.Equal(t, 110, x)
	assert.Equal(t, 105, y)

	// Test Right alignment
	configRight := PortalPositionConfig{
		Position:     types.PositionAbsolute,
		Anchor:       types.AnchorRight,
		AnchorX:      ax,
		AnchorY:      ay,
		AnchorWidth:  aw,
		AnchorHeight: ah,
		Right:        ptrInt(10),
		PortalWidth:  pw,
		PortalHeight: ph,
	}
	x, y = calculator.CalculatePosition(configRight)
	// x = ax + aw - pw - right = 100 + 200 - 150 - 10 = 140
	// y = ay + (ah-ph)/2 = 100 + (50-40)/2 = 105
	assert.Equal(t, 140, x)
	assert.Equal(t, 105, y)
}

// TestCalculateRootBasedPosition tests default Root-based positioning
func TestCalculateRootBasedPosition(t *testing.T) {
	calculator := NewPortalPositionCalculator()

	config := PortalPositionConfig{
		Position: types.PositionRelative,
		Left:     ptrInt(50),
		Top:      ptrInt(30),
	}

	x, y := calculator.CalculatePosition(config)
	assert.Equal(t, 50, x)
	assert.Equal(t, 30, y)
}

// TestFindAnchorPosition tests finding anchor elements in layout tree
func TestFindAnchorPosition(t *testing.T) {
	// Build a layout tree
	child1 := &LayoutBox{
		ID:     "child-1",
		X:      10,
		Y:      20,
		Width:  100,
		Height: 50,
	}

	child2 := &LayoutBox{
		ID:     "button-123", // This is our target anchor
		X:      120,
		Y:      30,
		Width:  200,
		Height: 50,
	}

	root := &LayoutBox{
		ID:       "root",
		X:        0,
		Y:        0,
		Width:    500,
		Height:   300,
		Children: []*LayoutBox{child1, child2},
	}

	// Find anchor position
	x, y, w, h, found := FindAnchorPosition(root, "button-123")

	assert.True(t, found)
	assert.Equal(t, 120, x)
	assert.Equal(t, 30, y)
	assert.Equal(t, 200, w)
	assert.Equal(t, 50, h)
}

// TestFindAnchorPosition_NotFound tests finding non-existent anchor
func TestFindAnchorPosition_NotFound(t *testing.T) {
	root := &LayoutBox{
		ID:     "root",
		X:      0,
		Y:      0,
		Width:  500,
		Height: 300,
	}

	x, y, w, h, found := FindAnchorPosition(root, "non-existent")

	assert.False(t, found)
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
}

// TestPortalPositionConfig_SetAnchorPosition tests setting anchor position
func TestPortalPositionConfig_SetAnchorPosition(t *testing.T) {
	config := PortalPositionConfig{
		PortalWidth:  100,
		PortalHeight: 50,
	}

	config.SetAnchorPosition(50, 30, 200, 60)

	assert.Equal(t, 50, config.AnchorX)
	assert.Equal(t, 30, config.AnchorY)
	assert.Equal(t, 200, config.AnchorWidth)
	assert.Equal(t, 60, config.AnchorHeight)
}

// TestPortalPositionConfig_String tests string representation
func TestPortalPositionConfig_String(t *testing.T) {
	config := PortalPositionConfig{
		Position:      types.PositionFixed,
		Anchor:        types.AnchorCenter,
		AnchorID:      "test-anchor",
		Left:          ptrInt(10),
		Top:           ptrInt(20),
		ViewportWidth: 800,
		ViewportHeight: 600,
		PortalWidth:   400,
		PortalHeight:  300,
		AnchorX:       100,
		AnchorY:       200,
		AnchorWidth:   200,
		AnchorHeight:  100,
	}

	s := config.String()
	assert.Contains(t, s, "fixed")  // PositionFixed.String() returns "fixed"
	assert.Contains(t, s, "center") // AnchorCenter.String() returns "center"
	assert.Contains(t, s, "test-anchor")
	assert.Contains(t, s, "10")  // Left
	assert.Contains(t, s, "20")  // Top
	assert.Contains(t, s, "800x600")  // Viewport
	assert.Contains(t, s, "400x300")  // Portal
	assert.Contains(t, s, "(100,200-200x100)")  // Anchor
}

// TestCalculateFixedPosition_AllAnchors tests all PositionFixed anchor variants
func TestCalculateFixedPosition_AllAnchors(t *testing.T) {
	vw, vh := 800, 600
	pw, ph := 400, 300

	calculator := NewPortalPositionCalculator()

	tests := []struct {
		name         string
		anchor       types.Anchor
		expectedX    int
		expectedY    int
	}{
		{"TopLeft", types.AnchorTopLeft, 0, 0},
		{"Top", types.AnchorTop, (vw - pw) / 2, 0},
		{"TopRight", types.AnchorTopRight, vw - pw, 0},
		{"Left", types.AnchorLeft, 0, (vh - ph) / 2},
		{"Center", types.AnchorCenter, (vw - pw) / 2, (vh - ph) / 2},
		{"Right", types.AnchorRight, vw - pw, (vh - ph) / 2},
		{"BottomLeft", types.AnchorBottomLeft, 0, vh - ph},
		{"Bottom", types.AnchorBottom, (vw - pw) / 2, vh - ph},
		{"BottomRight", types.AnchorBottomRight, vw - pw, vh - ph},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PortalPositionConfig{
				Position:      types.PositionFixed,
				Anchor:        tt.anchor,
				ViewportWidth: vw,
				ViewportHeight: vh,
				PortalWidth:   pw,
				PortalHeight:  ph,
			}

			x, y := calculator.CalculatePosition(config)
			assert.Equal(t, tt.expectedX, x, "X mismatch for "+tt.name)
			assert.Equal(t, tt.expectedY, y, "Y mismatch for "+tt.name)
		})
	}
}

// TestCalculateAnchorBasedPosition_AllAnchors tests all Anchor-based positioning variants
func TestCalculateAnchorBasedPosition_AllAnchors(t *testing.T) {
	ax, ay, aw, ah := 100, 100, 200, 50
	pw, ph := 150, 40

	calculator := NewPortalPositionCalculator()

	tests := []struct {
		name      string
		anchor    types.Anchor
		expectedX int
		expectedY int
	}{
		{"TopLeft", types.AnchorTopLeft, ax, ay},
		{"Top", types.AnchorTop, ax + (aw-pw)/2, ay},
		{"TopRight", types.AnchorTopRight, ax + aw - pw, ay},
		{"Left", types.AnchorLeft, ax, ay + (ah-ph)/2},
		{"Center", types.AnchorCenter, ax + (aw-pw)/2, ay + (ah-ph)/2},
		{"Right", types.AnchorRight, ax + aw - pw, ay + (ah-ph)/2},
		{"BottomLeft", types.AnchorBottomLeft, ax, ay + ah - ph},
		{"Bottom", types.AnchorBottom, ax + (aw-pw)/2, ay + ah - ph},
		{"BottomRight", types.AnchorBottomRight, ax + aw - pw, ay + ah - ph},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PortalPositionConfig{
				Position:    types.PositionAbsolute,
				Anchor:      tt.anchor,
				AnchorX:     ax,
				AnchorY:     ay,
				AnchorWidth: aw,
				AnchorHeight: ah,
				PortalWidth: pw,
				PortalHeight: ph,
			}

			x, y := calculator.CalculatePosition(config)
			assert.Equal(t, tt.expectedX, x, "X mismatch for "+tt.name)
			assert.Equal(t, tt.expectedY, y, "Y mismatch for "+tt.name)
		})
	}
}

// Helper function to create int pointer
func ptrInt(v int) *int {
	return &v
}
