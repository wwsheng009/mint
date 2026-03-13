# Fiber-First 渲染管线实施任务清单

**创建日期**: 2026-02-19
**最后更新**: 2026-02-20
**状态**: 进行中
**基于文档**: `docs/render/paint/optimized/refactor/phase*.md`

---

## 实施进度总览

| Phase | 名称 | 状态 | 预计时间 |
|-------|------|------|----------|
| Phase 1 | Fiber 结构优化 | ✅ 完成 | 1-2 天 |
| Phase 2 | Layout 引擎优化 | ✅ 完成 | 2-3 天 |
| Phase 3 | Paint 引擎优化 | ✅ 完成 | 1-2 天 |
| Phase 4 | 渲染管线集成 | ✅ 完成 | 3-5 天 |
| Phase 5 | 组件迁移 | ⏳ 待开始 | 5-7 天 |

---

## Phase 1: Fiber 结构优化

### 状态: ✅ 完成

### 已完成
- [x] Fiber 结构包含 `Instance` 字段
- [x] Fiber 结构包含 `Style` 字段
- [x] `GetInstance()` 方法已实现
- [x] `GetPaintableInstance()` 方法已实现
- [x] `GetFocusableInstance()` 方法已实现
- [x] **Task 1.1**: 添加 `FlagLayoutDirty` 和 `FlagPaintDirty` 常量
- [x] **Task 1.4**: 添加辅助方法 (`HasInstance`, `HasStyle`, `IsLayoutDirty`, `MarkLayoutDirty`, `ClearLayoutDirty`, `IsPaintDirty`, `MarkPaintDirty`, `ClearPaintDirty`)

### 待完成

#### Task 1.2: 删除重复的 ComponentInstance 字段
**文件**: `runtime/ui/fiber.go`
**优先级**: P1

- [ ] 搜索所有使用 `fiber.ComponentInstance` 的代码
- [ ] 替换为 `fiber.Instance`
- [ ] 删除 `ComponentInstance` 字段定义
- [ ] 更新 `GetInstance()` 方法

#### Task 1.3: 删除 Deprecated 字段
**文件**: `runtime/ui/fiber.go`
**优先级**: P2

- [ ] 确认所有 Focusable 操作已迁移到 Instance
- [ ] 删除 `FocusableVNode` 字段
- [ ] 删除 `FocusableMeta` 字段

#### Task 1.4: 添加辅助方法
**文件**: `runtime/ui/fiber.go`
**优先级**: P2

```go
// HasInstance returns true if fiber has an instance.
func (f *Fiber) HasInstance() bool {
    return f.Instance != nil
}

// HasStyle returns true if fiber has style defined.
func (f *Fiber) HasStyle() bool {
    return f.Style != nil
}
```

---

## Phase 2: Layout 引擎优化

### 状态: ✅ 完成

### 已完成
- [x] `FiberToNodeAdapter` 已实现 (`internal/render/fiber_adapter.go`)
- [x] 适配器实现 `layout.Node` 接口
- [x] 适配器实现 `layout.Layered` 接口
- [x] 适配器实现 `layout.Marginal` 接口
- [x] 适配器实现 `layout.Positionable` 接口
- [x] `LayoutSwitcher` 已实现
- [x] **Task 2.1**: 实现 `Measurable` 接口 (`Measure` 方法)
- [x] **Task 2.2**: 实现 `Dirtyable` 接口 (`IsLayoutDirty`, `ClearLayoutDirty`, `MarkLayoutDirty`)

### 待完成

#### Task 2.3: 完善 FiberToPaintableConverter
**文件**: `internal/render/converter.go`
**优先级**: P1

- [ ] 添加 `ConvertLayoutBox` 方法
- [ ] 确保 Fiber + LayoutBox 正确转换为 PaintableBox

#### Task 2.4: 添加单元测试
**文件**: `internal/render/fiber_adapter_test.go`
**优先级**: P1

- [x] 已在 `fiber_first_test.go` 中添加相关测试

---

## Phase 3: Paint 引擎优化

### 状态: ✅ 完成

### 已完成
- [x] `PaintEngine` 已实现 (`internal/render/paint_engine.go`)
- [x] `PaintLayout` 方法已实现
- [x] **Task 3.1**: 标记 `PaintVNode` 为 Deprecated

### 待完成

#### Task 3.2: 添加性能监控
**文件**: `internal/render/paint_engine.go`
**优先级**: P2

```go
// DebugPaintLayout 打印 PaintableLayout 结构
func DebugPaintLayout(layout *paint.PaintableLayout) string {
    // 实现调试输出
}
```

---

## Phase 4: 渲染管线集成

### 状态: ✅ 完成

### 已完成
- [x] **Task 4.1**: 重构 `DeclarativeNode.Paint()` - 添加 Fiber-first 字段
  - 添加 `RenderMode` 类型
  - 添加 `layoutSwitcher`, `paintEngine`, `converter`, `fiberFirstEnabled` 字段
- [x] **Task 4.2**: 实现双轨运行机制
  - `RenderModeLegacy`, `RenderModeFiberFirst`, `RenderModeBoth`
  - 根据 `MINT_FIBER_FIRST` 环境变量选择渲染模式
- [x] **Task 4.3**: 实现 `fiberFirstPaint` 方法
  - Phase 1: Fiber Reconciliation
  - Phase 2: Fiber-based Layout
  - Phase 3: Paint using PaintableBox
- [x] **Task 4.4**: 添加单元测试 (`fiber_first_test.go`)

### 待完成

#### Task 4.5: 删除 renderWithFiberContext 和 nonFiberRender
**文件**: `internal/render/declarative_node.go`
**优先级**: P1

- [ ] 删除 `renderWithFiberContext` 方法
- [ ] 删除 `nonFiberRender` 方法

---

## Phase 5: 组件迁移

### 状态: ⏳ 待开始

### P0: 基础组件

#### Task 5.1: Text 组件迁移
**目标位置**: `ui/components/text/`
- [ ] 创建 `vnode.go`
- [ ] 创建 `instance.go`
- [ ] 创建测试文件

#### Task 5.2: Stack 组件迁移 (VStack/HStack)
**目标位置**: `ui/components/stack/`
- [ ] 创建 `vnode.go`
- [ ] 创建 `instance.go`
- [ ] 创建测试文件

#### Task 5.3: Spacer 组件迁移
**目标位置**: `ui/components/spacer/`
- [ ] 创建 `vnode.go`
- [ ] 创建 `instance.go`

### P1: 交互组件

#### Task 5.4: Input 组件迁移
**目标位置**: `ui/components/input/`
- [ ] 创建 `vnode.go`
- [ ] 创建 `instance.go`
- [ ] 实现 `ActionHandlerInstance` 接口

#### Task 5.5: Checkbox 组件迁移
**目标位置**: `ui/components/checkbox/`
- [ ] 创建 `vnode.go`
- [ ] 创建 `instance.go`

---

## 验证检查清单

### Phase 1 验证
- [ ] `go test ./runtime/ui -run TestFiber -v`
- [ ] 无编译警告
- [ ] Fiber 不再持有 VNode 引用

### Phase 2 验证
- [ ] `go test ./internal/render -run TestFiberToNodeAdapter -v`
- [ ] `go test ./runtime/layout -v`
- [ ] Layout 不访问 VNode

### Phase 3 验证
- [ ] `go test ./internal/render -run TestPaintEngine -v`
- [ ] Paint 只使用 PaintableBox

### Phase 4 验证
- [ ] `go test ./internal/render -run TestDeclarativeNode -v`
- [ ] 示例应用正常运行
- [ ] 双轨对比无差异

### Phase 5 验证
- [ ] 所有迁移组件测试通过
- [ ] 示例应用正常运行

---

## 架构约束检查

在实施过程中，必须确保以下约束：

```
❌ runtime/layout 不导入 runtime/ui
❌ runtime/layout 不导入 internal/reconciler
❌ Layout 阶段不访问 VNode
❌ Paint 阶段不访问 VNode
❌ Fiber 不持有 VNode 引用
❌ VNode 不持有闭包（使用 ActionTargetID）
```

---

## 依赖关系图

```
Phase 1 (Fiber 结构)
    ↓
Phase 2 (Layout 引擎) ← 依赖 Phase 1 Task 1.1
    ↓
Phase 3 (Paint 引擎)
    ↓
Phase 4 (渲染管线) ← 依赖 Phase 2 + Phase 3
    ↓
Phase 5 (组件迁移)
```

---

## 立即可开始的任务

以下任务无依赖，可以立即开始：

1. **Task 1.1**: 添加 Fiber Flags 常量 (15分钟)
2. **Task 1.4**: 添加辅助方法 (15分钟)
3. **Task 3.1**: 标记 PaintVNode 为 Deprecated (15分钟)

---

**维护者**: Fiber-first 架构团队
**最后更新**: 2026-02-19
