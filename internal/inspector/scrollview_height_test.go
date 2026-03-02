package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/scrollview"
)

// TestScrollViewHeightClipping tests if ScrollView actually clips content to height
func TestScrollViewHeightClipping(t *testing.T) {
	t.Log("\n=== Testing ScrollView Height Clipping ===\n")

	// Create content with 50 lines
	var content []ui.VNode
	for i := 0; i < 50; i++ {
		content = append(content, ui.Text(fmt.Sprintf("Line %d", i+1)))
	}
	vstackContent := app.VStack(content...)

	// Wrap in ScrollView with height=10
	viewportHeight := 10
	scrollContainer := app.ScrollView().
		Child(vstackContent).
		Height(viewportHeight).
		Width(80).
		ScrollOffset(0).
		Build()

	t.Logf("ScrollView type: %T", scrollContainer)
	t.Logf("ScrollView children: %d", len(scrollContainer.Children()))

	// ScrollView handles its own painting, so Children() returns nil
	// But we can access the VNode properties directly
	if sv, ok := scrollContainer.(*scrollview.VNode); ok {
		t.Logf("✅ Successfully identified ScrollView VNode")
		t.Logf("   Width: %d, Height: %d, ScrollOffset: %d",
			sv.Width(), sv.Height(), sv.ScrollOffset())
		t.Logf("   ShowBorder: %v, ShowIndicator: %v",
			sv.ShowBorder(), sv.ShowIndicator())

		// Create an instance to test rendering behavior
		inst := sv.CreateInstance()
		if svInst, ok := inst.(*scrollview.Instance); ok {
			t.Logf("✅ Successfully created ScrollView Instance")
			t.Logf("   Total Lines: %d", svInst.GetTotalLines())
			t.Logf("   Is Scrollable: %v", svInst.IsScrollable())

			// Test that the viewport height is respected
			viewportSize := svInst.GetViewportSize()
			if viewportSize == viewportHeight {
				t.Logf("✅ ScrollView viewport height correctly set to %d", viewportSize)
			} else {
				t.Errorf("❌ ScrollView viewport height is %d, expected %d", viewportSize, viewportHeight)
			}

			// Test that content can be scrolled
			if svInst.IsScrollable() {
				t.Logf("✅ ScrollView is scrollable (content > viewport)")

				// Test scrolling doesn't allow going beyond bounds
				initialOffset := svInst.GetScrollOffset()
				svInst.ScrollBy(100) // Try to scroll way past end
				finalOffset := svInst.GetScrollOffset()
				maxOffset := svInst.GetTotalLines() - viewportHeight

				if finalOffset <= maxOffset {
					t.Logf("✅ ScrollView correctly clamps scroll offset to maximum %d", finalOffset)
				} else {
					t.Errorf("❌ Scroll offset %d exceeds maximum %d", finalOffset, maxOffset)
				}

				// Reset for next test
				svInst.ScrollTo(initialOffset)
				svInst.ScrollBy(-100) // Try to scroll way past top
				finalOffset = svInst.GetScrollOffset()

				if finalOffset >= 0 {
					t.Logf("✅ ScrollView correctly clamps scroll offset to minimum %d", finalOffset)
				} else {
					t.Errorf("❌ Scroll offset %d is below minimum 0", finalOffset)
				}
			}
		}
	} else {
		t.Errorf("❌ ScrollContainer is not *scrollview.VNode, got %T", scrollContainer)
	}

	t.Log("\n=== Test Complete ===")
}
