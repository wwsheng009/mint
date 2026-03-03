// Test Modal Centered Property (Phase 1.4)
// Tests the Modal's centered property to verify layout phase centering behavior
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	modal "github.com/wwsheng009/mint/ui/components/modal"
	text "github.com/wwsheng009/mint/ui/components/text"
)

// Test case struct
type ModalCenteringTest struct {
	name     string
	node     rtui.VNode
	expected *ExpectedPosition
	desc     string
}

type ExpectedPosition struct {
	// Expected position range (for approximate checking)
	MinX, MaxX int
	MinY,MaxY  int
	// ShouldCenter expected value
	ShouldCenter bool
}

// Test 1: Modal with Centered(true) - should be centered in viewport
func TestCenteredTrue() rtui.VNode {
	// Note: Using Centered(true) directly (Phase 1.4)
	return modal.NewBuilder().
		Key("modal-centered-true").
		Title("Centered Modal").
		Content(text.NewBuilder("Modal with Centered(true)").Build()).
		Width(40).
		Height(10).
		Centered(true).
		Open(true).
		Build()
}

// Test 2: Modal with Centered(false) - should follow parent flow
func TestCenteredFalse() rtui.VNode {
	// Note: Using Centered(false) directly (Phase 1.4)
	return modal.NewBuilder().
		Key("modal-centered-false").
		Title("Non-Centered Modal").
		Content(text.NewBuilder("Modal with Centered(false)").Build()).
		Width(40).
		Height(10).
		Centered(false).
		Open(true).
		Build()
}

// Test 3: Modal without any wrapper + Centered(true)
// Should be at (20, 5) in 80x15 viewport (centered)
func TestModalAloneCentered() rtui.VNode {
	return modal.NewBuilder().
		Key("modal-alone-center").
		Title("Alone Centered").
		Content(text.NewBuilder("Modal alone, centered").Build()).
		Width(40).
		Height(10).
		Centered(true).
		Open(true).
		Build()
}

// Test 4: Modal in VStack with Centered(false)
// Should follow VStack flow (not centered)
func TestModalInVStackNotCentered() rtui.VNode {
	return ui.VStack(
		ui.Text("Before modal"),
		modal.NewBuilder().
			Key("modal-vstack-not-center").
			Title("In VStack").
			Content(text.NewBuilder("Modal in VStack, not centered").Build()).
			Width(40).
			Height(10).
			Centered(false).
			Open(true).
			Build(),
		ui.Text("After modal"),
	)
}

// Test 5: Modal in HStack with Centered(false)
// Should follow HStack flow (not centered)
func TestModalInHStackNotCentered() rtui.VNode {
	return ui.HStack(
		modal.NewBuilder().
			Key("modal-hstack-not-center").
			Title("In HStack").
			Content(text.NewBuilder("Modal in HStack, not centered").Build()).
			Width(40).
			Height(10).
			Centered(false).
			Open(true).
			Build(),
		ui.Spacer().Flex(1).Build(),
	)
}

// LayoutInfo captures modal layout information
type LayoutInfo struct {
	X, Y         int
	AbsX, AbsY   int
	Width, Height int
	Layer        layout.Layer
	ShouldCenter bool
}

// getModalLayoutInfo extracts layout info for all modals
func getModalLayoutInfo(boxes []*layout.LayoutBox) []LayoutInfo {
	var results []LayoutInfo
	for _, box := range boxes {
		if box.Layer == layout.LayerModal {
			results = append(results, LayoutInfo{
				X:           box.X,
				Y:           box.Y,
				AbsX:        box.AbsX,
				AbsY:        box.AbsY,
				Width:       box.Width,
				Height:      box.Height,
				Layer:       box.Layer,
				ShouldCenter: box.ShouldCenter,
			})
		}
	}
	return results
}

// verifyCentering checks if modal is properly centered
func verifyCentering(info LayoutInfo, viewportWidth, viewportHeight int, expectedCentered bool) (bool, string) {
	if expectedCentered {
		// Expected to be centered
		expectedX := (viewportWidth - info.Width) / 2
		expectedY := (viewportHeight - info.Height) / 2

		// Allow small tolerance (modal might be off by 1 due to layout engine)
		tolerance := 1
		diffX := abs(info.X - expectedX)
		diffY := abs(info.Y - expectedY)

		if diffX > tolerance || diffY > tolerance {
			return false, fmt.Sprintf("NOT centered: got (%d, %d), expected (%d, %d), diff: X=%d, Y=%d",
				info.X, info.Y, expectedX, expectedY, diffX, diffY)
		}

		if !info.ShouldCenter {
			return false, "ShouldCenter flag is false, but centered position detected"
		}

		return true, fmt.Sprintf("CENTERED at (%d, %d), ShouldCenter=%v ✓", info.X, info.Y, info.ShouldCenter)
	} else {
		// Expected to be NOT centered (follow parent flow)
		if info.ShouldCenter {
			return false, "ShouldCenter flag is true, but expected non-centered layout"
		}

		// Just check position is reasonable (not at viewport center)
		centerX := (viewportWidth - info.Width) / 2
		centerY := (viewportHeight - info.Height) / 2

		// Allow 2 cell tolerance
		tolerance := 2
		diffX := abs(info.X - centerX)
		diffY := abs(info.Y - centerY)

		// If it's close to center, that's unexpected
		if diffX <= tolerance && diffY <= tolerance {
			return false, fmt.Sprintf("Unexpectedly centered near (%d, %d), expected to follow parent flow",
				info.X, info.Y)
		}

		return true, fmt.Sprintf("NOT centered (follows parent flow) at (%d, %d), ShouldCenter=%v ✓",
			info.X, info.Y, info.ShouldCenter)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Phase 1.4: Modal Centered Property Testing                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	const (
		viewportWidth  = 80
		viewportHeight = 25
	)

	tests := []struct {
		name            string
		node            rtui.VNode
		expectedCenter  bool
		desc            string
		checkPosition   bool
	}{
		{
			"Test 1: Centered(true)",
			TestCenteredTrue(),
			true,
			"Modal with Centered(true) - should be centered in viewport",
			true,
		},
		{
			"Test 2: Centered(false)",
			TestCenteredFalse(),
			false,
			"Modal with Centered(false) - should follow parent flow",
			true,
		},
		{
			"Test 3: Modal alone + Centered(true)",
			TestModalAloneCentered(),
			true,
			"Modal without wrapper, Centered(true) - exact center position",
			true,
		},
		{
			"Test 4: Modal in VStack + Centered(false)",
			TestModalInVStackNotCentered(),
			false,
			"Modal in VStack with Centered(false) - should flow with stack",
			false,
		},
		{
			"Test 5: Modal in HStack + Centered(false)",
			TestModalInHStackNotCentered(),
			false,
			"Modal in HStack with Centered(false) - should flow with stack",
			false,
		},
	}

	passed := 0
	failed := 0

	for i, test := range tests {
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Printf("  %s\n", test.name)
		fmt.Printf("  %s\n", test.desc)
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		fwApp := framework.NewApp()
		node := render.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode { return test.node })
    node.SetApp(fwApp)
		node.SetRenderMode(render.RenderModeFiberFirst)

		buf := paint.NewBuffer(viewportWidth, viewportHeight)
		ctx := component.PaintContext{
			Bounds:          paint.Rect{X: 0, Y: 0, Width: viewportWidth, Height: viewportHeight},
			AvailableWidth:  viewportWidth,
			AvailableHeight: viewportHeight,
		}

		// Render
		node.Paint(ctx, buf)

		// Get layout info
		boxes := node.GetLayoutBoxes()
		modals := getModalLayoutInfo(boxes)

		// Test verification
		fmt.Println("\n--- Layout Analysis ---")
		if len(modals) == 0 {
			fmt.Println("  ❌ FAIL: No modals found in layout")
			failed++
			continue
		}

		allPassed := true
		fmt.Printf("  Found %d modal(s)\n", len(modals))

		for j, info := range modals {
			fmt.Printf("\n  Modal #%d:\n", j+1)
			fmt.Printf("    Position: (%d, %d), Abs: (%d, %d)\n", info.X, info.Y, info.AbsX, info.AbsY)
			fmt.Printf("    Size: %dx%d\n", info.Width, info.Height)
			fmt.Printf("    Layer: %d (%s)\n", info.Layer, getLayerName(info.Layer))
			fmt.Printf("    ShouldCenter: %v\n", info.ShouldCenter)

			if test.checkPosition {
				ok, msg := verifyCentering(info, viewportWidth, viewportHeight, test.expectedCenter)
				if ok {
					fmt.Printf("    ✓ %s\n", msg)
				} else {
					fmt.Printf("    ❌ %s\n", msg)
					allPassed = false
				}
			}
		}

		// 打印所有 LayoutBoxes 用于调试
		if len(modals) > 0 {
			fmt.Println("\n--- All Layout Boxes (Debug) ---")
			for _, box := range boxes {
				layerName := getLayerName(box.Layer)
				fmt.Printf("  [%s] ID=%s, Pos=(%d,%d), Abs=(%d,%d), Size=%dx%d, ShouldCenter=%v\n",
					layerName, box.ID, box.X, box.Y, box.AbsX, box.AbsY,
					box.Width, box.Height, box.ShouldCenter)
			}
		}

		if len(modals) > 0 {
			firstModal := modals[0]
			if !test.checkPosition {
				// Just checking ShouldCenter flag
				if firstModal.ShouldCenter == test.expectedCenter {
					fmt.Printf("\n  ✓ ShouldCenter flag correct: %v\n", firstModal.ShouldCenter)
					passed++
				} else {
					fmt.Printf("\n  ❌ ShouldCenter flag wrong: expected %v, got %v\n",
						test.expectedCenter, firstModal.ShouldCenter)
					failed++
				}
			}
		}

		// Show render output (just top portion)
		if i < 2 { // Only show first two tests' render
			fmt.Println("\n--- Render Output (Top 20 rows) ---")
			printBuffer(buf, viewportWidth, 20)
		}

		if test.checkPosition {
			if allPassed {
				passed++
			} else {
				failed++
			}
		}
	}

	// Summary
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("  SUMMARY")
	fmt.Printf("%s\n", strings.Repeat("=", 80))
	fmt.Printf("  Total Tests: 5\n")
	fmt.Printf("  Passed: %d\n", passed)
	fmt.Printf("  Failed: %d\n", failed)

	if failed == 0 {
		fmt.Println("\n  ✅ All tests passed! Modal centered property works correctly.")
		fmt.Println("     - Centered(true): Modal is centered in viewport")
		fmt.Println("     - Centered(false): Modal follows parent layout flow")
	} else {
		fmt.Printf("\n  ❌ %d test(s) failed\n", failed)
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
}

func getLayerName(l layout.Layer) string {
	switch l {
	case layout.LayerBase:
		return "LayerBase"
	case layout.LayerOverlay:
		return "LayerOverlay"
	case layout.LayerModal:
		return "LayerModal"
	case layout.LayerTooltip:
		return "LayerTooltip"
	case layout.LayerInspector:
		return "LayerInspector"
	default:
		return fmt.Sprintf("Unknown(%d)", int(l))
	}
}

func printBuffer(buf *paint.Buffer, w, h int) {
	fmt.Println("┌" + strings.Repeat("─", w) + "┐")
	for y := 0; y < h; y++ {
		line := "|"
		for x := 0; x < w; x++ {
			if y < len(buf.Cells) && x < len(buf.Cells[y]) {
				cell := buf.Cells[y][x]
				if len(cell.Cluster) == 0 || cell.Cluster == " " {
					line += " "
				} else {
					for _, r := range cell.Cluster {
						line += string(r)
						break
					}
				}
			} else {
				line += " "
			}
		}
		line += "|"
		fmt.Println(line)
	}
	fmt.Println("└" + strings.Repeat("─", w) + "┘")
}
