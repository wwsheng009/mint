# Mint UI 声明式框架开发指南

本文档详细说明在 Mint UI 声明式框架中开发需要遵循的规则、最佳实践、常见问题排查方法和解决问题的思路。

---

## 目录

1. [架构概述](#1-架构概述)
2. [核心规则：必须遵守](#2-核心规则必须遵守)
3. [禁止事项：绝对不能做](#3-禁止事项绝对不能做)
4. [最佳实践](#4-最佳实践)
5. [常见问题排查](#5-常见问题排查)
6. [调试技巧](#6-调试技巧)
7. [案例分析](#7-案例分析)

---

## 1. 架构概述

### 1.1 分层架构

```
┌─────────────────────────────────────────────────────────┐
│                    ui/ (声明式层)                        │
│  VNode, Hooks, ComponentFunc, declarativeRoot           │
├─────────────────────────────────────────────────────────┤
│                 framework/ (应用层)                      │
│  App, Component, Event, Form, Input                     │
├─────────────────────────────────────────────────────────┤
│                  runtime/ (运行时层)                     │
│  Paint, Buffer, Platform, Style                         │
└─────────────────────────────────────────────────────────┘
```

### 1.2 核心概念

| 概念 | 说明 | 位置 |
|------|------|------|
| **VNode** | 虚拟节点，描述 UI 结构 | `ui/vnode.go` |
| **Hooks** | 状态管理（useState 等） | `ui/hooks.go` |
| **ComponentFunc** | 函数式组件 `func() VNode` | `ui/types.go` |
| **declarativeRoot** | 桥接声明式和命令式框架 | `ui/app.go` |

### 1.3 渲染周期

```
用户交互 → HandleEvent → 状态更新 → MarkDirty → Paint → VNode 树 → Buffer → 终端
```

---

## 2. 核心规则：必须遵守

### 2.1 Hooks 规则

#### 规则 1：Hooks 必须在组件函数顶层调用

```go
// ✅ 正确：在函数顶层调用
func Counter() ui.VNode {
    count, setCount, getCount := ui.UseStateInt(0)
    // ...
}

// ❌ 错误：在条件语句中调用
func Counter() ui.VNode {
    if someCondition {
        count, setCount, _ := ui.UseStateInt(0) // PANIC!
    }
}

// ❌ 错误：在循环中调用
func Counter() ui.VNode {
    for i := 0; i < 3; i++ {
        _, _, _ = ui.UseStateInt(i) // PANIC!
    }
}
```

#### 规则 2：Hooks 调用顺序必须一致

每次渲染时，Hooks 必须以相同的顺序调用。框架通过 `HookValidator` 验证这一点。

```go
// ✅ 正确：顺序一致
func MyComponent() ui.VNode {
    name, setName, _ := ui.UseStateString("")
    age, setAge, _ := ui.UseStateInt(0)
    // 每次渲染都是 string → int
}

// ❌ 错误：顺序不一致
func MyComponent() ui.VNode {
    if showName {
        name, _, _ := ui.UseStateString("") // 有时调用，有时不调用
    }
    age, _, _ := ui.UseStateInt(0)
}
```

### 2.2 闭包规则

#### 规则 3：事件处理器中使用 `getValue()` 获取最新状态

**这是最重要的规则之一！**

```go
// ✅ 正确：使用 getCount() 获取最新值
func Counter() ui.VNode {
    count, setCount, getCount := ui.UseStateInt(0)
    
    increment := func() {
        setCount(getCount() + 1) // 使用 getCount() 获取最新值
    }
    // ...
}

// ❌ 错误：直接使用捕获的 count 值
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)
    
    increment := func() {
        setCount(count + 1) // count 是闭包创建时的值，永远是旧值！
    }
    // ...
}
```

**原因解释**：
- 闭包在创建时捕获变量的值（对于基本类型）
- 按钮对象在 `Paint` 时创建并缓存
- 事件触发时，闭包中的值已经过时
- `getCount()` 直接从 Hook 存储中读取当前值

### 2.3 组件规则

#### 规则 4：组件函数必须是纯函数

```go
// ✅ 正确：纯函数，相同状态产生相同输出
func Counter() ui.VNode {
    count, _, _ := ui.UseStateInt(0)
    return ui.Text(fmt.Sprintf("Count: %d", count))
}

// ❌ 错误：有副作用
var globalCounter = 0
func Counter() ui.VNode {
    globalCounter++ // 副作用！
    return ui.Text(fmt.Sprintf("Count: %d", globalCounter))
}
```

#### 规则 5：不要在组件中直接修改状态

```go
// ✅ 正确：通过 setter 修改状态
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)
    onClick := func() { setCount(count + 1) }
    // ...
}

// ❌ 错误：直接修改变量
func Counter() ui.VNode {
    count, _, _ := ui.UseStateInt(0)
    onClick := func() { count++ } // 不会触发重新渲染！
    // ...
}
```

---

## 3. 禁止事项：绝对不能做

### 3.1 禁止在 HandleEvent 中调用组件函数

```go
// ❌ 绝对禁止
func (d *declarativeRoot) HandleEvent(ev Event) bool {
    vnode := d.appFn() // 会导致 Hook 验证失败！
    // ...
}
```

**原因**：`appFn()` 内部调用 Hooks，会被验证器检测为额外的 Hook 调用。

### 3.2 禁止在渲染过程中修改状态

```go
// ❌ 绝对禁止
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)
    setCount(count + 1) // 在渲染时调用 setter！无限循环！
    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

### 3.3 禁止跨组件共享 Hook 上下文

```go
// ❌ 绝对禁止
var sharedCtx *ComponentContext

func ComponentA() ui.VNode {
    sharedCtx = getCurrentContext() // 保存上下文
    // ...
}

func ComponentB() ui.VNode {
    setCurrentContext(sharedCtx) // 使用其他组件的上下文！
    // ...
}
```

### 3.4 禁止在 goroutine 中直接调用 setter

```go
// ❌ 危险：可能导致竞态条件
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)
    
    go func() {
        time.Sleep(time.Second)
        setCount(count + 1) // 在 goroutine 中调用！
    }()
    // ...
}

// ✅ 如果必须使用 goroutine，使用 channel 或 sync 机制
```

---

## 4. 最佳实践

### 4.1 状态管理

#### 使用类型安全的 Hooks

```go
// ✅ 推荐：使用类型安全版本
count, setCount, getCount := ui.UseStateInt(0)
name, setName, getName := ui.UseStateString("")
visible, setVisible, getVisible := ui.UseStateBool(false)

// ⚠️ 避免：使用通用版本（需要类型断言）
value, setValue := ui.useState(0)
count := value.(int) // 需要类型断言
```

#### 状态粒度

```go
// ✅ 推荐：细粒度状态
func Form() ui.VNode {
    name, setName, _ := ui.UseStateString("")
    email, setEmail, _ := ui.UseStateString("")
    // 每个字段独立更新
}

// ⚠️ 避免：粗粒度状态（除非真的需要）
type FormData struct { Name, Email string }
func Form() ui.VNode {
    data, setData, _ := ui.UseState(FormData{})
    // 任何字段变化都需要创建新对象
}
```

### 4.2 事件处理

#### 提取事件处理器

```go
// ✅ 推荐：提取为命名函数，提高可读性
func Counter() ui.VNode {
    count, setCount, getCount := ui.UseStateInt(0)
    
    handleDecrement := func() {
        setCount(getCount() - 1)
    }
    handleIncrement := func() {
        setCount(getCount() + 1)
    }
    
    return ui.HStack(
        ui.ButtonBuilder("-").OnClick(handleDecrement).Build(),
        ui.ButtonBuilder("+").OnClick(handleIncrement).Build(),
    )
}
```

### 4.3 组件组合

#### 使用布局组件

```go
// ✅ 推荐：使用语义化布局
func App() ui.VNode {
    return ui.VStack(
        ui.Header(),
        ui.Content(),
        ui.Footer(),
    ).Gap(1).Padding(1, 2, 1, 2)
}

// ⚠️ 避免：手动计算位置
```

#### 组件拆分

```go
// ✅ 推荐：拆分为小组件
func Header() ui.VNode {
    return ui.NewTextBuilder("My App").Bold(true).Build()
}

func Counter() ui.VNode {
    // 计数器逻辑
}

func App() ui.VNode {
    return ui.VStack(Header(), Counter())
}
```

### 4.4 样式

```go
// ✅ 推荐：使用 Builder 模式
ui.NewTextBuilder("Hello").
    FgColor("cyan").
    Bold(true).
    Build()

ui.ButtonBuilder("Click").
    FgColor("white").
    BgColor("blue").
    Build()
```

---

## 5. 常见问题排查

### 5.1 Panic: "Hook call count exceeds first render"

**症状**：
```
panic: [Hooks Error] Hook call count exceeds first render in component 'App': 
expected 1 hooks, got 2
```

**原因**：
1. 在 `HandleEvent` 或其他地方额外调用了 `appFn()`
2. 条件语句中调用了 Hooks
3. 循环中调用了 Hooks

**排查步骤**：
1. 检查是否在 `HandleEvent` 中调用了 `d.appFn()`
2. 检查组件函数中是否有条件/循环中的 Hook 调用
3. 使用 `runtime.Caller` 追踪调用栈

**修复**：
```go
// 移除 HandleEvent 中的组件调用
// 确保所有 Hooks 在组件顶层调用
```

### 5.2 状态更新不生效

**症状**：点击按钮后，显示的值没有变化

**原因**：
1. 闭包捕获了旧的状态值
2. 没有调用 `MarkDirty()` 触发重新渲染

**排查步骤**：
1. 检查事件处理器是否使用 `getValue()` 获取最新值
2. 检查 `scheduleRender` 是否正确调用 `MarkDirty()`
3. 添加调试日志确认 setter 被调用

**修复**：
```go
// 使用 getValue() 而不是捕获的值
increment := func() {
    setCount(getCount() + 1) // ✅
    // setCount(count + 1)   // ❌
}
```

### 5.3 按钮点击无响应

**症状**：按 Enter/Space 没有反应

**原因**：
1. `KeyEnter` 或 `KeyTab` 的值在 framework 和 platform 层不匹配
2. 焦点系统没有正确收集按钮
3. `OnClick` 处理器为 nil

**排查步骤**：
1. 验证键值匹配：
   ```go
   fmt.Printf("framework KeyEnter = %d\n", frameworkevent.KeyEnter)
   fmt.Printf("platform KeyEnter = %d\n", platforminput.KeyEnter)
   ```
2. 检查 `collectButtons` 是否收集到按钮
3. 检查按钮的 `OnClick()` 是否返回非 nil 函数

### 5.4 数组越界 Panic

**症状**：
```
panic: runtime error: index out of range [5] with length 3
```

**常见位置**：
- `renderLine` 中访问 buffer
- `swapBuffers` 中复制数据

**排查步骤**：
1. 检查 buffer 大小是否正确初始化
2. 检查渲染时的边界检查

**修复**：
```go
// 添加边界检查
if y < 0 || y >= buffer.Height {
    return
}
if x < 0 || x >= buffer.Width {
    return
}
```

### 5.5 程序无法退出

**症状**：按 q/Esc 没有反应

**原因**：
1. `quit` 通道无缓冲，发送被阻塞
2. `HandleEvent` 返回值不正确
3. 事件没有正确路由

**修复**：
```go
// 使用带缓冲的 quit 通道
quit: make(chan struct{}, 1)
```

---

## 6. 调试技巧

### 6.1 添加调试日志

```go
import "fmt"
import "os"

// 输出到 stderr（不干扰终端 UI）
fmt.Fprintf(os.Stderr, "Debug: count=%d\n", count)

// 或使用结构化日志
log.Printf("HandleEvent: key=%q special=%d", keyEv.Key.Rune, keyEv.Special)
```

### 6.2 验证 Hook 状态

```go
func Counter() ui.VNode {
    count, setCount, getCount := ui.UseStateInt(0)
    
    // 调试：打印 Hook 信息
    ctx := getCurrentContext()
    fmt.Fprintf(os.Stderr, "Hooks: %+v\n", ctx.Hooks)
    
    // ...
}
```

### 6.3 追踪事件流

```go
func (d *declarativeRoot) HandleEvent(ev frameworkevent.Event) bool {
    if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
        fmt.Fprintf(os.Stderr, "KeyEvent: Rune=%q Special=%d Mods=%d\n",
            keyEv.Key.Rune, keyEv.Special, keyEv.Modifiers)
    }
    // ...
}
```

### 6.4 验证渲染周期

```go
func (d *declarativeRoot) Paint(ctx component.PaintContext, buffer *paint.Buffer) {
    fmt.Fprintf(os.Stderr, "Paint called: renderCount=%d\n", d.ctx.RenderCount)
    // ...
}
```

---

## 7. 案例分析

### 7.1 案例：Counter 闭包问题

**问题描述**：
Counter 示例中，按 Enter 只能减少计数，无法增加。

**排查过程**：

1. **验证事件处理**
   - 添加调试日志确认 KeyEnter 被正确识别 ✓

2. **验证键值映射**
   ```go
   // framework KeyEnter = 2
   // platform KeyEnter = 2
   // 匹配正确 ✓
   ```

3. **检查闭包行为**
   - 发现 `setCount(count + 1)` 中的 `count` 是旧值
   - 闭包在 Paint 时创建，捕获当时的 `count` 值
   - HandleEvent 时使用缓存的按钮对象

4. **尝试修复 1（失败）**
   ```go
   // 在 HandleEvent 中重新调用 appFn()
   vnode := d.appFn() // PANIC: Hook 验证失败
   ```

5. **最终修复**
   - 扩展 `UseStateInt` 返回 `getValue` 函数
   - 在事件处理器中使用 `getCount()` 获取最新值

**教训**：
- Go 闭包捕获值类型时，捕获的是值的副本
- React-like 框架中，需要特殊机制获取最新状态
- 不能随意调用组件函数，会破坏 Hook 验证

### 7.2 案例：Buffer 越界

**问题描述**：
渲染时偶发 panic: index out of range

**排查过程**：

1. 定位到 `renderLine` 函数
2. 发现 front/back buffer 大小不一致
3. resize 时只更新了一个 buffer

**修复**：
```go
func (d *declarativeRoot) renderText(node *TextVNode, x, y int, buffer *paint.Buffer) {
    // 添加边界检查
    if y < 0 || y >= buffer.Height {
        return
    }
    // ...
}
```

### 7.3 案例：Quit 信号丢失

**问题描述**：
按 q 键程序不退出

**排查过程**：

1. 确认 HandleEvent 被调用 ✓
2. 确认 `d.app.Quit()` 被调用 ✓
3. 发现 quit 通道无缓冲，main loop 可能没有在接收

**修复**：
```go
// framework/app.go
quit: make(chan struct{}, 1) // 带缓冲
```

---

## 8. 快速参考

### 8.1 文件位置

| 功能 | 文件 |
|------|------|
| VNode 定义 | `ui/vnode.go` |
| Hooks 实现 | `ui/hooks.go` |
| Hook 验证 | `ui/validator.go` |
| 应用入口 | `ui/app.go` |
| 按钮组件 | `ui/button.go` |
| 布局组件 | `ui/layout.go` |
| 文本组件 | `ui/text.go` |

### 8.2 常用命令

```bash
# 编译
go build ./...

# 运行示例
go run examples/declarative/counter.go

# 运行测试
go test ./ui/... -v

# 检查竞态条件
go run -race examples/declarative/counter.go
```

### 8.3 Hook API

```go
// 整数状态
count, setCount, getCount := ui.UseStateInt(0)

// 字符串状态
name, setName, getName := ui.UseStateString("")

// 布尔状态
visible, setVisible, getVisible := ui.UseStateBool(false)
```

### 8.4 VNode 构建

```go
// 文本
ui.Text("Hello")
ui.NewTextBuilder("Hello").FgColor("cyan").Bold(true).Build()

// 按钮
ui.Button("Click")
ui.ButtonBuilder("Click").OnClick(handler).Build()

// 布局
ui.VStack(children...).Gap(1).Padding(1, 2, 1, 2)
ui.HStack(children...).Gap(2)
```

---

## 9. 总结

### 核心原则

1. **Hooks 规则是铁律** - 违反会导致 panic
2. **使用 getValue() 获取最新状态** - 避免闭包捕获问题
3. **组件函数必须是纯函数** - 无副作用
4. **边界检查不可省略** - 防止越界 panic

### 调试思路

1. **先验证假设** - 添加日志确认代码路径
2. **缩小范围** - 二分法定位问题
3. **理解机制** - 了解渲染周期和事件流
4. **查看历史** - 参考类似问题的解决方案

### 问题分类

| 类型 | 典型症状 | 首先检查 |
|------|---------|---------|
| Hook 错误 | panic: Hook call count | 条件/循环中的 Hook |
| 状态不更新 | UI 不变化 | getValue() 使用 |
| 事件无响应 | 点击无效 | HandleEvent 路由 |
| 渲染错误 | panic: index out of range | 边界检查 |
| 程序卡死 | 无法退出 | 通道缓冲 |

---

*文档版本: 1.0*
*最后更新: 2025-01-31*
*适用于: Mint UI 声明式框架*
