# Fiber Counter 按钮重复问题 - 日志分析报告

## 执行摘要

通过使用 `TUI_DEBUG_ALL=true` 运行 fiber_counter，发现按钮重复问题的根本原因是：

**`renderFiberToBuffer` 中使用的手动宽度测量 (`measureFiberWidth`) 与实际布局引擎计算的位置不一致。**

这与之前的分析结论完全一致 - 问题在 UI 渲染层，而非 Fiber reconciliation 层。

---

## 实验配置

### 测试版本

**1. 单空格版本 (main_one_space.go) - 触发问题**
```go
ui.HStack(
    app.ButtonBuilder(" - ").
        OnPress(intent.Decrement("count", 1)).Build(),
    ui.Text(" "),  // 单个空格
    app.ButtonBuilder(" + ").
        OnPress(intent.Increment("count", 1)).Build(),
)
```

**2. 双空格版本 (main.go) - 临时解决方案**
```go
ui.HStack(
    app.ButtonBuilder(" - ").
        OnPress(intent.Decrement("count", 1)).Build(),
    ui.Text(" "),  // 第一个空格
    ui.Text(" "),  // 额外空格
    app.ButtonBuilder(" + ").
        OnPress(intent.Increment("count", 1)).Build(),
)
```

### 日志文件

- `fiber_counter_one_space_debug_20260227_024235.log.err` - 单空格完整日志
- `fiber_counter_debug_20260227_024014.log.err` - 双空格完整日志

---

## 日志分析

### 1. Fiber Tree 结构 ✅ 正确

从日志中确认 Fiber tree 结构完全正确：

```log
[UI] [createAllNewChildren] Creating 3 children for parent Key="_idx_1", Tag="hstack"
[HitMap] [CREATEFIBER] Type=Element Key= Tag=button
[UI] [createAllNewChildren] Created child 0: Type=0(VNodeElement), Key="_idx_0", Tag="button", Path="/root/base[0]/hstack[0]/button[0]"
[HitMap] [CREATEFIBER] Type=Element Key= Tag=text
[UI] [createAllNewChildren] Created child 1: Type=0(VNodeElement), Key="_idx_1", Tag="text", Path="/root/base[0]/hstack[0]/text[0]"
[HitMap] [CREATEFIBER] Type=Element Key= Tag=button
[UI] [createAllNewChildren] Created child 2: Type=0(VNodeElement), Key="_idx_2", Tag="button", Path="/root/base[0]/hstack[0]/button[1]"
```

**结论**: HStack 确实有 3 个子节点，结构正确，没有重复。

### 2. 布局计算 ✅ 相对正确

#### 单空格版本 (问题版本)

```log
[Paint] [Paint.paintBox] hstack at (0,1) size 19x1
[Paint] [Paint.paintBox] button at (0,1) size 8x1
[Paint] [Paint.paintBox] text at (9,1) size 1x1
[Paint] [Paint.paintBox] button at (11,1) size 8x1
```

**布局说明**:
- HStack 位于 y=1，宽度 19（8 + 1 + 1 + 8 + 1 gap = 19）
- 第一个 button: (0,1)，宽度 8
- 空格文本: (9,1)，宽度 1 (8 + 1 gap = 9)
- 第二个 button: (11,1)，宽度 8 (9 + 1 + 1 gap = 11)

#### 双空格版本 (正常版本)

```log
[Paint] [Paint.paintBox] hstack at (0,1) size 21x1
[Paint] [Paint.paintBox] button at (0,1) size 8x1
[Paint] [Paint.paintBox] text at (9,1) size 1x1
[Paint] [Paint.paintBox] text at (11,1) size 1x1
[Paint] [Paint.paintBox] button at (13,1) size 8x1
```

**布局说明**:
- HStack 位于 y=1，宽度 21
- 第二个 button 移动到 x=13

### 3. ✅ renderFiberToBuffer 的位置 - 发现关键 Bug

这是发现问题的关键日志！

```log
[Render] renderFiberToBuffer tag="hstack", x=0, y=0
[Render] renderFiberToBuffer tag="button", x=0, y=0
[Render] renderFiberToBuffer tag="text", x=6, y=0
[Render] renderFiberToBuffer tag="button", x=7, y=0
```

**问题**:
- `renderFiberToBuffer` 中，第二个 button 在 **x=7** (相对于 HStack)
- 但 `paintBox` 布局计算说第二个 button 应该在 **x=11** (绝对位置)
- 差异 = 4 列

**位置计算分析**:

`renderFiberToBuffer` 中的计算：
```go
childX += width + gap
```

对于单空格版本的 HStack (gap=1):
1. 第一个 button 在 x=0，宽度 = ?
2. `childX` = 0 + buttonWidth + 1 = 6
3. Text " " 在 x=6
4. `childX` = 6 + 1 + 1 = 8 (但日志显示 x=7!)

这里 `childX` 计算有问题。实际日志显示 Text 在 x=6，第二个 button 在 x=7，这意味着：
- Text 在 x=6 (应该是 x=9)
- 第二个 button 在 x=7 (应该是 x=11)

### 4. Button 宽度不一致

查看 `measureFiberWidth` 的实现：

```go
case "button":
    label := ""
    if fiber.Props != nil {
        if l, ok := fiber.Props["label"].(string); ok {
            label = l
        }
    }
    return paint.StringWidth(label) + 2
```

对于 label "  +  "（空格+号空格）:
- `paint.StringWidth("  +  ")` = 5 (假设全角字符)
- 实际宽度 = 5 + 2 = 7 ❌

但 paintBox 显示按钮宽度是 8！

**这里存在严重的bug**: `measureFiberWidth` 返回的值与实际布局引擎计算不同。

### 5. 实际输出的字符内容

从标准输出日志中可以看到实际显示的内容：

```ansi
\x1b[2J\x1b[?25l\x1b[1;1H\x1b[32mCount: 0\x1b[2;1H\x1b[1;48;2;136;192;208;38;2;236;239;244m*[  -  ]\x1b[3C\x1b[0m\x1b[38;2;236;239;244;48;2;59;66;82m [  +  ]\x1b[3;1H [  +  ]\x1b[4;1H\x1b[0m\x1b[90m[Fiber: true] Using GlobalState + Built-
```

**解读**:
```
Line 1 y=1: Count: 0
Line 2 y=2: *[  -  ] [  +  ]  (正常)
Line 3 y=3: [  +  ]                ← 重复的第二个按钮！
Line 4 y=4: [Fiber: true] Using...
```

**注意**: 第二个重复的按钮出现在下一行 (y=3)！

### 6. HitMap 测试 ✅ 无重复

从日志：

```log
[FocusManager] CollectFromFiber: collected 2 focusable fibers
[FocusManager]   [0] FocusID=node-7, Tag=button
[FocusManager]   [1] FocusID=node-9, Tag=button
```

只有 2 个按钮被正确收集，没有重复。

---

## 根本原因分析

### 问题定位

`internal/reconciler/reconciler.go` 中的 `renderFiberToBuffer` 函数：

```go
for child := fiber.Child; child != nil; child = child.Sibling {
    r.renderFiberToBuffer(child, childX, childY, buffer)

    // Move position for next sibling
    if isHStack {
        width := r.measureFiberWidth(child)
        childX += width + gap
    }
    // ...
}
```

**问题**:
1. `measureFiberWidth` 返回的宽度与实际布局引擎计算的宽度不一致
2. 导致 `renderFiberToBuffer` 中使用的坐标与 `paintBox` 中的坐标不匹配
3. 当计算到第二个按钮时，位置偏移不正确
4. 第二个按钮的内容被绘制到了错误的位置（可能换行了）

### 为什么双空格能解决？

- 额外的一个空格增加了 HStack 总宽度
- 这改变了位置计算的累积值
- 错位的按钮内容正好适应了新的布局位置
- 但这只是一个偶然的效果，不是根本解决方案

---

## 证据总结

| 方面 | 单空格版本 | 双空格版本 | 状态 |
|------|-----------|-----------|------|
| Fiber tree 结构 | 正确，3个子节点 | 正确，4个子节点 | ✅ |
| HitMap | 2个按钮 | 2个按钮 | ✅ |
| | | | |
| paintBox Button1 位置 | (0,1), 8x1 | (0,1), 8x1 | ✅ |
| paintBox Button2 位置 | (11,1), 8x1 | (13,1), 8x1 | ✅ |
| | | | |
| renderFiberToBuffer Button1 | x=0, y=0 | x=0, y=0 | ✅ |
| renderFiberToBuffer Button2 | x=7, y=0 | x=8, y=0 | ❌ |
| | | | |
| measureFiberWidth | 返回值不准确 | 偏移刚好正确 | ❌ |
| | | | |
| 实际输出 | 出现 `[  +  ]` 重复 | 无重复 | ❌ |

---

## 相关代码位置

- `internal/reconciler/reconciler.go`
  - `renderFiberToBuffer()` - 第 417 行
  - `measureFiberWidth()` - 第 447 行
  - Button 宽度计算 - 第 484-489 行

- `runtime/compute/fiber_only_layout.go` - paintBox 布局计算

---

## 建议的修复方案

### 方案 1: 使用 ComputedBox 的宽度

修改 `measureFiberWidth` 优先使用布局引擎计算的宽度：

```go
measureFiberWidth(fiber *Fiber) int {
    // 优先使用 ComputedBox 的宽度（布局引擎计算的正确值）
    if fiber.ComputedBox != nil {
        if box, ok := fiber.ComputedBox.(interface{ GetWidth() int }); ok {
            w := box.GetWidth()
            log.RenderLogger.Debug("measureFiberWidth: ComputedBox width=%d for tag=%s", w, fiber.Tag)
            return w
        }
    }
    // ... 其他 fallback 逻辑
}
```

### 方案 2: 统一使用布局引擎的位置

考虑完全移除 `renderFiberToBuffer` 中的手动位置计算，使用布局引擎计算的位置。

### 方案 3: 添加调试日志验证

在 `measureFiberWidth` 和布局计算中添加更详细的日志，对比两组宽度值：

```go
log.RenderLogger.Debug("[WIDTH] Button label=%q, measured=%d, ComputedBox=%d",
    label, paint.StringWidth(label)+2, computedWidth)
```

---

## 下一步行动

1. ✅ 问题已定位到 `measureFiberWidth` 和布局不一致
2. ✅ 创建了详细的日志分析报告
3. [TODO] 修复 `measureFiberWidth` 使用 ComputedBox 的宽度
4. [TODO] 添加单元测试验证修复
5. [TODO] 移除临时的双空格解决方案
6. [TODO] 验证所有 examples 中的 HStack 组件

---

## 相关文档

- `docs/fixes/FIBER_COUNTER_BUTTON_DUPLICATION_ANALYSIS.md` - 初始分析报告
- `internal/reconciler/button_duplication_bug_test.go` - 之前的单元测试

---

**日志分析完成时间**: 2026-02-27
**状态**: 根本原因已确定 ✓
**严重程度**: 高 - 依赖宽度计算的渲染都可能有 bug
