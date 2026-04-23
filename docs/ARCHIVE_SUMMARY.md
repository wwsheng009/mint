# 文档归档清理总结

**日期**: 2026-03-10（更新：2026-03-13）

## 2026-03-13 追加归档（阶段/计划/总结类）

- 归档 `docsArchive/plans/**` 全部文档 → `docsArchive/plans/**`（含 `action/`、`event/`、`fiber/`、`inspector/`、`layer/` 等）
- 归档命名包含 `PHASE/REPORT/SUMMARY/PLAN/ROADMAP/TODO/TASKS/COMPLETION` 的文档 → `docsArchive/` 对应路径
- 覆盖目录示例：`docs/layout/layout/20260215/`、`docs/render/paint/optimized/`、`docs/sandbox/fiber_debug/`
- 本轮新增归档 55 个文件

## 已归档的文档

以下文档已被移动到 `docsArchive/` 目录，这些文档是过程性的、临时的或已修复问题的记录。

### 1. Inspector 相关文档（约 80+ 文件）

#### `docsArchive/inspector/`
- `integration/` - Inspector 集成修复记录（约 20 个文档）
- `investigation/` - 问题调查报告（约 10 个文档）
- `architecture/` - 架构问题分析（约 4 个文档）
- 主目录修复文档：
  - `FIX_SUMMARY.md`
  - `KEY_DETECTION_FIX.md`
  - `KEY_RESPONSE_ISSUE.md`
  - `RENDERING_ISSUE.md`
  - `elements_panel_path_fix.md`
  - `DEBUG_INSPECTOR_CLICK.md`
  - 等 10+ 个修复文档

#### `docsArchive/examples/inspector/`
- `DEMO2_INSPECTOR_FIX.md`
- `INSPECTOR_ALL_FIXES_SUMMARY.md`
- `INSPECTOR_ENHANCEMENTS_SUMMARY.md`
- 以及约 25 个其他修复/调查文档

#### `docsArchive/inspector/implementation/`
- `INSPECTOR_HARDCODED_BORDER_FIX.md`
- `INSPECTOR_PATH_INDEX_FIX.md`
- `INSPECTOR_POINTER_ID_FIX.md`
- `INSPECTOR_POSITION_FIX.md`
- 等 10+ 个修复文档
- Phase 报告：`PHASE2_REPORT.md` ~ `PHASE7_REPORT.md`

### 2. 问题修复和调试文档（约 60+ 文件）

#### `docsArchive/issue/`
- 整个 `event_refactor/` 目录
- 日期命名的问题报告（约 15 个）
- 架构复杂度分析等

#### `docsArchive/fixes/`
- `CLONEFIBER_SIBLING_BUG.md`
- `FIBER_COUNTER_BUTTON_DUPLICATION_ANALYSIS.md`
- 等所有修复文档

#### `docsArchive/report/`
- `button_paint_analysis.md`
- `demo2_button_layout_analysis.md`
- `fillwidth_final_solution.md`
- 等 10+ 个分析报告

### 3. 阶段性完成报告（约 30+ 文件）

- `docsArchive/plan/event/PHASE_1-6_COMPLETION.md` ~ `PHASE_6_COMPLETION.md`
- `docsArchive/layout/phase1_completion.md` 等
- 以及多个子系统的完成报告

### 4. 计划和设计文档（约 40+ 文件）

#### `docsArchive/plan/`
- `REFACTOR_PLAN.md`
- `REFACTOR_TODO.md`
- `box_model_interface_design.md`
- `button_text_alignment_design.md`
- `layout_optimization_plan.md`
- `two_phase_rendering_refactor.md`
- 等各种临时计划和设计方案

#### `docsArchive/layout/`
- 方案文档（方案A、方案B）
- refactoring 计划
- 实现计划
- 等分析文档

### 5. 其他临时文档（约 30+ 文件）

- `docsArchive/fiber/` - Fiber 架构分析和 diff 文档
- `docsArchive/performance/` - 性能优化修复报告
- `docsArchive/debug/` - 调试文档
- `docsArchive/render/tui/` - TUI buffer 修复文档
- `docsArchive/theme/` - 主题 gap 修复文档
- `docsArchive/state/MIGRATION.md`
- 等等

## 保留的核心文档

以下文档已保留在 `docs/` 目录中，它们是：

1. **索引和指导文档**：
   - `docs/README.md`
   - 各子目录的 `README.md`

2. **用户指南**：
   - `docs/guide/key-handling/` - 按键处理完整指南
   - `docs/guide/component-development-guide.md`
   - `docs/guide/migration-guide.md`

3. **API 参考**：
   - `docs/api/` - API 参考文档

4. **调试工具**：
   - `docs/debug/` - Debug 工具使用指南
   - `docs/debug/quick_start.md`
   - `docs/debug/environment_variables.md`

5. **架构文档**（核心的，非临时分析）：
   - `docs/architecture/README.md`
   - `docsArchive/architecture/mvp/MVP_MIGRATION_GUIDE.md`
   - `docsArchive/architecture/mvp/COMPONENT_INTENT_REVIEW.md`

6. **组件文档**：
   - `docs/components/README.md`
   - `docsArchive/components/COMPONENTS_SUMMARY.md`

7. **布局系统**（用户导向的）：
   - `docs/layout/README.md`
   - `docs/layout/user_guide/` - 用户指南
   - `docs/layout/core_concepts/` - 核心概念
   - `docs/layout/visualizer_usage_guide.md`
   - `docs/layout/constraint_tracer_guide.md`

8. **Inspector 使用指南**：
   - `docs/inspector/README.md`
   - `docs/inspector/QUICK_START.md`
   - `docs/inspector/FINAL_GUIDE.md`

9. **主题系统**：
   - `docs/theme/theme_system_guide.md`
   - `docs/theme/theme_system_guide.md`
   - `docsArchive/theme/ant_design_summary.md`

10. **存储系统**：
    - `docs/ui/store/README.md`
    - `docs/ui/store/guides/` - 开发指南

11. **其他实用文档**：
    - `docs/layer/` - Layer 系统架构
    - `docs/platform/` - 平台相关
    - `docs/state/README.md`
    - 等

## 归档统计

| 类别 | 归档文件数 | 保留文件数 |
|------|-----------|-----------|
| Inspector 相关 | ~80 | ~10 |
| 问题修复/调试 | ~60 | 0 |
| 阶段报告 | ~30 | 0 |
| 计划设计 | ~40 | ~20 |
| 性能/优化 | ~10 | ~2 |
| 其他临时 | ~30 | ~10 |
| **总计** | **~250** | **~50** |

## 清理效果

- **文档数量减少**：约 250 个临时文档被归档
- **结构清晰**：`docs/` 目录现在只包含用户和开发者需要参考的核心文档
- **易于维护**：归档的文档可以随时查阅历史，但不会干扰日常使用

## 建议后续操作

1. **定期清理**：建议每季度检查 `docsArchive/`，将不再需要的历史文档彻底删除
2. **合并文档**：可以考虑将同类别的完成报告合并为一份最终摘要
3. **版本控制**：考虑将 `docsArchive/` 添加到 `.gitignore`，或建立独立的归档仓库

## 注意事项

- 所有移动的文件仍然可访问，位于 `docsArchive/` 目录
- 没有文件被永久删除，只是移动以清理 `docs/` 目录
- 如需查找已归档的文档，可以使用 `grep` 在 `docsArchive/` 中搜索
