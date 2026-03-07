package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func main() {
	fmt.Println("=== 检查：SetString写入后的back buffer状态 ===\n")

	renderer := paint.NewRenderer(40, 5)

	// Frame 1
	buf := renderer.GetBackBuffer()
	buf.SetString(5, 3, "Button →", style.Style{})
	renderer.Render()

	// Frame 2
	fmt.Println("Frame 2: SetString(5, 3, \"Button-\")")
	buf2 := renderer.GetBackBuffer()

	fmt.Println("SetString前的状态:")
	for x := 11; x < 14; x++ {
		cell := buf2.GetContent(x, 3)
		fmt.Printf("  位置%d: %q w=%d cont=%v\n", x, cell.Cluster, cell.Width, cell.IsContinuation)
	}

	buf2.SetString(5, 3, "Button-", style.Style{})

	fmt.Println("\nSetString后的状态:")
	fmt.Println(" back buffer:")
	for x := 11; x < 14; x++ {
		cell := buf2.GetContent(x, 3)
		fmt.Printf("  位置%d: %q w=%d cont=%v\n", x, cell.Cluster, cell.Width, cell.IsContinuation)
	}

	front := renderer.GetFrontBuffer()
	fmt.Println(" front buffer:")
	for x := 11; x < 14; x++ {
		cell := front.GetContent(x, 3)
		fmt.Printf("  位置%d: %q w=%d cont=%v\n", x, cell.Cluster, cell.Width, cell.IsContinuation)
	}

	fmt.Println("\n关键：位置12-13在SetString后是否被清空？")
	if buf2.GetContent(12, 3).Cluster == "" {
		fmt.Println("✓ 位置12被清空")
	} else {
		fmt.Printf("❌ 位置12仍然是: %q\n", buf2.GetContent(12, 3).Cluster)
	}
}
