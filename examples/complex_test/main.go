package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/runtime/style"
	ui "github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/divider"
	"github.com/wwsheng009/mint/ui/components/progress"
	"github.com/wwsheng009/mint/ui/components/text"
)

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("TUI_DEBUG_UI", "false")

	ui.Run(RootApp,
		ui.WithSize(100, 42),
		ui.WithTitle("Mint UI - Complex Test"),
	)
}

// Global state
var (
	currentTab    = "form"
	formName      = ""
	formEmail     = ""
	formAgree     = false
	listItems     = []string{"Item 1", "Item 2", "Item 3"}
	showModal     = false
	progressValue = 0

	// Key event tracking for debugging
	lastKeyName  = ""
	lastKeyValue = 0
	isSpecialKey = false
	keyModifiers = ""
)

// RootApp 根组件
func RootApp() ui.VNode {
	return ui.VStack(
		createText("=== Mint UI Complex Test ===", style.Color("cyan"), true),
		divider.H(),
		ui.Text(""),

		// 标签页导航
		ui.HStack(
			tabButton("Form", "form"),
			ui.Text(" "),
			tabButton("List", "list"),
			ui.Text(" "),
			tabButton("Modal", "modal"),
			ui.Text(" "),
			tabButton("Progress", "progress"),
		),

		ui.Text(""),
		divider.H(),

		// Key event info display (for debugging)
		renderKeyInfo(),

		ui.Text(""),

		// 标签页内容
		renderTabContent(),
	)
}

// renderKeyInfo 渲染按键信息
func renderKeyInfo() ui.VNode {
	// Build key info string
	var infoParts []string
	if lastKeyName == "" {
		infoParts = append(infoParts, "No key pressed yet")
	} else {
		infoParts = append(infoParts, fmt.Sprintf("Key: %s", lastKeyName))
		if isSpecialKey {
			infoParts = append(infoParts, fmt.Sprintf("Value: %d", lastKeyValue))
			infoParts = append(infoParts, "Type: Special")
		} else {
			infoParts = append(infoParts, fmt.Sprintf("Rune: %c", lastKeyValue))
			infoParts = append(infoParts, "Type: Regular")
		}
		if keyModifiers != "" {
			infoParts = append(infoParts, fmt.Sprintf("Modifiers: %s", keyModifiers))
		}
	}

	// Tab key info (hardcoded reference)
	tabInfo := "Tab key value: 3 (KeyTab=3, KeyUnknown=0, KeyEscape=1, KeyEnter=2)"

	infoText := strings.Join(infoParts, " | ")

	return ui.VStack(
		text.New(infoText).Foreground(style.Color("yellow")),
		text.New(tabInfo).Foreground(style.Color("gray")),
	)
}

// createText 创建带样式的文本
func createText(content string, color style.Color, bold bool) ui.VNode {
	return createStyledText(content, color, bold)
}

// createStyledText 创建带样式的文本
func createStyledText(content string, color style.Color, bold bool) ui.VNode {
	t := text.New(content).Foreground(color)
	if bold {
		t = t.Bold(true)
	}
	return t
}

// tabButton 标签按钮
func tabButton(label, tabID string) ui.VNode {
	var fg style.Color
	if currentTab == tabID {
		fg = style.Color("green")
	} else {
		fg = style.Color("white")
	}
	return ui.NewButtonBuilder(label).
		FgColor(fg).
		Build()
}

// renderTabContent 渲染标签内容
func renderTabContent() ui.VNode {
	switch currentTab {
	case "form":
		return renderFormTab()
	case "list":
		return renderListTab()
	case "modal":
		return renderModalTab()
	case "progress":
		return renderProgressTab()
	default:
		return ui.Text("Unknown tab")
	}
}

// renderFormTab 表单标签
func renderFormTab() ui.VNode {
	return ui.VStack(
		createText("--- Form Tab ---", style.Color("yellow"), true),
		ui.Text(""),

		ui.HStack(
			ui.Text("Name:  "),
			ui.Text(fmt.Sprintf("[%30s]", formName)),
		),

		ui.Text(""),
		ui.HStack(
			ui.Text("Email: "),
			ui.Text(fmt.Sprintf("[%30s]", formEmail)),
		),

		ui.Text(""),
		ui.HStack(
			checkbox.New("I agree to terms").
				Checked(formAgree).
				Build(),
		),

		ui.Text(""),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder("Submit").
				Variant(ui.NewButtonVariantPrimary).
				Disabled(!formAgree).
				Build(),

			ui.Text("  "),

			ui.NewButtonBuilder("Reset").
				Variant(ui.NewButtonVariantSecondary).
				Build(),
		),

		ui.Text(""),
		createText(fmt.Sprintf("Name: %s | Email: %s | Agree: %v",
			formName, formEmail, formAgree), style.Color("gray"), false),
	)
}

// renderListTab 列表标签
func renderListTab() ui.VNode {
	itemCount := len(listItems)

	children := []ui.VNode{
		createText("--- List Tab ---", style.Color("yellow"), true),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder("Add Item").
				Variant(ui.ButtonVariantSuccess).
				Build(),
			ui.Text("  "),
			ui.NewButtonBuilder("Remove Last").
				Variant(ui.ButtonVariantDanger).
				Build(),
			ui.Text("  "),
			ui.NewButtonBuilder("Clear All").
				Variant(ui.ButtonVariantSecondary).
				Build(),
		),
		ui.Text(""),
		createStyledText(fmt.Sprintf("Items: %d", itemCount), style.Color("white"), true),
		ui.Text(""),
	}

	for i, item := range listItems {
		children = append(children, ui.HStack(
			ui.Text(fmt.Sprintf("%2d. ", i+1)),
			ui.Text(item),
		))
	}

	if len(listItems) == 0 {
		children = append(children, createStyledText("(empty)", style.Color("gray"), false))
	}

	return ui.VStack(children...)
}

// renderModalTab 模态框标签
func renderModalTab() ui.VNode {
	return ui.VStack(
		createText("--- Modal Tab ---", style.Color("yellow"), true),
		ui.Text(""),

		ui.NewButtonBuilder("Show Modal").
			Variant(ui.ButtonVariantPrimary).
			Build(),

		ui.Text(""),
		createStyledText(fmt.Sprintf("Modal Visible: %v", showModal), style.Color("gray"), false),
	)
}

// renderProgressTab 进度条标签
func renderProgressTab() ui.VNode {
	return ui.VStack(
		createText("--- Progress Tab ---", style.Color("yellow"), true),
		ui.Text(""),

		progress.New().
			Value(progressValue).
			Max(100).
			Build(),

		ui.Text(""),
		ui.Text(fmt.Sprintf("Progress: %d%%", progressValue)),
		ui.Text(""),

		ui.HStack(
			ui.NewButtonBuilder("0%").Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("25%").Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("50%").Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("75%").Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("100%").Build(),
		),

		ui.Text(""),
		ui.Text(""),

		ui.NewButtonBuilder("Auto Increment").
			Variant(ui.ButtonVariantSuccess).
			Build(),
	)
}
