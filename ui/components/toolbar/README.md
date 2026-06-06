# Toolbar

`toolbar` 是面向数据页、运维页和管理台页面的 Fiber-first 操作工具栏组件。它用于承载当前范围、状态摘要、刷新、打开详情、危险操作入口等页面级命令。

## 能力

- 三槽布局：left / center / right。
- 标题区：`Title(...)` 与 `TitleWidth(...)`。
- 操作项：text、badge、button、separator、custom。
- 受控下拉菜单项：`Dropdown(...)` / `Menu(...)` 会把 toolbar button 与 anchored `menu.Popup` 组合起来。
- intent-first：按钮通过 `intent.Intent` 派发动作，不持有闭包。
- 禁用态：`WithDisabled(...)`。
- 禁用原因：`WithDisabledReason(...)` 会同时禁用操作并提供 hover/statusbar help；`WithDisabledReasonIf(condition, reason)` 适合按 loading、selection、lookup 等业务条件顺序追加禁用原因。
- 多前置条件禁用原因：`WithDisabledReasons(...)` / `JoinDisabledReasons(...)` 会过滤空原因并合并说明，适合 reason、选中对象、loading 等条件同时影响同一个操作按钮。
- 操作目标摘要：`ActionTargetSummary(...)` / `ActionTargetSummaryWithScope(...)` / `CompactActionTargetPart(...)` 用于危险操作或审计操作区附近展示 `scope key · group default · provider openai · key ****` 这类目标链路，避免把操作目标误放进筛选摘要语义。
- 帮助文本：`WithHelp(...)` / `WithTooltip(...)`，普通工具栏会渲染 hover tooltip，`UseStatusBar(true)` 模式下复用 `StatusBar` overlay help。
- 密集模式：`Dense(true)` 会使用小尺寸按钮。
- 子目录组件结构：`ui/components/toolbar/`。
- Fiber-first：`VNode + Instance + RuntimeChildrenProvider`。
- 运维语义预设：`KeyValue(...)`、`MutedKeyValue(...)`、`Scope(...)`、`Selection(...)`、`StateBadge(...)`、`BusyBadge(...)`、`ErrorBadge(...)` 可直接构造 endpoint、scope、selection、运行态和错误态等常见工具栏片段。
- Shell 预设：`ShellHeader(...)` / `ShellNav(...)` 用于标准化运维 TUI 顶层的应用标识、endpoint/auth 状态、全局操作和页面导航行。
- 分页导航配置：`PaginationConfig` 可用同一份分页上下文生成 `Prev()`、`Next()` 与 `Controls()`，适合日志、token、jobs 和 alert history 这类服务端分页工具栏。
- 选择导航配置：`SelectionConfig` 可用同一份选择上下文生成 `Prev()`、`Next()`、`Controls()` 与 `ActionControls(...)`，适合详情面板、下钻面板和测试断言共享同一套边界/禁用原因。
- 操作分组布局：`ActionGroupWithLayout(...)` 与 `ActionGroupConfig.Width` / `TitleWidth` 可按页面宽度和分组标题长度调整操作组，默认仍保持标准 `ActionGroup(...)` 布局。

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

顶层 Shell 建议把环境确认、会话状态和全局动作放在第一行，把页面导航放在第二行：

```go
toolbar.ShellHeader(
    "manager.header",
    "ai-gateway-manager",
    "http://localhost:8080",
    "admin: ops",
    toolbar.Button("refresh", "Refresh", RefreshIntent{}).Primary(),
    toolbar.Button("logout", "Logout", LogoutIntent{}),
)

toolbar.ShellNav("manager.nav", []toolbar.Item{
    toolbar.Button("nav-overview", "F1 Overview", NavigateIntent{Page: "overview"}).Primary(),
    toolbar.Button("nav-actions", "F8 Actions", NavigateIntent{Page: "actions"}),
})
```

服务端分页工具栏需要 summary、Prev 和 Next 共用同一份分页上下文时，优先使用 `PaginationConfig`。如果页面要先归一化 page/pageSize/total/totalPages，并把 search、provider、status 等 scope 放进摘要，可使用 `NormalizePaginationScope(...)` 和 `PaginationControlsWithScope(...)`：

```go
scope := toolbar.NormalizePaginationScope(state.Page, state.PageSize, state.Total, state.TotalPages, len(rows))

toolbar.PaginationControlsWithScope(
    "logs.pagination",
    126,
    scope,
    PrevPageIntent{},
    NextPageIntent{},
    state.Busy,
    "Request logs are loading.",
    84,
    toolbar.PageSummaryPart{Label: "search", Value: state.Search},
)
```

User-entered scope values can be width-bounded with `CompactPageSummaryPart(...)`
so a long search/provider/status value does not crowd out Prev/Next controls:

```go
toolbar.PaginationControlsWithScope(
    "logs.pagination",
    126,
    scope,
    PrevPageIntent{},
    NextPageIntent{},
    state.Busy,
    "Request logs are loading.",
    84,
    toolbar.CompactPageSummaryPart("search", state.Search, 24),
)
```

For optional filters, prefer `PageSummaryPartIfValue(...)` or
`CompactPageSummaryPartIfValue(...)` so inactive search/provider values are
omitted instead of occupying space as `search -`:

```go
toolbar.PaginationControlsWithScope(
    "tokens.pagination",
    126,
    scope,
    PrevPageIntent{},
    NextPageIntent{},
    state.Busy,
    "Provider tokens are loading.",
    84,
    toolbar.CompactPageSummaryPartIfValue("provider", state.Provider, 24),
    toolbar.CompactPageSummaryPartIfValue("search", state.Search, 24),
)
```

For select filters with a known default value, prefer
`PageSummaryPartUnless(...)` or `CompactPageSummaryPartUnless(...)` so defaults
such as `status=all` stay visible in the filter control but are omitted from the
pagination summary until they narrow the result set:

```go
toolbar.PaginationControlsWithScope(
    "jobs.pagination",
    126,
    scope,
    PrevPageIntent{},
    NextPageIntent{},
    state.Busy,
    "Jobs are loading.",
    84,
    toolbar.PageSummaryPartUnless("status", state.Status, "all"),
    toolbar.CompactPageSummaryPartUnless("provider", state.Provider, "all", 24),
)
```

选中对象详情区需要 Up/Down 和当前对象动作时，优先把选择上下文放入 `SelectionConfig`：

```go
selection := toolbar.SelectionConfig{
    Key:           "logs.selection.actions",
    Title:         "Log",
    Width:         56,
    PrevIntent:    SelectLogIntent{Delta: -1},
    NextIntent:    SelectLogIntent{Delta: 1},
    Busy:          state.Busy,
    Index:         state.SelectedLogIndex,
    Total:         len(logs),
    ItemLabel:     "log",
    LoadingReason: "Request logs are loading.",
}

selection.ActionControls(
    toolbar.Button("open-trace", "Open Trace", OpenTraceIntent{}).Primary(),
)
```

危险操作区需要展示当前操作目标时，优先使用 target summary，而不是复用 filter 或 pagination summary：

```go
summary := toolbar.ActionTargetSummary(
    toolbar.CompactActionTargetPart("group", groupName, 24),
    toolbar.CompactActionTargetPart("provider", providerName, 28),
    toolbar.CompactActionTargetPart("key", maskedKeyID, 28),
)

toolbar.ActionGroups("actions.groups", groups, summary, "No actions matched.")
```

如果同一个操作区同时包含 group/provider/key 多种目标层级，使用
`ActionTargetSummaryWithScope(...)` 把当前操作层级放在摘要最前面：

```go
summary := toolbar.ActionTargetSummaryWithScope(
    activeScope,
    toolbar.CompactActionTargetPart("group", groupName, 24),
    toolbar.CompactActionTargetPart("provider", providerName, 28),
    toolbar.CompactActionTargetPart("key", maskedKeyID, 28),
)
```

如果同一个操作区同时展示多个可点击层级，并且摘要需要说明当前可见操作范围，使用
`ActionTargetSummaryWithScopes(...)`：单一 scope 会保持 `scope key` 旧格式，多 scope 会显示
`scopes global/group/provider/key`，后面继续追加当前选中目标链路。

```go
summary := toolbar.ActionTargetSummaryWithScopes(
    []string{"global", "group", "provider", "key"},
    toolbar.CompactActionTargetPart("endpoint", endpoint, 40),
    toolbar.CompactActionTargetPart("group", groupName, 24),
    toolbar.CompactActionTargetPart("provider", providerName, 28),
    toolbar.CompactActionTargetPart("key", maskedKeyID, 28),
)
```

`ActionTargetSummary` 会把空值显示为 `-`，让操作者在准备操作前就能看到缺失目标；敏感值仍应由调用方先脱敏，再传给 summary helper。

当操作区需要按影响范围拆成多行时，可在 `ActionGroupConfig` 上设置
`Width` 与 `TitleWidth`，让 group/provider/key 等短标题不占用过多空间，
也避免较长按钮行被固定默认宽度截断：

```go
toolbar.ActionGroups("actions.groups", []toolbar.ActionGroupConfig{
    {
        Key:        "actions.key",
        Title:      "Key",
        Width:      126,
        TitleWidth: 10,
        Items: []toolbar.Item{
            toolbar.Button("disable-key", "Disable Key", DisableKeyIntent{}).Danger(),
        },
    },
}, summary, "No actions matched.")
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
