# Charts Optimization Backlog

> 本文档记录 `ui/components/charts/` 的活跃待办与实现细节收口项。

## 当前焦点

当前第一批图表已经落地，第二阶段里 `heatmap / scatterplot / candlestick` 也已进入可用状态。**活跃焦点转为扩展组件收口与高密度场景增强**。

## 结构与依赖

- [ ] 明确哪些共享类型放 `charts/model`，哪些继续留在各组件本地
- [x] 建立图表组件 import 约束检查，避免子组件反向依赖 `ui`
- [x] 建立具体图表组件之间不得互相 import 的自动检查
- [x] 为 `charts/internal/*` 增加最小白名单依赖检查
- [ ] 评估是否需要把部分通用 helper 下沉到 `ui/components/internal/`

## 渲染基础设施

- [x] 定义基础 render mode 与共享绘制能力
- [ ] 明确不同 render mode 的降级规则
- [ ] 定义 clipping 行为，避免绘制越出 plot area
- [x] 明确宽字符标签处理策略，统一使用 `paint.StringWidth()`
- [ ] 确定 grid line 和 axis label 的冲突处理方式

## 性能

- [x] 为按宽度降采样建立稳定基础算法
- [ ] 评估图表实例的绘制缓存键
- [ ] 为高频刷新场景预留脏标记与缓存复用策略

## 主题与样式

- [x] 把图表系列色映射写成可复用 helper
- [x] 明确 `heatmap` 的主题梯度策略
- [x] 评估并实现 truecolor / 256 色 / 16 色 / none 下的基础降级策略

## 组件级待办

### Sparkline

- [x] 支持最小/最大点高亮
- [x] 支持 inline label
- [x] 支持固定高度与自适应高度

### BulletChart

- [x] 支持 qualitative ranges
- [x] 支持更细粒度的目标线 marker
- [x] 支持值标签布局增强

### BarChart

- [x] 支持 grouped bar
- [x] 支持 stacked bar
- [x] 支持 horizontal bar
- [x] 支持 value label
- [x] 支持分类过多时的标签折叠优化

### LineChart

- [x] 支持多系列
- [x] 支持 point marker
- [x] 支持基础 grid / legend
- [x] 支持平滑降采样后的视觉连续性校验

### Heatmap

- [x] 定义第一版基础数据模型
- [x] 初始化最小骨架与单测
- [x] 增加更稳定的主题梯度策略
- [x] 明确低色彩终端的降级方案
- [x] 支持 `full / compact` legend mode
- [x] 支持 `global / viewport` 缩放模式
- [x] 支持 `auto` 缩放策略选择
- [x] 支持可选窗口摘要行
- [x] 增加 gallery 展示
- [x] 增加专门的 e2e 用例

### ScatterPlot

- [x] 定义第一版二维点数据模型
- [x] 支持单系列 / 多系列
- [x] 支持基础 grid / legend / axis
- [x] 支持系列级 glyph
- [x] 增加专门的 e2e 与 snapshot 用例
- [x] 支持可命名参考线
- [x] 支持参考区间带
- [x] 支持自定义 domain
- [x] 支持 viewport
- [x] 支持更密集点位下的 collision 策略

### Candlestick

- [x] 定义第一版 OHLC 数据模型
- [x] 支持单系列蜡烛渲染
- [x] 支持基础 axis / grid / legend
- [x] 支持可选 volume 子图
- [x] 支持密集时间轴下的标签折叠
- [x] 支持更细粒度的实体 / 影线 / volume 样式定制
- [x] 增加专门的 e2e 与 snapshot 用例

## 文档与示例

- [x] 建立 `charts_gallery_demo`
- [x] 为第一批公开组件补 README 中的最小示例
- [x] 为 `heatmap` 补最小示例与设计说明
- [x] 为 `scatterplot` 补最小示例与设计说明
- [x] 为 `candlestick` 补最小示例与设计说明
