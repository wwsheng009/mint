package main

import (
	"fmt"
)

func main() {
	lines := []string{
		"✨ Elegant VNode Builder API Demo",
		"────────────────────────────────",
		"",
		"1. Flex buttons (no SetProp needed):",
		"*[ Left ]                        [ Center ]                        [ Right ]",
		"",
		"2. Text with PaddingAll(2):",
		"  Padded Text",
		"",
		"3. Buttons with MarginV(0, 1):",
		"  [ Btn1 ]",
		"  [ Btn2 ]",
		"  [ Btn3 ]",
		"",
		"4. Combined: Padding + Margin + Flex:",
		"  [ Spacious ]                                                                  ",
		"",
		"5. Just like CSS:",
		"  button {",
		"    padding: 2px;",
		"    margin: 1px;",
		"    flex: 1;",
		"  }",
	}

	for _, line := range lines {
		fmt.Println(line)
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("分析:")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Analyze row 1
	fmt.Println("第 1 行 (Flex buttons):")
	fmt.Println("  *[ Left ]                        [ Center ]                        [ Right ]")
	fmt.Println("  ↑                               ↑                                ↑")
	fmt.Println("  focus + text                    居中文字                         右对齐文字")
	fmt.Println()

	// Count characters in first button row
	line1 := "*[ Left ]                        [ Center ]                        [ Right ]"
	fmt.Printf("  总宽度: %d 字符\n", len(line1))
	fmt.Printf("  Button 1 (Left): \"[ Left ]\" + 右边空白 ~20 字符\n")
	fmt.Printf("  Button 2 (Center): ~20 字符 + \"[ Center ]\" + ~20 字符\n")
	fmt.Printf("  Button 3 (Right): ~40 字符 + \"[ Right ]\"\n")
	fmt.Println()

	// Issue identification
	fmt.Println("❌ 发现的问题:")
	fmt.Println("─────────────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("1. **按钮没有等宽分布**")
	fmt.Println("   预期: 每个 button 约 25 字符宽 (76 / 3)")
	fmt.Println("   实际: Button 1 很窄，Button 2 居中，Button 3 很宽")
	fmt.Println()
	fmt.Println("2. **PaddingH 没有生效**")
	fmt.Println("   Button 1: PaddingH(1, 2) 应该左边少、右边多")
	fmt.Println("   但看起来按钮整体就很窄，padding 没有明显效果")
	fmt.Println()
	fmt.Println("3. **TextAlign 没有明显效果**")
	fmt.Println("   Button 的文字应该根据 TextAlign 调整位置")
	fmt.Println("   但由于按钮整体很窄，看不出对齐效果")
	fmt.Println()

	// Root cause
	fmt.Println("🔍 根本原因:")
	fmt.Println("─────────────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("**HStack 没有传递边界约束给子元素**")
	fmt.Println()
	fmt.Println("当 HStack 测量时:")
	fmt.Println("  如果父容器没有边界宽度 → HStack.MaxWidth = Infinity")
	fmt.Println("  → Flex 子元素被测量为自然宽度")
	fmt.Println("  → Button 保持自然宽度 9-11 字符")
	fmt.Println()
	fmt.Println("预期行为:")
	fmt.Println("  HStack 应该从父容器获得约束 (如 MaxWidth = 76)")
	fmt.Println("  → 分配空间: 每个 Button 获得 MinWidth = MaxWidth = 25")
	fmt.Println("  → Button.Measure(25, 25) 返回 Width = 25")
	fmt.Println("  → 按钮拉伸填充分配的空间")
}
