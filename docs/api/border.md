# Border API 参考

本文档介绍 Mint UI 框架中容器边框属性的 API 使用方法。

## 概述

边框（Border）现已成为所有布局容器的**原生属性**，不再需要通过包装组件来实现。容器组件直接支持设置边框样式、标签和颜色。

### 支持的容器组件

- **Stack**：`stack.NewVStack()`, `stack.NewHStack()`
- **Grid**：`grid.New()`
- **Wrap**：`wrap.New()`
- **Absolute**：`absolute.New()`

---

## 核心方法

### 边框样式设置方法

以下方法返回容器 VNode 本身，支持链式调用：

| 方法 | 描述 | 返回值 |
|------|------|--------|
| `Border(style, label string)` | 设置边框样式和标签 | `*VNode` |
| `Bordered(style string)` | 设置边框样式（无标签） | `*VNode` |
| `NoBorder()` | 移除边框 | `*VNode` |
| `SingleBorder(label ...string)` | 单线边框 | `*VNode` |
| `DoubleBorder(label ...string)` | 双线边框 | `*VNode` |
| `RoundedBorder(label ...string)` | 圆角边框 | `*VNode` |
| `DashedBorder(label ...string)` | 虚线边框 | `*VNode` |
| `BorderLabel(label string)` | 仅设置边框标签（保留当前样式）| `*VNode` |
| `BorderColor(color string)` | 设置边框颜色 | `*VNode` |
| `SetBorderColor(color style.Color)` | 设置边框颜色（使用 style.Color） | `*VNode` |

---

## 边框样式

支持以下边框样式：

| 样式名称 | 描述 | 视觉效果 |
|----------|------|----------|
| `single` | 单线边框（默认） | `┌─┐ │ └─┘` |
| `double` | 双线边框 | `╔═╗ ║ ╚═╝` |
| `rounded` | 圆角边框 | `╭─╮ │ ╰─╯` |
| `dashed` | 虚线边框 | `+-+ | +-+` |
| `none` | 无边框 | 无 |


---

## 边框颜色

边框颜色通过 `BorderColor()` 或 `SetBorderColor()` 方法设置。

### 支持 4 种颜色格式

```go
// ANSI 标准颜色名
stack.NewVStack().BorderColor("green")

// ANSI 颜色代码 (0-255)
stack.NewVStack().BorderColor("#2")

// 亮度变体
stack.NewVStack().BorderColor("bright-red")

// 标准 16 色
// black, red, green, yellow, blue, magenta, cyan, white
// bright-black, bright-red, bright-green, bright-yellow
// bright-blue, bright-magenta, bright-cyan, bright-white
```

### 常用颜色

```go
// 成功/信息
stack.NewVStack().BorderColor("green")
stack.NewVStack().BorderColor("cyan")

// 警告/错误
stack.NewVStack().BorderColor("yellow")
stack.NewVStack().BorderColor("red")

// 高亮
stack.NewVStack().BorderColor("blue")
stack.NewVStack().BorderColor("magenta")

// 中性色
stack.NewVStack().BorderColor("gray")
stack.NewVStack().BorderColor("white")
```


---

## API 使用示例

### Stack 容器

```go
import "github.com/wwsheng009/mint/ui/components/stack"

// 基本用法 - 单线边框
stack.NewVStack().SingleBorder().SetChildren(...)

// 双线边框带标签
stack.NewVStack().DoubleBorder(" Settings ").SetChildren(...)

// 圆角边框带颜色
stack.NewHStack().
    RoundedBorder(" Actions ").
    BorderColor("cyan").
    SetGap(1).
    SetChildren(...)

// 设置边框颜色（style.Color 类型）
stack.NewVStack().
    SingleBorder().
    SetBorderColor(style.Color("#ff0000")).
    SetChildren(...)

// 无边框
stack.NewVStack().NoBorder().SetChildren(...)
```

### Grid 容器

```go
import "github.com/wwsheng009/mint/ui/components/grid"

// Grid 边框
grid.New(
    grid.WithColumns(3),
    grid.WithRows(2),
).
    SingleBorder(" Data Grid ").
    BorderColor("blue").
    SetChildren(...)

// 使用 Builder API
grid.NewBuilder().
    Columns(3).
    Rows(2).
    Bordered("double").
    Build()
```

### Wrap 容器

```go
import "github.com/wwsheng009/mint/ui/components/wrap"

// Wrap 边框
wrap.New().
    SetWidth(40).
    RoundedBorder(" Tags ").
    BorderColor("magenta").
    SetGap(1).
    SetChildren(...)

// Builder API
wrap.NewBuilder().
    Width(40).
    Wrap("rounded").
    BorderColor("magenta").
    Build()
```

### Absolute 容器

```go
import "github.com/wwsheng009/mint/ui/components/absolute"

// Absolute 边框
absolute.New(content).
    Left(10).
    Top(5).
    DoubleBorder(" Popup ").
    BorderColor("yellow")

// Builder API
absolute.NewBuilder(content).
    Left(10).Top(5).
    Bordered("double").
    Build()
```

---

## 边框对布局的影响

边框会增加容器的视觉尺寸，计算方式如下：

| 边框样式 | 宽度增加 | 高度增加 |
|----------|----------|----------|
| Single / Rounded / Dashed | +2 | +2 |
| Double | +4 | +4 |

边框占用的是**内边距**空间（padding），不会改变组件自身的 width/height 设置。

例如：
```go
// 设置宽度为 30 的容器
stack.NewVStack().
    SetWidth(30).
    SingleBorder()  // 总宽度 = 30 + 2 = 32

// 设置宽度为 30 且有双线边框的容器  
stack.NewVStack().
    SetWidth(30).
    DoubleBorder()  // 总宽度 = 30 + 4 = 34
```


---

## Builder API 支持

大部分容器组件的 Builder API 也支持边框设置：

### Stack Builder

```go
stack.NewBuilder().
    Direction(stack.Vertical).
    Border("single", "Title").           // 设置边框样式和标签
    Bordered("rounded").                 // 设置边框样式
    NoBorder().                          // 移除边框
    BorderColor("green").                // 设置边框颜色
    SingleBorder("Label").               // 单线边框
    DoubleBorder("Label").               // 双线边框
    RoundedBorder("Label").              // 圆角边框
    DashedBorder("Label").               // 虚线边框
    Build()
```

### Grid Builder

```go
grid.NewBuilder().
    Columns(3).
    Rows(2).
    Border("single", "Grid").
    Bordered("double").
    NoBorder().
    BorderColor("blue").
    Build()
```

### Wrap Builder

```go
wrap.NewBuilder().
    Width(40).
    Wrap("rounded").                     // 边框样式
    BorderColor("magenta").
    Build()
```

### Absolute Builder

```go
absolute.NewBuilder(content).
    Left(10).Top(5).
    Bordered("double").
    BorderColor("yellow").
    Build()
```


---

## 迁移指南

### 从旧版 Border 包装组件迁移

旧版设计（已弃用）：
```go
// ❌ 旧版：使用 Border 包装器
import "github.com/wwsheng009/mint/ui/components/border"

border.New(
    stack.NewVStack().SetChildren(...),
    "single",
    "Settings",
    "blue",
)
```

新版设计（推荐）：
```go
// ✅ 新版：使用原生边框属性
import "github.com/wwsheng009/mint/ui/components/stack"

stack.NewVStack().
    SingleBorder(" Settings ").
    BorderColor("blue").
    SetChildren(...)
```

### 迁移差异对比

| 特性 | 旧版（Border 包装组件） | 新版（原生属性） |
|------|------------------------|------------------|
| 代码复杂度 | 需要嵌套包装 | 直接链式调用 |
| 布局计算 | 额外布局节点 | 无额外节点 |
| 性能 | 更慢（额外组件） | 更快（直接属性） |
| 代码可读性 | 低（嵌套结构） | 高（线性调用） |


---

## 完整示例：对话框

```go
package main

import (
    "github.com/wwsheng009/mint/ui/components/stack"
    "github.com/wwsheng009/mint/ui/components/button"
    "github.com/wwsheng009/mint/ui/components/text"
    "github.com/wwsheng009/mint/framework/component"
    "github.com/wwsheng009/mint/runtime/paint"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
    // 创建一个带边框的确认对话框
    dialog := stack.NewVStack().
        DoubleBorder(" Confirm ").
        BorderColor("yellow").
        SetGap(1).
        SetPadding(2).
        SetWidth(40).
        SetChildrenList([]rtui.VNode{
            text.New("Delete this file?").Left(),
            
            stack.NewHStack().
                SetGap(2).
                SetAlignment(stack.AlignRight).  // 按钮右对齐
                SetChildrenList([]rtui.VNode{
                    button.New("Yes").SetVariant(button.VariantDanger),
                    button.New("No"),
                }),
        })

    // 渲染到 buffer
    buf := paint.NewBuffer(60, 10)
    ctx := component.PaintContext{
        Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 10},
        AvailableWidth:  60,
        AvailableHeight: 10,
    }
    dialog.Paint(ctx, buf)

    // 输出（带颜色）
    fmt.Print(buf.String())
}
```


---

## 完整示例：彩色消息框

```go
package main

import (
    "github.com/wwsheng009/mint/ui/components/stack"
    "github.com/wwsheng009/mint/ui/components/text"
    "github.com/wwsheng009/mint/framework/component"
    "github.com/wwsheng009/mint/runtime/paint"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

type MessageType int

const (
    MessageSuccess MessageType = iota
    MessageError
    MessageWarning
    MessageInfo
)

func createMessageBox(title, content string, msgType MessageType) rtui.VNode {
    var borderStyle string
    var borderColor string

    switch msgType {
    case MessageSuccess:
        borderStyle = "rounded"
        borderColor = "green"
    case MessageError:
        borderStyle = "single"
        borderColor = "red"
    case MessageWarning:
        borderStyle = "single"
        borderColor = "yellow"
    case MessageInfo:
        borderStyle = "rounded"
        borderColor = "cyan"
    }

    return stack.NewVStack().
        Border(borderStyle, title).
        BorderColor(borderColor).
        SetPadding(1).
        SetGap(0).
        SetChildrenList([]rtui.VNode{
            text.New(content),
        })
}

func main() {
    // 创建不同类型的消息框
    successBox := createMessageBox(" Success ", "Operation completed successfully!", MessageSuccess)
    errorBox := createMessageBox(" Error ", "Failed to process request", MessageError)
    warningBox := createMessageBox(" Warning ", "Low disk space warning", MessageWarning)
    infoBox := createMessageBox(" Info ", "System is running normally", MessageInfo)

    // 组合展示
    container := stack.NewVStack().SetGap(1).SetChildrenList([]rtui.VNode{
        successBox,
        errorBox,
        warningBox,
        infoBox,
    })

    // 渲染
    buf := paint.NewBuffer(50, 15)
    ctx := component.PaintContext{
        Bounds:          paint.Rect{X: 0, Y: 0, Width: 50, Height: 15},
        AvailableWidth:  50,
        AvailableHeight: 15,
    }
    container.Paint(ctx, buf)
    fmt.Print(buf.String())
}
```


---

## 完整示例：嵌套边框

```go
package main

import (
    "github.com/wwsheng009/mint/ui/components/stack"
    "github.com/wwsheng009/mint/ui/components/text"
    "github.com/wwsheng009/mint/framework/component"
    "github.com/wwsheng009/mint/runtime/paint"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
    // 外层容器
    outer := stack.NewVStack().
        DoubleBorder(" Outer Container ").
        BorderColor("blue").
        SetGap(1).
        SetPadding(1).
        SetWidth(40).
        SetChildrenList([]rtui.VNode{
            // 内层容器 1
            stack.NewVStack().
                RoundedBorder(" Inner Section 1 ").
                BorderColor("green").
                SetChildrenList([]rtui.VNode{
                    text.New("Section 1 content"),
                }),

            // 内层容器 2
            stack.NewVStack().
                SingleBorder(" Inner Section 2 ").
                BorderColor("yellow").
                SetChildrenList([]rtui.VNode{
                    text.New("Section 2 content"),
                }),

            // 内层容器 3
            stack.NewVStack().
                DashedBorder(" Inner Section 3 ").
                BorderColor("magenta").
                SetChildrenList([]rtui.VNode{
                    text.New("Section 3 content"),
                }),
        })

    // 渲染
    buf := paint.NewBuffer(45, 20)
    ctx := component.PaintContext{
        Bounds:          paint.Rect{X: 0, Y: 0, Width: 45, Height: 20},
        AvailableWidth:  45,
        AvailableHeight: 20,
    }
    outer.Paint(ctx, buf)
    fmt.Print(buf.String())
}
```


---

## 常见问题

### Q1: 为什么边框颜色不生效？

确认：
1. 使用了 `buf.String()` 而非不保留颜色的打印方法（如 `utils.PrintBuffer()`）
2. 颜色名称正确（参考 [色彩列表](#常用颜色)）
3. 终端支持 ANSI 颜色转义码

### Q2: 如何动态改变边框颜色？

```go
// 使用 state 管理边框颜色
borderColor := "blue"
stack.NewVStack().
    SingleBorder().
    BorderColor(borderColor).
    SetChildren(...)

// 更新状态后重新渲染
borderColor = "red"
// 注意：需要触发重新渲染
```

### Q3: 边框标签太长怎么办？

标签会自动适应容器的最小宽度。如果内容太窄，可以使用 `SetWidth()` 显式设置宽度：

```go
stack.NewVStack().
    SetWidth(50).  // 确保有足够空间显示长标签
    SingleBorder("This is a very long border label").
    SetChildren(...)
```

### Q4: 可以混合使用不同的边框样式吗？

是的，可以在 UI 中混合使用不同的边框样式，但同一个容器同时只能有一种边框样式：

```go
// ✅ 容器间使用不同样式
stack.NewVStack().SetChildrenList([]rtui.VNode{
    stack.NewVStack().SingleBorder(" A ").SetChildren(...),
    stack.NewVStack().DoubleBorder(" B ").SetChildren(...),
    stack.NewVStack().RoundedBorder(" C ").SetChildren(...),
})

// ❌ 同一容器不能同时有两种样式
stack.NewVStack().
    SingleBorder().    // 这个会被覆盖
    DoubleBorder().    // 只有这个生效
    SetChildren(...)
```


---

## 参考资源

- [Flex 布局（HStack/VStack）](../layout/core_concepts/flex_layout.md)
- [Grid 组件文档](../components/grid/ARCHITECTURE.md)
- [Wrap 组件文档](../layout/core_concepts/wrap_component.md)
- [Absolute 组件代码](../../ui/components/absolute/)
- [Style 颜色参考](../theme/ant_design_quick_reference.md)
- [示例：border_demo](../../examples/fiber_firsts/border_demo/)

---

*最后更新日期：2026-03-02*
