# AI Integration

AI 功能集成层。提供 AI Agent 与 Mint TUI 交互的语义化接口。

## 职责

- **Controller 接口**: 定义 AI 与 TUI 交互的标准接口
- **Runtime 实现**: Runtime 内置的 Controller 实现
- **Operations**: 可组合的操作序列，用于构建复杂的 AI 行为
- **感知能力**: Inspector、Query、Wait 等功能让 AI"看到"UI 状态
- **操作能力**: Click、Input、Navigate 等功能让 AI"操作"UI

## 核心概念

### 1. 设计理念：AI 与人类平级

AI Agent 通过与人类相同的语义化接口与 TUI 交互：

```
人类用户:       AI Agent:
   ↓              ↓
   键盘        →  Action
   鼠标        →  Dispatch()
   输入        →  Input()
   点击        →  Click()
   导航        →  Navigate()
```

**关键原则**:
- AI 和人类使用相同的 Action 类型
- AI 通过语义化接口操作，而非直接修改内部状态
- AI 只能通过 Controller 提供的能力与 TUI 交互
- 不提供"上帝模式"或"作弊菜单"给 AI

### 2. Controller 接口

Controller 是 AI 与 TUI 交互的核心接口：

**感知能力**:
- `Inspect()`: 获取完整的 UI 状态快照
- `Find(selector)`: 类似 DOM 选择器，查找组件（支持 `#id`, `.Type`, `[key=value]`）
- `Query(query)`: 查询状态
- `WaitUntil(condition, timeout)`: 等待特定状态出现

**操作能力**:
- `Dispatch(action)`: 发送语义化 Action
- `Click(id)`: 点击组件
- `Input(id, text)`: 输入文本
- `Navigate(direction)`: 焦点导航

**高级能力**:
- `Execute(ops...)`: 执行操作序列
- `Watch(callback)`: 监控状态变化

### 3. Operation 操作

Operation 是可组合的操作单元：

**基础操作**:
- `ClickOperation`: 点击组件
- `InputOperation`: 输入文本
- `NavigateOperation`: 导航焦点
- `DispatchOperation`: 分发 Action

**控制流操作**:
- `BatchOperation`: 批量执行（可选原子性）
- `RepeatOperation`: 重复执行
- `RetryOperation`: 重试执行
- `WaitOperation`: 等待条件
- `WaitValueOperation`: 等待特定值

**快捷函数**:
- `Click(id)` → Operation
- `Input(id, text)` → Operation
- `Navigate(dir)` → Operation
- `Batch(ops...)` → Operation
- `AtomicBatch(ops...)` → Operation

### 4. 选择器语法

AI 可以使用类似 DOM 的选择器查找组件：

```go
// ID 选择器: 精确匹配组件 ID
"#login-form"       // 找到 ID 为 login-form 的组件

// 类型选择器: 匹配特定类型的所有组件
".Button"           // 找到所有 Button 类型的组件
".TextInput"        // 找到所有 TextInput 类型的组件

// 属性选择器: 匹配具有特定属性值的组件
"[placeholder='Email']"          // 匹配 placeholder=Email 的组件
"[label='Submit']"               // 匹配 label=Submit 的组件
"[disabled='true']"              // 匹配禁用的组件

// 通配符: 匹配所有组件
"*"                 // 返回所有组件
```

## 使用示例

### 创建 RuntimeController

```go
// 从 Runtime 创建 Controller
ctrl := ai.NewRuntimeController(
    runtime.GetActionDispatcher(),
    runtime.GetStateTracker(),
    runtime.GetFocusManager(),
)
```

### 感知 UI 状态

```go
// 获取完整状态快照
snapshot, err := ctrl.Inspect()
if err != nil {
    log.Fatalf("Failed to inspect: %v", err)
}

// 查看所有组件
for id, comp := range snapshot.Components {
    fmt.Printf("ID: %s, Type: %s, State: %v\n", id, comp.Type, comp.State)
}

// 使用查找器查找组件
// 查找特定的 Button
btns, err := ctrl.Find(".Button")
if err != nil {
    log.Printf("No buttons found: %v", err)
}

// 查找特定的组件 ID
loginForm, err := ctrl.Find("#login-form")
if err != nil {
    log.Printf("Login form not found: %v", err)
}

// 查找特定属性
emailInput, err := ctrl.Find("[placeholder='Email']")
if err != nil {
    log.Printf("Email input not found: %v", err)
}

// 查询状态
query := ai.StateQuery{
    ComponentID: "input-username",
    StateKey:    "value",
}
result, err := ctrl.Query(query)
if err == nil {
    fmt.Printf("Username value: %v\n", result["value"])
}
```

### 操作 UI

```go
// 点击组件
err := ctrl.Click("button-submit")
if err != nil {
    log.Printf("Click failed: %v", err)
}

// 输入文本
err = ctrl.Input("input-username", "john_doe")
if err != nil {
    log.Printf("Input failed: %v", err)
}

// 导航焦点
err = ctrl.Navigate(ai.DirectionNext)  // 焦点到下一个
err = ctrl.Navigate(ai.DirectionPrev)  // 焦点到上一个
err = ctrl.Navigate(ai.DirectionDown)  // 焦点向下

// 直接分发 Action
action := action.NewAction(action.ActionSubmit).
    WithTarget("login-form").
    WithPayload(map[string]interface{}{"username": "john"})
err = ctrl.Dispatch(action)
```

### 使用 Operations 构建复杂行为

```go
// 填写并提交表单
ops := []ai.Operation{
    ai.Input("input-username", "john_doe"),
    ai.Input("input-password", "secret123"),
    ai.Click("button-login"),
    ai.Wait(func(s *state.Snapshot) bool {
        // 等待登录成功
        comp, ok := s.GetComponent("dashboard")
        return ok && comp.Visible
    }, 5*time.Second),
}

err := ctrl.Execute(ops...)
if err != nil {
    log.Printf("Form submission failed: %v", err)
}
```

### 批量操作

```go
// 批量操作（非原子，按顺序执行）
batch := ai.Batch(
    ai.Input("field1", "value1"),
    ai.Input("field2", "value2"),
    ai.Input("field3", "value3"),
)

err := batch.Execute(ctrl)

// 原子批量操作（任一失败则整体失败）
atomicBatch := ai.AtomicBatch(
    ai.Input("field1", "value1"),
    ai.Input("field2", "value2"),
    ai.Input("field3", "value3"),
)

err = atomicBatch.Execute(ctrl)
```

### 重复和重试

```go
// 重复点击 5 次
repeat := ai.NewRepeatOperation(
    ai.Click("button-add-item"),
    5,
    100*time.Millisecond, // 每次点击间隔 100ms
)

err := repeat.Execute(ctrl)

// 重试操作（最多 3 次，每次间隔 1 秒）
retry := ai.NewRetryOperation(
    ai.Click("button-load-more"),
    3,                    // 最多 3 次尝试
    1*time.Second,        // 每次间隔 1 秒
    func(err error) bool { // 根据错误决定是否重试
        return err != nil
    },
)

err = retry.Execute(ctrl)
```

### 等待条件

```go
// 等待特定条件满足
err := ctrl.WaitUntil(func(snapshot *state.Snapshot) bool {
    comp, ok := snapshot.GetComponent("loading-spinner")
    return ok && !comp.Visible // 等待加载完成
}, 10*time.Second)

if err != nil {
    log.Printf("Timeout waiting for loading to complete: %v", err)
}

// 使用 Operation 等待特定值
waitValue := ai.WaitValue(
    "input-result",
    "value",
    "expected_result",
    5*time.Second,
)

err = waitValue.Execute(ctrl)
```

### 监控状态变化

```go
// 监控所有状态变化
unsubscribe := ctrl.Watch(func(snapshot *state.Snapshot) {
    // 检查特定的组件状态
    if comp, ok := snapshot.GetComponent("user-status"); ok {
        fmt.Printf("New status: %v\n", comp.State["status"])
    }
})

// 取消监控
// unsubscribe()
```

### 快捷方法

```go
// 获取组件状态
value, err := ctrl.GetState("input-username", "value")

// 设置组件状态
err = ctrl.SetValue("input-username", "value", "new_value")

// 检查可见性
visible, err := ctrl.IsVisible("modal-dialog")

// 检查是否禁用
disabled, err := ctrl.IsDisabled("button-submit")

// 获取当前焦点
focusID, err := ctrl.GetFocused()

// 等待组件可见
err = ctrl.WaitForVisible("modal-dialog", 3*time.Second)

// 等待状态值
err = ctrl.WaitForValue(
    "loading-spinner",
    "visible",
    false,
    5*time.Second,
)
```

## 核心类型

### Controller

```go
type Controller interface {
    // 感知能力
    Inspect() (*state.Snapshot, error)
    Find(selector string) ([]ComponentInfo, error)
    Query(query StateQuery) (map[string]interface{}, error)
    WaitUntil(condition func(*state.Snapshot) bool, timeout time.Duration) error

    // 操作能力
    Dispatch(a *action.Action) error
    Click(componentID string) error
    Input(componentID, text string) error
    Navigate(direction Direction) error

    // 高级能力
    Execute(ops ...Operation) error
    Watch(callback func(*state.Snapshot)) func()
}
```

### RuntimeController

```go
type RuntimeController struct {
    dispatcher *action.Dispatcher
    tracker    *state.Tracker
    focusMgr   *focus.Manager
}

func NewRuntimeController(dispatcher *action.Dispatcher, tracker *state.Tracker, focusMgr *focus.Manager) *RuntimeController

// 快捷方法
func (c *RuntimeController) GetState(componentID, stateKey string) (interface{}, error)
func (c *RuntimeController) SetValue(componentID string, stateKey string, value interface{}) error
func (c *RuntimeController) IsVisible(componentID string) (bool, error)
func (c *RuntimeController) IsDisabled(componentID string) (bool, error)
func (c *RuntimeController) GetFocused() (string, error)
func (c *RuntimeController) WaitForVisible(componentID string, timeout time.Duration) error
func (c *RuntimeController) WaitForValue(componentID, stateKey string, expected interface{}, timeout time.Duration) error
```

### Operation

```go
type Operation interface {
    Execute(ctrl Controller) error
}

// 基础操作
func Click(id string) Operation
func Input(id, text string) Operation
func Navigate(dir Direction) Operation

// 等待操作
func Wait(condition func(*state.Snapshot) bool, timeout time.Duration) Operation
func WaitValue(id, key string, expected interface{}, timeout time.Duration) Operation

// 批量操作
func Batch(ops ...Operation) Operation
func AtomicBatch(ops ...Operation) Operation
```

### ComponentInfo

```go
type ComponentInfo struct {
    ID       string
    Type     string
    Props    map[string]interface{}
    State    map[string]interface{}
    Rect     state.Rect
    Visible  bool
    Disabled bool
}
```

## 文件结构

- `controller.go` - Controller 接口定义和相关类型
- `runtime_controller.go` - RuntimeController 实现
- `operations.go` - 各种 Operation 类型和构建器

## 依赖

**可以依赖**:
- `runtime/action` - Action 类型定义
- `runtime/state` - State 类型和 Snapshot
- `runtime/focus` - 焦点管理
- 标准库: `time`, `fmt`, `strings`

**不能依赖**:
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 与其他模块集成

### 与 Runtime 集成

```go
// Runtime 暴露 Controller
runtime := core.NewRuntime(platform)
ctrl := ai.NewRuntimeController(
    runtime.GetActionDispatcher(),
    runtime.GetStateTracker(),
    runtime.GetFocusManager(),
)

// AI 使用 Controller 进行交互
ctrl.Click("button-start")
```

### 与 DevTools 集成

AI Controller 可用于：
- 自动化测试
- 回放用户操作
- 生成测试脚本
- UI 可访问性检查

## 最佳实践

### 1. 总是使用语义化操作

```go
// 推荐：使用语义化操作
ctrl.Click("button-submit")

// 不推荐：直接发送原始 Action（AI 不应该这样做）
ctrl.Dispatch(action.NewAction(0x1234))
```

### 2. 使用 Wait 确保操作成功

```go
// 填写表单并等待提交结果
ops := []ai.Operation{
    ai.Input("email", "test@example.com"),
    ai.Click("button-submit"),
    ai.Wait(func(s *state.Snapshot) bool {
        comp, ok := s.GetComponent("result-message")
        return ok && comp.Visible
    }, 5*time.Second),
}
```

### 3. 使用 Batch 进行批量操作

```go
// 一次性填充多个字段
batch := ai.Batch(
    ai.Input("first-name", "John"),
    ai.Input("last-name", "Doe"),
    ai.Input("email", "john.doe@example.com"),
)

err := batch.Execute(ctrl)
```

### 4. 处理错误

```go
err := ctrl.Click("button-submit")
if err != nil {
    if _, ok := err.(*ai.ComponentNotFoundError); ok {
        log.Printf("Button not found")
    } else if _, ok := err.(*ai.ComponentDisabledError); ok {
        log.Printf("Button is disabled")
    } else {
        log.Printf("Click failed: %v", err)
    }
}
```

### 5. 使用 Watch 监控状态变化

```go
unsubscribe := ctrl.Watch(func(snapshot *state.Snapshot) {
    // 检查状态是否满足预期
    // 如果不满足，AI 可以采取行动
})

// 不要忘记取消订阅
defer unsubscribe()
```

## 常见问题

### Q: AI 能否绕过用户的权限限制？

A: 不能。AI 通过相同的 Controller 接口操作，所有操作都会经过相同的验证和权限检查。AI 不能访问"后门"或"上帝模式"。

### Q: AI 如何学习复杂的操作流程？

A: AI 可以录制人类操作序列，然后重放。Operations 可以组合成复杂的行为，AI 可以通过自然语言描述生成这些 Operations。

### Q: 如何确保 AI 不会破坏应用状态？

A:
1. 使用 `AtomicBatch` 确保操作要么全部成功，要么全部失败
2. 使用 `Wait` 确保操作完成后再进行下一步
3. 使用 `Snapshot` 记录操作前后的状态，支持回滚

### Q: AI 可以修改源代码吗？

A: 不能。AI 只能通过 Controller 提供的接口与运行中的应用交互，不能修改源代码。

### Q: 选择器支持组合查询吗？

A: 当前支持简单的选择器。复杂的组合查询需要在应用层实现。例如：
```go
// 查找 Button 且 id 包含 "submit"
buttons, _ := ctrl.Find(".Button")
for _, btn := range buttons {
    if strings.Contains(btn.ID, "submit") {
        // 找到了
    }
}
```
