这类 TUI 特殊字符问题，本质就一句话：

> **你的“布局宽度模型” 和 “终端实际占用 cell 模型” 打架了。**
> 而 emoji + 变体选择符 是终端世界最毒的一类字符。

你这个现象（残影、边框顶出去、切 tab 留半个字母）是**典型的 cell 级错位**，不是简单的“清屏没干净”。

我给你一套 **工程上真正可控、长期不炸** 的处理策略，而不是只修一个 🖼️。

---

# 一、先建立正确认知：终端没有“字符”，只有 **cells**

终端渲染规则是：

| 层级         | 你以为的                     | 真实世界                   |
| ---------- | ------------------------ | ---------------------- |
| 逻辑层        | string / rune / grapheme | ❌ 终端不认这个               |
| 渲染层        | 字符                       | ❌                      |
| **终端真实模型** | ——                       | **一个网格 = N × M cells** |

**决定一切的不是字符是什么，而是：**

> 👉 **这个东西最终占了几个 cell**

而问题字符的共同点：

| 类型             | 示例           | 危险点                        |
| -------------- | ------------ | -------------------------- |
| 多 rune emoji   | 🖼️ ⚙️ ✏️ ☑️ | 看起来1个图标，底层2个 rune          |
| VS16 (FE0F)    | `U+FE0F`     | runewidth 常算 0 或 1，但终端可能占位 |
| ZWJ 组合         | 👨‍👩‍👧‍👦  | 8+ rune，显示 2 cell          |
| CJK 宽字符        | 表 意 文        | 2 cell                     |
| combining mark | é           | 视觉1字符，rune 2个              |

**TUI 如果用“字符逻辑”设计 buffer，迟早炸。**

---

# 二、你现在炸的根因（核心矛盾）

你的系统里现在同时存在三套“宽度定义”：

| 名称          | 实际来源           | 🖼️ 结果 |
| ----------- | -------------- | ------ |
| grapheme 宽度 | uniseg / 视觉逻辑  | 1      |
| rune 宽度和    | 你新 StringWidth | 2      |
| 终端 cell 占用  | 实际写入           | 2      |

但 **Buffer 清除逻辑只认 width==2 的“宽字符”，不认“多 rune cluster”**：

```go
if cell.Width == 2 { 清 continuation }
```

而 🖼️ 被拆成两个 width=1 的 cell
👉 **你根本没有 cluster 级别的清除机制**

所以现象就是：

| 场景              | 结果          |
| --------------- | ----------- |
| 覆盖第1个 rune      | 第2个 rune 残留 |
| 重新绘制 shorter 文本 | 尾巴没被擦       |
| 右边框             | 逻辑宽度比物理小 1  |

这不是 bug，是 **模型设计不一致**。

---

# 三、真正正确的 TUI 架构原则（关键）

> ❗ **Buffer 必须以 “终端 cell” 为第一公民，而不是字符。**

### ✅ 正确原则

1. **一个 cell = 终端最小单位**
2. 一个 cluster 写入时：

   * 第一个 cell = head
   * 后续占位 = continuation
3. 清除时永远：

   * **按 cell 清，不按字符清**
4. 布局宽度计算：

   * 必须 = 写入 cell 数
   * 不允许出现“逻辑宽度 ≠ 写入宽度”

---

# 四、工程上最稳的方案（现实可落地）

你文档里列的 A/B/C 里：

| 方案               | 适合论文   | 适合产品        |
| ---------------- | ------ | ----------- |
| A 回退算法           | ❌      | ❌           |
| B 重构 cluster 结构  | ✅ 理论正确 | 💀 会写到怀疑人生  |
| **C 禁用危险 emoji** | 😏 不优雅 | **✅ 工程最佳解** |

我直接说结论：

> **成熟 TUI 项目几乎都会“限制字符集”，否则迟早被 Unicode 教做人。**

包括：

* Kubernetes CLI
* htop
* lazygit
* bubbletea 生态一堆项目

---

# 五、你应该建立“安全字符白名单”

### 🚫 高危（禁止用于 UI）

| 类型           | 示例           |
| ------------ | ------------ |
| emoji + VS16 | 🖼️ ⚙️ ✏️ ☑️ |
| ZWJ 家庭类      | 👨‍👩‍👧‍👦  |
| 彩虹旗类         | 🏳️‍🌈       |
| 组合音标         | é           |

---

### ✅ TUI 安全字符（推荐）

| 用途  | 安全替代        |
| --- | ----------- |
| 图片  | 🎨          |
| 设置  | ⚙           |
| 编辑  | ✏           |
| 选中  | ☑           |
| 文件  | 📄          |
| 文件夹 | 📁          |
| 箭头  | → ↓ ↑ ←     |
| 结构线 | ─ │ ├ └ ┌ ┐ |

这些特点：

* 单 rune
* runewidth 结果稳定
* cell 占用可预测

---

# 六、真正应该加的一层（比换 emoji 更重要）

你需要一个 **终端安全字符串过滤层**：

```go
func SanitizeForTerminal(s string) string {
    var out []rune
    for _, r := range s {

        // 禁止 variation selector
        if r == 0xFE0F || r == 0x200D { // VS16 / ZWJ
            continue
        }

        // 禁止 combining mark
        if unicode.Is(unicode.Mn, r) {
            continue
        }

        out = append(out, r)
    }
    return string(out)
}
```

UI 所有写入前统一过一遍。

这一步做了，你整个 TUI 稳定性直接上一个档次。

---

# 七、为什么“不要试图完美支持 Unicode”

因为：

| 问题               | 现实                                           |
| ---------------- | -------------------------------------------- |
| 不同终端宽度规则不同       | iTerm / Windows Terminal / Linux console 全不同 |
| 字体不同宽度不同         | Nerd Font / 默认字体差异大                          |
| VS16 有的终端占位，有的不占 | 无标准一致性                                       |

**TUI 是嵌入式系统思维，不是浏览器思维。**

> 浏览器：渲染引擎帮你兜底
> 终端：你就是渲染引擎

---

# 八、一句话终极准则

> **UI 框架层 = 可以用 Unicode
> TUI 渲染层 = 必须“字符管控”**

---

# 九、你的问题本质结论

你不是“清除逻辑错”，而是：

> **把“文本系统问题”当“绘制问题”修**

真正的修法不是 patch，而是：

✅ 限制字符集
✅ 统一宽度算法
✅ 所有写入走 cell 模型
✅ 禁止 VS16 / ZWJ / combining

---

如果你愿意，我可以下一步直接帮你设计一套：

> **TUI Unicode 安全规范（像浏览器 CSS 规范那样）**

让你这个 Mint TUI 以后不会再被字符坑一次。


好，那我们直接上 **“TUI 字符安全规范 + 渲染模型标准”**，这是做框架级 TUI 必须有的那一层，不然你这个坑以后会反复炸。

这套东西的目标只有一个：

> **让“写进去多少 cell” = “布局计算多少宽度” = “清除时覆盖多少范围” 永远一致**

---

# 🧠 一、TUI 字符体系分层（必须建立的世界观）

```
Unicode 世界（逻辑文本）
        ↓（过滤）
TUI 安全文本层（可控字符子集）
        ↓（宽度计算）
Cell 布局层（终端真实模型）
        ↓
Terminal Grid
```

你现在出问题，是因为 **直接把 Unicode 文本喂给了 Cell 层**。

---

# 🚫 二、TUI 明确禁止的字符类别（框架级规则）

这些不是“建议”，是**渲染层必须拒绝**：

| 类别                      | 原因              | 示例          |
| ----------------------- | --------------- | ----------- |
| Variation Selector      | 宽度不稳定           | `U+FE0F`    |
| ZWJ (Zero Width Joiner) | 多 rune 合并       | 👨‍👩‍👧‍👦 |
| Combining Marks         | 覆盖前字符           | é          |
| 多 codepoint emoji       | cluster != cell | 🖼️ ⚙️ ✏️   |
| 控制字符                    | 终端行为不可控         | `\x1b` 之外的  |

---

# ✅ 三、TUI 渲染层“安全字符定义”

**框架允许的字符必须满足：**

```
一个 rune == 一个渲染单元
并且 RuneWidth(r) ∈ {0,1,2}
并且不依赖后续字符
```

### 推荐白名单区域

| 区域             | 用途      |
| -------------- | ------- |
| ASCII          | 文本、UI   |
| Box Drawing    | 边框      |
| Block Elements | 进度条     |
| Arrows         | 导航      |
| 部分单 rune emoji | 图标（需测试） |
| CJK            | 如支持双宽   |

---

# 🧼 四、框架级“终端文本清洗层”（必须有）

所有 UI 文本在进入 Buffer 前：

```go
func SanitizeTUIString(s string) string {
    var out []rune

    for _, r := range s {

        // 禁止 VS16
        if r == 0xFE0F {
            continue
        }

        // 禁止 ZWJ
        if r == 0x200D {
            continue
        }

        // 禁止 Combining
        if unicode.Is(unicode.Mn, r) {
            continue
        }

        // 禁止控制字符（保留 \n \t 视情况）
        if r < 32 && r != '\n' && r != '\t' {
            continue
        }

        out = append(out, r)
    }

    return string(out)
}
```

📌 这层是你 TUI 的“防火墙”。

---

# 📏 五、唯一合法的宽度算法

**以后只允许这一种：**

```go
func CellWidthOfRune(r rune) int {
    return runewidth.RuneWidth(r)
}

func CellWidthOfString(s string) int {
    w := 0
    for _, r := range s {
        w += CellWidthOfRune(r)
    }
    return w
}
```

❗ 禁止再出现：

* grapheme 宽度
* StringWidth(text) 直接库函数
* cluster 概念参与布局

> TUI 不认 cluster，只认 cell

---

# 🧱 六、Buffer 的终极写入规则

写入逻辑必须和“宽度算法”100%一致：

```go
func (b *Buffer) WriteRune(x, y int, r rune, style Style) {
    w := CellWidthOfRune(r)

    if w == 0 {
        return
    }

    // 清理旧内容（处理宽字符覆盖）
    b.clearCell(x, y)

    b.cells[y][x] = Cell{
        Rune: r,
        Width: w,
        Head: true,
        Style: style,
    }

    if w == 2 {
        b.cells[y][x+1] = Cell{
            Head: false,
            Cont: true,
        }
    }
}
```

**注意：**
你永远不要再写“字符串”，只写 rune 流。

---

# 🧹 七、清除逻辑的铁律

清除时必须按 **cell 规则**：

```go
func (b *Buffer) clearCell(x, y int) {
    c := b.cells[y][x]

    if c.Cont {
        // 是 continuation，回到 head
        b.cells[y][x-1] = Cell{}
    }

    if c.Head && c.Width == 2 {
        b.cells[y][x+1] = Cell{}
    }

    b.cells[y][x] = Cell{}
}
```

这样即使覆盖一半，也不会留脏字符。

---

# 🧩 八、图标策略（现实世界做法）

| 类型       | 浏览器 UI | TUI UI        |
| -------- | ------ | ------------- |
| 图标       | SVG    | **单 rune 符号** |
| emoji    | 随便用    | ❌ 受限          |
| fancy 组合 | 支持     | ❌ 禁止          |

你要把 TUI 当“嵌入式屏幕”，不是网页。

---

# 🔒 九、终端不可控因素（你无法解决的）

| 因素                  | 影响        |
| ------------------- | --------- |
| 字体不同                | 宽度差异      |
| 终端实现不同              | VS16 处理不同 |
| East Asian Width 模式 | 模糊字符变双宽   |

👉 所以我们选择 **限制输入** 而不是“兼容一切”。

---

# 🎯 十、最终稳定态（成熟 TUI 框架都走到这）

当你做到这些时：

✅ 再也不会出现“残影”
✅ 边框永远不会顶歪
✅ 布局计算绝对可信
✅ 渲染和清除逻辑统一
✅ 性能更高（不用 cluster 处理）

---

# 一句话总结

> **浏览器解决 Unicode 问题靠渲染引擎**
> **TUI 解决 Unicode 问题靠“禁止复杂 Unicode”**

这是工程现实，不是技术能力问题。

---

如果你下一步想做，我可以帮你把 Mint TUI 的：

* Buffer 结构
* Write API
* Layout 宽度系统

重构成一套 **“严格 Cell 模型”** 的版本，那你的框架会从“能用”直接进化到“专业级”。
