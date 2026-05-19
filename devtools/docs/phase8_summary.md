# Phase 8: 远程调试实施总结

> **Remote Debugging & Cloud Sync** - WebSocket 协议、HTTP API、Chromium 集成

## 概述

Phase 8 实现了远程调试功能，允许通过浏览器或远程工具实时调试 TUI 应用。兼容 Chrome DevTools Protocol (CDP)。

## 核心组件

### 1. 协议定义 (`remote/protocol.go`)

```go
// 消息类型
const (
    TypeHandshake        = "handshake"
    TypeGetSnapshot      = "get_snapshot"
    TypeGetRange         = "get_range"
    TypeGetDiff          = "get_diff"
    TypeSetBreakpoint    = "set_breakpoint"
    TypeClearBreakpoint  = "clear_breakpoint"
    TypeSubscribe        = "subscribe"
    TypeUnsubscribe      = "unsubscribe"

    TypeHandshakeAck     = "handshake_ack"
    TypeEvent            = "event"
    TypeSnapshot         = "snapshot"
    TypeDiff             = "diff"
    TypeError            = "error"
)

// Message 通用消息格式
type Message struct {
    Version string      `json:"version"`
    Type    string      `json:"type"`
    ID      string      `json:"id,omitempty"`
    Payload interface{} `json:"payload,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// Session 调试会话
type Session struct {
    id        string
    clientID  string
    createdAt time.Time
    lastSeen  time.Time
    subs      map[string]bool
}
```

### 2. Chromium Bridge (`remote/chromium.go`)

```go
// ChromiumBridge Chromium DevTools 桥接器
type ChromiumBridge struct {
    server    *Server
    devtools  *devtools.DevTools
    snapshots *snapshot.Manager
    enabled   bool
}

// Server 远程调试服务器
type Server struct {
    mu        sync.RWMutex
    sessions  map[string]*Session
    port      int
    path      string
}

// 主要方法
func NewChromiumBridge(dt *devtools.DevTools, sm *snapshot.Manager) *ChromiumBridge
func (b *ChromiumBridge) Enable()
func (b *ChromiumBridge) Disable()
func (b *ChromiumBridge) IsEnabled() bool
func (b *ChromiumBridge) GetInspectorHTML() string
func (b *ChromiumBridge) ExportForChromium() ([]byte, error)
```

### 3. WebSocket 服务器 (`remote/websocket_server.go`)

```go
// WebSocketServer WebSocket 处理器
type WebSocketServer struct {
    mu      sync.RWMutex
    bridge  *ChromiumBridge
    clients map[*websocket.Conn]string
}

// DevToolsServer 完整的 DevTools 服务器
type DevToolsServer struct {
    mu              sync.RWMutex
    wsServer       *WebSocketServer
    httpServeMux    *http.ServeMux
    bridge          *ChromiumBridge
    snapshotManager *snapshot.Manager
    port            int
}

// 创建服务器
func NewDevToolsServer(port int, dt *devtools.DevTools, sm *snapshot.Manager) *DevToolsServer
func (s *DevToolsServer) Start() error
func (s *DevToolsServer) StartInBackground() error
```

### 4. HTTP 服务器 (`remote/http_server.go`)

```go
// HTTPServer HTTP API 服务器 (无 WebSocket)
type HTTPServer struct {
    mu              sync.RWMutex
    bridge          *ChromiumBridge
    snapshotManager *snapshot.Manager
    port            int
    mux             *http.ServeMux
}

// 创建服务器
func NewHTTPServer(port int, dt *devtools.DevTools, sm *snapshot.Manager) *HTTPServer
func (s *HTTPServer) Start() error
```

## API 端点

### HTTP REST API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/debug` | GET | Inspector HTML 页面 |
| `/health` | GET | 健康检查 |
| `/api/snapshots` | GET | 获取所有快照列表 |
| `/api/snapshot/` | GET | 获取指定帧快照 |
| `/api/diff` | GET | 比较两个快照 |
| `/api/export` | GET | 导出 Chromium 格式 |

### WebSocket 端点

| 端点 | 说明 |
|------|------|
| `/ws` | WebSocket 连接端点 |

## 使用示例

### 启动远程调试服务器

```go
package main

import (
    "github.com/wwsheng009/mint/devtools"
    "github.com/wwsheng009/mint/devtools/remote"
    "github.com/wwsheng009/mint/devtools/snapshot"
)

func main() {
    // 创建 DevTools 和快照管理器
    dt := devtools.New()
    dt.Enable()

    snapshotMgr := snapshot.NewManager(100)

    // 创建并启动服务器 (HTTP + WebSocket)
    server := remote.NewDevToolsServer(9222, dt, snapshotMgr)
    go func() {
        if err := server.Start(); err != nil {
            log.Fatal(err)
        }
    }()

    // 服务器运行在:
    // - http://localhost:9222/debug    (Inspector UI)
    // - ws://localhost:9222/ws          (WebSocket)
    // - http://localhost:9222/api/*    (REST API)

    select {} // 保持运行
}
```

### HTTP API 调用

```bash
# 健康检查
curl http://localhost:9222/health
# {"server":"mint-devtools","version":"1.0.0","ws_clients":1,"snapshots":4}

# 获取所有快照
curl http://localhost:9222/api/snapshots
# [{"id":"snap-0","frame_id":0,"timestamp":"...",...}, ...]

# 比较两个帧
curl "http://localhost:9222/api/diff?from=0&to=9"
# {"from":0,"to":9,"changes":[{...}],...}
```

### WebSocket 客户端

```go
package main

import (
    "encoding/json"
    "golang.org/x/net/websocket"
)

func main() {
    // 连接 WebSocket
    ws, err := websocket.Dial("ws://localhost:9222/ws", "", "http://localhost")
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    // 发送握手
    handshake := map[string]interface{}{
        "version": "1.0.0",
        "type":    "handshake",
        "id":      "client-1",
        "payload": map[string]string{"client_id": "my-client"},
    }
    websocket.JSON.Send(ws, handshake)

    // 接收握手确认
    var msg map[string]interface{}
    websocket.JSON.Receive(ws, &msg)
    fmt.Printf("Handshake: %+v\n", msg)

    // 请求快照范围
    getRange := map[string]interface{}{
        "version": "1.0.0",
        "type":    "get_range",
        "id":      "req-1",
        "payload": map[string]int{"from": 0, "to": 100},
    }
    websocket.JSON.Send(ws, getRange)

    // 接收响应
    websocket.JSON.Receive(ws, &msg)

    // 解析快照
    payload := msg["payload"].(map[string]interface{})
    frames := payload["frames"].([]interface{})
    fmt.Printf("Received %d frames\n", len(frames))
}
```

### 浏览器 Inspector

打开 `http://localhost:9222/debug`，自动连接 WebSocket 并显示：

```
┌─────────────────────────────────────────────────────────────┐
│              Mint TUI DevTools Inspector                     │
├─────────────────────────────────────────────────────────────┤
│ Connection                                                   │
│ Status: Connected                                           │
│ [Connect] [Disconnect]                                       │
├─────────────────────────────────────────────────────────────┤
│ Snapshots                                                    │
│ [Refresh]                                                    │
│ ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│ │ Frame 0      │  │ Frame 3      │  │ Frame 6      │        │
│ │ ID: snap-0   │  │ ID: snap-3   │  │ ID: snap-6   │        │
│ │ Components: 1│  │ Components: 1│  │ Components: 1│        │
│ └──────────────┘  └──────────────┘  └──────────────┘        │
├─────────────────────────────────────────────────────────────┤
│ Diff                                                         │
│ From: [0] To: [9] [Compare]                                 │
│                                                              │
│ ~ node-0: focused                                           │
│   Old: true                                                  │
│   New: false                                                 │
│                                                              │
│ ~ node-0: bounds.y                                          │
│   Old: 0                                                     │
│   New: 9                                                     │
└─────────────────────────────────────────────────────────────┘
```

## 协议消息格式

### 握手

```json
// 客户端 → 服务器
{
  "version": "1.0.0",
  "type": "handshake",
  "id": "handshake-123",
  "payload": {
    "client_id": "inspector-456",
    "capabilities": ["snapshots", "events", "diffs"]
  }
}

// 服务器 → 客户端
{
  "version": "1.0.0",
  "type": "handshake_ack",
  "payload": {
    "server_id": "mint-devtools",
    "version": "1.0.0",
    "session_id": "session-abc",
    "capabilities": ["snapshots", "events", "diffs", "breakpoints"]
  }
}
```

### 获取快照范围

```json
// 客户端 → 服务器
{
  "version": "1.0.0",
  "type": "get_range",
  "id": "req-1",
  "payload": {
    "from": 0,
    "to": 100
  }
}

// 服务器 → 客户端
{
  "version": "1.0.0",
  "type": "get_range",
  "id": "req-1",
  "payload": {
    "frames": [
      {
        "frame_id": 0,
        "timestamp": "2024-01-30T10:00:00Z",
        "events": 3,
        "mutations": 2,
        "layouts": 1
      }
    ]
  }
}
```

### 获取差异

```json
// 客户端 → 服务器
{
  "version": "1.0.0",
  "type": "get_diff",
  "id": "req-2",
  "payload": {
    "from": 0,
    "to": 9
  }
}

// 服务器 → 客户端
{
  "version": "1.0.0",
  "type": "diff",
  "id": "req-2",
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

Inspector 是一个自包含的 HTML 页面，内嵌在服务器中：

```javascript
// 自动连接
function connect() {
    ws = new WebSocket('ws://localhost:9222/ws');
    ws.onopen = () => {
        // 发送握手
        ws.send(JSON.stringify({
            version: '1.0.0',
            type: 'handshake',
            id: 'handshake-' + Date.now(),
            payload: { client_id: 'inspector-' + Date.now() }
        }));
    };
    ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        handleMessage(msg);
    };
}

// 处理消息
function handleMessage(msg) {
    switch(msg.type) {
        case 'snapshot':
            displaySnapshot(msg.payload);
            break;
        case 'get_range':
            displaySnapshots(msg.payload);
            break;
        case 'diff':
            displayDiff(msg.payload);
            break;
    }
}
```

## 测试客户端

测试客户端位于 `tests/wsclient/main.go`：

```bash
cd tests/wsclient
go run main.go
```

输出：
```
=== WebSocket Test Client ===
Connected to ws://localhost:9222/ws
Received: map[server_id:mint-devtools ...]
Sent handshake
Received: map[id:test-1 ...]
Sent get_range

=== Response ===
{
  "payload": {
    "frames": [
      {"frame_id": 0, "events": 0, ...},
      {"frame_id": 3, "events": 0, ...},
      ...
    ]
  }
}

Frames count: 4
```

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                      Browser / Client                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │  Inspector UI (http://localhost:9222/debug)             │   │
│   │  - Snapshots view                                        │   │
│   │  - Diff view                                             │   │
│   │  - Timeline view                                         │   │
│   └─────────────────────────────────────────────────────────┘   │
│                              ↕                                   │
│                    WebSocket (JSON)                             │
│                              ↕                                   │
└──────────────────────────────────────────────────────────────────┘
                                ↕
┌─────────────────────────────────────────────────────────────────┐
│                    DevTools Server                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│   │  HTTP Server │  │ WebSocket    │  │  Chromium    │        │
│   │  (REST API)  │  │  Server      │  │  Bridge      │        │
│   └──────────────┘  └──────────────┘  └──────────────┘        │
│                                                                  │
│   Protocol Handlers:                                             │
│   - handshake / handshake_ack                                    │
│   - get_snapshot / snapshot                                      │
│   - get_range / get_range (response)                             │
│   - get_diff / diff                                              │
│   - set_breakpoint / breakpoint_hit                              │
│   - subscribe / event                                            │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
                                ↕
┌─────────────────────────────────────────────────────────────────┐
│                     Mint DevTools Core                           │
├─────────────────────────────────────────────────────────────────┤
│   DevTools   SnapshotManager   ObservationLayer                  │
└──────────────────────────────────────────────────────────────────┘
```

## 测试

```bash
# 启动服务器
cd examples/devtools_demo
go run ./remote_demo

# 运行 WebSocket 测试客户端
cd tests/wsclient
go run main.go

# 在浏览器中打开
open http://localhost:9222/debug
```

## 状态

✅ **已完成**
- [x] 协议定义 (protocol.go)
- [x] Chromium Bridge (chromium.go)
- [x] WebSocket 服务器 (websocket_server.go)
- [x] HTTP API 服务器 (http_server.go)
- [x] Inspector UI (内嵌 HTML)
- [x] 错误处理 (JSON 错误响应)
- [x] 测试客户端
- [x] 文档
