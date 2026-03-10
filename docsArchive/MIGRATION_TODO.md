# 组件迁移 TODO List

## 状态说明

- [ ] 待开始
- [~] 进行中
- [x] 已完成
- [!] 阻塞

---

## 已完成组件

### 布局组件

- [x] **Absolute** (`ui/components/absolute/`)
  - [x] vnode.go - VNode 定义
  - [x] instance.go - Fiber 实例
  - [x] builder.go - Builder API
  - [x] absolute_test.go - 单元测试
  - [x] 布局引擎集成 (AbsoluteStyleProvider)
  - [x] 示例验证 (fiber_firsts/absolute_demo)

- [x] **Grid** (`ui/components/grid/`)
  - [x] vnode.go - VNode 定义
  - [x] instance.go - Fiber 实例
  - [x] builder.go - Builder API
  - [x] grid_test.go - 单元测试
  - [x] 布局引擎集成 (GridStyleProvider)
  - [x] 示例验证 (fiber_firsts/grid_demo)

- [x] **Stack (VStack/HStack)** (`ui/components/stack/`)
  - [x] vnode.go - VNode 定义
  - [x] instance.go - Fiber 实例
  - [x] builder.go - Builder API
  - [x] stack_test.go - 单元测试
  - [x] 布局引擎集成

### 基础组件

- [x] **Text** (`ui/components/text/`)
  - [x] vnode.go
  - [x] instance.go
  - [x] builder.go
  - [x] text_test.go

- [x] **Divider** (`ui/components/divider/`)
  - [x] vnode.go
  - [x] instance.go
  - [x] builder.go
  - [x] divider_test.go

- [x] **Border** (`ui/components/border/`)
  - [x] vnode.go
  - [x] instance.go
  - [x] builder.go

### 表单组件

- [x] **Button** (`ui/components/button/`)
  - [x] vnode.go
  - [x] instance.go
  - [x] builder.go
  - [x] button_test.go

- [x] **Checkbox** (`ui/components/checkbox/`)
  - [x] vnode.go
  - [x] instance.go
  - [x] builder.go
  - [x] checkbox_test.go

- [x] **Input** (`ui/components/input/`)
  - [x] vnode.go
  - [x] instance.go
  - [x] builder.go
  - [x] input_test.go

- [x] **Select** (`ui/components/select/`)
  - [x] vnode.go
  - [x] instance.go
  - [x] builder.go
  - [x] select_test.go

- [x] **Textarea** (`ui/components/textarea/`)
  - [x] vnode.go
  - [x] instance.go
  - [x] builder.go
  - [x] textarea_test.go

- [x] **Progress** (`ui/components/progress/`)
  - [x] vnode.go
  - [x] instance.go
  - [x] builder.go
  - [x] progress_test.go

---

## 第一阶段：基础布局组件

### Wrap (换行布局)

> **参考**: 已有实现 `components/layout/wrap.go` (377行)
> 
> 详细实现文档:
> - [wrap_component_implementation.md](./wrap_component_implementation.md)
> - [wrap_implementation_summary.md](./wrap_implementation_summary.md)

- [x] **分析现有实现**
  - [x] 阅读 `components/layout/wrap.go` (377行)
  - [x] 理解换行算法
  - [x] 确定依赖项

- [ ] **创建目录结构**
  - [ ] `ui/components/wrap/vnode.go`
  - [ ] `ui/components/wrap/instance.go`
  - [ ] `ui/components/wrap/builder.go`

- [ ] **实现 VNode**
  - [ ] 定义 WrapVNode 结构体
  - [ ] 实现 SetGap, SetRowGap, SetAlign, SetWidth
  - [ ] 实现 SetChildrenList, SetChildrenAuto

- [ ] **实现 Instance**
  - [ ] 创建 WrapInstance 结构体
  - [ ] 实现 UpdateInstance
  - [ ] 实现 Measure (换行计算)

- [ ] **布局引擎集成**
  - [ ] 定义 WrapStyleProvider 接口 (如需要)
  - [ ] 在 FiberToNodeAdapter 中添加支持

- [ ] **测试**
  - [ ] 参考 `components/layout/wrap_test.go`
  - [ ] 创建 `ui/components/wrap/wrap_test.go`
  - [ ] 测试基本换行
  - [ ] 测试间隙
  - [ ] 测试对齐

- [ ] **示例**
  - [ ] 创建 fiber_firsts/wrap_demo

---

### ScrollView (滚动视图)

- [ ] **分析现有实现**
  - [ ] 阅读 `components/layout/scroll_view.go` (510行)
  - [ ] 理解滚动机制
  - [ ] 理解视口裁剪

- [ ] **创建目录结构**
  - [ ] `ui/components/scroll/vnode.go`
  - [ ] `ui/components/scroll/instance.go`
  - [ ] `ui/components/scroll/builder.go`

- [ ] **实现 VNode**
  - [ ] 定义 ScrollVNode 结构体
  - [ ] 实现 SetContent, SetWidth, SetHeight
  - [ ] 实现 SetScrollOffset

- [ ] **实现 Instance**
  - [ ] 创建 ScrollInstance 结构体
  - [ ] 实现 UpdateInstance
  - [ ] 实现 Measure
  - [ ] 实现滚动状态管理

- [ ] **布局引擎集成**
  - [ ] 定义 ScrollViewStyleProvider 接口
  - [ ] 处理视口裁剪

- [ ] **Action 支持**
  - [ ] 实现 ScrollableActionTarget

- [ ] **测试**
  - [ ] 创建 scroll_test.go
  - [ ] 测试滚动偏移
  - [ ] 测试视口裁剪

- [ ] **示例**
  - [ ] 创建 fiber_firsts/scroll_demo

---

### Panel (面板容器)

- [ ] **分析现有实现**
  - [ ] 阅读 `components/container/panel.go` (245行)
  - [ ] 理解 header/content/footer 布局

- [ ] **创建目录结构**
  - [ ] `ui/components/panel/vnode.go`
  - [ ] `ui/components/panel/instance.go`
  - [ ] `ui/components/panel/builder.go`

- [ ] **实现 VNode**
  - [ ] 定义 PanelVNode 结构体
  - [ ] 实现 SetTitle, SetHeader, SetContent, SetFooter
  - [ ] 实现 SetBorderStyle, SetPadding

- [ ] **实现 Instance**
  - [ ] 考虑基于 VStack + Border 组合实现
  - [ ] 实现 UpdateInstance
  - [ ] 实现 Measure

- [ ] **事件转发**
  - [ ] 实现 HandleEvent
  - [ ] 转发事件到 content

- [ ] **测试**
  - [ ] 创建 panel_test.go
  - [ ] 测试布局
  - [ ] 测试事件转发

- [ ] **示例**
  - [ ] 创建 fiber_firsts/panel_demo

---

## 第二阶段：数据展示组件

### Table (表格)

- [ ] **分析现有实现**
  - [ ] 阅读 `components/data/table.go` (320行)
  - [ ] 确定是否可以基于 Grid 实现

- [ ] **创建目录结构**
  - [ ] `ui/components/table/vnode.go`
  - [ ] `ui/components/table/instance.go`
  - [ ] `ui/components/table/builder.go`

- [ ] **实现 VNode**
  - [ ] 定义 TableVNode 结构体
  - [ ] 实现 SetColumns, SetRows
  - [ ] 实现 SetHeaderStyle

- [ ] **实现 Instance**
  - [ ] 考虑基于 Grid 实现
  - [ ] 实现 UpdateInstance

- [ ] **测试**
  - [ ] 创建 table_test.go

- [ ] **示例**
  - [ ] 创建 fiber_firsts/table_demo

---

### List (列表)

- [ ] **分析现有实现**
  - [ ] 阅读 `components/data/list.go` (1096行)
  - [ ] 理解选择、滚动、焦点机制

- [ ] **创建目录结构**
  - [ ] `ui/components/list/vnode.go`
  - [ ] `ui/components/list/instance.go`
  - [ ] `ui/components/list/builder.go`

- [ ] **实现 VNode**
  - [ ] 定义 ListVNode 结构体
  - [ ] 实现 SetHeader, SetRows, SetSeparator
  - [ ] 实现 SetRowStyle 回调

- [ ] **实现 Instance**
  - [ ] 实现 UpdateInstance
  - [ ] 实现 Measure (宽度约束)
  - [ ] 实现状态管理 (选择、焦点、滚动)

- [ ] **Action 支持**
  - [ ] 实现 FocusableActionTarget
  - [ ] 实现 ScrollableActionTarget
  - [ ] 实现 SelectableActionTarget

- [ ] **测试**
  - [ ] 创建 list_test.go
  - [ ] 测试选择
  - [ ] 测试滚动
  - [ ] 测试键盘导航

- [ ] **示例**
  - [ ] 创建 fiber_firsts/list_demo

---

### VirtualScroll (虚拟滚动)

- [ ] **分析现有实现**
  - [ ] 阅读 `components/layout/virtual_scroll.go` (474行)
  - [ ] 比较 `components/data/virtuallist.go`
  - [ ] 确定是否合并

- [ ] **创建目录结构**
  - [ ] `ui/components/virtualscroll/vnode.go`
  - [ ] `ui/components/virtualscroll/instance.go`
  - [ ] `ui/components/virtualscroll/builder.go`

- [ ] **实现 VNode**
  - [ ] 定义 VirtualScrollVNode 结构体
  - [ ] 实现 SetItemCount, SetItemHeight
  - [ ] 实现 SetViewportHeight, SetScrollOffset
  - [ ] 实现 SetRenderItem 回调

- [ ] **实现 Instance**
  - [ ] 实现 UpdateInstance
  - [ ] 实现可见项计算
  - [ ] 实现增量更新

- [ ] **Action 支持**
  - [ ] 实现 ScrollableActionTarget

- [ ] **测试**
  - [ ] 创建 virtualscroll_test.go

- [ ] **示例**
  - [ ] 创建 fiber_firsts/virtualscroll_demo

---

## 第三阶段：交互组件

### Modal (模态对话框)

- [ ] **分析现有实现**
  - [ ] 阅读 `components/overlay/modal.go` (444行)

- [ ] **创建目录结构**
  - [ ] `ui/components/modal/vnode.go`
  - [ ] `ui/components/modal/instance.go`
  - [ ] `ui/components/modal/builder.go`

- [ ] **实现 VNode**
  - [ ] 定义 ModalVNode 结构体
  - [ ] 实现 SetTitle, SetContent, SetFooter
  - [ ] 实现 SetOpen, SetOnClose
  - [ ] 实现 SetWidth, SetHeight, SetCentered

- [ ] **实现 Instance**
  - [ ] 实现模态显示/隐藏
  - [ ] 实现焦点捕获
  - [ ] 实现 ESC 关闭

- [ ] **层叠上下文**
  - [ ] 设计 z-index 机制

- [ ] **测试**
  - [ ] 创建 modal_test.go

- [ ] **示例**
  - [ ] 创建 fiber_firsts/modal_demo

---

### Tooltip (提示框)

- [ ] **分析现有实现**
  - [ ] 阅读 `components/overlay/tooltip.go` (602行)

- [ ] **创建目录结构**
  - [ ] `ui/components/tooltip/vnode.go`
  - [ ] `ui/components/tooltip/instance.go`
  - [ ] `ui/components/tooltip/builder.go`

- [ ] **实现 VNode**
  - [ ] 定义 TooltipVNode 结构体
  - [ ] 实现 SetContent, SetText
  - [ ] 实现 SetPosition, SetDelay

- [ ] **实现 Instance**
  - [ ] 实现定位算法
  - [ ] 实现边界检测
  - [ ] 实现延迟显示

- [ ] **测试**
  - [ ] 创建 tooltip_test.go

---

### Toast (吐司通知)

- [ ] **分析现有实现**
  - [ ] 阅读 `components/overlay/tooltip.go` 中的 Toast 部分

- [ ] **创建目录结构**
  - [ ] `ui/components/toast/vnode.go`
  - [ ] `ui/components/toast/instance.go`
  - [ ] `ui/components/toast/builder.go`

- [ ] **实现 VNode**
  - [ ] 定义 ToastVNode 结构体
  - [ ] 实现 SetMessage, SetDuration
  - [ ] 实现 SetPosition

- [ ] **实现 Instance**
  - [ ] 实现动画 (淡入/淡出)
  - [ ] 实现自动消失

- [ ] **测试**
  - [ ] 创建 toast_test.go

---

## 第四阶段：复杂组件

### Tabs (标签页)

- [ ] **分析现有实现**
  - [ ] 阅读 `components/navigation/tabs.go` (1098行)
  - [ ] 设计组件拆分方案

- [ ] **创建目录结构**
  - [ ] `ui/components/tabs/vnode.go`
  - [ ] `ui/components/tabs/instance.go`
  - [ ] `ui/components/tabs/builder.go`
  - [ ] `ui/components/tabs/tab_item.go`

- [ ] **实现 VNode**
  - [ ] 定义 TabsVNode 结构体
  - [ ] 定义 TabItem 结构体
  - [ ] 实现 SetTabs, SetPosition
  - [ ] 实现 SetActiveTab, SetOnTabChange

- [ ] **实现 Instance**
  - [ ] 实现标签切换逻辑
  - [ ] 实现标签栏布局
  - [ ] 实现内容区切换

- [ ] **Action 支持**
  - [ ] 实现 SelectableActionTarget
  - [ ] 实现 ScrollableActionTarget

- [ ] **测试**
  - [ ] 创建 tabs_test.go
  - [ ] 测试标签切换
  - [ ] 测试键盘导航

- [ ] **示例**
  - [ ] 创建 fiber_firsts/tabs_demo

---

### TreeView (树形视图)

- [ ] **分析现有实现**
  - [ ] 阅读 `components/display/treeview.go` (1513行)
  - [ ] 设计组件拆分方案

- [ ] **创建目录结构**
  - [ ] `ui/components/treeview/vnode.go`
  - [ ] `ui/components/treeview/instance.go`
  - [ ] `ui/components/treeview/builder.go`
  - [ ] `ui/components/treeview/tree_node.go`
  - [ ] `ui/components/treeview/tree_line.go`

- [ ] **实现 VNode**
  - [ ] 定义 TreeViewVNode 结构体
  - [ ] 实现 SetNodes, SetExpandState
  - [ ] 实现 SetOnExpand, SetOnSelect

- [ ] **实现 Instance**
  - [ ] 实现展开/折叠逻辑
  - [ ] 实现选择逻辑
  - [ ] 实现虚拟滚动 (可选)

- [ ] **Action 支持**
  - [ ] 实现 FocusableActionTarget
  - [ ] 实现 ScrollableActionTarget
  - [ ] 实现 SelectableActionTarget
  - [ ] 实现 ExpandableActionTarget

- [ ] **测试**
  - [ ] 创建 treeview_test.go
  - [ ] 测试展开/折叠
  - [ ] 测试选择
  - [ ] 测试键盘导航

- [ ] **示例**
  - [ ] 创建 fiber_firsts/treeview_demo

---

## 进度跟踪

| 阶段 | 组件数 | 完成 | 进度 |
|------|--------|------|------|
| 已完成 | 13 | 13 | 100% |
| 第一阶段 | 3 | 0 | 0% |
| 第二阶段 | 3 | 0 | 0% |
| 第三阶段 | 3 | 0 | 0% |
| 第四阶段 | 2 | 0 | 0% |
| **总计** | **24** | **13** | **54%** |

---

## 更新日志

- **2026-02-21**: 创建迁移分析和 TODO list
  - 完成 13 个组件迁移 (布局 3, 基础 3, 表单 6, 进度 1)
  - 待迁移 11 个组件
