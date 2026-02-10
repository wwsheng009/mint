package navigation

import (
	"testing"

	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/ui"
)

// TestTabs_ActionTarget 测试 ActionTarget 基础接口
func TestTabs_ActionTarget(t *testing.T) {
	tabs := createTestTabs()

	// 测试 GetSupportedActions
	supported := tabs.GetSupportedActions()
	if len(supported) == 0 {
		t.Error("GetSupportedActions() should return non-empty list")
	}

	// 验证支持导航 Action
	hasNavigateNext := false
	for _, act := range supported {
		if act == action.ActionNavigateNext {
			hasNavigateNext = true
			break
		}
	}
	if !hasNavigateNext {
		t.Error("GetSupportedActions() should include ActionNavigateNext")
	}
}

// TestTabs_HandleAction_Navigation 测试导航 Action
func TestTabs_HandleAction_Navigation(t *testing.T) {
	tabs := createTestTabs()

	// 初始状态：第一个 tab 激活
	if tabs.ActiveTab() != 0 {
		t.Errorf("Initial active tab should be 0, got %d", tabs.ActiveTab())
	}

	// 测试 NavigateNext
	act := action.NewAction(action.ActionNavigateNext)
	if !tabs.HandleAction(act) {
		t.Error("HandleAction(NavigateNext) should return true")
	}
	if tabs.ActiveTab() != 1 {
		t.Errorf("NavigateNext should switch to tab 1, got %d", tabs.ActiveTab())
	}

	// 测试 NavigatePrev
	act = action.NewAction(action.ActionNavigatePrev)
	if !tabs.HandleAction(act) {
		t.Error("HandleAction(NavigatePrev) should return true")
	}
	if tabs.ActiveTab() != 0 {
		t.Errorf("NavigatePrev should switch to tab 0, got %d", tabs.ActiveTab())
	}

	// 测试 NavigateRight (等同于 Next)
	act = action.NewAction(action.ActionNavigateRight)
	if !tabs.HandleAction(act) {
		t.Error("HandleAction(NavigateRight) should return true")
	}

	// 测试 NavigateLeft (等同于 Prev)
	act = action.NewAction(action.ActionNavigateLeft)
	if !tabs.HandleAction(act) {
		t.Error("HandleAction(NavigateLeft) should return true")
	}

	// 测试 NavigateHome
	tabs.SetActiveTab(2)
	act = action.NewAction(action.ActionNavigateHome)
	if !tabs.HandleAction(act) {
		t.Error("HandleAction(NavigateHome) should return true")
	}
	if tabs.ActiveTab() != 0 {
		t.Errorf("NavigateHome should switch to first tab, got %d", tabs.ActiveTab())
	}

	// 测试 NavigateEnd
	act = action.NewAction(action.ActionNavigateEnd)
	if !tabs.HandleAction(act) {
		t.Error("HandleAction(NavigateEnd) should return true")
	}
	expectedEnd := len(tabs.tabs) - 1
	if tabs.ActiveTab() != expectedEnd {
		t.Errorf("NavigateEnd should switch to last tab %d, got %d", expectedEnd, tabs.ActiveTab())
	}
}

// TestTabs_HandleAction_Selection 测试选择 Action
func TestTabs_HandleAction_Selection(t *testing.T) {
	tabs := createTestTabs()

	// 测试 Select
	act := action.NewAction(action.ActionSelect)
	if !tabs.HandleAction(act) {
		t.Error("HandleAction(Select) should return true")
	}

	// 测试 Enter
	act = action.NewAction(action.ActionEnter)
	if !tabs.HandleAction(act) {
		t.Error("HandleAction(Enter) should return true")
	}
}

// TestTabs_HandleAction_Scroll 测试滚动 Action
func TestTabs_HandleAction_Scroll(t *testing.T) {
	tabs := createTestTabs()
	tabs.SetActiveTab(0)

	// 测试 Scroll Down (Next)
	act := action.NewActionWithPayload(action.ActionScroll, 1)
	if !tabs.HandleAction(act) {
		t.Error("HandleAction(Scroll down) should return true")
	}
	if tabs.ActiveTab() != 1 {
		t.Errorf("Scroll down should switch to tab 1, got %d", tabs.ActiveTab())
	}

	// 测试 Scroll Up (Prev)
	act = action.NewActionWithPayload(action.ActionScroll, -1)
	if !tabs.HandleAction(act) {
		t.Error("HandleAction(Scroll up) should return true")
	}
	if tabs.ActiveTab() != 0 {
		t.Errorf("Scroll up should switch to tab 0, got %d", tabs.ActiveTab())
	}
}

// TestTabs_CanHandleAction 测试 CanHandleAction
func TestTabs_CanHandleAction(t *testing.T) {
	tabs := createTestTabs()

	// 测试支持的 Action
	navigateAct := action.NewAction(action.ActionNavigateNext)
	if !tabs.CanHandleAction(navigateAct) {
		t.Error("CanHandleAction() should return true for NavigateNext")
	}

	selectAct := action.NewAction(action.ActionSelect)
	if !tabs.CanHandleAction(selectAct) {
		t.Error("CanHandleAction() should return true for Select")
	}

	scrollAct := action.NewAction(action.ActionScroll)
	if !tabs.CanHandleAction(scrollAct) {
		t.Error("CanHandleAction() should return true for Scroll")
	}

	// 测试不支持的 Action
	inputAct := action.NewAction(action.ActionInputText)
	if tabs.CanHandleAction(inputAct) {
		t.Error("CanHandleAction() should return false for InputText")
	}

	// 测试 nil Action
	if tabs.CanHandleAction(nil) {
		t.Error("CanHandleAction() should return false for nil action")
	}
}

// TestTabs_FocusableActionTarget 测试 FocusableActionTarget 接口
func TestTabs_FocusableActionTarget(t *testing.T) {
	tabs := createTestTabs()

	// 测试 Focus
	if !tabs.Focus() {
		t.Error("Focus() should return true")
	}
	if !tabs.IsFocused() {
		t.Error("IsFocused() should return true after Focus()")
	}

	// 测试 Blur
	tabs.Blur()
	// Blur doesn't change focus state for tabs
	if !tabs.IsFocused() {
		t.Error("IsFocused() should still return true after Blur()")
	}

	// 测试 IsFocusable
	if !tabs.IsFocusable() {
		t.Error("IsFocusable() should return true")
	}

	// 测试空 tabs
	emptyTabs := NewTabs()
	if emptyTabs.Focus() {
		t.Error("Focus() should return false for empty tabs")
	}
	if emptyTabs.IsFocusable() {
		t.Error("IsFocusable() should return false for empty tabs")
	}
}

// TestTabs_ScrollableActionTarget 测试 ScrollableActionTarget 接口
func TestTabs_ScrollableActionTarget(t *testing.T) {
	tabs := createTestTabs()
	tabs.SetActiveTab(1)

	// 测试 CanScroll
	if !tabs.CanScroll(1) {
		t.Error("CanScroll(1) should return true")
	}
	if !tabs.CanScroll(-1) {
		t.Error("CanScroll(-1) should return true")
	}

	// 移动到边界
	tabs.SetActiveTab(0)
	if tabs.CanScroll(-1) {
		t.Error("CanScroll(-1) should return false at first tab")
	}

	tabs.SetActiveTab(len(tabs.tabs) - 1)
	if tabs.CanScroll(1) {
		t.Error("CanScroll(1) should return false at last tab")
	}

	// 测试 Scroll
	tabs.SetActiveTab(1)
	if !tabs.Scroll(1) {
		t.Error("Scroll(1) should return true")
	}
	if tabs.ActiveTab() != 2 {
		t.Errorf("Scroll(1) should increase active tab to 2, got %d", tabs.ActiveTab())
	}

	// 测试 GetScrollPosition
	current, total, visible := tabs.GetScrollPosition()
	if current != tabs.ActiveTab() {
		t.Error("GetScrollPosition() current should match active tab")
	}
	if total != len(tabs.tabs) {
		t.Error("GetScrollPosition() total should match tab count")
	}
	if visible != 1 {
		t.Errorf("GetScrollPosition() visible should be 1, got %d", visible)
	}
}

// TestTabs_SelectableActionTarget 测试 SelectableActionTarget 接口
func TestTabs_SelectableActionTarget(t *testing.T) {
	tabs := createTestTabs()

	// 测试 Select
	if !tabs.Select() {
		t.Error("Select() should return true")
	}
	if !tabs.IsSelected() {
		t.Error("IsSelected() should return true after Select()")
	}

	// 测试 ToggleSelection
	initialTab := tabs.ActiveTab()
	tabs.ToggleSelection()
	if tabs.ActiveTab() == initialTab {
		// Toggle should navigate to next tab
		t.Error("ToggleSelection() should change active tab")
	}

	// 测试 GetSelectedCount
	count := tabs.GetSelectedCount()
	if count != 1 {
		t.Errorf("GetSelectedCount() should return 1, got %d", count)
	}

	// 测试空 tabs
	emptyTabs := NewTabs()
	if emptyTabs.GetSelectedCount() != 0 {
		t.Error("GetSelectedCount() should return 0 for empty tabs")
	}
}

// TestTabs_NavigationWithDisabledTabs 测试禁用 tab 的导航
func TestTabs_NavigationWithDisabledTabs(t *testing.T) {
	tabs := createTestTabs()

	// 禁用第二个 tab
	tabs.SetTabEnabled(1, false)

	// 从第一个 tab 向前导航
	tabs.SetActiveTab(0)
	act := action.NewAction(action.ActionNavigateNext)
	tabs.HandleAction(act)

	// 应该跳过禁用的 tab，直接到第三个 tab
	if tabs.ActiveTab() != 2 {
		t.Errorf("NavigateNext should skip disabled tab, expected tab 2, got %d", tabs.ActiveTab())
	}
}

// TestTabs_NavigationBoundaries 测试边界情况
func TestTabs_NavigationBoundaries(t *testing.T) {
	tabs := createTestTabs()

	// 在第一个 tab 向前导航
	tabs.SetActiveTab(0)
	act := action.NewAction(action.ActionNavigatePrev)
	if tabs.HandleAction(act) {
		t.Error("NavigatePrev at first tab should return false")
	}
	if tabs.ActiveTab() != 0 {
		t.Errorf("Active tab should remain 0, got %d", tabs.ActiveTab())
	}

	// 在最后一个 tab 向后导航
	tabs.SetActiveTab(len(tabs.tabs) - 1)
	act = action.NewAction(action.ActionNavigateNext)
	if tabs.HandleAction(act) {
		t.Error("NavigateNext at last tab should return false")
	}
	if tabs.ActiveTab() != len(tabs.tabs)-1 {
		t.Errorf("Active tab should remain at last, got %d", tabs.ActiveTab())
	}
}

// TestTabs_ScrollBoundaries 测试滚动边界
func TestTabs_ScrollBoundaries(t *testing.T) {
	tabs := createTestTabs()
	tabs.SetActiveTab(0)

	// 尝试向前滚动
	if tabs.Scroll(-1) {
		t.Error("Scroll(-1) at first tab should return false")
	}

	// 尝试向后滚动
	tabs.SetActiveTab(len(tabs.tabs) - 1)
	if tabs.Scroll(1) {
		t.Error("Scroll(1) at last tab should return false")
	}
}

// createTestTabs 创建一个测试用的 Tabs
func createTestTabs() *TabsVNode {
	builder := TabsBuilder()

	// 添加 3 个 tabs
	builder.AddTab("tab1", "Tab 1")
	builder.AddTab("tab2", "Tab 2")
	builder.AddTab("tab3", "Tab 3")

	// 添加内容
	builder.Content("tab1", ui.Text("Content 1"))
	builder.Content("tab2", ui.Text("Content 2"))
	builder.Content("tab3", ui.Text("Content 3"))

	return builder.Build().(*TabsVNode)
}
