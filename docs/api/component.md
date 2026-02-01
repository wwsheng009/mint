# 组件 API 参考

本文档介绍 mint UI 框架中组件相关的核心接口和类型。

## 目录

- [VNode 接口](#vnode-接口)
- [组件 VNode](#组件-vnode)
- [组件实例](#组件实例)
- [元素 VNode](#元素-vnode)
- [交互组件](#交互组件)
- [布局组件](#布局组件)

---

## VNode 接口

### VNode

所有虚拟节点的基础接口。

```go
type VNode interface {
    Type() VNodeType
    Props() Props
    SetProps(p Props)
    Children() []VNode
    SetChildren(children []VNode)
    Key() string
    SetKey(key string)
    Style() style.Style
    SetStyle(s style.Style)
}
```

### VNodeType

VNode 类型枚举。

```go
const (
    VNodeElement   VNodeType = iota  // 标准元素 (div, span 等)
    VNodeText                          // 文本节点
    VNodeComponent                    // 函数组件
    VNodeFragment                     // Fragment (不添加额外节点)
)
```

### Props

属性映射类型。

```go
type Props map[string]interface{}

// 方法
func (p Props) Get(key string) interface{}
func (p Props) Set(key string, value interface{}) Props
func (p Props) GetString(key string) string
func (p Props) GetInt(key string) int
func (p Props) GetBool(key string) bool
func (p Props) GetFunc(key string) func()
func (p Props) Merge(other Props) Props
func (p Props) Clone() Props
```

---

## 组件 VNode

### ComponentVNode

函数组件 VNode 类型。

```go
type ComponentVNode struct {
    // 包含 *ElementVNode 的嵌入字段
    // 有 name, fn, fnWithProps, props, key, style 等字段
}
```

#### 创建组件

```go
// 无 Props 组件
func Component(name string, fn ComponentFunc) *ComponentBuilder

// 带 Props 组件
func ComponentWithProps(name string, fn ComponentFuncWithProps) *ComponentBuilder
```

#### ComponentBuilder

```go
type ComponentBuilder struct {
    node *ComponentVNode
}

// 方法
func (b *ComponentBuilder) Props(p Props) *ComponentBuilder
func (b *ComponentBuilder) Prop(key string, value interface{}) *ComponentBuilder
func (b *ComponentBuilder) Key(key string) *ComponentBuilder
func (b *ComponentBuilder) Build() VNode
```

#### 示例

```go
// 定义组件函数
func Greeting() ui.VNode {
    return ui.Text("Hello, World!")
}

// 使用组件
ui.Component("Greeting", Greeting).Build()

// 带 key
ui.Component("ListItem", ListItem).Key("item-1").Build()

// 带 Props
ui.ComponentWithProps("UserCard", UserCard).
    Prop("name", "Alice").
    Prop("age", 30).
    Build()
```

---

## 组件实例

### ComponentInstance

持久化的组件实例接口。

```go
type ComponentInstance interface {
    Key() string
    SetKey(key string)
    SetProps(props Props) bool
    GetProps() Props
    GetState() map[string]interface{}
    SetState(key string, value interface{})
    GetContext() *ComponentContext
    Render() VNode
    OnMount()
    OnUpdate(newProps, oldProps Props) bool
    OnUnmount()
    MarkDirty()
    IsDirty() bool
}
```

### BaseComponentInstance

基础组件实例实现。

```go
type BaseComponentInstance struct {
    key      string
    props    Props
    context  *ComponentContext
    fn       ComponentFunc
    fnWithProps ComponentFuncWithProps
    dirty    bool
    mounted  bool
}

// 创建实例
func NewBaseComponentInstance(key string, fn ComponentFunc) *BaseComponentInstance
func NewBaseComponentInstanceWithProps(key string, fn ComponentFuncWithProps, props Props) *BaseComponentInstance
```

#### 实例生命周期

```go
// 1. 创建实例
inst := ui.NewBaseComponentInstance("my-component", MyComponentFunc)

// 2. 挂载
inst.OnMount()

// 3. 更新 props
inst.SetProps(newProps)

// 4. 渲染
vnode := inst.Render()

// 5. 卸载
inst.OnUnmount()
```

---

## 元素 VNode

### ElementVNode

标准元素节点。

```go
// 创建
func NewElement(tag string) *ElementVNode

// Builder 模式
func Element(tag string) *ElementBuilder
```

### TextVNode

文本节点。

```go
// 创建
func NewText(content string) *TextVNode

// Builder 模式
func Text(content string) *TextBuilder

// 示例
ui.Text("Hello").FgColor("blue").Bold(true).Build()
```

### FragmentVNode

Fragment 节点，不添加额外 DOM 节点。

```go
func Fragment(children ...VNode) *FragmentVNode
```

---

## 交互组件

### ButtonVNode

按钮组件。

```go
func Button(label string) *ButtonBuilder

// 方法
func (b *ButtonBuilder) Label(label string) *ButtonBuilder
func (b *ButtonBuilder) Style(s style.Style) *ButtonBuilder
func (b *ButtonBuilder) Disabled(disabled bool) *ButtonBuilder
func (b *ButtonBuilder) OnClick(fn func()) *ButtonBuilder
func (b *ButtonBuilder) OnMouseEnter(fn func()) *ButtonBuilder
func (b *ButtonBuilder) OnMouseLeave(fn func()) *ButtonBuilder

// 示例
ui.Button("Click me").
    Style(ui.NewStyle().Background("blue").Foreground("white")).
    OnClick(func() {
        fmt.Println("Clicked!")
    }).
    Build()
```

### InputVNode

输入框组件。

```go
func Input(placeholder string, value string, maxLength int) *InputVNode

// 方法
func (i *InputVNode) Placeholder(s string) *InputVNode
func (i *InputVNode) Value() string
func (i *InputVNode) SetValue(s string)
func (i *InputVNode) OnChange(fn func(string)) *InputVNode
func (i *InputVNode) OnSubmitFunc(fn func()) *InputVNode
func (i *InputVNode) InputType() InputType
```

### TextareaVNode

多行文本框组件。

```go
func Textarea(placeholder string, value string, rows, cols int) *TextareaVNode

// 方法
func (t *TextareaVNode) Value() string
func (t *TextareaVNode) SetValue(s string)
func (t *TextareaVNode) Rows() int
func (t *TextareaVNode) Cols() int
func (t *TextareaVNode) Placeholder() string
func (t *TextareaVNode) OnChange(fn func(string)) *TextareaVNode
```

### CheckboxVNode

复选框组件。

```go
func Checkbox(label string, checked bool) *CheckboxBuilder

// 方法
func (c *CheckboxBuilder) Label(label string) *CheckboxBuilder
func (c *CheckboxBuilder) Checked(checked bool) *CheckboxBuilder
func (c *CheckboxBuilder) OnChange(fn func(bool)) *CheckboxBuilder
```

### SelectVNode

下拉选择组件。

```go
func Select(options []SelectOption, selected int) *SelectBuilder

// 方法
func (s *SelectBuilder) Options(options []SelectOption) *SelectBuilder
func (s *SelectBuilder) Selected(selected int) *SelectBuilder
func (s *SelectBuilder) OnChange(fn func(interface{})) *SelectBuilder

// SelectOption 结构
type SelectOption struct {
    Label string
    Value interface{}
}
```

---

## 布局组件

### LayoutNode

弹性布局容器。

```go
// 创建
func NewLayout(direction LayoutDirection) *LayoutNode

// 方向
const (
    DirectionRow LayoutDirection = iota
    DirectionColumn
)

// 方法
func (l *LayoutNode) Padding(top, right, bottom, left int) *LayoutNode
func (l *LayoutNode) Gap(gap int) *LayoutNode
func (l *LayoutNode) Children(children ...VNode) *LayoutNode

// 示例
ui.NewLayout(ui.DirectionRow).
    Padding(10, 20, 10, 20).
    Gap(5).
    Children(
        ui.Text("First"),
        ui.Text("Second"),
    )
```

### GridVNode

网格布局容器。

```go
func NewGrid(columns int) *GridBuilder

// 方法
func (g *GridBuilder) Padding(top, right, bottom, left int) *GridBuilder
func (g *GridBuilder) Gap(row, col int) *GridBuilder
func (g *GridBuilder) Cells(cells ...*GridCell) *GridBuilder

// GridCell 结构
type GridCell struct {
    Child     VNode
    Row       int
    Col       int
    RowSpan   int
    ColSpan   int
}

// 示例
ui.NewGrid(3).Cells(
    ui.NewGridCell(ui.Text("Cell 1"), 0, 0, 1, 1),
    ui.NewGridCell(ui.Text("Cell 2"), 0, 1, 1, 1),
    ui.NewGridCell(ui.Text("Span"), 1, 0, 1, 2), // 跨2列
)
```

### AbsoluteVNode

绝对定位容器。

```go
func Absolute(x, y int, child VNode) *AbsoluteVNode

// 示例
ui.Absolute(10, 5, ui.Text(" positioned"))
```

---

## 构建器模式

大多数组件都支持链式调用构建：

```go
ui.Text("Hello").
    FgColor("blue").
    Bold(true).
    Background("black").
    Build()
```

---

## 相关文档

- [Hooks API](./hooks.md)
- [内存安全工具](./memory-safety.md)
- [组件开发指南](../guide/component-development-guide.md)
