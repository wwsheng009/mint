# Platform (V3)

平台抽象层。

## 职责

- **平台接口**：定义 `RuntimePlatform` 接口，为 Runtime 提供统一接口
- **屏幕管理**：屏幕输出、备用屏幕、清屏功能
- **输入处理**：读取原始输入（RawInput），不涉及语义转换
- **信号处理**：处理终端信号（SIGINT, SIGWINCH 等）
- **光标控制**：光标隐藏/显示、定位等功能
- **终端控制**：终端模式切换、恢复

## 设计原则

Platform 层只提供**能力抽象**，不包含**语义**：
- ✅ 提供：`ReadInput()` → 返回 `RawInput`
- ❌ 不提供：`ReadKey()` → 返回 `Action`（语义转换由 Runtime 的 KeyMap 负责）
- ✅ 提供：`WriteString()` 和 `Clear()`
- ❌ 不提供：`RenderComponent()`（渲染由上层负责）

Platform 不应该知道：Focus、Event、Component、Layout

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 核心概念

### RuntimePlatform

`RuntimePlatform` 将所有底层能力组合在一起，避免接口方法冲突：

```go
type RuntimePlatform interface {
    Init() error
    Close() error
    Size() (width, height int)
    ReadInput() *RawInput
    WriteString(s string) (int, error)
    Clear() error
}
```

### RawInput

`RawInput` 是平台无关的原始输入表示，`InputReader` 只产生原始输入：

```go
type RawInput struct {
    Type       RawInputType    // 输入类型
    Key        rune            // 键盘字符
    Special    SpecialKey      // 特殊键（Tab, Enter 等）
    Modifiers  KeyModifier     // 修饰键（Shift, Alt, Ctrl）
    MouseX     int             // 鼠标 X 坐标
    MouseY     int             // 鼠标 Y 坐标
    MouseButton MouseButton    // 鼠标按钮
    MouseAction MouseAction    // 鼠标动作
    Width      int             // 窗口宽度（Resize）
    Height     int             // 窗口高度（Resize）
    Data       []byte          // 其他数据
    Timestamp  time.Time       // 时间戳
}
```

**重要**：`RawInput` 不是 `Action`！Platform 只产生原始输入，不进行语义转换。

### 输入类型

| 类型 | 说明 |
|------|------|
| `InputKeyPress` | 按键按下 |
| `InputKeyRelease` | 按键释放 |
| `InputMouse` | 鼠标事件 |
| `InputResize` | 窗口大小变化 |
| `InputPaste` | 剪贴板粘贴 |
| `InputSignal` | 系统信号 |

### SpecialKey

```go
const (
    KeyEscape      SpecialKey
    KeyEnter       SpecialKey
    KeyTab         SpecialKey
    KeyBackspace   SpecialKey
    KeyDelete      SpecialKey
    // 光标键
    KeyUp, KeyDown, KeyLeft, KeyRight
    KeyHome, KeyEnd
    KeyPageUp, KeyPageDown
    // 功能键
    KeyF1, KeyF2, ..., KeyF12
)
```

### KeyModifier

```go
const (
    ModShift KeyModifier = 1 << iota
    ModAlt
    ModCtrl
    ModMeta
)
```

### 多层防御的恢复机制

`RestoreTerminal()` 是应用层的恢复机制第 2 层：

1. **第 1 层（Engine 层）**：`Engine.Run()` 的 `defer cleanup()` → `inputReader.Stop()` → 恢复 `originalMode`
2. **第 2 层（应用层）**：`main()` 的 `defer RestoreTerminal()` → 强制恢复到安全模式（保险）
3. **第 3 层（进程层）**：`init()` 的信号处理 → Ctrl+C 时强制恢复（兜底）

终端模式污染是致命问题，多层防御确保即使某一层失败，其他层仍能恢复终端。

## 使用示例

### 使用 RuntimePlatform

```go
import "github.com/wwsheng009/mint/runtime/platform"

// 创建平台
plat, err := platform.NewDefaultPlatform()
if err != nil {
    log.Fatal(err)
}

// 初始化平台
if err := plat.Init(); err != nil {
    log.Fatal(err)
}

defer plat.Close()
```

### 读取输入

```go
// 方式 1：同步读取
for {
    input := plat.ReadInput()
    processRawInput(input)
}

// 方式 2：异步读取（使用 InputReader）
reader, err := platform.NewInputReader()
if err != nil {
    log.Fatal(err)
}

events := make(chan platform.RawInput, 100)
reader.Start(events)
defer reader.Stop()

for event := range events {
    processRawInput(event)
}
```

### 写入屏幕

```go
// 写入字符串
plat.WriteString("Hello, World!")

// 清屏
plat.Clear()

// 获取屏幕尺寸
width, height := plat.Size()
fmt.Printf("屏幕大小: %dx%d\n", width, height)
```

### 备用屏幕模式

备用屏幕允许 TUI 应用在不影响原有终端内容的情况下运行：

```go
import "github.com/wwsheng009/mint/runtime/platform"

plat, _ := platform.NewDefaultPlatform()

// 进入备用屏幕
plat.DefaultScreen.EnterAlternateScreen()

defer plat.DefaultScreen.ExitAlternateScreen()

// 运行 TUI 应用
RunTUIApp(plat)
```

### 处理原始输入

```go
func processRawInput(input *platform.RawInput) {
    switch input.Type {
    case platform.InputKeyPress:
        if input.Special != platform.KeyUnknown {
            fmt.Printf("特殊键: %s\n", input.Special)
            // 转换为 Action
            action := convertSpecialKeyToAction(input.Special)
        } else if input.Key > 0 {
            fmt.Printf("字符: %c\n", input.Key)
            action := action.NewAction(action.ActionInputChar).WithPayload(input.Key)
        }
    case platform.InputMouse:
        fmt.Printf("鼠标: (%d, %d), 按钮: %v\n", input.MouseX, input.MouseY, input.MouseButton)
    case platform.InputResize:
        fmt.Printf("窗口大小变化: %dx%d\n", input.Width, input.Height)
    }
}
```

### 恢复终端（多层防御）

```go
func main() {
    // 第 2 层防御：应用层恢复终端
    defer platform.RestoreTerminal()

    plat, _ := platform.NewDefaultPlatform()
    plat.Init()
    defer plat.Close()

    // 运行 TUI 应用
    RunApp(plat)
}
```

## 核心类型

| 类型 | 说明 |
|------|------|
| `RuntimePlatform` | 平台统一接口 |
| `DefaultPlatform` | 默认平台实现 |
| `Screen` | 屏幕输出接口 |
| `DefaultScreen` | 默认屏幕实现 |
| `InputReader` | 输入读取接口 |
| `RawInput` | 原始输入类型 |
| `RawInputType` | 输入类型枚举 |
| `SpecialKey` | 特殊键枚举 |
| `KeyModifier` | 修饰键枚举 |
| `MouseButton` | 鼠标按钮枚举 |
| `MouseAction` | 鼠标动作枚举 |

## 文件结构

| 文件 | 说明 |
|------|------|
| `platform.go` | RuntimePlatform 接口和 DefaultPlatform 实现 |
| `screen.go` | Screen 接口和 DefaultScreen 实现 |
| `input.go` | InputReader 接口和 RawInput 类型 |
| `cursor.go` | 光标控制（如果有） |
| `terminal.go` | 终端控制（如果有） |
| `signal.go` | 信号处理器（如果有） |
| `input_unix.go` | Unix 平台输入实现（build tag） |
| `input_windows.go` | Windows 平台输入实现（build tag） |

## 平台特定实现

使用 Go 的 build tags 选择正确的平台实现：

```go
// input_unix.go
// +build unix

package platform

func newInputReaderImpl() inputReaderImpl {
    return &unixInputReader{}
}
```

```go
// input_windows.go
// +build windows

package platform

func newInputReaderImpl() inputReaderImpl {
    return &windowsInputReader{}
}
```

## 最佳实践

### 1. 始终在 main() 函数中恢复终端

```go
func main() {
    // 第 2 层防御
    defer platform.RestoreTerminal()

    // ...
}
```

### 2. 使用备用屏幕模式

```go
plat.DefaultScreen.EnterAlternateScreen()
defer plat.DefaultScreen.ExitAlternateScreen()
```

### 3. 异步读取输入

```go
reader, _ := platform.NewInputReader()
events := make(chan platform.RawInput, 100)
reader.Start(events)
defer reader.Stop()

for event := range events {
    // 处理事件
}
```

### 4. 错误处理

```go
if err := plat.Init(); err != nil {
    platform.RestoreTerminal() // 失败时恢复终端
    log.Fatal(err)
}
```

## 与其他模块集成

### 与 Input（KeyMap）集成

```go
// Platform 产生 RawInput
rawInput := plat.ReadInput()

// KeyMap 转换为 Action
action := keyMap.Convert(rawInput)

// Dispatcher 分发 Action
dispatcher.Dispatch(action)
```

### 与 Runtime 集成

Runtime 通过 `RuntimePlatform` 接口与 Platform 交互：

```go
type Engine struct {
    platform platform.RuntimePlatform
}

func (e *Engine) Run() error {
    e.platform.Init()
    defer e.platform.Close()

    // 主循环
    for {
        input := e.platform.ReadInput()
        // ...
    }
}
```
