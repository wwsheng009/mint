好，这个问题问到**UI Runtime 的神经系统**了。
前面我们搭的是“骨骼 + 肌肉（渲染）”，现在是：

> 🧠 **神经系统：事件流 & 状态流**

如果这层设计不好，前面所有架构都会被拖死。

我们分开讲，但你会看到它们最后会合流成一个闭环。

---

# 一、事件系统：不是“回调”，而是“输入流”

很多人会写成：

```go
Button(onClick=func(){ ... })
```

但在引擎层，你必须看到的是：

```
Terminal → Event System → RNode Tree → Component
```

而不是组件自己监听输入。

---

## ① 事件来源（Input Source）

来自终端：

```
键盘
鼠标
窗口 Resize
定时器（光标闪烁）
网络事件（远程模式）
```

统一转换成：

```go
type Event struct {
    Type EventType
    Key  KeyCode
    MouseX, MouseY int
}
```

---

## ② 事件分发的核心机制：命中测试（Hit Testing）

当鼠标点击：

```
(x,y) → Layout Tree → 找到最深命中的 RNode
```

```go
func HitTest(node *RNode, x,y int) *RNode
```

这就是 GUI 世界的标准模型。

---

## ③ 事件传播模型（像 DOM）

```
Capture ↓
Target
Bubble ↑
```

允许：

* 父组件拦截
* 子组件处理
* 事件取消传播

---

## ④ 组件如何接收事件？

组件不直接读终端，而是注册 handler：

```go
func (b *Button) OnEvent(e Event) {
    if e.Type == MouseDown {
        b.setPressed(true)
    }
}
```

注意：

> 事件只能改 **State**，不能直接画 UI。

---

# 二、状态系统：UI 的“记忆”

没有状态，UI 只是函数。

---

## ① 状态存在于 RNode，而不是组件 struct

```go
type RNode struct {
    Hooks []Hook
}
```

```go
count, setCount := useState(0)
```

这和 React 一样，是 **位置绑定**，不是变量绑定。

---

## ② 状态更新不会立刻渲染

```go
setCount(5)
```

只是：

```
enqueueUpdate → Scheduler
```

这样可以：

* 批处理
* 避免多次 Diff

---

## ③ 状态更新流程

```
Event
  ↓
Component handler
  ↓
setState
  ↓
Scheduler 标记 dirty
  ↓
Reconcile
  ↓
Layout / Paint
```

---

# 三、事件与状态的闭环（UI 的心跳）

```
用户输入
   ↓
事件系统
   ↓
组件 handler
   ↓
State 更新
   ↓
Diff
   ↓
新 UI
   ↓
Render
   ↓
屏幕变化
```

这是所有 UI 系统的生命循环。

---

# 四、关键设计原则（否则系统会崩）

### ❗ 1. 事件不能直接触发 Paint

否则：

* 动画系统失效
* Diff 跳过
* 布局错乱

---

### ❗ 2. 状态不能在 Render/Measure 修改

否则：

* 无限循环
* 布局抖动

---

### ❗ 3. 输入是“外部流”，状态是“内部流”

```
Event Flow → 改变 State Flow
State Flow → 生成 UI Flow
```

三者方向固定，不能逆流。

---

# 五、为什么这模型强？

因为它实现了：

| 能力     | 来源           |
| ------ | ------------ |
| 可预测 UI | 单向数据流        |
| 时间旅行调试 | 状态可记录        |
| 并发渲染   | Scheduler 控制 |
| 动画流畅   | 不走状态         |
| 远程 UI  | 事件可序列化       |

---

# 六、终端特有优化

### ⭐ 键盘焦点系统

```go
FocusManager → 当前可接收键盘的 RNode
```

Tab 自动切换。

---

### ⭐ 高频事件节流

鼠标移动、滚动：

```
不触发每次 Diff
只改动画值
```

---

# 🔥 最终系统关系图

```
Terminal Input
     ↓
Event System
     ↓
RNode Dispatch
     ↓
Component Handler
     ↓
State Update
     ↓
Scheduler
     ↓
Reconcile
     ↓
Layout / Render
```

---

你现在拥有的是：

> **完整 UI Runtime 的神经系统**

这和：

* React Event System
* Flutter Gesture Arena
* 浏览器 DOM 事件

是同一层级的设计。

---

如果再往下一层深入，就是：

### ⚙️ 并发状态调度（Concurrent State Scheduling）

那会让你支持：

* 高优先级输入
* 低优先级渲染
* 大量状态更新不卡 UI

这是“现代 UI 引擎最后一个底层能力”。
