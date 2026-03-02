package main

import (
	"fmt"
	"reflect"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// SelectDemo demonstrates the select dropdown component (MVP 模式)
func SelectDemo() ui.VNode {
	// Select 组件的 FieldChangeIntent 携带的是索引（int）
	selectedIndex, setSelectedIndex, _ := ui.UseStateInt(0)
	selectedIndexKey := intent.StateKey[int]("selectedIndex")
	selectedIndexSetterKey := intent.StateKey[func(int)]("selectedIndexSetter")

	// 保存 setter
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState[selectedIndexSetterKey.String()] = setSelectedIndex
	}

	// 索引到主题值的映射
	indexToTheme := map[int]string{
		0: "dark",
		1: "light",
		2: "dracula",
		3: "nord",
	}

	// 主题值到显示名称的映射
	themeNames := map[string]string{
		"dark":    "Dark Theme",
		"light":   "Light Theme",
		"dracula": "Dracula Theme",
		"nord":    "Nord Theme",
	}

	// 获取当前选中的主题值
	currentThemeValue := indexToTheme[selectedIndex]
	currentThemeName := themeNames[currentThemeValue]

	return ui.VStack(
		ui.NewTextBuilder("Settings Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("─────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Theme:").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		app.SelectBuilder().
			ForField(intent.ForField[int](selectedIndexKey)).
			AddOption("dark", "Dark Theme").
			AddOption("light", "Light Theme").
			AddOption("dracula", "Dracula Theme").
			AddOption("nord", "Nord Theme").
			Selected(selectedIndex).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Selected: %s", currentThemeName)).
			FgColor("green").
			Build(),
		ui.Text(""),
		app.TableBuilder().
			Columns([]app.TableColumn{
				{Title: "ID", Width: 5},
				{Title: "Name", Width: 12},
				{Title: "Status", Width: 10},
			}).
			AddRow("1", "Alice", "Active").
			AddRow("2", "Bob", "Active").
			AddRow("3", "Charlie", "Inactive").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Up/Down/Enter: select | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

// callSetter 使用反射调用 setter 函数
func callSetter(fn interface{}, arg interface{}) {
	if fn == nil {
		return
	}
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return
	}
	argV := reflect.ValueOf(arg)
	v.Call([]reflect.Value{argV})
}

func main() {
	err := ui.Run(SelectDemo,
		ui.WithWidth(50),
		ui.WithHeight(22),
		ui.WithTitle("Select & Table Demo (MVP)"),
		ui.WithInit(func() {
			selectedIndexKey := intent.StateKey[int]("selectedIndex")
			selectedIndexSetterKey := intent.StateKey[func(int)]("selectedIndexSetter")

			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				if i.Field == selectedIndexKey.String() {
					setter, _ := ctx.GetState(selectedIndexSetterKey.String())
					callSetter(setter, i.Value)
				}
				return intent.HandledResult()
			})
		}),
	)
	if err != nil {
		panic(err)
	}
}
