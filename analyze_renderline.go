# RenderLine处理continuation bug分析

## 测试输出分析

```
Frame 2 back buffer (SetString写入"Button-"):
  [12]: "→" w=2 (旧箭头未覆盖)
  [13]: continuation (旧continuation未覆盖)

Frame 2 front buffer (上一帧):
  [12]: "→" w=2
  [13]: continuation

Diff检测: [12]相同，[13]相同

Frame 2 output: "\x1b[4;12H \x1b[0m"
```

## 问题：为什么输出空格到位置12？

如果diff检测显示无变化，renderLine不应该输出任何内容。

**可能的解释**：应用层在绘制Frame 2之前，主动清除了某些区域。

## 真正的问题

即使renderLine输出了空格到位置12，箭头是宽度2：
- 位置12输出1个空格 → 清除箭头的"头"
- 位置13的continuation没有被输出 → **continuation残留**

## 根本原因

### renderLine输出空格时的问题

```go
if cell.Cluster == "" {
    x++
    continue
}
```

当检测到cell为空时，跳过输出。但这个空格是手动输出的：

```go
// 某个地方输出了空格到位置12
r.emitRunWithWidth(12, 3, style, " ", 1)
```

注意：**输出的是1个空格**，但箭头宽度是2！

## 正确的修复

在renderLine中，当输出空格清除内容时，需要检查前一个cell是否是宽字符头：

```go
// 在emitRunWithWidth之前检查
if cell.Cluster == "" && prevCell.Width == 2 {
    // 输出2个空格清除整个宽字符
    r.emitRunWithWidth(x, y, cell.Style, " ", 2)
} else {
    // 正常输出
    r.emitRunWithWidth(x, y, runStyle, runText, totalWidth)
}
```

但更深层的问题：**谁在调用emitRunWithWidth输出空格？**

需要找到输出"\x1b[4;12H "的代码位置。
