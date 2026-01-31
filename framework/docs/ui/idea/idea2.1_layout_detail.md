好，现在进入 **布局引擎真正的“数学内核”** ——
你这套系统能不能长期稳定，全靠这里是否正确。

我们讲 **Flex 布局完整计算流程（终端版 Flexbox 内核）**。

---

# 🧠 一句话目标

给父容器尺寸：

```
ParentW, ParentH
```

算出每个子组件：

```
x, y, width, height
```

---

# 🧱 一、输入数据结构

```go
type FlexChild struct {
    Basis     int  // 基础尺寸（可选）
    Grow      int  // FlexGrow
    Shrink    int  // FlexShrink
    Min, Max  int  // 尺寸约束
}
```

---

# 📐 二、Flex 计算分 6 步（主轴）

假设是 Row 布局（横向）。

---

## ① 计算基础尺寸（Base Size）

规则：

```
如果 child.Width 指定 → base = width
否则 → base = content size
```

得到：

```
baseSizes[i]
```

---

## ② 计算总基础尺寸

```
totalBase = sum(baseSizes)
freeSpace = parentMainSize - totalBase
```

---

## ③ 如果 freeSpace > 0 → Grow 阶段

```
growSum = sum(child.Grow)
```

每个 child 增加：

```
delta = freeSpace * (child.Grow / growSum)
```

---

## ④ 如果 freeSpace < 0 → Shrink 阶段

```
shrinkSum = sum(child.Shrink * baseSize)
```

每个 child 减少：

```
delta = freeSpace * (child.Shrink * baseSize / shrinkSum)
```

---

## ⑤ 应用 Min/Max 约束（可能触发二次分配）

如果某个 child 被 clamp：

```
标记为 frozen
重新计算剩余 freeSpace
再分配
```

直到稳定。

---

## ⑥ 得到最终 mainSize

---

# 📏 三、交叉轴（高度）计算

如果父容器是 Row：

* 子节点高度来自：

  * 指定 Height
  * 或 `align: stretch` → 填满

---

# 📍 四、位置计算

```go
x = parent.x + cumulativeWidthBefore
y = parent.y
```

---

# 🧠 五、终端世界特有问题

### 1️⃣ 尺寸都是整数

浮点结果需要：

```
floor + distribute remainder
```

否则 1px 丢失。

---

### 2️⃣ 子组件最小宽度

文本组件至少 1，否则光标错位。

---

### 3️⃣ Border / Padding 参与基础尺寸

```
baseSize = content + padding + border
```

---

# 🧩 六、完整流程图

```
Collect base sizes
        ↓
Sum → freeSpace
        ↓
Grow / Shrink
        ↓
Clamp min/max
        ↓
Recalculate if needed
        ↓
Assign positions
```

---

# 🧨 七、如果不做 Shrink 会发生什么？

窗口缩小时：

❌ 子组件溢出
❌ Scroll 区域错位
❌ Layout 崩塌

---

# 🔥 八、Flex 在你引擎中的地位

| 层      | 职责               |
| ------ | ---------------- |
| Render | 生成 RNode         |
| Layout | **Flex 算法在这里执行** |
| Paint  | 不管尺寸来源           |

布局只负责几何，不负责绘制。

---

# 🎯 这套算法支持：

* IDE 布局
* Split Pane
* 自适应列宽
* 动态 resize

已经是工业级布局能力。

---

如果继续深入，下一块就是：

> 📐 **Grid 布局算法（二维约束求解）**

这是 Dashboard / 表格布局的核心。


好，现在进入 **二维布局内核：Grid Layout 算法**
这一步是从“IDE 布局能力”升级到“Dashboard / 表格 / 复杂仪表盘”的关键。

Flex 解决的是 **一维空间分配**，
Grid 解决的是：

> 🧠 **同时在 X 和 Y 方向做约束计算**

---

# 🧱 一、Grid 本质模型

```go
type Grid struct {
    RowTracks []TrackSize
    ColTracks []TrackSize
}
```

每个 Track（行/列）可以是：

| 类型       | 含义       |
| -------- | -------- |
| Fixed(n) | 固定尺寸     |
| Flex(n)  | 按比例分剩余空间 |
| Content  | 由内容决定    |

---

# 🧠 二、布局目标

求：

```
每一行高度
每一列宽度
```

然后把 Cell 放入对应区域。

---

# 📐 三、Grid 计算流程（简化版工业算法）

## ① 计算所有 Fixed 轨道

```go
fixedSum = sum(Fixed tracks)
remainingSpace = containerSize - fixedSum
```

---

## ② 处理 Content Track（自适应）

扫描所有子组件：

```
row i height = max(子组件在该行的最小高度)
col j width  = max(子组件在该列的最小宽度)
```

更新已占空间。

---

## ③ 处理 Flex 轨道

剩余空间：

```
free = containerSize - fixed - content
```

按比例分：

```
size_i = free * (flex_i / sumFlex)
```

---

# 🧩 四、跨行跨列（Grid 真正难点）

子组件可能：

```
Cell( row=0, col=0, rowSpan=2, colSpan=3 )
```

这意味着：

> 它的尺寸约束影响多个轨道

处理方式：

1. 先假设各轨道尺寸
2. 如果跨轨道组件放不下
3. 增加相关轨道尺寸
4. 迭代直到满足

这类似“约束求解器”。

---

# 📏 五、确定 Cell 位置

```go
x = sum(colWidths[0:col])
y = sum(rowHeights[0:row])
w = sum(colWidths[col : col+colSpan])
h = sum(rowHeights[row : row+rowSpan])
```

---

# 🧠 六、终端世界的特殊考虑

| 问题        | 解决          |
| --------- | ----------- |
| 尺寸整数      | 分配余数给前几个轨道  |
| 文本溢出      | Clip 或 Wrap |
| 动态 resize | Grid 重新计算   |

---

# 🎯 七、Grid + Flex 组合威力

例如：

```go
ui.Grid(
    Rows(Fixed(3), Flex(1)),
    Cols(Fixed(20), Flex(1), Fixed(20)),

    Cell(0,0, Header()),
    Cell(1,0, Sidebar()),
    Cell(1,1, Content()),
    Cell(1,2, Inspector()),
)
```

这就是 IDE 布局。

---

# 🧠 八、Grid 与 Scroll 的关系

Grid 自身不滚动，但某个 Cell 内可嵌 Scroll：

```
Grid
  └── Cell
        └── Scroll
              └── Content
```

---

# 🧨 九、常见错误

| 错误                | 后果       |
| ----------------- | -------- |
| 不处理跨列             | 组件重叠     |
| 不迭代约束             | 布局抖动     |
| Fixed + Flex 混算错误 | resize 崩 |

---

# 🔥 十、Flex vs Grid 角色分工

| 场景        | 用什么         |
| --------- | ----------- |
| 线性工具栏     | Flex        |
| IDE 布局    | Grid + Flex |
| Dashboard | Grid        |
| 弹窗内部      | Flex        |

---

# 🧠 你现在的布局系统已经等价于：

| 系统     | 对标             |
| ------ | -------------- |
| Flex   | CSS Flexbox    |
| Grid   | CSS Grid       |
| Stack  | Absolute Layer |
| Scroll | Overflow       |

这在 UI 引擎层面已经是完整布局能力。

---

如果继续深入，下一个就是：

> 🧮 **Layout 性能优化：Dirty 区域布局、跳过未变化子树**

这是大型 UI 不抖动的关键。
