# Grid Cell 边框设计文档（修复版）

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
        bottomRight:  "┘",
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

### 4.1 绘制策略

```
Paint 方法只负责边框绘制（不负责内容）：
1. Instance.Paint(originX, originY) 返回 Cell 边框的绘制命令
2. 子节点内容由渲染引擎独立绘制（递归调用子节点的 Paint()）
3. 边框和内容的绘制顺序：由渲染引擎决定，通常后绘制的覆盖先绘制的
```

### 4.2 边框坐标系统

```
坐标系统定义：
┌─────────────────────────────────────────┐
│ y=originY                                    │ ← contentY (内容区域起点，包含左边框)
│ y=originY+1                                  │ ← 第 0 行内容区域起点（跳过上边框）
│ y=originY+rowHeights[0]                      │ ← 第 0 行内容区域终点
│ y=originY+rowHeights[0]+1                    │ ← 第 0 行底边框/中间分隔线
│                                              │
│ y=originY+rowHeights[0]+1                    │ ← 第 1 行内容区域起点
│ y=originY+rowHeights[0]+1+rowHeights[1]          │ ← 第 1 行内容区域终点
│ y=originY+rowHeightTotal+borderCharHeight       │ ← 底边框 (y = originY + ΣrowHeights + (numRows+1))
└─────────────────────────────────────────┘

坐标系转换：
- 相对坐标：getRowPosition(row, col) 返回相对于 contentY 的坐标（已跳过边框）
- 绝对坐标：渲染引擎需要将相对坐标转换为绝对坐标
- 边框绘制从 contentY 开始，使用 rowHeights 和 borderChars 计算
```

### 4.3 边框绘制步骤

```
伪代码：
func GenCellBorderDrawCmds(originX, originY int) []DrawCmd:
    contentX = originX + inst.padding[3]  // 加左 padding
    contentY = originY + inst.padding[0]  // 加上 padding

    for row = 0 to numRows:
        // 计算当前行边框的 Y 坐标
        lineY = contentY
        for r = 0 to row-1:
            lineY += inst.rowHeights[r] + 1  // 累加内容高度 + 下边框
        end
        
        // 绘制该行的边框线（交点 + 水平线 + 垂直线）
        for col = 0 to numCols:
            // 计算 X 坐标
            lineX = contentX
            for c = 0 to col-1:
                lineX += inst.colWidths[c] + 1
            end
            
            // 绘制交点字符
            char = selectBorderChar(row, col, numRows, numCols)
            cmds.append(DrawCmd{x: lineX, y: lineY, text: char})
            
            // 绘制水平边框（除了交点）
            if col < numCols:
                drawHorizontalBorder(lineX, col, lineY, cmds)
            end
        end
        
        // 绘制垂直边框（每列）
        for col = 0 to numCols:
            drawVerticalBorder(col, lineY, cmds)
        end
    end
```

## 5. 尺寸计算设计

### 5.1 边框占用空间计算

```
┌────────────────────────────────────────────────┐
│ 🔵 Cell Borders 占用空间                            │
└────────────────────────────────────────────────┘

垂直边框：每条线占 1 字符
- 垂直线数量 = numCols + 1
- 垂直边框总宽 = (numCols + 1) * 1 = numCols + 1

水平边框：每条线占 1 字符
- 水平线数量 = numRows + 1
- 水平边框总高 = (numRows + 1) * 1 = numRows + 1
```

### 5.2 总尺寸计算公式

#### 5.2.1 无 Cell 边框

```
TotalWidth  = padding[1] + padding[3]
             + Σ(colWidths) 
             + Σ(columnGaps)

TotalHeight = padding[0] + padding[2]
              + Σ(rowHeights)
              + Σ(rowGaps)
```

#### 5.2.2 有 Cell 边框

```
TotalWidth  = padding[1] + padding[3]
             + Σ(colWidths)
             + (numCols + 1)  ← 垂直边框占用

TotalHeight = padding[0] + padding[2]
              + Σ(rowHeights)
              + (numRows + 1)  ← 水平边框占用
```

### 5.3 容器边框 + Cell 边框

```
┌─────────┐  ← 容器上边框 (1 行)
│ Table   │  ← 容器标题栏（1 行）
├─────────┤  ← 容器分隔线（1 行）

┌──────────┼──────────┼──────────┐  ← Cell 上边框 (contentY)
│   ID     │ Name     │ Role     │    ← Row 0 内容
├══════════╪══════════╪════════╣      ← Row 0 底边框/行间分隔
│   001    │ Alice    │ Admin    │    ← Row 1 内容
└──────────┴──────────┴──────────┘      ← Cell 底边框
├────────────────────────────────────┤  ← 容器下边框
└────────────────────────────────────┘

TotalHeight 计算公式：
TotalHeight = containerTopBorder     // 容器上边框
             + containerTitleBar      // 容器标题栏
             + containerTopPadding
             + Σ(rowHeights)           // 内容高度
             + (numRows + 1) * 1       // Cell 水平边框
             + containerBottomPadding
             + containerBottomBorder
```

### 5.4 可用高度计算

```
AvailableHeight = constraint.MaxHeight - containerTop - containerTopPadding - containerBottomPadding

// 分配给 Row 和 Cell Borders 的可用高度
contentAvailableH = AvailableHeight - containerTitleBar - (numRows + 1) * 1

// RowHeights 必须满足：
Σ(rowHeights) ≤ contentAvailableH
```

## 6. Auto 行高设计（修复核心缺陷）

### 6.1 设计原则

**问题背景：**
- Auto 行高的最小值为 1 时，Cell 边框需要额外的空间
- 当 rowHeights = [1, 1] 且 showCellBorders=true 时：
  - 边框需要 3 行（上边框+分隔线+下边框）
  - 内容只有 2 行（每行 1 字符）
  - 导致内容和边框渲染位置冲突

### 6.2 Auto 行高最小值计算

```go
// ✅ 修复后的 Auto 行高设计

case Auto:
    minContentHeight := 1  // 每行内容至少 1 字符
    heights[i] = minContentHeight
```

### 6.3 可用高度不足时的分配策略

**情况 1：可用高度足够**
```
contentAvailableH = 100
numRows = 2
borderCharsH = 3

// 可用内容高度 = 100 - 3 = 97
// 97 > 2 * 1 = 2
// 策式：正常分配，每行高度 = 97 / 2 = 48
// ✓ OK
```

**情况 2：可用高度刚好**
```
contentAvailableH = 5
numRows = 2
borderCharsH = 3

// 可用内容高度 = 5 - 3 = 2
// 每行高度 = 2 / 2 = 1
// ✓ OK（刚好够用）
```

**情况 3：可用高度不足**
```
contentAvailableH = 2
numRows = 2
borderCharsH = 3

// 可用内容高度 = 2 - 3 = -1  ← 不足！
// 策略：压缩可用空间
adjustedAvailableH = max(MinContentSize, constraint.MaxHeight - borderCharsH - containerPadding)
```

### 6.4 最小内容尺寸（MinContentSize）

```go
const (
    MinGridContentSize = 1  // Grid 可用空间的最小值
 MinGridSizeWithBorders = MinGridContentSize + 4  // 边框至少需要 4 行
)
```

## 7. Paint() 方法契约（明确修复）

### 7.1 方法签名

```go
// Instance.Paint 绘制 Cell 边框
func (inst *Instance) Paint(originX, originY int) []paint.DrawCmd

// 参数定义：
//   - originX, originY: Grid 组件的绝对位置（包含容器边框和 padding）
//   - 内容区域起点：（通过 padding 机制获得自动处理）
//
// 返回值：
//   - Cell 边框的绘制命令列表（不包含子节点内容）
//
// 坐标计算：
//   contentX = originX                      （容器边框内起点）
//   contentY = originY                      （容器边框内起点）
//   Cell 边框绘制从 contentX, contentY 开始
//
//   边框绘制范围：
//   - X: [contentX, contentX + TotalWidth - 1]
//   - Y: [contentY, contentY + TotalHeight - 1]
//
//   其中 TotalWidth = Σ(colWidths) + (numCols + 1)
//         TotalHeight = Σ(rowHeights) + (numRows + 1)
```

### 7.2 调用上下文

```
渲染引擎流程：
┌─────────────────────────────────────────────┐
│ 1. LayoutEngine 调用 Instance.Paint()         │
│    inst.Paint(box.X, box.Y)                     │
│    ↓                                         │
│    box.X, box.Y 是 LayoutBox 的绝对坐标        │
│    （已在布局引擎中包含了所有偏移）             │
│    ↓                                         │
│ 2. 渲染引擎处理 Instance.Paint() 返回的边框命令│
│    buffer.SetString(cmd.X, cmd.Y, cmd.Text, ...)    │
│    ↓                                         │
│ 3. 渲染引擎递归绘制子节点                    │
│    child.Paint(childAbsX, childAbsY, ...)       │
│    child.AbsX = box.X + (从 layoutBox.X 转换) │
└─────────────────────────────────────────────┘
```

### 7.3 参数说明

**originX, originY 的含义：**
- 这就是 Grid 组件在父容器中的**绝对位置**
- 传递给 `Instance.Paint()` 时，参数就是 `LayoutBox.X, LayoutBox.Y`
- 内部已经包含了：
  - 父组件的绝对偏移
  - Grid 组件自己的 padding
  - Grid 组件的容器边框（如果有）

**正确使用示例：**
```go
// 在 Instance.Paint() 中：
contentX := originX                    // ✅ 直接使用，不要再加 padding
contentY := originY                    // ✅ 已经包含了所有偏移

// ❌ 错误用法：
contentX := originX + inst.padding[3]  // ❌ 重复计算，导致偏移错误
contentY := originY + inst.padding[0]
```

## 8. 边界条件设计

### 8.1 最小 Grid 尺寸

```go
const (
    // 无 Cell 边框
    MinGridSizeNoBorders = MinContentSize + padding
    
    // 有 Cell 边框
    MinGridSizeWithBorders = MinContentSize 
                                    + (minRows + 1)  // 水平边框
                                    + padding
```

### 8.2 可用高度不足时的处理

```
if availableHeight < MinGridSizeWithBorders:
    // 策略 1：返回最小尺寸
    return Size{Width: MinGridSizeWithBorders, Height: MinGridSizeWithBorders}
    
    // 策略 2：按比例缩放 rowHeights（优先保证最小内容）
    scale := float64(MinGridSizeWithBorders) / float64(availableHeight)
    if scale > 1.0 {
        scale = 1.0  // 最多使用全部可用高度
    }
    // 缩放 rowHeights...
```

### 8.3 零值空间处理

```
if numCols == 0:
    numCols = 1
    colWidths = [availableWidth]
    
if numRows == 0:
    numRows = 1
    // 列表行数
```

## 9. 使用示例

### 9.1 基本用法

```go
// 单线边框
grid.New().
    SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
    SetRows(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
    ShowCellBorders().
    SetChildrenAuto([]ui.VNode{
        text.New("A1"), text.New("A2"),
        text.New("B1"), text.New("B2"),
    })

// 双线边框
grid.New().
    SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
    SetRows(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
    DoubleCellBorders().
    SetChildrenAuto(...)

// 轻边框 + 自定义颜色
grid.New().
    LightCellBorders().
    SetCellBorderColor("cyan").
    SetChildrenAuto(...)

// 混合：容器边框 + cell 边框
grid.New().
    SingleBorder("表格").         // 容器单线边框
    SetCellBorderStyle("double"). // cell 双线边框
    ShowCellBorders().
    SetChildrenAuto(...)
```

### 9.2 Auto 行高的正确使用

```go
// ✅ 正确：给 Auto 足够的空间
grid.New().
    SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
    SetRows(
        grid.Auto{},  // 会自动分配足够空间
        grid.Auto{},
    ).
    ShowCellBorders().
    SetConstraints(layout.BoxConstraints{
        MinWidth:  10,
        MinHeight: 20,  // 提供足够高度给 2 行 + 边框
    })

// ❌ 错误：Auto 行高太小
grid.New().
    SetRows(grid.Auto{}, grid.Auto{}).
    ShowCellBorders().
    SetConstraints(layout.BoxConstraints{
        MinHeight: 3,  // 不足以容纳 2 行 x 1 + 边框 3 = 5
    })
```

### 9.3 表格样式示例（容器边框 + Cell 边框）

```go
stack.NewVStack().
    SetGap(1).
    AddChild(text.New("表格样式").Bold(true)).
    AddChild(
        grid.New().
            SetColumns(grid.Fixed(10), grid.Fixed(10), grid.Fixed(10)).
            SetRows(grid.Auto{}, grid.Auto{}).
            SingleBorder("Users").      // 容器边框
            SetCellBorderStyle("single").  // Cell 边框样式
            ShowCellBorders().
            SetConstraints(layout.BoxConstraints{
                MinWidth: 30,
                MinHeight: 10,  // 确保有足够空间给边框和内容
            }).
            SetChildrenAuto([]ui.VNode{
                text.New("ID"), text.New("Name"), text.New("Role"),
                text.New("001"), text.New("Alice"), text.New("Admin"),
            }),
    ).
    AddChild(
        grid.New().
            ShowCellBorders().
            SetChildrenAuto([]ui.VNode{
                text.New("001"), text.New("Bob"), text.New("User"),
            }),
    )
```

## 10. 实现优先级

### Phase 1 🔴 核心修复
1. ✅ 修复 Auto 行高最小值设计（添加 minContentHeight = 1）
2. ✅ 明确 Paint() 方法契约（定义 originX, originY 的含义）
3. ✅ 确保边框绘制逻辑与 getCellPosition() 坐标计算一致

### Phase 2 🟡 完善设计
1. 容器边框 + Cell 边框叠加设计
2. 边界条件和 MinGridSize 设计
3. Auto 行高在可用高度不足时的分配策略

### Phase 3 🟢 增强
1. Cell Span 的边框绘制规则
2. 混合样式的自动应用
3. 压缩和溢出处理策略

## 11. 注意事项

### 11.1 边缝处理
```
- ✅ 边框字符正确选择（考虑交点、角）
- ✅ 相邻 cell 的边框共享（避免重复绘制）
- ✅ Cell Span 时边框字符正确跳过跨越的列/行
```

### 11.2 坐标一致性
```
- ⚠️ 关键：边框绘制的 Y 坐标必须与 getCellPosition() 的 Y 坐标计算完全一致
- 边框线的 Y 坐标 = contentY + 前面内容高度 + 前面下边框
- 内容区域的 Y 坐标 = contentY + 前面上边框（已跳过）
```

### 11.3 绘制顺序
```
当前设计：
- 边框和内容由渲染引擎独立绘制
- 绘制顺序由渲染引擎决定
- 正确的顺序：先或后，都需要绝对位置正确

⚠️ 注意：不能假设边框一定先于或后于内容绘制
```

### 11.4 性能考虑
```
- ✅ 边框绘制只在 showCellBorders=true 时执行
- ✅ 边框字符预定义在 map 中，避免运行时计算
- ⚠️ 边框绘制是 O(numCols * numRows)，对于大 Grid 可能需要优化
```

## 12. 与其他组件的兼容性

### 12.1 与 Container 边框的兼容
```
- 容器边框（SingleBorder）绘制在 Grid 最外层
- Cell 边框绘制在 Grid 内部（包含容器边框之后）
- 两者不会冲突：容器边框在外围，Cell 边框在内部
```

### 12.2 与 Padding 的兼容
```
-Padding 与 Cell 边框是独立的
- Padding 影响内容区域但不影响边框线绘制
边框绘制从 contentX = originX + padding[3] 开始
子节点也从 content区域开始绘制
```

### 12.3 与 Gap 的兼容
```
- ColumnGap 仅影响列之间的间距
- RowGap 仅影响行之间的间距
- 边框字符不与 Gap 重叠
- Gap 空间不参与边框字符的绘制
```

---

## 附录：关键设计决策记录

### A. 为什么 Auto 行高最小值是 1？

**决策：** Auto 行高的最小内容高度 = 1

**理由：**
1. 文本内容至少需要 1 个字符
2. 边框字符单独占用空间（垂直线和水平线）
3. 分离设计：内容高度和边框高度独立计算
4. 符合直觉：用户设置 Auto 时就是希望内容自适应，不应该为 0

### B. 为什么 Paint() 参数是 Grid 的绝对位置？

**决策：** Paint(x, y) 接收 Grid 的绝对位置

**理由：**
1. 消除歧义：明确参数含义，避免调用者猜测是否需要加 padding
2. 简化代码：内部直接使用参数，无需再次添加偏移
3. 一致性：与其他组件的 Paint() 方法保持相同行为

### C. 边框的 Y 坐标计算公式

**决策：**
```
边框线的 Y 坐标 = contentY + 前面内容高度累积 + 行号（从0 到 row）
内容区域的 Y 坐标 = contentY + 前面上边框（第 row 行的起点的 y）
```

**理由：**
1. 确保边框线和分隔线正确对齐
2. 让内容的 Y 坐标独立于边框（已跳过上边框）
3. 公式简单清晰，易于理解和维护

---

**文档版本：** v2.0 (修复版)  
**修复日期：** 2025-02-25  
**修复内容：** Auto 行高最小值、Paint() 契约、容器边框叠加设计
