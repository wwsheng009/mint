# Grid Cell 边框设计文档

## 1. API 设计

### VNode API

```go
// Grid VNode 新增方法

// 设置是否显示 cell 边框
func (g *VNode) SetShowCellBorders(show bool) *VNode {
    g.showCellBorders = show
    return g
}

// 设置 cell 边框样式
func (g *VNode) SetCellBorderStyle(style string) *VNode {
    g.cellBorderStyle = style // "none", "single", "double", "light"
    return g
}

// 设置 cell 边框是否带圆角
func (g *VNode) SetCellBorderRounded(rounded bool) *VNode {
    g.cellBorderRounded = rounded
    return g
}

// 设置 cell 边框颜色
func (g *VNode) SetCellBorderColor(color string) *VNode {
    g.cellBorderColor = color
    return g
}

// 快捷方法
func (g *VNode) ShowCellBorders() *VNode {
    return g.SetShowCellBorders(true)
}

func (g *VNode) HideCellBorders() *VNode {
    return g.SetShowCellBorders(false)
}

func (g *VNode) SingleCellBorders() *VNode {
    return g.SetShowCellBorders(true).SetCellBorderStyle("single")
}

func (g *VNode) DoubleCellBorders() *VNode {
    return g.SetShowCellBorders(true).SetCellBorderStyle("double")
}

func (g *VNode) LightCellBorders() *VNode {
    return g.SetShowCellBorders(true).SetCellBorderStyle("light")
}
```

## 2. 属性设计

### VNode 属性

```go
type VNode struct {
    // ... 现有属性

    // ✨ Cell Borders (新增)
    showCellBorders   bool   // 是否显示格子边框
    cellBorderStyle   string // 边框样式: "none", "single", "double", "light"
    cellBorderRounded bool   // cell 边框是否带圆角
    cellBorderColor   string // 边框颜色
}
```

### Instance 属性

```go
type Instance struct {
    // ... 现有属性

    // ✨ Cell Borders (新增)
    showCellBorders   bool
    cellBorderStyle   string
    cellBorderRounded bool
    cellBorderColor   string
}
```

## 3. 边框样式定义

```go
// Cell 边框样式
const (
    CellBorderStyleNone   = "none"
    CellBorderStyleSingle = "single"
    CellBorderStyleDouble = "double"
    CellBorderStyleLight  = "light"   // 轻边框 (│ ─)
)

// 边框字符集
var cellBorderCharacters = map[string]BorderChars{
    "single": {
        top:         "─",
        bottom:      "─",
        left:        "│",
        right:       "│",
        topLeft:     "┌",
        topRight:    "┐",
        bottomLeft:  "└",
        bottomRight: "┘",
        cross:       "┼",
        topCross:    "┬",
        bottomCross: "┴",
        leftCross:   "├",
        rightCross:  "┤",
    },
    "double": {
        top:         "═",
        bottom:      "═",
        left:        "║",
        right:       "║",
        topLeft:     "╔",
        topRight:    "╗",
        bottomLeft:  "╚",
        bottomRight: "╝",
        cross:       "╬",
        topCross:    "╦",
        bottomCross: "╩",
        leftCross:   "╠",
        rightCross:  "╣",
    },
    "light": {
        top:         "─",
        bottom:      "─",
        left:        "│",
        right:       "│",
        topLeft:     "┌",
        topRight:    "┐",
        bottomLeft:  "└",
        bottomRight: "┘",
        cross:       "┼",
        topCross:    "┬",
        bottomCross: "┴",
        leftCross:   "├",
        rightCross:  "┤",
    },
}

type BorderChars struct {
    top, bottom, left, right string
    topLeft, topRight string
    bottomLeft, bottomRight string
    cross string
    topCross, bottomCross string
    leftCross, rightCross string
}
```

## 4. 边框绘制逻辑

### 绘制策略

在 Paint 方法中：
1. 先绘制 cell 内容（子节点）
2. 然后在 cell 边界位置绘制边框字符
3. 使用正确的边框字符（考虑圆角、交点）

### 边框位置计算

```go
// 对于 cell (row, col):
// - 左边框: col * (cellWidth + gap) + padding[3]
// - 上边框: row * (cellHeight + gap) + padding[0]
// - 每个边框占用 1 字符宽度
```

### 边框绘制步骤

```pseudo
for each row in 0..numRows-1:
    for each col in 0..numCols-1:
        cell = getCell(row, col)
        if cell.showBorder:
            # 绘制四条边
            drawTopBorder(row, col)
            drawBottomBorder(row, col)
            drawLeftBorder(row, col)
            drawRightBorder(row, col)

            # 绘制四个角
            drawTopLeft(row, col)
            drawTopRight(row, col)
            drawBottomLeft(row, col)
            drawBottomRight(row, col)

            # 绘制交点（与相邻 cell 共享）
            drawCrossPoints(row, col)
```

## 5. 尺寸计算

### 边框占用

- **单格边框**：每个 cell 四周各 1 字符
- **但相邻 cell 共享边框**：总的边框占用 = (cols+1) + (rows+1) 字符

### 新的尺寸计算

```go
// totalWidth = padding[1] + padding[3]
//             + colWidths (内容宽度)
//             + columnGaps
//             + (numCols - 0) * borderWidth  // 右边框

// totalHeight = padding[0] + padding[2]
//              + rowHeights (内容高度)
//              + rowGaps
//              + (numRows - 0) * borderWidth  // 下边框
```

## 6. 使用示例

```go
// 基本用法
grid.New().
    SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
    SetRows(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
    ShowCellBorders().  // 单线边框
    SetChildrenAuto([]ui.VNode{
        text.New("A1"), text.New("A2"),
        text.New("B1"), text.New("B2"),
    })

// 双线边框
grid.New().
    SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
    DoubleCellBorders().  // 双线边框
    SetChildrenAuto(...)

// 轻边框 + 颜色
grid.New().
    LightCellBorders().
    SetCellBorderColor("cyan").
    SetChildrenAuto(...)

// 混合：容器边框 + cell 边框
grid.New().
    SingleBorder("Table").      // 容器单线边框
    DoubleCellBorders().        // cell 双线边框
    SetChildrenAuto(...)

// 表格样式（表头双线，数据单线）
stack.NewVStack().
    SetChildrenList([]ui.VNode{
        grid.New().
            DoubleCellBorders().  // 表头
            SetChildrenAuto(headers),
        grid.New().
            SingleCellBorders().  // 数据行
            SetChildrenAuto(row1),
        grid.New().
            SingleCellBorders().  // 数据行
            SetChildrenAuto(row2),
    })
```

## 7. 渲染效果示例

```
┌────────────────────────────────────┐
│        Table                        │  ← 容器边框 (Single)
├────────────────────────────────────┤
│┌───────┬───────┬───────┐          │
││ Name  │ Age   │ Role  │          │  ← Cell 边框 (Double)
│├───────┼───────┼───────┤          │
││ Alice │ 30    │ Admin │          │
│├───────┼───────┼───────┤          │
││ Bob   │ 25    │ User  │          │
│└───────┴───────┴───────┘          │
└────────────────────────────────────┘
```

## 8. 实现优先级

Phase 1: 单线边框 (single)
Phase 2: 双线边框 (double)
Phase 3: 轻边框 (light)
Phase 4: 圆角边框
Phase 5: 边框颜色自定义

## 9. 注意事项

1. **边框重叠处理**：相邻 cell 的边框要正确绘制，避免重复
2. **交点字符**：使用正确的交叉字符（┼ ╬）
3. **内容偏移**：cell 内容需要向内偏移 1 字符（边框宽度）
4. **Gap 兼容性**：cell 边框与 gap 共同使用时正确计算位置
5. **容器边框优先级**：容器边框在最外层，cell 边框在内部
