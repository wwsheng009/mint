# Client - 调试客户端模块

> 调试界面、协议处理、可视化组件

## 功能概述

Client 模块提供 DevTools 的用户界面和通信协议实现，包括 TUI 调试面板、Web Dashboard 和 WebSocket 协议处理。

## 核心组件

### 1. TUI Debug Panel (`panel.go`)

```go
// TuiDebugPanel TUI 调试面板
type TuiDebugPanel struct {
    devtools *DevTools
    views    []PanelView
    current  int
}

// 创建面板
func NewTuiDebugPanel(dt *DevTools) *TuiDebugPanel

// 渲染面板
func (p *TuiDebugPanel) Render() string

// 处理输入
func (p *TuiDebugPanel) HandleInput(key rune) bool
```

**功能：**
- 多视图切换（Timeline、Causal、Stats、Patterns）
- 键盘导航（方向键、Tab、数字快捷键）
- 实时状态显示
- 组件高亮

### 2. Protocol (`protocol.go`)

```go
// WebSocket 通信协议
type Message struct {
    Version  string      `json:"version"`
    Type     string      `json:"type"`
    ID       string      `json:"id,omitempty"`
    Payload  interface{} `json:"payload,omitempty"`
}

// 消息类型
const (
    TypeHandshake   = "handshake"
    TypeGetSnapshot = "get_snapshot"
    TypeGetRange    = "get_range"
    TypeGetDiff     = "get_diff"
    TypeSubscribe   = "subscribe"
)
```

### 3. Visualizer (`visualizer.go`)

```go
// 因果链可视化
func VisualizeCausalGraph(graph *causal.CausalGraph) string

// 时间线可视化
func VisualizeTimeline(timeline *Timeline) string
```

### 4. Web Dashboard (`webdashboard.go`)

```go
// WebDashboard HTTP 服务器
type WebDashboard struct {
    port    int
    devtools *DevTools
}

// 启动 Web 服务
func (wd *WebDashboard) Start() error
```

## 使用方法

### TUI 面板

```go
import "github.com/wwsheng009/mint/devtools/client"

// 创建面板
panel := client.NewTuiDebugPanel(devtools)
panel.Enable()

// 渲染
output := panel.Render()
fmt.Println(output)

// 处理键盘输入
panel.HandleInput('t')  // 切换到 Timeline 视图
panel.HandleInput('c')  // 切换到 Causal 视图
panel.HandleInput('s')  // 切换到 Stats 视图
panel.HandleInput('p')  // 切换到 Patterns 视图
panel.HandleInput('q')  // 退出

// 组件检查
result := panel.Inspect("button-id")
fmt.Printf("Type: %s\n", result.Type)
```

### Web Dashboard

```go
// 创建并启动 Dashboard
dashboard := client.NewWebDashboard(8080, devtools)
go dashboard.Start()

// 访问 http://localhost:8080
```

### WebSocket 协议

```go
// 处理 WebSocket 消息
handler := func(msg *client.Message) *client.Message {
    switch msg.Type {
    case client.TypeGetSnapshot:
        // 返回快照
        return &client.Message{
            Type:    "snapshot",
            Payload: snapshotData,
        }
    case client.TypeGetDiff:
        // 返回差异
        return &client.Message{
            Type:    "diff",
            Payload: diffData,
        }
    }
    return nil
}
```

## 键盘快捷键

| 按键 | 功能 |
|------|------|
| `t` | Timeline 视图 |
| `c` | Causal Graph 视图 |
| `s` | Statistics 视图 |
| `p` | Patterns 视图 |
| `←` / `→` | 导航帧 |
| `Tab` | 切换视图 |
| `1-4` | 直接切换到视图 1-4 |
| `q` / `Esc` | 退出 |

## 相关模块

| 模块 | 关系 |
|------|------|
| `devtools` | 核心功能，面板观察和展示其数据 |
| `observation` | 统计数据和模式检测，面板展示其结果 |
| `causal` | 因果链数据，面板可视化因果关系 |
| `snapshot` | 快照数据，面板展示历史状态 |
| `remote` | WebSocket 服务器，与协议模块协同 |

## API 参考

### PanelView 接口

```go
type PanelView interface {
    Name() string
    Render(dt *DevTools) string
    HandleInput(key rune) bool
}
```

### InspectorResult

```go
type InspectorResult struct {
    NodeID    string
    Type      string
    Position  string
    Props     map[string]interface{}
    State     map[string]interface{}
    Bounds    string
}
```

## 文件列表

- `panel.go` - TUI 调试面板主文件
- `panel_v2.go` - 面板 V2 版本
- `protocol.go` - WebSocket 协议定义
- `visualizer.go` - 可视化工具
- `webdashboard.go` - Web Dashboard
- `tui/` - TUI 组件
- `web/` - Web 组件
