# Group D `linechart` Image Renderer 技术规格

## 1. 文档目的

Group A 解决图形能力层，Group B/C 解决 Scene 与 `App.render()` 接入。  
Group D 才真正回答：

> `linechart` 的 image path 第一阶段到底应该渲染什么、保留什么、舍弃什么。

本文件的目标是把 `linechart` image prototype 收敛成一条明确可编码的实现线，而不是“先写起来再说”。

## 2. 当前 `linechart` 的现实基础

当前公开组件位于：

- `ui/components/charts/linechart/`

当前文本路径已经具备：

- 单系列与多系列
- `axis`
- `grid`
- `legend`
- `showPoints`
- `axis label mode`
- 单测、e2e 与文本 snapshot

这说明：

- `linechart` 已经不是一个空组件
- image path 不能推翻它现有的文本结构和测试价值

## 3. Group D 的核心目标

Group D 第一阶段只做一件事：

> 用 image path 显著改善 plot 区折线连续性，而不是把整张图都改成像素 UI。

换句话说，真正要“像素化”的核心区域只有：

- plot area

而不是整张组件。

## 4. 第一阶段必须固定的设计决策

### 4.1 决策一：只把 plot area 交给 image renderer

第一阶段建议保留这些部分在文本层：

- 标题
- legend
- axis label
- summary/footer

第一阶段交给 image layer 的只包括：

- 折线
- 点位
- plot 内网格
- plot 背景

这样做的收益非常大：

- layout 仍然是 cell-first
- 文本快照和文本回退仍然稳定
- image layer 的范围最小

### 4.2 决策二：第一阶段默认只要求单系列 image path 成立

多系列是有价值的，但不是第一阶段的硬门槛。

Phase 1 最低验收应是：

- 单系列 image linechart 明显优于文本 linechart

多系列可以按优先级排：

- P0：单系列
- P1：双系列
- P2：完整多系列

### 4.3 决策三：第一阶段不追求 image legend / image text

所有文字说明继续走文本层。

原因：

- 终端图像文字渲染会立刻引入字体、抗锯齿、度量和对齐问题
- 但 `linechart` 的核心视觉收益主要来自连续折线，不来自图片文字

### 4.4 决策四：小尺寸下 image path 也允许关闭点位

点位在文本图里已经会和线段争空间，在 image path 里也一样会影响可读性。

因此第一阶段建议：

- 小 plot 默认 `showPoints=false`
- 或由 image backend 自行做最小尺寸抑制

## 5. 推荐的公开控制项

建议沿用此前文档中的 backend 选择思路：

```go
type ChartRenderBackend int

const (
    ChartRenderBackendAuto ChartRenderBackend = iota
    ChartRenderBackendText
    ChartRenderBackendImage
)
```

第一阶段建议在 `linechart` 侧新增：

- `RenderBackend(...)`
- `ImageMode()`
- `TextMode()`

并保留：

- 默认仍然是 `Auto`

## 6. `linechart` 内部结构建议

### 6.1 不改现有 builder/vnode 主体结构

第一阶段建议只做增量字段：

- `renderBackend`

不要同时引入一整组新的 image-only props。

### 6.2 推荐新增单独文件

建议新增：

- `ui/components/charts/linechart/image_renderer.go`

而不是把所有 image 逻辑直接塞进 `instance.go`。

这样能保持：

- 文本路径仍在 `instance.go`
- image path 在独立文件里实验

### 6.3 推荐的内部职责划分

建议拆成三步：

1. 文本布局继续计算完整 chart frame
2. 从 frame 中切出 plot rect
3. 只对 plot rect 做 bitmap raster

## 7. Plot-only image path 的推荐流程

### 7.1 继续复用现有布局

第一阶段不要重写 `linechart` 的整体布局。

推荐流程：

1. 继续通过当前 `layout` 逻辑得到标题区、legend 区、plot 区、axis 区
2. plot 区仍先在 cell 坐标中确定
3. 再把 plot cell rect 映射成 pixel rect

### 7.2 只 rasterize plot 内容

第一阶段 image renderer 只需要知道：

- plot 宽高
- series 数据
- domain
- 是否显示 points
- 是否显示 plot grid

不需要负责：

- 标题文本
- legend 文本
- axis label 文本

### 7.3 结果回填为一个 `ImageLayer`

也就是说，`linechart` 最终应输出：

- 文本 `Buffer`
- 一个覆盖 plot 区域的 `ImageLayer`

## 8. 推荐的图像栅格策略

### 8.1 Phase 1 使用固定像素 surface

建议按照 plot cell rect 和 capability 得到：

- `plotPixelWidth = plotCellWidth * cellPixelWidth`
- `plotPixelHeight = plotCellHeight * cellPixelHeight`

这足够支撑第一阶段。

### 8.2 先用最小线宽策略

建议第一阶段线宽固定为：

- 1px 或 2px

不要在第一阶段引入：

- 自适应线宽
- 平滑曲线
- 样式主题级线宽

### 8.3 点位策略

建议第一阶段：

- 点位可选
- 小尺寸下默认关闭
- 点位只画简单圆点或方点

### 8.4 Plot grid 策略

建议：

- 如果 `showGrid=true`，网格画进 plot bitmap
- axis label 仍在文本层

这样视觉上更统一，也不会增加文本层复杂度。

## 9. Phase 1 支持范围建议

### 9.1 P0 必须支持

- 单系列
- 文本标题
- 文本 axis label
- plot bitmap
- text fallback

### 9.2 P1 可以支持

- 双系列
- plot 内 grid
- 点位开关

### 9.3 P2 留待后续

- 完整多系列
- image legend
- image tooltip
- image crosshair
- image viewport interaction

## 10. 回退策略

### 10.1 backend=Auto

建议逻辑：

- 图形能力可靠且 plot 尺寸足够时走 image
- 否则回退 text

### 10.2 backend=Image

如果用户显式要求 `Image` 但能力不可用：

- 记录 diagnostics
- 自动回退到 text
- 不要让组件直接失败

### 10.3 backend=Text

始终走当前文本路径，不触发图像层。

## 11. 与现有测试体系的关系

### 11.1 文本单测和 snapshot 不应被破坏

当前 `linechart` 的文本测试价值很高，不能因为 image path 而弱化。

### 11.2 Group D 需要新增的验证

建议新增：

- image renderer 单测
- prototype 级人工视觉对照
- diagnostics 输出检查

但第一阶段不要求：

- 完整图像 golden CI

## 12. 成功标准

### 12.1 视觉成功标准

- 相同尺寸下，image plot 的折线连续性明显优于文本路径
- 小尺寸下走势辨识度明显提升

### 12.2 工程成功标准

- 文本路径不被打坏
- image path 只新增有限写入范围
- 不需要改动所有 chart 公共 API

### 12.3 复杂度成功标准

- 第一阶段 `linechart` image path 只覆盖 plot area
- 不需要同时解 `legend + labels + interaction + async`

## 13. 风险与止损点

### 13.1 风险一：试图把整张 chart 都图像化

这会迅速引入：

- 图像文字布局
- legend 图像化
- 额外字体问题

建议第一阶段明确禁止。

### 13.2 风险二：过早追求多系列完整兼容

如果单系列还没证明收益，就不应该在多系列上继续扩复杂度。

### 13.3 风险三：为了 image path 重写整个 `linechart`

第一阶段 image path 应该是增量旁路，而不是替代文本实现。

## 14. 一句话结论

Group D 第一阶段最稳的方向是：

**保留 `linechart` 的文本布局和文字说明，只把 plot area 切出来做 image raster，从而用最小改动验证 image backend 的真实收益。**
