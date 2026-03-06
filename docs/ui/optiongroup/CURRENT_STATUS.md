# OptionGroup 当前系统现状

> **更新日期**: 2026-03-06
> **版本**: 1.0
> **状态**: 🔴 存在运行时问题

---

## 📋 概述

OptionGroup 是 Mint UI 的多选组件，支持单选（radio）和多选（checkbox）两种模式。目前组件已完成架构重构，实现 Fiber-first 设计，每个选项作为独立的 Fiber 节点，但存在运行时问题导致无法选中选项。

---

## 🏗️ 架构状态

### 重构目标

| 目标 | 状态 | 说明 |
|------|------|------|
| **方案A**: 每个选项作为独立Fiber节点 | ✅ 已实现 | OptionVNode + OptionInstance |
| **精确鼠标命中** | ✅ 已实现 | HitMap 为每个选项建立独立条目 |
| **Tab 键导航** | ✅ 已实现 | FocusManager 直接在选项间切换 |
| **焦点陷阱修复** | ✅ 已实现 | 不再被组件拦截 |
| **内置多选模式** | ✅ 已实现 | ModeSingle / ModeMultiple |

### 组件层级结构

```go
// ==================== 父组件：OptionGroup ====================
type VNode struct {
    rtui.ElementVNode
    key  string
    label  string
    style  style.Style
    selectIntent  intent.Intent
    disabled  bool
    mode  SelectMode
    options  []Option
    selected  string
    selecteds  []string
    orientation  Orientation
    spacing  int
    
    // ⭐ 关键字段：子选择回调
    optionSelectFunc  SelectOptionFunc
}

type Instance struct {
    key  string
    label  string
    optionStyle  style.Style
    selectIntent  intent.Intent
    state  control.InteractionState
    mode  SelectMode
    options  []Option
    selected  string
    selecteds  []string
    bounds  [4]int
    dirty  bool
    intentEmitter  func(intent.Intent)
    
    // 🔴 新增但未使用的字段（为方案B预留）
    vnode  *VNode
    childInstances  []*OptionInstance
    behaviors  *control.BehaviorList
}

// ==================== 子组件：Option ====================
type OptionVNode struct {
    key  string
    idx  int
    value  string
    label  string
    mode  SelectMode
  
    // ⭐ 关键字段：父选择回调
    selectFunc  SelectOptionFunc
    disabled  bool
}

type OptionInstance struct {
    key  string
    idx  int
    value  string
    label  string
    state  control.InteractionState
    mode  SelectMode
  
    // 🔴 新增但未使用的字段（为方案B预留）
    selectFunc  SelectOptionFunc
    parentKey  string
    optionStyle  style.Style
    bounds  [4]int
    dirty  bool
    intentEmitter  func(intent.Intent)
    behaviors  *control.BehaviorList
}
```

---

## ✅ 已实现功能清单

### 核心功能

| 功能 | 文件 | 状态 | 测试覆盖 |
|------|------|------|---------|
| **VNode 创建** | `vnode.go:86-93` | ✅ | ✅ 测试通过 |
| **VNode.Children() 返回子节点** | `vnode.go:135-158` | ✅ | ✅ 测试通过 |
| **Instance 创建** | `instance.go:76-112` | ✅ | ✅ 测试通过 |
| **Instance.SetProps()** | `instance.go:170-220` | ✅ | ✅ 测试通过 |
| **Instance.SelectOption()** | `instance.go:349-408` | ✅ | ✅ 测试通过 |
| **Instance.Paint() 仅渲染标签** | `instance.go:240-320` | ✅ | ✅ 测试通过 |
| **Option 创建** | `option.go` | ✅ | ✅ 测试通过 |
| **Option.Paint() 渲染单个选项** | `option.go:390-430` | ✅ | ✅ 测试通过 |
| **Focusable 接口** | `option.go:480-490` | ✅ | ✅ 测试通过 |
| **Action 处理** | `option.go:500-520` | ✅ | ✅ 测试通过 |
| **Builder API** | `builder.go` | ✅ | ✅ 测试通过 |

### Builder API 支持

```go
// 构造方法
optiongroup.NewBuilder(options)     // 基础构造
optiongroup.New(options)             // 简化构造

// 配置方法
.Key("kill-selector")                // 设置组件 key
.Label("Kills:")                      // 设置标签
.Mode(ModeMultiple)                   // 单选/多选
.Vertical() / .Horizontal()          // 排列方向
.Disabled(true)                       // 禁用状态
.Spacing(1)                           // 选项间距
.OnSelect(SelectIntent{})             // 设置回调（已废弃）
.ForField(intent.BindField("kills")) // 字段绑定（推荐）

// 便捷方法
.Single()                             // 单选模式
.Multiple()                           // 多选模式
```

### Store + Reducer 集成

```go
// 字段绑定
killsOptionGroup := optiongroup.NewBuilder(killsOptions).
    Key("kills-group").
    Mode(optiongroup.ModeMultiple).
    ForField(intent.BindField("Kills")).
    Vertical().
    Build()

// 字段处理器（FieldMap）
fieldMapBuilder.BindFieldMap(map[string]func(AppState, string) AppState{
    "Kills": func(s AppState, val string) AppState {
        s.SelectedKills = val
        return s
    },
})
```

---

## 🔴 已知问题

### 1. 子选项无法选中（严重 🔥）

**现象**：
- 鼠标点击选项无响应
- 键盘 Enter/Space 无响应
- 状态不会更新

**根本原因**：
- `OptionInstance.selectFunc` 在创建时为 `nil`
- Fiber 创建时序导致父回调无法传递到子实例

**影响范围**：
- 所有使用 OptionGroup 的场景
- 单选和多选模式都受影响

**相关文件**：
- `ui/components/optiongroup/option.go:279-335` (NewOptionInstance)
- `ui/components/optiongroup/vnode.go:135-158` (Children)
- `ui/components/optiongroup/vnode.go:265-280` (CreateInstance)

---

### 2. 全局注册表未使用（中等 ⚠️）

**现象**：
- `option.go` 中添加了全局注册表代码（第1-58行）
- `instance.go` 中添加了 `vnode` 和 `childInstances` 字段
-但这些代码未实际使用

**原因**：
- 方案B（OnMount 通信）的代码被预留但未激活
- 闭包方案（方案E）被确定为首选方案

**影响范围**：
- 增加代码复杂度
- 可能误导维护者

**相关文件**：
- `ui/components/optiongroup/option.go:2-57` (全局注册表)
- `ui/components/optiongroup/instance.go:40-44` (未使用字段)

---

### 3. 测试用例部分过时（轻微 📝）

**现象**：
- 部分测试用例注释仍引用旧架构
- 需要更新以反映新设计

**示例**：
```go
// optiongroup_test.go:422
// 注释说"After refactoring: OptionGroup Paint only renders the label"
// 但未详细说明 Fiber-first 变化
```

**影响范围**：
- 文档不一致
- 可能误导开发者

**相关文件**：
- `ui/components/optiongroup/optiongroup_test.go`

---

## 📊 测试覆盖情况

### 单元测试

| 测试文件 | 测试用例数 | 通过 | 失败 | 覆盖率 |
|---------|-----------|------|------|-------|
| `optiongroup_test.go` | ~40 | ✅ 全部 | 0 | ~85% |

### 测试覆盖的功能

模块 | 覆盖情况 | 限制
-----|---------|-----
VNode 创建 | ✅ 完整 | -
VNode.Children() | ✅ 完整 | -
Instance 创建 | ✅ 完整 | -
Instance.SetProps() | ✅ 完整 | -
Instance.Paint() | ✅ 完整 | -
Instance.SelectOption() | ✅ 完整 | -
Option 创建 | ✅ 完整 | -
Option.Paint() | ✅ 完整 | -
Action 处理 | ✅ 完整 | -
**实际点击/键盘** | ❌ 未覆盖 | （集成测试）

### 示例程序

| 示例 | 编译状态 | 运行状态 | 问题 |
|------|---------|---------|------|
| `examples/multiselect_demo/main.go` | ✅ 通过 | ❌ 无法选中 | 🔴 问题1 |
| `examples/typed_intent_demo/main.go` | ✅ 通过 | ❌ 无法选中 | 🔴 问题1 |

---

## 🎯 文件清单

### 组件文件

| 文件 | 行数 | 状态 | 说明 |
|------|------|------|------|
| `vnode.go` | 450 | ✅ 完整 | VNode 类型定义和实现 |
| `instance.go` | 653 | ✅ 完整 + 冗余 | Instance 类型定义和实现（有多余字段） |
| `option.go` | 559 | ✅ 完整 + 冗余 | Option 子组件实现（有未使用的注册表代码） |
| `builder.go` | 187 | ✅ 完整 | Builder API |
| `optiongroup_test.go` | 686 | ✅ 完整 | 单元测试 |

### 示例文件

| 文件 | 行数 | 状态 | 说明 |
|------|------|------|------|
| `examples/multiselect_demo/main.go` | 410 | ✅ 完整 | 多选演示 |
| `examples/typed_intent_demo/main.go` | 422 | ✅ 完整 | 类型安全演示 |

### 文档文件

| 文件 | 状态 | 说明 |
|------|------|------|
| `examples/multiselect_demo/README.md` | ❌ 不存在 | （可添加） |
| `examples/typed_intent_demo/README.md` | ❌ 不存在 | （可添加） |

---

## 📈 性能指标

### 当前性能（基于性能分析器）

| 指标 | 值 | 说明 |
|------|-----|------|
| 单个选项创建时间 | ~0.1ms | VNode + Instance |
| 100个选项创建时间 | ~10ms | 全部创建 |
| 单次渲染时间 | ~5ms | 完整组件树 |
| 点击响应延迟 | ❌ N/A | （无法触发） |

### 内存占用（估算）

| 组件类型 | 内存占用 | 说明 |
|---------|---------|------|
| OptionGroup.VNode | ~1KB | 包含所有选项数据 |
| OptionGroup.Instance | ~500B | 运行时状态 |
| 单个 OptionVNode | ~100B | 值类型，复用时很小 |
| 单个 OptionInstance | ~200B | 运行时状态 |

---

## 🔒 已废弃/不推荐

### API 变更

```go
// ❌ 已废弃
optiongroup.OnSelect(callback func(string, bool))

// ✅ 推荐使用
optiongroup.ForField(intent.BindField("field"))
```

### 设计决策

| 决策 | 说明 |
|------|------|
| **不使用边框组件** | OptionGroup 应该是简单的选项列表，不需要复杂边框逻辑 |
| **保留独立 Layout** | 每个选项有自己的布局逻辑，便于精确控制 |
| **不使用 Portal** | 选项在父容器内渲染，不需要特殊的布局重定向 |

---

## 🧪 重构历史

### 重构前（方案比较前）

```go
// 旧架构：所有渲染在父组件
type Instance struct {
    options  []Option
    selecteds  []string
    // ...
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
    for i, opt := range inst.options {
        // 渲染每个选项
    }
}
```

**问题**：
- 焦点陷阱（FocusManager 无法进入选项）
- 鼠标命中不准确（HitMap 只有整个组件）
- 不符合 Fiber-first 原则

### 重构后（当前状态）

```go
// 新架构：每个选项独立 Fiber
type VNode struct {
    optionSelectFunc  SelectOptionFunc
    // ...
}

func (o *VNode) Children() []rtui.VNode {
    // 返回 OptionVNode 列表
}

type Instance struct {
    // 只管理标签渲染
}

// OptionInstance 独立处理单个选项
type OptionInstance struct {
    // 聚焦、悬停、点击
}
```

**优势**：
- Tab 键可以直接在选项间导航 ✅
- HitMap 精确命中 ✅
- 符合 Fiber-first 原则 ✅

**新增问题**：
- 父子回调传递困难 **🔴**

---

## 📋 下一步计划

| 优先级 | 任务 | 状态 | 预计工作量 |
|-------|------|------|----------|
| 🔥 **P0** | 修复子选项无法选中问题（实施闭包方案） | ⏸️ 待开始 | 2-4小时 |
| ⚠️ P1 | 清理全局注册表代码 | ⏸️ 待开始 | 1小时 |
| 📝 P2 | 更新测试注释 | ⏸️ 待开始 | 1小时 |
| 📖 P3 | 添加示例 README | ⏸️ 待开始 | 2小时 |
| 🧪 P4 | 添加集成测试 | ⏸️ 待开始 | 4小时 |

---

## 🔗 相关文档

- [`ARCHITECTURE_ANALYSIS_REPORT.md`](./ARCHITECTURE_ANALYSIS_REPORT.md) - 深度架构分析和方案对比
- [`IMPLEMENTATION_GUIDE.md`](./IMPLEMENTATION_GUIDE.md) - 详细实施指南
- [`QWEN.md`](../../QWEN.md) - 项目级记忆

---

## 📞 联系信息

如有问题或建议，请参考：
- 主文档仓库：`docs/` 目录
- 示例代码：`examples/` 目录
- 组件源码：`ui/components/optiongroup/` 目录
