# Examples

示例代码和演示程序。

## 职责

- 提供组件使用示例
- 演示 Framework 功能
- 学习资源

## 示例列表

| 示例 | 描述 | 运行命令 |
|------|------|----------|
| `hello` | Hello World 示例 | `go run framework/examples/hello/main.go` |
| `demo` | 完整组件演示 | `go run framework/examples/demo/main.go` |
| `theme` | 主题样式演示 | `go run framework/examples/theme/main.go` |
| `login/interactive` | 交互式登录表单 | `go run framework/examples/login/interactive/main.go` |
| `login/simple` | 简单输入框示例 | `go run framework/examples/login/simple/main.go` |
| `login/form` | 登录表单示例 | `go run framework/examples/login/form/main.go` |

## 快速开始

### Hello World

最简单的示例，展示基本的文本组件和事件处理：

```bash
go run framework/examples/hello/main.go
```

### 组件演示

展示所有可用组件的功能：

```bash
go run framework/examples/demo/main.go
```

### 主题演示

展示不同主题和样式的效果：

```bash
go run framework/examples/theme/main.go
```

### 登录表单

交互式登录表单，支持：
- 字段导航 (Tab/方向键)
- 输入验证
- 提交/取消

```bash
go run framework/examples/login/interactive/main.go
```

## 主题演示

主题系统展示了完整的样式功能：

### 可用主题

| 主题 | 说明 | 适用场景 |
|------|------|----------|
| `light` | 亮色主题 | 白天使用 |
| `dark` | 暗色主题 | 夜间使用 |
| `dracula` | Dracula 配色 | 流行暗色主题 |
| `nord` | Nord 冷色调 | 清爽风格 |
| `monokai` | Monokai 暗色 | 代码编辑器风格 |
| `tokyo-night` | Tokyo Night | 现代暗色 |

### 在代码中使用主题

```go
import (
    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/framework/theme"
    "github.com/wwsheng009/mint/runtime/style"
)

func main() {
    app := framework.NewApp()

    // 初始化主题
    if err := app.InitTheme("dark"); err != nil {
        panic(err)
    }

    // 运行时切换主题
    app.SetTheme("light")

    // 获取当前主题
    current := app.GetTheme()
    println("Current theme:", current)

    // 运行应用
    app.Run()
}
```

### 组件使用主题样式

```go
import "github.com/wwsheng009/mint/runtime/style"

// 在组件内部获取主题样式
func (c *MyComponent) getStyle() style.Style {
    // 获取焦点状态样式
    return style.Style{}.Foreground(style.Blue).Bold(true)
}
```

## 相关文档

- [../../docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md) - 架构设计
- [../../docs/COMPONENTS.md](../../docs/COMPONENTS.md) - 组件系统
- [../../docs/THEME_SYSTEM.md](../../docs/THEME_SYSTEM.md) - 主题系统
- [../../docs/FORM_VALIDATION.md](../../docs/FORM_VALIDATION.md) - 表单验证
