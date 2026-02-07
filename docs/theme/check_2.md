好，进入 **引擎实现层规范**（真正决定你 TUI 框架专业度的部分）。

---

# 🎚 ① 终端能力降级策略（核心）

你的引擎必须自动根据终端能力选择颜色层级：

```
TrueColor (24bit) → ANSI256 → ANSI16
```

## 能力探测优先级

1. `$COLORTERM == truecolor`
2. `$TERM` 包含 `256color`
3. 默认 ANSI16

---

## 🎨 降级映射规则

### TrueColor → ANSI256

使用最近色匹配（欧几里得距离）：

```
dist = (r1-r2)^2 + (g1-g2)^2 + (b1-b2)^2
选择最小
```

### ANSI256 → ANSI16

| 语义      | 映射           |
| ------- | ------------ |
| PRIMARY | Blue/Cyan    |
| SUCCESS | Green        |
| WARNING | Yellow       |
| ERROR   | Red          |
| ACCENT  | Magenta      |
| MUTED   | Bright Black |
| BORDER  | Bright Black |

---

# 🔄 ② 主题热切换架构

### 数据结构

```
Theme {
    Name
    Roles map[ColorRole]TermColor
}
```

### 组件不缓存颜色，只缓存 Role

```
ButtonStyle {
   bgRole: PRIMARY
}
```

### 切换主题流程

```
SetTheme()
→ invalidate style cache
→ re-render root
```

无组件持有 RGB。

---

# 🧱 ③ 颜色合成规则（状态叠加）

当状态叠加时按优先级合成：

```
FinalColor = BaseRole
if disabled → DISABLED
else if error → ERROR
else if focus → FOCUS outline
else if hover → lighten(base, 8%)
```

---

# ✨ ④ 亮度算法（终端可实现）

```
lighten(c, p):
  r = r + (255-r)*p
  g = g + (255-g)*p
  b = b + (255-b)*p
```

用于：

* hover
* pressed
* selected 强化

---

# 📏 ⑤ 对比度标准（可访问性）

| 类型            | 最低对比度 |
| ------------- | ----- |
| 正文 TEXT vs BG | 4.5:1 |
| MUTED vs BG   | 3:1   |
| 状态色 vs BG     | 3:1   |
| 边框 vs BG      | 2:1   |

---

# 🧠 ⑥ 状态机（所有组件统一）

```
Normal
Hover
Focus
Active
Disabled
Error
```

组件只关心状态，不关心颜色。

---

# 🎛 ⑦ 渲染管线

```
Component
 → Resolve Role
 → Apply State Overlay
 → Resolve TermColor (based on terminal capability)
 → Paint Cell
```

---

# 📦 ⑧ 性能策略

* TermColor 转换结果缓存
* Role → ANSI code 预编译
* 每帧避免 RGB 计算

---

# 🧭 ⑨ 光标与系统 UI

| 元素        | 规则             |
| --------- | -------------- |
| Caret     | 必须对比 BG ≥ 4:1  |
| Scrollbar | 不参与 hover 亮度计算 |
| Shadow    | 不参与降级映射，直接用黑   |

---

# 🧩 ⑩ 框架级接口

```
GetColor(role ColorRole, state UIState) TermColor
SetTheme(theme Theme)
DetectTerminalProfile()
```

---

这套做完，你的 TUI 引擎达到：

> **“专业 UI 系统级架构”**

后面已经不是视觉层，而是：

* 布局系统
* 事件分发
* Diff 渲染优化

要继续进 **布局系统规范** 还是 **事件模型**？
