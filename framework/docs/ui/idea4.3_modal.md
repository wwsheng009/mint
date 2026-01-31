好问题，这个点是 **TUI 引擎从“能用”到“像真正 UI 系统”** 的分水岭。

弹出框（Modal / Popup / Dropdown / Tooltip）本质不是“组件”，而是：

> 🧱 **Layer（图层系统）问题**

如果只把它当普通组件，你后面一定会遇到：

* 被其他组件盖住
* 焦点混乱
* 事件穿透
* Clip 错误
* Z-index 地狱

所以要从引擎层解决。

---

# 🧠 一句话核心

> **弹出框 = 脱离父布局树的“顶层渲染子树”**

它属于 UI 树逻辑结构，但**不受父节点布局约束**。

这和 Web 的：

```
position: fixed / portal
```

是同类问题。

---

# 🏗 一、必须引入「Layer 系统」

在 `internal/render` 增加：

```go
type Layer int

const (
    LayerBase Layer = iota   // 正常 UI
    LayerOverlay             // 遮罩
    LayerModal               // 弹窗
    LayerTooltip             // 提示
)
```

每个 RNode 带：

```go
rnode.Layer Layer
```

---

## 渲染顺序变成：

```
for layer := LayerBase → LayerTooltip {
    paint(layer)
}
```

这一步解决：

✅ Z 轴覆盖问题

---

# 🧩 二、Modal 的声明方式（SDK）

```go
ui.Modal(func(){
    ui.Box().
        Width(40).
        Height(10).
        Border(ui.BorderRounded).
        Child(
            ui.Column(
                ui.Text("Confirm?"),
                ui.Button("OK", onOK),
            ),
        )
})
```

内部做的事：

```
创建 RNode
标记 LayerModal
挂到 Runtime.overlayRoot
```

**不是父组件的 children。**

---

# 🧱 三、布局如何处理？

Modal 不参与正常 Layout。

它走 **Overlay Layout Pass**：

```go
screenW, screenH := runtime.ScreenSize()

modal.X = (screenW - modal.W) / 2
modal.Y = (screenH - modal.H) / 2
```

也可以支持：

* Anchor to component
* Cursor position
* Absolute

但都在“Overlay 布局系统”。

---

# 🧠 四、事件系统必须“截断”

当有 Modal 打开时：

```
事件流 = 只发给最上层 Layer
```

实现：

```go
if topLayerActive {
    dispatchTo(layerNodes[top])
    return
}
```

这解决：

✅ 事件穿透
✅ 背景组件误响应

---

# 🎯 五、Focus 管理必须切换 Scope

当 Modal 出现：

```
FocusScope = modal subtree
```

关闭后恢复。

否则 Tab 会跑到背景组件。

---

# 🎨 六、遮罩（Dim 背景）

通常 Modal 还需要：

```go
ui.Overlay(
    ui.Box().Bg(ui.ColorDim)
)
```

这个也是一个 LayerOverlay 节点，画在 Base 上、Modal 下。

---

# 🧩 七、Clip 与弹层

普通组件 Clip 来自父节点。
Modal Clip 来自：

```
屏幕边界
```

否则会被父容器裁掉。

---

# 🧠 八、Dropdown / Tooltip 也是同一机制

| 类型       | Layer   | 布局方式            |
| -------- | ------- | --------------- |
| Modal    | Modal   | 居中              |
| Dropdown | Overlay | Anchor          |
| Tooltip  | Tooltip | Cursor / Anchor |

不是不同系统，只是 Layer + 定位策略不同。

---

# 🚨 绝对不要这样做

❌ 把 Modal 当作 Column 的 child
❌ 用 z-index 数字在同一树里排序
❌ 让 Modal 参与普通 Layout

这会导致：

> 后期全面推翻渲染管线

---

# 🧠 正确的引擎结构应该是：

```
UI Tree（逻辑）
   ↓ Reconcile
RNode Tree
   ├── Base Layer Tree
   ├── Overlay Tree
   ├── Modal Tree
   └── Tooltip Tree
        ↓
    Layout (不同策略)
        ↓
    Paint (按 Layer 顺序)
```

---

# 🎯 最终效果

你就具备：

* Modal
* Dropdown
* Tooltip
* Context Menu
* Toast

全部统一在 Layer 系统里。

这在 TUI 领域是**高级引擎级能力**，不是普通库能力。

---

如果继续深入，下一步可以讲：

> 🧩 **Modal 动画、焦点陷阱（Focus Trap）、ESC 关闭机制**

这是弹层系统的“交互完成度”层。

好，现在进入弹层系统的**高级交互完整形态**，这一步做完，你的 Modal 体验就已经是“现代 UI 框架级”。

我们解决 4 件事：

> 🎭 动画
> 🎯 Focus Trap
> ⎋ ESC 关闭
> 🧠 背景冻结

---

# 🎞 一、Modal 动画（不能影响主树性能）

关键原则：

> **动画只影响 Paint，不影响 Layout**

---

## 1️⃣ SDK 使用形态

```go
ui.Modal(func() {
    ui.Box().
        Width(40).
        Height(10).
        Border(ui.BorderRounded).
        Animate(ui.FadeIn | ui.ScaleIn).
        Child(...)
})
```

---

## 2️⃣ 引擎做的事

Modal RNode 增加：

```go
type OverlayState struct {
    Opacity float64
    Scale   float64
}
```

Scheduler 每帧更新：

```
PaintDirty only
```

LayoutBox 不变。

---

## 3️⃣ Paint 阶段处理

```go
drawBox(x, y, w*scale, h*scale, opacity)
```

这保证：

✅ 动画流畅
✅ 不触发布局抖动
✅ 不增加 Diff 压力

---

# 🎯 二、Focus Trap（最关键 UX）

当 Modal 打开时：

> 焦点必须“困在弹窗内”

否则 Tab 会跑到背景 UI。

---

## 实现机制

引入：

```go
type FocusScope struct {
    Root *RNode
}
```

Runtime 维护栈：

```go
focusScopes []FocusScope
```

Modal 打开：

```
push(scopeOfModal)
```

关闭：

```
pop()
```

---

## Tab 键逻辑变为：

```
只在当前 FocusScope 子树中循环
```

---

# ⎋ 三、ESC 自动关闭机制

Modal 是“短生命周期 Layer”。

需要统一行为：

```go
ui.Modal(...).CloseOnEsc(true)
```

引擎层：

```go
if key == ESC && topLayer.CloseOnEsc {
    runtime.CloseTopModal()
}
```

这是 UI 直觉行为。

---

# 🧠 四、背景冻结（交互隔离）

当 Modal 打开：

```
Base Layer 仍渲染
但不响应事件
```

事件系统改为：

```go
if modalActive {
    dispatch only to modal layer
    return
}
```

并且：

* 禁止 Base Layer setState（可选）
* 暂停背景动画（可选）

---

# 🌫 五、遮罩淡入淡出

Overlay 层可以有独立动画：

```go
overlayOpacity 0 → 0.6
```

让 Modal 更有层次。

---

# 🧩 六、关闭流程必须是“动画安全”的

关闭不是直接销毁节点，而是：

```
State = Closing
播放 fade-out
动画结束 → 真正销毁 RNode
```

否则 UI 会“瞬间消失”。

---

# 🎯 七、完整弹层系统状态机

```text
Idle
  ↓ open()
Opening (fade in)
  ↓ done
Active
  ↓ esc/close()
Closing (fade out)
  ↓ done
Destroyed
```

---

# 🧠 八、为什么这套机制高级？

因为它解决的是：

| 问题    | 解决方式          |
| ----- | ------------- |
| 视觉层级  | Layer         |
| 交互隔离  | Event gating  |
| 焦点混乱  | FocusScope    |
| 动画卡顿  | Paint-only 动画 |
| UI 闪烁 | 延迟销毁          |

这是浏览器 / Flutter / React Native 同级设计。

---

# 🔥 现在你的 TUI 引擎已经支持：

* Modal
* Dropdown
* Tooltip
* Context Menu
* Toast

全部具备：

* 动画
* 焦点管理
* ESC 关闭
* 背景冻结

这在终端领域是**顶级 UI 能力集合**。

---

如果继续下一步，就进入：

> 🧠 **复杂输入组件（Input / Form / IME / 光标管理）**

这是另一个“引擎级难点区”。
