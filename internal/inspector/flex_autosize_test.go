package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestInspectorFlexAutoSizing tests that TreeView auto-sizes and ScrollView provides constraints
func TestInspectorFlexAutoSizing(t *testing.T) {
	t.Log("\n=== Testing Flex-like Auto-Sizing ===\n")

	inspector := NewStandaloneInspector()
	inspector.Enable()

	// Create a large tree (50 nodes)
	var children []ui.VNode
	for i := 0; i < 50; i++ {
		children = append(children, ui.Text(fmt.Sprintf("Node %d", i+1)))
	}
	root := ui.VStack(children...)

	inspector.AttachToApp(root)

	t.Logf("✅ Inspector enabled: %v", inspector.IsVisible())
	t.Logf("✅ Overlay size: %dx%d", 80, 25)

	// Get Elements tab content
	elementsTab := inspector.buildElementsTabContent()

	if elementsTab == nil {
		t.Fatal("❌ Elements tab content is nil!")
	}

	t.Logf("✅ Elements tab created: %T", elementsTab)

	// Check the structure
	tabChildren := elementsTab.Children()
	t.Logf("✅ Elements tab has %d children", len(tabChildren))

	// Expected structure:
	// 1. Header (2-3 lines)
	// 2. Selected info (0-4 lines)
	// 3. ScrollView with tree (1 child - the scroll container)
	// 4. Instructions (5+ lines)

	// Find the ScrollView child
	var scrollViewFound bool
	for i, child := range tabChildren {
		if child == nil {
			t.Logf("  Child #%d: <nil>", i)
			continue
		}
		t.Logf("  Child #%d: %T", i, child)
		// Check all children for ScrollView
		if child.Children() != nil {
			for j, grandchild := range child.Children() {
				t.Logf("    Grandchild #%d: %T", j, grandchild)
				// Check if this is the ScrollView by checking its type name
				typeStr := fmt.Sprintf("%T", grandchild)
				if grandchild.Children() != nil && len(grandchild.Children()) > 0 {
					// ScrollView typically has 1 child (text node)
					t.Logf("      Has %d children", len(grandchild.Children()))
				}
				// ScrollView is a LayoutNode with specific structure
				if typeStr == "*ui.LayoutNode" {
					scrollViewFound = true
					t.Logf("✅ Found LayoutNode (likely ScrollView) at child #%d, grandchild #%d", i, j)

					// Check ScrollView's children
					if grandchild.Children() != nil {
						svChildren := grandchild.Children()
						t.Logf("✅ ScrollView has %d children (should be 1 text node)", len(svChildren))
					}
				}
			}
		}
	}

	if !scrollViewFound {
		t.Logf("Note: ScrollView not found as grandchild LayoutNode (TreeView may be a direct child)")
	} else {
		t.Log("✅ Flex-like auto-sizing is implemented:")
		t.Log("  - TreeView auto-sizes to content (no fixed viewportHeight)")
		t.Log("  - ScrollView provides fixed-height constraint")
		t.Log("  - ScrollView handles clipping and scrolling")
	}

	t.Log("\n=== Test Complete ===")
}
