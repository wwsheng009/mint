# Registry

组件注册表和工厂系统，提供组件的注册、查找和实例化能力。

> **当前状态**：此包为占位包，未包含实现代码。框架层实现了完整的组件工厂系统（`framework/component/factory.go`），运行时层可能需要更轻量的注册机制。

## 职责

- **组件注册**：注册组件类型到注册表（计划中）
- **组件创建**：通过工厂模式创建组件实例
- **类型查找**：按类型名称查找注册的组件创建器
- **声明式支持**：支持从 DSL/Spec 创建组件

## 核心概念

### 工厂模式

工厂模式（Factory Pattern）用于创建对象而不需要指定创建对象的具体类。在 Mint 框架中：

1. **组件注册**：将组件类型映射到创建函数
2. **Spec 驱动创建**：通过组件规范（Spec）声明式地创建组件
3. **嵌套组件支持**：递归创建子组件树
4. **全局工厂**：提供全局访问的组件工厂实例

### 工厂 vs 注册表

- **工厂**：负责对象的创建逻辑，管理类型到创建函数的映射
- **注册表**：负责类型的注册和查找，通常与工厂结合使用

在 Mint 中，`Factory` 内部维护了一个注册表（`creators map[string]Creator`），两者合而为一。

### 现有实现

#### Framework 层：Factory (framework/component/factory.go)

框架层提供了完整的组件工厂系统：

```go
// Factory 组件工厂
type Factory struct {
    mu       sync.RWMutex
    creators map[string]Creator  // 注册表
}

// Creator 组件创建器函数类型
type Creator func(spec *Spec) (Node, error)

// Spec 组件规范
type Spec struct {
    Type     string                    // 组件类型
    ID       string                    // 组件 ID
    Props    map[string]interface{}    // 静态属性
    Children []*Spec                   // 子组件规范
    Actions  map[string]interface{}    // 事件处理
}
```

**核心方法**：
- `Register(typ string, creator Creator)` - 注册组件创建器
- `CreateFromSpec(spec *Spec) (Node, error)` - 从规范创建组件
- `Create(typ string, props map[string]interface{}) (Node, error)` - 从类型和属性创建
- `HasType(typ string) bool` - 检查类型是否已注册
- `GetTypes() []string` - 获取所有已注册类型
- `Count() int` - 返回已注册类型数量
- `Clear()` - 清空所有注册

#### 全局工厂便利函数

```go
// 全局工厂实例
var globalFactory = NewFactory()

// 便利函数
func Register(typ string, creator Creator)               // 注册到全局工厂
func CreateFromSpec(spec *Spec) (Node, error)           // 使用全局工厂创建
func Create(typ string, props map[string]interface{}) (Node, error)
func HasType(typ string) bool                          // 检查类型
func GetFactory() *Factory                             // 获取全局工厂
```

#### Spec 构建器

`Spec` 提供链式调用 API 便于构建组件规范：

```go
spec := NewSpec("button").
    WithID("submit-btn").
    WithProp("label", "Submit").
    WithProp("disabled", false).
    WithAction("click", func() { ... })
```

## 使用示例

### 组件注册

```go
import (
    "github.com/wwsheng009/mint/framework/component"
)

// 定义组件创建器
func createButton(spec *component.Spec) (component.Node, error) {
    label := spec.GetPropString("label", "Button")
    disabled := spec.GetPropBool("disabled", false)
    
    btn := button.New(label)
    if disabled {
        btn.SetDisabled(true)
    }
    return btn, nil
}

// 注册到工厂
func init() {
    component.Register("button", createButton)
}
```

### 从 Spec 创建组件

```go
// 定义组件规范
btnSpec := component.NewSpec("button").
    WithID("submit-btn").
    WithProp("label", "Submit").
    WithProp("disabled", false)

// 创建组件实例
btnNode, err := component.CreateFromSpec(btnSpec)
if err != nil {
    log.Fatal(err)
}
```

### 创建嵌套组件树

```go
// 定义复杂的组件树
formSpec := component.NewSpec("form").
    WithID("login-form").
    WithChild(
        component.NewSpec("input").
            WithID("username").
            WithProp("placeholder", "Username"),
    ).
    WithChild(
        component.NewSpec("input").
            WithID("password").
            WithProp("placeholder", "Password").
            WithProp("type", "password"),
    ).
    WithChild(
        component.NewSpec("button").
            WithID("submit".
            WithProp("label", "Login").
            WithAction("click", handleSubmit),
    )

// 递归创建整个组件树
formNode, err := component.CreateFromSpec(formSpec)
```

### 使用全局工厂

```go
// 直接使用全局工厂函数
btn, err := component.Create("button", map[string]interface{}{
    "label":    "Click Me",
    "disabled": false,
})

// 检查类型是否已注册
if component.HasType("button") {
    // 类型已注册
}

// 获取所有已注册类型
types := component.GetFactory().GetTypes()
fmt.Println("Registered types:", types)
```

### 动态组件加载

```go
// 从配置文件加载组件定义
func loadComponentsFromConfig(configPath string) error {
    var config struct {
        Components []struct {
            Type    string                 `json:"type"`
            Config  map[string]interface{} `json:"config"`
        } `json:"components"`
    }
    
    data, err := os.ReadFile(configPath)
    if err != nil {
        return err
    }
    
    if err := json.Unmarshal(data, &config); err != nil {
        return err
    }
    
    // 动态注册组件
    for _, compConfig := range config.Components {
        creator := createCreatorForType(compConfig.Type)
        component.Register(compConfig.Type, creator)
    }
    
    return nil
}
```

## 核心类型

### Factory (framework/component/factory.go)

```go
type Factory struct {
    mu       sync.RWMutex       // 读写锁，保护 creators map
    creators map[string]Creator // 类型到创建器的映射
}
```

### Creator

```go
type Creator func(spec *Spec) (Node, error)
```

创建器函数类型，接收组件规范 `Spec`，返回组件节点 `Node` 或错误。

### Spec

```go
type Spec struct {
    Type     string                    // 组件类型（必需）
    ID       string                    // 组件 ID（可选）
    Props    map[string]interface{}    // 静态属性（可选）
    Children []*Spec                   // 子组件规范（可选）
    Actions  map[string]interface{}    // 事件处理（可选）
}
```

**方法**：
- `WithID(id string) *Spec` - 设置 ID
- `WithProp(key string, value any) *Spec` - 设置属性
- `WithProps(props map[string]any) *Spec` - 批量设置属性
- `WithChildren(children ...*Spec) *Spec` - 设置子组件
- `WithChild(child *Spec) *Spec` - 添加子组件
- `WithAction(name string, handler any) *Spec` - 添加事件处理
- `GetProp(key string) (any, bool)` - 获取属性
- `GetPropString(key, default string) string` - 获取字符串属性
- `GetPropInt(key string, default int) int` - 获取整数属性
- `GetPropBool(key string, default bool) bool` - 获取布尔属性
- `Validate() error` - 验证规范
- `Clone() *Spec` - 克隆规范

## 文件结构

```
runtime/registry/
└── README.md       # 本文档
                   # registry.go 未实现（计划中）
                   # factory.go 未实现（计划中）

framework/component/
└── factory.go      # 框架层组件工厂实现
    ├── Factory                 # 工厂结构
    ├── Creator                 # 创建器类型
    ├── Spec                    # 组件规范
    ├── NewFactory              # 创建工厂
    ├── Register/Unregister     # 注册/注销
    ├── Create/CreateFromSpec   # 创建组件
    ├── HasType/GetTypes/Count  # 查询方法
    ├── Clear                   # 清空注册
    ├── 全局便利函数           # Register, Create, 等
    └── Spec 链式方法          # WithID, WithProp, 等
```

## 依赖

### 外部依赖
- 无（当前 `runtime/registry` 为空包）
- **framework/component/factory.go**: 仅使用标准库 (`sync`, `fmt`)

### 内部依赖
- 无（当前 `runtime/registry` 为空包）
- **framework/component/factory.go**: 依赖 `framework/component` 包的 `Node` 和 `Container` 接口

### 依赖此包的模块
- `framework/component` - 使用全局工厂创建组件
- `framework/layout` - 可能需要从 Spec 创建布局节点

## 与其他模块集成

### 与 DSL 模块集成

工厂系统与 DSL 模块协同工作，实现声明式 UI：

```go
// DSL 解析器解析配置后生成 Spec
func parseDSLConfig(dslConfig map[string]interface{}) *component.Spec {
    typ := dslConfig["type"].(string)
    spec := component.NewSpec(typ)
    
    if id, ok := dslConfig["id"]; ok {
        spec.WithID(id.(string))
    }
    
    if props, ok := dslConfig["props"]; ok {
        spec.WithProps(props.(map[string]interface{}))
    }
    
    // 递归解析子组件
    if childrenData, ok := dslConfig["children"]; ok {
        for _, childData := range childrenData.([]map[string]interface{}) {
            spec.WithChild(parseDSLConfig(childData))
        }
    }
    
    return spec
}
```

### 与 State/Binding 模块集成

组件工厂可以与状态绑定结合，自动处理属性绑定：

```go
func createBoundTextComponent(spec *component.Spec) (component.Node, error) {
    // 检查是否有状态绑定
    if bindKey, ok := spec.GetProp("bindValue"); ok {
        // 创建绑定组件
        text := text.New()
        binding.Bind(text, "content", bindKey.(string))
        return text, nil
    }
    
    // 创建普通组件
    content := spec.GetPropString("content", "")
    return text.New(content), nil
}
```

### 与 AI Controller 集成

AI Controller 可以使用工厂创建组件：

```go
// AI Agent 通过查询组件类型创建实例
func (c *RuntimeController) CreateComponent(typ string, props map[string]interface{}) (ID, error) {
    node, err := component.Create(typ, props)
    if err != nil {
        return "", err
    }
    
    // 将创建的组件添加到运行时
    id := c.AddComponent(node)
    return id, nil
}
```

### 与 DevTools 集成

DevTools 可以查询工厂中注册的所有组件类型：

```go
// DevTools 列出可用的组件类型
func (dt *DevTools) ListComponentTypes() []string {
    return component.GetFactory().GetTypes()
}

// DevTools 显示组件类型信息
func (dt *DevTools) DescribeComponentType(typ string) (ComponentTypeInfo, error) {
    if !component.HasType(typ) {
        return ComponentTypeInfo{}, fmt.Errorf("type %s not registered", typ)
    }
    
    // 返回组件类型元数据（可能在注册时提供）
    return component.GetFactory().DescribeType(typ)
}
```

## 最佳实践

### 1. 在包初始化时注册组件

```go
// good_components/button/button.go
package button

import "github.com/wwsheng009/mint/framework/component"

func init() {
    component.Register("button", createButton)
}

func createButton(spec *component.Spec) (component.Node, error) {
    // 创建逻辑
}
```

### 2. 创建器应处理所有必需参数

```go
func createInput(spec *component.Spec) (component.Node, error) {
    // 必需参数
    label := spec.GetPropString("label", "")
    if label == "" {
        return nil, fmt.Errorf("input requires 'label' property")
    }
    
    // 可选参数
    placeholder := spec.GetPropString("placeholder", "")
    value := spec.GetPropString("value", "")
    
    input := NewInput(label)
    input.SetPlaceholder(placeholder)
    input.SetValue(value)
    return input, nil
}
```

### 3. 使用 Spec 链式调用提高可读性

```go
// 推荐：使用链式调用
spec := component.NewSpec("button").
    WithID("submit").
    WithProp("label", "Submit").
    WithProp("disabled", false)

// 不推荐：手动填充 map
spec := &component.Spec{
    Type:  "button",
    ID:    "submit",
    Props: map[string]interface{}{
        "label":    "Submit",
        "disabled": false,
    },
}
```

### 4. 处理子组件时验证 Container 接口

```go
func createContainer(spec *component.Spec) (component.Node, error) {
    container := NewContainer()
    
    // 处理子组件
    for _, childSpec := range spec.Children {
        child, err := component.CreateFromSpec(childSpec)
        if err != nil {
            return nil, err
        }
        
        // 验证实现了 Container 接口
        if cont, ok := container.(component.Container); ok {
            cont.Add(child)
        } else {
            return nil, fmt.Errorf("container does not support children")
        }
    }
    
    return container, nil
}
```

### 5. 提供组件元数据

```go
type ComponentMeta struct {
    Name        string                 // 组件名称
    Description string                 // 描述
    Properties  map[string]Property    // 属性定义
    Events      []string               // 支持的事件
    Examples    []string               // 示例
}

// 扩展 Factory 以支持元数据
type Factory struct {
    mu       sync.RWMutex
    creators map[string]Creator
    metadata map[string]ComponentMeta
}

func (f *Factory) RegisterWithMeta(typ string, creator Creator, meta ComponentMeta) {
    f.mu.Lock()
    defer f.mu.Unlock()
    f.creators[typ] = creator
    f.metadata[typ] = meta
}
```

### 6. 错误处理和验证

```go
func createButton(spec *component.Spec) (component.Node, error) {
    // 验证 spec
    if err := spec.Validate(); err != nil {
        return nil, err
    }
    
    // 验证必需属性
    if !spec.HasProp("label") {
        return nil, fmt.Errorf("button requires 'label' property")
    }
    
    // 验证属性类型
    label, ok := spec.GetPropString("label", "")
    if !ok {
        return nil, fmt.Errorf("'label' property must be a string")
    }
    
    // 创建组件
    btn, err := NewButton(label)
    if err != nil {
        return nil, fmt.Errorf("failed to create button: %w", err)
    }
    
    return btn, nil
}
```

## 常见问题

### Q: runtime/registry 包为什么是空的？

A: 这是一个占位包，用于未来可能需要的运行时层组件注册机制。当前框架层（`framework/component/factory.go`）提供了完整的工厂系统。运行时层可能需要更轻量的注册机制，或者根本不需要（运行时更多是被动响应框架层的组件实例）。

### Q: 如何防止组件类型冲突？

A: 几种策略：
1. **使用命名空间**：注册前添加命名空间前缀
   ```go
   component.Register("myapp.button", createButton)
   component.Register("otherlib.button", createOtherButton)
   ```

2. **使用唯一注册器**：每个包维护自己的工厂实例
   ```go
   var myAppFactory = component.NewFactory()
   myAppFactory.Register("button", createButton)
   ```

3. **先检查后注册**
   ```go
   if !component.HasType("button") {
       component.Register("button", createButton)
   }
   ```

### Q: 工厂如何处理组件生命周期？

A: 工厂只负责创建组件，不管理生命周期。组件的生命周期由：
- 框架层的组件容器管理
- 运行时的布局和渲染系统管理
- 主应用的代码管理

如需生命周期控制，可在组件自身实现（如 `Close()` 方法）。

### Q: 如何注册带有泛型的组件？

A: Go 1.18+ 支持泛型。可以为工厂添加泛型方法：

```go
// 泛型创建器
type GenericCreator[T any] func(spec *Spec) (T, error)

// 泛型工厂方法
func CreateGeneric[T any](typ string, props map[string]interface{}) (T, error) {
    creator, ok := globalFactory.GetCreator(typ)
    if !ok {
        var zero T
        return zero, fmt.Errorf("unknown type: %s", typ)
    }
    return creator(spec).(T), nil
}
```

### Q: 工厂性能如何？如何优化？

A: 工厂的主要性能因素：
1. **map 查找**：O(1) 时间复杂度，已优化
2. **反射**：如果使用 `reflect` 处理属性会有开销
3. **锁竞争**：多个 goroutine 并发注册/创建会有锁开销

优化建议：
- 在应用启动时预注册所有组件类型
- 避免频繁调用 `Unregister` / `Clear`
- 对于性能敏感的场景，使用多个工厂实例减少锁竞争

### Q: 是否支持组件版本管理？

A: 当前工厂不支持版本控制。如果需要：

```go
// 策略 1：在类型名称中包含版本
component.Register("button.v1", createButtonV1)
component.Register("button.v2", createButtonV2)

// 策略 2：扩展 Factory 支持版本
type Factory struct {
    creators map[string]map[string]Creator  // type -> version -> creator
}

func (f *Factory) RegisterVersion(typ, version string, creator Creator) {
    // ...
}
```

### Q: 如何实现组件的热重载？

A: 热重载需要：
1. 追踪组件的来源（哪个 `.so` 或插件）
2. 卸载时调用 `Unregister`
3. 重新加载时调用 `Register`

```go
type Plugin struct {
    Name      string
    Path      string
    LoadFunc  func() []ComponentRegistration
}

func (p *Plugin) Reload() error {
    // 1. 卸载旧的
    p.Unregister()
    
    // 2. 重新加载插件
    plugin, err := plugin.Open(p.Path)
    if err != nil {
        return err
    }
    
    // 3. 注册新的
    registrations := p.LoadFunc()
    for _, reg := range registrations {
        component.Register(reg.Type, reg.Creator)
    }
    
    return nil
}
```

### Q: 工厂模式 vs 反射？

A:
- **工厂模式**：编译期类型安全，性能好，但需要手动注册
- **反射**：动态创建，无需注册，但性能差，无类型安全

推荐使用工厂模式，除非特殊场景需要完全动态的组件创建。
