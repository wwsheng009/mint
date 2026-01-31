# State Management (V3)

状态管理系统核心实现。

## 职责

- **状态快照（Snapshot）**：记录特定时间点的完整 UI 状态
- **状态追踪（Tracker）**：跟踪状态变化历史，支持 Undo/Redo
- **状态监听**：在状态变化时通知订阅者
- **时间旅行**：为 DevTools 提供状态回溯能力

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 核心概念

### 状态设计原则

所有状态必须是：

1. **可枚举**：可以完整列出所有状态（无隐藏内部状态）
2. **可快照**：可以保存和恢复状态
3. **可追溯**：每个状态变化都能追溯到 Action

### Snapshot（状态快照）

`Snapshot` 是在某个时间点的完整 UI 状态：

```go
type Snapshot struct {
    Timestamp  time.Time                   // 时间戳
    FocusPath  FocusPath                   // 焦点路径
    Components map[string]ComponentState  // 组件状态
    Modals     []ModalState                // Modal 栈
    Dirty      DirtyRegion                 // 脏区域
    Metadata   map[string]interface{}      // 元数据
}
```

### 组件状态

每个组件的状态包含静态属性和动态状态：

```go
type ComponentState struct {
    ID       string              // 组件标识
    Type     string              // 组件类型
    Props    map[string]interface{} // 静态属性（配置）
    State    map[string]interface{} // 动态状态（运行时）
    Rect     Rect                // 布局位置
    Visible  bool               // 可见性
    Disabled bool               // 可交互性
}
```

**Props vs State**：
- `Props`：配置属性，由父组件传递，组件内部不应修改
- `State`：动态状态，由组件内部维护，通过 Action 修改

### Tracker（状态追踪器）

`Tracker` 管理状态历史和变化监听：

```go
type Tracker struct {
    current    *Snapshot               // 当前状态
    past       []*Snapshot            // 历史状态（Undo 用）
    future     []*Snapshot            // 未来状态（Redo 用）
    maxHistory int                    // 最大历史记录数
    listeners  []ChangeListener       // 变化监听器
}
```

**工作流程**：

1. `BeforeAction()`：在执行 Action 前记录状态
2. 执行 Action（修改 UI）
3. `AfterAction(before)`：比较前后差异，更新历史

### 焦点路径（FocusPath）

使用路径而非索引或布尔值表示焦点，支持深层嵌套和 Modal：

```go
type FocusPath []string

// 例如：["app", "mainPanel", "form", "nameInput"]
```

**优势**：
- 支持深层嵌套
- 不受组件树顺序变化影响
- 便于时间旅行回溯

### 状态差异和脏区域

`DirtyRegion` 记录需要重新渲染的区域，用于优化渲染性能：

```go
type DirtyRegion struct {
    Cells []CellRef  // 脏单元格列表
    Rects []Rect     // 脏矩形列表（优化渲染）
}
```

## 使用示例

### 创建 Tracker

```go
import "github.com/wwsheng009/mint/runtime/state"

// 创建状态追踪器
tracker := state.NewTracker()

// 设置最大历史记录数
tracker.SetMaxHistory(100)
```

### Action 前后记录状态

```go
// 执行 Action 前记录
before := tracker.BeforeAction()

// 执行 Action（修改组件状态）
myComponent.ExecuteAction(action)

// 执行 Action 后记录（自动比较并更新历史）
after := tracker.AfterAction(before)
```

### 直接修改组件状态

```go
// 设置组件状态
tracker.SetComponentState("button1", map[string]interface{}{
    "pressed": true,
    "text":    "正在提交...",
})

// 获取组件状态
if state, ok := tracker.GetComponentState("button1"); ok {
    fmt.Println("按钮状态:", state)
}
```

### 焦点路径操作

```go
// 设置焦点
path := state.FocusPath{"app", "mainPanel", "form", "nameInput"}
tracker.SetFocusPath(path)

// 获取焦点
focusPath := tracker.GetFocusPath()
fmt.Println("当前焦点:", focusPath.String()) // app.mainPanel.form.nameInput

// 聚焦路径操作
parentPath := focusPath.Parent()        // ["app", "mainPanel", "form"]
currentFocus := focusPath.Current()     // "nameInput"
extendedPath := focusPath.Append("otherField") // ["app", "mainPanel", "form", "nameInput", "otherField"]
```

### Undo / Redo

```go
// 撤销
if tracker.CanUndo() {
    tracker.Undo()
}

// 重做
if tracker.CanRedo() {
    tracker.Redo()
}

// 检查可用性
fmt.Println("可撤销:", tracker.CanUndo())
fmt.Println("可重做:", tracker.CanRedo())
fmt.Println("历史记录数:", tracker.GetHistorySize())
```

### 状态变化监听

```go
// 订阅状态变化
unsubscribe := tracker.Subscribe(func(old, new *state.Snapshot) {
    fmt.Println("状态变化:")
    fmt.Println("  旧焦点:", old.FocusPath.String())
    fmt.Println("  新焦点:", new.FocusPath.String())
})

// 取消订阅
unsubscribe()
```

### 获取当前状态和历史

```go
// 获取当前快照
current := tracker.Current()
fmt.Println("包含组件数:", len(current.Components))

// 获取组件状态
if comp, ok := current.GetComponent("button1"); ok {
    fmt.Printf("组件 %s: %+v\n", comp.ID, comp.State)
}

// 获取历史记录
history := tracker.GetHistory()
for i, snap := range history {
    fmt.Printf("[%d] %s\n", i, snap.Timestamp)
}
```

### 与 DevTools 集成

```go
// 时间旅行调试：跳转到指定的历史状态
func TimeTravel(tracker *state.Tracker, index int) {
    history := tracker.GetHistory()
    if index < 0 || index >= len(history) {
        return
    }
    
    // 恢复到指定状态
    target := history[index]
    tracker.Update(target)
    
    // 重新渲染 UI
    RenderUIFromSnapshot(target)
}
```

## 核心类型

| 类型 | 说明 |
|------|------|
| `Tracker` | 状态追踪器，管理历史和监听器 |
| `Snapshot` | 状态快照，记录某个时间点的完整状态 |
| `ComponentState` | 组件状态（Props + State） |
| `FocusPath` | 焦点路径（字符串切片） |
| `Rect` | 矩形区域（X, Y, Width, Height） |
| `DirtyRegion` | 脏区域，用于渲染优化 |
| `ModalState` | Modal 状态 |
| `CellRef` | 单元格引用 |

## 文件结构

| 文件 | 说明 |
|------|------|
| `tracker.go` | Tracker 实现，状态历史和 Undo/Redo |
| `snapshot.go` | Snapshot 和相关类型定义 |
| `diff.go` | 状态差异计算（如果有） |
| `serialize.go` | 状态序列化/反序列化（如果有） |

## 最佳实践

### 1. Props vs State 的区分

```go
// Props：配置，由父组件设置
component := &Button{
    Props: map[string]interface{}{
        "label":  "提交",
        "primary": true,
    },
}

// State：动态，由组件内部更新
component.State = map[string]interface{}{
    "loading": false,
    "pressed": false,
}
```

### 2. 使用 BeforeAction/AfterAction 模式

```go
func ExecuteActionWithTracking(tracker *state.Tracker, action *action.Action) {
    before := tracker.BeforeAction()
    
    // 执行 Action
    dispatch(action)
    
    // 记录变化
    tracker.AfterAction(before)
}
```

### 3. 清理历史记录

```go
// 在特定事件后清理历史（如提交表单）
tracker.ClearHistory()

// 或限制历史大小
tracker.SetMaxHistory(50)
```

### 4. 时间旅行和 DevTools

```go
// 记录每个快照的时间戳
snap := tracker.Current()
fmt.Println("状态时间:", snap.Timestamp)

// 使用元数据记录额外信息
snap.Metadata["action"] = action.Type
snap.Metadata["source"] = action.Source
```

### 5. 组件状态更新

```go
// 推荐：完整替换状态（便于比较）
tracker.SetComponentState("id", map[string]interface{}{
    "field1": value1,
    "field2": value2,
})

// 不推荐：部分更新（会影响差分比较）
```
