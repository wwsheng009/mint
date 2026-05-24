package framework

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/framework/debug"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/framework/theme"
	aiservice "github.com/wwsheng009/mint/internal/ai/service"
	"github.com/wwsheng009/mint/internal/log"
	runtimepkg "github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/bridge/actionbridge"
	"github.com/wwsheng009/mint/runtime/core"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/input"
	"github.com/wwsheng009/mint/runtime/interaction"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/render"
	"github.com/wwsheng009/mint/runtime/selection"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// AppState 应用状态
type AppState int

const (
	StateCreated AppState = iota
	StateInitializing
	StateRunning
	StatePaused
	StateStopping
	StateStopped
	StateError
)

// InteractionMode controls how mouse/text selection behaves at runtime.
//
// - Interactive: regular app interaction (mouse capture on)
// - AppSelection: app-managed selection via runtime/selection (mouse capture on)
// - TerminalSelection: terminal-native text selection (mouse capture off)
type InteractionMode int

const (
	InteractionModeInteractive InteractionMode = iota
	InteractionModeAppSelection
	InteractionModeTerminalSelection
)

func (m InteractionMode) String() string {
	switch m {
	case InteractionModeInteractive:
		return "interactive"
	case InteractionModeAppSelection:
		return "app_selection"
	case InteractionModeTerminalSelection:
		return "terminal_selection"
	default:
		return "unknown"
	}
}

// App 主应用程序
type App struct {
	// 组件树
	root component.Node

	// 事件
	router       *frameworkevent.Router
	keyMap       *frameworkevent.KeyMap
	pump         *frameworkevent.Pump
	eventFilter  func(frameworkevent.Event) bool // 事件过滤器回调，返回 false 表示拦截
	focusManager *rtui.FiberFocusManager         // Focus manager for KeyMsg routing (Fiber-first)

	// ============================================================================
	// Phase 1: Action 系统 - 统一消息传播机制
	// ============================================================================
	actionRouter    *action.Router          // Action 分发器
	actionBridge    *actionbridge.Bridge    // Fiber → Action 桥接器
	inputProcessor  *action.InputProcessor  // Msg → Action 转换器
	scopeDispatcher *action.ScopeDispatcher // Scope-based action dispatcher
	// legacyMode is DEPRECATED - Action system is now the primary path
	// Set to true only for debugging/fallback purposes
	legacyMode bool // 是否启用兼容模式（默认 false）

	// 自定义事件源（测试时使用，如 MockSandbox）
	customSource frameworkevent.EventSource
	inputReader  platform.InputReader

	// 生命周期
	state int32 // accessed atomically; use AppState constants cast to int32
	quit  chan struct{}
	dirty bool
	// activeTickables caches whether the current fiber tree contains any instance
	// that still wants periodic ticks. This avoids full-tree scans on idle frames.
	activeTickables bool

	// AI / external invoke support
	aiMu           sync.RWMutex
	aiService      *aiservice.Service
	invokeQ        chan invokeRequest
	invokeDone     chan struct{}
	invokeDoneOnce sync.Once
	closeOnce      sync.Once
	renderSeq      uint64

	// 终端尺寸
	terminalWidth  int
	terminalHeight int

	// 配置尺寸（用户通过 WithWidth/Height 设置的固定大小）
	// 这个大小用于布局约束，不受终端实际大小影响
	configWidth  int
	configHeight int

	// 首次渲染标记
	firstRender bool

	// ============================================================================
	// 渲染器 - 使用新的双缓冲 Renderer (idea1 集成)
	// ============================================================================
	// Renderer 提供双缓冲、diff、run merging 等优化
	renderer *paint.Renderer
	// Async renderer (enabled by default; disable via MINT_ASYNC_RENDER=false)
	asyncRenderer      *paint.AsyncRenderer
	asyncRenderEnabled bool
	asyncFrameInterval time.Duration
	graphicsPresenter  platform.GraphicsPresenter
	graphicsImagesOn   bool
	graphicsLayout     []presentedGraphicsLayer
	graphicsLayoutNext []presentedGraphicsLayer

	// 上一帧缓冲区（用于局部刷新） - deprecated，保留用于兼容
	prevBuffer [][]paint.Cell

	// 光标位置跟踪（用于强制刷新光标区域） - deprecated
	lastCursorX int
	lastCursorY int

	// 配置
	tickInterval time.Duration

	// Panic 恢复管理器
	recovery *core.Recovery

	// 渲染节流器
	throttler *render.Throttler

	// renderMu guards render() against concurrent calls (e.g. ForceRenderNow from test goroutine)
	renderMu sync.Mutex

	// 上下文管理器
	contextMgr *core.ContextManager

	// 主题管理器
	themeMgr     *theme.Manager
	themeName    string // 当前主题名称
	themeEnabled bool   // 是否启用主题系统

	// 用户数据存储（用于存储任意用户定义数据）
	userData map[string]interface{}

	// ============================================================================
	// UI Inspector 支持
	// ============================================================================
	// Inspector 实例（用于框架级覆盖层模式）
	inspector        interface{} // *inspector.StandaloneInspector (avoid import cycle)
	inspectorEnabled bool
	inspectorVisible bool

	// ============================================================================
	// HitMap 系统（Phase 1: 统一命中测试）
	// ============================================================================
	// hitMap 存储从布局树构建的命中映射表
	// 在每次渲染后构建，用于鼠标事件的快速命中测试
	hitMap *runtimeevent.HitMap

	// ============================================================================
	// InputTracker + InteractionFSM (Phase 1-3: Pressed State 解决方案)
	// ============================================================================

	// ============================================================================
	// Test Probes (testing only)
	// ============================================================================
	testMsgProbe    func(runtimemsg.Msg)
	testActionProbe func(*action.Action, bool, string)
	// Based on the archived pressed-state design:
	// docsArchive/cleanup-2026-05-19/docs/event/PRESSED_STATE_COMPLETE_SOLUTION.md
	// - InputTracker: 追踪输入状态变化，推断边缘事件
	// - InteractionContext: 全局交互状态管理，分配 Click/Cancel/ResetPressed
	inputTracker   *input.InputTracker
	interactionCtx *interaction.InteractionContext

	// 文本选择与交互模式
	selectionAdapter *selection.RuntimeAdapter
	interactionMode  InteractionMode
	hoveredFiber     *rtui.Fiber
	mouseCaptureID   uint64
	mouseCaptureBtn  runtimemsg.MouseButton
	mouseCaptureOn   bool
	mouseCaptureRef  string

	// ============================================================================
	// 调试支持
	// ============================================================================
	debugMode     bool            // 调试模式开关
	debugLogFile  string          // 调试日志文件路径
	debugRecorder *debug.Recorder // 调试记录器
}

type presentedGraphicsLayer struct {
	ID          string
	Bounds      paint.Rect
	PixelWidth  int
	PixelHeight int
	ContentHash uint64
}

// NewApp 创建新应用 (Phase 1: 初始化 Action 系统)
func NewApp() *App {
	asyncRenderEnabled, asyncFrameInterval := asyncRenderConfigFromEnv()

	app := &App{
		router:             frameworkevent.NewRouter(),
		keyMap:             frameworkevent.NewKeyMap(),
		focusManager:       rtui.NewFiberFocusManager(),                        // Fiber-first: Focus manager
		eventFilter:        func(ev frameworkevent.Event) bool { return true }, // 默认放行所有事件
		quit:               make(chan struct{}, 1),
		invokeQ:            make(chan invokeRequest, 32),
		invokeDone:         make(chan struct{}),
		tickInterval:       16 * time.Millisecond, // ~60fps
		firstRender:        true,
		throttler:          render.NewThrottler(60), // 默认 60 FPS
		contextMgr:         core.NewContextManager(context.Background()),
		userData:           make(map[string]interface{}),
		renderer:           paint.NewRenderer(80, 24), // 新增：初始化 Renderer
		asyncRenderEnabled: asyncRenderEnabled,
		asyncFrameInterval: asyncFrameInterval,

		// Phase 1: 初始化 Action 系统
		actionRouter:    action.NewRouter(nil), // 根节点稍后设置
		inputProcessor:  action.NewInputProcessor(),
		scopeDispatcher: action.NewScopeDispatcherWithName(nil, "root"), // Scope-based dispatcher
		legacyMode:      false,                                          // Action 系统优先，legacy 仅用于调试

		// Phase 1-3: Pressed State 解决方案
		inputTracker:    input.NewInputTracker(),
		interactionCtx:  interaction.NewInteractionContext(),
		interactionMode: InteractionModeInteractive,
	}

	// 初始化 ActionBridge (Fiber → Action 桥接器)
	app.actionBridge = actionbridge.New(app.actionRouter)

	// 设置当前 ScopeDispatcher (用于 Builder 注册 closure)
	action.SetCurrentScopeDispatcher(app.scopeDispatcher)

	// 设置 ActionBridge 的 ScopeDispatcher
	app.actionBridge.SetScopeDispatcher(app.scopeDispatcher)

	// 设置 InputProcessor 的 KeyMap
	app.inputProcessor.SetKeyMap(action.NewKeyMap())

	// Phase 4: 设置默认中间件链
	// 根据环境变量选择中间件链
	var middlewareChain *action.MiddlewareChain
	if os.Getenv("ACTION_DEBUG") == "true" {
		middlewareChain = action.DebugMiddlewareChain()
	} else if os.Getenv("ACTION_PROD") == "true" {
		middlewareChain = action.ProductionMiddlewareChain()
	} else {
		middlewareChain = action.DefaultMiddlewareChain()
	}

	app.actionRouter.SetMiddleware(middlewareChain)

	return app
}

// NewAppWithSource 创建使用自定义 EventSource 的应用 (Phase 1: 初始化 Action 系统)
// 允许测试时使用 MockSandbox 或其他事件源替代真实的平台输入
func NewAppWithSource(source frameworkevent.EventSource) *App {
	asyncRenderEnabled, asyncFrameInterval := asyncRenderConfigFromEnv()

	app := &App{
		router:             frameworkevent.NewRouter(),
		keyMap:             frameworkevent.NewKeyMap(),
		focusManager:       rtui.NewFiberFocusManager(), // Fiber-first: Focus manager
		eventFilter:        func(ev frameworkevent.Event) bool { return true },
		quit:               make(chan struct{}, 1),
		invokeQ:            make(chan invokeRequest, 32),
		invokeDone:         make(chan struct{}),
		tickInterval:       16 * time.Millisecond,
		firstRender:        true,
		throttler:          render.NewThrottler(60),
		contextMgr:         core.NewContextManager(context.Background()),
		userData:           make(map[string]interface{}),
		renderer:           paint.NewRenderer(80, 24),
		asyncRenderEnabled: asyncRenderEnabled,
		asyncFrameInterval: asyncFrameInterval,
		customSource:       source, // 使用自定义事件源

		// Phase 1: 初始化 Action 系统
		actionRouter:    action.NewRouter(nil),
		inputProcessor:  action.NewInputProcessor(),
		scopeDispatcher: action.NewScopeDispatcherWithName(nil, "root"), // Scope-based dispatcher
		legacyMode:      true,
		interactionMode: InteractionModeInteractive,
	}

	// 初始化 ActionBridge
	app.actionBridge = actionbridge.New(app.actionRouter)

	// 设置当前 ScopeDispatcher (用于 Builder 注册 closure)
	action.SetCurrentScopeDispatcher(app.scopeDispatcher)

	// 设置 ActionBridge 的 ScopeDispatcher
	app.actionBridge.SetScopeDispatcher(app.scopeDispatcher)

	return app
}

func asyncRenderConfigFromEnv() (bool, time.Duration) {
	asyncRenderEnabled := true
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MINT_ASYNC_RENDER"))) {
	case "", "true", "1", "yes", "on":
		asyncRenderEnabled = true
	case "false", "0", "no", "off":
		asyncRenderEnabled = false
	default:
		// Unknown value: keep default enabled.
		asyncRenderEnabled = true
	}

	asyncFrameInterval := 16 * time.Millisecond // default ~60fps
	if fpsStr := os.Getenv("MINT_ASYNC_FPS"); fpsStr != "" {
		if fps, err := strconv.Atoi(fpsStr); err == nil && fps > 0 {
			asyncFrameInterval = time.Second / time.Duration(fps)
		}
	}
	return asyncRenderEnabled, asyncFrameInterval
}

// SetDebugMode 设置调试模式
func (a *App) SetDebugMode(enabled bool) {
	a.debugMode = enabled
	if enabled && a.debugRecorder == nil {
		logFile := a.debugLogFile
		if logFile == "" {
			logFile = fmt.Sprintf("tui_debug_%s.log", time.Now().Format("20060102_150405"))
		}
		recorder, err := debug.NewRecorder(logFile)
		if err != nil {
			log.UILogger.IfEnabled().Debug("Failed to create debug recorder: %v", err)
			return
		}
		a.debugRecorder = recorder
		log.UILogger.IfEnabled().Debug("Debug mode enabled, logging to: %s", logFile)
	}
}

// IsDebugMode 检查是否在调试模式
func (a *App) IsDebugMode() bool {
	return a.debugMode
}

// ============================================================================
// 事件过滤器配置
// ============================================================================

// SetEventFilter 设置事件过滤器回调
// 返回 false 表示拦截该事件，不再继续处理
func (a *App) SetEventFilter(filter func(frameworkevent.Event) bool) {
	a.eventFilter = filter
}

// ClearEventFilter 清除事件过滤器
func (a *App) ClearEventFilter() {
	a.eventFilter = func(ev frameworkevent.Event) bool { return true }
}

// ============================================================================
// Panic 恢复配置
// ============================================================================

// EnableRecovery 启用 panic 恢复
func (a *App) EnableRecovery() {
	if a.recovery == nil {
		a.recovery = core.NewRecovery(a)
	}
}

// SetPanicLog 设置 panic 日志文件
func (a *App) SetPanicLog(filename string) error {
	a.EnableRecovery()
	return a.recovery.EnablePanicLog(filename)
}

// AddPanicHandler 添加 panic 处理器
func (a *App) AddPanicHandler(handler core.PanicHandler) {
	a.EnableRecovery()
	a.recovery.AddHandler(handler)
}

// ============================================================================
// 渲染节流配置
// ============================================================================

// SetFPS 设置目标帧率
func (a *App) SetFPS(fps int) {
	a.throttler.SetFPS(fps)
	a.tickInterval = time.Second / time.Duration(a.throttler.FPS())
}

// FPS 获取当前帧率
func (a *App) FPS() int {
	return a.throttler.FPS()
}

// ActualFPS 获取实际帧率
func (a *App) ActualFPS() float64 {
	return a.throttler.ActualFPS()
}

// EnableAdaptiveFPS 启用自适应帧率
func (a *App) EnableAdaptiveFPS(enable bool) {
	a.throttler.EnableAdaptive(enable)
}

// GetRenderStats 获取渲染统计信息
func (a *App) GetRenderStats() render.Stats {
	return a.throttler.Stats()
}

// ForceRender 强制下次渲染（跳过节流限制）
func (a *App) ForceRender() {
	a.throttler.ForceRender()
	a.dirty = true
}

// ============================================================================
// 主题系统配置
// ============================================================================

// InitTheme 初始化主题系统
// 如果未指定主题名称，则使用主题包中的默认主题。
func (a *App) InitTheme(themeName string) error {
	mgr, err := theme.InitThemes(themeName)
	if err != nil {
		return fmt.Errorf("failed to initialize theme: %w", err)
	}
	a.themeMgr = mgr
	a.themeName = mgr.Current().Name
	a.themeEnabled = true
	return nil
}

// SetTheme 切换主题
func (a *App) SetTheme(name string) error {
	if a.themeMgr == nil {
		return errors.New("theme manager not initialized, call InitTheme first")
	}
	if err := a.themeMgr.Set(name); err != nil {
		return err
	}
	a.themeName = name
	a.dirty = true // 触发重绘
	return nil
}

// GetTheme 获取当前主题名称
func (a *App) GetTheme() string {
	return a.themeName
}

// ThemeManager 获取主题管理器
func (a *App) ThemeManager() *theme.Manager {
	return a.themeMgr
}

// IsThemeEnabled 检查主题系统是否启用
func (a *App) IsThemeEnabled() bool {
	return a.themeEnabled
}

// SetUserData 设置用户数据
func (a *App) SetUserData(key string, value interface{}) {
	a.userData[key] = value
}

// GetUserData 获取用户数据
func (a *App) GetUserData(key string) interface{} {
	return a.userData[key]
}

// SetInteractionMode sets runtime interaction mode.
func (a *App) SetInteractionMode(mode InteractionMode) error {
	switch mode {
	case InteractionModeInteractive, InteractionModeAppSelection, InteractionModeTerminalSelection:
	default:
		return fmt.Errorf("invalid interaction mode: %d", mode)
	}

	a.interactionMode = mode
	a.updateHoveredFiber(nil, nil)

	// Keep selection adapter state consistent.
	if mode == InteractionModeAppSelection {
		adapter := a.ensureSelectionAdapter()
		adapter.SetEnabled(true)
	} else if a.selectionAdapter != nil {
		a.selectionAdapter.SetEnabled(false)
		a.selectionAdapter.ClearSelection()
	}

	// Ctrl+C should copy selection in app-selection mode, not quit app.
	if a.pump != nil {
		a.pump.SetCtrlCAsQuit(mode != InteractionModeAppSelection)
	}

	// TerminalSelection means terminal-native selection: disable mouse capture.
	if err := a.applyMouseCaptureForMode(); err != nil {
		return err
	}

	a.dirty = true
	return nil
}

// GetInteractionMode returns current runtime interaction mode.
func (a *App) GetInteractionMode() InteractionMode {
	return a.interactionMode
}

// CycleInteractionMode cycles through all interaction modes and returns the new one.
func (a *App) CycleInteractionMode() (InteractionMode, error) {
	next := InteractionModeInteractive
	switch a.interactionMode {
	case InteractionModeInteractive:
		next = InteractionModeAppSelection
	case InteractionModeAppSelection:
		next = InteractionModeTerminalSelection
	case InteractionModeTerminalSelection:
		next = InteractionModeInteractive
	}
	return next, a.SetInteractionMode(next)
}

func (a *App) updateMouseHoverState(mouseMsg *runtimemsg.MouseMsg) {
	if mouseMsg == nil || mouseMsg.Action != runtimemsg.MouseActionMove {
		return
	}
	next := mouseTargetFiber(mouseMsg)
	a.updateHoveredFiber(next, mouseMsg)
}

func (a *App) updateHoveredFiber(next *rtui.Fiber, payload interface{}) {
	if a.hoveredFiber == next {
		return
	}
	if a.actionBridge != nil && a.hoveredFiber != nil && a.shouldDispatchToFiberTarget(a.hoveredFiber) {
		if a.actionBridge.DispatchFromFiber(a.hoveredFiber, action.ActionMouseLeave, payload) {
			a.dirty = true
		}
	}
	if a.actionBridge != nil && next != nil && a.shouldDispatchToFiberTarget(next) {
		if a.actionBridge.DispatchFromFiber(next, action.ActionMouseEnter, payload) {
			a.dirty = true
		}
	}
	a.hoveredFiber = next
}

func (a *App) ensureSelectionAdapter() *selection.RuntimeAdapter {
	if a.selectionAdapter == nil {
		a.selectionAdapter = selection.NewRuntimeAdapter()
	}
	return a.selectionAdapter
}

func (a *App) wantsMouseCapture() bool {
	return a.interactionMode != InteractionModeTerminalSelection
}

func (a *App) applyMouseCaptureForMode() error {
	if a.inputReader == nil {
		return nil
	}
	controller, ok := a.inputReader.(platform.MouseCaptureController)
	if !ok {
		return nil
	}
	return controller.SetMouseCaptureEnabled(a.wantsMouseCapture())
}

// ============================================================================
// 上下文管理
// ============================================================================

// Context 获取应用上下文
func (a *App) Context() context.Context {
	return a.contextMgr.Context()
}

// Shutdown 优雅关闭
func (a *App) Shutdown(timeout time.Duration) error {
	atomic.StoreInt32(&a.state, int32(StateStopping))
	return a.contextMgr.Shutdown(timeout)
}

// SetRoot 设置根组件
func (a *App) SetRoot(comp component.Node) {
	a.root = comp
	a.dirty = true

	// 使用 ComponentContext 注入运行时资源
	a.injectComponentContext(comp)
}

// createComponentContext 创建组件上下文
func (a *App) createComponentContext() *component.ComponentContext {
	ctx := component.NewComponentContext()
	ctx.SetDirtyCallback(func() {
		a.dirty = true
	})
	return ctx
}

// injectComponentContext 递归地为组件注入上下文
// 使用接口而非具体类型，遵循依赖倒置原则
func (a *App) injectComponentContext(node component.Node) {
	ctx := a.createComponentContext()

	// 1. 如果组件实现了 MountableWithContext，调用它
	if mountable, ok := node.(component.MountableWithContext); ok {
		// 获取父容器（如果有的话）
		var parent component.Container
		if p, ok := node.(interface{ GetParent() component.Container }); ok {
			parent = p.GetParent()
		}
		mountable.MountWithContext(parent, ctx)
		return
	}

	// 2. 兼容旧接口：通过辅助接口注入
	if dn, ok := node.(component.DirtyNotifiable); ok {
		dn.SetDirtyCallback(func() { a.dirty = true })
	}

	// 3. 递归处理子节点
	a.injectContextToChildren(node, ctx)
}

// injectContextToChildren 递归地为子节点注入上下文
func (a *App) injectContextToChildren(node component.Node, ctx *component.ComponentContext) {
	// 检查是否有子节点
	type childrenProvider interface {
		Children() []component.Node
	}

	if provider, ok := node.(childrenProvider); ok {
		for _, child := range provider.Children() {
			// 对子节点递归注入
			if mountable, ok := child.(component.MountableWithContext); ok {
				var parent component.Container
				if p, ok := child.(interface{ GetParent() component.Container }); ok {
					parent = p.GetParent()
				}
				mountable.MountWithContext(parent, ctx)
			} else if dn, ok := child.(component.DirtyNotifiable); ok {
				// 使用辅助接口设置 dirty callback
				dn.SetDirtyCallback(func() { a.dirty = true })
			}

			// 递归处理子节点的子节点
			a.injectContextToChildren(child, ctx)
		}
	}
}

// GetRoot 获取根组件
func (a *App) GetRoot() component.Node {
	return a.root
}

// OnKey 注册键盘事件处理
func (a *App) OnKey(key rune, handler func()) {
	a.keyMap.BindFunc(string(key), func(ev *frameworkevent.KeyEvent) {
		handler()
	})
}

// OnKeyCombo 注册快捷键处理
func (a *App) OnKeyCombo(combo string, handler func()) {
	a.keyMap.BindFunc(combo, func(ev *frameworkevent.KeyEvent) {
		handler()
	})
}

// ============================================================================
// Inspector 快捷键支持
// ============================================================================

// SetInspector 设置 Inspector 实例并自动注册渲染hook
// 这个方法用于注册 Inspector，使其可以通过 F12/Ctrl+D 切换
// Inspector会自动通过hook系统集成到渲染流程中
//
// 使用示例:
//
//	inspector := inspector.NewStandaloneInspector()
//	app.SetInspector(inspector)
//	app.SetupInspectorShortcut()
func (a *App) SetInspector(inspector interface{}) {
	a.inspector = inspector
	if i, ok := inspector.(interface{ IsVisible() bool }); ok {
		a.inspectorVisible = i.IsVisible()
	}

	// 自动注册Inspector hook到渲染系统
	// 使用接口避免import循环
	a.registerInspectorHook(inspector)
}

// GetInspector 返回当前的 Inspector 实例
func (a *App) GetInspector() interface{} {
	return a.inspector
}

// RegisterGlobalHandler registers a global action handler with the action router.
// Global handlers are called for actions without specific targets.
// This allows UI components to register cross-cutting concerns (like Modal's ESC key handling).
func (a *App) RegisterGlobalHandler(handler action.GlobalActionHandler) {
	if a.actionRouter != nil {
		a.actionRouter.AddGlobalHandler(handler)
	}
}

// AddMiddleware adds a middleware to the action router's middleware chain.
// Middleware are called before action dispatch and can intercept, modify, or observe actions.
// This allows UI components to add behavior like click-outside-to-close handling.
func (a *App) AddMiddleware(middleware action.ActionMiddleware) {
	if a.actionRouter != nil {
		a.actionRouter.AddMiddleware(middleware)
	}
}

// isInspectorVisible checks if the Inspector overlay is currently visible
// This is used to determine if keyboard events should be routed to the Inspector
func (a *App) isInspectorVisible() bool {
	if a.inspector == nil {
		return false
	}
	if inspector, ok := a.inspector.(interface{ IsVisible() bool }); ok {
		return inspector.IsVisible()
	}
	return false
}

// SetupInspectorShortcut 设置 F12 快捷键来切换 Inspector 显示
// 这是一个便捷方法，用于在框架级别启用 Inspector 快捷键
//
// 使用示例:
//
//	app := framework.NewApp()
//	app.SetInspector(inspector)
//	app.SetupInspectorShortcut() // 启用 F12 切换 Inspector
func (a *App) SetupInspectorShortcut() {
	if a.inspector == nil {
		log.UILogger.IfEnabled().Debug("[APP] Warning: SetupInspectorShortcut() called but no Inspector set")
		log.UILogger.IfEnabled().Debug("[APP] Call SetInspector() first")
		return
	}

	// 注册 F12 快捷键（在大多数终端中是 F12）
	// Note: Use lowercase because SpecialKey.String() returns lowercase
	a.OnKeyCombo("f12", func() {
		a.toggleInspector()
	})

	// 也注册 Ctrl+D 作为备用快捷键（更容易输入）
	a.OnKeyCombo("ctrl+d", func() {
		a.toggleInspector()
	})

	// 注册 Alt+H/J/K/L 快捷键来移动 Inspector 面板
	// Alt+H: 向左移动
	a.OnKeyCombo("alt+h", func() {
		a.moveInspector(-2, 0)
	})
	a.OnKeyCombo("alt+left", func() {
		a.moveInspector(-2, 0)
	})

	// Alt+L: 向右移动
	a.OnKeyCombo("alt+l", func() {
		a.moveInspector(2, 0)
	})
	a.OnKeyCombo("alt+right", func() {
		a.moveInspector(2, 0)
	})

	// Alt+K: 向上移动
	a.OnKeyCombo("alt+k", func() {
		a.moveInspector(0, -1)
	})
	a.OnKeyCombo("alt+up", func() {
		a.moveInspector(0, -1)
	})

	// Alt+J: 向下移动
	a.OnKeyCombo("alt+j", func() {
		a.moveInspector(0, 1)
	})
	a.OnKeyCombo("alt+down", func() {
		a.moveInspector(0, 1)
	})

	// REMOVED: Number keys 1-5 are now handled directly by the Inspector's HandleKeyEvent
	// via the event routing fallback. This avoids interface signature mismatches and
	// allows the Inspector to handle any number of tabs dynamically.
	//
	// The routing logic at the end of handleEvent() will forward any unhandled keys
	// to the Inspector if it is visible.

	log.UILogger.IfEnabled().Debug("[APP] Inspector shortcuts registered: F12, Ctrl+D (toggle)")
	log.UILogger.IfEnabled().Debug("[APP] Panel movement: Alt+H/J/K/L or Alt+Arrow keys")
	log.UILogger.IfEnabled().Debug("[APP] Tab switching: 1-6 (handled dynamically)")
	log.UILogger.IfEnabled().Debug("[APP] Tree scroll: PgUp/PgDn, Home/End (when Elements tab active)")
}

// toggleInspector 切换 Inspector 显示状态
// 这个方法会被快捷键触发
func (a *App) toggleInspector() {
	if a.inspector == nil {
		log.UILogger.IfEnabled().Debug("[APP] Inspector not initialized, ignoring toggle")
		return
	}

	// 调用 Inspector 的 ToggleVisibility 方法
	// 注意：这里使用接口调用避免直接导入 inspector 包
	if inspectorObj, ok := a.inspector.(interface {
		ToggleVisibility()
		IsVisible() bool
	}); ok {
		inspectorObj.ToggleVisibility()
		a.inspectorVisible = inspectorObj.IsVisible()
		a.dirty = true // 触发重绘

		log.UILogger.IfEnabled().Debug("[APP] Inspector toggled: now visible=%v", a.inspectorVisible)
	}
}

// moveInspector 移动 Inspector 面板位置
func (a *App) moveInspector(dx, dy int) {
	if a.inspector == nil {
		return
	}

	// 调用 Inspector 的 Move 方法
	if inspectorObj, ok := a.inspector.(interface {
		Move(dx, dy int)
		GetPosition() (x, y int)
	}); ok {
		inspectorObj.Move(dx, dy)
		a.dirty = true // 触发重绘

		x, y := inspectorObj.GetPosition()
		log.UILogger.IfEnabled().Debug("[APP] Inspector moved to (%d, %d)", x, y)
	}
}

// switchInspectorTab 切换 Inspector 标签页
func (a *App) switchInspectorTab(tabNum int) {
	if a.inspector == nil {
		return
	}

	// 调用 Inspector 的 HandleKeyEvent 方法
	if inspectorObj, ok := a.inspector.(interface {
		HandleKeyEvent(key string, alt, ctrl, shift bool) bool
	}); ok {
		key := fmt.Sprintf("%d", tabNum)
		if inspectorObj.HandleKeyEvent(key, false, false, false) {
			a.dirty = true // 触发重绘

			log.UILogger.IfEnabled().Debug("[APP] Inspector switched to tab %d", tabNum)
		}
	}
}

// OnEvent 注册事件处理
func (a *App) OnEvent(eventType frameworkevent.EventType, handler frameworkevent.EventHandler) func() {
	return a.router.Subscribe(eventType, handler)
}

// Init 初始化应用
func (a *App) Init() error {
	if AppState(atomic.LoadInt32(&a.state)) != StateCreated {
		return errors.New("app already initialized")
	}

	atomic.StoreInt32(&a.state, int32(StateInitializing))

	// 设置默认终端尺寸
	a.terminalWidth = 80
	a.terminalHeight = 24

	// 设置路由器
	a.setupRouter()

	// 创建并启动事件泵
	if a.customSource != nil {
		// 使用自定义事件源（测试模式）
		a.pump = frameworkevent.NewPumpWithSource(a.customSource)
	} else {
		// 使用默认的平台输入源（生产模式）
		log.UILogger.IfEnabled().Debug("[APP] Init: Creating input reader")
		inputReader, err := platform.NewInputReader()
		if err != nil {
			return err
		}
		a.inputReader = inputReader
		if err := a.applyMouseCaptureForMode(); err != nil {
			return err
		}
		log.UILogger.IfEnabled().Debug("[APP] Init: Input reader created")
		a.pump = frameworkevent.NewPump(inputReader)
	}

	if a.pump != nil {
		a.pump.SetCtrlCAsQuit(a.interactionMode != InteractionModeAppSelection)
	}

	log.UILogger.IfEnabled().Debug("[APP] Init: Starting pump")
	if err := a.pump.Start(); err != nil {
		return err
	}

	// 让根组件获得焦点
	if a.root != nil {
		if focusable, ok := a.root.(interface{ OnFocus() }); ok {
			focusable.OnFocus()
		}
	}

	if a.asyncRenderEnabled {
		a.asyncRenderer = paint.NewAsyncRenderer(80, 24, paint.AsyncRendererOptions{
			FrameInterval: a.asyncFrameInterval,
			Output: func(out string) {
				if out != "" {
					fmt.Print(out)
				}
			},
		})
		a.asyncRenderer.Start()
		log.RenderLogger.IfEnabled().Debug("[APP] async renderer enabled, frame interval=%s", a.asyncFrameInterval)
	}

	atomic.StoreInt32(&a.state, int32(StateRunning))
	if aiService := a.getAIService(); aiService != nil && aiService.ShouldAutoStart() {
		if err := aiService.Start(); err != nil {
			atomic.StoreInt32(&a.state, int32(StateError))
			return err
		}
	}
	a.dirty = true

	log.UILogger.IfEnabled().Debug("[APP] Init: Complete, state=StateRunning")

	return nil
}

// setupRouter 设置事件路由
func (a *App) setupRouter() {
	// 订阅退出事件
	a.router.Subscribe(frameworkevent.EventQuit, frameworkevent.EventHandlerFunc(func(ev frameworkevent.Event) bool {
		a.Quit()
		return true
	}))
}

// Run 运行应用
func (a *App) Run() error {
	// 启用 panic 恢复（如果配置）
	if a.recovery != nil {
		defer func() {
			if r := recover(); r != nil {
				a.recovery.Handle(r)
			}
		}()
	}

	if err := a.Init(); err != nil {
		return err
	}
	defer a.Close()

	// 主循环
	ticker := time.NewTicker(a.tickInterval)
	defer ticker.Stop()

	// 使用事件泵的通道
	eventChan := a.pump.Events()
	quitAppChan := a.pump.QuitAppRequested()
	renderStartTime := time.Now()

	// DEBUG 主循环状态
	log.UILogger.Debug("[APP] Starting main loop, state=%d, pump running=%v",
		atomic.LoadInt32(&a.state), a.pump != nil && a.pump.IsRunning())
	log.UILogger.Debug("[APP] eventChan=%p, pump.Events()=%p",
		eventChan, a.pump.Events())

	for AppState(atomic.LoadInt32(&a.state)) == StateRunning {
		// 等待事件或定时器（优先处理事件）
		select {
		case req := <-a.invokeQ:
			a.handleInvokeRequest(req)
			if a.dirty {
				renderStartTime = time.Now()
				a.render()
				a.throttler.RecordFrameTime(time.Since(renderStartTime))
			}
		case msg := <-eventChan:
			if msg == nil {
				// 通道关闭，退出
				break
			}

			// Drain all pending events to prevent backlog
			// Process keyboard events immediately, coalesce mouse move events
			var eventsToProcess []runtimemsg.Msg
			var latestMouseMove *runtimemsg.MouseMsg

			// Add the first event we got
			eventsToProcess = append(eventsToProcess, msg)

			// Drain the channel non-blockingly
			for {
				select {
				case extraMsg := <-eventChan:
					if extraMsg == nil {
						break
					}
					log.MessageLogger.IfEnabled().Debug("[APP] Msg from channel: Type=%v Message=%v", msg.Type(), msg)

					// Keyboard events: always queue
					if extraMsg.Type() == runtimemsg.MsgTypeKey {
						eventsToProcess = append(eventsToProcess, extraMsg)
						continue
					}

					// Mouse events: coalesce MouseMove
					if extraMsg.Type() == runtimemsg.MsgTypeMouse {
						if mouseMsg, ok := extraMsg.(*runtimemsg.MouseMsg); ok {
							if mouseMsg.IsMove() {
								// Keep only the latest mouse move event
								latestMouseMove = mouseMsg
								continue
							}
							// Other mouse events (Press, Release, Wheel) always queue
							eventsToProcess = append(eventsToProcess, extraMsg)
							continue
						}
					}

					// Other events: always queue
					eventsToProcess = append(eventsToProcess, extraMsg)

				default:
					// No more events to drain
					goto DRAIN_COMPLETE
				}
			}
		DRAIN_COMPLETE:

			// If we have a coalesced mouse move, add it as the last event
			if latestMouseMove != nil {
				eventsToProcess = append(eventsToProcess, latestMouseMove)
			}

			// Process all collected events
			hasUserInput := false // Track if we processed user input events
			for _, msg := range eventsToProcess {

				// Track if this is a user input event (for immediate rendering)
				if !hasUserInput {
					switch msg.Type() {
					case runtimemsg.MsgTypeKey:
						// Key events (except special keys like Tab, Enter) are user input
						if keyMsg, ok := msg.(*runtimemsg.KeyMsg); ok {
							if keyMsg.Rune != 0 {
								// Regular character key
								hasUserInput = true
							}
						}
					case runtimemsg.MsgTypeMouse:
						// Mouse events (except Move) are user input
						if mouseMsg, ok := msg.(*runtimemsg.MouseMsg); ok {
							if mouseMsg.IsPress() || mouseMsg.IsRelease() || mouseMsg.IsScroll() {
								hasUserInput = true
							}
						}
					}
				}

				// Phase 1: Try Action unified path (if enabled)
				if a.actionRouter != nil && a.inputProcessor != nil {
					a.processMsg(msg)
					// Action path marks dirty internally if needed
					continue
				}

				// Phase 2: Direct Msg routing for targeted mouse events (legacy path)
				handled := a.handleMsg(msg)
				// If not handled by direct routing, fall back to Event path
				if !handled {
					ev := frameworkevent.MsgToEvent(msg)
					if ev != nil {
						a.handleEvent(ev)
					}
				}
			}

			// Immediately render after processing user input events for better responsiveness
			// Skip throttler check for user input events to minimize input latency
			// For non-input events (resize, mouse move), use throttler to prevent excessive rendering
			needsRender := a.dirty
			if !hasUserInput {
				// Only check throttler for non-input events
				needsRender = a.dirty && a.throttler.ShouldRender()
			}
			if needsRender {
				log.UILogger.IfEnabled().Debug("[APP] Immediate render after event processing")
				renderStartTime = time.Now()
				a.render()
				a.throttler.RecordFrameTime(time.Since(renderStartTime))

				if a.inspector != nil {
					if provider, ok := a.root.(interface{ GetRenderedRoot() rtui.VNode }); ok {
						if renderedRoot := provider.GetRenderedRoot(); renderedRoot != nil {
							if inspector, ok := a.inspector.(interface{ AttachToApp(rtui.VNode) }); ok {
								inspector.AttachToApp(renderedRoot)
							}
						}
					}
				}
			}

		case <-ticker.C:
			log.UILogger.IfEnabled().Debug("[APP] Tick triggered")
			if !a.activeTickables && !a.dirty {
				continue
			}
			if a.activeTickables {
				a.handleTick()
			}

			// 处理完 tick 后，如果需要渲染则渲染
			needsRender := a.dirty && a.throttler.ShouldRender()
			log.UILogger.IfEnabled().Debug("[APP] needsRender=%v, dirty=%v", needsRender, a.dirty)
			if needsRender {
				log.UILogger.IfEnabled().Debug("[APP] Calling render()")
				renderStartTime = time.Now()
				a.render()
				a.throttler.RecordFrameTime(time.Since(renderStartTime))
				log.UILogger.IfEnabled().Debug("[APP] render() complete")

				// Pull pattern: Inspector pulls rendered tree from App after reconciliation
				// App provides GetRenderedRoot() interface, Inspector calls AttachToApp()
				if a.inspector != nil {
					if provider, ok := a.root.(interface{ GetRenderedRoot() rtui.VNode }); ok {
						if renderedRoot := provider.GetRenderedRoot(); renderedRoot != nil {
							// Inspector pulls the tree via AttachToApp
							if inspector, ok := a.inspector.(interface{ AttachToApp(rtui.VNode) }); ok {
								inspector.AttachToApp(renderedRoot)
							}
						}
					}
				}
			}

		case <-quitAppChan:
			// Ctrl+C 退出
			atomic.StoreInt32(&a.state, int32(StateStopping))
			return nil
		case <-a.quit:
			atomic.StoreInt32(&a.state, int32(StateStopping))
			return nil
		case <-a.contextMgr.Context().Done():
			atomic.StoreInt32(&a.state, int32(StateStopping))
			return nil
		}
	}

	return nil
}

// handleMsg 处理 Msg，通过 ActionBridge 路由
//
// Fiber-first Action Architecture:
// - App 只调用 ActionBridge，不直接访问 Fiber 内部
// - ActionBridge 是唯一知道 Fiber 和 Router 的模块
// - 支持语义 Action 和闭包两种模式
//
// 返回 true 表示消息已被处理
func (a *App) handleMsg(message runtimemsg.Msg) bool {
	// 鼠标事件：通过 HitMap 找到的 TargetFiber 路由
	if mouseMsg, ok := message.(*runtimemsg.MouseMsg); ok {
		if fiber := mouseTargetFiber(mouseMsg); fiber != nil {
			actionType := a.mouseActionToActionType(mouseMsg.Action)
			if actionType != "" {
				if a.actionBridge.DispatchFromFiber(fiber, actionType, mouseMsg) {
					a.dirty = true
					return true
				}
			}
		}
	}

	// 键盘事件：通过 FocusManager 路由到焦点组件
	if keyMsg, ok := message.(*runtimemsg.KeyMsg); ok {
		if a.focusManager != nil {
			focused := a.focusManager.GetCurrent()
			if focused != nil {
				actionType := a.keyMsgToActionType(keyMsg)
				if actionType != "" {
					if a.actionBridge.DispatchFromFiber(focused, actionType, keyMsg) {
						a.dirty = true
						return true
					}
				}
			}
		}
	}

	return false
}

// mouseActionToActionType 转换鼠标动作到 ActionType
func (a *App) mouseActionToActionType(mouseAction runtimemsg.MouseAction) action.ActionType {
	switch mouseAction {
	case runtimemsg.MouseActionPress:
		return action.ActionClick
	case runtimemsg.MouseActionRelease:
		return action.ActionSelect
	default:
		return ""
	}
}

// keyMsgToActionType 转换键盘消息到 ActionType
func (a *App) keyMsgToActionType(keyMsg *runtimemsg.KeyMsg) action.ActionType {
	switch {
	case keyMsg.IsEnter():
		return action.ActionEnter
	case keyMsg.IsTab():
		if keyMsg.HasShift() {
			return action.ActionNavigatePrev
		}
		return action.ActionNavigateNext
	case keyMsg.IsEscape():
		return action.ActionCancel
	case keyMsg.Rune == ' ':
		return action.ActionClick
	default:
		return action.ActionInputText
	}
}

// buildComponentRegistry 从布局树构建组件注册表（Phase 2）
//
// 已废弃：根据 fix1.md 重构，不再使用 Component Registry
// 现在使用 Instance Tree 直接处理事件
// 保留此方法仅供参考，将来会删除
/*
 */

// ============================================================================
// Phase 1: Action 系统集成
// ============================================================================

// processMsg 统一的消息处理入口
// 这是新的 Action 统一路径的核心入口
// 根据 fiber_confict.md：统一 Action Runtime
func (a *App) processMsg(msg runtimemsg.Msg) {
	if msg == nil {
		return
	}
	if a.testMsgProbe != nil {
		a.testMsgProbe(msg)
	}

	var mappedAction *action.Action
	var actionHandled bool
	var actionStage string
	defer func() {
		if a.testActionProbe != nil && mappedAction != nil {
			a.testActionProbe(mappedAction, actionHandled, actionStage)
		}
	}()

	// Global key shortcuts (OnKeyCombo/OnKey) must work in Action path too.
	if a.handleGlobalKeyShortcut(msg) {
		a.dirty = true
		return
	}

	// AppSelection mode: selection system gets first chance to consume input.
	if a.interactionMode == InteractionModeAppSelection {
		if a.dispatchSelectionEvent(msg) {
			a.dirty = true
			return
		}
	}

	// ========================================================================
	// Phase 1-3: Pressed State 解决方案
	// ========================================================================
	// Based on the archived pressed-state design:
	// docsArchive/cleanup-2026-05-19/docs/event/PRESSED_STATE_COMPLETE_SOLUTION.md
	// 1. 将 Msg 转换为 InputSnapshot
	// 2. 使用 InputTracker 推断边缘事件 (Press/Release/Move/Keyboard)
	// 3. 使用 InteractionContext 更新交互状态，分发 Click/Cancel/ResetPressed
	// ========================================================================
	snapshot := a.msgToSnapshot(msg)
	if snapshot != nil {
		intents := a.inputTracker.Update(snapshot)
		if len(intents) > 0 {
			a.interactionCtx.Update(intents, a.hitTest)
		}
	}

	// 1. 尝试转换为 Action
	act := a.inputProcessor.ProcessMsg(msg)
	mappedAction = act

	// 2. 处理无法转换的消息（系统消息）
	if act == nil {
		a.handleSystemMsg(msg)
		return
	}

	// Focus-aware remapping: in text editors, arrow/home/end should move caret,
	// not trigger global focus navigation.
	if focused := a.getFocusedFiber(); focused != nil && isTextInputFiber(focused) {
		switch act.Type {
		case action.ActionNavigateLeft:
			act.Type = action.ActionCursorLeft
		case action.ActionNavigateRight:
			act.Type = action.ActionCursorRight
		case action.ActionNavigateHome:
			act.Type = action.ActionCursorHome
		case action.ActionNavigateEnd:
			act.Type = action.ActionCursorEnd
		case action.ActionNavigateUp:
			if supportsVerticalCursorMove(focused) {
				act.Type = action.ActionCursorUp
			}
		case action.ActionNavigateDown:
			if supportsVerticalCursorMove(focused) {
				act.Type = action.ActionCursorDown
			}
		}
	}

	if mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg); ok {
		a.updateMouseHoverState(mouseMsg)
	}

	act = a.applyActionMiddlewareBefore(act)
	if act == nil {
		a.dirty = true
		actionHandled = true
		actionStage = "middleware_before"
		return
	}
	if act.IsStopped() {
		a.applyActionMiddlewareAfter(act, &action.RouterResult{
			Handled: true,
			Stopped: true,
			Phase:   action.ActionPhaseNone,
		})
		a.dirty = true
		actionHandled = true
		actionStage = "middleware_stop"
		return
	}

	// 3. 导航 Action 由焦点管理器直接处理
	// Ctrl+Tab / Ctrl+Shift+Tab should stay on the focused component so widgets
	// like tabs can use them for intra-component navigation.
	isCtrlTab := false
	if keyMsg, ok := msg.(*runtimemsg.KeyMsg); ok && keyMsg != nil && keyMsg.IsTab() && keyMsg.HasCtrl() {
		isCtrlTab = true
	}
	if act.IsNavigation() && !isCtrlTab {
		if a.handleNavigationAction(act) {
			a.applyActionMiddlewareAfter(act, &action.RouterResult{
				Handled: true,
				Phase:   action.ActionPhaseTarget,
			})
			actionHandled = true
			actionStage = "navigation"
			return
		}
	}

	// 4. 统一 Action 路由：通过 ActionBridge 分发

	// 4.1 鼠标事件：使用 MouseMsg 中的 TargetFiber
	// 通过 Payload 类型识别鼠标事件（更可靠，因为 Source 可能为空）
	mouseDispatchCleanup := func() {}
	if mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg); ok {
		if log.RenderLogger.Enabled() {
			targetTag := "<nil>"
			targetRef := ""
			if fiber := mouseTargetFiber(mouseMsg); fiber != nil {
				targetTag = fiber.Tag
				targetRef = fiber.ID
			}
			log.RenderLogger.Debug("[Mouse] msgAction=%v mappedAction=%s button=%v screen=(%d,%d) local=(%d,%d) targetNodeID=%d targetTag=%s targetRef=%s captureOn=%v captureRef=%s",
				mouseMsg.Action, act.Type, mouseMsg.Button, mouseMsg.X, mouseMsg.Y, mouseMsg.LocalX, mouseMsg.LocalY,
				mouseMsg.TargetID, targetTag, targetRef, a.mouseCaptureOn, a.mouseCaptureRef)
		}
		switch mouseMsg.Action {
		case runtimemsg.MouseActionPress:
			a.beginMouseCapture(mouseMsg)
		case runtimemsg.MouseActionRelease:
			a.applyMouseCapture(mouseMsg)
			mouseDispatchCleanup = func() {
				a.clearMouseCapture(mouseMsg.Button)
			}
		}
		if fiber := mouseTargetFiber(mouseMsg); fiber != nil {
			// Mouse click: Check if target is focusable and transfer focus
			if act.Type == action.ActionClick && a.focusManager != nil {
				if focusFiber := nearestFocusableFiber(fiber); focusFiber != nil {
					// Transfer focus to the nearest focusable ancestor by NodeID.
					focusID := fmt.Sprintf("%d", focusFiber.NodeID)
					if a.focusManager.SetFocusByID(focusID) {
						a.dirty = true
					}
				}
			}

			// 传入 act.Payload (MouseMsg) 而不是 act
			if a.shouldDispatchToFiberTarget(fiber) && a.actionBridge.DispatchFromFiber(fiber, act.Type, act.Payload) {
				mouseDispatchCleanup()
				a.applyActionMiddlewareAfter(act, &action.RouterResult{
					Handled: true,
					Phase:   action.ActionPhaseTarget,
				})
				a.dirty = true
				actionHandled = true
				actionStage = "mouse_target"
				return
			}
		}
	}

	// 4.2 键盘事件：使用焦点 Fiber
	// 非 MouseMsg 的都是键盘事件（包括特殊键和可打印字符）
	if _, ok := msg.(*runtimemsg.KeyMsg); ok {
		// 对于 ESC 键（ActionCancel），优先尝试通过 ActionRouter 的全局处理器处理
		// 这样 Modal 的 GlobalActionHandler 可以拦截并关闭 modal
		if act.Type == action.ActionCancel || act.Type == action.ActionQuit {
			result := a.actionRouter.DispatchWithoutMiddleware(act)
			a.applyActionMiddlewareAfter(act, result)
			if result.Handled {
				a.dirty = true
				actionHandled = true
				actionStage = "router_cancel"
				return
			}
		}

		// 否则正常路由到焦点元素
		if focused := a.focusManager.GetCurrent(); focused != nil {
			if a.actionBridge.DispatchFromFiber(focused, act.Type, act.Payload) {
				a.applyActionMiddlewareAfter(act, &action.RouterResult{
					Handled: true,
					Phase:   action.ActionPhaseTarget,
				})
				a.dirty = true
				actionHandled = true
				actionStage = "keyboard_target"
				return
			}
		}
	}

	// 4.3 回退：统一通过 ActionRouter 分发剩余未处理 Action。
	// 这里不能再要求必须有 TargetID；全局中间件和无目标动作
	// （例如点击菜单外部关闭 popup）也需要进入 ActionRouter。
	result := a.actionRouter.DispatchWithoutMiddleware(act)
	mouseDispatchCleanup()
	a.applyActionMiddlewareAfter(act, result)
	if result.Handled {
		a.dirty = true
		actionHandled = true
		actionStage = "router_fallback"
	} else {
		actionStage = "unhandled"
	}
}

func (a *App) applyActionMiddlewareBefore(act *action.Action) *action.Action {
	if act == nil || a.actionRouter == nil || a.actionRouter.Middleware == nil {
		return act
	}
	return a.actionRouter.Middleware.Before(act)
}

func (a *App) applyActionMiddlewareAfter(act *action.Action, result *action.RouterResult) {
	if act == nil || a.actionRouter == nil || a.actionRouter.Middleware == nil {
		return
	}
	a.actionRouter.Middleware.After(act, result)
}

func mouseTargetFiber(mouseMsg *runtimemsg.MouseMsg) *rtui.Fiber {
	if mouseMsg == nil || mouseMsg.TargetFiber == nil {
		return nil
	}
	fiber, ok := mouseMsg.TargetFiber.(*rtui.Fiber)
	if !ok || fiber == nil {
		return nil
	}
	return fiber
}

func (a *App) beginMouseCapture(mouseMsg *runtimemsg.MouseMsg) {
	if mouseMsg == nil || mouseMsg.Action != runtimemsg.MouseActionPress {
		return
	}
	fiber := mouseTargetFiber(mouseMsg)
	targetID := mouseMsg.TargetID
	if targetID == 0 && fiber != nil {
		targetID = fiber.NodeID
	}
	if targetID == 0 || fiber == nil {
		a.clearMouseCapture(runtimemsg.MouseButtonUnknown)
		return
	}
	a.mouseCaptureID = targetID
	a.mouseCaptureBtn = mouseMsg.Button
	a.mouseCaptureOn = true
	a.mouseCaptureRef = fiber.ID
	if log.RenderLogger.Enabled() {
		log.RenderLogger.Debug("[MouseCapture] begin nodeID=%d ref=%s button=%v", a.mouseCaptureID, a.mouseCaptureRef, a.mouseCaptureBtn)
	}
}

func (a *App) applyMouseCapture(mouseMsg *runtimemsg.MouseMsg) {
	if mouseMsg == nil || !a.mouseCaptureOn {
		return
	}
	if mouseMsg.Button != a.mouseCaptureBtn {
		return
	}
	if mouseMsg.Action != runtimemsg.MouseActionRelease {
		return
	}
	if a.hitMap == nil || a.mouseCaptureID == 0 {
		return
	}
	entry := a.hitMap.FindByID(a.mouseCaptureID)
	if entry == nil && a.mouseCaptureRef != "" {
		for _, candidate := range a.hitMap.AllEntries() {
			fiber, ok := candidate.TargetFiber.(*rtui.Fiber)
			if !ok || fiber == nil {
				continue
			}
			if fiber.ID != a.mouseCaptureRef {
				continue
			}
			entryCopy := candidate
			entry = &entryCopy
			break
		}
	}
	if entry == nil {
		if log.RenderLogger.Enabled() {
			log.RenderLogger.Debug("[MouseCapture] release no target found for nodeID=%d ref=%s", a.mouseCaptureID, a.mouseCaptureRef)
		}
		return
	}
	localX, localY := entry.LocalXY(mouseMsg.X, mouseMsg.Y)
	mouseMsg.TargetID = entry.NodeID
	mouseMsg.TargetFiber = entry.TargetFiber
	mouseMsg.LocalX = localX
	mouseMsg.LocalY = localY
	mouseMsg.TargetBounds = runtimepkg.Box{
		X:      entry.Bounds.X,
		Y:      entry.Bounds.Y,
		Width:  entry.Bounds.Width,
		Height: entry.Bounds.Height,
	}
	if log.RenderLogger.Enabled() {
		targetTag := "<nil>"
		if fiber := mouseTargetFiber(mouseMsg); fiber != nil {
			targetTag = fiber.Tag
		}
		log.RenderLogger.Debug("[MouseCapture] retarget release nodeID=%d ref=%s targetTag=%s local=(%d,%d)",
			mouseMsg.TargetID, a.mouseCaptureRef, targetTag, mouseMsg.LocalX, mouseMsg.LocalY)
	}
}

func (a *App) clearMouseCapture(button runtimemsg.MouseButton) {
	if !a.mouseCaptureOn {
		return
	}
	if button != runtimemsg.MouseButtonUnknown && a.mouseCaptureBtn != button {
		return
	}
	a.mouseCaptureID = 0
	a.mouseCaptureBtn = runtimemsg.MouseButtonUnknown
	a.mouseCaptureOn = false
	a.mouseCaptureRef = ""
	if log.RenderLogger.Enabled() {
		log.RenderLogger.Debug("[MouseCapture] cleared")
	}
}

func nearestFocusableFiber(fiber *rtui.Fiber) *rtui.Fiber {
	for node := fiber; node != nil; node = node.Return {
		if node.Instance == nil {
			continue
		}
		focusable, ok := node.Instance.(rtui.FocusableInstance)
		if !ok || focusable.IsDisabled() {
			continue
		}
		return node
	}
	return nil
}

func (a *App) shouldDispatchToFiberTarget(start *rtui.Fiber) bool {
	if a == nil || start == nil {
		return false
	}

	for node := start; node != nil; node = node.Return {
		if node.Instance != nil {
			if _, ok := node.Instance.(rtui.ActionHandlerInstance); ok {
				return true
			}
		}
		if node.ActionTargetID == "" {
			continue
		}
		if a.scopeDispatcher != nil && a.scopeDispatcher.HasHandler(node.ActionTargetID) {
			return true
		}
		if a.actionRouter != nil {
			if _, ok := a.actionRouter.TargetHandlers[node.ActionTargetID]; ok {
				return true
			}
		}
	}

	// Legacy router bubble/capture chains still require targeted dispatch even
	// when the Fiber path itself has no ActionHandlerInstance.
	if a.actionRouter == nil {
		return false
	}
	if a.actionRouter.Root != nil {
		return true
	}
	if len(a.actionRouter.CaptureHandlers) > 0 || len(a.actionRouter.BubbleHandlers) > 0 {
		return true
	}
	return false
}

func (a *App) handleGlobalKeyShortcut(msg runtimemsg.Msg) bool {
	if a.keyMap == nil {
		return false
	}
	keyMsg, ok := msg.(*runtimemsg.KeyMsg)
	if !ok {
		return false
	}
	ev := frameworkevent.MsgToEvent(keyMsg)
	keyEv, ok := ev.(*frameworkevent.KeyEvent)
	if !ok {
		return false
	}
	handler, found := a.keyMap.Lookup(keyEv)
	if !found || handler == nil {
		return false
	}
	return handler.HandleEvent(keyEv)
}

func (a *App) dispatchSelectionEvent(msg runtimemsg.Msg) bool {
	if a.interactionMode != InteractionModeAppSelection {
		return false
	}
	adapter := a.ensureSelectionAdapter()
	if adapter == nil || !adapter.IsEnabled() {
		return false
	}

	switch m := msg.(type) {
	case *runtimemsg.MouseMsg:
		handled := adapter.OnEvent(mouseMsgToRuntimeSelectionEvent(m))
		if !handled {
			return false
		}

		// In app-selection mode, keep mouse clicks usable for UI controls.
		// Only consume drag-move and drag-end to avoid accidental click-through
		// while selecting text.
		switch m.Action {
		case runtimemsg.MouseActionMove:
			return adapter.IsDragging()
		case runtimemsg.MouseActionRelease:
			return adapter.IsDragging()
		case runtimemsg.MouseActionPress, runtimemsg.MouseActionWheel:
			return false
		default:
			return handled
		}
	case *runtimemsg.KeyMsg:
		return adapter.OnEvent(keyMsgToRuntimeSelectionEvent(m))
	default:
		return false
	}
}

func keyMsgToRuntimeSelectionEvent(keyMsg *runtimemsg.KeyMsg) *runtimeevent.KeyEvent {
	mod := runtimeevent.KeyModifier(0)
	if keyMsg.Mod.Shift {
		mod |= runtimeevent.ModShift
	}
	if keyMsg.Mod.Ctrl {
		mod |= runtimeevent.ModCtrl
	}
	if keyMsg.Mod.Alt {
		mod |= runtimeevent.ModAlt
	}
	return &runtimeevent.KeyEvent{
		Key:     keyMsg.Rune,
		Special: keyMsg.Special,
		Type:    runtimeevent.KeyPress,
		Mod:     mod,
	}
}

func mouseMsgToRuntimeSelectionEvent(mouseMsg *runtimemsg.MouseMsg) *runtimeevent.MouseEvent {
	mouseType := runtimeevent.MouseMove
	mouseAction := runtimeevent.MouseActionMove
	switch mouseMsg.Action {
	case runtimemsg.MouseActionPress:
		mouseType = runtimeevent.MousePress
		mouseAction = runtimeevent.MouseActionPress
	case runtimemsg.MouseActionRelease:
		mouseType = runtimeevent.MouseRelease
		mouseAction = runtimeevent.MouseActionRelease
	case runtimemsg.MouseActionMove:
		mouseType = runtimeevent.MouseMove
		mouseAction = runtimeevent.MouseActionMove
	case runtimemsg.MouseActionWheel:
		mouseType = runtimeevent.MouseScroll
		mouseAction = runtimeevent.MouseActionWheel
	}

	button := runtimeevent.MouseNone
	switch mouseMsg.Button {
	case runtimemsg.MouseLeft:
		button = runtimeevent.MouseLeft
	case runtimemsg.MouseMiddle:
		button = runtimeevent.MouseMiddle
	case runtimemsg.MouseRight:
		button = runtimeevent.MouseRight
	}

	return &runtimeevent.MouseEvent{
		X:        mouseMsg.X,
		Y:        mouseMsg.Y,
		Type:     mouseType,
		Action:   mouseAction,
		TargetID: fmt.Sprintf("%d", mouseMsg.TargetID),
		LocalX:   mouseMsg.LocalX,
		LocalY:   mouseMsg.LocalY,
		Button:   button,
		Delta:    mouseMsg.Delta,
	}
}

// handleNavigationAction 处理导航 Action（Tab, 方向键等）
// 导航由焦点管理器处理，不经过 ActionRouter
func (a *App) handleNavigationAction(act *action.Action) bool {
	if a.focusManager == nil {
		return false
	}
	if keyMsg, ok := act.Payload.(*runtimemsg.KeyMsg); ok && keyMsg != nil && keyMsg.IsTab() && keyMsg.HasCtrl() {
		return false
	}

	// 根据导航类型调用焦点管理器
	var handled bool
	switch act.Type {
	case action.ActionNavigateNext:
		handled = a.focusManager.FocusNext()
	case action.ActionNavigatePrev:
		handled = a.focusManager.FocusPrev()
	case action.ActionNavigateHome:
		handled = a.focusManager.FocusFirst()
	case action.ActionNavigateEnd:
		handled = a.focusManager.FocusLast()
	// 方向键暂时不支持（FiberFocusManager 没有对应方法）
	case action.ActionNavigateUp, action.ActionNavigateDown,
		action.ActionNavigateLeft, action.ActionNavigateRight:
		handled = false
	}

	if handled {
		a.dirty = true
	}
	return handled
}

func (a *App) getFocusedFiber() *rtui.Fiber {
	if a.focusManager == nil {
		return nil
	}
	return a.focusManager.GetCurrent()
}

func (a *App) getFiberRoot() *rtui.Fiber {
	if a.root == nil {
		return nil
	}
	if provider, ok := a.root.(interface{ GetFiberRoot() *rtui.Fiber }); ok {
		return provider.GetFiberRoot()
	}
	return nil
}

func isTextInputFiber(fiber *rtui.Fiber) bool {
	if fiber == nil {
		return false
	}
	if fiber.Tag == "input" || fiber.Tag == "textarea" {
		return true
	}
	// Fallback: some component wrappers may use non-leaf tags while the runtime
	// instance still supports text cursor navigation.
	if fiber.Instance != nil {
		_, ok := fiber.Instance.(interface{ CursorPos() int })
		return ok
	}
	return false
}

func supportsVerticalCursorMove(fiber *rtui.Fiber) bool {
	if fiber == nil || fiber.Instance == nil {
		return false
	}
	_, ok := fiber.Instance.(interface {
		MoveCursorUp() bool
		MoveCursorDown() bool
	})
	return ok
}

// handleSystemMsg 处理无法转换为 Action 的系统消息
// 例如：Resize, Quit 等系统级事件
func (a *App) handleSystemMsg(msg runtimemsg.Msg) {
	// 处理 Resize 事件
	if resizeMsg, ok := msg.(*runtimemsg.ResizeMsg); ok {
		a.Resize(resizeMsg.NewWidth, resizeMsg.NewHeight)
		return
	}

	// 处理 Quit 事件
	if msg.Type() == runtimemsg.MsgTypeQuit {
		a.Quit()
		return
	}

	// 其他系统事件...
	log.UILogger.IfEnabled().Debug("[processMsg] Unhandled system message: Type=%v", msg.Type())
}

// dispatchAction 分发 Action 到 ActionRouter
func (a *App) dispatchAction(act *action.Action) *action.RouterResult {
	// 直接分发，注册表已在 render() 中构建
	return a.actionRouter.Dispatch(act)
}

// SetLegacyMode 设置兼容模式（已废弃）
// DEPRECATED: Action 系统现在是主路径，保留此方法仅用于调试
func (a *App) SetLegacyMode(enabled bool) {
	a.legacyMode = enabled
	log.UILogger.IfEnabled().Debug("[App] ⚠️  Legacy mode enabled - Action system bypassed")
}

// handleEvent 处理事件（已废弃）
// DEPRECATED: Action 系统现在是主路径，此函数仅用于调试/回退
func (a *App) handleEvent(ev frameworkevent.Event) {
	// 调试模式：记录所有事件类型
	if a.debugMode {
		log.UILogger.Debug("[EVENT] Type: %d (%s), IsMouse: %v",
			ev.Type(), ev.Type(), ev.Type().IsMouse())
	}
	// 调试模式：记录事件
	if a.debugMode && a.debugRecorder != nil {
		a.debugRecorder.RecordEvent(ev)
	}

	// 通过事件过滤器处理
	if !a.eventFilter(ev) {
		// 事件被过滤器拦截
		return
	}

	// 路由事件
	if a.router != nil {
		a.router.Route(ev)
	}

	// 窗口大小调整事件处理
	if ev.Type() == frameworkevent.EventResize {
		if resizeEv, ok := ev.(*frameworkevent.ResizeEvent); ok {
			a.Resize(resizeEv.NewWidth, resizeEv.NewHeight)
		}
		return
	}

	// 键盘事件处理
	if ev.Type() == frameworkevent.EventKeyPress {
		// DEBUG: 调试键盘事件
		log.UILogger.IfEnabled().Debug("[APP] KeyPress event received")

		// 首先检查快捷键映射
		if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
			if handler, found := a.keyMap.Lookup(keyEv); found {
				log.UILogger.Debug("[APP] KeyMap found handler for key '%s' (modifiers=%d)",
					keyEv.Key.Name, keyEv.Modifiers)
				if handler.HandleEvent(ev) {
					a.dirty = true
					return
				}
			} else {
				log.UILogger.Debug("[APP] KeyMap: No handler found for key '%s' (modifiers=%d)",
					keyEv.Key.Name, keyEv.Modifiers)
			}
		}

		// Route to Inspector if visible (NEW!)
		// Inspector gets second chance at keyboard events after registered shortcuts
		// If Inspector handles the event, it won't be sent to the VNode tree
		if a.inspector != nil && a.isInspectorVisible() {
			if inspectorObj, ok := a.inspector.(interface {
				HandleKeyEvent(key string, alt, ctrl, shift bool) bool
			}); ok {
				if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
					// Use Name for special keys, Rune for character keys
					var keyName string
					if keyEv.Key.Name != "" {
						keyName = keyEv.Key.Name
					} else if keyEv.Key.Rune > 0 {
						keyName = string(keyEv.Key.Rune)
					}

					alt := keyEv.Key.Alt
					ctrl := keyEv.Key.Ctrl
					shift := keyEv.Key.Shift

					log.UILogger.Debug("[APP] Routing key '%s' to Inspector (visible=%v, alt=%v)",
						keyName, a.isInspectorVisible(), alt)

					// Call HandleKeyEvent and check return value
					handled := inspectorObj.HandleKeyEvent(keyName, alt, ctrl, shift)

					// Always trigger re-render when Inspector processes a key event
					// This ensures UI updates even when event propagates (handled=false)
					a.dirty = true

					log.UILogger.IfEnabled().Debug("[APP] Inspector processed key '%s' (handled=%v)", keyName, handled)

					// If Inspector handled the event, don't send to VNode tree
					if handled {
						return
					}
					// If not handled, event continues to VNode tree (but re-render will happen)
				}
			}
		}

		// 然后发送到根组件
		if a.root != nil {
			log.UILogger.IfEnabled().Debug("[APP] Sending event to root, type=%T", a.root)
			// 使用 event.Component 接口检查，而不是匿名接口
			// 这样可以避免类型别名导致的类型断言失败
			if handler, ok := a.root.(frameworkevent.Component); ok {
				log.UILogger.IfEnabled().Debug("[APP] root implements Component, calling HandleEvent")
				if handler.HandleEvent(ev) {
					a.dirty = true
				}
			} else {
				log.UILogger.IfEnabled().Debug("[APP] root does NOT implement Component")
			}
		}
		return
	}

	// 鼠标事件处理 - 发送到根组件进行 hit testing
	// 支持的鼠标事件类型: EventMousePress, EventMouseRelease, EventMouseMove,
	// EventMouseWheel, EventMouseEnter, EventMouseLeave
	if ev.Type().IsMouse() {
		// DEBUG: 打印鼠标事件
		log.UILogger.IfEnabled().Debug("[handleEvent] Mouse event type=%d, sending to root Component", ev.Type())

		// Route mouse events to Inspector first (for hover tracking, overlay hit test, etc.)
		if a.inspector != nil && a.isInspectorVisible() {
			if inspectorObj, ok := a.inspector.(interface {
				HandleMouseEvent(frameworkevent.EventType, *frameworkevent.MouseEvent) bool
			}); ok {
				if mouseEv, ok := ev.(*frameworkevent.MouseEvent); ok {
					log.UILogger.IfEnabled().Debug("[APP] Routing mouse (%d,%d) to Inspector (type=%v)", mouseEv.X, mouseEv.Y, ev.Type())
					handled := inspectorObj.HandleMouseEvent(ev.Type(), mouseEv)
					a.dirty = true // refresh overlay with latest mouse info
					if handled {
						log.UILogger.IfEnabled().Debug("[handleEvent] Inspector handled mouse event, returning")
						return
					}
				}
			}
		}

		// 发送到根组件处理，由根组件负责 hit testing 和分发
		if a.root != nil {
			log.UILogger.IfEnabled().Debug("[handleEvent] Calling root.HandleEvent for mouse event")
			if handler, ok := a.root.(frameworkevent.Component); ok {
				handled := handler.HandleEvent(ev)
				log.UILogger.IfEnabled().Debug("[handleEvent] root.HandleEvent returned=%v", handled)
				if handled {
					a.dirty = true
				}
			} else {
				log.UILogger.IfEnabled().Debug("[handleEvent] root does NOT implement Component interface!")
			}
		} else {
			log.UILogger.IfEnabled().Debug("[handleEvent] root is nil!")
		}
		return
	}

	// Click 事件（已包含目标信息）
	if ev.Type() == frameworkevent.EventClick {
		if a.debugMode {
			log.UILogger.IfEnabled().Debug("[CLICK] Target: %v", ev.Target())
		}
		// 直接分发到目标组件
		if target := ev.Target(); target != nil {
			if handler, ok := target.(frameworkevent.Component); ok {
				if handler.HandleEvent(ev) {
					a.dirty = true
				}
			}
		}
		return
	}

	// 如果有目标组件，分发到组件
	if target := ev.Target(); target != nil {
		if handler, ok := target.(frameworkevent.Component); ok {
			if handler.HandleEvent(ev) {
				a.dirty = true
			}
		}
	}
}

// handleTick drives optional TickableInstance components (e.g. blinking cursors).
// Components opt in by implementing rtui.TickableInstance.
func (a *App) handleTick() {
	rootFiber := a.getFiberRoot()
	if rootFiber == nil {
		a.activeTickables = false
		return
	}

	now := time.Now()
	hasActiveTickables := false
	rtui.WalkFiberDepthFirst(rootFiber, func(fiber *rtui.Fiber) bool {
		if fiber == nil || fiber.Instance == nil {
			return true
		}

		tickable, ok := rtui.AsTickableInstance(fiber.Instance)
		if !ok || !tickable.WantsTick() {
			return true
		}
		hasActiveTickables = true

		if tickable.Tick(now) {
			a.dirty = true
		}
		return true
	})
	a.activeTickables = hasActiveTickables
}

// render 渲染界面
func (a *App) render() {
	a.renderMu.Lock()
	defer a.renderMu.Unlock()
	if a.root == nil {
		return
	}

	// 使用 V3 Paintable 接口渲染
	if paintable, ok := a.root.(component.Paintable); ok {
		// 使用 Renderer 的 back buffer
		buf := a.renderer.GetBackBuffer()

		// 清空并调整 buffer 大小（Renderer 复用 buffer）
		// buffer 大小使用实际终端大小（用于渲染）
		buf.Reset(a.terminalWidth, a.terminalHeight)

		// 布局约束使用用户配置的尺寸（用于布局计算）
		// 这样即使终端是 156x44，布局仍按用户配置的 80x24 计算
		layoutWidth, layoutHeight := a.GetConfigSize()

		ctx := component.PaintContext{
			AvailableWidth:  layoutWidth,
			AvailableHeight: layoutHeight,
			X:               0,
			Y:               0,
		}

		if log.UILogger.Enabled() {
			log.UILogger.Debug("Render: terminal=%dx%d, layout=%dx%d",
				a.terminalWidth, a.terminalHeight, layoutWidth, layoutHeight)
		}

		var scene *paint.SceneFrame
		if scenePaintable, ok := a.root.(component.ScenePaintable); ok {
			scene = scenePaintable.PaintScene(ctx, buf)
			if scene != nil && scene.Buffer == nil {
				scene.Buffer = buf
			}
		} else {
			paintable.Paint(ctx, buf)
		}

		// Apply app-managed text selection highlight (mode C).
		if a.interactionMode == InteractionModeAppSelection {
			frame := runtimepkg.Frame{
				Buffer: buf,
				Width:  buf.Width,
				Height: buf.Height,
				Dirty:  true,
			}
			a.ensureSelectionAdapter().OnRender(&frame)
		}

		// ========================================================================
		// Phase 1-3: Pressed State 解决方案 - 更新组件注册表
		// ========================================================================
		// 在每次渲染后，更新 InteractionContext 的组件注册表
		// 确保新创建的组件实例可以被 InteractionContext 访问
		a.updateInteractionInstances()

		if os.Getenv("MINT_DEBUG_TEST") == "true" {
			// Count non-empty cells after Paint
			count := 0
			for y := 0; y < buf.Height; y++ {
				for x := 0; x < buf.Width; x++ {
					if buf.Cells[y][x].Cluster != "" && buf.Cells[y][x].Cluster != " " {
						count++
					}
				}
			}
			log.UILogger.IfEnabled().Debug("[App.render] AFTER Paint: back buffer non-empty cells: %d", count)
		}

		// 调试模式：记录渲染状态
		if a.debugMode && a.debugRecorder != nil {
			a.debugRecorder.RecordRender(buf)
		}

		outputMode := os.Getenv("TUI_OUTPUT_MODE")
		noAltScreen := os.Getenv("MINT_NO_ALTERNATE_SCREEN") == "true"
		dirtyHints := a.collectPaintDirtyHints()
		sceneBypassAsync := a.shouldBypassAsyncForScene(scene)
		var sceneGraphicsLayout []presentedGraphicsLayer
		if a.graphicsPresenter != nil && (a.graphicsImagesOn || sceneBypassAsync) {
			a.graphicsLayoutNext = snapshotPresentedGraphicsInto(a.graphicsLayoutNext, scene)
			sceneGraphicsLayout = a.graphicsLayoutNext
		}
		clearGraphicsBeforeText := a.shouldClearGraphicsBeforeText(sceneGraphicsLayout)
		if clearGraphicsBeforeText {
			if err := a.clearPresentedGraphics(); err != nil {
				log.RenderLogger.IfEnabled().Debug("[APP] scene graphics pre-clear failed: %v", err)
			} else {
				a.prepareFullTextRepaintAfterGraphicsClear()
			}
		}

		sceneGraphicsChanged := a.hasSceneGraphicsChanged(sceneGraphicsLayout)
		sceneGraphicsPresent := a.shouldPresentSceneFrame(sceneGraphicsLayout)
		forceFullSceneText := a.shouldForceFullTextRenderForScene(scene, sceneGraphicsChanged || clearGraphicsBeforeText)
		if sceneBypassAsync {
			a.maskSceneImageTextRegions(buf, scene)
			if sceneGraphicsChanged {
				dirtyHints = a.appendSceneImageDirtyHints(dirtyHints, scene)
			}
			a.renderTextFrame(buf, dirtyHints, outputMode, noAltScreen, false, forceFullSceneText)
			if sceneGraphicsPresent {
				if err := a.presentSceneFrame(scene, sceneGraphicsLayout); err != nil {
					log.RenderLogger.IfEnabled().Debug("[APP] scene graphics present failed: %v", err)
					if !clearGraphicsBeforeText {
						if clearErr := a.clearPresentedGraphics(); clearErr != nil {
							log.RenderLogger.IfEnabled().Debug("[APP] scene graphics clear failed: %v", clearErr)
						}
					}
				}
			}
		} else {
			a.renderTextFrame(buf, dirtyHints, outputMode, noAltScreen, !clearGraphicsBeforeText, false)
			if !clearGraphicsBeforeText {
				if clearErr := a.clearPresentedGraphics(); clearErr != nil {
					log.RenderLogger.IfEnabled().Debug("[APP] scene graphics clear failed: %v", clearErr)
				}
			}
		}
	}

	// ============================================================================
	// Phase 1: 获取 HitMap（在每次渲染后）
	// ============================================================================
	// 优先从 RenderingPipeline 获取 HitMap（包含 Layer centering 等变换）
	// 如果不可用，回退到从布局树构建 HitMap
	if a.root != nil {
		// DEBUG: 输出 root 类型
		log.RenderLogger.IfEnabled().Debug("[APP] root type: %T", a.root)

		// 方法1：尝试从 DeclarativeNode 获取 RenderingPipeline 的 HitMap（推荐）
		// 这个 HitMap 包含了所有布局变换后的最终位置（包括 Layer centering）
		if declNode, ok := a.root.(interface{ GetHitMap() *runtimeevent.HitMap }); ok {
			a.hitMap = declNode.GetHitMap()

			if a.hitMap != nil {
				log.RenderLogger.IfEnabled().Debug("[APP] ✅ Got HitMap from RenderingPipeline: %d entries (includes layer transforms)", a.hitMap.Size())

			} else {
				log.RenderLogger.IfEnabled().Debug("[APP] ⚠️  RenderingPipeline returned nil HitMap, falling back to BuildHitMap")
			}
		}
		log.HitMapLogger.IfEnabled().Debug("[APP] Phase 2: Fiber Reconciler handles VNode → Instance reconciliation")
		// Phase 1-6: 将 HitMap 传递给 Pump 用于鼠标事件命中测试
		// HitMap already contains Instance references (enriched in DeclarativeNode.fiberFirstPaint)
		if a.pump != nil && a.hitMap != nil {
			a.pump.SetHitMap(a.hitMap)
		}
	}
	if aiService := a.getAIService(); aiService != nil {
		a.renderSeq++
		aiService.OnAfterRender(aiservice.RenderInfo{
			RenderSeq:  a.renderSeq,
			RenderedAt: time.Now(),
		})
	}
	a.refreshActiveTickables()
	a.dirty = false
	// 清除首次渲染标记
	if a.firstRender {
		a.firstRender = false
	}
}

func (a *App) refreshActiveTickables() {
	rootFiber := a.getFiberRoot()
	if rootFiber == nil {
		a.activeTickables = false
		return
	}

	active := false
	rtui.WalkFiberDepthFirst(rootFiber, func(fiber *rtui.Fiber) bool {
		if fiber == nil || fiber.Instance == nil {
			return true
		}
		tickable, ok := rtui.AsTickableInstance(fiber.Instance)
		if !ok || !tickable.WantsTick() {
			return true
		}
		active = true
		return false
	})
	a.activeTickables = active
}

// outputBuffer 输出缓冲区到终端（局部刷新优化版）
// 使用 output_diff.go 中的函数来处理差异比较和 ANSI 格式化
func (a *App) outputBuffer(buf *paint.Buffer) {
	// 调整 prevBuffer 大小（如果需要）
	a.prevBuffer = EnsurePrevBufferSize(a.prevBuffer, buf.Width, buf.Height)

	// 比较新旧 buffer，获取变化列表
	diffResult := CompareBuffers(buf, a.prevBuffer, a.lastCursorX, a.lastCursorY)

	// 更新光标位置
	a.lastCursorX = diffResult.CursorX
	a.lastCursorY = diffResult.CursorY

	// 如果没有变化，跳过输出
	if !diffResult.HasChanges {
		return
	}

	// 调试模式：记录输出
	if a.debugMode && a.debugRecorder != nil {
		log.UILogger.IfEnabled().Debug("[OUTPUT] %d changes detected", len(diffResult.Changes))
	}

	// 排序变化（从上到下，从左到右）
	SortChanges(diffResult.Changes)

	// 格式化为 ANSI 输出
	output := FormatChangesAsANSI(buf, BufferDiffResult{
		Changes:    diffResult.Changes,
		CursorX:    diffResult.CursorX,
		CursorY:    diffResult.CursorY,
		HasChanges: true,
	}, a.firstRender)

	// 一次性输出
	fmt.Print(output)

	// 更新 prevBuffer - 在所有输出完成后更新
	UpdatePrevBuffer(a.prevBuffer, buf)

	// 清除首次渲染标记
	if a.firstRender {
		a.firstRender = false
	}
}

// outputBufferDirect 输出缓冲区到终端（全量刷新版）
// 每次都输出完整的屏幕内容，不使用局部刷新优化
func (a *App) outputBufferDirect(buf *paint.Buffer) {
	var output bytes.Buffer

	noAltScreen := os.Getenv("MINT_NO_ALTERNATE_SCREEN") == "true"

	// 首次渲染时清屏（除非 MINT_NO_ALTERNATE_SCREEN=true）
	if a.firstRender {
		if !noAltScreen {
			output.WriteString("\x1b[2J") // 清屏
		}
		a.firstRender = false
	}

	// 隐藏终端光标
	output.WriteString("\x1b[?25l")

	// 调试模式：记录输出
	if a.debugMode && a.debugRecorder != nil {
		log.UILogger.IfEnabled().Debug("[OUTPUT DIRECT] about to write %d cells to terminal", buf.Height*buf.Width)
	}

	// 移动光标到左上角
	output.WriteString("\x1b[1;1H")

	// 跟踪当前样式
	var currentStyle style.Style

	// 构建输出内容 - 输出所有单元格
	for y := 0; y < buf.Height; y++ {
		skipNextCell := false
		for x := 0; x < buf.Width; x++ {
			// 如果上一个字符是宽字符，跳过它占据的下一个单元格
			if skipNextCell {
				skipNextCell = false
				continue
			}

			cell := buf.Cells[y][x]

			// 跳过宽字符的填充单元格 (Width == 0)
			if cell.Width == 0 {
				continue
			}

			// 设置字符 - extract first rune from cluster
			char := rune(0)
			for _, c := range cell.Cluster {
				char = c
				break
			}
			if char == 0 {
				char = ' '
			}

			// 应用样式（如果改变）
			if cell.Style != currentStyle {
				if currentStyle != (style.Style{}) {
					output.WriteString("\x1b[0m")
				}
				if cell.Style != (style.Style{}) {
					output.WriteString(cell.Style.ToANSI())
				}
				currentStyle = cell.Style
			}

			output.WriteRune(char)
			// 如果是宽字符，标记跳过下一个单元格
			skipNextCell = (cell.Width == 2)
		}

		// 行末重置样式并换行（除了最后一行）
		if y < buf.Height-1 {
			if currentStyle != (style.Style{}) {
				output.WriteString("\x1b[0m")
				currentStyle = style.Style{}
			}
			output.WriteString("\r\n")
		}
	}

	// 重置样式
	if currentStyle != (style.Style{}) {
		output.WriteString("\x1b[0m")
	}

	// 移动光标到末尾（避免残留）
	output.WriteString(fmt.Sprintf("\x1b[%d;%dH", buf.Height, 1))

	// 一次性输出
	fmt.Print(output.String())
}

// Quit 退出应用
func (a *App) Quit() {
	select {
	case a.quit <- struct{}{}:
	default:
	}
}

// Close 关闭应用
func (a *App) Close() error {
	a.closeOnce.Do(func() {
		atomic.StoreInt32(&a.state, int32(StateStopped))
		a.invokeDoneOnce.Do(func() {
			close(a.invokeDone)
		})
		if aiService := a.getAIService(); aiService != nil {
			_ = aiService.Stop()
		}

		// 让根组件失去焦点
		a.renderMu.Lock()
		if a.root != nil {
			if focusable, ok := a.root.(interface{ OnBlur() }); ok {
				focusable.OnBlur()
			}
		}

		if unmountable, ok := a.root.(interface{ Unmount() }); ok {
			unmountable.Unmount()
		}
		a.root = nil
		a.hitMap = nil
		a.renderMu.Unlock()

		// 停止事件泵
		if a.pump != nil {
			a.pump.Stop()
		}
		if a.asyncRenderer != nil {
			a.asyncRenderer.Stop()
			a.asyncRenderer = nil
		}
		if err := a.clearPresentedGraphics(); err != nil {
			log.RenderLogger.IfEnabled().Debug("[APP] graphics clear on close failed: %v", err)
		}

		// 调试模式：保存日志
		if a.debugMode && a.debugRecorder != nil {
			if err := a.debugRecorder.DumpToFile(); err != nil {
				log.UILogger.IfEnabled().Debug("Failed to save debug log: %v", err)
			} else {
				log.UILogger.IfEnabled().Debug("Debug log saved")
			}
		}

		// 显示终端光标
		a.ShowCursor()

		// 清屏，避免退出时残留内容
		// 除非 MINT_NO_ALTERNATE_SCREEN=true（保留输出以便复制）
		if os.Getenv("MINT_NO_ALTERNATE_SCREEN") != "true" {
			a.clearScreen()
		} else {
			// 在 NoAlternateScreen 模式下，打印一个空行分隔输出
			fmt.Println()
		}

		// 关闭 panic 恢复管理器
		if a.recovery != nil {
			a.recovery.Close()
		}
	})
	return nil
}

// ============================================================================
// Terminal 接口实现（用于 Panic Recovery）
// ============================================================================

// SetNormalMode 恢复终端正常模式
func (a *App) SetNormalMode() {
	// TUI framework 使用事件泵，这里不需要额外操作
}

// ShowCursor 显示光标
func (a *App) ShowCursor() {
	fmt.Print("\x1b[?25h")
}

// ExitAltScreen 退出备用屏幕（并清屏，因为我们并未真正使用备用屏幕模式）
func (a *App) ExitAltScreen() {
	fmt.Print("\x1b[?1049l")
	// 由于我们未使用备用屏幕模式（只用 \x1b[2J 清屏），
	// panic 时需要主动清屏以避免 TUI 内容残留
	a.clearScreen()
}

// EnableEcho 启用回显
func (a *App) EnableEcho() {
	// 事件泵会处理回显
}

// Flush 刷新输出
func (a *App) Flush() {
	os.Stdout.Sync()
}

// GetState 获取应用状态
func (a *App) GetState() AppState {
	return AppState(atomic.LoadInt32(&a.state))
}

// IsRunning 检查是否在运行
func (a *App) IsRunning() bool {
	return AppState(atomic.LoadInt32(&a.state)) == StateRunning
}

// SetTickInterval 设置定时器间隔
func (a *App) SetTickInterval(interval time.Duration) {
	a.tickInterval = interval
}

// GetSize 获取终端尺寸
func (a *App) GetSize() (width, height int) {
	return a.terminalWidth, a.terminalHeight
}

// SetConfigSize 设置用户配置的布局尺寸
// 这个尺寸用于布局约束，不受终端实际大小影响
func (a *App) SetConfigSize(width, height int) {
	a.configWidth = width
	a.configHeight = height

	log.UILogger.IfEnabled().Debug("SetConfigSize: config=%dx%d", width, height)
}

// GetConfigSize 获取用户配置的布局尺寸
func (a *App) GetConfigSize() (width, height int) {
	// 如果没有配置过，使用当前终端大小
	if a.configWidth == 0 {
		return a.terminalWidth, a.terminalHeight
	}
	return a.configWidth, a.configHeight
}

// Resize 调整终端尺寸（但不改变布局约束）
// 注意：这只是更新 buffer 大小，不会改变用户配置的布局约束
func (a *App) Resize(width, height int) {
	sizeChanged := a.terminalWidth != width || a.terminalHeight != height
	if !sizeChanged {
		return
	}
	a.terminalWidth = width
	a.terminalHeight = height
	a.dirty = true

	if log.UILogger.Enabled() {
		log.UILogger.Debug("Resize: terminal=%dx%d, config=%dx%d",
			width, height, a.configWidth, a.configHeight)
	}

	// 更新 Renderer 的尺寸（buffer 大小）
	a.renderer.Resize(width, height)
	if a.asyncRenderer != nil {
		a.asyncRenderer.Resize(width, height)
	}

	// 更新 Inspector 的屏幕大小
	if a.inspector != nil {
		if inspectorObj, ok := a.inspector.(interface {
			SetScreenSize(width, height int)
		}); ok {
			inspectorObj.SetScreenSize(width, height)
			log.UILogger.IfEnabled().Debug("[APP] Inspector screen size updated to %dx%d", width, height)
		}
	}

	// 尺寸变化时清屏，避免残留内容
	if sizeChanged && !a.firstRender {
		a.clearScreen()
	}
}

// clearScreen 清屏
func (a *App) clearScreen() {
	fmt.Print("\x1b[2J") // 清屏
	fmt.Print("\x1b[H")  // 移动光标到左上角
}

// ==============================================================================
// Renderer 访问方法 (idea1 集成)
// ==============================================================================

// GetRenderer 获取渲染器（用于高级用途）
func (a *App) GetRenderer() *paint.Renderer {
	return a.renderer
}

// SetGraphicsPresenter installs an optional graphics presenter used by
// experimental scene/image rendering paths.
func (a *App) SetGraphicsPresenter(presenter platform.GraphicsPresenter) {
	a.graphicsPresenter = presenter
	if presenter == nil {
		a.graphicsImagesOn = false
	}
}

// GetHitMap 获取当前的命中映射表（Phase 1: HitMap 集成）
// 返回从最新渲染构建的 HitMap，用于鼠标事件命中测试
//
// 返回：
//
//	*runtimeevent.HitMap - 当前的 HitMap，如果未渲染则返回 nil
//
// 示例：
//
//	hitMap := app.GetHitMap()
//	if hitMap != nil {
//	    entry := hitMap.HitTest(x, y)
//	    if entry != nil {
//	        log.UILogger.IfEnabled().Debug("Hit node: %s", entry.NodeID)
//	    }
//	}
func (a *App) GetHitMap() *runtimeevent.HitMap {
	return a.hitMap
}

// ForceFullRender 强制下一帧进行全量渲染
func (a *App) ForceFullRender() {
	a.renderer.ForceFullRender()
	a.dirty = true
}

// MarkDirty 标记需要重新渲染
func (a *App) MarkDirty() {
	a.dirty = true
}

// ForceRenderNow 强制立即渲染一次（用于测试）
// 注意：此方法仅用于测试
func (a *App) ForceRenderNow() {
	a.render()
	a.dirty = false
}

func (a *App) collectPaintDirtyHints() []paint.Rect {
	if dirtyProvider, ok := a.root.(interface{ GetPaintDirtyRects() []paint.Rect }); ok {
		return dirtyProvider.GetPaintDirtyRects()
	}
	return nil
}

func (a *App) shouldBypassAsyncForScene(scene *paint.SceneFrame) bool {
	if scene == nil || !scene.HasImageLayers() || a.graphicsPresenter == nil {
		return false
	}
	return a.graphicsPresenter.Capabilities().HasReliableGraphics()
}

func (a *App) shouldClearGraphicsBeforeText(nextLayout []presentedGraphicsLayer) bool {
	if a == nil || a.graphicsPresenter == nil || !a.graphicsImagesOn {
		return false
	}
	if a.graphicsPresenter.Capabilities().SupportsDelete {
		return false
	}

	if len(nextLayout) == 0 {
		return true
	}
	return !presentedGraphicsGeometryEqual(a.graphicsLayout, nextLayout)
}

func (a *App) hasSceneGraphicsChanged(nextLayout []presentedGraphicsLayer) bool {
	if a == nil || len(nextLayout) == 0 || a.graphicsPresenter == nil {
		return false
	}
	if !a.graphicsImagesOn {
		return true
	}
	return !presentedGraphicsLayoutEqual(a.graphicsLayout, nextLayout)
}

func (a *App) shouldPresentSceneFrame(nextLayout []presentedGraphicsLayer) bool {
	if a == nil || len(nextLayout) == 0 || a.graphicsPresenter == nil {
		return false
	}
	if !a.graphicsImagesOn {
		return true
	}
	if a.graphicsPresenter.Capabilities().UsesTerminalFramePresentation() {
		return true
	}
	return !presentedGraphicsLayoutEqual(a.graphicsLayout, nextLayout)
}

func (a *App) shouldForceFullTextRenderForScene(scene *paint.SceneFrame, graphicsUpdate bool) bool {
	if a == nil || scene == nil || !scene.HasImageLayers() || a.graphicsPresenter == nil {
		return false
	}
	return graphicsUpdate && a.graphicsPresenter.Capabilities().UsesTerminalFramePresentation()
}

func (a *App) invalidateRenderedTextState() {
	if a == nil || a.renderer == nil {
		return
	}
	a.renderer.ForceFullRender()
}

func (a *App) prepareFullTextRepaintAfterGraphicsClear() {
	if a == nil {
		return
	}
	a.invalidateRenderedTextState()
	a.firstRender = true
}

func (a *App) renderTextFrame(buf *paint.Buffer, dirtyHints []paint.Rect, outputMode string, noAltScreen bool, allowAsync bool, forceFull bool) {
	if outputMode == "direct" {
		a.outputBufferDirect(buf)
		return
	}

	if forceFull {
		a.invalidateRenderedTextState()
	}
	if a.firstRender {
		if !noAltScreen {
			fmt.Print("\x1b[2J")
		}
		fmt.Print("\x1b[?25l")
	}

	if allowAsync && a.asyncRenderer != nil {
		a.asyncRenderer.SubmitFrame(buf, dirtyHints, a.firstRender)
		return
	}

	if a.firstRender {
		a.renderer.ForceFullRender()
	}
	for _, rect := range dirtyHints {
		a.renderer.MarkDirtyRect(rect)
	}

	output := a.renderer.Render()

	if os.Getenv("MINT_DEBUG_TEST") == "true" {
		back := a.renderer.GetBackBuffer()
		front := a.renderer.GetFrontBuffer()
		backCount := 0
		frontCount := 0
		for y := 0; y < back.Height; y++ {
			for x := 0; x < back.Width; x++ {
				if back.Cells[y][x].Cluster != "" && back.Cells[y][x].Cluster != " " {
					backCount++
				}
				if front.Cells[y][x].Cluster != "" && front.Cells[y][x].Cluster != " " {
					frontCount++
				}
			}
		}
		log.UILogger.IfEnabled().Debug("[App.render] AFTER renderer.Render(): back=%d cells, front=%d cells", backCount, frontCount)
	}

	log.RenderLogger.IfEnabled().Debug("[APP] FirstRender=%v, OutputLen=%d, Dirty=%v, AllowAsync=%v", a.firstRender, len(output), a.dirty, allowAsync)

	if output != "" {
		fmt.Print(output)
	}
}

func (a *App) appendSceneImageDirtyHints(dst []paint.Rect, scene *paint.SceneFrame) []paint.Rect {
	if scene == nil || !scene.HasImageLayers() {
		return dst
	}

	for _, layer := range scene.ImageLayers {
		dst = append(dst, layer.Bounds)
	}
	return dst
}

func (a *App) maskSceneImageTextRegions(buf *paint.Buffer, scene *paint.SceneFrame) {
	if a == nil || buf == nil || scene == nil || !scene.HasImageLayers() {
		return
	}

	for _, layer := range scene.ImageLayers {
		if layer.Bounds.Width <= 0 || layer.Bounds.Height <= 0 {
			continue
		}
		x0 := appMaxInt(layer.Bounds.X, 0)
		y0 := appMaxInt(layer.Bounds.Y, 0)
		x1 := appMinInt(layer.Bounds.X+layer.Bounds.Width, buf.Width)
		y1 := appMinInt(layer.Bounds.Y+layer.Bounds.Height, buf.Height)
		maskBufferTextRegion(buf, x0, y0, x1, y1)
	}
}

func maskBufferTextRegion(buf *paint.Buffer, x0, y0, x1, y1 int) {
	if x0 >= x1 || y0 >= y1 {
		return
	}

	space := paint.Cell{Cluster: " ", Style: style.Style{}, Width: 1}
	for y := y0; y < y1; y++ {
		row := buf.Cells[y]
		if x0 > 0 {
			head := row[x0-1]
			if row[x0].IsContinuation && head.Width == 2 && head.Cluster != "" {
				row[x0-1] = space
			}
		}
		if x1 < buf.Width {
			last := row[x1-1]
			if last.Width == 2 {
				row[x1] = space
			}
		}
		for x := x0; x < x1; x++ {
			row[x] = space
		}
	}
}

func appMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func appMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (a *App) presentSceneFrame(scene *paint.SceneFrame, nextState []presentedGraphicsLayer) error {
	if scene == nil || !scene.HasImageLayers() || a.graphicsPresenter == nil {
		return nil
	}

	presentedAny := false
	for _, layer := range scene.ImageLayers {
		if !layer.HasPixels() {
			return fmt.Errorf("scene image layer %q has no pixels", layer.ID)
		}
		_, err := a.graphicsPresenter.Present(platform.DrawImageRequest{
			ID:              layer.ID,
			PixelWidth:      layer.PixelWidth,
			PixelHeight:     layer.PixelHeight,
			CellX:           layer.Bounds.X,
			CellY:           layer.Bounds.Y,
			CellWidth:       layer.Bounds.Width,
			CellHeight:      layer.Bounds.Height,
			RGBA:            append([]byte(nil), layer.RGBA...),
			AltText:         layer.AltText,
			ReplaceIfExists: true,
		})
		if err != nil {
			if presentedAny {
				a.graphicsImagesOn = true
			}
			return err
		}
		presentedAny = true
	}

	a.graphicsImagesOn = presentedAny
	if presentedAny {
		a.graphicsLayout = append(a.graphicsLayout[:0], nextState...)
	}
	return nil
}

func (a *App) clearPresentedGraphics() error {
	if a.graphicsPresenter == nil || !a.graphicsImagesOn {
		return nil
	}
	if err := a.graphicsPresenter.Clear(); err != nil {
		return err
	}
	a.graphicsImagesOn = false
	a.graphicsLayout = a.graphicsLayout[:0]
	return nil
}

func snapshotPresentedGraphics(scene *paint.SceneFrame) []presentedGraphicsLayer {
	return snapshotPresentedGraphicsInto(nil, scene)
}

func snapshotPresentedGraphicsInto(dst []presentedGraphicsLayer, scene *paint.SceneFrame) []presentedGraphicsLayer {
	dst = dst[:0]
	if scene == nil || !scene.HasImageLayers() {
		return dst
	}

	for _, layer := range scene.ImageLayers {
		dst = append(dst, presentedGraphicsLayer{
			ID:          layer.ID,
			Bounds:      layer.Bounds,
			PixelWidth:  layer.PixelWidth,
			PixelHeight: layer.PixelHeight,
			ContentHash: hashImageLayerContent(layer),
		})
	}
	return dst
}

func hashImageLayerContent(layer paint.ImageLayer) uint64 {
	h := fnv1aUint64(fnv64aOffset, uint64(layer.PixelWidth))
	h = fnv1aUint64(h, uint64(layer.PixelHeight))
	for _, b := range layer.RGBA {
		h ^= uint64(b)
		h *= fnv64aPrime
	}
	return h
}

const (
	fnv64aOffset uint64 = 14695981039346656037
	fnv64aPrime  uint64 = 1099511628211
)

func fnv1aUint64(h, v uint64) uint64 {
	for i := 0; i < 8; i++ {
		h ^= uint64(byte(v))
		h *= fnv64aPrime
		v >>= 8
	}
	return h
}

func presentedGraphicsGeometryEqual(a, b []presentedGraphicsLayer) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID {
			return false
		}
		if a[i].Bounds != b[i].Bounds {
			return false
		}
		if a[i].PixelWidth != b[i].PixelWidth || a[i].PixelHeight != b[i].PixelHeight {
			return false
		}
	}
	return true
}

func presentedGraphicsLayoutEqual(a, b []presentedGraphicsLayer) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID {
			return false
		}
		if a[i].Bounds != b[i].Bounds {
			return false
		}
		if a[i].PixelWidth != b[i].PixelWidth || a[i].PixelHeight != b[i].PixelHeight {
			return false
		}
		if a[i].ContentHash != b[i].ContentHash {
			return false
		}
	}
	return true
}

// ============================================================================
// 测试支持 - 事件注入接口
// ============================================================================

// InjectEvent 用于测试时注入事件
// 注意：此方法仅用于测试，不应用于生产代码
func (a *App) InjectEvent(raw platform.RawInput) error {
	if a.pump == nil {
		return errors.New("event pump not initialized")
	}
	if !a.pump.IsRunning() {
		log.UILogger.IfEnabled().Debug("[APP] InjectEvent: pump not running, state=%d, pump=%v", atomic.LoadInt32(&a.state), a.pump)
		return errors.New("event pump not running")
	}
	a.pump.Inject(raw)
	return nil
}

// InjectMsg injects a runtime message directly for test-only scenarios.
// This bypasses raw-input conversion and is useful when tests need precise
// control over Msg payloads such as punctuation key input.
func (a *App) InjectMsg(msg runtimemsg.Msg) error {
	if msg == nil {
		return errors.New("message is nil")
	}
	a.processMsg(msg)
	return nil
}

// GetPump 获取事件泵（用于高级测试场景）
// 注意：此方法仅用于测试
func (a *App) GetPump() *frameworkevent.Pump {
	return a.pump
}

// GetFocusManager returns the focus manager for keyboard navigation (Fiber-first)
// This is shared with DeclarativeNode to ensure focus state is synchronized
func (a *App) GetFocusManager() *rtui.FiberFocusManager {
	return a.focusManager
}

// SetFocusManagerFromDeclarativeNode syncs focusManager from DeclarativeNode
// This is called during render to ensure event routing uses the correct focus state
func (a *App) SetFocusManagerFromDeclarativeNode(fm *rtui.FiberFocusManager) {
	a.focusManager = fm
}

// SetTestMessageProbe installs a test-only callback for observing processed Msg values.
func (a *App) SetTestMessageProbe(fn func(runtimemsg.Msg)) {
	a.testMsgProbe = fn
}

// SetTestActionProbe installs a test-only callback for observing mapped Action values.
func (a *App) SetTestActionProbe(fn func(*action.Action, bool, string)) {
	a.testActionProbe = fn
}
