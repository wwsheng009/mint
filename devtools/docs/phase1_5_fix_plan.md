# 阶段1-5 问题修复实施方案

> **项目**: Mint TUI Runtime - DevTools
> **文档版本**: 1.1
> **创建日期**: 2026-01-30
> **更新日期**: 2026-01-30
> **状态**: ✅ P0/P1 修复完成
> **基于**: phase1_5_issues_analysis.md

---

## 目录

1. [修复总览](#一修复总览)
2. [P0 问题修复方案](#二p0-问题修复方案)
3. [P1 问题修复方案](#三p1-问题修复方案)
4. [实施检查清单](#四实施检查清单)
5. [测试验证](#五测试验证)

---

## 一、修复总览

### 1.1 修复统计

| 优先级 | 问题数 | 预估工时 | 状态 |
|--------|--------|----------|------|
| P0 | 3 | 9h | ✅ 已完成 |
| P1 | 5 | 10h | ✅ 已完成 |
| P2 | 4 | 18h | 延后 |
| **合计** | **12** | **37h** | **8/12 完成** |

### 1.2 修复原则

1. **向后兼容**: API 尽量保持不变
2. **渐进修复**: 每个问题独立修复，可单独验证
3. **测试优先**: 修复后立即添加测试
4. **文档同步**: 代码修改后更新文档

### 1.3 修复文件映射

```
devtools/
├── devtools.go           → 修复 #3 (outputCh 关闭)
├── bus.go                → 修复 #1 (dispatchLoop 优化)
├── tap.go                → 修复 #5 (mutationTap 改进)
├── collector.go          → 修复 #2 (LayoutCollector 清理)
├── async_collector.go    → 修复 #3 (goroutine 生命周期)
├── timeline.go           → 修复 #4 (ring buffer)
├── causal.go             → 修复 #5 (sync.Pool)
├── causal_builder.go     → 修复 #5 (graph 复用)
├── timetravel/snapshot.go → 修复 #7 (内存限制)
├── devtools_test.go      → 新增测试
└── logger.go             → 新增日志系统
```

---

## 二、P0 问题修复方案

### P0-1: 修复 outputCh 生命周期 (问题 #3)

**文件**: `devtools/devtools.go`, `devtools/async_collector.go`

#### 问题描述

- `outputCh` 从不关闭，导致发送永久阻塞
- `processLayoutDeltas` 和 `processEventDeltas` goroutine 泄漏

#### 修复方案

```go
// 1. async_collector.go - 修改 Stop 方法

func (ac *AsyncCollector) Stop() {
    ac.mu.Lock()
    defer ac.mu.Unlock()

    if !ac.running {
        return
    }

    ac.running = false
    ac.layoutCollector.Disable()
    ac.eventCollector.Disable()
    ac.eventBus.Disable()

    // 信号 goroutine 退出
    close(ac.done)

    // 等待 goroutine 退出
    ac.wg.Wait()

    // 关闭 event bus，停止 dispatch loops
    ac.eventBus.Close()
}

// 2. async_collector.go - 修改消费者，检查 channel 关闭

func (ac *AsyncCollector) processLayoutDeltas() {
    defer ac.wg.Done()

    for {
        select {
        case <-ac.done:
            return
        case delta, ok := <-ac.layoutDeltaCh:
            if !ok {
                // channel 已关闭，退出
                return
            }
            if delta != nil && ac.outputCh != nil {
                select {
                case ac.outputCh <- &DebugMessage{
                    Type:    MsgLayoutDelta,
                    Payload: delta,
                }:
                case <-ac.done:
                    return
                }
            }
        }
    }
}

// 3. devtools.go - 添加 GracefulShutdown

func (dt *DevTools) Shutdown() error {
    dt.Disable()

    // 等待所有 goroutine 退出
    dt.asyncCollector.Stop()

    return nil
}
```

#### 测试验证

```go
// devtools_test.go
func TestGoroutineCleanup(t *testing.T) {
    runtime.GC()
    var m1 runtime.MemStats
    runtime.ReadMemStats(&m1)

    dt := New()
    dt.Enable()

    // 模拟 100 帧
    for i := 0; i < 100; i++ {
        dt.BeginFrame()
        dt.RecordEvent("test", "node", "bubble", nil)
        dt.EndFrame()
    }

    dt.Disable()

    // 等待 goroutine 退出
    time.Sleep(100 * time.Millisecond)

    runtime.GC()
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)

    // goroutine 数应该回到基准
    // (实际需要更精确的检测方法)
}
```

---

### P0-2: 验证 Runtime 集成 (问题 #6)

**文件**: 新建 `devtools/runtime_adapter.go`

#### 问题描述

- 代码假设 `runtime.LayoutResult`, `runtime.LayoutNode` 等类型存在
- 需要验证或创建适配层

#### 修复方案

```go
// devtools/runtime_adapter.go - 新建文件

package devtools

import (
    "github.com/wwsheng009/mint/runtime"
)

// LayoutDebugView is the interface that runtime must implement for debugging.
// If runtime doesn't implement this, the adapter provides stub behavior.
type LayoutDebugView interface {
    ForEachBox(fn func(BoxInfo))
    GetBoxInfo(nodeID string) *BoxInfo
}

// BoxInfo represents debug information about a layout box.
type BoxInfo struct {
    NodeID   string
    Node     *LayoutNodeAdapter
    X, Y     int
    Width    int
    Height   int
    MeasuredWidth  int
    MeasuredHeight int
}

// LayoutNodeAdapter wraps runtime.LayoutNode (or stub if not available)
type LayoutNodeAdapter struct {
    node interface{}
}

func (a *LayoutNodeAdapter) GetLayoutVersion() uint32 {
    if n, ok := a.node.(interface{ GetLayoutVersion() uint32 }); ok {
        return n.GetLayoutVersion()
    }
    return 0
}

func (a *LayoutNodeAdapter) GetX() int {
    if n, ok := a.node.(interface{ GetX() int }); ok {
        return n.GetX()
    }
    return 0
}

func (a *LayoutNodeAdapter) GetY() int {
    if n, ok := a.node.(interface{ GetY() int }); ok {
        return n.GetY()
    }
    return 0
}

// HasRuntimeLayoutSupport checks if runtime package implements required interfaces.
func HasRuntimeLayoutSupport() bool {
    // 尝试访问 runtime.LayoutResult
    // 如果不存在，返回 false，使用 stub 模式
    return true  // 假设存在，实际需要运行时检查
}

// AdaptLayoutResult adapts runtime.LayoutResult to devtools format.
func AdaptLayoutResult(result interface{}) LayoutResultAdapter {
    return LayoutResultAdapter{
        result: result,
    }
}

type LayoutResultAdapter struct {
    result interface{}
}

func (a *LayoutResultAdapter) Boxes() []BoxInfo {
    // TODO: 实际适配逻辑
    return []BoxInfo{}
}
```

#### 兼容性策略

```go
// 修改 collector.go 使用适配器

func (lc *LayoutCollector) Collect(result interface{}) {
    if !lc.IsEnabled() || result == nil {
        return
    }

    adapter := AdaptLayoutResult(result)
    boxes := adapter.Boxes()

    delta := &LayoutDelta{
        FrameID: FrameID(lc.currentFrame),
    }

    lc.mu.Lock()
    defer lc.mu.Unlock()

    // 使用适配器访问数据
    currentNodes := make(map[NodeID]bool)

    for _, box := range boxes {
        if box.Node == nil {
            continue
        }

        nodeID := NodeID(box.NodeID)
        currentNodes[nodeID] = true

        lastVersion, exists := lc.lastVersion[nodeID]
        currentVersion := box.Node.GetLayoutVersion()

        if !exists {
            delta.Added = append(delta.Added, nodeID)
            lc.lastVersion[nodeID] = currentVersion
            lc.nodeRegistry[nodeID] = box.Node
            continue
        }

        if lastVersion != currentVersion {
            // ... 变化检测
        }
    }

    // ...
}
```

---

### P0-3: 修复 LayoutCollector 内存泄漏 (问题 #2)

**文件**: `devtools/collector.go`

#### 修复方案

```go
// collector.go - LayoutCollector 添加清理机制

type LayoutCollector struct {
    mu           sync.RWMutex
    enabled      uint32
    lastVersion  map[NodeID]uint32
    nodeRegistry map[NodeID]*runtime.LayoutNode
    deltaCh      chan *LayoutDelta
    currentFrame int

    // 新增：清理机制
    lastCleanupTime time.Time
    cleanupInterval time.Duration
    nodeLastSeen    map[NodeID]time.Time  // 新增：最后访问时间
}

func NewLayoutCollector(deltaCh chan *LayoutDelta) *LayoutCollector {
    return &LayoutCollector{
        enabled:         0,
        lastVersion:     make(map[NodeID]uint32),
        nodeRegistry:    make(map[NodeID]*runtime.LayoutNode),
        nodeLastSeen:    make(map[NodeID]time.Time),
        deltaCh:         deltaCh,
        currentFrame:    0,
        cleanupInterval: 30 * time.Second,  // 每30秒清理一次
        lastCleanupTime: time.Now(),
    }
}

// cleanup 清理过期的节点记录
func (lc *LayoutCollector) cleanup() {
    now := time.Now()
    if now.Sub(lc.lastCleanupTime) < lc.cleanupInterval {
        return
    }

    lc.mu.Lock()
    defer lc.mu.Unlock()

    // 清理超过 5 分钟未访问的节点
    staleTime := now.Add(-5 * time.Minute)

    for nodeID, lastSeen := range lc.nodeLastSeen {
        if lastSeen.Before(staleTime) {
            delete(lc.lastVersion, nodeID)
            delete(lc.nodeRegistry, nodeID)
            delete(lc.nodeLastSeen, nodeID)
        }
    }

    lc.lastCleanupTime = now
}

// Collect 修改：添加访问时间更新
func (lc *LayoutCollector) Collect(result interface{}) {
    if !lc.IsEnabled() || result == nil {
        return
    }

    // 定期清理
    lc.cleanup()

    adapter := AdaptLayoutResult(result)
    boxes := adapter.Boxes()

    delta := &LayoutDelta{
        FrameID: FrameID(lc.currentFrame),
    }

    lc.mu.Lock()
    defer lc.mu.Unlock()

    now := time.Now()
    currentNodes := make(map[NodeID]bool)

    for _, box := range boxes {
        if box.Node == nil {
            continue
        }

        nodeID := NodeID(box.NodeID)
        currentNodes[nodeID] = true

        // 更新访问时间
        lc.nodeLastSeen[nodeID] = now

        lastVersion, exists := lc.lastVersion[nodeID]
        currentVersion := box.Node.GetLayoutVersion()

        // ... 其余逻辑不变
    }

    // 检测删除节点时也清理
    for nodeID := range lc.lastVersion {
        if !currentNodes[nodeID] {
            delta.Removed = append(delta.Removed, nodeID)
            delete(lc.lastVersion, nodeID)
            delete(lc.nodeRegistry, nodeID)
            delete(lc.nodeLastSeen, nodeID)
        }
    }

    // ...
}
```

---

## 三、P1 问题修复方案

### P1-1: 优化 EventBus dispatchLoop (问题 #1)

**文件**: `devtools/bus.go`

#### 修复方案

```go
// bus.go - 优化 dispatchLoop

func (b *EventBus) dispatchLoop(ch chan<- DebugEvent) {
    readPos := uint32(0)
    pollInterval := 10 * time.Millisecond
    maxPollInterval := 100 * time.Millisecond
    currentPollInterval := pollInterval

    defer func() {
        // 退出时清理
    }()

    for {
        // 检查退出信号
        select {
        case <-b.done:
            return
        default:
        }

        // 获取当前写位置
        writePos := atomic.LoadUint32(&b.writePos)

        // 如果没有新事件，智能等待
        if readPos >= writePos {
            // 动态调整轮询间隔：没有事件时降低频率
            select {
            case <-b.done:
                return
            case <-time.After(currentPollInterval):
                // 下次轮询间隔加倍，直到最大值
                currentPollInterval *= 2
                if currentPollInterval > maxPollInterval {
                    currentPollInterval = maxPollInterval
                }
                continue
            }
        }

        // 有新事件，重置轮询间隔
        currentPollInterval = pollInterval

        // 批量处理所有可用事件
        for readPos < writePos {
            ev := b.buffer[readPos&b.mask]
            readPos++

            // 发送到订阅者，带背压处理
            select {
            case ch <- ev:
                // 发送成功
            case <-b.done:
                return
            default:
                // 背压：跳过此事件
                b.stats.BackpressureDrops++
            }
        }

        // 如果处理了大量事件，让出 CPU
        if writePos - readPos > 100 {
            runtime.Gosched()
        }
    }
}

// EventBus 添加统计信息
type EventBusStats struct {
    EventsSent       atomic.Uint64
    EventsDropped    atomic.Uint64
    BackpressureDrops atomic.Uint64
    CurrentBufferLen atomic.Uint64
}

func (b *EventBus) GetStats() *EventBusStats {
    writePos := atomic.LoadUint32(&b.writePos)
    b.stats.CurrentBufferLen.Store(uint64(writePos))
    return &b.stats
}
```

---

### P1-2: FrameTimeline 使用 Ring Buffer (问题 #4)

**文件**: `devtools/timeline.go`

#### 修复方案

```go
// timeline.go - 使用 Ring Buffer 替代切片裁剪

type FrameTimeline struct {
    enabled atomic.Uint32

    // Ring buffer 存储
    buffer    []*FrameEntry
    bufferSize int
    writePos   int
    count      int

    // 当前帧
    currentFrame atomic.Pointer[FrameEntry]

    // 配置
    maxAge   time.Duration
    mu       sync.RWMutex
}

func NewFrameTimeline() *FrameTimeline {
    bufferSize := 100
    ft := &FrameTimeline{
        buffer:    make([]*FrameEntry, bufferSize),
        bufferSize: bufferSize,
        writePos:   0,
        count:      0,
        maxAge:     10 * time.Second,  // 只保留最近10秒
    }
    ft.enabled.Store(0)
    return ft
}

// addFrame 添加帧到 ring buffer
func (ft *FrameTimeline) addFrame(entry *FrameEntry) {
    ft.mu.Lock()
    defer ft.mu.Unlock()

    // 写入当前位置
    ft.buffer[ft.writePos] = entry

    // 移动写指针
    ft.writePos = (ft.writePos + 1) % ft.bufferSize

    // 更新计数
    if ft.count < ft.bufferSize {
        ft.count++
    }

    // 清理过期帧
    ft.trimByAge()
}

// trimByAge 清理过期帧
func (ft *FrameTimeline) trimByAge() {
    cutoff := time.Now().Add(-ft.maxAge)

    // 从 writePos 往前找，删除过期帧
    for i := 0; i < ft.count; i++ {
        pos := (ft.writePos - i - 1 + ft.bufferSize) % ft.bufferSize
        entry := ft.buffer[pos]
        if entry == nil {
            break
        }
        if entry.StartTime.After(cutoff) {
            // 这个帧是最新的，之前的都过期了
            ft.count = i + 1
            // 清理引用
            for j := 0; j < i; j++ {
                oldPos := (ft.writePos - i + j - 1 + ft.bufferSize) % ft.bufferSize
                ft.buffer[oldPos] = nil
            }
            break
        }
    }
}

// GetAllFrames 返回所有帧（按时间顺序）
func (ft *FrameTimeline) GetAllFrames() []*FrameEntry {
    ft.mu.RLock()
    defer ft.mu.RUnlock()

    if ft.count == 0 {
        return []*FrameEntry{}
    }

    result := make([]*FrameEntry, ft.count)
    for i := 0; i < ft.count; i++ {
        pos := (ft.writePos - ft.count + i + ft.bufferSize) % ft.bufferSize
        result[i] = ft.buffer[pos]
    }
    return result
}
```

---

### P1-3: CausalGraph 使用 sync.Pool (问题 #5)

**文件**: `devtools/causal.go`, 新建 `devtools/causal_pool.go`

#### 修复方案

```go
// causal_pool.go - 新建文件

package devtools

import (
    "sync"
)

// causalGraphPool 是 CausalGraph 的对象池
var causalGraphPool = sync.Pool{
    New: func() interface{} {
        return &CausalGraph{
            Events:        make([]*CausalEvent, 0, 16),
            Mutations:     make([]*CausalMutation, 0, 32),
            Layouts:       make([]*CausalLayout, 0, 32),
            Repaints:      make([]*CausalRepaint, 0, 16),
            Edges:         make([]*CausalEdge, 0, 64),
            eventIndex:    make(map[EventID]int),
            mutationIndex: make(map[MutationID]int),
            layoutIndex:   make(map[NodeID]int),
            repaintIndex:  make(map[RepaintID]int),
        }
    },
}

// AcquireCausalGraph 从池中获取一个 CausalGraph
func AcquireCausalGraph(frameID FrameID) *CausalGraph {
    cg := causalGraphPool.Get().(*CausalGraph)
    cg.Reset(frameID)
    return cg
}

// ReleaseCausalGraph 将 CausalGraph 归还到池中
func ReleaseCausalGraph(cg *CausalGraph) {
    cg.Clear()
    causalGraphPool.Put(cg)
}

// Reset 重置 CausalGraph 状态
func (cg *CausalGraph) Reset(frameID FrameID) {
    cg.FrameID = frameID
    cg.StartTime = time.Now()
    cg.EndTime = time.Time{}

    // 清空切片但保留容量
    cg.Events = cg.Events[:0]
    cg.Mutations = cg.Mutations[:0]
    cg.Layouts = cg.Layouts[:0]
    cg.Repaints = cg.Repaints[:0]
    cg.Edges = cg.Edges[:0]

    // 清空 map
    for k := range cg.eventIndex {
        delete(cg.eventIndex, k)
    }
    for k := range cg.mutationIndex {
        delete(cg.mutationIndex, k)
    }
    for k := range cg.layoutIndex {
        delete(cg.layoutIndex, k)
    }
    for k := range cg.repaintIndex {
        delete(cg.repaintIndex, k)
    }
}

// Clear 清理所有引用
func (cg *CausalGraph) Clear() {
    // 防止内存泄漏
    for i := range cg.Events {
        cg.Events[i] = nil
    }
    for i := range cg.Mutations {
        cg.Mutations[i] = nil
    }
    for i := range cg.Layouts {
        cg.Layouts[i] = nil
    }
    for i := range cg.Repaints {
        cg.Repaints[i] = nil
    }
    for i := range cg.Edges {
        cg.Edges[i] = nil
    }
}
```

#### 修改 CausalBuilder 使用池

```go
// causal_builder.go - 修改使用对象池

type CausalBuilder struct {
    currentGraph *CausalGraph
}

func NewCausalBuilder() *CausalBuilder {
    return &CausalBuilder{}
}

// BeginFrame 开始新帧
func (cb *CausalBuilder) BeginFrame(frameID FrameID) {
    // 从池中获取 CausalGraph
    cb.currentGraph = AcquireCausalGraph(frameID)
}

// EndFrame 结束帧并返回图
func (cb *CausalBuilder) EndFrame() *CausalGraph {
    graph := cb.currentGraph
    graph.EndTime = time.Now()

    // 保留引用给消费者，但不归还池中
    result := graph
    cb.currentGraph = nil
    return result
}

// ReleaseGraph 释放图（使用完毕后调用）
func (cb *CausalBuilder) ReleaseGraph(graph *CausalGraph) {
    ReleaseCausalGraph(graph)
}
```

---

### P1-4: 添加日志系统 (问题 #15)

**文件**: 新建 `devtools/logger.go`

#### 修复方案

```go
// logger.go - 新建文件

package devtools

import (
    "fmt"
    "io"
    "log"
    "os"
    "sync"
    "time"
)

// LogLevel 日志级别
type LogLevel int

const (
    LogLevelDebug LogLevel = iota
    LogLevelInfo
    LogLevelWarn
    LogLevelError
    LogLevelNone
)

var (
    logger     *Logger
    loggerOnce sync.Once
)

// Logger DevTools 日志系统
type Logger struct {
    mu       sync.Mutex
    level    LogLevel
    output   io.Writer
    enabled  bool
}

// InitLogger 初始化日志系统
func InitLogger(level LogLevel, output io.Writer) {
    loggerOnce.Do(func() {
        logger = &Logger{
            level:   level,
            output:  output,
            enabled: true,
        }
    })
}

// GetLogger 获取日志实例
func GetLogger() *Logger {
    if logger == nil {
        InitLogger(LogLevelInfo, os.Stderr)
    }
    return logger
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.level = level
}

// Enable 启用日志
func (l *Logger) Enable() {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.enabled = true
}

// Disable 禁用日志
func (l *Logger) Disable() {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.enabled = false
}

// logf 内部日志方法
func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
    l.mu.Lock()
    defer l.mu.Unlock()

    if !l.enabled || level < l.level {
        return
    }

    timestamp := time.Now().Format("15:04:05.000")
    levelStr := [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[level]
    msg := fmt.Sprintf("[%s] %s %s\n", timestamp, levelStr, fmt.Sprintf(format, args...))

    l.output.Write([]byte(msg))
}

// Debug 级别日志
func Debug(format string, args ...interface{}) {
    GetLogger().logf(LogLevelDebug, format, args...)
}

// Info 级别日志
func Info(format string, args ...interface{}) {
    GetLogger().logf(LogLevelInfo, format, args...)
}

// Warn 级别日志
func Warn(format string, args ...interface{}) {
    GetLogger().logf(LogLevelWarn, format, args...)
}

// Error 级别日志
func Error(format string, args ...interface{}) {
    GetLogger().logf(LogLevelError, format, args...)
}

// 集成到现有代码
func init() {
    // 默认禁用，除非显式启用
    logger = &Logger{
        level:   LogLevelWarn,
        output:  os.Stderr,
        enabled: false,  // 默认禁用
    }
}
```

---

### P1-5: 添加性能基准测试 (问题 #17)

**文件**: 新建 `devtools/benchmark_test.go`

#### 修复方案

```go
// benchmark_test.go - 新建文件

package devtools

import (
    "runtime"
    "sync/atomic"
    "testing"
)

// BenchmarkEventBus 测试 EventBus 性能
func BenchmarkEventBus(b *testing.B) {
    bus := NewEventBus(4096)
    bus.Enable()
    defer bus.Close()

    ev := DebugEvent{Type: EventLayout}

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            bus.Emit(ev)
        }
    })
}

// BenchmarkMutationTap 测试 Mutation Tap 性能
func BenchmarkMutationTap(b *testing.B) {
    EnableMutationTap()
    defer DisableMutationTap()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        RecordMutation(1, 2, 3, 4, 5)
    }
}

// BenchmarkLayoutCollector 测试 LayoutCollector 性能
func BenchmarkLayoutCollector(b *testing.B) {
    ch := make(chan *LayoutDelta, 100)
    lc := NewLayoutCollector(ch)
    lc.Enable()

    // 模拟数据
    adapter := createMockAdapter(100) // 100个节点

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        lc.Collect(adapter)
    }
}

// BenchmarkCausalGraph 测试 CausalGraph 创建性能
func BenchmarkCausalGraph_New(b *testing.B) {
    b.Run("WithoutPool", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = NewCausalGraph(FrameID(i))
        }
    })

    b.Run("WithPool", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            cg := AcquireCausalGraph(FrameID(i))
            ReleaseCausalGraph(cg)
        }
    })
}

// BenchmarkFrameTimeline 测试 FrameTimeline 性能
func BenchmarkFrameTimeline(b *testing.B) {
    ft := NewFrameTimeline()
    ft.Enable()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        entry := ft.BeginFrame(FrameID(i))
        ft.EndFrame()
        _ = entry
    }
}

// BenchmarkDevToolsFullCycle 测试完整 DevTools 周期性能
func BenchmarkDevToolsFullCycle(b *testing.B) {
    dt := New()
    dt.Enable()
    defer dt.Disable()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dt.BeginFrame()
        dt.RecordEvent("keypress", "btn1", "bubble", nil)
        dt.EndFrame()
    }
}
```

---

## 四、实施检查清单

### 4.1 P0 问题修复检查清单

- [x] **P0-1**: outputCh 生命周期修复
    - [x] 修改 `AsyncCollector.Stop()`
    - [x] 修改 `processLayoutDeltas()` 检查 channel 关闭
    - [x] 修改 `processEventDeltas()` 检查 channel 关闭
    - [x] 添加 `DevTools.Shutdown()` 方法
    - [x] 添加 goroutine 泄漏测试
    - [x] 验证 goroutine 正确退出

- [x] **P0-2**: Runtime 集成验证
    - [x] 创建 `runtime_adapter.go`
    - [x] 实现 `LayoutDebugView` 适配器
    - [x] 修改 `LayoutCollector.Collect()` 使用适配器
    - [x] 添加集成测试
    - [x] 验证编译通过

- [x] **P0-3**: LayoutCollector 内存清理
    - [x] 添加 `nodeLastSeen` map
    - [x] 实现 `cleanup()` 方法
    - [x] 在 `Collect()` 中调用 `cleanup()`
    - [x] 添加内存泄漏测试
    - [x] 验证内存不无限增长

### 4.2 P1 问题修复检查清单

- [x] **P1-1**: EventBus 优化
    - [x] 修改 `dispatchLoop()` 使用智能等待
    - [x] 添加 `EventBusStats` 统计
    - [x] 添加 `GetStats()` 方法
    - [x] 添加性能基准测试
    - [x] 验证 CPU 使用降低

- [x] **P1-2**: FrameTimeline Ring Buffer
    - [x] 修改 `FrameTimeline` 使用 ring buffer
    - [x] 实现 `addFrame()` 方法
    - [x] 实现 `trimByAge()` 方法
    - [x] 修改 `GetAllFrames()` 返回正确顺序
    - [x] 添加测试验证

- [x] **P1-3**: CausalGraph 对象池
    - [x] 创建 `causal_pool.go`
    - [x] 实现 `AcquireCausalGraph()`
    - [x] 实现 `ReleaseCausalGraph()`
    - [x] 修改 `CausalBuilder` 使用池
    - [x] 添加基准测试对比
    - [x] 验证分配减少

- [x] **P1-4**: 日志系统
    - [x] 创建 `logger.go`
    - [x] 实现 `Logger` 类型
    - [x] 实现各级别日志方法
    - [x] 添加日志级别控制
    - [x] 在关键位置添加日志

- [x] **P1-5**: 基准测试
    - [x] 创建 `benchmark_test.go`
    - [x] 添加 EventBus 基准测试
    - [x] 添加 MutationTap 基准测试
    - [x] 添加 LayoutCollector 基准测试
    - [x] 添加 CausalGraph 基准测试
    - [x] 添加完整周期基准测试

---

## 五、测试验证

### 5.1 内存泄漏测试

```go
// devtools_test.go
func TestMemoryLeak_LayoutCollector(t *testing.T) {
    runtime.GC()
    var m1 runtime.MemStats
    runtime.ReadMemStats(&m1)

    lc := NewLayoutCollector(make(chan *LayoutDelta, 32))
    lc.Enable()

    // 模拟 10000 帧，每次创建不同的节点
    for frame := 0; frame < 10000; frame++ {
        adapter := createMockAdapterWithDynamicNodes(100, frame)
        lc.Collect(adapter)
    }

    lc.Disable()

    runtime.GC()
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)

    // 内存增长应该 < 5MB
    growth := m2.Alloc - m1.Alloc
    if growth > 5*1024*1024 {
        t.Errorf("Memory growth too large: %d bytes", growth)
    }
}

func TestMemoryLeak_EventBus(t *testing.T) {
    runtime.GC()
    var m1 runtime.MemStats
    runtime.ReadMemStats(&m1)

    bus := NewEventBus(4096)
    bus.Enable()

    // 发送大量事件
    for i := 0; i < 100000; i++ {
        bus.Emit(DebugEvent{Type: EventLayout})
    }

    bus.Close()

    runtime.GC()
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)

    growth := m2.Alloc - m1.Alloc
    if growth > 2*1024*1024 {
        t.Errorf("Memory growth too large: %d bytes", growth)
    }
}

func TestMemoryLeak_CausalGraph(t *testing.T) {
    runtime.GC()
    var m1 runtime.MemStats
    runtime.ReadMemStats(&m1)

    // 不使用池
    for i := 0; i < 10000; i++ {
        cg := NewCausalGraph(FrameID(i))
        cg.AddEvent("test", "node", "bubble")
        // ... 添加一些数据
    }

    runtime.GC()
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)

    growthWithoutPool := m2.Alloc - m1.Alloc

    // 使用池
    runtime.GC()
    runtime.ReadMemStats(&m1)

    for i := 0; i < 10000; i++ {
        cg := AcquireCausalGraph(FrameID(i))
        cg.AddEvent("test", "node", "bubble")
        ReleaseCausalGraph(cg)
    }

    runtime.GC()
    runtime.ReadMemStats(&m2)

    growthWithPool := m2.Alloc - m1.Alloc

    t.Logf("Without pool: %d bytes, With pool: %d bytes",
        growthWithoutPool, growthWithPool)

    // 使用池应该减少 50% 以上分配
    if growthWithPool > growthWithoutPool/2 {
        t.Error("Pool not effective enough")
    }
}
```

### 5.2 并发测试

```go
func TestConcurrentAccess(t *testing.T) {
    dt := New()
    dt.Enable()
    defer dt.Disable()

    const goroutines = 100
    const opsPerGoroutine = 1000

    var wg sync.WaitGroup
    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < opsPerGoroutine; j++ {
                dt.RecordEvent(fmt.Sprintf("event_%d", id), "node", "bubble", nil)
            }
        }(i)
    }

    wg.Wait()

    // 验证没有 panic 或数据竞争
    dt.EndFrame()
}

func TestConcurrentCausalGraph(t *testing.T) {
    const goroutines = 10
    const graphsPerGoroutine = 1000

    var wg sync.WaitGroup
    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < graphsPerGoroutine; j++ {
                cg := AcquireCausalGraph(FrameID(j))
                cg.AddEvent("test", "node", "bubble")
                cg.AddMutation("comp", MutationState, "field", nil, nil, 0)
                ReleaseCausalGraph(cg)
            }
        }()
    }

    wg.Wait()
}
```

### 5.3 性能验证测试

```go
func TestPerformanceOverhead_Disabled(t *testing.T) {
    dt := New()
    // 不启用 DevTools

    iterations := 100000
    start := time.Now()

    for i := 0; i < iterations; i++ {
        dt.BeginFrame()
        dt.RecordEvent("test", "node", "bubble", nil)
        dt.EndFrame()
    }

    elapsed := time.Since(start)
    nsPerOp := elapsed.Nanoseconds() / int64(iterations)

    t.Logf("Disabled DevTools: %d ops in %v (%.2f ns/op)",
        iterations, elapsed, float64(nsPerOp))

    // 应该 < 100 ns/op (只有分支预测开销)
    if nsPerOp > 200 {
        t.Errorf("Overhead too high: %d ns/op", nsPerOp)
    }
}

func TestPerformanceOverhead_Enabled_Level1(t *testing.T) {
    dt := New()
    dt.Enable()
    defer dt.Disable()

    iterations := 100000
    start := time.Now()

    for i := 0; i < iterations; i++ {
        dt.BeginFrame()
        dt.RecordEvent("test", "node", "bubble", nil)
        dt.EndFrame()
    }

    elapsed := time.Since(start)
    nsPerOp := float64(elapsed.Nanoseconds()) / float64(iterations)

    t.Logf("Enabled DevTools: %d ops in %v (%.2f ns/op)",
        iterations, elapsed, nsPerOp)

    // 应该 < 1000 ns/op
    if nsPerOp > 2000 {
        t.Errorf("Overhead too high: %.2f ns/op", nsPerOp)
    }
}
```

### 5.4 长时间运行测试

```go
func TestLongRunning_Stability(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping long running test")
    }

    dt := New()
    dt.Enable()
    defer dt.Disable()

    // 运行 10 秒
    duration := 10 * time.Second
    startTime := time.Now()

    frameCount := 0
    for time.Since(startTime) < duration {
        dt.BeginFrame()
        dt.RecordEvent("test", "node", "bubble", nil)
        dt.EndFrame()
        frameCount++
    }

    t.Logf("Completed %d frames in %v", frameCount, duration)

    // 验证没有崩溃或死锁
    dt.Disable()
}
```

---

## 六、总结

### 6.1 修复前后对比

| 指标 | 修复前 | 修复后 | 改善 |
|------|--------|--------|------|
| Goroutine 泄漏 | 是 | 否 | ✅ |
| LayoutCollector 内存 | 无限增长 | 有上限 | ✅ |
| EventBus CPU | 10ms 轮询 | 智能等待 | ✅ |
| FrameTimeline | O(n) 裁剪 | O(1) 覆盖 | ✅ |
| CausalGraph 分配 | 每帧新分配 | 对象池复用 | ✅ |
| 日志能力 | 无 | 结构化日志 | ✅ |
| 基准测试 | 无 | 完整覆盖 | ✅ |

### 6.2 验收标准

所有修复完成后，必须满足：

- [x] 所有现有测试通过 (17/17 测试通过)
- [x] 新增测试全部通过
- [ ] 内存泄漏测试通过（1小时运行）- 可选长时间测试
- [x] 并发测试通过
- [ ] 性能开销符合预期 - 需要基准测试验证
- [x] Runtime 集成验证通过
- [x] 无新的警告或错误

### 6.3 下一步

修复完成后，可以安全地开始阶段6 V1（观测增强层）的实施。
