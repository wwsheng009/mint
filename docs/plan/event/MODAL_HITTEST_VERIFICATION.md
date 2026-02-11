# Modal 按钮 HitTest 坐标验证指南

## 问题现象

- **按钮显示位置**：正确（在屏幕上的视觉位置对）
- **HitTest 位置**：错误（点击时需要移动鼠标到不同的位置才能命中）
- **其他按钮**：正常工作（只有 modal 按钮有问题）

## 验证步骤

### 1. 运行 demo 并查看 HitMap 数据

```bash
cd examples/ui_demos/demo1_full_featured
AUTO_OPEN_MODAL=true TUI_DEBUG_HITMAP=true go run main.go 2>&1 | grep -E "button.*Bounds|centerModal" | head -10
```

**预期输出**：
```
[centerModal] modal=(0,0) size=40x9 container=80x24 offset=(20,7)
[centerModal] after shift: modal=(20,7)
[buildHitMapFromComputedBox] Entry: ID=button, Bounds=(25,11,15x1)
[buildHitMapFromComputedBox] Entry: ID=button, Bounds=(43,11,11x1)
```

### 2. 查看实际渲染位置

从 ANSI 输出中找到按钮位置：
```
[12;25H[ [ Cancel ] ]
```

- `[12;25H` = Line 12 (1-based), Column 25
- 转换为 0-based：Y=11

### 3. 对比 HitMap 坐标

| 元素 | HitMap Y坐标 | ANSI Line (1-based) | 实际Y (0-based) | 状态 |
|------|--------------|---------------------|-----------------|------|
| Modal 边框 | Y=7 | Line 8 | Y=7 | ✅ |
| 标题 | Y=9 | Line 10 | Y=9 | ✅ |
| **按钮** | **Y=11** | **Line 12** | **Y=11** | ✅ |

**HitMap 记录正确！**

### 4. 测试鼠标 HitTest

运行 demo 并移动鼠标：

```bash
AUTO_OPEN_MODAL=true TUI_DEBUG_MOUSE_POSITION=true go run main.go
```

移动鼠标到 modal 按钮位置，查看 stderr 输出：

```
[MOUSE] Raw position: (25, 11) | Action: 0
[MOUSE] HitTest: Found 'button' at Bounds=(25,11,15x1)
```

**关键验证点**：
1. 视觉上按钮在哪一行？
2. 鼠标在哪一行时 HitTest 找到按钮？
3. 这两个值是否一致？

## 可能的问题

如果坐标不一致，可能的原因：

### 问题 A：HitMap 在 centering 之前构建

**现象**：HitMap 记录 Y=2（centering 之前），实际显示 Y=11（centering 之后）

**验证方法**：检查日志顺序
```
[Engine.Layout] Built HitMap with 10 entries
[centerModal] after shift: modal=(20,7)
[buildHitMapFromComputedBox] Entry: ID=button, Bounds=(25,11,15x1)
```

**正确顺序**：centering → rebuild HitMap → HitMap 有正确坐标

**错误顺序**：build HitMap → centering → HitMap 有旧坐标

### 问题 B：鼠标事件坐标系统不一致

**现象**：
- 点击视觉 Y=40 的按钮
- 鼠标事件传入 Y=20
- 但 HitTest 还是命中了

**验证方法**：在 `pump.go` 中添加日志：
```go
fmt.Fprintf(os.Stderr, "[MOUSE] Raw position: (%d, %d)\n", raw.MouseX, raw.MouseY)
fmt.Fprintf(os.Stderr, "[MOUSE] HitTest result: %s\n", entry.NodeID)
```

**检查点**：
- Windows Console：0-based 坐标
- Unix X10：1-based 坐标（需要 -1）
- Unix SGR：0-based 坐标

### 问题 C：Layer HitMap 合并错误

**现象**：Base layer 按钮 HitTest 正常，Modal layer 按钮 HitTest 错误

**验证方法**：检查 `GetMergedHitMap` 输出
```
[GetMergedHitMap] Layer 0: HitMap has 27 entries
[GetMergedHitMap] Layer 2: HitMap has 10 entries
[GetMergedHitMap] Merged HitMap: 37 entries from 2 layers
```

**检查点**：Modal 按钮 entries 是否在合并后的 HitMap 中？

## 当前状态

根据日志 `b4b718b.output`：

✅ **HitMap 构建时机正确**（centering 后 rebuild）
✅ **HitMap 坐标记录正确**（Y=11 匹配视觉位置）
✅ **ANSI 渲染位置正确**（Line 12 = Y=11）

**待验证**：
❓ 鼠标事件传入的 Y 坐标是多少？
❓ HitTest 时使用的 Y 坐标是多少？

## 下一步

请运行 demo 并告诉我：

1. **视觉上按钮在哪一行？**（从屏幕上数）
2. **鼠标在哪一行时按钮被点击？**（移动鼠标直到按钮响应）
3. **stderr 输出的鼠标 Y 坐标是多少？**

这样我就能计算出实际的坐标偏移量。
