# Tabs

标签页组件，适合多面板切换、文档页签和工具区分栏。

## 已支持

- top / bottom / left / right
- `line` / `card`，其中 card 变体会绘制顶部边框并使用更强的单行工作区样式
- closable tabs
- drag reorder
- hotkey hints
- active tab text marker
- 受控 / 非受控 active tab
- change / close / reorder intents
- workspace tabs preset for dense operational pages

## 示例

```go
ui.NewTabsBuilder().
    Tabs([]ui.TabItem{
        ui.NewTabItem("home", "Home"),
        ui.NewTabItem("logs", "Logs"),
    }).
    Card().
    ActiveTabID("home").
    Build()
```

快捷函数：

```go
ui.Tabs([]ui.TabItem{
    ui.NewTabItem("home", "Home"),
})
```

密集运维页面建议使用 `ui.WorkspaceTabs(...)`。该 preset 默认使用 card tabs，启用横向换行，并采用更克制的高对比 active rail 配合 reverse、bold 和 muted inactive labels，让当前工作区在窄终端或弱色彩主题下仍然清晰；同时通过 `FieldIntent` 绑定页面状态。

```go
ui.WorkspaceTabs(
    "runtime.workspace.tabs",
    "runtime.workspace.tabs",
    "runtimeWorkspace",
    state.RuntimeWorkspaceTab,
    []ui.TabItem{
        ui.NewTabItem("sections", "Sections"),
        ui.NewTabItem("operations", "Operations"),
        ui.NewTabItem("http", "HTTP"),
        ui.NewTabItem("concurrency", "Concurrency"),
    },
    126,
)
```

适用边界：当一个 TUI 页面有多个同级主工作面时，优先用 workspace tabs 只渲染当前工作区，避免把表格、详情、reason 表单和操作按钮纵向堆到终端屏幕外。preset 的 wrap 只解决 tab bar 自身在窄宽度下的可见性，不替代工作区拆分或页面 viewport。不要用文本型 `ScrollView` 包裹复杂交互页面来替代 workspace；真正交互式滚动容器需要保留 child tree、layout、hitmap、focus 和事件派发能力。
