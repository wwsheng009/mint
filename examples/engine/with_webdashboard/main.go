// Engine Example with DevTools Server (New Unified Protocol)
//
// This example shows how to integrate the DevTools server into an existing TUI application.
// The key points:
//   1. Start DevTools server at application startup
//   2. Use BeginFrame/EndFrame to automatically capture frame data
//   3. Component updates are recorded automatically
//   4. No manual simulation needed - data comes from real application activity
//
// Usage: go run main.go
// Then open http://localhost:8080/ in your browser
package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/protocol"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/engine"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// 全局状态
// =============================================================================

var (
	debugState struct {
		sync.RWMutex
		mouseX    int
		mouseY    int
		lastEvent string
	}

	// DevTools 实例
	dt *devtools.DevTools

	// DevToolsServer 实例 - 使用新的统一协议包
	server *protocol.Server
)

func init() {
	debugState.mouseX = -1
	debugState.mouseY = -1

	// 初始化 DevTools
	dt = devtools.New()

	// 初始化 DevToolsServer (使用新的 protocol 包)
	server = protocol.NewServer(protocol.ServerConfig{
		Port:              8080,
		EnableDashboard:   true,
		EnableTuiCommands: true,
	})
}

// =============================================================================
// 集成说明
// =============================================================================
//
// 在现有应用中集成 DevTools 只需几步:
//
// 1. 在 main() 或 init() 中创建 DevToolsServer:
//    server := protocol.NewServer(protocol.ServerConfig{Port: 8080})
//
// 2. 在应用启动时启动它:
//    server.Start()
//    defer server.Stop()
//
// 3. 在渲染循环中使用 BeginFrame/EndFrame (这会自动记录帧数据):
//    dt.BeginFrame()
//    ... 渲染逻辑 ...
//    dt.EndFrame()
//
// 4. (可选) 更新组件状态:
//    server.UpdateComponent(componentID, &protocol.DashboardComponentData{...})
//
// 5. (可选) 定期更新性能指标:
//    server.UpdateMetrics(&protocol.Metrics{...})
//
// 不需要 runSimulation! 数据来自真实的应用活动。
//
// =============================================================================

func logEvent(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	debugState.Lock()
	debugState.lastEvent = msg
	debugState.Unlock()

	// 记录到 DevTools (自动收集)
	if dt.IsEnabled() {
		dt.RecordEvent("log", "system", "target", map[string]interface{}{
			"message": msg,
		})
	}
}

// =============================================================================
// Button 组件
// =============================================================================

type Button struct {
	id      string
	x, y    int
	width   int
	height  int
	text    string
	focused bool
	onClick func()
}

func NewButton(id, text string, x, y int) *Button {
	return &Button{
		id:      id,
		x:       x,
		y:       y,
		width:   len(text) + 4,
		height:  3,
		text:    text,
		onClick: func() { logEvent("Button %s clicked", id) },
	}
}

func (b *Button) ID() string { return b.id }

func (b *Button) Paint(buf *paint.Buffer) {
	startX, startY := b.x, b.y
	endX, endY := b.x+b.width-1, b.y+b.height-1

	borderColor := style.White
	if b.focused {
		borderColor = style.Yellow
	}

	borderStyle := style.Style{}.Foreground(borderColor)
	textStyle := style.Style{}.Foreground(style.Cyan)

	// Draw border
	for x := startX; x <= endX; x++ {
		buf.SetCell(x, startY, '─', borderStyle)
		buf.SetCell(x, endY, '─', borderStyle)
	}
	for y := startY; y <= endY; y++ {
		buf.SetCell(startX, y, '│', borderStyle)
		buf.SetCell(endX, y, '│', borderStyle)
	}
	buf.SetCell(startX, startY, '┌', borderStyle)
	buf.SetCell(endX, startY, '┐', borderStyle)
	buf.SetCell(startX, endY, '└', borderStyle)
	buf.SetCell(endX, endY, '┘', borderStyle)

	// Draw text (centered)
	textX := startX + (b.width-len(b.text))/2
	textY := startY + b.height/2
	buf.SetString(textX, textY, b.text, textStyle)
}

func (b *Button) HandleMouse(ev *event.MouseEvent, localX, localY int) bool {
	if dt.IsEnabled() {
		clickType := "none"
		switch ev.Click {
		case event.MouseLeft:
			clickType = "left"
		case event.MouseMiddle:
			clickType = "middle"
		case event.MouseRight:
			clickType = "right"
		}
		dt.RecordEvent("mouse", b.id, "target", map[string]interface{}{
			"type":  string(ev.Type),
			"click": clickType,
		})
	}

	if ev.Type == event.MousePress && ev.Click == event.MouseLeft {
		logEvent("Button %s CLICKED!", b.id)
		if b.onClick != nil {
			b.onClick()
		}
		return true
	}
	return false
}

func (b *Button) SetFocus(focus bool) {
	b.focused = focus
	if dt.IsEnabled() {
		dt.RecordEvent("focus", b.id, "target", map[string]interface{}{
			"focused": focus,
		})
	}
}

func (b *Button) SetOnClick(fn func()) {
	b.onClick = fn
}

func (b *Button) IsFocusable() bool { return true }
func (b *Button) Bounds() (int, int, int, int) { return b.x, b.y, b.width, b.height }
func (b *Button) Contains(x, y int) bool {
	return x >= b.x && x < b.x+b.width && y >= b.y && y < b.y+b.height
}

// =============================================================================
// Root 组件
// =============================================================================

type Root struct {
	id         string
	width      int
	height     int
	buttons    []*Button
	frameCount int
	lastMetricsUpdate time.Time
}

func NewRoot(id string, width, height int) *Root {
	return &Root{
		id:      id,
		width:   width,
		height:  height,
		buttons: make([]*Button, 0),
	}
}

func (r *Root) ID() string { return r.id }

func (r *Root) AddButton(btn *Button) {
	r.buttons = append(r.buttons, btn)
}

func (r *Root) SetFocusedButton(btnID string) {
	for _, btn := range r.buttons {
		if btn.ID() == btnID {
			btn.SetFocus(true)
		} else {
			btn.SetFocus(false)
		}
	}
}

// Paint 渲染方法
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
	buf.SetString(2, 1, "Engine with DevTools Server - Real-time Debugging", titleStyle)

	// 绘制说明
	infoStyle := style.Style{}.Foreground(style.White)
	buf.SetString(2, 3, "Open http://localhost:8080/ in your browser", infoStyle)
	buf.SetString(2, 4, "Click buttons - events are captured in real-time", infoStyle)

	// 绘制所有按钮
	for _, btn := range r.buttons {
		btn.Paint(buf)
	}

	// 绘制状态栏
	r.drawStatusBar(buf)
}

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

	// 显示帧计数
	frameText := fmt.Sprintf("Frame: %d", r.frameCount)
	buf.SetString(r.width-len(frameText)-2, r.height-3, frameText, style.Style{}.Foreground(style.BrightBlack))

	// 显示 DevTools 状态
	devtoolsStatus := "DevTools: ON"
	buf.SetString(2, r.height-4, devtoolsStatus, style.Style{}.Foreground(style.Yellow))

	// 显示最后事件
	debugState.RLock()
	lastEvent := debugState.lastEvent
	debugState.RUnlock()

	if lastEvent != "" {
		if len(lastEvent) > r.width-4 {
			lastEvent = lastEvent[:r.width-7] + "..."
		}
		buf.SetString(2, r.height-1, lastEvent, style.Style{}.Foreground(style.Yellow))
	}
}

func (r *Root) Update(dt time.Duration) {
	r.frameCount++

	// 添加帧数据到 DevToolsServer
	server.AddFrame(&protocol.FrameData{
		FrameID:       devtools.FrameID(r.frameCount),
		Timestamp:     time.Now(),
		Duration:      dt,
		EventCount:    0, // Could track actual events
		MutationCount: 0,
		LayoutCount:   1,
		RepaintCount:  1,
	})

	// 每 10 帧更新一次性能指标到 DevToolsServer
	if time.Since(r.lastMetricsUpdate) > 500*time.Millisecond {
		r.updateMetrics()
		r.lastMetricsUpdate = time.Now()
	}

	// 更新组件状态到 DevToolsServer
	r.updateComponents()
}

// updateMetrics 更新性能指标
func (r *Root) updateMetrics() {
	server.UpdateMetrics(&protocol.Metrics{
		FPS:            60.0,                 // 目标 FPS
		FrameTime:      16 * time.Millisecond, // ~60fps
		LayoutTime:     5 * time.Millisecond,
		PaintTime:      8 * time.Millisecond,
		MemoryUsage:    50 * 1024 * 1024,    // 50MB
		ComponentCount: len(r.buttons) + 2,   // buttons + root + cursor
		FrameCount:     r.frameCount,
	})
}

// updateComponents 更新组件状态到 DevToolsServer
func (r *Root) updateComponents() {
	// 更新每个按钮的状态
	for _, btn := range r.buttons {
		server.UpdateComponent(btn.id, &protocol.DashboardComponentData{
			ID:   btn.id,
			Type: "Button",
			Properties: map[string]interface{}{
				"text":    btn.text,
				"focused": btn.focused,
			},
			Styles: map[string]interface{}{
				"x":      btn.x,
				"y":      btn.y,
				"width":  btn.width,
				"height": btn.height,
			},
		})
	}
}

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

func (r *Root) GetButtonByID(id string) *Button {
	for _, btn := range r.buttons {
		if btn.ID() == id {
			return btn
		}
	}
	return nil
}

// =============================================================================
// 主程序 - 集成 DevToolsServer 的标准方式
// =============================================================================

func EngineExample() error {
	const (
		width  = 70
		height = 15
	)

	// 步骤 1: 启动 DevToolsServer (使用新的 protocol 包)
	fmt.Println("Starting DevToolsServer on http://localhost:8080/")
	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start DevToolsServer: %w", err)
	}
	defer server.Stop()

	// 步骤 2: 启用 DevTools
	dt.Enable()

	// 步骤 3: 创建应用
	root := NewRoot("root", width, height)

	// 创建按钮
	btn1 := NewButton("btn1", "Button 1", 5, 7)
	btn2 := NewButton("btn2", "Button 2", 25, 7)
	btn3 := NewButton("btn3", "Exit", 45, 7)

	eng := engine.New(width, height, root)
	eng.SetFixedSize(true)

	// 设置按钮回调
	btn1.SetOnClick(func() {
		logEvent(">>> Button 1 CLICKED! <<<")
	})

	btn2.SetOnClick(func() {
		logEvent(">>> Button 2 CLICKED! <<<")
	})

	btn3.SetOnClick(func() {
		logEvent(">>> Exit clicked, quitting... <<<")
		eng.Stop()
	})

	root.AddButton(btn1)
	root.AddButton(btn2)
	root.AddButton(btn3)

	eng.SetOutputFunc(func(output string) {
		if output != "" {
			fmt.Print(output)
		}
	})

	eng.SetKeyHandler(func(ev *event.EventStruct) {
		if ev.Key != nil && ev.Key.Special == platform.KeyEscape {
			logEvent("ESC pressed, stopping engine...")
			eng.Stop()
		}
	})

	boxes := root.BuildLayoutBoxes()
	eng.SetLayoutBoxes(boxes)

	// 终端初始化
	fmt.Print("\x1b[?25l")
	fmt.Print("\x1b[2J")
	fmt.Print("\x1b[H")

	fmt.Println("=== Engine Example with DevTools Server ===")
	fmt.Println()
	fmt.Println("Web Dashboard: http://localhost:8080/")
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  - Click buttons to generate events")
	fmt.Println("  - ESC or Ctrl+C: Exit")
	fmt.Println()
	fmt.Println("All events are captured in real-time on the web dashboard!")

	logEvent("Engine started...")

	// 步骤 4: 运行应用 (引擎的渲染循环会自动处理帧调度)
	if err := eng.Run(); err != nil {
		return err
	}

	fmt.Println("\n[Engine] Exited cleanly.")

	// 步骤 5: 清理
	dt.Disable()

	return nil
}

func main() {
	defer func() {
		platform.RestoreTerminal()
		fmt.Print("\x1b[?25h")
		fmt.Print("\x1b[0m")
		fmt.Print("\x1b[H")
		fmt.Println()

		if dt.IsEnabled() {
			dt.Disable()
		}

		if server != nil && server.IsRunning() {
			server.Stop()
		}
	}()

	if err := EngineExample(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
}
