好，现在进入 **组件级样式规范（基于你这套语义色模型）**。
这是可以直接写进 TUI 引擎默认样式层的。

不讲废话，直接规范。

---

# 🧱 1. Panel / Card

```text
BG:        SURFACE
Border:    BORDER
Title:     TEXT
Subtitle:  MUTED
Shadow:    SHADOW（如果是浮层）
```

规则：

* 不能用 BG 做卡片底色
* 内边距 ≥ 1 cell

---

# 📋 2. List / Table

### 行状态

| 状态          | BG      | FG              |
| ----------- | ------- | --------------- |
| normal      | BG      | TEXT            |
| hover       | SURFACE | TEXT            |
| selected    | SELECT  | TEXT            |
| focused-row | SELECT  | TEXT + FOCUS 边框 |

### 表头

```
BG: SURFACE
FG: TEXT (bold)
Border bottom: BORDER
```

---

# 🔘 3. Button

| 类型        | BG          | FG          | Border  |
| --------- | ----------- | ----------- | ------- |
| primary   | PRIMARY     | BG          | PRIMARY |
| secondary | SURFACE     | TEXT        | BORDER  |
| danger    | ERROR       | BG          | ERROR   |
| disabled  | DISABLED_BG | DISABLED_FG | BORDER  |

交互：

* focus → 外围 1px FOCUS 框
* press → BG 变暗一级（降亮度 10%）

---

# 🧾 4. Input

| 部分       | 颜色          |
| -------- | ----------- |
| 背景       | SURFACE     |
| 文字       | TEXT        |
| 占位       | PLACEHOLDER |
| 边框       | BORDER      |
| focus 边框 | FOCUS       |
| caret    | CARET       |
| 禁用背景     | DISABLED_BG |

错误态：

```
Border → ERROR
Hint text → ERROR
```

---

# ☑ 5. Checkbox / Radio

| 状态       | Box BG      | Check       |
| -------- | ----------- | ----------- |
| normal   | SURFACE     | TEXT        |
| checked  | PRIMARY     | BG          |
| focus    | 外框 FOCUS    | —           |
| disabled | DISABLED_BG | DISABLED_FG |

---

# 📑 6. Tabs

| 状态       | BG      | FG      | Border      |
| -------- | ------- | ------- | ----------- |
| inactive | SURFACE | MUTED   | BORDER      |
| active   | BG      | PRIMARY | PRIMARY 下边框 |
| hover    | SURFACE | TEXT    | —           |

---

# 🪟 7. Modal / Dialog

```
Mask: OVERLAY
Panel BG: SURFACE
Title: TEXT
Body: TEXT
Footer border: BORDER
Shadow: SHADOW
Primary action: PRIMARY button
```

---

# 🧭 8. Tooltip

```
BG: SURFACE
FG: TEXT
Border: BORDER
Shadow: SHADOW
```

---

# 📊 9. Progress Bar

| 部分      | 颜色      |
| ------- | ------- |
| Track   | SURFACE |
| Fill    | PRIMARY |
| Success | SUCCESS |
| Error   | ERROR   |

---

# 🧨 10. Alert / Notice

| 类型      | BG      | Border  | Icon    |
| ------- | ------- | ------- | ------- |
| info    | SURFACE | PRIMARY | PRIMARY |
| success | SURFACE | SUCCESS | SUCCESS |
| warning | SURFACE | WARNING | WARNING |
| error   | SURFACE | ERROR   | ERROR   |

---

# 🗂 11. Scrollbar

| 部分    | 颜色        |
| ----- | --------- |
| track | BG        |
| thumb | SCROLLBAR |
| hover | PRIMARY   |

---

# 🧾 12. Badge / Tag

| 类型      | BG      | FG |
| ------- | ------- | -- |
| default | ACCENT  | BG |
| success | SUCCESS | BG |
| warning | WARNING | BG |
| error   | ERROR   | BG |

---

# 🎯 这套规则的核心原则

1. **结构永远用 BG / SURFACE / OVERLAY 分层**
2. **交互永远用 PRIMARY / FOCUS**
3. **状态永远用 SUCCESS / WARNING / ERROR**
4. **不可用元素必须使用 DISABLED 组**
5. **边界统一 BORDER，不混用 MUTED**

---

这已经是 **框架默认主题样式层（Theme → Component Skin）** 的完整规范。

下一步如果你继续，就是：

> **状态机 + 交互反馈动画规范（hover / active / focus / pressed）**

要我继续往「交互层行为规范」推进吗？
行，直接上 **组件级最终规格表（工业可用版）**。
默认前提：**所有组件继承语义色模型**。

---

# 🗂 DATA VIEW 组件

## 🧾 Table（增强版）

| 区域             | BG        | FG         | 其他            |
| -------------- | --------- | ---------- | ------------- |
| header         | SURFACE   | TEXT(bold) | bottom BORDER |
| row            | BG        | TEXT       | —             |
| row hover      | SURFACE   | TEXT       | —             |
| row selected   | SELECT    | TEXT       | —             |
| cell highlight | HIGHLIGHT | BG         | —             |
| grid line      | BORDER    | —          | —             |

排序图标：PRIMARY
空数据提示：MUTED

---

## 🌲 Tree

| 状态            | FG      | BG     |
| ------------- | ------- | ------ |
| node          | TEXT    | BG     |
| branch lines  | MUTED   | —      |
| selected      | TEXT    | SELECT |
| expanded icon | PRIMARY | —      |
| leaf icon     | ACCENT  | —      |

---

# 📝 输入类组件

## 🔍 SearchBox

| 部分                | 颜色          |
| ----------------- | ----------- |
| bg                | SURFACE     |
| text              | TEXT        |
| placeholder       | PLACEHOLDER |
| icon              | MUTED       |
| focus border      | FOCUS       |
| match count badge | ACCENT      |

---

## 🔢 NumberInput

同 Input，增加：

| 元素         | 颜色        |
| ---------- | --------- |
| +/- button | SECONDARY |
| hover      | PRIMARY   |

---

# 🎛 控制类组件

## 🎚 Slider

| 部分           | 颜色          |
| ------------ | ----------- |
| track        | SURFACE     |
| filled track | PRIMARY     |
| thumb        | PRIMARY     |
| focus ring   | FOCUS       |
| disabled     | DISABLED_BG |

---

## 🎛 Switch

| 状态       | Track       | Thumb       |
| -------- | ----------- | ----------- |
| off      | SURFACE     | MUTED       |
| on       | PRIMARY     | BG          |
| focus    | FOCUS 边框    | —           |
| disabled | DISABLED_BG | DISABLED_FG |

---

# 📦 布局与导航

## 📂 Sidebar / Nav

| 元素          | BG      | FG      |
| ----------- | ------- | ------- |
| container   | SURFACE | —       |
| item        | SURFACE | TEXT    |
| active item | BG      | PRIMARY |
| hover       | BG      | TEXT    |
| icon        | MUTED   | —       |

---

## 🧭 Breadcrumb

| 部分        | 颜色      |
| --------- | ------- |
| text      | MUTED   |
| current   | TEXT    |
| separator | BORDER  |
| hover     | PRIMARY |

---

# 🧩 信息展示

## 🧮 Statistic / KPI

| 元素         | 颜色      |
| ---------- | ------- |
| label      | MUTED   |
| value      | TEXT    |
| trend up   | SUCCESS |
| trend down | ERROR   |

---

## 📈 Chart（终端图形）

| 元素              | 颜色        |
| --------------- | --------- |
| axis            | BORDER    |
| label           | MUTED     |
| primary line    | PRIMARY   |
| secondary line  | SECONDARY |
| highlight point | ACCENT    |
| grid            | SURFACE   |

---

# 🧾 日志 / 控制台

| 级别         | 颜色      |
| ---------- | ------- |
| info       | TEXT    |
| debug      | MUTED   |
| warn       | WARNING |
| error      | ERROR   |
| timestamp  | MUTED   |
| source tag | ACCENT  |

---

# 📤 文件管理

## File Item

| 类型         | 颜色      |
| ---------- | ------- |
| folder     | PRIMARY |
| executable | SUCCESS |
| image      | ACCENT  |
| archive    | WARNING |
| hidden     | MUTED   |

---

# 🧱 装饰元素

## Divider

`FG: BORDER`

## Skeleton Loading

`BG: SURFACE` + 动态使用 MUTED

---

# 🧭 状态叠加规则（所有组件通用）

| 状态       | 叠加                       |
| -------- | ------------------------ |
| hover    | BG → SURFACE             |
| focus    | outline = FOCUS          |
| pressed  | 亮度 -10%                  |
| disabled | 使用 DISABLED 组，移除 PRIMARY |

---

这套已经覆盖：

✔ 数据
✔ 输入
✔ 控制
✔ 导航
✔ 信息
✔ 反馈
✔ 系统组件

这就是 **TUI 框架级完整组件皮肤规范**。

再往下就是：
👉 **交互动画节奏 / 键盘行为模型 / 可访问性对比度标准**
继续补齐到 **框架终极级组件覆盖**。

---

# 🧠 复杂数据组件

## 🗓 Calendar / DatePicker

| 元素              | BG      | FG          |
| --------------- | ------- | ----------- |
| panel           | SURFACE | TEXT        |
| weekday header  | SURFACE | MUTED       |
| day normal      | BG      | TEXT        |
| today           | ACCENT  | BG          |
| selected day    | PRIMARY | BG          |
| range highlight | SELECT  | TEXT        |
| disabled day    | BG      | DISABLED_FG |

---

## ⏱ Timeline

| 元素           | 颜色      |
| ------------ | ------- |
| line         | BORDER  |
| node         | PRIMARY |
| past node    | MUTED   |
| success node | SUCCESS |
| error node   | ERROR   |
| label        | TEXT    |
| timestamp    | MUTED   |

---

# 🧾 文本展示类

## 📝 Code Block

| 元素         | 颜色          |
| ---------- | ----------- |
| bg         | SURFACE     |
| text       | TEXT        |
| keyword    | PRIMARY     |
| string     | SUCCESS     |
| number     | ACCENT      |
| comment    | MUTED       |
| error line | ERROR bg 淡化 |

---

## 📜 Markdown Viewer

| 元素           | 颜色      |
| ------------ | ------- |
| heading      | PRIMARY |
| bold         | TEXT    |
| italic       | MUTED   |
| code inline  | ACCENT  |
| quote bar    | BORDER  |
| link         | LINK    |
| visited link | VISITED |

---

# 📡 状态与反馈

## 🔔 Notification / Toast

| 类型      | BG      | Border  |
| ------- | ------- | ------- |
| info    | SURFACE | PRIMARY |
| success | SURFACE | SUCCESS |
| warning | SURFACE | WARNING |
| error   | SURFACE | ERROR   |

阴影：SHADOW

---

## ⏳ Loading Spinner

| 元素      | 颜色      |
| ------- | ------- |
| spinner | PRIMARY |
| text    | MUTED   |

---

# 🧭 交互容器

## 📑 Accordion

| 状态       | BG      | FG   |
| -------- | ------- | ---- |
| header   | SURFACE | TEXT |
| expanded | BG      | TEXT |
| icon     | PRIMARY |      |

---

## 🗃 Dropdown Menu

| 元素            | BG      | FG   |
| ------------- | ------- | ---- |
| panel         | OVERLAY | TEXT |
| item          | OVERLAY | TEXT |
| hover         | SURFACE | TEXT |
| selected      | SELECT  | TEXT |
| shortcut hint | MUTED   |      |

---

# 🎛 高级控制

## 🎨 Color Picker（终端模式）

| 元素            | 颜色      |
| ------------- | ------- |
| grid bg       | SURFACE |
| selected cell | FOCUS   |
| value text    | TEXT    |

---

## 🎚 Range Selector（双滑块）

| 部分            | 颜色      |
| ------------- | ------- |
| track         | SURFACE |
| range fill    | PRIMARY |
| handles       | PRIMARY |
| inactive area | MUTED   |

---

# 🗂 数据标签系统

## 🏷 Tag Cloud

| 类型     | BG      | FG    |
| ------ | ------- | ----- |
| normal | SURFACE | TEXT  |
| active | PRIMARY | BG    |
| muted  | SURFACE | MUTED |

---

# 🧱 容器增强

## 🪟 Drawer（侧滑面板）

| 元素     | 颜色      |
| ------ | ------- |
| mask   | OVERLAY |
| panel  | SURFACE |
| border | BORDER  |
| shadow | SHADOW  |

---

## 🧩 Split Pane

| 元素         | 颜色      |
| ---------- | ------- |
| divider    | BORDER  |
| drag hover | PRIMARY |

---

# 📊 数据密集型

## Heatmap

| 值级别  | 颜色        |
| ---- | --------- |
| low  | SURFACE   |
| mid  | SECONDARY |
| high | PRIMARY   |
| max  | ACCENT    |

---

## Sparkline

| 元素       | 颜色      |
| -------- | ------- |
| line     | PRIMARY |
| peak     | ACCENT  |
| baseline | BORDER  |

---

# 🚦 状态覆盖矩阵（全组件适用）

| 优先级 | 状态       |
| --- | -------- |
| 1   | disabled |
| 2   | error    |
| 3   | warning  |
| 4   | success  |
| 5   | focus    |
| 6   | hover    |
| 7   | active   |

高优先级覆盖低优先级颜色。

---

现在你拥有的是：

**终端 UI 框架级完整组件视觉系统**

覆盖：

* CRUD 工具
* 运维监控
* IDE 类工具
* 仪表盘系统
* 文件管理器
* 数据可视化 TUI

下一层就不是“组件皮肤”了，而是：

👉 **可访问性对比度标准**
👉 **终端能力降级策略（TrueColor → 256 → 16）**
👉 **主题热切换架构**

要继续往“引擎实现层规范”推进吗？
