# 宽字符渲染问题修复文档

## 问题描述

在 TUI 应用中，当出现宽字符（如 "•"，U+2022 BULLET，宽度为 2）后，如果内容缩小（如提交表单），终端会出现以下异常：

### 问题现象

1. **散落字符残留**：提交表单后，终端显示散乱的字符，无法正确清除
2. **额外空格**：当"•"后面跟着空格时，会输出两个空格而不是一个

示例：
```
预期：• Mouse Click: Click an input/button to focus
异常：•  Mouse Click: Click an input/button to focus  ← 两个空格

预期：• Backspace
异常：•  Delete character  ← 额外的空格
```

## 根本原因分析

### 问题1：宽字符的特性

"•" (U+2022 BULLET) 是一个宽字符：
- `runewidth.RuneWidth('•') = 2`（占用 2 个显示单元格）
- 在 buffer 中占用两个位置：x 和 x+1
- Cell 结构：
  - `Cells[x].Cluster = "•"`, `Width = 2`, `IsContinuation = false`
  - `Cells[x+1].Cluster = ""`, `IsContinuation = true`

### 问题2：emitRunWithWidth 的 cursorX 不同步

在 `runtime/paint/renderer.go` 中：

```go
func (r *Renderer) emitRunWithWidth(x, y int, runStyle style.Style, text string, textWidth int) {
    if text == "" {
        return  // ❌ 问题：直接返回，但没有更新 cursorX
    }

    // ... 输出文本 ...

    r.cursorX = x + textWidth
}
```

**问题**：当 `text` 为空字符串时，函数直接返回，但没有更新 `cursorX`。这导致：
- 光标位置和实际输出位置不同步
- 后续的光标移动计算错误

### 问题3：renderLine 的 run merging 逻辑缺陷

当"•" (width=2) 后面跟着空格时，场景如下：

**场景：清除 "• " (bullet + space)**

```
Front buffer (渲染前):
  cell[0]: "•" (Width=2, IsContinuation=false)
  cell[1]: ""  (IsContinuation=true)
  cell[2]: " " (Width=1)

Back buffer (渲染后 - 清除"•"后):
  cell[0]: " " (Width=1)
  cell[1]: " " (Width=1)
  cell[2]: " " (Width=1)  ← 未变化
```

**修复前的 run merging 逻辑**：

```go
// 收集当前单元格以及后续相邻的、同样式的单元格
for ; x < endX; {
    nextCell := r.back.Cells[y][x]
    nextPrevCell := r.front.Cells[y][x]

    // ❌ 问题：检查 IsCellChanged
    if !IsCellChanged(nextCell, nextPrevCell) ||
        nextCell.IsContinuation ||
        nextCell.Style != runStyle ||
        nextCell.Cluster == "" {
        break
    }

    runText.WriteString(nextCell.Cluster)
    totalWidth += nextCell.Width
    x += nextCell.Width
}
```

**问题过程**：
1. x=0: `IsCellChanged(" ", "•") = true` → 合并
2. x=1: `IsCellChanged(" ", "") = true` → 合并
3. x=2: `IsCellChanged(" ", " ") = false` → **停止合并** ❌

**结果**：
- 输出 `runText = "  "` (两个空格)
- `totalWidth = 2`
- x=2 的空格没有被输出（因为被认为是未变化）
- 终端显示：`•  ` (bullet + 两个空格)

## 修复方案

### 修复1：emitRunWithWidth 的 cursorX 同步

**修改位置**：`runtime/paint/renderer.go`

**修复前**：
```go
func (r *Renderer) emitRunWithWidth(x, y int, runStyle style.Style, text string, textWidth int) {
    if text == "" {
        return
    }

    // ... 输出逻辑 ...

    r.cursorX = x + textWidth
}
```

**修复后**：
```go
func (r *Renderer) emitRunWithWidth(x, y int, runStyle style.Style, text string, textWidth int) {
    // ✓ 即使 text 为空，也要更新 cursorX 以保持光标同步
    r.cursorX = x + textWidth
    r.cursorY = y

    if text == "" {
        return
    }

    // ... 输出逻辑 ...
}
```

**效果**：确保 cursorX 始终与 buffer 位置同步，即使没有实际输出。

### 修复2：run merging 移除 IsCellChanged 检查

**修改位置**：`runtime/paint/renderer.go`，renderLine 函数

**修复前**：
```go
for ; x < endX; {
    nextCell := r.back.Cells[y][x]
    nextPrevCell := r.front.Cells[y][x]

    // ❌ 检查 IsCellChanged
    if !IsCellChanged(nextCell, nextPrevCell) ||
        nextCell.IsContinuation ||
        nextCell.Style != runStyle ||
        nextCell.Cluster == "" {
        break
    }

    runText.WriteString(nextCell.Cluster)
    totalWidth += nextCell.Width
    x += nextCell.Width
}
```

**修复后**：
```go
for ; x < endX; {
    nextCell := r.back.Cells[y][x]

    // ✓ 只检查样式、continuation 和空字符串
    // 移除 IsCellChanged 检查，因为一旦决定输出，就应该连续输出整个区域
    if nextCell.IsContinuation ||
        nextCell.Style != runStyle ||
        nextCell.Cluster == "" {
        break
    }

    runText.WriteString(nextCell.Cluster)
    totalWidth += nextCell.Width
    x += nextCell.Width
}
```

**原理**：
- 外层循环（`for x < endX`）负责找出第一个变化的单元格
- 内层循环负责从这个位置开始，连续收集所有相同样式的单元格
- 这样确保输出是连续的，光标移动和文本输出完全匹配

**修复效果**：
```
场景：清除 "• " 后：

Front buffer: "• "  (bullet + space)
Back buffer:  "   " (三个空格)

修复后的行为：
1. x=0 发现变化（"•" → " "）
2. 开始合并，收集所有相同样式的空格
3. 输出 runText = "   " (三个空格), totalWidth = 3
4. 光标移动 3 个位置，输出 3 个空格
结果 ✓：连续输出，无额外空格
```

## 验证测试

创建了以下测试验证修复：

1. **test_wide_space_analysis.go**：分析宽字符+空格场景的开头
2. **test_run_merging_debug.go**：模拟 run merging 的完整流程
3. **test_fix_validation.go**：验证修复后的逻辑

测试结果：
```text
修复前：
  - 检查 IsCellChanged 时遇到未变化单元格就停止合并
  - 导致部分空格被合并，部分被跳过
  - 结果：输出的文本宽度与光标移动不匹配

修复后：
  - 一旦开始合并，就连续收集所有相同样式的字符
  - 移除 IsCellChanged 检查，避免部分合并
  - 结果：输出的文本宽度与光标移动完全匹配
  ✓ totalWidth(4) == runText长度(4)
```

## 修复文件

- `runtime/paint/renderer.go`
  - emitRunWithWidth 函数
  - renderLine 函数（run merging 逻辑）

## 相关技术点

### TUI 渲染核心概念

1. **Front Buffer vs Back Buffer**
   - Front Buffer: 上次渲染的内容
   - Back Buffer: 本次将要渲染的内容
   - 通过 diff 找出需要更新的区域

2. **Dirty Region (脏区域)**
   - 只渲染变化的区域，优化性能
   - 通过 `extractDirtyRegions` 识别

3. **Run Merging (运行合并)**
   - 合并相邻的、同样式的文本输出
   - 减少 ANSI 转义序列数量
   - 提高渲染效率

4. **Cursor Tracking (光标跟踪)**
   - 精确跟踪光标位置
   - 使用 `moveCursorOptimized` 最小化光标移动

### 宽字符处理

1. **Cell 结构**
   ```go
   type Cell struct {
       Cluster        string  // 字符或字符串
       Style          Style   样式
       Width          int     显示宽度（1 或 2）
       IsContinuation bool    // 是否是宽字符的延续单元格
       ZIndex         int
       Selected       bool
   }
   ```

2. **宽字符布局**
   - 宽字符头：`Cluster = "•"`, `Width = 2`, `IsContinuation = false`
   - 延续单元格：`Cluster = ""`, `IsContinuation = true`

3. **清除宽字符**
   ```go
   func (b *Buffer) clearCellAt(x, y int) {
       // 清除主单元格和延续单元格
       if cell.IsContinuation && x > 0 {
           head := b.Cells[y][x-1]
           if head.Width == 2 {
               b.Cells[y][x-1] = Cell{}
           }
       }
       if cell.Width == 2 {
           b.Cells[y][x+1] = Cell{}
       }
       b.Cells[y][x] = Cell{}
   }
   ```

## 经验总结

1. **Cursor 同步优先**：即使没有实际输出，也要保持 cursorX 更新，避免位置漂移
2. **Run Merging 的完整性**：一旦开始输出一段区域，要确保完整输出，不能部分合并
3. **宽度计算要准确**：使用 cell.Width 而非 runewidth.StringWidth，避免边框字符等问题
4. **测试驱动修复**：通过单元测试和模拟测试快速定位问题根源

## 后续优化建议

1. **性能优化**：考虑更激进的 run merging 策略（如合并不同样式的空格）
2. **日志优化**：在生产环境中禁用 RenderLogger 的 Debug 输出
3. **测试覆盖**：增加更多宽字符组合的回归测试
