# 迁移指南

本指南帮助你将现有代码迁移到新的组件状态持久化架构。

## 概述

新的架构引入了：

1. **组件实例 (ComponentInstance)**: 持久化的组件实例，状态在重新渲染时保持不变
2. **实例管理器 (InstanceManager)**: 管理组件实例的生命周期和匹配
3. **useHoverState Hook**: 专门用于管理悬停状态的 hook

## 迁移步骤

### 步骤 1: 识别状态丢失问题

如果你的组件在重新渲染时丢失状态，需要检查：

- 状态是否存储在 VNode 的字段中？
- 组件是否使用 key 进行标识？
- 是否有悬停状态在交互后丢失？

**问题示例**:

```go
// 旧代码：悬停状态在按钮 VNode 中
btn := ui.Button("Click me")
// 每次渲染创建新的 ButtonVNode，isHovered 丢失
```

### 步骤 2: 使用 Hooks 替代直接状态存储

**旧方式** (状态存储在 VNode 中):

```go
// 按钮内部实现（框架代码）
type ButtonVNode struct {
    *ElementVNode
    isHovered bool // 状态在 VNode 中
}
```

**新方式** (使用 Hooks):

```go
func MyButton() ui.VNode {
    isHovered, setHovered := ui.UseHoverState()

    var style ui.StyleBuilder
    if isHovered() {
        style = ui.NewStyle().Background("cyan")
    } else {
        style = ui.NewStyle().Background("blue")
    }

    return ui.Button("Click me").
        Style(style.Build()).
        OnMouseEnter(func() { setHovered(true) }).
        OnMouseLeave(func() { setHovered(false) })
}
```

### 步骤 3: 为动态组件添加 Key

**旧代码** (无 key):

```go
func UserList(users []User) ui.VNode {
    children := make([]ui.VNode, len(users))
    for i, user := range users {
        children[i] = ui.ComponentWithProps("UserItem", UserItem).
            Prop("name", user.Name).
            Build() // 没有 key！
    }
    return ui.Column(children...)
}
```

**新代码** (添加稳定的 key):

```go
func UserList(users []User) ui.VNode {
    children := make([]ui.VNode, len(users))
    for i, user := range users {
        children[i] = ui.ComponentWithProps("UserItem", UserItem).
            Prop("name", user.Name).
            Key(fmt.Sprintf("user-%d", user.ID)). // 添加稳定的 key
            Build()
    }
    return ui.Column(children...)
}
```

### 步骤 4: 使用 useRef 保持可变引用

如果你需要在渲染之间保持可变值而不触发重新渲染：

```go
func TimerComponent() ui.VNode {
    // 使用 ref 存储定时器引用
    timerRef := ui.UseRef(nil)

    ui.UseEffect(func() func() {
        ticker := time.NewTicker(time.Second)
        timerRef.Value = ticker

        go func() {
            for range ticker.C {
                // 更新逻辑
            }
        }()

        return func() {
            // 清理时使用保存的引用
            if ticker, ok := timerRef.Value.(*time.Ticker); ok {
                ticker.Stop()
            }
        }
    }, nil)

    return ui.Text("Timer running...")
}
```

## 常见迁移场景

### 场景 1: 简单计数器

**旧代码** (状态丢失):

```go
func App() ui.VNode {
    // 每次渲染 count 都重置为 0
    count := 0

    return ui.Row(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("[+]").OnClick(func() {
            count++ // 修改不会反映到 UI
        }),
    )
}
```

**新代码** (使用 useState):

```go
func App() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.Row(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("[+]").OnClick(func() {
            setCount(count + 1)
        }),
    )
}
```

### 场景 2: 列表项状态

**旧代码** (列表项状态丢失):

```go
func TodoList(todos []string) ui.VNode {
    items := make([]ui.VNode, len(todos))
    for i, todo := range todos {
        // 每次渲染创建新的 Done 状态
        done := false

        items[i] = ui.Row(
            ui.Checkbox("", done),
            ui.Text(todo),
        )
    }
    return ui.Column(items...)
}
```

**新代码** (使用组件实例):

```go
// 创建独立的 TodoItem 组件
func TodoItem(props ui.Props) ui.VNode {
    todo := props.GetString("todo")

    // 每个 TodoItem 有自己的 done 状态
    done, setDone := ui.UseStateBool(false)

    return ui.Row(
        ui.CheckBox("", done).OnChange(setDone),
        ui.Text(todo),
    )
}

func TodoList(todos []string) ui.VNode {
    items := make([]ui.VNode, len(todos))
    for i, todo := range todos {
        items[i] = ui.ComponentWithProps("TodoItem", TodoItem).
            Prop("todo", todo).
            Key(fmt.Sprintf("todo-%d", i)). // 重要：添加 key
            Build()
    }
    return ui.Column(items...)
}
```

### 场景 3: 表单输入

**旧代码** (输入值丢失):

```go
func Form() ui.VNode {
    value := "" // 每次渲染重置

    return ui.Column(
        ui.Input("Name", value, 20).
            OnChange(func(newValue string) {
                value = newValue // 不会反映到 UI
            }),
        ui.Button("Submit"),
    )
}
```

**新代码** (使用 useState):

```go
func Form() ui.VNode {
    value, setValue := ui.UseStateString("")

    return ui.Column(
        ui.Input("Name", value, 20).
            OnChange(func(newValue string) {
                setValue(newValue)
            }),
        ui.Button("Submit").OnClick(func() {
            fmt.Printf("Submitted: %s\n", value)
        }),
    )
}
```

### 场景 4: 带悬停效果的交互元素

**旧代码** (依赖框架的临时解决方案):

```go
// 框架内部使用 bounds 匹配恢复悬停状态
// 问题：元素位置改变时会失效
```

**新代码** (使用 useHoverState):

```go
func InteractiveButton() ui.VNode {
    isHovered, setHovered := ui.UseHoverState()

    var style ui.StyleBuilder
    if isHovered() {
        style = ui.NewStyle().
            Background("cyan").
            Foreground("black").
            Bold(true)
    } else {
        style = ui.NewStyle().
            Background("blue").
            Foreground("white")
    }

    return ui.Button("Hover me").
        Style(style.Build()).
        OnMouseEnter(func() { setHovered(true) }).
        OnMouseLeave(func() { setHovered(false) })
}
```

## API 变更

### 新增 API

| API | 用途 |
|-----|------|
| `UseStateInt(initial int)` | 整数状态 |
| `UseStateString(initial string)` | 字符串状态 |
| `UseStateBool(initial bool)` | 布尔状态 |
| `UseRef(initial interface{}) *Ref` | 持久化引用 |
| `UseHoverState() (func() bool, func(bool))` | 悬停状态 |
| `UseEffect(callback, deps)` | 副作用 |
| `UseMemo(func(), deps)` | 缓存值 |
| `UseCallback(func(), deps)` | 缓存函数 |

### 组件 Builder 方法

| 方法 | 说明 |
|------|------|
| `.Key(key string)` | 设置组件 key（重要！） |
| `.Prop(key, value)` | 设置单个 prop |
| `.Props(Props)` | 批量设置 props |

## 兼容性

### 向后兼容

现有代码继续工作：

- 直接使用 `ui.Button()`, `ui.Input()` 等内置组件无需修改
- 内置组件的悬停状态仍然通过 bounds 匹配恢复
- 框架的临时解决方案仍然有效

### 推荐迁移

对于新代码或需要可靠状态管理的组件：

- 使用函数组件 + Hooks
- 为动态组件设置 key
- 使用 `useHoverState` 管理悬停状态

## 检查清单

迁移时确保：

- [ ] 所有使用状态的组件都使用 Hooks（useState, useRef 等）
- [ ] 动态列表中的每个组件都有唯一的 key
- [ ] 悬停状态使用 useHoverState 管理
- [ ] useEffect 的清理函数正确设置
- [ ] 复杂状态考虑使用自定义 Hook 或状态管理库
- [ ] 测试重新渲染后状态是否保持

## 故障排除

### 问题：状态在重新渲染时丢失

**原因**：状态存储在 VNode 字段中而非使用 Hooks

**解决**：改用 `useState` 或其他状态 Hook

### 问题：列表项顺序错乱

**原因**：缺少稳定的 key 或使用索引作为 key

**解决**：为每个项设置唯一的、稳定的 key

```go
// 不好
Key(fmt.Sprintf("item-%d", i))

// 好
Key(fmt.Sprintf("item-%d", item.ID))
```

### 问题：useEffect 清理函数未执行

**原因**：组件实例未正确匹配

**解决**：确保组件 key 在渲染间保持一致

### 问题：悬停状态仍然丢失

**原因**：可能未使用 useHoverState 或组件没有 key

**解决**：
1. 确保使用 `UseHoverState()` hook
2. 为组件设置唯一的 key
3. 确保悬停状态更新函数正确调用

## 示例项目

完整示例请参考：

- `examples/counter/` - 基础计数器
- `examples/todolist/` - 带状态的列表
- `examples/hover/` - 悬停状态演示

## 获取帮助

如有问题或需要帮助：

- 查阅 [组件开发指南](./component-development-guide.md)
- 参考 [Hooks API 文档](../api/hooks.md)
- 提交 Issue 到 GitHub
