# TUI DevTools 实施文档 V2.0

> **项目**: Mint TUI Runtime
> **文档版本**: 2.0 (根据架构审查更新)
> **创建日期**: 2025-01-30
> **状态**: 设计阶段

> **V2 更新说明**: 根据架构审查文档，从"快照式调试器"升级为"增量时序调试引擎"

---

## 目录

1. [概述](#一概述)
2. [架构审查总结](#二架构审查总结)
3. [核心架构转变](#三核心架构转变)
4. [增量数据收集](#四增量数据收集)
5. [异步调试架构](#五异步调试架构)
6. [因果链引擎](#六因果链引擎)
7. [时间回溯系统](#七时间回溯系统)
8. [确定性回放](#八确定性回放)
9. [实施路线图](#九实施路线图)

---

## 一、概述

### 1.1 目标演进

从 V1.0 的"调试工具"升级为：

> **UI Runtime Observability + Causality Engine**

这不是"看发生了什么"，而是：
> **解释"为什么会发生"**

### 1.2 设计原则（V2 更新）

| 原则 | V1.0 | V2.0 |
|------|------|------|
| 数据收集 | 全量快照 | **增量 Delta** |
| 处理方式 | 同步观察 | **异步处理** |
| 性能影响 | O(n) 每帧 | **O(changed) 每帧** |
| 耦合方式 | 深度集成 | **零侵入 Hook** |
| 可扩展性 | 调试面板 | **因果链 + 回放 + AI** |

### 1.3 最终能力矩阵

```
┌─────────────────────────────────────────────────────────────┐
│                   TUI DevTools 能力矩阵                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  L1: 观察能力    Layout Inspector / Repaint Debug            │
│  L2: 追踪能力    Event Trace / Focus Inspector               │
│  L3: 理解能力    Causal Graph Engine                        │
│  L4: 回溯能力    Time Travel System                          │
│  L5: 复现能力    Deterministic Replay                       │
│  L6: 智能能力    Behavior Intelligence (未来)                │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## 九、实施路线图

### 9.1 分阶段实施

```
阶段 1: 增量基础 (P0)
  ├─ Layout Delta Collector
  ├─ Mutation Tap (Lock-Free Ring Buffer)
  ├─ 异步事件总线
  └─ 独立 Overlay Buffer

阶段 2: 因果链 (P1)
  ├─ Event → Mutation 关联
  ├─ Mutation → Layout 关联
  ├─ Layout → Repaint 关联
  └─ FrameTimeline 模型

阶段 3: 时间旅行 (P2)
  ├─ 稀疏快照系统
  ├─ 重放引擎
  └─ 时间轴 UI

阶段 4: 确定性回放 (P2)
  ├─ Input Recorder
  ├─ Replay 模式
  └─ 回放比对

阶段 5: 客户端 (P3)
  ├─ TUI Panel
  ├─ Web Dashboard
  └─ 协议优化

阶段 6: 高级功能 (未来)
  ├─ 性能分析 AI
  ├─ 自动优化建议
  └─ 代码重写
```

### 9.2 文件结构

```
runtime/devtools/
├── devtools.go              # 主入口
├── types.go                  # 核心类型定义
│
├── bus.go                    # 异步事件总线
├── tap.go                    # Lock-Free Ring Buffer Tap
│
├── delta/
│   ├── layout.go             # 布局增量收集
│   ├── repaint.go            # 重绘增量收集
│   └── event.go              # 事件增量收集
│
├── causal/
│   ├── graph.go              # 因果图
│   ├── builder.go            # 构建器
│   └── mutation.go            # 变更节点
│
├── timetravel/
│   ├── store.go              # 时间旅行存储
│   ├── snapshot.go           # 快照系统
│   └── replay.go             # 重放引擎
│
├── replay/
│   ├── input.go              # 输入记录
│   ├── replay.go             # 回放系统
│   └── compare.go            # 差异对比
│
├── overlay/
│   ├── buffer.go             # 独立覆盖层 Buffer
│   └── compose.go            # 合成输出
│
├── protocol/
│   ├── message.go            # 消息格式
│   ├── encode.go             # 编码器
│   └── decode.go             # 解码器
│
└── client/
    ├── tui/
    │   └── panel.go           # TUI 调试面板
    └── web/
        ├── server.go         # WebSocket 服务
        └── frontend/          # 前端资源
```
---

## 二、架构审查总结

### 2.1 一级风险（P0 - 必须优先改）

#### ❗ 问题 1：每帧全量世界快照

**问题代码**:
```go
snapshot := &LayoutSnapshot{
    Boxes: lc.extractBoxInfo(boxes),  // 深拷贝所有 box
    Tree:  lc.buildTree(root, boxes),   // 递归构建树
}
```

**问题**:
| 问题 | 后果 |
|------|------|
| GC 暴涨 | 帧率不稳 |
| CPU 抢占 | Debug 开启程序卡顿 |
| 大型 TUI 直接崩溃 | 不可用 |

**解决方案 → 增量 Delta 模型**:

```go
type LayoutDelta struct {
    FrameID  int

    // 增量变化
    Added    []NodeID      // 新增节点
    Removed  []NodeID      // 删除节点
    Changed  []NodeDelta   // 变化节点
}

type NodeDelta struct {
    ID   NodeID
    Mask ChangeMask       // 位掩码
    Rect *Rect            // 只存变化字段
}
```

#### ❗ 问题 2：Hook 层是同步观察者

**问题代码**:
```go
root.Paint(buf)      // 主渲染线程
lc.Collect(...)      // 同步做复杂逻辑
rc.Collect(...)
```

**解决方案 → 异步事件总线**:

```go
// 渲染线程：只记录
debugBus.Emit(DebugEvent{Type: EventLayout, Data: ...})

// 调试线程：异步处理
for ev := range debugBus {
    collector.Process(ev)
}
```

#### ❗ 问题 3：LayoutCollector 复制布局引擎

**问题代码**:
```go
buildTree()           // 又跑了一套布局
calculateMetrics()
getDepth()
```

**解决方案 → Runtime 暴露调试视图**:

```go
// Runtime 侧
type LayoutDebugView interface {
    ForEachBox(func(BoxDebugInfo))
}

// DevTools 只消费
runtime.ForEachBox(func(info BoxDebugInfo) {
    collector.Record(info)
})
```

### 2.2 二级问题（P1）

| 问题 | 解决方案 |
|------|----------|
| DebugOverlay 污染渲染模型 | 独立 Overlay Buffer |
| Event Trace 缺时间线 | FrameTimeline 模型 |
| 协议层缺流控 | 背压机制 |

---

## 三、核心架构转变

### 3.1 从 Snapshot 到 Delta

```
V1 (Snapshot Model):
  Frame N:  复制完整 UI 世界树
  Frame N+1: 再复制一棵
  → O(n * frames) 内存, GC 疯狂

V2 (Delta Model):
  Frame N:  记录哪些节点变了
  Frame N+1: 只记录新增的变化
  → O(changed nodes) 内存, GC 友好
```

### 3.2 架构对比

```
┌─────────────────────────────────────────────────────────────┐
│                       V1 架构 (Snapshot)                     │
├─────────────────────────────────────────────────────────────┤
│  Runtime                                                    │
│    ↓                                                         │
│  深拷贝完整状态                                              │
│    ↓                                                         │
│  构建调试数据结构                                            │
│    ↓                                                         │
│  DevTools                                                   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                       V2 架构 (Delta)                        │
├─────────────────────────────────────────────────────────────┤
│  Runtime                                                    │
│    ↓ Emit轻量事件                                           │
│  Lock-Free Ring Buffer                                     │
│    ↓                                                         │
│  Debug Goroutine (异步处理)                                  │
│    ↓                                                         │
│  增量数据流                                                  │
│    ↓                                                         │
│  DevTools                                                   │
└─────────────────────────────────────────────────────────────┘
```

### 3.3 核心数据结构转变

```go
// ===== V1: Snapshot =====
type LayoutSnapshot struct {
    Boxes     []BoxInfo       // 全量
    Tree      *TreeNode       // 全量
    Metrics   LayoutMetrics   // 全量计算
}

// ===== V2: Delta =====
type LayoutDelta struct {
    FrameID   int

    // 增量
    Added     []NodeID
    Removed   []NodeID
    Changed   []NodeDelta     // 只存变化

    // 统计（可选）
    Metrics   *LayoutMetrics  // 懒计算
}

type NodeDelta struct {
    ID      NodeID
    Mask    ChangeMask       // 位掩码标记哪些字段变了

    // 只填充变化字段
    Rect    *Rect
    ZIndex  *int
    Visible *bool
    Props   map[string]any   // 只包含变化的 props
}

type ChangeMask uint8

const (
    ChangeRect      ChangeMask = 1 << iota
    ChangeZ
    ChangeVisibility
    ChangeFlex
    ChangeProps
)
```

---

## 四、增量数据收集

### 4.1 Layout Delta Collector

```go
// runtime/devtools/layout_delta.go

package devtools

import (
    "sync"
    "sync/atomic"

    "github.com/wwsheng009/mint/runtime"
)

// LayoutCollector 增量布局收集器
type LayoutCollector struct {
    mu              sync.RWMutex
    enabled         uint32  // atomic, 用于快速禁用检查

    // 增量追踪
    lastVersion     map[NodeID]uint32
    nodeRegistry    map[NodeID]*runtime.LayoutNode

    // 输出流
    deltaChan       chan<- *LayoutDelta
}

// NewLayoutCollector 创建布局收集器
func NewLayoutCollector(deltaChan chan<- *LayoutDelta) *LayoutCollector {
    return &LayoutCollector{
        enabled:      0,
        lastVersion:  make(map[NodeID]uint32),
        nodeRegistry: make(map[NodeID]*runtime.LayoutNode),
        deltaChan:    deltaChan,
    }
}

// Enable 启用收集
func (lc *LayoutCollector) Enable() {
    atomic.StoreUint32(&lc.enabled, 1)
}

// Disable 禁用收集
func (lc *LayoutCollector) Disable() {
    atomic.StoreUint32(&lc.enabled, 0)
}

// IsEnabled 快速检查是否启用（无锁）
func (lc *LayoutCollector) IsEnabled() bool {
    return atomic.LoadUint32(&lc.enabled) != 0
}

// Collect 收集增量布局数据（每帧调用）
func (lc *LayoutCollector) Collect(boxes []runtime.LayoutBox) {
    // 快速路径：未启用直接返回
    if !lc.IsEnabled() {
        return
    }

    delta := &LayoutDelta{
        FrameID: lc.currentFrame(),
    }

    lc.mu.Lock()
    defer lc.mu.Unlock()

    // 检测新增和变化节点
    currentNodes := make(map[NodeID]bool)

    for _, box := range boxes {
        if box.Node == nil {
            continue
        }

        nodeID := NodeID(box.NodeID)
        currentNodes[nodeID] = true

        lastVersion, exists := lc.lastVersion[nodeID]

        // 新增节点
        if !exists {
            delta.Added = append(delta.Added, nodeID)
            lc.lastVersion[nodeID] = box.Node.LayoutVersion
            lc.nodeRegistry[nodeID] = box.Node
            continue
        }

        // 变化检测
        if lastVersion != box.Node.LayoutVersion {
            nodeDelta := lc.buildNodeDelta(box.Node, lastVersion)
            if nodeDelta != nil {
                delta.Changed = append(delta.Changed, *nodeDelta)
            }
            lc.lastVersion[nodeID] = box.Node.LayoutVersion
        }
    }

    // 检测删除节点
    for nodeID := range lc.lastVersion {
        if !currentNodes[nodeID] {
            delta.Removed = append(delta.Removed, nodeID)
            delete(lc.lastVersion, nodeID)
            delete(lc.nodeRegistry, nodeID)
        }
    }

    // 只在有变化时发送
    if len(delta.Added) > 0 || len(delta.Removed) > 0 || len(delta.Changed) > 0 {
        select {
        case lc.deltaChan <- delta:
        default:
            // 背压：丢弃此帧数据
        }
    }
}

// buildNodeDelta 构建节点增量
func (lc *LayoutCollector) buildNodeDelta(node *runtime.LayoutNode, lastVersion uint32) *NodeDelta {
    delta := &NodeDelta{
        ID:   NodeID(node.ID),
        Mask: 0,
    }

    // 检测位置变化
    oldRect, exists := lc.getNodeRect(node.ID)
    newRect := runtime.Rect{
        X: node.X,
        Y: node.Y,
        W: node.MeasuredWidth,
        H: node.MeasuredHeight,
    }

    if !exists || oldRect != newRect {
        delta.Rect = &newRect
        delta.Mask |= ChangeRect
    }

    // 检测 Z-Index 变化
    if node.Style.ZIndex != lc.getNodeZIndex(node.ID) {
        z := node.Style.ZIndex
        delta.ZIndex = &z
        delta.Mask |= ChangeZ
    }

    // 检测可见性变化
    if node.Style.Visible != lc.getNodeVisible(node.ID) {
        v := node.Style.Visible
        delta.Visible = &v
        delta.Mask |= ChangeVisibility
    }

    // 如果没有任何变化，返回 nil
    if delta.Mask == 0 {
        return nil
    }

    return delta
}

// 辅助方法：获取缓存的节点属性
func (lc *LayoutCollector) getNodeRect(id string) (runtime.Rect, bool) {
    if node, ok := lc.nodeRegistry[NodeID(id)]; ok {
        return runtime.Rect{
            X: node.X,
            Y: node.Y,
            W: node.MeasuredWidth,
            H: node.MeasuredHeight,
        }, true
    }
    return runtime.Rect{}, false
}

func (lc *LayoutCollector) getNodeZIndex(id string) int {
    if node, ok := lc.nodeRegistry[NodeID(id)]; ok {
        return node.Style.ZIndex
    }
    return 0
}

func (lc *LayoutCollector) getNodeVisible(id string) bool {
    if node, ok := lc.nodeRegistry[NodeID(id)]; ok {
        return node.Style.Visible
    }
    return true
}

func (lc *LayoutCollector) currentFrame() int {
    // 从帧计数器获取
    return 0
}
```

### 4.2 Runtime 侧修改：添加 LayoutVersion

```go
// runtime/node.go

type LayoutNode struct {
    ID string

    // ... 现有字段

    // 调试支持：布局版本号
    LayoutVersion uint32
}

// InvalidateLayout 使布局失效
func (n *LayoutNode) InvalidateLayout() {
    n.LayoutVersion++
    // ... 其他失效逻辑
}

// SetPosition 设置位置（自动递增版本）
func (n *LayoutNode) SetPosition(x, y int) {
    if n.X != x || n.Y != y {
        n.X = x
        n.Y = y
        n.LayoutVersion++
    }
}
```

### 4.3 Repaint Delta（重绘增量）

```go
// runtime/devtools/repaint_delta.go

package devtools

type RepaintDelta struct {
    FrameID      int

    // 脏区域（增量）
    DirtyRegions []paint.Rect

    // 统计
    ChangedCells int
    TotalCells   int
}

type RepaintCollector struct {
    enabled    uint32
    deltaChan  chan<- *RepaintDelta
}

// Collect 收集重绘数据
func (rc *RepaintCollector) Collect(diff *paint.DiffResult) {
    if !atomic.LoadUint32(&rc.enabled) != 0 {
        return
    }

    if !diff.HasChanges {
        return
    }

    delta := &RepaintDelta{
        DirtyRegions: diff.DirtyRegions,
        ChangedCells: diff.ChangedCells,
    }

    select {
    case rc.deltaChan <- delta:
    default:
        // 背压：丢弃
    }
}
```

### 4.4 Event Delta（事件增量）

```go
// runtime/devtools/event_delta.go

package devtools

import (
    "github.com/wwsheng009/mint/runtime/event"
)

type EventDelta struct {
    FrameID int

    // 此帧的事件
    Events []EventEntry

    // 因果关联
    CausedMutations []MutationID  // 此事件触发的状态变更
}

type EventEntry struct {
    Type    event.EventType
    Target  string
    Phase   event.EventPhase
    Time    time.Time
}

// 集成到事件分发
func recordEvent(ev *event.EventStruct) {
    // 极轻量记录，不做复杂处理
    tap.Emit(EventEntry{
        Type:   ev.Type(),
        Target: getTargetID(ev),
        Phase:  ev.Phase(),
    })
}
```

---

## 五、异步调试架构

### 5.1 核心设计原则

> **Render Thread 只做记录 → Debug Goroutine 做分析**

### 5.2 调试事件总线

```go
// runtime/devtools/bus.go

package devtools

import (
    "sync/atomic"
)

// DebugEvent 调试事件（极轻量）
type DebugEvent struct {
    Type   DebugEventType
    Data   uintptr  // 或指向预分配内存的指针
}

type DebugEventType uint8

const (
    EventLayout   DebugEventType = iota
    EventRepaint
    EventInput
    EventMutation
    EventFocus
)

// EventBus 调试事件总线
type EventBus struct {
    enabled     uint32
    writePos    uint32
    buffer      []DebugEvent
    mask        uint32

    // 订阅者
    subscribers []chan<- DebugEvent
}

// NewEventBus 创建事件总线
func NewEventBus(size int) *EventBus {
    // size 必须是 2 的幂
    if size & (size - 1) != 0 {
        size = nextPowerOfTwo(size)
    }

    return &EventBus{
        buffer: make([]DebugEvent, size),
        mask:  uint32(size - 1),
    }
}

// Emit 发送事件（无锁，极快）
func (b *EventBus) Emit(ev DebugEvent) {
    if atomic.LoadUint32(&b.enabled) == 0 {
        return
    }

    // 原子操作获取位置并写入
    pos := atomic.AddUint32(&b.writePos, 1)
    b.buffer[pos & b.mask] = ev
}

// Subscribe 订阅事件
func (b *EventBus) Subscribe(ch chan<- DebugEvent) {
    b.subscribers = append(b.subscribers, ch)
    go b.dispatchLoop(ch)
}

// dispatchLoop 分发循环（每个订阅者一个 goroutine）
func (b *EventBus) dispatchLoop(ch chan<- DebugEvent) {
    readPos := uint32(0)

    for {
        // 等待新事件
        for {
            writePos := atomic.LoadUint32(&b.writePos)
            if readPos < writePos {
                break
            }
            time.Sleep(10 * time.Millisecond)
        }

        // 读取事件
        ev := b.buffer[readPos & b.mask]
        readPos++

        // 发送给订阅者
        select {
        case ch <- ev:
        default:
            // 背压：跳过
        }
    }
}

// Enable 启用总线
func (b *EventBus) Enable() {
    atomic.StoreUint32(&b.enabled, 1)
}

// Disable 禁用总线
func (b *EventBus) Disable() {
    atomic.StoreUint32(&b.enabled, 0)
}
```

### 5.3 异步收集器架构

```go
// runtime/devtools/async_collector.go

package devtools

import (
    "time"
)

// AsyncCollector 异步收集器
type AsyncCollector struct {
    eventBus      *EventBus
    layoutDeltaCh chan *LayoutDelta
    repaintDeltaCh chan *RepaintDelta
    eventDeltaCh  chan *EventDelta

    // 输出
    outputCh      chan<- *DebugMessage
}

// Start 启动异步处理
func (ac *AsyncCollector) Start() {
    // 启动事件总线
    ac.eventBus.Enable()

    // 启动处理 goroutines
    go ac.processLayoutDeltas()
    go ac.processRepaintDeltas()
    go ac.processEventDeltas()
    go ac.processFrameTimeline()
}

// processFrameTimeline 处理帧时间线
func (ac *AsyncCollector) processFrameTimeline() {
    ticker := time.NewTicker(16 * time.Millisecond) // 60fps
    defer ticker.Stop()

    timeline := &FrameTimeline{
        FrameID:   0,
        StartTime: time.Now(),
    }

    for {
        select {
        case <-ticker.C:
            // 完成一帧
            ac.outputCh <- &DebugMessage{
                Type: MsgFrameTimeline,
                Payload: timeline,
            }

            // 开始新帧
            timeline = &FrameTimeline{
                FrameID:   timeline.FrameID + 1,
                StartTime: time.Now(),
            }
        }
    }
}
```

### 5.4 独立 Overlay Buffer

```go
// runtime/devtools/overlay.go

package devtools

import (
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
)

// Overlay 独立的调试覆盖层
type Overlay struct {
    buffer *paint.Buffer
    shown  map[string]bool  // 哪些组件需要高亮
}

// NewOverlay 创建覆盖层
func NewOverlay(width, height int) *Overlay {
    return &Overlay{
        buffer: paint.NewBuffer(width, height),
        shown:  make(map[string]bool),
    }
}

// Highlight 高亮组件
func (o *Overlay) Highlight(id string, rect paint.Rect, s style.Style) {
    o.shown[id] = true
    o.drawBoxBorder(rect, s)
}

// Clear 清除覆盖
func (o *Overlay) Clear() {
    for id := range o.shown {
        delete(o.shown, id)
    }
    // 清空 buffer
    o.buffer = paint.NewBuffer(o.buffer.Width, o.buffer.Height)
}

// Compose 合成到主输出
func (o *Overlay) Compose(mainOutput string) string {
    if len(o.shown) == 0 {
        return mainOutput
    }

    // 将 overlay 内容追加到主输出
    overlayOutput := o.buffer.String() // 或使用 diff
    return mainOutput + overlayOutput
}

// drawBoxBorder 绘制调试边框
func (o *Overlay) drawBoxBorder(rect paint.Rect, s style.Style) {
    // 上边
    for x := rect.X; x < rect.X+rect.W; x++ {
        o.buffer.SetCell(x, rect.Y, '─', s)
    }
    // 下边
    for x := rect.X; x < rect.X+rect.W; x++ {
        o.buffer.SetCell(x, rect.Y+rect.H-1, '─', s)
    }
    // 左边
    for y := rect.Y; y < rect.Y+rect.H; y++ {
        o.buffer.SetCell(rect.X, y, '│', s)
    }
    // 右边
    for y := rect.Y; y < rect.Y+rect.H; y++ {
        o.buffer.SetCell(rect.X+rect.W-1, y, '│', s)
    }

    // 角落
    o.buffer.SetCell(rect.X, rect.Y, '┌', s)
    o.buffer.SetCell(rect.X+rect.W-1, rect.Y, '┐', s)
    o.buffer.SetCell(rect.X, rect.Y+rect.H-1, '└', s)
    o.buffer.SetCell(rect.X+rect.W-1, rect.Y+rect.H-1, '┘', s)
}
```

---

## 六、因果链引擎

### 6.1 核心思想

从"事件日志"升级为"事件因果链":

```
Event  →  State Change  →  Layout Delta  →  Repaint
  ↓            ↓                ↓              ↓
形成因果链，可以回答"为什么这个按钮闪了一下？"
```

### 6.2 数据结构

```go
// runtime/devtools/causal.go

package devtools

// FrameRecord 帧记录
type FrameRecord struct {
    FrameID int
    Time    time.Time

    // 输入
    Events   []*EventNode

    // 中间状态变化
    Mutations []*MutationNode

    // 输出
    LayoutDelta  *LayoutDelta
    RepaintDelta *RepaintDelta
}

// EventNode 事件节点
type EventNode struct {
    ID       uint64
    Type     EventType
    TargetID string
    Phase    EventPhase
}

// MutationNode 状态变更节点
type MutationNode struct {
    ID        uint64
    Component string
    Kind      MutationKind
    Field     string
    OldValue  any
    NewValue  any
}

type MutationKind int

const (
    MutationState MutationKind = iota
    MutationProp
    MutationStyle
    MutationFocus
)

// Edge 因果边
type Edge struct {
    From uint64
    To   uint64
    Type EdgeType
}

type EdgeType int

const (
    EdgeEventToMutation EdgeType = iota
    EdgeMutationToLayout
    EdgeLayoutToRepaint
)
```

### 6.3 Mutation 捕获机制（零侵入）

```go
// runtime/devtools/mutation_tap.go

package devtools

import (
    "sync/atomic"
)

// MutationRecord 变更记录（极轻量）
type MutationRecord struct {
    ComponentID uint32
    FieldID     uint16
    Kind        uint8
}

// mutationTap 全局变更 Tap
var mutationTap = struct {
    enabled uint32
    writePos uint32
    buffer   []MutationRecord
    mask     uint32
}{
    buffer: make([]MutationRecord, 1<<14), // 16K ring
    mask:   (1 << 14) - 1,
}

// recordMutation 记录变更（极快，无分配）
func recordMutation(compID uint32, field uint16, kind uint8) {
    if atomic.LoadUint32(&mutationTap.enabled) == 0 {
        return
    }

    i := atomic.AddUint32(&mutationTap.writePos, 1)
    mutationTap.buffer[i & mutationTap.mask] = MutationRecord{
        ComponentID: compID,
        FieldID:     field,
        Kind:        kind,
    }
}

// Enable 启用捕获
func EnableMutationTap() {
    atomic.StoreUint32(&mutationTap.enabled, 1)
}

// Disable 禁用捕获
func DisableMutationTap() {
    atomic.StoreUint32(&mutationTap.enabled, 0)
}

// PollMutations 消费变更记录（调试线程调用）
func PollMutations(fromPos *uint32) []MutationRecord {
    currentPos := atomic.LoadUint32(&mutationTap.writePos)
    if *fromPos >= currentPos {
        return nil
    }

    var result []MutationRecord
    for *fromPos < currentPos {
        rec := mutationTap.buffer[*fromPos & mutationTap.mask]
        result = append(result, rec)
        *fromPos++
    }

    return result
}
```

### 6.4 Runtime 侧集成

```go
// runtime/component.go

// Component 组件基类
type Component struct {
    // ... 现有字段

    // 调试 ID（预注册，避免字符串）
    debugID uint32
}

// SetState 设置状态（带自动调试）
func (c *Component) SetState(key string, value any) {
    old := c.state[key]

    // 记录变更（如果 DevTools 启用）
    fieldID := stateFieldID(key)  // 预计算的 field ID
    devtools.recordMutation(c.debugID, fieldID, devtools.MutationState)

    c.state[key] = value

    if old != value {
        c.InvalidateLayout()
    }
}

// 预计算的 State Field ID（在初始化时分配）
func stateFieldID(key string) uint16 {
    // 从全局注册表获取
    return fieldIDRegistry.Get(key)
}
```

### 6.5 因果链构建

```go
// runtime/devtools/causal_builder.go

package devtools

// CausalBuilder 因果链构建器
type CausalBuilder struct {
    currentFrame *FrameRecord
    eventIndex   map[uint64]int      // EventID -> EventNode index
    mutationIndex map[uint64]int      // MutationID -> MutationNode index
}

// NewCausalBuilder 创建构建器
func NewCausalBuilder() *CausalBuilder {
    return &CausalBuilder{
        eventIndex:   make(map[uint64]int),
        mutationIndex: make(map[uint64]int),
    }
}

// BeginFrame 开始新帧
func (cb *CausalBuilder) BeginFrame(frameID int) {
    cb.currentFrame = &FrameRecord{
        FrameID: frameID,
        Time:    time.Now(),
    }
}

// AddEvent 添加事件
func (cb *CausalBuilder) AddEvent(ev EventEntry) uint64 {
    node := &EventNode{
        ID:      nextEventID(),
        Type:    ev.Type,
        TargetID: ev.Target,
        Phase:   ev.Phase,
    }
    cb.currentFrame.Events = append(cb.currentFrame.Events, node)
    cb.eventIndex[node.ID] = len(cb.currentFrame.Events) - 1
    return node.ID
}

// AddMutation 添加状态变更
func (cb *CausalBuilder) AddMutation(rec MutationRecord, causedBy uint64) uint64 {
    node := &MutationNode{
        ID:     nextMutationID(),
        Kind:   MutationKind(rec.Kind),
        Field:  fieldNameFromID(rec.FieldID),
    }

    cb.currentFrame.Mutations = append(cb.currentFrame.Mutations, node)
    cb.mutationIndex[node.ID] = len(cb.currentFrame.Mutations) - 1

    // 建立因果边
    if causedBy != 0 {
        // 从事件到变更的边
        cb.addEdge(Edge{
            From: causedBy,
            To:   node.ID,
            Type: EdgeEventToMutation,
        })
    }

    return node.ID
}

// Build 构建因果链
func (cb *CausalBuilder) Build() *FrameRecord {
    // 处理所有变更到布局的边
    cb.linkMutationsToLayout()

    // 处理布局到重绘的边
    cb.linkLayoutToRepaint()

    return cb.currentFrame
}
```

---

## 七、时间回溯系统

### 7.1 核心设计：Snapshot + Mutation Log

```
Frame 0     → Snapshot (完整状态)
Frame 1-119 → 只记录 Mutation
Frame 120   → Snapshot
```

### 7.2 数据结构

```go
// runtime/devtools/timetravel.go

package devtools

// TimeTravelStore 时间旅行存储
type TimeTravelStore struct {
    mu         sync.RWMutex

    // 稀疏快照
    snapshots map[int]*StateSnapshot

    // Mutation 日志
    mutations map[int][]MutationRecord

    // 快照间隔
    snapshotInterval int
}

// StateSnapshot 状态快照（稀疏）
type StateSnapshot struct {
    FrameID int

    // 只存关键数据
    ComponentStates map[CompID]StateBlob
    FocusState      FocusBlob
    LayoutVersion   map[NodeID]uint32
}

// StateBlob 状态 Blob（紧凑存储）
type StateBlob []uint64  // 按 fieldID 存储，避免 map
```

### 7.3 回溯算法

```go
// Rewind 回溯到指定帧
func (s *TimeTravelStore) Rewind(frameID int) (*StateSnapshot, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    // 1. 找到最近的快照
    snapshot, snapshotFrame := s.findNearestSnapshot(frameID)

    // 2. 从快照重放到目标帧
    return s.replay(snapshot, snapshotFrame, frameID)
}

// replay 重放
func (s *TimeTravelStore) replay(snapshot *StateSnapshot, from, to int) (*StateSnapshot, error) {
    // 设置回放模式
    // runtime.SetReplayMode(true)
    // defer runtime.SetReplayMode(false)

    state := snapshot.Copy()

    // 应用 mutation
    for f := from + 1; f <= to; f++ {
        mutations := s.mutations[f]
        for _, m := range mutations {
            state.Apply(m)
        }
    }

    return state, nil
}
```

---

## 八、确定性回放

### 8.1 Input Recorder

```go
// runtime/devtools/input_recorder.go

package devtools

// InputRecorder 输入记录器
type InputRecorder struct {
    enabled bool
    inputs  []InputEvent
}

// InputEvent 输入事件（确定性）
type InputEvent struct {
    FrameID   int
    Type      InputType
    Key       uint16
    MouseX    uint16
    MouseY    uint16
    Modifiers uint8
}

type InputType uint8

const (
    InputKeyPress InputType = iota
    InputKeyRelease
    InputMousePress
    InputMouseRelease
    InputMouseMotion
)

// Record 记录输入
func (r *InputRecorder) Record(ev InputEvent) {
    if !r.enabled {
        return
    }

    // 只记录确定性数据，不记录时间戳
    r.inputs = append(r.inputs, ev)
}
```

### 8.2 确定性前提

| 因素 | 处理方式 |
|------|----------|
| 时间 | 禁止用 `time.Now()`，改用 FrameTime |
| 随机数 | 固定种子 PRNG |
| 并发 | UI 主线程单线程模型 |
| IO | 不允许直接 IO 改 UI |

### 8.3 Replay 模式

```go
// Replay 重放输入
func (r *InputRecorder) Replay(runtime *Runtime, snapshot *StateSnapshot) error {
    // 设置回放模式
    runtime.SetReplayMode(true)
    defer runtime.SetReplayMode(false)

    // 加载快照
    runtime.LoadState(snapshot)

    // 重放输入
    frame := 0
    for _, input := range r.inputs {
        runtime.DispatchInput(input)
        runtime.Frame()
        frame++
    }

    return nil
}
```

---


### 9.3 性能目标

| 指标 | 目标 | 测量方法 |
|------|------|----------|
| Debug 关闭时开销 | < 0.1% | 分支预测成功 |
| Debug 开启时开销 | < 5% | 帧时间对比 |
| 内存占用 | < 10MB (1000 帧) | 堆分析 |
| GC 影响 | < 5% 额外 GC | GC 日志 |
| 端到端延迟 | < 16ms | 事件到显示 |

### 9.4 最终能力矩阵

```
┌────────────────────────────────────────────────────────────────┐
│                      DevTools 能力进化图                        │
├────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Phase 1: 增量观察                                             │
│    ✓ Layout Delta    - 只记录变化                              │
│    ✓ Repaint Delta   - 只记录脏区                              │
│    ✓ Event Delta     - 只记录事件                              │
│    ✓ 异步处理        - 不阻塞主循环                            │
│                                                                  │
│  Phase 2: 因果理解                                             │
│    ✓ Event → Mutation   - 事件触发了什么状态变化               │
│    ✓ Mutation → Layout  - 状态变化导致了什么布局变化             │
│    ✓ Layout → Repaint   - 布局变化导致了什么重绘                 │
│    ✓ 时间线可视化       - 一帧内的完整因果链                    │
│                                                                  │
│  Phase 3: 时间旅行                                             │
│    ✓ 稀疏快照          - 定期保存完整状态                        │
│    ✓ Mutation 日志     - 记录所有状态变更                       │
│    ✓ 状态重建          - 回到任意帧                             │
│    ✓ 时间轴导航        - 可视化时间线                            │
│                                                                  │
│  Phase 4: 确定性回放                                           │
│    ✓ 输入记录          - 记录所有用户输入                        │
│    ✓ 重放引擎          - 完全复现 UI 行为                       │
│    ✓ 回放对比          - 发现非确定性问题                       │
│    ✓ Bug 录制          - 保存和分享问题场景                      │
│                                                                  │
│  Phase 5: 智能分析 (未来)                                      │
│    ⏳ 性能热区          - 自动发现性能瓶颈                      │
│    ⏳ 无效刷新          - 自动发现多余重绘                      │
│    ⏳ 优化建议          - AI 驱动的代码优化                      │
│                                                                  │
└────────────────────────────────────────────────────────────────┘
```

---

## 附录

### A. 与 V1 的主要差异

| 方面 | V1 | V2 |
|------|----|----|
| 数据模型 | 全量快照 | 增量 Delta |
| 处理方式 | 同步 | 异步 |
| 性能影响 | O(n) 每帧 | O(changed) 每帧 |
| 可扩展性 | 调试面板 | 因果链 + 回放 |
| 工程级别 | 调试工具 | Observability Engine |

### B. 参考实现

- Chrome DevTools Protocol
- Flutter Observatory
- React DevTools (Fiber)
- Browser Performance Timeline

### C. 性能优化技巧

1. **使用位掩码** 代替多个 bool 字段
2. **使用预分配 ID** 代替字符串
3. **使用 Ring Buffer** 避免 GC
4. **使用原子操作** 避免锁
5. **批量处理** 减少上下文切换

### D. 调试检查清单

- [ ] DevTools 关闭时性能无影响
- [ ] Debug 开启时帧率稳定
- [ ] 大型 UI (1000+ 组件) 可用
- [ ] 长时间运行无内存泄漏
- [ ] 因果链完整无丢失
- [ ] 回放结果一致
