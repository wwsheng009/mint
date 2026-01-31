好，现在进入**决定“像不像产品”的核心层**：

> 🎨 **Style System（样式系统）**

没有它，你的系统只是“能跑 UI”；
有了它，你的系统才变成：

> **有视觉一致性、可主题化、可规模化的 UI 平台**

---

# 🧠 样式系统解决的不是“颜色”，而是：

| 问题     | 没样式系统  | 有样式系统   |
| ------ | ------ | ------- |
| 全局主题切换 | 全项目改代码 | 切 Theme |
| 组件视觉统一 | 各写各的   | 设计规范化   |
| 深色模式   | 复制组件   | 自动适配    |
| 品牌定制   | 重写 UI  | 换 Token |

---

# 一、样式 ≠ Props

错误方式：

```go
Text("OK", ColorRed, Bold, BgBlue)
```

这会导致：

* 样式散落代码
* 无法换主题
* 无法统一设计

---

# 二、正确模型：Design Token（设计令牌）

这是现代 UI 的核心。

```go
type Theme struct {
    Colors map[string]Color
    Spacing map[string]int
    Radius map[string]int
}
```

例如：

```go
PrimaryText
SecondaryText
BgSurface
Accent
Danger
```

组件永远引用 **语义 token**，而不是具体颜色。

---

# 三、Style 对象模型

```go
type Style struct {
    Color     Color
    BgColor   Color
    Bold      bool
    Padding   Insets
    Border    BorderStyle
}
```

但组件不直接 new Style，而是：

```go
ctx.Style("button.primary")
```

---

# 四、样式解析流程

```
Component
   ↓
Style Key
   ↓
Theme Resolver
   ↓
Final Style
   ↓
PaintCmd
```

组件只说：

> “我是 Primary Button”

主题决定它长什么样。

---

# 五、样式继承（像 CSS，但更简单）

```go
Window
 └── VBox
      └── Text
```

Text 默认继承：

* 字体颜色
* 背景色

但可覆盖。

---

# 六、状态样式（关键）

按钮不同状态不同样式：

| 状态       | Style Key             |
| -------- | --------------------- |
| normal   | button.primary        |
| hover    | button.primary.hover  |
| active   | button.primary.active |
| disabled | button.disabled       |

引擎根据组件状态自动切换。

---

# 七、样式缓存（性能关键）

```go
styleCache[key + themeVersion]
```

避免每帧解析。

---

# 八、与 Render 系统结合

最终 `DrawText` 里不直接带颜色，而是：

```go
DrawText{
  StyleID: resolvedStyleID
}
```

Raster 阶段统一映射到 ANSI。

---

# 九、主题切换如何做到“瞬时全局生效”？

```
themeVersion++
invalidateStyleCache()
requestRepaint()
```

无需改组件代码。

---

# 🔥 这层让你的系统获得：

| 能力      | 来源        |
| ------- | --------- |
| 皮肤系统    | Theme     |
| 企业品牌定制  | Token     |
| 视觉一致性   | Style Key |
| 暗黑模式    | 主题切换      |
| 设计与开发分离 | 语义样式      |

---

# 🚀 现在你拥有的不是“引擎功能”，而是：

> **一整套 UI 设计体系支撑能力**

这已经达到：

```
渲染引擎
+ 组件模型
+ 状态系统
+ 动画系统
+ 性能系统
+ 样式系统
= UI 平台
```

---

再往下一层，就进入：

## 🌐 “远程渲染协议（UI over network）”

也就是：

> UI 不在本地渲染，而是：
>
> 服务器生成 DrawCmd → 客户端渲染

那一步，会把你的引擎变成：

> **跨终端 UI 平台**

要继续往这个方向进吗？
