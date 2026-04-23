# Mint UI 重构 TODO 清单

> **创建日期**: 2026-02-01
> **状态**: 📋 进行中
> **目标**: 将声明式组件从单一 declarativeRoot 架构重构为可独立渲染的多组件架构

---

## 目录

1. [Phase 0: 准备阶段](#phase-0-准备阶段)
2. [Phase 1: 类型基础包迁移](#phase-1-类型基础包迁移)
3. [Phase 2: 基础架构重组](#phase-2-基础架构重组)
4. [Phase 3: 内部模块迁移](#phase-3-内部模块迁移)
5. [Phase 4: 渲染系统重构](#phase-4-渲染系统重构)
6. [Phase 5: 多组件支持](#phase-5-多组件支持)
7. [Phase 6: API 入口层](#phase-6-api-入口层)
8. [Phase 7: 测试与验证](#phase-7-测试与验证)
8. [Phase 8: 文档更新](#phase-8-文档更新)

---

## Phase 0: 准备阶段

### 0.1 环境准备

- [x] 创建 `components/` 目录结构
  - [x] `components/basic/`
  - [x] `components/layout/`
  - [x] `components/form/`
  - [x] `components/button/`
  - [x] `components/feedback/`
  - [x] `components/data/`
  - [x] `components/navigation/`
  - [x] `components/overlay/`
  - [x] `components/container/`

- [x] 创建 `internal/` 目录结构
  - [x] `internal/reconciler/`

- [x] 创建 `runtime/ui/` 目录结构（新增）
  - [x] `runtime/ui/vnode.go`
  - [x] `runtime/ui/element.go`
  - [x] `runtime/ui/component.go`
  - [x] `runtime/ui/fragment.go`
  - [x] `runtime/ui/fiber.go`
  - [x] `runtime/ui/fiber_util.go`
  - [x] `runtime/ui/hooks.go`
  - [x] `runtime/ui/instance.go`
  - [x] `runtime/ui/validator.go`
  - [x] `runtime/ui/layout.go`

### 0.2 代码分析

- [x] 统计 `ui/` 目录下所有文件及其依赖关系
- [x] 识别需要迁移的文件列表
- [x] 分析循环依赖风险

### 0.3 测试基准

- [x] 运行所有测试，记录通过率
- [x] 运行所有示例，记录当前行为

---

## Phase 1: 类型基础包迁移 ✅ 已完成

> **完成日期**: 2026-02-02

### 1.1 创建 runtime/ui/ 包

- [x] `runtime/ui/vnode.go`
  - [x] 定义 `VNode` 接口
  - [x] 定义 `VNodeType` 枚举
  - [x] 定义 `Props` 类型
  - [x] 定义 `ComponentFunc` 和 `ComponentFuncWithProps`
  - [x] Props 方法: Get, Set, GetString, GetInt, GetBool, GetFunc, Merge, Clone

- [x] `runtime/ui/element.go`
  - [x] 定义 `ElementVNode` 结构
  - [x] 定义 `ElementBuilder`
  - [x] 实现 `NewElement()` 函数
  - [x] 实现 `Element()` 函数

- [x] `runtime/ui/component.go`
  - [x] 定义 `ComponentVNode` 结构
  - [x] 定义 `ComponentBuilder`
  - [x] 实现 `NewComponent()` 函数
  - [x] 实现 `NewComponentWithProps()` 函数
  - [x] 实现 `Component()` 函数

- [x] `runtime/ui/fragment.go`
  - [x] 定义 `FragmentVNode` 结构
  - [x] 实现 `NewFragment()` 函数
  - [x] 实现 `Fragment()` 函数

- [x] `runtime/ui/fiber.go`
  - [x] 定义 `Lane` 类型
  - [x] 定义 Lane 常量 (LaneNoLane, LaneSyncLane, etc.)
  - [x] 定义 `EffectFlag` 类型
  - [x] 定义 `Fiber` 结构
  - [x] 定义 `Update` 和 `UpdateQueue` 结构
  - [x] Fiber 方法: HasNoPendingWork, HasEffect, HasSubtreeEffect, MarkUpdate, EnqueueUpdate

- [x] `runtime/ui/fiber_util.go`
  - [x] 实现 `CreateFiber(vnode) *Fiber`
  - [x] 实现 `CreateFiberFromVNode(vnode) *Fiber`
  - [x] 实现 `buildFiberTree()`
  - [x] 实现 `CloneFiber(fiber) *Fiber`
  - [x] 实现 `WalkFiberDepthFirst()`
  - [x] 实现 `WalkFiberBreadthFirst()`
  - [x] 实现 `MergeLanes(a, b Lane) Lane`
  - [x] 实现 `FindFiberByKey()`, `CountFibers()`, `GetFiberDepth()`, `CollectFibersWithFlags()`

- [x] `runtime/ui/hooks.go`
  - [x] 定义 `HookType` 枚举
  - [x] 定义 `Hook` 结构
  - [x] 定义 `ComponentContext` 结构
  - [x] 定义 `Ref` 结构
  - [x] 定义 `EffectCallback` 和 `CleanupFunc` 类型
  - [x] 实现 `NextComponentID()` 函数
  - [x] 实现 `SetCurrentContext()` 函数
  - [x] 实现 `GetCurrentContext()` 函数
  - [x] 实现 `NewComponentContext()` 函数
  - [x] 实现 `NewComponentContextForRoot()` 函数
  - [x] 实现 `ComponentContext` 方法: ResetContext, FinishRender, RunEffects, CleanupAll, GetOrCreateHook

- [x] `runtime/ui/instance.go`
  - [x] 定义 `ComponentInstance` 接口
  - [x] 定义 `BaseComponentInstance` 结构
  - [x] 实现 `NewBaseComponentInstance()`
  - [x] 实现 `NewBaseComponentInstanceWithProps()`
  - [x] 实现 ComponentInstance 接口的所有方法
  - [x] 实现 BaseComponentInstance 的所有方法

- [x] `runtime/ui/validator.go`
  - [x] 定义 `HookValidator` 结构
  - [x] 定义 `HookOrderError` 结构
  - [x] 实现 `NewHookValidator()`
  - [x] 实现 `ValidateHookCall()`
  - [x] 实现 `FinishRender()`
  - [x] 实现 `Reset()`

- [x] `runtime/ui/layout.go`
  - [x] 定义 `Direction` 枚举
  - [x] 定义 `Align` 枚举
  - [x] 定义 `LayoutNode` 结构
  - [x] 实现 `HStack()` 函数
  - [x] 实现 `VStack()` 函数
  - [x] 实现 `Box()` 函数
  - [x] 实现 `Spacer()` 函数
  - [x] 实现 `LayoutBuilder` 结构及方法
  - [x] 实现 `BoxLayoutBuilder` 结构及方法
  - [x] 实现 `SpacerBuilder` 结构及方法

### 1.2 更新 ui/ 包以重导出 runtime/ui

- [x] `ui/vnode.go`
  - [x] 通过类型别名重导出 `VNode = types.VNode`
  - [x] 通过类型别名重导出 `VNodeType = types.VNodeType`
  - [x] 通过常量重导出 VNode 类型常量
  - [x] 通过类型别名重导出 `Props = types.Props`
  - [x] 通过类型别名重导出 `ComponentFunc` 和 `ComponentFuncWithProps`

- [x] `ui/element.go`
  - [x] 通过类型别名重导出 `ElementVNode = types.ElementVNode`
  - [x] 实现 `NewElement()` 调用 `types.NewElement()`
  - [x] 实现 `Element()` 调用 `types.Element()`

- [x] `ui/component.go`
  - [x] 通过类型别名重导出 `ComponentVNode = types.ComponentVNode`
  - [x] 实现 `NewComponent()` 调用 `types.NewComponent()`
  - [x] 实现 `NewComponentWithProps()` 调用 `types.NewComponentWithProps()`
  - [x] 实现 `Component()` 调用 `types.Component()`

- [x] `ui/fragment.go`
  - [x] 通过类型别名重导出 `FragmentVNode = types.FragmentVNode`
  - [x] 实现 `NewFragment()` 调用 `types.NewFragment()`
  - [x] 实现 `Fragment()` 调用 `types.Fragment()`

- [x] `ui/layout.go`
  - [x] 通过类型别名重导出 `LayoutNode = types.LayoutNode`
  - [x] 通过类型别名重导出 `Direction = types.Direction`
  - [x] 通过类型别名重导出 `Align = types.Align`
  - [x] 重导出 Direction 和 Align 常量
  - [x] 实现 `HStack()` 调用 `types.HStack()`
  - [x] 实现 `VStack()` 调用 `types.VStack()`
  - [x] 实现 `Box()` 调用 `types.Box()`
  - [x] 实现 `Spacer()` 调用 `types.Spacer()`

- [x] `ui/fiber.go`
  - [x] 通过类型别名重导出 `Lane = types.Lane`
  - [x] 重导出 Lane 常量
  - [x] 通过类型别名重导出 `EffectFlag = types.EffectFlag`
  - [x] 重导出 EffectFlag 常量
  - [x] 通过类型别名重导出 `Fiber = types.Fiber`
  - [x] 通过类型别名重导出 `Update = types.Update`
  - [x] 通过类型别名重导出 `UpdateQueue = types.UpdateQueue`
  - [x] 重导出 Fiber 函数: CreateFiber, CreateFiberFromVNode, WalkFiberDepthFirst, WalkFiberBreadthFirst, CloneFiber, MergeLanes, FindFiberByKey, CountFibers, GetFiberDepth, CollectFibersWithFlags

- [x] `ui/hooks.go`
  - [x] 通过类型别名重导出 `HookType = types.HookType`
  - [x] 重导出 HookType 常量
  - [x] 通过类型别名重导出 `ComponentContext = types.ComponentContext`
  - [x] 通过类型别名重导出 `Ref = types.Ref`
  - [x] 通过类型别名重导出 `EffectCallback = types.EffectCallback`
  - [x] 通过类型别名重导出 `CleanupFunc = types.CleanupFunc`
  - [x] 实现 SetCurrentContext() 调用 types.SetCurrentContext()
  - [x] 实现 GetCurrentContext() 调用 types.GetCurrentContext()
  - [x] 实现 NewComponentContextForRoot() 调用 types.NewComponentContextForRoot()
  - [x] 更新 useState() 使用 types.GetCurrentContext()
  - [x] 更新 UseStateInt() 使用 types.GetCurrentContext()
  - [x] 更新 useEffect() 使用 types.GetCurrentContext()
  - [x] 更新 UseRef() 使用 types.GetCurrentContext()
  - [x] 更新 UseMemo() 使用 types.GetCurrentContext()
  - [x] 更新 useHoverState() 使用 types.GetCurrentContext()
  - [x] 更新 UseStateIntWithDebug() 使用 types.GetCurrentContext()

- [x] `ui/instance.go`
  - [x] 通过类型别名重导出 `ComponentInstance = types.ComponentInstance`
  - [x] 通过类型别名重导出 `BaseComponentInstance = types.BaseComponentInstance`
  - [x] 实现 NewBaseComponentInstance() 调用 types.NewBaseComponentInstance()
  - [x] 实现 NewBaseComponentInstanceWithProps() 调用 types.NewBaseComponentInstanceWithProps()

- [x] `ui/validator.go`
  - [x] 通过类型别名重导出 `HookValidator = types.HookValidator`
  - [x] 通过类型别名重导出 `HookOrderError = types.HookOrderError`
  - [x] 实现 NewHookValidator() 调用 types.NewHookValidator()

- [x] `ui/compat.go` (临时兼容层)
  - [x] 为 stub 类型添加 accessor 方法，支持 internal/reconciler
  - [x] InputVNode: Value(), Placeholder()
  - [x] TextareaVNode: Value(), Placeholder()
  - [x] CheckboxVNode: Label(), Checked()
  - [x] SelectVNode: Selected(), Options(), SelectOption 类型
  - [x] ModalVNode: Width(), Height(), Title(), IsOpen(), Content(), Footer()
  - [x] TabsVNode: ActiveTab()
  - [x] TableVNode: Columns(), Rows()
  - [x] VirtualListVNode: ListHeight(), ItemCount(), ItemHeight()
  - [x] ProgressVNode: Value(), Max(), Percent(), Width()

### 1.3 更新 internal/reconciler 使用 runtime/ui

- [x] `internal/reconciler/vnode_converter.go`
  - [x] 更新导入使用 `github.com/wwsheng009/mint/runtime/ui`
  - [x] 通过 `ui.` 前缀访问类型（ui 从 runtime/ui 重导出）

---

## Phase 2: 基础架构重组

### 2.1 核心接口定义

- [x] `runtime/paint/paintable.go`
  - [x] 定义 `Paintable` 接口

### 2.2 VNode 接口扩展

- [x] `runtime/ui/` - VNode 接口完整定义
- [x] `ui/` - 重导出 VNode 接口

---

## Phase 3: 内部模块迁移

### 3.1 Reconciler 系统

- [x] `ui/reconciler.go` → `internal/reconciler/reconciler.go`
  - [x] 修改 package 为 `reconciler`
  - [x] 更新所有导入路径
  - [x] 使用 `runtime/ui` 类型

- [x] `internal/reconciler/vnode_converter.go`
  - [x] 实现 VNode 到 LayoutNode 的转换
  - [x] 支持所有组件类型的转换

### 3.2 Scheduler 系统

- [x] `ui/scheduler.go` → `internal/scheduler/`
  - [x] 修改 package 为 `scheduler`
  - [x] 更新所有导入路径

### 3.3 State 系统

- [x] `ui/instance_manager.go` → `internal/state/instance_manager.go`
  - [x] 修改 package 为 `state`
  - [x] 更新所有导入路径

- [x] `ui/interaction_state.go` → `internal/state/interaction_state.go`
  - [x] 修改 package 为 `state`
  - [x] 更新所有导入路径

---

## Phase 4: 渲染系统重构

### 4.1 Paintable 接口

- [x] `runtime/paint/paintable.go`
  - [x] 定义 `Paintable` 接口
  - [x] `Paint(x, y int) []DrawCmd` 方法

### 4.2 组件 Measure/Paint 实现

- [x] `components/basic/text.go`
  - [x] 实现 `Measure()` 方法
  - [x] 实现 `Paint()` 方法

- [x] `components/button/button.go`
  - [x] 实现 `Measure()` 方法
  - [x] 实现 `Paint()` 方法

- [x] `components/form/` (input.go, textarea.go, checkbox.go, select.go)
  - [x] 实现 `Measure()` 方法
  - [x] 实现 `Paint()` 方法

- [x] `components/feedback/` (progress.go, spinner.go)
  - [x] 实现 `Measure()` 方法
  - [x] 实现 `Paint()` 方法

- [x] `components/data/` (table.go, virtuallist.go)
  - [x] 实现 `Measure()` 方法
  - [x] 实现 `Paint()` 方法

- [x] `components/navigation/` (tabs.go)
  - [x] 实现 `Measure()` 方法
  - [x] 实现 `Paint()` 方法

- [x] `components/overlay/` (modal.go)
  - [x] 实现 `Measure()` 方法
  - [x] 实现 `Paint()` 方法

---

## Phase 5: 多组件支持

### 5.1 DeclarativeNode

- [x] `internal/render/declarative_node.go`
  - [x] 定义 `DeclarativeNode` 结构
  - [x] 实现 `component.Node` 接口
  - [x] 实现 `component.Measurable` 接口
  - [x] 实现 `component.Paintable` 接口
  - [x] 实现 `component.Mountable` 接口
  - [x] 实现 `Root()` 和 `UpdateRoot()` 方法

### 5.2 混合渲染支持

- [x] `internal/render/declarative_node.go`
  - [x] 支持将 VNode 包装为 framework.Component
  - [x] 支持 VNode 树的渲染
  - [x] Paint 方法遍历 VNode 树并绘制

---

## Phase 6: API 入口层

### 6.1 shortcuts.go

- [ ] `ui/shortcuts.go`
  - [ ] 基础组件快捷函数
  - [ ] 布局组件快捷函数
  - [ ] 表单组件快捷函数
  - [ ] 按钮组件快捷函数

### 6.2 app.go 精简

- [x] `ui/app.go`
  - [x] 移除已迁移的组件代码
  - [x] 保留 `Run()` 入口
  - [x] 保留 `declarativeRoot` (向后兼容)

---

## Phase 7: 测试与验证

### 7.1 单元测试

| 包 | 当前覆盖率 | 目标 | 状态 |
|---|---|---|---|
| runtime/action | 71.5% | 70% | ✅ |
| runtime/core | 70.4% | 70% | ✅ |
| runtime/focus | 71.1% | 70% | ✅ |
| runtime/layout | 89.8% | 70% | ✅ |
| runtime/paint | 86.9% | 70% | ✅ |
| runtime/render | 85.8% | 70% | ✅ |
| runtime/scheduler | 68.0% | 70% | 🟡 |
| runtime/selection | 67.3% | 70% | 🟡 |
| runtime/ui | 60.8% | 60% | ✅ |
| runtime (overall) | 51.3% | 50% | ✅ |
| runtime/event | 38.2% | 50% | ⏳ |
| runtime/input | 49.1% | 50% | 🟡 |
| internal/render | 61.6% | 60% | ✅ |
| internal/reconciler | 45.1% | 60% | ⏳ |

- [x] `runtime/ui/` 测试覆盖率 > 60% ✅ (60.8%)
- [x] `runtime/focus` 测试覆盖率 > 70% ✅ (71.1%)
- [x] `runtime/core` 测试覆盖率 > 70% ✅ (70.4%)
- [x] `internal/render` 测试覆盖率 > 60% ✅ (61.6%)
- [ ] `internal/reconciler/` 测试覆盖率 > 60% ⏳ (45.1%)
- [ ] `components/*/` 测试覆盖率 > 70% ⏳

### 7.2 已添加测试文件

| 测试文件 | 覆盖内容 |
|---|---|
| runtime/ui/compat_test.go | VNode wrapper types (TextVNode, ButtonVNode, InputVNode, etc.) |
| runtime/runtime_test.go | BoxConstraints, Position, Frame, SetContentRuntime |
| runtime/focus_test.go | FocusManager 完整功能测试 |
| runtime/core/runtime_test.go | Runtime lifecycle, state management |
| runtime/selection/selection_test.go | Selection, Renderer, RuntimeAdapter |
| internal/render/declarative_node_extensions_test.go | DeclarativeNode extension methods |

### 7.3 集成测试

- [ ] 创建 `tests/integration/` 目录
- [ ] `integration/render_test.go`
- [ ] `integration/multi_component_test.go`
- [ ] `integration/hooks_test.go`

### 7.4 回归测试

- [x] 运行所有现有示例 ✅
- [x] 确保行为一致 ✅

---

## Phase 8: 文档更新

### 8.1 API 文档

- [x] 更新 `docs/plan/COMPREHENSIVE_REFACTOR_PLAN.md`
- [ ] 更新 `README.md`
- [ ] 更新 `docs/API.md`
- [ ] 为每个组件包添加 `README.md`

### 8.2 设计文档

- [ ] 更新 `docs/ARCHITECTURE.md`
- [ ] 更新 `docs/RENDERING.md`
- [ ] 更新 `docs/COMPONENTS.md`

### 8.3 迁移指南

- [ ] 完成组件迁移指南
- [ ] 添加 API 迁移对照表
- [ ] 添加常见问题解答

---

## 检查清单

### 每个迁移项完成后检查

- [ ] 文件已复制到目标位置
- [ ] package 声明已更新
- [ ] 导入路径已更新
- [ ] 组件功能保持一致
- [ ] Builder 模式正常工作
- [ ] 单元测试已编写并通过
- [ ] 集成测试已通过
- [ ] 示例程序正常运行
- [ ] 文档已同步更新

### 发布前检查

- [ ] `go build ./...` 通过
- [ ] `go test ./... -cover` 覆盖率达标
- [ ] `go vet ./...` 无警告
- [ ] `golangci-lint run` 通过
- [ ] 所有示例正常运行
- [ ] 性能无明显退化
- [ ] 文档完整

---

## 进度跟踪

```
Phase 0: 准备阶段           [████████████████████████████] 100%
Phase 1: 类型基础包迁移       [████████████████████████████] 100% ✅
Phase 2: 基础架构重组       [████████████████████████████] 100%
Phase 3: 内部模块迁移       [████████████████████████████] 100% ✅
Phase 4: 渲染系统重构       [████████████████████████████] 100% ✅
Phase 5: 多组件支持         [████████████████████████████] 100% ✅
Phase 6: API 入口层         [███████████████░░░░░░░░░░░░░░░░░░] 70%
Phase 7: 测试与验证         [███████████████████░░░░░░░░░░] 65%
Phase 8: 文档更新           [████████████░░░░░░░░░░░░░░░░░░░] 50%
```

---

## 已完成组件清单

### Phase 1: 类型基础包迁移 (已完成)

| 文件 | 状态 |
|------|------|
| runtime/ui/vnode.go | ✅ |
| runtime/ui/element.go | ✅ |
| runtime/ui/component.go | ✅ |
| runtime/ui/fragment.go | ✅ |
| runtime/ui/fiber.go | ✅ |
| runtime/ui/fiber_util.go | ✅ |
| runtime/ui/hooks.go | ✅ |
| runtime/ui/instance.go | ✅ |
| runtime/ui/validator.go | ✅ |
| runtime/ui/layout.go | ✅ |

| 文件 | 状态 |
|------|------|
| ui/vnode.go (重导出) | ✅ |
| ui/element.go (重导出) | ✅ |
| ui/component.go (重导出) | ✅ |
| ui/fragment.go (重导出) | ✅ |
| ui/layout.go (重导出) | ✅ |
| ui/fiber.go (重导出) | ✅ |
| ui/hooks.go (使用 types) | ✅ |
| ui/instance.go (重导出) | ✅ |
| ui/validator.go (重导出) | ✅ |
| ui/compat.go (存根) | ✅ |

### Phase 3: 组件库迁移 (已完成)

| 组件分类 | 组件名 | 状态 |
|---------|--------|------|
| basic | Text | ✅ |
| basic | Divider | ✅ |
| layout | HStack, VStack, Box, Spacer | ✅ |
| form | Input | ✅ |
| form | TextArea | ✅ |
| form | Checkbox | ✅ |
| form | Select | ✅ |
| button | Button | ✅ |
| feedback | Progress | ✅ |
| feedback | Spinner | ✅ |
| data | Table | ✅ |
| data | VirtualList | ✅ |
| navigation | Tabs | ✅ |
| overlay | Modal | ✅ |

### 待迁移组件

| 组件分类 | 组件名 | 状态 |
|---------|--------|------|
| basic | Icon | ⏳ |
| layout | Absolute, Grid | ⏳ |
| form | Switch, Slider | ⏳ |
| button | IconButton | ⏳ |
| feedback | Toast, Alert | ⏳ |
| overlay | Tooltip | ⏳ |
| container | Panel | ⏳ |

---

**文档版本**: v1.4
**最后更新**: 2026-02-04
**更新内容**:
- ✅ 更新 Phase 7: 测试与验证 - 添加详细的测试覆盖率表格
- ✅ 新增已添加测试文件清单
- ✅ 更新进度跟踪 (Phase 4, 5, 7 进度)
- ✅ 测试覆盖率更新:
  - runtime/ui: 50.9% → 60.8% ✅
  - runtime/focus: 71.1% ✅
  - runtime/core: 70.4% ✅
  - runtime (overall): 31.2% → 51.3% ✅
  - internal/render: 57.5% → 61.6% ✅
  - runtime/selection: 61.3% → 67.3% 🟡
