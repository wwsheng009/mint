// +build ignore

package main

/*
这是一个性能分析辅助程序，用于统计渲染循环中的热点

使用方法：
go run profile_helper.go

输出：
1. 每秒渲染次数
2. dirty 标记设置次数
3. 各函数调用次数
4. CPU 性能热点定位
*/

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// 性能计数器
// ============================================================

type PerformanceStats struct {
	mu                sync.Mutex
	startTime         time.Time
	renderCount       int64
	dirtySetCount     int64
	tickCount         int64
	eventProcessCount int64
	paintCount        int64
	hitmapBuildCount  int64
	bufferWriteCount  int64
	lastReportTime    time.Time
	lastRenderCount   int64
}


var stats = &PerformanceStats{
	startTime:      time.Now(),
	lastReportTime: time.Now(),
}

func (s *PerformanceStats) RecordRender() {
	atomic.AddInt64(&s.renderCount, 1)
}

func (s *PerformanceStats) RecordDirtySet() {
	atomic.AddInt64(&s.dirtySetCount, 1)
}

func (s *PerformanceStats) RecordTick() {
	atomic.AddInt64(&s.tickCount, 1)
}

func (s *PerformanceStats) RecordEventProcess() {
	atomic.AddInt64(&s.eventProcessCount, 1)
}

func (s *PerformanceStats) RecordPaint() {
	atomic.AddInt64(&s.paintCount, 1)
}

func (s *PerformanceStats) RecordHitMapBuild() {
	atomic.AddInt64(&s.hitmapBuildCount, 1)
}

func (s *PerformanceStats) RecordBufferWrite() {
	atomic.AddInt64(&s.bufferWriteCount, 1)
}

func (s *PerformanceStats) PrintReport() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(s.startTime).Seconds()
	interval := now.Sub(s.lastReportTime).Seconds()

	currentRenderCount := atomic.LoadInt64(&s.renderCount)
	currentDirtySetCount := atomic.LoadInt64(&s.dirtySetCount)
	currentTickCount := atomic.LoadInt64(&s.tickCount)
	currentEventProcessCount := atomic.LoadInt64(&s.eventProcessCount)
	currentPaintCount := atomic.LoadInt64(&s.paintCount)
	currentHitMapBuildCount := atomic.LoadInt64(&s.hitmapBuildCount)
	currentBufferWriteCount := atomic.LoadInt64(&s.bufferWriteCount)

	renderDelta := currentRenderCount - s.lastRenderCount
	fps := float64(renderDelta) / interval

	s.lastRenderCount = currentRenderCount
	s.lastReportTime = now

	fmt.Printf("\n=== 性能统计报告 ===\n")
	fmt.Printf("运行时间: %.2f 秒\n", elapsed)
	fmt.Printf("当前 FPS: %.2f (最近 %.1f 秒)\n", fps, interval)
	fmt.Printf("\n累计操作计数:\n")
	fmt.Printf("  - render() 调用次数: %d (%.2f/s)\n", currentRenderCount, float64(currentRenderCount)/elapsed)
	fmt.Printf("  - dirty=true 设置次数: %d (%.2f/s)\n", currentDirtySetCount, float64(currentDirtySetCount)/elapsed)
	fmt.Printf("  - handleTick() 调用次数: %d (%.2f/s)\n", currentTickCount, float64(currentTickCount)/elapsed)
	fmt.Printf("  - 事件处理次数: %d (%.2f/s)\n", currentEventProcessCount, float64(currentEventProcessCount)/elapsed)
	fmt.Printf("  - Paint() 调用次数: %d (%.2f/s)\n", currentPaintCount, float64(currentPaintCount)/elapsed)
	fmt.Printf("  - HitMap 构建次数: %d (%.2f/s)\n", currentHitMapBuildCount, float64(currentHitMapBuildCount)/elapsed)
	fmt.Printf("  - 缓冲区写入次数: %d (%.2f/s)\n", currentBufferWriteCount, float64(currentBufferWriteCount)/elapsed)

	// 计算指标
	renderEfficiency := 0.0
	if currentTickCount > 0 {
		renderEfficiency = float64(currentRenderCount) / float64(currentTickCount) * 100
	}
	fmt.Printf("\n关键指标:\n")
	fmt.Printf("  - 渲染效率 (render/tick): %.1f%%\n", renderEfficiency)
	fmt.Printf("  - Dirty 设置率: %.2f%%\n", float64(currentDirtySetCount)/float64(currentTickCount)*100)

	// 检查是否存在不必要的渲染
	if renderEfficiency > 90.0 {
		fmt.Printf("\n⚠️  警告: 几乎每次 tick 都在渲染！\n")
		fmt.Printf("   这意味着即使没有内容变化，也在持续重绘。\n")
		fmt.Printf("   可能原因: handleTick() 每次都设置 dirty=true\n")
	}

	fmt.Printf("\n按 Enter 查看实时统计 (自动每 5 秒更新)...\n")
}

func startPerformanceMonitor() {
	go func() {
		// 每 5 秒打印一次统计
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			stats.PrintReport()
		}
	}()

	// 同时监听键盘输入
	go func() {
		for {
			var input string
			fmt.Scanln(&input)
			if input == "" || input == "q" {
				stats.PrintReport()
				if input == "q" {
					os.Exit(0)
				}
			}
		}
	}()
}

// ============================================================
// 插桩版本: 在关键位置添加计数
// ============================================================

// 下面是要插入到 framework/app.go 中的代码片段

/*
// 在 App 结构体中添加:
type App struct {
    // ... 现有字段 ...
    perfStats *PerformanceStats
}

// 修改 handleTick():
func (a *App) handleTick() {
    a.perfStats.RecordTick()    // 记录 tick

    // 定期触发重绘以支持光标闪烁
    // TextInput 会在 Paint 时自己检查时间并切换光标状态
    a.dirty = true
    a.perfStats.RecordDirtySet() // 记录 dirty 设置
}

// 修改 render():
func (a *App) render() {
    a.perfStats.RecordRender()   // 记录渲染调用

    if a.root == nil {
        return
    }

    // ... 现有代码 ...
    paintable.Paint(ctx, buf)
    a.perfStats.RecordPaint()    // 记录 Paint 调用

    // ... HitMap 构建 ...
    a.perfStats.RecordHitMapBuild() // 记录 HitMap 构建

    // ... 输出到终端 ...
    a.perfStats.RecordBufferWrite() // 记录缓冲区写入
}

// 修改 processMsg():
func (a *App) processMsg(msg runtimemsg.Msg) {
    a.perfStats.RecordEventProcess() // 记录事件处理

    // ... 现有代码 ...
}
*/

// ============================================================
// 当前代码分析报告
// ============================================================

func analyzeCurrentCode() {
	fmt.Println("=== 当前代码性能分析 ===\n")

	fmt.Println("1. tick 配置:")
	fmt.Printf("   - ticker 间隔: 16ms (~60 FPS)\n")
	fmt.Printf("   - 每秒 tick 次数: %.1f\n", 1000.0/16.0)

	fmt.Println("\n2. handleTick() 行为:")
	fmt.Println("   - 每次调用都设置 dirty = true")
	fmt.Println("   - 导致: 每 16ms 触发一次完整渲染")
	fmt.Println("   - 影响: 即使应用空闲，也在高频重绘")

	fmt.Println("\n3. 完整渲染流程:")
	fmt.Println("   handleTick() → dirty=true → render()")
	fmt.Println("   render() → Paint() → HitMap构建 → 终端输出")

	fmt.Println("\n4. 性能问题诊断:")
	fmt.Println("   问题1: handleTick() 无差别的标记 dirty=true")
	fmt.Println("       - 即使没有焦点 Input 组件也在重绘")
	fmt.Println("       - 每秒 60 次完整渲染是过量的")
	fmt.Println("")
	fmt.Println("   问题2: Render() 包含昂贵的操作:")
	fmt.Println("       - Paint(): 完整的布局计算 + 样式渲染")
	fmt.Println("       - HitMap 构建: 遍历组件树构建命中映射")
	fmt.Println("       - Renderer.Render(): Diff + 比较输出")
	fmt.Println("")
	fmt.Println("   问题3: 即使内容未改变也执行完整流程:")
	fmt.Println("       - Throttler 只限制最大帧率")
	fmt.Println("       - 但 dirty=true 导致强制渲染")
	fmt.Println("")
	fmt.Println("5. 改进建议:")
	fmt.Println("   方案1: 条件性 dirty 标记")
	fmt.Println("       - 只在有光标组件需要刷新时标记 dirty")
	fmt.Println("")
	fmt.Println("   方案2: 智能节流")
	fmt.Println("       - 空闲时降低 tick 频率 (如 5-10 FPS)")
	fmt.Println("       - 有用户输入时提升帧率")
	fmt.Println("")
	fmt.Println("   方案3: 渐进式渲染")
	fmt.Println("       - 光标闪烁只更新光标区域，不重绘整个界面")
	fmt.Println("")
	fmt.Println("   方案4: 分离 ticker")
	fmt.Println("       - 光标闪烁使用独立的高频 ticker")
	fmt.Println("       - 内容更新使用低频ticker或事件驱动")
}

// ============================================================
// 性能对比测试
// ============================================================

func benchmarkScenarios() {
	fmt.Println("\n=== 性能场景对比 ===\n")

	scenarios := []struct {
		name       string
		fps        int
		renderCost time.Duration
	}{
		{"当前配置 (60 FPS, 完整渲染)", 60, 5 * time.Millisecond},
		{"优化配置 (30 FPS, 完整渲染)", 30, 5 * time.Millisecond},
		{"光标优化 (60 FPS, 部分更新)", 60, 1 * time.Millisecond},
		{"空闲模式 (10 FPS, 仅事件渲染)", 10, 5 * time.Millisecond},
	}

	for _, s := range scenarios {
		cpuUsage := float64(s.fps) * float64(s.renderCost) / 1e6 * 100
		fmt.Printf("\n%s:\n", s.name)
		fmt.Printf("   FPS: %d\n", s.fps)
		fmt.Printf("   单次渲染耗时: %v\n", s.renderCost)
		fmt.Printf("   预估 CPU 占用: %.2f%% (单线程)\n", cpuUsage)
	}
}

// ============================================================
// Main - 运行分析工具
// ============================================================

func main() {
	fmt.Println("=== Mint UI 性能分析工具 ===\n")
	fmt.Println("选择功能:")
	fmt.Println("1. 分析当前代码")
	fmt.Println("2. 运行性能对比基准测试")
	fmt.Println("3. 启动带计数的版本 (需要修改代码)")

	analyzeCurrentCode()
	benchmarkScenarios()

	fmt.Println("\n=== 建议的修复代码 ===\n")
	fmt.Println(`
// 修改 framework/app.go 的 handleTick()
func (a *App) handleTick() {
	// 优化：只在有光标组件时才标记dirty
	hasCursorComponent := false

	if a.focusManager != nil {
		if focused := a.focusManager.GetCurrent(); focused != nil {
			// 检查焦点组件是否需要光标刷新
			if _, ok := focused.Instance.(interface{ HasCursor() bool }); ok {
				hasCursorComponent = true
			}
		}
	}

	if hasCursorComponent {
		a.dirty = true
	}
	// 无光标组件时不标记dirty，避免空闲时的高频重绘
}

// 或使用更激进的方法：降低空闲时的 FPS
func (a *App) handleTick() {
	// 检查距离上次交互的时间
	if time.Since(a.lastInteractionTime) > 2*time.Second {
		// 空闲超过2秒，降低到10 FPS
		a.currentFPS = 10
		a.throttler.SetFPS(a.currentFPS)
	} else {
		// 有交互，恢复到标准 FPS
		a.currentFPS = 30
		a.throttler.SetFPS(a.currentFPS)
	}

	a.dirty = true
}
`)

	// 输出当前 Go 版本和 GOMAXPROCS
	fmt.Printf("\n=== 系统信息 ===\n")
	fmt.Printf("Go 版本: %s\n", runtime.Version())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("CPU 核心数: %d\n", runtime.NumCPU())

	// 显示内存统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("当前分配内存: %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("堆分配内存: %.2f MB\n", float64(m.HeapAlloc)/1024/1024)

	// 计算理论 CPU
	fmt.Printf("\n=== 理论 CPU 占用计算 ===\n")
	fps := 60.0
	renderTimeMs := 5.0 // 假设每次渲染5ms
	cpuPercent := fps * renderTimeMs / 10.0 // 10ms per tick
	fmt.Printf("当前配置理论上限: %.1f%% (FPS=%.0f, 渲染=%.1fms)\n", cpuPercent, fps, renderTimeMs)
	fmt.Printf("注: 实际可能更高，因为 Dirty=true 导致跳过节流器的限制\n")
}
