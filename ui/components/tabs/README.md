# Tabs

标签页组件，适合多面板切换、文档页签和工具区分栏。

## 已支持

- top / bottom / left / right
- `line` / `card`
- closable tabs
- drag reorder
- hotkey hints
- 受控 / 非受控 active tab
- change / close / reorder intents

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
