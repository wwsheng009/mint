# Menu

> 状态：`MenuBar` / `Popup` / `ContextMenu`、级联 submenu、outside-click / ESC 收口、全局 shortcut 安装 helper 已可用。本文后半仍保留部分设计草案，阅读时以前置“当前实现摘要”为准。

`menu` 目标是提供一套可复用的菜单家族能力，覆盖：

- 顶部菜单栏（MenuBar）
- 下拉菜单（Dropdown Menu）
- 右键菜单（Context Menu）
- 级联子菜单（Cascading Menu）
- 工具栏下拉 / 操作菜单（Action Menu）

设计重点：

- **统一模型**：菜单栏、右键菜单、按钮下拉共用同一套 Item/Theme/Controller
- **键盘优先**：支持方向键、Enter、ESC、Tab、类型前缀搜索、快捷键显示与触发
- **主题完备**：基础态 / hover / focus / active / disabled / checked / separator / shortcut / overlay 阴影全部可配
- **可扩展**：支持 checkbox/radio/submenu/custom item/lazy loading
- **贴合 Mint 现有架构**：基于 `intent`、focus manager、overlay layer、scrollview、tooltip、theme manager 构建
- **Phase 1.5**：已支持 popup anchor/portal 配置、typeahead 导航、submenu path helpers
- **Phase 1.6**：已支持单实例级联 submenu 渲染与 `ActivePath(...)` 受控展开
- **Phase 1.7**：已支持 outside-click / ESC 关闭中间件，以及 app 级 shortcut 注册 helper
- **Phase 1.8**：已支持 anchored popup 根面板的 `Placement(...)` 对齐映射（`bottom/top/right/left-start` 与 `bottom/top-end`），`auto` 会按现有 anchor 方向推导默认 placement
- **接入方式**：可通过 `menu.NewMiddleware()` 接入 action router，通过 `ui.BindMenuGlobalShortcuts(...)` 接入全局快捷键
- **一键安装**：可通过 `menu.Install(...)` 或 `ui.InstallMenu(...)` 一次性安装中间件与全局快捷键
- **默认挂载点**：`ui.Run` / `ui.RunApp` 会自动注入 `overlay/modal/tooltip` PortalRoot，菜单 popup 默认挂到 overlay host

## 当前实现摘要

- 可直接使用 `menu.NewPopup(...)`、`menu.NewContextMenu(...)`、`menu.NewBuilder()` 构建 anchored popup、context menu 与 menu bar 族能力。
- `ActivePath(...)` 可显式控制 submenu 展开路径；单实例级联 submenu 渲染已经接通。
- `Placement(...)`、`AnchorTo(...)`、`PortalRoot(...)`、`PortalOffset(...)` 已可用于 anchored popup / context menu 的定位与挂载。
- anchored root popup 现在会基于真实外框尺寸做 viewport-aware candidate fallback 与 clamp，覆盖 `bottom/top/right/left-start` 和 `bottom/top-end`。
- context menu 仍以 `PortalOffset(...)` 作为目标原点，但当根面板超出 viewport 时会自动 clamp 回可见区。
- submenu 级联面板现在会在右侧放不下时自动翻转到左侧，并对纵向位置做 viewport clamp；多级 submenu 在翻转后会尽量沿同一侧继续展开，极窄 viewport 下也会按最终 clamp 到的一侧继续推导后续方向。这套规则已开始下沉到 `internal/overlayposition` 的共享 cascade helper，组件对外暴露的 hit bounds 也会覆盖整棵 cascade。
- `menu.Install(app, nil)` 可接入 outside-click / ESC 中间件；如果还要注册全局快捷键，则传入 `emit` 和开启 `RegisterShortcuts(true)` 的 builder。
- 当前剩余 gap 主要是把这套级联 candidate / direction 规则进一步推广到更多 overlay 场景，以及继续补更复杂角落组合回归；`menu` 与共享 helper 这边已经补了单层、多层、极窄 viewport、双轴 clamp、bottom-right upward clamp、窄底角同时 left-edge clamp + upward clamp，以及“left-edge clamp 后下一层镜像回右侧”的方向传递矩阵；其中最后一条现在也已经有真实 e2e。

## 安装方式

- 只需要关闭中间件时：

```go
menu.Install(app, nil)
```

- 需要同时安装全局快捷键时：

```go
menu.Install(app, emit, builder)
```

其中 `builder` 需要开启 `RegisterShortcuts(true)`，且对应快捷键的 `Scope` 不能是 `local`。

## 状态语义

- `Open(bool)` 用于显式控制 popup / context menu 是否打开。
- `ActivePath(...)` 用于显式控制级联 submenu 的活动路径。
- `ComponentID(...)` / `SetID(...)` 会影响路由与安装阶段生成的 menu ID，建议对业务菜单保持稳定。
- `auto` placement 会结合当前 anchor 方向推导默认落点，而不是简单固定到底部。

## 测试入口

- 单测：`go test ./ui/components/menu`
- 重点覆盖：`menu_test.go` 中的 placement、submenu path、cascade corner/clamp 回归、typeahead、middleware 和 shortcut 安装；`internal/overlayposition/cascade_test.go` 负责共享 cascade helper 的矩阵校验；`menu_e2e_test.go` 已覆盖单层左翻、多层连续左翻、极窄 corner clamp、bottom-right upward clamp、窄底角 left-edge clamp + upward clamp，以及“left-edge clamp 后下一层镜像回右侧”的真实交互链路
- E2E：`go test ./ui/e2e -run TestE2EMenu`

---

## 1. 设计目标

### 1.1 功能目标

菜单组件需要支持：

- 菜单项点击与键盘触发
- 菜单项快捷键展示（accelerator）
- 全局/局部快捷键绑定
- 复选菜单项（checkbox）
- 单选菜单项（radio group）
- 分隔线、标题、危险操作、禁用态
- 级联子菜单
- 菜单最大高度与滚动
- 自动避让和 overlay 渲染
- 鼠标 hover 打开、键盘导航打开
- 打开/关闭动画预留（首版不强依赖）
- 类型前缀搜索（typeahead）
- 延迟/懒加载子菜单项（例如文件系统、动态命令）

### 1.2 非目标（首版不做）

- 真正的图形 icon raster 渲染
- 富文本菜单项排版
- 命令面板（Command Palette）本体
- 多列 mega menu
- 复杂动画系统

命令面板后续可复用 `menu` 的 item model 与 shortcut model，但不直接混在首版里。

---

## 2. 与现有系统的衔接

菜单设计直接复用仓库已存在的基础能力：

- **Intent 系统**：菜单动作最终发出 `intent.Intent`
- **Focus Manager**：菜单项进入 Tab / 方向键导航体系
- **Overlay Layer**：弹出菜单使用 `LayerOverlay`，必要时允许 `LayerModal`
- **Tooltip Layer**：菜单项说明或帮助提示可复用现有 tooltip 能力
- **ScrollView**：长菜单、文件浏览型菜单使用滚动容器
- **Theme Manager**：未显式传入 Theme 时可接主题系统默认值
- **Hover/Press/Disable Behavior**：菜单项实例复用现有交互状态模型

建议菜单默认绘制到 `LayerOverlay`，保证不会被正文覆盖；右键菜单与级联子菜单都走 overlay，而不是普通流式布局。

---

## 3. 组件家族与分层

建议不是做一个“大一统巨型组件”，而是做一个**共享核心 + 多个入口组件**。

### 3.1 推荐目录结构

```text
ui/components/menu/
  README.md
  types.go
  theme.go
  model.go
  controller.go
  shortcuts.go
  builder.go
  bar_builder.go
  popup_builder.go
  context_builder.go
  menubar_component.go
  popup_component.go
  item_component.go
  overlay_component.go
  middleware.go
```

### 3.2 组件分层

- **MenuModel**：静态配置与数据结构（items / shortcuts / options）
- **MenuController**：运行时状态（open path / active index / typeahead / anchor / scroll）
- **MenuSurface**：单个弹出面板
- **MenuItem**：单项渲染与交互
- **MenuBar**：顶部菜单入口行
- **ContextMenu**：基于坐标或 anchor 弹出的菜单
- **DropdownMenu**：绑定按钮或任意 anchor 的下拉菜单

### 3.3 变体（Variant）

```go
type Variant string

const (
    VariantMenuBar  Variant = "menubar"
    VariantDropdown Variant = "dropdown"
    VariantContext  Variant = "context"
    VariantPopup    Variant = "popup"
)
```

---

## 4. 核心数据模型

### 4.1 MenuItem

```go
type ItemKind string

const (
    ItemAction    ItemKind = "action"
    ItemSeparator ItemKind = "separator"
    ItemCheckbox  ItemKind = "checkbox"
    ItemRadio     ItemKind = "radio"
    ItemSubmenu   ItemKind = "submenu"
    ItemCustom    ItemKind = "custom"
    ItemLabel     ItemKind = "label"
)

type MenuItem struct {
    Key           string
    Label         string
    SecondaryText string
    Description   string
    Icon          string
    Shortcut      Shortcut
    Kind          ItemKind
    Disabled      bool
    Visible       bool
    Checked       bool
    Danger        bool
    Group         string
    Children      []MenuItem
    Intent        intent.Intent
    OnSelect      func() intent.Intent
    CloseOnSelect bool
    KeepOpen      bool
    TestID        string
    ThemeSlot     string
    Metadata      map[string]any
}
```

说明：

- `Shortcut` 既用于展示，也可参与键位触发
- `OnSelect` 适合运行时生成 intent
- `CloseOnSelect` 和 `KeepOpen` 处理 checkbox/radio/menu action 的差异
- `ThemeSlot` 允许某些菜单项走 danger / success / muted 等语义样式

### 4.2 Shortcut

```go
type ShortcutScope string

const (
    ShortcutLocal  ShortcutScope = "local"
    ShortcutGlobal ShortcutScope = "global"
)

type Shortcut struct {
    Key       string
    Combo     string
    Display   string
    Scope     ShortcutScope
    Enabled   bool
    Priority  int
    When      string
    PreventDefault bool
}
```

建议约定：

- `Combo` 用于匹配，例如 `Ctrl+K`、`Alt+F`、`Shift+F10`
- `Display` 用于菜单右侧展示，可允许平台差异格式（Windows/Linux/macOS）
- `When` 预留表达式能力，例如 `editorFocused && !readOnly`

### 4.3 MenuModel

```go
type Model struct {
    ID               string
    Variant          Variant
    Items            []MenuItem
    Open             bool
    MultipleOpen     bool
    ActivePath       []int
    SelectedPath     []int
    AnchorID         string
    AnchorRect       rtui.Rect
    Placement        Placement
    PlacementOffsetX int
    PlacementOffsetY int
    MaxWidth         int
    MaxHeight        int
    MinWidth         int
    LoopNavigation   bool
    Typeahead        bool
    HoverOpenDelay   time.Duration
    HoverCloseDelay  time.Duration
    CloseOnBlur      bool
    CloseOnOutside   bool
    CloseOnEscape    bool
    Scrollable       bool
    ShowShortcuts    bool
    ShowIcons        bool
    ShowDescriptions bool
    ShowCheckMarks   bool
    Layer            rtui.Layer
}
```

---

## 5. 主题设计

菜单必须从一开始就支持完整主题，而不是只给一个 `Style()`。

### 5.1 Theme

```go
type Theme struct {
    BarStyle             style.Style
    BarActiveStyle       style.Style
    SurfaceStyle         style.Style
    SurfaceBorderStyle   style.Style
    SurfaceShadowStyle   style.Style
    ItemStyle            style.Style
    ItemHoverStyle       style.Style
    ItemFocusStyle       style.Style
    ItemActiveStyle      style.Style
    ItemDisabledStyle    style.Style
    ItemDangerStyle      style.Style
    ItemCheckedStyle     style.Style
    SeparatorStyle       style.Style
    ShortcutStyle        style.Style
    DescriptionStyle     style.Style
    IconStyle            style.Style
    CheckmarkStyle       style.Style
    SubmenuArrowStyle    style.Style
    TitleStyle           style.Style
    ScrollbarStyle       style.Style
    ScrollbarThumbStyle  style.Style
    OverlayBackdropStyle style.Style
    TooltipStyle         style.Style
    TooltipBorderStyle   style.Style
}
```

### 5.2 内置主题建议

```go
menu.DefaultTheme()
menu.MutedTheme()
menu.ContrastTheme()
menu.DenseTheme()
menu.MinimalTheme()
```

### 5.3 主题能力要求

- Menubar 与 Popup Surface 主题分离
- 基础态和交互态分离，避免 hover/focus 覆盖掉基础颜色
- 支持 danger item、checked item、disabled item 单独着色
- 支持 shortcut 列单独颜色
- 支持边框、阴影、滚动条样式
- 支持“紧凑模式”和“常规模式”

---

## 6. 交互设计

### 6.1 鼠标交互

- 点击 menubar 项：打开/切换对应菜单
- 悬停 menubar 项：若已有菜单打开，则切换到对应菜单
- 悬停 submenu 项：延迟打开子菜单
- 点击菜单项：触发 action，并按配置关闭或保持打开
- 点击菜单外：关闭所有菜单
- 右键：可切换/重建 context menu
- 滚轮：菜单超高时滚动面板内容

### 6.2 键盘交互

- `Enter` / `Space`：触发当前项
- `Esc`：关闭当前级；在根级时关闭整个菜单
- `↑` `↓`：同层菜单导航
- `←` `→`：menubar 间导航 / submenu 展开收起
- `Home` `End`：跳到首尾
- `Tab`：默认关闭菜单并回到正常焦点流；可配置保留
- 字符输入：typeahead 前缀搜索
- `Alt+<Mnemonic>`：打开 menubar 对应顶级菜单（后续增强）

### 6.3 快捷键策略

快捷键分成两层：

1. **显示型快捷键（Accelerator）**
   - 显示在菜单右侧，例如 `Ctrl+S`
   - 用户看到的是提示，不一定自动注册成全局快捷键

2. **可执行快捷键（Binding）**
   - 由菜单组件或 app 层实际注册
   - 可以是 Local，也可以是 Global

建议首版 API 同时支持：

```go
Builder.RegisterShortcuts(true)
Builder.ShortcutScope(menu.ShortcutLocal)
Builder.ShortcutResolver(func(item MenuItem) []Shortcut { ... })
```

并区分：

- **Global**：通过 app 全局快捷键系统分发
- **Local**：仅在菜单打开、或焦点位于所属区域时生效

### 6.4 复选 / 单选

- `checkbox`：点击切换 `Checked`
- `radio`：同一 `Group` 下互斥
- 首版建议用 intent 驱动状态，不在组件内部偷存业务状态

---

## 7. overlay / popover 定位设计

### 7.1 Placement

```go
type Placement string

const (
    PlacementBottomStart Placement = "bottom-start"
    PlacementBottomEnd   Placement = "bottom-end"
    PlacementTopStart    Placement = "top-start"
    PlacementTopEnd      Placement = "top-end"
    PlacementRightStart  Placement = "right-start"
    PlacementLeftStart   Placement = "left-start"
    PlacementAuto        Placement = "auto"
)
```

### 7.2 定位规则

- anchored popup 的根面板已接入 `Placement(...)`：
  `bottom-start` / `bottom-end` / `top-start` / `top-end` / `right-start` / `left-start`
- `PlacementAuto` 当前会按现有 anchor 方向推导默认 placement：
  `BottomLeft -> bottom-start`、`BottomRight -> bottom-end`、`TopLeft -> top-start`、`TopRight -> top-end`、`Right -> right-start`、`Left -> left-start`
- submenu 默认优先从父面板右侧级联展开；右侧放不下时会翻转到左侧，并对纵向位置做 viewport clamp；多级 submenu 会优先延续父级已选中的展开方向，在极窄 viewport 下则按最终解析出来的位置继续传递方向；对应规则已抽到共享 cascade helper
- context menu 仍按 `PortalOffset(...)` 直接给出目标原点；如果右侧或底部超出 viewport，会自动 clamp 到可见区内
- overlay 一律走多层渲染，不能作为普通流式节点下沉，否则会被正文覆盖

### 7.3 碰撞与边界处理

- 根 popup 已支持 viewport-aware candidate fallback 与 clamp；显式 `bottom/top/right/left-start`、`bottom/top-end` 和 `PlacementAuto` 都会基于根面板真实外框尺寸求最终落点
- context menu 已支持基于真实外框尺寸的 viewport clamp；submenu 也已支持 `left-start` 翻转与级联 hit bounds 扩展
- 高度超出可视区时仍可通过 `MaxHeight` + scrollable 模式控制
- 宽度超出限制时仍按现有 `MaxWidth` / 截断规则处理

---

## 8. 菜单状态机

建议显式引入 Controller，而不是把状态散在 item instance 上。

```go
type Controller struct {
    Open           bool
    OpenPath       []int
    ActivePath     []int
    HoverPath      []int
    Typeahead      string
    TypeaheadUntil time.Time
    AnchorRect     rtui.Rect
    ScrollOffsets  map[string]int
}
```

### 状态机要点

- 根菜单打开 → 选中第一可用项
- 悬停可延迟切换 `HoverPath`
- Enter 进入 submenu 或触发 action
- ESC 逐级关闭并回退到父级
- controller 只维护 UI 状态，不持久化业务数据

---

## 9. Builder API 草案

### 9.1 通用 Builder

```go
menu.NewBuilder().
    ID("main-menu").
    Variant(menu.VariantMenuBar).
    Items(items).
    Theme(menu.DefaultTheme()).
    Layer(ui.LayerOverlay).
    Placement(menu.PlacementAuto).
    MaxWidth(48).
    MaxHeight(20).
    Typeahead(true).
    ShowShortcuts(true).
    CloseOnOutside(true).
    RegisterShortcuts(true).
    Build()
```

### 9.2 更符合使用习惯的快捷入口

```go
menu.NewMenuBar(items)
menu.NewDropdown(anchor, items)
menu.NewContextMenu(items)
menu.NewPopup(items)
```

或在 `ui` 顶层提供：

```go
ui.MenuBar(...)
ui.DropdownMenu(...)
ui.ContextMenu(...)
ui.MenuThemeDefault()
```

### 9.3 受控 / 非受控

建议同时支持：

- **受控**：`Open(bool)` + `OnOpenChange(...)`
- **非受控**：内部 controller 管理打开状态

这样既能用于简单 demo，也能适配复杂业务场景。

---

## 10. 菜单项布局建议

一个标准菜单项建议支持四列：

```text
[check/icon] [label + description] [shortcut] [submenu-arrow]
```

### 列布局规则

- 第 1 列：固定宽度（check / icon）
- 第 2 列：主内容，自适应
- 第 3 列：shortcut，右对齐
- 第 4 列：submenu arrow，固定宽度

分隔线和标题项不参与激活导航。

---

## 11. 与现有组件的复用建议

### 11.1 复用 `scrollview`

- 超长菜单必须复用滚动容器，而不是自己重造滚动逻辑
- 文件浏览 / 大列表菜单可以直接接滚动条主题

### 11.2 复用 `tooltip`

- 菜单项若 `Description` 太长，可 hover 后走 tooltip layer
- tooltip 不替代 shortcut 栏，只补充说明信息

### 11.3 复用 `statusbar` 设计经验

可借鉴：

- 显式 Theme 结构，而不是散落 props
- hover / focus / pressed / disabled 独立样式层
- overlay 帮助与 anchor 定位经验

---

## 12. 事件与 Intent 设计

建议内置几个框架无关的 intent，便于调试和测试：

```go
type OpenMenuIntent struct {
    MenuID string
    Path   []int
}

type CloseMenuIntent struct {
    MenuID string
}

type ToggleMenuIntent struct {
    MenuID string
    Path   []int
}

type ActivateMenuItemIntent struct {
    MenuID  string
    ItemKey string
    Path    []int
}
```

业务动作仍然由具体 `MenuItem.Intent` 决定。

推荐行为：

- 组件内部先发 `ActivateMenuItemIntent`
- 然后再触发业务 intent
- 这样方便埋点、日志和测试回放

---

## 13. 快捷键绑定设计建议

### 13.1 菜单显示 ≠ 自动注册

建议把“显示 shortcut”和“注册 shortcut”分开：

- `ShowShortcuts(true)`：只负责显示右侧提示
- `RegisterShortcuts(true)`：将 menu item 的 binding 接入 app/global handler

### 13.2 平台展示格式

内部 combo 建议统一写：

```text
Ctrl+Shift+P
Alt+Enter
F10
Shift+F10
```

展示层允许做平台转写：

- Windows/Linux：`Ctrl+S`
- macOS：可后续转成 `⌘S`

### 13.3 冲突处理

建议提供冲突回调：

```go
Builder.OnShortcutConflict(func(existing, incoming Shortcut) ShortcutDecision)
```

默认策略：

- 同 scope、同 combo：后注册警告并忽略
- Global 优先级高于 Local

---

## 14. 可访问性与可观测性

建议从第一版就补齐：

- `TestID`
- `MenuID`
- `ItemKey`
- 焦点态可见反馈
- 选中态 / checked 态可见反馈
- debug 日志：打开、关闭、选择、快捷键命中、定位翻转、滚动偏移

---

## 15. 实现阶段建议

### Phase 1：基础可用

- `MenuItem` / `Theme` / `Builder`
- `MenuBar` + `Dropdown` + `ContextMenu`
- overlay 渲染
- 鼠标点击/hover
- 键盘导航
- shortcut 展示
- checkbox/radio
- max height + scroll

### Phase 2：高级体验

- submenu 级联
- typeahead
- local/global shortcut 注册
- auto flip / collision handling
- tooltip description
- 主题预设与 ui 顶层快捷 API

### Phase 3：增强能力

- lazy children
- mnemonic（Alt+F）
- 自定义 item renderer
- 打开/关闭动画钩子
- menu telemetry / action tracing

---

## 16. 推荐首版 API 示例

### 16.1 MenuBar

```go
menu.NewMenuBar([]menu.MenuItem{
    {
        Key:   "file",
        Label: "File",
        Kind:  menu.ItemSubmenu,
        Children: []menu.MenuItem{
            {Key: "new", Label: "New File", Shortcut: menu.Shortcut{Combo: "Ctrl+N", Display: "Ctrl+N"}, Intent: NewFileIntent{}},
            {Key: "open", Label: "Open...", Shortcut: menu.Shortcut{Combo: "Ctrl+O", Display: "Ctrl+O"}, Intent: OpenIntent{}},
            {Kind: menu.ItemSeparator},
            {Key: "readonly", Label: "Read Only", Kind: menu.ItemCheckbox, Checked: true, Intent: ToggleReadonlyIntent{}},
        },
    },
    {
        Key:   "edit",
        Label: "Edit",
        Kind:  menu.ItemSubmenu,
        Children: []menu.MenuItem{
            {Key: "undo", Label: "Undo", Shortcut: menu.Shortcut{Combo: "Ctrl+Z", Display: "Ctrl+Z"}, Intent: UndoIntent{}},
        },
    },
}).
    Theme(menu.DefaultTheme()).
    RegisterShortcuts(true).
    Build()
```

### 16.2 ContextMenu

```go
menu.NewContextMenu(items).
    Theme(menu.ContrastTheme()).
    Placement(menu.PlacementAuto).
    MaxHeight(14).
    CloseOnOutside(true).
    Build()
```

---

## 17. 设计结论

推荐按下面原则实现：

- **一个共享核心，多个产品化入口**：MenuBar / Dropdown / ContextMenu 只是不同入口，不是不同体系
- **先把 Theme 和 Shortcut 做对**：这是菜单“专业感”的关键
- **所有弹出面板都走 overlay**：避免再遇到被正文覆盖的问题
- **状态集中在 Controller**：不要把 open path / active path 分散到多个 item instance
- **显示快捷键与注册快捷键分离**：避免 API 模糊和副作用过强
- **首版先做好 keyboard + mouse + scroll + submenu**：这四项决定菜单是否真正可用

---

## 18. 下一步建议

建议下一步按顺序落地：

1. 先建 `types.go` / `theme.go` / `builder.go`
2. 先实现 `MenuSurface + MenuItem` 两个底层件
3. 再做 `MenuBar` 和 `ContextMenu`
4. 最后接入 shortcut 注册、submenu、typeahead

如果继续实现，推荐先从 **MenuBar + Dropdown Surface** 开始，因为它最容易验证主题、快捷键展示和 overlay 定位是否正确。

