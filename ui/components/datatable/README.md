# DataTable

`datatable/` 是面向业务数据列表的声明式组合组件，基于 `table/` builder 生成 VNode，不持有运行时状态。

## 适用场景

- 后台列表
- 运维资源表
- 带分页、搜索、选择绑定的管理台表格
- 服务端分页列表
- loading / error / empty 状态
- 基于业务 id 的稳定行选择和显式激活

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
    ui.DataTableRowKeys([]string{"provider.openai", "provider.azure"}),
    ui.DataTableSelectedKey("provider.azure"),
    ui.DataTableSelectedKeyField("selectedProviderKey"),
    ui.DataTableActivateKeyField("activatedProviderKey"),
    ui.DataTableSortState(state.ProviderSortColumn, state.ProviderSortDescending),
    ui.DataTableServerPagination(2, 25, 76),
    ui.DataTableOperationalStyle(),
)
```

## 运维状态

- `DataTableLoading(true)`：清空当前 rows，显示 loading 空态，并把 footer 设为 `Loading`。
- `DataTableErrorText("...")`：清空当前 rows，显示错误空态，并把 footer 设为 `Error · ...`。
- `DataTableServerPagination(page, pageSize, total)`：覆盖 table footer 为服务端分页摘要，例如 `Page 2/4 · Total 76 · Size 25`。
- `DataTableStatusText("...")`：直接覆盖 footer 文案，适合游标分页或自定义聚合摘要。

## 稳定选择

- `DataTableRowKeys(keys)`：按 source row index 提供稳定业务 key，适合 provider/key/token/job id。
- `DataTableSelectedKey(key)`：以 row key 控制当前高亮行，排序、过滤和刷新后不依赖易变行号。
- `DataTableSelectedKeyField(field)`：行选择变化时向字段写入当前 row key。
- `DataTableActivateKeyField(field)`：用户按 Enter 或确认当前行时向字段写入激活 row key。

## 受控排序

- `DataTableSortBy(column, descending)`：显式设置排序列和方向，适合调用方确定一定存在排序状态的场景。
- `DataTableSortState(column, descending)`：使用 `column < 0` 表示未排序，`column == 0 && descending == false` 仍保留为第一列升序。该入口适合直接转接业务 reducer 中的表格排序状态，避免把第 0 列升序误判成默认空状态。
- 客户端排序继承底层 `table` 的运维格式比较能力：RFC3339/API 时间文本会按时间排序，纯数字、千分位数字、百分比、`ms/s/m/h` 时长和 `x/y` 比例会按数值排序，适合日志时间、观察时间、负载、延迟、成功率、配额等列的当前页扫描。
- 服务端排序仍应由业务层把 `table.StateChangeIntent.SortColumn` 映射到明确支持的 API 字段；组件不发起 IO，也不假设后端支持任意列排序。

## Fiber-first 约束

- 只负责声明式 VNode 组合。
- 业务状态通过 `SelectedField(...)`、`SelectedKeyField(...)`、`ActivateKeyField(...)` 等 intent 绑定传递。
- 不在组件包内发起 IO、保存业务状态或直接处理运行时副作用。

## 测试

```powershell
go test ./ui/components/datatable
go test ./ui -run DataTable
go test ./ui/e2e -run "^TestE2EDataTable"
```
