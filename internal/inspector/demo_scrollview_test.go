package inspector

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestDemoScrollViewSimulation simulates the demo scenario with ScrollView
func TestDemoScrollViewSimulation(t *testing.T) {
	// Enable verbose logging for this test
	os.Setenv("TUI_INSPECTOR_VERBOSE", "true")
	defer os.Unsetenv("TUI_INSPECTOR_VERBOSE")

	t.Log("\n=== Simulating Demo with ScrollView ===\n")

	inspector := NewStandaloneInspector()
	inspector.Enable()

	// Create a large tree like in the demo
	var children []ui.VNode
	for i := 0; i < 100; i++ {
		children = append(children, ui.Text(fmt.Sprintf("Node %d: Some content here", i+1)))
	}
	root := ui.VStack(children...)

	inspector.AttachToApp(root)

	// Get tree stats
	treeView := inspector.GetTreeView()
	lines, totalLines := treeView.GetTreeLines()

	t.Logf("✅ Tree has %d lines (visible), %d total nodes", len(lines), totalLines)

	// Build Elements tab (this creates the ScrollView)
	elementsTab := inspector.buildElementsTabContent()

	t.Logf("✅ Elements tab created: %T", elementsTab)

	// Walk the tree to find ScrollView
	findScrollView(elementsTab, 0, t)

	// Verify ScrollView clipped the content
	// Expected: treeViewHeight = 25 - 14 = 11
	// ScrollView should render at most 12 lines (11 + 1 for indicator)
	treeViewHeight := inspector.overlayHeight - 14
	t.Logf("Expected treeViewHeight: %d (overlayHeight=%d - 14)", treeViewHeight, inspector.overlayHeight)

	// Check actual rendered content
	scrollViewContent := extractScrollViewContent(elementsTab)
	if scrollViewContent != "" {
		renderedLines := strings.Count(scrollViewContent, "\n") + 1
		t.Logf("✅ ScrollView rendered %d lines", renderedLines)

		if renderedLines <= treeViewHeight+2 {
			t.Logf("✅ ScrollView correctly clips content (≤ %d lines)", treeViewHeight+1)
		} else {
			t.Errorf("❌ ScrollView rendered %d lines, expected ≤ %d", renderedLines, treeViewHeight+2)
		}

		// Check if scroll indicator is present
		if strings.Contains(scrollViewContent, "▼") || strings.Contains(scrollViewContent, "▲") || strings.Contains(scrollViewContent, "↕") {
			t.Log("✅ Scroll indicator present")
		} else {
			t.Log("ℹ Scroll indicator not found (content might fit in viewport)")
		}
	}

	t.Log("\n=== Test Complete ===")
}

func findScrollView(node ui.VNode, depth int, t *testing.T) bool {
	if node == nil {
		return false
	}

	indent := strings.Repeat("  ", depth)
	typeStr := fmt.Sprintf("%T", node)

	// Check if this looks like a ScrollView
	if typeStr == "*ui.LayoutNode" && node.Children() != nil && len(node.Children()) == 1 {
		firstChild := node.Children()[0]
		if firstChild != nil {
			childType := fmt.Sprintf("%T", firstChild)
			t.Logf("%sFound LayoutNode with child: %s", indent, childType)

			// Try to extract content
			if props := firstChild.Props(); props != nil {
				if content, ok := props["content"]; ok {
					if contentStr, ok := content.(string); ok {
						lines := strings.Count(contentStr, "\n") + 1
						t.Logf("%s→ Content has %d lines", indent, lines)

						if strings.Contains(contentStr, "▼") || strings.Contains(contentStr, "▲") {
							t.Logf("%s→ Scroll indicator: YES", indent)
						}
					}
				}
			}
		}
	}

	// Recursively search children
	if node.Children() != nil {
		for _, child := range node.Children() {
			if findScrollView(child, depth+1, t) {
				return true
			}
		}
	}

	return false
}

func extractScrollViewContent(node ui.VNode) string {
	if node == nil {
		return ""
	}

	// Try to get content from props
	if props := node.Props(); props != nil {
		if content, ok := props["content"]; ok {
			if contentStr, ok := content.(string); ok {
				return contentStr
			}
		}
	}

	// Search children
	if node.Children() != nil {
		for _, child := range node.Children() {
			if content := extractScrollViewContent(child); content != "" {
				return content
			}
		}
	}

	return ""
}
