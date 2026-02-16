# Layout ASCII Visualization Demo

## Overview

本程序使用 ASCII 艺术可视化布局结果，显示每个节点的位置、尺寸和层级关系。

## Running the Demo

```bash
go run examples/fiber_demos/layout_ascii_visualization/main.go
```

## What It Does

程序为每个 `component_fixtures` 组件生成三种可视化：

1. **Layout Statistics** - 布局统计信息
2. **Layout Tree (ASCII View)** - 树形结构图
3. **Layout Grid Visualization** - 80x24 网格可视化
4. **Node Details** - 节点详细信息

## Visualization Examples

### 1. Layout Statistics

```
Layout Statistics:
  Root Size: 80x30
  Root Position: (0, 0)
  Total Nodes: 43
  Max Depth: 6
```

显示：
- 根节点尺寸
- 根节点位置
- 总节点数
- 最大深度

### 2. Layout Tree (ASCII View)

使用树形结构显示布局层次：

```
├─ [Element] Element
│  Size: 80x30
│  Position: (0, 0)
│  ├─ [Element] Element
│  │  Size: 80x3
│  │  Position: (0, 0)
│  │  └─ [Text] Text
│  │      Size: 15x1
│  │      Position: (1, 1)
│  │      Children: 0 (leaf node)
```

**符号说明**:
- `├─` 分支
- `└─` 最后一个分支
- `[Type]` 节点类型
- `│` 垂直连接线
- `Size` 节点尺寸
- `Position` 节点位置
- `Children` 子节点数

**节点类型**:
- `Border` - 边框容器
- `VStack` - 垂直堆栈
- `HStack` - 水平堆栈
- `Flex` - Flex布局
- `Text` - 文本节点
- `Element` - 通用元素

### 3. Layout Grid Visualization

使用 80x24 网格显示节点分布：

```
  ┌────────────────────────────────────────────────────────────────────────────────┐
  │EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE│
  │ETTTTTTTTTTTTTTTETTTTTTTTTTTTETTTTTTTTTTTTTEEEEEEEEEEEEEEEEEEEEEEEEEEEEE│
  │EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE│
  │EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE│
  │... (更多行) ...│
  │EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE│
  └────────────────────────────────────────────────────────────────────────────────┘
```

**字符说明**：
- `E` = Element（通用元素）
- `V` = VStack（垂直堆栈）
- `H` = HStack（水平堆栈）
- `F` = Flex（Flex布局）
- `T` = Text（文本节点）
- `·` = 其他类型

**示例**:
```
  ┌────────────────────────────────────────────────────────────────────────────────┐
  │EEEEEEEEEEEEEE                                                              │
  │ETTTTTTTTTTTETTTTTTTTTTTTTTTEEEEEEEEEEEEEEEEEEEEEEEEEEEEE│
  │EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE│
  └────────────────────────────────────────────────────────────────────────────────┘

说明:
- 顶层容器 (E) 占据整个宽度 (80列)
- 第二层包含标题行 (T) 和其他内容
- 节点字符表示节点在网格中的位置
- 空格表示未使用的空间
```

### 4. Node Details

显示每个节点的详细信息：

```
[Node #1]
Type: Element
Size: 80x30
Position: (0, 0)
Bounds: (0, 0, 80, 30)
Children: 3
  [Node #2]
  Type: Element
  Size: 80x3
  Position: (0, 0)
  Bounds: (0, 0, 80, 3)
  Children: 1
```

**字段说明**:
- `[Node #N]` - 节点编号
- `Type` - 节点类型
- `Size` - 节点尺寸
- `Position` - 相对父节点位置
- `Bounds` - 绝对边界 (x, y, x+w, y+h)
- `Children` - 子节点数量

## Test Components

程序测试所有10个预定义组件：

| Fixture | Description | Nodes | Depth |
|---------|-------------|-------|-------|
| demo1_full_app | Complete Demo1 application | 43 | 6 |
| demo1_header | Header component | 6 | 3 |
| demo1_main_body | Main body with sidebar | 15 | 4 |
| demo1_modal | Confirmation modal | 16 | 3 |
| simple_vstack | Simple vertical stack | 4 | 2 |
| simple_hstack | Simple horizontal stack | 4 | 2 |
| nested_layout | Nested VStack inside HStack | 7 | 3 |
| bordered_content | Bordered container | 2 | 2 |
| flex_layout | HStack with flex items | 4 | 2 |
| keyed_items | VStack with keyed items | 4 | 2 |

## Key Features

### 1. 多视图可视化

提供4种不同的可视化方式：
- 统计信息
- 树形结构
- 网格分布
- 详细信息

### 2. 递归遍历

正确处理任意深度的节点树：
- 递归显示所有节点
- 正确的缩进和连接线
- 识别叶子节点

### 3. 网格填充

将节点位置映射到80x24网格：
- 显示节点分布
- 使用不同字符标记不同类型
- 正确处理节点重叠

### 4. 详细信息

每个节点包含完整信息：
- 类型
- 尺寸
- 位置
- 边界
- 子节点数

## Implementation Details

### 树形结构生成

使用递归生成树形ASCII图：
- 正确的缩进
- `├─` 表示分支
- `└─` 表示最后分支
- `│` 表示垂直连接

### 网格填充算法

```go
func fillGridWithNodes(box, grid, offsetX, offsetY) {
    // 计算节点在网格中的位置
    x := box.Box.X + offsetX
    y := box.Box.Y + offsetY
    
    // 填充节点区域
    for dy := 0; dy < h; dy++ {
        for dx := 0; dx < w; dx++ {
            grid[y+dy][x+dx] = marker
        }
    }
    
    // 递归填充子节点
    for _, child := range box.Children {
        fillGridWithNodes(child, grid, offsetX, offsetY)
    }
}
```

### 节点类型识别

```go
func getNodeType(box *ComputedBox) string {
    typeStr := box.VNode.Type().String()
    
    switch {
    case strings.Contains(typeStr, "border"):
        return "Border"
    case strings.Contains(typeStr, "vstack"):
        return "VStack"
    case strings.Contains(typeStr, "hstack"):
        return "HStack"
    // ... 更多类型
    }
}
```

## Usage Examples

### 1. 查看布局结构

树形视图清晰显示层级关系：
```
├─ [Element] Element
│  Size: 80x30
│  ├─ [Element] Element
│  │  Size: 80x3
│  │  └─ [Text] Text
│  │      Size: 15x1
```

### 2. 查看节点分布

网格视图显示节点在80x24空间中的分布：
```
  ┌────────────────────────────────────────────────────────────────────────────────┐
  │EEEEEEEEEEEEEE                                                              │
  │ETTTTTTTTTTTT                                                             │
  │EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE│
  └────────────────────────────────────────────────────────────────────────────────┘
```

### 3. 调试布局问题

详细信息帮助调试：
- 查看每个节点的精确位置
- 验证边界计算
- 检查节点重叠

## Files

- **Main Program**: `examples/fiber_demos/layout_ascii_visualization/main.go`
- **This README**: `examples/fiber_demos/layout_ascii_visualization/README.md`

## Related Documentation

- [Layout Integration Test Report](../../../docs/layout/INTEGRATION_TEST_REPORT.md)
- [Engine Comparison Report](../../../docs/layout/ENGINE_COMPARISON_REPORT.md)
- [Border Style Constraint Analysis](../../../docs/layout/BORDER_STYLE_CONSTRAINT_ANALYSIS.md)

## Benefits of ASCII Visualization

1. **快速理解布局结构**
   - 树形视图显示层级
   - 网格视图显示分布
   - 统计信息显示概况

2. **调试布局问题**
   - 精确位置和尺寸
   - 边界计算验证
   - 节点重叠检测

3. **文档和沟通**
   - 清晰的可视化
   - 易于分享
   - 适合文档使用

4. **性能分析**
   - 节点数量统计
   - 深度分析
   - 复杂度评估

## Limitations

1. **网格分辨率**
   - 使用字符网格（不是像素）
   - 80x24 固定尺寸
   - 可能不够精确

2. **节点重叠**
   - 使用字符标记，无法显示重叠
   - 重叠节点会被覆盖

3. **文本内容**
   - 不显示实际文本内容
   - 只显示节点类型和位置

## Future Enhancements

可能的改进：

1. **彩色输出**
   - 使用 ANSI 颜色代码
   - 不同类型不同颜色
   - 更清晰的可视化

2. **交互式视图**
   - 使用光标移动
   - 显示节点详情
   - 缩放和滚动

3. **内容预览**
   - 显示实际文本内容
   - 标记节点类型
   - 显示样式信息

4. **导出功能**
   - 导出为图像
   - 导出为 SVG
   - 导出为 JSON

## Running Individual Components

可以修改程序只测试特定组件：

```go
// 只测试一个组件
fixture := component_fixtures.GetFixture("demo1_header")
// ... 其余代码
```

或修改fixtures列表：

```go
// 修改 main.go 中的测试循环
fixtures := []struct{
    component_fixtures.GetFixture("simple_vstack"),
    component_fixtures.GetFixture("bordered_content"),
}
```

## Conclusion

ASCII可视化工具提供了：
- ✅ 清晰的布局结构视图
- ✅ 直观的节点分布显示
- ✅ 详细的节点信息
- ✅ 多种可视化方式

这个工具非常适合：
- 调试布局问题
- 理解布局结构
- 验证布局计算
- 文档和演示

---

**Demo Date**: 2025-02-15
**Components Tested**: 10/10
**Visualizations**: 4 per component
**Status**: ✅ All working perfectly
