# Portal 系统实施 (Phase 3)

## 概述

Portal 系统允许组件将子组件渲染到组件树的不同位置，这对于以下场景特别有用：

- **模态框和对话框** (Modals & Dialogs)
- **提示框** (Tooltips)
- **下拉菜单** (Dropdowns)
- **通知** (Toasts)
- **悬浮层** (Overlays)

## 架构设计

### PortalRoot 创建流程

```
1. 应用顶层创建 PortalRoot 组件
   ↓
   props["portalRootId"] = "root-id"
   ↓
2. Render 阶段：VNode 转换为 Fiber
   ↓
3. Commit 阶段：Reconciler.linkPortalsToRoots()
   ├─ 收集所有 props["portalRootId"] 的节点
   └─ 建立映射表：portalRootId -> Fiber
   ↓
4. 再次遍历树，链接 Portal → PortalRoot
   ├─ 读取 props["portalRoot"]
   ├─ 查找映射表中的目标 Fiber
   └─ 设置 fiber.PortalRoot = targetFiber
   ↓
5. Layout 阶段：使用 fiber.PortalRoot 进行布局计算
   ↓
6. Paint 阶段：在 PortalRoot 的位置绘制 Portal 内容
```

### 核心组件

#### 1. Fiber.PortalRoot 字段

**位置**: `runtime/ui/fiber.go`

```go
type Fiber struct {
    // ...
    // ✨ Portal Root (Phase 3.1)
    // 指定该节点在布局/渲染阶段应该挂载的目标 Fiber
    // 用于 Portal 组件将子组件挂载到树的不同位置
    // nil 表示正常挂载（遵循父级树结构）
    PortalRoot *Fiber
}
```

**用途**：
- 存储 Portal 目标 Fiber 的引用
- 在布局阶段，节点将使用 PortalRoot 作为父级进行布局计算
- 在渲染阶段，节点将在 PortalRoot 的位置绘制

#### 2. OverlayManager

**位置**: `runtime/layout/overlay_manager.go`

```go
type OverlayManager struct {
    mu    sync.RWMutex
    stack []*OverlayEntry  // 按优先级排序的 Portal 条目栈
    entries map[string]*OverlayEntry  // Portal ID 到条目的映射
    dirty bool  // 缓存失效标志
}

type OverlayEntry struct {
    ID           string       // 唯一标识符
    Box          *LayoutBox   // 布局框
    PortalRootID string       // 目标 PortalRoot 的 ID
    Priority     int          // Z 顺序优先级（越高越靠上）
    Active       bool         // 是否活跃（未移除/隐藏）
}
```

**核心方法**：

- `Push(id, box, portalRootID, priority)` - 添加 Portal 到栈
- `Pop()` - 移除并返回栈顶 Portal
- `Top()` - 获取最高优先级的 Portal
- `GetAll()` - 获取所有活跃的 Portal（按优先级排序）
- `GetByID(id)` - 按ID获取 Portal
- `Remove(id)` - 移除指定的 Portal
- `Clear()` - 清空所有 Portal
- `Size()` - 获取活跃 Portal 数量

#### 3. Engine.overlayManager

**位置**: `runtime/layout/types.go`

```go
type Engine struct {
    // ...
    overlayManager *OverlayManager  // 管理 Portal 浮层
}

// GetOverlayManager 返回布局引擎的 OverlayManager
func (e *Engine) GetOverlayManager() *OverlayManager
```

#### 4. Reconciler.linkPortalsToRoots

**位置**: `internal/reconciler/reconciler.go`

**调用时机**: `CommitRoot()` 阶段

```go
func (r *Reconciler) CommitRoot() {
    // ...
    // Phase 0.5: Link portals to their PortalRoot targets
    r.linkPortalsToRoots(r.root)
    // ...
}

func (r *Reconciler) linkPortalsToRoots(root *Fiber) {
    // Step 1: Collect all PortalRoot nodes
    //   - Find all fibers with props["portalRootId"]
    //   - Build map: portalRootId -> Fiber

    // Step 2: Link Portal nodes to PortalRoot
    //   - For each fiber with props["portalRoot"]
    //   - Look up target Fiber in map
    //   - Set fiber.PortalRoot = target
}
```

**处理流程**：

1. **收集 PortalRoot**：遍历 Fiber 树，查找 `props["portalRootId"]` 属性
   ```go
   portalRoots := make(map[string]*Fiber)
   if portalRootID, ok := fiber.Props["portalRootId"].(string); ok {
       portalRoots[portalRootID] = fiber
   }
   ```

2. **链接 Portal**：再次遍历树，为每个 Portal 设置 `PortalRoot` 指针
   ```go
   if portalRootID, ok := fiber.Props["portalRoot"].(string); ok {
       if target, exists := portalRoots[portalRootID]; exists {
           fiber.PortalRoot = target
       }
   }
   ```

#### 5. syncPortalProperties

**位置**: `internal/reconciler/complete_work.go`

```go
// syncPortalProperties 从 Props 同步 Portal 属性到 Fiber
func syncPortalProperties(fiber *Fiber)
```

**处理逻辑**：

```go
func syncPortalProperties(fiber *Fiber) {
    switch fiber.Tag {
    case "portal":
        // PortalRoot 处理将在 commit 阶段完成
        // 因为目标 Fiber 可能还未创建
    case "modal":
        // Modal 使用内部 Portal 机制
        // 在单独的层 (LayerModal) 渲染
    }
}
```

## 使用流程

### 1. 定义 PortalRoot

在应用顶层使用 `props["portalRootId"]` 标识 PortalRoot：

```go
ui.Box(
    rtui.Props{
        "portalRootId": "main-root",  // 标识为 PortalRoot
        "layer":       types.LayerOverlay,
    },
    // PortalRoot 可以包含默认内容
)
```

### 2. 定义 Portal

使用 `props["portalRoot"]` 引用目标 PortalRoot：

```go
ui.Box(
    rtui.Props{
        "portalRoot": "main-root",  // 引用目标 PortalRoot 的 ID
        "position":  "fixed",
        "anchor":    "bottom",
    },
    ui.Text("This is a Portal!"),
)
```

### 3. 完整生命周期

```
用户代码
  ↓
Render 阶段: VNode → Fiber
  ├─ 创建 PortalRoot Fiber (props["portalRootId"])
  └─ 创建 Portal Fiber (props["portalRoot"])
  ↓
CompleteWork 阶段: syncPortalProperties()
  └─ 解析 Props，暂不链接
  ↓
CommitRoot 阶段: linkPortalsToRoots()
  ├─ 收集 PortalRoot: portalRootId → Fiber
  └─ 链接 Portal: Portal.PortalRoot → PortalRoot
  ↓
Layout 阶段
  └─ 使用 Fiber.PortalRoot 进行布局计算
  ↓
Paint 阶段
  └─ 在 PortalRoot 的位置绘制 Portal 内容
```

### 4. 示例：带 Tooltip 的按钮

```go
// 应用顶层
renderApp := func() rtui.VNode {
    return rtui.VNode{
        Type: rtui.VNodeElement,
        Tag:  "app",
        Children: []rtui.VNode{
            // PortalRoot (第1步)
            rtui.VNode{
                Type: rtui.VNodeElement,
                Tag:  "div",
                Props: rtui.Props{
                    "portalRootId": "main-root",
                },
            },
            // 主内容
            ui.Box(
                rtui.Props{},
                // 按钮
                ui.Box(
                    rtui.Props{
                        "onEnter": func(e Event) {
                            // 显示 tooltip - 动态创建 Portal
                        },
                    },
                    ui.Text("Hover me"),
                ),
            ),
        },
    }
}
```

## 优先级和 Z 顺序

### Layer 优先级

```
LayerBase      = 0   // 基础层
LayerOverlay   = 1   // 覆盖层
LayerModal     = 2   // 模态层
LayerTooltip   = 3   // 提示层
LayerInspector = 4   // 检查器层
```

### 优先级规则

1. **Layer 优先级**：高 Layer 在低 Layer 之上
2. **同一 Layer 内**：高 Priority 的节点在低 Priority 之上
3. **Portal 内部**：在 PortalRoot 的子树内，遵循正常的 Flex/Grid 布局

### 典型优先级设置

```go
// Modal
priority: 500, layer: LayerModal

// Toast
priority: 600, layer: LayerOverlay

// Tooltip
priority: 1000, layer: LayerTooltip

// Inspector
priority: 9999, layer: LayerInspector
```

## 测试覆盖

### OverlayManager 测试

`examples/fiber_firsts/portal_demo/portal_test.go` 包含：

- `TestOverlayManager_Basic` - 基本 Push/Pop/Top 操作
- `TestOverlayManager_PriorityOrdering` - 优先级排序
- `TestOverlayManager_GetByID` - ID 查询
- `TestOverlayManager_Remove` - 移除操作
- `TestOverlayManager_Clear` - 清空操作
- `TestOverlayManager_PushSameID` - 同 ID 覆盖
- `TestEngine_OverlayManager` - Engine 集成
- `TestOverlayEntry_Fields` - 条目字段验证

所有测试通过 ✅ (8/8)

### PortalRoot 链接测试

`internal/reconciler/portal_test.go` 包含：

- `TestPortalRoot_Linking` - 基本 Portal → PortalRoot 链接
- `TestLinkPortalsToRoots` - 双 PortalRoot 链接测试
- `TestPortalRoot_MultiplePortals` - 多 Portal 同目标测试
- `TestPortalRoot_NonExistentTarget` - 不存在目标处理
- `TestPortalRoot_NestedPortals` - 嵌套 Portal 结构

所有测试通过 ✅ (5/5)

### Portal 使用示例

`examples/fiber_firsts/portal_demo/portal_example.go` 包含：

- `TestPortalRootAndPortal_Simple` - 简单 PortalRoot + Portal
- `TestMultiplePortals` - 多 Portal 管理
- `TestPortalInRealScenario` - 真实场景模拟

## 与现有系统的集成

### 与 Layer 系统

- **Layer 系统**：处理同一树内的 Z 顺序（渲染顺序）
- **Portal 系统**：处理跨树的挂载（布局和渲染位置分离）
- 两者协同工作：
  - Portal 可以挂载到任意 Layer
  - 同一 Layer 内的 Portal 按 Priority 排序

### 与 PositionFixed

- Portal 通常与 `PositionFixed` 配合使用
- Fixed 定位使 Portal 脱离父级布局流
- 以 viewport 为参考系，实现自适应布局

### 与 Modal

- Modal 使用内部 Portal 机制
- 自动挂载到 `LayerModal`
- 支持 `centered` 属性（`PositionFixed + AnchorCenter`）

## 未来扩展

### 短期

1. **Portal 组件实现**
   - 实现可用的 `<Portal>` 组件 API
   - 支持 `<PortalRoot>` 和 `<Portal>` 组件

2. **Focus Trap**
   - Modal 打开时将焦点 Trap 在 Modal 内
   - 支持 Tab/Shift+Tab 循环导航

### 中期

1. **Portal 动画**
   - Portal 进入/退出动画
   - 支持缩放、淡入淡出等效果

2. **Portal 上下文**
   - Portal 可以访问目标 PortalRoot 的上下文
   - 支持数据共享和事件传递

### 长期

1. **嵌套 Portal**
   - 支持 Portal 内部再嵌套 Portal
   - 管理 Portal 层级关系

2. **Portal 生命周期钩子**
   - `onPortalOpen` / `onPortalClose`
   - `onPortalEnter` / `onPortalExit`

## 技术要点

### 并发安全

OverlayManager 使用 `sync.RWMutex` 保证并发安全：

- `Push` / `Pop` / `Remove` / `Clear` 使用写锁
- `Top` / `GetAll` / `GetByID` / `Size` 使用写锁（简化实现）

### 性能优化

- `dirty` 标志避免不必要的栈重建
- `entries` 映射提供 O(1) ID 查找
- 延迟重建栈（只在访问时重建）

### 内存管理

- `Remove` / `Pop` 会从映射中删除条目
- `Clear` 清空所有条目和映射
- 避免内存泄漏

## 示例代码

### 简单 Tooltip Portal

```go
// 按钮组件
button := ui.Box(
    rtui.Props{
        "onEnter": func(e Event) {
            // 显示 tooltip
            overlayManager.Push("tooltip", tooltipBox, "main-root", 1000)
        },
        "onLeave": func(e Event) {
            // 隐藏 tooltip
            overlayManager.Remove("tooltip")
        },
    },
    ui.Text("Hover me"),
)

// Tooltip 内容
tooltipBox := &layout.LayoutBox{
    Width:  20,
    Height: 3,
    X:      10,
    Y:      20,
    // PositionFixed + Anchor 定位
}
```

### Modal Portal

```go
// Modal 组件
modal := ui.Box(
    rtui.Props{
        "position": "fixed",
        "anchor":   "center",
        "layer":    types.LayerModal,
        "width":    50,
        "height":   20,
    },
    ui.Text("Modal Title"),
    ui.Text("Modal Content"),
    ui.Button("Close"),
)
```

## 总结

Portal 系统为 Mint TUI 框架提供了跨树渲染能力，使得 Modal、Tooltip、Toast 等组件可以灵活地渲染在应用的不同位置，而不受组件树结构的限制。

核心特性：

✅ 脱离原有树结构
✅ 支持优先级和 Z 顺序
✅ 与 Layer 系统无缝集成
✅ 支持PositionFixed 定位
✅ 并发安全
✅ 良好的测试覆盖

未来方向：

📌 Portal 组件 API
📌 Focus Trap
📌 Portal 动画
📌 嵌套 Portal
