# Paint 渲染调试详解

本文档详细说明如何使用 `TUI_UI_DEBUG_PAINT` 调试组件的渲染过程。

## 📖 前置知识

**如果您是第一次使用 Paint Debug**，建议先阅读：
- [README.md#2-paint-debug](README.md#2-paint-debug---渲染过程调试) - Paint 模式概览
- [quick_start.md#2-渲染调试](quick_start.md#2-渲染调试-paint-debug) - 基本使用方法

**本文档适合**:
- ✅ 已经启用过 Paint Debug，理解基本输出
- ✅ 遇到具体的渲染问题，需要深入分析
- ✅ 想要了解宽度计算、文本对齐等细节

## 🎨 Paint 调试概述

Paint 阶段是两阶段渲染的第二阶段，负责将布局计算的结果实际绘制到终端缓冲区。

### 启用 Paint Debug

```bash
TUI_UI_DEBUG_PAINT=true ./demo2.exe
```

## 📊 Debug 输出内容

### 1. Paint 方法调用

```
[Paint.paintNode] Element at (0,0) size 80x19, vnode_type=*ui.LayoutNode
[Paint.paintNode]   ❌ Paintable: NO (type assertion failed)
[Paint.paintNode] Element at (0,0) size 80x3, vnode_type=*ui.BorderedNode
[Paint.paintNode]   ❌ Paintable: NO (type assertion failed)
[Paint.paintNode] Text at (19,1) size 41x1, vnode_type=*basic.TextVNode
[Paint.paintNode]   ✅ Paintable: YES, calling Paint(19, 1)
```

**字段说明**:
- `Element/Text at (x,y) size wxh` - 节点位置和尺寸
- `vnode_type` - VNode 类型
- `✅/❌ Paintable` - 是否实现了 Paintable 接口

### 2. Button 渲染详情

```
[DEBUG-PAINT] label="[1] Event", bounds=[3 12 19 1], x=3, y=12
[DEBUG-PAINT]   contentWidth=14, naturalWidth=14, layoutWidth=19
[DEBUG-PAINT]   paddingLeft=0, paddingRight=0, willStretch=true
[DEBUG-PAINT]   final buttonText length=19, text=">[ [1] Event ]     "
```

**字段说明**:
- `bounds` - [x, y, width, height] 从布局阶段得到的边界
- `contentWidth` - 文本内容实际宽度（不含符号）
- `naturalWidth` - 不受约束时的自然宽度
- `layoutWidth` - 布局分配的宽度（flex拉伸后）
- `willStretch` - 是否需要拉伸填充
- `buttonText` - 最终渲染的文本

### 3. 文本对齐计算

```
[DEBUG-ALIGN] label="Left", textAlign=Start, availableSpace=14, buttonText len=25
[DEBUG-ALIGN] label="Spacious", textAlign=Center, availableSpace=65, buttonText len=79
```

**字段说明**:
- `textAlign` - 对齐方式 (Start/Center/End)
- `availableSpace` - 可用于填充的空格数
- `buttonText len` - 最终文本长度

### 4. DrawCmd 返回值

```
[DEBUG-RETURN] label="Left", cmd.X=0, cmd.Y=3, cmd.Text="*[ Left ]                " (len=25)
[DEBUG-RETURN] label="Center", cmd.X=27, cmd.Y=3, cmd.Text="        [ Center ]        " (len=26)
[DEBUG-RETURN] label="Right", cmd.X=54, cmd.Y=3, cmd.Text="                [ Right ]" (len=25)
```

**字段说明**:
- `cmd.X, cmd.Y` - 绘制位置
- `cmd.Text` - 实际绘制的文本
- `len` - 文本长度

## 🔍 宽度计算详解

### Button.Paint() 宽度计算流程

```go
// 1. 计算文本内容宽度
contentWidth := utf8.RuneCountInString(label) + 2  // 文本 + "[]"

// 2. 计算 naturalWidth（不受约束）
naturalWidth := contentWidth  // 不包括 padding！

// 3. 获取 layoutWidth（从 bounds）
layoutWidth := naturalWidth
if b.bounds[2] > 0 {  // bounds[2] = width
    layoutWidth = b.bounds[2]  // 使用布局分配的宽度
}

// 4. 判断是否需要拉伸
willStretch := layoutWidth > naturalWidth
availableSpace := layoutWidth - naturalWidth

// 5. 根据 textAlign 填充空格
switch textAlign {
case AlignCenter:
    leftSpace := availableSpace / 2
    rightSpace := availableSpace - leftSpace
    buttonText = strings.Repeat(" ", leftSpace) + buttonText +
                 strings.Repeat(" ", rightSpace)
case AlignStart:
    buttonText = buttonText + strings.Repeat(" ", availableSpace)
case AlignEnd:
    buttonText = strings.Repeat(" ", availableSpace) + buttonText
}
```

### 宽度验证清单

✅ **正确的宽度关系**:
```
contentWidth ≤ naturalWidth ≤ layoutWidth
```

✅ **拉伸计算正确**:
```
availableSpace = layoutWidth - naturalWidth
buttonText长度 ≈ layoutWidth（差异 ≤ 1）
```

⚠️ **常见问题**:

| 问题 | 症状 | 原因 | 解决方案 |
|------|------|------|---------|
| 文本被截断 | `len < layoutWidth` | paddingRight 少计算 | 检查空格填充逻辑 |
| 文本过长 | `len > layoutWidth` | 没有正确截断 | 添加截断检查 |
| 没有拉伸 | `naturalWidth == layoutWidth` | Flex 没生效 | 检查 bounds 设置 |
| 不对称 | Center 对齐时左右不等 | availableSpace 为奇数 | 正常现象，左边少一个 |

## 🎯 实际调试案例

### 案例 1: 按钮没有填充到布局宽度

**症状**:
```
[DEBUG-PAINT] label="Button", naturalWidth=12, layoutWidth=26
[DEBUG-PAINT]   buttonText len=12, text="[ Button ]"
```

**问题**: `len=12` 应该是 26

**调试步骤**:
1. 检查 `willStretch` 是否为 `true`
2. 检查 `availableSpace` 是否计算正确
3. 检查 textAlign 是否正确应用
4. 检查 strings.Repeat 是否正确执行

**预期输出**:
```
[DEBUG-PAINT] label="Button", naturalWidth=12, layoutWidth=26
[DEBUG-PAINT]   willStretch=true, availableSpace=14
[DEBUG-PAINT]   buttonText len=26, text="[ Button ]              "
```

### 案例 2: Center 对齐不对称

**症状**:
```
[DEBUG-ALIGN] label="Test", textAlign=Center, availableSpace=13
[DEBUG-RETURN] cmd.Text="      [ Test ]     " (左边6空格，右边7空格)
```

**说明**: 这是**正常行为**！
- `availableSpace = 13`（奇数）
- `leftSpace = 13 / 2 = 6`
- `rightSpace = 13 - 6 = 7`
- 左边比右边少1个空格

**验证**:
```
总长度 = 9 (文本) + 6 + 7 = 22 ✅
或者接近 layoutWidth ✅
```

### 案例 3: Focus 指示符影响宽度

**症状**:
```
[DEBUG-RETURN] label="Button", cmd.Text="*[ Button ]                " (len=25)
预期 len=26（layoutWidth）
```

**分析**:
- Focus 指示符 `*` 会替换左括号 `[`
- 文本变成 `*[ Button ]` 而不是 `[ [ Button ] ]`
- 长度可能少1个字符

**这是正常的**，因为 `*` 和 `[` 都是1个字符，替换不影响总宽度。

## 📈 性能分析

### 使用 Paint Debug 分析性能

```bash
# 统计 Paint 调用次数
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "Paintable: YES" | wc -l

# 查找最慢的渲染（如果添加了 timing）
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "Paint time"

# 对比不同场景的渲染次数
# 场景1: 初始渲染
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | tee initial.log

# 场景2: 重渲染
# (触发 UI 更新后)
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | tee reflow.log

# 对比
diff initial.log reflow.log
```

## 🔧 代码中添加 Paint Debug

### 在 Button.Paint() 中添加调试

```go
func (b *Button) Paint(x, y int) []paint.DrawCmd {
    // ... 原有代码 ...

    // 添加 debug 输出
    if os.Getenv("TUI_UI_DEBUG_PAINT") == "true" {
        fmt.Fprintf(os.Stderr, "[DEBUG-PAINT] label=%q, bounds=[%d %d %d %d], x=%d, y=%d\n",
            b.label, b.bounds[0], b.bounds[1], b.bounds[2], b.bounds[3], x, y)
        fmt.Fprintf(os.Stderr, "[DEBUG-PAINT]   contentWidth=%d, naturalWidth=%d, layoutWidth=%d\n",
            contentWidth, naturalWidth, layoutWidth)
        fmt.Fprintf(os.Stderr, "[DEBUG-PAINT]   willStretch=%v\n", willStretch)
        fmt.Fprintf(os.Stderr, "[DEBUG-PAINT]   buttonText len=%d, text=%q\n",
            utf8.RuneCountInString(buttonText), buttonText)
    }

    return []paint.DrawCmd{...}
}
```

### 在 Text.Paint() 中添加调试

```go
func (t *Text) Paint(x, y int) []paint.DrawCmd {
    if os.Getenv("TUI_UI_DEBUG_PAINT") == "true" {
        fmt.Fprintf(os.Stderr, "[DEBUG-TEXT] text=%q, x=%d, y=%d, len=%d\n",
            t.text, x, y, utf8.RuneCountInString(t.text))
    }
    // ...
}
```

## 💡 最佳实践

### 1. 调试新组件

```bash
# 只看特定组件的 Paint
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "Button"

# 过滤特定标签
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "label=\"MyButton\""
```

### 2. 验证宽度计算

```bash
# 检查所有按钮的宽度
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "layoutWidth"

# 查找异常宽度（比如太短）
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | awk '/layoutWidth=/{if ($NF < 20) print}'
```

### 3. 对齐问题调试

```bash
# 查看 Center 对齐的计算
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "textAlign=Center"

# 验证最终文本长度
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "buttonText len="
```

### 4. DrawCmd 验证

```bash
# 检查返回的绘制命令
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "DEBUG-RETURN"

# 验证位置是否连续
TUI_UI_DEBUG_PAINT=true ./demo2.exe 2>&1 | grep "cmd.X=" | awk '{print $NF}' | sort -n
```

## 🔄 与 Layout Debug 配合使用

```bash
# 同时启用 Layout 和 Paint debug
TUI_UI_DEBUG_LAYOUT=true TUI_UI_DEBUG_PAINT=true ./demo2.exe

# 保存完整输出
TUI_UI_DEBUG_LAYOUT=true TUI_UI_DEBUG_PAINT=true ./demo2.exe > full_debug.txt 2>&1

# 对比 layout 和 paint 的宽度
# Layout 输出:
#   Size: 19x1
# Paint 输出:
#   layoutWidth=19
# 两者应该匹配！
```

## 📖 相关文档

- [环境变量参考](environment_variables.md) - 所有调试选项
- [布局调试 API](layout_api.md) - Layout phase 调试
- [快速入门](quick_start.md) - 快速开始指南

---

**版本**: v1.0
**最后更新**: 2025-02-08
**维护者**: Claude Sonnet 4.5
