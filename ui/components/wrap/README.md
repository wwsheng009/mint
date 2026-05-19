# Wrap Layout Component

Fiber-first 换行布局组件，实现类似 CSS `flex-wrap: wrap` 的功能。当子元素超出容器宽度时自动换行。

## 目录

- [概述](#概述)
- [架构设计](#架构设计)
- [布局实施](#布局实施)
- [API 参考](#api-参考)
- [Builder 模式](#builder-模式)
- [示例](#示例)
- [模块关系](#模块关系)
- [注意事项](#注意事项)
- [性能建议](#性能建议)

## 概述

Wrap 组件是一个**换行布局容器**，核心特性：
- 当子元素总宽度超过容器宽度时**自动换行**
- 支持**行内间距**和**行间距**分别设置
- 支持每行**对齐方式**（左对齐/居中/右对齐）
- 支持**内边距**

```
容器宽度 = 20
子元素: [Button1] [Button2] [Button3] [Button4] [Button5]

渲染结果:
┌──────────────────┐
│[Button1] [Button2]│  ← 行1
│[Button3] [Button4]│  ← 行2
│[Button5]          │  ← 行3
└──────────────────┘
```

## 架构设计

Wrap 组件遵循 **Fiber-first 架构**，严格遵守分层设计：

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Fiber-first 架构                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   ┌──────────────┐      CreateInstance()      ┌──────────────────┐  │
│   │    VNode     │  ─────────────────────────▶ │     Instance     │  │
│   │   (描述层)    │                            │    (运行时实体)   │  │
│   │  - 无状态     │                            │  - 持久化状态     │  │
│   │  - 无闭包     │                            │  - Measure()     │  │
│   │  - 无渲染     │                            │  - Paint()       │  │
│   └──────────────┘                            └──────────────────┘  │
│          │                                              │           │
│          │ Props                                        │           │
│          ▼                                              │           │
│   ┌──────────────┐                                     │           │
│   │    Fiber     │                                     │           │
│   │  (状态保持)   │                                     │           │
│   │  - 树结构     │                                     │           │
│   │  - 调度单元   │                                     │           │
│   └──────────────┘                                     │           │
│          │                                              │           │
│          │ WrapStyleProvider                            │           │
│          ▼                                              │           │
│   ┌──────────────┐      WrapLayout.LayoutChildren()    │           │
│   │  LayoutBox   │  ◀───────────────────────────────────┘           │
│   │  (布局结果)   │                                                  │
│   │  - X, Y      │                                                  │
│   │  - W, H      │                                                  │
│   └──────────────┘                                                  │
│          │                                                           │
│          │ 坐标                                                      │
│          ▼                                                           │
│   ┌──────────────┐                                                  │
│   │ PaintableBox │                                                  │
│   │  (绘制层)     │                                                  │
│   │  - 递归绘制   │                                                  │
│   └──────────────┘                                                  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 核心文件

| 文件 | 职责 |
|------|------|
| `vnode.go` | VNode 定义，纯描述层，实现 `rtui.VNode` 和 `rtui.InstanceFactory` |
| `instance.go` | Instance 定义，运行时实体，实现 `Measure()` 和 `Paint()` |
| `builder.go` | Builder API，提供流畅的链式构造方法 |

### 接口实现

```go
// VNode 实现的接口
var (
    _ rtui.VNode           = (*VNode)(nil)  // 节点描述
    _ rtui.InstanceFactory = (*VNode)(nil)  // 实例工厂
)

// Instance 实现的接口
var (
    _ rtui.ComponentInstance  = (*Instance)(nil)  // 组件实例
    _ rtui.PaintableInstance  = (*Instance)(nil)  // 可绘制
    _ layout.Measurable       = (*Instance)(nil)  // 可测量
)
```

## 布局实施

### 两遍布局算法

Wrap 布局采用**两遍布局**策略：

```
┌─────────────────────────────────────────────────────────────────┐
│                      第一遍: Measure                             │
├─────────────────────────────────────────────────────────────────┤
│ 1. 测量所有子元素尺寸 (childMeasure[i])                          │
│ 2. 根据容器宽度计算换行 (calculateRows)                          │
│ 3. 计算每行高度 (rowHeights)                                     │
│ 4. 返回总尺寸 (Size{Width, Height})                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      第二遍: Layout                              │
├─────────────────────────────────────────────────────────────────┤
│  由 WrapLayout.LayoutChildren() 执行:                           │
│ 1. 使用 Measure 阶段的 rows 和 rowHeights                        │
│ 2. 根据 Align 计算每行的起始 X 位置                               │
│ 3. 计算每个子元素的 (X, Y) 坐标                                   │
│ 4. 返回 LayoutBox 数组                                           │
└─────────────────────────────────────────────────────────────────┘
```

### 换行算法

```go
// calculateRows 将子元素划分到各行
func (inst *Instance) calculateRows(childWidths []int, availableWidth int) [][]int {
    var rows [][]int
    currentRow := []int{}
    currentWidth := 0

    for i, childWidth := range childWidths {
        // 检查是否需要换行
        shouldWrap := len(currentRow) > 0 &&
            (currentWidth + childWidth + inst.gap > availableWidth)

        if shouldWrap {
            // 完成当前行，开始新行
            rows = append(rows, currentRow)
            currentRow = []int{i}
            currentWidth = childWidth
        } else {
            // 添加到当前行
            currentRow = append(currentRow, i)
            currentWidth += childWidth
            if len(currentRow) > 1 {
                currentWidth += inst.gap
            }
        }
    }

    // 保存最后一行
    if len(currentRow) > 0 {
        rows = append(rows, currentRow)
    }

    return rows
}
```

### 布局引擎集成

Wrap 通过 `WrapStyleProvider` 接口与布局引擎集成：

```go
// runtime/layout/wrap.go
type WrapStyleProvider interface {
    GetWrapStyle() *WrapStyle
}

type WrapStyle struct {
    Width      int           // 容器宽度
    Gap        int           // 行内间距
    RowGap     int           // 行间距 (0 = 使用 Gap)
    Align      WrapAlignment // 行对齐
    Padding    Padding       // 内边距
    FillWidth  bool          // 拉伸行填满宽度
    FillHeight bool          // 拉伸容器填满父高度
}
```

### 布局引擎处理流程

```
                    Fiber Tree
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              FiberToNodeAdapter                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  GetWrapStyle() *WrapStyle                          │    │
│  │  - 从 Fiber.Props 提取布局属性                       │    │
│  │  - 转换 Align 类型 (rtui.Align → layout.WrapAlign)  │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              layout.Engine.layoutNodeWithDepth()             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  // 检查 WrapStyleProvider 接口                      │    │
│  │  if wrapProvider, ok := node.(WrapStyleProvider) {  │    │
│  │      wrapStyle := wrapProvider.GetWrapStyle()       │    │
│  │      wrap := NewWrapLayout(node.ID(), wrapStyle)    │    │
│  │      wrap.SetChildren(node.Children())              │    │
│  │      childBoxes := wrap.LayoutChildren(w, h)        │    │
│  │      // 递归布局子节点...                            │    │
│  │  }                                                  │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
                   LayoutBox Tree
```

## API 参考

### 构造函数

| 函数 | 说明 |
|------|------|
| `New()` | 创建 Wrap VNode |
| `Wrap(children ...)` | 快捷创建带子元素的 Wrap |
| `WrapWithWidth(w, children ...)` | 创建指定宽度的 Wrap |
| `WrapWithGap(gap, children ...)` | 创建指定间距的 Wrap |
| `WrapConfig(w, gap, align, children ...)` | 完整配置创建 |
| `W()` | 创建 Builder |

### 布局属性

| 方法 | 说明 | 默认值 |
|------|------|--------|
| `SetWidth(w int)` | 设置容器宽度（决定何时换行） | `80` |
| `SetGap(gap int)` | 设置行内元素间距 | `1` |
| `SetRowGap(rowGap int)` | 设置行间距（0 = 使用 Gap） | `0` |
| `SetAlign(a Align)` | 设置行内元素对齐 | `AlignStart` |
| `SetPadding(t, r, b, l int)` | 设置内边距 | `0, 0, 0, 0` |
| `SetFillWidth(bool)` | 拉伸每行填满容器 | `false` |
| `SetFillHeight(bool)` | 拉伸容器填满父高度 | `false` |

### 对齐方式

```go
const (
    AlignStart  = rtui.AlignStart   // 左对齐（默认）
    AlignCenter = rtui.AlignCenter  // 居中对齐
    AlignEnd    = rtui.AlignEnd     // 右对齐
)
```

### 子元素管理

| 方法 | 说明 |
|------|------|
| `SetChildrenList(children)` | 设置子元素列表 |
| `AddChild(child)` | 添加单个子元素 |

## Builder 模式

### 链式调用

```go
wrap.New().
    SetWidth(50).
    SetGap(2).
    SetAlign(wrap.AlignCenter).
    SetChildrenList([]rtui.VNode{
        text.New("A"),
        text.New("B"),
        text.New("C"),
    })
```

### Builder 类型

```go
// 使用 W() 创建 Builder
wrap.W().
    Width(40).
    Gap(2).
    Align(wrap.AlignCenter).
    Children(
        text.New("A"),
        text.New("B"),
    ).
    Build()

// 使用 NewBuilder
wrap.NewBuilder().
    Width(40).
    Gap(2).
    Children(...).
    Build()
```

### 快捷函数

```go
// 最简创建
wrap.Wrap(
    text.New("A"),
    text.New("B"),
    text.New("C"),
)

// 指定宽度
wrap.WrapWithWidth(30,
    text.New("Short"),
    text.New("Medium"),
    text.New("Very Long"),
)

// 指定间距
wrap.WrapWithGap(3,
    button.New("OK"),
    button.New("Cancel"),
)

// 完整配置
wrap.WrapConfig(50, 2, wrap.AlignCenter,
    text.New("#golang"),
    text.New("#fiber"),
    text.New("#tui"),
)
```

## 示例

### 基础换行

```go
wrap.New().
    SetWidth(20).
    SetGap(1).
    SetChildrenList([]rtui.VNode{
        text.New("[Button1]"),
        text.New("[Button2]"),
        text.New("[Button3]"),
        text.New("[Button4]"),
        text.New("[Button5]"),
    })

// 渲染结果:
// [Button1] [Button2]
// [Button3] [Button4]
// [Button5]
```

### 按钮工具栏

```go
wrap.New().
    SetWidth(50).
    SetGap(2).
    SetChildrenList([]rtui.VNode{
        button.New("New"),
        button.New("Open"),
        button.New("Save"),
        button.New("Cut"),
        button.New("Copy"),
        button.New("Paste"),
        button.New("Undo"),
        button.New("Redo"),
    })

// 渲染结果:
// [ New ]  [ Open ]  [ Save ]  [ Cut ]
// [ Copy ]  [ Paste ]  [ Undo ]  [ Redo ]
```

### 行间距设置

```go
// RowGap = 0 时使用 Gap 作为行间距
wrap.New().
    SetWidth(25).
    SetGap(1).
    SetRowGap(0).  // 行间距 = gap = 1
    SetChildrenList(...)

// RowGap > 0 时使用独立的行间距
wrap.New().
    SetWidth(25).
    SetGap(1).      // 行内间距 = 1
    SetRowGap(3).   // 行间距 = 3
    SetChildrenList(...)

// 渲染结果:
// [Item1] [Item2] [Item3]
//                        ← 3 行间距
//                        ← 2 行间距
//                        ← 1 行间距
// [Item4] [Item5]
```

### 行对齐

```go
// 左对齐 (默认)
wrap.New().
    SetWidth(40).
    SetAlign(wrap.AlignStart).
    SetChildrenList([]rtui.VNode{
        text.New("[A]"),
        text.New("[B]"),
    })
// [A] [B]

// 居中对齐
wrap.New().
    SetWidth(40).
    SetAlign(wrap.AlignCenter).
    SetChildrenList(...)
//              [A] [B]

// 右对齐
wrap.New().
    SetWidth(40).
    SetAlign(wrap.AlignEnd).
    SetChildrenList(...)
//                             [A] [B]
```

### 带内边距

```go
wrap.New().
    SetWidth(30).
    SetGap(1).
    SetPadding(1, 2, 1, 2).  // top, right, bottom, left
    SetChildrenList([]rtui.VNode{
        text.New("[A]"),
        text.New("[B]"),
        text.New("[C]"),
    })

// 渲染结果:
// ┌──────────────────────┐
// │                      │ ← top: 1
// │  [A] [B]             │
// │  [C]                 │ ← left: 2
// │                      │ ← bottom: 1
// └──────────────────────┘
```

### Tag Cloud

```go
wrap.New().
    SetWidth(55).
    SetGap(2).
    SetRowGap(1).
    SetChildrenList([]rtui.VNode{
        text.New("#golang"),
        text.New("#fiber"),
        text.New("#tui"),
        text.New("#terminal"),
        text.New("#ui"),
        text.New("#layout"),
        text.New("#wrap"),
        text.New("#flex"),
        text.New("#component"),
        text.New("#render"),
    })

// 渲染结果:
// #golang  #fiber  #tui  #terminal  #ui  #layout  #wrap
// #flex  #component  #render
```

### 窄容器

```go
wrap.New().
    SetWidth(15).
    SetGap(0).
    SetChildrenList([]rtui.VNode{
        text.New("[Btn1]"),
        text.New("[Btn2]"),
        text.New("[Btn3]"),
        text.New("[Btn4]"),
    })

// 渲染结果:
// [Btn1][Btn2]
// [Btn3][Btn4]
```

## 模块关系

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           模块依赖关系                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────┐                                                    │
│  │ ui/components/  │                                                    │
│  │     wrap/       │◀──────────────────────────────────────────┐       │
│  │  ┌───────────┐  │                                          │       │
│  │  │ vnode.go  │──┼──▶ rtui.VNode, rtui.InstanceFactory      │       │
│  │  │instance.go│──┼──▶ rtui.ComponentInstance, Paintable     │       │
│  │  │builder.go │──┼──▶ 便捷 API                               │       │
│  │  └───────────┘  │                                          │       │
│  └─────────────────┘                                          │       │
│          │                                                    │       │
│          │ Props                                              │       │
│          ▼                                                    │       │
│  ┌─────────────────┐                                          │       │
│  │  runtime/ui/   │                                          │       │
│  │   (rtui)       │◀──────────────────────────────────────────┤       │
│  │  - VNode       │                                           │       │
│  │  - Props       │                                           │       │
│  │  - Fiber       │                                           │       │
│  └─────────────────┘                                          │       │
│          │                                                    │       │
│          │ WrapStyleProvider                                  │       │
│          ▼                                                    │       │
│  ┌─────────────────┐                                          │       │
│  │ runtime/layout/ │                                          │       │
│  │   wrap.go       │◀─── WrapStyle, WrapLayout               │       │
│  │   types.go      │◀─── Engine.WrapStyleProvider 检查       │       │
│  └─────────────────┘                                          │       │
│          │                                                    │       │
│          │ LayoutBox                                          │       │
│          ▼                                                    │       │
│  ┌─────────────────┐                                          │       │
│  │ internal/render │                                          │       │
│  │ fiber_adapter.go│◀─── GetWrapStyle() 实现                 │       │
│  │ declarative_*.go│◀─── PaintEngine 递归渲染                │       │
│  └─────────────────┘                                          │       │
│                                                                │       │
└─────────────────────────────────────────────────────────────────────────┘
```

### 数据流

```
1. VNode 创建
   wrap.New().SetWidth(40).SetChildrenList([...])
                    │
                    ▼
2. Fiber 创建 (通过 InstanceFactory)
   vnode.CreateInstance() → Instance
   Fiber.Instance = Instance
   Fiber.Props = {width: 40, gap: 1, children: [...]}
                    │
                    ▼
3. 布局阶段 (通过 WrapStyleProvider)
   adapter.GetWrapStyle() → &WrapStyle{Width: 40, Gap: 1, ...}
   NewWrapLayout(id, style).LayoutChildren(w, h) → []*LayoutBox
                    │
                    ▼
4. 渲染阶段 (通过 PaintableInstance)
   PaintEngine 使用 LayoutBox 坐标递归调用子元素的 Paint()
```

## 注意事项

### 宽度必须设置

Wrap 的 `Width` 决定何时换行，**必须设置合理的宽度**：

```go
// ❌ 错误：使用默认宽度 80，可能不符合预期
wrap.New().SetChildrenList(lotsOfButtons)

// ✅ 正确：明确设置宽度
wrap.New().
    SetWidth(50).
    SetChildrenList(lotsOfButtons)
```

### Gap vs RowGap

- `Gap`: 行内元素之间的水平间距
- `RowGap`: 行与行之间的垂直间距，**为 0 时使用 Gap 值**

### 与 Border 组件结合使用

当 Wrap 作为 Border 的子元素时，**Border 需要显式设置高度**：

```go
// Border.Measure() 不自动测量子元素高度
// 需要先测量 Wrap 高度，再设置到 Border

w := wrap.New().
    SetWidth(40).
    SetGap(1).
    SetChildrenList(children)

// 测量 Wrap 高度
inst := wrap.NewInstance(w.Props())
size := inst.Measure(layout.Constraints{
    MaxWidth:  200,
    MaxHeight: 200,
})

// 创建 Border 并设置显式高度
border.New().
    SetWidth(w.Width()).
    SetHeight(size.Height).  // 必须设置！
    SetChild(w)
```

**重要**：Border 组件的 `Measure()` 方法只使用显式设置的 `width` 和 `height`，不会自动测量子元素。这是 Fiber-first 架构的设计：布局引擎分别测量各层，避免重复计算。

```go
// Gap=2, RowGap=0 (行间距 = Gap = 2)
wrap.New().SetGap(2).SetRowGap(0)

// Gap=2, RowGap=5 (行间距 = 5)
wrap.New().SetGap(2).SetRowGap(5)
```

### Instance.Paint() 返回 nil

Wrap 是**纯布局容器**，没有自身需要绘制的内容：

```go
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
    // 纯布局容器，布局由 LayoutBox 处理
    // 子元素渲染由 PaintEngine 递归处理
    return nil
}
```

### 子元素必须有尺寸

Wrap 依赖子元素的 `Measure()` 返回的尺寸来计算布局。如果子元素没有实现 `Measurable`，会使用默认值：

```go
// 默认尺寸: Width=10, Height=1
func (inst *Instance) estimateChildSize(child rtui.VNode, constraints layout.Constraints) Size {
    // 尝试从 Props 获取
    // 尝试 Measurable 接口
    // 最后使用默认值
}
```

### 不支持跨行元素

当前实现**不支持**元素跨行（类似 CSS 的 `rowspan`）。每个元素只能在其计算的行内。

### 对齐作用于每行

`Align` 设置的是**每行的对齐方式**，而不是整体对齐：

```go
// 每行独立对齐
wrap.New().
    SetWidth(40).
    SetAlign(wrap.AlignCenter).
    SetChildrenList(...)
// [A] [B]
//        ← 这行居中
// [C]
//      ← 这行也居中
```

## 性能建议

### 1. 复用 VNode

VNode 是不可变的，可以安全复用：

```go
// 创建一次，多次使用
btnStyle := button.New("Item")

// ❌ 避免在循环中重复创建相同的 VNode
for i := 0; i < 100; i++ {
    // 每次都创建新的 VNode
}

// ✅ 预创建或使用工厂函数
createButton := func(label string) *button.VNode {
    return button.New(label).SetVariant(button.VariantPrimary)
}
```

### 2. 合理设置宽度

过大的宽度导致单行元素过多，过小导致频繁换行：

```go
// ✅ 根据内容和显示区域设置合理宽度
wrap.New().
    SetWidth(terminalWidth - padding).
    SetChildrenList(items)
```

### 3. 避免过深的嵌套

虽然支持嵌套，但建议控制在合理深度：

```go
// ❌ 过深嵌套
wrap.New().SetChildrenList([]rtui.VNode{
    wrap.New().SetChildrenList([]rtui.VNode{
        wrap.New().SetChildrenList([]rtui.VNode{
            // ...
        }),
    }),
})

// ✅ 扁平化结构
wrap.New().SetChildrenList([]rtui.VNode{
    stack.NewVStack().SetChildrenList(...),
    stack.NewVStack().SetChildrenList(...),
})
```

### 4. 批量设置子元素

```go
// ✅ 一次性设置所有子元素
wrap.New().SetChildrenList(children)

// ⚠️ 多次添加会有额外开销
wrap.New().AddChild(a).AddChild(b).AddChild(c)
```

## 测试

运行单元测试：

```bash
go test ./ui/components/wrap/... -v
```

历史单组件示例已归档到 `docsArchive/cleanup-2026-05-19/_examples/fiber_firsts/wrap_demo/`。当前建议通过组件单测和 `examples/ui_demos/demo2_runtime_internals` 中的组合界面验证 Wrap 行为。

## 相关组件

- **Stack**: 单行/单列布局（不换行）
- **Grid**: 网格布局（固定行列）
- **Wrap**: 换行布局（自动换行）

选择建议：
- 需要固定行列数 → **Grid**
- 单行/单列 → **Stack**
- 元素数量不确定，需要自动换行 → **Wrap**
