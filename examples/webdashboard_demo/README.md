# WebDashboard Demo

This example demonstrates how to use the `WebDashboard` for real-time debugging and monitoring of TUI applications.

## Features

- **Real-time Dashboard UI**: Modern web-based interface with dark theme
- **WebSocket Support**: Live updates pushed to connected clients
- **HTTP API**: RESTful endpoints for querying dashboard data
- **Performance Metrics**: FPS, frame time, memory usage tracking
- **Frame Timeline**: View recent frames with event and mutation counts
- **Component Tracking**: Monitor component state changes

## Running

```bash
cd examples/webdashboard_demo
go run main.go
```

Then open your browser to:

- **Dashboard**: http://localhost:8080/
- **Inspector**: http://localhost:8080/debug
- **Health**: http://localhost:8080/health

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Dashboard UI |
| `/debug` | GET | Dashboard UI (alias) |
| `/api/frames` | GET | Get all captured frames |
| `/api/metrics` | GET | Get current performance metrics |
| `/api/components` | GET | Get all components |
| `/api/report` | GET | Generate debug report |
| `/api/export` | GET | Export dashboard data as JSON |
| `/api/import` | POST | Import dashboard data |
| `/health` | GET | Health check |
| `/ws` | WebSocket | Real-time updates |

## WebSocket Events

The WebSocket server at `/ws` broadcasts the following event types:

- `frame` - New frame captured
- `metrics` - Performance metrics updated
- `component` - Component state changed

Example WebSocket message:

```json
{
  "type": "frame",
  "data": {
    "frame_id": 1,
    "timestamp": "2024-01-30T12:34:56Z",
    "duration": 16666666,
    "event_count": 3,
    "mutation_count": 2
  },
  "timestamp": "2024-01-30T12:34:56Z"
}
```

## Usage in Your Application

```go
import "github.com/wwsheng009/mint/devtools/client"

// Create and start dashboard
dashboard := client.NewWebDashboard(8080)
dashboard.Start()
defer dashboard.Stop()

// Add frame data
dashboard.AddFrame(&client.DashboardFrame{
    FrameID:   devtools.FrameID(1),
    Timestamp: time.Now(),
    Duration:  16 * time.Millisecond,
    EventCount: 5,
})

// Update metrics
dashboard.UpdateMetrics(&client.DashboardMetrics{
    FPS:          60.0,
    FrameTime:    16 * time.Millisecond,
    MemoryUsage:  50 * 1024 * 1024,
})
```

## Dashboard UI

The dashboard includes:

- **Header**: Status indicators (connection, client count, frame count)
- **Sidebar**: Navigation (Dashboard, Frames, Components, Report)
- **Metrics Panel**: Real-time performance metrics (FPS, Frame Time, Memory, Components)
- **Frames Table**: Recent frames with timing and event information
- **Real-time Updates**: WebSocket connection for live data

## Stopping

Press `Ctrl+C` to stop the server and exit.
