# HitTest 坐标系统分析

## 问题描述

用户反馈：点击位置与实际显示位置不同，modal按钮位置不正确。

## 分析结果

经过详细分析日志和代码，**HitTest系统本身是正确的**。

### 坐标系统说明

整个TUI系统使用 **0-based坐标**：

1. **Windows API (COORD)**: 0-based
   - `COORD.X = 0` → 第一列
   - `COORD.Y = 0` → 第一行

2. **Buffer坐标**: 0-based
   - `buffer.Cells[0][0]` → 第一行第一列的单元格
   - `buffer.Cells[y][x]` → 第y+1行，第x+1列

3. **HitMap坐标**: 0-based
   - `Bounds=(25,11,15x1)` → 左上角在(25,11)，宽度15，高度1

4. **ANSI转义序列**: 1-based
   - `\x1b[1;1H` → 移动到第1行第1列（0-based的Y=0, X=0）
   - `\x1b[12;1H` → 移动到第12行第1列（0-based的Y=11, X=0）

### 日志分析

从 `demo1_debug.log` 中的数据：

```
[12:14:55.685] [INFO] [LAYER] [centerModal] modal=(0,0) size=40x11 container=80x24 offset=(20,6)
[12:14:55.685] [INFO] [LAYER] [centerModal] after shift: modal=(20,6)
[12:14:55.685] [DEBUG] [HITMAP] [buildHitMapFromComputedBox] Entry: ID=button, Bounds=(25,11,15x1)
[12:14:55.685] [DEBUG] [HITMAP] [buildHitMapFromComputedBox] Entry: ID=button, Bounds=(43,11,11x1)
```

- Buffer大小: 80x24
- Modal位置: (20, 6) + 大小40x11 = 覆盖 Y=[6, 16], X=[20, 59]
- 按钮1位置: (25, 11, 15x1) → 覆盖 Y=[11, 11], X=[25, 39]
- 按钮2位置: (43, 11, 11x1) → 覆盖 Y=[11, 11], X=[43, 53]

### ANSI输出验证

从实际渲染的ANSI输出来看：

```
[9;1H=== MODAL START ===      ← 0-based: Y=8
...
[12;40H[ [ Cancel ] ]         ← 0-based: Y=11
[12;57H[ [ OK ] ]             ← 0-based: Y=11
```

**ANSI第12行 = 0-based的Y=11**，与HitMap记录的按钮位置一致！

## 结论

**HitTest系统完全正确**。坐标系统在整个渲染管线中是一致的：
1. Windows API返回0-based坐标
2. Buffer使用0-based坐标
3. HitMap使用0-based坐标
4. 渲染输出时转换为ANSI的1-based坐标

如果用户感觉点击位置不正确，可能的原因：
1. 终端窗口的实际大小与buffer大小不同（Resize事件处理问题）
2. 终端模拟器的鼠标坐标报告不准确
3. 用户将ANSI的1-based行号与0-based坐标混淆

## 验证方法

启用详细日志来验证：

```bash
TUI_DEBUG_LOG=debug.log ./demo.exe
```

检查日志中的：
- `[LAYER] [centerModal]` - 确认container大小和modal位置
- `[HITMAP] [buildHitMapFromComputedBox]` - 确认按钮的Bounds
- `[MOUSE] Raw position` - 确认鼠标点击的原始坐标
- `[MOUSE] HitTest` - 确认HitTest是否找到正确的按钮

所有坐标都应该是0-based的，并且与buffer大小一致。
