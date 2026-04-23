# Mint UI 全面重构计划

> **创建日期**: 2026-02-01
> **更新日期**: 2026-02-02
> **状态**: 📋 规划中 (已更新类型基础包 runtime/ui)
> **目标**: 构建符合设计愿景的 Terminal UI Runtime Platform
> **参考文档**: framework/docs/ui/idea/*.md, docs/plan/REFACTOR_TODO.md

---

## 目录

1. [愿景与现状对比](#一愿景与现状对比)
2. [架构分层设计](#二架构分层设计)
3. [组件解耦方案](#三组件解耦方案)
4. [核心模块重构](#四核心模块重构)
5. [实施路线图](#五实施路线图)
6. [验证标准](#六验证标准)

---

## 一、愿景与现状对比

### 1.1 设计愿景（来自 idea 文档）

根据 `framework/docs/ui/idea/idea7_final.md`，目标架构是：

```
Application Layer     ← 业务逻辑 / 页面 / 组件组合
    ↓
Declarative UI Layer  ← VNode / Hooks / State / Props
    ↓
Reconciler Layer      ← Diff / Fiber / Scheduler
    ↓
Layout Engine         ← Constraint Layout / Flex / Virtualized
    ↓
Render Engine         ← DrawCmd / Clip / Transform
    ↓
Animation Subsystem   ← Timeline / Easing / Physics
    ↓
Rasterizer            ← DrawCmd → Cells
    ↓
Dirty Region System   ← Diff Cells / Rect Merge
    ↓
Terminal Backend      ← ANSI Driver / Input / Resize
```

### 1.2 当前架构现状

```
ui/app.go (2169行)
    ├── declarativeRoot (单一根组件)
    │   ├── reconciler (Fiber协调器)
    │   ├── instanceManager
    │   ├── buttons/inputs/... (集中收集的交互元素)
    │   └── focusedIndex (全局焦点)
    └── renderVNode() (直接写Buffer)
```

### 1.3 差距分析

| 维度 | 设计愿景 | 当前实现 | 差距 |
|------|---------|---------|------|
| **类型基础** | 独立类型包 | 分散在 ui/ | ✅ 已创建 runtime/ui |
| **组件契约** | Render/Measure/Paint 分离 | VNode 没有这些方法 | ⚠️ 需补充 |
| **渲染方式** | DrawCmd → Rasterizer | 直接写 Buffer | ⚠️ 部分实现 |
| **组件组织** | 按功能分类 (basic/form/button) | 全在 ui/ 目录 | ✅ components/ 已组织 |
| **状态管理** | Hooks Slot (位置绑定) | 部分在 VNode 字段 | ⚠️ 需统一 |
| **多组件支持** | 可独立渲染 | 单一 declarativeRoot | ⚠️ 需解耦 |
| **动画系统** | 与状态分离 | 未分离 | ⚠️ 需新增 |

---

## 二、架构分层设计

### 2.1 目标目录结构

```
mint/
├── cmd/                            # 可执行程序
│   └── examples/                   # 示例入口
│
├── ui/                             # 公开 API 层 (精简，重导出 runtime/ui)
│   ├── app.go                      # Run() 入口 ~100行
│   ├── hooks.go                    # Hooks API 实现
│   ├── vnode.go                    # VNode 接口重导出
│   ├── element.go                  # ElementVNode 重导出
│   ├── component.go                # ComponentVNode 重导出
│   ├── fragment.go                  # FragmentVNode 重导出
│   ├── layout.go                    # LayoutNode 重导出
│   ├── fiber.go                     # Fiber 类型重导出
│   ├── instance.go                  # ComponentInstance 重导出
│   ├── validator.go                 # HookValidator 重导出
│   ├── compat.go                    # 兼容性存根 (临时)
│   └── shortcuts.go                # 快捷函数
│
├── runtime/                        # 运行时层 (类型定义)
│   ├── types/                       # ✨ 新增: 核心类型包
│   │   ├── vnode.go                # VNode 接口, VNodeType, Props
│   │   ├── element.go              # ElementVNode, ElementBuilder
│   │   ├── component.go            # ComponentVNode, ComponentBuilder
│   │   ├── fragment.go              # FragmentVNode
│   │   ├── fiber.go                 # Lane, Fiber, EffectFlag
│   │   ├── fiber_util.go            # CreateFiber, CloneFiber, 等
│   │   ├── hooks.go                 # ComponentContext, Hook, Ref
│   │   ├── instance.go              # ComponentInstance, BaseComponentInstance
│   │   ├── validator.go             # HookValidator, HookOrderError
│   │   └── layout.go                # LayoutNode, Direction, Align
│   │
│   ├── paint/                       # 渲染接口
│   │   ├── paintable.go             # Paintable 接口
│   │   └── batch.go                 # DrawCmd
│   │
│   ├── layout/                      # 布局引擎
│   │   ├── constraints.go           # BoxConstraints
│   │   └── size.go                  # Size
│   │
│   ├── style/                       # 样式系统
│   ├── event/                       # 事件系统
│   └── platform/                    # 平台抽象
│
├── components/                     # 声明式组件库 (新增)
│   ├── basic/                      # 基础组件
│   │   ├── text.go                 # Text 组件
│   │   ├── divider.go              # Divider 组件
│   │   └── ...
│   ├── layout/                     # 布局组件
│   ├── form/                       # 表单组件
│   ├── button/                     # 按钮组件
│   ├── feedback/                   # 反馈组件
│   ├── data/                       # 数据展示
│   ├── navigation/                 # 导航组件
│   ├── overlay/                    # 覆盖层组件
│   └── container/                  # 容器组件
│
├── internal/                       # 内部实现 (不对外暴露)
│   ├── reconciler/                 # 协调器系统
│   │   ├── fiber.go                # 已迁移到 runtime/ui
│   │   ├── reconciler.go           # Reconciler 实现
│   │   ├── vnode_converter.go      # VNode → LayoutNode 转换
│   │   └── public.go               # 公开接口
│   │
│   ├── scheduler/                  # 调度器
│   │   └── priority.go
│   │
│   └── state/                      # 状态系统
│       ├── instance_manager.go
│       └── interaction_state.go
│
├── framework/                      # 框架层 (保持)
│   ├── app.go
│   ├── component/
│   └── event/
│
└── docs/                           # 文档
```

### 2.2 依赖关系图（单向，无循环）

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           单向依赖关系图 (已更新 runtime/ui)                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  用户代码 (examples/)                                                         │
│      ↓ import                                                                    │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │ ui/ (API 入口层 - 精简，重导出 runtime/ui)                                   │    │
│  │ ├── vnode.go              # VNode = types.VNode (重导出)                   │    │
│  │ ├── element.go            # ElementVNode = types.ElementVNode (重导出)       │    │
│  │ ├── component.go           # ComponentVNode = types.ComponentVNode (重导出)   │    │
│  │ ├── hooks.go               # UseState, UseEffect (实现)                       │    │
│  │ ├── app.go                 # Run() 入口                                              │    │
│  │ └── compat.go              # 兼容性存根类型 (临时)                           │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│      ↓ import (重导出 + 类型别名)                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │ components/ (组件实现层)                                                   │    │
│  │ ├── basic/text.go         # TextVNode 实现 ui.VNode + Paintable              │    │
│  │ ├── button/button.go     # ButtonVNode 实现 ui.VNode + Paintable            │    │
│  │ └── ...                  # 其他组件同样实现                                   │    │
│  │                                                                              │    │
│  │ 组件实现:                                                                     │  │    │
│  │   - 实现 ui.VNode 接口 (来自 runtime/ui)                                   │  │    │
│  │   - 实现 runtime.Measurable 接口                                            │  │    │
│  │   - 实现 paint.Paintable 接口                                              │  │    │
│  │   - 使用 types.ComponentContext (来自 runtime/ui)                             │  │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│      ↓ import (导入 runtime/ui)                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │ internal/ (内部实现层)                                                       │    │
│  │ ├── reconciler/           # VNodeConverter 使用 types.VNode                    │    │
│  │ ├── scheduler/            # 使用 types.Lane, types.Fiber                       │    │
│  │ └── state/                # 使用 types.ComponentContext                     │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│      ↓ import                                                                    │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │ runtime/ui/ (类型基础层 - 新增)                                          │    │
│  │ ├── vnode.go              # VNode 接口, VNodeType, Props                    │    │
│  │ ├── element.go            # ElementVNode, ElementBuilder                   │    │    │
│  │ ├── component.go            # ComponentVNode, ComponentFunc                 │    │
│  │ ├── fragment.go              # FragmentVNode                                   │    │
│  │ ├── fiber.go                 # Lane, Fiber, EffectFlag, UpdateQueue      │    │
│  │ ├── fiber_util.go            # CreateFiber, CloneFiber, WalkFiber*        │    │
│  │ ├── hooks.go                 # ComponentContext, Hook, Ref, EffectCallback    │    │
│  │ ├── instance.go              # ComponentInstance, BaseComponentInstance    │    │
│  │ ├── validator.go             # HookValidator, HookOrderError                 │    │
│  │ └── layout.go                # LayoutNode, Direction, Align, HStack, VStack  │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│                                                                                  │
│  ✅ 关键设计:                                                                  │
│  - runtime/ui/ 作为类型基础包，所有核心类型定义在此                               │
│  - ui/ 通过类型别名重导出 types，保持向后兼容                                      │
│  - components/ 实现 types.VNode 接口                                                │
│  - internal/ 导入 types，不导入 ui (避免循环)                                      │
│  - 无循环依赖: components/ → runtime/ui ← ui/ (重导出)                             │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 类型分层架构

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                      runtime/ui/ 类型基础包架构                                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │    │
│  │ 1. VNode 系统 (vnode.go)                                                      │  │    │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │  │    │
│  │  │ type VNode interface {                                                     │   │  │    │
│  │  │     Type() VNodeType                                                        │   │  │    │
│  │  │     Props() Props                                                            │   │  │    │
│  │  │     Children() []VNode                                                       │   │  │    │
│  │  │     Key() string                                                             │   │  │    │
│  │  │     Style() style.Style                                                     │   │  │    │
│  │  │ }                                                                            │   │  │    │
│  │  │                                                                             │   │  │    │
│  │  │ type VNodeType int (Element/Text/Component/Fragment)                       │   │  │    │
│  │  │ type Props map[string]interface{}                                             │   │  │    │
│  │  │ type ComponentFunc func() VNode                                              │   │  │    │
│  │  │ type ComponentFuncWithProps func(Props) VNode                                │   │  │    │  │    │
│  │  └─────────────────────────────────────────────────────────────────────┘   │  │    │
│  └───────────────────────────────────────────────────────────────────────────┘  │    │
│                                                                                  │    │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │    │
│  │ 2. 具体节点类型 (element.go, component.go, fragment.go)                               │  │    │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │  │    │
│  │  │ type ElementVNode struct { ... }          // 通用元素节点                │   │  │    │
│  │  │ type ComponentVNode struct { ... }        // 函数组件节点                │   │  │    │
│  │  │ type FragmentVNode struct { ... }         // 片段节点                    │   │  │    │
│  │  │                                                                             │   │  │    │
│  │  │ // Builder 模式支持                                                           │   │  │    │
│  │  │ type ElementBuilder struct { node *ElementVNode }                           │   │  │    │
│  │  │ type ComponentBuilder struct { node *ComponentVNode }                         │   │  │  │    │
│  │  └─────────────────────────────────────────────────────────────────────┘   │  │    │
│  └───────────────────────────────────────────────────────────────────────────┘  │    │
│                                                                                  │    │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │    │
│  │ 3. Fiber 系统 (fiber.go, fiber_util.go)                                           │  │    │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │  │    │
│  │  │ type Lane uint64                           // 优先级车道                    │   │  │    │
│  │  │ type Fiber struct { ... }                  // 工作单元                      │   │  │    │
│  │  │ type EffectFlag int                         // 副作用标志                    │   │  │  │    │
│  │  │                                                                             │   │  │    │
│  │  │ // 工具函数                                                                   │   │  │    │
│  │  │ func CreateFiber(vnode VNode) *Fiber                                      │   │  │    │
│  │  │ func CreateFiberFromVNode(vnode VNode) *Fiber                             │   │  │    │
│  │  │ func CloneFiber(fiber *Fiber) *Fiber                                        │   │  │    │
│  │  │ func WalkFiberDepthFirst(root *Fiber, callback) bool                        │   │  │    │
│  │  │ func MergeLanes(a, b Lane) Lane                                             │   │  │    │
│  │  └─────────────────────────────────────────────────────────────────────┘   │  │    │
│  └───────────────────────────────────────────────────────────────────────────┘  │    │
│                                                                                  │    │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │    │
│  │ 4. Hooks 系统 (hooks.go)                                                         │  │    │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │  │    │
│  │  │ type HookType int (State/Effect/Context/Memo/Ref)                          │   │  │    │
│  │  │ type Hook struct { Type, Value, Deps, Cleanup, Initialized }                    │   │  │    │
│  │  │ type ComponentContext struct {                                                │   │  │    │
│  │  │     ComponentID string                                                         │   │  │    │
│  │  │     Hooks []Hook                                                               │   │  │    │
│  │  │     HookIndex int                                                             │   │  │    │
│  │  │     Validator *HookValidator                                                │   │  │    │
│  │  │ }                                                                            │   │  │    │
│  │  │ type Ref struct { Value interface{} }                                         │   │  │    │
│  │  │ type EffectCallback func() CleanupFunc                                        │   │  │    │
│  │  │                                                                             │   │  │    │
│  │  │ // 上下文管理                                                                 │   │  │    │
│  │  │ func SetCurrentContext(ctx *ComponentContext)                                 │   │  │    │
│  │  │ func GetCurrentContext() *ComponentContext                                  │   │  │    │
│  │  │ func NewComponentContext(name string) *ComponentContext                       │   │  │    │
│  │  └─────────────────────────────────────────────────────────────────────┘   │  │    │
│  └───────────────────────────────────────────────────────────────────────────┘  │    │
│                                                                                  │    │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │    │
│  │ 5. Component 实例系统 (instance.go)                                               │  │    │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │  │    │
│  │  │ type ComponentInstance interface { ... }                                   │   │  │    │
│  │  │ type BaseComponentInstance struct { ... }                                │   │  │    │
│  │  │ func NewBaseComponentInstance(key string, fn ComponentFunc) *BaseComponentInstance  │   │  │    │
│  │  └─────────────────────────────────────────────────────────────────────┘   │  │    │
│  └───────────────────────────────────────────────────────────────────────────┘  │    │
│                                                                                  │
│  ✅ runtime/ui/ 提供的类型供以下包使用:                                        │
│  - ui/ 重导出 (通过类型别名)                                                       │
│  - components/ 实现 (VNode 接口)                                                   │
│  - internal/reconciler/ 使用 (Fiber, Lane, Context, etc.)                             │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 三、组件解耦与自绘架构

### 3.1 组件自绘接口（避免循环引用）

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    组件自绘架构 - runtime/ui/ 作为类型中心                         │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  原则: runtime/ui/ 定义所有类型 → components/ 实现 VNode → reconciler/ 调用              │
│                                                                                  │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │    │
│  │ 1. runtime/ui/ (类型定义层 - 新增核心)                                        │  │    │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │  │    │
│  │  │ // vnode.go - VNode 接口定义                                                │   │  │    │
│  │  │ type VNode interface { ... }                                                 │   │  │    │
│  │  │                                                                             │   │  │    │
│  │  │ // fiber.go - Fiber 相关类型                                                   │   │  │    │
│  │  │ type Lane uint64                                                           │   │  │    │
│  │  │ type Fiber struct { ... }                                                    │   │  │  │    │
│  │  │ type EffectFlag int                                                          │   │  │    │
│  │  │                                                                             │   │  │    │
│  │  │ // hooks.go - Hooks 相关类型                                                   │   │  │    │
│  │  │ type ComponentContext struct { ... }                                         │   │  │    │
│  │  │ type Hook struct { ... }                                                     │   │  │    │
│  │  │ type Ref struct { Value interface{} }                                         │   │  │    │
│  │  │                                                                             │   │  │    │
│  │  │ // instance.go - Component 实例类型                                           │   │  │    │
│  │  │ type ComponentInstance interface { ... }                                    │   │  │    │
│  │  │ type BaseComponentInstance struct { ... }                                 │   │  │    │
│  │  └─────────────────────────────────────────────────────────────────────┘   │  │    │
│  └───────────────────────────────────────────────────────────────────────────┘  │    │
│                              ↑ 实现 VNode 接口                                  │    │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │    │
│  │ 2. components/ (组件实现层)                                                     │  │    │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │  │    │
│  │  │ // components/basic/text.go                                                   │   │  │    │
│  │  │ package basic                                                                  │   │  │    │
│  │  │ import (                                                                       │   │  │    │
│  │  │     "github.com/wwsheng009/mint/runtime/ui"                             │   │  │    │
│  │  │     "github.com/wwsheng009/mint/runtime/paint"                              │   │  │    │
│  │  │ )                                                                               │   │  │    │
│  │  │                                                                               │   │  │    │
│  │  │ // TextVNode 实现 types.VNode 接口                                              │   │  │    │
│  │  │ type TextVNode struct {                                                         │   │  │    │
│  │  │     content string                                                               │   │  │    │
│  │  │     key     string                                                               │   │  │    │
│  │  │     props   types.Props                                                        │   │  │    │
│  │  │     style   runtime.Style                                                      │   │  │    │
│  │  │ }                                                                                │   │  │    │
│  │  │                                                                               │   │  │    │
│  │  │ // 实现 types.VNode 接口方法                                                   │   │  │    │
│  │  │ func (t *TextVNode) Type() types.VNodeType { return types.VNodeText }        │   │  │    │
│  │  │ func (t *TextVNode) Props() types.Props { return t.props }                     │   │  │    │
│  │  │ // ... 其他方法                                                                 │   │  │    │
│  │  │                                                                               │   │  │    │
│  │  │ // 实现 paint.Paintable 接口                                                    │   │  │    │
│  │  │ func (t *TextVNode) Paint(x, y int) []paint.DrawCmd { ... }                 │   │  │    │
│  │  │                                                                               │   │ │    │
│  │  │ // 实现 runtime.Measurable 接口                                              │   │  │  │    │
│  │  │ func (t *TextVNode) Measure(c runtime.BoxConstraints) runtime.Size {  │   │  │    │
│  │  │     // ...                                                                         │   │  │    │
│  │  │ }                                                                                │   │  │    │
│  │  └─────────────────────────────────────────────────────────────────────┘   │  │    │
│  └───────────────────────────────────────────────────────────────────────────┘  │    │
│                              ↑ 通过类型断言调用 Paint/Measure                      │    │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │    │
│  │ 3. internal/reconciler/ (协调器 - 渲染协调)                                       │  │    │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │  │    │
│  │  │ // renderFiberToBuffer 渲染 Fiber 树到 Buffer                                     │   │  │    │
│  │  │ func (r *Reconciler) renderFiberToBuffer(fiber, x, y int, buffer) {   │   │  │    │
│  │  │     switch v := fiber.VNode.(type) {                                          │   │  │    │
│  │  │     case types.VNodeText:  // 通过类型断言获取组件方法                          │   │  │  │    │
│  │  │         if p, ok := v.(paint.Paintable); ok {                                │   │  │    │
│  │  │             cmds := p.Paint(x, y)                                           │   │  │    │
│  │  │             // 执行绘制命令                                                   │   │  │  │    │
│  │  │         }                                                                   │   │  │    │
│  │  │     }                                                                       │   │  │    │
│  │  │ }                                                                             │   │  │    │
│  │  │ }                                                                             │   │  │    │
│  │  └─────────────────────────────────────────────────────────────────────┘   │  │    │
│  └───────────────────────────────────────────────────────────────────────────┘  │    │
│                                                                                  │
│  ✅ 无循环依赖:                                                                  │
│  - components/ 导入 runtime/ui (类型定义)                                      │
│  │             runtime/paint (接口)                                               │
│  - ui/ 导入 runtime/ui 并重导出 (类型别名)                                       │
│  - internal/reconciler/ 导入 runtime/ui 和 runtime/paint                             │
│  - reconciler/ 不导入 ui/，避免循环                                              │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 ui/ 与 components/ 的正确分工

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                      ui/ 和 components/ 的职责划分 (已更新 runtime/ui)                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │    │
│  │ ui/ (API 入口层 - 精简，重导出 runtime/ui)                                      │  │    │
│  │ ┌─────────────────────────────────────────────────────────────────────┐  │  │    │
│  │ │ 只保留:                                                                        │  │  │    │
│  │ │ ├── vnode.go              # VNode = types.VNode (类型别名重导出)           │  │  │    │
│  │ │ ├── element.go            # ElementVNode = types.ElementVNode (重导出)    │  │ │    │
│  │ │ ├── component.go           # ComponentVNode = types.ComponentVNode (重导出) │  │  │    │
│  │ │ ├── fragment.go            # FragmentVNode = types.FragmentVNode (重导出)   │  │  │    │
│  │ │ ├── fiber.go               # Fiber = types.Fiber (重导出)                    │  │  │    │
│  │ │ ├── hooks.go               # UseState, UseEffect (实现，使用 types.Context)  │  │ │    │
│  │ │ ├── instance.go            # ComponentInstance = types.ComponentInstance   │  │  │    │
│  │ │ ├── validator.go           # HookValidator = types.HookValidator         │  │  │    │
│ │ │ ├── layout.go              # HStack, VStack, Box (重导出 types)            │  │  │    │
│  │ │ ├── compat.go              # 兼容性存根 (临时，用于 reconciler)        │  │  │    │
│  │ │ └── shortcuts.go            # 重新导出 components (如需要)             │  │  │    │
│  │ └─────────────────────────────────────────────────────────────────────┘  │  │    │
│  └───────────────────────────────────────────────────────────────────────────┘  │    │
│                              ↓ 实现 types.VNode 接口                             │    │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │    │
│  │ components/ (组件实现层 - 唯一定义)                                             │  │    │
│  │ ┌─────────────────────────────────────────────────────────────────────┐  │  │    │
│  │ │ 组件的完整实现:                                                                  │  │  │    │
│  │ │                                                                               │  │  │    │
│  │ │ components/basic/text.go:                                                       │  │ │    │
│  │ │   type TextVNode struct { content, key, props, style }                             │  │  │    │
│  │ │   func (t *TextVNode) Type() types.VNodeType { return types.VNodeText }       │  │  │    │
│  │ │   func (t *TextVNode) Props() types.Props { return t.props }                      │  │  │    │
│  │ │   // ... 其他 types.VNode 接口方法                                              │  │  │    │
│  │ │   func (t *TextVNode) Paint(x, y int) []paint.DrawCmd { ... }                 │  │  │    │
│  │ │                                                                               │  │  │    │
│  │ │ components/button/button.go:                                                    │  │  │    │
│  │ │   type ButtonVNode struct { label, onClick, ... }                            │  │  │    │
│  │ │   func (b *ButtonVNode) Type() types.VNodeType { ... }                         │  │  │  │    │
│  │ │   func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd { ... }               │  │  │  │    │
│  │ └─────────────────────────────────────────────────────────────────────┘  │  │    │
│  └───────────────────────────────────────────────────────────────────────────┘  │    │
│                                                                                  │
│  ✅ 关键设计:                                                                      │
│  - runtime/ui/ 作为唯一的类型定义来源                                            │
│  - ui/ 通过类型别名重导出，保持向后兼容                                                 │
│  - components/ 实现 types.VNode 接口                                                 │
│  - internal/ 使用 types 类型，不依赖 ui/                                            │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 四、核心模块重构

### 4.1 runtime/ui/ 包结构（已完成）

```
runtime/ui/
├── vnode.go         # VNode 接口, VNodeType, Props, ComponentFunc
├── element.go       # ElementVNode, ElementBuilder
├── component.go     # ComponentVNode, ComponentBuilder
├── fragment.go       # FragmentVNode, Fragment()
├── fiber.go          # Lane, Fiber, EffectFlag, UpdateQueue
├── fiber_util.go     # CreateFiber, CloneFiber, WalkFiber*, MergeLanes
├── hooks.go          # ComponentContext, Hook, Ref, EffectCallback
├── instance.go       # ComponentInstance, BaseComponentInstance
├── validator.go      # HookValidator, HookOrderError
└── layout.go         # LayoutNode, Direction, Align, HStack, VStack
```

### 4.2 类型重导出设计（ui/ → runtime/ui/）

```go
// ui/vnode.go - 通过类型别名重导出 runtime/ui

package ui

import "github.com/wwsheng009/mint/runtime/ui"

// VNode 是虚拟节点接口 - 核心的声明式 UI 系统
type VNode = types.VNode

// VNodeType 代表 VNode 的类型
type VNodeType = types.VNodeType

const (
    VNodeElement   = types.VNodeElement
    VNodeText      = types.VNodeText
    VNodeComponent = types.VNodeComponent
    VNodeFragment  = types.VNodeFragment
)

// Props 表示 VNode 的属性映射
type Props = types.Props

// ComponentFunc 代表返回 VNode 的函数组件
type ComponentFunc = types.ComponentFunc

// ComponentFuncWithProps 代表接受 Props 的组件
type ComponentFuncWithProps = types.ComponentFuncWithProps
```

### 4.3 渲染流程重构

目标流程：
```
VNode Tree (components/ 实现)
    ↓ Measure (runtime.Measurable)
Layout Tree (with x,y,w,h)
    ↓ Paint (paint.Paintable)
Render Tree (DrawCmd[])
    ↓ Rasterize
Buffer
```

---

## 五、实施路线图

### Phase 1: 类型基础迁移 ✅ 已完成

**目标**: 创建 runtime/ui/ 包，迁移所有核心类型

```
✅ runtime/ui/vnode.go          # VNode 接口, VNodeType, Props
✅ runtime/ui/element.go        # ElementVNode, ElementBuilder
✅ runtime/ui/component.go      # ComponentVNode, ComponentBuilder
✅ runtime/ui/fragment.go      # FragmentVNode
✅ runtime/ui/fiber.go         # Lane, Fiber, EffectFlag
✅ runtime/ui/fiber_util.go    # CreateFiber, CloneFiber, 等
✅ runtime/ui/hooks.go         # ComponentContext, Hook, Ref
✅ runtime/ui/instance.go      # ComponentInstance
✅ runtime/ui/validator.go     # HookValidator
✅ runtime/ui/layout.go        # LayoutNode, HStack, VStack

✅ ui/vnode.go                   # 重导出 types.VNode
✅ ui/element.go                  # 重导出 types.ElementVNode
✅ ui/component.go                # 重导出 types.ComponentVNode
✅ ui/fragment.go                  # 重导出 types.FragmentVNode
✅ ui/layout.go                    # 重导出 types.LayoutNode
✅ ui/fiber.go                     # 重导出 types.Fiber
✅ ui/hooks.go                     # 使用 types.ComponentContext
✅ ui/instance.go                  # 重导出 types.ComponentInstance
✅ ui/validator.go                 # 重导出 types.HookValidator

✅ ui/compat.go                    # 兼容性存根（临时）
```

### Phase 2: 组件库迁移 ✅ 已完成

```
✅ components/basic/text.go
✅ components/basic/divider.go
✅ components/layout/
✅ components/form/
✅ components/button/
✅ components/feedback/
✅ components/data/
✅ components/navigation/
✅ components/overlay/
```

### Phase 3: Reconciler 迁移 ✅ 进行中

```
✅ internal/reconciler/vnode_converter.go
```

---

## 六、验证标准

### 6.1 功能验证

| 验证项 | 标准 | 测试方法 |
|--------|------|----------|
| API 兼容性 | 所有公开 API 仍然可用 | 运行现有示例 |
| 类型独立 | runtime/ui/ 可独立导入 | `import "github.com/wwsheng009/mint/runtime/ui"` |
| 渲染正确 | 与当前实现输出一致 | 视觉对比测试 |
| 性能不退化 | 关键操作性能保持 | 基准测试 |

### 6.2 架构验证

| 验证项 | 标准 |
|--------|------|
| 分层清晰 | runtime/ui/ → ui/ → components/ → internal/ |
| 职责单一 | runtime/ui/ 只负责类型定义 |
| 依赖单向 | internal/ 不依赖 ui/，只依赖 runtime/ui/ |
| 接口稳定 | runtime/ui/ 定义的接口稳定 |

### 6.3 代码质量

```bash
# 验证命令
go build ./ui/...                # ui/ 编译通过
go build ./internal/...           # internal/ 编译通过
go build ./runtime/ui/...       # runtime/ui/ 编译通过
go build ./components/...         # components/ 编译通过
```

---

**文档版本**: v1.2 (runtime/ui 基础包)
**最后更新**: 2026-02-02
**更新内容**:
- ✅ 新增 runtime/ui/ 作为类型基础包
- ✅ 更新架构分层图，展示 runtime/ui/ 的中心地位
- ✅ 明确类型重导出设计（ui/ 通过类型别名重导出）
- ✅ 更新依赖关系图，消除循环依赖
