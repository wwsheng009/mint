# Layer Visual Demo - 交互式 Layer 可视化演示

## 概述

这是一个使用 `framework.App` 的**真实交互式应用**，展示 Fiber-first 多层渲染系统的正确性。

### 特点

✅ **使用 framework.App** - 完整的事件循环和渲染系统
✅ **自动启用 Fiber-first** - `SetRoot()` 自动触发新的多层渲染路径
✅ **真实组件** - 使用 `ui/components` 目录的 Text, Stack, Toast 组件
✅ **5 层 Toast** - 展示 Base, Overlay, Modal, Tooltip, Inspector 5 个层级

## 文件说明

| 文件 | 类型 | 说明 |
|-----|------|------|
| `main.go` | 交互式应用 | 使用 framework.App.Run() 进入交互模式 |
| `verify_render.go` | 非交互验证 | 单次渲染，输出渲染结果到控制台 |

## 运行方式

### 方式 1：非交互验证（推荐快速测试）

```bash
cd examples/fiber_firsts/layer_visual_demo
go run verify_render.go
```

**输出示例：**

```
配置:
  Render Mode: 1
  Fiber-First: true

================================================================================
渲染 Layer Visual (非交互模式)...
================================================================================

┌────────────────────────────────────────────────────────────────────────────────┐
|╔════════════════════════════════════════════════════════╗                      |
|║   Layer Visual Demo - Render Verification               ║                     |
|╚════════════════════════════════════════════════════════╝                      |
...
|    Layer 0: Base layer                                                         |
|    Layer 1: Overlay layer                                                      |
|    Layer 2: Modal layer                                                       |  🟡 (Warning - 黄色)
|    Layer 3: Tooltip layer                                                     |  🟣
|    Layer 4: Inspector layer (highest)                                         |  🟢 (Success - 绿色)
...
================================================================================
验证完成！
================================================================================
```

### 方式 2：交互式应用（完整体验）

```bash
cd examples/fiber_firsts/layer_visual_demo
go run .
# 或
go build . && ./layer_visual_demo
```

**输出说明：**

- 进入交互模式，应用程序会持续运行
- 按 **ESC** 或 **Ctrl+C** 退出

## Layer 颜色和对应类型

| Layer | 名称 | 颜色 | Toast 类型 | 说明 |
|------|------|------|-----------|------|
| 0 | Base | 🔴 红色 | - | 主要内容层（最底） |
| 1 | Overlay | 🔵 蓝色 | Info | 标准通知层 |
| 2 | Modal | 🟡 黄色 | Warning | 模态对话框层 |
| 3 | Tooltip | 🟣 紫色 | - | 悬浮提示层 |
| 4 | Inspector | 🟢 绿色 | Success | 调试/覆盖层（最高） |

## 技术实现

### fiberFirstPaint() 渲染流程

```
DeclarativeNode.Paint()
         ↓
   fiberFirstPaint()
         ↓
   ├─ Phase 1: Reconcile (VNode → Fiber)
   ├─ Phase 2: Layout (Fiber → LayoutBox)
   │
   └─ Phase 3: Paint
         ├─ Convert LayoutBox → PaintableLayout
         ├─ Build PaintablePlanes (按 Layer 分组)
         │
         └─ PaintPaintablePlanes() ← NEW PATH
               ↓
         按 renderOrder 顺序绘制：
         [Base, Overlay, Modal, Tooltip, Inspector]
```

### SetRoot() 自动启用

```go
// framework/app.go
func (a *App) SetRoot(comp component.Node) {
    if fiberFirstSetter, ok := comp.(interface{ SetRenderMode(int) }); ok {
        fiberFirstSetter.SetRenderMode(1) // RenderModeFiberFirst
    }
    a.injectComponentContext(comp)
}
```

### 使用的关键组件

```go
import (
    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/internal/render"
    "github.com/wwsheng009/mint/ui/components/text"
    "github.com/wwsheng009/mint/ui/components/stack"
    "github.com/wwsheng009/mint/ui/components/tooltip"
)

// 1. 创建 framework App
fwApp := framework.NewApp()

// 2. 创建 DeclarativeNode
declarativeNode := render.NewDeclarativeNodeFromFuncWithFiber(
    LayerVisualApp,
    fwApp,
)

// 3. 设置 Root（自动启用 Fiber-first）
fwApp.SetRoot(declarativeNode)

// 4. 运行应用
fwApp.Run()
```

## 验证要点

运行 `verify_render.go` 后，检查：

1. ✅ **Render Mode = 1** (Fiber-first)
2. ✅ **5 个 Toast 都渲染**
   - Layer 0 (红色)
   - Layer 1 (蓝色 - Info)
   - Layer 2 (黄色 - Warning)
   - Layer 3 (紫色)
   - Layer 4 (绿色 - Success)
3. ✅ **Toast 按顺序自上而下显示**（ToastManager 自动定位）
4. ✅ **Layer 信息正确传递**（通过 VNode → Fiber → LayoutBox → PaintableBox）

## Z-Order 渲染顺序

```
┌───────────────────────────────┐
│ 4. Inspector  (最高) ← 最后绘制 │
├───────────────────────────────┤
│ 3. Tooltip                   │
├───────────────────────────────┤
│ 2. Modal                     │
├───────────────────────────────┤
│ 1. Overlay                   │
├───────────────────────────────┤
│ 0. Base        (最低) ← 先绘制  │
└───────────────────────────────┘
    先绘制 ↑              后绘制 ↓
```

- **高层级后绘制**，覆盖同位置的低层级内容
- **低层级先绘制**，作为背景

## 相关资源

| 文件 | 说明 |
|-----|------|
| `docsArchive/cleanup-2026-05-19/docs/layer/FIBER_FIRST_AUTO_ENABLE.md` | framework.App 自动启用 Fiber-first 机制的历史说明 |
| `examples/fiber_firsts/layer_zorder_test/` | Layer Z-Order 详细测试 |
| `examples/fiber_firsts/tooltip_demo/` | Tooltip/Toast 组件示例 |
| `internal/render/declarative_node.go:424-504` | fiberFirstPaint() 实现 |
| `internal/render/paint_engine.go:496` | PaintPaintablePlanes() 实现 |

## 故障排除

### 找不到 `ui.Text` 方法

使用 `ui/components` 组件而不是 runtime/ui VNode API：
```go
// ❌ 错误
ui.Text("Hello")

// ✅ 正确
newtext.New("Hello")  // from ui/components/text
```

### 编译错误：import cycle

避免在 `framework` 包的测试中直接导入 `internal/render`：
- 将测试放在 `internal/render_test` 包
- 或放在独立目录

### Fiber-first 未启用

检查：
1. 是否使用 `render.NewDeclarativeNodeFromFuncWithFiber()`
2. 是否使用 `fwApp.SetRoot(declarativeNode)`
3. 或设置环境变量 `MINT_FIBER_FIRST=true`
