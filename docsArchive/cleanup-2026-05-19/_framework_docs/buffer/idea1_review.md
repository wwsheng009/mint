# idea1.md 审查报告

## 审查日期
2026-01-29

## 审查范围
根据 `idea1.md` 文档建议，审查 `mint` TUI 框架当前实施情况。

---

## 一、已完全实现 (⭐⭐⭐⭐⭐)

| 文档建议 | 当前实现 | 代码位置 |
|---------|---------|---------|
| **Grapheme Cluster 支持** | `Cell.Cluster` 使用 `string` 类型 | `runtime/paint/cell.go:13` |
| 使用 `uniseg` 库 | `uniseg.NewGraphemes()` 处理字形簇 | `runtime/paint/buffer.go:104-121` |
| **宽字符清理 (clearCellAt)** | 正确清理 head + continuation | `runtime/paint/buffer.go:219-241` |
| **Style 差分编码** | `StyleStateMachine` 最小化 VT 码变化 | `runtime/paint/style_state.go:9-212` |
| **命令合并 (Run Builder)** | `CommandBatch.mergeCommands()` 合并相邻命令 | `runtime/paint/batch.go:86-125` |
| **Dirty Region 跟踪** | `DirtyTracker` + flood fill 区域提取 | `runtime/paint/dirty.go:1-503` |
| **Layer/ZIndex 系统** | `Layer` + `Compositor` 多层合成 | `runtime/paint/layer.go`, `compositor.go` |
| **两阶段布局** | Measure + Layout 分离 | `runtime/layout/engine.go:39-80` |
| **优先级调度** | High/Normal/Low 三级优先级 | `runtime/scheduler/scheduler.go` |

---

## 二、部分实现 / 需要改进 (⭐⭐⭐⭐)

### 2.1 runeWidth 混用问题

**问题**: `SetString` 正确使用 `runewidth.StringWidth`，但 `runeWidth()` 函数仍用手写 CJK 判断

**当前代码** (`buffer.go:74-93`):
```go
func runeWidth(r rune) int {
    // 手写的 CJK 范围判断...
    if r >= 0x1100 && ... {
        return 2
    }
    // ...
}
```

**问题影响**:
- 无法正确处理 Nerd Font 图标
- 无法处理特殊 Unicode 字符
- 与 `runewidth.StringWidth` 行为不一致

**修复建议**: 删除手写函数，统一使用 `runewidth.RuneWidth`

---

### 2.2 Cursor 优化不完整

**当前实现** (`batch.go:138-140`):
```go
func (b *CommandBatch) moveCursor(x, y int) string {
    return "\x1b[" + itoa(y+1) + ";" + itoa(x+1) + "H"
}
```

**文档建议**:
| 场景 | 用法 |
|-----|------|
| 同一行右移 | 直接输出字符 |
| 同行小步移动 | `\x1b[nC` |
| 下一行开头 | `\n` |

**修复建议**: 添加光标位置跟踪和优化逻辑

---

### 2.3 IsCellChanged 逻辑

**当前实现** (`buffer.go:184-198`):
```go
func IsCellChanged(cell, prevCell Cell) bool {
    if cell.IsContinuation {
        return false  // 总是跳过
    }
    if prevCell.IsContinuation {
        return cell.Style != prevCell.Style
    }
    return cell.Cluster != prevCell.Cluster || cell.Style != prevCell.Style
}
```

**文档建议的表格**:
| 情况 | 是否刷新 |
|-----|---------|
| continuation → continuation | ❌ |
| head → continuation | ✅ |
| continuation → head | ✅ |

**当前问题**: `head → continuation` 变化时可能漏刷

---

## 三、未实现 (⭐⭐)

### 3.1 双缓冲系统

**文档建议**:
```go
type Renderer struct {
    front *Buffer  // 当前屏幕状态
    back  *Buffer  // 新一帧
}
```

**当前状态**: `Compositor` 有多层合成，但缺少明确的 front/back buffer 结构和交换机制

---

### 3.2 完整 Frame Scheduler

**文档建议的结构**:
```go
type Engine struct {
    renderer      *Renderer
    scheduler     *Scheduler
    frameInterval time.Duration  // 16ms = 60FPS
    eventQueue    chan Event
    quit          chan struct{}
}
```

**当前状态**:
- `runtime/scheduler/scheduler.go` 存在，但只实现了 **Update Scheduler**
- 缺少主事件循环 (`ticker` + `select`)
- 缺少帧同步机制
- 缺少 Idle Detection

**scheduler.go 当前功能**:
- ✅ Update batching (批处理)
- ✅ Priority processing (优先级)
- ✅ Time slicing (时间分片)
- ✅ Dirty queue (脏队列)
- ❌ 主事件循环
- ❌ 帧 ticker 驱动
- ❌ Idle detection

---

### 3.3 事件分发顺序

**文档建议** (Layer System):
```go
// 从 ZIndex 最大的 Layer 往下找
for i := len(layers)-1; i >= 0; i-- {
    if layers[i].Root.HandleEvent(ev) {
        return
    }
}
```

**当前状态**: Layer 系统存在，但事件未按 ZIndex 反向分发

---

## 四、整体完成度评估

| 模块 | 完成度 | 说明 |
|-----|-------|------|
| Buffer (Grapheme Cluster) | ⭐⭐⭐⭐⭐ | 完整实现 |
| 宽字符处理 | ⭐⭐⭐⭐ | 需统一 runewidth |
| clearCellAt | ⭐⭐⭐⭐⭐ | 正确实现 |
| Layer/ZIndex | ⭐⭐⭐⭐⭐ | 完整实现 |
| Dirty Region | ⭐⭐⭐⭐⭐ | 带 flood fill 合并 |
| Style 差分 | ⭐⭐⭐⭐⭐ | 完整的状态机 |
| Run 合并 | ⭐⭐⭐⭐ | 基础实现 |
| 双缓冲 | ⭐⭐ | 概念存在，结构未明确 |
| Update Scheduler | ⭐⭐⭐⭐ | 完整实现 |
| Frame Scheduler | ⭐⭐ | 只有 Update 部分 |
| Cursor 优化 | ⭐⭐⭐ | 基础实现 |

---

## 五、关键发现

### 5.1 架构优势

当前系统已经实现了 TUI 引擎的核心组件，代码质量高：

1. **正确的 Grapheme Cluster 处理** - 使用 `uniseg` 处理复杂 Unicode
2. **完善的脏区域系统** - 包含 flood fill 区域提取和合并
3. **样式状态机** - 最小化 ANSI 输出
4. **层级系统** - 支持多层合成

### 5.2 主要差距

要达到文档描述的 "实时渲染系统" 级别，主要缺失：

1. **Frame Engine 主循环** - 需要构建带 ticker 的事件驱动引擎
2. **双缓冲 Diff 渲染** - front/back buffer 交换机制
3. **Idle Detection** - 空闲时停止渲染以降低 CPU 占用

---

## 六、文件清单

### 已实现的关键文件

| 文件 | 功能 | 对应文档章节 |
|-----|------|-------------|
| `runtime/paint/cell.go` | Cell 结构 (Grapheme Cluster) | Grapheme Cluster |
| `runtime/paint/buffer.go` | Buffer + 宽字符处理 | Buffer + clearCellAt |
| `runtime/paint/style_state.go` | Style 差分状态机 | ANSI Encoder |
| `runtime/paint/batch.go` | 命令批处理和合并 | Run Builder |
| `runtime/paint/dirty.go` | 脏区域跟踪和合并 | Dirty Region |
| `runtime/paint/layer.go` | Layer 抽象 | Layer System |
| `runtime/paint/compositor.go` | 多层合成 | Layer 合成 |
| `runtime/layout/engine.go` | 两阶段布局 | Layout Engine |
| `runtime/scheduler/scheduler.go` | 更新调度器 | (部分) Frame Scheduler |

---

## 七、总结

当前系统的 **核心渲染组件** 已经达到文档描述的高级水平，主要缺失的是：

1. **帧调度引擎** - 需要构建主循环来驱动整个渲染管线
2. **双缓冲结构** - 需要明确的 front/back buffer 和交换机制
3. **光标优化** - 需要添加位置跟踪和移动优化

这些是将引擎从 "组件级" 提升到 "实时渲染系统级" 的关键拼图。
