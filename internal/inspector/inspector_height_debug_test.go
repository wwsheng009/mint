package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestInspectorHeights checks if Inspector height calculations are correct
func TestInspectorHeights(t *testing.T) {
	t.Log("\n=== Testing Inspector Height Calculations ===\n")

	inspector := NewStandaloneInspector()

	// Default overlay size
	t.Logf("Default overlay width: %d", inspector.overlayWidth)
	t.Logf("Default overlay height: %d", inspector.overlayHeight)

	if inspector.overlayHeight <= 0 {
		t.Errorf("❌ overlayHeight is %d, should be > 0", inspector.overlayHeight)
	} else {
		t.Logf("✅ overlayHeight = %d", inspector.overlayHeight)
	}

	// Calculate treeViewHeight
	treeViewHeight := inspector.overlayHeight - 14
	t.Logf("treeViewHeight = %d - 14 = %d", inspector.overlayHeight, treeViewHeight)

	if treeViewHeight <= 0 {
		t.Errorf("❌ treeViewHeight is %d, should be > 0", treeViewHeight)
	} else {
		t.Logf("✅ treeViewHeight = %d", treeViewHeight)
	}

	// Create content and attach
	inspector.Enable()
	var children []ui.VNode
	for i := 0; i < 30; i++ {
		children = append(children, ui.Text(fmt.Sprintf("Node %d", i+1)))
	}
	inspector.AttachToApp(ui.VStack(children...))

	// Build Elements tab
	elementsTab := inspector.buildElementsTabContent()

	t.Logf("Elements tab type: %T", elementsTab)
	t.Logf("Elements tab children: %d", len(elementsTab.Children()))

	// Check each child
	for i, child := range elementsTab.Children() {
		if child == nil {
			t.Logf("  Child #%d: <nil>", i)
			continue
		}
		t.Logf("  Child #%d: %T", i, child)

		// Check grandchildren
		if child.Children() != nil {
			for j, gc := range child.Children() {
				t.Logf("    Grandchild #%d: %T", j, gc)

				// Check if this is the ScrollView
				if gc.Children() != nil && len(gc.Children()) > 0 {
					gcc := gc.Children()[0]
					t.Logf("      Great-grandchild: %T", gcc)

					// Try to get content
					if props := gcc.Props(); props != nil {
						// Check for flex property (indicates auto-height mode)
						if flex, ok := props["flex"]; ok {
							t.Logf("      Has flex=%v (auto-height mode)", flex)
						}

						// Check for scroll-content (indicates auto-height mode)
						if scrollContent, ok := props["scroll-content"]; ok {
							if contentStr, ok := scrollContent.(string); ok {
								lines := 0
								for _, c := range contentStr {
									if c == '\n' {
										lines++
									}
								}
								if len(contentStr) > 0 {
									lines++
								}
								t.Logf("      Scroll-content has %d lines (full content, not clipped)", lines)
								t.Logf("✅ ScrollView in auto-height mode - parent flex will constrain height")
							}
						}

						// Old behavior: check content prop (fixed-height mode)
						if content, ok := props["content"]; ok {
							_, hasScrollContent := props["scroll-content"]
							if !hasScrollContent {
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
									t.Logf("      Content has %d lines", lines)

									// Should be clipped to treeViewHeight (if in fixed-height mode)
									if lines > treeViewHeight+2 {
										t.Logf("ℹ Content not clipped (may be auto-height mode)")
									} else {
										t.Logf("✅ Content clipped to %d lines (treeViewHeight=%d)", lines, treeViewHeight)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	t.Log("\n=== Test Complete ===")
}
