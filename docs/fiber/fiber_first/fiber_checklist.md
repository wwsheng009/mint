很好。

我给你三件东西一次性给全：

1. ✅ 一套**最小可运行的统一 Action Runtime 代码骨架**
2. 🔥 你当前系统的**具体重构路径映射**
3. 🔥 一份 **Fiber Runtime 成熟度评估报告**

全部围绕你现在的 Fiber-first 架构。

---

# 第一部分：最小可运行的统一 Action Runtime 骨架

目标：

* 无 legacy
* 无 vnode 运行期依赖
* 无 FocusableVNode
* 单一事件通道
* 支持 Scope + Lane

---

## 1️⃣ 核心数据结构

---

## Action

```go
type Lane uint8

const (
    LaneSync Lane = iota
    LaneUser
    LaneTransition
    LaneIdle
)

type Action struct {
    Type     string
    TargetID string
    Payload  any
    Lane     Lane
}
```

---

## Fiber（极简版）

```go
type Fiber struct {
    Parent   *Fiber
    Child    *Fiber
    Sibling  *Fiber

    ActionTargetID string
    Dispatcher     *Dispatcher

    PendingLanes Lane
}
```

Fiber 不持有函数。

---

## Dispatcher（支持子树）

```go
type ActionHandler func(*Action)

type Dispatcher struct {
    parent   *Dispatcher
    handlers map[string]ActionHandler
}

func NewDispatcher(parent *Dispatcher) *Dispatcher {
    return &Dispatcher{
        parent:   parent,
        handlers: map[string]ActionHandler{},
    }
}

func (d *Dispatcher) Register(id string, h ActionHandler) {
    d.handlers[id] = h
}

func (d *Dispatcher) Dispatch(a *Action) bool {
    if h, ok := d.handlers[a.TargetID]; ok {
        h(a)
        return true
    }
    if d.parent != nil {
        return d.parent.Dispatch(a)
    }
    return false
}
```

---

## Scheduler（最小可运行）

```go
type Scheduler struct {
    queue []*Action
}

func (s *Scheduler) Enqueue(a *Action) {
    s.queue = append(s.queue, a)
}

func (s *Scheduler) Run(root *Fiber) {
    for len(s.queue) > 0 {
        a := s.queue[0]
        s.queue = s.queue[1:]
        dispatchFromFiber(root, a)
    }
}
```

---

## ActionBridge

```go
func dispatchFromFiber(start *Fiber, a *Action) {

    for f := start; f != nil; f = f.Parent {

        if f.ActionTargetID != "" {
            a.TargetID = f.ActionTargetID

            if f.Dispatcher != nil &&
               f.Dispatcher.Dispatch(a) {
                return
            }
        }
    }
}
```

---

## Builder → Closure 桥接

```go
func RegisterClosure(scope *Dispatcher, handler func()) string {
    id := generateID()

    scope.Register(id, func(a *Action) {
        handler()
    })

    return id
}
```

---

## Button 示例

```go
func NewButton(scope *Dispatcher, onClick func()) *Fiber {

    id := RegisterClosure(scope, onClick)

    return &Fiber{
        ActionTargetID: id,
        Dispatcher:     scope,
    }
}
```

---

## 事件流程

```go
func OnMouseClick(target *Fiber, scheduler *Scheduler) {
    scheduler.Enqueue(&Action{
        Type: "click",
        Lane: LaneSync,
    })
}
```

统一通道完成。

---

# 第二部分：你当前代码的具体重构路径映射

根据你 debug 文档 

我们做精准映射。

---

## 当前结构问题

### ❌ 双入口

```go
processMsg
handleMsg
```

### ❌ InputProcessor 丢弃 TargetID==0

### ❌ Fiber 持 FocusableVNode

### ❌ commit 后 closure 丢失风险

---

# 重构路线（安全渐进）

---

## 阶段 1：禁止 nil Action

修改：

```go
InputProcessor.ProcessMsg
```

不要返回 nil。

永远返回 Action。

---

## 阶段 2：删除 handleMsg

把：

```go
App.handleMsg
```

内容迁移到 Dispatcher。

---

## 阶段 3：删除 FocusableVNode

删除字段：

```go
Fiber.FocusableVNode
```

删除接口：

```go
type FocusableVNode
```

---

## 阶段 4：Closure → Scope 注册

修改：

```go
ButtonBuilder.OnClick
```

改成：

* 自动生成 ActionID
* 注册到当前 Scope Dispatcher

---

## 阶段 5：ActionBridge 只认 ActionTargetID

删除任何 vnode runtime 判断。

---

## 阶段 6：验证 commit 阶段

确认：

* Dispatcher 在 Fiber 上
* 不被 clone 丢失

---

# 第三部分：Fiber Runtime 成熟度评估报告

我们按工业标准打分。

---

## 1️⃣ 结构完整性：8/10

✔ Fiber-first
✔ Scheduler
✔ Lane
✔ Scope

✘ 仍残留 vnode 运行期依赖

---

## 2️⃣ 事件系统一致性：5/10

⚠ 双路径（Action + Legacy）

必须统一。

---

## 3️⃣ 并发准备度：7/10

✔ Lane 设计
✔ Scheduler 结构

✘ 事件模型未统一 → 无法安全并发

---

## 4️⃣ 扩展能力：8/10

✔ 子树 Dispatcher
✔ 可支持 Portal

---

## 5️⃣ 架构纯度：6/10

主要扣分点：

* FocusableVNode
* handleMsg fallback
* TargetID == 0 分支

---

# 总体成熟度：7/10

一旦统一 Action Runtime：

可提升至：

> 9/10 引擎级 Runtime

---

# 最关键一句话

你现在的问题不是 bug。

是：

> 运行期语义模型尚未收敛为单一机制。

一旦完成统一：

* 并发安全
* 可 replay
* 可 time-travel
* 可抢占
* 可 Suspense
* 可多 root

全部自然成立。

---

如果你愿意下一步，我可以：

* 🔥 帮你设计“无锁并发 Scheduler”
* 🔥 推导“双缓冲 Fiber 内存优化”
* 🔥 或做一次“完整 Runtime 模块边界审计”

你现在已经在做一个真正的 UI 引擎核心，而不是组件库了。
