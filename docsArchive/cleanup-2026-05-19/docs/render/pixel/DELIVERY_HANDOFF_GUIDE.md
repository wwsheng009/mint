# Pixel 渲染方案交付与移交指南

## 1. 文档目的

当前 `docs/render/pixel` 已经不是一份单文档方案，而是一套完整的交付包。  
这意味着它已经足够支持：

- 架构评审
- 实施拆解
- 代码开工
- benchmark 设计
- prototype 页面落地

但文档一多，就会出现一个现实问题：

> 不同角色进入这套方案时，应该先看哪几份文档，哪些是必须先读，哪些是按需阅读。

本文件就是这份移交指南。

## 2. 这套交付包已经包含什么

当前 `pixel` 目录已经覆盖了 6 类信息。

### 2.1 方向与总体架构

- [PIXEL_CHART_RENDERING_ARCHITECTURE.md](./PIXEL_CHART_RENDERING_ARCHITECTURE.md)

回答：

- 为什么不能再只打磨字符图表
- 为什么这是引擎升级，而不是单个组件增强
- 目标架构是什么

### 2.2 当前系统冲击与可行性审查

- [CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md](./CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md)

回答：

- 当前代码链到底是什么
- 哪些层会被冲击
- 哪些层必须保持稳定
- 可行性验证怎么做

### 2.3 性能、优化与止损

- [PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md](./PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md)

回答：

- 为什么 image mode 会更重
- 要如何优化
- 应该测哪些指标
- 哪些情况必须止损

### 2.4 第一阶段的路线与拆解

- [LINECHART_IMAGE_PROTOTYPE_PLAN.md](./LINECHART_IMAGE_PROTOTYPE_PLAN.md)
- [PHASE1_TASK_BREAKDOWN.md](./PHASE1_TASK_BREAKDOWN.md)
- [IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md](./IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md)

回答：

- 为什么第一阶段只选 `linechart`
- 为什么先做 capability，再做 renderer，再做 prototype
- PR 应该怎么切

### 2.5 API 与实现规格

- [GROUP_A_GRAPHICS_CAPABILITY_SPEC.md](./GROUP_A_GRAPHICS_CAPABILITY_SPEC.md)
- [GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md](./GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md)
- [GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md](./GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md)
- [GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md](./GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md)
- [RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md](./RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md)
- [RUNTIME_PAINT_SCENE_API_SKETCH.md](./RUNTIME_PAINT_SCENE_API_SKETCH.md)
- [FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md](./FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md)

回答：

- 具体文件怎么拆
- 第一版接口怎么写
- `App.render()` 顺序怎么接
- `linechart` image path 该做到哪一步

### 2.6 Prototype 与产物落位

- [EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md](./EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md)
- [PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md](./PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md)

回答：

- prototype 目录怎么建
- benchmark 与 diagnostics 放哪里
- 什么能进仓库，什么不能进仓库

## 3. 不同角色的推荐阅读顺序

## 3.1 架构设计者

推荐阅读顺序：

1. [PIXEL_CHART_RENDERING_ARCHITECTURE.md](./PIXEL_CHART_RENDERING_ARCHITECTURE.md)
2. [CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md](./CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md)
3. [PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md](./PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md)
4. [DECISION_LOG.md](./DECISION_LOG.md)

目标：

- 判断这条路线是否值得立项
- 明确哪些决策已经固定，哪些还留待后续

## 3.2 平台层实现者

推荐阅读顺序：

1. [GROUP_A_GRAPHICS_CAPABILITY_SPEC.md](./GROUP_A_GRAPHICS_CAPABILITY_SPEC.md)
2. [RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md](./RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md)
3. [IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md](./IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md)
4. [DECISION_LOG.md](./DECISION_LOG.md)

目标：

- 直接进入 PR1 / PR3 范围

## 3.3 `paint` / render-loop 实现者

推荐阅读顺序：

1. [GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md](./GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md)
2. [RUNTIME_PAINT_SCENE_API_SKETCH.md](./RUNTIME_PAINT_SCENE_API_SKETCH.md)
3. [FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md](./FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md)
4. [DECISION_LOG.md](./DECISION_LOG.md)

目标：

- 直接进入 PR2 / PR4 范围

## 3.4 `linechart` 组件实现者

推荐阅读顺序：

1. [LINECHART_IMAGE_PROTOTYPE_PLAN.md](./LINECHART_IMAGE_PROTOTYPE_PLAN.md)
2. [GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md](./GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md)
3. [FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md](./FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md)
4. [DECISION_LOG.md](./DECISION_LOG.md)

目标：

- 直接进入 PR5 范围

## 3.5 Prototype / benchmark 实现者

推荐阅读顺序：

1. [GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md](./GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md)
2. [EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md](./EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md)
3. [PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md](./PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md)
4. [IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md](./IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md)

目标：

- 直接进入 PR6 范围

## 3.6 评审者 / Tech Lead

推荐阅读顺序：

1. [DECISION_LOG.md](./DECISION_LOG.md)
2. [IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md](./IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md)
3. [CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md](./CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md)
4. [PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md](./PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md)

目标：

- 快速确认这条方案的边界、风险和执行顺序

## 4. 当前已经固定、不应再反复争论的事项

这些事项建议直接视为 Phase 1 基线：

1. 第一阶段只选 `linechart image prototype`
2. 第一阶段只要求 `GraphicsModeNone + GraphicsModeKitty`
3. `paint.Buffer` 继续只承载文本层
4. 图像层通过 `SceneFrame` 增量表达
5. 含图像层的实验帧先绕过 `AsyncRenderer`
6. `linechart` 第一阶段只像素化 plot area
7. 标题、legend、axis label 继续留在文本层
8. 当前 `HitMap` 仍按文本路径更新
9. diagnostics / benchmark artifact 不进 `ui/e2e/testdata`
10. 失败时必须稳定回退到 text mode

更完整版本见：

- [DECISION_LOG.md](./DECISION_LOG.md)

## 5. 当前还保留开放空间的事项

这些项可以后续再做，不应挡住 Phase 1：

- 是否支持 `Sixel`
- 是否需要 image-aware `AsyncRenderer`
- 是否支持 image 交互命中测试
- 是否做图像 golden CI
- `scatterplot / heatmap / candlestick` 何时进入 image mode

## 6. 推荐启动方式

如果现在开始真正开工，推荐只做两步：

### 6.1 第一步

按 [IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md](./IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md) 开 PR1：

- `runtime/platform/graphics.go`
- `graphics_env.go`
- `graphics_probe.go`
- 对应测试

### 6.2 第二步

PR1 稳定后再开 PR2：

- `runtime/paint/scene.go`
- 最小 `SceneFrame / ImageLayer / SceneDiagnostics`

也就是说：

- 不要直接跳到 `linechart image renderer`

## 7. 这一交付包适合用在哪些场景

当前这套交付包已经适合：

- 立项评审
- 技术方案评审
- 团队拆任务
- 开第一批代码 PR

但它还不是：

- 已完成实现
- 已验证跨终端兼容
- 已具备生产稳定性结论

## 8. 一句话结论

这份移交指南的目的只有一个：

**让不同角色拿到这套 pixel 文档后，都能快速知道先看什么、先做什么、哪些决策已经锁定。**
