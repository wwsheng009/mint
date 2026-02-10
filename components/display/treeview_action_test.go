package display

import (
	"testing"

	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/runtime/platform"
)

// TestTreeView_ActionTarget 测试 ActionTarget 基础接口
func TestTreeView_ActionTarget(t *testing.T) {
	tree := createTestTreeView()

	// 测试 GetSupportedActions
	supported := tree.GetSupportedActions()
	if len(supported) == 0 {
		t.Error("GetSupportedActions() should return non-empty list")
	}

	// 验证支持导航 Action
	hasNavigateUp := false
	for _, act := range supported {
		if act == action.ActionNavigateUp {
			hasNavigateUp = true
			break
		}
	}
	if !hasNavigateUp {
		t.Error("GetSupportedActions() should include ActionNavigateUp")
	}
}

// TestTreeView_HandleAction_Navigation 测试导航 Action
func TestTreeView_HandleAction_Navigation(t *testing.T) {
	tree := createTestTreeView()

	// 设置初始焦点
	tree.SetFocusIndex(2)
	initialFocus := tree.GetFocusIndex()

	// 测试 NavigateUp
	act := action.NewAction(action.ActionNavigateUp)
	if !tree.HandleAction(act) {
		t.Error("HandleAction(NavigateUp) should return true")
	}
	if tree.GetFocusIndex() >= initialFocus {
		t.Errorf("NavigateUp should decrease focus index, got %d, was %d", tree.GetFocusIndex(), initialFocus)
	}

	// 测试 NavigateDown
	tree.SetFocusIndex(2)
	act = action.NewAction(action.ActionNavigateDown)
	if !tree.HandleAction(act) {
		t.Error("HandleAction(NavigateDown) should return true")
	}
	if tree.GetFocusIndex() <= 2 {
		t.Errorf("NavigateDown should increase focus index, got %d", tree.GetFocusIndex())
	}

	// 测试 NavigateHome
	act = action.NewAction(action.ActionNavigateHome)
	if !tree.HandleAction(act) {
		t.Error("HandleAction(NavigateHome) should return true")
	}
	if tree.GetFocusIndex() != 0 {
		t.Errorf("NavigateHome should set focus to 0, got %d", tree.GetFocusIndex())
	}

	// 测试 NavigateEnd
	act = action.NewAction(action.ActionNavigateEnd)
	if !tree.HandleAction(act) {
		t.Error("HandleAction(NavigateEnd) should return true")
	}
	expectedEnd := len(tree.lines) - 1
	if tree.GetFocusIndex() != expectedEnd {
		t.Errorf("NavigateEnd should set focus to %d, got %d", expectedEnd, tree.GetFocusIndex())
	}
}

// TestTreeView_HandleAction_Selection 测试选择 Action
func TestTreeView_HandleAction_Selection(t *testing.T) {
	tree := createTestTreeView()
	tree.SetFocusIndex(2)

	// 测试 Select
	act := action.NewAction(action.ActionSelect)
	if !tree.HandleAction(act) {
		t.Error("HandleAction(Select) should return true")
	}
	if !tree.HasSelection() {
		t.Error("HasSelection() should return true after Select action")
	}
	if tree.GetSelectedLine().NodeID != tree.lines[2].NodeID {
		t.Error("Selected line should match focus index")
	}
}

// TestTreeView_HandleAction_Toggle 测试切换 Action
func TestTreeView_HandleAction_Toggle(t *testing.T) {
	tree := createTestTreeView()
	tree.SetFocusIndex(1)

	// 测试 Toggle (展开/折叠)
	act := action.NewAction(action.ActionToggle)
	if !tree.HandleAction(act) {
		t.Error("HandleAction(Toggle) should return true")
	}

	// Verify expand state changed flag
	if !tree.ExpandStateChanged() {
		t.Error("ExpandStateChanged() should return true after Toggle action")
	}

	tree.ClearExpandStateChanged()
}

// TestTreeView_HandleAction_Scroll 测试滚动 Action
func TestTreeView_HandleAction_Scroll(t *testing.T) {
	tree := createTestTreeView()
	tree.SetFocusIndex(2)

	// 测试 Scroll Down
	act := action.NewActionWithPayload(action.ActionScroll, 1)
	if !tree.HandleAction(act) {
		t.Error("HandleAction(Scroll down) should return true")
	}
	if tree.GetFocusIndex() <= 2 {
		t.Errorf("Scroll down should increase focus index, got %d", tree.GetFocusIndex())
	}

	// 测试 Scroll Up
	tree.SetFocusIndex(5)
	act = action.NewActionWithPayload(action.ActionScroll, -1)
	if !tree.HandleAction(act) {
		t.Error("HandleAction(Scroll up) should return true")
	}
	if tree.GetFocusIndex() >= 5 {
		t.Errorf("Scroll up should decrease focus index, got %d", tree.GetFocusIndex())
	}
}

// TestTreeView_CanHandleAction 测试 CanHandleAction
func TestTreeView_CanHandleAction(t *testing.T) {
	tree := createTestTreeView()

	// 测试支持的 Action
	navigateAct := action.NewAction(action.ActionNavigateUp)
	if !tree.CanHandleAction(navigateAct) {
		t.Error("CanHandleAction() should return true for NavigateUp")
	}

	selectAct := action.NewAction(action.ActionSelect)
	if !tree.CanHandleAction(selectAct) {
		t.Error("CanHandleAction() should return true for Select")
	}

	scrollAct := action.NewAction(action.ActionScroll)
	if !tree.CanHandleAction(scrollAct) {
		t.Error("CanHandleAction() should return true for Scroll")
	}

	// 测试不支持的 Action
	inputAct := action.NewAction(action.ActionInputText)
	if tree.CanHandleAction(inputAct) {
		t.Error("CanHandleAction() should return false for InputText")
	}

	// 测试 nil Action
	if tree.CanHandleAction(nil) {
		t.Error("CanHandleAction() should return false for nil action")
	}
}

// TestTreeView_FocusableActionTarget 测试 FocusableActionTarget 接口
func TestTreeView_FocusableActionTarget(t *testing.T) {
	tree := createTestTreeView()

	// 测试 Focus
	if !tree.Focus() {
		t.Error("Focus() should return true")
	}
	if !tree.IsFocused() {
		t.Error("IsFocused() should return true after Focus()")
	}

	// 测试 Blur
	tree.Blur()
	if tree.IsFocused() {
		t.Error("IsFocused() should return false after Blur()")
	}

	// 测试 IsFocusable
	if !tree.IsFocusable() {
		t.Error("IsFocusable() should return true")
	}
}

// TestTreeView_ScrollableActionTarget 测试 ScrollableActionTarget 接口
func TestTreeView_ScrollableActionTarget(t *testing.T) {
	tree := createTestTreeView()
	tree.SetFocusIndex(5)

	// 测试 CanScroll
	if !tree.CanScroll(1) {
		t.Error("CanScroll(1) should return true")
	}
	if !tree.CanScroll(-1) {
		t.Error("CanScroll(-1) should return true")
	}

	// 移动到边界
	tree.SetFocusIndex(0)
	if tree.CanScroll(-1) {
		t.Error("CanScroll(-1) should return false at top")
	}

	tree.SetFocusIndex(len(tree.lines) - 1)
	if tree.CanScroll(1) {
		t.Error("CanScroll(1) should return false at bottom")
	}

	// 测试 Scroll
	tree.SetFocusIndex(5)
	if !tree.Scroll(1) {
		t.Error("Scroll(1) should return true")
	}
	if tree.GetFocusIndex() != 6 {
		t.Errorf("Scroll(1) should increase focus index to 6, got %d", tree.GetFocusIndex())
	}

	// 测试 GetScrollPosition
	current, total, visible := tree.GetScrollPosition()
	if current != tree.GetFocusIndex() {
		t.Error("GetScrollPosition() current should match focus index")
	}
	if total != len(tree.lines) {
		t.Error("GetScrollPosition() total should match line count")
	}
	if visible <= 0 {
		t.Error("GetScrollPosition() visible should be positive")
	}
}

// TestTreeView_SelectableActionTarget 测试 SelectableActionTarget 接口
func TestTreeView_SelectableActionTarget(t *testing.T) {
	tree := createTestTreeView()

	// 测试 Select
	if !tree.Select() {
		t.Error("Select() should return true")
	}
	if !tree.IsSelected() {
		t.Error("IsSelected() should return true after Select()")
	}

	// 测试 ToggleSelection
	tree.ToggleSelection()
	if tree.IsSelected() {
		t.Error("IsSelected() should be false after ToggleSelection()")
	}

	tree.ToggleSelection()
	if !tree.IsSelected() {
		t.Error("IsSelected() should be true after ToggleSelection()")
	}

	// 测试 GetSelectedCount
	count := tree.GetSelectedCount()
	if count != 1 {
		t.Errorf("GetSelectedCount() should return 1, got %d", count)
	}

	tree.ClearSelection()
	if tree.GetSelectedCount() != 0 {
		t.Error("GetSelectedCount() should return 0 after ClearSelection()")
	}
}

// TestTreeView_ActionWithKeyboard 测试键盘 Action 处理
func TestTreeView_ActionWithKeyboard(t *testing.T) {
	tree := createTestTreeView()

	// 测试 PageUp
	tree.SetFocusIndex(50)
	act := action.NewAction(action.ActionNavigatePageUp)
	tree.HandleAction(act)
	if tree.GetFocusIndex() >= 50 {
		t.Error("NavigatePageUp should decrease focus index")
	}

	// 测试 PageDown
	tree.SetFocusIndex(5)
	act = action.NewAction(action.ActionNavigatePageDown)
	tree.HandleAction(act)
	if tree.GetFocusIndex() <= 5 {
		t.Error("NavigatePageDown should increase focus index")
	}
}

// TestTreeView_BackwardCompatibility 测试与原有 HandleKey 方法的兼容性
func TestTreeView_BackwardCompatibility(t *testing.T) {
	tree := createTestTreeView()

	// 原有的 HandleKey 方法应该仍然工作
	if !tree.HandleKey(platform.KeyUp, 0) {
		t.Error("HandleKey(KeyUp) should return true")
	}

	// 新的 HandleAction 方法应该提供相同的功能
	act := action.NewAction(action.ActionNavigateUp)
	if !tree.HandleAction(act) {
		t.Error("HandleAction(NavigateUp) should return true")
	}

	// 两种方法的效果应该相同
	tree.SetFocusIndex(5)
	tree.HandleKey(platform.KeyDown, 0)
	focusAfterKey := tree.GetFocusIndex()

	tree.SetFocusIndex(5)
	tree.HandleAction(action.NewAction(action.ActionNavigateDown))
	focusAfterAction := tree.GetFocusIndex()

	if focusAfterKey != focusAfterAction {
		t.Errorf("HandleKey and HandleAction should produce same result: key=%d, action=%d",
			focusAfterKey, focusAfterAction)
	}
}

// createTestTreeView 创建一个测试用的 TreeView
func createTestTreeView() *TreeView {
	lines := []string{
		"Root",
		"├── Node 1",
		"│   ├── Node 1.1",
		"│   └── Node 1.2",
		"├── Node 2",
		"│   ├── Node 2.1",
		"│   └── Node 2.2",
		"└── Node 3",
		"    ├── Node 3.1",
		"    └── Node 3.2",
	}

	builder := NewTreeView().
		FromLines(lines).
		ExpandLevel(2)

	tree := builder.Build().(*TreeView)

	// Set viewport height for testing
	tree.SetViewportHeight(10)

	return tree
}
