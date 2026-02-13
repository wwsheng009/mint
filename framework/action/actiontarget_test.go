package action

import (
	"testing"
)

// ============================================================================
// Mock 实现
// ============================================================================

// mockActionTarget 是一个简单的 ActionTarget 实现
type mockActionTarget struct {
	*BaseActionTarget
	id          string
	handleFunc  func(*Action) bool
	canHandleFunc func(*Action) bool
}

func (m *mockActionTarget) HandleAction(action *Action) bool {
	if m.handleFunc != nil {
		return m.handleFunc(action)
	}
	return m.BaseActionTarget.HandleAction(action)
}

func (m *mockActionTarget) CanHandleAction(action *Action) bool {
	if m.canHandleFunc != nil {
		return m.canHandleFunc(action)
	}
	return m.BaseActionTarget.CanHandleAction(action)
}

// mockFocusableTarget 是一个支持焦点的目标
type mockFocusableTarget struct {
	*mockActionTarget
	focused bool
}

func (m *mockFocusableTarget) Focus() bool {
	m.focused = true
	return true
}

func (m *mockFocusableTarget) Blur() {
	m.focused = false
}

func (m *mockFocusableTarget) IsFocused() bool {
	return m.focused
}

func (m *mockFocusableTarget) IsFocusable() bool {
	return true
}

// mockScrollableTarget 是一个支持滚动的目标
type mockScrollableTarget struct {
	*mockActionTarget
	position int
	total    int
	visible  int
}

func (m *mockScrollableTarget) CanScroll(delta int) bool {
	newPos := m.position + delta
	return newPos >= 0 && newPos <= m.total
}

func (m *mockScrollableTarget) Scroll(delta int) bool {
	if !m.CanScroll(delta) {
		return false
	}
	m.position += delta
	return true
}

func (m *mockScrollableTarget) GetScrollPosition() (int, int, int) {
	return m.position, m.total, m.visible
}

// mockEditableTarget 是一个可编辑的目标
type mockEditableTarget struct {
	*mockActionTarget
	text       string
	cursorPos  int
}

func (m *mockEditableTarget) InsertText(text string) bool {
	m.text = m.text[:m.cursorPos] + text + m.text[m.cursorPos:]
	m.cursorPos += len(text)
	return true
}

func (m *mockEditableTarget) DeleteText(direction int) bool {
	if direction < 0 {
		// Backspace
		if m.cursorPos == 0 {
			return false
		}
		m.text = m.text[:m.cursorPos-1] + m.text[m.cursorPos:]
		m.cursorPos--
	} else {
		// Delete
		if m.cursorPos >= len(m.text) {
			return false
		}
		m.text = m.text[:m.cursorPos] + m.text[m.cursorPos+1:]
	}
	return true
}

func (m *mockEditableTarget) ReplaceText(text string) bool {
	m.text = text
	m.cursorPos = len(text)
	return true
}

func (m *mockEditableTarget) GetText() string {
	return m.text
}

func (m *mockEditableTarget) GetCursorPosition() int {
	return m.cursorPos
}

func (m *mockEditableTarget) SetCursorPosition(pos int) bool {
	if pos < 0 || pos > len(m.text) {
		return false
	}
	m.cursorPos = pos
	return true
}

// mockSelectableTarget 是一个可选择的目标
type mockSelectableTarget struct {
	*mockActionTarget
	selected bool
}

func (m *mockSelectableTarget) Select() bool {
	m.selected = true
	return true
}

func (m *mockSelectableTarget) IsSelected() bool {
	return m.selected
}

func (m *mockSelectableTarget) ToggleSelection() bool {
	m.selected = !m.selected
	return true
}

func (m *mockSelectableTarget) GetSelectedCount() int {
	if m.selected {
		return 1
	}
	return 0
}

// mockExpandableTarget 是一个可展开的目标
type mockExpandableTarget struct {
	*mockActionTarget
	expanded bool
}

func (m *mockExpandableTarget) Expand() bool {
	m.expanded = true
	return true
}

func (m *mockExpandableTarget) Collapse() bool {
	m.expanded = false
	return true
}

func (m *mockExpandableTarget) IsExpanded() bool {
	return m.expanded
}

func (m *mockExpandableTarget) Toggle() bool {
	m.expanded = !m.expanded
	return true
}

// ============================================================================
// ActionTarget 接口测试
// ============================================================================

// TestActionTarget_HandleAction 测试基础 HandleAction
func TestActionTarget_HandleAction(t *testing.T) {
	target := &mockActionTarget{
		BaseActionTarget: NewBaseActionTarget(ActionClick),
		id:                "button-1",
		handleFunc: func(action *Action) bool {
			return action.Type == ActionClick && action.TargetID == 12345
		},
	}

	// 测试匹配的 Action
	action1 := NewActionFromMouse(ActionClick, 12345, 10, 20)
	if !target.HandleAction(action1) {
		t.Error("HandleAction() should return true for matching action")
	}

	// 测试不匹配的 Action
	action2 := NewAction(ActionNavigateDown)
	if target.HandleAction(action2) {
		t.Error("HandleAction() should return false for non-matching action")
	}
}

// TestActionTarget_GetSupportedActions 测试获取支持的 Actions
func TestActionTarget_GetSupportedActions(t *testing.T) {
	supported := []ActionType{ActionClick, ActionHover}
	target := NewBaseActionTarget(supported...)

	actions := target.GetSupportedActions()
	if len(actions) != len(supported) {
		t.Errorf("GetSupportedActions() returned %d actions, want %d", len(actions), len(supported))
	}

	for i, action := range actions {
		if action != supported[i] {
			t.Errorf("GetSupportedActions()[%d] = %v, want %v", i, action, supported[i])
		}
	}
}

// TestActionTarget_CanHandleAction 测试 CanHandleAction
func TestActionTarget_CanHandleAction(t *testing.T) {
	target := NewBaseActionTarget(ActionClick, ActionHover)

	// 测试支持的 Action
	clickAction := NewAction(ActionClick)
	if !target.CanHandleAction(clickAction) {
		t.Error("CanHandleAction() should return true for supported action")
	}

	// 测试不支持的 Action
	navigateAction := NewAction(ActionNavigateDown)
	if target.CanHandleAction(navigateAction) {
		t.Error("CanHandleAction() should return false for unsupported action")
	}
}

// ============================================================================
// 辅助接口测试
// ============================================================================

// TestFocusableActionTarget 测试可聚焦目标
func TestFocusableActionTarget(t *testing.T) {
	target := &mockFocusableTarget{
		mockActionTarget: &mockActionTarget{
			BaseActionTarget: NewBaseActionTarget(ActionFocus),
		},
		focused: false,
	}

	// 测试 Focus
	if !target.Focus() {
		t.Error("Focus() should return true")
	}
	if !target.IsFocused() {
		t.Error("IsFocused() should return true after Focus()")
	}

	// 测试 Blur
	target.Blur()
	if target.IsFocused() {
		t.Error("IsFocused() should return false after Blur()")
	}

	// 测试 IsFocusable
	if !target.IsFocusable() {
		t.Error("IsFocusable() should return true")
	}
}

// TestScrollableActionTarget 测试可滚动目标
func TestScrollableActionTarget(t *testing.T) {
	target := &mockScrollableTarget{
		mockActionTarget: &mockActionTarget{
			BaseActionTarget: NewBaseActionTarget(ActionScroll),
		},
		position: 50,
		total:    100,
		visible:  20,
	}

	// 测试 CanScroll
	if !target.CanScroll(10) {
		t.Error("CanScroll(10) should return true")
	}
	if !target.CanScroll(-50) {
		t.Error("CanScroll(-50) should return true")
	}
	if target.CanScroll(100) { // 超出上边界
		t.Error("CanScroll(100) should return false")
	}
	if target.CanScroll(-100) { // 超出下边界
		t.Error("CanScroll(-100) should return false")
	}

	// 测试 Scroll
	initialPos := target.position
	if !target.Scroll(10) {
		t.Error("Scroll(10) should return true")
	}
	if target.position != initialPos+10 {
		t.Errorf("Scroll(10) changed position to %d, want %d", target.position, initialPos+10)
	}

	// 测试 GetScrollPosition
	current, total, visible := target.GetScrollPosition()
	if current != target.position || total != target.total || visible != target.visible {
		t.Error("GetScrollPosition() returned incorrect values")
	}
}

// TestEditableActionTarget 测试可编辑目标
func TestEditableActionTarget(t *testing.T) {
	target := &mockEditableTarget{
		mockActionTarget: &mockActionTarget{
			BaseActionTarget: NewBaseActionTarget(ActionInputText, ActionBackspace),
		},
		text:      "hello",
		cursorPos: 5,
	}

	// 测试 InsertText
	if !target.InsertText(" world") {
		t.Error("InsertText() should return true")
	}
	if target.GetText() != "hello world" {
		t.Errorf("InsertText() resulted in %q, want %q", target.GetText(), "hello world")
	}
	if target.GetCursorPosition() != 11 {
		t.Errorf("CursorPosition after InsertText() = %d, want 11", target.GetCursorPosition())
	}

	// 测试 DeleteText (Delete - 向前删除)
	// 光标在末尾，无法向前删除
	if target.DeleteText(1) {
		t.Error("DeleteText(1) should return false when at end of text")
	}
	if target.GetText() != "hello world" {
		t.Errorf("Text should not change after failed DeleteText(1), got %q", target.GetText())
	}

	// 移动光标到空格后的位置（指向 'w'）
	target.SetCursorPosition(6)

	// 测试 DeleteText (Delete - 向前删除，删除光标位置的字符 'w')
	if !target.DeleteText(1) {
		t.Error("DeleteText(1) should return true")
	}
	if target.GetText() != "hello orld" {
		t.Errorf("DeleteText(1) resulted in %q, want %q", target.GetText(), "hello orld")
	}
	if target.GetCursorPosition() != 6 {
		t.Errorf("CursorPosition should stay at %d after DeleteText(1), got %d", 6, target.GetCursorPosition())
	}

	// 测试 DeleteText (Backspace - 向后删除)
	// 删除光标前的字符 (空格)
	if !target.DeleteText(-1) {
		t.Error("DeleteText(-1) should return true")
	}
	if target.GetText() != "helloorld" {
		t.Errorf("DeleteText(-1) resulted in %q, want %q", target.GetText(), "helloorld")
	}
	if target.GetCursorPosition() != 5 {
		t.Errorf("CursorPosition should be at 5 after DeleteText(-1), got %d", target.GetCursorPosition())
	}

	// 测试 ReplaceText
	if !target.ReplaceText("new text") {
		t.Error("ReplaceText() should return true")
	}
	if target.GetText() != "new text" {
		t.Errorf("ReplaceText() resulted in %q, want %q", target.GetText(), "new text")
	}

	// 测试 SetCursorPosition
	if !target.SetCursorPosition(3) {
		t.Error("SetCursorPosition(3) should return true")
	}
	if target.GetCursorPosition() != 3 {
		t.Errorf("CursorPosition = %d, want 3", target.GetCursorPosition())
	}

	// 测试无效的 SetCursorPosition
	if target.SetCursorPosition(100) {
		t.Error("SetCursorPosition(100) should return false for out of range")
	}
}

// TestSelectableActionTarget 测试可选择目标
func TestSelectableActionTarget(t *testing.T) {
	target := &mockSelectableTarget{
		mockActionTarget: &mockActionTarget{
			BaseActionTarget: NewBaseActionTarget(ActionSelect),
		},
		selected: false,
	}

	// 测试 Select
	if !target.Select() {
		t.Error("Select() should return true")
	}
	if !target.IsSelected() {
		t.Error("IsSelected() should return true after Select()")
	}
	if target.GetSelectedCount() != 1 {
		t.Errorf("GetSelectedCount() = %d, want 1", target.GetSelectedCount())
	}

	// 测试 ToggleSelection
	target.ToggleSelection()
	if target.IsSelected() {
		t.Error("IsSelected() should be false after ToggleSelection()")
	}
	if target.GetSelectedCount() != 0 {
		t.Errorf("GetSelectedCount() = %d, want 0", target.GetSelectedCount())
	}
}

// TestExpandableActionTarget 测试可展开目标
func TestExpandableActionTarget(t *testing.T) {
	target := &mockExpandableTarget{
		mockActionTarget: &mockActionTarget{
			BaseActionTarget: NewBaseActionTarget(ActionExpand, ActionCollapse),
		},
		expanded: false,
	}

	// 测试 Expand
	if !target.Expand() {
		t.Error("Expand() should return true")
	}
	if !target.IsExpanded() {
		t.Error("IsExpanded() should return true after Expand()")
	}

	// 测试 Collapse
	if !target.Collapse() {
		t.Error("Collapse() should return true")
	}
	if target.IsExpanded() {
		t.Error("IsExpanded() should return false after Collapse()")
	}

	// 测试 Toggle
	target.Toggle()
	if !target.IsExpanded() {
		t.Error("IsExpanded() should be true after Toggle()")
	}
}

// ============================================================================
// 辅助函数测试
// ============================================================================

// TestHandleActionWithFallback 测试带回退的处理
func TestHandleActionWithFallback(t *testing.T) {
	target := &mockActionTarget{
		BaseActionTarget: NewBaseActionTarget(ActionClick),
		handleFunc: func(action *Action) bool {
			// 只处理 Click，不处理其他
			return action.Type == ActionClick
		},
	}

	clickAction := NewAction(ActionClick)
	navigateAction := NewAction(ActionNavigateDown)

	// 目标能处理的 Action
	called := false
	fallback := func(action *Action) bool {
		called = true
		return false
	}

	if !HandleActionWithFallback(target, clickAction, fallback) {
		t.Error("HandleActionWithFallback() should return true when target handles action")
	}
	if called {
		t.Error("Fallback should not be called when target handles action")
	}

	// 目标不能处理的 Action，但 fallback 也不处理
	called = false
	if HandleActionWithFallback(target, navigateAction, fallback) {
		t.Error("HandleActionWithFallback() should return false when neither handles action")
	}
	if !called {
		t.Error("Fallback should still be called even when it returns false")
	}
}

// TestGetActionTargets 测试提取 ActionTarget
func TestGetActionTargets(t *testing.T) {
	// 创建混合列表
	target1 := &mockActionTarget{BaseActionTarget: NewBaseActionTarget(ActionClick)}
	target2 := &struct{ string }{} // 不是 ActionTarget
	target3 := &mockActionTarget{BaseActionTarget: NewBaseActionTarget(ActionNavigateDown)}

	nodes := []interface{}{target1, target2, target3}

	targets := GetActionTargets(nodes)
	if len(targets) != 2 {
		t.Errorf("GetActionTargets() returned %d targets, want 2", len(targets))
	}
}

// TestFilterActionTargets 测试过滤 ActionTarget
func TestFilterActionTargets(t *testing.T) {
	target1 := NewBaseActionTarget(ActionClick, ActionHover)
	target2 := NewBaseActionTarget(ActionClick, ActionNavigateDown)
	target3 := NewBaseActionTarget(ActionNavigateDown)

	targets := []ActionTarget{target1, target2, target3}

	// 过滤支持 Click 的目标
	filtered := FilterActionTargets(targets, ActionClick)
	if len(filtered) != 2 {
		t.Errorf("FilterActionTargets() returned %d targets, want 2", len(filtered))
	}

	// 过滤支持 NavigateDown 的目标
	filtered = FilterActionTargets(targets, ActionNavigateDown)
	if len(filtered) != 2 {
		t.Errorf("FilterActionTargets() returned %d targets, want 2", len(filtered))
	}
}

// TestFindActionTarget 测试查找 ActionTarget
func TestFindActionTarget(t *testing.T) {
	target1 := NewBaseActionTarget(ActionClick)
	target2 := NewBaseActionTarget(ActionNavigateDown)

	targets := []ActionTarget{target1, target2}

	// 查找支持 Click 的目标
	found := FindActionTarget(targets, ActionClick)
	if found == nil {
		t.Error("FindActionTarget() should find a target for Click")
	}
	if found != target1 {
		t.Error("FindActionTarget() should return first matching target")
	}

	// 查找不支持的 Action
	found = FindActionTarget(targets, ActionScroll)
	if found != nil {
		t.Error("FindActionTarget() should return nil for unsupported action")
	}
}

// TestDispatchActionToTargets 测试分发 Action
func TestDispatchActionToTargets(t *testing.T) {
	target1 := &mockActionTarget{
		BaseActionTarget: NewBaseActionTarget(ActionClick),
		handleFunc: func(action *Action) bool {
			return action.Type == ActionClick
		},
	}
	target2 := &mockActionTarget{
		BaseActionTarget: NewBaseActionTarget(ActionNavigateDown),
		handleFunc: func(action *Action) bool {
			return action.Type == ActionNavigateDown
		},
	}

	clickAction := NewAction(ActionClick)
	navigateAction := NewAction(ActionNavigateDown)
	scrollAction := NewAction(ActionScroll)

	// 测试 Click 被第一个目标处理
	if !DispatchActionToTargets(clickAction, target1, target2) {
		t.Error("DispatchActionToTargets() should return true for Click")
	}

	// 测试 NavigateDown 被第二个目标处理
	if !DispatchActionToTargets(navigateAction, target1, target2) {
		t.Error("DispatchActionToTargets() should return true for NavigateDown")
	}

	// 测试不支持的 Action（两个目标都不处理）
	if DispatchActionToTargets(scrollAction, target1, target2) {
		t.Error("DispatchActionToTargets() should return false when no target handles action")
	}
}

// TestDispatchActionToTargetsWithFallback 测试带回退的分发
func TestDispatchActionToTargetsWithFallback(t *testing.T) {
	target := &mockActionTarget{
		BaseActionTarget: NewBaseActionTarget(ActionClick),
		handleFunc: func(action *Action) bool {
			return action.Type == ActionClick
		},
	}
	clickAction := NewAction(ActionClick)
	navigateAction := NewAction(ActionNavigateDown)

	fallbackCalled := false
	fallback := func(action *Action) bool {
		fallbackCalled = true
		return action.Type == ActionNavigateDown
	}

	// 目标能处理 Click
	if !DispatchActionToTargetsWithFallback(clickAction, fallback, target) {
		t.Error("Should return true when target handles action")
	}
	if fallbackCalled {
		t.Error("Fallback should not be called when target handles action")
	}

	// 目标不能处理 NavigateDown，但 fallback 可以
	fallbackCalled = false
	if !DispatchActionToTargetsWithFallback(navigateAction, fallback, target) {
		t.Error("Should return true when fallback handles action")
	}
	if !fallbackCalled {
		t.Error("Fallback should be called when target doesn't handle action")
	}
}

// ============================================================================
// BaseActionTarget 测试
// ============================================================================

// TestBaseActionTarget_NewBaseActionTarget 测试创建基础目标
func TestBaseActionTarget_NewBaseActionTarget(t *testing.T) {
	actions := []ActionType{ActionClick, ActionHover}
	target := NewBaseActionTarget(actions...)

	if target.HandleAction(NewAction(ActionClick)) {
		t.Error("HandleAction() should return false for BaseActionTarget")
	}

	supported := target.GetSupportedActions()
	if len(supported) != len(actions) {
		t.Errorf("GetSupportedActions() returned %d actions, want %d", len(supported), len(actions))
	}
}

// TestBaseActionTarget_AddSupportedActions 测试添加支持的 Actions
func TestBaseActionTarget_AddSupportedActions(t *testing.T) {
	target := NewBaseActionTarget(ActionClick)
	target.AddSupportedActions(ActionHover, ActionScroll)

	supported := target.GetSupportedActions()
	expectedCount := 3 // Click + Hover + Scroll
	if len(supported) != expectedCount {
		t.Errorf("GetSupportedActions() returned %d actions, want %d", len(supported), expectedCount)
	}
}

// ============================================================================
// CompositeActionTarget 测试
// ============================================================================

// TestCompositeActionTarget_HandleAction 测试组合目标
func TestCompositeActionTarget(t *testing.T) {
	target1 := &mockActionTarget{
		BaseActionTarget: NewBaseActionTarget(ActionClick),
		handleFunc: func(action *Action) bool {
			return action.Type == ActionClick && action.TargetID == 11111
		},
	}
	target2 := &mockActionTarget{
		BaseActionTarget: NewBaseActionTarget(ActionClick),
		handleFunc: func(action *Action) bool {
			return action.Type == ActionClick && action.TargetID == 22222
		},
	}

	composite := NewCompositeActionTarget(target1, target2)

	// 测试第一个目标处理
	action1 := NewActionFromMouse(ActionClick, 11111, 10, 20)
	if !composite.HandleAction(action1) {
		t.Error("HandleAction() should return true for first target")
	}

	// 测试第二个目标处理
	action2 := NewActionFromMouse(ActionClick, 22222, 10, 20)
	if !composite.HandleAction(action2) {
		t.Error("HandleAction() should return true for second target")
	}

	// 测试两个目标都不处理
	action3 := NewAction(ActionNavigateDown)
	if composite.HandleAction(action3) {
		t.Error("HandleAction() should return false when no target handles action")
	}
}

// TestCompositeActionTarget_GetSupportedActions 测试组合目标的支持列表
func TestCompositeActionTarget_GetSupportedActions(t *testing.T) {
	target1 := NewBaseActionTarget(ActionClick)
	target2 := NewBaseActionTarget(ActionClick, ActionHover)
	target3 := NewBaseActionTarget(ActionNavigateDown)

	composite := NewCompositeActionTarget(target1, target2, target3)

	supported := composite.GetSupportedActions()

	// 应该去重：Click, Hover, NavigateDown
	expectedCount := 3
	if len(supported) != expectedCount {
		t.Errorf("GetSupportedActions() returned %d actions, want %d", len(supported), expectedCount)
	}
}

// TestCompositeActionTarget_AddTarget 测试添加子目标
func TestCompositeActionTarget_AddTarget(t *testing.T) {
	target1 := NewBaseActionTarget(ActionClick)
	target2 := NewBaseActionTarget(ActionHover)

	composite := NewCompositeActionTarget(target1)

	initialCount := len(composite.GetSupportedActions())
	composite.AddTarget(target2)

	newCount := len(composite.GetSupportedActions())
	if newCount != initialCount+1 {
		t.Errorf("AddTarget() should increase supported actions count, got %d, want %d", newCount, initialCount+1)
	}
}

// ============================================================================
// ActionTargetAdapter 测试
// ============================================================================

// TestActionTargetAdapter_HandleAction 测试适配器
func TestActionTargetAdapter_HandleAction(t *testing.T) {
	called := false
	handler := func(action *Action) bool {
		called = true
		return action.Type == ActionClick
	}

	adapter := NewActionTargetAdapter(
		[]ActionType{ActionClick},
		handler,
	)

	// 测试支持的 Action
	clickAction := NewAction(ActionClick)
	if !adapter.HandleAction(clickAction) {
		t.Error("HandleAction() should return true for Click")
	}
	if !called {
		t.Error("Handler should be called for Click")
	}

	// 测试不支持的 Action
	navigateAction := NewAction(ActionNavigateDown)
	if adapter.HandleAction(navigateAction) {
		t.Error("HandleAction() should return false for NavigateDown")
	}
}

// ============================================================================
// 调试和诊断测试
// ============================================================================

// TestGetActionTargetInfo 测试获取目标信息
func TestGetActionTargetInfo(t *testing.T) {
	target := NewBaseActionTarget(ActionClick, ActionNavigateDown)

	info := GetActionTargetInfo(target)

	if info.Target != target {
		t.Error("Target field should be set")
	}

	if len(info.SupportedActions) != 2 {
		t.Errorf("SupportedActions count = %d, want 2", len(info.SupportedActions))
	}

	if !info.CanHandleClick {
		t.Error("CanHandleClick should be true")
	}

	if !info.CanHandleNavigate {
		t.Error("CanHandleNavigate should be true")
	}
}

// TestDumpActionTargets 测试导出目标信息
func TestDumpActionTargets(t *testing.T) {
	target1 := NewBaseActionTarget(ActionClick)
	target2 := NewBaseActionTarget(ActionNavigateDown)

	targets := []ActionTarget{target1, target2}

	dump := DumpActionTargets(targets)
	if dump == "" {
		t.Error("DumpActionTargets() should not return empty string")
	}

	// 验证包含关键信息
	if len(dump) < 20 {
		t.Errorf("Dump string too short: %q", dump)
	}
}
