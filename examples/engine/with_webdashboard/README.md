# WebDashboard 集成指南

## 核心概念

**不需要** `runSimulation` 函数！那是示例程序用来模拟活动的。

在实际应用中，数据来自真实的用户交互和渲染循环。

## 集成步骤 (3步)

### 1. 创建并启动 WebDashboard

```go
import "github.com/wwsheng009/mint/devtools/client"

var dashboard *client.WebDashboard

func init() {
    dashboard = client.NewWebDashboard(8080) // 端口号
}

func main() {
    // 启动
    dashboard.Start()
    defer dashboard.Stop()

    // ... 你的应用代码 ...
}
```

### 2. 在渲染循环中使用 BeginFrame/EndFrame

```go
import "github.com/wwsheng009/mint/devtools"

var dt = devtools.New()

func main() {
    dt.Enable()
    defer dt.Disable()

    for running {
        dt.BeginFrame()      // 开始帧记录
        // ... 渲染逻辑 ...
        dt.EndFrame()        // 结束帧记录，自动发送到 WebDashboard
    }
}
```

### 3. (可选) 更新组件和指标

```go
// 组件状态变化时
dashboard.UpdateComponent("button-1", &client.DashboardComponent{
    ID:   "button-1",
    Type: "Button",
    Properties: map[string]interface{}{
        "text":    "Click Me",
        "focused": true,
    },
})

// 定期更新性能指标
dashboard.UpdateMetrics(&client.DashboardMetrics{
    FPS:          60.0,
    FrameTime:    16 * time.Millisecond,
    MemoryUsage:  50 * 1024 * 1024,
})
```

## 完整示例

参见 `main.go` 文件，展示了：
- 在现有 TUI 应用中集成 WebDashboard
- 使用 BeginFrame/EndFrame 自动捕获帧数据
- 实时更新组件状态到 Web UI

## 运行

```bash
go run main.go
```

然后打开 http://localhost:8080/ 查看实时调试面板。

## 与现有 DevTools 的关系

| 组件 | 用途 |
|------|------|
| `devtools.DevTools` | 核心数据收集 |
| `client.WebDashboard` | Web UI 展示 |
| `client.RemoteDebugSession` | 远程调试会话 |

WebDashboard 自动使用 DevTools 收集的数据，无需手动同步。
