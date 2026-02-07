继续给 **高级组件与专业场景配色规范**。

仍然只基于语义色，不依赖主题具体数值。

---

# 🌲 1️⃣ Tree / File Explorer

| 元素           | 颜色规则                  |
| ------------ | --------------------- |
| 普通节点         | `TEXT`                |
| 目录节点         | `PRIMARY`             |
| 展开图标         | `MUTED`               |
| 当前选中节点       | `BG` + `SELECT` 背景    |
| Hover 节点     | `TEXT` + `SURFACE` 背景 |
| 禁用节点         | `MUTED`               |
| 符号线（│├└）     | `BORDER`              |
| Git Modified | `WARNING`             |
| Git Added    | `SUCCESS`             |
| Git Deleted  | `ERROR`               |

---

# 📜 2️⃣ Log Viewer / Console 输出

| 日志级别    | 文字色         | 背景         |
| ------- | ----------- | ---------- |
| TRACE   | `MUTED`     | `BG`       |
| DEBUG   | `SECONDARY` | `BG`       |
| INFO    | `TEXT`      | `BG`       |
| SUCCESS | `SUCCESS`   | `BG`       |
| WARN    | `WARNING`   | `BG`       |
| ERROR   | `ERROR`     | `BG`       |
| FATAL   | `BG`        | `ERROR` 背景 |

**当前行高亮** → 背景 `SURFACE`

---

# 📊 3️⃣ Chart / Graph（关键：可区分性）

图表线条必须固定顺序映射，不跟主题随机：

| 数据序号     | 颜色角色        |
| -------- | ----------- |
| Series 1 | `PRIMARY`   |
| Series 2 | `ACCENT`    |
| Series 3 | `SECONDARY` |
| Series 4 | `SUCCESS`   |
| Series 5 | `WARNING`   |
| Series 6 | `ERROR`     |
| 网格线      | `BORDER`    |
| 坐标文字     | `MUTED`     |
| 标题       | `TEXT`      |

---

# 🧭 4️⃣ Menu / Command Palette

| 元素    | 颜色              |
| ----- | --------------- |
| 背景    | `SURFACE`       |
| 普通项   | `TEXT`          |
| 选中项   | `BG` + `SELECT` |
| 快捷键提示 | `MUTED`         |
| 分组标题  | `PRIMARY`       |
| 描述文本  | `MUTED`         |

---

# 🪟 5️⃣ Modal / Dialog

| 元素     | 颜色                     |
| ------ | ---------------------- |
| 背景     | `SURFACE`              |
| 遮罩层    | `BG` (可降低亮度)           |
| 标题     | `TEXT`                 |
| 内容     | `TEXT`                 |
| 描述     | `MUTED`                |
| 主要按钮   | `PRIMARY` 背景 + `BG` 文字 |
| 取消按钮   | `SURFACE` 背景 + `TEXT`  |
| 危险操作按钮 | `ERROR` 背景 + `BG`      |

---

# 🧩 6️⃣ Scrollbar

| 部位       | 颜色        |
| -------- | --------- |
| 轨道       | `SURFACE` |
| 滑块       | `MUTED`   |
| Hover 滑块 | `PRIMARY` |
| 激活拖动     | `FOCUS`   |

---

# 🧠 7️⃣ Selection & Cursor 规则

| 场景     | 用色              |
| ------ | --------------- |
| 文本选中   | `BG` + `SELECT` |
| 光标所在字符 | `BG` + `FOCUS`  |
| 多选区域边框 | `PRIMARY`       |
| 拖拽选择框  | `FOCUS`         |

---

# 📁 8️⃣ File Type（可选扩展）

| 类型              | 颜色        |
| --------------- | --------- |
| 可执行文件           | `SUCCESS` |
| 配置文件            | `PRIMARY` |
| 日志文件            | `MUTED`   |
| 图片/资源           | `ACCENT`  |
| 压缩包             | `WARNING` |
| 危险文件（.env/.key） | `ERROR`   |

---

# 🚦 9️⃣ 状态栏 / Status Bar

| 状态   | 背景        | 文字     |
| ---- | --------- | ------ |
| 普通   | `SURFACE` | `TEXT` |
| 活动状态 | `PRIMARY` | `BG`   |
| 警告状态 | `WARNING` | `BG`   |
| 错误状态 | `ERROR`   | `BG`   |
| 成功状态 | `SUCCESS` | `BG`   |

---

# 🔬 🔟 Heat / Utilization 映射（监控类）

| 使用率     | 颜色        |
| ------- | --------- |
| 0–50%   | `SUCCESS` |
| 50–70%  | `PRIMARY` |
| 70–85%  | `WARNING` |
| 85–100% | `ERROR`   |

---

# 🧭 全局约束（防止 UI 失控）

| 规则                            | 原因     |
| ----------------------------- | ------ |
| 同屏强调色 ≤ 2                     | 避免噪声   |
| SUCCESS/WARNING/ERROR 不参与普通装饰 | 保持语义稳定 |
| 组件背景最多两层：BG / SURFACE         | 防止花屏   |
| SELECT 不与 PRIMARY 同时大面积使用     | 焦点冲突   |
| 图表颜色顺序固定                      | 记忆成本低  |

---

这套已经覆盖：

* 文件管理器
* DevTool
* 监控系统
* 日志系统
* 命令面板
* 数据可视化

属于 **完整 TUI 设计系统配色层规范**。

如果你下一步是做 **交互动效（闪烁/反色/光标节奏）规范** 或 **16色终端降级策略表**，那就是最后一层“终端级视觉语言”了。
