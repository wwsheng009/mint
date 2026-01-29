# Scheduler 在 Framework 中的集成

本指南展示如何将 `runtime/scheduler` 集成到 framework 层的 App 中。

## 架构概览

```
framework/App
    │
    ├── sched *scheduler.Scheduler      ← 新增：调度器
    │
    ├── render()                        ← 修改：使用调度器渲染
    │       │
    │       └── sched.ProcessNext()
    │               │
    │               └── FrameworkRenderer
    │                       ├── Layout()
    │                       └── Paint()
    │
    └── createComponentContext()        ← 修改：配置脏回调
            └── SetDirtyCallback()
                    └── sched.MarkDirty()
```

## 集成步骤

### 1. 更新 App 结构

在 `framework/app.go` 中添加 scheduler 字段：

```go
import "github.com/wwsheng009/mint/runtime/scheduler"
import "github.com/wwsheng009/mint/runtime/priority"

type App struct {
    // ... 现有字段

    // 调度器 - 用于优化渲染性能
    sched *scheduler.Scheduler
}
```

### 2. 初始化 Scheduler

在 `NewApp()` 中创建调度器：

```go
func NewApp() *App {
    return &App{
        // ... 现有字段

        // 创建调度器（可选：自定义时间预算）
        sched: scheduler.NewWithBudget(2 * time.Millisecond),
    }
}
```

### 3. 更新组件上下文

修改 `createComponentContext()` 使脏标记与调度器集成：

```go
func (a *App) createComponentContext() *component.ComponentContext {
    ctx := component.NewComponentContext()

    // 脏回调现在使用调度器
    ctx.SetDirtyCallback(func() {
        // 注意：这里我们无法直接访问组件 ID
        // 实际使用时，组件需要通过专门的接口标记自己为脏
        a.dirty = true
    })

    return ctx
}
```

### 4. 创建 Framework Renderer

实现 `scheduler.Renderer` 接口：

```go
// frameworkRenderer 实现 scheduler.Renderer 接口
type frameworkRenderer struct {
    app *App
    buf *paint.Buffer
}

func (r *frameworkRenderer) Layout(node interface{}) {
    // Framework 组件的布局在 Paint 前自动完成
    // 这里可以添加额外的布局逻辑（如果需要）
}

func (r *frameworkRenderer) Paint(node interface{}) {
    if n, ok := node.(component.Paintable); ok {
        x, y := n.GetPosition()
        w, h := n.GetSize()

        ctx := component.PaintContext{
            AvailableWidth:  w,
            AvailableHeight: h,
            X:               x,
            Y:               y,
        }
        n.Paint(ctx, r.buf)
    }
}
```

### 5. 更新 render() 方法

修改渲染逻辑使用调度器：

```go
func (a *App) render() {
    if a.root == nil {
        return
    }

    buf := paint.NewBuffer(a.terminalWidth, a.terminalHeight)

    // 1. 标记根组件为脏（如果需要全局刷新）
    if a.dirty {
        a.markNodeDirty(a.root, priority.DirtyHigh)
        a.dirty = false
    }

    // 2. 使用调度器处理渲染
    renderer := &frameworkRenderer{
        app: a,
        buf: buf,
    }

    result := a.sched.ProcessNext(renderer, scheduler.DefaultProcessOptions())

    // 3. 输出缓冲区到终端
    a.outputBuffer(buf)

    // 调试：显示调度统计
    if a.debugMode && result.Remaining > 0 {
        fmt.Fprintf(os.Stderr, "[SCHEDULER] Processed: %d, Remaining: %d, OutOfTime: %v\n",
            result.Processed, result.Remaining, result.OutOfTime)
    }
}

// markNodeDirty 递归标记节点及其子节点为脏
func (a *App) markNodeDirty(node component.Node, level priority.DirtyLevel) {
    a.sched.MarkDirty(node.ID(), node, level)

    // 递归标记子节点
    if container, ok := node.(component.Container); ok {
        for i := 0; i < container.ChildCount(); i++ {
            child := container.GetChild(i)
            a.markNodeDirty(child, level)
        }
    }
}
```

## 使用示例

### 基础用法

```go
package main

import (
    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/framework/layout"
    "github.com/wwsheng009/mint/framework/component"
)

type MyComponent struct {
    *component.BaseComponent
    app *framework.App
}

func NewMyComponent(app *framework.App) *MyComponent {
    c := &MyComponent{
        BaseComponent: component.NewBaseComponent("mycomp"),
        app:           app,
    }
    return c
}

func (c *MyComponent) OnDataChange(newData string) {
    // 更新状态
    c.SetState("data", newData)

    // 标记为脏（通过 ComponentContext）
    c.MarkDirty()

    // 或者直接使用 App 的调度器（如果有高级需求）
    // c.app.MarkComponentDirty(c, priority.DirtyHigh)
}
```

### 批量更新优化

```go
func (c *MyComponent) LoadItems(items []Item) {
    app := c.app

    // 开始批处理
    app.sched.BeginBatch()
    defer app.sched.EndBatch(true)

    for _, item := range items {
        comp := c.createItemComponent(item)
        c.AddChild(comp)

        // 在批处理期间，这些更新会被合并
        app.sched.MarkDirty(comp.ID(), comp, priority.DirtyNormal)
    }
}
```

### 优先级渲染

```go
func (c *MyComponent) UpdateByPriority(changeType ChangeType) {
    var level priority.DirtyLevel

    switch changeType {
    case ChangeCritical:  // 关键交互
        level = priority.DirtyHigh
    case ChangeContent:   // 内容变化
        level = priority.DirtyNormal
    case ChangeStyle:     // 样式变化
        level = priority.DirtyLow
    }

    c.app.sched.MarkDirty(c.ID(), c, level)
}
```

## 完整集成示例

```go
package framework

import (
    "github.com/wwsheng009/mint/runtime/scheduler"
    "github.com/wwsheng009/mint/runtime/priority"
)

// App 扩展方法 - Scheduler 集成

// MarkComponentDirty 标记组件为脏（使用调度器）
func (a *App) MarkComponentDirty(comp component.Node, level priority.DirtyLevel) {
    if a.sched == nil {
        // 回退到简单模式
        a.dirty = true
        return
    }
    a.sched.MarkDirty(comp.ID(), comp, level)
}

// BeginBatchUpdates 开始批量更新
func (a *App) BeginBatchUpdates() {
    if a.sched != nil {
        a.sched.BeginBatch()
    }
}

// EndBatchUpdates 结束批量更新
func (a *App) EndBatchUpdates(flush bool) {
    if a.sched != nil {
        a.sched.EndBatch(flush)
    }
}

// FlushBatch 刷新批处理队列
func (a *App) FlushBatch() bool {
    if a.sched != nil {
        return a.sched.FlushBatch()
    }
    return false
}

// GetSchedulerStats 获取调度器统计
func (a *App) GetSchedulerStats() map[priority.DirtyLevel]int {
    if a.sched != nil {
        return a.sched.DirtyCount()
    }
    return nil
}

// SetSchedulerBudget 设置调度器时间预算
func (a *App) SetSchedulerBudget(budget time.Duration) {
    if a.sched != nil {
        a.sched.SetTimeBudget(budget)
    }
}
```

## 组件层使用

### 组件实现

```go
package mycomponent

import (
    "github.com/wwsheng009/mint/framework/component"
    "github.com/wwsheng009/mint/runtime/priority"
)

type ListComponent struct {
    *component.BaseComponent
    items []string
}

func (l *ListComponent) AddItem(item string) {
    l.items = append(l.items, item)

    // 标记为脏 - 普通优先级
    l.MarkDirty()
}

func (l *ListComponent) SetError(err error) {
    l.SetState("error", err)

    // 标记为脏 - 高优先级（错误状态更重要）
    l.MarkDirty()
}
```

### 与 State 集成

```go
type StatefulComponent struct {
    *component.BaseComponent
    *component.StateHolder
    store *binding.Store
}

func NewStatefulComponent(store *binding.Store) *StatefulComponent {
    c := &StatefulComponent{
        BaseComponent: component.NewBaseComponent("stateful"),
        StateHolder:   component.NewStateHolder(),
        store:         store,
    }

    // 订阅状态变化
    store.Subscribe("count", func(newValue interface{}) {
        c.SetState("count", newValue)
        c.MarkDirty() // 状态变化时标记为脏
    })

    return c
}
```

## 性能优化场景

### 1. 大列表更新

```go
func (l *ListComponent) LoadMany(items []string) {
    app := l.GetApp()

    app.BeginBatchUpdates()
    defer app.EndBatchUpdates(true)

    for _, item := range items {
        l.AddItem(item)
    }
}
```

### 2. 动画帧

```go
func (c *AnimatedComponent) Animate() {
    app := c.GetApp()

    for {
        // 更新动画状态
        c.UpdateAnimation()

        // 高优先级（需要流畅）
        app.MarkComponentDirty(c, priority.DirtyHigh)

        time.Sleep(16 * time.Millisecond) // ~60fps
    }
}
```

### 3. 后台数据更新

```go
func (c *DataComponent) OnDataReceived(data Data) {
    // 低优先级（不影响交互）
    c.UpdateData(data)
    c.MarkDirty() // 普通优先级
}
```

## 调试

### 查看调度器状态

```go
func (a *App) DebugScheduler() {
    if a.sched == nil {
        fmt.Println("Scheduler not enabled")
        return
    }

    counts := a.sched.DirtyCount()
    fmt.Printf("Dirty nodes: High=%d, Normal=%d, Low=%d\n",
        counts[priority.DirtyHigh],
        counts[priority.DirtyNormal],
        counts[priority.DirtyLow])

    fmt.Printf("Batching: %v, BatchSize: %d\n",
        a.sched.IsBatching(),
        a.sched.GetBatchSize())
}
```

## 配置选项

### 自定义时间预算

```go
// 在 NewApp 中
sched := scheduler.NewWithConfig(
    2*time.Millisecond,  // 每优先级时间预算
    16*time.Millisecond, // 批次最大时长
    1000,                // 批次最大节点数
)
```

### 目标帧率配置

```go
// 根据目标帧率调整
fps := 60
msPerFrame := 1000 / fps
msPerPriority := msPerFrame / 3

sched := scheduler.NewWithBudget(
    time.Duration(msPerPriority) * time.Millisecond,
)
```

## 总结

Scheduler 集成到 Framework 层后：

1. **自动批处理**: 多个组件更新自动合并
2. **优先级渲染**: 重要更新优先处理
3. **时间切片**: 避免单帧渲染时间过长
4. **性能监控**: 可获取渲染统计信息

开发者只需调用 `MarkDirty()`，调度器会自动优化渲染性能。
