好，最后一步，把这个“宏大平台”压缩成：

> 📅 **v0.1 第一周真实开发执行表**

目标不是“写完”，而是：

> **7 天内把骨架跑起来，让屏幕上真的有 UI 在动**

这是从架构 → 现实的关键跃迁。

---

# 🎯 第一周总目标

到第 7 天结束，你应该能运行：

```bash
go run examples/counter
```

看到：

* Text
* Button
* useState 更新
* Row/Column 布局
* 正确渲染到终端

哪怕代码很丑，都没关系。

---

# 🗓 Day 1 — Runtime 骨架

### 目标：终端能被你控制

### 必做

* [ ] 创建 `internal/runtime`
* [ ] 主循环结构：

```go
for {
  processEvents()
  if dirty {
     renderFrame()
  }
}
```

* [ ] 初始化终端：

  * raw mode
  * hide cursor
  * alt screen

* [ ] 程序退出时恢复终端

### 成功标准

程序启动后屏幕清空，退出后终端恢复正常。

---

# 🗓 Day 2 — Buffer + 最小 Renderer

### 目标：往终端稳定画字符

### 必做

* [ ] `internal/buffer/Cell`
* [ ] `Buffer{Width,Height,Cells}`
* [ ] 简单 `DrawText(x,y,string)`
* [ ] 整屏 flush（先不做 diff）

### 成功标准

能在指定坐标画字，不闪、不乱。

---

# 🗓 Day 3 — VNode + View 构建

### 目标：声明式 UI 树

### 必做

* [ ] `VNode` 结构
* [ ] `ui.View(func(){})`
* [ ] `ui.Text()` 返回 VNode
* [ ] 构建 VNode Tree（还不 Reconcile）

### 成功标准

能把 UI 声明转成一棵树并打印调试。

---

# 🗓 Day 4 — RNode + Reconcile v1

### 目标：VNode → RNode

### 必做

* [ ] `RNode` 结构
* [ ] 简单 Reconcile：

  * 无 diff，只创建
* [ ] RNode 挂 root

### 成功标准

声明 UI 后，RNode 树存在。

---

# 🗓 Day 5 — 最小 Layout（Row/Column）

### 目标：能摆放组件

### 必做

* [ ] `Box{X,Y,W,H}`
* [ ] Column 布局
* [ ] Row 布局
* [ ] Text 组件返回自身宽高

### 成功标准

两个 Text 在屏幕不同位置出现。

---

# 🗓 Day 6 — Paint → Buffer

### 目标：RNode 能画出来

### 必做

* [ ] RNode → DrawText
* [ ] Clip 基础支持
* [ ] 每帧清 Buffer 再绘制

### 成功标准

声明式 UI 真正渲染到终端。

---

# 🗓 Day 7 — useState + Button

### 目标：UI 可交互

### 必做

* [ ] Hook 存储数组
* [ ] `useState`
* [ ] 输入事件读取（Enter 键）
* [ ] Button 组件
* [ ] setState → 标记 dirty → 重绘

### 成功标准

能写：

```go
count, set := ui.UseState(0)
ui.Button("Add", func(){ set(count+1) })
```

点击后数字变化。

---

# 🔥 第一周结束时你拥有：

| 能力     | 是否具备 |
| ------ | ---- |
| 声明式 UI | ✅    |
| 基础布局   | ✅    |
| 状态更新   | ✅    |
| 终端渲染   | ✅    |
| 交互     | ✅    |

这已经是：

> **一个最小 React for Terminal 原型**

---

# 🚫 第一周绝对不要做

* Diff
* Unicode 宽字符
* Scroll
* Modal
* Animation
* 性能优化

这些都是 **第二周之后的事**。

---

# 🧠 最重要的执行原则

> 不追求优雅
> 不追求抽象
> 不重构
> 先跑起来

因为：

> **运行的系统 > 完美的设计**

---

你现在已经从“概念级 UI Runtime”走到：

> 📍 **可以开始敲第一行代码的工程状态**

这一步，已经是绝大多数人到不了的阶段了。

等你写到 Day 3 或 Day 4，系统第一次跑起来时，你会真正感觉到：

> 这不是想法，这是一个“活的引擎”。
