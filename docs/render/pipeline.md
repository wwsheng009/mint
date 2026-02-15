# 渲染流程文档

## 概述

Mint 框架使用双阶段渲染管线：
1. **布局阶段** (Layout Phase): 计算每个节点的位置和尺寸
2. **绘制阶段** (Paint Phase): 将布局结果渲染到缓冲区

## 主要组件

### 1. Runtime (运行时核心)

位置: `runtime/core/runtime.go`

```go
type Runtime struct {
    platform         platform.RuntimePlatform    // 平台抽象（屏幕、输入、信号）
    layoutEngine     *layout.Engine             // 布局引擎
    focusManager    *focus.ManagerV3           // 焦点管理器
    stateTracker    *state.Tracker            // 状态追踪器
    actionDispatcher *action.Dispatcher        // Action 分发器
    keyMap          *input.KeyMap             // 输入映射
    contextManager  *ContextManager        // 上下文管理器
    root            layout.Node                // 根节点
    buffer           *paint.Buffer              // 绘制缓冲区
    dirtyTracker    *paint.DirtyTracker        // 脏区域跟踪器
    running         bool                        // 是否运行中
    windowWidth     int                         // 窗口尺寸
    windowHeight    int
}
```

### 2. 渲染流程入口

#### `Runtime.Render()` - 主渲染方法

位置: `runtime/core/runtime.go:184-209`

```go
func (r *Runtime) Render() error {
    // 1. 清空缓冲区（或只清空脏区域）
    if r.dirtyTracker.IsAllDirty() {
        r.buffer = paint.NewBuffer(r.windowWidth, r.windowHeight)
    }

    // 2. 绘制组件
    if r.root != nil {
        r.paintNode(r.root, 0, 0)
    }

    // 3. 将缓冲区内容写入屏幕
    r.writeToScreen()

    // 4. 清除脏标记
    r.dirtyTracker.Clear()

    return nil
}
```

#### `Runtime.Update()` - 更新循环

位置: `runtime/core/runtime.go:153-181`

```go
func (r *Runtime) Update() error {
    // 1. 记录更新前状态
    before := r.stateTracker.BeforeAction()

    // 2. 处理输入
    if err := r.processInput(); err != nil {
        return err
    }

    // 3. 记录更新后状态
    r.stateTracker.AfterAction(before)

    // 4. 布局
    if err := r.layout(); err != nil {
        return err
    }

    // 5. 标记脏区域
    r.markDirty()

    return nil
}
```

### 3. 布局阶段

#### `Runtime.layout()` - 布局入口

位置: `runtime/core/runtime.go` (需要查找 layout 方法)

布局引擎负责：
1. 测量每个组件的理想尺寸
2. 根据约束分配实际空间
3. 计算每个组件的最终位置

### 4. 绘制阶段

#### `Runtime.paintNode()` - 绘制节点

位置: `runtime/core/runtime.go:198-200`

这个方法需要从 VNode 获取对应的绘制逻辑：
- 检查节点类型（Text, Element, Component, Fragment）
- 调用相应的绘制方法

## 数据流

```
VNode (声明式)
    ↓
Fiber (树结构)
    ↓
ComputedBox (布局结果)
    ↓
PaintContext (绘制上下文)
    ↓
Buffer (最终输出)
```

## 关键接口

### VNode 接口

声明式节点接口，所有 UI 组件都实现此接口：
- `Type()`: 节点类型
- `Props()`: 属性
- `Children()`: 子节点
- `SetChildren()`: 设置子节点

### Paintable 接口

可绘制组件接口：
- `Paint(buffer, x, y)`: 绘制到缓冲区

## 渲染顺序

1. **从上到下，从左到右**: 父节点先绘制，子节点后绘制
2. **后绘制的覆盖先绘制的**: 子节点会覆盖父节点的绘制区域
3. **脏区域优化**: 只重新绘制标记为脏的区域

## 组件绘制优先级

1. 背景/背景层
2. 普通组件层
3. 浮动层/焦点层
4. 覆盖层/弹窗层
