# 使用Hook集成Inspector - 设计方案

## 核心思想

通过Hook系统实现Inspector的自动集成，避免：
- ❌ 应用层手动调用SetLayer()
- ❌ 框架层直接处理渲染细节
- ❌ 侵入性修改VNode树

## Hook系统设计

### 1. Hook定义

```go
// runtime/render/hook.go
package render

// VNodeHook 在VNode树构建后修改VNode
type VNodeHook func(VNode) VNode

// RenderHook 在渲染前修改VNode或buffer
type RenderHook func(VNode, *paint.Buffer) error

// HookManager 管理所有hook
type HookManager struct {
    vnodeHooks  []VNodeHook
    renderHooks []RenderHook
}

// RegisterVNodeHook 注册VNode处理hook
func (hm *HookManager) RegisterVNodeHook(hook VNodeHook) {
    hm.vnodeHooks = append(hm.vnodeHooks, hook)
}

// RegisterRenderHook 注册渲染hook
func (hm *HookManager) RegisterRenderHook(hook RenderHook) {
    hm.renderHooks = append(hm.renderHooks, hook)
}

// ApplyVNodeHooks 应用所有VNode hook
func (hm *HookManager) ApplyVNodeHooks(vnode VNode) VNode {
    for _, hook := range hm.vnodeHooks {
        vnode = hook(vnode)
    }
    return vnode
}
```

### 2. 渲染层集成Hook

```go
// internal/render/pipeline_renderer.go
type PipelineRenderer struct {
    pipeline *RenderingPipeline
    hooks    *render.HookManager
}

func (r *PipelineRenderer) Render(vnode VNode, x, y int, buf *paint.Buffer) error {
    // 1. 应用VNode hooks（自动注入Inspector等）
    vnode = r.hooks.ApplyVNodeHooks(vnode)

    // 2. 正常渲染流程
    return r.pipeline.Render(vnode, x, y, buf)
}
```

### 3. Inspector Hook实现

```go
// internal/inspector/hook.go
package inspector

import (
    "github.com/wwsheng009/mint/runtime/render"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// CreateInspectorHook 创建Inspector集成hook
func CreateInspectorHook(inspector *StandaloneInspector) render.VNodeHook {
    return func(vnode rtui.VNode) rtui.VNode {
        // 如果Inspector不可见，直接返回原VNode
        if !inspector.IsVisible() {
            return vnode
        }

        // 获取Inspector内容
        overlay := inspector.RenderContent()

        // 设置Layer属性（在这里设置，不污染应用代码）
        overlay.SetLayer(rtui.LayerInspector)
        overlay.SetProps(rtui.Props{
            "x": inspector.GetPosition(),
            "y": inspector.GetPosition(),
        })

        // 包装成Fragment
        return rtui.Fragment(vnode, overlay)
    }
}

// RegisterInspector 注册Inspector到渲染系统
func RegisterInspector(
    inspector *StandaloneInspector,
    hooks *render.HookManager,
) {
    hook := CreateInspectorHook(inspector)
    hooks.RegisterVNodeHook(hook)
}
```

### 4. 框架层注册Hook

```go
// framework/app.go
type App struct {
    inspector        interface{}
    hooks            *render.HookManager  // 新增
}

func (a *App) SetInspector(inspector interface{}) {
    a.inspector = inspector

    // 自动注册Inspector hook
    if insp, ok := inspector.(*inspector.StandaloneInspector); ok {
        inspector.RegisterInspector(insp, a.hooks)
    }
}
```

### 5. 应用层完全透明

```go
// examples/demo/main.go
func main() {
    app := framework.NewApp()

    // 注册Inspector
    inspector := inspector.NewStandaloneInspector()
    app.SetInspector(inspector)

    // 设置根组件
    app.SetRoot(myApp)

    // 运行
    app.Run()
}

// 应用组件完全不知道Inspector的存在
func myApp() ui.VNode {
    return ui.VStack(
        ui.Text("Hello"),
        ui.Button("Click"),
    )
}
```

## 优势

### 1. 解耦
- Inspector不需要知道framework的存在
- Framework不需要知道Layer的细节
- 应用代码完全不知道Inspector

### 2. 可测试
```go
// 测试Hook
func TestInspectorHook(t *testing.T) {
    inspector := NewMockInspector()
    hook := CreateInspectorHook(inspector)

    baseVNode := ui.Text("Base")
    result := hook(baseVNode)

    // 验证Fragment包装
    assert.IsType(*rtui.FragmentVNode{}, result)
}
```

### 3. 可扩展
```go
// 未来可以添加其他overlay hook
tooltipHook := CreateTooltipHook()
modalHook := CreateModalHook()

hooks.RegisterVNodeHook(tooltipHook)
hooks.RegisterVNodeHook(modalHook)
```

### 4. 非侵入
- Hook可以按需注册/注销
- 不修改核心渲染代码
- 符合开闭原则

## 实施步骤

### Phase 1: 创建Hook系统
1. 创建 `runtime/render/hook.go`
2. 定义Hook接口和HookManager
3. 在PipelineRenderer中集成HookManager

### Phase 2: 实现Inspector Hook
1. 创建 `internal/inspector/hook.go`
2. 实现CreateInspectorHook()
3. 添加RegisterInspector()函数

### Phase 3: 框架集成
1. App添加hooks字段
2. SetInspector()自动注册hook
3. 移除useLayers配置

### Phase 4: 简化示例
1. 更新demo2，移除手动Fragment包装
2. 验证Inspector正常显示
3. 添加测试

### Phase 5: 清理
1. 移除TUI_USE_LAYERS环境变量
2. 移除应用层的SetLayer()调用
3. 更新文档

## 总结

Hook方式是正确的架构选择：
- ✅ 关注点分离
- ✅ 非侵入性集成
- ✅ 易于测试
- ✅ 可扩展性强
