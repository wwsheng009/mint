package button

import (
	"testing"

	"github.com/wwsheng009/mint/framework/action"
)

// TestButton_ActionTarget 测试 ActionTarget 基础接口
func TestButton_ActionTarget(t *testing.T) {
	button := NewButton("Test")

	// 测试 GetSupportedActions
	supported := button.GetSupportedActions()
	if len(supported) == 0 {
		t.Error("GetSupportedActions() should return non-empty list")
	}

	// 验证支持 Click Action
	hasClick := false
	hasEnter := false
	for _, act := range supported {
		if act == action.ActionClick {
			hasClick = true
		}
		if act == action.ActionEnter {
			hasEnter = true
		}
	}
	if !hasClick {
		t.Error("GetSupportedActions() should include ActionClick")
	}
	if !hasEnter {
		t.Error("GetSupportedActions() should include ActionEnter")
	}
}

// TestButton_HandleAction_Click 测试点击 Action
func TestButton_HandleAction_Click(t *testing.T) {
	clicked := false
	button := NewButton("Test")
	button.SetOnClick(func() {
		clicked = true
	})

	// 测试 Click Action
	act := action.NewAction(action.ActionClick)
	if !button.HandleAction(act) {
		t.Error("HandleAction(Click) should return true")
	}
	if !clicked {
		t.Error("onClick should be called after HandleAction(Click)")
	}
}

// TestButton_HandleAction_Enter 测试 Enter Action
func TestButton_HandleAction_Enter(t *testing.T) {
	clicked := false
	button := NewButton("Test")
	button.SetOnClick(func() {
		clicked = true
	})

	// 测试 Enter Action
	act := action.NewAction(action.ActionEnter)
	if !button.HandleAction(act) {
		t.Error("HandleAction(Enter) should return true")
	}
	if !clicked {
		t.Error("onClick should be called after HandleAction(Enter)")
	}
}

// TestButton_HandleAction_Disabled 测试禁用状态
func TestButton_HandleAction_Disabled(t *testing.T) {
	clicked := false
	button := NewButton("Test")
	button.SetOnClick(func() {
		clicked = true
	})
	button.SetDisabled(true)

	// 测试禁用状态下不应触发点击
	act := action.NewAction(action.ActionClick)
	if button.HandleAction(act) {
		t.Error("HandleAction(Click) should return false when button is disabled")
	}
	if clicked {
		t.Error("onClick should not be called when button is disabled")
	}
}

// TestButton_HandleAction_NoHandler 测试无回调函数
func TestButton_HandleAction_NoHandler(t *testing.T) {
	button := NewButton("Test")
	// 不设置 onClick

	// 测试无回调函数时应返回 false
	act := action.NewAction(action.ActionClick)
	if button.HandleAction(act) {
		t.Error("HandleAction(Click) should return false when onClick is nil")
	}
}

// TestButton_HandleAction_Unsupported 测试不支持的 Action
func TestButton_HandleAction_Unsupported(t *testing.T) {
	clicked := false
	button := NewButton("Test")
	button.SetOnClick(func() {
		clicked = true
	})

	// 测试不支持的 Action
	act := action.NewAction(action.ActionNavigateUp)
	if button.HandleAction(act) {
		t.Error("HandleAction(NavigateUp) should return false")
	}
	if clicked {
		t.Error("onClick should not be called for unsupported action")
	}
}

// TestButton_HandleAction_Nil 测试 nil Action
func TestButton_HandleAction_Nil(t *testing.T) {
	button := NewButton("Test")

	// 测试 nil Action
	if button.HandleAction(nil) {
		t.Error("HandleAction(nil) should return false")
	}
}

// TestButton_CanHandleAction 测试 CanHandleAction
func TestButton_CanHandleAction(t *testing.T) {
	button := NewButton("Test")

	// 测试支持的 Action
	clickAct := action.NewAction(action.ActionClick)
	if !button.CanHandleAction(clickAct) {
		t.Error("CanHandleAction() should return true for Click")
	}

	enterAct := action.NewAction(action.ActionEnter)
	if !button.CanHandleAction(enterAct) {
		t.Error("CanHandleAction() should return true for Enter")
	}

	// 测试不支持的 Action
	navigateAct := action.NewAction(action.ActionNavigateUp)
	if button.CanHandleAction(navigateAct) {
		t.Error("CanHandleAction() should return false for NavigateUp")
	}

	// 测试 nil Action
	if button.CanHandleAction(nil) {
		t.Error("CanHandleAction() should return false for nil action")
	}
}

// TestButton_CanHandleAction_Disabled 测试禁用状态
func TestButton_CanHandleAction_Disabled(t *testing.T) {
	button := NewButton("Test")
	button.SetDisabled(true)

	// 禁用状态下不应处理任何 Action
	clickAct := action.NewAction(action.ActionClick)
	if button.CanHandleAction(clickAct) {
		t.Error("CanHandleAction() should return false when button is disabled")
	}
}

// TestButton_FocusableActionTarget 测试 FocusableActionTarget 接口
func TestButton_FocusableActionTarget(t *testing.T) {
	button := NewButton("Test")

	// 测试 Focus
	if !button.Focus() {
		t.Error("Focus() should return true")
	}
	if !button.IsFocused() {
		t.Error("IsFocused() should return true after Focus()")
	}

	// 测试 Blur
	button.Blur()
	if button.IsFocused() {
		t.Error("IsFocused() should return false after Blur()")
	}

	// 测试 IsFocusable
	if !button.IsFocusable() {
		t.Error("IsFocusable() should return true")
	}
}

// TestButton_FocusableActionTarget_Disabled 测试禁用状态的焦点
func TestButton_FocusableActionTarget_Disabled(t *testing.T) {
	button := NewButton("Test")
	button.SetDisabled(true)

	// 测试禁用状态下不能获得焦点
	if button.Focus() {
		t.Error("Focus() should return false when button is disabled")
	}
	if button.IsFocusable() {
		t.Error("IsFocusable() should return false when button is disabled")
	}
}

// TestButton_MultipleActions 测试多次 Action
func TestButton_MultipleActions(t *testing.T) {
	clickCount := 0
	button := NewButton("Test")
	button.SetOnClick(func() {
		clickCount++
	})

	// 触发多次点击
	for i := 0; i < 3; i++ {
		act := action.NewAction(action.ActionClick)
		if !button.HandleAction(act) {
			t.Errorf("HandleAction(Click) #%d should return true", i+1)
		}
	}

	if clickCount != 3 {
		t.Errorf("onClick should be called 3 times, got %d", clickCount)
	}
}

// TestButton_ActionWithSource 测试带来源的 Action
func TestButton_ActionWithSource(t *testing.T) {
	clicked := false
	button := NewButton("Test")
	button.SetOnClick(func() {
		clicked = true
	})

	// 测试带来源的 Action
	act := action.NewAction(action.ActionClick).WithSource("keyboard")
	if !button.HandleAction(act) {
		t.Error("HandleAction(Click with source) should return true")
	}
	if !clicked {
		t.Error("onClick should be called for Action with source")
	}
}

// TestButton_EnterSameAsClick 测试 Enter 和 Click 行为相同
func TestButton_EnterSameAsClick(t *testing.T) {
	clickCount := 0
	button := NewButton("Test")
	button.SetOnClick(func() {
		clickCount++
	})

	// 测试 Enter 和 Click 都触发相同的回调
	clickAct := action.NewAction(action.ActionClick)
	enterAct := action.NewAction(action.ActionEnter)

	button.HandleAction(clickAct)
	button.HandleAction(enterAct)

	if clickCount != 2 {
		t.Errorf("Both Click and Enter should trigger onClick, got %d calls", clickCount)
	}
}
