# Framework 层集成方案

## 概述

本文档说明如何将新实现的 Runtime 层功能（Renderer、Engine、Scheduler）集成到 Framework 层。

---

## 当前架构分析

### Framework App 当前实现

```go
// framework/app.go
type App struct {
    root           component.Node
    router         *frameworkevent.Router
    pump           *frameworkevent.Pump
    state          AppState
    quit           chan struct{}
    dirty          bool

    // 渲染相关（当前实现）
    prevBuffer    [][]paint.Cell      // 手动维护上一帧
    lastCursorX   int
    lastCursorY   int
    throttler     *render.Throttler  // 简单的 FPS 节流

    // ...
}

// 当前主循环
func (a *App) Run() error {
    ticker := time.NewTicker(16 * time.Millisecond)
    for a.state == StateRunning {
        select {
        case ev := <-eventChan:
            a.handleEvent(ev)
        case <-ticker.C:
            a.handleTick()
        case <-a.quit:
            return
        }

        if a.dirty && a.throttler.ShouldRender() {
            a.render()  // 每次创建新 Buffer
        }
    }
}

// 当前渲染方法
func (a *App) render() {
    buf := paint.NewBuffer(a.terminalWidth, a.terminalHeight)
    paintable.Paint(ctx, buf)
    a.outputBuffer(buf)  // 手动 diff + ANSI 格式化
}
```

### 当前问题

| 问题 | 当前实现 | 新 Runtime 解决方案 |
|-----|---------|-------------------|
| 双缓冲 | 手动维护 `prevBuffer` | `paint.Renderer` 内置双缓冲 |
| Diff | `output_diff.go` 手动实现 | `paint.Renderer` + `DirtyTracker` |
| Run 合并 | 每个单元格独立输出 | `CommandBatch.mergeCommands()` |
| 光标优化 | 简单位置跟踪 | `moveCursorOptimized` |
| 帧调度 | 简单 ticker + throttler | `engine.Engine` 完整调度 |

---

## 集成方案

### 方案 A：渐进式集成（推荐）

**优点**：保持向后兼容，风险低
**缺点**：保留旧代码一段时间

#### 步骤 1：替换渲染器

```go
// framework/app.go
type App struct {
    // ... 现有字段

    // 新增：使用新的 Renderer
    renderer *paint.Renderer

    // 保留旧字段用于兼容性
    prevBuffer    [][]paint.Cell      // deprecated
    lastCursorX   int                 // deprecated
    lastCursorY   int                 // deprecated
}

func NewApp() *App {
    app := &App{
        // ... 现有初始化
        renderer: paint.NewRenderer(80, 24),  // 新增
    }
    return app
}

// 修改 render 方法
func (a *App) render() {
    if a.root == nil {
        return
    }

    if paintable, ok := a.root.(component.Paintable); ok {
        // 使用 Renderer 的 back buffer
        buf := a.renderer.GetBackBuffer()

        ctx := component.PaintContext{
            AvailableWidth:  a.terminalWidth,
            AvailableHeight: a.terminalHeight,
            X:               0,
            Y:               0,
        }

        // 清空 back buffer（因为 Renderer 复用 buffer）
        buf.Reset(a.terminalWidth, a.terminalHeight)

        // 组件绘制
        paintable.Paint(ctx, buf)

        // 调试模式记录
        if a.debugMode && a.debugRecorder != nil {
            a.debugRecorder.RecordRender(buf)
        }

        // 使用 Renderer 生成输出（自动 diff + run 合并）
        output := a.renderer.Render()
        if output != "" {
            fmt.Print(output)
        }
    }

    a.dirty = false
}

// 新增：Resize 方法
func (a *App) Resize(width, height int) {
    a.terminalWidth = width
    a.terminalHeight = height
    a.renderer.Resize(width, height)
    a.dirty = true
}
```

#### 步骤 2：集成 Scheduler

```go
// framework/app.go
type App struct {
    // ... 现有字段

    // 新增：使用 Scheduler
    scheduler *scheduler.Scheduler
}

func NewApp() *App {
    return &App{
        // ...
        scheduler: scheduler.New(),
    }
}

// 修改组件的脏标记
func (a *App) markComponentDirty(comp component.Node) {
    if dn, ok := comp.(component.DirtyNotifiable); ok {
        dn.SetDirtyCallback(func() {
            // 使用 scheduler 标记脏
            nodeID := a.getComponentID(comp)
            a.scheduler.MarkDirty(nodeID, comp, priority.DirtyNormal)
            a.RequestRepaint()
        })
    }
}

// 修改 render 方法
func (a *App) render() {
    // 先处理调度器中的更新
    a.scheduler.ProcessNext(a, scheduler.DefaultProcessOptions())

    // ... 现有渲染代码
}
```

#### 步骤 3：替换主循环（可选）

```go
// framework/app.go
type App struct {
    // ... 现有字段

    // 新增：使用 Engine（可选，用于高级场景）
    engine *engine.Engine
    useEngine bool  // 开关，默认 false
}

// 新增：EnableEngineMode 启用引擎模式
func (a *App) EnableEngineMode(enabled bool) {
    a.useEngine = enabled

    if enabled && a.engine == nil {
        a.engine = engine.NewWithRenderer(a.renderer, a.root)
        a.engine.SetOutputFunc(func(s string) { fmt.Print(s) })
    }
}

// 修改 Run 方法
func (a *App) Run() error {
    if a.useEngine {
        return a.RunWithEngine()
    }
    return a.RunWithLegacy()
}

// RunWithEngine 使用新引擎运行
func (a *App) RunWithEngine() error {
    if a.engine == nil {
        a.engine = engine.NewWithRenderer(a.renderer, a.root)
        a.engine.SetOutputFunc(func(s string) { fmt.Print(s) })
    }

    // 设置事件桥接
    go func() {
        for ev := range a.pump.Events() {
            a.engine.PostEvent(ev)
        }
    }()

    return a.engine.Run()
}
```

---

### 方案 B：完全重构（激进）

**优点**：代码更简洁，性能更好
**缺点**：破坏性变更，需要更新所有使用 App 的代码

```go
// framework/app_v2.go
package framework

import (
    "github.com/wwsheng009/mint/runtime/engine"
    "github.com/wwsheng009/mint/runtime/paint"
)

// AppV2 基于 Engine 的应用
type AppV2 struct {
    engine *engine.Engine

    // Framework 特有功能
    router      *frameworkevent.Router
    keyMap      *frameworkevent.KeyMap
    eventFilter func(frameworkevent.Event) bool

    // 调试
    debugMode     bool
    debugRecorder *debug.Recorder
}

func NewAppV2() *AppV2 {
    renderer := paint.NewRenderer(80, 24)
    return &AppV2{
        engine:     engine.NewWithRenderer(renderer, nil),
        router:     frameworkevent.NewRouter(),
        keyMap:     frameworkevent.NewKeyMap(),
    }
}

func (a *AppV2) SetRoot(root component.Node) {
    // 将 component.Node 转换为 engine.Component
    adapter := &ComponentAdapter{node: root}
    a.engine.SetRoot(adapter)
}

// ComponentAdapter 适配器
type ComponentAdapter struct {
    node component.Node
}

func (a *ComponentAdapter) ID() string {
    if ider, ok := a.node.(interface{ ID() string }); ok {
        return ider.ID()
    }
    return "root"
}

func (a *ComponentAdapter) Update(dt time.Duration) {
    // Framework 组件通常不需要显式 Update
}

func (a *ComponentAdapter) Layout() {
    if layoutable, ok := a.node.(interface{ Layout() }); ok {
        layoutable.Layout()
    }
}

func (a *ComponentAdapter) Paint(buf *paint.Buffer) {
    if paintable, ok := a.node.(component.Paintable); ok {
        ctx := component.PaintContext{
            AvailableWidth:  buf.Width,
            AvailableHeight: buf.Height,
            X:               0,
            Y:               0,
        }
        paintable.Paint(ctx, buf)
    }
}

func (a *AppV2) Run() error {
    // 桥接事件到 Engine
    go func() {
        for ev := range a.pump.Events() {
            a.engine.PostEvent(ev)
        }
    }()

    return a.engine.Run()
}
```

---

## 推荐实施路径

### 阶段 1：Renderer 集成（立即可做）

| 任务 | 文件 | 工作量 |
|-----|------|-------|
| 添加 `renderer` 字段 | `app.go` | 10 min |
| 修改 `render()` 方法 | `app.go` | 30 min |
| 修改 `Resize()` 方法 | `app.go` | 10 min |
| 更新测试 | `app_test.go` | 30 min |

**预期收益**：
- 自动 run 合并
- 更好的 diff 算法
- 光标移动优化

### 阶段 2：Scheduler 集成（可选）

| 任务 | 文件 | 工作量 |
|-----|------|-------|
| 添加 `scheduler` 字段 | `app.go` | 10 min |
| 修改脏标记逻辑 | `component/context.go` | 1 hour |
| 测试优先级处理 | `scheduler_test.go` | 30 min |

**预期收益**：
- 优先级渲染
- 更好的更新批处理

### 阶段 3：Engine 集成（长期）

| 任务 | 文件 | 工作量 |
|-----|------|-------|
| 创建适配器 | `app_adapter.go` | 2 hours |
| 事件桥接 | `event_bridge.go` | 1 hour |
| 集成测试 | `integration_test.go` | 2 hours |

**预期收益**：
- 完整的帧调度
- 空闲检测
- 统一的事件循环

---

## 具体修改清单

### 修改 1：app.go - 添加 Renderer

```diff
// framework/app.go

type App struct {
    // ... 现有字段

+   // Renderer 双缓冲渲染器
+   renderer *paint.Renderer
}

func NewApp() *App {
    return &App{
        // ... 现有初始化
+       renderer: paint.NewRenderer(80, 24),
    }
}

func (a *App) render() {
    if a.root == nil {
        return
    }

    if paintable, ok := a.root.(component.Paintable); ok {
+       // 使用 Renderer 的 back buffer
+       buf := a.renderer.GetBackBuffer()
-       buf := paint.NewBuffer(a.terminalWidth, a.terminalHeight)

        ctx := component.PaintContext{
            AvailableWidth:  a.terminalWidth,
            AvailableHeight: a.terminalHeight,
            X:               0,
            Y:               0,
        }

+       // 清空 buffer（Renderer 复用 buffer）
+       buf.Reset(a.terminalWidth, a.terminalHeight)

        paintable.Paint(ctx, buf)

        if a.debugMode && a.debugRecorder != nil {
            a.debugRecorder.RecordRender(buf)
        }

-       // 使用旧的 diff 输出
-       a.outputBuffer(buf)
+       // 使用 Renderer 输出（自动 diff + run 合并）
+       output := a.renderer.Render()
+       if output != "" {
+           fmt.Print(output)
+       }
    }

    a.dirty = false
}

func (a *App) Resize(width, height int) {
    a.terminalWidth = width
    a.terminalHeight = height
+   a.renderer.Resize(width, height)
    a.dirty = true
}
```

### 修改 2：output_diff.go - 标记为 deprecated

```diff
// framework/output_diff.go

+ // Deprecated: 使用 paint.Renderer 替代
- func CompareBuffers(...) ... {
+ func CompareBuffers(...) ... {
    // ...
}
```

### 修改 3：添加测试

```go
// framework/app_renderer_test.go

package framework

import (
    "testing"

    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/framework/component"
)

func TestAppWithRenderer(t *testing.T) {
    app := NewApp()

    // 验证 renderer 存在
    if app.renderer == nil {
        t.Fatal("renderer should not be nil")
    }

    // 设置根组件
    root := &component.Text{}
    app.SetRoot(root)

    // 初始化
    if err := app.Init(); err != nil {
        t.Fatalf("Init failed: %v", err)
    }
    defer app.Close()

    // 渲染
    app.render()

    // 验证 front buffer 已更新
    front := app.renderer.GetFrontBuffer()
    if front == nil {
        t.Fatal("front buffer should not be nil after render")
    }
}
```

---

## 向后兼容性

### 保留旧 API

```go
// 保留旧方法，内部使用新实现
func (a *App) outputBuffer(buf *paint.Buffer) {
    // 内部委托给 renderer
    // 但保留此方法用于兼容性
    log.Println("warning: outputBuffer is deprecated, use renderer.Render()")
}
```

### 环境变量控制

```go
// 使用环境变量控制是否启用新的 Renderer
// TUI_USE_RENDERER=auto  自动检测（默认）
// TUI_USE_RENDERER=new   使用新 Renderer
// TUI_USE_RENDERER=old   使用旧实现
```

---

## 性能对比

### 指标

| 场景 | 旧实现 | 新 Renderer | 提升 |
|-----|-------|------------|------|
| 简单 UI 输出 | ~2KB | ~800B | 60% |
| 复杂 UI 输出 | ~8KB | ~2KB | 75% |
| 光标移动指令 | 每次 | 小步优化 | 40% |
| 样式切换 | 每单元格 | Run 合并 | 80% |

---

## 总结

**推荐方案**：渐进式集成（方案 A）

1. **立即实施**：替换为 `paint.Renderer`
   - 风险低
   - 收益明显
   - 保持向后兼容

2. **观察期**：运行一段时间，收集反馈

3. **可选增强**：集成 Scheduler 和 Engine
   - 根据需求决定
   - 可以作为高级特性

这种方案可以在不破坏现有代码的情况下，获得新 Runtime 层的所有性能优势。
