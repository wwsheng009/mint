package main

import (
	"fmt"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/internal/inspector"
)

func main() {
	fmt.Println("=== Inspector Mouse Click Test ===")
	fmt.Println()

	// Create inspector
	insp := inspector.NewStandaloneInspector()
	insp.Enable()

	// Set overlay size (typical inspector size)
	insp.SetOverlaySize(100, 40)

	// Make inspector visible
	insp.ToggleVisibility()

	// Simulate various mouse clicks
	fmt.Println("Testing mouse click handling...")
	fmt.Println()

	testCases := []struct {
		name string
		x    int
		y    int
		btn  frameworkevent.MouseButton
	}{
		{"Click on tab bar area", 10, 2, frameworkevent.MouseLeft},
		{"Click on TreeView line 5", 10, 20, frameworkevent.MouseLeft},
		{"Click on TreeView line 10", 10, 25, frameworkevent.MouseLeft},
		{"Click outside overlay", 150, 50, frameworkevent.MouseLeft},
	}

	for i, tc := range testCases {
		fmt.Printf("Test %d: %s\n", i+1, tc.name)
		fmt.Printf("  Position: (%d, %d)\n", tc.x, tc.y)

		// Create mouse event
		mouseEvent := &frameworkevent.MouseEvent{
			X:      tc.x,
			Y:      tc.y,
			Button: tc.btn,
		}

		// Check if in overlay bounds
		inOverlay := tc.x >= 0 && tc.x < 100 && tc.y >= 0 && tc.y < 40
		fmt.Printf("  In overlay bounds: %v\n", inOverlay)

		if inOverlay {
			// Try to handle the click
			handled := insp.HandleMouseEvent(frameworkevent.EventMousePress, mouseEvent)
			fmt.Printf("  Handled: %v\n", handled)
		}

		fmt.Println()
	}

	// Show TreeView component info
	tvComponent := insp.GetTreeViewComponent()
	if tvComponent != nil {
		fmt.Println("TreeView Component Info:")
		fmt.Printf("  Line count: %d\n", tvComponent.GetLineCount())
		fmt.Printf("  Focus index: %d\n", tvComponent.GetFocusIndex())
	} else {
		fmt.Println("TreeView Component: nil")
	}

	fmt.Println()
	fmt.Println("=== Test Complete ===")
}
