# Toolbar

`toolbar` 是面向数据页、运维页和管理台页面的 Fiber-first 操作工具栏组件。它用于承载当前范围、状态摘要、刷新、打开详情、危险操作入口等页面级命令。

## 能力

- 三槽布局：left / center / right。
- 标题区：`Title(...)` 与 `TitleWidth(...)`。
- 操作项：text、badge、button、separator、custom。
- intent-first：按钮通过 `intent.Intent` 派发动作，不持有闭包。
- 禁用态：`WithDisabled(...)`。
- 帮助文本：`WithHelp(...)` / `WithTooltip(...)`，在 `UseStatusBar(true)` 模式下复用 `StatusBar` overlay help。
- 密集模式：`Dense(true)` 会使用小尺寸按钮。
- 子目录组件结构：`ui/components/toolbar/`。
- Fiber-first：`VNode + Instance + RuntimeChildrenProvider`。

## 示例

```go
toolbar.NewBuilder().
    Key("lb.toolbar").
    Title("Load Balancer").
    TitleWidth(16).
    Left(toolbar.Text("scope", "group: default").WithWidth(20)).
    Center(toolbar.Badge("state", "degraded").WithColors("black", "yellow")).
    Right(toolbar.Button("refresh", "Refresh", RefreshIntent{}).Primary()).
    Right(toolbar.Button("reset", "Reset Runtime", ResetIntent{}).Danger().WithDisabled(true).WithHelp("Requires selected target and reason")).
    Build()
```

## 状态栏模式

需要底部快捷键或 hover 帮助时，可以开启 `UseStatusBar(true)`：

```go
toolbar.NewBuilder().
    Key("manager.status").
    Title("ai-gateway-manager").
    UseStatusBar(true).
    Left(toolbar.Badge("mode", "ADMIN")).
    Center(toolbar.Text("focus", "F2 Load Balancer")).
    Right(toolbar.Button("help", "F10 Help", HelpIntent{}).WithHelp("Open contextual help")).
    Build()
```

## 运维页面建议

- 顶部 `Toolbar` 只放当前页面最常用的命令和范围信息。
- 搜索、状态、group/provider 等过滤条件优先使用 `FilterBar`。
- 危险操作可以在 `Toolbar` 中提供入口，但应跳转到 `ConfirmDialog`、`Popconfirm`、`Modal` 或业务确认流程。
- 不要把复杂表单永久放入首屏工具栏。

## 测试

- 单测：`go test ./ui/components/toolbar`
- E2E：`go test ./ui/e2e -run "^TestE2EToolbar"`
