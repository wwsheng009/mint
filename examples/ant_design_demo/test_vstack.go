package main

import (
	"fmt"
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	// 创建嵌套 VStack
	// Root VStack (width=80, auto height)
	//   - Text("Header)        (height=1)
	//   - VStack (gap=1)       (auto height)
	//     - Text("Item 1")    (height=1)
	//     - Text("Item 2")    (height=1)
	//     - Text("Item 3")    (height=1)

	// 构建 VNode 树
	root := ui.VStackBuilder(
		ui.Text("Header"),
		ui.VStackBuilder(
			ui.Text("Item 1"),
			ui.Text("Item 2"),
			ui.Text("Item 3"),
		).Gap(1).Build(),
	).Build()

	// 渲染
	renderer := render.NewRenderer()
	err := renderer.Render(root)
	if err != nil {
		fmt.Printf("渲染失败: %v\n", err)
		return
	}

	// 获取 Fiber
	fiber := renderer.GetFiber()
	if fiber == nil {
		fmt.Println("Fiber 为 nil")
		return
	}

	// 转换为 Node
	node := render.FiberToNode(fiber)

	// 测量 Root VStack
	constraints := layout.Constraints{
		MinWidth:  80,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 9999,
	}
	size := node.Measure(constraints)

	fmt.Printf("Root VStack 约束: W=[80,80] H=[0,9999]\n")
	fmt.Printf("Root VStack 测量结果: W=%d, H=%d\n", size.Width, size.Height)
	fmt.Printf("期望: W=80, H=6 (1 + 1 + 1 + 1 + 1 + 1, 其中 inner VStack H=5)\n")

	// 测量子节点
	if adapter, ok := node.(*render.FiberToNodeAdapter); ok {
		children := adapter.Children()
		if len(children) >= 2 {
			header := children[0]
			headerSize := header.Measure(constraints)
			fmt.Printf("\nHeader: W=%d, H=%d\n", headerSize.Width, headerSize.Height)

			innerVStack := children[1]
			innerConstraints := layout.Constraints{
				MinWidth:  80,
				MaxWidth:  80,
				MinHeight: 0,
				MaxHeight: 9999,
			}
			innerSize := innerVStack.Measure(innerConstraints)
			fmt.Printf("Inner VStack: W=%d, H=%d\n", innerSize.Width, innerSize.Height)
			fmt.Printf("期望: W=80, H=5 (1 + 1 + 1 + 1 + 1, 3 items + 2 gaps)\n")
		}
	}
}
