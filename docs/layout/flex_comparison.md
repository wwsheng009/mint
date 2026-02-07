# Mint TUI 布局系统 vs CSS Flexbox

本文档对比 Mint TUI 的布局系统与 CSS Flexbox，帮助前端开发者快速理解和使用。

## 快速对照表

| CSS Flexbox | Mint TUI | 说明 |
|-------------|----------|------|
| `display: flex` | `ui.HStack()` / `ui.VStack()` | 创建 flex 容器 |
| `flex-direction: row` | `ui.HStack()` | 水平布局 |
| `flex-direction: column` | `ui.VStack()` | 垂直布局 |
| `justify-content` | `.Align()` | 主轴对齐 |
| `align-items` | `.AlignCross()` | 交叉轴对齐 |
| `gap` | `.Gap()` | 子元素间距 |
| `flex-grow` | `.Flex()` | 拉伸因子 |
| `padding` | `.Padding()` | 内边距 |
| `align-items: stretch` | `.Stretch()` | 子元素拉伸到交叉轴 |
| `width: 100%` | `.FillWidth()` | 填充父容器宽度 ✨ |
| `height: 100%` | `.FillHeight()` | 填充父容器高度 ✨ |

---

## 完整对比

### 1. 创建 Flex 容器

#### CSS
```css
.container {
  display: flex;
  flex-direction: row; /* or column */
}
```

#### Mint TUI
```go
// 水平布局 (flex-direction: row)
ui.HStack(item1, item2, item3)

// 垂直布局 (flex-direction: column)
ui.VStack(item1, item2, item3)

// 使用 Builder 模式（推荐）
ui.HStackBuilder(item1, item2, item3).
    Gap(1).
    Align(ui.AlignCenter).
    Build()
```

---

### 2. 主轴对齐 (Main Axis Alignment)

#### CSS - justify-content
```css
.container {
  justify-content: flex-start;    /* align-items: flex-start */
  justify-content: center;        /* align-items: center */
  justify-content: flex-end;      /* align-items: flex-end */
  justify-content: space-between; /* align-items: space-between */
  justify-content: space-around;  /* align-items: space-around */
}
```

#### Mint TUI - .Align()
```go
ui.HStackBuilder(items...).
    Align(ui.AlignStart).       // flex-start
    Align(ui.AlignCenter).      // center
    Align(ui.AlignEnd).         // flex-end
    Align(ui.AlignSpaceBetween).// space-between
    Align(ui.AlignSpaceAround). // space-around
    Build()
```

**对应关系：**
- HStack: 主轴 = 水平方向 (`.Align()` 控制 X 轴)
- VStack: 主轴 = 垂直方向 (`.Align()` 控制 Y 轴)

---

### 3. 交叉轴对齐 (Cross Axis Alignment)

#### CSS - align-items
```css
.container {
  align-items: flex-start;
  align-items: center;
  align-items: flex-end;
  align-items: stretch; /* 拉伸到填充交叉轴 */
}
```

#### Mint TUI - .AlignCross()
```go
ui.HStackBuilder(items...).
    AlignCross(ui.AlignStart).  // flex-start
    AlignCross(ui.AlignCenter). // center
    AlignCross(ui.AlignEnd).    // flex-end
    Build()

// 拉伸所有子元素到交叉轴
ui.HStackBuilder(items...).
    Stretch(). // 相当于 align-items: stretch
    Build()
```

**对应关系：**
- HStack: 交叉轴 = 垂直方向 (`.AlignCross()` 控制 Y 轴)
- VStack: 交叉轴 = 水平方向 (`.AlignCross()` 控制 X 轴)

---

### 4. 子元素间距 (Gap)

#### CSS
```css
.container {
  gap: 8px;          /* 统一间距 */
  gap: 8px 16px;     /* 行间距 列间距 */
  row-gap: 8px;      /* 行间距 */
  column-gap: 16px;  /* 列间距 */
}
```

#### Mint TUI
```go
ui.HStackBuilder(items...).
    Gap(1). // 1 个空格的间距
    Build()

ui.VStackBuilder(items...).
    Gap(0). // 无间距
    Build()
```

**注意：** TUI 不支持行/列分别设置（终端环境下通常不需要）

---

### 5. Flex Grow (拉伸因子)

#### CSS
```css
.item {
  flex-grow: 0;  /* 不拉伸（默认） */
  flex-grow: 1;  /* 拉伸填充可用空间 */
  flex-grow: 2;  /* 拉伸 2 倍比例 */
}

/* 简写 */
.item {
  flex: 1; /* flex-grow: 1, flex-shrink: 1, flex-basis: 0% */
}
```

#### Mint TUI
```go
// 方式 1: 使用 Flex() 函数
ui.Flex(item, 1) // flex-grow: 1
ui.Flex(item, 2) // flex-grow: 2

// 方式 2: 使用 Builder
ui.HStackBuilder(
    ui.Text("Item 1"),
    ui.Text("Item 2"),
).
    Flex(1). // 自身拉伸（在 VStack 中 = 垂直拉伸）
    Build()
```

**重要区别：**
- CSS `flex-grow` 在主轴方向拉伸
- Mint TUI `Flex()` 也只在主轴方向拉伸
  - HStack 的主轴 = 水平 → `.Flex()` 拉伸宽度
  - VStack 的主轴 = 垂直 → `.Flex()` 拉伸高度

---

### 6. 拉伸到父容器大小 (Fill Parent)

#### CSS
```css
.item {
  width: 100%;   /* 填充父容器宽度 */
  height: 100%;  /* 填充父容器高度 */
}
```

#### Mint TUI ✨ 新增
```go
// 填充父容器宽度（用于 VStack 的子元素）
ui.Bordered().
    Child(content).
    FillWidth().  // width: 100%
    Build()

// 填充父容器高度（用于 HStack 的子元素）
ui.Bordered().
    Child(content).
    FillHeight(). // height: 100%
    Build()
```

**与 Flex() 的区别：**
| 方法 | 拉伸方向 | 适用场景 |
|------|----------|----------|
| `Flex(n)` | 主轴 | 按比例分配主轴空间 |
| `FillWidth()` | 宽度（VStack 交叉轴） | 填充整个宽度 |
| `FillHeight()` | 高度（HStack 交叉轴） | 填充整个高度 |

---

### 7. Padding (内边距)

#### CSS
```css
.container {
  padding: 10px;           /* 全方向 */
  padding: 10px 20px;      /* 垂直 水平 */
  padding: 10px 20px 30px 40px; /* 上 右 下 左 */
}
```

#### Mint TUI
```go
ui.HStackBuilder(items...).
    Padding(10, 20, 30, 40). // 上, 右, 下, 左
    Build()

// 简写
ui.HStackBuilder(items...).
    Padding(10, 0, 10, 0). // 上下 10，左右 0
    Build()
```

---

## 实战示例

### 示例 1: 居中的单行内容

#### CSS
```css
.container {
  display: flex;
  flex-direction: row;
  justify-content: center; /* 水平居中 */
  align-items: center;     /* 垂直居中 */
  height: 100vh;
}
```

#### Mint TUI
```go
ui.HStackBuilder(
    ui.Text("Centered Content"),
).
    Align(ui.AlignCenter).   // 水平居中
    AlignCross(ui.AlignCenter). // 垂直居中
    FillHeight().            // height: 100vh
    Build()
```

---

### 示例 2: 两侧固定，中间自适应

#### CSS
```css
.container {
  display: flex;
  flex-direction: row;
}

.left { width: 200px; }
.right { width: 200px; }
.center { flex-grow: 1; }
```

#### Mint TUI
```go
ui.HStack(
    ui.Box().Width(50).Build(),   // 左侧固定
    ui.Box().Flex(1).Build(),      // 中间自适应 (flex-grow: 1)
    ui.Box().Width(50).Build(),   // 右侧固定
)
```

---

### 示例 3: 响应式头部（demo2 实际应用）

#### CSS
```css
.header {
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 60px;
}

.title {
  text-align: center;
}
```

#### Mint TUI
```go
func HeaderPanel() ui.VNode {
    title := ui.HStackBuilder(
        app.NewTextBuilder("Runtime Scheduling Pipeline Visualization").
            Style(style.FgBold(theme.Text())).
            Build(),
    ).
        Gap(0).
        Align(ui.AlignCenter). // justify-content: center
        Build()

    return ui.Bordered().
        Style(string(theme.Primary())).
        Child(title).
        FillWidth(). // width: 100%
        Build()
}
```

---

### 示例 4: 卡片布局 (Grid-like)

#### CSS
```css
.card-container {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  gap: 16px;
}

.card {
  flex: 0 0 200px; /* 不拉伸，固定 200px */
}
```

#### Mint TUI
```go
// 注意：TUI 不支持 flex-wrap，需要手动分行
row1 := ui.HStack(
    ui.Box().Width(30).Build(),
    ui.Box().Width(30).Build(),
    ui.Box().Width(30).Build(),
).Gap(2).Build()

row2 := ui.HStack(
    ui.Box().Width(30).Build(),
    ui.Box().Width(30).Build(),
    ui.Box().Width(30).Build(),
).Gap(2).Build()

ui.VStack(row1, row2).Gap(1).Build()
```

---

## 高级用法

### 1. 嵌套布局

#### CSS
```css
<div style="display: flex; flex-direction: column; height: 100vh;">
  <header style="flex: 0 0 60px;">Header</header>
  <main style="flex: 1;">Main Content</main>
  <footer style="flex: 0 0 40px;">Footer</footer>
</div>
```

#### Mint TUI
```go
ui.VStack(
    ui.Bordered().Child(ui.Text("Header")).Height(3).Build(),
    ui.Box().Flex(1).Build(),      // flex: 1 (填充剩余空间)
    ui.Bordered().Child(ui.Text("Footer")).Height(2).Build(),
).
    Gap(0).
    FillHeight(). // height: 100vh
    Build()
```

---

### 2. 混合对齐

#### CSS
```css
.container {
  display: flex;
  flex-direction: row;
  justify-content: space-between; /* 主轴：两端对齐 */
  align-items: center;            /* 交叉轴：居中 */
}
```

#### Mint TUI
```go
ui.HStackBuilder(left, right).
    Align(ui.AlignSpaceBetween). // justify-content: space-between
    AlignCross(ui.AlignCenter).  // align-items: center
    Build()
```

---

### 3. 比例布局

#### CSS
```css
.container {
  display: flex;
}

.sidebar { flex: 1; }   /* 1 份 */
.main { flex: 3; }      /* 3 份 */
```

#### Mint TUI
```go
ui.HStack(
    ui.Box().Flex(1).Build(),  // 1 份
    ui.Box().Flex(3).Build(),  // 3 份
)
```

---

## Mint TUI 特有功能

### 1. FillWidth / FillHeight

Mint TUI 新增的功能，用于精确控制单个组件的拉伸：

```go
// 只让标题横向拉伸，其他组件保持原样
ui.VStackBuilder(
    ui.Bordered().FillWidth().Build(),  // ✅ 横向拉伸
    ui.Box().Build(),                   // ❌ 保持原样
    ui.Box().Build(),                   // ❌ 保持原样
).Build()
```

**优势：**
- 不影响其他子元素
- 明确的拉伸方向（只横向或只纵向）
- 比 `Flex()` 更清晰的语义

---

### 2. 主题集成

```go
ui.HStackBuilder(items...).
    FgColor(theme.Primary()).  // 直接使用主题颜色
    BgColor(theme.BG()).
    Build()
```

---

## 限制和注意事项

### 不支持的 CSS 功能

1. **flex-wrap**: TUI 不支持自动换行
   - 解决方案：手动分行，使用 VStack + 多个 HStack

2. **order**: TUI 不支持子元素重排序
   - 解决方案：调整代码中的子元素顺序

3. **align-self**: TUI 不支持单个子元素覆盖对齐设置
   - 解决方案：使用嵌套布局

4. **flex-basis**: TUI 不支持设置基础大小
   - 解决方案：使用 `.Width()` / `.Height()`

---

## 最佳实践

### 1. 优先使用 Builder 模式

```go
// ✅ 推荐：清晰易读
ui.HStackBuilder(items...).
    Gap(1).
    Align(ui.AlignCenter).
    AlignCross(ui.AlignCenter).
    Build()

// ❌ 不推荐：使用 Flex 包装器（不灵活）
ui.Flex(
    ui.HStack(items...),
    1,
)
```

---

### 2. 明确拉伸方向

```go
// ✅ 推荐：语义明确
ui.Bordered().FillWidth().Build()   // 我要横向拉伸
ui.Bordered().FillHeight().Build()  // 我要纵向拉伸

// ⚠️ 谨慎使用：需要理解上下文
ui.Bordered().Flex(1).Build() // 在 VStack 中纵向拉伸，在 HStack 中横向拉伸
```

---

### 3. 组合使用对齐选项

```go
// 完全居中
ui.HStackBuilder(content).
    Align(ui.AlignCenter).      // 水平居中
    AlignCross(ui.AlignCenter). // 垂直居中
    FillWidth().                // 填充宽度
    Build()
```

---

### 4. 使用 Gap 而非 Margin

```go
// ✅ 推荐：使用 Gap
ui.HStack(items...).Gap(1).Build()

// ❌ 不推荐：手动添加 spacer
ui.HStack(item1, ui.Text(" "), item2, ui.Text(" "), item3)
```

---

## 总结

Mint TUI 的布局系统在设计上参考了 CSS Flexbox，但针对终端环境进行了优化：

| 特性 | CSS Flexbox | Mint TUI |
|------|-------------|----------|
| 学习曲线 | 中等 | 简单（API 更直观） |
| 语义 | `.Align()` = justify-content **或** align-items | `.Align()` = 主轴，`.AlignCross()` = 交叉轴 |
| 拉伸控制 | `flex-grow` + `width/height` | `Flex()` + `FillWidth/Height` ✨ |
| 性能 | 浏览器优化 | 终端优化（无重排） |
| 响应式 | 媒体查询 | 自适应（终端固定大小） |

**核心优势：**
- ✅ 更清晰的 API（分离主轴和交叉轴）
- ✅ 新增 `FillWidth/Height` 精确控制
- ✅ 类型安全（编译时检查）
- ✅ 链式调用（代码更简洁）
