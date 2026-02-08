# 布局调试 API 使用指南

## 概述

Mint TUI 现在提供了一个强大的布局调试 API，可以获取所有组件的详细布局信息，包括：

- 位置信息 (X, Y)
- 尺寸信息 (Width, Height)
- 布局约束 (MinWidth, MaxWidth, MinHeight, MaxHeight)
- 布局属性 (Flex, Gap, Align, Padding, Margin)
- 组件层级关系

## API 核心

### 1. LayoutTree 结构

```go
package debug

type LayoutTree struct {
    Root LayoutInfo
}

type LayoutInfo struct {
    // 组件标识
    Type     string  // VNode 类型 (button, text, hstack, etc.)
    Tag      string  // Tag (如果可用)
    Key      string  // Key (如果可用)
    Label    string  // Label (用于按钮/文本)
    Path     string  // 从根节点的路径 (如 "root.0.child.2")

    // 位置和尺寸
    X      int     // X 坐标 (绝对位置)
    Y      int     // Y 坐标 (绝对位置)
    Width  int     // 宽度
    Height int     // 高度

    // 约束
    MinWidth  int    // 最小宽度约束
    MaxWidth  int    // 最大宽度约束
    MinHeight int    // 最小高度约束
    MaxHeight int    // 最大高度约束

    // 布局属性
    Flex        int     // Flex 系数 (0 表示非弹性)
    Gap         int     // 布局容器的 gap
    Align       string  // 主轴对齐方式
    CrossAlign  string  // 交叉轴对齐方式
    Padding     [4]int  // Padding [top, right, bottom, left]
    Margin      [4]int  // Margin [top, right, bottom, left]
    IsContainer bool    // 是否为布局容器

    // 子组件
    Children []LayoutInfo  // 子组件的布局信息
}
```

## 使用方法

### 方法 1: 从 RenderingPipeline 获取布局信息

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/app"
    "github.com/wwsheng009/mint/internal/render"
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/debug"
)

func main() {
    // 1. 创建 UI
    root := buildUI()

    // 2. 创建渲染管线
    pipeline := render.NewRenderingPipeline()

    // 3. 设置约束
    constraints := runtime.NewBoxConstraints(0, 80, 0, 25)

    // 4. 执行布局计算
    layout, err := pipeline.ComputeLayout(root, constraints)
    if err != nil {
        panic(err)
    }

    // 5. 提取布局信息
    tree := debug.GetLayoutTree(layout)

    // 6. 打印格式化的布局树
    fmt.Println(debug.FormatLayoutTree(tree))
}
```

### 方法 2: 使用 Engine 直接获取布局信息

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/app"
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/compute"
    "github.com/wwsheng009/mint/runtime/debug"
)

func main() {
    // 1. 创建 UI
    root := buildUI()

    // 2. 创建布局引擎
    engine := compute.NewEngine()

    // 3. 设置约束
    constraints := runtime.NewBoxConstraints(0, 80, 0, 25)

    // 4. 执行布局计算
    layout, err := engine.Layout(root, constraints)
    if err != nil {
        panic(err)
    }

    // 5. 提取布局信息
    tree := debug.GetLayoutTree(layout)

    // 6. 打印格式化的布局树
    fmt.Println(debug.FormatLayoutTree(tree))
}
```

## 实际使用示例

### 示例 1: 获取所有按钮的信息

```go
layout, _ := engine.Layout(root, constraints)
tree := debug.GetLayoutTree(layout)

// 查找所有按钮
buttons := debug.FindComponentsByType(tree, "button")

for _, btn := range buttons {
    fmt.Printf("按钮: %s\n", btn.Label)
    fmt.Printf("  位置: (%d, %d)\n", btn.X, btn.Y)
    fmt.Printf("  尺寸: %dx%d\n", btn.Width, btn.Height)

    if btn.Flex > 0 {
        fmt.Printf("  Flex: %d\n", btn.Flex)
    }

    if btn.Padding != [4]int{} {
        fmt.Printf("  Padding: [top=%d, right=%d, bottom=%d, left=%d]\n",
            btn.Padding[0], btn.Padding[1], btn.Padding[2], btn.Padding[3])
    }

    fmt.Println()
}
```

**输出示例**:
```
按钮: Button 1
  位置: (3, 12)
  尺寸: 19x1
  Flex: 1

按钮: Button 2
  位置: (23, 12)
  尺寸: 19x1
  Flex: 1
  Padding: [top=1, right=2, bottom=1, left=2]
```

### 示例 2: 按路径查找组件

```go
layout, _ := engine.Layout(root, constraints)
tree := debug.GetLayoutTree(layout)

// 查找特定路径的组件
if comp, found := debug.GetComponentInfo(tree, "root.0.1"); found {
    fmt.Printf("找到组件: %s\n", comp.Type)
    fmt.Printf("  标签: %s\n", comp.Tag)
    fmt.Printf("  位置: (%d, %d)\n", comp.X, comp.Y)
    fmt.Printf("  尺寸: %dx%d\n", comp.Width, comp.Height)
}
```

### 示例 3: 命中测试 (Hit Testing)

```go
layout, _ := engine.Layout(root, constraints)
tree := debug.GetLayoutTree(layout)

// 查找特定位置的组件
x, y := 50, 10
if comp, found := debug.GetComponentAtPoint(tree, x, y); found {
    fmt.Printf("位置 (%d, %d) 的组件: %s\n", x, y, comp.Type)
    if comp.Label != "" {
        fmt.Printf("  标签: %q\n", comp.Label)
    }
    fmt.Printf("  位置: (%d, %d), 尺寸: %dx%d\n",
        comp.X, comp.Y, comp.Width, comp.Height)
} else {
    fmt.Printf("位置 (%d, %d) 没有组件\n", x, y)
}
```

### 示例 4: 查找所有布局容器

```go
layout, _ := engine.Layout(root, constraints)
tree := debug.GetLayoutTree(layout)

// 查找所有 HStack
hstacks := debug.FindComponentsByType(tree, "hstack")

for _, hs := range hstacks {
    fmt.Printf("HStack: path=%s\n", hs.Path)
    fmt.Printf("  位置: (%d, %d)\n", hs.X, hs.Y)
    fmt.Printf("  尺寸: %dx%d\n", hs.Width, hs.Height)
    fmt.Printf("  Gap: %d\n", hs.Gap)
    fmt.Printf("  Align: %s\n", hs.Align)
    fmt.Printf("  子组件数量: %d\n", len(hs.Children))
    fmt.Println()
}
```

## 输出格式示例

```
Layout Tree:
════════════════════════════════════════════════════════════════════════════════
📍 Root (vstack)
   Position: (0, 0) Size: 80x25

   ├─ root.0 (hstack) tag=hstack
      Position: (0, 1) Size: 80x3
      Layout: gap=1, align=Start

      ├─ root.0.0 (text) tag=text
      │  Position: (0, 1) Size: 13x1
      │  Label: "Actions:"
      │
      ├─ root.0.1 (button) tag=button
      │  Position: (14, 1) Size: 21x1
      │  label="Button 1"
      │  Flex: 1
      │
      ├─ root.0.2 (button) tag=button
      │  Position: (36, 1) Size: 21x1
      │  label="Button 2"
      │  Flex: 1
      │  Padding: [1 2 1 2]
      │
      └─ root.0.3 (button) tag=button
         Position: (58, 1) Size: 21x1
         label="Button 3"
         Flex: 1
════════════════════════════════════════════════════════════════════════════════
```

## 集成到现有应用

### 在 demo2 中添加布局调试

```go
// 在 demo2 的 main 函数中添加调试模式
func main() {
    // ... 原有代码 ...

    // 如果设置了 TUI_LAYOUT_DEBUG 环境变量，打印布局信息
    if os.Getenv("TUI_LAYOUT_DEBUG") == "true" {
        printLayoutInfo(RuntimeDemo)
    }

    // 运行应用
    err := ui.Run(RuntimeDemo, /* ... */)
    // ...
}

func printLayoutInfo(vnode ui.VNode) {
    engine := compute.NewEngine()
    constraints := runtime.NewBoxConstraints(0, 100, 0, 35)

    layout, err := engine.Layout(vnode, constraints)
    if err != nil {
        fmt.Printf("❌ 布局计算失败: %v\n", err)
        return
    }

    tree := debug.GetLayoutTree(layout)
    fmt.Println(debug.FormatLayoutTree(tree))
}
```

### 运行时查看布局

```bash
# 启用布局调试
TUI_LAYOUT_DEBUG=true go run examples/ui_demos/demo2_runtime_internals/main.go

# 或者编译后运行
go build -o demo2.exe examples/ui_demos/demo2_runtime_internals/main.go
TUI_LAYOUT_DEBUG=true ./demo2.exe
```

## 高级用法

### 可视化组件边界

```go
func visualizeLayout(tree *debug.LayoutTree) {
    fmt.Println("\n📐 组件边界可视化:")
    fmt.Println(strings.Repeat("─", 80))

    visualizeComponent(&tree.Root, 0)
}

func visualizeComponent(info *debug.LayoutInfo, depth int) {
    indent := strings.Repeat("  ", depth)

    // 绘制边框
    fmt.Printf("%s┌─ %s (%s) %dx%d\n",
        indent, info.Path, info.Type, info.Width, info.Height)
    fmt.Printf("%s│  位置: (%d, %d)\n", indent, info.X, info.Y)

    if info.Label != "" {
        fmt.Printf("%s│  标签: %q\n", indent, info.Label)
    }

    if info.Flex > 0 {
        fmt.Printf("%s│  Flex: %d\n", indent, info.Flex)
    }

    for _, child := range info.Children {
        visualizeComponent(&child, depth+1)
    }

    fmt.Printf("%s└\n", indent)
}
```

### 统计布局信息

```go
func analyzeLayout(tree *debug.LayoutTree) {
    // 统计组件类型
    typeCount := make(map[string]int)
    countComponents(&tree.Root, typeCount)

    fmt.Println("\n📊 组件统计:")
    fmt.Println(strings.Repeat("─", 40))
    for t, count := range typeCount {
        fmt.Printf("  %s: %d\n", t, count)
    }

    // 查找最宽和最高的组件
    widest := findWidest(&tree.Root)
    tallest := findTallest(&tree.Root)

    fmt.Println("\n📐 尺寸极值:")
    fmt.Printf("  最宽: %s (%dx%d)\n", widest.Type, widest.Width, widest.Height)
    fmt.Printf("  最高: %s (%dx%d)\n", tallest.Type, tallest.Width, tallest.Height)
}
```

## 常见用例

### 1. 调试布局问题

当组件没有按预期布局时：

```go
layout, _ := engine.Layout(root, constraints)
tree := debug.GetLayoutTree(layout)

// 找到有问题的组件
comp, _ := debug.GetComponentInfo(tree, "root.1.2")

fmt.Printf("组件约束: MinWidth=%d, MaxWidth=%d\n",
    comp.MinWidth, comp.MaxWidth)
fmt.Printf("组件实际尺寸: %dx%d\n", comp.Width, comp.Height)

if comp.Width < comp.MinWidth || comp.Width > comp.MaxWidth {
    fmt.Printf("⚠️ 警告: 组件宽度 %d 超出约束 [%d, %d]\n",
        comp.Width, comp.MinWidth, comp.MaxWidth)
}
```

### 2. 验证 Flex 布局

```go
// 检查 flex 子节点是否正确分配空间
layout, _ := engine.Layout(root, constraints)
tree := debug.GetLayoutTree(layout)

hstack, _ := debug.GetComponentInfo(tree, "root.0")
totalFlex := 0
totalWidth := 0

for _, child := range hstack.Children {
    if child.Flex > 0 {
        totalFlex += child.Flex
        totalWidth += child.Width
        fmt.Printf("Flex %d 子组件: 宽度=%d\n", child.Flex, child.Width)
    }
}

fmt.Printf("Flex 总和: %d, 总宽度: %d\n", totalFlex, totalWidth)
```

### 3. 检测重叠组件

```go
func detectOverlaps(tree *debug.LayoutTree) [][]debug.LayoutInfo {
    var allComponents []debug.LayoutInfo
    collectComponents(&tree.Root, &allComponents)

    var overlaps [][]debug.LayoutInfo

    for i, c1 := range allComponents {
        for j, c2 := range allComponents {
            if i >= j {
                continue
            }

            if isOverlapping(&c1, &c2) {
                overlaps = append(overlaps, []debug.LayoutInfo{c1, c2})
                fmt.Printf("⚠️ 重叠检测:\n")
                fmt.Printf("  %s (%d,%d,%dx%d)\n",
                    c1.Type, c1.X, c1.Y, c1.Width, c1.Height)
                fmt.Printf("  %s (%d,%d,%dx%d)\n",
                    c2.Type, c2.X, c2.Y, c2.Width, c2.Height)
            }
        }
    }

    return overlaps
}

func isOverlapping(a, b *debug.LayoutInfo) bool {
    return !(a.X+a.Width <= b.X ||
             b.X+b.Width <= a.X ||
             a.Y+a.Height <= b.Y ||
             b.Y+b.Height <= a.Y)
}
```

## API 参考

### debug.GetLayoutTree

```go
func GetLayoutTree(layout *compute.ComputedLayout) *LayoutTree
```

从 ComputedLayout 提取布局信息树。

**参数**:
- `layout`: ComputedLayout 实例，从 Engine.Layout() 或 RenderingPipeline.ComputeLayout() 获取

**返回**:
- `*LayoutTree`: 包含完整布局信息的树

### debug.FormatLayoutTree

```go
func FormatLayoutTree(tree *LayoutTree) string
```

将布局树格式化为易读的字符串。

**参数**:
- `tree`: LayoutTree 实例

**返回**:
- `string`: 格式化的布局树字符串

### debug.FindComponentsByType

```go
func FindComponentsByType(tree *LayoutTree, vtype string) []LayoutInfo
```

查找所有指定类型的组件。

**参数**:
- `tree`: LayoutTree 实例
- `vtype`: VNode 类型或tag (如 "button", "text", "hstack")

**返回**:
- `[]LayoutInfo`: 匹配的组件列表

### debug.GetComponentInfo

```go
func GetComponentInfo(tree *LayoutTree, path string) (LayoutInfo, bool)
```

按路径查找组件。

**参数**:
- `tree`: LayoutTree 实例
- `path`: 组件路径 (如 "root.0.1.2")

**返回**:
- `LayoutInfo`: 找到的组件信息
- `bool`: 是否找到

### debug.GetComponentAtPoint

```go
func GetComponentAtPoint(tree *LayoutTree, x, y int) (LayoutInfo, bool)
```

查找指定位置的组件。

**参数**:
- `tree`: LayoutTree 实例
- `x`, `y`: 位置坐标

**返回**:
- `LayoutInfo`: 该位置的组件
- `bool`: 是否找到组件

## 注意事项

1. **必须在布局计算后调用**: GetLayoutTree 需要 ComputedLayout，所以必须先调用 Engine.Layout() 或 RenderingPipeline.ComputeLayout()

2. **性能考虑**: 对于大型 UI 树，频繁调用布局计算可能影响性能。建议仅在调试模式下使用。

3. **线程安全**: 布局引擎不是线程安全的，不要在多个 goroutine 中同时调用。

4. **缓存无效**: 如果修改了 VNode 树，需要重新计算布局，旧的 LayoutTree 将不再有效。

## 完整示例

参见 `docs/debug_api_example.go` 获取完整的使用示例。

---

**版本**: v1.0
**最后更新**: 2025-02-07
**维护者**: Claude Sonnet 4.5
