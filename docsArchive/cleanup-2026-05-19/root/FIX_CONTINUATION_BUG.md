# 宽字符Continuation Cell清除Bug修复方案

## 问题描述

当宽字符（宽度2）被短字符（宽度1）或空字符覆盖时，continuation cell未被清除，导致残留。

示例：
```
Frame 1: "Button →"  →  back buffer[12]="→"(w=2), [13]=continuation
Frame 2: "Button-"   →  back buffer[12]="-"(w=1), [13]=""(空)

期望结果：位置13的continuation被清除
实际结果：位置13的continuation残留，终端显示残影
```

## 根本原因

### renderLine的处理逻辑问题

位置13的处理流程：
1. `IsCellChanged(cell, prevCell)` 检测到变化：
   - `prevCell.IsContinuation = true` (旧箭头的continuation)
   - `cell.IsContinuation = false` (新内容)
   - 返回 `true` → 需要渲染

2. `renderLine` 检查：
   ```go
   if cell.IsContinuation {
       x++
       continue
   }
   ```
   由于当前cell不是continuation，继续处理

3. 检查是否为空cluster：
   ```go
   if cell.Cluster == "" || cell.Cluster == "\\x00" {
       x++
       continue
   }
   ```
   cell为空，跳过输出 → **问题！没有发出清除指令**

4. 结果：终端位置13的旧内容被保留

### IsCellChanged的逻辑

```go
if prevCell.IsContinuation {
    return true  // 检测到变化
}
```
检测到了变化，但renderLine没有处理这个变化。

## 修复方案

在 `renderLine` 中，当检测到 `prevCell.IsContinuation` 且当前cell为空时，需要：
1. 往回找到宽字符的头
2. 输出足够数量的空格清除整个宽字符区域

### 代码修改位置

文件：`runtime/paint/renderer.go`
位置：约第165-175行，在 `if cell.IsContinuation` 之后

### 修复代码

```go
// 如果是延续单元格，跳过（由主单元格处理）
if cell.IsContinuation {
    x++
    continue
}

// CRITICAL FIX: 如果前一个单元格是宽字符的continuation，需要用空格清除它
// 这发生在短字符覆盖宽字符时（如"Button-"覆盖"Button →"）
if prevCell.IsContinuation && cell.Cluster == "" {
    // 往回找到宽字符的头
    headX := x - 1
    if headX >= 0 {
        headCell := r.front.Cells[y][headX]
        // 检查是否是宽字符头
        if headCell.Width == 2 && !headCell.IsContinuation {
            // 输出空格清除整个宽字符（头+continuation）
            r.emitRunWithWidth(headX, y, cell.Style, " ", headCell.Width)
            // x已经处理到continuation位置，继续前进
        }
    }
    x++
    continue
}

// 跳过空 cluster（避免无限循环和无效输出）
if cell.Cluster == "" || cell.Cluster == "\\x00" {
    x++
    continue
}
```

### 修复说明

1. **检测条件**：`prevCell.IsContinuation && cell.Cluster == ""`
   - 旧cell是continuation（表示这里是宽字符的一部分）
   - 新cell为空（表示宽字符被清除）

2. **查找头**：`headX := x - 1`
   - continuation的前一个单元格一定是宽字符的头

3. **清除整个宽字符**：
   - 位置headX输出2个空格（对于宽度2的字符）
   - 这样同时清除头和continuation两个位置的终端显示

4. **x更新**：`x++`
   - 正常前进到下一个位置

## 测试验证

修复后运行测试：
```bash
go run ./tests/arrow_width_bug.go
```

期望结果：
- Frame 2 output包含清除指令
- front buffer中没有残留的箭头或continuation cell

## 注意事项

1. 宽字符可能是宽度3（某些emoji），代码中使用了 `headCell.Width` 来支持任意宽度
2. 清除操作使用 `emitRunWithWidth` 确保光标位置正确
3. 使用 `cell.Style` 确保清除后的样式一致
