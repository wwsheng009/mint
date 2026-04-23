# Pixel 图表渲染架构方案

## 1. 背景

当前 `charts` 组件族已经具备一套可用的文本渲染能力：

- `sparkline`
- `bulletchart`
- `barchart`
- `linechart`
- `heatmap`
- `scatterplot`
- `candlestick`

这套能力在终端里已经足够完成：

- 趋势表达
- 对比分析
- 紧凑仪表盘
- 文本快照回归

但它仍然有一个天然上限：

- `linechart` 折线会受字符栅格限制而显得不连贯
- `scatterplot` 在高密度点位下只能近似表达
- `heatmap` 受字符密度和颜色层数限制
- `candlestick` 在小尺寸下 wick/body 细节不够

如果目标是进一步提升：

- 视觉连续性
- 数据表达密度
- 小尺寸可读性
- 接近真正图表库的表现力

仅靠继续打磨字符字形，不足以从根本上解决问题。此时需要引入“像素图表渲染”思路。

## 2. 路线定义

这里需要先区分两条不同路线。

### 2.1 伪像素路线

仍然是文本终端，但通过更高密度字符模拟像素：

- Braille
- Half Block / Full Block
- Unicode 斜线和点位
- 前景色 / 背景色组合

优点：

- 对现有引擎侵入小
- fallback 自然
- e2e 和 snapshot 机制基本不变

缺点：

- 仍然受 cell 网格限制
- 很难做到真正平滑折线
- 复杂图表表达力有限

### 2.2 真像素路线

先把图表 rasterize 成 bitmap，再通过终端图形协议显示：

- Kitty Graphics Protocol
- Sixel
- iTerm2 Inline Images

优点：

- 连续性和精度显著提升
- 可以做真正的抗锯齿折线、细粒度 heatmap、散点透明叠加
- 更接近图形界面图表

缺点：

- 需要引擎升级
- 终端兼容性复杂
- 需要强 fallback

本方案讨论的是第二条路线，即真像素图表渲染。

## 3. 当前引擎限制

### 3.1 `runtime/paint.Buffer` 是字符单元格缓冲

当前缓冲区以 `Width x Height` 的 cell 网格为核心，每个 cell 存：

- `Cluster`
- `Style`
- `Width`
- continuation 信息

这适合：

- 文本
- 边框
- 字符图

但不适合：

- bitmap
- 图像 tile
- 像素裁剪
- 区域级图像缓存

### 3.2 `runtime/platform.Terminal` 没有图形能力接口

现有终端抽象只提供：

- raw mode
- cursor
- string output
- 字符尺寸

缺少：

- 图像协议能力探测
- 像素尺寸
- 图像上传
- 图像删除/替换
- 图像裁剪和 anchor

### 3.3 当前主线仍然是 `App.render() + Renderer + AsyncRenderer`

对当前 charts 主线更关键的现实不是某个单独的 screen manager，而是：

- `framework/app.go`
- `runtime/paint/renderer.go`
- `runtime/paint/async_renderer.go`

它们共同决定了当前系统的真实假设：

- 一帧的核心载体仍然是文本 `Buffer`
- 同步和异步两条路径最终都只输出 ANSI 文本
- 当前没有 image-aware 的 frame model

仓库中的 `framework/screen/manager.go` 仍然有参考价值，因为它也体现了逐 cell diff 的历史设计；但 pixel 方案第一阶段的主要冲击点，实际上是 `App.render()`、`Renderer` 和 `AsyncRenderer`。

### 3.4 能力探测只有颜色，没有图形

当前主题输出层只区分：

- `TrueColor`
- `256`
- `16`
- `None`

缺少：

- `GraphicsNone`
- `GraphicsKitty`
- `GraphicsSixel`
- `GraphicsInlineImage`

## 4. 目标架构

目标不是替换现有文本渲染，而是在其上新增一层混合渲染能力。

### 4.1 新的能力模型

建议引入一套新的图形能力探测模型：

```go
type GraphicsMode int

const (
    GraphicsModeNone GraphicsMode = iota
    GraphicsModeKitty
    GraphicsModeSixel
    GraphicsModeInlineImage
)
```

并配套探测结果：

```go
type GraphicsCapabilities struct {
    Mode           GraphicsMode
    CellPixelWidth int
    CellPixelHeight int
    SupportsCrop   bool
    SupportsReuse  bool
    SupportsLayers bool
}
```

### 4.2 新的渲染场景模型

不要直接把图像塞进现有 `paint.Buffer`。

建议新增一层 scene/compositor 抽象，例如：

```go
type Scene struct {
    TextLayers  []TextLayer
    ImageLayers []ImageLayer
}
```

至少需要支持：

- 文本层
- 图像层
- overlay / z-order
- clip rect

### 4.3 新的绘制原语

当前组件输出文本绘制命令，后续建议扩展成：

- `DrawText`
- `DrawRect`
- `DrawImage`
- `DrawImageRegion`
- `DrawClip`

图表组件的 image mode 输出的就不再是字符线段，而是：

- 离屏 raster 结果
- 或 image handle + placement 信息

## 5. 必须新增的引擎能力

### 5.1 Cell / Pixel 双坐标体系

布局系统仍然应该基于 cell。

但在图像图表路径中，需要新增映射：

- `cell rect -> pixel rect`
- `pixel rect -> clipped image region`

否则无法：

- 准确控制图像大小
- 保持和文本布局一致
- 在 resize 后稳定重排

### 5.2 图像缓存和 dirty rect

图像图表必须具备缓存，否则性能不可接受。

至少需要：

- image hash
- tile 或 region dirty 标记
- 静态层 / 动态层分离
- 图像对象复用

建议：

- 坐标轴、背景网格、标题、图例独立缓存
- 数据层单独 rasterize
- resize 时整体失效
- 数据更新时优先局部失效

### 5.3 图像协议适配层

不同终端协议不应该污染 chart 组件。

建议新增 `runtime/platform/graphics` 或等价层，专门负责：

- 探测终端图形能力
- 生成协议输出
- 管理图像 ID
- 清理和替换旧图像

也就是说：

- `charts/*` 不直接输出 kitty/sixel escape sequence
- `platform` 或 renderer backend 负责协议细节

### 5.4 区域级渲染合成

屏幕层不应继续只做 cell diff。

要支持图像图表，渲染器至少要支持：

- 文本区域 diff
- 图像区域 diff
- 文本图像混排
- 图像 z-order

如果未来 tooltip、crosshair、brush selection 要叠在图像图表之上，这层会更重要。

补充一个第一阶段决策：

- Phase 1 不应把 image frame 直接并入 `AsyncRenderer`
- 更稳的方式是先让含图像的实验帧走同步提交闭环
- 等 capability、scene、prototype 和缓存策略稳定后，再设计异步图像管线

### 5.5 事件映射

如果后续图表支持交互：

- 鼠标点击
- hover-less summary
- 框选
- viewport 拖拽

就要能把：

- cell 坐标
- 映射到像素区域
- 再映射到 chart domain

建议提前把 hit test 模型设计好，而不是事后补。

## 6. 图表组件侧需要的变化

### 6.1 引入 Render Backend 概念

建议图表组件新增明确的 backend 选择，而不是只依赖终端能力自动推断：

```go
type ChartRenderBackend int

const (
    ChartRenderBackendAuto ChartRenderBackend = iota
    ChartRenderBackendText
    ChartRenderBackendImage
)
```

这样可以：

- 在测试里强制走 text
- 在支持图像的终端里强制走 image
- 避免“自动行为不透明”

### 6.2 先从少数图表切入

不要所有图表同时迁移。

建议顺序：

1. `linechart`
2. `heatmap`
3. `scatterplot`
4. `candlestick`
5. 其他图表继续按需推进

原因：

- 这些图表最能受益于像素渲染
- 也是当前文本路径最容易碰到上限的组件

### 6.3 保持双路径

每个图表短期内都应该同时保留：

- text mode
- image mode

而不是直接替换。

## 7. 回退策略

回退链路必须设计在第一版里。

建议顺序：

1. `Image`：终端支持图像协议，走 bitmap
2. `Unicode High Density`：终端不支持图像，走 Braille / Block
3. `ASCII`：能力最低时回退到 ASCII

这样：

- 文本快照测试仍然可用
- 不支持图像的终端不受影响
- 现有 charts 生态可以持续运行

## 8. 测试与验证策略

### 8.1 保留文本路径测试

现有：

- 组件单测
- e2e render assert
- snapshot 基线

仍然要保留，因为它们验证的是 fallback 路径。

### 8.2 新增图像路径测试

图像路径不能再只靠文本快照。

建议新增：

- 后端能力模拟测试
- image command 输出测试
- bitmap hash snapshot
- 终端协议级 golden test

### 8.3 诊断能力

建议工具层支持保存：

- 当前 scene dump
- 图像 tile hash
- 最终 bitmap
- 合成后的文本/图像边界信息

否则图像问题会很难调。

## 9. 推荐实施顺序

### Phase 1：能力建模

目标：

- 定义 `GraphicsMode`
- 定义图形能力探测结果
- 定义 cell/pixel 映射接口

不做：

- 具体图表 image mode

### Phase 2：实验性后端

目标：

- 做一个最小 image backend
- 只支持单一协议
- 只支持单个 chart

建议先选：

- `linechart`

### Phase 3：混合渲染整合

目标：

- 引入文本/图像混合渲染
- 支持 clip、缓存、dirty rect
- 保证 resize、rerender 可用

### Phase 4：图表扩展

逐步把：

- `heatmap`
- `scatterplot`
- `candlestick`

接进 image mode。

### Phase 5：交互与优化

最后再做：

- 局部放大
- crosshair
- brush selection
- partial rerasterization

## 10. 风险

### 10.1 终端兼容性

最大风险不是图表，而是终端差异。

包括：

- 支持协议不同
- 像素尺寸获取不一致
- 图像与文本叠加语义不同
- 某些协议在 tmux / ssh / CI 中表现不稳定

### 10.2 调试复杂度

纯文本 TUI 的优势是“可快照、可 grep、可肉眼 diff”。  
一旦进入图像层，问题定位会明显变难。

### 10.3 架构复杂度上升

如果混合渲染层设计不好，会让：

- renderer
- platform
- chart components
- testing

同时变复杂。

因此一定要：

- 先做能力建模
- 再做最小原型
- 最后才推到全局

## 11. 当前建议

对当前项目，最实际的结论是：

1. 不要继续把主要精力放在微调字符折线观感上
2. 先把 pixel / image rendering 方案在引擎层定型
3. 优先定义能力探测、混合绘制原语和缓存策略
4. 先用 `linechart` 做实验型 image backend

一句话总结：

**图片像素图表不是一个组件优化项，而是一个渲染引擎升级项目。**
