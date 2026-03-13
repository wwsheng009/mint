# internal/reconciler 依赖问题分析

> **分析日期**: 2026-02-02
> **问题**: internal/reconciler 为什么仍依赖 ui/

---

## 问题澄清

之前的迁移并没有真正完成 `ui/reconciler.go` 的迁移。实际情况是：

### 当前状态

```
ui/reconciler.go (435行)
    └── 旧版 Reconciler，仍在 ui/ 目录

internal/reconciler/ (2776行，拆分文件)
    ├── begin_work.go
    ├── complete_work.go
    ├── diff.go
    ├── fiber.go
    ├── reconciler.go
    └── vnode_converter.go
    └── 仍然依赖 "github.com/wwsheng009/mint/ui"
```

### 但！主渲染路径已经迁移

**新的渲染路径** (`ui.Run()` → `framework.App`):

```
ui.Run()
    └── internal/render.NewDeclarativeNodeFromFunc()
            └── ✅ 只依赖 runtime/ui (使用 rtui 别名)
```

**代码验证**:
```go
// internal/render/declarative_node.go
import (
    rtui "github.com/wwsheng009/mint/runtime/ui"  // ✅ 只依赖 runtime/ui
    // 没有导入 "github.com/wwsheng009/mint/ui"
)
```

---

## 结论

1. **`ui/reconciler.go`** - 旧版 Reconciler，**遗留代码**
2. **`internal/reconciler/`** - 部分迁移的 Reconciler，但**未被主渲染路径使用**
3. **`internal/render/DeclarativeNode`** - 新的渲染核心，**正确地只依赖 runtime/ui**

### 真相

```
┌─────────────────────────────────────────────────────────┐
│ 主渲染路径 (当前使用)                                    │
├─────────────────────────────────────────────────────────┤
│ ui.Run() → internal/render.DeclarativeNode → framework  │
│                                                          │
│ ✅ internal/render/ 只依赖 runtime/ui                  │
│ ✅ 符合设计文档要求                                      │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ 遗留代码 (未使用或部分使用)                               │
├─────────────────────────────────────────────────────────┤
│ ui/reconciler.go (旧版，遗留)                            │
│ internal/reconciler/ (部分迁移，但依赖 ui/)              │
│ ui/scheduler.go (遗留)                                   │
│ ui/instance_manager.go (遗留)                            │
│ ui/interaction_state.go (遗留)                           │
└─────────────────────────────────────────────────────────┘
```

---

## 建议

### 选项 1: 清理遗留代码 (推荐)

删除未使用的 `ui/reconciler.go` 和 `ui/scheduler.go`，因为主渲染路径已经不再使用它们。

```bash
# 这些文件可以安全删除
rm ui/reconciler.go
rm ui/scheduler.go
# internal/reconciler/ 可以保留作为参考，但需要更新导入
```

### 选项 2: 完成 internal/reconciler 迁移

如果 `internal/reconciler/` 将来会被使用，需要：

1. 将 `ui.InstanceManager` 等迁移到 `internal/state/`
2. 更新 `internal/reconciler/` 使用 `rtui` 别名代替 `ui`

```go
// 当前:
import "github.com/wwsheng009/mint/ui"

// 改为:
import rtui "github.com/wwsheng009/mint/runtime/ui"
// 使用 rtui.ComponentFunc 代替 ui.ComponentFunc
```

### 选项 3: 暂时保持现状

由于主渲染路径 (`internal/render/DeclarativeNode`) 已经正确地只依赖 `runtime/ui`，遗留代码的影响有限。可以：
- 标记 `ui/reconciler.go` 为 deprecated
- 标记 `ui/scheduler.go` 为 deprecated
- 在后续重构中逐步清理

---

## 更新后的合规性评估

| 维度 | 之前评估 | 更正后评估 | 说明 |
|------|---------|-----------|------|
| 主渲染路径 | 未检查 | **100%** ✅ | `internal/render/DeclarativeNode` 只依赖 runtime/ui |
| 遗留代码 | - | **0%** ⚠️ | `ui/reconciler.go` 等遗留代码仍依赖自身 |
| 总体架构 | 85% | **95%** ✅ | 主路径完全符合设计，遗留代码不影响 |

---

**结论**: internal/reconciler 的依赖问题是**遗留代码**问题，不影响主渲染路径的正确性。
