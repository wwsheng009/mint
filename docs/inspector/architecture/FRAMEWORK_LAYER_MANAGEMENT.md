# 框架Layer管理架构 - 正确方案

## 当前架构理解

### 渲染流程
```
Application
  → component.Paintable.Paint()
    → internal.render.DeclarativeNode
      → internal.render.PipelineRenderer
        → PipelineRenderer.Render()
          ├─ hasLayerNodes() - 检测是否有非Base Layer
          ├─ if hasLayers:
          │    → pipeline.RenderLayers()  // 多层渲染
          └─ else:
               → pipeline.Render()        // 单层渲染
```

### 关键发现
1. **PipelineRenderer已经自动处理多层渲染**
   - 通过`hasLayerNodes()`检测VNode树中的Layer属性
   - 自动选择`RenderLayers()`或`Render()`

2. **问题是Inspector需要手动SetLayer()**
   - 当前：Inspector或应用代码调用`SetLayer(LayerInspector)`
   - 这是侵入性的

## 正确的解决方案

### 方案A: 框架自动注入Inspector（推荐）

**在DeclarativeNode.Paint()中自动处理**

```go
// internal/render/declarative_node.go
func (n *DeclarativeNode) Paint(ctx PaintContext, buf *paint.Buffer) {
    // 1. 获取应用根节点
    rootNode := n.buildVNode() // 调用组件的Render()

    // 2. 检查框架是否有Inspector
    if fwApp := n.GetFrameworkApp(); fwApp != nil {
        if inspector := fwApp.GetInspector(); inspector != nil && inspector.IsVisible() {
            // 3. 框架自动包装Inspector，不需要应用知道
            inspectorContent := inspector.RenderContent()
            inspectorContent.SetLayer(rtui.LayerInspector)
            inspectorContent.SetProps(rtui.Props{
                "x": inspector.GetPosition(),
                "y": inspector.GetPosition(),
            })

            // 4. 使用Fragment组合
            rootNode = rtui.Fragment(rootNode, inspectorContent)
        }
    }

    // 5. 正常渲染（PipelineRenderer会自动检测Layer）
    if adapter, ok := n.renderer.(*PipelineRendererAdapter); ok {
        adapter.GetPipeline().Render(rootNode, 0, 0, buf)
    }
}
```

**优点：**
- ✅ 应用代码完全不需要知道Layer
- ✅ Inspector代码也不需要调用SetLayer()
- ✅ 框架统一管理，职责清晰
- ✅ 没有import循环（DeclarativeNode已经知道framework）

**应用代码：**
```go
// main.go - 完全不需要知道Layer
func RuntimeDemo() ui.VNode {
    return ui.VStack(
        HeaderPanel(),
        ContentPanel(),
        ControlPanel(),
    )
}
```

**Inspector代码：**
```go
// inspector.go - 也不需要知道Layer
func (si *Inspector) RenderContent() ui.VNode {
    return ui.Bordered().
        Label("INSPECTOR").
        Child(si.buildTabs()).
        Build()
    // 不调用SetLayer()！
}
```

### 方案B: 应用层手动组装（当前方案，不推荐）

```go
// main.go - 应用需要知道Layer细节
func RuntimeDemoWithInspector() ui.VNode {
    appContent := buildAppContent()

    if inspector.IsVisible() {
        inspectorOverlay := inspector.RenderOverlay()
        inspectorOverlay.SetLayer(rtui.LayerInspector)  // ❌ 侵入性
        inspectorOverlay.SetProps(ui.Props{"x": 40, "y": 5})  // ❌ 侵入性
        return ui.Fragment(appContent, inspectorOverlay)  // ❌ 应用层处理
    }

    return appContent
}
```

**缺点：**
- ❌ 应用需要理解Layer、Fragment、SetLayer
- ❌ Inspector需要调用SetLayer()
- ❌ 容易出错（SetProps覆盖问题）
- ❌ 职责不清

## 实施计划

### Phase 1: 在DeclarativeNode中添加Inspector自动处理

1. 修改`internal/render/declarative_node.go`
2. 添加`GetFrameworkApp()`方法
3. 在`Paint()`中自动检测并注入Inspector

### Phase 2: 简化Inspector接口

1. 添加`RenderContent() ui.VNode`方法
2. 保留`RenderOverlay()`用于向后兼容
3. 移除Inspector中的SetLayer()调用

### Phase 3: 更新示例

1. 简化demo2，移除Fragment包装
2. 移除SetLayer()调用
3. 验证功能正常

### Phase 4: 添加测试

1. 测试框架自动处理Inspector
2. 测试Layer正确设置
3. 测试渲染顺序正确

## 关键点

1. **框架的职责边界：**
   - framework/app.go: 管理Inspector生命周期（注册、切换）
   - internal/render/declarative_node.go: 自动注入Inspector到VNode树
   - internal/render/pipeline_renderer.go: 检测Layer并选择渲染路径

2. **应用不需要知道：**
   - Layer的存在
   - Fragment包装
   - SetLayer()调用

3. **Inspector只需要：**
   - 提供RenderContent()方法
   - 返回纯UI内容，不设置Layer

这样既符合当前的架构，又避免了侵入性调用。
