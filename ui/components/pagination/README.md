# Pagination

独立分页组件，复用 `table` 的 0-based `currentPage` 语义。

- `Total(total)` + `PageSize(size)` 计算页数
- `CurrentPage(page)` 走受控页码
- `PageForField(intent.BindField("currentPage"))` 发出页码变更
- `PageChangeIntent` 提供结构化分页变更事件
