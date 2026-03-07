package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func main() {
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

	// Step 3: 检查IsCellChanged - 需要使用buffer.IsCellChanged
	fmt.Println("\nStep 3: IsCellChanged 检查")
	cell := back2.GetContent(12, 3)
	prevCell := front.GetContent(12, 3)

	fmt.Printf("  cell: Cluster=%q, Width=%d\n", cell.Cluster, cell.Width)
	fmt.Printf("  prevCell: Cluster=%q, Width=%d\n", prevCell.Cluster, prevCell.Width)

	// 调用buffer包中的IsCellChanged函数
	// changed := paint.IsCellChanged(cell, prevCell)
	// 由于不能直接调用私有包的函数，我们手动检查逻辑
	fmt.Println("\n手动检测逻辑:")
	fmt.Printf("  cell.IsContinuation: %v\n", cell.IsContinuation)
	fmt.Printf("  prevCell.IsContinuation: %v\n", prevCell.IsContinuation)
	fmt.Printf("  prevCell.Width == 2 && cell.Width == 0: %v\n", prevCell.Width == 2 && cell.Width == 0)
	fmt.Printf("  cell.Cluster != prevCell.Cluster: %v\n", cell.Cluster != prevCell.Cluster)

	// 模拟IsCellChanged的判断
	changed := false
	if cell.IsContinuation {
		changed = false
	} else if prevCell.IsContinuation {
		changed = true
	} else if prevCell.Width == 2 && cell.Width == 0 {
		changed = true
	} else if cell.Cluster != prevCell.Cluster {
		changed = true
	}

	fmt.Printf("  预期 IsCellChanged() = %v\n", changed)

	if !changed {
		fmt.Println("\n❌ BUG: IsCellChanged没有检测到变化！")
		fmt.Println("\n原因:")
		fmt.Println("  prevCell.Width == 2 && cell.Width == 0 → FALSE")
		fmt.Println("  因为cell.Width = 1 (Reset设置的), 不是0")
		fmt.Println("  正常比较 cell.Cluster != prevCell.Cluster → TRUE")
	} else {
		fmt.Println("\n✓ IsCellChanged检测到变化")
		fmt.Println("\n但是问题来了：")
		fmt.Println("  cell.Cluster = \" \" (空格，不是\"\"或\"\\x00\")")
		fmt.Println("  renderLine会执行正常输出，输出一个空格")
		fmt.Println("  问题：一个空格能否清除一个宽字符的整个区域？")
	}

	fmt.Println("\n=== 核心问题分析 ===")
	if cell.Cluster == " " && prevCell.Width == 2 {
		fmt.Println("\n问题场景：")
		fmt.Println("  当前: Cluster=\" \" (宽度1)")
		fmt.Println("  前帧: Cluster=\"→\" (宽度2)")
		fmt.Println("  \n检测结果: IsCellChanged = TRUE")
		fmt.Println("  处理方式: renderLine输出一个空格 (宽度1)")
		fmt.Println("  结 果: 只清除了位置12，位置13的continuation残留！")

		fmt.Println("\n解决方案:")
		fmt.Println("  1. 在IsCellChanged中检测 prevCell.Width==2 && (cell.Cluster==\" \" || cell.Width==1)")
		fmt.Println("  2. OR在renderLine中，当检测到prev是宽字符时，输出prevCell.Width个空格")
	}
}
