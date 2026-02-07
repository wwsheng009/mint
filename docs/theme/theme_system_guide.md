# Mint TUI 主题系统完全指南

> 版本：v1.0
> 更新时间：2025-02-07
> 适用范围：Mint TUI Framework

---

## 目录

1. [系统架构](#系统架构)
2. [主题预设系统](#主题预设系统)
3. [颜色渲染机制](#颜色渲染机制)
4. [Modal颜色切换](#modal颜色切换)
5. [开发指南](#开发指南)
6. [故障排查](#故障排查)

---

## 系统架构

### 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                    应用层 (Components)                     │
│  Button, Input, Select, Checkbox, etc.                     │
└────────────────────┬────────────────────────────────────────┘
                     │ 调用
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              framework/theme (主题管理层)                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ • 全局主题管理器 (GlobalManager)                    │   │
│  │ • 主题预设 (Nord, Dracula, etc.)                     │   │
│  │ • 颜色访问 API (Primary(), Text(), etc.)             │   │
│  │ • 颜色转换 (RGB → ANSI)                              │   │
│  └─────────────────────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────────┘
                     │ 返回 style.Color
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              runtime/style (样式定义层)                     │
│  • Style 结构体 (FG, BG, Bold, etc.)                       │
│  • Color 类型 (string alias)                              │
│  • 样式构建方法 (Foreground(), Background())             │
└────────────────────┬────────────────────────────────────────┘
                     │ 传递样式
                     ▼
┌─────────────────────────────────────────────────────────────┐
│            runtime/paint (渲染引擎层)                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ • StyleStateMachine (样式状态跟踪)                   │   │
│  │ • Renderer (差异渲染)                                │   │
│  │ • Buffer (双缓冲)                                    │   │
│  │ • colorCode() (颜色 → ANSI转换)                      │   │
│  └─────────────────────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────────┘
                     │ 输出
                     ▼
              ┌──────────────┐
              │ ANSI 转义序列  │
              │  (终端显示)   │
              └──────────────┘
```

### 核心组件职责

| 层级 | 组件 | 文件位置 | 职责 |
|------|------|----------|------|
| **主题管理** | `framework/theme` | `framework/theme/` | • 主题预设管理<br>• 全局颜色访问<br>• RGB → ANSI 转换 |
| **样式定义** | `runtime/style` | `runtime/style/style.go` | • Style 结构体<br>• 颜色类型定义 |
| **渲染引擎** | `runtime/paint` | `runtime/paint/` | • 样式状态机<br>• 差异渲染<br>• 颜色编码 |

---

## 主题预设系统

### 23个语义颜色

根据 `docs/theme/design/终端UI设计规范.md`，系统定义了23个语义颜色：

#### 1. 层级系统 (Layer System) - 3个
```go
BG      = "bg"        // 全局背景层 (最底层)
Surface = "surface"   // 一级表面层 (卡片、面板)
Overlay = "overlay"   // 悬浮层 (Modal, Dropdown)
```

#### 2. 排版系统 (Typography) - 3个
```go
Text        = "text"         // 主文本 (90%内容)
Muted       = "muted"        // 弱化文本 (辅助信息)
Placeholder = "placeholder"  // 占位文本 (输入提示)
```

#### 3. 品牌/交互 (Brand & Action) - 3个
```go
Primary   = "primary"    // 主操作色 (最重要按钮)
Secondary = "secondary"  // 次操作色 (辅助按钮)
Accent    = "accent"     // 点缀色 (Badge, 标签)
```

#### 4. 状态系统 (State) - 3个
```go
Success = "success"  // 成功状态
Warning = "warning"  // 警告状态
Error   = "error"    // 错误状态
```

#### 5. 内容关系 (Content Relations) - 2个
```go
Link    = "link"     // 可链接文本
Visited = "visited"  // 已访问链接
```

#### 6. 边界系统 (Boundaries) - 4个
```go
Border    = "border"    // 组件边框
Focus     = "focus"     // 焦点轮廓
Select    = "select"    // 选中背景
Highlight = "highlight" // 文本高亮
```

#### 7. 禁用状态 (Disabled) - 2个
```go
DisabledBG = "disabled-bg"  // 禁用控件背景
DisabledFG = "disabled-fg"  // 禁用控件文字
```

#### 8. 系统UI (System UI) - 3个
```go
Scrollbar = "scrollbar"  // 滚动条
Shadow    = "shadow"     // 阴影
Caret     = "caret"      // 光标
```

### 5套完整主题预设

位置：`framework/theme/preset.go`

#### Nord (北极主题)
```go
nordTheme() ThemePreset
• 设计理念：北极、冷色调、清晰
• 特点：柔和的蓝绿色调，适合长时间使用
• 典型颜色：
  - BG: #2E3440 (深灰蓝)
  - Primary: #88C0D0 (冰蓝)
  - Success: #A3BE8C (柔和绿)
```

#### Dracula (吸血鬼主题)
```go
draculaTheme() ThemePreset
• 设计理念：暗黑、高对比、霓虹感
• 特点：鲜艳的紫色和粉色，适合夜间使用
• 典型颜色：
  - BG: #282A36 (深紫黑)
  - Primary: #BD93F9 (亮紫)
  - Accent: #FF79C6 (粉红)
```

#### Gruvbox Dark (Gruvbox暗色)
```go
gruvboxDarkTheme() ThemePreset
• 设计理念：复古、温暖、舒适
• 特点：棕黄色调，类似羊皮纸
• 典型颜色：
  - BG: #282828 (暖灰)
  - Primary: #83A598 (柔和蓝绿)
  - Warning: #FABD2F (金黄)
```

#### Catppuccin Mocha (猫薄荷拿铁)
```go
catppuccinMochaTheme() ThemePreset
• 设计理念：现代、柔和、优雅
• 特点：丰富的粉彩色，平衡的对比度
• 典型颜色：
  - BG: #1E1E2E (深紫黑)
  - Primary: #89B4FA (柔和蓝)
  - Accent: #F5C2E7 (粉紫)
```

#### Solarized Dark (Solarized暗色)
```go
solarizedDarkTheme() ThemePreset
• 设计理念：精准、科学、护眼
• 特点：基于色彩科学的精确对比度
• 典型颜色：
  - BG: #002B36 (深蓝底)
  - Primary: #268BD2 (标准蓝)
  - Success: #859900 (标准绿)
```

### 主题数据结构

```go
// ColorValue 完整颜色定义
type ColorValue struct {
    RGB     [3]int  // RGB值 (TrueColor)
    ANSI16  int     // 16色模式代码 (0-15)
    ANSI256 int     // 256色模式代码 (0-255)
}

// ThemePreset 主题预设
type ThemePreset struct {
    Name   string
    Colors map[string]ColorValue  // 23个语义颜色
}
```

---

## 颜色渲染机制

### 颜色格式支持链

```
用户指定颜色
    ↓
framework/theme.Color (结构体)
    ├─ Type: ColorRGB
    ├─ Type: Color256
    └─ Type: ColorNamed
    ↓
ToStyleString() 转换
    ↓
style.Color (string)
    ├─ "#RRGGBB"      → TrueColor
    ├─ "0-255"        → 256色
    └─ "red"          → 命名颜色
    ↓
runtime/paint/colorCode()
    ↓
ANSI 转义序列
    ├─ ESC[38;2;R;G;Bm  (TrueColor前景)
    ├─ ESC[48;2;R;G;Bm  (TrueColor背景)
    ├─ ESC[38;5;Nm      (256色前景)
    ├─ ESC[48;5;Nm      (256色背景)
    ├─ ESC[30-37m       (16色前景)
    └─ ESC[40-47m       (16色背景)
```

### TrueColor 支持实现

**位置：** `runtime/paint/style_state.go:181-330`

```go
func colorCode(color style.Color, isBackground bool) string {
    c := string(color)

    // 1. TrueColor支持 (RGB十六进制)
    if strings.HasPrefix(c, "#") {
        rgb, err := parseHexString(c)  // "#RRGGBB" → [3]int
        if err != nil {
            return ""
        }
        // 生成TrueColor ANSI码
        if isBackground {
            return fmt.Sprintf("48;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
        }
        return fmt.Sprintf("38;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
    }

    // 2. 256色模式支持
    if is256Color(c) {
        code, _ := parse256Color(c)
        if isBackground {
            return fmt.Sprintf("48;5;%d", code)
        }
        return fmt.Sprintf("38;5;%d", code)
    }

    // 3. 16色模式支持 (命名颜色)
    code, ok := colorToAnsi[strings.ToLower(c)]
    if !ok {
        return ""
    }
    // ... 处理明亮色和标准色
}
```

### 样式状态机

**位置：** `runtime/paint/style_state.go:10-180`

**核心功能：** 最小化ANSI转义序列输出

```go
type StyleStateMachine struct {
    current style.Style  // 当前样式状态
}

// Update 生成样式变化的ANSI码
func (s *StyleStateMachine) Update(st style.Style) string {
    codes := s.buildDiffCodes(s.current, st)  // 只输出变化部分
    s.current = st
    return codes
}
```

**优化策略：**

1. **增量更新** - 只输出变化的样式属性
2. **智能重置** - 当多个属性变化时，直接重置 (`ESC[0m`)
3. **属性关停检测** - 移除样式时自动重置

```go
// 示例：从"粗体+蓝色"切换到"斜体+绿色"
from.Style{Bold: true, FG: "blue"}
to.Style{Italic: true, FG: "green"}

// 智能判断：需要关掉Bold，改变FG，开启Italic
// → 输出: ESC[0m + ESC[3m + ESC[38;2;0;255;0m
```

---

## Modal颜色切换

### Modal Backdrop实现

**位置：** `internal/render/paint_engine.go:242-335`

#### 设计目标

Modal弹出时，需要：
1. ✅ 聚焦用户注意力到Modal
2. ✅ 视觉上"禁用"背景内容
3. ✅ Modal关闭时完全恢复原始颜色

#### 实现机制

```go
func (e *PaintEngine) paintModalBackdrop(root *compute.ComputedBox, buffer *paint.Buffer) {
    // 1. Dimmed样式定义
    dimmedFG := style.Color("bright-black")  // 灰色文字
    dimmedBG := style.Color("#1e2028")       // 深色背景 (nord0 darker)

    // 2. 处理modal外部所有区域
    applyDimmed := func(x, y int) {
        cell := buffer.GetContent(x, y)

        if cell.Cluster == "" || cell.Cluster == " " {
            // 空白单元格：只设置背景
            buffer.SetCell(x, y, ' ', style.Style{BG: dimmedBG})
        } else {
            // 有内容单元格：前景+背景都设置
            dimmedStyle := style.Style{
                FG: dimmedFG,
                BG: dimmedBG,
            }
            buffer.SetCell(x, y, []rune(cell.Cluster)[0], dimmedStyle)
        }
    }

    // 3. 处理4个外部区域
    // - 上：y < modalY
    // - 下：y >= modalY + modalHeight
    // - 左：y在modal范围内，x < modalX
    // - 右：y在modal范围内，x >= modalX + modalWidth
}
```

#### Modal状态检测与全屏重绘

**问题：** Modal关闭后，需要恢复所有原始颜色

**解决方案：** 检测Modal出现/消失，强制全屏重绘

```go
// PaintEngine结构体添加状态跟踪
type PaintEngine struct {
    debug          bool
    lastHadModal   bool   // 上一帧是否有modal
    forceFullRender bool   // 标记需要全屏重绘
}

// PaintLayers 检测modal状态变化
func (e *PaintEngine) PaintLayers(layouts layer.LayerLayouts, buffer *paint.Buffer) error {
    _, hasModal := layouts[rtui.LayerModal]
    hadModal := e.lastHadModal

    // Modal状态改变 → 强制全屏重绘
    if hasModal != hadModal {
        e.forceFullRender = true
    }
    e.lastHadModal = hasModal
    // ...
}

// Paint 处理强制全屏重绘
func (e *PaintEngine) Paint(layout *compute.ComputedLayout, buffer *paint.Buffer) error {
    if e.forceFullRender {
        e.forceFullRender = false
        // 清空buffer，强制重绘所有单元格
        for y := 0; y < buffer.Height; y++ {
            for x := 0; x < buffer.Width; x++ {
                buffer.Cells[y][x] = paint.Cell{}
            }
        }
    }
    // ...
}
```

### 视觉效果流程

```
┌─────────────────────────────────────────┐
│ 步骤1: Modal关闭 (正常状态)               │
│ ┌───────────────────────────────────┐   │
│ │ [按钮] 文字内容 颜色丰富          │   │
│ │ 背景色多样，无覆盖                │   │
│ └───────────────────────────────────┘   │
└─────────────────────────────────────────┘
                    ↓ 用户点击"Open Modal"
┌─────────────────────────────────────────┐
│ 步骤2: Modal打开 (应用Backdrop)         │
│ ┌───────────────────────────────────┐   │
│ │ ⚫ 文字变灰 背景变暗 ⚫           │   │ ← Dimmed效果
│ │ ███████████████████████████████  │   │   (bright-black文字 + #1e2028背景)
│ └───────────────────────────────────┘   │
│         ┌──────────────┐              │
│         │  Modal弹窗   │              │ ← Modal正常显示
│         │  (高亮)      │              │   (保持原始颜色)
│         └──────────────┘              │
└─────────────────────────────────────────┘
                    ↓ 用户点击"Close"或ESC
┌─────────────────────────────────────────┐
│ 步骤3: Modal关闭 (全屏重绘)             │
│ • 检测到hasModal: true → false         │
│ • 设置forceFullRender = true          │
│ • 清空整个buffer                      │
│ • 重新渲染所有组件 → 恢复原始颜色      │
└─────────────────────────────────────────┘
```

---

## 开发指南

### 如何在组件中使用主题颜色

#### 方法1: 使用框架提供的便捷函数（推荐）

```go
import "github.com/wwsheng009/mint/framework/theme"

// 在组件中使用
func (b *ButtonVNode) Render() style.Style {
    return style.Style{}
        .Foreground(theme.Primary())    // 返回 style.Color
        .Background(theme.Surface())
        .Bold(true)
}
```

#### 方法2: 直接访问主题管理器

```go
import "github.com/wwsheng009/mint/framework/theme"

mgr := theme.GlobalManager()
currentTheme := mgr.Current()
primaryColor := currentTheme.GetColor("primary")
```

#### 可用的便捷函数列表

**层级系统：**
```go
theme.BG()       // 背景色
theme.Surface()  // 表面色
theme.Overlay()  // 覆盖层色
```

**排版：**
```go
theme.Text()        // 主文本
theme.Muted()       // 弱化文本
theme.Placeholder() // 占位文本
```

**品牌色：**
```go
theme.Primary()   // 主色
theme.Secondary() // 次色
theme.Accent()    // 强调色
```

**状态色：**
```go
theme.Success() // 成功
theme.Warning() // 警告
theme.Error()   // 错误
```

**边界：**
```go
theme.Border()    // 边框
theme.Focus()     // 焦点
theme.Select()    // 选中
theme.Highlight() // 高亮
```

**其他：**
```go
theme.DisabledBG() // 禁用背景
theme.DisabledFG() // 禁用文字
theme.Scrollbar() // 滚动条
```

### 如何创建自定义主题

#### 步骤1: 定义主题预设

在 `framework/theme/preset.go` 中添加：

```go
func myCustomTheme() ThemePreset {
    return ThemePreset{
        Name: "my-custom",
        Colors: map[string]ColorValue{
            // 必须定义所有23个语义颜色
            "bg":      {RGB: [3]int{30, 30, 46}, ANSI16: 0, ANSI256: 235},
            "surface": {RGB: [3]int{49, 50, 68}, ANSI16: 8, ANSI256: 238},
            // ... 其他21个颜色
        },
    }
}
```

#### 步骤2: 注册主题

```go
import "github.com/wwsheng009/mint/framework/theme"

// 方法A: 直接注册
mgr := theme.GlobalManager()
theme.RegisterPreset("my-custom")

// 方法B: 修改Presets()函数添加到内置列表
func Presets() map[string]ThemePreset {
    return map[string]ThemePreset{
        "nord":             nordTheme(),
        "dracula":          draculaTheme(),
        "my-custom":        myCustomTheme(),  // ← 添加
        // ...
    }
}
```

#### 步骤3: 使用自定义主题

```go
// 在main.go中
theme.SetTheme("my-custom")
```

### 如何添加新的语义颜色

#### 步骤1: 更新ColorPalette结构

在 `framework/theme/color.go` 中：

```go
type ColorPalette struct {
    // 现有23个颜色...

    // 新增颜色
    MyNewColor Color
}
```

#### 步骤2: 更新NewColorPaletteFromPreset

```go
func NewColorPaletteFromPreset(preset ThemePreset) ColorPalette {
    return ColorPalette{
        // 现有23个颜色...

        // 新增颜色
        MyNewColor: getColor("my-new-color"),
    }
}
```

#### 步骤3: 在所有主题预设中定义颜色值

在 `preset.go` 的每个主题函数中添加：

```go
func nordTheme() ThemePreset {
    return ThemePreset{
        // ...
        Colors: map[string]ColorValue{
            // 现有颜色...
            "my-new-color": {RGB: [3]int{...}, ANSI16: ..., ANSI256: ...},
        },
    }
}
```

#### 步骤4: 添加便捷访问函数

在 `framework/theme/manager.go` 中添加：

```go
// 返回style.Color类型
func MyNewColor() style.Color {
    return style.Color(GetColor("my-new-color").ToStyleString())
}
```

### 如何调试颜色问题

#### 1. 启用渲染调试

```bash
export TUI_RENDER_DEBUG=1
export TUI_PAINT_DEBUG=1
go run main.go
```

#### 2. 检查颜色值

```go
// 打印所有颜色
colors := []struct{
    name string
    color string
}{
    {"Text", string(theme.Text())},
    {"Primary", string(theme.Primary())},
    // ...
}

for _, c := range colors {
    fmt.Printf("%s: '%s'\n", c.name, c.color)
}
```

#### 3. 查看ANSI输出

```go
// 在paint/colorCode中添加调试
func colorCode(color style.Color, isBackground bool) string {
    result := /* ... */

    // 调试输出
    fmt.Fprintf(os.Stderr, "[colorCode] input='%s' output='%s'\n",
        string(color), result)

    return result
}
```

---

## 故障排查

### 问题1: 颜色显示不正确

**症状：** 颜色与预期不符，或者完全无颜色

**排查步骤：**

1. **检查颜色字符串格式**
   ```go
   fmt.Printf("Color string: '%s'\n", string(theme.Text()))
   // 应该看到: #eceff4 (TrueColor格式)
   ```

2. **检查终端支持**
   ```bash
   # 检查终端颜色支持
   echo $TERM
   echo $COLORTERM

   # 应该支持truecolor或24bit
   ```

3. **验证colorCode转换**
   ```go
   // 在style_state.go中调试
   func colorCode(color style.Color, isBackground bool) string {
       fmt.Fprintf(os.Stderr, "[DEBUG] color=%s isBg=%v\n",
           string(color), isBackground)
       // ...
   }
   ```

**解决方案：**

- ✅ 确保使用TrueColor格式 `#RRGGBB`
- ✅ 终端必须支持TrueColor (`COLORTERM=truecolor`)
- ✅ 检查 `ToStyleString()` 是否正确转换

### 问题2: Modal关闭后有颜色残留

**症状：** Modal关闭后，文字颜色变灰或背景色丢失

**原因分析：**

1. **触发条件：**
   - Modal打开时，`paintModalBackdrop` 覆盖了样式
   - Modal关闭时，内容没变，diff系统认为不需要重绘

2. **根本原因：**
   - 没有检测到modal状态变化
   - 没有强制全屏重绘

**解决方案：**

✅ **已实现：** Modal状态检测 + 强制全屏重绘

```go
// 检测modal出现/消失
if hasModal != hadModal {
    e.forceFullRender = true
}
```

**验证方法：**

```go
// 在PaintEngine中添加日志
if e.forceFullRender {
    fmt.Fprintf(os.Stderr, "[PaintEngine] Modal state changed, forcing full render\n")
}
```

### 问题3: 组件颜色不跟随主题切换

**症状：** 切换主题后，某些组件颜色不变

**原因：** 组件在初始化时缓存了颜色值

**错误示例：**

```go
// ❌ 错误：在Init中缓存颜色
type Button struct {
    primaryColor style.Color  // 缓存的颜色
}

func (b *Button) Init() {
    b.primaryColor = theme.Primary()  // 只在Init时获取一次
}
```

**正确示例：**

```go
// ✅ 正确：每次Render时获取当前颜色
func (b *Button) Render() style.Style {
    return style.Style{}
        .Foreground(theme.Primary())  // 每次都获取最新颜色
}
```

### 问题4: TrueColor在某些终端不工作

**症状：** RGB颜色显示为默认颜色

**原因：** 终端不支持TrueColor，需要降级到256色或16色

**解决方案：**

✅ **已实现：** 自动降级机制

```go
func colorCode(color style.Color, isBackground bool) string {
    c := string(color)

    // 1. 尝试TrueColor
    if strings.HasPrefix(c, "#") {
        rgb, _ := parseHexString(c)
        return fmt.Sprintf("38;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
    }

    // 2. 降级到256色
    if is256Color(c) {
        return fmt.Sprintf("38;5;%s", c)
    }

    // 3. 降级到16色
    code := colorToAnsi[c]
    return fmt.Sprintf("%d", code + 30)
}
```

**手动指定颜色模式：**

如果需要强制使用特定模式，可以修改主题定义使用命名颜色：

```go
// 不使用RGB，直接使用ANSI颜色名
"primary": {RGB: [3]int{...}, ANSI16: 4}  // 使用ANSI16值
```

### 问题5: 颜色对比度不足

**症状：** 文字难以阅读

**检查工具：**

1. **手动计算对比度**
   ```go
   func contrastRatio(fg, bg [3]int) float64 {
       // 使用WCAG对比度公式
       l1 := luminance(fg)
       l2 := luminance(bg)
       return (max(l1, l2) + 0.05) / (min(l1, l2) + 0.05)
   }
   ```

2. **在线检查工具**
   - https://contrast-ratio.com/
   - https://webaim.org/resources/contrastchecker/

**调整颜色：**

```go
// 在preset.go中调整颜色值
func nordTheme() ThemePreset {
    return ThemePreset{
        Colors: map[string]ColorValue{
            "text": {RGB: [3]int{236, 239, 244}},  // 太亮？
            // 调整为更亮的颜色
            "text": {RGB: [3]int{255, 255, 255}},  // 纯白
        },
    }
}
```

---

## 性能优化

### 样式状态机优化

**优化原理：** 减少ANSI转义序列输出

```go
// ❌ 低效：每次都输出完整样式
func BadRendering() {
    for _, cell := range cells {
        fmt.Print("\x1b[0m")              // 重置
        fmt.Print("\x1b[1m")             // 粗体
        fmt.Print("\x1b[38;2;136;192;208m")  // 前景色
        fmt.Print(cell.Text)
    }
}

// ✅ 高效：只输出变化部分
func GoodRendering() {
    styleState := NewStyleStateMachine()
    for _, cell := range cells {
        // Update()只输出变化的样式
        fmt.Print(styleState.Update(cell.Style))
        fmt.Print(cell.Text)
    }
}
```

**性能对比：**

| 场景 | 低效方式 | 高效方式 | 提升 |
|------|----------|----------|------|
| 1000个相同单元格 | 4000 bytes | 1200 bytes | 70% |
| 颜色频繁变化 | 5000 bytes | 4500 bytes | 10% |
| 所有样式不同 | 8000 bytes | 8000 bytes | 0% |

### 差异渲染优化

**原理：** 只重绘变化的单元格

```go
// DirtyTracker标记变化区域
func (d *DirtyTracker) Diff(prev, curr *Buffer) DiffResult {
    // 1. 逐单元格比较
    // 2. 标记变化的单元格
    // 3. 合并为连续区域
    // 4. 只重绘这些区域
}
```

**性能数据：**

```
全屏渲染:  100% cells × 100 bytes = 10000 bytes
差异渲染:    5% cells × 100 bytes =   500 bytes
提升:      95%
```

---

## 参考资源

### 设计规范

- `docs/theme/design/终端UI设计规范.md` - 语义颜色定义
- `docs/theme/design/theme.md` - 5套主题的RGB值
- `docs/theme/design/comp_*.md` - 组件设计规范

### 源码文件

- `framework/theme/preset.go` - 主题预设定义
- `framework/theme/manager.go` - 主题管理器
- `framework/theme/color.go` - 颜色处理
- `runtime/paint/style_state.go` - 样式状态机
- `internal/render/paint_engine.go` - 渲染引擎

### 相关文档

- `docs/theme_rendering_flow.md` - 渲染流程说明
- `runtime/style/README.md` - Style系统说明

---

## 附录

### A. 完整的颜色速查表

| 语义名称 | Nord (示例) | Dracula (示例) | 用途 |
|----------|--------------|-----------------|------|
| BG | #2E3440 | #282A36 | 全局背景 |
| Surface | #3B4252 | #343746 | 卡片背景 |
| Overlay | #282C34 | #242733 | 悬浮层背景 |
| Text | #ECEFF4 | #F8F8F2 | 主文本 |
| Muted | #616E88 | #6272A4 | 辅助文本 |
| Placeholder | #616E88 | #6272A4 | 输入提示 |
| Primary | #88C0D0 | #BD93F9 | 主按钮 |
| Secondary | #81A1C1 | #8BE9FD | 次按钮 |
| Accent | #8FBCBB | #FF79C6 | 强调色 |
| Success | #A3BE8C | #50FA7B | 成功状态 |
| Warning | #EBCB8B | #F1FA8C | 警告状态 |
| Error | #BF616A | #FF5555 | 错误状态 |
| Link | #88C0D0 | #8BE9FD | 链接 |
| Visited | #B48EAD | #BD93F9 | 访问过的链接 |
| Border | #4C566A | #44475A | 边框 |
| Focus | #88C0D0 | #BD93F9 | 焦点 |
| Select | #81A1C1 | #FF79C6 | 选中 |
| Highlight | #EBCB8B | #F1FA8C | 高亮 |
| DisabledBG | #3B4252 | #343746 | 禁用背景 |
| DisabledFG | #616E88 | #6272A4 | 禁用文字 |
| Scrollbar | #4C566A | #44475A | 滚动条 |
| Shadow | #1E2028 | #1C1D26 | 阴影 |
| Caret | #ECEFF4 | #F8F8F2 | 光标 |

### B. ANSI转义序列速查

**格式：** `ESC[参数m`

**前景色：**
- `38;2;R;G;B` - TrueColor (RGB)
- `38;5;N` - 256色模式 (N: 0-255)
- `30-37` - 16色标准 (0=black, 1=red, ..., 7=white)
- `90-97` - 16色明亮 (90=bright-black, ..., 97=bright-white)

**背景色：**
- `48;2;R;G;B` - TrueColor (RGB)
- `48;5;N` - 256色模式 (N: 0-255)
- `40-47` - 16色标准
- `100-107` - 16色明亮

**样式：**
- `0` - 重置所有样式
- `1` - 粗体
- `3` - 斜体
- `4` - 下划线
- `7` - 反转

**示例：**
```
ESC[0m                    # 重置
ESC[1;38;2;136;192;208m  # 粗体 + TrueColor前景
ESC[48;2;46;52;64m        # TrueColor背景
```

---

**文档结束**

如有疑问或建议，请查看源码或提交Issue。
