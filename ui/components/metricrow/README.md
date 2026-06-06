# MetricRow

`metricrow/` 是紧凑指标行组合组件，基于 `panel/` 和 `text/` 生成横向 VNode 树。

## 适用场景

- 运维概览
- 仪表盘摘要
- 表格上方的关键指标区

## 示例

```go
node := ui.MetricRow(
    []ui.MetricItem{
        ui.MetricRowValue("Runtime", "healthy"),
        ui.MetricRowCount("Providers", 12),
        ui.MetricRowCount("Failed", 0),
        ui.MetricRowCompactValue("Trace", traceID, 18),
        ui.MetricRowCompactValue("Filters", filterSummary, 42).WithWidth(46),
    },
    ui.MetricRowItemWidth(20),
    ui.MetricRowGap(1),
)
```

## Fiber-first 约束

- 只构造 VNode，不创建运行时实例状态。
- 指标值格式化可以通过显式 `Formatter(...)` 传入；常见空值兜底、计数归零和宽度压缩优先使用 `Value(...)` / `FallbackValue(...)` / `CompactValue(...)` / `Count(...)` 这些 item helper，根包对应 `ui.MetricRowValue(...)` 等入口。
- 默认每个指标使用同一 `ItemWidth(...)`；当单个指标需要承载较长的筛选、scope 或目标摘要时，可对该 item 调用 `WithWidth(...)`，或在根包使用 `ui.MetricRowItemWithWidth(...)`，避免把整行所有短计数一起拉宽。
- 不耦合具体业务领域或数据来源。

## 测试

```powershell
go test ./ui/components/metricrow
```
