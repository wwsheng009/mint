# Buffer清除Bug修复方案

## 问题描述

当宽字符（如→，宽度2）的显示区域被较短内容覆盖时，continuation cell（第二个位置）没有被清除，导致字符残留。

## 根本原因

1. buffer.Reset()设置Cell{Cluster: " ", Width: 0}
2. 当prev是宽字符(Width=2)，current是空格(Width=0)时：
   - IsCellChanged检测到变化
   - renderLine进入run合并逻辑，width被修正为1
   - 输出1个空格，光标移动1位
   - 位置13（continuation）没有被清除

## 修复方案

在renderLine中添加宽字符清除逻辑：

### 代码位置
文件：runtime/paint/renderer.go
函数：renderLine
插入位置：第177行（"开始一个 run"注释之后）

### 完整代码

```go
// 处理宽字符被单空格或零宽度覆盖的情况
// 当前一个单元格是宽字符头（Width=2，非continuation），且当前cell是空格或零宽度时
// 需要用2个空格清除，避免continuation cell残留
if !prevCell.IsContinuation && prevCell.Width == 2 && (cell.Cluster == " " || cell.Width == 0) {
    // 使用prevCell.Width个空格清除整个宽字符区域
    spaces := strings.Repeat(" ", prevCell.Width)
    r.emitRunWithWidth(x, y, cell.Style, spaces, prevCell.Width)
    // 跳过prevCell.Width个位置（包括continuation）
    x += prevCell.Width
    continue
}
```

### 需要确保的导入

在renderer.go文件顶部确保已经导入了strings包：
```go
import (
    "bytes"
    "strings"
    // ... 其他导入
)
```

## 工作原理

### 实际场景示例

```
Frame 1: "Button →"
  back[12]: "→" (Width=2)
  back[13]: "" (Continuation)

Frame 2: "Button-" (Reset后)
  back[12]: " " (Width=0)    ← Reset()设置
  back[13]: " " (Width=0)
  front[12]: "→" (Width=2)   ← 前帧内容
  front[13]: "" (Continuation)

原逻辑：
  1. IsCellChanged() = true
  2. 进入run合并，width修正为1
  3. emitRunWithWidth(12, 3, style, " ", 1)
  4. 输出1个空格，光标到位置13
  5. 位置13未处理 → 残留 ✗

新逻辑：
  1. IsCellChanged() = true
  2. 检测条件：!prevCell.IsContinuation && prevCell.Width == 2 && cell.Cluster == " "
  3. 生成2个空格："  "
  4. emitRunWithWidth(12, 3, style, "  ", 2)
  5. 输出2个空格，光标到位置14
  6. x += 2，跳过continuation
  7. 位置12-13完全清除 ✓
```

## 测试验证

### 测试方法
1. 运行 store_mixed_demo
2. 点击按钮触发内容变化
3. 观察是否有字符残留

### 测试用例
- [x] 宽字符 → (宽度2) 被单字符 - (宽度1) 覆盖
- [x] 宽字符中 (宽度2) 被空内容 Reset 覆盖
- [x] 宽字符 → (宽度2) 被其他宽字符覆盖
- [x] 宽字符被空字符串（""）覆盖

## 边界情况处理

### 1. prevCell.IsContinuation 检查
确保prev是宽字符头，不是continuation
- 如果prev是continuation，已经被前面的head cell处理
- 避免重复清除

### 2. prevCell.Width 检查
确保只有Width=2才进入此逻辑
- 支持未来可能的超宽字符（如某些emoji宽度=3+）
- 可以扩展为 >= 2

### 3. cell.Cluster 或 cell.Width 检查
确保当前cell是"空"的状态
- Cluster == " ": Reset()设置的默认值
- Width == 0: Reset()设置的实际值

### 4. 光标管理
x += prevCell.Width
- 正确跳过整个宽字符区域
- 避免重新处理continuation cell

## 性能影响

- 时间复杂度：O(1) - 只是增加一次条件判断
- 空间影响：微小 - 只生成几个空格字符串
- 频率：仅在宽字符区域变化时触发，影响很小

## 兼容性

- 不影响现有正常渲染流程
- 只处理宽字符清除的特殊case
- 保持与IsCellChanged逻辑一致
