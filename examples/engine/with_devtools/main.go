package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/engine"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// 全局状态 - 用于调试和显示
// =============================================================================

var (
	debugState struct {
		sync.RWMutex
		mouseX     int
		mouseY     int
		mouseClick string
		lastEvent  string
		eventLog   []string
	}

	// DevTools 实例
	dt *devtools.DevTools
)

func init() {
	debugState.eventLog = make([]string, 0, 10)
	debugState.mouseX = -1
	debugState.mouseY = -1

	// 初始化 DevTools
	dt = devtools.New()
}

func logEvent(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	debugState.Lock()
	debugState.lastEvent = msg
	debugState.eventLog = append(debugState.eventLog, msg)
	if len(debugState.eventLog) > 5 {
		debugState.eventLog = debugState.eventLog[1:]
	}
	debugState.Unlock()
	// fmt.Printf("[DEBUG] %s\n", msg)

	// 记录到 DevTools
	if dt.IsEnabled() {
		dt.RecordEvent("log", "system", "target", map[string]interface{}{
			"message": msg,
		})
	}
}

func updateMousePos(x, y int) {
	debugState.Lock()
	debugState.mouseX = x
	debugState.mouseY = y
	debugState.Unlock()

	// 记录鼠标移动到 DevTools
	if dt.IsEnabled() {
		dt.RecordEvent("mousemove", "root", "target", map[string]interface{}{
			"x": x,
			"y": y,
		})
	}
}

func getMousePos() (int, int) {
	debugState.RLock()
	defer debugState.RUnlock()
	return debugState.mouseX, debugState.mouseY
}

// =============================================================================
// Button - 可点击的按钮组件
// =============================================================================

// Button 按钮组件
type Button struct {
	id       string
	x, y     int // 位置
	width    int
	height   int
	text     string
	focused  bool
	onClick  func()
	onFocus  func(bool)
}

// NewButton 创建按钮
func NewButton(id, text string, x, y int) *Button {
	return &Button{
		id:      id,
		x:       x,
		y:       y,
		width:   len(text) + 4,
		height:  3,
		text:    text,
		focused: false,
		onClick: func() { logEvent("Button %s clicked (default)", id) },
		onFocus: func(focused bool) {},
	}
}

// engine.Renderable 接口
func (b *Button) ID() string {
	return b.id
}

func (b *Button) Paint(buf *paint.Buffer) {
	// 计算按钮在缓冲区的实际位置
	startX := b.x
	startY := b.y
	endX := b.x + b.width - 1
	endY := b.y + b.height - 1

	// 确保不超出缓冲区
	if startX >= buf.Width || startY >= buf.Height {
		return
	}
	if endX >= buf.Width {
		endX = buf.Width - 1
	}
	if endY >= buf.Height {
		endY = buf.Height - 1
	}

	// 选择边框颜色
	borderColor := style.White
	if b.focused {
		borderColor = style.Yellow
	}

	// 检查鼠标是否悬停在按钮上
	mx, my := getMousePos()
	hover := mx >= b.x && mx < b.x+b.width && my >= b.y && my < b.y+b.height
	if hover {
		borderColor = style.Green
	}

	borderStyle := style.Style{}.Foreground(borderColor)
	textStyle := style.Style{}.Foreground(style.Cyan)

	// 绘制边框
	for x := startX; x <= endX; x++ {
		buf.SetCell(x, startY, '─', borderStyle)
	}
	for x := startX; x <= endX; x++ {
		buf.SetCell(x, endY, '─', borderStyle)
	}
	for y := startY; y <= endY; y++ {
		buf.SetCell(startX, y, '│', borderStyle)
	}
	for y := startY; y <= endY; y++ {
		buf.SetCell(endX, y, '│', borderStyle)
	}
	// 角落
	buf.SetCell(startX, startY, '┌', borderStyle)
	buf.SetCell(endX, startY, '┐', borderStyle)
	buf.SetCell(startX, endY, '└', borderStyle)
	buf.SetCell(endX, endY, '┘', borderStyle)

	// 绘制文字（居中）
	textX := startX + (b.width-len(b.text))/2
	textY := startY + b.height/2
	buf.SetString(textX, textY, b.text, textStyle)
}

// event.MouseEventHandler 接口
func (b *Button) HandleMouse(ev *event.MouseEvent, localX, localY int) bool {
	logEvent("Button %s received mouse event: Type=%v Click=%v Pos=(%d,%d)",
		b.id, ev.Type, ev.Button, localX, localY)

	// 记录鼠标事件到 DevTools
	if dt.IsEnabled() {
		clickType := "none"
		switch ev.Button {
		case event.MouseLeft:
			clickType = "left"
		case event.MouseMiddle:
			clickType = "middle"
		case event.MouseRight:
			clickType = "right"
		}
		eventData := map[string]interface{}{
			"button_id": b.id,
			"local_x":   localX,
			"local_y":   localY,
			"type":      string(ev.Type),
			"click":     clickType,
		}
		dt.RecordEvent("mouse", b.id, "target", eventData)
	}

	if ev.Type == event.MousePress && ev.Button == event.MouseLeft {
		logEvent("Button %s CLICKED! Triggering callback", b.id)
		if b.onClick != nil {
			b.onClick()
		}
		return true // 事件已处理
	}
	return false
}

// runtime.FocusableComponent 接口
func (b *Button) SetFocus(focus bool) {
	b.focused = focus
	logEvent("Button %s SetFocus(%v)", b.id, focus)

	// 记录焦点变化到 DevTools
	if dt.IsEnabled() {
		dt.RecordEvent("focus", b.id, "target", map[string]interface{}{
			"focused": focus,
		})
	}

	if b.onFocus != nil {
		b.onFocus(focus)
	}
}

func (b *Button) IsFocusable() bool {
	return true
}

// SetOnFocus 设置焦点变化回调
func (b *Button) SetOnFocus(fn func(bool)) {
	b.onFocus = fn
}

// SetOnClick 设置点击回调
func (b *Button) SetOnClick(fn func()) {
	b.onClick = fn
}

// Bounds 返回按钮边界
func (b *Button) Bounds() (int, int, int, int) {
	return b.x, b.y, b.width, b.height
}

// Contains 检查点是否在按钮内
func (b *Button) Contains(x, y int) bool {
	return x >= b.x && x < b.x+b.width && y >= b.y && y < b.y+b.height
}

// =============================================================================
// BlinkingCursor - 闪烁光标组件
// =============================================================================

// BlinkingCursor 闪烁光标组件
type BlinkingCursor struct {
	id        string
	x, y      int
	visible   bool
	lastBlink time.Time
}

// NewBlinkingCursor 创建闪烁光标
func NewBlinkingCursor(id string, x, y int) *BlinkingCursor {
	return &BlinkingCursor{
		id:        id,
		x:         x,
		y:         y,
		visible:   true,
		lastBlink: time.Now(),
	}
}

func (b *BlinkingCursor) ID() string {
	return b.id
}

func (b *BlinkingCursor) Paint(buf *paint.Buffer) {
	if b.visible && b.x < buf.Width && b.y < buf.Height {
		cursorStyle := style.Style{}.Reverse(true)
		buf.SetCell(b.x, b.y, ' ', cursorStyle)
	}
}

func (b *BlinkingCursor) Update(dt time.Duration) {
	if time.Since(b.lastBlink) > 500*time.Millisecond {
		b.visible = !b.visible
		b.lastBlink = time.Now()
	}
}

// =============================================================================
// Root - 根组件
// =============================================================================

// Root 根组件
type Root struct {
	id         string
	width      int
	height     int
	buttons    []*Button
	cursor     *BlinkingCursor
	lastUpdate time.Time
	frameCount int
}

// NewRoot 创建根组件
func NewRoot(id string, width, height int) *Root {
	return &Root{
		id:         id,
		width:      width,
		height:     height,
		buttons:    make([]*Button, 0),
		cursor:     NewBlinkingCursor("cursor", 25, 8),
		lastUpdate: time.Now(),
	}
}

func (r *Root) ID() string {
	return r.id
}

// AddButton 添加按钮
func (r *Root) AddButton(btn *Button) {
	r.buttons = append(r.buttons, btn)

	// 记录组件添加到 DevTools
	if dt.IsEnabled() {
		dt.RecordEvent("component_add", btn.id, "target", map[string]interface{}{
			"type":   "Button",
			"text":   btn.text,
			"x":      btn.x,
			"y":      btn.y,
			"width":  btn.width,
			"height": btn.height,
		})
	}
}

// SetFocusedButton 设置焦点按钮
func (r *Root) SetFocusedButton(btnID string) {
	for _, btn := range r.buttons {
		if btn.ID() == btnID {
			btn.SetFocus(true)
		} else {
			btn.SetFocus(false)
		}
	}
}

// Paint 绘制
func (r *Root) Paint(buf *paint.Buffer) {
	// 清空缓冲区
	emptyStyle := style.Style{}
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			buf.SetCell(x, y, ' ', emptyStyle)
		}
	}

	// 绘制标题
	titleStyle := style.Style{}.Foreground(style.Cyan).Bold(true)
	buf.SetString(2, 1, "Engine Example - Interactive Demo (with DevTools)", titleStyle)

	// 绘制说明
	infoStyle := style.Style{}.Foreground(style.White)
	buf.SetString(2, 3, "Click buttons to test interaction.", infoStyle)
	buf.SetString(2, 4, "Press ESC or Ctrl+C to exit.", infoStyle)

	// 绘制所有按钮
	for _, btn := range r.buttons {
		btn.Paint(buf)
	}

	// 绘制光标
	r.cursor.Paint(buf)

	// 绘制状态栏
	r.drawStatusBar(buf)
}

// drawStatusBar 绘制状态栏
func (r *Root) drawStatusBar(buf *paint.Buffer) {
	statusStyle := style.Style{}.Foreground(style.Green)

	// 显示焦点状态
	focusedBtn := ""
	for _, btn := range r.buttons {
		if btn.focused {
			focusedBtn = btn.text
			break
		}
	}
	statusText := fmt.Sprintf("Focused: %s", focusedBtn)
	if focusedBtn == "" {
		statusText = "Focused: None"
	}
	buf.SetString(2, r.height-3, statusText, statusStyle)

	// 显示鼠标位置
	mx, my := getMousePos()
	mouseText := "Mouse: (-,-)"
	if mx >= 0 {
		mouseText = fmt.Sprintf("Mouse: (%d,%d)", mx, my)
	}
	buf.SetString(2, r.height-2, mouseText, style.Style{}.Foreground(style.Magenta))

	// 显示帧计数
	frameText := fmt.Sprintf("Frame: %d", r.frameCount)
	buf.SetString(r.width-len(frameText)-2, r.height-3, frameText, style.Style{}.Foreground(style.BrightBlack))

	// 显示 DevTools 状态
	devtoolsStatus := "DevTools: OFF"
	if dt.IsEnabled() {
		devtoolsStatus = "DevTools: ON (recording events)"
	}
	buf.SetString(2, r.height-4, devtoolsStatus, style.Style{}.Foreground(style.Yellow))

	// 显示最后事件
	debugState.RLock()
	lastEvent := debugState.lastEvent
	debugState.RUnlock()

	if lastEvent != "" {
		// 截断过长的事件
		if len(lastEvent) > r.width-4 {
			lastEvent = lastEvent[:r.width-7] + "..."
		}
		buf.SetString(2, r.height-1, lastEvent, style.Style{}.Foreground(style.Yellow))
	}
}

// Update 更新状态
func (r *Root) Update(dt time.Duration) {
	r.cursor.Update(dt)
	r.frameCount++
	r.lastUpdate = time.Now()
}

// BuildLayoutBoxes 构建布局框用于命中测试
func (r *Root) BuildLayoutBoxes() []runtime.LayoutBox {
	boxes := make([]runtime.LayoutBox, len(r.buttons))

	for i, btn := range r.buttons {
		x, y, w, h := btn.Bounds()

		node := &runtime.LayoutNode{
			ID: btn.ID(),
			Component: &runtime.ComponentRef{
				Instance: btn,
			},
			X:              x,
			Y:              y,
			MeasuredWidth:  w,
			MeasuredHeight: h,
		}

		boxes[i] = runtime.LayoutBox{
			X:      x,
			Y:      y,
			W:      w,
			H:      h,
			NodeID: btn.ID(),
			Node:   node,
		}
	}

	return boxes
}

// GetButtonByID 根据 ID 获取按钮
func (r *Root) GetButtonByID(id string) *Button {
	for _, btn := range r.buttons {
		if btn.ID() == id {
			return btn
		}
	}
	return nil
}

// HandleMouseMove 处理鼠标移动（用于显示位置）
func (r *Root) HandleMouseMove(x, y int) {
	updateMousePos(x, y)
}

// =============================================================================
// 主程序
// =============================================================================

// EngineExample 引擎示例
func EngineExample() error {
	const (
		width  = 70
		height = 15
	)

	// 启用 DevTools
	dt.Enable()
	logEvent("DevTools enabled")

	// 创建根组件
	root := NewRoot("root", width, height)

	// 创建按钮
	btn1 := NewButton("btn1", "Button 1", 5, 7)
	btn2 := NewButton("btn2", "Button 2", 25, 7)
	btn3 := NewButton("btn3", "Exit", 45, 7)

	// 创建引擎（需要先创建以便按钮可以调用 Stop）
	eng := engine.New(width, height, root)

	// 启用固定大小模式
	eng.SetFixedSize(true)

	// 设置按钮回调
	btn1.SetOnClick(func() {
		logEvent(">>> Button 1 CLICKED! <<<")
	})

	btn2.SetOnClick(func() {
		logEvent(">>> Button 2 CLICKED! <<<")
	})

	btn3.SetOnClick(func() {
		logEvent(">>> Exit button clicked, quitting... <<<")
		fmt.Println("\n[Exit] Stopping engine...")
		eng.Stop()
	})

	// 设置焦点变化回调
	btn1.SetOnFocus(func(focused bool) {
		if focused {
			logEvent("Button 1 gained focus")
		}
	})

	btn2.SetOnFocus(func(focused bool) {
		if focused {
			logEvent("Button 2 gained focus")
		}
	})

	btn3.SetOnFocus(func(focused bool) {
		if focused {
			logEvent("Exit button gained focus")
		}
	})

	root.AddButton(btn1)
	root.AddButton(btn2)
	root.AddButton(btn3)

	// 设置输出函数
	eng.SetOutputFunc(func(output string) {
		if output != "" {
			fmt.Print(output)
		}
	})

	// 设置鼠标移动回调
	eng.SetMouseMoveCallback(func(x, y int) {
		updateMousePos(x, y)
	})

	// 设置键盘事件处理回调 - 处理 ESC 退出
	eng.SetKeyHandler(func(ev *event.EventStruct) {
		if ev.Key != nil && ev.Key.Special == platform.KeyEscape {
			logEvent("ESC pressed, stopping engine...")
			eng.Stop()
		}
	})

	// 设置布局框
	boxes := root.BuildLayoutBoxes()
	eng.SetLayoutBoxes(boxes)

	// 终端初始化
	fmt.Print("\x1b[?25l") // 隐藏光标
	fmt.Print("\x1b[2J")  // 清屏
	fmt.Print("\x1b[H")   // 光标移到左上角

	// 打印启动信息
	fmt.Println("=== Engine Example with DevTools ===")
	fmt.Println("Components: 3 Buttons")
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  - Mouse move: Shows mouse position")
	fmt.Println("  - Mouse click: Click buttons (green = hover, yellow = focused)")
	fmt.Println("  - ESC or Ctrl+C: Exit")
	fmt.Println()
	fmt.Println("DevTools Features:")
	fmt.Println("  - Event tracking: All events are recorded")
	fmt.Println("  - Check console for debug output")
	fmt.Println("  - Events are logged with timestamps")
	fmt.Println()

	logEvent("Engine started, waiting for input...")

	// 运行引擎
	if err := eng.Run(); err != nil {
		return err
	}

	// fmt.Println("\n[Engine] Exited cleanly.")

	// 禁用 DevTools
	dt.Disable()
	logEvent("DevTools disabled")

	return nil
}

func main() {
	// 设置清理函数
	defer func() {
		// 最先恢复终端控制台模式（必须在所有其他操作之前）
		// 这会恢复 ENABLE_LINE_INPUT 和 ENABLE_ECHO_INPUT，让 fmt.Scanln 等正常工作
		platform.RestoreTerminal()

		// 禁用 DevTools
		if dt.IsEnabled() {
			dt.Disable()
		}
	}()

	if err := EngineExample(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
}
