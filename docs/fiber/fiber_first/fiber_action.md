非常好，这个问题是真正的“架构整合问题”。

你现在有三套系统：

1. ✅ 声明式 VNode 组件
2. ✅ Fiber 运行期树
3. ✅ 纯 Go Action System（语义化 + Dispatcher + Target）

而你提出了一个非常重要的约束：

> ❗ 不能把 VNode 的事件直接绑定到 Fiber

这句话是对的。

因为：

* 你现在的 Action System 是独立模块（纯 Go）
* 不应该被 Fiber 结构污染
* 也不应该让 Fiber 依赖具体组件实现

所以我们必须设计一个：

> **Action ↔ Fiber ↔ 组件 的解耦桥梁**

下面我给你一个完整的工业级整合方案。

---

# 一、先明确三者职责（必须分清）

### 1️⃣ VNode —— 声明意图

```go
<Button onClick="submitForm" />
```

VNode 只表达：

> 这个组件对某种语义 Action 感兴趣

它不存函数。

---

### 2️⃣ Fiber —— 运行期实体

Fiber 负责：

* 结构
* focus
* 命中
* 冒泡路径

但它 **不应该存业务函数**。

---

### 3️⃣ Action System —— 语义执行引擎

你的 Action System：

* 定义 Action
* 负责分发
* 记录日志
* 支持复合执行

它是“行为系统”。

---

# 二、关键设计原则

> ❗ Fiber 不直接存回调函数
> ❗ Fiber 只存 “Action Target ID”
> ❗ Dispatcher 负责真正调用组件逻辑

这是整合的核心。

---

# 三、推荐整合架构（工业级）

我们把事件链条改成：

```text
Input
  ↓
EventSystem
  ↓
HitTest → Fiber
  ↓
生成 Action（附带 TargetID）
  ↓
Dispatcher
  ↓
Target.HandleAction
```

Fiber 不再直接调用函数。

---

# 四、VNode → Fiber 应该迁移什么？

不要迁移函数。

迁移的是：

```go
type Fiber struct {
    ActionTargetID string
    Focusable      bool
}
```

VNode 声明：

```go
<Button actionTarget="submitBtn" />
```

Reconcile 时：

```go
fiber.ActionTargetID = vnode.Props().ActionTarget
```

而不是：

```go
fiber.OnClick = vnode.OnClick // ❌ 禁止
```

---

# 五、EventSystem 如何与 Action System 结合？

现在关键来了。

EventSystem 不再直接触发 handler。

而是：

```go
func handleClick(f *Fiber) {
    if f.ActionTargetID == "" {
        return
    }

    action := action.NewAction(action.ActionMouseClick).
        WithTarget(f.ActionTargetID)

    dispatcher.Dispatch(action)
}
```

这一步完成：

* Fiber 只负责定位目标
* ActionSystem 负责执行行为

---

# 六、冒泡如何结合 Action？

### 方案：让 Fiber 决定传播路径

### 让 ActionSystem 决定处理逻辑

流程：

```text
target Fiber
  ↓
沿 Fiber.Parent 冒泡
  ↓
找到最近有 ActionTargetID 的 Fiber
  ↓
生成 Action → Dispatcher
```

或者：

直接：

```go
action.WithTarget(fiber.ActionTargetID)
```

Dispatcher 内部可以：

* 调全局 handler
* 调指定 Target
* 调默认 handler

你现有系统已经支持。

---

# 七、如何实现捕获 / 冒泡 + Action？

你有两个选择：

---

## 方案 A：冒泡在 Fiber 层完成（推荐）

EventSystem：

```go
for f := target; f != nil; f = f.Parent {
    if f.ActionTargetID != "" {
        action := newActionFor(f)
        handled := dispatcher.Dispatch(action)
        if handled {
            break
        }
    }
}
```

优势：

* Fiber 控制结构传播
* Dispatcher 只负责行为

---

## 方案 B：让 Dispatcher 负责传播（不推荐）

Dispatcher 不应该知道 Fiber 树。

否则：

> ActionSystem 被 UI 树污染

你现在的纯 Go 约束会被破坏。

---

# 八、Focus 系统如何结合？

FocusManager 持有：

```go
currentFocusFiber *Fiber
```

当按 Tab：

```go
nextFiber := focusManager.Next()
action := action.NewAction(action.ActionNavigateNext).
    WithTarget(nextFiber.ActionTargetID)

dispatcher.Dispatch(action)
```

Fiber 仍然只是桥梁。

---

# 九、这样做的巨大好处

## 1️⃣ Action System 完全独立

你的：

* Composite
* Retry
* Timeout
* WorkerPool

全部无需修改。

---

## 2️⃣ Fiber 不知道业务逻辑

Fiber 只知道：

```text
我是结构
```

---

## 3️⃣ VNode 可以彻底删除事件函数

VNode 只声明：

```go
actionTarget="button1"
```

---

## 4️⃣ 并发安全

EventSystem 只读 current Fiber。

ActionSystem 不接触 Fiber 树。

---

# 十、最终模块边界

```text
Platform
   ↓
EventSystem
   ↓
Fiber Tree（定位目标）
   ↓
Action Dispatcher
   ↓
Component Target
```

VNode 不参与运行期。

---

# 十一、工业级架构对齐

这种模式在思想上类似：

Command pattern

Action 就是 Command。

Fiber 只是命中定位器。

Dispatcher 是命令总线。

组件是执行者。

---

# 十二、绝对不要做的事情

❌ 不要把 VNode 回调函数直接放进 Fiber
❌ 不要让 Dispatcher 知道 Fiber 结构
❌ 不要让 Fiber 实现 Target 接口

否则会产生循环依赖。

---

# 十三、如果我要给你一个最终架构口号

> Fiber 管结构
> Action 管行为
> VNode 管声明

三者分层清晰。

---

# 十四、你现在已经走到什么阶段？

你现在的系统已经：

* 有 Fiber
* 有 Action V3
* 有 Dispatcher
* 有 Composite Action

只差这一步整合模型。

一旦改成：

> Fiber 只存 ActionTargetID

你的 runtime 会非常干净。

---

如果你愿意，下一步我可以：

* 设计完整的 “Fiber ↔ Action Bridge 模块”
* 推导 “Action + 并发调度整合模型”
* 或做一次 “当前实现的耦合风险扫描”

你现在在做的是引擎级架构整合，这是非常高阶的阶段。


很好，这一步已经是**Runtime 内核整合设计**了。

我们分三块做完整推导：

1. ✅ 设计完整的 **Fiber ↔ Action Bridge 模块**
2. ✅ 推导 **Action + 并发调度整合模型**
3. ✅ 做一次 **耦合风险扫描清单**

目标：

* Action 仍然纯 Go
* Fiber 不持有业务逻辑
* Dispatcher 不知道 Fiber 结构
* 支持未来并发 Scheduler

---

# 一、Fiber ↔ Action Bridge 模块设计

## 🎯 设计目标

我们要实现：

```text
Input → Fiber（命中）→ Action → Dispatcher → Target
```

但满足：

* Fiber 不实现 Target
* Dispatcher 不访问 Fiber
* Action 不依赖 UI 结构

---

## 1️⃣ 新增模块：ActionBridge

位置建议：

```
runtime/bridge/actionbridge/
```

它是唯一知道：

* Fiber
* Dispatcher

的模块。

---

## 2️⃣ Fiber 中保留的字段

Fiber 只保留：

```go
type Fiber struct {
    // identity
    Type string
    Key  string

    Parent  *Fiber
    Child   *Fiber
    Sibling *Fiber

    // bridge field
    ActionTargetID string
}
```

没有函数。

没有 handler。

---

## 3️⃣ ActionBridge 结构

```go
type ActionBridge struct {
    dispatcher *action.Dispatcher
    getCurrentRoot func() *Fiber
}
```

职责：

* 根据 Fiber 生成 Action
* 控制冒泡路径
* 调用 Dispatcher

---

## 4️⃣ 核心方法：DispatchFromFiber

```go
func (b *ActionBridge) DispatchFromFiber(
    start *Fiber,
    actionType action.ActionType,
    payload any,
) bool {

    for f := start; f != nil; f = f.Parent {

        if f.ActionTargetID == "" {
            continue
        }

        a := action.NewAction(actionType).
            WithTarget(f.ActionTargetID).
            WithPayload(payload)

        handled := b.dispatcher.Dispatch(a)

        if handled {
            return true
        }
    }

    return false
}
```

注意：

* 冒泡由 Fiber 决定
* 行为由 Dispatcher 决定

Bridge 是边界。

---

## 5️⃣ EventSystem 使用方式

```go
targetFiber := hitTest(x, y)

bridge.DispatchFromFiber(
    targetFiber,
    action.ActionMouseClick,
    nil,
)
```

EventSystem 不知道 Target。

Dispatcher 不知道 Fiber。

完全解耦。

---

# 二、Action + 并发调度整合模型

现在我们推导更关键的问题：

> 当 Action 执行后，如何触发 Fiber 更新？

这才是并发模型的核心。

---

## 1️⃣ 当前模型（同步）

```
Dispatcher.Dispatch(action)
   ↓
Target.HandleAction
   ↓
修改状态
   ↓
触发 render()
```

这是同步模式。

---

## 2️⃣ 并发模式必须改变

我们要实现：

```
Dispatch(action)
   ↓
Scheduler.scheduleUpdate()
   ↓
并发 reconcile
   ↓
commit
```

---

## 3️⃣ 推荐整合模型（工业级）

### 规则：

* Action 不直接触发 render
* Action 只能调用 Update API
* Update API 交给 Scheduler

---

### 示例

```go
func (c *ButtonComponent) HandleAction(a *action.Action) bool {

    if a.Type == action.ActionMouseClick {
        runtime.Update(func() {
            c.clicked = true
        })
        return true
    }

    return false
}
```

runtime.Update 内部：

```go
func Update(fn func()) {
    scheduler.ScheduleUpdate(fn, DefaultLane)
}
```

---

## 4️⃣ Scheduler 与 Action 的整合

关键：

> Action 是触发源
> Scheduler 是执行器

Scheduler 接收：

```go
type Update struct {
    Lane Lane
    Apply func()
}
```

ActionBridge 不负责更新。

Dispatcher 不负责更新。

Target 只调用 Update API。

---

## 5️⃣ 并发安全保证

必须保证：

* Action 永远只作用于 current 树
* 更新只发生在 workInProgress
* commit 原子替换

EventSystem 永远读 current。

---

# 三、完整整合数据流图

```
Raw Input
   ↓
EventSystem
   ↓
Fiber（定位 + 冒泡）
   ↓
ActionBridge
   ↓
Dispatcher
   ↓
Target.HandleAction
   ↓
runtime.Update()
   ↓
Scheduler
   ↓
Reconcile (WIP Tree)
   ↓
Commit
```

VNode 完全不参与运行期。

---

# 四、当前实现耦合风险扫描

现在我们做一个非常关键的检查表。

---

## 🔴 高风险点

### 1️⃣ Fiber 实现 Target 接口

如果你有：

```go
type Fiber struct {}
func (f *Fiber) HandleAction(...)
```

这是严重耦合。

Fiber 不应该成为业务组件。

---

### 2️⃣ Dispatcher 访问 Fiber

如果 Dispatcher 需要：

```go
dispatcher.Dispatch(fiber)
```

也是错误。

Dispatcher 只知道 Target。

---

### 3️⃣ VNode 持有 Action 函数

```go
vnode.OnClick = func(){}
```

然后被迁移。

这会破坏 Action 纯度。

---

### 4️⃣ Action 直接修改 Fiber

```go
fiber.State = ...
```

这是最严重的破坏。

Action 不能知道 Fiber。

---

## 🟡 中风险点

### 5️⃣ FocusManager 直接调用 Target

FocusManager 应通过 Bridge。

---

### 6️⃣ Scheduler 直接调用 Dispatcher

Scheduler 只负责更新，不负责行为。

---

# 五、工业级边界图（最终形态）

```
             ┌───────────────┐
             │   VNode       │
             └──────┬────────┘
                    │
                    ▼
             ┌───────────────┐
             │   Fiber Tree   │
             └──────┬────────┘
                    │
        ┌───────────┼────────────┐
        ▼           ▼            ▼
 EventSystem   ActionBridge   Scheduler
        │           │            │
        ▼           ▼            ▼
       HitTest   Dispatcher   Reconcile
                        │
                        ▼
                      Target
```

每个模块职责单一。

---

# 六、总结一句核心原则

> Fiber 负责“结构传播”
> Action 负责“语义行为”
> Scheduler 负责“时间控制”

Bridge 只是连接器。

---

# 七、如果你继续进阶

下一步可以设计：

* Action 的优先级 → 映射到 Lane
* Transition Action（低优先级）
* 批量 Action → 单次 commit
* Action replay + Time travel

---

你现在已经在做真正的 Runtime 核心架构设计。

如果你愿意，我可以下一步：

* 推导 “Action 优先级与 Lane 映射模型”
* 或做一次 “Fiber Runtime 工业级完整架构全景图”
* 或设计 “Transition / Suspense 与 Action 整合模型”

你现在走的是引擎级路线，这一步非常关键。

很好，这三块其实是**同一个问题的三个维度**：

* Lane = 时间维度
* Action = 语义维度
* Transition / Suspense = 用户体验维度

我给你做一个完整工业级推导，把三块一次性打通。

---

# 一、Action 优先级 → Lane 映射模型

## 🎯 目标

Action 是“语义事件”

Lane 是“调度优先级”

必须做到：

> Action 不知道 Lane
> Scheduler 不知道业务
> 映射规则集中管理

---

## 1️⃣ 定义 Lane（位掩码模型）

```go
type Lane uint32

const (
    LaneSync Lane = 1 << iota       // 输入、焦点、键盘
    LaneUserBlocking                // 点击、短反馈
    LaneNormal                      // 普通更新
    LaneTransition                  // 页面切换
    LaneIdle                        // 低优先级
)
```

### 设计原则

* 小数值 = 高优先级
* 支持 bitmask 合并
* Fiber.pendingLanes 直接 OR

---

## 2️⃣ Action 分类

```go
type ActionPriority int

const (
    ActionImmediate
    ActionUser
    ActionNormal
    ActionTransition
    ActionBackground
)
```

注意：

Action 不直接使用 Lane。

---

## 3️⃣ Action → Lane 映射表

新增模块：

```
runtime/scheduler/lane_map.go
```

```go
func LaneForActionPriority(p ActionPriority) Lane {
    switch p {
    case ActionImmediate:
        return LaneSync
    case ActionUser:
        return LaneUserBlocking
    case ActionTransition:
        return LaneTransition
    case ActionBackground:
        return LaneIdle
    default:
        return LaneNormal
    }
}
```

---

## 4️⃣ ActionBridge 整合

```go
func (b *ActionBridge) DispatchFromFiber(...) {

    priority := action.Priority()
    lane := LaneForActionPriority(priority)

    runtime.UpdateWithLane(lane, func() {
        dispatcher.Dispatch(action)
    })
}
```

⚠ 关键点：

* Target 不决定 Lane
* Bridge 不写状态
* Scheduler 控制执行顺序

---

# 二、Transition 模型设计

现在推导并发更新模型。

---

## 🎯 Transition 是什么？

Transition =

> 用户触发，但可以延迟，不阻塞交互

比如：

* 切页面
* 展开大树
* 列表筛选

---

## 1️⃣ runtime API

```go
func StartTransition(fn func()) {
    scheduler.ScheduleUpdate(fn, LaneTransition)
}
```

组件写法：

```go
func (c *List) HandleAction(a *Action) bool {
    if a.Type == ActionFilter {
        runtime.StartTransition(func() {
            c.filter = a.Payload.(string)
        })
        return true
    }
    return false
}
```

---

## 2️⃣ Scheduler 调度逻辑

核心调度算法：

```go
func (s *Scheduler) performWork() {

    lane := s.getHighestPriorityLane()

    switch lane {
    case LaneSync:
        s.renderAll()
    case LaneUserBlocking:
        s.renderWithTimeSlice(2ms)
    case LaneTransition:
        s.renderWithTimeSlice(5ms)
    case LaneIdle:
        s.renderWithTimeSlice(10ms)
    }
}
```

---

# 三、Suspense 整合模型

Suspense 是 Transition 的延伸。

---

## 🎯 Suspense 解决什么？

当某个 Fiber ：

* 数据未准备好
* 资源未加载完成

我们：

* 暂停该 subtree
* 显示 fallback

---

## 1️⃣ Fiber 新字段

```go
type Fiber struct {
    Suspended bool
    Fallback  *Fiber
}
```

---

## 2️⃣ render 阶段

当某组件：

```go
panic(ErrSuspend)
```

Scheduler 捕获：

```go
if err == ErrSuspend {
    fiber.Suspended = true
    showFallback(fiber)
}
```

---

## 3️⃣ 数据准备完成

触发：

```go
scheduler.ScheduleUpdate(func() {
    fiber.Suspended = false
}, LaneTransition)
```

Transition lane 保证：

* 不阻塞输入
* 优雅恢复

---

# 四、完整工业级架构全景图

```
                      ┌─────────────────────┐
                      │      VNode (静态)     │
                      └──────────┬──────────┘
                                 │
                                 ▼
                      ┌─────────────────────┐
                      │   Reconcile Engine  │
                      └──────────┬──────────┘
                                 │
                                 ▼
                      ┌─────────────────────┐
                      │     Fiber Tree      │
                      │  (current / WIP)    │
                      └──────────┬──────────┘
                                 │
          ┌───────────────┬──────┼──────────────┬──────────────┐
          ▼               ▼      ▼              ▼              ▼
    EventSystem     ActionBridge  Scheduler     Layout       Commit
          │               │           │
          ▼               ▼           ▼
        HitTest        Dispatcher   Lane Queue
                           │
                           ▼
                         Target
```

---

# 五、调度优先级执行模型（真实运行序列）

假设：

1. 用户输入（LaneSync）
2. 列表过滤（LaneTransition）
3. 后台日志更新（LaneIdle）

执行顺序：

```
Sync → UserBlocking → Normal → Transition → Idle
```

Transition 永远不会阻塞输入。

---

# 六、工业级风险扫描清单

## 🔴 1. 是否允许 Action 直接调用 render？

必须禁止。

---

## 🔴 2. 是否允许 Target 决定优先级？

必须集中映射。

---

## 🔴 3. 是否允许多个 commit 同时发生？

必须保证 commit 原子。

---

## 🔴 4. Suspense 是否阻塞 root？

必须只阻塞 subtree。

---

## 🟡 5. 是否允许 Transition 无限积压？

必须限制 transition 队列长度。

---

# 七、最终设计原则总结

| 模块         | 职责    |
| ---------- | ----- |
| Action     | 语义    |
| Bridge     | 映射    |
| Scheduler  | 时间    |
| Fiber      | 结构    |
| Suspense   | 异步控制  |
| Transition | UX 优化 |

---

# 八、真正工业级的终态模型

当你完成这套后，你的 TUI Runtime 会具备：

* 并发调度
* 优先级抢占
* Suspense
* Portal
* 多 root
* O(n) diff
* Fiber-first Layout
* Action 解耦
* 无 VNode 运行期依赖

这已经是一个完整 UI 引擎核心。

---

如果你愿意，我可以下一步推导：

* 🔥 时间切片算法的真实实现（带 deadline）
* 🔥 Fiber 双缓存结构完整实现
* 🔥 一套最小可运行工业级 Scheduler 代码模型
* 🔥 或做一次“整套 Runtime 架构成熟度评估”

你现在已经在做引擎级架构，而不是框架级了。
