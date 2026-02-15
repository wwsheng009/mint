# Runtime Layout 系统优化实施方案

## 执行摘要

本方案旨在基于 `runtime/layout/` 的现有实现（V3），吸收 `runtime/compute/` 中的通用布局优化策略，增强布局引擎的性能和功能，同时保持其独立性（不依赖 VNode/Fiber）。

**当前状态（V3）**：
- ✅ 标准的 Flexbox 实现（flex-grow、flex-shrink、flex-basis）
- ✅ 支持反向排列（RowReverse、ColumnReverse）
- ✅ 支持主轴/交叉轴对齐（6 种主轴 + 4 种交叉轴）
- ✅ 基础缓存系统（基于 SHA256 树哈希）
- ✅ 简单的脏标记机制（Dirtyable 接口和 Engine.dirtyNodes）
- ✅ 清晰的接口设计（Node、Measurable、Dirtyable）

**优化目标**：
1. **增强缓存策略** - 实现 Flex 分布缓存、优化叶子节点缓存
2. **完善 Dirty Tracking** - 基于 Dirtyable 接口实现增量布局
3. **优化测量算法** - 减少重复测量，提升性能
4. **扩展布局功能** - 添加 HitMap、边界验证、文本对齐等
5. **提升可维护性** - 改进代码结构，增强可测试性

---

## 1. 当前架构分析

### 1.1 代码结构

```
runtime/layout/
├── types.go          # 核心类型定义 (Node、Engine、Constraints 等)
├── flex.go           # Flexbox 布局算法实现
├── cache.go          # 布局缓存系统（基于 SHA256）
├── README.md         # 使用文档
├── *_test.go         # 测试文件
```

### 1.2 核心功能现状

#### 核心类型（types.go）

```go
// Node - 布局节点接口
type Node interface {
    ID() string
    Type() string
    Children() []Node
    GetPosition/SetPosition
    GetSize/SetSize
    GetWidth/GetHeight
}

// Measurable - 可测量节点
type Measurable interface {
    Node
    Measure(constraints Constraints) Size
}

// Dirtyable - 脏标记节点
type Dirtyable interface {
    IsLayoutDirty() bool
    ClearLayoutDirty()
    MarkLayoutDirty()
}

// Engine - 布局引擎
type Engine struct {
    dirtyNodes map[string]bool  // 脏节点集合
    stats      LayoutStats      // 统计信息
    cache      *Cache           // 缓存实例
}

// LayoutBox - 布局结果盒子
type LayoutBox struct {
    ID      string
    X, Y    int
    Width, Height int
    Baseline int
    Children []*LayoutBox
}
```

**优势**：
- ✅ 接口设计清晰，职责分离
- ✅ 支持可选接口扩展（Measurable、Dirtyable）
- ✅ 内置缓存和统计功能
- ✅ LayoutBox 支持树形结构

#### Flexbox 实现（flex.go）

**完整支持**：
- ✅ `FlexDirection`：Row, Column, RowReverse, ColumnReverse
- ✅ `MainAxisAlignment`：Start, End, Center, SpaceBetween, SpaceAround, SpaceEvenly
- ✅ `CrossAxisAlignment`：Start, End, Center, Stretch
- ✅ `Flex` 配置：Grow, Shrink, Basis
- ✅ `Padding` 和 `Gap` 支持
- ✅ `LayoutChildren` 方法实现子节点布局（带位置和尺寸分配）

**测量流程（`Measure` 方法）**：
```
Phase 1: 测量所有子节点
  ↓
Phase 2: 计算固定尺寸总和和可伸缩因子
  ↓
Phase 3: 添加间距（Gap）
  ↓
Phase 4: 计算总尺寸（包含 Padding）
```

**布局流程（`LayoutChildren` 方法）**：
```
Phase 1: 测量所有子节点并识别 flex/fixed
  ↓
Phase 2: 计算剩余空间
  ↓
Phase 3: 分配剩余空间给可伸缩节点（根据 flex grow）
  ↓
Phase 4: 根据对齐方式计算主轴位置
  ↓
Phase 5: 根据对齐方式计算交叉轴位置（含 Stretch）
  ↓
Phase 6: 处理 SpaceBetween/Around/Evenly 的额外间距
  ↓
Phase 7: 布局每个子节点（处理 reverse、间距、stretch）
```

**问题**：
- ⚠️ `childConstraints` 方法的 `index` 参数未使用
- ⚠️ 未利用 Flex 分布缓存优化
- ⚠️ 可伸缩节点的 `shrink` 和 `basis` 在测量时未完全利用

#### 缓存系统（cache.go）

**当前实现**：
```go
type Cache struct {
    entries map[string]*CachedLayout
    maxSize int
}

type CachedLayout struct {
    Result     *LayoutResult
    Timestamp  time.Time
    HitCount   int
}
```

**缓存键生成**：
- 节点树 SHA256 哈希（递归所有子节点）
- 约束条件转字符串（当前实现有 bug）

**问题**：
- ❌ `constraintsKey` 使用 `string(rune())` 转换错误（会截断）
- ❌ 完整树哈希计算成本高
- ❌ 未利用 `Dirtyable` 接口的脏标记
- ❌ 未优化叶子节点缓存
- ❌ `cacheKey` 类型定义但未使用
- ❌ 缺少按节点 ID 失效的功能

### 1.3 当前缺失功能

| 功能 | 状态 | 优先级 | 来源参考 |
|-----|------|--------|---------|
| **性能优化** |
| Flex 分布缓存 | ❌ 缺失 | 高 | compute/types.go |
| 叶子节点优化缓存 | ❌ 缺失 | 高 | compute/cache.go |
| 增量布局（Dirty Tracking） | ⚠️ 基础支持 | 高 | compute/dirty_tracker.go |
| 单次测量优化 | ⚠️ 部分实现 | 中 | compute/engine.go |
| **布局功能** |
| HitMap（事件处理） | ❌ 缺失 | 中 | compute/engine.go |
| 边界验证 | ❌ 缺失 | 低 | compute/bounds_validator.go |
| 表格布局 | ❌ 缺失 | 中 | compute/engine.go |
| 文本对齐增强 | ❌ 缺失 | 中 | compute/engine.go |
| **接口设计** |
| 稳定标识符 | ❌ 缺失 | 中 | 抽象设计 |

---

## 2. 可移植的通用功能

### 2.1 从 compute/ 可移植的功能

基于 `runtime/compute/` 的代码分析，以下功能可以移植到 `runtime/layout/`，因为它们不依赖 VNode/Fiber：

| 模块 | 可移植内容 | 适配代价 |
|-----|-----------|---------|
| **缓存优化** |
| Flex 分布缓存 | `FlexDistributionInfo` 结构和缓存逻辑 | 低 |
| 叶子节点缓存 | 只缓存叶子节点和有 key 的节点策略 | 中 |
| 约束键优化 | 正确的约束键生成（避免 rune 转换） | 低 |
| **Dirty Tracking** |
| DirtyTracker 结构 | `dirty map[string]bool` + 并发保护 | 低 |
| 脏标记传播 | 向上传播到父节点的逻辑 | 中 |
| 增量布局 | 仅重新布局脏子树 | 中 |
| **布局算法** |
| HStack/VStack 参考 | 纯布局算法（去除 VNode 绑定） | 低 |
| Flex 分布算法 | 智能分配剩余空间的逻辑 | 低 |
| **验证和调试** |
| 边界验证 | BoundsValidator（抽象化） | 低 |
| 统计信息 | 扩展 LayoutStats | 低 |
| **事件处理** |
| HitMap | 将 LayoutBox 映射到可交互区域 | 中 |

### 2.2 不可移植的功能

以下功能**不应移植**，因为它们紧密依赖 VNode/Fiber：

- ❌ VNode 相关接口（直接使用 VNode 类型）
- ❌ Fiber 相关功能（NodeID、Layer、DiffKey）
- ❌ 组件特定的测量（文本、边框、表格等）
- ❌ rtui 特定的约束系统

**替代方案**：
- 通过适配器模式接入 VNode/Fiber（在应用层实现）
- 保留纯布局引擎职责
- 提供通用接口供应用层扩展

---

## 3. 优化方案

### 3.1 缓存系统增强

#### 目标
- 修复约束键生成的 bug
- 添加 Flex 分布缓存
- 实现叶子节点优化策略
- 支持按节点 ID 失效

#### 实现计划

**1. 修复 constraintsKey**

```go
// 当前（错误）
func (c *Cache) constraintsKey(constraints Constraints) string {
    return string(rune(constraints.MinWidth)) + "," +
           string(rune(constraints.MaxWidth)) + "," +
           string(rune(constraints.MinHeight)) + "," +
           string(rune(constraints.MaxHeight))
}

// 优化后
func (c *Cache) constraintsKey(constraints Constraints) string {
    return fmt.Sprintf("%d,%d,%d,%d",
        constraints.MinWidth, constraints.MaxWidth,
        constraints.MinHeight, constraints.MaxHeight)
}
```

**2. 添加 Flex 分布缓存**

```go
// FlexDistributionInfo 记录 Flex 布局的分布信息
type FlexDistributionInfo struct {
    TotalFlexFactor int       // flex-grow 总和
    FixedSize       int       // 固定尺寸子节点总尺寸
    ChildCount      int       // 子节点数量
    Valid           bool      // 是否有效
    Version         uint64    // 版本号（用于失效）
}

// FlexCache Flex 分布缓存
type FlexCache struct {
    entries map[string]*FlexDistributionInfo
    mu      sync.RWMutex
}

// Get 获取或计算 Flex 分布信息
func (fc *FlexCache) Get(
    nodeID string,
    children []Node,
    flexibleIndices []int,
    isRow bool,
    computeFunc func() *FlexDistributionInfo,
) *FlexDistributionInfo {
    key := fmt.Sprintf("%s:%d:%v", nodeID, len(children), isRow)

    fc.mu.RLock()
    if info, ok := fc.entries[key]; ok && info.Valid {
        fc.mu.RUnlock()
        return info
    }
    fc.mu.RUnlock()

    fc.mu.Lock()
    defer fc.mu.Unlock()

    info := computeFunc()
    fc.entries[key] = info
    return info
}

// Invalidate 失效指定节点的缓存
func (fc *FlexCache) Invalidate(nodeID string) {
    fc.mu.Lock()
    defer fc.mu.Unlock()

    for key, info := range fc.entries {
        if strings.HasPrefix(key, nodeID+":") {
            info.Valid = false
            delete(fc.entries, key)
        }
    }
}
```

**3. 实现叶子节点优化缓存**

```go
// Cache 添加叶子节点检测
func (c *Cache) isLeafNode(node Node) bool {
    return len(node.Children()) == 0
}

// Get 添加叶子节点优先逻辑
func (c *Cache) Get(node Node, constraints Constraints) *LayoutResult {
    // 叶子节点或实现了特定接口（如缓存键）的节点才使用缓存
    if !c.isLeafNode(node) {
        return nil
    }

    key := c.makeKey(node, constraints)
    if entry, ok := c.entries[key]; ok {
        entry.HitCount++
        return c.cloneResult(entry.Result)
    }
    return nil
}
```

**4. 支持按节点 ID 失效**

```go
// RemoveByNode 删除特定节点的缓存（优化实现）
func (c *Cache) RemoveByNode(id string) {
    nodeIDs := c.extractNodeIDs(id)

    for key := range c.entries {
        keyNodeID := c.extractNodeIDFromKey(key)
        for _, nid := range nodeIDs {
            if keyNodeID == nid {
                delete(c.entries, key)
                break
            }
        }
    }
}

// extractNodeIDs 从节点 ID 提取相关 ID（包括子节点）
func (c *Cache) extractNodeIDs(nodeID string) []string {
    // 实现：返回该节点及其所有祖先/子节点的 ID
    // 这里简化处理，实际需要配合树结构
    return []string{nodeID}
}
```

#### 文件变更

```
runtime/layout/
├── cache.go              # 修改：修复缓存键、添加叶子节点优化
├── flex_cache.go         # 新增：Flex 分布缓存
└── types.go              # 修改：Engine 添加 FlexCache 字段
```

---

### 3.2 Dirty Tracking 完善

#### 目标
- 完善 Dirtyable 接口的集成
- 实现脏标记传播
- 支持增量布局（仅重新布局脏子树）
- 添加子树脏标记功能

#### 实现计划

**1. 扩展 DirtyTracker**

```go
// DirtyTracker 脏标记跟踪器
type DirtyTracker struct {
    mu    sync.RWMutex
    dirty map[string]bool
}

// NewDirtyTracker 创建新的脏标记跟踪器
func NewDirtyTracker() *DirtyTracker {
    return &DirtyTracker{
        dirty: make(map[string]bool),
    }
}

// MarkLayoutDirty 标记节点为脏
func (dt *DirtyTracker) MarkLayoutDirty(id string) {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    dt.dirty[id] = true
}

// MarkSubtreeDirty 标记整个子树为脏
func (dt *DirtyTracker) MarkSubtreeDirty(node Node) {
    dt.markRecursive(node)
}

func (dt *DirtyTracker) markRecursive(node Node) {
    if node == nil {
        return
    }
    dt.MarkLayoutDirty(node.ID())
    for _, child := range node.Children() {
        dt.markRecursive(child)
    }
}

// IsLayoutDirty 检查节点是否需要布局
func (dt *DirtyTracker) IsLayoutDirty(id string) bool {
    dt.mu.RLock()
    defer dt.mu.RUnlock()
    return dt.dirty[id]
}

// Clear 清除所有脏标记
func (dt *DirtyTracker) Clear() {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    dt.dirty = make(map[string]bool)
}

// ClearKey 清除特定节点的脏标记
func (dt *DirtyTracker) ClearKey(id string) {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    delete(dt.dirty, id)
}
```

**2. 增强 LayoutBox 支持脏标记传播**

```go
// LayoutBox 添加脏标记传播
func (lb *LayoutBox) PropagateDirtyUpwards() {
    lb.LayoutDirty = true
    // 向上传播逻辑在 Engine 中实现
}
```

**3. Engine 实现增量布局**

```go
// Layout 增量布局实现
func (e *Engine) Layout(root Node, constraints Constraints) *LayoutResult {
    if root == nil {
        return &LayoutResult{}
    }

    result := &LayoutResult{
        Boxes: make([]LayoutBox, 0),
        Dirty: true,
    }

    // 检查是否需要增量布局
    if e.needsIncrementalLayout(root) {
        box := e.layoutIncremental(root, constraints, 0, 0)
        result.Root = box
    } else {
        // 完整布局（同现有实现）
        box := e.layoutNode(root, constraints, 0, 0)
        result.Root = box
    }

    result.Boxes = e.collectBoxes(result.Root)

    // 清除脏标记
    e.clearDirtyMarkers(result.Root)

    return result
}

// needsIncrementalLayout 检查是否需要增量布局
func (e *Engine) needsIncrementalLayout(root Node) bool {
    // 检查根节点是否有脏标记
    if dirtyable, ok := root.(Dirtyable); ok && dirtyable.IsLayoutDirty() {
        return true
    }
    // 检查引擎中是否有脏节点
    if len(e.dirtyNodes) > 0 {
        return true
    }
    return false
}

// layoutIncremental 增量布局
func (e *Engine) layoutIncremental(node Node, constraints Constraints, x, y int) *LayoutBox {
    box := &LayoutBox{
        ID:      node.ID(),
        X:       x,
        Y:       y,
        Children: make([]*LayoutBox, 0),
    }

    // 检查节点是否脏
    needsLayout := false
    if dirtyable, ok := node.(Dirtyable); ok {
        needsLayout = dirtyable.IsLayoutDirty()
    } else if e.dirtyNodes[node.ID()] {
        needsLayout = true
    }

    if needsLayout {
        // 重新测量和布局
        width, height := node.GetSize()
        if measurable, ok := node.(Measurable); ok {
            size := measurable.Measure(constraints)
            width, height = size.Width, size.Height
        }
        box.Width = width
        box.Height = height

        node.SetPosition(x, y)
        node.SetSize(width, height)

        // 递归布局子节点
        childX, childY := x, y
        for _, child := range node.Children() {
            childBox := e.layoutIncremental(child, constraints, childX, childY)
            if childBox != nil {
                box.Children = append(box.Children, childBox)
                childY += childBox.Height
            }
        }
    } else {
        // 复用缓存或仅更新位置
        // 这里简化处理，实际需要缓存支持
        e.layoutIncrementalReuse(box, node, constraints, x, y)
    }

    return box
}

// clearDirtyMarkers 清除脏标记
func (e *Engine) clearDirtyMarkers(box *LayoutBox) {
    if box == nil {
        return
    }
    box.LayoutDirty = false

    for _, child := range box.Children {
        e.clearDirtyMarkers(child)
    }

    // 清除引擎中的脏标记
    if dirtyable, ok := box; ok {
        if d, k := dirtyable.(interface{ ClearLayoutDirty() }); k {
            d.ClearLayoutDirty()
        }
    }
}
```

#### 文件变更

```
runtime/layout/
├── dirty.go              # 新增：DirtyTracker 完整实现
├── engine.go             # 修改（types.go）：添加增量布局逻辑
└── types.go              # 修改：LayoutBox 添加 LayoutDirty 字段
```

---

### 3.3 测量算法优化

#### 目标
- 减少 Flex 布局中的重复测量
- 利用 Flex 分布缓存
- 优化 childConstraints 计算
- 改进反向排列的实现

#### 实现计划

**1. 优化 FlexLayout.Measure**

```go
// Measure 优化版本
func (f *FlexLayout) Measure(constraints Constraints) Size {
    if len(f.children) == 0 {
        width := constraints.ConstrainWidth(f.style.Padding.Left + f.style.Padding.Right)
        height := constraints.ConstrainHeight(f.style.Padding.Top + f.style.Padding.Bottom)
        return Size{Width: width, Height: height}
    }

    isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse

    // 使用 Flex 分布缓存
    flexIndices := f.getFlexibleIndices()
    flexInfo := f.engine.flexCache.Get(
        f.id, f.children, flexIndices, isRow,
        func() *FlexDistributionInfo {
            return f.computeFlexDistribution(constraints, isRow)
        },
    )

    // Phase 2: 添加间距
    gapCount := len(f.children) - 1
    if gapCount > 0 {
        flexInfo.FixedSize += f.style.Gap * gapCount
    }

    // 计算总尺寸
    var width, height int
    if isRow {
        width = f.style.Padding.Left + flexInfo.FixedSize + f.style.Padding.Right
        height = f.style.Padding.Top + flexInfo.MaxCrossSize + f.style.Padding.Bottom
    } else {
        width = f.style.Padding.Left + flexInfo.MaxCrossSize + f.style.Padding.Right
        height = f.style.Padding.Top + flexInfo.FixedSize + f.style.Padding.Bottom
    }

    return Size{
        Width:  constraints.ConstrainWidth(width),
        Height: constraints.ConstrainHeight(height),
    }
}

// getFlexibleIndices 获取可伸缩子节点索引
func (f *FlexLayout) getFlexibleIndices() []int {
    indices := make([]int, 0)
    for i := range f.children {
        if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
            indices = append(indices, i)
        }
    }
    return indices
}

// computeFlexDistribution 计算 Flex 分布（缓存的核心逻辑）
func (f *FlexLayout) computeFlexDistribution(constraints Constraints, isRow bool) *FlexDistributionInfo {
    childSizes := make([]Size, len(f.children))
    totalMainSize := 0
    maxCrossSize := 0
    flexTotalFactor := 0

    for i, child := range f.children {
        childConstraints := f.childConstraints(constraints, i)
        if measurable, ok := child.(Measurable); ok {
            childSizes[i] = measurable.Measure(childConstraints)
        } else {
            childSizes[i] = Size{Width: childConstraints.MinWidth, Height: childConstraints.MinHeight}
        }

        if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
            flexTotalFactor += flex.Grow
            basis := flex.Basis
            if basis == 0 {
                basis = childSizes[i].Width
            }
            totalMainSize += basis
        } else {
            if isRow {
                totalMainSize += childSizes[i].Width
            } else {
                totalMainSize += childSizes[i].Height
            }
        }

        if isRow {
            if childSizes[i].Height > maxCrossSize {
                maxCrossSize = childSizes[i].Height
            }
        } else {
            if childSizes[i].Width > maxCrossSize {
                maxCrossSize = childSizes[i].Width
            }
        }
    }

    return &FlexDistributionInfo{
        TotalFlexFactor: flexTotalFactor,
        FixedSize:       totalMainSize,
        ChildCount:      len(f.children),
        MaxCrossSize:    maxCrossSize,
        Valid:           true,
        Version:         f.version,
    }
}
```

**2. 优化 childConstraints**

```go
// childConstraints 移除未使用的 index 参数并优化
func (f *FlexLayout) childConstraints(constraints Constraints) Constraints {
    isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse

    // 减去内边距
    availableMain := constraints.MaxWidth - f.style.Padding.Left - f.style.Padding.Right
    availableCross := constraints.MaxHeight - f.style.Padding.Top - f.style.Padding.Bottom

    // 可用空间非负保证
    if availableMain < 0 {
        availableMain = 0
    }
    if availableCross < 0 {
        availableCross = 0
    }

    if isRow {
        return NewConstraints(0, availableMain, 0, availableCross)
    }
    return NewConstraints(0, availableCross, 0, availableMain)
}
```

**3. 改进反向排列的实现**

当前 `LayoutChildren` 中的反向排列逻辑有重复代码，可优化：

```go
// LayoutChildren 优化：提取布局逻辑
func (f *FlexLayout) LayoutChildren(width, height int) []LayoutBox {
    if len(f.children) == 0 {
        return nil
    }

    isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse

    // ... Phase 1-3: 计算尺寸和 Flex 分布同现有实现 ...

    // Phase 4+: 统一处理位置计算
    boxes := make([]LayoutBox, len(f.children))
    mainPos, crossPos := f.calculateStartPositions(availableWidth, availableHeight, finalSizes, isRow)
    extraGap := f.calculateSpaceAround(remainingSpace, len(f.children), availableWidth, availableHeight, isRow)

    // 布局每个子节点
    for i := range f.children {
        actualIndex := i
        if f.isReverse() {
            actualIndex = len(f.children) - 1 - i
        }

        x, y := f.calculateChildPosition(actualIndex, mainPos, crossPos, isRow, finalSizes[actualIndex])

        // 应用 stretch
        if f.style.CrossAxis == Stretch {
            f.applyStretch(&finalSizes[actualIndex], availableWidth, availableHeight, isRow)
        }

        boxes[i] = LayoutBox{
            ID:     f.children[actualIndex].ID(),
            X:      x,
            Y:      y,
            Width:  finalSizes[actualIndex].Width,
            Height: finalSizes[actualIndex].Height,
        }

        f.children[actualIndex].SetPosition(x, y)
        f.children[actualIndex].SetSize(finalSizes[actualIndex].Width, finalSizes[actualIndex].Height)

        // 更新主轴位置
        if isRow {
            mainPos += finalSizes[actualIndex].Width + f.style.Gap
        } else {
            mainPos += finalSizes[actualIndex].Height + f.style.Gap
        }

        if extraGap > 0 && i < len(f.children)-1 {
            mainPos += extraGap
        }
    }

    return boxes
}

// isReverse 检查是否为反向排列
func (f *FlexLayout) isReverse() bool {
    return f.style.Direction == FlexRowReverse || f.style.Direction == FlexColumnReverse
}

// calculateStartPositions 计算起始位置
func (f *FlexLayout) calculateStartPositions(availW, availH int, sizes []Size, isRow bool) (int, int) {
    fixedTotal, maxCross := f.calculateFixedCross(sizes, isRow)
    gapTotal := f.style.Gap * (len(sizes) - 1)

    var mainPos, crossPos int

    switch f.style.MainAxis {
    case MainStart:
        mainPos = 0
    case MainEnd:
        if isRow {
            mainPos = availW - fixedTotal - gapTotal
        } else {
            mainPos = availH - fixedTotal - gapTotal
        }
    case Center:
        if isRow {
            mainPos = (availW - fixedTotal - gapTotal) / 2
        } else {
            mainPos = (availH - fixedTotal - gapTotal) / 2
        }
    case SpaceBetween, SpaceAround, SpaceEvenly:
        mainPos = 0
    }

    switch f.style.CrossAxis {
    case CrossStart:
        crossPos = 0
    case CrossEnd:
        if isRow {
            crossPos = availH - maxCross
        } else {
            crossPos = availW - maxCross
        }
    case CrossCenter:
        if isRow {
            crossPos = (availH - maxCross) / 2
        } else {
            crossPos = (availW - maxCross) / 2
        }
    case Stretch:
        crossPos = 0
    }

    return mainPos, crossPos
}

// calculateSpaceAround 计算间距分布
func (f *FlexLayout) calculateSpaceAround(remaining, childCount, availW, availH int, isRow bool) int {
    if len(f.children) <= 1 {
        return 0
    }

    gapCount := len(f.children) - 1
    switch f.style.MainAxis {
    case SpaceBetween:
        return remaining / gapCount
    case SpaceAround:
        return remaining / len(f.children)
    case SpaceEvenly:
        return remaining / (len(f.children) + 1)
    default:
        return 0
    }
}

// calculateChildPosition 计算子节点位置
func (f *FlexLayout) calculateChildPosition(index int, mainPos, crossPos int, isRow bool, size Size) (int, int) {
    if isRow {
        return f.style.Padding.Left + mainPos, f.style.Padding.Top + crossPos
    }
    return f.style.Padding.Left + crossPos, f.style.Padding.Top + mainPos
}

// applyStretch 应用拉伸
func (f *FlexLayout) applyStretch(size *Size, availW, availH int, isRow bool) {
    if isRow {
        size.Height = availH
    } else {
        size.Width = availW
    }
}
```

#### 文件变更

```
runtime/layout/
├── flex.go               # 修改：优化 Measure 和 LayoutChildren
├── flex_cache.go         # 新增：Flex 分布缓存
└── types.go              # 修改：Engine 添加 FlexCache
```

---

### 3.4 HitMap 与事件处理

#### 目标
- 实现通用 HitMap 功能
- 支持将 LayoutBox 映射到交互区域
- 为应用层提供坐标转换能力
- 支持多层 Z-order

#### 实现计划

```go
// hitmap.go
package layout

import (
    "sort"
)

// HitMapEntry HitMap 条目
type HitMapEntry struct {
    NodeID  uint64    // 稳定节点标识（由应用层提供）
    ID      string    // 节点 ID
    Bounds  Rect      // 边界矩形
    ZOrder  int       // Z 轴顺序
}

// HitMap 命中映射
type HitMap struct {
    entries    []*HitMapEntry
    version    uint64
}

// NewHitMap 创建新的 HitMap
func NewHitMap() *HitMap {
    return &HitMap{
        entries: make([]*HitMapEntry, 0),
    }
}

// BuildFromLayoutBox 从 LayoutBox 构建 HitMap
func (hm *HitMap) BuildFromLayoutBox(root *LayoutBox) {
    entries := hm.collectEntries(root, 0, 0, 0)

    // 按 Z-order 排序
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].ZOrder < entries[j].ZOrder
    })

    hm.entries = entries
    hm.version++
}

// collectEntries 递归收集条目
func (hm *HitMap) collectEntries(box *LayoutBox, x, y, zOrder int) []*HitMapEntry {
    entries := make([]*HitMapEntry, 0)

    entry := &HitMapEntry{
        ID:     box.ID,
        Bounds: Rect{X: x + box.X, Y: y + box.Y, Width: box.Width, Height: box.Height},
        ZOrder: zOrder,
    }
    entries = append(entries, entry)

    for _, child := range box.Children {
        childEntries := hm.collectEntries(child, x, y, zOrder)
        entries = append(entries, childEntries...)
    }

    return entries
}

// HitTest 命中测试（返回最上层的匹配）
func (hm *HitMap) HitTest(x, y int) *HitMapEntry {
    // 从后向前查找（Z-order 最大的最上层）
    for i := len(hm.entries) - 1; i >= 0; i-- {
        entry := hm.entries[i]
        if entry.Bounds.Contains(x, y) {
            return entry
        }
    }
    return nil
}

// HitTestAll 命中测试（返回所有匹配，按 Z-order 排序）
func (hm *HitMap) HitTestAll(x, y int) []*HitMapEntry {
    results := make([]*HitMapEntry, 0)
    for _, entry := range hm.entries {
        if entry.Bounds.Contains(x, y) {
            results = append(results, entry)
        }
    }
    return results
}

// GetVersion 获取版本号
func (hm *HitMap) GetVersion() uint64 {
    return hm.version
}

// LocalCoordinateTransformer 坐标转换器接口
type LocalCoordinateTransformer interface {
    // ScreenToLocal 屏幕坐标转局部坐标
    ScreenToLocal(screenX, screenY int) (localX, localY int)

    // LocalToScreen 局部坐标转屏幕坐标
    LocalToScreen(localX, localY int) (screenX, screenY int)
}

// LocalTransformer 局部坐标转换器实现
type LocalTransformer struct {
    origin Point
}

// NewLocalTransformer 创建局部坐标转换器
func NewLocalTransformer(originX, originY int) *LocalTransformer {
    return &LocalTransformer{
        origin: Point{X: originX, Y: originY},
    }
}

// ScreenToLocal 屏幕坐标转局部坐标
func (lt *LocalTransformer) ScreenToLocal(screenX, screenY int) (int, int) {
    return screenX - lt.origin.X, screenY - lt.origin.Y
}

// LocalToScreen 局部坐标转屏幕坐标
func (lt *LocalTransformer) LocalToScreen(localX, localY int) (int, int) {
    return localX + lt.origin.X, localY + lt.origin.Y
}
```

**Engine 集成 HitMap**：

```go
// types.go
type LayoutResult struct {
    Boxes       []LayoutBox
    Root        *LayoutBox
    ContentSize Size
    Dirty       bool
    HitMap      *HitMap  // 新增
}

// engine.go
func (e *Engine) Layout(root Node, constraints Constraints) *LayoutResult {
    // ... 现有布局逻辑 ...

    result := &LayoutResult{
        Boxes: make([]LayoutBox, 0),
        Dirty: true,
        HitMap: NewHitMap(),  // 新增
    }

    box := e.layoutNode(root, constraints, 0, 0)
    result.Root = box
    result.Boxes = e.collectBoxes(box)

    // 构建 HitMap
    result.HitMap.BuildFromLayoutBox(box)

    return result
}
```

#### 文件变更

```
runtime/layout/
├── hitmap.go             # 新增：HitMap 和坐标转换
└── types.go              # 修改：LayoutResult 添加 HitMap 字段
```

---

### 3.5 边界验证

#### 目标
- 检测布局计算与组件声明的一致性
- 提供调试和诊断工具
- 支持 Bounds 感知接口（应用层实现）

#### 实现计划

```go
// validator.go
package layout

import (
    "fmt"
)

// BoundsProvider 提供边界声明（由应用层实现）
type BoundsProvider interface {
    // GetBounds 返回组件声明的边界
    GetBounds() [4]int
}

// BoundsValidator 边界验证器
type BoundsValidator struct {
    enabled bool
}

// NewBoundsValidator 创建边界验证器
func NewBoundsValidator() *BoundsValidator {
    return &BoundsValidator{
        enabled: true,
    }
}

// Enable 启用验证
func (v *BoundsValidator) Enable() {
    v.enabled = true
}

// Disable 禁用验证
func (v *BoundsValidator) Disable() {
    v.enabled = false
}

// ValidateBox 验证单个盒子的边界
func (v *BoundsValidator) ValidateBox(box *LayoutBox, node Node) error {
    if !v.enabled {
        return nil
    }

    if node == nil {
        return fmt.Errorf("node is nil for box %s", box.ID)
    }

    // 检查节点是否实现了 BoundsProvider
    if provider, ok := node.(BoundsProvider); ok {
        declaredBounds := provider.GetBounds()
        expectedBounds := [4]int{box.X, box.Y, box.Width, box.Height}

        if declaredBounds != expectedBounds {
            return fmt.Errorf(
                "bounds inconsistency for node %s: declared %v, computed %v",
                box.ID, declaredBounds, expectedBounds,
            )
        }
    }

    return nil
}

// ValidateTree 递归验证整个树
func (v *BoundsValidator) ValidateTree(root *LayoutBox, rootNode Node) []error {
    return v.validateRecursive(root, rootNode)
}

func (v *BoundsValidator) validateRecursive(box *LayoutBox, node Node) []error {
    errors := make([]error, 0)

    // 验证当前节点
    if err := v.ValidateBox(box, node); err != nil {
        errors = append(errors, err)
    }

    // 递归验证子节点
    children := node.Children()
    for i, childBox := range box.Children {
        if i < len(children) {
            childErrors := v.validateRecursive(childBox, children[i])
            errors = append(errors, childErrors...)
        }
    }

    return errors
}
```

**Engine 集成验证**：

```go
// types.go
type Engine struct {
    dirtyNodes map[string]bool
    stats      LayoutStats
    cache      *Cache
    validator  *BoundsValidator  // 新增
}

// NewEngine 创建引擎（可选启用验证）
func NewEngine(options ...EngineOption) *Engine {
    engine := &Engine{
        dirtyNodes: make(map[string]bool),
        stats:      LayoutStats{},
        cache: &Cache{
            entries: make(map[string]*CachedLayout),
            maxSize: 1000,
        },
        validator: NewBoundsValidator(),
    }

    for _, opt := range options {
        opt(engine)
    }

    return engine
}

// EngineOption 引擎选项
type EngineOption func(*Engine)

// WithValidation 启用边界验证
func WithValidation() EngineOption {
    return func(e *Engine) {
        if e.validator != nil {
            e.validator.Enable()
        }
    }
}

// WithoutValidation 禁用边界验证
func WithoutValidation() EngineOption {
    return func(e *Engine) {
        if e.validator != nil {
            e.validator.Disable()
        }
    }
}

// Layout 带验证的布局
func (e *Engine) Layout(root Node, constraints Constraints) *LayoutResult {
    if root == nil {
        return &LayoutResult{}
    }

    // ... 现有布局逻辑 ...

    result := &LayoutResult{
        Boxes: make([]LayoutBox, 0),
        Dirty: true,
    }

    box := e.layoutNode(root, constraints, 0, 0)
    result.Root = box
    result.Boxes = e.collectBoxes(box)

    // 边界验证（如果启用）
    if e.validator != nil && e.validator.enabled {
        if errors := e.validator.ValidateTree(box, root); len(errors) > 0 {
            // 可以选择记录日志、panic 或返回错误
            fmt.Printf("Layout validation errors: %v\n", errors)
        }
    }

    return result
}
```

#### 文件变更

```
runtime/layout/
├── validator.go          # 新增：边界验证器
└── types.go              # 修改：Engine 添加验证器
```

---

### 3.6 接口扩展

#### 目标
- 提供稳定标识符接口（用于缓存失效）
- 支持版本跟踪（更智能的缓存）
- 添加布局信息接口（Flexbox 配置）

#### 实现计划

```go
// types.go 添加新接口

// Identifiable 可标识节点
type Identifiable interface {
    Node
    // GetStableID 返回稳定的节点标识（跨渲染周期保持不变）
    GetStableID() uint64
}

// Versioned 版本化节点
type Versioned interface {
    Node
    // GetVersion 返回节点版本号（内容变化时递增）
    GetVersion() uint64
}

// LayoutInfoProvider 布局信息提供者
type LayoutInfoProvider interface {
    Node
    // GetLayoutInfo 返回布局配置
    GetLayoutInfo() *LayoutInfo
}

// LayoutInfo 布局信息
type LayoutInfo struct {
    // Flexbox 配置
    Direction     FlexDirection
    MainAxis      MainAxisAlignment
    CrossAxis     CrossAxisAlignment
    Gap           int
    CrossGap      int
    Padding       Padding

    // 弹性配置映射
    FlexibleChildren map[int]*Flex

    // 填充行为
    FillWidth  bool
    FillHeight bool
}
```

**适配器模式示例（应用层实现）**：

```go
// 示例：应用层适配 VNode 到 layout.Node

// VNodeAdapter VNode 适配器（示例）
type VNodeAdapter struct {
    vnode     interface{}  // rtui.VNode
    fiber     interface{}  // *rtui.Fiber
    layoutInfo *layout.LayoutInfo
}

func (a *VNodeAdapter) ID() string {
    // 从 VNode 获取
    return ""
}

func (a *VNodeAdapter) Type() string {
    return ""
}

func (a *VNodeAdapter) Children() []layout.Node {
    // 递归适配子节点
    return nil
}

func (a *VNodeAdapter) GetStableID() uint64 {
    // 从 Fiber.NodeID 获取
    return 0
}

func (a *VNodeAdapter) GetVersion() uint64 {
    // 从 Fiber 版本获取
    return 0
}

func (a *VNodeAdapter) GetLayoutInfo() *layout.LayoutInfo {
    return a.layoutInfo
}

// ... 其他接口实现
```

#### 文件变更

```
runtime/layout/
└── types.go              # 修改：添加新接口
```

---

## 4. 实施计划

### 4.1 阶段划分

| 阶段 | 任务 | 优先级 | 预计时间 |
|-----|------|-------|---------|
| **阶段 1：缓存优化** |
| - 修复 constraintsKey bug | P0 | 0.5 天 |
| - 实现 FlexCache | P0 | 1 天 |
| - 实现叶子节点优化缓存 | P1 | 1 天 |
| - 添加按节点 ID 失效 | P1 | 0.5 天 |
| **阶段 2：Dirty Tracking** |
| - 完善 DirtyTracker | P0 | 1 天 |
| - 实现增量布局 | P0 | 2 天 |
| - 脏标记传播优化 | P1 | 1 天 |
| **阶段 3：测量优化** |
| - 优化 FlexLayout.Measure | P1 | 1 天 |
| - 重构 LayoutChildren | P1 | 1.5 天 |
| - 优化 childConstraints | P2 | 0.5 天 |
| **阶段 4：高级功能** |
| - 实现 HitMap | P1 | 1.5 天 |
| - 实现边界验证器 | P2 | 1 天 |
| - 扩展接口定义 | P2 | 1 天 |
| **阶段 5：测试与文档** |
| - 单元测试补充 | P0 | 2 天 |
| - 集成测试编写 | P1 | 1.5 天 |
| - 更新使用文档 | P2 | 1 天 |
| - 性能基准测试 | P1 | 1 天 |

**总计**：约 17 天（3-4 周）

### 4.2 任务依赖关系

```
修复缓存 bug ────────┐
                    ├──→ 实现 FlexCache ──→ FlexLayout 优化 ──→ 测试
实现叶子节点缓存 ────┤
                    │
完善 DirtyTracker ───┴──→ 增量布局 ──────────────────────→ 集成测试
                    │
HitMap 实现 ────────┼──→ 边界验证 ──────────────────────────→ 文档更新
                    │
接口扩展 ────────────┘
```

### 4.3 测试策略

**单元测试**：
- FlexCache 缓存失效逻辑
- DirtyTracker 脏标记传播
- HitMap 命中测试
- Validator 边界检测

**集成测试**：
- 带缓存的完整布局流程
- 增量布局场景
- HitMap 与布局结果对应

**性能测试**：
- 缓存命中率对比
- 增量布局 vs 完整布局
- Flex 分布缓存效果

---

## 5. 风险评估与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|-----|------|------|---------|
| 缓存失效不正确 | 高 | 中 | 详细测试、逐步替换 |
| 增量布局状态不一致 | 高 | 中 | 版本号跟踪、全量回退 |
| 性能回归 | 中 | 低 | 性能基准对比 |
| API 变更影响现有代码 | 中 | 低 | 向后兼容、废弃警告 |

---

## 6. 成功指标

### 6.1 功能指标
- ✅ 缓存约束键正确生成
- ✅ Flex 分布缓存生效
- ✅ Dirty Tracking 支持增量布局
- ✅ HitMap 正确映射布局结果
- ✅ 边界验证检测不一致

### 6.2 性能指标
- 缓存命中率 > 50%
- 增量布局性能提升 > 60%
- Flex 测量时间减少 > 30%

### 6.3 质量指标
- 单元测试覆盖率 > 80%
- 集成场景覆盖率 > 70%
- 无严重 bug

---

## 7. 后续优化方向

### 7.1 短期（3-6 个月）
- 表格布局算法
- 绝对定位支持
- 文本对齐增强

### 7.2 中期（6-12 个月）
- Grid 布局
- 动画布局插值
- 多线程布局计算

### 7.3 长期（12+ 个月）
- 自适应布局优化
- AI 辅助布局预测

---

## 8. 总结

本方案基于 `runtime/layout/` 的现有 V3 实现，吸收 `runtime/compute/` 的通用布局优化策略，在保持独立性的前提下，从缓存、Dirty Tracking、测量算法、HitMap、边界验证等方面进行全面增强。

**核心原则**：
1. **保持独立性** - 不依赖 VNode/Fiber，提供通用接口供应用层适配
2. **渐进式增强** - 基于现有代码优化，避免大规模重写
3. **性能优先** - 复制 compute/ 的成功经验，重点优化缓存和增量布局
4. **可测试性** - 完善测试覆盖，确保质量

**预期收益**：
- 缓存优化减少重复计算 > 40%
- Dirty Tracking 支持高效增量更新
- HitMap 支持事件系统集成
- 边界验证提升调试能力

---

**文档版本**: v2.0
**更新日期**: 2026-02-15
**作者**: AI Assistant (Crush)
**状态**: 待审核
