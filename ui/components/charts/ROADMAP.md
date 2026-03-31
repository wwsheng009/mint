# Charts Roadmap

> 本文档记录 `ui/components/charts/` 组件族的分阶段实施路线。

## 当前状态

当前 charts 组件族已经完成第一阶段主干：

- [x] `charts/` 目录与命名空间结构建立完成
- [x] `ui/shortcuts_charts.go` 已建立
- [x] `internal/canvas / scale / layout / palette / downsample / axis / raster` 已落地
- [x] `sparkline / bulletchart / barchart / linechart` 已有最小可用实现
- [x] 第一批公开组件 README 最小示例已补齐
- [x] charts gallery demo 已建立
- [x] charts e2e 与 gallery e2e 已建立
- [x] `heatmap / scatterplot / candlestick` 已进入第二阶段公开组件

当前最重要的下一步不是继续无限横向扩张，而是把第二阶段新增组件继续往业务图表方向推进。

## 设计目标

### 1. 结构稳定

- 图表作为组件族统一收纳在 `ui/components/charts/`
- 图表共享底层能力集中在 `charts/internal/*`
- 根目录不承担父包聚合职责，避免父子包互相引用

### 2. 优先级清晰

- 先做终端里信息密度高、实现性价比高的图表
- 先打通“数据 -> 测量 -> 绘制 -> 测试 -> 示例 -> e2e”的闭环
- 复杂交互和更昂贵的图表延后

### 3. 复用当前组件架构

- 组件目录遵循 `builder.go + vnode.go + instance.go + *_test.go`
- 主体绘制采用 `PaintableInstance + Measure`
- 单测和 e2e 复用现有 `ui/components/*_test.go` 和 `ui/e2e/` 模式

## 目标组件分层

### Phase 0: 结构与规则

- [x] 建立 `charts/` 目录及文档
- [x] 建立 `internal/` 子目录
- [x] 建立图表子组件目录
- [x] 在 `ui` 顶层新增 `shortcuts_charts.go`

### Phase 1: 共享底层能力

- [x] `internal/canvas`
- [x] `internal/scale`
- [x] `internal/layout`
- [x] `internal/palette`
- [x] `internal/downsample`
- [x] `internal/axis`
- [x] `internal/raster`

### Phase 2: 第一批公开组件

- [x] `sparkline`
- [x] `bulletchart`
- [x] `barchart`
- [x] `linechart`

### Phase 3: 扩展组件

- [x] `heatmap`
- [x] `scatterplot`
- [x] `candlestick`

## 当前推荐顺序

1. 继续增强 `candlestick / scatterplot / heatmap` 的业务表达和样式回归
2. 视业务需求补更复杂的 domain / viewport / dense-data 能力
3. 再评估是否进入更复杂的金融或高密度图表

## 第一批组件当前状态

- `sparkline`
  - 已支持固定高度、自适应高度、最值高亮和 inline label
- `bulletchart`
  - 已支持 `qualitative ranges`、更细粒度 `target marker`、`Auto / Inline / Below` 值标签布局
- `barchart`
  - 已支持 grouped / stacked / horizontal / value label，并补齐密集标签折叠
- `linechart`
  - 已支持多系列、grid / legend / axis、point marker，以及连续性优先降采样

## Scatterplot 第一版范围

- 单系列和多系列散点渲染
- 基础 `grid / legend / axis`
- 系列级颜色和 glyph
- 自定义 `domain`
- 显式 `viewport`
- 可命名的 `x / y` 参考线
- 可命名的 `x / y` 参考区间带
- 密集点位 collision 计数标记
- 文本 snapshot 与 e2e 回归

## Candlestick 已完成范围

- OHLC 数据模型
- 单系列蜡烛渲染
- 上下影线与实体区分
- 基础 `axis / grid`
- 基础 `legend`
- 可选 `volume` 子图
- 密集时间轴标签折叠
- 更细粒度的 `body / wick / volume` 样式覆盖
- e2e 与文本 snapshot 回归

## Heatmap 已完成范围

- 静态矩阵渲染
- 行标签和列标签
- 强度 legend
- 低色彩终端可读的字符密度映射
- `full / compact` legend mode
- gallery / e2e / snapshot 回归

暂不进入：

- 大矩阵滚动与虚拟化
- tooltip / focus / selection

## 约束

### 包约束

- 不在 `ui/components/charts` 根目录实现可编译父包
- 不允许图表子组件 import `ui`
- 不允许 `charts/internal/*` 反向 import 具体图表组件
- 不允许图表子组件之间互相 import
- `charts/internal/*` 只允许依赖 `runtime/*`、`framework/theme`、`ui/components/internal/*`、`charts/internal/*` 和 `charts/model`
- 上述硬约束由 `ui/components/charts/import_rules_test.go` 做根层自动检查

### 视觉约束

- 图表轴线使用主题 `Border`
- 标签使用主题 `Muted`
- 系列色顺序固定为 `Primary -> Accent -> Secondary -> Success -> Warning -> Error`

### 交互约束

第一阶段默认展示优先：

- 不强制 tooltip
- 不强制缩放
- 不强制复杂十字光标

## 测试目标

每一批实现都应覆盖：

- internal 算法单测
- 组件 `Paint()` / `Measure()` 单测
- 至少一个示例 demo
- 必要时补 `ui/e2e/` 验证视觉和交互语义
- 关键视图补文本 snapshot 基线
