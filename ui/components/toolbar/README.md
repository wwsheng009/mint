# Toolbar

`toolbar` 是面向数据页、运维页和管理台页面的 Fiber-first 操作工具栏组件。它用于承载当前范围、状态摘要、刷新、打开详情、危险操作入口等页面级命令。

## 能力

- 三槽布局：left / center / right。
- 标题区：`Title(...)` 与 `TitleWidth(...)`。
- 操作项：text、badge、button、separator、custom。
- 受控下拉菜单项：`Dropdown(...)` / `Menu(...)` 会把 toolbar button 与 anchored `menu.Popup` 组合起来。
- intent-first：按钮通过 `intent.Intent` 派发动作，不持有闭包。
- 禁用态：`WithDisabled(...)`。
- 禁用原因：`WithDisabledReason(...)` 会同时禁用操作并提供 hover/statusbar help。
- 帮助文本：`WithHelp(...)` / `WithTooltip(...)`，普通工具栏会渲染 hover tooltip，`UseStatusBar(true)` 模式下复用 `StatusBar` overlay help。
- 密集模式：`Dense(true)` 会使用小尺寸按钮。
- 子目录组件结构：`ui/components/toolbar/`。
- Fiber-first：`VNode + Instance + RuntimeChildrenProvider`。
- 运维语义预设：`KeyValue(...)`、`MutedKeyValue(...)`、`Scope(...)`、`Selection(...)`、`StateBadge(...)`、`BusyBadge(...)`、`ErrorBadge(...)` 可直接构造 endpoint、scope、selection、运行态和错误态等常见工具栏片段。

## 示例

```go
toolbar.NewBuilder().
    Key("lb.toolbar").
    Title("Load Balancer").
    TitleWidth(16).
    Left(toolbar.Scope("group: default").WithWidth(24)).
    Center(toolbar.StateBadge("state", "degraded")).
    Right(toolbar.Button("refresh", "Refresh", RefreshIntent{}).Primary()).
    Right(toolbar.Button("reset", "Reset Runtime", ResetIntent{}).Danger().WithDisabledReason("Requires selected target and reason")).
    Build()
```

## 运维预设

```go
toolbar.Endpoint("http://localhost:8080").WithWidth(36)
toolbar.Scope("group: default").WithWidth(24)
toolbar.Selection("provider/openai").WithWidth(28)
toolbar.StateBadge("state", "healthy")
toolbar.StateBadge("state", "pending_restart")
toolbar.BusyBadge("sync", "refreshing")
toolbar.ErrorBadge("error", "failed")
```

默认状态映射复用 `statusbar.DefaultTone(...)`：

- `healthy` / `active` / `available` / `effective` / `enabled` / `ready` / `running` / `in_sync` -> normal
- `degraded` / `rate_limited` / `pending_restart` / `pending` / `warning` / `lagging` / `reloading` -> warn
- `unhealthy` / `disabled` / `unauthorized` / `unavailable` / `failed` / `error` / `blocked` / `out_of_sync` -> error
- `processing` / `loading` / `syncing` / `refreshing` -> info
- 未识别状态 -> neutral

## 下拉菜单

`Dropdown(...)` 适合工具栏上的更多操作、过滤预设、页面级命令菜单。打开态由应用状态控制；点击按钮默认发出 `menu.OpenMenuIntent`，业务 reducer 可以据此把对应 item 渲染为 `MenuOpen(true)`。

```go
toolbar.NewBuilder().
    Key("runtime.toolbar").
    Title("Runtime").
    Right(toolbar.Dropdown("actions", "Actions", menu.Items(
        menu.Action("reload", "Reload Runtime", ReloadIntent{}).WithDescription("Reload runtime configuration"),
        menu.Action("diagnostics", "Open Diagnostics", OpenDiagnosticsIntent{}),
    ), state.ActionsMenuOpen).
        WithMenuID("runtime-actions").
        WithMenuPlacement(menu.PlacementBottomStart).
        WithMenuDescriptions(true)).
    Build()
```

说明：

- `WithMenuID(...)` 提供稳定菜单 ID；未设置时使用 toolbar item key 派生。
- `WithMenuPlacement(...)` 支持 `bottom-start` / `bottom-end` / `top-start` / `top-end` / `right-start` / `left-start` / `auto`。
- `WithMenuMinWidth(...)`、`WithMenuMaxHeight(...)`、`WithMenuShortcuts(...)`、`WithMenuDescriptions(...)` 会透传给 `menu.Popup`。
- 下拉菜单在普通 toolbar 行模式下自动锚定到按钮；`UseStatusBar(true)` 模式仍只保留按钮/intent 行为，复杂菜单建议放在页面顶部工具栏。

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
- 多个低频命令优先收进 `Dropdown`，避免在首屏形成操作墙。
- 搜索、状态、group/provider 等过滤条件优先使用 `FilterBar`。
- 危险操作可以在 `Toolbar` 中提供入口，但应跳转到 `ConfirmDialog`、`Popconfirm`、`Modal` 或业务确认流程。
- 不要把复杂表单永久放入首屏工具栏。

## 测试

- 单测：`go test ./ui/components/toolbar`
- E2E：`go test ./ui/e2e -run "^TestE2EToolbar"`
