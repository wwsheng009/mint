# TUI DevTools 实施文档

> **项目**: Mint TUI Runtime
> **文档版本**: 1.0
> **创建日期**: 2025-01-30
> **状态**: 设计阶段

---

## 目录

1. [概述](#一概述)
2. [当前系统分析](#二当前系统分析)
3. [DevTools 架构设计](#三devtools-架构设计)
4. [核心模块实现](#四核心模块实现)
5. [协议层设计](#五协议层设计)
6. [客户端实现](#六客户端实现)
7. [实施计划](#七实施计划)
8. [文件结构](#八文件结构)

---

## 一、概述

### 1.1 目标

为 Mint TUI Runtime 构建一套完整的调试工具系统，实现：

- **Layout Inspector**: 可视化组件布局和边界
- **Repaint Debug**: 重绘区域可视化，性能分析
- **Component Tree Viewer**: 组件树实时查看和状态检查
- **Event Trace**: 事件流追踪和传播路径可视化
- **Focus Inspector**: 焦点状态和 Tab 顺序检查

### 1.2 设计原则

| 原则 | 说明 |
|------|------|
| **零侵入** | 通过 Hook 机制集成，不修改核心代码 |
| **可插拔** | DevTools 可独立启用/禁用 |
| **低开销** | 禁用时性能影响 < 1% |
| **多客户端** | 支持 TUI/Web/VSCode 等多种客户端 |
| **生产安全** | 生产环境可完全移除 |

---

## 二、当前系统分析

### 2.1 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Mint TUI Runtime                        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐ │
│  │  Engine        │  │  Event System  │  │  Focus Manager │ │
│  │  - 主循环       │  │  - 三阶段传播   │  │  - Tab 导航     │ │
│  │  - 帧调度       │  │  - 命中测试     │  │  - Focus Trap  │ │
│  │  - 渲染管线     │  │  - 事件分发     │  │                │ │
│  └────────────────┘  └────────────────┘  └────────────────┘ │
│           │                     │                     │       │
│           └─────────────────────┴─────────────────────┘       │
│                           │                                  │
│  ┌────────────────────────────────────────────────────────┐  │
│  │              Renderer (paint 包)                       │  │
│  │  - 双缓冲渲染                                          │  │
│  │  - Diff 算法 (dirty.go)                                │  │
│  │  - 脏区域跟踪                                          │  │
│  │  - Run merging 优化                                    │  │
│  └────────────────────────────────────────────────────────┘  │
│                           │                                  │
│  ┌────────────────────────────────────────────────────────┐  │
│  │              Layout System                             │  │
│  │  - LayoutBox / LayoutNode                              │  │
│  │  - BoxConstraints                                      │  │
│  │  - Flex 布局                                            │  │
│  └────────────────────────────────────────────────────────┘  │
│                           │                                  │
│  ┌────────────────────────────────────────────────────────┐  │
│  │              Components                                │  │
│  │  - ComponentRef                                        │  │
│  │  - Measurable / Renderer                               │  │
│  │  - FocusableComponent                                  │  │
│  └────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 现有调试能力

| 模块 | 文件 | 能力 | 可用于 DevTools |
|------|------|------|-----------------|
| Render Debug | `runtime/debug.go` | Frame 分析、Box 信息、JSON 输出 | ✅ 布局检查器基础 |
| Event Dispatch | `runtime/event/dispatch.go` | 三阶段传播、EventResult | ✅ 事件追踪 |
| Dirty Tracker | `runtime/paint/dirty.go` | 脏区域检测、合并优化 | ✅ 重绘调试 |
| Focus Manager | `runtime/focus/manager.go` | 焦点管理、Tab 顺序 | ✅ 焦点检查器 |
| Layout System | `runtime/layout/` | 布局计算、约束 | ✅ 组件树构建 |

### 2.3 Hook 点分析

| 位置 | Hook 类型 | 用途 |
|------|----------|------|
| `Engine.frame()` | Render Hook | 收集布局信息、绘制调试边框 |
| `Renderer.Render()` | Repaint Hook | 记录脏区域、计算渲染统计 |
| `DispatchEvent()` | Event Hook | 记录事件流、传播路径 |
| `FocusManager` 操作 | Focus Hook | 记录焦点变化、验证状态 |
| `Component.Paint()` | Component Hook | 组件级调试信息收集 |

---

## 三、DevTools 架构设计

### 3.1 整体架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                         App Runtime                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Debug Hook Layer                          │   │
│  │                                                              │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │   │
│  │  │ RenderHook   │  │ EventHook    │  │ FocusHook    │      │   │
│  │  │              │  │              │  │              │      │   │
│  │  │ - PrePaint   │  │ - PreDispatch│  │ - PreChange  │      │   │
│  │  │ - PostPaint  │  │ - PostDispatch│ │ - PostChange │      │   │
│  │  │ - DirtyRect  │  │ - Propagation│  │ - Validate   │      │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘      │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                   Data Collector                             │   │
│  │                                                              │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │   │
│  │  │ LayoutData   │  │ EventData    │  │ FocusData    │      │   │
│  │  │ - Boxes      │  │ - Trace      │  │ - State      │      │   │
│  │  │ - Tree       │  │ - Path       │  │ - Order      │      │   │
│  │  │ - Metrics    │  │ - Timestamp  │  │ - History    │      │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘      │   │
│  │                                                              │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │   │
│  │  │ RepaintData  │  │ ComponentData│  │ MetricsData  │      │   │
│  │  │ - DirtyRects │  │ - Props      │  │ - FPS        │      │   │
│  │  │ - CellCount  │  │ - State      │  │ - FrameTime  │      │   │
│  │  │ - Blended    │  │ - Children   │  │ - Memory     │      │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘      │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                   Protocol Layer                             │   │
│  │                                                              │   │
│  │              DebugMessage { Type, Payload }                  │   │
│  │                                                              │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │   │
│  │  │ Encoder      │  │ Transport    │  │ Decoder      │      │   │
│  │  │ - JSON       │  │ - Channel    │  │ - JSON       │      │   │
│  │  │ - Binary     │  │ - WebSocket  │  │ - Binary     │      │   │
│  │  │ - CBOR       │  │ - HTTP       │  │ - CBOR       │      │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘      │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         DevTools Clients                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │ TUI Panel    │  │ Web Dashboard│  │ VSCode       │             │
│  │              │  │              │  │ Extension    │             │
│  │ - 内嵌调试    │  │ - 可视化图表  │  │ - IDE 集成   │             │
│  │ - 快捷键切换  │  │ - 远程调试    │  │ - 断点支持   │             │
│  │ - 实时更新    │  │ - 历史回放    │  │ - 代码跳转   │             │
│  └──────────────┘  └──────────────┘  └──────────────┘             │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 核心接口设计

```go
// runtime/devtypes.go - DevTools 核心类型

// Collector 数据收集器接口
type Collector interface {
    // 启用/禁用收集
    Enable(flags DebugFlag)
    Disable(flags DebugFlag)
    IsEnabled(flag DebugFlag) bool

    // 获取收集的数据
    GetLayout() *LayoutSnapshot
    GetRepaintLog() []*RepaintEntry
    GetEventLog() []*EventEntry
    GetFocusState() *FocusSnapshot

    // 清空数据
    Clear(flags DebugFlag)
}

// Hook 钩子接口
type Hook interface {
    // 优先级，数字越小越先执行
    Priority() int

    // 是否启用
    Enabled() bool
}

// RenderHook 渲染钩子
type RenderHook interface {
    Hook
    PrePaint(buf *paint.Buffer)
    PostPaint(buf *paint.Buffer, dirty *paint.DiffResult)
}

// EventHook 事件钩子
type EventHook interface {
    Hook
    PreDispatch(ev *event.EventStruct)
    PostDispatch(ev *event.EventStruct, result event.EventResult)
}

// FocusHook 焦点钩子
type FocusHook interface {
    Hook
    PreChange(from, to string)
    PostChange(from, to string)
}
```

---

## 四、核心模块实现

### 4.1 Layout Inspector（布局检查器）

#### 功能需求

1. 可视化组件边界（支持多层级）
2. 显示组件属性（ID、Type、Size、Position）
3. 高亮选中组件
4. 显示组件层级关系

#### 数据结构

```go
// runtime/devtools/layout.go

package devtools

import (
    "sync"
    "time"

    "github.com/wwsheng009/mint/runtime"
)

// LayoutSnapshot 布局快照
type LayoutSnapshot struct {
    Time      time.Time
    Frame     int
    Boxes     []BoxInfo
    Tree      *TreeNode
    Metrics   LayoutMetrics
}

// BoxInfo 组件盒信息
type BoxInfo struct {
    ID     string
    Type   string
    Rect   runtime.Rect     // X, Y, W, H
    ZIndex int

    // 布局属性
    Margin   runtime.Insets
    Padding  runtime.Insets
    Flex     runtime.FlexConfig

    // 渲染信息
    Visible      bool
    Clip         bool
    Transform    runtime.Transform

    // 统计
    CellCount    int
    NonEmptyCell int
}

// TreeNode 组件树节点
type TreeNode struct {
    ID       string
    Type     string
    Rect     runtime.Rect
    Children []*TreeNode

    // 扩展信息
    Props    map[string]interface{}
    State    map[string]interface{}
}

// LayoutMetrics 布局指标
type LayoutMetrics struct {
    TotalBoxes      int
    VisibleBoxes    int
    TotalCells      int
    UsedCells       int
    Density         float64
    MaxDepth        int
    OverlapCount    int
}

// LayoutCollector 布局数据收集器
type LayoutCollector struct {
    mu              sync.RWMutex
    enabled         bool
    snapshots       []*LayoutSnapshot
    maxSnapshots    int
    currentFrame    int
}

// NewLayoutCollector 创建布局收集器
func NewLayoutCollector() *LayoutCollector {
    return &LayoutCollector{
        enabled:      false,
        snapshots:    make([]*LayoutSnapshot, 0, 100),
        maxSnapshots: 100,
    }
}

// Collect 收集当前帧的布局信息
func (lc *LayoutCollector) Collect(boxes []runtime.LayoutBox, root runtime.Renderable) *LayoutSnapshot {
    if !lc.enabled {
        return nil
    }

    lc.mu.Lock()
    defer lc.mu.Unlock()

    snapshot := &LayoutSnapshot{
        Time:  time.Now(),
        Frame: lc.currentFrame,
        Boxes: lc.extractBoxInfo(boxes),
        Tree:  lc.buildTree(root, boxes),
    }

    snapshot.Metrics = lc.calculateMetrics(snapshot)

    // 保留最近 N 个快照
    lc.snapshots = append(lc.snapshots, snapshot)
    if len(lc.snapshots) > lc.maxSnapshots {
        lc.snapshots = lc.snapshots[1:]
    }
    lc.currentFrame++

    return snapshot
}

// extractBoxInfo 从 LayoutBox 提取信息
func (lc *LayoutCollector) extractBoxInfo(boxes []runtime.LayoutBox) []BoxInfo {
    info := make([]BoxInfo, len(boxes))

    for i, box := range boxes {
        info[i] = BoxInfo{
            ID:     box.NodeID,
            Type:   lc.getComponentType(box),
            Rect:   runtime.Rect{X: box.X, Y: box.Y, W: box.W, H: box.H},
            ZIndex: box.ZIndex,
        }
    }

    return info
}

// buildTree 构建组件树
func (lc *LayoutCollector) buildTree(root runtime.Renderable, boxes []runtime.LayoutBox) *TreeNode {
    tree := &TreeNode{
        ID:   root.ID(),
        Type: lc.getTypeName(root),
    }

    // 从 LayoutBox 构建树结构
    boxMap := make(map[string]*runtime.LayoutNode)
    for _, box := range boxes {
        if box.Node != nil {
            boxMap[box.NodeID] = box.Node
        }
    }

    // 递归构建
    for _, box := range boxes {
        if box.Node != nil {
            node := lc.nodeToTreeNode(box.Node)
            tree.Children = append(tree.Children, node)
        }
    }

    return tree
}

// nodeToTreeNode 将 LayoutNode 转换为 TreeNode
func (lc *LayoutCollector) nodeToTreeNode(node *runtime.LayoutNode) *TreeNode {
    treeNode := &TreeNode{
        ID:     node.ID,
        Type:   node.Component.Type,
        Rect:   runtime.Rect{X: node.X, Y: node.Y, W: node.MeasuredWidth, H: node.MeasuredHeight},
    }

    for _, child := range node.Children {
        treeNode.Children = append(treeNode.Children, lc.nodeToTreeNode(child))
    }

    return treeNode
}

// calculateMetrics 计算布局指标
func (lc *LayoutCollector) calculateMetrics(snapshot *LayoutSnapshot) LayoutMetrics {
    metrics := LayoutMetrics{
        TotalBoxes: len(snapshot.Boxes),
    }

    totalCells := 0
    usedCells := 0
    maxDepth := 0

    for _, box := range snapshot.Boxes {
        if box.Visible {
            metrics.VisibleBoxes++
        }
        totalCells += box.Rect.W * box.Rect.H
        usedCells += box.NonEmptyCell

        depth := lc.getDepth(snapshot.Tree, box.ID)
        if depth > maxDepth {
            maxDepth = depth
        }
    }

    metrics.TotalCells = totalCells
    metrics.UsedCells = usedCells
    metrics.Density = float64(usedCells) / float64(totalCells)
    metrics.MaxDepth = maxDepth

    return metrics
}

// getDepth 获取节点深度
func (lc *LayoutCollector) getDepth(tree *TreeNode, id string) int {
    return lc.depthHelper(tree, id, 0)
}

func (lc *LayoutCollector) depthHelper(node *TreeNode, id string, current int) int {
    if node.ID == id {
        return current
    }

    for _, child := range node.Children {
        if depth := lc.depthHelper(child, id, current+1); depth > 0 {
            return depth
        }
    }

    return 0
}

// GetComponentType 获取组件类型
func (lc *LayoutCollector) getComponentType(box runtime.LayoutBox) string {
    if box.Node != nil && box.Node.Component != nil {
        return box.Node.Component.Type
    }
    return "unknown"
}

// GetTypeName 获取类型名称
func (lc *LayoutCollector) getTypeName(v interface{}) string {
    return runtime.GetTypeName(v)
}

// GetLatest 获取最新的快照
func (lc *LayoutCollector) GetLatest() *LayoutSnapshot {
    lc.mu.RLock()
    defer lc.mu.RUnlock()

    if len(lc.snapshots) == 0 {
        return nil
    }
    return lc.snapshots[len(lc.snapshots)-1]
}

// GetAll 获取所有快照
func (lc *LayoutCollector) GetAll() []*LayoutSnapshot {
    lc.mu.RLock()
    defer lc.mu.RUnlock()

    result := make([]*LayoutSnapshot, len(lc.snapshots))
    copy(result, lc.snapshots)
    return result
}

// Enable 启用收集
func (lc *LayoutCollector) Enable() {
    lc.mu.Lock()
    defer lc.mu.Unlock()
    lc.enabled = true
}

// Disable 禁用收集
func (lc *LayoutCollector) Disable() {
    lc.mu.Lock()
    defer lc.mu.Unlock()
    lc.enabled = false
}

// Clear 清空快照
func (lc *LayoutCollector) Clear() {
    lc.mu.Lock()
    defer lc.mu.Unlock()
    lc.snapshots = lc.snapshots[:0]
    lc.currentFrame = 0
}
```

#### 引擎集成

```go
// runtime/engine/engine.go - 添加 LayoutInspector 支持

type Engine struct {
    // ... 现有字段

    // DevTools 支持
    layoutCollector   *devtools.LayoutCollector
    repaintCollector  *devtools.RepaintCollector
    eventCollector    *devtools.EventCollector
    focusCollector    *devtools.FocusCollector

    debugMode         bool
    debugOverlay      bool
}

func (e *Engine) frame() {
    startTime := time.Now()

    // 1. 获取后缓冲区
    buf := e.renderer.GetBackBuffer()

    // 2. 清空（可选，取决于渲染策略）
    // buf.Clear()

    // 3. 调用组件的 Paint
    e.rootMu.RLock()
    root := e.root
    e.rootMu.RUnlock()

    if root != nil {
        root.Paint(buf)
    }

    // 4. 收集布局信息
    if e.layoutCollector != nil && e.layoutCollector.IsEnabled() {
        snapshot := e.layoutCollector.Collect(e.layoutBoxes, root)
        _ = snapshot // 存储以便后续查询
    }

    // 5. 绘制调试覆盖层
    if e.debugOverlay {
        e.drawDebugOverlay(buf)
    }

    // 6. 渲染
    output := e.renderer.Render()

    // 7. 收集重绘信息
    if e.repaintCollector != nil && e.repaintCollector.IsEnabled() {
        e.repaintCollector.Collect(e.renderer.GetStats())
    }

    // 8. 输出
    if output != "" {
        e.outputFunc(output)
    }

    // 9. 记录帧时间
    frameTime := time.Since(startTime)
    if e.metricsCollector != nil {
        e.metricsCollector.RecordFrame(frameTime)
    }
}

// drawDebugOverlay 绘制调试覆盖层
func (e *Engine) drawDebugOverlay(buf *paint.Buffer) {
    const debugChar = '·'
    debugStyle := style.Style{}.
        Foreground(style.RGB(255, 255, 0)).
        Background(style.RGB(100, 100, 0))

    e.layoutMu.RLock()
    boxes := e.layoutBoxes
    e.layoutMu.RUnlock()

    for _, box := range boxes {
        // 绘制组件边框
        e.drawBoxBorder(buf, box, debugStyle)

        // 绘制组件 ID（空间足够时）
        if box.W > len(box.NodeID) + 2 {
            buf.SetString(box.X+1, box.Y, box.NodeID, debugStyle)
        }
    }
}

// drawBoxBorder 绘制单个盒子的边框
func (e *Engine) drawBoxBorder(buf *paint.Buffer, box runtime.LayoutBox, s style.Style) {
    // 上边
    for x := box.X; x < box.X+box.W && x < buf.Width; x++ {
        if box.Y < buf.Height {
            buf.SetCell(x, box.Y, '─', s)
        }
    }
    // 下边
    for x := box.X; x < box.X+box.W && x < buf.Width; x++ {
        if box.Y+box.H-1 >= 0 && box.Y+box.H-1 < buf.Height {
            buf.SetCell(x, box.Y+box.H-1, '─', s)
        }
    }
    // 左边
    for y := box.Y; y < box.Y+box.H && y < buf.Height; y++ {
        if box.X >= 0 && box.X < buf.Width {
            buf.SetCell(box.X, y, '│', s)
        }
    }
    // 右边
    for y := box.Y; y < box.Y+box.H && y < buf.Height; y++ {
        if box.X+box.W-1 >= 0 && box.X+box.W-1 < buf.Width {
            buf.SetCell(box.X+box.W-1, y, '│', s)
        }
    }
}
```

### 4.2 Repaint Debug（重绘调试）

#### 功能需求

1. 追踪每帧的重绘区域
2. 高亮脏区域（闪烁效果）
3. 统计渲染性能指标
4. 检测过度重绘

#### 数据结构

```go
// runtime/devtools/repaint.go

package devtools

import (
    "sync"
    "time"

    "github.com/wwsheng009/mint/runtime/paint"
)

// RepaintEntry 重绘条目
type RepaintEntry struct {
    Time         time.Time
    Frame        int

    // 脏区域信息
    DirtyRegions []paint.Rect
    RegionCount  int

    // 变化统计
    ChangedCells int
    TotalCells   int
    ChangeRatio  float64

    // 渲染统计
    OutputBytes  int
    RenderTime   time.Duration
}

// RepaintCollector 重绘数据收集器
type RepaintCollector struct {
    mu             sync.RWMutex
    enabled        bool
    log            []*RepaintEntry
    maxEntries     int
    currentFrame   int

    // 性能统计
    totalChanged   int64
    totalCells     int64
    frameCount     int64
}

// NewRepaintCollector 创建重绘收集器
func NewRepaintCollector() *RepaintCollector {
    return &RepaintCollector{
        enabled:    false,
        log:        make([]*RepaintEntry, 0, 1000),
        maxEntries: 1000,
    }
}

// Collect 收集重绘数据
func (rc *RepaintCollector) Collect(stats paint.RenderStats, dirtyRegions []paint.Rect) {
    if !rc.enabled {
        return
    }

    rc.mu.Lock()
    defer rc.mu.Unlock()

    totalCells := stats.Width * stats.Height
    entry := &RepaintEntry{
        Time:         time.Now(),
        Frame:        rc.currentFrame,
        DirtyRegions: dirtyRegions,
        RegionCount:  len(dirtyRegions),
        ChangedCells: stats.ChangedCells,
        TotalCells:   totalCells,
        ChangeRatio:  float64(stats.ChangedCells) / float64(totalCells),
        OutputBytes:  stats.OutputBytes,
    }

    rc.log = append(rc.log, entry)
    if len(rc.log) > rc.maxEntries {
        rc.log = rc.log[1:]
    }

    rc.totalChanged += int64(stats.ChangedCells)
    rc.totalCells += int64(totalCells)
    rc.frameCount++
    rc.currentFrame++
}

// GetStats 获取统计信息
type RepaintStats struct {
    AvgChangedCells    float64
    AvgChangeRatio     float64
    MaxChangedCells    int
    PeakRegionCount    int
    TotalFrames        int64
}

func (rc *RepaintCollector) GetStats() *RepaintStats {
    rc.mu.RLock()
    defer rc.mu.RUnlock()

    if len(rc.log) == 0 {
        return nil
    }

    stats := &RepaintStats{
        TotalFrames: rc.frameCount,
    }

    var sumChanged, sumRatio int64
    var maxChanged int

    for _, entry := range rc.log {
        sumChanged += int64(entry.ChangedCells)
        sumRatio += int64(entry.ChangeRatio * 1000)

        if entry.ChangedCells > maxChanged {
            maxChanged = entry.ChangedCells
        }
        if entry.RegionCount > stats.PeakRegionCount {
            stats.PeakRegionCount = entry.RegionCount
        }
    }

    stats.AvgChangedCells = float64(sumChanged) / float64(len(rc.log))
    stats.AvgChangeRatio = float64(sumRatio) / float64(len(rc.log)) / 1000
    stats.MaxChangedCells = maxChanged

    return stats
}

// DetectOverpaint 检测过度重绘
// 返回频繁重绘的区域列表
func (rc *RepaintCollector) DetectOverpaint(threshold float64) []paint.Rect {
    rc.mu.RLock()
    defer rc.mu.RUnlock()

    // 统计每个区域的重绘次数
    regionFreq := make(map[string]int)
    regionRect := make(map[string]paint.Rect)

    for _, entry := range rc.log {
        for _, rect := range entry.DirtyRegions {
            key := rc.rectKey(rect)
            regionFreq[key]++
            regionRect[key] = rect
        }
    }

    // 找出超过阈值的区域
    var result []paint.Rect
    thresholdCount := int(float64(len(rc.log)) * threshold)

    for key, count := range regionFreq {
        if count > thresholdCount {
            result = append(result, regionRect[key])
        }
    }

    return result
}

// rectKey 生成区域的唯一键
func (rc *RepaintCollector) rectKey(r paint.Rect) string {
    return fmt.Sprintf("%d,%d,%d,%d", r.X, r.Y, r.W, r.H)
}
```

### 4.3 Component Tree Viewer（组件树查看器）

#### 功能需求

1. 实时显示组件树结构
2. 查看组件 Props 和 State
3. 定位组件在屏幕上的位置
4. 搜索和过滤组件

#### 数据结构

```go
// runtime/devtools/tree.go

package devtools

import (
    "sync"

    "github.com/wwsheng009/mint/runtime"
)

// ComponentNode 组件树节点
type ComponentNode struct {
    ID        string
    Type      string
    Rect      runtime.Rect
    Visible   bool

    // 子组件
    Children  []*ComponentNode

    // 组件详情
    Info      *ComponentInfo
}

// ComponentInfo 组件详细信息
type ComponentInfo struct {
    // 基本属性
    Props     map[string]interface{}
    State     map[string]interface{}

    // 布局信息
    Layout    *LayoutInfo

    // 事件处理器
    Handlers  []string

    // 焦点状态
    Focusable bool
    Focused   bool
}

// LayoutInfo 布局详情
type LayoutInfo struct {
    Constraints runtime.BoxConstraints
    Size        runtime.Size
    Position    runtime.Point
    Flex        runtime.FlexConfig
}

// TreeSnapshot 组件树快照
type TreeSnapshot struct {
    Root    *ComponentNode
    Flattened []*ComponentNode  // 扁平化列表，便于搜索
    Time    time.Time
}

// TreeCollector 组件树收集器
type TreeCollector struct {
    mu            sync.RWMutex
    enabled       bool
    snapshots     []*TreeSnapshot
    maxSnapshots  int
}

// NewTreeCollector 创建组件树收集器
func NewTreeCollector() *TreeCollector {
    return &TreeCollector{
        enabled:      false,
        snapshots:    make([]*TreeSnapshot, 0, 50),
        maxSnapshots: 50,
    }
}

// Collect 收集组件树
func (tc *TreeCollector) Collect(root *runtime.LayoutNode) *TreeSnapshot {
    if !tc.enabled {
        return nil
    }

    tc.mu.Lock()
    defer tc.mu.Unlock()

    snapshot := &TreeSnapshot{
        Root:      tc.buildNode(root),
        Time:      time.Now(),
    }

    // 构建扁平化列表
    snapshot.Flattened = tc.flatten(snapshot.Root)

    tc.snapshots = append(tc.snapshots, snapshot)
    if len(tc.snapshots) > tc.maxSnapshots {
        tc.snapshots = tc.snapshots[1:]
    }

    return snapshot
}

// buildNode 递归构建组件节点
func (tc *TreeCollector) buildNode(node *runtime.LayoutNode) *ComponentNode {
    if node == nil {
        return nil
    }

    cn := &ComponentNode{
        ID:      node.ID,
        Type:    tc.getComponentType(node),
        Rect:    runtime.Rect{
            X: node.X,
            Y: node.Y,
            W: node.MeasuredWidth,
            H: node.MeasuredHeight,
        },
        Visible: node.Style.Visible,
        Info:    tc.extractInfo(node),
    }

    for _, child := range node.Children {
        cn.Children = append(cn.Children, tc.buildNode(child))
    }

    return cn
}

// extractInfo 提取组件信息
func (tc *TreeCollector) extractInfo(node *runtime.LayoutNode) *ComponentInfo {
    info := &ComponentInfo{
        Props:    make(map[string]interface{}),
        State:    make(map[string]interface{}),
        Layout: &LayoutInfo{
            Size: runtime.Size{
                Width:  node.MeasuredWidth,
                Height: node.MeasuredHeight,
            },
        },
    }

    // 提取 Props（如果有）
    if propsProvider, ok := node.Component.Instance.(PropsProvider); ok {
        info.Props = propsProvider.GetProps()
    }

    // 提取 State（如果有）
    if stateProvider, ok := node.Component.Instance.(StateProvider); ok {
        info.State = stateProvider.GetState()
    }

    // 检查焦点能力
    if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
        info.Focusable = true
        info.Focused = false // 需要从 FocusManager 获取
    }

    return info
}

// flatten 扁平化组件树
func (tc *TreeCollector) flatten(root *ComponentNode) []*ComponentNode {
    var result []*ComponentNode

    var traverse func(*ComponentNode)
    traverse = func(node *ComponentNode) {
        if node == nil {
            return
        }
        result = append(result, node)
        for _, child := range node.Children {
            traverse(child)
        }
    }

    traverse(root)
    return result
}

// Search 搜索组件
func (tc *TreeCollector) Search(query string) []*ComponentNode {
    tc.mu.RLock()
    defer tc.mu.RUnlock()

    if len(tc.snapshots) == 0 {
        return nil
    }

    latest := tc.snapshots[len(tc.snapshots)-1]
    var result []*ComponentNode

    for _, node := range latest.Flattened {
        if tc.matchQuery(node, query) {
            result = append(result, node)
        }
    }

    return result
}

// matchQuery 检查节点是否匹配查询
func (tc *TreeCollector) matchQuery(node *ComponentNode, query string) bool {
    // 按ID匹配
    if strings.Contains(strings.ToLower(node.ID), strings.ToLower(query)) {
        return true
    }

    // 按Type匹配
    if strings.Contains(strings.ToLower(node.Type), strings.ToLower(query)) {
        return true
    }

    return false
}
```

### 4.4 Event Trace（事件追踪）

#### 功能需求

1. 记录所有事件及其传播路径
2. 显示事件处理结果
3. 支持事件过滤和搜索
4. 事件回放功能

#### 数据结构

```go
// runtime/devtools/trace.go

package devtools

import (
    "sync"
    "time"

    "github.com/wwsheng009/mint/runtime/event"
)

// EventEntry 事件条目
type EventEntry struct {
    Time     time.Time
    Frame    int

    // 事件信息
    Type     event.EventType
    Phase    event.EventPhase

    // 传播路径
    Path     []string  // 事件传播经过的组件ID

    // 目标信息
    Target   string    // 命中测试的目标
    Handled  bool      // 是否被处理
    Updated  bool      // 是否触发更新

    // 详细数据
    Mouse    *event.MouseEvent
    Key      *event.KeyEvent
}

// EventCollector 事件收集器
type EventCollector struct {
    mu           sync.RWMutex
    enabled      bool
    log          []*EventEntry
    maxEntries   int
    currentFrame int

    // 过滤器
    filters      []EventFilter
}

// EventFilter 事件过滤器
type EventFilter struct {
    Types  []event.EventType
    Phases []event.EventPhase
    Target string
}

// NewEventCollector 创建事件收集器
func NewEventCollector() *EventCollector {
    return &EventCollector{
        enabled:    false,
        log:        make([]*EventEntry, 0, 1000),
        maxEntries: 1000,
    }
}

// Record 记录事件
func (ec *EventCollector) Record(ev *event.EventStruct, result event.EventResult, path []string) {
    if !ec.enabled {
        return
    }

    // 检查过滤器
    if !ec.matchFilters(ev) {
        return
    }

    ec.mu.Lock()
    defer ec.mu.Unlock()

    entry := &EventEntry{
        Time:    time.Now(),
        Frame:   ec.currentFrame,
        Type:    ev.Type(),
        Phase:   ev.Phase(),
        Path:    path,
        Handled: result.Handled,
        Updated: result.Updated,
    }

    // 记录目标
    if ev.Target() != nil {
        entry.Target = ev.Target().ID()
    }

    // 记录详细数据
    switch ev.Type() {
    case event.EventMousePress, event.EventMouseRelease, event.EventMouseMove:
        if ev.Mouse != nil {
            mouseCopy := *ev.Mouse
            entry.Mouse = &mouseCopy
        }
    case event.EventKeyPress, event.EventKeyRelease:
        if ev.Key != nil {
            keyCopy := *ev.Key
            entry.Key = &keyCopy
        }
    }

    ec.log = append(ec.log, entry)
    if len(ec.log) > ec.maxEntries {
        ec.log = ec.log[1:]
    }
}

// matchFilters 检查是否匹配过滤器
func (ec *EventCollector) matchFilters(ev *event.EventStruct) bool {
    if len(ec.filters) == 0 {
        return true
    }

    for _, filter := range ec.filters {
        if ec.matchFilter(ev, filter) {
            return true
        }
    }

    return false
}

// matchFilter 检查单个过滤器
func (ec *EventCollector) matchFilter(ev *event.EventStruct, filter EventFilter) bool {
    // 类型过滤
    if len(filter.Types) > 0 {
        typeMatch := false
        for _, t := range filter.Types {
            if ev.Type() == t {
                typeMatch = true
                break
            }
        }
        if !typeMatch {
            return false
        }
    }

    // 阶段过滤
    if len(filter.Phases) > 0 {
        phaseMatch := false
        for _, p := range filter.Phases {
            if ev.Phase() == p {
                phaseMatch = true
                break
            }
        }
        if !phaseMatch {
            return false
        }
    }

    return true
}

// GetLog 获取事件日志
func (ec *EventCollector) GetLog(offset, limit int) []*EventEntry {
    ec.mu.RLock()
    defer ec.mu.RUnlock()

    if offset >= len(ec.log) {
        return nil
    }

    end := offset + limit
    if end > len(ec.log) {
        end = len(ec.log)
    }

    result := make([]*EventEntry, end-offset)
    copy(result, ec.log[offset:end])
    return result
}

// GetStats 获取事件统计
type EventStats struct {
    TotalEvents    int64
    ByType         map[event.EventType]int64
    HandledRatio   float64
    AvgPropagation float64
}

func (ec *EventCollector) GetStats() *EventStats {
    ec.mu.RLock()
    defer ec.mu.RUnlock()

    stats := &EventStats{
        TotalEvents: int64(len(ec.log)),
        ByType:      make(map[event.EventType]int64),
    }

    var handled int64
    var totalPath int

    for _, entry := range ec.log {
        stats.ByType[entry.Type]++
        if entry.Handled {
            handled++
        }
        totalPath += len(entry.Path)
    }

    if len(ec.log) > 0 {
        stats.HandledRatio = float64(handled) / float64(len(ec.log))
        stats.AvgPropagation = float64(totalPath) / float64(len(ec.log))
    }

    return stats
}
```

#### 事件系统集成

```go
// runtime/event/dispatch.go - 添加 Hook 支持

package event

// EventHook 事件钩子函数
type EventHook func(ev *EventStruct, result EventResult, path []string)

var (
    debugEventHook EventHook
    eventHooks     []EventHook
)

// SetEventHook 设置调试钩子
func SetEventHook(hook EventHook) {
    debugEventHook = hook
}

// AddEventHook 添加事件钩子
func AddEventHook(hook EventHook) {
    eventHooks = append(eventHooks, hook)
}

// DispatchEventWithHooks 带钩子的事件分发
func DispatchEventWithHooks(ev Event, boxes []runtime.LayoutBox) EventResult {
    es := ev.(*EventStruct)
    result := dispatchEventWithPropagation(es, ev.Mouse, boxes)

    // 收集传播路径
    path := collectPropagationPath(es)

    // 调用调试钩子
    if debugEventHook != nil {
        debugEventHook(es, result, path)
    }

    // 调用其他钩子
    for _, hook := range eventHooks {
        hook(es, result, path)
    }

    return result
}

// collectPropagationPath 收集传播路径
func collectPropagationPath(ev *EventStruct) []string {
    var path []string

    // 从事件中提取经过的组件
    // 这需要在事件分发过程中记录
    // 具体实现需要修改 dispatchMouseEventWithPropagation

    return path
}
```

### 4.5 Focus Inspector（焦点检查器）

#### 功能需求

1. 显示当前焦点组件
2. 显示 Tab 顺序
3. 验证焦点状态一致性
4. 检测焦点陷阱

#### 数据结构

```go
// runtime/devtools/focus.go

package devtools

import (
    "sync"
    "time"

    "github.com/wwsheng009/mint/runtime/focus"
)

// FocusSnapshot 焦点状态快照
type FocusSnapshot struct {
    Time      time.Time

    // 当前焦点
    Current   string

    // Tab 顺序
    TabOrder  []string

    // 焦点陷阱
    Traps     []*FocusTrapInfo

    // 验证结果
    Issues    []string
}

// FocusTrapInfo 焦点陷阱信息
type FocusTrapInfo struct {
    Root      string
    Scope     []string
    Cyclic    bool
}

// FocusCollector 焦点数据收集器
type FocusCollector struct {
    mu           sync.RWMutex
    enabled      bool
    snapshots    []*FocusSnapshot
    maxSnapshots int
}

// NewFocusCollector 创建焦点收集器
func NewFocusCollector() *FocusCollector {
    return &FocusCollector{
        enabled:      false,
        snapshots:    make([]*FocusSnapshot, 0, 100),
        maxSnapshots: 100,
    }
}

// Collect 从 FocusManager 收集焦点状态
func (fc *FocusCollector) Collect(fm *focus.Manager) *FocusSnapshot {
    if !fc.enabled {
        return nil
    }

    fc.mu.Lock()
    defer fc.mu.Unlock()

    snapshot := &FocusSnapshot{
        Time:     time.Now(),
        Current:  fc.getCurrent(fm),
        TabOrder: fc.getTabOrder(fm),
        Traps:    fc.getTraps(fm),
    }

    // 验证焦点状态
    snapshot.Issues = fc.validate(snapshot)

    fc.snapshots = append(fc.snapshots, snapshot)
    if len(fc.snapshots) > fc.maxSnapshots {
        fc.snapshots = fc.snapshots[1:]
    }

    return snapshot
}

// getCurrent 获取当前焦点
func (fc *FocusCollector) getCurrent(fm *focus.Manager) string {
    // 通过反射或接口获取
    if fmer, ok := fm.(interface{ Current() string }); ok {
        return fmer.Current()
    }
    return ""
}

// getTabOrder 获取 Tab 顺序
func (fc *FocusCollector) getTabOrder(fm *focus.Manager) []string {
    // 通过反射或接口获取
    if fmer, ok := fm.(interface{ TabOrder() []string }); ok {
        return fmer.TabOrder()
    }
    return nil
}

// getTraps 获取焦点陷阱
func (fc *FocusCollector) getTraps(fm *focus.Manager) []*FocusTrapInfo {
    // 通过反射或接口获取
    if fmer, ok := fm.(interface{ GetTraps() []*focus.Trap }); ok {
        traps := fmer.GetTraps()
        result := make([]*FocusTrapInfo, len(traps))
        for i, trap := range traps {
            result[i] = &FocusTrapInfo{
                Root:   trap.Root,
                Scope:  trap.Scope,
                Cyclic: trap.Cyclic,
            }
        }
        return result
    }
    return nil
}

// validate 验证焦点状态
func (fc *FocusCollector) validate(snapshot *FocusSnapshot) []string {
    var issues []string

    // 检查1: 当前焦点是否在 Tab 顺序中
    if snapshot.Current != "" {
        found := false
        for _, id := range snapshot.TabOrder {
            if id == snapshot.Current {
                found = true
                break
            }
        }
        if !found {
            issues = append(issues, fmt.Sprintf("Current focus '%s' not in tab order", snapshot.Current))
        }
    }

    // 检查2: 空焦点
    if snapshot.Current == "" && len(snapshot.TabOrder) > 0 {
        issues = append(issues, "No focus set but focusable components exist")
    }

    // 检查3: 焦点陷阱循环
    for _, trap := range snapshot.Traps {
        if trap.Cyclic && len(trap.Scope) > 1 {
            // 这是正常的，不需要警告
        }
    }

    return issues
}
```

---

## 五、协议层设计

### 5.1 调试协议

```go
// runtime/devtools/protocol.go

package devtools

import (
    "encoding/json"
    "time"
)

// MessageType 消息类型
type MessageType string

const (
    MsgLayoutSnapshot  MessageType = "layout_snapshot"
    MsgRepaintLog      MessageType = "repaint_log"
    MsgEventLog        MessageType = "event_log"
    MsgFocusState      MessageType = "focus_state"
    MsgComponentTree   MessageType = "component_tree"
    MsgMetrics         MessageType = "metrics"
    MsgCommand         MessageType = "command"
    MsgResponse        MessageType = "response"
)

// DebugMessage 调试消息
type DebugMessage struct {
    Type      MessageType    `json:"type"`
    Timestamp time.Time      `json:"timestamp"`
    Payload   json.RawMessage `json:"payload"`
}

// Encoder 消息编码器
type Encoder struct {
    // 可选：压缩、加密等
}

func (e *Encoder) Encode(msg *DebugMessage) ([]byte, error) {
    return json.Marshal(msg)
}

// Decoder 消息解码器
type Decoder struct{}

func (d *Decoder) Decode(data []byte) (*DebugMessage, error) {
    var msg DebugMessage
    err := json.Unmarshal(data, &msg)
    return &msg, err
}

// Command 命令类型
type Command struct {
    Action string                 `json:"action"`
    Params map[string]interface{} `json:"params"`
}

const (
    CmdEnable     = "enable"
    CmdDisable    = "disable"
    CmdGetSnapshot = "get_snapshot"
    CmdGetTree    = "get_tree"
    CmdSearch     = "search"
    CmdHighlight  = "highlight"
)

// Response 响应
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

### 5.2 传输层

```go
// runtime/devtools/transport.go

package devtools

import (
    "context"
)

// Transport 传输接口
type Transport interface {
    // Send 发送消息
    Send(ctx context.Context, msg *DebugMessage) error

    // Receive 接收消息
    Receive(ctx context.Context) (*DebugMessage, error)

    // Close 关闭连接
    Close() error
}

// ChannelTransport 基于内存 Channel 的传输
type ChannelTransport struct {
    sendCh chan<- *DebugMessage
    recvCh <-chan *DebugMessage
}

func NewChannelTransport(send chan<- *DebugMessage, recv <-chan *DebugMessage) *ChannelTransport {
    return &ChannelTransport{
        sendCh: send,
        recvCh: recv,
    }
}

func (t *ChannelTransport) Send(ctx context.Context, msg *DebugMessage) error {
    select {
    case t.sendCh <- msg:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (t *ChannelTransport) Receive(ctx context.Context) (*DebugMessage, error) {
    select {
    case msg := <-t.recvCh:
        return msg, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (t *ChannelTransport) Close() error {
    return nil
}
```

---

## 六、客户端实现

### 6.1 TUI Dev Panel

```go
// runtime/devtools/panel/panel.go

package panel

import (
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/devtools"
    "github.com/wwsheng009/mint/runtime/engine"
)

// PanelMode 面板模式
type PanelMode int

const (
    ModeLayout PanelMode = iota
    ModeRepaint
    ModeEvent
    ModeFocus
    ModeTree
)

// Panel TUI 调试面板
type Panel struct {
    mu           sync.Mutex
    active       bool
    mode         PanelMode
    engine       *engine.Engine
    collector    *devtools.Collector

    // UI 组件
    width        int
    height       int
    buffer       *paint.Buffer
}

// NewPanel 创建调试面板
func NewPanel(eng *engine.Engine) *Panel {
    return &Panel{
        engine:    eng,
        collector: devtools.NewCollector(),
        width:     80,
        height:    24,
    }
}

// Toggle 切换面板显示
func (p *Panel) Toggle() {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.active = !p.active
    if p.active {
        p.collector.Enable()
        go p.renderLoop()
    } else {
        p.collector.Disable()
    }
}

// SetMode 设置面板模式
func (p *Panel) SetMode(mode PanelMode) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.mode = mode
}

// renderLoop 渲染循环
func (p *Panel) renderLoop() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for p.active {
        select {
        case <-ticker.C:
            p.render()
        }
    }
}

// render 渲染面板
func (p *Panel) render() {
    p.mu.Lock()
    defer p.mu.Unlock()

    switch p.mode {
    case ModeLayout:
        p.renderLayout()
    case ModeRepaint:
        p.renderRepaint()
    case ModeEvent:
        p.renderEventLog()
    case ModeFocus:
        p.renderFocus()
    case ModeTree:
        p.renderTree()
    }
}

// renderLayout 渲染布局检查器
func (p *Panel) renderLayout() {
    snapshot := p.collector.GetLayout()
    if snapshot == nil {
        return
    }

    fmt.Println("\x1b[2J\x1b[H") // 清屏
    fmt.Println("=== Layout Inspector ===")
    fmt.Printf("Frame: %d | Boxes: %d\n", snapshot.Frame, len(snapshot.Boxes))
    fmt.Println()

    for i, box := range snapshot.Boxes {
        fmt.Printf("[%2d] %s at (%d,%d) size %dx%d",
            i, box.ID, box.Rect.X, box.Rect.Y, box.Rect.W, box.Rect.H)
        if box.Type != "" {
            fmt.Printf(" [%s]", box.Type)
        }
        fmt.Println()
    }

    fmt.Println()
    fmt.Println("Press q to quit, 1-5 to switch mode")
}
```

### 6.2 Web Dashboard

```
./devtools/web/
├── index.html
├── app.js
├── styles.css
└── ws_client.go      # WebSocket 服务端
```

---

## 七、实施计划

### 7.1 阶段划分

| 阶段 | 任务 | 优先级 | 预计时间 |
|------|------|--------|----------|
| **Phase 1** | 基础设施 | P0 | 2 天 |
| | - Hook 系统 | | |
| | - 数据收集器接口 | | |
| | - 协议层 | | |
| **Phase 2** | Layout Inspector | P0 | 2 天 |
| | - 布局快照 | | |
| | - 组件树构建 | | |
| | - TUI 面板 | | |
| **Phase 3** | 事件追踪 | P1 | 2 天 |
| | - 事件钩子 | | |
| | - 日志记录 | | |
| | - 可视化 | | |
| **Phase 4** | 重绘调试 | P1 | 1 天 |
| | - 脏区域追踪 | | |
| | - 性能统计 | | |
| **Phase 5** | 焦点检查器 | P2 | 1 天 |
| | - 状态快照 | | |
| | - 验证逻辑 | | |
| **Phase 6** | Web Dashboard | P2 | 3 天 |
| | - WebSocket 服务 | | |
| | - React 前端 | | |
| **Phase 7** | 文档和测试 | P3 | 2 天 |

### 7.2 里程碑

| 里程碑 | 标准 |
|--------|------|
| M1: 基础完成 | Hook 系统可用，数据能收集 |
| M2: TUI 面板可用 | 可在终端内查看布局和事件 |
| M3: 功能完整 | 5 大调试能力全部实现 |
| M4: Web 版本 | 可通过浏览器调试 |
| M5: 生产就绪 | 性能优化、文档完善 |

---

## 八、文件结构

```
runtime/devtools/
├── devtools.go          # 主入口
├── collector.go         # 收集器接口
├── protocol.go          # 调试协议
├── transport.go         # 传输层
│
├── layout.go            # 布局检查器
├── repaint.go           # 重绘调试
├── trace.go             # 事件追踪
├── focus.go             # 焦点检查器
├── tree.go              # 组件树
├── metrics.go           # 性能指标
│
├── hook.go              # Hook 系统
├── overlay.go           # 调试覆盖层绘制
│
└── panel/               # TUI 面板
    ├── panel.go
    ├── layout_view.go
    ├── repaint_view.go
    ├── event_view.go
    └── focus_view.go

devtools/web/            # Web Dashboard
├── server.go            # WebSocket 服务
├── index.html
├── app.js
└── styles.css

devtools/vscode/         # VSCode 扩展
├── extension.go
└── protocol.ts
```

---

## 附录

### A. 性能考虑

| 功能 | 开销 | 优化方案 |
|------|------|----------|
| 布局快照 | O(n) | 限制频率、增量更新 |
| 事件追踪 | O(1) | 过滤、环形缓冲 |
| 重绘追踪 | O(1) | 采样统计 |
| 组件树 | O(n) | 懒加载、虚拟化 |

### B. 安全考虑

- 生产环境完全禁用（编译标签）
- 数据脱敏（敏感信息过滤）
- 访问控制（本地连接）

### C. 扩展性

- 插件机制（自定义 Collector）
- 远程调试（WebSocket）
- 第三方客户端（协议开放）
