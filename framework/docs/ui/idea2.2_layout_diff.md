好，进入 **布局系统真正决定性能的部分**：

> 🧠 **Dirty Layout 优化（只重新计算“受影响区域”）**

没有这个，你的布局算法再完美，UI 一复杂就卡。

---

# 🎯 核心问题

现在流程是：

```
setState → Render → Reconcile → Layout 整棵树
```

如果树有 5000 个节点，即使只改了一个按钮：

> ❌ 你也在全树算布局

这就是很多 UI 框架“理论优雅，实际卡顿”的原因。

---

# 🧱 一、为每个节点引入“脏标记”

```go
type RNode struct {
    LayoutDirty bool
    PaintDirty  bool
}
```

---

# 🧠 二、哪些变化会导致 Layout Dirty？

| 变化                | 是否 LayoutDirty |
| ----------------- | -------------- |
| Width / Height 改变 | ✅              |
| Padding / Border  | ✅              |
| FlexGrow 改变       | ✅              |
| 文字内容变化            | 可能             |
| 颜色变化              | ❌（只 Paint）     |

---

# 🔁 三、脏标记必须向上传播

因为子节点尺寸变化会影响父容器分配：

```go
func markLayoutDirty(node *RNode) {
    if node.LayoutDirty { return }
    node.LayoutDirty = true
    markLayoutDirty(node.Parent)
}
```

---

# 🧩 四、Layout 阶段变成“选择性递归”

```go
func layout(node *RNode) {
    if !node.LayoutDirty {
        return
    }

    computeLayout(node)

    for child in node.Children {
        layout(child)
    }

    node.LayoutDirty = false
}
```

---

# ⚡ 五、但还可以更狠：**剪枝**

如果父节点尺寸没变：

```
子节点布局可跳过
```

条件：

```
parent.size unchanged
AND
child has no size-affecting changes
```

---

# 📦 六、Scroll 区域的特殊优化

虚拟列表只 Layout：

```
visible items
```

不可见部分：

```
直接跳过 Layout
```

这一步是万行日志不卡的核心。

---

# 🧠 七、Paint Dirty 与 Layout Dirty 分离

很多变化只需要重绘：

| 操作       | Layout | Paint |
| -------- | ------ | ----- |
| Hover 状态 | ❌      | ✅     |
| 光标闪烁     | ❌      | ✅     |
| 文本颜色变    | ❌      | ✅     |

这样动画就不会触发布局。

---

# 🔥 八、最终布局调度流程变成

```
setState
  ↓
Reconcile（标记 Dirty）
  ↓
Layout(只处理 LayoutDirty 子树)
  ↓
Paint(只处理 PaintDirty)
```

---

# 🧠 九、再进阶：区域布局裁剪（Viewport Culling）

如果一个节点完全在屏幕外：

```
不 Layout 子树
```

适用于：

* Scroll 超长内容
* 折叠面板

---

# 📊 十、优化前后差异

| 场景       | 无优化       | 有 Dirty Layout |
| -------- | --------- | -------------- |
| 按钮 hover | 全树 Layout | 0 布局           |
| Modal 打开 | 全树        | 只 Overlay      |
| 输入字符     | 全树        | Input 子树       |
| 日志滚动     | 全树        | 可见区域           |

---

# 🧠 这一步意味着

你的引擎已经进入：

> **“增量布局引擎”级别**

这是浏览器、Flutter、React Native 的核心能力之一。

---

到现在为止，你已经构建了：

* 渲染管线
* 组件系统
* 布局系统（Flex + Grid）
* Layer 系统
* 状态系统
* 事件系统
* 输入系统
* Scroll 虚拟化
* 调度并发
* 样式系统
* 输出优化
* **增量布局优化**

这已经是一套完整 UI Runtime 内核。

---
