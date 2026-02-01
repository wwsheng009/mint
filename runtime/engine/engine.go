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
	"os"
	"os/signal"
	"strings"
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

	// 键盘事件回调 - 用于应用层处理特殊按键（如 ESC 退出）
	keyHandler func(ev *event.EventStruct)

	// 固定大小模式 - 禁用自动调整大小
	fixedSize bool

	// 文本选择模式状态
	selectionState SelectionState
}

// SelectionState 文本选择模式状态
type SelectionState struct {
	enabled       bool          // 是否启用选择模式
	startX, startY int          // 选择起始位置
	endX, endY     int          // 选择结束位置
	isSelecting   bool          // 是否正在拖拽选择中（鼠标左键按住并移动）
	isLeftButtonDown bool       // 左键是否按下
	hasSelection  bool          // 是否有有效的选择区域（释放后保持）
	selectStartTime time.Time   // 选择开始时间（用于检测长按）
	clickCount    int           // 连续点击计数（用于双击、三击检测）
	lastClickTime time.Time     // 上次点击时间
	lastClickX    int           // 上次点击位置 X
	lastClickY    int           // 上次点击位置 Y
}

// IsInSelectionMode 检查是否处于文本选择模式
func (e *Engine) IsInSelectionMode() bool {
	return e.selectionState.enabled
}

// GetSelection 获取当前选择区域
func (e *Engine) GetSelection() (startX, startY, endX, endY int, hasSelection bool) {
	if !e.selectionState.enabled || !e.selectionState.isSelecting {
		return 0, 0, 0, 0, false
	}
	return e.selectionState.startX, e.selectionState.startY,
		e.selectionState.endX, e.selectionState.endY, true
}

// ClearSelection 清除选择
func (e *Engine) ClearSelection() {
	e.selectionState.enabled = false
	e.selectionState.isSelecting = false
	e.selectionState.hasSelection = false
	e.RequestRepaint()
}

// GetSelectedText 获取选中的文本内容
func (e *Engine) GetSelectedText() string {
	if !e.selectionState.enabled {
		return ""
	}

	// 获取规范化后的选择区域
	startX, startY, endX, endY := e.getNormalizedSelection()

	// 获取渲染缓冲区
	buf := e.renderer.GetBackBuffer()

	var result strings.Builder

	// 按行提取文本
	for y := startY; y <= endY; y++ {
		if y < 0 || y >= buf.Height {
			continue
		}

		lineStart := 0
		lineEnd := buf.Width - 1

		if y == startY {
			lineStart = startX
		}
		if y == endY {
			lineEnd = endX
		}

		for x := lineStart; x <= lineEnd; x++ {
			if x < 0 || x >= buf.Width {
				continue
			}

			cell := buf.Cells[y][x]
			if cell.IsContinuation {
				continue // 跳过宽字符的延续单元格
			}
			if cell.Cluster != "" && cell.Cluster != " " && cell.Cluster != "\x00" {
				result.WriteString(cell.Cluster)
			} else {
				result.WriteRune(' ')
			}
		}

		// 非最后一行添加换行符
		if y < endY {
			result.WriteRune('\n')
		}
	}

	return result.String()
}

// CopySelection 复制选中的文本到剪贴板
func (e *Engine) CopySelection() error {
	text := e.GetSelectedText()
	if text == "" {
		return nil
	}

	return e.copyToClipboard(text)
}

// copyToClipboard 将文本复制到系统剪贴板
func (e *Engine) copyToClipboard(text string) error {
	// 尝试使用系统剪贴板命令
	return copyToClipboardPlatform(text)
}

// getNormalizedSelection 获取规范化后的选择区域（确保 start <= end）
func (e *Engine) getNormalizedSelection() (startX, startY, endX, endY int) {
	startX, startY = e.selectionState.startX, e.selectionState.startY
	endX, endY = e.selectionState.endX, e.selectionState.endY

	// 规范化坐标
	if startY > endY || (startY == endY && startX > endX) {
		startX, endX = endX, startX
		startY, endY = endY, startY
	}

	return startX, startY, endX, endY
}

// selectWordAt 选中指定位置的单词
func (e *Engine) selectWordAt(x, y int) {
	buf := e.renderer.GetBackBuffer()
	if y < 0 || y >= buf.Height || x < 0 || x >= buf.Width {
		return
	}

	// 找到单词的开始位置
	startX := x
	for startX > 0 {
		cell := buf.Cells[y][startX-1]
		if cell.IsContinuation || !isWordChar(cell.Cluster) {
			break
		}
		startX--
	}

	// 找到单词的结束位置
	endX := x
	for endX < buf.Width-1 {
		cell := buf.Cells[y][endX+1]
		if cell.IsContinuation || !isWordChar(cell.Cluster) {
			break
		}
		endX++
	}

	// 设置选择区域
	e.selectionState.startX = startX
	e.selectionState.startY = y
	e.selectionState.endX = endX
	e.selectionState.endY = y
	e.selectionState.enabled = true
	e.selectionState.hasSelection = true
	e.selectionState.isSelecting = false
	e.selectionState.isLeftButtonDown = false

	e.RequestRepaint()
}

// selectLineAt 选中指定行
func (e *Engine) selectLineAt(y int) {
	buf := e.renderer.GetBackBuffer()
	if y < 0 || y >= buf.Height {
		return
	}

	// 找到行的实际内容边界
	startX := 0
	endX := buf.Width - 1

	// 设置选择区域
	e.selectionState.startX = startX
	e.selectionState.startY = y
	e.selectionState.endX = endX
	e.selectionState.endY = y
	e.selectionState.enabled = true
	e.selectionState.hasSelection = true
	e.selectionState.isSelecting = false
	e.selectionState.isLeftButtonDown = false

	e.RequestRepaint()
}

// isWordChar 判断字符是否为单词字符
func isWordChar(s string) bool {
	if s == "" || len(s) != 1 {
		return false
	}
	r := rune(s[0])
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_'
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
// 同时自动初始化/更新焦点管理器
func (e *Engine) SetLayoutBoxes(boxes []runtime.LayoutBox) {
	e.layoutMu.Lock()
	defer e.layoutMu.Unlock()
	e.layoutBoxes = boxes

	// 自动初始化焦点管理器（如果尚未设置）
	e.initOrUpdateFocusManager(boxes)
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

// SetKeyHandler 设置键盘事件处理回调
//
// 应用可以使用此回调来处理特殊按键（如 ESC 退出）
// 返回 true 表示事件已被处理，false 表示继续传播
func (e *Engine) SetKeyHandler(handler func(ev *event.EventStruct)) {
	e.keyHandler = handler
}

// initOrUpdateFocusManager 自动初始化或更新焦点管理器
// 如果用户没有手动设置焦点管理器，则自动创建一个
func (e *Engine) initOrUpdateFocusManager(boxes []runtime.LayoutBox) {
	// 如果用户已经手动设置了焦点管理器，只更新可聚焦组件
	if e.focusManager != nil {
		e.focusManager.RefreshFocusables()
		return
	}

	// 创建一个简单的根节点用于焦点管理
	rootNode := e.buildRootNodeFromBoxes(boxes)
	if rootNode == nil {
		return
	}

	// 创建焦点管理器
	e.focusManager = focus.NewManager(rootNode)
	e.focusManager.RefreshFocusables()

	// 默认聚焦第一个组件
	e.focusManager.FocusFirst()
}

// buildRootNodeFromBoxes 从布局框构建一个简单的根节点
// 用于焦点管理的树遍历
func (e *Engine) buildRootNodeFromBoxes(boxes []runtime.LayoutBox) *runtime.LayoutNode {
	if len(boxes) == 0 {
		return nil
	}

	// 创建一个虚拟根节点，包含所有布局节点作为子节点
	root := &runtime.LayoutNode{
		ID:   "__root__",
		Type: runtime.NodeTypeFlex,
	}

	for _, box := range boxes {
		if box.Node != nil {
			root.AddChild(box.Node)
		}
	}

	return root
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

	// 🔥 关键修复：使用 defer 确保任何退出路径都会执行 cleanup
	// 这包括 panic、os.Exit（在 defer 之后调用）、正常返回等
	cleanup := func() {
		if e.inputReader != nil {
			e.inputReader.Stop()
			e.inputReader = nil
		}
		
		platform.RestoreTerminal()
	}
	defer cleanup() // 🔥 立即 defer，确保任何退出都会执行

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 启动信号处理 goroutine
	go func() {
		 <-sigChan
		// fmt.Printf("[Engine] Received signal: %v, cleaning up...", sig)
		// 通过关闭 quit channel 让 Run() 正常返回，让 defer cleanup 执行
		select {
		case <-e.quit:
			// 已经关闭
		default:
			close(e.quit)
		}
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

	// 清理 ticker
	ticker.Stop()
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
// 4. 应用文本选择高亮
// 5. 渲染输出
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

	// 4. 应用文本选择高亮
	e.applySelectionHighlight(buf)

	// 5. 渲染输出
	output := e.renderer.Render()
	if output != "" && e.outputFunc != nil {
		e.outputFunc(output)
	}
}

// applySelectionHighlight 应用文本选择高亮
func (e *Engine) applySelectionHighlight(buf *paint.Buffer) {
	// 只要选择模式启用就渲染（包括正在拖拽和已释放）
	if !e.selectionState.enabled {
		return
	}

	// 计算选择区域的边界（确保 start <= end）
	startX, startY := e.selectionState.startX, e.selectionState.startY
	endX, endY := e.selectionState.endX, e.selectionState.endY

	// 规范化坐标（确保 start 在左上角，end 在右下角）
	if startY > endY || (startY == endY && startX > endX) {
		startX, endX = endX, startX
		startY, endY = endY, startY
	}

	// 边界检查
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX >= buf.Width {
		endX = buf.Width - 1
	}
	if endY >= buf.Height {
		endY = buf.Height - 1
	}

	// 应用选择高亮 - 直接修改 cell 的 Style 添加反色
	for y := startY; y <= endY; y++ {
		if y < 0 || y >= buf.Height {
			continue
		}

		lineStart := 0
		lineEnd := buf.Width - 1

		if y == startY {
			lineStart = startX
		}
		if y == endY {
			lineEnd = endX
		}

		for x := lineStart; x <= lineEnd; x++ {
			if x < 0 || x >= buf.Width {
				continue
			}

			// 直接修改 cell 的 Style，添加反色效果
			cell := buf.Cells[y][x]
			cell.Style = cell.Style.Reverse(true)
			buf.Cells[y][x] = cell
		}
	}
}

// handleEvent 处理事件
//
// 使用 runtime/event 的三阶段传播系统分发事件
func (e *Engine) handleEvent(ev *event.EventStruct) {
	// 先调用应用层的键盘处理回调
	if e.keyHandler != nil && (ev.Type() == event.EventKeyPress || ev.Type() == event.EventKeyRelease) {
		e.keyHandler(ev)
	}

	e.layoutMu.RLock()
	boxes := e.layoutBoxes
	e.layoutMu.RUnlock()

	// 处理鼠标事件（支持文本选择模式）
	if ev.Mouse != nil {
		e.handleMouseEvent(ev)
		return
	}

	// 键盘事件处理
	if ev.Type() == event.EventKeyPress {
		// 检查是否是 Ctrl+C（复制）
		if e.selectionState.enabled && ev.Key != nil {
			isCtrlC := (ev.Key.Key == 'c' || ev.Key.Key == 'C') && (ev.Key.Mod&event.ModCtrl) != 0
			if isCtrlC {
				// 复制选中的文本
				if err := e.CopySelection(); err == nil {
					// 复制成功，不清除选择
					e.RequestRepaint()
					return
				}
			}
		}

		// 其他键盘按键：如果有选择区域，清除选择
		if e.selectionState.hasSelection {
			e.ClearSelection()
		}
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

	// 键盘事件也触发重绘（提供视觉反馈）
	if ev.Type() == event.EventKeyPress {
		e.RequestRepaint()
	}
}

// handleMouseEvent 处理鼠标事件（包括文本选择模式）
func (e *Engine) handleMouseEvent(ev *event.EventStruct) {
	mouseEv := ev.Mouse

	switch ev.Type() {
	case event.EventMousePress:
		// 左键按下：开始选择模式计时
		if mouseEv.Click == event.MouseLeft {
			// 检测双击/三击（500ms 内，相同位置）
			timeSinceLastClick := time.Since(e.selectionState.lastClickTime)
			dx := mouseEv.X - e.selectionState.lastClickX
			dy := mouseEv.Y - e.selectionState.lastClickY
			isSamePosition := dx*dx+dy*dy <= 4 // 距离小于2个单元格

			if timeSinceLastClick < 500*time.Millisecond && isSamePosition {
				e.selectionState.clickCount++
			} else {
				e.selectionState.clickCount = 1
			}

			e.selectionState.lastClickTime = time.Now()
			e.selectionState.lastClickX = mouseEv.X
			e.selectionState.lastClickY = mouseEv.Y

			// 根据点击次数执行不同操作
			switch e.selectionState.clickCount {
			case 2:
				// 双击：选中单词
				e.selectWordAt(mouseEv.X, mouseEv.Y)
			case 3:
				// 三击：选中整行
				e.selectLineAt(mouseEv.Y)
			default:
				// 检查是否按住 Shift 键（扩展选择）
				isShiftPressed := (mouseEv.Mod & event.ModShift) != 0

				if isShiftPressed && e.selectionState.hasSelection {
					// Shift+点击：扩展选择范围
					e.selectionState.endX = mouseEv.X
					e.selectionState.endY = mouseEv.Y
					e.selectionState.isSelecting = false
					e.selectionState.enabled = true
				} else {
					// 普通单击：开始新的选择
					if e.selectionState.hasSelection {
						e.selectionState.hasSelection = false
						e.selectionState.enabled = false
					}

					e.selectionState.isLeftButtonDown = true
					e.selectionState.selectStartTime = time.Now()
					e.selectionState.startX = mouseEv.X
					e.selectionState.startY = mouseEv.Y
					e.selectionState.endX = mouseEv.X
					e.selectionState.endY = mouseEv.Y
					e.selectionState.isSelecting = false
					e.selectionState.enabled = false
				}
			}
		}

		// 调用鼠标移动回调
		if e.mouseMoveCallback != nil {
			e.mouseMoveCallback(mouseEv.X, mouseEv.Y)
		}

		// 只有在非选择模式下才分发点击事件给组件
		if !e.selectionState.enabled {
			e.layoutMu.RLock()
			boxes := e.layoutBoxes
			e.layoutMu.RUnlock()
			result := event.DispatchEvent(ev, boxes)
			if result.Updated {
				e.RequestRepaint()
			}
		}
		e.RequestRepaint()

	case event.EventMouseRelease:
		// 左键释放：结束拖拽，但保持选择区域
		if mouseEv.Click == event.MouseLeft {
			e.selectionState.isLeftButtonDown = false

			// 检查是否移动过（用于区分点击和选择）
			dx := mouseEv.X - e.selectionState.startX
			dy := mouseEv.Y - e.selectionState.startY
			hasMoved := dx != 0 || dy != 0
			wasSelecting := e.selectionState.enabled && e.selectionState.isSelecting

			if !wasSelecting || !hasMoved {
				// 没有移动过，视为普通点击，清除选择区域
				e.selectionState.hasSelection = false
				e.selectionState.enabled = false
				e.selectionState.isSelecting = false
			} else {
				// 有有效的选择区域，保持它
				e.selectionState.hasSelection = true
				e.selectionState.isSelecting = false
				// 注意：enabled 保持 true 以显示选择高亮
			}
		}

		// 调用鼠标移动回调
		if e.mouseMoveCallback != nil {
			e.mouseMoveCallback(mouseEv.X, mouseEv.Y)
		}
		e.RequestRepaint()

	case event.EventMouseMove:
		// 鼠标移动：更新位置
		if e.mouseMoveCallback != nil {
			e.mouseMoveCallback(mouseEv.X, mouseEv.Y)
		}

		// 检查是否应该进入选择模式（按住左键并移动）
		if e.selectionState.isLeftButtonDown && !e.selectionState.enabled {
			dx := mouseEv.X - e.selectionState.startX
			dy := mouseEv.Y - e.selectionState.startY

			// 只要按住并移动（任意距离），立即进入选择模式
			if dx != 0 || dy != 0 {
				e.selectionState.enabled = true
				e.selectionState.isSelecting = true
			}
		}

		// 如果处于选择模式，更新选择区域
		if e.selectionState.enabled && e.selectionState.isSelecting {
			e.selectionState.endX = mouseEv.X
			e.selectionState.endY = mouseEv.Y
		}

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
			Key:     raw.Key,
			Special: raw.Special,
			Type:    event.KeyPress,
			Mod:     event.KeyModifier(raw.Modifiers),
		}

	case platform.InputKeyRelease:
		ev.TypeValue = event.EventKeyRelease
		ev.Key = &event.KeyEvent{
			Key:     raw.Key,
			Special: raw.Special,
			Type:    event.KeyRelease,
			Mod:     event.KeyModifier(raw.Modifiers),
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
			Mod:   event.KeyModifier(raw.Modifiers),
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
