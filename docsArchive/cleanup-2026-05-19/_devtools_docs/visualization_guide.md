# DevTools 可视化使用指南

> **项目**: Mint TUI Runtime - DevTools
> **文档版本**: 1.0
> **创建日期**: 2026-01-30

---

## 目录

1. [快速开始](#一快速开始)
2. [TUI 调试面板](#二tui-调试面板)
3. [可视化视图](#三可视化视图)
4. [使用示例](#四使用示例)
5. [集成方法](#五集成方法)

---

## 一、快速开始

### 1.1 启用 DevTools

```go
import (
    "github.com/wwsheng009/mint/devtools"
    "github.com/wwsheng009/mint/devtools/client"
)

func main() {
    // 创建 DevTools 实例
    dt := devtools.New()

    // 创建 TUI 调试面板
    panel := client.NewTuiDebugPanel(dt)
    panel.Enable()

    // 在应用中使用
    app := createMyApp()

    // 渲染循环
    for {
        dt.BeginFrame()

        // 应用逻辑
        app.Update()
        app.Render()

        // 收集调试数据
        dt.CollectLayout(app.GetLayoutResult())

        dt.EndFrame()

        // 显示调试面板 (可选)
        if devtoolsKeyWasPressed() {
            showDebugPanel(panel)
        }
    }
}
```

### 1.2 快捷键

| 按键 | 功能 |
|------|------|
| `Ctrl+D` | 切换 DevTools 面板 |
| `Ctrl+T` | 切换时间线视图 |
| `Ctrl+C` | 切换因果图视图 |
| `Ctrl+S` | 切换快照视图 |
| `Ctrl+R` | 切换回放视图 |
| `F1` | 显示帮助 |

---

## 二、TUI 调试面板

### 2.1 面板布局

```
┌─────────────────────────────────────────────────────────────────┐
│                    DevTools Debug Panel                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Timeline View                                            │   │
│  │ ─────────────────────────────────────────────────────── │   │
│  │ Current Frame: 1234                                   │   │
│  │                                                         │   │
│  │ Frame Timeline:                                        │   │
│  │ ▓▓▓▓▓▓▓░░░▒▒▒▒░░░▓▓▓▓▓▓░░░▒▒▒▒░░░░░░░░░░░░░░░░░░░░░░        │   │
│  │ │││││││   │││││   ││││││││   ││││││                        │   │
│  │                                                         │   │
│  │ Metrics: Events: 45  Mutations: 12  Layouts: 8        │   │
│  │                                                         │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Causal Graph View                                        │   │
│  │ ─────────────────────────────────────────────────────── │   │
│  │                                                         │   │
│  │  [E] KeyPress (btn1)                                     │   │
│  │     │                                                    │   │
│  │     ├─> [M] SetText (btn1)                               │   │
│  │     │     │                                              │   │
│  │     │     └─> [L] Relayout (container)                    │   │
│  │     │           │                                        │   │
│  │     │           └─> [R] Repaint (btn1)                    │   │
│  │     │                                                  │   │
│  │     └─> [M] Focus (btn2)                                │   │
│  │                                                         │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ [t]imeline [c]ausal [s]napshot [r]eplay [h]elp [q]uit           │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 视图说明

#### Timeline View (时间线视图)

显示每帧的性能指标和活动：

```
┌─────────────────────────────────────────────────────────────┐
│ Timeline View                                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ Frame: 1234  Time: 12:34:56.789  Duration: 16.7ms         │
│                                                              │
│ ┌─ Layout ────┬─ Paint ───┬─ Input ───┬─ Idle ─────┐    │
│ │ ████░░░░░░░ │ ░░░░░░░░░░░ │ ████░░░░░ │ ░░░░░░░░░░░░ │    │
│ └──────────┴──────────┴─────────┴────────────┘    │
│   5ms         8ms        2ms        1.7ms            │
│                                                              │
│ Event Timeline:                                              │
│  ├─ KeyPress('a') @ 12:34:56.100                          │
│  ├─ Mutation(btn1.text) @ 12:34:56.120                    │
│  ├─ Layout(container) @ 12:34:56.150                        │
│  └─ Repaint(btn1) @ 12:34:56.160                            │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

#### Causal Graph View (因果图视图)

显示事件之间的因果关系：

```
┌─────────────────────────────────────────────────────────────┐
│ Causal Graph View                                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│    ┌────────────────────────────────────────────┐           │
│    │     Frame #1234                              │           │
│    └────────────────────────────────────────────┘           │
│                                                              │
│    [E1] click                                          │
│      │                                                 │
│      ├─────> [M1] button.onClick                     │
│      │              │                                 │
│      │              ├─────> [M2] form.submit          │
│      │              │              │                 │
│      │              │              └─────> [L1] Relayout  │
│      │              │                       │             │
│      │              │                       └─────> [R1] Paint │
│      │              │                                         │
│      │              └─────> [M3] state.isLoading          │
│      │                      │                             │
│      │                      └─────> [L2] Relayout          │
│      │                                │                 │
│      └─────> [M4] input.value                    │
│                                                 │
│    Legend:                                           │
│    [E] Event  [M] Mutation  [L] Layout  [R] Repaint │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## 三、可视化视图

### 3.1 组件树视图

```
┌─────────────────────────────────────────────────────────────┐
│ Component Tree                                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ 📦 App (root)                                              │
│   ├─ 📦 VBox (main)                                       │
│   │   ├─ 🔘 Button (submit_btn)  [focused]                │
│   │   │   └─ 📝 Text: "Submit"                             │
│   │   │                                                  │
│   │   ├─ 📝 Input (username)                             │
│   │   │   └─ 📝 Text: "user@example.com"                   │
│   │   │                                                  │
│   │   ├─ 📝 Input (password)                             │
│   │   │   └─ 🔒 Text: "********"                           │
│   │   │                                                  │
│   │   └─ 📦 HBox (actions)                                │
│   │       ├─ 🔘 Button (cancel)                            │
│   │       └─ 🔘 Button (help)                              │
│   │                                                      │
│   └─ 📦 StatusBar (status)                                │
│       └─ 📝 Text: "Ready"                                  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 3.2 性能火焰图

```
┌─────────────────────────────────────────────────────────────┐
│ Performance Flamegraph                                     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  100% ┤                                                    │
│   90% ┤ ████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░        │
│   80% ┤ ████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░         │
│   70% ┤ ████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░         │
│   60% ┤ ███████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░        │
│   50% ┤ ██████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░      │
│   40% ┤ ███████████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░   │
│   30% ┤ █████████████████████████████████████████░░░░░░░░░░░░░░░░░░░   │
│   20% ┤ █████████████████████████████████████████████████░░░░░░░░░░░░░░░ │
│   10% ┤ ████████████████████████████████████████████████████████████████ │
│    0% ┼────────────────────────────────────────────────────────────── │
│       └───────────────────────────────────────────────────────────   │
│         Layout  Paint  Input  Event  Idle  GC   Other                │
│                                                              │
│  Hot Functions:                                              │
│    • Container.Layout()          45% (450ms)                 │
│    • Button.Render()             25% (250ms)                 │
│    • Input.HandleKey()            15% (150ms)                 │
│    • EventBus.Dispatch()          10% (100ms)                 │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 3.3 布局盒视图

```
┌─────────────────────────────────────────────────────────────┐
│ Layout Boxes                                                │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  App (0, 0, 80x24)                                         │
│  ┌────────────────────────────────────────────────────────┐   │
│  │ VBox (0, 0, 80x20)                                     │   │
│  │ ┌──────────────────────────────────────────────────┐ │   │
│  │ │ Button (10, 2, 20x3) [Focused]                    │ │   │
│  │ │ ┌────────────────────────────────────────────────┐│ │   │
│  │ │ │ Text: "Submit"                                 ││ │   │
│  │ │ └────────────────────────────────────────────────┘│ │   │
│  │ └──────────────────────────────────────────────────┘ │   │
│  │ ┌──────────────────────────────────────────────────┐ │   │
│  │ │ Input (35, 2, 30x3)                               │ │   │
│  │ │ ┌────────────────────────────────────────────────┐│ │   │
│  │ │ │ Text: "user@example.com"  [Dirty]              ││ │   │
│  │ │ └────────────────────────────────────────────────┘│ │   │
│  │ └──────────────────────────────────────────────────┘ │   │
│  │ [ ... 3 more children ... ]                          │   │
│  └────────────────────────────────────────────────────┘   │
│  [ + 15 more components not shown ]                        │
│                                                              │
│  Legend:                                                     │
│    [Focused]  = 当前焦点                                     │
│    [Dirty]    = 需要重新布局                                   │
│    [Hidden]   = 不可见                                       │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 3.4 时间旅行控制

```
┌─────────────────────────────────────────────────────────────┐
│ Time Travel Control                                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Frame: 1234 / 5000                                       │
│  ┌────────────────────────────────────────────────────────┐   │
│  │ ◀◀◀◀◀◀◀◀◀◀◀◀◀░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│   │
│  │              ■■■■                              │   │
│  │          ■■■■■■■■■                              │   │
│  │       ■■■■■■■■■■■■■■■                           │   │
│  │    ■■■■■■■■■■■■■■■■■■■■■■■                       │   │
│  │ ■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■                   │   │
│  └────────────────────────────────────────────────────────┘   │
│                                                              │
│  Controls:                                                   │
│    [◀◀]  Prev Frame    [Next >]  [┇ Snapshot]  [▶ Replay]  │
│                                                              │
│  Current Frame Details:                                    │
│    • Time: 12:34:56.789                                   │
│    • Duration: 16.7ms                                     │
│    • Events: 3                                            │
│    • Mutations: 2                                         │
│    • Layouts: 1                                            │
│    • Repaints: 2                                           │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## 四、使用示例

### 4.1 基础用法

```go
package main

import (
    "github.com/wwsheng009/mint/devtools"
    devtools_client "github.com/wwsheng009/mint/devtools/client"
)

func main() {
    // 1. 创建 DevTools
    dt := devtools.New()

    // 2. 创建调试面板
    panel := devtools_client.NewTuiDebugPanel(dt)
    panel.Enable()

    // 3. 创建应用
    app := NewMyApp()

    // 4. 主循环
    for {
        dt.BeginFrame()

        // 应用更新
        app.Update()

        // 渲染
        screen := app.Render()

        // 收集布局数据
        dt.CollectLayout(screen)

        // 收集重绘数据
        dt.CollectRepaint(screen.DirtyRegions,
            screen.ChangedCells,
            screen.TotalCells)

        dt.EndFrame()

        // 显示调试面板 (按 Ctrl+D)
        if isDebugKeyPressed() {
            renderDebugPanel(panel)
        }
    }
}

func renderDebugPanel(panel *devtools_client.TuiDebugPanel) {
    // 渲染面板到屏幕
    output := panel.Render()
    print(output)
}
```

### 4.2 组件高亮

```go
// 当用户悬停在组件上时高亮显示
func onHover(nodeID string, rect Rect) {
    dt.Highlight(nodeID, rect.X, rect.Y, rect.Width, rect.Height)
}

// 在渲染时绘制高亮边框
func renderWithHighlight(screen Screen, dt *devtools.DevTools) {
    // 正常渲染
    screen.Render()

    // 绘制调试高亮
    overlay := dt.GetOverlay()
    for _, highlight := range overlay.GetHighlights() {
        drawDebugBorder(screen, highlight.Region, highlight.Color)
    }

    // 清理过期的高亮
    overlay.Update()
}
```

### 4.3 命令面板

```
DevTools Command Line
> inspect button1

Node: button1
Type: Button
Position: x: 10, y: 5, width: 20, height: 3
Properties:
  text: "Submit"
  enabled: true
  visible: true
Styles:
  bg: blue
  fg: white
  bold: true
Children: none
Metrics:
  LayoutTime: 0.5ms
  PaintTime: 0.2ms
  RepaintCount: 5

> highlight button1 green

Highlighted node button1 with color green

> stats

Statistics:
  Total Frames: 5000
  Avg Frame Time: 16.7ms
  FPS: 59.8
  Events: 234
  Mutations: 89
  Layouts: 123
  Repaints: 456
```

### 4.4 时间旅行使用

```go
// 创建时间旅行管理器
tt := devtools_client.NewTimeTravelClient(dt)

// 记录当前状态
tt.RecordSnapshot("before_action")

// 执行操作
button.Click()

// 记录操作后状态
tt.RecordSnapshot("after_action")

// 时间旅行到之前
tt.RestoreSnapshot("before_action")

// 自动回放
replayer := tt.NewReplayer()
replayer.ReplayRange("before_action", "after_action")
```

---

## 五、集成方法

### 5.1 在应用中集成

```go
package app

import (
    "github.com/wwsheng009/mint/devtools"
    devtools_client "github.com/wwsheng009/mint/devtools/client"
)

type Application struct {
    // 原有字段
    root      Component
    screen    *Screen

    // DevTools 字段
    devtools     *devtools.DevTools
    debugPanel   *devtools_client.TuiDebugPanel
    showDebug    bool
}

func NewApplication() *Application {
    dt := devtools.New()

    return &Application{
        root:       NewRootComponent(),
        screen:     NewScreen(),
        devtools:   dt,
        debugPanel: devtools_client.NewTuiDebugPanel(dt),
        showDebug:   false,
    }
}

func (app *Application) Run() {
    // 启用 DevTools
    app.devtools.Enable()
    app.debugPanel.Enable()

    for {
        // 处理输入
        event := app.readEvent()

        // 检查调试快捷键
        if event.Type == KeyEvent && event.Key == 'D' && event.Ctrl {
            app.showDebug = !app.showDebug
            continue
        }

        // 运行应用
        app.runFrame(event)
    }
}

func (app *Application) runFrame(event Event) {
    app.devtools.BeginFrame()
    defer app.devtools.EndFrame()

    // 更新
    app.root.HandleEvent(event)
    app.root.Update()

    // 布局
    layoutResult := app.root.Layout()
    app.devtools.CollectLayout(layoutResult)

    // 渲染
    screen := app.root.Render()
    app.devtools.CollectRepaint(
        screen.DirtyRegions,
        screen.ChangedCells,
        screen.TotalCells,
    )

    // 显示调试面板
    if app.showDebug {
        app.renderDebugPanel()
    }
}

func (app *Application) renderDebugPanel() {
    // 更新面板数据
    panel := app.debugPanel

    // 获取当前帧信息
    frameID := app.devtools.GetCurrentFrame()
    panel.SetSelectedFrame(frameID)

    // 渲染面板
    output := panel.Render()

    // 叠加显示在主屏幕上
    app.screen.RenderOverlay(output)
}
```

### 5.2 快捷键处理

```go
type DebugKeyBindings struct {
    TogglePanel      rune // Ctrl+D
    ToggleTimeline   rune // Ctrl+T
    ToggleCausal     rune // Ctrl+C
    ToggleSnapshot   rune // Ctrl+S
    ToggleReplay     rune // Ctrl+R
    StepBack         rune // Ctrl+Left
    StepForward      rune // Ctrl+Right
    InspectMode      rune // F12
}

func (app *Application) handleDebugKey(key rune) bool {
    bindings := &DebugKeyBindings{
        TogglePanel:    'D',
        ToggleTimeline: 'T',
        ToggleCausal:   'C',
        ToggleSnapshot: 'S',
        ToggleReplay:   'R',
    }

    switch key {
    case bindings.TogglePanel:
        app.showDebug = !app.showDebug
        return true

    case bindings.ToggleTimeline:
        app.debugPanel.ToggleTimeline()
        return true

    case bindings.ToggleCausal:
        app.debugPanel.ToggleCausal()
        return true

    case bindings.StepBack:
        // 时间旅行: 上一帧
        app.timeTravelToFrame(app.debugPanel.GetSelectedFrame() - 1)
        return true

    case bindings.StepForward:
        // 时间旅行: 下一帧
        app.timeTravelToFrame(app.debugPanel.GetSelectedFrame() + 1)
        return true
    }

    return false
}
```

### 5.3 命令行接口

```go
// DevTools 命令处理
type DevToolsCommand struct {
    Name string
    Args []string
}

func (app *Application) handleDevToolsCommand(cmd DevToolsCommand) string {
    handler := devtools_client.NewCommandHandler(app.debugPanel)
    return handler.Execute(cmd.Name + " " + strings.Join(cmd.Args, " "))
}

// 示例命令
func ExampleCommands() {
    // 检查组件
    handleCommand("inspect button1")

    // 高亮组件
    handleCommand("highlight button2 red")

    // 选择帧
    handleCommand("frame 1234")

    // 显示统计
    handleCommand("stats")

    // 时间旅行
    handleCommand("goto 1200")
    handleCommand("snapshot save before_edit")
    handleCommand("restore before_edit")
}
```

---

## 六、高级用法

### 6.1 自定义可视化

```go
// 创建自定义视图
type CustomView struct {
    panel *devtools_client.TuiDebugPanel
}

func (cv *CustomView) RenderCustomMetrics() string {
    // 收集自定义指标
    metrics := cv.collectCustomMetrics()

    // 渲染自定义视图
    var builder strings.Builder
    builder.WriteString("│ Custom Metrics\n")
    builder.WriteString("│ ─────────────\n")

    for name, value := range metrics {
        builder.WriteString(fmt.Sprintf("│ %s: %v\n", name, value))
    }

    return builder.String()
}
```

### 6.2 导出数据

```go
// 导出为 JSON
func ExportDevToolsData(dt *devtools.DevTools) ([]byte, error) {
    exporter := devtools_client.NewDataExporter(dt)
    return exporter.ToJSON()
}

// 导出为图表
func ExportCharts(dt *devtools.DevTools, outputPath string) error {
    exporter := devtools_client.NewChartExporter(dt)
    return exporter.ExportFlamegraph(outputPath)
}
```

### 6.3 远程调试

```go
// WebSocket 远程调试
func StartRemoteDebugServer(dt *devtools.DevTools, addr string) error {
    handler := devtools_client.NewWebSocketHandler(dt)

    http.HandleFunc("/ws", handler.HandleWebSocket)

    go func() {
        log.Fatal(http.ListenAndServe(addr, nil))
    }()

    return nil
}
```

---

## 七、故障排查

### 7.1 常见问题

| 问题 | 解决方法 |
|------|----------|
| 面板不显示 | 检查 `panel.Enable()` 是否调用 |
| 数据不更新 | 确保在每帧调用 `dt.EndFrame()` |
| 高亮不显示 | 检查 `overlay.Update()` 是否在渲染循环中调用 |
| 性能下降 | 使用运行级别 `dt.SetLevel(LevelLight)` 降低开销 |

### 7.2 调试技巧

```go
// 1. 启用详细日志
devtools.SetDefaultLogLevel(devtools.LevelDebug)

// 2. 检查统计信息
stats := dt.GetEventBus().GetStats()
log.Printf("Events: %d, Dropped: %d",
    stats.EventsSent.Load(),
    stats.BackpressureDrops.Load())

// 3. 验证因果链
graph := dt.GetCausalGraph()
summary := graph.GetFrameSummary()
log.Printf("Frame %d: E=%d M=%d L=%d R=%d",
    summary.FrameID,
    summary.EventCount,
    summary.MutationCount,
    summary.LayoutCount,
    summary.RepaintCount)
```

---

## 八、最佳实践

1. **生产环境**: 使用 `LevelNone` 或 `LevelLight`
2. **开发环境**: 使用 `LevelFull` 获取完整信息
3. **性能敏感**: 限制历史窗口大小
4. **内存限制**: 定期清理快照
5. **异步处理**: 使用非阻塞发送避免死锁
