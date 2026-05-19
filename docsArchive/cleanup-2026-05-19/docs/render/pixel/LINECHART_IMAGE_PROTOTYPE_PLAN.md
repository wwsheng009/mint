# LineChart Image Prototype 计划

## 1. 目标

本计划用于把 `docs/render/pixel` 目录下的总体方案，收敛成第一份真正可执行的 PoC。

本阶段只做一件事：

> 为 `linechart` 做一个最小可验证的 image rendering prototype

目标不是把整套 charts 一次性升级成图片图表，而是验证下列关键判断是否成立：

1. 当前引擎能否在不破坏文本路径的前提下，引入最小图像渲染能力
2. `linechart` 是否能从 image mode 获得明显优于字符折线的观感收益
3. 图像路径的性能、复杂度、测试成本是否在可接受范围内

## 2. 为什么选 `linechart`

PoC 阶段优先选择 `linechart`，原因有三点。

### 2.1 收益最直观

当前文本 `linechart` 的主要问题最适合通过像素渲染验证：

- 线条不够连续
- 小尺寸下走势容易断裂
- 点位和线段相互争抢字符格

如果 image mode 连 `linechart` 的收益都不明显，后续对 `scatterplot / heatmap / candlestick` 的说服力也会下降。

### 2.2 数据模型相对简单

`linechart` 当前公开模型已经稳定：

- 单系列
- 多系列
- labels
- axis label mode
- legend
- grid

相比 `heatmap` 或 `candlestick`，`linechart` 更适合作为“第一张图”验证渲染架构。

### 2.3 容易做文本 fallback 对照

当前已有：

- 组件单测
- e2e fixture
- render snapshot
- 独立 demo

这使得 image prototype 可以很容易和现有文本路径做 A/B 对比。

## 3. 第一阶段明确不做的事情

为了避免 PoC 失控，本阶段必须明确 non-goals。

### 3.1 不做全图表支持

本阶段不扩展到：

- `heatmap`
- `scatterplot`
- `candlestick`
- `sparkline`

### 3.2 不做所有终端协议兼容

本阶段只允许先支持一种图像协议。

推荐优先顺序：

1. `Kitty Graphics`
2. 之后再评估 `Sixel`

不要在 PoC 阶段同时兼容多协议。

### 3.3 不做复杂交互

本阶段不做：

- zoom
- pan
- brush selection
- hover tooltip
- crosshair

第一阶段只验证静态显示和受控更新。

### 3.4 不做全局 renderer 重写

本阶段不能一开始就替换现有文本 renderer。

必须坚持：

- 文本路径继续存在
- image path 通过实验性旁路或扩展接入
- 失败时可以完整回退

## 4. 推荐 PoC 形态

### 4.1 先做“实验性 image backend”，而不是直接改 `linechart` 默认行为

建议先加一个实验性 backend 选择：

```go
type ChartRenderBackend int

const (
    ChartRenderBackendAuto ChartRenderBackend = iota
    ChartRenderBackendText
    ChartRenderBackendImage
)
```

然后只让 `linechart` 识别它。

### 4.2 先做单终端、单图场景

PoC 第一版建议只验证：

- 本地图形终端
- 单张 `linechart`
- 固定尺寸
- 静态和低频更新

### 4.3 先做 demo / prototype 页面，而不是直接接入所有业务页面

推荐新增一个专门的 prototype 示例，而不是直接改现有 `charts_gallery_demo`：

- `examples/charts_linechart_image_prototype/`

这样失败成本最低，也便于独立 benchmark。

## 5. 建议的包级切入点

### 5.1 `runtime/platform`

建议新增一层图像能力抽象，优先位置建议之一：

- `runtime/platform/graphics.go`
- 或 `runtime/platform/image.go`

第一阶段最小接口建议：

```go
type GraphicsMode int

const (
    GraphicsModeNone GraphicsMode = iota
    GraphicsModeKitty
)

type GraphicsCapabilities struct {
    Mode            GraphicsMode
    CellPixelWidth  int
    CellPixelHeight int
}

type Graphics interface {
    Capabilities() GraphicsCapabilities
    DrawImage(req DrawImageRequest) error
    ClearImage(id string) error
}
```

注意：

- `Terminal` 和 `Screen` 现有接口先不要直接推翻
- 第一版可以通过可选能力接口挂接

### 5.2 `runtime/paint`

不建议第一步就改造 `paint.Buffer` 本身。

建议新增一个实验性 scene/frame 抽象，例如：

- `runtime/paint/scene.go`

第一版最小结构可以非常保守：

```go
type Scene struct {
    TextBuffer *Buffer
    Images     []ImageLayer
}
```

其中：

- 文本仍然走 `paint.Buffer`
- 图像层作为增量附加能力

### 5.3 `framework/app.go`

`App.render()` 是第一阶段最关键的接入点，但必须控制范围。

建议策略：

- 继续生成文本 `Buffer`
- 如果根树或某组件声明了 image layer，则额外收集图像绘制命令
- Phase 1 含图像层的实验帧优先绕过 `AsyncRenderer`
- 先在同步 `Render()` 结束时追加图像输出

原因：

- 当前 `AsyncRenderer` 只服务文本 `Buffer`
- 如果第一阶段就并入异步图像链，会让 PoC 范围快速失控

第一版容忍这不是最终优雅结构，只要满足：

- 不破坏文本路径
- 可验证 image prototype

### 5.4 `ui/components/charts/linechart`

建议只对 `linechart` 做最小新增：

- `RenderBackend(...)`
- `ImageMode()` / `TextMode()`

内部保留两条渲染路径：

- 现有文本路径
- 新增 image path

### 5.5 `examples`

新增一个 PoC 示例目录：

- `examples/charts_linechart_image_prototype/`

示例建议包含：

- 文本 `linechart`
- image `linechart`
- 同数据、同尺寸、并排展示

这样最容易做视觉和性能对照。

## 6. 第一阶段建议文件范围

为了让 PoC 可控，我建议第一阶段写入范围尽量收敛到这些区域：

### 必要写入

- `runtime/platform/*`
- `runtime/paint/*`
- `framework/app.go`
- `ui/components/charts/linechart/*`
- `examples/charts_linechart_image_prototype/*`

### 尽量不动

- 其他 chart 组件
- 全局 shortcuts 之外的大范围公开 API
- 现有 e2e 主链路
- `charts_gallery_demo`

## 7. 第一阶段建议功能切片

### Slice A：能力探测与占位接口

实现：

- `GraphicsMode`
- `GraphicsCapabilities`
- `Graphics` 可选接口

验收：

- 能在运行时拿到 graphics capability
- 不支持时明确返回 `GraphicsModeNone`

### Slice B：最小图像绘制命令

实现：

- `DrawImageRequest`
- 图像 placement
- 最小 image layer 容器

验收：

- `App.render()` 能在文本帧之外额外提交一张图像

### Slice C：`linechart` image renderer

实现：

- 数据转 pixel domain
- bitmap raster
- image backend 输出

第一版可接受的限制：

- 固定线宽
- 不做复杂 legend
- 不做多系列样式定制

验收：

- 同一张图在 image mode 下明显优于文本折线

### Slice D：prototype demo 与 benchmark

实现：

- 单独 demo
- benchmark / 计时日志
- diagnostics 输出

验收：

- 可做文本与 image 对照
- 可测首帧时间、更新时间、输出量

## 8. 第一阶段验证方式

### 8.1 功能验证

必须验证：

- 文本路径不被破坏
- image path 能显示
- resize 后恢复
- 切回 text mode 正常

### 8.2 视觉验证

必须直接比较：

- 相同数据
- 相同尺寸
- text mode 与 image mode

重点看：

- 连续性
- 小尺寸走势辨识度
- label 是否仍可用

### 8.3 性能验证

至少采集：

- 首帧时间
- 单次数据更新重绘时间
- 输出字节数
- cache 是否命中

### 8.4 稳定性验证

至少验证：

- 多次重绘是否残影
- alternate screen 切换后是否残留图像
- 不支持图像协议时是否稳定 fallback

## 9. 成功标准

本 PoC 不要求“全面上线”，但至少要满足这些条件。

### 9.1 架构成功标准

- 不破坏现有文本 renderer 主链
- 图像协议细节不泄漏到 chart 组件业务层
- `linechart` 能双路径运行

### 9.2 视觉成功标准

- 在相同尺寸下，image `linechart` 连续性明显优于文本路径
- 小图场景下走势可读性有实质提升

### 9.3 性能成功标准

- 静态或低频更新场景下，性能仍然可接受
- 不会因为 image path 让整页 render 明显失控

### 9.4 工程成功标准

- 有独立 demo
- 有最小 benchmark
- 有 diagnostics
- 失败可完整回退

## 10. 止损条件

如果出现下列情况，应立即止损，不要继续扩大 image mode 范围：

- 需要大规模重写 `Renderer` 才能勉强跑通第一张图
- 图像协议能力探测在目标终端上不可靠
- 单图 prototype 性能远差于文本路径，且没有明显视觉收益
- 文本路径回归被持续打坏
- 调试与测试成本显著超过收益

## 11. 推荐下一步

在这份 PoC 计划之后，建议继续补一份更细的工程任务清单，例如：

- `PHASE1_TASK_BREAKDOWN.md`

把第一阶段再拆成：

1. capability 层
2. image layer 最小接入
3. `linechart` image renderer
4. prototype demo
5. benchmark 与 diagnostics

一句话总结：

**第一阶段不要追求“把图片图表做完”，而要验证“最小 image linechart 路线是否值得继续投产”。**
