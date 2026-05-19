# Pixel 渲染实施顺序与 PR 切分计划

## 1. 文档目的

到目前为止，`docs/render/pixel` 已经具备：

- 总架构
- 当前系统冲击审查
- 性能影响与优化
- PoC 路线
- Phase 1 分组拆解
- Group A 到 Group E 的技术规格
- `runtime/platform`、`runtime/paint`、`framework/app` 的 API 草案
- prototype 目录与 artifact 布局规范

现在缺的不是更多方案，而是一份真正面向交付的文档：

> 代码阶段应该按什么顺序实施，应该切成哪些 PR，每个 PR 的目标、边界、风险和验收是什么。

本文件就是这份交付计划。

## 2. 总体交付原则

### 2.1 先证实可行，再扩大范围

整个 pixel 方案第一阶段必须坚持：

- 先证实 `linechart image prototype` 值得做
- 再考虑是否把能力扩展到更多图表

### 2.2 先搭能力层，再碰组件层

顺序必须是：

1. `runtime/platform`
2. `runtime/paint`
3. `framework/app`
4. `ui/components/charts/linechart`
5. `examples`

不要反过来先写 `linechart image renderer`。

### 2.3 每个 PR 都必须可止损

也就是说，每个 PR 即使后面整体方案不继续推进，也应当：

- 有清晰边界
- 不破坏现有文本主链
- 可独立评审
- 可独立回滚

## 3. 推荐实施顺序

### Step 1：平台图形能力层

目标：

- 建立 `GraphicsMode / GraphicsCapabilities`
- 能稳定区分 `None` 和 `Kitty`
- 提供环境变量覆盖

对应文档：

- [GROUP_A_GRAPHICS_CAPABILITY_SPEC.md](./GROUP_A_GRAPHICS_CAPABILITY_SPEC.md)
- [RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md](./RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md)

### Step 2：`paint.SceneFrame` 最小层

目标：

- 在不动 `Buffer` 的前提下，新增 `SceneFrame / ImageLayer / SceneDiagnostics`

对应文档：

- [GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md](./GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md)
- [RUNTIME_PAINT_SCENE_API_SKETCH.md](./RUNTIME_PAINT_SCENE_API_SKETCH.md)

### Step 3：`App.render()` 接入同步图像旁路

目标：

- 文本帧仍按原路径工作
- 含 image layer 的实验帧先走同步提交
- `AsyncRenderer` 第一阶段不改

对应文档：

- [GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md](./GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md)
- [FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md](./FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md)

### Step 4：`linechart` plot-only image renderer

目标：

- 只把 plot area 切到 image path
- 标题、legend、axis label 继续留在文本层

对应文档：

- [GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md](./GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md)
- [LINECHART_IMAGE_PROTOTYPE_PLAN.md](./LINECHART_IMAGE_PROTOTYPE_PLAN.md)

### Step 5：prototype 页面与 benchmark/diagnostics

目标：

- 新增 `examples/charts_linechart_image_prototype`
- 形成 text vs image 对照页面
- 输出 capability / benchmark / diagnostics artifact

对应文档：

- [GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md](./GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md)
- [EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md](./EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md)
- [PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md](./PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md)

## 4. 推荐 PR 切分

建议至少拆成 6 个 PR，而不是 1 个大 PR。

## PR1：Graphics Capability 基础层

### 目标

- 新增 `runtime/platform/graphics.go`
- 新增 `graphics_env.go`
- 新增 `graphics_probe.go`
- 建立 `GraphicsCapabilities` 与 override 探测逻辑

### 主要写入范围

- `runtime/platform/*`
- 必要测试文件

### 验收标准

- 能稳定返回 `GraphicsModeNone`
- `MINT_GRAPHICS` 覆盖逻辑可测
- `MINT_CELL_PIXELS` 解析可测
- 不影响现有平台实现编译

### 风险控制

- 本 PR 不碰 `App.render()`
- 本 PR 不碰 charts 组件

## PR2：实验性 SceneFrame 与图像层数据模型

### 目标

- 新增 `runtime/paint/scene.go`
- 建立 `SceneFrame / ImageLayer / SceneDiagnostics`
- 不修改 `Buffer`

### 主要写入范围

- `runtime/paint/*`

### 验收标准

- Scene 类型可独立编译与测试
- 不影响现有 `Renderer` / `AsyncRenderer`

### 风险控制

- 本 PR 不接 `App.render()`
- 本 PR 不接任何图像协议

## PR3：Kitty Presenter 与最小图像生命周期

### 目标

- 新增实验性 `graphics_kitty.go`
- 建立 `Present / Replace / Delete / Clear`

### 主要写入范围

- `runtime/platform/*`

### 验收标准

- presenter 对象可独立工作
- 生命周期操作具备最小测试

### 风险控制

- 本 PR 仍不改 charts
- 本 PR 仍不改 `App.render()`

## PR4：`App.render()` 图像旁路接入

### 目标

- 允许根节点提供 `SceneFrame`
- 含图像层的实验帧走同步文本输出 + 图像提交

### 主要写入范围

- `framework/app.go`
- 少量 `runtime/paint` / `runtime/platform` 辅助代码

### 验收标准

- 纯文本页面保持原行为
- 含图像层时不经过 `AsyncRenderer`
- 图像失败时文本回退稳定

### 风险控制

- 不修改 `AsyncRenderer.SubmitFrame()` 签名
- 不改现有文本 e2e 主链

## PR5：`linechart` image prototype backend

### 目标

- 给 `linechart` 新增实验性 backend 选择
- plot-only image renderer

### 主要写入范围

- `ui/components/charts/linechart/*`

### 验收标准

- 文本路径既有测试通过
- image path 可运行
- plot 连续性明显改善

### 风险控制

- 第一阶段只要求单系列优先
- 不把 legend/title/axis label 图像化

## PR6：Prototype 页面、benchmark 与 diagnostics

### 目标

- 新增 `examples/charts_linechart_image_prototype`
- 输出 `artifacts/pixel/...`
- 完成最小验证闭环

### 主要写入范围

- `examples/charts_linechart_image_prototype/*`
- 少量 benchmark/diagnostics 辅助代码

### 验收标准

- text vs image 对照页面可运行
- capability / benchmark / scene summary 可导出
- 有明确 go / no-go 判断依据

### 风险控制

- 不把实验产物混进 `ui/e2e/testdata`
- 不强求 image golden CI

## 5. 每个 PR 的评审重点

为了防止评审时重新发散，建议每个 PR 都只关注对应问题。

### PR1 评审重点

- 类型是否克制
- override 与启发式是否保守
- 是否引入了不必要的平台改动

### PR2 评审重点

- Scene 是否足够轻
- 是否污染了 `Buffer`
- 是否引入了 `paint -> platform` 反向依赖

### PR3 评审重点

- presenter 生命周期边界是否清晰
- 协议细节是否封装在 `platform`

### PR4 评审重点

- `App.render()` 时序是否稳定
- 纯文本主链是否保持不变
- 是否错误地把图像接进了 `AsyncRenderer`

### PR5 评审重点

- `linechart` image path 是否只覆盖 plot area
- 是否保留了文本 fallback
- 是否过早承载了多系列/复杂交互

### PR6 评审重点

- prototype 是否真正可用于决策
- benchmark 指标是否足够
- artifact 目录与仓库边界是否清晰

## 6. 推荐的中止点

整个 Phase 1 不应该“一路硬做到底”，而应在几个节点设置中止判断。

### 中止点 A：PR1 后

如果图形能力探测本身就不稳定：

- 暂停后续 image mode 研发

### 中止点 B：PR4 后

如果 `App.render()` 接入图像旁路后明显扰乱文本主链：

- 暂停后续 `linechart` image backend

### 中止点 C：PR5 后

如果 `linechart` image path 视觉收益有限或复杂度过高：

- 不进入 PR6 的 prototype 扩展

## 7. 推荐里程碑

为了便于团队协作和状态汇报，建议设 3 个里程碑。

### 里程碑 M1：能力层就绪

完成：

- PR1
- PR2

意味着：

- 具备 capability 与 scene 最小基础

### 里程碑 M2：渲染链旁路就绪

完成：

- PR3
- PR4

意味着：

- 平台与 `App.render()` 已具备实验性图像提交闭环

### 里程碑 M3：原型验证完成

完成：

- PR5
- PR6

意味着：

- 拥有可以做 go / no-go 判断的 `linechart image prototype`

## 8. 和当前文档体系的关系

这份文档不是替代已有文档，而是把它们串成执行顺序。

建议把阅读顺序固定为：

1. [PIXEL_CHART_RENDERING_ARCHITECTURE.md](./PIXEL_CHART_RENDERING_ARCHITECTURE.md)
2. [CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md](./CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md)
3. [PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md](./PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md)
4. [PHASE1_TASK_BREAKDOWN.md](./PHASE1_TASK_BREAKDOWN.md)
5. Group A-E 规格与 API 草案
6. 本文

也就是说：

- 前面解决“为什么”和“做什么”
- 本文解决“按什么顺序交付”

## 9. 一句话结论

这份实施计划的核心目标是：

**把 pixel 方案从“文档已完整”推进到“开发团队可以直接按 PR 顺序开工、评审与止损”的状态。**
