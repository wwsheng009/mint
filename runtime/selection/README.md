# Selection System

文本选择系统。支持鼠标和键盘选择，复制到剪贴板，以及选择高亮渲染。

## 职责

- **选择管理**: 跟踪选择状态和范围
- **多种选择模式**: 字符、单词、行级选择
- **事件处理**: 响应鼠标和键盘事件
- **文本提取**: 获取选中的文本
- **剪贴板集成**: 跨平台复制功能（Windows/macOS/Linux）
- **渲染支持**: 高亮显示选中的文本

## 核心概念

### 1. 选择模式

选择系统支持三种选择模式：

- **Char 模式**: 字符级选择（默认）
  - 单击拖动选择字符
  - 适用于精确选择

- **Word 模式**: 单词级选择
  - 鼠标双击选择完整单词
  - "单词"定义为连续的非空白字符

- **Line 模式**: 行级选择
  - 鼠标三击选择整行
  - 选择从行首到行尾

### 2. 选择状态

Manager 维护以下状态：

```go
type Manager struct {
    active    bool  // 是否有活动选择
    startX, startY int     // 选择起始位置
    currentX, currentY int  // 当前选择结束位置
    anchorX, anchorY int    // 选择锚点（用于扩展选择）
    mode      SelectionMode  // 当前选择模式
    buffer    TextBuffer     // 关联的文本缓冲区
}
```

### 3. 事件类型

支持多种事件触发选择操作：

**鼠标事件**:
- `MousePress`: 单击开始选择
- `MouseMotion`: 拖动更新选择
- `MouseRelease`: 释放结束选择
- `MouseWheelUp/Down`: 滚动
- 鼠标点击次数（单击/双击/三击）决定选择模式

**键盘快捷键**:
- `Ctrl+C`: 复制选中文本
- `Ctrl+X`: 剪切
- `Ctrl+A`: 全选
- `Escape`: 清除选择
- `Shift+方向键`: 扩展选择

### 4. 剪贴板支持

系统针对不同平台提供了剪贴板集成：

| 平台 | 复制方法 |
|------|---------|
| Windows | `Set-Clipboard` (PowerShell) |
| macOS | `pbcopy` |
| Linux (Wayland) | `wl-copy` |
| Linux (X11) | `xclip` / `xsel` |

## 使用示例

### 基本使用

```go
import "github.com/wwsheng009/mint/runtime/selection"

// 创建选择管理器
buffer := runtime.NewCellBuffer(80, 24)
manager := selection.NewManager(buffer)

// 开始选择（鼠标按下）
manager.Start(x, y)

// 更新选择（鼠标拖动）
manager.Update(newX, newY)

// 检查单元格是否被选中
if manager.IsSelected(x, y) {
    // 此单元格被选中
}

// 获取选中的文本
text := manager.GetSelectedText()

// 清除选择
manager.Clear()
```

### 单词和行选择

```go
// 单词选择（双击）
manager.SelectWord(x, y)

// 行选择（三击）
manager.SelectLine(y)

// 全选
manager.SelectAll()

// 检查选择模式
mode := manager.GetMode()
if mode == selection.SelectionModeWord {
    fmt.Println("单词选择模式")
}
```

### 扩展选择（Shift+点击）

```go
// 开始选择
manager.Start(startX, startY)

// 移动锚点到新位置（Shift+点击）
manager.Extend(newX, newY)

// 轴点会保持在原始位置，选择从锚点扩展到新位置
```

### 复制到剪贴板

```go
// 获取选中文本
text := manager.GetSelectedText()

// 复制到剪贴板（自动检测平台）
err := selection.CopyToClipboard(text)
if err != nil {
    fmt.Printf("复制失败: %v\n", err)
}

// 或使用紧凑格式（去除尾部空行）
textCompact := manager.GetSelectedTextCompact()
```

### 选择高亮渲染

```go
// 在渲染循环中
func render(buffer *runtime.CellBuffer, manager *selection.Manager) {
    // 1. 渲染所有组件
    renderComponents(buffer)

    // 2. 应用选择高亮
    applySelectionHighlight(buffer, manager)
}

func applySelectionHighlight(buffer *runtime.CellBuffer, manager *selection.Manager) {
    if !manager.IsActive() {
        return
    }

    // 获取所有选中单元格
    cells := manager.GetSelectedCells()
    for _, cell := range cells {
        // 反色显示
        original := buffer.GetCell(cell.X, cell.Y)
        buffer.SetContent(
            cell.X, cell.Y,
            original.Cluster[0],  // 简化处理
            reverseStyle(original.Style),
        )
    }
}
```

### 事件集成

```go
// 处理鼠标事件
func handleMouseEvent(ev MouseEvent, manager *selection.Manager) {
    switch ev.Action {
    case platform.MousePress:
        switch ev.ClickCount {
        case 2:
            // 双击：单词选择
            manager.SelectWord(ev.X, ev.Y)
        case 3:
            // 三击：行选择
            manager.SelectLine(ev.Y)
        default:
            // 单击：字符选择
            manager.Start(ev.X, ev.Y)
        }

    case platform.MouseMotion:
        if ev.Modifier&platform.ModShift != 0 {
            // Shift+拖动：扩展选择
            manager.Extend(ev.X, ev.Y)
        } else {
            // 普通拖动：更新选择
            manager.Update(ev.X, ev.Y)
        }

    case platform.MouseRelease:
        // 释放鼠标，选择完成
    }
}

// 处理键盘事件
func handleKeyEvent(key platform.RawInput, manager *selection.Manager) {
    switch key.Special {
    case platform.KeyC:
        if key.Modifiers&platform.ModCtrl != 0 {
            // Ctrl+C: 复制
            text := manager.GetSelectedText()
            selection.CopyToClipboard(text)
        }

    case platform.KeyA:
        if key.Modifiers&platform.ModCtrl != 0 {
            // Ctrl+A: 全选
            manager.SelectAll()
        }

    case platform.KeyEscape:
        // Escape: 清除选择
        manager.Clear()
    }
}
```

## 核心类型

### Manager

```go
type Manager struct {
    active    bool
    startX, startY int
    currentX, currentY int
    anchorX, anchorY int
    mode      SelectionMode
    buffer    TextBuffer
}

func NewManager(buffer TextBuffer) *Manager
func (m *Manager) Start(x, y int)
func (m *Manager) Update(x, y int)
func (m *Manager) Extend(x, y int)
func (m *Manager) Clear()
func (m *Manager) IsActive() bool
func (m *Manager) IsSelected(x, y int) bool
func (m *Manager) GetSelectedText() string
func (m *Manager) GetSelectedTextCompact() string
func (m *Manager) SelectWord(x, y int)
func (m *Manager) SelectLine(y int)
func (m *Manager) SelectAll()
```

### SelectionMode

```go
type SelectionMode int

const (
    SelectionModeChar SelectionMode = iota  // 字符选择
    SelectionModeWord                      // 单词选择
    SelectionModeLine                      // 行选择
)
```

### TextBuffer

```go
type TextBuffer interface {
    GetCell(x, y int) Cell
    Width() int
    Height() int
}
```

### Cell

```go
type Cell struct {
    Cluster string
    Empty   bool
}
```

### SelectionRegion

```go
type SelectionRegion struct {
    StartX, EndX int
    StartY, EndY int
}

func (r SelectionRegion) Contains(x, y int) bool
func (r SelectionRegion) IsEmpty() bool
func (r SelectionRegion) Width() int
func (r SelectionRegion) Height() int
```

### 剪贴板函数

```go
// 复制到剪贴板
func CopyToClipboard(text string) error

// 从剪贴板读取
func ReadFromClipboard() (string, error)

// 检查平台是否支持剪贴板操作
func IsClipboardSupported() bool
```

## 文件结构

- `selection.go` - 核心选择管理器和类型
- `clipboard.go` - 剪贴板集成（跨平台）
- `render.go` - 选择高亮渲染
- `mouse_handler.go` - 鼠标事件处理
- `keyboard.go` - 键盘事件处理
- `integration.go` - 统一 API 接口
- `adapter.go` - Runtime 适配器
- `selection_test.go` - 单元测试

## 文档

更详细的文档：
- `IMPLEMENTATION_SUMMARY.md` - 实现总结和快速开始
- `INTEGRATION.md` - 集成指南
- `INTEGRATION_EXAMPLE.md` - 集成示例
- `INTEGRATION_STEPS.md` - 集成步骤

## 依赖

**可以依赖**:
- `runtime/paint` - CellBuffer 类型
- 标准库: `strconv`, `os/exec`, `runtime`, `strings`, `unicode/utf8`

**不能依赖**:
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 平台兼容性

| 平台 | 选择功能 | 复制功能 | 依赖 |
|------|---------|---------|------|
| Windows | ✅ | ✅ | PowerShell |
| macOS | ✅ | ✅ | pbcopy/pbpaste |
| Linux (Wayland) | ✅ | ✅ | wl-copy/wl-paste |
| Linux (X11) | ✅ | ✅ | xclip 或 xsel |

## 最佳实践

### 1. 渲染顺序

`ApplySelection()` 必须在所有组件渲染后调用：

```go
renderComponents(buffer)   // 先渲染组件
applySelectionHighlight(buffer, manager)  // 再应用选择高亮
```

### 2. 事件优先级

选择事件应在组件处理前检查：

```go
func handleEvent(ev interface{}) {
    // 检查是否是选择相关事件
    if isSelectionEvent(ev) {
        manager.HandleSelectionEvent(ev)
        return  // 如果是选择事件，跳过组件处理
    }

    // 处理组件事件
    component.HandleEvent(ev)
}
```

### 3. 清除选择

当发生某些操作时，自动清除选择：

```go
// 用户开始输入时
func onTextInput(text string) {
    manager.Clear()
    // 处理输入...
}

// 窗口大小改变时
func onResize(width, height int) {
    manager.Clear()
    // 重新布局...
}
```

### 4. 错误处理

复制操作可能失败：

```go
text := manager.GetSelectedText()
if err := selection.CopyToClipboard(text); err != nil {
    // 显示错误消息
    showErrorToUser("复制失败: " + err.Error())
}
```

## 常见问题

### Q: 如何检测是否支持剪贴板？

A: 使用 `IsClipboardSupported()`:

```go
if !selection.IsClipboardSupported() {
    fmt.Println("当前平台不支持剪贴板操作")
}
```

### Q: 为什么 Linux 上复制失败？

A: Linux 需要安装剪贴板工具：
- X11: `sudo apt install xclip` 或 `sudo apt install xsel`
- Wayland: `sudo apt install wl-clipboard`

### Q: 选择高亮颜色如何自定义？

A: 使用不同的渲染函数自定义高亮样式：

```go
func applyCustomHighlight(buffer *runtime.CellBuffer, manager *selection.Manager) {
    cells := manager.GetSelectedCells()
    for _, cell := range cells {
        original := buffer.GetCell(cell.X, cell.Y)
        // 自定义样式，例如蓝色背景
        buffer.SetContent(
            cell.X, cell.Y,
            original.Cluster[0],
            customStyle,
        )
    }
}
```

### Q: 支持列选择吗？

A: 当前版本不支持列选择（Alt+拖动），但可以通过扩展 `SelectionMode` 实现。

### Q: 如何获取选择区域的详细信息？

A: 使用 `GetRegion()`:

```go
region := manager.GetRegion()
fmt.Printf("选择区域: (%d, %d) -> (%d, %d)\n",
    region.StartX, region.StartY,
    region.EndX, region.EndY)
fmt.Printf("宽度: %d, 高度: %d\n",
    region.Width(), region.Height())
```
