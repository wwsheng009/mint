# Header 背景断开的根本原因分析

## 问题现象

从测试输出的 ANSI 转义码可以看到：

```
[1;38;2;236;239;244;48;2;136;192;208mTUI Engine Demo[1C[0m[38;2;59;66;82;48;2;136;192;208m              [1C[1;93m>[ [Open Modal] ][1C[0m[38;2;59;66;82;48;2;136;192;208m  [1C[1;38;2;46;52;64mClicks: 0
↑ PRIMARY背景                             ↑ PRIMARY背景          ↑ 问题！                          ↑ PRIMARY背景
```

**关键发现**：
1. `TUI Engine Demo` → PRIMARY 背景 ✅
2. `              ` → PRIMARY 背景 ✅
3. `>[ [Open Modal] ]` → **只有前景色 `[1;93m`，没有背景！** ❌
4. `  ` → PRIMARY 背景 ✅
5. `Clicks: 0` → PRIMARY 背景（从前面延续）✅

## 根本原因

### Button 焦点指示符的实现问题

**文件**: `components/button/button.go`

**代码位置**: Line 610-619

```go
// Add focus indicator: * before focused button (only for Reverse style)
var focusIndicator string
if b.hasFocus && !b.disabled && b.focusStyle == FocusStyleReverse {
    focusIndicator = "*"
} else if b.hasFocus && !b.disabled && b.focusStyle == FocusStyleUnderline {
    focusIndicator = ">" // Visible indicator for underline style
} else if b.hasFocus && !b.disabled && b.focusStyle == FocusStyleBracket {
    focusIndicator = ">" // ❌ 问题：只是字符串，没有独立样式
} else {
    focusIndicator = " "
}

// Line 623: 组合文本
buttonText := focusIndicator + labelText  // ">" + "[Open Modal]"

// Line 632-633: 作为一个整体渲染
return []paint.DrawCmd{
    paint.NewTextCmd(x, y, buttonText, buttonStyle),
}
```

**问题分析**：

1. `focusIndicator = ">"` 是纯字符串
2. 与 `labelText = "[Open Modal]"` 组合成 `buttonText`
3. **整个** `buttonText` 使用 `buttonStyle` 渲染
4. 但在相邻的间隔字符之间，可能会有样式重置序列 `[0m`

### 为什么会产生 gap？

从 ANSI 转义码看：

```
[38;2;59;66;82;48;2;136;192;208m              ← 间隔字符结束
[1C                                       ← 光标移动
[0m                                      ← ❌ 样式重置！
[1;93m                                   ← 焦点指示符的前景（黄色）
>[ [Open Modal] ]                        ← 按钮文本
[1C                                       ← 光标移动
[0m                                      ← ❌ 样式重置！
[38;2;59;66;82;48;2;136;192;208m               ← 下一个间隔字符开始
```

**样式重置** `[0m` 导致背景色被清除，在间隔和按钮之间产生了一个透明的 gap！

### 为什么按钮本身没有设置背景？

查看 `button.go:554-556`:

```go
case ButtonVariantSecondary:
    // Secondary: BG=SURFACE, FG=TEXT
    buttonStyle = buttonStyle.Foreground(theme.Text()).Background(theme.Surface())
```

Secondary variant 默认使用 `SURFACE` 背景（灰色），不是 `PRIMARY`（蓝色）。

虽然在 `main.go:92` 中我们用 `.Style(style.Style{}.Background(theme.Primary()))` 覆盖了背景，但**覆盖可能不完整**。

---

## 解决方案

### 方案 1: 确保 focusIndicator 继承背景色（推荐）

修改 `button.go` 的 Paint 方法，确保焦点指示符使用相同的背景色：

```go
// Line 632-633 修改为：
return []paint.DrawCmd{
    // 如果有焦点指示符，单独渲染以确保背景连续
    paint.NewTextCmd(x, y, buttonText, buttonStyle),
}
```

但这样还是不够，因为问题在于**相邻元素之间的样式重置**。

### 方案 2: 移除样式重置（修改渲染引擎）

检查 `internal/render` 或 `runtime/paint` 中是否有不必要的 `[0m` 重置。

### 方案 3: 让按钮的背景色更明确

在 demo1 中，不仅覆盖按钮的背景，还要确保样式完全应用：

```go
// main.go:90-97
app.ButtonBuilder("[Open Modal]").
    Variant(app.ButtonVariantSecondary).
    Style(style.Style{}.
        Foreground(theme.Text()).      // 明确设置前景
        Background(theme.Primary()).   // 明确设置背景
        Bold(true)).                    // 明确设置加粗
    OnClick(func() {
        setShowModal(true)
    }).
    FocusStyle(app.FocusStyleBracket).
    Build(),
```

### 方案 4: 检查样式合并逻辑

检查 `style.Style` 的合并是否正确处理背景色。

---

## 验证步骤

1. 检查当前 Button 的 `buttonStyle` 是否包含 PRIMARY 背景
2. 检查渲染过程中是否有样式重置
3. 确认间隔字符和按钮之间的过渡是否平滑

---

## 下一步

需要进一步调查：
1. 运行时检查 `buttonStyle.BG` 的实际值
2. 追踪 ANSI 转义码的生成过程
3. 查看是否有地方插入了 `[0m` 重置序列

---

## 临时解决方案

如果问题出在样式重置上，可以尝试在 demo1 中调整 Header 的实现：

```go
// 不使用单独的间隔字符，而是让 HStack 的 Gap 功能处理间距
headerContent := ui.HStackBuilder(
    app.NewTextBuilder("TUI Engine Demo").
        Style(style.Style{}.Foreground(theme.Text()).Background(theme.Primary()).Bold(true)).
        Build(),
    app.ButtonBuilder("[Open Modal]").
        Variant(app.ButtonVariantSecondary).
        Style(style.Style{}.Background(theme.Primary())).
        FocusStyle(app.FocusStyleBracket).
        Build(),
    app.NewTextBuilder(fmt.Sprintf("Clicks: %d", count)).
        Style(style.Style{}.Foreground(theme.BG()).Background(theme.Primary()).Bold(true)).
        Build(),
).Gap(1).Align(ui.AlignCenter).Build()
```

但这样还是解决不了 `focusIndicator` 的背景问题。

**最根本的解决方案**：修改 Button 组件，确保焦点指示符和按钮文本作为一个整体时，背景色是连续的，不会被样式重置打断。
