# Mint TUI 架构问题分析

## 1. 核心设计原则（来自 idea2_layout.md）

1. **布局必须完全从"组件行为"中剥离**
2. **两阶段协议**：
   - Phase 1: Measure (父给约束，子返回尺寸)
   - Phase 2: Layout (父决定子最终坐标)
3. **组件只提供 Measure 和 Paint，不知道自己的位置**
4. **渲染管线应该是**：
   ```
   VNode → Reconcile → RNode → Layout Engine → Paint Engine
   ```

## 2. 当前实现的架构缺陷

### 2.1 Paint 和 Layout 混合

**位置**：`internal/render/declarative_node.go` PaintVNode

```go
if isHStack {
    childX := x
    for _, child := range children {
        n.PaintVNode(child, childX, y, buf)  // 直接绘制！
        childX += n.MeasureVNodeWidth(child) + gap  // 边计算边画
    }
}
```

**问题**：
- 位置计算 (`childX`) 和绘制 (`PaintVNode`) 在同一个循环中
- 没有独立的 Layout 阶段预先计算所有位置
- 违反了 "Layout 是纯计算阶段，不可有副作用" 的原则

### 2.2 没有独立的 Layout Engine

**设计要求**：
```
VNode → Reconcile → RNode → Layout Engine → Paint Engine
```

**当前实现**：
```
VNode → 直接 PaintVNode (没有中间的 Layout Engine)
```

**缺失**：
- 没有 Layout Tree
- 没有统一的 Layout 阶段
- 位置计算散布在 Paint 逻辑中

### 2.3 Measure 接口没有被正确使用

虽然组件实现了 `Measure(constraints) Size`：
```go
type Measurable interface {
    Measure(constraints runtime.BoxConstraints) runtime.Size
}
```

但绘制时使用的是：
```go
childWidth := n.MeasureVNodeWidth(child)  // 没有使用约束！
```

**问题**：
- 没有父容器传给子组件的约束
- 子组件不知道父容器的可用空间
- 无法实现 "父给约束 → 子返回尺寸 → 父再排布" 的单向数据流

### 2.4 边框组件的职责混乱

**位置**：`internal/render/declarative_node.go` paintBordered

```go
func (n *DeclarativeNode) paintBordered(vnode rtui.VNode, ...) {
    contentWidth := n.MeasureVNodeWidth(child)   // 测量
    contentHeight := n.MeasureVNodeHeight(child)
    renderer.Paint(x, y, contentWidth, contentHeight, ...)  // 绘制边框
    n.PaintVNode(child, x+offsetX, y+offsetY, buf)  // 绘制内容
}
```

**问题**：
- 边框既是布局组件（有尺寸），又是装饰组件（需要绘制）
- 边框的 Measure 和 Paint 职责没有清晰分离
- 导致位置计算和绘制逻辑耦合

## 3. 当前 Bug 的直接原因

**调试输出**：
```
[HSTACK] child 1 (tag=bordered): x=17, width=32, nextX=50
[BORDER.Paint] cornerTL at (17,3): '┌'
```

**ANSI 输出**：
```
[4;18H┌  ← 实际在 column 18！
```

**原因分析**：

1. HStack 计算 contentArea 应该在 x=17
2. 但实际边框的左上角出现在 x=18
3. 说明 `paintBordered` 或边框绘制逻辑有额外的 +1 偏移

**可能的 Bug 位置**：
- `BorderedNode.Measure` 返回的宽度是否正确？
- `paintBordered` 中的 `offsetX` 是否正确？
- 边框和内容的位置关系是否一致？

## 4. 设计要求 vs 当前实现对比

| 设计要求 | 当前实现 | 状态 |
|---------|---------|------|
| 独立的 Layout 阶段 | 没有，和 Paint 混合 | ❌ |
| Layout Tree | 没有 | ❌ |
| 约束驱动的布局 | 有 Measure 接口但未使用 | ❌ |
| 组件不知道自己的位置 | Paint 时传入 x,y | ❌ |
| Layout 是纯计算 | Paint 时计算位置 | ❌ |
| 增量布局 (LayoutDirty) | 没有实现 | ❌ |
| Layout 和 Paint 分离 | 混在一起 | ❌ |

## 5. 重构方向

### 5.1 创建独立的 Layout Engine

```go
type LayoutEngine struct {
    // 约束驱动的布局计算
}

func (e *LayoutEngine) Layout(root VNode, constraints BoxConstraints) LayoutTree {
    // 预先计算所有节点位置
    // 返回 LayoutTree，包含每个节点的 Box (x, y, width, height)
}
```

### 5.2 分离 Layout 和 Paint

```go
// Layout 阶段：只计算位置
layoutTree := LayoutEngine.Layout(vnode, constraints)

// Paint 阶段：只绘制，不再计算位置
PaintEngine.Paint(layoutTree, buffer)
```

### 5.3 边框组件职责分离

```go
// BorderedNode 只负责：
// 1. Measure: 返回包含边框的总尺寸
// 2. 提供 GetContentOffset() 返回内容偏移

// 边框的绘制由 Layout Engine 或 Paint Engine 负责
```

## 6. 短期修复方案

在不重构整个架构的情况下，修复当前的 bug：

1. 在 `paintBordered` 中添加调试，确认坐标计算
2. 检查 `BorderedNode.Measure` 返回的宽度是否正确
3. 确保 `offsetX` 与边框绘制逻辑一致

## 7. 长期重构方案

按照设计文档的要求，实现完整的 Layout Engine：

1. 创建 Layout Tree 数据结构
2. 实现 Layout Engine，基于约束计算所有节点位置
3. 将 Layout 阶段从 Paint 阶段分离
4. 实现 LayoutDirty 标记，支持增量布局
