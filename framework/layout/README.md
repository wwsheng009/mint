# Layout - 布局组件

提供 TUI 应用的布局容器组件，基于 runtime flex 实现的 framework 层封装。

## 组件概览

| 组件 | 说明 | 文件 |
|------|------|------|
| `Box` | 盒子容器，支持边框、内边距、背景 | `box.go` |
| `Row` | 水平弹性布局（从左到右） | `flex.go` |
| `Column` | 垂直弹性布局（从上到下） | `flex.go` |
| `Flex` | 通用弹性布局容器 | `flex.go` |

## Row - 水平布局

将子组件从左到右排列。

```go
import "github.com/wwsheng009/mint/framework/layout"

row := layout.NewRow().
    Gap(1).                    // 子组件间距
    Padding(1).                // 内边距
    MainAlign(layout.MainCenter).  // 主轴居中
    CrossAlign(layout.CrossCenter). // 交叉轴居中
    AddChild(
        text1,
        text2,
        button,
    )
```

## Column - 垂直布局

将子组件从上到下排列。

```go
col := layout.NewColumn().
    Gap(1).                       // 子组件间距
    PaddingLTRB(1, 0, 1, 0).      // 上下内边距
    MainAlign(layout.MainStart).  // 从顶部开始
    AddChild(
        header,
        content,
        footer,
    )
```

## Flex - 弹性布局

通用的弹性布局容器，支持完整配置。

```go
flex := layout.NewFlex().
    Direction(layout.FlexRow).         // 方向
    MainAlign(layout.SpaceBetween).    // 两端对齐
    CrossAlign(layout.Center).         // 居中
    Gap(2).                            // 间距
    Padding(1).                        // 内边距
    Flex(0, layout.FlexConfig{         // 第一个子元素弹性
        Grow: 1,                        // 可放大
        Shrink: 1,                      // 可缩小
        Basis: 0,                       // 基础尺寸
    }).
    AddChild(child1, child2, child3)
```

## 对齐方式

### 主轴对齐 (MainAlign)

```go
layout.MainStart        // 起点（左/上）
layout.MainEnd          // 终点（右/下）
layout.MainCenter       // 居中
layout.MainSpaceBetween // 两端对齐
layout.MainSpaceAround  // 两侧间距相等
layout.MainSpaceEvenly  // 所有间距相等
```

### 交叉轴对齐 (CrossAlign)

```go
layout.CrossStart   // 起点
layout.CrossEnd     // 终点
layout.CrossCenter  // 居中
layout.CrossStretch // 拉伸填满
```

## 弹性配置

```go
// 子组件将按比例分配剩余空间
flex.
    AddChild(child1, child2, child3).
    Flex(0, layout.FlexConfig{Grow: 1}).  // child1 占 1 份
    Flex(1, layout.FlexConfig{Grow: 2}).  // child2 占 2 份
    // child3 固定尺寸

// 或者使用便捷方法
flex.
    AddChild(child1, child2).
    FlexGrow(0, 1).   // child1 可放大
    FlexBasis(1, 100) // child2 基础宽度 100
```

## 方向

```go
layout.FlexRow          // 水平，从左到右
layout.FlexColumn       // 垂直，从上到下
layout.FlexRowReverse   // 水平，从右到左
layout.FlexColumnReverse // 垂直，从下到上
```

## 内边距

```go
// 四边相同
flex.Padding(1)

// 分别设置
flex.PaddingLTRB(left, top, right, bottom)

// 垂直/水平
flex.PaddingV(1)  // 上下
flex.PaddingH(1)  // 左右
```

## 完整示例

### 表单布局

```go
func buildForm() *layout.Flex {
    return layout.NewColumn().
        Gap(1).
        Padding(2).
        AddChild(
            NewLabel("用户名:"),
            NewInput(),
            NewLabel("密码:"),
            NewInput(),
            layout.NewRow().
                Gap(2).
                MainAlign(layout.MainEnd).
                AddChild(
                    NewButton("取消"),
                    NewButton("确定"),
                ),
        )
}
```

### 卡片布局

```go
func buildCard() *layout.Box {
    return layout.NewBox().
        WithBorder(true).
        WithBorderType("rounded").
        WithBorderColor("gray").
        WithPadding(1).
        WithChild(
            layout.NewColumn().
                Gap(1).
                AddChild(
                    NewTitle("标题").
                        MainAlign(layout.MainCenter),
                    NewDivider(),
                    NewContent("内容..."),
                ),
        )
}
```

### 弹性布局

```go
func buildSplitView() *layout.Flex {
    return layout.NewRow().
        Gap(1).
        AddChild(
            // 左侧面板 - 固定宽度
            NewSidebar(),
            // 右侧内容 - 占据剩余空间
            NewContent().
                FlexGrow(0, 1),
        )
}
```

### 居中内容

```go
func buildCentered() *layout.Flex {
    return layout.NewRow().
        MainAlign(layout.MainCenter).    // 水平居中
        CrossAlign(layout.CrossCenter).  // 垂直居中
        AddChild(
            NewContent("居中内容"),
        )
}
```

## 子组件管理

```go
flex := layout.NewRow()

// 添加子组件
flex.AddChild(child1)
flex.AddChildren(child2, child3)

// 获取子组件
count := flex.ChildCount()
child := flex.GetChild(0)
children := flex.GetChildren()

// 移除子组件
flex.Remove(child2)
flex.RemoveAt(0)

// 清空所有子组件
flex.ClearChildren()
```

## 获取配置

```go
direction := flex.GetDirection()
mainAlign := flex.GetMainAlign()
crossAlign := flex.GetCrossAlign()
gap := flex.GetGap()
padding := flex.GetPadding()
bg := flex.GetBackground()
flexChildren := flex.GetFlexChildren()
```

## 与 Box 组合使用

```go
layout.NewBox().
    WithBorder(true).
    WithBorderType("double").
    WithPadding(1).
    WithChild(
        layout.NewColumn().
            Gap(1).
            AddChild(
                NewHeader("标题"),
                NewBody("内容"),
            ),
    )
```

## 架构说明

Flex 组件包装了 `runtime/layout.FlexLayout`:

```
framework/layout/Flex
    │
    ├── rtFlex *layout.FlexLayout (runtime)
    │   └── 实际布局计算引擎
    │
    ├── 配置方法 → 委托给 rtFlex
    │   ├── Direction()
    │   ├── MainAlign()
    │   ├── Gap()
    │   └── ...
    │
    └── componentNodeAdapter
        └── 将 framework.Node 转换为 runtime.layout.Node
```

## API 参考

### 构造函数

| 函数 | 说明 |
|------|------|
| `NewRow()` | 创建水平布局 |
| `NewColumn()` | 创建垂直布局 |
| `NewFlex()` | 创建通用弹性布局 |

### 配置方法（链式）

| 方法 | 说明 |
|------|------|
| `Direction(dir)` | 设置方向 |
| `MainAlign(align)` | 设置主轴对齐 |
| `CrossAlign(align)` | 设置交叉轴对齐 |
| `Gap(n)` | 设置主轴间距 |
| `CrossGap(n)` | 设置交叉轴间距 |
| `Padding(n)` | 设置四边内边距 |
| `PaddingV(n)` | 设置垂直内边距 |
| `PaddingH(n)` | 设置水平内边距 |
| `PaddingLTRB(l,t,r,b)` | 分别设置四边内边距 |
| `Background(color)` | 设置背景色 |
| `Flex(idx, cfg)` | 设置弹性配置 |
| `FlexGrow(idx, n)` | 设置放大比例 |
| `FlexBasis(idx, n)` | 设置基础尺寸 |

### 子组件管理

| 方法 | 说明 |
|------|------|
| `AddChild(child)` | 添加子组件 |
| `AddChildren(children...)` | 添加多个子组件 |
| `Remove(child)` | 移除子组件 |
| `RemoveAt(idx)` | 移除指定位置子组件 |
| `ClearChildren()` | 清空所有子组件 |
| `ChildCount()` | 获取子组件数量 |
| `GetChild(idx)` | 获取指定位置子组件 |
| `GetChildren()` | 获取所有子组件 |

### 获取器

| 方法 | 说明 |
|------|------|
| `GetDirection()` | 获取方向 |
| `GetMainAlign()` | 获取主轴对齐 |
| `GetCrossAlign()` | 获取交叉轴对齐 |
| `GetGap()` | 获取主轴间距 |
| `GetCrossGap()` | 获取交叉轴间距 |
| `GetPadding()` | 获取内边距 |
| `GetBackground()` | 获取背景色 |
| `GetFlexChildren()` | 获取弹性配置 |

## 最佳实践

1. **使用便捷构造器**: 大多数情况用 `NewRow()` / `NewColumn()` 即可
2. **合理使用间距**: `Gap()` 比手动添加 Space 组件更高效
3. **弹性配置**: `FlexGrow` 用于分配剩余空间，`FlexBasis` 用于设置基础尺寸
4. **与 Box 组合**: Box 用于边框/内边距，Flex 用于子组件布局
5. **嵌套布局**: 复杂布局可以拆分为多个嵌套的 Flex 容器
