package framework

import (
	"testing"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// TestApp_InputTracker 测试 InputTracker 集成
func TestApp_InputTracker(t *testing.T) {
	app := NewApp()

	if app.inputTracker == nil {
		t.Fatal("InputTracker should be initialized")
	}
	if app.interactionCtx == nil {
		t.Fatal("InteractionContext should be initialized")
	}

	// 测试 mouse press
	mousePress := runtimemsg.NewMouseMsg(10, 10, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	snapshot := app.msgToSnapshot(mousePress)

	if snapshot == nil {
		t.Fatal("snapshot should not be nil for MousePress")
	}
	if snapshot.MouseX != 10 || snapshot.MouseY != 10 {
		t.Errorf("Expected position (10,10), got (%d,%d)", snapshot.MouseX, snapshot.MouseY)
	}
	if snapshot.MouseButton != runtimemsg.MouseLeft {
		t.Errorf("Expected MouseButton=MouseLeft, got %v", snapshot.MouseButton)
	}

	// 测试 InputTracker Update
	// 首次 Press 会产生 2 个 intent：InputMoveIntent (初始位置) + InputPressIntent
	intents := app.inputTracker.Update(snapshot)
	if len(intents) != 2 {
		t.Errorf("Expected 2 intents on first press (Move + Press), got %d", len(intents))
	}

	// 测试 mouse release
	mouseRelease := runtimemsg.NewMouseMsg(10, 10, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)
	snapshot = app.msgToSnapshot(mouseRelease)

	if snapshot == nil {
		t.Fatal("snapshot should not be nil for MouseRelease")
	}
	if snapshot.MouseButton != runtimemsg.MouseButtonUnknown {
		t.Errorf("Expected MouseButton=Unknown on release, got %v", snapshot.MouseButton)
	}
}

// TestApp_HitTest 测试 hitTest 集成
func TestApp_HitTest(t *testing.T) {
	app := NewApp()

	// 没有 HitMap 时应该返回 0
	if app.hitTest(10, 10) != 0 {
		t.Error("Expected hitTest=0 when hitMap is nil")
	}

	// TODO: 添加 HitMap 后测试实际命中
}

// TestApp_MsgToSnapshot_KeyMsg 测试 KeyMsg 转换
func TestApp_MsgToSnapshot_KeyMsg(t *testing.T) {
	app := NewApp()

	keyMsg := runtimemsg.NewKeyMsg('a', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{})
	snapshot := app.msgToSnapshot(keyMsg)

	if snapshot == nil {
		t.Fatal("snapshot should not be nil for KeyMsg")
	}
	if snapshot.KeyboardKey != 'a' {
		t.Errorf("Expected KeyboardKey='a', got '%c'", snapshot.KeyboardKey)
	}
}

// TestApp_MsgToSnapshot_IgnoreResize 测试忽略其他消息
func TestApp_MsgToSnapshot_IgnoreResize(t *testing.T) {
	app := NewApp()

	resizeMsg := runtimemsg.NewResizeMsg(0, 0, 80, 24)
	snapshot := app.msgToSnapshot(resizeMsg)

	if snapshot != nil {
		t.Error("ResizeMsg should not generate a snapshot")
	}
}

func TestApp_ResizeSameSizeDoesNotMarkDirty(t *testing.T) {
	app := NewApp()

	app.Resize(80, 24)
	app.dirty = false

	app.Resize(80, 24)

	if app.dirty {
		t.Fatal("expected same-size resize to leave app clean")
	}
	if app.terminalWidth != 80 || app.terminalHeight != 24 {
		t.Fatalf("terminal size = %dx%d, want 80x24", app.terminalWidth, app.terminalHeight)
	}
}

func TestApp_ProcessMsgResizeRoutesEvent(t *testing.T) {
	app := NewApp()

	var got *frameworkevent.ResizeEvent
	app.OnEvent(frameworkevent.EventResize, frameworkevent.EventHandlerFunc(func(ev frameworkevent.Event) bool {
		resizeEv, ok := ev.(*frameworkevent.ResizeEvent)
		if !ok {
			t.Fatalf("resize event type = %T, want *ResizeEvent", ev)
		}
		copy := *resizeEv
		got = &copy
		return false
	}))

	app.processMsg(runtimemsg.NewResizeMsg(80, 24, 120, 36))

	if got == nil {
		t.Fatal("expected resize event to be routed")
	}
	if got.OldWidth != 80 || got.OldHeight != 24 || got.NewWidth != 120 || got.NewHeight != 36 {
		t.Fatalf("resize event = %+v, want old 80x24 and new 120x36", got)
	}
	if app.terminalWidth != 120 || app.terminalHeight != 36 {
		t.Fatalf("terminal size = %dx%d, want 120x36", app.terminalWidth, app.terminalHeight)
	}
}

// TestApp_InteractionContext_Registration 测试 InteractionContext 组件注册
func TestApp_InteractionContext_Registration(t *testing.T) {
	app := NewApp()

	// 创建一个 mock PressedResetHandler
	mockHandler := &mockPressedResetHandler{
		resetCalled: false,
	}

	// 直接注册到 InteractionContext
	app.interactionCtx.RegisterInstance(1, mockHandler)

	// 验证组件已注册
	if len(app.interactionCtx.Instances) != 1 {
		t.Errorf("Expected 1 registered instance, got %d", len(app.interactionCtx.Instances))
	}

	// 模拟键盘输入触发 InputKeyboardIntent
	snapshot := app.msgToSnapshot(runtimemsg.NewKeyMsg('a', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{}))
	intents := app.inputTracker.Update(snapshot)

	// 更新 InteractionContext（应该触发 ResetPressed）
	app.interactionCtx.Update(intents, app.hitTest)

	// 验证 ResetPressed() 被调用
	if !mockHandler.resetCalled {
		t.Error("Expected ResetPressed() to be called, but was not")
	}
}

// TestApp_InteractionContext_MultipleInstances 测试多个组件实例
func TestApp_InteractionContext_MultipleInstances(t *testing.T) {
	app := NewApp()

	// 创建 3 个 handler
	mockHandler1 := &mockPressedResetHandler{resetCalled: false}
	mockHandler2 := &mockPressedResetHandler{resetCalled: false}
	mockHandler3 := &mockPressedResetHandler{resetCalled: false}

	// 注册到 InteractionContext
	app.interactionCtx.RegisterInstance(1, mockHandler1)
	app.interactionCtx.RegisterInstance(2, mockHandler2)
	app.interactionCtx.RegisterInstance(3, mockHandler3)

	// 模拟键盘输入触发
	snapshot := app.msgToSnapshot(runtimemsg.NewKeyMsg('x', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{}))
	intents := app.inputTracker.Update(snapshot)
	app.interactionCtx.Update(intents, app.hitTest)

	// 验证所有 handler 的 ResetPressed() 都被调用
	if !mockHandler1.resetCalled || !mockHandler2.resetCalled || !mockHandler3.resetCalled {
		t.Error("Expected all ResetPressed() to be called")
	}
}

// mockPressedResetHandler 是一个 mock 实现，用于测试
type mockPressedResetHandler struct {
	resetCalled bool
}

// ResetPressed 实现 PressedResetHandler 接口
func (m *mockPressedResetHandler) ResetPressed() {
	m.resetCalled = true
}
