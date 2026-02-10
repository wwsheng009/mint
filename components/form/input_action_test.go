package form

import (
	"testing"

	"github.com/wwsheng009/mint/framework/action"
)

// TestInput_ActionTarget 测试 ActionTarget 基础接口
func TestInput_ActionTarget(t *testing.T) {
	input := NewInput()

	// 测试 GetSupportedActions
	supported := input.GetSupportedActions()
	if len(supported) == 0 {
		t.Error("GetSupportedActions() should return non-empty list")
	}

	// 验证支持的 Actions
	hasInputText := false
	hasBackspace := false
	hasDeleteChar := false
	hasEnter := false

	for _, act := range supported {
		if act == action.ActionInputText {
			hasInputText = true
		}
		if act == action.ActionBackspace {
			hasBackspace = true
		}
		if act == action.ActionDeleteChar {
			hasDeleteChar = true
		}
		if act == action.ActionEnter {
			hasEnter = true
		}
	}

	if !hasInputText {
		t.Error("GetSupportedActions() should include ActionInputText")
	}
	if !hasBackspace {
		t.Error("GetSupportedActions() should include ActionBackspace")
	}
	if !hasDeleteChar {
		t.Error("GetSupportedActions() should include ActionDeleteChar")
	}
	if !hasEnter {
		t.Error("GetSupportedActions() should include ActionEnter")
	}
}

// TestInput_HandleAction_InputText 测试文本输入 Action
func TestInput_HandleAction_InputText(t *testing.T) {
	input := NewInput()

	// 测试 InputText Action
	act := action.NewActionWithPayload(action.ActionInputText, "Hello")
	if !input.HandleAction(act) {
		t.Error("HandleAction(InputText) should return true")
	}
	if input.Value() != "Hello" {
		t.Errorf("Value should be \"Hello\", got %q", input.Value())
	}
	if input.GetCursorPosition() != 5 {
		t.Errorf("Cursor position should be 5, got %d", input.GetCursorPosition())
	}
}

// TestInput_HandleAction_MaxLength 测试最大长度限制
func TestInput_HandleAction_MaxLength(t *testing.T) {
	input := NewInput()
	input.SetMaxLength(5)

	// 测试超出最大长度
	act := action.NewActionWithPayload(action.ActionInputText, "HelloWorld")
	if input.HandleAction(act) {
		t.Error("HandleAction(InputText) should return false when exceeding max length")
	}

	// 测试在最大长度内
	act = action.NewActionWithPayload(action.ActionInputText, "Hi")
	if !input.HandleAction(act) {
		t.Error("HandleAction(InputText) should return true within max length")
	}
	if input.Value() != "Hi" {
		t.Errorf("Value should be \"Hi\", got %q", input.Value())
	}
}

// TestInput_HandleAction_Backspace 测试 Backspace Action
func TestInput_HandleAction_Backspace(t *testing.T) {
	input := NewInput()
	input.SetValue("Hello")
	input.SetCursorPosition(5)

	// 测试 Backspace Action
	act := action.NewAction(action.ActionBackspace)
	if !input.HandleAction(act) {
		t.Error("HandleAction(Backspace) should return true")
	}
	if input.Value() != "Hell" {
		t.Errorf("Value should be \"Hell\", got %q", input.Value())
	}
	if input.GetCursorPosition() != 4 {
		t.Errorf("Cursor position should be 4, got %d", input.GetCursorPosition())
	}
}

// TestInput_HandleAction_BackspaceAtStart 测试在开头 Backspace
func TestInput_HandleAction_BackspaceAtStart(t *testing.T) {
	input := NewInput()
	input.SetValue("Hello")
	input.SetCursorPosition(0)

	// 测试在开头 Backspace
	act := action.NewAction(action.ActionBackspace)
	if input.HandleAction(act) {
		t.Error("HandleAction(Backspace) should return false at cursor position 0")
	}
	if input.Value() != "Hello" {
		t.Errorf("Value should remain \"Hello\", got %q", input.Value())
	}
}

// TestInput_HandleAction_DeleteChar 测试 Delete Action
func TestInput_HandleAction_DeleteChar(t *testing.T) {
	input := NewInput()
	input.SetValue("Hello")
	input.SetCursorPosition(2) // Position at 'l'

	// 测试 DeleteChar Action
	act := action.NewAction(action.ActionDeleteChar)
	if !input.HandleAction(act) {
		t.Error("HandleAction(DeleteChar) should return true")
	}
	if input.Value() != "Helo" {
		t.Errorf("Value should be \"Helo\", got %q", input.Value())
	}
	if input.GetCursorPosition() != 2 {
		t.Errorf("Cursor position should remain 2, got %d", input.GetCursorPosition())
	}
}

// TestInput_HandleAction_DeleteCharAtEnd 测试在末尾 Delete
func TestInput_HandleAction_DeleteCharAtEnd(t *testing.T) {
	input := NewInput()
	input.SetValue("Hello")
	input.SetCursorPosition(5) // At end

	// 测试在末尾 Delete
	act := action.NewAction(action.ActionDeleteChar)
	if input.HandleAction(act) {
		t.Error("HandleAction(DeleteChar) should return false at end of text")
	}
	if input.Value() != "Hello" {
		t.Errorf("Value should remain \"Hello\", got %q", input.Value())
	}
}

// TestInput_HandleAction_Enter 测试 Enter Action
func TestInput_HandleAction_Enter(t *testing.T) {
	submitted := false
	input := NewInput()
	input.SetOnSubmit(func() {
		submitted = true
	})

	// 测试 Enter Action
	act := action.NewAction(action.ActionEnter)
	if !input.HandleAction(act) {
		t.Error("HandleAction(Enter) should return true")
	}
	if !submitted {
		t.Error("onSubmit should be called after HandleAction(Enter)")
	}
}

// TestInput_HandleAction_Disabled 测试禁用状态
func TestInput_HandleAction_Disabled(t *testing.T) {
	input := NewInput()
	input.SetDisabled(true)

	// 测试禁用状态下不应处理 Action
	act := action.NewActionWithPayload(action.ActionInputText, "test")
	if input.HandleAction(act) {
		t.Error("HandleAction(InputText) should return false when disabled")
	}
	if input.Value() != "" {
		t.Errorf("Value should remain empty, got %q", input.Value())
	}
}

// TestInput_HandleAction_ReadOnly 测试只读状态
func TestInput_HandleAction_ReadOnly(t *testing.T) {
	input := NewInput()
	input.SetReadOnly(true)

	// 测试只读状态下不应处理 Action
	act := action.NewActionWithPayload(action.ActionInputText, "test")
	if input.HandleAction(act) {
		t.Error("HandleAction(InputText) should return false when read-only")
	}
	if input.Value() != "" {
		t.Errorf("Value should remain empty, got %q", input.Value())
	}
}

// TestInput_CanHandleAction 测试 CanHandleAction
func TestInput_CanHandleAction(t *testing.T) {
	input := NewInput()

	// 测试支持的 Action
	inputAct := action.NewActionWithPayload(action.ActionInputText, "test")
	if !input.CanHandleAction(inputAct) {
		t.Error("CanHandleAction() should return true for InputText")
	}

	backspaceAct := action.NewAction(action.ActionBackspace)
	if !input.CanHandleAction(backspaceAct) {
		t.Error("CanHandleAction() should return true for Backspace")
	}

	// 测试不支持的 Action
	navigateAct := action.NewAction(action.ActionNavigateUp)
	if input.CanHandleAction(navigateAct) {
		t.Error("CanHandleAction() should return false for NavigateUp")
	}
}

// TestInput_CanHandleAction_Disabled 测试禁用状态能力检查
func TestInput_CanHandleAction_Disabled(t *testing.T) {
	input := NewInput()
	input.SetDisabled(true)

	// 禁用状态下不能处理任何 Action
	act := action.NewAction(action.ActionInputText)
	if input.CanHandleAction(act) {
		t.Error("CanHandleAction() should return false when disabled")
	}
}

// TestInput_FocusableActionTarget 测试 FocusableActionTarget 接口
func TestInput_FocusableActionTarget(t *testing.T) {
	input := NewInput()

	// 测试 Focus
	if !input.Focus() {
		t.Error("Focus() should return true")
	}
	if !input.IsFocused() {
		t.Error("IsFocused() should return true after Focus()")
	}

	// 测试 Blur
	input.Blur()
	if input.IsFocused() {
		t.Error("IsFocused() should return false after Blur()")
	}

	// 测试 IsFocusable
	if !input.IsFocusable() {
		t.Error("IsFocusable() should return true")
	}
}

// TestInput_FocusableActionTarget_Disabled 测试禁用状态焦点
func TestInput_FocusableActionTarget_Disabled(t *testing.T) {
	input := NewInput()
	input.SetDisabled(true)

	// 测试禁用状态下不能获得焦点
	if input.Focus() {
		t.Error("Focus() should return false when disabled")
	}
	if input.IsFocusable() {
		t.Error("IsFocusable() should return false when disabled")
	}
}

// TestInput_EditableActionTarget_InsertText 测试 InsertText
func TestInput_EditableActionTarget_InsertText(t *testing.T) {
	input := NewInput()
	input.SetValue("He")
	input.SetCursorPosition(2)

	// 测试插入文本
	if !input.InsertText("llo") {
		t.Error("InsertText() should return true")
	}
	if input.Value() != "Hello" {
		t.Errorf("Value should be \"Hello\", got %q", input.Value())
	}
	if input.GetCursorPosition() != 5 {
		t.Errorf("Cursor position should be 5, got %d", input.GetCursorPosition())
	}
}

// TestInput_EditableActionTarget_InsertTextAtEnd 测试在末尾插入
func TestInput_EditableActionTarget_InsertTextAtEnd(t *testing.T) {
	input := NewInput()
	input.SetValue("Hello")
	input.SetCursorPosition(5)

	// 测试在末尾插入
	if !input.InsertText(" World") {
		t.Error("InsertText() should return true")
	}
	if input.Value() != "Hello World" {
		t.Errorf("Value should be \"Hello World\", got %q", input.Value())
	}
	if input.GetCursorPosition() != 11 {
		t.Errorf("Cursor position should be 11, got %d", input.GetCursorPosition())
	}
}

// TestInput_EditableActionTarget_DeleteText 测试 DeleteText
func TestInput_EditableActionTarget_DeleteText(t *testing.T) {
	input := NewInput()
	input.SetValue("Hello")
	input.SetCursorPosition(3) // Position at 'l'

	// 测试向前删除
	if !input.DeleteText(1) {
		t.Error("DeleteText(1) should return true")
	}
	if input.Value() != "Helo" {
		t.Errorf("Value should be \"Helo\", got %q", input.Value())
	}
	if input.GetCursorPosition() != 3 {
		t.Errorf("Cursor position should remain 3, got %d", input.GetCursorPosition())
	}
}

// TestInput_EditableActionTarget_ReplaceText 测试 ReplaceText
func TestInput_EditableActionTarget_ReplaceText(t *testing.T) {
	input := NewInput()

	// 测试替换文本
	if !input.ReplaceText("Hello World") {
		t.Error("ReplaceText() should return true")
	}
	if input.Value() != "Hello World" {
		t.Errorf("Value should be \"Hello World\", got %q", input.Value())
	}
	if input.GetCursorPosition() != 11 {
		t.Errorf("Cursor position should be 11, got %d", input.GetCursorPosition())
	}
}

// TestInput_EditableActionTarget_GetText 测试 GetText
func TestInput_EditableActionTarget_GetText(t *testing.T) {
	input := NewInput()
	input.SetValue("Test Value")

	if input.GetText() != "Test Value" {
		t.Errorf("GetText() should return \"Test Value\", got %q", input.GetText())
	}
}

// TestInput_EditableActionTarget_SetCursorPosition 测试 SetCursorPosition
func TestInput_EditableActionTarget_SetCursorPosition(t *testing.T) {
	input := NewInput()
	input.SetValue("Hello")

	// 测试设置有效位置
	if !input.SetCursorPosition(3) {
		t.Error("SetCursorPosition(3) should return true")
	}
	if input.GetCursorPosition() != 3 {
		t.Errorf("Cursor position should be 3, got %d", input.GetCursorPosition())
	}

	// 测试设置无效位置（超出范围）
	if input.SetCursorPosition(100) {
		t.Error("SetCursorPosition(100) should return false")
	}
	if input.GetCursorPosition() != 3 {
		t.Errorf("Cursor position should remain 3, got %d", input.GetCursorPosition())
	}

	// 测试设置负数位置
	if input.SetCursorPosition(-1) {
		t.Error("SetCursorPosition(-1) should return false")
	}
}

// TestInput_OnChange 测试 onChange 回调
func TestInput_OnChange(t *testing.T) {
	changeCount := 0
	var lastValue string

	input := NewInput()
	input.SetOnChange(func(value string) {
		changeCount++
		lastValue = value
	})

	// 测试 InsertText 触发 onChange
	input.InsertText("Hello")
	if changeCount != 1 {
		t.Errorf("onChange should be called once, got %d times", changeCount)
	}
	if lastValue != "Hello" {
		t.Errorf("onChange value should be \"Hello\", got %q", lastValue)
	}

	// 测试 DeleteText 触发 onChange
	input.DeleteText(-1)
	if changeCount != 2 {
		t.Errorf("onChange should be called twice, got %d times", changeCount)
	}
	if lastValue != "Hell" {
		t.Errorf("onChange value should be \"Hell\", got %q", lastValue)
	}
}

// TestInput_MultipleActions 测试多次 Action
func TestInput_MultipleActions(t *testing.T) {
	input := NewInput()

	// 连续插入文本
	actions := []string{"H", "e", "l", "l", "o"}
	for _, text := range actions {
		act := action.NewActionWithPayload(action.ActionInputText, text)
		if !input.HandleAction(act) {
			t.Errorf("HandleAction(InputText, %q) should return true", text)
		}
	}

	if input.Value() != "Hello" {
		t.Errorf("Value should be \"Hello\", got %q", input.Value())
	}
}

// TestInput_CursorMovement 测试光标移动
func TestInput_CursorMovement(t *testing.T) {
	input := NewInput()
	input.SetValue("Hello World")

	// 测试移动光标到不同位置
	positions := []int{0, 5, 11}
	for _, pos := range positions {
		if !input.SetCursorPosition(pos) {
			t.Errorf("SetCursorPosition(%d) should return true", pos)
		}
		if input.GetCursorPosition() != pos {
			t.Errorf("Cursor position should be %d, got %d", pos, input.GetCursorPosition())
		}
	}
}
