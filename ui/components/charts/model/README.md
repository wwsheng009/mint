# Chart Model

这个目录用于放置图表组件之间确实需要共享、并且需要被外部 import 的模型类型。

当前阶段先保留为轻量目录，不主动把各组件本地类型上提到这里。`model/` 的门槛应当比“两个组件都能复用”更高，必须同时满足“多个图表共享”和“外部调用方值得直接 import”。

## 放到 `model/` 的条件

- 至少两个公开图表组件都需要这个类型
- 该类型本身是稳定的数据契约，而不是某个组件内部实现细节
- 外部调用方直接 import 这个类型会提升可读性或可组合性
- 该类型不依赖具体图表组件包，否则会引入反向耦合

## 明确不放到 `model/` 的内容

- 某个单一图表组件的局部数据结构
- 渲染期中间态，例如 layout frame、raster point、sampling bucket
- 仅供 `instance.go` 内部使用的 helper type
- 为了“看起来统一”而勉强抽出来、但还没有稳定复用面的类型

## 当前建议

以下类型继续留在各组件目录更合理：

- `scatterplot.Point`、`scatterplot.Series`、`scatterplot.ReferenceLine`、`scatterplot.ReferenceBand`
- `candlestick.Candle`
- `linechart.Series`
- `barchart.Series`
- `bulletchart.QualitativeRange`

原因很直接：

- 它们的字段语义仍然明显绑定具体图表
- 当前公开 API 通过组件包本身 import 已经足够清晰
- 过早上提到 `model/` 会让 `charts` 族看起来像一个大型 schema 包，反而增加维护成本

## 哪些类型将来可能进入 `model/`

只有在后续出现下面这些情况时，再考虑新增共享模型：

- 多个图表都开始依赖统一的 `AxisDomain` / `Viewport` 契约
- 多个图表都需要公开的 `LegendItem` 或 `SeriesStyleSpec`
- 业务层明确希望在多个图表之间共享同一种 threshold / annotation 数据结构

## 约束

- 只有在多个图表组件都需要共享且必须公开的数据结构时，才放到这里
- 不要把实现细节 helper 放到 `model/`
- 一旦类型只服务于单个图表组件，就留在该组件目录中
- `model/` 中的类型不应 import 任何具体图表组件
