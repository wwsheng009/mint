# Layer系统架构重构方案

## 当前问题

### 1. 侵入性调用
应用代码直接调用底层Layer API：
```go
// main.go - 侵入性调用
inspectorOverlay.SetLayer(rtui.LayerInspector)  // ❌ 应用不应该知道Layer
inspectorOverlay.SetProps(ui.Props{"x": 40, "y": 5})
root := ui.Fragment(appContent, inspectorOverlay)  // ❌ 应用手动组装
```

### 2. 职责混乱
- **框架**: 应该负责管理多层渲染
- **应用**: 应该只关注业务逻辑
- **当前**: 应用需要理解Layer、Fragment、SetLayer等底层概念

### 3. Inspector集成混乱
```go
// standalone_inspector.go
content.SetProps(ui.Props{"x": x, "y": y})  // SetProps会覆盖！
content.SetLayer(rtui.LayerInspector)       // SetLayer会丢失！
```

## 正确的架构

### 框架层职责
```
framework/App
  ├─ SetInspector(inspector)           // 注册Inspector
  ├─ SetupInspectorShortcut()          // 配置快捷键
  └─ render()                          // 自动处理多层渲染
       ├─ if inspector.Visible():
       │    ├─ 渲染应用内容 (LayerBase)
       │    └─ 渲染Inspector覆盖层 (LayerInspector)
       └─ Paint到buffer
```

### 应用层职责
```go
// 应用只需要：
app.SetInspector(myInspector)  // 注册
app.SetupInspectorShortcut()   // 启用快捷键
app.Run()                       // 运行

// 不需要：
// ❌ 调用SetLayer()
// ❌ 手动Fragment包装
// ❌ 处理x/y坐标
```

### Inspector职责
```go
type Inspector interface {
    // 显示/隐藏
    ToggleVisibility()
    IsVisible() bool

    // 渲染内容
    RenderContent() ui.VNode  // 只返回内容，不设置Layer

    // 位置调整（可选）
    Move(dx, dy int)
    GetPosition() (x, y int)
}
```

## 重构方案

### Phase 1: 框架自动管理Inspector Layer

**framework/app.go**
```go
func (a *App) render() {
    // 1. 渲染基础内容
    baseContent := a.renderBaseContent()

    // 2. 检查Inspector是否可见
    if a.isInspectorVisible() {
        inspectorContent := a.getInspectorContent()
        // 框架自动使用多层渲染
        a.renderWithLayers(baseContent, inspectorContent)
    } else {
        a.renderSingle(baseContent)
    }
}
```

**framework/layer_renderer.go** (新文件)
```go
// LayerRenderer 统一管理多层渲染
type LayerRenderer struct {
    engine *compute.Engine
    paintEngine *render.PaintEngine
}

// RenderLayers 渲染多个层
func (lr *LayerRenderer) RenderLayers(
    base VNode,
    overlay VNode,
    buffer *paint.Buffer,
) error {
    // 自动为overlay设置LayerInspector
    // 自动处理x/y坐标
    // 应用不需要知道任何Layer细节
}
```

### Phase 2: 清理Inspector接口

**internal/inspector/standalone_inspector.go**
```go
// RenderContent 返回Inspector内容（不设置Layer）
func (si *StandaloneInspector) RenderContent() rtui.VNode {
    // 只负责构建UI，不设置Layer、不设置x/y
    return si.buildOverlayContent()
}

// 移除：
// - RenderOverlay()  // 混合了UI构建和Layer设置
// - SetLayer() 调用  // 框架负责
// - SetProps(x, y)   // 框架负责
```

### Phase 3: 添加测试

**framework/layer_integration_test.go**
```go
func TestFrameworkManagesLayers(t *testing.T) {
    app := NewApp()
    inspector := NewMockInspector()

    app.SetInspector(inspector)
    app.SetRoot(mockContent)

    // 框架自动处理多层渲染
    app.render()

    // 验证：
    // 1. Inspector内容被渲染
    // 2. Inspector在正确的位置
    // 3. 基础内容不被影响
}
```

## 实施步骤

1. **停止修复打补丁** - 撤销所有SetProps/SetLayer相关的修改
2. **创建框架层渲染管理** - framework/layer_manager.go
3. **修改App.render()** - 自动检测Inspector并使用多层渲染
4. **简化Inspector接口** - 移除Layer相关的代码
5. **添加集成测试** - 验证框架正确处理多层渲染
6. **更新示例** - demo2不再需要手动处理Layer

## 成功标准

1. ✅ 应用代码不需要调用SetLayer()
2. ✅ 应用代码不需要手动Fragment包装
3. ✅ Inspector不需要知道Layer的存在
4. ✅ 框架自动管理所有层的渲染
5. ✅ 测试覆盖多层渲染场景
