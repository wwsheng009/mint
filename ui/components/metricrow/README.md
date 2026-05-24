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
        {Title: "Runtime", Value: "healthy"},
        {Title: "Providers", Value: 12},
        {Title: "Failed", Value: 0},
    },
    ui.MetricRowItemWidth(20),
    ui.MetricRowGap(1),
)
```

## Fiber-first 约束

- 只构造 VNode，不创建运行时实例状态。
- 指标值格式化通过显式 `Formatter(...)` 传入。
- 不耦合具体业务领域或数据来源。

## 测试

```powershell
go test ./ui/components/metricrow
```
