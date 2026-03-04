// Package main demonstrates focus switching between multiple components
// This demo shows Tab-based keyboard navigation across focusable components
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
	buttonComp "github.com/wwsheng009/mint/ui/components/button"
	checkboxComp "github.com/wwsheng009/mint/ui/components/checkbox"
	inputComp "github.com/wwsheng009/mint/ui/components/input"
)

// ClickButtonIntent 按钮点击 intent
type ClickButtonIntent struct{}

func (ClickButtonIntent) IntentType() string { return "ClickButton" }
func (ClickButtonIntent) StayPressed() bool  { return true }

// FocusApp demonstrates focusable components with MVP pattern
func FocusApp() ui.VNode {
	// 使用 UseState 创建表单状态
	input1Value, setInput1Value := ui.UseStateString("")
	input2Value, setInput2Value := ui.UseStateString("")

	// 按钮按下计数器 - 展示按钮被点击
	clickCount, setClickCount, _ := ui.UseStateInt(0)

	// Checkbox 状态
	checked1, setChecked1 := ui.UseStateBool(false)
	checked2, setChecked2 := ui.UseStateBool(false)

	// ✅ 关键：将 setter 保存到 GlobalState，供 Intent handler 调用
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["input1-value-setter"] = setInput1Value
		ctx.GlobalState["input2-value-setter"] = setInput2Value
		ctx.GlobalState["setClickCount"] = setClickCount
		ctx.GlobalState["checked1-setter"] = setChecked1
		ctx.GlobalState["checked2-setter"] = setChecked2
	}

	return ui.VStack(
		ui.NewTextBuilder("─────────────────────────────────────────").FgColor("cyan").Build(),
		ui.NewTextBuilder("Focus Switching Demo").Bold(true).FgColor("yellow").Build(),
		ui.Text(""),
		ui.NewTextBuilder("Use TAB to navigate between components").FgColor("gray").Build(),
		ui.NewTextBuilder("Press ENTER to activate button/checkbox").FgColor("gray").Build(),
		ui.Text(""),
		ui.NewTextBuilder("─────────────────────────────────────────").FgColor("cyan").Build(),
		ui.Text(""),

		// Button 1
		buttonComp.NewBuilder("Button 1 - First").
			OnPress(ClickButtonIntent{}).
			Build().
			SetKey("btn1"),

		// Input 1 - 使用 UseState
		InputBorder("Input 1:", "input1-value", "input1-value-setter", "Enter name...", 25, input1Value),

		// Checkbox 1
		checkboxComp.NewBuilder().
			Label("Option A").
			ForField(intent.BindField("checked1")).
			Checked(checked1).
			Build().
			SetKey("chk1"),

		// Button 2
		buttonComp.NewBuilder("Button 2 - Middle").
			OnPress(ClickButtonIntent{}).
			Build().
			SetKey("btn2"),

		// Input 2 - 使用 UseState
		InputBorder("Input 2:", "input2-value", "input2-value-setter", "Enter email...", 25, input2Value),

		// Checkbox 2
		checkboxComp.NewBuilder().
			Label("Option B").
			ForField(intent.BindField("checked2")).
			Checked(checked2).
			Build().
			SetKey("chk2"),

		// Button 3 - Disabled
		buttonComp.NewBuilder("Button 3 - Last").
			Disabled(true).
			OnPress(ClickButtonIntent{}).
			Build().
			SetKey("btn3"),

		// Disabled Input 3
		ui.VStack(
			ui.NewTextBuilder("Input 3 (disabled):").FgColor("blue").Build(),
			inputComp.NewBuilder().
				ForField(intent.BindField("input3-value")).
				Placeholder("Disabled Input").
				Disabled(true).
				Build().
				SetKey("input3"),
		),

		// Disabled Checkbox
		checkboxComp.NewBuilder().
			Label("Disabled Checkbox").
			Disabled(true).
			OnToggle(intent.Toggle("chk3-checked")).
			Build().
			SetKey("chk3"),

		ui.Text(""),
		ui.NewTextBuilder("─────────────────────────────────────────").FgColor("cyan").Build(),

		// 显示按钮点击计数
		ui.NewTextBuilder(fmt.Sprintf("Button Click Count: %d", clickCount)).
			FgColor("green").
			Bold(true).
			Build(),

		// 显示 checkbox 状态
		ui.NewTextBuilder(fmt.Sprintf("Checkbox A: %t  |  Checkbox B: %t", checked1, checked2)).
			FgColor("yellow").
			Build(),
	)
}

// InputBorder 创建带显示值的输入框组件
// fieldKey: 字段键（用于 ForField）
// setterKey: setter 的键名（从 GlobalState 获取）
// value: 当前值（从 UseState 获取）
func InputBorder(label string, fieldKey string, setterKey string, placeholder string, width int, value string) ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder(label).FgColor("blue").Build(),
		inputComp.NewBuilder().
			ForField(intent.BindField(fieldKey)).
			Value(value).
			Placeholder(placeholder).
			Width(width).
			Build().
			SetKey(fieldKey),
		// 显示当前输入值（调试）
		ui.NewTextBuilder(fmt.Sprintf("Value: %s", value)).
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Focus Switching Demo - Interactive Keyboard Navigation  ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println("")
	fmt.Println("This demo showcases focus management with multiple component types:")
	fmt.Println("  - Buttons (3 focusable)")
	fmt.Println("  - Input fields (2 focusable)")
	fmt.Println("  - Checkboxes (2 focusable)")
	fmt.Println("  - Disabled components (3, skipped during navigation)")
	fmt.Println("")
	fmt.Println("Intent Usage Examples (MVP Pattern):")
	fmt.Println("  Button:   OnPress(intent.Click(\"btn1\"))")
	fmt.Println("  Input:    ForField(intent.BindField(\"key\")) + Value(state)")
	fmt.Println("  Checkbox: OnToggle(intent.Toggle(\"key\"))")
	fmt.Println("")
	fmt.Println("Using ui.Run() to start the application...")
	fmt.Println("")
	fmt.Println("Press TAB to move focus between components.")
	fmt.Println("Press ESC or CTRL+C to exit.")
	fmt.Println("")

	err := ui.Run(FocusApp,
		ui.WithWidth(60),
		ui.WithHeight(35),
		ui.WithTitle("Focus Switching Demo"),
		ui.WithInit(func() {
			// 1. 注册 FieldChangeIntent handler 将输入值同步到 UseState
			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				switch i.Field {
				case "input1-value", "input2-value":
					if fn, ok := ctx.GetState(i.Field+"-setter"); ok {
						if setter, ok := fn.(func(string)); ok {
							setter(i.Value)
						}
					}
				case "checked1", "checked2":
					if fn, ok := ctx.GetState(i.Field+"-setter"); ok {
						value := i.Value == "true"
						if setter, ok := fn.(func(bool)); ok {
							setter(value)
						}
					}
				}
				return intent.HandledResult()
			})

			// 2. 注册 ClickButtonIntent handler 更新按钮计数
			ui.RegisterIntent(func(ctx *intent.ActionContext, i ClickButtonIntent) intent.IntentResult {
				if fn, ok := ctx.GetState("setClickCount"); ok {
					// UseStateInt 的setter类型是 func(interface{})
					if setter, ok := fn.(func(interface{})); ok {
						// 使用 functional update：基于当前值递增
						setter(func(c int) int {
							return c + 1
						})
					}
				}
				return intent.HandledResult()
			})
		}),
	)
	if err != nil {
		fmt.Printf("Error running app: %v\n", err)
	}
}
