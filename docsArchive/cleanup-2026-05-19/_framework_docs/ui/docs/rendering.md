# 渲染管线 (Rendering Pipeline)

本文档描述 Mint UI 声明式框架的渲染管线实现。

---

## 概述

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Rendering Pipeline                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  VNode Tree                                                            │
│       │                                                                │
│       ▼                                                                │
│  ┌─────────────┐                                                       │
│  │   Layout    │  计算每个组件的位置和大小                               │
│  └─────────────┘                                                       │
│       │                                                                │
│       ▼                                                                │
│  ┌─────────────┐                                                       │
│  │    Draw     │  生成绘制命令 (DrawCmd)                                │
│  └─────────────┘                                                       │
│       │                                                                │
│       ▼                                                                │
│  ┌─────────────┐    ┌──────────────┐                                  │
│  │   Buffer    │───▶│ DirtyTracker │  跟踪变化区域                      │
│  │   (Back)    │    └──────────────┘                                  │
│  └─────────────┘                                                       │
│       │                                                                │
│       ▼                                                                │
│  ┌─────────────┐    ┌──────────────┐                                  │
│  │  Diff       │◀───│ StyleState   │  最小化 ANSI 切换                  │
│  └─────────────┘    └──────────────┘                                  │
│       │                                                                │
│       ▼                                                                │
│  ┌─────────────┐                                                       │
│  │  RLE Encode │  压缩连续相同样式的单元格                              │
│  └─────────────┘                                                       │
│       │                                                                │
│       ▼                                                                │
│  ┌─────────────┐                                                       │
│  │  ANSI Out   │  生成终端输出                                          │
│  └─────────────┘                                                       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 核心组件

### 1. Buffer (缓冲区)

**文件**: `runtime/paint/buffer.go`

```go
type Buffer struct {
    Width  int
    Height int
    Cells  [][]Cell
}

type Cell struct {
    Cluster        string      // 图形集群（可见字符）
    Style          style.Style // 视觉样式
    Width          int         // 显示宽度 (1 或 2)
    IsContinuation bool        // 宽字符续单元格
    ZIndex         int         // 层级
    NodeID         uint64      // 关联的 VNode ID
    Selected       bool        // 选中状态
}
```

**特性**:
- 双缓冲渲染（front buffer 和 back buffer）
- 支持宽字符（中文、emoji、组合字符）
- Z-Index 支持层级排序

---

### 2. DirtyTracker (脏区域跟踪)

**文件**: `runtime/paint/dirty.go`

```go
type DirtyTracker struct {
    cells        map[cellRef]struct{}
    rects        []Rect
    allDirty     bool
    prevBuffer   *Buffer
    changedCells int
}

type DiffResult struct {
    DirtyRegions []Rect // 发生变化的矩形区域
    HasChanges   bool
    ChangedCells int
}
```

**功能**:
- 跟踪发生变化的单元格
- 合并相邻的脏区域
- 只重绘变化的部分

**示例**:
```go
tracker := NewDirtyTracker()

// 标记单个单元格
tracker.MarkCell(x, y)

// 标记矩形区域
tracker.MarkRect(x, y, width, height)

// 标记全部
tracker.MarkAll()

// 获取变化结果
diff := tracker.GetDiff(currentBuffer)
```

---

### 3. StyleStateMachine (样式状态机)

**文件**: `runtime/paint/style_state.go`

```go
type StyleStateMachine struct {
    current style.Style
}

func (s *StyleStateMachine) Update(st style.Style) string
```

**功能**:
- 追踪当前终端样式
- 只输出变化的样式属性
- 减少输出字节数

**效果对比**:
```
无优化: \x1b[31m\x1b[1mText\x1b[0m\x1b[32m\x1b[1mText\x1b[0m
有优化: \x1b[31;1mText\x1b[0m\x1b[32mText
减少:    ~40%
```

---

### 4. RLE (Run-Length Encoding)

**文件**: `runtime/paint/rle.go`

```go
type Run struct {
    Cell  Cell  // 单元格值（样式 + 集群）
    Count int   // 连续单元格数量
    X     int   // 起始 X 位置
    Y     int   // Y 位置
}

func EncodeRLE(row []Cell, width int) []Run
```

**功能**:
- 压缩连续相同样式的单元格
- 减少渲染命令数量

**压缩示例**:
```
输入: [Cell{Style:red, Text:"H"},
      Cell{Style:red, Text:"e"},
      Cell{Style:red, Text:"l"},
      Cell{Style:red, Text:"l"},
      Cell{Style:red, Text:"o"}]

输出: [Run{Cell{Style:red, Text:"H"}, Count:1, X:0},
      Run{Cell{Style:red, Text:"e"}, Count:1, X:1},
      Run{Cell{Style:red, Text:"l"}, Count:2, X:2},  ← 压缩
      Run{Cell{Style:red, Text:"o"}, Count:1, X:4}]
```

---

### 5. Renderer (渲染器)

**文件**: `runtime/paint/renderer.go`

```go
type Renderer struct {
    front      *Buffer
    back       *Buffer
    dirty      *DirtyTracker
    styleState *StyleStateMachine
    output     *bytes.Buffer
    mu         sync.Mutex
}

func (r *Renderer) Render() string
```

**渲染流程**:
1. 比较 front buffer 和 back buffer
2. 收集脏区域
3. 生成 ANSI 输出
4. 交换缓冲区

---

## 辅助功能

### cursorMove (光标移动)

```go
func cursorMove(fromX, toX, y int) string
```

**策略**:
- 小距离移动 (< 10 字符): 使用相对命令 `\x1b[nC` / `\x1b[nD`
- 大距离移动: 使用绝对定位 `\x1b[y;xH`

**示例**:
```go
cursorMove(5, 7, 0)  // "\x1b[2C"  (向前移动 2 位)
cursorMove(7, 5, 0)  // "\x1b[2D"  (向后移动 2 位)
cursorMove(0, 20, 5) // "\x1b[6;21H" (绝对定位到第6行第21列)
```

---

### styleToANSI (样式转换)

```go
func styleToANSI(s style.Style) string
```

**支持的样式**:
| 属性 | ANSI 代码 |
|------|-----------|
| Bold | 1 |
| Italic | 3 |
| Underline | 4 |
| Blink | 5 |
| Reverse | 7 |
| Strikethrough | 9 |
| 前景色 | `颜色名` 或 `#RRGGBB` |
| 背景色 | `颜色名` 或 `#RRGGBB` |

---

## 性能优化

### 优化层级

```
┌─────────────────────────────────────────────────────────────┐
│  Level 1: 脏区域跟踪                                         │
│  → 只渲染变化的部分                                         │
│  → 节省: 50-99%                                            │
├─────────────────────────────────────────────────────────────┤
│  Level 2: 样式状态机                                        │
│  → 只输出变化的样式属性                                     │
│  → 节省: 40-60%                                            │
├─────────────────────────────────────────────────────────────┤
│  Level 3: RLE 编码                                          │
│  → 合并连续相同样式的单元格                                 │
│  → 节省: 50-99% (重复内容)                                 │
├─────────────────────────────────────────────────────────────┤
│  Level 4: 智能光标移动                                      │
│  → 选择最短的光标移动路径                                   │
│  → 节省: 10-30%                                            │
└─────────────────────────────────────────────────────────────┘
```

### 性能指标

| 场景 | 无优化 | 有优化 | 节省 |
|------|--------|--------|------|
| 静态文本 | ~5000 bytes | ~200 bytes | 96% |
| 重复字符 | ~3000 bytes | ~100 bytes | 97% |
| 全屏刷新 | ~40000 bytes | ~800 bytes | 98% |

---

## 使用示例

### 基本渲染

```go
package main

import (
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
)

func main() {
    // 创建缓冲区
    buf := paint.NewBuffer(80, 24)

    // 绘制文本
    style := style.Style{}.
        Foreground("red").
        Bold(true)

    buf.SetCell(10, 5, 'H', style)
    buf.SetCell(11, 5, 'e', style)
    buf.SetCell(12, 5, 'l', style)
    buf.SetCell(13, 5, 'l', style)
    buf.SetCell(14, 5, 'o', style)

    // 创建渲染器
    renderer := paint.NewRenderer(buf)

    // 渲染
    output := renderer.Render()
    print(output)
}
```

### RLE 渲染

```go
// 使用 RLE 优化渲染
renderer := paint.NewRLERenderer()

// 渲染单行
row := buf.Cells[5]
output := renderer.RenderRow(row, 80, 5)
print(output)
```

### 优化输出

```go
// 只输出变化的部分
diff := &paint.DiffResult{
    HasChanges: true,
    DirtyRegions: []paint.Rect{
        {X: 10, Y: 5, Width: 5, Height: 1},
    },
    ChangedCells: 5,
}

output := paint.OptimizedOutput(buf, diff)
print(output)
```

---

## 测试

### 运行测试

```bash
# 运行所有 paint 测试
go test ./runtime/paint/... -v

# 运行 RLE 测试
go test ./runtime/paint/... -v -run RLE

# 运行性能测试
go test ./runtime/paint/... -bench=. -benchmem
```

### 测试覆盖

```bash
go test ./runtime/paint/... -cover
# ok  	github.com/wwsheng009/mint/runtime/paint	2.005s	coverage: 49.5%
```

---

## API 参考

### Buffer

| 方法 | 描述 |
|------|------|
| `NewBuffer(width, height int)` | 创建新缓冲区 |
| `SetCell(x, y int, ch rune, style Style)` | 设置单元格 |
| `SetString(x, y int, s string, style Style)` | 设置字符串 |
| `Clear()` | 清空缓冲区 |
| `GetSize() (width, height int)` | 获取尺寸 |

### Renderer

| 方法 | 描述 |
|------|------|
| `NewRenderer() *Renderer` | 创建渲染器 |
| `Render() string` | 渲染整个缓冲区 |
| `GetStats() RenderStats` | 获取渲染统计 |

### RLE

| 函数 | 描述 |
|------|------|
| `EncodeRLE(row []Cell, width int) []Run` | RLE 编码 |
| `NewRLERenderer() *RLERenderer` | 创建 RLE 渲染器 |
| `OptimizedOutput(buf *Buffer, diff *DiffResult) string` | 优化输出 |

---

## 相关文档

- [SYSTEM_ARCHITECTURE.md](design/SYSTEM_ARCHITECTURE.md) - 系统架构
- [STYLE_DIFF_DESIGN.md](design/STYLE_DIFF_DESIGN.md) - 样式优化设计
- [BENCHMARK.md](design/BENCHMARK.md) - 性能基准

---

**文档版本**: v1.0
**最后更新**: 2026-02-01
**维护者**: Mint UI Team
