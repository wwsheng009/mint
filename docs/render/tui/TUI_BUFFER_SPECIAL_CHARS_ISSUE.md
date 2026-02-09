# TUI Buffer 特殊字符处理问题分析

## 1. 问题概述

### 1.1 问题描述
在 Mint TUI framework 的 Inspector 组件中，使用某些 emoji 图标（🖼️、⚙️）时，出现以下问题：

1. **内容残余**：Tab 切换时，图标后的字符（如 "element" 中的 'e'）未能正确清除
2. **边框突出**：包含 🖼️ 的行，右侧边框超出预期宽度 1 个字符

### 1.2 问题截图

```
└── 🖼️ BorderedNode                                    │ender ]    [ [5]Reconcile ]  │
│  └── 📦 LayoutNode                                  │
│   │  └── 📝 TextVNode(Runtime Schedulin...)        │
│  └── 🖼️ BorderedNode                                │───────────────────────────┘
   ^^^^^^^^^^^^^^^^                                    ^^^^^^^^^^^^^^^^^^^^^^^^^^^
   问题区域：右侧超出                                    切换后 'e' 未清除
```

---

## 2. 根本原因分析

### 2.1 Unicode 字符分类

| 字符 | Unicode | Rune数量 | runewidth宽度 | 实际占用cells | 问题 |
|------|---------|----------|---------------|--------------|------|
| 📦 | U+1F4E6 | 1 | 2 | 2 | ❌ 无 - 宽字符，width=2 正确 |
| 🖼️ | U+1F5BC + U+FE0F | 2 | 1 | 2 | ⚠️ **多rune grapheme cluster** |
| ⚙️ | U+2699 + U+FE0F | 2 | 1 | 2 | ⚠️ **多rune grapheme cluster** |
| 📝 | U+1F4DD | 1 | 2 | 2 | ❌ 无 - 宽字符，width=2 正确 |
| 🔵 | U+1F535 | 1 | 2 | 2 | ❌ 无 - 宽字符，width=2 正确 |
| → | U+2192 | 1 | 1 | 1 | ❌ 无 |

### 2.2 Multi-rune Grapheme Cluster 的特殊性

**🖼️ 的组成**：
```
U+1F5BC '🖼'  (FRAME WITH PICTURE)
U+FE0F  '️'   (VARIATION SELECTOR-16, 让表情显示为彩色)
```

**关键矛盾**：

1. **UNISEG 视角**：`🖼️` 是 1 个 grapheme cluster（视觉上整体）
   ```go
   g := uniseg.NewGraphemes("🖼️")
   g.Next() // 只有 1 次迭代
   cluster := g.Str() // "🖼️"
   ```

2. **StringWidth 结果**：`runewidth.StringWidth("🖼️")` 返回 **1**
   - 将其视为单个逻辑字符，宽度为 1

3. **TUI Buffer 现实**：**每个 rune 占用 1 个 cell**
   ```
   Buffer.Cells[y][x]   = Cell{Cluster: "🖼️", Width: 1}
   Buffer.Cells[y][x+1] = Cell{Cluster: "️",   Width: 1}
   ```
   - 实际占用 2 个 cell，但每个 cell 的 `Width` 都标记为 1

4. **布局计算错误**：
   ```go
   // Layout 时计算宽度为 1
   width = runewidth.StringWidth("🖼️") // = 1

   // 但实际书写时占用 2 cells
   buffer.WriteString(x, y, "🖼️", style)
   // 写入 (x, y)   = '🖼'
   // 写入 (x+1, y) = '️'
   ```

---

## 3. 当前 Buffer 配置

### 3.1 Cell 结构

```go
// runtime/paint/buffer.go
type Cell struct {
    // 图形簇内容（支持 grapheme clusters）
    Cluster        string

    // 样式信息
    Style          style.Style

    // 这个单元格的显示宽度（1 或 2）
    Width          int

    // 是否是宽字符的延续单元格
    IsContinuation bool

    // 节点 ID（用于调试）
    NodeID         string

    // 选择状态
    Selected       bool
}
```

**关键设计决策**：
- `Width` 只标记 **0 或 2**用于宽字符（如📦）
- 对于多rune emoji（🖼️），每个rune的 `Width = 1`（不认为是宽字符）

### 3.2 SetString 写入逻辑

```go
// runtime/paint/buffer.go
func (b *Buffer) SetString(x, y int, text string, s style.Style) {
    col := x
    g := uniseg.NewGraphemes(text)

    for g.Next() {
        cluster := g.Str()                // 完整图形簇
        width := getClusterWidth(cluster) // cluster宽度

        // 边界检查
        if col >= b.Width {
            break
        }
        if width == 2 && col+1 >= b.Width {
            break
        }

        b.setCluster(col, y, cluster, width, s)
        col += width
    }
}
```

### 3.3 setCluster 继续处理

```go
func (b *Buffer) setCluster(x, y int, cluster string, width int, s style.Style) {
    b.clearCellAt(x, y)

    b.Cells[y][x] = Cell{
        Cluster:        cluster,
        Style:          s,
        Width:          width,
        IsContinuation: false,
    }

    // 只对 width == 2 的字符设置 continuation
    if width == 2 && x+1 < b.Width {
        b.Cells[y][x+1] = Cell{
            IsContinuation: true,
        }
    }
}
```

**对于 🖼️ 的问题**：
- width = 1（`getClusterWidth` 返回 `runewidth.StringWidth("🖼️")` = 1）
- 不设置 `IsContinuation`
- 两个rune独立写入，无关联

### 3.4 clearCellAt 清除逻辑

```go
func (b *Buffer) clearCellAt(x, y int) {
    cell := b.Cells[y][x]

    // 如果是 continuation，清除 head
    if cell.IsContinuation && x > 0 {
        head := b.Cells[y][x-1]
        if head.Width == 2 && head.Cluster != "" {
            b.Cells[y][x-1] = Cell{}
        }
    }

    // 如果是宽字符头，清除右侧 continuation
    if cell.Width == 2 && x+1 < b.Width {
        b.Cells[y][x+1] = Cell{}
    }

    b.Cells[y][x] = Cell{}
}
```

**对于 🖼️ 的清除**：
- 两个rune的 `Width = 1`
- 都不触发任何清除关联逻辑
- 清除第一个rune '🖼' 时，第二个rune '️' 不会被清除 → **残余**

---

## 4. StringWidth 计算演进

### 4.1 老实现（已被替换）

```go
func StringWidthLegacy(text string) int {
    runes := []rune(text)
    if len(runes) == 1 {
        ch := runes[0]
        // Unicode Box Drawing block - 返回 1
        if ch >= 0x2500 && ch <= 0x257F {
            return 1
        }
    }
    // 否则使用 runewidth.StringWidth
    return runewidth.StringWidth(text)
}
```

**问题**：
- `runewidth.StringWidth("🖼️")` = 1
- 但实际占用 2 cells
- 导致布局计算与实际写入不一致

### 4.2 新实现（当前）

```go
func StringWidth(text string) int {
    totalWidth := 0
    for _, r := range text {
        totalWidth += RuneWidth(r)
    }
    return totalWidth
}

func RuneWidth(r rune) int {
    if r >= 0x2500 && r <= 0x257F {
        return 1  // Box Drawing Characters
    }
    return runewidth.RuneWidth(r)
}
```

**改进**：
- 按rune计算宽度
- `"🖼️"` = `RuneWidth('🖼') + RuneWidth('️')` = 1 + 1 = **2**
- **正确反映实际占用cells**

**新问题**：
- 与 layout 计算不一致（layout 可能还使用旧方法）
- 导致边框缩进（计算宽度大1，内容右移）

---

## 5. 当前系统状态的矛盾

### 5.1 三个不同的"宽度"概念

| 概念 | 定义 | 🖼️的值 | 说明 |
|------|------|--------|------|
| **runewidth.StringWidth** | grapheme cluster视觉宽度 | 1 | 将🖼️视为单个字符 |
| **paint.StringWidth (NEW)** | rune宽度之和 | 2 | 按rune计算 |
| **实际buffer占用** | TUI cells数量 | 2 | 每rune占1 cell |

### 5.2 Layout 阶段的假设

```go
// 假设（老的runewidth计算）
width = runewidth.StringWidth("🖼️ element") // = 9

// 但实际buffer占用
runes: [U+1F5BC, U+FE0F, ' ', 'e', 'l', 'e', 'm', 'e', 'n', 't']
cells: 10 个
```

### 5.3 清除阶段的遗漏

```go
// 清除 "🖼️ element" 时
buffer.SetCell(x, y, ' ', style)   // 清除 U+1F5BC '🖼'
buffer.SetCell(x+1, y, ' ', style) // 清除 U+FE0F  '️'

// 但在某些情况下（如Tab切换），只调用：
buffer.clearRegion(bounds)

// clearRegion中：
for x := bounds.X; x < bounds.X + bounds.Width; x++ {
    buffer.SetCell(x, y, ' ', style)
}

// 如果 bounds.Width 是按 runewidth.StringWidth 计算的
// bounds.Width = 9，只清除 9 个cells
// 但 "🖼️ element" 实际 10 个cells
// → 第10个cell（'t'）或第7个cell（'e'）可能未被清除
```

---

## 6. 解决方案评估

### 6.1 方案A：回退到老的StringWidth

```go
func StringWidth(text string) int {
    return runewidth.StringWidth(text)
}
```

**优点**：
- 简单，与layout计算一致
- 不改变现有buffer结构

**缺点**：
- 🖼️ 的宽度计算为1，但实际占用2 cells
- **无法解决边框突出问题**

### 6.2 方案B：改进buffer结构处理多rune clusters

```go
type Cell struct {
    Cluster           string
    Width             int
    IsContinuation    bool
    ClusterHeadIndex  int      // 新增：指向同一cluster的head
    ClusterRuneCount  int      // 新增：这个cluster有多少rune
}
```

**优点**：
- 理论上可以正确处理任何字符

**缺点**：
- **需要大规模重构buffer结构**
- 需要反向查找head（性能问题）
- 边界情况极其复杂（部分覆盖、跨cluster等）
- 与现有width=2的宽字符逻辑冲突
- **实现复杂度极高，风险大**

### 6.3 方案C：禁用多rune emoji

```go
// 替换所有多rune emoji为单rune替代品
🖼️ → 🎨  // art
⚙️  → ⚙   // gear without variation selector
✏️  → ✏   // pencil without variation selector
☑️  → ☑   // checkbox without variation selector
```

**优点**：
- **完美解决问题**
- 实现简单
- 性能无影响
- 所有单rune字符：`runewidth.StringWidth` = `实际占用cells`

**缺点**：
- 失去多rune emoji的视觉效果（但功能相同）
- 可能需要重新测试

### 6.4 方案D：混合方案（推荐）

**核心思想**：
1. 保持当前 `paint.StringWidth`（按rune计算）
2. 替换多rune emoji为单rune版本
3. 审查所有使用 `runewidth.StringWidth` 的地方，改为 `paint.StringWidth`

**实施步骤**：
1. 替换icon
2. 全局搜索 `runewidth.StringWidth` 替换为 `paint.StringWidth`
3. 测试验证

---

## 7. 推荐实施：方案C/D

### 7.1 为何推荐方案C/D

| 因素 | 方案A | 方案B | 方案C/D |
|------|-------|-------|---------|
| **能否解决问题** | ❌ 否 | ✅ 是 | ✅ 是 |
| **实现复杂度** | ⭐ 简单 | ⭐⭐⭐⭐⭐ 极复杂 | ⭐ 简单 |
| **风险** | ⚠️ 中 | 🔴 极高 | ✅ 低 |
| **性能影响** | ✅ 无 | ⚠️ 可能有 | ✅ 无 |
| **维护性** | ⚠️ 中 | ❌ 差 | ✅ 好 |
| **视觉效果** | ✅ 同原 | ✅ 同原 | ⚠️ 轻微差异 |

### 7.2 实施方案C的具体代码

```go
// internal/inspector/tree_view.go
func getIconForType(typeName string) string {
    typeLower := strings.ToLower(typeName)

    switch {
    case strings.Contains(typeLower, "button"):
        return "🔵"
    case strings.Contains(typeLower, "text"):
        return "📝"
    case strings.Contains(typeLower, "hstack"):
        return "→"
    case strings.Contains(typeLower, "vstack"):
        return "↓"
    case strings.Contains(typeLower, "box"):
        return "📦"
    case strings.Contains(typeLower, "border"):
        return "🖼️" → "🎨"  // 替换：art / picture frame
    case strings.Contains(typeLower, "input"):
        return "✏️" → "✏"   // 替换：pencil
    case strings.Contains(typeLower, "checkbox"):
        return "☑️" → "☑"   // 替换：checkbox
    case strings.Contains(typeLower, "select"):
        return "📋"
    default:
        return "📦"
    }
}

// components/display/treeview.go
func (b *TreeViewBuilder) getNodeIcon(nodeType string) string {
    icons := map[string]string{
        "vstack":    "📦",
        "hstack":    "📦",
        "text":      "📝",
        "button":    "🔵",
        "bordered":  "🖼️" → "🎨",  // 替换
        "flex":      "🔧",
        "element":   "📦",
        "component": "⚙️" → "⚙",   // 替换：gear
    }
    // ...
}
```

### 7.3 回退 StringWidth 到老实现（可选）

如果需要完全与layout对齐，可以回退：

```go
// runtime/paint/width.go
func StringWidth(text string) int {
    // 回退到 runewidth.StringWidth
    runes := []rune(text)
    if len(runes) == 1 {
        ch := runes[0]
        if ch >= 0x2500 && ch <= 0x257F {
            return 1
        }
    }
    return runewidth.StringWidth(text)
}
```

**但需要注意**：回退后，如果保留了🖼️，边框问题会再次出现。因此**必须同时替换emoji**。

---

## 8. 测试验证清单

实施方案C后，需验证：

### 8.1 功能测试

- [ ] Inspector 所有tab切换正常
- [ ] TreeView 显示正确，无残余字符
- [ ] TreeView 切换焦点/选中状态正常
- [ ] 所有图标显示正确（🎨, ⚙, ✏, ☑）
- [ ] 边框不对齐问题已解决

### 8.2 宽度计算验证

```go
// 运行测试
for icon, expectedWidth := range map[string]int{"🎨":2, "⚙":1, "📦":2} {
    got := paint.StringWidth(icon)
    if got != expectedWidth {
        log.Fatalf("%s width: got %d, want %d", icon, got, expectedWidth)
    }
}
```

### 8.3 Buffer写入验证

```go
buf := paint.NewBuffer(20, 10)
buf.WriteString(0, 0, "🎨 element", style.Style{})

// 验证 cells
for i := 0; i < 12; i++ {
    cell := buf.Cells[0][i]
    if cell.Cluster == "" {
        log.Fatalf("Cell %d should not be empty", i)
    }
}
```

---

## 9. 总结

### 9.1 核心问题

**TUI Buffer 的物理限制**：每个rune占用1个cell，与Unicode grapheme cluster的逻辑抽象矛盾。

### 9.2 最佳解决方案

**方案C/D**：替换多rune emoji为单rune版本 + 保持当前 `paint.StringWidth`

**理由**：
- 简单、可靠、高风险低
- 完全解决问题
- 不需要大规模重构
- 视觉损失极小

### 9.3 实施优先级

1. **立即执行**：替换🖼️ → 🎨, ⚙️ → ⚙, ✏️ → ✏, ☑️ → ☑
2. **验证**：运行所有测试，确认问题解决
3. **可选**：如果layout仍有问题，审查并统一所有 `StringWidth` 调用

---

## 10. 参考文档

- [Unicode Variation Selectors](https://en.wikipedia.org/wiki/Variation_Selectors_(Unicode_block))
- [runewidth 库](https://github.com/mattn/go-runewidth)
- [uniseg 库]((https://pkg.go.dev/github.com/rivo/uniseg)
- [Unicode Text Segmentation](https://unicode.org/reports/tr29/)
