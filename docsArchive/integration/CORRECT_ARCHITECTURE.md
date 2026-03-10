# 正确的架构分层

## 架构层次

```
┌─────────────────────────────────────────────────────┐
│  应用层 (examples/)                                  │
│  - 业务逻辑                                          │
│  - UI组件组合                                        │
│  - 不需要知道Layer、渲染细节                          │
└─────────────────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────────────────┐
│  框架层 (framework/)                                 │
│  - App生命周期管理                                   │
│  - 事件处理                                          │
│  - Inspector注册/切换                                │
│  - 不处理渲染细节                                    │
└─────────────────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────────────────┐
│  渲染层 (internal/render/, runtime/compute/)        │
│  - VNode → 布局计算                                  │
│  - 布局 → Buffer绘制                                 │
│  - **自动检测和处理Layer**                           │
│  - 应用/框架不需要知道Layer的存在                     │
└─────────────────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────────────────┐
│  运行时层 (runtime/ui/, runtime/layer/)             │
│  - VNode定义                                         │
│  - Layer枚举                                        │
│  - Layer收集/管理（供渲染层使用）                     │
└─────────────────────────────────────────────────────┘
```

## 核心原则

### 1. 关注点分离

**应用层 (examples/)**
```go
// 只关注业务逻辑
func MyApp() ui.VNode {
    return ui.VStack(
        ui.Text("Hello"),
        ui.Button("Click"),
    )
    // 完全不需要知道Layer
}
```

**框架层 (framework/)**
```go
// 只管理应用状态
app.SetInspector(inspector)  // 注册
app.Run()                     // 运行
// 不处理渲染
```

**渲染层 (internal/render/)**
```go
// 自动处理多层渲染
func (r *PipelineRenderer) Render(vnode, buffer) {
    // 自动检测：
    // - 是否有Modal？
    // - 是否有Inspector？
    // - 是否有其他overlay？

    // 自动选择：
    // - 单层渲染
    // - 多层渲染
}
```

### 2. Inspector应该如何集成？

**当前错误的方式：**
```go
// ❌ 应用代码处理Layer
if inspector.IsVisible() {
    overlay := inspector.RenderOverlay()
    overlay.SetLayer(rtui.LayerInspector)  // 侵入性！
    return ui.Fragment(content, overlay)    // 应用层处理！
}
```

**正确的方式：**
```go
// ✅ 应用只返回业务内容
func MyApp() ui.VNode {
    return ui.VStack(
        Header(),
        Content(),
    )
}

// ✅ 框架只管理状态
app.SetInspector(inspector)
app.Run()

// ✅ 渲染层自动检测并处理
// PipelineRenderer自动发现：
// - vnode树中有Inspector组件
// - Inspector在app中已注册且可见
// - 自动将Inspector作为overlay渲染
```

## 如何实现自动检测？

### 方案：Inspector作为特殊组件标记

```go
// runtime/ui/component.go
type ComponentVNode struct {
    // ...
    isOverlay bool  // 新增：标记为overlay组件
}

// internal/inspector/inspector.go
func NewInspector() *ComponentVNode {
    comp := &ComponentVNode{Name: "Inspector"}
    comp.isOverlay = true  // 标记为overlay
    return comp
}

// internal/render/pipeline_renderer.go
func (r *PipelineRenderer) detectOverlays(vnode) []VNode {
    var overlays []VNode

    // 遍历VNode树，查找isOverlay=true的组件
    WalkVNode(vnode, func(node VNode) {
        if comp, ok := node.(*ComponentVNode); ok && comp.isOverlay {
            overlays = append(overlays, comp)
        }
    })

    return overlays
}

func (r *PipelineRenderer) Render(vnode, buffer) {
    overlays := r.detectOverlays(vnode)

    if len(overlays) > 0 {
        // 自动使用多层渲染
        r.renderWithLayers(vnode, overlays, buffer)
    } else {
        // 单层渲染
        r.renderSingle(vnode, buffer)
    }
}
```

### 或者：通过组件类型识别

```go
// internal/render/pipeline_renderer.go
func (r *PipelineRenderer) isInspectorComponent(vnode VNode) bool {
    if comp, ok := vnode.(*ComponentVNode); ok {
        return comp.Name() == "Inspector"
    }
    return false
}

func (r *PipelineRenderer) autoWrapInspector(base VNode, inspector VNode) VNode {
    // 自动为Inspector设置Layer属性
    inspector.SetLayer(rtui.LayerInspector)

    // 自动包装成Fragment
    return rtui.Fragment(base, inspector)
}
```

## 总结

1. **应用层** - 只写业务逻辑，不关心Layer
2. **框架层** - 只管理应用状态，不处理渲染
3. **渲染层** - 自动检测overlay组件，自动使用多层渲染
4. **运行时层** - 提供Layer API，由渲染层调用

这样各层职责清晰，没有侵入性调用。
