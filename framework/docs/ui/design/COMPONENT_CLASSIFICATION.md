# Mint UI 组件分类重组方案

**版本**: v1.0
**日期**: 2026-01-31

---

## 一、当前组件分类分析

### 1.1 现有目录结构

```
framework/
├── component/          # 组件核心（基础设施）
├── display/            # 显示组件
│   ├── text.go        # Text
│   ├── list.go        # List
│   └── table.go       # Table
├── input/              # 输入组件
│   └── textinput.go   # TextInput
├── interactive/        # 交互组件
│   └── button.go      # Button
├── layout/             # 布局组件
│   ├── box.go         # Box
│   └── flex.go        # Flex
├── form/               # 表单处理
└── widget/             # 小部件（空目录）
```

### 1.2 分类问题分析

| 问题 | 说明 | 影响 |
|------|------|------|
| **界限模糊** | Button 在 `interactive/`，TextInput 在 `input/`，两者都是可交互组件 | 用户难以查找 |
| **分类不一致** | Text/List/Table 是纯显示，但 List/Table 也可交互 | 分类逻辑混乱 |
| **widget/ 空置** | 创建了 `widget/` 但未使用 | 目录冗余 |
| **组件分散** | 功能相近的组件分散在不同目录 | 维护困难 |
| **缺少常用组件** | 无 Checkbox, Radio, Select, Modal, Tabs 等 | 功能不完整 |

### 1.3 现有组件清单

| 组件 | 当前位置 | 类型 | 说明 |
|------|---------|------|------|
| Text | display/ | 显示 | 纯文本显示 |
| List | display/ | 显示 | 列表显示（可滚动） |
| Table | display/ | 显示 | 表格显示 |
| Button | interactive/ | 交互 | 按钮组件 |
| TextInput | input/ | 输入 | 文本输入框 |
| Box | layout/ | 布局 | 容器组件 |
| Flex | layout/ | 布局 | Flex 布局 |
| VirtualList | component/ | 显示 | 虚拟滚动列表 |

---

## 二、推荐分类方案

### 2.1 新的目录结构

```
framework/
├── components/              # 🔵 统一的组件目录（新建）
│   ├── basic/              # 基础组件
│   │   ├── text.go         # Text
│   │   ├── icon.go         # Icon
│   │   ├── separator.go    # Separator
│   │   └── spacer.go       # Spacer
│   │
│   ├── layout/             # 布局组件
│   │   ├── box.go          # Box
│   │   ├── flex.go         # Flex
│   │   ├── stack.go        # Stack (HStack/VStack)
│   │   ├── grid.go         # Grid
│   │   └── overlay.go      # Overlay
│   │
│   ├── form/              # 表单组件
│   │   ├── input.go        # TextInput
│   │   ├── textarea.go     # TextArea
│   │   ├── checkbox.go     # CheckBox
│   │   ├── radio.go        # Radio
│   │   ├── select.go       # Select
│   │   ├── switch.go       # Switch
│   │   ├── slider.go       # Slider
│   │   ├── field.go        # Field (表单字段包装)
│   │   └── form.go         # Form (表单容器)
│   │
│   ├── button/            # 按钮组件
│   │   ├── button.go       # Button
│   │   ├── icon_button.go  # IconButton
│   │   └── button_group.go # ButtonGroup
│   │
│   ├── feedback/          # 反馈组件
│   │   ├── progress.go     # ProgressBar
│   │   ├── spinner.go      # Spinner
│   │   ├── toast.go        # Toast
│   │   ├── alert.go        # Alert
│   │   └── badge.go        # Badge
│   │
│   ├── data/              # 数据展示组件
│   │   ├── list.go         # List
│   │   ├── table.go        # Table
│   │   ├── tree.go         # Tree
│   │   ├── virtual_list.go # VirtualList
│   │   └── calendar.go     # Calendar
│   │
│   ├── navigation/        # 导航组件
│   │   ├── tabs.go         # Tabs
│   │   ├── menu.go         # Menu
│   │   ├── sidebar.go      # Sidebar
│   │   └── breadcrumb.go   # Breadcrumb
│   │
│   ├── overlay/           # 覆盖层组件
│   │   ├── modal.go        # Modal
│   │   ├── dialog.go       # Dialog
│   │   ├── dropdown.go     # Dropdown
│   │   └── tooltip.go      # Tooltip
│   │
│   ├── container/         # 容器组件
│   │   ├── panel.go        # Panel
│   │   ├── split.go        # SplitPane
│   │   ├── scroll.go       # ScrollArea
│   │   └── resizable.go    # Resizable
│   │
│   └── advanced/          # 高级组件
│       ├── editor.go       # TextEditor
│       ├── terminal.go     # Terminal
│       ├── chart.go        # Chart
│       └── log_viewer.go   # LogViewer
│
├── component/              # 🟢 组件基础设施（保留）
│   ├── base.go
│   ├── capabilities.go
│   ├── container.go
│   ├── context.go
│   ├── state_holder.go
│   └── ...
│
├── _legacy/               # ⚫ 旧组件（兼容层，新建）
│   ├── display/
│   │   ├── text.go        # → components/basic/text.go
│   │   ├── list.go        # → components/data/list.go
│   │   └── table.go       # → components/data/table.go
│   ├── input/
│   │   └── textinput.go   # → components/form/input.go
│   └── interactive/
│       └── button.go      # → components/button/button.go
```

### 2.2 组件分类标准

| 分类 | 标准 | 示例 |
|------|------|------|
| **basic** | 不可交互的基础显示组件 | Text, Icon, Separator |
| **layout** | 用于组织和排列其他组件 | Box, Flex, Stack, Grid |
| **form** | 用于数据输入和表单构建 | Input, Checkbox, Select |
| **button** | 触发操作的按钮类组件 | Button, IconButton |
| **feedback** | 显示状态或进度信息 | ProgressBar, Toast, Alert |
| **data** | 展示结构化数据 | List, Table, Tree |
| **navigation** | 导航和路由 | Tabs, Menu, Sidebar |
| **overlay** | 覆盖在其他内容之上 | Modal, Dialog, Dropdown |
| **container** | 提供特定布局行为 | Panel, SplitPane, ScrollArea |
| **advanced** | 复杂功能组件 | Editor, Terminal, Chart |

---

## 三、组件迁移计划

### 3.1 第一阶段：目录创建 (Week 1)

```
framework/components/
├── basic/              # 从 display/ 迁移 Text
├── button/             # 从 interactive/ 迁移 Button
├── form/               # 从 input/ 迁移 TextInput
├── data/               # 从 display/ 迁移 List, Table
└── layout/             # 从 layout/ 迁移 Box, Flex
```

### 3.2 第二阶段：新组件开发 (Week 2-4)

```
# 新建组件
components/basic/
├── icon.go             # 新建
├── separator.go        # 新建
└── spacer.go           # 新建

components/form/
├── textarea.go         # 新建
├── checkbox.go         # 新建
├── radio.go            # 新建
├── select.go           # 新建
└── switch.go           # 新建

components/feedback/
├── progress.go         # 新建
├── toast.go            # 新建
└── alert.go            # 新建

components/navigation/
├── tabs.go             # 新建
└── menu.go             # 新建

components/overlay/
├── modal.go            # 新建
└── dialog.go           # 新建
```

### 3.3 第三阶段：高级组件 (Week 5+)

```
components/advanced/
├── editor.go           # 新建
├── terminal.go         # 新建
└── chart.go            # 新建
```

---

## 四、声明式组件 API

### 4.1 统一的导入路径

```go
// 旧 API（兼容）
import "github.com/wwsheng009/mint/framework/display"
text := display.NewText("Hello")

// 新 API（推荐）
import "github.com/wwsheng009/mint/ui"
text := ui.Text("Hello")

// 直接使用组件包
import "github.com/wwsheng009/mint/framework/components/basic"
text := basic.NewText("Hello")
```

### 4.2 组件 Builder 模式

```go
// ui 包提供简洁的声明式 API
package ui

// 基础组件
func Text(content string) *TextBuilder
func Icon(name string) *IconBuilder
func Separator() *SeparatorBuilder

// 布局组件
func HStack(children ...VNode) *LayoutBuilder
func VStack(children ...VNode) *LayoutBuilder
func Box() *BoxBuilder

// 表单组件
func Input(placeholder string) *InputBuilder
func Checkbox(label string) *CheckboxBuilder
func Select(options []string) *SelectBuilder

// 按钮组件
func Button(label string) *ButtonBuilder

// 数据组件
func List(items []string) *ListBuilder
func Table(headers []string) *TableBuilder

// 反馈组件
func Progress(value float64) *ProgressBuilder
func Toast(message string) *ToastBuilder

// 导航组件
func Tabs(tabs ...Tab) *TabsBuilder

// 覆盖层组件
func Modal(content VNode) *ModalBuilder
```

---

## 五、组件接口规范

### 5.1 组件基础接口

```go
// 所有组件应实现的基础接口
type Component interface {
    // 组件标识
    ID() string
    Type() string

    // 生命周期
    Mount(ctx Context) error
    Update(ctx Context) error
    Unmount(ctx Context) error

    // 渲染
    Measure(constraints Constraint) Size
    Paint(ctx PaintContext, buffer *Buffer)
}

// 可选能力接口
type Focuable interface {
    FocusID() string
    OnFocus()
    OnBlur()
}

type Clickable interface {
    OnClick(handler func(Event))
}

type Validatable interface {
    Validate() error
}
```

### 5.2 组件规范模板

```go
// components/XXX/xxx.go

package xxx

import (
    "github.com/wwsheng009/mint/framework/component"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
)

// Xxx 组件说明
type Xxx struct {
    *component.BaseComponent
    *component.StateHolder

    // 组件属性
    prop1 Type1
    prop2 Type2

    // 样式
    style style.Style
}

// NewXxx 创建组件
func NewXxx() *Xxx {
    return &Xxx{
        BaseComponent: component.NewBaseComponent("xxx"),
        StateHolder:   component.NewStateHolder(),
        // 初始化默认值
    }
}

// Measure 测量尺寸
func (x *Xxx) Measure(constraints Constraint) Size {
    // 实现测量逻辑
}

// Paint 绘制
func (x *Xxx) Paint(ctx PaintContext, buffer *Buffer) {
    // 实现绘制逻辑
}

// 链式设置方法
func (x *Xxx) SetProp1(v Type1) *Xxx {
    x.prop1 = v
    x.MarkDirty()
    return x
}
```

---

## 六、命名规范

### 6.1 组件命名

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件结构体 | 名词，首字母大写 | `Button`, `TextInput` |
| 构造函数 | `New` + 组件名 | `NewButton()`, `NewTextInput()` |
| Getter 方法 | `Get` + 属性名 | `GetLabel()`, `GetValue()` |
| Setter 方法 | `Set` + 属性名 | `SetLabel()`, `SetValue()` |
| 事件处理 | `On` + 事件名 | `OnClick()`, `OnChange()` |

### 6.2 文件命名

```
components/basic/
├── text.go           # Text 组件
├── text_test.go      # 测试文件
├── icon.go           # Icon 组件
└── separator.go      # Separator 组件
```

---

## 七、迁移检查清单

### 7.1 迁移前检查

- [ ] 确认组件当前功能
- [ ] 确认组件依赖关系
- [ ] 确认组件使用位置
- [ ] 编写迁移测试

### 7.2 迁移步骤

1. 在新目录创建组件文件
2. 复制并调整代码
3. 添加适配器保持兼容
4. 更新测试文件
5. 更新文档和示例
6. 标记旧代码为 `_legacy`

### 7.3 迁移后验证

- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 示例程序正常运行
- [ ] 文档更新完整

---

## 八、总结

### 分类原则

1. **按功能分类** - 而非按"显示/输入/交互"分类
2. **单一职责** - 每个目录职责明确
3. **易于查找** - 用户能快速找到需要的组件
4. **便于扩展** - 新组件有明确的归属

### 迁移策略

1. **渐进式迁移** - 一次迁移一个目录
2. **保持兼容** - 通过 `_legacy` 和适配器
3. **优先常用** - 先迁移常用的基础组件
4. **文档同步** - 代码和文档同步更新

### 下一步行动

1. 创建 `framework/components/` 目录
2. 创建各子目录（basic, layout, form, button, data 等）
3. 迁移 Text 组件作为试点
4. 逐步迁移其他组件
