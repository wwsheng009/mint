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
