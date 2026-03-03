package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
)

// 测试 margin 对测量的影响
func MarginMeasureTest() rtui.VNode {
	return rtui.VStackBuilder(
		ui.Text("=== Margin & Measure Test ==="),
		ui.Text(""),
		ui.Text("Parent: width=60, height=20"),
		ui.Text(""),

		// 测试 1: 子节点宽度 + margin = 父容器宽度
		ui.Text("Test 1: Child with margin that might overflow"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("W30").SetID("W30").
				Flex(1).  // 占用全部可用空间
				MarginAll(5).Build(),
		).
			Gap(0).
			Build(),

		ui.Text(""),
		ui.Text("Test 2: Multiple children with different margins"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("A").SetID("A").
				Flex(1).
				MarginH(10, 0).Build(),  // left margin: 10
			ui.NewButtonBuilder("B").SetID("B").
				Flex(1).
				MarginH(0, 10).Build(), // right margin: 10
		).
			Gap(1).
			Build(),

		ui.Text(""),
		ui.Text("Test 3: VStack children with vertical margins"),
		rtui.VStackBuilder(
			ui.NewButtonBuilder("C1").SetID("C1").
				Flex(1).MarginV(5, 5).Build(),
			ui.NewButtonBuilder("C2").SetID("C2").
				Flex(1).MarginV(5, 5).Build(),
		).
			Gap(0).
			Build(),

		ui.Text(""),
		ui.Text("Test 4: Button with large margin in small container"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("Small").SetID("Small").
				MarginAll(10).Build(),  // Large margin
			ui.NewButtonBuilder("Btn").SetID("Btn").
				MarginH(5, 5).Build(),
		).
			Gap(2).
			Build(),
	).
		Gap(0).
		Build()
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Margin & Measure Test                                      ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║   Testing how margin affects measurement and constraints    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(MarginMeasureTest)
    node.SetApp(fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	// 渲染
	buf := paint.NewBuffer(80, 25)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 25},
		AvailableWidth:  80,
		AvailableHeight: 25,
	}

	node.Paint(ctx, buf)

	fmt.Println("\n=== Rendering Output ===")
	printBuffer(buf, 80, 25)

	// 分析 Layout Boxes
	boxes := node.GetLayoutBoxes()
	analyzeMeasurements(boxes)
}

func analyzeMeasurements(boxes []*layout.LayoutBox) {
	fmt.Println("\n=== Layout Box Analysis ===")

	// 查找所有容器和子节点
	parentHStack1 := findByPropsID(boxes, "hstack-1")
	childW30 := findByPropsID(boxes, "W30")

	parentHStack2 := findByPropsID(boxes, "hstack-2")
	childA := findByPropsID(boxes, "A")
	childB := findByPropsID(boxes, "B")

	parentVStack3 := findByPropsID(boxes, "vstack-3")
	childC1 := findByPropsID(boxes, "C1")
	childC2 := findByPropsID(boxes, "C2")

	parentHStack4 := findByPropsID(boxes, "hstack-4")
	childSmall := findByPropsID(boxes, "Small")
	childBtn := findByPropsID(boxes, "Btn")

	fmt.Println("\n--- Test 1: Single child with margin ---")
	if parentHStack1 != nil {
		fmt.Printf("  Parent HStack: Pos=(%d,%d), Size=%dx%d\n",
			parentHStack1.X, parentHStack1.Y,
			parentHStack1.Width, parentHStack1.Height)
	}
	if childW30 != nil {
		fmt.Printf("  Child W30 (MarginAll(5)): Pos=(%d,%d), Size=%dx%d, Layer=%d\n",
			childW30.X, childW30.Y,
			childW30.Width, childW30.Height, childW30.Layer)

		// child 的 X 是否等于 5 (marginLeft)?
		expectedX := 5
		if childW30.X == expectedX {
			fmt.Printf("    ✓ Left margin (5) applied correctly\n")
		} else {
			fmt.Printf("    ✗ Left margin not applied (expected X=5, got X=%d)\n", childW30.X)
		}

		if parentHStack1 != nil {
			// child 右边界是否在父容器 - marginRight 之内？
			childRight := childW30.X + childW30.Width
			expectedRight := parentHStack1.X + parentHStack1.Width - 5
			if childRight <= expectedRight {
				fmt.Printf("    ✓ Child within parent bounds (right margin respected)\n")
			} else {
				fmt.Printf("    ✗ Child may overflow parent (childRight=%d, expected <= %d)\n",
					childRight, expectedRight)
			}
		}
	}

	fmt.Println("\n--- Test 2: Two children with asymmetric margins ---")
	if parentHStack2 != nil {
		fmt.Printf("  Parent HStack: Pos=(%d,%d), Size=%dx%d\n",
			parentHStack2.X, parentHStack2.Y,
			parentHStack2.Width, parentHStack2.Height)
	}
	if childA != nil && childB != nil {
		fmt.Printf("  Child A (MarginH(10,0)): Pos=(%d,%d), Size=%dx%d, Layer=%d\n",
			childA.X, childA.Y, childA.Width, childA.Height, childA.Layer)
		fmt.Printf("  Child B (MarginH(0,10)): Pos=(%d,%d), Size=%dx%d, Layer=%d\n",
			childB.X, childB.Y, childB.Width, childB.Height, childB.Layer)

		if parentHStack2 != nil {
			// Check if A's left margin is applied
			if childA.X >= parentHStack2.X+10 {
				fmt.Printf("    ✓ Child A has left margin >= 10\n")
			} else {
				fmt.Printf("    ✗ Child A left margin not applied (X=%d, expected > %d)\n",
					childA.X, parentHStack2.X+10)
			}

			// Check spacing between A and B
			h_gap := childB.X - (childA.X + childA.Width)
			fmt.Printf("    Gap between A and B: %d cells (includes gap=1 + margins)\n", h_gap)

			// Check if B's right margin is respected
			childBRight := childB.X + childB.Width
			parentRight := parentHStack2.X + parentHStack2.Width
			if parentRight - childBRight >= 10 {
				fmt.Printf("    ✓ Child B has right margin >= 10\n")
			} else {
				fmt.Printf("    ✗ Child B right margin violated (available=%d, required=10)\n",
					parentRight-childBRight)
			}
		}
	}

	fmt.Println("\n--- Test 3: VStack children with vertical margins ---")
	if parentVStack3 != nil {
		fmt.Printf("  Parent VStack: Pos=(%d,%d), Size=%dx%d\n",
			parentVStack3.X, parentVStack3.Y,
			parentVStack3.Width, parentVStack3.Height)
	}
	if childC1 != nil && childC2 != nil {
		fmt.Printf("  Child C1 (MarginV(5,5)): Pos=(%d,%d), Size=%dx%d, Layer=%d\n",
			childC1.X, childC1.Y, childC1.Width, childC1.Height, childC1.Layer)
		fmt.Printf("  Child C2 (MarginV(5,5)): Pos=(%d,%d), Size=%dx%d, Layer=%d\n",
			childC2.X, childC2.Y, childC2.Width, childC2.Height, childC2.Layer)

		if parentVStack3 != nil {
			// Check vertical spacing
			v_gap := childC2.Y - (childC1.Y + childC1.Height)
			fmt.Printf("    Gap between C1 and C2: %d cells\n", v_gap)
			if v_gap >= 10 {
				fmt.Printf("    ✓ Vertical margins (5+5=10) working\n")
			} else {
				fmt.Printf("    ✗ Vertical margins may not be fully working (expected >= 10)\n")
			}
		}
	}

	fmt.Println("\n--- Test 4: Large margin in constrained container ---")
	if parentHStack4 != nil {
		fmt.Printf("  Parent HStack: Pos=(%d,%d), Size=%dx%d\n",
			parentHStack4.X, parentHStack4.Y,
			parentHStack4.Width, parentHStack4.Height)
	}
	if childSmall != nil && childBtn != nil {
		fmt.Printf("  Child Small (MarginAll(10)): Pos=(%d,%d), Size=%dx%d, Layer=%d\n",
			childSmall.X, childSmall.Y, childSmall.Width, childSmall.Height, childSmall.Layer)
		fmt.Printf("  Child Btn (MarginH(5,5)): Pos=(%d,%d), Size=%dx%d, Layer=%d\n",
			childBtn.X, childBtn.Y, childBtn.Width, childBtn.Height, childBtn.Layer)

		if parentHStack4 != nil {
			// Check total width including margins
			childSmallTotalWidth := childSmall.X + childSmall.Width + 10 - parentHStack4.X
			if childSmallTotalWidth > parentHStack4.Width {
				fmt.Printf("    ⚠ Child Small margin may cause width overflow\n")
				fmt.Printf("    But flex layout should have reduced content width to fit\n")
			}
		}
	}

	fmt.Println("\n=== Key Findings ===")
	fmt.Println("1. Parent constraints are applied BEFORE child measurement")
	fmt.Println("2. Margins are DEDUCTED from child constraints in layout phase:")
	fmt.Println("   - Main axis: Accumulated in spacing offset")
	fmt.Println("   - Cross axis: Handled separately")
	fmt.Println("3. Child width/height are measured WITHOUT margins")
	fmt.Println("4. Margins are then added to child position")
	fmt.Println("5. Child may overflow parent if (content + margins) > parent")
	fmt.Println("\n   Example (HStack, Parent Width: 60):")
	fmt.Println("   Child content: 50 + Margin(10, 10) = Total 70 → OVERFLOW!")
	fmt.Println("   But Flex layout reduces child content width to fit:")
	fmt.Println("   Child content: 40 + Margin(10, 10) = Total 60 → OK!")
}

func findByPropsID(boxes []*layout.LayoutBox, propsID string) *layout.LayoutBox {
	for _, box := range boxes {
		if box.PropsID == propsID {
			return box
		}
	}
	return nil
}

func printBuffer(buf *paint.Buffer, w, h int) {
	fmt.Println("┌" + "─" + "┐") // Simplified
	for y := 0; y < h && y < len(buf.Cells); y++ {
		line := "|"
		for x := 0; x < w && x < len(buf.Cells[y]); x++ {
			cell := buf.Cells[y][x]
			if len(cell.Cluster) == 0 || cell.Cluster == " " {
				line += " "
			} else {
				line += string([]rune(cell.Cluster)[0])
			}
		}
		line += "|"
		fmt.Println(line)
	}
	fmt.Println("└" + "─" + "┘")
}
