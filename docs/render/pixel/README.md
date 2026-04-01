# Pixel 渲染方案文档

本目录保存“通过图片像素方式渲染图表”和“为此需要的渲染引擎升级”相关方案。

这条线的核心结论是：

- 单纯继续优化 `linechart`、`scatterplot`、`heatmap` 的字符字形，并不能根本解决“连续性、细节表达、信息密度”问题
- 如果目标是让图表真正进入更高保真度的渲染层，必须把当前引擎从“纯字符 cell 渲染”扩展成“文本 + 图像”的混合渲染架构
- 这不是单个 chart 组件可以独立完成的工作，而是 `framework/app`、`runtime/paint`、`runtime/platform`、能力探测、缓存与测试链路的联合改造

## 当前结论

当前引擎仍然是标准的字符单元格渲染系统，而且主线渲染链以 `framework.App.render()` 为中心：

- `runtime/paint.Buffer` 以 cell 为核心，单元格存储 `Cluster + Style`
- `runtime/platform.Terminal` 只暴露字符层能力，没有图像协议和像素尺寸接口
- `runtime/paint.Renderer` 与 `runtime/paint.AsyncRenderer` 仍然只服务文本帧
- `framework/theme/output` 只有颜色能力，没有图形协议能力

补充说明：

- `framework/screen/manager.go` 在仓库中仍然存在，且同样体现了逐 cell diff 的历史设计思路
- 但当前 charts 主线真正经过的是 `framework/app.go -> runtime/paint.Renderer / AsyncRenderer`
- 因此 pixel 方案的第一落点应优先围绕 `App.render()`、`Renderer`、`AsyncRenderer` 和 `runtime/platform`

因此，现阶段的图表组件只能走：

- Unicode 线段
- Braille / Block / ASCII
- 颜色梯度和字符密度

还不能真正走“图片像素图表”。

## 文档列表

### 0. [DELIVERY_HANDOFF_GUIDE.md](./DELIVERY_HANDOFF_GUIDE.md)

整套交付包的移交指南，覆盖：

- 不同角色该先看什么
- 当前已经固定的事项
- 当前仍然开放的事项
- 推荐的启动方式

### 0.1 [DECISION_LOG.md](./DECISION_LOG.md)

Phase 1 决策日志，覆盖：

- 哪些关键决策已经锁定
- 哪些事项仍保留开放空间
- 后续如果要变更，应如何更新

### 1. [PIXEL_CHART_RENDERING_ARCHITECTURE.md](./PIXEL_CHART_RENDERING_ARCHITECTURE.md)

主方案文档，覆盖：

- 为什么当前引擎不能直接支持图片图表
- 引擎需要补哪些抽象层
- 推荐的接口拆分
- 渐进式落地顺序
- 风险与 fallback 设计

### 2. [CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md](./CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md)

结合当前系统实现的审查文档，覆盖：

- 当前真实渲染路径
- 方案对 `App / Renderer / Buffer / Platform / Event / Testing` 的冲击
- 哪些层可以保留，哪些层必须升级
- 如何做最小可行原型验证
- 成功标准与止损条件

### 3. [PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md](./PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md)

性能专项文档，覆盖：

- 像素图表对 CPU、带宽、内存与 renderer 的影响
- 可以采用哪些系统级优化手段
- 应该采集哪些 benchmark 指标
- 何时应当止损或自动回退到文本路径

### 4. [LINECHART_IMAGE_PROTOTYPE_PLAN.md](./LINECHART_IMAGE_PROTOTYPE_PLAN.md)

第一份可执行的 PoC 计划，覆盖：

- 为什么第一阶段应从 `linechart` 切入
- 第一阶段明确不做什么
- 应从哪些包和文件开始改
- 如何定义成功标准与止损条件

### 5. [PHASE1_TASK_BREAKDOWN.md](./PHASE1_TASK_BREAKDOWN.md)

第一阶段执行拆解文档，覆盖：

- Phase 1 的任务分组
- 每组建议改动的包与文件
- 推荐执行顺序
- 验收清单
- 风险控制点与建议的提交拆分

### 6. [GROUP_A_GRAPHICS_CAPABILITY_SPEC.md](./GROUP_A_GRAPHICS_CAPABILITY_SPEC.md)

第一阶段 Group A 的技术规格，覆盖：

- 为什么图形能力层必须先于 image prototype 落地
- `GraphicsMode / GraphicsCapabilities` 应包含哪些最小字段
- 如何以可选接口接入 `runtime/platform`，而不破坏现有 `Terminal / Screen`
- Phase 1 的探测策略、环境变量覆盖和测试范围

### 7. [GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md](./GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md)

第一阶段 Group B / Group C 的接入规格，覆盖：

- 最小 `SceneFrame / ImageLayer` 应长什么样
- 为什么 Phase 1 不能直接把图像塞进 `paint.Buffer`
- `App.render()` 应如何收集和提交 image layer
- 为什么第一阶段图像帧应先绕过 `AsyncRenderer`，走同步提交闭环

### 8. [GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md](./GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md)

第一阶段 Group D 的 `linechart` image renderer 规格，覆盖：

- 为什么第一阶段只应该把 plot area 交给 image renderer
- 哪些内容继续保留在文本层
- `linechart` image path 的支持范围、回退策略和成功标准

### 9. [GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md](./GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md)

第一阶段 Group E 的验证闭环规格，覆盖：

- prototype 页面应展示什么
- benchmark 至少要测哪些指标
- diagnostics 应保存哪些产物
- image prototype 的 go / no-go 判断条件

### 10. [RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md](./RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md)

面向实际编码的 `runtime/platform/graphics` 草案，覆盖：

- `graphics.go / graphics_env.go / graphics_probe.go / graphics_kitty.go` 建议如何拆
- `GraphicsCapabilities / DrawImageRequest / GraphicsPresenter` 的第一版签名
- 环境变量覆盖、探测伪代码和第一批测试清单

### 11. [RUNTIME_PAINT_SCENE_API_SKETCH.md](./RUNTIME_PAINT_SCENE_API_SKETCH.md)

面向 `runtime/paint/scene.go` 的 API 草案，覆盖：

- `SceneFrame / ImageLayer / SceneDiagnostics` 第一版应有哪些字段
- 为什么 `SceneFrame` 应存渲染语义，而不是直接存 `DrawImageRequest`
- Scene helper、内存策略与测试建议

### 12. [FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md](./FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md)

面向 `framework.App.render()` 的接入流程规格，覆盖：

- 文本层、Scene 构造、图像提交、`HitMap` 更新的推荐顺序
- 为什么第一阶段含图像帧必须绕过 `AsyncRenderer`
- 首帧、resize、失败回退和清理的建议行为

### 13. [EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md](./EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md)

面向 prototype 示例目录的布局规格，覆盖：

- `examples/charts_linechart_image_prototype` 建议目录结构
- 页面应展示的最小区域
- 它和现有 `charts_linechart_demo` 的职责分工
- 第一阶段与 `ui/e2e` 的衔接边界

### 14. [PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md](./PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md)

面向 benchmark 与 diagnostics 产物的布局规格，覆盖：

- 哪些产物进仓库，哪些不进仓库
- `artifacts/pixel/...` 的建议目录结构
- capability、benchmark、scene、bitmap 的建议文件格式
- 为什么这些产物不能直接混进 `ui/e2e/testdata`

### 15. [IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md](./IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md)

面向真正交付的实施计划，覆盖：

- Phase 1 的推荐实施顺序
- 从 PR1 到 PR6 的切分建议
- 每个 PR 的写入边界、验收重点与中止点
- 里程碑划分与团队执行顺序

## 推荐路线

建议按三段推进，而不是一步到位：

1. 先定义图形能力探测、像素尺寸、混合绘制原语
2. 再做实验性 image backend，只验证少数终端协议和单个 chart
3. 最后让 `linechart / heatmap / scatterplot` 逐步拥有 text mode 与 image mode 双路径

## 不建议的做法

以下方向短期内不建议继续投入过多：

- 继续只靠 `─ │ ╱ ╲` 打磨字符折线
- 在没有图像能力探测的前提下硬写某种终端图像协议
- 直接把图片图表嵌入现有 cell diff 流程，而不引入图层或区域缓存

## 相关代码位置

理解本方案时，建议先对照这些当前实现：

- `runtime/paint/buffer.go`
- `runtime/paint/renderer.go`
- `runtime/paint/async_renderer.go`
- `runtime/platform/terminal.go`
- `runtime/platform/screen.go`
- `framework/app.go`
- `framework/theme/output.go`
- `ui/components/charts/*`

## 状态

- 当前状态：方案保存，尚未实现
- 推荐优先级：中高
- 前置条件：charts 第一阶段文本渲染链路已基本稳定
