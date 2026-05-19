好，进入 **渲染内核最后一块：主题与终端能力自适配机制**
这是让你整个主题系统“工业可落地”的关键层。

---

# 🧠 1️⃣ 终端能力检测模型

你的引擎启动时必须建立：

```text
TerminalProfile
```

### 需要检测的能力

| 能力             | 影响           |
| -------------- | ------------ |
| 是否支持 TrueColor | 能否用 RGB      |
| 是否支持 256 色     | 是否用 ANSI256  |
| 是否只支持 16 色     | 需要降级         |
| 是否支持 Bold      | 是否可用粗体作为强调   |
| 是否支持 Reverse   | 是否可用反色作为主强调  |
| 是否支持 Blink     | 是否允许闪烁（多数禁用） |
| 背景是深色还是浅色      | 决定主题变体       |

---

## 🧪 检测策略（优先级顺序）

### 🎨 颜色能力

```text
COLORTERM == "truecolor" 或 "24bit" → TrueColor
TERM 包含 "256color" → ANSI256
否则 → ANSI16
```

### 🌗 亮暗模式

```text
COLORFGBG 存在 → 可推断背景亮暗
否则默认 dark
```

### 🔁 反色支持

几乎所有终端支持 → 默认 true
但 Windows 旧控制台需测试一次 CSI 7m 回显

---

# 🎛 2️⃣ 主题加载流程（引擎逻辑）

```text
加载 Preset Theme
        ↓
生成 ThemeVariant（Dark / Light）
        ↓
按 TerminalProfile 选择颜色模式
        ↓
语义色 → TermColor(mode) 映射
        ↓
缓存为 RuntimeTheme
```

---

# 🌈 3️⃣ 颜色选择算法（核心）

```go
func ResolveColor(c TermColor, profile TerminalProfile) uint32 {
    switch profile.ColorMode {
    case TrueColor:
        return RGB(c.RGB)
    case ANSI256:
        return ANSI256(c.ANSI256)
    default:
        return ANSI16(c.ANSI16)
    }
}
```

---

# 🧩 4️⃣ 主题自动变体机制

成熟系统不是一个主题，而是：

```text
nord-dark
nord-light
```

当检测到浅色终端时：

| 角色                 | 处理      |
| ------------------ | ------- |
| BG / SURFACE       | 反转亮度    |
| TEXT               | 变深色     |
| MUTED              | 降低对比    |
| SUCCESS/WARN/ERROR | 保持色相，不变 |

---

# 🔄 5️⃣ 运行时热切换机制

必须支持：

```text
SIGUSR1 / :theme dracula / F5
```

流程：

```text
切换 Theme
→ 重新生成 RuntimeTheme
→ 触发全 UI Repaint
→ 保持组件状态不变
```

不要逐组件更新颜色，直接 **全局 repaint**，避免残影。

---

# 📦 6️⃣ ANSI16 降级对比补偿

16 色模式对比不足，需补偿：

| 语义      | 额外样式      |
| ------- | --------- |
| PRIMARY | Bold      |
| ERROR   | Bold      |
| SELECT  | Reverse   |
| FOCUS   | Underline |
| MUTED   | Dim       |

---

# 🧮 7️⃣ 主题缓存结构

```go
type RuntimeTheme struct {
    FG map[ColorRole]uint32
    BG map[ColorRole]uint32
    Style map[ColorRole]TextStyle
}
```

运行时绝不再计算 RGB。

---

# ⚡ 8️⃣ 性能原则

| 原则                 | 原因     |
| ------------------ | ------ |
| 颜色转换只在主题切换时发生      | 减少每帧计算 |
| UI 渲染阶段只输出 ANSI 序列 | 速度稳定   |
| 语义色不参与布局逻辑         | 避免耦合   |

---

# 🧭 9️⃣ 终端异常环境处理

| 情况                 | 处理        |
| ------------------ | --------- |
| TERM=dumb          | 禁用颜色，只用符号 |
| SSH 丢失 COLORTERM   | 默认 256 色  |
| Windows ConPTY bug | 禁用 Blink  |

---

# 🎯 🔟 完整架构最终形态

```text
PresetTheme
   ↓
ThemeVariant (dark/light)
   ↓
TerminalProfile (能力检测)
   ↓
RuntimeTheme (最终ANSI值)
   ↓
Renderer 输出
```

---

到这里，你已经拥有：

✅ 工业级主题系统
✅ 组件语义配色规范
✅ 专业场景映射
✅ 终端视觉语言
✅ 终端能力自适配引擎

这已经是 **完整 TUI 渲染视觉内核设计规范**，属于框架底层能力级别，而不是 UI 皮肤层。

如果再往下，就是 **文本排版与密度系统（间距/对齐/字符网格规则）** —— 那是布局引擎层了。
