好，这一步开始，你从“引擎作者”正式进入：

> **平台设计者模式**

前面你搭好了“操作系统内核”，
现在要定义的是：

> **别人如何给你的系统写“应用程序”**

这就是——

# 🧩 组件标准规范（Component Contract Spec）

如果没有这一层，你的引擎再强，也只能你自己用。

---

# 🧠 组件规范的本质

组件不是函数，不是 struct，而是：

> **一种受引擎调度、参与渲染管线的“受控实体”**

它必须遵守规则，否则会破坏：

* Diff 稳定性
* 布局系统
* 动画系统
* 性能模型

---

# 一、组件生命周期契约

```go
type Component interface {
    Init(ctx *Context)                // 创建
    Render() VNode                    // 描述 UI
    Measure(c Constraints) Size       // 布局阶段
    Paint(ctx *PaintCtx)              // 绘制阶段
    Destroy()                         // 卸载
}
```

### 为什么不能只要 Render？

因为这是：

| 阶段      | 职责    |
| ------- | ----- |
| Render  | UI 结构 |
| Measure | 空间需求  |
| Paint   | 绘制内容  |

把三件事混一起 = 性能灾难。

---

# 二、组件必须遵守的“铁律”

### ❗ 1. Render 必须纯函数

```
相同 props + state → 相同 VNode
```

否则 Diff 失效。

---

### ❗ 2. Measure 不能产生副作用

不能：

* 改 state
* 发事件
* 调网络

否则布局会死循环。

---

### ❗ 3. Paint 不能访问业务状态

Paint 只能依赖：

```
LayoutBox + props + animation value
```

否则动画会触发 Diff。

---

# 三、状态规范

```go
count := useState(0)
```

状态更新必须：

* 异步
* 批处理
* 不立即执行 Render

否则 Scheduler 无法优化。

---

# 四、布局规范

组件不能直接设置坐标：

❌

```go
SetPosition(10, 5)
```

✅

```go
return VBox(
    AlignCenter(Text("OK")),
)
```

布局只能通过容器完成。

---

# 五、绘制规范（Render Cmd）

组件不能直接操作 Buffer：

❌

```go
buf.SetCell(...)
```

✅

```go
ctx.DrawText(...)
```

否则：

* 无法裁剪
* 无法动画
* 无法脏区

---

# 六、动画规范

组件不能用定时器自己动：

❌

```go
go loop { updateState() }
```

✅

```go
Animate(value, to=100, duration=200ms)
```

动画必须交给 Animation Engine。

---

# 七、事件规范

输入必须走事件系统：

```
Terminal → Event Loop → RNode → Component
```

不能直接读取 stdin。

---

# 八、资源管理规范

组件如果持有：

* goroutine
* 订阅
* 文件句柄

必须在 Destroy() 释放。

否则长时间运行会内存泄漏。

---

# 九、为什么这一层决定生态？

因为这定义了：

> **第三方开发者“能做什么、不能做什么”**

这等同于：

| 平台      | 对应规范            |
| ------- | --------------- |
| React   | Hooks Rules     |
| Flutter | RenderObject 协议 |
| Android | View 生命周期       |

没有规范 → 项目变“脚本集合”。

---

# 🔥 最终你拥有的不是框架，而是：

```
UI Runtime
+ Rendering Engine
+ Component Contract
= UI Platform
```

---

# 🚀 下一步如果继续进化，就不是“引擎内部”了，而是：

## 🌍 平台生态层

* UI DSL 设计（类似 JSX）
* 组件市场机制
* 主题与样式系统
* 远程渲染协议（UI over SSH / WebSocket）
* DevTools 协议

这一步，才是从“技术系统”变成“产品平台”。

---

如果你想，我可以下一步带你设计：

### 🎨 样式系统（Style System）

这将决定你的 UI 是否能做到：

* 主题切换
* 皮肤系统
* 深色模式
* 组件视觉统一

这一步，会让你的引擎“看起来像产品”，而不是技术演示。
