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
