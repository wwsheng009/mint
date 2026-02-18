很好，这个不是简单 bug。

这是一次**架构冲突暴露**。

你现在的系统里存在：

```
Fiber-first Action 语义模式
        +
VNode 闭包回调模式
```

两套机制在运行期发生冲突。

而你现在的失败，本质不是“某个 nil”，而是：

> 你的事件管线已经优先走 Action path，但 closure 模式依赖 legacy path。

这是一种“半迁移状态”的结构撕裂。

---

# 一、问题本质（架构级）

根据你给的调试文档 

关键点：

```
processMsg 优先走 Action 路径
InputProcessor 对 TargetID == 0 返回 nil
closure 模式 TargetID == 0
→ 事件被丢弃
```

你现在的架构是：

```
Input → InputProcessor → Action
```

但 closure 模式根本不生成 Action。

这就是冲突。

---

# 二、真正冲突点在哪里？

### Fiber-first 设计假设：

> 所有事件都通过 ActionBridge

### closure 模式假设：

> 事件通过 handleMsg 直接调用 vnode 回调

但现在：

```go
if actionRouter != nil && inputProcessor != nil {
    processMsg(msg)
    continue
}
```

你强制所有事件走 Action path。

但 closure 模式没有 ActionTargetID。

这就是断层。

---

# 三、工业级正确方向

你必须选一个方向：

---

## 方案 A（推荐）：彻底 Action 化（终极方向）

禁止 closure 模式。

Button 不再持有 onClick 函数。

Builder：

```go
OnClick(func(){})
```

自动：

* 生成 ActionID
* 注册到 Scope Dispatcher
* Fiber 保存 ID

也就是说：

> closure 只是声明期语法糖
> 运行期全部是 Action 机制

这样 InputProcessor 永远返回 Action。

没有 TargetID==0 的情况。

---

## 方案 B（临时兼容）：ActionProcessor 支持 Fiber 路由

不要在 InputProcessor 阶段拒绝 TargetID==0。

改为：

```go
func processMouseMsg(mouseMsg *MouseMsg) *Action {

    // ❌ 删除这个
    // if mouseMsg.TargetID == 0 { return nil }

    return &Action{
        Type: ActionClick,
        Payload: mouseMsg,
        // 不设置 TargetID
    }
}
```

然后让 ActionBridge 决定：

```go
if fiber.ActionTargetID != "" {
    // 语义模式
} else if fiber.FocusableVNode != nil {
    // closure 模式
}
```

这样所有事件都变成 Action。

InputProcessor 不再丢弃。

---

# 四、你当前最可能的真正错误

根据你描述：

> 修复 fallback 仍然失败

我怀疑真正问题是：

```
Fiber.FocusableVNode 在 reconcile 后丢失
```

极可能发生在：

* cloneFiber()
* workInProgress 构建
* commit 阶段

很多人会忘记复制这个字段。

你要确认：

```go
func cloneFiber(f *Fiber) *Fiber {
    return &Fiber{
        ...
        FocusableVNode: f.FocusableVNode, // 是否复制？
    }
}
```

如果没复制：

HitMap 会指向 current 树

但点击后 dispatch 走的是 WIP 树

导致：

```
TargetFiber != current tree fiber
FocusableVNode 为空
```

事件自然不触发。

---

# 五、真正推荐的架构修复（长期）

不要让 closure 依赖 FocusableVNode。

让 Button 在构建时：

```
closure → 注册 ActionID → Scope Dispatcher
```

Fiber 永远只有：

```
ActionTargetID
```

删除：

```
Fiber.FocusableVNode
```

删除：

```
handleMsg legacy path
```

删除：

```
InputProcessor nil fallback
```

统一模型：

```
Input → Action → Scheduler → Dispatcher
```

只有一个通道。

---

# 六、你的系统当前状态诊断

从你现象看，极可能是：

### 1️⃣ InputProcessor 把 closure 模式挡掉了

### 2️⃣ 或 Fiber.FocusableVNode 在 commit 后丢失

### 3️⃣ 或 HitMap 指向旧 fiber

优先排查顺序：

1. 打印 mouseMsg.TargetFiber 是否为 current fiber
2. 打印 fiber.FocusableVNode 是否非 nil
3. 打印 Button.onClick 是否非 nil
4. 打印 commit 后 fiber 地址是否变化

---

# 七、真正的架构建议（必须做的决定）

你现在必须决定：

---

## 你要：

### A）兼容双模式（复杂）

需要：

* Action 路径
* Legacy 路径
* 双重 hitTest
* 双重 dispatcher

复杂度翻倍。

---

### B）统一 Action 模型（推荐）

closure 只存在于：

```
Builder 阶段
```

运行期永远是：

```
ActionTargetID + Dispatcher
```

这是唯一长期稳定的结构。

---

# 八、我给你的明确建议

你已经：

* 有 Fiber-first
* 有 Scheduler
* 有 Lane
* 有 Scope

现在不要退回 legacy。

彻底统一。

删掉：

```
FocusableVNode 运行期依赖
legacy handleMsg
InputProcessor nil 判断
```

把 closure 转换为 ActionID。

---

# 九、一句话总结

你现在的问题不是 bug。

是：

> 运行期混用了两套事件系统。

要么统一 Action。
要么永远处于结构撕裂状态。

---

如果你愿意，我可以：

* 🔥 给你一份“彻底统一 Action Runtime 的重构步骤清单”
* 🔥 或帮你设计一个“Closure → Action 编译期桥接系统”
* 🔥 或帮你画出当前系统的撕裂点结构图

你现在正处于架构进化的关键节点。

很好。

你现在已经明确走 **Fiber-first + Scheduler + Lane** 方向，那就不要再犹豫了。

我们一次性做完整设计：

1. ✅ 彻底统一 Action Runtime 的重构步骤清单
2. 🔥 Closure → Action 桥接系统设计（声明期转换）
3. 🔥 当前系统撕裂点结构图

我会按“工业级可落地重构顺序”写。

---

# 一、当前系统的撕裂点结构图

你现在实际上有两条事件通道：

```
                ┌──────────────┐
                │    Input     │
                └──────┬───────┘
                       │
        ┌──────────────┴──────────────┐
        ▼                             ▼
  Action Path                    Legacy Path
(processMsg)                    (handleMsg)
        │                             │
InputProcessor                  FocusableVNode
        │                             │
ActionBridge                  vnode.onClick()
        │
Dispatcher
```

问题：

* Action Path 优先
* InputProcessor 会丢弃 TargetID == 0
* closure 模式没有 TargetID
* 于是事件丢失

这不是 bug。

这是双系统冲突。

---

# 二、目标结构（唯一通道）

我们要变成：

```
Input
  ↓
InputProcessor
  ↓
Action
  ↓
ActionBridge
  ↓
Scheduler
  ↓
Dispatcher
  ↓
Component Logic
  ↓
runtime.Update
```

没有 legacy。

没有 vnode runtime。

没有 FocusableVNode。

---

# 三、🔥 彻底统一 Action Runtime 重构步骤清单

必须按顺序做。

---

## Step 1：删除 Legacy 入口

删除：

```
App.handleMsg
FocusableVNode.HandleAction
legacy fallback
```

强制：

```
所有事件都必须生成 Action
```

InputProcessor 永远返回 Action（不再返回 nil）。

---

## Step 2：删除 Fiber.FocusableVNode

Fiber 结构中移除：

```go
FocusableVNode
```

Fiber 只允许：

```go
ActionTargetID string
```

Fiber 是结构节点。

不是行为节点。

---

## Step 3：InputProcessor 不再依赖 TargetID

改成：

```go
func processMouseMsg(msg *MouseMsg) *Action {
    return &Action{
        Type: ActionClick,
        Payload: ClickPayload{...},
        // TargetID 不在这里设置
    }
}
```

TargetID 由 ActionBridge 决定。

---

## Step 4：ActionBridge 统一路由

```go
func (b *ActionBridge) DispatchFromFiber(
    fiber *Fiber,
    action *Action,
) bool {

    for f := fiber; f != nil; f = f.Parent {

        if f.ActionTargetID != "" {
            action.TargetID = f.ActionTargetID
            return dispatcher.Dispatch(action)
        }
    }

    return false
}
```

唯一逻辑来源是 Fiber.ActionTargetID。

---

## Step 5：删除 vnode 运行期依赖

* 运行期不能访问 vnode
* vnode 只用于构建 Fiber
* commit 后不再读取 vnode

---

## Step 6：闭包只存在于声明期

closure 不再被 Fiber 持有。

而是：

```
Builder → 注册 ActionID → Scope Dispatcher
```

---

# 四、🔥 Closure → Action 编译期桥接系统设计

这是关键。

你要保留优雅语法：

```go
Button().
    OnClick(func() {
        setShowModal(true)
    })
```

但运行期必须是：

```
ActionID
```

---

## 1️⃣ Builder 阶段转换

```go
func (b *ButtonBuilder) OnClick(handler func()) *ButtonBuilder {

    actionID := runtime.GenerateActionID()

    scope := runtime.CurrentScope()

    scope.Register(actionID, func(a *Action) {
        handler()
    })

    b.actionID = actionID

    return b
}
```

---

## 2️⃣ Scope 结构

```go
type ComponentScope struct {
    dispatcher *Dispatcher
}
```

每个组件 Fiber 对应一个 Scope。

---

## 3️⃣ Fiber 构建阶段

```go
fiber.ActionTargetID = builder.actionID
fiber.Dispatcher = scope.dispatcher
```

---

## 4️⃣ 运行期

点击：

```
Input → Action → Bridge → Dispatcher
```

Dispatcher 调用闭包。

Fiber 不持函数。

VNode 不参与。

---

# 五、🔥 子树隔离（局部 Dispatcher）

每个 Fiber 可以有自己的 Dispatcher。

```
RootDispatcher
   ├── AppScope
   │     ├── ModalScope
   │     └── SidebarScope
```

Dispatch 查找规则：

```
当前 Dispatcher → parent → parent
```

支持：

* Modal 隔离
* Portal
* 子树拦截

---

# 六、🔥 Action 冒泡 + 捕获 + 优先级抢占

统一模型下流程：

```
Build Fiber Path
   ↓
Capture Phase
   ↓
Target Phase
   ↓
Bubble Phase
```

Lane 控制优先级：

```
LaneSync > LaneUser > LaneTransition > LaneIdle
```

Scheduler 永远抢占高优先级。

---

# 七、最终 Runtime 交互图（统一后）

```
                    ┌──────────────┐
                    │    Input     │
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │InputProcessor│
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │    Action    │
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │ ActionBridge │
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │  Scheduler   │
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │ Dispatcher   │
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │ Component    │
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │runtime.Update│
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │  Reconcile   │
                    └──────────────┘
```

唯一通道。

---

# 八、为什么必须统一？

如果不统一，你将永远面对：

* InputProcessor 分叉
* Fiber 与 vnode 交叉引用
* commit 后丢失 closure
* 并发更新错乱
* Scheduler 与 legacy 冲突

统一后：

* 事件可 replay
* 支持 time travel
* 支持并发
* 支持优先级
* 支持 Suspense

---

# 九、你的最佳重构顺序建议

1. 删除 legacy handleMsg
2. 删除 FocusableVNode
3. 删除 InputProcessor nil 逻辑
4. 实现 Closure → Action 注册
5. 验证单 root
6. 再做局部 Dispatcher
7. 最后做 Lane 抢占

不要一次性做完。

逐步迁移。

---

# 十、最终一句话

你现在遇到的不是 bug。

是引擎升级过程中：

> 运行期行为模型未完全统一。

统一 Action Runtime 是唯一长期解。

---

如果你愿意，我可以下一步给你：

* 🔥 一套最小可运行的统一 Action Runtime 代码骨架
* 🔥 或帮你做一次“当前代码的具体重构路径映射”
* 🔥 或给你做一份“Fiber Runtime 成熟度评估报告”

你现在已经在做真正的 UI 引擎核心了。
