# Border Container Component

Fiber-first 边框容器组件，用于为内容添加装饰性边框。

## 目录

- [概述](#概述)
- [快速开始](#快速开始)
- [API 参考](#api-参考)
- [边框样式](#边框样式)
- [边框对布局的影响](#边框对布局的影响)
- [Builder 模式](#builder-模式)
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
