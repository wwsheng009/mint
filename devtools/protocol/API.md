# Mint DevTools API 文档

版本: 1.0.0

## 概述

Mint DevTools 提供了 HTTP REST API 和 WebSocket 接口，用于 TUI 应用的实时调试和监控。

**服务器地址**: `http://localhost:8080` (默认)
**WebSocket**: `ws://localhost:8080/ws`

---

## HTTP REST API

### 1. 健康检查

检查服务器运行状态。

**端点**: `GET /health`

**响应**:
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

---

### 2. 性能指标 (Planned)

获取当前性能指标。

**端点**: `GET /api/metrics`

**响应**:
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

| 字段 | 类型 | 说明 |
|------|------|------|
| `fps` | float64 | 当前帧率 |
| `frameTime` | int64 | 帧时间(毫秒) |
| `layoutTime` | int64 | 布局时间(毫秒) |
| `paintTime` | int64 | 绘制时间(毫秒) |
| `memoryUsage` | uint64 | 内存使用(字节) |
| `componentCount` | int | 组件数量 |
| `frameCount` | int | 总帧数 |

---

### 3. 帧列表 (Planned)

获取所有已记录的帧。

**端点**: `GET /api/frames`

**响应**:
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

---

### 4. 组件列表 (Planned)

获取所有组件的状态。

**端点**: `GET /api/components`

**响应**:
```json
[
  {
    "id": "component-1",
    "type": "Button",
    "properties": {
      "label": "Click Me",
      "active": true
    },
    "styles": {
      "color": "#4ec9b0",
      "width": 120,
      "height": 30
    },
    "children": ["component-2"],
    "visible": true,
    "focused": false,
    "bounds": {
      "x": 10,
      "y": 10,
      "width": 120,
      "height": 30
    }
  }
]
```

---

### 5. 快照列表

获取所有组件树快照。

**端点**: `GET /api/snapshots`

**响应**:
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

---

### 6. 获取单个快照

获取指定帧的完整快照。

**端点**: `GET /api/snapshot?frame={frame_id}`

**参数**:
- `frame` (required): 帧ID

**响应**:
```json
{
  "frame_id": 10,
  "timestamp": "2026-01-31T09:00:30.1234567+08:00",
  "components": [
    {
      "node_id": "node-1",
      "type": "Container",
      "props": {"padding": 10},
      "state": {},
      "bounds": {"x": 0, "y": 0, "width": 800, "height": 600},
      "children": ["node-2", "node-3"],
      "visible": true,
      "focused": false
    }
  ]
}
```

---

### 7. 快照对比

对比两个快照之间的差异。

**端点**: `GET /api/diff?from={frame_id}&to={frame_id}`

**参数**:
- `from` (required): 起始帧ID
- `to` (required): 结束帧ID

**响应**:
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
    },
    {
      "node_id": "node-2",
      "type": "removed",
      "path": "",
      "old_value": null,
      "new_value": null
    }
  ]
}
```

**变更类型**:
- `added`: 新增节点
- `removed`: 删除节点
- `modified`: 修改节点属性

---

### 8. 调试报告 (Planned)

生成完整的调试报告。

**端点**: `GET /api/report`

**响应**:
```json
{
  "generatedAt": "2026-01-31T09:00:30.1234567+08:00",
  "metrics": {
    "fps": 60.0,
    "frameTime": 16,
    "componentCount": 100,
    "frameCount": 50
  },
  "frames": [
    {
      "frameId": 1,
      "timestamp": "2026-01-31T09:00:18.1234567+08:00",
      "events": 3,
      "mutations": 2,
      "layouts": 1
    }
  ],
  "components": [...]
}
```

---

### 9. 导出数据

导出所有调试数据。

**端点**: `GET /api/export`

**响应**:
```json
{
  "version": "1.0.0",
  "exported_at": "2026-01-31T09:00:30.1234567+08:00",
  "metrics": {...},
  "frames": [...],
  "components": [...]
}
```

---

### 10. 导入数据 (Planned)

导入调试数据。

**端点**: `POST /api/import`

**请求体**: 与 `/api/export` 响应格式相同

**响应**:
```json
{
  "success": true,
  "imported": {
    "frames_count": 100,
    "components_count": 50
  }
}
```

---

## WebSocket API

### 连接

**URL**: `ws://localhost:8080/ws`

连接后服务器会发送心跳消息，客户端应响应以保持连接。

---

### 服务器推送消息

#### 1. metrics_updated - 性能指标更新

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

#### 2. frame_added - 新帧记录

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

#### 3. component_updated - 组件更新

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

---

### 客户端请求消息

#### 1. handshake - 握手

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

#### 2. get_snapshot - 获取快照

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

#### 3. get_diff - 获取差异

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

#### 4. subscribe - 订阅事件

```json
{
  "version": "1.0.0",
  "type": "subscribe",
  "id": "req-3",
  "payload": {
    "event_type": "keypress"
  }
}
```

#### 5. inspect - 检查节点 (TUI)

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

#### 6. highlight - 高亮节点 (TUI)

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

#### 7. evaluate - 表达式求值

```json
{
  "version": "1.0.0",
  "type": "evaluate",
  "id": "req-6",
  "payload": {
    "expression": "component.props.label",
    "context": {"node_id": "node-5"},
    "frame_id": 10
  }
}
```

**响应 (evaluation_result)**:

```json
{
  "version": "1.0.0",
  "type": "evaluation_result",
  "id": "req-6",
  "payload": {
    "result": "Click Me",
    "type": "string",
    "frame_id": 10
  }
}
```

---

## 错误响应

所有 API 在错误时返回相应的 HTTP 状态码：

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

错误响应格式:
```json
{
  "error": "错误描述信息"
}
```

---

## UI 页面

### 主仪表盘
- **URL**: `http://localhost:8080/`
- **描述**: 主要的调试仪表盘，显示性能指标、帧列表、组件树等

### 调试检查器
- **URL**: `http://localhost:8080/debug`
- **描述**: 快照和差异对比检查器
