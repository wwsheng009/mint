很好，这个问题比前面更关键。

你现在已经确认：

* 需要 **稳定 ID**
* HitTest 依赖 ID
* Instance 依赖 ID
* Event routing 依赖 ID
* Layer 会重排结构

那我们就必须把 identity 设计成**运行时一级结构**。

下面给你一个完整、可落地的方案。

---

# 🎯 目标

HitTest 需要：

```
(x, y) → 找到 Node → 找到 Instance → 触发事件
```

关键点：

> 命中结构必须依赖“稳定运行时 ID”，不能依赖 VNode.Key。

---

# 🚨 先说一个核心原则

> ❌ 不要用 VNode.Key 作为 HitTest ID
> ❌ 不要用 Fiber.Path 作为长期 ID
> ✔ 要用独立 Runtime NodeID

---

# 🧠 正确的 ID 分层模型

我们定义三种 identity：

```
1️⃣ DiffKey        → vnode.Key（只用于 sibling diff）
2️⃣ RuntimeNodeID  → Fiber.NodeID（全局唯一）
3️⃣ InstanceID     → Instance.NodeID（= Fiber.NodeID）
```

关系应该是：

```
VNode
  ↓ diff
Fiber (生成 RuntimeNodeID)
  ↓ layout
ComputedBox (保存 RuntimeNodeID)
  ↓ hittest
HitMap (用 RuntimeNodeID)
  ↓ lookup
InstanceMap[RuntimeNodeID]
```

VNode 不参与 runtime identity。

---

# 🧱 正确的 HitTest 架构

## 第一步：Fiber 创建时生成稳定 NodeID

在 createFiber 时：

```go
fiber.NodeID = idAllocator.Next()
```

⚠️ 不要用 Path。
⚠️ 不要用字符串拼接。
⚠️ 用 uint64 或 int64。

---

## 第二步：Layout 阶段把 NodeID 写入 ComputedBox

```go
type ComputedBox struct {
    Rect   Rect
    NodeID uint64
}
```

Layout 时：

```go
box.NodeID = fiber.NodeID
```

不要再通过 VNode.Key 读取。

---

## 第三步：构建 HitMap 只存 NodeID

```go
hitMap.Add(box.Rect, box.NodeID)
```

不要存 key 字符串。

---

## 第四步：HitTest 返回 NodeID

```go
func (h *HitMap) HitTest(x, y int) uint64
```

---

## 第五步：用 NodeID 找 Instance

```go
instance := instanceRegistry[nodeID]
```

---

# 💥 这样有什么好处？

### 1️⃣ Layer 剥离完全不影响 ID

StripLayers 再怎么 clone VNode，都不会影响 Fiber.NodeID。

---

### 2️⃣ Path 变化不会影响 ID

你将来：

* reorder
* insert
* animate
* portal
* overlay

都不会破坏 ID。

---

### 3️⃣ 不需要 cloneExistingFiber 同步 key

你可以直接删除：

```go
vnode.SetKey(fiber.Path)
```

整个问题消失。

---

# 🔥 进阶问题：如何让 NodeID 稳定？

有两种策略：

---

## 策略 A：纯 runtime ID（推荐）

每次 mount 生成新 ID：

```
mount → 分配 NodeID
unmount → 删除
```

优点：

* 简单
* 性能好
* 绝对唯一

缺点：

* 热更新后 ID 改变

但 TUI 一般不需要跨 session 稳定 ID。

---

## 策略 B：结构性稳定 ID（高级玩法）

NodeID = hash(parent.NodeID + diffKey + index)

优点：

* 可预测
* 可调试

缺点：

* reorder 会变
* 实现复杂

一般不建议。

---

# 🧩 那 Fiber.Path 还有什么用？

可以保留：

* 调试
* DevTools
* Inspector
* 错误堆栈

但不要再参与 identity。

---

# 🏗 推荐的最终结构

```go
type Fiber struct {
    NodeID   uint64  // 唯一运行时ID
    DiffKey  string  // vnode.Key
    Path     string  // 调试用
}
```

```go
type Instance struct {
    NodeID uint64
}
```

```go
type ComputedBox struct {
    NodeID uint64
}
```

```go
type HitEntry struct {
    Rect   Rect
    NodeID uint64
}
```

---

# 🧠 回答你一个隐含问题

你现在的架构其实已经是：

> React + Layout Engine + Event System

一旦有：

* HitMap
* Layer
* Instance
* Inspector

identity 就必须独立。

否则每增加一个系统，耦合会指数级增长。

---

# ⚠️ 你现在继续用 key 会怎样？

未来会出现：

* Modal 打开 → path 变 → ID 变 → 事件错乱
* Overlay 插入 → 全树 Path shift
* Reorder list → 全部 identity 改变
* Devtools 强制 rerender → key 漂移

这些是必然的。

---

# 🎯 终极建议（明确）

现在就做这件事：

```
1. 引入 Fiber.NodeID
2. ComputedBox 存 NodeID
3. HitMap 改为 NodeID
4. InstanceMap 改为 NodeID
5. 删除 vnode.SetKey(fiber.Path)
```

这是一次性架构修复。

越早越好。

---

# 🚀 如果你愿意

我可以帮你：

* 设计完整的 Runtime Identity 迁移方案
* 写出替换步骤
* 列出需要修改的所有模块
* 给你一个“无痛渐进迁移路径”

你现在这个 TUI，已经值得做成真正的 UI Runtime 了。

别再让 key 背锅了。

很好，这一步非常关键。

你现在的系统已经具备：

* Fiber 协调层
* Layout 引擎
* Layer 管理
* HitMap
* Instance Registry
* 事件路由

这已经是一个 **UI Runtime**，不是简单 TUI renderer 了。

所以我们要做的不是“修 bug”，而是：

> 把 identity 从 key/path hack 升级为正式 Runtime Identity 系统。

下面给你一个完整可执行的迁移方案。

---

# 🎯 一、最终目标架构（Target Architecture）

## 1️⃣ Identity 分层模型

```text
VNode
  └─ DiffKey (string)        ← 仅用于 sibling diff

Fiber
  ├─ NodeID (uint64)         ← 唯一运行时ID ⭐⭐⭐
  ├─ DiffKey (string)
  └─ Path (string, debug only)

ComputedBox
  └─ NodeID (uint64)

HitMap
  └─ NodeID (uint64)

InstanceRegistry
  └─ map[NodeID]Instance
```

核心原则：

> 运行时身份只认 NodeID，不认 key，不认 path。

---

# 🧠 二、NodeID 设计

## 1️⃣ 类型选择

```go
type NodeID uint64
```

理由：

* 比 string 快
* 比 path 稳定
* 比 hash 可控
* 足够大

---

## 2️⃣ 分配策略（推荐）

纯 runtime allocator：

```go
type IDAllocator struct {
    next uint64
}

func (a *IDAllocator) Next() NodeID {
    a.next++
    return NodeID(a.next)
}
```

在 Fiber mount 时分配。

不要 hash。
不要用 path。
不要结构生成。

---

# 🧱 三、需要修改的模块（完整列表）

按系统层次列出。

---

## 🔹 1️⃣ Fiber 层

文件：

* internal/reconciler/diff.go
* fiber.go

### 修改：

```go
type Fiber struct {
    NodeID  NodeID   // ⭐ 新增
    DiffKey string
    Path    string   // debug only
}
```

### 在 createChildFiber 时：

```go
fiber.NodeID = idAllocator.Next()
```

### 在 cloneExistingFiber 时：

保留：

```go
fiber.NodeID = current.NodeID
```

删除：

```go
vnode.SetKey(fiber.Path)  // ❌ 移除
```

---

## 🔹 2️⃣ Instance Registry

文件：

* framework/app.go
* internal/runtime/instance_registry.go

### 修改：

```go
map[NodeID]*Instance
```

Instance 结构：

```go
type Instance struct {
    NodeID NodeID
}
```

注册：

```go
registry[fiber.NodeID] = instance
```

删除：

```go
"vnode:" + key
```

---

## 🔹 3️⃣ Layout Engine

文件：

* internal/layout/*
* runtime/layout_engine.go

修改 ComputedBox：

```go
type ComputedBox struct {
    Rect   Rect
    NodeID NodeID
}
```

布局时：

```go
box.NodeID = fiber.NodeID
```

不要读取 VNode.Key。

---

## 🔹 4️⃣ HitMap

文件：

* internal/reconciler/reconciler.go
* internal/render/hitmap.go

修改：

```go
type HitEntry struct {
    Rect   Rect
    NodeID NodeID
}
```

HitTest：

```go
func (h *HitMap) HitTest(x, y int) NodeID
```

删除：

```go
box.VNode.Key()
```

---

## 🔹 5️⃣ Event Routing

文件：

* framework/app.go

修改：

```go
nodeID := hitMap.HitTest(x,y)
instance := instanceRegistry[nodeID]
```

删除：

```go
instanceKey := "vnode:" + nodeIDString
```

---

## 🔹 6️⃣ Layer System

文件：

* runtime/layer/collector.go

LayerNode 结构增加：

```go
type LayerNode struct {
    NodeID NodeID
}
```

walk 时：

```go
node.NodeID = fiber.NodeID
```

不要用 vnode.Key。

---

## 🔹 7️⃣ Inspector / DevTools

如果有 Inspector 依赖 Path：

保留 Path，但只用于显示。

不要用于查找。

---

# 🛠 四、无痛渐进迁移路径

重点来了。

不能一次性全改，否则会炸。

我们做 3 阶段迁移。

---

# 🟢 阶段 1：引入 NodeID（不删除旧逻辑）

### 步骤：

1️⃣ Fiber 增加 NodeID
2️⃣ ComputedBox 增加 NodeID
3️⃣ HitMap 同时存：

```go
NodeID
VNodeKey
```

4️⃣ InstanceRegistry 同时支持：

```go
map[string]Instance
map[NodeID]Instance
```

5️⃣ Event Routing 优先用 NodeID，如果没有 fallback 到旧 key。

系统仍然能跑。

---

# 🟡 阶段 2：双轨验证

加 debug 断言：

```go
if instanceByKey != instanceByNodeID {
    panic("identity mismatch")
}
```

运行完整测试：

* Modal
* Overlay
* Reorder
* Dynamic insert
* Focus change

验证无 mismatch。

---

# 🔴 阶段 3：删除旧 identity

删除：

* vnode.SetKey(fiber.Path)
* 所有 "vnode:" + key
* HitMap 读取 VNode.Key
* cloneExistingFiber 同步逻辑

彻底只保留 NodeID。

---

# 📦 五、迁移影响面评估

| 模块       | 风险 | 复杂度 |
| -------- | -- | --- |
| Fiber    | 低  | 低   |
| Layout   | 低  | 低   |
| HitMap   | 中  | 中   |
| Event    | 中  | 中   |
| Layer    | 中  | 中   |
| DevTools | 低  | 低   |

总体风险：可控。

---

# 🧠 六、迁移后的收益

### ✔ Layer 重排不再影响 identity

### ✔ Path 改变不影响事件

### ✔ Reorder 不破坏 Instance

### ✔ 不需要同步 key

### ✔ 不会再出现 key mismatch bug

---

# 🔥 七、迁移完成后的最终结构

```text
VNode (声明层)
    ↓
Fiber (协调层)
    ├─ NodeID ⭐
    └─ DiffKey
    ↓
Layout
    └─ ComputedBox(NodeID)
    ↓
HitMap
    └─ NodeID
    ↓
InstanceRegistry[NodeID]
    ↓
Event Dispatch
```

这是一套真正 UI Runtime 的 identity 模型。

---

# 🚀 八、未来扩展能力

迁移完成后，你可以轻松实现：

* Portal
* 多 Root
* Suspense
* 动画系统
* 离屏缓存
* 虚拟列表
* DevTools Time Travel

而不会再被 key 限制。

---

# 🎯 最终建议

你现在正处在“架构升级临界点”。

如果不做这次迁移：

未来每新增一个系统都会加倍复杂。

如果做了：

你的 TUI 就是一个真正的 UI Runtime。

---

如果你愿意，我可以：

* 给你画完整的“迁移前 vs 迁移后结构对比图”
* 或者帮你写一版 NodeID Patch 模板代码
* 或者帮你设计 Fiber/Runtime 的下一步演进路线

你这套已经很接近框架级别了，不要停在 hack 阶段。

