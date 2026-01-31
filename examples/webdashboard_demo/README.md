# WebDashboard Demo

This example demonstrates how to use the unified `protocol.Server` for real-time debugging and monitoring of TUI applications.

## Features

- **Real-time Dashboard UI**: Modern web-based interface with dark theme
- **WebSocket Support**: Live updates pushed to connected clients
- **HTTP API**: RESTful endpoints for querying dashboard data
- **Performance Metrics**: FPS, frame time, memory usage tracking
- **Frame Timeline**: View recent frames with event and mutation counts
- **Component Tracking**: Monitor component state changes
- **Snapshots & Diff**: Component tree snapshots and diff comparison

## Running

```bash
cd examples/webdashboard_demo
go run main.go
```

Then open your browser to:

- **Dashboard**: http://localhost:8080/
- **Health**: http://localhost:8080/health

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Main Dashboard UI |
| `/health` | GET | Health check and server info |
| `/api/frames` | GET | Get all captured frames |
| `/api/metrics` | GET | Get current performance metrics |
| `/api/components` | GET | Get all components |
| `/api/snapshots` | GET | Get all snapshots |
| `/api/snapshot?frame={id}` | GET | Get specific snapshot |
| `/api/diff?from={id}&to={id}` | GET | Compare two snapshots |
| `/api/report` | GET | Generate debug report |
| `/api/export` | GET | Export dashboard data as JSON |
| `/api/import` | POST | Import dashboard data |
| `/ws` | WebSocket | Real-time updates |

## WebSocket Events

The WebSocket server at `/ws` supports the following message types:

### Client → Server

| Type | Description |
|------|-------------|
| `handshake` | Initial connection handshake |
| `get_snapshot` | Request a specific snapshot |
| `get_diff` | Request diff between frames |
| `subscribe` | Subscribe to event types |
| `inspect` | Inspect a component node (TUI) |
| `highlight` | Highlight a component (TUI) |
| `evaluate` | Evaluate an expression |

### Server → Client

| Type | Description |
|------|-------------|
| `handshake_ack` | Handshake confirmation |
| `metrics_updated` | Performance metrics update |
| `frame_added` | New frame captured |
| `component_updated` | Component state changed |
| `evaluation_result` | Expression evaluation result |

Example WebSocket handshake:

```json
{
  "version": "1.0.0",
  "type": "handshake",
  "id": "web-dashboard",
  "payload": {
    "client_id": "web-dashboard",
    "capabilities": ["snapshots", "metrics", "frames"],
    "version": "1.0.0",
    "protocol": "remote"
  }
}
```

## Usage in Your Application

```go
import "github.com/wwsheng009/mint/devtools/protocol"
import "github.com/wwsheng009/mint/devtools/snapshot"

// Create and start server
server := protocol.NewServer(protocol.ServerConfig{
    Port:              8080,
    EnableDashboard:   true,
    EnableTuiCommands: true,
})

// Set snapshot manager
snapshotMgr := snapshot.NewManager(1000)
server.SetSnapshotManager(snapshotMgr)

// Start server
go server.Start()
defer server.Stop()

// Update metrics
server.UpdateMetrics(&protocol.Metrics{
    FPS:            60.0,
    FrameTime:      16 * time.Millisecond,
    MemoryUsage:    50 * 1024 * 1024,
    ComponentCount: 100,
})

// Add frame
server.AddFrame(&protocol.FrameData{
    FrameID:       1,
    Timestamp:     time.Now(),
    Duration:      16 * time.Millisecond,
    EventCount:    5,
    MutationCount: 2,
})
```

## Dashboard UI

The dashboard includes:

- **Header**: Status indicators (FPS, frame time, component count, connection status)
- **Sidebar Navigation**:
  - Dashboard - Overview with metrics and charts
  - Metrics - Detailed performance metrics
  - Frames - Frame timeline
  - Components - Component tree
  - Snapshots - Snapshot list
  - Diff - Compare two snapshots
  - Report - Full debug report
- **Real-time Updates**: WebSocket connection for live data
- **Dark Theme**: Modern, easy-on-eyes interface

## Stopping

Press `Ctrl+C` to stop the server and exit.
