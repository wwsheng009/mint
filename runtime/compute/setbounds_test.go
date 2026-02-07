package compute

import (
	"os"
	"testing"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestSetBoundsCalledForButtons verifies that SetBounds is called on buttons
// before Paint() would be invoked
func TestSetBoundsCalledForButtons(t *testing.T) {
	// Create a simple Wrap with FillWidth
	button1 := app.ButtonBuilder("Button1").Build()
	button2 := app.ButtonBuilder("Button2").Build()
	button3 := app.ButtonBuilder("Button3").Build()
	button4 := app.ButtonBuilder("Button4").Build()

	wrap := rtui.WrapBuilder(button1, button2, button3, button4).
		Gap(1).
		ScreenWidth(80).
		Build()

	// Create layout engine
	engine := NewEngine()

	// Set up constraints
	constraints := runtime.NewBoxConstraints(0, 80, 0, 24)

	// Enable debug output
	os.Setenv("TUI_LAYOUT_DEBUG", "true")
	defer os.Unsetenv("TUI_LAYOUT_DEBUG")

	t.Log("=== Starting Layout Phase ===")

	// Phase 1: Layout
	layout, err := engine.Layout(wrap, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	t.Log("=== Layout Complete ===")

	// Find all button boxes
	var buttonBoxes []*ComputedBox
	findButtons(layout.Root, &buttonBoxes)

	if len(buttonBoxes) == 0 {
		t.Fatal("No button boxes found in layout")
	}

	t.Logf("Found %d button boxes", len(buttonBoxes))

	// Verify bounds were set
	for i, box := range buttonBoxes {
		t.Logf("Button %d: ComputedBox bounds=(%d,%d,%dx%d)",
			i, box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height)

		if box.Box.Width == 0 {
			t.Errorf("Button %d has zero width - SetBounds may not have been called correctly", i)
		}

		// Check if VNode's internal bounds field is set
		if btn, ok := box.VNode.(interface{ Bounds() [4]int }); ok {
			vnodeBounds := btn.Bounds()
			t.Logf("Button %d: VNode internal bounds = %v", i, vnodeBounds)

			if vnodeBounds[2] == 0 {
				t.Errorf("Button %d: VNode internal bounds width is 0 - SetBounds was not called or not effective", i)
			}

			if vnodeBounds[2] != box.Box.Width {
				t.Errorf("Button %d: VNode bounds width (%d) != ComputedBox width (%d) - SetBounds not synced",
					i, vnodeBounds[2], box.Box.Width)
			}
		} else {
			t.Errorf("Button %d: VNode does not have Bounds() method", i)
		}
	}
}

// findButtons recursively finds all button ComputedBoxes
func findButtons(box *ComputedBox, buttons *[]*ComputedBox) {
	if box == nil {
		return
	}

	// Check if this box is a button
	if box.VNode != nil {
		if tagger, ok := box.VNode.(interface{ Tag() string }); ok {
			if tagger.Tag() == "button" {
				*buttons = append(*buttons, box)
				return // Don't recurse into buttons
			}
		}
	}

	// Recurse into children
	for _, child := range box.Children {
		findButtons(child, buttons)
	}
}
