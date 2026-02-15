# Runtime Layout 系统优化实施方案

## 执行摘要

本方案旨在将 `runtime/compute` 的高级功能移植到 `runtime/layout`，创建一个**独立于 VNode/Fiber** 的通用布局引擎，同时保留现有 `runtime/layout` 的完整 Flexbox 特性。

**核心目标**：
1. 创建抽象接口层，解耦对 VNode/Fiber 的依赖
2. 移植 `compute/` 的性能优化（单次测量、缓存、Dirty Tracking）
3. 增强 `layout/` 的 Flexbox 实现，保持其标准性
4. 保持向后兼容，不影响现有使用

---

## 1. 调研总结

### 1.1 当前架构分析

#### runtime/compute/engine.go - 生产系统
```
优势：
✅ 单次测量优化 (LayoutMeasurer 接口)
✅ 智能缓存策略（叶子节点缓存）
✅ Dirty Tracking（增量布局）
✅ Fiber 集成（NodeID、Layer、DiffKey 匹配）
✅ HitMap 构建（事件处理）
✅ Bounds Validator（一致性检查）

不足：
❌ 紧密依赖 VNode/Fiber 架构
❌ 缺少完整的 Flexbox 支持（无 flex-shrink、flex-basis、反向排列）
❌ 代码复杂度高（2000+ 行）
```

#### runtime/layout/ - 通用库
```
优势：
✅ 标准的 Flexbox 实现
✅ 支持 flex-grow、flex-shrink、flex-basis
✅ 支持反向排列（RowReverse、ColumnReverse）
✅ 支持交叉轴间距（CrossGap）
✅ 零依赖（纯 Go，可独立使用）
✅ 清晰的接口设计

不足：
❌ 缺少单次测量优化
❌ 缓存策略简单（完整树哈希）
❌ 无 Dirty Tracking
❌ 无增量更新支持
❌ 未与实际系统集成
```

### 1.2 关键功能对比矩阵

| 功能类别 | compute/ | layout/ | 优化后 |
|---------|----------|---------|---------|
| **核心算法** |
| Flexbox 布局 | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Flex-grow | ✅ | ✅ | ✅ |
| Flex-shrink | ❌ | ✅ | ✅ |
| Flex-basis | ❌ | ✅ | ✅ |
| 反向排列 | ❌ | ✅ | ✅ |
| 主轴对齐 | ✅ | ✅ | ✅ |
| 交叉轴对齐 | ✅ | ✅ | ✅ |
| **性能优化** |
| 单次测量 | ✅ | ❌ | ✅ |
| 叶子缓存 | ✅ | ❌ | ✅ |
| Dirty Tracking | ✅ | ❌ | ✅ |
| 增量布局 | ✅ | ❌ | ✅ |
| Flex 分布缓存 | ✅ | ❌ | ✅ |
| **集成特性** |
| 抽象接口 | ❌ (VNode) | ✅ | ✅ |
| 稳定标识 | ✅ (NodeID) | ❌ | ✅ |
| 事件处理 | ✅ (HitMap) | ❌ | ✅ |
| 边界验证 | ✅ | ❌ | ✅ |
| **特殊布局** |
| 文本对齐 | ✅ | ❌ | ✅ |
| 边框处理 | ✅ | ❌ | ✅ |
| 表格布局 | ✅ | ❌ | ✅ |

---

## 2. 抽象接口设计

### 2.1 核心抽象层

为解耦对 VNode/Fiber 的依赖，设计以下抽象接口：

```go
package layout

// =============================================================================
// Core Abstraction Layer
// =============================================================================

// LayoutNode represents any node that can be laid out
// This replaces the VNode dependency from compute/
type LayoutNode interface {
    // Identity
    ID() string
    Type() NodeType

    // Tree structure
    Children() []LayoutNode
    Parent() LayoutNode

    // Layout info
    GetLayoutInfo() LayoutInfo
    GetProps() Props
    GetStyle() Style

    // Optional: measurement
    Measure(constraints Constraints) Size
}

// LayoutableNode extends LayoutNode with layout capabilities
type LayoutableNode interface {
    LayoutNode

    // Layout children
    LayoutChildren(constraints Constraints, width, height int) []LayoutBox
}

// MeasurableNode extends LayoutNode with measurement
type MeasurableNode interface {
    LayoutNode
    Measurable
}

// NodeIDProvider provides stable identity (like Fiber.NodeID)
type NodeIDProvider interface {
    LayoutNode
    NodeID() uint64
}

// LayeredNode provides rendering layer information
type LayeredNode interface {
    LayoutNode
    Layer() Layer
}

// ========================================================================
// Layout Information (extracted from compute/LayoutInfo)
// ========================================================================

// LayoutInfo contains layout configuration for a node
// This abstracts rtui.LayoutInfo from compute/
type LayoutInfo struct {
    // Flexbox properties
    Direction     FlexDirection
    MainAxis      MainAxisAlignment
    CrossAxis     CrossAxisAlignment
    Flex          int
    StretchCross  bool

    // Spacing
    Gap       int
    CrossGap  int
    Padding   [4]int // top, right, bottom, left

    // Fill behavior
    FillWidth  bool
    FillHeight bool
}

// Props represents node properties
type Props interface {
    Get(key string) interface{}
    GetInt(key string, defaultVal int) int
    GetString(key string, defaultVal string) string
    GetBool(key string, defaultVal bool) bool
    Has(key string) bool
}

// Style represents visual style
type Style struct {
    FG    Color
    BG    Color
    Bold  bool
    Italic bool
}

// =============================================================================
// Layout Results (independent of compute/)
// =============================================================================

// LayoutBox represents a laid out box
type LayoutBox struct {
    // Identity
    ID      string
    NodeID  uint64  // Stable identifier
    Layer   Layer   // Rendering layer

    // Position and size
    Box    Rect
    Bounds Rect

    // Layout state
    LayoutDirty bool
    NaturalSize Size  // For alignment calculations

    // Tree structure
    Parent   *LayoutBox
    Children []*LayoutBox

    // Content
    Content    string  // Rendered text (with padding)
    Props      Props
    Style      Style
}

// LayoutResult represents complete layout output
type LayoutResult struct {
    Root      *LayoutBox
    Boxes     []LayoutBox
    HitMap    *HitMap     // For event handling
    Dirty     bool
    Version   uint64      // For cache invalidation
}

// =============================================================================
// Hit Map (event handling)
// =============================================================================

// HitMap maps screen positions to nodes
type HitMap struct {
    entries map[string]*HitMapEntry
    version uint64
}

// HitMapEntry represents a hit-able region
type HitMapEntry struct {
    NodeID   uint64
    Bounds   Rect
    LocalXY  func(screenX, screenY int) (int, int)
    ZOrder   int
}

// =============================================================================
// Constraints (enhanced from layout/)
// =============================================================================

// Constraints defines layout space constraints
type Constraints struct {
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
}

// Enhanced methods
func (c Constraints) IsTight() bool
func (c Constraints) IsBounded() bool
func (c Constraints) HasBoundedWidth() bool
func (c Constraints) HasBoundedHeight() bool
func (c Constraints) Constrain(width, height int) (int, int)
func (c Constraints) SubtractPadding(h, v int) Constraints
func (c Constraints) WithWidth(min, max int) Constraints
func (c Constraints) WithHeight(min, max int) Constraints
```

### 2.2 适配器设计

为了兼容现有 VNode/Fiber，提供适配器：

```go
// =============================================================================
// Adapters for VNode/Fiber compatibility
// =============================================================================

// VNodeAdapter wraps rtui.VNode to implement LayoutNode
type VNodeAdapter struct {
    vnode rtui.VNode
    fiber *rtui.Fiber
}

func (a *VNodeAdapter) ID() string {
    if a.vnode != nil {
        return a.vnode.Key()
    }
    return ""
}

func (a *VNodeAdapter) Type() NodeType {
    // Convert VNodeType to layout.NodeType
    return NodeType(a.vnode.Type().String())
}

func (a *VNodeAdapter) Children() []LayoutNode {
    children := a.vnode.Children()
    result := make([]LayoutNode, len(children))
    for i, child := range children {
        result[i] = &VNodeAdapter{vnode: child}
    }
    return result
}

func (a *VNodeAdapter) GetLayoutInfo() LayoutInfo {
    // Extract from VNode props or rtui.GetLayoutInfo
    return LayoutInfo{
        Direction:    a.directionFromVNode(),
        MainAxis:     a.alignFromVNode(),
        CrossAxis:    a.crossAlignFromVNode(),
        Flex:         a.flexFromVNode(),
        Gap:          a.gapFromVNode(),
        Padding:      a.paddingFromVNode(),
        StretchCross: a.stretchFromVNode(),
    }
}

func (a *VNodeAdapter) NodeID() uint64 {
    if a.fiber != nil {
        return a.fiber.NodeID
    }
    return 0
}

func (a *VNodeAdapter) Layer() Layer {
    if a.fiber != nil {
        return a.fiber.Layer
    }
    return LayerBase
}

// ... other methods

// =============================================================================
// FiberAdapter wraps rtui.Fiber to provide NodeID/Layer
// =============================================================================

type FiberAdapter struct {
    fiber *rtui.Fiber
}

func (a *FiberAdapter) NodeID() uint64 {
    return a.fiber.NodeID
}

func (a *FiberAdapter) Layer() Layer {
    return a.fiber.Layer
}
```

---

## 3. 功能增强方案

### 3.1 单次测量优化（Single-Pass Layout）

移植 `compute/` 的 `LayoutMeasurer` 接口：

```go
// =============================================================================
// Single-Pass Layout Measurement
// =============================================================================

// LayoutMeasurer enables single-pass layout measurement
// This avoids O(N²) re-measurement by returning child constraints
type LayoutMeasurer interface {
    // IsLayoutMeasurer is a marker method
    IsLayoutMeasurer()

    // MeasureLayout performs single-pass measurement
    // Returns:
    //   - Size: total size of this node
    //   - ChildConstraints: constraints for each child (in order)
    //   - ChildSizes: measured size of each child
    MeasureLayout(
        measurer ChildMeasurer,
        constraints Constraints,
    ) LayoutMeasurement
}

// LayoutMeasurement contains measurement results
type LayoutMeasurement struct {
    Size             Size
    ChildConstraints []Constraints
    ChildSizes       []Size
}

// ChildMeasurer measures a child with constraints
type ChildMeasurer interface {
    MeasureChild(child interface{}, constraints Constraints) Size
}

// =============================================================================
// Enhanced FlexLayout with Single-Pass
// =============================================================================

// MeasureLayout implements LayoutMeasurer for FlexLayout
func (f *FlexLayout) IsLayoutMeasurer() {}

func (f *FlexLayout) MeasureLayout(
    measurer ChildMeasurer,
    constraints Constraints,
) LayoutMeasurement {
    children := f.children
    if len(children) == 0 {
        return LayoutMeasurement{
            Size: Size{
                Width:  constraints.MinWidth,
                Height: constraints.MinHeight,
            },
        }
    }

    // Calculate padding
    paddingWidth := f.style.Padding.Left + f.style.Padding.Right
    paddingHeight := f.style.Padding.Top + f.style.Padding.Bottom

    // Determine main/cross axis
    isRow := f.style.Direction == FlexRow ||
              f.style.Direction == FlexRowReverse

    // Phase 1: Identify flex vs fixed children
    var flexChildren []int
    var fixedSize int
    flexTotalFactor := 0

    childConstraints := make([]Constraints, len(children))
    childSizes := make([]Size, len(children))

    for i, child := range children {
        childInfo := f.getChildLayoutInfo(child)

        if childInfo.Flex > 0 {
            flexChildren = append(flexChildren, i)
            flexTotalFactor += childInfo.Flex
        } else {
            // Fixed child: measure with natural size
            childConstraints[i] = f.getChildConstraints(
                constraints, childInfo, isRow)
            childSizes[i] = measurer.MeasureChild(
                child, childConstraints[i])
            if isRow {
                fixedSize += childSizes[i].Width
            } else {
                fixedSize += childSizes[i].Height
            }
        }
    }

    // Phase 2: Distribute space to flex children
    if len(flexChildren) > 0 && isRow && constraints.HasBoundedWidth() {
        available := constraints.MaxWidth - paddingWidth
        if isRow {
            available -= (len(children) - 1) * f.style.Gap
        }
        remaining := available - fixedSize

        for _, idx := range flexChildren {
            flexSize := (remaining * f.style.FlexibleChildren[idx].Grow) /
                       flexTotalFactor
            if flexSize < 0 {
                flexSize = 0
            }
            childConstraints[idx] = Constraints{
                MinWidth: flexSize,
                MaxWidth: flexSize,
                MinHeight: 0,
                MaxHeight: constraints.MaxHeight - paddingHeight,
            }
            childSizes[idx] = measurer.MeasureChild(
                f.children[idx], childConstraints[idx])
        }
    } else if len(flexChildren) > 0 && !isRow &&
              constraints.HasBoundedHeight() {
        // Similar logic for VStack
        available := constraints.MaxHeight - paddingHeight
        if !isRow {
            available -= (len(children) - 1) * f.style.Gap
        }
        remaining := available - fixedSize

        for _, idx := range flexChildren {
            flexSize := (remaining * f.style.FlexibleChildren[idx].Grow) /
                       flexTotalFactor
            if flexSize < 0 {
                flexSize = 0
            }
            childConstraints[idx] = Constraints{
                MinWidth: 0,
                MaxWidth: constraints.MaxWidth - paddingWidth,
                MinHeight: flexSize,
                MaxHeight: flexSize,
            }
            childSizes[idx] = measurer.MeasureChild(
                f.children[idx], childConstraints[idx])
        }
    }

    // Phase 3: Calculate total size
    var totalMainSize, maxCrossSize int

    if isRow {
        totalMainSize = fixedSize
        maxCrossSize = 0
        for _, size := range childSizes {
            if size.Height > maxCrossSize {
                maxCrossSize = size.Height
            }
        }
        totalMainSize += (len(children) - 1) * f.style.Gap
    } else {
        totalMainSize = fixedSize
        maxCrossSize = 0
        for _, size := range childSizes {
            if size.Width > maxCrossSize {
                maxCrossSize = size.Width
            }
        }
        totalMainSize += (len(children) - 1) * f.style.Gap
    }

    totalWidth := totalMainSize
    totalHeight := maxCrossSize

    if isRow {
        totalWidth += paddingWidth
        totalHeight += paddingHeight
    } else {
        totalWidth += paddingWidth
        totalHeight += paddingHeight
    }

    // Cross-axis filling
    if isRow && constraints.HasBoundedHeight() &&
       totalHeight < constraints.MaxHeight {
        totalHeight = constraints.MaxHeight
    } else if !isRow && constraints.HasBoundedWidth() &&
              totalWidth < constraints.MaxWidth {
        totalWidth = constraints.MaxWidth
    }

    return LayoutMeasurement{
        Size:             Size{Width: totalWidth, Height: totalHeight},
        ChildConstraints: childConstraints,
        ChildSizes:       childSizes,
    }
}
```

### 3.2 增强缓存策略

移植 `compute/` 的智能缓存：

```go
// =============================================================================
// Enhanced Layout Cache
// =============================================================================

// CacheStrategy defines cache strategy
type CacheStrategy int

const (
    CacheLeafOnly CacheStrategy = iota  // Only cache leaf nodes
    CacheFullTree                  // Cache entire tree
    CacheHybrid                    // Hybrid approach
)

// LayoutCache provides intelligent caching
type LayoutCache struct {
    mu           sync.RWMutex
    entries      map[CacheKey]*CacheEntry
    strategy     CacheStrategy
    maxSize      int
    evictPolicy  EvictionPolicy

    // Statistics
    hits     int64
    misses   int64
    evictions int64
}

// CacheKey combines multiple factors for precise caching
type CacheKey struct {
    NodeID      uint64
    NodeType    string
    Constraints Constraints
    PropsHash   uint64
    ContentHash uint64
    Version     uint64  // For cache invalidation
}

// CacheEntry represents a cached layout result
type CacheEntry struct {
    Box         LayoutBox
    Size        Size
    Timestamp   time.Time
    HitCount    int
    IsLeaf      bool
    Valid       bool
}

// EvictionPolicy defines how to evict entries
type EvictionPolicy int

const (
    EvictLRU EvictionPolicy = iota
    EvictLFU
    EvictFIFO
)

// NewLayoutCache creates a new cache
func NewLayoutCache(strategy CacheStrategy) *LayoutCache {
    return &LayoutCache{
        entries:     make(map[CacheKey]*CacheEntry),
        strategy:    strategy,
        maxSize:     1000,
        evictPolicy: EvictLRU,
    }
}

// Get retrieves a cached entry
func (c *LayoutCache) Get(key CacheKey) (*CacheEntry, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, ok := c.entries[key]
    if ok {
        // Check validity
        if !entry.Valid {
            c.misses++
            return nil, false
        }
        entry.HitCount++
        c.hits++
        return entry, true
    }
    c.misses++
    return nil, false
}

// Set stores a cache entry
func (c *LayoutCache) Set(key CacheKey, entry *CacheEntry) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Check if we need to evict
    if len(c.entries) >= c.maxSize {
        c.evict()
    }

    c.entries[key] = entry
}

// evict removes entries based on policy
func (c *LayoutCache) evict() {
    switch c.evictPolicy {
    case EvictLRU:
        c.evictLRU()
    case EvictLFU:
        c.evictLFU()
    case EvictFIFO:
        c.evictFIFO()
    }
}

func (c *LayoutCache) evictLRU() {
    var oldestKey *CacheKey
    var oldestTime time.Time

    for key, entry := range c.entries {
        if oldestKey == nil || entry.Timestamp.Before(oldestTime) {
            oldestKey = &key
            oldestTime = entry.Timestamp
        }
    }

    if oldestKey != nil {
        delete(c.entries, *oldestKey)
        c.evictions++
    }
}

// Invalidate marks entries as invalid
func (c *LayoutCache) Invalidate(predicate func(CacheKey) bool) {
    c.mu.Lock()
    defer c.mu.Unlock()

    for key, entry := range c.entries {
        if predicate(key) {
            entry.Valid = false
        }
    }
}

// InvalidateByNodeID invalidates entries for a node
func (c *LayoutCache) InvalidateByNodeID(nodeID uint64) {
    c.Invalidate(func(key CacheKey) bool {
        return key.NodeID == nodeID
    })
}

// Stats returns cache statistics
func (c *LayoutCache) Stats() CacheStats {
    c.mu.RLock()
    defer c.mu.RUnlock()

    total := c.hits + c.misses
    hitRate := 0.0
    if total > 0 {
        hitRate = float64(c.hits) / float64(total)
    }

    return CacheStats{
        Size:      len(c.entries),
        Hits:      c.hits,
        Misses:    c.misses,
        Evictions: c.evictions,
        HitRate:   hitRate,
    }
}
```

### 3.3 Dirty Tracking

移植 `compute/` 的 Dirty Tracking：

```go
// =============================================================================
// Dirty Tracking
// =============================================================================

// DirtyTracker tracks which nodes need to be re-laid out
type DirtyTracker struct {
    mu     sync.RWMutex
    dirty   map[uint64]bool  // NodeID -> dirty flag
    version uint64           // Current version for invalidation
}

// NewDirtyTracker creates a new dirty tracker
func NewDirtyTracker() *DirtyTracker {
    return &DirtyTracker{
        dirty:   make(map[uint64]bool),
        version: 1,
    }
}

// MarkDirty marks a node as needing layout
func (t *DirtyTracker) MarkDirty(nodeID uint64) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.dirty[nodeID] = true
}

// MarkSubtreeDirty marks a subtree as needing layout
func (t *DirtyTracker) MarkSubtreeDirty(box *LayoutBox) {
    if box == nil {
        return
    }
    box.LayoutDirty = true
    t.markDescendantsDirty(box)
}

func (t *DirtyTracker) markDescendantsDirty(box *LayoutBox) {
    box.LayoutDirty = true
    if box.NodeID != 0 {
        t.MarkDirty(box.NodeID)
    }
    for _, child := range box.Children {
        t.markDescendantsDirty(child)
    }
}

// IsDirty checks if a node needs layout
func (t *DirtyTracker) IsDirty(nodeID uint64) bool {
    t.mu.RLock()
    defer t.mu.RUnlock()
    return t.dirty[nodeID]
}

// Clear clears all dirty flags
func (t *DirtyTracker) Clear() {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.dirty = make(map[uint64]bool)
    t.version++
}

// BumpVersion increments version for cache invalidation
func (t *DirtyTracker) BumpVersion() uint64 {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.version++
    return t.version
}

// Version returns current version
func (t *DirtyTracker) Version() uint64 {
    t.mu.RLock()
    defer t.mu.RUnlock()
    return t.version
}
```

### 3.4 Flexbox 增强

保留 `layout/` 的完整 Flexbox 支持并优化：

```go
// =============================================================================
// Enhanced FlexLayout with shrink/basis/reverse
// =============================================================================

// FlexLayout implements standard Flexbox with all features
type FlexLayout struct {
    id       string
    children []LayoutNode
    style    *FlexStyle
    size     Size
    position Point
}

// FlexStyle contains complete Flexbox configuration
type FlexStyle struct {
    // Direction
    Direction FlexDirection

    // Alignment
    MainAxis   MainAxisAlignment
    CrossAxis  CrossAxisAlignment

    // Spacing
    Gap       int
    CrossGap  int
    Padding   [4]int

    // Flex properties
    FlexibleChildren map[int]*Flex
    StretchCross    bool

    // Fill behavior
    FillWidth  bool  // Children fill width
    FillHeight bool  // Children fill height
}

// Flex contains complete flex configuration
type Flex struct {
    Grow   int  // Grow factor (default 0)
    Shrink int  // Shrink factor (default 1)
    Basis  int  // Basis size (default 0 = auto)
}

// FlexDirection with reverse support
type FlexDirection int

const (
    FlexRow FlexDirection = iota
    FlexColumn
    FlexRowReverse
    FlexColumnReverse
)

// Measure with shrink/basis support
func (f *FlexLayout) Measure(constraints Constraints) Size {
    if len(f.children) == 0 {
        return Size{
            Width:  constraints.MinWidth + f.style.Padding.Left + f.style.Padding.Right,
            Height: constraints.MinHeight + f.style.Padding.Top + f.style.Padding.Bottom,
        }
    }

    isRow := f.style.Direction == FlexRow ||
              f.style.Direction == FlexRowReverse
    isReverse := f.style.Direction == FlexRowReverse ||
                 f.style.Direction == FlexColumnReverse

    // Phase 1: Measure all children
    childSizes := make([]Size, len(f.children))
    var flexChildren []struct {
        index  int
        config *Flex
    }

    totalMainSize := 0
    maxCrossSize := 0
    totalGrow := 0
    totalBasis := 0

    for i, child := range f.children {
        // Measure child with initial constraints
        childConstraints := f.getChildConstraints(constraints, child, isRow)
        childSizes[i] = f.measureChild(child, childConstraints)

        // Check flex configuration
        if config, ok := f.style.FlexibleChildren[i]; ok {
            flexChildren = append(flexChildren, struct {
                index  int
                config *Flex
            }{i, config})

            totalGrow += config.Grow
            if config.Basis > 0 {
                totalBasis += config.Basis
            } else {
                // Auto basis = natural size
                if isRow {
                    totalBasis += childSizes[i].Width
                } else {
                    totalBasis += childSizes[i].Height
                }
            }
        } else {
            // Fixed child
            if isRow {
                totalMainSize += childSizes[i].Width
                if childSizes[i].Height > maxCrossSize {
                    maxCrossSize = childSizes[i].Height
                }
            } else {
                totalMainSize += childSizes[i].Height
                if childSizes[i].Width > maxCrossSize {
                    maxCrossSize = childSizes[i].Width
                }
            }
        }

        // Add gap (except after last)
        if i < len(f.children)-1 {
            if isRow {
                totalMainSize += f.style.Gap
            } else {
                totalMainSize += f.style.Gap
            }
        }
    }

    // Phase 2: Handle shrinking if needed
    if isRow && constraints.HasBoundedWidth() {
        available := constraints.MaxWidth - f.style.Padding.Left - f.style.Padding.Right
        if totalMainSize > available {
            // Need to shrink flex children
            f.shrinkChildren(childSizes, flexChildren,
                available, isRow, totalGrow, totalBasis)
        }
    } else if !isRow && constraints.HasBoundedHeight() {
        available := constraints.MaxHeight - f.style.Padding.Top - f.style.Padding.Bottom
        if totalMainSize > available {
            f.shrinkChildren(childSizes, flexChildren,
                available, isRow, totalGrow, totalBasis)
        }
    }

    // Phase 3: Calculate final size
    var width, height int
    if isRow {
        width = totalMainSize + f.style.Padding.Left + f.style.Padding.Right
        height = maxCrossSize + f.style.Padding.Top + f.style.Padding.Bottom
    } else {
        width = maxCrossSize + f.style.Padding.Left + f.style.Padding.Right
        height = totalMainSize + f.style.Padding.Top + f.style.Padding.Bottom
    }

    // Apply constraints
    width = max(constraints.MinWidth, width)
    height = max(constraints.MinHeight, height)

    if constraints.HasBoundedWidth() && width > constraints.MaxWidth {
        width = constraints.MaxWidth
    }
    if constraints.HasBoundedHeight() && height > constraints.MaxHeight {
        height = constraints.MaxHeight
    }

    return Size{Width: width, Height: height}
}

// shrinkChildren shrinks flex children to fit available space
func (f *FlexLayout) shrinkChildren(
    childSizes []Size,
    flexChildren []struct { index int; config *Flex },
    available int,
    isRow bool,
    totalGrow int,
    totalBasis int,
) {
    // Calculate shrink factor
    totalShrink := 0
    for _, fc := range flexChildren {
        totalShrink += fc.config.Shrink
    }

    if totalShrink == 0 {
        // No shrinking possible
        return
    }

    // Distribute shrinkage
    overflow := 0
    if isRow {
        for _, fc := range flexChildren {
            if isRow {
                overflow += childSizes[fc.index].Width
            } else {
                overflow += childSizes[fc.index].Height
            }
        }
        overflow -= available
    } else {
        for _, fc := range flexChildren {
            if isRow {
                overflow += childSizes[fc.index].Width
            } else {
                overflow += childSizes[fc.index].Height
            }
        }
        overflow -= available
    }

    for _, fc := range flexChildren {
        shrinkAmount := (overflow * fc.config.Shrink) / totalShrink
        if isRow {
            childSizes[fc.index].Width -= shrinkAmount
            if childSizes[fc.index].Width < 0 {
                childSizes[fc.index].Width = 0
            }
        } else {
            childSizes[fc.index].Height -= shrinkAmount
            if childSizes[fc.index].Height < 0 {
                childSizes[fc.index].Height = 0
            }
        }
    }
}

// LayoutChildren with reverse support
func (f *FlexLayout) LayoutChildren(width, height int) []LayoutBox {
    if len(f.children) == 0 {
        return nil
    }

    isRow := f.style.Direction == FlexRow ||
              f.style.Direction == FlexRowReverse
    isReverse := f.style.Direction == FlexRowReverse ||
                 f.style.Direction == FlexColumnReverse

    // Calculate available space
    availableWidth := width - f.style.Padding.Left - f.style.Padding.Right
    availableHeight := height - f.style.Padding.Top - f.style.Padding.Bottom

    // Measure children
    childSizes := make([]Size, len(f.children))
    for i, child := range f.children {
        constraints := Constraints{
            MinWidth:  0,
            MaxWidth:  availableWidth,
            MinHeight: 0,
            MaxHeight: availableHeight,
        }
        childSizes[i] = f.measureChild(child, constraints)
    }

    // Calculate flex distribution
    fixedSize := 0
    totalFlex := 0
    for i, child := range f.children {
        if config, ok := f.style.FlexibleChildren[i]; ok {
            totalFlex += config.Grow
        } else {
            if isRow {
                fixedSize += childSizes[i].Width
            } else {
                fixedSize += childSizes[i].Height
            }
        }
    }

    // Distribute remaining space
    remainingSpace := 0
    if isRow {
        remainingSpace = availableWidth - fixedSize - (len(f.children)-1)*f.style.Gap
    } else {
        remainingSpace = availableHeight - fixedSize - (len(f.children)-1)*f.style.Gap
    }

    finalSizes := make([]Size, len(f.children))
    for i := range f.children {
        if config, ok := f.style.FlexibleChildren[i]; ok && config.Grow > 0 {
            extra := (remainingSpace * config.Grow) / totalFlex
            if isRow {
                finalSizes[i] = Size{
                    Width:  childSizes[i].Width + extra,
                    Height: childSizes[i].Height,
                }
            } else {
                finalSizes[i] = Size{
                    Width:  childSizes[i].Width,
                    Height: childSizes[i].Height + extra,
                }
            }
        } else {
            finalSizes[i] = childSizes[i]
        }
    }

    // Calculate positions with reverse support
    boxes := make([]LayoutBox, len(f.children))
    mainPos := 0

    // Handle reverse direction
    if isReverse {
        if isRow {
            mainPos = availableWidth
        } else {
            mainPos = availableHeight
        }
    }

    for i := range f.children {
        var x, y int

        // Calculate cross-axis position
        if isRow {
            y = f.style.Padding.Top
            if f.style.CrossAxis == CrossCenter {
                y = f.style.Padding.Top + (availableHeight-finalSizes[i].Height)/2
            } else if f.style.CrossAxis == CrossEnd {
                y = f.style.Padding.Top + availableHeight - finalSizes[i].Height
            }
        } else {
            x = f.style.Padding.Left
            if f.style.CrossAxis == CrossCenter {
                x = f.style.Padding.Left + (availableWidth-finalSizes[i].Width)/2
            } else if f.style.CrossAxis == CrossEnd {
                x = f.style.Padding.Left + availableWidth - finalSizes[i].Width
            }
        }

        // Calculate main-axis position
        if isRow {
            x = mainPos
            if isReverse {
                x -= finalSizes[i].Width
            }
            mainPos += finalSizes[i].Width
        } else {
            y = mainPos
            if isReverse {
                y -= finalSizes[i].Height
            }
            mainPos += finalSizes[i].Height
        }

        boxes[i] = LayoutBox{
            ID:      f.getChildID(i),
            Box:     Rect{X: x, Y: y, Width: finalSizes[i].Width, Height: finalSizes[i].Height},
            Size:    finalSizes[i],
            Props:   f.getChildProps(i),
        }
    }

    return boxes
}
```

---

## 4. 实施计划

### 4.1 阶段划分

```
阶段 0：准备工作（1-2天）
├── 创建新的包结构
├── 设置测试框架
└── 编写接口设计文档

阶段 1：抽象接口层（3-5天）
├── 实现 LayoutNode 接口
├── 实现 LayoutInfo 结构
├── 实现 Constraints 增强
├── 实现适配器（VNodeAdapter, FiberAdapter）
└── 编写单元测试

阶段 2：核心算法增强（5-7天）
├── 移植单次测量（LayoutMeasurer）
├── 增强 FlexLayout（shrink/basis/reverse）
├── 优化布局算法性能
└── 编写集成测试

阶段 3：缓存优化（3-4天）
├── 实现智能缓存策略
├── 实现 Dirty Tracking
├── 集成到 Engine
└── 性能基准测试

阶段 4：高级功能（4-5天）
├── 实现 HitMap
├── 实现 Bounds Validator
├── 添加文本对齐支持
├── 添加边框处理
└── 添加表格布局

阶段 5：集成与测试（3-4天）
├── 与现有系统集成
├── 端到端测试
├── 性能测试
└── 文档编写

阶段 6：迁移与优化（5-7天）
├── 逐步迁移现有代码
├── 性能调优
├── 修复发现的问题
└── 最终验证
```

### 4.2 详细任务清单

#### 阶段 0：准备工作
- [ ] 创建 `runtime/layout/v2/` 目录
- [ ] 设计接口文档（Markdown）
- [ ] 设置测试框架（`layout/v2_test.go`）
- [ ] 准备基准测试套件

#### 阶段 1：抽象接口层
- [ ] 实现 `LayoutNode` 接口（`node.go`）
- [ ] 实现 `LayoutInfo` 结构（`layout_info.go`）
- [ ] 增强 `Constraints` 方法（`constraints.go`）
- [ ] 实现 `VNodeAdapter`（`adapter_vnode.go`）
- [ ] 实现 `FiberAdapter`（`adapter_fiber.go`）
- [ ] 编写接口测试（`node_test.go`, `adapter_test.go`）

#### 阶段 2：核心算法增强
- [ ] 实现 `LayoutMeasurer` 接口（`measurer.go`）
- [ ] 增强 `FlexLayout.Measure()`（`flex_measure.go`）
  - [ ] 添加 shrink 支持
  - [ ] 添加 basis 支持
  - [ ] 添加 reverse 支持
- [ ] 实现 `FlexLayout.MeasureLayout()`（`flex_single_pass.go`）
- [ ] 优化布局计算性能
- [ ] 编写 Flexbox 测试（`flex_enhanced_test.go`）

#### 阶段 3：缓存优化
- [ ] 实现 `LayoutCache`（`cache.go`）
  - [ ] 支持叶子节点缓存
  - [ ] 支持 LRU 驱逐
  - [ ] 支持按 NodeID 失效
- [ ] 实现 `DirtyTracker`（`dirty_tracker.go`）
- [ ] 集成到 `Engine`（`engine.go`）
- [ ] 编写缓存测试（`cache_test.go`）
- [ ] 编写性能基准测试（`cache_bench_test.go`）

#### 阶段 4：高级功能
- [ ] 实现 `HitMap`（`hitmap.go`）
- [ ] 实现 `BoundsValidator`（`validator.go`）
- [ ] 添加文本对齐支持（`text_align.go`）
- [ ] 添加边框处理（`border.go`）
- [ ] 添加表格布局（`table.go`）
- [ ] 编写功能测试

#### 阶段 5：集成与测试
- [ ] 创建集成测试套件（`integration_test.go`）
- [ ] 端到端场景测试
- [ ] 性能对比测试（vs compute/）
- [ ] 编写使用文档（`README_v2.md`）
- [ ] 编写 API 文档（`godoc`）

#### 阶段 6：迁移与优化
- [ ] 迁移 `runtime/compute` 用户到新 API
- [ ] 性能调优和瓶颈分析
- [ ] 修复发现的问题
- [ ] 代码审查和重构
- [ ] 最终验证和发布准备

---

## 5. 文件结构设计

```
runtime/layout/v2/
├── README.md                    # 新 API 文档
├── node.go                     # 核心抽象接口
├── constraints.go              # 约束系统增强
├── layout_info.go            # 布局信息
├── flex.go                   # Flexbox 算法
├── flex_enhanced.go         # 增强 Flexbox
├── flex_single_pass.go       # 单次测量
├── table.go                 # 表格布局
├── border.go                # 边框处理
├── text_align.go            # 文本对齐
├── cache.go                 # 缓存系统
├── dirty_tracker.go         # Dirty Tracking
├── hitmap.go               # HitMap
├── validator.go             # 边界验证
├── engine.go                # 布局引擎
├── adapter_vnode.go        # VNode 适配器
├── adapter_fiber.go        # Fiber 适配器
├── adapters.go             # 通用适配器
├── result.go                # 布局结果
│
├── node_test.go            # 接口测试
├── flex_test.go            # Flexbox 测试
├── cache_test.go           # 缓存测试
├── integration_test.go      # 集成测试
└── benchmark_test.go        # 性能基准
```

---

## 6. 向后兼容性

### 6.1 兼容策略

**目标**：现有代码可以无缝迁移到新 API

```go
// 策略 1：保留旧 API，添加新 API
package layout

// 旧 API（保持不变）
type Node interface { ... }
type Engine struct { ... }

// 新 API（v2）
package layout/v2

type LayoutNode interface { ... }
type EngineV2 struct { ... }

// 提供迁移辅助
package layout/migrate

// Convert VNode to LayoutNode
func FromVNode(vnode rtui.VNode) v2.LayoutNode {
    return &v2.VNodeAdapter{vnode: vnode}
}

// Convert Fiber to LayoutNode
func FromFiber(fiber *rtui.Fiber) v2.LayoutNode {
    return &v2.FiberAdapter{fiber: fiber}
}
```

**策略 2：提供桥接层**

```go
// runtime/compute 提供 Bridge API
package compute

// 使用新的 layout/v2 引擎
type EngineV2 struct {
    v2Engine *layoutv2.Engine
}

func (e *EngineV2) Layout(
    vnode VNode,
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
) (*ComputedLayout, error) {
    // 转换到 v2 接口
    layoutNode := layoutmigrate.FromVNode(vnode)

    // 使用 v2 引擎
    result := e.v2Engine.Layout(layoutNode, constraints)

    // 转换回 compute/ 类型
    return e.convertToComputeLayout(result), nil
}
```

### 6.2 渐进式迁移路径

```
阶段 1：并存（当前）
├── runtime/compute 继续使用
├── runtime/layout/v2 可选使用
└── 两者互不干扰

阶段 2：迁移（1-2个月）
├── 新功能使用 v2 API
├── 旧功能逐步迁移
└── 性能对比验证

阶段 3：切换（2-3个月）
├── 默认使用 v2 API
├── 保留旧 API 作为兼容层
└── 文档更新

阶段 4：弃用（3-6个月）
├── 标记旧 API 为 deprecated
├── 提供迁移指南
└── 计划移除时间表
```

---

## 7. 测试策略

### 7.1 测试金字塔

```
        /\
       /  \
      / E2E \       ← 集成测试（真实场景）
     /--------\
    /  单元测试  \   ← 单元测试（快速反馈）
   /------------\
  /   性能基准测试   \  ← 性能测试（对比验证）
 /------------------\
```

### 7.2 测试覆盖目标

| 测试类型 | 当前覆盖率 | 目标覆盖率 |
|---------|-----------|------------|
| 单元测试 | 40% | 85% |
| 集成测试 | 20% | 70% |
| 性能测试 | 10% | 60% |
| 边界测试 | 30% | 80% |

### 7.3 关键测试场景

#### Flexbox 测试
```go
func TestFlexLayout_CompleteFeatures(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() *FlexLayout
        expect  Size
    }{
        {
            name: "flex-grow with reverse",
            setup: func() *FlexLayout {
                container := NewFlexLayout("test", children)
                container.SetDirection(FlexRowReverse)
                container.SetFlex(0, 2, 0)  // grow=2
                return container
            },
            expect: Size{Width: 200, Height: 50},
        },
        {
            name: "flex-shrink with basis",
            setup: func() *FlexLayout {
                container := NewFlexLayout("test", children)
                container.SetFlex(0, 0, 1)  // shrink=1
                container.SetFlexBasis(0, 100)  // basis=100
                return container
            },
            expect: Size{Width: 100, Height: 50},
        },
        // ... 更多场景
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            flex := tt.setup()
            result := flex.Measure(UnboundedConstraints())
            assert.Equal(t, tt.expect, result)
        })
    }
}
```

#### 缓存测试
```go
func TestCache_LeafNodeOptimization(t *testing.T) {
    engine := NewEngineV2()
    engine.SetCacheStrategy(CacheLeafOnly)

    // First layout - cache miss
    result1 := engine.Layout(node, constraints)
    assert.Equal(t, 0, engine.CacheStats().Hits)

    // Second layout - cache hit
    result2 := engine.Layout(node, constraints)
    assert.Equal(t, 1, engine.CacheStats().Hits)

    // Verify results are identical
    assert.Equal(t, result1.Root.Box, result2.Root.Box)
}
```

#### Dirty Tracking 测试
```go
func TestDirtyTracking_IncrementalLayout(t *testing.T) {
    engine := NewEngineV2()

    // Initial layout
    result1 := engine.Layout(root, constraints)

    // Mark single child as dirty
    engine.MarkDirty(childNodeID)

    // Incremental layout - should only re-layout dirty subtree
    result2 := engine.Layout(root, constraints)

    // Verify only dirty subtree changed
    assertLayoutBoxEquals(t, result1.FindByNodeID(childNodeID),
        result2.FindByNodeID(childNodeID))
    assertLayoutBoxEquals(t, result1.FindByNodeID(otherNodeID),
        result2.FindByNodeID(otherNodeID))
}
```

### 7.4 性能基准测试

```go
func BenchmarkLayoutV2_SimpleFlex(b *testing.B) {
    engine := NewEngineV2()
    node := createSimpleFlexTree()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = engine.Layout(node, UnboundedConstraints())
    }
}

func BenchmarkLayoutV2_ComplexNested(b *testing.B) {
    engine := NewEngineV2()
    node := createComplexNestedTree(depth=10)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = engine.Layout(node, UnboundedConstraints())
    }
}

func BenchmarkLayoutV2_WithCache(b *testing.B) {
    engine := NewEngineV2()
    engine.SetCacheStrategy(CacheLeafOnly)
    node := createSimpleFlexTree()

    // Warm up cache
    _ = engine.Layout(node, UnboundedConstraints())

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = engine.Layout(node, UnboundedConstraints())
    }
}

// 对比测试
func BenchmarkLayout_ComputeVsV2(b *testing.B) {
    // 运行两个引擎的基准测试
    b.Run("compute", func(b *testing.B) {
        // ... runtime/compute 引擎
    })

    b.Run("v2", func(b *testing.B) {
        // ... runtime/layout/v2 引擎
    })
}
```

---

## 8. 风险评估与缓解

### 8.1 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 接口设计不当 | 高 | 中 | 早期原型验证、迭代设计 |
| 性能回归 | 高 | 中 | 性能基准对比、渐进式迁移 |
| 兼容性问题 | 中 | 高 | 保留旧 API、提供适配器 |
| 缓存失效bug | 中 | 中 | 完善测试、版本控制 |
| 单次测量bug | 高 | 低 | 详细单元测试、边界用例 |

### 8.2 项目风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 时间超期 | 中 | 中 | 分阶段实施、优先级管理 |
| 资源不足 | 高 | 低 | 明确资源需求、提前协调 |
| 需求变更 | 中 | 中 | 灵活架构设计 |
| 知识传承 | 低 | 中 | 详细文档、代码审查 |

### 8.3 回滚计划

```
回滚触发条件：
❌ 性能下降 > 20%
❌ 功能缺失 > 10%
❌ 严重 bug > 5 个/周
❌ 用户反馈负面

回滚步骤：
1. 切换回 runtime/compute 引擎
2. 禁用 runtime/layout/v2
3. 修复发现的问题
4. 重新评估方案

回滚时间：< 2 小时
回滚影响：< 5% 用户
```

---

## 9. 成功指标

### 9.1 功能指标

- [ ] ✅ 支持完整的 Flexbox（grow/shrink/basis/reverse）
- [ ] ✅ 单次测量优化减少 50% 测量次数
- [ ] ✅ 缓存命中率 > 60%
- [ ] ✅ Dirty Tracking 减少计算量 > 70%
- [ ] ✅ 独立于 VNode/Fiber 架构

### 9.2 性能指标

```
基准测试目标：
├── 简单布局：< 5ms (vs compute/: 8ms)
├── 复杂嵌套：< 20ms (vs compute/: 30ms)
├── 大型树（1000 节点）：< 100ms (vs compute/: 150ms)
└── 增量更新（单个脏节点）：< 2ms (vs compute/: 10ms)

内存使用：
├── 布局内存：< 80% of compute/
├── 缓存内存：< 60% of compute/ (叶子节点优化)
└── 峰值内存：< 90% of compute/
```

### 9.3 质量指标

- [ ] ✅ 单元测试覆盖率 > 85%
- [ ] ✅ 集成测试覆盖率 > 70%
- [ ] ✅ 无严重 bug（critical/高）
- [ ] ✅ 文档完整性 > 95%
- [ ] ✅ API 易用性评分 > 4/5

---

## 10. 后续优化方向

### 10.1 短期优化（3-6个月）

1. **Grid 布局**
   - 实现标准 CSS Grid 算法
   - 支持网格模板和区域

2. **绝对定位**
   - 支持绝对定位布局
   - 与 Flexbox 混合使用

3. **动画支持**
   - 布局动画插值
   - 过渡效果

4. **响应式布局**
   - 媒体查询支持
   - 断点系统

### 10.2 长期优化（6-12个月）

1. **异步布局**
   - Web Worker 支持
   - 非阻塞布局计算

2. **GPU 加速**
   - WebGL 渲染集成
   - 硬件加速布局

3. **AI 优化**
   - 机器学习预测布局
   - 自动布局建议

4. **跨平台一致性**
   - 统一的布局行为
   - 平台特定优化

---

## 11. 总结

### 11.1 核心价值

**本方案的核心价值**：

1. **架构解耦**
   - 创建独立于 VNode/Fiber 的抽象层
   - 提高可维护性和可测试性
   - 支持多平台复用

2. **性能提升**
   - 单次测量减少 O(N²) 问题
   - 智能缓存提高命中率
   - Dirty Tracking 减少计算量

3. **功能完整**
   - 保留 `layout/` 的标准 Flexbox
   - 添加 `compute/` 的高级功能
   - 支持更复杂的布局场景

4. **平滑迁移**
   - 向后兼容现有代码
   - 渐进式迁移路径
   - 降低迁移风险

### 11.2 实施建议

**推荐实施方式**：

1. **快速原型**（1周）
   - 实现核心抽象接口
   - 验证设计可行性
   - 获取早期反馈

2. **迭代开发**（4-6周）
   - 按阶段逐步实施
   - 每个阶段验证测试
   - 及时调整方向

3. **灰度发布**（2-4周）
   - 部分功能灰度
   - 收集真实用户反馈
   - 监控性能指标

4. **全量发布**（1-2周）
   - 功能完整发布
   - 文档和培训
   - 监控和调优

### 11.3 关键成功因素

**成功的关键因素**：

1. **设计质量**
   - 抽象接口设计合理
   - API 易用且一致
   - 扩展性强

2. **性能表现**
   - 达到或超过性能目标
   - 内存使用合理
   - 无性能回归

3. **测试覆盖**
   - 单元测试完善
   - 集成测试全面
   - 性能基准可靠

4. **文档和培训**
   - 文档清晰完整
   - 使用示例丰富
   - 迁移指南详细

---

## 附录

### A. 接口映射表

| compute/ 类型 | layout/v2 类型 | 说明 |
|-------------|---------------|------|
| VNode | LayoutNode | 节点抽象 |
| BoxConstraints | Constraints | 约束系统 |
| ComputedBox | LayoutBox | 布局结果 |
| ComputedLayout | LayoutResult | 完整结果 |
| LayoutMeasurer | LayoutMeasurer | 单次测量接口 |
| ChildMeasurer | ChildMeasurer | 子节点测量 |
| LayoutCache | LayoutCache | 缓存系统 |
| DirtyTracker | DirtyTracker | 脏跟踪 |

### B. API 对比示例

```go
// compute/ API
func (e *Engine) Layout(
    vnode VNode,
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
) (*ComputedLayout, error)

// layout/v2 API
func (e *EngineV2) Layout(
    node LayoutNode,
    constraints Constraints,
) *LayoutResult

// 迁移示例
func MigrateToV2(oldEngine *compute.Engine, vnode VNode) *LayoutResult {
    v2Engine := layoutv2.NewEngine()
    layoutNode := layoutmigrate.FromVNode(vnode)
    return v2Engine.Layout(layoutNode, convertConstraints(constraints))
}
```

### C. 性能基准目标

```
场景 1：简单 Flexbox（10个节点）
├── compute/: 8ms
├── layout/v2: < 5ms  ← 目标
└── 改进: > 37%

场景 2：深度嵌套（10层）
├── compute/: 30ms
├── layout/v2: < 20ms  ← 目标
└── 改进: > 33%

场景 3：大型树（1000节点）
├── compute/: 150ms
├── layout/v2: < 100ms  ← 目标
└── 改进: > 33%

场景 4：增量更新（1个脏节点）
├── compute/: 10ms
├── layout/v2: < 2ms  ← 目标
└── 改进: > 80%
```

---

**文档版本**: v1.0
**创建日期**: 2025-02-15
**作者**: AI Assistant (Crush)
**状态**: 待审核
