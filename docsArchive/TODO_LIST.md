# Fiber 统一架构重构 - 任务清单

**版本**: 1.0
**日期**: 2026-02-14
**负责人**: [待分配]
**状态**: 进行中

---

## 📋 目录

1. [Phase 1: 基础设施](#phase-1-基础设施)
2. [Phase 2: Layout 重构](#phase-2-layout-重构)
3. [Phase 3: RenderPlane 引入](#phase-3-renderplane-引入)
4. [Phase 4: 废弃 StripLayers](#phase-4-废弃-striplayers)
5. [Phase 5: Render 更新](#phase-5-render-更新)
6. [Phase 6: HitMap 更新](#phase-6-hitmap-更新)
7. [Phase 7: 清理和优化](#phase-7-清理和优化)
8. [Phase 8: 综合测试](#phase-8-综合测试)

---

## 任务状态说明

| 状态标记 | 含义 |
|---------|------|
| ⬜ 未开始 | 任务尚未开始 |
| ⏳ 进行中 | 任务正在进行 |
| ✅ 已完成 | 任务已完成 |
| ⚠️ 阻塞 | 任务被阻塞 |
| ❌ 失败 | 失败需要修复 |
| 🔄 重做 | 需要重做 |
| 🔲 跳过 | 跳过此任务 |

---

## Phase 1: 基础设施

**时间**: 1-2 周
**目标**: 添加必要的字段和方法，不破坏现有功能

### 1.1 Fiber 结构更新

- [ ] **Fiber 新增 Layer 字段** (`runtime/ui/fiber.go`)
  - [ ] ⬜ 添加 `Layer rtui.Layer` 字段
  - [ ] ⬜ 更新字段注释
  - [ ] ⬜ 运行 `go test ./runtime/ui/...` 验证编译通过
  - [ ] ⬜ Code Review: 字段命名和类型

- [ ] **Fiber 新增 ComputedBox 字段** (`runtime/ui/fiber.go`)
  - [ ] ⬜ 添加 `ComputedBox *compute.ComputedBox` 字段
  - [ ] ⬜ 更新字段注释
  - [ ] ⬜ 运行 `go test ./runtime/ui/...` 验证编译通过
  - [ ] ⬜ Code Review: 指针引用的正确性

### 1.2 Fiber 构造函数更新

- [ ] **更新 CreateFiber()** (`runtime/ui/fiber.go`)
  - [ ] ⬜ 从 `vnode.GetLayer()` 拷贝 Layer 到 `fiber.Layer`
  - [ ] ⬜ 处理 Layer 默认值（无效值设为 LayerBase）
  - [ ] ⬜ 设置 `fiber.ComputedBox = nil`
  - [ ] ⬜ 添加单元测试 `TestCreateFiberLayerCopy`
  - [ ] ⬜ 添加单元测试 `TestCreateFiberLayerDefaultValue`
  - [ ] ⬜ 运行测试验证通过
  - [ ] ⬜ Code Review

- [ ] **更新 CloneFiber()** (`runtime/ui/fiber.go`)
  - [ ] ⬜ 拷贝 `current.Layer` 到 `clone.Layer`
  - [ ] ⬜ 设置 `clone.ComputedBox = nil`（Layout 阶段重新计算）
  - [ ] ⬜ 保持 `clone.NodeID = current.NodeID`（重要！）
  - [ ] ⬜ 添加单元测试 `TestCloneFiberLayerPreserved`
  - [ ] ⬜ 添加单元测试 `TestCloneFiberNodeIDPreserved`
  - [ ] ⬜ 运行测试验证通过
  - [ ] ⬜ Code Review

### 1.3 Reconciler 更新

- [ ] **更新 complete_work() 拷贝 Layer** (`internal/reconciler/complete_work.go`)
  - [ ] ⬜ 在 `CompleteWork()` 中添加 Layer 拷贝逻辑
  - [ ] ⬜ `workInProgress.Layer = workInProgress.VNode.GetLayer()`
  - [ ] ⬜ 添加单元测试 `TestCompleteWorkLayerCopied`
  - [ ] ⬜ 运行 reconciler 测试
  - [ ] ⬜ Code Review

- [ ] **验证 diff 算法不变** (`internal/reconciler/diff.go`)
  - [ ] ⬜ 确认 `shouldUpdate()` 仍然使用 DiffKey，不使用 NodeID
  - [ ] ⬜ 确认 `reconcileChildren()` 仍然使用 DiffKey 匹配
  - [ ] ⬜ 添加集成测试 `TestReconcileLayerPreserved`
  - [ ] ⬜ 运行所有 reconciler 测试
  - [ ] ⬜ Code Review: 确认无引入 bug

### 1.4 ComputedBox 结构更新

- [ ] **ComputedBox 新增 NodeID 字段** (`runtime/compute/box.go`)
  - [ ] ⬜ 添加 `NodeID uint64` 字段
  - [ ] ⬜ 更新字段注释
  - [ ] ⬜ 运行 `go test ./runtime/compute/...`

- [ ] **ComputedBox 新增 Layer 字段** (`runtime/compute/box.go`)
  - [ ] ⬜ 添加 `Layer rtui.Layer` 字段
  - [ ] ⬜ 更新字段注释
  - [ ] ⬜ 运行 `go test ./runtime/compute/...`

- [ ] **ComputedBox 新增 Children 字段** (`runtime/compute/box.go`)
  - [ ] ⬜ 添加 `Children []*ComputedBox` 字段
  - [ ] ⬜ 更新字段注释
  - [ ] ⬜ 运行 `go test ./runtime/compute/...`

### 1.5 HitMap API 添加

- [ ] **添加 BuildHitMapFromFiber() API** (`runtime/event/hitmap.go`)
  - [ ] ⬜ 实现 `BuildHitMapFromFiber(root *Fiber) *HitMap`
  - [ ] ⬜ 实现 `walkAndBuild(fiber *Fiber, treeDepth int)` 遍历方法
  - [ ] ⬜ 计算 Z-order: `int(fiber.Layer) * 10000 + treeDepth`
  - [ ] ⬜ 实现 `sortByLayerAndZOrder()` 排序方法
  - [ ] ⬜ 添加单元测试 `TestBuildHitMapFromFiberBasic`
  - [ ] ⬜ 添加单元测试 `TestBuildHitMapFromFiberZOrder`
  - [ ] ⬜ 添加单元测试 `TestBuildHitMapFromFiberLayers`
  - [ ] ⬜ 运行所有 event 测试
  - [ ] ⬜ Code Review

### 1.6 测试覆盖率

- [ ] **添加 Fiber 单元测试** (`runtime/ui/fiber_test.go`)
  - [ ] ⬜ `TestFiberLayerCopyFromVNode`
  - [ ] ⬜ `TestFiberLayerDefaultValue`
  - [ ] ⬜ `TestFiberComputedBoxAfterLayout`
  - [ ] ⬜ `TestCloneFiberLayerPreserved`
  - [ ] ⬜ `TestCloneFiberNodeIDPreserved`

- [ ] **添加 Reconciler 单元测试** (`internal/reconciler/complete_work_test.go`)
  - [ ] ⬜ `TestCompleteWorkLayerCopied`
  - [ ] ⬜ `TestReconcileLayerPreserved`
  - [ ] ⬜ `TestReconcileComplexTreeWithLayers`

- [ ] **添加 HitMap 单元测试** (`runtime/event/hitmap_fiber_test.go`)
  - [ ] ⬜ `TestBuildHitMapFromFiberBasic`
  - [ ] ⬜ `TestBuildHitMapFromFiberZOrder`
  - [ ] ⬜ `TestBuildHitMapFromFiberLayers`
  - [ ] ⬜ `TestBuildHitMapFromFiberModalPriority`

### 1.7 集成测试

- [ ] **运行所有现有测试**
  - [ ] ⬜ `go test ./...` - 运行所有测试
  - [ ] ⬜ 确保新增字段不影响现有功能
  - [ ] ⬜ 修复任何失败的测试
  - [ ] ⬜ 记录失败原因（如果有）

- [ ] **运行示例程序**
  - [ ] ⬜ `cd examples/counter && go run .`
  - [ ] ⬜ `cd examples/modal && go run .`
  - [ ] ⬜ `cd examples/overlay && go run .`
  - [ ] ⬜ 确保所有示例正常运行

### 1.8 文档更新

- [ ] **更新 Fiber 文档** (`runtime/ui/fiber.go`)
  - [ ] ⬜ 更新结构体注释
  - [ ] ⬜ 添加 Layer 字段说明
  - [ ] ⬜ 添加 ComputedBox 字段说明
  - [ ] ⬜ 更新示例代码

- [ ] **更新 AGENTS.md**
  - [ ] ⬜ 记录 Fiber 新字段
  - [ ] ⬜ 记录新 API

### 1.9 Phase 1 验收

- [ ] ⬜ 所有单元测试通过
- [ ] ⬜ 所有集成测试通过
- [ ] ⬜ 所有示例程序正常运行
- [ ] ⬜ Code Review 通过
- [ ] ⬜ 测试覆盖率 > 70%
- [ ] ⬜ 无性能退化（通过基准测试）

---

## Phase 2: Layout 重构

**时间**: 1-2 周
**目标**: Layout 基于 Fiber 而不是 VNode

### 2.1 Layout API 更新

- [ ] **修改 Engine.Layout() 签名** (`runtime/compute/engine.go`)
  - [ ] ⬜ 添加 `fiber *reconciler.Fiber` 参数
  - [ ] ⬜ 保持向后兼容（可选：保留旧 API 作为 Deprecated）
  - [ ] ⬜ 更新方法注释
  - [ ] ⬜ 运行测试确保不破坏现有调用

- [ ] **实现 layoutFiber() 方法** (`runtime/compute/engine.go`)
  - [ ] ⬜ `layoutFiber(fiber *Fiber, constraints BoxConstraints, depth int) *ComputedBox`
  - [ ] ⬜ 创建/更新 `fiber.ComputedBox`
  - [ ] ⬜ 拷贝 `fiber.NodeID` 到 `computedBox.NodeID`
  - [ ] ⬜ 拷贝 `fiber.Layer` 到 `computedBox.Layer`
  - [ ] ⬜ 设置 `computedBox.Children`
  - [ ] ⬜ 保存 `fiber.ComputedBox = computedBox`
  - [ ] ⬜ 添加单元测试 `TestLayoutFiberBasic`
  - [ ] ⬜ 添加单元测试 `TestLayoutFiberNodeIDPropagated`
  - [ ] ⬜ 添加单元测试 `TestLayoutFiberLayerPropagated`
  - [ ] ⬜ 运行测试
  - [ ] ⬜ Code Review

- [ ] **实现 measureFiber() 方法** (`runtime/compute/engine.go`)
  - [ ] ⬜ `measureFiber(fiber *Fiber, constraints BoxConstraints) Box`
  - [ ] ⬜ 基于 `fiber.VNode.Type()` 计算
  - [ ] ⬜ 基于 `fiber.Props` 计算
  - [ ] ⬜ ⚠️ 不再从 `vnode.GetBounds()` 读取（那是运行时信息）
  - [ ] ⬜ 添加单元测试 `TestMeasureFiber`
  - [ ] ⬜ 运行测试
  - [ ] ⬜ Code Review

- [ ] **实现 layoutFiberChildren() 方法** (`runtime/compute/engine.go`)
  - [ ] ⬜ `layoutFiberChildren(fiber *Fiber, constraints BoxConstraints, depth int) []*ComputedBox`
  - [ ] ⬜ 遍历 fiber.Child 和 fiber.Sibling
  - [ ] ⬜ 递归调用 layoutFiber()
  - [ ] ⬜ 添加单元测试 `TestLayoutFiberChildren`
  - [ ] ⬜ 运行测试
  - [ ] ⬜ Code Review

- [ ] **实现 buildHitMapFromFiber() 方法** (`runtime/compute/engine.go`)
  - [ ] ⬜ `buildHitMapFromFiber(root *Fiber) *event.HitMap`
  - [ ] ⬜ 遍历 Fiber 树，收集 ComputedBox
  - [ ] ⬜ 调用 `event.BuildHitMapFromEntries()`
  - [ ] ⬜ 替代原有 `BuildHitMapFromLayout()`
  - [ ] ⬜ 添加单元测试 `TestEngineBuildHitMapFromFiber`
  - [ ] ⬜ 运行测试
  - [ ] ⬜ Code Review

### 2.2 更新 Layout 调用点

- [ ] **更新 App.Render()** (`framework/app.go`)
  - [ ] ⬜ 传递 Fiber 到 `engine.Layout()`
  - [ ] ⬜ `layout, err := engine.Layout(a.rootComponent, newFiberRoot, a.constraints)`
  - [ ] ⬜ 添加集成测试 `TestAppRenderWithFiberLayout`
  - [ ] ⬜ 运行测试
  - [ ] ⬜ Code Review

- [ ] **查找所有 Layout 调用**
  - [ ] ⬜ `grep -r "engine\.Layout" ./...` 查找所有调用
  - [ ] ⬜ 列出需要更新的文件
  - [ ] ⬜ 逐个更新调用点
  - [ ] ⬜ 验证每个调用点

- [ ] **更新 LayerManager 调用** (`runtime/layer/manager.go`)
  - [ ] ⬜ `CollectAndLayout()` 中传递 Fiber 参数
  - [ ] ⬜ `LayoutBase` 传递 Fiber
  - [ ] ⬜ `LayoutLayer` 传递 Fiber
  - [ ] ⬜ 添加集成测试 `TestLayerManagerWithFiberLayout`
  - [ ] ⬜ 运行测试
  - [ ] ⬜ Code Review

### 2.3 Modal 和 Overlay 布局

- [ ] **调整 Modal 布局逻辑** (`runtime/layer/manager.go`)
  - [ ] ⬜ Modal 使用 `LayerModal` 层级
  - [ ] ⬜ Modal 仍然使用全屏约束
  - [ ] ⬜ 验证 Modal 居中逻辑仍然正确
  - [ ] ⬜ 添加集成测试 `TestModalLayoutWithFiber`
  - [ ] ⬜ 运行测试
  - [ ] ⬜ 手动验证 Modal 位置

- [ ] **调整 Overlay 布局逻辑** (`runtime/layer/manager.go`)
  - [ ] ⬜ Overlay 使用 `LayerOverlay` 层级
  - [ ] ⬜ 验证 Overlay 位置逻辑
  - [ ] ⬜ 添加集成测试 `TestOverlayLayoutWithFiber`
  - [ ] ⬜ 运行测试
  - [ ] ⬜ 手动验证 Overlay 位置

### 2.4 验证布局正确性

- [ ] **编写布局对比测试**
  - [ ] ⬜ 对比 Phase 1 和 Phase 2 的布局结果
  - [ ] ⬜ 验证 NodeID 一致性
  - [ ] ⬜ 验证 Layer 一致性
  - [ ] ⬜ 验证位置（X, Y）一致性
  - [ ] ⬜ 验证尺寸（Width, Height）一致性
  - [ ] ⬜ 记录任何不一致

- [ ] **编写边界测试**
  - [ ] ⬜ 测试空树 layout
  - [ ] ⬜ 测试单节点 layout
  - [ ] ⬜ 测试深层嵌套 layout
  - [ ] ⬜ 测试大量节点 layout

### 2.5 性能优化

- [ ] **Layout 性能基准测试**
  - [ ] ⬜ 添加 `BenchmarkLayoutFiber` 基准测试
  - [ ] ⬜ 对比旧 Layout 的性能
  - [ ] ⬜ 确保无明显退化 (+20% 以内)
  - [ ] ⬜ 如果退化，分析原因并优化

- [ ] **内存使用分析**
  - [ ] ⬜ 运行 `go test -bench=. -benchmem`
  - [ ] ⬜ 检查内存分配情况
  - [ ] ⬜ 优化不必要的分配
  - [ ] ⬜ 验证无内存泄漏

### 2.6 测试覆盖

- [ ] **添加 Layout 单元测试** (`runtime/compute/engine_test.go`)
  - [ ] ⬜ `TestLayoutFiberBasic`
  - [ ] ⬜ `TestLayoutFiberNodeIDPropagated`
  - [ ] ⬜ `TestLayoutFiberLayerPropagated`
  - [ ] ⬜ `TestLayoutFiberChildren`
  - [ ] ⬜ `TestLayoutFiberComplexTree`

- [ ] **添加 Layout 集成测试** (`runtime/compute/engine_integration_test.go`)
  - [ ] ⬜ `TestLayoutWithRealVNode`
  - [ ] ⬜ `TestLayoutWithModal`
  - [ ] ⬜ `TestLayoutWithOverlay`

### 2.7 文档更新

- [ ] **更新 Layout 文档** (`runtime/compute/engine.go`)
  - [ ] ⬜ 更新 Layout() 方法注释
  - [ ] ⬜ 添加 Fiber 参数说明
  - [ ] ⬜ 更新示例代码

- [ ] **更新 AGENTS.md**
  - [ ] ⬜ 记录 Layout 新参数
  - [ ] ⬜ 记录新的布局流程

### 2.8 Phase 2 验收

- [ ] ⬜ 所有单元测试通过
- [ ] ⬜ 所有集成测试通过
- [ ] ⬜ 所有示例程序正常运行
- [ ] ⬜ Modal、Overlay 布局正确
- [ ] ⬜ 性能无明显退化（+20% 以内）
- [ ] ⬜ 测试覆盖率 > 75%
- [ ] ⬜ Code Review 通过

---

## Phase 3: RenderPlane 引入

**时间**: 1 周
**目标**: 引入 RenderPlane，不破坏现有功能

### 3.1 RenderPlanes 类型实现

- [ ] **创建 renderplanes.go 文件** (`runtime/layer/renderplanes.go`)
  - [ ] ⬜ 实现 `RenderPlanes` 结构体
  - [ ] ⬜ 使用 `map[rtui.Layer][]*compute.ComputedBox`
  - [ ] ⬜ 使用 `[]rtui.Layer` 记录渲染顺序
  - [ ] ⬜ 添加结构体注释

- [ ] **实现 NewRenderPlanes() 方法** (`runtime/layer/renderplanes.go`)
  - [ ] ⬜ 初始化 Planes map
  - [ ] ⬜ 初始化 RenderOrder 数组（LayerBase → LayerInspector）
  - [ ] ⬜ 添加单元测试 `TestNewRenderPlanes`
  - [ ] ⬜ 运行测试
  - [ ] ⬜ Code Review

- [ ] **实现 BuildFromFiber() 方法** (`runtime/layer/renderplanes.go`)
  - [ ] ⬜ `BuildFromFiber(root *Fiber)`
  - [ ] ⬜ 实现 `walkAndCollect(fiber *Fiber)` 遍历方法
  - [ ] ⬜ 按 Layer 分桶 ComputedBox
  - [ ] ⬜ 添加单元测试 `TestBuildFromFiberBasic`
  - [ ] 添加单元测试 `TestBuildFromFiberMultipleLayers`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **实现 sortPlanes() 方法** (`runtime/layer/renderplanes.go`)
  - [ ] ⬜ 对每个 Plane 中的 ComputedBox 排序
  - [ ] ⬜ 按 Y 排序，Y 相同按 X 排序
  - [ ] 添加单元测试 `TestSortPlanes`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **实现 GetPlane() 方法** (`runtime/layer/renderplanes.go`)
  - [ ] ⬜ `GetPlane(layer rtui.Layer) []*compute.ComputedBox`
  - [ ] ⬜ 添加单元测试 `TestGetPlane`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **实现 Iterate() 方法** (`runtime/layer/renderplanes.go`)
  - [ ] ⬜ `Iterate(fn func(layer rtui.Layer, box *compute.ComputedBox) bool)`
  - [ ] ⬜ 按渲染顺序遍历
  - [ ] ⬜ 支持提前退出（fn 返回 false）
  - [ ] 添加单元测试 `TestIterate`
  - [ ] 运行测试
  - ] Code Review

- [ ] **实现辅助方法** (`runtime/layer/renderplanes.go`)
  - [ ] ⬜ `AllPlanes() map[rtui.Layer][]*compute.ComputedBox`
  - [ ] ⬜ `HasLayer(layer rtui.Layer) bool`
  - [ ] ⬜ `GetHighestLayer() rtui.Layer`
  - [ ] ⬜ 添加单元测试 `TestRenderPlanesHelpers`
  - [ ] 运行测试
  - [ ] Code Review

### 3.2 LayerManager 更新

- [ ] **添加 BuildRenderPlanes() 方法** (`runtime/layer/manager.go`)
  - [ ] ⬜ `BuildRenderPlanes(root *Fiber) *RenderPlanes`
  - [ ] ⬜ 直接从 Fiber 树构建
  - [ ] ⬜ 调用 `renderPlanes.BuildFromFiber(root)`
  - [ ] ⬜ 添加单元测试 `TestLayerManagerBuildRenderPlanes`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **添加 GetRenderPlanes() 方法** (`runtime/layer/manager.go`)
  - [ ] ⬜ `GetRenderPlanes() *RenderPlanes`
  - [ ] ⬜ 返回内部 renderPlanes 引用
  - [ ] ⬜ 运行测试
  - [ ] Code Review

- [ ] **更新 HasModal() 等方法** (`runtime/layer/manager.go`)
  - [ ] ⬜ `HasModal()` 基于 `renderPlanes.HasLayer(LayerModal)`
  - [ ] ⬜ `HasOverlay()` 基于 `renderPlanes.HasLayer(LayerOverlay)`
  - [ ] ⬜ `GetHighestLayer()` 基于 `renderPlanes.GetHighestLayer()`
  - [ ] 添加单元测试 `TestLayerManagerQueries`
  - [ ] 运行测试
  - [ ] Code Review

### 3.3 与旧 API 共存

- [ ] **保持 CollectAndLayout() 可用** (`runtime/layer/manager.go`)
  - [ ] ⬜ 添加 Deprecated 注释
  - [ ] ⬜ 确保仍然可以调用
  - [ ] ⬜ 内部实现可以暂时保持不变（或者转换到新 API）
  - [ ] 添加集成测试 `TestCollectAndLayoutStillWorks`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **提供转换方法**（可选）
  - [ ] ⬜ `ConvertToRenderPlanes(oldLayouts LayerLayouts) *RenderPlanes`
  - [ ] ⬜ 用于逐步迁移旧代码
  - [ ] ⬜ 添加单元测试 `TestConvertToRenderPlanes`
  - [ ] 运行测试
  - [ ] Code Review

### 3.4 测试覆盖

- [ ] **添加 RenderPlanes 单元测试** (`runtime/layer/renderplanes_test.go`)
  - [ ] ⬜ `TestNewRenderPlanes`
  - [ ] ⬜ `TestBuildFromFiberBasic`
  - [ ] ⬜ `TestBuildFromFiberMultipleLayers`
  - [ ] ⬜ `TestSortPlanes`
  - [ ] ⬜ `TestGetPlane`
  - [ ] ⬜ `TestIterate`
  - [ ] `TestRenderPlanesHelpers`

- [ ] **添加 LayerManager 单元测试** (`runtime/layer/manager_test.go`)
  - [ ] ⬜ `TestLayerManagerBuildRenderPlanes`
  - [ ] ⬜ `TestLayerManagerGetRenderPlanes`
  - [ ] ⬜ `TestLayerManagerQueries`

- [ ] **添加集成测试** (`runtime/layer/layer_integration_test.go`)
  - [ ] ⬜ `TestRenderPlanesWithRealApp`
  - [ ] ⬜ `TestRenderPlanesWithModal`
  - [ ] ⬜ `TestRenderPlanesWithOverlay`

### 3.5 性能测试

- [ ] **RenderPlanes 性能基准测试**
  - [ ] ⬜ `BenchmarkBuildRenderPlanes` 基准测试
  - [ ] ⬜ 测试 100、1000、10000 节点的性能
  - [ ] 确保时间复杂度 O(n)
  - [ ] 识别性能瓶颈并优化

### 3.6 文档更新

- [ ] **更新 RenderPlanes 文档** (`runtime/layer/renderplanes.go`)
  - [ ] ⬜ 添加包注释
  - [ ] 添加类型注释
  - [ ] 添加方法注释
  - ] 添加使用示例

- [ ] **更新 LayerManager 文档** (`runtime/layer/manager.go`)
  - [ ] ⬜ 标记旧 API 为 Deprecated
  - [ ] ⬜ 文档化新 API
  - [ ] 添加迁移指南

- [ ] **更新 AGENTS.md**
  - [ ] ⬜ 记录 RenderPlanes 类型
  - [ ] ⬜ 记录新 API

### 3.7 Phase 3 验收

- [ ] ⬜ 所有单元测试通过
- [ ] ⬜ 所有集成测试通过
- [ ] ⬜ 旧 API 仍然可用
- [ ] ⬜ 性能无明显退化
- [ ] ⬜ 测试覆盖率 > 80%
- [ ] ⬜ Code Review 通过

---

## Phase 4: 废弃 StripLayers

**时间**: 1 周
**目标**: 移除 StripLayers 相关代码

### 4.1 标记 Deprecated

- [ ] **标记 StripLayers 为 Deprecated** (`runtime/layer/collector.go`)
  - [ ] ⬜ 添加 `// Deprecated: Use BuildRenderPlanes instead` 注释
  - [ ] ⬜ 添加 Go 1.13+ 的 deprecation marker（如果支持）
  - [ ] ⬜ 运行测试确保不破坏现有调用

- [ ] **标记 cloneWithoutLayers 为 Deprecated** (`runtime/layer/collector.go`)
  - [ ] ⬜ 添加 `// Deprecated: This method is no longer needed` 注释
  - [ ] ⬜ 运行测试

- [ ] **标记 CollectAndLayout 为 Deprecated** (`runtime/layer/manager.go`)
  - [ ] ⬜ 添加 `// Deprecated: Use BuildRenderPlanes instead` 注释
  - [ ] ⬜ 运行测试

### 4.2 查找所有调用点

- [ ] **查找 StripLayers 调用**
  - [ ] ⬜ `grep -r "StripLayers" ./...` 查找所有调用
  - [ ] ⬜ 列出所有文件和行号
  - [ ] ⬜ 评估每个调用点的优先级

- [ ] **查找 cloneWithoutLayers 调用**
  - [ ] ⬜ `grep -r "cloneWithoutLayers" ./...`
  - [ ] ⬜ 列出所有文件和行号

- [ ] **查找 CollectAndLayout 调用**
  - [ ] ⬜ `grep -r "CollectAndLayout" ./...`
  - [ ] ⬜ 列出所有文件和行号

### 4.3 替换调用点

- [ ] **替换 framework/app.go 中的调用**
  - [ ] ⬜ 将 `CollectAndLayout()` 替换为 `BuildRenderPlanes()`
  - [ ] ⬜ 更新相关逻辑
  - [ ] ⬜ 添加集成测试 `TestAppWithRenderPlanes`
  - [ ] ⬜ 运行测试
  - [ ] ⬜ Code Review

- [ ] **替换 examples 中的调用**
  - [ ] ⬜ 遍历所有 examples/ 目录
  - [ ] ⬜ 找到使用 StripLayers 的示例
  - [ ] ⬜ 更新为使用 RenderPlanes
  - [ ] ⬜ 运行每个示例验证

- [ ] **替换测试代码中的调用**
  - [ ] ⬜ 遍历所有 *_test.go 文件
  - [ ] ⬜ 找到使用 StripLayers 的测试
  - [ ] ⬜ 更新为使用 RenderPlanes
  - [ ] ⬜ 运行所有测试

- [ ] **替换 internal/ 中的调用**
  - [ ] ⬜ 遍历 internal/ 目录
  - [ ] ⬜ 找到使用 StripLayers 的内部代码
  - [ ] ⬜ 更新为使用 RenderPlanes
  - [ ] ⬜ 运行测试

### 4.4 移除无用代码

- [ ] **移除 StripLayers 实现** (`runtime/layer/collector.go`)
  - [ ] ⬜ 移除 `StripLayers()` 方法实现
  - [ ] ⬜ 移除 `cloneWithoutLayers()` 方法实现
  - [ ] ⬜ 运行 `go test ./runtime/layer/...`
  - [ ] 确保无编译错误

- [ ] **移除 CollectAndLayout 实现** (`runtime/layer/manager.go`)
  - [ ] ⬜ 移除 `CollectAndLayout()` 方法实现
  - [ ] ⬜ 移除 `layoutLayer()` 方法实现（如果不再需要）
  - [ ] ⬜ 移除 `centerModal()` 方法实现（如果不再需要）
  - [ ] ⬜ 移除 `buildHitMapFromComputedBox()` 方法实现（如果不再需要）
  - [ ] ⬜ 运行 `go test ./runtime/layer/...`
  - [ ] 确保无编译错误

- [ ] **清理 Collector 类型** (`runtime/layer/collector.go`)
  - [ ] ⬜ 评估 Collector 是否还有用途
  - [ ] ⬜ 如果只有 StripLayers 功能，考虑废弃
  - [ ] ⬜ 如果还有其他用途，保留并更新文档

### 4.5 验证功能不变

- [ ] **运行所有测试**
  - [ ] ⬜ `go test ./...`
  - [ ] ⬜ 确保所有测试通过
  - [ ] ⬜ 修复任何失败的测试

- [ ] **运行所有示例**
  - [ ] ⬜ `cd examples/counter && go run .`
  - [ ] ⬜ `cd examples/modal && go run .`
  - [ ] ⬜ `cd examples/overlay && go run .`
  - [ ] ⬜ 确保所有示例正常运行

- [ ] **手动验证渲染**
  - [ ] ⬜ 运行 Modal 示例，验证 Modal 正确显示
  - [ ] ⬜ 运行 Overlay 示例，验证 Overlay 正确显示
  - [ ] ⬜ 运行 Tooltip 示例，验证 Tooltip 正确显示
  - [ ] 记录任何视觉差异

### 4.6 文档更新

- [ ] **更新文档**
  - [ ] ⬜ 从 AGENTS.md 移除 StripLayers 相关内容
  - [ ] ⬜ 从 README 移除 StripLayers 说明
  - [ ] ⬜ 更新 Migration Guide

- [ ] **添加迁移指南**
  - [ ] ⬜ 创建 MIGRATION_TO_RENDERPLANES.md
  - [ ] ⬜ 提供旧 API 到新 API 的映射
  - [ ] ⬜ 提供迁移示例

### 4.7 Phase 4 验收

- [ ] ⬜ 所有 StripLayers 调用已移除
- [ ] ⬜ 所有 CollectAndLayout 调用已移除
- [ ] ⬜ 所有单元测试通过
- [ ] ⬜ 所有集成测试通过
- [ ] ⬜ 所有示例程序正常运行
- [ ] ⬜ 无编译错误
- [ ] ⬜ Code Review 通过

---

## Phase 5: Render 更新

**时间**: 1 周
**目标**: Render 基于 RenderPlanes

### 5.1 FiberRenderer 更新

- [ ] **修改 Render() 方法** (`internal/render/vnode_renderer.go`)
  - [ ] ⬜ 获取 Fiber 根节点
  - [ ] ⬜ 获取 LayerManager
  - [ ] ⬜ 调用 `BuildRenderPlanes(fiberRoot)`
  - [ ] ⬜ 按 Layer 顺序渲染
  - [ ] ⬜ 添加集成测试 `TestRenderWithRenderPlanes`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **实现 renderComputedBox() 方法** (`internal/render/vnode_renderer.go`)
  - [ ] ⬜ `renderComputedBox(box *compute.ComputedBox, buffer interface{})`
  - [ ] ⬜ 基于 `box.VNode` 渲染
  - [ ] ⬜ 递归渲染子节点
  - [ ] ⬜ 添加单元测试 `TestRenderComputedBox`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **实现 renderByRenderPlanes() 方法** (`internal/render/vnode_renderer.go`)
  - [ ] ⬜ `renderByRenderPlanes(renderPlanes *RenderPlanes, buffer interface{})`
  - [ ] ⬜ 遍历 `renderPlanes.RenderOrder`
  - [ ] ⬜ 对每个 Layer 的每个 ComputedBox 调用 `renderComputedBox()`
  - [ ] ⬜ 添加单元测试 `TestRenderByRenderPlanes`
  - [ ] 运行测试
  - [ ] Code Review

### 5.2 更新渲染顺序

- [ ] **验证渲染顺序** (`internal/render/vnode_renderer.go`)
  - [ ] ⬜ LayerBase → LayerOverlay → LayerModal → LayerTooltip → LayerInspector
  - [ ] ⬜ 添加集成测试 `TestRenderOrderCorrect`
  - [ ] ⬜ 运行测试
  - [ ] Code Review

- [ ] **添加调试日志**（可选）
  - [ ] ⬜ 记录每个 Layer 的渲染
  - [ ] ⬜ 记录每个 ComputedBox 的渲染
  - [ ] ⬜ 便于调试渲染问题

### 5.3 测试覆盖

- [ ] **添加 Render 单元测试** (`internal/render/vnode_renderer_test.go`)
  - [ ] ⬜ `TestRenderComputedBox`
  - [ ] ⬜ `TestRenderByRenderPlanes`
  - [ ] ⬜ `TestRenderOrderCorrect`

- [ ] **添加 Render 集成测试** (`internal/render/render_integration_test.go`)
  - [ ] ⬜ `TestRenderWithRealApp`
  - [ ] ⬜ `TestRenderWithModal`
  - [ ] ⬜ `TestRenderWithOverlay`
  - [ ] ⬜ `TestRenderWithTooltip`

### 5.4 视觉测试

- [ ] **运行 Modal 示例**
  - [ ] ⬜ `cd examples/modal && go run .`
  - [ ] ⬜ 验证 Modal 正确显示
  - [ ] ⬜ 验证 Modal 在最上层
  - [ ] ⬜ 记录截图对比

- [ ] **运行 Overlay 示例**
  - [ ] ⬜ `cd examples/overlay && go run .`
  - [ ] ⬜ 验证 Overlay 正确显示
  - [ ] ⬜ 验证 Overlay 在 Modal 之下
  - [ ] 记录截图对比

- [ ] **运行复杂场景**
  - [ ] ⬜ 运行同时有 Modal 和 Overlay 的场景
  - [ ] ⬜ 验证渲染顺序正确
  - [ ] 记录截图对比

### 5.5 性能优化

- [ ] **Render 性能基准测试**
  - [ ] ⬜ `BenchmarkRenderWithRenderPlanes` 基准测试
  - [ ] ⬜ 对比旧 Render 的性能
  - [ ] ⬜ 确保无明显退化
  - [ ] 优化瓶颈

- [ ] **减少重复渲染**
  - [ ] ⬜ 识别可以缓存的渲染结果
  - [ ] ⬜ 实现渲染缓存（可选）

### 5.6 文档更新

- [ ] **更新 Render 文档** (`internal/render/vnode_renderer.go`)
  - [ ] ⬜ 更新 Render() 方法注释
  - [ ] ⬜ 添加渲染顺序说明
  - [ ] 添加使用示例

- [ ] **更新 AGENTS.md**
  - [ ] ⬜ 记录新的渲染流程
  - [ ] ⬜ 记录 RenderPlanes 渲染顺序

### 5.7 Phase 5 验收

- [ ] ⬜ 所有单元测试通过
- [ ] ⬜ 所有集成测试通过
- [ ] ⬜ 渲染顺序正确
- [ ] ⬜ Modal、Overlay、Tooltip 正确显示
- [ ] ⬜ 性能无明显退化
- [ ] ⬜ 测试覆盖率 > 80%
- [ ] ⬜ Code Review 通过

---

## Phase 6: HitMap 更新

**时间**: 1 周
**目标**: HitMap 基于 Fiber 树

### 6.1 移除旧 API

- [ ] **移除 GetMergedHitMap() 方法** (`runtime/layer/manager.go`)
  - [ ] ⬜ 添加 `// Deprecated: Use BuildHitMapFromFiber instead` 注释
  - [ ] ⬜ 查找所有调用点
  - [ ] ⬜ 替换为 `BuildHitMapFromFiber()`
  - [ ] ⬜ 运行测试
  - [ ] Code Review

- [ ] **移除 getLayouts 相关代码** (`runtime/layer/manager.go`)
  - [ ] ⬜ 评估 LayerLayouts 是否还有用途
  - [ ] ⬜ 如果只有 HitMap 用途，移除它
  - [ ] ⬜ 运行测试

### 6.2 更新 HitMap 调用点

- [ ] **更新 App.Render()** (`framework/app.go`)
  - [ ] ⬜ `hitMap := event.BuildHitMapFromFiber(fiberRoot)`
  - [ ] ⬜ 替代 `hitMap := layerManager.GetMergedHitMap()`
  - [ ] ⬜ 添加集成测试 `TestAppWithFiberHitMap`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **查找所有 GetMergedHitMap 调用**
  - [ ] ⬜ `grep -r "GetMergedHitMap" ./...`
  - [ ] ⬜ 列出所有文件和行号
  - [ ] ⬜ 逐个替换为 `BuildHitMapFromFiber()`
  - [ ] 运行测试

### 6.3 验证 HitTest 正确性

- [ ] **HitTest Z-order 排序验证**
  - [ ] ⬜ 创建测试树（包含多个 Layer）
  - [ ] ⬜ 调用 `BuildHitMapFromFiber()`
  - [ ] ⬜ 验证 entries 按 Layer 降序排序
  - [ ] ⬜ 添加单元测试 `TestHitTestZOrdering`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **Modal 优先命中验证**
  - [ ] ⬜ 创建测试树（Base + Modal）
  - [ ] ⬜ 在 Modal 区域调用 HitTest
  - [ ] ⬜ 验证返回 Modal 节点（而非 Base）
  - [ ] ⬜ 添加集成测试 `TestHitTestModalPriority`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **Overlay 优先命中验证**
  - [ ] ⬜ 创建测试树（Base + Overlay）
  - [ ] ⬜ 在 Overlay 区域调用 HitTest
  - [ ] ⬜ 验证返回 Overlay 节点
  - [ ] ⬜ 添加集成测试 `TestHitTestOverlayPriority`
  - [ ] 运行测试
  - [ ] Code Review

### 6.4 验证事件冒泡

- [ ] **事件冒泡测试**
  - [ ] ⬜ 创建包含父子节点的测试树
  - [ ] ⬜ 在子节点触发事件
  - [ ] ⬜ 验证冒泡到父节点
  - [ ] ⬜ 添加集成测试 `TestEventBubbling`
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **Modal 事件冒泡测试**
  - [ ] ⬜ 创建包含 Modal 的测试树
  - [ ] ⬜ 在 Modal 内部触发事件
  - [ ] ⬜ 验证冒泡逻辑（可能需要特殊处理）
  - [ ] 添加集成测试 `TestModalEventBubbling`
  - [ ] 运行测试
  - [ ] Code Review

### 6.5 Modal 点击测试

- [ ] **Modal 点击响应测试**
  - [ ] ⬜ 运行 Modal 示例
  - [ ] ⬜ 点击 Modal 内的按钮
  - [ ] ⬜ 验证按钮响应
  - [ ] 添加用户场景测试
  - [ ] 运行测试
  - [ ] Code Review

- [ ] **Modal 背景点击测试**
  - [ ] ⬜ 运行 Modal 示例
  - [ ] ⬜ 点击 Modal 外部的背景
  - [ ] ⬜ 验证背景不响应（Modal 阻断）
  - [ ] 验证 Modal 不自动关闭（如果需要）
  - [ ] 添加用户场景测试
  - [ ] 运行测试
  - [ ] Code Review

### 6.6 调试支持

- [ ] **添加 HitMap 调试日志**
  - [ ] ⬜ 记录 HitMap 构建过程
  - [ ] ⬜ 记录每个 entry 的 NodeID 和 Layer
  - [ ] ⬜ 记录 sort 结果
  - [ ] 便于调试事件问题

- [ ] **添加 HitTest 调试日志**
  - [ ] ⬜ 记录 HitTest 查找过程
  - [ ] ⬜ 记录命中的 entry
  - [ ] 便于调试事件分发问题

### 6.7 性能测试

- [ ] **BuildHitMapFromFiber 性能基准测试**
  - [ ] ⬜ `BenchmarkBuildHitMapFromFiber` 基准测试
  - [ ] ⬜ 测试不同规模树的性能
  - [ ] 确保时间复杂度 O(n)
  - [ ] 优化瓶颈

- [ ] **HitTest 性能基准测试**
  - [ ] ⬜ `BenchmarkHitTest` 基准测试
  - [ ] 确保命中测试足够快（< 100μs）
  - [ ] 优化查找算法

### 6.8 测试覆盖

- [ ] **添加 HitMap 单元测试** (`runtime/event/hitmap_fiber_test.go`)
  - [ ] ⬜ `TestHitTestZOrdering`
  - [ ] ⬜ `TestHitTestModalPriority`
  - [ ] ⬜ `TestHitTestOverlayPriority`
  - [ ] ⬜ `TestHitTestComplexMultipleLayers`

- [ ] **添加事件集成测试** (`runtime/event/event_integration_test.go`)
  - [ ] ⬜ `TestEventBubbling`
  - [ ] ⬜ `TestModalEventBubbling`
  - [ ] ⬜ `TestMultiLayerEventDispatch`

### 6.9 文档更新

- [ ] **更新 HitMap 文档** (`runtime/event/hitmap.go`)
  - [ ] ⬜ 更新 BuildHitMapFromFiber() 文档
  - [ ] ⬜ 添加 Layer 和 Z-order 说明
  - [ ] 添加使用示例

- [ ] **更新 AGENTS.md**
  - [ ] ⬜ 记录新的 HitMap API
  - [ ] 记录事件分发流程

### 6.10 Phase 6 验收

- [ ] ⬜ 所有单元测试通过
- [ ] ⬜ 所有集成测试通过
- [ ] ⬜ HitTest 正确（优先命中高层）
- [ ] ⬜ 事件冒泡正确
- [ ] ⬜ Modal 点击响应正确
- [ ] ⬜ 性能无明显退化
- [ ] ⬜ 测试覆盖率 > 80%
- [ ] ⬜ Code Review 通过

---

## Phase 7: 清理和优化

**时间**: 1 周
**目标**: 清理废弃代码，优化性能

### 7.1 移除所有 Deprecated 代码

- [ ] **移除 StripLayers 实现** (`runtime/layer/collector.go`)
  - [ ] ⬜ 移除 `StripLayers()` 方法
  - [ ] ⬜ 移除 `cloneWithoutLayers()` 方法
  - [ ] ⬜ 运行 `go test ./...`

- [ ] **移除 CollectAndLayout 实现** (`runtime/layer/manager.go`)
  - [ ] ⬜ 移除 `CollectAndLayout()` 方法
  - [ ] ⬜ 移除相关私有方法
  - [ ] ⬜ 移除 LayerLayouts 类型（如果不再需要）
  - [ ] 运行 `go test ./...`

- [ ] **移除 GetMergedHitMap 实现** (`runtime/layer/manager.go`)
  - [ ] ⬜ 移除 `GetMergedHitMap()` 方法
  - [ ] 运行 `go test ./...`

- [ ] **清理其他 Deprecated 代码**
  - [ ] ⬜ 遍历所有文件
  - [ ] 查找带有 Deprecated 注释的代码
  - [ ] 评估是否可移除
  - [ ] 逐个移除
  - [ ] 运行测试

### 7.2 清理 unused imports

- [ ] **查找 unused imports**
  - [ ] ⬜ 运行 `golangci-lint run ./...`
  - [ ] 查找所有 unused imports
  - [ ] 逐个移除
  - ] 运行测试

### 7.3 清理未使用的代码

- [ ] **查找未使用的类型和函数**
  - [ ] ⬜ 运行 `golangci-lint run ./...`
  - [ ] 查找所有 unused exports
  - [ ] 查找所有 unused private functions
  - [ ] 评估是否可移除
  - [ ] 逐个移除（如果确认无用）
  - [ ] 运行测试

### 7.4 优化 RenderPlanes 性能

- [ ] **分析 RenderPlanes 构建性能**
  - [ ] ⬜ 运行 `BenchmarkBuildRenderPlanes`
  - [ ] 分析热点
  - [ ] 优化以下方面：
    - [ ] ⬜ 减少内存分配
    - [ ] ⬜ 使用更高效的数据结构
    - [ ] ⬜ 减少不必要的拷贝
  - ] 运行基准测试对比
  - ] 目标：性能提升 > 20%

### 7.5 优化 HitMap 性能

- [ ] **分析 BuildHitMapFromFiber 性能**
  - [ ] ⬜ 运行 `BenchmarkBuildHitMapFromFiber`
  - [ ] 分析热点
  - [ ] 优化以下方面：
    - [ ] ⬜ 减少内存分配
    - [ ] ⬜ 优化排序算法
    - [ ] ⬜ 预分配切片容量
  - ] 运行基准测试对比
  - ] 目标：性能提升 > 20%

- [ ] **优化 HitTest 性能**
  - [ ] ⬜ 运行 `BenchmarkHitTest`
  - [ ] 分析热点
  - [ ] 优化以下方面：
    - [ ] ⬜ 使用更高效的查找算法
    - ⬜ 预先排序减少查找时间
    - ⬜ 使用空间换时间（如构建网格索引）
  - ] 运行基准测试对比
  - ] 目标：HitTest < 100μs

### 7.6 添加更多单元测试

- [ ] **补充 Fiber 测试**
  - [ ] ⬜ 遍历 Fiber 的所有方法
  - [ ] ⬜ 为未测试的方法添加单元测试
  - [ ] 达到 80% 覆盖率
  - ] 运行测试

- [ ] **补充 RenderPlanes 测试**
  - [ ] ⬜ 遍历 RenderPlanes 的所有方法
  - [ ] ⬜ 为未测试的方法添加单元测试
  - [ ] 达到 80% 覆盖率
  - ] 运行测试

- [ ] **补充 HitMap 测试**
  - [ ] ⬜ 遍历 HitMap 的所有方法
  - [ ] ⬜ 为未测试的方法添加单元测试
  - [ ] 达到 80% 覆盖率
  - ] 运行测试

### 7.7 更新文档

- [ ] **更新主文档**
  - [ ] ⬜ 更新 README.md
  - [ ] ⬜ 移除所有提到的 StripLayers
  - ] 添加 RenderPlanes 说明
  - ] 添加新的 API 文档

- [ ] **更新 API 文档**
  - [ ] ⬜ 更新 `docs/api/` 下的文档
  - [ ] 废弃的 API 标记为 Deprecated
  - ] 新增的 API 添加文档
  - ] 添加迁移指南

- [ ] **更新 AGENTS.md**
  - [ ] ⬜ 移除 StripLayers 相关内容
  - ] 添加 RenderPlanes 相关内容
  - ] 添加新的数据流说明

- [ ] **更新 Migration Guide**
  - [ ] ⬜ 创建或更新 docs/migration/migration_guide.md
  - ] 提供详细的迁移步骤
  - ] 提供常见问题和解决方案

### 7.8 Code Style 和 Lint

- [ ] **运行 golangci-lint**
  - [ ] ⬜ `golangci-lint run ./...`
  - ] 修复所有 lint 错误
  - ] 修复所有 lint 警告（可选）

- [ ] **运行 go vet**
  - [ ] ⬜ `go vet ./...`
  - ] 修复所有错误
  - ] 修复所有警告（可选）

- [ ] **检查代码风格**
  - [ ] ⬜ 使用 `gofmt` 格式化代码
  - ] 使用 `goimports` 格式化 imports
  [ ] 确保代码风格一致

### 7.9 Phase 7 验收

- [ ] ⬜ 废弃代码全部移除
- [ ] ⬜ 无 unused imports
- [ ] ⬜ 无编译警告
- [ ] ⬜ 无 lint 错误
- [ ] ⬜ 性能提升 > 20%（RenderPlanes 和 HitMap）
- [ ] ⬜ 测试覆盖率 > 80%
- [ ] ⬜ 文档完整
- [ ] ⬜ Code Review 通过

---

## Phase 8: 综合测试

**时间**: 1 周
**目标**: 全面测试，确保功能完整

### 8.1 现有测试

- [ ] **运行所有单元测试**
  - [ ] ⬜ `go test -short ./...`
  - ] 确保所有测试通过
  - ] 记录失败的测试
  - ] 修复失败的测试

- [ ] **运行所有集成测试**
  - [ ] ⬜ `go test ./... -tags=integration`
  - ] 确保所有测试通过
  - ] 记录失败的测试
  - ] 修复失败的测试

### 8.2 添加新的集成测试

- [ ] **端到端测试**
  - [ ] ⬜ 创建端到端测试（完整应用流程）
  - ] 验证：Render → Reconcile → Layout → RenderPlanes → Render → HitMap → Event
  - ] 添加集成测试 `TestEndToEndSimple`
  - ] 添加集成测试 `TestEndToEndWithModal`
  - ] 添加集成测试 `TestEndToEndWithOverlay`
  - ] 运行测试

- [ ] **复杂场景测试**
  - [ ] ⬜ 创建同时包含 Modal、Overlay、Tooltip 的测试
  - ] 验证所有 Layer 正确显示
  - ] 验证 HitTest 正确
  - ] 验证事件正确分发
  [ ] 添加集成测试 `TestComplexMultiLayer`
  [ ] 运行测试

### 8.3 用户场景测试

- [ ] **Counter 示例测试**
  - [ ] ⬜ `cd examples/counter && go run .`
  - ] 验证计数器正常工作
  - ] 验证点击事件正确
  [ ] 记录任何问题

- [ ] **Modal 示例测试**
  - [ ] ⬜ `cd examples/modal && go run .`
  - ] 验证 Modal 正确显示
  [ ] 验证 Modal 按钮点击正确
  - ] 验证 Modal 关闭正确
  [ ] 记录任何问题

- [ ] **Overlay 示例测试**
  - [ ] ⬜ `cd examples/overlay && go run .`
  [ ] 验证 Overlay 正确显示
  ] 验证 Overlay 点击正确
  ] 验证 Overlay 关闭正确
  ] 记录任何问题

- [ ] **Tooltip 示例测试**
  - [ ] ⬜ `cd examples/tooltip && go run .`
  ] 验证 Tooltip 正确显示
  ] 验证 Tooltip 位置正确
  ] 记录任何问题

- [ ] **Table 示例测试**
  - [ ] ⬜ `cd examples/table && go run .`
  ] 验证 Table 正确显示
  ] 验证表头正确
  ] 验证数据正确
  ] 记录任何问题

- [ ] **其他示例测试**
  - [ ] ⬜ 遍历所有 examples/
  ] 逐个运行验证
  ] 记录任何问题

### 8.4 性能测试

- [ ] **整体性能基准测试**
  - [ ] ⬜ `BenchmarkAppRendering` - 完整渲染流程
  ] `BenchmarkAppReconcile` - Reconcile 性能
  ] `BenchmarkAppLayout` - Layout 性能
  ] `BenchmarkAppHitTest` - HitTest 性能
  ] 运行所有基准测试
  ] 对比重构前后的性能
  ] 目标：无性能退化（+10% 以内）

- [ ] **内存泄漏测试**
  - [ ] ⬜ 运行应用长时间
  ] 使用 pprof 监控内存
  ] 确保无内存泄漏
  [ ] 记录内存使用情况

- [ ] **大负载测试**
  - [ ] ⬜ 创建超大组件树（10000+ 节点）
  ] 运行渲染测试
  ] 验证性能可接受
  ] 记录性能数据

### 8.5 回归测试

- [ ] **对比重构前后的行为**
  - [ ] ⬜ 重新提交重构前的代码到分支
  ] 运行相同测试
  ] 对比重构前后结果
  ] 验证无行为变化（除了预期的改进）

- [ ] **验证未引入新 bug**
  - [ ] ⬜ 确保所有旧功能仍然正常
  ] 确保未破坏任何边缘情况
  ] 确保错误处理正确

### 8.6 用户验收测试（UAT）

- [ ] **邀请测试用户**
  - [ ] ⬜ 创建测试版本
  ] 分发给测试用户
  ] 收集反馈
  ] 修复发现的问题

- [ ] **用户报告问题处理**
  - [ ] ⬜ 记录所有用户报告的问题
  [ ] 分析问题原因
  [ ] 修复问题
  [ ] 回归测试

### 8.7 压力测试

- [ ] **大规模并发测试**
  - [ ] ⬜ 创建多个并发渲染请求
  ] 验证系统稳定性
  ] 验证无数据竞争
  ] 使用 `go test -race` 检测

- [ ] **长时间运行测试**
  - [ ] ⬜ 运行应用 24 小时
  ] 验证无内存泄漏
  ] 验证性能稳定
  ] 记录性能数据

### 8.8 文档和示例更新

- [ ] **更新所有示例**
  - [ ] ⬜ 遍历所有 examples/
  ] 更新到新的 API
  ] 运行验证
  [ ] 更新 README

- [ ] **创建新的示例**
  - [ ] ⬜ 创建 `examples/multilayer/` 示例
  ] 展示多个 Layer 的使用
  ] 添加 README
  [ ] 运行验证

### 8.9 发布准备

- [ ] **更新 CHANGELOG**
  - [ ] ⬜ 记录所有变更
  [ ] 记录破坏性变更
  [ ] 记录新功能
  [ ] 记录 bug 修复

- [ ] **更新版本号**
  - [ ] ⬜ 更新 go.mod 中的版本
  [ ] 更新 README 中的版本
  [ ] 标记为 Breaking Change

- [ ] **创建 Release Notes**
  - [ ] ⬜ 创建详细的 Release Notes
  [ ] 包含迁移指南
  [ ] 包含已知问题
  [ ] 包含升级路径

### 8.10 Phase 8 验收

- [ ] ⬜ 所有单元测试通过
- [ ] ⬜ 所有集成测试通过
- [ ] ⬜ 所有示例程序正常运行
- [ ] ⬜ 所有用户场景验证通过
- [ ] ⬜ 性能无退化
- [ ] ⬜ 无内存泄漏
- [ ] ⬜ 无数据竞争
- [ ] ⬜ 文档完整
- [ ] ⬜ Code Review 通过
- [ ] ⬜ 准备发布

---

## 总体验收标准

### 功能验收

- [ ] ⬜ 所有核心功能正常（Render、Reconcile、Layout、Event）
- [ ] ⬜ 多层渲染正确（Modal、Overlay、Tooltip）
- [ ] ⬜ HitTest 正确（优先命中高层）
- [ ] ⬜ 事件冒泡正确

### 性能验收

- [ ] ⬜ 渲染性能：无退化（+10% 以内）
- [ ] ⬜ Layout 性能：无退化（+10% 以内）
- [ ] ⬜ HitTest 性能：< 100μs
- [ ] ⬜ 内存使用：无明显增加（+10% 以内）

### 质量验收

- [ ] ⬜ 测试覆盖率：> 80%
- [ ] ⬜ 无 lint 错误
- [ ] ⬜ 无编译警告
- [ ] ⬜ 无数据竞争

### 文档验收

- [ ] ⬜ API 文档完整
- [ ] ⬜ 迁移指南完整
- [ ] ⬜ 示例程序更新
- [ ] ⬜ CHANGELOG 更新

---

## 附录

### A. 任务优先级

| 优先级 | 说明 |
|--------|------|
| P0 | 核心功能，必须完成 |
| P1 | 重要功能，强烈建议完成 |
| P2 | 次要功能，建议完成 |
| P3 | 可选功能，时间允许时完成 |

### B. 任务依赖关系

```
Phase 1 (基础设施)
  ↓
Phase 2 (Layout)
  ↓
Phase 3 (RenderPlane)
  ↓
Phase 4 (废弃 StripLayers)
    ↓
    ├─→ Phase 5 (Render)
    ├─→ Phase 6 (HitMap)
    │     ↓
    ├─────────────┘
    ↓
Phase 7 (清理和优化)
  ↓
Phase 8 (综合测试)
```

### C. 风险任务标记

标记以下任务为高风险：
- ⚠️ Phase 2: Layout 重构（影响所有布局）
- ⚠️ Phase 4: 废弃 StripLayers（大量代码变更）
- ⚠️ Phase 6: HitMap 更新（影响事件系统）

高风险任务需要：
- 更详细的测试
- Code Review
- 逐步迁移策略

---

**文档版本**: 1.0
**最后更新**: 2026-02-14
**责任人**: [待分配]
**审核人**: [待指定]
