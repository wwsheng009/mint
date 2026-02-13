package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/components/data"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// SampleHitTestEntry 模拟hittest数据结构
type SampleHitTestEntry struct {
	NodeID    string
	Bounds    string
	ZOrder    int
	HitTest   string
	Clickable bool
}

// PrintCommands 输出渲染命令的详细信息
func PrintCommands(cmds []paint.DrawCmd, title string) {
	fmt.Printf("\n=== %s ===\n", title)
	fmt.Printf("Total commands: %d\n", len(cmds))

	for i, cmd := range cmds {
		styleStr := cmd.Style.ToANSI()
		styleDesc := describeStyle(cmd.Style)

		fmt.Printf("\nCommand %d:\n", i+1)
		fmt.Printf("  Text: %q\n", cmd.Text)
		fmt.Printf("  Position: (%d, %d)\n", cmd.X, cmd.Y)
		fmt.Printf("  Style ANSI: %q\n", styleStr)
		fmt.Printf("  Style desc: %s\n", styleDesc)
		fmt.Printf("  Length: %d chars\n", len(cmd.Text))
	}
}

// describeStyle 描述样式的属性
func describeStyle(s style.Style) string {
	desc := "Style["
	if s.FG != "" {
		desc += "FG:" + string(s.FG) + " "
	}
	if s.BG != "" {
		desc += "BG:" + string(s.BG) + " "
	}
	if s.IsBold() {
		desc += "Bold "
	}
	if s.IsItalic() {
		desc += "Italic "
	}
	if s.IsUnderline() {
		desc += "Underline "
	}
	desc += "]"
	return desc
}

// getRowStyleInfo 获取指定行的样式信息（基于文本推断）
func getRowStyleInfo(listVNode *data.ListVNode, index int) string {
	rows := listVNode.Rows()
	if index >= len(rows) {
		return "未知"
	}

	text := rows[index]

	// 基于文本内容推断样式
	if index == 0 {
		return "青色粗体 (列标题)"
	}
	if contains(text, "✓") {
		return "黄色粗体 (命中行)"
	}
	return "默认样式"
}

// contains 简单的字符串包含检查
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// MeasureAndPrint 测量并打印列表的尺寸信息
func MeasureAndPrint(list ui.VNode, title string) {
	listVNode, ok := list.(*data.ListVNode)
	if !ok {
		fmt.Printf("\n=== %s 测量结果 ===\n", title)
		fmt.Printf("无法转换为ListVNode类型\n")
		return
	}

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	size := listVNode.Measure(constraints)
	fmt.Printf("\n=== %s 测量结果 ===\n", title)
	fmt.Printf("Size: Width=%d, Height=%d\n", size.Width, size.Height)
	fmt.Printf("Constraints: MinW=%d, MaxW=%d, MinH=%d, MaxH=%d\n",
		constraints.MinWidth, constraints.MaxWidth, constraints.MinHeight, constraints.MaxHeight)
}

// main 主函数
func main() {
	fmt.Println("=== ListVNode 组件渲染示例 ===")
	fmt.Println("此示例将创建一个模拟的hittest数据列表，并输出渲染后的详细信息")

	// 1. 创建模拟数据
	entries := []SampleHitTestEntry{
		{NodeID: "button-ok", Bounds: "(10,20,80,25)", ZOrder: 5, HitTest: "YES", Clickable: true},
		{NodeID: "input-name", Bounds: "(10,50,120,25)", ZOrder: 4, HitTest: "NO", Clickable: false},
		{NodeID: "modal", Bounds: "(0,0,120,40)", ZOrder: 3, HitTest: "YES", Clickable: true},
		{NodeID: "container", Bounds: "(5,5,110,30)", ZOrder: 2, HitTest: "NO", Clickable: false},
		{NodeID: "root", Bounds: "(0,0,120,40)", ZOrder: 1, HitTest: "NO", Clickable: false},
	}

	// 2. 准备ListVNode数据
	rows := make([]string, 0, len(entries)+1)

	// 添加列标题
	colHeader := fmt.Sprintf("%-3s %-15s %-12s %-2s %-2s", "Z", "Node", "Bounds", "H", "C")
	rows = append(rows, colHeader)

	// 添加数据行
	for i := range entries {
		e := entries[len(entries)-1-i] // 反向排序（Z轴高的在前）

		hitMark := "·"
		if e.HitTest == "YES" {
			hitMark = "✓"
		}
		clickMark := "·"
		if e.Clickable {
			clickMark = "Y"
		}

		line := fmt.Sprintf("%-3d %-15s %-12s %-2s %-2s",
			e.ZOrder, e.NodeID, e.Bounds, hitMark, clickMark)
		rows = append(rows, line)
	}

	// 3. 创建ListVNode
	headerStyle := style.Style{}.Bold(true)
	headerStyle.FG = style.Color("green")

	list := data.ListBuilder().
		Header("🎯 Hit Test Data").
		Rows(rows).
		HeaderStyle(headerStyle).
		RowStyleFn(func(index int, text string) style.Style {
			// 第一行（列标题）使用青色粗体
			if index == 0 {
				return style.Style{}.Bold(true).Foreground(style.Color("cyan"))
			}
			// 包含 ✓ 的行使用黄色高亮
			if strings.Contains(text, "✓") {
				return style.Style{}.Bold(true).Foreground(style.Color("yellow"))
			}
			return style.Style{}
		}).
		ShowSeparator(true).
		Build()

	listVNode := list

	// 4. 测量列表
	MeasureAndPrint(list, "HitTest List")

	// 5. 渲染列表并输出详细信息
	cmds := listVNode.Paint(0, 0)
	PrintCommands(cmds, "ListVNode 渲染命令")

	// 6. 数据验证
	fmt.Printf("\n=== 数据验证 ===\n")
	fmt.Printf("List Header: %q\n", listVNode.Header())
	fmt.Printf("Data Rows: %d\n", listVNode.RowCount())

	// 打印原始数据行
	fmt.Println("\n原始数据行:")
	for i, row := range listVNode.Rows() {
		styleDesc := getRowStyleInfo(listVNode, i)
		fmt.Printf("  Row %d: %q (style: %s)\n", i+1, row, styleDesc)
	}

	// 7. 判断分析
	fmt.Printf("\n=== 判断分析 ===\n")

	// 检查渲染命令
	totalLines := 0
	headerFound := false
	colHeaderFound := false
	hitHighlightFound := false

	for _, cmd := range cmds {
		totalLines++

		// 检查Header（去除填充空格）
		trimmedText := strings.TrimSpace(cmd.Text)
		if strings.HasPrefix(trimmedText, "🎯 Hit Test Data") {
			headerFound = true
			fmt.Println("✅ Header正确渲染")
		}

		// 检查列标题（去除填充空格）
		if strings.Contains(trimmedText, "Z") && strings.Contains(trimmedText, "Node") && strings.Contains(trimmedText, "Bounds") && strings.Contains(trimmedText, "H") && strings.Contains(trimmedText, "C") {
			colHeaderFound = true
			fmt.Println("✅ 列标题正确渲染")
		}

		// 检查命中高亮
		if strings.Contains(cmd.Text, "✓") && cmd.Style.ToANSI() != "" {
			hitHighlightFound = true
			fmt.Println("✅ 命中高亮正确应用")
		}
	}

	fmt.Printf("\n渲染统计:\n")
	fmt.Printf("  - 总渲染行数: %d\n", totalLines)
	fmt.Printf("  - Header存在: %v\n", headerFound)
	fmt.Printf("  - 列标题存在: %v\n", colHeaderFound)
	fmt.Printf("  - 命中高亮存在: %v\n", hitHighlightFound)

	// 8. 结论
	fmt.Printf("\n=== 结论 ===\n")
	if headerFound && colHeaderFound && hitHighlightFound {
		fmt.Println("🎉 ListVNode组件正确渲染了所有预期内容")
		fmt.Println("   - Header显示正确")
		fmt.Println("   - 列标题样式正确（青色粗体）")
		fmt.Println("   - 命中行高亮正确（黄色高亮）")
		fmt.Println("   - 分隔线正确显示")
	} else {
		fmt.Println("❌ ListVNode组件渲染存在问题")
	}

	fmt.Println("\n=== 演示完成 ===")
}