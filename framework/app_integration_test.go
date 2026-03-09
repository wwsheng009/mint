package framework

import (
	"context"
	"testing"
	"time"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	irender "github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/core"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/render"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/cursor"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/textarea"
)

// TestApp_Throttler 测试 App 的节流器集成
func TestApp_Throttler(t *testing.T) {
	app := NewApp()

	// 测试默认帧率
	if fps := app.throttler.FPS(); fps != 60 {
		t.Errorf("expected FPS 60, got %d", fps)
	}

	// 测试设置帧率
	app.SetFPS(30)
	if fps := app.FPS(); fps != 30 {
		t.Errorf("expected FPS 30, got %d", fps)
	}

	// 测试统计信息
	stats := app.GetRenderStats()
	if stats.TargetFPS != 30 {
		t.Errorf("expected TargetFPS 30, got %d", stats.TargetFPS)
	}
}

// TestApp_ContextManager 测试 App 的上下文管理器集成
func TestApp_ContextManager(t *testing.T) {
	app := NewApp()

	// 测试获取上下文
	ctx := app.Context()
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	// 测试上下文取消
	app.Shutdown(100 * time.Millisecond)
	select {
	case <-ctx.Done():
		// 上下文已取消，预期行为
	case <-time.After(200 * time.Millisecond):
		t.Fatal("context should be cancelled after Shutdown")
	}
}

// TestApp_Recovery 测试 App 的 panic 恢复集成
func TestApp_Recovery(t *testing.T) {
	app := NewApp()

	// 启用恢复
	app.EnableRecovery()

	if app.recovery == nil {
		t.Fatal("expected recovery manager to be initialized")
	}

	// 测试添加处理器 - 创建一个简单的测试处理器
	app.AddPanicHandler(&testPanicHandler{})
}

// TestApp_EventFilter 测试 App 的事件过滤器集成
func TestApp_EventFilter(t *testing.T) {
	app := NewApp()
	filterCalled := false

	// 设置过滤器
	app.SetEventFilter(func(ev frameworkevent.Event) bool {
		filterCalled = true
		return true
	})

	// 清除过滤器
	app.ClearEventFilter()

	// 设置一个拦截所有事件的过滤器
	app.SetEventFilter(func(ev frameworkevent.Event) bool {
		return false
	})
	_ = filterCalled // 避免未使用变量警告
}

// TestApp_ForceRender 测试强制渲染
func TestApp_ForceRender(t *testing.T) {
	app := NewApp()

	// 强制渲染应该设置 dirty 标记
	app.ForceRender()
	if !app.dirty {
		t.Error("expected dirty to be true after ForceRender")
	}
}

// TestApp_AdaptiveFPS 测试自适应帧率
func TestApp_AdaptiveFPS(t *testing.T) {
	app := NewApp()

	// 启用自适应帧率
	app.EnableAdaptiveFPS(true)

	stats := app.GetRenderStats()
	if !stats.Adaptive {
		t.Error("expected adaptive mode to be enabled")
	}

	// 禁用自适应帧率
	app.EnableAdaptiveFPS(false)

	stats = app.GetRenderStats()
	if stats.Adaptive {
		t.Error("expected adaptive mode to be disabled")
	}
}

// TestApp_GracefulShutdown 测试优雅关闭
func TestApp_GracefulShutdown(t *testing.T) {
	app := NewApp()

	// 获取上下文
	ctx := app.Context()

	// 在另一个 goroutine 中触发关闭
	go func() {
		time.Sleep(50 * time.Millisecond)
		app.Shutdown(100 * time.Millisecond)
	}()

	// 等待上下文取消
	select {
	case <-ctx.Done():
		// 预期行为
	case <-time.After(200 * time.Millisecond):
		t.Fatal("context should be cancelled after Shutdown")
	}
}

// TestThrottler_Behavior 详细测试节流器行为
func TestThrottler_Behavior(t *testing.T) {
	throttler := render.NewThrottler(60)

	// 首次调用应该返回 true
	if !throttler.ShouldRender() {
		t.Error("first render should be allowed")
	}

	// 立即再次调用应该返回 false（间隔太短）
	if throttler.ShouldRender() {
		t.Error("immediate second render should be throttled")
	}

	// 等待超过最小间隔后应该允许渲染
	time.Sleep(20 * time.Millisecond)
	if !throttler.ShouldRender() {
		t.Error("render after interval should be allowed")
	}
}

// TestContextManager_Integration 测试上下文管理器集成
func TestContextManager_Integration(t *testing.T) {
	ctx := context.Background()
	mgr := core.NewContextManager(ctx)

	// 测试上下文传播
	appCtx := mgr.Context()
	if appCtx == nil {
		t.Fatal("expected non-nil context from manager")
	}

	// 测试优雅关闭
	done := make(chan error, 1)
	go func() {
		done <- mgr.Shutdown(100 * time.Millisecond)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("shutdown failed: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("shutdown timed out")
	}
}

func TestApp_HandleTick_DrivesTickableInstances(t *testing.T) {
	app := NewApp()

	caret := cursor.NewInstance(rtui.Props{
		"visible": true,
		"config": cursor.Config{
			Blink:         true,
			BlinkInterval: time.Nanosecond,
		},
	})

	rootFiber := &rtui.Fiber{
		Tag:      "cursor",
		Instance: caret,
	}
	app.root = &mockFiberRootNode{fiberRoot: rootFiber}

	app.dirty = false
	time.Sleep(time.Millisecond)
	app.handleTick()

	if !app.dirty {
		t.Fatal("handleTick should mark app dirty when a tickable instance advances")
	}
}

func TestApp_HandleTick_DrivesFocusedInputCursor(t *testing.T) {
	app := NewApp()
	app.Resize(80, 24)

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return input.NewBuilder().
			Key("test-input").
			Value("abc").
			CursorConfig(cursor.Config{
				Blink:         true,
				BlinkInterval: time.Millisecond,
			}).
			Build()
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	// First render builds the Fiber tree and focusable list.
	app.dirty = true
	app.render()
	// Second render applies focus state onto instances.
	app.dirty = true
	app.render()

	rootFiber := app.getFiberRoot()
	if rootFiber == nil {
		t.Fatal("expected fiber root after rendering")
	}

	foundFocusedInput := false
	rtui.WalkFiberDepthFirst(rootFiber, func(fiber *rtui.Fiber) bool {
		if fiber == nil || fiber.Instance == nil {
			return true
		}
		if inst, ok := fiber.Instance.(*input.Instance); ok && inst.HasFocus() {
			foundFocusedInput = true
		}
		return true
	})
	if !foundFocusedInput {
		t.Fatal("expected focused input instance after second render")
	}

	app.dirty = false
	time.Sleep(2 * time.Millisecond)
	app.handleTick()
	if !app.dirty {
		t.Fatal("handleTick should mark app dirty for focused blinking input")
	}
}

func TestApp_ProcessMsg_LeftArrowMovesInputCursor(t *testing.T) {
	app := NewApp()
	app.Resize(80, 24)

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return input.NewBuilder().
			Key("test-input").
			Value("abc").
			Build()
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	// Build tree, then apply focus state.
	app.dirty = true
	app.render()
	app.dirty = true
	app.render()

	var focusedInput *input.Instance
	rtui.WalkFiberDepthFirst(app.getFiberRoot(), func(fiber *rtui.Fiber) bool {
		if fiber == nil || fiber.Instance == nil {
			return true
		}
		if inst, ok := fiber.Instance.(*input.Instance); ok && inst.HasFocus() {
			focusedInput = inst
		}
		return true
	})
	if focusedInput == nil {
		t.Fatal("expected focused input instance")
	}
	if focusedInput.CursorPos() != 3 {
		t.Fatalf("initial cursor = %d, want 3", focusedInput.CursorPos())
	}

	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyLeft, runtimemsg.Modifiers{}))
	if focusedInput.CursorPos() != 2 {
		t.Fatalf("cursor after left = %d, want 2", focusedInput.CursorPos())
	}
}

func TestApp_ProcessMsg_LeftArrowMovesInputCursor_WithWrapperTag(t *testing.T) {
	app := NewApp()

	inst := input.NewInstance(rtui.Props{
		"value": "abc",
	})
	fiber := &rtui.Fiber{
		Tag:      "component",
		NodeID:   1,
		Instance: inst,
	}

	app.focusManager.UpdateFocusableList([]*rtui.Fiber{fiber})
	if ok := app.focusManager.SetFocusByIndex(0); !ok {
		t.Fatal("SetFocusByIndex(0) should succeed")
	}

	if inst.CursorPos() != 3 {
		t.Fatalf("initial cursor = %d, want 3", inst.CursorPos())
	}

	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyLeft, runtimemsg.Modifiers{}))
	if inst.CursorPos() != 2 {
		t.Fatalf("cursor after left = %d, want 2", inst.CursorPos())
	}
}

func TestApp_ProcessMsg_UpDownMoveTextareaCursor_WithWrapperTag(t *testing.T) {
	app := NewApp()

	inst := textarea.NewInstance(rtui.Props{
		"value": "abcd\nef\nxyz",
	})
	inst.SetCursorPos(3)

	fiber := &rtui.Fiber{
		Tag:      "component",
		NodeID:   1,
		Instance: inst,
	}

	app.focusManager.UpdateFocusableList([]*rtui.Fiber{fiber})
	if ok := app.focusManager.SetFocusByIndex(0); !ok {
		t.Fatal("SetFocusByIndex(0) should succeed")
	}

	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyDown, runtimemsg.Modifiers{}))
	if inst.CursorPos() != 7 {
		t.Fatalf("cursor after down = %d, want 7", inst.CursorPos())
	}

	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyUp, runtimemsg.Modifiers{}))
	if inst.CursorPos() != 3 {
		t.Fatalf("cursor after up = %d, want 3", inst.CursorPos())
	}
}

func TestApp_ProcessMsg_TypedNilMouseTargetFiber_NoPanic(t *testing.T) {
	app := NewApp()

	mouseMsg := runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	var targetFiber *rtui.Fiber
	mouseMsg.TargetFiber = targetFiber

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("processMsg should ignore typed nil TargetFiber, got panic: %v", recovered)
		}
	}()

	app.processMsg(mouseMsg)
}

func TestApp_ProcessMsg_UnhandledNavigationFallsThroughToFocusedFiber(t *testing.T) {
	app := NewApp()

	recorder := &navigationActionRecorder{}
	fiber := &rtui.Fiber{
		Tag:      "menu-popup",
		NodeID:   42,
		Instance: recorder,
	}

	app.focusManager.UpdateFocusableList([]*rtui.Fiber{fiber})
	if ok := app.focusManager.SetFocusByIndex(0); !ok {
		t.Fatal("SetFocusByIndex(0) should succeed")
	}

	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyDown, runtimemsg.Modifiers{}))

	if len(recorder.handled) == 0 || recorder.handled[0] != action.ActionNavigateDown {
		t.Fatalf("handled actions = %#v, want first action %q", recorder.handled, action.ActionNavigateDown)
	}
}

func TestApp_ProcessMsg_UntargetedClickDispatchesActionRouterMiddleware(t *testing.T) {
	app := NewApp()
	middleware := &recordingMiddleware{intercept: true}
	app.AddMiddleware(middleware)

	app.dirty = false
	app.processMsg(runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress))

	if len(middleware.beforeCalls) != 1 {
		t.Fatalf("middleware before calls = %d, want 1", len(middleware.beforeCalls))
	}
	if middleware.beforeCalls[0] != action.ActionClick {
		t.Fatalf("middleware first action = %q, want %q", middleware.beforeCalls[0], action.ActionClick)
	}
	if !app.dirty {
		t.Fatal("app should be marked dirty when middleware intercepts untargeted click")
	}
}

type mockFiberRootNode struct {
	fiberRoot *rtui.Fiber
}

func (n *mockFiberRootNode) ID() string {
	return "mock-fiber-root"
}

func (n *mockFiberRootNode) Type() string {
	return "mock-fiber-root"
}

func (n *mockFiberRootNode) GetFiberRoot() *rtui.Fiber {
	return n.fiberRoot
}

type navigationActionRecorder struct {
	handled []action.ActionType
}

func (m *navigationActionRecorder) Key() string                        { return "" }
func (m *navigationActionRecorder) SetKey(key string)                  {}
func (m *navigationActionRecorder) Init(props rtui.Props)              {}
func (m *navigationActionRecorder) Destroy()                           {}
func (m *navigationActionRecorder) OnMount()                           {}
func (m *navigationActionRecorder) OnUnmount()                         {}
func (m *navigationActionRecorder) SetProps(props rtui.Props) bool     { return false }
func (m *navigationActionRecorder) GetProps() rtui.Props               { return nil }
func (m *navigationActionRecorder) MarkDirty()                         {}
func (m *navigationActionRecorder) IsDirty() bool                      { return false }
func (m *navigationActionRecorder) GetContext() *rtui.ComponentContext { return nil }
func (m *navigationActionRecorder) SetFocus(bool)                      {}
func (m *navigationActionRecorder) HasFocus() bool                     { return true }
func (m *navigationActionRecorder) IsDisabled() bool                   { return false }
func (m *navigationActionRecorder) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}
	m.handled = append(m.handled, act.Type)
	return act.Type == action.ActionNavigateDown
}

type recordingMiddleware struct {
	beforeCalls []action.ActionType
	intercept   bool
}

func (m *recordingMiddleware) Name() string {
	return "recording"
}

func (m *recordingMiddleware) Before(act *action.Action) *action.Action {
	if act == nil {
		return nil
	}
	m.beforeCalls = append(m.beforeCalls, act.Type)
	if m.intercept {
		return nil
	}
	return act
}

func (m *recordingMiddleware) After(act *action.Action, result *action.RouterResult) {}

// testPanicHandler 测试用的 panic 处理器
type testPanicHandler struct{}

func (h *testPanicHandler) HandlePanic(r interface{}, stack []byte) {
	// 测试实现，什么都不做
}
