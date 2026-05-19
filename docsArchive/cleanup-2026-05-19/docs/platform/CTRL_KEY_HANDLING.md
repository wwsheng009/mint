# Ctrl 键处理详解

## 问题描述

Ctrl+字母组合键（如 Ctrl+X、Ctrl+V）在 tcell 平台上需要特殊处理。

## tcell 的行为

tcell 对于 Ctrl+字母组合**不会**返回 `KeyRune` 类型，而是返回特殊的 Key 常量：

| 按键 | tcell 返回 |
|------|-----------|
| Ctrl+A | `tcell.KeyCtrlA` |
| Ctrl+B | `tcell.KeyCtrlB` |
| ... | ... |
| Ctrl+X | `tcell.KeyCtrlX` |
| Ctrl+Z | `tcell.KeyCtrlZ` |

## 错误的实现（不可行）

```go
// ❌ 这样写无法捕获 Ctrl+X
if ev.Key() == tcell.KeyRune {
    ch := ev.Rune()
    if ev.Modifiers()&tcell.ModCtrl != 0 {
        // 永远不会执行！Ctrl+X 不是 KeyRune 类型
    }
    input.Key = ch
}
```

## 正确的实现

```go
// ✅ 正确：显式处理每个 Ctrl+字母组合
switch ev.Key() {
case tcell.KeyCtrlA:
    input.Key = 'a'
    input.Modifiers |= ModCtrl
case tcell.KeyCtrlB:
    input.Key = 'b'
    input.Modifiers |= ModCtrl
// ...
case tcell.KeyCtrlX:
    input.Key = 'x'
    input.Modifiers |= ModCtrl
case tcell.KeyCtrlZ:
    input.Key = 'z'
    input.Modifiers |= ModCtrl
}

// ✅ 同时保留普通字符处理
if ev.Key() == tcell.KeyRune {
    input.Key = ev.Rune()
}
```

## 数据流

```
tcell.EventKey
    ↓
parseKeyEvent()
    ↓
RawInput{Key: 'x', Modifiers: ModCtrl}
    ↓
Pump.convertToKeyMsg()
    ↓
KeyMsg{Rune: 'x', Mod: Modifiers{Ctrl: true}}
    ↓
MsgToEvent()
    ↓
KeyEvent{Key: Key{Rune: 'x'}, Modifiers: ModCtrl}
    ↓
Inspector 显示: "Ctrl+x"
```

## 相关文件

- `runtime/platform/input_linux_tcell.go` - Linux 实现
- `runtime/platform/input_darwin_tcell.go` - Darwin 实现
- `runtime/platform/input.go` - RawInput 定义
- `runtime/msg/key_msg.go` - KeyMsg 定义
- `framework/event/msg_adapter.go` - Msg 到 Event 转换
- `framework/event/pump.go` - RawInput 到 Msg 转换
