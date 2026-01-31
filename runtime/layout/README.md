# Layout Engine (布局引擎)

布局引擎是 Mint Runtime 的核心子系统，负责计算 UI 组件的位置和尺寸。

## 职责

- **Flexbox 布局算法** - 实现标准的 Flexbox 布局规则
- **约束计算和传递** - 自顶向下传递空间约束
- **尺寸测量** - 自底向上计算组件内在尺寸
- **缓存优化** - 缓存布局结果避免重复计算
- **脏标记系统** - 跟踪需要重新布局的组件

## 核心概念

### 布局流程

布局分为两个阶段：

```
1. Measure Phase (测量阶段)
   ↓
   自底向上遍历组件树
   ↓
   每个组件返回其理想尺寸

2. Layout Phase (布局阶段)
   ↓
   自顶向下遍历组件树
   ↓
   根据约束分配最终位置和尺寸
```

### BoxConstraints

约束定义了组件可用的最小和最大空间：

```go
type BoxConstraints struct {
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
}
```

**类型**:
- **Tight Constraints** (紧约束): Min == Max - 固定尺寸
- **Loose Constraints** (松约束): Min = 0, Max = 已知 - 弹性尺寸
- **Unbounded Constraints** (无界约束): Max < 0 - 无限制

### 布局方向

支持两种主要的布局方向：

- **Column (垂直)**: 子组件垂直排列（默认）
- **Row (水平)**: 子组件水平排列

## 使用示例

### 基本使用

```go
import "github.com/wwsheng009/mint/runtime/layout"

// 创建布局引擎
engine := layout.NewEngine()

// 准备节点树
nodes := []runtime.LayoutNode{
    *runtime.NewLayoutNode("box1", runtime.NodeTypeBox, runtime.Style{}),
    *runtime.NewLayoutNode("box2", runtime.NodeTypeBox, runtime.Style{}),
}

// 计算布局
constraints := runtime.NewBoxConstraints(0, 80, 0, 24)
result := engine.Layout(nodes, constraints)

// 查看结果
for _, box := range result.Boxes {
    fmt.Printf("%s: x=%d, y=%d, w=%d, h=%d\n",
        box.ID, box.X, box.Y, box.Width, box.Height)
}
```

### Flexbox 布局

```go
// 创建垂直 Flex 容器
container := flex.NewContainer(flex.Column, flex.Stretch)
container.SetGap(1)
container.SetPadding(1, 1, 1, 1)

// 添加子组件
container.AddChild(child1, flex.Grow(1))
container.AddChild(child2, flex.Fixed(10)) // 固定 10 列宽

// 计算布局
constraints := runtime.NewBoxConstraints(0, 80, 0, 24)
container.Layout(constraints)
```

### 布局缓存

```go
// 引擎自动缓存布局结果
engine := layout.NewEngine()

// 第一次布局 - 缓存未命中
result1 := engine.Layout(nodes, constraints)
//  stats.CacheMisses++

// 第二次布局（相同输入）- 缓存命中
result2 := engine.Layout(nodes, constraints)
//  stats.CacheHits++
```

## 核心类型

### Engine

布局引擎主类：

```go
type Engine struct {
    cache *Cache
    stats *Stats
}

func NewEngine() *Engine
func (e *Engine) Layout(nodes []Node, constraints Constraints) *LayoutResult
func (e *Engine) GetStats() Stats
func (e *Engine) ClearCache()
```

### LayoutResult

布局结果：

```go
type LayoutResult struct {
    Boxes []LayoutBox
}

type LayoutBox struct {
    ID     string
    Type   NodeType
    X, Y   int
    Width, Height int
    Style  Style
}
```

### Cache

布局缓存：

```go
type Cache struct {
    entries map[string]*CachedLayout
    maxSize int
}

func (c *Cache) Get(nodes []Node, constraints Constraints) *LayoutResult
func (c *Cache) Put(nodes []Node, constraints Constraints, result *LayoutResult)
```

### Node接口

节点必须实现此接口：

```go
type Node interface {
    ID() string
    Type() NodeType
    Children() []Node
    SetParent(parent Node)
    Parent() Node

    // 可选接口
    Measurable  // 如果可测量
    Layoutable  // 如果可布局
}
```

### Measurable

可测量组件接口：

```go
type Measurable interface {
    Measure(constraints Constraints) Size
}
```

**实现示例**:

```go
func (t *TextComponent) Measure(constraints Constraints) Size {
    textWidth := len(t.Text)
    textHeight := strings.Count(t.Text, "\n") + 1

    // 应用约束
    w := clamp(textWidth, constraints.MinWidth, constraints.MaxWidth)
    h := clamp(textHeight, constraints.MinHeight, constraints.MaxHeight)

    return Size{Width: w, Height: h}
}
```

## 文件结构

```
layout/
├── README.md              # 本文档
├── engine.go              # 布局引擎
├── flex.go                # Flexbox 算法
├── cache.go               # 布局缓存
├── types.go               # 相关类型定义
└── (未来添加)
    ├── grid.go            # Grid 布局
    ├── absolute.go        # 绝对定位
    └── utils.go           # 工具函数
```

## 性能优化

### 缓存策略

- 基于节点树和约束的哈希键
- 最多缓存 `maxSize` 个布局结果（默认 1000）
- LRU 淘汰策略

### 脏标记

组件可以标记为脏，触发重新布局：

```go
// 标记整个子树为脏
node.MarkDirty()

// 仅标记布局为脏
node.MarkLayoutDirty()

// 仅标记绘制为脏
node.MarkPaintDirty()
```

### 统计信息

```go
stats := engine.GetStats()
fmt.Printf("Total Layouts: %d\n", stats.TotalLayouts)
fmt.Printf("Cache Hits: %d\n", stats.CacheHits)
fmt.Printf("Cache Misses: %d\n", stats.CacheMisses)
fmt.Printf("Hit Rate: %.2f%%\n",
    float64(stats.CacheHits)/float64(stats.TotalLayouts)*100)
```

## 常见问题

### Q: 为什么我的组件没有正确布局？

A: 确保组件实现了 `Measurable` 和 `Layoutable` 接口。

### Q: 如何处理动态内容尺寸？

A: 在 `Measure` 方法中根据内容计算尺寸，在 `Layout` 方法中应用约束。

### Q: 缓存会影响实时性吗？

A: 不会。缓存仅在输入完全相同时复用，任何样式或结构变化都会触发重新计算。

### Q: 如何实现嵌套布局？

A: 容器组件在测量时递归调用子组件的 `Measure`，在布局时递归调用子组件的 `Layout`。

## 相关文档

- [Flexbox 规范](https://www.w3.org/TR/css-flexbox/)
- [Layout 约束设计](../../framework/docs/layout/constraints.md)
- [DevTools 集成](../../devtools/docs/)

## 测试

```bash
go test ./runtime/layout/... -v
```

## 注意事项

此目录是纯布局内核，必须保持纯 Go 实现，不能依赖：
- ✅ Bubble Tea
- ✅ DSL 解析器
- ✅ 具体组件
- ✅ lipgloss
