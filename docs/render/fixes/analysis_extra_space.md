分析：为什么会出现额外的空格

# 问题场景
```text
• Mouse Click: Click an input/button to focus
```
显示为：
```text
•  Mouse Click: Click an input/button to focus
    ^^多一个空格
```

# 可能原因分析

## 原因1: run merging 循环逻辑错误

在新的 run merging 逻辑中：

```go
// 收集当前单元格以及后续相邻的、同样式的单元格
for ; x < endX; {
    nextCell := r.back.Cells[y][x]
    nextPrevCell := r.front.Cells[y][x]

    // 如果单元格不是 changed 或者样式不同，停止合并
    if !IsCellChanged(nextCell, nextPrevCell) ||
        nextCell.IsContinuation ||
        nextCell.Style != runStyle ||
        nextCell.Cluster == "" ||
        nextCell.Cluster == "\\x00" {
        break
    }
    ...
}
```

这个循环的问题：
- 在内层循环中，`nextPrevCell` 被用来检查
- 但第一次调用时，我们已经知道 `cell` 是 changed 的
- 所以内层循环应该从第一个单元格开始合并，而不是检查 IsCellChanged

## 原因2: emitRunWithWidth 中的 textWidth 计算问题

新逻辑使用 `totalWidth` 作为 textWidth：
```go
r.emitRunWithWidth(startX, y, runStyle, runText.String(), totalWidth)
```

但是 `runText.String()` 的字符宽度可能与 `totalWidth` 不一致：
- 例如：runText = "• "，totalWidth = 2 + 1 = 3
- 但 runewidth.StringWidth("• ") = 2 + 1 = 3
- 这是一致的

## 原因3: 内层循环的 IsCellChanged 检查位置不当

问题可能在这里：
```go
if !IsCellChanged(nextCell, nextPrevCell) || ... {
    break
}
```

- 第一次循环时，`nextCell = cell`, `nextPrevCell = prevCell`
- 我们已经知道 `IsCellChanged(cell, prevCell) == true`（否则不会进入这个分支）
- 所以第一次检查应该通过
- 但如果后续检查失败，循环会 break，这可能导致部分合并

## 根本问题：内层循环的起始位置

正确的逻辑应该是：
1. 外层循环找到第一个 changed 单元格
2. 开始内层循环，从当前 x 位置开始收集**所有相邻的、同样式的单元格**（不管是否 changed）
3. 输出合并后的 run

而不是：
1. 外层循环找到第一个 changed 单元格
2. 开始内层循环，从当前 x 位置开始收集**所有相邻的、同样式的、changed 的单元格**
3. 输出合并后的 run

因为一旦我们决定输出一个区域，我们应该输出整个区域的完整内容，而不是只输出 changed 的部分。

# 修复方案

方案1：只合并空格字符（最简单，风险最小）
```
runText.WriteString(cell.Cluster)
// 如果当前单元格是空格，尝试合并后续的相邻空格
if cell.Cluster == " " && cell.Style == runStyle {
    for lookAhead := x + cell.Width; lookAhead < endX; {
        lookAheadCell := r.back.Cells[y][lookAhead]
        if lookAheadCell.Cluster == " " && lookAheadCell.Style == runStyle &&
           !lookAheadCell.IsContinuation {
            // 合并这个空格
            // 需要检查是否真的需要清除
        }
    }
}
```

方案2：回到原始的逐单元格输出（最稳定）
```
// 逐个单元格输出，不合并
runText.WriteString(cell.Cluster)
width := cell.Width
if width <= 0 { width = 1 }
x += width
r.emitRunWithWidth(startX, y, runStyle, runText.String(), width)
```

建议：先使用方案2确保稳定性，如果有性能需求再考虑方案1。
