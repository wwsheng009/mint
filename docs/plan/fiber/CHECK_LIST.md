# Fiber 统一架构重构 - 验收检查清单

**版本**: 1.0
**日期**: 2026-02-14
**负责人**: [待分配]
**状态**: 进行中

---

## 📋 目录

1. [通用检查项](#通用检查项)
2. [Phase 1: 基础设施](#phase-1-基础设施)
3. [Phase 2: Layout 重构](#phase-2-layout-重构)
4. [Phase 3: RenderPlane 引入](#phase-3-renderplane-引入)
5. [Phase 4: 废弃 StripLayers](#phase-4-废弃-striplayers)
6. [Phase 5: Render 更新](#phase-5-render-更新)
7. [Phase 6: HitMap 更新](#phase-6-hitmap-更新)
8. [Phase 7: 清理和优化](#phase-7-清理和优化)
9. [Phase 8: 综合测试](#phase-8-综合测试)
10. [最终验收](#最终验收)
11. [回滚决策](#回滚决策)
12. [签收模板](#签收模板)

---

## 通用检查项

### 所有阶段必须执行的检查

| 检查项 | 验证方法 | 责任人 | 完成时间 |
|--------|---------|--------|----------|
| ✅ 代码编译通过 | `go build ./...` | 开发者 | 阶段完成后 |
| ✅ 所有单元测试通过 | `go test ./...` | 开发者 | 阶段完成后 |
| ✅ 无新的编译警告 | 检查编译输出 | 开发者 | 阶段完成后 |
| ✅ 代码符合项目风格 | 检查代码格式 | 开发者 | 阶段完成后 |
| ✅ 添加了必要的单元测试 | 代码审查 | Reviewer | 阶段完成后 |
| ✅ 更新了相关文档 | 检查文档更新 | 开发者 | 阶段完成后 |
| ✅ 性能无显著退化 | 运行性能测试 | 开发者 | 阶段完成后 |
| ✅ 无内存泄漏 | 运行 `go test -race` | 开发者 | 阶段完成后 |

### 代码审查清单

| 检查项 | 标准 |
|--------|------|
| ✅ 代码清晰易懂 | 变量命名准确，逻辑清晰 |
| ✅ 错误处理完整 | 所有错误路径都有处理 |
| ✅ 边界情况考虑完整 | 空/nil/极端情况有处理 |
| ✅ 并发安全正确 | 共享数据正确加锁 |
| ✅ API 设计合理 | 函数签名清晰，易于使用 |
| ✅ 测试覆盖充分 | 核心逻辑有测试覆盖 |
| ✅ 文档注释完整 | 导出函数/类型有注释 |

---

## Phase 1: 基础设施

### ⏰ 预检查清单

**开始 Phase 1 前必须验证以下项：**

- [ ] ✅ 理解 Fiber 和 VNode 的区分原则（见 `diff_key.md`）
- [ ] ✅ 理解 Layer 是渲染维度而非结构维度（见 `diff_layer.md`）
- [ ] ✅ 阅读 `runtime/ui/fiber.go` 了解当前 Fiber 结构
- [ ] ✅ 阅读 `runtime/compute/box.go` 了解当前 ComputedBox 结构
- [ ] ✅ 准备好分支：`feature/fiber-layer-refactor-phase1`
- [ ] ✅ 记录当前测试基线：`go test ./... -cover` 保存输出
- [ ] ✅ 记录当前性能基线：`go test -bench=. ...` 保存输出

### ✅ Phase 1 完成检查清单

**所有子任务完成后，按以下清单验收：**

#### 1.1 Fiber 结构更新

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] Fiber.Layer 字段存在 | `grep "Layer rtui.Layer" runtime/ui/fiber.go` | 找到定义 |
| [ ] Fiber.Layer 字段有注释 | View fiber.go 文件 | 有清晰的中文注释 |
| [ ] 编译通过 | `go build ./runtime/ui/...` | 无错误 |
| [ ] 单元测试通过 | `go test ./runtime/ui/... -run Fiber` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] Fiber.ComputedBox 字段存在 | `grep "ComputedBox \*compute.ComputedBox" runtime/ui/fiber.go` | 找到定义 |
| [ ] Fiber.ComputedBox 字段有注释 | View fiber.go 文件 | 有清晰的中文注释 |
| [ ] 编译通过 | `go build ./runtime/ui/...` | 无错误 |
| [ ] 单元测试通过 | `go test ./runtime/ui/... -run Fiber` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 1.2 Fiber 构造函数更新

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] CreateFiber 从 VNode 拷贝 Layer | 查看 CreateFiber 实现 | 有 `fiber.Layer = vnode.GetLayer()` |
| [ ] CreateFiber 处理 Layer 默认值 | 查看 CreateFiber 实现 | 无效值设为 LayerBase |
| [ ] CreateFiber 初始化 ComputedBox | 查看 CreateFiber 实现 | `fiber.ComputedBox == nil` |
| [ ] TestCreateFiberLayerCopy 存在 | 查看 fiber*_test.go | 测试存在 |
| [ ] TestCreateFiberLayerDefaultValue 存在 | 查看 fiber*_test.go | 测试存在 |
| [ ] 运行测试 | `go test ./runtime/ui/... -run TestCreateFiber` | 全部通过 |

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] CloneFiber 拷贝 Layer | 查看 CloneFiber 实现 | 有 `clone.Layer = current.Layer` |
| [ ] CloneFiber 重置 ComputedBox | 查看 CloneFiber 实现 | `clone.ComputedBox == nil` |
| [ ] CloneFiber 保持 NodeID | 查看 CloneFiber 实现 | `clone.NodeID == current.NodeID` |
| [ ] TestCloneFiberLayerPreserved 存在 | 查看 fiber*_test.go | 测试存在 |
| [ ] TestCloneFiberNodeIDPreserved 存在 | 查看 fiber*_test.go | 测试存在 |
| [ ] 运行测试 | `go test ./runtime/ui/... -run TestCloneFiber` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 1.3 Reconciler 更新

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] CompleteWork 拷贝 Layer | 查看 complete_work.go | 有 `workInProgress.Layer = workInProgress.VNode.GetLayer()` |
| [ ] shouldUpdate 仍用 DiffKey | 查看 diff.go | 使用 DiffKey 匹配 |
| [ ] reconcileChildren 仍用 DiffKey | 查看 reconciler.go | 使用 DiffKey 匹配 |
| [ ] TestCompleteWorkLayerCopied 存在 | 查看 complete_work_test.go | 测试存在 |
| [ ] TestReconcileLayerPreserved 存在 | 查看 reconciler_test.go | 测试存在 |
| [ ] 运行 reconciler 测试 | `go test ./internal/reconciler/...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 1.4 ComputedBox 结构更新

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] ComputedBox.NodeID 字段存在 | grep "NodeID uint64" runtime/compute/box.go | 找到定义 |
| [ ] ComputedBox.NodeID 有注释 | View box.go | 有清晰注释 |
| [ ] ComputedBox.Layer 字段存在 | grep "Layer rtui.Layer" runtime/compute/box.go | 找到定义 |
| [ ] ComputedBox.Layer 有注释 | View box.go | 有清晰注释 |
| [ ] 编译通过 | `go build ./runtime/compute/...` | 无错误 |
| [ ] 单元测试通过 | `go test ./runtime/compute/... -run ComputedBox` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 1.5 BuildHitMapFromFiber API

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] BuildHitMapFromFiber 函数存在 | grep "func BuildHitMapFromFiber" runtime/event/hitmap.go | 找到定义 |
| [ ] 接受 Fiber 作为输入 | 查看函数签名 | 参数类型为 `*Fiber` 或类似 |
| [ ] 返回 HitMap | 查看函数签名 | 返回类型为 `*HitMap` |
| [ ] TestBuildHitMapFromFiber 存在 | 查看 hitmap_test.go | 测试存在 |
| [ ] 测试覆盖空树 | 测试代码 | 有空 Fiber 测试用例 |
| [ ] 测试覆盖单节点 | 测试代码 | 有单节点 Fiber 测试用例 |
| [ ] 测试覆盖多节点 | 测试代码 | 有多节点 Fiber 测试用例 |
| [ ] 测试覆盖多个层级 | 测试代码 | 有多层 Fiber 测试用例 |
| [ ] 运行测试 | `go test ./runtime/event/... -run HitMap` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

### 📊 Phase 1 性能检查

| 指标 | 测试方法 | 允许偏差 |
|------|---------|---------|
| [ ] 编译时间无显著增加 | 计时编译 | < 5% |
| [ ] 内存占用无显著增加 | `runtime.ReadMemStats` | < 5% |
| [ ] Fiber 创建速度无显著下降 | BenchmarkCreateFiber | < 5% |
| [ ] Fiber 克隆速度无显著下降 | BenchmarkCloneFiber | < 5% |

### 🚨 Phase 1 阻塞风险

如果以下任何项为 ❌，需要先解决再继续：

- [ ] ❌ 编译失败 → 检查语法错误
- [ ] ❌ 基础测试失败 → 修复测试或代码
- [ ] ❌ 性能退化超过 5% → 优化实现
- [ ] ❌ Review 不通过 → 根据反馈修改

---

## Phase 2: Layout 重构

### ⏰ 预检查清单

**开始 Phase 2 前必须验证以下项：**

- [ ] ✅ Phase 1 已完成并通过验收
- [ ] ✅ 理解当前的 Layout 流程（`runtime/layout/engine.go`）
- [ ] ✅ 理解 Layout 如何遍历 VNode 树
- [ ] ✅ 准备好分支：`feature/fiber-layer-refactor-phase2`
- [ ] ✅ 记录当前 Layout 性能基线
- [ ] ✅ 了解所有 Layout 的调用位置

### ✅ Phase 2 完成检查清单

#### 2.1 Engine.Layout() 更新

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 签名从 Layout(root *VNode) 改为 Layout(root *Fiber) | 查看 engine.go 新签名 | 参数为 Fiber |
| [ ] 内部调用 layoutFiber(root) | 查看 Layout 实现 | 调用新函数 |
| [ ] 废弃旧签名（添加 Deprecated 注释） | grep "Deprecated" engine.go | 有注释 |
| [ ] 编译通过 | `go build ./runtime/layout/...` | 无错误 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 2.2 layoutFiber() 实现

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] layoutFiber 函数存在 | grep "func layoutFiber" engine.go | 找到定义 |
| [ ] 接受 Fiber 作为输入 | 查看函数签名 | 参数为 `*Fiber` |
| [ ] 遍历 Fiber 树（而非 VNode） | 查看实现逻辑 | 使用 fiber.FirstChild/fiber.Sibling |
| [ ] 调用 measureFiber() | 查看实现 | 有 measureFiber 调用 |
| [ ] 返回 Fiber 树（每个 Fiber 有 ComputedBox） | 查看返回值 | 返回更新后的 Fiber |
| [ ] TestLayoutFiber 存在 | 查看 engine_test.go | 测试存在 |
| [ ] 测试覆盖空树 | 测试代码 | 有空 Fiber 测试 |
| [ ] 测试覆盖单节点 | 测试代码 | 有单节点测试 |
| [ ] 测试覆盖 VStack 布局 | 测试代码 | 有 VStack 测试 |
| [ ] 测试覆盖 HStack 布局 | 测试代码 | 有 HStack 测试 |
| [ ] 测试覆盖嵌套布局 | 测试代码 | 有嵌套测试 |
| [ ] 运行测试 | `go test ./runtime/layout/...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 2.3 measureFiber() 实现

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] measureFiber 函数存在 | grep "func measureFiber" engine.go | 找到定义 |
| [ ] 接受 Fiber 作为输入 | 查看函数签名 | 参数为 `*Fiber` |
| [ ] 创建/更新 fiber.ComputedBox | 查看实现 | `fiber.ComputedBox != nil` |
| [ ] ComputedBox.NodeID 填充 | 查看实现 | `fiber.ComputedBox.NodeID == fiber.NodeID` |
| [ ] ComputedBox.Layer 填充 | 查看实现 | `fiber.ComputedBox.Layer == fiber.Layer` |
| [ ] 递归 measure 子节点 | 查看实现 | 遍历 FirstChild |
| [ ] TestMeasureFiberNodeID 存在 | 查看测试文件 | 测试存在 |
| [ ] TestMeasureFiberLayer 存在 | 查看测试文件 | 测试存在 |
| [ ] 运行测试 | `go test ./runtime/layout/...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 2.4 buildHitMapFromFiber() 实现

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] buildHitMapFromFiber 函数存在 | grep "func buildHitMapFromFiber" engine.go | 找到定义 |
| [ ] 遍历 Fiber 树 | 查看实现 | 遍历 FirstChild/Sibling |
| [ ] 从 fiber.ComputedBox 构建 HitMapEntry | 查看实现 | 使用 ComputedBox 数据 |
| [ ] HitMapEntry.NodeID 正确填充 | 查看实现 | 来自 ComputedBox.NodeID |
| [ ] HitMapEntry.Layer 正确填充 | 查看实现 | 来自 ComputedBox.Layer |
| [ ] 按 Layer 排序（Z-order） | 查看实现 | 有排序逻辑 |
| [ ] TestBuildHitMapFromFiber 存在 | 查看测试文件 | 测试存在 |
| [ ] 测试覆盖单层级 | 测试代码 | 有单层测试 |
| [ ] 测试覆盖多层级 Base/Overlay | 测试代码 | 有多层测试 |
| [ ] 测试覆盖 Modal | 测试代码 | 有 Modal 测试 |
| [ ] 运行测试 | `go test ./runtime/layout/...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 2.5 更新所有 Layout 调用位置

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] framework/app.go 更新到新 API | grep "Layout" framework/app.go | 调用新 API |
| [ ] framework/component.go 更新到新 API | grep "Layout" framework/component.go | 调用新 API |
| [ ] devtools/* 更新到新 API | grep "Layout" devtools/** | 调用新 API |
| [ ] examples/* 更新到新 API（如有） | grep "Layout" examples/** | 调用新 API |
| [ ] 所有旧 API 调用已移除 | grep -r "Layout.*VNode" . | 无结果（或只有注释） |
| [ ] 编译通过 | `go build ./...` | 无错误 |
| [ ] 运行所有测试 | `go test ./...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 2.6 Modal/Overlay Layout 验证

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] Modal 组件仍然正常布局 | 运行 Modal 示例 | 可以正确显示 |
| [ ] Overlay 组件仍然正常布局 | 运行 Overlay 示例 | 可以正确显示 |
| [ ] Modal Layer 正确识别 | 查看调试输出 | Layer == LayerModal |
| [ ] Overlay Layer 正确识别 | 查看调试输出 | Layer == LayerOverlay |
| [ ] Modal 在 Overlay 之上 | 查看渲染顺序 | Z-order 正确 |
| [ ] 运行示例测试 | `go run examples/modal` | 无错误 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

### 📊 Phase 2 性能检查

| 指标 | 测试方法 | 允许偏差 |
|------|---------|---------|
| [ ] 布局计算时间无显著增加 | BenchmarkLayout | < 10% |
| [ ] HitMap 构建时间无显著增加 | BenchmarkBuildHitMap | < 10% |
| [ ] 内存占用无显著增加 | `runtime.ReadMemStats` | < 10% |

### 🔬 Phase 2 集成测试

| 测试场景 | 验证方法 | 预期结果 |
|---------|---------|---------|
| [ ] 简单页面布局 | 运行 counter 示例 | 正常显示 |
| [ ] 复杂嵌套布局 | 运行复杂组件示例 | 正确布局 |
| [ ] Modal 弹框测试 | 运行 modal 示例 | 正确弹出 |
| [ ] Overlay 测试 | 运行 overlay 示例 | 正确覆盖 |
| [ ] 键盘导航测试 | 手动测试示例 | 可以导航 |

### 🚨 Phase 2 阻塞风险

- [ ] ❌ 编译失败 → 检查所有调用位置
- [ ] ❌ 布局结果错误 → 调试 measureFiber 逻辑
- [ ] ❌ Modal/Overlay 不显示 → 检查 Layer 填充
- [ ] ❌ 性能退化超过 10% → 优化遍历逻辑

---

## Phase 3: RenderPlane 引入

### ⏰ 预检查清单

**开始 Phase 3 前必须验证以下项：**

- [ ] ✅ Phase 2 已完成并通过验收
- [ ] ✅ 理解 RenderPlane 概念（见 `diff_layer.md`）
- [ ] ✅ 理解旧的 LayerManager 实现
- [ ] ✅ 准备好分支：`feature/fiber-layer-refactor-phase3`
- [ ] ✅ 了解 RenderPlanes 类型设计

### ✅ Phase 3 完成检查清单

#### 3.1 RenderPlanes 类型实现

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] RenderPlanes 类型定义存在 | grep "type RenderPlanes" runtime/render/planes.go | 找到定义 |
| [ ] 包含 Base, Overlay, Modal, Tooltip, Inspector | 查看类型定义 | 5 个字段存在 |
| [ ] 每个字段是 []*compute.ComputedBox | 查看字段类型 | 正确类型 |
| [ ] BuildFromFiber() 方法存在 | grep "func.*BuildFromFiber" planes.go | 找到方法 |
| [ ] BuildFromFiber 遍历一次 Fiber 树 | 查看实现 | 单次遍历 |
| [ ] BuildFromFiber 按 Layer 分桶 | 查看实现 | 有分桶逻辑 |
| [ ] 编译通过 | `go build ./runtime/render/...` | 无错误 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 3.2 RenderPlanes 单元测试

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] TestRenderPlanesBuildFromFiber 存在 | 查看测试文件 | 测试存在 |
| [ ] 测试单节点 Base | 测试代码 | Base 长度 1 |
| [ ] 测试多节点混合层 | 测试代码 | 每层都有数据 |
| [ ] 测试空 Fiber | 测试代码 | 所有层为空 |
| [ ] 测试未设置 Layer 的节点 | 测试代码 | 归到 Base |
| [ ] 运行测试 | `go test ./runtime/render/...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 3.3 LayerManager 更新

| 检查项 | 验证方法 | 预期结果 |
| [ ] BuildRenderPlanes() 方法存在 | grep "func.*BuildRenderPlanes" runtime/layer/manager.go | 找到方法 |
| [ ] BuildRenderPlanes 接受 Fiber | 查看函数签名 | 参数为 Fiber |
| [ ] BuildRenderPlanes 返回 RenderPlanes | 查看函数签名 | 返回 RenderPlanes |
| [ ] 旧方法 StripLayers() 已标记 Deprecated | grep "Deprecated" layer/collector.go | 有注释 |
| [ ] 编译通过 | `go build ./runtime/layer/...` | 无错误 |
| [ ] 运行测试 | `go test ./runtime/layer/...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 3.4 共存验证

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 旧 API StripLayers() 仍然可用 | grep "func StripLayers" layer/collector.go | 函数存在 |
| [ ] 新 API BuildRenderPlanes() 可用 | grep "func.*BuildRenderPlanes" | 函数存在 |
| [ ] 两套 API 可以同时编译 | `go build ./...` | 无冲突 |
| [ ] 运行所有测试 | `go test ./...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

### 📊 Phase 3 性能检查

| 指标 | 测试方法 | 允许偏差 |
|------|---------|---------|
| [ ] RenderPlanes 构建时间可接受 | BenchmarkBuildRenderPlanes | < 旧 Layout 的 50% |
| [ ] 内存占用可接受 | 内存分析 | 与 StripLayers 相当 |

### 🚨 Phase 3 阻塞风险

- [ ] ❌ RenderPlanes 构建错误 → 检查分桶逻辑
- [ ] ❌ 与旧 API 冲突 → 检查命名冲突
- [ ] ❌ 性能不达标 → 优化遍历逻辑

---

## Phase 4: 废弃 StripLayers

### ⏰ 预检查清单

**开始 Phase 4 前必须验证以下项：**

- [ ] ✅ Phase 3 已完成并通过验收
- [ ] ✅ 确定所有 StripLayers 的调用位置
- [ ] ✅ 准备好替换代码
- [ ] ✅ 准备好分支：`feature/fiber-layer-refactor-phase4`

### ✅ Phase 4 完成检查清单

#### 4.1 标记 Deprecated

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] StripLayers() 添加 Deprecated 注释 | 查看 collector.go | 有注释 |
| [ ] cloneWithoutLayers() 添加 Deprecated 注释 | 查看 collector.go | 有注释 |
| [ ] CollectAndLayout() 添加 Deprecated 注释 | 查看 manager.go | 有注释 |
| [ ] 注释说明原因和替代方案 | 查看注释内容 | 清晰说明 |
| [ ] Deprecated 注释符合 Go 规范 | 检查注释格式 | 正确格式 |

#### 4.2 替换所有调用位置

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 找到所有 StripLayers 调用 | `grep -r "StripLayers" . --include="*.go"` | 列出所有位置 |
| [ ] 替换为 BuildRenderPlanes | 代码检查 | 无旧调用 |
| [ ] framework/ 调用已替换 | 检查 framework/*.go | 使用新 API |
| [ ] devtools/ 调用已替换 | 检查 devtools/*.go | 使用新 API |
| [ ] examples/ 调用已替换 | 检查 examples/*.go | 使用新 API |
| [ ] 编译通过 | `go build ./...` | 无错误 |
| [ ] 运行所有测试 | `go test ./...` | 全部通过 |

#### 4.3 移除实现

| 检查项 | 验证方法 | 预期结果 |
| [ ] 移除 StripLayers() 实现 | grep -c "func StripLayers" collector.go | 0 |
| [ ] 移除 cloneWithoutLayers() 实现 | grep -c "func cloneWithoutLayers" collector.go | 0 |
| [ ] 移除 CollectAndLayout() 实现 | grep -c "func CollectAndLayout" manager.go | 0 |
| [ ] 移除辅助函数（如有） | 检查未使用的函数 | 清理完成 |
| [ ] 编译通过 | `go build ./...` | 无错误 |
| [ ] 运行所有测试 | `go test ./...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 4.4 功能验证

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 运行所有示例 | 逐个运行 examples/ | 正常运行 |
| [ ] Modal 示例正常 | `go run examples/modal` | 正常显示 |
| [ ] Overlay 示例正常 | `go run examples/overlay` | 正常显示 |
| [ ] 复杂 UI 示例正常 | 运行复杂示例 | 正常显示 |

### 🚨 Phase 4 阻塞风险

- [ ] ❌ 漏掉某个调用位置 → 全面代码搜索
- [ ] ❌ 删除后某些功能失效 → 回退检查
- [ ] ❌ 示例无法运行 → 调式检查

---

## Phase 5: Render 更新

### ⏰ 预检查清单

**开始 Phase 5 前必须验证以下项：**

- [ ] ✅ Phase 4 已完成并通过验收
- [ ] ✅ 理解 FiberRenderer 当前实现
- [ ] ✅ 理解渲染顺序规则（Base → Overlay → Modal → Tooltip → Inspector）
- [ ] ✅ 准备好分支：`feature/fiber-layer-refactor-phase5`

### ✅ Phase 5 完成检查清单

#### 5.1 FiberRenderer.Render() 更新

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] Render() 接受 RenderPlanes | 查看函数签名 | 参数包含 RenderPlanes |
| [ ] 按层顺序渲染（0-4） | 查看实现逻辑 | LayerBase → LayerInspector |
| [ ] 每层使用 renderComputedBox() | 查看实现 | 有调用 |
| [ ] renderComputedBox() 存在 | grep "func renderComputedBox" renderer 实现 | 找到函数 |
| [ ] renderComputedBox() 接受 ComputedBox | 查看函数签名 | 参数正确 |
| [ ] NonFiberRenderer 已废弃 | grep "Deprecated" renderer 实现 | 有注释 |
| [ ] 编译通过 | `go build ./internal/render/...` | 无错误 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 5.2 renderComputedBox() 实现

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 读取 ComputedBox.BoxModel | 查看实现 | 正确读取 |
| [ ] 读取 ComputedBox.Content | 查看实现 | 正确读取 |
| [ ] 读取 ComputedBox.Style | 查看实现 | 正确读取 |
| [ ] 调用底层渲染 API | 查看实现 | 正确调用 |
| [ ] 处理边界情况（nil） | 查看实现 | 有判断 |
| [ ] 单元测试存在 | 查看 renderer_test.go | 测试存在 |
| [ ] 运行测试 | `go test ./internal/render/...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 5.3 渲染顺序验证

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] Base 层最先渲染 | 手动测试/查看输出 | 正确 |
| [ ] Overlay 在 Base 之上 | 手动测试 Modal/Overlay | 正确 |
| [ ] Modal 在 Overlay 之上 | 手动测试 Modal | 正确 |
| [ ] Tooltip 在 Modal 之上 | 手动测试 Tooltip | 正确 |
| [ ] Inspector 在最上层 | 手动测试 DevTools | 正确 |
| [ ] 运行视觉测试 | 对比截图 | 无差异 |

#### 5.4 视觉测试

| 测试场景 | 验证方法 | 预期结果 |
|---------|---------|---------|
| [ ] 简单文本渲染 | 运行文本示例 | 正确显示 |
| [ ] 按钮渲染 | 运行按钮示例 | 正确显示 |
| [ ] Modal 弹框 | 运行 Modal 示例 | 正确显示，覆盖下方 |
| [ ] Dropdown/Overlay | 运行 Overlay 示例 | 正确显示，覆盖下方 |
| [ ] Tooltip 提示 | 运行 Tooltip 示例 | 正确显示，在 Modal 之上 |
| [ ] DevTools Inspector | 打开 DevTools | 正确显示，在最上层 |

### 📊 Phase 5 性能检查

| 指标 | 测试方法 | 允许偏差 |
|------|---------|---------|
| [ ] 渲染时间无显著增加 | BenchmarkRender | < 10% |
| [ ] 重绘时间可接受 | 手动测试响应 | < 16ms (60fps) |

### 🚨 Phase 5 阻塞风险

- [ ] ❌ 渲染顺序错乱 → 检查循环顺序
- [ ] ❌ 某些内容不显示 → 检查 renderComputedBox 实现
- [ ] ❌ 性能退化 → 优化渲染逻辑

---

## Phase 6: HitMap 更新

### ⏰ 预检查清单

**开始 Phase 6 前必须验证以下项：**

- [ ] ✅ Phase 5 已完成并通过验收
- [ ] ✅ 理解 HitMap 用途（事件命中测试）
- [ ] ✅ 理解 Z-order 对 HitTest 的影响
- [ ] ✅ 找到所有 HitMap 的调用位置
- [ ] ✅ 准备好分支：`feature/fiber-layer-refactor-phase6`

### ✅ Phase 6 完成检查清单

#### 6.1 移除 GetMergedHitMap()

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] GetMergedHitMap() 已移除 | grep "GetMergedHitMap" runtime/event/hitmap.go | 无结果 |
| [ ] 移除相关的合并逻辑 | 检查 hitmap.go | 清理完成 |
| [ ] 编译通过 | `go build ./runtime/event/...` | 无错误 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 6.2 更新所有 HitMap 调用位置

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] framework/app.go 使用 BuildHitMapFromFiber | 检查代码 | 调用新 API |
| [ ] devtools/* 使用 BuildHitMapFromFiber | 检查代码 | 调用新 API |
| [ ] examples/* 使用 BuildHitMapFromFiber | 检查代码 | 调用新 API |
| [ ] 无 GetMergedHitMap 调用 | `grep -r "GetMergedHitMap" .` | 无结果 |
| [ ] 编译通过 | `go build ./...` | 无错误 |
| [ ] 运行所有测试 | `go test ./...` | 全部通过 |

#### 6.3 HitTest Z-ordering 验证

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] HitTest 优先返回上层元素 | 手动测试点击 | 正确命中 |
| [ ] Base 层元素可点击 | 手动测试 Base | 可以点击 |
| [ ] Modal 阻挡 Base 点击 | 手动测试 Modal | 阻挡成功 |
| [ ] Overlay 阻挡 Base 点击 | 手动测试 Overlay | 阻挡成功 |
| [ ] Tooltip 可点击且不影响下层 | 手动测试 Tooltip | 可以点击 |
| [ ] Inspector 可点击 | 手动测试 DevTools | 可以点击 |

#### 6.4 Modal/Overlay HitTest 验证

| 测试场景 | 验证方法 | 预期结果 |
|---------|---------|---------|
| [ ] 点击 Modal 外部不触发 | 手动测试 Modal | 正确 |
| [ ] 点击 Modal 内部按钮 | 手动测试 | 触发按钮 |
| [ ] 点击 Overlay 外部不触发 | 手动测试 Overlay | 正确 |
| [ ] 点击 Overlay 内部元素 | 手动测试 | 触发元素 |
| [ ] 多层 Modal 命中正确 | 手动测试嵌套 | 最上层命中 |

#### 6.5 事件冒泡验证

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 事件在正解层级冒泡 | 手动测试 | 正确冒泡 |
| [ ] StopPropagation 有效 | 手动测试 | 停止冒泡 |
| [ ] 跨层事件不冒泡 | 手动测试 Modal | 正确 |
| [ ] 单元测试存在 | 查看 event_test.go | 测试存在 |
| [ ] 运行测试 | `go test ./runtime/event/...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

### 📊 Phase 6 性能检查

| 指标 | 测试方法 | 允许偏差 |
|------|---------|---------|
| [ ] HitMap 构建时间 | BenchmarkBuildHitMap | < 5ms (1000节点) |
| [ ] HitTest 时间 | BenchmarkHitTest | < 1ms |

### 🚨 Phase 6 阻塞风险

- [ ] ❌ HitMap 构建错误 → 检查 BuildHitMapFromFiber
- [ ] ❌ HitTest 返回错误元素 → 检查排序逻辑
- [ ] ❌ 事件不冒泡 → 检查事件系统

---

## Phase 7: 清理和优化

### ⏰ 预检查清单

**开始 Phase 7 前必须验证以下项：**

- [ ] ✅ Phase 6 已完成并通过验收
- [ ] ✅ 所有功能正常工作
- [ ] ✅ 准备好分支：`feature/fiber-layer-refactor-phase7`

### ✅ Phase 7 完成检查清单

#### 7.1 移除所有 Deprecated 代码

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 搜索所有 Deprecated 注释 | `grep -r "Deprecated:" . --include="*.go"` | 列出所有 |
| [ ] 移除非必须的 Deprecated 函数 | 代码检查 | 清理完成 |
| [ ] 移除非必须的 Deprecated 类型 | 代码检查 | 清理完成 |
| [ ] 移除非必须的 Deprecated 字段 | 代码检查 | 清理完成 |
| [ ] 编译通过 | `go build ./...` | 无错误 |
| [ ] 运行所有测试 | `go test ./...` | 全部通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 7.2 清理未使用的导入和代码

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 运行 goimports/gofmt | `goimports -w .` | 格式化完成 |
| [ ] 移除未使用的导入 | linter 检查 | 无警告 |
| [ ] 移除未使用的变量 | linter 检查 | 无警告 |
| [ ] 移除未使用的函数 | linter 检查 | 无警告 |
| [ ] 编译通过 | `go build ./...` | 无错误 |
| [ ] 运行 linter | `golangci-lint run` | 无新错误 |

#### 7.3 优化 RenderPlanes 性能

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] Benchmark BuildRenderPlanes | `go test -bench=.` | 有基准 |
| [ ] 分析性能热点 | `pprof` | 找到瓶颈 |
| [ ] 优化热路径 | 代码优化 | 提升 >20% |
| [ ] 优化后重新测试 | Benchmark 对比 | 性能提升 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 7.4 优化 HitMap 性能

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] Benchmark BuildHitMapFromFiber | `go test -bench=.` | 有基准 |
| [ ] 分析性能热点 | `pprof` | 找到瓶颈 |
| [ ] 优化热路径 | 代码优化 | 提升 >20% |
| [ ] 优化后重新测试 | Benchmark 对比 | 性能提升 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 7.5 添加更多单元测试

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] Fiber 结构测试覆盖率 >80% | `go test -cover` | 达标 |
| [ ] RenderPlanes 测试覆盖率 >80% | `go test -cover` | 达标 |
| [ ] HitMap 测试覆盖率 >80% | `go test -cover` | 达标 |
| [ ] Layout 测试覆盖率 >70% | `go test -cover` | 达标 |
| [ ] 运行所有测试 | `go test ./...` | 全部通过 |

#### 7.6 更新文档

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 更新 runtime/README.md | 检查文档 | 反映新架构 |
| [ ] 更新 docs/render/fiber/diff_layer.md | 检查文档 | 与实现一致 |
| [ ] 更新 internal/reconciler/doc.go | 检查文档 | 更新描述 |
| [ ] 更新 AGENTS.md | 检查文档 | 更新模式说明 |
| [ ] 添加迁移指南 | 文档存在 | 有清晰指南 |
| [ ] 代码注释完整 | 检查导出代码 | 都有注释 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

### 📊 Phase 7 性能检查

| 指标 | 测试方法 | 允许偏差 |
|------|---------|---------|
| [ ] 渲染帧率 > 60fps | 手动测试 | 流畅 |
| [ ] 常见操作延迟 < 100ms | 手动测试 | 响应快 |
| [ ] 内存占用稳定 | 长时间运行 | 无泄漏 |

### 🚨 Phase 7 阻塞风险

- [ ] ❌ 移除代码导致功能失效 → 回退修改
- [ ] ❌ 优化引入 bug → 回退优化
- [ ] ❌ 测试覆盖率不足 → 添加测试

---

## Phase 8: 综合测试

### ⏰ 预检查清单

**开始 Phase 8 前必须验证以下项：**

- [ ] ✅ Phase 7 已完成并通过验收
- [ ] ✅ 准备好完整测试环境
- [ ] ✅ 准备好测试数据集

### ✅ Phase 8 完成检查清单

#### 8.1 运行所有现有测试

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 所有单元测试通过 | `go test ./...` | 100% 通过 |
| [ ] 所有集成测试通过 | `go test -tags=integration ./...` | 100% 通过 |
| [ ] Race detector 测试通过 | `go test -race -short ./...` | 无竞态 |
| [ ] 基准测试无退化 | `go test -bench=. -benchmem` | 性能提升 |
| [ ] 记录测试报告 | 保存输出 | 有记录 |

#### 8.2 添加端到端测试

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] E2E 测试：完整应用生命周期 | 运行 E2E 测试 | 通过 |
| [ ] E2E 测试：复杂 UI 交互 | 运行 E2E 测试 | 通过 |
| [ ] E2E 测试：Modal/Overlay 交互 | 运行 E2E 测试 | 通过 |
| [ ] E2E 测试：事件冒泡 | 运行 E2E 测试 | 通过 |
| [ ] E2E 测试：键盘导航 | 运行 E2E 测试 | 通过 |
| [ ] E2E 测试：多层渲染 | 运行 E2E 测试 | 通过 |
| [ ] Code Review 通过 | PR Review | 至少 1 人通过 |

#### 8.3 用户场景测试

| 场景 | 验证方法 | 预期结果 |
|------|---------|---------|
| [ ] Counter 示例运行正常 | `cd examples/counter && go run .` | 正常 |
| [ ] Timer 示例运行正常 | `cd examples/timer && go run .` | 正常 |
| [ ] Input 示例运行正常 | `cd examples/input && go run .` | 正常 |
| [ ] Modal 示例运行正常 | `cd examples/modal && go run .` | 正常 |
| [ ] Overlay 示例运行正常 | `cd examples/overlay && go run .` | 正常 |
| [ ] 复杂示例运行正常 | `cd examples/complex && go run .` | 正常 |
| [ ] 所有示例无 panic 观察 | 手动测试 | 无崩溃 |

#### 8.4 性能基准测试

| 检查项 | 测试方法 | 基线要求 |
|--------|---------|----------|
| [ ] Layout 性能 | BenchmarkLayout | 不低于 Phase 6 |
| [ ] Render 性能 | BenchmarkRender | 不低于 Phase 6 |
| [ ] HitMap 构建 | BenchmarkBuildHitMap | <Phase 6 |
| [ ] HitTest | BenchmarkHitTest | <Phase 6 |
| [ ] RenderPlanes 构建 | BenchmarkBuildRenderPlanes | 记录基线 |
| [ ] 帧率测试 | 实际运行 | > 60fps |
| [ ] 内存占用 | pprof | 稳定 |
| [ ] 无内存泄漏 | `go test -race` + 长时间运行 | 无泄漏 |

#### 8.5 压力测试

| 测试项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 深层级组件树（100层） | 创建测试用例 | 不崩溃 |
| [ ] 宽节点数（1000个子节点） | 创建测试用例 | 不崩溃 |
| [ ] 快速连续渲染（1000次） | 循环测试 | 无泄漏 |
| [ ] 大量事件注入（10000次） | 注入测试 | 无泄漏 |
| [ ] 长时间运行（1小时） | 长期测试 | 稳定 |

#### 8.6 回归测试

| 测试项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] Fiber 节点匹配逻辑 | 运行 reconciler 测试 | 与之前一致 |
| [ ] Diff 结果一致性 | 对比旧版本 | 一致 |
| [ ] 组件状态保持 | 测试示例 | 状态正确 |
| [ ] 事件处理正确性 | 测试示例 | 正确处理 |
| [ ] 布局结果一致性 | 对比旧版本 | 一致 |
| [ ] 渲染结果一致性 | 视觉对比 | 无差异 |

#### 8.7 发布准备

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| [ ] 版本号更新 | 检查 go.mod | 新版本号 |
| [ ] CHANGELOG.md 更新 | 检查文件 | 有 CHANGELOG |
| [ ] Breaking Changes 文档 | 检查 docs/ | 有文档 |
| [ ] Migration Guide 完整 | 检查指南 | 完整清晰 |
| [ ] 示例代码更新 | 检查 examples/ | 已更新 |
| [ ] README 更新 | 检查 README.md | 反映新特性 |
| [ ] Release Notes 准备 | 检查 release notes | 完整 |
| [ ] Code Review 全部通过 | PR Review | 至少 2 人通过 |
| [ ] 主要维护者批准 | 获得批准 | 有人批准 |

### 📊 Phase 8 各类指标汇总

| 指标类别 | 子指标 | 目标值 | 实际值 | 状态 |
|---------|-------|--------|--------|------|
| **测试通过率** | 单元测试 | 100% | [待填] | ⬜ |
| | 集成测试 | 100% | [待填] | ⬜ |
| | E2E 测试 | 100% | [待填] | ⬜ |
| **代码覆盖率** | Fiber | >80% | [待填] | ⬜ |
| | RenderPlanes | >80% | [待填] | ⬜ |
| | HitMap | >80% | [待填] | ⬜ |
| | Layout | >70% | [待填] | ⬜ |
| **性能指标** | 帧率 | >60fps | [待填] | ⬜ |
| | Layout 延迟 | <10ms | [待填] | ⬜ |
| | Render 延迟 | <16ms | [待填] | ⬜ |
| | 内存占用 | 稳定 | [待填] | ⬜ |
| **稳定性** | 无 Panic | 0 | [待填] | ⬜ |
| | 无 Race | 0 | [待填] | ⬜ |
| | 无崩溃 | 0 | [待填] | ⬜ |

### 🚨 Phase 8 阻塞风险

- [ ] ❌ 测试失败 → 回滚到前一个阶段
- [ ] ❌ 性能退化严重 → 优化或回滚
- [ ] ❌ 出现严重 bug → 修复或回滚
- [ ] ❌ 不满足发布条件 → 延期发布

---

## 最终验收

### 🎯 核心验收标准

#### 1. 功能完整性

| 标准 | 验证方法 | 状态 |
|------|---------|------|
| ✅ 所有现有功能正常工作 | 运行所有示例 | ⬜ |
| ✅ Modal/Overlay 正常显示和交互 | 测试 Modal/Overlay | ⬜ |
| ✅ 多层渲染正确（Base→Overlay→Modal） | 视觉测试 | ⬜ |
| ✅ HitTest 正确（考虑 Z-order） | 点击测试 | ⬜ |
| ✅ 事件冒泡正确 | 事件测试 | ⬜ |
| ✅ 组件状态保持 | 状态测试 | ⬜ |
| ✅ Fiber 节点匹配正确 | Diff 测试 | ⬜ |

#### 2. 架构正确性

| 标准 | 验证方法 | 状态 |
|------|---------|------|
| ✅ Fiber 是唯一运行时结构 | 代码审查 | ⬜ |
| ✅ StripLayers 完全移除 | 代码审查 | ⬜ |
| ✅ RenderPlanes 使用投影而非克隆 | 代码审查 | ⬜ |
| ✅ Layer 是渲染维度而非结构维度 | 代码审查 | ⬜ |
| ✅ NodeID 用于运行时标识 | 代码审查 | ⬜ |
| ✅ DiffKey 用于 Diff 匹配 | 代码审查 | ⬜ |

#### 3. 质量标准

| 标准 | 目标值 | 实际值 | 状态 |
|------|--------|--------|------|
| ✅ 测试覆盖率 | >75% | [待填] | ⬜ |
| ✅ 所有测试通过 | 100% | [待填] | ⬜ |
| ✅ 无编译警告 | 0 | [待填] | ⬜ |
| ✅ 无竞态条件 | 0 | [待填] | ⬜ |
| ✅ 性能无显著退化 | <5% | [待填] | ⬜ |
| ✅ 无内存泄漏 | 0 | [待填] | ⬜ |

#### 4. 文档完整性

| 标准 | 验证方法 | 状态 |
|------|---------|------|
| ✅ 代码注释完整 | linter | ⬜ |
| ✅ API 文档更新 | godoc | ⬜ |
| ✅ 架构文档更新 | 检查 docs/ | ⬜ |
| ✅ Migration Guide 可用 | 检查指南 | ⬜ |
| ✅ README 更新 | 检查 README | ⬜ |
| ✅ CHANGELOG 更新 | 检查 CHANGELOG | ⬜ |

### 📋 最终验收检查清单

**所有以下项目必须为 ✅ 才能发布：**

#### Phase 完成检查

- [ ] ✅ Phase 1 验收通过
- [ ] ✅ Phase 2 验收通过
- [ ] ✅ Phase 3 验收通过
- [ ] ✅ Phase 4 验收通过
- [ ] ✅ Phase 5 验收通过
- [ ] ✅ Phase 6 验收通过
- [ ] ✅ Phase 7 验收通过
- [ ] ✅ Phase 8 验收通过

#### 代码质量检查

- [ ] ✅ 所有编译警告已修复
- [ ] ✅ 所有测试通过
- [ ] ✅ 代码格式正确（gofmt）
- [ ] ✅ linter 无新错误
- [ ] ✅ 竞态检测通过
- [ ] ✅ Code Review 通过

#### 功能检查

- [ ] ✅ 所有示例正常运行
- [ ] ✅ Modal/Overlay 正常工作
- [ ] ✅ 多层渲染正确
- [ ] ✅ HitTest 正确
- [ ] ✅ 事件系统正常
- [ ] ✅ 组件状态保持

#### 性能检查

- [ ] ✅ 渲染帧率 > 60fps
- [ ] ✅ 常见操作延迟 < 100ms
- [ ] ✅ 无显著性能退化
- [ ] ✅ 内存占用稳定
- [ ] ✅ 无内存泄漏

#### 文档检查

- [ ] ✅ 代码注释完整
- [ ] ✅ API 文档更新
- [ ] ✅ 架构文档更新
- [ ] ✅ Migration Guide 可用
- [ ] ✅ README 更新
- [ ] ✅ CHANGELOG 更新

#### 发布准备

- [ ] ✅版本号更新
- [ ] ✅ Release Notes 完成
- [ ] ✅ Breaking Changes 文档
- [ ] ✅ 主要维护者批准
- [ ] ✅ 可以合并到主分支

---

## 回滚决策

### 🚨 何时触发回滚

如果以下任何条件为 ❌，**必须立即停止**并评估是否需要回滚：

#### 阻塞性问题（必须回滚）

| 情况 | 描述 | 决策 |
|------|------|------|
| ❌ **核心功能失效** | Modal 无法显示、事件不工作、布局错误等 | **立即回滚** |
| ❌ **严重性能退化** | 帧率 < 30fps、操作延迟 > 500ms | **评估回滚** |
| ❌ **内存泄漏** | 长时间运行内存持续增长 | **必须回滚** |
| ❌ **崩溃频繁** | 运行示例频繁 panic | **必须回滚** |
| ❌ **数据破坏** | 组件状态丢失、数据损坏 | **必须回滚** |
| ❌ **无法修复** | 多次尝试仍无法解决 | **回滚** |

#### 警告性问题（评估回滚）

| 情况 | 描述 | 决策 |
|------|------|------|
| ⚠️ **轻微性能退化** | 帧率 30-60fps、延迟 100-500ms | 评估优化 |
| ⚠️ **少数测试失败** | 非核心测试失败 | 评估修复 |
| ⚠️ **UI 细节问题** | 样式、间距不一致 | 评估修复 |
| ⚠️ **兼容性问题** | 部分第三方组件不兼容 | 评估适配 |
| ⚠️ **时间压力** | 预定时间无法完成 | 评估延期或回滚 |

### 📋 回滚检查清单

如果需要回滚，按以下流程操作：

#### 发现问题（0-2小时）

- [ ] 记录问题描述
- [ ] 记录复现步骤
- [ ] 记录影响范围
- [ ] 记录严重程度（阻塞/警告）
- [ ] 评估是否可以快速修复

#### 评估回滚（2-4小时）

- [ ] 确定要回滚到的阶段
- [ ] 评估回滚影响
- [ ] 评估回滚工作量
- [ ] 评估修复 vs 回滚
- [ ] 做出回滚决策

#### 执行回滚（4-8小时）

- [ ] 创建回滚分支
- [ ] 回退代码到上一个稳定状态
- [ ] 验证回滚后的功能
- [ ] 运行所有测试
- [ ] 更新文档

#### 回滚后验证（8-12小时）

- [ ] 运行所有示例
- [ ] 运行性能测试
- [ ] 确认问题解决
- [ ] 记录回滚原因
- [ ] 更新计划

### 🔄 回滚策略

#### 阶段回滚

| 当前阶段 | 回滚到 | 难度 | 数据丢失 |
|---------|--------|------|----------|
| Phase 1 | 无（第一阶段） | - | - |
| Phase 2 | Phase 1 | 中 | 否 |
| Phase 3 | Phase 2 | 低 | 否 |
| Phase 4 | Phase 3 | 低 | 否 |
| Phase 5 | Phase 4 | 中 | 否 |
| Phase 6 | Phase 5 | 中 | 否 |
| Phase 7 | Phase 6 | 高（代码已清理） | 否 |
| Phase 8 | Phase 7 | 高 | 否 |

#### 完全回滚

如果整个重构失败，需要完全回滚：

- [ ] 删除所有新分支
- [ ] 恢复到重构前的主分支
- [ ] 记录失败原因
- [ ] 重新设计方案
- [ ] 更新文档

#### 部分回滚

如果只是某个功能需要回滚：

- [ ] 确定需要回滚的功能
- [ ] 回退相关代码
- [ ] 保留已完成的独立部分
- [ ] 评估影响范围
- [ ] 更新文档

---

## 签收模板

### Phase 验收签收

**Phase XX 验收签收**

| 项目 | 内容 |
|------|------|
| **Phase 编号** | X |
| **Phase 名称** | [名称] |
| **开始日期** | [日期] |
| **完成日期** | [日期] |
| **开发者** | [姓名] |
| **Reviewer** | [姓名] |
| **代码审查链接** | [PR 链接] |

#### 验收结果

- [ ] ✅ 所有预检查项通过
- [ ] ✅ 所有完成检查项通过
- [ ] ✅ 所有代码审查通过
- [ ] ✅ 所有测试通过
- [ ] ✅ 性能检查通过
- [ ] ✅ 集成测试通过
- [ ] ✅ 文档已更新

#### 验收结论

**[ ] 通过**  **[ ] 不通过**

**Reviewer 签名**: _________________ **日期**: __________________

**开发者签名**: _________________ **日期**: __________________

---

### 最终验收签收

**Fiber 统一架构重构 - 最终验收签收**

| 项目 | 内容 |
|------|------|
| **项目名称** | Fiber 统一架构重构 |
| **开始日期** | [日期] |
| **完成日期** | [日期] |
| **总工期** | [天数] |
| **项目负责人** | [姓名] |

#### Phase 完成情况

| Phase | 状态 | 签收人 | 日期 |
|-------|------|--------|------|
| Phase 1: 基础设施 | [待填] | [待填] | [待填] |
| Phase 2: Layout 重构 | [待填] | [待填] | [待填] |
| Phase 3: RenderPlane 引入 | [待填] | [待填] | [待填] |
| Phase 4: 废弃 StripLayers | [待填] | [待填] | [待填] |
| Phase 5: Render 更新 | [待填] | [待填] | [待填] |
| Phase 6: HitMap 更新 | [待填] | [待填] | [待填] |
| Phase 7: 清理和优化 | [待填] | [待填] | [待填] |
| Phase 8: 综合测试 | [待填] | [待填] | [待填] |

#### 核心指标

| 指标 | 目标值 | 实际值 | 状态 |
|------|--------|--------|------|
| 测试覆盖率 | >75% | [待填] | [待填] |
| 平均帧率 | >60fps | [待填] | [待填] |
| 内存泄漏 | 0 | [待填] | [待填] |
| 测试通过率 | 100% | [待填] | [待填] |

#### 验收检查项

**功能完整性**
- [ ] ✅ 所有现有功能正常工作
- [ ] ✅ Modal/Overlay 正常工作
- [ ] ✅ 多层渲染正确
- [ ] ✅ HitTest 正确
- [ ] ✅ 事件系统正常

**架构正确性**
- [ ] ✅ Fiber 是唯一运行时结构
- [ ] ✅ StripLayers 完全移除
- [ ] ✅ RenderPlanes 使用投影
- [ ] ✅ NodeID/DiffKey 职责清晰

**质量标准**
- [ ] ✅ 测试覆盖率达标
- [ ] ✅ 所有测试通过
- [ ] ✅ 代码符合规范
- [ ] ✅ 性能无显著退化

**文档完整性**
- [ ] ✅ 代码注释完整
- [ ] ✅ API 文档更新
- [ ] ✅ Migration Guide 可用
- [ ] ✅ README 更新

#### 风险评估

| 风险类型 | 风险等级 | 缓解措施 |
|---------|---------|---------|
| [待填] | [待填] | [待填] |

#### 验收结论

**[ ] 通过 - 可以合并到主分支并发布**

**[ ] 不通过 - 需要 [修复/回滚/延期]**

**详细说明**:
[填写详细说明...]

---

**技术负责人签名**: _________________ **日期**: __________________

**项目经理签名**: _________________ **日期**: __________________

**主要维护者批准**: _________________ **日期**: __________________

---

## 附录

### A. 检查清单使用说明

1. **每个 Phase 开始前**，执行该 Phase 的预检查清单
2. **每个 Phase 完成后**，执行该 Phase 的完成检查清单
3. **所有项目必须为 ✅** 才能进入下一个 Phase
4. **出现 ⚠️ 或 ❌ 项目**，需要先解决才能继续
5. **定期更新进度**，填写实际值到表格中

### B. 签收流程

1. **Phase 签收**: 每个 Phase 完成后，填写 Phase 验收签收模板
2. **Code Review**: 所有代码更改必须经过 Code Review
3. **测试验证**: 所有测试通过才能签收
4. **文档检查**: 文档更新完成才能签收
5. **最终验收**: 所有 Phase 完成后，填写最终验收签收模板

### C. 故障处理流程

1. **发现问题**: 记录问题描述、复现步骤
2. **评估影响**: 确定影响范围和严重程度
3. **决策**: 确定是修复还是回滚
4. **执行**: 按决策执行修复或回滚
5. **验证**: 测试验证问题已解决
6. **记录**: 更新文档和计划

### D. 联系方式

| 角色 | 姓名 | 邮箱 | GitHub |
|------|------|------|--------|
| 项目负责人 | [待填] | [待填] | [待填] |
| 技术负责人 | [待填] | [待填] | [待填] |
| Reviewer | [待填] | [待填] | [待填] |
| 测试负责人 | [待填] | [待填] | [待填] |

---

**文档版本**: 1.0
**最后更新**: 2026-02-14
**维护者**: [待分配]
