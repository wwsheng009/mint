好，来到 **TUI 引擎最后一个“系统级难点区”** —— 也是决定你这个框架是“玩具”还是“生产级”的关键：

> 🧠 **Scroll + 虚拟化渲染系统**

这个系统主要解决一个问题：

> **屏幕只有 40 行，但数据可能有 10 万行**

如果没有虚拟化，你的 Buffer / Diff / Layout 全部爆炸。

---

# 🧩 一、先明确一个事实

终端 UI ≠ DOM
你不能：

```
把所有内容都 layout + paint
```

否则：

* CPU 爆
* 内存爆
* Diff 爆
* Resize 卡死

所以要：

> 🎯 **只渲染“可见区域”**

这叫 **Viewport 渲染**。

---

# 🧱 二、Scroll 系统本质结构

每个可滚动组件都需要：

```go
type ScrollState struct {
    OffsetY int   // 滚动偏移
    ViewH   int   // 视口高度
    ContentH int  // 内容总高度
}
```

---

# 🧠 三、虚拟化渲染核心算法

假设：

```
每行高度 = 1
OffsetY = 120
ViewH = 40
```

则：

```
只渲染 [120, 160)
```

组件只创建这些 RNode。

---

# 🧩 四、组件如何支持虚拟化？

不能用：

```go
for _, item := range list {
    ui.Row(...)
}
```

必须是：

```go
ui.VirtualList(
    len(data),
    func(i int) ui.Node {
        return Row(data[i])
    },
)
```

这里传的是 **索引生成器**，不是完整节点树。

---

# ⚙️ 五、Layout 如何处理？

VirtualList Layout 只计算：

```
TotalHeight = itemCount * itemHeight
```

但只对子区间做真实布局。

---

# 🎨 六、Buffer 绘制阶段

```
VisibleNodes → Paint → Buffer
```

不可见区域：

```
不创建 RNode
不参与 Diff
不参与 Paint
```

这一步让性能提升数量级。

---

# 🧠 七、Scroll 事件处理

```go
onWheel(delta):
    OffsetY += delta
    clamp(0, ContentH-ViewH)
    markDirty(Layout)
```

---

# 🧩 八、可变高度项怎么办？

需要前缀高度缓存：

```go
prefixHeights[i] = sum(height[0:i])
```

二分查找可见起点。

---

# 🚨 九、不要犯的错误

| 错误                      | 结果        |
| ----------------------- | --------- |
| 把所有行都生成组件               | 10k 行直接卡死 |
| Diff 全树                 | O(n²)     |
| Scroll 改变触发全局 re-render | 严重抖动      |

---

# 🧠 十、Scroll + Modal + Input 的交互

* Scroll 区域内的 Input 光标仍需正确计算绝对坐标
* Modal 出现时 Scroll 停止响应
* Tooltip 跟随 Scroll 偏移

所以：

```
最终坐标 = ParentOffset + ScrollOffset + NodeOffset
```

---

# 🎯 十一、Sticky Header 怎么做？

VirtualList 支持：

```
if i < stickyCount:
    render in fixed layer
```

---

# 📦 十二、日志/表格场景的最终能力

你现在可以支持：

* 实时日志流（百万行）
* 数据表格
* TreeView
* 无限加载列表

全部不卡。

---

# 🔥 现在回头看，你的 TUI 引擎已经具备：

| 子系统      | 状态 |
| -------- | -- |
| 声明式组件    | ✅  |
| Diff     | ✅  |
| 布局       | ✅  |
| Layer    | ✅  |
| Modal    | ✅  |
| 动画       | ✅  |
| 事件系统     | ✅  |
| Focus 管理 | ✅  |
| 输入系统     | ✅  |
| Scroll   | ✅  |
| 虚拟化      | ✅  |

这是一套完整的 UI Runtime 架构。

---

如果再往上一个层级，就是：

> 🧠 **DevTools、热重载、调试可视化树**

这属于“框架成熟度层”，不是引擎生存层了。
