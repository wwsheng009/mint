# Charts Gallery Demo

这个示例把当前已经落地的图表组件集中放到一个页面里，便于肉眼检查整体观感和组合布局。

包含的组件：

- `sparkline`
- `bulletchart`
- `barchart`
- `linechart`
- `heatmap`
- `candlestick`
- `scatterplot`

运行方式：

```bash
go run ./examples/charts_gallery_demo
```

这个 demo 重点展示：

- 紧凑 KPI 卡片中的 `sparkline`
- 目标对比场景里的 `bulletchart`
- `barchart` 的 horizontal / legend / value label 组合
- `linechart` 的多系列、legend 与紧凑趋势视图
- 小型矩阵场景中的 `heatmap`，包含 viewport scale 与 summary
- 迷你 `candlestick` 的 OHLC 趋势展示
- 迷你 `scatterplot` 的点位分布展示
