package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func main() {
	fmt.Println("=== 正确模拟真实渲染流程 ===\n")

	renderer := paint.NewRenderer(40, 5)

	// === Frame 1 ===
	fmt.Println("=== Frame 1 ===")
	// 1. Get Back Buffer
	back1 := renderer.GetBackBuffer()

	// 2. Reset (app.go line 1407)
	back1.Reset(40, 5)

	// 3. Paint
	back1.SetString(5, 3, "Button →", style.Style{})

	fmt.Println("After Paint(\"Button →\"):")
	for x := 11; x < 14; x++ {
		cell := back1.GetContent(x, 3)
		fmt.Printf("  back1[%d]: %q w=%d\n", x, cell.Cluster, cell.Width)
	}

	// 4. Render (includes swapBuffers)
	renderer.Render()
	front1 := renderer.GetFrontBuffer()
	fmt.Println("After Render + swap:")
	for x := 11; x < 14; x++ {
		cell := front1.GetContent(x, 3)
		fmt.Printf("  front1[%d]: %q w=%d\n", x, cell.Cluster, cell.Width)
	}

	// === Frame 2 ===
	fmt.Println("\n=== Frame 2 ===")
	// 1. Get Back Buffer
	back2 := renderer.GetBackBuffer()

	fmt.Println("Before Reset (swap后的状态):")
	for x := 11; x < 14; x++ {
		cell := back2.GetContent(x, 3)
		fmt.Printf("  back2[%d]: %q w=%d\n", x, cell.Cluster, cell.Width)
	}

	// 2. Reset
	back2.Reset(40, 5)

	fmt.Println("After Reset:")
	for x := 11; x < 14; x++ {
		cell := back2.GetContent(x, 3)
		fmt.Printf("  back2[%d]: %q w=%d\n", x, cell.Cluster, cell.Width)
	}

	// 3. Paint
	back2.SetString(5, 3, "Button-", style.Style{})

	fmt.Println("After Paint(\"Button-\"):")
	for x := 11; x < 14; x++ {
		cell := back2.GetContent(x, 3)
		fmt.Printf("  back2[%d]: %q w=%d\n", x, cell.Cluster, cell.Width)
	}

	// === 检查diff ===
	fmt.Println("\n=== Diff Check ===")
	cell12 := back2.GetContent(12, 3)
	prevCell12 := front1.GetContent(12, 3)

	fmt.Printf("Position 12:\n")
	fmt.Printf("  current (back):  %q w=%d\n", cell12.Cluster, cell12.Width)
	fmt.Printf("  previous (front): %q w=%d\n", prevCell12.Cluster, prevCell12.Width)

	// 手动实现IsCellChanged逻辑
	changed := false
	if cell12.IsContinuation {
		changed = false
	} else if prevCell12.IsContinuation {
		changed = true
	} else if prevCell12.Width == 2 && cell12.Width == 0 {
		changed = true
	} else if cell12.Cluster != prevCell12.Cluster {
		changed = true
		fmt.Println("  ✓ 变化检测: Cluster不同")
	} else if cell12.Style != prevCell12.Style {
		changed = true
	} else if cell12.Selected != prevCell12.Selected {
		changed = true
	}

	fmt.Printf("  IsCellChanged = %v\n\n", changed)

	if changed {
		fmt.Println("问题:")
		fmt.Println("  - IsCellChanged检测到变化 ✓")
		fmt.Printf("  - 当前cell是 %q, 宽度为%d\n", cell12.Cluster, cell12.Width)
		fmt.Println("  - renderLine会输出: ")
		fmt.Printf("    emitRunWithWidth(12, 3, style, %q, %d)\n", cell12.Cluster, cell12.Width)
		fmt.Println("  - 这会输出1个空格，但需要清除2个宽度的区域！")
	} else {
		fmt.Println("问题:")
		fmt.Println("  - IsCellChanged没有检测到变化 ✗")
		fmt.Println("  - 因为:")
		fmt.Printf("    prevCell.Width == 2 && cell.Width == 0 → %v\n", prevCell12.Width == 2 && cell12.Width == 0)
		fmt.Println("    但cell.Width是1（Reset设置的），不是0")
		fmt.Println("  - 结果: renderLine跳过，不输出清除指令")
	}
}
