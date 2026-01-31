
好，现在进入 **终端 UI 性能的“隐藏王炸”**：

> 🎨 **Style Diff 优化（避免 ANSI 频繁切换）**

很多 TUI 框架已经有 Diff，但还是慢，原因就是：

> ❌ **字符 Diff 做了，样式 Diff 没做**

终端里，**切换样式比打印字符还贵**。

---

# 🧠 一句话理解问题

如果你渲染：

```
AAAAAA (红色)
BBBBBB (红色)
CCCCCC (红色)
```

但每个 Cell 都输出：

```
ESC[31m A
ESC[31m A
ESC[31m A
```

你就在疯狂浪费 IO。

---

# 🧱 一、Cell 结构（回顾）

```go
type Cell struct {
    Rune  rune
    Style Style
}
```

Buffer Diff 不能只比较 Rune，还要比 Style。

---

# 🔍 二、输出阶段必须维护“当前终端样式状态”

```go
type TerminalState struct {
    Fg Color
    Bg Color
    Bold bool
    Underline bool
}
```

---

# ⚙️ 三、Style Diff 算法

```go
func ApplyStyleDiff(prev, next Style) {
    if prev.Fg != next.Fg {
        writeAnsi(FgCode(next.Fg))
    }
    if prev.Bg != next.Bg {
        writeAnsi(BgCode(next.Bg))
    }
    if prev.Bold != next.Bold {
        writeAnsi(BoldCode(next.Bold))
    }
}
```

**只改变化的部分。**

---

# 🧩 四、字符输出流程变成

```go
for each changed cell {
    MoveCursor(x,y)

    ApplyStyleDiff(currentTermStyle, cell.Style)

    WriteRune(cell.Rune)

    currentTermStyle = cell.Style
}
```

---

# 🔥 五、进一步优化：行内样式合并（Run-Length Encoding）

一整行如果：

```
AAAAAA (same style)
```

不要逐格输出：

```
cursor(1,1) ESC[31m A
cursor(2,1) A
cursor(3,1) A
```

而是：

```
cursor(1,1) ESC[31m AAAAAA
```

---

## 实现方式

扫描行：

```go
while i < width {
    start := i
    style := cells[i].Style
    run := collect until style changes
    output run
}
```

---

# 🚀 六、样式重置问题

终端有“遗留状态”问题。

最佳策略：

```
Frame 开始 → Reset Style 一次
之后全靠 diff
```

避免状态错乱。

---

# 🧠 七、这优化带来的量级提升

| 情况       | 输出量               |
| -------- | ----------------- |
| 无样式 Diff | 10,000 ANSI codes |
| 有 Diff   | ~200              |
| 再加 RLE   | ~20               |

IO 直接减少两个数量级。

---

# ⚡ 八、动画场景尤为重要

动画通常只改变：

* 位置
* 少量字符

但如果样式每格都写，就会卡。

Style Diff 能让动画丝滑。

---

# 🧩 九、最终终端输出流水线变为

```
Buffer Diff
  ↓
行扫描
  ↓
样式 Diff
  ↓
字符合并
  ↓
ANSI 输出
```

---

# 🧠 这一步意味着什么？

大部分 TUI 框架：

> Diff = 字符级

你的系统现在是：

> **字符 + 样式级 Diff + RLE 合并**

这是“高性能终端渲染引擎”的标志。

---

到这里，你的引擎已经覆盖：

* 架构层
* 交互层
* 状态层
* 布局层
* Layer 层
* 输入层
* Scroll 虚拟化
* 调度并发
* 样式系统
* **终端输出优化**

这已经是完整 UI Runtime。

---

如果再往上，就是最后的“工程化终点”：

> 🧰 DevTools：可视化 RNode 树、Layout 边界、重绘热点分析
