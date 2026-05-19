好眼力，这问题非常关键 👍

前面的 demo 偏重“系统能力验证”，**样式系统没展开**，但如果没有完整样式层，你的引擎会变成：

> “能跑，但不像 UI 框架”

现在我们把 **Styling System（TUI 版 CSS）**补齐。

---

# 🧠 一句话定位

你的样式系统 =

```
CSS Box Model + 终端颜色/属性系统 + 继承规则
```

但要适配字符网格世界。

---

# 🧱 一、Style 结构设计（核心）

```go
type Style struct {
    Fg        Color
    Bg        Color
    Bold      bool
    Italic    bool
    Underline bool

    Padding   Insets
    Margin    Insets
    Border    BorderStyle
    BorderColor Color

    Width, Height Dimension
    FlexGrow int
    Align    AlignType
}
```

这是组件的“视觉声明层”。

---

# 🎨 二、SDK 使用方式（声明式）

```go
ui.Box().
    Style(
        ui.Style().
            Bg(ui.ColorBlue).
            Fg(ui.ColorWhite).
            Padding(1).
            Border(ui.BorderRounded).
    )
```

---

# ✨ 三、样式继承规则（像 CSS）

终端 UI 也需要继承：

| 属性      | 是否继承 |
| ------- | ---- |
| Fg      | ✅    |
| Bg      | ❌    |
| Bold    | ✅    |
| Border  | ❌    |
| Padding | ❌    |

实现：

```go
func ResolveStyle(parent, self Style) Style
```

---

# 📐 四、Box Model 在终端的实现

```
Margin
  Border
    Padding
      Content
```

最终 LayoutBox 计算：

```go
contentW = width - padding - border
```

---

# 🧩 五、样式如何进入 Paint？

每个 Cell：

```go
Cell{
    Rune: 'A',
    Style: cellStyle,
}
```

最终转 ANSI：

```
\033[fg;bg;boldm
```

---

# 🧠 六、主题系统（Theme）

可以全局注入：

```go
ui.SetTheme(ui.Theme{
    Primary: ColorBlue,
    Danger:  ColorRed,
})
```

组件：

```go
ui.Button("Delete").Variant(ui.Danger)
```

---

# 🔥 七、把样式加回 Demo

### Header

```go
ui.Box().
    Height(3).
    Style(
        ui.Style().
            Bg(ui.ColorBlue).
            Fg(ui.ColorWhite).
            Padding(1),
    )
```

---

### Sidebar

```go
ui.Box().
    Width(20).
    Style(
        ui.Style().
            Border(ui.BorderRounded).
            Padding(1).
            Bg(ui.ColorGrayDark),
    )
```

---

### Input

```go
ui.Input().
    Style(
        ui.Style().
            Border(ui.BorderRounded).
            Fg(ui.ColorGreen),
    )
```

---

### Modal

```go
ui.Box().
    Style(
        ui.Style().
            Bg(ui.ColorBlack).
            Border(ui.BorderDouble).
            BorderColor(ui.ColorYellow).
            Padding(1),
    )
```

---

# 🎯 样式系统解决的核心问题

| 能力        | 作用                 |
| --------- | ------------------ |
| 继承        | 减少重复声明             |
| 主题        | UI 统一              |
| Box Model | 布局可预测              |
| ANSI 映射   | 正确终端显示             |
| 状态样式      | Hover/Focus/Active |

---

# 🧠 Hover / Focus 样式

组件内部：

```go
if node.Focused {
    style.BorderColor = ui.ColorCyan
}
```

---

# 🚀 最终效果

有了样式系统，你的引擎从：

> “功能型 TUI”

进化成：

> **“可设计的 UI 框架”**

---

如果你愿意，下一步可以讲：

> 🎨 **Style Diff 优化（避免重复 ANSI 切换）**

这是终端性能的关键技巧。
