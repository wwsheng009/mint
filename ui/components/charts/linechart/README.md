# LineChart

`linechart` 是第一批优先实现的图表组件之一。

状态说明：

- `linechart` 当前只承诺稳定的文本渲染路径
- 为兼容保留的 `ImagePlotBackend()` API 目前会在运行时回退到文本 backend
- 终端像素图像能力仍然保留给专用图片显示控件，不再作为 charts 的接入面

目标定位：

- 时间序列
- 趋势分析
- 多系列折线对比

当前已实现：

- 单系列或少量多系列
- Braille 优先渲染
- 基础 grid / axis / legend
- point marker 可选
- 窄宽度下优先保留 turning point 的连续性采样
- 公开的轴标签策略 `auto / dense / sparse`

最小示例：

```go
linechart.NewBuilder([]float64{1, 3, 2, 5, 4}).
    Title("Latency Trend").
    Width(9).
    Height(3).
    ShowAxis(true).
    ShowPoints(true).
    Build()
```

多系列示例：

```go
linechart.NewBuilder(nil).
    Title("Traffic").
    Series(
        linechart.Series{Name: "API", Data: []float64{1, 3, 2, 5, 4}},
        linechart.Series{Name: "Worker", Data: []float64{2, 2, 4, 3, 5}},
    ).
    Width(9).
    Height(3).
    ShowLegend(true).
    ShowGrid(true).
    ShowAxis(true).
    ShowPoints(true).
    Build()
```

轴标签策略示例：

```go
linechart.NewBuilder([]float64{1, 9, 2, 8, 3, 7}).
    Title("Latency").
    Labels([]string{"03/24", "03/25", "03/26", "03/27", "03/28", "03/29"}).
    DenseAxisLabels().
    Width(11).
    Height(4).
    ShowGrid(true).
    ShowAxis(true).
    Build()
```

说明：

- `AutoAxisLabels()` 保持当前默认启发式，尽量兼顾信息量和不拥挤
- `DenseAxisLabels()` 尽量保留所有可见标签
- `SparseAxisLabels()` 主动降低标签密度，适合窄宽度或更强调走势的场景
- 如果旧代码仍然调用 `ImagePlotBackend()`，当前也只会得到文本 `linechart`

实现建议：

- 等 `internal/canvas` 和 `internal/raster` 基本稳定后再开始
- 先保证静态绘制正确，再逐步考虑交互
- 窄宽度场景优先保留峰谷和 turning point，而不是盲目最近邻抽样
