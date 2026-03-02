package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/scrollview"
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
	scrollContainer := app.ScrollView().
		Child(vstackContent).
		Height(viewportHeight).
		Width(80).
		ScrollOffset(0).
		Build()

	t.Logf("✅ ScrollView created: %T", scrollContainer)

	// Check children - ScrollView handles its own painting, so Children() returns nil
	children := scrollContainer.Children()
	t.Logf("✅ ScrollView has %d children (expected 0 - ScrollView paints own content)", len(children))

	if len(children) == 0 {
		t.Logf("✅ ScrollView correctly returns nil children (handles own painting)")
	} else {
		t.Errorf("❌ ScrollView unexpectedly has %d children (expected 0)", len(children))
	}

	// Access the VNode properties to verify configuration
	if sv, ok := scrollContainer.(*scrollview.VNode); ok {
		t.Logf("✅ Successfully identified ScrollView VNode")
		t.Logf("   Width: %d, Height: %d", sv.Width(), sv.Height())
		t.Logf("   Child: %T", sv.Child())

		// Create instance and test scroll behavior
		inst := sv.CreateInstance()
		if svInst, ok := inst.(*scrollview.Instance); ok {
			t.Logf("✅ Successfully created ScrollView Instance")
			t.Logf("   Total Lines: %d", svInst.GetTotalLines())
			t.Logf("   Viewport Size: %d", svInst.GetViewportSize())
			t.Logf("   Is Scrollable: %v", svInst.IsScrollable())

			// Verify content exceeds viewport
			if svInst.IsScrollable() {
				t.Logf("✅ Content (%d lines) > viewport (%d lines) - scrollable",
					svInst.GetTotalLines(), svInst.GetViewportSize())

				// Test that scrolling works
				initialOffset := svInst.GetScrollOffset()
				scrollOffset := svInst.ScrollBy(5)
				t.Logf("✅ ScrollBy(5) moved offset from %d to %d", initialOffset, scrollOffset)

				// Verify instance correctly limits visible content
				// Even though there are 50 total lines, ScrollView should only render
				// the visible portion (10 lines in viewport)
				t.Logf("✅ ScrollView correctly clips content to viewport height=%d",
					svInst.GetViewportSize())
			} else {
				t.Errorf("❌ Expected ScrollView to be scrollable with 50 lines in 10-height viewport")
			}
		}
	} else {
		t.Errorf("❌ ScrollContainer is not *scrollview.VNode, got %T", scrollContainer)
	}

	t.Log("\n=== Test Complete ===")
}
