# Remote - 远程调试模块

> WebSocket 协议、HTTP API、Chromium 集成、Inspector UI

## 功能概述

Remote 模块提供远程调试功能，允许通过浏览器或远程工具实时调试 TUI 应用。兼容 Chrome DevTools Protocol (CDP)。

## 模块结构

```
remote/
├── protocol.go        # 消息协议定义
├── chromium.go        # Chromium 桥接器
├── websocket_server.go # WebSocket 服务器
├── http_server.go     # HTTP API 服务器
└── simple_server.go   # 简单服务器
```

## 核心组件

### 1. Protocol (`protocol.go`)

```go
// 消息类型
const (
    TypeHandshake       = "handshake"
    TypeGetSnapshot     = "get_snapshot"
    TypeGetRange        = "get_range"
    TypeGetDiff         = "get_diff"
    TypeSetBreakpoint   = "set_breakpoint"
    TypeSubscribe       = "subscribe"
)

// Message 通用消息
type Message struct {
    Version string      `json:"version"`
    Type    string      `json:"type"`
    ID      string      `json:"id,omitempty"`
    Payload interface{} `json:"payload,omitempty"`
}

// Session 调试会话
type Session struct {
    id        string
    clientID  string
    createdAt time.Time
    subs      map[string]bool
}
```

### 2. Chromium Bridge (`chromium.go`)

```go
// ChromiumBridge CDP 桥接器
type ChromiumBridge struct {
    server    *Server
    devtools  *devtools.DevTools
    snapshots *snapshot.Manager
    enabled   bool
}

// 创建桥接器
func NewChromiumBridge(dt *devtools.DevTools, sm *snapshot.Manager) *ChromiumBridge

// 获取 Inspector HTML
func (b *ChromiumBridge) GetInspectorHTML() string

// 导出 Chromium 格式
func (b *ChromiumBridge) ExportForChromium() ([]byte, error)
```

### 3. WebSocket Server (`websocket_server.go`)

```go
// WebSocketServer WebSocket 处理器
type WebSocketServer struct {
    bridge  *ChromiumBridge
    clients map[*websocket.Conn]string
}

// DevToolsServer 完整服务器
type DevToolsServer struct {
    wsServer       *WebSocketServer
    httpServeMux    *http.ServeMux
    bridge          *ChromiumBridge
    snapshotManager *snapshot.Manager
    port            int
}

// 创建服务器
func NewDevToolsServer(port int, dt *devtools.DevTools, sm *snapshot.Manager) *DevToolsServer

// 启动服务器
func (s *DevToolsServer) Start() error
```

### 4. HTTP Server (`http_server.go`)

```go
// HTTPServer HTTP API 服务器
type HTTPServer struct {
    bridge          *ChromiumBridge
    snapshotManager *snapshot.Manager
    port            int
}

// 创建 HTTP 服务器
func NewHTTPServer(port int, dt *devtools.DevTools, sm *snapshot.Manager) *HTTPServer

// 启动服务器
func (s *HTTPServer) Start() error
```

## API 端点

### HTTP REST API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/debug` | GET | Inspector HTML 页面 |
| `/health` | GET | 健康检查 |
| `/api/snapshots` | GET | 获取所有快照列表 |
| `/api/diff` | GET | 比较两个快照 |
| `/api/export` | GET | 导出 Chromium 格式 |

### WebSocket 端点

| 端点 | 说明 |
|------|------|
| `/ws` | WebSocket 连接端点 |

## 使用方法

### 启动远程调试服务器

```go
import "github.com/wwsheng009/mint/devtools/remote"

// 创建 DevTools 和快照管理器
dt := devtools.New()
dt.Enable()
snapshotMgr := snapshot.NewManager(100)

// 创建并启动服务器
server := remote.NewDevToolsServer(9222, dt, snapshotMgr)
go server.Start()

// 服务器运行在:
// - http://localhost:9222/debug    (Inspector UI)
// - ws://localhost:9222/ws          (WebSocket)
// - http://localhost:9222/api/*    (REST API)
```

### HTTP API 调用

```bash
# 健康检查
curl http://localhost:9222/health

# 获取所有快照
curl http://localhost:9222/api/snapshots

# 比较两个帧
curl "http://localhost:9222/api/diff?from=0&to=9"
```

### WebSocket 客户端

```go
import "golang.org/x/net/websocket"

// 连接
ws, err := websocket.Dial("ws://localhost:9222/ws", "", "http://localhost")

// 发送握手
websocket.JSON.Send(ws, map[string]interface{}{
    "version": "1.0.0",
    "type": "handshake",
    "id": "client-1",
    "payload": map[string]string{"client_id": "my-client"},
})

// 请求快照范围
websocket.JSON.Send(ws, map[string]interface{}{
    "version": "1.0.0",
    "type": "get_range",
    "id": "req-1",
    "payload": map[string]int{"from": 0, "to": 100},
})

// 接收响应
var msg map[string]interface{}
websocket.JSON.Receive(ws, &msg)
```

## 协议消息格式

### 握手

```json
// 客户端 → 服务器
{
  "version": "1.0.0",
  "type": "handshake",
  "payload": {"client_id": "inspector-123"}
}

// 服务器 → 客户端
{
  "version": "1.0.0",
  "type": "handshake_ack",
  "payload": {
    "server_id": "mint-devtools",
    "capabilities": ["snapshots", "events", "diffs"]
  }
}
```

### 获取差异

```json
// 客户端 → 服务器
{
  "version": "1.0.0",
  "type": "get_diff",
  "payload": {"from": 0, "to": 9}
}

// 服务器 → 客户端
{
  "version": "1.0.0",
  "type": "diff",
  "payload": {
    "from": 0,
    "to": 9,
    "changes": [
      {
        "node_id": "node-0",
        "type": "modified",
        "path": "focused",
        "old_value": true,
        "new_value": false
      }
    ]
  }
}
```

## Inspector UI

内置的 HTML Inspector 页面提供：
- 连接状态显示
- 快照列表
- 差异对比
- 实时更新

访问 `http://localhost:9222/debug` 打开。

## 相关模块

| 模块 | 关系 |
|------|------|
| `devtools` | 核心功能，被远程调试观察和控制 |
| `snapshot` | 快照数据，通过 API 提供 |
| `client` | 协议定义，与远程模块共享 |

## API 参考

### Server Stats

```go
type ServerStats struct {
    Port         int
    SessionCount int
    Enabled      bool
}
```

### Connection

```go
type Connection struct {
    ID        string
    ClientID  string
    Connected time.Time
    Subscribed []string
}
```

## 文件列表

- `protocol.go` - 消息协议定义
- `chromium.go` - Chromium 桥接器
- `websocket_server.go` - WebSocket 服务器
- `http_server.go` - HTTP API 服务器
- `simple_server.go` - 简单服务器
