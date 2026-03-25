# Mint UI 组件迁移指南

本文档指导将 `components/` 目录下的组件迁移到 Fiber-first 架构的 `ui/components/` 目录。

## 目录

- [概述](#概述)
- [架构变化](#架构变化)
- [目录结构变化](#目录结构变化)
- [迁移步骤](#迁移步骤)
- [代码示例](#代码示例)
- [布局系统新特性](#布局系统新特性)
- [示例程序要求](#示例程序要求)
- [验证清单](#验证清单)
- [迁移进度追踪](#迁移进度追踪)

---

## 概述

### 迁移目标

将旧架构组件迁移到 Fiber-first 架构：
- **描述与状态分离**：VNode 只负责描述，Instance 负责状态
- **事件处理解耦**：使用 Intent 替代闭包
- **行为组合**：使用 Behavior 模式复用通用行为
- **两阶段布局**：Measure（测量）→ SetBounds（布局）→ Paint（绘制）

### 已迁移组件参考

以下目录可作为迁移参考：
- ✅ `ui/components/button/` - 标准交互组件
- ✅ `ui/components/input/` - 表单输入组件
- ✅ `ui/components/form/` - FormItem、字段绑定、验证联动
- ✅ `ui/components/grid/` - 布局/测量/调试支持
- ✅ `ui/components/wrap/` - 布局容器
- ✅ `ui/components/absolute/` - 绝对定位容器
- ✅ `ui/components/panel/` - 容器组件
- ✅ `ui/components/select/` - 复合选择器 + overlay
- ✅ `ui/components/table/` - 数据组件
- ✅ `ui/components/treeview/` - 复杂树形数据组件
- ✅ `ui/components/drawer/` - 覆盖层组件
- ✅ `ui/components/statusbar/` - 组合式 builder 特例

### ✅ 当前状态

`ui/components/` 已从“迁移中”进入“主体组件面基本补齐、以文档和增强收口为主”的阶段。

- 排除 `docs/` 与 `internal/` 后，当前共有 58 个目录
- 其中 54 个严格遵循 `builder.go + vnode.go + instance.go` 规范
- `toast`、`statusbar`、`control`、`validation` 是刻意保留的特例/支撑模块
- 历史上的 `stack` / `border` 现在主要表现为运行时布局/边框能力与示例，不再对应 `ui/components/stack/`、`ui/components/border/` 目录
- 完整组件现状与 backlog 以 [ROADMAP.md](./ROADMAP.md) 和 [OPTIMIZATION_BACKLOG.md](./OPTIMIZATION_BACKLOG.md) 为准

---

## 架构变化

### 旧架构

```
components/<category>/<component>.go
├── VNode struct
│   ├── state (状态直接在 VNode 中)
│   ├── callbacks (闭包回调)
│   ├── Paint() (rendering logic)
│   ├── Measure()
│   └── HandleAction()
└── Builder pattern
```

**问题**：
- VNode 混合了状态、逻辑和渲染
- 闭包导致序列化和对比困难
- 难以测试和行为复用

### 新架构 (Fiber-first)

```
ui/components/<component>/
├── vnode.go      - 纯描述性 VNode
│   ├── props (只读属性)
│   ├── CreateInstance()
│   └── 无状态、无闭包、无 Paint
│
├── builder.go    - Fluent API
│   └── 链式构建 VNode
│
├── instance.go   - 运行时实例
│   ├── state (内部状态)
│   ├── behaviors (行为组合)
│   ├── Measure()
│   ├── Paint()
│   └── HandleAction()
│
└── tracing.go (可选) - Fiber 调试支持
```

**优势**：
- 描述与实现分离，易于测试
- Intent 替代闭包，支持序列化
- Behavior 复用通用行为模式

---

## 目录结构变化

### 旧结构

```
components/
├── basic/
│   ├── divider.go
│   ├── text.go
│   └── doc.go
├── button/
│   ├── button.go
│   ├── button_instance.go
│   ├── button_action_test.go
│   └── ...
├── form/
│   ├── input.go
│   ├── checkbox.go
│   ├── select.go
│   └── textarea.go
├── layout/
│   ├── stack.go
│   ├── absolute.go
│   ├── grid.go
│   └── ...
└── ...
```

### 新结构

```
ui/components/
├── button/
│   ├── vnode.go       # VNode 描述
│   ├── builder.go     # Fluent API
│   ├── instance.go    # 运行时实例
│   ├── tracing.go     # 调试钩子
│   └── button_test.go # 测试
├── wrap/
│   ├── vnode.go
│   ├── builder.go
│   ├── instance.go
│   ├── README.md
│   └── wrap_test.go
└── ...
```

### 映射关系

| 旧路径 | 新路径 | 组件类型 |
|--------|--------|----------|
| `components/button/` | `ui/components/button/` | 交互组件 |
| `components/basic/text.go` | `ui/components/text/` | 基础组件 |
| `components/basic/divider.go` | `ui/components/divider/` | 基础组件 |
| `components/form/input.go` | `ui/components/input/` | 表单组件 |
| `components/form/checkbox.go` | `ui/components/checkbox/` | 表单组件 |
| `components/form/select.go` | `ui/components/select/` | 表单组件 |
| `components/form/textarea.go` | `ui/components/textarea/` | 表单组件 |
| `components/layout/stack.go` | `runtime/ui` 的 `VStackBuilder` / `HStackBuilder` | 布局原语 |
| `components/layout/grid.go` | `ui/components/grid/` | 布局组件 |
| `components/layout/absolute.go` | `ui/components/absolute/` | 布局组件 |
| `components/layout/wrap.go` | `ui/components/wrap/` | 布局组件 |
| `components/layout/scroll_view.go` | `ui/components/scrollview/` | 布局组件 |
| `components/container/panel.go` | `ui/components/panel/` | 容器组件 |
| `components/feedback/progress.go` | `ui/components/progress/` | 反馈组件 |
| `components/overlay/modal.go` | `ui/components/modal/` | 覆盖层组件 |
| `components/navigation/tabs.go` | `ui/components/tabs/` | 导航组件 |
| `components/data/table.go` | `ui/components/table/` | 数据组件 |
| `components/data/virtuallist.go` | `ui/components/virtuallist/` | 数据组件 |
| `display/treeview.go` | `ui/components/treeview/` | 展示组件（简化版）|
| `components/data/list.go` | `ui/components/list/` | 数据组件 |

---

## 迁移步骤

### 第一步：创建目录和基础文件

```bash
# 创建新组件目录
mkdir -p ui/components/<component>

# 创建基础文件
touch ui/components/<component>/vnode.go
touch ui/components/<component>/builder.go
touch ui/components/<component>/instance.go
touch ui/components/<component>/<component>_test.go
```

### 第二步：编写 VNode (vnode.go)

**原则**：
- VNode 只包含声明性属性
- 无状态、无闭包、无 Paint 逻辑
- 实现 `rtui.InstanceFactory` 接口

**代码模板**：

```go
package <component>

import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/style"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types - 使用 rtui 标准类型
// =============================================================================

// Variant 组件变体类型
type Variant int

const (
    VariantDefault Variant = iota
    VariantPrimary
    // ...
)

// =============================================================================
// VNode - 纯描述性 VNode
// =============================================================================

// VNode 是组件的纯描述性结构
// 只包含声明性信息 - 无状态、无闭包、无 Paint 逻辑
type VNode struct {
    *rtui.ElementVNode

    // === 识别 ===
    key string

    // === 视觉属性 ===
    variant  Variant
    style    style.Style

    // === 状态属性（声明性）===
    disabled bool

    // === Intent 属性（无闭包！）===
    pressIntent intent.Intent
}

// 确保实现所需接口
var (
    _ rtui.VNode           = (*VNode)(nil)
    _ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New 创建新的 VNode
func New(/* 参数列表 */) *VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("<component>"),
        // 初始化默认值
    }
}

// =============================================================================
// rtui.VNode 接口实现
// =============================================================================

func (v *VNode) Key() string           { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode { v.key = key; return v }

func (v *VNode) Tag() string           { return "<component>" }

func (v *VNode) Style() style.Style    { return v.style }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.style = s; return v }

func (v *VNode) Children() []rtui.VNode { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer   { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
    return rtui.Props{
        "key":         v.key,
        "variant":     v.variant,
        "disabled":    v.disabled,
        "pressIntent": v.pressIntent,
    }
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
    if v, ok := p["key"].(string); ok {
        v.key = v
    }
    if v, ok := p["variant"].(Variant); ok {
        v.variant = v
    }
    if v, ok := p["disabled"].(bool); ok {
        v.disabled = v
    }
    if v, ok := p["pressIntent"].(intent.Intent); ok {
        v.pressIntent = v
    }
    return v
}

// =============================================================================
// InstanceFactory 实现
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
    return NewInstance(rtui.Props{
        "key":         v.key,
        "variant":     v.variant,
        "disabled":    v.disabled,
        "pressIntent": v.pressIntent,
    })
}

// =============================================================================
// Setter 方法
// =============================================================================

func (v *VNode) SetVariant(vt Variant) *VNode { v.variant = vt; return v }
func (v *VNode) SetDisabled(d bool) *VNode { v.disabled = d; return v }
func (v *VNode) SetIntent(i intent.Intent) *VNode { v.pressIntent = i; return v }
// ...
```

### 第三步：编写 Instance (instance.go)

**原则**：
- Instance 承载运行时状态
- 实现 ComponentInstance, PaintableInstance, FocusableInstance
- 使用 Behavior 组合通用行为
- 实现 Measure(), Paint()

**代码模板**：

```go
package <component>

import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/ui/components/control"
)

// =============================================================================
// Instance - 运行时实体
// =============================================================================

// Instance 是组件的运行时实体
// 在渲染之间持久存在并持有所有状态
type Instance struct {
    // === 识别 ===
    key string

    // === Props（来自 VNode，每次渲染可能变化）===
    variant     Variant
    componentStyle style.Style
    pressIntent intent.Intent

    // === 运行时状态（由 Instance 管理）===
    state   control.InteractionState
    bounds  [4]int // x, y, w, h
    dirty   bool

    // === Intent Emitter ===
    intentEmitter func(intent.Intent)

    // === Behaviors ===
    behaviors *control.BehaviorList
}

// 确保实现所需接口
var (
    _ rtui.ComponentInstance     = (*Instance)(nil)
    _ rtui.PaintableInstance     = (*Instance)(nil)
    _ rtui.FocusableInstance     = (*Instance)(nil)
    _ rtui.ActionHandlerInstance = (*Instance)(nil)
    _ control.Instance           = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance 从 props 创建新的 Instance
func NewInstance(props rtui.Props) *Instance {
    inst := &Instance{
        key:          getStringProp(props, "key", ""),
        variant:      getVariantProp(props, VariantDefault),
        componentStyle: getStyleProp(props),
        pressIntent:  getIntentProp(props),
        dirty:        true,
    }

    // 初始化状态
    inst.state = control.InteractionState{
        Disabled: getBoolProp(props, "disabled", false),
    }

    // 初始化行为
    inst.initBehaviors()

    return inst
}

// initBehaviors 初始化行为组合
func (inst *Instance) initBehaviors() {
    pressable := control.NewPressableBehavior(inst.pressIntent)
    inst.behaviors = control.NewBehaviorList(
        &control.FocusableBehavior{},
        pressable,
        &control.HoverableBehavior{},
        &control.DisableableBehavior{},
    )
}

// =============================================================================
// ComponentInstance 接口
// =============================================================================

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }

func (inst *Instance) Init(props rtui.Props) {
    inst.SetProps(props)
}

func (inst *Instance) Destroy() {
    inst.behaviors.OnUnmount(inst)
}

func (inst *Instance) OnMount() {
    inst.behaviors.OnMount(inst)
}

func (inst *Instance) OnUnmount() {
    inst.behaviors.OnUnmount(inst)
}

func (inst *Instance) SetProps(props rtui.Props) bool {
    oldVariant := inst.variant
    oldDisabled := inst.state.Disabled
    oldIntent := inst.pressIntent

    inst.variant = getVariantProp(props, inst.variant)
    inst.componentStyle = getStyleProp(props)
    inst.pressIntent = getIntentProp(props)

    newDisabled := getBoolProp(props, "disabled", inst.state.Disabled)
    if newDisabled != inst.state.Disabled {
        inst.state.Disabled = newDisabled
    }

    // 更新 pressable behavior intent
    if inst.pressIntent != oldIntent {
        if pressable := inst.behaviors.Get("Pressable"); pressable != nil {
            if p, ok := pressable.(*control.PressableBehavior); ok {
                p.SetIntent(inst.pressIntent)
            }
        }
    }

    changed := oldVariant != inst.variant ||
        oldDisabled != inst.state.Disabled ||
        oldIntent != inst.pressIntent

    if changed {
        inst.dirty = true
    }
    return changed
}

func (inst *Instance) GetProps() rtui.Props {
    return rtui.Props{
        "key":      inst.key,
        "variant":  inst.variant,
        "disabled": inst.state.Disabled,
    }
}

func (inst *Instance) MarkDirty()         { inst.dirty = true }
func (inst *Instance) IsDirty() bool      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

// =============================================================================
// PaintableInstance 接口
// =============================================================================

// Paint 实现绘制逻辑
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
    // 构建绘制命令
    // 示例：
    // style := inst.resolveStyle()
    // return []paint.DrawCmd{{
    //     X:     x,
    //     Y:     y,
    //     Text:  inst.buildText(),
    //     Style: style,
    // }}
    return nil
}

// resolveStyle 解析当前状态对应的样式
func (inst *Instance) resolveStyle() style.Style {
    s := inst.componentStyle

    // 应用状态样式
    if inst.state.Disabled {
        // ...
    } else if inst.state.Focused {
        // ...
    } else if inst.state.Hovered {
        // ...
    }

    return s
}

// =============================================================================
// FocusableInstance 接口
// =============================================================================

func (inst *Instance) SetFocus(focused bool) {
    if inst.state.Focused != focused {
        oldState := inst.state
        inst.state.Focused = focused
        inst.dirty = true
        inst.behaviors.OnStateChange(inst, oldState, inst.state)
    }
}

func (inst *Instance) HasFocus() bool { return inst.state.Focused }
func (inst *Instance) IsDisabled() bool { return inst.state.Disabled }

// =============================================================================
// ActionHandlerInstance 接口
// =============================================================================

func (inst *Instance) CanHandleAction(actionType string) bool {
    if inst.state.Disabled {
        return false
    }
    return inst.behaviors.OnAction(inst, actionType, nil)
}

func (inst *Instance) HandleAction(actionType string, payload interface{}) bool {
    if inst.state.Disabled {
        return false
    }
    return inst.behaviors.OnAction(inst, actionType, payload)
}

// =============================================================================
// control.Instance 接口（用于 Behaviors）
// =============================================================================

func (inst *Instance) GetState() *control.InteractionState { return &inst.state }
func (inst *Instance) SetState(state control.InteractionState) {
    oldState := inst.state
    inst.state = state
    inst.behaviors.OnStateChange(inst, oldState, inst.state)
}

func (inst *Instance) EmitIntent(i intent.Intent) {
    if inst.intentEmitter != nil {
        inst.intentEmitter(i)
    }
}

func (inst *Instance) GetBounds() (x, y, w, h int) {
    return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *Instance) SetBounds(x, y, w, h int) {
    inst.bounds = [4]int{x, y, w, h}
}

func (inst *Instance) GetStyle() style.Style {
    return inst.componentStyle
}

func (inst *Instance) SetStyle(s style.Style) {
    inst.componentStyle = s
}

func (inst *Instance) GetProp(key string) (interface{}, bool) {
    // ...
}

func (inst *Instance) SetProp(key string, value interface{}) {
    // ...
}

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
    inst.intentEmitter = fn
}

func (inst *Instance) ClearDirty() { inst.dirty = false }

// =============================================================================
// Measurable 接口（两阶段布局）
// =============================================================================

// Measure 实现布局测量，两阶段布局 Phase 1
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // 计算组件的自然尺寸
    // ...
    return layout.Size{Width: 0, Height: 0}
}

// =============================================================================
// Prop 提取辅助函数
// =============================================================================

func getStringProp(props rtui.Props, key, def string) string {
    if v, ok := props[key]; ok {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
    // ...
}

func getVariantProp(props rtui.Props, def Variant) Variant {
    // ...
}

func getStyleProp(props rtui.Props) style.Style {
    // ...
}

func getIntentProp(props rtui.Props) intent.Intent {
    // ...
}
```

### 第四步：编写 Builder (builder.go)

**原则**：
- 提供链式调用 API
- 方法返回 *Builder 以支持链式调用
- 最后调用 Build() 返回 VNode

**代码模板**：

```go
package <component>

import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/style"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder 提供用于创建 VNode 的 Fluent API
type Builder struct {
    node *VNode
}

// NewBuilder 创建新的 Builder
func NewBuilder(/* 初始参数 */) *Builder {
    return &Builder{
        node: New(/* 参数 */),
    }
}

// Key 设置 diffing key
func (b *Builder) Key(key string) *Builder {
    b.node.SetKey(key)
    return b
}

// Variant 设置组件变体
func (b *Builder) Variant(v Variant) *Builder {
    b.node.SetVariant(v)
    return b
}

// Primary 设置为主样式（快捷方法）
func (b *Builder) Primary() *Builder {
    return b.Variant(VariantPrimary)
}

// Disabled 设置禁用状态
func (b *Builder) Disabled(disabled bool) *Builder {
    b.node.SetDisabled(disabled)
    return b
}

// OnPress 设置按下时发的 Intent
// 替代旧的 OnClick(func()) 闭包模式
func (b *Builder) OnPress(pressIntent intent.Intent) *Builder {
    b.node.SetIntent(pressIntent)
    return b
}

// Style 设置视觉样式
func (b *Builder) Style(s style.Style) *Builder {
    b.node.SetStyleProps(s)
    return b
}

// FgColor 设置前景色
func (b *Builder) FgColor(c interface{}) *Builder {
    s := b.node.Style()
    switch v := c.(type) {
    case string:
        s.FG = style.Color(v)
    case style.Color:
        s.FG = v
    }
    b.node.SetStyleProps(s)
    return b
}

// BgColor 设置背景色
func (b *Builder) BgColor(c interface{}) *Builder {
    s := b.node.Style()
    switch v := c.(type) {
    case string:
        s.BG = style.Color(v)
    case style.Color:
        s.BG = v
    }
    b.node.SetStyleProps(s)
    return b
}

// Build 返回 VNode
func (b *Builder) Build() rtui.VNode {
    return b.node
}

// BuildInstance 直接创建 Instance
func (b *Builder) BuildInstance() *Instance {
    return NewInstance(b.node.Props())
}

// =============================================================================
// 便捷函数
// =============================================================================

// 快捷创建函数示例
func <Component>(/* 参数 */) rtui.VNode {
    return NewBuilder(/* 参数 */).Build()
}
```

### 第五步：编写测试

**参考**：`ui/components/button/button_test.go`, `ui/components/text/text_test.go`

**测试要点**：
- VNode 接口实现测试
- Instance Measure/Paint 测试
- Builder 链式调用测试
- Props 更新测试

---

## 代码示例

### 迁移前：旧架构 Button

```go
// components/button/button.go

type ButtonVNode struct {
    *ui.ElementVNode
    label    string
    variant  ButtonVariant
    disabled bool
    onClick  func()  // ← 闭包
    // ...
}

func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd {
    // 渲染逻辑在 VNode 中
}

func (b *ButtonVNode) OnClick(fn func()) *ButtonVNode {
    b.onClick = fn  // ← 使用闭包
    return b
}
```

### 迁移后：Fiber-first Button

```go
// ui/components/button/vnode.go

type VNode struct {
    *rtui.ElementVNode
    label       string
    variant     Variant
    pressIntent intent.Intent  // ← Intent 替代闭包
    disabled    bool
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
    return NewInstance(rtui.Props{
        "label":       v.label,
        "pressIntent": v.pressIntent,
        // ...
    })
}

// ui/components/button/instance.go

type Instance struct {
    pressIntent intent.Intent
    state       control.InteractionState
    behaviors   *control.BehaviorList
    // ...
}

func (inst *Instance) HandleAction(actionType string, payload interface{}) bool {
    return inst.behaviors.OnAction(inst, actionType, payload)
}
```

---

## 布局系统新特性

### 概述

`docs/layout/plan/` 目录中实施了一系列布局系统优化，新迁移的组件应当参考并应用这些新特性。

### Phase 1: 约束传播优化（短期）

#### 1.1 约束传递规则

**优先级原则**：显式维度 > 父约束 > 自动测量

```go
// 组件应优先使用显式设置的维度（如 width > 0）
func (inst *Instance) computeChildConstraints(constraints layout.Constraints) layout.Constraints {
    cc := constraints

    // 规则: 显式维度优先
    if inst.width > 0 {
        cc.MinWidth = inst.width
        cc.MaxWidth = inst.width
    }

    if inst.height > 0 {
        cc.MinHeight = inst.height
        cc.MaxHeight = inst.height
    }

    // 规则: 边框内边距
    cc.MinWidth = max(0, cc.MinWidth - inst.borderPadding)
    cc.MaxWidth = max(0, cc.MaxWidth - inst.borderPadding)
    cc.MinHeight = max(0, cc.MinHeight - inst.borderPadding)
    cc.MaxHeight = max(0, cc.MaxHeight - inst.borderPadding)

    return cc
}
```

#### 1.2 约束追踪工具

新组件应支持约束追踪，用于调试：

```go
import "github.com/wwsheng009/mint/ui/layout/constraints"

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    childConstraints := inst.computeChildConstraints(constraints)

    // 追踪约束传递
    if constraints.IsEnabled() {
        constraints.Trace(
            inst.Tag(), inst.child.Tag(),
            inst.GetPath(),
            constraints, childConstraints,
            layout.Size{},
            "Applied child constraints",
        )
    }

    childSize := inst.child.Measure(childConstraints)

    // 更新追踪结果
    if constraints.IsEnabled() {
        entries := constraints.GetEntries()
        if len(entries) > 0 {
            entries[len(entries)-1].Dimension = inst.computeOuterSize(childSize)
        }
    }

    return inst.computeOuterSize(childSize)
}
```

### Phase 2: API 改进（中期）

#### 2.1 明确的维度语义

对于容器类组件（如 Panel, Border），应提供明确的内外部维度 API：

```go
// 外部维度 API（包含边框/内边距）
func (v *VNode) SetOuterWidth(w int) *VNode    // 或 SetWidth(w)
func (v *VNode) SetOuterHeight(h int) *VNode  // 或 SetHeight(h)

// 内部维度 API（内容区域，不含边框/内边距）
func (v *VNode) SetInnerWidth(w int) *VNode
func (v *VNode) SetInnerHeight(h int) *VNode

// 内容维度别名（更直观）
func (v *VNode) SetContentWidth(w int) *VNode  // = SetInnerWidth
func (v *VNode) SetContentSize(lineCount int) *VNode  // = SetInnerHeight

// 获取维度信息
func (v *VNode) GetOuterDimensions() (width, height int)
func (v *VNode) GetInnerDimensions() (width, height int)
func (v *VNode) GetPadding() (width, height int)
```

#### 2.2 Builder API 增强

```go
// Builder 应提供便捷方法
func (b *Builder) WithFixed(w, h int) *Builder       // 固定外部尺寸
func (b *Builder) WithFixedInner(w, h int) *Builder  // 固定内部尺寸
func (b *Builder) WithContentWidth(w int) *Builder   // 固定内容宽度
func (b *Builder) WithAutoHeight() *Builder          // 自动高度

// 便捷的内容设置
func (b *Builder) WithWrappedText(text string, width int) *Builder
func (b *Builder) WithTextContent(text string) *Builder
```

### Phase 3: 布局引擎和可视化工具（长期）

#### 3.1 维度计算规则

**边框内边距**：
- 单线/双线/圆角边框都占用 1 个字符单元格
- `GetBorderWidth(style)` 返回 0 或 1
- 总 padding = `2 * GetBorderWidth(style)`

**维度转换**：
```
外部维度 = 内部维度 + 边框 padding
内部维度 = 外部维度 - 边框 padding
```

#### 3.2 Text.Wrap 双重约束处理

```go
// Measure 阶段：根据 MaxWidth 计算行数
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    if inst.wrap {
        maxWidth := constraints.MaxWidth
        lines := wordWrap(inst.content, maxWidth)
        return Size{Width: maxWidth, Height: len(lines)}
    }
}

// Paint 阶段：同时遵守 MaxWidth 和 maxHeight
func (inst *Instance) Paint(ctx PaintContext, buf *Buffer) {
    if inst.wrap {
        lines := wordWrap(inst.content, ctx.Bounds[2]) // 使用 MaxWidth
        for i, line := range lines {
            if ctx.Y+i >= ctx.Bounds[3] {  // 检查 maxHeight
                break  // 超出部分裁剪
            }
            render(line, ctx.X, ctx.Y+i)
        }
    }
}
```

### 参考文档

- `docs/layout/plan/01-analysis.md` - 布局系统详细分析
- `docs/layout/plan/02-optimization.md` - 优化方案（Phase 1-3）
- `docs/layout/plan/03-testing.md` - 测试方案
- `docs/layout/plan/04-debug-tools.md` - 调试工具

---

## 示例程序要求

### 概述

每个新迁移的组件都需要在 `examples/fiber_firsts/<component>_demo/` 目录下提供示例程序，展示组件的使用方法和特性。

### 示例程序结构

```
examples/fiber_firsts/<component>_demo/
├── main.go              # 示例主程序
└── (可选) README.md      # 组件特性说明
```

### 示例程序模板

```go
// Fiber-first <Component> Component Demo
// Demonstrates the new <Component> component following the Fiber-first architecture
package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/framework/component"
    "github.com/wwsheng009/mint/framework/theme"
    "github.com/wwsheng009/mint/internal/render"
    "github.com/wwsheng009/mint/runtime/paint"
    rtui "github.com/wwsheng009/mint/runtime/ui"

    // 导入新组件
    "<component> "github.com/wwsheng009/mint/ui/components/<component>"
    newtext "github.com/wwsheng009/mint/ui/components/text"
    // 导入其他依赖组件
)

// DemoApp renders <Component> features using the Fiber-first components
func DemoApp() rtui.VNode {
    return newtext.New("Demo content")
}

func main() {
    os.Setenv("MINT_USE_FIBER", "true")
    os.Setenv("MINT_FIBER_FIRST", "true")
    os.Setenv("MINT_DEBUG_TEST", "true")

    fmt.Println("╔════════════════════════════════════════════════════════════╗")
    fmt.Println("║   Fiber-First <Component> Rendering Demo                   ║")
    fmt.Println("╚════════════════════════════════════════════════════════════╝")

    // Create framework app (required for Fiber reconciler)
    fwApp := framework.NewApp()

    // Create DeclarativeNode WITH Fiber reconciler
    node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)

    // Enable Fiber-first mode
    node.SetRenderMode(render.RenderModeFiberFirst)

    fmt.Printf("\nConfiguration:\n")
    fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())
    fmt.Printf("  Fiber-First Enabled: %v\n", node.IsFiberFirstEnabled())

    // Create buffer
    buf := paint.NewBuffer(60, 40)

    // Create paint context
    ctx := component.PaintContext{
        Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 40},
        AvailableWidth:  60,
        AvailableHeight: 40,
    }

    fmt.Printf("\n%s\n", strings.Repeat("=", 60))
    fmt.Println("Rendering <Component> with Fiber-first pipeline...")
    fmt.Printf("%s\n\n", strings.Repeat("=", 60))

    // Render
    node.Paint(ctx, buf)

    // Output result
    printBuffer(buf, 60, 40)

    fmt.Println("\n" + strings.Repeat("=", 60))
    fmt.Println("<Component> Features:")
    fmt.Println(strings.Repeat("=", 60))
    fmt.Println("  - Feature 1")
    fmt.Println("  - Feature 2")
    fmt.Println(strings.Repeat("=", 60))
}

func printBuffer(buf *paint.Buffer, width, height int) {
    fmt.Printf("┌%s┐\n", strings.Repeat("─", width))
    for y := 0; y < height; y++ {
        var line strings.Builder
        for x := 0; x < width; x++ {
            cell := buf.GetContent(x, y)
            if cell.Cluster != "" {
                line.WriteString(cell.Cluster)
            } else {
                line.WriteString(" ")
            }
        }
        trimmed := strings.TrimRight(line.String(), " ")
        if trimmed != "" {
            fmt.Printf("|%-*s|\n", width, trimmed)
        }
    }
    fmt.Printf("└%s┘\n", strings.Repeat("─", width))
}
```

### 示例内容要求

参考 `examples/fiber_firsts/stack_demo/main.go` 的结构，示例程序应包含：

1. **基础用法**：展示组件的基本功能
2. **不同变体**：展示组件的不同变体/样式
3. **组合使用**：展示组件与其他组件的组合
4. **边界情况**：展示组件在边界情况下的行为
5. **布局特性**：如适用，展示组件在不同布局中的表现

### 参考示例

- `examples/fiber_firsts/stack_demo/` - Stack 组件完整示例
- `examples/fiber_firsts/panel_demo/` - Panel 组件示例
- `examples/fiber_firsts/button_demo/` - Button 组件示例

---

## 验证清单

迁移完成后，验证以下内容：

### Vnode 验证
- [ ] VNode 只包含声明性属性
- [ ] 无状态字段（如 hasFocus, isHovered）
- [ ] 无闭包字段
- [ ] 无 Paint() 方法
- [ ] 实现 rtui.InstanceFactory
- [ ] Props() / SetProps() 正确实现

### Instance 验证
- [ ] 实现 ComponentInstance
- [ ] 实现 PaintableInstance（如需自绘）
- [ ] 实现 FocusableInstance（如有焦点）
- [ ] 实现 ActionHandlerInstance
- [ ] Measure() 使用 layout.Constraints
- [ ] 使用 Behavior 组合
- [ ] SetProps() 正确更新状态

### Builder 验证
- [ ] 所有 setter 返回 *Builder
- [ ] Build() 返回 rtui.VNode
- [ ] BuildInstance() 返回 *Instance
- [ ] 便捷函数可用

### 测试验证
- [ ] 单元测试覆盖 VNode
- [ ] 单元测试覆盖 Instance
- [ ] Builder 测试通过
- [ ] 所有现有测试通过
- [ ] 约束传播测试通过（如果是布局组件）
- [ ] 边界情况测试通过

### 文档验证
- [ ] 导入路径已更新到 rtui
- [ ] intent 包已引入
- [ ] 旧代码引用已更新

### 布局特性验证（如适用）
- [ ] 约束传递规则正确（显式维度 > 父约束 > 自动测量）
- [ ] 支持约束追踪（可以记录约束传递）
- [ ] 外部/内部维度语义清晰
- [ ] 边框内边距正确计算
- [ ] Text.Wrap 双重约束正确处理

### 示例程序验证
- [ ] 示例程序位于 `examples/fiber_firsts/<component>_demo/`
- [ ] 示例程序展示组件基本功能
- [ ] 示例程序展示不同变体
- [ ] 示例程序展示组合使用
- [ ] 示例程序可以独立运行
- [ ] 示例程序有清晰的说明和特性总结

---

## 迁移进度追踪

### 已迁移

| 组件 | 旧路径 | 新路径 | 示例程序 |
|------|--------|--------|----------|
| Button | `components/button/` | `ui/components/button/` | ✅ README + 单测 |
| Stack | `components/layout/stack.go` | `runtime/ui` 的 `VStackBuilder` / `HStackBuilder` | ✅ stack_demo |
| Text | `components/basic/text.go` | `ui/components/text/` | ✅ text_demo |
| Divider | `components/basic/divider.go` | `ui/components/divider/` | ✅ divider_demo |
| Input | `components/form/input.go` | `ui/components/input/` | ✅ input_demo |
| Checkbox | `components/form/checkbox.go` | `ui/components/checkbox/` | ✅ checkbox_demo |
| Panel | `components/container/panel.go` | `ui/components/panel/` | ✅ panel_demo |
| Grid | `components/layout/grid.go` | `ui/components/grid/` | ✅ grid_demo |
| ScrollView | `components/layout/scroll_view.go` | `ui/components/scrollview/` | ✅ scrollview_demo |
| Select | `components/form/select.go` | `ui/components/select/` | ✅ select_demo |
| Textarea | `components/form/textarea.go` | `ui/components/textarea/` | ✅ textarea_demo |
| Wrap | `components/layout/wrap.go` | `ui/components/wrap/` | ✅ wrap_demo |
| Absolute | `components/layout/absolute.go` | `ui/components/absolute/` | ✅ absolute_demo |
| Border | - | `runtime/ui` 容器边框能力 | ✅ border_demo |
| Progress | `components/feedback/progress.go` | `ui/components/progress/` | ✅ progress_demo |
| Modal | `components/overlay/modal.go` | `ui/components/modal/` | ✅ modal_demo |
| Tabs | `components/navigation/tabs.go` | `ui/components/tabs/` | ✅ tabs_demo |
| Table | `components/data/table.go` | `ui/components/table/` | ✅ table_demo |
| VirtualList | `components/data/virtuallist.go` | `ui/components/virtuallist/` | ✅ virtuallist_demo |
| TreeView | `display/treeview.go` | `ui/components/treeview/` | ✅ treeview_demo |
| List | `components/data/list.go` | `ui/components/list/` | ✅ list_demo |

### 历史会话记录

以下内容保留为早期迁移会话的归档记录，反映当时一次集中迁移的产出，不代表当前目录的完整清单。

#### 本次迁移会话完成的组件：
1. **Modal** - 模态对话框 (443行旧代码)
2. **Tabs** - 标签页导航 (1097行旧代码)
3. **Table** - 数据表格 (319行旧代码)
4. **VirtualList** - 虚拟列表 (660行旧代码，42个测试)
5. **TreeView** - 树形视图 (1512行旧代码，简化版，30个测试)
6. **List** - 列表组件 (1095行旧代码)

总计迁移：6个组件，约5,126行代码全部迁移到Fiber-first架构

### 迁移注意点

#### 1. 复杂交互组件 (Tabs, TreeView)
- 需要特别处理状态管理
- 考虑拆分多个 Behavior
- 可能需要自定义 State 结构

#### 2. 覆盖层组件 (Modal, Tooltip)
- 需要处理层级
- 可能与 Fiber 的 Layer 系统集成
- 需要管理焦点和事件路由

#### 3. 数据组件 (List, Table)
- 需要虚拟滚动支持
- 考虑与 ScrollView 集成
- 注意性能优化

---

## 参考文档

### 相关架构文档
- `fiber_paint.md` - Fiber-first 架构详解
- `fiber_action.md` - Action/Intent 框架
- `ui/components/control/types.go` - Behavior 实现

### 示例组件
- `ui/components/button/` - 完整的交互组件参考
- `ui/components/grid/` - 布局与测量参考
- `ui/components/wrap/` - 布局容器参考
- `ui/components/text/` - 简单组件参考

### 已迁移组件对比
对比以下组件了解迁移差异：
- `components/button/button.go` vs `ui/components/button/vnode.go`
- `components/layout/stack.go` vs `runtime/ui/layout.go`

---

## 迁移支持

### 常见问题

**Q: 如何处理复杂的组件状态？**
A: 创建自定义 State 结构，在 Instance 中管理，通过 Behavior 暴露行为。

**Q: 如何迁移使用闭包的回调？**
A: 创建对应的 Intent 类型，使用 `intent.NewAction()` 或自定义 Intent。

**Q: 组件有多个行为如何处理？**
A: 使用 `control.NewBehaviorList()` 组合多个 Behavior。

**Q: 布局组件如何迁移？**
A: 优先参考 `runtime/ui/layout.go` 中的 `VStackBuilder` / `HStackBuilder`，以及 `ui/components/grid/`、`ui/components/wrap/` 的 `Measure()` 实现。

---

## 总结

迁移到 Fiber-first 架构的关键改进：

1. **关注点分离**：VNode 描述、Instance 实现
2. **可测试性**：无状态 VNode 易于单元测试
3. **可序列化**：Intent 替代闭包支持 diffing
4. **可复用**：Behavior 模式复用通用行为
5. **可维护**：清晰的架构分层

遵循本指南，可以系统地将现有组件迁移到 Fiber-first 架构。
