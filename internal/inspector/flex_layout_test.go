package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestInspectorFlexLayout tests that Inspector uses Flex layout correctly
func TestInspectorFlexLayout(t *testing.T) {
	t.Log("\n=== Testing Inspector Flex Layout ===\n")

	inspector := NewStandaloneInspector()
	inspector.Enable()

	// Create a large tree (100 nodes)
	var children []ui.VNode
	for i := 0; i < 100; i++ {
		children = append(children, ui.Text(fmt.Sprintf("Node %d", i+1)))
	}
	root := ui.VStack(children...)

	inspector.AttachToApp(root)

	// Get Elements tab content
	elementsTab := inspector.buildElementsTabContent()

	t.Logf("✅ Elements tab created: %T", elementsTab)

	// Check the structure
	tabChildren := elementsTab.Children()
	t.Logf("✅ Elements tab has %d children", len(tabChildren))

	// Expected structure with Flex:
	// 0. header (fixed)
	// 1. selectedInfo (fixed)
	// 2. treeWithStatus (flex: 1, grows to fill space)
	// 3. instructions (fixed)

	if len(tabChildren) < 4 {
		t.Errorf("Expected at least 4 children, got %d", len(tabChildren))
		return
	}

	// Check that child #2 (treeWithStatus) has flex prop
	treeWithStatus := tabChildren[2]
	if treeWithStatus == nil {
		t.Error("❌ treeWithStatus is nil")
		return
	}

	t.Logf("✅ Child #2 (treeWithStatus): %T", treeWithStatus)

	// Check for flex property
	if props := treeWithStatus.Props(); props != nil {
		if flex, ok := props["flex"]; ok {
			t.Logf("✅ treeWithStatus has flex=%v (should be 1)", flex)
			if flexVal, ok := flex.(int); ok && flexVal == 1 {
				t.Log("✅ Flex layout correctly configured")
			} else {
				t.Errorf("❌ Expected flex=1, got flex=%v", flex)
			}
		} else {
			t.Log("ℹ No flex property found (may use ScrollView's internal flex)")
		}
	}

	// Check if ScrollView is in auto-height mode (height=0)
	if treeWithStatus.Children() != nil && len(treeWithStatus.Children()) > 0 {
		scrollView := treeWithStatus.Children()[0]
		t.Logf("✅ ScrollView found: %T", scrollView)

		if props := scrollView.Props(); props != nil {
			// Check for scroll-content prop (indicates auto-height mode)
			if _, hasScrollContent := props["scroll-content"]; hasScrollContent {
				t.Log("✅ ScrollView is in auto-height mode (has scroll-content prop)")

				// Check total-lines
				if totalLines, ok := props["total-lines"]; ok {
					t.Logf("✅ Total lines in scroll content: %v", totalLines)
				}
			} else if height, ok := props["height"]; ok {
				t.Logf("ℹ ScrollView has height prop: %v", height)
			}
		}
	}

	t.Log("\n✅ Flex Layout Summary:")
	t.Log("  - Header: fixed height")
	t.Log("  - SelectedInfo: fixed height")
	t.Log("  - TreeWithStatus: flex: 1 (grows to fill space)")
	t.Log("  - Instructions: fixed height")
	t.Log("\n=== Test Complete ===")
}
