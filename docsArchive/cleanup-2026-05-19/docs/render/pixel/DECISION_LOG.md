# Pixel 渲染 Phase 1 决策日志

## 1. 文档目的

当方案文档逐渐变多时，最容易出现的问题不是“没有方案”，而是：

> 同一个关键问题在不同实现阶段被反复重新讨论。

本文件的作用就是把 Phase 1 已经锁定的关键决策单独提出来，形成一份可供评审和实现共同遵守的决策日志。

## 2. 决策范围

本决策日志只覆盖：

- `linechart image prototype`
- Phase 1

不覆盖：

- Phase 2 以后
- 其他图表的 image mode 扩展
- 多协议并行支持

## 3. 已锁定决策

## D1：第一阶段只验证 `linechart`

结论：

- Phase 1 只选 `linechart`

原因：

- 收益最直观
- 文本基线最完整
- 失败成本最低

影响：

- `heatmap / scatterplot / candlestick` 暂不进入 image prototype 主线

## D2：第一阶段只要求 `GraphicsModeNone + GraphicsModeKitty`

结论：

- Phase 1 不要求 `Sixel` 或其他协议实现

原因：

- 同时支持多协议会显著放大复杂度
- 第一个阶段更需要稳定验证，而不是兼容大全

影响：

- `GraphicsModeSixel / GraphicsModeInlineImage` 可以存在于枚举中，但不要求落地

## D3：`paint.Buffer` 继续只承载文本层

结论：

- 不把 bitmap 或图像对象塞进 `Buffer.Cells`

原因：

- 当前 `Buffer` 的 cell 语义很稳定
- 污染 `Buffer` 会放大对文本渲染主链的冲击

影响：

- 图像层必须通过新的 `SceneFrame` 增量表达

## D4：第一阶段引入最小 `SceneFrame`

结论：

- 新增 `SceneFrame / ImageLayer / SceneDiagnostics`

原因：

- 需要在文本帧之外承载图像层
- 但又不值得为 Phase 1 直接上重型 compositor

影响：

- `paint` 层新增最小 Scene 包装层

## D5：第一阶段不修改 `component.Paintable` 主契约

结论：

- 现有 `Paint(ctx, buf)` 保持不变

原因：

- 改动范围过大
- 与 Phase 1 “最小验证”目标不匹配

影响：

- 通过可选扩展接口提供实验性 Scene 能力

## D6：含图像层的实验帧先绕过 `AsyncRenderer`

结论：

- Phase 1 的 image frame 一律走同步提交闭环

原因：

- 当前 `AsyncRenderer` 只适配文本 `Buffer`
- 过早接入图像层会迅速扩大改造面

影响：

- 纯文本页面仍走现有同步/异步路径
- 只有实验性 image frame 走同步旁路

## D7：`linechart` 第一阶段只像素化 plot area

结论：

- image path 只负责 plot 区

原因：

- 真正需要像素连续性的主要是折线本体
- 标题、legend、axis label 没必要在第一阶段图像化

影响：

- 文本层继续保留标题、legend、axis label、summary/footer

## D8：第一阶段默认优先单系列

结论：

- 单系列 image `linechart` 是 P0

原因：

- 先证明单图收益最重要
- 多系列会增加样式、legend、点位冲突等复杂度

影响：

- 多系列不是 Phase 1 的硬门槛

## D9：图像层第一阶段视为非交互视觉层

结论：

- 图像层暂不承担复杂命中测试与交互

原因：

- hit testing 一旦纳入 image region，会迅速把 scope 扩到 viewport/crosshair/selection

影响：

- 当前 `HitMap` 继续按文本路径更新

## D10：图像失败必须稳定回退到文本路径

结论：

- capability 不可靠、scene 构建失败、presenter 提交失败，都必须允许 text fallback

原因：

- Phase 1 不能因为 image prototype 破坏既有页面输出

影响：

- `Image` 永远不是必经路径

## D11：实验性产物不进入 `ui/e2e/testdata`

结论：

- capability dump、benchmark JSON、bitmap、presenter.log 不进入文本 snapshot 目录

原因：

- `ui/e2e/testdata` 当前语义是稳定文本回归资产

影响：

- prototype 产物应进入 `artifacts/pixel/...`

## D12：Phase 1 不以 image golden CI 为前置条件

结论：

- 第一阶段允许先依赖本地 prototype + diagnostics + benchmark 做判断

原因：

- image golden CI 会额外拉高环境依赖与测试复杂度

影响：

- 文本 CI 继续稳定运行
- image prototype 先走实验闭环

## 4. 仍保留开放性的决策

这些项目前仍是开放项，不应被误读为已定结论：

1. `Sixel` 是否进入 Phase 2
2. image-aware `AsyncRenderer` 是否值得做
3. image region 是否要进入 `HitMap`
4. `heatmap / scatterplot / candlestick` 的 image mode 启动顺序
5. image golden CI 何时引入

## 5. 决策变更规则

如果未来要推翻本文件中的某个决策，建议满足两个条件：

1. 在对应实现阶段已经拿到明确证据
2. 修改相应规格文档，而不是只在代码中悄悄偏离

也就是说：

- 先更新决策日志
- 再改代码

## 6. 一句话结论

这份决策日志的作用是：

**把 Phase 1 已经锁定的边界单独固化，避免后续实现阶段在关键问题上反复摇摆。**
