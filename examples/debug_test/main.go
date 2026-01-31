package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wwsheng009/mint/devtools/standalone"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/engine"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
)

var (
	logger *standalone.Logger
)

func init() {
	var err error
	logger, err = standalone.NewLogger(nil)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	fmt.Printf("[DevTools] Logger started: %s\n", logger.GetPath())
}

func logMsg(msg string) {
	// fmt.Printf("[DEBUG] %s\n", msg)
	logger.LogMessage(msg)
}

// =============================================================================
// Button - 简单按钮组件
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
		focused: false,
	}
}

func (b *Button) ID() string {
	return b.id
}

func (b *Button) Paint(buf *paint.Buffer) {
	borderColor := style.White
	if b.focused {
		borderColor = style.Yellow
	}

	borderStyle := style.Style{}.Foreground(borderColor)
	textStyle := style.Style{}.Foreground(style.Cyan)

	// 绘制边框
	startX := b.x
	startY := b.y
	endX := b.x + b.width - 1
	endY := b.y + b.height - 1

	if startX >= buf.Width || startY >= buf.Height {
		return
	}
	if endX >= buf.Width {
		endX = buf.Width - 1
	}
	if endY >= buf.Height {
		endY = buf.Height - 1
	}

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
	buf.SetCell(startX, startY, '┌', borderStyle)
	buf.SetCell(endX, startY, '┐', borderStyle)
	buf.SetCell(startX, endY, '└', borderStyle)
	buf.SetCell(endX, endY, '┘', borderStyle)

	// 绘制文字
	textX := startX + (b.width-len(b.text))/2
	textY := startY + b.height/2
	buf.SetString(textX, textY, b.text, textStyle)
}

func (b *Button) HandleMouse(ev *event.MouseEvent, localX, localY int) bool {
	logMsg(fmt.Sprintf("Button %s HandleMouse: Type=%v Click=%v localX=%d localY=%d", b.id, ev.Type, ev.Click, localX, localY))
	logger.LogMouseEvent(b.id, localX, localY, string(ev.Type), "left")

	// 严格检查：必须是鼠标按下事件，且是左键
	if ev.Type == event.MousePress && ev.Click == event.MouseLeft {
		logMsg(fmt.Sprintf("Button %s CLICKED! (Type=%v, Click=%v)", b.id, ev.Type, ev.Click))
		if b.onClick != nil {
			b.onClick()
		}
		return true
	}
	return false
}

func (b *Button) SetFocus(focus bool) {
	b.focused = focus
	logMsg(fmt.Sprintf("Button %s SetFocus(%v)", b.id, focus))
	logger.LogFocusEvent(b.id, focus)
}

func (b *Button) IsFocusable() bool {
	return true
}

func (b *Button) SetOnClick(fn func()) {
	b.onClick = fn
}

func (b *Button) Bounds() (int, int, int, int) {
	return b.x, b.y, b.width, b.height
}

// 实现 KeyEventHandler 接口
func (b *Button) HandleKey(ev *event.KeyEvent) bool {
	// 记录详细的按键信息
	charStr := "NUL"
	if ev.Key > 0 {
		charStr = fmt.Sprintf("%c", ev.Key)
	}
	if ev.Key == ' ' {
		charStr = "SPACE"
	}

	// 记录特殊键
	specialStr := "none"
	if ev.Special > 0 {
		specialStr = ev.Special.String()
	}

	logMsg(fmt.Sprintf("Button %s HandleKey: Key=%d(%s) Special=%s Type=%s Mod=%d",
		b.id, ev.Key, charStr, specialStr, ev.Type, ev.Mod))

	logger.LogCustom("key", b.id, map[string]interface{}{
		"key":       fmt.Sprintf("%d", ev.Key),
		"char":      charStr,
		"special":   specialStr,
		"type":      string(ev.Type),
		"modifiers": fmt.Sprintf("%d", ev.Mod),
	})
	return false // 不处理键盘事件，只记录
}

// =============================================================================
// Root - 根组件
// =============================================================================

type Root struct {
	id      string
	width   int
	height  int
	buttons []*Button
}

func NewRoot(id string, width, height int) *Root {
	return &Root{
		id:      id,
		width:   width,
		height:  height,
		buttons: make([]*Button, 0),
	}
}

func (r *Root) ID() string {
	return r.id
}

func (r *Root) AddButton(btn *Button) {
	r.buttons = append(r.buttons, btn)
	logger.LogComponentAdd(btn.id, "Button", map[string]interface{}{
		"text": btn.text,
		"x":    btn.x,
		"y":    btn.y,
	})
}

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
	buf.SetString(2, 1, "Debug Test - Press any key!", titleStyle)

	// 绘制按钮
	for _, btn := range r.buttons {
		btn.Paint(buf)
	}

	// 绘制提示
	infoStyle := style.Style{}.Foreground(style.White)
	buf.SetString(2, r.height-2, "Press ESC to exit, or click buttons", infoStyle)
}

func (r *Root) Update(dt time.Duration) {
	// 记录帧更新
	logger.BeginFrame()
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

// =============================================================================
// 主程序
// =============================================================================

func DebugExample() error {
	const (
		width  = 60
		height = 12
	)

	logMsg("=== DebugExample starting ===")

	// 创建根组件
	root := NewRoot("root", width, height)

	// 创建按钮
	btn1 := NewButton("btn1", "Test", 10, 5)
	btn2 := NewButton("btn2", "Exit", 30, 5)

	root.AddButton(btn1)
	root.AddButton(btn2)

	logMsg("Created 2 buttons")

	// 创建引擎
	eng := engine.New(width, height, root)
	eng.SetFixedSize(true)

	// 设置 Test 按钮回调
	btn1.SetOnClick(func() {
		logMsg(">>> Test button clicked! <<<")
		fmt.Println("[Test] Button clicked!")
	})

	// 设置 Exit 按钮回调
	btn2.SetOnClick(func() {
		logMsg(">>> Exit button clicked, stopping... <<<")
		eng.Stop()
	})

	logMsg("Engine created, callbacks set")

	// 设置输出函数
	eng.SetOutputFunc(func(output string) {
		if output != "" {
			fmt.Print(output)
		}
	})

	// 设置布局框
	boxes := root.BuildLayoutBoxes()
	eng.SetLayoutBoxes(boxes)

	logMsg(fmt.Sprintf("Layout boxes set: %d boxes", len(boxes)))

	// 终端初始化
	fmt.Print("\x1b[?25l") // 隐藏光标
	fmt.Print("\x1b[2J")  // 清屏
	fmt.Print("\x1b[H")   // 光标移到左上角

	fmt.Println("=== Debug Test ===")
	fmt.Println("Press keys or click buttons with mouse, ESC to exit")
	fmt.Printf("Log: %s\n", logger.GetPath())
	fmt.Println()

	logMsg("About to call eng.Run()...")

	// 运行引擎
	if err := eng.Run(); err != nil {
		logMsg(fmt.Sprintf("Engine.Run() returned error: %v", err))
		return err
	}

	logMsg("Engine.Run() completed successfully")

	// fmt.Println("\n[Engine] Exited cleanly.")

	if logger != nil {
		logger.Flush()
	}

	return nil
}

func main() {
	defer func() {
		// 最先恢复终端控制台模式（必须在所有其他操作之前）
		// 这会恢复 ENABLE_LINE_INPUT 和 ENABLE_ECHO_INPUT，让 fmt.Scanln 等正常工作
		platform.RestoreTerminal()

		if logger != nil {
			logger.Close()
			fmt.Printf("[DevTools] Log saved to: %s\n", logger.GetPath())
			fmt.Println("Run 'mint-debugger' to analyze")
		}
	}()

	if err := DebugExample(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
