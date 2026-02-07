package render

import (
	"testing"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestSetBoundsCalledForButtons verifies that SetBounds is called on buttons
// before Paint() is invoked
func TestSetBoundsCalledForButtons(t *testing.T) {
	// Create a simple HStack with buttons
	button1 := app.ButtonBuilder("Button1").Build()
	button2 := app.ButtonBuilder("Button2").Build()

	hstack := rtui.HStackBuilder(button1, button2).
		Gap(1).
		Build()

	// Create rendering pipeline
	pipeline := NewRenderingPipeline()

	// Set up constraints
	constraints := runtime.NewBoxConstraints(0, 80, 0, 24)

	// Create layout
	layout, err := pipeline.GetLayoutEngine().Layout(hstack, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	// Phase 1: Layout complete - verify SetBounds was called
	t.Log("=== Phase 1: Layout Complete ===")

	// Find all button boxes
	var buttonBoxes []*runtime.ComputedBox
	findButtons(layout.Root, &buttonBoxes)

	if len(buttonBoxes) == 0 {
		t.Fatal("No button boxes found in layout")
	}

	t.Logf("Found %d button boxes", len(buttonBoxes))

	// Verify bounds were set
	for i, box := range buttonBoxes {
		t.Logf("Button %d: bounds=(%d,%d,%dx%d)",
			i, box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height)

		if box.Box.Width == 0 {
			t.Errorf("Button %d has zero width - SetBounds may not have been called", i)
		}

		// Check if VNode's internal bounds field is set
		if btn, ok := box.VNode.(interface{ Bounds() [4]int }); ok {
			vnodeBounds := btn.Bounds()
			t.Logf("Button %d: VNode bounds = %v", i, vnodeBounds)

			if vnodeBounds[2] == 0 {
				t.Errorf("Button %d: VNode internal bounds width is 0 - SetBounds was not called or not effective", i)
			}
		} else {
			t.Errorf("Button %d: VNode does not have Bounds() method", i)
		}
	}
}

// findButtons recursively finds all button ComputedBoxes
func findButtons(box *runtime.ComputedBox, buttons *[]*runtime.ComputedBox) {
	if box == nil {
		return
	}

	// Check if this box is a button
	if box.VNode != nil {
		if tagger, ok := box.VNode.(interface{ Tag() string }); ok {
			if tagger.Tag() == "button" {
				*buttons = append(*buttons, box)
				return // Don't recurse into buttons - they handle their own children
			}
		}
	}

	// Recurse into children
	for _, child := range box.Children {
		findButtons(child, buttons)
	}
}
