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

## 数据应用快捷构造

对于后台、运维、管理台一类常见表格，可以使用根包的 `ui.DataTable(...)` 快速配置分页、选择绑定、搜索、空态和标准高亮样式：

```go
tableView := ui.DataTable(
    []ui.TableColumn{
        {Title: "Provider", Width: 20},
        {Title: "Status", Width: 12},
    },
    [][]string{
        {"openai", "healthy"},
        {"azure", "degraded"},
    },
    ui.DataTablePageSize(10),
    ui.DataTableSelectedIndex(state.SelectedProviderIndex),
    ui.DataTableSelectedField("selectedProviderIndex"),
    ui.DataTableSearch(state.Search),
    ui.DataTableEmptyText("No providers"),
    ui.DataTableOperationalStyle(),
)
```

如需扩展内置选项，可以实现 `ui.DataTableOption` 并修改公开的 `ui.DataTableConfig`。

## 测试入口

- 单测：`go test ./ui/components/table`
- 重点覆盖：`table_test.go` 中的排序 / 分页 / 过滤、expandable rows、fixed columns、tree data、字段同步与 `StateChangeIntent`
- E2E：`go test ./ui/e2e -run TestE2ETable`
