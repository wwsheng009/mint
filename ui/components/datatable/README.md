# DataTable

`datatable/` 是面向业务数据列表的声明式组合组件，基于 `table/` builder 生成 VNode，不持有运行时状态。

## 适用场景

- 后台列表
- 运维资源表
- 带分页、搜索、选择绑定的管理台表格
- 服务端分页列表
- loading / error / empty 状态

## 示例

```go
node := ui.DataTable(
    []ui.TableColumn{
        {Title: "Provider", Width: 20},
        {Title: "Status", Width: 12},
    },
    [][]string{
        {"openai", "healthy"},
        {"azure", "degraded"},
    },
    ui.DataTablePageSize(10),
    ui.DataTableSelectedField("selectedProvider"),
    ui.DataTableServerPagination(2, 25, 76),
    ui.DataTableOperationalStyle(),
)
```

## 运维状态

- `DataTableLoading(true)`：清空当前 rows，显示 loading 空态，并把 footer 设为 `Loading`。
- `DataTableErrorText("...")`：清空当前 rows，显示错误空态，并把 footer 设为 `Error · ...`。
- `DataTableServerPagination(page, pageSize, total)`：覆盖 table footer 为服务端分页摘要，例如 `Page 2/4 · Total 76 · Size 25`。
- `DataTableStatusText("...")`：直接覆盖 footer 文案，适合游标分页或自定义聚合摘要。

## Fiber-first 约束

- 只负责声明式 VNode 组合。
- 业务状态通过 `SelectedField(...)` 等 intent 绑定传递。
- 不在组件包内发起 IO、保存业务状态或直接处理运行时副作用。

## 测试

```powershell
go test ./ui/components/datatable
go test ./ui -run DataTable
go test ./ui/e2e -run "^TestE2EDataTable"
```
