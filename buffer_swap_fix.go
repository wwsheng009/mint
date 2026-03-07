package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func main() {
	fmt.Println("=== 测试：正确的buffer管理 ===\n")

	renderer := paint.NewRenderer(40, 10)

	// Frame 1: 正确的方式 - 每次Render前重新获取back buffer
	fmt.Println("Frame 1: 'Button →'")
	buf1 := renderer.GetBackBuffer()
	buf1.SetString(5, 3, "Button →", style.Style{})
	fmt.Printf("  buf1 address: %p (should be back buffer)\n", buf1)

	frame1Output := renderer.Render()
	fmt.Printf("  Output: %q\n", frame1Output)

	// Frame 2: 重新获取back buffer（正确的方式）
	fmt.Println("\nFrame 2: 'Button-'")
	buf2 := renderer.GetBackBuffer()  ← 关键：重新获取！
	fmt.Printf("  buf2 address: %p (should be back buffer)\n", buf2)
	fmt.Printf("  buf1 address: %p (now is front buffer)\n", buf1)
	buf2.SetString(5, 3, "Button-", style.Style{})

	frame2Output := renderer.Render()
	fmt.Printf("  Output: %q\n", frame2Output)

	// Check
	frontBuf := renderer.GetFrontBuffer()
	fmt.Printf("\nFront buffer after Frame 2:\n")
	for x := 11; x < 14; x++ {
		cell := frontBuf.GetContent(x, 3)
		cluster := cell.Cluster
		if cluster == "" {
			if cell.IsContinuation {
				cluster = "[cont]"
			} else {
				cluster = "[empty]"
			}
		}
		fmt.Printf("  [%d]: %q w=%d cont=%v\n", x, cluster, cell.Width, cell.IsContinuation)
	}

	// 检查输出
	if strings.Contains(frame2Output, "  ") && strings.Count(frame2Output, "  ") >= 2 {
		fmt.Println("\n✅ Output has 2 spaces (correct for width=2 arrow)")
	} else if strings.Count(frame2Output, " ") == 1 {
		fmt.Println("\n❌ BUG: Output has only 1 space")
	}
}
