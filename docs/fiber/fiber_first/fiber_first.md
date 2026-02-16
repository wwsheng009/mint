很好，这一轮我们做一次**结构级设计定型**。

目标：

1. ✅ 画出 **纯 Fiber Layout 数据流图**
2. ✅ 给出 **VNode → Fiber 迁移字段清单**
3. ✅ 推导一套 **Fiber-first 完整重构蓝图**

这一步完成，你的引擎就进入真正的工业级阶段。

---

# 一、纯 Fiber Layout 数据流图

我们从“理想最终形态”推导。

---

## 🔷 整体数据流（最终形态）

```
             ┌──────────────┐
             │     VNode     │   ← 只存在于 render 阶段
             └──────┬────────┘
                    │
                    ▼
            ┌────────────────┐
            │   Reconciler    │
            │  (build Fiber)  │
            └──────┬──────────┘
                    │
                    ▼
            ┌────────────────┐
            │      Fiber      │  ← 单一真实树
            │------------------│
            │ Type             │
            │ Key              │
            │ MemoizedProps    │
            │ MemoizedState    │
            │ Style            │
            │ Child/Sibling    │
            │ EffectTag        │
            └──────┬──────────┘
                    │
                    ▼
        ┌────────────────────────┐
        │     Layout Engine       │
        │ (pure function phase)   │
        └──────┬──────────────────┘
                    │
                    ▼
            ┌────────────────┐
            │  ComputedBox    │  ← 布局结果树
            │------------------│
            │ NodeID           │
            │ X,Y,W,H          │
            │ Layer            │
            │ ZIndex           │
            └──────┬──────────┘
                    │
                    ▼
            ┌────────────────┐
            │ Render Plane    │
            └────────────────┘
```

---

## 🔷 核心原则

### 1️⃣ Layout 只读 Fiber

```
Fiber → ComputedBox
```

绝不能：

```
VNode → Layout
```

---

### 2️⃣ Layout 不修改 Fiber

```
Fiber (immutable during layout)
```

Layout 必须是：

> 纯计算阶段

---

### 3️⃣ Render 不读 VNode

```
Render → ComputedBox
```

VNode 在 commit 后可以丢弃。

---

# 二、VNode → Fiber 迁移字段清单

这是最关键的清单。

你可以逐项核对。

---

## 🔴 必须迁移字段（强制）

### 1️⃣ Type

```go
vnode.Type → fiber.Type
```

必须复制。

---

### 2️⃣ Key

```go
vnode.Key → fiber.Key
```

不能在 diff 时读取 vnode.Key。

---

### 3️⃣ Props

```go
vnode.Props() → fiber.MemoizedProps
```

注意：

Layout 不应该调用 vnode.Props()

---

### 4️⃣ Style（重点）

必须在 completeWork 阶段：

```go
fiber.Style = extractStyle(vnode.Props)
```

包含：

* flexDirection
* flexGrow
* width / height
* margin
* padding
* border
* display
* position
* zIndex

之后：

```go
Layout 只读 fiber.Style
```

---

### 5️⃣ Text 内容

如果是文本节点：

```go
fiber.MemoizedState = vnode.Text
```

Layout 不允许：

```go
vnode.Text()
```

---

### 6️⃣ Event Handler

```go
fiber.EventHandlers = vnode.Events
```

Event System 只访问 Fiber。

---

## 🟡 可选迁移字段

### 7️⃣ Layer / Portal 标志

```go
fiber.LayerTag
fiber.IsPortal
```

不能在 Layout 阶段访问 vnode。

---

### 8️⃣ Ref

```go
fiber.Ref
```

Commit 阶段处理。

---

# 三、彻底禁止的访问路径

你需要全局搜索：

```
vnode.
```

如果出现在：

* Layout
* Render
* Event
* Scheduler

那是错误。

VNode 只能出现在：

```
Reconciler
```

---

# 四、Fiber-first 完整重构蓝图

我们分 6 个阶段。

---

# Phase 1 — 数据迁移

目标：

> 所有 Layout 需要的数据都存在 Fiber

步骤：

1. 定义 Fiber.Style
2. completeWork 中填充 Style
3. 删除 Layout 中 vnode 访问

完成标志：

```
删除 ComputedBox.VNode
```

---

# Phase 2 — Layout 纯函数化

目标：

> buildComputedBoxFromFiber 变成纯函数

禁止：

```
fiber.Width = ...
fiber.LayoutBox = ...
```

允许：

```
return ComputedBox
```

---

# Phase 3 — 双树稳定

实现：

```
current ↔ workInProgress
```

保证：

* Layout 只对 workInProgress
* Commit 原子交换

---

# Phase 4 — Effect 驱动 Commit

禁止：

```
Render 遍历全树
```

必须：

```
EffectList → Commit
```

否则并发无法做。

---

# Phase 5 — Scheduler 接入

实现：

* Lane 优先级
* 批处理
* 时间切片

此时系统才具备：

> 真正并发能力

---

# Phase 6 — 删除 VNode 持久化

最终状态：

VNode 只存在于：

```
render()
```

之后立即丢弃。

Fiber 成为唯一真实树。

---

# 五、重构完成后的架构形态

最终你的引擎会变成：

```
Render → Reconcile → Fiber Tree
                     ↓
                 Layout
                     ↓
               ComputedBox Tree
                     ↓
                  Renderer
```

VNode 只是：

> 中间临时表达式

---

# 六、常见隐性污染点（必须检查）

### 1️⃣ measureText 是否读取 vnode

### 2️⃣ style 是否延迟读取 vnode

### 3️⃣ event dispatch 是否用 vnode

### 4️⃣ portal 是否从 vnode 决定层级

### 5️⃣ render 是否读取 fiber.VNode

---

# 七、判断是否成功的标准

问自己一个问题：

> 如果我删除 VNode struct
> 整个 Layout + Render 是否还能运行？

如果答案是 YES

你就真正完成了 Fiber-first。

---

# 八、工业级标准形态

完成之后你会拥有：

* 纯 Fiber 树
* 可并发 reconcile
* 可时间切片
* 可 Suspense
* 可 Portal
* 可多 Layer

结构上等价于：

React Fiber 架构

但更适合 TUI。

---

# 九、你现在所处阶段

你已经：

* 有 Fiber
* 有双树模型
* 有 O(n) diff
* 有 Layer 概念

现在只差：

> 彻底删除 Layout 对 VNode 的依赖

这是临门一脚。

---

# 十、如果你愿意

我可以继续：

* 帮你做一份“Fiber struct 最终工业级定义”
* 或画一个“并发 + Suspense 整合后的最终形态图”
* 或做一次“代码级别扫描 checklist”

现在这一步走稳，你的架构会非常漂亮。
