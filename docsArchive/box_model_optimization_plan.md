# Box Model 全面优化方案

## 目标

1. **统一盒模型接口**：将 Padding, Margin, Border 统一到 `BoxModel`
2. **正确的约束传播**：测量时扣除 padding/border
3. **正确的尺寸计算**：返回时加回 padding/border
4. **正确的位置偏移**：布局时应用所有偏移

---

## 一、接口设计

### 1.1 新增 BoxModel 结构体

```go
// runtime/layout/box_model.go
package layout

// BoxModel 统一表示节点的盒模型属性
type BoxModel struct {
    Margin  Margin
    Padding Padding
    Border  Border
}

// HorizontalPadding 返回水平方向占用的总空间
// 包括左边框 + 左padding + 右padding + 右边框
func (b BoxModel) HorizontalPadding() int {
    return b.Border.HorizontalPadding() +
        b.Padding.Left +
        b.Padding.Right
}

// VerticalPadding 返回垂直方向占用的总空间
func (b BoxModel) VerticalPadding() int {
    return b.Border.VerticalPadding() +
        b.Padding.Top +
        b.Padding.Bottom
}

// ContentOffsetX 返回内容区域的 X 偏移
func (b BoxModel) ContentOffsetX() int {
    return b.Padding.Left + b.Border.HorizontalPadding()/2
}

// ContentOffsetY 返回内容区域的 Y 偏移
func (b BoxModel) ContentOffsetY() int {
    return b.Padding.Top + b.Border.VerticalPadding()/2
}

// TotalWidth 计算总宽度
func (b BoxModel) TotalWidth(contentWidth int) int {
    return contentWidth + b.HorizontalPadding()
}

// TotalHeight 计算总高度
func (b BoxModel) TotalHeight(contentHeight int) int {
    return contentHeight + b.VerticalPadding()
}

// InnerWidth 计算内部可用宽度
func (b BoxModel) InnerWidth(totalWidth int) int {
    return max(0, totalWidth - b.HorizontalPadding())
}

// InnerHeight 计算内部可用高度
func (b BoxModel) InnerHeight(totalHeight int) int {
    return max(0, totalHeight - b.VerticalPadding())
}
```

### 1.2 BoxModelProvider 接口

```go
// BoxModelProvider 提供盒模型信息的接口
// 实现此接口的节点将自动处理 padding/border 的约束传播
type BoxModelProvider interface {
    Node
    GetBoxModel() BoxModel
}
```

### 1.3 为现有类型实现接口

```go
// runtime/box/box.go
type Box struct {
    // 现有字段
    margin  Margin
    padding [4]int
    border  Border
}

func (b *Box) GetBoxModel() layout.BoxModel {
    return layout.BoxModel{
        Margin:  b.margin,
        Padding: layout.Padding{
            Left:   b.padding[3],
            Right:  b.padding[1],
            Top:    b.padding[0],
            Bottom: b.padding[2],
        },
        Border: b.border,
    }
}
```

---

## 二、测量阶段修复

### 2.1 Engine.Measure 修复

```go
// runtime/layout/types.go

// Measure 测量节点的尺寸
// 返回的尺寸包含 padding 和 border，但不包含 margin
func (e *Engine) Measure(node Node, constraints Constraints) Size {
    // Step 1: 获取盒模型
    var boxModel BoxModel
    if provider, ok := node.(BoxModelProvider); ok {
        boxModel = provider.GetBoxModel()
    }

    // Step 2: 计算内部约束
    minInnerWidth := max(0, constraints.MinWidth - boxModel.HorizontalPadding())
    maxInnerWidth := max(0, constraints.MaxWidth - boxModel.HorizontalPadding())
    minInnerHeight := max(0, constraints.MinHeight - boxModel.VerticalPadding())
    maxInnerHeight := max(0, constraints.MaxHeight - boxModel.VerticalPadding())

    // 如果内部空间变为负数或太小，返回最小尺寸
    if maxInnerWidth < 0 {
        maxInnerWidth = 0
    }
    if maxInnerHeight < 0 {
        maxInnerHeight = 0
    }

    innerConstraints := Constraints{
        MinWidth:  minInnerWidth,
        MaxWidth:  maxInnerWidth,
        MinHeight: minInnerHeight,
        MaxHeight: maxInnerHeight,
        Bounded:   constraints.Bounded,
    }

    // Step 3: 测量内容
    var contentSize Size
    if measurable, ok := node.(Measurable); ok {
        contentSize = measurable.Measure(innerConstraints)
    } else {
        // 非可测量的节点，使用其尺寸
        w, h := node.GetSize()
        contentSize.Width = innerConstraints.ConstrainWidth(w)
        contentSize.Height = innerConstraints.ConstrainHeight(h)
    }

    // Step 4: 返回总尺寸（包含 padding/border）
    totalWidth := contentSize.Width + boxModel.HorizontalPadding()
    totalHeight := contentSize.Height + boxModel.VerticalPadding()

    // 确保不超过约束
    totalWidth = constraints.ConstrainWidth(totalWidth)
    totalHeight = constraints.ConstrainHeight(totalHeight)

    return Size{Width: totalWidth, Height: totalHeight}
}
```

### 2.2 FlexLayout.Measure 修复

```go
// runtime/layout/flex.go

// Measure 测量布局的尺寸
// 返回的尺寸包含 padding
func (f *FlexLayout) Measure(constraints Constraints) Size {
    // 计算可用空间（扣除 padding）
    availableWidth := max(0, constraints.MaxWidth - f.style.Padding.Horizontal())
    availableHeight := max(0, constraints.MaxHeight - f.style.Padding.Vertical())

    if len(f.children) == 0 {
        // 空容器返回 padding 占用
        contentWidth := constraints.ConstrainWidth(0)
        contentHeight := constraints.ConstrainHeight(0)
        return Size{
            Width:  contentWidth + f.style.Padding.Horizontal(),
            Height: contentHeight + f.style.Padding.Vertical(),
        }
    }

    // 创建内部约束
    innerConstraints := Constraints{
        MinWidth:  max(0, constraints.MinWidth - f.style.Padding.Horizontal()),
        MaxWidth:  availableWidth,
        MinHeight: max(0, constraints.MinHeight - f.style.Padding.Vertical()),
        MaxHeight: availableHeight,
        Bounded:   constraints.Bounded,
    }

    // Phase 1: 测量所有子节点
    childSizes := make([]Size, len(f.children))
    for i, child := range f.children {
        if measurable, ok := child.(Measurable); ok {
            childSizes[i] = measurable.Measure(innerConstraints)
        }
    }

    // Phase 2: 计算内容尺寸
    isRow := f.style.Direction == FlexRow
    contentWidth := 0
    contentHeight := 0
    totalGaps := 0

    if isRow {
        // 水平布局
        for i, size := range childSizes {
            contentWidth += size.Width
            contentHeight = max(contentHeight, size.Height)
        }
        totalGaps = (len(f.children) - 1) * f.style.Gap
        contentWidth += totalGaps
    } else {
        // 垂直布局
        for i, size := range childSizes {
            contentHeight += size.Height
            contentWidth = max(contentWidth, size.Width)
        }
        totalGaps = (len(f.children) - 1) * f.style.Gap
        contentHeight += totalGaps
    }

    // Phase 3: 考虑 flex 子节点的 grow/shrink
    // 如果有约束限制，可能需要调整
    if constraints.Bounded {
        if isRow && contentWidth > availableWidth {
            // 可能需要 shrink
            // ...
        } else if !isRow && contentHeight > availableHeight {
            // ...
        }
    }

    // 返回总尺寸（包含 padding）
    totalWidth := contentWidth + f.style.Padding.Horizontal()
    totalHeight := contentHeight + f.style.Padding.Vertical()

    return Size{Width: totalWidth, Height: totalHeight}
}
```

### 2.3 FlexLayout.LayoutChildren 修复

```go
// LayoutChildren 布局子节点
// 输入的 width/height 是容器的内部空间（已扣除 padding）
func (f *FlexLayout) LayoutChildren(width, height int) []LayoutBox {
    // Phase 1: 测量子节点
    n := len(f.children)
    childSizes := make([]Size, n)

    // 保存 margin 信息
    childMarginContent := make([]int, n)   // 主轴 margin
    childMarginStart := make([]int, n)    // 主轴起始 margin
    childMarginCross := make([]int, n)     // 跨轴 margin

    innerConstraints := NewConstraints(0, width, 0, height)

    for i, child := range f.children {
        if measurable, ok := child.(Measurable); ok {
            childSizes[i] = measurable.Measure(innerConstraints)
        }

        // 获取 margin
        if marginal, ok := child.(Marginal); ok {
            m := marginal.GetMargin()
            if f.style.Direction == FlexRow {
                childMarginContent[i] = m.Horizontal() // Left + Right
                childMarginStart[i] = m.Left
                childMarginCross[i] = m.Vertical()    // Top + Bottom
            } else {
                childMarginContent[i] = m.Vertical()   // Top + Bottom
                childMarginStart[i] = m.Top
                childMarginCross[i] = m.Horizontal()   // Left + Right
            }
        }
    }

    // Phase 2: 计算固定尺寸（包含主轴 margin）
    fixedTotal := 0
    for i, size := range childSizes {
        if f.style.Direction == FlexRow {
            fixedTotal += size.Width  // 仅内容宽度，不含 margin 在内
                                       // margin 后续单独处理
        } else {
            fixedTotal += size.Height
        }
    }

    // 添加主轴 margin
    for i := range childMarginContent {
        fixedTotal += childMarginContent[i]
    }

    // 添加 gap
    totalGaps := max(0, n-1) * f.style.Gap
    fixedTotal += totalGaps

    // 计算剩余空间
    var remainingSpace int
    if f.style.Direction == FlexRow {
        remainingSpace = width - fixedTotal
    } else {
        remainingSpace = height - fixedTotal
    }
    if remainingSpace < 0 {
        remainingSpace = 0
    }

    // Phase 3: 分配 flex 空间
    totalGrowFlex := 0
    for idx, flex := range f.style.FlexibleChildren {
        if idx < n && flex != nil && flex.Grow > 0 {
            totalGrowFlex += flex.Grow
        }
    }

    finalSizes := make([]Size, n)
    for i, size := range childSizes {
        finalSizes[i] = size
        if flex, ok := f.style.FlexibleChildren[i]; ok && flex != nil && flex.Grow > 0 {
            extra := int(float64(remainingSpace) * float64(flex.Grow) / float64(totalGrowFlex))
            if f.style.Direction == FlexRow {
                finalSizes[i].Width += extra
            } else {
                finalSizes[i].Height += extra
            }
        }
    }

    // Phase 4: 计算位置
    boxes := make([]LayoutBox, n)
    mainPos := 0

    for i := range boxes {
        size := finalSizes[i]
        marginStart := childMarginStart[i]
        marginContent := childMarginContent[i]

        if f.style.Direction == FlexRow {
            // 主轴位置（包含 margin）
            boxes[i].X = mainPos + marginStart
            mainPos += marginStart + size.Width + (marginContent - marginStart) + f.style.Gap

            // 跨轴位置（仅处理 margin，padding 在返回后处理）
            // 跨轴 margin：Top for row
            marginTop := 0
            if marginal, ok := f.children[i].(Marginal); ok {
                marginTop = marginal.GetMargin().Top
            }
            boxes[i].Y = marginTop
            boxes[i].Width = size.Width
            boxes[i].Height = min(size.Height, height)
        } else {
            // VStack
            boxes[i].Y = mainPos + marginStart
            mainPos += marginStart + size.Height + (marginContent - marginStart) + f.style.Gap

            // 跨轴 margin：Left for column
            marginLeft := 0
            if marginal, ok := f.children[i].(Marginal); ok {
                marginLeft = marginal.GetMargin().Left
            }
            boxes[i].X = marginLeft
            boxes[i].Width = min(size.Width, width)
            boxes[i].Height = size.Height
        }
    }

    // Phase 5: 应用跨轴对齐
    f.applyContentAlignment(boxes, width, height)

    return boxes
}

// 注意：返回的 boxes 的 X/Y 是相对于 padding 内部的内容区域
// 调用者需要添加 padding 偏移
```

---

## 三、布局阶段修复

### 3.1 Engine.layoutNodeWithDepth 修复

```go
// runtime/layout/types.go

// layoutNodeWithDepth 递归布局节点
func (e *Engine) layoutNodeWithDepth(node Node, x, y int, constraints Constraints, depth int) *LayoutBox {
    e.depth = depth

    // Step 1: 获取盒模型
    var boxModel BoxModel
    if provider, ok := node.(BoxModelProvider); ok {
        boxModel = provider.GetBoxModel()
    }

    // Step 2: 测量节点（已包含 padding/border）
    size := e.Measure(node, constraints)
    width, height := size.Width, size.Height

    // Step 3: 创建 LayoutBox
    box := &LayoutBox{
        Node:     node,
        X:        x,
        Y:        y,
        Width:    width,
        Height:   height,
        Children: nil,
    }

    // Step 4: 如果没有内部布局，直接返回
    if len(node.Children()) == 0 {
        return box
    }

    // Step 5: 计算内部空间（扣除 padding/border）
    innerWidth := boxModel.InnerWidth(width)
    innerHeight := boxModel.InnerHeight(height)

    // Step 6: 使用 FlexLayout 布局子节点
    if flexLayout, ok := node.(*FlexLayout); ok {
        childrenBoxes := flexLayout.LayoutChildren(innerWidth, innerHeight)

        // Step 7: 应用 padding/bargin/border 偏移
        contentOffsetX := boxModel.ContentOffsetX()
        contentOffsetY := boxModel.ContentOffsetY()

        childNodes := node.Children()
        for i, childBox := range childrenBoxes {
            if i < len(childNodes) {
                child := childNodes[i]

                // 计算子节点位置
                childX := x + childBox.X + contentOffsetX
                childY := y + childBox.Y + contentOffsetY

                // 获取子节点的 margin
                childMargin := Margin{}
                if marginal, ok := child.(Marginal); ok {
                    childMargin = marginal.GetMargin()
                }

                // HStack: 主轴是水平，margin.Left 包含在 childBox.X 中
                // 但需要确保跨轴 margin（Top）被应用
                if flexLayout.style.Direction == FlexRow {
                    // childBox.X 已经包含了 margin.Start + content + margin.End
                    // (在 LayoutChildren 中计算)
                    childX = x + childBox.X + contentOffsetX
                    childY = y + childBox.Y + contentOffsetY
                    // childBox.Y 已经包含 margin.Top
                } else {
                    // VStack
                    childX = x + childBox.X + contentOffsetX
                    childY = y + childBox.Y + contentOffsetY
                    // childBox.X 已经包含 margin.Left
                }

                // 创建子节点约束
                childConstraints := NewConstraints(0, childBox.Width, 0, childBox.Height)

                // 递归布局子节点
                childBox := e.layoutNodeWithDepth(child, childX, childY, childConstraints, depth+1)
                box.Children = append(box.Children, childBox)
            }
        }
    }

    return box
}
```

### 3.2 调整 FlexLayout 布局逻辑

关键点：
1. `LayoutChildren` 返回的是内部内容区域的布局
2. 主轴 margin 包含在 `box.X`/`box.Y` 中
3. 跨轴 margin 作为起始偏移
4. Padding 帏 `layoutNodeWithDepth` 添加

---

## 四、组件适配

### 4.1 LayoutNode 实现

```go
// runtime/ui/layout.go

// 实现 BoxModelProvider
func (l *LayoutNode) GetBoxModel() layout.BoxModel {
    return layout.BoxModel{
        Margin: layout.Margin{
            Left:   0,  // LayoutNode 暂不支持 margin
            Right:  0,
            Top:    0,
            Bottom: 0,
        },
        Padding: layout.Padding{
            Left:   l.padding[3],
            Right:  l.padding[1],
            Top:    l.padding[0],
            Bottom: l.padding[2],
        },
        Border: layout.Border{},  // LayoutNode 暂不支持 border
    }
}
```

### 4.2 Box 实现

```go
// runtime/box/box.go

// 实现 BoxModelProvider
func (b *Box) GetBoxModel() layout.BoxModel {
    return layout.BoxModel{
        Margin:  b.margin,
        Padding: layout.Padding{
            Left:   b.padding[3],
            Right:  b.padding[1],
            Top:    b.padding[0],
            Bottom: b.padding[2],
        },
        Border: b.border,
    }
}
```

### 4.3 UI 层测量逻辑简化

由于 `Engine.Measure` 已经处理 padding，UI 层的测量逻辑可以简化：

```go
// runtime/ui/layout.go

// Measure 简化版本
func (l *LayoutNode) Measure(constraints Constraints) Size {
    if measurable, ok := l.node.(Measurable); ok {
        // 直接测量，padding/border 由布局引擎处理
        return measurable.Measure(constraints)
    }

    // 默认行为
    w, h := l.GetSize()
    return Size{Width: w, Height: h}
}
```

---

## 五、测试策略

### 5.1 单元测试

```go
// runtime/layout/box_model_test.go
func TestBoxModelCalculations(t *testing.T) {
    boxModel := BoxModel{
        Padding: Padding{Top: 10, Right: 20, Bottom: 30, Left: 40},
        Border:  Border{Style: BorderSingle},
    }

    // 水平占用：边框(2) + padding(40+20) = 62
    assert.Equal(t, 62, boxModel.HorizontalPadding())

    // 垂直占用：边框(2) + padding(10+30) = 42
    assert.Equal(t, 42, boxModel.VerticalPadding())

    // 内容偏移
    assert.Equal(t, 41, boxModel.ContentOffsetX())  // 40 + 2/2
    assert.Equal(t, 11, boxModel.ContentOffsetY())  // 10 + 2/2
}

func TestMeasureWithPadding(t *testing.T) {
    engine := NewEngine()

    // 创建带 padding 的节点
    node := &LayoutNode{
        padding: [4]int{10, 20, 30, 40}, // top, right, bottom, left
        node: &Text{Content: "Hello"},
    }

    constraints := NewConstraints(0, 200, 0, 100)
    size := engine.Measure(node, constraints)

    // 内容宽度 = 5 ("Hello")
    // 总宽度 = 5 + 2 + 40 + 20 = 67
    assert.Equal(t, 67, size.Width)
    // 内容高度 = 1
    // 总高度 = 1 + 2 + 10 + 30 = 43
    assert.Equal(t, 43, size.Height)
}
```

### 5.2 集成测试

```go
func TestBoxLayoutWithPaddingMargin(t *testing.T) {
    // 创建包含 padding 的容器
    container := HStack().
        Padding(10).
        Children(
            Text("A").Margin(5),
            Text("B").Margin(10),
        )

    // 布局
    engine := NewEngine()
    box := engine.Layout(container)

    // 验证尺寸
    // 内部内容：A(5) + margin(5+5) + gap(假设0) + B(5) + margin(10+10) = 40
    // 加上 padding(10+10) = 20
    // 总宽度 = 40 + 20 = 60
    assert.Equal(t, 60, box.Width)
}
```

---

## 六、迁移步骤

### 阶段 1：接口定义（第 1-2 天）
- [x] 创建 `box_model.go` 文件
- [ ] 定义 `BoxModel` 结构体
- [ ] 定义 `BoxModelProvider` 接口
- [ ] 编写单元测试

### 阶段 2：测量阶段（第 3-4 天）
- [ ] 修改 `Engine.Measure()`
- [ ] 修改 `FlexLayout.Measure()`
- [ ] 编写测量测试

### 阶段 3：布局阶段（第 5-6 天）
- [ ] 修改 `Engine.layoutNodeWithDepth()`
- [ ] 修改 `FlexLayout.LayoutChildren()`
- [ ] 修复 margin 处理
- [ ] 编写布局测试

### 阶段 4：组件适配（第 7-8 天）
- [ ] `LayoutNode` 实现 `BoxModelProvider`
- [ ] `Box` 实现 `BoxModelProvider`
- [ ] 适配其他组件

### 阶段 5：测试和文档（第 9-10 天）
- [ ] 集成测试
- [ ] 端到端测试
- [ ] 更新文档

---

## 七、向后兼容性

### 渐进式迁移

1. **保留旧接口**
```go
// 保留 Marginal 接口
type Marginal interface {
    GetMargin() Margin
}

// 新的 BoxModelProvider 同时实现 Marginal
type BoxModelProvider interface {
    Marginal  // 隐式嵌入
    GetBoxModel() BoxModel
}
```

2. **辅助方法**
```go
// 为布局引擎提供辅助方法处理旧接口
func getBoxModel(node Node) BoxModel {
    // 尝试 BoxModelProvider
    if provider, ok := node.(BoxModelProvider); ok {
        return provider.GetBoxModel()
    }

    // 回退到旧接口
    var boxModel BoxModel

    if marginal, ok := node.(Marginal); ok {
        boxModel.Margin = marginal.GetMargin()
    }

    return boxModel
}
```

---

## 八、性能考虑

### 8.1 缓存

```go
// BoxModel 可以缓存
func (n *LayoutNode) GetBoxModel() BoxModel {
    if n.cachedBoxModel == nil {
        n.cachedBoxModel = &BoxModel{
            Margin:  Margin{...},
            Padding: Padding{...},
            Border:  Border{...},
        }
    }
    return *n.cachedBoxModel
}
```

### 8.2 避免重复计算

```go
// 预计算常用值
func (b BoxModel) Precompute() BoxModelCache {
    return BoxModelCache{
        HorizontalPadding: b.HorizontalPadding(),
        VerticalPadding:   b.VerticalPadding(),
        ContentOffsetX:    b.ContentOffsetX(),
        ContentOffsetY:    b.ContentOffsetY(),
    }
}
```

---

## 总结

本方案提供了一个全面的优化路径，主要改进：

1. ✅ 统一的 `BoxModel` 接口简化了 API
2. ✅ 正确的约束传播确保布局精确
3. ✅ 清晰的语义：测量 = 内容，布局 = 位置
4. ✅ 向后兼容的迁移路径
5. ✅ 完整的测试策略

实施后，开发者可以容易地为任何组件添加 padding/border/margin，而无需关心底层的约束传播细节。
