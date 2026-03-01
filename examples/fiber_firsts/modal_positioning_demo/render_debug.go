// Modal Positioning Render Debug
// Demonstrates modal positioning using Stack+Spacer and Layer system
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
	panel "github.com/wwsheng009/mint/ui/components/panel"
	modal "github.com/wwsheng009/mint/ui/components/modal"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DebugApp shows UI components with modals positioned using Stack+Spacer
func DebugApp() rtui.VNode {
	return rtui.VStack(
		// Header panel
		panel.NewBuilder().
			Title("Header").
			Content(newtext.New("This is the header panel with background content")).
			Width(70).
			Height(3).
			Build(),

		// Main content area
		rtui.VStack(
			// Centered Modal using VStack + Spacer
			// This makes the modal centered in the viewport
			rtui.HStack(
				// Left spacer centers the modal
				rtui.Spacer().Flex(1).Build(),

				// Centered Modal
				modal.NewBuilder().
					Title("Centered Modal").
					Content(newtext.New("This modal is centered using Stack+Spacer")).
					Width(34).
					Height(10).
					Single().
					Open(true).
					Build(),

				// Right spacer centers the modal
				rtui.Spacer().Flex(1).Build(),
			),

			// Vertical spacer for vertical centering
			rtui.Spacer().Flex(1).Build(),

			// Left Aligned Modal using HStack (no left spacer)
			rtui.HStack(
				// Modal left aligned (no spacer on left)
				modal.NewBuilder().
					Title("Left Aligned Modal").
					Content(newtext.New("Left aligned - no spacer before modal")).
					Width(30).
					Height(8).
					Single().
					Open(true).
					Build(),

				// Spacer pushes modal to left
				rtui.Spacer().Flex(1).Build(),
			),

			rtui.Spacer().Flex(1).Build(),

			// Right Aligned Modal using HStack (spacer on left)
			rtui.HStack(
				// Spacer pushes modal to right
				rtui.Spacer().Flex(1).Build(),

				// Modal right aligned
				modal.NewBuilder().
					Title("Right Aligned Modal").
					Content(newtext.New("Right aligned - spacer before modal")).
					Width(30).
					Height(8).
					Single().
					Open(true).
					Build(),
			),

			// Bottom spacer
			rtui.Spacer().Flex(1).Build(),
		),

		// Footer panel
		panel.NewBuilder().
			Title("Footer").
			Content(newtext.New("Footer area")).
			Width(70).
			Height(2).
			Build(),
	)
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
	// os.Setenv("MINT_DEBUG_TEST", "true") // Uncomment for debug output

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Modal Positioning with Stack+Spacer & Layer System       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(DebugApp, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())
	fmt.Printf("  Fiber-First Enabled: %v\n", node.IsFiberFirstEnabled())

	// Create buffer
	buf := paint.NewBuffer(80, 45)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 45},
		AvailableWidth:  80,
		AvailableHeight: 45,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("Rendering with Stack+Spacer positioning and Layer system...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	// Render
	node.Paint(ctx, buf)

	fmt.Println("\n=== Render Output ===")
	printBuffer(buf, 80, 45)

	// Get layout boxes (positions computed by layout engine)
	fmt.Println("\n=== Layout Boxes (Computed by Layout Engine) ===")
	boxes := node.GetLayoutBoxes()
	if boxes != nil {
		modalCount := 0
		panelCount := 0
		for _, box := range boxes {
			layerName := getLayerName(rtui.Layer(box.Layer))
			nodeType := "Unknown"
			if box.Border.Style != layout.BorderNone {
				if box.Layer == 2 { // LayerModal
					nodeType = "Modal"
					modalCount++
				} else {
					nodeType = "Panel"
					panelCount++
				}
			}

			if nodeType == "Modal" {
				fmt.Printf("  [Modal #%d] Node ID: %s\n", modalCount, box.ID)
				fmt.Printf("         Position: X=%d, Y=%d\n", box.X, box.Y)
				fmt.Printf("         Size: Width=%d, Height=%d\n", box.Width, box.Height)
				fmt.Printf("         Layer: %s (%d)\n", layerName, box.Layer)
			}
		}
		fmt.Printf("\n  Summary: %d Panels, %d Modals\n", panelCount, modalCount)
	}

	// Get paintable boxes (after layer normalization)
	fmt.Println("\n=== Paintable Boxes (After Layer Normalization) ===")
	paintableBoxes := node.GetPaintableBoxes()
	if paintableBoxes != nil {
		modalCount := 0
		for _, box := range paintableBoxes {
			if box.BorderStyle != layout.BorderNone && box.Layer == 2 {
				modalCount++
				fmt.Printf("  [Modal #%d] Position: X=%d, Y=%d, Size: %dx%d, Layer=%d\n",
					modalCount, box.X, box.Y, box.Width, box.Height, box.Layer)
			}
		}
	}

	fmt.Println("\n=== Architecture Review ===")
	fmt.Println("  Layout Phase: Stack+Spacer controls modal positioning")
	fmt.Println("    - Centered: HStack(Spacer(1), Modal, Spacer(1))")
	fmt.Println("    - Left Aligned: HStack(Modal, Spacer(1))")
	fmt.Println("    - Right Aligned: HStack(Spacer(1), Modal)")
	fmt.Println("  Layer Phase: LayerManager normalizes all layers to (0, 0)")
	fmt.Println("    - LayerBase (0): Background content")
	fmt.Println("    - LayerModal (2): Modal dialogs (on top)")

	fmt.Println("\n=== Expected Results ===")
	fmt.Println("  ✓ Centered Modal should appear in middle of screen")
	fmt.Println("  ✓ Left Aligned Modal should appear on left side")
	fmt.Println("  ✓ Right Aligned Modal should appear on right side")
	fmt.Println("  ✓ All Modals should appear on top (LayerModal > LayerBase)")

	fmt.Println("\n=== Test Complete ===")
}

func getLayerName(l rtui.Layer) string {
	switch l {
	case rtui.LayerBase:
		return "LayerBase"
	case rtui.LayerOverlay:
		return "LayerOverlay"
	case rtui.LayerModal:
		return "LayerModal"
	case rtui.LayerTooltip:
		return "LayerTooltip"
	case rtui.LayerInspector:
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
