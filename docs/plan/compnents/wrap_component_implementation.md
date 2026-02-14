# Wrap 组件实现计划

## 背景

用户需要实现类似 CSS `flex-wrap: wrap` 的功能，让子元素在容器宽度不够时自动换行。

当前 Mint TUI 的布局系统（HStack/VStack）**不支持自动换行**，这是一个被明确记录的限制（参见 `docs/layout/flex_wrap_limitation.md`）。

## 目标

实现一个 Wrap 组件，支持：
- 根据容器宽度自动将子元素分到多行
- 类似 CSS 的 `flex-wrap: wrap`
- 遵循 Mint TUI 现有的 Builder 模式和 API 设计

## 设计策略

### 核心思想

Wrap 组件不需要自定义布局逻辑，而是：
1. 在 **Build 阶段**计算如何分行
2. 将自己转换为 **VStack 包含多个 HStack**（每行一个）
3. 让现有的布局引擎处理所有定位和渲染

这种"委托模式"与 `flex_wrap_limitation.md` 中提到的方案一致，已经被证明可行。

### 架构优势

- ✅ **无需修改布局引擎** - 复用现有的 HStack/VStack 逻辑
- ✅ **易于理解和调试** - 转换后的结构清晰可见
- ✅ **性能优化** - Build 时计算一次，布局时直接使用
- ✅ **遵循现有模式** - 与 HStack/VStack/Grid 保持一致

---

## 数据结构

### 1. WrapNode 结构

**文件**: `components/layout/wrap.go` (新建)

```go
package layout

import (
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/style"
    ui "github.com/wwsheng009/mint/ui"
)

// WrapNode represents a wrapping layout container
// Similar to CSS flex-wrap: wrap
type WrapNode struct {
    *ui.ElementVNode

    // Configuration
    gap        int          // Spacing between items in the same row
    rowGap     int          // Spacing between rows (0 = use gap)
    align      ui.Align     // Main-axis alignment for each row
    screenWidth int         // Container width for wrapping calculation

    // Internal state (calculated during Build)
    rows [][]ui.VNode       // Pre-calculated rows (for debugging)
}
```

### 2. WrapBuilder 结构

```go
// WrapBuilder provides fluent API for building wrap layouts
type WrapBuilder struct {
    node     *WrapNode
    children []ui.VNode
    // Optional: width cache for performance
    widthCache map[ui.VNode]int
}

// WrapBuilder creates a new wrap layout builder
func WrapBuilder(children ...ui.VNode) *WrapBuilder {
    return &WrapBuilder{
        node: &WrapNode{
            ElementVNode: ui.NewElement("wrap"),
            gap:          1,              // Default gap
            rowGap:       0,              // Default row gap (use gap)
            align:        ui.AlignStart,  // Default alignment
            screenWidth:  80,             // Default width
        },
        children: children,
        widthCache: make(map[ui.VNode]int), // Pre-initialize cache
    }
}
```

---

## 实现步骤

### 步骤 1: Builder 方法

**文件**: `components/layout/wrap.go`

```go
// Gap sets spacing between items in the same row
func (b *WrapBuilder) Gap(n int) *WrapBuilder {
    b.node.gap = n
    return b
}

// RowGap sets spacing between rows (0 = use gap value)
func (b *WrapBuilder) RowGap(n int) *WrapBuilder {
    b.node.rowGap = n
    return b
}

// Align sets main-axis alignment for each row
func (b *WrapBuilder) Align(a ui.Align) *WrapBuilder {
    b.node.align = a
    return b
}

// ScreenWidth sets container width for wrap calculation
// This determines when to break to a new row
func (b *WrapBuilder) ScreenWidth(width int) *WrapBuilder {
    b.node.screenWidth = width
    return b
}

// Width sets explicit width (alias for ScreenWidth)
func (b *WrapBuilder) Width(n int) *WrapBuilder {
    return b.ScreenWidth(n)
}

// Style sets visual style
func (b *WrapBuilder) Style(s style.Style) *WrapBuilder {
    b.node.SetStyle(s)
    return b
}

// Key sets key for diffing
func (b *WrapBuilder) Key(key string) *WrapBuilder {
    b.node.SetKey(key)
    return b
}
```

---

### 步骤 2: 宽度估算

**文件**: `components/layout/wrap.go`

```go
// estimateWidth estimates the display width of a VNode
// Priority order:
// 1. Explicit width prop
// 2. Measure interface (if available)
// 3. Component-specific logic (Button, Text, Input, etc.)
// 4. Default fallback
func (b *WrapBuilder) estimateWidth(child ui.VNode) int {
    if child == nil {
        return 0
    }

    // Check cache first
    if w, ok := b.widthCache[child]; ok {
        return w
    }

    var width int

    // PRIORITY 1: Check explicit width prop
    if props := child.Props(); props != nil {
        if w := props.GetInt("width"); w > 0 {
            width = w
            goto cache
        }
    }

    // PRIORITY 2: Use Measure interface if available
    if measurable, ok := child.(interface {
        Measure(runtime.BoxConstraints) runtime.Size
    }); ok {
        // Measure with unbounded constraints
        size := measurable.Measure(runtime.BoxConstraints{
            MinWidth:  0,
            MaxWidth:  runtime.Infinity,
            MinHeight: 0,
            MaxHeight: runtime.Infinity,
        })
        width = size.Width
        goto cache
    }

    // PRIORITY 3: Component-specific logic

    // Button: label + brackets and focus indicator
    if btn, ok := child.(interface{ Label() string }); ok {
        label := btn.Label()
        if label == "" {
            label = " "
        }
        width = utf8.RuneCountInString(label) + 4 // "[label]" + focus
        goto cache
    }

    // Text: content length
    if text := ui.GetTextContent(child); text != "" {
        width = utf8.RuneCountInString(text)
        goto cache
    }

    // Input: value or placeholder width
    if input, ok := child.(interface{
        Value() string
        Placeholder() string
    }); ok {
        value := input.Value()
        if value == "" {
            value = input.Placeholder()
        }
        if value != "" {
            width = utf8.RuneCountInString(value) + 2 // ":value:"
        }
        goto cache
    }

    // PRIORITY 4: Default fallback
    width = 10 // Minimum reasonable width

cache:
    // Cache result
    b.widthCache[child] = width
    return width
}
```

---

### 步骤 3: 行计算逻辑

**文件**: `components/layout/wrap.go`

```go
// calculateRows divides children into rows based on screenWidth
func (b *WrapBuilder) calculateRows() [][]ui.VNode {
    var rows [][]ui.VNode
    currentRow := []ui.VNode{}
    currentWidth := 0
    screenWidth := b.node.screenWidth

    for i, child := range b.children {
        childWidth := b.estimateWidth(child)

        // Check if we need to wrap
        // Wrap if:
        // 1. Current row is not empty
        // 2. Adding this child would exceed screen width (including gap)
        shouldWrap := len(currentRow) > 0 &&
                     (currentWidth + childWidth + b.node.gap > screenWidth)

        if shouldWrap {
            // Finish current row
            rows = append(rows, currentRow)
            // Start new row with this child
            currentRow = []ui.VNode{child}
            currentWidth = childWidth
        } else {
            // Add to current row
            currentRow = append(currentRow, child)
            if len(currentRow) > 1 {
                // Add gap for all but first item
                currentWidth += b.node.gap
            }
            currentWidth += childWidth
        }

        // Debug logging (optional)
        if os.Getenv("TUI_WRAP_DEBUG") == "true" {
            fmt.Fprintf(os.Stderr, "[Wrap] child %d: width=%d, currentWidth=%d, shouldWrap=%v\n",
                i, childWidth, currentWidth, shouldWrap)
        }
    }

    // Don't forget the last row
    if len(currentRow) > 0 {
        rows = append(rows, currentRow)
    }

    return rows
}
```

---

### 步骤 4: Build 方法

**文件**: `components/layout/wrap.go`

```go
// Build converts the Wrap node into a VStack of HStacks
// This is where the wrapping logic happens
func (b *WrapBuilder) Build() ui.VNode {
    // Handle empty children
    if len(b.children) == 0 {
        return ui.VStack()
    }

    // Calculate rows based on child widths and screen width
    rows := b.calculateRows()

    // Store rows for debugging
    b.node.rows = rows

    // Convert each row to an HStack
    var rowNodes []ui.VNode
    for _, row := range rows {
        rowBuilder := HStackBuilder(row...).
            Gap(b.node.gap).
            Align(b.node.align)
        rowNodes = append(rowNodes, rowBuilder.Build())
    }

    // Determine row gap
    rowGap := b.node.rowGap
    if rowGap == 0 {
        rowGap = b.node.gap // Use gap if rowGap not specified
    }

    // Return VStack containing all rows
    result := VStackBuilder(rowNodes...).
        Gap(rowGap).
        Build()

    // Copy style and key from WrapNode to result
    result.SetStyle(b.node.Style())
    if b.node.Key() != "" {
        result.SetKey(b.node.Key())
    }

    return result
}
```

---

## 集成点

### 1. 在 app 包导出

**文件**: `app/app.go`

在现有的布局组件导出部分添加：

```go
var (
    // ... existing exports ...

    // Layout components
    HStack        = layout.HStack
    VStack        = layout.VStack
    Box           = layout.Box
    Spacer        = layout.Spacer
    Wrap          = layout.Wrap          // NEW
    WrapBuilder   = layout.WrapBuilder   // NEW
)
```

### 2. 在 ui 包提供便捷函数

**文件**: `ui/layout.go`

```go
// Wrap creates a wrapping layout container
// Automatically wraps children to multiple rows based on width
func Wrap(children ...VNode) VNode {
    return layout.Wrap(children...)
}

// WrapBuilder creates a wrapping layout container builder for method chaining
func WrapBuilder(children ...VNode) *layout.WrapBuilder {
    return layout.WrapBuilder(children...)
}
```

---

## API 使用示例

### 示例 1: demo2 的 ControlPanel

**之前的实现**（手动分行）：
```go
row1 := ui.HStackBuilder(btn1, btn2, btn3, btn4).Gap(1).Build()
row2 := ui.HStackBuilder(btn5, btn6, btn7, btn8).Gap(1).Build()
content := ui.VStack(row1, ui.Text(""), row2)
```

**使用 Wrap 后**：
```go
allButtons := ui.WrapBuilder(
    btn1, btn2, btn3, btn4,
    btn5, btn6, btn7, btn8,
).
    Gap(1).
    RowGap(0).
    ScreenWidth(98). // 100 - border (2)
    Build()
```

### 示例 2: 不同对齐方式

```go
// 居中对齐每一行
wrapped := ui.WrapBuilder(buttons...).
    Gap(2).
    Align(ui.AlignCenter).  // 每行居中
    ScreenWidth(80).
    Build()

// 两端对齐（第一个靠左，最后一个靠右）
wrapped := ui.WrapBuilder(buttons...).
    Gap(1).
    Align(ui.AlignSpaceBetween).
    ScreenWidth(80).
    Build()
```

### 示例 3: 响应式布局

```go
// 可以根据不同终端宽度使用不同配置
if screenWidth >= 120 {
    // 宽终端：所有按钮在一行
    wrapped = ui.WrapBuilder(buttons...).
        Gap(1).
        ScreenWidth(screenWidth).
        Build()
} else {
    // 窄终端：自动换行
    wrapped = ui.WrapBuilder(buttons...).
        Gap(1).
        ScreenWidth(screenWidth).
        Build()
}
```

---

## 文件清单

### 新建文件

1. **`components/layout/wrap.go`** (约 350 行)
   - WrapNode 结构体
   - WrapBuilder 结构体
   - estimateWidth() 方法
   - calculateRows() 方法
   - Build() 方法
   - 所有 Builder 方法（Gap, RowGap, Align, ScreenWidth 等）

2. **`components/layout/wrap_test.go`** (约 250 行)
   - 基本功能测试
   - 边缘案例测试
   - 对齐方式测试
   - 性能基准测试

3. **`examples/wrap_demo/main.go`** (约 100 行)
   - Wrap 组件演示程序
   - 展示自动换行功能
   - 提供交互式测试

4. **`docs/layout/wrap_component.md`** (约 200 行)
   - Wrap 组件使用文档
   - API 参考
   - 与 CSS flex-wrap 对比
   - 实际使用示例

### 修改文件

1. **`app/app.go`**
   - 添加 Wrap 和 WrapBuilder 的导出
   - 在布局组件部分添加

2. **`ui/layout.go`**
   - 添加 Wrap() 和 WrapBuilder() 便捷函数

3. **`docs/layout/flex_wrap_limitation.md`**
   - 更新限制说明，添加 Wrap 组件作为解决方案
   - 添加到 wrap_component.md 的链接

---

## 测试验证

### 单元测试

**文件**: `components/layout/wrap_test.go`

```go
func TestWrap_BasicWrapping(t *testing.T) {
    btn1 := ui.Text("Button1")
    btn2 := ui.Text("Button2")
    btn3 := ui.Text("Button3")

    // 屏幕宽度只有 20，应该自动换行
    wrapped := WrapBuilder(btn1, btn2, btn3).
        Gap(1).
        ScreenWidth(20).
        Build()

    // 验证返回的是 VStack
    // 验证有多个子元素（多行）
}

func TestWrap_NoWrappingNeeded(t *testing.T) {
    items := []ui.VNode{
        ui.Text("A"),
        ui.Text("B"),
        ui.Text("C"),
    }

    // 屏幕宽度 100，应该不需要换行
    wrapped := WrapBuilder(items...).
        Gap(1).
        ScreenWidth(100).
        Build()

    // 验证只有一个 HStack 子元素
}

func TestWrap_EmptyChildren(t *testing.T) {
    wrapped := WrapBuilder().
        Gap(1).
        ScreenWidth(80).
        Build()

    // 应该返回空的 VStack
}

func TestWrap_AlignCenter(t *testing.T) {
    items := createButtons(5)

    wrapped := WrapBuilder(items...).
        Gap(1).
        Align(ui.AlignCenter).
        ScreenWidth(40).
        Build()

    // 验证每个 HStack 使用 AlignCenter
}
```

### 集成测试

**文件**: `examples/wrap_demo/main.go`

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/app"
    ui "github.com/wwsheng009/mint/ui"
)

func main() {
    err := ui.Run(WrapDemo,
        ui.WithWidth(80),
        ui.WithHeight(20),
        ui.WithTitle("Wrap Component Demo"),
    )
    if err != nil {
        panic(err)
    }
}

func WrapDemo() ui.VNode {
    // 创建 10 个按钮演示自动换行
    var buttons []ui.VNode
    for i := 1; i <= 10; i++ {
        label := fmt.Sprintf("[%d]", i)
        buttons = append(buttons,
            app.ButtonBuilder(label).
                FocusStyle(app.FocusStyleBracket).
                Build(),
        )
    }

    return ui.VStack(
        ui.Text("Wrap Component Demo"),
        ui.Text("─────────────────"),
        app.WrapBuilder(buttons...).
            Gap(1).
            RowGap(0).
            ScreenWidth(76). // 80 - border
            Align(ui.AlignStart).
            Build(),
    )
}
```

---

## 关键文件引用

### 实现参考

1. **`runtime/ui/layout.go:36-228`** - LayoutNode 和 LayoutBuilder 模式
2. **`runtime/ui/layout.go:570-624`** - estimateChildWidth() 实现
3. **`runtime/compute/engine.go:886-1066`** - HStack/VStack 布局逻辑
4. **`components/layout/stack.go`** - 现有布局组件实现参考
5. **`components/button/button.go:456-508`** - Button.Measure() 实现参考

### 测试参考

1. **`components/layout/stack_test.go`** - 布局组件测试模式
2. **`examples/ui_demos/demo1_full_featured/main.go`** - 实际应用示例

---

## 成功标准

实现完成的标准：

1. ✅ **API 一致性** - 与 HStack/VStack 使用相同的 Builder 模式
2. ✅ **自动换行** - 根据配置的宽度自动分行
3. ✅ **间距控制** - Gap 和 RowGap 正确应用
4. ✅ **对齐支持** - Align 参数正确传递到每行的 HStack
5. ✅ **宽度估算** - 准确估算常见组件（Button, Text, Input）的宽度
6. ✅ **性能可接受** - 没有明显的性能退化
7. ✅ **测试覆盖** - 单元测试和集成测试通过
8. ✅ **文档完整** - 使用文档和 API 参考完整
9. ✅ **demo2 集成** - 可以成功应用到 demo2 的 ControlPanel

---

## 实现优先级

### Phase 1: 核心功能（必须）

1. WrapNode 和 WrapBuilder 基础结构
2. estimateWidth() 基本实现
3. calculateRows() 核心算法
4. Build() 转换逻辑
5. 基本测试

### Phase 2: 集成（重要）

1. app 包和 ui 包导出
2. demo2 ControlPanel 应用
3. 完善测试用例
4. 基础文档

### Phase 3: 完善（可选）

1. 宽度缓存优化
2. 支持更多组件类型
3. 高级测试（性能、边缘案例）
4. 详细文档和示例

---

## 预期效果

实现后，demo2 的 ControlPanel 将变为：

```go
func ControlPanel(...) ui.VNode {
    allButtons := []ui.VNode{
        app.ButtonBuilder("[1] Event").Build(),
        app.ButtonBuilder("[2] setState").Build(),
        app.ButtonBuilder("[3] Scheduler").Build(),
        app.ButtonBuilder("[4] Render").Build(),
        app.ButtonBuilder("[5] Reconcile").Build(),
        app.ButtonBuilder("[6] Layout").Build(),
        app.ButtonBuilder("[7] Paint").Build(),
        app.ButtonBuilder("[0] Idle").Build(),
    }

    return ui.Bordered().
        Style(string(theme.Border())).
        Child(app.WrapBuilder(allButtons...).
            Gap(1).
            RowGap(0).
            ScreenWidth(98). // 100 - border (2)
            Align(ui.AlignStart).
            Build()).
        FillWidth().
        Build()
}
```

**效果**：
- ✅ 在宽屏（≥120 字符）：所有按钮在一行
- ✅ 在标准屏（80-100 字符）：自动换行到 2 行
- ✅ 在窄屏（<80 字符）：自动换行到多行
- ✅ 无需手动分行，自动适配不同终端宽度
