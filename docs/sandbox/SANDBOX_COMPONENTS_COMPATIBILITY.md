# Sandbox 与当前 UI 组件兼容性

本文记录 Sandbox 测试系统与当前 `ui/`、`runtime/ui`、`ui/components` 架构的兼容关系。旧文档曾使用 `app.ButtonBuilder().OnClick()`、`app.Text()` 等 API；当前源码中 `app` 包主要保留 `app.Run` / `app.Quit` 兼容入口，组件创建应优先使用 `ui` 包。

## 当前结论

- 测试入口：推荐 `ui.RunTest(...)` 和 `ui.RunTestWithSandbox(...)`。
- 组件入口：推荐 `ui.Text(...)`、`ui.NewTextBuilder(...)`、`ui.NewButtonBuilder(...)`、`ui.NewInputBuilder(...)` 等。
- Button 行为：当前 Fiber-first 模型使用 `OnPress(intent.Intent)`，不是旧 `OnClick(func())`。
- Sandbox 高级功能仍在 `sandbox/`、`sandbox/mock`、`sandbox/replay`、`sandbox/testing` 中。

## 当前架构

```text
runtime/ui/
  核心 VNode、Fiber、ComponentInstance、hooks、focus 等类型

ui/
  应用入口、测试入口、组件快捷函数、builder 工厂、类型别名

ui/components/
  组件实现：VNode + Instance + Builder + tests

sandbox/
  mock / real / replay / testing 等测试基础设施
```

## 测试入口

| API | 状态 | 说明 |
|---|---|---|
| `ui.RunTest(app, opts...)` | 推荐 | 运行测试应用，返回 `*ui.TestableApp` |
| `ui.RunTestWithSandbox(app, opts...)` | 推荐 | 使用 MockSandbox 作为事件源，支持事件注入、快照、队列等能力 |
| `ui.TestRun(app, opts...)` | legacy | 仍存在于 `ui/test.go`，用于兼容旧测试 |
| `sandbox/mock.New(width, height)` | 可用 | 直接使用底层 MockSandbox |
| `sb.Helper()` | 可用 | 链式测试辅助器 |

常用尺寸选项：

```go
ui.WithSize(80, 24)
ui.WithWidth(80)
ui.WithHeight(24)
```

## 当前组件 API

| 组件类型 | 当前 API |
|---|---|
| Text | `ui.Text("content")`、`ui.NewTextBuilder("content").Build()` |
| Button | `ui.Button("label")`、`ui.NewButtonBuilder("label").OnPress(intent).Build()` |
| Input | `ui.NewInputBuilder().Placeholder("...").Build()` |
| Checkbox | `ui.NewCheckboxBuilder().Label("...").Build()` |
| Layout | `ui.VStack(...)`、`ui.HStack(...)`、`ui.Wrap(...)`、`ui.WrapBuilder(...)` |
| VirtualList | `ui.NewVirtualListBuilder()`、`ui.VirtualList()`、`ui.VirtualListOfSize(...)` |
| ScrollView | `ui.NewScrollViewBuilder()`、`ui.ScrollView(...)`、`ui.ScrollBordered(...)` |

示例：

```go
func App() ui.VNode {
    return ui.VStack(
        ui.NewTextBuilder("Hello").FgColor("green").Bold(true).Build(),
        ui.NewButtonBuilder("Submit").
            OnPress(intent.Click("submit")).
            Build(),
        ui.NewInputBuilder().
            Placeholder("Name").
            Build(),
    )
}
```

## Sandbox 能力

| 功能 | 入口 |
|---|---|
| 键盘注入 | `InjectKey`、`InjectSpecialKey` 等测试 helper |
| 鼠标注入 | MockSandbox / TestableApp 对应 mouse 注入 API |
| 事件录制 | `sandbox/replay`、示例 `examples/sandbox/01_event_recording` |
| 快照 | `Snapshot()` / `Restore()` 相关 Sandbox API |
| 队列统计 | `QueueStats()` |
| Dump buffer | `examples/sandbox/dump_buffer` |

推荐从示例目录看当前用法：

```text
examples/sandbox/01_event_recording
examples/sandbox/02_snapshot
examples/sandbox/03_test_helper
examples/sandbox/04_queue_stats
examples/sandbox/05_injection_strategy
examples/sandbox/06_comprehensive
examples/sandbox/dump_buffer
```

## 迁移对照

| 旧写法 | 当前写法 |
|---|---|
| `app.Text("content")` | `ui.Text("content")` |
| `app.NewTextBuilder("content")` | `ui.NewTextBuilder("content")` |
| `app.ButtonBuilder("label").OnClick(fn).Build()` | `ui.NewButtonBuilder("label").OnPress(intent).Build()` |
| `app.InputBuilder()` | `ui.NewInputBuilder()` |
| `ui.Button("label", onClick)` | `ui.NewButtonBuilder("label").OnPress(intent).Build()` |
| `ui.TestRun(app, ui.TestWithSize(w, h))` | `ui.RunTest(app, ui.WithSize(w, h))` |

如果测试只需要展示静态按钮，可以使用：

```go
ui.Button("Click")
```

如果按钮需要触发行为，应使用 intent：

```go
ui.ButtonWithIntent("Click", intent.Click("click"))
```

或：

```go
ui.NewButtonBuilder("Click").
    OnPress(intent.Click("click")).
    Build()
```

## 当前测试建议

```bash
go test ./sandbox/... -count=1
go test ./examples/sandbox/... -count=1
go test ./ui -run RunTest -count=1
```

若资源有限，不建议把 `go test ./...` 当作本地必跑基线；可按 sandbox、examples、ui/e2e 分层执行。

## 维护约束

- 新文档不要再推荐 `app.ButtonBuilder`、`app.InputBuilder` 或 `OnClick(func())`。
- `app.Run` / `app.Quit` 仍可作为兼容入口，但组件创建以 `ui` 包为准。
- 交互行为优先用 `intent`、`runtime/action`、store/reducer，而不是在 VNode 中保存闭包。
