# Fiber-first Paint Architecture

## Overview

This document describes the Fiber-first Paint Architecture that strictly separates VNode (description) from Instance (runtime entity).

## Core Design Principle

**VNode = Description (created every render)**
**Instance = Runtime Entity (persists across renders)**

## Architecture

```
用户代码创建 ButtonVNode
       ↓
CreateFiber(vnode)
       ↓
vnode.CreateInstance() → 创建 ButtonInstance
       ↓
Fiber.Instance = instance (持久化)
       ↓
VNode 被丢弃，不再被引用
       ↓
Paint阶段: instance.Paint(x, y)
```

## Three Roles

| Role | Responsibility | Lifetime |
|------|----------------|----------|
| VNode | Description (type, props) | Created each render, then discarded |
| Fiber | Tree structure, scheduling | Runtime persistent |
| Instance | Behavior + State | Runtime persistent |

## ComponentInstance Interface

```go
type ComponentInstance interface {
    // Lifecycle
    Init(props Props)
    Destroy()
    OnMount()
    OnUnmount()

    // Props Management
    SetProps(props Props) bool
    GetProps() Props

    // Identification
    Key() string
    SetKey(key string)

    // Dirty flag
    MarkDirty()
    IsDirty() bool
}
```

## Specialized Interfaces

```go
// For components that can paint
type PaintableInstance interface {
    ComponentInstance
    Paint(x, y int) []paint.DrawCmd
}

// For components that can receive focus
type FocusableInstance interface {
    ComponentInstance
    SetFocus(focused bool)
    HasFocus() bool
    IsDisabled() bool
}

// For components that handle actions
type ActionHandlerInstance interface {
    ComponentInstance
    CanHandleAction(actionType string) bool
    HandleAction(actionType string, payload interface{}) bool
}
```

## InstanceFactory

```go
type InstanceFactory interface {
    CreateInstance() ComponentInstance
}
```

## Key Implementation Rules

1. **VNode 不持有运行期状态** - VNode 只描述配置
2. **Instance 持有所有运行期状态** - focus, hover, disabled 等
3. **Fiber.Instance 持有 Instance 引用** - 不持有 VNode 引用
4. **CloneFiber 复用 Instance** - Instance 永远不被 clone
5. **Paint 阶段使用 Instance.Paint()** - 不访问 VNode

## Button Implementation Example

### ButtonVNode (Description)

```go
// ButtonVNode is the description for a Button
type ButtonVNode struct {
    *ui.ElementVNode
    label   string
    variant ButtonVariant
    size    ButtonSize
    // ... configuration only
}

// CreateInstance implements InstanceFactory
func (b *ButtonVNode) CreateInstance() rtui.ComponentInstance {
    props := rtui.Props{
        "label":   b.label,
        "variant": b.variant,
        "size":    b.size,
        // ...
    }
    return NewButtonInstance(props)
}
```

### ButtonInstance (Runtime Entity)

```go
// ButtonInstance is the runtime entity for Button
type ButtonInstance struct {
    // Props from VNode
    label   string
    variant ButtonVariant
    size    ButtonSize

    // Runtime state
    hasFocus  bool
    isHovered bool
    disabled  bool
}

// Paint generates draw commands
func (inst *ButtonInstance) Paint(x, y int) []paint.DrawCmd {
    // Use instance state (hasFocus, isHovered, etc.)
    // No VNode dependency!
}
```

## Fiber Flow

```go
// CreateFiber - creates Fiber and Instance
func CreateFiber(vnode VNode) *Fiber {
    // ... extract props from vnode

    // Create Instance from VNode
    var instance ComponentInstance
    if factory, ok := vnode.(InstanceFactory); ok {
        instance = factory.CreateInstance()
    }

    return &Fiber{
        Props:    props,
        Style:    style,
        Instance: instance,  // Instance persists
        // NO VNode reference!
    }
}

// CloneFiber - reuses Instance
func CloneFiber(fiber *Fiber) *Fiber {
    return &Fiber{
        // ... copy other fields
        Instance: fiber.Instance,  // REUSE, never clone
    }
}
```

## Benefits

1. **VNode 生命周期清晰** - 每次渲染创建，之后丢弃
2. **Instance 状态持久** - 跨渲染保持 focus, hover 等状态
3. **无 VNode 运行期依赖** - Fiber 只持有 Instance
4. **并发安全** - WIP 树和 current 树共享 Instance
5. **支持 time-slicing** - Instance 不受重渲染影响
6. **符合文档设计** - 严格按照 fiber_paint.md 实现

## Files Changed

- `runtime/ui/instance.go` - ComponentInstance interface and BaseComponentInstance
- `runtime/ui/fiber.go` - Added Instance field
- `runtime/ui/fiber_util.go` - CreateFiber creates Instance, CloneFiber reuses Instance
- `runtime/compute/adapter_fiber.go` - Uses Instance.Paint()
- `components/button/button_instance.go` - ButtonInstance implementation
- `components/button/button.go` - ButtonVNode.CreateInstance()

