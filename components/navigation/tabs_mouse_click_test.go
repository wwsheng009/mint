package navigation

import (
	"testing"

	"github.com/wwsheng009/mint/framework/component"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// TestTabsUpdate_MouseClick 直接测试 Tabs 组件的 Update(Msg) 方法
// 绕过 HitMap 和 ComponentRegistry，直接验证鼠标点击是否工作
func TestTabsUpdate_MouseClick(t *testing.T) {
	// 创建一个 Tabs 组件
	tabs := NewTabs()
	tabs.SetTabs([]TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
		{ID: "tab3", Label: "Tab 3"},
	})

	// 验证初始状态
	if tabs.ActiveTab() != 0 {
		t.Errorf("Expected active tab to be 0, got %d", tabs.ActiveTab())
	}

	// 验证 Tabs 实现了 Updater 接口
	var _ component.Updater = tabs

	// 测试鼠标点击切换到第二个 Tab (LocalX 在 "Tab 2" 范围内)
	// Tab bar 格式: "Tab 1 | [Tab 2] | Tab 3"
	// "Tab 1" 长度 5 + " | " 长度 3 = 8
	// "[Tab 2]" 长度 7，所以点击 LocalX=8-14 应该切换到 Tab 2

	testCases := []struct {
		name          string
		localX        int
		localY        int
		expectedTab   int
		shouldSwitch  bool
	}{
		{
			name:         "Click on Tab 1 (first tab)",
			localX:       2, // Within "Tab 1"
			localY:       0, // Tab bar is at y=0
			expectedTab:  0,
			shouldSwitch: false, // Already on Tab 1
		},
		{
			name:         "Click on Tab 2",
			localX:       10, // Within "[Tab 2]"
			localY:       0,  // Tab bar is at y=0
			expectedTab:  1,
			shouldSwitch: true,
		},
		{
			name:         "Click on Tab 3",
			localX:       20, // Within "Tab 3"
			localY:       0,  // Tab bar is at y=0
			expectedTab:  2,
			shouldSwitch: true,
		},
		{
			name:         "Click below tab bar",
			localX:       10,
			localY:       5, // Below tab bar
			expectedTab:  0, // Should not change
			shouldSwitch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset to first tab
			tabs.SetActiveTab(0)

			// 创建鼠标点击消息
			mouseMsg := &runtimemsg.MouseMsg{
				BaseMsg: runtimemsg.BaseMsg{
					TypeValue: runtimemsg.MsgTypeMouse,
				},
				Action:  runtimemsg.MouseActionPress,
				Button:  runtimemsg.MouseLeft,
				LocalX:  tc.localX,
				LocalY:  tc.localY,
			}

			// 调用 Update 方法
			cmd := tabs.Update(mouseMsg)
			if cmd != nil {
				t.Logf("Update returned Cmd: %v", cmd)
			}

			// 验证结果
			if tabs.ActiveTab() != tc.expectedTab {
				t.Errorf("After click at (%d,%d), expected active tab=%d, got %d",
					tc.localX, tc.localY, tc.expectedTab, tabs.ActiveTab())
			} else {
				t.Logf("✅ Click at (%d,%d) correctly switched to tab %d", tc.localX, tc.localY, tabs.ActiveTab())
			}
		})
	}
}

// TestTabsUpdate_WithOnChangeCallback 测试 onChange 回调是否被触发
func TestTabsUpdate_WithOnChangeCallback(t *testing.T) {
	// 创建一个 Tabs 组件
	tabs := NewTabs()
	tabs.SetTabs([]TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	})

	// 设置 onChange 回调
	capturedTabID := ""
	tabs.SetOnChange(func(tabID string) {
		capturedTabID = tabID
	})

	// 点击第二个 Tab
	mouseMsg := &runtimemsg.MouseMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue: runtimemsg.MsgTypeMouse,
		},
		Action:  runtimemsg.MouseActionPress,
		Button:  runtimemsg.MouseLeft,
		LocalX:  10, // Within "Tab 2"
		LocalY:  0,
	}

	tabs.Update(mouseMsg)

	// 验证回调被触发
	if capturedTabID != "tab2" {
		t.Errorf("Expected onChange to be called with 'tab2', got '%s'", capturedTabID)
	} else {
		t.Logf("✅ onChange callback was correctly called with 'tab2'")
	}
}

// TestTabsUpdate_MouseRelease 测试鼠标释放事件不应触发 Tab 切换
func TestTabsUpdate_MouseRelease(t *testing.T) {
	tabs := NewTabs()
	tabs.SetTabs([]TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	})
	tabs.SetActiveTab(0)

	// 发送鼠标释放事件（不是点击）
	mouseMsg := &runtimemsg.MouseMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue: runtimemsg.MsgTypeMouse,
		},
		Action: runtimemsg.MouseActionRelease,
		LocalX: 10,
		LocalY:  0,
	}

	tabs.Update(mouseMsg)

	// 验证 Tab 没有切换
	if tabs.ActiveTab() != 0 {
		t.Errorf("Mouse release should not switch tab, expected 0, got %d", tabs.ActiveTab())
	} else {
		t.Logf("✅ Mouse release correctly did not switch tab")
	}
}

// TestTabsUpdate_RightClick 测试右键点击不应触发 Tab 切换
func TestTabsUpdate_RightClick(t *testing.T) {
	tabs := NewTabs()
	tabs.SetTabs([]TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	})
	tabs.SetActiveTab(0)

	// 发送右键点击
	mouseMsg := &runtimemsg.MouseMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue: runtimemsg.MsgTypeMouse,
		},
		Action: runtimemsg.MouseActionPress,
		Button: runtimemsg.MouseRight,
		LocalX: 10,
		LocalY:  0,
	}

	tabs.Update(mouseMsg)

	// 验证 Tab 没有切换
	if tabs.ActiveTab() != 0 {
		t.Errorf("Right click should not switch tab, expected 0, got %d", tabs.ActiveTab())
	} else {
		t.Logf("✅ Right click correctly did not switch tab")
	}
}
