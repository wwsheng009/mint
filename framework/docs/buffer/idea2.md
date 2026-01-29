好，现在来到最后一块真正让框架“活起来”的核心层：

> 🧠 **Event Dispatch Tree + Focus System + State Reconciliation**

这一步完成后，你的项目正式从“渲染引擎”进化为：

> 💎 **完整终端 UI 应用框架（Terminal App Framework）**

---

# 🧠 一、事件分发树（像浏览器那样工作）

现在不能再是：

```
root.HandleEvent()
```

而是：

> **命中测试 → 捕获 → 目标 → 冒泡**

---

## 1️⃣ 事件结构

```go
type Event struct {
    Type EventType
    X, Y int        // 鼠标
    Key  KeyCode    // 键盘
}
```

---

## 2️⃣ 命中测试（Hit Test）

从 **最高 ZIndex layer** 开始：

```go
func hitTest(node *Node, x, y int) *Node {
    if !node.Rect.Contains(x,y) {
        return nil
    }

    for i := len(node.Kids)-1; i >= 0; i-- {
        if hit := hitTest(node.Kids[i], x, y); hit != nil {
            return hit
        }
    }
    return node
}
```

---

## 3️⃣ 事件传播阶段

```
Capture ↓
Target
Bubble ↑
```

```go
func dispatch(ev Event) {
    target := hitTest(topLayer.Root, ev.X, ev.Y)

    path := buildPathToRoot(target)

    // Capture
    for _, n := range path {
        n.OnCapture(ev)
    }

    // Target
    target.OnEvent(ev)

    // Bubble
    for i := len(path)-1; i >= 0; i-- {
        path[i].OnBubble(ev)
    }
}
```

这让你实现：

| 行为      | 实现      |
| ------- | ------- |
| 父容器拦截事件 | Capture |
| 子控件响应   | Target  |
| 向上传播    | Bubble  |

---

# 🎯 二、Focus System（键盘世界核心）

终端 UI 的灵魂不是鼠标，而是 **焦点系统**

---

## Focus Manager

```go
type FocusManager struct {
    focused *Node
}
```

---

## 设置焦点

```go
func (f *FocusManager) SetFocus(n *Node) {
    if f.focused != nil {
        f.focused.OnBlur()
    }
    f.focused = n
    n.OnFocus()
}
```

---

## 键盘事件走向

```
KeyEvent → Focused Node
```

```go
func handleKey(ev KeyEvent) {
    if focusManager.focused != nil {
        focusManager.focused.OnKey(ev)
    }
}
```

---

## Tab 导航

构建可聚焦节点链表：

```go
func nextFocusable(curr *Node) *Node
```

---

# 🧠 三、状态协调系统（让 UI 自动更新）

否则你会陷入：

```
改状态 → 手动 repaint
```

我们需要：

> **声明式状态驱动 UI**

---

## 组件状态模型

```go
type Component struct {
    state map[string]any
    dirty bool
}
```

---

## setState

```go
func (c *Component) SetState(key string, val any) {
    c.state[key] = val
    c.dirty = true
    engine.RequestRepaint()
}
```

---

## Reconciliation（轻量版 Virtual DOM 思想）

每帧：

```go
func update(node *Node) {
    if node.Component.dirty {
        node.Component.Render()
        node.Component.dirty = false
    }

    for _, kid := range node.Kids {
        update(kid)
    }
}
```

这实现：

| 行为     | 结果            |
| ------ | ------------- |
| 状态变化   | 自动刷新          |
| UI 描述式 | 不用手动操作 buffer |
| 局部更新   | 性能高           |

---

# 🧠 四、完整事件流

```
Input (Terminal)
   ↓
Event Parser
   ↓
Layer Hit Test
   ↓
Dispatch Tree
   ↓
Component State Change
   ↓
Mark Dirty
   ↓
Next Frame
   ↓
Reconcile → Layout → Paint → Render
```

这就是现代 UI 引擎闭环。

---

# 🏁 你现在已经构建的系统等级

```
Terminal Driver
Input Parser
Event System
Focus System
State Reconciliation
Layer System
Layout Engine
Paint System
Buffer
Diff Engine
Renderer
Frame Scheduler
```

这已经是：

> 🧠 **完整 UI 渲染引擎架构（只是输出设备是终端）**

---

# 🚀 这套东西能做什么级别的产品？

* 终端 IDE
* 终端数据库管理器
* 终端监控面板
* 类 Electron 的终端应用生态

---

如果再往上，就是生态层了：

### 🧩 组件库系统

### 🎨 主题系统

### 🔌 插件系统

### 🌍 多窗口 / 远程渲染

---

你现在已经不在“写 TUI”，而是在：

> **实现一个终端图形操作系统的 UI 子系统**

要继续冲 **组件系统设计（终端版 Material UI）** 吗？
那是“框架生态层”的开始。
