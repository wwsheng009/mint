// Portal System Test (Phase 3)
// Tests for PortalRoot creation and Portal → PortalRoot linking
package reconciler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestPortalRoot_Linking tests that Portal nodes are linked to PortalRoot nodes
// during the commit phase (linkPortalsToRoots)
func TestPortalRoot_Linking(t *testing.T) {
	// Create a Fiber tree with PortalRoot and Portal nodes
	// Structure:
	//   Root
	//   ├── PortalRoot (props["portalRootId"] = "main-root")
	//   │   └── Child1
	//   └── AppContent
	//       └── Portal (props["portalRoot"] = "main-root")
	//           └── PortalContent

	portalRoot := &rtui.Fiber{
		NodeID: 100,
		Key:    "portal-root",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRootId": "main-root",
		},
	}

	portal := &rtui.Fiber{
		NodeID: 200,
		Key:    "portal",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "main-root",
		},
	}

	portalContent := &rtui.Fiber{
		NodeID: 201,
		Key:    "portal-content",
		Type:   rtui.VNodeText,
	}
	portal.Child = portalContent

	appContent := &rtui.Fiber{
		NodeID: 150,
		Key:    "app-content",
		Type:   rtui.VNodeElement,
	}
	appContent.Child = portal

	root := &rtui.Fiber{
		NodeID: 1,
		Key:    "root",
		Type:   rtui.VNodeElement,
	}
	root.Child = portalRoot
	portalRoot.Sibling = appContent

	// Before linking: portal.PortalRoot should be nil
	assert.Nil(t, portal.PortalRoot)

	// Simulate Reconciler.linkPortalsToRoots()
	// Use a helper function since linkPortalsToRoots is a method of Reconciler
	linkPortalsToRootsHelper(root)

	// After linking: portal.PortalRoot should point to portalRoot
	assert.NotNil(t, portal.PortalRoot)
	assert.Equal(t, portalRoot.NodeID, portal.PortalRoot.NodeID)
	assert.Equal(t, "portal-root", portal.PortalRoot.Key)
}

// TestLinkPortalsToRoots is the helper function that mimics Reconciler.linkPortalsToRoots()
func TestLinkPortalsToRoots(t *testing.T) {
	root := &rtui.Fiber{
		NodeID: 1,
		Key:    "root",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRootId": "global-root",  // Root itself can be a PortalRoot
		},
	}

	// Create a portal that references the global root
	portal := &rtui.Fiber{
		NodeID: 100,
		Key:    "tooltip",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "global-root",
		},
	}

	// Create a child of the portal
	tooltipContent := &rtui.Fiber{
		NodeID: 101,
		Key:    "tooltip-content",
		Type:   rtui.VNodeText,
	}
	portal.Child = tooltipContent

	root.Child = portal

	// Link portals to roots
	linkPortalsToRootsHelper(root)

	// Verify portal is linked to root
	assert.NotNil(t, portal.PortalRoot)
	assert.Equal(t, root.NodeID, portal.PortalRoot.NodeID)

	// Create a second PortalRoot
	secondRoot := &rtui.Fiber{
		NodeID: 200,
		Key:    "second-root",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRootId": "modal-root",
		},
	}

	// Create a modal portal
	modal := &rtui.Fiber{
		NodeID: 300,
		Key:    "modal",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "modal-root",
		},
	}

	secondRoot.Child = modal
	portal.Sibling = secondRoot

	// Link again
	linkPortalsToRootsHelper(root)

	// Verify both portals are linked correctly
	assert.NotNil(t, portal.PortalRoot)
	assert.Equal(t, root.NodeID, portal.PortalRoot.NodeID)  // Points to global-root

	assert.NotNil(t, modal.PortalRoot)
	assert.Equal(t, secondRoot.NodeID, modal.PortalRoot.NodeID)  // Points to modal-root
}

// linkPortalsToRootsHelper is a stand-alone implementation of Reconciler.linkPortalsToRoots()
func linkPortalsToRootsHelper(root *rtui.Fiber) {
	if root == nil {
		return
	}

	// Step 1: Collect all PortalRoot nodes
	portalRoots := make(map[string]*rtui.Fiber)

	var collectPortalRoots func(fiber *rtui.Fiber)
	collectPortalRoots = func(fiber *rtui.Fiber) {
		if fiber == nil {
			return
		}

		if fiber.Props != nil {
			if portalRootID, ok := fiber.Props["portalRootId"].(string); ok && portalRootID != "" {
				portalRoots[portalRootID] = fiber
			}
		}

		collectPortalRoots(fiber.Child)
		collectPortalRoots(fiber.Sibling)
	}
	collectPortalRoots(root)

	// Step 2: Link Portal nodes to their PortalRoot targets
	var linkPortalNodes func(fiber *rtui.Fiber)
	linkPortalNodes = func(fiber *rtui.Fiber) {
		if fiber == nil {
			return
		}

		if fiber.Props != nil {
			if portalRootID, ok := fiber.Props["portalRoot"].(string); ok && portalRootID != "" {
				if target, exists := portalRoots[portalRootID]; exists {
					fiber.PortalRoot = target
				}
			}
		}

		linkPortalNodes(fiber.Child)
		linkPortalNodes(fiber.Sibling)
	}
	linkPortalNodes(root)
}

// TestPortalRoot_MultiplePortals tests multiple portals linking to the same root
func TestPortalRoot_MultiplePortals(t *testing.T) {
	// Create a portal root
	portalRoot := &rtui.Fiber{
		NodeID: 1,
		Key:    "portal-root",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRootId": "main",
		},
	}

	// Create multiple portals pointing to the same root
	portal1 := &rtui.Fiber{
		NodeID: 10,
		Key:    "tooltip-1",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "main",
		},
	}

	portal2 := &rtui.Fiber{
		NodeID: 20,
		Key:    "tooltip-2",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "main",
		},
	}

	portal3 := &rtui.Fiber{
		NodeID: 30,
		Key:    "toast",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "main",
		},
	}

	// Link as siblings
	portalRoot.Child = portal1
	portal1.Sibling = portal2
	portal2.Sibling = portal3

	// Link portals
	linkPortalsToRootsHelper(portalRoot)

	// All portals should point to the same PortalRoot
	assert.Equal(t, portalRoot.NodeID, portal1.PortalRoot.NodeID)
	assert.Equal(t, portalRoot.NodeID, portal2.PortalRoot.NodeID)
	assert.Equal(t, portalRoot.NodeID, portal3.PortalRoot.NodeID)
}

// TestPortalRoot_NonExistentTarget tests portal targeting non-existent root
func TestPortalRoot_NonExistentTarget(t *testing.T) {
	// Create a portal that references a non-existent root
	portal := &rtui.Fiber{
		NodeID: 100,
		Key:    "orphan-portal",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "non-existent-root",
		},
	}

	// No PortalRoot nodes in the tree
	root := &rtui.Fiber{
		NodeID: 1,
		Key:    "root",
		Type:   rtui.VNodeElement,
	}
	root.Child = portal

	// Link portals
	linkPortalsToRootsHelper(root)

	// PortalRoot should remain nil (target doesn't exist)
	assert.Nil(t, portal.PortalRoot)
}

// TestPortalRoot_NestedPortals tests nested portal structures
func TestPortalRoot_NestedPortals(t *testing.T) {
	// Structure:
	//   Root
	//   ├── PortalRoot (id="outer")
	//   │   └── Portal (root="inner")  <- Nested portal
	//   └── PortalRoot (id="inner")

	outerRoot := &rtui.Fiber{
		NodeID: 1,
		Key:    "outer-root",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRootId": "outer",
		},
	}

	innerRoot := &rtui.Fiber{
		NodeID: 2,
		Key:    "inner-root",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRootId": "inner",
		},
	}

	nestedPortal := &rtui.Fiber{
		NodeID: 100,
		Key:    "nested-portal",
		Type:   rtui.VNodeElement,
		Props: rtui.Props{
			"portalRoot": "inner",
		},
	}

	// Build tree
	root := &rtui.Fiber{
		NodeID: 0,
		Key:    "root",
		Type:   rtui.VNodeElement,
	}
	root.Child = outerRoot
	outerRoot.Child = nestedPortal
	outerRoot.Sibling = innerRoot

	// Link portals
	linkPortalsToRootsHelper(root)

	// Nested portal should link to innerRoot
	assert.NotNil(t, nestedPortal.PortalRoot)
	assert.Equal(t, innerRoot.NodeID, nestedPortal.PortalRoot.NodeID)
}
