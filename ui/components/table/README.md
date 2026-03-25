# Table

数据表格组件，适合后台列表、分页数据和可筛选数据集展示。

## 已支持

- columns / rows
- 排序、过滤、搜索
- 分页
- 单选 / 多选
- expandable rows
- fixed columns
- tree data
- `Field` / selection / page binding

## 示例

```go
ui.NewTableBuilder().
    Columns([]ui.TableColumn{
        {Title: "ID", Width: 6},
        {Title: "Name", Width: 20},
    }).
    Rows([][]string{
        {"1", "Mint"},
        {"2", "Fiber"},
    }).
    PageSize(10).
    ShowBorder(true).
    Build()
```

根包不提供 `ui.Table(...)` 快捷函数，因为该名字已被布局 helper 占用；表格请直接用 `ui.NewTableBuilder()`。
