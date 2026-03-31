# Sparkline

`sparkline` 是第一批优先实现的图表组件之一。

目标定位：

- 行内趋势图
- KPI 卡片中的迷你趋势图
- `table` / `list` 的紧凑型时间序列展示

当前已实现：

- 单系列
- 固定高度与基于约束的自适应高度
- `auto` / `braille` / `block` / `ascii` 渲染模式
- 可选最小值/最大值高亮
- 可选 inline label
- `Auto` 单行宽图默认优先 `braille`，窄图降级到 `ascii`
- 多行模式统一以 `block` 为默认降级，显式 `ascii` 会保留 ASCII 字形

最小示例：

```go
sparkline.NewBuilder([]float64{7, 8, 8, 9, 11, 12, 13, 12, 14, 16}).
    Title("Requests").
    Width(12).
    Braille().
    HighlightMinMax(true).
    Build()
```

多行紧凑示例：

```go
sparkline.NewBuilder([]float64{42, 45, 44, 51, 47, 58, 55, 61}).
    Title("Latency").
    Width(8).
    Height(3).
    InlineLabel("live").
    HighlightMinMax(true).
    Build()
```

实现建议：

- 组件主体采用 `PaintableInstance + Measure`
- 按 plot 宽度对原始数据做列级降采样
- 默认保持单行兼容路径，只有显式 `Height(...)` 或 `AutoHeight()` 才进入多行渲染
