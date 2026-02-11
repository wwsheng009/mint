package framework

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/framework/debug"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/framework/theme"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/core"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/render"
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

// App 主应用程序
type App struct {
	// 组件树
	root component.Node

	// 事件
	router        *frameworkevent.Router
	keyMap        *frameworkevent.KeyMap
	pump          *frameworkevent.Pump
	eventFilter   func(frameworkevent.Event) bool // 事件过滤器回调，返回 false 表示拦截
	componentReg  *component.Registry             // Component registry for direct Msg routing (Phase 2)
	focusManager  *rtui.VNodeFocusManager         // Focus manager for KeyMsg routing (Phase 3)

	// 自定义事件源（测试时使用，如 MockSandbox）
	customSource frameworkevent.EventSource

	// 生命周期
	state AppState
	quit  chan struct{}
	dirty bool

	// 终端尺寸
	terminalWidth  int
	terminalHeight int

	// 首次渲染标记
	firstRender bool

	// ============================================================================
	// 渲染器 - 使用新的双缓冲 Renderer (idea1 集成)
	// ============================================================================
	// Renderer 提供双缓冲、diff、run merging 等优化
	renderer *paint.Renderer

	// 上一帧缓冲区（用于局部刷新） - deprecated，保留用于兼容
	prevBuffer [][]paint.Cell

	// 光标位置跟踪（用于强制刷新光标区域） - deprecated
	lastCursorX int
	lastCursorY int

	// 配置
	tickInterval time.Duration

	// 调试模式
	debugMode     bool
	debugRecorder *debug.Recorder
	debugLogFile  string

	// Panic 恢复管理器
	recovery *core.Recovery

	// 渲染节流器
	throttler *render.Throttler

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
}

// NewApp 创建新应用
func NewApp() *App {
	app := &App{
		router:       frameworkevent.NewRouter(),
		keyMap:       frameworkevent.NewKeyMap(),
		componentReg: component.NewRegistry(), // Phase 2: Component registry for direct Msg routing
		focusManager: rtui.NewVNodeFocusManager(), // Phase 3: Focus manager for KeyMsg routing
		eventFilter:  func(ev frameworkevent.Event) bool { return true }, // 默认放行所有事件
		quit:         make(chan struct{}, 1),
		tickInterval: 16 * time.Millisecond, // ~60fps
		firstRender:  true,
		debugMode:    os.Getenv("TUI_DEBUG") == "true",
		debugLogFile: os.Getenv("TUI_DEBUG_LOG"),
		throttler:    render.NewThrottler(60), // 默认 60 FPS
		contextMgr:   core.NewContextManager(context.Background()),
		userData:     make(map[string]interface{}),
		renderer:     paint.NewRenderer(80, 24), // 新增：初始化 Renderer
	}

	return app
}

// NewAppWithSource 创建使用自定义 EventSource 的应用
// 允许测试时使用 MockSandbox 或其他事件源替代真实的平台输入
func NewAppWithSource(source frameworkevent.EventSource) *App {
	return &App{
		router:       frameworkevent.NewRouter(),
		keyMap:       frameworkevent.NewKeyMap(),
		componentReg: component.NewRegistry(), // Phase 2: Component registry
		focusManager: rtui.NewVNodeFocusManager(), // Phase 3: Focus manager
		eventFilter:  func(ev frameworkevent.Event) bool { return true },
		quit:         make(chan struct{}, 1),
		tickInterval: 16 * time.Millisecond,
		firstRender:  true,
		debugMode:    os.Getenv("TUI_DEBUG") == "true",
		debugLogFile: os.Getenv("TUI_DEBUG_LOG"),
		throttler:    render.NewThrottler(60),
		contextMgr:   core.NewContextManager(context.Background()),
		userData:     make(map[string]interface{}),
		renderer:     paint.NewRenderer(80, 24),
		customSource: source, // 使用自定义事件源
	}
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
			fmt.Fprintf(os.Stderr, "Failed to create debug recorder: %v\n", err)
			return
		}
		a.debugRecorder = recorder
		fmt.Fprintf(os.Stderr, "Debug mode enabled, logging to: %s\n", logFile)
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
// 如果未指定主题名称，则使用默认主题 "dark"
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

// ============================================================================
// 上下文管理
// ============================================================================

// Context 获取应用上下文
func (a *App) Context() context.Context {
	return a.contextMgr.Context()
}

// Shutdown 优雅关闭
func (a *App) Shutdown(timeout time.Duration) error {
	a.state = StateStopping
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
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] Warning: SetupInspectorShortcut() called but no Inspector set\n")
			fmt.Fprintf(os.Stderr, "[APP] Call SetInspector() first\n")
		}
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

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] Inspector shortcuts registered: F12, Ctrl+D (toggle)\n")
		fmt.Fprintf(os.Stderr, "[APP] Panel movement: Alt+H/J/K/L or Alt+Arrow keys\n")
		fmt.Fprintf(os.Stderr, "[APP] Tab switching: 1-6 (handled dynamically)\n")
		fmt.Fprintf(os.Stderr, "[APP] Tree scroll: PgUp/PgDn, Home/End (when Elements tab active)\n")
	}
}

// toggleInspector 切换 Inspector 显示状态
// 这个方法会被快捷键触发
func (a *App) toggleInspector() {
	if a.inspector == nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] Inspector not initialized, ignoring toggle\n")
		}
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

		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] Inspector toggled: now visible=%v\n", a.inspectorVisible)
		}
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

		if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			x, y := inspectorObj.GetPosition()
			fmt.Fprintf(os.Stderr, "[APP] Inspector moved to (%d, %d)\n", x, y)
		}
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

			if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[APP] Inspector switched to tab %d\n", tabNum)
			}
		}
	}
}

// OnEvent 注册事件处理
func (a *App) OnEvent(eventType frameworkevent.EventType, handler frameworkevent.EventHandler) func() {
	return a.router.Subscribe(eventType, handler)
}

// Init 初始化应用
func (a *App) Init() error {
	if a.state != StateCreated {
		return errors.New("app already initialized")
	}

	a.state = StateInitializing

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] Init: Starting initialization\n")
	}

	// 设置默认终端尺寸
	a.terminalWidth = 80
	a.terminalHeight = 24

	// 设置路由器
	a.setupRouter()

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] Init: Router setup complete\n")
	}

	// 创建并启动事件泵
	if a.customSource != nil {
		// 使用自定义事件源（测试模式）
		a.pump = frameworkevent.NewPumpWithSource(a.customSource)
	} else {
		// 使用默认的平台输入源（生产模式）
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] Init: Creating input reader\n")
		}
		inputReader, err := platform.NewInputReader()
		if err != nil {
			return err
		}
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] Init: Input reader created\n")
		}
		a.pump = frameworkevent.NewPump(inputReader)
	}

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] Init: Starting pump\n")
	}
	if err := a.pump.Start(); err != nil {
		return err
	}
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] Init: Pump started\n")
	}

	// 让根组件获得焦点
	if a.root != nil {
		if focusable, ok := a.root.(interface{ OnFocus() }); ok {
			focusable.OnFocus()
		}
	}

	a.state = StateRunning
	a.dirty = true

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] Init: Complete, state=StateRunning\n")
	}

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
	renderStartTime := time.Now()

	// DEBUG 主循环状态
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] Starting main loop, state=%d, pump running=%v\n",
			a.state, a.pump != nil && a.pump.IsRunning())
		fmt.Fprintf(os.Stderr, "[APP] eventChan=%p, pump.Events()=%p\n",
			eventChan, a.pump.Events())
	}

	for a.state == StateRunning {
		// 等待事件或定时器（优先处理事件）
		select {
		case msg := <-eventChan:
			if msg == nil {
				// 通道关闭，退出
				break
			}
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[APP] Msg from channel: Type=%v\n", msg.Type())
			}

			// Phase 2: Direct Msg routing for targeted mouse events
			handled := a.handleMsg(msg)

			// If not handled by direct routing, fall back to Event path
			if !handled {
				ev := frameworkevent.MsgToEvent(msg)
				if ev != nil {
					a.handleEvent(ev)
				}
			}

		case <-ticker.C:
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[APP] Tick triggered\n")
			}
			a.handleTick()

			// 处理完 tick 后，如果需要渲染则渲染
			needsRender := a.dirty && a.throttler.ShouldRender()
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[APP] needsRender=%v, dirty=%v\n", needsRender, a.dirty)
			}
			if needsRender {
				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] Calling render()\n")
				}
				renderStartTime = time.Now()
				a.render()
				a.throttler.RecordFrameTime(time.Since(renderStartTime))
				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] render() complete\n")
				}
			}

		case <-a.quit:
			a.state = StateStopping
			return nil
		case <-a.contextMgr.Context().Done():
			a.state = StateStopping
			return nil
		}
	}

	return nil
}

// handleMsg 直接处理 Msg（Phase 2/3: 绕过 Event 系统直接路由）
//
// 返回 true 表示消息已被处理，false 表示需要回退到 Event 系统
func (a *App) handleMsg(message runtimemsg.Msg) bool {
	if a.componentReg == nil {
		return false
	}

	// 处理带 TargetID 的鼠标消息（直接路由）
	if mouseMsg, ok := message.(*runtimemsg.MouseMsg); ok {
		if mouseMsg.TargetID != "" {
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[APP] Direct routing: MouseMsg → %s\n", mouseMsg.TargetID)
			}

			// 从组件注册表查找目标组件
			component := a.componentReg.Lookup(mouseMsg.TargetID)
			if component != nil {
				// 调用组件的 Update 方法
				cmd := component.Update(mouseMsg)
				if cmd != nil {
					// TODO: 执行 Cmd（需要实现 Cmd 执行系统）
					if os.Getenv("TUI_DEBUG_UI") == "true" {
						fmt.Fprintf(os.Stderr, "[APP] Component returned Cmd: %v\n", cmd)
					}
				}

				// 标记需要重新渲染
				a.dirty = true
				return true // 消息已处理
			}

			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[APP] Component not found: %s\n", mouseMsg.TargetID)
			}
		}
	}

	// Phase 3: 处理键盘消息（通过焦点管理器路由）
	if keyMsg, ok := message.(*runtimemsg.KeyMsg); ok {
		if a.focusManager != nil {
			// 获取当前焦点组件
			focused := a.focusManager.GetCurrent()
			if focused != nil {
				// 检查焦点组件是否实现 Updater 接口
				if updater, ok := focused.(component.Updater); ok {
					if os.Getenv("TUI_DEBUG_UI") == "true" {
						focusID := focused.GetFocusID()
						fmt.Fprintf(os.Stderr, "[APP] Direct routing: KeyMsg → focused component %s\n", focusID)
					}

					// 调用焦点组件的 Update 方法
					cmd := updater.Update(keyMsg)
					if cmd != nil {
						// TODO: 执行 Cmd
						if os.Getenv("TUI_DEBUG_UI") == "true" {
							fmt.Fprintf(os.Stderr, "[APP] Focused component returned Cmd: %v\n", cmd)
						}
					}

					// 标记需要重新渲染
					a.dirty = true
					return true // 消息已处理
				}
			}
		}
	}

	// 其他情况回退到 Event 系统
	return false
}

// buildComponentRegistry 从布局树构建组件注册表（Phase 2）
//
// 遍历布局树，注册所有实现 Updater 接口的组件
func (a *App) buildComponentRegistry(root layout.Node) {
	if a.componentReg == nil {
		return
	}

	// 清空旧的注册表
	a.componentReg.Clear()

	// 递归遍历布局树
	var traverse func(node layout.Node)
	traverse = func(node layout.Node) {
		if node == nil {
			return
		}

		// 获取节点的 ID
		nodeID := node.ID()
		if nodeID != "" {
			// 检查是否实现了 Updater 接口
			if updater, ok := node.(component.Updater); ok {
				a.componentReg.Register(nodeID, updater)

				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] Registered component: %s\n", nodeID)
				}
			}
		}

		// 递归处理子节点
		children := node.Children()
		for _, child := range children {
			traverse(child)
		}
	}

	traverse(root)

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] Component registry built: %d components\n", a.componentReg.Size())
	}
}

// updateFocusManager 从布局树更新焦点管理器（Phase 3）
//
// 遍历布局树，收集所有实现 FocusableVNode 接口的组件
func (a *App) updateFocusManager(root layout.Node) {
	if a.focusManager == nil {
		return
	}

	// 收集所有 focusable 节点
	var focusableNodes []rtui.FocusableVNode

	var traverse func(node layout.Node)
	traverse = func(node layout.Node) {
		if node == nil {
			return
		}

		// 检查是否实现 FocusableVNode 接口
		if focusable, ok := node.(rtui.FocusableVNode); ok {
			focusableNodes = append(focusableNodes, focusable)
		}

		// 递归处理子节点
		children := node.Children()
		for _, child := range children {
			traverse(child)
		}
	}

	traverse(root)

	// 更新焦点管理器
	a.focusManager.SetFocusable(focusableNodes)

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] Focus manager updated: %d focusable nodes\n", len(focusableNodes))
	}
}

// handleEvent 处理事件
func (a *App) handleEvent(ev frameworkevent.Event) {
	// 调试模式：记录所有事件类型
	if a.debugMode {
		fmt.Fprintf(os.Stderr, "[EVENT] Type: %d (%s), IsMouse: %v\n",
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
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] KeyPress event received\n")
		}

		// 首先检查快捷键映射
		if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
			if handler, found := a.keyMap.Lookup(keyEv); found {
				if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] KeyMap found handler for key '%s' (modifiers=%d)\n",
						keyEv.Key.Name, keyEv.Modifiers)
				}
				if handler.HandleEvent(ev) {
					a.dirty = true
					return
				}
			} else {
				if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] KeyMap: No handler found for key '%s' (modifiers=%d)\n",
						keyEv.Key.Name, keyEv.Modifiers)
				}
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

					if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
						fmt.Fprintf(os.Stderr, "[APP] Routing key '%s' to Inspector (visible=%v, alt=%v)\n",
							keyName, a.isInspectorVisible(), alt)
					}

					// Call HandleKeyEvent and check return value
					handled := inspectorObj.HandleKeyEvent(keyName, alt, ctrl, shift)

					// Always trigger re-render when Inspector processes a key event
					// This ensures UI updates even when event propagates (handled=false)
					a.dirty = true

					if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
						fmt.Fprintf(os.Stderr, "[APP] Inspector processed key '%s' (handled=%v)\n", keyName, handled)
					}

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
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[APP] Sending event to root, type=%T\n", a.root)
			}
			// 使用 event.Component 接口检查，而不是匿名接口
			// 这样可以避免类型别名导致的类型断言失败
			if handler, ok := a.root.(frameworkevent.Component); ok {
				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] root implements Component, calling HandleEvent\n")
				}
				if handler.HandleEvent(ev) {
					a.dirty = true
				}
			} else {
				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] root does NOT implement Component\n")
				}
			}
		}
		return
	}

	// 鼠标事件处理 - 发送到根组件进行 hit testing
	// 支持的鼠标事件类型: EventMousePress, EventMouseRelease, EventMouseMove,
	// EventMouseWheel, EventMouseEnter, EventMouseLeave
	if ev.Type().IsMouse() {
		// DEBUG: 打印鼠标事件
		if a.debugMode {
			fmt.Fprintf(os.Stderr, "[MOUSE] Event type: %d\n", ev.Type())
		}

		// Route mouse events to Inspector first (for hover tracking, overlay hit test, etc.)
		if a.inspector != nil && a.isInspectorVisible() {
			if inspectorObj, ok := a.inspector.(interface {
				HandleMouseEvent(frameworkevent.EventType, *frameworkevent.MouseEvent) bool
			}); ok {
				if mouseEv, ok := ev.(*frameworkevent.MouseEvent); ok {
					if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
						fmt.Fprintf(os.Stderr, "[APP] Routing mouse (%d,%d) to Inspector (type=%v)\n", mouseEv.X, mouseEv.Y, ev.Type())
					}
					handled := inspectorObj.HandleMouseEvent(ev.Type(), mouseEv)
					a.dirty = true // refresh overlay with latest mouse info
					if handled {
						return
					}
				}
			}
		}

		// 发送到根组件处理，由根组件负责 hit testing 和分发
		if a.root != nil {
			if handler, ok := a.root.(frameworkevent.Component); ok {
				if handler.HandleEvent(ev) {
					a.dirty = true
				}
			}
		}
		return
	}

	// Click 事件（已包含目标信息）
	if ev.Type() == frameworkevent.EventClick {
		if a.debugMode {
			fmt.Fprintf(os.Stderr, "[CLICK] Target: %v\n", ev.Target())
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

// handleTick 处理定时器
// 光标闪烁现在由 TextInput.Paint 自己处理，不需要外部 Tick
func (a *App) handleTick() {
	// 定期触发重绘以支持光标闪烁
	// TextInput 会在 Paint 时自己检查时间并切换光标状态
	a.dirty = true
}

// render 渲染界面
func (a *App) render() {
	if a.root == nil {
		return
	}

	// 使用 V3 Paintable 接口渲染
	if paintable, ok := a.root.(component.Paintable); ok {
		// 使用 Renderer 的 back buffer
		buf := a.renderer.GetBackBuffer()

		// 清空并调整 buffer 大小（Renderer 复用 buffer）
		buf.Reset(a.terminalWidth, a.terminalHeight)

		ctx := component.PaintContext{
			AvailableWidth:  a.terminalWidth,
			AvailableHeight: a.terminalHeight,
			X:               0,
			Y:               0,
		}

		paintable.Paint(ctx, buf)

		// 调试模式：记录渲染状态
		if a.debugMode && a.debugRecorder != nil {
			a.debugRecorder.RecordRender(buf)
		}

		// 将缓冲区内容输出到终端
		// 使用环境变量控制输出模式：
		// TUI_OUTPUT_MODE=direct  使用全量刷新（绕过差异比较）
		// TUI_OUTPUT_MODE=diff    使用差异比较优化（默认）
		// TUI_OUTPUT_MODE=debug   调试模式，显示 diff 信息
		// MINT_NO_ALTERNATE_SCREEN=true  不清屏，允许复制/滚动
		outputMode := os.Getenv("TUI_OUTPUT_MODE")
		noAltScreen := os.Getenv("MINT_NO_ALTERNATE_SCREEN") == "true"
		if outputMode == "direct" {
			a.outputBufferDirect(buf)
		} else {
			// 首次渲染：清屏、隐藏光标、强制全量渲染
			// 除非 MINT_NO_ALTERNATE_SCREEN=true
			if a.firstRender {
				if !noAltScreen {
					fmt.Print("\x1b[2J") // 清屏
				}
				fmt.Print("\x1b[?25l")       // 隐藏光标
				a.renderer.ForceFullRender() // 强制全屏渲染
			}

			// 使用新的 Renderer 输出（自动 diff + run merging + 光标优化）
			output := a.renderer.Render()

			// DEBUG: 输出渲染信息（每次）
			if os.Getenv("TUI_OUTPUT_MODE") == "debug" {
				fmt.Fprintf(os.Stderr, "[APP] FirstRender=%v, OutputLen=%d, Dirty=%v\n", a.firstRender, len(output), a.dirty)
			}

			if output != "" {
				fmt.Print(output)
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
		if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] root type: %T\n", a.root)
		}

		// 方法1：尝试从 DeclarativeNode 获取 RenderingPipeline 的 HitMap（推荐）
		// 这个 HitMap 包含了所有布局变换后的最终位置（包括 Layer centering）
		if declNode, ok := a.root.(interface{ GetHitMap() *runtimeevent.HitMap }); ok {
			a.hitMap = declNode.GetHitMap()

			if a.hitMap != nil {
				if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] ✅ Got HitMap from RenderingPipeline: %d entries (includes layer transforms)\n", a.hitMap.Size())
				}
			} else {
				if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] ⚠️  RenderingPipeline returned nil HitMap, falling back to BuildHitMap\n")
				}
			}
		}

		// 方法2：如果 RenderingPipeline 的 HitMap 不可用，回退到从 layout.Node 构建
		if a.hitMap == nil {
			if layoutRoot, ok := a.root.(layout.Node); ok {
				a.hitMap = runtimeevent.BuildHitMap(layoutRoot)

				if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] ⚠️  HitMap built from layout.Node: %d entries (may not include layer transforms)\n", a.hitMap.Size())
				}
			} else if vnodeRoot, ok := a.root.(rtui.VNode); ok {
				// 通过 VNodeAdapter 将 VNode 转换为 layout.Node
				layoutAdapter := rtui.AsLayoutNode(vnodeRoot)
				a.hitMap = runtimeevent.BuildHitMap(layoutAdapter)

				if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] ⚠️  HitMap built from VNode: %d entries (may not include layer transforms)\n", a.hitMap.Size())
				}
			} else {
				// DEBUG: root 不是 layout.Node 也不是 VNode
				if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
					fmt.Fprintf(os.Stderr, "[APP] root is neither layout.Node nor VNode, type=%T\n", a.root)
				}
			}
		}

		// Phase 1-6: 将 HitMap 传递给 Pump 用于鼠标事件命中测试
		if a.pump != nil && a.hitMap != nil {
			a.pump.SetHitMap(a.hitMap)
		}

		// Phase 2: 构建组件注册表（从布局树提取 Updater 组件）
		if layoutRoot, ok := a.root.(layout.Node); ok {
			a.buildComponentRegistry(layoutRoot)
		} else if vnodeRoot, ok := a.root.(rtui.VNode); ok {
			layoutAdapter := rtui.AsLayoutNode(vnodeRoot)
			a.buildComponentRegistry(layoutAdapter)
		}

		// Phase 3: 更新焦点管理器（从布局树提取 Focusable 组件）
		if layoutRoot, ok := a.root.(layout.Node); ok {
			a.updateFocusManager(layoutRoot)
		} else if vnodeRoot, ok := a.root.(rtui.VNode); ok {
			layoutAdapter := rtui.AsLayoutNode(vnodeRoot)
			a.updateFocusManager(layoutAdapter)
		}
	}

	a.dirty = false

	// 清除首次渲染标记
	if a.firstRender {
		a.firstRender = false
	}
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
	if a.debugMode && a.debugRecorder != nil && os.Getenv("TUI_OUTPUT_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[OUTPUT] %d changes detected\n", len(diffResult.Changes))
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
	if a.debugMode && a.debugRecorder != nil && os.Getenv("TUI_OUTPUT_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[OUTPUT DIRECT] about to write %d cells to terminal\n", buf.Height*buf.Width)
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
	a.state = StateStopped

	// 让根组件失去焦点
	if a.root != nil {
		if focusable, ok := a.root.(interface{ OnBlur() }); ok {
			focusable.OnBlur()
		}
	}

	// 停止事件泵
	if a.pump != nil {
		a.pump.Stop()
	}

	// 调试模式：保存日志
	if a.debugMode && a.debugRecorder != nil {
		if err := a.debugRecorder.DumpToFile(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save debug log: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Debug log saved\n")
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
	return a.state
}

// IsRunning 检查是否在运行
func (a *App) IsRunning() bool {
	return a.state == StateRunning
}

// SetTickInterval 设置定时器间隔
func (a *App) SetTickInterval(interval time.Duration) {
	a.tickInterval = interval
}

// GetSize 获取终端尺寸
func (a *App) GetSize() (width, height int) {
	return a.terminalWidth, a.terminalHeight
}

// Resize 调整尺寸
func (a *App) Resize(width, height int) {
	sizeChanged := a.terminalWidth != width || a.terminalHeight != height
	a.terminalWidth = width
	a.terminalHeight = height
	a.dirty = true

	// 更新 Renderer 的尺寸
	a.renderer.Resize(width, height)

	// 更新 Inspector 的屏幕大小
	if a.inspector != nil {
		if inspectorObj, ok := a.inspector.(interface {
			SetScreenSize(width, height int)
		}); ok {
			inspectorObj.SetScreenSize(width, height)
			if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[APP] Inspector screen size updated to %dx%d\n", width, height)
			}
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

// GetHitMap 获取当前的命中映射表（Phase 1: HitMap 集成）
// 返回从最新渲染构建的 HitMap，用于鼠标事件命中测试
//
// 返回：
//   *runtimeevent.HitMap - 当前的 HitMap，如果未渲染则返回 nil
//
// 示例：
//   hitMap := app.GetHitMap()
//   if hitMap != nil {
//       entry := hitMap.HitTest(x, y)
//       if entry != nil {
//           fmt.Printf("Hit node: %s\n", entry.NodeID)
//       }
//   }
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
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] InjectEvent: pump not running, state=%d, pump=%v\n", a.state, a.pump)
		}
		return errors.New("event pump not running")
	}
	a.pump.Inject(raw)
	return nil
}

// GetPump 获取事件泵（用于高级测试场景）
// 注意：此方法仅用于测试
func (a *App) GetPump() *frameworkevent.Pump {
	return a.pump
}
