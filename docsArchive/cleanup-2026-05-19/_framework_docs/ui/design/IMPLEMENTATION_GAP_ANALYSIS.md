# Mint UI 声明式架构 - 实现差距分析报告

**文档版本**: v1.0
**分析日期**: 2026-01-31
**目标文档**: SYSTEM_ARCHITECTURE.md v0.1
**分析范围**: framework, runtime, devtools

---

## 执行摘要

本报告详细对比了 Mint UI 的新架构设计文档与当前系统实现，识别出可复用、需改造和需新建的功能模块。

### 关键发现

| 分类 | 模块数量 | 说明 |
|------|---------|------|
| **可直接复用** | 15 | 基础设施成熟，无需修改 |
| **需要改造** | 12 | 架构适配，API 调整 |
| **需要新建** | 10 | 核心新特性 |

---

## 第一部分：可复用功能 (复用率 ~60%)

### 1. Runtime 层 - 高复用价值

#### 1.1 平台抽象层 (platform/) ✅ 完全复用

**当前实现**:
```go
type RuntimePlatform interface {
    Init() error
    Close() error
    Size() (width, height int)
    ReadInput() *RawInput
    WriteString(s string) (int, error)
    Clear() error
}
```

**结论**: 完全符合目标架构需求，无需修改。

#### 1.2 样式系统 (style/) ✅ 完全复用

**当前功能**:
- Color 结构 (RGB, ANSI 256, TrueColor)
- Style 结构 (Bold, Italic, Underline, Strikethrough)
- 样式修饰和合并

**结论**: 架构兼容，API 可直接使用。

#### 1.3 渲染系统 (paint/) ✅ 完全复用

**当前实现**:
```go
type Buffer struct {
    Width  int
    Height int
    Cells  [][]Cell
}

type Cell struct {
    Cluster        string
    Style          style.Style
    Width          int
    IsContinuation bool
    ZIndex         int
    NodeID         string
}
```

**目标需求**:
```go
type Cell struct {
    Char  rune
    Fg    Color
    Bg    Color
    Style Style
}

type Buffer struct {
    Width  int
    Height int
    Cells  [][]Cell
}
```

**差异分析**:
- 当前实现更丰富（支持宽字符、Z-Index、NodeID）
- 可通过适配器层简化为目标 API

**结论**: 复用核心，添加适配器。

#### 1.4 布局引擎 (layout/) ✅ 部分复用

**当前功能**:
- BoxConstraints 约束驱动布局
- Measure Phase / Layout Phase
- Flexbox 布局支持
- 布局缓存

**与目标对比**:

| 特性 | 当前实现 | 目标需求 | 状态 |
|------|---------|---------|------|
| 约束驱动 | ✅ | ✅ | 完全匹配 |
| Flexbox | ✅ | ✅ | 完全匹配 |
| HStack/VStack | ❌ | ✅ | 需封装 |
| 虚拟化 | ❌ | ✅ | 需新建 |

**结论**: 核心算法复用，添加 HStack/VStack 封装。

#### 1.5 焦点管理 (focus/) ✅ 完全复用

**当前实现**:
- FocusManager V3
- 焦点域 (Scope) 支持
- 焦点陷阱 (Trap)
- 几何感知导航

**结论**: 功能完备，无需修改。

#### 1.6 输入处理 (input/) ✅ 部分复用

**当前实现**:
- RawInput 解析
- KeyMap 快捷键映射
- 终端输入序列处理

**差异**: Action 系统与声明式事件系统需桥接

**结论**: 复用底层解析，上层适配。

---

### 2. Framework 层 - 中等复用价值

#### 2.1 组件基础能力 (component/capabilities.go) ✅ 复用接口设计

**当前接口**:
```go
type Node interface {
    ID() string
    Type() string
}

type Measurable interface {
    Measure(maxWidth, maxHeight int) (width, height int)
}

type Paintable interface {
    Paint(ctx PaintContext, buf *paint.Buffer)
}

type Focusable interface {
    FocusID() string
    OnFocus()
    OnBlur()
}
```

**结论**: 接口设计优秀，可作为新架构基础。

#### 2.2 容器组件 (component/container.go) ✅ 部分复用

**当前实现**: BaseContainer
- 子节点管理
- 布局支持
- 上下文注入

**改造需求**: 适配声明式 VNode 树

**结论**: 内核复用，API 重构。

#### 2.3 主题系统 (theme/) ✅ 完全复用

**当前实现**:
- ColorPalette
- StyleConfig
- 主题继承和扩展

**结论**: 满足目标需求，无需修改。

#### 2.4 数据绑定 (binding/) ⚠️ 需评估

**当前实现**:
- ReactiveStore
- 依赖追踪
- 表达式求值

**与目标状态系统的关系**: 可能存在功能重叠

**结论**: 需要与新的 useState 系统一体化设计。

---

### 3. DevTools 层 - 高复用价值

#### 3.1 协议服务器 (protocol/) ✅ 完全复用

**当前实现**:
- 统一协议服务器
- WebSocket 支持
- HTTP API
- 消息定义

**目标需求**: 远程渲染协议

**结论**: 基础设施完备，添加 DrawCmd 流式支持。

#### 3.2 事件总线 (core/bus.go) ✅ 完全复用

**当前实现**:
- 原子操作 + 环形缓冲区
- 零拷贝设计
- 多订阅者支持

**结论**: 性能优秀，可直接复用。

#### 3.3 快照系统 (snapshot/) ✅ 完全复用

**当前实现**:
- 对象池模式
- 持久化支持
- 增量快照

**结论**: 功能完备，无需修改。

#### 3.4 性能分析 (memory/, observation/) ✅ 完全复用

**当前实现**:
- 内存监控
- 性能采样
- 模式检测

**结论**: 满足 DevTools 需求。

---

## 第二部分：需要改造的功能

### 1. 核心架构改造

#### 1.1 组件系统改造 (最优先)

**当前状态**: 命令式组件
```go
// 当前实现
text := NewText("Hello")
text.SetPosition(1, 1)
text.SetColor(FgRed)
container.Add(text)
```

**目标状态**: 声明式组件
```go
// 目标实现
func App() VNode {
    return ui.HStack(
        ui.Text("Hello").FgColor(FgRed),
        ui.Text("World").FgColor(FgBlue),
    )
}
```

**改造方案**:

1. **创建 VNode 抽象层**
```go
// 新建: ui/vnode.go
type VNode interface {
    Type() VNodeType
    Children() []VNode
    Props() Props
    Key() string
}

type VNodeType int
const (
    VNodeComponent VNodeType = iota
    VNodeElement
    VNodeText
    VNodeFragment
)
```

2. **适配现有组件**
```go
// 改造: component/base.go
func (c *BaseComponent) ToVNode() VNode {
    return &ElementVNode{
        component: c,
        props:     c.props,
        children:  c.children,
    }
}
```

3. **创建声明式 Builder**
```go
// 新建: ui/builder.go
func Text(text string) *TextBuilder {
    return &TextBuilder{content: text}
}

func (b *TextBuilder) FgColor(c Color) *TextBuilder {
    b.props["fg"] = c
    return b
}

func (b *TextBuilder) Build() VNode {
    // 转换为内部组件
}
```

#### 1.2 事件系统改造

**当前状态**: Runtime 三阶段事件传播
**目标状态**: 声明式事件绑定

**改造方案**:

1. **保留核心分发逻辑**
2. **添加声明式事件 API**
```go
// 新建: ui/events.go
func (n *VNode) OnClick(handler func(Event)) *VNode {
    n.props["onClick"] = handler
    return n
}

func (n *VNode) OnClickCapture(handler func(Event)) *VNode {
    n.props["onClickCapture"] = handler
    return n
}
```

3. **桥接层连接两个系统**
```go
// 改造: event/dispatch.go
func dispatchVNodeEvent(vnode VNode, event Event) {
    if handler, ok := vnode.Props()["onClick"]; ok {
        handler.(func(Event))(event)
    }
}
```

#### 1.3 布局 API 改造

**当前状态**: 直接使用 FlexLayout
**目标状态**: HStack/VStack 声明式 API

**改造方案**:

```go
// 新建: ui/layout.go
func HStack(children ...VNode) VNode {
    return Element("hstack").Children(children...)
}

func VStack(children ...VNode) VNode {
    return Element("vstack").Children(children...)
}

// 内部使用现有 FlexLayout
func (e *ElementVNode) Measure(constraints Constraint) Size {
    switch e.tag {
    case "hstack":
        return runtime.FlexLayoutRow(e, constraints)
    case "vstack":
        return runtime.FlexLayoutColumn(e, constraints)
    }
}
```

---

### 2. 渲染管线改造

#### 2.1 添加 DrawCmd 抽象

**当前状态**: 直接调用 Paint()
**目标状态**: DrawCmd 中间表示

**改造方案**:

```go
// 新建: render/drawcmd.go
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

// 改造: paint/buffer.go
func (b *Buffer) Execute(cmds []DrawCmd) {
    for _, cmd := range cmds {
        switch c := cmd.(type) {
        case *DrawText:
            b.drawText(c.X, c.Y, c.Text, c.Style)
        case *DrawRect:
            b.drawRect(c.X, c.Y, c.W, c.H, c.Style)
        case *DrawClip:
            b.setClip(c.X, c.Y, c.W, c.H)
        }
    }
}
```

#### 2.2 Buffer Diff 优化

**当前状态**: 全量刷新或脏区域
**目标状态**: Cell 级别的 Diff

**新建**: render/diff.go
```go
func DiffBuffer(old, new *Buffer) []CellChange {
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

// 优化 ANSI 输出
func ApplyChanges(changes []CellChange) string {
    // 合并相同 style 的连续 cell
    // 减少 ANSI 切换
}
```

---

### 3. 状态系统一体化

#### 3.1 合并状态管理方案

**当前状态**:
- framework/binding: ReactiveStore
- framework/component: StateHolder

**目标状态**: useState Hook

**一体化方案**:

```go
// 新建: state/hooks.go
type HookContext struct {
    componentID string
    hooks       []Hook
    hookIndex   int
}

var (
    hookContexts sync.Map // map[string]*HookContext
    hookMutex    sync.Mutex
)

func useState(initial interface{}) (interface{}, func(interface{})) {
    ctx := currentHookContext()
    idx := ctx.hookIndex
    ctx.hookIndex++

    if idx >= len(ctx.hooks) {
        // 首次创建
        ctx.hooks = append(ctx.hooks, Hook{
            Type:  HookState,
            State: initial,
        })
    }

    hook := &ctx.hooks[idx]

    // 返回值和 setter
    setState := func(newValue interface{}) {
        hook.State = newValue
        scheduleRender(ctx.componentID)
    }

    return hook.State, setState
}

// 复用现有 ReactiveStore 用于全局状态
func useStore(store *ReactiveStore) interface{} {
    // ...
}
```

---

## 第三部分：需要新建的功能

### 1. Reconciler 系统 (新建)

**优先级**: 🔴 最高

**文件结构**:
```
internal/reconciler/
├── diff.go        # Diff 算法
├── fiber.go       # Fiber 节点
├── scheduler.go   # 调度器
└── workloop.go    # 工作循环
```

#### 1.1 Diff 算法

```go
// 新建: internal/reconciler/diff.go
func Diff(old, new VNode) Patch {
    switch {
    case old == nil && new != nil:
        return &CreatePatch{Node: new}
    case old != nil && new == nil:
        return &DeletePatch{Node: old}
    case old.Type() != new.Type():
        return &ReplacePatch{Old: old, New: new}
    case old.Key() != new.Key():
        return &ReplacePatch{Old: old, New: new}
    default:
        return diffChildren(old, new)
    }
}

func diffChildren(old, new VNode) Patch {
    oldChildren := old.Children()
    newChildren := new.Children()

    // 使用双指针算法优化
    patches := make([]Patch, 0)
    // ... diff 逻辑
    return &ChildrenPatch{Patches: patches}
}
```

#### 1.2 Fiber 架构

```go
// 新建: internal/reconciler/fiber.go
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
    Passive
)

func (f *Fiber) BeginWork() {
    // 处理组件开始阶段
}

func (f *Fiber) CompleteWork() {
    // 完成工作，标记 effects
}
```

#### 1.3 Scheduler 调度器

```go
// 新建: internal/reconciler/scheduler.go
type Lane uint64

const (
    SyncLane Lane = 0b00000001
    InputLane Lane = 0b00000010
    AnimationLane Lane = 0b00000100
    TransitionLane Lane = 0b00001000
    IdleLane Lane = 0b10000000
)

type Scheduler struct {
    taskQueue  []*Task
    lanes      Lanes
    running    bool
}

type Task struct {
    Callback func()
    Lanes    Lanes
}

func (s *Scheduler) Schedule(callback func(), lanes Lanes) {
    task := &Task{
        Callback: callback,
        Lanes:    lanes,
    }
    s.taskQueue = append(s.taskQueue, task)
}

func (s *Scheduler) WorkLoop(deadline time.Time) {
    for {
        if time.Now().After(deadline) {
            break // 时间片用完
        }
        s.performUnitOfWork()
    }
}
```

---

### 2. Hooks 系统 (新建)

**优先级**: 🔴 最高

**文件结构**:
```
internal/hooks/
├── state.go       # useState
├── effect.go      # useEffect
├── context.go     # useContext
├── memo.go        # useMemo, useCallback
├── ref.go         # useRef
└── context.go     # Context 实现
```

#### 2.1 useState

```go
// 新建: internal/hooks/state.go
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

#### 2.2 useEffect

```go
// 新建: internal/hooks/effect.go
func useEffect(effect func(), deps []interface{}) {
    ctx := currentContext()
    hook := ctx.nextHook()

    if !depsEqual(hook.Deps, deps) {
        hook.Deps = deps
        hook.Effect = effect
        ctx.registerEffect(hook)
    }
}

func (h *Hook) cleanup() {
    if h.Cleanup != nil {
        h.Cleanup()
    }
}
```

#### 2.3 useContext

```go
// 新建: internal/hooks/context.go
type Context struct {
    value interface{}
}

var contexts sync.Map

func useContext(ctx *Context) interface{} {
    current := currentContext()
    return current.lookupContext(ctx)
}

func ProvideContext(ctx *Context, value interface{}, child VNode) VNode {
    // 创建 provider 节点
}
```

---

### 3. 虚拟化渲染 (新建)

**优先级**: 🟡 中

**文件结构**:
```
internal/layout/
└── virtual.go
```

```go
// 新建: internal/layout/virtual.go
type VirtualList struct {
    Items      []interface{}
    ItemHeight int
    RenderItem func(interface{}) VNode

    scrollTop  int
    viewportHeight int
}

func (vl *VirtualList) Measure(constraints Constraint) Size {
    return constraints.Constrain(Size{
        Width:  constraints.MaxWidth,
        Height: vl.viewportHeight,
    })
}

func (vl *VirtualList) Build() VNode {
    // 计算可见范围
    start := vl.scrollTop / vl.ItemHeight
    visibleCount := vl.viewportHeight / vl.ItemHeight + 2
    end := min(start + visibleCount, len(vl.Items))

    // 只渲染可见项
    children := make([]VNode, 0)
    for i := start; i < end; i++ {
        item := vl.Items[i]
        child := vl.RenderItem(item)
        child.SetKey(fmt.Sprintf("item-%d", i))
        children = append(children, child)
    }

    return VStack(children...)
}
```

---

### 4. 动画系统 (新建)

**优先级**: 🟡 中

**注**: runtime/animation 已存在，需扩展

**扩展示例**:

```go
// 扩展: animation/hooks.go
func useAnimation(config AnimationConfig) (float64, func()) {
    value, setValue := useState(config.From)

    useEffect(func() {
        timeline := animation.NewTimeline(config.Duration, config.Easing)
        timeline.OnUpdate(func(v float64) {
            setValue(v)
        })
        timeline.Play()

        cleanup := func() {
            timeline.Cancel()
        }
        return cleanup
    }, []interface{}{config})

    return value.(float64), func() {
        // 取消动画
    }
}
```

---

### 5. SDK 层 (新建)

**优先级**: 🔴 最高

**文件结构**:
```
ui/
├── app.go         # Run 函数
├── vnode.go       # VNode 类型
├── builder.go     # 组件构建器
├── layout.go      # HStack, VStack
├── hooks.go       # Hooks 导出
└── components.go  # 内置组件
```

```go
// 新建: ui/app.go
func Run(app func() VNode, opts ...Option) error {
    options := defaultOptions()
    for _, opt := range opts {
        opt(&options)
    }

    // 创建运行时
    rt := runtime.New(options.width, options.height)

    // 创建 Fiber 根节点
    root := reconciler.CreateRoot(app())

    // 主循环
    for {
        // 调度器工作循环
        reconciler.WorkLoop(5 * time.Millisecond)

        // 渲染
        rt.Render(root)
    }
}

// 新建: ui/hooks.go
func UseState = hooks.useState
func UseEffect = hooks.useEffect
func UseContext = hooks.useContext
// ...
```

---

## 第四部分：实施优先级路线图

### Phase 1: 核心基础 (Week 1-2)

| 任务 | 文件 | 依赖 | 优先级 |
|------|------|------|--------|
| VNode 抽象 | ui/vnode.go | - | P0 |
| Builder API | ui/builder.go | VNode | P0 |
| 基础 Hooks | internal/hooks/*.go | VNode | P0 |
| Diff 算法 | internal/reconciler/diff.go | VNode | P0 |
| Fiber 节点 | internal/reconciler/fiber.go | Diff | P0 |

### Phase 2: 渲染集成 (Week 3-4)

| 任务 | 文件 | 依赖 | 优先级 |
|------|------|------|--------|
| DrawCmd 抽象 | render/drawcmd.go | - | P1 |
| Buffer Diff | render/diff.go | Buffer | P1 |
| Paint 适配 | render/adapt.go | DrawCmd | P1 |
| 布局 API | ui/layout.go | runtime/layout | P1 |

### Phase 3: 高级特性 (Week 5-6)

| 任务 | 文件 | 依赖 | 优先级 |
|------|------|------|--------|
| Scheduler | internal/reconciler/scheduler.go | Fiber | P1 |
| 时间切片 | internal/reconciler/workloop.go | Scheduler | P1 |
| 虚拟化 | internal/layout/virtual.go | Layout | P2 |
| 动画 Hooks | animation/hooks.go | Hooks | P2 |

### Phase 4: DevTools 集成 (Week 7-8)

| 任务 | 文件 | 依赖 | 优先级 |
|------|------|------|--------|
| 组件树导出 | devtools/tree.go | Fiber | P2 |
| 性能监控 | devtools/profiler.go | Scheduler | P2 |
| 布局调试 | devtools/layout.go | Layout | P3 |

---

## 第五部分：复用策略总结

### 直接复用 (15 模块)

| 模块 | 路径 | 说明 |
|------|------|------|
| 平台抽象 | runtime/platform | 完全兼容 |
| 样式系统 | runtime/style | 完全兼容 |
| 焦点管理 | runtime/focus | 完全兼容 |
| Flex 布局 | runtime/layout | 核心算法 |
| 事件传播 | runtime/event | 三阶段模型 |
| Action 系统 | runtime/action | 语义化事件 |
| 输入处理 | runtime/input | 底层解析 |
| 状态追踪 | runtime/state | 快照/回放 |
| 绘制缓冲 | runtime/paint | 核心渲染 |
| 组件接口 | framework/component | 能力模型 |
| 主题系统 | framework/theme | 样式管理 |
| 事件总线 | devtools/bus | 高性能通信 |
| 协议服务 | devtools/protocol | WebSocket/API |
| 快照系统 | devtools/snapshot | 时间旅行 |
| 内存监控 | devtools/memory | 性能分析 |

### 适配改造 (12 模块)

| 模块 | 改造方式 | 工作量 |
|------|---------|--------|
| 组件系统 | 添加 VNode 适配层 | 中 |
| 容器组件 | 声明式 API 封装 | 小 |
| 布局引擎 | HStack/VStack 封装 | 小 |
| 事件系统 | 声明式事件绑定 | 小 |
| 渲染管线 | DrawCmd 中间层 | 中 |
| 状态管理 | Hooks 一体化 | 大 |
| 数据绑定 | 与 useState 合并 | 中 |
| 主题应用 | 声明式 API | 小 |
| 远程渲染 | DrawCmd 流式 | 中 |
| DevTools API | Fiber 树导出 | 中 |
| 性能分析 | 调度器集成 | 中 |
| 测试框架 | 组件测试 | 中 |

### 新建模块 (10 模块)

| 模块 | 优先级 | 复用基础 |
|------|--------|---------|
| VNode 系统 | P0 | runtime.Node |
| Diff 算法 | P0 | - |
| Fiber 架构 | P0 | runtime.Node |
| Hooks 系统 | P0 | framework/binding |
| Scheduler | P1 | runtime/scheduler |
| DrawCmd | P1 | runtime/paint |
| Buffer Diff | P1 | runtime/paint |
| 声明式 API | P0 | framework/component |
| 虚拟化 | P2 | runtime/layout |
| 动画 Hooks | P2 | runtime/animation |

---

## 第六部分：风险评估与缓解措施 (增强版)

### 高风险项

#### 1. 状态系统合并

| 维度 | 详情 |
|------|------|
| **风险描述** | ReactiveStore 与 useState 功能重叠，可能导致开发者困惑 |
| **影响等级** | 🔴 高 |
| **发生概率** | 中 |

**详细缓解措施**：

```go
// 1. 明确职责边界（已在 API_DESIGN.md 文档化）
//    - useState: 组件本地状态
//    - ReactiveStore: 全局/跨组件状态

// 2. 编译时提示（通过代码注释）
// useState 适用于：UI 状态、表单输入、动画状态
// ReactiveStore 适用于：用户数据、应用配置、缓存数据

// 3. Lint 规则建议
// 检测在多个组件中重复定义相同状态的模式
```

**验证标准**：
- [ ] 文档明确分工（已完成 ✅）
- [ ] 示例代码展示正确用法
- [ ] 代码审查检查清单包含状态选择

---

#### 2. Fiber 架构复杂度

| 维度 | 详情 |
|------|------|
| **风险描述** | Fiber 树管理复杂，指针操作容易出错，调试困难 |
| **影响等级** | 🔴 高 |
| **发生概率** | 高 |

**详细缓解措施**：

```go
// 1. 不变性检查
type Fiber struct {
    // 添加调试标记
    debugID    string    // 唯一标识
    createTime time.Time // 创建时间
    updateCount int      // 更新次数
}

// 2. 树完整性验证
func (f *Fiber) ValidateTree() error {
    // 检查父子关系一致性
    if f.Child != nil && f.Child.Return != f {
        return fmt.Errorf("parent-child mismatch at %s", f.debugID)
    }
    // 检查兄弟链表完整性
    // ...
    return nil
}

// 3. 开发模式增强日志
func (f *Fiber) DebugLog(phase string) {
    if isDevMode() {
        log.Printf("[Fiber %s] %s: type=%v, effect=%v",
            f.debugID, phase, f.VNode.Type(), f.EffectTag)
    }
}
```

**测试策略**：
- [ ] 单元测试覆盖所有 Fiber 操作
- [ ] 模糊测试随机生成树结构
- [ ] 集成测试验证完整渲染周期
- [ ] 内存泄漏检测

---

#### 3. Hooks 调用规则

| 维度 | 详情 |
|------|------|
| **风险描述** | Go 无法静态检查，顺序调用约束容易被违反 |
| **影响等级** | 🔴 高 |
| **发生概率** | 高 |

**详细缓解措施**（已在 SYSTEM_ARCHITECTURE.md 实现）：

```go
// 1. 运行时验证器（已实现 ✅）
// 见 SYSTEM_ARCHITECTURE.md 中的 HookValidator

// 2. 开发模式堆栈追踪
// 见 SYSTEM_ARCHITECTURE.md 中的 DevModeValidator

// 3. 错误信息增强
// 提供具体的错误位置和修复建议

// 4. 文档和示例
// 常见错误模式和正确写法对比
```

**验证标准**：
- [x] 运行时验证器实现（已完成 ✅）
- [ ] 错误信息包含修复建议
- [ ] 文档包含常见错误示例
- [ ] 测试覆盖所有错误场景

---

### 中风险项

#### 4. 性能回归

| 维度 | 详情 |
|------|------|
| **风险描述** | 新增抽象层（VNode、Fiber、DrawCmd）可能影响性能 |
| **影响等级** | 🟡 中 |
| **发生概率** | 中 |

**详细缓解措施**（已在 BENCHMARK.md 实现）：

```bash
# 1. 建立性能基准（已完成 ✅）
go test ./framework/benchmark/... -bench=. -benchmem

# 2. CI 集成性能检测
# 每次 PR 运行基准测试，与基准线比较

# 3. 性能阈值告警
# 超过 10% 回归自动标记 PR

# 4. 热点优化
# 使用 pprof 分析 CPU 和内存热点
go tool pprof cpu.prof
```

**验证标准**：
- [x] 基准测试套件完成（已完成 ✅）
- [ ] CI 集成性能检测
- [ ] 性能回归自动告警
- [ ] 关键路径优化完成

---

#### 5. API 兼容性

| 维度 | 详情 |
|------|------|
| **风险描述** | 新旧 API 共存期间可能造成开发者困惑 |
| **影响等级** | 🟡 中 |
| **发生概率** | 中 |

**详细缓解措施**：

```go
// 1. 废弃标记
// Deprecated: Use ui.Text instead
func NewText(content string) *Text {
    log.Println("DEPRECATED: NewText is deprecated, use ui.Text")
    return &Text{content: content}
}

// 2. 适配器层
// framework/compat/adapter.go
func LegacyToVNode(legacy *component.Text) ui.VNode {
    return ui.Text(legacy.Content()).
        FgColor(legacy.FgColor()).
        BgColor(legacy.BgColor())
}

// 3. 迁移工具
// tools/migrate/main.go
// 自动转换旧代码到新 API
```

**迁移时间表**：
| 阶段 | 时间 | 行动 |
|------|------|------|
| 共存期 | v0.1-v0.3 | 新旧 API 并存，废弃警告 |
| 过渡期 | v0.4-v0.5 | 默认新 API，旧 API 需显式启用 |
| 移除期 | v1.0 | 移除旧 API |

---

#### 6. 布局错误处理

| 维度 | 详情 |
|------|------|
| **风险描述** | Grid/Absolute 布局边界情况可能导致崩溃或显示异常 |
| **影响等级** | 🟡 中 |
| **发生概率** | 中 |

**详细缓解措施**（已在 GRID_LAYOUT_DESIGN.md 实现）：

```go
// 1. 容错模式（已实现 ✅）
// RecoveryStrict / RecoveryClamp / RecoveryExpand

// 2. 错误可视化（已实现 ✅）
// 开发模式显示布局错误边界

// 3. 日志记录
// 记录所有布局警告供调试
```

**验证标准**：
- [x] 容错策略实现（已完成 ✅）
- [x] 错误可视化实现（已完成 ✅）
- [ ] 测试覆盖所有边界情况

---

### 低风险项

#### 7. 内存泄漏

| 维度 | 详情 |
|------|------|
| **风险描述** | Hooks Effect 清理不当、Fiber 节点未释放 |
| **影响等级** | 🟢 低 |
| **发生概率** | 低 |

**缓解措施**：

```go
// 1. Effect 清理追踪
type Effect struct {
    cleanup   func()
    cleaned   bool
    debugInfo string
}

func (e *Effect) Cleanup() {
    if e.cleanup != nil && !e.cleaned {
        e.cleanup()
        e.cleaned = true
    }
}

// 2. Fiber 节点池
var fiberPool = sync.Pool{
    New: func() interface{} {
        return &Fiber{}
    },
}

// 3. 定期内存检查（开发模式）
func StartMemoryMonitor() {
    go func() {
        for range time.Tick(10 * time.Second) {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            if m.HeapAlloc > threshold {
                log.Printf("[Memory Warning] Heap: %d MB", m.HeapAlloc/1024/1024)
            }
        }
    }()
}
```

---

### 风险矩阵总览

```
影响 ↑
高   │ ③Hooks规则  ②Fiber复杂度
     │ ①状态合并
中   │ ④性能回归   ⑤API兼容   ⑥布局错误
     │
低   │             ⑦内存泄漏
     └──────────────────────────────────→ 概率
         低           中           高
```

---

## 第七部分：结论

### 复用率评估

- **Runtime 层**: 75% 可复用
- **Framework 层**: 50% 可复用
- **DevTools 层**: 80% 可复用
- **整体**: ~60% 可复用

### 关键建议

1. **分阶段实施**: 先完成 VNode + Hooks + Diff，再扩展高级特性
2. **保持兼容**: 新旧 API 共存一段时间，提供迁移工具
3. **测试先行**: 每个新模块都应有对应的测试
4. **文档完善**: API 文档、迁移指南、示例代码

### 已完成的改进 (v1.1)

| 改进项 | 文档 | 状态 |
|--------|------|------|
| MVP 优先策略 | IMPLEMENTATION_PLAN.md | ✅ 完成 |
| Hooks 运行时验证 | SYSTEM_ARCHITECTURE.md | ✅ 完成 |
| useState vs ReactiveStore 分工 | API_DESIGN.md | ✅ 完成 |
| 性能基准测试结果 | BENCHMARK.md | ✅ 完成 |
| Grid 容错策略 | GRID_LAYOUT_DESIGN.md | ✅ 完成 |
| 风险缓解措施细化 | 本文档 | ✅ 完成 |

### 下一步行动

1. ~~创建详细实施计划~~ (`IMPLEMENTATION_PLAN.md`) ✅
2. ~~建立 API 设计文档~~ (`API_DESIGN.md`) ✅
3. 准备迁移指南 (`MIGRATION_GUIDE.md`)
4. ~~设置性能基准~~ (`BENCHMARK.md`) ✅
5. 开始 Phase 0: MVP 核心实现

---

**文档结束**

**版本历史**:
- v1.1 (2026-01-31): 细化风险缓解措施，添加已完成改进清单
- v1.0 (2026-01-31): 初始版本
