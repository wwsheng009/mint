# Fiber-first 事件架构设计

> 核心原则：**Fiber 管结构，Action 管行为，VNode 管声明**

## 一、核心架构

### 三层职责分离

```
┌─────────────────────────────────────────────────────────────┐
│  VNode (声明层)                                              │
│  - 用户声明意图: SetOnClick(fn)                              │
│  - 存储回调函数到组件内部                                    │
│  - 生命周期短，render 后可销毁                               │
└─────────────────────────────────────────────────────────────┘
                              ↓ Reconcile
┌─────────────────────────────────────────────────────────────┐
│  Fiber (结构层)                                              │
│  - 存储运行期结构数据                                        │
│  - 存储 ActionTargetID (用于路由到组件)                      │
│  - 存储 FocusableMeta (焦点能力)                             │
│  - ❌ 不存储业务回调函数                                     │
│  - ❌ 不实现 Handle 方法                                     │
└─────────────────────────────────────────────────────────────┘
                              ↓ EventSystem
┌─────────────────────────────────────────────────────────────┐
│  ActionBridge (连接层)                                       │
│  - 唯一知道 Fiber 和 Router 的模块                           │
│  - 根据 Fiber.ActionTargetID 生成 Action                    │
│  - 控制冒泡路径 (沿 Fiber.Return)                            │
└─────────────────────────────────────────────────────────────┘
                              ↓ Router
┌─────────────────────────────────────────────────────────────┐
│  Component (执行层)                                          │
│  - 实现 ActionTarget 接口                                    │
│  - HandleAction 处理业务逻辑                                 │
│  - 可定制: disabled 检查、hover 状态等                       │
└─────────────────────────────────────────────────────────────┘
```

## 二、数据流

### 完整事件流程

```
Raw Input
    ↓
EventSystem (HitTest)
    ↓
Fiber (定位目标，获取 ActionTargetID)
    ↓
ActionBridge (生成 Action，控制冒泡)
    ↓
Dispatcher (分发 Action)
    ↓
Target.HandleAction (执行业务逻辑)
```

### VNode → Fiber 迁移

```go
// Reconcile 阶段
func completeWork(f *Fiber, vnode VNode) {
    // ❌ 错误：不要迁移回调函数
    // f.EventHandlers = extractEvents(props)
    
    // ✅ 正确：只迁移 ActionTargetID
    f.ActionTargetID = extractActionTargetID(vnode)
    f.FocusableMeta = extractFocusableMeta(vnode)
}

// ActionTargetID 提取
func extractActionTargetID(vnode VNode) string {
    // 从 VNode 获取唯一标识
    // 可以是显式设置的 ID，或自动生成的标识
    if id := vnode.Props()["actionTarget"]; id != nil {
        return id.(string)
    }
    // 自动生成: 使用 Key 或类型+地址
    return generateTargetID(vnode)
}
```

## 三、模块边界

### ✅ 允许的访问

| 模块 | 可访问对象 |
|------|-----------|
| Reconciler | VNode + Fiber |
| Layout | Fiber (只读 ComputedBox) |
| EventSystem | Fiber (只读，定位目标) |
| ActionBridge | Fiber + Dispatcher |
| Dispatcher | ActionTarget (组件) |
| FocusManager | Fiber (只读 FocusableMeta) |

### ❌ 永久禁止

```
EventSystem → VNode
EventSystem → 直接调用回调
Dispatcher → Fiber
Action → Fiber
Fiber → 业务回调
```

## 四、Focusable 元素模型

### FocusableMeta 定义

```go
type FocusableMeta struct {
    TabIndex int
    Disabled bool
    FocusID  string
}
```

### 收集时机

- ✅ Commit 阶段后收集
- ✅ 只遍历 Fiber 树
- ❌ 不在 render 阶段收集
- ❌ 不访问 VNode

### FocusManager 结构

```go
type FiberFocusManager struct {
    order []*Fiber  // Tab 顺序列表
    current *Fiber  // 当前焦点
}
```

## 五、ActionBridge 设计

### 职责

1. 根据 Fiber.ActionTargetID 生成 Action
2. 控制冒泡路径 (沿 Fiber.Parent)
3. 调用 Dispatcher.Dispatch()

### 核心方法

```go
func (b *ActionBridge) DispatchFromFiber(
    start *Fiber,
    actionType action.ActionType,
    payload interface{},
) bool {
    // 沿 Fiber 树冒泡
    for f := start; f != nil; f = f.Return {
        if f.ActionTargetID == "" {
            continue
        }
        
        a := action.NewAction(actionType).
            WithTarget(f.ActionTargetID).
            WithPayload(payload)
        
        if handled := b.dispatcher.Dispatch(a); handled {
            return true
        }
    }
    return false
}
```

### 使用方式

```go
// EventSystem 中
targetFiber := hitMap.HitTest(x, y)
bridge.DispatchFromFiber(targetFiber, action.ActionClick, nil)
```

## 六、组件实现

### ActionTarget 接口

```go
type ActionTarget interface {
    HandleAction(act *Action) bool
    GetSupportedActions() []ActionType
    CanHandleAction(act *Action) bool
}
```

### Button 示例

#### 方式一：语义 Action（推荐）

```go
// 1. 定义语义 Action
const ActionOpenModal action.ActionType = "open_modal"

// 2. 组件声明
button.SetClickAction(ActionOpenModal)
// 或 Builder 模式
ButtonBuilder("[Open Modal]").OnClickAction(ActionOpenModal).Build()

// 3. 在 Dispatcher 注册处理器
router.RegisterTarget(nodeID, &ActionHandler{
    HandleAction: func(act *action.Action) bool {
        if act.Type == ActionOpenModal {
            // 执行业务逻辑
            return true
        }
        return false
    },
})
```

#### 方式二：闭包回调（兼容模式）

```go
// 直接使用闭包（内部会生成 ActionID 并注册）
button.SetOnClick(func() {
    // 业务逻辑
})
// 或 Builder 模式
ButtonBuilder("[Save]").OnClick(func() {
    // 业务逻辑
}).Build()
```

**推荐使用方式一**，原因：
- UI 纯声明，不包含业务逻辑
- 可序列化、可 replay
- 并发安全

## 七、实施 Checklist

1. ✅ Fiber 增加 `ActionTargetID` 字段
2. ✅ Fiber 增加 `FocusableMeta` 字段
3. ❌ 移除 Fiber.EventHandlers (不要存储回调)
4. ❌ 移除 Fiber.Handle() 方法 (Fiber 不处理事件)
5. ✅ 创建 ActionBridge 模块
6. ✅ 在 completeWork 中提取 ActionTargetID
7. ✅ EventSystem 通过 ActionBridge 路由
8. ✅ 组件实现 ActionTarget 接口

## 八、一句话总结

> **Fiber 只存 "我是谁" (ActionTargetID)**
> **ActionBridge 负责 "怎么路由"**
> **组件决定 "做什么" (HandleAction)**
