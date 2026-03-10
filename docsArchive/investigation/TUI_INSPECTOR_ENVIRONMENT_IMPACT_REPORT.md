# TUI_INSPECTOR 环境变量影响分析报告

## 概述

`TUI_INSPECTOR` 环境变量控制 Inspector 的启用和自动显示。本报告分析了所有受该环境变量影响的关键代码路径。

---

## 环境变量检查点

### 1. 应用启动时检查 - `main.go`

**文件**: `examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go`

**代码位置**: 第 46-51 行

```go
// Auto-show inspector from environment
if os.Getenv("TUI_INSPECTOR") == "true" {
    globalInspector.ToggleVisibility()  // 设置 visible = true
    fmt.Println("UI Inspector auto-enabled - Press F12 or Ctrl+D to toggle")
} else {
    fmt.Println("UI Inspector ready - Press [I] button or F12/Ctrl+D to toggle")
}
```

**影响**:
- `TUI_INSPECTOR=true` → 调用 `ToggleVisibility()` → `inspector.visible = true`
- `TUI_INSPECTOR=false` 或未设置 → `inspector.visible = false` (默认)

**状态变化**:
```
┌─────────────────────────────────────────────────────┐
│ TUI_INSPECTOR=true                                 │
│   ↓                                               │
│ globalInspector.ToggleVisibility()                │
│   ↓                                               │
│ inspector.visible = true                          │
└─────────────────────────────────────────────────────┘
```

---

## 渲染流程关键点

### 2. VNode 树构建 - `RuntimeDemoWithInspectorOverlay()`

**文件**: `examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go`

**代码位置**: 第 125-157 行

```go
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    // 获取 Inspector 可见性状态
    inspectorVisible := globalInspector.IsVisible()  // ← 关键检查点

    // 构建 appContent
    appContent := buildDemoContent(...)

    // 根据 Inspector 可见性决定返回的 VNode 树
    if inspectorVisible {
        // TUI_INSPECTOR=true → 走这个分支
        inspectorOverlay := globalInspector.RenderOverlay()
        return ui.Fragment(               // ← 修复后使用 Fragment
            appContent,
            inspectorOverlay,
        )
    }

    // TUI_INSPECTOR=false → 直接返回 appContent
    return appContent
}
```

**数据流**:

```
TUI_INSPECTOR=true
  ↓
inspector.visible = true
  ↓
IsVisible() = true
  ↓
inspectorVisible = true
  ↓
if inspectorVisible { ... }  ← 进入此分支
  ↓
return Fragment(appContent, inspectorOverlay)
  ↓
VNode 树包含 LayerInspector 节点
```

**对比**:

```
TUI_INSPECTOR=false
  ↓
inspector.visible = false
  ↓
IsVisible() = false
  ↓
inspectorVisible = false
  ↓
return appContent
  ↓
VNode 树不包含任何 layer 节点
```

---

### 3. Layer 检测 - `PipelineRenderer.hasLayerNodes()`

**文件**: `internal/render/pipeline_renderer.go`

**代码位置**: 第 105-124 行

```go
func (r *PipelineRenderer) hasLayerNodes(vnode rtui.VNode) bool {
    if vnode == nil {
        return false
    }

    // 检查当前节点
    layer := vnode.GetLayer()
    if layer != rtui.LayerBase && layer.IsValid() {
        return true  // ← 检测到 LayerInspector
    }

    // 递归检查子节点
    for _, child := range vnode.Children() {
        if r.hasLayerNodes(child) {
            return true
        }
    }

    return false
}
```

**影响**:

```
TUI_INSPECTOR=true
  ↓
VNode 树 = Fragment(appContent, inspectorOverlay)
  ↓
hasLayerNodes(Fragment) 递归检查
  ↓
发现 inspectorOverlay.GetLayer() = LayerInspector
  ↓
返回 true → hasLayers = true
  ↓
调用 RenderLayers() 而不是 Render()
```

**对比**:

```
TUI_INSPECTOR=false
  ↓
VNode 树 = appContent
  ↓
hasLayerNodes(appContent) 递归检查
  ↓
所有节点都是 LayerBase
  ↓
返回 false → hasLayers = false
  ↓
调用 Render() 标准渲染流程
```

---

### 4. 渲染路径选择 - `PipelineRenderer.Render()`

**文件**: `internal/render/pipeline_renderer.go`

**代码位置**: 第 69-82 行

```go
var err error
if hasLayers {
    // TUI_INSPECTOR=true → hasLayers = true → 走这里
    if r.debug || os.Getenv("TUI_DEBUG_RENDER") == "true" {
        fmt.Fprintf(os.Stderr, "[PipelineRenderer] Using RenderLayers for multi-layer rendering\n")
    }
    err = r.pipeline.RenderLayers(vnode, constraints, buf)
} else {
    // TUI_INSPECTOR=false → hasLayers = false → 走这里
    if r.debug || os.Getenv("TUI_DEBUG_RENDER") == "true" {
        fmt.Fprintf(os.Stderr, "[PipelineRenderer] Using standard Render\n")
    }
    err = r.pipeline.Render(vnode, constraints, buf)
}
```

**两条渲染路径**:

#### 路径 A: `TUI_INSPECTOR=false` (标准渲染)

```
PipelineRenderer.Render()
  ↓
hasLayers = false
  ↓
RenderingPipeline.Render(vnode, constraints, buffer)
  ↓
LayoutEngine.Layout(appContent, constraints)
  ↓
PaintEngine.Paint(layout, buffer)
  ↓
✅ 正常显示
```

#### 路径 B: `TUI_INSPECTOR=true` (多 layer 渲染)

```
PipelineRenderer.Render()
  ↓
hasLayers = true
  ↓
RenderingPipeline.RenderLayers(vnode, constraints, buffer)
  ↓
LayerManager.CollectAndLayout(vnode, constraints, engine)
  ├─ Collector.Collect(vnode)
  ├─ StripLayers(vnode)
  ├─ Layout(baseTree, constraints)
  └─ Layout(inspectorOverlay, constraints)
  ↓
PaintEngine.PaintLayers(layouts, buffer)
  ├─ Paint(baseLayout)
  └─ Paint(inspectorLayout)
  ↓
✅ 显示 (如果修复正确)
```

---

## 关键配置信息

### Inspector 配置

**文件**: `internal/inspector/standalone_inspector.go`

**代码位置**: 第 114-132 行

```go
func NewStandaloneInspector() *StandaloneInspector {
    defaultX := 40  // 屏幕宽度 120 - Inspector 宽度 80 = 40
    defaultY := 5   // 顶部边距

    return &StandaloneInspector{
        enabled:       false,
        visible:       false,          // ← 默认不可见
        activeTab:     TabElements,
        overlayWidth:  80,             // ← Inspector 宽度
        overlayHeight: 25,             // ← Inspector 高度
        floatX:        defaultX,       // ← X 位置 (已修复)
        floatY:        defaultY,       // ← Y 位置
    }
}
```

### Inspector Overlay 创建

**文件**: `internal/inspector/standalone_inspector.go`

**代码位置**: 第 253-291 行

```go
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
    // Build overlay content
    content := si.buildOverlayContent()

    // IMPORTANT: SetProps BEFORE SetLayer
    content.SetProps(ui.Props{
        "x": si.floatX,  // ← Inspector X 位置 (默认 40)
        "y": si.floatY,  // ← Inspector Y 位置 (默认 5)
    })

    // Mark as Inspector layer
    content.SetLayer(rtui.LayerInspector)  // ← 关键：设置为 Inspector layer

    return content
}
```

---

## 问题根源分析

### 问题 1: VStack 创建额外容器 (已修复)

**原始代码** (问题):

```go
if inspectorVisible {
    inspectorOverlay := globalInspector.RenderOverlay()
    return ui.VStack(          // ← 问题：创建额外的布局容器
        appContent,
        inspectorOverlay,
    )
}
```

**结果 VNode 树**:

```
VStack (新增容器) ← 外层
  └─ appContent (VStack) ← 内层
```

**问题**:
1. StripLayers 返回 `VStack{appContent}` (嵌套结构)
2. LayoutEngine 布局嵌套 VStack 时高度计算错误
3. 导致界面不显示 ❌

**修复代码** (解决方案):

```go
if inspectorVisible {
    inspectorOverlay := globalInspector.RenderOverlay()
    return ui.Fragment(          // ← 解决方案：使用虚拟容器
        appContent,
        inspectorOverlay,
    )
}
```

**结果 VNode 树**:

```
Fragment (虚拟容器，不参与布局)
  ├─ appContent
  └─ inspectorOverlay
```

**优势**:
1. StripLayers 返回 `appContent` (扁平结构)
2. LayoutEngine 直接布局 appContent
3. 高度计算正确 ✅

---

## 完整的数据流图

### TUI_INSPECTOR=false (正常工作)

```
┌─────────────────────────────────────────────────────────────┐
│ 1. main(): TUI_INSPECTOR=false                             │
│    inspector.visible = false                                │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. RuntimeDemo()                                            │
│    IsVisible() = false                                      │
│    return appContent                                        │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. PipelineRenderer.Render()                                │
│    hasLayerNodes(appContent) = false                        │
│    → Use standard Render()                                  │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. RenderingPipeline.Render()                              │
│    LayoutEngine.Layout(appContent, constraints)             │
│    → Correct layout calculated                              │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. PaintEngine.Paint()                                      │
│    → Render appContent to buffer                            │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 6. Terminal                                                 │
│    ✅ Interface displays correctly                          │
└─────────────────────────────────────────────────────────────┘
```

### TUI_INSPECTOR=true (修复后正常工作)

```
┌─────────────────────────────────────────────────────────────┐
│ 1. main(): TUI_INSPECTOR=true                              │
│    ToggleVisibility() → inspector.visible = true            │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. RuntimeDemo()                                            │
│    IsVisible() = true                                       │
│    return Fragment(appContent, inspectorOverlay)            │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. PipelineRenderer.Render()                                │
│    hasLayerNodes(Fragment) = true (has LayerInspector)      │
│    → Use RenderLayers()                                     │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. RenderingPipeline.RenderLayers()                        │
│    LayerManager.CollectAndLayout()                          │
│      ├─ Collect: Extract LayerBase & LayerInspector        │
│      ├─ StripLayers: Remove inspector from tree            │
│      ├─ Layout(appContent) → baseLayout                    │
│      └─ Layout(inspectorOverlay) → inspectorLayout         │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. PaintEngine.PaintLayers()                               │
│    ├─ Paint(baseLayout) at (0, 0)                          │
│    └─ Paint(inspectorLayout) at (40, 5)                    │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 6. Terminal                                                 │
│    ✅ Interface displays correctly                          │
│    ✅ Inspector overlay visible on right side               │
└─────────────────────────────────────────────────────────────┘
```

---

## 修复总结

### 修复内容

**文件**: `examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go`

**修改行**: 第 151-154 行

**修改前**:
```go
return ui.VStack(
    appContent,
    inspectorOverlay,
)
```

**修改后**:
```go
return ui.Fragment(
    appContent,
    inspectorOverlay,
)
```

### 为什么修复有效

| 方面 | VStack (问题) | Fragment (修复) |
|------|--------------|----------------|
| **创建布局节点** | ✅ 是 (LayoutNode) | ❌ 否 (虚拟容器) |
| **参与布局计算** | ✅ 是 | ❌ 否 |
| **StripLayers 结果** | `VStack{appContent}` | `appContent` |
| **LayoutEngine 处理** | 嵌套 VStack，高度错误 | 扁平结构，高度正确 |
| **最终显示** | ❌ 不显示 | ✅ 正常显示 |

---

## 验证方法

### 单元测试

```bash
cd runtime/layer
go test -v -run TestStripLayers
```

**预期结果**: 所有测试通过 ✅

### 集成测试

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay

# 测试 Without Inspector
./demo_fixed.exe
# 预期: 界面正常显示

# 测试 With Inspector
TUI_INSPECTOR=true ./demo_fixed.exe
# 预期: 界面正常显示，Inspector 在右侧可见
```

---

## 相关文件清单

### 核心代码

1. **`examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go`**
   - 环境变量检查 (第 46-51 行)
   - VNode 树构建 (第 125-157 行) ← 修复位置

2. **`internal/inspector/standalone_inspector.go`**
   - Inspector 配置 (第 114-132 行)
   - RenderOverlay 方法 (第 253-291 行)

3. **`internal/render/pipeline_renderer.go`**
   - Layer 检测 (第 105-124 行)
   - 渲染路径选择 (第 69-82 行)

4. **`internal/render/rendering_pipeline.go`**
   - RenderLayers 方法 (第 145-186 行)

5. **`runtime/layer/collector.go`**
   - Collect 方法 (第 127-130 行)
   - StripLayers 方法 (第 214-228 行)

### 测试文件

1. **`runtime/layer/collector_test.go`**
   - StripLayers 单元测试

2. **`runtime/layer/inspector_integration_test.go`**
   - Inspector 集成测试

---

## 结论

**环境变量 `TUI_INSPECTOR` 的影响路径**:

1. **应用启动**: 设置 `inspector.visible = true`
2. **VNode 构建**: 决定是否在树中包含 `inspectorOverlay`
3. **Layer 检测**: `hasLayerNodes()` 返回 `true`
4. **渲染路径**: 从 `Render()` 切换到 `RenderLayers()`
5. **Layer 处理**: `StripLayers` 分离 base 和 inspector 层
6. **最终渲染**: `PaintLayers` 渲染两个独立的 layer

**关键修复**:

将 `ui.VStack(appContent, inspectorOverlay)` 改为 `ui.Fragment(appContent, inspectorOverlay)`

这确保了：
- ✅ Inspector 作为独立 layer 被正确处理
- ✅ appContent 不被嵌套在额外的容器中
- ✅ LayoutEngine 正确计算布局
- ✅ 界面和 Inspector 都正常显示
