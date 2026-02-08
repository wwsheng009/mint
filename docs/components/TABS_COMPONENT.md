# Tabs 组件完整文档

**Tabs Component - Complete Documentation**

---

## 概述

`Tabs` 组件实现了类似 Web 浏览器或 IDE 的选项卡功能，允许用户在多个面板之间切换显示。这是 TUI 应用中组织和展示内容的常见模式。

### 主要特性

- ✅ **简单易用的 API** - Builder 模式，链式调用
- ✅ **灵活的位置** - 支持 tabs 在上、下、左、右
- ✅ **键盘导航** - 支持快捷键切换标签
- ✅ **禁用标签** - 可以禁用特定标签
- ✅ **回调支持** - 标签切换时触发回调
- ✅ **样式自定义** - 激活和非激活标签分别设置样式

---

## 基本用法

### 最简单的示例

```go
import (
    "github.com/wwsheng009/mint/components/navigation"
    ui "github.com/wwsheng009/mint/ui"
)

func MyComponent() ui.VNode {
    // 创建标签内容
    content1 := ui.Text("这是标签页 1 的内容")
    content2 := ui.Text("这是标签页 2 的内容")
    content3 := ui.Text("这是标签页 3 的内容")

    // 创建 tabs 组件
    tabs := navigation.NewTabs().
        AddTab("tab1", "首页").
        AddTab("tab2", "设置").
        AddTab("tab3", "关于").
        Content("tab1", content1).
        Content("tab2", content2).
        Content("tab3", content3).
        ActiveTab(0).
        Build()

    return tabs
}
```

### 使用 TabItem 结构

```go
func MyComponent() ui.VNode {
    tabs := []*navigation.TabItem{
        {ID: "home", Label: "首页", Content: ui.Text("首页内容")},
        {ID: "settings", Label: "设置", Content: ui.Text("设置内容")},
        {ID: "about", Label: "关于", Content: ui.Text("关于内容")},
    }

    return navigation.NewTabsBuilder(tabs).
        ActiveTab(0).
        Build()
}
```

---

## Builder API

### 创建 Tabs

```go
// 方式 1: 从头创建
tabs := navigation.NewTabs().
    AddTab("id1", "Tab 1").
    AddTab("id2", "Tab 2").
    Content("id1", content1).
    Content("id2", content2).
    Build()

// 方式 2: 使用 TabItem 切片
tabs := navigation.NewTabsBuilder([]*navigation.TabItem{
    {ID: "tab1", Label: "First", Content: content1},
    {ID: "tab2", Label: "Second", Content: content2},
}).
    Build()
```

### 配置方法

| 方法 | 参数 | 说明 | 默认值 |
|------|------|------|--------|
| `AddTab(id, label)` | id: 唯一ID, label: 显示文本 | 添加一个标签页 | - |
| `Content(id, content)` | id: 标签ID, content: VNode | 设置标签内容 | - |
| `ActiveTab(index)` | index: 标签索引 | 设置激活的标签 | 0 |
| `Position(pos)` | pos: 位置枚举 | 设置标签位置 | TabPositionTop |
| `Width(n)` | n: 宽度 | 设置总宽度 | 自动 |
| `Height(n)` | n: 高度 | 设置内容高度 | 自动 |
| `OnChange(fn)` | fn: 回调函数 | 设置切换回调 | nil |
| `Key(key)` | key: diff 键 | 设置键 | 自动 |
| `Vertical(v)` | v: 是否垂直 | **已废弃**，使用 Position | false |

### TabPosition 枚举

```go
const (
    TabPositionTop    TabPosition = iota  // 标签在内容上方
    TabPositionBottom                       // 标签在内容下方
    TabPositionLeft                         // 标签在内容左侧
    TabPositionRight                        // 标签在内容右侧
)
```

### 示例：不同位置的 Tabs

```go
// 标签在顶部（默认）
tabsTop := navigation.NewTabs().
    AddTab("t1", "Tab 1").
    Position(navigation.TabPositionTop).
    Build()

// 标签在底部
tabsBottom := navigation.NewTabs().
    AddTab("t1", "Tab 1").
    Position(navigation.TabPositionBottom).
    Build()

// 标签在左侧（垂直）
tabsLeft := navigation.NewTabs().
    AddTab("t1", "Tab 1").
    Position(navigation.TabPositionLeft).
    Build()

// 标签在右侧（垂直）
tabsRight := navigation.NewTabs().
    AddTab("t1", "Tab 1").
    Position(navigation.TabPositionRight).
    Build()
```

---

## 运行时 API

### 标签切换

```go
// 通过索引切换
tabs.SetActiveTab(2)

// 通过标签切换
ok := tabs.SetActiveTabByLabel("设置")
if !ok {
    // 标签不存在
}

// 通过 ID 切换
ok = tabs.SetActiveTabByID("settings")
if !ok {
    // ID 不存在
}
```

### 导航方法

```go
// 下一个标签
tabs.NextTab()

// 上一个标签
tabs.PreviousTab()

// 第一个标签
tabs.FirstTab()

// 最后一个标签
tabs.LastTab()
```

### 查询方法

```go
// 获取当前激活标签索引
activeIndex := tabs.ActiveTab()

// 获取当前激活标签的文本
label := tabs.GetActiveTabLabel()  // "首页"

// 获取当前激活标签的 ID
id := tabs.GetActiveTabID()  // "home"

// 获取当前激活标签的内容
content := tabs.GetActiveTabContent()

// 获取标签总数
count := tabs.GetTabCount()

// 检查是否可以向前/向后导航
canNext := tabs.CanGoNext()
canPrev := tabs.CanGoPrevious()

// 检查标签是否启用
enabled := tabs.IsTabEnabled(2)

// 查找标签索引
idx := tabs.FindTabByLabel("设置")   // -1 if not found
idx = tabs.FindTabByID("settings")   // -1 if not found

// 获取指定位置的标签
tab := tabs.GetTabByIndex(0)
if tab != nil {
    fmt.Println(tab.Label, tab.ID)
}

// 获取/设置位置
pos := tabs.GetPosition()
tabs.SetPosition(navigation.TabPositionBottom)
```

### 标签管理

```go
// 禁用/启用标签
tabs.SetTabEnabled(2, false)  // 禁用第 3 个标签
tabs.SetTabEnabled(2, true)   // 启用第 3 个标签

// 如果当前标签被禁用，会自动切换到第一个可用标签
```

---

## 高级用法

### 示例 1: Inspector 标签页

```go
// internal/inspector/standalone_inspector.go

func (si *StandaloneInspector) buildOverlayContent() rtui.VNode {
    // 创建各个标签页的内容
    elementsContent := si.buildElementsTab()
    consoleContent := si.buildConsoleTab()
    performanceContent := si.buildPerformanceTab()
    diagnosticsContent := si.buildDiagnosticsTab()
    networkContent := si.buildNetworkTab()

    // 创建 tabs 组件
    tabs := navigation.NewTabs().
        AddTab("elements", "Elements").
        AddTab("console", "Console").
        AddTab("performance", "Performance").
        AddTab("diagnostics", "Diagnostics").
        AddTab("network", "Network").
        Content("elements", elementsContent).
        Content("console", consoleContent).
        Content("performance", performanceContent).
        Content("diagnostics", diagnosticsContent).
        Content("network", networkContent).
        ActiveTab(int(si.activeTab)).
        OnChange(func(tabID string) {
            // 标签切换时的回调
            fmt.Printf("Switched to tab: %s\n", tabID)
        }).
        Build()

    return tabs
}
```

### 示例 2: 带键盘导航的 Tabs

```go
type MyComponentState struct {
    activeTab int
    tabs      *navigation.TabsVNode
}

func (s *MyComponentState) Render() ui.VNode {
    s.tabs = navigation.NewTabs().
        AddTab("tab1", "Tab 1").
        AddTab("tab2", "Tab 2").
        AddTab("tab3", "Tab 3").
        ActiveTab(s.activeTab).
        Build()

    return ui.VStack(
        s.tabs,
        ui.Text("Press 1-3 to switch tabs"),
    )
}

func (s *MyComponentState) HandleKeyEvent(key string) bool {
    switch key {
    case "1":
        return s.tabs.SetActiveTabByLabel("Tab 1")
    case "2":
        return s.tabs.SetActiveTabByLabel("Tab 2")
    case "3":
        return s.tabs.SetActiveTabByLabel("Tab 3")
    case "ctrl+tab":
        return s.tabs.NextTab()
    case "ctrl+shift+tab":
        return s.tabs.PreviousTab()
    }
    return false
}
```

### 示例 3: 动态 Tabs

```go
type DynamicTabsState struct {
    tabs      []string
    contents  map[string]ui.VNode
    activeTab string
}

func (s *DynamicTabsState) Render() ui.VNode {
    builder := navigation.NewTabs()

    // 动态添加标签
    for _, tabID := range s.tabs {
        builder.AddTab(tabID, strings.ToUpper(tabID))
        if content, ok := s.contents[tabID]; ok {
            builder.Content(tabID, content)
        }
    }

    // 设置激活标签
    if idx := builder.node.FindTabByID(s.activeTab); idx >= 0 {
        builder.ActiveTab(idx)
    }

    return builder.OnChange(func(id string) {
        s.activeTab = id
    }).Build()
}

func (s *DynamicTabsState) AddTab(id string, content ui.VNode) {
    s.tabs = append(s.tabs, id)
    if s.contents == nil {
        s.contents = make(map[string]ui.VNode)
    }
    s.contents[id] = content

    // 如果是第一个标签，自动激活
    if len(s.tabs) == 1 {
        s.activeTab = id
    }
}

func (s *DynamicTabsState) RemoveTab(id string) {
    // 找到并移除标签
    for i, tabID := range s.tabs {
        if tabID == id {
            s.tabs = append(s.tabs[:i], s.tabs[i+1:]...)
            delete(s.contents, id)

            // 如果移除的是当前标签，切换到第一个
            if s.activeTab == id && len(s.tabs) > 0 {
                s.activeTab = s.tabs[0]
            }
            break
        }
    }
}
```

### 示例 4: 禁用标签

```go
func CreateTabsWithDisabled() ui.VNode {
    return navigation.NewTabs().
        AddTab("enabled", "Enabled Tab").
        AddTab("disabled", "Disabled Tab").
        AddTab("locked", "Locked Tab").
        Content("enabled", ui.Text("This tab is enabled")).
        Content("disabled", ui.Text("This tab is disabled")).
        Content("locked", ui.Text("This tab is locked")).
        Build()
}

// 在运行时禁用标签
func DisableTab(tabs *navigation.TabsVNode) {
    tabs.SetTabEnabled(1, false)  // 禁用第 2 个标签

    // 如果禁用的是当前标签，会自动切换
}
```

### 示例 5: 带图标的标签

```go
func CreateTabsWithIcons() ui.VNode {
    return navigation.NewTabs().
        AddTab("home", "🏠 Home").
        AddTab("settings", "⚙️ Settings").
        AddTab("about", "ℹ️ About").
        Content("home", ui.Text("Home page content")).
        Content("settings", ui.Text("Settings page content")).
        Content("about", ui.Text("About page content")).
        Build()
}
```

---

## 样式自定义

### 激活标签样式

Tabs 组件会自动为激活的标签添加方括号 `[]` 和加粗样式：

```
Tab 1 | [Tab 2] | Tab 3
       ↑ 激活
```

### 自定义样式

Tabs 组件继承自 `ElementVNode`，可以使用 `SetStyle()` 方法：

```go
tabs := navigation.NewTabs().
    AddTab("t1", "Tab 1").
    AddTab("t2", "Tab 2").
    Build()

tabs.SetStyle(style.NewStyle().
    Foreground(style.Cyan).
    Bold(true))
```

---

## 与 Inspector 的集成

Inspector 使用 Tabs 组件来组织不同的功能面板：

```go
// 1. 定义标签枚举
type InspectorTab int

const (
    TabElements InspectorTab = iota
    TabConsole
    TabPerformance
    TabDiagnostics
    TabNetwork
)

// 2. 创建 Inspector 字段
type StandaloneInspector struct {
    activeTab  InspectorTab
    // ...
}

// 3. 构建标签页
func (si *StandaloneInspector) buildActiveTabContent() ui.VNode {
    switch si.activeTab {
    case TabElements:
        return si.buildElementsTab()
    case TabConsole:
        return si.buildConsoleTab()
    case TabPerformance:
        return si.buildPerformanceTab()
    case TabDiagnostics:
        return si.buildDiagnosticsTab()
    case TabNetwork:
        return si.buildNetworkTab()
    default:
        return ui.Text("Tab not implemented")
    }
}

// 4. 创建 tabs 组件
func (si *StandaloneInspector) buildTabs() ui.VNode {
    tabLabels := []string{"Elements", "Console", "Performance", "Diagnostics", "Network"}
    tabIDs := []string{"elements", "console", "performance", "diagnostics", "network"}

    builder := navigation.NewTabs()
    for i, label := range tabLabels {
        builder.AddTab(tabIDs[i], label)
        builder.Content(tabIDs[i], si.buildActiveTabContent())
    }

    return builder.ActiveTab(int(si.activeTab)).
        OnChange(func(tabID string) {
            // 更新 activeTab
            for i, id := range tabIDs {
                if id == tabID {
                    si.activeTab = InspectorTab(i)
                    break
                }
            }
        }).
        Build()
}
```

---

## 键盘快捷键

### 常用快捷键模式

```go
func (app *App) SetupKeyboardShortcuts() {
    // Ctrl+Tab: 下一个标签
    app.OnKeyCombo("ctrl+tab", func() {
        if tabs.NextTab() {
            app.MarkDirty()
        }
    })

    // Ctrl+Shift+Tab: 上一个标签
    app.OnKeyCombo("ctrl+shift+tab", func() {
        if tabs.PreviousTab() {
            app.MarkDirty()
        }
    })

    // 1-9: 数字键切换到对应标签
    for i := 1; i <= 9; i++ {
        tabNum := i
        key := fmt.Sprintf("%d", i)
        app.OnKeyCombo(key, func() {
            if tabs.SetActiveTab(tabNum - 1) {
                app.MarkDirty()
            }
        })
    }
}
```

---

## 与其他组件的对比

### vs 手动实现

| 特性 | 手动实现 | Tabs 组件 |
|------|---------|---------|
| **代码量** | ~50-100 行 | ~10 行 |
| **维护性** | 分散在应用逻辑中 | 集中管理 |
| **可复用性** | ❌ 不可复用 | ✅ 完全可复用 |
| **功能完整性** | 取决于实现 | 完整功能 |

### vs VStack + 条件渲染

```go
// ❌ 手动方式：复杂且易错
var content ui.VNode
switch activeTab {
case 0:
    content = ui.Text("Content 1")
case 1:
    content = ui.Text("Content 2")
}

return ui.VStack(
    ui.Text("[Tab1] | Tab2"),
    content,
)

// ✅ Tabs 组件：简洁清晰
return navigation.NewTabs().
    AddTab("t1", "Tab 1").
    AddTab("t2", "Tab 2").
    Content("t1", ui.Text("Content 1")).
    Content("t2", ui.Text("Content 2")).
    ActiveTab(0).
    Build()
```

---

## 性能考虑

### 内存占用

Tabs 组件本身内存占用很小：
- 每个 TabItem: ~100 bytes
- TabsVNode: ~200 bytes + tabs 数组

### 渲染性能

- **Measure** - O(n)，n 是标签数量
- **Paint** - O(1)，只渲染一行标签文本
- **切换** - O(1)，只需要更新 activeTab 索引

### 优化建议

1. **避免频繁的标签重建**
   ```go
   // ❌ 不好：每次渲染都创建新 tabs
   func Render() ui.VNode {
       return navigation.NewTabs().
           AddTab("t1", "Tab 1").
           Build()
   }

   // ✅ 好：复用 tabs 实例
   type MyState struct {
       tabs *navigation.TabsVNode
   }
   func (s *MyState) Render() ui.VNode {
       if s.tabs == nil {
           s.tabs = navigation.NewTabs().
               AddTab("t1", "Tab 1").
               Build()
       }
       return s.tabs
   }
   ```

2. **预构建标签内容**
   ```go
   // ❌ 不好：每次都创建新内容
   Content("tab1", buildExpensiveContent())

   // ✅ 好：缓存内容
   type MyState struct {
       tab1Content ui.VNode
   }
   // 初始化时构建
   s.tab1Content = buildExpensiveContent()
   Content("tab1", s.tab1Content)
   ```

---

## 限制和注意事项

### 当前限制

1. **不支持嵌套 tabs** - Tabs 内部不能包含另一个 Tabs
2. **不支持拖拽重排** - 标签顺序是固定的
3. **不支持关闭标签** - 标签数量是固定的
4. **不支持滚动** - 如果标签太多，会换行显示

### 使用注意事项

1. **唯一 ID** - 确保每个标签的 ID 是唯一的
   ```go
   // ❌ 不好：重复的 ID
   tabs.AddTab("tab", "First").
         AddTab("tab", "Second")  // ID 重复！

   // ✅ 好：唯一的 ID
   tabs.AddTab("tab1", "First").
         AddTab("tab2", "Second")
   ```

2. **内容匹配** - 确保每个 ID 都有对应的内容
   ```go
   // ❌ 不好：缺少内容
   tabs.AddTab("t1", "Tab 1").
         Build()  // t1 没有内容！

   // ✅ 好：完整的内容
   tabs.AddTab("t1", "Tab 1").
         Content("t1", ui.Text("Content")).
         Build()
   ```

3. **索引范围** - 确保索引在有效范围内
   ```go
   // ❌ 不好：越界索引
   tabs.SetActiveTab(10)  // 只有 3 个标签！

   // ✅ 好：检查索引
   if tabs.GetTabCount() > 10 {
       tabs.SetActiveTab(10)
   }
   ```

---

## 故障排查

### 问题：标签不显示

**原因**：
- 标签 ID 或内容未正确设置
- 标签索引越界

**解决方案**：
```go
tabs := navigation.NewTabs().
    AddTab("tab1", "Label 1").
    Content("tab1", content1).  // ← 确保设置内容
    ActiveTab(0).               // ← 确保索引有效
    Build()
```

### 问题：切换标签无反应

**原因**：
- 标签被禁用
- 标签索引越界
- 回调函数阻止了切换

**解决方案**：
```go
// 检查是否禁用
if !tabs.IsTabEnabled(index) {
    fmt.Println("Tab is disabled")
}

// 检查索引范围
if index < 0 || index >= tabs.GetTabCount() {
    fmt.Println("Index out of range")
}

// 检查回调
tabs.OnChange(func(id string) {
    fmt.Println("Switching to:", id)
    // 不要在这里阻止切换
})
```

### 问题：样式不生效

**原因**：
- 样式设置在错误的节点上
- 样式被覆盖

**解决方案**：
```go
// ✅ 正确：在 tabs 节点上设置样式
tabs.SetStyle(style.NewStyle().
    Foreground(style.Cyan).
    Bold(true))

// ❌ 错误：在包含 tabs 的 VStack 上设置
container := ui.VStack(tabs)
container.SetStyle(...)  // 样式不会应用到 tabs
```

---

## 完整示例

### 文件编辑器示例

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/app"
    ui "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/components/navigation"
)

type FileEditorState struct {
    tabs *navigation.TabsVNode
    fileContent string
    settingsContent ui.VNode
    aboutContent ui.VNode
}

func NewFileEditor() *FileEditorState {
    s := &FileEditorState{}

    // 创建标签页内容
    s.fileContent = "File content here..."
    s.settingsContent = ui.Text("Settings...")
    s.aboutContent = ui.Text("About...")

    // 创建 tabs
    s.tabs = navigation.NewTabs().
        AddTab("editor", "Editor").
        AddTab("settings", "Settings").
        AddTab("about", "About").
        Content("editor", ui.Text(s.fileContent)).
        Content("settings", s.settingsContent).
        Content("about", s.aboutContent).
        ActiveTab(0).
        OnChange(func(tabID string) {
            fmt.Printf("Switched to %s tab\n", tabID)
        }).
        Build()

    return s
}

func (s *FileEditorState) Render() ui.VNode {
    return ui.VStack(
        s.tabs,
        ui.Text("─────────────────"),
        ui.Text("Ctrl+Tab: Switch tabs"),
    )
}

func (s *FileEditorState) HandleKeyEvent(key string) bool {
    switch key {
    case "ctrl+tab":
        return s.tabs.NextTab()
    case "ctrl+shift+tab":
        return s.tabs.PreviousTab()
    }
    return false
}

func main() {
    editor := NewFileEditor()

    err := ui.Run(func() ui.VNode {
        return editor.Render()
    })
    if err != nil {
        panic(err)
    }
}
```

---

## 导出位置

Tabs 组件已从 `app` 包导出：

```go
// app/app.go
var (
    Tabs         = navigation.Tabs
    NewTabs      = navigation.NewTabs
    TabsBuilder  = navigation.TabsBuilder
)
```

使用时可以直接导入：

```go
import "github.com/wwsheng009/mint/app"

// 使用 app.Tabs 类型
var myTabs *app.Tabs

// 使用 app.NewTabs 创建
myTabs = app.NewTabs()
```

---

## 测试

创建测试文件验证功能：

```go
// components/navigation/tabs_test.go

package navigation

import (
    "testing"

    ui "github.com/wwsheng009/mint/ui"
)

func TestTabs_BasicCreation(t *testing.T) {
    tabs := NewTabs().
        AddTab("t1", "Tab 1").
        AddTab("t2", "Tab 2").
        ActiveTab(0).
        Build()

    if tabs == nil {
        t.Fatal("Expected non-nil tabs")
    }

    if tabs.ActiveTab() != 0 {
        t.Errorf("Expected active tab 0, got %d", tabs.ActiveTab())
    }
}

func TestTabs_SwitchTabs(t *testing.T) {
    tabs := NewTabs().
        AddTab("t1", "Tab 1").
        AddTab("t2", "Tab 2").
        ActiveTab(0).
        Build()

    // Switch to second tab
    ok := tabs.SetActiveTab(1)
    if !ok {
        t.Error("Failed to switch to tab 1")
    }

    if tabs.ActiveTab() != 1 {
        t.Errorf("Expected active tab 1, got %d", tabs.ActiveTab())
    }
}

func TestTabs_Navigation(t *testing.T) {
    tabs := NewTabs().
        AddTab("t1", "Tab 1").
        AddTab("t2", "Tab 2").
        AddTab("t3", "Tab 3").
        ActiveTab(0).
        Build()

    // Test NextTab
    tabs.SetActiveTab(0)
    if !tabs.NextTab() {
        t.Error("NextTab should succeed from tab 0")
    }
    if tabs.ActiveTab() != 1 {
        t.Errorf("Expected active tab 1, got %d", tabs.ActiveTab())
    }

    // Test PreviousTab
    tabs.PreviousTab()
    if tabs.ActiveTab() != 0 {
        t.Errorf("Expected active tab 0, got %d", tabs.ActiveTab())
    }

    // Test FirstTab
    tabs.SetActiveTab(2)
    tabs.FirstTab()
    if tabs.ActiveTab() != 0 {
        t.Errorf("Expected active tab 0, got %d", tabs.ActiveTab())
    }

    // Test LastTab
    tabs.LastTab()
    if tabs.ActiveTab() != 2 {
        t.Errorf("Expected active tab 2, got %d", tabs.ActiveTab())
    }
}

func TestTabs_SetActiveTabByLabel(t *testing.T) {
    tabs := NewTabs().
        AddTab("t1", "First").
        AddTab("t2", "Second").
        ActiveTab(0).
        Build()

    ok := tabs.SetActiveTabByLabel("Second")
    if !ok {
        t.Error("Failed to switch by label")
    }

    if tabs.GetActiveTabLabel() != "Second" {
        t.Errorf("Expected label 'Second', got '%s'", tabs.GetActiveTabLabel())
    }
}

func TestTabs_DisabledTabs(t *testing.T) {
    tabs := NewTabs().
        AddTab("t1", "Tab 1").
        AddTab("t2", "Tab 2").
        AddTab("t3", "Tab 3").
        ActiveTab(0).
        Build()

    // Disable second tab
    tabs.SetTabEnabled(1, false)

    // Try to switch to disabled tab - should fail
    ok := tabs.SetActiveTab(1)
    if ok {
        t.Error("Should not be able to switch to disabled tab")
    }

    // NextTab should skip disabled tab
    tabs.SetActiveTab(0)
    if !tabs.NextTab() {
        t.Error("NextTab should succeed")
    }
    if tabs.ActiveTab() != 2 {
        t.Errorf("Expected active tab 2 (skipping disabled), got %d", tabs.ActiveTab())
    }
}
```

运行测试：

```bash
go test ./components/navigation -run TestTabs -v
```

---

## 总结

Tabs 组件提供了一个简洁、强大的方式来组织和切换内容。它的 Builder API 和丰富的运行时方法使得在 TUI 应用中实现标签页变得非常简单。

### 关键要点

- ✅ **简单 API** - Builder 模式，链式调用
- ✅ **灵活布局** - 支持上、下、左、右四个位置
- ✅ **键盘导航** - 完整的导航方法
- ✅ **禁用支持** - 可以禁用特定标签
- ✅ **回调机制** - 标签切换时触发回调
- ✅ **样式继承** - 完全支持自定义样式

### 何时使用

| 场景 | 推荐方案 |
|------|---------|
| 简单的 2-3 个面板 | VStack + 条件渲染 |
| 3-10 个功能面板 | Tabs 组件 |
| 动态标签数量 | Tabs 组件 + 动态管理 |
| 需要键盘导航 | Tabs 组件 + 快捷键 |

---

**版本**: 2.0
**状态**: ✅ 已增强并测试
**日期**: 2025-02-08
**新增功能**:
- TabPosition 支持（上、下、左、右）
- 禁用标签功能
- 丰富的导航方法
- 运行时 API 增强
