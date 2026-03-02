// Package render tests Portal HitMap functionality
package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime/layout"
)

// TestPortalZIndex test that Portals use separate Z-index range
func TestPortalZIndex(t *testing.T) {
	// Main tree layout box
	mainTreeBox := &layout.LayoutBox{
		ID:     "main-1",
		X:      0,
		Y:      0,
		Width:  100,
		Height: 100,
		Children: []*layout.LayoutBox{
			{
				ID:     "main-2",
				X:      10,
				Y:      10,
				Width:  50,
				Height: 50,
				Children: []*layout.LayoutBox{
					{
						ID:     "main-3",
						X:      20,
						Y:      20,
						Width:  20,
						Height: 20,
					},
				},
			},
		},
	}

	// Portal box with ZIndexBase + priority
	portalBox := &layout.LayoutBox{
		ID:     "portal-1",
		X:      50,
		Y:      50,
		Width:  80,
		Height: 60,
		ZIndex: 1005, // PortalZIndexBase (1000) + priority (5)
		Children: []*layout.LayoutBox{
			{
				ID:     "portal-content",
				X:      10,
				Y:      10,
				Width:  60,
				Height: 40,
				ZIndex: 1006, // Base + 5 + 1
			},
		},
	}

	// Build HitMap from combined tree
	combinedBox := &layout.LayoutBox{
		ID:     "root",
		X:      0,
		Y:      0,
		Width:  200,
		Height: 200,
		Children: []*layout.LayoutBox{
			mainTreeBox,
			portalBox,
		},
	}

	hitMap := layout.NewHitMap()
	hitMap.BuildFromLayoutBox(combinedBox)

	// Test 1: Portal should have higher ZIndex than main tree
	portalEntry := hitMap.Get("portal-1")
	assert.NotNil(t, portalEntry)
	assert.Equal(t, 1005, portalEntry.ZIndex, "Portal should have ZIndex 1005")

	mainEntry2 := hitMap.Get("main-2")
	assert.NotNil(t, mainEntry2)
	assert.Less(t, mainEntry2.ZIndex, portalEntry.ZIndex, "Main tree ZIndex should be less than Portal")

	// Test 2: Hit point where Portal and main tree overlap
	hitX, hitY := 55, 55 // Overlap region
	entry := hitMap.HitTest(hitX, hitY)

	assert.NotNil(t, entry)
	assert.Equal(t, "portal-1", entry.NodeID, "Portal should win HitTest")
}

// TestPortalZIndexMultiplePortals tests multiple Portals with different priorities
func TestPortalZIndexMultiplePortals(t *testing.T) {
	portal1 := &layout.LayoutBox{
		ID:     "portal-1",
		X:      10,
		Y:      10,
		Width:  50,
		Height: 50,
		ZIndex: 1000, // Base + priority 0
	}

	portal2 := &layout.LayoutBox{
		ID:     "portal-2",
		X:      30,
		Y:      30,
		Width:  50,
		Height: 50,
		ZIndex: 1005, // Base + priority 5
	}

	portal3 := &layout.LayoutBox{
		ID:     "portal-3",
		X:      20,
		Y:      20,
		Width:  50,
		Height: 50,
		ZIndex: 1010, // Base + priority 10
	}

	combinedBox := &layout.LayoutBox{
		ID:     "root",
		Children: []*layout.LayoutBox{portal1, portal2, portal3},
	}

	hitMap := layout.NewHitMap()
	hitMap.BuildFromLayoutBox(combinedBox)

	// Test: Overlap point (30,30) - all three portals overlap
	entry := hitMap.HitTest(30, 30)

	assert.NotNil(t, entry)
	assert.Equal(t, "portal-3", entry.NodeID, "Highest priority Portal should win")
}

// TestPortalZIndexChildren verifies children inherit proper ZIndex via setPortalZIndex
func TestPortalZIndexChildren(t *testing.T) {
	parent := &layout.LayoutBox{
		ID:     "portal-parent",
		X:      0,
		Y:      0,
		Width:  100,
		Height: 100,
		Children: []*layout.LayoutBox{
			{
				ID:     "child-1",
				X:      10,
				Y:      10,
				Width:  30,
				Height: 30,
			},
			{
				ID:     "child-2",
				X:      50,
				Y:      50,
				Width:  30,
				Height: 30,
			},
		},
	}

	// Test setPortalZIndex method
	engine := NewPortalAwareLayoutEngine()
	engine.setPortalZIndex(parent, 1000)

	// Verify ZIndex values
	assert.Equal(t, 1000, parent.ZIndex, "Portal root ZIndex should be 1000")
	assert.Equal(t, 1001, parent.Children[0].ZIndex, "First child ZIndex should be 1001")
	assert.Equal(t, 1002, parent.Children[1].ZIndex, "Second child ZIndex should be 1002")
}

// TestCalculatePortalZIndex tests the setPortalZIndex method with nested children
func TestCalculatePortalZIndex(t *testing.T) {
	// Create a mock engine
	engine := NewPortalAwareLayoutEngine()

	// Create a portal box tree
	portalBox := &layout.LayoutBox{
		ID:     "portal",
		Width:  100,
		Height: 100,
		Children: []*layout.LayoutBox{
			{
				ID:     "child1",
				Width:  50,
				Height:  50,
				Children: []*layout.LayoutBox{
					{
						ID:     "grandchild",
						Width:  20,
						Height: 20,
					},
				},
			},
			{
				ID:     "child2",
				Width:  30,
				Height:  30,
			},
		},
	}

	// Set ZIndex starting from PortalZIndexBase (1000)
	engine.setPortalZIndex(portalBox, 1000)

	// Verify ZIndex values
	assert.Equal(t, 1000, portalBox.ZIndex, "Portal root ZIndex should be 1000")
	assert.Equal(t, 1001, portalBox.Children[0].ZIndex, "First child ZIndex should be 1001")
	assert.Equal(t, 1002, portalBox.Children[1].ZIndex, "Second child ZIndex should be 1002")
	assert.Equal(t, 1003, portalBox.Children[0].Children[0].ZIndex, "Grandchild ZIndex should be 1003")
}
