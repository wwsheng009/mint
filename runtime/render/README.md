# Render Module

渲染辅助工具模块。提供帧率限制和视口管理功能。

## 职责

- **帧率限制**: 提供智能帧率控制，避免过度渲染导致的性能浪费
- **视口管理**: 提供虚拟滚动支持，只渲染可见区域的内容
- **虚拟列表**: 支持大数据量的虚拟列表渲染

## 注意

虽然此模块是 `runtime` 中唯一**可以**使用 `lipgloss` 的模块，但当前实现只提供轻量级的渲染辅助工具，不依赖 `lipgloss`。真正的样式应用由 `runtime/paint` 层处理。

## 核心概念

### 1. 帧率限制系统 (Throttler)

Throttler 用于控制渲染频率，防止过度渲染：

**特性**:
- 目标帧率控制（默认 60 FPS，范围 1-120 FPS）
- 帧时间记录和统计
- 自适应帧率调整
- 渲染策略支持（总是渲染、仅在脏标记时渲染、自适应渲染）

**渲染策略**:
- `StrategyAlways`: 总是渲染（受帧率限制）
- `StrategyOnDirty`: 仅在组件有脏标记时渲染
- `StrategyAdaptive`: 根据待渲染队列长度和脏标记自适应

### 2. 智能渲染器 (SmartRenderer)

SmartRenderer 封装了 Throttler 和渲染策略，提供智能化的渲染调度：

```go
type SmartRenderer struct {
    throttler *Throttler
    strategy  RenderStrategy
}
```

### 3. 视口管理 (Viewport)

Viewport 表示大的内容区域中的可见区域，支持虚拟滚动：

**特性**:
- 可见范围计算和缓存
- 滚动位置管理（ScrollTo, ScrollBy）
- 可见性检测（行、列、区域）
- 虚拟列表支持

**关键方法**:
- `GetVisibleRange()`: 获取可见区域范围
- `ScrollTo(x, y)`: 滚动到指定位置
- `IsRowVisible(row)`: 检测行是否可见
- `GetVisibleRows(itemHeight)`: 获取可见行索引列表

### 4. 虚拟列表 (VirtualListState)

VirtualListState 结合 Viewport 提供虚拟列表功能：

**特性**:
- 仅渲染可见项
- 滚动时动态更新可见范围
- 自动滚动选中项到视野内

## 使用示例

### 帧率限制基础使用

```go
// 创建 60 FPS 的帧率限制器
throttler := render.NewThrottler(60)

// 主渲染循环
for {
    if throttler.ShouldRender() {
        start := time.Now()
        renderScene()
        throttler.RecordFrameTime(time.Since(start))
    }
}
```

### 智能渲染器

```go
// 创建智能渲染器
renderer := render.NewSmartRenderer(60)

// 设置渲染策略为仅在脏标记时渲染
renderer.SetStrategy(render.StrategyOnDirty)

// 渲染时检查
if renderer.ShouldRender(isDirty) {
    renderer.Render(func() time.Duration {
        start := time.Now()
        renderComponents()
        return time.Since(start)
    })
}
```

### 自适应帧率

```go
throttler := render.NewThrottler(60)

// 启用自适应帧率
throttler.EnableAdaptive(true)

// 自适应模式下，系统会根据渲染性能自动调整帧率
// 如果渲染时间过长，会自动降低帧率
// 如果渲染很快，会适当提高帧率
```

### 帧率监控

```go
// 获取帧率统计信息
stats := throttler.Stats()
fmt.Printf("目标帧率: %d FPS\n", stats.TargetFPS)
fmt.Printf("实际帧率: %.2f FPS\n", stats.ActualFPS)
fmt.Printf("平均帧时间: %v\n", stats.AvgFrameTime)
fmt.Printf("待渲染帧数: %d\n", stats.PendingCount)
```

### 视口基础使用

```go
// 创建视口：内容 100x200，可见区域 80x60
viewport := render.NewViewport(100, 200, 80, 60)

// 滚动到指定位置
viewport.ScrollTo(10, 20)

// 获取可见范围
startX, endX, startY, endY := viewport.GetVisibleRange()
fmt.Printf("可见区域: (%d, %d) -> (%d, %d)\n", startX, startY, endX, endY)

// 检查某个区域是否可见
if viewport.IsVisible(5, 10, 20, 30) {
    // 渲染该区域
    renderRegion(5, 10, 20, 30)
}
```

### 虚拟列表

```go
// 配置虚拟列表
config := render.VirtualListConfig{
    ItemCount:      1000,      // 1000 个项目
    ItemHeight:     3,         // 每个项目 3 行高
    ViewportWidth:  80,        // 可见宽度
    ViewportHeight: 20,        // 可见高度
}

// 创建虚拟列表状态
state := render.NewVirtualListState(config)

// 选择项目（自动滚动到视野内）
state.SelectItem(50, 1000, 3, 20)

// 获取可见项目索引
visibleItems := state.GetVisibleItems(3)
for _, idx := range visibleItems {
    // 仅渲染可见项目
    renderItem(idx, idx == state.SelectedIndex)
}

// 渲染虚拟列表
render.RenderVirtualList(buffer, state, config, renderItemFunc)
```

## 核心类型

### Throttler

```go
type Throttler struct {
    targetFPS       int           // 目标帧率
    minInterval     time.Duration // 最小渲染间隔
    lastRender      time.Time     // 上次渲染时间
    pendingCount    int           // 待渲染帧数
    frameTimes      []time.Duration // 帧时间历史
    adaptive        bool          // 是否启用自适应
}

func NewThrottler(fps int) *Throttler
func (t *Throttler) ShouldRender() bool
func (t *Throttler) RecordFrameTime(d time.Duration)
func (t *Throttler) SetFPS(fps int)
func (t *Throttler) FPS() int
func (t *Throttler) ActualFPS() float64
func (t *Throttler) EnableAdaptive(enable bool)
func (t *Throttler) Stats() Stats
```

### SmartRenderer

```go
type SmartRenderer struct {
    throttler *Throttler
    strategy  RenderStrategy
}

func NewSmartRenderer(fps int) *SmartRenderer
func (r *SmartRenderer) ShouldRender(isDirty bool) bool
func (r *SmartRenderer) Render(fn func() time.Duration)
func (r *SmartRenderer) SetStrategy(s RenderStrategy)
```

### Viewport

```go
type Viewport struct {
    contentWidth   int  // 内容宽度
    contentHeight  int  // 内容高度
    viewportWidth  int  // 视口宽度
    viewportHeight int  // 视口高度
    scrollX, scrollY int // 滚动位置
}

func NewViewport(contentWidth, contentHeight, viewportWidth, viewportHeight int) *Viewport
func (v *Viewport) GetVisibleRange() (startX, endX, startY, endY int)
func (v *Viewport) ScrollTo(x, y int)
func (v *Viewport) ScrollBy(dx, dy int)
func (v *Viewport) IsRowVisible(row int) bool
func (v *Viewport) GetVisibleRows(itemHeight int) []int
```

### VirtualListState

```go
type VirtualListState struct {
    Viewport      *Viewport
    SelectedIndex int
}

func NewVirtualListState(config VirtualListConfig) *VirtualListState
func (l *VirtualListState) SelectItem(index int, itemCount, itemHeight, viewportHeight int)
func (l *VirtualListState) GetVisibleItems(itemHeight int) []int
```

## 文件结构

- `throttle.go` - 帧率限制器和智能渲染器
- `viewport.go` - 视口管理和虚拟列表支持

## 依赖约束

**可以使用**:
- `github.com/wwsheng009/mint/runtime` - 核心运行时类型
- `lipgloss`（虽然可以，但当前实现未使用）

**不能依赖**:
- Bubble Tea
- DSL 解析器
- 具体组件

## 与其他模块集成

### 与 Paint 集成

```go
// 在 Paint 层使用 SmartRenderer 控制 CellBuffer 更新频率
renderer := render.NewSmartRenderer(60)
renderer.SetStrategy(render.StrategyOnDirty)

if renderer.ShouldRender(buffer.IsDirty()) {
    buffer.Update()
    renderer.Render(func() time.Duration {
        // 渲染到终端
        return platform.Sync()
    })
}
```

### 与 Layout 集成

```go
// Layout 计算组件位置后，可以使用 Viewport 确定可见区域
viewport := render.NewViewport(totalWidth, totalHeight, screenWidth, screenHeight)

// 仅布局并渲染可见组件
if viewport.IsVisible(compX, compY, compWidth, compHeight) {
    layoutComponent(comp)
    renderComponent(comp)
}
```

### 与 Animation 集成

```go
// 动画使用高帧率
renderer.SetFPS(60)
renderer.SetStrategy(render.StrategyAlways)

// 静态内容使用低帧率
renderer.SetFPS(20)
renderer.SetStrategy(render.StrategyOnDirty)
```

## 最佳实践

### 1. 根据场景选择帧率

```go
// 动画场景：高帧率
fps := render.RecommendFPS("animation")  // 60 FPS

// 静态表单：低帧率
fps := render.RecommendFPS("form")  // 30 FPS

// 空闲状态：最低帧率
fps := render.RecommendFPS("idle")  // 20 FPS
```

### 2. 使用智能渲染策略

```go
// 对于大部分 UI，使用 StrategyOnDirty
renderer.SetStrategy(render.StrategyOnDirty)

// 对于动画，使用 StrategyAlways
renderer.SetStrategy(render.StrategyAlways)

// 对于复杂场景，使用 StrategyAdaptive
renderer.SetStrategy(render.StrategyAdaptive)
renderer.EnableAdaptive(true)
```

### 3. 虚拟列表优化

对于大数据量的列表，使用虚拟列表只渲染可见项：

```go
// 配置虚拟列表
config := render.VirtualListConfig{
    ItemCount:      10000,  // 大数据量
    ItemHeight:     2,
    ViewportWidth:  80,
    ViewportHeight: 20,
}

// 仅渲染约 10-12 个可见项，而非 10000 个
```

### 4. 监控帧率性能

```go
// 定期检查帧率统计
go func() {
    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        stats := renderer.Stats()
        log.Printf("FPS: %.2f (目标: %d), 待渲染: %d\n",
            stats.ActualFPS, stats.TargetFPS, stats.PendingCount)
    }
}()
```

## 常见问题

### Q: 为什么不直接使用 lipgloss？

A: 本模块提供的是渲染辅助工具（帧率控制、视口管理），而非样式应用。样式处理在 `runtime/paint` 层，那里可以决定是否使用 `lipgloss`。但当前的 `runtime/paint` 实现也是纯 Go 的，所以本模块保持纯 Go 更简单。

### Q: 如何选择合适的帧率？

A: 使用 `RecommendFPS(scene)` 或 `PresetFPS(presetName)`：
- 动画/过渡: 60 FPS
- 表单/输入: 30 FPS
- 静态内容: 15-20 FPS

### Q: 虚拟列表有什么优势？

A: 虚拟列表只渲染可见项，而非整个列表。对于 10000 项的列表：
- 传统列表: 渲染 10000 项 → 慢
- 虚拟列表: 渲染约 20 项 → 快

### Q: 启用自适应帧率后为什么帧率会自动降低？

A: 启用自适应后，如果平均渲染时间超过目标帧时间的 90%，系统会自动降低目标帧率以避免卡顿。如果你发现帧率低于预期，可以检查渲染性能或禁用自适应。

### Q: Viewport 如何处理内容小于视口的情况？

A: Viewport 会自动限制滚动位置。如果内容小于视口，`maxScrollX()` 和 `maxScrollY()` 会返回 0，无法滚动。
