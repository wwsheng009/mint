package framework

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	irender "github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/core"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/render"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/cursor"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/notification"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
	"github.com/wwsheng009/mint/ui/components/textarea"
	"github.com/wwsheng009/mint/ui/components/toast"
	"github.com/wwsheng009/mint/ui/components/tooltip"
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
	if app.tickInterval != time.Second/30 {
		t.Errorf("expected tickInterval %v, got %v", time.Second/30, app.tickInterval)
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

func TestApp_HandleTick_ExpiresTimedNotification(t *testing.T) {
	app := NewApp()

	inst := notification.NewBuilder("done").
		Title("Saved").
		Duration(time.Millisecond).
		BuildInstance()

	rootFiber := &rtui.Fiber{
		Tag:      "notification",
		Instance: inst,
	}
	app.root = &mockFiberRootNode{fiberRoot: rootFiber}

	app.dirty = false
	time.Sleep(10 * time.Millisecond)
	app.handleTick()

	if !app.dirty {
		t.Fatal("handleTick should mark app dirty when a timed notification expires")
	}
	if !inst.IsExpired() {
		t.Fatal("notification should be expired after handleTick")
	}
	if inst.IsVisible() {
		t.Fatal("expired notification should be hidden")
	}
}

func TestApp_HandleTick_ExpiresTimedToast(t *testing.T) {
	app := NewApp()

	inst := toast.NewToastBuilder("Saved").
		Duration(time.Millisecond).
		BuildInstance()

	rootFiber := &rtui.Fiber{
		Tag:      "toast",
		Instance: inst,
	}
	app.root = &mockFiberRootNode{fiberRoot: rootFiber}

	app.dirty = false
	time.Sleep(10 * time.Millisecond)
	app.handleTick()

	if !app.dirty {
		t.Fatal("handleTick should mark app dirty when a timed toast expires")
	}
	if !inst.IsExpired() {
		t.Fatal("toast should be expired after handleTick")
	}
	if inst.IsVisible() {
		t.Fatal("expired toast should be hidden")
	}
}

func TestApp_HandleTick_ShowsDelayedTooltip(t *testing.T) {
	app := NewApp()

	inst := tooltip.NewInstance(rtui.Props{
		"text":  "hint",
		"delay": time.Millisecond,
	})
	inst.HandleAction(action.NewAction(action.ActionMouseEnter))

	rootFiber := &rtui.Fiber{
		Tag:      "tooltip",
		Instance: inst,
	}
	app.root = &mockFiberRootNode{fiberRoot: rootFiber}

	app.dirty = false
	time.Sleep(10 * time.Millisecond)
	app.handleTick()

	if !app.dirty {
		t.Fatal("handleTick should mark app dirty when a delayed tooltip becomes visible")
	}
	if children := inst.RuntimeChildren(); len(children) != 1 {
		t.Fatalf("runtime children after tooltip delay = %d, want 1", len(children))
	}
	if inst.WantsTick() {
		t.Fatal("visible tooltip should stop requesting ticks after delay completes")
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

func TestApp_ProcessMsg_SelectArrowNavigationAndEnterCommit(t *testing.T) {
	app := NewApp()
	app.Resize(80, 24)
	selectcomp.Install(app)
	app.actionRouter.SetMiddleware(action.NewMiddlewareChain())

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.VStack(
			rtui.NewElement("box").SetProps(rtui.Props{
				"portalRootId": rtui.DefaultOverlayPortalRootID,
				"position":     "absolute",
				"left":         0,
				"top":          0,
				"width":        1,
				"height":       1,
			}),
			selectcomp.NewBuilder().
				SetID("country-select").
				OverlayPopup(true).
				Options([]selectcomp.Option{
					{Value: "us", Label: "United States"},
					{Value: "cn", Label: "China"},
					{Value: "jp", Label: "Japan"},
				}).
				Selected(0).
				Build(),
		)
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	app.dirty = true
	app.render()
	app.dirty = true
	app.render()

	var selectInst *selectcomp.Instance
	rtui.WalkFiberDepthFirst(app.getFiberRoot(), func(fiber *rtui.Fiber) bool {
		if fiber == nil || fiber.Instance == nil {
			return true
		}
		if inst, ok := fiber.Instance.(*selectcomp.Instance); ok && inst.HasFocus() {
			selectInst = inst
			return false
		}
		return true
	})
	if selectInst == nil {
		t.Fatal("expected focused select instance")
	}

	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyEnter, runtimemsg.Modifiers{}))
	if open, _ := selectInst.GetProp("open"); open != true {
		t.Fatalf("select open after enter = %#v, want true", open)
	}
	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyDown, runtimemsg.Modifiers{}))
	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyEnter, runtimemsg.Modifiers{}))

	if selectInst.SelectedIndex() != 1 {
		t.Fatalf("SelectedIndex after down+enter = %d, want 1", selectInst.SelectedIndex())
	}
	if open, _ := selectInst.GetProp("open"); open != false {
		t.Fatalf("select open after commit = %#v, want false", open)
	}
}

func TestApp_ProcessMsg_SelectMouseOpenThenKeyboardCommit(t *testing.T) {
	app := NewApp()
	app.Resize(80, 24)
	selectcomp.Install(app)
	app.actionRouter.SetMiddleware(action.NewMiddlewareChain())

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.VStack(
			rtui.NewElement("text").SetProps(rtui.Props{"content": "prefix"}),
			rtui.NewElement("box").SetProps(rtui.Props{
				"portalRootId": rtui.DefaultOverlayPortalRootID,
				"position":     "absolute",
				"left":         0,
				"top":          0,
				"width":        1,
				"height":       1,
			}),
			selectcomp.NewBuilder().
				SetID("country-select").
				OverlayPopup(true).
				Options([]selectcomp.Option{
					{Value: "us", Label: "United States"},
					{Value: "cn", Label: "China"},
					{Value: "jp", Label: "Japan"},
				}).
				Selected(0).
				Build(),
		)
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	app.dirty = true
	app.render()
	app.dirty = true
	app.render()

	var selectInst *selectcomp.Instance
	rtui.WalkFiberDepthFirst(app.getFiberRoot(), func(fiber *rtui.Fiber) bool {
		if fiber == nil || fiber.Instance == nil {
			return true
		}
		if inst, ok := fiber.Instance.(*selectcomp.Instance); ok {
			selectInst = inst
			return false
		}
		return true
	})
	if selectInst == nil {
		t.Fatal("expected select instance")
	}

	detail := selectTriggerHitDetail(t, app, 1, 0)
	mouseMsg := runtimemsg.NewMouseMsg(detail.ScreenX, detail.ScreenY, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	mouseMsg.TargetFiber = detail.Entry.TargetFiber
	mouseMsg.LocalX = detail.LocalX
	mouseMsg.LocalY = detail.LocalY
	app.processMsg(mouseMsg)
	mouseRelease := runtimemsg.NewMouseMsg(detail.ScreenX, detail.ScreenY, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)
	mouseRelease.TargetFiber = detail.Entry.TargetFiber
	mouseRelease.LocalX = detail.LocalX
	mouseRelease.LocalY = detail.LocalY
	app.processMsg(mouseRelease)

	if !selectInst.HasFocus() {
		t.Fatal("select should receive focus after trigger click")
	}
	if open, _ := selectInst.GetProp("open"); open != true {
		t.Fatalf("select open after trigger click = %#v, want true", open)
	}

	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyDown, runtimemsg.Modifiers{}))
	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyEnter, runtimemsg.Modifiers{}))

	if selectInst.SelectedIndex() != 1 {
		t.Fatalf("SelectedIndex after mouse-open down+enter = %d, want 1", selectInst.SelectedIndex())
	}
}

func TestApp_SelectPopupTakesFocusAfterOpenRender(t *testing.T) {
	app := NewApp()
	app.Resize(80, 24)
	selectcomp.Install(app)
	app.actionRouter.SetMiddleware(action.NewMiddlewareChain())

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.VStack(
			rtui.NewElement("box").SetProps(rtui.Props{
				"portalRootId": rtui.DefaultOverlayPortalRootID,
				"position":     "absolute",
				"left":         0,
				"top":          0,
				"width":        1,
				"height":       1,
			}),
			selectcomp.NewBuilder().
				SetID("country-select").
				OverlayPopup(true).
				Options([]selectcomp.Option{
					{Value: "us", Label: "United States"},
					{Value: "cn", Label: "China"},
					{Value: "jp", Label: "Japan"},
				}).
				Selected(0).
				Build(),
		)
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	app.dirty = true
	app.render()
	app.dirty = true
	app.render()

	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyEnter, runtimemsg.Modifiers{}))
	app.dirty = true
	app.render()

	focused := app.GetFocusManager().GetCurrent()
	if focused == nil {
		t.Fatal("expected focused fiber")
	}
	if focused.Tag != "select-popup" {
		t.Fatalf("focused fiber tag after popup open = %q, want %q", focused.Tag, "select-popup")
	}

	var selectInst *selectcomp.Instance
	rtui.WalkFiberDepthFirst(app.getFiberRoot(), func(fiber *rtui.Fiber) bool {
		if fiber == nil || fiber.Instance == nil {
			return true
		}
		if inst, ok := fiber.Instance.(*selectcomp.Instance); ok {
			selectInst = inst
			return false
		}
		return true
	})
	if selectInst == nil {
		t.Fatal("expected select instance")
	}
	if open, _ := selectInst.GetProp("open"); open != true {
		t.Fatalf("select open after popup takes focus = %#v, want true", open)
	}
}

func TestApp_ProcessMsg_SelectPopupMouseClickCommitsOption(t *testing.T) {
	app := NewApp()
	app.Resize(80, 24)
	selectcomp.Install(app)
	app.actionRouter.SetMiddleware(action.NewMiddlewareChain())

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.VStack(
			rtui.NewElement("box").SetProps(rtui.Props{
				"portalRootId": rtui.DefaultOverlayPortalRootID,
				"position":     "absolute",
				"left":         0,
				"top":          0,
				"width":        1,
				"height":       1,
			}),
			selectcomp.NewBuilder().
				SetID("country-select").
				OverlayPopup(true).
				Options([]selectcomp.Option{
					{Value: "us", Label: "United States"},
					{Value: "cn", Label: "China"},
					{Value: "jp", Label: "Japan"},
				}).
				Selected(0).
				Build(),
		)
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	app.dirty = true
	app.render()
	app.dirty = true
	app.render()

	var selectInst *selectcomp.Instance
	rtui.WalkFiberDepthFirst(app.getFiberRoot(), func(fiber *rtui.Fiber) bool {
		if fiber == nil || fiber.Instance == nil {
			return true
		}
		if inst, ok := fiber.Instance.(*selectcomp.Instance); ok {
			selectInst = inst
			return false
		}
		return true
	})
	if selectInst == nil {
		t.Fatal("expected select instance")
	}

	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyEnter, runtimemsg.Modifiers{}))
	app.dirty = true
	app.render()

	detail := popupHitDetail(t, app, 1, 2)

	mouseMsg := runtimemsg.NewMouseMsg(detail.ScreenX, detail.ScreenY, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	mouseMsg.TargetFiber = detail.Entry.TargetFiber
	mouseMsg.LocalX = detail.LocalX
	mouseMsg.LocalY = detail.LocalY
	app.processMsg(mouseMsg)
	mouseRelease := runtimemsg.NewMouseMsg(detail.ScreenX, detail.ScreenY, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)
	mouseRelease.TargetFiber = detail.Entry.TargetFiber
	mouseRelease.LocalX = detail.LocalX
	mouseRelease.LocalY = detail.LocalY
	app.processMsg(mouseRelease)
	app.dirty = true
	app.render()

	if selectInst.SelectedIndex() != 1 {
		t.Fatalf("SelectedIndex after popup click = %d, want 1", selectInst.SelectedIndex())
	}
	if open, _ := selectInst.GetProp("open"); open != false {
		t.Fatalf("select open after popup click = %#v, want false", open)
	}
}

func TestApp_SelectPopupHitMapLocalCoordinates(t *testing.T) {
	app := NewApp()
	app.Resize(80, 24)
	selectcomp.Install(app)
	app.actionRouter.SetMiddleware(action.NewMiddlewareChain())

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.VStack(
			rtui.NewElement("box").SetProps(rtui.Props{
				"portalRootId": rtui.DefaultOverlayPortalRootID,
				"position":     "absolute",
				"left":         0,
				"top":          0,
				"width":        1,
				"height":       1,
			}),
			selectcomp.NewBuilder().
				SetID("country-select").
				OverlayPopup(true).
				Options([]selectcomp.Option{
					{Value: "us", Label: "United States"},
					{Value: "cn", Label: "China"},
					{Value: "jp", Label: "Japan"},
				}).
				Selected(0).
				Build(),
		)
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	app.dirty = true
	app.render()
	app.dirty = true
	app.render()

	var selectInst *selectcomp.Instance
	rtui.WalkFiberDepthFirst(app.getFiberRoot(), func(fiber *rtui.Fiber) bool {
		if fiber == nil || fiber.Instance == nil {
			return true
		}
		if inst, ok := fiber.Instance.(*selectcomp.Instance); ok {
			selectInst = inst
			return false
		}
		return true
	})
	if selectInst == nil {
		t.Fatal("expected select instance")
	}

	app.processMsg(runtimemsg.NewKeyMsg(0, runtimeplatform.KeyEnter, runtimemsg.Modifiers{}))
	app.dirty = true
	app.render()

	detail := popupHitDetail(t, app, 1, 2)
	if detail.LocalX != 1 || detail.LocalY != 2 {
		t.Fatalf("popup local coords = (%d,%d), want (1,2)", detail.LocalX, detail.LocalY)
	}
	fiber, ok := detail.Entry.TargetFiber.(*rtui.Fiber)
	if !ok || fiber == nil {
		t.Fatal("expected target fiber from popup hit")
	}
	if fiber.Tag != "select-popup" {
		t.Fatalf("popup target fiber tag = %q, want %q", fiber.Tag, "select-popup")
	}
}

func popupHitDetail(t *testing.T, app *App, localX, localY int) *runtimeevent.DetailedHitTestResult {
	t.Helper()
	hitMap := app.GetHitMap()
	if hitMap == nil {
		t.Fatal("expected hitmap after render")
	}

	var popupEntry *runtimeevent.HitMapEntry
	for _, entry := range hitMap.AllEntries() {
		fiber, ok := entry.TargetFiber.(*rtui.Fiber)
		if !ok || fiber == nil {
			continue
		}
		if fiber.Tag == "select-popup" {
			entryCopy := entry
			popupEntry = &entryCopy
			break
		}
	}
	if popupEntry == nil {
		t.Fatal("expected select-popup entry in hitmap")
	}

	screenX := popupEntry.Bounds.X + localX
	screenY := popupEntry.Bounds.Y + localY
	detail := hitMap.HitTestDetailed(screenX, screenY)
	if detail == nil || !detail.Found || detail.Entry == nil {
		t.Fatalf("expected popup hit at (%d,%d)", screenX, screenY)
	}
	return detail
}

func hitDetailForFiberID(t *testing.T, app *App, fiberID string, localX, localY int) *runtimeevent.DetailedHitTestResult {
	t.Helper()
	hitMap := app.GetHitMap()
	if hitMap == nil {
		t.Fatal("expected hitmap after render")
	}

	var targetEntry *runtimeevent.HitMapEntry
	for _, entry := range hitMap.AllEntries() {
		fiber, ok := entry.TargetFiber.(*rtui.Fiber)
		if !ok || fiber == nil {
			continue
		}
		if fiber.ID == fiberID {
			entryCopy := entry
			targetEntry = &entryCopy
			break
		}
	}
	if targetEntry == nil {
		t.Fatalf("expected hitmap entry for fiber id %q", fiberID)
	}

	screenX := targetEntry.Bounds.X + localX
	screenY := targetEntry.Bounds.Y + localY
	detail := hitMap.HitTestDetailed(screenX, screenY)
	if detail == nil || !detail.Found || detail.Entry == nil {
		t.Fatalf("expected hit at (%d,%d) for fiber id %q", screenX, screenY, fiberID)
	}
	return detail
}

func staticTerminalFrameSceneVNode() rtui.VNode {
	return newSceneImageTestVNode("linechart-image-prototype-image", "Image Plot Backend", 31, 6, 248, 72)
}

func describeActionPath(app *App, target any) string {
	fiber, ok := target.(*rtui.Fiber)
	if !ok || fiber == nil {
		return "<nil>"
	}

	parts := make([]string, 0, 8)
	for node := fiber; node != nil; node = node.Return {
		instanceType := "<nil>"
		if node.Instance != nil {
			instanceType = fmt.Sprintf("%T", node.Instance)
		}
		_, hasInstanceHandler := node.Instance.(rtui.ActionHandlerInstance)
		hasScopeHandler := false
		if app != nil && app.scopeDispatcher != nil && node.ActionTargetID != "" {
			hasScopeHandler = app.scopeDispatcher.HasHandler(node.ActionTargetID)
		}
		hasRouterHandler := false
		if app != nil && app.actionRouter != nil && node.ActionTargetID != "" {
			_, hasRouterHandler = app.actionRouter.TargetHandlers[node.ActionTargetID]
		}
		parts = append(parts, fmt.Sprintf("%s[id=%s target=%s inst=%s instHandler=%v scope=%v router=%v]",
			node.Tag, node.ID, node.ActionTargetID, instanceType, hasInstanceHandler, hasScopeHandler, hasRouterHandler))
	}
	return strings.Join(parts, " -> ")
}

func describeLayoutActionPath(app *App, target any) string {
	fiber, ok := target.(*rtui.Fiber)
	if !ok || fiber == nil || app == nil || app.actionRouter == nil || app.actionRouter.Root == nil {
		return "<nil>"
	}

	targetID := fiber.GetActionTargetID()
	if targetID == "" {
		return "<no-target-id>"
	}

	node := findLayoutNodeByID(app.actionRouter.Root, targetID)
	if node == nil {
		return "<layout-miss>"
	}

	parts := make([]string, 0, 8)
	for current := node; current != nil; current = current.Parent {
		instanceType := "<nil>"
		hasTarget := false
		if current.Component != nil && current.Component.Instance != nil {
			instanceType = fmt.Sprintf("%T", current.Component.Instance)
			_, hasTarget = current.Component.Instance.(action.Target)
		}
		parts = append(parts, fmt.Sprintf("%v[id=%s inst=%s target=%v]", current.Type, current.ID, instanceType, hasTarget))
	}
	return strings.Join(parts, " -> ")
}

func findLayoutNodeByID(root *runtime.LayoutNode, id string) *runtime.LayoutNode {
	if root == nil {
		return nil
	}
	if root.ID == id {
		return root
	}
	for _, child := range root.Children {
		if found := findLayoutNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

func selectTriggerHitDetail(t *testing.T, app *App, localX, localY int) *runtimeevent.DetailedHitTestResult {
	t.Helper()
	hitMap := app.GetHitMap()
	if hitMap == nil {
		t.Fatal("expected hitmap after render")
	}

	var triggerEntry *runtimeevent.HitMapEntry
	for _, entry := range hitMap.AllEntries() {
		fiber, ok := entry.TargetFiber.(*rtui.Fiber)
		if !ok || fiber == nil {
			continue
		}
		if nearestFocusableFiber(fiber) != nil {
			if focusable, ok := nearestFocusableFiber(fiber).Instance.(*selectcomp.Instance); ok && focusable != nil {
				entryCopy := entry
				triggerEntry = &entryCopy
				break
			}
		}
	}
	if triggerEntry == nil {
		t.Fatal("expected select trigger entry in hitmap")
	}

	screenX := triggerEntry.Bounds.X + localX
	screenY := triggerEntry.Bounds.Y + localY
	detail := hitMap.HitTestDetailed(screenX, screenY)
	if detail == nil || !detail.Found || detail.Entry == nil {
		t.Fatalf("expected trigger hit at (%d,%d)", screenX, screenY)
	}
	return detail
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

func TestApp_ProcessMsg_TargetedClickRunsMiddlewareBeforeFiberDispatch(t *testing.T) {
	app := NewApp()
	middleware := &recordingMiddleware{}
	app.AddMiddleware(middleware)

	recorder := &clickActionRecorder{}
	fiber := &rtui.Fiber{
		Tag:      "button",
		NodeID:   99,
		Instance: recorder,
	}

	mouseMsg := runtimemsg.NewMouseMsg(3, 4, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	mouseMsg.TargetFiber = fiber

	app.processMsg(mouseMsg)

	if len(middleware.beforeCalls) != 1 || middleware.beforeCalls[0] != action.ActionClick {
		t.Fatalf("middleware before calls = %#v, want [ActionClick]", middleware.beforeCalls)
	}
	if len(recorder.handled) != 1 || recorder.handled[0] != action.ActionClick {
		t.Fatalf("handled actions = %#v, want [ActionClick]", recorder.handled)
	}
}

func TestApp_ProcessMsg_StaticTerminalFrameSceneMouseMoveDoesNotMarkDirty(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	app := NewApp()
	app.Resize(104, 24)
	var handled bool
	var stage string
	var actionType action.ActionType
	app.SetTestActionProbe(func(act *action.Action, wasHandled bool, handledStage string) {
		actionType = act.Type
		handled = wasHandled
		stage = handledStage
	})
	presenter := &recordingGraphicsPresenter{
		caps: runtimeplatform.GraphicsCapabilities{
			Mode:              runtimeplatform.GraphicsModeSixel,
			PresentationModel: runtimeplatform.GraphicsPresentationModelTerminalFrame,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			SupportsDelete:    false,
		},
	}
	app.SetGraphicsPresenter(presenter)

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return staticTerminalFrameSceneVNode()
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	captureStdout(t, func() {
		app.render()
	})
	if presenter.presentCalls != 1 {
		t.Fatalf("initial present calls = %d, want 1", presenter.presentCalls)
	}

	detail := hitDetailForFiberID(t, app, "linechart-image-prototype-image", 2, 2)
	mouseMsg := runtimemsg.NewMouseMsg(detail.ScreenX, detail.ScreenY, runtimemsg.MouseButtonUnknown, runtimemsg.MouseActionMove)
	mouseMsg.TargetID = detail.Entry.NodeID
	mouseMsg.TargetFiber = detail.Entry.TargetFiber
	mouseMsg.LocalX = detail.LocalX
	mouseMsg.LocalY = detail.LocalY

	app.dirty = false
	app.processMsg(mouseMsg)

	if app.dirty {
		t.Fatalf("expected static scene mouse move to leave app clean, action=%q handled=%v stage=%q fiberPath=%s layoutPath=%s", actionType, handled, stage, describeActionPath(app, detail.Entry.TargetFiber), describeLayoutActionPath(app, detail.Entry.TargetFiber))
	}
	if presenter.presentCalls != 1 {
		t.Fatalf("present calls after mouse move = %d, want 1", presenter.presentCalls)
	}
}

func TestApp_ProcessMsg_StaticTerminalFrameSceneMouseClickDoesNotMarkDirty(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	app := NewApp()
	app.Resize(104, 24)
	var handled bool
	var stage string
	var actionType action.ActionType
	app.SetTestActionProbe(func(act *action.Action, wasHandled bool, handledStage string) {
		actionType = act.Type
		handled = wasHandled
		stage = handledStage
	})
	presenter := &recordingGraphicsPresenter{
		caps: runtimeplatform.GraphicsCapabilities{
			Mode:              runtimeplatform.GraphicsModeSixel,
			PresentationModel: runtimeplatform.GraphicsPresentationModelTerminalFrame,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			SupportsDelete:    false,
		},
	}
	app.SetGraphicsPresenter(presenter)

	decl := irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return staticTerminalFrameSceneVNode()
	})
	decl.SetApp(app)
	if fm := decl.GetFocusManager(); fm != nil {
		app.SetFocusManagerFromDeclarativeNode(fm)
	}
	app.SetRoot(decl)

	captureStdout(t, func() {
		app.render()
	})
	if presenter.presentCalls != 1 {
		t.Fatalf("initial present calls = %d, want 1", presenter.presentCalls)
	}

	detail := hitDetailForFiberID(t, app, "linechart-image-prototype-image", 2, 2)
	mousePress := runtimemsg.NewMouseMsg(detail.ScreenX, detail.ScreenY, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	mousePress.TargetID = detail.Entry.NodeID
	mousePress.TargetFiber = detail.Entry.TargetFiber
	mousePress.LocalX = detail.LocalX
	mousePress.LocalY = detail.LocalY

	mouseRelease := runtimemsg.NewMouseMsg(detail.ScreenX, detail.ScreenY, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)
	mouseRelease.TargetID = detail.Entry.NodeID
	mouseRelease.TargetFiber = detail.Entry.TargetFiber
	mouseRelease.LocalX = detail.LocalX
	mouseRelease.LocalY = detail.LocalY

	app.dirty = false
	app.processMsg(mousePress)
	if app.dirty {
		t.Fatalf("expected static scene mouse press to leave app clean, action=%q handled=%v stage=%q fiberPath=%s layoutPath=%s", actionType, handled, stage, describeActionPath(app, detail.Entry.TargetFiber), describeLayoutActionPath(app, detail.Entry.TargetFiber))
	}

	app.processMsg(mouseRelease)
	if app.dirty {
		t.Fatalf("expected static scene mouse release to leave app clean, action=%q handled=%v stage=%q fiberPath=%s layoutPath=%s", actionType, handled, stage, describeActionPath(app, detail.Entry.TargetFiber), describeLayoutActionPath(app, detail.Entry.TargetFiber))
	}
	if presenter.presentCalls != 1 {
		t.Fatalf("present calls after mouse click = %d, want 1", presenter.presentCalls)
	}
}

func TestApp_Render_TextOnlyPathStillOutputsText(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	app := NewApp()
	app.Resize(20, 4)
	app.root = &renderTextNode{text: "hello scene"}

	output := captureStdout(t, func() {
		app.render()
	})

	if !strings.Contains(app.GetRenderer().GetRenderSnapshot(), "hello scene") {
		t.Fatalf("render snapshot = %q, want text content", app.GetRenderer().GetRenderSnapshot())
	}
	if !strings.Contains(output, "hello scene") {
		t.Fatalf("stdout output = %q, want text content", output)
	}
}

func TestApp_Render_SceneWithImagesBypassesAsyncRenderer(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")
	app := NewApp()
	app.Resize(20, 4)
	app.asyncRenderer = paint.NewAsyncRenderer(20, 4, paint.AsyncRendererOptions{
		Output: func(string) {},
	})
	presenter := &recordingGraphicsPresenter{}
	app.SetGraphicsPresenter(presenter)
	app.root = &renderSceneNode{
		text: "image frame",
		scene: &paint.SceneFrame{
			ImageLayers: []paint.ImageLayer{{
				ID:          "plot",
				Bounds:      paint.Rect{X: 1, Y: 1, Width: 6, Height: 2},
				PixelWidth:  12,
				PixelHeight: 4,
				RGBA:        []byte{255, 0, 0, 255},
			}},
		},
	}

	output := captureStdout(t, func() {
		app.render()
	})

	if got := app.asyncRenderer.Stats().SubmittedFrames; got != 0 {
		t.Fatalf("async submitted frames = %d, want 0 for image bypass", got)
	}
	if presenter.presentCalls != 1 {
		t.Fatalf("present calls = %d, want 1", presenter.presentCalls)
	}
	if !app.graphicsImagesOn {
		t.Fatal("expected graphicsImagesOn to be true after successful image present")
	}
	if !strings.Contains(output, "image frame") {
		t.Fatalf("stdout output = %q, want image frame text", output)
	}
}

func TestApp_Render_ScenePresenterFailureFallsBackToText(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")
	app := NewApp()
	app.Resize(20, 4)
	app.asyncRenderer = paint.NewAsyncRenderer(20, 4, paint.AsyncRendererOptions{
		Output: func(string) {},
	})
	presenter := &recordingGraphicsPresenter{failPresent: true}
	app.SetGraphicsPresenter(presenter)
	app.root = &renderSceneNode{
		text: "fallback text",
		scene: &paint.SceneFrame{
			ImageLayers: []paint.ImageLayer{{
				ID:          "plot",
				Bounds:      paint.Rect{X: 1, Y: 1, Width: 6, Height: 2},
				PixelWidth:  12,
				PixelHeight: 4,
				RGBA:        []byte{255, 0, 0, 255},
			}},
		},
	}

	output := captureStdout(t, func() {
		app.render()
	})

	if got := app.asyncRenderer.Stats().SubmittedFrames; got != 0 {
		t.Fatalf("async submitted frames = %d, want 0 for failed image bypass", got)
	}
	if presenter.presentCalls != 1 {
		t.Fatalf("present calls = %d, want 1", presenter.presentCalls)
	}
	if app.graphicsImagesOn {
		t.Fatal("expected graphicsImagesOn to remain false after presenter failure")
	}
	if !strings.Contains(output, "fallback text") {
		t.Fatalf("stdout output = %q, want fallback text", output)
	}
}

func TestApp_Render_DeclarativeSceneLayersBypassAsyncRenderer(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	app := NewApp()
	app.Resize(24, 8)
	app.asyncRenderer = paint.NewAsyncRenderer(24, 8, paint.AsyncRendererOptions{
		Output: func(string) {},
	})

	presenter := &recordingGraphicsPresenter{}
	app.SetGraphicsPresenter(presenter)
	app.root = irender.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return newSceneImageTestVNode("trend-image-scene", "Trend", 5, 4, 40, 48)
	})

	output := captureStdout(t, func() {
		app.render()
	})

	if got := app.asyncRenderer.Stats().SubmittedFrames; got != 0 {
		t.Fatalf("async submitted frames = %d, want 0 for declarative scene bypass", got)
	}
	if presenter.presentCalls != 1 {
		t.Fatalf("present calls = %d, want 1", presenter.presentCalls)
	}
	if len(presenter.requests) != 1 {
		t.Fatalf("recorded requests = %d, want 1", len(presenter.requests))
	}

	req := presenter.requests[0]
	if req.CellX != 0 || req.CellY != 1 {
		t.Fatalf("request cell origin = (%d,%d), want (0,1)", req.CellX, req.CellY)
	}
	if req.CellWidth != 5 || req.CellHeight != 4 {
		t.Fatalf("request cell size = %dx%d, want 5x4", req.CellWidth, req.CellHeight)
	}
	if req.PixelWidth != 40 || req.PixelHeight != 48 {
		t.Fatalf("request pixel size = %dx%d, want 40x48", req.PixelWidth, req.PixelHeight)
	}
	if len(req.RGBA) != req.PixelWidth*req.PixelHeight*4 {
		t.Fatalf("len(req.RGBA) = %d, want %d", len(req.RGBA), req.PixelWidth*req.PixelHeight*4)
	}
	if !app.graphicsImagesOn {
		t.Fatal("expected graphicsImagesOn to be true after declarative scene present")
	}
	if !strings.Contains(output, "Trend") {
		t.Fatalf("stdout output = %q, want title text", output)
	}
}

func TestApp_Render_NonDeletableGraphicsClearBeforeTextRepaint(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	app := NewApp()
	app.Resize(20, 4)
	app.asyncRenderer = paint.NewAsyncRenderer(20, 4, paint.AsyncRendererOptions{
		Output: func(string) {},
	})
	presenter := &recordingGraphicsPresenter{
		caps: runtimeplatform.GraphicsCapabilities{
			Mode:              runtimeplatform.GraphicsModeSixel,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			SupportsDelete:    false,
		},
	}
	app.SetGraphicsPresenter(presenter)
	app.root = &renderSceneNode{
		text: "image frame",
		scene: &paint.SceneFrame{
			ImageLayers: []paint.ImageLayer{{
				ID:          "plot",
				Bounds:      paint.Rect{X: 1, Y: 1, Width: 6, Height: 2},
				PixelWidth:  12,
				PixelHeight: 4,
				RGBA:        []byte{255, 0, 0, 255},
			}},
		},
	}

	captureStdout(t, func() {
		app.render()
	})
	if !app.graphicsImagesOn {
		t.Fatal("expected graphicsImagesOn to be true after image frame")
	}

	app.root = &renderSceneNode{text: "text only"}
	output := captureStdout(t, func() {
		app.render()
	})

	if presenter.clearCalls != 1 {
		t.Fatalf("clear calls = %d, want 1", presenter.clearCalls)
	}
	if got := app.asyncRenderer.Stats().SubmittedFrames; got != 0 {
		t.Fatalf("async submitted frames = %d, want 0 after non-deletable graphics clear", got)
	}
	if app.graphicsImagesOn {
		t.Fatal("expected graphicsImagesOn to be false after text repaint")
	}
	if !strings.Contains(output, "text only") {
		t.Fatalf("stdout output = %q, want text only content", output)
	}
}

func TestApp_Render_NonDeletableGraphicsSameLayoutDoesNotClearBetweenFrames(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	app := NewApp()
	app.Resize(20, 4)
	app.asyncRenderer = paint.NewAsyncRenderer(20, 4, paint.AsyncRendererOptions{
		Output: func(string) {},
	})
	presenter := &recordingGraphicsPresenter{
		caps: runtimeplatform.GraphicsCapabilities{
			Mode:              runtimeplatform.GraphicsModeSixel,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			SupportsDelete:    false,
		},
	}
	app.SetGraphicsPresenter(presenter)

	scene := &paint.SceneFrame{
		ImageLayers: []paint.ImageLayer{{
			ID:          "plot",
			Bounds:      paint.Rect{X: 1, Y: 1, Width: 6, Height: 2},
			PixelWidth:  12,
			PixelHeight: 4,
			RGBA:        []byte{255, 0, 0, 255},
		}},
	}

	app.root = &renderSceneNode{text: "frame one", scene: scene}
	captureStdout(t, func() {
		app.render()
	})
	if !app.graphicsImagesOn {
		t.Fatal("expected graphicsImagesOn to be true after first image frame")
	}

	app.root = &renderSceneNode{text: "frame two", scene: scene}
	output := captureStdout(t, func() {
		app.render()
	})

	if presenter.clearCalls != 0 {
		t.Fatalf("clear calls = %d, want 0 for same-layout rerender", presenter.clearCalls)
	}
	if presenter.presentCalls != 2 {
		t.Fatalf("present calls = %d, want 2 for terminal-frame same-layout rerender", presenter.presentCalls)
	}
	if !app.graphicsImagesOn {
		t.Fatal("expected graphicsImagesOn to remain true after same-layout rerender")
	}
	if !strings.Contains(output, "frame two") {
		t.Fatalf("stdout output = %q, want updated text content", output)
	}
}

func TestApp_Render_SameGeometryChangedImageContentPresentsAgain(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	app := NewApp()
	app.Resize(20, 4)
	presenter := &recordingGraphicsPresenter{
		caps: runtimeplatform.GraphicsCapabilities{
			Mode:              runtimeplatform.GraphicsModeSixel,
			PresentationModel: runtimeplatform.GraphicsPresentationModelTerminalFrame,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			SupportsDelete:    false,
		},
	}
	app.SetGraphicsPresenter(presenter)

	firstScene := &paint.SceneFrame{
		ImageLayers: []paint.ImageLayer{{
			ID:          "plot",
			Bounds:      paint.Rect{X: 1, Y: 1, Width: 6, Height: 2},
			PixelWidth:  1,
			PixelHeight: 1,
			RGBA:        []byte{255, 0, 0, 255},
		}},
	}
	secondScene := &paint.SceneFrame{
		ImageLayers: []paint.ImageLayer{{
			ID:          "plot",
			Bounds:      paint.Rect{X: 1, Y: 1, Width: 6, Height: 2},
			PixelWidth:  1,
			PixelHeight: 1,
			RGBA:        []byte{0, 255, 0, 255},
		}},
	}

	app.root = &renderSceneNode{text: "frame", scene: firstScene}
	captureStdout(t, func() {
		app.render()
	})
	app.root = &renderSceneNode{text: "frame", scene: secondScene}
	captureStdout(t, func() {
		app.render()
	})

	if presenter.clearCalls != 0 {
		t.Fatalf("clear calls = %d, want 0 for same-geometry content update", presenter.clearCalls)
	}
	if presenter.presentCalls != 2 {
		t.Fatalf("present calls = %d, want 2 when image content changes", presenter.presentCalls)
	}
}

func TestApp_Render_TerminalFrameGraphicsStableSceneRepresentsWithoutFullTextRepaint(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	app := NewApp()
	app.Resize(20, 4)
	presenter := &recordingGraphicsPresenter{
		caps: runtimeplatform.GraphicsCapabilities{
			Mode:              runtimeplatform.GraphicsModeSixel,
			PresentationModel: runtimeplatform.GraphicsPresentationModelTerminalFrame,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			SupportsDelete:    false,
		},
	}
	app.SetGraphicsPresenter(presenter)

	scene := &paint.SceneFrame{
		ImageLayers: []paint.ImageLayer{{
			ID:          "plot",
			Bounds:      paint.Rect{X: 1, Y: 1, Width: 6, Height: 2},
			PixelWidth:  12,
			PixelHeight: 4,
			RGBA:        []byte{255, 0, 0, 255},
		}},
	}

	app.root = &renderSceneNode{text: "stable frame", scene: scene}
	captureStdout(t, func() {
		app.render()
	})

	app.root = &renderSceneNode{text: "stable frame", scene: scene}
	output := captureStdout(t, func() {
		app.render()
	})

	if presenter.clearCalls != 0 {
		t.Fatalf("clear calls = %d, want 0 for stable terminal-frame rerender", presenter.clearCalls)
	}
	if presenter.presentCalls != 2 {
		t.Fatalf("present calls = %d, want 2 for stable terminal-frame rerender", presenter.presentCalls)
	}
	if strings.Contains(output, "stable frame") {
		t.Fatalf("stdout output = %q, want diff-based text rendering for unchanged terminal-frame graphics", output)
	}
}

func TestApp_Render_OverlayGraphicsStableSceneKeepsDiffBasedTextRendering(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	app := NewApp()
	app.Resize(20, 4)
	presenter := &recordingGraphicsPresenter{
		caps: runtimeplatform.GraphicsCapabilities{
			Mode:              runtimeplatform.GraphicsModeKitty,
			PresentationModel: runtimeplatform.GraphicsPresentationModelOverlay,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			SupportsDelete:    true,
		},
	}
	app.SetGraphicsPresenter(presenter)

	scene := &paint.SceneFrame{
		ImageLayers: []paint.ImageLayer{{
			ID:          "plot",
			Bounds:      paint.Rect{X: 1, Y: 1, Width: 6, Height: 2},
			PixelWidth:  12,
			PixelHeight: 4,
			RGBA:        []byte{255, 0, 0, 255},
		}},
	}

	app.root = &renderSceneNode{text: "stable frame", scene: scene}
	captureStdout(t, func() {
		app.render()
	})

	app.root = &renderSceneNode{text: "stable frame", scene: scene}
	output := captureStdout(t, func() {
		app.render()
	})

	if presenter.presentCalls != 1 {
		t.Fatalf("present calls = %d, want 1 for unchanged overlay rerender", presenter.presentCalls)
	}
	if strings.Contains(output, "stable frame") {
		t.Fatalf("stdout output = %q, want diff-based text rendering without forced full repaint", output)
	}
}

func TestApp_Render_SceneBypassMasksTextUnderImageBounds(t *testing.T) {
	t.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")

	app := NewApp()
	app.Resize(20, 4)
	presenter := &recordingGraphicsPresenter{
		caps: runtimeplatform.GraphicsCapabilities{
			Mode:              runtimeplatform.GraphicsModeSixel,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			SupportsDelete:    false,
		},
	}
	app.SetGraphicsPresenter(presenter)
	app.root = &renderSceneTextRowsNode{
		rows: map[int]string{
			0: "title",
			1: "ABCDE",
			2: "footer",
		},
		scene: &paint.SceneFrame{
			ImageLayers: []paint.ImageLayer{{
				ID:          "plot",
				Bounds:      paint.Rect{X: 0, Y: 1, Width: 5, Height: 1},
				PixelWidth:  10,
				PixelHeight: 2,
				RGBA:        []byte{255, 0, 0, 255},
			}},
		},
	}

	output := captureStdout(t, func() {
		app.render()
	})

	if strings.Contains(output, "ABCDE") {
		t.Fatalf("stdout output = %q, want image-covered text to be masked", output)
	}
	if !strings.Contains(output, "title") || !strings.Contains(output, "footer") {
		t.Fatalf("stdout output = %q, want surrounding text rows to remain", output)
	}
	if presenter.presentCalls != 1 {
		t.Fatalf("present calls = %d, want 1", presenter.presentCalls)
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

type clickActionRecorder struct {
	handled []action.ActionType
}

func (m *clickActionRecorder) Key() string                        { return "" }
func (m *clickActionRecorder) SetKey(key string)                  {}
func (m *clickActionRecorder) Init(props rtui.Props)              {}
func (m *clickActionRecorder) Destroy()                           {}
func (m *clickActionRecorder) OnMount()                           {}
func (m *clickActionRecorder) OnUnmount()                         {}
func (m *clickActionRecorder) SetProps(props rtui.Props) bool     { return false }
func (m *clickActionRecorder) GetProps() rtui.Props               { return nil }
func (m *clickActionRecorder) MarkDirty()                         {}
func (m *clickActionRecorder) IsDirty() bool                      { return false }
func (m *clickActionRecorder) GetContext() *rtui.ComponentContext { return nil }
func (m *clickActionRecorder) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}
	m.handled = append(m.handled, act.Type)
	return act.Type == action.ActionClick
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

func mustPropBool(t *testing.T, inst interface {
	GetProp(string) (interface{}, bool)
}, key string) bool {
	t.Helper()
	value, ok := inst.GetProp(key)
	if !ok {
		t.Fatalf("prop %q missing", key)
	}
	b, ok := value.(bool)
	if !ok {
		t.Fatalf("prop %q type = %T, want bool", key, value)
	}
	return b
}

// testPanicHandler 测试用的 panic 处理器
type testPanicHandler struct{}

func (h *testPanicHandler) HandlePanic(r interface{}, stack []byte) {
	// 测试实现，什么都不做
}

type renderTextNode struct {
	text string
}

func (n *renderTextNode) ID() string {
	return "render-text-node"
}

func (n *renderTextNode) Type() string {
	return "render-text-node"
}

func (n *renderTextNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	buf.SetString(ctx.X, ctx.Y, n.text, style.Style{})
}

type renderSceneNode struct {
	text  string
	scene *paint.SceneFrame
}

func (n *renderSceneNode) ID() string {
	return "render-scene-node"
}

func (n *renderSceneNode) Type() string {
	return "render-scene-node"
}

func (n *renderSceneNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	buf.SetString(ctx.X, ctx.Y, n.text, style.Style{})
}

func (n *renderSceneNode) PaintScene(ctx component.PaintContext, buf *paint.Buffer) *paint.SceneFrame {
	n.Paint(ctx, buf)
	if n.scene == nil {
		return nil
	}

	sceneCopy := *n.scene
	sceneCopy.Buffer = buf
	sceneCopy.ImageLayers = paint.CloneImageLayers(n.scene.ImageLayers)
	sceneCopy.Diagnostics.Notes = append([]string(nil), n.scene.Diagnostics.Notes...)
	return &sceneCopy
}

type renderSceneTextRowsNode struct {
	rows  map[int]string
	scene *paint.SceneFrame
}

func (n *renderSceneTextRowsNode) ID() string {
	return "render-scene-text-rows-node"
}

func (n *renderSceneTextRowsNode) Type() string {
	return "render-scene-text-rows-node"
}

func (n *renderSceneTextRowsNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	for row, text := range n.rows {
		buf.SetString(ctx.X, ctx.Y+row, text, style.Style{})
	}
}

func (n *renderSceneTextRowsNode) PaintScene(ctx component.PaintContext, buf *paint.Buffer) *paint.SceneFrame {
	n.Paint(ctx, buf)
	if n.scene == nil {
		return nil
	}

	sceneCopy := *n.scene
	sceneCopy.Buffer = buf
	sceneCopy.ImageLayers = paint.CloneImageLayers(n.scene.ImageLayers)
	sceneCopy.Diagnostics.Notes = append([]string(nil), n.scene.Diagnostics.Notes...)
	return &sceneCopy
}

type sceneImageTestVNode struct {
	*rtui.ElementVNode
}

const (
	sceneImageTestPropTitle       = "title"
	sceneImageTestPropPlotWidth   = "plotWidth"
	sceneImageTestPropPlotHeight  = "plotHeight"
	sceneImageTestPropPixelWidth  = "pixelWidth"
	sceneImageTestPropPixelHeight = "pixelHeight"
)

func newSceneImageTestVNode(id, title string, plotWidth, plotHeight, pixelWidth, pixelHeight int) *sceneImageTestVNode {
	node := &sceneImageTestVNode{ElementVNode: rtui.NewElement("scene-image-test")}
	node.SetID(id)
	node.SetProps(rtui.Props{
		sceneImageTestPropTitle:       title,
		sceneImageTestPropPlotWidth:   plotWidth,
		sceneImageTestPropPlotHeight:  plotHeight,
		sceneImageTestPropPixelWidth:  pixelWidth,
		sceneImageTestPropPixelHeight: pixelHeight,
	})
	return node
}

func (v *sceneImageTestVNode) CreateInstance() rtui.ComponentInstance {
	return newSceneImageTestInstance(v.ID(), v.Props())
}

type sceneImageTestInstance struct {
	key         string
	id          string
	title       string
	plotWidth   int
	plotHeight  int
	pixelWidth  int
	pixelHeight int
	bounds      [4]int
	dirty       bool
}

var (
	_ rtui.ComponentInstance      = (*sceneImageTestInstance)(nil)
	_ rtui.PaintableInstance      = (*sceneImageTestInstance)(nil)
	_ rtui.ScenePaintableInstance = (*sceneImageTestInstance)(nil)
)

func newSceneImageTestInstance(id string, props rtui.Props) *sceneImageTestInstance {
	inst := &sceneImageTestInstance{id: id, dirty: true}
	inst.Init(props)
	return inst
}

func (inst *sceneImageTestInstance) Key() string                        { return inst.key }
func (inst *sceneImageTestInstance) SetKey(key string)                  { inst.key = key }
func (inst *sceneImageTestInstance) Destroy()                           {}
func (inst *sceneImageTestInstance) OnMount()                           {}
func (inst *sceneImageTestInstance) OnUnmount()                         {}
func (inst *sceneImageTestInstance) MarkDirty()                         { inst.dirty = true }
func (inst *sceneImageTestInstance) IsDirty() bool                      { return inst.dirty }
func (inst *sceneImageTestInstance) GetContext() *rtui.ComponentContext { return nil }

func (inst *sceneImageTestInstance) Init(props rtui.Props) {
	inst.SetProps(props)
}

func (inst *sceneImageTestInstance) SetProps(props rtui.Props) bool {
	oldTitle := inst.title
	oldPlotWidth := inst.plotWidth
	oldPlotHeight := inst.plotHeight
	oldPixelWidth := inst.pixelWidth
	oldPixelHeight := inst.pixelHeight

	if title, ok := props[sceneImageTestPropTitle].(string); ok {
		inst.title = title
	}
	if plotWidth, ok := props[sceneImageTestPropPlotWidth].(int); ok {
		inst.plotWidth = plotWidth
	}
	if plotHeight, ok := props[sceneImageTestPropPlotHeight].(int); ok {
		inst.plotHeight = plotHeight
	}
	if pixelWidth, ok := props[sceneImageTestPropPixelWidth].(int); ok {
		inst.pixelWidth = pixelWidth
	}
	if pixelHeight, ok := props[sceneImageTestPropPixelHeight].(int); ok {
		inst.pixelHeight = pixelHeight
	}

	changed := oldTitle != inst.title ||
		oldPlotWidth != inst.plotWidth ||
		oldPlotHeight != inst.plotHeight ||
		oldPixelWidth != inst.pixelWidth ||
		oldPixelHeight != inst.pixelHeight
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *sceneImageTestInstance) GetProps() rtui.Props {
	return rtui.Props{
		sceneImageTestPropTitle:       inst.title,
		sceneImageTestPropPlotWidth:   inst.plotWidth,
		sceneImageTestPropPlotHeight:  inst.plotHeight,
		sceneImageTestPropPixelWidth:  inst.pixelWidth,
		sceneImageTestPropPixelHeight: inst.pixelHeight,
	}
}

func (inst *sceneImageTestInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *sceneImageTestInstance) Measure(constraints layout.Constraints) layout.Size {
	width := inst.plotWidth
	if width < len(inst.title) {
		width = len(inst.title)
	}
	height := inst.plotHeight + 1
	return layout.Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(height),
	}
}

func (inst *sceneImageTestInstance) Paint(x, y int) []paint.DrawCmd {
	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  inst.title,
		Style: style.Style{},
	}}
}

func (inst *sceneImageTestInstance) SceneLayers() []paint.ImageLayer {
	if inst.plotWidth <= 0 || inst.plotHeight <= 0 || inst.pixelWidth <= 0 || inst.pixelHeight <= 0 {
		return nil
	}
	return []paint.ImageLayer{{
		ID:          inst.id + ":plot-image",
		Bounds:      paint.Rect{X: inst.bounds[0], Y: inst.bounds[1] + 1, Width: inst.plotWidth, Height: inst.plotHeight},
		PixelWidth:  inst.pixelWidth,
		PixelHeight: inst.pixelHeight,
		RGBA:        solidSceneImageRGBA(inst.pixelWidth, inst.pixelHeight, 255, 0, 0, 255),
		AltText:     inst.title + " plot image",
	}}
}

func solidSceneImageRGBA(width, height int, r, g, b, a byte) []byte {
	if width <= 0 || height <= 0 {
		return nil
	}
	rgba := make([]byte, width*height*4)
	for i := 0; i < len(rgba); i += 4 {
		rgba[i] = r
		rgba[i+1] = g
		rgba[i+2] = b
		rgba[i+3] = a
	}
	return rgba
}

type recordingGraphicsPresenter struct {
	caps         runtimeplatform.GraphicsCapabilities
	presentCalls int
	clearCalls   int
	failPresent  bool
	requests     []runtimeplatform.DrawImageRequest
}

func (p *recordingGraphicsPresenter) Capabilities() runtimeplatform.GraphicsCapabilities {
	caps := p.caps
	if caps.Mode == runtimeplatform.GraphicsModeNone {
		caps.Mode = runtimeplatform.GraphicsModeKitty
		caps.Reliable = true
		caps.SupportsPlacement = true
		caps.SupportsReplace = true
		caps.SupportsDelete = true
	}
	return caps
}

func (p *recordingGraphicsPresenter) Present(req runtimeplatform.DrawImageRequest) (string, error) {
	p.presentCalls++
	p.requests = append(p.requests, req)
	if p.failPresent {
		return "", os.ErrInvalid
	}
	if req.ID != "" {
		return req.ID, nil
	}
	return "generated", nil
}

func (p *recordingGraphicsPresenter) Replace(id string, req runtimeplatform.DrawImageRequest) error {
	if id == "" {
		return os.ErrInvalid
	}
	return nil
}

func (p *recordingGraphicsPresenter) Delete(id string) error {
	if id == "" {
		return os.ErrInvalid
	}
	return nil
}

func (p *recordingGraphicsPresenter) Clear() error {
	p.clearCalls++
	return nil
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()

	os.Stdout = w
	defer func() {
		os.Stdout = originalStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("stdout writer close error = %v", err)
	}

	bytes, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	return string(bytes)
}
