# idea1.md 实施方案

## 实施日期
2026-01-29

## 概述

根据 `idea1_review.md` 的审查结果，制定详细的实施方案，将系统提升到 "实时渲染系统" 级别。

---

## 一、修复阶段 (低风险，高优先级)

### 1.1 统一 runeWidth 使用

**目标**: 完全使用 `runewidth.RuneWidth`，删除手写实现

**影响文件**: `runtime/paint/buffer.go`

**步骤**:

1. 删除 `runeWidth` 函数 (line 74-93)
2. 更新 `SetCell` 方法:
   ```go
   func (b *Buffer) SetCell(x, y int, char rune, s style.Style) {
       b.setCluster(x, y, string(char), runewidth.RuneWidth(char), s)
   }
   ```

**预期收益**:
- 正确处理 Nerd Font 图标
- 正确处理特殊 Unicode 字符
- 行为一致性

**风险**: 低

**工作量**: 0.5 小时

---

### 1.2 修复 IsCellChanged 逻辑

**目标**: 正确处理 head ↔ continuation 的变化

**影响文件**: `runtime/paint/buffer.go`

**修改**:
```go
func IsCellChanged(cell, prevCell Cell) bool {
    // 如果当前是 continuation，跳过（由主单元格处理）
    if cell.IsContinuation {
        return false
    }

    // 如果前一个是 continuation，当前是 head → 需要刷新
    if prevCell.IsContinuation && cell.Width > 0 {
        return true
    }

    // 正常比较
    return cell.Cluster != prevCell.Cluster || cell.Style != prevCell.Style
}
```

**风险**: 低

**工作量**: 0.5 小时

---

### 1.3 优化 Cursor 移动

**目标**: 添加光标位置跟踪，优化移动指令

**影响文件**: `runtime/paint/batch.go`

**修改**:

1. 在 `CommandBatch` 添加位置跟踪:
   ```go
   type CommandBatch struct {
       cmds    []DrawCmd
       styleVM *StyleStateMachine
       curX    int  // 当前光标 X
       curY    int  // 当前光标 Y
   }
   ```

2. 优化 `moveCursor`:
   ```go
   func (b *CommandBatch) moveCursor(x, y int) string {
       // 同一行，右侧，直接输出
       if y == b.curY && x >= b.curX {
           b.curX = x
           return ""  // 调用者直接输出字符
       }
       // 同一行，右侧小步移动
       if y == b.curY && x > b.curX && (x-b.curX) <= 5 {
           b.curX = x
           return "\x1b[" + itoa(x-b.curX) + "C"
       }
       // 绝对定位
       b.curX, b.curY = x, y
       return "\x1b[" + itoa(y+1) + ";" + itoa(x+1) + "H"
   }
   ```

**风险**: 中等（需要仔细测试边界情况）

**工作量**: 1 小时

---

## 二、核心功能实现 (中等风险，高优先级)

### 2.1 双缓冲 Renderer

**目标**: 实现完整的双缓冲渲染管线

**新文件**: `runtime/paint/renderer.go`

**结构**:
```go
package paint

import (
    "bytes"
    "sync"
)

// Renderer 双缓冲渲染器
type Renderer struct {
    mu sync.Mutex

    // 双缓冲
    front *Buffer  // 当前屏幕状态
    back  *Buffer  // 新一帧

    // 脏区域
    dirtyTracker *DirtyTracker

    // 渲染状态
    styleState  *StyleStateMachine
    cursorX     int
    cursorY     int

    // 输出
    output bytes.Buffer
}

// NewRenderer 创建渲染器
func NewRenderer(width, height int) *Renderer {
    return &Renderer{
        front:       NewBuffer(width, height),
        back:        NewBuffer(width, height),
        dirtyTracker: NewDirtyTracker(),
        styleState:  NewStyleStateMachine(),
    }
}

// Render 对比 front/back 并生成差异输出
func (r *Renderer) Render() string {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.output.Reset()

    // 执行 diff
    diff := r.dirtyTracker.Diff(r.front, r.back)

    // 渲染脏区域
    for _, region := range diff.DirtyRegions {
        r.renderRegion(region)
    }

    // 交换缓冲区
    r.swapBuffers()

    return r.output.String()
}

// renderRegion 渲染单个区域（使用 run 合并）
func (r *Renderer) renderRegion(region Rect) {
    for y := region.Y; y < region.Y+region.Height; y++ {
        r.renderLine(y, region)
    }
}

// renderLine 渲染单行，合并连续相同样式的片段
func (r *Renderer) renderLine(y int, region Rect) {
    x := region.X
    endX := region.X + region.Width

    for x < endX {
        cell := r.back.Cells[y][x]

        // 跳过未变化和 continuation
        if !r.isChanged(x, y, cell) {
            x++
            continue
        }

        // 开始一个 run
        startX := x
        style := cell.Style
        var text bytes.Buffer

        // 收集连续相同样式的片段
        for x < endX {
            c := r.back.Cells[y][x]
            if !r.isChanged(x, y, c) || c.Style != style || c.IsContinuation {
                break
            }
            text.WriteString(c.Cluster)
            x += c.Width
        }

        r.emitRun(startX, y, style, text.String())
    }
}

// emitRun 输出一个渲染批次
func (r *Renderer) emitRun(x, y int, style style.Style, text string) {
    // 移动光标
    if x != r.cursorX || y != r.cursorY {
        r.output.WriteString(r.moveCursor(x, y))
        r.cursorX, r.cursorY = x, y
    }

    // 设置样式
    if r.styleState.NeedsUpdate(style) {
        r.output.WriteString(r.styleState.Update(style))
    }

    // 输出文本
    r.output.WriteString(text)
    r.cursorX += len(text)
}

// isChanged 检查单元格是否变化
func (r *Renderer) isChanged(x, y int, cell Cell) bool {
    prevCell := r.front.Cells[y][x]
    return IsCellChanged(cell, prevCell)
}

// swapBuffers 交换前后缓冲区
func (r *Renderer) swapBuffers() {
    r.front, r.back = r.back, r.front

    // 清空新的 back buffer
    // (可选: 复制 front 内容作为起点)
}

// Resize 调整大小
func (r *Renderer) Resize(width, height int) {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.front = NewBuffer(width, height)
    r.back = NewBuffer(width, height)
    r.dirtyTracker.MarkAll()
}

// GetBackBuffer 获取用于绘制的缓冲区
func (r *Renderer) GetBackBuffer() *Buffer {
    return r.back
}
```

**风险**: 中等

**工作量**: 4 小时

**测试要点**:
- 宽字符边界处理
- 部分区域刷新
- 缓冲区交换正确性

---

### 2.2 Frame Engine

**目标**: 实现主事件循环，驱动整个渲染管线

**新文件**: `runtime/engine/engine.go`

**结构**:
```go
package engine

import (
    "sync/atomic"
    "time"

    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/scheduler"
)

// Engine 帧调度引擎
type Engine struct {
    // 渲染器
    renderer *paint.Renderer

    // 调度器
    scheduler *scheduler.Scheduler

    // 帧配置
    frameInterval time.Duration  // 16ms = 60fps

    // 状态
    running      atomic.Bool
    repaintNeeded atomic.Bool

    // 事件
    eventQueue chan Event
    quit       chan struct{}

    // 组件树根节点
    root Component
}

// Component 组件接口
type Component interface {
    ID() string
    Update(dt time.Duration)
    Layout()
    Paint(buf *paint.Buffer)
}

// Event 事件接口
type Event interface{}

// New 创建引擎
func New(width, height int, root Component) *Engine {
    return &Engine{
        renderer:      paint.NewRenderer(width, height),
        scheduler:     scheduler.New(),
        frameInterval: 16 * time.Millisecond,
        eventQueue:    make(chan Event, 100),
        quit:          make(chan struct{}),
        root:          root,
    }
}

// Run 启动主循环
func (e *Engine) Run() {
    if !e.running.CompareAndSwap(false, true) {
        return  // 已经在运行
    }

    ticker := time.NewTicker(e.frameInterval)
    defer ticker.Stop()
    defer e.running.Store(false)

    idleTimer := time.NewTimer(3 * time.Second)
    idleTimer.Stop()

    for {
        select {
        case ev := <-e.eventQueue:
            e.handleEvent(ev)
            e.repaintNeeded.Store(true)
            idleTimer.Reset(3 * time.Second)

        case <-ticker.C:
            if e.repaintNeeded.Load() {
                e.frame()
                e.repaintNeeded.Store(false)
                idleTimer.Reset(3 * time.Second)
            }

        case <-idleTimer.C:
            // 空闲检测：3秒无变化，停止 ticker
            // 事件会重新触发

        case <-e.quit:
            return
        }
    }
}

// frame 执行一帧
func (e *Engine) frame() {
    // 1. 处理调度器中的更新
    result := e.scheduler.ProcessNext(e, scheduler.DefaultProcessOptions())

    // 2. 更新组件状态
    e.root.Update(e.frameInterval)

    // 3. 布局
    e.root.Layout()

    // 4. 绘制到 back buffer
    buf := e.renderer.GetBackBuffer()
    e.root.Paint(buf)

    // 5. 渲染输出
    output := e.renderer.Render()
    print(output)
}

// handleEvent 处理事件
func (e *Engine) handleEvent(ev Event) {
    // 分发事件到组件树
    // ...
}

// PostEvent 投递事件
func (e *Engine) PostEvent(ev Event) {
    select {
    case e.eventQueue <- ev:
    default:
        // 队列满，丢弃
    }
}

// RequestRepaint 请求重绘
func (e *Engine) RequestRepaint() {
    e.repaintNeeded.Store(true)
}

// Stop 停止引擎
func (e *Engine) Stop() {
    close(e.quit)
}

// Resize 调整大小
func (e *Engine) Resize(width, height int) {
    e.renderer.Resize(width, height)
    e.RequestRepaint()
}

// Layout 实现 scheduler.Renderer
func (e *Engine) Layout(node interface{}) {
    if comp, ok := node.(Component); ok {
        comp.Layout()
    }
}

// Paint 实现 scheduler.Renderer
func (e *Engine) Paint(node interface{}) {
    if comp, ok := node.(Component); ok {
        buf := e.renderer.GetBackBuffer()
        comp.Paint(buf)
    }
}
```

**风险**: 中高

**工作量**: 6 小时

**测试要点**:
- 帧率稳定性
- 事件处理延迟
- 空闲检测正确性
- 优雅退出

---

## 三、高级功能实现 (中等风险，中优先级)

### 3.1 事件分发 (ZIndex 反向)

**目标**: 事件从最高层开始分发

**修改**: `runtime/engine/engine.go`

```go
// handleEvent 从最高 ZIndex 开始分发事件
func (e *Engine) handleEvent(ev Event) {
    // 获取所有 layer (需要从 Compositor 获取)
    layers := e.getLayersSortedByZIndex()

    // 从高到低分发
    for i := len(layers) - 1; i >= 0; i-- {
        if e.dispatchToLayer(layers[i], ev) {
            return  // 事件被处理
        }
    }
}

func (e *Engine) dispatchToLayer(layer *paint.Layer, ev Event) bool {
    // 命中测试和事件分发
    return false
}
```

---

### 3.2 Idle Detection

**已在 Engine 中实现**，使用 `idleTimer` 检测 3 秒无变化后停止 ticker。

---

## 四、实施顺序

### 阶段 1: 修复 (2 小时)
1. 统一 runeWidth (0.5h)
2. 修复 IsCellChanged (0.5h)
3. 优化 Cursor 移动 (1h)

### 阶段 2: 双缓冲 (4 小时)
1. 实现 Renderer
2. 添加测试

### 阶段 3: Frame Engine (6 小时)
1. 实现 Engine
2. 集成 Scheduler
3. 添加测试

### 阶段 4: 高级功能 (2 小时)
1. ZIndex 事件分发
2. 集成测试

**总计**: 约 14 小时

---

## 五、测试计划

### 5.1 单元测试

| 测试 | 文件 | 覆盖 |
|-----|------|------|
| `TestRenderer_Render` | `renderer_test.go` | 双缓冲渲染 |
| `TestRenderer_RunMerge` | `renderer_test.go` | Run 合并 |
| `TestEngine_Frame` | `engine_test.go` | 帧循环 |
| `TestEngine_Idle` | `engine_test.go` | 空闲检测 |
| `TestWideChar_Cleanup` | `buffer_wide_test.go` | 宽字符清理 |
| `TestWideChar_Override` | `buffer_wide_test.go` | 宽字符覆盖 |

### 5.2 集成测试

- 完整 UI 应用渲染
- 宽字符混排渲染
- 高频更新场景
- SSH 低延迟场景

---

## 六、向后兼容

所有修改保持 API 兼容：

- `Buffer` API 不变
- `Layer` API 不变
- `Scheduler` API 不变
- 新增 `Renderer` 和 `Engine` 为独立组件

---

## 七、性能指标

### 目标

| 指标 | 当前 | 目标 |
|-----|------|------|
| 每帧输出量 | ~5KB | <2KB |
| SSH 延迟 | 一般 | 顺滑 |
| CPU 占用 (空闲) | 高 | ~0% |
| 动画掉帧 | 有 | 无 |

### 测量方法

```go
// 添加性能统计
type PerfStats struct {
    FramesPerSecond int
    AvgFrameBytes   int
    AvgFrameTime    time.Duration
    CPUUsage        float64
}
```

---

## 八、风险和缓解

| 风险 | 概率 | 影响 | 缓解 |
|-----|------|------|------|
| 双缓冲 bug | 中 | 高 | 充分测试 |
| 帧循环阻塞 | 低 | 中 | 使用 goroutine |
| 内存泄漏 | 低 | 中 | 定期 review |
| 性能回退 | 低 | 中 | benchmark |

---

## 九、完成标准

实施完成后，系统应达到：

1. ✅ 所有 `runeWidth` 使用 `runewidth`
2. ✅ 完整的双缓冲渲染管线
3. ✅ 帧调度主循环 (60fps)
4. ✅ 空闲检测 (3秒)
5. ✅ Run 合并输出
6. ✅ 光标移动优化
7. ✅ ZIndex 事件分发

此时系统达到 **Neovim / Helix / WezTerm 同级思路**。
