// Portal Render Debug
// Demonstrates Portal rendering and box data output
//
// HOW TO RUN:
// This file is designed to be run alongside main.go, so the main function is renamed to DebugMain.
// To run this debug program, you have two options:
//
// Option 1 - Quick Edit (easiest):
//  1. Open this file in an editor
//  2. Find line ~107: "func main() {"
//  3. Change it to: "func main() {"
//  4. Run: go run render_debug.go
//  5. Revert change when done
//
// Option 2 - Build and Run:
//  1. Copy this file to a separate location (e.g., debug_run.go)
//  2. Change line ~107: "func main() {"
//  3. Run: go run debug_run.go
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
	"github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DebugApp shows UI components with Portal rendering
func DebugApp() rtui.VNode {
	// Main layout tree (content sections)
	mainContent := rtui.VStack(
		// Header section
		rtui.VStack(
			newtext.New("=== Portal Rendering Debug ==="),
			newtext.New("This demonstrates the two-phase Portal layout system"),
			newtext.New(""),
		),

		// Button 1 with a tooltip
		rtui.VStack(
			newtext.New("Hover over this button to see a tooltip:"),
			rtui.NewElement("button").SetProps(rtui.Props{
				"text":   "Click Me",
				"width":  15,
				"height": 3,
			}).SetID("button-1"),  // ✨ 使用 SetID() 业务标识
		),

		// Button 2 with another tooltip
		rtui.VStack(
			newtext.New(""),
			newtext.New("And this button:"),
			rtui.NewElement("button").SetProps(rtui.Props{
				"text":   "Another Button",
				"width":  20,
				"height": 3,
			}).SetID("button-2"),  // ✨ 使用 SetID() 业务标识
		),

		// Footer
		rtui.VStack(
			newtext.New(""),
			newtext.New("─────────────────────────────────────────────────────"),
			newtext.New("Portals are rendered independently in Phase 2"),
		),
	)

	// Combine main content with PortalRoot and Portals
	return ui.NewVStack().SetGap(0).SetChildrenList([]rtui.VNode{
		// PortalRoot for tooltips (Phase 2: Overlay Layout target)
		rtui.NewElement("div").SetProps(rtui.Props{
			"portalRootId": "tooltip-root",
		}),

		// Main content
		mainContent,

		// Tooltip 1: PositionFixed, anchored to bottom-left of button-1
		// 使用新的SetID API，其余保持Props方式
		rtui.NewElement("portal").SetProps(rtui.Props{
			"portalRoot": "tooltip-root",
			"anchorId":   "button-1",
			"anchor":     types.AnchorBottomLeft,
			"position":   types.PositionFixed,
			"top":        5,
			"left":       0,
			"priority":   5,
		}).SetChildren([]rtui.VNode{
			newtext.New("Tooltip 1: Below button"),
		}),

		// Tooltip 2: PositionFixed, anchored to top-right of button-2
		rtui.NewElement("portal").SetProps(rtui.Props{
			"portalRoot": "tooltip-root",
			"anchorId":   "button-2",
			"anchor":     types.AnchorTopRight,
			"position":   types.PositionFixed,
			"top":        -10,
			"priority":   10,
		}).SetChildren([]rtui.VNode{
			newtext.New("Tooltip 2: Above button"),
		}),

		// Modal: Centered using AnchorCenter with PositionFixed
		rtui.NewElement("portal").SetProps(rtui.Props{
			"portalRoot": "tooltip-root",
			"anchor":     types.AnchorCenter,
			"position":   types.PositionFixed,
			"top":        0,
			"priority":   1,
		}).SetChildren([]rtui.VNode{
			newtext.New("Centered Modal (Fixed Position)"),
		}),
	})
}

// DebugMain is the entry point for the portal render debug program
// Run with: go run render_debug.go -- (using a build tag or renaming main)
// For simplicity, rename main to DebugMain temporarily when running
func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║       Portal Rendering Debug - Box Data Output           ║")
	fmt.Println("║       Two-Phase Layout System: Main Tree + Portals       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(DebugApp)
    node.SetApp(fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	// Enable Portal-aware layout
	node.SetUsePortalLayout(true)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())
	fmt.Printf("  Fiber-First Enabled: %v\n", node.IsFiberFirstEnabled())
	fmt.Printf("  Portal Layout Enabled: %v\n", node.IsPortalLayoutEnabled())

	// Create buffer
	bufWidth := 80
	bufHeight := 45
	buf := paint.NewBuffer(bufWidth, bufHeight)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: bufWidth, Height: bufHeight},
		AvailableWidth:  bufWidth,
		AvailableHeight: bufHeight,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("Rendering with Portal two-phase layout system...")
	fmt.Printf("  Phase 1: Main Tree Layout (skip Portals)")
	fmt.Printf("  Phase 2: Overlay Layout (each Portal independently)")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	// Render
	node.Paint(ctx, buf)

	fmt.Println("=== Render Output ===")
	printBuffer(buf, bufWidth, bufHeight)

	// Get layout boxes (positions computed by layout engine)
	fmt.Println("\n=== Layout Boxes (Computed by Layout Engine) ===")
	boxes := node.GetLayoutBoxes()
	if boxes != nil {
		fmt.Printf("  Total boxes: %d\n\n", len(boxes))

		// Categorize boxes
		mainTreeBoxes := make([]*layout.LayoutBox, 0)
		portalBoxes := make([]*layout.LayoutBox, 0)

		for _, box := range boxes {
			if len(box.ID) > 6 && box.ID[:6] == "portal-" {
				portalBoxes = append(portalBoxes, box)
			} else {
				mainTreeBoxes = append(mainTreeBoxes, box)
			}
		}

		// Print main tree boxes summary
		fmt.Printf("  [Main Tree] Total: %d boxes\n", len(mainTreeBoxes))
		mainTreeMaxZ := 0
		for _, box := range mainTreeBoxes {
			if box.ZIndex > mainTreeMaxZ {
				mainTreeMaxZ = box.ZIndex
			}
		}
		fmt.Printf("  [Main Tree] Max Z-index: %d\n\n", mainTreeMaxZ)

		// Print portal boxes in detail
		fmt.Printf("  [Portal Boxes] Total: %d\n", len(portalBoxes))
		if len(portalBoxes) > 0 {
			fmt.Println()
			for i, box := range portalBoxes {
				fmt.Printf("  [Portal #%d]\n", i+1)
				fmt.Printf("    ID:         %s\n", box.ID)
				fmt.Printf("    Position:   X=%d, Y=%d\n", box.X, box.Y)
				fmt.Printf("    Abs Pos:    AbsX=%d, AbsY=%d\n", box.AbsX, box.AbsY)
				fmt.Printf("    Size:       Width=%d, Height=%d\n", box.Width, box.Height)
				fmt.Printf("    Z-index:    %d (PortalBase=%d + priority)\n", box.ZIndex, 1000)
				fmt.Printf("    ShouldCenter: %v\n", box.ShouldCenter)
				if len(box.Children) > 0 {
					fmt.Printf("    Children:   %d\n", len(box.Children))
				}
				fmt.Println()
			}
		}

		// Z-index comparison
		fmt.Printf("  [Z-index Layer Separation]\n")
		fmt.Printf("    Main Tree Z-index range: 0 - %d\n", mainTreeMaxZ)
		if len(portalBoxes) > 0 {
			minPortalZ := portalBoxes[0].ZIndex
			maxPortalZ := portalBoxes[0].ZIndex
			for _, box := range portalBoxes {
				if box.ZIndex < minPortalZ {
					minPortalZ = box.ZIndex
				}
				if box.ZIndex > maxPortalZ {
					maxPortalZ = box.ZIndex
				}
			}
			fmt.Printf("    Portal Z-index range: %d - %d (1000+ base)\n", minPortalZ, maxPortalZ)
			if minPortalZ > mainTreeMaxZ {
				fmt.Printf("    ✓ Portals will render ABOVE main tree (%d > %d)\n", minPortalZ, mainTreeMaxZ)
			}
		}
		fmt.Println()
	}

	// Get portal boxes specifically
	fmt.Println("=== Portal Boxes (GetPortalBoxes()) ===")
	portalBoxes := node.GetPortalBoxes()
	if portalBoxes != nil {
		fmt.Printf("  Portal boxes count: %d\n", len(portalBoxes))
		for i, box := range portalBoxes {
			fmt.Printf("  [Portal %d] ID=%s, PropsID='%s', pos=(%d,%d), size=%dx%d, ZIndex=%d\n",
				i, box.ID, box.PropsID, box.X, box.Y, box.Width, box.Height, box.ZIndex)
		}
	} else {
		fmt.Println("  No portal boxes found")
	}
	fmt.Println()

	// ✨ Debug: Show all boxes with PropsID to verify SetID() is working
	fmt.Println("=== Debug: Boxes with PropsID (SetID lookup table) ===")
	var debugBoxes func(box *layout.LayoutBox, indent int)
	debugBoxes = func(box *layout.LayoutBox, indent int) {
		if box == nil {
			return
		}
		if box.PropsID != "" {
			tag := ""
			if box.BoxModel.Border.Label != "" {
				tag = fmt.Sprintf("[%s]", box.BoxModel.Border.Label)
			}
			fmt.Printf("  %sBox: ID=%s %s PropsID='%s', pos=(%d,%d), size=%dx%d, children=%d\n",
				strings.Repeat("  ", indent), box.ID, tag, box.PropsID, box.X, box.Y, box.Width, box.Height, len(box.Children))
		}
		for _, child := range box.Children {
			debugBoxes(child, indent+1)
		}
	}
	debugBoxes(node.GetLayoutBoxes()[0], 0)
	fmt.Println()

	// ✨ Debug: Print all boxes in the layout tree to understand structure
	fmt.Println("=== Debug: All Boxes (Full Tree) ===")
	var printAllBoxes func(box *layout.LayoutBox, indent int)
	printAllBoxes = func(box *layout.LayoutBox, indent int) {
		if box == nil {
			return
		}
		tag := ""
		if box.BoxModel.Border.Label != "" {
			tag = fmt.Sprintf("[%s]", box.BoxModel.Border.Label)
		}
		propsIDInfo := ""
		if box.PropsID != "" {
			propsIDInfo = fmt.Sprintf(", PropsID='%s'", box.PropsID)
		}
		typeStr := box.ID
		if len(typeStr) > 4 {
			typeStr = typeStr[:4]
		}
		fmt.Printf("  %sBox (ID=%s%s%s): pos=(%d,%d), size=%dx%d\n",
			strings.Repeat("  ", indent), box.ID, tag, propsIDInfo, box.X, box.Y, box.Width, box.Height)
		for _, child := range box.Children {
			printAllBoxes(child, indent+1)
		}
	}
	printAllBoxes(node.GetLayoutBoxes()[0], 0)
	fmt.Println()

	// ✨ Debug: Test FindAnchorPosition manually
	fmt.Println("=== Debug: Test FindAnchorPosition ===")
	boxes = node.GetLayoutBoxes()
	if len(boxes) > 0 && boxes[0] != nil {
		x1, y1, w1, h1, found1 := layout.FindAnchorPosition(boxes[0], "button-1")
		fmt.Printf("  FindAnchorPosition('button-1'): found=%v, pos=(%d,%d), size=%dx%d\n", found1, x1, y1, w1, h1)

		x2, y2, w2, h2, found2 := layout.FindAnchorPosition(boxes[0], "button-2")
		fmt.Printf("  FindAnchorPosition('button-2'): found=%v, pos=(%d,%d), size=%dx%d\n", found2, x2, y2, w2, h2)
	}
	fmt.Println()

	// Get paintable boxes
	fmt.Println("=== Paintable Boxes (After Layer Normalization) ===")
	paintableBoxes := node.GetPaintableBoxes()
	if paintableBoxes != nil {
		fmt.Printf("  Total paintable boxes: %d\n", len(paintableBoxes))

		// Find portal paintable boxes (node is nil for portal boxes)
		portalPaintableCount := 0
		portalMaxZ := 0
		mainTreeMaxZ := 0

		for _, box := range paintableBoxes {
			if box.Node == nil {
				portalPaintableCount++
				if box.ZIndex > portalMaxZ {
					portalMaxZ = box.ZIndex
				}
			} else {
				if box.ZIndex > mainTreeMaxZ {
					mainTreeMaxZ = box.ZIndex
				}
			}

			// Log high Z-index boxes (likely portals)
			if box.ZIndex >= 1000 {
				fmt.Printf("  [High-Z] pos=(%d,%d), size=%dx%d, ZIndex=%d, Node=%v\n",
					box.X, box.Y, box.Width, box.Height, box.ZIndex, box.Node != nil)
			}
		}

		fmt.Printf("\n  Summary:\n")
		fmt.Printf("    Main tree boxes: %d (max ZIndex=%d)\n", len(paintableBoxes)-portalPaintableCount, mainTreeMaxZ)
		fmt.Printf("    Portal boxes: %d (max ZIndex=%d)\n", portalPaintableCount, portalMaxZ)
	}
	fmt.Println()

	fmt.Println("=== Architecture Review ===")
	fmt.Println("  Two-Phase Layout System:")
	fmt.Println("    Phase 1 - Main Tree Layout:")
	fmt.Println("      - Traverse the main Fiber tree")
	fmt.Println("      - Skip Portal nodes (collect to queue)")
	fmt.Println("      - Normal layout for all other nodes")
	fmt.Println("")
	fmt.Println("    Phase 2 - Overlay Layout:")
	fmt.Println("      - Layout each Portal independently using Root coordinates")
	fmt.Println("      - Use PortalRoot as the anchor/positioning context")
	fmt.Println("      - Support PositionFixed and Anchor-based positioning")
	fmt.Println("      - Merge Portal layout results into final layout tree")
	fmt.Println("")
	fmt.Println("  Z-index Strategy:")
	fmt.Println("    - Main tree: Z-index 0 - ~100 (depth-based)")
	fmt.Println("    - Portals: Z-index 1000+ (PortalBase 1000 + priority)")
	fmt.Println("    - Ensures Portals render ABOVE main tree")

	fmt.Println("\n=== Expected Results ===")
	fmt.Println("  ✓ Tooltip 1 should be positioned below button-1")
	fmt.Println("  ✓ Tooltip 2 should be positioned above button-2")
	fmt.Println("  ✓ Centered Modal should appear in middle of screen")
	fmt.Println("  ✓ All Portals should have Z-index >= 1000")
	fmt.Println("  ✓ Portals should render on top of main content")

	fmt.Println("\n=== Test Complete ===")
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

// For running this file directly, uncomment the following line:
// func main() { DebugMain() }
//
// Quick run command from portal_demo directory:
//   sed 's/func DebugMain/\/\/ func DebugMain/' render_debug.go | sed 's/\/\/ func main()/func main()/' | sed '1s/^/\/\/ Quick fix - temp main func added\n/' | go run -
//
// Or use this simpler approach:
//   1. Edit render_debug.go
//   2. Change line ~107: DebugMain -> main
//   3. Add a new line after line ~236: // func main() { DebugMain() }
//   4. Run: go run render_debug.go
