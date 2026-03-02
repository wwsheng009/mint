# Border Container Component

Fiber-first 边框容器组件，用于为内容添加装饰性边框。

> ⚠️ **重要通知：API 迁移中**
>
> Border 正在从独立包装组件迁移为容器的原生属性。
> 新代码应优先使用容器的边框方法，而不是包装在 Border 组件中。
> 请参考下方的 [迁移指南](#迁移指南)。

## 目录

- [概述](#概述)
- [快速开始](#快速开始)
- [API 参考](#api-参考)
- [边框样式](#边框样式)
- [边框对布局的影响](#边框对布局的影响)
- [Builder 模式](#builder-模式)
- [迁移指南](#迁移指南) ⚠️
- [示例](#示例)
- [架构说明](#架构说明)

## 概述

Border 组件遵循 **Fiber-first 架构**：
- **VNode** 仅包含声明式描述（边框样式、颜色、标签等）
- **Instance** 是运行时实体，负责测量和渲染边框
- **布局引擎** 自动计算边框占用的空间
- **渲染引擎** 在内容周围绘制边框字符

```
┌─────────────────────────────────────────────────────────────┐
│                    职责分离                                  │
├─────────────────────────────────────────────────────────────┤
│  VNode         │ 仅描述边框样式、颜色、标签                  │
│  Instance      │ 运行时状态，测量尺寸，生成绘制命令          │
│  LayoutBox     │ 自动计算边框占用的额外空间                  │
│  PaintEngine   │ 使用边框字符绘制边框                        │
└─────────────────────────────────────────────────────────────┘
```

## 快速开始

```go
package main

import (
    "github.com/wwsheng009/mint/ui/components/border"
    "github.com/wwsheng009/mint/ui/components/text"
)

func App() rtui.VNode {
    return border.New().
        Label("Title").
        SetChild(text.New("Content inside border")).
        SetWidth(20).
        SetHeight(3)
}
```

## API 参考

### 构造函数

| 函数 | 说明 |
|------|------|
| `New()` | 创建默认单线边框容器 |
| `NewWithStyle(s)` | 创建指定样式的边框容器 |
| `NewBuilder()` | 创建 Builder 用于链式调用 |

### 快捷函数

| 函数 | 说明 |
|------|------|
| `B(child)` | 创建默认边框容器 |
| `Single(child)` | 创建单线边框容器 |
| `Double(child)` | 创建双线边框容器 |
| `Rounded(child)` | 创建圆角边框容器 |
| `Dashed(child)` | 创建虚线边框容器 |
| `WithLabel(label, child)` | 创建带标签的边框容器 |
| `WithColor(color, child)` | 创建指定颜色的边框容器 |

### VNode 方法

| 方法 | 说明 | 默认值 |
|------|------|--------|
| `SetBorderStyle(s)` | 设置边框样式 | `BorderSingle` |
| `SetBorderColor(c)` | 设置边框颜色 | `"blue"` |
| `SetBorderLabel(label)` | 设置边框标签 | `""` |
| `SetWidth(w)` | 设置内容宽度 | `0` (自动) |
| `SetHeight(h)` | 设置内容高度 | `0` (自动) |
| `SetFlex(f)` | 设置弹性因子 | `0` |
| `SetChild(child)` | 设置子元素 | `nil` |
| `Single()` | 设置为单线边框 | - |
| `Double()` | 设置为双线边框 | - |
| `Rounded()` | 设置为圆角边框 | - |
| `Dashed()` | 设置为虚线边框 | - |
| `None()` | 移除边框 | - |
| `Label(label)` | 设置标签（便捷方法） | - |
| `Color(c)` | 设置颜色（便捷方法） | - |

### 辅助函数

| 函数 | 说明 |
|------|------|
| `Box(w, h, style)` | 计算包含边框的总尺寸 |
| `Content(w, h, style)` | 计算可用内容尺寸 |
| `Padding(style)` | 返回边框的水平/垂直内边距 |
| `Offset(style)` | 返回内容的 x,y 偏移量 |

## 边框样式

```
BorderSingle (默认)    BorderDouble         BorderRounded        BorderDashed
┌──────────────┐     ╔══════════════╗     ╭──────────────╮     +--------------+
│              │     ║              ║     │              │     |              |
│   Content    │     ║   Content    ║     │   Content    │     |   Content    |
│              │     ║              ║     │              │     |              |
└──────────────┘     ╚══════════════╝     ╰──────────────╯     +--------------+
Width: 1              Width: 2             Width: 1             Width: 1
```

### 边框宽度

| 样式 | 宽度 | 总尺寸增加 |
|------|------|-----------|
| `BorderNone` | 0 | 0 × 0 |
| `BorderSingle` | 1 | 2 × 2 |
| `BorderDouble` | 2 | 4 × 4 |
| `BorderRounded` | 1 | 2 × 2 |
| `BorderDashed` | 1 | 2 × 2 |

## 边框对布局的影响

### 空间计算

边框会占用额外空间，布局引擎会自动处理：

```
无边框:               单线边框:             双线边框:

                      ┌────────────┐       ╔══════════════╗
                      │            │       ║              ║
┌────────────┐        │  Content   │       ║   Content    ║
│  Content   │        │            │       ║              ║
└────────────┘        └────────────┘       ╚══════════════╝

Size: W × H          Size: (W+2) × (H+2)   Size: (W+4) × (H+4)
Content: (0,0)       Content: (1,1)        Content: (2,2)
```

### 布局计算公式

```go
// 边框宽度
borderWidth := getBorderWidth(borderStyle)
// BorderSingle = 1, BorderDouble = 2, BorderNone = 0

// 水平边距 (左 + 右)
horizontalPadding := borderWidth * 2

// 垂直边距 (上 + 下)
verticalPadding := borderWidth * 2

// 内容偏移 (x, y)
contentOffsetX := borderWidth
contentOffsetY := borderWidth

// 总尺寸
totalWidth := contentWidth + horizontalPadding
totalHeight := contentHeight + verticalPadding
```

### 与 Stack 组件配合

```go
// 带边框的 VStack
stack.NewVStack().
    SetGap(0).
    SetChildrenList([]rtui.VNode{
        border.New().
            Label("Panel A").
            SetWidth(20).
            SetHeight(5).
            SetChild(contentA),
        border.New().
            Label("Panel B").
            SetWidth(20).
            SetHeight(5).
            SetChild(contentB),
    })

// 布局结果:
// ┌───── Panel A ─────┐  <- 高度 5 + 边框 2 = 7
// │                   │
// │     Content A     │
// │                   │
// └───────────────────┘
// ┌───── Panel B ─────┐  <- 高度 5 + 边框 2 = 7
// │                   │
// │     Content B     │
// │                   │
// └───────────────────┘
```

### 嵌套边框

```go
border.New().
    Label("Outer").
    SetChild(
        border.New().
            Label("Inner").
            SetChild(content).
            SetWidth(10).
            SetHeight(3),
    ).
    SetWidth(16).
    SetHeight(7)

// 结果:
// ┌─────── Outer ─────────┐
// │                       │
// │  ┌───── Inner ─────┐  │
// │  │    Content      │  │
// │  └─────────────────┘  │
// │                       │
// └───────────────────────┘
```

## Builder 模式

### 链式调用

```go
border.NewBuilder().
    Label("Settings").
    Color("green").
    Rounded().
    Width(30).
    Height(10).
    Child(settingsContent).
    Build()
```

### 快捷创建

```go
// 单线边框
border.Single(text.New("Simple content"))

// 带标签
border.WithLabel("Important", text.New("Content"))

// 带颜色
border.WithColor("red", text.New("Warning!"))

// 双线边框
border.Double(text.New("Emphasized content"))

// 圆角边框
border.Rounded(text.New("Modern look"))
```

## 迁移指南 ⚠️

### 背景

Border 正在从独立包装组件迁移为容器的原生属性。新架构下：
- **Stack、Grid、Wrap、Absolute** 等容器原生支持边框
- 不再需要将内容包装在 `border.New()` 中
- 边框尺寸自动计算，布局更高效

### 迁移对照表

| 旧 API (包装模式) | 新 API (容器属性) | 说明 |
|------------------|-----------------|------|
| `border.New().Label("T").SetChild(c)` | `stack.SingleBorder("T").SetChildrenList([c])` | Stack 边框 |
| `border.Double().SetChild(c)` | `grid.DoubleBorder().Children(c)` | Grid 边框 |
| `border.Rounded().SetChild(c)` | `wrap.RoundedBorder().Children(c)` | Wrap 边框 |
| `border.New().Color("red").SetChild(c)` | `stack.SingleBorder().SetStyle({...})` | 带颜色 |

### 逐个容器迁移示例

#### Stack (VStack/HStack)

**旧代码**：
```go
border.New().
    Label("Title").
    SetChild(
        stack.NewVStack().
            SetGap(0).
            SetChildrenList([]ui.VNode{
                text.New("Item 1"),
                text.New("Item 2"),
            }),
    ).
    SetWidth(20).
    SetHeight(5)
```

**新代码**：
```go
stack.NewVStack().
    SingleBorder("Title").
    SetGap(0).
    SetWidth(20).  // 注意：现在指包含边框的总宽度
    SetHeight(5).
    SetChildrenList([]ui.VNode{
        text.New("Item 1"),
        text.New("Item 2"),
    })
```

#### Grid

**旧代码**：
```go
border.New().
    Double().
    SetChild(
        grid.New().
            SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
            SetChildrenAuto(cell1, cell2, cell3, cell4),
    ).
    SetWidth(40).
    SetHeight(10)
```

**新代码**：
```go
grid.New().
    DoubleBorder().
    SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
    SetWidth(40).
    SetHeight(10).
    SetChildrenAuto(cell1, cell2, cell3, cell4)
```

#### Wrap

**旧代码**：
```go
border.New().
    Rounded().Label(" Tags ").
    SetChild(
        wrap.New().
            SetWidth(30).
            SetGap(1).
            SetChildrenList(tagItems),
    )
```

**新代码**：
```go
wrap.New().
    RoundedBorder(" Tags ").
    SetWidth(30).
    SetGap(1).
    SetChildrenList(tagItems)
```

#### Absolute

**旧代码**：
```go
border.New().
    SetChild(
        absolute.NewBuilder(child).
            Left(absolute.AbsolutePos(10)).
            Top(absolute.AbsolutePos(5)).
            Build(),
    )
```

**新代码**：
```go
absolute.NewBuilder(child).
    Left(absolute.AbsolutePos(10)).
    Top(absolute.AbsolutePos(5)).
    SingleBorder().  // 直接在 absolute 上设置边框
    Build()
```

### Builder 模式迁移

**旧代码 (Border Builder)**：
```go
border.NewBuilder().
    Label("Settings").
    Color("green").
    Rounded().
    Child(settingsContent).
    Build()
```

**新代码 (Stack Builder)**：
```go
stack.NewBuilder(stack.Column).
    SingleBorder("Settings").
    FgColor("green").
    Children(settingsContent).
    Build()
```

### 边框方法参考

以下方法在所有容器（Stack、Grid、Wrap、Absolute）上通用：

| 方法 | 说明 | 示例 |
|------|------|------|
| `Border(style, label)` | 设置边框样式和标签 | `.Border("single", "Title")` |
| `Bordered(style)` | 设置边框样式（无标签） | `.Bordered("double")` |
| `NoBorder()` | 移除边框 | `.NoBorder()` |
| `SingleBorder(label...)` | 单线边框 | `.SingleBorder("Title")` |
| `DoubleBorder(label...)` | 双线边框 | `.DoubleBorder()` |
| `RoundedBorder(label...)` | 圆角边框 | `.RoundedBorder("Tags")` |
| `DashedBorder(label...)` | 虚线边框 | `.DashedBorder()` |
| `BorderLabel(label)` | 只设置标签 | `.BorderLabel("Info")` |

### 为什么迁移？

1. **API 更简洁**：直接在容器上设置，无需额外包装层
2. **性能更好**：减少一层 Fiber 节点，内存占用更少
3. **自然布局**：边框尺寸自动计算，无需手动调整
4. **统一风格**：与 Padding、Margin 等属性保持一致
5. **向后兼容**：旧 API 继续可用，渐进式迁移

### 完全迁移步骤

1. 第一阶段：新代码使用新 API
2. 第二阶段：核心组件迁移到新 API
3. 第三阶段：更新所有示例和文档
4. 第四阶段：标记旧 API 为 `Deprecated`
5. 第五阶段：完全移除 Border 包装组件

当前处于**第四阶段**：旧 API 已标记为废弃，但仍可使用。

## 示例

### 基础边框

```go
// 简单边框
border.New().
    SetChild(text.New("Hello, World!")).
    SetWidth(15).
    SetHeight(1)
// ┌───────────────┐
// │Hello, World!  │
// └───────────────┘
```

### 带标签的边框

```go
border.New().
    Label(" Configuration ").
    SetChild(configContent).
    SetWidth(30).
    SetHeight(10)
// ┌─── Configuration ────────┐
// │                          │
// │    config content...     │
// │                          │
// └──────────────────────────┘
```

### 彩色边框

```go
border.New().
    Color("red").
    SetChild(text.New("Error: Something went wrong")).
    SetWidth(30).
    SetHeight(1)
// (红色边框)
// ┌──────────────────────────┐
// │Error: Something went wro│
// └──────────────────────────┘
```

### 双线边框（强调）

```go
border.New().
    Double().
    Label(" Important ").
    SetChild(importantContent).
    SetWidth(20).
    SetHeight(5)
// ╔═════ Important ═════╗
// ║                     ║
// ║   important content ║
// ║                     ║
// ╚═════════════════════╝
```

### 圆角边框（现代风格）

```go
border.New().
    Rounded().
    SetChild(modernContent).
    SetWidth(25).
    SetHeight(3)
// ╭─────────────────────────╮
// │                         │
// │    modern content       │
// │                         │
// ╰─────────────────────────╯
```

### 虚线边框（次要区域）

```go
border.New().
    Dashed().
    SetChild(optionalContent).
    SetWidth(20).
    SetHeight(3)
// +--------------------+
// |                    |
// |  optional content  |
// |                    |
// +--------------------+
```

### 组合 Stack + Border

```go
stack.NewHStack().
    SetGap(1).
    SetChildrenList([]rtui.VNode{
        border.New().
            Label("Left Panel").
            SetWidth(15).
            SetHeight(8).
            SetChild(leftContent),
        border.New().
            Label("Right Panel").
            SetWidth(15).
            SetHeight(8).
            SetChild(rightContent),
    })
// ┌── Left Panel ──┐ ┌── Right Panel ──┐
// │                │ │                 │
// │  left content  │ │  right content  │
// │                │ │                 │
// └────────────────┘ └─────────────────┘
```

### 带状态变化的边框

```go
func renderPanel(focused bool) rtui.VNode {
    b := border.New()
    if focused {
        b.Color("yellow").Label("● Active")
    } else {
        b.Color("gray").Label("Inactive")
    }
    b.SetChild(content).
        SetWidth(20).
        SetHeight(5)
    return b
}
```

## 架构说明

### Fiber-first 设计

Border 组件遵循 Fiber-first 架构原则：

1. **VNode 是纯描述**
   - 只包含声明式信息（样式、颜色、标签、尺寸）
   - 无状态、无闭包、无渲染逻辑
   - 可序列化、可缓存

2. **Instance 是运行时实体**
   - 持久化状态
   - 实现布局测量 (`Measure`)
   - 实现渲染 (`Paint`)

3. **布局与渲染分离**
   ```
   VNode → CreateInstance() → Instance
                    ↓
   Fiber → FiberToNodeAdapter → layout.Node
                    ↓
          Engine.Layout() → LayoutBox (自动计算边框空间)
                    ↓
          PaintEngine.PaintLayout() → 调用 Instance.Paint()
                    ↓
          Instance.Paint() → 生成边框绘制命令
   ```

### 布局接口

Border 组件实现的接口：

```go
// layout.Bordered - 提供边框配置
type Bordered interface {
    GetBorder() Border
}

// Measurable - 提供测量能力
type Measurable interface {
    Measure(constraints Constraints) Size
}

// PaintableInstance - 提供渲染能力
type PaintableInstance interface {
    Paint(x, y int) []paint.DrawCmd
}
```

### 边框字符

| 样式 | 左上 | 右上 | 左下 | 右下 | 水平 | 垂直 |
|------|------|------|------|------|------|------|
| Single | `┌` | `┐` | `└` | `┘` | `─` | `│` |
| Double | `╔` | `╗` | `╚` | `╝` | `═` | `║` |
| Rounded | `╭` | `╮` | `╰` | `╯` | `─` | `│` |
| Dashed | `+` | `+` | `+` | `+` | `-` | `│` |

## 性能建议

1. **避免过度嵌套**: 深层嵌套边框会增加布局计算复杂度
2. **使用 BorderNone**: 当不需要边框时，使用 `None()` 而不是创建边框后隐藏
3. **固定尺寸**: 明确设置 `Width` 和 `Height` 可以减少测量开销
4. **复用 VNode**: VNode 是不可变的，可以安全复用

## 迁移指南

### 从旧版 `runtime/ui.Bordered` 迁移

旧版：
```go
// 旧版使用 runtime/ui.Bordered()
ui.Bordered().
    Style("double").
    Label("Title").
    Child(content).
    Build()
```

新版：
```go
// 新版使用 ui/components/border
border.New().
    Double().
    Label("Title").
    SetChild(content)
```

### 主要变化

1. **VNode 不可变**: 使用 Builder 模式或链式调用创建
2. **类型统一**: 使用 `layout.BorderStyle` 和 `layout.Border`
3. **方法命名**: `SetChild()` 替代 `Child()`，避免与 Builder 混淆
4. **尺寸单位**: `Width`/`Height` 表示内容尺寸，边框自动添加
