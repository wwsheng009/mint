package button

import (
	"testing"

	"github.com/wwsheng009/mint/framework/component"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// TestButtonUpdate_MouseClick 直接测试 Button 组件的 Update(Msg) 方法
// 绕过 HitMap 和 ComponentRegistry，直接验证鼠标点击是否工作
func TestButtonUpdate_MouseClick(t *testing.T) {
	// 创建一个 Button 组件
	clicked := false
	button := NewButton("Click Me")
	button.SetOnClick(func() {
		clicked = true
	})

	// 验证 Button 实现了 Updater 接口
	var _ component.Updater = button

	// 测试鼠标点击
	mouseMsg := &runtimemsg.MouseMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue: runtimemsg.MsgTypeMouse,
		},
		Action:  runtimemsg.MouseActionPress,
		Button:  runtimemsg.MouseLeft,
		LocalX:  5,
		LocalY:  0,
	}

	// 调用 Update 方法
	cmd := button.Update(mouseMsg)
	if cmd != nil {
		t.Logf("Update returned Cmd: %v", cmd)
	}

	// 验证 onClick 被触发
	if !clicked {
		t.Error("Button click did not trigger onClick callback")
	} else {
		t.Logf("✅ Button click correctly triggered onClick callback")
	}
}

// TestButtonUpdate_MouseRelease 测试鼠标释放不应触发 onClick
func TestButtonUpdate_MouseRelease(t *testing.T) {
	clicked := false
	button := NewButton("Click Me")
	button.SetOnClick(func() {
		clicked = true
	})

	// 发送鼠标释放事件
	mouseMsg := &runtimemsg.MouseMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue: runtimemsg.MsgTypeMouse,
		},
		Action: runtimemsg.MouseActionRelease,
		Button: runtimemsg.MouseLeft,
	}

	button.Update(mouseMsg)

	// 验证 onClick 没有被触发
	if clicked {
		t.Error("Mouse release should not trigger onClick")
	} else {
		t.Logf("✅ Mouse release correctly did not trigger onClick")
	}
}

// TestButtonUpdate_RightClick 测试右键点击不应触发 onClick
func TestButtonUpdate_RightClick(t *testing.T) {
	clicked := false
	button := NewButton("Click Me")
	button.SetOnClick(func() {
		clicked = true
	})

	// 发送右键点击
	mouseMsg := &runtimemsg.MouseMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue: runtimemsg.MsgTypeMouse,
		},
		Action: runtimemsg.MouseActionPress,
		Button: runtimemsg.MouseRight,
	}

	button.Update(mouseMsg)

	// 验证 onClick 没有被触发
	if clicked {
		t.Error("Right click should not trigger onClick")
	} else {
		t.Logf("✅ Right click correctly did not trigger onClick")
	}
}

// TestButtonUpdate_Disabled 测试禁用的按钮不应该响应点击
func TestButtonUpdate_Disabled(t *testing.T) {
	clicked := false
	button := NewButton("Click Me")
	button.SetOnClick(func() {
		clicked = true
	})
	button.SetDisabled(true)

	// 发送鼠标点击
	mouseMsg := &runtimemsg.MouseMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue: runtimemsg.MsgTypeMouse,
		},
		Action: runtimemsg.MouseActionPress,
		Button: runtimemsg.MouseLeft,
	}

	button.Update(mouseMsg)

	// 验证 onClick 没有被触发
	if clicked {
		t.Error("Disabled button should not trigger onClick")
	} else {
		t.Logf("✅ Disabled button correctly did not trigger onClick")
	}
}
