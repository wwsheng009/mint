# 组件迁移指南

> **目标**: 将 `ui/` 中的组件迁移到 `components/` 各分类目录
> **原则**: 保持功能一致，改进组织结构
> **日期**: 2026-02-01

---

## 目录

1. [API 兼容性承诺](#api-兼容性承诺)
2. [迁移模式](#迁移模式)
3. [基本组件迁移](#基本组件迁移)
4. [表单组件迁移](#表单组件迁移)
5. [布局组件迁移](#布局组件迁移)
6. [验证清单](#验证清单)

---

## API 兼容性承诺

### 核心承诺：声明式组件功能完全保留

重构后，用户可以继续使用声明式方式创建组件，**无需修改现有代码**：

```go
// ✅ 完全支持的写法 (重构前后保持一致)
func Counter() ui.VNode {
    count, setCount := ui.UseState(0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.HStack(
            ui.Button("-").OnClick(func() { setCount(count - 1) }),
            ui.Text(fmt.Sprintf("%d", count)),
            ui.Button("+").OnClick(func() { setCount(count + 1) }),
        ),
        ui.Input("Name").Value(name),
    )
}

// 运行应用
func main() {
    ui.Run(Counter,
        ui.WithWidth(40),
        ui.WithHeight(20),
    )
}
```

### 三种使用方式并存

重构后提供三种使用方式，用户可以根据场景选择：

| 方式 | 适用场景 | 示例 |
|------|---------|------|
| **声明式函数** | 快速构建 UI | `ui.Text("Hello")` |
| **Builder 模式** | 复杂配置 | `ui.Input("Name").Placeholder("Enter name").Value(v)` |
| **直接导入** | 精细控制 | `import "github.com/wwsheng009/mint/components/form"` |

### API 层次结构

```
┌─────────────────────────────────────────────────────────┐
│ 用户代码层 (User Code)                                   │
│                                                          │
│   func App() ui.VNode {                                  │
│       return ui.VStack(                                 │
│           ui.Text("Hello"),                             │
│           ui.Button("Click"),                           │
│       )                                                  │
│   }                                                      │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│ ui/ 公开 API 层 (Shortcut Layer)                         │
│                                                          │
│   func Text(content string) ui.VNode {                  │
│       return basic.Text(content).Build()                │
│   }                                                      │
│                                                          │
│   func VStack(children ...ui.VNode) ui.VNode {          │
│       return layout.VStack(children...)                  │
│   }                                                      │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│ components/ 组件库层 (Component Library)                  │
│                                                          │
│   package basic                                         │
│   func Text(content string) *TextBuilder { ... }        │
│                                                          │
│   package layout                                        │
│   func VStack(children ...ui.VNode) ui.VNode { ... }     │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│ ui/ 核心层 (Core)                                         │
│                                                          │
│   - VNode 接口                                            │
│   - Hooks (UseState, UseEffect, ...)                      │
│   - ComponentFunc 类型                                    │
│   - Run() 入口                                            │
└─────────────────────────────────────────────────────────┘
```

---

---

## 迁移模式

### 模式 1: 直接迁移 (简单组件)

适用于：结构简单、无复杂依赖的组件

```go
// 原文件: ui/text.go

// 步骤 1: 复制到新位置
// cp ui/text.go components/basic/text.go

// 步骤 2: 修改 package 声明
- package ui
+ package basic

// 步骤 3: 更新导入
import (
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
)

// 步骤 4: 保留导出的构造函数
func Text(content string) VNode {
    return NewText(content)
}
```

### 模式 2: 重构迁移 (复杂组件)

适用于：需要拆分的组件

```go
// 原文件: ui/button.go (200+ 行)

// 拆分为:
// components/button/button.go       - 核心组件逻辑
// components/button/builder.go      - Builder 模式
// components/button/public.go       - 公开接口
```

---

## 基本组件迁移

### Text 组件

**源文件**: `ui/text.go`

**目标位置**: `components/basic/text.go`

```go
// components/basic/text.go

package basic

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/runtime/style"
)

// Text 文本组件
type Text struct {
    vnode ui.VNode
    content string
}

// NewText 创建文本组件
func NewText(content string) *Text {
    return &Text{
        vnode:   ui.NewElement("text"),
        content: content,
    }
}

// Build 构建 VNode
func (t *Text) Build() ui.VNode {
    t.vnode.SetProp("content", t.content)
    return t.vnode
}

// TextBuilder 文本构建器
type TextBuilder struct {
    node *Text
}

// Text 创建文本构建器
func Text(content string) *TextBuilder {
    return &TextBuilder{
        node: NewText(content),
    }
}

// Content 设置文本内容
func (b *TextBuilder) Content(content string) *TextBuilder {
    b.node.content = content
    return b
}

// Style 设置样式
func (b *TextBuilder) Style(s style.Style) *TextBuilder {
    b.node.vnode.SetStyle(s)
    return b
}

// Build 构建 VNode
func (b *TextBuilder) Build() ui.VNode {
    return b.node.Build()
}
```

**ui/shortcuts.go 更新**:

```go
// ui/shortcuts.go

import "github.com/wwsheng009/mint/components/basic"

// Text 创建文本组件 (快捷方式)
func Text(content string) ui.VNode {
    return basic.Text(content).Build()
}
```

---

## 表单组件迁移

### Input 组件

**源文件**: `ui/input.go`

**目标位置**: `components/form/input.go`

```go
// components/form/input.go

package form

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/runtime/style"
)

// InputVNode 输入组件
type InputVNode struct {
    *ui.ElementVNode
    value      string
    placeholder string
    readOnly   bool
    disabled   bool
    maxLength  int
    password   bool
    onChange   func(string)
}

// NewInput 创建输入组件
func NewInput() *InputVNode {
    return &InputVNode{
        ElementVNode: ui.NewElement("input"),
        value:        "",
        placeholder:  "",
        readOnly:     false,
        disabled:     false,
        maxLength:    0,
        password:     false,
    }
}

// Input 创建输入构建器
func Input(placeholder string) *InputBuilder {
    return &InputBuilder{
        node: NewInput(),
    }
}

// InputBuilder 输入构建器
type InputBuilder struct {
    node *InputVNode
}

// Value 设置值
func (b *InputBuilder) Value(v string) *InputBuilder {
    b.node.value = v
    return b
}

// Placeholder 设置占位符
func (b *InputBuilder) Placeholder(p string) *InputBuilder {
    b.node.placeholder = p
    return b
}

// OnChange 设置变化回调
func (b *InputBuilder) OnChange(fn func(string)) *InputBuilder {
    b.node.onChange = fn
    return b
}

// Style 设置样式
func (b *InputBuilder) Style(s style.Style) *InputBuilder {
    b.node.SetStyle(s)
    return b
}

// Build 构建 VNode
func (b *InputBuilder) Build() ui.VNode {
    return b.node
}
```

---

## 布局组件迁移

### HStack/VStack 组件

**源文件**: `ui/layout.go`

**目标位置**: `components/layout/stack.go`

```go
// components/layout/stack.go

package layout

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/runtime/style"
)

// Stack 堆叠布局
type Stack struct {
    vnode   ui.VNode
    dir     Direction  // H 或 V
    align   Alignment
    justify Justification
    spacing int
}

type Direction int
const (
    Horizontal Direction = iota
    Vertical
)

// NewStack 创建堆叠布局
func NewStack(dir Direction) *Stack {
    return &Stack{
        vnode:   ui.NewElement("stack"),
        dir:     dir,
        align:   Start,
        justify: Start,
        spacing: 0,
    }
}

// HStack 创建水平堆叠
func HStack(children ...ui.VNode) ui.VNode {
    stack := NewStack(Horizontal)
    stack.vnode.SetChildren(children)
    return stack.vnode
}

// VStack 创建垂直堆叠
func VStack(children ...ui.VNode) ui.VNode {
    stack := NewStack(Vertical)
    stack.vnode.SetChildren(children)
    return stack.vnode
}

// StackBuilder 堆叠布局构建器
type StackBuilder struct {
    node *Stack
}

// Stack 创建堆叠构建器
func Stack(dir Direction) *StackBuilder {
    return &StackBuilder{
        node: NewStack(dir),
    }
}

// Children 设置子元素
func (b *StackBuilder) Children(children ...ui.VNode) *StackBuilder {
    b.node.vnode.SetChildren(children)
    return b
}

// Spacing 设置间距
func (b *StackBuilder) Spacing(n int) *StackBuilder {
    b.node.spacing = n
    return b
}

// Align 设置对齐
func (b *StackBuilder) Align(a Alignment) *StackBuilder {
    b.node.align = a
    return b
}

// Build 构建 VNode
func (b *StackBuilder) Build() ui.VNode {
    return b.node.vnode
}
```

---

## 验证清单

### 迁移后检查

- [ ] 文件已复制到目标位置
- [ ] package 声明已更新
- [ ] 导入路径已更新
- [ ] 组件功能保持一致
- [ ] Builder 模式正常工作
- [ ] 单元测试已更新
- [ ] 示例程序正常运行
- [ ] 文档已同步更新

### 导入路径映射表

| 原导入 | 新导入 |
|--------|--------|
| `ui/text.go` | `components/basic/text.go` |
| `ui/button.go` | `components/button/button.go` |
| `ui/input.go` | `components/form/input.go` |
| `ui/checkbox.go` | `components/form/checkbox.go` |
| `ui/layout.go` | `components/layout/stack.go` |

---

## 注意事项

1. **循环依赖**: components/ 不应依赖 ui/ 的实现细节
2. **类型引用**: VNode 等核心类型仍在 ui/ 包
3. **测试路径**: 测试文件同步迁移到 components/xxx/xxx_test.go
4. **文档更新**: 同步更新 API 文档

---

**文档版本**: v1.0
**最后更新**: 2026-02-01
