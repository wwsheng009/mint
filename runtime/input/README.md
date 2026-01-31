# Input Processing

输入处理核心层。负责将 Platform 提供的原始输入（RawInput）转换为语义化的 Action。

## 职责

- **输入解析**: 将 Platform 层的 RawInput 解析为结构化数据
- **按键映射**: 将原始按键转换为语义化的 Action 类型（KeyMap）
- **鼠标跟踪**: 跟踪鼠标状态，支持双击、三击、拖拽等高级操作（MouseTracker）
- **输入读取**: 提供可取消的输入读取能力（Reader、TerminalInput）
- **配置加载**: 支持 YAML 格式的按键映射配置

## 核心概念

### 1. RawInput → Action 转换流程

```
Platform RawInput → KeyMap → Action → Dispatcher → Component Handler
```

**输入处理职责划分**:
- **Platform**: 提供原始输入数据（RawInput），只提供"能力"而非"语义"
- **Runtime/KeyMap**: 将 RawInput 转换为语义化的 Action，这是 Runtime 的职责
- **Framework/Component**: 处理语义化的 Action，不应该直接处理 RawInput

### 2. KeyMap 按键映射系统

KeyMap 负责将 Platform 的特殊键和字符输入映射为 Action 类型。

**映射优先级**:
1. 自定义绑定 (`customBinds`)
2. 上下文相关映射 (`contextMaps`)
3. 默认映射 (`defaultMap`)

**上下文支持**:
- 支持上下文栈 (`contextStack`)，不同上下文拥有不同的按键映射
- 使用 `PushContext()`/`PopContext()` 管理上下文
- 适用于模态对话框、菜单等场景

**默认按键映射**:
- 导航: `↑↓←→`、`Home`、`End`、`PageUp/PageDown`
- Vim 风格: `k/j/h/l`
- 编辑: `Backspace`、`Delete`、`Enter`
- 功能键: `Esc` (取消)、`Tab` (下一焦点)、`F1` (帮助)、`F5` (刷新)

### 3. MouseTracker 鼠标跟踪

MouseTracker 追踪鼠标状态以支持高级操作：

**支持的鼠标操作**:
- 单击: `ActionMouseClick`
- 双击: `ActionMouseDoubleClick` (500ms 内连续点击)
- 三击: `ActionMouseTripleClick`
- 拖拽: `ActionMouseDrag` (包含起始位置和偏移量)
- 滚轮: `ActionMouseWheelUp/Down`
- 移动: `ActionMouseMotion`

**配置参数**:
- `doubleClickTimeout`: 双击超时（默认 500ms）
- `doubleClickDistance`: 双击位置容差（默认 5 像素）

### 4. Reader 可取消输入读取器

提供跨平台的可取消输入读取能力：

**核心接口**:
```go
type Reader interface {
    io.Reader
    Cancel() error   // 取消正在进行的读取操作
    Close() error    // 关闭读取器
}
```

**特性**:
- 支持 `context.Context` 超时控制
- 线程安全的取消机制
- 支持 `ReadEvent()` 带超时读取
- 提供 `ReadLine()` 和 `ReadLineWithTimeout()` 辅助函数

## 使用示例

### 基本按键映射

```go
// 创建 KeyMap
keymap := input.NewKeyMap()

// 转换 RawInput 为 Action
rawInput := platform.RawInput{
    Type:    platform.InputKeyPress,
    Special: platform.KeyEnter,
}
action := keymap.Map(rawInput)
// action.Type == action.ActionSubmit
```

### 自定义按键绑定

```go
keymap := input.NewKeyMap()

// 绑定 Ctrl+C 到退出操作
keymap.Bind("C-c", action.ActionQuit)

// 绑定 Ctrl+S 到保存操作
keymap.Bind("C-s", action.ActionSave)

// 取消绑定
keymap.Unbind("C-c")
```

### 上下文相关映射

```go
keymap := input.NewKeyMap()

// 设置模态对话框上下文
keymap.BindContext("modal", map[platform.SpecialKey]action.ActionType{
    platform.KeyEnter: action.ActionSubmit,
    platform.KeyEscape: action.ActionCancel,
})

// 进入模态状态
keymap.PushContext("modal")

// 此时 Enter 会触发 ActionSubmit
// ...

// 退出模态状态
keymap.PopContext()
```

### 鼠标事件处理

```go
tracker := input.NewMouseTracker()

// 处理鼠标输入
rawInput := platform.RawInput{
    Type:       platform.InputMouse,
    MouseAction: platform.MousePress,
    MouseButton: platform.MouseLeft,
    MouseX:     10,
    MouseY:     20,
}

result := tracker.ProcessInput(rawInput)

if result != nil {
    if result.IsDoubleClick {
        fmt.Println("双击事件")
    }
    if result.IsDragStart {
        fmt.Printf("拖拽开始: (%d, %d)\n",
            result.DragStartX, result.DragStartY)
    }
    if result.Action != nil {
        // 处理 Action
        dispatchAction(result.Action)
    }
}
```

### 从 YAML 加载配置

```go
// 从文件加载
keymap, err := input.LoadKeyMap("config/keymap.yaml")
if err != nil {
    // 使用默认 keymap
    keymap = input.NewKeyMap()
}

// 从字符串加载
yamlStr := `
navigation:
  up: "Up"
  down: "Down"
  left: "Left"
  right: "Right"

editing:
  submit: "Enter"
  cancel: "Escape"
`
keymap, err := input.LoadKeyMapFromString(yamlStr)
```

### 可取消的输入读取

```go
// 创建终端输入
termInput, err := input.NewTerminalInput()
if err != nil {
    panic(err)
}
defer termInput.Close()

// 带超时读取事件
data, err := termInput.ReadEvent(1 * time.Second)
if err != nil {
    if err == input.ErrTimeout {
        fmt.Println("读取超时")
    }
}

// 取消读取
go func() {
    time.Sleep(100 * time.Millisecond)
    termInput.Cancel()
}()

// 立即取消后，Read 会返回 ErrCanceled
```

## 核心类型

### KeyMap

```go
type KeyMap struct {
    defaultMap   map[platform.SpecialKey]action.ActionType
    contextMaps  map[string]map[platform.SpecialKey]action.ActionType
    contextStack []string
    customBinds  map[string]action.ActionType
}
```

**主要方法**:
- `Map(input platform.RawInput) *action.Action`: 将 RawInput 转换为 Action
- `Bind(combo string, actionType action.ActionType)`: 自定义绑定
- `Unbind(combo string)`: 解除绑定
- `BindContext(context string, bindings map)`: 绑定上下文
- `PushContext(context string)`: 推入上下文
- `PopContext()`: 弹出上下文

### MouseTracker

```go
type MouseTracker struct {
    lastClickButton      platform.MouseButton
    lastClickTime        time.Time
    lastClickX, lastClickY int
    lastClickCount       int
    isDragging           bool
    dragStartX, dragStartY int
    dragButton           platform.MouseButton
    doubleClickTimeout   time.Duration
    doubleClickDistance  int
}
```

**主要方法**:
- `ProcessInput(input platform.RawInput) *MouseEvent`: 处理鼠标输入

### MouseEvent

```go
type MouseEvent struct {
    Action       *action.Action
    IsDoubleClick bool
    IsTripleClick bool
    IsDragStart   bool
    IsDragMove    bool
    IsDragEnd     bool
    DragStartX, DragStartY int
    DragDeltaX, DragDeltaY int
}
```

### Reader / CancelReader

```go
type Reader interface {
    io.Reader
    Cancel() error
    Close() error
}

type CancelReader struct {
    r          io.Reader
    cancel      func()
    cancelChan  chan struct{}
    closed      bool
    mu          sync.Mutex
}
```

### TerminalInput

```go
type TerminalInput struct {
    reader Reader
    ctx    context.Context
    cancel context.CancelFunc
    mu     sync.Mutex
}
```

**主要方法**:
- `Read(p []byte) (n int, err error)`: 读取（可取消）
- `ReadEvent(timeout time.Duration) ([]byte, error)`: 带超时读取
- `Cancel()`: 取消读取
- `Close() error`: 关闭输入
- `Context() context.Context`: 获取上下文

## 文件结构

- `keymap.go` - KeyMap 实现和按键映射
- `keymap_config.go` - YAML 配置加载器
- `mouse_tracker.go` - 鼠标状态跟踪器
- `reader.go` - 可取消输入读取器

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

但是可以依赖：
- `runtime/platform` - 平台抽象层
- `runtime/action` - Action 类型定义
- `gopkg.in/yaml.v3` - YAML 配置解析（外部配置文件工具）

## 与其他模块集成

### 与 Platform 集成

```
Platform.RawInput → KeyMap.Map() → Action
```

KeyMap 是 Platform 和 Runtime 之间的桥梁，Platform 只提供原始输入，KeyMap 负责语义化转换。

### 与 Action 集成

KeyMap 生成的 Action 被传递给 Dispatcher 进行分发：
```go
action := keymap.Map(platformInput)
if action != nil {
    dispatcher.Dispatch(action)
}
```

### 与 DevTools 集成

KeyMap 可以记录按键映射配置，用于调试：
```go
// 查看当前所有自定义绑定
for combo, actionType := range keymap.customBinds {
    fmt.Printf("%s -> %s\n", combo, actionType)
}
```

## 最佳实践

### 1. 按键绑定命名规范

使用标准的按键组合格式：
- `C-c` (Ctrl+C)
- `A-x` (Alt+X)
- `S-a` (Shift+A)
- `C-S-a` (Ctrl+Shift+A)

### 2. 上下文管理

成对使用 `PushContext` 和 `PopContext`：
```go
keymap.PushContext("modal")
defer keymap.PopContext()
```

### 3. 鼠标交互优化

根据应用场景调整双击检测参数：
```go
tracker := input.NewMouseTracker()
tracker.doubleClickTimeout = 800 * time.Millisecond  // 更宽松
tracker.doubleClickDistance = 10  // 更宽松的点击位置
```

### 4. 输入读取超时

合理设置读取超时，避免阻塞 UI 渲染循环：
```go
data, err := termInput.ReadEvent(50 * time.Millisecond)
```

## 常见问题

### Q: 如何处理复合按键（如 Ctrl+Shift+A）？

A: 使用自定义绑定：
```go
keymap.Bind("C-S-a", action.ActionSpecial)
```

### Q: 如何禁用某些按键？

A: 绑定到一个空 Action 或不处理该 Action：
```go
keymap.Bind("C-c", action.ActionNone)
```

### Q: 鼠标拖拽如何区分点击和拖拽？

A: MouseTracker 会根据拖拽距离判断，如果拖拽距离小于 `doubleClickDistance`，则不视为拖拽。

### Q: 如何实现 Vim 模式？

A: 使用上下文栈管理不同模式：
```go
// 普通模式
keymap.BindContext("normal", normalModeBinds)

// 插入模式
keymap.BindContext("insert", insertModeBinds)

// 切换模式
keymap.PushContext("normal")
// 按 i 键进入插入模式
keymap.PushContext("insert")
```

### Q: 输入读取为什么会卡住？

A: 使用带超时的读取或可取消的读取：
```go
data, err := termInput.ReadEvent(100 * time.Millisecond)
if err == input.ErrTimeout {
    // 超时，继续其他逻辑
}
```
