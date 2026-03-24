# Mint UI - 声明式终端UI框架

一个现代化的 Go 终端用户界面 (TUI) 框架，支持声明式UI开发、组件化架构和React-like Hooks。

## 特性

- **声明式UI** - React-like组件开发体验
- **Hooks系统** - useState, useEffect, useMemo, useCallback, useRef
- **组件库** - 丰富的预置组件 (Button, Input, Table, Progress等)
- **类型安全** - 完整的类型定义和类型推断
- **事件处理** - 完整的键盘事件和焦点管理
- **主题系统** - 支持多种内置主题

## 架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Application Layer                               │
│                    声明式UI组件 (ui/)                                   │
│  - Text, Button, Input, Checkbox, Table, Progress, Select, Spinner    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Framework Layer                                  │
│              组件系统 + 事件处理 + 布局引擎                                 │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Runtime Layer                                    │
│              渲染引擎 + 样式系统 + 焦点管理                                 │
└─────────────────────────────────────────────────────────────────────────┘
```

## 快速开始

### 环境要求

- Go 1.21+

### 安装

```bash
git clone https://github.com/wwsheng009/mint.git
cd mint
go mod download
```

## 运行示例

```bash
# Counter Demo - 计数器
cd examples/counter
go run .

# Timer Demo - 状态管理
cd examples/timer
go run .

# Input Demo - 输入组件
cd examples/input
go run .

# Checkbox Demo - 复选框
cd examples/checkbox
go run .

# Progress Demo - 进度条
cd examples/progress
go run .

# Select Demo - 选择器+表格
cd examples/select
go run .

# Date + Time Picker Demo
cd examples/date_time_picker_demo
go run .

# Demo - 综合演示
cd examples/demo
go run .
```

## 组件列表

### 基础组件

| 组件 | 说明 | 状态 |
|------|------|------|
| `Text` | 文本显示 | ✅ |
| `Button` | 按钮 | ✅ |
| `Input` | 单行输入 | ✅ |
| `Textarea` | 多行输入 | ✅ |
| `Checkbox` | 复选框 | ✅ |
| `DatePicker` | 日期选择 | ✅ |
| `TimePicker` | 时间选择 | ✅ |
| `Progress` | 进度条 | ✅ |
| `Spinner` | 加载动画 | ✅ |
| `Select` | 下拉选择器 | ✅ |
| `Table` | 表格 | ✅ |

### 布局组件

| 组件 | 说明 | 状态 |
|------|------|------|
| `VStack` | 垂直布局 | ✅ |
| `HStack` | 水平布局 | ✅ |
| `Fragment` | 片段容器 | ✅ |

### Hooks

| Hook | 说明 | 状态 |
|------|------|------|
| `UseStateInt` | 整数状态 | ✅ |
| `UseStateString` | 字符串状态 | ✅ |
| `UseStateBool` | 布尔状态 | ✅ |
| `UseEffect` | 副作用管理 | ✅ |
| `UseMemo` | 记忆化计算 | ✅ |
| `UseCallback` | 记忆化回调 | ✅ |
| `UseRef` | 引用 | ✅ |

## 使用示例

### Counter - 计数器

```go
package main

import "github.com/wwsheng009/mint/ui"

func main() {
    count, setCount, _ := ui.UseStateInt(0)

    ui.Run(func() ui.VNode {
        return ui.VStack(
            ui.NewTextBuilder("Count: %d", count).Build(),
            ui.ButtonBuilder("+").OnClick(func() {
                setCount(count + 1)
            }).Build(),
        )
    }, ui.WithWidth(40), ui.WithHeight(10))
}
```

### Input - 输入框

```go
text, setText := ui.UseStateString("")

ui.InputBuilder().
    Value(text).
    Placeholder("Type here...").
    MaxLength(20).
    OnChange(setText).
    Build()
```

### Table - 表格

```go
ui.TableBuilder().
    Columns([]ui.TableColumn{
        {Title: "ID", Width: 5},
        {Title: "Name", Width: 15},
        {Title: "Status", Width: 10},
    }).
    AddRow("1", "Alice", "Active").
    AddRow("2", "Bob", "Inactive").
    Build()
```

### Checkbox - 复选框

```go
checked, setChecked := ui.UseStateBool(false)

ui.CheckboxBuilder().
    Label("Accept terms").
    Checked(checked).
    OnChange(setChecked).
    Build()
```

### Progress - 进度条

```go
progress, setProgress, _ := ui.UseStateInt(0)

ui.ProgressBuilder().
    Label("Loading:").
    Value(progress).
    Max(100).
    ShowPercent(true).
    Build()
```

## 键盘快捷键

| 按键 | 功能 |
|------|------|
| `Tab` | 下一个焦点元素 |
| `Shift+Tab` | 上一个焦点元素 |
| `Enter` | 激活按钮/切换选项 |
| `Space` | 切换复选框 |
| `↑` `↓` | 选择器导航 |
| `←` `→` | 元素导航 |
| `Backspace` | 删除字符 |
| `q` | 退出程序 |

## 编译测试

```bash
# 运行所有测试
go test ./...

# 运行UI包测试
go test ./ui/... -v

# 查看测试覆盖率
go test ./ui/... -cover
```

## 目录结构

```
mint/
├── ui/                  # 声明式UI包
│   ├── vnode.go        # 虚拟节点
│   ├── component.go    # 组件系统
│   ├── layout.go       # 布局组件
│   ├── hooks.go        # Hooks系统
│   ├── button.go       # 按钮
│   ├── text.go         # 文本
│   ├── input.go        # 输入框
│   ├── checkbox.go     # 复选框
│   ├── progress.go     # 进度条/加载动画
│   ├── select.go       # 选择器/表格
│   ├── app.go          # 应用运行时
│   └── *_test.go       # 测试文件
│
├── framework/          # 框架层
├── runtime/            # 运行时层
├── examples/           # 示例程序
│   ├── counter/        # 计数器示例
│   ├── timer/          # 状态管理示例
│   ├── input/          # 输入组件示例
│   ├── checkbox/       # 复选框示例
│   ├── date_time_picker_demo/ # 日期时间选择示例
│   ├── progress/       # 进度条示例
│   ├── select/         # 选择器示例
│   └── demo/           # 综合演示
│
└── docs/               # 文档
```

## 开发路线图

- [x] 基础组件 (Text, Button)
- [x] 输入组件 (Input, Textarea)
- [x] 选择组件 (Checkbox, Select)
- [x] 反馈组件 (Progress, Spinner)
- [x] 数据展示 (Table)
- [x] Hooks系统 (useState, useEffect, useMemo, useCallback, useRef)
- [x] 布局组件 (VStack, HStack)
- [ ] 更多组件 (Modal, Tabs, Slider, List...)
- [ ] 动画系统
- [ ] 路由系统

## 测试状态

```
ok  	github.com/wwsheng009/mint/ui	coverage: 40.7%
```

- **136个测试全部通过** ✅

## 许可证

MIT License
