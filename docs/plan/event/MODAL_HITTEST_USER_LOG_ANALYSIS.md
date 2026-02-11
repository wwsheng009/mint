# Modal HitTest 用户日志分析

## 用户提供的日志数据

```
[MOUSE] Raw position: (65, 21) | Action: 2
[MOUSE] HitTest: Found 'button' at Bounds=(63,21,15x1) Local=(2,0)
```

```
[LAYER] [centerModal] modal=(0,0) size=40x11 container=156x44 offset=(58,16)
[LAYER] [GetMergedHitMap] Modal entry: ID=button, Bounds=(63,21,15x1)
[LAYER] [GetMergedHitMap] Modal entry: ID=button, Bounds=(81,21,11x1)
```

## 问题分析

### 配置 vs 实际

| 项目 | 配置值 | 实际值 |
|------|--------|--------|
| Buffer 大小 | 80x24 | 156x44 |
| Container | 80x24 | 156x44 |
| Modal 原始位置 | (0,0) | (0,0) |
| Modal 大小 | 40x11 | 40x11 |
| Modal 居中偏移 | (20,6) | (58,16) |
| Modal 最终位置 | (20,6) | (58,16) |
| 按钮 Y 坐标 | Y=11 | Y=21 |

### 根本原因

**Resize 事件导致 Container 大小变化**:

1. 用户配置了 `ui.WithWidth(80), ui.WithHeight(24)`
2. 但实际终端窗口是 156x44
3. 系统检测到终端大小后，通过 Resize 事件更新了 container
4. Modal 基于 156x44 重新计算居中位置

### HitTest 验证

```
鼠标点击: (65, 21)
按钮 Bounds: (63, 21, 15x1)
命中结果: ✅ 成功 (Local=(2,0))
```

**HitTest 工作正常**！点击位置 (65, 21) 在按钮范围 (63-77, 21) 内。

## 用户感知的问题

用户可能觉得"按钮显示在 Y=40，但点击 Y=20 能命中"是因为：

1. **视觉参考**: 用户可能以 80x24 的预期来观察
2. **实际渲染**: Modal 在 156x44 的窗口中居中于 Y=16
3. **坐标对应**:
   - 在 80x24 中，按钮应该在 Y=11
   - 在 156x44 中，按钮实际在 Y=21
   - 差异: 21 - 11 = 10 行

## 解决方案

### 选项 1: 固定 Buffer 大小（推荐用于测试）

如果希望 Modal 始终在 80x24 的区域内居中，需要确保 buffer 大小不受终端窗口影响：

```go
// 在 framework/app.go 中，Resize 事件不应该改变 buffer 大小
// 而是应该让 buffer 居中显示在实际终端中
```

### 选项 2: 接受终端大小

当前实现是正确的 - Modal 会根据实际终端大小自动居中。

### 选项 3: 添加调试信息

在 Modal 上显示实际的 buffer 大小，帮助用户理解：
```
Buffer: 156x44 (Terminal: 156x44)
Modal: centered at (58, 16)
Buttons at Y=21 (0-based)
```

## 结论

**HitTest 系统完全正确**。问题在于用户期望的 80x24 与实际终端 156x44 不匹配。

如果要验证这个，可以：
1. 将终端窗口调整为 80x24
2. 或者在代码中禁用自动 resize
