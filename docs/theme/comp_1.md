好，直接给 **组件级配色规则表** —— 基于你上面那套语义色（BG / SURFACE / PRIMARY…），这是**TUI 引擎级 UI 规范**，不是审美建议。

---

# 🧱 基础规则（全组件适用）

| 元素    | 用色规则      |
| ----- | --------- |
| 页面背景  | `BG`      |
| 面板/卡片 | `SURFACE` |
| 主文字   | `TEXT`    |
| 次级说明  | `MUTED`   |
| 分割线   | `BORDER`  |
| 焦点框   | `FOCUS`   |
| 选中底色  | `SELECT`  |

---

# 🧩 1️⃣ Button 按钮

| 状态         | 前景色     | 背景色       | 边框        |
| ---------- | ------- | --------- | --------- |
| 默认         | `TEXT`  | `SURFACE` | `BORDER`  |
| Primary 按钮 | `BG`    | `PRIMARY` | `PRIMARY` |
| Hover      | `BG`    | `ACCENT`  | `ACCENT`  |
| Focus      | `BG`    | `PRIMARY` | `FOCUS`   |
| Disabled   | `MUTED` | `SURFACE` | `BORDER`  |

---

# 📋 2️⃣ List / Table 行

| 状态      | 前景     | 背景                     |
| ------- | ------ | ---------------------- |
| 普通行     | `TEXT` | `BG`                   |
| 斑马纹     | `TEXT` | `SURFACE`              |
| Hover 行 | `TEXT` | `SURFACE`              |
| 选中行     | `BG`   | `SELECT`               |
| 焦点行     | `TEXT` | `SURFACE` + `FOCUS` 边框 |

---

# 🧾 3️⃣ Input 输入框

| 部位          | 颜色        |
| ----------- | --------- |
| 背景          | `SURFACE` |
| 文字          | `TEXT`    |
| Placeholder | `MUTED`   |
| 边框默认        | `BORDER`  |
| Focus 边框    | `FOCUS`   |
| 错误态边框       | `ERROR`   |
| 禁用文字        | `MUTED`   |

---

# 📦 4️⃣ Panel / Card

| 元素     | 颜色        |
| ------ | --------- |
| 背景     | `SURFACE` |
| 标题文字   | `TEXT`    |
| 描述文字   | `MUTED`   |
| 边框     | `BORDER`  |
| 激活面板边框 | `FOCUS`   |

---

# 📑 5️⃣ Tabs 标签页

| 状态    | 文字      | 背景        | 下划线/标识    |
| ----- | ------- | --------- | --------- |
| 未选中   | `MUTED` | `BG`      | `BORDER`  |
| Hover | `TEXT`  | `SURFACE` | `PRIMARY` |
| 选中    | `TEXT`  | `SURFACE` | `PRIMARY` |
| 焦点    | `TEXT`  | `SURFACE` | `FOCUS`   |

---

# 🧭 6️⃣ Status / Badge

| 类型      | 背景        | 文字     |
| ------- | --------- | ------ |
| Success | `SUCCESS` | `BG`   |
| Warning | `WARNING` | `BG`   |
| Error   | `ERROR`   | `BG`   |
| Info    | `PRIMARY` | `BG`   |
| Neutral | `SURFACE` | `TEXT` |

---

# 🧮 7️⃣ Progress / Gauge

| 部位   | 颜色        |
| ---- | --------- |
| 轨道   | `SURFACE` |
| 进度条  | `PRIMARY` |
| 成功完成 | `SUCCESS` |
| 警告区段 | `WARNING` |
| 危险区段 | `ERROR`   |
| 文字   | `TEXT`    |

---

# 🔔 8️⃣ Toast / Alert

| 类型      | 背景        | 标题   | 内容   |
| ------- | --------- | ---- | ---- |
| Info    | `PRIMARY` | `BG` | `BG` |
| Success | `SUCCESS` | `BG` | `BG` |
| Warning | `WARNING` | `BG` | `BG` |
| Error   | `ERROR`   | `BG` | `BG` |

---

# 🎯 9️⃣ Border 规则（重要）

| 场景   | 用色        |
| ---- | --------- |
| 普通边框 | `BORDER`  |
| 焦点边框 | `FOCUS`   |
| 错误框  | `ERROR`   |
| 选中框  | `PRIMARY` |

---

# 🧠 关键原则（工业级约束）

1. **不要超过 1 个强调色同时出现**
   → 页面最多一个 `PRIMARY/ACCENT` 焦点源

2. **错误色只用于错误**
   不允许当装饰

3. **背景永远不用纯黑**
   用主题 BG（护眼）

4. **禁用态 = 只降对比，不换色相**

5. **选中态优先“反色”逻辑**
   `FG=BG, BG=SELECT`

---

# ✅ 结果

这套规则保证：

* 任何主题切换 UI 不崩
* 16 色终端仍可辨识
* SSH / 本地 / IDE 终端一致

这已经是 **TUI 组件配色工业规范层**，可以直接写进你框架的 design system 文档。

如果你下一步要做 **图表 / 树形结构 / 日志级别配色规范**，我可以继续给那一层规则。
