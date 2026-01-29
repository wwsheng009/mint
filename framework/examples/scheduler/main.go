package main

import (
	"fmt"
	"os"
	"time"

	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/framework/display"
	"github.com/wwsheng009/mint/framework/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/priority"
	"github.com/wwsheng009/mint/runtime/scheduler"
	"github.com/wwsheng009/mint/runtime/style"
)

// SchedulerApp 演示 Scheduler 在 Framework 中的应用
type SchedulerApp struct {
	sched        *scheduler.Scheduler
	root         *layout.Flex
	prevBuffer   [][]paint.Cell
	prevStatusMsg string

	// UI 组件
	title   *display.Text
	clock   *ClockWidget
	counter *CounterWidget
	logPanel *LogPanel

	// 状态
	quit      bool
	statusMsg string
}

// ClockWidget 实时时钟组件（高优先级更新）
type ClockWidget struct {
	*component.BaseComponent
	*component.StateHolder
	app      *SchedulerApp
	lastTime string
}

func NewClockWidget(app *SchedulerApp) *ClockWidget {
	return &ClockWidget{
		BaseComponent: component.NewBaseComponent("clock"),
		StateHolder:   component.NewStateHolder(),
		app:           app,
		lastTime:      "",
	}
}

func (c *ClockWidget) Measure(maxW, maxH int) (int, int) {
	return 20, 1
}

func (c *ClockWidget) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	now := time.Now().Format("15:04:05")
	s := style.Style{}.Foreground(style.Cyan).Bold(true)
	// 使用 SetString 正确处理宽字符
	buf.SetString(ctx.X, ctx.Y, now, s)
}

func (c *ClockWidget) Start() {
	// 初始标记根容器为脏（它会递归渲染所有子组件）
	c.app.sched.MarkDirty(c.app.root.ID(), c.app.root, priority.DirtyHigh)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for !c.app.quit {
			<-ticker.C
			// 标记根容器为脏，触发整个 UI 重新渲染
			c.app.sched.MarkDirty(c.app.root.ID(), c.app.root, priority.DirtyHigh)
		}
	}()
}

// CounterWidget 计数器组件（普通优先级）
type CounterWidget struct {
	*component.BaseComponent
	*component.StateHolder
	app   *SchedulerApp
	count int
}

func NewCounterWidget(app *SchedulerApp) *CounterWidget {
	return &CounterWidget{
		BaseComponent: component.NewBaseComponent("counter"),
		StateHolder:   component.NewStateHolder(),
		app:           app,
		count:         0,
	}
}

func (c *CounterWidget) Measure(maxW, maxH int) (int, int) {
	return 30, 3
}

func (c *CounterWidget) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	label := fmt.Sprintf("计数: %d", c.count)
	s1 := style.Style{}.Foreground(style.Green)

	info := "按 Space 增加计数"
	s2 := style.Style{}.Foreground(style.BrightBlack)

	hint := "按 + 快速增加 (+10)"
	s3 := style.Style{}.Foreground(style.Yellow)

	// for i, r := range label {
	// 	buf.SetCell(ctx.X+i, ctx.Y, r, s1)
	// }
	// 使用 SetString 正确处理宽字符
	buf.SetString(ctx.X, ctx.Y, label, s1)
	// for i, r := range info {
	// 	buf.SetCell(ctx.X+i, ctx.Y+1, r, s2)
	// }
	// 使用 SetString 正确处理宽字符
	buf.SetString(ctx.X, ctx.Y, info, s2)
	// for i, r := range hint {
	// 	buf.SetCell(ctx.X+i, ctx.Y+2, r, s3)
	// }
	// 使用 SetString 正确处理宽字符
	buf.SetString(ctx.X, ctx.Y, hint, s3)
}

func (c *CounterWidget) Increment() {
	c.count++
	// 标记根容器为脏
	c.app.sched.MarkDirty(c.app.root.ID(), c.app.root, priority.DirtyNormal)
}

func (c *CounterWidget) Add(n int) {
	c.count += n
	// 标记根容器为脏
	c.app.sched.MarkDirty(c.app.root.ID(), c.app.root, priority.DirtyNormal)
}

// LogPanel 日志面板（低优先级）
type LogPanel struct {
	*component.BaseComponent
	*component.StateHolder
	app     *SchedulerApp
	logs    []string
	maxLogs int
}

func NewLogPanel(app *SchedulerApp) *LogPanel {
	return &LogPanel{
		BaseComponent: component.NewBaseComponent("logpanel"),
		StateHolder:   component.NewStateHolder(),
		app:           app,
		logs:          make([]string, 0, 100),
		maxLogs:       10,
	}
}

func (l *LogPanel) Measure(maxW, maxH int) (int, int) {
	return 40, 11
}

func (l *LogPanel) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	titleStyle := style.Style{}.Foreground(style.Blue).Bold(true)
	title := "=== 日志面板 ==="
	// 使用 SetString 正确处理宽字符
	buf.SetString(ctx.X, ctx.Y, title, titleStyle)

	logStyle := style.Style{}.Foreground(style.White)
	for i, log := range l.logs {
		if i >= l.maxLogs {
			break
		}
		// 使用 SetString 正确处理宽字符
		buf.SetString(ctx.X, ctx.Y+1+i, log, logStyle)
	}

	// 填充空行
	emptyStyle := style.Style{}.Foreground(style.BrightBlack)
	for i := len(l.logs); i < l.maxLogs; i++ {
		for j := 0; j < ctx.AvailableWidth; j++ {
			buf.SetCell(ctx.X+j, ctx.Y+1+i, ' ', emptyStyle)
		}
	}
}

func (l *LogPanel) AddLog(log string) {
	l.logs = append(l.logs, log)
	if len(l.logs) > l.maxLogs {
		l.logs = l.logs[len(l.logs)-l.maxLogs:]
	}
	// 标记根容器为脏
	l.app.sched.MarkDirty(l.app.root.ID(), l.app.root, priority.DirtyLow)
}

func main() {
	app := &SchedulerApp{
		sched: scheduler.NewWithBudget(2 * time.Millisecond),
		quit:  false,
	}

	app.buildUI()

	fmt.Println("=== Scheduler 演示程序 ===")
	fmt.Println("操作说明:")
	fmt.Println("  Space - 增加计数")
	fmt.Println("  +     - 快速增加 (+10)")
	fmt.Println("  l     - 添加日志")
	fmt.Println("  b     - 批量添加日志 (演示批处理)")
	fmt.Println("  s     - 显示调度器统计")
	fmt.Println("  q     - 退出")
	fmt.Println()

	// 启动时钟
	app.clock.Start()

	// 模拟后台日志
	go app.backgroundLogs()

	// 简单的事件循环（实际应用中应该使用 framework.App）
	app.runEventLoop()

	fmt.Println("\n程序已退出")
}

func (a *SchedulerApp) buildUI() {
	a.title = display.NewText("Scheduler 演示")
	a.title.SetStyle(style.Style{}.Foreground(style.Cyan).Bold(true))

	a.clock = NewClockWidget(a)
	a.counter = NewCounterWidget(a)

	a.logPanel = NewLogPanel(a)

	// 构建布局 - 修复 AddChild 调用链
	titleRow := layout.NewRow().
		MainAlign(layout.MainCenter).
		AddChild(a.title)

	leftCol := layout.NewColumn().
		Gap(1).
		AddChild(a.clock).
		AddChild(a.counter)

	contentRow := layout.NewRow().
		Gap(2).
		AddChild(leftCol).
		AddChild(a.logPanel)

	mainCol := layout.NewColumn().
		Gap(1).
		Padding(1).
		AddChild(titleRow).
		AddChild(contentRow)

	a.root = mainCol
}

func (a *SchedulerApp) runEventLoop() {
	// 简化版事件循环
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	input := make(chan byte)
	go func() {
		var b [1]byte
		for {
			n, _ := os.Stdin.Read(b[:])
			if n > 0 {
				input <- b[0]
			}
		}
	}()

	// 设置终端为原始模式（简化）
	fmt.Print("\x1b[2J\x1b[H") // 清屏
	defer fmt.Print("\x1b[0m") // 重置

	renderBuf := paint.NewBuffer(80, 24)
	statusMsg := "就绪 - 按 q 退出"

	for !a.quit {
		select {
		case ch := <-input:
			a.handleInput(ch)
		case <-ticker.C:
			a.render(renderBuf, statusMsg)
		}
	}
}

func (a *SchedulerApp) handleInput(ch byte) {
	switch ch {
	case 'q', 'Q':
		a.quit = true
	case ' ':
		a.counter.Increment()
		a.statusMsg = "计数 +1"
	case '+':
		a.counter.Add(10)
		a.statusMsg = "计数 +10"
	case 'l', 'L':
		a.logPanel.AddLog(fmt.Sprintf("[LOG] 日志条目 %d", time.Now().Unix()))
		a.statusMsg = "添加日志"
	case 'b', 'B':
		a.batchAddLogs()
		a.statusMsg = "批量添加日志"
	case 's', 'S':
		a.showStats()
	}
}

func (a *SchedulerApp) batchAddLogs() {
	// 演示批处理
	a.sched.BeginBatch()
	defer a.sched.EndBatch(true)

	for i := 0; i < 5; i++ {
		a.logPanel.AddLog(fmt.Sprintf("[BATCH] 日志 %d", i+1))
	}

	fmt.Printf("\n[批处理] 添加了 5 条日志，批次大小: %d\n", a.sched.GetBatchSize())
}

func (a *SchedulerApp) showStats() {
	counts := a.sched.DirtyCount()
	fmt.Printf("\n=== 调度器统计 ===\n")
	fmt.Printf("高优先级:   %d\n", counts[priority.DirtyHigh])
	fmt.Printf("普通优先级: %d\n", counts[priority.DirtyNormal])
	fmt.Printf("低优先级:   %d\n", counts[priority.DirtyLow])
	fmt.Printf("总脏节点:   %d\n", a.sched.TotalDirtyCount())
	fmt.Printf("批处理中:   %v\n", a.sched.IsBatching())
	fmt.Println("==================")
}

func (a *SchedulerApp) backgroundLogs() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for !a.quit {
		select {
		case <-ticker.C:
			a.logPanel.AddLog(fmt.Sprintf("[BG] 后台日志 %s", time.Now().Format("15:04:05")))
		}
	}
}

func (a *SchedulerApp) render(buf *paint.Buffer, statusMsg string) {
	// 每次渲染时确保根组件大小正确
	a.root.SetPosition(0, 0)
	a.root.SetSize(buf.Width, buf.Height)

	// 使用调度器渲染 - Flex 容器会递归渲染子组件
	renderer := &appRenderer{
		buf: buf,
		app: a,
	}

	// 首次渲染时，只标记 Flex 容器（它们会递归渲染子组件）
	if a.sched.TotalDirtyCount() == 0 {
		a.markFlexContainersDirty(a.root)
	}

	// 处理脏节点
	result := a.sched.ProcessNext(renderer, scheduler.DefaultProcessOptions())

	// 只有当有处理结果时才输出
	if result.Processed > 0 {
		a.simpleOutput(buf, statusMsg)
	}
}

// markAllDirty 递归标记所有组件为脏（仅首次渲染使用）
func (a *SchedulerApp) markAllDirtyRecursive(node component.Node) {
	a.sched.MarkDirty(node.ID(), node, priority.DirtyHigh)

	// 递归标记所有子组件
	if container, ok := node.(interface{ GetChildren() []component.Node }); ok {
		for _, child := range container.GetChildren() {
			a.markAllDirtyRecursive(child)
		}
	}
}

// markFlexContainersDirty 只标记 Flex 容器为脏
func (a *SchedulerApp) markFlexContainersDirty(node component.Node) {
	if _, isFlex := node.(*layout.Flex); isFlex {
		a.sched.MarkDirty(node.ID(), node, priority.DirtyHigh)
	}
	// 递归处理子组件
	if container, ok := node.(interface{ GetChildren() []component.Node }); ok {
		for _, child := range container.GetChildren() {
			a.markFlexContainersDirty(child)
		}
	}
}

func (a *SchedulerApp) simpleOutput(buf *paint.Buffer, statusMsg string) {
	// 移动光标到左上角
	fmt.Print("\x1b[H")

	// 确保缓冲区大小正确
	a.prevBuffer = ensureBufferSize(a.prevBuffer, buf.Width, buf.Height)

	// 只输出变化的单元格
	for y := 0; y < buf.Height-1; y++ {
		x := 0
		for x < buf.Width {
			cell := buf.Cells[y][x]
			prevCell := a.prevBuffer[y][x]

			// 使用辅助函数跳过延续单元格
			if paint.ShouldSkipCell(cell) {
				x++
				continue
			}

			// 使用辅助函数检查是否有变化
			if paint.IsCellChanged(cell, prevCell) {
				// 移动到该位置
				fmt.Printf("\x1b[%d;%dH", y+1, x+1)

				// 输出字符
				if cell.Char != 0 {
					if cell.Style != (style.Style{}) {
						fmt.Print(cell.Style.ToANSI())
					}
					fmt.Print(string(cell.Char))
				} else {
					fmt.Print(" ")
				}
			}

			// 使用辅助函数获取字符宽度来增加位置
			width := paint.GetCellWidth(cell)
			if width > 0 {
				x += width
			} else {
				x++
			}
		}
	}

	// 状态消息
	if statusMsg != a.prevStatusMsg {
		fmt.Printf("\x1b[%d;1H", buf.Height)
		fmt.Print("\x1b[K")
		fmt.Print(statusMsg)
		a.prevStatusMsg = statusMsg
	}

	// 更新 prevBuffer
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			a.prevBuffer[y][x] = buf.Cells[y][x]
		}
	}
}

// ensureBufferSize 确保缓冲区大小正确
func ensureBufferSize(buf [][]paint.Cell, w, h int) [][]paint.Cell {
	if buf == nil || len(buf) != h || len(buf[0]) != w {
		newBuf := make([][]paint.Cell, h)
		for i := range newBuf {
			newBuf[i] = make([]paint.Cell, w)
		}
		return newBuf
	}
	return buf
}

// appRenderer 实现 scheduler.Renderer
type appRenderer struct {
	buf *paint.Buffer
	app *SchedulerApp
}

func (r *appRenderer) Layout(node interface{}) {
	// 布局在 Paint 之前由 Flex 自动处理
}

func (r *appRenderer) Paint(node interface{}) {
	if n, ok := node.(component.Paintable); ok {
		// 使用 component.Positionable 和 component.Sizable 接口
		var x, y int
		if pos, ok := node.(component.Positionable); ok {
			x, y = pos.GetPosition()
		}
		var w, h int
		if sz, ok := node.(component.Sizable); ok {
			w, h = sz.GetSize()
		}

		ctx := component.PaintContext{
			AvailableWidth:  w,
			AvailableHeight: h,
			X:               x,
			Y:               y,
		}
		n.Paint(ctx, r.buf)
	}
}
