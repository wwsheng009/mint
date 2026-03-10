# Lane 调度系统

## 概述

Lane 调度系统是一个优先级驱动的任务调度器，灵感来自 React 的 Lane 系统。它解决了以下问题：

- **UI 阻塞**：大量渲染工作阻塞用户输入响应
- **优先级倒置**：后台任务占用资源导致高优先级任务延迟
- **可中断渲染**：支持 Fiber 架构的中断/恢复机制

## 快速开始

```go
package main

import (
    "fmt"
    "time"
    "github.com/wwsheng009/mint/runtime/scheduler"
)

func main() {
    // 创建调度器
    s := scheduler.NewScheduler()
    defer s.Shutdown()

    // 调度高优先级任务（用户输入）
    s.ScheduleFunc(scheduler.InputLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
        fmt.Println("Processing user input...")
        return true // completed
    })

    // 调度低优先级任务（后台数据预取）
    s.ScheduleFunc(scheduler.IdleLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
        fmt.Println("Prefetching data...")
        // 检查是否应该让出控制权
        if shouldYield() {
            return false // 未完成，稍后继续
        }
        return true
    })

    // 执行所有待处理任务
    s.Flush()
}
```

---

## Lane 优先级

### 优先级层次

从高到低：

| Lane | 优先级 | 用途 | 可中断 |
|------|--------|------|--------|
| `SyncLane` | 最高 | 同步操作，必须立即完成 | ❌ |
| `InputLane` | 高 | 用户输入（键盘、鼠标、表单） | ✅ |
| `AnimationLane` | 中高 | 动画、过渡效果 | ✅ |
| `DefaultLane` | 中 | 默认状态更新 | ✅ |
| `TransitionLane` | 低 | 路由切换、大数据更新 | ✅ |
| `IdleLane` | 最低 | 后台预取、分析、非关键更新 | ✅ |

### Lane 操作

```go
// 合并多个 Lane
lanes := scheduler.MergeLanes(scheduler.InputLane, scheduler.TransitionLane)

// 检查是否包含特定 Lane
if lanes.Includes(scheduler.InputLane) {
    fmt.Println("Has input lane")
}

// 获取最高优先级 Lane
highest := scheduler.PickHighestPriorityLane(lanes) // InputLane

// 比较 Lane 优先级
if scheduler.InputLane.IsHigherPriorityThan(scheduler.TransitionLane) {
    fmt.Println("Input is more important")
}
```

---

## 调度器 API

### 创建调度器

```go
// 基本创建
s := scheduler.NewScheduler()

// 带回调
s := scheduler.NewScheduler(
    scheduler.WithOnWorkStart(func(task *scheduler.ScheduledTask) {
        fmt.Printf("Starting task %d on lane %s\n", task.ID, task.Lane)
    }),
    scheduler.WithOnWorkComplete(func(task *scheduler.ScheduledTask) {
        fmt.Printf("Completed task %d\n", task.ID)
    }),
    scheduler.WithOnWorkYield(func(task *scheduler.ScheduledTask) {
        fmt.Printf("Task %d yielded\n", task.ID)
    }),
)
```

### 调度任务

```go
// 调度简单函数
task := s.ScheduleFunc(scheduler.InputLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
    // 处理用户输入
    return true // 返回 true 表示完成
})

// 取消任务
task.Cancel()
```

### 执行任务

```go
// 同步执行所有待处理任务
s.Flush()

// 执行一次工作循环
s.PerformWork()

// 检查是否有待处理任务
if s.HasPendingWork() {
    s.PerformWork()
}
```

### 批量调度

```go
batch := scheduler.NewBatchScheduler(s, scheduler.TransitionLane)

batch.AddFunc(func(shouldYield scheduler.ShouldYieldFunc) bool {
    // 任务 1
    return true
})

batch.AddFunc(func(shouldYield scheduler.ShouldYieldFunc) bool {
    // 任务 2
    return true
})

batch.Flush()
```

---

## Intent 集成

### IntentWithLane

```go
import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/scheduler"
)

// 为 Intent 分配优先级
myIntent := intent.FieldChangeIntent{Field: "username", Value: "alice"}

// 高优先级包装
wrapped := scheduler.HighPriority(myIntent)

// 低优先级包装
wrapped := scheduler.LowPriority(myIntent)

// 自动推断优先级
wrapped := scheduler.AutoWrap(myIntent)
```

### 优先级助手函数

```go
// 同步（立即执行，不可中断）
scheduler.SyncPriority(intent)

// 高优先级（用户输入）
scheduler.HighPriority(intent)

// 正常优先级（默认）
scheduler.NormalPriority(intent)

// 低优先级（可延迟）
scheduler.LowPriority(intent)

// 后台优先级（空闲时执行）
scheduler.BackgroundPriority(intent)
```

### Intent Batch

```go
batch := scheduler.NewIntentBatch(scheduler.TransitionLane,
    intent1,
    intent2,
    intent3,
)

// 获取所有带 Lane 的 Intent
wrappedIntents := batch.WithLane()
```

---

## 工作原理

### 调度循环

```
┌─────────────────────────────────────────────────────────────┐
│                     Scheduler Loop                          │
│                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    │
│  │ Input Queue │───>│  Animation  │───>│ Transition  │    │
│  │ (High)      │    │   Queue     │    │   Queue     │    │
│  └─────────────┘    └─────────────┘    └─────────────┘    │
│         │                 │                   │            │
│         v                 v                   v            │
│  ┌─────────────────────────────────────────────────┐      │
│  │           PickHighestPriorityLane()             │      │
│  └─────────────────────────────────────────────────┘      │
│                          │                                 │
│                          v                                 │
│  ┌─────────────────────────────────────────────────┐      │
│  │                 Execute Work                    │      │
│  │  - Check shouldYield() for interruptible lanes │      │
│  │  - Resume on next tick if yielded              │      │
│  └─────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### 中断机制

```go
// 长时间运行的任务应该检查 shouldYield
s.ScheduleFunc(scheduler.TransitionLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
    items := getLargeList()

    for i, item := range items {
        // 每处理 100 个项目检查一次
        if i % 100 == 0 && shouldYield() {
            // 保存进度，返回 false 让调度器稍后继续
            saveProgress(i)
            return false
        }

        processItem(item)
    }

    return true // 全部完成
})
```

---

## 最佳实践

### 1. 选择正确的 Lane

```go
// ✅ 用户输入 → InputLane
input.OnChange(func(v string) {
    s.ScheduleFunc(scheduler.InputLane, processInput)
})

// ✅ 数据获取 → DefaultLane 或 TransitionLane
fetchData := func() {
    s.ScheduleFunc(scheduler.TransitionLane, func(shouldYield scheduler.ShouldYieldFunc) bool {
        // 获取数据...
        return true
    })
}

// ✅ 分析日志 → IdleLane
logAnalytics := func() {
    s.ScheduleFunc(scheduler.IdleLane, sendAnalytics)
}
```

### 2. 长任务分片

```go
// ❌ 避免：一次性处理大量数据
processAllItems := func() {
    for _, item := range largeList {
        processItem(item) // 可能阻塞 UI
    }
}

// ✅ 推荐：分片处理
processItemsInChunks := func(shouldYield scheduler.ShouldYieldFunc) bool {
    for i := 0; i < chunkSize && currentIndex < totalItems; i++ {
        processItem(items[currentIndex])
        currentIndex++
    }

    if shouldYield() {
        return false // 让出控制权
    }

    return currentIndex >= totalItems
}
```

### 3. 使用 LaneSelector

```go
selector := scheduler.NewLaneSelector()

// 统一管理 Lane 选择
s.ScheduleFunc(selector.ForUserInput(), handleInput)
s.ScheduleFunc(selector.ForDataFetch(false), fetchBackground)
s.ScheduleFunc(selector.ForAnimation(), runAnimation)
s.ScheduleFunc(selector.ForBackground(), prefetchData)
```

---

## 性能考虑

### Deadline 配置

默认 deadline 间隔：

| Lane | Deadline | 说明 |
|------|----------|------|
| SyncLane | 0 | 无 deadline，必须完成 |
| InputLane | 100ms | 确保响应性 |
| AnimationLane | 33ms | 维持 30fps |
| DefaultLane | 50ms | 平衡性能 |
| TransitionLane | 100ms | 可容忍延迟 |
| IdleLane | 1s | 宽松预算 |

### 自定义 Deadline

```go
// 修改默认 deadline
scheduler.DeadlineInterval[scheduler.AnimationLane] = 16 * time.Millisecond // 60fps
```

---

## 调试

### 监控调度器状态

```go
// 检查队列长度
fmt.Printf("Pending work: %d\n", s.GetTotalQueueLength())
fmt.Printf("Input queue: %d\n", s.GetQueueLength(scheduler.InputLane))

// 检查当前状态
fmt.Printf("Pending lanes: %s\n", s.GetPendingLanes())
fmt.Printf("Is working: %v\n", s.IsPerformingWork())
```

### 使用回调调试

```go
s := scheduler.NewScheduler(
    scheduler.WithOnWorkStart(func(task *scheduler.ScheduledTask) {
        log.Printf("[START] Lane=%s ID=%d", task.Lane, task.ID)
    }),
    scheduler.WithOnWorkComplete(func(task *scheduler.ScheduledTask) {
        log.Printf("[DONE] ID=%d Duration=%v", task.ID, time.Since(task.CreatedAt))
    }),
    scheduler.WithOnWorkYield(func(task *scheduler.ScheduledTask) {
        log.Printf("[YIELD] ID=%d", task.ID)
    }),
)
```

---

## 与 Fiber 架构集成

Lane 调度系统与 Fiber 架构配合使用：

```go
// 在 Fiber 渲染循环中使用
func (r *Reconciler) workLoop() {
    for r.scheduler.HasPendingWork() {
        lane := scheduler.PickHighestPriorityLane(r.scheduler.GetPendingLanes())

        // 执行渲染工作
        r.performUnitOfWork(lane)

        // 检查是否应该中断（低优先级 Lane）
        if lane.IsInterruptible() && r.shouldYield() {
            return // 让出控制权，等待下一帧
        }
    }

    // 提交更新
    r.commitRoot()
}
```

---

## 相关文档

- [Fiber 架构](./FIBER_ARCHITECTURE.md)
- [Store + Reducer 指南](./STORE_REDUCER_GUIDE.md)
- [重构计划](./REFACTOR_PLAN.md)
