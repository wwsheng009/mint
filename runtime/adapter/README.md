# Adapter

组件适配器接口定义层，用于桥接不同组件系统或层次之间的接口差异。

> **当前状态**：此包为占位包，目前未包含实现代码。项目中各模块使用内部适配器模式，而非统一的适配器接口。

## 职责

- **适配器接口定义**：定义通用的组件适配接口（计划中）
- **跨层桥接**：为 runtime 与 framework 之间的交互提供适配层（计划中）
- **接口转换**：将一种组件接口转换为另一种（计划中）

## 核心概念

### 适配器模式

适配器模式（Adapter Pattern）用于将一个类的接口转换成客户期望的另一个接口，使原本接口不兼容的类可以一起工作。在 Mint 框架中，适配器用于：

1. **模块间集成**：如 Selection 模块通过 `RuntimeAdapter` 与 Runtime 集成
2. **DevTools 集成**：通过 `LayoutNodeAdapter` 桥接 Runtime 布局树与 DevTools
3. **框架层适配**：`componentNodeAdapter` 将框架组件适配为运行时布局节点
4. **状态绑定**：`StoreToComponentAdapter` 将状态流转为组件可用的格式

### 现有适配器实现

虽然 `runtime/adapter` 包本身是占位符，但项目中有多个具体的适配器实现：

#### 1. RuntimeAdapter (runtime/selection/adapter.go)

将文本选择功能集成到 Runtime 系统，作为 Runtime 和 Selection 模块之间的桥梁。

```go
type RuntimeAdapter struct {
    textSelection *TextSelection
    enabled       bool
}
```

**核心功能**：
- `OnRender(frame)` - 在每次渲染后应用选择高亮
- `OnEvent(ev)` - 处理选择相关的事件
- `Copy()`, `SelectAll()`, `ClearSelection()` - 选择操作
- `SetHighlightStyle()`, `SetSelectionMode()` - 样式和模式配置

#### 2. CellStyleAdapter (runtime/selection/render.go)

为选择高亮提供样式适配功能。

#### 3. LayoutNodeAdapter (devtools/runtime_adapter.go)

将 Runtime 的布局树节点适配为 DevTools 可用的格式。

```go
type LayoutNodeAdapter struct {
    ID          NodeID
    Type        string
    Position    geometry.Rect
    Children    []*LayoutNodeAdapter
    Properties  map[string]interface{}
}
```

#### 4. componentNodeAdapter (framework/layout/flex.go)

将框架层的 Component 适配为运行时布局系统所需的节点接口。

```go
type componentNodeAdapter struct {
    component.Component
}
```

#### 5. StoreToComponentAdapter (framework/component/binding/integration.go)

将状态 Store 的变化适配为组件可用的格式。

### 适配器类型

按照适配的范围和用途，可以分为：

1. **集成适配器**：将独立模块集成到主系统（如 `RuntimeAdapter`）
2. **格式适配器**：转换数据格式以适应不同系统（如 `LayoutNodeAdapter`）
3. **接口适配器**：包装实现以满足特定接口要求（如 `componentNodeAdapter`）
4. **流适配器**：转换事件流以适应消费者需求（如 `StoreToComponentAdapter`）

## 使用示例

### RuntimeAdapter 使用示例

```go
import (
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/selection"
)

func main() {
    // 创建 Runtime
    rt := runtime.NewRuntime()
    
    // 创建 RuntimeAdapter 集成选择功能
    selAdapter := selection.NewRuntimeAdapter()
    
    // 在 Runtime.Render() 后调用应用选择高亮
    rt.OnRender(func() {
        frame := rt.LastFrame()
        selAdapter.OnRender(frame)
    })
    
    // 在事件分发前处理选择事件
    rt.OnEvent(func(ev interface{}) bool {
        return selAdapter.OnEvent(ev)
    })
}
```

### 创建自定义适配器

当需要将一个外部组件系统集成到 Mint 时，可以创建适配器：

```go
// 假设有一个第三方文本编辑器组件
type ThirdPartyEditor struct {
    // ...
}

type EditorAdapter struct {
    editor *ThirdPartyEditor
}

// 适配为 runtime.Paintable 接口
func (a *EditorAdapter) Paint(ctx *runtime.PaintContext) error {
    // 将第三方编辑器的内容绘制到 Buffer
    content := a.editor.GetContent()
    ctx.Write(content)
    return nil
}

// 适配为 component.Actionable 接口
func (a *EditorAdapter) HandleAction(act runtime.Action) bool {
    switch act.Type() {
    case "insert":
        a.editor.Insert(act.Value)
    case "delete":
        a.editor.Delete()
    default:
        return false
    }
    return true
}
```

## 核心类型

**当前状态**：`runtime/adapter` 包未定义核心类型。以下是从其他模块抽取的适配器接口建议：

### 建议的接口定义（规划中）

```go
// ComponentAdapter 将外部组件适配为 Mint 可用的节点
type ComponentAdapter interface {
    // 返回组件的唯一标识
    ID() string
    
    // 返回组件类型
    Type() string
    
    // 测量组件大小
    Measure(constraints geometry.Constraints) geometry.Size
    
    // 绘制组件到缓冲区
    Paint(ctx *runtime.PaintContext) error
    
    // 处理事件
    HandleAction(act runtime.Action) bool
}

// StateAdapter 将外部状态源适配为 Mint 可用的 Store 接口
type StateAdapter interface {
    // 订阅状态变化
    Subscribe(key string, handler func(value interface{}))
    
    // 获取状态值
    Get(key string) interface{}
    
    // 设置状态值
    Set(key string, value interface{})
}
```

## 文件结构

```
runtime/adapter/
└── README.md       # 本文档
                   # adapter.go 未实现（计划中）
```

### 依赖此包的其他模块

*无* - 当前 `runtime/adapter` 包为空。

### 实际适配器实现位置

```
runtime/selection/
└── adapter.go              # RuntimeAdapter（选择与 Runtime 集成）

runtime/selection/
└── render.go               # CellStyleAdapter（样式适配）

devtools/
└── runtime_adapter.go      # LayoutNodeAdapter（DevTools 集成）

framework/layout/
└── flex.go                 # componentNodeAdapter（组件到布局节点）

framework/component/binding/
└── integration.go          # StoreToComponentAdapter（状态绑定）
```

## 依赖

### 外部依赖
- 无（当前为空包）

### 内部依赖
- 无依赖（当前为空包）

### 依赖此包的模块
- 无（当前为空包）

## 与其他模块集成

虽然没有统一的 `runtime/adapter` 包，但各模块使用内部适配器实现集成：

### 与 Selection 模块集成

`RuntimeAdapter` 作为 Selection 与 Runtime 的桥梁：

```go
// Runtime 通过适配器启用选择功能
selAdapter := selection.NewRuntimeAdapter()
selAdapter.SetEnabled(true)

// 渲染后应用高亮
selAdapter.OnRender(runtime.Frame)

// 复制选中文本
selAdapter.Copy()
```

### 与 DevTools 集成

`LayoutNodeAdapter` 将运行时布局树转换为 DevTools 可用的格式：

```go
// DevTools 通过适配器获取布局信息
nodeAdapter := devtools.NewLayoutNodeAdapter(runtime.Layout())
layoutResult := nodeAdapter.ToLayoutResult()

// DevTools 使用适配后的数据进行性能分析
collector.AnalyzeLayout(layoutResult)
```

### 与 Framework 集成

`componentNodeAdapter` 将框架层组件适配为运行时可用的布局节点：

```go
// Flex 布局使用包装器适配组件
func adaptComponent(c component.Component) layout.Node {
    return &componentNodeAdapter{Component: c}
}
```

### 与 State 绑定集成

`StoreToComponentAdapter` 将状态流转为组件可用格式：

```go
// 绑定层使用适配器连接 Store 和 Component
adapter := binding.NewStoreToComponentAdapter(store, component)
store.Subscribe(key, func(newVal interface{}) {
    adapter.UpdateComponent(newVal)
})
```

## 最佳实践

### 1. 适配器职责单一

每个适配器应专注于一个特定的桥接任务：

```go
// 好的设计：一个适配器只负责一个桥接任务
type RuntimeAdapter struct{}  // Runtime 与 Selection 之间的桥接
type LayoutNodeAdapter struct{} // Runtime 布局与 DevTools 之间的桥接

// 避免设计：一个适配器同时处理多个无关的桥接任务
type MultiAdapter struct {}  // 不推荐
```

### 2. 保持轻量

适配器应只做必要的转换，不添加业务逻辑：

```go
// 好的设计：轻量级转换
func (a *EditorAdapter) Measure() geometry.Size {
    return geometry.Size{Width: a.editor.Width, Height: a.editor.Height}
}

// 避免设计：在适配器中添加复杂逻辑
func (a *EditorAdapter) Measure() geometry.Size {
    // 不推荐：在适配器中进行复杂计算
    result := make(chan geometry.Size)
    go func() {
        // 大量计算...
        result <- calculateSize()
    }()
    return <-result
}
```

### 3. 使用零开销包装

对于高性能场景，优化适配器的内存占用：

```go
// 好的设计：使用包含类型，无额外字段
type componentNodeAdapter struct {
    component.Component  // 直接嵌入，零额外开销
}

// 或使用接口，避免复制
type PaintableAdapter struct {
    interface{ Paint() }  // 仅嵌入必要的接口方法
}
```

### 4. 提供清晰的生命周期管理

适配器应明确初始化和清理时机：

```go
// 好的设计：清晰的初始化和清理
func NewAdapter() *Adapter {
    a := &Adapter{}
    a.init()  // 初始化
    return a
}

func (a *Adapter) Close() error {
    a.cleanup()  // 清理资源
    return nil
}
```

### 5. 文档化适配的边界条件

明确适配器的能力和限制：

```go
// RuntimeAdapter 将文本选择功能集成到 Runtime 系统。
//
// 限制：
//   - 仅支持 CellBuffer 类型的 Frame
//   - 不支持嵌套选择区域
//   - 剪贴板支持依赖于操作系统
type RuntimeAdapter struct { ... }
```

## 常见问题

### Q: runtime/adapter 包为什么是空的？

A: 这是一个占位包，用于未来可能需要的通用适配器接口。目前项目中各模块使用内部适配器实现（如 `RuntimeAdapter`），而不是统一的适配器接口。如果未来需要适配多个外部组件系统，可以考虑在此包中定义通用接口。

### Q: 什么时候应该创建适配器？

A: 在以下情况下应考虑使用适配器：
1. 集成第三方库或旧代码到 Mint 框架
2. 连接模块 A 和模块 B 的不兼容接口
3. 将运行时组件适配为框架层组件（或反之）
4. 为特定性能优化提供替代实现

### Q: 适配器与包装器（Wrapper）有什么区别？

A:
- **适配器**：改变接口，使不兼容的类可以一起工作
- **包装器**：保持接口不变，添加额外功能

示例：
```go
// 适配器：改变接口
type EditorAdapter struct {
    editor *ThirdPartyEditor
}
func (a *EditorAdapter) Paint() { ... }  // 不同的接口

// 包装器：添加功能
type LoggingWrapper struct {
    Component component.Component
}
func (w *LoggingWrapper) Paint() {
    log.Println("Painting")
    w.Component.Paint()  // 相同的接口
}
```

### Q: 如何测试适配器？

A: 适配器测试应验证：
1. 输入是否正确转换为目标格式
2. 事件是否正确转发
3. 边界条件是否正确处理

示例：
```go
func TestRuntimeAdapter_OnEvent(t *testing.T) {
    adapter := NewRuntimeAdapter()
    
    // 模拟事件
    ev := MouseEvent{X: 10, Y: 20}
    handled := adapter.OnEvent(ev)
    
    // 验证事件被正确转发和处理
    assert.True(t, handled)
    assert.True(t, adapter.IsSelectionActive())
}
```

### Q: 适配器是否应该缓存数据？

A: 通常不推荐缓存，因为这可能导致状态不一致。适配器应该只是薄层包装。如需缓存，应在目标系统中实现，而非在适配器中。

### Q: 适配器的性能影响如何？

A: 良好的适配器设计应具有接近零的性能开销：
- 使用嵌入类型避免数据复制
- 避免在适配路径中进行复杂计算
- 使用接口而非具体类型减少内存占用

### Q: 是否需要统一的 `runtime/adapter` 接口？

A: 目前不需要。各模块的适配器针对不同的桥接需求，统一接口反而会增加复杂性。如果未来有多个外部组件系统需要适配，再考虑定义通用接口。

### Q: 适配器和依赖注入（DI）有什么关系？

A: 适配器和 DI 是不同的概念：
- **适配器**：解决接口不兼容问题
- **依赖注入**：解耦组件，便于测试和配置

可以在 DI 容器中注入适配器实例，实现松耦合：

```go
// DI 容器配置
func (c *Container) ProvideAdapter() *RuntimeAdapter {
    return selection.NewRuntimeAdapter()
}

// 使用
func (app *App) Run(adapter *RuntimeAdapter) {
    // 注入的适配器
}
```
