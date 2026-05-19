# Engine Example

这是一个展示 `runtime/engine` 包基本功能的示例程序。

## 功能演示

本示例展示了引擎的核心功能：

1. **帧调度渲染** - 60fps 帧率驱动渲染循环
2. **鼠标事件处理** - 点击按钮触发回调
3. **光标动画** - 移动的闪烁光标效果
4. **焦点管理** - 点击按钮时显示焦点状态（黄色边框）
5. **动态状态栏** - 显示当前获得焦点的按钮

## 运行方式

```bash
go run ./examples/engine/origin
```

## 控制说明

- **鼠标点击** - 点击按钮触发点击事件，焦点会切换到被点击的按钮
- **ESC 或 Ctrl+C** - 退出程序
- 点击 "Exit" 按钮也可以退出程序

## 代码结构

### Button - 按钮组件
实现了以下接口：
- `engine.Renderable` - 绘制按钮
- `event.MouseEventHandler` - 处理鼠标点击
- `runtime.FocusableComponent` - 管理焦点状态

### BlinkingCursor - 闪烁光标
- 实现了 `engine.Updatable` 接口
- 每 500ms 切换可见状态
- 带有移动动画效果

### Root - 根组件
- 管理所有子组件
- 构建 `LayoutBox` 用于事件命中测试
- 绘制标题、说明和状态栏

## 关键点

### 事件处理

为了让事件系统正确工作，需要：

1. 组件实现 `event.MouseEventHandler` 接口：
```go
func (b *Button) HandleMouse(ev *event.MouseEvent, localX, localY int) bool {
    if ev.Type == event.MousePress && ev.Click == event.MouseLeft {
        // 处理点击
        return true
    }
    return false
}
```

2. 创建 `LayoutBox` 时设置 `Node` 字段：
```go
node := &runtime.LayoutNode{
    ID: btn.ID(),
    Component: &runtime.ComponentRef{
        Instance: btn,  // 指向实际的组件实例
    },
    // ... 位置信息
}

boxes[i] = runtime.LayoutBox{
    // ...
    Node: node,  // 设置 Node 引用
}
```

### 焦点管理

组件需要实现 `runtime.FocusableComponent` 接口：
```go
func (b *Button) SetFocus(focus bool) {
    b.focused = focus
    // 触发重绘
}

func (b *Button) IsFocusable() bool {
    return true
}
```

## 架构说明

```
Engine
├── Frame Loop (60fps)
│   ├── Update()    - 更新组件状态 (光标闪烁、移动)
│   ├── Paint()     - 绘制到缓冲区
│   └── Render()    - 输出到终端
└── Event Loop
    └── Hit Test -> Dispatch to Component -> HandleMouse()
```
