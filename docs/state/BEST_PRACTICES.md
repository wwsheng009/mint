# Fiber-first 状态管理最佳实践

本文档提供 Mint Fiber-first 模式下状态管理的最佳实践指南。

## 目录

- [状态类型选择](#状态类型选择)
- [架构模式](#架构模式)
- [性能优化](#性能优化)
- [常见陷阱](#常见陷阱)
- [代码示例](#代码示例)

---

## 状态类型选择

### 决策树

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         状态类型选择                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  状态需要跨越多个组件吗？                                               │
│       │                                                                │
│       ├─ 是 ──→ 使用全局状态 (Intent + RootContext.State) │
│       │          示例：步骤索引、用户信息、表单数据                       │
│       │                                                                │
│       └─ 否 ──→ 组件是叶子节点吗？                                     │
│                     │                                                  │
│                     ├─ 是 ──→ 使用局部状态 (useState) │
│                     │          示例：toggle、input 值、展开/折叠         │
│                     │                                                  │
│                     └─ 否 ──→ 是父组件给子组件传递数据吗？            │
│                                   │                                    │
│                                   ├─ 是 ──→ 使用 Props ↓
│                                   │          示例：表格行数据           │
│                                   │                                    │
                                   └─ 否 ──→ 重新评估设计                 │
│                                             可能在组件边界划分不当      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────┘
```

### 使用场景对比

| 状态类型 | 场景 | 生命周期 | 示例 |
|---------|------|---------|------|
| **useState** | 组件内部状态 | 组件销毁 | 按钮 toggle、Input 值、Accordion 展开/折叠 |
| **全局状态** | 跨组件共享 | 应用运行时 | 当前步骤、用户信息、主题、表单验证状态 |
| **Props** | 父→子单向传递 | Fiber 更新 | 列表项、模态框标题、图表数据 |
| **Context** | 深层组件访问 | 应用运行时 | 主题、Locale、认证状态 |

### 最佳实践示例

```go
// ✅ 推荐：组件内状态用 useState
func CollapsibleSection() ui.VNode {
    isOpen, setIsOpen := rtui.UseState(false)

    return ui.VStack(
        ui.Button(if isOpen ? "▼ 收起" : "► 展开").
            OnClick(func() { setIsOpen(!isOpen) }).
            Build(),
        ui.Text("...内容...").
            Visible(isOpen).
            Build(),
    )
}

// ✅ 推荐：跨组件状态用全局状态
func App() ui.VNode {
    ctx := rtui.GetCurrentContext()
    currentStep := ctx.GetIntState("step", 1)
    username := ctx.GetStringState("username", "")

    return ui.VStack(
        StepIndicator(currentStep),       // 读取全局状态
        UserProfile(username),             // 读取全局状态
        NextButton(currentStep),           // 触发全局状态更新
    )
}

// ❌ 不推荐：跨组件状态用 useState
// 问题：状态无法在其他组件中访问
func WrongExample() ui.VNode {
    step, setStep := rtui.UseState(1)

    // StepIndicator 无法访问 step
    return ui.VStack(
        ui.Text(fmt.Sprintf("Step: %d", step)),
        StepIndicator(),  // ❌ 无法读取 step
    )
}
```

---

## 架构模式

### 模式 1：中心化状态管理（推荐用于复杂应用）

适用于：多步骤表单、导航状态、主题等全局状态。

```go
// main.go - 定义全局状态 Intent
type AppIntent struct {
    Type  string
    Value interface{}
}

func (AppIntent) IntentType() string { return "AppIntent" }

// 注册 Intent Handler
ui.RegisterIntent(func(ctx *intent.ActionContext, i AppIntent) intent.IntentResult {
    switch i.Type {
    case "SET_STEP":
        ctx.SetState("step", i.Value.(int))
    case "SET_USERNAME":
        ctx.SetState("username", i.Value.(string))
    case "SET_THEME":
        ctx.SetState("theme", i.Value.(string))
    }
    return intent.HandledResult()
})

// 组件通过 Intent 更新全局状态
func NextButton(currentStep int) ui.VNode {
    return app.Button("Next").
        OnPress(AppIntent{Type: "SET_STEP", Value: currentStep + 1}).
        Build()
}
```

### 模式 2：表单状态聚合

适用于：多步骤表单、复杂表单。

```go
// 表单状态 Intent
type FormUpdateIntent struct {
    Field string  // 字段名
    Value interface{}
}
func (FormUpdateIntent) IntentType() string { return "FormUpdate" }

// 注册 Handler
ui.RegisterIntent(func(ctx *intent.ActionContext, i FormUpdateIntent) intent.IntentResult {
    ctx.SetState(i.Field, i.Value)
    return intent.HandledResult()
})

// 表单组件
func Step1() ui.VNode {
    ctx := rtui.GetCurrentContext()
    username := ctx.GetStringState("username", "")
    email := ctx.GetStringState("email", "")

    return ui.VStack(
        InputField("Username:", username, "username", FormUpdateIntent{Field: "username"}),
        InputField("Email:", email, "email", FormUpdateIntent{Field: "email"}),
    )
}

func InputField(label, value, field string, onChangeIntent FormUpdateIntent) ui.VNode {
    return ui.VStack(
        ui.Text(label),
        app.Input().Value(value).OnChange(changeIntent).
            OnInput(func(newValue string) {
                // 发射更新 Intent
                onChangeIntent.Value = newValue
                EmitIntent(onChangeIntent)
            }).
            Build(),
    )
}
```

### 模式 3：局部组件封装

适用于：独立 UI 组件（如 Button、Checkbox、Accordion）。

```go
// ✅ Button - 纯组件，状态完全对外
type ButtonProps struct {
    Label    string
    OnPress  intent.Intent
    Variant  string
}

func Button(props ButtonProps) ui.VNode {
    // 无内部状态，完全通过 Props 和 Intent 控制
    return app.Button(props.Label).
        Variant(props.Variant).
        OnPress(props.OnPress).
        Build()
}

// 使用
func App() ui.VNode {
    ctx := rtui.GetCurrentContext()
    step := ctx.GetIntState("step", 1)

    return ui.HStack(
        Button{
            Label:   "Previous",
            OnPress: UpdateStepIntent{Step: step - 1},
            Variant: ButtonVariantSecondary,
        },
        Button{
            Label:   "Next",
            OnPress: UpdateStepIntent{Step: step + 1},
            Variant: ButtonVariantPrimary,
        },
    )
}
```

### 模式 4：混合状态（需谨慎）

适用于：同时需要全局和局部状态的复杂组件。

```go
// ✅ 推荐：清晰分离全局和局部状态
func AccordionItem(title string) ui.VNode {
    // 局部状态：展开/折叠
    isExpanded, setIsExpanded := rtui.UseState(false)

    // 全局状态：用户选择
    ctx := rtui.GetCurrentContext()
    isSelected := ctx.GetBoolState(fmt.Sprintf("selected:%s", title), false)

    return ui.Bordered().Child(
        ui.VStack(
            ui.HStack(
                ui.Text(title).
                    Bold(isSelected).
                    FgColor(fgColor(isSelected)).
                    Build(),
                ui.Text(if isExpanded ? "▼" : "►").
                    OnClick(func() { setIsExpanded(!isExpanded) }).
                    Build(),
            ),
            // 全局状态更新：选中/取消选中
            ui.Text("Select").
                OnClick(func() {
                    ctx.SetState(fmt.Sprintf("selected:%s", title), !isSelected)
                }).
                Build(),
            // 局部状态展开内容
            ui.Text("...内容...").
                Visible(isExpanded).
                Build(),
        ),
    )
}

// ❌ 不推荐：混合使用时不加注释
func AmbiguousExample() ui.VNode {
    // 读者难以理解哪个是全局状态，哪个是局部状态
    expanded, _ := rtui.UseState(false)
    selected := ctx.GetBoolState("selected", false)
    // ...
}
```

---

## 性能优化

### 优化 1：避免不必要的重渲染

```go
// ❌ 不推荐：每次点击都触发全局状态更新
func HeavyComponent() ui.VNode {
    ctx := rtui.GetCurrentContext()
    counter := ctx.GetIntState("counter", 0)

    // 这个按钮是局部的，不应该更新全局状态
    return app.Button(fmt.Sprintf("Clicks: %d", counter)).
        OnClick(func() {
            ctx.SetState("counter", counter + 1)  // ❌ 全局状态更新
        }).
        Build()
}

// ✅ 推荐：局部状态用 useState
func FixedHeavyComponent() ui.VNode {
    clickCount, setClickCount := rtui.UseState(0)

    return app.Button(fmt.Sprintf("Clicks: %d", clickCount)).
        OnClick(func() {
            setClickCount(clickCount + 1)  // ✅ 局部状态，不影响其他组件
        }).
        Build()
}
```

### 优化 2：批量更新

```go
// ❌ 不推荐：多次独立更新
func MultipleUpdatesBad() ui.VNode {
    ctx := rtui.GetCurrentContext()

    return app.Button("Update All").
        OnClick(func() {
            ctx.SetState("field1", "value1")  // 触发重渲染
            ctx.SetState("field2", "value2")  // 再次触发重渲染
            ctx.SetState("field3", "value3")  // 第三次触发重渲染
        }).
        Build()
}

// ✅ 推荐：使用批量 Intent
type BatchUpdateIntent struct {
    Updates map[string]interface{}
}
func (BatchUpdateIntent) IntentType() string { return "BatchUpdate" }

// 注册 Handler
ui.RegisterIntent(func(ctx *intent.ActionContext, i BatchUpdateIntent) intent.IntentResult {
    for key, value := range i.Updates {
        ctx.SetState(key, value)  // 一次性更新多个字段
    }
    return intent.HandledResult()
})

// 使用
func MultipleUpdatesGood() ui.VNode {
    return app.Button("Update All").
        OnPress(BatchUpdateIntent{
            Updates: map[string]interface{}{
                "field1": "value1",
                "field2": "value2",
                "field3": "value3",
            },
        }).
        Build()
}
```

### 优化 3：Memo 和 Key

```go
// ✅ 推荐：为列表项设置 Key
func UserList(users []User) ui.VNode {
    items := make([]ui.VNode, len(users))
    for i, user := range users {
        items[i] = UserItem(user).
            Key(fmt.Sprintf("user:%d", user.ID)).  // ✅ 唯一 Key
            Build()
    }
    return ui.VStack(items...)
}

// ✅ 推荐：使用 Memo 避免不变 Props 的重渲染
func UserItem(user User) ui.VNode {
    return ui.Memo(
        ui.VStack(
            ui.Text(user.Name),
            ui.Text(user.Email),
        ),
        func(old, new interface{}) bool {
            oldProps := old.(Props)
            newProps := new.(Props)
            return oldProps["id"] == newProps["id"]  // 相等则跳过重渲染
        },
    )
}
```

### 优化 4：延迟加载（Lazy）

```go
// ✅ 推荐：只在需要时加载组件
func LazyTabPanel() ui.VNode {
    ctx := rtui.GetCurrentContext()
    activeTab := ctx.GetIntState("activeTab", 1)

    contents := map[int]ui.ComponentFunc{
        1: Tab1Content,
        2: Tab2Content,
        3: Tab3Content,
    }

    return contents[activeTab]()  // 只渲染当前激活的
}

// ❌ 不推荐：渲染所有标签页的内容
func AllTabsAlwaysRender() ui.VNode {
    return ui.VStack(
        Tab1Content(),  // 即使不可见也渲染
        Tab2Content(),
        Tab3Content(),
    )
}
```

---

## 常见陷阱

### 陷阱 1：混合使用 Props 和 State

```go
// ❌ 不推荐：Props 和 State 的优先级不清晰
func ConfusingComponent(props Props) ui.VNode {
    ctx := rtui.GetCurrentContext()
    // value 来自 Props 还是 State？
    value := props["value"]  // 可能与 ctx.GetState("value") 冲突
    // ...
}

// ✅ 推荐：明确来源
func ClearComponent(props Props) ui.VNode {
    ctx := rtui.GetCurrentContext()

    // Props：外部控制（只读）
    username := props["username"].(string)

    // State：内部控制（可变）
    selfValue := ctx.GetIntState("selfValue", 0)

    // ...
}
```

### 陷阱 2：直接修改全局 State

```go
// ❌ 不推荐：绕过 Intent 直接修改
func BypassIntent() ui.VNode {
    ctx := rtui.GetCurrentContext()

    return app.Button("Increment").
        OnClick(func() {
            ctx.State["counter"] = ctx.GetIntState("counter", 0) + 1  // ❌ 直接修改
            fwApp.MarkDirty()  // 手动标记更新
        }).
        Build()
}

// ✅ 推荐：通过 Intent 更新
func UseIntent() ui.VNode {
    ctx := rtui.GetCurrentContext()

    return app.Button("Increment").
        OnPress(IncrementCounterIntent{}).
        Build()
}
```

### 陷阱 3：状态同步问题

```go
// ❌ 不推荐：状态分散，同步困难
func UnsyncedState() ui.VNode {
    ctx := rtui.GetCurrentContext()
    step := ctx.GetIntState("step", 1)

    // 问题：两个不同步的状态
    maxStep := ctx.GetIntState("maxStep", 3)
    // 如果 maxStep 改变，但 step 没相应调整，会出现 bug

    return ui.Text(fmt.Sprintf("Step %d of %d", step, maxStep))
}

// ✅ 推荐：从单个状态派生
func DerivedState() ui.VNode {
    ctx := rtui.GetCurrentContext()
    step := ctx.GetIntState("step", 1)
    maxStep := 3  // ✅ 硬编码或从配置读取

    return ui.Text(fmt.Sprintf("Step %d of %d", step, maxStep))
}
```

### 陷阱 4：过度使用全局状态

```go
// ❌ 不推荐：所有状态都是全局的
func OverGlobalized() ui.VNode {
    ctx := rtui.GetCurrentContext()
    isExpanded := ctx.GetBoolState("item1:expanded", false)  // 应该是局部的
    selectedText := ctx.GetStringState("item1:selectedText", "")  // 应该是局部的

    return ui.VStack(
        app.TextInput().Value(selectedText).OnInput(func(v string) {
            ctx.SetState("item1:selectedText", v)  // ❌ 全局污染
        }).Build(),
    )
}

// ✅ 推荐：局部状态用 useState
func ProperScoped() ui.VNode {
    isExpanded, setIsExpanded := rtui.UseState(false)
    selectedText, setSelectedText := rtui.UseState("")

    return ui.VStack(
        app.TextInput().Value(selectedText).OnInput(setSelectedText).Build(),
    )
}
```

---

## 代码示例

### 完整示例：多步骤表单

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/app"
    "github.com/wwsheng009/mint/runtime/intent"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Intent 定义
type UpdateStepIntent struct {
    Step int
}
func (UpdateStepIntent) IntentType() string { return "UpdateStep" }

type UpdateFieldIntent struct {
    Field string
    Value interface{}
}
func (UpdateFieldIntent) IntentType() string { return "UpdateField" }

// 主应用
func App() ui.VNode {
    ctx := rtui.GetCurrentContext()

    // 从全局状态读取
    step := ctx.GetIntState("step", 1)
    username := ctx.GetStringState("username", "")
    email := ctx.GetStringState("email", "")
    password := ctx.GetStringState("password", "")

    return ui.VStack(
        Header(),
        ProgressBar(step),
        StepContent(step, username, email, password),
        ActionButtons(step),
    )
}

// Header - 纯展示组件
func Header() ui.VNode {
    return ui.Text("User Registration Form").Bold(true)
}

// ProgressBar - 读取全局状态
func ProgressBar(step int) ui.VNode {
    const totalSteps = 3
    progress := float64(step) / float64(totalSteps)

    // 绘制进度条
    bar := make([]rune, 20)
    for i := 0; i < 20; i++ {
        if float64(i) < progress*20 {
            bar[i] = '█'
        } else {
            bar[i] = '░'
        }
    }

    return ui.Text(string(bar))
}

// StepContent - 根据步骤渲染不同表单
func StepContent(step, username, email, password string) ui.VNode {
    switch step {
    case 1:
        return AccountForm(username, email, password)
    case 2:
        return ProfileForm()
    case 3:
        return ConfirmForm(username, email)
    default:
        return ui.Text("Unknown step")
    }
}

// AccountForm - 表单字段通过 Intent 更新全局状态
func AccountForm(username, email, password string) ui.VNode {
    return ui.VStack(
        InputField("Username:", username, "username"),
        InputField("Email:", email, "email"),
        InputField("Password:", password, "password"),
    )
}

func InputField(label, value, field string) ui.VNode {
    return ui.VStack(
        ui.Text(label),
        app.NewTextBuilder().
            Value(value).
            OnInput(func(newValue string) {
                // 发射 Intent 更新全局状态
                EmitIntent(UpdateFieldIntent{Field: field, Value: newValue})
            }).
            Build(),
    )
}

// ActionButtons - 通过 Intent 更新步骤
func ActionButtons(currentStep int) ui.VNode {
    var buttons []ui.VNode

    if currentStep > 1 {
        buttons = append(buttons,
            app.Button("Previous").
                OnPress(UpdateStepIntent{Step: currentStep - 1}).
                Build(),
        )
    }

    if currentStep < 3 {
        buttons = append(buttons,
            app.Button("Next").
                OnPress(UpdateStepIntent{Step: currentStep + 1}).
                Build(),
        )
    } else {
        buttons = append(buttons,
            app.Button("Submit").Build(),
        )
    }

    return ui.HStack(buttons...)
}

func main() {
    ui.Run(App,
        ui.WithInit(func() {
            // 注册 Intent Handler
            ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateStepIntent) intent.IntentResult {
                ctx.SetState("step", i.Step)
                return intent.HandledResult()
            })

            ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateFieldIntent) intent.IntentResult {
                ctx.SetState(i.Field, i.Value)
                return intent.HandledResult()
            })
        }),
    )
}
```

---

## 总结

### 关键要点

1. **选择正确的状态类型**：
   - 组件内部 ⇒ `useState`
   - 跨组件 ⇒ 全局状态 + Intent
   - 父子通信 ⇒ Props

2. **避免常见陷阱**：
   - 不要混合 Props 和 State
   - 不要绕过 Intent 直接修改
   - 不要过度使用全局状态

3. **性能优化**：
   - 批量更新
   - Memo + Key
   - 延迟加载

### 快速参考

```go
// 局部状态
count, setCount := rtui.UseState(0)

// 全局状态 - 读取
ctx := rtui.GetCurrentContext()
step := ctx.GetIntState("step", 1)

// 全局状态 - 更新（通过 Intent）
ctx.SetState("step", 2)  // 在 Intent Handler 中
EmitIntent(UpdateStepIntent{Step: 2})  // 在组件中
```

### 相关文档

- [FIBER_STATE_ARCHITECTURE.md](./FIBER_STATE_ARCHITECTURE.md) - 核心架构
- [PERFORMANCE.md](/docsArchive/performance/FINAL_SUMMARY.md) - 性能优化
- [MIGRATION.md](/docsArchive/MIGRATION.md) - 从闭包模式迁移
