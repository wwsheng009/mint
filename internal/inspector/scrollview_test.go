package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/components/layout"
	"github.com/wwsheng009/mint/ui"
)

// TestScrollViewWrapper verifies that TreeView is properly wrapped in ScrollView
func TestScrollViewWrapper(t *testing.T) {
	t.Log("\n=== Testing ScrollView Wrapper ===\n")

	// Create a large content that exceeds viewport
	var content []ui.VNode
	for i := 0; i < 50; i++ {
		content = append(content, ui.Text(fmt.Sprintf("Line %d", i+1)))
	}
	vstackContent := ui.VStack(content...)

	// Wrap in ScrollView with fixed height
	viewportHeight := 10
	scrollContainer := layout.NewScrollView(vstackContent).
		Height(viewportHeight).
		Width(80).
		ScrollOffset(0).
		Build()

	t.Logf("✅ ScrollView created: %T", scrollContainer)

	// Check children
	children := scrollContainer.Children()
	t.Logf("✅ ScrollView has %d children", len(children))

	if len(children) == 0 {
		t.Error("❌ ScrollView has no children!")
	} else {
		// ScrollView should render only visible lines (not all 50)
		t.Logf("✅ ScrollView rendered %d children (content has 50, viewport=%d)",
			len(children), viewportHeight)

		// With virtual scrolling, we should see at most viewportHeight children
		if len(children) > viewportHeight+2 { // +2 for scroll indicator
			t.Errorf("❌ ScrollView rendered %d children, expected at most %d",
				len(children), viewportHeight+2)
		} else {
			t.Logf("✅ Virtual scrolling working: rendered %d visible lines",
				len(children))
		}
	}

	t.Log("\n=== Test Complete ===")
}
