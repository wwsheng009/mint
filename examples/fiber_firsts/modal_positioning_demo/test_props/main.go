// Test Props - 测试 Modal 属性正确传递到 Fiber
package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	uitext "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/modal"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Test: Modal Props 传递到 Fiber                               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	// 1. 创建两个 Modal：一个使用 Centered() API，一个使用显式 Props
	centeredModal := modal.NewBuilder().
		Title("Centered Modal").
		Content(uitext.New("Uses Centered()")).
		Width(38).
		Height(12).
		Centered(true). // ✅ Centered API
		Open(true).
		BuildVNode()

	explicitModal := modal.NewBuilder().
		Title("Explicit Modal").
		Content(uitext.New("Uses explicit props")).
		Width(38).
		Height(10).
		Centered(false).
		BuildVNode()
	explicitModal.SetProps(rtui.Props{
		"position": "fixed",
		"anchor":   "center",
	})

	// 2. 创建 VNode 树
	app := rtui.Fragment(
		uitext.New("Background"),
		centeredModal,
		explicitModal,
	)

	// 3. 创建 Fiber 树
	fmt.Printf("\n%s Fiber 树结构 %s\n", strings.Repeat("=", 54), strings.Repeat("=", 54))
	
	fiberRoot := reconciler.CreateFiberFromVNode(app)
	printFiberPropsBefore(fiberRoot, "Before SyncPositioningProperties")

	// 4. 应用 SyncPositioningProperties
	fmt.Printf("\n%s 应用 SyncPositioningProperties %s\n", strings.Repeat("=", 52), strings.Repeat("=", 52))
	
	child := fiberRoot.Child
	for child != nil {
		if child.Tag == "modal" {
			reconciler.SyncPositioningProperties(child)
		}
		child = child.Sibling
	}

	printFiberPropsAfter(fiberRoot, "After SyncPositioningProperties")

	// 5. 验证结果
	fmt.Printf("\n%s 验证结果 %s\n", strings.Repeat("=", 58), strings.Repeat("=", 58))
	
	allPassed := true
	
	// 检查 centeredModal
	centeredModalFiber := findFiberByTag(fiberRoot, "Centered Modal")
	if centeredModalFiber != nil {
		if centeredModalFiber.Position == types.PositionFixed && centeredModalFiber.Anchor == types.AnchorCenter {
			fmt.Printf("✅ centeredModal: Position=fixed, Anchor=center (正确)\n")
		} else {
			fmt.Printf("❌ centeredModal: Position=%v, Anchor=%v (错误，期望 PositionFixed + AnchorCenter)\n", 
				centeredModalFiber.Position, centeredModalFiber.Anchor)
			allPassed = false
		}
	}

	// 检查 explicitModal
	explicitModalFiber := findFiberByTag(fiberRoot, "Explicit Modal")
	if explicitModalFiber != nil {
		if explicitModalFiber.Position == types.PositionFixed && explicitModalFiber.Anchor == types.AnchorCenter {
			fmt.Printf("✅ explicitModal: Position=fixed, Anchor=center (正确)\n")
		} else {
			fmt.Printf("❌ explicitModal: Position=%v, Anchor=%v (错误)\n", explicitModalFiber.Position, explicitModalFiber.Anchor)
			allPassed = false
		}
	}

	fmt.Println()
	if allPassed {
		fmt.Println("🎉 所有测试通过！Modal Props 正确传递到 Fiber")
	} else {
		fmt.Println("⚠️  存在问题！Modal Props 传递不正确")
	}
}

func printFiberPropsBefore(fiber *reconciler.Fiber, title string) {
	fmt.Printf("\n--- %s ---\n", title)
	printFiberRecursive(fiber, 0, nil)
}

func printFiberPropsAfter(fiber *reconciler.Fiber, title string) {
	fmt.Printf("\n--- %s ---\n", title)
	printFiberRecursive(fiber, 0, nil)
}

func printFiberRecursive(fiber *reconciler.Fiber, depth int, visited map[uint64]bool) {
	if fiber == nil {
		return
	}

	if visited == nil {
		visited = make(map[uint64]bool)
	}

	visited[fiber.NodeID] = true

	indent := strings.Repeat("  ", depth)
	
	props := ""
	if fiber.Props != nil {
		centered := ""
		if c, ok := fiber.Props["centered"].(bool); ok {
			centered = fmt.Sprintf("centered=%v", c)
		}
		position := ""
		if p, ok := fiber.Props["position"].(string); ok {
			position = fmt.Sprintf("position=%q", p)
		}
		anchor := ""
		if a, ok := fiber.Props["anchor"].(string); ok {
			anchor = fmt.Sprintf("anchor=%q", a)
		}
		props = fmt.Sprintf("[%s, %s, %s]", centered, position, anchor)
	}

	fmt.Printf("%sFiber NodeID=%d, Tag=%q, Position=%v, Anchor=%v %s\n",
		indent, fiber.NodeID, fiber.Tag, fiber.Position, fiber.Anchor, props)

	printFiberRecursive(fiber.Child, depth+1, visited)
	printFiberRecursive(fiber.Sibling, depth, visited)
}

func findFiberByTag(fiber *reconciler.Fiber, tag string) *reconciler.Fiber {
	if fiber == nil {
		return nil
	}
	
	if fiber.Tag == tag {
		return fiber
	}
	
	if result := findFiberByTag(fiber.Child, tag); result != nil {
		return result
	}
	
	return findFiberByTag(fiber.Sibling, tag)
}
