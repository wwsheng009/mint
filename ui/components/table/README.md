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

## 状态语义

- 默认交互状态由组件内部维护。
- 需要外部受控时，可显式传入 `CurrentPage(...)`、`SortBy(...)`、`SelectedIndex(...)`、`CheckedIndices(...)`、`ExpandedIndices(...)`。
- 需要监听完整交互快照时，给表格设置 `ComponentID(...)` 并订阅 `table.StateChangeIntent`。
- 字段联动可按职责拆开：`ForField(...)` 绑定当前选中行，`SelectionForField(...)` 绑定勾选集合，`PageForField(...)` 绑定页码，`OnExpand(...)` 处理展开态变化。

## 示例

```go
ui.NewTableBuilder().
    ComponentID("orders.table").
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

## 测试入口

- 单测：`go test ./ui/components/table`
- 重点覆盖：`table_test.go` 中的排序 / 分页 / 过滤、expandable rows、fixed columns、tree data、字段同步与 `StateChangeIntent`
- E2E：`go test ./ui/e2e -run TestE2ETable`
