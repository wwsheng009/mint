# Fiber-first 渲染模式自动启用

## 概述

`framework.App` 和 `internal/render.DeclarativeNode` 已经改进，**自动启用 Fiber-first 渲染模式**，确保新的多层渲染路径（`PaintPaintablePlanes`）被正确使用。

## 关键改进

### 1. `App.SetRoot()` 自动检测并启用 Fiber-first

```go
// framework/app.go
func (a *App) SetRoot(comp component.Node) {
    a.root = comp
    a.dirty = true

    // 自动启用 Fiber-first 模式：如果 root 支持 SetRenderMode 接口
    if fiberFirstSetter, ok := comp.(interface{ SetRenderMode(int) }); ok {
        // 设置 Fiber-first 渲染模式
        fiberFirstSetter.SetRenderMode(1) // RenderModeFiberFirst
        log.UILogger.Debug("[App.SetRoot] ✅ Auto-enabled Fiber-first render mode (Multi-Layer Support)")
    }

    a.injectComponentContext(comp)
}
```

### 2. `DeclarativeNode.SetRenderMode()` 实现

```go
// internal/render/declarative_node.go
func (n *DeclarativeNode) SetRenderMode(mode RenderMode) {
    n.mu.Lock()
    defer n.mu.Unlock()
    n.renderMode = mode
    if mode == RenderModeFiberFirst {
        n.fiberFirstEnabled = true
        // 初始化新的 layout engine 和 paint engine
        if n.newLayoutEngine == nil {
            n.newLayoutEngine = NewNewLayoutEngineAdapter()
        }
        if n.paintEngine == nil {
            n.paintEngine = NewPaintEngine()
        }
    }
}
```

### 3. 自动启用条件

满足以下任一条件时，会自动使用 Fiber-first 渲染路径：

1. **使用 `NewDeclarativeNodeFromFuncWithFiber` 创建 DeclarativeNode**
2. **使用 `ui.RunTest()` 或 `ui.RunTestWithSandbox()`**

## 渲染路径流程

```
User Code                                  framework.App              internal/render
────────────────────────────────────────────────────────────────────────────────────────
fwApp.SetRoot(declarativeNode)         →  SetRoot(comp)         →  comp.SetRenderMode(1)
                                          └───────────────────────────────────────→
                                    DeclarativeNode.SetRenderMode(1)
                                          └───────────────────────────────────────→
                                    NewPaintEngine()
                                    NewNewLayoutEngineAdapter()

                                    ┌───────────────────────────────────────────────────┐
                                    │  Paint() Method Execution Flow                   │
                                    ├───────────────────────────────────────────────────┤
                                    │                                                   │
                                    │  if fiberFirstEnabled && useFiber {              │
                                    │      case RenderModeFiberFirst:                  │
                                    │          fiberFirstPaint(ctx, buf)               │
                                    │                                                   │
                                    │          Phase 1: Reconcile (VNode → Fiber)     │
                                    │          Phase 2: Layout (Fiber → LayoutBox)     │
                                    │          Phase 3: Paint:                         │
                                    │                   ┌─────────────────────────┐   │
                                    │                   │ Convert LayoutBox to   │   │
                                    │                   │ PaintableLayout         │   │
                                    │                   └─────────────────────────┘   │
                                    │                           ↓                      │
                                    │                   ┌─────────────────────────┐   │
                                    │                   │ Build PaintablePlanes  │   │
                                    │                   │ (group by Layer)       │   │
                                    │                   └─────────────────────────┘   │
                                    │                           ↓                      │
                                    │                   ┌─────────────────────────┐   │
                                    │                   │ PaintPaintablePlanes() │   │
                                    │                   │ ← NEW PATH             │   │
                                    │                   └─────────────────────────┘   │
                                    │  }                                              │
                                    └───────────────────────────────────────────────────┘
```

## 多层渲染关键实现

### PaintPaintablePlanes() 执行逻辑

```go
// internal/render/paint_engine.go
func (e *PaintEngine) PaintPaintablePlanes(planes *paint.PaintablePlanes, buf *paint.Buffer) error {
    // 按 renderOrder 顺序绘制各层
    for _, layer := range planes.GetRenderOrder() {
        layerBoxes := planes.GetLayer(layer)
        // 按 Y 坐标排序（优化渲染顺序）
        sort.Slice(layerBoxes, func(i, j int) bool {
            return layerBoxes[i].Y < layerBoxes[j].Y
        })
        // 绘制该层的所有 box
        for _, box := range layerBoxes {
            e.PaintLayout(box, buf)
        }
    }
    return nil
}
```

### renderOrder 定义

```go
// runtime/paint/paintable_planes.go
func NewPaintablePlanes() *PaintablePlanes {
    return &PaintablePlanes{
        planes: make(map[RenderLayer][]*PaintableBox),
        renderOrder: []RenderLayer{
            RenderLayerBase,      // 0
            RenderLayerOverlay,   // 1
            RenderLayerModal,     // 2
            RenderLayerTooltip,   // 3
            RenderLayerInspector, // 4
        },
    }
}
```

## 使用示例

### 方式 1：使用 framework.App（推荐）

```go
package main

import (
    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/internal/render"
    "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
    // 1. 创建 framework App
    fwApp := framework.NewApp()
    fwApp.Resize(80, 24)

    // 2. 创建 DeclarativeNode
    declarativeNode := render.NewDeclarativeNodeFromFuncWithFiber(myApp, fwApp)

    // 3. 设置 Root（自动启用 Fiber-first 模式）
    fwApp.SetRoot(declarativeNode)  // ← 自动调用 SetRenderMode(1)

    // 4. 运行 App
    fwApp.Run()
}

func myApp() ui.VNode {
    return ui.VStack(
        ui.Text("Using Multi-Layer Rendering"),
    )
}
```

### 方式 2：使用测试框架

```go
func TestLayerRendering(t *testing.T) {
    // 自动使用 Fiber-first 渲染模式
    ta, err := ui.RunTestWithSandbox(app, ui.WithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer ta.Close()

    ta.ForceRender()

    // 验证多层渲染内容
    ta.AssertContains("Layer 0 Toast")
    ta.AssertContains("Layer 4 Toast")
}
```

## 多层组件示例

### Tooltip（Layer 3）

```go
tooltip.New().
    SetContent("This tooltip renders at Layer 3").
    SetPosition(ui.TooltipPositionTop).
    Build()
```

**默认 Layer**: `LayerTooltip (3)`

### Toast（可自定义 Layer）

```go
toast.NewToastBuilder("Custom Layer 2 Toast").
    Layer(rtui.LayerModal).  // 设置为 Modal 层
    Build()
```

**默认 Layer**: `LayerOverlay (1)`  
**Layer 选项**: 0-4

## 验证方法

### 1. 检查渲染模式

```go
fmt.Printf("Render Mode: %v\n", node.GetRenderMode())
// 输出: Render Mode: 1 (FiberFirst)
```

### 2. 检查 Fiber-first 是否启用

```go
fmt.Printf("Fiber-First Enabled: %v\n", node.IsFiberFirstEnabled())
// 输出: Fiber-First Enabled: true
```

### 3. 运行 Layer 测试

```bash
cd examples/fiber_firsts/layer_zorder_test
go run main.go

# 验证要点：
# ✓ Render Mode: 1
# ✓ Fiber-First Enabled: true
# ✓ 5层 Toast 按顺序渲染: Layer 0 → Layer 4
```

## Layer 渲染顺序

```
┌───────────────────────────────┐
│ 4. Inspector  (最高)          │ ← 最后绘制，覆盖所有
├───────────────────────────────┤
│ 3. Tooltip                    │ ← 悬浮提示层
├───────────────────────────────┤
│ 2. Modal                      │ ← 模态对话框层
├───────────────────────────────┤
│ 1. Overlay                    │ ← 标准 Toast 层
├───────────────────────────────┤
│ 0. Base        (最低)         │ ← 主要内容层
└───────────────────────────────┘
    先绘制 ↑                           后绘制 ↓
```

## 修复历史

| 问题 | 根本原因 | 修复方案 |
|-----|---------|---------|
| Fiber-first 多层渲染不工作 | `fiberFirstPaint()` 使用 `PaintLayout()`，忽略 Layer | 改用 `PaintPaintablePlanes()`，按 renderOrder 绘制 |
| 需要手动设置环境变量 | `MINT_FIBER_FIRST` 必须显式设置 | `App.SetRoot()` 自动调用 `SetRenderMode(1)` |
| 代码不直观 | 新旧路径切换需要手动控制 | 自动检测并启用 Fiber-first 模式 |

## 相关文件

| 文件 | 职责 |
|-----|------|
| `framework/app.go` | App.SetRoot() 自动启用 Fiber-first |
| `internal/render/declarative_node.go` | fiberFirstPaint() 使用 PaintPaintablePlanes() |
| `internal/render/paint_engine.go` | PaintPaintablePlanes() 执行逻辑 |
| `runtime/paint/paintable_planes.go` | renderOrder 定义 |
| `examples/fiber_firsts/layer_zorder_test/` | Layer Z-Order 验证 |
| `examples/fiber_firsts/tooltip_demo/` | 真实组件 + Layer 演示 |
