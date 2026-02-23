# Mint UI 组件迁移计划

本文档安排剩余组件从 `components/` 迁移到 Fiber-first 架构的 `ui/components/` 目录。

## 目录

- [概述](#概述)
- [组件复杂度分析](#组件复杂度分析)
- [迁移优先级和时间表](#迁移优先级和时间表)
- [详细迁移计划](#详细迁移计划)
- [依赖关系](#依赖关系)
- [验证标准](#验证标准)

---

## 概述

### 当前状态

| 状态 | 数量 | 组件 |
|------|------|------|
| 已迁移 | 16 | Button, Stack, Text, Divider, Input, Checkbox, Panel, Grid, ScrollView, Select, Textarea, Wrap, Absolute, Border, Progress, Tooltip |
| 待迁移 | 6 | Tabs, TreeView, Modal, List, VirtualList, Table |

### 迁移目标

1. **高优先级组件**（Tabs, TreeView）：先迁移，优先提供示例程序
2. **中优先级组件**：关注布局和状态管理
3. **低优先级组件**：逐步完善

---

## 组件复杂度分析

| 组件 | 旧路径 | 代码量 | 复杂度 | 主要挑战 | 估计工时 |
|------|--------|--------|--------|----------|----------|
| **TreeView** | `components/display/treeview.go` | ~1500行 | 很高 | 树结构管理、expand/collapse、虚拟滚动、焦点管理 | 5-7天 |
| **Tabs** | `components/navigation/tabs.go` | ~1000行 | 高 | activeTab状态、键盘导航、wrapTabs计算、回调 | 3-5天 |
| **List** | `components/data/list.go` | ~1000行 | 中 | 滚动、键盘导航、选择状态 | 2-3天 |
| **VirtualList** | `components/data/virtuallist.go` | ~650行 | 中 | 虚拟滚动、renderItem回调 | 2-3天 |
| **Modal** | `components/overlay/modal.go` | ~400行 | 中 | 层级管理、焦点管理、键盘关闭 | 2天 |
| **Tooltip** | `components/overlay/tooltip.go` | ~500行（含Toast） | 低-中 | 定位逻辑、延迟显示、Toast通知 | 1-2天 | ✅ 已完成 |
| **Table** | `components/data/table.go` | ~300行 | 低 | 表格布局、列宽管理 | 1天 |

---

## 迁移优先级和时间表

### Phase 1: 高优先级组件（Week 1-2）

**时间：2026-02-24 ~ 2026-03-09**

| 组件 | 开始日期 | 结束日期 | 负责人 |
|------|----------|----------|--------|
| Tabs | 2026-02-24 | 2026-02-28 | 待定 |
| TreeView | 2026-03-03 | 2026-03-09 | 待定 |

### Phase 2: 中优先级组件（Week 3-4）

**时间：2026-03-10 ~ 2026-03-23**

| 组件 | 开始日期 | 结束日期 | 负责人 |
|------|----------|----------|--------|
| List | 2026-03-10 | 2026-03-12 | 待定 |
| VirtualList | 2026-03-13 | 2026-03-15 | 待定 |
| Modal | 2026-03-17 | 2026-03-18 | 待定 |
| Tooltip | 2026-03-19 | 2026-03-20 | 待定 |

### Phase 3: 低优先级组件（Week 5）

**时间：2026-03-24 ~ 2026-03-28**

| 组件 | 开始日期 | 结束日期 | 负责人 |
|------|----------|----------|--------|
| Table | 2026-03-24 | 2026-03-26 | 待定 |

---

## 详细迁移计划

### 1. Tabs 组件迁移

**路径**：`components/navigation/tabs.go` → `ui/components/tabs/`

**功能特性**：
- Tab 切换（鼠标点击、键盘导航）
- Tab 位置（Top/Bottom/Left/Right）
- Tab 包装（WrapTabs）
- 禁用状态
- onChange 回调
- Builder API

**迁移任务**：

| 任务 | 描述 | 验证 |
|------|------|------|
| 1.1 分析旧代码 | 理解 TabsVNode 结构、状态管理 | 查看旧代码结构 |
| 1.2 创建新目录 | `ui/components/tabs/` | 目录存在 |
| 1.3 编写 vnode.go | 纯描述性 VNode、无状态、无闭包 | props 只读、无 Paint |
| 1.4 编写 instance.go | 运行时状态、Behavior 组合 | Measure正确、状态管理 |
| 1.5 编写 builder.go | Fluent API | 链式调用可用 |
| 1.6 键盘导航 | 实现左箭头/右箭头/Enter/Home/End | 键盘切换正常 |
| 1.7 Wrap 支持 | 实现 wrapTabs 计算 | 多行tab显示正确 |
| 1.8 单元测试 | 测试各种tab切换场景 | 所有测试通过 |
| 1.9 示例程序 | `examples/fiber_firsts/tabs_demo/` | 示例可运行 |

**代码结构**：
```
ui/components/tabs/
├── vnode.go        # TabPosition, TabItem, VNode
├── builder.go      # TabsBuilder + Builder
├── instance.go     # 运行时实例、状态管理
├── behavior.go     # Tabs专用Behavior（可选）
├── navigation.go   # 键盘导航逻辑
└── tabs_test.go
```

**关键挑战**：
1. activeTab 状态管理
2. wrapTabs 的计算逻辑
3. TabItem 内容的动态切换
4. onChange 回转换为 Intent

**Intent 设计**：
```go
type ActionType string

const (
    TabSwitch ActionType = "switch"  // 切换tab
    TabNavigate ActionType = "navigate"  // 键盘导航
)

type SwitchTabIntent struct {
    intent.IntentBase
    TabID string
    Index int
}
```

---

### 2. TreeView 组件迁移

**路径**：`components/display/treeview.go` → `ui/components/treeview/`

**功能特性**：
- 树形结构显示
- 展开/折叠 (expand/collapse)
- 虚拟滚动（大列表）
- 键盘导航（上下、Enter、Space）
- 鼠标点击选择
- 图标显示

**迁移任务**：

| 任务 | 描述 | 验证 |
|------|------|------|
| 2.1 分析旧代码 | 理解树结构、expand状态、虚拟滚动 | 查看旧代码结构 |
| 2.2 创建新目录 | `ui/components/treeview/` | 目录存在 |
| 2.3 编写 vnode.go | 纯描述性 VNode、树节点定义 | props 只读、无 Paint |
| 2.4 编写 instance.go | 运行时状态、树状态管理 | expand状态正确 |
| 2.5 编写 builder.go | Fluent API | 链式调用可用 |
| 2.6 虚拟滚动实现 | 只渲染可见节点 | 性能测试通过 |
| 2.7 键盘导航 | 上下、Enter（展开/折叠） | 导航正常 |
| 2.8 单元测试 | 测试各种树操作 | 所有测试通过 |
| 2.9 示例程序 | `examples/fiber_firsts/treeview_demo/` | 示例可运行 |

**代码结构**：
```
ui/components/treeview/
├── vnode.go        # TreeViewLine, VNode
├── builder.go      # TreeViewBuilder + Builder
├── instance.go     # 运行时实例、树状态
├── tree.go         # 树结构管理
├── virtual.go      # 虚拟滚动实现
├── navigation.go   # 键盘导航逻辑
└── treeview_test.go
```

**关键挑战**：
1. 树结构的状态管理（expand/collapse）
2. 虚拟滚动的实现
3. 焦点管理（focusIndex）
4. 路径稳定化（Path 用于diffing）

**Intent 设计**：
```go
type ActionType string

const (
    NodeExpand   ActionType = "expand"    // 展开节点
    NodeCollapse ActionType = "collapse"  // 折叠节点
    NodeSelect   ActionType = "select"    // 选择节点
    NodeNavigate ActionType = "navigate"  // 键盘导航
)

type ExpandNodeIntent struct {
    intent.IntentBase
    NodeID int
    Path   string  // 节点路径
}
```

---

### 3. List 组件迁移

**路径**：`components/data/list.go` → `ui/components/list/`

**功能特性**：
- 列表显示
- 滚动支持
- 选择状态
- 键盘导航
- 鼠标点击
- 头部样式

**迁移任务**：

| 任务 | 描述 | 验证 |
|------|------|------|
| 3.1 分析旧代码 | 理解列表结构、滚动、选择 | 查看旧代码结构 |
| 3.2 创建新目录 | `ui/components/list/` | 目录存在 |
| 3.3 编写 vnode.go | 纯描述性 VNode | props 只读 |
| 3.4 编写 instance.go | 运行时状态、滚动实现 | Measure正确、滚动正常 |
| 3.5 编写 builder.go | Fluent API | 链式调用可用 |
| 3.6 单元测试 | 测试列表操作 | 所有测试通过 |
| 3.7 示例程序 | `examples/fiber_firsts/list_demo/` | 示例可运行 |

**代码结构**：
```
ui/components/list/
├── vnode.go        # VNode
├── builder.go      # ListBuilder + Builder
├── instance.go     # 运行时实例、滚动
├── scroll.go       # 滚动逻辑
└── list_test.go
```

---

### 4. VirtualList 组件迁移

**路径**：`components/data/virtuallist.go` → `ui/components/virtuallist/`

**功能特性**：
- 虚拟化列表
- 只渲染可见项
- 大数据集支持
- 滚动加载

**迁移任务**：

| 任务 | 描述 | 验证 |
|------|------|------|
| 4.1 分析旧代码 | 理解虚拟滚动逻辑 | 查看旧代码结构 |
| 4.2 创建新目录 | `ui/components/virtuallist/` | 目录存在 |
| 4.3 编写 vnode.go | 纯描述性 VNode | props 只读 |
| 4.4 编写 instance.go | 运行时实例、虚拟滚动 | Measure正确、渲染可见项 |
| 4.5 编写 builder.go | Fluent API | 链式调用可用 |
| 4.6 单元测试 | 测试虚拟滚动和性能 | 所有测试通过 |
| 4.7 示例程序 | `examples/fiber_firsts/virtuallist_demo/` | 示例可运行 |

**关键挑战**：
1. virtualRender 计算（只渲染可见项）
2. renderItem 回转换为 Props
3. 性能测试（大数据集）

---

### 5. Modal 组件迁移

**路径**：`components/overlay/modal.go` → `ui/components/modal/`

**功能特性**：
- 模态对话框
- 居中显示
- 键盘关闭（ESC）
- 边界检测

**迁移任务**：

| 任务 | 描述 | 验证 |
|------|------|------|
| 5.1 分析旧代码 | 理解modal逻辑、层级管理 | 查看旧代码结构 |
| 5.2 创建新目录 | `ui/components/modal/` | 目录存在 |
| 5.3 编写 vnode.go | 纯描述性 VNode | props 只读 |
| 5.4 编写 instance.go | 运行时状态、焦点管理 | 显示正确、焦点管理 |
| 5.5 编写 builder.go | Fluent API | 链式调用可用 |
| 5.6 层级集成 | 与 Fiber Layer 系统集成 | 层级正确 |
| 5.7 单元测试 | 测试modal开关和关闭 | 所有测试通过 |
| 5.8 示例程序 | `examples/fiber_firsts/modal_demo/` | 示例可运行 |

**关键挑战**：
1. Modal 层级管理（需要在最上层）
2. 焦点管理（打开时获取焦点、关闭时恢复）
3. 键盘事件（ESC关闭）

---

### 6. Tooltip 组件迁移

**路径**：`components/overlay/tooltip.go` → `ui/components/tooltip/`

**功能特性**：
- Tooltip 提示
- 定位逻辑（Top/Bottom/Left/Right/Auto）
- 延迟显示
- Toast 通知

**迁移任务**：

| 任务 | 描述 | 验证 |
|------|------|------|
| 6.1 分析旧代码 | 理解定位逻辑、延迟显示 | ✅ 已完成 |
| 6.2 创建新目录 | `ui/components/tooltip/` | ✅ 已完成 |
| 6.3 编写 vnode.go | 纯描述性 VNode | ✅ 已完成 |
| 6.4 编写 instance.go | 运行时实例、定位逻辑 | ✅ 已完成 |
| 6.5 编写 builder.go | Fluent API | ✅ 已完成 |
| 6.6 Toast 管理 | ToastManager 实现 | ✅ 已完成 |
| 6.7 单元测试 | 测试tooltip和toast | ✅ 所有测试通过 |
| 6.8 示例程序 | `examples/fiber_firsts/tooltip_demo/` | ✅ 示例可运行 |

**代码结构**：
```
ui/components/tooltip/
├── vnode.go        # TooltipVNode, ToastVNode
├── builder.go      # TooltipBuilder + ToastBuilder
├── instance.go     # Tooltip实例
├── toast.go        # ToastVNode实例、ToastManager
└── tooltip_test.go
```

---

### 7. Table 组件迁移

**路径**：`components/data/table.go` → `ui/components/table/`

**功能特性**：
- 表格显示
- 列宽管理
- 头部样式

**迁移任务**：

| 任务 | 描述 | 验证 |
|------|------|------|
| 7.1 分析旧代码 | 理解表格布局、列宽 | 查看旧代码结构 |
| 7.2 创建新目录 | `ui/components/table/` | 目录存在 |
| 7.3 编写 vnode.go | 纯描述性 VNode | props 只读 |
| 7.4 编写 instance.go | 运行时实例、表格布局 | Measure正确 |
| 7.5 编写 builder.go | Fluent API | 链式调用可用 |
| 7.6 单元测试 | 测试表格显示 | 所有测试通过 |
| 7.7 示例程序 | `examples/fiber_firsts/table_demo/` | 示例可运行 |

---

## 依赖关系

```
┌─────────────────────────────────────────────────────────────────┐
│                         迁移依赖图                               │
└─────────────────────────────────────────────────────────────────┘

Panel (已迁移)
    │
    ├──> ScrollContainer (已迁移)
    │         │
    │         ├──> VirtualList (待迁移)
    │         │         │
    │         │         └──> TreeView (待迁移)
    │         │
    │         └──> List (待迁移)
    │
    └──> Stack (已迁移)
              │
              ├──> Tabs (待迁移)
              └──> Modal (待迁移)
                        │
                        └──> Tooltip (待迁移)

Table (待迁移) - 独立组件
```

**迁移建议**：
1. 先迁移 Toast（简单，独立）
2. 再迁移 List/VirtualList（中等复杂度）
3. 然后迁移 Tabs（基于Stack）
4. 接着 Modal（需要Toast支持）
5. 最后 TreeView（最复杂）
6. Table 最后（独立，简单）

---

## 验证标准

### 通用验证

每个组件迁移完成后必须通过以下验证：

- [ ] **VNode 验证**
  - [ ] 只包含声明性属性
  - [ ] 无状态字段
  - [ ] 无闭包字段
  - [ ] 无 Paint() 方法
  - [ ] 实现 rtui.InstanceFactory

- [ ] **Instance 验证**
  - [ ] 实现 ComponentInstance
  - [ ] 实现 PaintableInstance（如需）
  - [ ] 实现 FocusableInstance（如有焦点）
  - [ ] 实现 ActionHandlerInstance
  - [ ] Measure() 使用 layout.Constraints
  - [ ] 使用 Behavior 组合

- [ ] **Builder 验证**
  - [ ] 所有 setter 返回 *Builder
  - [ ] Build() 返回 rtui.VNode

- [ ] **测试验证**
  - [ ] 单元测试覆盖 VNode
  - [ ] 单元测试覆盖 Instance
  - [ ] 所有测试通过

- [ ] **布局特性验证**（如适用）
  - [ ] 约束传递正确
  - [ ] 支持约束追踪
  - [ ] 边框内边距正确计算

- [ ] **示例程序验证**
  - [ ] 示例程序位于 `examples/fiber_firsts/<component>_demo/`
  - [ ] 示例程序可以独立运行
  - [ ] 示例程序展示组件特性

### 特定组件验证

**Tabs 验证**：
- [ ] Tab 切换正常（鼠标/键盘）
- [ ] wrapTabs 多行显示正确
- [ ] 禁用tab不可点击
- [ ] onChange 回调触发

**TreeView 验证**：
- [ ] 展开/折叠正常
- [ ] 虚拟滚动性能良好
- [ ] 键盘导航正确
- [ ] 鼠标点击选择正常

**List/VirtualList 验证**：
- [ ] 滚动流畅
- [ ] 键盘导航正确
- [ ] 选择状态正确
- [ ] VirtualList 只渲染可见项

**Modal 验证**：
- [ ] Modal 显示/隐藏正确
- [ ] 居中显示正确
- [ ] ESC 关闭
- [ ] 焦点管理正确

**Tooltip/Toast 验证**：
- [ ] Tooltip 定位正确
- [ ] 延迟显示正常
- [ ] Toast 显示/关闭正常
- [ ] ToastManager 功能正常

**Table 验证**：
- [ ] 表格显示正确
- [ ] 列宽计算正确
- [ ] 头部样式应用正确

---

## 成功指标

### 数量指标

| 指标 | 目标 | 当前 |
|------|------|------|
| 已迁移组件数 | 22 | 15/22 |
| 示例程序覆盖率 | 100% | 15/22 |
| 测试覆盖率 | ≥80% | 待验证 |

### 质量指标

| 指标 | 目标 |
|------|------|
| 所有组件通过验证 | ✅ |
| 示例程序全部可运行 | ✅ |
| 测试覆盖率 ≥80% | ✅ |
| 无 P0/P1 bug | ✅ |

### 时间指标

| 阶段 | 目标日期 |
|------|----------|
| Phase 1 完成 | 2026-03-09 |
| Phase 2 完成 | 2026-03-23 |
| Phase 3 完成 | 2026-03-28 |

---

## 附录

### A. 参考文档

- `ui/components/COMPONENT_MIGRATION_GUIDE.md` - 迁移指导文档
- `docs/layout/plan/` - 布局系统特性
- `ui/components/button/` - 迁移参考示例
- `ui/components/stack/` - 布局组件参考

### B. 示例程序模板

参考 `examples/fiber_firsts/stack_demo/main.go`

### C. 已迁移组件列表

| 组件 | 路径 | 示例程序 |
|------|------|----------|
| Button | `ui/components/button/` | ❌ |
| Stack | `ui/components/stack/` | ✅ |
| Text | `ui/components/text/` | ✅ |
| Divider | `ui/components/divider/` | ✅ |
| Input | `ui/components/input/` | ✅ |
| Checkbox | `ui/components/checkbox/` | ✅ |
| Panel | `ui/components/panel/` | ✅ |
| Grid | `ui/components/grid/` | ✅ |
| ScrollView | `ui/components/scrollview/` | ✅ |
| Select | `ui/components/select/` | ✅ |
| Textarea | `ui/components/textarea/` | ✅ |
| Wrap | `ui/components/wrap/` | ✅ |
| Absolute | `ui/components/absolute/` | ✅ |
| Border | `ui/components/border/` | ✅ |
| Progress | `ui/components/progress/` | ✅ |

---

**文档版本**：1.0
**创建日期**：2026-02-23
**最后更新**：2026-02-23
