# List

列表组件，适合菜单、结果列表、搜索结果和可滚动行选择场景。

## 已支持

- 纯 `rows` 和结构化 `items`
- 单选 / 多选
- 搜索过滤
- 单列文本排序：`SortAscending()` / `SortDescending()` / `SortRows(enabled, descending)`
- `header`
- border / scrollbar
- `Field` 绑定与 selection intent
- `VirtualList` bridge

## 示例

```go
ui.List().
    Header("Tasks").
    Rows([]string{"Build", "Test", "Deploy"}).
    SortAscending().
    MaxRows(5).
    ShowBorder(true).
    Build()
```

需要 richer row 展示时可以切到 `Items(...)`，大数据量场景可以直接 `BuildVirtualList()`。

根包 SDK 暴露 `ui.ListItem`、`ui.NewListItem(...)` 和 `ui.NewListItemWithDescription(...)`，业务代码通常不需要直接导入 `ui/components/list`：

```go
ui.List().
    Header("Providers").
    Items([]ui.ListItem{
        ui.NewListItem("openai").WithPrefix("[ok]").WithDescription("healthy"),
        ui.NewListItemWithDescription("anthropic", "degraded").WithPrefix("[warn]"),
    }).
    SortAscending().
    Build()
```

`List` 的排序面向单列候选列表，排序 key 使用 `RowItem.Title`，没有 title 时使用行文本。多列数据、列头点击排序、服务端分页或稳定 row key 选择应使用 `Table` / `DataTable`。
