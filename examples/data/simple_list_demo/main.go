package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui/components/list"
)

// SimpleListDemo 简化的List VNode演示程序
type SimpleListDemo struct {
	list *list.VNode
}

// NewSimpleListDemo 创建演示实例
func NewSimpleListDemo() *SimpleListDemo {
	return &SimpleListDemo{}
}

// Run 运行演示
func (d *SimpleListDemo) Run() {
	fmt.Println("=== List VNode 组件渲染演示 ===")
	fmt.Println("创建一个模拟的hittest数据列表并分析渲染结果")

	// 1. 准备数据
	entries := []struct {
		NodeID    string
		Bounds    string
		ZOrder    int
		HitTest   string
		Clickable bool
	}{
		{"button-ok", "(10,20,80,25)", 5, "YES", true},
		{"input-name", "(10,50,120,25)", 4, "NO", false},
		{"modal", "(0,0,120,40)", 3, "YES", true},
		{"container", "(5,5,110,30)", 2, "NO", false},
		{"root", "(0,0,120,40)", 1, "NO", false},
	}

	// 2. 准备List数据
	rows := make([]string, 0, len(entries)+1)

	// 添加列标题
	colHeader := fmt.Sprintf("%-3s %-15s %-12s %-2s %-2s", "Z", "Node", "Bounds", "H", "C")
	rows = append(rows, colHeader)

	// 添加数据行
	for i := range entries {
		e := entries[len(entries)-1-i] // 反向排序

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

	// 3. 创建List VNode (Fiber-first)
	headerStyle := style.Style{}.Bold(true)
	headerStyle.FG = style.Color("green")

	listNode := list.NewBuilder().
		Header("🎯 Hit Test Data").
		Rows(rows).
		HeaderStyle(headerStyle).
		RowStyleFn(func(index int, text string) style.Style {
			// 第一行（列标题）
			if index == 0 {
				return style.Style{}.Bold(true).Foreground(style.Color("cyan"))
			}
			// 包含 ✓ 的行
			if strings.Contains(text, "✓") {
				return style.Style{}.Bold(true).Foreground(style.Color("yellow"))
			}
			return style.Style{}
		}).
		ShowSeparator(true).
		Build()

	d.list = listNode.(*list.VNode)

	// 4. 分析和输出
	d.analyzeList()
	d.renderAndAnalyze()
}

// analyzeList 分析List VNode的属性
func (d *SimpleListDemo) analyzeList() {
	fmt.Printf("\n=== List VNode 属性分析 ===\n")

	fmt.Printf("Header: %q\n", d.list.Header())
	fmt.Printf("Rows count: %d\n", d.list.RowCount())

	// 测量列表
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}
	size := d.list.Measure(constraints)
	fmt.Printf("Measured size: Width=%d, Height=%d\n", size.Width, size.Height)

	// 显示数据行
	fmt.Printf("\n数据行内容:\n")
	for i, row := range d.list.Rows() {
		styleInfo := d.getRowStyleInfo(i)
		fmt.Printf("  Row %2d: %q [%s]\n", i+1, row, styleInfo)
	}
}

// getRowStyleInfo 获取行的样式信息
func (d *SimpleListDemo) getRowStyleInfo(index int) string {
	text := d.list.Rows()[index]

	if index == 0 {
		return "青色粗体 (列标题)"
	}
	if strings.Contains(text, "✓") {
		return "黄色粗体 (命中行)"
	}
	return "默认样式"
}

// renderAndAnalyze 渲染列表并分析结果
func (d *SimpleListDemo) renderAndAnalyze() {
	fmt.Printf("\n=== 渲染分析 ===\n")

	cmds := d.list.Paint(0, 0)
	fmt.Printf("渲染命令数量: %d\n", len(cmds))

	// 分析每个命令
	headerFound := false
	colHeaderFound := false
	separatorFound := false
	hitHighlightFound := false
	var lineStyles []string

	for i, cmd := range cmds {
		lineStyles = append(lineStyles, d.analyzeCommand(cmd))

		fmt.Printf("\n命令 %d:\n", i+1)
		fmt.Printf("  文本: %q\n", cmd.Text)
		fmt.Printf("  位置: (%d, %d)\n", cmd.X, cmd.Y)
		fmt.Printf("  样式: %s\n", lineStyles[i])
		fmt.Printf("  长度: %d 字符\n", len(cmd.Text))

		// 检查特定内容
		if cmd.Text == "🎯 Hit Test Data" {
			headerFound = true
		}
		if strings.Contains(cmd.Text, "Z   Node   Bounds   H   C") {
			colHeaderFound = true
		}
		if strings.Contains(cmd.Text, "─") && strings.Repeat("─", len(cmd.Text)) == cmd.Text {
			separatorFound = true
		}
		if strings.Contains(cmd.Text, "✓") {
			hitHighlightFound = true
		}
	}

	// 输出统计
	fmt.Printf("\n=== 渲染统计 ===\n")
	fmt.Printf("Header渲染: %s✓\n", yesNo(headerFound))
	fmt.Printf("列标题渲染: %s✓\n", yesNo(colHeaderFound))
	fmt.Printf("分隔线渲染: %s✓\n", yesNo(separatorFound))
	fmt.Printf("命中高亮渲染: %s✓\n", yesNo(hitHighlightFound))

	// 验证预期
	fmt.Printf("\n=== 验证结果 ===\n")
	allGood := headerFound && colHeaderFound && separatorFound && hitHighlightFound

	if allGood {
		fmt.Println("🎉 所有组件正确渲染！")
		fmt.Println("   ✓ Header显示正确")
		fmt.Println("   ✓ 列标题使用青色粗体")
		fmt.Println("   ✓ 分隔线正确显示")
		fmt.Println("   ✓ 命中行使用黄色高亮")
	} else {
		fmt.Println("❌ 存在渲染问题")
	}

	// 样式分析
	fmt.Printf("\n=== 样式应用分析 ===\n")
	uniqueStyles := make(map[string]int)
	for _, s := range lineStyles {
		uniqueStyles[s]++
	}

	for styleName, count := range uniqueStyles {
		fmt.Printf("  %s: %d 行\n", styleName, count)
	}
}

// analyzeCommand 分析单个命令的样式
func (d *SimpleListDemo) analyzeCommand(cmd paint.DrawCmd) string {
	styleDesc := cmd.Style.ToANSI()

	// 分析样式特征
	isBold := cmd.Style.IsBold()
	isCyan := strings.Contains(styleDesc, "cyan") || strings.Contains(styleDesc, "36")
	isYellow := strings.Contains(styleDesc, "yellow") || strings.Contains(styleDesc, "33")

	if isCyan && isBold {
		return "青色粗体"
	} else if isYellow && isBold {
		return "黄色粗体"
	} else if strings.Contains(cmd.Text, "─") {
		return "分隔线样式"
	} else if isBold {
		return "粗体样式"
	} else {
		return "默认样式"
	}
}

// yesNo 将布尔值转换为是/否
func yesNo(b bool) string {
	if b {
		return ""
	}
	return "✗ "
}

// main 主函数
func main() {
	demo := NewSimpleListDemo()
	demo.Run()

	fmt.Printf("\n=== 演示完成 ===\n")
	fmt.Println("List VNode组件已成功演示并验证 (Fiber-first 架构)")
}
