# Focus System

焦点管理系统核心实现。

## 职责

- 焦点栈管理：维护当前 focused 组件
- 焦点导航逻辑：支持线性导航（Next/Prev）和几何导航（上下左右）
- 焦点陷阱（Focus Trap）：支持模态对话框、菜单等场景的焦点限制
- 焦点事件处理：向组件分发 Focus/Blur 事件

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 核心概念

### 组件焦点

组件通过实现 `FocusableComponent` 接口声明自己可聚焦：

```go
type FocusableComponent interface {
    IsFocusable() bool
    SetFocus(focused bool)
}
```

### 焦点管理器 (Manager)

`Manager` 是焦点系统的核心，负责：

1. **扫描可聚焦组件**：从 LayoutNode 树中收集所有 `IsFocusable() == true` 的组件
2. **维护焦点状态**：记录当前聚焦组件的索引
3. **导航功能**：
   - `FocusNext()` / `FocusPrev()`：线性导航（支持循环）
   - `FocusFirst()`：跳到第一个可聚焦组件
   - `FocusUp()` / `FocusDown()` / `FocusLeft()` / `FocusRight()`：几何导航
   - `FocusSpecific(id)`：聚焦指定组件
4. **焦点陷阱管理**：
   - `PushFocusTrap(trap)`：激活焦点陷阱（如模态对话框）
   - `PopFocusTrap()`：移除焦点陷阱
   - 在焦点陷阱激活时，焦点只能在其子树内循环

### 几何导航 (Geometric Navigation)

`GeometricNavigator` 基于组件的几何位置进行智能导航：

- **距离优先**：优先选择距离最近的组件
- **重叠加权**：水平/垂直方向的重叠区域会得到额外加分
- **适用场景**：网格布局、表单等需要空间导航的界面

导航评分算法：
```
Score = (maxDistance - distance) / maxDistance + overlapBonus * 0.5
```

### 焦点陷阱 (Focus Trap)

`FocusTrap` 用于限制焦点导航范围：

- **类型**：`TrapModal`（模态）、`TrapMenu`（菜单）、`TrapPopover`（弹出框）
- **行为**：陷阱激活时，`FocusNext/Prev` 只会在陷阱子树内循环
- **堆栈**：支持嵌套陷阱（如模套模），使用栈管理

## 使用示例

### 基本使用

```go
import "github.com/wwsheng009/mint/runtime/focus"

// 创建焦点管理器
fm := focus.NewManager(rootNode)

// 扫描可聚焦组件
fm.RefreshFocusables()

// 导航焦点
fm.FocusNext()
fom.FocusPrev()
fom.FocusFirst()

// 检查当前焦点
if id, ok := fm.GetFocused(); ok {
    fmt.Println("当前聚焦:", id)
}
```

### 几何导航

```go
// 使用方向键导航
leftId, _ := fm.FocusLeft()
rightId, _ := fm.FocusRight()
upId, _ := fm.FocusUp()
downId, _ := fm.FocusDown()
```

### 焦点陷阱（模态对话框）

```go
// 创建模态对话框的焦点陷阱
modalTrap := focus.NewFocusTrap(
    "modal1",
    focus.TrapModal,
    modalRootNode,
)

// 激活陷阱（关闭后调用 PopFocusTrap）
fm.PushFocusTrap(modalTrap)

// 现在焦点只能在 modalTrap 内循环
for i := 0; i < 10; i++ {
    fm.FocusNext() // 只在模态框内移动
}

// 关闭模态框
fm.PopFocusTrap()
```

### 检查焦点状态

```go
// 检查是否有焦点陷阱
if fm.HasActiveFocusTrap() {
    trap := fm.GetCurrentFocusTrap()
    fmt.Println("当前陷阱:", trap.ID)
}

// 检查特定组件是否有焦点
if fm.HasFocus("button1") {
    fmt.Println("button1 被聚焦")
}

// 获取所有可聚焦组件
focusables := fm.GetFocusableComponents()
fmt.Println("可聚焦组件数:", fm.GetFocusableCount())
```

## 核心类型

| 类型 | 说明 |
|------|------|
| `Manager` | 焦点管理器
| `GeometricNavigator` | 几何导航器 |
| `FocusTrap` | 焦点陷阱 |
| `TrapManager` | 陷阱堆栈管理器 |
| `NavigationDirection` | 导航方向枚举 |
| `TrapType` | 陷阱类型枚举 |
| `ComponentBounds` | 组件几何边界 |

## 文件结构

| 文件 | 说明 |
|------|------|
| `manager.go` | 焦点管理器主实现，包含线性导航和陷阱管理 |
| `geometric.go` | 几何导航算法和组件边界计算 |
| `trap.go` | 焦点陷阱实现和堆栈管理 |
| `v3.go` | V3 版本的焦点系统接口（如存在） |

## 最佳实践

### 1. 动态更新焦点列表

当组件树变化（添加/删除组件）时，调用 `RefreshFocusables()`：

```go
// 添加新组件后
rootNode.ChildNodes = append(rootNode.ChildNodes, newNode)
fm.RefreshFocusables()
```

### 2. 模态对话框使用陷阱

```go
// 打开模态框
func ShowModal(fm *focus.Manager, modal *Modal) {
    trap := focus.NewFocusTrap("modal-"+modal.ID, focus.TrapModal, modal.RootNode)
    fm.PushFocusTrap(trap)
}

// 关闭模态框
func CloseModal(fm *focus.Manager, modal *Modal) {
    fm.PopFocusTrap()
}
```

### 3. 几何导航适用场景

- **推荐**：表单输入、网格布局、需要空间感知的界面
- **不推荐**：纯列表（使用线性导航更高效）

### 4. 跨平台键盘映射

```go
// 将按键映射为焦点导航
switch key {
case key.Tab:
    fm.FocusNext()
case key.ShiftTab:
    fm.FocusPrev()
case key.Up:
    fm.FocusUp()
case key.Down:
    fm.FocusDown()
// ...
}
```

## 常见问题

### Q: 为什么 FocusNext 没有效果？

可能原因：
1. 没有可聚焦组件（检查 `GetFocusableCount()`）
2. 焦点陷阱激活且陷阱内只有一个组件
3. 组件没有实现 `FocusableComponent` 接口

### Q: 几何导航跳到了意想不到的位置？

检查组件的 `MeasuredWidth` 和 `MeasuredHeight` 是否正确。几何导航依赖布局后的实际尺寸。

### Q: 焦点陷阱丢失了？

确保在关闭模态对话框时调用 `PopFocusTrap()`。如果可能抛出异常，使用 `defer` 保证清理：

```go
fm.PushFocusTrap(trap)
defer fm.PopFocusTrap()
// ...
```
