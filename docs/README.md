# Mint TUI Framework 文档

本文档目录包含 Mint 终端 UI 框架的设计文档、实现指南和问题分析。

## 目录结构

### 📚 API 文档 (`api/`)
组件和 Hooks 的 API 参考
- `component.md` - 组件 API 文档
- `hooks.md` - Hooks API 文档
- `memory-safety.md` - 内存安全规范

### 🏗️ 架构设计 (`architecture/`)
系统架构设计和问题分析
- `ARCHITECTURE_ISSUES.md` - 架构问题汇总

### 🐛 调试指南 (`debugging/`)
调试工具和环境配置
- `DEBUG_ENVIRONMENT_VARIABLES.md` - 调试环境变量说明

### 📝 示例 (`examples/`)
示例应用优化和改进记录
- `demo2_optimization_report.md` - demo2 优化报告

### ✨ 功能特性 (`features/`)
功能特性实现和分析
- `mouse-text-selection.md` - 鼠标文本选择功能
- `focus/` - 焦点管理相关
  - `mouse_click_focus_implementation.md` - 鼠标点击焦点切换实现
  - `mouse_click_focus_issue.md` - 鼠标焦点问题分析
  - `tab_key_focus_implementation.md` - Tab 键焦点切换实现

### 📖 开发指南 (`guide/`)
组件开发和迁移指南
- `component-development-guide.md` - 组件开发指南
- `migration-guide.md` - 迁移指南

### 📋 问题追踪 (`issue/`)
历史问题记录和修复报告
- 按日期命名的问题报告
- 实现计划和测试报告

### 📐 布局系统 (`layout/`)
布局引擎设计和实现
- `flex_layout.md` - Flex 布局说明
- `layer_system_guide.md` - 图层系统指南
- `stretch_layout.md` - Stretch 布局
- `issue/` - 布局问题修复记录
  - `layout_engine_fix_summary.md` - 布局引擎修复总结
  - `LAYOUT_FIX_PLAN.md` - 布局修复计划
  - `modal_centering_fix.md` - Modal 居中修复
- 渲染流程和架构审查文档

### 📋 计划文档 (`plan/`)
开发计划和路线图
- `REFACTOR_PLAN.md` - 重构计划
- `FIBER_RECONCILER_MIGRATION.md` - Fiber 协调器迁移
- `COMPONENT_MIGRATION_GUIDE.md` - 组件迁移指南
- 各阶段实现报告

### 🧪 测试沙箱 (`sandbox/`)
沙箱测试环境文档
- `API_REFERENCE.md` - API 参考
- `APP_LIFECYCLE_AND_SANDBOX.md` - 应用生命周期
- `SANDBOX_DESIGN.md` - 沙箱设计
- `fiber_debug/` - Fiber 调试相关文档

### 🎨 样式系统 (`style/`)
样式 API 和设计系统
- `style_api_refactor_proposal.md` - 样式 API 重构提案
- `migration_examples.md` - 迁移示例
- `demo1_migration_report.md` - demo1 迁移报告

### 📊 任务管理 (`task/`)
开发路线图和任务规划
- `roadmap.md` - 项目路线图

### 🧪 测试文档 (`testing/`)
测试工具和设计
- `TESTABLE_INPUT_DESIGN.md` - 可测试输入设计
- `TESTING_TOOL.md` - 测试工具文档

### 🎨 主题系统 (`theme/`)
主题设计和实现
- `theme_system_guide.md` - 主题系统指南
- `theme_rendering_flow.md` - 主题渲染流程
- `design/` - 主题设计规范
  - `theme.md` - 主题系统设计
  - `ant_design.md` - Ant Design 适配
- Header 间隙问题分析

## 文档分类说明

### 按用途分类

**学习使用**：
- `guide/` - 开发指南
- `api/` - API 参考
- `examples/` - 示例代码

**理解实现**：
- `architecture/` - 架构设计
- `layout/` - 布局系统
- `features/` - 功能实现
- `theme/` - 主题系统
- `style/` - 样式系统

**问题排查**：
- `issue/` - 历史问题
- `debugging/` - 调试指南
- `testing/` - 测试文档

**项目管理**：
- `plan/` - 开发计划
- `task/` - 任务追踪
- `sandbox/` - 沙箱测试

### 按时间线

- **设计阶段**: `architecture/`, `plan/`, `theme/design/`
- **实现阶段**: `features/`, `layout/`, `style/`
- **测试阶段**: `testing/`, `sandbox/`, `issue/`
- **维护阶段**: `debugging/`, `api/`, `guide/`

## 贡献指南

添加新文档时，请根据文档类型选择合适的目录：

1. **功能实现** → `features/功能名/`
2. **问题分析** → `issue/YYYY-MM-DD-问题描述.md`
3. **设计方案** → `plan/` 或对应的子系统目录
4. **API 文档** → `api/`
5. **使用指南** → `guide/`
6. **测试相关** → `testing/` 或 `sandbox/`

## 文档命名规范

- 使用英文文件名，用下划线分隔
- 问题报告: `YYYY-MM-DD-问题描述.md`
- 实现报告: `feature_implementation.md`
- 设计文档: `feature_design.md`
- 优化报告: `demo_optimization_report.md`

---

## 📖 English Documentation

The following directories contain English documentation for international users:

### Components (`components/`)
Component library documentation
- ScrollView, VirtualScroll, Tabs, TreeView
- Component usage and API reference

### Key Handling Guide (`guide/key-handling/`)
Keyboard input system documentation
- [Complete Guide](guide/key-handling/KEY_HANDLING_COMPLETE_GUIDE.md) - Comprehensive key handling reference
- [Alt+K Debug Guide](guide/key-handling/ALT_K_DEBUG_GUIDE.md) - Debugging Alt+K issues
- [Ctrl Key Fix Summary](guide/key-handling/CTRL_KEY_FIX_SUMMARY.md) - Ctrl detection fixes
- [Case Preservation](guide/key-handling/CTRL_SHIFT_CASE_PRESERVATION.md) - Shift handling

### Architecture (`architecture/`)
System architecture and design
- [Layer System Architecture](architecture/LAYER_SYSTEM_ARCHITECTURE.md) - Layer system overview
- [Implementation Summary](architecture/LAYER_SYSTEM_IMPLEMENTATION_SUMMARY.md) - Implementation details
- [Agents](architecture/AGENTS.md) - Agent system
- [Two Rendering Systems](architecture/TWO_RENDERING_SYSTEMS_EXPLAINED.md) - Rendering approaches

### Inspector (`examples/inspector/`)
UI Inspector tool documentation
- [Quick Reference](examples/inspector/INSPECTOR_QUICK_REFERENCE.md) - All shortcuts and features
- [All Fixes Summary](examples/inspector/INSPECTOR_ALL_FIXES_SUMMARY.md) - Complete fix history
- [Key Debug Feature](examples/inspector/INSPECTOR_KEY_DEBUG.md) - Debug key presses
- [Modal Behavior](examples/inspector/INSPECTOR_MODAL_BEHAVIOR_FIXED.md) - Modal behavior

### Testing (`testing/`)
Test documentation and results
- [BoxModel Test Results](testing/BOXMODEL_TEST_RESULTS.md) - Test results

### Sandbox (`sandbox/`)
Testing sandbox documentation
- [Event Integration](sandbox/SANDBOX_EVENT_INTEGRATION_ANALYSIS.md) - Event system integration

