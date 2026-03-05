# Store + Reducer 架构文档

**版本**: v0.11  
**最后更新**: 2026-03-05

---

## 📂 目录结构

本文档目录按照文档类型进行了分类整理，便于查找和管理。

```
store/
├── guides/           # 📖 使用指南 - 如何使用 Store + Reducer
├── api/              # 📚 API 参考 - 完整的 API 文档
├── optimization/     # ⚡ 优化指南 - 性能优化相关
├── migration/        # 🔄 迁移指南 - 从旧 API 迁移
├── status/           # 📊 状态报告 - 当前进度和状态评估
├── plans/            # 📋 计划文档 - 废弃计划、重构计划
├── reviews/          # ✅ 审查报告 - 实现审查和总结
├── fixes/            # 🔧 问题修复 - Bug 修复文档
├── issues/           # ❌ 问题文档 - 已知问题
├── features/         # 🎯 特性文档 - 类型安全、日志等特性
└── design/           # 🎨 设计文档 - 原始设计方案
```

---

## 📖 快速开始

### 新手入门

1. **阅读指南**: 从 `guides/README.md` 开始了解 Store + Reducer 架构
2. **开发指南**: 参考 `guides/DEVELOPMENT_GUIDE.md` 学习如何构建应用
3. **API 参考**: 查看 `api/API_REFERENCE.md` 了解完整 API
4. **迁移指南**: 如果有旧代码，参考 `guides/MIGRATION_GUIDE.md` 进行迁移

### 进阶学习

- **字段绑定优化**: `optimization/FIELD_BINDING_OPTIMIZATION.md` 
- **类型安全 Intent**: `features/TYPE_SAFE_INTENT.md`
- **日志和错误处理**: `features/LOGGING_AND_ERROR_HANDLING_GUIDE.md`

### 状态和进度

- **当前状态**: `status/CURRENT_STATUS.md` - 架构完善度评估
- **迁移进度**: `status/MIGRATION_PROGRESS.md` - 示例迁移进度

---

## 📖 分类说明

### guides/ 📖 使用指南

使用指南类文档，帮助开发者理解和使用 Store + Reducer 架构。

| 文档 | 说明 |
|------|------|
| `README.md` | Store + Reducer 架构概述和快速开始 |
| `DEVELOPMENT_GUIDE.md` | 完整的开发指南，包含最佳实践 |
| `STORE_REDUCER_GUIDE.md` | Store + Reducer 模式详细指南 |
| `HOOK_USAGE_GUIDE.md` | Hooks 使用指南和最佳实践 |
| `MIGRATION_GUIDE.md` | 从 UseState 迁移到 Store + Reducer |
| `RUNAPP_GUIDE.md` | AppRuntime 和 RunApp 使用指南 |

---

### api/ 📚 API 参考

完整的 API 文档，包含所有公开接口的详细说明。

| 文档 | 说明 |
|------|------|
| `API_REFERENCE.md` | Store、Reducer、FieldBinding、Intent 等完整 API |

---

### optimization/ ⚡ 优化指南

性能优化相关文档，帮助提升应用性能。

| 文档 | 说明 |
|------|------|
| `FIELD_BINDING_OPTIMIZATION.md` | **推荐**: 使用 FieldMap 消除 switch-case 硬编码 |
| `SYSTEM_ANALYSIS_OPTIMIZATION.md` | Store + Intent 系统分析与优化方案 |
| `APPVIEW_TYPE_OPTIMIZATION.md` | AppView 类型优化 |
| `APPVIEW_TYPE_SUMMARY.md` | AppView 类型优化总结 |

---

### migration/ 🔄 迁移指南

从旧 API 或其他模式迁移到 Store + Reducer 的指南。

| 文档 | 说明 |
|------|------|
| `INTENT_HANDLER_MIGRATION.md` | Intent Handler 迁移指南 |
| `FORM_FIELDMAP_MIGRATION.md` | 表单字段映射迁移指南 |
| `TYPESAFE_FORM_MIGRATION_SUMMARY.md` | 类型安全表单迁移总结 |

---

### status/ 📊 状态报告

当前架构的完善度评估和迁移进度。

| 文档 | 说明 |
|------|------|
| `CURRENT_STATUS.md` | Store + Reducer 完善性评估报告 (93%) |
| `MIGRATION_PROGRESS.md` | Store + Reducer 迁移进度报告 |
| `STATUS_UPDATE.md` | 状态更新报告 |

---

### plans/ 📋 计划文档

废弃计划和重构计划。

| 文档 | 说明 |
|------|------|
| `DEPRECATION_PLAN.md` | UseState 废弃计划 |
| `REFACTOR_PLAN.md` | Mint UI 架构重构完整方案 |

---

### reviews/ ✅ 审查报告

实现审查和总结。

| 文档 | 说明 |
|------|------|
| `IMPLEMENTATION_SUMMARY.md` | 实现总结 |
| `IMPLEMENTATION_REVIEW.md` | 实现审查 |

---

### fixes/ 🔧 问题修复

已知问题的修复文档。

| 文档 | 说明 |
|------|------|
| `TIMETRAVEL_FIX.md` | 时间旅行 Undo/Redo 修复 |
| `RUN_VS_RUNAPP.md` | Run vs RunApp 问题 |

---

### issues/ ❌ 问题文档

已知问题文档。

| 文档 | 说明 |
|------|------|
| `FMT_PRINT_ISSUE.md` | Fmt.Print 问题 |

---

### features/ 🎯 特性文档

特定功能的详细文档。

| 文档 | 说明 |
|------|------|
| `TYPE_SAFE_INTENT.md` | 类型安全 Intent DSL |
| `LOGGING_AND_ERROR_HANDLING_GUIDE.md` | 日志和错误处理指南 |

---

### design/ 🎨 设计文档

原始设计方案和概念文档。

| 文档 | 说明 |
|------|------|
| `store.md` | Store 原始设计文档 |
| `mini_demo.md` | Mini 演示文档 |

---

## 🎯 推荐阅读路径

### 路径 1: 新手入门 (完整学习曲线)

1. `guides/README.md` - 了解概述
2. `guides/DEVELOPMENT_GUIDE.md` - 学习开发
3. `api/API_REFERENCE.md` - 查阅 API
4. `optimization/FIELD_BINDING_OPTIMIZATION.md` - 学习优化

### 路径 2: 从 UseState 迁移

1. `guides/MIGRATION_GUIDE.md` - 迁移指南
2. `status/MIGRATION_PROGRESS.md` - 查看进度
3. `migration/INTENT_HANDLER_MIGRATION.md` - Intent Handler 迁移

### 路径 3: 深入优化

1. `optimization/FIELD_BINDING_OPTIMIZATION.md` - 字段绑定优化
2. `features/TYPE_SAFE_INTENT.md` - 类型安全 Intent
3. `status/CURRENT_STATUS.md` - 理解当前状态

---

## 📊 当前状态

- **架构完善度**: 93% ✅
- **核心功能**: Store[T] ✅ (100%), Reducer[T] ✅ (100%), AppRuntime[T] ⚠️ (90%)
- **文档完整度**: 100% ✅
- **示例迁移**: 核心示例已完成 ✅

详细状态参考: `status/CURRENT_STATUS.md`

---

## 🔗 相关资源

- **Mint UI 主仓库**: [github.com/wwsheng009/mint](https://github.com/wwsheng009/mint)
- **示例代码**: `examples/store_reducer_demo/`

---

**文档维护**: Mint UI 团队  
**最后更新**: 2026-03-05  
**版本**: v0.11
