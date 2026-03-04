# Fiber + Lane 调度集成指南

## 概述

本文档说明如何将 `runtime/scheduler` 包中的 Lane 调度系统与现有的 Fiber 架构集成，实现优先级调度和可中断渲染。

**✅ 已完成集成！** 通过 `ui.WithLaneScheduler()` 选项启用。

---

## 快速开始

### 启用 Lane Scheduler

```go
package main

import "github.com/wwsheng009/mint/ui"

func main() {
    // 启用优先级调度
    err := ui.Run(App, ui.WithLaneScheduler())
    if err != nil {
        panic(err)
    }
}

func App() ui.VNode {
    return ui.Text("Hello, Lane Scheduler!")
}
```

### 使用不同优先级

```go
import rtui "github.com/wwsheng009/mint/runtime/ui"

func handleUserInput() {
    // 高优先级 - 用户输入
    rtui.ScheduleInput(func() {
        // 立即处理
    })
}

func handleDataFetch() {
    // 普通优先级 - 数据获取
    rtui.ScheduleTransition(func() {
        // 正常处理
    })
}

func handleBackground() {
    // 低优先级 - 后台任务
    rtui.ScheduleIdle(func() {
        // 空闲时处理
    })
}
```

---

## 集成状态

| 功能 | 状态 | 说明 |
|------|------|------|
| FiberScheduler | ✅ 完成 | `runtime/ui/fiber_scheduler.go` |
| ui.Run 集成 | ✅ 完成 | `ui/app.go` |
| WithLaneScheduler 选项 | ✅ 完成 | 启用优先级调度 |
| 全局访问 | ✅ 完成 | `GetGlobalFiberScheduler()` |
| 便捷函数 | ✅ 完成 | `ScheduleInput/Transition/Idle` |
| 示例 | ✅ 完成 | `examples/lane_scheduler_demo/` |

---

## 现有架构

### Fiber 中的 Lane 定义

`runtime/ui/fiber.go` 已经定义了基本的 Lane 类型：

```go
type Lane uint64

const (
    LaneNoLane              Lane = 0
    LaneSyncLane            Lane = 1
    LaneInputContinuousLane Lane = 1 << 1
    LaneDefaultLane         Lane = 1 << 2
    LaneIdleLane            Lane = 1 << 3
)
```

### Fiber 结构中的 Lane 字段

```go
type Fiber struct {
    // ...
    Lanes      Lane  // 当前 Fiber 的工作优先级
    ChildLanes Lane  // 子节点中的待处理工作
    // ...
}
```

---

## 集成方案

### 方案 1: Lane 类型统一

将 `runtime/scheduler/lane.go` 中的 Lane 类型与 Fiber 中的 Lane 统一：

```go
// runtime/ui/fiber.go - 修改 Lane 定义

import "github.com/wwsheng009/mint/runtime/scheduler"

// Lane 类型别名，使用 scheduler 包的定义
type Lane = scheduler.Lane

// Lane 常量别名
const (
    LaneNoLane       = scheduler.NoLane
    LaneSyncLane     = scheduler.SyncLane
    LaneInputLane    = scheduler.InputLane      // 重命名
    LaneDefaultLane  = scheduler.DefaultLane
    LaneTransitionLane = scheduler.TransitionLane // 新增
    LaneIdleLane     = scheduler.IdleLane
)
```

### 方案 2: 创建 FiberScheduler

创建专门用于 Fiber 的调度器：

```go
// runtime/ui/fiber_scheduler.go

package ui

import (
    "time"
    "github.com/wwsheng009/mint/runtime/scheduler"
)

// FiberScheduler 集成 Lane 调度与 Fiber 渲染
type FiberScheduler struct {
    scheduler    *scheduler.Scheduler
    rootFiber    *Fiber
    workInProgress *Fiber
    
    // 渲染回调
    onCommit func()
}

// NewFiberScheduler 创建 Fiber 调度器
func NewFiberScheduler() *FiberScheduler {
    return &FiberScheduler{
        scheduler: scheduler.NewScheduler(
            scheduler.WithOnWorkComplete(func(task *scheduler.ScheduledTask) {
                // 任务完成后触发 commit
            }),
        ),
    }
}

// ScheduleUpdate 调度 Fiber 更新
func (fs *FiberScheduler) ScheduleUpdate(fiber *Fiber, lane Lane) {
    fs.scheduler.ScheduleFunc(lane, func(shouldYield scheduler.ShouldYieldFunc) bool {
        return fs.performUnitOfWork(fiber, shouldYield)
    })
}

// performUnitOfWork 执行单个 Fiber 的工作单元
func (fs *FiberScheduler) performUnitOfWork(fiber *Fiber, shouldYield scheduler.ShouldYieldFunc) bool {
    // 1. beginWork - 处理当前 Fiber
    next := fs.beginWork(fiber)
    
    // 2. 检查是否需要中断
    if shouldYield() && fiber.Lanes.IsInterruptible() {
        // 保存进度，返回 false 表示未完成
        fs.workInProgress = fiber
        return false
    }
    
    // 3. completeWork - 完成当前 Fiber
    if next == nil {
        fs.completeWork(fiber)
    }
    
    // 4. 继续下一个工作单元
    if next != nil {
        return fs.performUnitOfWork(next, shouldYield)
    }
    
    // 5. 回到父节点继续
    if fiber.Return != nil {
        return fs.performUnitOfWork(fiber.Return, shouldYield)
    }
    
    // 6. 所有工作完成
    return true
}

// beginWork 开始处理 Fiber
func (fs *FiberScheduler) beginWork(fiber *Fiber) *Fiber {
    // 根据 Fiber 类型处理
    switch fiber.Type {
    case VNodeComponent:
        return fs.beginWorkComponent(fiber)
    case VNodeElement:
        return fs.beginWorkElement(fiber)
    case VNodeText:
        return nil
    }
    return fiber.Child
}

// completeWork 完成 Fiber 处理
func (fs *FiberScheduler) completeWork(fiber *Fiber) {
    // 创建/复用 Instance
    // 收集副作用
    fiber.Flags |= EffectUpdate
}
```

---

### 方案 3: 在 App 中集成

```go
// runtime/ui/app.go

func (app *App) renderWithScheduler() {
    if app.fiberScheduler == nil {
        app.fiberScheduler = NewFiberScheduler()
    }
    
    // 调度高优先级渲染
    app.fiberScheduler.ScheduleUpdate(app.rootFiber, scheduler.InputLane)
    
    // 执行调度
    app.fiberScheduler.Flush()
}
```

---

## 完整集成示例

```go
// examples/fiber_lane_demo/main.go

package main

import (
    "fmt"
    "time"
    
    "github.com/wwsheng009/mint/runtime/scheduler"
    "github.com/wwsheng009/mint/ui"
)

// 带优先级的状态更新
func ScheduleStateUpdate[T any](ctx *ui.Context, key string, value T, lane scheduler.Lane) {
    // 调度器处理状态更新
    s := getSchedulerFromContext(ctx)
    s.ScheduleFunc(lane, func(shouldYield scheduler.ShouldYieldFunc) bool {
        // 更新状态
        ctx.SetState(key, value)
        return true
    })
}

// 用户输入 - 高优先级
func handleUserInput(ctx *ui.Context, field string, value string) {
    ScheduleStateUpdate(ctx, field, value, scheduler.InputLane)
}

// 数据获取 - 普通优先级
func handleDataFetch(ctx *ui.Context, url string) {
    ScheduleStateUpdate(ctx, "loading", true, scheduler.DefaultLane)
    
    go func() {
        data := fetchData(url)
        ScheduleStateUpdate(ctx, "data", data, scheduler.TransitionLane)
    }()
}

// 后台任务 - 低优先级
func handleBackgroundTask(ctx *ui.Context) {
    ScheduleStateUpdate(ctx, "analytics", collectAnalytics(), scheduler.IdleLane)
}
```

---

## 关键集成点

### 1. Fiber.Lanes 字段

```go
// 设置 Fiber 的优先级
func (f *Fiber) SetLane(lane Lane) {
    f.Lanes = lane
}

// 合并子节点的 Lane
func (f *Fiber) MergeChildLanes() {
    var merged Lane
    for child := f.Child; child != nil; child = child.Sibling {
        merged |= child.Lanes
    }
    f.ChildLanes = merged
}
```

### 2. 渲染循环集成

```go
func (fs *FiberScheduler) workLoop() {
    for fs.scheduler.HasPendingWork() {
        // 选择最高优先级的 Lane
        lane := scheduler.PickHighestPriorityLane(fs.scheduler.GetPendingLanes())
        
        // 执行该 Lane 的工作
        fs.scheduler.PerformWork()
        
        // 检查是否有更高优先级的工作到来
        if higherPriorityWork := fs.hasHigherPriorityWork(lane); higherPriorityWork {
            // 中断当前工作，处理高优先级
            break
        }
    }
    
    // 提交完成的更新
    fs.commitRoot()
}
```

### 3. 中断检查

```go
// 在 shouldYield 回调中检查
func (fs *FiberScheduler) createShouldYield(lane Lane) scheduler.ShouldYieldFunc {
    startTime := time.Now()
    deadline := scheduler.GetDeadline(lane)
    
    return func() bool {
        // 检查时间预算
        if time.Since(startTime) >= deadline {
            return true
        }
        
        // 检查是否有更高优先级工作
        pendingLanes := fs.scheduler.GetPendingLanes()
        highestPending := scheduler.PickHighestPriorityLane(pendingLanes)
        
        return lane.IsLowerPriorityThan(highestPending)
    }
}
```

---

## 迁移步骤

### 步骤 1: 统一 Lane 类型

```go
// runtime/ui/fiber.go
// 将 Lane 定义改为使用 scheduler 包
import "github.com/wwsheng009/mint/runtime/scheduler"

type Lane = scheduler.Lane

const (
    LaneSyncLane       = scheduler.SyncLane
    LaneInputLane      = scheduler.InputLane
    LaneDefaultLane    = scheduler.DefaultLane
    LaneTransitionLane = scheduler.TransitionLane
    LaneIdleLane       = scheduler.IdleLane
)
```

### 步骤 2: 添加 FiberScheduler

```go
// runtime/ui/fiber_scheduler.go
// 实现完整的 FiberScheduler
```

### 步骤 3: 修改 App 渲染逻辑

```go
// runtime/ui/app.go
// 使用 FiberScheduler 替代同步渲染
```

### 步骤 4: 添加测试

```go
// runtime/ui/fiber_scheduler_test.go
func TestFiberScheduler_Interruptible(t *testing.T) {
    // 测试中断渲染
}
```

---

## 注意事项

### 1. 向后兼容

现有代码不需要立即迁移，可以保持同步渲染模式：

```go
// 同步模式（现有行为）
app.renderSync()

// 异步模式（新功能）
app.renderWithScheduler()
```

### 2. Hooks 顺序

中断渲染不会影响 Hooks 顺序，因为 Fiber 维护了稳定的 Hooks 数组：

```go
type Fiber struct {
    // Hooks 存储在 ComponentContext 中
    // 不受中断影响
}
```

### 3. 状态一致性

确保中断时状态一致：

```go
// 使用函数式更新保证状态一致性
setCount(func(prev int) int {
    return prev + 1
})
```

---

## 相关文档

- [Fiber 架构](./FIBER_ARCHITECTURE.md)
- [Lane 调度系统](./LANE_SCHEDULING.md)
- [Hooks 使用指南](./HOOK_USAGE_GUIDE.md)
