# Statistic

`statistic/` 提供单值统计展示组件，适合仪表盘、概览面板和摘要卡片。

支持：

- `title` / `value`
- `prefix` / `suffix`
- `precision`、千分位和小数分隔符
- `trend` 上下箭头
- `bordered` 样式与 `extra` 补充内容

## 示例

```go
ui.NewStatisticBuilder().
    Title("Revenue").
    Prefix("$").
    Value(12345.67).
    Precision(2).
    Up().
    Build()
```

## 指标行快捷构造

如果需要在运维概览、后台首页或详情页顶部展示多个紧凑指标，可以使用根包的 `ui.MetricRow(...)`。它基于现有 `Panel` 和 `Text` 组合，不引入新的运行时组件。

```go
summary := ui.MetricRow(
    []ui.MetricItem{
        {Title: "Runtime", Value: "healthy"},
        {Title: "Providers", Value: 12},
        {Title: "Failed", Value: 0},
    },
    ui.MetricRowItemWidth(20),
    ui.MetricRowGap(1),
)
```

常用选项：

- `MetricRowItemWidth(width)`
- `MetricRowGap(gap)`
- `MetricRowBorder(style, color)`
- `MetricRowValueStyle(style)`
- `MetricRowPanelStyle(style)`
- `MetricRowFormatter(func(interface{}) string)`
