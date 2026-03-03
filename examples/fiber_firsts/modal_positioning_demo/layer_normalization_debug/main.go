// Layer Normalization Debug
// Demonstrates how LayerManager normalizes coordinates for each layer
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
	modal "github.com/wwsheng009/mint/ui/components/modal"
	panel "github.com/wwsheng009/mint/ui/components/panel"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DebugApp creates a test UI with multiple layers
func DebugApp() rtui.VNode {
	return rtui.VStack(
		// LayerBase: Header panel
		panel.NewBuilder().
			Title("Header (LayerBase)").
			Content(newtext.New("Background content")).
			Width(60).
			Height(3).
			Build(),

		// Main content with centered and left-aligned modals
		rtui.VStack(
			rtui.Spacer().Flex(1).Build(),
			
			// LayerModal: Centered modal using HStack + Spacers
			rtui.HStack(
				rtui.Spacer().Flex(1).Build(),
				modal.NewBuilder().
					Title("Centered Modal").
					Content(newtext.New("Centered using Stack+Spacer")).
					Width(34).
					Height(10).
					Single().
					Open(true).
					Build(),
				rtui.Spacer().Flex(1).Build(),
			),

			rtui.Spacer().Flex(1).Build(),
		),
	)
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Layer Normalization Debug - Coordinate Analysis          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(DebugApp)
    node.SetApp(fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	// Create buffer
	buf := paint.NewBuffer(80, 30)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 30},
		AvailableWidth:  80,
		AvailableHeight: 30,
	}

	fmt.Println("\nRendering with Layer system...")
	node.Paint(ctx, buf)

	// ====================================
	// Step 1: Analyze Layout Boxes (BEFORE normalization)
	// ====================================
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("STEP 1: Layout Boxes (Computed by Layout Engine)")
	fmt.Println(strings.Repeat("=", 80))
	
	// Access layout result via reflection/internal API - but for this demo, use GetLayoutBoxes
	boxes := node.GetLayoutBoxes()
	if boxes != nil {
		// Group by layer
		byLayer := make(map[int][]*layout.LayoutBox)
		for _, box := range boxes {
			byLayer[int(box.Layer)] = append(byLayer[int(box.Layer)], box)
		}

		// Display each layer
		for layer := 0; layer < 5; layer++ {
			if layerBoxes, ok := byLayer[layer]; ok && len(layerBoxes) > 0 {
				fmt.Printf("\n  Layer %d (%s):\n", layer, getLayerName(rtui.Layer(layer)))
				
				minX, minY, maxX, maxY := layerBoxes[0].X, layerBoxes[0].Y, layerBoxes[0].X + layerBoxes[0].Width, layerBoxes[0].Y + layerBoxes[0].Height
				for _, box := range layerBoxes {
					if box.X < minX { minX = box.X }
					if box.Y < minY { minY = box.Y }
					right := box.X + box.Width
					bottom := box.Y + box.Height
					if right > maxX { maxX = right }
					if bottom > maxY { maxY = bottom }
				}

				fmt.Printf("    Bounding Box: (%d, %d) to (%d, %d), Size %dx%d\n", 
					minX, minY, maxX, maxY, maxX-minX, maxY-minY)

				fmt.Printf("    Boxes:\n")
				for _, box := range layerBoxes {
					fmt.Printf("      [%s] Pos:(%d,%d) Size:%dx%d Border:%v\n",
						box.ID, box.X, box.Y, box.Width, box.Height,
						box.BoxModel.Border.Style != layout.BorderNone)
				}
			}
		}
	}

	// ====================================
	// Step 2: Analyze PaintableBoxes (AFTER normalization)
	// ====================================
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("STEP 2: Paintable Boxes (After Layer Normalization)")
	fmt.Println(strings.Repeat("=", 80))

	paintableBoxes := node.GetPaintableBoxes()
	if paintableBoxes != nil {
		// Group by layer
		byLayer := make(map[int][]*paint.PaintableBox)
		for _, box := range paintableBoxes {
			byLayer[box.Layer] = append(byLayer[box.Layer], box)
		}

		for layer := 0; layer < 5; layer++ {
			if layerBoxes, ok := byLayer[layer]; ok && len(layerBoxes) > 0 {
				fmt.Printf("\n  Layer %d (%s):\n", layer, getLayerName(rtui.Layer(layer)))
				
				minX, minY, maxX, maxY := layerBoxes[0].X, layerBoxes[0].Y, layerBoxes[0].X + layerBoxes[0].Width, layerBoxes[0].Y + layerBoxes[0].Height
				for _, box := range layerBoxes {
					if box.X < minX { minX = box.X }
					if box.Y < minY { minY = box.Y }
					right := box.X + box.Width
					bottom := box.Y + box.Height
					if right > maxX { maxX = right }
					if bottom > maxY { maxY = bottom }
				}

				isNormalized := minX == 0 && minY == 0
				fmt.Printf("    Bounding Box: (%d, %d) to (%d, %d), Size %dx%d %s\n",
					minX, minY, maxX, maxY, maxX-minX, maxY-minY,
					map[bool]string{true: "✓ Normalized", false: "✗ NOT Normalized"}[isNormalized])

				fmt.Printf("    Boxes:\n")
				for _, box := range layerBoxes {
					id := ""
					if box.Node != nil {
						id = box.Node.ID()
					}
					fmt.Printf("      [%s] Pos:(%d,%d) Size:%dx%d Border:%v\n",
						id, box.X, box.Y, box.Width, box.Height,
						box.BorderStyle != layout.BorderNone)
				}
			}
		}
	}

	// ====================================
	// Step 3: Visual Output
	// ====================================
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("STEP 3: Render Output")
	fmt.Println(strings.Repeat("=", 80))
	printBuffer(buf, 80, 30)

	// ====================================
	// Summary
	// ====================================
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SUMMARY: Coordinate Transformation")
	fmt.Println(strings.Repeat("=", 80))
	
	// Find modal boxes before and after
	fmt.Println("\nModal Position Analysis:")
	if boxes != nil && paintableBoxes != nil {
		// Find modal in layout boxes (Layer 2)
		layoutModal := findBoxByLayer(boxes, 2)
		if layoutModal != nil {
			fmt.Printf("  Layout Phase:    Modal at (%d, %d) (absolute coordinates)\n", 
				layoutModal.X, layoutModal.Y)
		}

		// Find modal in paintable boxes (Layer 2)
		paintableModal := findPaintableBoxByLayer(paintableBoxes, 2)
		if paintableModal != nil {
			fmt.Printf("  Layer Phase:      Modal at (%d, %d) (layer-relative)\n", 
				paintableModal.X, paintableModal.Y)

			// Calculate offset
			if layoutModal != nil {
				offsetX := layoutModal.X - paintableModal.X
				offsetY := layoutModal.Y - paintableModal.Y
				if offsetX != 0 || offsetY != 0 {
					fmt.Printf("  Normalization:  Offset (-%d, -%d) applied to bring layer to (0, 0)\n",
						offsetX, offsetY)
				}
			}
		}
	}

	fmt.Println("\nKey Points:")
	fmt.Println("  1. Layout Engine computes ABSOLUTE coordinates for ALL elements")
	fmt.Println("  2. LayerManager groups boxes by Layer property")
	fmt.Println("  3. LayerManager normalizes: for each layer, subtract min(X, Y) from all boxes")
	fmt.Println("  4. Result: Each layer starts at (0, 0) for independent rendering")
	fmt.Println("  5. Centering is handled by Stack+Spacer in layout, NOT by LayerManager")

	fmt.Println("\n=== Debug Complete ===")
}

func findBoxByLayer(boxes []*layout.LayoutBox, layer int) *layout.LayoutBox {
	for _, box := range boxes {
		if int(box.Layer) == layer && box.BoxModel.Border.Style != layout.BorderNone {
			return box
		}
	}
	return nil
}

func findPaintableBoxByLayer(boxes []*paint.PaintableBox, layer int) *paint.PaintableBox {
	for _, box := range boxes {
		if box.Layer == layer && box.BorderStyle != layout.BorderNone {
			return box
		}
	}
	return nil
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
