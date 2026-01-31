# Style Diff 优化设计文档

**版本**: v1.0
**日期**: 2026-01-31
**来源**: idea/idea5.1_style_diff.md
**状态**: 🔴 高优先级

---

## 一、问题分析

### 1.1 终端渲染的性能瓶颈

在终端 TUI 渲染中，**样式切换比字符输出更昂贵**：

```go
// 低效方式：每个字符都切换样式
for _, cell := range cells {
    fmt.Print(ansi.FgColor(cell.Fg))  // 切换前景色
    fmt.Print(ansi.BgColor(cell.Bg))  // 切换背景色
    fmt.Print(ansi.Bold)               // 切换粗体
    fmt.Print(cell.Char)
}

// 对于 1000 个字符的屏幕：
// - 样式切换序列: 1000 * 3 = 3000 次
// - ANSI 代码长度: ~20 字节/次
// - 总输出: ~60,000 字节
```

### 1.2 优化目标

| 指标 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| 样式切换次数 | ~3000 | ~20 | 150x |
| ANSI 代码长度 | ~60KB | ~1KB | 60x |
| 渲染时间 | ~50ms | ~5ms | 10x |

---

## 二、核心设计

### 2.1 终端状态追踪

```go
// render/terminal_state.go

package render

import "github.com/wwsheng009/mint/runtime/style"

// TerminalState 终端当前样式状态
type TerminalState struct {
    FgColor     *style.Color
    BgColor     *style.Color
    Bold        bool
    Italic      bool
    Underline   bool
    Strikethrough bool
    Reverse     bool
}

// Equals 比较两个状态是否相等
func (s *TerminalState) Equals(other *TerminalState) bool {
    if s.Bold != other.Bold {
        return false
    }
    if s.Italic != other.Italic {
        return false
    }
    // ... 其他属性
    return true
}

// Diff 计算状态差异
func (s *TerminalState) Diff(target *TerminalState) *StyleChange {
    changes := &StyleChange{}

    if !colorEqual(s.FgColor, target.FgColor) {
        changes.FgColor = target.FgColor
    }
    if s.Bold != target.Bold {
        changes.Bold = &target.Bold
    }
    // ...

    return changes
}

func colorEqual(a, b *style.Color) bool {
    if a == nil && b == nil {
        return true
    }
    if a == nil || b == nil {
        return false
    }
    return a.R == b.R && a.G == b.G && a.B == b.B
}
```

### 2.2 样式变更

```go
// render/style_change.go

package render

// StyleChange 样式变更
type StyleChange struct {
    FgColor       *style.Color
    BgColor       *style.Color
    Bold          *bool
    Italic        *bool
    Underline     *bool
    Strikethrough *bool
    Reverse       *bool
}

// IsEmpty 检查是否为空变更
func (c *StyleChange) IsEmpty() bool {
    return c.FgColor == nil &&
           c.BgColor == nil &&
           c.Bold == nil &&
           c.Italic == nil &&
           c.Underline == nil &&
           c.Strikethrough == nil &&
           c.Reverse == nil
}

// ToANSI 转换为 ANSI 序列
func (c *StyleChange) ToANSI(lastStyle *style.Style) string {
    var buf strings.Builder

    if c.FgColor != nil {
        buf.WriteString(ansi.FgColor(*c.FgColor))
    }
    if c.BgColor != nil {
        buf.WriteString(ansi.BgColor(*c.BgColor))
    }
    if c.Bold != nil {
        if *c.Bold {
            buf.WriteString(ansi.Bold)
        } else {
            buf.WriteString(ansi.ResetBold)
        }
    }
    // ...

    return buf.String()
}
```

---

## 三、Run-Length Encoding 优化

### 3.1 Run 定义

```go
// render/rle.go

package render

// Run 连续相同样式的字符序列
type Run struct {
    Style  style.Style
    Start  int  // 在 Buffer 中的起始位置
    Length int  // 字符数
    Text   strings.Builder
}

// RunLengthEncoder RLE 编码器
type RunLengthEncoder struct {
    runs   []Run
    buffer *Buffer
}

// Encode 编码 Buffer 为 Runs
func (e *RunLengthEncoder) Encode(buffer *Buffer) []Run {
    e.buffer = buffer
    e.runs = make([]Run, 0)

    for y := 0; y < buffer.Height; y++ {
        e.encodeRow(y)
    }

    return e.runs
}

// encodeRow 编码一行
func (e *RunLengthEncoder) encodeRow(y int) {
    row := e.buffer.Cells[y]
    if len(row) == 0 {
        return
    }

    currentRun := Run{
        Style:  row[0].Style,
        Start:  0,
        Length: 1,
    }
    currentRun.Text.WriteRune(row[0].Char)

    for x := 1; x < len(row); x++ {
        cell := row[x]

        if cell.Style.Equals(currentRun.Style) {
            // 相同样式，合并到当前 Run
            currentRun.Length++
            currentRun.Text.WriteRune(cell.Char)
        } else {
            // 样式变化，保存当前 Run，开始新 Run
            e.runs = append(e.runs, currentRun)

            currentRun = Run{
                Style:  cell.Style,
                Start:  x,
                Length: 1,
            }
            currentRun.Text.WriteRune(cell.Char)
        }
    }

    // 保存最后一个 Run
    e.runs = append(e.runs, currentRun)
}
```

---

## 四、输出优化器

### 4.1 优化器接口

```go
// render/optimizer.go

package render

import (
    "io"
    "strings"
)

// Optimizer 输出优化器
type Optimizer struct {
    writer      io.Writer
    state       *TerminalState
    ansiBuilder *strings.Builder
}

// NewOptimizer 创建优化器
func NewOptimizer(w io.Writer) *Optimizer {
    return &Optimizer{
        writer:      w,
        state:       &TerminalState{},
        ansiBuilder: &strings.Builder{},
    }
}

// WriteBuffer 优化写入 Buffer
func (o *Optimizer) WriteBuffer(buffer *Buffer) error {
    // 1. RLE 编码
    encoder := &RunLengthEncoder{}
    runs := encoder.Encode(buffer)

    // 2. 输出每个 Run
    for _, run := range runs {
        o.writeRun(run, buffer.Width, buffer.Height)
    }

    // 3. 重置样式到默认状态
    o.resetStyle()

    return nil
}

// writeRun 写入一个 Run
func (o *Optimizer) writeRun(run Run, width, height int) error {
    // 1. 计算目标状态
    targetState := styleToTerminalState(run.Style)

    // 2. 计算差异
    change := o.state.Diff(targetState)

    // 3. 只输出变化的样式
    if !change.IsEmpty() {
        ansi := change.ToANSI(nil)
        o.state.Apply(change)
        o.writer.Write([]byte(ansi))
    }

    // 4. 输出文本
    o.writer.Write([]byte(run.Text.String()))

    return nil
}

// resetStyle 重置样式
func (o *Optimizer) resetStyle() {
    o.writer.Write([]byte(ansi.Reset))
    o.state = &TerminalState{}
}
```

### 4.2 位置计算

```go
// render/position.go

package render

// PositionCalculator 位置计算器
type PositionCalculator struct {
    bufferWidth  int
    bufferHeight int
}

// CalculatePosition 计算光标位置
func (p *PositionCalculator) CalculatePosition(index int) (x, y int) {
    y = index / p.bufferWidth
    x = index % p.bufferWidth
    return
}

// MoveCursor 移动光标（使用 ANSI 序列）
func (p *PositionCalculator) MoveCursor(x, y int, writer io.Writer) {
    // 使用 ANSICursorPosition 而非逐字符移动
    seq := fmt.Sprintf("\x1b[%d;%dH", y+1, x+1)
    writer.Write([]byte(seq))
}
```

---

## 五、性能基准

### 5.1 基准测试

```go
// render/style_diff_bench_test.go

package render

import (
    "testing"
)

// BenchmarkStyleDiffOptimized 优化版本
func BenchmarkStyleDiffOptimized(b *testing.B) {
    buffer := createTestBuffer(80, 24)
    optimizer := NewOptimizer(io.Discard)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        optimizer.WriteBuffer(buffer)
    }
}

// BenchmarkStyleDiffNaive 未优化版本
func BenchmarkStyleDiffNaive(b *testing.B) {
    buffer := createTestBuffer(80, 24)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        writeNaive(buffer, io.Discard)
    }
}

// BenchmarkStyleDiffReduction 样式减少比例
func BenchmarkStyleDiffReduction(b *testing.B) {
    buffer := createTestBuffer(80, 24)

    // 统计样式切换次数
    naiveCount := countNaiveStyleChanges(buffer)
    optimizedCount := countOptimizedStyleChanges(buffer)

    b.ReportMetric(float64(naiveCount/optimizedCount), "reduction")
}
```

### 5.2 目标性能

| 指标 | 目标值 |
|------|--------|
| 全屏渲染 (80x24) | < 5ms |
| 样式切换减少 | ≥ 95% |
| 输出字节数减少 | ≥ 90% |
| 内存分配 | 最小化 |

---

## 六、集成方案

### 6.1 与渲染管线集成

```go
// render/renderer.go

package render

type Renderer struct {
    buffer    *Buffer
    optimizer *Optimizer
    writer    io.Writer
}

func (r *Renderer) Render() error {
    // 1. 生成 DrawCmd
    cmds := r.generateDrawCmds()

    // 2. 光栅化到 Buffer
    r.buffer = r.rasterize(cmds)

    // 3. 优化输出
    return r.optimizer.WriteBuffer(r.buffer)
}
```

### 6.2 兼容性处理

```go
// render/compat.go

package render

// CompatibilityMode 兼容性模式
type CompatibilityMode int

const (
    // ModeOptimized 完全优化模式
    ModeOptimized CompatibilityMode = iota

    // ModeSafe 安全模式（某些终端可能不支持）
    ModeSafe

    // ModeLegacy 传统模式（用于调试）
    ModeLegacy
)

// SetCompatibilityMode 设置兼容模式
func (o *Optimizer) SetCompatibilityMode(mode CompatibilityMode) {
    o.mode = mode

    if mode == ModeSafe || mode == ModeLegacy {
        // 禁用某些优化
        o.optimizeLevel = 0
    } else {
        o.optimizeLevel = 2
    }
}
```

---

## 七、实施计划

### 阶段 1: 基础实现 (Day 12)

- [ ] 实现 TerminalState
- [ ] 实现 StyleChange
- [ ] 实现 Diff 算法

### 阶段 2: RLE 优化 (Day 13)

- [ ] 实现 RunLengthEncoder
- [ ] 实现 Run 结构
- [ ] 编写单元测试

### 阶段 3: 输出优化器 (Day 14)

- [ ] 实现 Optimizer
- [ ] 集成到渲染管线
- [ ] 性能测试

---

## 八、测试策略

### 8.1 单元测试

```go
func TestTerminalStateDiff(t *testing.T) {
    state := &TerminalState{Bold: true}
    target := &TerminalState{Bold: false}

    change := state.Diff(target)

    assert.NotNil(t, change.Bold)
    assert.False(t, *change.Bold)
}

func TestRLEncoding(t *testing.T) {
    buffer := createBuffer(
        "AAAA",  // 相同样式
        "BBBB",  // 相同样式
    )

    encoder := &RunLengthEncoder{}
    runs := encoder.Encode(buffer)

    assert.Equal(t, 2, len(runs))
    assert.Equal(t, 4, runs[0].Length)
    assert.Equal(t, 4, runs[1].Length)
}
```

### 8.2 集成测试

```go
func TestOptimizerOutput(t *testing.T) {
    buffer := createTestBuffer(80, 24)

    var output strings.Builder
    optimizer := NewOptimizer(&output)

    err := optimizer.WriteBuffer(buffer)

    assert.NoError(t, err)
    assert.Less(t, output.Len(), 10000) // 应该远小于全量输出
}
```

---

## 九、API 文档

```go
// render/style_diff.go (公开 API)

package render

// OptimizeOutput 优化输出（便捷函数）
func OptimizeOutput(buffer *Buffer, writer io.Writer) error {
    optimizer := NewOptimizer(writer)
    return optimizer.WriteBuffer(buffer)
}

// CountStyleChanges 统计样式变化（用于调试）
func CountStyleChanges(buffer *Buffer) int {
    count := 0
    // ...
    return count
}
```

---

**文档版本**: v1.0
**最后更新**: 2026-01-31
