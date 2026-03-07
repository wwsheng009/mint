package paint

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func TestBufferResidueBug(t *testing.T) {
	fmt.Println("=== 测试Buffer残留Bug ===\n")

	renderer := paint.NewRenderer(40, 5)

	// Step 1: Frame 1 - 写入 "Button →"
	back := renderer.GetBackBuffer()
	back.SetString(5, 3, "Button →", style.Style{})

	fmt.Println("Step 1: SetString(5, 3, \"Button →\") 后")
	backCell := back.GetContent(12, 3)
	fmt.Printf("  back[12]: Cluster=%q, Width=%d, IsContinuation=%v\n",
		backCell.Cluster, backCell.Width, backCell.IsContinuation)

	// Render
	renderer.Render()
	front := renderer.GetFrontBuffer()
	fmt.Printf("  render后 front[12]: Cluster=%q, Width=%d\n",
		front.GetContent(12, 3).Cluster, front.GetContent(12, 3).Width)

	// Step 2: Frame 2 - 写入 "Button-"
	back2 := renderer.GetBackBuffer()
	fmt.Println("\nStep 2: SetString(5, 3, \"Button-\") 前")
	for x := 11; x < 14; x++ {
		cell := back2.GetContent(x, 3)
		fmt.Printf("  back2[%d]: Cluster=%q, Width=%d\n", x, cell.Cluster, cell.Width)
	}

	back2.SetString(5, 3, "Button-", style.Style{})

	fmt.Println("SetString后:")
	for x := 11; x < 14; x++ {
		cell := back2.GetContent(x, 3)
		fmt.Printf("  back2[%d]: Cluster=%q, Width=%d\n", x, cell.Cluster, cell.Width)
	}

	// Step 3: 检查IsCellChanged
	fmt.Println("\nStep 3: IsCellChanged 检查")
	cell := back2.GetContent(12, 3)
	prevCell := front.GetContent(12, 3)

	fmt.Printf("  cell: Cluster=%q, Width=%d\n", cell.Cluster, cell.Width)
	fmt.Printf("  prevCell: Cluster=%q, Width=%d\n", prevCell.Cluster, prevCell.Width)

	changed := paint.IsCellChanged(cell, prevCell)
	fmt.Printf("  IsCellChanged() = %v\n", changed)

	// 详细分析检测逻辑
	fmt.Println("\n详细检测逻辑:")
	fmt.Printf("  cell.IsContinuation: %v\n", cell.IsContinuation)
	fmt.Printf("  prevCell.IsContinuation: %v\n", prevCell.IsContinuation)
	fmt.Printf("  prevCell.Width == 2 && cell.Width == 0: %v\n", prevCell.Width == 2 && cell.Width == 0)
	fmt.Printf("  cell.Cluster != prevCell.Cluster: %v\n", cell.Cluster != prevCell.Cluster)

	if !changed {
		fmt.Println("\n❌ BUG: IsCellChanged没有检测到变化！")
	} else {
		fmt.Println("\n✓ IsCellChanged检测到变化")
		fmt.Println("\n问题：如果检测到变化，renderLine会如何处理？")
		fmt.Println("  cell.Cluster是空格\" \"，不是\"\"或\"\\x00\"")
		fmt.Println("  所以会进入正常输出逻辑，输出空格")
		fmt.Println("  但问题是：空格能否正确清除宽字符？")
	}
}
