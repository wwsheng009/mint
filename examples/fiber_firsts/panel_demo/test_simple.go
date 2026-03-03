// Package main demonstrates the Fiber-first Panel component.
// Panel is a high-level container that manages borders, headers, and content layout.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/examples/utils"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// DemoApp creates the demo UI
func DemoApp() rtui.VNode {
	return ui.NewVStack().SingleBorder().SetChildrenList([]ui.VNode{ui.Text("Single Border")}).Build()
}

// printLayoutBoxes prints detailed layout box information for debugging
// Uses flattened view for backward compatibility
func printLayoutBoxes(boxes []*layout.LayoutBox) {
	if boxes == nil {
		fmt.Println("  No layout boxes found!")
		return
	}

	// Count boxes by type
	typeCount := make(map[string]int)
	hstackCount := 0
	vstackCount := 0
	panelCount := 0

	// Print ALL boxes with detail (flattened view for backward compatibility)
	fmt.Printf("Total boxes: %d\n\n", len(boxes))
	fmt.Printf("%-10s | %-8s | PropsID      | %-8s | %-10s | %-7s\n", "Node ID", "Tag", "Pos", "Size", "Children")
	fmt.Println(strings.Repeat("-", 90))

	for _, box := range boxes {
		propsID := box.PropsID
		if len(propsID) > 12 {
			propsID = propsID[:9] + "..."
		}
		if propsID == "" {
			propsID = "-"
		}

		fmt.Printf("%-10s | %-8s | %-12s | (%3d,%3d) | %-10s | %d\n",
			box.ID, box.Tag, propsID,
			box.X, box.Y,
			fmt.Sprintf("%dx%d", box.Width, box.Height),
			len(box.Children))

		// Count by type using Tag (new way - more accurate)
		switch box.Tag {
		case "hstack":
			hstackCount++
			typeCount["HStack"]++
		case "vstack":
			vstackCount++
			typeCount["VStack"]++
		case "panel":
			panelCount++
			typeCount["Panel"]++
		case "text":
			typeCount["Text"]++
		default:
			if box.Tag != "" {
				typeCount[box.Tag]++
			} else {
				typeCount["Other"]++
			}
		}
	}

	// Print summary
	fmt.Println("\nType Summary:")
	fmt.Printf("  Panel:   %d\n", panelCount)
	fmt.Printf("  VStack:  %d\n", vstackCount)
	fmt.Printf("  HStack:  %d\n", hstackCount)
	fmt.Printf("  Text:    %d\n", typeCount["Text"])
	fmt.Printf("  Other:   %d\n", typeCount["Other"])
}

// printBufferCoordinates prints each coordinate of the buffer
// Focus on rows 0-40 to see the panel area in detail
func printBufferCoordinates(buf *paint.Buffer, width, height int) {
	// Print first 40 rows, highlighting the first panel (around y=6-10)
	maxY := 40
	if maxY > height {
		maxY = height
	}

	for y := 0; y < maxY; y++ {
		// Highlight rows that are likely part of the first panel
		rowMarker := "   "
		if y >= 6 && y <= 10 {
			rowMarker = ">>> " // Mark panel area
		}
		fmt.Printf("%sY%02d: ", rowMarker, y)

		// Print each cell in this row
		for x := 0; x < width; x++ {
			cell := buf.GetContent(x, y)
			if cell.Cluster == "" || cell.Cluster == " " {
				fmt.Print(".")
			} else {
				// Highlight border characters
				cluster := cell.Cluster
				runes := []rune(cluster)
				r := runes[0] // Get first rune for comparison

				// Highlight border characters
				if r == '╭' || r == '╮' || r == '╰' || r == '╯' ||
					r == '─' || r == '│' {
					fmt.Printf("[%c]", r) // Border in brackets
				} else if x == 39 || x == 0 {
					// Highlight border edges at column 0 and 39
					fmt.Printf("[%c]", r)
				} else {
					// Regular content - show 2 chars per cell
					if len(cluster) > 1 {
						fmt.Printf("%2s", cluster[:2])
					} else {
						fmt.Printf(" %c", r)
					}
				}
			}
		}
		fmt.Println()
	}

	// Print detailed border character positions for the first panel
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("Border Character Positions (First Panel at Y=6-10):")
	fmt.Println(strings.Repeat("-", 70))

	// Check rows 6-10 for border characters, focusing on edges
	for y := 6; y <= 10; y++ {
		borderChars := []string{}
		// Check left border (x=0)
		for x := 0; x < 5; x++ {
			cell := buf.GetContent(x, y)
			if cell.Cluster != "" && cell.Cluster != " " {
				borderChars = append(borderChars, fmt.Sprintf("(%d,%d)=%s", x, y, cell.Cluster))
			}
		}
		// Check right border (around x=39)
		for x := 35; x < width && x < 42; x++ {
			cell := buf.GetContent(x, y)
			if cell.Cluster != "" && cell.Cluster != " " {
				borderChars = append(borderChars, fmt.Sprintf("(%d,%d)=%s", x, y, cell.Cluster))
			}
		}
		if len(borderChars) > 0 {
			fmt.Printf("Y%02d: %s\n", y, strings.Join(borderChars, ", "))
		}
	}
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Panel Component Demo                               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")

	// Create framework app (required for Fiber reconciler)
	fwApp := framework.NewApp()

	// Create DeclarativeNode WITH Fiber reconciler
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)

	// Enable Fiber-first mode
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	// Create buffer (70 wide, 90 tall)
	buf := paint.NewBuffer(70, 90)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 70, Height: 90},
		AvailableWidth:  70,
		AvailableHeight: 90,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("Rendering Panel components...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	utils.PrintBuffer(buf, 70, 90)

	// Debug: Print each coordinate of the buffer
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Buffer Coordinates (first 40 rows, focused on panel area)")
	fmt.Println(strings.Repeat("=", 70))
	printBufferCoordinates(buf, 70, 90)

	// Get layout boxes for debugging
	// var boxes []*layout.LayoutBox
	// nodeBoxes := node.GetLayoutBoxes()
	// if nodeBoxes != nil {
	// 	boxes = nodeBoxes
	// }

	// Print layout box debug info
	// fmt.Println("\n" + strings.Repeat("=", 70))
	// fmt.Println("Layout Box Debug Info (Flattened)")
	// fmt.Println(strings.Repeat("=", 70))
	// printLayoutBoxes(boxes)

	// Print layout tree with hierarchical structure
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Print(node.GetLayoutTreeString())

	// Print paintable tree with hierarchical structure
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Print(node.GetPaintableTreeString())

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Panel Component Features:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  Layout Options:")
	fmt.Println("    - Width/Height: Fixed dimensions")
	fmt.Println("    - Height=0: Auto height based on content")
	fmt.Println("    - Flex: Flex factor for expansion")
	fmt.Println("    - Padding: Inner padding")
	fmt.Println("")
	fmt.Println("  Border Styles:")
	fmt.Println("    - Rounded: ╭╮╰╯ (default for titled panels)")
	fmt.Println("    - Single:  ┌┐└┘")
	fmt.Println("    - Double:  ╔╗╚╝")
	fmt.Println("    - None:    No border")
	fmt.Println("")
	fmt.Println("  Content Areas:")
	fmt.Println("    - Header: Optional header component")
	fmt.Println("    - Content: Main content (required)")
	fmt.Println("    - Footer: Optional footer component")
	fmt.Println("")
	fmt.Println("  Styling:")
	fmt.Println("    - BorderColor: Color of the border")
	fmt.Println("    - Title: Sets title (appears in border label)")
	fmt.Println("    - Label: Custom border label")
	fmt.Println("")
	fmt.Println("  Convenience Functions:")
	fmt.Println("    - panel.Of(content)")
	fmt.Println("    - panel.OfSize(content, w, h)")
	fmt.Println("    - panel.Titled(title, content)")
	fmt.Println("    - panel.Bordered(content, w, h)")
	fmt.Println(strings.Repeat("=", 70))
}
