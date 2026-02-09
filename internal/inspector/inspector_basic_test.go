package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestInspectorBasicRendering tests basic Inspector rendering
func TestInspectorBasicRendering(t *testing.T) {
	t.Log("\n=== Testing Basic Inspector Rendering ===\n")

	inspector := NewStandaloneInspector()
	inspector.Enable()

	// Attach a larger tree to test virtual scrolling
	var children []ui.VNode
	for i := 0; i < 30; i++ {
		children = append(children, ui.Text(fmt.Sprintf("Node %d", i+1)))
	}
	root := ui.VStack(children...)

	inspector.AttachToApp(root)

	t.Logf("Inspector enabled: %v", inspector.IsVisible())
	t.Logf("Overlay size: %dx%d", 80, 25) // Default values

	// Get Elements tab content
	elementsTab := inspector.buildElementsTabContent()

	if elementsTab == nil {
		t.Error("❌ Elements tab content is nil!")
	} else {
		t.Logf("✅ Elements tab content created: %T", elementsTab)

		// Check if it has children
		children := elementsTab.Children()
		t.Logf("✅ Elements tab has %d children", len(children))

		// Check the tree
		if inspector.treeViewComponent == nil {
			t.Error("❌ TreeViewComponent is nil!")
		} else {
			t.Logf("✅ TreeViewComponent exists: %T", inspector.treeViewComponent)

			// Check if it has lines
			lines := inspector.treeViewComponent.GetLines()
			t.Logf("✅ TreeViewComponent has %d lines", len(lines))

			// Get render
			render := inspector.treeViewComponent.GetRender()
			if render == nil {
				t.Error("❌ GetRender() returned nil!")
			} else {
				t.Logf("✅ GetRender() returned: %T", render)

				renderChildren := render.Children()
				if len(renderChildren) == 0 {
					t.Error("❌ Render has no children!")
				} else {
					t.Logf("✅ Render has %d children", len(renderChildren))

					// Check first child
					firstChild := renderChildren[0]
					t.Logf("✅ First child type: %T", firstChild)

					// Check grandchildren
					if firstChild.Children() != nil {
						textChildren := firstChild.Children()
						t.Logf("✅ First child has %d text children (tree lines)", len(textChildren))

						if len(textChildren) == 0 {
							t.Error("❌ NO TEXT CHILDREN - TreeView is not rendering any lines!")
						} else {
							t.Logf("✅ SUCCESS: TreeView is rendering %d lines", len(textChildren))
						}
					}
				}
			}
		}
	}

	t.Log("\n=== Test Complete ===")
}
