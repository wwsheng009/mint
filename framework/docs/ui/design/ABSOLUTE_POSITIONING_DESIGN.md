# Absolute 定位设计文档

**版本**: v1.0
**日期**: 2026-01-31
**来源**: demo4_layout.md, demo5_ide.md
**状态**: 🟡 中优先级

---

## 一、概述

### 1.1 设计目标

实现绝对定位系统，用于脱离正常布局流的组件定位。

### 1.2 使用场景

- 徽章 (Badge) - 右上角提示
- 提示框 (Tooltip) - 相对目标定位
- 浮动按钮 (FAB)
- 覆盖层内容

### 1.3 与 Stack 的区别

| 特性 | Stack | Absolute |
|------|-------|----------|
| 布局流 | 叠加布局 | 脱离流 |
| 定位 | 相对父容器 | 绝对坐标 |
| Z-Index | 文档顺序 | 显式控制 |
| 尺寸 | 父容器约束 | 自定义 |

---

## 二、Absolute 类型定义

### 2.1 Position 类型

```go
// framework/layout/position.go

package layout

// Position 位置类型
type Position struct {
    Value int
    Unit  PositionUnit
}

type PositionUnit int

const (
    UnitPixel PositionUnit = iota // 像素
    UnitPercent                    // 百分比 (相对父容器)
    UnitAuto                       // 自动
)

// 预定义位置
func Pixel(n int) Position {
    return Position{Value: n, Unit: UnitPixel}
}

func Percent(n int) Position {
    return Position{Value: n, Unit: UnitPercent}
}

func AutoPosition() Position {
    return Position{Value: 0, Unit: UnitAuto}
}
```

### 2.2 Absolute Props

```go
// framework/layout/absolute.go

package layout

// AbsoluteProps 绝对定位属性
type AbsoluteProps struct {
    // 位置约束
    Top    *Position
    Bottom *Position
    Left   *Position
    Right  *Position

    // 尺寸约束
    Width  *Dimension
    Height *Dimension

    // Z-Index（可选，默认为 0）
    ZIndex int

    // 锚点（相对位置）
    Anchor AnchorType
}

// AnchorType 锚点类型
type AnchorType int

const (
    AnchorTopLeft AnchorType = iota
    AnchorTop
    AnchorTopRight
    AnchorLeft
    AnchorCenter
    AnchorRight
    AnchorBottomLeft
    AnchorBottom
    AnchorBottomRight
)
```

---

## 三、Absolute API 设计

### 3.1 声明式 API

```go
// framework/layout/absolute.go

// Absolute 创建绝对定位容器
func Absolute(props AbsoluteProps, child VNode) VNode {
    return &VNodeElement{
        Type:     "Absolute",
        Props:    props,
        Children: []VNode{child},
    }
}
```

### 3.2 链式 API

```go
// framework/layout/absolute_builder.go

package layout

// AbsoluteBuilder 绝对定位构建器
type AbsoluteBuilder struct {
    props AbsoluteProps
    child VNode
}

// NewAbsolute 创建构建器
func NewAbsolute(child VNode) *AbsoluteBuilder {
    return &AbsoluteBuilder{
        child: child,
    }
}

// Top 设置顶部位置
func (b *AbsoluteBuilder) Top(v int) *AbsoluteBuilder {
    b.props.Top = &Position{Value: v, Unit: UnitPixel}
    return b
}

// TopPercent 设置顶部百分比
func (b *AbsoluteBuilder) TopPercent(v int) *AbsoluteBuilder {
    b.props.Top = &Position{Value: v, Unit: UnitPercent}
    return b
}

// Bottom 设置底部位置
func (b *AbsoluteBuilder) Bottom(v int) *AbsoluteBuilder {
    b.props.Bottom = &Position{Value: v, Unit: UnitPixel}
    return b
}

// Left 设置左侧位置
func (b *AbsoluteBuilder) Left(v int) *AbsoluteBuilder {
    b.props.Left = &Position{Value: v, Unit: UnitPixel}
    return b
}

// Right 设置右侧位置
func (b *AbsoluteBuilder) Right(v int) *AbsoluteBuilder {
    b.props.Right = &Position{Value: v, Unit: UnitPixel}
    return b
}

// Width 设置宽度
func (b *AbsoluteBuilder) Width(v int) *AbsoluteBuilder {
    dim := Fixed(v)
    b.props.Width = &dim
    return b
}

// Height 设置高度
func (b *AbsoluteBuilder) Height(v int) *AbsoluteBuilder {
    dim := Fixed(v)
    b.props.Height = &dim
    return b
}

// ZIndex 设置层级
func (b *AbsoluteBuilder) ZIndex(v int) *AbsoluteBuilder {
    b.props.ZIndex = v
    return b
}

// Anchor 设置锚点
func (b *AbsoluteBuilder) Anchor(v AnchorType) *AbsoluteBuilder {
    b.props.Anchor = v
    return b
}

// Build 构建 VNode
func (b *AbsoluteBuilder) Build() VNode {
    return Absolute(b.props, b.child)
}
```

### 3.3 UI 层 API

```go
// ui/absolute.go

package ui

// Absolute 快捷函数
func Absolute(child VNode) *layout.AbsoluteBuilder {
    return layout.NewAbsolute(child)
}

// TopLeft 定位到左上角
func TopLeft(child VNode) *layout.AbsoluteBuilder {
    return layout.NewAbsolute(child).
        Top(0).Left(0)
}

// TopRight 定位到右上角
func TopRight(child VNode) *layout.AbsoluteBuilder {
    return layout.NewAbsolute(child).
        Top(0).Right(0)
}

// BottomLeft 定位到左下角
func BottomLeft(child VNode) *layout.AbsoluteBuilder {
    return layout.NewAbsolute(child).
        Bottom(0).Left(0)
}

// BottomRight 定位到右下角
func BottomRight(child VNode) *layout.AbsoluteBuilder {
    return layout.NewAbsolute(child).
        Bottom(0).Right(0)
}

// Center 居中定位
func Center(child VNode) *layout.AbsoluteBuilder {
    return layout.NewAbsolute(child).
        Anchor(layout.AnchorCenter)
}
```

---

## 四、布局算法

### 4.1 位置计算

```go
// framework/layout/absolute_algorithm.go

package layout

// AbsoluteLayout 绝对定位布局
type AbsoluteLayout struct {
    Props     AbsoluteProps
    Child     VNode
    ParentBox Box // 父容器布局信息
}

// Layout 计算绝对定位
func (a *AbsoluteLayout) Layout(parentX, parentY, parentW, parentH int) Box {
    // 1. 解析位置约束
    top := a.resolvePosition(a.props.Top, parentH, 0)
    bottom := a.resolvePosition(a.props.Bottom, parentH, 0)
    left := a.resolvePosition(a.props.Left, parentW, 0)
    right := a.resolvePosition(a.props.Right, parentW, 0)

    // 2. 解析尺寸约束
    width := a.resolveDimension(a.props.Width, parentW)
    height := a.resolveDimension(a.props.Height, parentH)

    // 3. 计算最终位置和尺寸
    x := parentX
    y := parentY
    w := width
    h := height

    // 水平位置
    if left != nil {
        x = parentX + *left
        if right != nil {
            // left + right 都指定，计算宽度
            w = (parentX + parentW - *right) - x
        }
    } else if right != nil {
        x = (parentX + parentW - *right) - w
    }

    // 垂直位置
    if top != nil {
        y = parentY + *top
        if bottom != nil {
            // top + bottom 都指定，计算高度
            h = (parentY + parentH - *bottom) - y
        }
    } else if bottom != nil {
        y = (parentY + parentH - *bottom) - h
    }

    // 4. 处理锚点
    x, y = a.applyAnchor(x, y, w, h, parentW, parentH)

    // 5. 递归布局子元素
    a.child.Layout(x, y, w, h)

    return Box{X: x, Y: y, W: w, H: h}
}

// resolvePosition 解析位置值
func (a *AbsoluteLayout) resolvePosition(pos *Position, parentSize, defaultValue int) *int {
    if pos == nil {
        return &defaultValue
    }
    switch pos.Unit {
    case UnitPixel:
        return &pos.Value
    case UnitPercent:
        v := pos.Value * parentSize / 100
        return &v
    case UnitAuto:
        return &defaultValue
    }
    return &defaultValue
}

// resolveDimension 解析尺寸值
func (a *AbsoluteLayout) resolveDimension(dim *Dimension, parentSize int) int {
    if dim == nil {
        // 默认由内容决定
        return a.child.Measure(parentSize).W
    }
    switch d := (*dim).(type) {
    case Fixed:
        return int(d)
    case Flex:
        return parentSize * d.Grow
    case Auto:
        return a.child.Measure(parentSize).W
    case Min:
        childSize := a.child.Measure(parentSize).W
        return max(d.Min, childSize)
    case Max:
        childSize := a.child.Measure(parentSize).W
        return min(d.Max, childSize)
    }
    return 0
}

// applyAnchor 应用锚点偏移
func (a *AbsoluteLayout) applyAnchor(x, y, w, h, parentW, parentH int) (int, int) {
    switch a.props.Anchor {
    case AnchorTop:
        x = x + (parentW-w)/2
    case AnchorTopRight:
        x = parentW - w
    case AnchorLeft:
        y = y + (parentH-h)/2
    case AnchorCenter:
        x = x + (parentW-w)/2
        y = y + (parentH-h)/2
    case AnchorRight:
        x = parentW - w
        y = y + (parentH-h)/2
    case AnchorBottomLeft:
        y = parentH - h
    case AnchorBottom:
        x = x + (parentW-w)/2
        y = parentH - h
    case AnchorBottomRight:
        x = parentW - w
        y = parentH - h
    }
    // AnchorTopLeft 是默认，无需调整
    return x, y
}
```

---

## 五、使用示例

### 5.1 基础定位

```go
// 示例：右上角徽章
func ButtonWithBadge() VNode {
    return ui.Box().
        Padding(2).
        Child(
            ui.Stack(
                ui.Text("Click Me"),
                ui.Absolute(
                    ui.Text("NEW").
                        FgColor(ui.ColorRed).
                        Bold(true),
                ).
                    Top(0).
                    Right(0),
            ),
        )
}
```

### 5.2 居中定位

```go
// 示例：居中 Modal
func CenterModal() VNode {
    return ui.Box().Flex(1).Child(
        ui.Absolute(
            ui.Box().
                Width(40).
                Height(10).
                Border(true).
                Child(
                    ui.Text("Modal Content"),
                ),
        ).
            Anchor(layout.AnchorCenter),
    )
}
```

### 5.3 百分比定位

```go
// 示例：相对父容器百分比定位
func PercentPosition() VNode {
    return ui.Box().
        Width(80).
        Height(20).
        Child(
            ui.Absolute(
                ui.Box().
                    Width(20).
                    Height(5).
                    Border(true),
            ).
                LeftPercent(10).  // 父容器宽度的 10%
                TopPercent(20),   // 父容器高度的 20%
        )
}
```

### 5.4 Tooltip 定位

```go
// 示例：跟随目标的 Tooltip
func WithTooltip(target VNode, tip VNode) VNode {
    return ui.Box().Child(
        ui.Stack(
            target,
            ui.Absolute(tip).
                Top(2).       // 目标下方 2 行
                Left(0),
        ),
    )
}
```

### 5.5 浮动按钮

```go
// 示例：右下角浮动按钮
func FAB() VNode {
    return ui.Box().Flex(1).Child(
        ui.Absolute(
            ui.Button("+").OnClick(func() {
                // 处理点击
            }),
        ).
            Bottom(2).
            Right(4).
            ZIndex(100),
    )
}
```

---

## 六、Z-Index 管理

### 6.1 层级规则

```go
// framework/layout/zindex.go

package layout

// ZOrder 按照 Z-Index 排序
func ZOrder(children []VNode) []VNode {
    result := make([]VNode, len(children))
    copy(result, children)

    sort.Slice(result, func(i, j int) bool {
        zi := getZIndex(result[i])
        zj := getZIndex(result[j])
        return zi < zj
    })

    return result
}

func getZIndex(node VNode) int {
    if elem, ok := node.(*VNodeElement); ok {
        if props, ok := elem.Props.(AbsoluteProps); ok {
            return props.ZIndex
        }
    }
    return 0
}
```

### 6.2 渲染顺序

Z-Index 决定绘制顺序：
- 低值先绘制（在底层）
- 高值后绘制（在上层）

---

## 七、性能优化

### 7.1 缓存计算结果

```go
// framework/layout/absolute_cache.go

type AbsoluteCache struct {
    lastParentSize Size
    lastResult     Box
    dirty          bool
}

func (c *AbsoluteCache) Invalidate() {
    c.dirty = true
}
```

### 7.2 避免频繁计算

只有以下情况需要重新计算：
1. 父容器尺寸变化
2. 位置/尺寸属性变化
3. 子元素尺寸变化

---

## 八、边界情况处理

### 8.1 超出父容器

```go
// 允许超出父容器边界
// 这是 Absolute 的特性
```

### 8.2 负值位置

```go
// 允许负值，会向左/上偏移
ui.Absolute(content).
    Left(-5).  // 向左偏移 5 像素
    Top(-2),   // 向上偏移 2 像素
```

### 8.3 尺寸冲突

```go
// width + left + right 超过父容器宽度时的处理
// 优先级：left > right > width
```

---

## 九、实施计划

### 阶段 1: 基础实现

- [ ] 实现 Position 类型
- [ ] 实现 AbsoluteProps
- [ ] 实现基础 API

### 阶段 2: 布局算法

- [ ] 实现位置解析
- [ ] 实现尺寸解析
- [ ] 实现锚点支持
- [ ] 实现 Z-Index 排序

### 阶段 3: 集成测试

- [ ] 编写单元测试
- [ ] 创建 Badge 示例
- [ ] 创建 Modal 示例
- [ ] 创建 Tooltip 示例

---

## 十、测试策略

```go
// framework/layout/absolute_test.go

func TestAbsoluteTopLeft(t *testing.T) {
    abs := layout.Absolute(
        ui.Box().Width(10).Height(5),
    ).Top(0).Left(0).Build()

    box := abs.Layout(100, 100, 80, 20)
    assert.Equal(t, 100, box.X)
    assert.Equal(t, 100, box.Y)
    assert.Equal(t, 10, box.W)
    assert.Equal(t, 5, box.H)
}

func TestAbsoluteCenter(t *testing.T) {
    abs := layout.Absolute(
        ui.Box().Width(10).Height(5),
    ).Anchor(layout.AnchorCenter).Build()

    box := abs.Layout(0, 0, 80, 20)
    assert.Equal(t, 35, box.X) // (80-10)/2
    assert.Equal(t, 7, box.Y)  // (20-5)/2
}

func TestAbsolutePercent(t *testing.T) {
    abs := layout.Absolute(
        ui.Box().Width(20).Height(10),
    ).LeftPercent(10).TopPercent(20).Build()

    box := abs.Layout(0, 0, 100, 50)
    assert.Equal(t, 10, box.X)  // 100 * 0.1
    assert.Equal(t, 10, box.Y)  // 50 * 0.2
}
```

---

## 十一、与其他布局的组合

### 11.1 Absolute + Stack

```go
// Stack 提供叠加容器，Absolute 控制精确位置
ui.Stack(
    ui.Box().Child(ui.Text("Background")),
    ui.Absolute(
        ui.Box().Child(ui.Text("Foreground")),
    ).Top(2).Left(2),
)
```

### 11.2 Absolute + Grid

```go
// Grid 单元格内使用 Absolute
ui.UICell(0, 0,
    ui.Box().Child(
        ui.Absolute(
            ui.Box().Child(ui.Text("Overlay")),
        ).Top(0).Right(0),
    ),
)
```

### 11.3 Absolute + Layer

```go
// Layer 决定渲染层级，Absolute 控制位置
ui.Layer(ui.LayerOverlay, "tooltip",
    ui.Absolute(
        ui.Box().Child(ui.Text("Tooltip")),
    ).Bottom(2).Left(5),
)
```

---

**文档版本**: v1.0
**最后更新**: 2026-01-31
