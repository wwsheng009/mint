基于你提供的 `yao/tui/framework` 源代码（特别是 `component.go`, `base.go`, `textinput.go`, `theme.go`），我们可以看出 Yao TUI 的架构是基于 **保留模式（Retained Mode）** 和 **直接绘制（Painter/Buffer）** 的混合架构。

它不依赖 Bubble Tea 的 Model/View/Update 循环来生成字符串，而是直接操作 `paint.Buffer`。因此，设计 UI 组件库时，核心在于 **状态管理**、**绘制管线** 和 **事件分发**。

以下是针对 Yao TUI 架构的 UI 组件设计与实施指南（去 Bubbles 化，纯原生实现）：

---

### 1. 核心架构原则 (基于现有代码)

在设计任何新组件之前，必须遵循现有的接口契约：

* **StateHolder (状态)**: 组件必须维护自己的线程安全状态 (`sync.RWMutex`)。
* **Measurable (测量)**: 父组件询问子组件需要多大空间 (`Measure`)。
* **Paintable (绘制)**: 组件将自己绘制到 `paint.Buffer` 上的 `(x, y)` 坐标。
* **ActionTarget (交互)**: 通过 `HandleAction` 处理语义化事件（如 `ActionInputChar`），而不是处理原始按键。

---

### 2. 主题系统设计实施 (Theme Implementation)

你已经有了 `theme.go`，设计组件时应当通过 **Token** 而非硬编码颜色来引用样式。

**设计策略：**

1. **定义语义化 Token**: 在 `Theme` 结构体中，除了基础颜色（Primary, Secondary），还需要定义组件级 Token。
* `Input.Background`, `Input.Border`, `Input.Placeholder`
* `List.Item.Active`, `List.Item.Normal`


2. **样式解析器**: 组件初始化或绘制时，从全局 Theme 获取 `StyleConfig`。

**代码示例 (组件中应用主题):**

```go
// 在组件绘制方法中
func (c *MyComponent) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 从上下文中获取主题 (假设 ctx 包含 Theme 引用，或者通过全局 ThemeManager 获取)
    t := theme.GetGlobal() 
    
    // 获取当前状态的样式 (Normal, Focus, Disabled)
    state := "normal"
    if c.IsDisabled() {
        state = "disabled"
    } else if c.IsFocused() {
        state = "focus"
    }
    
    // 获取计算后的样式
    styleCfg := t.GetComponentStyle(c.Type(), state)
    
    // 将 StyleConfig 转换为 paint.Style 并应用
    // ...绘制逻辑
}

```

---

### 3. 基础输入组件设计指导 (Input & TextArea)

你提供的 `TextInput` 实现是一个很好的起点。为了使其成为生产级组件，需要增加 **视口滚动 (Viewport)** 和 **掩码** 功能。

#### A. 增强型 TextInput (单行输入)

**设计痛点**: 当文本长度超过组件宽度时，当前实现可能会截断。
**解决方案**: 引入 `viewOffset`。

1. **状态扩展**:
```go
type TextInput struct {
    // ... 原有字段
    viewOffset int // 可视区域的起始字符索引
}

```


2. **绘制逻辑 (Paint)**:
* 计算 `cursor` 相对于 `viewOffset` 的位置。
* 如果光标移出右边界，增加 `viewOffset`。
* 如果光标移出左边界，减少 `viewOffset`。
* 绘制字符串时，截取 `value[viewOffset : viewOffset+width]`。



#### B. TextArea (多行输入)

**设计方案**:

* **数据结构**: `lines []string` 或直接使用 `string` 并动态计算换行。
* **自动换行 (Soft Wrap)**: 在 `Measure` 阶段或 `Paint` 阶段，根据宽度计算每一行显示的文本。
* **垂直滚动**: 需要 `yOffset`。
* **光标移动**: 上下移动光标时，需要计算目标行的字符位置（保持视觉上的 X 轴位置，即 "Desired Column"）。

---

### 4. Form (表单) 组件设计方案

Form 是一个容器组件，负责管理子组件的布局和焦点流转。

**核心职责**:

1. **焦点管理 (Focus Trap)**: 拦截 Tab / Shift+Tab 键。
2. **数据聚合**: 提供 `GetData()` 方法返回所有子组件的值。
3. **验证协调**: 调用所有子组件的 `Validate()`。

**实现逻辑**:

```go
type Form struct {
    *component.BaseComponent
    children []component.Component // 包含 Input, Select, DatePicker 等
    focusIdx int                   // 当前焦点在 children 中的索引
}

// HandleAction 拦截导航键
func (f *Form) HandleAction(a action.Action) bool {
    // 1. 优先让当前有焦点的子组件处理事件
    current := f.children[f.focusIdx]
    if target, ok := current.(component.ActionTarget); ok {
        if target.HandleAction(a) {
            return true // 子组件消费了事件
        }
    }

    // 2. 子组件没处理（例如 Input到了行尾，或者按了 Tab），Form 接管
    switch a.Type {
    case action.ActionNextFocus: // Tab
        f.cycleFocus(1)
        return true
    case action.ActionPrevFocus: // Shift+Tab
        f.cycleFocus(-1)
        return true
    case action.ActionSubmit:    // Enter (如果在按钮上)
        f.Submit()
        return true
    }
    return false
}

func (f *Form) cycleFocus(direction int) {
    // 1. Blur 当前组件
    f.children[f.focusIdx].OnBlur()
    
    // 2. 循环计算下一个索引
    // ... logic to find next focusable item ...
    
    // 3. Focus 新组件
    f.children[f.focusIdx].OnFocus()
}

```

---

### 5. 列表与表格组件 (List & Table)

这两个组件最复杂的地方在于 **渲染性能**。必须实现 **虚拟滚动 (Virtualization)**，只渲染可视区域的行。

#### A. Table (表格)

**设计要素**:

* **Columns**: 定义列宽（固定宽度、百分比或自适应）。
* **Rows**: 数据源。
* **Offset**: 当前滚动到的行索引。
* **Selection**: 当前高亮的行。

**绘制流程**:

1. 绘制 Header（固定在顶部）。
2. 计算可视区域高度 `H`。
3. 从数据源的 `Offset` 索引开始，循环绘制 `H` 行数据。
4. 如果 `rowIndex == selectedIndex`，应用 `Theme.Table.Row.Active` 样式（通常是反色背景）。

#### B. List (列表)

与 Table 类似，但每个 Item 的渲染逻辑更灵活。可以引入 **ItemRenderer** 接口。

```go
type ListItem interface {
    Title() string
    Description() string
}

type List struct {
    items []ListItem
    // ...
}

// 绘制时，可以支持简单的两行模式（标题+描述）
func (l *List) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 虚拟滚动逻辑...
    for i := 0; i < visibleCount; i++ {
        item := l.items[l.offset + i]
        // 绘制 Title
        // 绘制 Description
    }
}

```

---

### 6. 弹窗类组件 (Select / DatePicker / Modal)

在 TUI 中实现弹窗（Overlay）需要架构支持 **Z-Index** 或 **Layers（图层）** 的概念。由于 `paint.Buffer` 是平面的，通常通过渲染顺序来控制。

#### DatePicker (日期选择器)

这是最复杂的交互组件之一。

**设计方案**:

1. **Composition**: 由一个 `TextInput` (显示日期) 和一个 `CalendarPopup` (弹出层) 组成。
2. **Calendar 逻辑**:
* **State**: `CurrentMonth`, `CurrentYear`, `SelectedDate`.
* **Grid Calculation**: 一个 7x6 的网格。
* 算法：计算当月1号是周几，确定日历起始格。


* **Navigation**: 左右方向键在天之间移动，上下键在周之间移动。PageUp/Down 切换月份。


3. **弹出机制**:
* 当 `TextInput` 获得焦点并按下 Enter/Space 时，设置 `ShowCalendar = true`。
* 在 `Paint` 阶段，如果是 `ShowCalendar` 状态，**在所有其他组件绘制完成后**，在指定坐标（Input 下方）绘制日历矩形区域。*注意：这需要 Paint 系统支持后处理或图层，或者在父容器的 Paint 最后手动调用。*



---

### 7. 实施路线图

建议按以下顺序开发：

1. **完善基础设施**:
* 确保 `paint.Buffer` 支持中文宽字符（wcwidth）。
* 完善 `Theme` 加载机制。


2. **Level 1 组件 (无子级)**:
* `Button`, `Label`, `ProgressBar`, `Separator`.


3. **Level 2 组件 (交互核心)**:
* 完善 `TextInput` (增加滚动)。
* `Checkbox`, `Radio`.


4. **Level 3 组件 (布局与管理)**:
* `Form` (焦点管理)。
* `Modal` / `Dialog` (居中覆盖层)。


5. **Level 4 组件 (复杂数据)**:
* `Table` (虚拟滚动)。
* `Select` / `Combobox` (输入+列表的组合)。



### 8. 关键代码技巧

* **双缓冲 (Double Buffering)**: 你的架构似乎是直接写入 Buffer。为了防止闪烁，确保整个 Buffer 绘制完成后再一次性 Flush 到终端。
* **脏矩形 (Dirty Rects)**: 为了性能，`StateHolder` 应该记录组件是否需要重绘。如果组件状态未变，且布局未变，`Paint` 方法可以直接跳过（或者复用缓存的 Buffer，如果实现了离屏 Buffer）。
* **坐标系转换**: 始终清晰区分 **绝对坐标** (屏幕坐标) 和 **相对坐标** (容器内坐标)。你的 `BaseComponent` 存储的是相对父容器的 `x, y`，但在 `Paint(ctx)` 时，`ctx.X, ctx.Y` 通常是绝对坐标。

通过遵循这些设计模式，你可以构建一套高性能、风格统一且不依赖第三方庞大库的 Go TUI 组件库。

好的，我们继续深入 `yao/tui/framework` 架构下的组件开发指导。上一部分我们讨论了核心架构、基础输入框和表单容器。

这一部分我们将重点放在更复杂的 **交互组件**（Select/下拉框、Checkbox/复选框）、**布局系统** 的集成，以及 **浮层（Overlay）** 的管理策略。

---

### 9. Select / Dropdown (下拉选择组件) 设计

下拉框是 TUI 中最复杂的组件之一，因为它涉及“弹出层”的概念。在终端中，我们没有原生的 Z-轴，必须通过绘制顺序或独立的 Overlay 层来实现。

#### 设计方案

1. **状态管理 (`SelectState`)**:
* `Options`: 选项列表 `[{Label, Value}]`。
* `SelectedIndex`: 当前选中的索引。
* `IsOpen`: 是否展开。
* `ListState`: 嵌入一个 `List` 组件的状态，用于管理下拉菜单的滚动。


2. **交互逻辑**:
* **收起状态**: 显示当前选中的值。按下 `Enter` 或 `Space` -> 设置 `IsOpen = true`，请求重绘。
* **展开状态**:
* 拦截上下键 -> 改变高亮选项。
* 拦截 `Enter` -> 选中当前项，`IsOpen = false`。
* 拦截 `Esc` -> `IsOpen = false` (不改变选中项)。




3. **绘制策略 (关键)**:
* 通常 TUI 框架是从上到下、从父到子绘制的。如果在 `Paint` 方法中直接绘制下拉列表，它可能会被后续绘制的组件（如下一行的输入框）覆盖。
* **解决方案**: 使用 **Deferred Painting (延迟绘制)** 或 **Overlay Manager**。组件在 `Paint` 阶段只绘制输入框本身，并将“绘制下拉菜单”的任务注册到全局 Overlay 队列中，等待所有基础层绘制完毕后再统一绘制 Overlay。



#### 代码实现 (Select)

```go
type SelectOption struct {
    Label string
    Value string
}

type Select struct {
    *component.BaseComponent
    options     []SelectOption
    selectedIdx int
    isOpen      bool
    listView    *List // 复用 List 组件来渲染下拉部分
}

func NewSelect(options []SelectOption) *Select {
    return &Select{
        BaseComponent: component.NewBase(),
        options:       options,
        listView:      NewList(), // 简化的内部 List
    }
}

// HandleAction 处理交互
func (s *Select) HandleAction(a action.Action) bool {
    if !s.IsFocused() {
        return false
    }

    if !s.isOpen {
        // 收起状态：按 Enter/Space 展开
        if a.Type == action.ActionSubmit || a.Key == action.KeySpace {
            s.isOpen = true
            return true
        }
        return false // 让 Form 处理 Tab 等导航
    }

    // 展开状态：接管所有导航键
    switch a.Type {
    case action.ActionUp, action.ActionDown:
        // 委托给内部的 List 处理滚动
        return s.listView.HandleAction(a)
    case action.ActionSubmit:
        // 确认选择
        s.selectedIdx = s.listView.GetSelectedIndex()
        s.isOpen = false
        return true
    case action.ActionCancel: // Esc
        s.isOpen = false
        return true
    }
    return true // 展开时吞掉其他按键
}

// Paint 绘制
func (s *Select) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 1. 绘制主输入框区域 (模拟 TextInput 样式)
    style := theme.GetGlobal().GetStyle("input.normal")
    if s.IsFocused() {
        style = theme.GetGlobal().GetStyle("input.focus")
    }
    
    text := ""
    if s.selectedIdx >= 0 && s.selectedIdx < len(s.options) {
        text = s.options[s.selectedIdx].Label
    }
    // 绘制边框和文字...
    // 绘制右侧箭头 "▼"
    
    // 2. 如果展开，注册 Overlay
    if s.isOpen {
        // 计算下拉框的绝对坐标
        absX, absY := s.GetAbsolutePosition() 
        dropdownHeight := 5 // 假设最大高度
        
        // 向上下文注册一个 Overlay 绘制任务
        // 注意：ctx 需要有 AddOverlay 能力
        if overlayMgr, ok := ctx.OverlayManager(); ok {
            overlayMgr.Add(func(topBuf *paint.Buffer) {
                // 在最顶层绘制下拉列表
                // 绘制背景框 (覆盖下面的内容)
                // 绘制 List 内容
            })
        }
    }
}

```

---

### 10. Checkbox & Radio (选择控件)

这两个组件逻辑简单，核心在于符号的渲染。

#### Checkbox (复选框)

* **外观**: `[ ]` (未选中), `[x]` (选中), `[-]` (半选/Indeterminate).
* **交互**: Space 键切换状态。

```go
func (c *Checkbox) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    icon := "[ ]"
    if c.checked {
        icon = "[x]" // 或者使用 Nerd Fonts 图标
    }
    // 绘制图标 + Label
    buf.SetString(ctx.X, ctx.Y, icon+" "+c.label, c.GetStyle())
}

```

#### RadioGroup (单选组)

* **布局**: 通常包含多个子项，可以是水平排列或垂直排列。
* **导航**:
* 如果 RadioGroup 作为一个整体获得焦点，使用 左/右 或 上/下 键在内部选项间切换 `SelectedValue`。
* 需要处理 `Focus` 状态：整个组有焦点时，当前选中的项显示高亮。



---

### 11. Modal / Dialog (模态对话框)

模态框是应用级别的组件，它会打断正常的用户流。

#### 核心设计

1. **全局单例或栈**: 通常 App 中同一时间只显示一个 Modal（或者通过栈管理多层 Modal）。
2. **Focus Trap (焦点陷阱)**: 当 Modal 显示时，**必须** 阻止 Tab 键切换到 Modal 之外的组件（如底层的 Form）。
3. **渲染**: 居中绘制，通常需要计算 `(ScreenWidth - ModalWidth) / 2`。

#### 实施指导

在 `App` 或 `Root` 组件的 `Paint` 方法最后调用：

```go
func (app *App) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 1. 绘制主界面
    app.rootComponent.Paint(ctx, buf)

    // 2. 绘制 Modal (如果存在)
    if app.currentModal != nil {
        // 绘制半透明遮罩 (Dim Background)
        // 这是一个技巧：遍历 Buffer 所有单元格，将背景色变暗
        app.dimBackground(buf)

        // 计算居中位置
        w, h := app.currentModal.Measure(app.width, app.height)
        x := (app.width - w) / 2
        y := (app.height - h) / 2
        
        // 绘制 Modal
        // 这里需要创建一个偏移后的 Context
        modalCtx := ctx.WithOffset(x, y)
        app.currentModal.Paint(modalCtx, buf)
    }
}

// 焦点分发
func (app *App) HandleAction(a action.Action) {
    if app.currentModal != nil {
        // 如果有 Modal，事件只分发给 Modal
        app.currentModal.HandleAction(a)
        return
    }
    // 否则分发给主界面
    app.rootComponent.HandleAction(a)
}

```

---

### 12. 布局引擎 (Layout Engine)

不要手动计算每个组件的 `x, y`，应该实现一套简单的布局容器。

#### FlexBox (弹性盒子)

参考 CSS Flexbox 或 Flutter 的 Row/Column。

**结构**:

```go
type Flex struct {
    Direction Axis // Horizontal | Vertical
    Children  []FlexChild
}

type FlexChild struct {
    Component component.Component
    Fixed     int // 固定像素
    Proportion int // 比例 (flex-grow)
}

```

**Measure 算法 (两遍扫描)**:

1. **Pass 1**: 遍历所有 `Fixed` 大小的子组件，减去它们占用的空间。
2. **Pass 2**: 剩余空间按 `Proportion` 分配给弹性子组件，调用它们的 `Measure`。

**Layout/Paint 算法**:

* 记录当前的 `cursorX, cursorY`。
* 遍历子组件，将计算好的位置传递给子组件的 `Paint` 方法。

---

### 13. 组件开发 Checklist

在实现每个组件时，请检查以下几点：

1. **尺寸自适应**:
* 组件是否尊重父容器传入的 `maxWidth/maxHeight`？
* 如果空间不足，是截断还是换行？


2. **Unicode 支持**:
* **必须** 使用 `runewidth.StringWidth()` 而不是 `len()` 来计算字符串长度。
* 处理中文输入时的光标位置（一个中文字符占2个单元格，光标通常只移动1次逻辑位置，但视觉上要移动2格）。


3. **样式统一**:
* 是否使用了 `Theme` 中的 Token？
* Focus 状态和 Error 状态是否有明显的视觉区分？


4. **性能优化**:
* `Paint` 方法会被高频调用（每次按键可能都会触发）。避免在 `Paint` 中进行耗时的计算（如复杂的布局计算），布局结果应该缓存，只有当尺寸变化时才重新计算（Dirty Flag 模式）。



### 14. 推荐的目录结构

为了保持项目整洁，建议如下组织代码：

```text
yao/tui/framework/
├── widget/           # 具体组件实现
│   ├── button/
│   ├── input/        # TextInput, TextArea
│   ├── select/       # Dropdown
│   ├── checkbox/
│   ├── table/
│   └── list/
├── layout/           # 布局容器
│   ├── flex.go
│   ├── grid.go
│   └── container.go  # Padding, Margin, Border wrapper
├── overlay/          # 浮层管理
│   └── manager.go
├── theme/            # 主题定义
│   ├── default.go
│   └── types.go
├── event/            # 事件定义
└── backend/          # 终端交互层 (tcell/termbox 适配)

```

通过这套方案，你可以构建出一个健壮、可扩展且去 Bubbles 依赖的 TUI 组件库。

继续我们在 Yao TUI 架构下的组件开发指导。在前两部分，我们讨论了基础架构、输入类组件、弹窗和浮层管理。

现在，我们将深入探讨 **复杂数据展示组件（Table & Tree）** 以及 **布局系统（FlexBox）** 的具体实现细节。这是构建复杂 TUI 界面（如仪表盘、文件管理器）的基石。

---

### 15. 高级数据组件：Table (表格)

表格是 TUI 应用中最核心的组件之一。在 `paint.Buffer` 架构下，为了保证性能，必须实现 **虚拟滚动 (Virtual Scrolling)**。

#### 设计要素

* **Model**: 数据模型，通常是一个二维数组或对象切片。
* **Columns**: 列定义（标题、宽度、对齐方式、渲染器）。
* **Viewport**: 视口，包含 `OffsetX` (水平滚动) 和 `OffsetY` (垂直滚动)。
* **Selection**: 行选择模式（单选/多选）或单元格选择模式。

#### 核心实现 (Virtual Table)

```go
type Column struct {
    Title    string
    Width    int     // 0 表示自适应权重，>0 表示固定宽度
    Align    string  // "left", "center", "right"
    RenderFn func(data interface{}) string // 自定义渲染
}

type Table struct {
    *component.BaseComponent
    columns     []Column
    data        []interface{} // 原始数据
    offsetY     int           // 当前滚动起始行
    selectedRow int           // 当前选中行
    rowHeight   int           // 通常为 1
}

// Measure 计算表格尺寸
// 表格通常贪婪地占用剩余空间，但在 Flex 布局中可能受限
func (t *Table) Measure(maxWidth, maxHeight int) (int, int) {
    // 简单实现：尝试填满最大空间
    return maxWidth, maxHeight
}

// Paint 绘制逻辑 (关键：只绘制可见部分)
func (t *Table) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    if !t.IsVisible() || ctx.AvailableHeight < 2 { // 至少要有 Header + 1 Row
        return
    }

    // 1. 计算列宽
    // 实际实现中应该缓存列宽计算结果，避免每次 Paint 都计算
    colWidths := t.calculateColumnWidths(ctx.AvailableWidth)

    // 2. 绘制表头 (Header)
    t.paintHeader(ctx, buf, colWidths)

    // 3. 绘制数据行 (Body)
    // 可用高度减去表头高度
    bodyHeight := ctx.AvailableHeight - 1 
    if bodyHeight <= 0 {
        return
    }

    // 计算可视范围
    startRow := t.offsetY
    endRow := startRow + bodyHeight
    if endRow > len(t.data) {
        endRow = len(t.data)
    }

    y := ctx.Y + 1 // Header 占用一行
    for i := startRow; i < endRow; i++ {
        rowState := "normal"
        if i == t.selectedRow {
            rowState = "active" // 选中高亮
        }
        
        t.paintRow(ctx, buf, y, t.data[i], colWidths, rowState)
        y++
    }
    
    // 4. 绘制滚动条 (可选)
    t.paintScrollBar(ctx, buf, bodyHeight, len(t.data))
}

// HandleAction 处理滚动和选择
func (t *Table) HandleAction(a action.Action) bool {
    if !t.IsFocused() {
        return false
    }

    switch a.Type {
    case action.ActionUp:
        if t.selectedRow > 0 {
            t.selectedRow--
            // 自动跟随滚动
            if t.selectedRow < t.offsetY {
                t.offsetY = t.selectedRow
            }
            return true
        }
    case action.ActionDown:
        if t.selectedRow < len(t.data)-1 {
            t.selectedRow++
            // 自动跟随滚动
            // 注意：需要知道当前的 bodyHeight，这里简化处理
            if t.selectedRow >= t.offsetY + 10 { // 假设高度10
                t.offsetY++
            }
            return true
        }
    // ... PageUp, PageDown 处理
    }
    return false
}

```

---

### 16. 高级布局系统：FlexBox 实现

你不能依赖绝对坐标来构建复杂的 UI。你需要一个布局容器，它接受子组件，并根据规则自动计算子组件的 `x, y, width, height`。

#### Flex 容器设计

```go
type FlexDirection int
const (
    FlexRow FlexDirection = iota
    FlexColumn
)

type FlexItem struct {
    Component component.Component
    Fixed     int // 固定尺寸 (px/cells)
    Grow      int // 伸缩权重 (0不伸缩)
}

type Flex struct {
    *component.BaseComponent
    direction FlexDirection
    items     []*FlexItem
    gap       int // 子项间距
}

// Measure 阶段：确定 Flex 容器自身的大小（通常由父容器决定）
// 并协调子组件的 Measure
func (f *Flex) Measure(maxWidth, maxHeight int) (int, int) {
    // Flex 容器通常填充父容器给予的空间
    f.SetSize(maxWidth, maxHeight) 
    
    // 这里的关键是：我们不在这里 Measure 子组件
    // 而是在 Layout/Paint 阶段，或者专门的 Layout 方法中
    // 根据 f.width / f.height 去分配空间给子组件
    return maxWidth, maxHeight
}

// Layout 阶段：核心算法
// 这个方法可以在 Paint 之前调用，或者在 Paint 的开始调用（如果有缓存机制）
func (f *Flex) layoutChildren() {
    width, height := f.GetSize()
    
    // 1. 计算主轴可用空间 (Main Axis Space)
    totalSpace := width
    if f.direction == FlexColumn {
        totalSpace = height
    }
    
    // 减去间距
    if len(f.items) > 1 {
        totalSpace -= (len(f.items) - 1) * f.gap
    }

    // 2. 扣除固定尺寸 (Fixed)
    availableSpace := totalSpace
    totalGrow := 0
    
    for _, item := range f.items {
        if item.Fixed > 0 {
            availableSpace -= item.Fixed
        }
        totalGrow += item.Grow
    }
    
    if availableSpace < 0 {
        availableSpace = 0
    }

    // 3. 分配坐标和尺寸
    currentPos := 0
    
    for _, item := range f.items {
        // 计算该项在主轴上的尺寸
        itemSize := item.Fixed
        if item.Grow > 0 && totalGrow > 0 {
            // 按权重分配剩余空间
            share := (item.Grow * availableSpace) / totalGrow
            itemSize += share
            // 处理整数除法余数分配问题（略）
        }
        
        // 计算交叉轴尺寸 (Cross Axis)
        crossSize := height
        if f.direction == FlexColumn {
            crossSize = width
        }
        
        // 设置子组件位置和尺寸
        // 注意：这里设置的是相对坐标
        if f.direction == FlexRow {
            item.Component.SetPosition(currentPos, 0)
            item.Component.SetSize(itemSize, crossSize)
            
            // 重要：通知子组件重新测量内容
            // 很多组件（如 Text）需要根据宽度决定高度（换行）
            // 如果是 Row 布局，高度已定；如果是 Column，宽度已定
            if m, ok := item.Component.(component.Measurable); ok {
                m.Measure(itemSize, crossSize)
            }
            
            currentPos += itemSize + f.gap
        } else {
            // Column Logic...
            item.Component.SetPosition(0, currentPos)
            item.Component.SetSize(crossSize, itemSize)
            
            if m, ok := item.Component.(component.Measurable); ok {
                m.Measure(crossSize, itemSize)
            }
            
            currentPos += itemSize + f.gap
        }
    }
}

func (f *Flex) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 确保布局已更新
    f.layoutChildren()
    
    // 递归绘制子组件
    for _, item := range f.items {
        child := item.Component
        x, y := child.GetPosition()
        w, h := child.GetSize()
        
        // 创建子 Context，坐标系转换为绝对坐标
        subCtx := ctx
        subCtx.X += x
        subCtx.Y += y
        subCtx.AvailableWidth = w
        subCtx.AvailableHeight = h
        
        // 剪裁区域 (Clipping) - 防止子组件画出界
        subCtx.ClipRect = ctx.ClipRect.Intersect(&component.Rect{
            X: subCtx.X, Y: subCtx.Y, Width: w, Height: h,
        })
        
        child.Paint(subCtx, buf)
    }
}

```

---

### 17. 树形组件：TreeView (文件树/导航)

TreeView 适合展示层级结构。

#### 设计方案

1. **Node 结构**:
```go
type TreeNode struct {
    ID       string
    Label    string
    Icon     string // e.g. "📁", "📄"
    Level    int    // 缩进层级
    Expanded bool
    Children []*TreeNode
    Parent   *TreeNode
}

```


2. **扁平化渲染 (Flattening)**:
为了简化渲染和键盘导航（上下键），不要递归渲染。而是将树 **扁平化** 为一个可视列表。
* 遍历树，如果节点是 `Expanded`，将其子节点加入列表。
* 如果节点是 `Collapsed`，跳过其子节点。
* 结果是一个 `[]*TreeNode` 切片，可以直接复用 `List` 组件的逻辑来渲染。


3. **交互**:
* `Enter`: 切换 `Expanded` 状态（如果是文件夹），或触发选中（如果是文件）。
* `Right`: 展开。
* `Left`: 折叠。如果已折叠，跳转到父节点。



---

### 18. 事件与焦点管理系统 (Event & Focus System)

为了让组件协同工作，你需要一个健壮的事件分发系统。

#### Focus Chain (焦点链)

不要让每个组件自己去抢焦点。应由 **Container** 管理。

```go
type FocusManager struct {
    focusables []component.Focusable // 扁平化的可聚焦组件列表
    currentIdx int
}

func (fm *FocusManager) Next() {
    // Blur current
    // Index++ (loop)
    // Focus next
}

```

#### 事件冒泡 (Event Bubbling)

在 `HandleAction` 中，如果组件处理不了事件，是否应该冒泡给父组件？

* **Yes**: 例如在 `Input` 中按 `Tab`，输入框不处理，冒泡给 `Form`，`Form` 切换焦点。
* **实现**: `BaseComponent` 可以持有一个 `Parent` 引用。

```go
// 在 App 层面分发事件
func (app *App) DispatchAction(a action.Action) {
    // 1. 获取当前焦点组件
    target := app.FocusManager.Current()
    
    // 2. 尝试让目标处理
    if target.HandleAction(a) {
        return
    }
    
    // 3. 冒泡 (Bubble Up)
    parent := target.GetParent()
    for parent != nil {
        if pAction, ok := parent.(component.ActionTarget); ok {
            if pAction.HandleAction(a) {
                return
            }
        }
        parent = parent.GetParent()
    }
    
    // 4. 全局快捷键处理 (Global Shortcuts)
    if a.Key == action.KeyCtrlC {
        app.Quit()
    }
}

```

---

### 19. 最终集成：App Loop (应用主循环)

将所有部分组合在一起。

```go
func (app *App) Run() error {
    // 1. 初始化终端 (tcell/termbox)
    screen := app.initScreen()
    defer screen.Fini()

    // 2. 事件通道
    events := make(chan event.Event)
    go app.pollEvents(screen, events)

    // 3. 主循环
    for {
        // A. 渲染阶段
        app.layout() // 计算布局
        app.paint()  // 绘制到 Buffer
        app.flush()  // 将 Buffer 同步到 Screen

        // B. 事件处理阶段
        select {
        case ev := <-events:
            switch e := ev.(type) {
            case *event.ResizeEvent:
                app.width, app.height = e.Width, e.Height
                app.root.Measure(app.width, app.height) // 触发重新测量
            case *event.KeyEvent:
                // 转换为语义化 Action
                act := action.MapKeyToAction(e)
                app.DispatchAction(act)
            }
        case <-app.quitChan:
            return nil
        }
    }
}

```

### 总结

这套设计方案（Level 1 基础 -> Level 2 复合 -> Layout -> System）完整覆盖了从底层绘制到上层交互的所有环节。

**关键建议**:

1. **优先实现 Flex 布局**: 没有布局系统，手写坐标会极其痛苦且难以维护。
2. **Table 是重中之重**: 几乎所有业务系统都需要 Table。
3. **保持 State 和 UI 分离**: 就像在 React 中一样，修改 State，然后请求重绘 (MarkDirty)，不要直接修改 Buffer。

遵循这些指导，你将能够为 Yao 项目开发出一套专业、高性能且架构清晰的原生 TUI 组件库。

继续深入 Yao TUI 架构的组件开发设计。在掌握了基础交互、布局和数据列表之后，我们需要关注提升应用“专业度”和“即时响应能力”的高级特性。

这部分将涵盖 **Canvas 绘图与图表**、**异步并发模式**、**命令面板（Command Palette）** 以及 **Grid 响应式布局**。

---

### 20. 高级视觉组件：Canvas 与图表 (Charts)

在终端中，字符单元格（Cell）通常是高度大于宽度的（约 2:1）。为了绘制精细的图表（折线图、波形图），我们需要突破单元格限制，使用 **Braille (盲文)** 字符或 **Block (方块)** 字符来模拟“子像素”渲染。

#### A. Braille Canvas 设计

Braille 模式允许在一个字符单元格中通过 2x4 的点阵来绘制，将分辨率提高 8 倍。

```go
// Canvas 是一个低级组件，提供像素级操作接口
type Canvas struct {
    *component.BaseComponent
    width, height int // 逻辑像素尺寸 (Canvas Width = Cell Width * 2, Height = Cell Height * 4)
    buffer        []rune // 存储盲文字符
    colors        []theme.Color
}

// SetPixel 设置逻辑像素点
func (c *Canvas) SetPixel(x, y int, color theme.Color) {
    // 1. 计算所属的 Cell 坐标
    cellX := x / 2
    cellY := y / 4
    
    // 2. 计算在该 Cell 内部的偏移 (0-1, 0-3)
    subX := x % 2
    subY := y % 4
    
    // 3. 映射到 Braille Unicode 偏移量 (0x2800 base)
    // 盲文点位映射表:
    // (0,0)->0x1, (0,1)->0x2, (0,2)->0x4, (0,3)->0x40
    // (1,0)->0x8, (1,1)->0x10, (1,2)->0x20, (1,3)->0x80
    mask := c.computeBrailleMask(subX, subY)
    
    // 4. 更新 Buffer
    c.buffer[cellY*c.width + cellX] |= mask
}

func (c *Canvas) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 将内部 Braille buffer 绘制到 TUI buffer
    for y := 0; y < c.cellHeight; y++ {
        for x := 0; x < c.cellWidth; x++ {
            char := c.buffer[y*c.cellWidth + x]
            buf.SetCell(ctx.X+x, ctx.Y+y, char + 0x2800, c.GetStyle())
        }
    }
}

```

#### B. BarChart (柱状图)

使用 Block 字符 (`  ▂ ▃ ▄ ▅ ▆ ▇ █`) 实现。

```go
type BarChart struct {
    *component.BaseComponent
    data   []float64
    labels []string
    maxVal float64
}

func (b *BarChart) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 1. 确定比例尺
    scale := float64(ctx.AvailableHeight) / b.maxVal
    blocks := []rune{' ', ' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

    barWidth := 2 // 每个柱子占2列
    gap := 1      // 间距

    x := ctx.X
    for i, val := range b.data {
        height := int(val * scale)
        
        // 绘制柱子 (从下往上)
        for h := 0; h < height; h++ {
            // 计算当前格子的填充度 (如果是最顶端，可能不需要全满)
            char := '█' 
            // ... 精细计算逻辑 ...
            
            drawY := ctx.Y + ctx.AvailableHeight - 1 - h
            buf.SetCell(x, drawY, char, theme.GetColor("primary"))
            buf.SetCell(x+1, drawY, char, theme.GetColor("primary")) // 加宽
        }
        
        // 绘制 Label
        buf.SetString(x, ctx.Y+ctx.AvailableHeight, b.labels[i], theme.GetStyle("text.small"))
        
        x += barWidth + gap
    }
}

```

---

### 21. 异步并发与 Loading 模式 (Async & Spinner)

在 TUI 中，**主循环 (Event Loop) 不能阻塞**。任何 HTTP 请求、数据库查询都必须在 Goroutine 中执行。

#### 设计模式：Command & Callback

1. **Spinner 组件**: 这是一个基于时间的动画组件。
```go
type Spinner struct {
    *component.BaseComponent
    frames    []rune // e.g. ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
    frameIdx  int
    interval  time.Duration
    stopChan  chan struct{}
}

// Tick 驱动动画
func (s *Spinner) Start() {
    ticker := time.NewTicker(s.interval)
    go func() {
        for {
            select {
            case <-ticker.C:
                s.mu.Lock()
                s.frameIdx = (s.frameIdx + 1) % len(s.frames)
                s.mu.Unlock()
                // 关键：通知 App 状态已变，需要重绘
                s.MarkDirty() 
            case <-s.stopChan:
                return
            }
        }
    }()
}

```


2. **异步数据加载器 (DataLoader)**:
不要在 `Paint` 中加载数据。组件初始化或 `OnMount` 时启动加载。
```go
type DataView struct {
    *component.BaseComponent
    isLoading bool
    data      interface{}
    err       error
    spinner   *Spinner
}

func (d *DataView) LoadData() {
    d.SetLoading(true)
    go func() {
        // 模拟耗时操作
        result, err := db.Query(...)

        // 回到主线程上下文更新状态
        // 注意：直接修改状态通常是安全的（只要有锁），
        // 但触发重绘 (MarkDirty) 需要通过回调或事件总线
        d.mu.Lock()
        d.data = result
        d.err = err
        d.isLoading = false
        d.mu.Unlock()

        d.MarkDirty() // 请求下一帧重绘
    }()
}

func (d *DataView) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    if d.isLoading {
        d.spinner.Paint(ctx, buf) // 居中绘制 Spinner
        return
    }
    // 绘制正常数据...
}

```



---

### 22. 命令面板 (Command Palette / Spotlight)

参考 VS Code 的 `Ctrl+P` 或 `Cmd+K`，这是提升键盘效率的关键。

#### 组合式设计

Command Palette 本质上是：`Overlay` + `FilterInput` + `FilteredList`。

```go
type CommandItem struct {
    Title    string
    Shortcut string
    Action   func() // 触发的回调
}

type CommandPalette struct {
    *component.BaseComponent
    input    *TextInput
    list     *List
    allItems []CommandItem
    visible  bool
}

// 核心逻辑：模糊搜索
func (cp *CommandPalette) OnInputChange(text string) {
    // 使用简单的包含匹配，或引入 fuzzy 库
    var filtered []ListItem
    for _, item := range cp.allItems {
        if strings.Contains(strings.ToLower(item.Title), strings.ToLower(text)) {
            filtered = append(filtered, item)
        }
    }
    cp.list.SetItems(filtered)
    cp.list.Select(0) // 默认选中第一项
}

// 全局集成
func (app *App) HandleAction(a action.Action) {
    // 全局拦截 Ctrl+P
    if a.Key == action.KeyCtrlP {
        app.commandPalette.Show()
        app.SetFocus(app.commandPalette.input) // 焦点劫持
        return
    }
    
    // 如果 Palette 打开，它独占输入
    if app.commandPalette.IsVisible() {
        if a.Key == action.KeyEsc {
            app.commandPalette.Hide()
            app.RestoreFocus() // 恢复之前的焦点
            return
        }
        app.commandPalette.HandleAction(a)
        return
    }
    
    // ... 其他逻辑
}

```

---

### 23. 综合布局系统：Grid Layout

Flex 适合一维布局（行或列），但构建复杂的 Dashboard 需要二维 Grid。

#### 定义 Grid

借鉴 CSS Grid 概念，但简化。

```go
type Grid struct {
    *component.BaseComponent
    Rows    []int // 行高，0表示自适应(1fr)，>0表示固定行数
    Columns []int // 列宽
    Cells   map[string]component.Component // "0:0" -> Comp, "0:1" -> Comp
    Gaps    int
}

func (g *Grid) Measure(maxWidth, maxHeight int) (int, int) {
    // 1. 计算固定行/列的总占用
    // 2. 将剩余空间分配给自适应行/列 (Fr unit calculation)
    // 3. 缓存计算出的每个网格单元的 (x, y, w, h)
    return maxWidth, maxHeight
}

func (g *Grid) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 遍历所有单元格组件
    for key, comp := range g.Cells {
        // 解析 key "row:col"
        r, c := parseKey(key)
        
        // 获取预计算的布局区域
        rect := g.getCellRect(r, c)
        
        // 创建子上下文并绘制
        subCtx := ctx.WithRect(rect)
        comp.Paint(subCtx, buf)
    }
}

```

#### Dashboard 示例结构

```go
dashboard := NewGrid(
    []int{3, 0, 1},    // Rows: Header(3), Content(Flex), Footer(1)
    []int{20, 0},      // Cols: Sidebar(20), Main(Flex)
)

dashboard.AddCell(0, 0, NewHeader(), 1, 2) // Header 跨两列
dashboard.AddCell(1, 0, NewSidebar())      // Sidebar 左侧
dashboard.AddCell(1, 1, NewContent())      // Content 右侧
dashboard.AddCell(2, 0, NewFooter(), 1, 2) // Footer 跨两列

```

---

### 24. 终极 Checklist：生产级质量保证

在完成上述组件后，确保进行以下优化：

1. **Resize Resilience (抗缩放)**:
* 在终端窗口极小的情况下（如 20x10），你的 Grid 或 Flex 是否会 Panic？
* **策略**: 设置 `MinSize`，如果空间不足，隐藏非关键组件（如 Sidebar），或者显示一个 "Screen too small" 的提示层。


2. **Mouse Support (鼠标支持)**:
* 虽然是 TUI，但现代终端都支持鼠标点击和滚轮。
* **实现**: `BaseComponent` 增加 `HitTest(x, y) bool`。
* 在事件循环中，如果收到 `MouseLeftClick`，遍历组件树找到最深层的 `HitTest` 为 true 的组件，触发其 `OnClick`。


3. **Panic Recovery**:
* 个别组件的崩溃不应导致整个 App 退出。
* 在 `app.Paint()` 内部做 `defer recover()`，如果某个组件绘制失败，在该区域绘制一个红色的 "Error" 框代替。


4. **Logger 集成**:
* TUI 会独占 stdout/stderr。
* **必须** 将日志输出到文件（如 `app.log`）或侧载到内存中供 Debug 组件查看。绝对不能使用 `fmt.Println`。



### 总结

通过这三部分的指导，我们构建了一个完整的 TUI 框架体系：

1. **基础层**: `Paint`, `Theme`, `Event Loop`。
2. **核心组件**: `Input`, `List`, `Table`, `Form`。
3. **高级特性**: `Canvas`, `Grid`, `Async`, `Overlay`。

遵循 `yao/tui/framework` 的设计哲学（Direct Mode + Component Interface），你可以实现一个不依赖 Bubbles/BubbleTea，但功能同样强大且更易于集成到现有 Yao 生态系统中的 TUI 库。

继续我们在 Yao TUI 架构下的组件开发设计。在前三个部分中，我们构建了从原子组件（Input）到复合组件（Table, Form）再到布局系统（Flex, Grid）的完整视图。

第四部分，我们将重点关注 **应用级架构**：如何管理多页面导航（Screen Management）、如何实现页内切换（Tabs）、以及最关键的——如何将这些 Go 原生组件与 **Yao 的 V8 脚本引擎** 集成，使 TUI 具备动态业务能力。

---

### 25. 屏幕管理与路由系统 (Screen Manager & Router)

一个完整的 TUI 应用不仅仅是一个界面，通常包含“登录”、“主菜单”、“详情页”等多个视图。我们需要一个 `ScreenManager` 来管理这些视图的切换。

#### 设计理念

* **Stack-based (栈式管理)**: 支持 `Push`（进入新页面）和 `Pop`（返回上一页）。
* **Lifecycle (生命周期)**: 页面切换时触发 `OnMount` (加载数据) 和 `OnUnmount` (清理资源/停止定时器)。

#### 接口定义

```go
// Screen 代表一个全屏的视图单元
type Screen interface {
    component.Component // 继承基础组件能力
    
    // 生命周期
    OnEnter(params map[string]interface{}) // 进入时触发，接收参数
    OnLeave()                              // 离开/被覆盖时触发
    OnResume()                             // 从上层页面返回时触发
}

// ScreenManager 管理器
type ScreenManager struct {
    stack []Screen
    mu    sync.RWMutex
}

func (sm *ScreenManager) Push(s Screen, params map[string]interface{}) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    // 暂停当前页面
    if len(sm.stack) > 0 {
        sm.stack[len(sm.stack)-1].OnLeave()
    }
    
    // 压栈并初始化新页面
    sm.stack = append(sm.stack, s)
    s.Mount(nil) // Screen 通常作为根节点，Parent 为 nil 或 App
    s.OnEnter(params)
}

func (sm *ScreenManager) Pop() {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    if len(sm.stack) <= 1 {
        return // 保持至少一个根页面
    }
    
    // 移除栈顶
    popped := sm.stack[len(sm.stack)-1]
    popped.OnLeave()
    popped.Unmount()
    sm.stack = sm.stack[:len(sm.stack)-1]
    
    // 恢复下层页面
    top := sm.stack[len(sm.stack)-1]
    top.OnResume()
}

// Paint 只绘制栈顶页面
func (sm *ScreenManager) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    if len(sm.stack) > 0 {
        sm.stack[len(sm.stack)-1].Paint(ctx, buf)
    }
}

```

---

### 26. 导航组件：Tabs (选项卡)

Tabs 是在同一 Screen 内切换视图的标准方式，常用于详情页（例如：基本信息 | 配置 | 日志）。

#### 结构设计

```go
type TabItem struct {
    ID      string
    Title   string
    Content component.Component
}

type Tabs struct {
    *component.BaseComponent
    items      []TabItem
    activeIdx  int
    activeStyle style.Style
}

```

#### 绘制逻辑

1. **Tab Bar (顶部栏)**: 水平排列 Title。
* 计算每个 Title 的宽度。
* 选中项应用 `activeStyle`（通常是下划线或反色）。


2. **Content (内容区)**:
* 获取当前激活的 `Content` 组件。
* 设置其位置为 `(x, y + 1)`，尺寸为 `(width, height - 1)`。
* 调用子组件的 `Paint`。



#### 交互逻辑

* `Left` / `Right`: 切换 `activeIdx`。
* **Lazy Loading (懒加载)**: 只有当 Tab 被激活时，才调用其 Content 的 `Init` 或 `LoadData`，避免一次性加载所有 Tab 的数据。

---

### 27. 核心集成：V8 脚本引擎绑定 (Yao Script Binding)

这是 Yao TUI 与普通 Go TUI 库最大的区别。组件的行为（Action）需要能触发 Yao 定义的 DSL 处理器或 JavaScript 函数。

#### 事件处理器设计

我们需要一个通用的 `ProcessHandler` 来桥接 Go 事件和 Yao 处理器。

```go
// ActionHandler 定义组件如何响应事件
// 例如: "onClick": {"process": "scripts.user.Login", "args": ["{{username}}", "{{password}}"]}
type ActionSpec struct {
    Process string        `json:"process"`
    Args    []interface{} `json:"args"`
}

// Component Mixin
type Scriptable struct {
    actions map[string]ActionSpec // "enter" -> Spec, "click" -> Spec
}

func (c *Scriptable) Trigger(eventName string, ctx *component.ComponentContext) {
    spec, ok := c.actions[eventName]
    if !ok {
        return
    }

    // 1. 解析参数 (支持 {{value}} 引用组件状态)
    resolvedArgs := c.resolveArgs(spec.Args, ctx)

    // 2. 异步执行 Yao Process
    go func() {
        res, err := yao.Process(spec.Process, resolvedArgs...)
        
        // 3. 处理回调 (更新 UI)
        if err != nil {
            ctx.App.NotifyError(err.Error())
        } else {
            // 可能需要根据返回值更新组件状态
            // 这通常需要一个回调机制或 UI 刷新指令
            ctx.App.Dispatch(action.NewUpdateStateAction(c.ID(), res))
        }
    }()
}

```

#### 实战示例：登录按钮

在 DSL 中定义：

```json
{
  "widget": "Button",
  "props": { "text": "Login" },
  "on": {
    "enter": {
      "process": "scripts.auth.login",
      "args": ["{{input_user.value}}", "{{input_pass.value}}"]
    }
  }
}

```

在 Go `HandleAction` 中：

```go
func (b *Button) HandleAction(a action.Action) bool {
    if a.Type == action.ActionSubmit {
        // 触发 "enter" 事件，将控制权交给 V8 桥接层
        b.Trigger("enter", b.GetContext())
        return true
    }
    return false
}

```

---

### 28. 数据绑定 (Data Binding)

除了事件调用，组件的值往往需要绑定到数据模型。

#### 响应式设计

1. **DataStore**: 全局或页面级的 Key-Value 存储。
2. **Binding Syntax**: 使用 `bind: "user.name"` 标记组件。
3. **Watcher**:
* 当 `TextInput` 修改值时 -> 更新 `DataStore["user.name"]`。
* 当 `DataStore` 变更时 -> 查找所有绑定该 Key 的组件 -> `SetValue()` -> `MarkDirty()`。



```go
type DataStore struct {
    data      map[string]interface{}
    listeners map[string][]component.Component
    mu        sync.RWMutex
}

func (ds *DataStore) Set(key string, val interface{}) {
    ds.mu.Lock()
    ds.data[key] = val
    comps := ds.listeners[key]
    ds.mu.Unlock()

    // 通知绑定的组件更新
    for _, c := range comps {
        if setter, ok := c.(component.ValueSetter); ok {
            setter.SetValue(val)
            c.MarkDirty()
        }
    }
}

```

---

### 29. 终极性能优化：Diffing & Partial Rendering

对于复杂的 Dashboard，每次全屏重绘（Full Repaint）开销很大。

#### 优化策略

1. **Render Cache (渲染缓存)**:
* 每个组件维护一个 `lastBuffer *paint.Buffer`。
* 在 `Paint` 开始前，检查 `IsDirty()`。
* 如果 `!IsDirty()`，直接将 `lastBuffer` 复制到当前 Frame，无需重新计算布局和字符。


2. **Z-Index Layers (逻辑层)**:
* 将 UI 分为：`Base Layer` (背景、边框) 和 `Active Layer` (光标、输入框、Spinner)。
* 大部分时间只需要重绘 `Active Layer`。


3. **IO 节流**:
* 不要对每次 `SetCell` 都进行 syscall 写终端。
* 使用 `bufio.Writer`，并在每一帧渲染结束后调用 `Flush()`。



---

### 总结：Yao TUI 组件库架构全景

通过这四部分的指导，我们建立了一个金字塔式的架构：

1. **顶层 (Scripting/DSL)**:
* 通过 JSON/Yao DSL 定义界面。
* 通过 `Scriptable` 接口与 V8 引擎交互。
* 通过 `DataStore` 实现数据驱动。


2. **管理层 (Framework)**:
* `ScreenManager` 处理页面流。
* `OverlayManager` 处理弹窗。
* `FocusManager` 处理键盘导航。


3. **组件层 (Widgets)**:
* **容器**: `Flex`, `Grid`, `Tabs`, `Form`.
* **交互**: `Input`, `Select`, `Button`, `Table` (Virtual).
* **视觉**: `Chart`, `Canvas`, `Sparkline`.


4. **底层 (Core)**:
* `Painter`: 基于 `(x,y)` 的绘制原语。
* `EventLoop`: 针对 TUI 优化的消息循环。



**下一步建议**:

* 先实现 `TextInput`, `List`, `Flex` 三个核心组件，这就足以构建 80% 的 CLI 工具界面。
* 然后实现 V8 绑定，打通 Yao 的业务逻辑。
* 最后打磨 `Table` 和 `Charts` 等高级组件。

这套方案完全去除了对 `bubbles` 等外部重型库的依赖，让 Yao TUI 保持轻量、可控且高度定制化。

继续我们在 Yao TUI 架构下的组件开发设计。在前三个部分中，我们构建了从原子组件（Input）到复合组件（Table, Form）再到布局系统（Flex, Grid）的完整视图。

第四部分，我们将重点关注 **应用级架构**：如何管理多页面导航（Screen Management）、如何实现页内切换（Tabs）、以及最关键的——如何将这些 Go 原生组件与 **Yao 的 V8 脚本引擎** 集成，使 TUI 具备动态业务能力。

---

### 25. 屏幕管理与路由系统 (Screen Manager & Router)

一个完整的 TUI 应用不仅仅是一个界面，通常包含“登录”、“主菜单”、“详情页”等多个视图。我们需要一个 `ScreenManager` 来管理这些视图的切换。

#### 设计理念

* **Stack-based (栈式管理)**: 支持 `Push`（进入新页面）和 `Pop`（返回上一页）。
* **Lifecycle (生命周期)**: 页面切换时触发 `OnMount` (加载数据) 和 `OnUnmount` (清理资源/停止定时器)。

#### 接口定义

```go
// Screen 代表一个全屏的视图单元
type Screen interface {
    component.Component // 继承基础组件能力
    
    // 生命周期
    OnEnter(params map[string]interface{}) // 进入时触发，接收参数
    OnLeave()                              // 离开/被覆盖时触发
    OnResume()                             // 从上层页面返回时触发
}

// ScreenManager 管理器
type ScreenManager struct {
    stack []Screen
    mu    sync.RWMutex
}

func (sm *ScreenManager) Push(s Screen, params map[string]interface{}) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    // 暂停当前页面
    if len(sm.stack) > 0 {
        sm.stack[len(sm.stack)-1].OnLeave()
    }
    
    // 压栈并初始化新页面
    sm.stack = append(sm.stack, s)
    s.Mount(nil) // Screen 通常作为根节点，Parent 为 nil 或 App
    s.OnEnter(params)
}

func (sm *ScreenManager) Pop() {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    if len(sm.stack) <= 1 {
        return // 保持至少一个根页面
    }
    
    // 移除栈顶
    popped := sm.stack[len(sm.stack)-1]
    popped.OnLeave()
    popped.Unmount()
    sm.stack = sm.stack[:len(sm.stack)-1]
    
    // 恢复下层页面
    top := sm.stack[len(sm.stack)-1]
    top.OnResume()
}

// Paint 只绘制栈顶页面
func (sm *ScreenManager) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    if len(sm.stack) > 0 {
        sm.stack[len(sm.stack)-1].Paint(ctx, buf)
    }
}

```

---

### 26. 导航组件：Tabs (选项卡)

Tabs 是在同一 Screen 内切换视图的标准方式，常用于详情页（例如：基本信息 | 配置 | 日志）。

#### 结构设计

```go
type TabItem struct {
    ID      string
    Title   string
    Content component.Component
}

type Tabs struct {
    *component.BaseComponent
    items      []TabItem
    activeIdx  int
    activeStyle style.Style
}

```

#### 绘制逻辑

1. **Tab Bar (顶部栏)**: 水平排列 Title。
* 计算每个 Title 的宽度。
* 选中项应用 `activeStyle`（通常是下划线或反色）。


2. **Content (内容区)**:
* 获取当前激活的 `Content` 组件。
* 设置其位置为 `(x, y + 1)`，尺寸为 `(width, height - 1)`。
* 调用子组件的 `Paint`。



#### 交互逻辑

* `Left` / `Right`: 切换 `activeIdx`。
* **Lazy Loading (懒加载)**: 只有当 Tab 被激活时，才调用其 Content 的 `Init` 或 `LoadData`，避免一次性加载所有 Tab 的数据。

---

### 27. 核心集成：V8 脚本引擎绑定 (Yao Script Binding)

这是 Yao TUI 与普通 Go TUI 库最大的区别。组件的行为（Action）需要能触发 Yao 定义的 DSL 处理器或 JavaScript 函数。

#### 事件处理器设计

我们需要一个通用的 `ProcessHandler` 来桥接 Go 事件和 Yao 处理器。

```go
// ActionHandler 定义组件如何响应事件
// 例如: "onClick": {"process": "scripts.user.Login", "args": ["{{username}}", "{{password}}"]}
type ActionSpec struct {
    Process string        `json:"process"`
    Args    []interface{} `json:"args"`
}

// Component Mixin
type Scriptable struct {
    actions map[string]ActionSpec // "enter" -> Spec, "click" -> Spec
}

func (c *Scriptable) Trigger(eventName string, ctx *component.ComponentContext) {
    spec, ok := c.actions[eventName]
    if !ok {
        return
    }

    // 1. 解析参数 (支持 {{value}} 引用组件状态)
    resolvedArgs := c.resolveArgs(spec.Args, ctx)

    // 2. 异步执行 Yao Process
    go func() {
        res, err := yao.Process(spec.Process, resolvedArgs...)
        
        // 3. 处理回调 (更新 UI)
        if err != nil {
            ctx.App.NotifyError(err.Error())
        } else {
            // 可能需要根据返回值更新组件状态
            // 这通常需要一个回调机制或 UI 刷新指令
            ctx.App.Dispatch(action.NewUpdateStateAction(c.ID(), res))
        }
    }()
}

```

#### 实战示例：登录按钮

在 DSL 中定义：

```json
{
  "widget": "Button",
  "props": { "text": "Login" },
  "on": {
    "enter": {
      "process": "scripts.auth.login",
      "args": ["{{input_user.value}}", "{{input_pass.value}}"]
    }
  }
}

```

在 Go `HandleAction` 中：

```go
func (b *Button) HandleAction(a action.Action) bool {
    if a.Type == action.ActionSubmit {
        // 触发 "enter" 事件，将控制权交给 V8 桥接层
        b.Trigger("enter", b.GetContext())
        return true
    }
    return false
}

```

---

### 28. 数据绑定 (Data Binding)

除了事件调用，组件的值往往需要绑定到数据模型。

#### 响应式设计

1. **DataStore**: 全局或页面级的 Key-Value 存储。
2. **Binding Syntax**: 使用 `bind: "user.name"` 标记组件。
3. **Watcher**:
* 当 `TextInput` 修改值时 -> 更新 `DataStore["user.name"]`。
* 当 `DataStore` 变更时 -> 查找所有绑定该 Key 的组件 -> `SetValue()` -> `MarkDirty()`。



```go
type DataStore struct {
    data      map[string]interface{}
    listeners map[string][]component.Component
    mu        sync.RWMutex
}

func (ds *DataStore) Set(key string, val interface{}) {
    ds.mu.Lock()
    ds.data[key] = val
    comps := ds.listeners[key]
    ds.mu.Unlock()

    // 通知绑定的组件更新
    for _, c := range comps {
        if setter, ok := c.(component.ValueSetter); ok {
            setter.SetValue(val)
            c.MarkDirty()
        }
    }
}

```

---

### 29. 终极性能优化：Diffing & Partial Rendering

对于复杂的 Dashboard，每次全屏重绘（Full Repaint）开销很大。

#### 优化策略

1. **Render Cache (渲染缓存)**:
* 每个组件维护一个 `lastBuffer *paint.Buffer`。
* 在 `Paint` 开始前，检查 `IsDirty()`。
* 如果 `!IsDirty()`，直接将 `lastBuffer` 复制到当前 Frame，无需重新计算布局和字符。


2. **Z-Index Layers (逻辑层)**:
* 将 UI 分为：`Base Layer` (背景、边框) 和 `Active Layer` (光标、输入框、Spinner)。
* 大部分时间只需要重绘 `Active Layer`。


3. **IO 节流**:
* 不要对每次 `SetCell` 都进行 syscall 写终端。
* 使用 `bufio.Writer`，并在每一帧渲染结束后调用 `Flush()`。



---

### 总结：Yao TUI 组件库架构全景

通过这四部分的指导，我们建立了一个金字塔式的架构：

1. **顶层 (Scripting/DSL)**:
* 通过 JSON/Yao DSL 定义界面。
* 通过 `Scriptable` 接口与 V8 引擎交互。
* 通过 `DataStore` 实现数据驱动。


2. **管理层 (Framework)**:
* `ScreenManager` 处理页面流。
* `OverlayManager` 处理弹窗。
* `FocusManager` 处理键盘导航。


3. **组件层 (Widgets)**:
* **容器**: `Flex`, `Grid`, `Tabs`, `Form`.
* **交互**: `Input`, `Select`, `Button`, `Table` (Virtual).
* **视觉**: `Chart`, `Canvas`, `Sparkline`.


4. **底层 (Core)**:
* `Painter`: 基于 `(x,y)` 的绘制原语。
* `EventLoop`: 针对 TUI 优化的消息循环。



**下一步建议**:

* 先实现 `TextInput`, `List`, `Flex` 三个核心组件，这就足以构建 80% 的 CLI 工具界面。
* 然后实现 V8 绑定，打通 Yao 的业务逻辑。
* 最后打磨 `Table` 和 `Charts` 等高级组件。

这套方案完全去除了对 `bubbles` 等外部重型库的依赖，让 Yao TUI 保持轻量、可控且高度定制化。

继续我们在 Yao TUI 架构下的组件开发指导。在前几个部分，我们已经构建了组件的功能骨架、布局系统和交互逻辑。现在，我们将核心聚焦在 **"绘制 (Rendering)"** 和 **"UI 设计 (Visual Design)"** 上。

在 TUI（终端用户界面）中，绘制不仅仅是输出文本，更是对 **字符网格 (Grid of Cells)** 的精确控制。如何在有限的字符空间内实现美观、清晰且高性能的界面，是 TUI 开发的精髓。

---

### 31. 绘制管线与坐标系统 (Rendering Pipeline)

在 `framework/component/component.go` 的设计中，`Paint` 方法接收一个 `PaintContext`。理解这个上下文的变换是绘制正确界面的关键。

#### A. 相对坐标与绝对坐标

通常组件内部只关心**相对坐标**（即：(0,0) 是组件的左上角），而 `PaintContext` 负责传递偏移量。

```go
// PaintContext 定义
type PaintContext struct {
    X, Y            int // 绝对坐标偏移量
    AvailableWidth  int // 组件可用宽度
    AvailableHeight int // 组件可用高度
    ClipRect        Rect // 剪裁区域 (关键！)
    ZIndex          int // 层级 (用于 Overlay)
}

// 在 Container (如 Flex/Box) 中调用子组件 Paint
func (f *Flex) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    for _, child := range f.children {
        relX, relY := child.GetPosition()
        w, h := child.GetSize()

        // 1. 坐标转换：父绝对坐标 + 子相对坐标
        childCtx := ctx
        childCtx.X = ctx.X + relX
        childCtx.Y = ctx.Y + relY
        childCtx.AvailableWidth = w
        childCtx.AvailableHeight = h

        // 2. 计算剪裁区域 (Clipping)
        // 确保子组件不会画到父容器外面（这对于 ScrollView 至关重要）
        childAbsRect := Rect{X: childCtx.X, Y: childCtx.Y, W: w, H: h}
        if intersected, ok := ctx.ClipRect.Intersect(childAbsRect); ok {
            childCtx.ClipRect = intersected
            child.Paint(childCtx, buf)
        }
    }
}

```

#### B. 剪裁 (Clipping) 的实现

在 `paint.Buffer` 的 `SetCell` 方法中，**必须** 检查坐标是否在 `ClipRect` 内。

```go
func (buf *Buffer) SetCell(x, y int, char rune, style Style) {
    // 全局边界检查
    if x < 0 || x >= buf.Width || y < 0 || y >= buf.Height {
        return
    }
    // 写入 Buffer
    buf.Cells[y][x] = Cell{Char: char, Style: style}
}

// 更好的做法是让 Paint 方法自己处理，或者封装一个带 Clip 的 Painter
type Painter struct {
    Buffer *Buffer
    Ctx    PaintContext
}

func (p *Painter) Draw(x, y int, char rune, style Style) {
    absX := p.Ctx.X + x
    absY := p.Ctx.Y + y
    
    // 剪裁检查
    if !p.Ctx.ClipRect.Contains(absX, absY) {
        return
    }
    p.Buffer.SetCell(absX, absY, char, style)
}

```

---

### 32. UI 设计：盒模型 (The Box Model in TUI)

在 CSS 中我们有 Margin/Border/Padding。在 TUI 中，由于最小单位是一个字符，这显得格外昂贵。

#### 设计策略

* **Border (边框)**: 占用 1 个字符宽度。
* **Padding (内边距)**: 通常至少 1 个字符，为了紧凑有时设为 0。
* **Margin (外边距)**: 用于组件间隙。

#### 边框绘制器 (Border Painter)

实现一个通用的边框绘制工具，支持不同风格（单线、双线、圆角）。

```go
type BorderStyle int
const (
    BorderNone BorderStyle = iota
    BorderSingle // ┌─┐│└─┘
    BorderDouble // ╔═╗║╚═╝
    BorderRounded // ╭─╮│╰─╯
)

func DrawBorder(p *Painter, w, h int, style BorderStyle, color theme.Color) {
    // 获取字符集
    chars := getBorderChars(style) 
    
    // 绘制四个角
    p.Draw(0, 0, chars.TopLeft, color)
    p.Draw(w-1, 0, chars.TopRight, color)
    p.Draw(0, h-1, chars.BottomLeft, color)
    p.Draw(w-1, h-1, chars.BottomRight, color)
    
    // 绘制线 (使用循环或 fast path)
    for x := 1; x < w-1; x++ {
        p.Draw(x, 0, chars.Horizontal, color)     // Top
        p.Draw(x, h-1, chars.Horizontal, color)   // Bottom
    }
    for y := 1; y < h-1; y++ {
        p.Draw(0, y, chars.Vertical, color)       // Left
        p.Draw(w-1, y, chars.Vertical, color)     // Right
    }
}

```

---

### 33. 视觉状态反馈 (Visual State Feedback)

用户在终端中没有鼠标指针的即时反馈（Hover 虽有但非必须），因此 **Focus（焦点）** 状态的设计至关重要。

#### 设计规范

1. **Normal**: 默认状态，使用次要颜色（如灰色、白色）。
2. **Focus**: 获得焦点，**必须** 有显著变化。
* 改变边框颜色（如变蓝/绿）。
* 文字加粗或变色。
* 如果是输入框，光标必须闪烁。


3. **Active/Selected**: 选中状态（如 List 中的选中项）。
* 通常使用 **反色 (Reverse)**：背景色变亮，前景色变暗。
* 或者行首添加指示符 `> Item 1`。


4. **Disabled**: 变暗（Dim），通常使用深灰色。

#### 代码示例 (样式状态机)

```go
func (c *Button) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 1. 确定基础样式 Key
    styleKey := "button.normal"
    if c.IsDisabled() {
        styleKey = "button.disabled"
    } else if c.IsPressed() {
        styleKey = "button.active"
    } else if c.IsFocused() {
        styleKey = "button.focus"
    }

    // 2. 从主题获取具体样式
    s := theme.GetGlobal().GetStyle(styleKey)
    
    // 3. 绘制
    // 比如：[ Submit ]
    // Focus 时：> [ Submit ] < 或者边框变色
}

```

---

### 34. 图标与符号 (Icons & Symbols)

现代 TUI 很大程度上依赖 [Nerd Fonts](https://www.nerdfonts.com/) 来提供图形化体验。

#### 最佳实践

1. **宽字符处理**: 很多图标看起来是 1 个字符宽，但实际上可能占用 2 个单元格，或者渲染时溢出。
* **强制后缀空格**: 在图标后显式加一个空格 ` `，防止与文字重叠。


2. **Fallback (回退)**: 并非所有用户都安装了 Nerd Fonts。
* 在配置中提供 `UseNerdFonts bool` 选项。
* 如果是 `false`，使用 ASCII 替代（如 `[x]` 代替 ``）。



```go
var Icons = map[string]struct{ Nerd, ASCII string }{
    "check":    {"", "[x]"},
    "uncheck":  {"", "[ ]"},
    "folder":   {"", "+"},
    "file":     {"", "-"},
    "search":   {"", "?"},
}

func GetIcon(name string) string {
    icon := Icons[name]
    if config.UseNerdFonts {
        return icon.Nerd + " " // 自动补空格
    }
    return icon.ASCII + " "
}

```

---

### 35. 响应式设计 (Responsive Design in TUI)

终端窗口大小随意变化，组件必须适应。

#### 策略

1. **断点隐藏 (Hide on Breakpoint)**:
* 在 Table 中，如果宽度 `< 80`，隐藏次要列（"Created At", "ID"）。
* 在 Dashboard 中，如果宽度太小，将侧边栏（Sidebar）变为抽屉式（Drawer/Overlay），默认隐藏。


2. **自适应布局 (Fluid Layout)**:
* 使用 `Flex` 的 `Grow` 属性。
* 文本过长时，使用 `Truncate`（截断: `text...`）或 `Wrap`（换行）。



```go
// 在 Table 的 Measure 或 Paint 中
func (t *Table) calculateVisibleColumns(totalWidth int) []Column {
    var visible []Column
    usedWidth := 0
    
    // 优先显示主要列
    for _, col := range t.columns {
        if col.Priority == High {
            visible = append(visible, col)
            usedWidth += col.Width
        }
    }
    
    // 空间足够则显示次要列
    for _, col := range t.columns {
        if col.Priority == Low && usedWidth + col.Width <= totalWidth {
            visible = append(visible, col)
            usedWidth += col.Width
        }
    }
    return visible
}

```

---

### 36. 性能优化：脏矩形渲染 (Dirty Rect Rendering)

虽然全屏重绘（Clear + Paint All）最简单，但在低带宽（SSH）或复杂界面下会闪烁或卡顿。

#### 设计思路

1. **Component Dirty Flag**: 每个组件维护一个 `isDirty` bool。
2. **Partial Paint**:
* 当 `Input` 输入文字时，只标记 `Input` 组件为 Dirty。
* 渲染循环只调用 Dirty 组件的 `Paint`。
* **难点**: 如果组件大小变了，或者位置变了，必须标记父容器甚至全屏 Dirty。


3. **Buffer Diffing (更实用的方案)**:
* 维护两个 Buffer: `CurrentFrame` 和 `LastFrame`。
* 每次全量绘制到 `CurrentFrame`（内存操作很快）。
* 在 `Flush` 阶段，逐个单元格比较。
* 只向终端发送**不同**的单元格指令，并尽量合并光标移动指令（跳过相同的区域）。



```go
// Flush 伪代码
func (eng *Engine) Flush() {
    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            newCell := eng.currentBuf[y][x]
            oldCell := eng.lastBuf[y][x]
            
            if newCell != oldCell {
                eng.term.MoveCursor(x, y)
                eng.term.SetStyle(newCell.Style)
                eng.term.Print(newCell.Char)
            }
        }
    }
    // Swap buffers
    eng.lastBuf, eng.currentBuf = eng.currentBuf, eng.lastBuf
    eng.currentBuf.Clear()
}

```

*注：这是 `tcell` 或 `tea` 底层已经在做的事情。如果你直接使用 `tcell`，利用好 `Show()` 即可，重点是**不要在每一帧都重新计算复杂的布局数据**。*

### 总结

在 Yao TUI 的组件绘制与设计中：

1. **精准控制**: 利用 `PaintContext` 和 `ClipRect` 实现精确的区域绘制。
2. **视觉分层**: 通过 Box Model、Focus 状态和颜色主题建立清晰的视觉层级。
3. **适应性**: 设计组件时始终考虑 "如果屏幕只有 40 列宽怎么办"。
4. **美学**: 在纯文本限制下，利用 unicode 字符、留白和对齐来提升品质感。

遵循这些原则，结合之前定义的组件架构，你将能开发出既美观又健壮的 TUI 系统。

继续深入 Yao TUI 架构的组件开发指导。在前几个部分，我们已经完成了从基础架构、核心组件到高级布局和绘制系统的设计。

这一部分我们将关注 **"动态交互与生命周期 (Dynamics & Lifecycle)"**，这是让界面“活”起来的关键，以及 **"工程化与测试 (Engineering & Testing)"**，这是保证组件库质量的基石。

---

### 37. 动画系统 (Animation System)

在 TUI 中实现动画比 GUI 更具挑战性，因为我们受限于字符网格，无法做到亚像素级的平滑移动。但我们仍然可以通过 **插值 (Interpolation)** 和 **定时器 (Ticker)** 来实现过渡效果（如进度条、加载动画、侧边栏滑出）。

#### 架构设计

动画不应阻塞主线程。我们需要一个 `AnimationManager` 来驱动数值的变化。

1. **Animation State**:
```go
type Animation struct {
    StartVal  float64
    EndVal    float64
    Duration  time.Duration
    StartTime time.Time
    Easing    func(t float64) float64 // Easing function: Linear, EaseIn, EaseOut
    OnUpdate  func(val float64)       // 回调更新组件状态
    OnFinish  func()
}

```


2. **Easing Functions (缓动函数)**:
```go
func EaseOutQuad(t float64) float64 { return t * (2 - t) }
func Linear(t float64) float64 { return t }

```


3. **驱动循环**:
在 `App` 的主循环中，增加一个 `UpdateAnimations` 阶段。
```go
func (app *App) updateAnimations() {
    now := time.Now()
    active := app.animations[:0] // In-place filtering

    for _, anim := range app.animations {
        elapsed := now.Sub(anim.StartTime).Seconds()
        total := anim.Duration.Seconds()

        if elapsed >= total {
            anim.OnUpdate(anim.EndVal)
            if anim.OnFinish != nil { anim.OnFinish() }
            continue
        }

        progress := elapsed / total
        val := anim.StartVal + (anim.EndVal - anim.StartVal) * anim.Easing(progress)
        anim.OnUpdate(val)
        active = append(active, anim)

        app.MarkDirty() // 触发重绘
    }
    app.animations = active
}

```



#### 实战：侧边栏滑出动画

假设侧边栏宽度从 0 变到 20。

```go
func (s *Sidebar) Toggle() {
    start := s.width
    end := 0
    if !s.visible {
        end = 20
        s.visible = true
    }
    
    app.Animate(Animation{
        StartVal: float64(start),
        EndVal:   float64(end),
        Duration: 200 * time.Millisecond,
        Easing:   EaseOutQuad,
        OnUpdate: func(v float64) {
            s.width = int(v) // TUI 中取整
        },
    })
}

```

---

### 38. 组件生命周期 (Component Lifecycle)

为了管理复杂的资源（如定时器、数据库连接、文件句柄），组件需要生命周期钩子。

#### 接口定义

在 `component/component.go` 中扩展接口：

```go
type Lifecycle interface {
    OnMount()   // 组件首次挂载到组件树时调用
    OnUnmount() // 组件从组件树移除时调用
    OnUpdate()  // 组件属性或状态变更后调用
}

```

#### 实现策略

1. **OnMount**: 启动 `Ticker`（例如时钟组件），注册全局事件监听（例如监听窗口大小变化），或者发起异步数据请求。
2. **OnUnmount**: **必须** 停止所有 Goroutine 和 Ticker，防止内存泄漏。
```go
type Clock struct {
    *component.BaseComponent
    stopChan chan struct{}
}

func (c *Clock) OnMount() {
    c.stopChan = make(chan struct{})
    go func() {
        ticker := time.NewTicker(time.Second)
        for {
            select {
            case <-ticker.C:
                c.refreshTime()
            case <-c.stopChan:
                ticker.Stop()
                return
            }
        }
    }()
}

func (c *Clock) OnUnmount() {
    if c.stopChan != nil {
        close(c.stopChan)
    }
}

```



---

### 39. 高级输入处理：快捷键与鼠标 (Advanced Input)

#### Key Binding System (键位映射)

不要在组件中硬编码 `if key == 'q'`。使用 `KeyMap` 使得快捷键可配置。

```go
type KeyMap struct {
    Up    []string // e.g. ["up", "k"]
    Down  []string // e.g. ["down", "j"]
    Enter []string
}

var DefaultKeyMap = KeyMap{
    Up:   []string{"up", "k", "ctrl+p"},
    Down: []string{"down", "j", "ctrl+n"},
}

// 在组件中检查
func (l *List) HandleAction(a action.Action) bool {
    if Matches(a, DefaultKeyMap.Up) {
        l.SelectionUp()
        return true
    }
    // ...
}

```

#### 鼠标支持 (Mouse Support)

虽然是键盘优先，但鼠标支持（点击选择、滚轮滚动）能显著提升体验。

1. **Hit Testing (碰撞检测)**:
组件需要知道鼠标点击是否落在自己身上。
```go
// BaseComponent 已经有了 x, y, width, height
func (c *BaseComponent) Contains(x, y int) bool {
    // 注意：x, y 必须转换到组件的坐标系，或者组件存储绝对坐标
    absX, absY := c.GetAbsolutePosition()
    return x >= absX && x < absX+c.width && 
           y >= absY && y < absY+c.height
}

```


2. **事件冒泡**:
鼠标点击是一个点 `(x, y)`。事件分发器需要从根节点开始，找到**最深层**且包含该点的组件（Z-Index 最高的），触发其 `OnMouseClick`。

---

### 40. 测试策略 (Testing Strategy)

TUI 组件的测试不仅是逻辑测试，还包括 **视觉测试 (Visual Testing)**。

#### A. Headless Canvas (无头画布)

为了测试 `Paint` 方法，我们需要一个内存中的 `Buffer` 实现，不依赖真实终端。

```go
// TestContext 创建一个模拟的绘制环境
func NewTestContext(w, h int) (component.PaintContext, *paint.Buffer) {
    buf := paint.NewBuffer(w, h)
    ctx := component.PaintContext{
        AvailableWidth:  w,
        AvailableHeight: h,
        ClipRect: component.Rect{W: w, H: h},
    }
    return ctx, buf
}

func TestInputPaint(t *testing.T) {
    input := NewTextInput()
    input.SetValue("Hello")
    
    ctx, buf := NewTestContext(10, 1)
    input.Paint(ctx, buf)
    
    // 断言 Buffer 内容
    // 预期: [ H e l l o     ]
    assert.Equal(t, '[', buf.GetCell(0, 0).Char)
    assert.Equal(t, 'H', buf.GetCell(1, 0).Char)
}

```

#### B. Golden File Testing (快照测试)

将组件渲染结果（Buffer 的字符串表示）保存为 `.golden` 文件。下次测试时，将新渲染结果与文件比对。如果 UI 发生预期外的变化，测试失败。

```go
func TestTableSnapshot(t *testing.T) {
    table := NewTable(...)
    // ... setup data ...
    
    renderStr := RenderToString(table, 80, 20)
    
    // 比较 renderStr 和 "testdata/table.golden"
    // 如果不同 -> Fail
}

```

---

### 41. 错误处理与恢复 (Panic Recovery)

在 TUI 应用中，如果发生 panic，终端可能会处于 "Raw Mode"（乱码状态），导致用户终端无法使用。

**必须** 实现全局 Recover 机制：

```go
func (app *App) Run() (err error) {
    // 1. 确保退出时恢复终端
    defer func() {
        if r := recover(); r != nil {
            app.screen.Fini() // 恢复终端到正常模式
            fmt.Fprintf(os.Stderr, "App Panicked: %v\nStack: %s", r, debug.Stack())
            os.Exit(1)
        }
    }()
    
    // ... 主循环 ...
}

```

---

### 42. 开发者体验 (DX)

为了让其他开发者（或未来的你）方便地开发组件，提供以下工具：

1. **Debug Overlay**: 按 `F12` 显示 FPS、当前焦点组件 ID、鼠标坐标。
2. **Color Palette Viewer**: 一个展示所有 Theme 颜色的页面，方便校对。
3. **Component Gallery (Storybook)**: 一个包含所有组件示例的 Demo App，展示各种状态（Normal, Focus, Disabled, Error）。

---

### 总结与下一步

至此，我们已经完成了一套 **企业级 TUI 框架** 的完整设计蓝图：

1. **Core**: `Buffer`, `Painter`, `Event Loop`.
2. **Widgets**: `Input`, `List`, `Table`, `Tree`, `Charts`.
3. **Layout**: `Flex`, `Grid`, `Absolute`.
4. **System**: `Theme`, `Focus`, `Animation`, `Lifecycle`.
5. **Engineering**: `Testing`, `Panic Recovery`.

**实施建议**:

* 不要试图一次性写完所有组件。
* 先实现 `Box` (Layout) + `Text` (Display) + `Input` (Interactive)。
* 用这三个组件构建一个简单的 "Hello World" 表单。
* 一旦这套流程跑通（渲染 -> 输入 -> 状态更新 -> 重绘），再逐步添加 Table 和更复杂的组件。

---



#### 帮助栏组件 (Help/Status Bar)

底部通常有一个状态栏显示当前可用的快捷键。

```go
func (h *HelpBar) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 获取当前焦点组件
    focused := h.app.FocusManager.Current()
    
    // 获取该组件的 KeyMap
    if km, ok := focused.(KeyMappable); ok {
        mappings := km.GetKeyMap()
        // 渲染: [j/k] Navigate  [Enter] Select  [q] Quit
        h.renderShortcuts(ctx, buf, mappings)
    }
}

```


---

### 52. 系统级组件：FilePicker (文件选择器)

在 TUI 中，文件选择是一个高频需求，且逻辑复杂（路径导航、隐藏文件、权限）。

#### 设计方案

`FilePicker` = `Input` (当前路径) + `List` (文件列表) + `KeyboardHandler`。

```go
type FileInfo struct {
    Name  string
    IsDir bool
    Size  int64
}

type FilePicker struct {
    *component.BaseComponent
    currentPath string
    files       []FileInfo
    cursor      int
    onSelect    func(path string)
    showHidden  bool
}

func (fp *FilePicker) LoadDir(path string) {
    // 1. 读目录 (os.ReadDir)
    entries, _ := os.ReadDir(path)
    
    // 2. 排序 (文件夹在前)
    // 3. 过滤
    // 4. 更新 State
    fp.files = processedFiles
    fp.MarkDirty()
}

func (fp *FilePicker) HandleAction(a action.Action) bool {
    switch a.Type {
    case action.ActionEnter:
        selected := fp.files[fp.cursor]
        if selected.IsDir {
            // 进入目录
            fp.LoadDir(filepath.Join(fp.currentPath, selected.Name))
            fp.cursor = 0
        } else {
            // 选中文件
            if fp.onSelect != nil {
                fp.onSelect(filepath.Join(fp.currentPath, selected.Name))
            }
        }
        return true
    case action.ActionBackspace:
        // 返回上级
        fp.LoadDir(filepath.Dir(fp.currentPath))
        return true
    }
    // ... 处理上下滚动
    return false
}

```

---

### 53. 日志与调试控制台 (Log Viewer)

TUI 应用独占了标准输出，这导致 `fmt.Println` 调试变得不可能。你需要一个内置的日志查看器组件。

#### 实现

利用 `ringbuffer` 存储最近 N 条日志。

```go
// global/logger.go
var LogBuffer = ring.New(1000)

func Info(msg string) {
    LogBuffer.Value = fmt.Sprintf("[INFO] %s", msg)
    LogBuffer = LogBuffer.Next()
    // 触发 LogViewer 重绘 (如果有)
    events.Emit("log_updated", nil)
}

// widgets/log_viewer.go
func (lv *LogViewer) Paint(...) {
    // 遍历 LogBuffer 绘制到 Viewport
}

```

开发者可以通过按快捷键（如 `F10`）弹出一个 Modal，里面包含 `LogViewer`，实时查看后台逻辑的执行情况。

---





### 54. 完整的 DSL 定义参考 (The Complete Picture)

为了让这套系统可用，你需要定义一套清晰的 JSON/YAML 规范。

**`user_dashboard.tui.json`**:

```json
{
  "name": "User Dashboard",
  "layout": "Grid",
  "theme": "dark",
  "props": {
    "rows": [3, 0, 1], // Header, Content(Flex), Footer
    "columns": [20, 0] // Sidebar, Main
  },
  "state": {
    "currentUser": {},
    "menuItems": []
  },
  "hooks": {
    "onMount": {
      "process": "scripts.dashboard.InitData",
      "target": "state" // 结果合并到 state
    }
  },
  "children": [
    // [0,0] Header (Span 2 cols)
    {
      "type": "Box",
      "gridArea": { "r": 0, "c": 0, "w": 2, "h": 1 },
      "style": { "border": "bottom" },
      "children": [
        { "type": "Text", "value": "Yao Admin TUI", "style": "h1" }
      ]
    },
    
    // [1,0] Sidebar
    {
      "type": "List",
      "id": "sidebar_menu",
      "gridArea": { "r": 1, "c": 0 },
      "props": {
        "items": "{{menuItems}}", 
        "key": "id",
        "label": "name"
      },
      "events": {
        "onSelect": {
          "type": "Script.Run",
          "script": "scripts.dashboard.OnMenuSelect"
        }
      }
    },
    
    // [1,1] Main Content Area
    {
      "type": "Container",
      "id": "main_view",
      "gridArea": { "r": 1, "c": 1 },
      "children": [
        // 初始为空，由 Router 填充
      ]
    },
    
    // [2,0] Footer
    {
      "type": "Text",
      "gridArea": { "r": 2, "c": 0, "w": 2 },
      "value": "Ready. Press F1 for Help.",
      "style": { "fg": "gray" }
    }
  ]
}

```