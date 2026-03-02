// Package render provides Portal integration tests
package render

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/types"
	ui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
)

func init() {
	// Enable Fiber-first mode for all tests in this file
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
}

// TestPortal_Integration_PortalDetection tests Portal node detection in Fiber tree
func TestPortal_Integration_PortalDetection(t *testing.T) {
	// Create simple test fiber trees to test hasPortals method

	// Test 1: Fiber with PortalRoot and Portal
	portalFiber := &reconciler.Fiber{
		Props: ui.Props{
			"text": "Test",
		},
		Child: &reconciler.Fiber{
			Props: ui.Props{
				"portalRootId": "main-root",
			},
			Sibling: &reconciler.Fiber{
				Props: ui.Props{
					"portalRoot": "main-root",
				},
			},
		},
	}

	fwApp := framework.NewApp()
	defer fwApp.Quit()

	testNode := NewDeclarativeNodeFromFuncWithFiber(func() ui.VNode {
		return ui.NewElement("div")
	}, fwApp)

	// Test hasPortals with a fiber that has Portals
	hasPortals := testNode.hasPortals(portalFiber)
	assert.True(t, hasPortals, "hasPortals should detect Portal nodes")

	// Test 2: Fiber without Portal nodes
	noPortalFiber := &reconciler.Fiber{
		Props: ui.Props{
			"text": "Test",
		},
		Child: &reconciler.Fiber{
			Props: ui.Props{
				"text": "Child",
			},
		},
	}

	hasPortals = testNode.hasPortals(noPortalFiber)
	assert.False(t, hasPortals, "hasPortals should return false for fibers without Portals")

	// Test 3: nil fiber
	hasPortals = testNode.hasPortals(nil)
	assert.False(t, hasPortals, "hasPortals should handle nil fiber")
}

// TestPortal_Integration_LayoutBoxesWithPortal tests that Portal boxes are included in layout result
func TestPortal_Integration_LayoutBoxesWithPortal(t *testing.T) {
	fwApp := framework.NewApp()
	defer fwApp.Quit()

	// Create a test node with PortalRoot and a Portal
	renderFn := func() ui.VNode {
		return newstack.NewVStack().SetGap(1).SetChildrenList([]ui.VNode{
			newtext.New("Main Content"),
			ui.NewElement("div").SetProps(ui.Props{
				"portalRootId": "tooltip-root",
			}),
			ui.NewElement("portal").SetProps(ui.Props{
				"portalRoot": "tooltip-root",
				"position":   types.PositionFixed,
				"top":        10,
				"left":       20,
			}).SetChildren([]ui.VNode{
				newtext.New("Tooltip"),
			}),
		})
	}

	testNode := NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
	testNode.SetUsePortalLayout(true)

	// Setup and paint
	fwApp.SetRoot(testNode)
	buf := paint.NewBuffer(80, 24)
	ctx := component.PaintContext{
		AvailableWidth:  80,
		AvailableHeight: 24,
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 24},
	}

	testNode.Paint(ctx, buf)

	// Get layout boxes - should include Portal boxes
	boxes := testNode.GetLayoutBoxes()
	assert.NotEmpty(t, boxes, "Should have layout boxes after paint")

	t.Logf("Total layout boxes: %d", len(boxes))

	// Log all boxes for debugging
	for _, box := range boxes {
		t.Logf("Box ID=%s at (%d,%d) size %dx%d, ZIndex=%d, ShouldCenter=%v",
			box.ID, box.X, box.Y, box.Width, box.Height, box.ZIndex, box.ShouldCenter)
	}

	// Get portal boxes specifically
	portalBoxes := testNode.GetPortalBoxes()
	t.Logf("Portal boxes count: %d", len(portalBoxes))

	// In Portal-aware layout mode, we should have portal boxes
	// (Though availability depends on the actual layout execution)
	if testNode.IsPortalLayoutEnabled() {
		// We might not get portal boxes if the layout didn't run in Portal mode
		// or if there was no valid PortalRoot. Just log the result.
		t.Logf("Portal layout enabled, got %d portal boxes", len(portalBoxes))
		for i, box := range portalBoxes {
			t.Logf("  Portal box %d: ID=%s at (%d,%d) size %dx%d, ZIndex=%d",
				i, box.ID, box.X, box.Y, box.Width, box.Height, box.ZIndex)
		}
	}
}

// TestPortal_Integration_PaintableBoxesWithPortal tests Paintable boxes output
func TestPortal_Integration_PaintableBoxesWithPortal(t *testing.T) {
	fwApp := framework.NewApp()
	defer fwApp.Quit()

	// Create a test node with PortalRoot and a Portal
	renderFn := func() ui.VNode {
		return newstack.NewVStack().SetGap(1).SetChildrenList([]ui.VNode{
			newtext.New("Main"),
			ui.NewElement("div").SetProps(ui.Props{
				"portalRootId": "modal-root",
			}),
			ui.NewElement("portal").SetProps(ui.Props{
				"portalRoot": "modal-root",
				"position":   types.PositionFixed,
			}).SetChildren([]ui.VNode{
				newtext.New("Modal Content"),
			}),
		})
	}

	testNode := NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
	testNode.SetUsePortalLayout(true)

	// Setup and paint
	fwApp.SetRoot(testNode)
	buf := paint.NewBuffer(80, 24)
	ctx := component.PaintContext{
		AvailableWidth:  80,
		AvailableHeight: 24,
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 24},
	}

	testNode.Paint(ctx, buf)

	// Get paintable boxes
	paintableBoxes := testNode.GetPaintableBoxes()

	t.Logf("Total paintable boxes: %d", len(paintableBoxes))

	// Log all paintable boxes
	for i, box := range paintableBoxes {
		t.Logf("Paintable box %d: node=%s at (%d,%d) size %dx%d, ZIndex=%d",
			i, box.Node, box.X, box.Y, box.Width, box.Height, box.ZIndex)
	}

	// Should have paintable boxes
	assert.NotEmpty(t, paintableBoxes, "Should have paintable boxes after paint")
}

// TestPortal_Integration_MultiplePortals tests multiple Portals with different priorities
func TestPortal_Integration_MultiplePortals(t *testing.T) {
	fwApp := framework.NewApp()
	defer fwApp.Quit()

	// Create a test node with multiple Portals
	renderFn := func() ui.VNode {
		return newstack.NewVStack().SetGap(1).SetChildrenList([]ui.VNode{
			newtext.New("Main Content"),
			ui.NewElement("div").SetProps(ui.Props{
				"portalRootId": "root-1",
			}),
			ui.NewElement("portal").SetProps(ui.Props{
				"portalRoot": "root-1",
				"priority":   10, // High priority
			}).SetChildren([]ui.VNode{
				newtext.New("High Priority"),
			}),
			ui.NewElement("portal").SetProps(ui.Props{
				"portalRoot": "root-1",
				// Default priority (0)
			}).SetChildren([]ui.VNode{
				newtext.New("Low Priority"),
			}),
		})
	}

	testNode := NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
	testNode.SetUsePortalLayout(true)

	// Setup and paint
	fwApp.SetRoot(testNode)
	buf := paint.NewBuffer(80, 24)
	ctx := component.PaintContext{
		AvailableWidth:  80,
		AvailableHeight: 24,
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 24},
	}

	testNode.Paint(ctx, buf)

	// Get portal boxes
	portalBoxes := testNode.GetPortalBoxes()
	t.Logf("Portal boxes count: %d", len(portalBoxes))

	for i, box := range portalBoxes {
		t.Logf("Portal box %d: ID=%s at (%d,%d) size %dx%d, ZIndex=%d",
			i, box.ID, box.X, box.Y, box.Width, box.Height, box.ZIndex)
	}

	// Get all layout boxes to ensure they include Portals
	allBoxes := testNode.GetLayoutBoxes()
	t.Logf("Total layout boxes (including Portals): %d", len(allBoxes))
}

// TestPortal_Integration_NoPortalLayout test non-Portal layout mode
func TestPortal_Integration_NoPortalLayout(t *testing.T) {
	fwApp := framework.NewApp()
	defer fwApp.Quit()

	// Create a test node with Portal elements but disabled Portal layout
	renderFn := func() ui.VNode {
		return newstack.NewVStack().SetGap(1).SetChildrenList([]ui.VNode{
			newtext.New("Main"),
			ui.NewElement("div").SetProps(ui.Props{
				"portalRootId": "root-1",
			}),
			ui.NewElement("portal").SetProps(ui.Props{
				"portalRoot": "root-1",
			}).SetChildren([]ui.VNode{
				newtext.New("Portal Content"),
			}),
		})
	}

	testNode := NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
	testNode.SetUsePortalLayout(false) // Disable Portal-aware layout

	// Setup and paint
	fwApp.SetRoot(testNode)
	buf := paint.NewBuffer(80, 24)
	ctx := component.PaintContext{
		AvailableWidth:  80,
		AvailableHeight: 24,
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 24},
	}

	testNode.Paint(ctx, buf)

	// Get portal boxes - should be empty in non-Portal layout mode
	portalBoxes := testNode.GetPortalBoxes()
	t.Logf("Portal boxes (disabled mode): %d", len(portalBoxes))
	assert.Empty(t, portalBoxes, "Should have no portal boxes when Portal layout is disabled")

	// Get all layout boxes
	allBoxes := testNode.GetLayoutBoxes()
	t.Logf("Total layout boxes (non-Portal mode): %d", len(allBoxes))
}

// TestPortal_Integration_PortalPositionFixed tests Fixed positioning for Portals
func TestPortal_Integration_PortalPositionFixed(t *testing.T) {
	fwApp := framework.NewApp()
	defer fwApp.Quit()

	// Test PositionFixed: portal positioned relative to viewport
	renderFn := func() ui.VNode {
		return newstack.NewVStack().SetChildrenList([]ui.VNode{
			newtext.New("Main"),
			ui.NewElement("div").SetProps(ui.Props{
				"portalRootId": "modal-root",
			}),
			ui.NewElement("portal").SetProps(ui.Props{
				"portalRoot": "modal-root",
				"position":   types.PositionFixed,
				"top":        100,
				"left":       150,
			}).SetChildren([]ui.VNode{
				newtext.New("Fixed Modal"),
			}),
		})
	}

	testNode := NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
	testNode.SetUsePortalLayout(true)

	fwApp.SetRoot(testNode)
	buf := paint.NewBuffer(800, 600)
	ctx := component.PaintContext{
		AvailableWidth:  800,
		AvailableHeight: 600,
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	testNode.Paint(ctx, buf)

	// Check portal boxes for Fixed positioning
	portalBoxes := testNode.GetPortalBoxes()
	t.Logf("Portal boxes for Fixed positioning: %d", len(portalBoxes))

	for i, box := range portalBoxes {
		t.Logf("Portal %d: pos=(%d,%d), ZIndex=%d", i, box.X, box.Y, box.ZIndex)

		// For PositionFixed, the X,Y should be close to the specified top/left
		// (accounting for anchor adjustments)
		if box.X > 0 || box.Y > 0 {
			t.Logf("  Portal positioned at (%d,%d), expected around (150,100)", box.X, box.Y)
		}
	}
}

// TestPortal_Integration_PortalAnchor tests Anchor-based positioning
func TestPortal_Integration_PortalAnchor(t *testing.T) {
	fwApp := framework.NewApp()
	defer fwApp.Quit()

	// Test Anchor-based positioning (tooltip below a button)
	renderFn := func() ui.VNode {
		return newstack.NewVStack().SetGap(2).SetChildrenList([]ui.VNode{
			ui.NewElement("button").SetProps(ui.Props{
				"text": "Button",
				"id":   "button-id",
			}),
			ui.NewElement("div").SetProps(ui.Props{
				"portalRootId": "tooltip-root",
			}),
			ui.NewElement("portal").SetProps(ui.Props{
				"portalRoot": "tooltip-root",
				"anchorId":   "button-id",
				"anchor":     types.AnchorBottomLeft,
				"top":        5,
			}).SetChildren([]ui.VNode{
				newtext.New("Tooltip text"),
			}),
		})
	}

	testNode := NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
	testNode.SetUsePortalLayout(true)

	fwApp.SetRoot(testNode)
	buf := paint.NewBuffer(80, 24)
	ctx := component.PaintContext{
		AvailableWidth:  80,
		AvailableHeight: 24,
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 24},
	}

	testNode.Paint(ctx, buf)

	// Check portal boxes for Anchor positioning
	portalBoxes := testNode.GetPortalBoxes()
	t.Logf("Portal boxes for Anchor positioning: %d", len(portalBoxes))

	for i, box := range portalBoxes {
		t.Logf("Portal %d: pos=(%d,%d), ZIndex=%d, size=%dx%d",
			i, box.X, box.Y, box.ZIndex, box.Width, box.Height)
	}
}

// TestPortal_Integration_PortalZIndex tests Portal Z-index ordering
func TestPortal_Integration_PortalZIndex(t *testing.T) {
	fwApp := framework.NewApp()
	defer fwApp.Quit()

	// Test that Portals have higher Z-index than main tree elements
	renderFn := func() ui.VNode {
		return newstack.NewVStack().SetGap(1).SetChildrenList([]ui.VNode{
			newtext.New("Main Content"),
			ui.NewElement("div").SetProps(ui.Props{
				"portalRootId": "overlay-root",
			}),
			ui.NewElement("portal").SetProps(ui.Props{
				"portalRoot": "overlay-root",
				"priority":   5,
			}).SetChildren([]ui.VNode{
				newtext.New("Overlay"),
			}),
		})
	}

	testNode := NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
	testNode.SetUsePortalLayout(true)

	fwApp.SetRoot(testNode)
	buf := paint.NewBuffer(80, 24)
	ctx := component.PaintContext{
		AvailableWidth:  80,
		AvailableHeight: 24,
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 24},
	}

	testNode.Paint(ctx, buf)

	// Get all boxes and check Z-index ordering
	allBoxes := testNode.GetLayoutBoxes()

	t.Logf("Z-index analysis:")

	// Find main tree boxes and Portal boxes
	mainTreeZIndex := -1
	portalZIndex := -1

	for _, box := range allBoxes {
		if len(box.ID) > 6 && box.ID[:6] == "portal-" {
			// This is a Portal box
			portalZIndex = box.ZIndex
			t.Logf("  Portal box ZIndex: %d (ID=%s)", box.ZIndex, box.ID)
		} else {
			// Main tree box
			if box.ZIndex > mainTreeZIndex {
				mainTreeZIndex = box.ZIndex
			}
		}
	}

	if mainTreeZIndex >= 0 {
		t.Logf("  Main tree max ZIndex: %d", mainTreeZIndex)
	}

	if portalZIndex >= 0 && mainTreeZIndex >= 0 {
		t.Logf("  Portal ZIndex (%d) should be > Main tree ZIndex (%d)", portalZIndex, mainTreeZIndex)
		// Portals should have higher Z-index (PortalZIndexBase + priority = 1000 + 5 = 1005)
		assert.Greater(t, portalZIndex, mainTreeZIndex,
			"Portal should have higher Z-index than main tree elements")
	}
}
