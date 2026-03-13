# Mint UI Components Roadmap

> 本文档记录 Mint UI 组件库的现状、与 Ant Design 的对比分析，以及后续开发计划。
>
> 更新日期：2026-03-14

---

## 当前组件清单（30个）

| 组件 | 目录 | 完整度 | 说明 |
|------|------|--------|------|
| Button | `button/` | ★★★★☆ | Primary/Secondary/Danger/Ghost变体，disabled，icon |
| Input | `input/` | ★★★★☆ | text/password/number/email，placeholder，maxLen，readOnly |
| Textarea | `textarea/` | ★★★☆☆ | 多行文本输入 |
| Select | `select/` | ★★★★☆ | 单选/多选，overlay popup，placeholder，disabled |
| Checkbox | `checkbox/` | ★★★☆☆ | 基本勾选，缺 indeterminate 和 CheckboxGroup |
| OptionGroup | `optiongroup/` | ★★★☆☆ | 选项组，支持 radio/checkbox 模式 |
| Form | `form/` | ★★☆☆☆ | 表单容器，submit/reset intent，缺 FormItem 封装 |
| Table | `table/` | ★★★★☆ | 排序、分页、过滤、多选、搜索、滚动 |
| List | `list/` | ★★★☆☆ | 基本列表，选择模式 |
| VirtualList | `virtuallist/` | ★★★☆☆ | 虚拟滚动列表 |
| TreeView | `treeview/` | ★★★★☆ | 懒加载、展开折叠、搜索、多选、受控模式 |
| Modal | `modal/` | ★★★★☆ | overlay定位，portal架构 |
| Tabs | `tabs/` | ★★★★☆ | intent，controlled模式 |
| Menu | `menu/` | ★★★★★ | menubar/dropdown/context/popup，submenu，shortcut |
| Alert | `alert/` | ★★★☆☆ | 内联提示，info/success/warning/error，可关闭 |
| Spin | `spin/` | ★★★☆☆ | 加载指示器，small/default/large，tip，TickFrame动画 |
| Toast | `toast/` | ★★★☆☆ | 已从 tooltip 独立，info/success/warning/error，自动消失 |
| Tooltip | `tooltip/` | ★★★☆☆ | placement 方位有限 |
| Progress | `progress/` | ★★☆☆☆ | 线形进度条，value/max，showPercent |
| Divider | `divider/` | ★★★☆☆ | 水平/垂直分隔线 |
| StatusBar | `statusbar/` | ★★★☆☆ | help、section 子组件 |
| Panel | `panel/` | ★★★☆☆ | 容器面板，有 enhanced builder |
| ScrollView | `scrollview/` | ★★★☆☆ | 可滚动容器 |
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
| **Notification** | 右上角通知弹窗，带标题和内容 | `Notification` |
| **Message** | 全局顶部消息提示（Toast 已部分实现，需独立为正式组件） | `message` |
| **Result** | 结果状态页，success/error/404/403 | `Result` |
| **Skeleton** | 骨架屏加载占位 | `Skeleton` |

### 表单类（高优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| **Radio** | 独立单选按钮（当前只有 OptionGroup，缺独立 Radio） | `Radio` |
| **Switch** | 开关，toggle，UX 与 Checkbox 不同 | `Switch` |
| **Slider** | 数值范围滑块，支持单值和范围 | `Slider` |
| **Rate** | 星级评分 | `Rate` |
| **DatePicker** | 日期选择器 | `DatePicker` |
| **TimePicker** | 时间选择器 | `TimePicker` |
| **Cascader** | 级联选择，层级数据选择 | `Cascader` |
| **Transfer** | 穿梭框，双列选择 | `Transfer` |

### 导航类（中优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| **Breadcrumb** | 面包屑导航，层级路径展示 | `Breadcrumb` |
| **Pagination** | 独立分页组件（Table 内置了分页，但缺独立组件） | `Pagination` |
| **Steps** | 步骤条，引导流程 | `Steps` |
| **Anchor** | 锚点导航 | `Anchor` |

### 数据展示类（中优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| **Badge** | 徽标数，消息计数角标 | `Badge` |
| **Tag** | 标签，支持可关闭、颜色 | `Tag` |
| **Collapse** | 折叠面板，手风琴模式 | `Collapse` |
| **Descriptions** | 描述列表，键值对展示 | `Descriptions` |
| **Empty** | 空状态占位 | `Empty` |
| **Statistic** | 统计数字展示 | `Statistic` |
| **Timeline** | 时间轴 | `Timeline` |

### 布局类（低优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| **Space** | 间距组件，统一子元素间距 | `Space` |
| **Layout** | 整体布局（Header/Sider/Content/Footer） | `Layout` |
| **Row/Col** | Flex 栅格行列布局 | `Row`, `Col` |

### 其他（低优先级）

| 组件 | 说明 | AntD 对应 |
|------|------|-----------|
| **Popover** | 气泡卡片，比 Tooltip 更复杂，有标题和内容区 | `Popover` |
| **Popconfirm** | 气泡确认框 | `Popconfirm` |
| **Drawer** | 侧边抽屉，从边缘滑出的覆盖层 | `Drawer` |

---

## 已有组件的功能增强计划

| 组件 | 缺失功能 | 优先级 |
|------|---------|--------|
| **Form** | FormItem 封装（label + 控件 + 错误提示三合一）；layout 配置（horizontal/vertical/inline） | 高 |
| **Input** | prefix/suffix 插槽；addonBefore/addonAfter；Search 变体；InputNumber 完整实现 | 高 |
| **Checkbox** | indeterminate 半选状态；CheckboxGroup 组件 | 高 |
| **Progress** | 圆形（Circle）和仪表盘（Dashboard）样式；status（success/exception/active） | 中 |
| **Select** | OptGroup 分组；filterOption 搜索过滤；tags 模式（自定义输入+选择） | 中 |
| **Table** | expandable 行展开；固定列（sticky column）；列宽百分比/自适应；树形数据展示 | 中 |
| **Modal** | confirm/info/success/error/warning 快捷静态方法 | 中 |
| **Tooltip** | 12方位 placement 精细控制；delay 配置 | 中 |
| **Tabs** | card 样式；可关闭标签页（closable）；标签页拖拽排序 | 低 |
| **TreeView** | drag-and-drop 拖拽排序；异步搜索高亮分页 | 低 |
| **List** | List.Item 子组件；虚拟滚动集成 | 低 |

---

## 开发路线图

### Phase 1 — 基础反馈体系（近期）

补全应用开发中最高频的反馈类组件，让用户能清晰感知操作结果和系统状态。

- [x] `Alert` — 内联状态提示（2026-03）
- [ ] `Spin` — 加载状态指示器
- [x] `Message`/`Toast` — 已独立为 `ui/components/toast/` 包（2026-03）
- [ ] `Notification` — 通知弹窗
- [ ] `Empty` — 空状态
- [ ] `Form` 增强 — 完成 FormItem 封装，支持 label + 验证错误提示联动

### Phase 2 — 表单体系完善（中期）

完善表单控件，使表单体系与 Form 验证体系闭环。

- [ ] `Radio` — 独立单选按钮 + RadioGroup
- [ ] `Switch` — 开关控件
- [ ] `Checkbox` 增强 — indeterminate 状态，CheckboxGroup
- [ ] `Input` 增强 — prefix/suffix，Search 变体，addonBefore/addonAfter
- [ ] `Select` 增强 — OptGroup，filterOption，tags 模式
- [ ] `Slider` — 数值滑块
- [ ] `Rate` — 星级评分

### Phase 3 — 导航与数据展示（中期）

- [ ] `Breadcrumb` — 面包屑
- [ ] `Pagination` — 独立分页
- [ ] `Steps` — 步骤条
- [ ] `Badge` — 徽标数
- [ ] `Tag` — 标签
- [ ] `Collapse` — 折叠面板
- [ ] `Descriptions` — 描述列表
- [ ] `Statistic` — 统计数字
- [ ] `Table` 增强 — expandable 行，固定列，树形数据

### Phase 4 — 高级交互组件（远期）

- [ ] `Drawer` — 侧边抽屉
- [ ] `Popover` — 气泡卡片
- [ ] `Popconfirm` — 气泡确认框
- [ ] `Progress` 增强 — Circle/Dashboard 样式
- [ ] `Timeline` — 时间轴
- [ ] `Result` — 结果状态页
- [ ] `Skeleton` — 骨架屏
- [ ] `DatePicker` — 日期选择（TUI 适配）
- [ ] `TimePicker` — 时间选择
- [ ] `Cascader` — 级联选择
- [ ] `Space` — 间距布局
- [ ] `Layout` — 整体布局框架

---

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
