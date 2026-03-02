package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/paint"
	ui "github.com/wwsheng009/mint/ui"
)

// ElegantAPIDemo - demo component for testing margin
func ElegantAPIDemo() rtui.VNode {
	return rtui.VStackBuilder(
		ui.Text("✨ Margin Debug"),
		ui.Text("────────────────────────────────"),
		ui.Text("3. Buttons with MarginV(1, 1):"),
		ui.NewButtonBuilder("Btn1").MarginV(1, 1).Build(),
		ui.NewButtonBuilder("Btn2").MarginV(1, 1).Build(),
		ui.NewButtonBuilder("Btn3").MarginV(1, 1).Build(),
	).Gap(0).Build()
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Margin Debug Test                                            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(ElegantAPIDemo, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())
	fmt.Printf("  Fiber-First Enabled: %v\n", node.IsFiberFirstEnabled())

	// Create buffer
	buf := paint.NewBuffer(80, 25)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 25},
		AvailableWidth:  80,
		AvailableHeight: 25,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("Rendering with margin support...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	// Render
	node.Paint(ctx, buf)

	fmt.Println("\n=== Render Output ===")
	printBuffer(buf, 80, 25)

	// Get layout boxes
	fmt.Println("\n=== Layout Boxes (Computed by Layout Engine) ===")
	boxes := node.GetLayoutBoxes()
	if boxes != nil {
		buttonCount := 0
		for _, box := range boxes {
			// 显示所有 button 类型节点
			if strings.Contains(box.ID, "Btn") || strings.Contains(strings.ToLower(box.ID), "button") {
				buttonCount++
				fmt.Printf("  [Button #%d] Node ID: %s\n", buttonCount, box.ID)
				fmt.Printf("             Position: X=%d, Y=%d\n", box.X, box.Y)
				fmt.Printf("             Size: Width=%d, Height=%d\n", box.Width, box.Height)
				fmt.Printf("             Layer: %d\n\n", box.Layer)
			}
		}
	}

	fmt.Println("\n=== Expected Margin Results ===")
	fmt.Println("  Btn1, Btn2, Btn3: Should have marginV(1, 1) - spaced apart")
	fmt.Println("  Between buttons: Should see 1 line gap between each button")

	// 打印 Fiber 结构中所有按钮的 LayoutMargin
	fmt.Println("\n=== Fiber LayoutMargin Values ===")
	printFiberLayoutMargin(node)

	fmt.Println("\n=== Test Complete ===")
}

func printFiberLayoutMargin(node *render.DeclarativeNode) {
	// 通过 NodeID 打印调试信息
	// GetLayoutBoxes() 返回的是 LayoutBox，但不包含 Fiber 信息
	// 我们需要通过 Fiber 树遍历来获取 LayoutMargin

	// 获取 Fiber 根节点（需要通过内部 API）
	// 但这里简化处理，只打印提示信息
	fmt.Println("  Note: GetLayoutBoxes() doesn't expose Fiber.LayoutMargin")
	fmt.Println("  The layout engine should use FiberToNodeAdapter.GetMargin()")
	fmt.Println("  which reads from Fiber.LayoutMargin field")
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

