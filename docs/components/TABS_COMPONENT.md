# Tabs 组件文档

## 概述

`Tabs` 是一个轻量级导航组件，用来在多个逻辑视图之间切换“当前选中项”。

当前实现聚焦于：

- 选项卡栏渲染
- 键盘 / 鼠标切换
- Intent 集成
- 受控激活状态
- 顶部 / 底部 / 左侧 / 右侧布局

当前实现**不负责**：

- 自动管理 tab 对应的内容面板
- 拖拽重排
- 关闭按钮
- 超长 tab 列表滚动

如果你需要内容面板，通常做法是：

1. 用 `tabs` 维护“当前选中 tab”
2. 在父组件里根据当前 tab 渲染内容区

---

## 快速开始

```go
package main

import (
    "github.com/wwsheng009/mint/ui"
)

func View() ui.VNode {
    items := []ui.TabItem{
        ui.NewTabItem("home", "Home").WithIcon("H").WithHotkey('h'),
        ui.NewTabItem("search", "Search").WithIcon("S").WithHotkey('s'),
        ui.NewTabItem("alerts", "Alerts").WithBadge("12"),
        ui.NewTabItem("locked", "Locked").WithDisabled(true),
    }

    return ui.NewTabsBuilder().
        ComponentID("workspace-tabs").
        Tabs(items).
        ActiveTabID("alerts").
        ShowHotkeys(true).
        LoopNavigation(true).
        Width(48).
        Build()
}
```

渲染效果类似：

```text
{H} H Home | {S} S Search | [Alerts (12)] | Locked
```

---

## TabItem

`TabItem` 定义：

```go
type TabItem struct {
    ID       string
    Label    string
    Icon     string
    Badge    string
    Hotkey   rune
    Disabled bool
    Hidden   bool
}
```

推荐通过 helper 构造：

```go
tab := ui.NewTabItem("metrics", "Metrics").
    WithIcon("M").
    WithBadge("99+").
    WithHotkey('m')
```

字段语义：

- `ID`: 业务标识，建议唯一
- `Label`: 显示文本
- `Icon`: 标签前缀
- `Badge`: 标签后缀提示
- `Hotkey`: 无修饰键快捷切换
- `Disabled`: 渲染但不可切换
- `Hidden`: 不渲染，也不会参与导航

---

## Builder API

常用构建方式：

```go
tabs := ui.NewTabsBuilder().
    AddTab("home", "Home").
    AddTab("settings", "Settings").
    Build()
```

或者：

```go
tabs := ui.NewTabsBuilder().
    Tabs([]ui.TabItem{
        ui.NewTabItem("home", "Home"),
        ui.NewTabItem("settings", "Settings"),
    }).
    Build()
```

### 常用配置

| 方法 | 说明 |
|------|------|
| `Tabs([]TabItem)` | 批量设置 tab |
| `AddTab(id, label)` | 添加普通 tab |
| `AddTabItem(tab)` | 添加完整 `TabItem` |
| `AddTabWithOptions(id, label, disabled)` | 添加含禁用态的 tab |
| `ComponentID(id)` | 设置 Intent 路由 ID |
| `ActiveTab(index)` | 通过索引声明激活 tab |
| `ActiveTabID(id)` | 通过 ID 声明激活 tab |
| `Position(pos)` | 设置位置 |
| `Top()/Bottom()/Left()/Right()` | 位置快捷方法 |
| `WrapTabs(true)` | 启用水平换行 |
| `TabGap(n)` | 垂直布局间距 |
| `Divider(text)` | 设置水平分隔符 |
| `LoopNavigation(true)` | 启用循环导航 |
| `ShowHotkeys(true)` | 显示热键提示 |
| `Style(s)` | 普通 tab 样式 |
| `ActiveTabStyle(s)` | 激活 tab 样式 |
| `DisabledTabStyle(s)` | 禁用 tab 样式 |
| `OnChange(intent)` | 切换时发出自定义 Intent |
| `FieldIntent(binding)` | 切换时发出 `FieldChangeIntent` |
| `Width/Height/Flex/Size` | 布局配置 |

### 位置枚举

```go
const (
    TabPositionTop TabPosition = iota
    TabPositionBottom
    TabPositionLeft
    TabPositionRight
)
```

说明：

- `Top` / `Bottom`: 水平布局
- `Left` / `Right`: 垂直布局
- `Bottom` / `Right` 会基于组件尺寸进行偏移，不再固定绘制在第 0 行

---

## 受控状态

`tabs` 支持两种声明式激活方式：

```go
ui.NewTabsBuilder().
    ActiveTab(2)
```

或者：

```go
ui.NewTabsBuilder().
    ActiveTabID("settings")
```

优先级：

1. `ActiveTabID`
2. `ActiveTab`
3. 现有运行时状态
4. 第一个可选 tab

如果声明的 tab 不可选（隐藏或禁用），会自动回退到第一个可选 tab。

---

## 运行时 API

以下方法在 `tabs.Instance` 上可用：

### 切换

```go
inst.SetActiveTab(1)
inst.SetActiveTabByID("settings")
inst.SetActiveTabByLabel("Settings")
inst.SetActiveVisibleOrdinal(0)
inst.NextTab()
inst.PreviousTab()
inst.FirstTab()
inst.LastTab()
```

### 查询

```go
active := inst.GetActiveTab()
id := inst.GetActiveTabID()
label := inst.GetActiveTabLabel()

count := inst.GetTabCount()
visibleCount := inst.GetVisibleTabCount()

canNext := inst.CanGoNext()
canPrev := inst.CanGoPrevious()

idx := inst.FindTabByID("settings")
idx2 := inst.FindTabByLabel("Settings")

tab, ok := inst.GetTabByIndex(0)
_ = tab
_ = ok
```

### 状态控制

```go
inst.SetTabEnabled(2, false)
enabled := inst.IsTabEnabled(2)

pos := inst.GetPosition()
inst.SetPosition(pos)
```

行为说明：

- 禁用当前激活 tab 时，会自动回退到第一个可选 tab
- `Hidden` tab 不参与绘制，也不参与导航
- `Disabled` tab 会绘制，但不可选

---

## Intent 集成

`tabs` 的 Intent 分成两类：状态变化事件和命令式切换。

### 1. 自动发出的状态变化 Intent

如果设置了 `ComponentID`，切换时会自动发出：

```go
tabs.TabChangeIntent{
    ComponentID: "workspace-tabs",
    ActiveTab:   1,
    TabID:       "settings",
    TabLabel:    "Settings",
}
```

如果设置了 `FieldIntent(...)`，还会额外发出：

```go
intent.FieldChangeIntent{
    Field: "ActiveTab",
    Value: "1",
}
```

如果设置了 `OnChange(customIntent)`，则会发出你传入的自定义 Intent。

### 2. 命令式切换 Intent

```go
ui.TabsNext("workspace-tabs")
ui.TabsPrevious("workspace-tabs")
ui.TabsSelect("workspace-tabs", "settings")
ui.TabsSelectIndex("workspace-tabs", 2)
```

底层对应：

- `tabs.TabNextIntent`
- `tabs.TabPreviousIntent`
- `tabs.TabSelectIntent`

---

## 键盘与鼠标行为

默认支持：

- `Left` / `Up`: 切到上一个可选 tab
- `Right` / `Down`: 切到下一个可选 tab
- `Home`: 第一个可选 tab
- `End`: 最后一个可选 tab
- `Ctrl+Tab`: 下一个 tab
- `Ctrl+Shift+Tab`: 上一个 tab
- `1-9`: 选中第 N 个**可见且可选** tab
- `Hotkey`: 选中匹配 `TabItem.Hotkey` 的 tab
- 鼠标点击：切换到点击的 tab

说明：

- 热键不需要修饰键
- 数字选择按“可见且可选顺序”计算
- 换行后的 tab 和带偏移的位置也能正确点击命中

---

## 样式

可分别设置三种样式：

```go
ui.NewTabsBuilder().
    Style(style.NewStyle().Foreground(style.White)).
    ActiveTabStyle(style.NewStyle().Foreground(style.Cyan).Bold(true)).
    DisabledTabStyle(style.NewStyle().Foreground(style.BrightBlack))
```

默认行为：

- 普通 tab: 白色
- 激活 tab: 青色 + 粗体
- 禁用 tab: 亮黑色

激活态始终会自动加 `[]` 包裹。

---

## 示例：工具面板导航

```go
package main

import (
    "github.com/wwsheng009/mint/runtime/style"
    "github.com/wwsheng009/mint/ui"
)

func ToolTabs() ui.VNode {
    return ui.NewTabsBuilder().
        ComponentID("tool-tabs").
        Tabs([]ui.TabItem{
            ui.NewTabItem("files", "Files").WithIcon("F").WithHotkey('f'),
            ui.NewTabItem("search", "Search").WithIcon("S").WithHotkey('s'),
            ui.NewTabItem("git", "Git").WithIcon("G").WithBadge("2"),
            ui.NewTabItem("trace", "Trace").WithIcon("T").WithDisabled(true),
        }).
        ActiveTabID("git").
        ShowHotkeys(true).
        Divider(" / ").
        LoopNavigation(true).
        Width(56).
        ActiveTabStyle(style.NewStyle().Foreground(style.Cyan).Bold(true)).
        DisabledTabStyle(style.NewStyle().Foreground(style.BrightBlack)).
        Build()
}
```

---

## `ui` 层快捷入口

`ui` 包已提供以下 helper：

```go
item := ui.NewTabItem("home", "Home")

next := ui.TabsNext("workspace-tabs")
prev := ui.TabsPrevious("workspace-tabs")
sel1 := ui.TabsSelect("workspace-tabs", "settings")
sel2 := ui.TabsSelectIndex("workspace-tabs", 2)
```

---

## 当前限制

1. 不负责 tab 内容面板管理，只负责 tab 栏本身
2. 不支持拖拽重排
3. 不支持关闭按钮
4. 不支持 overflow 滚动，水平模式下只能单行或换行
5. `Hidden` tab 不参与导航，不能通过数字键直接选中

---

## 验证

已覆盖的关键测试包括：

- 受控激活：`ActiveTab` / `ActiveTabID`
- 换行布局与本地点击命中
- `Bottom` / `Right` 定位
- 热键、`Ctrl+Tab`、循环导航
- `FieldIntent` 提取与发出
- `TabSelectIntent`

运行：

```bash
go test ./ui/components/tabs
```

---

**状态**: 已更新  
**文档日期**: 2026-03-12
