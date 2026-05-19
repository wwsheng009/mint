# Pixel 图表渲染对当前系统的冲击审查与可行性验证

## 1. 目的

`PIXEL_CHART_RENDERING_ARCHITECTURE.md` 给出了目标方向，但它还偏“目标架构”视角。

本文件补充两件更贴近当前仓库现实的问题：

1. 结合现有实现，真像素图表渲染会冲击当前系统的哪些位置
2. 应该如何分阶段验证方案是否真的可行，而不是停留在设计层

本文件默认讨论的是：

- 图表位图 rasterize
- 通过终端图形协议输出
- 与现有文本渲染系统并存

而不是单纯继续增强 Braille / Block / ASCII 文本图表。

## 2. 当前系统真实渲染路径

在评估冲击前，必须先明确当前系统到底是如何工作的。

### 2.1 应用主渲染入口在 `framework.App.render()`

当前主入口见：

- `framework/app.go`

关键事实：

- `App.render()` 从根节点获取 `component.Paintable`
- 使用 `a.renderer.GetBackBuffer()` 取得后缓冲
- 调用 `buf.Reset(a.terminalWidth, a.terminalHeight)`
- 通过 `paintable.Paint(ctx, buf)` 把当前 UI 画入文本缓冲
- 然后走 `a.renderer.Render()` 做 diff 和 ANSI 输出

这说明：

- 当前渲染结果的最终载体仍然是 `paint.Buffer`
- 整个 UI 渲染链默认假设“所有内容最终都能投射为字符单元格”

### 2.2 当前核心渲染器是文本 line diff 渲染器

当前 renderer 在：

- `runtime/paint/renderer.go`

关键事实：

- 仍然维护 `front/back` 双文本缓冲
- `Render()` 先对 `Buffer` 做 `Rehash()`
- 再做 line diff
- 最后把变化行格式化成 ANSI 输出

这说明：

- 现在的优化核心是“文本行 diff + run merging + dirty hint”
- 这套逻辑并没有图像对象、图层图像、像素区域的概念

### 2.3 当前还存在异步文本渲染管线

当前异步路径在：

- `runtime/paint/async_renderer.go`

关键事实：

- `AsyncRenderer` 内部仍然以 `Buffer`、`PartialFrameBuffer`、`RegionDiff` 为核心
- `SubmitFrame()` 提交的是文本 `Buffer`
- `renderPending()` 最终仍然调用 `renderer.Render()` 输出 ANSI 文本

这说明：

- 当前异步链并没有图像层概念
- 如果第一阶段直接把 image layer 塞进 `AsyncRenderer`，会把风险从“原型验证”放大成“渲染主链重构”

### 2.3 当前缓冲区抽象是 cell buffer

当前缓冲区在：

- `runtime/paint/buffer.go`

关键事实：

- `Buffer` 是 `Width x Height` 的 cell 网格
- `SetCell()`、`SetString()` 都围绕 grapheme cluster 和 cell width 工作
- continuation/wide char 是一等公民
- `GetRenderSnapshot()` 也是从文本 cell 重新拼装字符串

这说明：

- 当前缓冲区抽象对文本非常成熟
- 但它天然不承载 bitmap、tile、图像句柄和像素裁剪信息

### 2.4 当前平台层没有图像能力抽象

当前终端与屏幕抽象在：

- `runtime/platform/terminal.go`
- `runtime/platform/screen.go`

关键事实：

- `Terminal` 只有字符写入、光标、raw mode、字符尺寸
- `Screen` 只有 `Write/Flush/Clear/AltScreen`
- 没有图形协议能力探测
- 没有像素尺寸
- 没有 image upload / replace / delete

这说明：

- 当前平台层只服务于字符终端输出
- 不具备图像后端所需的最小接口

### 2.5 `framework/screen.Manager` 更像历史/并行路径，不是当前 charts 主线

仓库中仍然存在：

- `framework/screen/manager.go`

它同样体现了逐 cell diff 的屏幕模型，对评估系统历史包袱很有参考价值。  
但对当前 charts 主线来说，更关键的现实路径已经是：

- `framework/app.go`
- `runtime/paint/renderer.go`
- `runtime/paint/async_renderer.go`

因此 pixel 方案的第一批实现与审查，应优先围绕这些文件展开，而不是先改 `framework/screen.Manager`。

### 2.6 当前输入与命中测试依赖文本渲染结果

相关入口：

- `framework/event/pump.go`
- `framework/app.go` 中 `a.pump.SetHitMap(a.hitMap)`

关键事实：

- `HitMap` 在 render 后更新
- 鼠标命中测试依赖 render 后的布局/命中信息
- 当前命中测试天然基于“文本布局后的区域”

这说明：

- 如果图表切到 image mode，事件映射不能再只看文本 cell
- 需要补“图像区域到 chart domain”的映射层

## 3. 对当前系统的冲击面

### 3.1 对 `framework.App` 的冲击

冲击级别：高

原因：

- 现在 `App.render()` 默认假设所有内容先画到 `paint.Buffer`
- 图像图表一旦引入，`render()` 必须能处理“文本 + 图像”混合帧

需要考虑的变化：

- `App.render()` 是否继续只处理 `paint.Buffer`
- 是否新增 `Scene` / `Frame` / `CompositeFrame`
- `AI service renderSeq` 是否仍以文本帧为唯一依据
- `debug recorder` 如何记录图像帧
- `async renderer` 是否支持图像 payload

直接风险：

- 如果在 `App.render()` 里硬塞图像协议输出，很容易把当前文本管线打乱
- 异步渲染与图像上传如果不同步，会导致闪烁或覆盖顺序错误

### 3.2 对 `runtime/paint.Renderer` 的冲击

冲击级别：极高

原因：

- 当前 renderer 设计前提是“最终输出是 ANSI 文本”
- 它的优化都建立在 line diff、run merging、光标最小移动之上

图像图表要求新增：

- 图像图层
- 图像对象缓存
- 区域级 dirty 管理
- 文本层与图像层合成顺序
- 图像对象的生命周期管理

直接风险：

- 如果继续沿用 line diff renderer 作为唯一输出器，图像路径会非常别扭
- 图片与文本如果不分层，后续 tooltip/selection/hit testing 都会很难维护

### 3.3 对 `runtime/paint.AsyncRenderer` 的冲击

冲击级别：极高

原因：

- 当前 `AsyncRenderer` 只认识文本 `Buffer`
- 其 `stage`, `RegionDiff`, `pendingRects`, `copyBufferRect` 全是文本语义

直接风险：

- 如果第一阶段试图把 image layer 直接并入 `AsyncRenderer`，会迅速扩大写入范围
- 文本节流、区域更新和图像生命周期管理会纠缠在一起

建议：

- Phase 1 图像帧先走同步提交
- 等 image prototype、缓存策略和协议后端稳定后，再评估异步图像管线

### 3.4 对 `runtime/paint.Buffer` 的冲击

冲击级别：中高

这里要注意：

- 不一定要把 `paint.Buffer` 直接改成像素缓冲
- 更合理的是在它之上新增 scene/layer 抽象

但即便不直接改 `Buffer`，也会有冲击：

- `GetRenderSnapshot()` 只能表达文本
- 现有 e2e 体系默认把渲染结果视为字符串
- dirty rect 当前主要服务于文本渲染提示

建议：

- 尽量保留 `Buffer` 作为文本层
- 不要把 bitmap 强行塞进 `Buffer.Cells`
- 新增图像层抽象，而不是污染现有文本缓冲模型

### 3.5 对 `runtime/platform` 的冲击

冲击级别：极高

当前 platform 层是图片方案的最大缺口之一。

需要新增的最小能力：

- 图形协议能力探测
- 终端 cell 对应像素尺寸
- 图像写入接口
- 图像对象替换 / 删除
- 可能还要支持 clip / placement / anchor

直接风险：

- 如果平台层不抽象好，协议细节会外溢到 charts 组件
- 不同终端协议差异会污染 renderer 和 component 代码

### 3.6 对输入与交互的冲击

冲击级别：中高

如果 image mode 下图表是不可交互的，问题相对可控。

但如果未来支持：

- click
- viewport drag
- crosshair
- brushing

就必须解决：

- cell 坐标 -> pixel 区域
- pixel 区域 -> chart logical domain

否则当前 hit test 模型只适合文本节点，不适合图像图表内部交互。

### 3.7 对测试体系的冲击

冲击级别：高

当前 charts 测试体系非常依赖：

- 文本 e2e
- render snapshot
- plain-text diagnostics

这是一套非常有效的文本图表回归机制。

但 image mode 一旦引入：

- 文本 snapshot 无法完整表达图像结果
- `GetRenderSnapshot()` 只能覆盖 fallback 文本路径
- 当前 diagnostics 目录不包含图像资产

这意味着测试体系必须新增第二条验证路径。

## 4. 哪些部分可以保持不动

不是所有层都要一起推翻。

以下部分应该尽量保留：

### 4.1 现有文本图表路径

它仍然应该存在，原因：

- fallback 必须保留
- CI / e2e / 快照当前高度依赖文本渲染
- 并不是所有终端都支持图像协议

### 4.2 现有 charts 组件公开 API

除非必要，不建议一开始就把 `linechart`、`heatmap`、`scatterplot` 对外 API 全部推翻。

建议新增的是：

- backend selector
- capability-aware render policy

而不是破坏现有 builder。

### 4.3 现有测量与布局主模型

布局仍然应该以 cell 为主。

原因：

- 整个 UI 系统当前就是 cell-based layout
- 改成 pixel-first 会影响整个框架，不只是图表

更合理的做法是：

- layout 产出 cell rect
- image backend 再把它映射成 pixel rect

## 5. 设计冲击的核心判断

综合来看，图片像素图表渲染对当前系统的冲击主要集中在五层：

1. `platform`
2. `renderer`
3. `async renderer`
4. `mixed frame / scene model`
5. `testing`

而不是首先冲击各个 chart 组件。

这意味着：

- 不能把它当作一个普通组件增强任务
- 必须当作渲染架构扩展任务来设计

## 6. 可行性验证的正确方式

这类方案不能靠“感觉上可行”推进，必须做分阶段 PoC。

建议采用三段验证。

### 6.1 验证一：能力验证

目标：

- 当前运行环境能不能拿到图形协议能力
- 能不能拿到 cell 对应的像素尺寸

验证内容：

- 终端能力探测 API
- capability dump
- fallback 选择逻辑

成功标准：

- 能稳定区分 `no graphics` 与 `graphics available`
- 能拿到足够可靠的尺寸信息

失败即止损条件：

- 主流目标终端不能稳定识别
- 像素尺寸信息获取不可靠

### 6.2 验证二：单图表原型验证

目标：

- 只选一个图表做 image mode 原型

建议只选：

- `linechart`

原因：

- 它最能体现像素渲染的收益
- 也最容易看出文本路径的上限

验证内容：

- 把 `linechart` rasterize 成 bitmap
- 在支持协议的终端里显示
- 与文本 UI 共存
- resize 后重绘
- fallback 到 text mode 正常

成功标准：

- 能在一个真实终端中稳定显示
- 不破坏现有页面布局
- 不出现明显闪烁或残影

失败即止损条件：

- 图像与文本叠加顺序不稳定
- resize 后无法稳定恢复
- 图像更新成本过高

### 6.3 验证三：渲染链整合验证

目标：

- 验证 image mode 能不能进入现有 render loop，而不是只做单独 demo

验证内容：

- `App.render()` 集成
- async renderer 路径
- dirty rect 与图像缓存
- diagnostics 输出
- e2e 可观测性

成功标准：

- 不影响文本路径
- image/text 混合帧稳定
- 可做最小自动化回归
- 第一阶段允许 image frame 先绕过 `AsyncRenderer`

失败即止损条件：

- 主线 renderer 复杂度失控
- 为了图像原型不得不大改 `AsyncRenderer`
- 文本回归大量退化
- 调试成本不可接受

## 7. 建议的验证指标

方案可行性不能只靠“能显示出来”，应至少看这些指标。

### 7.1 功能指标

- 是否支持文本 + 图像共存
- 是否支持 resize
- 是否支持 fallback
- 是否支持多图表共存

### 7.2 性能指标

- 首帧时间
- 数据更新后的重绘时间
- 图像缓存命中率
- 输出字节数
- 高频刷新时 CPU 占用

### 7.3 稳定性指标

- 多次切换 alternate screen 是否残留图像
- 终端 resize 后是否错位
- 图像对象是否泄漏
- 非支持终端是否稳定回退

### 7.4 可测试性指标

- 是否能生成图像 golden
- 是否能保存诊断文件
- 是否能在 CI 环境对 fallback 路径做可靠验证

## 8. 最小可行原型建议

建议的最小原型不是直接进 charts 主线，而是做一个独立 PoC：

### Phase A

- 新增 `GraphicsCapabilities`
- 新增图像协议 backend 接口
- 不接任何 chart 组件

### Phase B

- 新增 `LineChartImagePrototype`
- 固定宽高
- 固定终端协议
- 只验证渲染和刷新

### Phase C

- 把 prototype 接进 `charts/linechart`
- 通过 `RenderBackendAuto/Text/Image` 控制

### Phase D

- 才考虑扩展到 `heatmap / scatterplot`

## 9. 当前最重要的非功能要求

在任何实现前，必须先满足这几个要求：

1. **不破坏现有文本图表路径**
2. **不让图像协议细节泄漏到 charts 组件层**
3. **不强迫 layout 改成 pixel-first**
4. **必须保留 fallback**
5. **必须有 PoC 阶段性止损点**

## 10. 最终结论

### 10.1 这件事可行吗

可行，但不是低成本增强。

它是一个真实的渲染架构升级任务，主要冲击：

- `platform`
- `renderer`
- `frame/scene model`
- `testing`

### 10.2 现在适合立刻全面开工吗

不适合直接全面开工。

更合理的顺序是：

1. 先把设计审查清楚
2. 先做最小能力验证
3. 先做单图表 image prototype
4. 验证通过后再考虑主线集成

### 10.3 当前最推荐的动作

如果要继续推进，下一步不应该直接改 charts 组件，而应该先产出一份更偏工程实施的文档，明确：

- 要新增哪些接口
- 哪些包需要 first-class 支持 graphics
- PoC 的最小 write scope 是什么
- 验证指标和止损条件是什么

一句话总结：

**这条路线值得研究，但必须先用最小原型验证，不适合直接在现有主线 renderer 上硬改。**
