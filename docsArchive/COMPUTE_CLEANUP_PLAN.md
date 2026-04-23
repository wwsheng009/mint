# Compute 包清理计划

## 一、现状分析

### 1.1 重复的布局引擎

| 引擎 | 位置 | 算法实现 | 状态 |
|------|------|----------|------|
| `compute.Engine` | `runtime/compute/engine.go` | buildComputedBox, calculatePositions, measureVNode | ⚠️ 重复 |
| `compute.Engine` | `runtime/compute/fiber_only_layout.go` | buildFiberOnlyBox, calculateFiberOnlyPositions | ⚠️ 重复 |
| `layout.Engine` | `runtime/layout/types.go` | layoutNode, layoutNodeWithDepth | ✅ 保留 |

### 1.2 重复的 Box 类型

| Box 类型 | 位置 | 特点 | 状态 |
|----------|------|------|------|
| `compute.ComputedBox` | `runtime/compute/types.go` | 包含 VNode 引用 + NodeID + 布局结果 | ⚠️ 混合职责 |
| `layout.LayoutBox` | `runtime/layout/types.go` | 纯布局数据，无框架依赖 | ✅ 保留 |
| `paint.PaintableBox` | `runtime/paint/paintable_box.go` | 纯绘制数据 | ✅ 保留 |

### 1.3 现有适配器状态

| 适配器 | 位置 | 功能 | 状态 |
|--------|------|------|------|
| `FiberToNodeAdapterPure` | `internal/render/fiber_adapter.go` | Fiber → layout.Node | ✅ 完整 |
| `VNodeToNodeAdapter` | `internal/render/fiber_adapter.go` | VNode → layout.Node | ⚠️ Deprecated |
| `FiberPaintableNode` | `internal/render/converter.go` | Fiber → paint.PaintableNode | ✅ 完整 |
| `FiberToPaintableConverter` | `internal/render/converter.go` | LayoutBox + Fiber → PaintableBox | ✅ 完整 |

## 二、目标架构

### 2.1 统一渲染流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                    目标: 单一布局引擎架构                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Phase 1: Reconcile                                                │
│   ─────────────────────                                             │
│   VNode → Reconciler → Fiber (VNode 丢弃)                           │
│                                                                     │
│   Phase 2: Layout (单一引擎)                                         │
│   ─────────────────────                                             │
│   Fiber → FiberToNodeAdapterPure → layout.Node → layout.Engine     │
│                                         ↓                           │
│                                   layout.LayoutBox                  │
│                                                                     │
│   Phase 3: Paint                                                    │
│   ─────────────────────                                             │
│   LayoutBox + Fiber → FiberToPaintableConverter → PaintableBox     │
│                                                    ↓                │
│                                            PaintEngine → Buffer     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 删除的代码

| 文件 | 代码 | 原因 |
|------|------|------|
| `compute/engine.go` | `buildComputedBox*` | 被 `layout.Engine` 替代 |
| `compute/engine.go` | `calculatePositions*` | 被 `layout.Engine` 替代 |
| `compute/engine.go` | `measureVNode*` | 被 `layout.Node.Measure()` 替代 |
| `compute/engine.go` | `measureLayoutChildren*` | 被 `layout.FlexLayout` 替代 |
| `compute/engine.go` | `measureBordered*` | 被 `layout.Border` 处理 |
| `compute/engine.go` | `measureTable*` | 被 `layout.GridLayout` 替代 |
| `compute/fiber_only_layout.go` | 整个文件 | 被 `layout.Engine` 替代 |
| `compute/adapter_vnode.go` | 整个文件 | 不再需要 VNode 适配 |
| `compute/adapter_fiber.go` | 整个文件 | 被适配器替代 |
| `internal/render/fiber_adapter.go` | `VNodeToNodeAdapter` | 不再使用 VNode |

### 2.3 保留的代码

| 文件 | 代码 | 原因 |
|------|------|------|
| `compute/types.go` | `ComputedBox` 结构体 | PaintEngine 仍使用 (过渡期) |
| `compute/cache.go` | `LayoutCache` | 可复用或移到 layout |
| `compute/dirty_tracker.go` | `DirtyTracker` | 可复用或移到 layout |
| `compute/bounds_validator.go` | `BoundsValidator` | 可复用或移到 layout |
| `compute/layout_validator.go` | `LayoutValidator` | 可复用或移到 layout |

## 三、迁移步骤

### 阶段 1: 验证适配器完整性 (1-2天)

- [ ] 确认 `FiberToNodeAdapterPure` 实现所有必要接口
  - [ ] `layout.Node` - 基础接口
  - [ ] `layout.Measurable` - 测量接口
  - [ ] `layout.Marginal` - 边距接口
  - [ ] `layout.Positionable` - 定位接口
  - [ ] `layout.FlexStyleProvider` - Flex 布局
  - [ ] `layout.GridStyleProvider` - Grid 布局
  - [ ] `layout.WrapStyleProvider` - Wrap 布局
  - [ ] `layout.AbsoluteStyleProvider` - 绝对定位

- [ ] 确认 `FiberToPaintableConverter` 正确工作
  - [ ] LayoutBox → PaintableBox 转换
  - [ ] Fiber 数据填充
  - [ ] 子节点递归

### 阶段 2: 迁移 PaintEngine (2-3天)

- [ ] 修改 `PaintEngine` 接受 `PaintableBox` 而非 `ComputedBox`
- [ ] 删除 `PaintEngine` 中对 `ComputedBox.VNode` 的直接访问
- [ ] 使用 `PaintableBox.Node` (PaintableNode 接口) 替代

### 阶段 3: 迁移调用点 (1-2天)

- [ ] `rendering_pipeline.go` - 使用 `layout.Engine`
- [ ] `layer/manager.go` - 使用 `layout.Engine`
- [ ] `declarative_node.go` - 使用 Fiber-first 路径

### 阶段 4: 删除重复代码 (1天)

- [ ] 删除 `compute/engine.go` 中的布局算法
- [ ] 删除 `compute/fiber_only_layout.go`
- [ ] 删除 `compute/adapter_*.go`
- [ ] 删除 `VNodeToNodeAdapter`

### 阶段 5: 清理与测试 (1-2天)

- [ ] 运行完整测试套件
- [ ] 确认 Fiber-first 模式正常工作
- [ ] 确认 Legacy 模式回退正常

## 四、ComputedBox 处理策略

### 选项 A: 完全删除 ComputedBox

**优点:**
- 最干净的架构
- 完全分离布局和绘制

**缺点:**
- 需要修改所有使用 `ComputedBox` 的代码
- 工作量较大

### 选项 B: 简化 ComputedBox 为 PaintableBox 别名 (推荐)

**优点:**
- 兼容性好
- 渐进式迁移

**实现:**
```go
// compute/types.go
type ComputedBox = paint.PaintableBox  // 类型别名
```

### 选项 C: 保留 ComputedBox 作为中间类型

**优点:**
- 最小改动
- 向后兼容

**缺点:**
- 保留冗余

## 五、当前行动项

### 立即行动

1. **验证 `FiberToNodeAdapterPure` 完整性**
   - 检查是否实现所有 `layout.*` 接口
   - 补充缺失的接口实现

2. **优化 `FiberToPaintableConverter`**
   - 简化 Fiber 查找逻辑
   - 确保 NodeID 正确传播

3. **创建 `VNodeToNodeAdapter` 优化版** (如需要)
   - 仅用于非 Fiber 模式回退
   - 复用 `FiberToNodeAdapterPure` 的逻辑

## 六、风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| PaintEngine 依赖 ComputedBox | 高 | 先迁移 PaintEngine 到 PaintableBox |
| 测试覆盖率不足 | 中 | 增加集成测试 |
| Legacy 模式破坏 | 中 | 保留回退路径 |

## 七、验收标准

- [ ] Fiber-first 模式 (`MINT_FIBER_FIRST=true`) 使用 `layout.Engine`
- [ ] Legacy 模式正常工作
- [ ] 所有测试通过
- [ ] `compute.Engine` 不再包含布局算法
- [ ] 代码行数减少 30%+
