# Inspector 三个核心问题的详细回答

## 问题 1: Inspector 为什么需要显式调用 setLayer？

### 当前实现

```go
// Inspector 内部代码
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
    content := si.buildOverlayContent()

    // ❌ Inspector 必须知道自己是 LayerInspector
    content.SetLayer(rtui.LayerInspector)
    content.SetProps(ui.Props{"x": 40, "y": 5})

    return content
}
```

### 为什么需要这样设计？

**历史原因**:
1. Layer 系统是后来添加的，不是原始设计的一部分
2. 为了向后兼容，使用"标记"模式而不是"包装"模式
3. 避免大规模重构现有组件

**技术原因**:
```go
// 渲染引擎通过检查 VNode 的 layer 属性来决定渲染路径
func (r *PipelineRenderer) hasLayerNodes(vnode rtui.VNode) bool {
    layer := vnode.GetLayer()  // ← 读取 layer 属性
    if layer != rtui.LayerBase && layer.IsValid() {
        return true  // 发现非 base layer
    }
    // ...
}
```

如果 Inspector 不调用 `SetLayer()`，渲染引擎无法知道它属于哪个层。

### 这样设计的缺点

1. **耦合度高**: Inspector 组件依赖于 Layer 系统的实现细节
2. **违反单一职责**: Inspector 既负责内容，又负责自己的层级
3. **不够灵活**: 无法在不修改 Inspector 代码的情况下改变它的层级

### 更好的设计（未来方向）

```go
// 理想设计: 组件只负责内容
func (si *StandaloneInspector) RenderContent() ui.VNode {
    return si.buildOverlayContent()  // 纯内容，无 layer 信息
}

// 应用层决定如何使用组件
func RuntimeDemo() ui.VNode {
    inspectorContent := globalInspector.RenderContent()

    // 使用 Layer 包装器
    return ui.Fragment(
        appContent,
        ui.InspectorLayer(  // 包装器负责设置 layer
            ui.Position(40, 5, inspectorContent),
        ),
    )
}
```

---

## 问题 2: 渲染引擎如何处理多个层？为什么 Inspector 位置是 (40, 5)？

### 2.1 渲染引擎的多层处理机制

```go
// paint_engine.go
func (e *PaintEngine) PaintLayers(layouts LayerLayouts, buffer *paint.Buffer) error {
    // 定义渲染顺序（Z-order）
    renderOrder := []rtui.Layer{
        rtui.LayerBase,      // 0: 最底层
        rtui.LayerOverlay,   // 1
        rtui.LayerModal,     // 2
        rtui.LayerTooltip,   // 3
        rtui.LayerInspector, // 4: 最顶层
    }

    // 按 Z-order 从低到高渲染
    for _, layer := range renderOrder {
        layout := layouts[layer]
        if layout == nil || layout.Root == nil {
            continue
        }

        // 每个 layer 依次绘制到同一个 buffer
        e.Paint(layout, buffer)
        //    ^^^^^
        //    直接覆盖，后绘制的覆盖先绘制的
    }

    return nil
}
```

### 覆盖机制详解

```
Buffer 视图（简化）:

初始状态:
┌────────────────────────────────────────┐
│                                        │
│           全空                          │
│                                        │
└────────────────────────────────────────┘

步骤 1: Paint(LayerBase) at (0, 0)
┌────────────────────────────────────────┐
│  主应用内容                             │
│  [Header]                              │
│  [Pipeline]                            │
│  [Stats]                               │
└────────────────────────────────────────┘
  0───────────────40                    119

步骤 2: Paint(LayerInspector) at (40, 5)
┌────────────────────────────────────────┐
│  主应用        │Inspector               │
│  [Header]      │+──────────+           │
│  [Pipeline]    ││Inspector │           │
│  [Stats]       ││Content   │           │
│                │+──────────+           │
└────────────────────────────────────────┘
  0──────────────40                    119
                ↑
             Inspector 从这里开始
```

**关键点**:
1. 后绘制的层覆盖先绘制的层
2. Inspector 在 (40, 5)，所以主应用的 (40-119, 5-29) 区域被覆盖
3. 主应用的 (0-39, 0-39) 区域仍然可见

### 2.2 为什么 Inspector 位置是 (40, 5) 而不是 (0, 0)？

#### 当前代码

```go
// internal/inspector/standalone_inspector.go:115-132
func NewStandaloneInspector() *StandaloneInspector() {
    defaultX := 40  // ← 为什么是 40？
    defaultY := 5   // ← 为什么是 5？

    return &StandaloneInspector{
        overlayWidth:  80,  // Inspector 宽度
        overlayHeight: 25,  // Inspector 高度
        floatX:        defaultX,
        floatY:        defaultY,
    }
}
```

#### 位置计算逻辑

```
屏幕尺寸: 120 x 40
Inspector 尺寸: 80 x 25

X 位置计算:
┌────────────────────────────────────────────┐
│  主应用区域        Inspector 区域            │
│  ←───────→        ←──────────→             │
│  0      39        40        119            │
│  ↑                ↑                         │
│  主应用宽度 40    Inspector 宽度 80          │
└────────────────────────────────────────────┘

计算公式:
X = 屏幕宽度 - Inspector 宽度
  = 120 - 80
  = 40

为什么不用 (0, 0)?
- (0, 0) 位置应该留给主应用
- Inspector 是"覆盖层"，应该浮动在主应用旁边或上方
- 如果 Inspector 在 (0, 0)，会完全遮挡主应用 ❌

Y 位置计算:
Y = 5 (顶部留白)
- 避免 Inspector 遮挡主应用的标题栏
- 给用户一些视觉空间
```

### Inspector 如何覆盖显示？

**机制**: 直接像素覆盖

```
// 伪代码
for y := 0; y < inspectorHeight; y++ {
    for x := 0; x < inspectorWidth; x++ {
        targetX := inspectorX + x  // 40 + x
        targetY := inspectorY + y  // 5 + y

        // 直接覆盖 buffer 中的像素
        buffer[targetY][targetX] = inspectorContent[y][x]
        //  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
        //  主应用的内容被 Inspector 替换
    }
}
```

**没有透明度**:
- 当前实现没有 alpha blending
- Inspector 完全遮挡主应用
- 未来可以添加半透明背景

---

## 问题 3: 如何在一个 root 节点下渲染多个节点？

### 当前限制

```go
// ui.Run() 只接受一个 VNode
func RuntimeDemo() ui.VNode {
    // ❌ 只能返回一个节点
    return appContent
}

// ❌ 如果要渲染多个节点，必须包装
func RuntimeDemoWithInspector() ui.VNode {
    return ui.Fragment(  // ← Fragment 是一个 workaround
        appContent,
        inspectorOverlay,
    )
}
```

### Fragment 的工作原理

```go
// FragmentVNode 实现
type FragmentVNode struct {
    children []ui.VNode
}

func (f *FragmentVNode) Children() []ui.VNode {
    return f.children
}

// 渲染时如何处理 Fragment?
func (r *PipelineRenderer) hasLayerNodes(vnode rtui.VNode) bool {
    // Fragment 会递归检查所有子节点
    for _, child := range vnode.Children() {
        if r.hasLayerNodes(child) {
            return true  // 如果任一子节点有 layer，返回 true
        }
    }
    return false
}
```

### 理想的多 Root 渲染架构

#### 方案 A: 直接返回数组

```go
// 改进 API: 支持返回节点数组
func RuntimeDemo() []ui.VNode {
    var nodes []ui.VNode
    nodes = append(nodes, appContent)

    if globalInspector.IsVisible() {
        nodes = append(nodes, globalInspector.RenderContent())
    }

    return nodes  // ← 返回数组而不是单个节点
}

// 渲染引擎处理数组
func (r *PipelineRenderer) Render(vnode interface{}, ...) error {
    if array, ok := vnode.([]ui.VNode); ok {
        // 渲染多个节点
        return r.renderMultiple(array, ...)
    }
    return r.renderSingle(vnode, ...)
}
```

#### 方案 B: Scene Graph

```go
// 引入 Scene 概念
type Scene struct {
    nodes []*SceneNode
}

type SceneNode struct {
    Content   ui.VNode
    Layer     rtui.Layer
    Position  struct{ X, Y int }
    ZIndex    int
}

func (s *Scene) Add(content ui.VNode, layer rtui.Layer, x, y int) {
    s.nodes = append(s.nodes, &SceneNode{
        Content:  content,
        Layer:    layer,
        Position: struct{ X, Y int }{X: x, Y: y},
        ZIndex:   layer.ZIndex(),
    })
}

// 使用
func RuntimeDemo() *Scene {
    scene := &Scene{}

    // 添加主应用
    scene.Add(appContent, rtui.LayerBase, 0, 0)

    // 条件性添加 Inspector
    if globalInspector.IsVisible() {
        scene.Add(
            globalInspector.RenderContent(),
            rtui.LayerInspector,
            40, 5,
        )
    }

    return scene
}

// 渲染引擎处理 Scene
func (r *PipelineRenderer) RenderScene(scene *Scene, ...) error {
    // 按 Z-index 排序
    sort.Slice(scene.nodes, func(i, j int) bool {
        return scene.nodes[i].ZIndex < scene.nodes[j].ZIndex
    })

    // 渲染每个节点
    for _, node := range scene.nodes {
        r.renderNodeAt(node.Content, node.Position.X, node.Position.Y, ...)
    }

    return nil
}
```

### 当前推荐的实践

```go
// ✅ 使用 Fragment（当前最佳实践）
func RuntimeDemoWithInspectorOverlay() ui.VNode {
    inspectorVisible := globalInspector.IsVisible()

    if inspectorVisible {
        return ui.Fragment(          // ← 使用 Fragment
            appContent,              // 主应用
            globalInspector.RenderOverlay(),  // Inspector
        )
    }

    return appContent
}
```

**为什么 Fragment 当前是最好的选择**:
1. ✅ 避免创建额外的布局容器（不像 VStack）
2. ✅ 渲染引擎自动处理 Fragment 的子节点
3. ✅ 不需要修改核心 API
4. ✅ 向后兼容

---

## 总结

### 问题 1: Inspector 为什么需要显式调用 setLayer？

**答**: 因为渲染引擎通过读取 VNode 的 layer 属性来决定渲染路径。这是当前架构的必要实现，但不是最优设计。

**未来改进**: 引入 Layer 包装器，让应用层决定组件的层级，而不是组件内部。

### 问题 2: 渲染引擎如何处理多个层？

**答**: 按 Z-order 从低到高依次渲染到同一个 buffer，后渲染的覆盖先渲染的。Inspector 在 (40, 5) 是为了避免遮挡主应用，它只占据屏幕的右侧部分。

**覆盖机制**: 直接像素替换（没有透明度），Inspector 区域完全覆盖主应用。

### 问题 3: 如何在一个 root 节点下渲染多个节点？

**答**: 当前使用 Fragment 作为 workaround。理想方案是支持返回节点数组或引入 Scene Graph，允许动态管理多个节点。

**当前推荐**: 使用 Fragment，它不会创建额外的布局容器，渲染引擎会自动处理。
