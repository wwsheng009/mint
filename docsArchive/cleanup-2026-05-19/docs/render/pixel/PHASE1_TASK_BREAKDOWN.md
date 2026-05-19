# Pixel 渲染 Phase 1 任务拆解

## 1. 文档目的

本文件把 `LINECHART_IMAGE_PROTOTYPE_PLAN.md` 再往前推进一层，目标是把第一阶段从“PoC 计划”拆成可执行任务。

第一阶段的核心目标只有一个：

> 在不破坏当前文本渲染主链的前提下，完成 `linechart image prototype` 的最小验证闭环。

这里的“闭环”指的是：

- 有能力探测
- 有最小图像 backend
- 有 `linechart` 的 image path
- 有 prototype demo
- 有性能与稳定性验证
- 失败时可完整回退

## 2. Phase 1 交付范围

Phase 1 只允许覆盖以下范围：

- `runtime/platform`
- `runtime/paint`
- `framework/app.go`
- `ui/components/charts/linechart`
- `examples/charts_linechart_image_prototype`
- 必要的 `docs/render/pixel`

明确不做：

- 其他 charts 组件 image path
- 全局 renderer 重构
- 多协议兼容
- 复杂交互
- 统一图像 snapshot 测试体系

## 3. 任务分组

建议按 5 组任务推进。

### Group A：Graphics 能力层

目标：

- 为图像渲染建立最小能力探测模型

建议文件：

- `runtime/platform/graphics.go`
- 必要时补测试文件

建议新增内容：

- `GraphicsMode`
- `GraphicsCapabilities`
- `Graphics` 可选接口
- capability 获取入口

本组验收：

- 支持明确返回 `GraphicsModeNone`
- 在支持终端环境里能拿到基础 capability
- 不影响现有 `Terminal` / `Screen` 代码路径

本组风险：

- 能力探测如果不稳定，后续全部都会建立在不可靠前提上

### Group B：Paint / Scene 最小扩展

目标：

- 在不推翻 `paint.Buffer` 的前提下，增加最小图像层承载能力

建议文件：

- `runtime/paint/scene.go`
- 必要时补 `renderer.go` 辅助结构

建议新增内容：

- `Scene`
- `ImageLayer`
- `DrawImageRequest`
- 最小 frame 组合结构

关键约束：

- 文本层继续使用现有 `paint.Buffer`
- 图像层作为附加层，而不是把 bitmap 塞进 `Buffer.Cells`

本组验收：

- 能表达“一帧文本 + 一张图像”
- 不影响现有文本 snapshot 与 e2e 能力

本组风险：

- 如果 scene 设计过重，Phase 1 会被架构工作吞掉

### Group C：`App.render()` 最小接入

目标：

- 让主渲染循环能在文本输出之外提交 image layer

建议文件：

- `framework/app.go`

建议实现策略：

- 继续正常生成 `paint.Buffer`
- 如果本帧存在 image layer，则在文本渲染之后做图像提交
- 不改掉现有文本 renderer 主路径

本组验收：

- 普通 UI 页面不受影响
- prototype 页面可输出 image layer
- 首帧、重绘、关闭 alternate screen 不出现明显破坏

本组风险：

- 文本和图像提交时序不同步
- 异步 renderer 路径出现闪烁或乱序

### Group D：`linechart` image path

目标：

- 给 `linechart` 新增实验性 image backend

建议文件：

- `ui/components/charts/linechart/vnode.go`
- `ui/components/charts/linechart/builder.go`
- `ui/components/charts/linechart/instance.go`
- 视需要新增 `image_renderer.go`

建议新增能力：

- `RenderBackend(...)`
- `ImageMode()`
- `TextMode()`

建议第一版限制：

- 单系列优先
- 固定线宽
- 点位先可选关闭
- legend 和复杂 grid 可先弱化

本组验收：

- 文本路径仍能工作
- image path 能在 prototype 里显示
- 同尺寸下图像线条连续性明显优于文本路径

本组风险：

- image path 过早承载多系列、legend、grid，会让 PoC 复杂度失控

### Group E：Prototype / Benchmark / Diagnostics

目标：

- 建立最小验证闭环

建议文件：

- `examples/charts_linechart_image_prototype/main.go`
- `examples/charts_linechart_image_prototype/README.md`
- prototype 测试或 benchmark 文件

建议能力：

- 文本与 image `linechart` 并排展示
- 输出 capability 信息
- 输出简单时间统计
- 保存 diagnostics

本组验收：

- 能肉眼对比 text vs image
- 能测首帧和更新成本
- 失败时能保留诊断信息

本组风险：

- 如果没有 benchmark 和 diagnostics，PoC 很容易陷入“感觉更好/感觉更慢”的争论

## 4. 推荐执行顺序

推荐顺序不是按组件，而是按依赖关系走：

1. Group A：能力层
2. Group B：scene 最小扩展
3. Group C：`App.render()` 接入
4. Group D：`linechart` image path
5. Group E：prototype 与 benchmark

原因：

- 如果先写 `linechart image path`，后面会反向逼迫平台层和渲染层返工
- 先把 capability 与 scene 立起来，后续路径会稳定很多

## 5. 每组的产出物

### Group A

产出物：

- 代码接口
- capability dump 示例
- 最小单测

### Group B

产出物：

- `Scene` / `ImageLayer` 基础结构
- 文本帧与图像层组合能力

### Group C

产出物：

- `App.render()` 可处理 image layer
- 最小集成测试

### Group D

产出物：

- `linechart` image backend
- builder 控制项
- 文本 / 图像双路径

### Group E

产出物：

- prototype demo
- benchmark 记录方式
- diagnostics 保存方式

## 6. 建议的验收检查表

### 6.1 基础检查

- [ ] 不支持图像协议时自动回退到文本路径
- [ ] 支持图像协议时能稳定进入 image path
- [ ] `linechart` 文本路径既有测试不被打坏
- [ ] prototype 页面可以独立运行

### 6.2 视觉检查

- [ ] image path 下折线连续性明显优于文本模式
- [ ] 小尺寸下走势辨识度明显提升
- [ ] resize 后图像不残影、不漂移

### 6.3 性能检查

- [ ] 首帧时间已记录
- [ ] 单次更新重绘时间已记录
- [ ] 输出量已记录
- [ ] 至少有一个缓存命中指标

### 6.4 稳定性检查

- [ ] 退出页面后没有图像残留
- [ ] alternate screen 切换正常
- [ ] 图像失败时不影响文本 UI

## 7. 建议的风险控制点

Phase 1 最容易失控的地方有三处。

### 7.1 过早做“通用图像架构”

风险：

- 在还没证明 `linechart image path` 值得做之前，就把渲染层抽象得过大

控制方式：

- 先做最小 scene
- 只满足单图 prototype

### 7.2 过早追求多协议

风险：

- 协议差异会把 Phase 1 复杂度放大

控制方式：

- Phase 1 只选一种协议
- 其余协议留到 Phase 2 以后

### 7.3 过早引入复杂交互

风险：

- 事件映射、命中测试、局部放大都会让 PoC 偏题

控制方式：

- Phase 1 只做静态显示和受控重绘

## 8. 建议的提交拆分

为了控制回滚与评审成本，建议至少拆成 4 到 5 个提交：

1. `platform: add graphics capability primitives`
2. `paint: add experimental scene/image layer types`
3. `framework: wire image layer into app render loop`
4. `charts/linechart: add experimental image backend`
5. `examples: add linechart image prototype and benchmark notes`

这样即使中途止损，也能保留：

- 能力层
- scene 层
- 组件层

的独立边界。

## 9. 推荐下一步

如果继续推进，建议优先完成并维护下面这组配套规格：

- `GROUP_A_GRAPHICS_CAPABILITY_SPEC.md`
- `GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md`
- `GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md`
- `GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md`

原因：

- 只有把 Group A 到 Group E 的关键约束都写清，Phase 1 才不会在实现阶段重新回到架构争论
- 其中 Group A 和 Group B/C 决定边界，Group D 和 Group E 决定 prototype 是否真的能闭环

一句话总结：

**Phase 1 的关键不是“先把图画出来”，而是“先把最小图像能力链路建立起来，并且可回退、可验证、可止损”。**
