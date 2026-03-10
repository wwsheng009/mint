# VNode 属性传输到 LayoutBox 完整指南

## 目录
- [1. 概述](#1-概述)
- [2. 属性传输流程](#2-属性传输流程)
- [3. 各层详细说明](#3-各层详细说明)
- [4. 常见错误与解决方案](#4-常见错误与解决方案)
- [5. 调试指南](#5-调试指南)
- [6. 新增属性步骤清单](#6-新增属性步骤清单)
- [7. 实战案例：margin 属性](#7-实战案例margin-属性)

---

## 1. 概述

Mint UI 框架采用分层的属性传输架构，从 UI 组件层的 Builder API 到布局引擎的 LayoutBox，需要经过多个层次的转换。

### 架构层次

```
┌─────────────────────────────────────────────────────────────┐
│  VNode 层        # 用户 API，设置属性                          │
│  ↘                                                              │
│    Builder API → VNode 结构体                                  │
│      • 属性存储在 VNode 字段中                                  │
│      • 通过 Props() 方法导出所有属性                            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Fiber 层        # 运行时状态，协调更新                         │
│  ↘                                                              │
│    CreateFiber() → Fiber 结构体                                │
│      • 从 VNode.Props() 提取属性到 Fiber 字段                   │
│      • Fiber 字段用于布局计算和状态管理                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  适配器层        # Fiber 到 layout.Node 的桥接                  │
│  ↘                                                              │
│    FiberToNodeAdapter → layout.Node 接口                      │
│      • 从 Fiber 字段转换为布局引擎接口                         │
│      • 实现布局引擎所需的接口方法（如 GetMargin()）              │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  布局引擎层      # 布局计算，生成 LayoutBox                   │
│  ↘                                                              │
│    layout.Engine → LayoutBox 树                                │
│      • 使用 fiber_adapter 接口读取属性                          │
│      • 进行布局计算，考虑所有布局属性                           │
└─────────────────────────────────────────────────────────────┘
```

### 核心设计原则

1. **单向数据流**：属性从 UI 组件流向布局引擎，不会反向传播
2. **接口隔离**：各层通过接口通信，降低耦合
3. **延迟计算**：布局计算在需要时才执行，避免不必要的计算
4. **类型安全**：使用结构体和接口确保类型正确性

---

## 2. 属性传输流程

### 2.1 完整流程图

```
用户代码
  ↓
Component API (Builder)
  ↓
VNode 结构体 (ui/components/*/vnode.go)
  ├─ 属性字段 (如 padding, margin)
  └─ Props() 方法 → rtui.Props (Map)
  ↓
CreateFiber() (runtime/ui/fiber_util.go)
  ├─ vnode.Props() 读取
  ├─ 提取到 Fiber 字段 (如 LayoutPadding, LayoutMargin)
  └─ Props → Fiber 字段映射
  ↓
Fiber 结构体 (runtime/ui/fiber.go)
  ├─ Layout 属性字段
  └─ Props (原始 VNode Props，用于兼容)
  ↓
FiberToNodeAdapter (internal/render/fiber_adapter.go)
  ├─ GetPadding() → 从 Fiber.LayoutPadding 读取
  ├─ GetMargin() → 从 Fiber.LayoutMargin 读取
  └─ 实现 layout.Node 接口方法
  ↓
layout.Engine (runtime/layout/types.go)
  ├─ 从 layout.Node 读取属性
  ├─ 进行布局计算
  └─ 生成 LayoutBox 树
  ↓
LayoutBox (带位置和尺寸信息的布局树)
```

### 2.2 关键数据结构

#### VNode 层
```go
// ui/components/button/vnode.go
type VNode struct {
    *rtui.ElementVNode
    rtui.BoxModelMixin  // 嵌入 mixin 提供 padding, margin, textAlign

    label      string
    variant    Variant
    // ... 其他字段
}

// BoxModelMixin 提供默认实现
type BoxModelMixin struct {
    padding   [4]int   // top, right, bottom, left
    margin    [4]int   // top, right, bottom, left
    textAlign Align
}
```

#### Fiber 层
```go
// runtime/ui/fiber.go
type Fiber struct {
    // ... 其他字段

    // Layout Properties (extracted from VNode)
    LayoutDirection  Direction
    LayoutAlign      Align
    LayoutCrossAlign Align
    LayoutGap        int
    LayoutPadding    [4]int   // top, right, bottom, left
    LayoutMargin     [4]int   // top, right, bottom, left ← 新增
    LayoutFlex       int

    // 原始 Props（用于兼容）
    Props Props
}
```

#### 适配器层
```go
// internal/render/fiber_adapter.go
type FiberToNodeAdapter struct {
    fiber *reconciler.Fiber
}

// 实现 layout.Marginal 接口
func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
    // ✅ 优先读取 Fiber 字段
    if a.fiber.LayoutMargin[0] != 0 || a.fiber.LayoutMargin[1] != 0 {
        return layout.Margin{
            Top:    a.fiber.LayoutMargin[0],
            Right:  a.fiber.LayoutMargin[1],
            Bottom: a.fiber.LayoutMargin[2],
            Left:   a.fiber.LayoutMargin[3],
        }
    }
    // 向后兼容：从 Props 读取
    if a.fiber.Props != nil {
        if m, ok := a.fiber.Props["margin"].([4]int); ok {
            return layout.Margin{...}
        }
    }
    return layout.Margin{}
}
```

---

## 3. 各层详细说明

### 3.1 VNode 层

#### 职责
- 提供 Builder API 供用户设置属性
- 存储属性在结构体字段中
- 通过 `Props()` 方法导出所有属性

#### 关键点
1. **属性存储方式**：
   - 推荐使用 `BoxModelMixin` 嵌入来获得 box model 支持
   - 直接结构体字段存储组件特定属性

2. **Props() 方法**：
   - 必须包含所有可用于布局的属性
   - 包括直接的字段和嵌入 mixin 的属性
   - 示例：
   ```go
   func (b *VNode) Props() rtui.Props {
       return rtui.Props{
           "key":       b.key,
           "label":     b.label,
           "padding":   b.padding,       // 直接字段
           "margin":    b.Margin(),      // Mixin 方法
           "textAlign": b.Margin(),      // Mixin 方法
           "flex":      b.flex,
       }
   }
   ```

#### 常见错误

**错误 1：Props() 没有包含新属性**
```go
// ❌ 错误：Props() 缺少 margin
func (b *VNode) Props() rtui.Props {
    return rtui.Props{
        "padding":  b.padding,
        "flex":     b.flex,
        // 缺少 "margin"
    }
}

// ✅ 正确：包含 margin
func (b *VNode) Props() rtui.Props {
    return rtui.Props{
        "padding":  b.padding,
        "margin":   b.Margin(),  // 添加这行
        "flex":     b.flex,
    }
}
```

**错误 2：使用错误的方法名**
```go
// ❌ 错误：混淆 Margin() 和 MarginH()
func (b *VNode) Props() rtui.Props {
    return rtui.Props{
        "marginH": b.Margin(),  // 错误的 key
    }
}

// ✅ 正确：使用标准的 key
func (b *VNode) Props() rtui.Props {
    return rtui.Props{
        "margin": b.Margin(),  // 使用 "margin"
    }
}
```

---

### 3.2 Fiber 层

#### 职责
- 从 VNode.Props() 提取属性到 Fiber 字段
- 在协调过程中保持属性不变（直到新的 VNode）
- 将属性传递给布局引擎

#### 关键点
1. **CreateFiber() 中的属性提取**：
   ```go
   // runtime/ui/fiber_util.go
   func CreateFiber(vnode VNode) *Fiber {
       // ... 初始化变量
       var layoutPadding [4]int
       var layoutMargin [4]int  // ✅ 必须声明变量
       var layoutFlex int

       // 从 Props 提取
       if props != nil {
           if p, ok := props["padding"].([4]int); ok {
               layoutPadding = p
           }
           if m, ok := props["margin"].([4]int); ok {
               layoutMargin = m  // ✅ 必须提取
           }
           if f, ok := props["flex"].(int); ok {
               layoutFlex = f
           }
       }

       // 设置到 Fiber
       return &Fiber{
           LayoutPadding: layoutPadding,
           LayoutMargin:  layoutMargin,  // ✅ 必须赋值
           LayoutFlex:    layoutFlex,
           Props:         props,  // 保留原始 Props
       }
   }
   ```

2. **Fiber 字段命名约定**：
   - Layout + 属性名（如 LayoutPadding, LayoutMargin）
   - 统一使用 `[4]int` 表示四个方向（top, right, bottom, left）
   - 布局属性单独存储，不和 Props 混用

#### 常见错误

**错误 1：忘记声明变量**
```go
// ❌ 错误：没有声明 layoutMargin 变量
func CreateFiber(vnode VNode) *Fiber {
    var layoutPadding [4]int
    // 缺少: var layoutMargin [4]int

    if props != nil {
        if m, ok := props["margin"].([4]int); ok {
            layoutMargin = m  // 编译错误：undefined
        }
    }
}

// ✅ 正确：声明变量
func CreateFiber(vnode VNode) *Fiber {
    var layoutPadding [4]int
    var layoutMargin [4]int  // ✅ 声明变量

    if props != nil {
        if m, ok := props["margin"].([4]int); ok {
            layoutMargin = m
        }
    }
}
```

**错误 2：忘记赋值到 Fiber**
```go
// ❌ 错误：提取了但没有赋值到 Fiber
if props != nil {
    if m, ok := props["margin"].([4]int); ok {
        layoutMargin = m  // 提取了
    }
    // 缺少: LayoutMargin: layoutMargin
}
return &Fiber{
    LayoutPadding: layoutPadding,
    // 缺少: LayoutMargin: layoutMargin
}

// ✅ 正确：赋值到 Fiber
if props != nil {
    if m, ok := props["margin"].([4]int); ok {
        layoutMargin = m  // 提取了
    }
}
return &Fiber{
    LayoutPadding: layoutPadding,
    LayoutMargin:  layoutMargin,  // ✅ 赋值到 Fiber
}
```

**错误 3：类型断言错误**
```go
// ❌ 错误：错误的类型断言
if m, ok := props["margin"].(int); ok {  // margin 是 [4]int 不是 int
    layoutMargin = [4]int{m, m, m, m}
}

// ✅ 正确：正确的类型断言
if m, ok := props["margin"].([4]int); ok {
    layoutMargin = m
}
```

---

### 3.3 适配器层 (FiberToNodeAdapter)

#### 职责
- 将 Fiber 字段转换为 layout.Node 接口
- 实现布局引擎所需的接口方法
- 提供属性访问方法

#### 关键点
1. **接口实现**：
   - 实现 `layout.Marginal` 接口用于 margin
   - 实现 `layout.Bordered` 接口用于边框
   - 实现 `layout.FlexStyleProvider` 接口用于 flex 样式

2. **读取顺序**：
   - 优先读取 Fiber 字段（新的方式）
   - 向后兼容 Props 读取（旧的方式）
   -示例：
   ```go
   func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
       if a.fiber == nil {
           return layout.Margin{}
       }

       // ✅ 优先读取 Fiber 字段
       if a.fiber.LayoutMargin[0] != 0 || a.fiber.LayoutMargin[1] != 0 ||
          a.fiber.LayoutMargin[2] != 0 || a.fiber.LayoutMargin[3] != 0 {
           return layout.Margin{
               Top:    a.fiber.LayoutMargin[0],
               Right:  a.fiber.LayoutMargin[1],
               Bottom: a.fiber.LayoutMargin[2],
               Left:   a.fiber.LayoutMargin[3],
           }
       }

       // 向后兼容：从 Props 读取
       if a.fiber.Props != nil {
           if m, ok := a.fiber.Props["margin"].([4]int); ok {
               return layout.Margin{
                   Top:    m[0],
                   Right:  m[1],
                   Bottom: m[2],
                   Left:   m[3],
               }
           }
       }

       return layout.Margin{}
   }
   ```

3. **零值检查**：
   - 对于可选属性，检查是否为零值
   - 只有非零值才返回有效属性

#### 常见错误

**错误 1：忘记实现接口**
```go
// ❌ 错误：没有实现 GetMargin() 方法
// 导致布局引擎无法读取 margin
type FiberToNodeAdapter struct {
    fiber *reconciler.Fiber
}

// 缺少: func (a *FiberToNodeAdapter) GetMargin() layout.Margin

// ✅ 正确：实现接口
func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
    // ... 实现
}
```

**错误 2：返回错误的类型**
```go
// ❌ 错误：返回 [4]int 而不是 layout.Margin
func (a *FiberToNodeAdapter) GetMargin() [4]int {
    return a.fiber.LayoutMargin
}

// ✅ 正确：返回 layout.Margin
func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
    return layout.Margin{
        Top:    a.fiber.LayoutMargin[0],
        Right:  a.fiber.LayoutMargin[1],
        Bottom: a.fiber.LayoutMargin[2],
        Left:   a.fiber.LayoutMargin[3],
    }
}
```

**错误 3：索引错误**
```go
// ❌ 错误：错误的索引顺序
func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
    return layout.Margin{
        Top:    a.fiber.LayoutMargin[1],  // 错误：应该是 [0]
        Right:  a.fiber.LayoutMargin[0],  // 错误：应该是 [1]
        Bottom: a.fiber.LayoutMargin[3],  // 错误：应该是 [2]
        Left:   a.fiber.LayoutMargin[2],  // 错误：应该是 [3]
    }
}

// ✅ 正确：正确的索引顺序
func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
    return layout.Margin{
        Top:    a.fiber.LayoutMargin[0],  // top
        Right:  a.fiber.LayoutMargin[1],  // right
        Bottom: a.fiber.LayoutMargin[2],  // bottom
        Left:   a.fiber.LayoutMargin[3],  // left
    }
}
```

---

### 3.4 布局引擎层 (layout.Engine)

#### 职责
- 从 layout.Node 接口读取属性
- 进行布局计算，考虑所有布局属性
- 生成 LayoutBox 树

#### 关键点
1. **在 layoutNodeWithDepth 中使用属性**：
   ```go
   // runtime/layout/types.go
   func (e *Engine) layoutNodeWithDepth(
       node layout.Node,
       constraints Constraints,
       x, y int,
       depth int,
       visited map[uint64]bool,
   ) *LayoutBox {
       // ... 计算尺寸

       // ✅ 检查节点是否实现了 Marginal 接口
       marginTop, marginBottom, marginLeft, marginRight := 0, 0, 0, 0
       if marginal, ok := child.(Marginal); ok {
           m := marginal.GetMargin()
           marginTop = m.Top
           marginBottom = m.Bottom
           marginLeft = m.Left
           marginRight = m.Right
       }

       // ✅ 应用 margin 偏移
       actualChildX := childX + marginLeft
       actualChildY := childY + marginTop

       // ✅ 调整约束，考虑 margin
       adjustedConstraints := childConstraints
       if childConstraints.MaxWidth > 0 {
           adjustedConstraints.MaxWidth = max(0, childConstraints.MaxWidth-marginLeft-marginRight)
       }

       // 递归布局子节点
       childBox := e.layoutNodeWithDepth(
           child,
           adjustedConstraints,
           actualChildX,
           actualChildY,
           depth+1,
           visited,
       )
   }
   ```

2. **Flex 布局中的 margin 处理**：
   - FlexLayout 不考虑 margin，需要在 types.go 中手动处理
   - 累积主轴方向的 margin 偏移
   - 跨轴方向的 margin 直接应用到位置

#### 常见错误

**错误 1：忘记检查接口**
```go
// ❌ 错误：没有检查节点是否实现了 Marginal 接口
marginTop := 0  // 永远是 0，导致 margin 无效

// ✅ 正确：检查接口
marginTop, marginBottom, marginLeft, marginRight := 0, 0, 0, 0
if marginal, ok := child.(Marginal); ok {
    m := marginal.GetMargin()
    marginTop = m.Top
    marginBottom = m.Bottom
    marginLeft = m.Left
    marginRight = m.Right
}
```

**错误 2：忘记偏移位置**
```go
// ❌ 错误：没有应用 margin 偏移
subBox := e.layoutNodeWithDepth(child, childConstraints, childX, childY, depth+1, visited)

// ✅ 正确：应用 margin 偏移
actualChildX := childX + marginLeft
actualChildY := childY + marginTop
subBox := e.layoutNodeWithDepth(
    child,
    childConstraints,
    actualChildX,
    actualChildY,
    depth+1,
    visited,
)
```

**错误 3：Flex 布局没有累积 margin**
```go
// ❌ 错误：Flex 布局没有累积 margin
childY := y + childBox.Y + borderOffsetY + marginTop
// 这样会导致所有子节点在相同 margin 上叠加

// ✅ 正确：累积 margin
mainAxisMarginOffset := 0
for i, childBox := range childBoxes {
    // ... 获取 margin

    childY := y + childBox.Y + borderOffsetY + mainAxisMarginOffset + marginTop
    mainAxisMarginOffset += marginTop + marginBottom  // 累积
}
```

---

## 4. 常见错误与解决方案

### 4.1 属性没有生效

#### 症状
- 设置了属性但没有效果
- 按钮之间没有间距等

#### 排查步骤
1. **检查 VNode.Props() 是否包含属性**
   ```go
   fmt.Printf("VNode Props: %+v\n", vnode.Props())
   ```

2. **检查 Fiber 字段是否正确提取**
   ```go
   fmt.Printf("Fiber LayoutMargin: %+v\n", fiber.LayoutMargin)
   ```

3. **检查适配器接口方法是否实现**
   ```go
   if marginal, ok := adapter.(layout.Marginal); ok {
       fmt.Printf("Margin: %+v\n", marginal.GetMargin())
   }
   ```

4. **检查布局引擎是否使用属性**
   - 在 types.go 中搜索属性名
   - 确认是否在布局计算中使用

#### 常见原因
1. VNode.Props() 没有包含属性
2. Fiber 字段没有正确提取
3. 适配器没有实现接口方法
4. 布局引擎没有使用属性

---

### 4.2 编译错误

#### undefined 错误
```
undefined: layoutMargin
```

**解决方案**：
```go
// 确保声明了变量
var layoutMargin [4]int
```

#### cannot use 错误
```
cannot use layoutMargin (type [4]int) as type [4]int
```

**解决方案**：
检查是否使用了错误的类型或数组长度。

---

### 4.3 运行时 panic

#### interface conversion 错误
```
panic: interface conversion: interface {} is string, not [4]int
```

**解决方案**：
```go
// ✅ 正确的类型断言
if m, ok := props["margin"].([4]int); ok {
    layoutMargin = m
}
```

#### nil pointer 错误
```
panic: runtime error: invalid memory address or nil pointer dereference
```

**解决方案**：
```go
// ✅ 添加 nil 检查
if a.fiber == nil || a.fiber.Props == nil {
    return layout.Margin{}
}
```

---

## 5. 调试指南

### 5.1 添加调试日志

#### 在 CreateFiber 中添加日志
```go
// runtime/ui/fiber_util.go
func CreateFiber(vnode VNode) *Fiber {
    // ... 提取属性
    if m, ok := props["margin"].([4]int); ok {
        layoutMargin = m
        fmt.Printf("[CreateFiber] Extracted margin: %+v\n", m)  // 调试日志
    }

    return &Fiber{
        LayoutMargin: layoutMargin,
    }
}
```

#### 在适配器中添加日志
```go
// internal/render/fiber_adapter.go
func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
    if a.fiber == nil {
        return layout.Margin{}
    }

    margin := layout.Margin{
        Top:    a.fiber.LayoutMargin[0],
        Right:  a.fiber.LayoutMargin[1],
        Bottom: a.fiber.LayoutMargin[2],
        Left:   a.fiber.LayoutMargin[3],
    }

    fmt.Printf("[GetMargin] FiberID=%d, Margin=%+v\n", a.fiber.NodeID, margin)
    return margin
}
```

#### 在布局引擎中添加日志
```go
// runtime/layout/types.go
func (e *Engine) layoutNodeWithDepth(...) *LayoutBox {
    // ... 获取 margin
    if marginal, ok := child.(Marginal); ok {
        m := marginal.GetMargin()
        fmt.Printf("[LayoutNode] ChildID=%s, Margin=%+v\n", child.ID(), m)
    }
}
```

### 5.2 使用调试工具

#### 查看 Fiber 树
```go
// 使用 framework 提供的工具
import "github.com/wwsheng009/mint/devtools"

// 打印 Fiber 树
devtools.PrintFiberTree(fiber)
```

#### 查看 LayoutBox 树
```go
// 使用提供的 API
boxes := node.GetLayoutBoxes()
for _, box := range boxes {
    fmt.Printf("Box: ID=%s, X=%d, Y=%d, W=%d, H=%d\n",
        box.ID, box.X, box.Y, box.Width, box.Height)
}
```

### 5.3 单步调试

#### 使用 delve 调试器
```bash
# 安装 delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug ./examples/elegant_api_demo/main.go

# 设置断点
(dlv) break CreateFiber
(dlv) break FiberToNodeAdapter.GetMargin
(dlv) continue

# 查看变量
(dlv) print vnode.Props()
(dlv) print layoutMargin
(dlv) print fiber.LayoutMargin
```

---

## 6. 新增属性步骤清单

### 步骤 1：为 VNode 添加属性字段
- [ ] 在 VNode 结构体中添加字段
- [ ] 如果适用，使用 BoxModelMixin
- [ ] 在 Props() 方法中添加属性

### 步骤 2：在 CreateFiber 中提取属性
- [ ] 声明对应的变量
- [ ] 从 Props 中提取属性
- [ ] 赋值到 Fiber 字段

### 步骤 3：在 Fiber 结构体中添加字段
- [ ] 添加到 Fiber 结构体
- [ ] 遵循命名约定（Layout + 属性名）

### 步骤 4：在适配器中实现接口方法
- [ ] 实现对应的 layout.* 接口
- [ ] 优先读取 Fiber 字段
- [ ] 向后兼容 Props 读取

### 步骤 5：在布局引擎中使用属性
- [ ] 在 layoutNodeWithDepth 中读取属性
- [ ] 应用属性到布局计算
- [ ] 调整约束和位置

### 步骤 6：特殊处理（如需要）
- [ ] Flex 布局需要特殊处理
- [ ] Grid 布局需要特殊处理
- [ ] Wrap 布局需要特殊处理

### 步骤 7：测试
- [ ] 编写单元测试
- [ ] 编写集成测试
- [ ] 创建 demo 验证

---

## 7. 实战案例：margin 属性

### 7.1 问题陈述
用户使用 `ui.NewButtonBuilder("Btn").MarginV(1, 1).Build()` 设置 margin，但按钮之间没有显示间距。

### 7.2 排查过程

#### 第一步：确认问题
```go
// 运行示例
go run ./examples/elegant_api_demo/main.go
// 输出：按钮之间没有间距
```

#### 第二步：检查 VNode.Props()
```go
// ui/components/button/vnode.go
func (b *VNode) Props() rtui.Props {
    return rtui.Props{
        "padding":  b.padding,
        "flex":     b.flex,
        // ❌ 缺少 "margin"
    }
}
```

**发现问题 1**：Props() 没有包含 margin。

#### 第三步：检查 Fiber 字段
```go
// 运行调试程序
go run ./examples/elegant_api_demo/debug_margin.go
// 输出：margin 没有被提取到 Fiber
```

**确认问题 1**：Fiber.LayoutMargin 是空的。

#### 第四步：检查适配器
```go
// internal/render/fiber_adapter.go
func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
    // 优先从 Props 读取 ← 错误的顺序
    if a.fiber.Props != nil {
        if m, ok := a.fiber.Props["margin"].([4]int); ok {
            return layout.Margin{...}
        }
    }
    // 然后从 Fiber 字段读取 ← 错误的顺序
    if a.fiber.LayoutMargin[0] != 0 {
        return layout.Margin{...}
    }
    return layout.Margin{}
}
```

**发现问题 2**：读取顺序错误，应该优先读取 Fiber 字段。

#### 第五步：检查布局引擎
```go
// runtime/layout/types.go
// Flex 布局没有累积 margin
for i, childBox := range childBoxes {
    childY := y + childBox.Y + borderOffsetY + marginTop
    // ❌ 没有累积 margin
}
```

**发现问题 3**：Flex 布局没有累积主轴方向的 margin。

### 7.3 修复过程

#### 修复 1：VNode.Props() 添加 margin
```go
// ui/components/button/vnode.go
func (b *VNode) Props() rtui.Props {
    return rtui.Props{
        "padding":  b.padding,
        "margin":   b.Margin(),  // ✅ 添加
        "flex":     b.flex,
    }
}
```

#### 修复 2：调整读取顺序
```go
// internal/render/fiber_adapter.go
func (a *FiberToNodeAdapter) GetMargin() layout.Margin {
    // ✅ 优先读取 Fiber 字段
    if a.fiber.LayoutMargin[0] != 0 || a.fiber.LayoutMargin[1] != 0 ||
       a.fiber.LayoutMargin[2] != 0 || a.fiber.LayoutMargin[3] != 0 {
        return layout.Margin{
            Top:    a.fiber.LayoutMargin[0],
            Right:  a.fiber.LayoutMargin[1],
            Bottom: a.fiber.LayoutMargin[2],
            Left:   a.fiber.LayoutMargin[3],
        }
    }

    // 向后兼容：从 Props 读取
    if a.fiber.Props != nil {
        if m, ok := a.fiber.Props["margin"].([4]int); ok {
            return layout.Margin{...}
        }
    }
    return layout.Margin{}
}
```

#### 修复 3：Flex 布局累积 margin
```go
// runtime/layout/types.go
mainAxisMarginOffset := 0
for i, childBox := range childBoxes {
    // ... 获取 margin

    childY := y + childBox.Y + borderOffsetY + mainAxisMarginOffset + marginTop
    mainAxisMarginOffset += marginTop + marginBottom  // ✅ 累积
}
```

### 7.4 验证
```go
// 运行测试
go test ./runtime/layout/...
// 输出：PASS

// 运行示例
go run ./examples/elegant_api_demo/debug_margin.go
// 输出：按钮之间显示正确的间距
```

---

## 8. 最佳实践

### 8.1 命名约定

| 层次 | 命名约定 | 示例 |
|------|---------|------|
| VNode 字段 | 描述性名称 | `padding`, `margin`, `flex` |
| Props Key | 小驼峰 | `"padding"`, `"margin"`, `"flex"` |
| Fiber 字段 | Layout + 属性名 | `LayoutPadding`, `LayoutMargin`, `LayoutFlex` |
| 接口方法 | Get + 属性名 | `GetPadding()`, `GetMargin()`, `GetFlex()` |
| 结构体方法 | Set + 属性名 | `SetPadding()`, `SetMargin()`, `SetFlex()` |

### 8.2 类型约定

| 属性类型 | 类型定义 | 示例 |
|---------|---------|------|
| 间距属性（四边） | `[4]int` | `[top, right, bottom, left]` |
| 单一间距 | `int` | `gap`, `flex` |
| 对齐属性 | `枚举` | `AlignStart`, `AlignCenter`, `AlignEnd` |
| 布尔属性 | `bool` | `stretchCross`, `fillWidth` |
| 复杂属性 | `结构体` | `layout.Margin`, `layout.Border` |

### 8.3 代码审查检查点

- [ ] VNode.Props() 包含了所有属性
- [ ] CreateFiber 正确提取了所有属性
- [ ] Fiber 字段正确赋值
- [ ] 适配器实现了所有接口方法
- [ ] 布局引擎正确使用了属性
- [ ] 添加了必要的测试
- [ ] 添加了调试日志（可选）

### 8.4 性能考虑

1. **避免重复计算**：在适配器中缓存常用属性
2. **使用接口检查**：使用类型断言而不是反射
3. **最小化 Props 大小**：只包含布局需要的属性
4. **延迟加载**：属性只在需要时才计算

---

## 9. 参考资料

### 相关文件
- `runtime/ui/fiber.go` - Fiber 结构体定义
- `runtime/ui/fiber_util.go` - CreateFiber 实现
- `runtime/ui/box_model.go` - BoxModelMixin 定义
- `internal/render/fiber_adapter.go` - FiberToNodeAdapter 实现
- `runtime/layout/types.go` - layout.Node 接口和布局引擎
- `ui/components/*/vnode.go` - 各种组件的 VNode 实现

### 相关文档
- `docs/fiber/fiber_first/` - Fiber-first 架构文档
- `docs/layout/` - 布局引擎文档
- `docs/component_design/` - 组件设计文档

### 相关接口
- `layout.Node` - 基础布局节点接口
- `layout.Marginal` - Margin 属性接口
- `layout.Bordered` - Border 属性接口
- `layout.FlexStyleProvider` - Flex 样式接口
- `layout.GridStyleProvider` - Grid 样式接口
