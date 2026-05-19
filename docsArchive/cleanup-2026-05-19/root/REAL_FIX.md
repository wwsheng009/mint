# 正确的修复方案

## 问题

`emitRunWithWidth(x, y, style, " ", 2)` 只输出1个空格，但需要清除宽度为2的宽字符。

## 修复代码

在 `runtime/paint/renderer.go` 中，修改 `emitRunWithWidth` 的调用：

```go
// 错误的做法（当前）：
r.emitRunWithWidth(headX, y, cell.Style, " ", headCell.Width)
// 即使headCell.Width=2，也只输出1个空格

// 正确的做法：
clearText := strings.Repeat(" ", headCell.Width)
r.emitRunWithWidth(headX, y, cell.Style, clearText, headCell.Width)
// 输出headCell.Width个空格

// 或者更简单的：
spaces := make([]byte, headCell.Width)
fill(spaces, ' ')
r.emitRunWithWidth(headX, y, cell.Style, string(spaces), headCell.Width)
```

## 实际修改位置

文件：`runtime/paint/renderer.go`，约第178行

修改前：
```go
r.emitRunWithWidth(headX, y, cell.Style, " ", headCell.Width)
```

修改后：
```go
// 生成足够数量的空格来清除整个宽字符
clearText := strings.Repeat(" ", headCell.Width)
r.emitRunWithWidth(headX, y, cell.Style, clearText, headCell.Width)
```

同时需要在 `import` 中添加 `"strings"` 包。
