// Test Constraints Flow - 分析 Modal 约束传递流程
package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime/types"
	"github.com/wwsheng009/mint/runtime/layout"
	uitext "github.com/wwsheng009/mint/runtime/ui"
	uitext "github.com/wwsheng009/mint/ui/components/text"
	uimodal "github.com/wwsheng009/mint/ui/components/modal"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Test: Modal 约束传递流程验证                                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════░")
	fmt.Println()

	// 创建一个 Modal（Centered API）
	centeredModal := uimodal.New().
		SetTitle("Test Modal").
		SetWidth(38).
		SetHeight(12).
		Centered(true).
		Open(true).
		BuildVNode()

	// 创建应用树 - Modal 直接作为 VStack 的第一个子元素
	app := rtui.VStack(
		centeredModal, // ✅ Modal 作为第一个子节点
	)

	// 创建 Fiber 树
	fiberRoot := reconciler.BuildFiberTree(app)

	fmt.Println("\n=== Fiber 树结构 ===")
	printFiberTree(fiberRoot, 0)

	// 应用 SyncPositioningProperties
	fmt.Println("\n=== 应用 SyncPositioningProperties ===")
	applySync(fiberRoot)

	// 创建布局引擎
	engine := layout.NewEngine()
	viewportConstraints := layout.Constraints{
		MinWidth:   0,
	MaxWidth:   80,
		MinHeight: 0,
		MaxHeight: 45,
	}
	fmt.Printf("Layout Engine viewportConstraints: (%d,%d)→(%dx%d)\n",
		viewportConstraints.MinWidth, viewportConstraints.MinHeight,
		viewportConstraints.MaxWidth, viewportConstraints.MaxHeight)

	// 查找 Modal 的 Fiber 节点
	fmt.Println("\n=== 查找 Modal Fiber ===")
	findModalAndShowInfo(fiberRoot, 0)

	// 创建应用树并使用 Modal 作为直接子节点
	fmt.Println("\n=== 测试 1: Modal 作为 VStack 直接子节点 ===")
	appDirect := centeredModal
	appDirectVStack := rtui.VStack(appDirect)
	fiberRootDirect := reconciler.BuildFiberTree(appDirectVStack)
	
	findModalAndShowInfo(fiberRootDirect, 0)
}

func printFiberTree(fiber *reconciler.Fiber, depth int) {
	if fiber == nil {
		return
	}
	
	indent := strings.Repeat("  ", depth)
	
	props := ""
	if fiber.Props != nil {
		centered, _ := fiber.Props["centered"].(bool)
		position, _ := fiber.Props["position"].(string)
		anchor, _ := fiber.Props["anchor"].(string)
		
		details := []string{}
		if hasCentered {
			details = append(details, fmt.Sprintf("centered=%v", centered))
		}
		if len(fiber.Children) > 0 {
			childTag := ""
			if c := fiber.Children[0]; c != nil {
				childTag = c.Type().String()
			}
			details = append(details, fmt.Sprintf("child=%s", childTag))
		}
		
		if len(details) > 0 {
			props = "[" + strings.Join(details, ", ") + "]"
		}
	
	fmt.Printf("%sFiber NodeID=%d, Tag=%q, %s\n",
		indent, fiber.NodeID, fiber.Tag, props)

	printFiberTree(fiber.Child, depth+1)
	printFiberTree(fiber.Sibling, depth)
}

func findModalAndShowInfo(fiber *reconciler.Fiber, depth int) {
	if fiber == nil {
		return
	}
	
	if fiber.Tag == "modal" || fiber.Tag == "Test Modal" || fiber.Tag == "TestModal" {
		fmt.Printf("📍 找到 Modal (NodeID=%d, Tag=%q)!\n", fiber.NodeID, fiber.Tag)
		fmt.Printf("   Before: Position=%v, Anchor=%v, Layer=%d\n", fiber.Position, fiber.Anchor, fiber.Layer)
		
		// 找到所有 Modal siblings（包括 Fiber.Sibling）
		fmt.Printf("   Modal 在层级中的位置:\n")
		showFiberInfo(fiber)
		printFiberTree(fiber, depth+1)
		
		// 应用 SyncPositioningProperties
		fmt.Printf("\n   应用 SyncPositioningProperties...\n")
		reconciler.SyncPositioningProperties(fiber)
		fmt.Printf("   After:  Position=%v, Anchor=%v, Layer=%d\n", fiber.Position, fiber.Anchor, fiber.Layer)
		
		// 验证结果
		if fiber.Position == types.PositionFixed && fiber.Anchor == types.AnchorCenter {
			fmt.Printf("   ✅ Modal 使用 Fixed + AnchorCenter 定位！\n")
		} else {
			fmt.Printf("   ❌ Modal 未使用 Fixed 定位，被父布局流控制\n")
		}
	} else {
		printFiberTree(fiber.Child, depth+1)
		printFiberTree(fiber.Sibling, depth)
	}
}

func showFiberInfo(fiber *reconciler.Fiber) {
	if fiber.Props == nil {
		return
	}
	
	fmt.Printf("   Props: centered=%v, position=%q, anchor=%q\n",
		fiber.Props["centered"], fiber.Props["position"], fiber.Props["anchor"])
}

func printFiberInfo(fiber *reconciler.Fiber) {
	fmt.Printf("   布局尺寸: Width=%d, Height=%d\n", fiber.GetInstance().GetSize())
	fmt.Printf("   子节点数: %d\n", len(fiber.Children()))
	if len(fiber.Children()) > 0 {
		child := fiber.Children()[0]
		fmt.Printf("   第一个子元素: Tag=%q\n", child.Type().String())
	}
}
