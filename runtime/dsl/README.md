# DSL Support

DSL（领域特定语言）支持层，为运行时提供声明式配置和解析能力。

## 职责

- **颜色 DSL 解析**：支持多种颜色格式（命名颜色、十六进制、RGB、ANSI 代码）转换为终端可用的 ANSI 颜色
- **语义化颜色映射**：提供预设的语义化颜色（primary、success、warning 等）
- **主题配色支持**：支持流行配色方案（Tokyo Night、Dracula、Nord 等）
- **声明式配置接口**：定义 DSL 解析接口规范（具体实现由 framework 层提供）

## 核心概念

### 颜色格式支持

`ColorNameToANSI` 函数支持以下颜色格式：

1. **命名颜色**：`"red"`, `"blue"`, `"green"` 等
2. **亮色**：`"brightRed"`, `"brightBlue"` 等
3. **十六进制**：`"#FF5733"`, `"ff5733"`
4. **RGB 格式**：`"rgb(255, 87, 51)"`
5. **ANSI 数字**：`"214"`, `"57"`
6. **十六进制简写**：`"F53"` → `"#FF5533"`

### 颜色分类

#### 基础颜色（ANSI 0-7）
- `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`
- 默认别名：`default: 0`, `gray/darkgray: 8`, `lightgray: 7`

#### 亮色（ANSI 8-15）
- `brightblack`, `brightred`, `brightgreen`, `brightblue`, 等
- 别名形式：`lightblack`, `lightred`（与 bright 互为别名）

#### 常用扩展颜色
- `orange: 208`, `purple: 127`, `pink: 204`, `brown: 130`
- `lime: 10`, `indigo: 54`, `violet: 213`, `gold: 220`

#### 语义化颜色
- **状态色**：`primary`, `secondary`, `success`, `info`, `warning`, `danger`, `error`, `critical`
- **UI 元素色**：`foreground`, `background`, `muted`, `border`
- **表格专用色**：`header`, `row`, `alternate`, `hover`
- **文本色**：`text`, `textdark`, `textlight`, `textmuted`

#### 主题配色方案
- **Tokyo Night**：`tokyo-night-blue`, `tokyo-night-cyan`, `tokyo-night-green`, 等
- **Dracula**：`dracula-purple`, `dracula-pink`, `dracula-red`, 等
- **Nord**：`nord-blue`, `nord-cyan`, `nord-green`, `nord-yellow`, 等

### 设计哲学

虽然此目录涉及 DSL，但运行时层应只定义基础接口和数据类型，具体的高级 DSL 解析（如完整的 UI 配置解析）由 framework 层提供。

**当前实现**：仅提供颜色相关的 DSL 功能，遵循"最小必要"原则。

## 使用示例

### 基础颜色转换

```go
import "yourproject/runtime/dsl"

// 命名颜色
ansiColor := dsl.ColorNameToANSI("red")        // "1"
ansiColor = dsl.ColorNameToANSI("red")         // "1"
ansiColor = dsl.ColorNameToANSI("brightgreen") // "10"

// 十六进制颜色
ansiColor = dsl.ColorNameToANSI("#FF5733")     // "#FF5733"
ansiColor = dsl.ColorNameToANSI("ff5733")      // "ff5733"

// RGB 格式
ansiColor = dsl.ColorNameToANSI("rgb(255, 87, 51)") // "rgb(255, 87, 51)"

// ANSI 数字（直接返回）
ansiColor = dsl.ColorNameToANSI("214")        // "214"

// 语义化颜色
ansiColor = dsl.ColorNameToANSI("danger")     // "196" (红色)
ansiColor = dsl.ColorNameToANSI("primary")    // "21" (蓝色)
ansiColor = dsl.ColorNameToANSI("tokyo-night-blue") // "39"
```

### 与 lipgloss 集成

```go
import (
    "github.com/charmbracelet/lipgloss"
    "yourproject/runtime/dsl"
)

// 解析为 lipgloss.Color
fgColor := dsl.ParseColorStyle("primary")     // lipgloss.Color("21")
bgColor := dsl.ParseColorStyle("#333333")     // lipgloss.Color("#333333")
successColor := dsl.ParseColorStyle("success") // lipgloss.Color("34")

// 在样式中使用
style := lipgloss.NewStyle().
    Foreground(fgColor).
    Background(bgColor).
    Bold(true)
```

### 主题配色使用

```go
// 使用主题配色方案
headerColor := dsl.ColorNameToANSI("tokyo-night-blue") // "39"
draculaPink := dsl.ColorNameToANSI("dracula-pink")     // "212"
nordCyan := dsl.ColorNameToANSI("nord-cyan")           // "109"

// 表格样式示例
headerStyle := lipgloss.NewStyle().
    Foreground(dsl.ParseColorStyle("header")).
    Background(dsl.ParseColorStyle("background")).
    Bold(true)

alternateRowStyle := lipgloss.NewStyle().
    Foreground(dsl.ParseColorStyle("row")).
    Background(dsl.ParseColorStyle("alternate"))
```

### 配置文件颜色映射

```go
// YAML 配置示例
/*
type ThemeConfig struct {
    Primary   string `yaml:"primary"`
    Secondary string `yaml:"secondary"`
    Danger    string `yaml:"danger"`
    Warning   string `yaml:"warning"`
    Success   string `yaml:"success"`
}

// 配置文件 (config.yaml)
primary: "primary"      # 使用语义化颜色
secondary: "secondary"
danger: "danger"
warning: "#FFA500"      # 直接使用十六进制
success: "success"
*/

// 解析配置
func (c *ThemeConfig) Parse() *ThemeColors {
    return &ThemeColors{
        Primary:   dsl.ParseColorStyle(c.Primary),
        Secondary: dsl.ParseColorStyle(c.Secondary),
        Danger:    dsl.ParseColorStyle(c.Danger),
        Warning:   dsl.ParseColorStyle(c.Warning),
        Success:   dsl.ParseColorStyle(c.Success),
    }
}
```

## 核心类型

### ColorNameToANSI(color string) string

将颜色名称或代码转换为 ANSI 颜色格式。支持格式：
- ANSI 数字： `"0"`, `"255"`
- 十六进制： `"#FF5733"`（保持原样）
- RGB： `"rgb(255, 87, 51)"`（保持原样）
- 命名颜色： `"red"`, `"brightred"`, `"primary"` 等

**处理逻辑**：
1. 如果已为 ANSI 数字 → 直接返回
2. 如果为十六进制格式（`#` 开头）→ 直接返回
3. 如果为 RGB 格式 → 直接返回
4. 在颜色映射表中查找 → 返回对应的 ANSI 代码
5. 用 lipgloss 尝试解析 → 如果成功，返回原值
6. 默认返回 `"15"`（白色）

### ParseColorStyle(color string) lipgloss.Color

将颜色字符串解析为 `lipgloss.Color` 类型，用于与 lipgloss 样式系统集成。

```go
func ParseColorStyle(color string) lipgloss.Color
```

**实现**：内部调用 `ColorNameToANSI` 后转换类型。

### isANSICode(s string) bool

检查字符串是否为 ANSI 颜色代码（0-255）。仅包含数字字符时返回 true。

### 全局颜色映射表

- **basicColors**: 基础 8 色（ANSI 0-7）映射表
- **brightColors**: 亮色（ANSI 8-15）映射表 + 常用扩展颜色
- **semanticColors**: 语义化颜色映射表 + 主题配色方案

## 文件结构

```
runtime/dsl/
├── README.md       # 本文档
└── colors.go       # 颜色 DSL 解析实现
    ├── ColorNameToANSI       # 主转换函数
    ├── ParseColorStyle       # lipgloss 集成
    ├── isANSICode            # 校验函数
    ├── basicColors           # 基础颜色映射
    ├── brightColors          # 亮色映射
    └── semanticColors        # 语义化颜色映射
```

## 依赖

### 外部依赖
- **github.com/charmbracelet/lipgloss**: 用于 `ParseColorStyle` 的类型返回和颜色验证

**注意**：虽然 README 中的"纯 Go 约束"指出不应依赖 lipgloss，但当前实现因与框架层的 lipgloss 样式系统集成需要此依赖。这是 DSL 层与框架层交互的边界示例。

### 内部依赖
- 无（DSL 层为独立的基础设施层）

## 与其他模块集成

### 与 Paint 模块集成

虽然颜色 DSL 提供颜色名称到 ANSI 的转换机制，但 `runtime/paint` 层使用纯 Go 的 `Cell` 结构和内联样式（前景色、背景色、样式位），不直接使用 lipgloss。DSL 颜色主要用于：

1. **主题配置**：在配置文件中使用语义化颜色名称
2. **框架层样式**：framework 层使用 lipgloss 时的颜色源
3. **运行时配置**：通过运行时 API 动态调整主题色

### 与 Framework 层集成

DSL 层为 framework 层提供基础的颜色解析能力：

```go
// 在 framework/component 或 framework/style 中使用
import "yourproject/runtime/dsl"

// 解析用户配置的颜色
userColor := dsl.ParseColorStyle(themeConfig.PrimaryColor)
componentStyle := lipgloss.NewStyle().Foreground(userColor)
```

## 最佳实践

### 1. 优先使用语义化颜色

```go
// 推荐使用语义化颜色，便于主题切换
foregroundColor := dsl.ColorNameToANSI("danger")     // 基于语义
foregroundColor := dsl.ColorNameToANSI("tokyo-night-blue") // 基于主题
```

### 2. 配置中声明颜色，运行时解析

```go
// 配置：使用可读性高的颜色名称
config := struct {
    SuccessColor string `yaml:"successColor"`
    WarningColor string `yaml:"warningColor"`
}{
    SuccessColor: "success",  // 而非 "34"
    WarningColor: "warning",  // 而非 "208"
}
```

### 3. 处理颜色转换错误

```go
func parseColor(input string) lipgloss.Color {
    color := dsl.ColorNameToANSI(input)
    // ColorNameToANSI 总是返回有效值（默认为 "15"）
    // 如果需要严格验证，可自行添加检查
    if color == "15" && input != "white" && input != "15" {
        // 记录警告：未知颜色
        log.Warnf("Unknown color '%s', falling back to white", input)
    }
    return dsl.ParseColorStyle(input)
}
```

### 4. 主题配色方案使用

```go
// 定义主题常量，便于统一切换
const (
    TokyoNightTheme = "tokyo-night-blue"
    DraculaTheme    = "dracula-purple"
    NordTheme       = "nord-cyan"
)

// 运行时动态切换主题
func SetTheme(theme string) {
    accentColor := dsl.ColorNameToANSI(theme)
    // 应用主题色到组件...
}
```

## 常见问题

### Q: 为什么 DSL 层会依赖 lipgloss，违反"纯 Go 约束"？

A: 这是一个设计权衡。运行时核心（runtime/paint, runtime/render 等）保持纯 Go 实现，确保无外部依赖。DSL 层作为与 framework 层的桥梁，少量使用 lipgloss 以提供类型兼容性（`lipgloss.Color`）。未来可以考虑引入接口抽象，避免直接依赖。

### Q: 如何添加自定义颜色映射？

A: 可以在语义化颜色映射表中添加自定义颜色：

```go
// 在 colors.go 中扩展 semanticColors
var semanticColors = map[string]string{
    // ... 现有映射
    "accent":      "99",   // 自定义强调色
    "customGreen": "118",  // 自定义绿色
}
```

### Q: ColorNameToANSI 为什么总返回有效值？

A: 为了提供容错性。当输入无法识别时，返回默认颜色 `"15"`（白色），避免颜色解析失败导致渲染错误。如果需要严格验证，可在调用方添加检查逻辑。

### Q: 如何支持 24-bit 真彩色（RGB）？

A: `ColorNameToANSI` 接受 RGB 格式且直接返回原值（`rgb(255, 87, 51)`），这可以由 lipgloss 解析为 24-bit 真彩色。对于自定义的 24-bit 颜色，直接使用 RGB 或十六进制格式即可：

```go
customColor := dsl.ParseColorStyle("rgb(120, 100, 80)")
hexColor := dsl.ParseColorStyle("#78C650")
```

### Q: 是否支持 HSL 或其他颜色格式？

A: 目前暂不支持 HSL、HSLA 等格式。可通过扩展 `ColorNameToANSI` 函数，添加 HSL 到 RGB 的转换逻辑：

```go
// 示例：扩展支持 HSL 格式
if strings.HasPrefix(color, "hsl(") {
    rgb := hslToRGB(color)      // 自定义转换函数
    return fmt.Sprintf("rgb(%d, %d, %d)", rgb.R, rgb.G, rgb.B)
}
```
