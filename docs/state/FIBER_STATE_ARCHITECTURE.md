# Fiber-first 状态管理架构

本文档描述 Mint Fiber-first 模式下的状态管理机制。

## 目录

- [概览](#概览)
- [核心架构](#核心架构)
- [状态流转](#状态流转)
- [存储层级](#存储层级)
- [Intent 系统](#intent-系统)
- [InstanceManager](#instancemanager)

---

## 概览

### 设计原则

1. **声明式分离**：VNode（描述）与 Fiber（运行时状态）分离
2. **双重状态管理**：组件级状态（Hooks）与全局状态（Intent）并存
3. **单向数据流**：Intent → Handler → State Update → Re-render

### 核心流程

```
VNode (声明式描述) → Fiber (运行时实例) → LayoutBox (布局) → PaintableBox (绘制)
```

---

## 核心架构

### 数据流图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          用户操作层                                      │
│  Button Click / Keyboard Input / Form Submit                             │
└──────────────────────────────────────┬──────────────────────────────────┘
                                       │ EmitIntent()
                                       ↓
┌─────────────────────────────────────────────────────────────────────────┐
│                          Intent 分发层                                    │
│  Runtime.Emit() → Dispatcher.Dispatch()                                  │
│  ┌─────────────────────────────────────────────────────────────────────┤
│  │  Dispatcher.StateSetter = Root ComponentContext (Global State Store) │
│  └─────────────────────────────────────────────────────────────────────┘
└──────────────────────────────────────┬──────────────────────────────────┘
                                       │ Handler.Handle()
                                       ↓
┌─────────────────────────────────────────────────────────────────────────┐
│                          Intent Handler 层                                │
│  registeredHandler.Execute() {                                          │
│      ctx.SetState("step", 2)  // 更新 RootContext.State map             │
│      ctx.ScheduleUpdate()    // 触发重新渲染                             │
│  }                                                                       │
└──────────────────────────────────────┬──────────────────────────────────┘
                                       │ fwApp.MarkDirty()
                                       ↓
┌─────────────────────────────────────────────────────────────────────────┐
│                          重新渲染层                                       │
│  DeclarativeNode.Paint() → Reconciler.Render() → BeginWork()             │
│  ┌─────────────────────────────────────────────────────────────────────┤
│  │  beginWorkComponent() {                                             │
│  │      if isRootComponent {                                          │
│  │          ctx = rootContext  // 使用共享的全局 State                │
│  │      } else {                                                       │
│  │          ctx = instance.GetContext()  // 使用组件独立的 Hooks State │
│  │      }                                                               │
│  │      SetCurrentContext(ctx)                                         │
│  │      renderFunc()  // App() 函数执行                                │
│  │  }                                                                   │
│  └─────────────────────────────────────────────────────────────────────┘
└──────────────────────────────────────┬──────────────────────────────────┘
                                       │ App() 函数调用
                                       ↓
┌─────────────────────────────────────────────────────────────────────────┐
│                          状态读取层                                       │
│  App() {                                                                 │
│      ctx := rtui.GetCurrentContext()  // → Root ComponentContext         │
│      step := ctx.GetIntState("step", 1)  // → State["step"] = 2 ✓        │
│      // ... 根据 step 渲染 UI                                           │
│  }                                                                       │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 状态流转

### Fiber 树构建流程

```go
// 1. 用户组件定义 (VNode 层)
func App() ui.VNode {
    return ui.VStack(
        StepIndicator(),  // 读取全局状态
        ActionButtons(),  // 发射 Intent
    )
}

// 2. VNode → Fiber 转换 (fiber_util.go)
func CreateFiber(vnode VNode) *Fiber {
    return &Fiber{
        NodeID: generateNodeID(),
        Key:    vnode.Key(),
        Type:   vnode.Type(),
        // Instance 在 beginWorkComponent 中创建
        Instance: nil,  // 稍后填充
        Props:    vnode.Props(),
    }
}

// 3. BeginWork - 创建/恢复 Instance (begin_work.go)
func beginWorkComponent(current, workInProgress *Fiber) *Fiber {
    componentKey := workInProgress.Key  // "root"

    if componentKey == "root" {
        // 根组件：使用共享的 rootContext
        ctx = currentReconciler.ctx  // DeclarativeNode.instance
    } else {
        // 子组件：使用独立的 ComponentInstance
        instance = instanceMgr.GetOrCreate(componentKey, creator)
        ctx = instance.GetContext()
    }

    workInProgress.Instance = instance
    workInProgress.Context = ctx

    // 渲染组件
    SetCurrentContext(ctx)
    children := workInProgress.ComponentFunc()  // 调用 App()
    // ...
}
```

### 布局绘制流程

```go
// 4. CommitWork - 提交阶段 reconcile/commit_work.go
func CommitWork(current, workInProgress *Fiber) {
    // 1. 应用 focus/hover 等交互状态
    if workInProgress.Type == rtui.VNodeElement {
        applyInteractionStates(workInProgress)
    }

    // 2. 计算布局 (LayoutEngine)
    layoutBox := layoutEngine.Layout(workInProgress)
    workInProgress.ComputedBox = layoutBox

    // 3. 转换为 PaintableBox
    paintableBox := paintEngine.ToPaintable(layoutBox, workInProgress.Style)

    // 4. 绘制到 buffer
    paintBox := PaintableBoxToPaintable(paintableBox)
    renderCallback(workInProgress, paintBox.X, paintBox.Y, buffer)
}
```

---

## 存储层级

### 状态存储的五个层级

| 层级 | 类型 | 位置 | 用途 | 生命周期 |
|-----|------|------|------|---------|
| **1. Hooks 状态** | `[]Hook` | `Fiber.Instance.Context.Hooks` | useState/useEffect 的局部值 | 组件实例生命周期 |
| **2. 全局状态** | `map[string]interface{}` | `DeclarativeNode.instance.State` | Intent Handler 更新的全局数据 | 应用生命周期 |
| **3. Props** | `Props` | `Fiber.Props` | 父组件传递的数据 | Fiber 更新周期 |
| **4. MemoizedProps** | `Props` | `Fiber.MemoizedProps` | Props 的副本用于 diff | Fiber 更新周期 |
| **5. Render 缓存** | `interface{}` | `Fiber.MemoizedState` | Text 内容、Layout 结果等 | Fiber 更新周期 |

### 数据结构定义

```go
// ComponentContext - 组件上下文（包含 Hooks 和全局状态）
type ComponentContext struct {
    ComponentID string
    Hooks       []Hook           // useState/useEffect 值
    HookIndex   int              // 当前 Hook 索引
    RenderCount int              // 渲染次数

    // GlobalState - 全局状态存储（Intent Handler 更新）
    // 在根组件中，这是 DeclarativeNode.instance 的 State
    GlobalState    map[string]interface{}
    GlobalStateMu  sync.RWMutex

    scheduleUpdate func()         // 触发重新渲染的回调
}

// Fiber - 工作单元
type Fiber struct {
    NodeID         uint64           // 唯一 ID
    Key            string           // diff key
    Type           VNodeType        // VNode type
    Tag            string           // element tag / component name

    // Instance - 组件运行时实例
    Instance       ComponentInstance
    Context        *ComponentContext  // 快速访问上下文（可能等于 Instance.Context）

    // Props - 属性
    Props          Props
    MemoizedProps  Props

    // Render cache
    MemoizedState  interface{}      // Text content or custom cache

    // Layout & Paint
    ComputedBox    *BoxConstraints  // 布局结果
    Child          *Fiber
    Sibling        *Fiber
}
```

### 状态访问模式

```go
// ✅ 组件内部状态 - 使用 Hooks
func Counter() ui.VNode {
    count, setCount := rtui.UseState(0)

    return ui.Button(fmt.Sprintf("Count: %d", count)).
        OnClick(func() {
            setCount(count + 1)  // 更新 Hooks 状态
        }).
        Build()
}

// ✅ 全局状态 - 使用 Intent
func App() ui.VNode {
    ctx := rtui.GetCurrentContext()
    step := ctx.GetIntState("step", 1)  // 从 GlobalState 读取

    return ui.Button("Next").
        OnPress(UpdateStepIntent{Step: step + 1}).  // 发射 Intent
        Build()
}

// ✅ Props - 父组件传递
func Parent() ui.VNode {
    title := "Hello"

    return ui.Child().
        Prop("title", title).  // Props 传递
        Build()
}

func Child(props rtui.Props) ui.VNode {
    title := props["title"].(string)
    return ui.Text(title)
}
```

---

## Intent 系统

### Intent 流程图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Intent 定义 (main.go)                             │
│  type UpdateStepIntent struct { Step int }                                │
│  func (UpdateStepIntent) IntentType() string { return "UpdateStep" }     │
└──────────────────────────────────────┬──────────────────────────────────┘
                                       │
                                       ↓
┌─────────────────────────────────────────────────────────────────────────┐
│                         Intent 注册 (ui.WithInit)                         │
│  ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateStepIntent)) │
│      intent.IntentResult {                                               │
│          ctx.SetState("step", i.Step)  // 更新 GlobalState               │
│          return intent.HandledResult()                                   │
│      })                                                                  │
└──────────────────────────────────────┬──────────────────────────────────┘
                                       │
                                       ↓
┌─────────────────────────────────────────────────────────────────────────┐
│                         Intent 发射 (组件中)                              │
│  app.ButtonBuilder().                                                   │
│      OnPress(UpdateStepIntent{Step: 2}).  // 构造 Intent                 │
│      Build()                                                             │
│      │                                                                   │
│      └─> Button.EmitIntent(intent)  // 触发发射                          │
└──────────────────────────────────────┬──────────────────────────────────┘
                                       │
                                       ↓
┌─────────────────────────────────────────────────────────────────────────┐
│                         Intent 运行时                                     │
│  Runtime.Emit(intent)                                                   │
│      │                                                                   │
│      ├─> Dispatcher.Dispatch(intent)                                    │
│      │       ├─> NewActionContext(stateSetter=rootContext)              │
│      │       └─> handler.Handle(ctx, intent)                           │
│      │                                                               ┌───┴───┐
│      │                                                               │       │
│      └─> fwApp.MarkDirty() <─────────────────────────────────────────┘       │
│                                                                   │       │
│                                                                   ↓       ↓
│                                                          SetState()    ScheduleUpdate()
└──────────────────────────────────────────────────────────────┬───────────┘
                                                                       │
                                                                       ↓
┌─────────────────────────────────────────────────────────────────────────┐
│                         重新渲染                                         │
│  App() 函数重新执行 → 读取最新 State → 生成新 Fiber 树 → 提交更新           │
└─────────────────────────────────────────────────────────────────────────┘
```

### Intent Handler 上下文

```go
// ActionContext - Intent Handler 的执行上下文
type ActionContext struct {
    context.Context
    Source   string
    Mutable bool  // 是否允许修改状态
    State    *ComponentContext  // 指向 Root ComponentContext
}

// SetState - 更新全局状态
func (c *ActionContext) SetState(key string, value interface{}) {
    if c.State != nil {
        c.State.SetGlobal(key, value)  // 更新 Root Context.GlobalState

        // 触发重新渲染
        if c.State.scheduleUpdate != nil {
            c.State.scheduleUpdate()
        }
    }
}

// GetState - 读取状态
func (c *ActionContext) GetState(key string) (interface{}, bool) {
    if c.State != nil {
        return c.State.GetGlobal(key)
    }
    return nil, false
}
```

---

## InstanceManager

### InstanceManager 架构

```go
type InstanceManager struct {
    mu            sync.RWMutex

    // 两个索引：字符串 key 和 NodeID
    instances     map[string]ComponentInstance      // old: key -> instance
    instancesByID map[uint64]ComponentInstance     // new: NodeID -> instance

    // LRU 管理和限制
    instanceOrder []string
    lastAccess    map[string]time.Time
    maxInstances  int  // 防止内存泄漏
}
```

### 实例复用逻辑

```go
func (m *InstanceManager) GetOrCreate(key string, creator func() ComponentInstance) ComponentInstance {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 尝试复用现有实例
    if inst, exists := m.instances[key]; exists {
        m.lastAccess[key] = time.Now()  // 更新访问时间
        m.moveToEnd(key)                // LRU: 移到最末尾
        return inst
    }

    // 创建新实例
    inst := creator()
    inst.OnMount()  // 生命周期：OnMount

    // 存储在两个索引中
    m.instances[key] = inst
    instanceOrder = append(instanceOrder, key)
    lastAccess[key] = time.Now()

    // 应用 LRU 限制
    m.cleanupLRU()

    return inst
}
```

### 实例生命周期

```go
// OnMount - 实例创建时调用（新实例）
func (b *BaseComponentInstance) OnMount() {
    // 初始化副作用
    // 注册事件监听器
    // 获取初始数据
}

// OnUnmount - 实例销毁时调用
func (b *BaseComponentInstance) OnUnmount() {
    // 清理副作用
    // 取消事件监听器
    // 释放资源
}

// Update Props - 更新属性时
func (b *BaseComponentInstance) SetProps(newProps Props) bool {
    // Props 变化检查
    if propsEqual(b.props, newProps) {
        return false  // 无变化
    }

    b.props = newProps
    return true  // Props 已更新
}
```

### 组件实例 vs 根组件 Context

| 类型 | 管理方式 | 用途 | State 来源 |
|-----|---------|------|-----------|
| **根组件** | `Reconciler.ctx` | 全局状态管理 | `DeclarativeNode.instance.State` |
| **子组件** | `InstanceManager` | Hooks 状态管理 | `BaseComponentInstance.Context.Hooks` |

```go
// beginWorkComponent() 中的选择逻辑
func beginWorkComponent(current, workInProgress *Fiber) *Fiber {
    isRoot := (workInProgress.Key == "root")

    if isRoot && currentReconciler.ctx != nil {
        // 根组件：直接使用 Reconciler 的 rootContext
        ctx = currentReconciler.ctx
        // ✅ ctx.GlobalState = 全局状态（Intent 更新）
    } else if instanceMgr != nil {
        // 子组件：从 InstanceManager 获取/创建实例
        instance = instanceMgr.GetOrCreate(componentKey, ...)
        ctx = instance.GetContext()
        // ✅ ctx.Hooks = useState 值
    }

    SetCurrentContext(ctx)
    // 渲染组件
    renderFunc()
}
```

---

## 性能优化

### 批量更新

为了避免每次 `SetState` 都触发重新渲染，可以实现批量更新队列：

```go
type ComponentContext struct {
    PendingUpdates  map[string]interface{}  // 待更新队列
    UpdateScheduled bool           // 是否已调度
}

func (ctx *ComponentContext) SetGlobal(key string, value interface{}) {
    newState := value
    oldState, exists := ctx.GlobalState[key]

    // 检查变化
    if exists && newState == oldState {
        return  // 无实际变化
    }

    // 加入更新队列
    ctx.PendingUpdates[key] = newState

    // 调度批量处理
    if !ctx.UpdateScheduled && ctx.scheduleUpdate != nil {
        ctx.UpdateScheduled = true
        ctx.scheduleUpdate()
    }
}

// FlushUpdates - 在重新渲染前应用批量更新
func (ctx *ComponentContext) FlushUpdates() {
    if len(ctx.PendingUpdates) == 0 {
        return
    }

    ctx.GlobalStateMu.Lock()
    for k, v := range ctx.PendingUpdates {
        ctx.GlobalState[k] = v
    }
    ctx.GlobalStateMu.Unlock()

    ctx.PendingUpdates = make(map[string]interface{})
    ctx.UpdateScheduled = false
}
```

### Memoization

```go
// 使用 Memo 避免不必要的重渲染
type MemoVNode struct {
    Compare PropsEqual
    // ...
}

func (m *MemoVNode) ShouldUpdate(newProps Props) bool {
    return m.Compare != nil && m.Compare(m.Props, newProps)
}
```

---

## 调试工具

### 状态追踪

```go
// 开启调试日志
os.Setenv("MINT_DEBUG_STATE", "true")

// 日志输出示例：
// [SET] step = 2 (source: button#123)
// [GET] step = 2 (component: App)
// [RENDER] App (hooks: 0, global: 5 entries)
```

### Fiber Inspector

```bash
# 查看当前 Fiber 树结构
mint-inspect fibers --tree

# 查看组件状态
mint-inspect fibers --component "App"

# 查看 InstanceManager
mint-inspect instances --all
```

---

## 总结

### 关键点

1. **VNode 是描述** - 无状态的纯数据结构
2. **Fiber 是实例** - 持有运行时状态
3. **双重状态** - Hooks（局部） + Global（全局）
4. **Intent 驱动** - 所有状态更新通过 Intent
5. **InstanceManager** - 组件实例复用和生命周期管理

### 最佳实践

- ✅ 组件内状态用 `useState`
- ✅ 跨组件状态用 Intent
- ✅ Props 用于父子通信
- ✅ 避免混合使用局部和全局状态
- ✅ 为组件设置明确的 `Key` 确保 Instance 复用

### 相关文档

- [BEST_PRACTICES.md](./BEST_PRACTICES.md) - 状态管理最佳实践
- [PERFORMANCE.md](./PERFORMANCE.md) - 性能优化指南
- [MIGRATION.md](./MIGRATION.md) - 从闭包模式迁移
