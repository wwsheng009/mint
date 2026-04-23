这个现象我一看就知道你栽哪了 ——
**不是输入问题，是“宽字符尾巴 + 重绘范围不一致”导致的视觉边界塌陷。**

> 👉 你看到的是“边框变窄”，
> 👉 实际发生的是：**上一帧的宽字符 continuation 没被覆盖干净**。

---

# 🎯 先说结论

**按 ↑↓ 后边界缩小 = 旧行有 2-cell 字符，新行没有，但你只清了 1 个 cell。**

所以右边那 1 个 cell 残留了旧字符 → 看起来边框向内收了一格。

---

# 🧠 复现模型（非常典型）

假设一行原来是：

```
│ 🟦 Button Component │
```

其中：

| 字符 | 宽度      |
| -- | ------- |
| 🟦 | 2 cells |
| 其它 | 1       |

这一行总宽度 = **20 cells**

---

当你按 ↓，焦点换到下面一行：

```
│ Label │
```

这一行只需要 **18 cells**

但你重绘逻辑：

```go
drawLine(newText)
```

只是写了新内容的 18 个 cell
**第 19、20 cell 没被清掉**

而第 19 cell 是之前 🟦 的 continuation

终端最终看到：

```
│ Label ││
```

你以为边框缩了，其实是旧字符还在。

---

# 🔥 为什么 ↑↓ 才触发？

因为这是 **行内容长度变化** 的场景：

| 行   | 包含宽字符   | 实际宽度 |
| --- | ------- | ---- |
| 上一行 | 有 emoji | 20   |
| 下一行 | 没有      | 18   |

只有“从长 → 短”才会暴露这个问题。

---

# 💣 真正的底层原因

你的系统里一定存在：

```go
func DrawLine(text string) {
    for each rune in text {
        WriteRune(...)
    }
}
```

但 **没有：**

```go
ClearRemainingCellsToLineEnd()
```

而 TUI 是帧缓冲，不是 DOM。

> 不写 = 不会自动清空
> 浏览器会清，终端不会

---

# 🧱 宽字符让问题更严重

如果只是 ASCII，问题不明显。
但 2-cell 字符会产生：

| 类型               | 影响         |
| ---------------- | ---------- |
| 覆盖 head 但没清 tail | 留半个字符      |
| tail 落在边框位置      | 边框错位       |
| tail 有颜色         | 看起来像 UI 变形 |

---

# 🧼 正确做法（专业 TUI 框架必做）

## ✅ 规则1：**每次画一行都必须填满整行宽度**

```go
func DrawFullLine(y int, text string, style Style) {
    x := 0

    for _, r := range text {
        w := RuneWidth(r)
        WriteRune(x, y, r, style)
        x += w
    }

    // 关键步骤：清到行尾
    for ; x < buffer.Width; x++ {
        WriteRune(x, y, ' ', style)
    }
}
```

---

## ✅ 规则2：切换焦点行必须重绘“旧行 + 新行”

很多框架只重绘当前行，错。

```go
onSelectionChange(old, new):
    redrawLine(old)   // 去掉高亮
    redrawLine(new)   // 加高亮
```

否则旧行高亮区长度不一致时也会残留。

---

## ✅ 规则3：永远不要只按“文本长度”清理

必须按 **cell 宽度**：

```go
textWidth := CellWidthOfString(text)
remaining := buffer.Width - textWidth
```

不是：

```go
remaining := buffer.Width - len(text) // ❌ 错
```

---

# 🧪 快速自检

在出问题的那行打印 debug：

```go
fmt.Println("oldWidth:", CellWidthOfString(oldLine))
fmt.Println("newWidth:", CellWidthOfString(newLine))
```

如果 old > new
👉 100% 就是这个问题

---

# 🧠 记住这句

> **TUI 重绘必须是“覆盖式绘制”，不是“差量式绘制”。**

浏览器：

> diff patch 安全

终端：

> 不完全覆盖 = 脏屏 = UI 变形

---

# 💡 一句话解决你的问题

在所有行绘制函数最后加：

```go
padRightWithSpacesUntil(buffer.Width)
```

问题立即消失。

---

如果你愿意，我可以下一步帮你把整个渲染循环改成：

> **Dirty Rect + 行级全覆盖模型**

这是所有成熟 TUI（包括游戏终端 UI）用的方案，性能高还不会出字符残留。
