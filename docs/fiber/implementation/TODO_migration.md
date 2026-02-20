# Fiber-first 渲染管线迁移待办事项

> 创建时间：2026-02-20
> 状态：进行中

---

## 一、当前状态分析

### 1.1 核心原则检查

| 原则 | 状态 | 说明 |
|------|------|------|
| **VNode 只存在于 Reconcile 阶段** | ⚠️ 部分符合 | fiberFirstPaint 路径符合，但 legacy 路径仍保留 VNode |
| **Layout 只读 Fiber** | ✅ 符合 | FiberToNodeAdapter 已清理，不再有特殊处理 |
| **Paint 只用 PaintableBox** | ✅ 符合 | PaintEngine.PaintLayout 只接受 PaintableLayout |

### 1.2 各阶段实施情况

#### Phase 1: Reconcile 阶段

**✅ Fiber-first 路径 (`fiberFirstPaint`)**
```go
// Phase 1: Fiber Reconciliation
// VNode is discarded after this
n.reconciler.Render(ctx, nullBuf, n.renderFn)
fiberRoot := n.getFiberRoot()
```

**❌ Legacy 路径 (`legacyPaint`)**
```go
// 仍然保存 VNode
n.root = n.renderWithFiberContext()
```

#### Phase 2: Layout 阶段

**✅ Fiber-first 路径**
- FiberToNodeAdapter.Measure - 已清理特殊处理
- 从 Fiber.Instance 获取尺寸（优先）
- 从 Fiber.Style/Props 获取（备选）
- 无 VNode 依赖

#### Phase 3: Paint 阶段

**✅ Fiber-first 路径**
```go
converter := NewFiberToPaintableConverter(fiberRoot)
paintableLayout := converter.ConvertToLayout(layoutBoxRoot)
n.paintEngine.PaintLayout(paintableLayout, buf)
```

**❌ Legacy 路径**
```go
n.PaintVNode(n.root, x, y, buf)  // 仍访问 VNode
```

---

## 二、待改进项

### 2.1 P1 - 核心架构清理

| 问题 | 位置 | 状态 | 说明 |
|------|------|------|------|
| `root rtui.VNode` 字段 | declarative_node.go:46 | 待处理 | 删除或移到 legacy 结构 |
| `legacyPaint` 方法 | declarative_node.go:524 | 待处理 | 保留但标记 deprecated |

### 2.2 P2 - Legacy 代码隔离

| 问题 | 位置 | 状态 | 说明 |
|------|------|------|------|
| `PaintVNode` 等方法 | declarative_node.go:792+ | 待处理 | 移到 legacy_paint.go |
| `renderWithFiberContext` | declarative_node.go:635 | 待处理 | 返回 VNode，违反原则 |
| `MeasureVNodeWidth/Height` | declarative_node.go:1070+ | 待处理 | Legacy 测量方法 |
| `expandComponents` | declarative_node.go:1298+ | 待处理 | VNode 展开逻辑 |

---

## 三、组件迁移清单

### 3.1 源目录：`components/`

#### basic/ (P0 - 基础组件)

| 文件 | 组件 | 状态 | 说明 |
|------|------|------|------|
| text.go | Text | ✅ 已迁移 | `ui/components/text/` |
| divider.go | Divider | ⬜ 待迁移 | 分隔线 |

#### button/ (P0 - ✅ 已完成)

| 文件 | 组件 | 状态 | 说明 |
|------|------|------|------|
| *.go | Button | ✅ 已迁移 | `ui/components/button/` |

#### layout/ (P0 - 布局组件)

| 文件 | 组件 | 状态 | 说明 |
|------|------|------|------|
| stack.go | VStack/HStack | ⬜ 待迁移 | 垂直/水平堆叠 |
| grid.go | Grid | ⬜ 待迁移 | 网格布局 |
| scroll_view.go | ScrollView | ⬜ 待迁移 | 滚动视图 |
| wrap.go | Wrap | ⬜ 待迁移 | 换行布局 |
| absolute.go | Absolute | ⬜ 待迁移 | 绝对定位 |

#### container/ (P1 - 容器组件)

| 文件 | 组件 | 状态 | 说明 |
|------|------|------|------|
| panel.go | Panel | ⬜ 待迁移 | 面板容器 |

#### display/ (P1 - 展示组件)

| 文件 | 组件 | 状态 | 说明 |
|------|------|------|------|
| treeview.go | TreeView | ⬜ 待迁移 | 树形视图 |

#### form/ (P1 - 表单组件)

| 文件 | 组件 | 状态 | 说明 |
|------|------|------|------|
| checkbox.go | Checkbox | ⬜ 待迁移 | 复选框 |
| input.go | Input | ⬜ 待迁移 | 输入框 |
| select.go | Select | ⬜ 待迁移 | 下拉选择 |
| textarea.go | TextArea | ⬜ 待迁移 | 多行文本 |

#### data/ (P2 - 数据组件)

| 文件 | 组件 | 状态 | 说明 |
|------|------|------|------|
| list.go | List | ⬜ 待迁移 | 列表 |
| table.go | Table | ⬜ 待迁移 | 表格 |
| virtuallist.go | VirtualList | ⬜ 待迁移 | 虚拟列表 |

#### feedback/ (P2 - 反馈组件)

| 文件 | 组件 | 状态 | 说明 |
|------|------|------|------|
| progress.go | Progress | ⬜ 待迁移 | 进度条 |

#### navigation/ (P2 - 导航组件)

| 文件 | 组件 | 状态 | 说明 |
|------|------|------|------|
| tabs.go | Tabs | ⬜ 待迁移 | 标签页 |

#### overlay/ (P2 - 覆盖层组件)

| 文件 | 组件 | 状态 | 说明 |
|------|------|------|------|
| modal.go | Modal | ⬜ 待迁移 | 模态框 |
| tooltip.go | Tooltip | ⬜ 待迁移 | 提示框 |

### 3.2 目标目录：`ui/components/`

每个组件迁移后目录结构：
```
ui/components/{component}/
    vnode.go      # 描述层：VNode、Props、Builder
    instance.go   # 运行层：Instance、Measure、Paint、Behaviors
    builder.go    # 构建器：流畅 API
    types.go      # 类型定义
    test.go       # 单元测试
```

---

## 四、迁移任务详情

### 4.1 Button ✅ (已完成)

- **状态**：已迁移
- **位置**：`ui/components/button/`
- **测试**：`runtime/ui/fiber_adapter_measure_test.go`
- **示例**：`examples/fiber_firsts/fiber_first_demo/`

### 4.2 Text (P0 - ✅ 已完成)

- **源位置**：`components/basic/text.go`
- **目标位置**：`ui/components/text/`
- **优先级**：P0
- **状态**：✅ 已完成

**已完成任务**：
- [x] 创建 `ui/components/text/` 目录
- [x] 实现 `vnode.go` (TextVNode, Props)
- [x] 实现 `instance.go` (TextInstance, Measure, Paint)
- [x] 实现 `builder.go` (流畅 API)
- [x] 编写单元测试 (`ui/components/text/text_test.go`)
- [x] 创建示例渲染代码 (`examples/fiber_firsts/text_demo/`)
- [x] 验证 Fiber-first 渲染管线 (`runtime/ui/fiber_adapter_measure_test.go`)

### 4.3 Divider (P0 - 待迁移)

- **源位置**：`components/basic/divider.go`
- **目标位置**：`ui/components/divider/`
- **优先级**：P0
- **状态**：⬜ 待迁移

### 4.4 VStack/HStack (P0 - 待迁移)

- **源位置**：`components/layout/stack.go`
- **目标位置**：`ui/components/stack/`
- **优先级**：P0
- **状态**：⬜ 待迁移

### 4.5 Grid (P0 - 待迁移)

- **源位置**：`components/layout/grid.go`
- **目标位置**：`ui/components/grid/`
- **优先级**：P0
- **状态**：⬜ 待迁移

### 4.6 ScrollView (P0 - 待迁移)

- **源位置**：`components/layout/scroll_view.go`
- **目标位置**：`ui/components/scrollview/`
- **优先级**：P0
- **状态**：⬜ 待迁移

### 4.7 Wrap (P0 - 待迁移)

- **源位置**：`components/layout/wrap.go`
- **目标位置**：`ui/components/wrap/`
- **优先级**：P0
- **状态**：⬜ 待迁移

### 4.8 Panel (P1 - 待迁移)

- **源位置**：`components/container/panel.go`
- **目标位置**：`ui/components/panel/`
- **优先级**：P1
- **状态**：⬜ 待迁移

### 4.9 TreeView (P1 - 待迁移)

- **源位置**：`components/display/treeview.go`
- **目标位置**：`ui/components/treeview/`
- **优先级**：P1
- **状态**：⬜ 待迁移

### 4.10 Checkbox (P1 - 待迁移)

- **源位置**：`components/form/checkbox.go`
- **目标位置**：`ui/components/checkbox/`
- **优先级**：P1
- **状态**：⬜ 待迁移

### 4.11 Input (P1 - 待迁移)

- **源位置**：`components/form/input.go`
- **目标位置**：`ui/components/input/`
- **优先级**：P1
- **状态**：⬜ 待迁移

### 4.12 Select (P1 - 待迁移)

- **源位置**：`components/form/select.go`
- **目标位置**：`ui/components/select/`
- **优先级**：P1
- **状态**：⬜ 待迁移

### 4.13 TextArea (P1 - 待迁移)

- **源位置**：`components/form/textarea.go`
- **目标位置**：`ui/components/textarea/`
- **优先级**：P1
- **状态**：⬜ 待迁移

### 4.14 List (P2 - 待迁移)

- **源位置**：`components/data/list.go`
- **目标位置**：`ui/components/list/`
- **优先级**：P2
- **状态**：⬜ 待迁移

### 4.15 Table (P2 - 待迁移)

- **源位置**：`components/data/table.go`
- **目标位置**：`ui/components/table/`
- **优先级**：P2
- **状态**：⬜ 待迁移

### 4.16 VirtualList (P2 - 待迁移)

- **源位置**：`components/data/virtuallist.go`
- **目标位置**：`ui/components/virtuallist/`
- **优先级**：P2
- **状态**：⬜ 待迁移

### 4.17 Progress (P2 - 待迁移)

- **源位置**：`components/feedback/progress.go`
- **目标位置**：`ui/components/progress/`
- **优先级**：P2
- **状态**：⬜ 待迁移

### 4.18 Tabs (P2 - 待迁移)

- **源位置**：`components/navigation/tabs.go`
- **目标位置**：`ui/components/tabs/`
- **优先级**：P2
- **状态**：⬜ 待迁移

### 4.19 Modal (P2 - 待迁移)

- **源位置**：`components/overlay/modal.go`
- **目标位置**：`ui/components/modal/`
- **优先级**：P2
- **状态**：⬜ 待迁移

### 4.20 Tooltip (P2 - 待迁移)

- **源位置**：`components/overlay/tooltip.go`
- **目标位置**：`ui/components/tooltip/`
- **优先级**：P2
- **状态**：⬜ 待迁移

---

## 五、迁移模板

按照 `docs/fiber/fiber_first/fiber_button_v2.md` 规范：

### 5.1 VNode 必须实现

```go
var (
    _ rtui.VNode           = (*VNode)(nil)
    _ rtui.InstanceFactory = (*VNode)(nil)
    _ rtui.FocusableVNode  = (*VNode)(nil)  // 如果可聚焦
    _ rtui.BoxModel        = (*VNode)(nil)
)
```

### 5.2 Instance 必须实现

```go
var (
    _ rtui.ComponentInstance     = (*Instance)(nil)
    _ rtui.PaintableInstance     = (*Instance)(nil)
    _ rtui.FocusableInstance     = (*Instance)(nil)  // 如果可聚焦
    _ rtui.ActionHandlerInstance = (*Instance)(nil)  // 如果处理 Action
    _ control.Instance           = (*Instance)(nil)
    _ interface {
        Measure(layout.Constraints) layout.Size
    } = (*Instance)(nil)  // ⭐ 必须
)
```

### 5.3 测试要求

每个组件迁移后必须：
1. **单元测试**：测试 Measure 计算正确性
2. **Fiber 集成测试**：测试 FiberToNodeAdapter 调用
3. **示例渲染**：创建独立示例验证渲染输出

---

## 六、完成标准

### 6.1 组件迁移完成标准

- [ ] VNode 只包含声明性 Props
- [ ] Instance 实现 Measure(layout.Constraints) layout.Size
- [ ] Instance 实现 Paint(x, y int) []paint.DrawCmd
- [ ] 无 VNode 闭包（OnClick → Intent）
- [ ] 单元测试通过
- [ ] 示例渲染正确

### 6.2 整体项目完成标准

- [ ] 所有 components/ 目录组件迁移到 ui/components/
- [ ] fiberFirstPaint 路径覆盖 100% 场景
- [ ] Legacy 路径代码隔离/删除
- [ ] 所有测试通过
- [ ] 文档更新

---

## 七、时间线

| 阶段 | 任务 | 状态 | 组件数 |
|------|------|------|--------|
| Phase 1 | Button 迁移 | ✅ 完成 | 1 |
| Phase 2 | Text 迁移 | ✅ 完成 | 1 |
| Phase 3 | Divider 迁移 | ⬜ 待开始 | 1 |
| Phase 4 | Stack + Grid + ScrollView + Wrap (layout) | ⬜ 待开始 | 4 |
| Phase 5 | Panel + TreeView | ⬜ 待开始 | 2 |
| Phase 6 | Checkbox + Input + Select + TextArea (form) | ⬜ 待开始 | 4 |
| Phase 7 | List + Table + VirtualList (data) | ⬜ 待开始 | 3 |
| Phase 8 | Progress + Tabs + Modal + Tooltip | ⬜ 待开始 | 4 |
| Phase 9 | Legacy 代码清理 | ⬜ 待开始 | - |

**总计**：20 个组件待迁移，2 个已完成

---

## 八、相关文档

- [Fiber-first Button 组件设计规范](../fiber_first/fiber_button_v2.md)
- [Fiber-first 渲染管线](../fiber_first/FIBER_FIRST_RENDER_PIPELINE.md)
- [两段布局引擎](../../layout/TWO_PASS_LAYOUT.md)
