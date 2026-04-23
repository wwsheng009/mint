# ui/ 目录清理分析报告

**日期**: 2026-02-02

---

## 1. 文件分类

### 1.1 保留文件 (核心功能)

| 文件 | 状态 | 说明 |
|------|------|------|
| `vnode.go` | ✅ 保留 | VNode 接口重导出 |
| `element.go` | ✅ 保留 | ElementVNode 重导出 |
| `component.go` | ✅ 保留 | ComponentVNode 重导出 |
| `fragment.go` | ✅ 保留 | FragmentVNode 重导出 |
| `fiber.go` | ✅ 保留 | Fiber 重导出 |
| `instance.go` | ✅ 保留 | ComponentInstance 重导出 |
| `validator.go` | ✅ 保留 | HookValidator 重导出 |
| `layout.go` | ✅ 保留 | HStack/VStack 重导出 |
| `hooks.go` | ✅ 保留 | useState/useEffect 等 |
| `app.go` | ✅ 保留 | Run 入口 |
| `shortcuts.go` | ✅ 保留 | Text/Styled 快捷函数 |
| `test.go` | ✅ 保留 | 测试辅助 |
| `keys.go` | ✅ 保留 | 键盘快捷键 |
| `compat.go` | ✅ 保留 | 向后兼容存根 |
| `begin_work.go` | ✅ 保留 | 工作流辅助 |
| `complete_work.go` | ✅ 保留 | 工作流辅助 |
| `memory_safety.go` | ✅ 保留 | 内存安全 |

### 1.2 迁移但保留的文件 (被 internal/ 使用)

| 文件 | 新位置 | 保留原因 |
|------|--------|----------|
| `instance_manager.go` | `internal/state/instance_manager.go` | 被 `internal/reconciler/` 依赖 |
| `interaction_state.go` | `internal/state/interaction_state.go` | 被 `internal/reconciler/` 依赖 |
| `scheduler.go` | `internal/scheduler/ui_scheduler.go` | 被 `internal/reconciler/` 依赖 |
| `reconciler.go` | `internal/reconciler/reconciler.go` | 被 framework 依赖 |

**依赖关系**:
```
framework → internal/reconciler → ui.InstanceManager
                                   → ui.InteractionStateManager
                                   → ui.Scheduler
```

### 1.3 测试文件

| 文件 | 状态 | 说明 |
|------|------|------|
| `*_test.go` (所有) | ✅ 保留 | 单元测试 |

### 1.4 组件相关文件 (已迁移到 components/)

以下组件已迁移到 `components/` 目录：
- Button → `components/button/`
- Input → `components/form/`
- Checkbox → `components/form/`
- Select → `components/form/`
- Textarea → `components/form/`
- Text → `components/basic/`
- Progress → `components/feedback/`
- Spinner → `components/feedback/`
- Table → `components/data/`
- VirtualList → `components/data/`
- Tabs → `components/navigation/`
- Modal → `components/overlay/`
- Tooltip → `components/overlay/`
- Toast → `components/overlay/`
- Divider → `components/basic/`
- HStack/VStack/Box → `components/layout/`
- Absolute/Grid → `components/layout/`

---

## 2. 清理建议

### 2.1 可以删除的文件

**无** - 目前 `ui/` 目录中的所有文件都被使用：
- 直接被 framework 使用
- 被 internal/reconciler 使用
- 被 app 包重导出
- 被测试使用
- 被示例使用

### 2.2 需要重构的部分

`instance_manager.go`, `interaction_state.go`, `scheduler.go`, `reconciler.go`

这些文件虽然已迁移到 `internal/`，但 `ui/` 版本仍然保留是因为：

1. **向后兼容** - 旧代码可能直接导入 `ui.InstanceManager`
2. **依赖链** - `internal/reconciler` 仍然使用 `ui.InstanceManager` 类型

**建议重构路径**:
1. 将 `internal/reconciler` 改为使用 `internal/state` 的类型
2. 添加类型别名或适配器
3. 然后删除 `ui/` 中的重复文件

---

## 3. 当前架构图

```
┌─────────────────────────────────────────────────────────────┐
│                         framework/App                         │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
                ┌─────────────────────────────┐
                │  internal/reconciler/        │
                │  - uses ui.InstanceManager    │
                │  - uses ui.InteractionState   │
                │  - uses ui.Scheduler          │
                └─────────────────────────────┘
                           │
                           ▼
                ┌─────────────────────────────┐
                │       ui/ (re-exports)       │
                │  - InstanceManager           │
                │  - InteractionState          │
                │  - Scheduler                 │
                └─────────────────────────────┘
                           │
                           ▼
                ┌─────────────────────────────┐
                │   internal/state/ (new)       │
                │   internal/scheduler/ (new)  │
                │   runtime/ui/ (new)           │
                └─────────────────────────────┘
```

---

## 4. 结论

**ui/ 目录中没有遗留可删除的代码**。所有文件都在使用中，主要分为三类：

1. **重导出文件** - 将 `runtime/ui` 类型重新导出
2. **兼容层** - 被内部 reconciler 使用的类型
3. **测试文件** - 单元测试和集成测试

要真正清理这些文件，需要先重构 `internal/reconciler` 使其不再依赖 `ui/` 包的这些类型。这是一个较大的重构工作，建议作为单独的阶段进行。

---

## 5. 待办事项 (如果需要进一步清理)

- [ ] Phase 9.1: 重构 internal/reconciler 使用 internal/state 类型
- [ ] Phase 9.2: 删除 ui/ 中的 instance_manager.go
- [ ] Phase 9.3: 删除 ui/ 中的 interaction_state.go
- [ ] Phase 9.4: 删除 ui/ 中的 scheduler.go
- [ ] Phase 9.5: 删除 ui/ 中的 reconciler.go
