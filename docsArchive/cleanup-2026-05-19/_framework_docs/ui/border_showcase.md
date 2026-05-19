# Border Showcase - 边框展示效果文档

本文档展示了 Mint TUI 框架中所有边框样式的预期效果。

## 运行测试

```bash
cd examples/sandbox/dump_buffer
go test -v -run TestBorderShowcaseAll
```

## 边框样式说明

### 1. Single Style (单线边框) - 默认样式

使用 Unicode 单线框字符：
- 左上角: `┌`
- 右上角: `┐`
- 左下角: `└`
- 右下角: `┘`
- 横线: `─`
- 竖线: `│`

#### 预期效果

**无边框标题:**
```
┌──────────────┐
│Single Border │
└──────────────┘
```

**带标题 "Title":**
```
┌─── Title ───┐
│Single with  │
│Label        │
└─────────────┘
```

---

### 2. Double Style (双线边框)

使用 Unicode 双线框字符：
- 左上角: `╔`
- 右上角: `╗`
- 左下角: `╚`
- 右下角: `╝`
- 横线: `═`
- 竖线: `║`

#### 预期效果

**无边框标题:**
```
╔══════════════╗
║Double Border ║
╚══════════════╝
```

**带标题 "Settings":**
```
╔══ Settings ══╗
║Double with   ║
║Label         ║
╚══════════════╝
```

---

### 3. Rounded Style (圆角边框)

使用圆角和单线字符：
- 左上角: `╭`
- 右上角: `╮`
- 左下角: `╰`
- 右下角: `╯`
- 横线: `─`
- 竖线: `│`

#### 预期效果

**无边框标题:**
```
╭──────────────╮
│Rounded Cornes│
╰──────────────╯
```

**带标题 "Info":**
```
╭─── Info ────╮
│Rounded with │
│Label        │
╰─────────────╯
```

---

### 4. Dashed Style (虚线/ASCII边框)

使用 ASCII 字符：
- 角落: `+`
- 横线: `-`
- 竖线: `|`

#### 预期效果

**无边框标题:**
```
+--------------+
|Dashed Border |
+--------------+
```

**带标题 "ASCII":**
```
+--- ASCII ---+
|Dashed with  |
|Label        |
+-------------+
```

---

## 多行内容

### 单线边框 - 多行内容

```
┌──── Multiple Lines ────┐
│Line 1: First content   │
│Line 2: Second content  │
│Line 3: Third content   │
└────────────────────────┘
```

### 双线边框 - 表格内容

```
╔════ Data Grid ═════════╗
║┌─────────┬─────────┐  ║
║│ Column1 │ Column2 │  ║
║├─────────┼─────────┤  ║
║│ Data A  │ Data B  │  ║
║└─────────┴─────────┘  ║
╚════════════════════════╝
```

## 宽字符支持

边框渲染正确处理宽字符（中文、日文、Emoji），不会出现错位：

```
┌── Wide Characters ────┐
│English: Hello         │
│Chinese: 你好世界       │
│Japanese: こんにちは   │
│Emoji: 😀🎉🚀          │
└───────────────────────┘
```

## 嵌套边框

支持多层边框嵌套：

### 双层嵌套

```
┌───── Outer Border ───────┐
│Content above nested      │
│┌───── Inner ───────┐    │
││Nested content      │    │
│└───────────────────┘    │
│Content below nested      │
└──────────────────────────┘
```

### 三层嵌套 (混合样式)

```
┌────── Level 1 ─────────┐
│╔═══ Level 2 ═════════╗ │
│║╭── Level 3 ──────╮  ║ │
│║│Deeply nested     │  ║ │
│║╰──────────────────╯  ║ │
│╚══════════════════════╝ │
└────────────────────────┘
```

## 所有样式网格

```
Border Style Showcase

┌── single ──┐  ╔══ double ══╗
│Style 1     │  ║Style 2      ║
└────────────┘  ╚════════════╝

╭── rounded ──╮  +--- dashed ---+
│Style 3      │  |Style 4       |
╰─────────────╯  +--------------+
```

## 颜色变体

边框支持多种颜色（需要终端支持 ANSI 颜色）：

```
Border Color Variations

┌─ red ─┐ ┌─ green ─┐ ┌─ blue ─┐
│Red    │ │Green    │ │Blue    │
└───────┘ └─────────┘ └────────┘

┌ yellow ┐ ┌ magenta ┐ ┌─ cyan ─┐
│Yellow  │ │Magenta  │ │Cyan    │
└────────┘ └─────────┘ └────────┘
```

## API 使用示例

```go
// 单线边框（默认）
ui.Bordered().Child(content).Build()

// 带标题的单线边框
ui.Bordered().Label("Title").Child(content).Build()

// 双线边框
ui.Bordered().Style("double").Child(content).Build()

// 圆角边框
ui.Bordered().Style("rounded").Child(content).Build()

// 虚线边框
ui.Bordered().Style("dashed").Child(content).Build()

// 自定义颜色
ui.Bordered().Color("red").Label("Error").Child(content).Build()
```

## 边框样式对比表

| 样式名称 | 样式参数 | 角落字符 | 线条字符 | 适用场景 |
|---------|---------|---------|---------|---------|
| Single | `"single"` 或 `""` | `┌┐└┘` | `─│` | 通用、默认 |
| Double | `"double"` | `╔╗╚╝` | `═║` | 突出显示、标题 |
| Rounded | `"rounded"` | `╭╮╰╯` | `─│` | 柔和界面 |
| Dashed | `"dashed"` | `++` | `-\|` | ASCII 兼容、调试 |

## 技术说明

### 边框不占用布局空间

边框是装饰性的，渲染在内容区域**外部**，不参与布局计算。

- 内容宽度: `contentWidth`
- 边框总宽度: `contentWidth + 2` (左右各加1个字符)
- 内容高度: `contentHeight`
- 边框总高度: `contentHeight + 2` (上下各加1个字符)

### 宽字符处理

边框字符被强制设置为单宽度 (`width=1`)，避免与宽字符（如中文）的延续单元格机制冲突。

修复前: `runewidth.RuneWidth('┌')` 返回 `2`
修复后: `getRuneWidth('┌')` 返回 `1`

这防止了边框字符之间的相互覆盖问题。
