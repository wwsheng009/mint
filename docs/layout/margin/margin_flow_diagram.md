# Margin 处理流程完整图解

## 完整流程图

```
┌────────────────────────────────────────────────────────────────────────────┐
│  Phase 1: 父容器 Measurement (测量阶段)                                      │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Engine.layoutNodeWithDepth(root, constraints, x, y, depth)                 │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │  1. Root 收到约束: MaxWidth=80, MaxHeight=25                         │  │
│  │                                                                      │  │
│  │  2. 调用 measurable.Measure(constraints)                             │  │
│  │     ┌─────────────────────────────────────────────────────┐         │  │
│  │     │  Measure(constraints) -> Size{Width, Height}        │         │  │
│  │     │  - 节点不考虑自己的 margin                            │         │  │
│  │     │  - 只返回内容的尺寸                                  │         │  │
│  │     └─────────────────────────────────────────────────────┘         │  │
│  │                                                                      │  │
│  │  3. 父容器确定自己的最终尺寸 (width, height)                         │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌────────────────────────────────────────────────────────────────────────────┐
│  Phase 2: 子节点布局 (Flex Layout 计算)                                     │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  FlexLayout.LayoutChildren(width, height)                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │  输入:                                                              │  │
│  │  • parentWidth = 60 (VStack 宽度)                                   │  │
│  │  • parentHeight = 20 (VStack 高度)                                  │  │
│  │                                                                     │  │
│  │  子节点: [ButtonA, ButtonB]                                         │  │
│  │  • ButtonA: Flex=1, MarginV(5,5)                                    │  │
│  │  • ButtonB: Flex=1, MarginV(5,5)                                    │  │
│  │                                                                     │  │
│  │  1. Phase 1: 测量子节点 ( Measure Loop )                           │  │
│  │     ┌─────────────────────────────────────────────────────────┐     │  │
│  │     │  childConstraints = f.childConstraints(constraints)     │     │  │
│  │     │                      ↓                                  │     │  │
│  │     │  ButtonA.Measure(MaxHeight=20 - padding)                 │     │  │
│  │     │  ButtonB.Measure(MaxHeight=20 - padding)                 │     │  │
│  │     │                      ↓                                  │     │  │
│  │     │  childContentSize = {Height: 1, Height: 1}              │     │  │
│  │     └─────────────────────────────────────────────────────────┘     │  │
│  │                                                                     │  │
│  │  2. Phase 2: 计算 Flex 分配                                        │  │
│  │     ┌─────────────────────────────────────────────────────────┐     │  │
│  │     │  totalContentHeight = 1 + 1 = 2                          │     │  │
│  │     │  remainingHeight = 20 - 2 = 18                          │     │  │
│  │     │  flexGrowTotal = 2 (两个子节点都是 Flex=1)               │     │  │
│  │     │                                                             │     │  │
│  │     │  每个子节点分配: 1 + (18 / 2) = 10                       │     │  │
│  │     │                                                             │     │  │
│  │     │  childBoxHeight = 10 (包含 margin 空间!)                 │     │  │
│  │     └─────────────────────────────────────────────────────────┘     │  │
│  │                                                                     │  │
│  │  3. Phase 3: 创建 LayoutBox                                        │  │
│  │     ┌─────────────────────────────────────────────────────────┐     │  │
│  │     │  对于每个子节点:                                         │     │  │
│  │     │  childBox = LayoutBox {                                  │     │  │
│  │     │      Width: childBoxWidth,   (包含 margin 空间)          │     │  │
│  │     │      Height: childBoxHeight, (包含 margin 空间)          │     │  │
│  │     │      X: 0, Y: 累积位置                                     │     │  │
│  │     │  }                                                        │     │  │
│  │     └─────────────────────────────────────────────────────────┘     │  │
│  │                                                                     │  │
│  │  返回: []LayoutBox {childBoxA, childBoxB}                           │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌────────────────────────────────────────────────────────────────────────────┐
│  Phase 3: 应用 Margin 到子节点位置 (layoutNodeWithDepth)                    │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  for i, childBox := range childBoxes {                                     │
│      child := node.Children()[i]                                           │
│                                                                             │
│      // 获取子节点的 margin                                                │
│      marginTop, marginBottom, marginLeft, marginRight := ...                │
│                                                                             │
│      // 计算 childX 和 childY                                              │
│      if isFlexColumn (VStack):                                             │
│          ┌─────────────────────────────────────────────────────────────┐   │
│          │  childY = containerY + childBox.Y + borderOffsetY +         │   │
│          │            mainAxisMarginOffset + marginTop                 │   │
│          │                                                             │   │
│          │  childX = containerX + childBox.X + borderOffsetX +         │   │
│          │            marginLeft (跨轴 margin)                          │   │
│          │                                                             │   │
│          │  mainAxisMarginOffset += marginTop + marginBottom           │   │
│          └─────────────────────────────────────────────────────────────┘   │
│                                                                             │
│      // ✅ 关键: 创建递归约束时扣除 margin                                 │
│      ┌─────────────────────────────────────────────────────────────────────┐│
│      │  childConstraints = Constraints {                                  ││
│      │      MinWidth:  max(0, childBox.Width-marginLeft-marginRight),    ││
│      │      MaxWidth:  max(0, childBox.Width-marginLeft-marginRight),    ││
│      │      MinHeight: max(0, childBox.Height-marginTop-marginBottom),   ││
│      │      MaxHeight: max(0, childBox.Height-marginTop-marginBottom),   ││
│      │  }                                                                ││
│      │                                                                    ││
│      │  递归布局:                                                         ││
│      │  subBox := e.layoutNodeWithDepth(                                 ││
│      │      child, childConstraints,                                     ││
│      │      childX, childY, depth+1)                                     ││
│      └─────────────────────────────────────────────────────────────────────┘│
│                                                                             │
│      if subBox != nil {                                                    │
│          box.Children = append(box.Children, subBox)                       │
│      }                                                                      │
│  }                                                                          │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌────────────────────────────────────────────────────────────────────────────┐
│  数值示例: VStack [BtnA, BtnV(5,5), BtnB]                                    │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  父容器: width=60, height=20                                                │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │  FlexLayout 计算:                                                      │ │
│  │                                                                        │ │
│  │  BtnA:                                                                  │ │
│  │  • Measure: Height=1                                                    │ │
│  │  • Flex 分配: 1 + (18/2) = 10                                          │ │
│  │  • childBox.Height = 10                                                │ │
│  │                                                                        │ │
│  │  第一子节点位置计算:                                                    │ │
│  │  • mainAxisMarginOffset = 0                                            │ │
│  │  • childY = 0 + 0 + 0 + 0 + 5 = 5                                      │ │
│  │  • childX = 0 + 0 + 0 + 0 = 0                                          │ │
│  │  • mainAxisMarginOffset += 5 + 5 = 10                                   │ │
│  │                                                                        │ │
│  │  递归约束扣除 margin:                                                   │ │
│  │  • MaxWidth = 60 - (0 + 0) = 60                                        │ │
│  │  • MaxHeight = 10 - (5 + 5) = 0 ⚠                                   │ │
│  │                                                                        │ │
│  │  ┌─────────────────────────────────────────────────────────────────┐   │ │
│  │  │  BtnA 最终:                                                   │   │ │
│  │  │  Pos: (0, 5)                                                  │   │ │
│  │  │  Size: 60x0 (高度被压缩到最小)                                 │   │ │
│  │  └─────────────────────────────────────────────────────────────────┘   │ │
│  │                                                                        │ │
│  │  BtnB:                                                                  │ │
│  │  • childBox.Y = 5 (累积偏移)                                           │ │
│  │  • childY = 0 + 5 + 0 + 10 + 5 = 20                                    │ │
│  │  • childX = 0 + 0 + 0 + 0 = 0                                          │ │
│  │                                                                        │ │
│  │  ┌─────────────────────────────────────────────────────────────────┐   │ │
│  │  │  第二个节点的位置被第一个节点的 marginTop + marginBottom 推后了   │   │ │
│  │  │  间距 = childY(BtnB) - (childY(BtnA) + heightA)                │   │ │
│  │  │       = 20 - (5 + 0) = 15                                        │   │ │
│  │  │   实际应该是: Top margin of BtnB + Bottom margin of BtnA          │   │ │
│  │  │                = 5 + 5 = 10 ✓                                    │   │ │
│  │  └─────────────────────────────────────────────────────────────────┘   │ │
│  │                                                                        │ │
│  │  (注: 上面的计算是简化的，实际还要考虑 gap 和其他因素)                 │ │
│  │                                                                        │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## 关键代码位置

### 1. Flex 布局计算子节点高度/宽度

**文件**: `runtime/layout/flex.go:280-350`
```go
// Measure 测量节点尺寸
func (f *FlexLayout) Measure(constraints Constraints) Size {
    // Phase 1: 测量所有子节点
    childConstraints := f.childConstraints(constraints, i)
    childSizes[i] = measurable.Measure(childConstraints)

    // Phase 2: 计算 flex 分配
    availableSpace := parentSize - fixedTotal
    for i, size := range childSizes {
        if flex, ok := f.style.FlexibleChildren[i]; ok {
            // 计算 flex 分配
            allocated := ...
            childBox Heights/Widths[i] = allocated
        }
    }
}
```

### 2. 创建子节点约束 (不含 margin)

**文件**: `runtime/layout/flex.go:388-425`
```go
// childConstraints 计算子节点约束
// 注意: 这里不扣除 margin!
func (f *FlexLayout) childConstraints(constraints Constraints, index int) Constraints {
    isRow := f.style.Direction == FlexRow

    // 减去内边距
    availableMain := constraints.MaxWidth - f.style.Padding.Left - f.style.Padding.Right
    availableCross := constraints.MaxHeight - f.style.Padding.Top - f.style.Padding.Bottom

    if isRow {
        return NewConstraints(0, availableMain, 0, availableCross)
    }
    return NewConstraints(0, availableCross, 0, availableMain)
}
```

### 3. 应用 Margin 到位置

**文件**: `runtime/layout/types.go:744-762`
```go
// 为子节点计算位置
var childX, childY int
if isFlexRow {
    // HStack
    childX = x + childBox.X + borderOffsetX + mainAxisMarginOffset + marginLeft
    childY = y + childBox.Y + borderOffsetY + marginTop  // 跨轴 margin
    mainAxisMarginOffset += marginLeft + marginRight
} else {
    // VStack
    childY = y + childBox.Y + borderOffsetY + mainAxisMarginOffset + marginTop
    childX = x + childBox.X + borderOffsetX + marginLeft  // 跨轴 margin
    mainAxisMarginOffset += marginTop + marginBottom
}
```

### 4. 创建递归约束 (扣除 margin)

**文件**: `runtime/layout/types.go:773-783`
```go
// ✨ 创建子节点的递归约束，扣除 margin
childConstraints := Constraints{
    MinWidth:  max(0, childBox.Width-marginLeft-marginRight),
    MaxWidth:  max(0, childBox.Width-marginLeft-marginRight),
    MinHeight: max(0, childBox.Height-marginTop-marginBottom),
    MaxHeight: max(0, childBox.Height-marginTop-marginBottom),
}

// 递归布局子节点
subBox := e.layoutNodeWithDepth(child, childConstraints, childX, childY, depth+1, visited)
```

---

## 数据流向图

```
┌──────────────────────────────────────────────────────────────────────┐
│                        数据流向                                       │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  constraints (父容器约束)                                              │
│        ↓                                                              │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Phase 1: FlexLayout 测量子节点                                  │ │
│  │  ┌────────────────────────────────────────────────────────────┐  │ │
│  │  │  子节点 1: Measure(constraints without margin)           │  │ │
│  │  │  → Size{Width: 15, Height: 1}                            │  │ │
│  │  └────────────────────────────────────────────────────────────┘  │ │
│  │  ┌────────────────────────────────────────────────────────────┐  │ │
│  │  │  子节点 2: Measure(constraints without margin)           │  │ │
│  │  │  → Size{Width: 15, Height: 1}                            │  │ │
│  │  └────────────────────────────────────────────────────────────┘  │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│        ↓                                                              │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Phase 2: FlexLayout 计算分配                                    │ │
│  │  childBox.Width[0] = 30 (包含 margin 空间)                      │ │
│  │  childBox.Width[1] = 30 (包含 margin 空间)                      │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│        ↓                                                              │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Phase 3: 应用 Margin 位置偏移                                   │ │
│  │  childX[0] = 0 + margin.left                                    │ │
│  │  childY[0] = 0 + margin.top                                     │ │
│  │  childX[1] = childBox[0].Width + margin.right[0] + margin.left[1]│
│  └─────────────────────────────────────────────────────────────────┘ │
│        ↓                                                              │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Phase 4: 创建递归约束 (扣除 margin)                            │ │
│  │  childConstraints[0] = {                                         │ │
│  │      MaxWidth: 30 - (10 + 0) = 20,                            │ │
│  │      MaxHeight: 1 - (0 + 0) = 1,                              │ │
│  │  }                                                              │ │
│  │  └─────────────────────────────────┐                            │ │
│  │         ↓                          │                            │ │
│  │  递归布局子节点 (可选)              │                            │ │
│  └────────────────────────────────────┴────────────────────────────┘ │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 相关文件

| 文件 | 说明 |
|------|------|
| `runtime/layout/types.go` | 布局引擎核心，包含 margin 位置应用和递归约束创建 |
| `runtime/layout/flex.go` | Flex 布局实现，包含子节点测量和分配 |
| `runtime/layout/constraints.go` | 约束定义和处理 |
| `docs/layout/margin/margin_and_measurement.md` | 详细的概念文档 |
