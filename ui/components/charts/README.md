# Mint Charts Components

`ui/components/charts/` 是图表组件族的命名空间目录。

当前它不再只是占位结构，已经具备一套可运行的第一阶段图表能力，并正在进入第二阶段扩展。

顶层 `ui` 快捷入口提供 `ASCIISparkline` 与 `ASCIIBarChart`，适合只需要简单趋势或分类对比、且希望输出完全兼容纯文本终端的运维面板。

## 当前已实现

公开组件：

- `sparkline`
- `bulletchart`
- `barchart`
- `linechart`
- `heatmap`
- `scatterplot`
- `candlestick`

共享内部能力：

- `internal/canvas`
- `internal/scale`
- `internal/layout`
- `internal/palette`
- `internal/downsample`
- `internal/axis`
- `internal/raster`

顶层入口：

- `ui/shortcuts_charts.go` 已接入 `sparkline / bulletchart / barchart / linechart / heatmap / scatterplot / candlestick`

示例与验证：

- `examples/charts_gallery_demo`
- `ui/e2e/charts_e2e_test.go`
- `ui/e2e/charts_gallery_e2e_test.go`

当前 gallery 已经不只是静态拼盘，它还会用 mini `heatmap` 直接展示 `viewport scale + summary` 这类高密度窗口能力。

第一批组件增强已收口：

- `sparkline` 已补齐固定高度、自适应高度、最值高亮和 inline label
- `bulletchart` 已补齐 `qualitative ranges`、可配置 `target marker` 和值标签布局模式
- `barchart` 已补齐分类过多时的标签折叠优化，并新增纯 ASCII glyph 模式用于高兼容监控面板
- `linechart` 已补齐窄宽度下优先保留 turning point 的连续性采样

当前约束说明：

- `linechart` 的 chart 级像素 backend 已暂停接入，组件层统一走文本渲染
- 终端图形协议和基础 image-layer 机制仍然保留，但它们现在只面向专用图片控件，不再作为 charts 的默认能力

## 当前结论

第一批图表已经跑通：

- `sparkline`
- `bulletchart`
- `barchart`
- `linechart`
- `heatmap`
- `scatterplot`
- `candlestick`

第二阶段扩展组件已经进入可用状态：

- `heatmap`
- `scatterplot`
- `candlestick`

当前更合理的主线不是继续快速堆新图表，而是把扩展组件的业务表达、样式回归和示例说明继续收口：

- `scatterplot` 已经验证了二维连续坐标的复用链路
- `candlestick` 已经验证了 OHLC 场景在终端字符栅格里的最小闭环
- 下一步更值得投入的是扩展组件增强，而不是立刻进入更昂贵的网络型图表

## 第一批组件示例入口

- [sparkline/README.md](./sparkline/README.md)
- [bulletchart/README.md](./bulletchart/README.md)
- [barchart/README.md](./barchart/README.md)
- [linechart/README.md](./linechart/README.md)

## 扩展组件示例入口

- [heatmap/README.md](./heatmap/README.md)
- [scatterplot/README.md](./scatterplot/README.md)
- [candlestick/README.md](./candlestick/README.md)
- [../../../examples/charts_gallery_demo](../../../examples/charts_gallery_demo)

独立 heatmap、scatterplot、candlestick 小示例已归档到 `../../../docsArchive/cleanup-2026-05-19/_examples/`，当前推荐使用综合图表示例。

## 目录原则

这个目录当前**不作为一个可编译的 Go 父包**使用。

原因：

- Go 的循环引用是按包而不是按目录检查的
- 如果把 `ui/components/charts` 做成 `package charts`，再让它 import 子图表包，后续极易形成父子互相依赖
- 图表族内部的共享能力应下沉到 `charts/internal/*`

推荐结构：

```text
ui/components/charts/
├── README.md
├── ROADMAP.md
├── OPTIMIZATION_BACKLOG.md
├── internal/
├── model/
├── sparkline/
├── bulletchart/
├── barchart/
├── linechart/
├── heatmap/
├── scatterplot/
└── candlestick/
```

## 依赖约束

- `ui` 顶层快捷入口可以 import 图表子组件
- 图表子组件不能 import `github.com/wwsheng009/mint/ui`
- `charts/internal/*` 不能 import 任意具体图表组件
- 图表子组件之间不得互相 import
- 共享实现放到 `charts/internal/*`
- 只有在多个图表子组件确实需要共享且必须公开的数据结构时，才放到 `charts/model`

当前这三条结构规则已经有根层自动检查兜底：

- `ui/components/charts/import_rules_test.go` 会检查 `charts` 根目录没有可编译父包代码
- 同一个测试会检查具体图表组件没有反向 import `github.com/wwsheng009/mint/ui`
- 同一个测试会检查 `charts/internal/*` 没有反向 import 任意具体图表组件
- 同一个测试会检查具体图表组件之间没有互相 import
- 同一个测试会对 `charts/internal/*` 执行最小白名单依赖检查

## 文档入口

- 总体规划见 [ROADMAP.md](./ROADMAP.md)
- 活跃待办见 [OPTIMIZATION_BACKLOG.md](./OPTIMIZATION_BACKLOG.md)
- 内部基础设施说明见 [internal/README.md](./internal/README.md)
- 示例入口见 [../../../examples/charts_gallery_demo/README.md](../../../examples/charts_gallery_demo/README.md)
