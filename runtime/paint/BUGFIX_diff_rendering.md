# TUI Diff 模式渲染问题分析报告

## 问题概述

### 问题描述
在 `TUI_OUTPUT_MODE="diff"` 模式下，运行 `go run ./framework/examples/login/interactive` 时出现以下问题：

1. **界面无响应**: 交互界面没有任何变化，无法响应用户输入
2. **内存泄漏**: 内存持续增长，测试时观察到占用 9GB 内存
3. **输出异常**: 日志显示 `HasChanges=true` 但 `OutputLen=0`

### 影响范围
- 所有使用 diff 渲染模式的 TUI 应用
- 涉及文件：`runtime/paint/renderer.go`, `runtime/paint/dirty.go`, `runtime/paint/buffer.go`

---

## 根本原因分析

### 问题 1: 无限循环导致 CPU 占用

**位置**: `runtime/paint/renderer.go:renderLine`

**原因代码**:
```go
// 只收集当前单元格
runText.WriteString(cell.Cluster)
x += cell.Width  // ❌ 如果 cell.Width == 0，x 不会前进
```

**触发条件**:
1. 单元格的 `Cluster` 为空字符串或 `"\x00"`
2. 单元格的 `Width` 为 0
3. `IsCellChanged` 返回 `true`（可能是因为 Style 变化）

**后果**:
- `x` 永远不会增加，循环永不终止
- CPU 100% 占用
- 程序挂起

### 问题 2: swapBuffers 逻辑错误

**位置**: `runtime/paint/renderer.go:swapBuffers`

**原始代码**:
```go
func (r *Renderer) swapBuffers() {
    r.front, r.back = r.back, r.front  // ❌ 只交换，不复制
}
```

**问题分析**:

双缓冲渲染的正确流程应该是：
```
初始状态: front = 空, back = 绘制区域 A
Render(): 输出 A 与空的差异
交换后:   front = A, back = 空
下一帧:   back = 绘制区域 B
Render(): 输出 B 与 A 的差异
```

但实际发生了什么：
```
初始状态: front = 空, back = 绘制区域 A
Render(): 输出 A
交换后:   front = A, back = 空 (旧 front)
下一帧:   back = 绘制区域 B
Render(): 比较 B 与 空 → 输出完整内容！
```

**后果**:
- 每次 Render 都输出全屏内容
- `Diff` 算法失效
- 无法正确检测增量变化

### 问题 3: 内存泄漏

**位置**: `runtime/paint/dirty.go:extractDirtyRegions`

**原因代码**:
```go
func (d *DirtyTracker) extractDirtyRegions(...) []Rect {
    visited := make([][]bool, height)  // ❌ 每帧分配新数组
    for i := range visited {
        visited[i] = make([]bool, width)
    }
    // ...
}
```

**计算**:
- 80x24 终端 = 1920 个 bool
- 每帧分配约 2KB
- 60 FPS × 2KB = 120KB/秒
- 如果频繁调用，累积效应明显

### 问题 4: IsCellChanged 与 cellsEqual 不一致

**位置**:
- `compareBuffersWithGrid` 使用 `cellsEqual`
- `renderLine` 使用 `IsCellChanged`

**差异**:
```go
// cellsEqual - 简单比较
func cellsEqual(a, b Cell) bool {
    return a.Cluster == b.Cluster && a.Style == b.Style && a.Width == b.Width
}

// IsCellChanged - 处理宽字符
func IsCellChanged(cell, prevCell Cell) bool {
    if cell.IsContinuation {
        return false  // 延续单元格不比较
    }
    if prevCell.IsContinuation {
        return true  // 前值是延续，需要刷新
    }
    return cell.Cluster != prevCell.Cluster || cell.Style != prevCell.Style
}
```

**后果**:
- Diff 认为某单元格变化了
- renderLine 认为没有变化
- 输出与预期不一致

---

## 处理过程

### 阶段 1: 识别无限循环

**日志分析**:
```
[renderLine] changed at x=3: cell.Cluster="请", prev.Cluster="请"
[renderLine] emitted 1 runs
```

发现 `emitted 1 runs` 但应该有多个变化，且 `OutputLen=0`。

**修复方案**:
```go
// 确保 x 至少前进 1
width := cell.Width
if width <= 0 {
    width = 1
}
x += width
```

### 阶段 2: 修复 swapBuffers

**修复方案**:
```go
func (r *Renderer) swapBuffers() {
    r.front, r.back = r.back, r.front

    // 将 front 的内容复制到 back，作为下一帧绘制的基准
    for y := 0; y < r.front.Height; y++ {
        copy(r.back.Cells[y], r.front.Cells[y])
    }
}
```

**原理**:
- 交换后，back buffer 需要与 front buffer 内容一致
- 应用程序只绘制变化的区域
- Diff 算法只检测真正的变化

### 阶段 3: 修复内存泄漏

**修复方案**:
```go
type DirtyTracker struct {
    // ... 其他字段

    // 复用的 visited 数组
    visited      [][]bool
    visitedWidth int
    visitedHeight int
}

func (d *DirtyTracker) extractDirtyRegions(...) []Rect {
    // 复用或重新分配 visited 数组
    if d.visited == nil || d.visitedWidth != width || d.visitedHeight != height {
        d.visited = make([][]bool, height)
        for i := range d.visited {
            d.visited[i] = make([]bool, width)
        }
        d.visitedWidth = width
        d.visitedHeight = height
    } else {
        // 清空而非重新分配
        for y := 0; y < height; y++ {
            for x := 0; x < width; x++ {
                d.visited[y][x] = false
            }
        }
    }
    // ...
}
```

### 阶段 4: 统一比较逻辑

**修复方案**:
```go
// compareBuffersWithGrid 使用 IsCellChanged 而非 cellsEqual
changed := IsCellChanged(currCell, prevCell)
```

---

## 防止措施

### 1. 代码审查 Checklist

**循环安全检查**:
- [ ] 所有循环都有明确的终止条件
- [ ] 循环变量必须在每次迭代中更新
- [ ] 考虑边界条件（Width=0, Height=0 等）

**内存管理检查**:
- [ ] 热路径避免频繁分配
- [ ] 考虑对象池或复用模式
- [ ] 使用 `pprof` 定期检查内存泄漏

**双缓冲模式检查**:
- [ ] swap 后是否需要复制/清空
- [ ] 下一帧的基准状态是否正确
- [ ] 边界情况（nil buffer, size change）是否处理

### 2. 测试策略

**单元测试覆盖**:
```go
// 测试空 cluster 不会导致无限循环
func TestRenderLine_EmptyClusterNoInfiniteLoop(t *testing.T)

// 测试宽字符的样式变化
func TestRenderLine_WideCharStyleChange(t *testing.T)

// 测试 swap 后的状态
func TestSwapBuffers_StateCorrectness(t *testing.T)
```

**基准测试**:
```go
func BenchmarkRenderer_FullRender(b *testing.B)
func BenchmarkRenderer_PartialUpdate(b *testing.B)
```

**内存测试**:
```go
func TestRenderer_NoMemoryLeak(t *testing.T) {
    var m1, m2 runtime.MemStats
    runtime.ReadMemStats(&m1)

    renderer := NewRenderer(80, 24)
    for i := 0; i < 1000; i++ {
        // ... 执行渲染
    }

    runtime.ReadMemStats(&m2)
    // 检查内存增长是否合理
}
```

### 3. 调试工具

**环境变量**:
```bash
# 启用渲染调试
export TUI_RENDER_DEBUG=1

# 启用内存分析
export TUI_MEM_PROFILE=1
```

**日志输出**:
```go
if os.Getenv("TUI_RENDER_DEBUG") == "1" {
    fmt.Fprintf(os.Stderr, "[renderLine] x=%d, width=%d\n", x, width)
}
```

### 4. 架构改进建议

**短期改进**:
1. 添加断言检查循环变量前进
2. 添加内存分配监控
3. 完善 diff 算法的边界处理

**长期改进**:
1. 考虑使用对象池管理 buffer
2. 实现增量式 dirty region 检测
3. 添加可视化的渲染调试工具

---

## 附录: 相关文件

| 文件 | 修改内容 |
|------|----------|
| `runtime/paint/renderer.go` | 修复 renderLine 无限循环，修复 swapBuffers 逻辑 |
| `runtime/paint/dirty.go` | 复用 visited 数组，统一使用 IsCellChanged |
| `runtime/paint/renderer_test.go` | 添加渲染器测试用例 |

---

## 总结

本次问题修复涉及三个核心问题：

1. **无限循环**: Width=0 时循环变量不前进 → 添加最小步长检查
2. **状态不一致**: swapBuffers 不复制内容 → 添加复制逻辑
3. **内存泄漏**: 每帧分配 visited 数组 → 复用数组

这些问题都是典型的"热路径"性能问题，在正常情况下难以发现，但在高频率调用时（60 FPS 渲染）会迅速暴露。

**经验教训**:
- 热路径代码需要特别关注循环不变式和内存分配
- 双缓冲模式的状态转换需要仔细设计
- 完善的测试覆盖是预防此类问题的关键
