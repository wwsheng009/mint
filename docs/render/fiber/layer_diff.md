很好，现在我们进入真正的 Runtime 设计阶段。

你已经有：

* Fiber Core（NodeID 独立）
* Layout
* HitMap
* Instance

现在要做的是：

> 在不破坏 identity 的前提下，引入 Layer（Modal / Overlay / Tooltip / Inspector）

目标：

* Layer 不影响 diff
* Layer 不破坏 NodeID
* Layer 不污染 VNode
* Layer 不依赖 Path hack
* HitTest 支持 Z-order

我们从最小模型开始推导。

---

# 🎯 一、Layer 的本质

Layer 不是“树结构改变”。

Layer 是：

> 同一棵 Fiber 树的不同“渲染平面（render plane）”

也就是说：

```text
Fiber 树 = 逻辑结构
Layer     = 渲染顺序分组
```

不要：

* 把 Layer 从 Fiber 树剥离
* 不要 clone VNode
* 不要 StripLayers

那是旧模型的问题根源。

---

# 🧠 二、最小 Layer 模型设计

---

## 1️⃣ Layer 枚举

```go
type Layer int

const (
    LayerBase Layer = iota
    LayerModal
    LayerOverlay
    LayerTooltip
    LayerInspector
)
```

---

## 2️⃣ VNode 增加 Layer 声明（声明层）

```go
type VNode struct {
    Type     string
    Key      string
    Layer    Layer
    Props    Props
    Children []VNode
}
```

默认：

```go
LayerBase
```

VNode 只是声明在哪个 Layer。

---

## 3️⃣ Fiber 保存 Layer

```go
type Fiber struct {
    NodeID  uint64
    DiffKey string
    Type    string
    Layer   Layer

    Parent   *Fiber
    Child    *Fiber
    Sibling  *Fiber

    Props Props
}
```

Layer 在 reconcile 时从 VNode 拷贝。

---

# 🔄 三、Reconcile 不需要特殊处理 Layer

关键原则：

> Layer 不参与 diff 规则
> 只参与渲染阶段排序

reconcile 时：

```go
fiber.Layer = vnode.Layer
```

就够了。

不要：

* 根据 Layer 改 path
* 根据 Layer 改 identity

---

# 🧱 四、Layout 阶段支持 Layer

Layout 仍然遍历整棵 Fiber 树：

```go
func Layout(root *Fiber)
```

每个 Fiber 生成一个 ComputedBox：

```go
type ComputedBox struct {
    NodeID uint64
    Rect   Rect
    Layer  Layer
}
```

Layout 时：

```go
box.Layer = fiber.Layer
```

---

# 🎯 五、Render 阶段按 Layer 排序

关键来了。

不要再 StripLayers。

而是：

> Layout 生成所有 ComputedBox
> Render 阶段按 Layer 分桶

---

## RenderBuckets

```go
type RenderBuckets struct {
    Buckets map[Layer][]*ComputedBox
}
```

遍历所有 boxes：

```go
for _, box := range allBoxes {
    buckets[box.Layer] = append(buckets[box.Layer], box)
}
```

---

## 渲染顺序

```text
LayerBase
LayerOverlay
LayerModal
LayerTooltip
LayerInspector
```

按顺序绘制。

---

# 🧠 六、HitTest 支持 Layer

HitTest 必须支持 Z-order。

核心规则：

> 从最高 Layer 开始检测

---

## 修改 HitMap

```go
type HitEntry struct {
    Rect   Rect
    NodeID uint64
    Layer  Layer
}
```

---

## HitTest 逻辑

```go
func (h *HitMap) HitTest(x, y int) uint64 {
    for layer := HighestLayer; layer >= LayerBase; layer-- {
        for _, entry := range h.entries[layer] {
            if entry.Rect.Contains(x,y) {
                return entry.NodeID
            }
        }
    }
    return 0
}
```

Layer 只是优先级控制。

---

# 🔥 七、Modal 阻断点击

如果是 Modal：

规则：

> 一旦命中 Modal 层
> 不再往下层检测

实现方式：

HitTest 一旦命中就 return。

不用特殊逻辑。

---

# 🧩 八、Layer 不改变 Fiber 结构

这点极其重要。

Fiber 树仍然是：

```text
Root
 ├─ App
 │   ├─ Button
 │   └─ Modal (LayerModal)
 │       └─ ModalContent
```

Layer 只是标记。

不需要：

* 生成独立树
* 从 Base 树中移除
* 克隆 VNode

---

# 🚀 九、为什么这样比 StripLayers 更优？

旧模型问题：

* identity 不一致
* vnode clone
* key 不同步
* layer 重排影响 path

新模型：

* NodeID 不变
* Fiber 不变
* VNode 不被修改
* Layout 不拆树
* Layer 只是渲染分组

结构稳定。

---

# 🧠 十、Portal / 多 Root 支持（自然扩展）

如果将来支持 Portal：

只需：

```go
fiber.Layer = LayerModal
fiber.RenderRoot = ModalRoot
```

无需改 diff。

---

# 📦 十一、完整流程图（带 Layer）

```text
Render()
  ↓
VNode Tree (带 Layer)
  ↓
Reconcile
  ↓
Fiber Tree (NodeID + Layer)
  ↓
Layout
  ↓
ComputedBoxes(NodeID + Layer)
  ↓
RenderBuckets
  ↓
按 Layer 顺序绘制
  ↓
HitMap(NodeID + Layer)
  ↓
Event(NodeID)
```

---

# 🎯 十二、最小 Fiber Core + Layer 最终结构

```go
type Fiber struct {
    NodeID  uint64
    DiffKey string
    Type    string
    Layer   Layer

    Parent   *Fiber
    Child    *Fiber
    Sibling  *Fiber

    Props Props
}
```

```go
type ComputedBox struct {
    NodeID uint64
    Rect   Rect
    Layer  Layer
}
```

```go
type HitEntry struct {
    NodeID uint64
    Rect   Rect
    Layer  Layer
}
```

没有：

* Path
* Key hack
* StripLayers
* vnode clone
* instance key string 拼接

---

# 🏁 十三、一句话总结

> Layer 是渲染排序维度
> NodeID 是身份维度
> DiffKey 是匹配维度

三者完全解耦。

---

# 🔥 你现在得到的是：

一个：

* identity 正交
* layer 稳定
* diff 简洁
* hittest 清晰
* 可扩展到 Portal / Suspense

的 Fiber Runtime 核心。

---

如果你愿意，下一步我可以：

* 推导“支持 Portal 的 Layer Fiber”
* 或者设计“支持局部重排的高性能 reconcile”
* 或者帮你把这个模型对照 React Fiber 做一次深度对比

你现在已经走到真正框架级设计阶段了。
