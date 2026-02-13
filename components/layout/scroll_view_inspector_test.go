package layout

import (
	"fmt"
	"os"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
)

// TestScrollViewWithLayoutNode tests if ScrollView can extract content from LayoutNode
func TestScrollViewWithLayoutNode(t *testing.T) {
	os.Setenv("TUI_DEBUG_INSPECTOR", "true")

	t.Log("\n=== Testing ScrollView with LayoutNode ===\n")

	// Create a LayoutNode with text children (simulating TreeView output)
	var textChildren []ui.VNode
	for i := 0; i < 5; i++ {
		textChildren = append(textChildren, ui.Text(fmt.Sprintf("Line %d", i+1)))
	}

	layoutNode := rtui.VStack(textChildren...)

	t.Logf("Created LayoutNode: %T", layoutNode)
	t.Logf("LayoutNode has %d children", len(layoutNode.Children()))

	// Create ScrollView with this content
	sv := NewScrollView(layoutNode).
		Width(80).
		// Height NOT set - auto-height mode
		Build()

	t.Logf("ScrollView created: %T", sv)

	// Check what was rendered
	if sv.Children() != nil && len(sv.Children()) > 0 {
		firstChild := sv.Children()[0]
		t.Logf("First child: %T", firstChild)

		// Try to get content
		if props := firstChild.Props(); props != nil {
			// VStackBuilder wraps textNode and puts content in props["content"]
			if content, ok := props["content"]; ok {
				if contentStr, ok := content.(string); ok {
					lines := 0
					for _, c := range contentStr {
						if c == '\n' {
							lines++
						}
					}
					if len(contentStr) > 0 {
						lines++
					}
					t.Logf("✅ content has %d lines", lines)
					t.Logf("Content preview: %q", contentStr)

					if lines == 5 {
						t.Log("✅ ScrollView correctly extracted content from LayoutNode")
					} else {
						t.Errorf("❌ Expected 5 lines, got %d", lines)
					}
				} else {
					t.Error("❌ content is not a string")
				}
			} else {
				t.Error("❌ No content prop found")
				t.Logf("Available props: %v", props)
			}
		} else {
			t.Error("❌ First child has no props")
		}
	} else {
		t.Error("❌ ScrollView has no children")
	}

	t.Log("\n=== Test Complete ===")
}
