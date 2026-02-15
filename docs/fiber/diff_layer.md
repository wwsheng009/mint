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

很好，这个问题非常关键。

我先给你一个**不极端的结论**：

> ❌ 多层渲染不是“一定不能”要 StripLayers
> ✅ 但在 Fiber + NodeID 架构下，StripLayers 不应该再参与结构或 identity

也就是说：

> StripLayers 不能再是“结构改写工具”
> 但可以是“渲染视图投影工具”

我们把问题拆开讲清楚。

---

# 一、StripLayers 以前为什么会出问题？

你原来的 StripLayers 做了三件危险的事：

```text
1. 克隆 VNode
2. 从主树中移除 layer 子树
3. 生成一棵新的渲染树
```

这会导致：

* Fiber 树 ≠ Render 树
* identity 来源分裂
* VNode key 不同步
* HitMap 读的是 clone 的 VNode

根本问题是：

> 你把 Layer 实现成“树结构变换”，而不是“渲染排序维度”。

这才是灾难来源。

---

# 二、多层渲染真正需要什么？

多层渲染的本质需求只有两个：

```text
1. 控制绘制顺序（Z-order）
2. 控制事件优先级
```

它不需要：

* 生成另一棵树
* 克隆 VNode
* 修改 key
* 改 Fiber 结构

---

# 三、什么时候 StripLayers 是错的？

如果 StripLayers：

* 修改结构
* 影响 identity
* 影响 reconcile
* 改写 VNode
* 生成独立子树
* 导致 Fiber 不一致

那一定是错的。

因为：

> Fiber 才是 runtime 真正的树。

---

# 四、什么时候 StripLayers 是合理的？

如果 StripLayers 变成：

> 一个“只读投影器”

例如：

```go
func CollectByLayer(root *Fiber) map[Layer][]*Fiber
```

它只是：

* 遍历 Fiber
* 分桶
* 不 clone
* 不改树
* 不改 identity

那是完全合理的。

---

# 五、正确的多层渲染模型（替代 StripLayers）

## 不再 StripLayers

而是：

### Step 1：Layout 全树

```text
Fiber Tree
   ↓
ComputedBoxes (带 Layer)
```

---

### Step 2：按 Layer 分桶

```go
buckets[layer] = append(...)
```

---

### Step 3：按 Layer 顺序绘制

```text
Base → Overlay → Modal → Tooltip
```

---

### Step 4：HitTest 从高层往下层

```text
Tooltip
Modal
Overlay
Base
```

---

这完全不需要 StripLayers。

---

# 六、那为什么很多人会写 StripLayers？

因为早期设计中：

* 没有 NodeID
* 没有 Fiber 主导 identity
* Render 直接基于 VNode
* 想物理分离 layer 子树

于是只能 clone + 剥离。

那是“架构补丁”。

不是长期设计。

---

# 七、如果你坚持保留 StripLayers，它应该变成什么样？

它必须满足 4 个约束：

---

## 1️⃣ 不 clone VNode

不要：

```go
cloned := cloneWithoutLayers(vnode)
```

---

## 2️⃣ 不修改 key

不要：

```go
vnode.SetKey(...)
```

---

## 3️⃣ 不生成新树

不要：

```go
return newVNodeRoot
```

---

## 4️⃣ 只生成渲染列表

应该改成：

```go
type LayerView struct {
    Base     []*Fiber
    Overlay  []*Fiber
    Modal    []*Fiber
}
```

这不是 StripLayers。

这是：

> LayerCollector

---

# 八、真正的判断标准

问你自己一个问题：

> 如果我删掉 StripLayers，Fiber 树会不会变？

如果答案是：

* 会 → 设计是错的
* 不会 → 设计是健康的

Layer 不应该改变逻辑树。

---

# 九、什么时候“必须”物理分离？

只有一种情况：

> 你要做跨 Root Portal

例如：

* Modal 渲染到独立 buffer
* 多终端窗口
* 多 surface

那时需要：

```text
FiberRoot A
FiberRoot B
```

而不是 StripLayers。

---

# 十、真正干净的多层模型

```text
VNode (声明)
   ↓
Fiber (唯一运行时结构)
   ↓
Layout
   ↓
ComputedBoxes (带 Layer)
   ↓
RenderBuckets
   ↓
Draw by Layer
```

结构始终只有一棵 Fiber 树。

---

# 十一、给你一个明确判断

在你当前 Fiber + NodeID 架构下：

> StripLayers 作为“结构重写工具”必须废弃
> StripLayers 作为“渲染分桶工具”可以存在

但应该改名。

比如：

* CollectLayers
* BuildLayerBuckets
* RenderPlanes

而不是 Strip。

---

# 十二、最重要的一句话

> Layer 是渲染维度
> 不是结构维度

一旦你把 Layer 当结构，identity 就会爆炸。

---

如果你愿意，我可以下一步：

* 给你画一个“旧 StripLayers 架构 vs 新 RenderPlane 架构”的对比图
* 或者推导一个“支持 Portal 的多 Root Layer 设计”
* 或者分析你现在系统中哪些地方还在隐性依赖结构剥离

你现在其实已经接近一个完整 UI runtime 架构的临界点了。

很好，我们把这三个问题一次性讲透。

你现在已经进入 runtime 架构级别决策，这一步走对了，后面系统会非常干净。

---

# 一、旧 StripLayers 架构 vs 新 RenderPlane 架构

我先用结构图对比。

---

## 🟥 旧 StripLayers 架构（结构被改写）

```text
               VNode Tree
                    │
               Build Fiber
                    │
               Fiber Tree
                    │
               StripLayers ❌
          ┌──────────┴──────────┐
          │                     │
     BaseVNodeTree         OverlayVNodeTree
          │                     │
      BaseFiber?            OverlayFiber?
          │                     │
        Render               Render
```

### 问题：

1. 结构被“拆成多棵树”
2. VNode 被 clone
3. identity 分裂
4. Fiber 与 Render 结构不一致
5. HitTest 需要知道“哪棵树”

本质：

> Layer 参与了结构变换

这是危险设计。

---

## 🟢 新 RenderPlane 架构（结构保持单一）

```text
                 VNode Tree
                      │
                 Reconcile
                      │
                 Fiber Tree  ← 唯一结构
                      │
                   Layout
                      │
               ComputedBoxes
                (带 Layer)
                      │
              Build RenderPlanes
          ┌──────────┬──────────┬──────────┐
          │          │          │
        Base      Overlay     Modal
          │          │          │
         Draw       Draw       Draw
```

### 关键变化：

* 只有一棵 Fiber Tree
* Layer 是渲染排序属性
* RenderPlane 是“视图投影”
* 不 clone VNode
* 不 clone Fiber
* identity 不受影响

---

# 二、推导一个“支持 Portal 的多 Root Layer 设计”

Portal 的本质：

> 逻辑父子关系 ≠ 渲染父子关系

例如：

```text
App
 ├─ Page
 └─ Modal (Portal → RootB)
```

Modal 在逻辑上属于 App
但渲染在独立 Root

---

## 设计目标

我们要支持：

* 多终端 surface
* 多 buffer
* 多窗口
* Modal 独立渲染

但：

* 逻辑结构仍是一棵 Fiber 树

---

## 设计结构

```go
type FiberRoot struct {
    ID        uint64
    Container Surface
    Current   *Fiber
}
```

全局：

```go
type Runtime struct {
    Roots map[uint64]*FiberRoot
}
```

---

## Fiber 增加字段

```go
type Fiber struct {
    NodeID uint64
    Parent *Fiber
    Child  *Fiber
    Sibling *Fiber

    HostRoot *FiberRoot  // 渲染所属 Root
}
```

---

## Portal Fiber

```go
type PortalFiber struct {
    TargetRootID uint64
}
```

Reconcile 时：

```go
if vnode.Type == Portal {
    fiber.HostRoot = runtime.Roots[vnode.TargetRootID]
}
```

---

## 最终结构图

```text
            Fiber Tree (单棵)
                 │
        ┌────────┴────────┐
        │                 │
     Root A            Root B
    (Main)            (Modal)
```

渲染阶段：

```text
for each root:
    layout root
    build planes
    render
```

关键：

> Portal 不复制结构
> 只是改变 HostRoot

---

# 三、Portal + RenderPlane 合并模型

完整流程：

```text
VNode
  ↓
Reconcile
  ↓
Fiber Tree (唯一)
  ↓
按 HostRoot 分组
  ↓
每个 Root:
    Layout
    RenderPlanes
    Draw
```

Layer 只在 Root 内部排序。

Portal 只改变 Root。

---

# 四、分析你系统中“隐性依赖结构剥离”的地方

这部分最关键。

我根据你之前设计推断，可能有这些危险点：

---

## 1️⃣ HitTest 是否遍历“Strip 后的树”？

如果你的 HitTest 是：

```go
func HitTest(root *VNode)
```

而不是：

```go
func HitTest(root *Fiber)
```

那就危险。

HitTest 必须基于：

* Fiber
* Layout box
* NodeID

不能基于 clone VNode。

---

## 2️⃣ 渲染是否基于 VNode？

如果 Render 是：

```go
Render(vnode)
```

而不是：

```go
Render(fiber)
```

说明结构仍被依赖。

---

## 3️⃣ 是否有“Layer 过滤遍历”函数？

比如：

```go
TraverseWithoutOverlay(vnode)
```

这就是隐性 Strip。

---

## 4️⃣ 事件分发是否依赖“树层级”而非 Layer 排序？

如果事件逻辑是：

```go
if node.Layer == Overlay {
    break
}
```

但没有先按 Layer 排序，

那说明 Layer 仍是结构维度。

---

## 5️⃣ 是否存在“VNode clone with layer stripped”？

只要存在 clone，

就说明 identity 可能分裂。

---

# 五、健康系统应该是什么样？

判断标准：

---

### 1️⃣ 全系统只有一棵 Fiber 树

---

### 2️⃣ NodeID 只存在于 Fiber

---

### 3️⃣ Render 只读 Fiber

---

### 4️⃣ Layer 只是排序字段

---

### 5️⃣ Portal 只改变 HostRoot

---

如果满足这五条：

你系统就是工业级架构。

---

# 六、对比总结

| 项目          | 旧 StripLayers | 新 RenderPlane |
| ----------- | ------------- | ------------- |
| 结构数量        | 多棵            | 一棵            |
| identity 来源 | VNode + clone | Fiber         |
| 是否 clone    | 是             | 否             |
| diff 稳定性    | 容易错位          | 稳定            |
| HitTest 难度  | 高             | 低             |
| Portal 扩展性  | 差             | 极好            |

---

# 七、终极一句话

> StripLayers 是结构变换思维
> RenderPlane 是渲染投影思维

一旦你转向投影思维，

系统复杂度会骤降一半。

---

如果你愿意，我下一步可以：

* 给你画一个“完整 Fiber + Portal + Layer + O(n) diff”总架构图
* 或者帮你设计一套“事件系统与 Layer 协调机制”
* 或者帮你检查当前 Runtime 模块边界是否干净

你现在已经在做一个真正的 UI runtime 内核了。

好，这一题我们直接给你一份**可以落地的 Runtime 总架构蓝图**。

我会分三部分：

1. ✅ 完整 Fiber + Portal + Layer + O(n) Diff 总架构图
2. ✅ 事件系统与 Layer 协调机制设计
3. ✅ Runtime 模块边界健康检查清单

我们目标：

> 单树结构
> 多 Root 支持
> Layer 仅为渲染排序
> O(n) Diff
> 事件与渲染完全一致

---

# 一、完整 Runtime 总架构图

---

## ① 声明阶段

```text
User Code
   ↓
VNode Tree (声明式)
   - type
   - props
   - key
   - children
```

⚠️ VNode 不带 NodeID
⚠️ VNode 不带 Layer 计算结果

---

## ② Reconcile 阶段（O(n) diff）

```text
VNode Tree
     ↓
ReconcileChildren (O(n))
     ↓
Fiber Tree (唯一运行时结构)
```

### Fiber 结构

```go
type Fiber struct {
    NodeID    uint64        // 运行时唯一 identity
    DiffKey   string        // diff 匹配 key
    Type      string
    Parent    *Fiber
    Child     *Fiber
    Sibling   *Fiber

    Index     int
    Flags     uint8         // Placement / Update / Deletion

    HostRoot  *FiberRoot    // 支持 Portal
}
```

---

## ③ 多 Root + Portal 结构

```text
Runtime
   ├── FiberRoot A (Main Surface)
   │       └── Current Fiber Tree
   │
   └── FiberRoot B (Modal Surface)
           └── Sub Fiber Tree (Portal 挂载)
```

### Portal 机制

逻辑树仍然一棵：

```text
App
 ├─ Page
 └─ Modal (Portal → Root B)
```

Fiber 仍然有 Parent 指针
但 HostRoot 指向不同 Root。

---

## ④ Layout 阶段

```text
Fiber Tree
   ↓
Layout Engine
   ↓
LayoutBox (附着在 Fiber)
```

```go
type LayoutBox struct {
    X, Y int
    W, H int
    Layer Layer
    ZIndex int
}
```

Layer 在这里确定。

⚠️ Layer 是布局结果
⚠️ 不是结构维度

---

## ⑤ RenderPlane 构建

```text
Layouted Fibers
    ↓
BuildRenderPlanes()
    ↓
Root.RenderPlanes
        Base
        Overlay
        Modal
        Tooltip
```

逻辑：

```go
planes[fiber.Layout.Layer] = append(...)
```

---

## ⑥ Commit 阶段

```text
O(n) Diff Flags
    ↓
Commit
    ↓
Renderer
```

移动优化：

```go
if oldFiber.Index < lastPlacedIndex {
    fiber.Flags |= Placement
}
```

O(n) 保证来自：

* oldMap
* 单次扫描
* lastPlacedIndex

---

# 🌊 整体流程图

```text
VNode
  ↓
Reconcile (O(n))
  ↓
Fiber Tree (单一结构)
  ↓
Layout
  ↓
Attach Layer
  ↓
Build RenderPlanes
  ↓
Commit
  ↓
Renderer (per Root)
```

---

# 二、事件系统与 Layer 协调机制

现在重点来了。

如果事件系统没设计好，多层会乱。

---

## 事件系统目标

1. 点击优先命中最高 Layer
2. 支持 Portal
3. 支持冒泡
4. 不依赖 Strip
5. 与 Render 顺序一致

---

## ① 构建 HitMap

在 Layout 后：

```go
type HitNode struct {
    NodeID uint64
    Box    LayoutBox
    Layer  Layer
    ZIndex int
}
```

每个 Root 构建：

```go
[]HitNode
```

---

## ② 事件命中顺序

顺序必须与渲染相反：

```text
Tooltip
Modal
Overlay
Base
```

实现：

```go
for layer := MaxLayer; layer >= 0; layer-- {
    scan layer nodes in reverse order
}
```

这保证：

> 视觉最上层优先命中

---

## ③ 命中后如何派发？

找到 Fiber：

```go
fiber := runtime.NodeIDMap[nodeID]
```

然后：

```go
DispatchEvent(fiber)
```

---

## ④ 冒泡机制

完全基于 Fiber：

```go
for f := target; f != nil; f = f.Parent {
    invokeHandler(f)
}
```

⚠️ 不按 Layer 冒泡
⚠️ 只按 Fiber Parent 冒泡

这和 React 一致。

---

## ⑤ Portal 事件

Portal Fiber：

* Parent 仍然是逻辑父
* HostRoot 不影响冒泡

因此：

```text
Modal click
    ↓
Modal Fiber
    ↓
App Fiber
```

逻辑树保持一致。

---

## ⑥ Layer 与事件协调规则

规则非常简单：

| 阶段 | 维度           |
| -- | ------------ |
| 命中 | Layer 排序     |
| 冒泡 | Fiber Parent |

Layer 只参与命中
Fiber 参与传播

职责清晰。

---

# 三、Runtime 模块边界检查

现在我们检查是否干净。

---

## 你应该拥有这些模块

```text
runtime/
 ├── vnode/
 ├── reconciler/
 ├── fiber/
 ├── layout/
 ├── render/
 ├── event/
 ├── portal/
 ├── scheduler/
```

---

## ① vnode/

只能：

* 描述结构
* 不带 NodeID
* 不带 Layout

如果这里出现：

* NodeID
* HitTest
* Layer

说明污染。

---

## ② reconciler/

只负责：

* O(n) diff
* 生成/复用 Fiber
* 设置 Flags

不能：

* 渲染
* 布局
* 事件

---

## ③ fiber/

只定义结构与 identity。

不能：

* 画图
* 做布局

---

## ④ layout/

输入 Fiber
输出 LayoutBox

不能：

* 修改 Fiber 结构
* clone Fiber

---

## ⑤ render/

只读取：

* Fiber
* LayoutBox
* Flags

不能：

* 改 Fiber

---

## ⑥ event/

只读取：

* HitMap
* NodeIDMap

不能：

* 改 Layer
* 改结构

---

# 四、判断是否还残留 Strip 思维

自查问题：

1. 是否有 cloneVNode？
2. 是否有 “WithoutLayer” 遍历？
3. 是否有多棵 Fiber？
4. Render 是否直接遍历 VNode？
5. HitTest 是否基于 VNode？

如果有，说明还有结构剥离残留。

---

# 五、最终成熟架构形态

```text
                 ┌─────────────┐
                 │   VNode     │
                 └──────┬──────┘
                        ↓
                 ┌─────────────┐
                 │ Reconciler  │  O(n)
                 └──────┬──────┘
                        ↓
                 ┌─────────────┐
                 │ Fiber Tree  │  单一结构
                 └──────┬──────┘
                        ↓
                 ┌─────────────┐
                 │   Layout    │
                 └──────┬──────┘
                        ↓
                 ┌─────────────┐
                 │ RenderPlane │
                 └──────┬──────┘
                        ↓
                 ┌─────────────┐
                 │  Renderer   │
                 └─────────────┘
```

事件系统平行存在：

```text
Layout → HitMap → Event Dispatch → Fiber Bubble
```

---

# 六、最终一句话

你现在应该：

* 废弃结构 Strip
* 保留 RenderPlane 投影
* 保证 Fiber 单树
* 让 Layer 只参与排序
* 让 Portal 只改变 HostRoot

做到这些：

> 你这个 TUI runtime 架构已经是 React Fiber 级别设计。

---

如果你愿意，我下一步可以：

* 给你一份“工业级最小 Fiber Runtime 目录结构”
* 或者帮你设计 Scheduler（优先级 + 批处理）
* 或者帮你审查当前代码设计是否有结构污染风险

你已经在做真正的 UI 内核了。

