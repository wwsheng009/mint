// Package main provides modal centering test
package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// TestModalCentering tests that modal content is centered
func TestModalCentering(t *testing.T) {
	// Create a simple modal app
	app := func() ui.VNode {
		return ui.VStack(
			ui.Text("Background content"),
			ui.Modal(
				ui.Bordered().
					Width(40).
					Child(
						ui.VStackBuilder(
							ui.Text(""),
							ui.HStackBuilder(
								ui.Text("*** Title ***"),
							).Align(ui.AlignCenter).Build(),
							ui.Text(""),
							ui.HStackBuilder(
								app.ButtonBuilder("[ Cancel ]").Build(),
								ui.Text(" "),
								app.ButtonBuilder("[ OK ]").Build(),
							).Align(ui.AlignCenter).Build(),
							ui.Text(""),
						).Build(),
					).
					Build(),
			).Build(),
		)
	}

	testApp, err := ui.RunTest(app, ui.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for render
	time.Sleep(100 * time.Millisecond)

	// Get rendered output
	rendered := testApp.GetRenderString()
	t.Logf("Rendered output:\n%s", rendered)

	// Check if modal is present
	if !strings.Contains(rendered, "*** Title ***") {
		t.Error("Modal title not found in rendered output")
	}

	// Check if content is centered
	// The modal is 40 chars wide (38 inner), content should be centered within modal
	lines := strings.Split(rendered, "\n")
	for _, line := range lines {
		if strings.Contains(line, "*** Title ***") {
			// Find modal borders to calculate relative position
			leftBorderPos := strings.Index(line, "│")
			rightBorderPos := strings.LastIndex(line, "│")
			if leftBorderPos == -1 || rightBorderPos == -1 {
				t.Error("Modal borders not found in title line")
				continue
			}

			// Title is 13 chars ("*** Title ***")
			// In a 38-char modal inner width, centered would be at (38-13)/2 = 12-13 chars from left edge
			titlePos := strings.Index(line, "***")
			if titlePos > 0 {
				titlePosInModal := titlePos - leftBorderPos - 1 // Position within modal content
				t.Logf("Title position in modal: %d (from left edge of content)", titlePosInModal)

				// Allow some tolerance
				if titlePosInModal < 10 || titlePosInModal > 15 {
					t.Errorf("Title not centered in modal, pos=%d, expected around 12-13", titlePosInModal)
				} else {
					t.Logf("Title appears centered! pos=%d", titlePosInModal)
				}
			}
		}
	}
}

// TestModalCenteringWithButtons tests button alignment in modal
func TestModalCenteringWithButtons(t *testing.T) {
	// Create a modal with buttons
	app := func() ui.VNode {
		return ui.VStack(
			ui.Text("Background"),
			ui.Modal(
				ui.Bordered().
					Width(40).
					Child(
						ui.VStackBuilder(
							ui.Text(""),
							ui.HStackBuilder(
								app.ButtonBuilder("[ Cancel ]").Build(),
								ui.Text(" "),
								app.ButtonBuilder("[ OK ]").Build(),
							).Align(ui.AlignCenter).Build(),
							ui.Text(""),
						).Build(),
					).
					Build(),
			).Build(),
		)
	}

	testApp, err := ui.RunTest(app, ui.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("Rendered output:\n%s", rendered)

	// Find button line
	lines := strings.Split(rendered, "\n")
	for _, line := range lines {
		if strings.Contains(line, "[ Cancel ]") && strings.Contains(line, "[ OK ]") {
			// Calculate positions
			cancelPos := strings.Index(line, "[ Cancel ]")
			okPos := strings.Index(line, "[ OK ]")
			t.Logf("Button positions: Cancel=%d, OK=%d", cancelPos, okPos)

			// Find modal borders to calculate relative position
			leftBorderPos := strings.Index(line, "│")
			rightBorderPos := strings.LastIndex(line, "│")
			if leftBorderPos == -1 || rightBorderPos == -1 {
				t.Error("Modal borders not found")
				return
			}
			modalInnerWidth := rightBorderPos - leftBorderPos - 1 // 38 for Width(40)
			cancelPosInModal := cancelPos - leftBorderPos - 1 // Position within modal content
			t.Logf("Modal: left border at %d, inner width %d, Cancel at relative pos %d",
				leftBorderPos, modalInnerWidth, cancelPosInModal)

			// Buttons should be centered within the modal content area
			// With actual widths (including focus indicators): ~15 + 1 + 11 = 27-29 chars
			// Centered would be around (modalInnerWidth - 29) / 2 from left edge
			// Allow wide tolerance for rendering variations
			expectedPos := (modalInnerWidth - 27) / 2  // rough estimate
			t.Logf("Expected centering position around %d", expectedPos)
			if cancelPosInModal < 3 || cancelPosInModal > 12 {
				t.Errorf("Buttons not centered in modal, relative pos=%d, expected around %d", cancelPosInModal, expectedPos)
			} else {
				t.Logf("Buttons appear centered in modal! relative pos=%d", cancelPosInModal)
			}
		}
	}
}

// TestModalAlignProp tests that align prop is set correctly
func TestModalAlignProp(t *testing.T) {
	hstack := ui.HStackBuilder(
		ui.Text("Test"),
	).Align(ui.AlignCenter).Build()

	props := hstack.Props()
	if props == nil {
		t.Fatal("Props is nil")
	}

	align, ok := props["align"].(int)
	if !ok {
		t.Error("align prop not found or not an int")
	} else if align != int(ui.AlignCenter) {
		t.Errorf("align prop = %d, expected %d (AlignCenter)", align, ui.AlignCenter)
	} else {
		t.Logf("align prop correctly set to %d", align)
	}
}
