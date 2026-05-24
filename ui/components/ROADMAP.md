# Mint UI Components Roadmap

> 本文档记录 Mint UI 组件库的现状、与 Ant Design 的对比分析，以及后续开发计划。
>
> 更新日期：2026-03-28
>
> 实施状态核对：截至 2026-03-28，本文中的 Phase 1-4 路线图条目已全部完成；`go test ./ui/components/...` 与 `go test ./ui/e2e/...` 当前通过，本轮 overlay 收口已把 `popover`、`popconfirm` 的 viewport-aware placement / fallback / clamp 进一步下沉到 `internal/overlayposition`，`statusbar` overlay help 也已补上真实 viewport 注入与 dedicated corner e2e。
>
> 详细增强任务拆分见：[OPTIMIZATION_BACKLOG.md](./OPTIMIZATION_BACKLOG.md)

---

## 当前组件清单（71 个目录 + 1 个内置 Scrollbar）

> 排除 `docs/` 与 `internal/` 后，当前共有 71 个组件/支撑目录；多数交互组件遵循 `builder.go + vnode.go + instance.go` 或等价 Fiber-first 结构，`toast`、`statusbar`、`control`、`validation` 等属于 manager / 组合式 / 基础设施特例。

| 组件 | 目录 | 完整度 | 说明 |
|------|------|--------|------|
| Button | `button/` | ★★★★☆ | Primary/Secondary/Danger/Ghost变体，disabled，icon |
| Breadcrumb | `breadcrumb/` | ★★★☆☆ | 面包屑导航，支持自定义分隔符、当前项高亮、窄宽度折叠 |
| Pagination | `pagination/` | ★★★☆☆ | 独立分页组件，支持页码跳转、ellipsis、field 绑定 |
| Steps | `steps/` | ★★★☆☆ | 基础步骤条，支持 horizontal/vertical、current/status、description、progressDot、percent、键鼠切换、`StepChangeIntent`/field 绑定，README 与 e2e 已补齐 |
| Badge | `badge/` | ★★★☆☆ | 徽标数/点状态，支持 count/text/dot、overflow、showZero、status、label |
| Collapse | `collapse/` | ★★★☆☆ | 折叠面板，支持多面板展开、accordion、受控/非受控 activeKeys、panel disabled、`CollapseToggleIntent`/`CollapseChangeIntent`、field 绑定 |
| Descriptions | `descriptions/` | ★★★☆☆ | 描述列表，支持多列布局、span、horizontal/vertical、title/extra、bordered |
| Statistic | `statistic/` | ★★★☆☆ | 统计数字展示，支持 title/value、prefix/suffix、precision、千分位/小数分隔符、trend、bordered、extra |
| Input | `input/` | ★★★★☆ | text/password/number/email，placeholder，maxLen，readOnly，prefix/suffix，addonBefore/addonAfter，Search变体，number 模式的负号/小数控制、`min/max/step` |
| Textarea | `textarea/` | ★★★☆☆ | 多行文本输入 |
| Select | `select/` | ★★★★★ | 单选/多选/tags，overlay popup，OptGroup，filterOption，placeholder，disabled |
| Checkbox | `checkbox/` | ★★★★☆ | 基本勾选、indeterminate、CheckboxGroup |
| Radio | `radio/` | ★★★☆☆ | 独立单选按钮，含 RadioGroup 包装层 |
| Switch | `switch/` | ★★★☆☆ | 开关切换，支持 label、自定义 on/off 文案、Field/Form 绑定 |
| Slider | `slider/` | ★★★★☆ | 数值滑块，支持键盘调节、Form 绑定、受控模式 |
| Rate | `rate/` | ★★★☆☆ | 星级评分，支持键盘调节、可清空、Form 绑定 |
| DatePicker | `datepicker/` | ★★★☆☆ | 日期选择，支持 `YYYY-MM-DD` 输入、弹出月视图、键盘/鼠标导航、Field/Form 绑定 |
| TimePicker | `timepicker/` | ★★★☆☆ | 时间选择，支持 `HH:mm` 输入、弹出时间面板、键盘/鼠标导航、Field/Form 绑定 |
| Cascader | `cascader/` | ★★★☆☆ | 级联选择，支持层级数据、列式展开、键盘/鼠标导航、叶子提交、`changeOnSelect`、Field/Form 绑定 |
| Transfer | `transfer/` | ★★★☆☆ | 穿梭框，支持双列表迁移、源/目标搜索过滤、`targetKeys` 受控/非受控、禁用项过滤、Field/Form 绑定 |
| OptionGroup | `optiongroup/` | ★★★☆☆ | 选项组，支持 radio/checkbox 模式 |
| Form | `form/` | ★★★★☆ | 表单容器，submit/reset intent，已补 FormItem、layout、validator 联动，以及 touched/dirty/submitted/submitCount 等字段状态 |
| Table | `table/` | ★★★★☆ | 排序、分页、过滤、多选、搜索、滚动，已补 expandable 行、固定列与树形数据 |
| List | `list/` | ★★★★★ | 基本列表，选择模式，List.Item/item 模型，VirtualList bridge/state sync |
| VirtualList | `virtuallist/` | ★★★☆☆ | 虚拟滚动列表 |
| TreeView | `treeview/` | ★★★★☆ | 懒加载、展开折叠、搜索、多选、受控模式、同父级 subtree 拖拽排序、异步搜索结果分页 |
| Modal | `modal/` | ★★★★☆ | overlay定位，portal架构，静态 helper 支持 prefix/footer layout/按钮文案与变体模板 |
| Drawer | `drawer/` | ★★★★☆ | 侧边抽屉，支持 placement、受控显隐、overlay 与 ESC/遮罩关闭 |
| Tabs | `tabs/` | ★★★★★ | intent，controlled 模式，card/closable/drag reorder |
| Menu | `menu/` | ★★★★★ | menubar/dropdown/context/popup，submenu，shortcut；anchored popup 根面板已支持显式 placement 与 anchor-aware auto，多级 submenu 也已覆盖 left-flip、viewport clamp 与共享 cascade helper |
| Anchor | `anchor/` | ★★★☆☆ | 锚点导航，支持层级链接、当前项高亮、键鼠切换、受控/非受控 `activeKey`、Field/Form 绑定 |
| Alert | `alert/` | ★★★☆☆ | 内联提示，info/success/warning/error，可关闭 |
| Spin | `spin/` | ★★★☆☆ | 加载指示器，small/default/large，tip，TickFrame动画 |
| Notification | `notification/` | ★★★☆☆ | 通知弹窗，info/success/warning/error，可关闭，placement，duration |
| Empty | `empty/` | ★★★☆☆ | 空状态占位，自定义描述和图片 |
| Skeleton | `skeleton/` | ★★★☆☆ | 骨架屏占位，支持 avatar/title/paragraph、loading gate、内容回落 |
| Result | `result/` | ★★★☆☆ | 结果状态页，支持 info/success/warning/error/403/404/500、title/subtitle、extra、bordered |
| Toast | `toast/` | ★★★☆☆ | 独立 manager + runtime，info/success/warning/error，自动消失 |
| Tooltip | `tooltip/` | ★★★★☆ | 已支持 12 方位 placement、auto/显式回退、viewport clamp、delay、layer，且定位 helper 已开始被其他 overlay 复用 |
| Tag | `tag/` | ★★★☆☆ | 标签，颜色变体，可关闭，可选图标前缀 |
| Progress | `progress/` | ★★★☆☆ | line/block/circle/dashboard 进度条，status（normal/success/exception/active），showPercent，`active` 与 indeterminate 动画 |
| Divider | `divider/` | ★★★☆☆ | 水平/垂直分隔线 |
| StatusBar | `statusbar/` | ★★★☆☆ | 组合式 builder，含三槽 section、固定宽度、交互 section、help/tooltip；overlay tooltip 支持 auto/top/bottom、fallback 与 clamp，并已复用共享候选位置 helper，overlay help 运行时已接入真实 viewport 注入 |
| Toolbar | `toolbar/` | ★★★☆☆ | 运维/数据页工具栏，支持 left/center/right 槽、title、button/badge/custom/statusbar 模式，以及受控 dropdown menu |
| Timer | `timer/` | ★★★☆☆ | elapsed/countdown/retry/auto-refresh 计时展示，支持 live ticking、静态渲染、固定宽度和 ASCII progress |
| Charts | `charts/` | ★★★★☆ | 图表组件族，包含 sparkline、bar、line、bullet、heatmap、scatter、candlestick；sparkline 支持 ASCII 模式 |
| Panel | `panel/` | ★★★☆☆ | 容器面板，有 enhanced builder |
| Popover | `popover/` | ★★★☆☆ | 气泡卡片，支持 title/body、click/hover/manual 触发、auto + top/bottom 6 方位 placement、viewport-aware fallback/clamp、local open intents，以及 Install 后的 ESC / outside-click 收口 |
| Popconfirm | `popconfirm/` | ★★★☆☆ | 气泡确认框，支持 title/description、OK/Cancel 操作、click/hover/manual 触发、top/bottom 系列 placement、viewport-aware fallback/clamp、confirm/cancel intents、按钮 variant/footer layout，以及 Install 后的 ESC / outside-click 收口 |
| Timeline | `timeline/` | ★★★☆☆ | 时间轴，支持 label/content/description、status、自定义 dot、pending、reverse |
| ScrollView | `scrollview/` | ★★★☆☆ | 可滚动容器 |
| Space | `space/` | ★★★☆☆ | 间距布局，支持 horizontal/vertical、size、wrap、split、cross-axis align |
| Layout | `layout/` | ★★★☆☆ | 整体布局，支持 Header/Sider/Content/Footer 组合、左右双侧边栏、body/content `flex` 填充 |
| Row/Col | `rowcol/` | ★★★☆☆ | 24 栅格 Flex 行列布局，支持 span、offset、gutter、wrap、justify、align |
| Grid | `grid/` | ★★★★☆ | 单元格边框，完整文档 |
| Text | `text/` | ★★★☆☆ | 文本显示 |
| Absolute | `absolute/` | ★★☆☆☆ | 绝对定位容器 |
| Wrap | `wrap/` | ★★☆☆☆ | 换行容器 |
| Cursor | `cursor/` | ★★☆☆☆ | 光标配置（内部工具） |
| Control | `control/` | ★★☆☆☆ | stay_pressed 等控制原语 |
| Validation | `validation/` | ★★★☆☆ | 表单验证支持 |
| Scrollbar | (内置) | ★★★☆☆ | 各组件内置 scrollbar 样式 |

---

## 与 Ant Design 对比 — 缺失组件

### 反馈类（高优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| ~~**Spin**~~ | ~~异步加载指示器，支持全屏遮罩和局部 loading~~ | `Spin` ✅ 已实现（2026-03）|
| ~~**Notification**~~ | ~~右上角通知弹窗，带标题和内容~~ | `Notification` ✅ 已实现（2026-03）|
| ~~**Message**~~ | ~~全局顶部消息提示（已由 `ui/components/toast/` 提供）~~ | `message` ✅ 已实现（2026-03）|
| ~~**Result**~~ | ~~结果状态页，success/error/404/403~~ | `Result` ✅ 已实现（2026-03-18） |
| ~~**Skeleton**~~ | ~~骨架屏加载占位~~ | `Skeleton` ✅ 已实现（2026-03-18） |

### 表单类（高优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| ~~**Radio**~~ | ~~独立单选按钮，含 RadioGroup~~ | `Radio` ✅ 已实现（2026-03）|
| ~~**Switch**~~ | ~~开关，toggle，UX 与 Checkbox 不同~~ | `Switch` ✅ 已实现（2026-03） |
| ~~**Slider**~~ | ~~数值范围滑块，支持单值和范围~~ | `Slider` ✅ 已实现（2026-03）|
| ~~**Rate**~~ | ~~星级评分~~ | `Rate` ✅ 已实现（2026-03）|
| ~~**DatePicker**~~ | ~~日期选择器~~ | `DatePicker` ✅ 已实现（2026-03-18） |
| ~~**TimePicker**~~ | ~~时间选择器~~ | `TimePicker` ✅ 已实现（2026-03-24） |
| ~~**Cascader**~~ | ~~级联选择，层级数据选择~~ | `Cascader` ✅ 已实现（2026-03-24） |
| ~~**Transfer**~~ | ~~穿梭框，双列选择~~ | `Transfer` ✅ 已实现（2026-03-25） |

### 导航类（中优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| ~~**Breadcrumb**~~ | ~~面包屑导航，层级路径展示~~ | `Breadcrumb` ✅ 已实现（2026-03） |
| ~~**Pagination**~~ | ~~独立分页组件（Table 内置了分页，但缺独立组件）~~ | `Pagination` ✅ 已实现（2026-03） |
| ~~**Steps**~~ | ~~步骤条，引导流程~~ | `Steps` ✅ 已实现（2026-03-17） |
| ~~**Anchor**~~ | ~~锚点导航~~ | `Anchor` ✅ 已实现（2026-03-25） |

### 数据展示类（中优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| ~~**Badge**~~ | ~~徽标数，消息计数角标~~ | `Badge` ✅ 已实现（2026-03-17） |
| ~~**Tag**~~ | ~~标签，支持可关闭、颜色~~ | `Tag` ✅ 已实现（2026-03）|
| ~~**Collapse**~~ | ~~折叠面板，手风琴模式~~ | `Collapse` ✅ 已实现（2026-03-18） |
| ~~**Descriptions**~~ | ~~描述列表，键值对展示~~ | `Descriptions` ✅ 已实现（2026-03-18） |
| ~~**Empty**~~ | ~~空状态占位~~ | `Empty` ✅ 已实现（2026-03）|
| ~~**Statistic**~~ | ~~统计数字展示~~ | `Statistic` ✅ 已实现（2026-03-18） |
| ~~**Timeline**~~ | ~~时间轴~~ | `Timeline` ✅ 已实现（2026-03-18） |

### 布局类（低优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| ~~**Space**~~ | ~~间距组件，统一子元素间距~~ | `Space` ✅ 已实现（2026-03-25） |
| ~~**Layout**~~ | ~~整体布局（Header/Sider/Content/Footer）~~ | `Layout` ✅ 已实现（2026-03-25） |
| ~~**Row/Col**~~ | ~~Flex 栅格行列布局~~ | `Row`, `Col` ✅ 已实现（2026-03-25） |

### 其他（低优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| ~~**Popover**~~ | ~~气泡卡片，比 Tooltip 更复杂，有标题和内容区~~ | `Popover` ✅ 已实现（2026-03-18） |
| ~~**Popconfirm**~~ | ~~气泡确认框~~ | `Popconfirm` ✅ 已实现（2026-03-18） |
| ~~**Drawer**~~ | ~~侧边抽屉，从边缘滑出的覆盖层~~ | `Drawer` ✅ 已实现（2026-03） |

---

## 已有组件的功能增强计划

| 组件 | 缺失功能 | 优先级 |
|------|---------|--------|
| ~~**Form**~~ | ~~字段级 touched/dirty/submitted、表单级 `HasSubmitted`/`GetSubmitCount`、`GetTouchedFields`/`GetDirtyFields`/`GetSubmittedFields`、`GetFormContext` 实例树解析与 compat registry 收口~~ ✅ 已完成（2026-03-25） | ~~中~~ |
| ~~**Input**~~ | ~~InputNumber 完整实现~~ ✅ 已完成（2026-03-17） | ~~高~~ |
| ~~**Checkbox**~~ | ~~indeterminate 半选状态；CheckboxGroup 组件~~ | ~~高~~ |
| ~~**Progress**~~ | ~~圆形（Circle）和仪表盘（Dashboard）样式；status（success/exception/active）~~ ✅ 已完成（2026-03-17） | ~~中~~ |
| ~~**Select**~~ | ~~OptGroup 分组；filterOption 搜索过滤；tags 模式（自定义输入+选择）~~ | ~~中~~ ✅ 已完成（2026-03） |
| ~~**Table**~~ | ~~expandable 行展开；固定列（sticky column）；列宽百分比/自适应；树形数据展示~~ ✅ 已完成（2026-03-17） | ~~中~~ |
| ~~**Modal**~~ | ~~confirm/info/success/error/warning 快捷静态方法~~ ✅ 已完成（2026-03-17） | ~~中~~ |
| ~~**Tooltip**~~ | ~~12方位 placement 精细控制；delay 配置~~ ✅ 已完成（2026-03-17） | ~~中~~ |
| ~~**Tabs**~~ | ~~card、closable、拖拽排序已完成~~ ✅（2026-03-17） | ~~低~~ |
| ~~**TreeView**~~ | ~~同父级 subtree 拖拽排序 + 异步搜索高亮分页~~ ✅ 已完成（2026-03-17，且已补 `ui/e2e` 的 search / lazy load / selection / drag reorder 回归） | ~~低~~ |
| ~~**List**~~ | ~~List.Item / item 模型、VirtualList bridge 与选择/搜索/高亮同步已完成~~ ✅（2026-03-17） | ~~低~~ |

---

## 开发路线图

> 当前状态：本节已从“待实施计划”收口为“已完成阶段记录”。如果后续继续做组件增强、README/demo 收口或测试覆盖补强，不再继续向下追加新的 Phase，而是以 [OPTIMIZATION_BACKLOG.md](./OPTIMIZATION_BACKLOG.md) 与各组件 README 为准。

### Phase 1 — 基础反馈体系（近期）

补全应用开发中最高频的反馈类组件，让用户能清晰感知操作结果和系统状态。

- [x] `Alert` — 内联状态提示（2026-03）
- [x] `Spin` — 加载状态指示器（2026-03）
- [x] `Message`/`Toast` — 已独立为 `ui/components/toast/` 包（2026-03）
- [x] `Notification` — 通知弹窗（2026-03）
- [x] `Empty` — 空状态（2026-03）
- [x] `Form` 增强 — 完成 FormItem 封装，支持 label + 验证错误提示联动（2026-03）

### Phase 2 — 表单体系完善（中期）

完善表单控件，使表单体系与 Form 验证体系闭环。

- [x] `Radio` — 独立单选按钮 + RadioGroup（2026-03）
- [x] `Switch` — 开关控件（2026-03）
- [x] `Checkbox` 增强 — indeterminate 状态，CheckboxGroup（2026-03）
- [x] `Input` 增强 — prefix/suffix，Search 变体，addonBefore/addonAfter（2026-03）
- [x] `Select` 增强 — OptGroup，filterOption，tags 模式（2026-03）
- [x] `Slider` — 数值滑块（2026-03）
- [x] `Rate` — 星级评分（2026-03）
- [x] `Transfer` — 穿梭框（2026-03-25）

### Phase 3 — 导航与数据展示（中期）

- [x] `Breadcrumb` — 面包屑（2026-03）
- [x] `Anchor` — 锚点导航（2026-03-25）
- [x] `Pagination` — 独立分页（2026-03）
- [x] `Steps` — 步骤条（2026-03-17）
- [x] `Badge` — 徽标数（2026-03-17）
- [x] `Tag` — 标签（2026-03）
- [x] `Collapse` — 折叠面板（2026-03-18）
- [x] `Descriptions` — 描述列表（2026-03-18）
- [x] `Statistic` — 统计数字（2026-03-18）
- [x] `Table` 增强 — expandable 行，固定列，树形数据（2026-03-17）

### Phase 4 — 高级交互组件（远期）

- [x] `Drawer` — 侧边抽屉（2026-03）
- [x] `Popover` — 气泡卡片（2026-03-18）
- [x] `Popconfirm` — 气泡确认框（2026-03-18）
- [x] `Progress` 增强 — Circle/Dashboard 样式（2026-03-17）
- [x] `Timeline` — 时间轴（2026-03-18）
- [x] `Result` — 结果状态页（2026-03-18）
- [x] `Skeleton` — 骨架屏（2026-03-18）
- [x] `DatePicker` — 日期选择（TUI 适配，2026-03-18）
- [x] `TimePicker` — 时间选择（TUI 适配，2026-03-24）
- [x] `Cascader` — 级联选择（2026-03-24）
- [x] `Space` — 间距布局（2026-03-25）
- [x] `Layout` — 整体布局框架（2026-03-25）
- [x] `Row/Col` — Flex 栅格行列布局（2026-03-25）

---

## 当前余量

当前 roadmap 级工作已经完成；剩余事项主要是“已有组件增强”和“文档/测试收口”，不再是新增组件缺口。

- 组件增强以 [OPTIMIZATION_BACKLOG.md](./OPTIMIZATION_BACKLOG.md) 为主，重点是较薄弱组件的能力补齐与语义收口
- 文档整理重点仍是复杂组件 README、迁移指南和历史设计文档的现状同步
- 测试层面已从“组件缺失”转向“dedicated e2e 覆盖密度”和 `internal/` 支撑模块空白补齐

## 组件实现规范

所有新组件应遵循现有架构约定：

1. **文件结构**
   ```
   component/
   ├── vnode.go      # 纯声明式描述，无状态无闭包
   ├── instance.go   # 运行时状态和渲染逻辑
   ├── builder.go    # Fluent API 构建器
   ├── intent.go     # 组件 Intent 类型定义（如需）
   └── *_test.go     # 单元测试
   ```

2. **设计原则**
   - VNode 只描述，不持有状态
   - Instance 持有运行时状态，处理 Intent
   - Builder 提供链式 API
   - 受控/非受控模式均需支持
   - 所有交互通过 Intent 系统传递，不直接回调

3. **TUI 适配原则**
   - 日期/时间类组件以文本输入为主，弹出选择为辅
   - 图片、媒体相关组件以 ASCII 字符渲染或占位符替代
   - 动画效果以字符动画（spinner frames）实现
   - 布局以终端字符网格为基础，不依赖像素

---

## 参考资料

- [Ant Design 组件总览](https://ant.design/components/overview)
- [Bubble Tea 组件生态](https://github.com/charmbracelet/bubbletea)
- [Lip Gloss 样式系统](https://github.com/charmbracelet/lipgloss)
- 项目内文档：`ui/components/docs/FOCUS_MANAGEMENT_ANALYSIS.md`
- 项目内文档：`ui/components/COMPONENT_MIGRATION_GUIDE.md`
