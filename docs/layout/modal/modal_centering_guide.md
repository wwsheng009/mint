# Modal 居中实现指南

## 目录

1. [核心设计](#核心设计)
2. [布局原理](#布局原理)
3. [配置方法](#配置方法)
4. [架构说明](#架构说明)
5. [使用示例](#使用示例)
6. [注意事项](#注意事项)
7. [最佳实践](#最佳实践)
8. [故障排除](#故障排除)
9. [底层实现详解](#底层实现详解)

---

## 核心设计

### 设计理念

**Modal 本质是"打破普通布局流"的节点**。

> ✅ 关键点：**Modal 仍在单一树中，但布局参考系提升到 Root**
> 
> 👉 本质：**逻辑父子 ≠ 布局父子**

### 核心机制

```
┌─────────────────────────────────────────┐
│  Root (Viewport: 80×45)                 │
│                                         │
│  ┌────────────┐                         │
│  │ VStack     │                         │
│  │  (父容器)   │                         │
│  │             │                         │
│  │  ┌─────────┐│    ✅ Modal Position   │
│  │  │ Modal   ││        Fixed +         │
│  │  │         ││        AnchorCenter    │
│  │  │居中!    ││                        │
│  │  └─────────┘│                        │
│  │    ↑        │                        │
│  │  计算逻辑   │                        │
│  │  (80-40)/2=20                     │
│  │  (45-12)/2≈17                     │
│  └────────────┘                         │
│                                         │
└─────────────────────────────────────────┘
```

**关键特性：**
1. Modal 逻辑上属于父容器（VStack）
2. Modal 布局上独立于父容器（使用 Root 作为参考系）
3. Modal 子元素仍然使用 Modal 自身尺寸约束（保持布局流）

---

## 布局原理

### 1️⃣ PositionFixed 定位机制

Modal 使用 `Position=fixed` 以 Root 为参考系进行定位：

```go
type PositionType int

const (
    PositionRelative PositionType = iota
    PositionAbsolute
    PositionFixed   // 🔑 Modal 使用这个
)
```

### 2️⃣ Anchor 锚点系统

```go
type AnchorType int

const (
    AnchorTopLeft     AnchorType = iota
    AnchorTop
    AnchorTopRight
    AnchorLeft
    AnchorCenter      // 🔑 Modal 居中使用这个
    AnchorRight
    AnchorBottomLeft
    AnchorBottom
    AnchorBottomRight
)
```

### 3️⃣ 居中计算公式

```
centeredModal.X  = (ViewportWidth  - ModalWidth)  / 2
centeredModal.Y  = (ViewportHeight - ModalHeight) / 2

例如：
  Viewport:  80 × 45
  Modal:     40 × 12
  计算结果：
    X = (80 - 40) / 2  = 20
    Y = (45 - 12) / 2 ≈ 16.5 → 16
```

### 4️⃣ 布局处理流程

```
┌─────────────────────────────────────────┐
│  1. VNode → Fiber → Props              │
│     centered: true →                    │
│     - Position: fixed                   │
│     - Anchor:   center                  │
│     - Layer:    Modal (2)               │
├─────────────────────────────────────────┤
│  2. 布局引擎 (Engine.Layout)           │
│     - Root 调用 Layout                  │
│     - 保存 viewportConstraints           │
│     - Modal 使用 Fixed 定位计算          │
│         • 使用 viewportConstraints       │
│         • 不是父容器 constraints         │
├─────────────────────────────────────────┤
│  3. Fixed 定位计算                      │
│     X = (80 - 40) / 2 = 20              │
│     Y = (45 - 12) / 2 ≈ 16              │
├─────────────────────────────────────────┤
│  4. 父容器处理                          │
│     ✅ Fixed 子节点位置不被覆盖         │
│     Flex/Grid/Absolute 布局不会重新设置 │
│     Modal 的 X, Y                       │
├─────────────────────────────────────────┤
│  5. 渲染输出                            │
│     Modal 渲染在 (20, 16)               │
│     在屏幕中央显示 ✅                   │
└─────────────────────────────────────────┘
```

---

## 配置方法

### Builder API 概述

Modal 提供两种居中配置方法：

#### Centered(bool) - Setter 方法

```go
modal.NewBuilder().
    Title("My Modal").
    Width(40).
    Height(12).
    Centered(true).  // 显式启用居中
    Open(true).
    Build()
```

**特点：**
- ✅ 可以动态控制（`true/false`）
- ✅ 提供完整控制权
- ✅ 适用于运行时决定是否居中

#### Center() - 快捷方法

```go
modal.NewBuilder().
    Title("My Modal").
    Center().          // 等价于 Centered(true)
    Width(40).
    Height(12).
    Open(true).
    Build()
```

**特点：**
- ✅ 语法简洁
- ✅ 提高可读性
- ✅ 适用于大多数"居中显示"场景

### 参数配置对比

| 方法 | 说明 | 底层调用 | 适用场景 |
|------|------|---------|---------|
| `Centered(true)` | **启用居中** | `SetCentered(true)` | 通用居中场景 |
| `Centered(false)` | **禁用居中** | `SetCentered(false)` | 自定义位置场景 |
| `Center()` | **启用居中（快捷）** | `Center()` - 内部调用 `SetCentered(true)` | 快速实现居中 |

### 完整配置示例

```go
// 标准 Modal（居中）
centeredModal := modal.NewBuilder().
    Title("Standard Modal").
    Content(newtext.New("This is a centered modal")).
    Width(40).
    Height(12).
    Center().
    Open(true).
    Build()

// 非居中 Modal（自定义位置）
customPosModal := modal.NewBuilder().
    Title("Custom Position Modal").
    Content(newtext.New("Fixed at top-right")).
    Width(30).
    Height(10).
    Centered(false).  // 禁用自动居中
    BuildVNode()

// 显式设置 Position 和 Anchor 属性
customPosModal.SetProps(rtui.Props{
    "position": "fixed",     // Fixed 定位
    "anchor":   "topright",  // 右上角锚点
})

// 结合使用
flexibleModal := modal.NewBuilder().
    Title("Toggle Center").
    Content(newtext.New("Center can be toggled")).
    Width(40).
    Height(12).
    Centered(true).  // 默认居中
    Open(true).
    Build()
```

---

## 架构说明

### 单树 + Layer 系统

```
┌─────────────────────────────────────┐
│         Single VNode Tree           │
│                                     │
│  Root                               │
│    ├── Page                         │
│    │   └── Content                  │
│    │                                 │
│    └── ModalLayer (Layer=2)          │
│        ├── Backdrop                 │
│        └── Modal (Position=fixed)   │
│            - Width: 40               │
│            - Height: 12              │
│            - Anchor: center          │
│            - AbsX: 20, AbsY: 16      │
│                                     │
└─────────────────────────────────────┘
```

### 关键组件

| 组件 | 职责 | 文件 |
|------|------|------|
| **Layout Engine** | 布局计算、Fixed 定位 | `runtime/layout/types.go` |
| **PositionProvider** | 提供 Position 和 Anchor 属性 | `runtime/layout/position.go` |
| **LayerManager** | Layer 管理（但不干预布局） | `runtime/layout/layer_manager.go` |
| **SyncPositioningProperties** | Props 到 Fiber 的同步 | `internal/reconciler/complete_work.go` |
| **Modal Builder** | Modal 配置 API | `ui/components/modal/vnode.go` |

### 关键接口

```go
// PositionProvider - 提供 Position 和 Anchor 属性
type PositionProvider interface {
    GetPositionType() PositionType
    GetAnchor()       Anchor
}

// Modal 居中配置（在 ModalVNode 中）
type ModalVNode struct {
    centered bool  // 控制 SetCentered()
    position PositionType
    anchor   Anchor
    layer    Layer
}

// SetCentered 设置居中标志
func (m *ModalVNode) SetCentered(centered bool) {
    m.centered = centered
    if centered {
        m.position = PositionFixed
        m.anchor   = AnchorCenter
    } else {
        m.position = PositionRelative
        m.anchor   = AnchorTopLeft
    }
}
```

---

## 使用示例

### 示例 1：基础居中 Modal

```go
package main

import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
    newtext "github.com/wwsheng009/mint/ui/components/text"
    uimodal "github.com/wwsheng009/mint/ui/components/modal"
)

func main() {
    // 创建居中的 Modal
    centeredModal := uimodal.NewBuilder().
        Title("Welcome").
        Content(newtext.New("This modal is centered!")).
        Width(40).
        Height(12).
        Center().          // Centered() 的快捷方式
        Open(true).
        BuildVNode()

    // 作为 VStack 的子元素使用
    root := rtui.VStack(
        newtext.New("Background content"),
        centeredModal,  // Modal 在 VStack 中，但不受其布局影响
    )
}
```

### 示例 2：动态控制居中

```go
type AppState struct {
    isCentered bool
}

func BuildApp(state *AppState) rtui.VNode {
    modalConfig := uimodal.NewBuilder().
        Title("Toggle Center").
        Content(newtext.New("Click to toggle")).
        Width(40).
        Height(12).
        Centered(state.isCentered).  // 动态控制居中
        Open(true)
    
    return rtui.VStack(
        rtui.Button("Toggle Center", func() {
            state.isCentered = !state.isCentered
        }),
        modalConfig.BuildVNode(),
    )
}
```

### 示例 3：不同锚点位置

```go
// 居中 Modal（默认）
centerModal := uimodal.NewBuilder().
    Title("Center").
    Width(40).
    Height(10).
    Center().               // center anchor
    BuildVNode()

// 顶部居中
topModal := uimodal.NewBuilder().
    Title("Top").
    Width(40).
    Height(8).
    Centered(false).
    BuildVNode()
topModal.SetProps(rtui.Props{
    "position": "fixed",
    "anchor":   "topcenter",  // topcenter anchor
})

// 右上角
topRightModal := uimodal.NewBuilder().
    Title("Top-Right").
    Width(30).
    Height(8).
    Centered(false).
    BuildVNode()
topRightModal.SetProps(rtui.Props{
    "position": "fixed",
    "anchor":   "topright",
})
```

### 示例 4：多个 Modal

```go
func App() rtui.VNode {
    return rtui.VStack(
        // 背景
        newtext.New("Background"),
        
        // 居中 Modal
        uimodal.NewBuilder().
            Title("Centered").
            Width(40).
            Height(12).
            Center().
            Open(true).
            BuildVNode(),
        
        // 左侧 Modal（非居中）
        uimodal.NewBuilder().
            Title("Left").
            Width(30).
            Height(10).
            Centered(false).
            BuildVNode().
            SetProps(rtui.Props{
                "position": "fixed",
                "anchor":   "topleft",
            }),
        
        // 右侧 Modal（非居中）
        uimodal.NewBuilder().
            Title("Right").
            Width(30).
            Height(10).
            Centered(false).
            BuildVNode().
            SetProps(rtui.Props{
                "position": "fixed",
                "anchor":   "topright",
            }),
    )
}
```

---

## 注意事项

### ⚠️ 1. Modal 作为子元素的位置

```go
// ❌ 错误：Modal 嵌套太深
root := rtui.VStack(
    rtui.HStack(
        rtui.VStack(
            uimodal.NewBuilder().Center().BuildVNode(),  // 太深！
        ),
    ),
)

// ✅ 正确：Modal 作为直接子元素
root := rtui.VStack(
    newtext.New("Content"),
    uimodal.NewBuilder().Center().BuildVNode(),  // 直接子元素
)
```

**原因：** Modal 虽然使用 Fixed 定位，但逻辑层级越深，Props 传递和 Layout 计算的延迟越大。

---

### ⚠️ 2. Props 覆盖顺序

```go
modal := uimodal.NewBuilder().
    Centered(true).              // 设置 centered=true
    BuildVNode()

// ❌ 错误：覆盖后居中失效
modal.SetProps(rtui.Props{
    "position": "relative",      // 覆盖了 fixed
    "anchor": "topleft",         // 覆盖了 center
})

// ✅ 正确：只覆盖需要的属性
modal.SetProps(rtui.Props{
    "title": "New Title",        // 只覆盖标题
})
```

**注意：** Props 会覆盖 Builder 配置，避免设置 `position`/`anchor` 除非明确需要非居中位置。

---

### ⚠️ 3. 尺寸约束

```go
// ⚠️ Modal 内的 Content 仍然使用 Modal 尺寸约束
modal := uimodal.NewBuilder().
    Title("My Modal").
    Width(40).         // Modal 宽度 = 40
    Height(12).        // Modal 高度 = 12
    Content(               // <─ Content 受这些尺寸约束
        newtext.New("Long text..."),  // 如果太长会被裁剪
    ).
    Center().
    Build()

// ✅ 如果需要使用完整 viewport，使用 Layer 系统
// 或创建全屏 Modal
fullscreenModal := uimodal.NewBuilder().
    Title("Fullscreen").
    Width(80).         // Viewport 宽度
    Height(45).        // Viewport 高度
    Center().
    Build()
```

---

### ⚠️ 4. 布局刷新

```go
// ⚠️ 如果 Modal 配置被修改，需要重新渲染
modal.SetCentered(false)  // 修改配置
// 需要调用 Invalidate 或 重新渲染
```

---

### ⚠️ 5. 测试配置

```go
// 测试 Modal 定位时，确保使用正确的 viewportConstraints
engine := layout.NewEngine()

// ✅ 正确：使用实际 viewport 尺寸
viewportConstraints := layout.Constraints{
    MaxWidth:  80,
    MaxHeight: 45,
}
result := engine.Layout(rootNode, viewportConstraints)

// ❌ 错误：使用被限制的 constraints（可能来自父容器）
restrictedConstraints := layout.Constraints{
    MaxWidth:  40,   // ❌ 限制在 Modal 宽度
    MaxHeight: 12,   // ❌ 限制在 Modal 高度
}
result := engine.Layout(modalNode, restrictedConstraints)  // 计算错误
```

---

## 最佳实践

### ✅ 推荐用法

#### 1. 使用 Center() 简化代码

```go
// ✅ 推荐：简洁直接
modal := uimodal.NewBuilder().
    Title("My Modal").
    Content(content).
    Width(40).
    Height(12).
    Center().
    BuildVNode()

// ⚠️ 可行但不简洁
modal := uimodal.NewBuilder().
    Title("My Modal").
    Content(content).
    Width(40).
    Height(12).
    Centered(true).
    BuildVNode()
```

#### 2. 明确配置尺寸

```go
// ✅ 推荐：明确尺寸
modal := uimodal.NewBuilder().
    Width(50).       // 明确宽度
    Height(15).      // 明确高度
    Center().
    Build()

// ⚠️ 不确定：依赖默认尺寸
modal := uimodal.NewBuilder().
    Center().        // 尺寸不明确
    Build()
```

#### 3. 验证居中结果

```go
// 测试 Modal 居中
func TestModalCentering(t *testing.T) {
    modal := uimodal.NewBuilder().
        Width(40).
        Height(12).
        Center().
        BuildVNode()
    
    engine := layout.NewEngine()
    constraints := layout.Constraints{MaxWidth: 80, MaxHeight: 45}
    result := engine.Layout(modal, constraints)
    
    // 预期居中位置
    expectedX := (80 - 40) / 2  // = 20
    expectedY := (45 - 12) / 2  // = 16
    
    if result.Root.AbsX != expectedX || result.Root.AbsY != expectedY {
        t.Errorf("Expected (%d,%d), got (%d,%d)", 
            expectedX, expectedY, result.Root.AbsX, result.Root.AbsY)
    }
}
```

#### 4. 尽量避免 Fragment 包裹 Modal

```go
// ✅ 推荐：Modal 在实际布局容器中
root := rtui.VStack(
    newtext.New("Content"),
    uimodal.NewBuilder().Center().BuildVNode(),
)

// ⚠️ 谨慎：Fragment 可能导致 constraints 传递问题
root := rtui.Fragment(
    // ...其他内容
    uimodal.NewBuilder().Center().BuildVNode(),
)
```

---

## 故障排除

### 问题 1：Modal 未居中

**症状：**
```
Modal 显示在错误位置，如 (0, 13)
```

**原因分析：**

1. **Props 未正确设置**
   ```go
   // ❌ 问题：Centered() 未应用
   modal := modal.NewBuilder().
       Width(40).
       // 丢失了 .Center()
       BuildVNode()
   
   // ✅ 解决：添加 Center()
   modal := modal.NewBuilder().
       Width(40).
       Center().
       BuildVNode()
   ```

2. **SyncPositioningProperties 未调用**
   - 检查是否有 Fiber 树构建
   - 确认 `SyncPositioningProperties()` 被调用

3. **被父容器覆盖位置**
   - 检查 `types.go` 中 Fixed 子节点是否被正确处理
   - 确认父容器没有覆盖 `subBox.X`/`subBox.Y`

**诊断步骤：**

```go
// 1. 检查 Props
fmt.Printf("Modal Props: centered=%v, position=%v, anchor=%v\n",
    modal.Props["centered"], modal.Props["position"], modal.Props["anchor"])

// 2. 检查 Fiber 节点
fiber := reconciler.BuildFiberTree(root)
modalFiber := findModalFiber(fiber)
fmt.Printf("Modal Fiber: Position=%v, Anchor=%v\n", 
    modalFiber.Position, modalFiber.Anchor)

// 3. 检查布局结果
result := engine.Layout(root, constraints)
fmt.Printf("Layout: AbsX=%d, AbsY=%d\n", 
    result.Root.AbsX, result.Root.AbsY)
```

---

### 问题 2：Modal 尺寸被约束限制

**症状：**
```
Modal 内容被裁剪或尺寸不正确
```

**原因：**
Modal 内的子元素使用 Modal 尺寸约束。

**解决：**

```go
// ✅ 增大 Modal 尺寸
modal := uimodal.NewBuilder().
    Width(60).          // 增大宽度
    Height(18).         // 增大高度
    Center().
    BuildVNode()

// 或：使用滚动容器（如果实现）
modal.WithContent(
    rtui.Scrollable(newtext.New(longContent)),
)
```

---

### 问题 3：多个 Modal 位置重叠

**症状：**
```
多个 Fixed Modal 显示在同一位置
```

**原因：**
缺少 Z-index 管理。

**解决：**
确保 Layer 系统（Modal Layer = 2）已正确应用：
```go
// ModalVNode 应该有 Layer = LayerModal (2)
type ModalVNode struct {
    layer Layer  // 应该是 2 (LayerModal)
}
```

---

### 问题 4：布局结果不稳定

**症状：**
```
Modal 位置在不同渲染中不一致
```

**原因：**
- 脏标记问题导致 Fixed 定位未重新计算
- 父容器布局变化影响 Modal

**解决：**

```go
// 强制刷新布局
engine.InvalidateNode(modalID)

// 确认 Modal 的 Fixed 定位在 Clean 路径也被执行
// (已在 types.go 的 !IsLayoutDirty() 分支中添加 Fixed 计算)
```

---

### 问题 5：测试环境与实际环境不一致

**症状：**
```
测试中 Modal 居中，实际运行不居中
```

**原因：**
viewportConstraints 不一致。

**解决：**

```go
// 测试中正确模拟 viewportConstraints
func TestModal(t *testing.T) {
    viewportConstraints := layout.Constraints{
        MaxWidth:  80,   // ✅ 使用实际 viewport 尺寸
        MaxHeight: 45,
    }
    
    // 不使用限制的 constraints
    engine.SetViewportConstraints(viewportConstraints)  // 新增 API
    result := engine.Layout(rootNode, viewportConstraints)
}
```

---

## 底层实现详解

### 关键实现文件

| 文件 | 功能 | 关键函数 |
|------|------|---------|
| `runtime/layout/types.go` | 布局引擎 | `layoutNodeWithDepth()` |
| `runtime/layout/position.go` | Position/Anchor | `PositionProvider` |
| `runtime/layout/layer.go` | Layer 系统 | `LayerManager` |
| `internal/reconciler/complete_work.go` | Props 同步 | `SyncPositioningProperties()` |
| `ui/components/modal/vnode.go` | Modal 实现 | `Centered()`, `Center()` |

### Fixed 定位计算流程

```go
// 在 layoutNodeWithDepth() 中：

// 1. 获取 Position 和 Anchor
position := PositionRelative
anchor := AnchorTopLeft
if posProvider, ok := node.(PositionProvider); ok {
    position = posProvider.GetPositionType()
    anchor = posProvider.GetAnchor()
}

// 2. Fixed 定位：使用 viewportConstraints
if position == PositionFixed && width > 0 && height > 0 {
    // 使用保存的 viewport 约束
    rootW := e.viewportConstraints.MaxWidth
    rootH := e.viewportConstraints.MaxHeight
    
    // 根据 Anchor 计算固定定位坐标
    switch anchor {
    case AnchorCenter:
        x = (rootW - width) / 2
        y = (rootH - height) / 2
    // ... 其他锚点
    }
}

// 3. Clean 路径也执行 Fixed 计算（避免跳过）
if !e.dirty.IsLayoutDirty(node.ID()) {
    // 获取 PositionProvider
    // Fixed 定位计算（同上）
}
```

### 父容器位置保护机制

```go
// 在子节点布局后：

// ❌ 错误：盲目覆盖位置
// subBox.X = childX
// subBox.Y = childY

// ✅ 正确：检查子节点 Position
childPosition := PositionRelative
if posProvider, ok := child.(PositionProvider); ok {
    childPosition = posProvider.GetPositionType()
}

// 只覆盖非 Fixed 子节点的位置
if childPosition != PositionFixed {
    subBox.X = childX
    subBox.Y = childY
}
// Fixed 子节点：保持 Fixed 定位计算的位置
```

### Props 同步机制

```go
// SyncPositioningProperties: Props → Fiber
func SyncPositioningProperties(fiber *Fiber) {
    // 从 Props 读取.centered
    if centered, ok := fiber.Props["centered"].(bool); ok {
        if centered {
            fiber.Position = PositionFixed
            fiber.Anchor = AnchorCenter
        }
    }
    
    // 显式设置
    if position, ok := fiber.Props["position"].(string); ok {
        fiber.Position = PositionTypeFromString(position)
    }
    if anchor, ok := fiber.Props["anchor"].(string); ok {
        fiber.Anchor = AnchorTypeFromString(anchor)
    }
}
```

---

## 调试指南

### 启用调试输出

```go
// 设置调试环境变量
os.Setenv("MINT_DEBUG_LAYOUT", "true")
os.Setenv("MINT_FIBER_FIRST", "true")

// 或使用调试 API
engine.SetDebug(true)

// 显示布局结果
result := engine.Layout(rootNode, constraints)
printLayoutBoxes(result.Root)

// 查看 Modal 布局信息
for _, box := range result.Boxes {
    if box.Layer == layout.LayerModal {
        fmt.Printf("Modal: AbsX=%d, AbsY=%d, %dx%d\n",
            box.AbsX, box.AbsY, box.Width, box.Height)
    }
}
```

### 工具函数

```go
// 检查 Modal 是否居中
func IsModalCentered(box *layout.LayoutBox, viewportConstraints layout.Constraints) bool {
    expectedX := (viewportConstraints.MaxWidth - box.Width) / 2
    expectedY := (viewportConstraints.MaxHeight - box.Height) / 2
    
    return box.AbsX == expectedX && box.AbsY == expectedY
}

// 打印 Modal 布局信息
func PrintModalLayout(box *layout.LayoutBox) {
    fmt.Printf("\nModal Layout:\n")
    fmt.Printf("  Position:  X=%d, Y=%d\n", box.X, box.Y)
    fmt.Printf("  Absolute:   AbsX=%d, AbsY=%d\n", box.AbsX, box.AbsY)
    fmt.Printf("  Size:     Width=%d, Height=%d\n", box.Width, box.Height)
    fmt.Printf("  Layer:    Layer=%s\n", box.Layer)
    fmt.Printf("  ShouldCenter: %v\n", box.ShouldCenter)
}
```

---

## 总结

### 关键要点

1. ✅ **Position=fixed + Anchor=center** - 实现居中
2. ✅ **使用 Root 作为参考系** - 脱离父布局流
3. ✅ **子元素使用 Modal 约束** - 保持约束传递
4. ✅ **父容器不覆盖 Fixed 位置** - 保护 Fixed 计算

### 快速检查清单

- [ ] 使用 `.Center()` 或 `.Centered(true)`
- [ ] 明确设置 Width 和 Height
- [ ] Modal 作为实际布局容器的子元素
- [ ] 避免覆盖 `position`/`anchor` Props
- [ ] 验证布局结果（AbsX/AbsY）
- [ ] 测试不同 viewports

### 参考配置

```go
// 推荐的居中 Modal 配置
modal := uimodal.NewBuilder().
    Title("Title").
    Content(newtext.New("Content")).
    Width(50).            // 适当的宽度
    Height(15).           // 适当的高度
    Center().             // Centered() 快捷方式
    Open(true).           // 打开 Modal
    BuildVNode()
```

---

## 附录

### A. 布局计算示例

```
场景：Viewport = 100×60, Modal = 60×20

居中计算：
--------------------------------
AbsX = (100 - 60) / 2 = 20
AbsY = (60  - 20) / 2 = 20

结果：Modal 显示在 (20, 20)
```

### B. 不同锚点位置

| Anchor | 说明 | 公式 |
|--------|------|------|
| `topleft` | 左上角 | ``(0, 0)`` |
| `top` | 顶部居中 | ``((rootW-w)/2, 0)`` |
| `topright` | 右上角 | ``(rootW-w, 0)`` |
| `left` | 左侧居中 | ``(0, (rootH-h)/2)`` |
| `center` | **中央** | **((rootW-w)/2, (rootH-h)/2)** |
| `right` | 右侧居中 | ``(rootW-w, (rootH-h)/2)`` |
| `bottomleft` | 左下角 | ``(0, rootH-h)`` |
| `bottom` | 底部居中 | ``((rootW-w)/2, rootH-h)`` |
| `bottomright` | 右下角 | ``(rootW-w, rootH-h)`` |

### C. LayoutEngine API

```go
type Engine struct {
    viewportConstraints Constraints  // 保存 Root 约束
}

func (e *Engine) Layout(root Node, constraints Constraints) *LayoutResult
func (e *Engine) LayoutIncremental(root Node, constraints Constraints) *LayoutResult
func (e *Engine) InvalidateNode(id string)
```

**注意：** `viewportConstraints` 在 `Layout()` 开始时保存，用于 Fixed 定位节点计算。

---

*文档版本：1.0*
*最后更新：2025-12-xx*
*适用版本：Mint Framework v0.x*
