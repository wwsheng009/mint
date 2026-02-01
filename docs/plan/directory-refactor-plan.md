# Mint UI 目录结构重构计划

> **创建日期**: 2026-02-01
> **目标**: 实现清晰、模块化、低耦合的分层架构
> **参考文档**: framework/docs/ui/design/SYSTEM_ARCHITECTURE.md

---

## 目录

1. [当前状态分析](#一当前状态分析)
2. [目标架构设计](#二目标架构设计)
3. [详细文件迁移计划](#三详细文件迁移计划)
4. [解耦与接口设计](#四解耦与接口设计)
5. [实施步骤](#五实施步骤)
6. [验证与测试](#六验证与测试)
7. [风险评估](#七风险评估)

---

## 一、当前状态分析

### 1.1 当前目录结构

```
mint/
├── ui/                     # 问题：职责过重
│   ├── app.go              # 1933 行，包含大量实现细节
│   ├── fiber.go            # Fiber 架构实现
│   ├── reconciler.go       # 协调器实现
│   ├── begin_work.go       # BeginWork 阶段
│   ├── complete_work.go    # CompleteWork 阶段
│   ├── diff.go             # Diff 算法
│   ├── scheduler.go        # UI 调度器
│   ├── instance.go         # 组件实例
│   ├── instance_manager.go # 实例管理器
│   ├── interaction_state.go # 交互状态
│   ├── vnode.go            # VNode 接口 (应保留)
│   ├── component.go        # 组件节点 (应保留)
│   ├── element.go          # 元素节点 (应保留)
│   ├── hooks.go            # Hooks API (应保留)
│   └── [各组件文件]         # 组件实现 (应保留)
│
├── framework/              # 框架层
│   ├── app.go
│   ├── component/
│   ├── event/
│   └── ...
│
└── runtime/                # 运行时层
    ├── paint/
    ├── event/
    └── ...
```

### 1.2 核心问题识别

| 问题 | 描述 | 影响 |
|------|------|------|
| **ui/ 职责过重** | 既包含公开 API 又包含核心实现 | 难以独立测试，边界不清 |
| **实现细节暴露** | fiber.go, reconciler.go 等暴露给用户 | API 表面过大，维护困难 |
| **依赖关系混乱** | ui/ 直接依赖 runtime/，跳过 framework | 违反分层原则 |
| **代码重复** | framework/ 和 runtime/ 有功能重叠 | 维护成本高 |

### 1.3 当前依赖关系

```
ui/ (app.go 1933行)
  ├─→ framework/
  │    └─→ runtime/
  └─→ runtime/ (直接依赖，跳过 framework)
```

### 1.4 需要迁移的文件统计

| 分类 | 文件数 | 代码行数 (估算) |
|------|--------|----------------|
| Reconciler 系统 | 5 | ~1500 |
| State 管理 | 3 | ~600 |
| Scheduler | 1 | ~200 |
| **合计** | **9** | **~2300** |

---

## 二、目标架构设计

### 2.1 目标目录结构

```
mint/
├── internal/               # 新增：核心系统实现（不对外暴露）
│   ├── reconciler/         # Fiber 协调器
│   │   ├── fiber.go
│   │   ├── reconciler.go
│   │   ├── begin_work.go
│   │   ├── complete_work.go
│   │   ├── diff.go
│   │   └── public.go       # 公开接口
│   │
│   ├── scheduler/          # 调度器
│   │   └── ui_scheduler.go
│   │
│   └── state/              # 状态系统
│       ├── instance.go
│       ├── instance_manager.go
│       ├── interaction_state.go
│       └── public.go       # 公开接口
│
├── ui/                     # 精简：只保留公开 API
│   ├── app.go              # 精简后 ~200 行
│   ├── vnode.go
│   ├── component.go
│   ├── element.go
│   ├── fragment.go
│   ├── text.go
│   ├── hooks.go
│   ├── layout.go
│   ├── absolute.go
│   ├── grid.go
│   ├── button.go
│   ├── input.go
│   ├── checkbox.go
│   ├── select.go
│   ├── textarea.go
│   ├── progress.go
│   ├── modal.go
│   ├── tooltip.go
│   ├── virtuallist.go
│   ├── memory_safety.go
│   └── validator.go
│
├── framework/              # 保持不变
└── runtime/                # 保持不变
```

### 2.2 目标依赖关系

```
ui/ (公开 API 层)
  ↓
internal/ (内部实现层)
  ├─ reconciler/
  ├─ scheduler/
  └─ state/
  ↓
framework/ (框架层)
  ↓
runtime/ (运行时层)
```

### 2.3 职责划分

| 层级 | 职责 | 包含内容 |
|------|------|----------|
| **ui/** | 公开 API | VNode 接口、组件构造器、Hooks API |
| **internal/reconciler/** | 协调算法 | Fiber 架构、Diff 算法、工作循环 |
| **internal/scheduler/** | 调度逻辑 | 优先级调度、时间切片 |
| **internal/state/** | 状态管理 | 组件实例、交互状态 |
| **framework/** | 框架服务 | 应用框架、事件路由、主题 |
| **runtime/** | 底层运行时 | 渲染、输入、平台适配 |

---

## 三、详细文件迁移计划

### 3.1 迁移到 internal/reconciler/

#### 3.1.1 ui/fiber.go → internal/reconciler/fiber.go

**原文件内容**：
- Fiber 结构体定义
- Lane 和 EffectFlag 类型
- Fiber 创建和遍历函数
- Fiber 工具函数

**迁移后改动**：
```go
// 修改 package 声明
- package ui
+ package reconciler

// 添加对 ui 包的依赖
import "github.com/wwsheng009/mint/ui"
```

**导出的类型**：
- `Fiber` - 结构体
- `Lane` - 类型别名
- `EffectFlag` - 类型别名
- `CreateFiber()` - 函数
- `CreateFiberFromVNode()` - 函数
- `CloneFiber()` - 函数
- `MergeLanes()` - 函数

#### 3.1.2 ui/reconciler.go → internal/reconciler/reconciler.go

**原文件内容**：
- Reconciler 结构体定义
- NewReconciler() 构造函数
- Render() 方法
- ScheduleUpdate() 方法
- workLoopSync() 方法
- CommitRoot() 方法

**迁移后改动**：
```go
// 修改 package 声明
- package ui
+ package reconciler

// 更新导入
import (
    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/framework/component"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/ui"  // VNode 类型
+   "github.com/wwsheng009/mint/internal/state"  // 实例管理器
)

// 更新字段类型
type Reconciler struct {
    // ...
-   instanceMgr         *InstanceManager
+   instanceMgr         *state.InstanceManager
-   interactionStateMgr *InteractionStateManager
+   interactionStateMgr *state.InteractionStateManager
}
```

#### 3.1.3 ui/begin_work.go → internal/reconciler/begin_work.go

**原文件内容**：
- BeginWork() 函数
- 组件渲染逻辑

**迁移后改动**：
```go
- package ui
+ package reconciler

import (
    "github.com/wwsheng009/mint/ui"
+   "github.com/wwsheng009/mint/internal/state"
)
```

#### 3.1.4 ui/complete_work.go → internal/reconciler/complete_work.go

**原文件内容**：
- CompleteWork() 函数
- 副作用收集逻辑

**迁移后改动**：
```go
- package ui
+ package reconciler

import (
    "github.com/wwsheng009/mint/ui"
)
```

#### 3.1.5 ui/diff.go → internal/reconciler/diff.go

**原文件内容**：
- Diff 算法实现
- VNode 比较逻辑

**迁移后改动**：
```go
- package ui
+ package reconciler

import (
    "github.com/wwsheng009/mint/ui"
)
```

#### 3.1.6 新增：internal/reconciler/public.go

**新增文件** - 定义公开接口：

```go
package reconciler

import (
    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/framework/component"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/internal/state"
)

// Reconciler 接口 - 协调器公开接口
type Reconciler interface {
    // Render 执行渲染过程
    Render(ctx component.PaintContext, buffer *paint.Buffer, renderFunc func() ui.VNode)

    // ScheduleUpdate 调度状态更新
    ScheduleUpdate(lane Lane)

    // MarkDirty 标记需要重新渲染
    MarkDirty()

    // GetInstanceManager 获取实例管理器
    GetInstanceManager() *state.InstanceManager

    // GetInteractionStateManager 获取交互状态管理器
    GetInteractionStateManager() *state.InteractionStateManager

    // SetRenderCallback 设置渲染回调
    SetRenderCallback(cb RenderFunc)

    // Stats 获取统计信息
    Stats() map[string]interface{}
}

// RenderFunc 渲染函数类型
type RenderFunc func(vnode ui.VNode, x, y int, buffer *paint.Buffer)

// ReconcilerConfig 协调器配置
type ReconcilerConfig struct {
    TimeBudget      // 时间片预算
    EnableProfiling bool // 是否启用性能分析
    EnableFiber     bool // 是否启用 Fiber 协调
}

// NewReconciler 创建协调器
func NewReconciler(app *framework.App, rootComponent ui.ComponentFunc, config ReconcilerConfig) Reconciler
```

### 3.2 迁移到 internal/scheduler/

#### 3.2.1 ui/scheduler.go → internal/scheduler/ui_scheduler.go

**迁移后改动**：
```go
- package ui
+ package scheduler
```

### 3.3 迁移到 internal/state/

#### 3.3.1 ui/instance.go → internal/state/instance.go

**原文件内容**：
- ComponentInstance 接口
- BaseComponentInstance 实现

**迁移后改动**：
```go
- package ui
+ package state

import (
    "github.com/wwsheng009/mint/ui"  // VNode, ComponentFunc 等
)
```

**导出的类型**：
- `ComponentInstance` - 接口
- `BaseComponentInstance` - 结构体
- `NewBaseComponentInstance()` - 函数
- `NewBaseComponentInstanceWithProps()` - 函数

#### 3.3.2 ui/instance_manager.go → internal/state/instance_manager.go

**迁移后改动**：
```go
- package ui
+ package state

import (
    "github.com/wwsheng009/mint/ui"
)
```

**导出的类型**：
- `InstanceManager` - 结构体
- `NewInstanceManager()` - 函数

#### 3.3.3 ui/interaction_state.go → internal/state/interaction_state.go

**迁移后改动**：
```go
- package ui
+ package state
```

**导出的类型**：
- `InteractionStateManager` - 结构体
- `NewInteractionStateManager()` - 函数

#### 3.3.4 新增：internal/state/public.go

**新增文件** - 定义公开接口：

```go
package state

// InstanceManager 接口 - 实例管理器公开接口
type InstanceManager interface {
    Get(key string) ComponentInstance
    GetOrCreate(key string, factory func() ComponentInstance) ComponentInstance
    Cleanup(activeKeys []string)
    Count() int
}

// ComponentInstance 接口 - 组件实例公开接口
type ComponentInstance interface {
    GetContext() *ui.ComponentContext
    SetProps(props ui.Props)
    Render() ui.VNode
}

// InteractionStateManager 接口 - 交互状态管理器公开接口
type InteractionStateManager interface {
    SetFocused(key string, focused bool)
    IsFocused(key string) bool
    SetHovered(key string, hovered bool)
    IsHovered(key string) bool
}
```

### 3.4 ui/ 目录精简

#### 3.4.1 ui/app.go 精简

**当前状态**：1933 行，包含大量实现细节

**精简后** (~200 行)：

```go
package ui

import (
    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/internal/reconciler"
)

// Option 配置应用
type Option func(*Options)

// Options 应用配置
type Options struct {
    Width          int
    Height         int
    Title          string
    FPS            int
    EnableDevTools bool
}

// WithWidth 设置宽度
func WithWidth(width int) Option { /* ... */ }

// WithHeight 设置高度
func WithHeight(height int) Option { /* ... */ }

// WithTitle 设置标题
func WithTitle(title string) Option { /* ... */ }

// WithFPS 设置帧率
func WithFPS(fps int) Option { /* ... */ }

// Run 启动声明式 UI 应用
func Run(app ComponentFunc, opts ...Option) error {
    options := &Options{
        Width:  80,
        Height: 24,
        Title:  "Mint UI App",
        FPS:    60,
    }

    for _, opt := range opts {
        opt(options)
    }

    // 创建框架应用
    fwApp := framework.NewApp()
    fwApp.Resize(options.Width, options.Height)

    // 创建协调器
    r := reconciler.NewReconciler(fwApp, app, reconciler.ReconcilerConfig{
        TimeBudget:      5 * time.Millisecond,
        EnableProfiling: false,
        EnableFiber:     os.Getenv("MINT_USE_FIBER") == "true",
    })

    // 创建声明式根组件
    declarativeRoot := newDeclarativeRoot(app, fwApp, r)
    fwApp.SetRoot(declarativeRoot)

    return fwApp.Run()
}
```

#### 3.4.2 保留在 ui/ 的文件

这些文件不需要迁移，保留在 ui/ 作为公开 API：

| 文件 | 说明 |
|------|------|
| vnode.go | VNode 接口定义 |
| component.go | 组件节点实现 |
| element.go | 元素节点实现 |
| fragment.go | 片段节点实现 |
| text.go | 文本节点实现 |
| hooks.go | Hooks API (useState, useEffect 等) |
| layout.go | 布局组件 (HStack, VStack) |
| absolute.go | 绝对布局组件 |
| grid.go | 网格布局组件 |
| button.go | 按钮组件 |
| input.go | 输入组件 |
| checkbox.go | 复选框组件 |
| select.go | 选择器组件 |
| textarea.go | 多行输入组件 |
| progress.go | 进度条组件 |
| modal.go | 模态框组件 |
| tooltip.go | 提示框组件 |
| virtuallist.go | 虚拟列表组件 |
| memory_safety.go | 内存安全 API |
| validator.go | Key 验证器 |

---

## 四、解耦与接口设计

### 4.1 依赖注入模式

使用依赖注入减少模块间的直接依赖：

```go
// ui/app.go
func Run(app ComponentFunc, opts ...Option) error {
    fwApp := framework.NewApp()

    // 创建协调器时注入依赖
    r := reconciler.NewReconciler(
        fwApp,           // framework app
        app,             // root component
        config,          // configuration
    )
    // ...
}
```

### 4.2 接口隔离原则

定义小而专一的接口：

```go
// internal/reconciler/public.go
type Renderer interface {
    Render(ctx PaintContext, buffer *Buffer, renderFunc func() VNode)
}

type Scheduler interface {
    ScheduleUpdate(lane Lane)
    MarkDirty()
}

type StateProvider interface {
    GetInstanceManager() *state.InstanceManager
    GetInteractionStateManager() *state.InteractionStateManager
}
```

### 4.3 避免循环依赖

**原则**：internal/ 包可以依赖 ui/ 的类型（接口），但 ui/ 不能依赖 internal/ 的具体实现。

```go
// ✅ 正确：internal/ 依赖 ui/ 的接口
package reconciler

import "github.com/wwsheng009/mint/ui"

type Reconciler struct {
    rootComponent ui.ComponentFunc  // 使用接口
}

// ❌ 错误：ui/ 依赖 internal/ 的具体实现
package ui

import "github.com/wwsheng009/mint/internal/reconciler"

var r *reconciler.Reconciler  // 不要这样做

// ✅ 正确：ui/ 使用接口
package ui

type Reconciler interface {
    Render(...)
    ScheduleUpdate(...)
}
```

---

## 五、实施步骤

### 阶段 1: 准备工作 (第 1 天)

#### 1.1 创建目录结构

```bash
cd /path/to/mint
mkdir -p internal/reconciler
mkdir -p internal/scheduler
mkdir -p internal/state
```

#### 1.2 创建分支

```bash
git checkout -b refactor/directory-structure
git push -u origin refactor/directory-structure
```

#### 1.3 更新 .gitignore

确保忽略任何临时文件：

```gitignore
# 临时文件
*.tmp
*.bak
*~

# IDE
.idea/
.vscode/
```

### 阶段 2: 迁移状态系统 (第 2 天)

#### 2.1 迁移 instance.go

```bash
# 复制文件
cp ui/instance.go internal/state/instance.go

# 编辑 internal/state/instance.go
# - 修改 package ui → package state
# - 更新导入路径
```

```go
// internal/state/instance.go
package state

import (
    "github.com/wwsheng009/mint/ui"
)

// ComponentInstance 组件实例接口
type ComponentInstance interface {
    // ...
}

// BaseComponentInstance 基础组件实例实现
type BaseComponentInstance struct {
    // ...
}

// NewBaseComponentInstance 创建基础组件实例
func NewBaseComponentInstance(key string, fn ui.ComponentFunc) *BaseComponentInstance {
    // ...
}
```

#### 2.2 迁移 instance_manager.go

```bash
cp ui/instance_manager.go internal/state/instance_manager.go
```

#### 2.3 迁移 interaction_state.go

```bash
cp ui/interaction_state.go internal/state/interaction_state.go
```

#### 2.4 创建 public.go

```bash
# 创建 internal/state/public.go
# 定义公开接口
```

#### 2.5 更新 ui/ 中的引用

查找并更新所有使用这些类型的文件：

```bash
# 查找需要更新的文件
grep -r "InstanceManager" ui/
grep -r "InteractionStateManager" ui/
grep -r "ComponentInstance" ui/
```

更新导入和类型引用：

```go
// ui/app.go
import (
    "github.com/wwsheng009/mint/internal/state"
)

type declarativeRoot struct {
    // ...
-   instanceManager     *InstanceManager
+   instanceManager     *state.InstanceManager
}
```

#### 2.6 测试

```bash
go test ./internal/state/...
go test ./ui/...
```

### 阶段 3: 迁移协调器系统 (第 3-4 天)

#### 3.1 迁移 fiber.go

```bash
cp ui/fiber.go internal/reconciler/fiber.go
```

修改：
- package 声明
- 导入路径

#### 3.2 迁移 diff.go

```bash
cp ui/diff.go internal/reconciler/diff.go
```

#### 3.3 迁移 begin_work.go

```bash
cp ui/begin_work.go internal/reconciler/begin_work.go
```

#### 3.4 迁移 complete_work.go

```bash
cp ui/complete_work.go internal/reconciler/complete_work.go
```

#### 3.5 迁移 reconciler.go

```bash
cp ui/reconciler.go internal/reconciler/reconciler.go
```

更新类型引用：

```go
// internal/reconciler/reconciler.go
import (
    "github.com/wwsheng009/mint/internal/state"
)

type Reconciler struct {
    instanceMgr         *state.InstanceManager
    interactionStateMgr *state.InteractionStateManager
}
```

#### 3.6 创建 public.go

```bash
# 创建 internal/reconciler/public.go
```

#### 3.7 更新 ui/ 中的引用

```bash
# 查找需要更新的文件
grep -r "Reconciler" ui/
grep -r "Fiber" ui/
```

#### 3.8 测试

```bash
go test ./internal/reconciler/...
go test ./ui/...
```

### 阶段 4: 迁移调度器 (第 5 天)

#### 4.1 迁移 scheduler.go

```bash
cp ui/scheduler.go internal/scheduler/ui_scheduler.go
```

#### 4.2 更新引用

```bash
grep -r "Lane" ui/
```

#### 4.3 测试

```bash
go test ./internal/scheduler/...
go test ./ui/...
```

### 阶段 5: 精简 ui/app.go (第 6 天)

#### 5.1 提取声明式根组件

创建 `ui/declarative_root.go`：

```go
package ui

import (
    "github.com/wwsheng009/mint/internal/reconciler"
)

// declarativeRoot 声明式根组件
type declarativeRoot struct {
    component.Node
    reconciler          reconciler.Reconciler  // 使用接口
    // ...
}

func newDeclarativeRoot(fn ComponentFunc, app *framework.App, r reconciler.Reconciler) component.Node {
    // ...
}
```

#### 5.2 精简 app.go

只保留 Run() 函数和 Option 类型。

#### 5.3 测试

```bash
go test ./ui/...
```

### 阶段 6: 删除旧文件 (第 7 天)

#### 6.1 确认所有测试通过

```bash
go test ./...
```

#### 6.2 删除已迁移的文件

```bash
# 备份分支
git branch backup-before-cleanup

# 删除 ui/ 中已迁移的文件
rm ui/fiber.go
rm ui/reconciler.go
rm ui/begin_work.go
rm ui/complete_work.go
rm ui/diff.go
rm ui/scheduler.go
rm ui/instance.go
rm ui/instance_manager.go
rm ui/interaction_state.go
```

#### 6.3 最终测试

```bash
go test ./...
go build ./...
```

### 阶段 7: 文档更新 (第 8 天)

#### 7.1 更新 README.md

#### 7.2 更新 API 文档

#### 7.3 创建迁移指南

---

## 六、验证与测试

### 6.1 编译验证

```bash
# 编译所有包
go build ./...

# 编译示例
cd examples/counter && go build .
cd examples/demo && go build .
```

### 6.2 单元测试

```bash
# 运行所有测试
go test ./... -v

# 运行特定包的测试
go test ./internal/reconciler/... -v
go test ./internal/scheduler/... -v
go test ./internal/state/... -v
go test ./ui/... -v
```

### 6.3 集成测试

运行示例应用验证功能：

```bash
# Counter 示例
cd examples/counter && go run main.go

# Demo 示例
cd examples/demo && go run main.go

# Modal 示例
cd examples/modal && go run main.go

# 其他示例...
```

### 6.4 检查清单

- [ ] 所有包编译通过
- [ ] 所有单元测试通过
- [ ] 至少 5 个示例应用运行正常
- [ ] 没有循环依赖
- [ ] 公开 API 仍然可用
- [ ] 文档已更新

### 6.5 依赖检查

使用 `go mod` 检查依赖：

```bash
go mod tidy
go mod verify
```

检查循环依赖：

```bash
go list -json ./... | jq -r 'select(.Imports != null) | .ImportPath + " -> " + (.Imports | join(", "))'
```

---

## 七、风险评估

### 7.1 风险矩阵

| 风险 | 概率 | 影响 | 等级 | 缓解措施 |
|------|------|------|------|----------|
| 破坏现有 API | 中 | 高 | **高** | 保留所有公开函数，只移动实现 |
| 循环依赖 | 中 | 高 | **高** | 使用接口隔离，internal/ 只依赖 ui/ 接口 |
| 测试失败 | 低 | 中 | 中 | 先迁移测试，再迁移实现 |
| 性能下降 | 低 | 低 | 低 | 性能基准测试对比 |
| 文档不同步 | 中 | 低 | 低 | 同步更新文档 |

### 7.2 回滚计划

如果重构出现问题，可以快速回滚：

```bash
# 回滚到迁移前
git checkout main

# 或使用备份分支
git checkout backup-before-cleanup
```

### 7.3 分支策略

```
main (稳定)
  └── refactor/directory-structure (重构分支)
       └── backup-before-cleanup (清理前备份)
```

---

## 八、时间估算

| 阶段 | 任务 | 预计时间 |
|------|------|----------|
| 1 | 准备工作 | 0.5 天 |
| 2 | 迁移状态系统 | 1 天 |
| 3 | 迁移协调器系统 | 2 天 |
| 4 | 迁移调度器 | 0.5 天 |
| 5 | 精简 ui/app.go | 1 天 |
| 6 | 删除旧文件 | 0.5 天 |
| 7 | 文档更新 | 1 天 |
| 8 | 测试与验证 | 1 天 |
| **总计** | | **7.5 天** |

---

## 九、预期效果

### 9.1 代码组织

**重构前**：
```
ui/
├── app.go (1933 行) ❌
├── fiber.go (444 行) ❌
├── reconciler.go (348 行) ❌
├── [组件文件] ✅
└── ...
```

**重构后**：
```
ui/
├── app.go (~200 行) ✅
├── declarative_root.go (~300 行) ✅
├── [组件文件] ✅
└── ...

internal/
├── reconciler/
│   ├── fiber.go (444 行)
│   ├── reconciler.go (348 行)
│   └── ...
├── state/
└── scheduler/
```

### 9.2 依赖关系

**重构前**：
```
ui/ → framework/
ui/ → runtime/ (直接依赖)
```

**重构后**：
```
ui/ → internal/ → framework/ → runtime/
```

### 9.3 可测试性

- **重构前**：难以独立测试协调器逻辑
- **重构后**：可以直接测试 internal/reconciler/ 包

---

## 十、后续改进

重构完成后，可以考虑以下改进：

1. **性能优化**
   - 添加性能基准测试
   - 优化热点代码

2. **文档完善**
   - 添加包级别文档
   - 添加更多示例

3. **测试覆盖**
   - 提高测试覆盖率到 80%+
   - 添加集成测试

4. **API 改进**
   - 评估公开 API 的易用性
   - 添加更多便捷函数

---

**文档版本**: v1.0
**最后更新**: 2026-02-01
