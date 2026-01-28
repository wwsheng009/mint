# TUI Framework

一个现代化的 Go 终端用户界面 (TUI) 框架，采用四层架构设计，支持组件化开发、主题切换和 AI 集成。

## 特性

- **组件化** - 基于能力接口的组件系统
- **事件驱动** - 完整的事件处理和 Action 系统
- **主题系统** - 支持多种内置主题和自定义主题
- **表单验证** - 内置表单组件和验证器
- **虚拟滚动** - 高性能的大数据列表支持
- **AI 友好** - 状态快照和操作回放，便于 AI 集成

## 架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Application Layer                               │
│  用户应用代码 - 使用 Framework API 构建具体应用                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Framework Layer                                  │
│  组件系统 + Action 路由 + 适配器                                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Runtime Layer                                    │
│  纯内核 - 无外部依赖，可独立测试                                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Platform Layer                                   │
│  平台抽象 - 最小化接口                                                    │
└─────────────────────────────────────────────────────────────────────────┘
```

## 目录结构

```
mint/
├── framework/          # 框架层（组件、样式、表单等）
│   ├── component/      # 组件系统
│   ├── display/        # 显示组件（Text、List、Table）
│   ├── input/          # 输入组件（TextInput、TextArea）
│   ├── interactive/    # 交互组件（Button、Checkbox）
│   ├── layout/         # 布局组件（Box、Flex）
│   ├── form/           # 表单组件
│   ├── styling/        # 样式系统
│   ├── theme/          # 主题系统
│   ├── validation/     # 验证器
│   ├── event/          # 事件系统
│   ├── cursor/         # 光标管理
│   ├── screen/         # 屏幕管理
│   └── examples/       # 示例代码
│
├── runtime/            # 运行时层（内核）
│   ├── action/         # Action 系统
│   ├── animation/      # 动画系统
│   ├── event/          # 底层事件处理
│   ├── focus/          # 焦点管理
│   ├── input/          # 输入处理
│   ├── layout/         # 布局引擎
│   ├── paint/          # 绘制系统
│   └── state/          # 状态管理
│
└── docs/               # 框架文档
```

## 快速开始

### 环境要求

- Go 1.21+

### 安装

```bash
git clone https://github.com/yaoapp/yao/tui.git
cd tui
go mod download
```

### 运行示例

```bash
# Hello World
go run framework/examples/hello/main.go

# 组件演示
go run framework/examples/demo/main.go

# 主题切换演示
go run framework/examples/theme/main.go

# 登录表单
go run framework/examples/login/interactive/main.go
```

### 编译测试

```bash
# 运行所有测试
go test ./...

# 运行单个包测试
go test ./framework/component
```

## 组件示例

### 文本显示

```go
import "github.com/yaoapp/yao/tui/framework/display"
import "github.com/yaoapp/yao/tui/runtime/style"

text := display.NewText("Hello, TUI!")
text.SetStyle(style.Style{}.Foreground(style.Blue).Bold(true))
```

### 文本输入

```go
import "github.com/yaoapp/yao/tui/framework/input"

input := input.NewTextInput()
input.SetPlaceholder("请输入用户名")
input.SetValue("demo")
```

### 按钮

```go
import "github.com/yaoapp/yao/tui/framework/interactive"

button := interactive.NewButton("提交")
button.SetOnClick(func() {
    fmt.Println("按钮被点击")
})
```

### 表单

```go
import "github.com/yaoapp/yao/tui/framework/form"
import "github.com/yaoapp/yao/tui/framework/validation"

form := form.NewForm()

usernameField := form.NewFormField("username")
usernameField.Label = "用户名"
usernameField.Input = input.NewTextInput()
usernameField.Validators = []validation.Validator{
    validation.Required(),
    validation.MinLength(3),
}

form.AddField(usernameField)

form.SetOnSubmit(func(data map[string]interface{}) error {
    fmt.Printf("提交: %v\n", data)
    return nil
})
```

### 表格

```go
import "github.com/yaoapp/yao/tui/framework/display"

table := display.NewTable([]display.TableColumn{
    {Title: "ID", Width: 10},
    {Title: "名称", Width: 30},
    {Title: "状态", Width: 15},
})

table.SetRows([][]string{
    {"1", "用户管理", "完成"},
    {"2", "订单系统", "进行中"},
})
```

## 可用主题

| 主题 | 说明 |
|------|------|
| `light` | 亮色主题 |
| `dark` | 暗色主题 |
| `dracula` | Dracula 配色 |
| `nord` | Nord 冷色调 |
| `monokai` | Monokai 暗色 |
| `tokyo-night` | Tokyo Night |

## 示例列表

| 示例 | 描述 |
|------|------|
| `hello` | Hello World |
| `demo` | 完整组件演示 |
| `theme` | 主题切换演示 |
| `login` | 登录表单示例 |

## 文档

详细文档请查看 [framework/docs](framework/docs/)

- [架构设计](framework/docs/ARCHITECTURE.md)
- [组件系统](framework/docs/COMPONENTS.md)
- [事件系统](framework/docs/EVENT_SYSTEM.md)
- [主题系统](framework/docs/THEME_SYSTEM.md)
- [表单验证](framework/docs/FORM_VALIDATION.md)

## 许可证

[MIT License](LICENSE)
