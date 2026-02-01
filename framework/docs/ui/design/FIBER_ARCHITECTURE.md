# Fiber 架构设计文档

## 📋 文档概述

**Mint UI Fiber** 是 React 16+ Fiber 架构的 Go 语言实现，用于构建可中断、增量式的声明式 UI 渲染系统。

**版本**: v1.0
**最后更新**: 2026-02-01
**状态**: 设计完成，待实施

---

## 🎯 设计目标

### 1. 可中断渲染 (Interruptible Rendering)

当前渲染问题是同步、不可中断的：

```go
// 当前实现：一次性渲染整个树
func (d *declarativeRoot) Paint(ctx, buffer) {
    vnode := d.appFn()           // 获取 VNode
    d.renderVNode(vnode, buffer) // 直接渲染 - 阻塞
}
```

**Fiber 解决方案**：将渲染工作分解为小单元，可随时中断和恢复：

```go
// Fiber 实现：可中断的工作循环
func WorkLoop(deadline time.Time) {
    for hasMoreWork() {
        if time.Now().After(deadline) {
            return // 时间片用完，中断渲染
        }
        performUnitOfWork()
    }
}
```

### 2. 优先级调度 (Priority Scheduling)

不同更新应该有不同的优先级：

| 优先级 | Lane | 场景 | 示例 |
|--------|------|------|------|
| 同步 | SyncLane | 用户输入 | 点击、按键 |
| 连续输入 | InputContinuousLane | 拖拽、连续输入 | 滑动条、文本输入 |
| 默认 | DefaultLane | 数据更新 | 列表刷新 |
| 空闲 | IdleLane | 低优先级 | 统计分析 |

### 3. 增量渲染 (Incremental Rendering)

不需要等待整棵树渲染完成，可以分帧渲染：

```
Frame 1: 渲染顶层组件
Frame 2: 渲染子组件
Frame 3: 渲染孙组件
...
```

---

## 🏗️ 核心架构

### Fiber 节点结构

```go
// Fiber 表示一个工作单元
type Fiber struct {
    // === VNode 关联 ===
    VNode VNode  // 对应的虚拟节点

    // === 树结构 ===
    Return   *Fiber  // 父节点
    Child    *Fiber  // 第一个子节点
    Sibling  *Fiber  // 下一个兄弟节点
    Alternate *Fiber  // 双缓冲的另一个树

    // === 工作状态 ===
    PendingProps  Props       // 待处理的 props
    MemoizedProps Props       // 记忆化的 props（上次渲染）
    MemoizedState interface{} // 记忆化的 state
    UpdateQueue   *UpdateQueue // 更新队列

    // === Effect 标记 ===
    Flags        EffectFlag   // 当前节点的 effect
    SubtreeFlags EffectFlag   // 子树的 effect

    // === 优先级 ===
    Lanes      Lane   // 当前节点的待处理 lanes
    ChildLanes Lane   // 子树的待处理 lanes

    // === Diff 相关 ===
    Key  string  // 用于 reconciliation
    Type VNodeType  // 节点类型
}
```

### Fiber 工作流程

```
┌─────────────────────────────────────────────────────────────┐
│                    Render Phase (可中断)                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  WorkLoop ──► PerformUnitOfWork ──► BeginWork               │
│       │              │                │                     │
│       │              │                ├─► Reconcile Children │
│       │              │                ├─► Process Updates   │
│       │              │                └─► Create Fiber      │
│       │              │                                     │
│       │              └──► CompleteWork                    │
│       │                      │                             │
│       │                      ├─► Calculate Layout          │
│       │                      ├─► Collect Effects           │
│       │                      └─► Mark Effects              │
│       │                                                    │
│       └──► Check Deadline (时间到则中断)                    │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Commit Phase (不可中断)                     │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  CommitRoot ──► BeforeMutation ──► Mutation                │
│                      │                │                     │
│                      │                ├─► Apply to Buffer    │
│                      │                ├─► Update Refs        │
│                      │                └─► DOM Operations    │
│                      │                                     │
│                      └──► Layout ──► Run Effects               │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## 📊 数据结构详解

### EffectFlag 标记位

```go
type EffectFlag int

const (
    EffectNoEffect EffectFlag = 0   // 无 effect
    EffectPlacement  = 1 << iota     // 新插入
    EffectUpdate                     // 更新
    EffectDeletion                   // 删除
    EffectRef                        // Ref 变化
    EffectSnapshot                   // 快照
)
```

### Lane 优先级系统

```go
type Lane uint64

const (
    LaneNoLane            Lane = 0b00000000  // 无 lane
    LaneSyncLane          Lane = 0b00000001  // 同步（最高优先级）
    LaneInputContinuousLane = 0b00000010  // 连续输入
    LaneDefaultLane       Lane = 0b00000100  // 默认
    LaneIdleLane          Lane = 0b10000000  // 空闲（最低优先级）
)

// LaneRoot 表示所有 lane
const LaneRoot Lane = LaneSyncLane | LaneInputContinuousLane |
                        LaneDefaultLane | LaneIdleLane
```

### Update 更新队列

```go
type Update struct {
    Payload    interface{}  // 更新内容（值或函数）
    Next       *Update      // 下一个更新
    Lane       Lane         // 优先级
    Callback   func()       // 完成后的回调
}

type UpdateQueue struct {
    First *Update  // 队列首
    Last  *Update  // 队列尾
}
```

---

## 🔧 核心算法

### 1. BeginWork - 处理 Fiber

```go
func BeginWork(current, workInProgress *Fiber) *Fiber {
    switch workInProgress.Type {
    case VNodeComponent:
        // 处理组件：调用 render，reconcile 子节点
        return beginWorkComponent(current, workInProgress)
    case VNodeText:
        // 处理文本：检查内容变化
        return beginWorkText(current, workInProgress)
    case VNodeElement:
        // 处理元素：reconcile 子节点
        return beginWorkElement(current, workInProgress)
    }
}
```

### 2. CompleteWork - 完成 Fiber

```go
func CompleteWork(current, workInProgress *Fiber) *Fiber {
    switch workInProgress.Type {
    case VNodeComponent:
        // 收集组件的 effects
        return completeWorkComponent(current, workInProgress)
    case VNodeElement:
        // 计算布局，收集 ref
        return completeWorkElement(current, workInProgress)
    }
}
```

### 3. ReconcileChildren - 子节点协调

```go
func reconcileChildren(
    returnFiber *Fiber,
    currentFirstChild *Fiber,
    newChildren []VNode,
    lanes Lane,
) *Fiber {
    // Phase 1: 简单协调（按索引）
    // Phase 2: Key 协调（通过 key 匹配）

    // 返回第一个子 fiber
}
```

---

## 🔄 状态管理

### 双缓冲机制 (Double Buffering)

```
Current Tree (当前显示)     WorkInProgress Tree (工作中)
┌──────────────┐            ┌──────────────┐
│   Root       │◄──Alternate──►│   Root       │
│   └─ Child   │            │   └─ Child   │
│       └─Child│            │       └─Child│
└──────────────┘            └──────────────┘
     (显示中)                    (计算中)
```

工作完成后，交换角色：

```go
// Commit 后交换
root.Alternate = workInProgress
workInProgress.Alternate = root
root = workInProgress  // 新树成为当前树
```

### 状态持久化

```go
// 组件状态保存在 Fiber 的 MemoizedState
type Fiber struct {
    MemoizedState interface{}  // 持久化状态
}

// Hooks 状态通过 ComponentContext 访问
type ComponentContext struct {
    Hooks []Hook  // Hooks 列表
}

// 每个 Hook 的状态也持久化
type Hook struct {
    Type  HookType
    Value interface{}  // useState 的值等
}
```

---

## ⏱️ 时间切片实现

### 时间预算

```go
type Reconciler struct {
    timeBudget time.Duration  // 每帧的时间预算
    deadline    time.Time     // 当前帧的截止时间
}

// 在 60fps 下，每帧约 16.67ms
// 分配 5ms 给 reconciliation，11.67ms 给渲染
const DefaultTimeBudget = 5 * time.Millisecond
```

### 可中断工作循环

```go
func (r *Reconciler) WorkLoop() error {
    r.deadline = time.Now().Add(r.timeBudget)

    for r.hasMoreWork() {
        // 检查是否超时
        if time.Now().After(r.deadline) {
            // 时间到，中断，请求下一帧继续
            r.requestWork()
            return nil
        }

        // 处理一个工作单元
        unit := r.getNextWorkUnit()
        r.performUnitOfWork(unit)
    }

    // 所有工作完成
    return r.CommitRoot()
}
```

---

## 🎨 集成点

### 与现有系统集成

| 现有组件 | 集成方式 |
|---------|---------|
| `runtime/engine/engine.go` | Fiber 在 frame ticker 中调度 |
| `ui/hooks.go` | 通过 ComponentContext 访问 hooks |
| `ui/instance_manager.go` | ComponentVNode 使用实例管理 |
| `runtime/paint/buffer.go` | Commit 阶段输出到 buffer |

### 渲染流程集成

```
[Engine.Run() 主循环]
    │
    ├─► [事件队列处理]
    │
    ├─► [60fps Ticker]
    │       │
    │       └─► [Reconciler.WorkLoop()] ◄── 可中断
    │               │
    │               ├─► BeginWork (reconcile)
    │               ├─► CompleteWork (finalize)
    │               └─► [检查 deadline]
    │
    └─► [Reconciler.CommitRoot()] ◄── 不可中断
            │
            ├─► renderFiberToBuffer()
            │
            └─► [Buffer 输出到终端]
```

---

## 📁 文件结构

### 新增文件

```
ui/
├── reconciler.go          # Reconciler 核心结构
├── begin_work.go          # BeginWork 阶段实现
├── complete_work.go       # CompleteWork 阶段实现
├── commit.go              # Commit 阶段实现
├── effects.go             # Effect 处理
└── reconcile.go           # 子节点协调算法
```

### 修改文件

```
ui/
├── app.go                 # 集成 Reconciler 到 declarativeRoot
└── fiber.go              # 可能添加 reconciliation 辅助方法
```

---

## ✅ 验收标准

### 功能验收

- [ ] 状态更新触发 reconciliation
- [ ] 渲染可以被时间切片中断
- [ ] 优先级调度生效（同步 > 默认 > 空闲）
- [ ] Effect 正确执行和清理
- [ ] 双缓冲机制工作正常

### 性能验收

- [ ] 大型组件树（1000+ 节点）不阻塞 UI
- [ ] 用户输入响应延迟 < 16ms
- [ ] 内存占用在合理范围
- [ ] 无内存泄漏

### 兼容性验收

- [ ] 现有组件无需修改即可工作
- [ ] Hooks 继续正常工作
- [ ] InstanceManager 继续管理组件实例

---

## 📚 参考资料

- [React Fiber Architecture](https://github.com/acdlite/react-fiber-architecture)
- [React reconciliation](https://react.dev/learn/understanding-reacts-reconciliation)
- [React Scheduler](https://github.com/facebook/react/tree/main/packages/scheduler)

---

**文档版本**: v1.0
**最后更新**: 2026-02-01
**维护者**: Mint UI Team
