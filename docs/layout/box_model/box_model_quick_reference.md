# Box Model 快速参考

## 核心概念

### 盒模型组成

```
┌──────────────────────────────────────────┐
│           Margin (外边距)                │  ← 节点与兄弟节点的间距
│  ┌──────────────────────────────────┐   │
│  │        Border (边框)             │   │  ← 视觉边界，占用 1 字符
│  │  ┌────────────────────────────┐  │   │
│  │  │     Padding (内边距)       │  │   │  ← 内容与边框的间距
│  │  │  ┌──────────────────────┐  │  │   │
│  │  │  │   Content (内容)     │  │  │   │
│  │  │  │                      │  │  │   │
│  │  │  └──────────────────────┘  │  │   │
│  │  └────────────────────────────┘  │   │
│  └──────────────────────────────────┘   │
└──────────────────────────────────────────┘
```

### 测量和布局的职责

| 阶段 | 处理的属性 | 不处理的属性 |
|------|-----------|-------------|
| **测量** | 内容尺寸、padding、border | margin |
| **布局** | 位置、margin | - |

---

## 约束传播规则

### 父节点 → 子节点

```
父节点约束 (Constraints)
    ↓
扣除父节点的 padding 和 border
    ↓
子节点约束 (InnerConstraints)
```

**公式：**
```
innerWidth  = maxWidth  - paddingLR - borderLR
innerHeight = maxHeight - paddingTB - borderTB

innerConstraints = Constraints{
    MinWidth:  max(0, minWidth  - paddingLR - borderLR),
    MaxWidth:  max(0, maxWidth  - paddingLR - borderLR),
    MinHeight: max(0, minHeight - paddingTB - borderTB),
    MaxHeight: max(0, maxHeight - paddingTB - borderTB),
}
```

### 子节点 → 父节点

```
子节点测量 (ContentSize)
    ↓
加上子节点的 padding 和 border
    ↓
父节点使用尺寸 (TotalSize)
```

**公式：**
```
totalWidth  = contentWidth  + paddingLR + borderLR
totalHeight = contentHeight + paddingTB + borderTB
```

---

## 布局位置计算

### 坐标系统

```
容器 (0, 0, W, H)
    │ Padding: [pt, pr, pb, pl]
    │ Border: Single (1 char)
    │
    ├─ 边框边界: (0, 0) 到 (W, H)
    ├─ 内容区域起点: (pl+1, pt+1)
    ├─ 内容区域尺寸: (W-pl-pr-2, H-pt-pb-2)
    └─ 偏移: OffsetX = pl + 1, OffsetY = pt + 1
```

### 子节点位置

```
child.X = container.X + internal.X + OffsetX
child.Y = container.Y + internal.Y + OffsetY
```

**其中 `internal.X/Y` 是 FlexLayout 返回的相对位置**

---

## 关键接口

### BoxModel（提议的新接口）

```go
type BoxModel struct {
    Margin  Margin
    Padding Padding
    Border  Border
}

// 辅助方法
func (b BoxModel) HorizontalPadding() int    // 左右 padding + 边框
func (b BoxModel) VerticalPadding() int      // 上下 padding + 边框
func (b BoxModel) ContentOffsetX() int       // 内容 X 偏移
func (b BoxModel) ContentOffsetY() int       // 内容 Y 偏移
func (b BoxModel) TotalWidth(content int) int
func (b BoxModel) TotalHeight(content int) int
func (b BoxModel) InnerWidth(total int) int
func (b BoxModel) InnerHeight(total int) int
```

---

## 测量阶段伪代码

```go
func Measure(node Node, constraints Constraints) Size {
    // 1. 获取 BoxModel
    boxModel := node.GetBoxModel()

    // 2. 扣除 padding/border，创建内部约束
    innerConstraints := Constraints{
        MaxWidth:  max(0, constraints.MaxWidth - boxModel.HorizontalPadding()),
        MaxHeight: max(0, constraints.MaxHeight - boxModel.VerticalPadding()),
        // ... 同样处理 Min
    }

    // 3. 测量内容
    contentSize := node.MeasureContent(innerConstraints)

    // 4. 加回 padding/border
    return Size{
        Width:  contentSize.Width + boxModel.HorizontalPadding(),
        Height: contentSize.Height + boxModel.VerticalPadding(),
    }
}
```

---

## 布局阶段伪代码

```go
func Layout(node Node, x, y int, constraints Constraints) LayoutBox {
    // 1. 测量节点
    size := Measure(node, constraints)

    // 2. 创建 LayoutBox
    box := LayoutBox{
        X:      x,
        Y:      y,
        Width:  size.Width,
        Height: size.Height,
    }

    // 3. 获取 BoxModel
    boxModel := node.GetBoxModel()

    // 4. 计算偏移
    offsetX := boxModel.ContentOffsetX()
    offsetY := boxModel.ContentOffsetY()

    // 5. 布局子节点（使用内部空间）
    innerWidth := boxModel.InnerWidth(size.Width)
    innerHeight := boxModel.InnerHeight(size.Height)

    // 6. 使用 FlexLayout 获取内部布局
    internalBoxes := flexLayout.LayoutChildren(innerWidth, innerHeight)

    // 7. 转换坐标，递归布局
    for i, internal := range internalBoxes {
        childNode := node.Children()[i]

        childX := x + internal.X + offsetX
        childY := y + internal.Y + offsetY

        childBox := Layout(childNode, childX, childY, ...)
        box.Children = append(box.Children, childBox)
    }

    return box
}
```

---

## 常见场景示例

### 场景 1: 简单的 Padding 容器

```go
container := HStack().
    Padding(10).  // 上下左右各 10
    Children(
        Text("Hello"),
        Text("World"),
    )

// 约束: 100x20

// 测量:
//   - 内部空间: 100-10-10 = 80
//   - 内容宽度: 5 + 5 = 10
//   - 总宽度: 10 + 10 + 10 = 30

// 布局:
//   - 第1个子: (10+0, 10+0) = (10, 10)
//   - 第2个子: (10+6, 10+0) = (16, 10)
//                                        ↑ 假设有 gap=6
```

### 场景 2: 嵌套的 Padding

```go
outer := HStack().
    Padding(5).
    Children(
        VStack().
            Padding(3).
            Children(
                Text("A"),
                Text("B"),
            ),
    )

// 测量:
//   outer.Measure(100, 50)
//     innerWidth = 100 - 5 - 5 = 90
//     innerHeight = 50 - 5 - 5 = 40
//
//     inner.Measure(90, 40)
//       innerInnerWidth = 90 - 3 - 3 = 84
//       innerInnerHeight = 40 - 3 - 3 = 34
//
//       A.Measure(84, 34) → Size{1, 1}
//       B.Measure(84, 34) → Size{1, 1}
//
//       contentWidth = 1
//       contentHeight = 1 + 1 = 2
//       innerSize = Size{1+6, 2+6} = Size{7, 8}
//
//     outerSize = Size{7+10, 8+10} = Size{17, 18}
```

### 场景 3: Padding + Border

```go
container := Box().
    Padding(10).
    Border(BorderSingle).
    Child(Text("Content"))

// 测量:
//   borderLR = 2  // 左右边框各 1
//   borderTB = 2  // 上下边框各 1
//   paddingLR = 20 // padding 左右各 10
//   paddingTB = 20 // padding 上下各 10
//
//   innerWidth = 100 - 20 - 2 = 78
//   innerHeight = 50 - 20 - 2 = 28
//
//   content.Measure(78, 28) → Size{7, 1}
//
//   totalWidth = 7 + 20 + 2 = 29
//   totalHeight = 1 + 20 + 2 = 23

// 布局:
//   OffsetX = 10 + 1 = 11  // padding.Left + border/2
//   OffsetY = 10 + 1 = 11
//
//   child.X = 0 + 11 = 11
//   child.Y = 0 + 11 = 11
```

### 场景 4: Margin（仅布局阶段）

```go
container := HStack().Children(
    Text("A").Margin(2),  // 左右各 2
    Text("B").Margin(4),  // 左右各 4
)

// 测量:
//   A.Measure(100, 20) → Size{1, 1}  // 不包含 margin
//   B.Measure(100, 20) → Size{1, 1}  // 不包含 margin

// 布局 (FlexLayout):
//   fixedTotal = 1 + 1 + (2+2) + (4+4) = 14
//
//   A.X = 0 + 2 = 2  // 左 margin
//   B.X = (1+4) + 4 = 9
//                ↑   ↑ A 的 content + 右 margin
//                    ↑ B 的左 margin
```

---

## 当前实现的问题

### 问题 1: Padding 在约束传播中未处理

```go
// 当前实现
func (e *Engine) Measure(node Node, constraints Constraints) Size {
    // ❌ 没有扣除 padding/border
    innerConstraints := constraints
    contentSize := node.Measure(innerConstraints)

    // ❌ 可能加回也可能不加
    return contentSize
}
```

### 问题 2: Padding 在布局偏移中未应用

```go
// 当前实现
func (e *Engine) layoutNodeWithDepth(...) *LayoutBox {
    // ✅ Border 的偏移被应用
    borderWidth := nodeBorder.HorizontalPadding()

    // ❌ Padding 的偏移未被应用
    // 没有 code 这里

    childX := x + childBox.X + borderWidth/2
}
```

---

## 调试清单

### 测量不正确时检查：

- [ ] 父节点的 padding/border 是否被正确扣除？
- [ ] 子节点返回的尺寸是内容尺寸还是总尺寸？
- [ ] 总尺寸是否正确加回了 padding/border？

### 布局不正确时检查：

- [ ] 容器的 padding 帧是否正确计算 `ContentOffset`？
- [ ] 子节点位置是否加上了 offset？
- [ ] 边框偏移是否被正确应用？
- [ ] Margin 的方向是否正确（主轴 vs 跨轴）？

### 尺寸计算时检查：

- [ ] `BoxModel.HorizontalPadding()` 是否正确？
- [ ] `BoxModel.VerticalPadding()` 是否正确？
- [ ] `ContentOffsetX/Y` 是否正确？
- [ ] `InnerWidth/Height` 是否正确？

---

## 测试用例模板

```go
func TestBoxModel(t *testing.T) {
    engine := NewEngine()

    // 创建节点
    node := CreateNode().
        Padding(10, 20, 10, 20).  // Top, Right, Bottom, Left
        Children(...)

    // 测量
    constraints := NewConstraints(0, 200, 0, 100)
    size := engine.Measure(node, constraints)

    // 验证
    assert.Equal(t, expectedWidth, size.Width)
    assert.Equal(t, expectedHeight, size.Height)

    // 布局
    box := engine.Layout(node, 0, 0, constraints)

    // 验证子节点位置
    assert.Equal(t, expectedChildX, box.Children[0].X)
    assert.Equal(t, expectedChildY, box.Children[0].Y)
}
```

---

## 快速决策表

| 场景 | 如何处理 Padding？ | 如何处理 Border？ | 如何处理 Margin？ |
|------|------------------|-----------------|-----------------|
| 测量容器 | 扣除后传给子节点 | 扣除后传给子节点 | 不处理 |
| 返回测量结果 | 加回 | 加回 | 不包含 |
| 布局容器 | 应用偏移 | 应用偏移 | 不处理 |
| 计算内部空间 | 扣除 | 扣除 | 不扣除 |
| 计算子节点位置 | 加偏移 | 加偏移 | 主轴方向累积 |
| FlexLayout | 用于内部约束 | - | 主轴 margin 累积 |

---

## 相关文档

- [Box Model 当前状态分析](box_model_current_state.md)
- [Box Model 优化方案](box_model_optimization_plan.md)
- [Box Model 流程图](box_model_flow_diagram.md)
- [Margin 与测量](margin_and_measurement.md)
- [Margin Bug 分析](margin_bug_analysis.md)
