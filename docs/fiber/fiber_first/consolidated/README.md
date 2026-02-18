# Fiber-First 文档整理说明

## 整理概述

本目录包含从原始 Fiber-first 文档整理后的统一架构文档。

### 原始文档列表

以下13个文档已被整合：

1. `fiber_first.md` - 整体架构和数据流
2. `fiber_action.md` - Action系统整合
3. `fiber_action_payload.md` - 参数传递模型
4. `fiber_button.md` - Button组件设计
5. `fiber_checklist.md` - 统一Action Runtime重构
6. `fiber_confict.md` - 架构冲突解决
7. `fiber_event.md` - 事件架构设计
8. `fiber_intent.md` - 声明式Intent API
9. `fiber_paint.md` - Paint架构
10. `FIBER_PAINT_ARCHITECTURE.md` - Paint架构详细设计
11. `fiber_vs_vnode.md` - Fiber与VNode对比
12. `fiber优化.md` - 性能优化指南
13. `REFACTOR_PLAN.md` - 重构计划

### 整理后文档

1. **FIBER_FIRST_ARCHITECTURE.md** - 统一架构文档
   - 整合所有核心架构概念
   - 包含完整的数据流图
   - 详细的实施指南

2. **FIBER_FIRST_QUICK_REFERENCE.md** - 快速参考指南
   - 核心概念速查
   - 常用代码模式
   - 调试技巧

## 主要变更

### 术语统一

- `computeBox` → `paint.PaintableBox`
- 所有布局结果树节点统称为 `paint.PaintableBox`
- Instance/ComponentInstance 统一使用 `paint.PaintableBox` 接口

### 结构重组

原始文档按主题分散，整理后按架构层次组织：

1. **架构概览** - 三层职责分离
2. **核心角色** - VNode/Fiber/Instance 对比
3. **数据流图** - 完整的数据流和事件流
4. **实施指南** - 详细的迁移步骤和检查清单
5. **代码示例** - 关键代码模式

### 内容优化

1. **去重** - 合并重复的架构描述
2. **补充** - 添加缺失的代码示例
3. **标准化** - 统一术语和格式
4. **索引** - 添加目录和快速参考

## 使用建议

### 新手入门

1. 先阅读 `FIBER_FIRST_ARCHITECTURE.md` 了解整体架构
2. 使用 `FIBER_FIRST_QUICK_REFERENCE.md` 作为日常参考
3. 遇到具体问题时查阅原始文档

### 日常开发

- 使用 `FIBER_FIRST_QUICK_REFERENCE.md` 查找代码模式
- 遇到架构问题时参考 `FIBER_FIRST_ARCHITECTURE.md`
- 实施重构时参考 Checklist 部分

### 架构演进

- 重构计划在 `FIBER_FIRST_ARCHITECTURE.md` 的"重构计划"章节
- 性能优化建议在快速参考指南中
- 保留原始文档供深入理解设计决策

## 文档维护

### 更新原则

1. **主文档优先** - 重大更新先更新架构文档
2. **保持同步** - 代码变更后更新快速参考
3. **向后兼容** - 保留原始文档供参考

### 贡献指南

1. 新增内容先讨论位置（架构 vs 快速参考）
2. 代码示例必须可运行
3. 保持术语一致性

## 注意事项

### computeBox → paint.PaintableBox 替换

所有原始文档中的 `computeBox` 或 `ComputedBox` 已替换为 `paint.PaintableBox`，包括：

- 数据流图中的节点名称
- 代码示例中的类型引用
- 接口定义
- 文字描述

### 架构演进

Fiber-first 架构仍在演进中，文档可能存在：

- 未完成的TODO
- 计划中的功能
- 待优化的部分

请参考最新的代码实现。

---

**最后更新**: 2024年
**维护者**: Fiber-first 架构团队
