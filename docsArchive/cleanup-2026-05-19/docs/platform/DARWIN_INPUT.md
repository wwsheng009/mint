# Darwin (macOS) 输入处理

## 概述

Mint 在 Darwin/macOS 平台使用 [tcell](https://github.com/gdamore/tcell) 库处理终端输入，实现与 Linux 平台完全一致。

## tcell 实现文件

- `runtime/platform/input_darwin_tcell.go` - Darwin 平台的 tcell 输入实现

## 构建标签

```go
//go:build darwin
// +build darwin
```

## 与 Linux 的一致性

Darwin 和 Linux 使用完全相同的 tcell 实现逻辑：

1. **Ctrl+字母处理** - 相同的 KeyCtrlA-Z 映射
2. **KeyModifier 位掩码** - 相同的位掩码定义
3. **鼠标事件处理** - 相同的鼠标事件转换逻辑

## macOS 特殊注意事项

### Terminal.app

默认的 Terminal.app 对某些键的支持有限。建议使用：
- iTerm2
- Alacritty
- WezTerm

### 功能键

某些功能键可能需要配置：
- F1-F12 可能被系统占用
- Delete/Backspace 行为可能因终端而异

## 调试

```bash
TUI_DEBUG_INPUT=true go run ./examples/ui_demos/demo2_runtime_internals/inspector_overlay/
```

## 参考文档

详细的 tcell 输入处理说明请参阅 [LINUX_INPUT.md](./LINUX_INPUT.md)。
