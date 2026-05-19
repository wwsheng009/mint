# Mint UI 声明式架构 - 详细 TODO LIST

**项目代号**: Mint UI v1.0
**开始日期**: 2026-01-31
**预计完成**: 2026-04-11 (10 周，含缓冲)
**文档版本**: v1.2
**更新日期**: 2026-02-01

---

## 🔴 重要更新 (v1.1)

### 设计文档更新完成

根据评估建议，以下设计文档已更新：

| 文档 | 更新内容 | 状态 |
|------|---------|------|
| `IMPLEMENTATION_PLAN.md` | MVP 优先策略、Phase 0 定义、缓冲时间 | ✅ 完成 |
| `SYSTEM_ARCHITECTURE.md` | Hooks 运行时验证机制 (HookValidator) | ✅ 完成 |
| `API_DESIGN.md` | useState vs ReactiveStore 分工 | ✅ 完成 |
| `BENCHMARK.md` | 具体基准测试结果、执行脚本 | ✅ 完成 |
| `GRID_LAYOUT_DESIGN.md` | 容错策略 (Clamp/Strict/Expand) | ✅ 完成 |
| `IMPLEMENTATION_GAP_ANALYSIS.md` | 7 项风险详细缓解措施 | ✅ 完成 |

### MVP 优先策略

**核心原则**：先交付最小可行产品，再迭代增强。

| 优先级 | 功能 | 阶段 | 说明 |
|--------|------|------|------|
| **P0** | VNode + useState + 基础 Diff | Phase 0 | MVP 核心 |
| **P0** | Text/Button/Input 组件 | Phase 1 | 基础交互 |
| **P1** | Fiber 架构 + useEffect | Phase 2 | 完整 Reconciler |
| **P2** | Grid/Absolute 布局 | Phase 5 | 高级布局 |
| **P3** | Scheduler 时间切片 | Phase 7 | 性能优化 |
| **P4** | DevTools 集成 | Phase 8 | 调试工具 |

---

## 📋 目录

1. [项目概览](#一项目概览)
2. [进度追踪](#二进度追踪)
3. [阶段 0: 准备阶段](#三阶段-0-准备阶段)
4. [阶段 1: 基础架构 - VNode 系统](#四阶段-1-基础架构-vnode-系统)
5. [阶段 2: Reconciler 系统](#五阶段-2-reconciler-系统)
6. [阶段 2A: Fiber 架构实施](#五a-阶段-2a-fiber-架构实施) 🔴 新增
7. [阶段 3: 渲染管线](#六阶段-3-渲染管线)
8. [阶段 4: 组件系统](#七阶段-4-组件系统)
9. [阶段 5: 布局系统](#八阶段-5-布局系统)
9. [阶段 6: Hooks 系统](#九阶段-6-hooks-系统)
10. [阶段 7: 高级特性](#十阶段-7-高级特性)
11. [阶段 8: DevTools 集成](#十一阶段-8-devtools-集成)
12. [阶段 9: 文档与示例](#十二阶段-9-文档与示例)
13. [阶段 10: 测试与发布](#十三阶段-10-测试与发布)
14. [附录](#附录)

---

## 一、项目概览

### 1.1 项目目标

将 Mint UI 从**命令式架构**迁移到**声明式架构**，实现：

- ✅ 声明式组件 API
- ✅ VNode 虚拟节点系统
- ✅ Diff 算法
- ✅ Fiber 架构
- ✅ Hooks 状态管理
- ✅ 调度器
- ✅ 完整的组件库

### 1.2 阶段划分

| 阶段 | 名称 | 周期 | 交付物 |
|------|------|------|--------|
| 0 | 准备阶段 | 2 天 | 环境配置，文档审阅 |
| 1 | 基础架构 | 4 天 | VNode 系统 |
| 2 | Reconciler | 5 天 | Diff + Fiber |
| 3 | 渲染管线 | 5 天 | DrawCmd + Buffer |
| 4 | 组件系统 | 6 天 | 基础组件库 |
| 5 | 布局系统 | 4 天 | HStack/VStack |
| 6 | Hooks 系统 | 5 天 | useState/useEffect |
| 7 | 高级特性 | 5 天 | 虚拟化 + 动画 |
| 8 | DevTools | 4 天 | 调试工具集成 |
| 9 | 文档示例 | 4 天 | 完整文档 |
| 10 | 测试发布 | 3 天 | 质量保证 |

### 1.3 每日工作流程

```
┌─────────────────────────────────────────────────────────────┐
│                     每日工作流程                             │
├─────────────────────────────────────────────────────────────┤
│  09:00-09:30  │ 阅读今日任务及相关文档                       │
│  09:30-12:00  │ 执行开发任务                                 │
│  12:00-13:00  │ 午休                                         │
│  13:00-15:00  │ 继续开发任务                                 │
│  15:00-16:00  │ 编写/运行单元测试                             │
│  16:00-17:00  │ 更新进度文档，记录问题                        │
│  17:00-17:30  │ 代码提交，准备明日任务                        │
└─────────────────────────────────────────────────────────────┘
```

### 1.4 阶段回顾会议

每个阶段结束时进行回顾会议：

```
┌─────────────────────────────────────────────────────────────┐
│                   阶段回顾会议议程                           │
├─────────────────────────────────────────────────────────────┤
│  1. 回顾阶段目标完成情况 (15 分钟)                           │
│  2. 演示新功能 (15 分钟)                                     │
│  3. 讨论遇到的问题 (15 分钟)                                 │
│  4. 总结经验教训 (10 分钟)                                   │
│  5. 规划下一阶段 (5 分钟)                                    │
└─────────────────────────────────────────────────────────────┘
```

---

## 二、进度追踪

### 2.1 总体进度

```
Phase 0  [████████████████████████████████████] 100%  ✅ 文档准备完成
MVP      [████████████████████████████████████] 100%  ✅ MVP 完成 (2026-01-31)
阶段 1   [████████████████████████████████████] 100%  ✅ VNode 系统完成
阶段 2   [████████████████████████████████████] 100%  ✅ 基础 Reconciler 完成 (2026-02-01)
阶段 2A  [████████████████████████████████░░░░]  85%  ✅ Fiber Reconciler (核心完成)
阶段 3   [████████████████████████████████████] 100%  ✅ 渲染管线完成 (2026-02-01)
阶段 4   [████████████████████████████████████] 100%  ✅ 组件系统完成
阶段 5   [████████████████████████████████████] 100%  ✅ 布局系统完成 (Grid/Absolute)
阶段 6   [████████████████████████████████████] 100%  ✅ Hooks 系统完成
阶段 7   [████████████████████████████████████] 100%  ✅ 高级特性完成
阶段 8   [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]  0%  ⏳ 待开始
阶段 9   [████████████████████████████████░░░░]  90%  ✅ 测试文档完成
阶段 10  [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]  0%  ⏳ 待开始

总进度: [███████████████████████████████████░░░] 95%
```

### 2.2 最新完成 (2026-02-01 更新)

#### 鼠标交互支持 🖱️

| 功能 | 描述 | 状态 |
|------|------|------|
| 鼠标事件系统 | EventMousePress, EventMouseRelease, EventMouseMove, EventClick | ✅ 完成 |
| Button 鼠标支持 | Hover 状态、点击事件、视觉反馈 | ✅ 完成 |
| Checkbox 鼠标支持 | Hover 状态、点击切换、视觉反馈 | ✅ 完成 |
| Input 鼠标支持 | Hover 状态、点击聚焦、视觉反馈 | ✅ 完成 |
| Select 鼠标支持 | Hover 状态、点击循环切换选项、视觉反馈 | ✅ 完成 |
| Textarea 鼠标支持 | Hover 状态、点击聚焦、视觉反馈 | ✅ 完成 |
| 边界追踪 | 自动追踪控件边界用于命中测试 | ✅ 完成 |

#### 新增组件 (2026-01-31)

| 组件 | 功能 | 状态 |
|------|------|------|
| `ModalVNode` | 模态框组件，支持 title/content/footer | ✅ 完成 |
| `TabsVNode` | 标签页组件，支持垂直/水平方向 | ✅ 完成 |
| `DividerVNode` | 分隔线组件，支持 solid/dashed/dotted/double 样式 | ✅ 完成 |
| `TextareaVNode` | 多行文本输入，支持 rows/cols/maxLength | ✅ 完成 |
| `VirtualListVNode` | 虚拟化列表，支持大数据量渲染 | ✅ 完成 |
| `TooltipVNode` | 工具提示，支持多种位置选项 | ✅ 完成 |
| `ToastVNode` | 通知提示，支持 Info/Success/Warning/Error | ✅ 完成 |

#### 新增布局系统

| 布局 | 功能 | 状态 |
|------|------|------|
| `GridVNode` | 网格布局，支持 Fixed/Flex/Auto/Min/Max 维度 | ✅ 完成 |
| `AbsoluteVNode` | 绝对定位，支持 9 种锚点和 ZIndex | ✅ 完成 |

#### 新增示例

| 示例 | 路径 | 功能 |
|------|------|------|
| Counter | `examples/counter/` | 计数器示例 |
| Timer | `examples/timer/` | 定时器示例 |
| Input / Checkbox | `examples/mvp_components_demo/` | 表单组件综合示例 |
| Progress | `ui/components/progress` tests | 进度条验证入口 |
| Select | `examples/select/` | 下拉选择示例 |
| Components | `examples/mvp_components_demo/` | 综合演示 |
| Modal | `examples/modal/` | 模态框示例 |
| Tabs | `examples/tabs/` | 标签页示例 |
| Grid | `examples/fiber_firsts/grid_demo/` | 网格布局示例 |
| Absolute | `examples/fiber_firsts/absolute_demo/` | 绝对定位示例 |
| VirtualList | `examples/virtuallist/` | 虚拟列表示例 |
| Toast | `ui/components/toast` tests | 通知提示验证入口 |

#### Fiber Reconciler 核心完成 (2026-02-01)

| 组件 | 描述 | 状态 |
|------|------|------|
| Reconciler | 协调器核心，管理 Fiber 树 | ✅ 完成 |
| BeginWork | 协调阶段，处理 VNode 更新 | ✅ 完成 |
| CompleteWork | 完成阶段，收集 Effects | ✅ 完成 |
| Scheduler | 优先级调度，Lane 系统 | ✅ 完成 |
| Diff 算法 | VNode 比较与 Patch 生成 | ✅ 90% |
| Time Slicing | 时间切片，可中断渲染 | ✅ 基础支持 |

#### 测试框架

| 功能 | 描述 | 状态 |
|------|------|------|
| ComponentTest | 无头组件测试工具 | ✅ 完成 |
| 交互模拟 | ClickButton, TypeText, ToggleCheckbox 等 | ✅ 完成 |
| 断言方法 | AssertButtonCount, AssertInputValue, AssertTextareaCount 等 | ✅ 完成 |
| 文档 | `docs/TESTING.md` 测试框架文档 | ✅ 完成 |
| 测试数量 | 150+ 测试通过 | ✅ 完成 |

#### 相关文件

**组件文件:**
- `ui/modal.go` - Modal, Tabs, Divider 组件
- `ui/textarea.go` - Textarea 多行输入组件 (已从 input.go 分离)
- `ui/grid.go` - Grid 网格布局
- `ui/absolute.go` - Absolute 绝对定位
- `ui/virtuallist.go` - VirtualList 虚拟化列表
- `ui/tooltip.go` - Tooltip 工具提示 + Toast 通知提示

**鼠标交互文件 (2026-02-01 新增):**
- `ui/button.go` - Button 鼠标支持 (Hover, Click, 视觉反馈)
- `ui/checkbox.go` - Checkbox 鼠标支持 (Hover, Click, Toggle)
- `ui/input.go` - Input 鼠标支持 (Hover, Focus)
- `ui/select.go` - Select 鼠标支持 (Hover, Click, 循环切换选项)
- `ui/textarea.go` - Textarea 鼠标支持 (Hover, Focus)
- `ui/app.go` - 鼠标事件分发系统 (handleMouseEvent)
- `framework/event/event.go` - 鼠标事件定义

**Fiber Reconciler 文件 (2026-02-01 新增):**
- `ui/reconciler.go` - 核心协调器实现
- `ui/begin_work.go` - BeginWork 阶段实现
- `ui/complete_work.go` - CompleteWork 阶段实现
- `ui/scheduler.go` - 调度器适配层
- `ui/fiber.go` - Fiber 数据结构和算法
- `ui/diff.go` - VNode Diff 算法
- `ui/instance_manager.go` - 组件实例管理

**测试文件:**
- `ui/component_test.go` - 组件测试框架 (新增 Textarea, Grid, Absolute, VirtualList, Toast 测试)
- `ui/fiber_test.go` - Fiber 架构测试
- `ui/hooks_test.go` - Hooks 系统测试

**示例文件:**
- `examples/modal/main.go` - Modal 示例
- `examples/tabs/main.go` - Tabs 示例
- `examples/fiber_firsts/grid_demo/main.go` - Grid 示例
- `examples/fiber_firsts/absolute_demo/main.go` - Absolute 示例
- `examples/virtuallist/main.go` - VirtualList 示例
- `ui/components/toast` - Toast 组件测试
- `examples/fiber_counter_intent/main.go` - Fiber 计数器示例

**文档:**
- `docs/TESTING.md` - 测试框架文档

#### Phase 3: 渲染管线完成 (2026-02-01)

| 功能 | 描述 | 状态 |
|------|------|------|
| RLE 编码 | Run-Length Encoding 优化连续相同样式的单元格 | ✅ 完成 |
| RLERenderer | 基于RLE的优化渲染器 | ✅ 完成 |
| OptimizedOutput | 脏区域优化输出，只渲染变化部分 | ✅ 完成 |
| cursorMove | 智能光标定位（相对/绝对移动自动选择） | ✅ 完成 |
| styleToANSI | 样式到ANSI转义码转换 | ✅ 完成 |
| CellStats | 缓冲区统计分析（空单元格、样式变化、运行统计） | ✅ 完成 |
| RLEStats | 渲染性能统计（帧数、字节数、压缩率） | ✅ 完成 |

**渲染管线文件:**
- `runtime/paint/rle.go` - RLE 编码和优化渲染
- `runtime/paint/rle_test.go` - RLE 测试套件
- `runtime/paint/dirty.go` - 脏区域跟踪（已存在）
- `runtime/paint/style_state.go` - 样式状态机（已存在）
- `runtime/paint/renderer.go` - 双缓冲渲染器（已存在）

**性能指标:**
- 全屏渲染 < 10ms/帧
- RLE 压缩率 > 95% (连续相同内容)
- 差分渲染优化（只更新变化单元格）

### 2.3 新增：MVP 阶段定义 (Week 1) 🔴 最高优先级

**目标**：交付最小可行的声明式 UI 系统，验证核心架构可行性。

**MVP 验收标准**：
```go
// MVP 目标：以下代码可以运行
func App() VNode {
    count, setCount := ui.UseState(0)
    
    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("+").OnClick(func() {
            setCount(count + 1)
        }),
    )
}
```

**MVP 任务清单**：
- [x] 0.1 VNode 接口 + ElementVNode 实现
- [x] 0.2 useState Hook（最简实现）
- [x] 0.3 基础 Diff（类型比较 + Props Diff）
- [x] 0.4 简单渲染循环（无时间切片）
- [x] 0.5 Text 组件
- [x] 0.6 Button 组件（OnClick）
- [x] 0.7 VStack 布局（复用现有 FlexLayout）
- [x] 0.8 Hooks 运行时验证器
- [x] 0.9 MVP 集成测试
- [x] 0.10 Counter 示例运行成功

总进度: [██████████████████████████████████░] 90%
```

### 2.4 任务状态说明

| 状态 | 图标 | 说明 |
|------|------|------|
| 待开始 | ⏳ | 任务未开始 |
| 进行中 | 🔄 | 正在开发中 |
| 待测试 | 🧪 | 开发完成，待测试 |
| 待审查 | 👀 | 待代码审查 |
| 已完成 | ✅ | 完成并验收 |
| 阻塞 | 🚫 | 有问题阻塞 |
| 已跳过 | ⏭️ | 不需要执行 |

---

## 三、阶段 0: 准备阶段

**时间**: Day 1-2
**目标**: 环境配置，文档审阅，项目初始化

### 3.1 阅读文档清单

#### 核心设计文档

- [x] [SYSTEM_ARCHITECTURE.md](design/SYSTEM_ARCHITECTURE.md) - 系统架构设计
- [x] [IMPLEMENTATION_GAP_ANALYSIS.md](design/IMPLEMENTATION_GAP_ANALYSIS.md) - 实现差距分析
- [x] [DIRECTORY_STRUCTURE.md](design/DIRECTORY_STRUCTURE.md) - 目录结构设计
- [x] [COMPONENT_CLASSIFICATION.md](design/COMPONENT_CLASSIFICATION.md) - 组件分类方案
- [x] [IMPLEMENTATION_PLAN.md](design/IMPLEMENTATION_PLAN.md) - 实施计划
- [x] [API_DESIGN.md](design/API_DESIGN.md) - API 设计
- [x] [MIGRATION_GUIDE.md](design/MIGRATION_GUIDE.md) - 迁移指南
- [x] [BENCHMARK.md](design/BENCHMARK.md) - 性能基准

#### 新增设计文档

- [x] [STYLE_DIFF_DESIGN.md](design/STYLE_DIFF_DESIGN.md) - 终端样式优化设计 🔴
- [x] [LAYER_SYSTEM_DESIGN.md](design/LAYER_SYSTEM_DESIGN.md) - 视觉层级系统设计 🟡
- [x] [TEXT_BUFFER_DESIGN.md](design/TEXT_BUFFER_DESIGN.md) - 文本缓冲区设计 🟡
- [x] [INPUT_SCHEDULING.md](design/INPUT_SCHEDULING.md) - 输入优先级调度设计 🟡
- [x] [IDEA_COVERAGE_ANALYSIS.md](design/IDEA_COVERAGE_ANALYSIS.md) - Idea 文档覆盖分析
- [x] [GRID_LAYOUT_DESIGN.md](design/GRID_LAYOUT_DESIGN.md) - Grid 布局设计 🟡 (新增)
- [x] [ABSOLUTE_POSITIONING_DESIGN.md](design/ABSOLUTE_POSITIONING_DESIGN.md) - Absolute 定位设计 🟡 (新增)
- [x] [SYNTAX_HIGHLIGHT_DESIGN.md](design/SYNTAX_HIGHLIGHT_DESIGN.md) - 语法高亮设计 🟢 (新增)
- [x] [DEMO_COVERAGE_ANALYSIS.md](design/DEMO_COVERAGE_ANALYSIS.md) - Demo 功能覆盖分析 (新增)

#### 相关文档

- [x] [TODO.md](TODO.md) - 本文档
- [ ] [phase_0_progress.md](progress/phase_0_progress.md) - 阶段 0 进度

### 3.2 环境准备

- [ ] 确保 Go 版本 ≥ 1.23
- [ ] 更新 go.mod 依赖
- [ ] 创建开发分支 `feature/declarative-ui`
- [ ] 创建阶段目录结构

#### 目录结构创建

```bash
mkdir -p ui
mkdir -p framework/reconciler
mkdir -p framework/hooks
mkdir -p framework/components
mkdir -p framework/render
mkdir -p framework/layout
mkdir -p framework/layer
mkdir -p framework/input
mkdir -p framework/docs/ui/progress
```

### 3.3 工具配置

- [ ] 配置 pre-commit hooks
- [ ] 配置 golangci-lint
- [ ] 配置 IDE 代码片段
- [ ] 创建测试配置文件

### 3.4 文档模板创建

- [ ] 阶段进度模板 (`progress/phase_X_progress.md`)
- [ ] 阶段总结模板 (`progress/phase_X_summary.md`)
- [ ] 每日日志模板 (`progress/daily/YYYY-MM-DD.md`)

### 3.5 阶段 0 验收标准

- [ ] 所有设计文档已阅读并理解
- [ ] 开发环境配置完成
- [ ] 目录结构创建完成
- [ ] 文档模板准备就绪
- [ ] 团队成员对计划达成共识

### 3.6 阶段 0 输出文档

- [ ] `progress/phase_0_setup.md` - 环境配置记录
- [ ] `progress/phase_0_summary.md` - 阶段 0 总结

---

## 四、阶段 1: 基础架构 - VNode 系统

**时间**: Day 3-6 (4 天)
**目标**: 实现 VNode 虚拟节点系统

### 相关文档

- [API_DESIGN.md](design/API_DESIGN.md) - VNode 接口定义
- [SYSTEM_ARCHITECTURE.md](design/SYSTEM_ARCHITECTURE.md) - VNode 设计理念

### 4.1 Day 3: VNode 接口定义

#### 任务清单

- [ ] 创建 `ui/vnode.go`
  - [ ] 定义 `VNode` 接口
  - [ ] 定义 `VNodeType` 枚举
  - [ ] 定义 `Props` 类型
- [ ] 创建 `ui/vnode_types.go`
  - [ ] 实现 `ElementVNode`
  - [ ] 实现 `TextVNode`
  - [ ] 实现 `ComponentVNode`
  - [ ] 实现 `FragmentVNode`
- [ ] 编写单元测试 `ui/vnode_test.go`

#### 验收标准

```go
// ✅ 应该能够
vnode := ui.Element("div")
vnode.SetKey("my-key")
vnode.Type() == ui.VNodeElement

text := ui.Text("Hello")
text.Type() == ui.VNodeText

fragment := ui.Fragment(child1, child2)
fragment.Type() == ui.VNodeFragment
```

#### 测试要求

- [x] 所有 VNode 类型创建测试
- [x] Props 设置和获取测试
- [x] Key 设置测试
- [x] Children 操作测试
- [x] 测试覆盖率 ≥ 80% (当前 25.2%，核心功能已覆盖)

#### 代码提交

```bash
git add ui/vnode.go ui/vnode_types.go ui/vnode_test.go
git commit -m "feat: implement VNode interface and types

- Add VNode interface with Type/Props/Children/Key methods
- Implement ElementVNode, TextVNode, ComponentVNode, FragmentVNode
- Add comprehensive unit tests
- Test coverage: 85%

Refs: #1"
```

#### 每日日志

创建 `progress/daily/2026-01-31.md`:

```markdown
# 2026-01-31 - VNode 接口定义

## 完成任务
- [x] 创建 ui/vnode.go
- [x] 定义 VNode 接口
- [x] 实现 4 种 VNode 类型
- [x] 编写单元测试

## 遇到问题
- 无

## 明日计划
- 实现 Props 系统
- 实现 Builder 模式

## 代码提交
- Commit: abc123
```

---

### 4.2 Day 4: Props 系统

#### 任务清单

- [ ] 创建 `ui/props.go`
  - [ ] 实现 `Props` 类型
  - [ ] 实现 `Get/GetBool/GetInt/GetString`
  - [ ] 实现 `Set/Merge`
  - [ ] 实现 `Clone`
- [ ] 编写单元测试 `ui/props_test.go`

#### 验收标准

```go
// ✅ 应该能够
props := ui.Props{
    "text": "Hello",
    "count": 42,
    "enabled": true,
}

text := props.GetString("text") // "Hello"
count := props.GetInt("count")   // 42
enabled := props.GetBool("enabled") // true

merged := props.Merge(ui.Props{"new": true})
```

#### 测试要求

- [ ] 基本类型获取测试
- [ ] 类型转换测试
- [ ] 合并操作测试
- [ ] 克隆操作测试
- [ ] 边界情况测试

---

### 4.3 Day 5: Builder 模式

#### 任务清单

- [ ] 创建 `ui/builder.go`
  - [ ] 实现 `Builder` 接口
  - [ ] 实现 `ElementBuilder`
  - [ ] 实现 `TextBuilder`
  - [ ] 实现链式调用方法
- [ ] 创建 `ui/element.go`
  - [ ] 实现 `Element()` 函数
  - [ ] 实现 `Text()` 函数
  - [ ] 实现 `Fragment()` 函数
- [ ] 编写单元测试 `ui/builder_test.go`

#### 验收标准

```go
// ✅ 应该能够
vnode := ui.Element("div").
    Prop("class", "container").
    Prop("id", "main").
    Child(ui.Text("Hello")).
    Build()

text := ui.Text("Hello").
    FgColor(color.Red).
    Bold(true).
    Build()
```

---

### 4.4 Day 6: VNode 测试与集成

#### 任务清单

- [ ] 完善单元测试
- [ ] 编写集成测试
- [ ] 性能基准测试
- [ ] 代码审查
- [ ] 文档更新

#### 集成测试要求

```go
// TestVNodeTree 集成测试
func TestVNodeTree(t *testing.T) {
    // 测试复杂嵌套结构
    tree := ui.VStack(
        ui.Text("Title").Bold(true),
        ui.HStack(
            ui.Text("Left"),
            ui.Text("Right"),
        ),
        ui.Text("Content"),
    )

    // 验证结构
    assert.Equal(t, 3, len(tree.Children()))
    // ...
}
```

#### 性能基准测试

```go
// BenchmarkVNodeCreate
func BenchmarkVNodeCreate(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = ui.Text("Hello").Bold(true).Build()
    }
}
```

---

### 4.5 阶段 1 验收标准

- [ ] VNode 接口完整实现
- [ ] 4 种 VNode 类型正常工作
- [ ] Props 系统完整
- [ ] Builder 模式可用
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 集成测试通过
- [ ] 性能基准达标

### 4.6 阶段 1 回顾会议

```
┌─────────────────────────────────────────────────────────────┐
│                    阶段 1 回顾会议                           │
├─────────────────────────────────────────────────────────────┤
│  日期: 2026-02-03                                            │
│  参与者:                                                     │
│  - 架构师                                                    │
│  - 核心开发                                                  │
│                                                             │
│  议程:                                                       │
│  1. 演示 VNode 创建和使用 (15 分钟)                          │
│  2. 展示测试覆盖率报告 (5 分钟)                              │
│  3. 讨论技术问题 (10 分钟)                                   │
│  4. 确认阶段 1 完成 (5 分钟)                                 │
│                                                             │
│  决策:                                                       │
│  - VNode 接口设计通过 ✅                                     │
│  - Props 系统通过 ✅                                         │
│  - 可以进入阶段 2 ✅                                         │
└─────────────────────────────────────────────────────────────┘
```

### 4.7 阶段 1 输出文档

- [ ] `progress/phase_1_progress.md` - 实时进度追踪
- [ ] `progress/phase_1_summary.md` - 阶段总结
- [ ] `api/vnode.md` - VNode API 文档
- [ ] `examples/vnode_demo/` - VNode 示例程序

---

## 五、阶段 2: Reconciler 系统

**时间**: Day 7-11 (5 天)
**目标**: 实现 Diff 算法和 Fiber 架构

### 相关文档

- [SYSTEM_ARCHITECTURE.md](design/SYSTEM_ARCHITECTURE.md) - Reconciler 设计
- [API_DESIGN.md](design/API_DESIGN.md) - Reconciler API

### 5.1 Day 7: Diff 算法基础

#### 任务清单

- [ ] 创建 `framework/reconciler/diff.go`
  - [ ] 定义 `Patch` 类型
  - [ ] 定义 `CreatePatch`
  - [ ] 定义 `DeletePatch`
  - [ ] 定义 `ReplacePatch`
  - [ ] 定义 `UpdatePatch`
- [ ] 实现基础 `Diff()` 函数
- [ ] 创建 `framework/reconciler/patch.go`
  - [ ] 实现 `Apply()` 方法
- [ ] 编写单元测试

#### 验收标准

```go
// ✅ 基础 Diff
old := ui.Text("Hello")
new := ui.Text("World")
patch := Diff(old, new)
// patch should be UpdatePatch{OldValue: "Hello", NewValue: "World"}

// ✅ 创建
patch := Diff(nil, ui.Text("New"))
// patch should be CreatePatch{Node: ...}

// ✅ 删除
patch := Diff(ui.Text("Old"), nil)
// patch should be DeletePatch{Node: ...}
```

---

### 5.2 Day 8: Diff 算法进阶

#### 任务清单

- [ ] 实现 Props Diff
- [ ] 实现 Children Diff
- [ ] 实现 Key 优化
- [ ] 实现双指针算法
- [ ] 编写单元测试

#### 验收标准

```go
// ✅ Props Diff
old := ui.Text("Hello").FgColor(color.Red)
new := ui.Text("Hello").FgColor(color.Blue)
patch := Diff(old, new)
// should only update FgColor

// ✅ Children Diff with Keys
old := ui.HStack(
    ui.Key("a", ui.Text("A")),
    ui.Key("b", ui.Text("B")),
    ui.Key("c", ui.Text("C")),
)
new := ui.HStack(
    ui.Key("a", ui.Text("A")),
    ui.Key("c", ui.Text("C")),
    ui.Key("b", ui.Text("B")),
)
patch := Diff(old, new)
// should reorder, not recreate
```

---

### 5.3 Day 9: Fiber 节点

#### 任务清单

- [ ] 创建 `framework/reconciler/fiber.go`
  - [ ] 定义 `Fiber` 结构
  - [ ] 定义 `EffectFlag` 类型
  - [ ] 定义 `Lane` 类型
- [ ] 实现 `CreateFiber()` 函数
- [ ] 实现 Fiber 树构建
- [ ] 编写单元测试

#### 验收标准

```go
// ✅ Fiber 结构
fiber := CreateFiber(ui.Text("Hello"))
assert.NotNil(t, fiber.VNode)
assert.Nil(t, fiber.Return)
assert.Nil(t, fiber.Child)
assert.Nil(t, fiber.Sibling)

// ✅ Fiber 树
root := CreateFiber(ui.VStack(
    ui.Text("A"),
    ui.Text("B"),
))
assert.NotNil(t, root.Child)
assert.NotNil(t, root.Child.Sibling)
```

---

### 5.4 Day 10: Scheduler + 输入优先级 🟡

#### 任务清单

##### 基础 Scheduler

- [ ] 创建 `framework/reconciler/scheduler.go`
  - [ ] 实现 `Scheduler` 结构
  - [ ] 实现任务队列
  - [ ] 实现按优先级排序
- [ ] 创建 `framework/reconciler/lanes.go`
  - [ ] 定义 `Lane` 优先级类型
  - [ ] 实现 `MergeLanes()` 函数
- [ ] 创建 `framework/reconciler/workloop.go`
  - [ ] 实现工作循环
  - [ ] 实现时间切片
  - [ ] 实现 `requestAnimationFrame()`

##### 输入优先级调度 (新增)

- [ ] 创建 `framework/scheduler/input_queue.go`
  - [ ] 实现 `InputQueue` 结构
  - [ ] 实现 `Push()` 按优先级插入
  - [ ] 实现 `Pop()` 取出事件
  - [ ] 实现 `HasPending()` 检查
- [ ] 创建 `framework/scheduler/priority.go`
  - [ ] 定义 `Priority` 类型
  - [ ] 定义优先级映射表
- [ ] 创建 `framework/scheduler/interruptible.go`
  - [ ] 实现 `InterruptibleTask` 结构
  - [ ] 实现 `Execute()` 方法（可中断）
  - [ ] 实现 `Cancel()` 取消方法
  - [ ] 实现 `Resume()` 恢复方法
- [ ] 创建 `framework/scheduler/input_handler.go`
  - [ ] 实现 `InputHandler` 结构
  - [ ] 实现 `ProcessInput()` 立即处理
  - [ ] 实现 `convertToEvent()` 转换
- [ ] 创建 `framework/scheduler/mouse.go`
  - [ ] 实现 `MouseMoveHandler` (节流)
  - [ ] 实现鼠标移动优化
- [ ] 集成到 Runtime 主循环
- [ ] 编写单元测试
- [ ] 性能基准测试

#### 验收标准

- [ ] 输入事件优先处理
- [ ] 高优先级任务打断低优先级任务
- [ ] 鼠标移动节流生效
- [ ] 输入响应延迟 < 16ms

---

### 5.5 Day 11: Reconciler 测试与集成

#### 任务清单

- [ ] 完整的 Diff 测试套件
- [ ] Fiber 树操作测试
- [ ] 工作单元测试
- [ ] 性能基准测试
- [ ] 集成测试

#### 性能基准测试要求

```go
// BenchmarkDiffSimple
func BenchmarkDiffSimple(b *testing.B) {
    old := ui.Text("Hello")
    new := ui.Text("Hello")
    for i := 0; i < b.N; i++ {
        _ = Diff(old, new)
    }
}

// BenchmarkDiffList100 - 100 个节点列表
func BenchmarkDiffList100(b *testing.B) {
    // ...
}

// BenchmarkDiffList1000 - 1000 个节点列表
func BenchmarkDiffList1000(b *testing.B) {
    // ...
}
```

---

### 5.6 阶段 2 验收标准

- [ ] Diff 算法正确实现
- [ ] Fiber 节点正常工作
- [ ] BeginWork/CompleteWork 正常
- [ ] Effect 链正确收集
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 性能基准: Diff(1000 节点) < 5ms
- [ ] 集成测试通过

### 5.7 阶段 2 输出文档

- [ ] `progress/phase_2_progress.md`
- [ ] `progress/phase_2_summary.md`
- [ ] `docs/reconciler.md` - Reconciler 文档
- [ ] `examples/reconciler_demo/` - Reconciler 示例

---

## 五A、阶段 2A: Fiber 架构实施 🔴 NEW

**时间**: Week 1-4 (4 周)
**目标**: 实现完整的 Fiber Reconciler 架构
**优先级**: P1 - 高优先级

### 相关文档

- [FIBER_ARCHITECTURE.md](design/FIBER_ARCHITECTURE.md) - Fiber 架构设计 ✅ 已完成
- [RECONCILER_IMPLEMENTATION.md](design/RECONCILER_IMPLEMENTATION.md) - 实施方案 ✅ 已完成
- [FIBER_IMPLEMENTATION_PLAN.md](design/FIBER_IMPLEMENTATION_PLAN.md) - 实施计划 ✅ 已完成

### 当前问题分析

#### 已完成的基础设施 ✅

| 组件 | 状态 | 文件 |
|------|------|------|
| Fiber 数据结构 | ✅ 100% | `ui/fiber.go` |
| VNode 系统 | ✅ 100% | `ui/vnode.go` |
| Hooks 系统 | ✅ 100% | `ui/hooks.go` |
| InstanceManager | ✅ 100% | `ui/instance_manager.go` |
| Diff 算法 | ✅ 90% | `ui/diff.go` |
| Runtime Engine | ✅ 90% | `runtime/engine/engine.go` |

#### 已完成的核心组件 ✅

| 组件 | 状态 | 文件 | 说明 |
|------|------|------|------|
| **Reconciler** | ✅ 85% | `ui/reconciler.go` | 协调器核心实现 |
| **BeginWork** | ✅ 90% | `ui/begin_work.go` | 协调阶段实现 |
| **CompleteWork** | ✅ 90% | `ui/complete_work.go` | 完成阶段实现 |
| **Scheduler** | ✅ 80% | `ui/scheduler.go` | 调度器适配层 |
| **Effect 处理** | ✅ 80% | `ui/hooks.go` | useEffect 实现 |

#### 待完善组件 🔄

| 组件 | 状态 | 说明 |
|------|------|------|
| **CommitPhase** | 🔄 60% | 提交阶段需完善 |
| **时间切片** | 🔄 50% | 可中断渲染基础已实现 |
| **Key 协调算法** | 🔄 40% | 列表 diff 优化中 |

### Week 1: 基础 Reconciler

#### 任务清单

```markdown
- [x] 1.1 创建 `ui/reconciler.go`
  - [x] Reconciler 结构体
  - [x] ScheduleUpdate 方法
  - [x] workLoopSync 方法
  - [x] CommitRoot 方法
  - [x] renderFiberToBuffer 方法

- [x] 1.2 创建 `ui/begin_work.go`
  - [x] BeginWork 入口函数
  - [x] beginWorkComponent
  - [x] beginWorkElement
  - [x] beginWorkText
  - [x] processUpdateQueue

- [x] 1.3 创建 `ui/complete_work.go`
  - [x] CompleteWork 入口函数
  - [x] completeWorkComponent
  - [x] completeWorkElement
  - [x] completeWorkText

- [x] 1.4 集成到 `ui/app.go`
  - [x] declarativeRoot 添加 reconciler 字段
  - [x] 添加 useFiber 开关
  - [x] 实现 paintWithFiber 方法
  - [x] 保留 paintLegacy 作为后备

- [x] 1.5 单元测试
  - [x] reconciler_test.go
  - [x] begin_work_test.go
  - [x] complete_work_test.go
```

#### 验收标准

```go
// ✅ 基础协调器工作
func TestReconciler_Basic(t *testing.T) {
    // 创建 reconciler
    reconciler := NewReconciler(app, ReconcilerConfig{})

    // 调度更新
    reconciler.ScheduleUpdate(LaneSyncLane)

    // 执行渲染
    reconciler.Render(ctx, buffer)

    // 验证输出正确
    assert.True(t, buffer.Modified)
}

// ✅ 组件树渲染
func TestFiber_ComponentTree(t *testing.T) {
    ui.Run(func() ui.VNode {
        count, setCount := ui.UseStateInt(0)
        return ui.VStack(
            ui.Text("Counter"),
            ui.Button("+").OnClick(func() {
                setCount(count + 1)
            }),
        )
    }, ui.UseFiber(true))
}
```

---

### Week 2: Commit 阶段 + Effects

#### 任务清单

```markdown
- [x] 2.1 创建 `ui/commit.go`
  - [x] CommitRoot 主方法
  - [x] commitBeforeMutationEffects
  - [x] commitMutationEffects (渲染到 buffer)
  - [x] commitLayoutEffects (执行 useEffect)

- [x] 2.2 创建 `ui/effects.go`
  - [x] Effect 结构体定义
  - [x] collectEffects 遍历收集
  - [x] runEffects 执行 effects
  - [x] flushPassiveEffects 执行被动 effects
  - [x] cleanupEffects 清理 effects

- [x] 2.3 Buffer 渲染集成
  - [x] renderFiberToBuffer 实现
  - [x] 处理所有 VNode 类型
  - [x] 样式应用
  - [x] 边界追踪

- [x] 2.4 测试
  - [x] fiber_test.go
  - [x] hooks_test.go
  - [x] 集成测试
```

#### 验收标准

- [ ] 状态更新正确反映到 buffer
- [ ] useEffect 正确执行和清理
- [ ] 组件状态持久化
- [ ] 无内存泄漏

---

### Week 3: 时间切片 + 优先级调度

#### 任务清单

```markdown
- [x] 3.1 增强 reconciler.go
  - [x] WorkLoopWithTimeSlice 方法
  - [x] deadline 检查逻辑
  - [x] requestWork 继续请求
  - [x] hasMoreWork 判断

- [x] 3.2 Lane 优先级系统
  - [x] ensureRootIsScheduled
  - [x] getNextLane 获取最高优先级
  - [x] markRootFinished 标记完成
  - [x] Lane 合并操作

- [x] 3.3 输入优先级
  - [x] SyncLane 处理（用户输入）
  - [x] InputContinuousLane 处理（拖拽等）
  - [x] DefaultLane 处理（数据更新）
  - [x] IdleLane 处理（后台任务）

- [x] 3.4 性能测试
  - [x] 大型组件树测试（1000+ 节点）
  - [x] 时间切片验证
  - [x] 内存泄漏检测
  - [x] 响应时间测试
```

#### 验收标准

| 指标 | 目标值 | 测试方法 |
|------|--------|---------|
| 组件树规模 | 1000+ 节点 | 创建大型测试树 |
| 响应延迟 | < 16ms | 输入延迟测试 |
| 时间切片 | 5ms 预算 | deadline 检查 |
| 内存 | 无泄漏 | 内存分析 |

---

### Week 4: Key 协调算法

#### 任务清单

```markdown
- [x] 4.1 创建 `ui/reconcile.go`
  - [x] reconcileChildren 主方法
  - [x] reconcileChildrenArray (Phase 1)
  - [ ] reconcileChildrenWithKeys (Phase 2) 🔄
  - [x] mapRemainingChildren 建立 key 映射
  - [x] updateFromMap 更新已有 fiber

- [ ] 4.2 Key 匹配算法
  - [x] O(1) key 查找
  - [x] fiber 复用逻辑
  - [x] 移动/删除处理
  - [ ] 列表重排优化 🔄

- [x] 4.3 测试
  - [x] Key 匹配测试
  - [x] 列表增删测试
  - [x] 移动/重排测试
  - [x] 性能基准测试
```

#### 验收标准

```go
// ✅ Key 协调正确
func TestReconcile_Keys(t *testing.T) {
    old := ui.HStack(
        ui.Key("a", ui.Text("A")),
        ui.Key("b", ui.Text("B")),
        ui.Key("c", ui.Text("C")),
    )
    new := ui.HStack(
        ui.Key("a", ui.Text("A")),
        ui.Key("c", ui.Text("C")),  // 交换 b/c 位置
        ui.Key("b", ui.Text("B")),
    )

    // 应该复用 fiber，而不是重建
    // fiber a 复用
    // fiber c 移到新位置
    // fiber b 移到新位置
}

// ✅ 动态列表正确处理
func TestReconcile_DynamicList(t *testing.T) {
    items, _ := ui.UseState([]string{"A", "B", "C"})

    ui.VStack(
        ui.For(items, func(item string) ui.VNode {
            return ui.Key(item, ui.Text(item))
        }),
    )
    // 当 items 变化时，应该正确复用/创建 fiber
}
```

---

### 文件创建清单

| 文件 | 说明 | Week | 状态 |
|------|------|------|------|
| `ui/reconciler.go` | 核心协调器 | Week 1 | ✅ 完成 |
| `ui/begin_work.go` | BeginWork 阶段 | Week 1 | ✅ 完成 |
| `ui/complete_work.go` | CompleteWork 阶段 | Week 1 | ✅ 完成 |
| `ui/commit.go` | Commit 阶段 | Week 2 | ✅ 完成 |
| `ui/effects.go` | Effect 处理 | Week 2 | ✅ 完成 |
| `ui/reconcile.go` | 子节点协调 | Week 4 | 🔄 进行中 |

---

### 修改文件清单

| 文件 | 修改内容 | Week | 状态 |
|------|---------|------|------|
| `ui/app.go` | 集成 reconciler | Week 1 | ✅ 完成 |
| `ui/fiber.go` | 可能添加辅助方法 | Week 1-4 | ✅ 完成 |

---

### 进度跟踪

```
Week 1 [████████████████████████████████████] 100%  ✅ 已完成
Week 2 [████████████████████████████████████] 100%  ✅ 已完成
Week 3 [████████████████████████████░░░░░░░░]  75%  🔄 进行中
Week 4 [████████████████░░░░░░░░░░░░░░░░░░░░]  40%  🔄 进行中

总进度: [████████████████████████████████░░░░]  85%
```

---

### 阶段 2A 验收标准

#### 功能完整性

- [x] 核心 Phase 任务完成 (Week 1-2)
- [x] 单元测试覆盖率 ≥ 75%
- [x] 集成测试通过
- [ ] 完整文档 (进行中)

#### 性能指标

| 指标 | 目标值 | 当前值 |
|------|--------|--------|
| 可中断渲染 | 支持 | ✅ 基础支持 |
| 优先级调度 | 支持 | ✅ Lane 系统实现 |
| 组件树规模 | 1000+ 节点 | ✅ 测试通过 |
| 输入响应 | < 16ms | ✅ 满足 |

#### 兼容性

- [x] 现有应用无需修改即可工作
- [x] Hooks 继续正常工作
- [x] InstanceManager 继续管理组件实例
- [x] 环境变量控制（MINT_USE_FIBER）

---

### 相关文档

#### 新增文档

- [FIBER_ARCHITECTURE.md](design/FIBER_ARCHITECTURE.md) ✅ 完成
- [RECONCILER_IMPLEMENTATION.md](design/RECONCILER_IMPLEMENTATION.md) ✅ 完成
- [FIBER_IMPLEMENTATION_PLAN.md](design/FIBER_IMPLEMENTATION_PLAN.md) ✅ 完成

#### 更新文档

- [ ] `TODO.md` - 本文件，添加阶段 2A
- [ ] `IMPLEMENTATION_PLAN.md` - 更新为包含 Fiber
- [ ] `SYSTEM_ARCHITECTURE.md` - 添加 reconciler 集成说明

---

## 六、阶段 3: 渲染管线

**时间**: Day 12-16 (5 天)
**目标**: 实现 DrawCmd 和 Buffer Diff

### 相关文档

- [SYSTEM_ARCHITECTURE.md](design/SYSTEM_ARCHITECTURE.md) - 渲染管线设计
- [BENCHMARK.md](design/BENCHMARK.md) - 渲染性能指标

### 6.1 Day 12: DrawCmd 定义

#### 任务清单

- [ ] 创建 `framework/render/drawcmd.go`
  - [ ] 定义 `DrawCmd` 接口
  - [ ] 定义 `DrawText`
  - [ ] 定义 `DrawRect`
  - [ ] 定义 `DrawClip`
  - [ ] 定义 `DrawTransform`
- [ ] 编写单元测试

#### 验收标准

```go
// ✅ DrawCmd 接口
cmd := &DrawText{
    X:     0,
    Y:     0,
    Text:  "Hello",
    Style: style.Style{Bold: true},
}
assert.Equal(t, "text", cmd.Type())
```

---

### 6.2 Day 13: 光栅化器

#### 任务清单

- [ ] 创建 `framework/render/rasterize.go`
  - [ ] 实现 `Rasterize()` 函数
  - [ ] 处理 DrawText
  - [ ] 处理 DrawRect
  - [ ] 处理 DrawClip
- [ ] 创建 `framework/render/buffer_adapter.go`
  - [ ] 适配 runtime.Buffer
- [ ] 编写单元测试

---

### 6.3 Day 14: Buffer Diff 与 Style 优化 🔴

#### 任务清单

##### Buffer Diff

- [ ] 创建 `framework/render/buffer_diff.go`
  - [ ] 实现 `DiffBuffer()` 函数
  - [ ] 实现 `CellChange` 结构
  - [ ] 优化 Diff 算法
- [ ] 编写单元测试

##### Style Diff 优化 (新增)

- [ ] 创建 `framework/render/terminal_state.go`
  - [ ] 实现 `TerminalState` 结构
  - [ ] 实现 `Equals()` 方法
  - [ ] 实现 `Diff()` 方法
- [ ] 创建 `framework/render/style_change.go`
  - [ ] 实现 `StyleChange` 结构
  - [ ] 实现 `ToANSI()` 方法
  - [ ] 实现 `IsEmpty()` 检查
- [ ] 创建 `framework/render/rle.go`
  - [ ] 实现 `Run` 结构
  - [ ] 实现 `RunLengthEncoder`
  - [ ] 实现 RLE 编码算法
- [ ] 创建 `framework/render/optimizer.go`
  - [ ] 实现 `Optimizer` 结构
  - [ ] 实现 `WriteBuffer()` 方法
  - [ ] 集成 TerminalState 追踪
- [ ] 创建 `framework/render/compat.go`
  - [ ] 实现 `CompatibilityMode`
  - [ ] 实现模式切换
- [ ] 编写单元测试
- [ ] 性能基准测试

#### 性能要求

- [ ] Buffer Diff (全屏) < 1ms
- [ ] ANSI 优化减少 ≥ 95% 切换 (目标: 99%)
- [ ] 输出字节数减少 ≥ 90%
- [ ] 全屏渲染 < 5ms

---

### 6.4 Day 15: 渲染器集成

#### 任务清单

- [ ] 创建 `framework/render/renderer.go`
  - [ ] 实现 `Renderer` 结构
  - [ ] 实现 `Render()` 方法
  - [ ] 集成 DrawCmd -> Buffer
  - [ ] 集成 Buffer Diff -> ANSI
- [ ] 编写集成测试

---

### 6.5 Day 16: 渲染测试与优化

#### 任务清单

- [ ] 完整的渲染测试套件
- [ ] 性能基准测试
- [ ] 内存泄漏检测
- [ ] 优化热点代码

#### 性能基准

```go
// BenchmarkRenderSimple
func BenchmarkRenderSimple(b *testing.B) {
    renderer := NewRenderer()
    vnode := ui.Text("Hello").Bold(true)
    for i := 0; i < b.N; i++ {
        renderer.Render(vnode)
    }
}

// 目标: > 1000 ops/sec
```

---

### 6.6 阶段 3 验收标准

- [ ] DrawCmd 系统完整
- [ ] 光栅化正确工作
- [ ] Buffer Diff 正确
- [ ] ANSI 优化有效
- [ ] 渲染帧率 ≥ 60 FPS
- [ ] 单元测试覆盖率 ≥ 80%

### 6.7 阶段 3 输出文档

- [ ] `progress/phase_3_progress.md`
- [ ] `progress/phase_3_summary.md`
- [ ] `docs/rendering.md` - 渲染管线文档
- [ ] `runtime/paint` tests and `docs/render/README.md` - 渲染示例与验证入口

---

## 七、阶段 4: 组件系统

**时间**: Day 17-22 (6 天)
**目标**: 实现基础组件库

### 相关文档

- [API_DESIGN.md](design/API_DESIGN.md) - 组件 API
- [COMPONENT_CLASSIFICATION.md](design/COMPONENT_CLASSIFICATION.md) - 组件分类

### 4.1 组件开发规范

每个新组件必须包含：

1. **组件实现** (`components/xxx/xxx.go`)
2. **单元测试** (`components/xxx/xxx_test.go`)
3. **示例程序** (`examples/xxx_demo/`)
4. **API 文档** (`docs/components/xxx.md`)

### 4.2 Day 17: 组件基础设施

#### 任务清单

- [ ] 创建 `framework/components/` 目录
  - [ ] `basic/` - 基础组件
  - [ ] `form/` - 表单组件
  - [ ] `button/` - 按钮组件
  - [ ] `data/` - 数据展示
  - [ ] `feedback/` - 反馈组件
- [ ] 创建 `ui/components.go`
  - [ ] 导出所有组件
- [ ] 创建组件模板

---

### 4.3 Day 18: Text 组件

#### 任务清单

- [ ] 实现 `components/basic/text.go`
- [ ] 实现 `ui.Text()` 函数
- [ ] 支持链式调用
- [ ] 编写测试 `text_test.go`
- [ ] 创建示例 `examples/text_demo/`

#### 组件 API

```go
func Text(content string) *TextBuilder

type TextBuilder struct {
    // ...
}

func (b *TextBuilder) Content(s string) *TextBuilder
func (b *TextBuilder) FgColor(c color.Color) *TextBuilder
func (b *TextBuilder) BgColor(c color.Color) *TextBuilder
func (b *TextBuilder) Bold(v bool) *TextBuilder
func (b *TextBuilder) Italic(v bool) *TextBuilder
func (b *TextBuilder) Underline(v bool) *TextBuilder
func (b *TextBuilder) Align(a TextAlign) *TextBuilder
func (b *TextBuilder) MaxLines(n int) *TextBuilder
func (b *TextBuilder) Build() VNode
```

#### 示例程序要求

```go
// examples/text_demo/main.go
package main

func main() {
    ui.Run(func() ui.VNode {
        return ui.VStack(
            ui.Text("Basic Text"),
            ui.Text("Bold Text").Bold(true),
            ui.Text("Colored Text").FgColor(color.Red),
            ui.Text("Aligned Text").Align(ui.AlignCenter),
        )
    })
}
```

#### 测试要求

- [ ] 基本渲染测试
- [ ] 样式应用测试
- [ ] 对齐方式测试
- [ ] 多行文本测试
- [ ] 行数限制测试

---

### 4.4 Day 19: Button 组件

#### 任务清单

- [ ] 实现 `components/button/button.go`
- [ ] 实现 `ui.Button()` 函数
- [ ] 支持多种变体
- [ ] 编写测试
- [ ] 创建示例

#### 示例程序要求

```go
// examples/button_demo/main.go
func main() {
    count, _ := ui.UseStateInt(0)

    ui.Run(func() ui.VNode {
        return ui.VStack(
            ui.Text("Button Demo"),
            ui.Button("Default"),
            ui.Button("Primary").Variant(ui.ButtonVariantPrimary),
            ui.Button("Danger").Variant(ui.ButtonVariantDanger),
            ui.Text(fmt.Sprintf("Clicked: %d", count)),
            ui.Button("Increment").OnClick(func() {
                setCount(count + 1)
            }),
        )
    })
}
```

---

### 4.5 Day 20: Input 组件

#### 任务清单

##### Input 组件
- [ ] 实现 `components/form/input.go`
- [ ] 实现 `ui.Input()` 函数
- [ ] 支持受控和非受控模式
- [ ] 编写测试
- [ ] 创建示例

##### TextBuffer 实现 (新增)
- [ ] 实现 `framework/input/buffer.go` - 核心 UTF-32 rune 存储
- [ ] 实现 `framework/input/selection.go` - 选择区域管理
- [ ] 实现 `framework/input/cursor.go` - 光标移动
- [ ] 实现 `framework/input/line.go` - 行操作
- [ ] 实现 `framework/input/scroll.go` - 滚动控制
- [ ] 实现 `framework/input/history.go` - 撤销/重做
- [ ] 编写 TextBuffer 单元测试
- [ ] 创建文本编辑示例

---

### 4.6 Day 21: List 和 Table 组件

#### 任务清单

- [ ] 实现 `components/data/list.go`
- [ ] 实现 `components/data/table.go`
- [ ] 编写测试
- [ ] 创建示例

---

### 4.7 Day 22: 其他基础组件

#### 任务清单

- [ ] CheckBox 组件
- [ ] Separator 组件
- [ ] ProgressBar 组件
- [ ] 编写测试
- [ ] 创建示例

---

### 4.8 阶段 4 验收标准

- [ ] 至少 8 个基础组件完成
- [ ] 每个组件有单元测试
- [ ] 每个组件有示例程序
- [ ] 每个组件有 API 文档
- [ ] 所有示例可运行
- [ ] 测试覆盖率 ≥ 75%

### 4.9 阶段 4 输出文档

- [ ] `progress/phase_4_progress.md`
- [ ] `progress/phase_4_summary.md`
- [ ] `docs/components/` - 组件文档目录
- [ ] `examples/*_demo/` - 示例程序目录

---

## 八、阶段 5: 布局系统

**时间**: Day 23-30 (8 天)
**目标**: 实现 HStack/VStack/Grid/Absolute 声明式布局

### 相关文档

- [API_DESIGN.md](design/API_DESIGN.md) - 布局 API
- [SYSTEM_ARCHITECTURE.md](design/SYSTEM_ARCHITECTURE.md) - 布局系统设计
- [GRID_LAYOUT_DESIGN.md](design/GRID_LAYOUT_DESIGN.md) - Grid 布局设计
- [ABSOLUTE_POSITIONING_DESIGN.md](design/ABSOLUTE_POSITIONING_DESIGN.md) - Absolute 定位设计

### 5.1 Day 23: HStack 实现

#### 任务清单

- [ ] 创建 `framework/layout/hstack.go`
- [ ] 实现 `ui.HStack()` 函数
- [ ] 封装 runtime FlexLayout
- [ ] 支持链式调用
- [ ] 编写测试
- [ ] 创建示例

#### API 设计

```go
func HStack(children ...VNode) *LayoutBuilder

type LayoutBuilder struct {
    // ...
}

func (b *LayoutBuilder) Align(a Align) *LayoutBuilder
func (b *LayoutBuilder) AlignCross(a Align) *LayoutBuilder
func (b *LayoutBuilder) Gap(n int) *LayoutBuilder
func (b *LayoutBuilder) Padding(top, right, bottom, left int) *LayoutBuilder
func (b *LayoutBuilder) Width(n int) *LayoutBuilder
func (b *LayoutBuilder) Height(n int) *LayoutBuilder
```

---

### 5.2 Day 24: VStack 实现

#### 任务清单

- [ ] 创建 `framework/layout/vstack.go`
- [ ] 实现 `ui.VStack()` 函数
- [ ] 封装 runtime FlexLayout
- [ ] 编写测试
- [ ] 创建示例

---

### 5.3 Day 25: 其他布局组件

#### 任务清单

- [ ] Spacer 组件
- [ ] Box 容器
- [ ] Overlay 组件
- [ ] 编写测试
- [ ] 创建示例

---

### 5.4 Day 26: 布局测试与示例

#### 任务清单

- [ ] 完整的布局测试套件
- [ ] 布局示例程序
- [ ] 性能测试
- [ ] 文档更新

#### 示例程序

```go
// examples/layout_demo/main.go
func main() {
    ui.Run(func() ui.VNode {
        return ui.VStack(
            ui.Text("Layout Demo"),
            ui.HStack(
                ui.Text("Left").Flex(1),
                ui.Text("Center").Flex(2),
                ui.Text("Right").Flex(1),
            ).Gap(2),
            ui.Box().Border(true).Padding(2).Child(
                ui.Text("Box Content"),
            ),
        )
    })
}
```

---

### 5.5 Day 27-28: Grid 布局 (新增)

#### 任务清单

##### 核心实现
- [ ] 创建 `framework/layout/dimension.go` - Dimension 类型
- [ ] 创建 `framework/layout/grid.go` - Grid 布局
- [ ] 创建 `framework/layout/grid_algorithm.go` - 布局算法
- [ ] 创建 `framework/layout/grid_cache.go` - 缓存优化

##### UI 集成
- [ ] 实现 `ui.Grid()` 函数
- [ ] 实现 `ui.Cell()` 函数
- [ ] 实现 `ui.CellSpan()` 函数
- [ ] 实现 `ui.Fixed()`, `ui.Flex()`, `ui.Auto()` 函数

##### 测试与示例
- [ ] 编写 Grid 单元测试
- [ ] 创建 Dashboard 示例
- [ ] 创建 IDE 布局示例
- [ ] 测试跨行跨列

#### 示例程序

```go
// examples/fiber_firsts/grid_demo/main.go
func main() {
    ui.Run(func() ui.VNode {
        return ui.Grid(
            []ui.Dimension{ui.Fixed(10), ui.Fixed(10), ui.Flex(1)},
            []ui.Dimension{ui.Fixed(5), ui.Flex(1)},
            ui.Cell(0, 0, CpuPanel()),
            ui.Cell(0, 1, MemPanel()),
            ui.CellSpan(1, 0, 1, 3, LogsPanel()),
        )
    })
}
```

---

### 5.6 Day 29-30: Absolute 定位 (新增)

#### 任务清单

##### 核心实现
- [ ] 创建 `framework/layout/position.go` - Position 类型
- [ ] 创建 `framework/layout/absolute.go` - Absolute 定位
- [ ] 创建 `framework/layout/absolute_algorithm.go` - 定位算法
- [ ] 创建 `framework/layout/absolute_builder.go` - 链式 API

##### UI 集成
- [ ] 实现 `ui.Absolute()` 函数
- [ ] 实现 `ui.TopLeft()` 等快捷函数
- [ ] 实现 `ui.Center()` 函数
- [ ] 实现 Z-Index 排序

##### 测试与示例
- [ ] 编写 Absolute 单元测试
- [ ] 创建 Badge 示例
- [ ] 创建 居中 Modal 示例
- [ ] 测试百分比定位

#### 示例程序

```go
// examples/fiber_firsts/absolute_demo/main.go
func main() {
    ui.Run(func() ui.VNode {
        count, setCount := ui.UseState(0)

        return ui.Box().Padding(5).Child(
            ui.Stack(
                ui.Button("Click Me").OnClick(func() {
                    setCount(count + 1)
                }),
                // 右上角徽章
                ui.Absolute(
                    ui.Text(fmt.Sprintf("Count: %d", count)).
                        FgColor(ui.ColorRed).
                        Bold(true),
                ).Top(0).Right(2),
            ),
        )
    })
}
```

---

### 5.7 阶段 5 验收标准

- [ ] HStack/VStack 正常工作
- [ ] Flex 参数正确
- [ ] 对齐方式正确
- [ ] Grid 布局支持跨行跨列
- [ ] Grid 弹性空间分配正确
- [ ] Absolute 定位位置准确
- [ ] Absolute 百分比定位正确
- [ ] Z-Index 层级正确
- [ ] 所有布局示例可运行
- [ ] 测试覆盖率 ≥ 75%

### 5.8 阶段 5 输出文档

- [ ] `progress/phase_5_progress.md`
- [ ] `progress/phase_5_summary.md`
- [ ] `docs/layout.md` - 布局系统文档
- [ ] `examples/layout_demo/` - 布局示例
- [ ] `examples/fiber_firsts/grid_demo/` - Grid 示例
- [ ] `examples/fiber_firsts/absolute_demo/` - Absolute 示例

---

## 九、阶段 6: Hooks 系统

**时间**: Day 31-35 (5 天)
**目标**: 实现完整的 Hooks 系统

### 相关文档

- [API_DESIGN.md](design/API_DESIGN.md) - Hooks API
- [SYSTEM_ARCHITECTURE.md](design/SYSTEM_ARCHITECTURE.md) - Hooks 设计

### 6.1 Day 27: useState

#### 任务清单

- [ ] 创建 `framework/hooks/context.go`
  - [ ] 实现 HookContext
  - [ ] 实现 context 管理
- [ ] 创建 `framework/hooks/state.go`
  - [ ] 实现 useState
  - [ ] 实现类型安全版本
- [ ] 编写测试
- [ ] 创建示例

#### 示例程序

```go
// examples/usestate_demo/main.go
func main() {
    ui.Run(func() ui.VNode {
        return Counter()
    })
}

func Counter() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("Increment").OnClick(func() {
            setCount(count + 1)
        }),
        ui.Button("Decrement").OnClick(func() {
            setCount(count - 1)
        }),
    )
}
```

---

### 6.2 Day 28: useEffect

#### 任务清单

- [ ] 创建 `framework/hooks/effect.go`
  - [ ] 实现 useEffect
  - [ ] 实现依赖比较
  - [ ] 实现清理函数
- [ ] 编写测试
- [ ] 创建示例

#### 示例程序

```go
// examples/useeffect_demo/main.go
func Timer() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    ui.UseEffect(func() {
        ticker := time.NewTicker(time.Second)
        done := make(chan bool)

        go func() {
            for {
                select {
                case <-ticker.C:
                    setCount(count + 1)
                case <-done:
                    return
                }
            }
        }()

        return func() {
            ticker.Stop()
            close(done)
        }
    }, nil)

    return ui.Text(fmt.Sprintf("Time: %d", count))
}
```

---

### 6.3 Day 29: useContext

#### 任务清单

- [ ] 创建 `framework/hooks/context_hook.go`
  - [ ] 实现 useContext
  - [ ] 实现 createContext
  - [ ] 实现 Provider
- [ ] 编写测试
- [ ] 创建示例

---

### 6.4 Day 30: useMemo 和 useCallback

#### 任务清单

- [ ] 创建 `framework/hooks/memo.go`
  - [ ] 实现 useMemo
  - [ ] 实现 useCallback
- [ ] 编写测试
- [ ] 创建示例

---

### 6.5 Day 31: useRef 和其他 Hooks

#### 任务清单

- [ ] 创建 `framework/hooks/ref.go`
  - [ ] 实现 useRef
  - [ ] 实现 useImperativeHandle
- [ ] 创建 `framework/hooks/reducer.go`
  - [ ] 实现 useReducer
- [ ] 编写测试
- [ ] 创建示例

---

### 6.6 阶段 6 验收标准

- [ ] 所有基础 Hooks 实现
- [ ] Hooks 调用规则检查
- [ ] 状态更新正确触发
- [ ] Effect 清理正确执行
- [ ] 示例程序完整
- [ ] 测试覆盖率 ≥ 80%

### 6.7 阶段 6 输出文档

- [ ] `progress/phase_6_progress.md`
- [ ] `progress/phase_6_summary.md`
- [ ] `docs/hooks.md` - Hooks 文档
- [ ] `examples/hooks_demo/` - Hooks 示例

---

## 十、阶段 7: 高级特性

**时间**: Day 36-42 (7 天)
**目标**: 实现虚拟化、动画和 Layer 系统

### 相关文档

- [LAYER_SYSTEM_DESIGN.md](design/LAYER_SYSTEM_DESIGN.md) - Layer 系统设计

### 7.1 Day 32-33: 虚拟化渲染

#### 任务清单

- [ ] 创建 `framework/components/data/virtual_list.go`
- [ ] 实现 VirtualList 组件
- [ ] 实现可见范围计算
- [ ] 编写测试
- [ ] 创建大列表示例 (10000+ 项)

#### 示例程序

```go
// examples/virtuallist/main.go
func main() {
    // 生成 100000 项数据
    items := make([]string, 100000)
    for i := 0; i < 100000; i++ {
        items[i] = fmt.Sprintf("Item %d", i)
    }

    ui.Run(func() ui.VNode {
        return ui.VirtualList(items).
            ItemHeight(1).
            RenderItem(func(item interface{}) ui.VNode {
                return ui.Text(item.(string))
            })
    })
}
```

---

### 7.2 Day 34-36: 动画系统

#### 任务清单

- [ ] 扩展 `framework/animation/`
- [ ] 实现 useAnimation Hook
- [ ] 实现过渡动画
- [ ] 编写测试
- [ ] 创建示例

---

### 7.3 Day 41-42: Layer 系统 (新增)

#### 任务清单

##### Layer 核心实现
- [ ] 创建 `framework/layer/layer.go` - Layer 类型定义
- [ ] 创建 `framework/layer/tree.go` - LayerTree 实现
- [ ] 创建 `framework/layer/manager.go` - Manager 实现
- [ ] 创建 `framework/layer/event.go` - 事件处理
- [ ] 创建 `framework/layer/render.go` - 渲染集成
- [ ] 创建 `framework/layer/layout.go` - 布局处理
- [ ] 创建 `framework/layer/focus.go` - 焦点管理

##### Layer UI 集成
- [ ] 实现 `ui.Layer()` 函数
- [ ] 实现 `ui.Modal()` 函数
- [ ] 实现 `ui.CloseModal()` 函数
- [ ] 实现 `ui.Tooltip()` 函数
- [ ] 实现 `ui.Toast()` 函数

##### 测试与示例
- [ ] 编写 LayerManager 单元测试
- [ ] 编写 Modal 示例
- [ ] 编写 Tooltip 示例
- [ ] 编写 Toast 示例
- [ ] 测试模态框嵌套
- [ ] 测试 ESC 键关闭

#### 示例程序

```go
// examples/modal_demo/main.go
func main() {
    showModal, setShowModal := ui.useState(false)

    ui.Run(func() ui.VNode {
        if showModal {
            ui.Modal("confirm-modal", ConfirmDialog())
        }

        return ui.VStack(
            ui.Text("Main Content"),
            ui.Button("Show Modal").OnClick(func() {
                setShowModal(true)
            }),
        )
    })
}

func ConfirmDialog() ui.VNode {
    return ui.Box().Border(true).Padding(2).Child(
        ui.VStack(
            ui.Text("Confirm").Bold(true),
            ui.Separator(),
            ui.Text("Are you sure?"),
            ui.HStack(
                ui.Button("Cancel").OnClick(func() {
                    ui.CloseModal()
                }),
                ui.Button("OK"),
            ),
        ),
    )
}
```

---

### 7.4 阶段 7 验收标准

- [ ] VirtualList 支持 100000+ 项
- [ ] 动画流畅 (≥ 60 FPS)
- [ ] Layer 系统正常工作
- [ ] Modal/Tooltip/Toast 可用
- [ ] 焦点陷阱正常工作
- [ ] 示例程序完整
- [ ] 测试覆盖率 ≥ 75%

### 7.5 阶段 7 输出文档

- [ ] `progress/phase_7_progress.md`
- [ ] `progress/phase_7_summary.md`
- [ ] `docs/advanced.md` - 高级特性文档
- [ ] `docs/layer.md` - Layer 系统文档
- [ ] `examples/virtuallist/` - 虚拟化示例
- [ ] `examples/modal_demo/` - Modal 示例

---

## 十一、阶段 8: DevTools 集成

**时间**: Day 43-46 (4 天)
**目标**: 集成 DevTools 调试支持

### 相关文档

- [SYSTEM_ARCHITECTURE.md](design/SYSTEM_ARCHITECTURE.md) - DevTools 设计

### 8.1 Day 39: DevTools 桥接

#### 任务清单

- [ ] 创建 `devtools/bridge/fiber.go`
  - [ ] 实现 Fiber 树导出
  - [ ] 实现组件状态导出
- [ ] 创建 `devtools/bridge/inspector.go`
  - [ ] 实现组件检查器
- [ ] 编写测试

---

### 8.2 Day 44-45: 性能分析

#### 任务清单

- [ ] 创建 `devtools/bridge/profiler.go`
  - [ ] 实现性能收集
  - [ ] 实现火焰图生成
- [ ] 创建 `devtools/bridge/layout.go`
  - [ ] 实现布局调试
- [ ] 编写测试

---

### 8.3 Day 46: DevTools UI

#### 任务清单

- [ ] 更新 Web Dashboard
- [ ] 添加 Fiber 树视图
- [ ] 添加性能面板
- [ ] 添加布局调试面板
- [ ] 测试

---

### 8.4 阶段 8 验收标准

- [ ] DevTools 可查看 Fiber 树
- [ ] 可查看组件 Props/State
- [ ] 性能分析正常工作
- [ ] 布局调试可用

### 8.5 阶段 8 输出文档

- [ ] `progress/phase_8_progress.md`
- [ ] `progress/phase_8_summary.md`
- [ ] `docs/devtools.md` - DevTools 文档

---

## 十二、阶段 9: 文档与示例

**时间**: Day 47-50 (4 天)
**目标**: 完善文档和示例

### 9.1 Day 47: API 文档

#### 任务清单

- [ ] 完善 `docs/api/` 目录
- [ ] 为每个组件编写 API 文档
- [ ] 为每个 Hook 编写 API 文档
- [ ] 生成 godoc

---

### 9.2 Day 48-49: 示例程序

#### 任务清单

- [ ] Hello World 示例
- [ ] Counter 示例
- [ ] Todo List 示例
- [ ] Form 示例
- [ ] Dashboard 示例
- [ ] 确保所有示例可运行

---

### 9.3 Day 50: 教程和指南

#### 任务清单

- [ ] 快速开始指南
- [ ] 组件开发教程
- [ ] Hooks 使用教程
- [ ] 最佳实践文档
- [ ] 迁移指南更新

---

### 9.4 阶段 9 验收标准

- [ ] API 文档完整
- [ ] 所有示例可运行
- [ ] 教程清晰易懂
- [ ] godoc 可生成

### 9.5 阶段 9 输出文档

- [ ] `progress/phase_9_progress.md`
- [ ] `progress/phase_9_summary.md`
- [ ] `docs/` - 完整文档目录
- [ ] `examples/` - 完整示例目录

---

## 十三、阶段 10: 测试与发布

**时间**: Day 51-53 (3 天)
**目标**: 质量保证和发布

### 10.1 Day 51: 全面测试

#### 任务清单

- [ ] 运行所有单元测试
- [ ] 运行所有集成测试
- [ ] 运行性能基准测试
- [ ] 竞态检测
- [ ] 内存泄漏检测

---

### 10.2 Day 52: 发布准备

#### 任务清单

- [ ] 更新 CHANGELOG.md
- [ ] 更新 README.md
- [ ] 创建 Release Notes
- [ ] 准备版本号
- [ ] 打 Git tag

---

### 10.3 Day 53: 发布

#### 任务清单

- [ ] 合并到 main 分支
- [ ] 创建 GitHub Release
- [ ] 发布到 go modules
- [ ] 发布公告
- [ ] 庆祝 🎉

---

### 10.4 阶段 10 验收标准

- [ ] 所有测试通过
- [ ] 无已知 Critical bug
- [ ] 性能达标
- [ ] 文档完整
- [ ] 发布成功

### 10.5 阶段 10 输出文档

- [ ] `progress/phase_10_progress.md`
- [ ] `progress/phase_10_summary.md`
- [ ] `CHANGELOG.md`
- [ ] `RELEASE_NOTES.md`

---

## 附录

### A. 每日日志模板

```markdown
# YYYY-MM-DD - [阶段名称] [任务名称]

## 今日完成
- [x] 任务 1
- [x] 任务 2

## 进行中
- [ ] 任务 3 (50%)

## 遇到问题
- 问题描述:
  - 影响:
  - 解决方案:

## 明日计划
- [ ] 任务 3
- [ ] 任务 4

## 代码提交
- Commit: hash
- PR: #

## 学习笔记
- 今日学到的新知识
```

---

### B. 阶段总结模板

```markdown
# 阶段 X 总结

## 概述
- 开始日期:
- 结束日期:
- 实际工期:
- 计划工期:

## 完成情况
- 计划任务数:
- 实际完成:
- 完成率:

## 交付物
- [ ] 交付物 1
- [ ] 交付物 2

## 遇到的问题
1. 问题描述
   - 解决方案:
   - 经验教训:

2. 问题描述
   - 解决方案:
   - 经验教训:

## 性能指标
- 测试覆盖率:
- 性能基准:
- 内存使用:

## 代码统计
- 新增文件:
- 修改文件:
- 新增代码行:
- 测试代码行:

## 下一步
- 进入阶段 X+1
- 重点注意事项:
```

---

### C. 进度更新命令

```bash
# 更新总进度
# 编辑 TODO.md 第一部分的进度条

# 更新阶段进度
# 编辑 progress/phase_X_progress.md

# 记录每日日志
# 创建 progress/daily/YYYY-MM-DD.md

# 更新任务状态
# 在 TODO.md 中将 [ ] 改为 [x]
```

---

### D. 代码提交规范

```bash
# 提交格式
git commit -m "<type>: <subject>

<body>

<footer>

# 类型
feat:     新功能
fix:      修复 bug
docs:     文档更新
test:     测试相关
refactor: 重构
perf:     性能优化
style:    代码格式
chore:    构建/工具

# 示例
git commit -m "feat(ui): implement VNode interface

- Add VNode interface with Type/Props/Children methods
- Implement ElementVNode and TextVNode
- Add unit tests with 85% coverage

Refs: #1
Related to #2"
```

---

### E. 检查清单

#### 开发前检查

- [ ] 阅读相关设计文档
- [ ] 理解任务要求
- [ ] 确认依赖已就绪
- [ ] 创建功能分支

#### 开发中检查

- [ ] 代码符合规范
- [ ] 添加必要注释
- [ ] 编写单元测试
- [ ] 运行测试验证

#### 提交前检查

- [ ] 所有测试通过
- [ ] 代码已自审
- [ ] 更新相关文档
- [ ] 更新 TODO.md

#### 阶段结束检查

- [ ] 所有任务完成
- [ ] 验收标准满足
- [ ] 文档已更新
- [ ] 总结已编写

---

## 文档索引

### 设计文档

| 文档 | 路径 | 状态 |
|------|------|------|
| 系统架构 | `design/SYSTEM_ARCHITECTURE.md` | ✅ |
| 差距分析 | `design/IMPLEMENTATION_GAP_ANALYSIS.md` | ✅ |
| 目录结构 | `design/DIRECTORY_STRUCTURE.md` | ✅ |
| 组件分类 | `design/COMPONENT_CLASSIFICATION.md` | ✅ |
| 实施计划 | `design/IMPLEMENTATION_PLAN.md` | ✅ |
| API 设计 | `design/API_DESIGN.md` | ✅ |
| 迁移指南 | `design/MIGRATION_GUIDE.md` | ✅ |
| 性能基准 | `design/BENCHMARK.md` | ✅ |
| Fiber 架构 | `design/FIBER_ARCHITECTURE.md` | ✅ |
| Reconciler 实现 | `design/RECONCILER_IMPLEMENTATION.md` | ✅ |
| Fiber 实施计划 | `design/FIBER_IMPLEMENTATION_PLAN.md` | ✅ |

### 进度文档

| 文档 | 路径 | 状态 |
|------|------|------|
| TODO 列表 | `TODO.md` | ✅ |
| 阶段 0 进度 | `progress/phase_0_progress.md` | ⏳ |
| 阶段 1 进度 | `progress/phase_1_progress.md` | ⏳ |
| ... | ... | ... |

---

**文档结束**

**最后更新**: 2026-02-01
**维护者**: Mint UI Team
