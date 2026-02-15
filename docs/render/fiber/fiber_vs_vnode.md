这个问题问得非常核心。

答案不是简单的“是”或“否”。

> ❌ Fiber 不需要复制 **所有** VNode 信息
> ✅ Fiber 只需要复制 **运行期必须长期存在的信息**

我们从架构层面推导一次。

---

# 一、先明确：VNode 和 Fiber 本质不同

### VNode 是什么？

* render() 的返回值
* 短生命周期
* 描述“下一帧想要什么”

它是：

> 描述结构的表达式

---

### Fiber 是什么？

* 持久化节点
* 参与调度
* 参与 diff
* 参与 layout
* 参与 commit

它是：

> 运行期实体

---

所以问题本质是：

> 哪些信息需要跨帧存在？

只有这些，才需要复制到 Fiber。

---

# 二、VNode 信息分类

我们把 VNode 信息分 4 类。

---

## 🟢 类别 A：结构信息（必须复制）

* type
* key
* props（经过归一化）
* children
* text

这些必须存在 Fiber 中。

否则：

* diff 做不了
* layout 做不了
* render 做不了

---

## 🟡 类别 B：运行期需要的信息（必须复制）

* style
* event handlers
* ref
* layer 标志
* portal 标志

因为：

> Layout / Event / Render 不能访问 VNode

这些必须进 Fiber。

---

## 🔵 类别 C：中间计算信息（不应该复制）

例如：

* JSX 生成时的临时数据
* dev-only metadata
* 源代码位置
* 编译期 helper

这些：

> 不应进入 Fiber

否则 Fiber 会变成垃圾桶。

---

## 🔴 类别 D：仅用于 diff 的瞬时信息

比如：

* 临时 props 比较数据
* diff 过程中的 map
* 旧 vnode 结构

这些只在 reconcile 阶段使用。

不能持久化。

---

# 三、真正的原则

> Fiber 只持有“跨帧稳定信息”

问自己一个问题：

> 这个字段在下一次 render 之前还需要吗？

如果答案是：

* 是 → 放 Fiber
* 否 → 不要放

---

# 四、一个非常常见的错误

很多人会做：

```go
fiber.VNode = vnode
```

然后所有数据都通过：

```go
fiber.VNode.Props()
```

读取。

这看起来简单。

但会导致：

* Layout 依赖 VNode
* 并发不安全
* Suspense 无法实现
* VNode 永远无法删除

这是假 Fiber。

---

# 五、工业级 Fiber 应该长什么样？

理想结构是：

```go
type Fiber struct {
    // identity
    Type      ElementType
    Key       string

    // tree
    Parent    *Fiber
    Child     *Fiber
    Sibling   *Fiber
    Alternate *Fiber

    // data
    MemoizedProps any
    PendingProps  any
    MemoizedState any

    // style
    Style     Style

    // effect
    Flags     Flags

    // scheduling
    Lanes     Lane
}
```

注意：

没有 VNode。

---

# 六、如果你全部复制 VNode 会怎样？

问题会非常明显：

### 1️⃣ 内存膨胀

VNode + Fiber 双份结构。

---

### 2️⃣ Layout 不纯

会忍不住读取 vnode。

---

### 3️⃣ 无法时间切片

因为 VNode 不是稳定结构。

---

### 4️⃣ Suspense 很难实现

VNode 是静态描述，不适合运行期挂起。

---

# 七、真正正确的模型

VNode 的使命：

```text
描述“下一帧”
```

Fiber 的使命：

```text
维持“当前运行状态”
```

VNode 是表达式。

Fiber 是实例。

---

# 八、给你一个判断公式

是否复制 = 是否参与以下任一阶段：

* diff
* layout
* commit
* event
* scheduling
* state

如果参与 → 复制
否则 → 不复制

---

# 九、和 React 的对比

成熟实现里：

VNode（Element）只包含：

```js
{
  type,
  key,
  props,
}
```

Fiber 会展开为：

* type
* key
* memoizedProps
* memoizedState
* lanes
* flags
* alternate
* effect list

VNode 不会被长期保存。

结构上等价于：

React 的 Fiber 模型。

---

# 十、你当前阶段建议

不要做：

* “把整个 vnode 存进 fiber”

要做：

* 明确 Fiber 的字段边界
* 逐个迁移 layout 所需字段
* 最终删除 fiber.VNode

---

# 十一、最重要的一句话

> Fiber 不是 VNode 的壳
> Fiber 是运行期实体

只复制“运行期必须长期存在”的信息。

---

如果你愿意，我可以：

* 帮你做一次 Fiber struct 精简设计
* 或帮你做“字段削减扫描表”
* 或推导一个“最小可运行 Fiber 定义”

现在你已经在做引擎级设计了，这一步非常关键。

很好，这一轮我们做**结构级收敛设计**。

目标：

1. ✅ 做一次 **Fiber struct 精简设计**
2. ✅ 输出 **字段削减扫描表**
3. ✅ 推导一个 **最小可运行 Fiber 定义（工业可扩展版）**

我们要的不是“像 React”，而是：

> 适合 TUI + Layer + Portal + O(n) diff + 并发演进 的 Fiber

---

# 一、Fiber 结构分层设计

先做一个原则：

> Fiber = 运行期实体
> 不是 VNode 的缓存

我们把字段分为 6 个模块。

---

## 1️⃣ Identity（身份层）

```go
Type ElementType
Key  string
```

作用：

* diff
* 复用判断
* reconcileSingleElement

必须保留。

---

## 2️⃣ Tree Structure（结构层）

```go
Parent    *Fiber
Child     *Fiber
Sibling   *Fiber
Alternate *Fiber
```

作用：

* 单链树遍历
* 双树模型
* commit 阶段交换

必须保留。

---

## 3️⃣ Data Layer（数据层）

```go
PendingProps  any
MemoizedProps any
MemoizedState any
```

作用：

* 更新前后对比
* bailout
* 状态存储

如果你还没做并发：

PendingProps 可以暂时保留。

---

## 4️⃣ Layout Data（布局输入层）

```go
Style Style
```

这里是关键。

Layout 只允许读取：

```go
fiber.Style
```

不允许读取 vnode。

Style 必须是归一化结构体：

```go
type Style struct {
    Width, Height int
    FlexGrow      int
    FlexDirection Direction
    Margin        BoxSpacing
    Padding       BoxSpacing
    Position      PositionType
    ZIndex        int
}
```

---

## 5️⃣ Effect Layer（提交层）

```go
Flags     Flags
SubtreeFlags Flags
NextEffect *Fiber
```

如果你还没有 effect list：

可以先只保留：

```go
Flags
```

但工业级必须有 effect list。

---

## 6️⃣ Scheduler Layer（调度层）

```go
Lanes      Lane
ChildLanes Lane
```

如果你还没做时间切片：

可以保留但不启用。

---

# 二、字段削减扫描表

现在我们做一张“是否保留”的扫描表。

---

| 字段            | 是否必须 | 原因           | 是否可删        |
| ------------- | ---- | ------------ | ----------- |
| VNode         | ❌    | 运行期不需要       | ✅ 必删        |
| NodeID        | 🟡   | 仅用于 render   | 可迁移到 commit |
| Style         | ✅    | layout 输入    | 不可删         |
| Parent        | ✅    | 树遍历          | 不可删         |
| Child         | ✅    | 树遍历          | 不可删         |
| Sibling       | ✅    | O(n) diff    | 不可删         |
| Alternate     | ✅    | 双树模型         | 不可删         |
| MemoizedProps | ✅    | diff 比较      | 不可删         |
| PendingProps  | 🟡   | 并发用          | 可阶段性保留      |
| MemoizedState | 🟡   | stateful 组件  | 看需求         |
| LayoutBox     | ❌    | 属于 layout 结果 | 应移出         |
| Width/Height  | ❌    | layout 结果    | 不应在 Fiber   |
| LayerIndex    | 🟡   | 可能用于 render  | 可迁移         |
| EffectList    | ✅    | commit 优化    | 工业级必须       |
| DevInfo       | ❌    | debug        | 不进生产        |

---

# 三、常见错误字段（必须删除）

### ❌ fiber.VNode

污染源。

---

### ❌ fiber.LayoutBox

Layout 结果不能存在 Fiber。

必须：

```text
Fiber → ComputedBox
```

而不是：

```text
Fiber + LayoutResult 混合
```

---

### ❌ fiber.Width / fiber.Height

布局结果属于：

ComputedBox。

---

### ❌ fiber.RenderNode

Render 层应该使用：

ComputedBox。

---

# 四、最小可运行 Fiber 定义（工业基础版）

这是你当前阶段推荐的定义：

```go
type Fiber struct {

    // -------------------------
    // Identity
    // -------------------------
    Type ElementType
    Key  string

    // -------------------------
    // Tree
    // -------------------------
    Parent    *Fiber
    Child     *Fiber
    Sibling   *Fiber
    Alternate *Fiber

    // -------------------------
    // Data
    // -------------------------
    PendingProps  any
    MemoizedProps any
    MemoizedState any

    // -------------------------
    // Layout Input
    // -------------------------
    Style Style

    // -------------------------
    // Effects
    // -------------------------
    Flags        Flags
    SubtreeFlags Flags
    NextEffect   *Fiber

    // -------------------------
    // Scheduling
    // -------------------------
    Lanes      Lane
    ChildLanes Lane
}
```

注意：

没有：

* VNode
* LayoutBox
* RenderNode
* Width/Height

这是最干净形态。

---

# 五、最小化版本（如果你还没做并发）

如果你还没做：

* 时间切片
* Suspense
* 优先级

可以压缩成：

```go
type Fiber struct {

    Type ElementType
    Key  string

    Parent    *Fiber
    Child     *Fiber
    Sibling   *Fiber
    Alternate *Fiber

    MemoizedProps any
    MemoizedState any

    Style Style

    Flags Flags
}
```

这是：

> 可运行 + 可扩展 + 可进化 的最小版本

---

# 六、结构边界最终图

```text
VNode（短生命周期）
        ↓
Reconciler
        ↓
Fiber（长期实体）
        ↓
Layout
        ↓
ComputedBox（布局结果）
        ↓
Renderer
```

VNode 不应参与：

* Layout
* Render
* Event
* Scheduler

---

# 七、判断是否精简成功

问自己：

1. Fiber 是否还引用 VNode？
2. Layout 是否只读 Fiber？
3. Fiber 是否只存“跨帧必要信息”？
4. Layout 结果是否完全独立？

全部 YES 才算完成。

---

# 八、一个关键建议

你现在不要急着：

* 加字段
* 做复杂优化

而是：

> 先删除字段

删除比增加更重要。

---

# 九、你现在的位置

你已经有：

* O(n) diff
* 双树模型
* Layer
* Portal

现在是：

> 做结构收敛

一旦 Fiber 干净：

并发、Suspense 都会自然长出来。

---

如果你愿意，我可以继续：

* 推导“Fiber 内存模型优化”
* 或画“最终并发架构全景图”
* 或做一次“当前实现风险扫描 checklist”

你现在已经在做引擎核心架构设计了。
这一步做对，你的系统会非常优雅。
