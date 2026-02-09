package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/components/layout"
	"github.com/wwsheng009/mint/ui"
)

// TestScrollViewHeightClipping tests if ScrollView actually clips content to height
func TestScrollViewHeightClipping(t *testing.T) {
	t.Log("\n=== Testing ScrollView Height Clipping ===\n")

	// Create content with 50 lines
	var content []ui.VNode
	for i := 0; i < 50; i++ {
		content = append(content, ui.Text(fmt.Sprintf("Line %d", i+1)))
	}
	vstackContent := ui.VStack(content...)

	// Wrap in ScrollView with height=10
	viewportHeight := 10
	scrollContainer := layout.NewScrollView(vstackContent).
		Height(viewportHeight).
		Width(80).
		ScrollOffset(0).
		Build()

	t.Logf("ScrollView type: %T", scrollContainer)
	t.Logf("ScrollView children: %d", len(scrollContainer.Children()))

	// Get the rendered text content
	if scrollContainer.Children() != nil && len(scrollContainer.Children()) > 0 {
		firstChild := scrollContainer.Children()[0]
		t.Logf("First child type: %T", firstChild)

		// Try to extract content
		if props := firstChild.Props(); props != nil {
			if content, ok := props["content"]; ok {
				if contentStr, ok := content.(string); ok {
					lines := countLines(contentStr)
					t.Logf("✅ Rendered content has %d lines", lines)

					// Should have at most viewportHeight lines
					if lines <= viewportHeight+1 { // +1 for scroll indicator
						t.Logf("✅ ScrollView correctly clips to %d lines (height=%d)", lines, viewportHeight)
					} else {
						t.Errorf("❌ ScrollView rendered %d lines, expected at most %d", lines, viewportHeight+1)
					}
				}
			}
		}
	}

	t.Log("\n=== Test Complete ===")
}

func countLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	if len(s) > 0 {
		count++ // Last line doesn't have trailing \n
	}
	return count
}
