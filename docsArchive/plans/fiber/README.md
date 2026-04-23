# Fiber 统一架构重构方案 - 概览

**版本**: 1.0
**日期**: 2026-02-14
**主文档**: [UNIFIED_FIBER_ARCHITECTURE_REFACTOR.md](./UNIFIED_FIBER_ARCHITECTURE_REFACTOR.md)

---

## 📋 重构目标

本次重构根据设计文档 `diff_key.md` 和 `diff_layer.md`，对 Mint TUI Framework 进行重大架构重组：

### 核心理念

1. **Fiber 作为唯一运行时结构** - 所有运行时状态、事件处理、布局计算都基于 Fiber
2. **Layer 作为渲染排序维度** - Layer 不参与结构变换，只控制 Z 轴渲染顺序
3. **废弃 StripLayers** - 不再使用 VNode 克隆和树结构分离的方式实现多层渲染
4. **RenderPlane 投影机制** - 使用渲染分桶而非树结构分离实现多层渲染

---

## 🎯 当前架构问题

### 问题 1：StripLayers 违背单一数据源原则

- StripLayers 创建 VNode 的克隆副本
- 导致 VNode 树 ≠ Fiber 树
- identity 来源分裂（VNode 和 Fiber 各有一套）
- Key 同步困难
- 增加内存开销

### 问题 2：Layer 作为结构维度参与 Layout

- Layer 参与了结构变换（stripping）
- 每个 Layer 有独立的 Layout 树
- 布局上下文不一致（base 和 modal 的父子关系丢失）

### 问题 3：HitMap 基于多源数据构建

- 需要从多个 Layout 构建多个 HitMap
- 需要合并多个 HitMap
- 需要手动管理 Z-order
- 容易出现同步问题

### 问题 4：Fiber 树与渲染树不一致

- Fiber 树：完整的 VNode 树结构
- 渲染树：被 StripLayers 后的多个独立树
- Fiber 树的 NodeID 无法直接映射到渲染树
- Event Dispatch 需要在不同的树之间查找

### 问题 5：VNode 承担了过多职责

- VNode 参与了运行时逻辑（Layer 管理）
- VNode 不应该是纯声明，它影响了运行时行为
- 违反关注点分离原则

---

## 🔧 新架构设计

### 核心原则

#### 原则 1：Fiber 是唯一运行时结构

> "Everything happens in Fiber. VNode is just declarative input."

- 所有运行时状态都保存在 Fiber
- 所有事件处理基于 Fiber.NodeID
- 所有布局计算基于 Fiber 树
- 所有渲染基于 Fiber 树

#### 原则 2：Layer 是渲染排序维度

> "Layer controls drawing order, not tree structure."

- Layer 只是 Fiber 的一个属性
- Layer 不参与 diff 规则
- Layer 不影响 Fiber 树结构
- Layer 只在渲染阶段用于分桶

#### 原则 3：单一 Fiber 树

> "There is only one tree - the Fiber tree."

- 系统中只有一棵完整的 Fiber 树
- 所有节点（包括 Modal、Overlay）都在这棵树中
- Layer 只是标记，不是独立子树

#### 原则 4：VNode 纯声明

> "VNode describes what to render, not how to render."

- VNode 只描述 UI 结构
- VNode 不包含运行时信息
- VNode 每次 render 都是新的快照

#### 原则 5：RenderPlane 投影而非 Strip

> "Project, don't strip."

- 使用 RenderPlane 进行渲染分桶
- 不修改树结构，只是创建投影视图
- HitMap 基于单一 Fiber 树构建

### 架构层次

```
用户代码 (User Code)
    ↓ ↓
声明层 (VNode Tree) - Type, Key, Props, Children
    ↓ ↓
协调层 (Fiber Tree) - NodeID, DiffKey, Type, Layer, Props
    ↓ ↓
布局层 (ComputedBox Tree) - Box, NodeID, Layer
    ↓ ↓
渲染投影层 (RenderPlanes) - 按 Layer 分桶
    ↓ ↓
渲染层 (Paint Buffer) - 按 Layer 顺序绘制
    ↓ ↓
事件层 (HitMap) - 基于 ComputedBox 构建
```

---

## 📦 数据结构更新

### Fiber

```go
type Fiber struct {
    // ✨ 新增：Layer 声明值（从 VNode 拷贝）
    Layer rtui.Layer

    // ✨ 新增：ComputedBox 保存 Layout 结果
    ComputedBox *compute.ComputedBox

    // 现有字段保持不变
    NodeID   uint64
    DiffKey  string
    Type     VNodeType
    Props    Props
    // ...
}
```

### ComputedBox

```go
type ComputedBox struct {
    // ✨ 新增：NodeID 从 Fiber 拷贝
    NodeID uint64

    // ✨ 新增：Layer 从 Fiber 拷贝
    Layer rtui.Layer

    // ✨ 新增：子节点的 ComputedBox
    Children []*ComputedBox

    Box Box
    VNode rtui.VNode
}
```

### RenderPlanes（✨ 新增）

```go
type RenderPlanes struct {
    // 按 Layer 分桶的 ComputedBox
    Planes map[rtui.Layer][]*compute.ComputedBox

    // 渲染顺序（从低到高）
    RenderOrder []rtui.Layer
}

// 从 Fiber 树构建 RenderPlanes
func (rp *RenderPlanes) BuildFromFiber(root *Fiber)

// 按渲染顺序遍历
func (rp *RenderPlanes) Iterate(fn func(layer rtui.Layer, box *compute.ComputedBox) bool)
```

---

## 🔄 数据流

```
User Code
   ↓ ↓
Render() → new VNode tree
   ↓ ↓
Reconcile(oldFiberTree, newVNodeTree)
   └─ 基于 DiffKey 匹配
   └─ 复用或创建 Fiber
   └─ 保留 NodeID（复用时）或分配新 NodeID（新建时）
   ↓ ↓
newFiberTree（单棵完整树）
   ├─ Normal nodes (LayerBase)
   ├─ Modal nodes (LayerModal)
   ├─ Overlay nodes (LayerOverlay)
   └─ ...
   ↓ ↓
Layout(newFiberTree)
   └─ 遍历整棵 Fiber 树
   └─ 为每个 Fiber 创建/更新 ComputedBox
   └─ 保存 Layout 结果到 Fiber.ComputedBox
   ↓ ↓
RenderPlanes（按 Layer 分桶）
   └─ 遍历所有 ComputedBox
   └─ 按 Layer 分桶
   ↓ ↓
Renderer
   └─ 按 Layer 顺序绘制
   ↓ ↓
HitMap（基于所有 ComputedBox）
   └─ 按 Layer 和 Z-order 排序
   ↓ ↓
Event Dispatch
   └─ 基于 NodeID 找到 Fiber
   └─ 基于 Fiber 找到 Instance
```

---

## 🚀 实施计划

### 8 个阶段，预计 8 周

#### Phase 1: 基础设施（1-2周）
- Fiber 新增 `Layer` 和 `ComputedBox` 字段
- Reconciler 从 VNode 拷贝 Layer 到 Fiber
- 添加 BuildHitMapFromFiber() API

#### Phase 2: Layout 重构（1-2周）
- Engine.Layout() 修改为基于 Fiber
- Layout 结果附加到 Fiber.ComputedBox
- 废弃独立的 Layer Layout 路径

#### Phase 3: RenderPlane 引入（1周）
- 添加 RenderPlanes 类型
- 实现 BuildFromFiber() 方法
- 与现有 CollectAndLayout() 并存

#### Phase 4: 废弃 StripLayers（1周）
- 标记 StripLayers 为 Deprecated
- 移除所有 StripLayers 调用点
- 移除克隆 VNode 逻辑

#### Phase 5: Render 更新（1周）
- 更新 FiberRenderer.Render()
- 按 Layer 顺序渲染 RenderPlanes

#### Phase 6: HitMap 更新（1周）
- 移除 GetMergedHitMap() 方法
- 所有 HitTest 使用 BuildHitMapFromFiber()
- 验证 HitTest 正确性

#### Phase 7: 清理和优化（1周）
- 移除所有 Deprecated 代码
- 优化 RenderPlanes 性能
- 添加更多单元测试

#### Phase 8: 综合测试（1周）
- 运行所有现有测试
- 添加集成测试
- 性能测试
- 用户场景测试

---

## ⚠️ 破坏性变更

### 废弃的 API

```go
// ❌ runtime/layer/manager.go
func (m *Manager) CollectAndLayout(...) error

// ❌ runtime/layer/collector.go
func (c *Collector) StripLayers(vnode rtui.VNode) rtui.VNode
func (c *Collector) cloneWithoutLayers(vnode rtui.VNode) rtui.VNode

// ❌ runtime/layer/manager.go
func (m *Manager) GetMergedHitMap() *event.HitMap
func (m *Manager) GetLayouts() LayerLayouts

// ❌ runtime/ui/vnode.go（如果有）
func (v VNode) SetBounds(x, y, width, height int)
func (v VNode) GetBounds() [4]int
```

### 新增的 API

```go
// ✨ runtime/layer/manager.go
func (m *Manager) BuildRenderPlanes(root *Fiber) *RenderPlanes
func (m *Manager) GetRenderPlanes() *RenderPlanes

// ✨ runtime/layer/manager.go
type RenderPlanes struct {
    Planes      map[rtui.Layer][]*compute.ComputedBox
    RenderOrder []rtui.Layer
}
func (rp *RenderPlanes) BuildFromFiber(root *Fiber)
func (rp *RenderPlanes) GetPlane(layer rtui.Layer) []*compute.ComputedBox
func (rp *RenderPlanes) Iterate(fn func(layer rtui.Layer, box *compute.ComputedBox) bool)

// ✨ runtime/compute/engine.go
func (e *Engine) Layout(vnode rtui.VNode, fiber *Fiber, constraints BoxConstraints) (*ComputedLayout, error)

// ✨ runtime/event/hitmap.go
func BuildHitMapFromFiber(root *Fiber) *HitMap
```

### 行为变更

- **Modal 布局不再独立** - Modal 在 Fiber 树中，与父组件在同一棵树
- **HitMap 基于 Fiber 树** - 直接从单一 Fiber 树构建，不需要合并
- **Debug 输出不同** - 所有节点都有 NodeID 标识

---

## 📖 迁移指南

### 组件开发者

```go
// ❌ 旧代码
baseTree := layerManager.CollectAndLayout(vnode, fiber, constraints)
modalNodes := layerManager.GetModalNodes()

// ✨ 新代码
fiberRoot := a.reconciler.GetFiberRoot()
renderPlanes := layerManager.BuildRenderPlanes(fiberRoot)
modalBoxes := renderPlanes.GetPlane(rtui.LayerModal)
```

### 第三方库

```go
// ❌ 旧代码
func MyLayoutEngine(vnode rtui.VNode) rtui.VNode {
    collector := layer.NewCollector()
    collector.Collect(vnode)
    return collector.StripLayers(vnode)
}

// ✨ 新代码
func MyLayoutEngine(fiberRoot *Fiber) *RenderPlanes {
    renderPlanes := NewRenderPlanes()
    renderPlanes.BuildFromFiber(fiberRoot)
    return renderPlanes
}
```

---

## 🧪 测试策略

- **单元测试**：验证 Fiber、RenderPlanes、HitMap 的正确性
- **集成测试**：验证 Reconcile、Layout、Render、Event 的集成
- **性能测试**：确保无性能退化（O(n) 时间复杂度）
- **用户场景测试**：测试 Modal、Overlay、Tooltip 的真实使用场景

---

## ⚠️ 风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| Layout 行为改变 | 中等 | 详细对比测试，保持 Modal 居中逻辑 |
| 性能退化 | 低 | 性能基准测试，优化 RenderPlanes |
| 事件处理错误 | 高 | 详细事件测试，验证 Z-order 排序 |
| 破坏性变更太多 | 中等 | 详细迁移指南，逐步废弃 |
| 测试覆盖不足 | 中等 | 大量单元测试和集成测试 |

---

## 📚 相关文档

- **主文档**: [UNIFIED_FIBER_ARCHITECTURE_REFACTOR.md](./UNIFIED_FIBER_ARCHITECTURE_REFACTOR.md)
- **设计文档 1**: [../render/fiber/diff_key.md](../../render/fiber/diff_key.md)
- **设计文档 2**: [../render/fiber/diff_layer.md](../../render/fiber/diff_layer.md)
- **架构分析**: [../layer_system_analysis.md](../layer_system_analysis.md)

---

## 📞 联系方式

如有问题或建议，请：

1. 查阅主文档中的详细说明
2. 查看 GitHub Issues
3. 提交 Pull Request

---

**文档版本**: 1.0
**最后更新**: 2026-02-14
