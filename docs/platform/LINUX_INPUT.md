# Linux 输入处理

## 概述

Mint 在 Linux 平台使用 [tcell](https://github.com/gdamore/tcell) 库处理终端输入。

## tcell 实现文件

- `runtime/platform/input_linux_tcell.go` - Linux 平台的 tcell 输入实现

## 关键实现细节

### 1. Ctrl+字母组合键处理

tcell 对于 Ctrl+字母组合返回的是**特殊的 Key 常量**（如 `tcell.KeyCtrlX`），而不是 `tcell.KeyRune` 类型。

```go
// tcell 定义的特殊 Ctrl 键
const (
    KeyCtrlA Key = iota + 65
    KeyCtrlB
    // ...
    KeyCtrlX
    KeyCtrlY
    KeyCtrlZ
)
```

**问题：** 如果代码只检查 `ev.Key() == tcell.KeyRune`，会完全忽略 Ctrl+字母组合。

**解决方案：** 显式处理所有 Ctrl+字母组合：

```go
switch ev.Key() {
case tcell.KeyCtrlA:
    input.Key = 'a'
    input.Modifiers |= ModCtrl
case tcell.KeyCtrlX:
    input.Key = 'x'
    input.Modifiers |= ModCtrl
// ... 其他 Ctrl 字母
}
```

### 2. KeyModifier 位掩码

KeyModifier 使用位掩码格式：

```go
const (
    ModShift      KeyModifier = 1 << iota // 1 (0001)
    ModCtrl                               // 2 (0010)
    ModAlt                                // 4 (0100)
    ModMeta                         // 8 (1000)
)
```

**重要：** 这个顺序必须与 tcell 的 ModMask 定义一致：
- tcell.ModShift = 1
- tcell.ModCtrl = 2
- tcell.ModAlt = 4
- tcell.ModMeta = 8

### 3. 修饰键检测

使用位掩码与运算检测修饰键：

```go
if ev.Modifiers()&tcell.ModCtrl != 0 {
    // Ctrl 被按下
}
```

### 4. 鼠标事件处理

tcell 鼠标事件使用 `tcell.EventMouse`：

```go
func (r *tcellInputReader) parseMouseEvent(ev *tcell.EventMouse, now time.Time) RawInput {
    // 获取位置
    x, y := ev.Position()
    input.MouseX = int(x)
    input.MouseY = int(y)

    // 处理滚轮
    if button == tcell.WheelUp {
        input.MouseAction = MouseWheelUp
        return input
    }

    // 处理按钮
    if button&tcell.ButtonPrimary != 0 {
        input.MouseButton = MouseLeft
        input.MouseAction = MousePress
    }
    // ...
}
```

## 常见问题

### Ctrl+X 无法识别

**原因：** 代码只检查 `tcell.KeyRune`，没有处理 `tcell.KeyCtrlX` 等特殊键。

**解决：** 在 parseKeyEvent 中添加对所有 Ctrl+字母的显式处理。

### Shift+X 显示错误

**原因：** KeyModifier 常量顺序错误（ModAlt 和 ModCtrl 位置颠倒）。

**解决：** 确保常量顺序为 Shift=1, Ctrl=2, Alt=4, Meta=8。

## 调试

启用调试输出来查看 tcell 返回的原始值：

```bash
TUI_DEBUG_INPUT=true go run ./examples/ui_demos/demo2_runtime_internals/inspector_overlay/
```

## 参考

- [tcell GitHub](https://github.com/gdamore/tcell)
- [tcell Godoc](https://pkg.go.dev/github.com/gdamore/tcell/v2)
