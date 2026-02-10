package testing

import (
	"testing"

	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/framework/msg"
	"github.com/wwsheng009/mint/framework/sandbox"
	"github.com/wwsheng009/mint/runtime/platform"
)

// TestInjector_KeySequence 测试按键序列注入
func TestInjector_KeySequence(t *testing.T) {
	injector := sandbox.NewInjector()

	// 注入按键序列
	injector.InjectKeySequence("hello")

	if injector.Count() != 5 {
		t.Errorf("Expected 5 messages, got %d", injector.Count())
	}

	messages := injector.GetMessages()
	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}
}

// TestInjector_NavigationKeys 测试导航键注入
func TestInjector_NavigationKeys(t *testing.T) {
	injector := NewInjector()

	// 注入导航键
	injector.InjectUp()
	injector.InjectDown()
	injector.InjectLeft()
	injector.InjectRight()

	if injector.Count() != 4 {
		t.Errorf("Expected 4 actions, got %d", injector.Count())
	}

	actions := injector.GetActions()
	if len(actions) != 4 {
		t.Errorf("Expected 4 actions, got %d", len(actions))
	}

	// 验证 Action 类型
	if actions[0].Type != action.ActionNavigateUp {
		t.Errorf("First action should be NavigateUp, got %s", actions[0].Type)
	}
	if actions[1].Type != action.ActionNavigateDown {
		t.Errorf("Second action should be NavigateDown, got %s", actions[1].Type)
	}
}

// TestInjector_ModifierKeys 测试修饰键注入
func TestInjector_ModifierKeys(t *testing.T) {
	injector := NewInjector()

	// 注入 Ctrl+C
	injector.InjectCtrlKey('c')

	messages := injector.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	sandboxMsg := messages[0].(*msg.SandboxMsg)
	if !sandboxMsg.KeyData.HasCtrl() {
		t.Error("Key should have Ctrl modifier")
	}
	if sandboxMsg.KeyData.Rune != 'c' {
		t.Errorf("Expected rune 'c', got %c", sandboxMsg.KeyData.Rune)
	}
}

// TestInjector_MouseClick 测试鼠标点击注入
func TestInjector_MouseClick(t *testing.T) {
	injector := NewInjector()

	// 注入鼠标点击
	injector.InjectMouseClick("button1", 10, 5)

	if injector.Count() != 1 {
		t.Errorf("Expected 1 action, got %d", injector.Count())
	}

	actions := injector.GetActions()
	if len(actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(actions))
	}

	act := actions[0]
	if act.Type != action.ActionClick {
		t.Errorf("Action should be Click, got %s", act.Type)
	}
	if act.TargetID != "button1" {
		t.Errorf("TargetID should be 'button1', got %s", act.TargetID)
	}

	x, y, _ := act.GetPayloadPoint()
	if x != 10 || y != 5 {
		t.Errorf("Expected coordinates (10, 5), got (%d, %d)", x, y)
	}
}

// TestInjector_StateMutation 测试状态修改注入
func TestInjector_StateMutation(t *testing.T) {
	injector := NewInjector()

	// 注入状态修改
	injector.InjectSetState("button1", "disabled", true)

	messages := injector.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	sandboxMsg := messages[0].(*msg.SandboxMsg)
	if !sandboxMsg.IsStateMutation() {
		t.Error("Message should be state mutation")
	}

	if sandboxMsg.StateMutation.TargetID != "button1" {
		t.Errorf("TargetID should be 'button1', got %s", sandboxMsg.StateMutation.TargetID)
	}

	if sandboxMsg.StateMutation.Path != "disabled" {
		t.Errorf("Path should be 'disabled', got %s", sandboxMsg.StateMutation.Path)
	}

	if sandboxMsg.StateMutation.Value != true {
		t.Errorf("Value should be true, got %v", sandboxMsg.StateMutation.Value)
	}
}

// TestInjector_SetValue 测试值设置快捷方法
func TestInjector_SetValue(t *testing.T) {
	injector := NewInjector()

	// 使用快捷方法设置值
	injector.InjectSetValue("input1", "test value")

	messages := injector.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	sandboxMsg := messages[0].(*msg.SandboxMsg)
	if sandboxMsg.StateMutation.Path != "value" {
		t.Errorf("Path should be 'value', got %s", sandboxMsg.StateMutation.Path)
	}

	if sandboxMsg.StateMutation.Value != "test value" {
		t.Errorf("Value should be 'test value', got %v", sandboxMsg.StateMutation.Value)
	}
}

// TestInjector_Clear 测试清空注入器
func TestInjector_Clear(t *testing.T) {
	injector := NewInjector()

	// 注入一些内容
	injector.InjectUp()
	injector.InjectDown()

	if injector.Count() != 2 {
		t.Errorf("Expected 2 actions before clear, got %d", injector.Count())
	}

	// 清空
	injector.Clear()

	if injector.Count() != 0 {
		t.Errorf("Expected 0 actions after clear, got %d", injector.Count())
	}

	if !injector.IsEmpty() {
		t.Error("Injector should be empty after clear")
	}
}

// TestInjector_Chaining 测试方法链
func TestInjector_Chaining(t *testing.T) {
	injector := NewInjector()

	// 链式调用
	injector.
		InjectUp().
		InjectDown().
		InjectEnter().
		InjectMouseClick("button1", 0, 0).
		InjectSetState("input1", "value", "test")

	if injector.Count() != 5 {
		t.Errorf("Expected 5 items, got %d", injector.Count())
	}
}

// TestTestableApp_InjectKeySequence 测试 TestableApp 按键序列注入
func TestTestableApp_InjectKeySequence(t *testing.T) {
	// 创建模拟的 Router
	router := action.NewRouter(nil)
	app := NewTestableApp(nil, router)

	// 注入按键序列
	app.InjectKeySequence("test")

	// 验证注入成功（没有错误）
	if app.GetLastError() != nil {
		t.Errorf("Expected no error, got %v", app.GetLastError())
	}
}

// TestTestableApp_InjectMouseClickByID 测试 TestableApp 鼠标点击注入
func TestTestableApp_InjectMouseClickByID(t *testing.T) {
	router := action.NewRouter(nil)
	app := NewTestableApp(nil, router)

	// 注入鼠标点击
	app.InjectMouseClickByID("button1", 10, 5)

	// 验证注入成功
	if app.GetLastError() != nil {
		t.Errorf("Expected no error, got %v", app.GetLastError())
	}
}

// TestTestableApp_InjectText 测试 TestableApp 文本注入
func TestTestableApp_InjectText(t *testing.T) {
	router := action.NewRouter(nil)
	app := NewTestableApp(nil, router)

	// 注入文本
	app.InjectText("input1", "hello world")

	// 验证注入成功
	if app.GetLastError() != nil {
		t.Errorf("Expected no error, got %v", app.GetLastError())
	}
}

// TestTestableApp_SetState 测试 TestableApp 状态设置
func TestTestableApp_SetState(t *testing.T) {
	router := action.NewRouter(nil)
	app := NewTestableApp(nil, router)

	// 设置状态
	app.SetState("button1", "disabled", true)

	// 验证消息被记录
	if app.GetLastMsg() == nil {
		t.Error("Last message should not be nil")
	}
}

// TestTestableApp_AssertFocused 测试焦点断言
func TestTestableApp_AssertFocused(t *testing.T) {
	router := action.NewRouter(nil)
	app := NewTestableApp(nil, router)

	// 断言焦点（应该通过）
	err := app.AssertFocused("button1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestTestableApp_AssertValue 测试值断言
func TestTestableApp_AssertValue(t *testing.T) {
	router := action.NewRouter(nil)
	app := NewTestableApp(nil, router)

	// 断言值（应该通过）
	err := app.AssertValue("input1", "test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestTestableApp_Navigation 测试导航操作
func TestTestableApp_Navigation(t *testing.T) {
	router := action.NewRouter(nil)
	app := NewTestableApp(nil, router)

	// 注入导航键
	app.InjectKey(0, platform.KeyDown, msg.Modifiers{})
	app.InjectKey(0, platform.KeyUp, msg.Modifiers{})
	app.InjectKey(0, platform.KeyEnter, msg.Modifiers{})

	// 验证没有错误
	if app.GetLastError() != nil {
		t.Errorf("Expected no error, got %v", app.GetLastError())
	}
}

// TestTestableApp_ComplexScenario 测试复杂场景
func TestTestableApp_ComplexScenario(t *testing.T) {
	router := action.NewRouter(nil)
	app := NewTestableApp(nil, router)

	// 模拟用户交互流程：
	// 1. 输入文本
	app.InjectText("input1", "hello")

	// 2. 按 Tab 键切换焦点
	app.InjectTab()

	// 3. 点击按钮
	app.InjectMouseClickByID("button1", 10, 5)

	// 4. 按 Enter 确认
	app.InjectEnter()

	// 验证所有操作都成功
	if app.GetLastError() != nil {
		t.Errorf("Expected no error, got %v", app.GetLastError())
	}
}

// TestInjector_HelperMethods 测试辅助方法
func TestInjector_HelperMethods(t *testing.T) {
	injector := NewInjector()

	// 测试 HasMessages 和 HasActions
	if injector.HasMessages() {
		t.Error("Should not have messages initially")
	}
	if injector.HasActions() {
		t.Error("Should not have actions initially")
	}
	if !injector.IsEmpty() {
		t.Error("Should be empty initially")
	}

	// 添加一些内容
	injector.InjectUp()

	if !injector.HasActions() {
		t.Error("Should have actions after injection")
	}
	if injector.IsEmpty() {
		t.Error("Should not be empty after injection")
	}
}
