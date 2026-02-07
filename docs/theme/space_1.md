好，这一步是 **TUI 版 Design Token（尺寸系统）**，决定 UI 看起来是“业余拼格子”还是“专业框架”。

终端单位 = **cell（字符格）**
下面是可直接写进框架默认值的规格。

---

# 📏 一、基础尺寸体系（Spacing Scale）

统一使用 **4 的节奏**，但适配终端字符密度：

| Token | Cell |
| ----- | ---- |
| XS    | 0    |
| S     | 1    |
| M     | 2    |
| L     | 3    |
| XL    | 4    |
| XXL   | 6    |

用途：

* padding
* margin
* gap
* 组件最小高度

---

# 🧱 二、容器类组件

## Panel / Card

| 属性               | 值   |
| ---------------- | --- |
| padding          | 1–2 |
| margin           | 1   |
| min height       | 3   |
| title bar height | 1   |
| border width     | 1   |

---

## Modal / Dialog

| 属性                    | 值          |
| --------------------- | ---------- |
| padding               | 2          |
| margin to screen edge | ≥2         |
| min width             | 30         |
| max width             | 70% screen |
| footer height         | 3          |

---

# 📋 三、数据类组件

## Table

| 属性               | 值      |
| ---------------- | ------ |
| row height       | 1      |
| header height    | 1      |
| cell padding     | 1 (左右) |
| column min width | 6      |
| column max width | 40     |
| row gap          | 0      |

---

## List

| 属性           | 值 |
| ------------ | - |
| item height  | 1 |
| padding left | 1 |
| gap          | 0 |

---

# 📝 四、输入类

## Input

| 属性                 | 值  |
| ------------------ | -- |
| height             | 1  |
| padding left/right | 1  |
| min width          | 10 |
| label gap          | 1  |

---

## TextArea

| 属性         | 值 |
| ---------- | - |
| min height | 3 |
| padding    | 1 |
| line gap   | 0 |

---

# 🔘 五、按钮

| 属性                  | 值 |
| ------------------- | - |
| height              | 1 |
| padding left/right  | 2 |
| min width           | 6 |
| gap between buttons | 1 |

按钮文本公式：

```
width = text_len + padding*2
```

---

# ☑ 六、选择类

## Checkbox / Radio

| 属性           | 值         |
| ------------ | --------- |
| box width    | 3 (`[ ]`) |
| gap to label | 1         |

---

## Switch

| 属性     | 值 |
| ------ | - |
| width  | 6 |
| height | 1 |

---

# 🧭 七、导航

## Tabs

| 属性         | 值 |
| ---------- | - |
| tab height | 1 |
| padding    | 2 |
| gap        | 1 |

---

## Sidebar

| 属性          | 值     |
| ----------- | ----- |
| width       | 18–24 |
| item height | 1     |
| padding     | 1     |

---

# 🧩 八、反馈组件

## Alert

| 属性         | 值 |
| ---------- | - |
| padding    | 1 |
| icon gap   | 1 |
| min height | 3 |

---

## Tooltip

| 属性        | 值  |
| --------- | -- |
| padding   | 1  |
| max width | 40 |

---

# 📊 九、进度与图形

## ProgressBar

| 属性        | 值  |
| --------- | -- |
| height    | 1  |
| min width | 10 |

---

## Chart

| 属性         | 值 |
| ---------- | - |
| min height | 6 |
| padding    | 1 |

---

# 🧭 十、布局规则

### 🔹 垂直节奏

```
组件之间最小垂直间距 = 1
区块之间 = 2
```

### 🔹 水平节奏

```
字段 label 与控件 = 1
按钮组内 = 1
表单列间距 = 3
```

---

# 🧠 设计哲学（终端专属）

| 规则               | 原因        |
| ---------------- | --------- |
| 高度几乎都是 1         | 字符 UI 密度高 |
| padding > margin | 终端空间小     |
| 避免高度 2 的控件       | 会打断行对齐    |
| 所有组件对齐到字符网格      | 保持稳定      |

---

# ✅ 这一层完成后，你的系统拥有：

* 色彩系统
* 组件视觉系统
* 状态系统
* 尺寸系统

已经是 **完整 TUI Design System**。

下一层才是：
👉 布局引擎（Flex / Grid 的 TUI 版本）
