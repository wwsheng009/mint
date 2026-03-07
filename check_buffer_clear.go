package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func main() {
	fmt.Println("=== 检查：Render后back buffer是否清理 ===\n")

	renderer := paint.NewRenderer(40, 5)

	// Frame 1
	fmt.Println("Frame 1: SetString(5, 3, \"Button →\")")
	buf := renderer.GetBackBuffer()
	buf.SetString(5, 3, "Button →", style.Style{})

	fmt.Printf("Render前 back buffer 位置12: %q (w=%d)\n",
		buf.GetContent(12, 3).Cluster, buf.GetContent(12, 3).Width)

	renderer.Render()

	// Frame 2: 获取新的back buffer（render后swap）
	buf2 := renderer.GetBackBuffer()
	fmt.Printf("\nFrame 2开始, 新back buffer 位置12: %q (w=%d)\n",
		buf2.GetContent(12, 3).Cluster, buf2.GetContent(12, 3).Width)

	// 检查front buffer
	front := renderer.GetFrontBuffer()
	fmt.Printf("Frame 2开始, front buffer 位置12: %q (w=%d)\n",
		front.GetContent(12, 3).Cluster, front.GetContent(12, 3).Width)

	fmt.Println("\n结论：")
	if buf2.GetContent(12, 3).Cluster == "→" {
		fmt.Println("❌ back buffer保留了旧内容，没有清理！")
	}
	if front.GetContent(12, 3).Cluster == "→" {
		fmt.Println("✓ front buffer保留着旧内容（这是预期的）")
	}
}
