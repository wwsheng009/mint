# Portal 架构核心总结

> **当前实现状态**: 共享 Fiber 实现（Pragmatic 方式）
> **设计文档建议**: 独立 Fiber 子树（Ideal 方式）
> **功能等价性**: 两种实现方式在功能上完全等价，差异仅在于 Fiber 层的实现细节

---

## 📌 实现方式对比表

| 维度 | 共享 Fiber（当前实现） | 独立 Fiber 子树（设计文档） | 影响 |
|------|---------------------|--------------------------|------|
| **Fiber 结构** | Portal 节点保留在主树中 | Portal 子树与主树完全分离 | Fiber 层结构差异 |
| **Diff** | Portal 作为普通节点参与 diff | Portal 子树独立 diff | Diff context 不同 |
| **Layout** | 主树跳过 Portal 子节点 | 主树不包含 Portal | Layout 适配层简化 |
| **LayoutBox** | Overlay 阶段重新生成 | Overlay 阶段直接使用 | LayoutBox 生成路径不同 |
| **Render** | 完全相同（基于 LayoutBox） | 完全相同（基于 LayoutBox） | 无影响 |
| **事件** | 完全相同（基于 Z-order） | 完全相同（基于 Z-order） | 无影响 |
| **Focus** | 基于单树遍历 | 基于多树遍历 | Focus 遍历逻辑不同 |
| **性能** | 单树 reconcile，开销低 | 双树 reconcile，开销中高 | 性能略有差异 |
| **复杂度** | 🟢 低（代码简洁） | 🔴 高（需管理多树） | 维护成本差异 |
| **功能等价性** | ✅ 完全满足 | ✅ 完全满足 | 功能无差异 |
| **改造工作量** | - | 🔴 10-15 天 | 迁移成本高 |

---

## 二、Fiber 层实现差异

### 🟢 共享 Fiber 实现（当前）

**架构图**:
```
主 Fiber Tree:
  Container
    ├── Button
    ├── PortalRoot (portalRootId="modal-root")
    └── Portal (portalRoot="modal-root")  ← 保留在主树中
          └── Modal  ← 共享 Fiber 节点，逻辑子节点
```

**关键特点**:
- ✅ Portal 节点作为普通 Fiber 节点存在于主树
- ✅ Portal.Parent 指向主树父节点（逻辑关系不变）
- ✅ Portal 子节点仍通过 `Portal.Child` 访问
- ✅ 在 Layout 阶段，通过 `PortalAwareFiberToNodeAdapter` 跳过 Portal 子节点
- ✅ 在 Overlay 阶段，通过 `NewFiberToNodeAdapterPure` 转换 Portal 子节点

**代码示例**:
```go
// internal/render/portal_layout_adapter.go

// 主树 Layout: Portal 子节点为空
func (a *PortalAwareFiberToNodeAdapter) initChildren() {
    if a.isPortal {
        a.children = make([]layout.Node, 0)  // 跳过
        return
    }
    // 正常构建子节点
}

// Overlay Layout: 使用共享 Fiber 子节点
func (pc *PortalCollector) AddPortal(portalFiber *reconciler.Fiber) error {
    // 使用 NewFiberToNodeAdapterPure 转换 Portal 子节点
    children := make([]layout.Node, 0)
    childFiber := portalFiber.Child
    for childFiber != nil {
        children = append(children, NewFiberToNodeAdapterPure(childFiber))
        childFiber = childFiber.Sibling
    }
    // ...
}
```

### 🔴 独立 Fiber 子树（设计文档建议）

**架构图**:
```
主 Fiber Tree:
  Container
    ├── Button
    └── PortalRoot (portalRootId="modal-root")

独立 Portal Fiber Tree #1:
  Modal  ← 完全独立的 Fiber 子树，指向原 Instance
```

**关键特点**:
- ✅ Portal 节点不在主树中（或仅保留 shell）
- ✅ Portal 子树完全独立，有自己的 Fiber 根节点
- ✅ 主树和 Portal 树分别 reconcile、diff、layout
- ✅ 在 Reconciler 层收集 Portal 子树并独立处理
- ✅ Layout 层直接使用独立 Fiber 子树，不需要适配器

**伪代码示例**:
```go
// internal/reconciler/portal_subtree.go (未实现)

type PortalSubtree struct {
    portalRootID  string
    portalFiber   *Fiber  // 原 Portal 节点（可能保留 shell）
    treeRoot      *Fiber  // 独立的 Fiber 根节点
    workInProgress *Fiber
}

// Reconciler 阶段收集 Portal 子树
func (r *Reconciler) collectPortalSubtrees(root *Fiber) []PortalSubtree {
    var subtrees []PortalSubtree

    var collect func(fiber *Fiber)
    collect = func(fiber *Fiber) {
        if fiber == nil {
            return
        }

        // 检查是否是 Portal 节点
        if fiber.Props != nil {
            if portalRootID, ok := fiber.Props["portalRoot"].(string); ok && portalRootID != "" {
                // 提取 Portal 子树（从 fiber.Child 开始）
                children := extractPortalChildren(fiber)

                // 从主树中移除 Portal 节点（或保留 shell）
                removePortalFromParent(fiber)

                subtrees = append(subtrees, PortalSubtree{
                    portalRootID: portalRootID,
                    portalFiber:  fiber,
                    treeRoot:     buildIndependentFiberTree(children),
                })
            }
        }

        collect(fiber.Child)
        collect(fiber.Sibling)
    }

    collect(root)
    return subtrees
}

// Layout 层直接使用独立 Fiber 子树
func (e *PortalAwareLayoutEngine) Layout(fiber *reconciler.Fiber, ...) {
    // Phase 1: 主树 Layout（不需要 PortalAwareFiberToNodeAdapter）
    mainNode := NewFiberToNodeAdapterPure(fiber)
    mainResult := e.engine.Layout(mainNode, layoutConstraints)

    // Phase 2: 使用独立 Portal Fiber 子树
    for _, portal := range portalSubtrees {
        portalNode := NewFiberToNodeAdapterPure(portal.treeRoot)
        portalResult := e.engine.Layout(portalNode, layoutConstraints)
    }
}
```

---

## 三、为什么当前实现选择共享 Fiber？

### 🟢 当前实现的优势

| 优势 | 详细说明 |
|------|---------|
| **代码简洁** | 无需管理多棵 Fiber 树，Reconciler 逻辑统一 |
| **Diff 自然** | Portal 作为普通 Fiber 节点，diff 逻辑无需特殊处理 |
| **状态统一** | Portal 子节点直接访问状态，无需额外同步机制 |
| **易于调试** | 单一 Fiber 树，便于追踪执行流程和调试问题 |
| **性能稳定** | 避免双树 reconcile 的额外开销，性能可预测 |
| **渐进式演进** | 如果未来性能瓶颈明显，可以逐步迁移到独立子树 |

### 🔴 独立 Fiber 子树的优势

| 优势 | 详细说明 |
|------|---------|
| **架构清晰** | 严格遵循"主树 + Overlay 多树"的设计理念 |
| **物理隔离** | Portal 子树完全独立，不会污染主树逻辑 |
| **语义明确** | Fiber 树结构与渲染结构完全一致，便于理解 |
| **扩展性好** | 支持 Portal 虚拟化、按需挂载等高级特性 |

### 📊 改造成本对比

| 项目 | 共享 Fiber → 独立 Fiber | 工作量 | 风险 |
|------|------------------------|--------|------|
| **Reconciler** | 添加 Portal 子树收集、独立 reconcile | 3-5天 | 🔴 高 |
| **Diff** | 支持独立 Portal 子树 diff | 2-3天 | 🟡 中 |
| **Layout** | 移除 PortalAwareFiberToNodeAdapter | 1-2天 | 🟡 中 |
| **Focus** | 多树遍历 | 1天 | 🟡 中 |
| **测试** | 多场景覆盖 | 2-3天 | 🟡 中 |
| **性能优化** | Lane 优先级调整 | 1-2天 | 🟢 低 |
| **总计** | - | **10-15天** | 🔴 高 |

---

## 四、实现差异影响分析

### 4.1 需要额外处理的模块（共享 Fiber）

| 模块 | 处理方式 | 复杂度 |
|------|---------|--------|
| **Layout Adapter** | `PortalAwareFiberToNodeAdapter` 跳过 Portal 子节点 | 🟡 中 |
| **Portal Collector** | 收集 Portal 节点，转换子节点为 layout.Node | 🟡 中 |

### 4.2 不需要额外处理的模块（等价）

| 模块 | 理由 |
|------|------|
| **Render** | 基于 LayoutBox，不关心 Fiber 结构 |
| **Event** | 基于 LayoutBox 和 Z-order，不关心 Fiber 结构 |
| **State** | Instance 持久化，不依赖 Fiber 树结构 |

### 4.3 需要差异处理的模块（独立 Fiber）

| 模块 | 差异点 | 复杂度 |
|------|--------|--------|
| **Reconciler** | 多树 reconcile，Portal 子树收集逻辑 | 🔴 高 |
| **Diff** | 需要独立的 DiffContext | 🟡 中 |
| **Focus** | 多树遍历，Focus Transfer | 🟡 中 |
| **Lifecycle** | Portal 子树独立生命周期管理 | 🟡 中 |

---

## 五、兼容性与迁移路径

### 当前实现的兼容性

✅ **完全兼容设计文档的核心原则**:
- Fiber 结构保持不变（逻辑关系不变）
- Layout 父可重定向（通过 PortalRoot）
- Render 顺序基于挂载点（通过 Layer 排序）

### 迁移到独立 Fiber 的建议路径

如果未来需要迁移，建议采用渐进式方式：

**Phase 1**: 添加 Portal 子树收集逻辑（保留共享 Fiber）
```go
// Reconciler 层添加收集逻辑，但不改变 reconciliation
func (r *Reconciler) analyzePortalStructure(root *Fiber) []PortalAnalysis {
    // 仅分析，不改变 tree 结构
}
```

**Phase 2**: 添加独立 reconcile 选项（并行运行）
```go
// 双轨制：同时支持共享和独立模式
type ReconcilerConfig struct {
    EnableIndependentPortalSubtrees bool
}
```

**Phase 3**: 逐步切换（通过 feature flag）
```go
// 默认保持共享模式，可选开启独立模式
if config.EnableIndependentPortalSubtrees {
    // 使用独立 Fiber 子树
} else {
    // 使用共享 Fiber（当前默认）
}
```

**Phase 4**: 完全迁移（经过充分验证后）
```go
// 移除共享模式相关代码
// 集中维护独立 Fiber 子树实现
```

---

## 六、最终评估

### ✅ 推荐保持当前实现

**理由总结**:
1. ✅ **功能等价**: 共享 Fiber 实现完全满足所有功能需求
2. ✅ **性能稳定**: 单树 reconcile，性能开销可预测
3. ✅ **代码简洁**: 无需管理多树，维护成本低
4. ✅ **易于调试**: 单一树结构，便于问题追踪
5. ✅ **渐进式演进**: 未来如有需要，可以平滑迁移

### 📋 持续优化方向

推荐将精力投入到更有价值的优化上：
- ✅ Portal 动画（淡入/淡出、缩放）
- ✅ Portal 虚拟化（大量 tooltip 按需挂载）
- ✅ Anchor + Scroll 联动（更精确的定位）
- ✅ 性能监控和调优

---

## 七、关键原则（重申）

### Portal 核心原则（两种实现方式均遵循）

```
✅ Fiber 不动（逻辑关系保持不变）
✅ Layout 重建（Overlay 阶段独立计算）
✅ Render 分层（按 Layer 排序）
```

### 当前实现的变体说明

```
当前实现 = 共享 Fiber（逻辑层）+ 独立 LayoutBox（布局层）
         = "Fiber 共享" + "Layout 分离"
         = Pragmatic 方式（工程实践优化）
```

---

## 八、参考文档

| 文档 | 说明 |
|------|------|
| `docs/layout/portal/portal_design.md` | Portal 设计文档（独立 Fiber 子树方式） |
| `docs/layout/portal/portal_implementation_plan.md` | 当前实现进度（共享 Fiber 方式） |
| `docs/layout/portal/portal_architecture_summary.md` | 本文档（架构对比） |

---

## 九、常见问题（FAQ）

### Q1: 共享 Fiber 是否会导致 Portal 子节点在 diff 时出现问题？

**A**: 不会。Portal 子节点作为普通 Fiber 节点参与 diff，diff 逻辑基于 Fiber.NodeID、DiffKey 等标识符，与树结构无关。Layout 阶段通过 PortalAwareFiberToNodeAdapter 跳过 Portal 子节点，确保不影响主树布局。

### Q2: 共享 Fiber 是否会影响性能？

**A**: 不会。当前实现的单树 reconcile 性能开销低于双树 reconcile。Portal 的 LayoutBox 在 Overlay 阶段独立生成，不参与主树布局计算。经过测试，Portal 打开/关闭的性能表现良好。

### Q3: 什么情况下需要迁移到独立 Fiber 子树？

**A**: 如果出现以下情况，可以考虑迁移：
- Portal 子树规模非常大，需要虚拟化
- Portal 子树需要独立的 reconcile 优先级（高优先级同步处理）
- 主树和 Portal 子树的 diff 日志需要完全隔离

### Q4: 独立 Fiber 子树的实现难度如何？

**A**: 较高。需要修改 Reconciler、Diff、Focus 等多个核心模块，工作量约 10-15 天，风险较高。建议仅在必要时进行迁移。

---

## 十、关键原则

### Portal 本质

```
Portal = Fiber 不动，Layout 重建
Portal = 布局跳过 + 渲染重接
Portal = 结构存在 + 布局消失 + Overlay 重生
```

### 核心不变量

1. **Fiber 结构不能断**：保持逻辑父子关系（用于 state、context、生命周期）
2. **Layout 父可重定向**：Portal 的 LayoutBox 的 Parent 是 OverlayRoot，不是原父
3. **Render 顺序基于挂载点**：不是 Fiber 顺序，而是按 Layer 排序

---

## 二、架构路线

### 整体分层

```
VNode（声明层）
    ↓ reconcile
Fiber（调度层，含 PortalRoot）
    ↓ commit
LayoutBox（布局层，分两阶段）
    ↓
LayerManager（合成层）
    ↓
PaintableBox（渲染层）
    ↓
Renderer（终端输出）
```

### 单向数据流

```
① Fiber → LayoutBox（结构 → 几何）
② LayoutBox → PaintableBox（几何 → 渲染指令）
③ PaintableBox → Terminal（绘制）
```

### 四层职责边界

| 层级 | 职责 | 绝对不做 |
|------|------|----------|
| Fiber | 结构、状态、diff | 坐标、渲染、clip |
| LayoutBox | 坐标、尺寸、clip、scroll | draw、style、diff |
| PaintableBox | 字符、颜色、边框 | 布局、结构 |
| Layer | 优先级、渲染顺序、事件顺序 | 坐标、树结构 |

---

## 三、Portal 与 Layout 的处理关系

### Portal 在各层的表现

| 层级 | Portal 状态 |
|------|-------------|
| VNode | 容器节点，声明 PortalTarget |
| Fiber | 节点存在，Parent 不变，设置 PortalRoot |
| 主 Layout | **不存在**，不参与布局计算 |
| Overlay Layout | 重新生成 LayoutBox，Root 坐标系 |

### 两阶段 Layout

**阶段一：主树 Layout（忽略 Portal）**

```go
func layout(node *LayoutBox) {
    for _, child := range node.Children {
        if child.IsPortal {
            collectPortal(child)  // 只收集，不布局
            continue
        }
        layout(child)
    }
    computeSize(node)  // Portal 不参与尺寸计算
}
```

**阶段二：Overlay Layout（独立计算）**

```go
func layoutOverlay(rootW, rootH int) {
    for _, item := range portalQueue {
        node := item.Node
        // 直接使用 Root 坐标系，完全忽略 parent
        switch item.Fiber.Anchor {
        case AnchorCenter:
            node.AbsX = (rootW - node.W) / 2
            node.AbsY = (rootH - node.H) / 2
        // ... 其他锚点
        }
    }
}
```

### 坐标计算差异

| 类型 | 计算方式 |
|------|----------|
| 普通节点 | `node.AbsX = parent.AbsX + node.X` |
| Portal 节点（居中） | `(rootW - node.W) / 2, (rootH - node.H) / 2` |
| Tooltip 锚点 | `anchor.AbsX, anchor.AbsY + anchor.H` |

### Fiber 层的 Layout 重定向

```go
func layout(f *Fiber, parent *LayoutBox, root *LayoutBox) {
    node := f.StateNode
    var layoutParent *LayoutBox

    if f.PortalRoot != nil {
        layoutParent = f.PortalRoot.StateNode   // 跳到 Overlay
    } else {
        layoutParent = parent
    }

    computeLayout(node, layoutParent, root)
}
```

### Portal 在 flex/stack 中的行为

**示例**：
```
Flex(
    Box(W:10),
    Portal("modal", Box(W:40, H:10)),
    Box(W:20),
)
```

**实际计算**：
- Flex 参与布局的子元素：Box(10) + Box(20)
- Portal 完全透明，不占空间
- flexWidth = 30（不是 50）

---

## 四、Overlay 子系统

### OverlayManager 结构

```go
type OverlayEntry struct {
    ID            string
    Box          *LayoutBox
    PortalRootID string
    Priority     int     // Z-order
    Active       bool
    Fiber        *Fiber  // 关联Fiber
}

type OverlayManager struct {
    stack   []*OverlayEntry  // 按优先级排序
    entries map[string]*OverlayEntry
}
```

### Z-index 分层

```
Layer 0   → Base UI
Layer 10  → Dropdown
Layer 20  → Tooltip
Layer 100 → Modal
Layer 200 → Toast
```

### 多 Portal 布局策略

**原则**：每个 Portal 独立 layout，共享 Root 坐标系

| 定位模式 | 计算公式 |
|---------|----------|
| Modal（全局居中） | `(rootW - W)/2, (rootH - H)/2` |
| Tooltip（锚点） | `anchor.AbsX, anchor.AbsY + H` |
| Toast（堆叠） | 累加 Y 坐标 |

### Layer 合成

```go
func buildLayers() {
    // 主树
    addToLayer(0, mainBoxes)

    // Overlay（按 Z 排序）
    for _, entry := range sortedEntries {
        layer := layerManager.Get(entry.Z)
        layer.Boxes = append(layer.Boxes, entry.PaintBoxes...)
    }
}
```

---

## 五、注意事项

### 必须满足的 3 个不变量

- [ ] Fiber Parent 不变
- [ ] Layout Parent 可变
- [ ] Render 顺序基于挂载点

### 常见错误（必须避免）

| 错误 | 后果 |
|------|------|
| 只在 Render 做 Portal | layout 错、clip 错 |
| `modal.Parent = overlay` | state 丢失、diff 崩 |
| Portal 不分 Layer | z-index 混乱 |
| Fixed 不配 Portal | 被 clip、被 scroll |
| 在主树生成 Portal LayoutBox | 污染 flow、影响父尺寸 |

### Portal ≠ Layer

| 能力 | Portal | Layer |
|------|--------|-------|
| 脱离父布局 | ✅ | ❌ |
| 脱离 scroll | ✅ | ❌ |
| 顶层显示 | ❌ | ✅ |
| 事件优先 | ❌ | ✅ |
| 渲染顺序 | ❌ | ✅ |

**正确的组合**：
```go
// Portal决定坐标系
if fiber.PortalRoot != nil {
    fiber.PortalRoot = overlayRoot
}

// Layer决定渲染顺序
fiber.Layer = LayerOverlay
if fiber.PortalRoot != nil {
    fiber.Layer = LayerModal // 自动提升
}
```

### 三个禁止的反向依赖

```
Paintable ❌→ Layout
Layout    ❌→ Fiber（除只读）
```

---

## 六、事件系统

### 命中检测顺序

```
Input
  ↓
Overlay（从上到下，Z 最大的先命中）
  ↓（如果没消费）
Main Tree
```

### Focus 系统

**正确**：基于 Fiber Tree
```go
focusNext(fiberTree)
```

**错误**：基于 flattenNodes（会打断逻辑焦点流）

### Modal 阻断

```go
if overlayActive {
    ignoreBelowOverlay()
}
```

---

## 七、实践步骤

### Step 1: VNode 声明

```go
VNode{
    Type: "Modal",
    Props: map[string]interface{}{
        "portal": "overlay",  // 指定目标层
    },
    Children: children,
}
```

### Step 2: Reconcile 解析

```go
func createFiber(v VNode, parent *Fiber) *Fiber {
    f := &Fiber{
        Type:   v.Type,
        Parent: parent,
    }
    if target, ok := v.Props["portal"]; ok {
        f.PortalRoot = resolvePortalTarget(target)
    }
    return f
}
```

### Step 3: Layout 主树（收集 Portal）

```go
var portalQueue []PortalItem

func layout(node *LayoutBox) {
    for _, child := range node.Children {
        if child.IsPortal {
            portalQueue = append(portalQueue, PortalItem{Node: child})
            continue
        }
        layout(child)
    }
}
```

### Step 4: Overlay Layout

```go
func layoutOverlay(rootW, rootH int) {
    for _, item := range portalQueue {
        buildLayoutTree(item.Node)
        layoutPortalNode(item.Node, rootW, rootH)
    }
}
```

### Step 5: Flatten

```go
func flatten(f *Fiber, layers map[Layer][]*LayoutBox) {
    node := f.StateNode
    layer := LayerNormal
    if f.PortalRoot != nil {
        layer = LayerOverlay
    }
    layers[layer] = append(layers[layer], node)
}
```

### Step 6: Layer 排序

```go
final := append(layers[LayerNormal], layers[LayerOverlay]...)
```

### Step 7: 渲染

```go
for _, layer := range layersSortedByZ {
    draw(layer.Boxes)
}
```

---

## 八、实现检查清单

- [ ] Fiber Parent 不变
- [ ] Layout Parent 可变（Portal）
- [ ] Flatten 分层
- [ ] Layer 控制顺序
- [ ] Clip 不继承父（Portal 使用 root.Clip）
- [ ] Scroll 不影响 Portal
- [ ] Event 逆序命中
- [ ] Diff 不感知 Portal（只基于 Fiber Tree）

---

## 九、多 Portal 模型

```
多 Portal = 多棵独立 Layout Tree + 单一 Layer 合成 + Z 排序 + 事件反向分发
```

### OverlayManager 管理

- **注册**：`Push(id, box, portalRootID, priority)`
- **获取**：`GetAll()` 按优先级排序
- **移除**：`Remove(id)` 或 `Pop()`
- **堆栈**：高优先级在前

---

## 十、性能优化

### Portal 独立 Diff

```go
func (r *Reconciler) diffOverlay(overlay *OverlayEntry, newFiber *Fiber) {
    diff(overlay.FiberRoot, newFiber)
}
```

### 脏区合并

```go
dirtyRects = merge(mainDirty, overlayDirty)
```

### 局部渲染

```go
for _, rect := range dirtyRects {
    redraw(rect)
}
```

Portal 更新不触发全屏重绘。
