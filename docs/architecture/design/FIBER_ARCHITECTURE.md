# Fiber Architecture

**版本**: v1.0
**创建时间**: 2026-03-04
**状态**: ✅ 已实现

---

## 概述

Mint UI 采用 Fiber 架构，类似于 React Fiber。Fiber 是一种**增量渲染**算法，将渲染工作分解为小单元，支持：

- **可中断渲染** - 高优先级任务可以打断低优先级任务
- **时间切片** - 避免长时间阻塞主线程
- **优先级调度** - 用户输入优先于数据获取
- **时间旅行调试** - 记录状态历史，支持撤销/重做

---

## 核心概念

### 1. Fiber 节点

每个 Fiber 节点代表 UI 中的一个组件：

```go
type Fiber struct {
    // 树结构
    Return   *Fiber  // 父节点
    Child    *Fiber  // 第一个子节点
    Sibling  *Fiber  // 下一个兄弟节点
    Alternate *Fiber // 双缓存：指向另一棵树的对应节点

    // 标识
    Type     VNodeType
    Tag      string
    Key      string
    NodeID   uint64

    // 状态
    Props         Props
    MemoizedProps Props
    MemoizedState interface{}
    Instance      ComponentInstance

    // 效果标记
    Flags       EffectFlag
    SubtreeFlags EffectFlag

    // 优先级
    Lanes      Lane
    ChildLanes Lane
}
```

### 2. 双缓存 (Double Buffering)

```
Current Tree (屏幕显示) ←→ WorkInProgress Tree (正在构建)
         ↑                              ↑
    Alternate ────────────────────────Alternate
```

- **Current Tree**: 当前显示的 Fiber 树
- **WorkInProgress Tree**: 正在构建的新树
- **Alternate**: 两棵树互相指向对方

**优势**：
- 无闪烁更新：新树构建完成后一次性切换
- 可中断：构建过程中不显示不完整的 UI
- 复用：节点可以在两棵树之间复用

### 3. Lane 优先级系统

```go
const (
    LaneSyncLane            Lane = 1       // 最高优先级，同步执行
    LaneInputContinuousLane Lane = 1 << 1  // 连续输入（拖拽、打字）
    LaneDefaultLane         Lane = 1 << 2  // 默认更新
    LaneIdleLane            Lane = 1 << 3  // 低优先级，空闲时执行
)
```

**调度规则**：
1. 高优先级任务打断低优先级任务
2. 同优先级任务按顺序执行
3. 低优先级任务在空闲时执行

---

## 渲染流程

### Phase 1: Render (可中断)

```go
func renderRoot(root *Fiber) {
    workInProgress := root

    for workInProgress != nil {
        // 检查是否应该中断
        if shouldYield() {
            return // 保存进度，等待恢复
        }

        // 处理当前节点
        workInProgress = performUnitOfWork(workInProgress)
    }
}
```

**beginWork** - 创建/复用子节点：
```go
func beginWork(fiber *Fiber) *Fiber {
    switch fiber.Type {
    case VNodeElement:
        return beginWorkElement(fiber)
    case VNodeComponent:
        return beginWorkComponent(fiber)
    case VNodeText:
        return beginWorkText(fiber)
    }
    return nil
}
```

**completeWork** - 完成节点处理：
```go
func completeWork(fiber *Fiber) {
    // 创建/更新 Instance
    if fiber.Instance == nil {
        fiber.Instance = createInstance(fiber)
    }

    // 收集 Effect
    if fiber.Flags != EffectNoEffect {
        appendEffect(fiber)
    }
}
```

### Phase 2: Commit (不可中断)

```go
func commitRoot(root *Fiber) {
    // 1. 执行删除
    commitDeletions()

    // 2. 执行插入和更新
    commitPlacement(root)
    commitUpdate(root)

    // 3. 执行 Effects
    commitEffects(root)

    // 4. 切换 Current 指针
    currentTree = workInProgressTree
}
```

---

## Hooks 系统

### 设计原则

**关键**：Hooks 不捕获闭包状态，而是通过索引访问 Fiber 中的 Hook 数组：

```go
func UseState[T any](initial T) (T, func(T)) {
    ctx := GetCurrentContext()
    hook := ctx.GetOrCreateHook(HookState)

    // 初始化
    if !hook.Initialized {
        hook.Value = initial
        hook.Initialized = true
    }

    // 获取当前值
    value := hook.Value.(T)

    // 创建 setter（不捕获值，只捕获索引）
    hookIndex := ctx.HookIndex - 1
    setter := func(newValue T) {
        ctx.Hooks[hookIndex].Value = newValue
        scheduleUpdate()
    }

    return value, setter
}
```

### Hook 类型

| Hook | 用途 | 存储 |
|------|------|------|
| `UseState` | 组件状态 | `Hook{Type: HookState, Value: state}` |
| `UseEffect` | 副作用 | `Hook{Type: HookEffect, Value: callback, Deps: deps}` |
| `UseMemo` | 计算缓存 | `Hook{Type: HookMemo, Value: result, Deps: deps}` |
| `UseRef` | 持久引用 | `Hook{Type: HookRef, Value: &Ref{Value: initial}}` |

---

## 时间旅行调试

### TimeTravelDebugger

```go
// 创建调试器
dbg := debug.NewTimeTravelDebugger[AppState](
    debug.WithMaxHistory(100),
    debug.WithApplyState(func(s AppState) {
        store.Set(s)
    }),
)

// 订阅状态变化
store.Subscribe(dbg.RecordFunc())

// 导航历史
dbg.Undo()       // 回退一步
dbg.Redo()       // 前进一步
dbg.JumpTo(5)    // 跳转到第 5 步

// 导出/导入
data, _ := dbg.Export()
dbg.Import(data)
```

### 与 Store 集成

```go
func main() {
    store := store.NewStore(AppState{})
    dbg := debug.NewTimeTravelDebugger[AppState]()

    // 订阅记录
    store.Subscribe(dbg.RecordFunc())

    // 创建带调试的 Runtime
    rt := statemachine.NewAppRuntime(
        store,
        reducer,
        view,
    )

    // 调试命令
    go func() {
        for {
            select {
            case <-undoCh:
                dbg.Undo()
            case <-redoCh:
                dbg.Redo()
            }
        }
    }()

    ui.Run(rt)
}
```

---

## Effect 处理

### Effect 标记

```go
const (
    EffectPlacement EffectFlag = 1 << iota  // 新插入
    EffectUpdate                              // 更新
    EffectDeletion                            // 删除
    EffectRef                                 // Ref 变化
    FlagLayoutDirty                           // 需要重新布局
    FlagPaintDirty                            // 需要重绘
)
```

### Effect 执行

```go
func commitEffects(fiber *Fiber) {
    // 执行清理函数
    if fiber.Flags&EffectDeletion != 0 {
        cleanupEffects(fiber)
    }

    // 执行新 Effect
    if fiber.Flags&EffectUpdate != 0 {
        runEffects(fiber)
    }

    // 递归处理子树
    for child := fiber.Child; child != nil; child = child.Sibling {
        commitEffects(child)
    }
}
```

---

## 性能优化

### 1. 批量更新

```go
func (ctx *ComponentContext) SetState(key string, value interface{}) {
    // 添加到待处理队列
    ctx.PendingUpdates[key] = value

    // 只调度一次更新
    if !ctx.UpdateScheduled {
        ctx.UpdateScheduled = true
        scheduleUpdate()
    }
}
```

### 2. 增量渲染

```go
func shouldYield() bool {
    // 检查是否有更高优先级任务
    if hasHigherPriorityWork() {
        return true
    }

    // 检查时间切片是否用尽
    if time.Since(workStartTime) > frameBudget {
        return true
    }

    return false
}
```

### 3. 树 Diff 优化

```go
func reconcileChildren(fiber *Fiber, newChildren []VNode) {
    // 使用 Key 进行匹配
    existingChildren := make(map[string]*Fiber)
    for child := fiber.Child; child != nil; child = child.Sibling {
        existingChildren[child.Key] = child
    }

    // 复用或创建
    for i, newChild := range newChildren {
        if existing, ok := existingChildren[newChild.Key()]; ok {
            // 复用
            updateFiber(existing, newChild)
        } else {
            // 创建新节点
            createFiber(newChild)
        }
    }
}
```

---

## 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                     应用入口                                 │
│                   ui.Run(App)                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Fiber 树管理                              │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │  Current Tree   │ ←→ │ WorkInProgress  │                │
│  │   (显示中)      │    │    (构建中)      │                │
│  └─────────────────┘    └─────────────────┘                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    渲染调度                                  │
│  ┌─────────┐   ┌─────────┐   ┌─────────┐                   │
│  │ Sync    │ → │ Input   │ → │ Default │ → Idle            │
│  │ Lane    │   │ Lane    │   │ Lane    │   Lane            │
│  └─────────┘   └─────────┘   └─────────┘                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    工作循环                                  │
│  beginWork → completeWork → beginWork → ...                 │
│       ↓                                                      │
│  shouldYield? ──Yes──→ 保存进度，等待恢复                    │
│       │                                                      │
│       No                                                     │
│       ↓                                                      │
│  commitRoot (不可中断)                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 相关文档

- [REFACTOR_PLAN.md](/docsArchive/REFACTOR_PLAN.md) - 架构重构计划
- [STORE_REDUCER_GUIDE.md](/docs/ui/store/guides/STORE_REDUCER_GUIDE.md) - Store + Reducer 指南
- [TYPE_SAFE_INTENT.md](/docs/ui/store/features/TYPE_SAFE_INTENT.md) - 类型安全 Intent

---

**最后更新**: 2026-03-04
