# Modal Component Plugin Architecture

本文档描述了 Modal 组件的插件化架构设计，说明其工作原理和其他 UI 组件如何借鉴这一模式。

---

## 目录

- [概述](#概述)
- [核心功能](#核心功能)
- [架构设计](#架构设计)
- [工作原理](#工作原理)
- [组件接口](#组件接口)
- [其他组件如何借鉴](#其他组件如何借鉴)
- [完整示例](#完整示例)

---

## 概述

Modal 组件实现了完整的**插件化架构**，使其能够：

1. **全局生效**：即使焦点在 Modal 内的按钮上，ESC 键也能关闭 Modal
2. **点击外部关闭**：用户点击 Modal 外部区域时触发关闭
3. **完全解耦**：框架层不依赖 Modal，应用层按需注册 Modal 支持
4. **多实例支持**：同时支持多个打开的 Modal（由全局注册表管理）

---

## 核心功能

### 1. 状态管理
- `isOpen`：Modal 是否打开
- `closeable`：是否允许关闭（可通过 Builder 设置锁定）

### 2. 事件处理
- `HandleAction`：处理 ActionCancel/ActionQuit（ESC 键）
- `HandleMouseMessage`：处理鼠标点击（检测点击在外部时关闭）
- `HandleKeyMessage`：处理 ESC 键（备用实现）

### 3. 全局注册表
- 跟踪所有打开的 Modal 实例
- 线程安全的注册/注销机制
- 避免重复注册（使用 `registered` 标志）

---

## 架构设计

### 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                        应用层 (App)                          │
│  ui.Run(App, ui.WithPluginSetup(func(app *framework.App){   │
│      app.AddMiddleware(modal.NewModalMiddleware())           │
│  }))                                                         │
└───────────────────────┬─────────────────────────────────────┘
                        │ 注册插件
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                        框架层 (Framework)                    │
│  framework/App.RegisterGlobalHandler()                      │
│  framework/App.AddMiddleware()                              │
└───────────────────────┬─────────────────────────────────────┘
                        │ 提供 HTTP/事件基础设施
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                      UI 组件层 (Modal)                       │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              ModalMiddleware                          │  │
│  │  • 拦截 ActionCancel/ActionQuit → 关闭 Modal         │  │
│  │  • 拦截 ActionClick (外部) → 关闭 Modal               │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           全局注册表 (globalRegistry)                  │  │
│  │  • 跟踪所有打开的 Modal                               │  │
│  │  • 支持多个 Modal 实例                                │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Modal 实例方法                           │  │
│  │  • HandleAction() - 处理 ESC 键                       │  │
│  │  • HandleMouseMessage() - 处理鼠标点击                │  │
│  │  • HandleKeyMessage() - 处理键盘消息                  │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 关键设计原则

#### 1. **框架完全解耦**
框架层 (`framework/app.go`) 不依赖任何 UI 组件：
```go
// ❌ 不好的做法：框架直接依赖 Modal
import "github.com/wwsheng009/mint/ui/components/modal"

// ✅ 好的做法：框架只提供注册接口
func (a *App) AddMiddleware(middleware action.ActionMiddleware) {
    a.actionRouter.AddMiddleware(middleware)
}
```

#### 2. **插件化注册**
应用层决定是否启用某个组件的功能：
```go
// 在你的 App 中注册 Modal 支持
ui.Run(App,
    ui.WithPluginSetup(func(app *framework.App) {
        app.AddMiddleware(modal.NewModalMiddleware())
    }),
)
```

#### 3. **单一职责**
每个类/组件只负责一件事：
- `ModalMiddleware`：拦截全局事件
- `globalRegistry`：管理打开的 Modal 列表
- `Instance`：单个 Modal 的渲染和组件级事件处理

---

## 工作原理

### 1. Modal 如何跟踪所有实例

Modal 使用一个全局注册表 (`modalRegistry`) 来跟踪所有打开的 Modal：

```go
var globalRegistry = &modalRegistry{
    modals: make(map[*Instance]bool),
}

// 注册（在 SetProps 中，当状态从关闭变为打开时）
func (r *modalRegistry) register(inst *Instance) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.modals[inst] = true
}

// 注销（在 SetProps 中，当状态从打开变为关闭时，或在 Destroy 中）
func (r *modalRegistry) unregister(inst *Instance) {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.modals, inst)
}

// 获取所有打开的 Modal
func (r *modalRegistry) getOpenModals() []*Instance {
    r.mu.RLock()
    defer r.mu.RUnlock()
    var result []*Instance
    for inst := range r.modals {
        if inst.isOpen {
            result = append(result, inst)
        }
    }
    return result
}
```

**关键点**：
- 使用 `sync.RWMutex` 保证线程安全
- 使用 `map` 实现快速查找
- 使用指针避免复制
- 带有 `isOpen` 检查，只返回真正打开的 Modal

### 2. 拦截 ESC 键

当用户按下 ESC 键时的流程：

```
ESC 键输入
    │
    ▼
InputProcessor → KeyMsg → ActionCancel
    │
    ▼
Router.Dispatch()
    │
    ▼
Middleware.Before()  // ModalMiddleware
    │
    ├── 检测到 ActionCancel
    │
    ├── 获取所有打开的 Modal：globalRegistry.getOpenModals()
    │
    ├── 遍历，找到第一个可关闭的 Modal
    │
    ├── modalImpl.isOpen = false
    │   modalImpl.emitCloseIntent()
    │
    └── Return nil  // 拦截 Action，阻止继续分发
```

代码实现：
```go
func (m *ModalMiddleware) handleKeyboardClose(act *action.Action) *action.Action {
    modals := globalRegistry.getOpenModals()
    for _, modalInst := range modals {
        if modalInst.isOpen && modalInst.closeable {
            modalInst.isOpen = false
            modalInst.dirty = true
            modalInst.emitCloseIntent()
            return nil  // 拦截
        }
    }
    return act  // 继续分发
}
```

### 3. 拦截点击外部

当用户点击鼠标时的流程：

```
鼠标点击 (X, Y)
    │
    ▼
InputProcessor → MouseMsg → ActionClick
    │
    ▼
Router.Dispatch()
    │
    ▼
Middleware.Before()  // ModalMiddleware
    │
    ├── 检测到 ActionClick 且 MouseMsg.Action == Press
    │
    ├── 获取所有打开的 Modal：globalRegistry.getOpenModals()
    │
    ├── 遍历模态框，检查点击坐标是否在任何一个内部
    │   │
    │   └─ modalInst.containsPoint(X, Y)  // 几何检测
    │
    ├── 如果所有 Modal 都不包含该点 → "点击外部"
    │
    ├── 关闭最顶层 Modal
    │   modalInst.isOpen = false
    │   modalInst.emitCloseIntent()
    │
    └── Return nil  // 拦截 Action，阻止目标组件接收点击
```

代码实现：
```go
func (m *ModalMiddleware) handleClickOutside(act *action.Action) *action.Action {
    mouseMsg, _ := act.Payload.(*runtimemsg.MouseMsg)
    modals := globalRegistry.getOpenModals()
    
    // 检查点击是否在任何 Modal 内部
    clickedInside := false
    for _, modalInst := range modals {
        if modalInst.isOpen && modalInst.inst.containsPoint(mouseMsg.X, mouseMsg.Y) {
            clickedInside = true
            break
        }
    }
    
    // 如果点击在外部，关闭最顶层的 Modal
    if !clickedInside {
        for i := len(modals) - 1; i >= 0; i-- {
            modalInst := modals[i]
            if modalInst.isOpen && modalInst.closeable {
                modalInst.isOpen = false
                modalInst.emitCloseIntent()
                return nil  // 拦截
            }
        }
    }
    
    return act  // 继续分发
}
```

### 4. 注册/注销机制

**注册时机**：在 `SetProps()` 中，当 `isOpen` 状态从 `false` 变为 `true` 时注册

```go
func (inst *Instance) SetProps(props Props) {
    oldOpen := inst.isOpen
    // ... 更新状态
    
    // 1. 状态从关闭变为打开：注册
    if !oldOpen && inst.isOpen {
        globalRegistry.register(inst)
        inst.registered = true
    }
    // 2. 状态从打开变为关闭：注销
    else if oldOpen && !inst.isOpen {
        globalRegistry.unregister(inst)
        inst.registered = false
    }
    // 3. 初始就是打开状态且未注册：注册
    else if inst.isOpen && !inst.registered {
        globalRegistry.register(inst)
        inst.registered = true
    }
}

func (inst *Instance) Destroy() {
    globalRegistry.unregister(inst)  // 确保注销
    inst.PaintableInstance.Destroy()
}
```

**关键设计**：
- 使用 `registered` 标志避免重复注册
- 在 `Destroy()` 中强制注销，防止内存泄漏
- 三种情况覆盖所有可能的启动场景

---

## 组件接口

Modal 实现了多个运行时接口：

### 1. ActionHandlerInstance
```go
type ActionHandlerInstance interface {
    HandleAction(act *action.Action) bool
}

// Modal 实现：处理 ESC 键
func (inst *Instance) HandleAction(act *action.Action) bool {
    switch act.Type {
    case action.ActionCancel, action.ActionQuit:
        inst.isOpen = false
        inst.emitCloseIntent()
        return true
    }
    return false
}
```

### 2. MouseMessageHandlerInstance
```go
type MouseMessageHandlerInstance interface {
    HandleMouseMessage(msg *MouseMsg) bool
}

// Modal 实现：处理点击外部
func (inst *Instance) HandleMouseMessage(msg *runtimemsg.MouseMsg) bool {
    if msg.Action == runtimemsg.MouseActionPress {
        if !inst.containsPoint(msg.X, msg.Y) {
            inst.isOpen = false
            inst.emitCloseIntent()
            return true
        }
    }
    return false
}
```

### 3. KeyMessageHandlerInstance
```go
type KeyMessageHandlerInstance interface {
    HandleKeyMessage(msg *KeyMsg) bool
}

// Modal 实现：处理 ESC 键
func (inst *Instance) HandleKeyMessage(msg *runtimemsg.KeyMsg) bool {
    if msg.Special == runtimeplatform.KeyEscape {
        inst.isOpen = false
        inst.emitCloseIntent()
        return true
    }
    return false
}
```

### 4. 事件传播层次

```
Router.Dispatch()
    │
    ├─ Middleware.Before()  ← 第一道防线：全局拦截 (ModalMiddleware)
    │   └─ 可返回 nil 拦截
    │
    ├─ Capture Phase        ↓
    │   └─ 从根到目标的捕获阶段
    │
    ├─ Target Phase         ↓
    │   └─ 目标组件的事件处理
    │       ├─ HandleAction()
    │       ├─ HandleMouseMessage()
    │       └─ HandleKeyMessage()
    │
    └─ Bubble Phase        ↓
        └─ 从目标到根的冒泡阶段
```

---

## 其他组件如何借鉴

### 场景 1：Dropdown

**需求**：
- 点击外部时关闭 Dropdown
- ESC 键关闭 Dropdown

**实现步骤**：

1. **创建 DropdownMiddleware**
```go
type DropdownMiddleware struct{}

func (m *DropdownMiddleware) Before(act *action.Action) *action.Action {
    // 处理 ESC 键
    if act.Type == action.ActionCancel {
        return m.closeDropdown()
    }
    
    // 处理点击外部
    if act.Type == action.ActionClick {
        return m.handleClickOutside(act)
    }
    
    return act
}
```

2. **创建全局注册表**
```go
var dropdownRegistry = struct {
    sync.RWMutex
    dropdowns map[*Instance]bool
}{
    dropdowns: make(map[*Instance]bool),
}

func registerDropdown(d *Instance) {
    dropdownRegistry.Lock()
    dropdownRegistry.dropdowns[d] = true
    dropdownRegistry.Unlock()
}

func getOpenDropdowns() []*Instance {
    dropdownRegistry.RLock()
    defer dropdownRegistry.RUnlock()
    var result []*Instance
    for d := range dropdownRegistry.dropdowns {
        if d.isOpen {
            result = append(result, d)
        }
    }
    return result
}
```

3. **在应用中注册**
```go
ui.Run(App,
    ui.WithPluginSetup(func(app *framework.App) {
        app.AddMiddleware(dropdown.NewDropdownMiddleware())
    }),
)
```

### 场景 2：Popover

**需求**：
- 点击外部关闭 Popover
- ESC 键关闭 Popover
- 关闭时弹出确认提示

**实现**：
```go
type PopoverMiddleware struct{
    showConfirm func() bool  // 支持注入确认回调
}

func (m *PopoverMiddleware) Before(act *action.Action) *action.Action {
    if act.Type == action.ActionClick {
        if m.handleClickOutside(act) {
            // 显示确认对话框
            if m.showConfirm() {
                return m.closePopover()
            }
            return nil  // 拦截，取消关闭
        }
    }
    return act
}
```

### 场景 3：ContextMenu

**需求**：
- 右键点击任意位置关闭
- 点击外部关闭
- ESC 键关闭

**实现**：
```go
func (m *ContextMenuMiddleware) Before(act *action.Action) *action.Action {
    // 右键点击（新建菜单前先关闭旧菜单）
    if act.Type == action.ActionRightClick {
        m.closeAllMenus()
        return act  // 继续处理新的右键点击
    }
    
    // 其他关闭机制...
    return act
}
```

### 关键模式总结

| 模式 | 用途 | 实现要点 |
|------|------|----------|
| **Middleware 拦截** | 全局事件拦截 | 在 `Before()` 中检查，返回 `nil` 拦截 |
| **全局注册表** | 跟踪实例 | 使用 `sync.Map` 或 `map + RWMutex` |
| **线程安全** | 防止并发问题 | 每个注册表操作都加锁 |
| **生命周期管理** | 避免内存泄漏 | 在 `Destroy()` 中强制注销 |
| **几何检测** | 外部点击检测 | 实现 `containsPoint(x, y)` 方法 |
| **分层处理** | 可选的组件级处理 | 实现 `Handle*Message()` 接口 |

---

## 完整示例

### Modal 集成到新应用

```go
package main

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/ui/components/modal"
    "github.com/wwsheng009/mint/framework"
)

func main() {
    ui.Run(myApp,
        ui.WithWidth(80),
        ui.WithHeight(30),
        
        // 关键：注册 Modal 支持
        ui.WithPluginSetup(func(app *framework.App) {
            app.AddMiddleware(modal.NewModalMiddleware())
        }),
    )
}

func myApp() ui.VNode {
    isOpen, setIsOpen := ui.UseStateBool(false)
    
    return ui.VStack(
        app.Button("Open Modal", func() {
            setIsOpen(true)
        }),
        
        modal.NewBuilder().
            Title("Hello").
            Content(app.Text("Modal content")).
            Open(isOpen).
            OnClose(func() {
                setIsOpen(false)
            }).
            Build(),
    )
}
```

---

## 最佳实践

### ✅ DO

1. **使用 Middleware 而不是 GlobalHandler**
   - Middleware 可以拦截任何 Action（有或无 target）
   - 代码更集中、更易维护

2. **全局注册表使用 `sync.RWMutex`**
   - 读多写少的场景，`RWMutex` 性能更好

3. **在 `Destroy()` 中强制注销**
   - 防止组件销毁后指针仍存在于注册表中

4. **实现 `containsPoint(x, y)` 用于几何检测**
   - 点击外部检测的核心方法

5. **使用插件化注册**
   - 应用层决定是否启用组件功能

### ❌ DON'T

1. **不要在框架层直接依赖 UI 组件**
   ```go
   // ❌ 错误
   import "github.com/wwsheng009/mint/ui/components/modal"
   
   // ✅ 正确
   // 框架只提供注册接口
   ```

2. **不要忘记线程安全**
   ```go
   // ❌ 错误
   func register(inst *Instance) {
       registry.modals[inst] = true  // 并发不安全
   }
   
   // ✅ 正确
   func register(inst *Instance) {
       registry.mu.Lock()
       defer registry.mu.Unlock()
       registry.modals[inst] = true
   }
   ```

3. **不要忽略边界情况**
   - 同时打开多个 Modal 的情况
   - Modal 被禁用（`closeable = false`）的情况
   - 点击坐标超出屏幕范围的情况

---

## 总结

Modal 组件展示了一个完整的、可复用的插件化架构：

1. **完全解耦**：框架不依赖任何具体 UI 组件
2. **插件化**：应用层按需注册组件功能
3. **全局拦截**：Middleware 可拦截任何 Action
4. **线程安全**：使用全局注册表管理实例
5. **生命周期管理**：正确的注册/注销机制

其他组件（Dropdown、Popover、ContextMenu 等）可以采用相同的模式来实现全局生效的交互行为。

---

## 相关文件

- `middleware.go` - Modal 的 Middleware 实现
- `instance.go` - Modal 实例和组件级事件处理
- `builder.go` - Builder 模式
- `props.go` - Props 定义
