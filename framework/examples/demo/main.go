package main

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/display"
	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/framework/form"
	"github.com/wwsheng009/mint/framework/input"
	"github.com/wwsheng009/mint/framework/interactive"
	"github.com/wwsheng009/mint/framework/layout"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/framework/validation"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("   TUI Framework V3 - 组件演示程序")
	fmt.Println("==============================================")
	fmt.Println()

	// 演示基础组件
	demoBasicComponents()

	// 演示布局组件
	demoLayouts()

	// 演示表单组件
	demoForms()

	// 演示列表和表格
	demoListsAndTables()

	// 演示事件系统
	demoEventSystem()

	fmt.Println("\n==============================================")
	fmt.Println("   所有演示完成！")
	fmt.Println("==============================================")
}

func demoBasicComponents() {
	fmt.Println(">>> 基础组件演示")

	// 1. Text 组件
	fmt.Println("\n--- Text 组件 ---")

	title := display.NewText("TUI Framework V3")
	title.SetStyle(style.Style{}.Foreground(style.Blue).Bold(true))
	fmt.Printf("标题: %s\n", title.GetContent())

	subtitle := display.NewText("新一代终端 UI 框架")
	fmt.Printf("副标题: %s\n", subtitle.GetContent())

	multiline := display.NewText("支持多行文本\n自动换行\n样式应用")
	fmt.Printf("多行文本:\n%s\n", multiline.GetContent())

	// 2. TextInput 组件
	fmt.Println("\n--- TextInput 组件 ---")

	username := input.NewTextInput()
	username.SetPlaceholder("请输入用户名")
	username.SetValue("demo_user")
	fmt.Printf("用户名: %s\n", username.GetValue())

	password := input.NewTextInput()
	password.SetPlaceholder("请输入密码")
	password.SetPassword(true)
	password.SetValue("secret123")
	fmt.Printf("密码: %s (长度=%d)\n", maskString(password.GetValue()), len(username.GetValue()))

	// 3. Button 组件
	fmt.Println("\n--- Button 组件 ---")

	submitBtn := interactive.NewButton("提交")
	fmt.Printf("按钮标签: %s\n", submitBtn.GetLabel())

	cancelBtn := interactive.NewButton("取消")
	cancelBtn.SetNormalStyle(style.Style{}.Foreground(style.Red))
	fmt.Printf("取消按钮: %s\n", cancelBtn.GetLabel())
}

func demoLayouts() {
	fmt.Println("\n>>> 布局组件演示")

	// Box 容器
	fmt.Println("\n--- Box 容器 ---")

	box := layout.NewBox().WithBorder(true).WithBorderColor(style.Cyan).WithPadding(1)

	fmt.Printf("Box 边框: %v\n", box.GetBorder() != nil)
	fmt.Printf("Box 内边距: %+v\n", box.GetPadding())

	fmt.Println("\n--- Flex 布局 (待实现) ---")
	fmt.Println("Flex Row: 水平布局")
	fmt.Println("Flex Column: 垂直布局")
}

func demoForms() {
	fmt.Println("\n>>> 表单组件演示")

	// 创建表单
	loginForm := form.NewForm()
	loginForm.SetLabelStyle(style.Style{}.Foreground(style.Cyan))
	loginForm.SetErrorStyle(style.Style{}.Foreground(style.Red))

	// 添加字段
	usernameField := form.NewFormField("username")
	usernameField.Label = "用户名"
	usernameField.Placeholder = "请输入用户名"
	usernameField.Input = input.NewTextInput()
	usernameField.Validators = []validation.Validator{
		validation.Required(),
		validation.MinLength(3),
	}

	passwordField := form.NewFormField("password")
	passwordField.Label = "密码"
	passwordField.Placeholder = "请输入密码"
	passwordField.Input = input.NewTextInput()
	passwordField.Input.(*input.TextInput).SetPassword(true)
	passwordField.Validators = []validation.Validator{
		validation.Required(),
		validation.MinLength(6),
	}

	emailField := form.NewFormField("email")
	emailField.Label = "邮箱"
	emailField.Placeholder = "请输入邮箱"
	emailField.Input = input.NewTextInput()
	emailField.Validators = []validation.Validator{
		validation.Required(),
		validation.Email(),
	}

	loginForm.AddField(usernameField)
	loginForm.AddField(passwordField)
	loginForm.AddField(emailField)

	// 设置提交回调
	loginForm.SetOnSubmit(func(data map[string]interface{}) error {
		fmt.Printf("\n✓ 表单提交成功!\n")
		fmt.Printf("  用户名: %v\n", data["username"])
		fmt.Printf("  密码: %v\n", data["password"])
		fmt.Printf("  邮箱: %v\n", data["email"])
		return nil
	})

	loginForm.SetOnCancel(func() {
		fmt.Println("\n✗ 表单已取消")
	})

	// 测试表单
	fmt.Println("\n表单字段:")
	for _, field := range loginForm.GetFields() {
		fmt.Printf("  - %s: %s\n", field.Label, field.Placeholder)
		if len(field.Validators) > 0 {
			fmt.Printf("    验证器: %d 个\n", len(field.Validators))
		}
	}

	// 验证测试
	fmt.Println("\n验证测试:")
	valid := loginForm.IsValid()
	fmt.Printf("  表单有效: %v (初始状态)\n", valid)

	// 设置值并测试验证
	_ = loginForm.SetValue("username", "ab") // 太短
	_ = loginForm.SetValue("password", "123") // 太短
	_ = loginForm.SetValue("email", "invalid") // 无效邮箱

	field, _ := loginForm.GetField("username")
	err := field.Validate()
	fmt.Printf("  用户名验证: %v\n", err)

	field, _ = loginForm.GetField("email")
	err = field.Validate()
	fmt.Printf("  邮箱验证: %v\n", err)

	// 有效数据
	_ = loginForm.SetValue("username", "demo_user")
	_ = loginForm.SetValue("password", "password123")
	_ = loginForm.SetValue("email", "user@example.com")

	valid = loginForm.IsValid()
	fmt.Printf("  表单有效: %v (有效数据)\n", valid)
}

func demoListsAndTables() {
	fmt.Println("\n>>> 列表和表格演示")

	// List 组件
	fmt.Println("\n--- List 组件 ---")

	items := []string{
		"项目 1: 学习 Go 语言",
		"项目 2: 开发 TUI 框架",
		"项目 3: 编写文档",
		"项目 4: 单元测试",
		"项目 5: 发布版本",
	}
	_ = display.NewListStrings(items)

	fmt.Printf("列表项数量: %d\n", len(items))

	// Table 组件
	fmt.Println("\n--- Table 组件 ---")

	table := display.NewTable([]display.TableColumn{
		{Title: "ID", Width: 10},
		{Title: "名称", Width: 30},
		{Title: "状态", Width: 15},
		{Title: "创建时间", Width: 20},
	})

	table.SetRows([][]string{
		{"1", "用户管理", "完成", "2024-01-20"},
		{"2", "订单系统", "进行中", "2024-01-21"},
		{"3", "数据同步", "待开始", "2024-01-22"},
	})

	fmt.Printf("表格列数: %d\n", table.GetColumnCount())
	fmt.Printf("表格行数: %d\n", table.GetRowCount())
}

func demoEventSystem() {
	fmt.Println("\n>>> 事件系统演示")

	// 创建各种事件
	fmt.Println("--- 事件类型 ---")

	// 特殊键事件
	specialEvent := event.NewSpecialKeyEvent(event.KeyEnter)
	fmt.Printf("特殊键事件: Enter (类型: %d)\n", specialEvent.Type())

	// 事件属性
	fmt.Println("\n--- 事件属性 ---")

	specialEvent = event.NewSpecialKeyEvent(event.KeyTab)
	fmt.Printf("Tab 键 - 类型: 导航键\n")

	specialEvent = event.NewSpecialKeyEvent(event.KeyEscape)
	fmt.Printf("Esc 键 - 类型: 系统键\n")

	specialEvent = event.NewSpecialKeyEvent(event.KeyF1)
	fmt.Printf("F1 键 - 类型: 功能键\n")

	// Vim 风格键
	kEvent := event.NewSpecialKeyEvent(event.KeyK)
	fmt.Printf("Vim K 键: %d\n", kEvent.Special)

	jEvent := event.NewSpecialKeyEvent(event.KeyJ)
	fmt.Printf("Vim J 键: %d\n", jEvent.Special)

	// 事件处理流程
	fmt.Println("\n--- 事件处理流程 ---")
	fmt.Println("  1. Platform → RawInput")
	fmt.Println("  2. Runtime → KeyMap → Action")
	fmt.Println("  3. Component → HandleAction()")
	fmt.Println("  4. State → Update & Render")

	// Action 类型示例
	fmt.Println("\n--- Action 类型 ---")
	fmt.Println("  导航: navigate_up, navigate_down, navigate_next, navigate_prev")
	fmt.Println("  编辑: input_char, delete_char, backspace")
	fmt.Println("  表单: submit, cancel, validate")
	fmt.Println("  视图: scroll_up, scroll_down, zoom_in, zoom_out")
	fmt.Println("  系统: quit, copy, paste, undo, redo")
}

func maskString(s string) string {
	if len(s) <= 2 {
		return "***"
	}
	return s[:1] + "***" + s[len(s)-1:]
}
