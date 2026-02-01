# Mint UI 重构 TODO 清单

> **创建日期**: 2026-02-01
> **状态**: 📋 待执行
> **目标**: 将声明式组件从单一 declarativeRoot 架构重构为可独立渲染的多组件架构

---

## 目录

1. [Phase 0: 准备阶段](#phase-0-准备阶段)
2. [Phase 1: 基础架构重组](#phase-1-基础架构重组)
3. [Phase 2: 内部模块迁移](#phase-2-内部模块迁移)
4. [Phase 3: 组件库迁移](#phase-3-组件库迁移)
5. [Phase 4: 渲染系统重构](#phase-4-渲染系统重构)
6. [Phase 5: 多组件支持](#phase-5-多组件支持)
7. [Phase 6: API 入口层](#phase-6-api-入口层)
8. [Phase 7: 测试与验证](#phase-7-测试与验证)
9. [Phase 8: 文档更新](#phase-8-文档更新)

---

## Phase 0: 准备阶段

### 0.1 环境准备

- [ ] 创建 `components/` 目录结构
  - [ ] `components/basic/`
  - [ ] `components/layout/`
  - [ ] `components/form/`
  - [ ] `components/button/`
  - [ ] `components/feedback/`
  - [ ] `components/data/`
  - [ ] `components/navigation/`
  - [ ] `components/overlay/`
  - [ ] `components/container/`

- [ ] 创建 `internal/` 目录结构
  - [ ] `internal/reconciler/`
  - [ ] `internal/scheduler/`
  - [ ] `internal/state/`
  - [ ] `internal/render/`

- [ ] 创建各目录的 `doc.go` 说明文件
- [ ] 创建各目录的 `public.go` 接口文件

### 0.2 代码分析

- [ ] 统计 `ui/` 目录下所有文件及其依赖关系
- [ ] 识别需要迁移的文件列表
- [ ] 分析循环依赖风险
- [ ] 记录当前所有公开 API

### 0.3 测试基准

- [ ] 运行所有测试，记录通过率
- [ ] 运行所有示例，记录当前行为
- [ ] 创建视觉回归测试基线
- [ ] 记录性能基准数据

---

## Phase 1: 基础架构重组

### 1.1 核心接口定义

- [ ] `internal/render/component.go`
  - [ ] 定义 `Component` 接口
    - [ ] `ID() string`
    - [ ] `Type() string`
    - [ ] `Mount(ctx Context) error`
    - [ ] `Update(ctx Context) error`
    - [ ] `Unmount(ctx Context) error`
    - [ ] `Measure(constraints Constraints) Size`
    - [ ] `Paint(ctx PaintContext)`

  - [ ] 定义 `Constraints` 类型
  - [ ] 定义 `Size` 类型
  - [ ] 定义 `Context` 接口
  - [ ] 定义 `PaintContext` 接口

### 1.2 VNode 接口扩展

- [ ] `ui/vnode.go`
  - [ ] 添加 `RenderableVNode` 接口
    - [ ] 继承 `VNode`
    - [ ] `Measure(constraints Constraints) Size`
    - [ ] `Paint(ctx PaintContext)`

  - [ ] 为 `ElementVNode` 添加 `Measure()` 方法
  - [ ] 为 `ElementVNode` 添加 `Paint()` 方法

### 1.3 组件构建器基础

- [ ] `components/builder/base.go`
  - [ ] 定义 `Builder` 基础接口
  - [ ] 定义 `Buildable` 接口
  - [ ] 通用构建器辅助方法

---

## Phase 2: 内部模块迁移

### 2.1 Reconciler 系统

- [ ] `ui/fiber.go` → `internal/reconciler/fiber.go`
  - [ ] 修改 package 为 `reconciler`
  - [ ] 更新所有导入路径
  - [ ] 编写单元测试

- [ ] `ui/reconciler.go` → `internal/reconciler/reconciler.go`
  - [ ] 修改 package 为 `reconciler`
  - [ ] 更新所有导入路径
  - [ ] 编写单元测试

- [ ] `ui/diff.go` → `internal/reconciler/diff.go`
  - [ ] 修改 package 为 `reconciler`
  - [ ] 更新所有导入路径
  - [ ] 编写单元测试

- [ ] `ui/begin_work.go` → `internal/reconciler/begin_work.go`
  - [ ] 修改 package 为 `reconciler`
  - [ ] 更新所有导入路径

- [ ] `ui/complete_work.go` → `internal/reconciler/complete_work.go`
  - [ ] 修改 package 为 `reconciler`
  - [ ] 更新所有导入路径

- [ ] `internal/reconciler/public.go`
  - [ ] 定义公开接口
  - [ ] 导出必要的类型

### 2.2 Scheduler 系统

- [ ] `ui/scheduler.go` → `internal/scheduler/ui_scheduler.go`
  - [ ] 修改 package 为 `scheduler`
  - [ ] 更新所有导入路径
  - [ ] 编写单元测试

- [ ] `internal/scheduler/priority.go`
  - [ ] 定义优先级常量
  - [ ] 优先级队列实现

- [ ] `internal/scheduler/public.go`
  - [ ] 定义公开接口

### 2.3 State 系统

- [ ] `ui/instance.go` → `internal/state/instance.go`
  - [ ] 修改 package 为 `state`
  - [ ] 更新所有导入路径
  - [ ] 编写单元测试

- [ ] `ui/instance_manager.go` → `internal/state/instance_manager.go`
  - [ ] 修改 package 为 `state`
  - [ ] 更新所有导入路径

- [ ] `ui/interaction_state.go` → `internal/state/interaction_state.go`
  - [ ] 修改 package 为 `state`
  - [ ] 更新所有导入路径

- [ ] `internal/state/focus_manager.go` (新建)
  - [ ] `FocusManager` 结构
  - [ ] `FocusScope` 结构
  - [ ] `Register()` 方法
  - [ ] `Focus()` 方法
  - [ ] `Next()` / `Prev()` 方法
  - [ ] 编写单元测试

- [ ] `internal/state/public.go`
  - [ ] 定义公开接口

### 2.4 更新导入路径

- [ ] 更新 `ui/app.go` 中的所有导入
- [ ] 更新 `ui/hooks.go` 中的所有导入
- [ ] 更新所有组件文件中的导入
- [ ] 确保 `go build ./...` 通过

---

## Phase 3: 组件库迁移

### 3.1 基础组件 (components/basic/)

- [ ] `ui/text.go` → `components/basic/text.go`
  - [ ] 拆分为 `text.go` + `builder.go`
  - [ ] 实现 `Component` 接口
  - [ ] 实现 `Measure()` 方法
  - [ ] 实现 `Paint()` 方法
  - [ ] 编写单元测试

- [ ] `components/basic/icon.go` (新建)
  - [ ] 定义 Icon 组件
  - [ ] 实现 Builder 模式
  - [ ] 编写单元测试

- [ ] `components/basic/separator.go` (新建)
  - [ ] 定义 Separator 组件
  - [ ] 支持水平/垂直方向
  - [ ] 编写单元测试

- [ ] `components/basic/spacer.go` (新建)
  - [ ] 定义 Spacer 组件
  - [ ] 支持固定/弹性尺寸
  - [ ] 编写单元测试

- [ ] `components/basic/public.go`
  - [ ] 导出公开接口

### 3.2 布局组件 (components/layout/)

- [ ] `ui/layout.go` → `components/layout/stack.go`
  - [ ] 提取 HStack/VStack
  - [ ] 实现 `Component` 接口
  - [ ] 实现 `Measure()` 方法
  - [ ] 实现 `Paint()` 方法
  - [ ] 编写单元测试

- [ ] `ui/absolute.go` → `components/layout/absolute.go`
  - [ ] 实现绝对定位布局
  - [ ] 编写单元测试

- [ ] `ui/grid.go` → `components/layout/grid.go`
  - [ ] 实现网格布局
  - [ ] 编写单元测试

- [ ] `components/layout/flex.go` (新建)
  - [ ] 实现 Flex 布局
  - [ ] 支持 flex-grow/shrink
  - [ ] 编写单元测试

- [ ] `components/layout/box.go` (新建)
  - [ ] 实现 Box 容器
  - [ ] 支持 padding/margin
  - [ ] 编写单元测试

- [ ] `components/layout/public.go`
  - [ ] 导出公开接口

### 3.3 表单组件 (components/form/)

- [ ] `ui/input.go` → `components/form/input.go`
  - [ ] 重构为 TextInput 组件
  - [ ] 实现 `Component` 接口
  - [ ] 实现 Builder 模式
  - [ ] 编写单元测试

- [ ] `ui/textarea.go` → `components/form/textarea.go`
  - [ ] 重构为 TextArea 组件
  - [ ] 支持多行输入
  - [ ] 编写单元测试

- [ ] `ui/checkbox.go` → `components/form/checkbox.go`
  - [ ] 重构为 Checkbox 组件
  - [ ] 实现 Builder 模式
  - [ ] 编写单元测试

- [ ] `ui/select.go` → `components/form/select.go`
  - [ ] 重构为 Select 组件
  - [ ] 实现下拉选择
  - [ ] 编写单元测试

- [ ] `components/form/switch.go` (新建)
  - [ ] 实现 Switch 组件
  - [ ] 编写单元测试

- [ ] `components/form/slider.go` (新建)
  - [ ] 实现 Slider 组件
  - [ ] 编写单元测试

- [ ] `components/form/field.go` (新建)
  - [ ] 实现 Field 包装器
  - [ ] 支持标签/验证状态
  - [ ] 编写单元测试

- [ ] `components/form/public.go`
  - [ ] 导出公开接口

### 3.4 按钮组件 (components/button/)

- [ ] `ui/button.go` → `components/button/button.go`
  - [ ] 重构为 Button 组件
  - [ ] 实现 `Component` 接口
  - [ ] 实现 Builder 模式
  - [ ] 编写单元测试

- [ ] `components/button/icon_button.go` (新建)
  - [ ] 实现 IconButton 组件
  - [ ] 编写单元测试

- [ ] `components/button/button_group.go` (新建)
  - [ ] 实现 ButtonGroup 组件
  - [ ] 编写单元测试

- [ ] `components/button/public.go`
  - [ ] 导出公开接口

### 3.5 反馈组件 (components/feedback/)

- [ ] `ui/progress.go` → `components/feedback/progress.go`
  - [ ] 重构为 ProgressBar 组件
  - [ ] 支持水平/垂直
  - [ ] 编写单元测试

- [ ] `components/feedback/spinner.go` (新建)
  - [ ] 实现 Spinner 加载组件
  - [ ] 支持多种样式
  - [ ] 编写单元测试

- [ ] `components/feedback/toast.go` (新建)
  - [ ] 实现 Toast 通知组件
  - [ ] 支持自动消失
  - [ ] 编写单元测试

- [ ] `components/feedback/alert.go` (新建)
  - [ ] 实现 Alert 警告组件
  - [ ] 编写单元测试

- [ ] `components/feedback/badge.go` (新建)
  - [ ] 实现 Badge 徽章组件
  - [ ] 编写单元测试

- [ ] `components/feedback/public.go`
  - [ ] 导出公开接口

### 3.6 数据展示 (components/data/)

- [ ] `ui/virtuallist.go` → `components/data/virtuallist.go`
  - [ ] 重构为 VirtualList 组件
  - [ ] 实现虚拟滚动
  - [ ] 编写单元测试

- [ ] `components/data/list.go` (新建)
  - [ ] 实现 List 组件
  - [ ] 编写单元测试

- [ ] `components/data/table.go` (新建)
  - [ ] 实现 Table 组件
  - [ ] 支持排序/筛选
  - [ ] 编写单元测试

- [ ] `components/data/tree.go` (新建)
  - [ ] 实现 Tree 组件
  - [ ] 支持展开/折叠
  - [ ] 编写单元测试

- [ ] `components/data/public.go`
  - [ ] 导出公开接口

### 3.7 导航组件 (components/navigation/)

- [ ] `components/navigation/tabs.go` (新建)
  - [ ] 实现 Tabs 组件
  - [ ] 支持动态切换
  - [ ] 编写单元测试

- [ ] `components/navigation/menu.go` (新建)
  - [ ] 实现 Menu 组件
  - [ ] 支持嵌套菜单
  - [ ] 编写单元测试

- [ ] `components/navigation/sidebar.go` (新建)
  - [ ] 实现 Sidebar 侧边栏
  - [ ] 支持折叠/展开
  - [ ] 编写单元测试

- [ ] `components/navigation/public.go`
  - [ ] 导出公开接口

### 3.8 覆盖层组件 (components/overlay/)

- [ ] `ui/modal.go` → `components/overlay/modal.go`
  - [ ] 重构为 Modal 组件
  - [ ] 支持遮罩层
  - [ ] 编写单元测试

- [ ] `ui/tooltip.go` → `components/overlay/tooltip.go`
  - [ ] 重构为 Tooltip 组件
  - [ ] 支持自动定位
  - [ ] 编写单元测试

- [ ] `components/overlay/dialog.go` (新建)
  - [ ] 实现 Dialog 组件
  - [ ] 编写单元测试

- [ ] `components/overlay/dropdown.go` (新建)
  - [ ] 实现 Dropdown 组件
  - [ ] 编写单元测试

- [ ] `components/overlay/public.go`
  - [ ] 导出公开接口

### 3.9 容器组件 (components/container/)

- [ ] `components/container/panel.go` (新建)
  - [ ] 实现 Panel 面板组件
  - [ ] 支持标题/操作栏
  - [ ] 编写单元测试

- [ ] `components/container/split.go` (新建)
  - [ ] 实现 SplitPane 分割面板
  - [ ] 支持拖拽调整
  - [ ] 编写单元测试

- [ ] `components/container/scroll.go` (新建)
  - [ ] 实现 ScrollArea 滚动区域
  - [ ] 支持自动滚动条
  - [ ] 编写单元测试

- [ ] `components/container/public.go`
  - [ ] 导出公开接口

---

## Phase 4: 渲染系统重构

### 4.1 RNode 系统

- [ ] `internal/render/rnode.go`
  - [ ] 定义 `RNode` 结构
    - [ ] VNode 引用
    - [ ] 子节点列表
    - [ ] 布局信息 (x, y, width, height)
    - [ ] 绘制命令列表

  - [ ] `NewRNode(vnode VNode) *RNode`
  - [ ] `AddChild(child *RNode)`
  - [ ] `SetBounds(x, y, w, h int)`
  - [ ] `GetBounds() Rect`
  - [ ] 编写单元测试

### 4.2 VNode → RNode 转换

- [ ] `internal/render/converter.go`
  - [ ] `ConvertVNode(vnode VNode) *RNode`
  - [ ] 递归转换子节点
  - [ ] 保留 VNode 引用
  - [ ] 编写单元测试

### 4.3 Layout Engine

- [ ] `internal/render/layout_engine.go`
  - [ ] 定义 `LayoutEngine` 结构
  - [ ] `Layout(root *RNode, constraints Constraints) error`
    - [ ] 递归测量 (Measure Pass)
    - [ ] 递归布局 (Layout Pass)
  - [ ] 编写单元测试

- [ ] `internal/render/constraints.go`
  - [ ] 定义 `Constraints` 类型
  - [ ] `NewConstraints(minW, maxW, minH, maxH int)`
  - [ ] `Unbounded()` 常量
  - [ ] `Tight()` 构造函数

- [ ] `internal/render/size.go`
  - [ ] 定义 `Size` 类型
  - [ ] `Zero()` 常量
  - [ ] `Infinite()` 常量

### 4.4 Render Tree

- [ ] `internal/render/render_tree.go`
  - [ ] 定义 `RenderTree` 结构
  - [ ] `Build(root *RNode) *RenderTree`
  - [ ] `CollectDrawCmds(node *RNode) []DrawCmd`
  - [ ] 编写单元测试

- [ ] `internal/render/draw_cmd.go`
  - [ ] 定义 `DrawCmd` 接口
  - [ ] 定义 `CmdType` 枚举
  - [ ] 实现 `TextCmd`
  - [ ] 实现 `FillCmd`
  - [ ] 实现 `BoxCmd`
  - [ ] 实现 `CustomCmd`
  - [ ] 编写单元测试

### 4.5 Rasterizer

- [ ] `internal/render/rasterizer.go`
  - [ ] 定义 `Rasterizer` 结构
  - [ ] `Rasterize(cmds []DrawCmd, buffer *paint.Buffer) error`
  - [ ] 优化：批量相同类型命令
  - [ ] 优化：裁剪区域处理
  - [ ] 编写单元测试

---

## Phase 5: 多组件支持

### 5.1 DeclarativeNode

- [ ] `internal/render/declarative_node.go`
  - [ ] 定义 `DeclarativeNode` 结构
    - [ ] 嵌入 `component.Component`
    - [ ] `componentFn ComponentFunc`
    - [ ] `ctx *ui.ComponentContext`
    - [ ] `reconciler *reconciler.Reconciler`
    - [ ] `instanceMgr *state.InstanceManager`
    - [ ] `bounds paint.Rect`

  - [ ] 实现 `Component` 接口
    - [ ] `ID() string`
    - [ ] `Type() string`
    - [ ] `Mount(ctx Context) error`
    - [ ] `Update(ctx Context) error`
    - [ ] `Unmount(ctx Context) error`
    - [ ] `Measure(constraints Constraints) Size`
    - [ ] `Paint(ctx PaintContext)`

  - [ ] `RenderTo(x, y int, buffer *paint.Buffer)`
  - [ ] 编写单元测试

### 5.2 DeclarativeNode 工厂

- [ ] `ui/declarative_node.go`
  - [ ] `NewDeclarativeNode(fn ComponentFunc) *DeclarativeNode`
  - [ ] `Declarative(component ComponentFunc) VNode`
  - [ ] 编写单元测试

### 5.3 混合渲染支持

- [ ] `ui/app.go` 更新
  - [ ] 支持 Component 和 VNode 混合
  - [ ] 更新 `renderVNode()` 处理 DeclarativeNode
  - [ ] 编写集成测试

---

## Phase 6: API 入口层

### 6.1 shortcuts.go

- [ ] `ui/shortcuts.go`
  - [ ] 基础组件快捷函数
    - [ ] `Text(content string) VNode`
    - [ ] `Icon(name string) VNode`
    - [ ] `Separator() VNode`
    - [ ] `Spacer() VNode`

  - [ ] 布局组件快捷函数
    - [ ] `HStack(children ...VNode) VNode`
    - [ ] `VStack(children ...VNode) VNode`
    - [ ] `ZStack(children ...VNode) VNode`
    - [ ] `Box() *BoxBuilder`

  - [ ] 表单组件快捷函数
    - [ ] `Input(placeholder string) *InputBuilder`
    - [ ] `TextArea(placeholder string) *TextAreaBuilder`
    - [ ] `Checkbox(label string) *CheckboxBuilder`
    - [ ] `Select(options []string) *SelectBuilder`

  - [ ] 按钮组件快捷函数
    - [ ] `Button(label string) *ButtonBuilder`
    - [ ] `IconButton(icon string) *IconButtonBuilder`

  - [ ] 反馈组件快捷函数
    - [ ] `Progress(value float64) *ProgressBuilder`
    - [ ] `Spinner() *SpinnerBuilder`

### 6.2 app.go 精简

- [ ] `ui/app.go`
  - [ ] 移除已迁移的组件代码
  - [ ] 保留 `Run()` 入口
  - [ ] 保留 `declarativeRoot` (向后兼容)
  - [ ] 添加 `NewDeclarativeNode()` 支持
  - [ ] 清理不必要的代码

### 6.3 hooks.go

- [ ] `ui/hooks.go`
  - [ ] 验证 Hooks 仍正常工作
  - [ ] 更新导入路径
  - [ ] 编写单元测试

---

## Phase 7: 测试与验证

### 7.1 单元测试

- [ ] `internal/reconciler/` 测试覆盖率 > 80%
- [ ] `internal/scheduler/` 测试覆盖率 > 80%
- [ ] `internal/state/` 测试覆盖率 > 80%
- [ ] `internal/render/` 测试覆盖率 > 80%
- [ ] `components/*/` 测试覆盖率 > 70%

### 7.2 集成测试

- [ ] 创建 `tests/integration/` 目录
- [ ] `integration/render_test.go`
  - [ ] 测试完整渲染流程
  - [ ] 测试 VNode → RNode → DrawCmd → Buffer

- [ ] `integration/multi_component_test.go`
  - [ ] 测试多组件渲染
  - [ ] 测试声明式/imperative 混合

- [ ] `integration/hooks_test.go`
  - [ ] 测试所有 Hooks 功能
  - [ ] 测试 Hooks 与新架构兼容性

### 7.3 回归测试

- [ ] 运行所有现有示例
  - [ ] `examples/sandbox/demo/`
  - [ ] `examples/counter/`
  - [ ] `examples/todomvc/`
  - [ ] 确保行为一致

- [ ] 视觉回归测试
  - [ ] 对比重构前后输出
  - [ ] 记录任何差异

### 7.4 性能测试

- [ ] `benchmark/render_bench.go`
  - [ ] 渲染性能基准测试
  - [ ] 对比重构前后性能

- [ ] `benchmark/memory_bench.go`
  - [ ] 内存分配基准测试
  - [ ] 检测内存泄漏

### 7.5 示例更新

- [ ] 更新 `examples/` 中的导入路径
- [ ] 添加新架构示例
  - [ ] 混合使用示例
  - [ ] 直接导入组件示例
  - [ ] 多组件示例

---

## Phase 8: 文档更新

### 8.1 API 文档

- [ ] 更新 `README.md`
- [ ] 更新 `docs/API.md`
- [ ] 为每个组件包添加 `README.md`

### 8.2 设计文档

- [ ] 更新 `docs/ARCHITECTURE.md`
- [ ] 更新 `docs/RENDERING.md`
- [ ] 更新 `docs/COMPONENTS.md`

### 8.3 迁移指南

- [ ] 完成组件迁移指南
- [ ] 添加 API 迁移对照表
- [ ] 添加常见问题解答

### 8.4 示例文档

- [ ] 更新所有示例的 README
- [ ] 添加代码注释

---

## 检查清单

### 每个迁移项完成后检查

- [ ] 文件已复制到目标位置
- [ ] package 声明已更新
- [ ] 导入路径已更新
- [ ] 组件功能保持一致
- [ ] Builder 模式正常工作
- [ ] 单元测试已编写并通过
- [ ] 集成测试已通过
- [ ] 示例程序正常运行
- [ ] 文档已同步更新

### 发布前检查

- [ ] `go build ./...` 通过
- [ ] `go test ./... -cover` 覆盖率达标
- [ ] `go vet ./...` 无警告
- [ ] `golangci-lint run` 通过
- [ ] 所有示例正常运行
- [ ] 性能无明显退化
- [ ] 文档完整

---

## 优先级说明

| 优先级 | 说明 | 标记 |
|-------|------|------|
| P0 | 核心功能，必须完成 | 🔴 |
| P1 | 重要功能，尽快完成 | 🟡 |
| P2 | 增强功能，可延后 | 🟢 |

---

## 进度跟踪

```
Phase 0: 准备阶段           [████████████████████████████] 100%
Phase 1: 基础架构重组       [████████████████████████████] 100%
Phase 2: 内部模块迁移       [..................................] 0%
Phase 3: 组件库迁移         [████████████████████░░░░░░░░] 70%
Phase 4: 渲染系统重构       [..................................] 0%
Phase 5: 多组件支持         [..................................] 0%
Phase 6: API 入口层         [..................................] 0%
Phase 7: 测试与验证         [████████████████████████████] 100%
Phase 8: 文档更新           [..................................] 0%
```

---

**文档版本**: v1.1
**最后更新**: 2026-02-01
**预计总工期**: 7-8 周

---

## 已完成组件清单

### Phase 3: 组件库迁移 (已完成)

| 组件分类 | 组件名 | 状态 |
|---------|--------|------|
| basic | Text | ✅ |
| basic | Divider | ✅ |
| layout | HStack, VStack, Box, Spacer | ✅ |
| form | Input | ✅ |
| form | TextArea | ✅ |
| form | Checkbox | ✅ |
| form | Select | ✅ |
| button | Button | ✅ |
| feedback | Progress | ✅ |
| feedback | Spinner | ✅ |
| data | Table | ✅ |
| data | VirtualList | ✅ |
| navigation | Tabs | ✅ |
| overlay | Modal | ✅ |

### 待迁移组件

| 组件分类 | 组件名 | 状态 |
|---------|--------|------|
| basic | Icon, Spacer | ⏳ |
| layout | Absolute, Grid, Flex | ⏳ |
| form | Switch, Slider, Field | ⏳ |
| button | IconButton, ButtonGroup | ⏳ |
| feedback | Toast, Alert, Badge | ⏳ |
| overlay | Tooltip, Dialog, Dropdown | ⏳ |
| container | Panel, SplitPane, ScrollArea | ⏳ |
| navigation | Menu, Sidebar | ⏳ |
