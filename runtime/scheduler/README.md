# Scheduler - 更新调度器

Scheduler 提供**更新批处理**和**分帧渲染**功能，用于优化 TUI 应用的渲染性能。

## 核心功能

- **更新批处理 (Batching)**: 将多个脏更新合并为一次渲染
- **优先级处理**: High → Normal → Low 顺序处理
- **时间切片**: 每帧限制处理时间，避免阻塞
- **脏队列管理**: 高效跟踪需要更新的组件

## 基本使用

```go
import "github.com/wwsheng009/mint/runtime/scheduler"
```

### 创建调度器

```go
// 默认配置：每优先级 2ms 时间预算
s := scheduler.New()

// 自定义时间预算
s := scheduler.NewWithBudget(5 * time.Millisecond)

// 完整配置
s := scheduler.NewWithConfig(
    2*time.Millisecond,  // 时间预算
    16*time.Millisecond, // 批次最大时长 (~60fps)
    1000,                // 批次最大数量
)
```

### 标记组件为脏

```go
// 标记需要布局和绘制
s.MarkDirty("button1", buttonNode, priority.DirtyHigh)

// 仅标记需要布局
s.MarkLayoutDirty("panel1", panelNode, priority.DirtyNormal)

// 仅标记需要绘制
s.MarkPaintDirty("text1", textNode, priority.DirtyLow)
```

### 批量更新

```go
// 开始批处理 - 标记的更新会先缓存
s.BeginBatch()

for _, item := range newItems {
    s.MarkDirty(item.ID, item.Node, priority.DirtyNormal)
}

// 结束批处理并刷新
s.EndBatch(true) // true = 立即刷新到脏队列
```

### 处理更新

```go
// 定义渲染器
type MyRenderer struct {
    // ...
}

func (r *MyRenderer) Layout(node interface{}) {
    if n, ok := node.(*MyComponent); ok {
        n.PerformLayout()
    }
}

func (r *MyRenderer) Paint(node interface{}) {
    if n, ok := node.(*MyComponent); ok {
        n.Repaint()
    }
}

// 处理一帧
renderer := &MyRenderer{}
result := s.ProcessNext(renderer, scheduler.DefaultProcessOptions())

fmt.Printf("处理: %d, 剩余: %d, 超时: %v\n",
    result.Processed, result.Remaining, result.OutOfTime)
```

## 高级用法

### 自动批处理刷新

```go
s.BeginBatch()

for {
    // 添加更新...
    s.MarkDirty(id, node, priority.DirtyNormal)

    // 检查是否应该刷新
    if s.ShouldFlush() {
        s.FlushBatch()
        // 处理一帧...
        break
    }
}
```

### 限制处理数量

```go
// 每帧最多处理 10 个节点
opts := scheduler.ProcessOptions{
    MaxNodes: 10,
}
s.ProcessNext(renderer, opts)
```

### 处理特定优先级

```go
// 只处理高优先级更新
opts := scheduler.ProcessOptions{
    PriorityLevels: []priority.DirtyLevel{priority.DirtyHigh},
}
s.ProcessNext(renderer, opts)
```

### 统计信息

```go
// 获取各优先级的脏节点数量
counts := s.DirtyCount()
fmt.Printf("High: %d, Normal: %d, Low: %d\n",
    counts[priority.DirtyHigh],
    counts[priority.DirtyNormal],
    counts[priority.DirtyLow])

// 获取总脏节点数
total := s.TotalDirtyCount()

// 检查批处理状态
isBatching := s.IsBatching()
batchSize := s.GetBatchSize()
```

## 与框架集成

### 在 App 中使用

```go
type App struct {
    scheduler *scheduler.Scheduler
    renderer  *MyRenderer
}

func NewApp() *App {
    return &App{
        scheduler: scheduler.New(),
        renderer:  &MyRenderer{},
    }
}

func (a *App) MarkComponentDirty(c *Component, level priority.DirtyLevel) {
    a.scheduler.MarkDirty(c.ID(), c, level)
}

func (a *App) RenderFrame() {
    result := a.scheduler.ProcessNext(a.renderer, scheduler.DefaultProcessOptions())

    // 如果有剩余，在下一帧继续处理
    if result.Remaining > 0 {
        // 计划下一帧
    }
}

// 批量更新场景
func (a *App) UpdateMany(items []Item) {
    a.scheduler.BeginBatch()
    defer a.scheduler.EndBatch(true)

    for _, item := range items {
        a.scheduler.MarkDirty(item.ID, item.Component, priority.DirtyNormal)
    }
}
```

### 自定义时间预算

```go
// 根据目标帧率调整
fps := 60
msPerFrame := 1000 / fps
msPerPriority := msPerFrame / 3 // 分给 3 个优先级

s := scheduler.NewWithBudget(time.Duration(msPerPriority) * time.Millisecond)
```

## API 参考

### 类型

| 类型 | 说明 |
|------|------|
| `Scheduler` | 调度器 |
| `DirtyNode` | 脏节点信息 |
| `UpdateBatch` | 批次信息 |
| `Renderer` | 渲染器接口 |
| `ProcessOptions` | 处理选项 |
| `ProcessResult` | 处理结果 |

### 方法

#### Scheduler

| 方法 | 说明 |
|------|------|
| `New()` | 创建默认调度器 |
| `NewWithBudget(dur)` | 创建指定时间预算的调度器 |
| `MarkDirty(id, node, level)` | 标记节点为脏 |
| `MarkLayoutDirty(id, node, level)` | 标记需要布局 |
| `MarkPaintDirty(id, node, level)` | 标记需要绘制 |
| `BeginBatch()` | 开始批处理 |
| `EndBatch(flush)` | 结束批处理 |
| `FlushBatch()` | 刷新批次 |
| `ShouldFlush()` | 检查是否应该刷新 |
| `ProcessNext(renderer, opts)` | 处理下一批更新 |
| `Clear()` | 清空所有脏节点 |
| `DirtyCount()` | 获取各优先级脏节点数 |
| `TotalDirtyCount()` | 获取总脏节点数 |

#### ProcessOptions

| 字段 | 说明 |
|------|------|
| `TimeBudget` | 时间限制 |
| `MaxNodes` | 最大处理节点数 |
| `PriorityLevels` | 要处理的优先级列表 |

#### ProcessResult

| 字段 | 说明 |
|------|------|
| `Processed` | 已处理节点数 |
| `Remaining` | 剩余脏节点数 |
| `OutOfTime` | 是否因时间不足停止 |
| `ByPriority` | 各优先级处理数量 |

## 最佳实践

1. **批量更新**: 一次更新多个组件时使用 `BeginBatch/EndBatch`
2. **合理优先级**: 关键交互用 High，普通内容用 Normal，装饰用 Low
3. **时间预算**: 根据目标帧率调整，一般 2-5ms per priority
4. **监控剩余**: `Remaining > 0` 时考虑下一帧继续处理

## 示例

完整示例请参考 `scheduler_test.go`。
