# Protocol - 统一协议模块

> 类型定义、WebSocket 服务器、HTTP API、Web 仪表盘

## 数据流架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        DevTools Server                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────┐      ┌──────────────────────┐        │
│  │   HTTP REST API      │      │    WebSocket API     │        │
│  ├──────────────────────┤      ├──────────────────────┤        │
│  │ GET  /health         │      │ ws://host/ws         │        │
│  │ GET  /api/metrics    │      │ metrics_updated      │        │
│  │ GET  /api/frames     │      │ frame_added          │        │
│  │ GET  /api/components │      │ component_updated    │        │
│  │ GET  /api/snapshots  │      │ handshake_ack        │        │
│  │ GET  /api/snapshot   │      │                      │        │
│  │ GET  /api/diff       │      │                      │        │
│  │ GET  /api/report     │      │                      │        │
│  │ GET  /api/export     │      │                      │        │
│  │ POST /api/import     │      │                      │        │
│  └──────────────────────┘      └──────────────────────┘        │
│           │                              │                       │
│           │        ┌─────────────────────┼─────────────────┐   │
│           │        │                     │                 │   │
│           ▼        ▼                     ▼                 ▼   │
│  ┌───────────────┐ ┌───────────────┐ ┌─────────────────────┐  │
│  │   Metrics     │ │    Frames     │ │    Components       │  │
│  │   UpdateMetrics│ │   AddFrame    │ │   UpdateComponent   │  │
│  └───────────────┘ └───────────────┘ └─────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Web Dashboard (HTML/JS)                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  AppState:                                                       │
│  ├── metrics       ← WebSocket: metrics_updated                │
│  ├── frames        ← WebSocket: frame_added                    │
│  ├── components    ← WebSocket: component_updated              │
│  ├── snapshots     ← HTTP: GET /api/snapshots                  │
│  ├── fpsHistory    ← Local chart data                          │
│  └── memoryHistory ← Local chart data                          │
│                                                                  │
│  Views:                                                          │
│  ├── Dashboard    ← 混合: WebSocket + HTTP API                 │
│  ├── Metrics      ← HTTP: GET /api/metrics                     │
│  ├── Frames       ← HTTP: GET /api/frames                      │
│  ├── Components   ← HTTP: GET /api/components                  │
│  ├── Snapshots    ← HTTP: GET /api/snapshots                    │
│  ├── Diff         ← HTTP: GET /api/diff?from=X&to=Y            │
│  └── Report       ← HTTP: GET /api/report                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 消息流

### WebSocket 握手流程

```
Client                           Server
  │                                │
  │───── handshake ────────────────▶│
  │                                │
  │◀──── handshake_ack ────────────│
  │                                │
  │         (连接建立)               │
```

### 实时更新流程

```
Application                       Server                      Client
    │                              │                            │
    │ UpdateMetrics()              │                            │
    │─────────────────────────────▶│                            │
    │                              │ Broadcast(metrics_updated) │
    │                              │───────────────────────────▶│
    │                              │                            │ Update UI
    │                              │                            │
    │ AddFrame()                   │                            │
    │─────────────────────────────▶│                            │
    │                              │ Broadcast(frame_added)    │
    │                              │───────────────────────────▶│
    │                              │                            │ Append to list
```

### HTTP 请求流程

```
Client (Browser)                  Server                     SnapshotMgr
    │                                │                            │
    │ GET /api/snapshots             │                            │
    │───────────────────────────────▶│                            │
    │                                │───── GetAll() ────────────▶│
    │                                │◀───── Snapshot list ───────│
    │◀──── JSON response ────────────│                            │
    │                                │                            │
    │ GET /api/diff?from=1&to=10    │                            │
    │───────────────────────────────▶│                            │
    │                                │───── Get(1), Get(10) ─────▶│
    │                                │◀───── Snapshots ───────────│
    │                                │ Compare()                  │
    │◀──── Changes JSON ────────────│                            │
```

## 功能概述

Protocol 模块是 DevTools 的**统一协议服务器**，整合了之前 `remote` 和 `client` 的功能，提供：

1. **核心类型系统** - DevTools 的基础类型定义
2. **WebSocket 服务器** - 实时双向通信
3. **HTTP REST API** - 查询接口
4. **Web 仪表盘** - 内嵌的调试 UI
5. **CDP 兼容** - Chrome DevTools Protocol 支持

## 快速开始

```go
import "github.com/wwsheng009/mint/devtools/protocol"
import "github.com/wwsheng009/mint/devtools/snapshot"

// 创建服务器
server := protocol.NewServer(protocol.ServerConfig{
    Port:              8080,
    EnableDashboard:   true,
    EnableTuiCommands: true,
    EnableCdp:         true,
})

// 配置快照管理器（可选）
snapshotMgr := snapshot.NewManager(1000)
server.SetSnapshotManager(snapshotMgr)

// 启动服务器
go server.Start()
defer server.Stop()

// 访问:
// - http://localhost:8080/     (Web 仪表盘)
// - ws://localhost:8080/ws     (WebSocket)
// - http://localhost:8080/api/* (REST API)
```

## 服务器配置

### ServerConfig

```go
type ServerConfig struct {
    Port              int   // 服务器端口 (默认: 8080)
    EnableDashboard   bool  // 启用 Web 仪表盘 UI
    EnableTuiCommands bool  // 启用 TUI 特定命令 (inspect, highlight)
    EnableCdp         bool  // 启用 Chrome DevTools Protocol 兼容
}
```

### 默认配置

```go
config := protocol.DefaultServerConfig()
// 等价于:
// Port: 8080
// EnableDashboard: true
// EnableTuiCommands: true
// EnableCdp: true
```

## HTTP API 端点

### 健康检查

```
GET /health
```

响应:
```json
{
  "status": "ok",
  "server": "mint-devtools",
  "version": "1.0.0",
  "port": 8080,
  "ws_clients": 1,
  "running": true,
  "capabilities": ["snapshots", "diffs", "metrics", "frames", "components", "tui_commands"],
  "snapshots": {
    "totalSnapshots": 100,
    "oldestFrameId": 0,
    "newestFrameId": 99
  }
}
```

### 性能指标

```
GET /api/metrics
```

响应:
```json
{
  "timestamp": "2026-01-31T09:00:28.5875172+08:00",
  "fps": 60.0,
  "frameTime": 16,
  "layoutTime": 5,
  "paintTime": 4,
  "memoryUsage": 55000000,
  "componentCount": 100,
  "frameCount": 50
}
```

### 帧列表

```
GET /api/frames
```

响应:
```json
[
  {
    "frameId": 1,
    "timestamp": "2026-01-31T09:00:18.1234567+08:00",
    "eventCount": 3,
    "mutationCount": 2,
    "layoutCount": 1,
    "repaintCount": 1,
    "durationMs": 16
  }
]
```

### 快照列表

```
GET /api/snapshots
```

响应:
```json
[
  {
    "id": "snap-1234567890",
    "frame_id": 10,
    "timestamp": "2026-01-31T09:00:30.1234567+08:00",
    "components": 25
  }
]
```

### 获取单个快照

```
GET /api/snapshot?frame={frame_id}
```

### 快照对比

```
GET /api/diff?from={frame_id}&to={frame_id}
```

响应:
```json
{
  "from": 10,
  "to": 20,
  "changes": [
    {
      "node_id": "node-5",
      "type": "added",
      "path": "",
      "old_value": null,
      "new_value": null
    },
    {
      "node_id": "node-3",
      "type": "modified",
      "path": "props.label",
      "old_value": "Old Label",
      "new_value": "New Label"
    }
  ]
}
```

## WebSocket 协议

### 连接

```
ws://localhost:8080/ws
```

### 客户端消息

#### 握手

```json
{
  "version": "1.0.0",
  "type": "handshake",
  "id": "client-123",
  "payload": {
    "client_id": "web-dashboard",
    "capabilities": ["snapshots", "metrics", "frames"],
    "version": "1.0.0",
    "protocol": "remote"
  }
}
```

#### 获取快照

```json
{
  "version": "1.0.0",
  "type": "get_snapshot",
  "id": "req-1",
  "payload": {
    "frame_id": 10,
    "include_state": true,
    "include_children": true
  }
}
```

#### 获取差异

```json
{
  "version": "1.0.0",
  "type": "get_diff",
  "id": "req-2",
  "payload": {
    "from": 10,
    "to": 20
  }
}
```

#### 检查节点 (TUI)

```json
{
  "version": "1.0.0",
  "type": "inspect",
  "id": "req-4",
  "payload": {
    "node_id": "node-5"
  }
}
```

#### 高亮节点 (TUI)

```json
{
  "version": "1.0.0",
  "type": "highlight",
  "id": "req-5",
  "payload": {
    "node_id": "node-5",
    "color": "red"
  }
}
```

### 服务器消息

#### 握手确认

```json
{
  "version": "1.0.0",
  "type": "handshake_ack",
  "id": "client-123"
}
```

#### 性能指标更新

```json
{
  "type": "metrics_updated",
  "data": {
    "timestamp": "2026-01-31T09:00:28.5875172+08:00",
    "fps": 60.0,
    "frameTime": 16,
    "layoutTime": 5,
    "paintTime": 4,
    "memoryUsage": 55000000,
    "componentCount": 100,
    "frameCount": 50
  }
}
```

#### 新帧记录

```json
{
  "type": "frame_added",
  "data": {
    "frameId": 11,
    "timestamp": "2026-01-31T09:00:28.1234567+08:00",
    "eventCount": 3,
    "mutationCount": 2,
    "layoutCount": 1,
    "repaintCount": 1,
    "durationMs": 16
  }
}
```

#### 组件更新

```json
{
  "type": "component_updated",
  "data": {
    "id": "component-5",
    "type": "Button",
    "properties": {...},
    "styles": {...},
    "visible": true,
    "focused": false
  }
}
```

## 核心 API

### 更新性能指标

```go
server.UpdateMetrics(&protocol.Metrics{
    FPS:            60.0,
    FrameTime:      16 * time.Millisecond,
    LayoutTime:     5 * time.Millisecond,
    PaintTime:      4 * time.Millisecond,
    MemoryUsage:    50 * 1024 * 1024,
    ComponentCount: 100,
    FrameCount:     1000,
})
```

### 添加帧

```go
server.AddFrame(&protocol.FrameData{
    FrameID:       devtools.FrameID(1),
    Timestamp:     time.Now(),
    Duration:      16 * time.Millisecond,
    EventCount:    5,
    MutationCount: 2,
    LayoutCount:   1,
    RepaintCount:  1,
})
```

### 更新组件

```go
server.UpdateComponent("btn-1", &protocol.DashboardComponentData{
    ID:       "btn-1",
    Type:     "Button",
    Properties: map[string]interface{}{
        "label": "Click Me",
    },
    Visible:  true,
    Focused:  false,
})
```

### 获取 WebSocket 服务器

```go
wsServer := server.GetWebSocketServer()

// 设置快照处理器
wsServer.SetSnapshotHandler(func(frameID devtools.FrameID) (*snapshot.Snapshot, bool) {
    return snapshotMgr.Get(frameID)
})

// 设置范围处理器
wsServer.SetRangeHandler(func(from, to devtools.FrameID) []*snapshot.Snapshot {
    return snapshotMgr.GetRange(from, to)
})
```

## 核心类型

### NodeID - 节点标识符

```go
type NodeID string

// 生成节点 ID
func GenerateNodeID(prefix string) NodeID
```

### FrameID - 帧标识符

```go
type FrameID int
```

### Rect - 矩形区域

```go
type Rect struct {
    X      int
    Y      int
    Width  int
    Height int
}

func (r Rect) Contains(x, y int) bool
func (r Rect) Intersects(other Rect) bool
func (r Rect) Area() int
```

### Metrics - 性能指标

```go
type Metrics struct {
    Timestamp      time.Time
    FPS            float64
    FrameTime      time.Duration
    LayoutTime     time.Duration
    PaintTime      time.Duration
    MemoryUsage    uint64
    ComponentCount int
    FrameCount     int
}
```

### FrameData - 帧数据

```go
type FrameData struct {
    FrameID       devtools.FrameID
    Timestamp     time.Time
    Duration      time.Duration
    EventCount    int
    MutationCount int
    LayoutCount   int
    RepaintCount  int
}
```

### DashboardComponentData - 组件数据

```go
type DashboardComponentData struct {
    ID         string
    Type       string
    Properties map[string]interface{}
    Styles     map[string]interface{}
    Children   []string
    Visible    bool
    Focused    bool
    Bounds     RectData
}
```

## Web 仪表盘

### 功能

- **仪表盘** - 性能指标概览、FPS/内存趋势图
- **指标** - 详细性能数据
- **帧** - 帧时间线
- **组件** - 组件树视图
- **快照** - 快照列表
- **对比** - 快照差异对比
- **报告** - 完整调试报告

### 界面特性

- 深色主题
- 实时 WebSocket 更新
- 响应式布局
- 中文界面

## 文件列表

| 文件 | 描述 |
|------|------|
| `server.go` | 统一协议服务器 |
| `websocket.go` | WebSocket 服务器 |
| `message.go` | 消息类型定义和版本 |
| `html.go` | Web 仪表盘 HTML |
| `types.go` | 核心数据类型 |
| `handler.go` | WebSocket 消息处理器 |
| `API.md` | 完整 API 文档 |

## 相关模块

| 模块 | 关系 |
|------|------|
| **devtools** | 核心类型定义 |
| **snapshot** | 快照系统 |
| **client** | 调试客户端 |
| **memory** | 内存优化 |
| **observation** | 观察层 |

## 使用示例

### 完整示例

```go
package main

import (
    "time"
    "github.com/wwsheng009/mint/devtools"
    "github.com/wwsheng009/mint/devtools/protocol"
    "github.com/wwsheng009/mint/devtools/snapshot"
)

func main() {
    // 创建快照管理器
    snapshotMgr := snapshot.NewManager(1000)

    // 创建服务器
    server := protocol.NewServer(protocol.ServerConfig{
        Port:              8080,
        EnableDashboard:   true,
        EnableTuiCommands: true,
    })
    server.SetSnapshotManager(snapshotMgr)

    // 启动服务器
    server.Start()
    defer server.Stop()

    // 模拟更新
    for i := 0; i < 100; i++ {
        // 更新指标
        server.UpdateMetrics(&protocol.Metrics{
            FPS:            60.0,
            FrameTime:      16 * time.Millisecond,
            MemoryUsage:    50 * 1024 * 1024,
            ComponentCount: 100,
            FrameCount:     i + 1,
        })

        // 添加帧
        server.AddFrame(&protocol.FrameData{
            FrameID:       devtools.FrameID(i),
            Timestamp:     time.Now(),
            Duration:      16 * time.Millisecond,
            EventCount:    3,
            MutationCount: 1,
        })

        time.Sleep(100 * time.Millisecond)
    }

    // 等待退出
    select {}
}
```
