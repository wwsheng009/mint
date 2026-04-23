# Charts Internal Packages

`ui/components/charts/internal/` 用于存放图表组件族内部共享的实现。

这里的代码服务于：

- `sparkline`
- `bulletchart`
- `barchart`
- `linechart`
- `scatterplot`
- `candlestick`
- `heatmap`

## 规则

- 这里的包不应 import 任意具体图表组件
- 这里的包只依赖 `runtime/*`、`framework/theme`、必要时依赖 `ui/components/internal/*`
- 这里的包当前还允许依赖 `charts/internal/*` 彼此之间的共享实现，以及 `charts/model`
- 如果某个能力未来被非图表组件复用，再评估是否下沉到更高层的 `ui/components/internal/`

更细的约束见：

- [IMPORT_RULES.md](./IMPORT_RULES.md)
- [RENDERING_RULES.md](./RENDERING_RULES.md)

其中 `IMPORT_RULES.md` 里列出的硬约束已经落成根层自动检查，入口在 [`../import_rules_test.go`](../import_rules_test.go)。当前这组自动检查也会拦截：

- 具体图表组件之间的互相 import
- `charts/internal/*` 超出白名单层级的模块内依赖

## 子目录

- `canvas/`：字符画布与 Braille / Block 栅格映射
- `scale/`：比例尺
- `layout/`：plot 区布局与测量
- `palette/`：图表调色和系列色映射
- `downsample/`：降采样
- `axis/`：坐标轴与刻度
- `raster/`：点、线、柱的字符栅格化

## 当前职责边界

- 进入 `internal/` 的前提是“多个图表已经稳定复用”
- 进入 `internal/` 不等于自动变成对外公开 API
- `internal/` 里的包应该提供小而稳定的实现能力，而不是把某个图表完整抽象成第二套框架

## 当前空缺的处理方式

目前 backlog 里仍然存在一些尚未彻底统一的规则，例如：

- render mode 降级细则
- clipping 统一行为
- grid 与 label 冲突处理

这些规则在编码时应优先向共享 helper 靠拢；在真正出现两个以上组件重复实现之前，不要为了“理论完美”提前过度抽象。
