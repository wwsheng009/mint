package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/components/basic"
	"github.com/wwsheng009/mint/components/button"
	"github.com/wwsheng009/mint/components/form"
	"github.com/wwsheng009/mint/components/layout"
	"github.com/wwsheng009/mint/runtime/style"
	ui "github.com/wwsheng009/mint/ui"
)

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("TUI_DEBUG_UI", "false")

	ui.Run(RootApp,
		ui.WithSize(100, 40),
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
)

// RootApp 根组件
func RootApp() ui.VNode {
	return layout.VStack(
		createText("=== Mint UI Complex Test ===", style.Color("cyan"), true),
		basic.Divider(),
		basic.Text(""),

		// 标签页导航
		layout.HStack(
			tabButton("Form", "form"),
			basic.Text(" "),
			tabButton("List", "list"),
			basic.Text(" "),
			tabButton("Modal", "modal"),
			basic.Text(" "),
			tabButton("Progress", "progress"),
		),

		basic.Text(""),
		basic.Divider(),
		basic.Text(""),

		// 标签页内容
		renderTabContent(),
	)
}

// createText 创建带样式的文本
func createText(content string, color style.Color, bold bool) ui.VNode {
	return createStyledText(content, color, bold)
}

// createStyledText 创建带样式的文本
func createStyledText(content string, color style.Color, bold bool) ui.VNode {
	s := style.Style{}.Foreground(color)
	if bold {
		s = s.Bold(true)
	}
	t := basic.NewText(content)
	t.SetStyle(s)
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
	return button.ButtonBuilder(label).
		OnClick(func() {
			currentTab = tabID
		}).
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
		return basic.Text("Unknown tab")
	}
}

// renderFormTab 表单标签
func renderFormTab() ui.VNode {
	return layout.VStack(
		createText("--- Form Tab ---", style.Color("yellow"), true),
		basic.Text(""),

		layout.HStack(
			basic.Text("Name:  "),
			basic.Text(fmt.Sprintf("[%30s]", formName)),
		),

		basic.Text(""),
		layout.HStack(
			basic.Text("Email: "),
			basic.Text(fmt.Sprintf("[%30s]", formEmail)),
		),

		basic.Text(""),
		layout.HStack(
			form.CheckboxBuilder().
				Label("I agree to terms").
				Checked(formAgree).
				OnChange(func(bool) {
					formAgree = !formAgree
				}).
				Build(),
		),

		basic.Text(""),
		basic.Text(""),
		layout.HStack(
			button.ButtonBuilder("Submit").
				Variant(button.ButtonVariantPrimary).
				Disabled(!formAgree).
				OnClick(func() {
					fmt.Printf("Form: Name=%s, Email=%s, Agree=%v\n",
						formName, formEmail, formAgree)
				}).
				Build(),

			basic.Text("  "),

			button.ButtonBuilder("Reset").
				Variant(button.ButtonVariantSecondary).
				OnClick(func() {
					formName = ""
					formEmail = ""
					formAgree = false
				}).
				Build(),
		),

		basic.Text(""),
		createText(fmt.Sprintf("Name: %s | Email: %s | Agree: %v",
			formName, formEmail, formAgree), style.Color("gray"), false),
	)
}

// renderListTab 列表标签
func renderListTab() ui.VNode {
	itemCount := len(listItems)

	children := []ui.VNode{
		createText("--- List Tab ---", style.Color("yellow"), true),
		basic.Text(""),
		layout.HStack(
			button.ButtonBuilder("Add Item").
				Variant(button.ButtonVariantSuccess).
				OnClick(func() {
					listItems = append(listItems, fmt.Sprintf("Item %d", itemCount+1))
				}).
				Build(),
			basic.Text("  "),
			button.ButtonBuilder("Remove Last").
				Variant(button.ButtonVariantDanger).
				OnClick(func() {
					if len(listItems) > 0 {
						listItems = listItems[:len(listItems)-1]
					}
				}).
				Build(),
			basic.Text("  "),
			button.ButtonBuilder("Clear All").
				Variant(button.ButtonVariantSecondary).
				OnClick(func() {
					listItems = []string{}
				}).
				Build(),
		),
		basic.Text(""),
		createStyledText(fmt.Sprintf("Items: %d", itemCount), style.Color("white"), true),
		basic.Text(""),
	}

	for i, item := range listItems {
		children = append(children, layout.HStack(
			basic.Text(fmt.Sprintf("%2d. ", i+1)),
			basic.Text(item),
		))
	}

	if len(listItems) == 0 {
		children = append(children, createStyledText("(empty)", style.Color("gray"), false))
	}

	return layout.VStack(children...)
}

// renderModalTab 模态框标签
func renderModalTab() ui.VNode {
	return layout.VStack(
		createText("--- Modal Tab ---", style.Color("yellow"), true),
		basic.Text(""),

		button.ButtonBuilder("Show Modal").
			Variant(button.ButtonVariantPrimary).
			OnClick(func() {
				showModal = true
			}).
			Build(),

		basic.Text(""),
		createStyledText(fmt.Sprintf("Modal Visible: %v", showModal), style.Color("gray"), false),
	)
}

// renderProgressTab 进度条标签
func renderProgressTab() ui.VNode {
	return layout.VStack(
		createText("--- Progress Tab ---", style.Color("yellow"), true),
		basic.Text(""),

		app.ProgressBuilder().
			Value(progressValue).
			Max(100).
			Build(),

		basic.Text(""),
		basic.Text(fmt.Sprintf("Progress: %d%%", progressValue)),
		basic.Text(""),

		layout.HStack(
			button.ButtonBuilder("0%").
				OnClick(func() { progressValue = 0 }).
				Size(button.ButtonSizeSmall).
				Build(),
			basic.Text(" "),
			button.ButtonBuilder("25%").
				OnClick(func() { progressValue = 25 }).
				Size(button.ButtonSizeSmall).
				Build(),
			basic.Text(" "),
			button.ButtonBuilder("50%").
				OnClick(func() { progressValue = 50 }).
				Size(button.ButtonSizeSmall).
				Build(),
			basic.Text(" "),
			button.ButtonBuilder("75%").
				OnClick(func() { progressValue = 75 }).
				Size(button.ButtonSizeSmall).
				Build(),
			basic.Text(" "),
			button.ButtonBuilder("100%").
				OnClick(func() { progressValue = 100 }).
				Size(button.ButtonSizeSmall).
				Build(),
		),

		basic.Text(""),
		basic.Text(""),

		button.ButtonBuilder("Auto Increment").
			Variant(button.ButtonVariantSuccess).
			OnClick(func() {
				progressValue += 10
				if progressValue > 100 {
					progressValue = 0
				}
			}).
			Build(),
	)
}
