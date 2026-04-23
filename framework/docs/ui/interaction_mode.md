# Interaction Mode（三态交互模式）

Mint 现在支持三种运行时交互模式，用于同时满足：

- 常规交互（点击、滚动、输入）
- 应用内文本选择（高亮 + 复制）
- 终端原生鼠标选中（系统复制体验）

## 模式说明

- `interactive`：默认模式  
  - 开启鼠标捕获
  - 适合普通组件交互
- `app_selection`：应用内选择模式  
  - 开启鼠标捕获
  - 接入 `runtime/selection`，支持拖拽选区与 `Ctrl+C` 复制
- `terminal_selection`：终端原生选择模式  
  - 关闭鼠标捕获
  - 允许终端直接框选文本并复制

## UI 层 API

初始化时设置：

```go
ui.Run(App,
    ui.WithInteractionMode(ui.InteractionModeAppSelection),
)
```

运行时切换：

```go
_ = ui.SetInteractionMode(ui.InteractionModeTerminalSelection)
mode, _ := ui.GetInteractionMode()
next, _ := ui.CycleInteractionMode()
_ = mode
_ = next
```

## 设计原则

- 三种模式可共存，但同一时刻只激活一种（避免鼠标事件竞争）
- 长期建议以 `app_selection` 为主、`terminal_selection` 为辅
- `terminal_selection` 适合调试/日志查看/快速复制场景
