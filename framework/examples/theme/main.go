package main

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/display"
	"github.com/wwsheng009/mint/framework/input"
	"github.com/wwsheng009/mint/runtime/style"
)

// main 主题演示 - 展示不同主题的样式效果
func main() {
	fmt.Println("==============================================")
	fmt.Println("   TUI Framework - 主题演示")
	fmt.Println("==============================================")
	fmt.Println()

	// 可用主题列表
	themes := []string{
		"light", "dark", "dracula", "nord", "monokai", "tokyo-night",
	}

	fmt.Println("可用主题:")
	for i, theme := range themes {
		fmt.Printf("  %d. %s\n", i+1, theme)
	}
	fmt.Println()

	// 演示不同颜色
	fmt.Println("--- 颜色演示 ---")
	demoColors()

	// 演示文本样式
	fmt.Println("\n--- 文本样式演示 ---")
	demoTextStyles()

	// 演示输入框样式
	fmt.Println("\n--- 输入框演示 ---")
	demoInputStyles()
}

func demoColors() {
	colors := []struct {
		name  string
		color style.Color
	}{
		{"黑色", style.Black},
		{"红色", style.Red},
		{"绿色", style.Green},
		{"黄色", style.Yellow},
		{"蓝色", style.Blue},
		{"品红", style.Magenta},
		{"青色", style.Cyan},
		{"白色", style.White},
	}

	fmt.Println("前景色:")
	for _, c := range colors {
		text := display.NewText(c.name)
		text.SetStyle(style.Style{}.Foreground(c.color))
		fmt.Printf("  %s\n", text.GetContent())
	}

	fmt.Println("\n亮色前景色:")
	brightColors := []struct {
		name  string
		color style.Color
	}{
		{"亮黑", style.BrightBlack},
		{"亮红", style.BrightRed},
		{"亮绿", style.BrightGreen},
		{"亮黄", style.BrightYellow},
		{"亮蓝", style.BrightBlue},
		{"亮品红", style.BrightMagenta},
		{"亮青", style.BrightCyan},
		{"亮白", style.BrightWhite},
	}
	for _, c := range brightColors {
		text := display.NewText(c.name)
		text.SetStyle(style.Style{}.Foreground(c.color))
		fmt.Printf("  %s\n", text.GetContent())
	}
}

func demoTextStyles() {
	// 粗体
	boldText := display.NewText("这是粗体文本")
	boldText.SetStyle(style.Style{}.Bold(true))
	fmt.Printf("  粗体: %s\n", boldText.GetContent())

	// 斜体 (终端支持有限)
	italicText := display.NewText("这是斜体文本")
	italicText.SetStyle(style.Style{}.Italic(true))
	fmt.Printf("  斜体: %s\n", italicText.GetContent())

	// 下划线
	underlineText := display.NewText("这是下划线文本")
	underlineText.SetStyle(style.Style{}.Underline(true))
	fmt.Printf("  下划线: %s\n", underlineText.GetContent())

	// 反色
	reverseText := display.NewText("这是反色文本")
	reverseText.SetStyle(style.Style{}.Reverse(true))
	fmt.Printf("  反色: %s\n", reverseText.GetContent())

	// 组合样式
	comboText := display.NewText("这是粗体蓝色下划线文本")
	comboText.SetStyle(style.Style{}.
		Foreground(style.Blue).
		Bold(true).
		Underline(true))
	fmt.Printf("  组合: %s\n", comboText.GetContent())
}

func demoInputStyles() {
	// 普通输入框
	normalInput := input.NewTextInput()
	normalInput.SetPlaceholder("请输入内容...")
	normalInput.SetValue("普通文本")
	fmt.Printf("  普通: %s\n", normalInput.GetValue())

	// 密码输入框
	passwordInput := input.NewTextInput()
	passwordInput.SetPassword(true)
	passwordInput.SetPlaceholder("请输入密码...")
	passwordInput.SetValue("secret123")
	fmt.Printf("  密码: %s\n", maskString(passwordInput.GetValue()))

	// 带样式的输入框
	styledInput := input.NewTextInput()
	styledInput.SetValue("带样式的输入")
	fmt.Printf("  带样式: %s\n", styledInput.GetValue())
}

func maskString(s string) string {
	if len(s) <= 2 {
		return "***"
	}
	return s[:1] + "***" + s[len(s)-1:]
}
