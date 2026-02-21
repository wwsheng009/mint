# Panel API 使用指南

## 目录

- [维度概念](#维度概念)
- [API 对照表](#api-对照表)
- [使用示例](#使用示例)
- [最佳实践](#最佳实践)
- [迁移指南](#迁移指南)

---

## 维度概念

Panel 有两种维度概念：

### 外部维度 (Outer Dimensions)

包含边框的总尺寸。这是 Panel 在布局中占用的实际空间。

```
┌─────────────────────┐
│ Title               │  ← 外部高度 = 内容高度 + 边框占用
├─────────────────────┤
│                     │
│   内部内容区域       │
│                     │  ← 内部高度 = 内容高度（不含边框）
└─────────────────────┘
↑  左边框 ↑ 右边框
      边框内边距 (2 × borderWidth)
```

### 内部维度 (Inner Dimensions)

内容区域的实际尺寸（不含边框边距）。

```
外部宽度 = 内部宽度 + 边框内边距
内部宽度 = 外部宽度 - 边框内边距

边框内边距计算：
- 单线边框 (Single/Rounded): 2 字符
- 双线边框 (Double): 2 字符
- 无边框: 0 字符
```

---

## API 对照表

### 尺寸设置方法

| 场景 | 传统 API | 新 API | 说明 |
|------|---------|--------|------|
| 设置外部宽度 | `Width(20)` | `OuterWidth(20)`, `SetOuterWidth(20)` | 包含边框，向后兼容 |
| 设置外部高度 | `Height(6)` | `OuterHeight(6)`, `SetOuterHeight(6)` | 包含边框，向后兼容 |
| 设置内容宽度 | 无 | `ContentWidth(18)`, `SetContentWidth(18)` | 不含边框，自动调整 |
| 设置内容高度 | 无 | `ContentHeight(4)`, `SetContentHeight(4)` | 不含边框，自动调整 |
| 固定尺寸 | `Width(20).Height(6)` | `Fixed(20, 6)` | 一行设置外部尺寸 |

### 自动尺寸方法

| 场景 | API | 说明 |
|------|-----|------|
| 自动宽度 | `AutoWidth()`, `SetWidth(0)` | 从内容自动测量宽度 |
| 自动高度 | `AutoHeight()`, `SetHeight(0)` | 从内容自动测量高度 |
| 自动尺寸 | `AutoSize()`, `FixedSize(0, 0)` | 从内容自动测量宽高 |
| 固定宽度自动高度 | `FixedWidthAutoHeight(w)` | 宽度固定，高度自适应 |
| 固定高度自动宽度 | `FixedHeightAutoWidth(h)` | 高度固定，宽度自适应 |

### 文本内容方法

| 场景 | VNode API | Builder API | 说明 |
|------|-----------|-------------|------|
| 设置 Wrap 文本 | `SetTextContent(text)` | `WithTextContent(text)` | 自动 Wrap，Auto 尺寸 |
| 设置固定宽度的 Wrap 文本 | `SetWrappedTextContent(text, w)` | `WithWrappedText(text, w)` | 自动调整 Panel 宽度 |
| 设置纯文本（不 Wrap） | `SetPlainContent(text)` | `WithPlainContent(text)` | 自动 Wrap=false |

### 快捷工厂函数

| 场景 | 函数 | 说明 |
|------|------|------|
| 信息面板 | `InfoPanel(title, content)` | 蓝色单线边框 |
| 警告面板 | `WarningPanel(title, content)` | 黄色单线边框 |
| 错误面板 | `ErrorPanel(title, content)` | 红色双线边框 |
| 成功面板 | `SuccessPanel(title, content)` | 绿色双线边框 |
| 文本面板 | `TextPanel(title, content, w)` | 自动 Wrap，指定宽度 |

---

## 使用示例

### 场景 1：固定尺寸

使用外部维度设置固定大小的面板。

```go
import "github.com/wwsheng009/mint/ui/components/panel"
import "github.com/wwsheng009/mint/ui/components/text"

// 传统方式
p1 := panel.New().
    SetWidth(20).
    SetHeight(6).
    SetContent(text.New("Hello"))

// 新方式 - 更清晰
p2 := panel.New().
    SetOuterSize(20, 6).
    SetContent(text.New("Hello"))

// Builder 方式
p3 := panel.NewBuilder().
    Fixed(20, 6).
    Content(text.New("Hello")).
    Build()
```

### 场景 2：固定宽度，自动高度

创建具有固定宽度但高度自适应内容的面板。

```go
// 传统方式 - 需要手动计算
panel.New().
    SetWidth(20).
    SetHeight(0).  // 0 表示自动高度
    SetContent(text.New("Hello\nWorld"))

// 新方式 - 更直观
panel.New().
    FixedWidthAutoHeight(20).
    SetContent(text.New("Hello\nWorld"))

// Builder 方式
panel.NewBuilder().
    Width(20).
    AutoHeight().
    Content(text.New("Hello\nWorld")).
    Build()
```

### 场景 3：设置内容宽度（内部维度）

直接设置内容区域的宽度，Panel 自动计算外部尺寸。

```go
// 传统方式 - 需要知道边框占用
panel.New().
    SetWidth(22).  // 20 内容 + 2 边框
    SetContent(text.New("Hello").Wrap(true))

// 新方式 - 无需手动计算边框
panel.New().
    SetContentWidth(20).  // 直接指定内容宽度
    SetContent(text.New("Hello").Wrap(true))

// Builder 方式
panel.NewBuilder().
    ContentWidth(20).
    Content(text.New("Hello").Wrap(true)).
    Build()
```

### 场景 4：Wrap 文本面板

创建一个自动换行的文本面板。

```go
// 传统方式
panel.New().
    SetWidth(22).  // 20 内容 + 2 边框
    SetHeight(6).
    SetContent(text.New("This is a long text that should wrap").SetWrap(true))

// 新方式 - 一行搞定
panel.New().
    SetWrappedTextContent("This is a long text that should wrap", 20).
    SetTitle("Message")

// Builder 方式
panel.NewBuilder().
    TextPanel("Message", "This is a long text that should wrap", 20).
    Build()
```

### 场景 5：自动尺寸面板

完全根据内容自动调整大小。

```go
// 传统方式
panel.New().
    SetWidth(0).  // 自动宽度
    SetHeight(0). // 自动高度
    SetContent(text.New("Auto-sized content"))

// 新方式 - 更直观
panel.New().
    AutoSize().
    SetContent(text.New("Auto-sized content"))

// 或使用工厂函数
panel.AutoContent(text.New("Auto-sized content"))
```

### 场景 6：带标题和内容的面板

创建标准的带标题面板。

```go
// 使用 Builder 链式调用
panel.NewBuilder().
    Title("Settings").
    OuterSize(30, 10).
    Content(text.New("Settings go here")).
    Rounded().
    Build()

// 使用便利函数
panel.TitledAuto("Settings", text.New("Settings go here"))
```

### 场景 7：状态面板（信息/警告/错误）

快速创建不同类型的状态面板。

```go
info := panel.InfoPanel("Information", "Operation successful")
warning := panel.WarningPanel("Warning", "This might cause issues")
error := panel.ErrorPanel("Error", "Something went wrong")
success := panel.SuccessPanel("Success", "All good!")

// Builder 方式
panel.NewBuilder().
    Title("Info").
    WithBorder(layout.BorderSingle, style.Color("blue")).
    WithTextContent("Operation complete").
    Build()
```

### 场景 8：条件设置

使用 Maybe 方法进行条件设置。

```go
showBorder := true
title := "Optional Title"

panel.NewBuilder().
    MaybeTitle(title).
    MaybeBorder(layout.BorderSingle, style.Color("blue")).
    Content(text.New("Content")).
    Build()
```

---

## 最佳实践

### 1. 优先使用内容维度

当您知道内容应该有多大时，使用 `ContentWidth/Height` 而不是手动计算边框占用。

```go
// ❌ 不推荐 - 需要知道边框内边距
panel.New().SetWidth(22)  // 20 内容 + 2 边框

// ✅ 推荐 - 自动计算
panel.New().SetContentWidth(20)
```

### 2. 使用自动尺寸适应不同内容

让 Panel 根据内容自动调整大小，避免裁剪。

```go
// ❌ 不推荐 - 可能裁剪内容
panel.New().SetSize(20, 5)
panel.SetContent(text.New("Very long text that might be cut off"))

// ✅ 推荐 - 自动调整
panel.New().AutoSize()
panel.SetContent(text.New("Content at any size"))
```

### 3. 使用便捷方法处理常见场景

对于常见模式，使用专门的方法可以减少代码。

```go
// ❌ 可行但冗长
panel.New().
    SetTitle("Error").
    SetBorderStyle(layout.BorderDouble).
    SetBorderColor(style.Color("red")).
    SetContent(text.New("Something went wrong").SetWrap(true))

// ✅ 清晰且简洁
panel.ErrorPanel("Error", "Something went wrong")
```

### 4. Builder 模式用于复杂配置

对于复杂配置，使用 Builder 模式更清晰。

```go
panel.NewBuilder().
    Title("Complex Panel").
    OuterSize(40, 15).
    BorderStyle(layout.BorderDouble).
    BorderColor(style.Color("cyan")).
    Header(text.New("Header")).
    Content(text.New("Main Content").Wrap(true)).
    Footer(text.New("Footer")).
    Build()
```

### 5. VNode 方法用于简单修改

对于简单修改，直接使用 VNode 方法更方便。

```go
// 创建简单修改
panel.New().
    SetContentWidth(20).
    SetTextContent("Hello")

// 而不是
b := panel.NewBuilder()
b.ContentWidth(20)
b.WithTextContent("Hello")
p := b.Build()
```

---

## 迁移指南

### 从旧 API 迁移到新 API

#### 场景 1：手动计算边框的代码

```go
// 旧代码
p := panel.New().
    SetWidth(contentWidth + 2).  // 手动添加边框
    SetHeight(lineCount + 2)

// 新代码
p := panel.New().
    SetContentWidth(contentWidth).
    SetContentSize(lineCount)
```

#### 场景 2：设置 Wrap 文本

```go
// 旧代码
p := panel.New().SetWidth(22)
p.SetContent(text.New(text).SetWrap(true))

// 新代码
p := panel.New().SetWrappedTextContent(text, 20)
```

#### 场景 3：创建状态面板

```go
// 旧代码
errorPanel := panel.New().
    SetTitle("Error").
    SetBorderStyle(layout.BorderDouble).
    SetBorderColor(style.Color("red")).
    SetContent(text.New(msg).SetWrap(true))

// 新代码
errorPanel := panel.ErrorPanel("Error", msg)
```

### 向后兼容性

所有旧 API 仍然可用，新 API 是在旧 API 基础上的增强。

- `Width()` / `Height()` - 仍然工作，等同于 `OuterWidth()` / `OuterHeight()`
- `SetTitle()` / `SetContent()` - 保持不变
- `Builder` 模式 - 完全兼容，只是增加了新方法

---

## 高级技巧

### 1. 使用内部维度查询

```go
p := panel.New().SetOuterSize(20, 6)

// 获取内容区域实际大小
innerWidth, innerHeight := p.GetInnerDimensions()
fmt.Printf("Content area: %dx%d\n", innerWidth, innerHeight)
```

### 2. 根据边框样式调整

```go
// 在不同边框样式间切换时，使用对样式敏感的方法
panel.New().
    SetInnerWidthForStyle(20, layout.BorderDouble).  // 自动计算 Double 边框占用
    SetContent(...)
```

### 3. 使用工具函数计算

```go
// 独立的工具函数可以用于计算
outerWidth := panel.CalculateOuterWidth(20, layout.BorderDouble)
outerHeight := panel.CalculateOuterHeight(5, layout.BorderSingle)
```

### 4. 组合多个 With 方法

```go
// 使用 With 前缀的方法进行可选配置
panel.New().
   WithTitle("Optional Title").
    WithOuterDimensions(40, 15).
    WithBorderStyleAndColor(layout.BorderRounded, style.Color("blue")).
    WithContentText("Content")
```

---

## 完整示例

### 示例：创建一个设置面板

```go
package main

import (
    "github.com/wwsheng009/mint/ui/components/panel"
    "github.com/wwsheng009/mint/ui/components/text"
    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/style"
    "github.com/wwsheng009/mint/runtime/ui"
)

func createSettingsPanel() ui.VNode {
    return panel.NewBuilder().
        Title("Settings").
        OuterSize(40, 15).
        BorderStyle(layout.BorderDouble).
        BorderColor(style.Color("cyan")).
        Header(
            text.New("Configure your preferences").SetWrap(true),
        ).
        Content(
            text.New("Option 1: Enabled\nOption 2: Disabled\nOption 3: Custom").SetWrap(true),
        ).
        Footer(
            text.New("Press Enter to save").SetWrap(true),
        ).
        Build()
}

// 或者使用更简单的方式
func createSettingsPanelSimple() ui.VNode {
    return panel.New().
        SetOuterSize(40, 15).
        SetWrappedTextContent(
            "Option 1: Enabled\nOption 2: Disabled\nOption 3: Custom",
            38,
        ).
        SetBorderStyle(layout.BorderDouble).
        SetBorderColor(style.Color("cyan"))
}
```

---

## 参考资源

- [Panel API 源代码](../../ui/components/panel/)
- [布局系统优化计划](./02-optimization.md)
- [Border 组件文档](../../ui/components/border/)
- [Text 组件文档](../../ui/components/text/)

---

**文档版本**: 1.0  
**最后更新**: 2026-02-22  
**维护者**: Qwen Code
