# Focus Switching Demo

This demo demonstrates focus management in the Mint UI framework using the `ui.Run()` entry point.

## Features

### Focusable Components

| Component | Count | Focus ID Format | Intent Method |
|-----------|-------|-----------------|---------------|
| Button | 3 | `button:{key}` | `.SetIntent(intent.Click(targetID))` |
| Input | 2 | `input:{key}` | `.ForField(intent.BindField(key))` |
| Checkbox | 2 | `checkbox:{key}` | `.SetIntent(intent.Toggle(key))` |

### Navigation

- **TAB** - Move to next focusable element
- **SHIFT+TAB** - Move to previous focusable element
- **ENTER** - Activate focused button/checkbox
- **SPACE** - Toggle focused checkbox
- **ESC / CTRL+C** - Exit the app

## MVP Pattern - State Management

这个 demo 展示了 Mint 的 MVP 模式，使用 `UseState` + `FieldChangeIntent`：

```go
func FocusApp() ui.VNode {
    // 1. 使用 UseState 创建表单状态
    input1Value, setInput1Value := ui.UseStateString("")
    input2Value, setInput2Value := ui.UseStateString("")

    // 2. 将 setter 保存到 GlobalState，供 FieldChangeIntent handler 调用
    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalState["input1-value-setter"] = setInput1Value
        ctx.GlobalState["input2-value-setter"] = setInput2Value
    }

    return ui.VStack(
        // 3. Input 使用 ForField + Value 绑定
        inputComp.NewBuilder().
            ForField(intent.BindField("input1-value")).
            Value(input1Value).
            Placeholder("Enter name...").
            Build().
            SetKey("input1"),

        // 显示当前值（实时反馈）
        ui.NewTextBuilder(fmt.Sprintf("Value: %s", input1Value)).
            FgColor("bright-black").
            Build(),
        // ...
    )
}

func main() {
    err := ui.Run(FocusApp,
        ui.WithWidth(60),
        ui.WithHeight(35),
        ui.WithTitle("Focus Switching Demo"),
        ui.WithInit(func() {
            // 4. 注册 FieldChangeIntent handler
            // 从 GlobalState 获取 setter 并调用
            ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
                switch i.Field {
                case "input1-value":
                    if fn, ok := ctx.GetState("input1-value-setter"); ok {
                        if setter, ok := fn.(func(string)); ok {
                            setter(i.Value)  // 更新 UseState
                        }
                    }
                // ...
                }
                return intent.HandledResult()
            })
        }),
    )
}
```

### 数据流

```
用户输入
   ↓
Instance 内部状态（临时缓冲）
   ↓
FieldChangeIntent{Field, Value}
   ↓
Handler 从 GlobalState 获取 setter
   ↓
setter(i.Value) 更新 UseState
   ↓
自动重渲染
   ↓
VNode 更新
   ↓
Instance 更新显示
```

### 关键点

1. **UseState** 创建组件局部状态（存储在 hooks）
2. **ForField** 配置组件发射 `FieldChangeIntent`
3. **setter → GlobalState**：将闭包 setter 存储到 GlobalState
4. **Handler → setter**：从 GlobalState 获取 setter 并调用更新 UseState
5. **Value 绑定**：Input 的 `.Value()` 属性绑定 UseState 状态

## Intent Usage Summary

### Button - OnPress()

```go
// 使用自定义 Intent (推荐 - 完全控制)
buttonComp.NewBuilder("Button 1 - First").
    OnPress(ClickButtonIntent{}).
    Build().
    SetKey("btn1")
```

**注意事项**:
- `Click` intent **没有内置 handler**，使用时会产生警告
- 推荐使用**自定义 Intent** + 注册 handler，完全控制行为
- 或者使用内置 Intent：`Toggle`, `OpenModal`, `SetState`, `Navigate` 等

自定义 Intent 示例：
```go
type ClickButtonIntent struct{}

func (ClickButtonIntent) IntentType() string { return "ClickButton" }
func (ClickButtonIntent) StayPressed() bool  { return true }

// 状态管理
clickCount, setClickCount, _ := ui.UseStateInt(0)
ctx.GlobalState["setClickCount"] = setClickCount

// 注册 handler
ui.RegisterIntent(func(ctx *intent.ActionContext, i ClickButtonIntent) intent.IntentResult {
    if fn, ok := ctx.GetState("setClickCount"); ok {
        if setter, ok := fn.(func(interface{})); ok {
            // 使用 functional update：基于当前值递增
            setter(func(c int) int {
                return c + 1
            })
        }
    }
    return intent.HandledResult()
})
```

**Functional Update 说明**：
- `UseStateInt` 返回三个值：`(currentValue, setValue, getValue)`
- `setValue` 支持直接设置值：`setValue(5)`
- `setValue` 也支持 functional update：`setValue(func(c int) int { return c + 1 })`
- Functional update 会先通过 `getValue()` 获取当前值，然后应用函数更新

### Checkbox - OnToggle()

```go
checkboxComp.NewBuilder().
    Label("Option A").
    OnToggle(intent.Toggle("chk1-checked")).
    Build().
    SetKey("chk1")
```

### Input - ForField() + Value()

```go
inputComp.NewBuilder().
    ForField(intent.BindField("input1-value")).
    Value(input1Value).
    Placeholder("Enter name...").
    Build().
    SetKey("input1")
```

### Available Intents

| Intent | Constructor | Priority | Purpose |
|--------|-------------|----------|---------|
| Toggle | `intent.Toggle(key)` | UserBlocking | Toggle boolean state (used for button clicks) |
| SetState | `intent.SetState(key, value)` | Normal | Set a state value |
| Focus | `intent.Focus(targetID)` | Immediate | Focus an element |
| Blur | `intent.Blur(targetID)` | Immediate | Remove focus |
| Navigate | `intent.Navigate(path)` | UserBlocking | Navigate to a page |
| OpenModal | `intent.OpenModal(modalID)` | UserBlocking | Open a modal |
| CloseModal | `intent.CloseModal(modalID)` | UserBlocking | Close a modal |

**注意**: `Click` intent 没有内置 handler。在这个 demo 中，按钮使用 `Toggle` intent 来更新点击计数。

## Running the Demo

```bash
cd examples/fiber_firsts/focus_switching_demo
go run main.go
```

Or build and run:

```bash
go build -o focus_demo.exe main.go
./focus_demo.exe
```
