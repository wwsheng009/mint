很好。

现在我们进入 **真正的并发 Fiber 内核级设计**。

目标三件事：

1. ✅ 设计一个支持时间切片的并发 Scheduler
2. ✅ 推导 Transition / 非阻塞更新模型
3. ✅ 做一次架构风险扫描清单

我们会严格保持：

> 单 Fiber Tree
> Lane 优先级模型
> O(n) diff
> 可中断 render
> 不破坏 Commit 原子性

---

# 第一部分：支持时间切片的并发 Scheduler

---

## 一、问题本质

同步模型：

```text
render(root)
→ 一次性跑完整棵树
→ commit
```

问题：

* 大树会卡 UI
* 输入无法响应
* 无法中断

我们需要：

> render 可中断
> commit 仍然原子

---

## 二、核心原则

### 1️⃣ Render 可中断

### 2️⃣ Commit 不可中断

### 3️⃣ Lane 控制优先级

### 4️⃣ WorkLoop 可让出控制权

---

## 三、调度核心结构

```go
type Scheduler struct {
    roots []*FiberRoot
    currentTask *Task
}
```

---

### Task

```go
type Task struct {
    Root *FiberRoot
    Lane Lane
    Callback func()
}
```

---

## 四、时间切片模型

我们实现：

```go
const frameBudget = 5 * time.Millisecond
```

核心逻辑：

```go
func workLoopConcurrent(root *FiberRoot, lane Lane) {
    deadline := time.Now().Add(frameBudget)

    for nextUnitOfWork != nil {

        if time.Now().After(deadline) {
            scheduleContinuation(root, lane)
            return
        }

        nextUnitOfWork = performUnitOfWork(nextUnitOfWork)
    }

    root.finishedWork = root.workInProgress
}
```

---

## 五、关键：双 Fiber 树模型

必须有：

```go
Current Tree
WorkInProgress Tree
```

结构：

```go
fiber.Alternate
```

Render 时：

* 只改 workInProgress
* 不动 current

Commit 时：

```go
root.current = root.finishedWork
```

原子替换。

---

## 六、优先级抢占

当：

```text
正在 render NormalLane
突然来了 InputLane
```

我们：

1. 保存当前 nextUnitOfWork
2. 切换到更高 Lane
3. render 完高优先级
4. 再恢复低优先级

核心函数：

```go
func getHighestPriorityLane(lanes Lane) Lane {
    return lanes & -lanes
}
```

---

## 七、完整并发流程

```text
ScheduleUpdate
    ↓
markRootUpdated
    ↓
ensureRootScheduled
    ↓
performConcurrentWorkOnRoot
    ↓
workLoopConcurrent
    ↓
yield if needed
    ↓
finish
    ↓
commit
```

---

# 第二部分：Transition / 非阻塞更新模型

现在我们做真正高级的。

---

## 一、问题

用户输入：

```text
Input → setState
```

同时：

```text
FilterList → 重渲染 5000 项
```

如果同步：

输入卡住。

我们要：

> 输入立即响应
> 重计算延后

---

## 二、Lane 设计扩展

新增：

```go
const (
    SyncLane Lane = 1 << iota
    InputLane
    TransitionLane
    NormalLane
    IdleLane
)
```

---

## 三、startTransition 模型

API：

```go
func StartTransition(fn func())
```

实现：

```go
func StartTransition(fn func()) {
    currentLane = TransitionLane
    fn()
    currentLane = DefaultLane
}
```

---

## 四、调度行为

优先级顺序：

```text
Sync > Input > Transition > Normal > Idle
```

当 Transition 正在渲染：

如果 InputLane 更新来了：

立即中断 Transition。

---

## 五、Transition 的核心机制

### 1️⃣ 可中断

### 2️⃣ 可丢弃

如果中途来了更高优先级更新：

* 丢弃当前 WIP
* 重新 render

因为：

> render 是纯函数阶段

---

## 六、避免闪烁（双树保护）

Transition 不会立即 commit。

只有当：

```text
Transition render 完成
且没有更高优先级
```

才 commit。

这保证 UI 不闪烁。

---

## 七、非阻塞更新模型总结

| 类型   | Lane           | 是否可中断 | 是否可丢弃 |
| ---- | -------------- | ----- | ----- |
| Sync | SyncLane       | 否     | 否     |
| 输入   | InputLane      | 可     | 可     |
| 过渡   | TransitionLane | 可     | 可     |
| 空闲   | IdleLane       | 可     | 可     |

---

# 第三部分：架构风险扫描清单

这是最重要的部分。

---

# Ⅰ. 结构风险

### ❌ 是否存在多棵 Fiber 主树？

必须：

> 全局逻辑只有一棵 Fiber Tree

---

### ❌ 是否 clone VNode？

VNode 不能参与 identity。

---

### ❌ 是否 Layer 影响结构？

Layer 只能排序，不能重排 Fiber。

---

# Ⅱ. 调度风险

### ❌ 是否 render 修改 current 树？

必须：

```text
render → workInProgress
commit → 切换指针
```

---

### ❌ 是否 commit 可中断？

Commit 必须原子。

---

### ❌ 是否没有 lane 抢占？

没有优先级抢占 = 假并发。

---

# Ⅲ. 事件系统风险

### ❌ 是否命中基于结构剥离？

必须基于：

```text
LayoutBox + Layer 排序
```

---

### ❌ 冒泡是否按 Layer？

必须按 Fiber Parent。

---

# Ⅳ. Diff 风险

### ❌ 是否 index 作为 key？

会状态错位。

---

### ❌ 是否 O(n²) diff？

必须 oldMap + lastPlacedIndex。

---

# Ⅴ. 并发一致性风险

### ❌ 是否共享 mutable 状态？

render 阶段不能写全局状态。

---

### ❌ 是否在 render 中做 IO？

render 必须纯。

---

# Ⅵ. Portal 风险

### ❌ 是否 Portal 创建新 Fiber 树？

Portal 只能改变 HostRoot。

---

# 七、最终并发架构图

```text
                    ┌──────────────┐
                    │   Scheduler  │
                    │  (Lane + TS) │
                    └──────┬───────┘
                           ↓
                 ┌──────────────────┐
                 │ Concurrent Render │
                 │   WorkInProgress  │
                 └──────┬────────────┘
                        ↓
                 ┌──────────────────┐
                 │   FinishedWork   │
                 └──────┬────────────┘
                        ↓
                 ┌──────────────────┐
                 │     Commit       │
                 │   (Atomic Swap)  │
                 └──────────────────┘
```

---

# 八、最终判断标准

如果你现在满足：

* 双树模型
* Lane 优先级
* 时间切片
* 可中断 render
* 原子 commit
* O(n) diff
* Layer 只排序
* Portal 只改变 HostRoot

那你的 runtime：

> 已经达到 React 18 并发架构级别。

---

如果你愿意，

我可以下一步推导：

* Suspense / 延迟显示模型
* 异步数据加载与调度协作
* Fiber 内存优化策略
* 或者帮你做一次“性能瓶颈预测分析”

你现在真的已经在做一个 UI 引擎内核了。
