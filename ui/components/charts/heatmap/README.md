# Heatmap

`heatmap` 是 charts 组件族里的二维矩阵热度图组件，适合终端里做高密度状态分布、容量热点和时段矩阵这类视图。

## 当前已实现

- 静态矩阵渲染
- 行标签和列标签
- `full / compact` 两种 legend 模式
- `truecolor / 256 / 16 / none` 颜色模式降级
- 行列窗口裁剪 `Viewport / RowWindow / ColWindow`
- `global / viewport / auto` 三种缩放模式
- 行标签裁剪 `MaxRowLabelWidth`
- `none / compact / detailed` 三种摘要模式
- gallery、组件单测、e2e 和文本 snapshot 回归

## 最小示例

```go
heatmap.NewBuilder([][]float64{
	{1, 2, 4, 6, 8},
	{2, 3, 5, 7, 9},
	{1, 1, 2, 3, 5},
	{4, 5, 6, 7, 8},
}).
	Title("Service Hotspots").
	RowLabels([]string{
		"North America",
		"South America",
		"Europe",
		"Asia Pacific",
	}).
	ColLabels([]string{"Mon", "Tue", "Wed", "Thu", "Fri"}).
	RowWindow(1, 3).
	ColWindow(1, 3).
	ViewportScale().
	MaxRowLabelWidth(4).
	CompactLegend().
	ColorMode(ui.HeatmapColorMode16).
	ShowAxis(true).
	ShowLegend(true).
	Build()
```

这段配置会得到一个适合终端宽度的小窗口：

- 只显示矩阵中的局部窗口，而不是把所有行列硬挤进一个视图
- 行标签会被裁剪成类似 `Sou~`、`Euro`
- legend 会使用更紧凑的 `L ░▒▓█ H`
- 热度层级会按可见窗口重新拉伸，避免局部区域全都挤成低对比度

## 推荐用法

适合：

- 一周七天或一天 24 小时的流量热点
- 服务和区域组成的错误率矩阵
- 任务队列与处理分片的负载分布

不适合：

- 需要精确数值比较的单点指标
- 类别很少但需要强标签解释的场景
- 需要交互 selection / tooltip 的复杂分析页

## 设计说明

`heatmap` 在终端里优先依赖两种表达：

- 字符密度：`░ / ▒ / ▓ / █`
- 颜色梯度：在终端能力允许时进一步区分热度层级

当颜色能力下降时，组件仍然依赖字符密度保持可读性，所以即使退到 `ColorModeNone` 也不会完全失效。

大矩阵场景目前推荐显式使用窗口能力，而不是一次性渲染整张表：

- `Viewport(...)`
- `RowWindow(start, count)`
- `ColWindow(start, count)`
- `ViewportScale()` 或 `ScaleMode(ScaleModeViewport)`
- `ShowSummary(true)`
- `MaxRowLabelWidth(width)`

默认缩放模式是 `global`，也就是整个矩阵共享一套最小值/最大值。这样跨窗口对比更稳定。

如果你当前展示的是一个局部窗口，而且更关心窗口内的冷热变化，可以切到 `viewport` 缩放模式，让颜色和字形只基于可见窗口计算。

如果你不想手动判断，可以使用 `AutoScale()`。它会在当前视图明显小于整张矩阵时自动切到 `viewport` 缩放；如果当前窗口只做了很轻微的裁剪，或者其实覆盖了完整矩阵，则仍然保持 `global`，避免过早放大局部对比度。

如果你需要在无交互场景里补一点数值语义，可以开启摘要模式：

- `ShowSummary(true)`：兼容旧调用，等价于 `DetailedSummary()`
- `CompactSummary()`：输出更短的 `min..max avg value`
- `DetailedSummary()`：输出更明确的 `range min..max avg value`

摘要始终基于当前可见窗口计算，所以它能和 `Viewport(...)`、`RowWindow(...)`、`ColWindow(...)` 自然配合，适合小窗口、静态报表和 e2e 可视回归。

## 示例

- 独立最小示例见 [examples/charts_heatmap_demo/README.md](/E:/projects/yao/wwsheng009/mint/examples/charts_heatmap_demo/README.md)
- 组合展示见 [examples/charts_gallery_demo/README.md](/E:/projects/yao/wwsheng009/mint/examples/charts_gallery_demo/README.md)
