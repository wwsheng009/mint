# Pixel Phase 1 团队执行工单

## 1. 文档目的

本文件把 `docs/render/pixel` 现有交付包进一步收敛成可直接分派给团队成员的执行工单。

目标不是重复方案文档，而是明确：

- 每个工单谁负责
- 写入边界在哪里
- 前置依赖是什么
- 验收与止损点是什么
- 应该用什么分支和测试命令推进

本文件基于以下已固定文档和基线提交整理：

- `docs/pixel-rendering-poc-plan`
- `2aa95a26`
- [README.md](./README.md)
- [IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md](./IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md)
- [DECISION_LOG.md](./DECISION_LOG.md)
- [DELIVERY_HANDOFF_GUIDE.md](./DELIVERY_HANDOFF_GUIDE.md)

## 2. 角色建议

建议至少按 5 个 owner 组织执行：

- 成员 A：Platform Core Owner
- 成员 B：Paint Scene Owner
- 成员 C：Kitty Presenter Owner
- 成员 D：Render Loop Owner
- 成员 E：Chart + Prototype Owner
- Tech Lead：接口冻结、闸门判断、评审协调，不跨 owner 文件直接实现

如果团队只有 3 人，建议合并为：

- A：`PR1 + PR3`
- B：`PR2 + PR4`
- C：`PR5 + PR6`

## 3. 全局执行规则

### 3.1 唯一事实来源

实现阶段统一以以下文档为准：

1. [DECISION_LOG.md](./DECISION_LOG.md)
2. [IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md](./IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md)
3. 对应组规格与 API 草案

如果实现与文档冲突，先更新文档，再改代码。

### 3.2 合并顺序

合并顺序固定为：

1. `PR1`
2. `PR2`
3. `PR3`
4. `PR4`
5. `PR5`
6. `PR6`

允许提前准备后续 PR，但不建议跳过前置依赖直接合并。

### 3.3 文件 owner 锁

为避免高冲突文件被多人同时改动，执行阶段采用 owner 锁：

- `runtime/platform/graphics.go` 在 `PR1` 合并前只允许成员 A 修改
- `framework/app.go` 的 image bypass 相关改动只允许成员 D 修改
- `ui/components/charts/linechart/*.go` 的 image backend 改动只允许成员 E 修改
- 任何人都不得把 diagnostics / benchmark / bitmap 产物写入 `ui/e2e/testdata`

### 3.4 fallback 规则

Phase 1 有一条硬规则：

- 任何 image 相关失败都必须稳定回退到 text path

包括：

- capability 不可靠
- scene 构建失败
- presenter 提交失败
- prototype 环境不满足

### 3.5 分支命名

建议统一使用：

- `feature/pixel-pr1-graphics-capability-a`
- `feature/pixel-pr2-sceneframe-b`
- `feature/pixel-pr3-kitty-presenter-c`
- `feature/pixel-pr4-app-render-bypass-d`
- `feature/pixel-pr5-linechart-image-backend-e`
- `feature/pixel-pr6-prototype-diagnostics-e`

## 4. Ticket T1 / PR1

### 标题

`runtime/platform: add graphics capability foundation`

### 推荐 Owner

- 成员 A

### 推荐 Reviewer

- Tech Lead
- 成员 C

### 前置依赖

- 无

### 建议分支

- `feature/pixel-pr1-graphics-capability-a`

### 背景

Phase 1 必须先建立图形能力层，后续 `SceneFrame`、`Kitty presenter`、`App.render()` image bypass 和 `linechart` image backend 都会依赖这里定义的最小契约。

对应文档：

- [GROUP_A_GRAPHICS_CAPABILITY_SPEC.md](./GROUP_A_GRAPHICS_CAPABILITY_SPEC.md)
- [RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md](./RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md)

### 范围

- 新增 `runtime/platform/graphics.go`
- 新增 `runtime/platform/graphics_env.go`
- 新增 `runtime/platform/graphics_probe.go`
- 新增对应测试文件
- 冻结 `GraphicsMode / GraphicsCapabilities / DrawImageRequest / Provider` 第一版签名

### 非范围

- 不实现 `Kitty presenter`
- 不改 `App.render()`
- 不改 charts 组件
- 不改 `runtime/platform/platform.go`
- 不改 `runtime/platform/screen.go`
- 不改 `runtime/platform/terminal.go`

### 写入范围

- `runtime/platform/*`
- 必要测试文件

### 交付物

- `GraphicsModeNone + GraphicsModeKitty` 最小能力枚举
- `GraphicsCapabilities`
- `DrawImageRequest`
- `GraphicsCapabilityProvider`
- `GraphicsPresenter`
- `GraphicsPresenterProvider`
- `MINT_GRAPHICS`
- `MINT_CELL_PIXELS`
- 保守 probe 和 env override 逻辑

### 验收标准

- 能稳定返回 `GraphicsModeNone`
- `MINT_GRAPHICS=off/kitty/auto` 行为可测
- `MINT_CELL_PIXELS` 成功与失败解析可测
- 未知环境默认保守回退
- 不影响现有平台实现编译

### 风险与回退

- 风险：基础类型定义过重会拖慢后续所有 PR
- 风险：提前改主平台接口会把范围无谓放大
- 回退：如果 probe 不可靠，强制回退到 `GraphicsModeNone`

### 测试命令

```bash
go test ./runtime/platform
```

### 评审重点

- 类型是否克制
- override 优先级是否清晰
- heuristic 是否足够保守
- 是否没有把图像能力硬塞进现有 `Terminal / Screen / RuntimePlatform`

## 5. Ticket T2 / PR2

### 标题

`runtime/paint: add experimental SceneFrame model`

### 推荐 Owner

- 成员 B

### 推荐 Reviewer

- Tech Lead
- 成员 D

### 前置依赖

- 建议在 `PR1` 接口冻结后开始

### 建议分支

- `feature/pixel-pr2-sceneframe-b`

### 背景

Phase 1 已固定：

- `paint.Buffer` 继续只承载文本层
- 图像层通过新的 `SceneFrame` 增量表达

因此 `PR2` 的目标是先把最小数据模型落下来，而不是直接接入 render loop。

对应文档：

- [GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md](./GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md)
- [RUNTIME_PAINT_SCENE_API_SKETCH.md](./RUNTIME_PAINT_SCENE_API_SKETCH.md)

### 范围

- 新增 `runtime/paint/scene.go`
- 新增 `runtime/paint/scene_test.go`
- 定义 `SceneFrame / ImageLayer / SceneDiagnostics`

### 非范围

- 不改 `runtime/paint/buffer.go`
- 不改 `runtime/paint/renderer.go`
- 不改 `runtime/paint/async_renderer.go`
- 不改 `framework/app.go`
- 不引入终端协议细节

### 写入范围

- `runtime/paint/*`

### 依赖限制

- 只允许依赖 `paint.Buffer`、`paint.Rect` 和基础渲染类型
- 不允许依赖 `framework/*`
- 不允许依赖 `runtime/platform`
- 不允许依赖具体组件或终端协议

### 验收标准

- scene 类型可独立编译与测试
- `Buffer` 语义不被污染
- 现有 `Renderer / AsyncRenderer` 行为不变

### 风险与回退

- 风险：`SceneFrame` 过早承载协议请求，导致 `paint -> platform` 反向依赖
- 风险：把 bitmap 直接塞进 `Buffer` 破坏文本主链
- 回退：若模型定义偏重，收缩回最小语义层，只保留 Phase 1 真正消费字段

### 测试命令

```bash
go test ./runtime/paint
```

### 评审重点

- Scene 是否足够轻
- 是否没有污染 `Buffer`
- 是否没有引入反向依赖

## 6. Ticket T3 / PR3

### 标题

`runtime/platform: add Kitty graphics presenter`

### 推荐 Owner

- 成员 C

### 推荐 Reviewer

- Tech Lead
- 成员 A

### 前置依赖

- `PR1` 接口冻结或已合并

### 建议分支

- `feature/pixel-pr3-kitty-presenter-c`

### 背景

`PR1` 只解决 capability 和契约，不解决实际图像对象生命周期。`PR3` 要补上 Phase 1 的实验性 `Kitty presenter`，为后续 `App.render()` 同步旁路提供最小能力闭环。

对应文档：

- [RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md](./RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md)
- [IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md](./IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md)

### 范围

- 新增 `runtime/platform/graphics_kitty.go`
- 新增 presenter 生命周期最小测试

### 非范围

- 不改 `graphics.go` 已冻结基础类型
- 不改 `App.render()`
- 不改 charts 组件
- 不做多协议并行支持

### 写入范围

- `runtime/platform/*`
- 必要测试文件

### 交付物

- `Present`
- `Replace`
- `Delete`
- `Clear`
- 最小 object ID 生命周期管理

### 验收标准

- presenter 对象可独立工作
- `Replace / Delete / Clear` 语义可测
- 协议细节封装在 `platform` 内

### 风险与回退

- 风险：试图在 `PR3` 回头重构 `PR1` 类型，导致基础层失稳
- 风险：协议实现过重，过早引入缓存和高级 diff
- 回退：保持最小生命周期语义，复杂优化延后

### 测试命令

```bash
go test ./runtime/platform
```

### 评审重点

- 生命周期边界是否清晰
- 是否没有反向污染 capability 基础层
- 协议封装是否留在 `runtime/platform`

## 7. Ticket T4 / PR4

### 标题

`framework: wire SceneFrame image bypass into App.render`

### 推荐 Owner

- 成员 D

### 推荐 Reviewer

- Tech Lead
- 成员 B
- 成员 C

### 前置依赖

- `PR2`
- `PR3`

### 建议分支

- `feature/pixel-pr4-app-render-bypass-d`

### 背景

当前 image frame 的关键风险点不在组件层，而在 [framework/app.go](./../../framework/app.go) 的 render 主链。Phase 1 已固定：

- text-only 页面继续走原路径
- 含 image layer 的实验帧绕过 `AsyncRenderer`
- image failure 必须稳定 text fallback

对应文档：

- [GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md](./GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md)
- [FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md](./FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md)

### 范围

- 修改 `framework/app.go`
- 必要时补一个 scene-provider seam
- 少量补 `runtime/paint` / `runtime/platform` 辅助代码
- 补 `App.render()` 相关集成测试

### 非范围

- 不改 `AsyncRenderer.SubmitFrame()` 签名
- 不把 image 接进 `AsyncRenderer`
- 不改 charts 组件
- 不改变 text-only 页面主链

### 写入范围

- `framework/app.go`
- `framework/app_integration_test.go`
- 少量辅助代码

### 交付物

- 根节点可选提供 `SceneFrame`
- text-only frame 继续走原路径
- image frame 走同步文本输出 + presenter 提交
- presenter 失败时稳定 fallback 到 text path

### 验收标准

- 纯文本页面保持原行为
- 含 image layer 时不经过 `AsyncRenderer`
- image 提交失败时文本仍稳定输出

### 风险与回退

- 风险：`App.render()` 分支过重，扰乱现有文本主链
- 风险：错误地把 image path 并入 async 流程
- 回退：如果 image bypass 明显干扰 text path，在 `PR4` 节点止损

### 测试命令

```bash
go test ./framework ./runtime/paint
```

### 评审重点

- `App.render()` 时序是否稳定
- fallback 是否硬且可预测
- 纯文本页面是否零回归

## 8. Ticket T5 / PR5

### 标题

`linechart: add experimental image backend for plot area`

### 推荐 Owner

- 成员 E

### 推荐 Reviewer

- Tech Lead
- 成员 D

### 前置依赖

- `PR4`

### 建议分支

- `feature/pixel-pr5-linechart-image-backend-e`

### 背景

当前 `linechart` 文本链路入口清晰：

- `builder.go`
- `vnode.go`
- `instance.go`

其中 `instance.go` 已天然分成 `header / plot / footer` 三块，适合 Phase 1 只把 plot area 切到 image path。

对应文档：

- [LINECHART_IMAGE_PROTOTYPE_PLAN.md](./LINECHART_IMAGE_PROTOTYPE_PLAN.md)
- [GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md](./GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md)

### 范围

- 给 `linechart` 增加实验性 backend 选择
- plot-only image renderer
- 保留现有文本 backend

### 非范围

- 不图像化标题
- 不图像化 legend
- 不图像化 axis label
- 不要求多系列完整支持
- 不改现有 text demo 主链

### 写入范围

- `ui/components/charts/linechart/*`

### 交付物

- 实验性 backend API
- plot area image path
- 文本 fallback

### 验收标准

- 文本路径既有测试通过
- image path 可运行
- plot 连续性明显改善
- 单系列优先可用

### 风险与回退

- 风险：同时承载多系列、legend、label 图像化，复杂度失控
- 风险：修改 `builder / vnode / instance` 三文件时与 prototype 页面并行冲突
- 回退：Phase 1 只保留 plot-only，必要时把多系列显式降级到 text path

### 测试命令

```bash
go test ./ui/components/charts/linechart
```

### 评审重点

- 是否只覆盖 plot area
- 文本 fallback 是否保留
- 是否控制在 Phase 1 范围内

## 9. Ticket T6 / PR6

### 标题

`examples: add linechart image prototype and diagnostics artifacts`

### 推荐 Owner

- 成员 E

### 推荐 Reviewer

- Tech Lead
- 成员 D

### 前置依赖

- `PR5`

### 建议分支

- `feature/pixel-pr6-prototype-diagnostics-e`

### 背景

`PR6` 的目标不是继续发散架构，而是形成真正可以做 go / no-go 判断的验证闭环。

对应文档：

- [GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md](./GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md)
- [EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md](./EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md)
- [PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md](./PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md)

### 范围

- 新增 `examples/charts_linechart_image_prototype`
- 增加 capability dump、benchmark、scene diagnostics 辅助代码
- 输出 `artifacts/pixel/...` 目录闭环

### 非范围

- 不污染现有 `examples/charts_linechart_demo`
- 不修改现有 text e2e snapshot
- 不引入 image golden CI

### 写入范围

- `examples/charts_linechart_image_prototype/*`
- 少量 benchmark / diagnostics 辅助代码

### 交付物

- text vs image 对照页面
- capability dump
- benchmark 输出
- scene summary
- artifact 目录结构

### 验收标准

- prototype 页面可运行
- capability / benchmark / diagnostics 可导出
- 产物目录边界清晰
- 能支持 go / no-go 判断

### 风险与回退

- 风险：把实验资产混入稳定文本回归目录
- 风险：原型页面与 `linechart demo` 语义混淆
- 回退：prototype 独立目录、自带 smoke test、不改现有 demo

### 测试命令

```bash
go test ./examples/charts_linechart_demo ./ui/e2e -run Linechart
```

### 评审重点

- prototype 是否真正服务决策
- artifact 目录是否干净
- 是否没有破坏现有文本 demo 和 e2e

## 10. 推荐节奏

### 第 1 波

- A 正式实现 `T1`
- B 起草 `scene.go` 和模型测试
- C 做 `Kitty presenter` scratch spike
- D 先写 `PR4` 测试草图和 render 分支检查清单
- E 准备 prototype 数据集和页面结构草稿

### 第 2 波

- `T1` 接口冻结后
- B 正式实现 `T2`
- C 正式实现 `T3`

### 第 3 波

- `T2 + T3` 合并后
- D 实现 `T4`

### 第 4 波

- `T4` 合并后
- E 实现 `T5`
- `T5` 合并后
- E 实现 `T6`

## 11. 一句话执行结论

本文件的核心执行策略是：

**不要把 `PR1` 硬拆给多人同时写，而是用 owner 锁 + 前置闸门，让基础层单点收敛、后续层按独占文件边界并行准备。**
