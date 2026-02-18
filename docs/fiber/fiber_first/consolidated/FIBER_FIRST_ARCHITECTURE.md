# Fiber-First 统一架构文档

> **核心原则**：Fiber 管结构，Action 管行为，VNode 管声明

## 目录

1. [架构概览](#架构概览)
2. [核心角色](#核心角色)
3. [数据流图](#数据流图)
4. [VNode到Fiber迁移](#vnode到fiber迁移)
5. [事件系统](#事件系统)
6. [Paint架构](#paint架构)
7. [布局系统](#布局系统)
8. [Action Runtime](#action-runtime)
9. [实施Checklist](#实施checklist)
10. [重构计划](#重构计划)

---

## 架构概览

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
│  paint.PaintableBox (渲染层)                                 │
│  - Instance 运行时实体                                       │
│  - Paint 绘制逻辑                                            │
│  - 状态管理                                                  │
└─────────────────────────────────────────────────────────────┘
```

---

## 核心角色

### 三种角色对比

| 层级 | 角色 | 作用 | 是否运行期存在 |
|------|------|------|--------------|
| VNode | 描述 | render() 的返回值，描述"下一帧想要什么" | ❌ 不长期存在 |
| Fiber | 结构节点 | 持久化节点，参与调度、diff、layout、commit | ✅ |
| Instance (paint.PaintableBox) | 行为+状态 | 组件运行时实体，持有状态和渲染逻辑 | ✅ |

### Fiber-first 核心原则

> **VNode = Description (created every render)**
> **Instance = Runtime Entity (persists across renders)**
> **Fiber = Structure (scheduling node)**

---

## 数据流图

### 整体数据流（最终形态）

```
             ┌──────────────┐
             │     VNode     │   ← 只存在于 render 阶段
             └──────┬────────┘
                    │
                    ▼
            ┌────────────────┐
            │   Reconciler    │
            │  (build Fiber)  │
            └──────┬──────────┘
                    │
                    ▼
            ┌────────────────┐
            │      Fiber      │  ← 单一真实树
            │------------------│
            │ Type             │
            │ Key              │
            │ MemoizedProps    │
            │ MemoizedState    │
            │ Style            │
            │ Child/Sibling    │
            │ EffectTag        │
            │ Instance         │ → paint.PaintableBox
            └──────┬──────────┘
                    │
                    ▼
        ┌────────────────────────┐
        │     Layout Engine       │
        │ (pure function phase)   │
        └──────┬──────────────────┘
                    │
                    ▼
            ┌────────────────┐
            │ paint.PaintableBox │  ← 布局结果树
            │------------------│
            │ NodeID           │
            │ X,Y,W,H          │
            │ Layer            │
            │ ZIndex           │
            └──────┬──────────┘
                    │
                    ▼
            ┌────────────────┐
            │ Render Plane    │
            └────────────────┘
```

### 事件流程

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

---

## VNode到Fiber迁移

### 必须迁移字段（强制）

#### 1️⃣ Type
```go
vnode.Type → fiber.Type
```

#### 2️⃣ Key
```go
vnode.Key → fiber.Key
```

#### 3️⃣ Props
```go
vnode.Props() → fiber.MemoizedProps
```

#### 4️⃣ Style（重点）
必须在 completeWork 阶段：
```go
fiber.Style = extractStyle(vnode.Props)
```

包含：
- flexDirection
- flexGrow
- width / height
- margin
- padding
- border
- display
- position
- zIndex

之后：
```go
Layout 只读 fiber.Style
```

#### 5️⃣ Text 内容
如果是文本节点：
```go
fiber.MemoizedState = vnode.Text
```

#### 6️⃣ Event Handler
```go
fiber.EventHandlers = vnode.Events
```

### 可选迁移字段

#### 7️⃣ Layer / Portal 标志
```go
fiber.LayerTag
fiber.IsPortal
```

#### 8️⃣ Ref
```go
fiber.Ref
```

---

## 事件系统

### ActionBridge 设计

ActionBridge 是唯一知道 Fiber 和 Router 的模块。

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

### 模块边界

#### ✅ 允许的访问

| 模块 | 可访问对象 |
|------|-----------|
| Reconciler | VNode + Fiber |
| Layout | Fiber (只读 paint.PaintableBox) |
| EventSystem | Fiber (只读，定位目标) |
| ActionBridge | Fiber + Dispatcher |
| Dispatcher | ActionTarget (组件) |
| FocusManager | Fiber (只读 FocusableMeta) |

#### ❌ 永久禁止

```
EventSystem → VNode
EventSystem → 直接调用回调
Dispatcher → Fiber
Action → Fiber
Fiber → 业务回调
```

### 组件实现

#### 方式一：语义 Action（推荐）

```go
// 1. 定义语义 Action
const ActionOpenModal action.ActionType = "open_modal"

// 2. 组件声明
button.SetClickAction(ActionOpenModal)

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
```

---

## Paint架构

### 核心设计原则

**VNode = Description (created every render)**
**Instance = paint.PaintableBox (Runtime Entity)**

### 架构流程

```
用户代码创建 ButtonVNode
       ↓
CreateFiber(vnode)
       ↓
vnode.CreateInstance() → 创建 ButtonInstance
       ↓
Fiber.Instance = instance (持久化)
       ↓
VNode 被丢弃，不再被引用
       ↓
Paint阶段: instance.Paint(x, y)
```

### ComponentInstance 接口

```go
type ComponentInstance interface {
    // Lifecycle
    Init(props Props)
    Destroy()
    OnMount()
    OnUnmount()

    // Props Management
    SetProps(props Props) bool
    GetProps() Props

    // Identification
    Key() string
    SetKey(key string)

    // Dirty flag
    MarkDirty()
    IsDirty() bool
}
```

### paint.PaintableBox 接口

```go
// For components that can paint
type PaintableBox interface {
    ComponentInstance
    Paint(x, y int) []paint.DrawCmd
}

// For components that can receive focus
type FocusableInstance interface {
    ComponentInstance
    SetFocus(focused bool)
    HasFocus() bool
    IsDisabled() bool
}

// For components that handle actions
type ActionHandlerInstance interface {
    ComponentInstance
    CanHandleAction(actionType string) bool
    HandleAction(actionType string, payload interface{}) bool
}
```

### 实现示例

#### ButtonVNode (Description)

```go
type ButtonVNode struct {
    *ui.ElementVNode
    label   string
    variant ButtonVariant
    size    ButtonSize
}

func (b *ButtonVNode) CreateInstance() rtui.ComponentInstance {
    props := rtui.Props{
        "label":   b.label,
        "variant": b.variant,
        "size":    b.size,
    }
    return NewButtonInstance(props)
}
```

#### ButtonInstance (paint.PaintableBox)

```go
type ButtonInstance struct {
    // Props from VNode
    label   string
    variant ButtonVariant
    size    ButtonSize

    // Runtime state
    hasFocus  bool
    isHovered bool
    disabled  bool
}

func (inst *ButtonInstance) Paint(x, y int) []paint.DrawCmd {
    // Use instance state (hasFocus, isHovered, etc.)
    // No VNode dependency!
}
```

---

## 布局系统

### 核心原则

1️⃣ Layout 只读 Fiber
```
Fiber → paint.PaintableBox
```

绝不能：
```
VNode → Layout
```

2️⃣ Layout 不修改 Fiber
```
Fiber (immutable during layout)
```

Layout 必须是：
> 纯计算阶段

3️⃣ Render 不读 VNode
```
Render → paint.PaintableBox
```

VNode 在 commit 后可以丢弃。

### 最小可运行 Fiber 定义

```go
type Fiber struct {
    // -------------------------
    // Identity
    // -------------------------
    Type ElementType
    Key  string

    // -------------------------
    // Tree
    // -------------------------
    Parent    *Fiber
    Child     *Fiber
    Sibling   *Fiber
    Alternate *Fiber

    // -------------------------
    // Data
    // -------------------------
    PendingProps  any
    MemoizedProps any
    MemoizedState any

    // -------------------------
    // Layout Input
    // -------------------------
    Style Style
    Instance paint.PaintableBox  // 运行时实体

    // -------------------------
    // Effects
    // -------------------------
    Flags        Flags
    SubtreeFlags Flags
    NextEffect   *Fiber

    // -------------------------
    // Scheduling
    // -------------------------
    Lanes      Lane
    ChildLanes Lane
}
```

---

## Action Runtime

### 统一 Action Runtime 重构

**目标架构**:
```
Input
  ↓
InputProcessor
  ↓
Action (永远不返回 nil)
  ↓
ActionBridge
  ↓
ScopeDispatcher / Router
  ↓
Component Logic
```

**核心原则**：
- 单一事件通道 ✅
- 无 legacy 路径 ✅
- Closure → ActionID 注册 ✅
- Fiber 只存 ActionTargetID ✅

### Action 优先级 → Lane 映射

```go
type Lane uint32

const (
    LaneSync Lane = 1 << iota       // 输入、焦点、键盘
    LaneUserBlocking                // 点击、短反馈
    LaneNormal                      // 普通更新
    LaneTransition                  // 页面切换
    LaneIdle                        // 低优先级
)
```

---

## 实施Checklist

### ✅ 结构完整性检查

- [ ] VNode 不参与运行期
- [ ] Instance 独立
- [ ] Fiber 只做结构
- [ ] 无跨层访问

### ✅ 事件系统检查

- [ ] 冒泡
- [ ] 捕获
- [ ] 子树隔离
- [ ] Action Bridge

### ✅ 状态模型检查

- [ ] 统一交互状态
- [ ] 行为抽象层
- [ ] 无 closure 泄漏
- [ ] 状态驱动视觉

### ✅ 内存模型检查

- [ ] 双缓冲
- [ ] Instance 复用
- [ ] Fiber 轻量化
- [ ] 无 VNode clone

---

## 重构计划

### Phase 1 — 数据迁移
1. 定义 Fiber.Style
2. completeWork 中填充 Style
3. 删除 Layout 中 vnode 访问
4. 完成标志：删除 paint.PaintableBox.VNode

### Phase 2 — Layout 纯函数化
- buildComputedBoxFromFiber 变成纯函数
- 禁止：`fiber.Width = ...`
- 允许：`return paint.PaintableBox`

### Phase 3 — 双树稳定
- 实现：`current ↔ workInProgress`
- 保证：Layout 只对 workInProgress
- Commit 原子交换

### Phase 4 — Effect 驱动 Commit
- 禁止：`Render 遍历全树`
- 必须：`EffectList → Commit`

### Phase 5 — Scheduler 接入
- Lane 优先级
- 批处理
- 时间切片

### Phase 6 — 删除 VNode 持久化
- VNode 只存在于：`render()`
- 之后立即丢弃
- Fiber 成为唯一真实树

---

## 成功标准

### 判断是否 Fiber-first 成功

问自己一个问题：
> 如果我删除 VNode struct
> 整个 Layout + Render 是否还能运行？

如果答案是 **YES**，你就真正完成了 Fiber-first。

### 工业级标准形态

完成之后你会拥有：
- ✅ 纯 Fiber 树
- ✅ 可并发 reconcile
- ✅ 可时间切片
- ✅ 可 Suspense
- ✅ 可 Portal
- ✅ 可多 Layer

结构上等价于：**React Fiber 架构**（但更适合 TUI）

---

## 一句话总结

> **Fiber 只存 "我是谁" (ActionTargetID)**
> **ActionBridge 负责 "怎么路由"**
> **paint.PaintableBox 决定 "做什么" (HandleAction)**
> **组件决定 "怎么画" (Paint)**
