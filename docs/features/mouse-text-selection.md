# 鼠标文本选择功能

## 功能概述

Mint TUI 框架支持通过鼠标进行文本选择，用户可以通过按住鼠标左键并拖动来选择屏幕上的文本区域，选中的文本会以反色（reverse video）高亮显示。

## 使用方式

### 基本操作

| 操作 | 行为 |
|------|------|
| **按住左键 + 拖动** | 立即进入选择模式，实时显示选择区域（反色高亮） |
| **释放左键** | 结束选择，保持选择区域显示 |
| **点击（不拖动）** | 视为普通点击，清除之前的选择区域 |
| **按任意键** | 清除选择区域 |
| **Ctrl+C** | 复制选中的文本到剪贴板 |
| **双击** | 选中光标位置的单词 |
| **三击** | 选中光标所在整行 |
| **Shift+点击** | 扩展选择范围到点击位置 |

### 交互流程

```
鼠标左键按下
    │
    ▼
开始计时，记录起始位置 (startX, startY)
    │
    ▼
鼠标移动？ ──否──► 继续等待
    │
    是
    ▼
立即进入选择模式 (enabled = true, isSelecting = true)
    │
    ▼
实时更新选择区域 (endX, endY)，每帧渲染反色高亮
    │
    ▼
鼠标左键释放
    │
    ▼
结束拖拽 (isSelecting = false)，保持选择显示 (hasSelection = true)
```

## 技术实现

### 核心数据结构

```go
// SelectionState 文本选择模式状态
type SelectionState struct {
    enabled          bool      // 是否启用选择模式
    startX, startY   int       // 选择起始位置
    endX, endY       int       // 选择结束位置
    isSelecting      bool      // 是否正在拖拽选择中
    isLeftButtonDown bool      // 左键是否按下
    hasSelection     bool      // 是否有有效的选择区域（释放后保持）
    selectStartTime  time.Time // 选择开始时间
}
```

### 关键方法

#### 1. 鼠标事件处理 (`handleMouseEvent`)

**鼠标按下 (MousePress)**：
- 如果已有选择区域，先清除它
- 记录起始位置和时间
- 重置选择状态

**鼠标移动 (MouseMove)**：
- 检查是否按住左键且未进入选择模式
- 只要有任何移动 (`dx != 0 || dy != 0`)，**立即**进入选择模式
- 如果已在选择模式，实时更新结束位置

**鼠标释放 (MouseRelease)**：
- 检查是否移动过（区分点击和选择）
- 如果移动过，保持选择区域显示
- 如果没有移动，视为普通点击

#### 2. 选择高亮渲染 (`applySelectionHighlight`)

```go
func (e *Engine) applySelectionHighlight(buf *paint.Buffer) {
    // 只要选择模式启用就渲染
    if !e.selectionState.enabled {
        return
    }
    
    // 计算选择区域边界
    // 规范化坐标（确保 start 在左上角，end 在右下角）
    
    // 遍历选择区域内的所有单元格
    for y := startY; y <= endY; y++ {
        for x := lineStart; x <= lineEnd; x++ {
            // 直接修改 cell 的 Style，添加反色效果
            cell := buf.Cells[y][x]
            cell.Style = cell.Style.Reverse(true)
            buf.Cells[y][x] = cell
        }
    }
}
```

#### 3. 渲染流程集成

在 `frame()` 方法中，绘制完成后应用选择高亮：

```go
func (e *Engine) frame() {
    // 1. 更新组件状态
    // 2. 布局
    // 3. 绘制到 back buffer
    buf := e.renderer.GetBackBuffer()
    root.Paint(buf)
    
    // 4. 应用文本选择高亮
    e.applySelectionHighlight(buf)
    
    // 5. 渲染输出
    output := e.renderer.Render()
    // ...
}
```

### 与其他功能的协调

#### 与鼠标点击的区分

通过 `hasMoved` 判断是点击还是选择：

```go
dx := mouseEv.X - e.selectionState.startX
dy := mouseEv.Y - e.selectionState.startY
hasMoved := dx != 0 || dy != 0

if !wasSelecting || !hasMoved {
    // 没有移动，视为普通点击
} else {
    // 有移动，保持选择区域
}
```

#### 与键盘事件的协调

键盘事件会清除选择区域：

```go
if ev.Type() == event.EventKeyPress && e.selectionState.hasSelection {
    e.ClearSelection()
}
```

#### 与组件点击的协调

只有在非选择模式下才分发点击事件给组件：

```go
if !e.selectionState.enabled {
    result := event.DispatchEvent(ev, boxes)
    // ...
}
```

## 开发历程

### 问题 1：构建错误（跨平台兼容性）

**问题**：`helper.go` 中直接导入 `golang.org/x/sys/windows`，导致在 Linux/macOS 上编译失败。

**解决**：将 Windows 特定的代码移到 `helper_windows.go`，使用 build tags 隔离平台相关代码。

### 问题 2：选择高亮不显示

**问题**：使用 `buf.SetSelected(x, y, true)` 设置 `Selected` 字段，但 Renderer 的 `emitRun` 方法不检查这个字段。

**解决**：改为直接修改 cell 的 `Style`，使用 `cell.Style.Reverse(true)` 添加反色效果。

### 问题 3：Diff 算法不检测选择变化

**问题**：`IsCellChanged` 函数没有比较 `Selected` 字段，导致选择变化不会触发重绘。

**解决**：修改 `IsCellChanged` 函数，添加 `cell.Selected != prevCell.Selected` 比较。

### 问题 4：选择区域无法保持

**问题**：鼠标释放后立即清除选择区域，用户无法看到最终选择结果。

**解决**：添加 `hasSelection` 字段，释放后保持 `enabled = true`，直到下次点击或按键才清除。

### 问题 5：选择响应延迟

**问题**：需要按住 300ms 才进入选择模式，用户体验不佳。

**解决**：改为按住并**立即拖动**就进入选择模式，无需等待时间阈值。

### 问题 6：拖动时不显示选择区域

**问题**：`applySelectionHighlight` 要求 `hasSelection` 为 true，但该字段只在释放时设置。

**解决**：修改条件为只要 `enabled` 为 true 就渲染选择高亮。

## 关键设计决策

### 1. 立即响应拖动

不设置时间阈值，只要用户按住并拖动就立即进入选择模式，提供最流畅的体验。

### 2. 直接修改 Style

不引入额外的 `Selected` 字段，直接修改 `cell.Style` 添加反色，确保与现有渲染系统兼容。

### 3. 保持选择区域

释放鼠标后保持选择区域显示，方便用户查看和复制，直到进行下一次交互才清除。

### 4. 区分点击和选择

通过 `hasMoved` 判断用户意图，没有移动视为点击，有移动视为选择。

## 文件位置

- **核心实现**：`runtime/engine/engine.go`
  - `SelectionState` 结构体
  - `handleMouseEvent()` 方法
  - `applySelectionHighlight()` 方法
  - `frame()` 方法中的集成

- **Diff 检测**：`runtime/paint/buffer.go`
  - `IsCellChanged()` 函数

## 未来改进

1. **复制功能**：添加 Ctrl+C 复制选中文本到剪贴板
2. **多行选择优化**：支持跨行选择的文本提取
3. **双击选择**：双击选中单词，三击选中整行
4. **选择扩展**：按住 Shift + 点击扩展选择范围
