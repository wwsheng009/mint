# List

列表组件，适合菜单、结果列表、搜索结果和可滚动行选择场景。

## 已支持

- 纯 `rows` 和结构化 `items`
- 单选 / 多选
- 搜索过滤
- `header`
- border / scrollbar
- `Field` 绑定与 selection intent
- `VirtualList` bridge

## 示例

```go
ui.List().
    Header("Tasks").
    Rows([]string{"Build", "Test", "Deploy"}).
    MaxRows(5).
    ShowBorder(true).
    Build()
```

需要 richer row 展示时可以切到 `Items(...)`，大数据量场景可以直接 `BuildVirtualList()`。
