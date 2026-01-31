# DevToolsServer 集成指南

## 核心概念

**不需要** `runSimulation` 函数！那是示例程序用来模拟活动的。

在实际应用中，数据来自真实的用户交互和渲染循环。

DevToolsServer 是统一的调试服务器，整合了之前 WebDashboard 和 remote 包的功能。

## 集成步骤 (3步)

### 1. 创建并启动 DevToolsServer

```go
import "github.com/wwsheng009/mint/devtools/client"

var server *client.DevToolsServer

func init() {
    var err error
    server, err = client.NewDevToolsServerForApplication(8080) // 端口号
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    // 启动
    server.Start()
    defer server.Stop()

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
        dt.EndFrame()        // 结束帧记录，自动发送到 DevToolsServer
    }
}
```

### 3. (可选) 更新组件和指标

```go
// 组件状态变化时
server.UpdateComponent("button-1", &client.ComponentData{
    ID:   "button-1",
    Type: "Button",
    Properties: map[string]interface{}{
        "text":    "Click Me",
        "focused": true,
    },
})

// 定期更新性能指标
server.UpdateMetrics(&client.Metrics{
    FPS:          60.0,
    FrameTime:    16 * time.Millisecond,
    MemoryUsage:  50 * 1024 * 1024,
})
```

## 完整示例

参见 `main.go` 文件，展示了：
- 在现有 TUI 应用中集成 DevToolsServer
- 使用 BeginFrame/EndFrame 自动捕获帧数据
- 实时更新组件状态到 Web UI

## 运行

```bash
go run main.go
```

然后打开 http://localhost:8080/ 查看实时调试面板。

## DevToolsServer 功能

| 功能 | 说明 |
|------|------|
| `devtools.DevTools` | 核心数据收集 |
| `client.DevToolsServer` | 统一的调试服务器 (HTTP + WebSocket) |
| 性能指标 | FPS、帧时间、内存使用等 |
| 帧时间线 | 查看每一帧的事件、变更、渲染 |
| 组件树 | 实时查看组件状态 |
| 快照对比 | 比较不同帧之间的差异 |
| 调试报告 | 生成完整的调试报告 |

DevToolsServer 自动使用 DevTools 收集的数据，无需手动同步。
