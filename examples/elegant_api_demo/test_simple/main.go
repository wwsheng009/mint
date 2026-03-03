package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/paint"
	ui "github.com/wwsheng009/mint/ui"
)

func SimpleMarginTest() rtui.VNode {
	return rtui.VStackBuilder(
		ui.Text("=== Simple Margin Test ==="),
		ui.Text(""),

		// 测试 1: VStack 中的 MarginV
		ui.Text("Test 1: VStack + MarginV"),
		ui.Text("  Btn1: marginV(1,1)"),
		ui.NewButtonBuilder("Btn1").SetID("Btn1").MarginV(1, 1).Build(),
		ui.Text("  Btn2: marginV(1,1)"),
		ui.NewButtonBuilder("Btn2").SetID("Btn2").MarginV(1, 1).Build(),
		ui.Text(""),

		// 测试 2: HStack 中的 MarginV - 垂直间距
		ui.Text("Test 2: HStack + MarginV (vertical/cross-axis)"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("L").SetID("L1").MarginV(1, 1).Build(),
			ui.NewButtonBuilder("R").SetID("R1").MarginV(1, 1).Build(),
		).Gap(1).Build(),
		ui.Text(""),

		// 测试 2.5: HStack 中的 MarginH - 水平间距
		ui.Text("Test 2.5: HStack + MarginH (horizontal/main-axis)"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("L").SetID("L2").MarginH(1, 0).Build(),
			ui.NewButtonBuilder("R").SetID("R2").MarginH(0, 1).Build(),
		).Gap(1).Build(),
		ui.Text(""),

		// 测试 3: 单个按钮 MarginAll
		ui.Text("Test 3: MarginAll(2)"),
		ui.NewButtonBuilder("Big").SetID("Big").MarginAll(2).Build(),
	).Gap(0).Build()
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Simple Margin Test                                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(SimpleMarginTest, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	buf := paint.NewBuffer(80, 25)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 25},
		AvailableWidth:  80,
		AvailableHeight: 25,
	}

	// Render
	node.Paint(ctx, buf)

	fmt.Println("\n=== Render Output ===")
	printBuffer(buf, 80, 25)

	// Get layout boxes
	boxes := node.GetLayoutBoxes()

	fmt.Println("\n=== Layout Boxes (Buttons Only) ===")
	for _, box := range boxes {
		if box.PropsID != "" && (strings.Contains(box.PropsID, "Btn") ||
			strings.Contains(box.PropsID, "L") ||
			strings.Contains(box.PropsID, "R") ||
			box.PropsID == "Big") {
			fmt.Printf("  %-8s | Pos: (%3d,%3d) | Size: %dx%d | PropsID: %s\n",
				box.PropsID, box.X, box.Y, box.Width, box.Height, box.PropsID)
		}
	}

	fmt.Println("\n=== Summary ===")
	analyzeMargins(boxes)
}

func analyzeMargins(boxes []*layout.LayoutBox) {
	// 查找按钮
	var btn1, btn2 *layout.LayoutBox
	var l1, r1 *layout.LayoutBox
	var l2, r2 *layout.LayoutBox
	var big *layout.LayoutBox

	for _, box := range boxes {
		if box.PropsID == "Btn1" {
			btn1 = box
		} else if box.PropsID == "Btn2" {
			btn2 = box
		} else if box.PropsID == "L1" {
			l1 = box
		} else if box.PropsID == "R1" {
			r1 = box
		} else if box.PropsID == "L2" {
			l2 = box
		} else if box.PropsID == "R2" {
			r2 = box
		} else if box.PropsID == "Big" {
			big = box
		}
	}

	// 分析 VStack 中的按钮
	if btn1 != nil && btn2 != nil {
		gap := btn2.Y - (btn1.Y + btn1.Height)
		fmt.Printf("  VStack: Btn1 (Y=%d, H=%d) → Btn2 (Y=%d, H=%d)\n",
			btn1.Y, btn1.Height, btn2.Y, btn2.Height)
		fmt.Printf("    Gap: %d cells (expected: 2 cells for marginV(1,1))\n", gap)
		if gap >= 2 {
			fmt.Printf("    ✓ Vertical margin working\n")
		} else {
			fmt.Printf("    ✗ Vertical margin not working properly\n")
		}
	}

	// 分析 HStack 中的 MarginV (跨轴垂直 margin)
	if l1 != nil && r1 != nil {
		h_gap := r1.X - (l1.X + l1.Width)
		fmt.Printf("  HStack + MarginV: L1 (X=%d, W=%d, Y=%d) → R1 (X=%d, W=%d, Y=%d)\n",
			l1.X, l1.Width, l1.Y, r1.X, r1.Width, r1.Y)
		fmt.Printf("    Horizontal Gap: %d cells (from Gap only, marginV affects vertical)\n", h_gap)
		if l1.Y > r1.Y-5 { // 简单检查它们是否在同一行附近
			fmt.Printf("    Note: MarginV(1,1) only affects vertical spacing in HStack\n")
		}
	}

	// 分析 HStack 中的 MarginH (主轴水平 margin)
	if l2 != nil && r2 != nil {
		h_gap := r2.X - (l2.X + l2.Width)
		fmt.Printf("  HStack + MarginH: L2 (X=%d, W=%d) → R2 (X=%d, W=%d)\n",
			l2.X, l2.Width, r2.X, r2.Width)
		fmt.Printf("    L2: MarginH(1,0) = left=1, right=0\n")
		fmt.Printf("    R2: MarginH(0,1) = left=0, right=1\n")
		fmt.Printf("    Horizontal spacing between buttons: %d cells\n", h_gap)
		if h_gap >= 1 {
			fmt.Printf("    ✓ Horizontal margin working (includes gap and margins)\n")
		} else {
			fmt.Printf("    ✗ Horizontal margin not working properly\n")
		}
	}

	// 分析大按钮
	if big != nil {
		fmt.Printf("  Big Button (MarginAll(2)):\n")
		fmt.Printf("    Position: (%d, %d), Size: %dx%d\n",
			big.X, big.Y, big.Width, big.Height)
		fmt.Printf("    Note: Margins are incorporated into position, not shown separately\n")
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
