// Package engine provides frame scheduling and rendering loop for TUI applications.
//
// The Engine implements:
//   - Frame-based rendering (60fps)
//   - Event-driven updates with three-phase propagation
//   - Repaint coalescing
//   - Idle detection and power saving
//   - Platform input handling
//   - Integration with runtime/event system
package engine

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/focus"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
)

// Renderable 是可渲染组件的接口
//
// 这是引擎与组件交互的核心接口，任何实现此接口的类型都可以被引擎渲染
type Renderable interface {
	// ID 返回组件唯一标识
	ID() string

	// Paint 绘制组件到缓冲区
	Paint(buf *paint.Buffer)
}

// Layoutable 是可布局组件的接口
type Layoutable interface {
	Renderable
	// Layout 计算组件布局
	Layout()
}

// Updatable 是可更新组件的接口（每帧调用）
type Updatable interface {
	Renderable
	// Update 更新组件状态
	// dt 是自上一帧的时间增量
	Update(dt time.Duration)
}

// Engine 帧调度引擎
//
// 实现了完整的主事件循环，驱动整个渲染管线
type Engine struct {
	// 渲染器
	renderer *paint.Renderer

	// 帧配置
	frameInterval time.Duration // 16ms = 60fps
	idleTimeout   time.Duration // 空闲超时时间

	// 状态
	running       atomic.Bool
	repaintNeeded atomic.Bool
	idle          atomic.Bool

	// 事件处理
	eventQueue  chan *event.EventStruct
	quit        chan struct{}
	inputReader platform.InputReader
	inputEvents chan platform.RawInput

	// 组件
	root   Renderable
	rootMu sync.RWMutex

	// LayoutNode 用于命中测试（需要由外部设置）
	layoutBoxes []runtime.LayoutBox
	layoutMu    sync.RWMutex

	// 焦点管理
	focusManager *focus.Manager

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc

	// 输出
	outputFunc func(string) // 输出函数（可注入）

	// 最后活动时间
	lastActivityTime time.Time

	// 鼠标回调
	mouseMoveCallback func(x, y int)

	// 固定大小模式 - 禁用自动调整大小
	fixedSize bool
}

// New 创建新的引擎
func New(width, height int, root Renderable) *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		renderer:      paint.NewRenderer(width, height),
		frameInterval: 16 * time.Millisecond, // ~60fps
		idleTimeout:   3 * time.Second,
		eventQueue:    make(chan *event.EventStruct, 100),
		quit:          make(chan struct{}),
		inputEvents:   make(chan platform.RawInput, 50),
		root:          root,
		ctx:           ctx,
		cancel:        cancel,
		outputFunc:    func(s string) { print(s) },
	}
}

// NewWithRenderer 创建使用自定义渲染器的引擎
func NewWithRenderer(renderer *paint.Renderer, root Renderable) *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		renderer:      renderer,
		frameInterval: 16 * time.Millisecond,
		idleTimeout:   3 * time.Second,
		eventQueue:    make(chan *event.EventStruct, 100),
		quit:          make(chan struct{}),
		inputEvents:   make(chan platform.RawInput, 50),
		root:          root,
		ctx:           ctx,
		cancel:        cancel,
		outputFunc:    func(s string) { print(s) },
	}
}

// SetOutputFunc 设置输出函数
func (e *Engine) SetOutputFunc(fn func(string)) {
	e.outputFunc = fn
}

// SetMouseMoveCallback 设置鼠标移动回调
func (e *Engine) SetMouseMoveCallback(fn func(x, y int)) {
	e.mouseMoveCallback = fn
}

// SetFrameInterval 设置帧间隔
func (e *Engine) SetFrameInterval(interval time.Duration) {
	e.frameInterval = interval
}

// SetIdleTimeout 设置空闲超时时间
func (e *Engine) SetIdleTimeout(timeout time.Duration) {
	e.idleTimeout = timeout
}

// SetLayoutBoxes 设置用于命中测试的布局框
func (e *Engine) SetLayoutBoxes(boxes []runtime.LayoutBox) {
	e.layoutMu.Lock()
	defer e.layoutMu.Unlock()
	e.layoutBoxes = boxes
}

// GetLayoutBoxes 获取当前的布局框
func (e *Engine) GetLayoutBoxes() []runtime.LayoutBox {
	e.layoutMu.RLock()
	defer e.layoutMu.RUnlock()
	return e.layoutBoxes
}

// SetFocusManager 设置焦点管理器
func (e *Engine) SetFocusManager(fm *focus.Manager) {
	e.focusManager = fm
}

// GetFocusManager 获取焦点管理器
func (e *Engine) GetFocusManager() *focus.Manager {
	return e.focusManager
}

// SetFixedSize 设置固定大小模式
// 当启用时，引擎将忽略终端大小变化事件，保持指定的缓冲区大小
func (e *Engine) SetFixedSize(fixed bool) {
	e.fixedSize = fixed
}

// IsFixedSize 检查是否处于固定大小模式
func (e *Engine) IsFixedSize() bool {
	return e.fixedSize
}

// Run 启动主循环
//
// 这是引擎的核心方法，实现了完整的事件驱动渲染循环：
// 1. 监听事件队列
// 2. 定时帧渲染（60fps）
// 3. 空闲检测（无变化时停止渲染，节省资源）
// 4. 优雅退出
// 5. 信号处理（Ctrl+C）
func (e *Engine) Run() error {
	if !e.running.CompareAndSwap(false, true) {
		return nil // 已经在运行
	}

	defer e.running.Store(false)

	// 启动输入读取器
	inputReader, err := platform.NewInputReader()
	if err != nil {
		e.cancel()
		return err
	}
	e.inputReader = inputReader

	if err := e.inputReader.Start(e.inputEvents); err != nil {
		e.cancel()
		return err
	}

	// 清理函数 - 确保总是执行
	cleanup := func() {
		if e.inputReader != nil {
			e.inputReader.Stop()
			e.inputReader = nil
		}
		// 显示光标（如果被隐藏了）
		fmt.Print("\x1b[?25h")
		// 重置终端样式
		fmt.Print("\x1b[0m")
		// 清除屏幕（可选）
		// fmt.Print("\x1b[2J")
		// 移动光标到左下角
		fmt.Print("\x1b[H")
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 启动信号处理 goroutine
	go func() {
		sig := <-sigChan
		fmt.Printf("\n[Engine] Received signal: %v, cleaning up...\n", sig)
		cleanup()
		os.Exit(0)
	}()

	ticker := time.NewTicker(e.frameInterval)

	// 记录最后活动时间
	e.lastActivityTime = time.Now()
	e.repaintNeeded.Store(true) // 初始渲染

	// 启动输入转换 goroutine
	go e.convertInputLoop()

	// 主循环
	runErr := func() error {
		for {
			select {
			case ev := <-e.eventQueue:
				// 处理事件
				e.handleEvent(ev)
				e.recordActivity()
				e.exitIdle()

			case <-ticker.C:
				// 帧渲染
				if e.repaintNeeded.Load() && !e.idle.Load() {
					e.frame()
					e.repaintNeeded.Store(false)
				}

				// 检查空闲超时
				if !e.idle.Load() && time.Since(e.lastActivityTime) > e.idleTimeout {
					e.enterIdle()
				}

			case <-e.quit:
				// 退出
				return nil

			case <-e.ctx.Done():
				// 上下文取消
				return nil
			}
		}
	}()

	// 清理
	ticker.Stop()
	cleanup()
	signal.Stop(sigChan)
	close(sigChan)

	return runErr
}

// frame 执行一帧
//
// 帧执行顺序：
// 1. 更新组件状态（如果实现 Updatable）
// 2. 布局（如果实现 Layoutable）
// 3. 绘制到 back buffer
// 4. 渲染输出
func (e *Engine) frame() {
	e.rootMu.RLock()
	root := e.root
	e.rootMu.RUnlock()

	if root == nil {
		return
	}

	// 1. 更新组件状态（如果支持）
	if updatable, ok := root.(Updatable); ok {
		updatable.Update(e.frameInterval)
	}

	// 2. 布局（如果支持）
	if layoutable, ok := root.(Layoutable); ok {
		layoutable.Layout()
	}

	// 3. 绘制到 back buffer
	buf := e.renderer.GetBackBuffer()
	root.Paint(buf)

	// 4. 渲染输出
	output := e.renderer.Render()
	if output != "" && e.outputFunc != nil {
		e.outputFunc(output)
	}
}

// handleEvent 处理事件
//
// 使用 runtime/event 的三阶段传播系统分发事件
func (e *Engine) handleEvent(ev *event.EventStruct) {
	e.layoutMu.RLock()
	boxes := e.layoutBoxes
	e.layoutMu.RUnlock()

	// 处理鼠标移动事件 - 更新位置并触发重绘
	if ev.Type() == event.EventMouseMove && ev.Mouse != nil {
		if e.mouseMoveCallback != nil {
			e.mouseMoveCallback(ev.Mouse.X, ev.Mouse.Y)
		}
		e.RequestRepaint()
		return
	}

	// 使用事件分发器进行三阶段传播
	result := event.DispatchEvent(ev, boxes)

	// 处理焦点变化
	if e.focusManager != nil {
		switch result.FocusChange {
		case event.FocusChangeNext:
			e.focusManager.FocusNext()
			e.RequestRepaint()
		case event.FocusChangePrev:
			e.focusManager.FocusPrev()
			e.RequestRepaint()
		case event.FocusChangeSpecific:
			if result.FocusTarget != "" {
				e.focusManager.FocusSpecific(result.FocusTarget)
				e.RequestRepaint()
			}
		}
	}

	// 如果事件被处理，标记需要重绘
	if result.Updated {
		e.RequestRepaint()
	}

	// 任何鼠标事件都触发重绘（用于显示悬停效果）
	if ev.Type() == event.EventMousePress || ev.Type() == event.EventMouseRelease {
		e.RequestRepaint()
	}
}

// convertInputLoop 将平台输入转换为事件
func (e *Engine) convertInputLoop() {
	for {
		select {
		case <-e.ctx.Done():
			return
		case raw := <-e.inputEvents:
			ev := e.convertRawInput(raw)
			if ev != nil {
				select {
				case e.eventQueue <- ev:
				case <-e.ctx.Done():
					return
				}
			}
		}
	}
}

// convertRawInput 将原始输入转换为事件
func (e *Engine) convertRawInput(raw platform.RawInput) *event.EventStruct {
	ev := &event.EventStruct{
		TimestampValue: raw.Timestamp,
	}

	switch raw.Type {
	case platform.InputKeyPress:
		ev.TypeValue = event.EventKeyPress
		ev.Key = &event.KeyEvent{
			Key: raw.Key,
			Mod: event.KeyModifier(raw.Modifiers),
		}

	case platform.InputKeyRelease:
		ev.TypeValue = event.EventKeyRelease
		ev.Key = &event.KeyEvent{
			Key: raw.Key,
			Mod: event.KeyModifier(raw.Modifiers),
		}

	case platform.InputMouse:
		switch raw.MouseAction {
		case platform.MousePress:
			ev.TypeValue = event.EventMousePress
		case platform.MouseRelease:
			ev.TypeValue = event.EventMouseRelease
		case platform.MouseMotion:
			ev.TypeValue = event.EventMouseMove
		case platform.MouseWheelUp, platform.MouseWheelDown:
			ev.TypeValue = event.EventMouseWheel
		default:
			return nil
		}
		ev.Mouse = &event.MouseEvent{
			X:     raw.MouseX,
			Y:     raw.MouseY,
			Type:  mouseActionToEventType(raw.MouseAction),
			Click: mouseButtonToClickType(raw.MouseButton),
		}

	case platform.InputResize:
		ev.TypeValue = event.EventResize
		ev.Resize = &event.ResizeEvent{
			Width:  raw.Width,
			Height: raw.Height,
		}
		// 处理窗口大小变化（仅在非固定大小模式下）
		if !e.fixedSize {
			e.renderer.Resize(raw.Width, raw.Height)
			e.RequestRepaint()
		}

	default:
		return nil
	}

	return ev
}

// mouseActionToEventType 转换鼠标动作
func mouseActionToEventType(action platform.MouseAction) event.MouseEventType {
	switch action {
	case platform.MousePress:
		return event.MousePress
	case platform.MouseRelease:
		return event.MouseRelease
	case platform.MouseMotion:
		return event.MouseMove
	case platform.MouseWheelUp, platform.MouseWheelDown:
		return event.MouseScroll
	default:
		return event.MousePress
	}
}

// mouseButtonToClickType 转换鼠标按钮
func mouseButtonToClickType(button platform.MouseButton) event.MouseClickType {
	switch button {
	case platform.MouseLeft:
		return event.MouseLeft
	case platform.MouseMiddle:
		return event.MouseMiddle
	case platform.MouseRight:
		return event.MouseRight
	default:
		return event.MouseLeft
	}
}

// recordActivity 记录活动时间
func (e *Engine) recordActivity() {
	e.lastActivityTime = time.Now()
}

// enterIdle 进入空闲模式
func (e *Engine) enterIdle() {
	if e.idle.CompareAndSwap(false, true) {
		// 进入空闲模式
	}
}

// exitIdle 退出空闲模式
func (e *Engine) exitIdle() {
	if e.idle.CompareAndSwap(true, false) {
		e.RequestRepaint()
	}
}

// PostEvent 投递事件
//
// 线程安全，可以从任何 goroutine 调用
func (e *Engine) PostEvent(ev *event.EventStruct) {
	select {
	case e.eventQueue <- ev:
	default:
		// 队列满，丢弃事件
	}
}

// PostRawInput 投递原始输入（供外部使用）
func (e *Engine) PostRawInput(raw platform.RawInput) {
	select {
	case e.inputEvents <- raw:
	default:
		// 队列满，丢弃
	}
}

// RequestRepaint 请求重绘
//
// 线程安全
func (e *Engine) RequestRepaint() {
	e.repaintNeeded.Store(true)
	e.recordActivity()
	e.exitIdle()
}

// Stop 停止引擎
func (e *Engine) Stop() {
	select {
	case <-e.quit:
		// 已经关闭
	default:
		close(e.quit)
	}
	e.cancel()
}

// Resize 调整大小
func (e *Engine) Resize(width, height int) {
	e.renderer.Resize(width, height)

	// 发送 resize 事件
	ev := &event.EventStruct{
		TypeValue: event.EventResize,
		Resize: &event.ResizeEvent{
			Width:  width,
			Height: height,
		},
		TimestampValue: time.Now(),
	}
	e.PostEvent(ev)
}

// IsRunning 检查引擎是否在运行
func (e *Engine) IsRunning() bool {
	return e.running.Load()
}

// IsIdle 检查引擎是否处于空闲状态
func (e *Engine) IsIdle() bool {
	return e.idle.Load()
}

// SetRoot 设置根组件
func (e *Engine) SetRoot(root Renderable) {
	e.rootMu.Lock()
	e.root = root
	e.rootMu.Unlock()
	e.RequestRepaint()
}

// GetRoot 获取根组件
func (e *Engine) GetRoot() Renderable {
	e.rootMu.RLock()
	defer e.rootMu.RUnlock()
	return e.root
}

// GetRenderer 获取渲染器
func (e *Engine) GetRenderer() *paint.Renderer {
	return e.renderer
}

// ForceFullRender 强制下一帧进行全量渲染
func (e *Engine) ForceFullRender() {
	e.renderer.ForceFullRender()
	e.RequestRepaint()
}
