# ⚠️ 旧文档目录（已归档）

**状态**: **DEPRECATED** - 请参考新文档目录

---

## 📌 说明

本目录包含 Mint UI 早期架构设计（MVP 阶段）的文档。这些文档记录了开发过程中的探索、分析和临时设计方案。

**当前推荐架构**: **Store + Reducer**
**新文档位置**: `/docs/ui/store/`

---

## 🚀 迁移到新文档

### 架构文档

| 旧文档 | 新文档 |
|--------|--------|
| `INTENT_MANAGEMENT_PATTERNS.md` | [`/docs/ui/store/guides/DEVELOPMENT_GUIDE.md`](../../ui/store/guides/DEVELOPMENT_GUIDE.md) |
| `REFACTORING_SUMMARY.md` | [`/docs/ui/store/status/CURRENT_STATUS.md`](../../ui/store/status/CURRENT_STATUS.md) |
| `MVP_MIGRATION_GUIDE.md` | [`/docs/ui/store/guides/MIGRATION_GUIDE.md`](../../ui/store/guides/MIGRATION_GUIDE.md) |

### Intent 管理模式

**新文档路径**: [`/docs/ui/store/api/API_REFERENCE.md`](../../ui/store/api/API_REFERENCE.md#intent-系统)

**核心要点**:
- ✅ **Store + Reducer**: 推荐的生产架构
- ✅ **类型安全**: 编译期检查，无类型断言
- ✅ **单一数据源**: 所有状态在 Store 中
- ✅ **自动注册**: `RegisterToGlobal()` 自动处理

旧文档中的三种管理模式对比（组件级状态、全局状态、自定义 Intent）已经被 **Store + Reducer** 统一。

---

## 📂 新文档结构概览

```
docs/ui/store/
├── guides/           # 使用指南
│   ├── README.md                    # 快速开始
│   ├── DEVELOPMENT_GUIDE.md         # 完整开发指南 ⭐
│   ├── MIGRATION_GUIDE.md           # 迁移指南 ⭐
│   └── HOOK_USAGE_GUIDE.md          # Hooks 使用指南
├── api/              # API 参考
│   └── API_REFERENCE.md             # 完整 API 文档
├── optimization/     # 优化指南
│   └── FIELD_BINDING_OPTIMIZATION.md # 字段绑定优化 ⭐
├── status/           # 状态报告
│   ├── CURRENT_STATUS.md            # 架构状态 (93%)
│   └── MIGRATION_PROGRESS.md        # 迁移进度
└── migration/        # 迁移指南
    ├── INTENT_HANDLER_MIGRATION.md
    └── FORM_FIELDMAP_MIGRATION.md
```

---

## 📖 推荐阅读顺序

### 新手入门

1. 开始: [`/docs/ui/store/README.md`](../../ui/store/README.md)
2. 学习: [`/docs/ui/store/guides/DEVELOPMENT_GUIDE.md`](../../ui/store/guides/DEVELOPMENT_GUIDE.md)
3. 迁移: [`/docs/ui/store/guides/MIGRATION_GUIDE.md`](../../ui/store/guides/MIGRATION_GUIDE.md)

### 开发参考

- **API 参考**: [`/docs/ui/store/api/API_REFERENCE.md`](../../ui/store/api/API_REFERENCE.md)
- **优化指南**: [`/docs/ui/store/optimization/`](../../ui/store/optimization/)
- **当前状态**: [`/docs/ui/store/status/CURRENT_STATUS.md`](../../ui/store/status/CURRENT_STATUS.md)

---

## 🗂️ 本目录归档的文档

以下文档仅供参考历史开发过程，不应用于新项目：

| 文档 | 说明 | 新文档 |
|------|------|--------|
| `INTENT_MANAGEMENT_PATTERNS.md` | Intent 三种管理模式对比 | [`/docs/ui/store/guides/DEVELOPMENT_GUIDE.md`](../../ui/store/guides/DEVELOPMENT_GUIDE.md) |
| `REFACTORING_SUMMARY.md` | Store + Reducer 重构总结 | [`/docs/ui/store/reviews/IMPLEMENTATION_SUMMARY.md`](../../ui/store/reviews/IMPLEMENTATION_SUMMARY.md) |
| `MVP_MIGRATION_GUIDE.md` | MVP 模式迁移指南 | [`/docs/ui/store/guides/MIGRATION_GUIDE.md`](../../ui/store/guides/MIGRATION_GUIDE.md) |
| `CURRENT_ISSUES_AND_REFACTORING_PLAN.md` | 当前问题和重构计划 | [`/docs/ui/store/status/CURRENT_STATUS.md`](../../ui/store/status/CURRENT_STATUS.md) |
| `INTENT_DATA_FLOW_ANALYSIS.md` | Intent 数据流分析 | [`/docs/ui/store/guides/DEVELOPMENT_GUIDE.md`](../../ui/store/guides/DEVELOPMENT_GUIDE.md) |
| `INTENT_DATA_FLOW_REVIEW.md` | Intent 数据流审查 | [`/docs/ui/store/reviews/IMPLEMENTATION_REVIEW.md`](../../ui/store/reviews/IMPLEMENTATION_REVIEW.md) |
| `COMPONENT_INTENT_REVIEW.md` | 组件 Intent 检查报告 | `/docs/ui/store/reviews/` |
| `REFACTORING_COMPLETE.md` | 重构完成报告 | [`/docs/ui/store/status/MIGRATION_PROGRESS.md`](../../ui/store/status/MIGRATION_PROGRESS.md) |

---

## 🎯 关键迁移要点

### 1. 状态管理统一

**旧方式**（UseState + GlobalState）:
```go
// ❌ 多源状态，复杂 setter 管理
username, setUsername := ui.UseStateString("")
ctx.GlobalState["username-setter"] = setUsername
```

**新方式**（Store + Reducer）:
```go
// ✅ 单一数据源
type AppState struct { Username string }
appStore := store.NewStore(AppState{Username: ""})
state := appStore.Get()  // 读取状态
```

### 2. Intent 处理自动化

**旧方式**（手动注册 + 类型断言）:
```go
// ❌ 手动 WithInit + 类型断言
ui.WithInit(func() {
    ui.RegisterIntent(func(ctx *ActionContext, i intent.FieldChangeIntent) {
        if fn, ok := ctx.GetState("username-setter"); ok {
            if setter, ok := fn.(func(string)); ok {  // 类型断言!
                setter(i.Value)
            }
        }
    })
})
```

**新方式**（自动注册）:
```go
// ✅ 纯函数 Reducer + 自动注册
reducer := NewBuilder[AppState]()
reducer.On(FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
    fci := i.(FieldChangeIntent)
    s.Username = fci.Value  // 直接访问，无类型断言
    return s
})
reducer.RegisterToGlobal(appStore)
```

### 3. 字段绑定优化

**旧方式**（Switch-Case 硬编码）:
```go
// ❌ 手动 switch-case
switch i.Field {
case "username": s.Username = i.Value
case "email": s.Email = i.Value
// ... 硬编码所有字段
}
```

**新方式**（FieldMap 自动映射）:
```go
// ✅ 声明式字段映射
fieldMap := NewFieldMap[AppState]().
    Field(&s.Username, "username").
    Field(&s.Email, "email")

fieldMap.ApplyFieldChange(s, fieldChangeIntent)  // 自动处理
```

详细优化指南: [`/docs/ui/store/optimization/FIELD_BINDING_OPTIMIZATION.md`](../../ui/store/optimization/FIELD_BINDING_OPTIMIZATION.md)

---

## 📊 迁移进度

| 示例目录 | 状态 | 说明 |
|---------|------|------|
| `examples/counter/` | ✅ 已完成 | Store + Reducer 架构 |
| `examples/ui_demos/demo2_runtime_internals/inspector_demo/` | ✅ 已完成 | UseState → Store 迁移 |
| `examples/ui_demos/demo2_runtime_internals/inspector_overlay/` | ✅ 已完成 | UseState → Store 迁移 |
| `examples/ui_demos/demo2_runtime_internals/inspector_standalone/` | ✅ 已完成 | UseState → Store 迁移 |
| `examples/fiber_firsts/portal_demo/` | ✅ 已完成 | UseState → Store 迁移 |
| `examples/fiber_firsts/modal_demo/` | ✅ 已完成 | UseState → Store 迁移 |
| `examples/fiber_firsts/modal_positioning_demo/` | ✅ 已完成 | UseState → Store 迁移 |

剩余 UseState* 使用仅在测试文件中，保留用于 Hooks 机制验证。

详细进度: [`/docs/ui/store/status/MIGRATION_PROGRESS.md`](../../ui/store/status/MIGRATION_PROGRESS.md)

---

## 📚 参考资料

### 新文档核心推荐

- **架构概述**: [`/docs/ui/store/guides/README.md`](../../ui/store/guides/README.md)
- **开发指南**: [`/docs/ui/store/guides/DEVELOPMENT_GUIDE.md`](../../ui/store/guides/DEVELOPMENT_GUIDE.md) ⭐
- **迁移指南**: [`/docs/ui/store/guides/MIGRATION_GUIDE.md`](../../ui/store/guides/MIGRATION_GUIDE.md) ⭐
- **API 参考**: [`/docs/ui/store/api/API_REFERENCE.md`](../../ui/store/api/API_REFERENCE.md)
- **当前状态**: [`/docs/ui/store/status/CURRENT_STATUS.md`](../../ui/store/status/CURRENT_STATUS.md)
- **GlobalState 废弃**: [`/docs/ui/store/GLOBALSTATE_DEPRECATION.md`](../../ui/store/GLOBALSTATE_DEPRECATION.md)

### 代码示例

- **完整示例**: `examples/store_reducer_demo/`
- **迁移示例**: `examples/counter/`
- **Hooks 用法示例**: `examples/sandbox/demo/sandbox_compatibility_test.go`

---

## ⚙️ 维护说明

### 本目录的维护原则

1. **不再添加新文档** - 所有新架构文档应放在 `/docs/ui/store/`
2. **保留历史记录** - 本目录作为开发过程的历史档案保留
3. **更新映射关系** - 当旧文档内容被新文档替代时，更新本 README 的映射表

### 何时可以完全删除本目录？

以下情况满足时，可以考虑归档或删除本目录：
- ✅ 所有示例已迁移到 Store + Reducer
- ✅ UseState API 标记为 Deprecated 并提供弃用警告
- ✅ 用户社区完全采用新架构
- ⚠️ 当前：UseState 仍用于测试文件，保留作为 Hooks 机制验证

---

**文档迁移日期**: 2026-03-08
**维护团队**: Mint UI 团队
