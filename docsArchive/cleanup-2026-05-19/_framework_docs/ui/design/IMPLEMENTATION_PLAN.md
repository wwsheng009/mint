# Mint UI 声明式架构实施计划

**版本**: v1.1
**日期**: 2026-01-31
**基于**: SYSTEM_ARCHITECTURE.md, IMPLEMENTATION_GAP_ANALYSIS.md

---

## 执行摘要

本实施计划将 Mint UI 从当前命令式架构迁移到声明式架构，采用 **MVP 优先策略**，分为 6 个阶段，预计 **10 周**完成（含缓冲时间）。

### 🎯 MVP 优先策略

**核心原则**：先交付最小可行产品，再迭代增强。

| 优先级 | 功能 | 阶段 | 说明 |
|--------|------|------|------|
| **P0** | VNode + useState + 基础 Diff | Phase 0 | MVP 核心，必须首先完成 |
| **P0** | Text/Button/Input 组件 | Phase 1 | 基础交互能力 |
| **P1** | Fiber 架构 + useEffect | Phase 2 | 完整 Reconciler |
| **P1** | HStack/VStack 布局 | Phase 2 | 基础布局 |
| **P2** | Grid/Absolute 布局 | Phase 3 | 高级布局 |
| **P2** | Layer 系统 | Phase 3 | Modal/Tooltip |
| **P3** | Scheduler 时间切片 | Phase 4 | 性能优化 |
| **P3** | 虚拟化渲染 | Phase 4 | 大数据量支持 |
| **P4** | DevTools 集成 | Phase 5 | 调试工具 |

### 关键里程碑（含缓冲时间）

| 阶段 | 周期 | 缓冲 | 交付物 | 状态 |
|------|------|------|--------|------|
| **Phase 0** | Week 1 | +2天 | **MVP**: VNode + useState + 基础 Diff | ⏳ 待开始 |
| Phase 1 | Week 2 | +2天 | 基础组件 (Text/Button/Input) | ⏳ 待开始 |
| Phase 2 | Week 3-4 | +3天 | Fiber + Hooks + 布局 API | ⏳ 待开始 |
| Phase 3 | Week 5-6 | +3天 | 高级布局 + Layer 系统 | ⏳ 待开始 |
| Phase 4 | Week 7-8 | +2天 | Scheduler + 虚拟化 | ⏳ 待开始 |
| Phase 5 | Week 9-10 | +2天 | DevTools + 文档 + 测试 | ⏳ 待开始 |

**总缓冲时间**: 14 天（约 2 周），用于应对不可预见的技术挑战。

---

## Phase 0: MVP 核心 (Week 1) 🔴 最高优先级

### 目标
交付最小可行的声明式 UI 系统，验证核心架构可行性。

### MVP 定义

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

func main() {
    ui.Run(App)
}
```

### MVP 任务清单

```markdown
- [ ] 0.1 VNode 接口 + ElementVNode 实现
- [ ] 0.2 useState Hook（最简实现）
- [ ] 0.3 基础 Diff（类型比较 + Props Diff）
- [ ] 0.4 简单渲染循环（无时间切片）
- [ ] 0.5 Text 组件
- [ ] 0.6 Button 组件（OnClick）
- [ ] 0.7 VStack 布局（复用现有 FlexLayout）
- [ ] 0.8 MVP 集成测试
```

### MVP 验收标准

- [ ] Counter 示例可运行
- [ ] 状态更新触发 UI 刷新
- [ ] 基础 Diff 正确识别变化
- [ ] 无明显性能问题（手动测试）

---

## Phase 1: 基础架构 (Week 1-2)

### 目标
建立声明式 UI 的核心基础设施：VNode 系统和 Hooks 系统。

### 1.1 VNode 系统 (Week 1)

#### 任务清单

```markdown
- [ ] 1.1.1 创建 ui/ 目录结构
- [ ] 1.1.2 定义 VNode 接口 (ui/vnode.go)
- [ ] 1.1.3 实现各类型 VNode
  - [ ] ElementVNode
  - [ ] TextVNode
  - [ ] ComponentVNode
  - [ ] FragmentVNode
- [ ] 1.1.4 实现 Props 系统
- [ ] 1.1.5 实现 Key 机制
- [ ] 1.1.6 编写 VNode 单元测试
```

#### 文件结构

```
ui/
├── vnode.go              # VNode 接口和类型定义
├── vnode_types.go        # VNode 具体实现
├── props.go              # Props 定义
└── vnode_test.go         # 测试
```

#### 接口定义

```go
// ui/vnode.go
type VNode interface {
    Type() VNodeType
    Props() Props
    Children() []VNode
    Key() string
    SetKey(key string)
}

type VNodeType int

const (
    VNodeElement VNodeType = iota
    VNodeText
    VNodeComponent
    VNodeFragment
)
```

#### 验收标准

- [ ] VNode 接口定义完整
- [ ] 四种 VNode 类型实现完成
- [ ] Props 支持基本类型和函数
- [ ] Key 机制用于 Diff 优化
- [ ] 单元测试覆盖率 > 80%

---

### 1.2 Hooks 系统 (Week 1-2)

#### 任务清单

```markdown
- [ ] 1.2.1 创建 framework/hooks/ 目录
- [ ] 1.2.2 实现 useState
- [ ] 1.2.3 实现 useEffect
- [ ] 1.2.4 实现 useContext
- [ ] 1.2.5 实现 useMemo
- [ ] 1.2.6 实现 useCallback
- [ ] 1.2.7 实现 useRef
- [ ] 1.2.8 实现 useReducer
- [ ] 1.2.9 实现 Hooks 调用规则检查
- [ ] 1.2.10 编写 Hooks 测试
```

#### 文件结构

```
framework/hooks/
├── context.go            # Hooks 上下文管理
├── state.go              # useState
├── effect.go             # useEffect
├── context_hook.go       # useContext
├── memo.go               # useMemo, useCallback
├── ref.go                # useRef, useImperativeHandle
├── reducer.go            # useReducer
├── rules.go              # 调用规则检查
└── hooks_test.go         # 测试
```

#### 实现示例

```go
// framework/hooks/state.go
func useState(initial interface{}) (interface{}, func(interface{})) {
    ctx := currentContext()
    hook := ctx.nextHook()

    if hook.State == nil {
        hook.State = initial
    }

    setState := func(newValue interface{}) {
        hook.State = newValue
        ctx.markDirty()
    }

    return hook.State, setState
}
```

#### 验收标准

- [ ] 所有基础 Hooks 实现完成
- [ ] Hooks 调用顺序检查生效
- [ ] 状态更新触发重新渲染
- [ ] Effect 清理函数正确执行
- [ ] 单元测试覆盖所有 Hooks

---

## Phase 2: Reconciler + 渲染 (Week 3-4)

### 目标
实现 Reconciler 系统（Diff + Fiber）和渲染管线（DrawCmd）。

### 2.1 Diff 算法 (Week 3)

#### 任务清单

```markdown
- [ ] 2.1.1 创建 framework/reconciler/ 目录
- [ ] 2.1.2 实现 Patch 类型定义
- [ ] 2.1.3 实现基础 Diff 算法
- [ ] 2.1.4 实现 Key 优化 Diff
- [ ] 2.1.5 实现子节点 Diff（双指针算法）
- [ ] 2.1.6 编写 Diff 测试用例
```

#### 文件结构

```
framework/reconciler/
├── diff.go               # Diff 算法
├── patch.go              # Patch 定义
├── diff_test.go          # 测试
└── diff_bench_test.go    # 基准测试
```

#### 实现示例

```go
// framework/reconciler/diff.go
func Diff(old, new VNode) Patch {
    switch {
    case old == nil && new != nil:
        return &CreatePatch{Node: new}
    case old != nil && new == nil:
        return &DeletePatch{}
    case old.Type() != new.Type():
        return &ReplacePatch{New: new}
    case old.Key() != new.Key():
        return &ReplacePatch{New: new}
    default:
        return diffProps(old, new)
    }
}
```

#### 验收标准

- [ ] Create/Delete/Replace Patch 正确生成
- [ ] 同类型节点进行 Props Diff
- [ ] Key 优化减少不必要的替换
- [ ] 子节点 Diff 使用双指针算法
- [ ] 基准测试性能符合预期

---

### 2.2 Fiber 架构 (Week 3)

#### 任务清单

```markdown
- [ ] 2.2.1 实现 Fiber 节点结构
- [ ] 2.2.2 实现 Fiber 树构建
- [ ] 2.2.3 实现 BeginWork
- [ ] 2.2.4 实现 CompleteWork
- [ ] 2.2.5 实现 Effect 链
- [ ] 2.2.6 编写 Fiber 测试
```

#### 文件结构

```
framework/reconciler/
├── fiber.go              # Fiber 节点
├── fiber_tree.go         # Fiber 树操作
├── begin_work.go         # BeginWork
├── complete_work.go      # CompleteWork
├── effects.go            # Effect 处理
└── fiber_test.go         # 测试
```

#### 实现示例

```go
// framework/reconciler/fiber.go
type Fiber struct {
    // VNode 关联
    VNode VNode

    // 树结构
    Return  *Fiber
    Child   *Fiber
    Sibling *Fiber

    // 工作单元状态
    PendingProps  Props
    MemoizedProps Props
    UpdateQueue   *UpdateQueue

    // Effect
    EffectTag  EffectTag
    NextEffect *Fiber

    // 优先级
    Lanes Lanes
}

type EffectFlag uint32

const (
    Placement EffectFlag = 1 << iota
    Update
    Ref
)
```

#### 验收标准

- [ ] Fiber 节点正确映射 VNode 树
- [ ] BeginWork 处理组件渲染
- [ ] CompleteWork 收集 Effect
- [ ] Effect 链按正确顺序链接
- [ ] 测试覆盖所有工作阶段

---

### 2.3 Scheduler (Week 3)

#### 任务清单

```markdown
- [ ] 2.3.1 定义 Lane 优先级
- [ ] 2.3.2 实现任务队列
- [ ] 2.3.3 实现工作循环
- [ ] 2.3.4 实现时间切片
- [ ] 2.3.5 实现 requestAnimationFrame
- [ ] 2.3.6 编写 Scheduler 测试
```

#### 文件结构

```
framework/reconciler/
├── scheduler.go          # 调度器
├── lanes.go              # 优先级定义
├── workloop.go           # 工作循环
└── scheduler_test.go     # 测试
```

#### 实现示例

```go
// framework/reconciler/scheduler.go
type Lane uint64

const (
    SyncLane      Lane = 0b00000001
    InputLane     Lane = 0b00000010
    AnimationLane Lane = 0b00000100
    TransitionLane Lane = 0b00001000
    IdleLane      Lane = 0b10000000
)

type Scheduler struct {
    taskQueue  []*Task
    lanes      Lanes
    running    bool
}

func (s *Scheduler) Schedule(callback func(), lanes Lanes) {
    task := &Task{Callback: callback, Lanes: lanes}
    s.taskQueue = append(s.taskQueue, task)
}

func (s *Scheduler) WorkLoop(deadline time.Time) {
    for !time.Now().After(deadline) {
        s.performUnitOfWork()
    }
}
```

#### 验收标准

- [ ] 优先级正确排序任务
- [ ] 时间切片正确中断工作
- [ ] 高优先级任务抢占低优先级
- [ ] 空闲时间处理 IdleLane 任务

---

### 2.4 渲染管线 (Week 4)

#### 任务清单

```markdown
- [ ] 2.4.1 创建 framework/render/ 目录
- [ ] 2.4.2 定义 DrawCmd 类型
- [ ] 2.4.3 实现 DrawText
- [ ] 2.4.4 实现 DrawRect
- [ ] 2.4.5 实现 DrawClip
- [ ] 2.4.6 实现光栅化器
- [ ] 2.4.7 实现 Buffer Diff
- [ ] 2.4.8 实现 ANSI 优化输出
- [ ] 2.4.9 编写渲染测试
```

#### 文件结构

```
framework/render/
├── drawcmd.go            # DrawCmd 定义
├── rasterize.go          # 光栅化
├── buffer_diff.go        # Buffer Diff
├── ansi.go               # ANSI 优化
├── buffer_adapter.go     # Buffer 适配器
└── render_test.go        # 测试
```

#### 实现示例

```go
// framework/render/drawcmd.go
type DrawCmd interface {
    Type() DrawCmdType
}

type DrawText struct {
    X, Y  int
    Text  string
    Style style.Style
}

type DrawRect struct {
    X, Y, W, H int
    Style      style.Style
}

type DrawClip struct {
    X, Y, W, H int
}

// framework/render/buffer_diff.go
func DiffBuffer(old, new *runtime.Buffer) []CellChange {
    changes := []CellChange{}
    for y := 0; y < old.Height; y++ {
        for x := 0; x < old.Width; x++ {
            if !cellsEqual(old.Cells[y][x], new.Cells[y][x]) {
                changes = append(changes, CellChange{
                    X:    x,
                    Y:    y,
                    Cell: new.Cells[y][x],
                })
            }
        }
    }
    return changes
}
```

#### 验收标准

- [ ] DrawCmd 正确转换为 Buffer 操作
- [ ] Buffer Diff 正确识别变化
- [ ] ANSI 输出优化切换次数
- [ ] 渲染性能符合目标（60 FPS）

---

## Phase 3: 组件 + 布局 (Week 5-6)

### 目标
实现声明式组件 API 和布局封装。

### 3.1 声明式 API (Week 5)

#### 任务清单

```markdown
- [ ] 3.1.1 实现 ui.Builder
- [ ] 3.1.2 实现 ui.Text
- [ ] 3.1.3 实现 ui.Button
- [ ] 3.1.4 实现 ui.Input
- [ ] 3.1.5 实现 HStack/VStack
- [ ] 3.1.6 实现链式调用
- [ ] 3.1.7 实现事件绑定
```

#### 文件结构

```
ui/
├── app.go                # Run 函数
├── vnode.go              # VNode 接口
├── builder.go            # Builder 模式
├── components.go         # 组件构建器
├── layout.go             # 布局构建器
├── events.go             # 事件绑定
└── style.go              # 样式设置
```

#### API 示例

```go
// 声明式 API 示例
func App() VNode {
    count, setCount := ui.UseState(0)

    return ui.VStack(
        ui.Text("Counter Example").Style(style.Bold),
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.HStack(
            ui.Button("-").OnClick(func() {
                setCount(count - 1)
            }),
            ui.Button("+").OnClick(func() {
                setCount(count + 1)
            }),
        ),
    )
}
```

#### 验收标准

- [ ] 所有基础组件可用
- [ ] 链式调用流畅
- [ ] 事件绑定正确
- [ ] 布局组件正确排列子节点

---

### 3.2 组件库 (Week 5-6)

#### 任务清单

```markdown
- [ ] 3.2.1 创建 framework/components/ 目录
- [ ] 3.2.2 实现 basic 组件
  - [ ] Text
  - [ ] Icon
  - [ ] Separator
- [ ] 3.2.3 实现 form 组件
  - [ ] Input
  - [ ] TextArea
  - [ ] CheckBox
  - [ ] Select
- [ ] 3.2.4 实现 data 组件
  - [ ] List
  - [ ] Table
- [ ] 3.2.5 实现 feedback 组件
  - [ ] ProgressBar
  - [ ] Toast
- [ ] 3.2.6 迁移现有组件到新结构
```

#### 文件结构

```
framework/components/
├── basic/
│   ├── text.go
│   ├── icon.go
│   └── separator.go
├── form/
│   ├── input.go
│   ├── textarea.go
│   ├── checkbox.go
│   └── select.go
├── data/
│   ├── list.go
│   └── table.go
├── feedback/
│   ├── progress.go
│   └── toast.go
└── button/
    └── button.go
```

#### 验收标准

- [ ] 组件目录结构清晰
- [ ] 每个组件有完整测试
- [ ] 组件文档完整
- [ ] 旧组件迁移到 `_legacy`

---

### 3.3 布局系统 (Week 6)

#### 任务清单

```markdown
- [ ] 3.3.1 创建 framework/layout/ 目录
- [ ] 3.3.2 实现 HStack 封装
- [ ] 3.3.3 实现 VStack 封装
- [ ] 3.3.4 实现 Spacer
- [ ] 3.3.5 实现 Align 对齐
- [ ] 3.3.6 实现 Flex 参数
- [ ] 3.3.7 实现虚拟化布局
```

#### 文件结构

```
framework/layout/
├── hstack.go             # 水平布局
├── vstack.go             # 垂直布局
├── stack.go              # 通用堆叠
├── spacer.go             # 弹性空间
├── align.go              # 对齐
├── flex.go               # Flex 参数
└── virtual.go            # 虚拟化
```

#### 验收标准

- [ ] HStack/VStack 正确排列子节点
- [ ] Flex 参数正确分配空间
- [ ] 对齐方式正确应用
- [ ] 虚拟化只渲染可见项

---

## Phase 4: DevTools 集成 (Week 7)

### 目标
集成 DevTools，支持组件树查看和性能分析。

### 4.1 DevTools 桥接

#### 任务清单

```markdown
- [ ] 4.1.1 创建 devtools/bridge/ 目录
- [ ] 4.1.2 实现 Fiber 树导出
- [ ] 4.1.3 实现组件检查器
- [ ] 4.1.4 实现性能分析器
- [ ] 4.1.5 实现布局调试器
```

#### 文件结构

```
devtools/bridge/
├── fiber.go              # Fiber 树导出
├── inspector.go          # 组件检查器
├── profiler.go           # 性能分析
└── layout.go             # 布局调试
```

#### 验收标准

- [ ] DevTools 可查看 Fiber 树
- [ ] 可查看组件 Props 和 State
- [ ] 可查看渲染性能指标
- [ ] 可调试布局边界

---

## Phase 5: 文档 + 示例 + 测试 (Week 8)

### 目标
完善文档、示例和测试，确保系统可用性。

### 5.1 文档

#### 任务清单

```markdown
- [ ] 5.1.1 编写 API 参考
- [ ] 5.1.2 编写快速开始指南
- [ ] 5.1.3 编写组件文档
- [ ] 5.1.4 编写 Hooks 文档
- [ ] 5.1.5 编写最佳实践
```

### 5.2 示例

#### 任务清单

```markdown
- [ ] 5.2.1 Hello World 示例
- [ ] 5.2.2 Counter 示例
- [ ] 5.2.3 Todo List 示例
- [ ] 5.2.4 Form 示例
- [ ] 5.2.5 Dashboard 示例
```

#### 文件结构

```
examples/
├── hello/
│   └── main.go
├── counter/
│   └── main.go
├── todo/
│   └── main.go
├── form/
│   └── main.go
└── dashboard/
    └── main.go
```

### 5.3 测试

#### 任务清单

```markdown
- [ ] 5.3.1 完成单元测试（覆盖率 > 80%）
- [ ] 5.3.2 完成集成测试
- [ ] 5.3.3 完成性能基准测试
- [ ] 5.3.4 完成 E2E 测试
```

---

## 风险管理

### 高风险项

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| Hooks 调用规则违反 | 运行时错误 | Lint 检查 + 运行时验证 |
| Fiber 树复杂度 | 维护困难 | 充分文档 + 代码审查 |
| 性能回归 | 用户体验下降 | 基准测试 + 性能监控 |
| API 兼容性 | 迁移困难 | 适配器 + 迁移指南 |

### 中风险项

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 渲染不一致 | 显示错误 | 视觉回归测试 |
| 内存泄漏 | 资源耗尽 | 内存监控 + 清理检查 |
| 并发问题 | 数据竞争 | Go race detector |

---

## 资源分配

### 人员配置（假设）

| 角色 | 人数 | 职责 |
|------|------|------|
| 架构师 | 1 | 设计、代码审查 |
| 核心开发 | 2 | Reconciler、渲染 |
| 组件开发 | 1 | 组件库、API |
| DevTools | 1 | 调试工具 |
| 测试工程师 | 1 | 测试、质量保证 |

### 时间分配

| 阶段 | 工作量 | 缓冲 |
|------|--------|------|
| Phase 1 | 10 天 | 2 天 |
| Phase 2 | 10 天 | 2 天 |
| Phase 3 | 10 天 | 2 天 |
| Phase 4 | 5 天 | 1 天 |
| Phase 5 | 5 天 | 1 天 |
| **总计** | **40 天** | **8 天** |

---

## 交付物清单

### 代码交付物

- [ ] ui/ 包（声明式 API）
- [ ] framework/reconciler/（Reconciler 系统）
- [ ] framework/hooks/（Hooks 系统）
- [ ] framework/components/（组件库）
- [ ] framework/render/（渲染管线）
- [ ] devtools/bridge/（DevTools 桥接）

### 文档交付物

- [ ] API 参考文档
- [ ] 快速开始指南
- [ ] 迁移指南
- [ ] 最佳实践文档
- [ ] 示例代码

### 测试交付物

- [ ] 单元测试套件
- [ ] 集成测试套件
- [ ] 性能基准测试
- [ ] E2E 测试套件

---

## 成功标准

### 功能完整性

- [ ] 所有基础 Hooks 可用
- [ ] 所有基础组件可用
- [ ] 声明式 API 完整
- [ ] DevTools 功能完整

### 性能指标

- [ ] 渲染帧率 ≥ 60 FPS
- [ ] 布局计算 < 1 ms
- [ ] 支持 10,000+ 组件
- [ ] 内存占用 < 50 MB

### 质量标准

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 所有示例正常运行
- [ ] 无已知 critical bug
- [ ] 文档完整

---

## 下一步行动

1. **本周**：启动 Phase 1，创建 VNode 系统
2. **代码审查**：每次 PR 需要审查
3. **每日同步**：团队进度同步会议
4. **每周回顾**：检查里程碑完成情况

---

**文档结束**

**版本历史**:
- v1.0 (2026-01-31): 初始版本
