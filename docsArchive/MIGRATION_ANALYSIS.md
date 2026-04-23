# Fiber-first 组件迁移分析报告

## 概述

本文档分析 `components/` 目录与 `ui/components/` 目录的组件迁移状态，为 Fiber-first 架构迁移提供执行计划。

## 目录结构对比

### 旧版目录 (components/)

```
components/
├── basic/           # 基础组件
│   ├── text.go
│   └── divider.go
├── button/          # 按钮组件
├── container/       # 容器组件
│   └── panel.go
├── data/            # 数据组件
│   ├── list.go
│   ├── table.go
│   └── virtuallist.go
├── display/         # 显示组件
│   └── treeview.go
├── feedback/        # 反馈组件
│   └── progress.go
├── form/            # 表单组件
├── layout/          # 布局组件
│   ├── absolute.go
│   ├── grid.go
│   ├── stack.go
│   ├── scroll_view.go
│   ├── wrap.go
│   └── virtual_scroll.go
├── navigation/      # 导航组件
│   └── tabs.go
└── overlay/         # 覆盖层组件
    ├── modal.go
    └── tooltip.go
```

### 新版目录 (ui/components/)

```
ui/components/
├── absolute/        # ✅ 绝对定位
├── border/          # ✅ 边框
├── button/          # ✅ 按钮
├── checkbox/        # ✅ 复选框
├── control/         # ✅ 控件类型
├── divider/         # ✅ 分隔线
├── grid/            # ✅ 网格布局
├── input/           # ✅ 输入框
├── progress/        # ✅ 进度条
├── select/          # ✅ 下拉选择
├── stack/           # ✅ 堆栈布局
├── text/            # ✅ 文本
└── textarea/        # ✅ 多行文本
```

## 迁移状态总览

| 分类 | 已迁移 | 待迁移 | 完成率 |
|------|--------|--------|--------|
| 布局组件 | 3 | 3 | 50% |
| 容器组件 | 0 | 1 | 0% |
| 覆盖层组件 | 0 | 2 | 0% |
| 导航组件 | 0 | 1 | 0% |
| 显示组件 | 0 | 1 | 0% |
| 数据组件 | 0 | 3 | 0% |
| **总计** | **3** | **11** | **21%** |

## 待迁移组件详细分析

### 1. 布局组件

#### 1.1 ScrollView (scroll_view.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/layout/scroll_view.go` |
| 代码行数 | 510 行 |
| 复杂度 | ⭐⭐⭐ 高 |
| 依赖 | action.ActionTarget, action.ScrollableActionTarget |

**功能说明：**
- 可滚动容器，裁剪内容到视口大小
- 支持垂直滚动
- 显示滚动位置指示器
- 支持键盘和鼠标滚动

**迁移要点：**
- 需要实现 `ScrollViewStyleProvider` 接口
- 滚动状态管理需要 Fiber 架构适配
- 视口裁剪逻辑需要与 Paint 阶段配合

**建议目录结构：**
```
ui/components/scroll/
├── vnode.go      # VNode 定义
├── instance.go   # Fiber 实例
├── builder.go    # Builder API
└── scroll_test.go
```

---

#### 1.2 Wrap (wrap.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/layout/wrap.go` |
| 代码行数 | 377 行 |
| 复杂度 | ⭐⭐ 中 |
| 依赖 | runtime.Measurable |
| 状态 | ✅ **已有实现** |

> **参考文档**:
> - [wrap_component_implementation.md](./wrap_component_implementation.md) - 详细实现计划
> - [wrap_implementation_summary.md](./wrap_implementation_summary.md) - 实现总结

**功能说明：**
- 换行布局容器，类似 CSS `flex-wrap: wrap`
- 支持间隙 (gap, rowGap)
- 支持对齐 (align)
- 自动计算行数

**迁移要点：**
- 需要实现 `WrapStyleProvider` 接口
- 需要在 Measure 阶段计算换行
- 布局引擎需要支持多行分布
- 可以参考现有的 `components/layout/wrap.go` 实现

**建议目录结构：**
```
ui/components/wrap/
├── vnode.go
├── instance.go
├── builder.go
└── wrap_test.go
```

---

#### 1.3 VirtualScroll (virtual_scroll.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/layout/virtual_scroll.go` |
| 代码行数 | 474 行 |
| 复杂度 | ⭐⭐⭐ 高 |
| 依赖 | action.ScrollableActionTarget |

**功能说明：**
- 虚拟滚动，只渲染可见项
- 类似 react-window / react-virtualized
- 支持动态项高度
- 支持滚动偏移

**迁移要点：**
- 需要与 Fiber 的协调机制配合
- 可见项计算需要在 Layout 阶段完成
- 需要支持增量更新

**建议目录结构：**
```
ui/components/virtualscroll/
├── vnode.go
├── instance.go
├── builder.go
└── virtualscroll_test.go
```

---

### 2. 容器组件

#### 2.1 Panel (panel.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/container/panel.go` |
| 代码行数 | 245 行 |
| 复杂度 | ⭐⭐ 中 |
| 依赖 | frameworkevent.Component, component.Updater, action.ActionTarget |

**功能说明：**
- 高级容器组件
- 管理边框、标题、内容布局
- 自动处理高度计算和 flex 分布
- 支持事件转发

**迁移要点：**
- 可以基于 VStack + Border 组合实现
- 需要处理 header/content/footer 布局
- 事件转发需要 Fiber 架构适配

**建议目录结构：**
```
ui/components/panel/
├── vnode.go
├── instance.go
├── builder.go
└── panel_test.go
```

---

### 3. 覆盖层组件

#### 3.1 Modal (overlay/modal.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/overlay/modal.go` |
| 代码行数 | 444 行 |
| 复杂度 | ⭐⭐ 中 |
| 依赖 | frameworkevent.Component, component.Updater, action.ActionTarget |

**功能说明：**
- 模态对话框
- 支持标题、内容、底部
- 支持居中显示
- 支持 onClose 回调

**迁移要点：**
- 需要支持 z-index 或层叠上下文
- 需要处理焦点捕获
- 需要支持键盘事件 (ESC 关闭)

**建议目录结构：**
```
ui/components/modal/
├── vnode.go
├── instance.go
├── builder.go
└── modal_test.go
```

---

#### 3.2 Tooltip (overlay/tooltip.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/overlay/tooltip.go` |
| 代码行数 | 602 行 |
| 复杂度 | ⭐⭐ 中 |
| 依赖 | action.ActionTarget |

**功能说明：**
- 提示框组件
- 支持多种位置 (top/bottom/left/right/auto)
- 支持延迟显示
- 包含 Toast 组件

**迁移要点：**
- 需要支持绝对定位
- 需要处理边界检测
- Toast 需要支持动画

**建议目录结构：**
```
ui/components/tooltip/
├── vnode.go
├── instance.go
├── builder.go
└── tooltip_test.go

ui/components/toast/
├── vnode.go
├── instance.go
├── builder.go
└── toast_test.go
```

---

### 4. 导航组件

#### 4.1 Tabs (navigation/tabs.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/navigation/tabs.go` |
| 代码行数 | 1098 行 |
| 复杂度 | ⭐⭐⭐⭐ 很高 |
| 依赖 | frameworkevent.Component, component.Updater, action.ActionTarget, action.ScrollableActionTarget, action.SelectableActionTarget |

**功能说明：**
- 标签页导航组件
- 支持多种位置 (top/bottom/left/right)
- 支持禁用状态
- 支持键盘导航
- 支持滚动标签栏

**迁移要点：**
- 需要支持状态管理 (当前选中标签)
- 需要处理标签栏与内容区的布局
- 需要支持事件冒泡和捕获
- 可能需要拆分为多个子组件

**建议目录结构：**
```
ui/components/tabs/
├── vnode.go
├── instance.go
├── builder.go
├── tab_item.go
└── tabs_test.go
```

---

### 5. 显示组件

#### 5.1 TreeView (display/treeview.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/display/treeview.go` |
| 代码行数 | 1513 行 |
| 复杂度 | ⭐⭐⭐⭐⭐ 最高 |
| 依赖 | runtime.Measurable, frameworkevent.Component, component.Updater, action.ActionTarget, action.FocusableActionTarget, action.ScrollableActionTarget, action.SelectableActionTarget |

**功能说明：**
- 树形视图组件
- 支持展开/折叠
- 支持选择
- 支持键盘导航
- 支持虚拟滚动
- 支持自定义图标

**迁移要点：**
- 这是最复杂的组件，建议最后迁移
- 需要支持大量状态 (展开、选择、焦点、滚动)
- 需要支持增量更新
- 建议拆分为 TreeView + TreeNode + TreeLine 等子组件

**建议目录结构：**
```
ui/components/treeview/
├── vnode.go
├── instance.go
├── builder.go
├── tree_node.go
├── tree_line.go
└── treeview_test.go
```

---

### 6. 数据组件

#### 6.1 List (data/list.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/data/list.go` |
| 代码行数 | 1096 行 |
| 复杂度 | ⭐⭐⭐ 高 |
| 依赖 | runtime.Measurable, frameworkevent.Component, component.Updater, action.ActionTarget, action.FocusableActionTarget, action.ScrollableActionTarget, action.SelectableActionTarget |

**功能说明：**
- 通用列表组件
- 支持表头
- 支持分隔符
- 支持键盘导航
- 支持鼠标选择
- 支持滚动
- 支持行样式回调

**迁移要点：**
- 需要支持选择状态
- 需要支持滚动状态
- 需要支持焦点状态
- 需要在 Measure 阶段处理宽度约束

**建议目录结构：**
```
ui/components/list/
├── vnode.go
├── instance.go
├── builder.go
└── list_test.go
```

---

#### 6.2 Table (data/table.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/data/table.go` |
| 代码行数 | 320 行 |
| 复杂度 | ⭐⭐ 中 |
| 依赖 | action.ActionTarget |

**功能说明：**
- 表格组件
- 支持列定义
- 支持表头样式
- 支持行数据

**迁移要点：**
- 可以基于 Grid 实现
- 需要处理列宽计算
- 需要支持表头样式

**建议目录结构：**
```
ui/components/table/
├── vnode.go
├── instance.go
├── builder.go
└── table_test.go
```

---

#### 6.3 VirtualList (data/virtuallist.go)

| 属性 | 值 |
|------|-----|
| 文件 | `components/data/virtuallist.go` |
| 代码行数 | 待确认 |
| 复杂度 | ⭐⭐⭐ 高 |
| 依赖 | 待确认 |

**功能说明：**
- 虚拟列表组件
- 只渲染可见项
- 类似 layout/virtual_scroll.go

**迁移要点：**
- 可能与 VirtualScroll 合并
- 需要确认与 layout/virtual_scroll.go 的区别

---

## 迁移优先级建议

基于复杂度和依赖关系，建议按以下顺序迁移：

### 第一阶段：基础布局组件 (优先级: 高)

| 顺序 | 组件 | 复杂度 | 预计工作量 |
|------|------|--------|------------|
| 1 | Wrap | ⭐⭐ | 1-2 天 |
| 2 | ScrollView | ⭐⭐⭐ | 2-3 天 |
| 3 | Panel | ⭐⭐ | 1 天 |

### 第二阶段：数据展示组件 (优先级: 中)

| 顺序 | 组件 | 复杂度 | 预计工作量 |
|------|------|--------|------------|
| 4 | Table | ⭐⭐ | 1 天 |
| 5 | List | ⭐⭐⭐ | 2-3 天 |
| 6 | VirtualScroll | ⭐⭐⭐ | 2-3 天 |

### 第三阶段：交互组件 (优先级: 中)

| 顺序 | 组件 | 复杂度 | 预计工作量 |
|------|------|--------|------------|
| 7 | Modal | ⭐⭐ | 1-2 天 |
| 8 | Tooltip + Toast | ⭐⭐ | 1-2 天 |

### 第四阶段：复杂组件 (优先级: 低)

| 顺序 | 组件 | 复杂度 | 预计工作量 |
|------|------|--------|------------|
| 9 | Tabs | ⭐⭐⭐⭐ | 3-5 天 |
| 10 | TreeView | ⭐⭐⭐⭐⭐ | 5-7 天 |

---

## 迁移模板

每个组件迁移应遵循以下结构：

```
ui/components/{component}/
├── vnode.go        # VNode 定义和构建器
├── instance.go     # Fiber 实例和协调逻辑
├── builder.go      # Builder API (可选，可与 vnode.go 合并)
└── {component}_test.go  # 单元测试
```

### VNode 模板

```go
package {component}

import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// {Component}VNode 定义
type {Component}VNode struct {
    rtui.VNodeBase
    // 组件特定属性
}

// New{Component} 创建新的 {Component}
func New{Component}() *{Component}VNode {
    return &{Component}VNode{
        VNodeBase: rtui.NewVNodeBase("{component}"),
    }
}

// {Component} 便捷函数
func {Component}() *{Component}VNode {
    return New{Component}()
}

// SetXxx 设置属性 (链式调用)
func (n *{Component}VNode) SetXxx(value Type) *{Component}VNode {
    n.xxx = value
    return n
}

// VNode 返回 VNode (实现 rtui.VNode 接口)
func (n *{Component}VNode) VNode() rtui.VNode {
    return n
}
```

### Instance 模板

```go
package {component}

import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// {Component}Instance Fiber 实例
type {Component}Instance struct {
    *rtui.Fiber
}

// New{Component}Instance 创建实例
func New{Component}Instance() *{Component}Instance {
    return &{Component}Instance{
        Fiber: rtui.NewFiber("{component}"),
    }
}

// UpdateInstance 更新实例 (实现 InstanceUpdater 接口)
func (i *{Component}Instance) UpdateInstance(newVNode rtui.VNode) {
    // 更新逻辑
}

// Measure 测量尺寸 (实现 Measurable 接口)
func (i *{Component}Instance) Measure(constraints layout.Constraints) layout.Size {
    // 测量逻辑
}
```

---

## 风险与挑战

### 1. 状态管理复杂度

- **Tabs, TreeView, List** 等组件有复杂的状态管理
- 需要设计 Fiber 架构下的状态管理方案

### 2. 事件处理

- 需要适配 `frameworkevent.Component` 接口
- 需要处理事件冒泡和捕获

### 3. 滚动支持

- ScrollView, VirtualScroll, List, TreeView 都需要滚动支持
- 需要设计统一的滚动机制

### 4. 层叠上下文

- Modal, Tooltip 需要支持 z-index
- 需要设计层叠上下文管理

### 5. 向后兼容

- 旧 API 需要保持兼容或提供迁移指南
- 需要考虑渐进式迁移

---

## 参考资料

- [Fiber 架构文档](../fiber/)
- [布局引擎文档](../layout/)
- [组件迁移指南](../compnents/COMPONENT_MIGRATION_GUIDE.md)
- [Absolute 组件实现](../../ui/components/absolute/)
- [Grid 组件实现](../../ui/components/grid/)
