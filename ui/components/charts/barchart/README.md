# BarChart

`barchart` 是第一批优先实现的图表组件之一。

目标定位：

- 分类比较
- Top N 展示
- 适合终端中的离散型数据对比

当前已实现：

- 单系列
- grouped bar
- stacked bar
- 水平柱图
- 轴线和标签
- value label
- 分类过多时的采样与标签折叠策略

最小示例：

```go
barchart.NewBuilder([]float64{12, 7, 15}).
    Title("Orders").
    Labels([]string{"A", "B", "C"}).
    Width(5).
    Height(4).
    ShowAxis(true).
    Build()
```

多系列横向示例：

```go
barchart.NewBuilder(nil).
    Title("Throughput").
    Labels([]string{"North America", "Europe"}).
    Series(
        barchart.Series{Name: "Ingress", Values: []float64{22, 20}},
        barchart.Series{Name: "Egress", Values: []float64{17, 18}},
    ).
    Horizontal().
    ShowLegend(true).
    ShowValue(true).
    Width(18).
    Height(5).
    Build()
```

说明：

- 垂直图在密集分类下会压缩标签到更短的槽位表示
- 水平图会优先保留柱体宽度，并对长标签做缩写折叠
