# 渲染系统文档

本目录包含 Mint 渲染系统的设计文档、实现指南和问题分析。

## 目录结构

### 📐 计划文档 (`plan/`)

渲染系统开发和重构计划

| 文档 | 说明 | 状态 |
|------|------|------|
| `RENDER_ENGINE_DECOUPLING.md` | 渲染引擎解耦方案 | 设计 |
| `SIMPLIFIED_PIPELINE_V2.md` | 简化渲染管线设计 | 设计 |
| `SIMPLIFY_RENDER_PIPELINE.md` | 简化渲染管线 | 设计 |
| `RENDER_PIPELINE_SIMPLIFICATION_COMPLETED.md` | 渲染管线简化完成 (2026-03, 已归档) | ✅ 完成 |

### 🎨 Paint 系统 (`paint/`)

Paint 相关的文档

### 🔧 Hook 系统 (`hook/`)

渲染 Hook 文档

### 🐛 修复记录 (`fixes/`)

渲染相关的问题修复记录

### 🎮 TUI 相关 (`tui/`)

TUI 特定的渲染文档

### 🖼️ Pixel 渲染 (`pixel/`)

面向图片像素图表与混合文本/图像渲染的方案文档

| 文档 | 说明 | 状态 |
|------|------|------|
| `README.md` | Pixel 渲染目录说明 | 设计 |
| `DELIVERY_HANDOFF_GUIDE.md` | Pixel 渲染方案交付与移交指南 | 设计 |
| `DECISION_LOG.md` | Pixel 渲染 Phase 1 决策日志 | 设计 |
| `PIXEL_CHART_RENDERING_ARCHITECTURE.md` | 图片像素图表渲染架构方案 | 设计 |
| `CURRENT_SYSTEM_IMPACT_AND_FEASIBILITY_REVIEW.md` | 结合当前实现的冲击审查与可行性验证 | 设计 |
| `PERFORMANCE_IMPACT_AND_OPTIMIZATION_PLAN.md` | 性能影响、优化策略与验证指标 | 设计 |
| `LINECHART_IMAGE_PROTOTYPE_PLAN.md` | `linechart` image prototype PoC 计划 | 设计 |
| `PHASE1_TASK_BREAKDOWN.md` | Phase 1 分组拆解、验收和风险控制 | 设计 |
| `GROUP_A_GRAPHICS_CAPABILITY_SPEC.md` | 图形能力层技术规格 | 设计 |
| `GROUP_BC_SCENE_AND_APP_RENDER_INTEGRATION_SPEC.md` | Scene 与 `App.render()` 接入规格 | 设计 |
| `GROUP_D_LINECHART_IMAGE_RENDERER_SPEC.md` | `linechart` image renderer 技术规格 | 设计 |
| `GROUP_E_PROTOTYPE_BENCHMARK_AND_DIAGNOSTICS_SPEC.md` | Prototype、benchmark 与 diagnostics 规格 | 设计 |
| `RUNTIME_PLATFORM_GRAPHICS_API_SKETCH.md` | `runtime/platform/graphics` API 草案与伪代码清单 | 设计 |
| `RUNTIME_PAINT_SCENE_API_SKETCH.md` | `runtime/paint/scene` API 草案与最小数据模型 | 设计 |
| `FRAMEWORK_APP_RENDER_IMAGE_FLOW_SPEC.md` | `framework.App.render()` 图像帧接入流程规格 | 设计 |
| `EXAMPLES_LINECHART_IMAGE_PROTOTYPE_LAYOUT_SPEC.md` | `examples/charts_linechart_image_prototype` 目录与页面布局规格 | 设计 |
| `PROTOTYPE_BENCHMARK_AND_ARTIFACT_LAYOUT_SPEC.md` | Prototype benchmark 与 diagnostics 产物布局规格 | 设计 |
| `IMPLEMENTATION_SEQUENCE_AND_PR_PLAN.md` | Pixel 渲染实施顺序与 PR 切分计划 | 设计 |

---

## 最新变更 (2026-03)

### ✅ 渲染管线简化已完成

详见 [`/docsArchive/RENDER_PIPELINE_SIMPLIFICATION_COMPLETED.md`](/docsArchive/RENDER_PIPELINE_SIMPLIFICATION_COMPLETED.md)

#### 主要变更

1. **Scheduler 接口解耦**
   - 移除 `framework.App` 依赖
   - 通过 `Scheduler` 接口最小化依赖
   - `SetApp()` 现在是可选方法

2. **Fiber-first 默认启用**
   - 移除 `RenderModeBoth`
   - 移除环境变量 `MINT_FIBER_FIRST` 检查
   - `RenderModeFiberFirst` 是唯一活跃模式

3. **Portal-aware Layout 默认启用**
   - Portal layout 默认为启用状态
   - 可通过 `MINT_PORTAL_LAYOUT=0` 禁用

4. **移除 Legacy Paint Fallback**
   - 移除所有 8 处 `legacyPaint()` fallback 调用
   - 单一路径渲染，更易调试
   - `legacyPaint()` 标记为 DEPRECATED

5. **API 简化**
   ```go
   // 之前的 API
   node := render.NewDeclarativeNodeFromFuncWithFiber(app, buildUI)

   // 新的 API
   node := render.NewDeclarativeNodeFromFuncWithFiber(buildUI)
   node.SetApp(app)  // 可选，交互式应用需要
   ```

---

## 历史文档

### 旧版本计划

以下文档是之前的设计方案，已在新版本中实现或有替代方案：

- `SIMPLIFY_RENDER_PIPELINE.md` - 早期的简化方案
- `RENDER_ENGINE_DECOUPLING.md` - 解耦设计（已实现)
- `SIMPLIFIED_PIPELINE_V2.md` - 第二版简化设计

### 未来计划

以下文档是未来的发展方向，当前版本尚未实现：

- 详见 `task/` 和 `plan REFACTOR_PLAN.md`

---

## 相关文档

| 文档 | 说明 |
|------|------|
| `../layout/layout_system_guide.md` | 布局系统指南 |
| `../architecture/LAYER_SYSTEM_ARCHITECTURE.md` | 图层系统架构 |
| `../fiber/FIBER_RECONCILER_MIGRATION.md` | Fiber 协调器迁移 |

---

## 贡献

添加新的渲染相关文档时，请遵循以下目录结构：

- 设计文档 → `plan/`
- 问题修复 → `fixes/`
- 实现细节 → 对应子系统目录

命名规范：
- 设计文档: `feature_design.md`
- 实现文档: `feature_implementation.md`
- 总结文档: `feature_completion_summary.md`
