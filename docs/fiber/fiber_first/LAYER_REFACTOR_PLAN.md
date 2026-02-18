# Layer 系统重构计划

**Date**: 2026-02-17
**Status**: Phase 1 - Layer-aware Focus Manager
**Based on**: docs/fiber/diff_layer.md

## 设计原则

根据 `diff_layer.md`：

1. **Layer 是渲染排序维度**，不是结构维度
2. **Fiber 树保持单一结构**，不被 Strip 分裂
3. **Layer 不参与 diff 规则**，只参与渲染阶段排序
4. **HitTest 从最高 Layer 开始检测**
5. **事件冒泡按 Fiber Parent，不按 Layer**

## 当前状态

### ✅ 已完成

1. **Fiber.Layer 字段** - 每个 Fiber 存储其所属 Layer
2. **FiberFocusManager.activeLayer** - 焦点管理器支持 Layer 感知
3. **FocusFirst/FocusLast/FocusNext/FocusPrev** - 都支持 activeLayer
4. **Reconciler.hasLayerFibers** - 检测 Modal/Overlay 层存在
5. **Reconciler.updateFocusManagerFromFiber** - 自动设置 activeLayer

### ⏳ 待验证

1. Modal 打开时焦点是否自动跳转到 Modal 内
2. Tab 导航是否在 Modal 层内循环

## 关键文件改动

| 文件 | 改动 |
|------|------|
| `runtime/ui/fiber_focus_manager.go` | ✅ 添加 activeLayer 支持 |
| `internal/reconciler/reconciler.go` | ✅ 自动检测 Layer 并设置 activeLayer |

## 测试状态

当前测试仍失败，焦点在 Quit 按钮上。

**可能原因**：
1. 初始渲染时焦点已设置，之后 Tab 导航改变了焦点
2. Modal 检测时机不对
3. 测试逻辑问题（焦点导航方向）

## 下一步

1. 添加调试日志验证 activeLayer 设置
2. 检查焦点初始化逻辑
3. 验证 Modal 打开后焦点跳转