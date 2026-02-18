# Button Component - Fiber-first Architecture

## 概述

Button 组件是 Fiber-first 架构的参考实现，展示了如何正确分离 VNode（描述）和 Instance（运行期实体）。

### 核心原则

1. **VNode 是纯描述** - 无状态、无闭包、无 Paint 方法
2. **Instance 是运行期实体** - 持有状态、处理交互、执行渲染
3. **Behavior 是可组合逻辑** - 焦点、按压、悬停、禁用等行为可组合
4. **Intent 替代闭包** - 使用结构化 Intent 替代回调函数

## 架构

`
┌─────────────────────────────────────────────────────────────┐
│                    VNode (纯描述)                           │
│  - Props: Label, Variant, Size, FocusStyle, Disabled       │
│  - Intent: pressIntent (替代 OnClick func())               │
│  - 无状态、无闭包、无 Paint                                 │
└──────────────────────────┬──────────────────────────────────┘
                           │ CreateInstance()
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Instance (运行期实体)                    │
│  - State: InteractionState (Focused, Hovered, Pressed...)  │
│  - Paint: 基于状态渲染，不在 VNode 中                      │
│  - Behaviors: Focusable + Pressable + Hoverable + Disable  │
└──────────────────────────┬──────────────────────────────────┘
                           │ 发送
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Intent (结构化意图)                      │
│  - intent.OpenModal(\"settings\")                           │
│  - intent.Navigate(\"/dashboard\")                          │
│  - intent.Click(\"button-1\")                               │
└─────────────────────────────────────────────────────────────┘
`

## 使用方法

### 基础用法

`go
import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/ui/components/button"
)

// 创建按钮
btn := button.New("Click Me")

// 使用 Builder 模式
btn := button.B("Save").
    Primary().
    Large().
    OnPress(intent.OpenModal("settings")).
    Build()
`

### 变体 (Variant)

`go
button.B("Default").Build()      // 默认样式
button.B("Primary").Primary().Build()   // 主要操作
button.B("Secondary").Secondary().Build() // 次要操作
button.B("Danger").Danger().Build()   // 危险操作
button.B("Success").Success().Build() // 成功操作
`

### 尺寸 (Size)

`go
button.B("Small").Small().Build()
button.B("Medium").Medium().Build()  // 默认
button.B("Large").Large().Build()
`

### 焦点样式 (FocusStyle)

`go
button.B("Reverse").FocusStyle(button.FocusStyleReverse).Build()  // 反色（默认）
button.B("Underline").FocusStyle(button.FocusStyleUnderline).Build() // 下划线
button.B("Bracket").FocusStyle(button.FocusStyleBracket).Build()  // 括号
button.B("Bold").FocusStyle(button.FocusStyleBold).Build()    // 加粗
`

### 禁用状态

`go
button.B("Disabled").Disabled(true).Build()
`

### Intent（替代闭包）

`go
// ❌ 旧方式：闭包（不推荐）
button.OnClick(func() {
    setShowModal(true)
})

// ✅ 新方式：Intent（推荐）
button.B("Open").
    OnPress(intent.OpenModal("settings")).
    Build()

// 其他 Intent 示例
button.B("Navigate").OnPress(intent.Navigate("/dashboard")).Build()
button.B("Submit").OnPress(intent.SubmitForm("login", formData)).Build()
button.B("Toggle").OnPress(intent.Toggle("isExpanded")).Build()
`

### 自定义样式

`go
button.B("Custom").
    FgColor("white").
    BgColor("blue").
    PaddingAll(2).
    TextAlign(rtui.AlignCenter).
    Build()
`

## API 参考

### VNode 方法

| 方法 | 说明 |
|------|------|
| New(label) | 创建 Button VNode |
| SetKey(key) | 设置组件 Key |
| SetLabel(label) | 设置标签文本 |
| SetVariant(v) | 设置变体样式 |
| SetSize(s) | 设置尺寸 |
| SetFocusStyle(fs) | 设置焦点样式 |
| SetDisabled(d) | 设置禁用状态 |
| SetIntent(i) | 设置按压 Intent |
| OnPress(i) | 设置按压 Intent（别名） |

### Builder 方法

| 方法 | 说明 |
|------|------|
| B(label) | 创建 Builder |
| Key(k) | 设置 Key |
| Primary() | 设置 Primary 变体 |
| Secondary() | 设置 Secondary 变体 |
| Danger() | 设置 Danger 变体 |
| Success() | 设置 Success 变体 |
| Small() | 设置 Small 尺寸 |
| Medium() | 设置 Medium 尺寸 |
| Large() | 设置 Large 尺寸 |
| FocusStyle(fs) | 设置焦点样式 |
| Disabled(d) | 设置禁用状态 |
| OnPress(i) | 设置按压 Intent |
| Style(s) | 设置样式 |
| Padding(t, r, b, l) | 设置内边距 |
| PaddingAll(p) | 设置四边内边距 |
| TextAlign(a) | 设置文本对齐 |
| Build() | 构建 VNode |
| BuildInstance() | 直接创建 Instance |

### Instance 方法

| 方法 | 说明 |
|------|------|
| Key() | 获取 Key |
| SetFocus(f) | 设置焦点状态 |
| HasFocus() | 是否聚焦 |
| IsDisabled() | 是否禁用 |
| HandleAction(t, p) | 处理动作 |
| Paint(x, y) | 渲染 |
| MarkDirty() | 标记为脏 |

## 状态模型

### InteractionState

`go
type InteractionState struct {
    Focused  bool  // 焦点状态
    Hovered  bool  // 悬停状态
    Pressed  bool  // 按压状态
    Disabled bool  // 禁用状态
    Active   bool  // 激活状态
}
`

### 状态优先级

渲染时状态优先级：Disabled > Pressed > Focused > Hovered > Normal

## Behavior 组合

Button 通过 Behavior 组合实现交互逻辑：

`go
// Instance 内部组合了 4 个 Behavior
behaviors: NewBehaviorList(
    &FocusableBehavior{},   // 处理 Focus/Blur
    &PressableBehavior{},   // 处理 Press/Release
    &HoverableBehavior{},   // 处理 MouseEnter/MouseLeave
    &DisableableBehavior{}, // 处理 Enable/Disable
)
`

### FocusableBehavior

- 处理 Focus / Blur 动作
- 更新 state.Focused
- 发送 intent.Focus / intent.Blur

### PressableBehavior

- 处理 Press / Release 动作
- 更新 state.Pressed
- Release 时发送配置的 Intent

### HoverableBehavior

- 处理 MouseEnter / MouseLeave 动作
- 更新 state.Hovered

### DisableableBehavior

- 处理 Enable / Disable 动作
- 更新 state.Disabled
- 禁用时阻止其他交互

## 完整示例

`go
package main

import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/ui/components/button"
)

func main() {
    // 创建 Intent Runtime
    rt := intent.NewRuntime()
    
    // 注册 Intent 处理器
    intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i intent.OpenModalIntent) intent.IntentResult {
        ctx.SetState("activeModal", i.ModalID)
        ctx.ScheduleUpdate()
        return intent.HandledResult()
    })
    
    // 创建按钮
    openBtn := button.B("Open Settings").
        Key("open-settings").
        Primary().
        Large().
        OnPress(intent.OpenModal("settings")).
        PaddingAll(2).
        Build()
    
    closeBtn := button.B("Close").
        Key("close-btn").
        Secondary().
        OnPress(intent.CloseModal("settings")).
        Build()
    
    // 使用按钮
    _ = openBtn
    _ = closeBtn
}
`

## 迁移指南

### 从旧 Button 迁移

`go
// 旧代码
btn := button.NewButton("Click").
    OnClick(func() {
        doSomething()
    }).
    Build()

// 新代码
btn := button.B("Click").
    OnPress(intent.SetState("clicked", true)).
    Build()

// 或注册自定义 Intent
intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i DoSomethingIntent) intent.IntentResult {
    doSomething()
    return intent.HandledResult()
})

btn := button.B("Click").
    OnPress(DoSomethingIntent{}).
    Build()
`

## 文件结构

`
button/
├── vnode.go        # VNode 描述（纯声明）
├── instance.go     # Instance 运行期实体（状态 + 渲染）
├── builder.go      # Builder 流式 API
├── button_test.go  # 单元测试
└── README.md       # 本文档
`

## 设计参考

本组件遵循以下文档设计：

- docs/fiber/fiber_first/fiber_paint.md - Paint 架构
- docs/fiber/fiber_first/fiber_button.md - Button 设计原则
- docs/fiber/fiber_first/fiber_intent.md - Intent 系统
