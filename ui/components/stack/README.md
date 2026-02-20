# Stack Layout Components

Fiber-first 布局组件，提供 `HStack`（水平布局）和 `VStack`（垂直布局）功能。

## 目录

- [概述](#概述)
- [快速开始](#快速开始)
- [API 参考](#api-参考)
  - [构造函数](#构造函数)
  - [布局属性](#布局属性)
  - [对齐方式](#对齐方式)
  - [间距与边距](#间距与边距)
  - [尺寸控制](#尺寸控制)
  - [子元素管理](#子元素管理)
- [布局详解](#布局详解)
  - [主轴对齐 (Align)](#主轴对齐-align)
  - [交叉轴对齐 (CrossAlign)](#交叉轴对齐-crossalign)
  - [间距 (Gap)](#间距-gap)
  - [内边距 (Padding)](#内边距-padding)
  - [交叉轴拉伸 (Stretch)](#交叉轴拉伸-stretch)
  - [弹性空间 (Spacer)](#弹性空间-spacer)
- [Builder 模式](#builder-模式)
- [示例](#示例)
- [架构说明](#架构说明)

## 概述

Stack 组件遵循 **Fiber-first 架构**：
- **VNode** 仅包含声明式描述，无状态、无闭包、无渲染逻辑
- **Instance** 是运行时实体，持久的跨渲染状态
- **布局引擎** 负责计算所有元素的位置和尺寸
- **渲染引擎** 使用 LayoutBox 坐标递归渲染子元素

```
┌─────────────────────────────────────────────────────────────┐
│                    职责分离                                  │
├─────────────────────────────────────────────────────────────┤
│  VNode         │ 仅描述，不可变                              │
│  Instance      │ 运行时状态，持久化                          │
│  LayoutBox     │ 计算所有元素的位置和尺寸                    │
│  PaintEngine   │ 使用 LayoutBox 坐标递归渲染                 │
└─────────────────────────────────────────────────────────────┘
```

## 快速开始

```go
package main

import (
    "github.com/wwsheng009/mint/ui/components/stack"
    "github.com/wwsheng009/mint/ui/components/text"
    "github.com/wwsheng009/mint/ui/components/button"
)

func App() rtui.VNode {
    return stack.NewVStack().
        SetGap(1).
        SetChildrenList([]rtui.VNode{
            text.New("Hello, World!"),
            stack.NewHStack().
                SetGap(2).
                SetChildrenList([]rtui.VNode{
                    button.New("OK"),
                    button.New("Cancel"),
                }),
        })
}
```

## API 参考

### 构造函数

| 函数 | 说明 |
|------|------|
| `NewHStack()` | 创建水平布局容器 |
| `NewVStack()` | 创建垂直布局容器 |
| `New(dir Direction)` | 创建指定方向的布局容器 |
| `H(children ...)` | 快捷创建 HStack |
| `V(children ...)` | 快捷创建 VStack |

### 布局属性

| 方法 | 说明 | 默认值 |
|------|------|--------|
| `SetDirection(dir)` | 设置布局方向 | - |
| `SetAlign(a Align)` | 设置主轴对齐 | `AlignStart` |
| `SetCrossAlign(a Align)` | 设置交叉轴对齐 | `AlignStart` |
| `SetGap(gap int)` | 设置子元素间距 | `0` |
| `SetPadding(t, r, b, l int)` | 设置内边距 | `0, 0, 0, 0` |
| `SetStretchCross(bool)` | 交叉轴拉伸子元素 | `false` |
| `SetWidth(w int)` | 设置固定宽度 | `0` (自动) |
| `SetHeight(h int)` | 设置固定高度 | `0` (自动) |
| `SetFlex(f int)` | 设置弹性因子 | `0` |

### 对齐方式

```go
const (
    AlignStart       // 起点对齐
    AlignCenter      // 居中对齐
    AlignEnd         // 终点对齐
    AlignSpaceBetween // 两端对齐，间距平均分配
    AlignSpaceAround  // 每个子元素两侧间距相等
)
```

### 间距与边距

| 方法 | 说明 |
|------|------|
| `SetGap(n)` | 子元素之间的间距 |
| `SetPadding(top, right, bottom, left)` | 容器内边距 |

### 尺寸控制

| 方法 | 说明 |
|------|------|
| `SetWidth(w)` | 设置固定宽度，`0` 表示自动 |
| `SetHeight(h)` | 设置固定高度，`0` 表示自动 |
| `SetFlex(n)` | 设置弹性因子，用于 Spacer |

### 子元素管理

| 方法 | 说明 |
|------|------|
| `SetChildrenList(children)` | 设置子元素列表 |
| `AddChild(child)` | 添加单个子元素 |

## 布局详解

### 主轴对齐 (Align)

主轴是布局的主要方向：
- **HStack**: 水平方向（从左到右）
- **VStack**: 垂直方向（从上到下）

```
AlignStart (默认)     AlignCenter          AlignEnd           SpaceBetween
┌────────────────┐   ┌────────────────┐   ┌────────────────┐   ┌────────────────┐
│ [A][B][C]      │   │    [A][B][C]   │   │        [A][B][C]│   │[A]    [B]    [C]│
└────────────────┘   └────────────────┘   └────────────────┘   └────────────────┘
```

```go
// 左对齐（默认）
stack.NewHStack().SetAlign(stack.AlignStart)

// 居中对齐
stack.NewHStack().Center()
// 或
stack.NewHStack().SetAlign(stack.AlignCenter)

// 右对齐
stack.NewHStack().SetAlign(stack.AlignEnd)

// 两端对齐
stack.NewHStack().SetAlign(stack.AlignSpaceBetween)
```

### 交叉轴对齐 (CrossAlign)

交叉轴是垂直于主轴的方向：
- **HStack**: 垂直方向（从上到下）
- **VStack**: 水平方向（从左到右）

```
HStack with Height=3:

CrossStart          CrossCenter           CrossEnd
┌────────────────┐   ┌────────────────┐   ┌────────────────┐
│ A  B  C        │   │                │   │                │
│                │   │ A  B  C        │   │                │
│                │   │                │   │ A  B  C        │
└────────────────┘   └────────────────┘   └────────────────┘
```

```go
// 顶部对齐（默认）
stack.NewHStack().SetHeight(3).SetCrossAlign(stack.AlignStart)

// 垂直居中
stack.NewHStack().SetHeight(3).CenterCross()
// 或
stack.NewHStack().SetHeight(3).SetCrossAlign(stack.AlignCenter)

// 底部对齐
stack.NewHStack().SetHeight(3).SetCrossAlign(stack.AlignEnd)
```

### 间距 (Gap)

子元素之间的间距：

```go
// 无间距
stack.NewHStack().SetGap(0).SetChildrenList([]rtui.VNode{
    text.New("[A]"),
    text.New("[B]"),
    text.New("[C]"),
})
// 输出: [A][B][C]

// 间距为 2
stack.NewHStack().SetGap(2).SetChildrenList(...)
// 输出: [A]  [B]  [C]

// 间距为 5
stack.NewHStack().SetGap(5).SetChildrenList(...)
// 输出: [A]     [B]     [C]
```

### 内边距 (Padding)

容器的内边距，顺序为 `[top, right, bottom, left]`：

```go
stack.NewHStack().
    SetPadding(1, 2, 1, 2).  // 上1, 右2, 下1, 左2
    SetChildrenList([]rtui.VNode{
        text.New("Content"),
    })
```

```
┌──────────────────┐
│                  │ ← top: 1
│   ┌────────┐     │
│ L │Content │ R   │ ← left: 2, right: 2
│   └────────┘     │
│                  │ ← bottom: 1
└──────────────────┘
```

### 交叉轴拉伸 (Stretch)

让子元素填满交叉轴：

**HStack.Stretch()**: 子元素高度填满容器

```
┌────────────────┐
│ [A]  [B]  [C]  │ ← 无拉伸
└────────────────┘

┌────────────────┐
│ [A]  [B]  [C]  │
│ [A]  [B]  [C]  │ ← 拉伸后
│ [A]  [B]  [C]  │
└────────────────┘
```

**VStack.Stretch()**: 子元素宽度填满容器

```
无拉伸:              拉伸后:
┌─────┐            ┌────────────────┐
│ [A] │            │      [A]       │
├─────┤            ├────────────────┤
│ [B] │            │      [B]       │
├─────┤            ├────────────────┤
│ [C] │            │      [C]       │
└─────┘            └────────────────┘
```

```go
stack.NewHStack().
    SetHeight(3).
    Stretch().
    SetChildrenList([]rtui.VNode{
        text.New("A"),
        text.New("B"),
        text.New("C"),
    })
```

### 弹性空间 (Spacer)

`Spacer(flex)` 用于占据剩余空间，实现类似 CSS Flexbox 的弹性布局：

```go
// Left | Spacer | Right
stack.NewHStack().
    SetWidth(40).
    SetChildrenList([]rtui.VNode{
        text.New("Left"),
        stack.Spacer(1),  // 占据剩余空间
        text.New("Right"),
    })
// 输出: Left                              Right

// Multiple Spacers
stack.NewHStack().
    SetWidth(40).
    SetChildrenList([]rtui.VNode{
        text.New("A"),
        stack.Spacer(1),
        text.New("B"),
        stack.Spacer(1),
        text.New("C"),
    })
// 输出: A                  B                  C
```

Spacer 的 `flex` 参数决定空间分配比例：

```go
stack.Spacer(1)  // 占据 1 份
stack.Spacer(2)  // 占据 2 份
```

## Builder 模式

Stack 提供了两种 Builder 模式：

### 链式调用

```go
stack.NewBuilder(stack.Row).
    Gap(2).
    Center().
    Children(
        text.New("A"),
        text.New("B"),
        text.New("C"),
    ).
    Build()
```

### 专用 Builder

```go
// HStack Builder
stack.NewHStackBuilder().
    Gap(2).
    Center().
    Children(btn1, btn2, btn3).
    Build()

// VStack Builder
stack.NewVStackBuilder().
    Gap(1).
    Padding(1, 0, 1, 0).
    Children(elem1, elem2, elem3).
    Build()
```

### 快捷函数

```go
// 快速创建
stack.H(text.New("A"), text.New("B"))  // HStack
stack.V(text.New("A"), text.New("B"))  // VStack

// 带间距
stack.RowStack(2, children...)  // HStack with gap
stack.ColStack(1, children...)  // VStack with gap
```

## 示例

### 基础布局

```go
// 水平布局
stack.NewHStack().
    SetGap(2).
    SetChildrenList([]rtui.VNode{
        text.New("Left"),
        text.New("Center"),
        text.New("Right"),
    })

// 垂直布局
stack.NewVStack().
    SetGap(0).
    SetChildrenList([]rtui.VNode{
        text.New("Item 1"),
        text.New("Item 2"),
        text.New("Item 3"),
    })
```

### 按钮行

```go
stack.NewHStack().
    SetGap(2).
    SetChildrenList([]rtui.VNode{
        button.New("OK").SetVariant(button.VariantPrimary),
        button.New("Cancel"),
        button.New("Help"),
    })
```

### 两端对齐按钮

```go
stack.NewHStack().
    SetWidth(40).
    SetAlign(stack.AlignSpaceBetween).
    SetChildrenList([]rtui.VNode{
        button.New("Back"),
        button.New("Next").SetVariant(button.VariantPrimary),
    })
// [ Back ]                       [ Next ]
```

### Spacer 实现左右分布

```go
stack.NewHStack().
    SetWidth(40).
    SetChildrenList([]rtui.VNode{
        button.New("OK"),
        stack.Spacer(1),
        button.New("Cancel"),
    })
// [ OK ]                         [ Cancel ]
```

### 嵌套布局 (Grid)

```go
stack.NewVStack().
    SetGap(0).
    SetChildrenList([]rtui.VNode{
        stack.NewHStack().
            SetGap(1).
            SetChildrenList([]rtui.VNode{
                text.New("[1,1]"),
                text.New("[1,2]"),
                text.New("[1,3]"),
            }),
        stack.NewHStack().
            SetGap(1).
            SetChildrenList([]rtui.VNode{
                text.New("[2,1]"),
                text.New("[2,2]"),
                text.New("[2,3]"),
            }),
        stack.NewHStack().
            SetGap(1).
            SetChildrenList([]rtui.VNode{
                text.New("[3,1]"),
                text.New("[3,2]"),
                text.New("[3,3]"),
            }),
    })
```

### 带内边距的容器

```go
stack.NewVStack().
    SetPadding(1, 3, 1, 3).
    SetChildrenList([]rtui.VNode{
        text.New("┌─ Inner Content ─┐"),
        text.New("│  With Padding   │"),
        text.New("└─────────────────┘"),
    })
```

### 工具栏布局

```go
stack.NewHStack().
    SetWidth(50).
    SetAlign(stack.AlignSpaceBetween).
    CenterCross().
    SetHeight(3).
    SetChildrenList([]rtui.VNode{
        stack.NewHStack().
            SetGap(1).
            SetChildrenList([]rtui.VNode{
                text.New("📁"),
                text.New("File"),
                text.New("📝"),
                text.New("Edit"),
            }),
        stack.NewHStack().
            SetChildrenList([]rtui.VNode{
                button.New("?"),
            }),
    })
```

### 状态栏

```go
stack.NewHStack().
    SetWidth(50).
    SetAlign(stack.AlignSpaceBetween).
    SetChildrenList([]rtui.VNode{
        text.New("Ready"),
        text.New("Ln 1, Col 1"),
    })
```

## 架构说明

### Fiber-first 设计

Stack 组件遵循 Fiber-first 架构原则：

1. **VNode 是纯描述**
   - 只包含声明式信息
   - 无状态、无闭包、无渲染逻辑
   - 可序列化、可缓存

2. **Instance 是运行时实体**
   - 持久化状态
   - 实现布局测量 (`Measure`)
   - 实现渲染 (`Paint` - 但 Stack 是纯布局容器，返回 nil)

3. **布局与渲染分离**
   ```
   VNode → CreateInstance() → Instance
                    ↓
   Fiber → FiberToNodeAdapter → layout.Node
                    ↓
          Engine.Layout() → LayoutBox
                    ↓
          PaintEngine.PaintLayout() → 渲染到 Buffer
   ```

### 布局接口

Stack 相关的布局接口：

```go
// FlexStyleProvider - 容器提供 Flex 布局样式
type FlexStyleProvider interface {
    GetFlexStyle() *FlexStyle
}

// FlexChildProvider - 子节点提供 flex 属性
type FlexChildProvider interface {
    GetFlex() int
}
```

### 类型定义

```go
// Direction - 布局方向
type Direction = rtui.Direction
const (
    Row    = rtui.DirectionRow    // 水平
    Column = rtui.DirectionColumn // 垂直
)

// Align - 对齐方式
type Align = rtui.Align
const (
    AlignStart       = rtui.AlignStart
    AlignCenter      = rtui.AlignCenter
    AlignEnd         = rtui.AlignEnd
    AlignSpaceBetween = rtui.AlignSpaceBetween
    AlignSpaceAround  = rtui.AlignSpaceAround
)
```

## 迁移指南

### 从旧版 Stack 迁移

旧版：
```go
// 旧版使用 components/layout/stack
stack := components.NewStackLayout()
stack.SetDirection(ui.Row)
```

新版：
```go
// 新版使用 ui/components/stack
stack := stack.NewHStack()
```

### 主要变化

1. **VNode 不可变**: 使用 Builder 模式或链式调用创建
2. **无 Paint 逻辑**: 布局容器只负责布局，不参与渲染
3. **类型统一**: 使用 `rtui.Direction` 和 `rtui.Align`
4. **Spacer 支持**: 使用 `stack.Spacer(flex)` 创建弹性空间

## 性能建议

1. **避免深层嵌套**: 虽然支持无限嵌套，但建议控制在合理深度
2. **复用 VNode**: VNode 是不可变的，可以安全复用
3. **使用 Spacer**: Spacer 比多个固定宽度元素更高效
4. **合理设置 Gap**: Gap 在布局时一次性计算，比使用多个 Margin 更高效
