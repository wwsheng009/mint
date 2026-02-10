package action

import (
	"testing"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
)

// ============================================================================
// InputProcessor 测试
// ============================================================================

// TestInputProcessor_NewInputProcessor 测试创建 InputProcessor
func TestInputProcessor_NewInputProcessor(t *testing.T) {
	processor := NewInputProcessor()

	if processor == nil {
		t.Fatal("NewInputProcessor() returned nil")
	}

	if processor.keyMap != nil {
		t.Error("NewInputProcessor() should have nil keyMap by default")
	}
}

// TestInputProcessor_SetKeyMap 测试设置 KeyMap
func TestInputProcessor_SetKeyMap(t *testing.T) {
	processor := NewInputProcessor()
	keyMap := NewKeyMap()

	processor.SetKeyMap(keyMap)

	if processor.GetKeyMap() != keyMap {
		t.Error("SetKeyMap() failed")
	}
}

// TestInputProcessor_ProcessKeyEvent_Navigation 测试导航键转换
func TestInputProcessor_ProcessKeyEvent_Navigation(t *testing.T) {
	processor := NewInputProcessor()

	tests := []struct {
		name        string
		special     frameworkevent.SpecialKey
		expected    ActionType
	}{
		{"Up", frameworkevent.KeyUp, ActionNavigateUp},
		{"Down", frameworkevent.KeyDown, ActionNavigateDown},
		{"Left", frameworkevent.KeyLeft, ActionNavigateLeft},
		{"Right", frameworkevent.KeyRight, ActionNavigateRight},
		{"PageUp", frameworkevent.KeyPageUp, ActionNavigatePageUp},
		{"PageDown", frameworkevent.KeyPageDown, ActionNavigatePageDown},
		{"Home", frameworkevent.KeyHome, ActionNavigateHome},
		{"End", frameworkevent.KeyEnd, ActionNavigateEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &frameworkevent.KeyEvent{
				Special: tt.special,
			}

			action := processor.Process(ev)
			if action == nil {
				t.Fatalf("Process() returned nil for %v", tt.special)
			}

			if action.Type != tt.expected {
				t.Errorf("Process() = %v, want %v", action.Type, tt.expected)
			}

			if action.Source != "keyboard" {
				t.Errorf("Source = %q, want %q", action.Source, "keyboard")
			}
		})
	}
}

// TestInputProcessor_ProcessKeyEvent_Editing 测试编辑键转换
func TestInputProcessor_ProcessKeyEvent_Editing(t *testing.T) {
	processor := NewInputProcessor()

	tests := []struct {
		name     string
		special  frameworkevent.SpecialKey
		expected ActionType
	}{
		{"Enter", frameworkevent.KeyEnter, ActionEnter},
		{"Tab", frameworkevent.KeyTab, ActionNavigateNext},
		{"Backspace", frameworkevent.KeyBackspace, ActionBackspace},
		{"Delete", frameworkevent.KeyDelete, ActionDeleteChar},
		{"Escape", frameworkevent.KeyEscape, ActionCancel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &frameworkevent.KeyEvent{
				Special: tt.special,
			}

			action := processor.Process(ev)
			if action == nil {
				t.Fatalf("Process() returned nil for %v", tt.special)
			}

			if action.Type != tt.expected {
				t.Errorf("Process() = %v, want %v", action.Type, tt.expected)
			}
		})
	}
}

// TestInputProcessor_ProcessKeyEvent_FunctionKeys 测试功能键转换
func TestInputProcessor_ProcessKeyEvent_FunctionKeys(t *testing.T) {
	processor := NewInputProcessor()

	tests := []struct {
		name     string
		special  frameworkevent.SpecialKey
		expected ActionType
	}{
		{"F1", frameworkevent.KeyF1, ActionInspect},
		{"F5", frameworkevent.KeyF5, ActionRefresh},
		{"F10", frameworkevent.KeyF10, ActionQuit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &frameworkevent.KeyEvent{
				Special: tt.special,
			}

			action := processor.Process(ev)
			if action == nil {
				t.Fatalf("Process() returned nil for %v", tt.special)
			}

			if action.Type != tt.expected {
				t.Errorf("Process() = %v, want %v", action.Type, tt.expected)
			}
		})
	}
}

// TestInputProcessor_ProcessKeyEvent_TextInput 测试文本输入转换
func TestInputProcessor_ProcessKeyEvent_TextInput(t *testing.T) {
	processor := NewInputProcessor()

	tests := []rune{'a', 'Z', '0', ' ', '.', '?'}

	for _, r := range tests {
		t.Run(string(r), func(t *testing.T) {
			ev := &frameworkevent.KeyEvent{
				Key: frameworkevent.Key{Rune: r},
			}

			action := processor.Process(ev)
			if action == nil {
				t.Fatalf("Process() returned nil for rune %v", r)
			}

			if action.Type != ActionInputText {
				t.Errorf("Type = %v, want %v", action.Type, ActionInputText)
			}

			payload, ok := action.GetPayloadString()
			if !ok {
				t.Fatal("GetPayloadString() returned ok=false")
			}

			if payload != string(r) {
				t.Errorf("Payload = %q, want %q", payload, string(r))
			}
		})
	}
}

// TestInputProcessor_ProcessKeyEvent_Modifiers 测试修饰键组合
func TestInputProcessor_ProcessKeyEvent_Modifiers(t *testing.T) {
	processor := NewInputProcessor()

	tests := []struct {
		name        string
		key         rune
		ctrl, alt   bool
		expected    ActionType
	}{
		{"Ctrl+C", 'c', true, false, ActionCopy},
		{"Ctrl+V", 'v', true, false, ActionPaste},
		{"Ctrl+X", 'x', true, false, ActionCut},
		{"Ctrl+F", 'f', true, false, ActionSearch},
		{"Ctrl+Q", 'q', true, false, ActionQuit},
		{"Ctrl+S", 's', true, false, ActionSubmit},
		{"Ctrl+A", 'a', true, false, ActionNavigateHome},
		{"Ctrl+E", 'e', true, false, ActionNavigateEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &frameworkevent.KeyEvent{
				Key: frameworkevent.Key{
					Rune: tt.key,
					Ctrl: tt.ctrl,
					Alt:  tt.alt,
				},
			}

			action := processor.Process(ev)
			if action == nil {
				t.Fatalf("Process() returned nil for %v", tt.name)
			}

			if action.Type != tt.expected {
				t.Errorf("Process() = %v, want %v", action.Type, tt.expected)
			}
		})
	}
}

// TestInputProcessor_ProcessMouseEvent_Click 测试鼠标点击转换
func TestInputProcessor_ProcessMouseEvent_Click(t *testing.T) {
	processor := NewInputProcessor()

	tests := []struct {
		name        string
		button      frameworkevent.MouseButton
		expected    ActionType
	}{
		{"LeftClick", frameworkevent.MouseLeft, ActionClick},
		{"RightClick", frameworkevent.MouseRight, ActionRightClick},
		{"MiddleClick", frameworkevent.MouseMiddle, ActionMiddleClick},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &frameworkevent.MouseEvent{
				Action:   runtimeevent.MouseActionPress, // 使用正确的常量
				Button:   tt.button,
				TargetID: "button-1",
				LocalX:   10,
				LocalY:   20,
			}

			action := processor.Process(ev)
			if action == nil {
				t.Fatal("Process() returned nil")
			}

			if action.Type != tt.expected {
				t.Errorf("Type = %v, want %v", action.Type, tt.expected)
			}

			if action.TargetID != "button-1" {
				t.Errorf("TargetID = %q, want %q", action.TargetID, "button-1")
			}

			x, y, ok := action.GetPayloadPoint()
			if !ok {
				t.Fatal("GetPayloadPoint() returned ok=false")
			}

			if x != 10 || y != 20 {
				t.Errorf("Payload = (%d,%d), want (10,20)", x, y)
			}

			if action.Source != "mouse" {
				t.Errorf("Source = %q, want %q", action.Source, "mouse")
			}
		})
	}
}

// TestInputProcessor_ProcessMouseEvent_Hover 测试鼠标悬停转换
func TestInputProcessor_ProcessMouseEvent_Hover(t *testing.T) {
	processor := NewInputProcessor()

	ev := &frameworkevent.MouseEvent{
		Action:   runtimeevent.MouseActionMove, // 使用正确的常量
		TargetID: "list-item",
		LocalX:   5,
		LocalY:   15,
	}

	action := processor.Process(ev)
	if action == nil {
		t.Fatal("Process() returned nil")
	}

	if action.Type != ActionHover {
		t.Errorf("Type = %v, want %v", action.Type, ActionHover)
	}

	if action.TargetID != "list-item" {
		t.Errorf("TargetID = %q, want %q", action.TargetID, "list-item")
	}
}

// TestInputProcessor_ProcessMouseEvent_Scroll 测试滚轮滚动转换
func TestInputProcessor_ProcessMouseEvent_Scroll(t *testing.T) {
	processor := NewInputProcessor()

	tests := []struct {
		name        string
		delta       int
	}{
		{"ScrollUp", 1},
		{"ScrollDown", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &frameworkevent.MouseEvent{
				Action:   runtimeevent.MouseActionWheel, // 使用正确的常量
				TargetID: "list-1",
				Delta:    tt.delta,
			}

			action := processor.Process(ev)
			if action == nil {
				t.Fatal("Process() returned nil")
			}

			if action.Type != ActionScroll {
				t.Errorf("Type = %v, want %v", action.Type, ActionScroll)
			}

			delta, ok := action.GetPayloadInt()
			if !ok {
				t.Fatal("GetPayloadInt() returned ok=false")
			}

			if delta != tt.delta {
				t.Errorf("Delta = %d, want %d", delta, tt.delta)
			}
		})
	}
}

// TestInputProcessor_ProcessMouseEvent_NoTarget 测试没有目标的鼠标事件
func TestInputProcessor_ProcessMouseEvent_NoTarget(t *testing.T) {
	processor := NewInputProcessor()

	ev := &frameworkevent.MouseEvent{
		Action:   runtimeevent.MouseActionPress, // 使用正确的常量
		Button:   frameworkevent.MouseLeft,
		TargetID: "", // 空目标
		LocalX:   10,
		LocalY:   20,
	}

	action := processor.Process(ev)
	if action != nil {
		t.Errorf("Process() should return nil for event without target, got %v", action)
	}
}

// TestInputProcessor_ProcessBatch 测试批量处理
func TestInputProcessor_ProcessBatch(t *testing.T) {
	processor := NewInputProcessor()

	events := []frameworkevent.Event{
		&frameworkevent.KeyEvent{Special: frameworkevent.KeyUp},
		&frameworkevent.KeyEvent{Special: frameworkevent.KeyDown},
		&frameworkevent.KeyEvent{Special: frameworkevent.KeyLeft},
		&frameworkevent.KeyEvent{Special: frameworkevent.KeyRight},
	}

	actions := processor.ProcessBatch(events)

	if len(actions) != len(events) {
		t.Errorf("ProcessBatch() returned %d actions, want %d", len(actions), len(events))
	}

	for i, action := range actions {
		if action == nil {
			t.Errorf("actions[%d] is nil", i)
		}
	}
}

// ============================================================================
// KeyMap 测试
// ============================================================================

// TestKeyMap_NewKeyMap 测试创建 KeyMap
func TestKeyMap_NewKeyMap(t *testing.T) {
	km := NewKeyMap()

	if km == nil {
		t.Fatal("NewKeyMap() returned nil")
	}

	if km.Size() != 0 {
		t.Errorf("NewKeyMap() has size %d, want 0", km.Size())
	}
}

// TestKeyMap_Bind 测试绑定按键
func TestKeyMap_Bind(t *testing.T) {
	km := NewKeyMap()
	action := NewAction(ActionCopy)

	err := km.Bind("ctrl+c", action)
	if err != nil {
		t.Fatalf("Bind() returned error: %v", err)
	}

	if km.Size() != 1 {
		t.Errorf("Size() = %d, want 1", km.Size())
	}

	// 验证查找
	foundAction := km.Lookup("ctrl+c")
	if foundAction == nil {
		t.Fatal("Lookup() returned nil")
	}

	if foundAction.Type != ActionCopy {
		t.Errorf("Lookup() = %v, want %v", foundAction.Type, ActionCopy)
	}
}

// TestKeyMap_Bind_SimpleKeys 测试绑定简单按键
func TestKeyMap_Bind_SimpleKeys(t *testing.T) {
	km := NewKeyMap()

	// 特殊键
	err := km.Bind("enter", NewAction(ActionEnter))
	if err != nil {
		t.Fatalf("Bind(enter) failed: %v", err)
	}

	err = km.Bind("space", NewAction(ActionToggle))
	if err != nil {
		t.Fatalf("Bind(space) failed: %v", err)
	}

	// 字符键
	err = km.Bind("a", NewAction(ActionInputText))
	if err != nil {
		t.Fatalf("Bind(a) failed: %v", err)
	}

	if km.Size() != 3 {
		t.Errorf("Size() = %d, want 3", km.Size())
	}
}

// TestKeyMap_Bind_Modifiers 测试绑定修饰键组合
func TestKeyMap_Bind_Modifiers(t *testing.T) {
	km := NewKeyMap()

	tests := []struct {
		keySpec    string
		actionType ActionType
	}{
		{"ctrl+c", ActionCopy},
		{"ctrl+shift+c", ActionCopy}, // 复杂修饰键
		{"alt+f", ActionSearch},
		{"ctrl+alt+delete", ActionCancel},
	}

	for _, tt := range tests {
		t.Run(tt.keySpec, func(t *testing.T) {
			action := NewAction(tt.actionType)
			err := km.Bind(tt.keySpec, action)
			if err != nil {
				t.Fatalf("Bind(%s) failed: %v", tt.keySpec, err)
			}

			found := km.Lookup(tt.keySpec)
			if found == nil {
				t.Fatalf("Lookup(%s) returned nil", tt.keySpec)
			}

			if found.Type != tt.actionType {
				t.Errorf("Lookup() = %v, want %v", found.Type, tt.actionType)
			}
		})
	}
}

// TestKeyMap_Unbind 测试解除绑定
func TestKeyMap_Unbind(t *testing.T) {
	km := NewKeyMap()
	action := NewAction(ActionCopy)

	km.Bind("ctrl+c", action)

	if km.Size() != 1 {
		t.Fatalf("Size() = %d, want 1", km.Size())
	}

	err := km.Unbind("ctrl+c")
	if err != nil {
		t.Fatalf("Unbind() returned error: %v", err)
	}

	if km.Size() != 0 {
		t.Errorf("Size() = %d, want 0", km.Size())
	}

	if found := km.Lookup("ctrl+c"); found != nil {
		t.Errorf("Lookup() after Unbind() should return nil, got %v", found)
	}
}

// TestKeyMap_Context 测试上下文相关映射
func TestKeyMap_Context(t *testing.T) {
	km := NewKeyMap()

	// 全局映射
	globalAction := NewAction(ActionCopy)
	km.Bind("ctrl+c", globalAction)

	// 上下文特定映射
	inputAction := NewAction(ActionPaste)
	err := km.BindWithContext("input", "ctrl+v", inputAction)
	if err != nil {
		t.Fatalf("BindWithContext() failed: %v", err)
	}

	// 设置上下文
	km.SetCurrentContext("input")

	// 查找应该返回上下文特定映射
	ev := &frameworkevent.KeyEvent{
		Key: frameworkevent.Key{
			Rune: 'v',
			Ctrl: true,
		},
	}

	foundAction := km.LookupKeyEvent(ev)
	if foundAction == nil {
		t.Fatal("LookupKeyEvent() returned nil")
	}

	if foundAction.Type != ActionPaste {
		t.Errorf("LookupKeyEvent() = %v, want %v", foundAction.Type, ActionPaste)
	}
}

// TestKeyMap_ContextStack 测试上下文栈
func TestKeyMap_ContextStack(t *testing.T) {
	km := NewKeyMap()

	// 推入上下文
	km.PushContext("root")
	km.PushContext("child")

	current := km.GetCurrentContext()
	if current != "child" {
		t.Errorf("GetCurrentContext() = %q, want %q", current, "child")
	}

	// 弹出上下文
	km.PopContext()

	current = km.GetCurrentContext()
	if current != "root" {
		t.Errorf("GetCurrentContext() after Pop() = %q, want %q", current, "root")
	}
}

// TestKeyMap_LookupKeyEvent 测试从 KeyEvent 查找
func TestKeyMap_LookupKeyEvent(t *testing.T) {
	km := NewKeyMap()
	action := NewAction(ActionCopy)

	km.Bind("ctrl+c", action)

	ev := &frameworkevent.KeyEvent{
		Key: frameworkevent.Key{
			Rune: 'c',
			Ctrl: true,
		},
	}

	foundAction := km.LookupKeyEvent(ev)
	if foundAction == nil {
		t.Fatal("LookupKeyEvent() returned nil")
	}

	if foundAction.Type != ActionCopy {
		t.Errorf("LookupKeyEvent() = %v, want %v", foundAction.Type, ActionCopy)
	}

	if foundAction.Source != "keymap" {
		t.Errorf("Source = %q, want %q", foundAction.Source, "keymap")
	}
}

// TestKeyMap_DefaultKeyMap 测试默认键盘映射
func TestKeyMap_DefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	if km.Size() == 0 {
		t.Error("DefaultKeyMap() should have bindings")
	}

	// 测试一些默认映射
	tests := []struct {
		keySpec string
		actionType ActionType
	}{
		{"up", ActionNavigateUp},
		{"down", ActionNavigateDown},
		{"enter", ActionEnter},
		{"tab", ActionNavigateNext},
		{"ctrl+c", ActionCopy},
		{"f1", ActionInspect},
	}

	for _, tt := range tests {
		t.Run(tt.keySpec, func(t *testing.T) {
			action := km.Lookup(tt.keySpec)
			if action == nil {
				t.Fatalf("Lookup(%s) returned nil", tt.keySpec)
			}

			if action.Type != tt.actionType {
				t.Errorf("Lookup(%s) = %v, want %v", tt.keySpec, action.Type, tt.actionType)
			}
		})
	}
}

// TestKeyMap_Clear 测试清空映射
func TestKeyMap_Clear(t *testing.T) {
	km := DefaultKeyMap()

	if km.Size() == 0 {
		t.Fatal("DefaultKeyMap() should have bindings before Clear()")
	}

	km.Clear()

	if km.Size() != 0 {
		t.Errorf("Size() after Clear() = %d, want 0", km.Size())
	}

	currentContext := km.GetCurrentContext()
	if currentContext != "" {
		t.Errorf("GetCurrentContext() after Clear() = %q, want %q", currentContext, "")
	}
}
