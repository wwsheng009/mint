# Charts LineChart Demo

这个示例包含两部分，适合直接观察 TUI 里的折线图效果：

- 一个更宽、更高的 `Readability Baseline`，用于展示终端里可读的折线趋势
- 一排紧凑的轴标签模式对比

同时会展示：

- `AutoAxisLabels()`
- `DenseAxisLabels()`
- `SparseAxisLabels()`
- 同一组数据在三种模式下的标签密度差异

运行：

```bash
go test ./examples/charts_linechart_demo
go run ./examples/charts_linechart_demo
```
