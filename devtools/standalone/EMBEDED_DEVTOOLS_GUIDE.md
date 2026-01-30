# Engine 示例 嵌入式 DevTools 调试指南

本指南展示如何使用 DevTools 来调试 `examples/engine/main.go` 程序。

## 问题诊断

当 `main.go` 无法正常工作时，可以使用 `main_with_devtools.go` 来进行调试。

### 使用方法

```bash
# 运行带 DevTools 的版本
cd examples/engine
go run main_with_devtools.go
```

## DevTools 集成说明

### 1. 初始化 DevTools

```go
import "github.com/wwsheng009/mint/devtools"

var dt *devtools.DevTools

func init() {
    dt = devtools.New()
}
```

### 2. 启用 DevTools

```go
func EngineExample() error {
    // 启用 DevTools
    dt.Enable()
    defer dt.Disable()

    // ... 应用逻辑 ...
}
```

### 3. 记录事件

DevTools 自动记录以下事件：

#### 组件添加事件
```go
func (r *Root) AddButton(btn *Button) {
    r.buttons = append(r.buttons, btn)

    if dt.IsEnabled() {
        dt.RecordEvent("component_add", btn.id, "target", map[string]interface{}{
            "type":   "Button",
            "text":   btn.text,
            "x":      btn.x,
            "y":      btn.y,
            "width":  btn.width,
            "height": btn.height,
        })
    }
}
```

#### 鼠标事件
```go
func (b *Button) HandleMouse(ev *event.MouseEvent, localX, localY int) bool {
    if dt.IsEnabled() {
        eventData := map[string]interface{}{
            "button_id": b.id,
            "local_x":   localX,
            "local_y":   localY,
            "type":      string(ev.Type),
            "click":     clickType,
        }
        dt.RecordEvent("mouse", b.id, "target", eventData)
    }
    // ...
}
```

#### 焦点事件
```go
func (b *Button) SetFocus(focus bool) {
    b.focused = focus

    if dt.IsEnabled() {
        dt.RecordEvent("focus", b.id, "target", map[string]interface{}{
            "focused": focus,
        })
    }
    // ...
}
```

#### 日志事件
```go
func logEvent(format string, args ...interface{}) {
    msg := fmt.Sprintf(format, args...)

    if dt.IsEnabled() {
        dt.RecordEvent("log", "system", "target", map[string]interface{}{
            "message": msg,
        })
    }
}
```

### 4. 查看调试输出

运行程序后，所有事件都会在控制台输出：

```
[DEBUG] Engine started, waiting for input...
[DEBUG] DevTools enabled
[DEBUG] component_add: btn1 added at (5, 7)
[DEBUG] component_add: btn2 added at (25, 7)
[DEBUG] component_add: btn3 added at (45, 7)
[DEBUG] Button btn1 received mouse event: Type=press Click=left Pos=(2,1)
[DEBUG] Button btn1 CLICKED! Triggering callback
[DEBUG] >>> Button 1 CLICKED! <<<
```

## 常见问题诊断

### 问题 1: 鼠标点击无响应

**症状**: 点击按钮没有任何反应

**诊断方法**:
1. 查看 `[DEBUG] Button xxx received mouse event` 输出
2. 如果没有此输出，说明鼠标事件未到达组件
3. 检查 `BuildLayoutBoxes()` 是否正确设置

**检查点**:
```go
// 确保布局框已设置
boxes := root.BuildLayoutBoxes()
eng.SetLayoutBoxes(boxes)
```

### 问题 2: 焦点状态不更新

**症状**: 按钮焦点状态（黄色边框）不显示

**诊断方法**:
1. 查看 `[DEBUG] Button xxx SetFocus` 输出
2. 如果 `SetFocus` 被调用但 UI 不更新，检查 `Paint` 方法

**检查点**:
```go
// Paint 方法应该读取 focused 状态
func (b *Button) Paint(buf *paint.Buffer) {
    borderColor := style.White
    if b.focused {
        borderColor = style.Yellow
    }
    // ...
}
```

### 问题 3: 事件丢失

**症状**: 某些事件没有被处理

**诊断方法**:
1. DevTools 记录所有事件
2. 对比预期事件和实际记录的事件

**事件流示例**:
```
component_add → mousemove → focus → mousepress → click
```

## DevTools API 参考

### 核心方法

| 方法 | 描述 |
|------|------|
| `devtools.New()` | 创建新的 DevTools 实例 |
| `dt.Enable()` | 启用 DevTools |
| `dt.Disable()` | 禁用 DevTools |
| `dt.IsEnabled()` | 检查是否启用 |
| `dt.RecordEvent(type, targetID, phase, data)` | 记录事件 |
| `dt.BeginFrame()` | 标记帧开始 |
| `dt.EndFrame()` | 标记帧结束 |
| `dt.CollectLayout(result)` | 收集布局数据 |
| `dt.CollectRepaint(dirtyRegions, changed, total)` | 收集重绘数据 |

### 事件数据格式

```go
// 事件数据包含
map[string]interface{}{
    "button_id": "btn1",     // 目标组件 ID
    "local_x":   10,          // 本地 X 坐标
    "local_y":   5,           // 本地 Y 坐标
    "type":      "press",     // 事件类型
    "click":     "left",      // 鼠标按钮
    "focused":   true,        // 焦点状态
    "message":   "...",       // 日志消息
}
```

## 调试工作流

1. **启用 DevTools**: 在程序开始时调用 `dt.Enable()`

2. **添加事件记录**: 在关键位置添加 `dt.RecordEvent()`

3. **运行程序**: `go run main_with_devtools.go`

4. **观察输出**: 检查控制台的 `[DEBUG]` 输出

5. **分析事件流**: 确认事件按预期顺序发生

6. **定位问题**: 根据缺失或异常的事件定位问题

## 示例输出

```
=== Engine Example with DevTools ===
Components: 3 Buttons

Controls:
  - Mouse move: Shows mouse position
  - Mouse click: Click buttons (green = hover, yellow = focused)
  - ESC or Ctrl+C: Exit

DevTools Features:
  - Event tracking: All events are recorded
  - Check console for debug output
  - Events are logged with timestamps

[DEBUG] Engine started, waiting for input...
[DEBUG] DevTools enabled
[DEBUG] component_add: Button btn1 added
[DEBUG] component_add: Button btn2 added
[DEBUG] component_add: Button btn3 added
[DEBUG] Mouse move: (10, 8)
[DEBUG] Button btn1 SetFocus(true)
[DEBUG] Button btn1 received mouse event: Type=press Click=left Pos=(2,1)
[DEBUG] Button btn1 CLICKED! Triggering callback
[DEBUG] >>> Button 1 CLICKED! <<<
```

## 进阶使用

### 自定义事件类型

```go
// 定义自定义事件类型
const (
    EventCustomAction = "custom_action"
    EventStateChange  = "state_change"
)

// 记录自定义事件
dt.RecordEvent(EventCustomAction, "my-component", "target", map[string]interface{}{
    "action": "do_something",
    "param":  42,
})
```

### 时间旅行调试

DevTools 支持快照和时间旅行功能：

```go
import "github.com/wwsheng009/mint/devtools/snapshot"

// 创建快照管理器
snapshotMgr := snapshot.NewManager()

// 在关键点创建快照
snap := snapshot.NewSnapshot(frameID)
snap.Capture(dt)
snapshotMgr.Save(snap)

// 比较两个状态
diff := snapshotMgr.Compare(frame1, frame2)
```

### 远程调试

使用 WebSocket 连接进行远程调试：

```go
import "github.com/wwsheng009/mint/devtools/remote"

// 启动远程调试服务器
server := remote.NewDevToolsServer(9222, dt, snapshotMgr)
go server.Start()

// 在浏览器中访问 http://localhost:9222/debug
```

## 相关文档

- [DevTools 主文档](../../devtools/docs/README.md)
- [快照系统](../../devtools/snapshot/readme.md)
- [时间旅行](../../devtools/timetravel/readme.md)
- [远程调试](../../devtools/remote/readme.md)
