// Test Fixed Positioning - 直接测试 Modal 的 Fixed + AnchorCenter 定位
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/modal"
	uitext "github.com/wwsheng009/mint/ui/components/text"
)

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Test: Modal Fixed Centering (Position=fixed, Anchor=center)    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	// 1. 创建使用 Centered(true) 的 Modal
	centeredModal := modal.NewBuilder().
		Title("Centered Modal (Centered API)").
		Content(uitext.New("Uses .Center() API")).
		Width(38).
		Height(12).
		Centered(true). // ✅ 使用 Centered API
		Open(true).
		Single().
		BuildVNode()

	// 2. 创建使用显式 Props 的 Modal (对比)
	explicitModal := modal.NewBuilder().
		Title("Fixed Modal (Explicit Props)").
		Content(uitext.New("Uses explicit position=fixed, anchor=center props")).
		Width(38).
		Height(10).
		Centered(false).
		Double().
		Open(true).
		BuildVNode()
	explicitModal.SetProps(rtui.Props{
		"position": "fixed",
		"anchor":   "center",
	})

	// 3. 创建 Fiber 树 - Modal 作为 Root 直接子节点
	app := rtui.Fragment(
		uitext.New("Background"),
		centeredModal,
		explicitModal,
	)

	// 4. 创建 DeclarativeNode
	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode { return app })
    node.SetApp(fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	// 5. 创建 viewport 缓冲
	buf := paint.NewBuffer(80, 45)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 45},
		AvailableWidth:  80,
		AvailableHeight: 45,
	}

	fmt.Printf("\n%s Modal 配置 %s\n", strings.Repeat("=", 44), strings.Repeat("=", 44))
	fmt.Println("\nModal 1 (Centered API):")
	fmt.Println("  - width  = 38")
	fmt.Println("  - height = 12")
	fmt.Println("  - Centered(true) → Position=Fixed, Anchor=Center")
	fmt.Println("  Expected: X=(80-38)/2=21, Y=(45-12)/2=16")

	fmt.Println("\nModal 2 (Explicit Props):")
	fmt.Println("  - width  = 38")
	fmt.Println("  - height = 10")
	fmt.Println("  - position=fixed, anchor=center")
	fmt.Println("  Expected: X=(80-38)/2=21, Y=(45-10)/2=17")

	// 6. 渲染前的 Fiber 状态
	fmt.Printf("\n%s Fiber 状态 (渲染前) %s\n", strings.Repeat("=", 40), strings.Repeat("=", 40))
	fiberRoot := node.GetFiberRoot()
	printFiberInfo(fiberRoot, 0)

	// 7. 渲染
	fmt.Printf("\n%s 渲染中... %s\n", strings.Repeat("=", 38), strings.Repeat("=", 38))
	node.Paint(ctx, buf)

	// 8. 显示渲染结果
	fmt.Println("\n=== 渲染输出 ===")
	printBuffer(buf, 80, 45)

	// 9. 显示布局结果
	fmt.Println("\n=== 布局结果 (Layout Boxes) ===")
	boxes := node.GetLayoutBoxes()
	if boxes != nil {
		for i, box := range boxes {
			layerName := getLayerName(rtui.Layer(box.Layer))
			fmt.Printf("\n[Box %d] ID: %s\n", i, box.ID)
			fmt.Printf("  Position: X=%d, Y=%d\n", box.X, box.Y)
			fmt.Printf("  Absolute: AbsX=%d, AbsY=%d\n", box.AbsX, box.AbsY)
			fmt.Printf("  Size: %dx%d\n", box.Width, box.Height)
			fmt.Printf("  Layer: %s (%d)\n", layerName, box.Layer)
			fmt.Printf("  ShouldCenter: %v\n", box.ShouldCenter)

			if box.Layer == 2 { // LayerModal
				expectedX := (80 - box.Width) / 2
				expectedY := (45 - box.Height) / 2
				isCentered := (box.AbsX == expectedX) && (box.AbsY == expectedY)
				status := "❌"
				if isCentered {
					status = "✅"
				}
				fmt.Printf("  Expected (centered): X=%d, Y=%d %s\n", expectedX, expectedY, status)
			}
		}
	}

	fmt.Println("\n=== 结论 ===")
	fmt.Println("如果 Modal 已居中 (AbsX=21, AbsY=16)，说明 Fixed + AnchorCenter 正常工作")
	fmt.Println("如果 Modal 未居中 (AbsX≠21, AbsY≠16)，说明 constraints 传递有问题")
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

func printFiberInfo(fiber *reconciler.Fiber, depth int) {
	if fiber == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	fmt.Printf("%sFiber: Tag=%s, NodeID=%d, Props.centered=%v, Props.position=%v, Props.anchor=%v\n",
		indent, fiber.Tag, fiber.NodeID, fiber.Props["centered"], fiber.Props["position"], fiber.Props["anchor"])

	printFiberInfo(fiber.Child, depth+1)
	printFiberInfo(fiber.Sibling, depth)
}
