# Demo 3: Styling System (TUI CSS)

## 概述

样式系统 = **CSS Box Model + 终端颜色/属性系统 + 继承规则**

## 特性

| 特性    | 说明          |
| ----- | ----------- |
| Box Model | Padding, Margin, Border |
| Color | Foreground, Background   |
| Attributes | Bold, Italic, Underline |
| Inheritance | Fg/Bold 继承，Bg 不继承   |
| Theme | 全局主题系统         |

## 运行

```bash
cd examples/ui_demos/demo3_styling
go run main.go
```

## 标签页

### Colors (颜色)

展示所有可用颜色：
- 基础色：Red, Green, Blue, Yellow, Cyan, Magenta, White, Black, Gray
- 亮色：Bright Red, Bright Green, Bright Blue, Bright Yellow, Bright Cyan, Bright Magenta, Bright White

### Attributes (属性)

展示文本属性：
- Normal (普通)
- Bold (粗体)
- Italic (斜体)
- Underline (下划线)
- 组合：Bold + Italic + Underline

### Borders (边框)

展示边框样式：
- Single (单线): `┌───┐`
- Double (双线): `╔═══╗`
- Rounded (圆角): `╭───╮`
- Dashed (虚线)
- Thick (粗线)
- Dotted (点线)

### Inheritance (继承)

样式继承规则：
| 属性    | 是否继承 |
| ----- | ---- |
| Fg    | ✔    |
| Bg    | ✖    |
| Bold  | ✔    |
| Italic | ✔    |
| Underline | ✔    |
| Border | ✖    |
| Padding | ✖    |

### Themes (主题)

主题系统演示：
- Default (默认)
- Dark (暗色)
- Ocean (海洋)
- Forest (森林)
- Sunset (日落)
- Violet (紫罗兰)

## 验证目标

有了样式系统，引擎从：
> "功能型 TUI"

进化成：
> **"可设计的 UI 框架"**

## 基于文档

`../../../docsArchive/cleanup-2026-05-19/_framework_docs/ui/demo/demo3_with_style.md`
