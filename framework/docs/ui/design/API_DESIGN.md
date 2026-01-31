# Mint UI API 设计文档

**版本**: v1.0
**日期**: 2026-01-31

---

## 目录

1. [设计原则](#一设计原则)
2. [包结构](#二包结构)
3. [核心 API](#三核心-api)
4. [组件 API](#四组件-api)
5. [Hooks API](#五-hooks-api)
6. [样式 API](#六样式-api)
7. [事件 API](#七事件-api)
8. [布局 API](#八布局-api)
9. [工具 API](#九工具-api)
10. [类型定义](#十类型定义)

---

## 一、设计原则

### 1.1 声明式优先

```go
// ✅ 声明式（推荐）
func App() VNode {
    return ui.VStack(
        ui.Text("Hello").FgColor(color.Red),
        ui.Button("Click").OnClick(func() {}),
    )
}

// ❌ 命令式（避免）
text := display.NewText("Hello")
text.SetStyle(color.Red)
container.Add(text)
```

### 1.2 链式调用

```go
ui.Text("Hello")
    .FgColor(color.Red)
    .BgColor(color.Blue)
    .Bold(true)
```

### 1.3 类型安全

```go
// 强类型 Props
type ButtonProps struct {
    Label    string
    OnClick  func()
    Disabled bool
}

func Button(props ButtonProps) VNode {
    // ...
}
```

### 1.4 最小化 API

```go
// ✅ 简洁
ui.Text("Hello")

// ❌ 冗余
ui.CreateTextBuilder().WithContent("Hello").Build()
```

---

## 二、包结构

### 2.1 包组织

```
github.com/wwsheng009/mint/
├── ui/                    # 主 API 包（对外暴露）
│   ├── app.go
│   ├── vnode.go
│   ├── builder.go
│   ├── components.go
│   ├── layout.go
│   ├── hooks.go
│   ├── events.go
│   └── style.go
│
├── framework/             # 框架层（高级用户）
│   ├── reconciler/
│   ├── hooks/
│   ├── components/
│   └── render/
│
└── runtime/               # 运行时层（内部）
    ├── layout/
    ├── event/
    └── paint/
```

### 2.2 导入约定

```go
// 标准导入（大多数应用）
import "github.com/wwsheng009/mint/ui"

// 高级用法
import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/framework/hooks"
    "github.com/wwsheng009/mint/framework/components"
)
```

---

## 三、核心 API

### 3.1 应用入口

```go
package ui

// Run 运行应用
func Run(app func() VNode) error

// RunWithOptions 运行应用（带选项）
func RunWithOptions(app func() VNode, opts ...Option) error

// Option 运行时选项
type Option func(*Options)

// Options 可配置项
type Options struct {
    Width          int           // 窗口宽度
    Height         int           // 窗口高度
    Title          string        // 窗口标题
    FPS            int           // 帧率限制
    EnableDevTools bool          // 启用 DevTools
    Theme          theme.Theme   // 主题
    KeyMap         input.KeyMap  // 快捷键
}

// WithWidth 设置宽度
func WithWidth(width int) Option

// WithHeight 设置高度
func WithHeight(height int) Option

// WithTitle 设置标题
func WithTitle(title string) Option

// WithFPS 设置帧率
func WithFPS(fps int) Option

// WithDevTools 启用 DevTools
func WithDevTools(enabled bool) Option

// WithTheme 设置主题
func WithTheme(t theme.Theme) Option
```

#### 使用示例

```go
package main

import "github.com/wwsheng009/mint/ui"

func main() {
    ui.Run(App,
        ui.WithWidth(80),
        ui.WithHeight(24),
        ui.WithTitle("My App"),
        ui.WithFPS(60),
    )
}

func App() ui.VNode {
    return ui.Text("Hello, Mint UI!")
}
```

---

### 3.2 VNode 接口

```go
package ui

// VNode 虚拟节点接口
type VNode interface {
    // 类型标识
    Type() VNodeType

    // Props
    Props() Props
    SetProps(Props)

    // 子节点
    Children() []VNode
    SetChildren([]VNode)

    // Key（用于 Diff）
    Key() string
    SetKey(string)
}

// VNodeType 节点类型
type VNodeType int

const (
    VNodeElement VNodeType = iota
    VNodeText
    VNodeComponent
    VNodeFragment
)
```

---

### 3.3 Props 系统

```go
package ui

// Props 属性集合
type Props map[string]interface{}

// Get 获取属性
func (p Props) Get(key string) interface{}

// GetString 获取字符串属性
func (p Props) GetString(key string) string

// GetInt 获取整型属性
func (p Props) GetInt(key string) int

// GetFunc 获取函数属性
func (p Props) GetFunc(key string) func()

// Set 设置属性
func (p Props) Set(key string, value interface{}) Props

// Merge 合并属性
func (p Props) Merge(other Props) Props
```

---

## 四、组件 API

### 4.1 基础组件

#### Text

```go
package ui

// Text 创建文本组件
func Text(content string) *TextBuilder

// TextBuilder 文本构建器
type TextBuilder struct {
    node *elementNode
}

// Content 设置文本内容
func (b *TextBuilder) Content(content string) *TextBuilder

// FgColor 设置前景色
func (b *TextBuilder) FgColor(c color.Color) *TextBuilder

// BgColor 设置背景色
func (b *TextBuilder) BgColor(c color.Color) *TextBuilder

// Bold 设置粗体
func (b *TextBuilder) Bold(v bool) *TextBuilder

// Italic 设置斜体
func (b *TextBuilder) Italic(v bool) *TextBuilder

// Underline 设置下划线
func (b *TextBuilder) Underline(v bool) *TextBuilder

// Align 设置对齐
func (b *TextBuilder) Align(align TextAlign) *TextBuilder

// MaxLines 设置最大行数
func (b *TextBuilder) MaxLines(n int) *TextBuilder

// Build 构建 VNode
func (b *TextBuilder) Build() VNode
```

#### 使用示例

```go
// 简单文本
ui.Text("Hello, World!")

// 链式调用
ui.Text("Hello").
    FgColor(color.Red).
    Bold(true).
    Align(ui.AlignCenter)

// 多行文本
ui.Text("Line 1\nLine 2\nLine 3")

// 限制行数
ui.Text(longText).MaxLines(3)
```

---

#### Icon

```go
package ui

// Icon 创建图标组件
func Icon(name string) *IconBuilder

// IconBuilder 图标构建器
type IconBuilder struct {
    node *elementNode
}

// Name 设置图标名称
func (b *IconBuilder) Name(name string) *IconBuilder

// FgColor 设置颜色
func (b *IconBuilder) FgColor(c color.Color) *IconBuilder

// Size 设置图标大小
func (b *IconBuilder) Size(size int) *IconBuilder

// Build 构建 VNode
func (b *IconBuilder) Build() VNode
```

---

#### Separator

```go
package ui

// Separator 创建分隔符
func Separator() *SeparatorBuilder

// SeparatorBuilder 分隔符构建器
type SeparatorBuilder struct {
    node *elementNode
}

// Vertical 设置为垂直方向
func (b *SeparatorBuilder) Vertical(v bool) *SeparatorBuilder

// Style 设置样式
func (b *SeparatorBuilder) Style(s SeparatorStyle) *SeparatorBuilder

// Build 构建 VNode
func (b *SeparatorBuilder) Build() VNode

// SeparatorStyle 分隔符样式
type SeparatorStyle int

const (
    SeparatorSingle SeparatorStyle = iota
    SeparatorDouble
    SeparatorDashed
    SeparatorDotted
)
```

---

### 4.2 表单组件

#### Input

```go
package ui

// Input 创建输入框
func Input(placeholder string) *InputBuilder

// InputBuilder 输入框构建器
type InputBuilder struct {
    node *elementNode
}

// Placeholder 设置占位符
func (b *InputBuilder) Placeholder(text string) *InputBuilder

// Value 设置值（受控）
func (b *InputBuilder) Value(v string) *InputBuilder

// DefaultValue 设置默认值（非受控）
func (b *InputBuilder) DefaultValue(v string) *InputBuilder

// OnChange 值变化回调
func (b *InputBuilder) OnChange(fn func(string)) *InputBuilder

// OnSubmit 提交回调
func (b *InputBuilder) OnSubmit(fn func(string)) *InputBuilder

// Password 设置为密码模式
func (b *InputBuilder) Password(v bool) *InputBuilder

// MaxLength 设置最大长度
func (b *InputBuilder) MaxLength(n int) *InputBuilder

// Disabled 禁用
func (b *InputBuilder) Disabled(v bool) *InputBuilder

// Build 构建 VNode
func (b *InputBuilder) Build() VNode
```

#### 使用示例

```go
// 非受控组件
ui.Input("Enter text...").
    OnChange(func(value string) {
        fmt.Println("Input:", value)
    })

// 受控组件
func SearchBox() VNode {
    query, setQuery := ui.UseState("")

    return ui.Input("Search...").
        Value(query).
        OnChange(func(value string) {
            setQuery(value)
        })
}

// 密码输入
ui.Input("Password...").
    Password(true)

// 限制长度
ui.Input("Username...").
    MaxLength(20)
```

---

#### CheckBox

```go
package ui

// CheckBox 创建复选框
func CheckBox(label string) *CheckBoxBuilder

// CheckBoxBuilder 复选框构建器
type CheckBoxBuilder struct {
    node *elementNode
}

// Checked 设置选中状态（受控）
func (b *CheckBoxBuilder) Checked(v bool) *CheckBoxBuilder

// DefaultChecked 设置默认选中（非受控）
func (b *CheckBoxBuilder) DefaultChecked(v bool) *CheckBoxBuilder

// OnChange 变化回调
func (b *CheckBoxBuilder) OnChange(fn func(bool)) *CheckBoxBuilder

// Disabled 禁用
func (b *CheckBoxBuilder) Disabled(v bool) *CheckBoxBuilder

// Build 构建 VNode
func (b *CheckBoxBuilder) Build() VNode
```

---

#### Select

```go
package ui

// Select 创建下拉选择
func Select(options []string) *SelectBuilder

// SelectBuilder 下拉选择构建器
type SelectBuilder struct {
    node *elementNode
}

// Options 设置选项
func (b *SelectBuilder) Options(opts []string) *SelectBuilder

// Selected 设置选中项（受控）
func (b *SelectBuilder) Selected(index int) *SelectBuilder

// DefaultSelected 设置默认选中（非受控）
func (b *SelectBuilder) DefaultSelected(index int) *SelectBuilder

// OnChange 变化回调
func (b *SelectBuilder) OnChange(fn func(int)) *SelectBuilder

// Disabled 禁用
func (b *SelectBuilder) Disabled(v bool) *SelectBuilder

// Build 构建 VNode
func (b *SelectBuilder) Build() VNode
```

---

### 4.3 按钮组件

#### Button

```go
package ui

// Button 创建按钮
func Button(label string) *ButtonBuilder

// ButtonBuilder 按钮构建器
type ButtonBuilder struct {
    node *elementNode
}

// Label 设置标签
func (b *ButtonBuilder) Label(label string) *ButtonBuilder

// OnClick 点击事件
func (b *ButtonBuilder) OnClick(fn func()) *ButtonBuilder

// Disabled 禁用
func (b *ButtonBuilder) Disabled(v bool) *ButtonBuilder

// Variant 设置样式变体
func (b *ButtonBuilder) Variant(v ButtonVariant) *ButtonBuilder

// Size 设置大小
func (b *ButtonBuilder) Size(s ButtonSize) *ButtonBuilder

// Build 构建 VNode
func (b *ButtonBuilder) Build() VNode

// ButtonVariant 按钮变体
type ButtonVariant int

const (
    ButtonVariantDefault ButtonVariant = iota
    ButtonVariantPrimary
    ButtonVariantSecondary
    ButtonVariantDanger
    ButtonVariantSuccess
)

// ButtonSize 按钮大小
type ButtonSize int

const (
    ButtonSizeSmall ButtonSize = iota
    ButtonSizeMedium
    ButtonSizeLarge
)
```

#### 使用示例

```go
// 基础按钮
ui.Button("Click Me").
    OnClick(func() {
        fmt.Println("Clicked!")
    })

// 不同变体
ui.Button("Primary").Variant(ui.ButtonVariantPrimary)
ui.Button("Danger").Variant(ui.ButtonVariantDanger)

// 不同大小
ui.Button("Small").Size(ui.ButtonSizeSmall)
ui.Button("Large").Size(ui.ButtonSizeLarge)

// 禁用状态
ui.Button("Disabled").Disabled(true)
```

---

### 4.4 数据展示组件

#### List

```go
package ui

// List 创建列表
func List(items []string) *ListBuilder

// ListBuilder 列表构建器
type ListBuilder struct {
    node *elementNode
}

// Items 设置列表项
func (b *ListBuilder) Items(items []string) *ListBuilder

// Selected 设置选中项（受控）
func (b *ListBuilder) Selected(index int) *ListBuilder

// DefaultSelected 设置默认选中（非受控）
func (b *ListBuilder) DefaultSelected(index int) *ListBuilder

// OnSelect 选中回调
func (b *ListBuilder) OnSelect(fn func(int)) *ListBuilder

// MultiSelect 多选模式
func (b *ListBuilder) MultiSelect(v bool) *ListBuilder

// Build 构建 VNode
func (b *ListBuilder) Build() VNode
```

---

#### Table

```go
package ui

// Table 创建表格
func Table(headers []string) *TableBuilder

// TableBuilder 表格构建器
type TableBuilder struct {
    node *elementNode
}

// Headers 设置表头
func (b *TableBuilder) Headers(headers []string) *TableBuilder

// Rows 设置行数据
func (b *TableBuilder) Rows(rows [][]string) *TableBuilder

// OnSort 排序回调
func (b *TableBuilder) OnSort(fn func(int, bool)) *TableBuilder

// Sortable 可排序
func (b *TableBuilder) Sortable(v bool) *TableBuilder

// Build 构建 VNode
func (b *TableBuilder) Build() VNode
```

---

### 4.5 反馈组件

#### ProgressBar

```go
package ui

// ProgressBar 创建进度条
func ProgressBar(value float64) *ProgressBarBuilder

// ProgressBarBuilder 进度条构建器
type ProgressBarBuilder struct {
    node *elementNode
}

// Value 设置进度值 (0-1)
func (b *ProgressBarBuilder) Value(v float64) *ProgressBarBuilder

// Indeterminate 不确定进度
func (b *ProgressBarBuilder) Indeterminate(v bool) *ProgressBarBuilder

// ShowValue 显示百分比
func (b *ProgressBarBuilder) ShowValue(v bool) *ProgressBarBuilder

// Build 构建 VNode
func (b *ProgressBarBuilder) Build() VNode
```

---

#### Toast

```go
package ui

// Toast 显示提示（需要在应用中配置 ToastProvider）
func Toast(message string) *ToastBuilder

// ToastBuilder 提示构建器
type ToastBuilder struct {
    node *elementNode
}

// Message 设置消息
func (b *ToastBuilder) Message(msg string) *ToastBuilder

// Variant 设置类型
func (b *ToastBuilder) Variant(v ToastVariant) *ToastBuilder

// Duration 设置持续时间
func (b *ToastBuilder) Duration(d time.Duration) *ToastBuilder

// Build 构建 VNode
func (b *ToastBuilder) Build() VNode

// ToastVariant 提示类型
type ToastVariant int

const (
    ToastVariantDefault ToastVariant = iota
    ToastVariantSuccess
    ToastVariantError
    ToastVariantWarning
    ToastVariantInfo
)
```

---

## 五、Hooks API

### 5.1 useState

```go
package ui

// useState 状态 Hook
func useState(initial interface{}) (value interface{}, setValue func(interface{}))

// UseStateInt 整型状态（类型安全）
func UseStateInt(initial int) (int, func(int))

// UseStateString 字符串状态
func UseStateString(initial string) (string, func(string))

// UseStateBool 布尔状态
func UseStateBool(initial bool) (bool, func(bool))
```

#### 使用示例

```go
func Counter() VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("Increment").OnClick(func() {
            setCount(count + 1)
        }),
    )
}
```

---

### 5.2 useEffect

```go
package ui

// useEffect 副作用 Hook
func useEffect(effect func(), deps []interface{})

// useEffectWithCleanup 带清理的副作用
func useEffectWithCleanup(effect func() func(), deps []interface{})
```

#### 使用示例

```go
func Timer() VNode {
    count, setCount := ui.UseStateInt(0)

    ui.UseEffect(func() {
        // 启动定时器
        ticker := time.NewTicker(time.Second)
        done := make(chan bool)

        go func() {
            for {
                select {
                case <-ticker.C:
                    setCount(count + 1)
                case <-done:
                    return
                }
            }
        }()

        // 清理函数
        return func() {
            ticker.Stop()
            close(done)
        }
    }, nil) // 空依赖 = 只执行一次

    return ui.Text(fmt.Sprintf("Time: %d", count))
}
```

---

### 5.3 useContext

```go
package ui

// Context 上下文类型
type Context struct {
    value interface{}
}

// useContext 使用上下文
func useContext(ctx *Context) interface{}

// createContext 创建上下文
func createContext(value interface{}) *Context

// Provider 上下文提供者
func Provider(ctx *Context, value interface{}, children ...VNode) VNode
```

#### 使用示例

```go
// 创建主题上下文
var ThemeContext = ui.CreateContext(theme.Light)

// 提供主题
func App() VNode {
    currentTheme, setTheme := ui.UseState(theme.Light)

    return ui.Provider(ThemeContext, currentTheme,
        ui.Button("Toggle Theme").OnClick(func() {
            if currentTheme == theme.Light {
                setTheme(theme.Dark)
            } else {
                setTheme(theme.Light)
            }
        }),
        ThemedButton(),
    )
}

// 消费主题
func ThemedButton() VNode {
    t := ui.UseContext(ThemeContext).(theme.Theme)

    return ui.Button("Themed").
        FgColor(t.Primary)
}
```

---

### 5.4 useMemo & useCallback

```go
package ui

// useMemo 缓存计算结果
func useMemo(fn func() interface{}, deps []interface{}) interface{}

// useCallback 缓存函数
func useCallback(fn func(), deps []interface{}) func()
```

#### 使用示例

```go
func ExpensiveComponent() VNode {
    items, setItems := ui.UseState([]Item{})

    // 缓存过滤结果
    filteredItems := ui.UseMemo(func() interface{} {
        return filterItems(items)
    }, []interface{}{items})

    // 缓存回调函数
    handleItemClick := ui.UseCallback(func(id string) {
        // 处理点击
    }, nil)

    return ui.List(filteredItems.([]Item)).
        OnSelect(handleItemClick)
}
```

---

### 5.5 useRef

```go
package ui

// useRef 创建引用
func useRef(initial interface{}) *Ref

// Ref 引用类型
type Ref struct {
    current interface{}
}

// Get 获取当前值
func (r *Ref) Get() interface{}

// Set 设置值
func (r *Ref) Set(v interface{})
```

#### 使用示例

```go
func InputWithRef() VNode {
    inputRef := ui.UseRef(nil)

    ui.UseEffect(func() {
        // 聚焦输入框
        if input := inputRef.Get().(*Input); input != nil {
            input.Focus()
        }
    }, nil)

    return ui.Input("Focus me...").
        Ref(inputRef)
}
```

---

## 六、样式 API

### 6.1 颜色

```go
package color

// Color 颜色类型
type Color struct {
    R, G, B uint8
    A       uint8
}

// RGB 创建 RGB 颜色
func RGB(r, g, b uint8) Color

// RGBA 创建 RGBA 颜色
func RGBA(r, g, b, a uint8) Color

// Hex 从十六进制创建颜色
func Hex(s string) Color

// ANSI 从 ANSI 色号创建颜色
func ANSI(code uint8) Color

// 预定义颜色
var (
    Black   Color
    Red     Color
    Green   Color
    Yellow  Color
    Blue    Color
    Magenta Color
    Cyan    Color
    White   Color
)
```

---

### 6.2 样式

```go
package ui

// Style 样式类型
type Style struct {
    FgColor    *color.Color
    BgColor    *color.Color
    Bold       bool
    Italic     bool
    Underline  bool
    Strikethrough bool
    Reverse    bool
}

// Style 创建样式
func Style() *StyleBuilder

// StyleBuilder 样式构建器
type StyleBuilder struct {
    style Style
}

// Fg 前景色
func (b *StyleBuilder) Fg(c color.Color) *StyleBuilder

// Bg 背景色
func (b *StyleBuilder) Bg(c color.Color) *StyleBuilder

// Bold 粗体
func (b *StyleBuilder) Bold() *StyleBuilder

// Italic 斜体
func (b *StyleBuilder) Italic() *StyleBuilder

// Underline 下划线
func (b *StyleBuilder) Underline() *StyleBuilder

// Reverse 反色
func (b *StyleBuilder) Reverse() *StyleBuilder

// Build 构建样式
func (b *StyleBuilder) Build() Style
```

---

## 七、事件 API

### 7.1 事件类型

```go
package ui

// Event 事件接口
type Event interface {
    Type() EventType
    Target() VNode
    CurrentTarget() VNode
    StopPropagation()
    PreventDefault()
}

// EventType 事件类型
type EventType string

const (
    EventTypeClick    EventType = "click"
    EventTypeChange   EventType = "change"
    EventTypeFocus    EventType = "focus"
    EventTypeBlur     EventType = "blur"
    EventTypeKeyPress EventType = "keypress"
    EventTypeSubmit   EventType = "submit"
)

// MouseEvent 鼠标事件
type MouseEvent struct {
    X, Y   int
    Button MouseButton
}

// KeyEvent 键盘事件
type KeyEvent struct {
    Key   rune
    Mod   KeyMod
}

// KeyMod 键盘修饰符
type KeyMod int

const (
    ModNone KeyMod = 0
    ModCtrl KeyMod = 1 << iota
    ModAlt
    ModShift
)
```

---

### 7.2 事件绑定

```go
package ui

// OnClick 点击事件
func (b *ElementBuilder) OnClick(fn func(Event)) *ElementBuilder

// OnClickCapture 捕获阶段点击
func (b *ElementBuilder) OnClickCapture(fn func(Event)) *ElementBuilder

// OnChange 变化事件
func (b *ElementBuilder) OnChange(fn func(Event)) *ElementBuilder

// OnFocus 获得焦点
func (b *ElementBuilder) OnFocus(fn func(Event)) *ElementBuilder

// OnBlur 失去焦点
func (b *ElementBuilder) OnBlur(fn func(Event)) *ElementBuilder

// OnKeyPress 按键事件
func (b *ElementBuilder) OnKeyPress(fn func(KeyEvent)) *ElementBuilder

// OnSubmit 提交事件
func (b *ElementBuilder) OnSubmit(fn func(Event)) *ElementBuilder
```

---

## 八、布局 API

### 8.1 基础布局

```go
package ui

// HStack 水平布局
func HStack(children ...VNode) *LayoutBuilder

// VStack 垂直布局
func VStack(children ...VNode) *LayoutBuilder

// Stack 通用堆叠布局
func Stack(direction Direction, children ...VNode) *LayoutBuilder

// Direction 布局方向
type Direction int

const (
    DirectionRow Direction = iota
    DirectionColumn
)
```

---

### 8.2 布局修饰符

```go
package ui

// LayoutBuilder 布局构建器
type LayoutBuilder struct {
    node *elementNode
}

// Align 设置主轴对齐
func (b *LayoutBuilder) Align(a Align) *LayoutBuilder

// AlignCross 设置交叉轴对齐
func (b *LayoutBuilder) AlignCross(a Align) *LayoutBuilder

// Gap 设置间距
func (b *LayoutBuilder) Gap(n int) *LayoutBuilder

// Padding 设置内边距
func (b *LayoutBuilder) Padding(top, right, bottom, left int) *LayoutBuilder

// Width 设置宽度
func (b *LayoutBuilder) Width(n int) *LayoutBuilder

// Height 设置高度
func (b *LayoutBuilder) Height(n int) *LayoutBuilder

// Flex 设置弹性
func (b *LayoutBuilder) Flex(n int) *LayoutBuilder

// Build 构建 VNode
func (b *LayoutBuilder) Build() VNode

// Align 对齐方式
type Align int

const (
    AlignStart Align = iota
    AlignCenter
    AlignEnd
    AlignSpaceBetween
    AlignSpaceAround
    AlignSpaceEvenly
)
```

#### 使用示例

```go
// 水平排列，居中对齐
ui.HStack(
    ui.Text("Left"),
    ui.Text("Center"),
    ui.Text("Right"),
).Align(ui.AlignCenter).Gap(4)

// 垂直排列，两端对齐
ui.VStack(
    ui.Text("Top"),
    ui.Text("Middle"),
    ui.Text("Bottom"),
).Align(ui.AlignSpaceBetween)

// 弹性布局
ui.HStack(
    ui.Text("Left").Flex(1),
    ui.Text("Middle").Flex(2),
    ui.Text("Right").Flex(1),
)
```

---

### 8.3 容器

```go
package ui

// Box 创建容器
func Box() *BoxLayoutBuilder

// BoxLayoutBuilder 容器构建器
type BoxLayoutBuilder struct {
    node *elementNode
}

// Border 边框
func (b *BoxLayoutBuilder) Border(v bool) *BoxLayoutBuilder

// BorderStyle 边框样式
func (b *BoxLayoutBuilder) BorderStyle(s BorderStyle) *BoxLayoutBuilder

// Padding 内边距
func (b *BoxLayoutBuilder) Padding(n int) *BoxLayoutBuilder

// Background 背景色
func (b *BoxLayoutBuilder) Background(c color.Color) *BoxLayoutBuilder

// Child 子组件
func (b *BoxLayoutBuilder) Child(child VNode) *BoxLayoutBuilder

// Build 构建 VNode
func (b *BoxLayoutBuilder) Build() VNode

// BorderStyle 边框样式
type BorderStyle int

const (
    BorderStyleSingle BorderStyle = iota
    BorderStyleDouble
    BorderStyleRounded
    BorderStyleThick
)
```

---

### 8.4 Spacer

```go
package ui

// Spacer 创建弹性空间
func Spacer() *SpacerBuilder

// SpacerBuilder 弹性空间构建器
type SpacerBuilder struct {
    node *elementNode
}

// Flex 设置弹性比例
func (b *SpacerBuilder) Flex(n int) *SpacerBuilder

// Build 构建 VNode
func (b *SpacerBuilder) Build() VNode
```

#### 使用示例

```go
// 左右两端对齐
ui.HStack(
    ui.Text("Left"),
    ui.Spacer().Flex(1),
    ui.Text("Right"),
)

// 自定义弹性比例
ui.HStack(
    ui.Text("Small").Flex(1),
    ui.Spacer().Flex(2),
    ui.Text("Large").Flex(3),
)
```

---

## 九、工具 API

### 9.1 Key

```go
package ui

// Key 设置组件 key
func Key(key string, child VNode) VNode
```

#### 使用示例

```go
func List(items []Item) VNode {
    children := make([]VNode, len(items))
    for i, item := range items {
        children[i] = ui.Key(item.ID, ui.Text(item.Name))
    }
    return ui.VStack(children...)
}
```

---

### 9.2 Fragment

```go
package ui

// Fragment 创建片段（不添加额外节点）
func Fragment(children ...VNode) VNode
```

#### 使用示例

```go
// 避免额外容器
ui.VStack(
    ui.Text("Title"),
    ui.Fragment(
        ui.Text("Line 1"),
        ui.Text("Line 2"),
        ui.Text("Line 3"),
    ),
    ui.Text("Footer"),
)
```

---

### 9.3 条件渲染

```go
package ui

// If 条件渲染
func If(condition bool, then VNode, elseVNode ...VNode) VNode

// When 条件渲染（简洁版）
func When(condition bool, node VNode) VNode
```

#### 使用示例

```go
func LoggedInUser(user *User) VNode {
    return ui.VStack(
        ui.If(user != nil,
            ui.Text("Welcome, " + user.Name),
            ui.Text("Please login"),
        ),
        ui.When(user.IsAdmin,
            ui.Button("Admin Panel"),
        ),
    )
}
```

---

### 9.4 列表渲染

```go
package ui

// For 列表渲染
func For[T any](items []T, fn func(int, T) VNode) VNode
```

#### 使用示例

```go
func TodoList(todos []Todo) VNode {
    return ui.VStack(
        ui.For(todos, func(i int, todo Todo) ui.VNode {
            return ui.HStack(
                ui.Text(todo.Title),
                ui.When(todo.Done,
                    ui.Text("✓").FgColor(color.Green),
                ),
            ).Key(todo.ID)
        }),
    )
}
```

---

## 十、类型定义

### 10.1 核心类型

```go
package ui

// VNode 虚拟节点
type VNode interface {
    Type() VNodeType
    Props() Props
    Children() []VNode
    Key() string
    SetKey(string)
}

// ComponentFunc 组件函数类型
type ComponentFunc func() VNode

// ComponentFuncWithProps 带 Props 的组件函数
type ComponentFuncWithProps func(Props) VNode
```

---

### 10.2 组件 Props 类型

```go
package ui

// TextProps 文本属性
type TextProps struct {
    Content  string
    FgColor  color.Color
    BgColor  color.Color
    Bold     bool
    Italic   bool
    Underline bool
    Align    TextAlign
    MaxLines int
}

// ButtonProps 按钮属性
type ButtonProps struct {
    Label    string
    OnClick  func()
    Disabled bool
    Variant  ButtonVariant
    Size     ButtonSize
}

// InputProps 输入框属性
type InputProps struct {
    Placeholder  string
    Value        string
    DefaultValue string
    OnChange     func(string)
    OnSubmit     func(string)
    Password     bool
    MaxLength    int
    Disabled     bool
}
```

---

### 10.3 事件类型

```go
package ui

// EventHandler 事件处理器
type EventHandler func(Event)

// ClickHandler 点击处理器
type ClickHandler func(MouseEvent)

// ChangeHandler 变化处理器
type ChangeHandler func(interface{})

// KeyHandler 按键处理器
type KeyHandler func(KeyEvent)

// SubmitHandler 提交处理器
type SubmitHandler func(Event)
```

---

## 十一、最佳实践

### 11.1 组件定义

```go
// ✅ 推荐：函数组件
func MyComponent(props MyComponentProps) VNode {
    return ui.VStack(
        ui.Text(props.Title),
        // ...
    )
}

// ❌ 避免：在渲染中创建闭包
func BadComponent() VNode {
    count, _ := ui.UseState(0)
    return ui.Button("Click").OnClick(func() {
        // 这个闭包会捕获 count 的旧值！
        fmt.Println(count)
    })
}
```

---

### 11.2 依赖数组

```go
// ✅ 正确：依赖数组包含所有使用的外部变量
ui.UseEffect(func() {
    subscribe(userID)
    return func() { unsubscribe(userID) }
}, []interface{}{userID})

// ❌ 错误：遗漏依赖
ui.UseEffect(func() {
    subscribe(userID)  // userID 变化时不会重新执行
    return func() { unsubscribe(userID) }
}, nil)
```

---

### 11.3 Key 使用

```go
// ✅ 正确：使用稳定的 ID
ui.For(items, func(i int, item Item) VNode {
    return ui.Text(item.Name).Key(item.ID)
})

// ❌ 错误：使用索引作为 Key
ui.For(items, func(i int, item Item) VNode {
    return ui.Text(item.Name).Key(strconv.Itoa(i))
})
```

---

**文档结束**

**版本历史**:
- v1.0 (2026-01-31): 初始版本
