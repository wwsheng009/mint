package modal

import (
	"testing"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestModal_SyncPositioningProperties tests that syncPositioningProperties sets correct Position and Anchor
func TestModal_SyncPositioningProperties(t *testing.T) {
	tests := []struct {
		name         string
		centered     bool
		explicitPos  string
		expectedPos  types.PositionType
		expectedAnch types.Anchor
	}{
		{
			name:         "Centered modal uses Fixed+Center",
			centered:     true,
			explicitPos:  "",
			expectedPos:  types.PositionFixed,
			expectedAnch: types.AnchorCenter,
		},
		{
			name:         "Non-centered modal uses Relative+TopLeft",
			centered:     false,
			explicitPos:  "",
			expectedPos:  types.PositionRelative,
			expectedAnch: types.AnchorTopLeft,
		},
		{
			name:         "Explicit absolute position overrides centered",
			centered:     true,
			explicitPos:  "absolute",
			expectedPos:  types.PositionAbsolute,
			expectedAnch: types.AnchorTopLeft, // anchor prop not set, default TopLeft
		},
		{
			name:         "Explicit fixed position without centered",
			centered:     false,
			explicitPos:  "fixed",
			expectedPos:  types.PositionFixed,
			expectedAnch: types.AnchorTopLeft, // anchor prop not set, default TopLeft
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Fiber node from Modal VNode
			modalVNode := New().SetTitle("Test").SetCentered(tt.centered)
			modalVNode.SetOpen(true) // Set isOpen so modal is visible

			fiber := reconciler.CreateFiberFromVNode(modalVNode)

			// If explicit position is specified, set it in Props directly
			// (Modal.SetProps doesn't have position handling, but Props() includes it)
			if tt.explicitPos != "" {
				fiber.Props["position"] = tt.explicitPos
			}

			// Apply syncPositioningProperties
			reconciler.SyncPositioningProperties(fiber)

			// Verify Position
			if fiber.Position != tt.expectedPos {
				t.Errorf("Fiber.Position = %v, expected %v", fiber.Position, tt.expectedPos)
			}

			// Verify Anchor (only if position is Fixed)
			if tt.expectedPos == types.PositionFixed {
				if fiber.Anchor != tt.expectedAnch {
					t.Errorf("Fiber.Anchor = %v, expected %v", fiber.Anchor, tt.expectedAnch)
				}
			}

			t.Logf("Modal: centered=%v, position=%s => Position=%v, Anchor=%v",
				tt.centered, tt.explicitPos, fiber.Position, fiber.Anchor)
		})
	}
}

// TestModal_PositionInheritance tests that modal can use different anchors
func TestModal_PositionInheritance(t *testing.T) {
	// Test explicit anchor overrides centered behavior
	modalVNode := New().SetCentered(true)
	modalVNode.SetProps(rtui.Props{
		"anchortype": "TopRight", // Note: actual prop key would depend on implementation
	})

	fiber := reconciler.CreateFiberFromVNode(modalVNode)
	reconciler.SyncPositioningProperties(fiber)

	// If anchor prop is set explicitly, it should override the centered anchor
	// This test documents the expected behavior
	t.Logf("Modal with centered=true but explicit anchor: Position=%v, Anchor=%v",
		fiber.Position, fiber.Anchor)
}

// TestModal_LayoutCentering tests that layout engine calculates correct position for centered modal
func TestModal_LayoutCentering(t *testing.T) {
	// Create a mock modal node with PositionFixed + AnchorCenter
	modalNode := &mockPositionNode{
		width:        40,
		height:       12,
		positionType: layout.PositionFixed,
		anchor:       layout.AnchorCenter,
	}

	tests := []struct {
		name   string
		vw, vh int // viewport dimensions
		expectX, expectY int
	}{
		{
			name:    "80x40 viewport",
			vw:      80,
			vh:      40,
			expectX: 20, // (80 - 40) / 2
			expectY: 14, // (40 - 12) / 2
		},
		{
			name:    "100x50 viewport",
			vw:      100,
			vh:      50,
			expectX: 30, // (100 - 40) / 2
			expectY: 19, // (50 - 12) / 2
		},
		{
			name:    "120x60 viewport",
			vw:      120,
			vh:      60,
			expectX: 40, // (120 - 40) / 2
			expectY: 24, // (60 - 12) / 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints := layout.Constraints{
				MinWidth:  0,
				MaxWidth:  tt.vw,
				MinHeight: 0,
				MaxHeight: tt.vh,
			}

			engine := layout.NewEngine()
			result := engine.Layout(modalNode, constraints)

			if result.Root == nil {
				t.Fatal("Result Root is nil")
			}

			t.Logf("Layout: X=%d, Y=%d (expected: X=%d, Y=%d)",
				result.Root.X, result.Root.Y, tt.expectX, tt.expectY)

			if result.Root.X != tt.expectX {
				t.Errorf("X position = %d, expected %d", result.Root.X, tt.expectX)
			}
			if result.Root.Y != tt.expectY {
				t.Errorf("Y position = %d, expected %d", result.Root.Y, tt.expectY)
			}
		})
	}
}

// mockPositionNode is a mock Node that implements PositionProvider
type mockPositionNode struct {
	width, height  int
	positionType   layout.PositionType
	anchor         layout.Anchor
}

func (m *mockPositionNode) ID() string { return "mock-modal" }
func (m *mockPositionNode) Type() string { return "modal" }
func (m *mockPositionNode) Children() []layout.Node { return nil }
func (m *mockPositionNode) GetPosition() (int, int) { return 0, 0 }
func (m *mockPositionNode) SetPosition(x, y int) {}
func (m *mockPositionNode) GetSize() (int, int) { return m.width, m.height }
func (m *mockPositionNode) SetSize(w, h int) {}
func (m *mockPositionNode) GetWidth() int { return m.width }
func (m *mockPositionNode) GetHeight() int { return m.height }

// Measurable interface
func (m *mockPositionNode) Measure(constraints layout.Constraints) layout.Size {
	return layout.Size{Width: m.width, Height: m.height}
}

// PositionProvider interface
func (m *mockPositionNode) GetPositionType() layout.PositionType {
	return m.positionType
}

func (m *mockPositionNode) GetAnchor() layout.Anchor {
	return m.anchor
}

// Layered interface
func (m *mockPositionNode) GetLayer() layout.Layer {
	return layout.LayerModal
}

func (m *mockPositionNode) GetZIndex() int {
	return 100
}

// TestModal_AllAnchors tests all anchor types with PositionFixed
func TestModal_AllAnchors(t *testing.T) {
	width := 40
	height := 12
	vw := 100
	vh := 50

	tests := []struct {
		name    string
		anchor  layout.Anchor
		expectX int
		expectY int
	}{
		{"AnchorTopLeft", layout.AnchorTopLeft, 0, 0},
		{"AnchorTop", layout.AnchorTop, 30, 0}, // (100 - 40) / 2
		{"AnchorTopRight", layout.AnchorTopRight, 60, 0}, // 100 - 40
		{"AnchorLeft", layout.AnchorLeft, 0, 19}, // (50 - 12) / 2
		{"AnchorCenter", layout.AnchorCenter, 30, 19}, // (100-40)/2, (50-12)/2
		{"AnchorRight", layout.AnchorRight, 60, 19}, // 100-40, (50-12)/2
		{"AnchorBottomLeft", layout.AnchorBottomLeft, 0, 38}, // 50-12
		{"AnchorBottom", layout.AnchorBottom, 30, 38}, // (100-40)/2, 50-12
		{"AnchorBottomRight", layout.AnchorBottomRight, 60, 38}, // 100-40, 50-12
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modalNode := &mockPositionNode{
				width:        width,
				height:       height,
				positionType: layout.PositionFixed,
				anchor:       tt.anchor,
			}

			constraints := layout.Constraints{
				MinWidth:  0,
				MaxWidth:  vw,
				MinHeight: 0,
				MaxHeight: vh,
			}

			engine := layout.NewEngine()
			result := engine.Layout(modalNode, constraints)

			if result.Root.X != tt.expectX {
				t.Errorf("X position = %d, expected %d", result.Root.X, tt.expectX)
			}
			if result.Root.Y != tt.expectY {
				t.Errorf("Y position = %d, expected %d", result.Root.Y, tt.expectY)
			}

			t.Logf("Anchor %s: X=%d, Y=%d", tt.anchor, result.Root.X, result.Root.Y)
		})
	}
}
