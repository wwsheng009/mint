# Mint UI 测试框架文档

## 目录

- [概述](#概述)
- [设计理念](#设计理念)
- [核心组件](#核心组件)
- [使用指南](#使用指南)
- [API 参考](#api-参考)
- [最佳实践](#最佳实践)
- [示例](#示例)

---

## 概述

Mint UI 测试框架是一个专为声明式 TUI 组件设计的自动化测试工具。它提供了：

- **无头渲染测试**：无需终端即可渲染组件
- **交互模拟**：模拟按钮点击、键盘输入、文本输入等
- **状态验证**：验证组件状态变化
- **完整流程测试**：测试多步骤用户交互流程

### 特性

✅ 自动化测试，无需手动交互
✅ 支持所有组件类型
✅ 链式 API，易于编写测试
✅ 详细的断言信息
✅ 支持状态变化验证

---

## 设计理念

### 1. 组件即函数

测试框架基于 Mint UI 的核心设计理念：**组件是纯函数**。

```go
type ComponentFunc func() VNode
```

每个组件渲染时都会产生一个 VNode 树，测试框架通过捕获这个树来：
- 收集交互元素（按钮、输入框等）
- 验证组件结构
- 模拟用户交互

### 2. 隔离测试环境

测试在一个隔离的环境中运行：
- 不需要真实的终端
- 不启动主事件循环
- 不依赖外部依赖

### 3. 显式状态管理

测试通过直接调用事件处理器来模拟交互：
- 直接调用 `onClick` 处理器
- 直接调用 `onChange` 处理器
- 手动触发状态更新

---

## 核心组件

### ComponentTest

测试工具的核心结构：

```go
type ComponentTest struct {
    t          *testing.T
    root       *declarativeRoot
    rendered   bool
    eventQueue []frameworkevent.Event
}
```

### 创建测试实例

```go
test := NewComponentTest(t, MyComponent)
```

---

## 使用指南

### 基础测试流程

```go
func TestMyComponent(t *testing.T) {
    // 1. 定义组件
    component := func() VNode {
        return VStack(
            ButtonBuilder("Click Me").Build(),
        )
    }

    // 2. 创建测试并渲染
    test := NewComponentTest(t, component).Render()

    // 3. 验证结果
    test.AssertButtonCount(1)
}
```

### 完整的交互测试

```go
func TestButtonClick(t *testing.T) {
    clicked := false

    component := func() VNode {
        return ButtonBuilder("Click Me").
            OnClick(func() { clicked = true }).
            Build()
    }

    test := NewComponentTest(t, component).Render()

    // 验证初始状态
    if clicked != false {
        t.Error("Expected clicked to be false")
    }

    // 模拟点击
    test.ClickButton(0)

    // 验证点击后状态
    if clicked != true {
        t.Error("Expected clicked to be true")
    }
}
```

---

## API 参考

### 创建测试

```go
func NewComponentTest(t *testing.T, componentFunc ComponentFunc) *ComponentTest
```

### 渲染控制

| 方法 | 说明 |
|------|------|
| `Render() *ComponentTest` | 渲染组件并收集交互元素 |

### 交互模拟

| 方法 | 说明 |
|------|------|
| `ClickButton(index int)` | 点击指定索引的按钮 |
| `ClickButtonByLabel(label string)` | 点击指定标签的按钮 |
| `PressKey(key Key)` | 模拟按键 |
| `PressEnter()` | 模拟 Enter 键 |
| `PressTab()` | 模拟 Tab 键 |
| `TypeText(text string)` | 在第一个输入框输入文本 |
| `ToggleCheckbox(index int)` | 切换复选框状态 |
| `SelectOption(index int)` | 选择下拉选项 |

### 断言方法

| 方法 | 说明 |
|------|------|
| `AssertButtonCount(n int)` | 验证按钮数量 |
| `AssertButtonLabel(i int, label string)` | 验证按钮标签 |
| `AssertInputCount(n int)` | 验证输入框数量 |
| `AssertInputValue(text string)` | 验证输入框值 |
| `AssertCheckboxCount(n int)` | 验证复选框数量 |
| `AssertCheckboxChecked(i int, checked bool)` | 验证复选框状态 |

### 获取方法

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `GetButtonCount()` | 获取按钮数量 | int |
| `GetInputCount()` | 获取输入框数量 | int |
| `GetCheckboxCount()` | 获取复选框数量 | int |

---

## 最佳实践

### 1. 测试命名

使用描述性的测试名称：

```go
// 好的命名
func TestModalOpenCloseFlow(t *testing.T)
func TestButtonIncrementCounter(t *testing.T)
func TestInputValidationRejectsEmpty(t *testing.T)

// 避免使用
func TestComponent1(t *testing.T)
func TestFeature(t *testing.T)
```

### 2. 使用链式调用

```go
// 推荐：链式调用，代码简洁
test.Render().
    AssertButtonCount(2).
    AssertButtonLabel(0, "OK").
    AssertButtonLabel(1, "Cancel")
```

### 3. 测试状态变化

```go
func TestStateChange(t *testing.T) {
    var state string

    component := func() VNode {
        text, setText := UseStateString("initial")
        state = text // 捕获状态用于验证
        return ButtonBuilder("Change").
            OnClick(func() { setText("changed") }).
            Build()
    }

    test := NewComponentTest(t, component).Render()

    // 初始状态
    if state != "initial" {
        t.Errorf("Expected 'initial', got '%s'", state)
    }

    // 触发变化
    test.ClickButton(0)

    // 验证新状态
    if state != "changed" {
        t.Errorf("Expected 'changed', got '%s'", state)
    }
}
```

### 4. 测试完整用户流程

```go
func TestLoginFormFlow(t *testing.T) {
    username := ""
    password := ""
    loggedIn := false

    component := func() VNode {
        if loggedIn {
            return VStack(
                Text("Welcome!"),
                ButtonBuilder("Logout").
                    OnClick(func() { loggedIn = false }).
                    Build(),
            )
        }

        return VStack(
            InputBuilder().
                Placeholder("Username").
                OnChange(func(s string) { username = s }).
                Build(),
            InputBuilder().
                Placeholder("Password").
                OnChange(func(s string) { password = s }).
                Build(),
            ButtonBuilder("Login").
                OnClick(func() {
                    if username == "admin" && password == "pass" {
                        loggedIn = true
                    }
                }).
                Build(),
        )
    }

    test := NewComponentTest(t, component).Render()

    // 初始状态：登录表单
    test.AssertButtonCount(1).AssertButtonLabel(0, "Login")

    // 输入用户名
    test.TypeText("admin")
    test.TypeText("pass")

    // 点击登录
    test.ClickButton(0)

    // 验证登录成功
    test.AssertButtonLabel(0, "Logout")
    if !loggedIn {
        t.Error("Expected to be logged in")
    }
}
```

### 5. 测试边界条件

```go
func TestInputMaxLength(t *testing.T) {
    component := func() VNode {
        text, setText := UseStateString("")
        return InputBuilder().
            Value(text).
            MaxLength(5).
            OnChange(setText).
            Build()
    }

    test := NewComponentTest(t, component).Render()

    // 输入超过最大长度的文本
    test.TypeText("123456")

    // 验证被截断
    // 注意：需要组件实现 maxLength 逻辑
}
```

---

## 示例

### 示例 1: 简单按钮测试

```go
func TestSimpleButton(t *testing.T) {
    component := func() VNode {
        return ButtonBuilder("Click Me").Build()
    }

    NewComponentTest(t, component).
        Render().
        AssertButtonCount(1).
        AssertButtonLabel(0, "Click Me")
}
```

### 示例 2: 按钮点击测试

```go
func TestButtonClick(t *testing.T) {
    clickCount := 0

    component := func() VNode {
        return ButtonBuilder("Increment").
            OnClick(func() { clickCount++ }).
            Build()
    }

    test := NewComponentTest(t, component).Render()

    // 点击三次
    for i := 0; i < 3; i++ {
        test.ClickButton(0)
    }

    if clickCount != 3 {
        t.Errorf("Expected clickCount=3, got %d", clickCount)
    }
}
```

### 示例 3: 复选框测试

```go
func TestCheckboxToggle(t *testing.T) {
    component := func() VNode {
        checked, setChecked := UseStateBool(false)
        return CheckboxBuilder().
            Label("Accept Terms").
            Checked(checked).
            OnChange(setChecked).
            Build()
    }

    test := NewComponentTest(t, component).Render()

    // 初始状态：未选中
    test.AssertCheckboxChecked(0, false)

    // 切换一次
    test.ToggleCheckbox(0)
    test.AssertCheckboxChecked(0, true)

    // 再切换一次
    test.ToggleCheckbox(0)
    test.AssertCheckboxChecked(0, false)
}
```

### 示例 4: 模态框测试

```go
func TestModalOpenClose(t *testing.T) {
    var isModalOpen bool

    component := func() VNode {
        isOpen, setIsOpen, _ := UseStateInt(0)
        isModalOpen = isOpen == 1

        if isOpen == 1 {
            return VStack(
                Text("Modal Content"),
                ButtonBuilder("Close").
                    OnClick(func() { setIsOpen(0) }).
                    Build(),
            )
        }

        return VStack(
            Text("Main Content"),
            ButtonBuilder("Open").
                OnClick(func() { setIsOpen(1) }).
                Build(),
        )
    }

    test := NewComponentTest(t, component).Render()

    // 初始：弹窗关闭
    if isModalOpen {
        t.Error("Modal should be closed")
    }
    test.AssertButtonLabel(0, "Open")

    // 打开弹窗
    test.ClickButton(0)

    if !isModalOpen {
        t.Error("Modal should be open")
    }
    test.AssertButtonLabel(0, "Close")

    // 关闭弹窗
    test.ClickButton(0)

    if isModalOpen {
        t.Error("Modal should be closed")
    }
    test.AssertButtonLabel(0, "Open")
}
```

### 示例 5: 多步骤流程测试

```go
func TestWizardFlow(t *testing.T) {
    step := 0

    component := func() VNode {
        state, setState, _ := UseStateInt(0)
        step = state

        switch state {
        case 0:
            return VStack(
                Text("Step 1: Welcome"),
                ButtonBuilder("Next").
                    OnClick(func() { setState(1) }).
                    Build(),
            )
        case 1:
            return VStack(
                Text("Step 2: Confirm"),
                HStack(
                    ButtonBuilder("Back").
                        OnClick(func() { setState(0) }).
                        Build(),
                    ButtonBuilder("Next").
                        OnClick(func() { setState(2) }).
                        Build(),
                ),
            )
        case 2:
            return VStack(
                Text("Step 3: Done"),
                ButtonBuilder("Finish").
                    OnClick(func() { setState(0) }).
                    Build(),
            )
        default:
            return Text("")
        }
    }

    test := NewComponentTest(t, component).Render()

    // Step 0
    test.AssertButtonCount(1)
    test.ClickButton(0)

    // Step 1
    test.AssertButtonCount(2)
    test.ClickButton(1) // Click "Next"

    // Step 2
    test.AssertButtonCount(1)
    test.ClickButton(0) // Click "Finish"

    // Back to Step 0
    test.AssertButtonCount(1)
}
```

### 示例 6: 按标签查找按钮

```go
func TestClickByLabel(t *testing.T) {
    var clicked string

    component := func() VNode {
        return HStack(
            ButtonBuilder("Save").
                OnClick(func() { clicked = "save" }).
                Build(),
            ButtonBuilder("Cancel").
                OnClick(func() { clicked = "cancel" }).
                Build(),
            ButtonBuilder("Delete").
                OnClick(func() { clicked = "delete" }).
                Build(),
        )
    }

    test := NewComponentTest(t, component).Render()

    // 按标签点击
    test.ClickButtonByLabel("Cancel")
    if clicked != "cancel" {
        t.Errorf("Expected 'cancel', got '%s'", clicked)
    }

    test.ClickButtonByLabel("Save")
    if clicked != "save" {
        t.Errorf("Expected 'save', got '%s'", clicked)
    }
}
```

---

## 测试覆盖报告

运行测试并查看覆盖率：

```bash
# 运行所有测试
go test ./ui/... -v

# 查看覆盖率
go test ./ui/... -cover

# 生成覆盖率报告
go test ./ui/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 常见问题

### Q: 如何测试需要真实终端的组件？

A: 测试框架设计为无头测试，不需要真实终端。如果组件依赖终端特性（如尺寸），可以通过 `WithWidth` 和 `WithHeight` 选项配置。

### Q: 状态捕获的闭包问题

A: 由于 Go 闭包的特性，在组件外部捕获的变量可能显示旧值。使用 `getState()` 方法或直接在测试中验证状态变化：

```go
// ❌ 不推荐
isOpen, setIsOpen := UseStateBool(false)
// ... onClick 中调用 setIsOpen(true)
// isOpen 仍显示 false（闭包捕获旧值）

// ✅ 推荐
_, setIsOpen, getState := UseStateInt(0)
// ...
if getState() == 1 { /* 弹窗打开 */ }
```

### Q: 如何测试异步操作？

A: 当前测试框架是同步的。对于异步操作，使用 `time.Sleep` 或在测试中等待状态变化。

---

## 附录

### 完整测试示例文件

参见 `ui/component_test.go` 获取更多测试示例。

### 相关文档

- [组件 API 文档](./ui/README.md)
- [Hooks API 文档](./ui/README.md#hooks-api)
- [主 README](../README.md)
