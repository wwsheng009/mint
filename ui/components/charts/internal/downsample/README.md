# Chart Downsample

计划中的降采样包。

目标：

- 根据可用宽度压缩数据点
- 保留趋势特征
- 避免超出终端列宽后直接挤压成噪声

典型场景：

- sparkline 的列聚合
- linechart 的 min/max bucket
- bar chart 分类过多时的降级策略

