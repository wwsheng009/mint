# 5 套主题全部补齐到完整语义模型**。


语义集合：

```
BG SURFACE OVERLAY TEXT MUTED PLACEHOLDER
PRIMARY SECONDARY ACCENT
SUCCESS WARNING ERROR
LINK VISITED
BORDER FOCUS SELECT HIGHLIGHT
DISABLED_BG DISABLED_FG
SCROLLBAR SHADOW CARET
```

好，这部分就是**主题语义层设计文档**，不是配色，而是“每个颜色角色在 UI 系统里的职责边界”。
直接给可落地定义。

---

# 🎨 结构层（Layer System）

### **BG — 全局背景层**

> 应用最底层背景色

用于：

* 整个终端画布底色
* 全屏容器背景

不能用于：

* 卡片、面板内部

---

### **SURFACE — 一级表面层**

> 组件承载层（比 BG 浅一层）

用于：

* Panel / Card / Section
* 列表容器
* 表格区域

作用：建立“模块块状结构”

---

### **OVERLAY — 悬浮层**

> 比 SURFACE 更高层的 UI

用于：

* Modal
* Dropdown
* Tooltip
* Popover

特征：必须与 SURFACE 有明显区分

---

# 📝 文本层（Typography System）

### **TEXT — 主文本**

> 90% 正文内容

用于：

* 正文
* 标题
* 按钮文字

---

### **MUTED — 弱化文本**

> 辅助信息

用于：

* 描述文字
* 时间戳
* 注释

---

### **PLACEHOLDER — 占位文本**

> 临时提示，不属于内容

用于：

* Input placeholder
* 空数据提示中的弱文本

区别：比 MUTED 更“虚”

---

# 🎯 品牌与交互层（Brand & Action）

### **PRIMARY — 主操作色**

> 最重要交互

用于：

* 主按钮
* 当前页 Tab
* 选中主状态

---

### **SECONDARY — 次操作**

> 辅助操作

用于：

* 次按钮
* 次级标签
* 非关键交互

---

### **ACCENT — 点缀色**

> 视觉强调，但不是操作核心

用于：

* Badge
* 数量标记
* 特殊标签

---

# 🚦 状态层（State System）

### **SUCCESS**

✔ 成功状态、通过、在线

---

### **WARNING**

⚠ 警告、注意、待处理

---

### **ERROR**

✖ 失败、错误、危险

---

# 🔗 内容关系层

### **LINK**

可点击文本

---

### **VISITED**

访问过的链接

---

# 📦 结构边界层

### **BORDER**

组件分割线

---

### **FOCUS**

当前获得键盘焦点的元素轮廓

---

### **SELECT**

被选中条目背景（列表/表格）

---

### **HIGHLIGHT**

文本高亮（搜索命中）

---

# 🚫 禁用状态层

### **DISABLED_BG**

禁用控件背景

### **DISABLED_FG**

禁用控件文字

原则：必须显得“不可交互”，但仍可读

---

# 🧭 系统 UI 层

### **SCROLLBAR**

滚动条滑块

---

### **SHADOW**

模拟层级阴影（弹窗底部、浮层边缘）

---

### **CARET**

文本输入光标

---

# 🧠 设计逻辑总结

| 类型  | 角色数量 | 目的     |
| --- | ---- | ------ |
| 结构层 | 3    | 构建空间层次 |
| 文本层 | 3    | 信息层级   |
| 品牌层 | 3    | 操作优先级  |
| 状态层 | 3    | 系统反馈   |
| 关系层 | 2    | 内容连接   |
| 边界层 | 4    | 交互定位   |
| 系统层 | 3    | 系统控件支持 |

---

这套语义不是“颜色表”，而是：

> **TUI 设计系统的色彩职责划分模型**

你后面加主题，只是“换皮”，
语义层不用再改。

---

# 🌌 NORD — FULL

```yaml
name: nord
colors:

  BG: { ansi16: 0, ansi256: 235, rgb: [46,52,64] }
  SURFACE: { ansi16: 8, ansi256: 237, rgb: [59,66,82] }
  OVERLAY: { ansi16: 0, ansi256: 234, rgb: [40,44,52] }

  TEXT: { ansi16: 7, ansi256: 253, rgb: [236,239,244] }
  MUTED: { ansi16: 8, ansi256: 244, rgb: [97,110,136] }
  PLACEHOLDER: { ansi16: 8, ansi256: 244, rgb: [97,110,136] }

  PRIMARY: { ansi16: 6, ansi256: 110, rgb: [136,192,208] }
  SECONDARY: { ansi16: 4, ansi256: 109, rgb: [129,161,193] }
  ACCENT: { ansi16: 14, ansi256: 117, rgb: [143,188,187] }

  SUCCESS: { ansi16: 2, ansi256: 108, rgb: [163,190,140] }
  WARNING: { ansi16: 3, ansi256: 180, rgb: [235,203,139] }
  ERROR: { ansi16: 1, ansi256: 131, rgb: [191,97,106] }

  LINK: { ansi16: 6, ansi256: 110, rgb: [136,192,208] }
  VISITED: { ansi16: 5, ansi256: 139, rgb: [180,142,173] }

  BORDER: { ansi16: 8, ansi256: 240, rgb: [76,86,106] }
  FOCUS: { ansi16: 6, ansi256: 110, rgb: [136,192,208] }
  SELECT: { ansi16: 4, ansi256: 109, rgb: [129,161,193] }
  HIGHLIGHT: { ansi16: 3, ansi256: 180, rgb: [235,203,139] }

  DISABLED_BG: { ansi16: 8, ansi256: 237, rgb: [59,66,82] }
  DISABLED_FG: { ansi16: 8, ansi256: 244, rgb: [97,110,136] }

  SCROLLBAR: { ansi16: 8, ansi256: 240, rgb: [76,86,106] }
  SHADOW: { ansi16: 0, ansi256: 232, rgb: [30,32,40] }
  CARET: { ansi16: 7, ansi256: 253, rgb: [236,239,244] }
```

---

# 🧛 DRACULA — FULL

```yaml
name: dracula
colors:

  BG: { ansi16: 0, ansi256: 236, rgb: [40,42,54] }
  SURFACE: { ansi16: 8, ansi256: 238, rgb: [52,55,70] }
  OVERLAY: { ansi16: 0, ansi256: 235, rgb: [36,39,51] }

  TEXT: { ansi16: 7, ansi256: 255, rgb: [248,248,242] }
  MUTED: { ansi16: 8, ansi256: 103, rgb: [98,114,164] }
  PLACEHOLDER: { ansi16: 8, ansi256: 103, rgb: [98,114,164] }

  PRIMARY: { ansi16: 5, ansi256: 141, rgb: [189,147,249] }
  SECONDARY: { ansi16: 4, ansi256: 111, rgb: [139,233,253] }
  ACCENT: { ansi16: 13, ansi256: 212, rgb: [255,121,198] }

  SUCCESS: { ansi16: 2, ansi256: 84, rgb: [80,250,123] }
  WARNING: { ansi16: 3, ansi256: 228, rgb: [241,250,140] }
  ERROR: { ansi16: 1, ansi256: 203, rgb: [255,85,85] }

  LINK: { ansi16: 4, ansi256: 111, rgb: [139,233,253] }
  VISITED: { ansi16: 5, ansi256: 141, rgb: [189,147,249] }

  BORDER: { ansi16: 8, ansi256: 60, rgb: [68,71,90] }
  FOCUS: { ansi16: 5, ansi256: 141, rgb: [189,147,249] }
  SELECT: { ansi16: 13, ansi256: 212, rgb: [255,121,198] }
  HIGHLIGHT: { ansi16: 3, ansi256: 228, rgb: [241,250,140] }

  DISABLED_BG: { ansi16: 8, ansi256: 238, rgb: [52,55,70] }
  DISABLED_FG: { ansi16: 8, ansi256: 103, rgb: [98,114,164] }

  SCROLLBAR: { ansi16: 8, ansi256: 60, rgb: [68,71,90] }
  SHADOW: { ansi16: 0, ansi256: 234, rgb: [28,29,38] }
  CARET: { ansi16: 7, ansi256: 255, rgb: [248,248,242] }
```

---

# 🐻 GRUVBOX DARK — FULL

```yaml
name: gruvbox-dark
colors:

  BG: { ansi16: 0, ansi256: 235, rgb: [40,40,40] }
  SURFACE: { ansi16: 8, ansi256: 237, rgb: [60,56,54] }
  OVERLAY: { ansi16: 0, ansi256: 234, rgb: [32,30,29] }

  TEXT: { ansi16: 7, ansi256: 223, rgb: [235,219,178] }
  MUTED: { ansi16: 8, ansi256: 246, rgb: [146,131,116] }
  PLACEHOLDER: { ansi16: 8, ansi256: 246, rgb: [146,131,116] }

  PRIMARY: { ansi16: 4, ansi256: 109, rgb: [131,165,152] }
  SECONDARY: { ansi16: 6, ansi256: 108, rgb: [142,192,124] }
  ACCENT: { ansi16: 5, ansi256: 175, rgb: [211,134,155] }

  SUCCESS: { ansi16: 2, ansi256: 142, rgb: [184,187,38] }
  WARNING: { ansi16: 3, ansi256: 214, rgb: [250,189,47] }
  ERROR: { ansi16: 1, ansi256: 167, rgb: [251,73,52] }

  LINK: { ansi16: 4, ansi256: 109, rgb: [131,165,152] }
  VISITED: { ansi16: 5, ansi256: 175, rgb: [211,134,155] }

  BORDER: { ansi16: 8, ansi256: 239, rgb: [80,73,69] }
  FOCUS: { ansi16: 4, ansi256: 109, rgb: [131,165,152] }
  SELECT: { ansi16: 5, ansi256: 175, rgb: [211,134,155] }
  HIGHLIGHT: { ansi16: 3, ansi256: 214, rgb: [250,189,47] }

  DISABLED_BG: { ansi16: 8, ansi256: 237, rgb: [60,56,54] }
  DISABLED_FG: { ansi16: 8, ansi256: 246, rgb: [146,131,116] }

  SCROLLBAR: { ansi16: 8, ansi256: 239, rgb: [80,73,69] }
  SHADOW: { ansi16: 0, ansi256: 232, rgb: [28,27,26] }
  CARET: { ansi16: 7, ansi256: 223, rgb: [235,219,178] }
```

---

# 🍵 CATPPUCCIN MOCHA — FULL

```yaml
name: catppuccin-mocha
colors:

  BG: { ansi16: 0, ansi256: 235, rgb: [30,30,46] }
  SURFACE: { ansi16: 8, ansi256: 238, rgb: [49,50,68] }
  OVERLAY: { ansi16: 0, ansi256: 234, rgb: [24,24,37] }

  TEXT: { ansi16: 7, ansi256: 252, rgb: [205,214,244] }
  MUTED: { ansi16: 8, ansi256: 244, rgb: [108,112,134] }
  PLACEHOLDER: { ansi16: 8, ansi256: 244, rgb: [108,112,134] }

  PRIMARY: { ansi16: 4, ansi256: 111, rgb: [137,180,250] }
  SECONDARY: { ansi16: 5, ansi256: 176, rgb: [203,166,247] }
  ACCENT: { ansi16: 13, ansi256: 218, rgb: [245,194,231] }

  SUCCESS: { ansi16: 2, ansi256: 114, rgb: [166,227,161] }
  WARNING: { ansi16: 3, ansi256: 223, rgb: [249,226,175] }
  ERROR: { ansi16: 1, ansi256: 210, rgb: [243,139,168] }

  LINK: { ansi16: 4, ansi256: 111, rgb: [137,180,250] }
  VISITED: { ansi16: 5, ansi256: 176, rgb: [203,166,247] }

  BORDER: { ansi16: 8, ansi256: 239, rgb: [88,91,112] }
  FOCUS: { ansi16: 4, ansi256: 111, rgb: [137,180,250] }
  SELECT: { ansi16: 13, ansi256: 218, rgb: [245,194,231] }
  HIGHLIGHT: { ansi16: 3, ansi256: 223, rgb: [249,226,175] }

  DISABLED_BG: { ansi16: 8, ansi256: 238, rgb: [49,50,68] }
  DISABLED_FG: { ansi16: 8, ansi256: 244, rgb: [108,112,134] }

  SCROLLBAR: { ansi16: 8, ansi256: 239, rgb: [88,91,112] }
  SHADOW: { ansi16: 0, ansi256: 232, rgb: [22,22,33] }
  CARET: { ansi16: 7, ansi256: 252, rgb: [205,214,244] }
```

---

# 🌞 SOLARIZED DARK — FULL

```yaml
name: solarized-dark
colors:

  BG: { ansi16: 0, ansi256: 234, rgb: [0,43,54] }
  SURFACE: { ansi16: 8, ansi256: 237, rgb: [7,54,66] }
  OVERLAY: { ansi16: 0, ansi256: 233, rgb: [0,36,46] }

  TEXT: { ansi16: 7, ansi256: 254, rgb: [238,232,213] }
  MUTED: { ansi16: 8, ansi256: 244, rgb: [131,148,150] }
  PLACEHOLDER: { ansi16: 8, ansi256: 244, rgb: [131,148,150] }

  PRIMARY: { ansi16: 4, ansi256: 33, rgb: [38,139,210] }
  SECONDARY: { ansi16: 6, ansi256: 37, rgb: [42,161,152] }
  ACCENT: { ansi16: 5, ansi256: 136, rgb: [108,113,196] }

  SUCCESS: { ansi16: 2, ansi256: 64, rgb: [133,153,0] }
  WARNING: { ansi16: 3, ansi256: 136, rgb: [181,137,0] }
  ERROR: { ansi16: 1, ansi256: 160, rgb: [220,50,47] }

  LINK: { ansi16: 4, ansi256: 33, rgb: [38,139,210] }
  VISITED: { ansi16: 5, ansi256: 136, rgb: [108,113,196] }

  BORDER: { ansi16: 8, ansi256: 240, rgb: [88,110,117] }
  FOCUS: { ansi16: 4, ansi256: 33, rgb: [38,139,210] }
  SELECT: { ansi16: 6, ansi256: 37, rgb: [42,161,152] }
  HIGHLIGHT: { ansi16: 3, ansi256: 136, rgb: [181,137,0] }

  DISABLED_BG: { ansi16: 8, ansi256: 237, rgb: [7,54,66] }
  DISABLED_FG: { ansi16: 8, ansi256: 244, rgb: [131,148,150] }

  SCROLLBAR: { ansi16: 8, ansi256: 240, rgb: [88,110,117] }
  SHADOW: { ansi16: 0, ansi256: 232, rgb: [0,30,38] }
  CARET: { ansi16: 7, ansi256: 254, rgb: [238,232,213] }
```

---

至此，这 5 套主题已经是 **TUI 引擎级完整主题模型**，无需再补语义色。
