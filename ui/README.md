# Mint UI - 声明式UI框架

Mint UI 是一个现代化的 Go 终端用户界面 (TUI) 框架，采用 React-like 的声明式编程模式。

## 特性

- **声明式UI** - 类似 React 的组件开发体验
- **Hooks系统** - useState, useEffect, useMemo, useCallback, useRef
- **组件库** - 丰富的预置组件
- **Builder模式** - 流畅的链式API
- **类型安全** - 完整的类型支持

## 快速开始

```go
package main

import "github.com/wwsheng009/mint/ui"

func main() {
    ui.Run(func() ui.VNode {
        return ui.VStack(
            ui.NewTextBuilder("Hello, Mint UI!").Build(),
            ui.ButtonBuilder("Click Me").Build(),
        )
    },
        ui.WithWidth(40),
        ui.WithHeight(10),
        ui.WithTitle("My App"),
    )
}
```

## 组件API

### Button 按钮

```go
ui.ButtonBuilder("Click Me").
    OnClick(func() { println("clicked") }).
    Disabled(false).
    Build()
```

### Input 输入框

```go
text, setText := ui.UseStateString("")

ui.InputBuilder().
    Value(text).
    Placeholder("Type here...").
    MaxLength(20).
    OnChange(setText).
    Build()
```

### Checkbox 复选框

```go
checked, setChecked := ui.UseStateBool(false)

ui.CheckboxBuilder().
    Label("Accept terms").
    Checked(checked).
    OnChange(setChecked).
    Build()
```

### Progress 进度条

```go
progress, setProgress, _ := ui.UseStateInt(0)

ui.ProgressBuilder().
    Label("Loading:").
    Value(progress).
    Max(100).
    ShowPercent(true).
    Build()
```

### Select 选择器

```go
selected, setSelected := ui.UseStateString("option1")

ui.SelectBuilder().
    AddOption("option1", "Option 1").
    AddOption("option2", "Option 2").
    Selected(0).
    OnChange(setSelected).
    Build()
```

### Table 表格

```go
ui.TableBuilder().
    Columns([]ui.TableColumn{
        {Title: "ID", Width: 5},
        {Title: "Name", Width: 15},
        {Title: "Status", Width: 10},
    }).
    AddRow("1", "Alice", "Active").
    AddRow("2", "Bob", "Inactive").
    Build()
```

### Spinner 加载动画

```go
ui.SpinnerBuilder().
    Message("Loading...").
    FgColor("cyan").
    Build()
```

## Hooks API

### UseStateInt / UseStateString / UseStateBool

```go
count, setCount, _ := ui.UseStateInt(0)
text, setText := ui.UseStateString("")
active, setActive := ui.UseStateBool(false)

// 更新状态
setCount(42)
setText("hello")
setActive(true)

// 函数式更新
setCount(func(c int) int { return c + 1 })
```

### UseMemo

```go
expensiveValue := ui.UseMemo(func() interface{} {
    return count * count
}, []interface{}{count})
```

### UseCallback

```go
callback := ui.UseCallback(func() {
    println("callback")
}, []interface{}{dependency})
```

### UseRef

```go
ref := ui.UseRef(0)
ref.Value = ref.Value.(int) + 1
```

### UseEffect

```go
ui.UseEffect(func() ui.CleanupFunc {
    // 初始化逻辑
    ticker := time.NewTicker(time.Second)

    return func() {
        // 清理逻辑
        ticker.Stop()
    }
}, []interface{}{dependency})
```

## 布局组件

### VStack 垂直布局

```go
ui.VStack(
    ui.Text("Line 1"),
    ui.Text("Line 2"),
    ui.Text("Line 3"),
)
```

### HStack 水平布局

```go
ui.HStack(
    ui.Text("A"),
    ui.Text("B"),
    ui.Text("C"),
)
```

### Fragment 片段

```go
ui.Fragment(
    ui.Text("Item 1"),
    ui.Text("Item 2"),
)
```

## 样式API

```go
ui.NewTextBuilder("Styled Text").
    FgColor("cyan").
    BgColor("black").
    Bold(true).
    Build()
```

## 键盘事件

| 按键 | 功能 |
|------|------|
| `Tab` | 下一个焦点 |
| `Shift+Tab` | 上一个焦点 |
| `Enter` | 激活/选择 |
| `Space` | 切换复选框 |
| `↑` `↓` | Select选项导航 |
| `←` `→` | 元素导航 |
| `Backspace` | 删除字符 |
| `q` | 退出 |

## 运行选项

```go
ui.Run(component,
    ui.WithWidth(80),      // 窗口宽度
    ui.WithHeight(24),     // 窗口高度
    ui.WithTitle("App"),   // 窗口标题
    ui.WithFPS(60),         // 帧率限制
)
```

## 目录

```
ui/
├── vnode.go          # 虚拟DOM节点
├── component.go      # 组件系统
├── element.go        # 元素节点
├── fragment.go       # 片段
├── layout.go         # 布局容器
├── button.go         # 按钮
├── text.go           # 文本
├── input.go          # 输入
├── checkbox.go       # 复选框
├── progress.go       # 进度条/加载动画
├── select.go         # 选择器/表格
├── hooks.go          # Hooks
├── app.go            # 应用运行时
├── fiber.go          # Fiber架构
├── scheduler.go      # 调度器
└── *_test.go         # 测试
```

## 测试

```bash
go test ./ui/... -v -cover
```

当前测试覆盖率: **40.7%**

## Toolbar / StatusBar

```go
bar := ui.NewToolbarBuilder().
    Key("ops.toolbar").
    Title("Load Balancer").
    TitleWidth(16).
    Left(ui.ToolbarText("scope", "group: default").WithWidth(20)).
    Center(ui.ToolbarBadge("state", "degraded").WithColors("black", "yellow")).
    Right(ui.ToolbarButton("refresh", "Refresh", RefreshIntent{}).Primary()).
    Right(ui.ToolbarButton("reset", "Reset Runtime", ResetIntent{}).Danger().WithDisabled(true).WithHelp("Requires selected target and reason")).
    Build()

confirm := ui.ConfirmDangerOperation(
    "ops.confirm",
    "Disable Provider Key",
    "Disable the selected provider key and record an audit reason.",
    "Traffic may fail over to another available key.",
    "Disable",
    "actionReason",
    "maintenance",
    DisableKeyIntent{},
    CloseConfirmIntent{},
    ui.ConfirmDialogTargetItem("provider", "Provider", "openai"),
    ui.ConfirmDialogSensitiveTarget("key", "Key", "provider-key-demo"),
)
```

```go
bar := ui.NewStatusBarBuilder().
    Theme(ui.StatusBarThemeDefault().WithTooltipArrowStyle(ui.StatusBarTooltipArrowRounded)).
    HelpDisplayMode(ui.StatusBarHelpOverlay).
    TooltipPlacement(ui.StatusBarTooltipAuto).
    TooltipGapRows(1).
    TooltipMaxWidth(38).
    HelpPrefix("? ").
    HelpFallback("Hover actions to inspect overlay help").
    Left(ui.StatusBarActionBadge(" MODE ", "black", "yellow", CycleModeIntent{}).WithHelp("Cycle to the next mode")).
    Center(ui.StatusBarText("Mint Demo").WithWidth(20).WithAlign(ui.AlignCenter).WithEllipsis()).
    Right(ui.StatusBarActionBadge(" SAVE ", "black", "green", SaveIntent{}).WithHelp("Save current document")).
    BuildWithHelp()

bar2 := ui.StatusBarWithHelpMode(
    ui.StatusBarThemeMuted(),
    "F6 cycles mode | Ctrl+1/2/3 jumps directly",
    ui.StatusBarHelpOverlay,
    ui.StatusBarSections(
        ui.StatusBarActionBadge(" UI ", "black", "yellow", SetModeIntent{Mode: "interactive"}).WithHelp("Regular UI interaction mode"),
    ),
    ui.StatusBarSections(
        ui.StatusBarText("StatusBar supports hover-only overlay tooltips").WithWidth(24).WithAlign(ui.AlignCenter).WithEllipsis(),
    ),
    ui.StatusBarSections(
        ui.StatusBarActionBadge(" TERM ", "bright-white", "bright-black", SetModeIntent{Mode: "terminal_selection"}).WithHelp("Native terminal selection mode"),
    ),
)
```
