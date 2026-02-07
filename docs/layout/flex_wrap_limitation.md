# 布局系统限制：flex-wrap 自动换行

## 当前状态

Mint TUI 的布局系统**不支持** CSS `flex-wrap: wrap` 功能。

## 问题场景

当子元素总宽度超过父容器宽度时：

### CSS Flexbox (支持换行)
```css
.container {
  display: flex;
  flex-wrap: wrap;  /* ✅ 自动换行 */
  gap: 8px;
}
```

### Mint TUI (不支持换行)
```go
ui.HStackBuilder(
    item1, item2, item3, // ... many items
).
    Gap(1).
    Build()
// ❌ 如果总宽度超过容器，会溢出或被截断
```

## demo2 当前实现

所有按钮在一行显示：

```go
allButtons := ui.HStackBuilder(
    "[1] Event", "[2] setState", "[3] Scheduler",
    "[4] Render", "[5] Reconcile", "[6] Layout",
    "[7] Paint", "[0] Idle", // 8 个按钮
).
    Gap(1).
    Align(ui.AlignStart).
    Build()
```

**预期宽度计算：**
- 每个按钮约 15-20 字符
- 7 个 gap × 1 字符 = 7 字符
- 边框：2 字符
- 总计：约 130-170 字符

**问题：** 在 100 字符宽度的终端中，按钮会超出边界。

## 解决方案

### 方案 1: 手动分行（推荐）

最简单、最可靠的方法：

```go
row1 := ui.HStackBuilder(buttons[0:4]).Gap(1).Align(ui.AlignStart).Build()
row2 := ui.HStackBuilder(buttons[4:8]).Gap(1).Align(ui.AlignStart).Build()

ui.VStack(row1, row2).Gap(0).Build()
```

**优点：**
- ✅ 完全控制每行的内容
- ✅ 可预测的布局
- ✅ 适配不同屏幕宽度

**缺点：**
- ❌ 需要手动计算分行
- ❌ 固定分行，不响应宽度变化

---

### 方案 2: 实现自动换行组件

创建一个智能的 Wrap 组件，自动计算子元素宽度并换行。

**实现思路：**

```go
// WrapComponent - 自动换行的容器
func WrapComponent(children []ui.VNode, gap int) ui.VNode {
    const screenWidth = 100 // 可从运行时获取

    var rows [][]ui.VNode
    currentRow := []ui.VNode{}
    currentWidth := 0

    for _, child := range children {
        childWidth := estimateWidth(child)

        // 如果当前行放不下，开始新行
        if currentWidth + childWidth > screenWidth && len(currentRow) > 0 {
            rows = append(rows, currentRow)
            currentRow = []ui.VNode{child}
            currentWidth = childWidth
        } else {
            currentRow = append(currentRow, child)
            currentWidth += childWidth + gap
        }
    }

    // 添加最后一行
    if len(currentRow) > 0 {
        rows = append(rows, currentRow)
    }

    // 构建 VStack
    var rowNodes []ui.VNode
    for _, row := range rows {
        rowNodes = append(rowNodes, ui.HStackBuilder(row...).Gap(gap).Align(ui.AlignStart).Build())
    }

    return ui.VStackBuilder(rowNodes...).Gap(0).Build()
}
```

**使用：**

```go
buttons := []ui.VNode{
    btn1, btn2, btn3, btn4, btn5, btn6, btn7, btn8,
}

wrapped := WrapComponent(buttons, 1) // 自动换行，gap=1
```

**优点：**
- ✅ 自动适配屏幕宽度
- ✅ 响应式布局
- ✅ 类似 CSS flex-wrap

**缺点：**
- ❌ 需要实现 width estimation
- ❌ 性能开销（需要测量每个子元素）
- ❌ 复杂度增加

---

### 方案 3: 使用固定宽度容器

限制每个按钮的宽度，确保能放下：

```go
ui.HStackBuilder(
    btn1.Width(12).Build(),
    btn2.Width(12).Build(),
    btn3.Width(12).Build(),
    // ...
).
    Gap(1).
    Build()
```

**计算：**
- 8 个按钮 × 12 字符 = 96 字符
- 7 个 gap × 1 字符 = 7 字符
- 边框 = 2 字符
- 总计 = 105 字符（略微超出 100 字符）

**问题：** 即使固定宽度，仍然可能超出。

---

### 方案 4: 滚动容器

如果按钮太多，使用滚动：

```go
// 类似 CSS overflow-x: auto
scrollable := ui.Scrollable().Content(allButtons).Build()
```

**问题：** 终端 UI 的滚动实现复杂，性能开销大。

---

## 推荐做法

### 短期（当前实现）

**手动分行，适应最常见的场景：**

```go
// 根据终端宽度合理分组
row1 := ui.HStackBuilder(btn1, btn2, btn3, btn4).Gap(1).Build()
row2 := ui.HStackBuilder(btn5, btn6, btn7, btn8).Gap(1).Build()

ui.VStack(row1, row2).Gap(0).Build()
```

**优点：**
- 简单、可靠、性能好
- 适合大多数终端宽度（80-120 字符）
- 易于维护和调试

---

### 中期（功能增强）

**添加智能换行组件：**

```go
// 自动检测屏幕宽度并换行
wrapped := ui.Wrap(
    ui.Config{
        ScreenWidth: 100,  // 从运行时获取
        Gap: 1,
    },
    buttons...
)
```

**需要实现：**
1. `ui.Wrap()` 组件
2. 子元素宽度估算
3. 屏幕宽度检测
4. 动态分行逻辑

---

### 长期（架构改进）

**参考 CSS Flexbox，实现完整的 flex-wrap 支持：**

```go
ui.HStackBuilder(items...).
    FlexWrap(ui.WrapWrap). // 类似 flex-wrap: wrap
    AlignContent(ui.AlignStart).
    Build()
```

**需要的扩展：**
1. `FlexWrap` 属性
2. `AlignContent` 属性（控制多行对齐）
3. 布局引擎支持多行计算
4. 动态行高计算

---

## 当前 demo2 的建议

### 选项 1: 保持单行（当前）

```go
allButtons := ui.HStackBuilder(
    btn1, btn2, btn3, btn4, btn5, btn6, btn7, btn8,
).
    Gap(1).
    Align(ui.AlignStart).
    Build()
```

**适用场景：** 终端宽度 ≥ 120 字符

**问题：** 在 80-100 字符终端会溢出

---

### 选项 2: 手动分行（稳妥）

```go
row1 := ui.HStackBuilder(btn1, btn2, btn3, btn4).Gap(1).Build()
row2 := ui.HStackBuilder(btn5, btn6, btn7, btn8).Gap(1).Build()
content := ui.VStack(row1, row2).Gap(0).Build()
```

**适用场景：** 所有终端宽度

**优点：** 不会溢出，布局稳定

---

### 选项 3: 按功能分组

```go
row1 := ui.HStackBuilder(
    // Pipeline phases
    btn1, btn2, btn3, btn4,
).Gap(1).Build()

row2 := ui.HStackBuilder(
    // More phases + idle
    btn5, btn6, btn7, btn8,
).Gap(1).Build()
```

**优点：**
- 逻辑分组清晰
- 每行功能相关

---

## 技术限制总结

| 特性 | CSS Flexbox | Mint TUI | 说明 |
|------|-------------|----------|------|
| `flex-wrap: nowrap` | ✅ 默认 | ✅ 默认 | 单行，不换行 |
| `flex-wrap: wrap` | ✅ 支持 | ❌ 不支持 | 需要手动分行 |
| `flex-wrap: wrap-reverse` | ✅ 支持 | ❌ 不支持 | - |
| `align-content` | ✅ 支持 | ❌ 不支持 | 多行对齐 |
| 响应式宽度 | ✅ 媒体查询 | ⚠️ 有限 | 终端宽度固定 |

---

## 总结

**当前最佳实践：**

1. ✅ 使用 `HStackBuilder` + `Gap()` 代替手动空格
2. ✅ 使用 `Align(ui.AlignStart)` 从左到右排列
3. ✅ 使用 `FillWidth()` 让容器横向拉伸
4. ⚠️ 手动分行以适应不同终端宽度
5. ❌ 不支持自动换行（需要手动实现）

**未来改进方向：**

1. 实现 `Wrap` 组件（自动换行）
2. 支持屏幕宽度检测
3. 实现完整的 flex-wrap 支持
4. 添加响应式布局支持

---

## 相关资源

- **CSS Flexbox:** https://developer.mozilla.org/en-US/docs/Web/CSS/flex-wrap
- **Mint TUI 布局系统:** `docs/layout/flex_comparison.md`
- **布局问题修复:** `docs/layout/issue/`
