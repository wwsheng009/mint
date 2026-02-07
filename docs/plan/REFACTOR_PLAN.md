# Mint UI 目录结构重构计划

## 目标

将当前架构重构为清晰、模块化、低耦合的分层结构，使：
- **ui/** 仅保留公开 API 层
- **internal/** 保留核心实现细节
- **framework/** 保留框架层
- **runtime/** 保留底层运行时

---

## 一、当前问题分析

### 1.1 ui/ 目录职责过重

当前 ui/ 包含：
- **公开 API** (应该保留): vnode.go, component.go, element.go, hooks.go
- **核心实现** (应该移到 internal/): fiber.go, reconciler.go, begin_work.go, complete_work.go, diff.go
- **状态管理** (应该移到 internal/): instance.go, instance_manager.go, interaction_state.go
- **调度器** (应该移到 internal/): scheduler.go
- **具体组件** (应该保留): button.go, input.go, checkbox.go, layout.go 等

### 1.2 当前依赖关系

```
ui/ (1933行 app.go + 实现细节)
  ↓
framework/
  ↓
runtime/
```

问题：ui/ 暴露了太多实现细节，难以独立测试和扩展。

---

## 二、目标目录结构

```
mint/
├── cmd/
│   └── mint-debugger/      # 调试器入口 (已存在)
│
├── internal/               # 核心系统实现 (新增)
│   ├── reconciler/         # Fiber 协调器
│   │   ├── fiber.go        # 从 ui/fiber.go 迁移
│   │   ├── reconciler.go   # 从 ui/reconciler.go 迁移
│   │   ├── begin_work.go   # 从 ui/begin_work.go 迁移
│   │   ├── complete_work.go # 从 ui/complete_work.go 迁移
│   │   ├── diff.go         # 从 ui/diff.go 迁移
│   │   └── public.go       # 公开接口定义 (新增)
│   │
│   ├── scheduler/          # 调度器
│   │   └── ui_scheduler.go # 从 ui/scheduler.go 迁移
│   │
│   └── state/              # 状态系统
│       ├── instance.go     # 从 ui/instance.go 迁移
│       ├── instance_manager.go # 从 ui/instance_manager.go 迁移
│       └── interaction_state.go # 从 ui/interaction_state.go 迁移
│
├── ui/                     # 对外 SDK (精简后)
│   ├── app.go              # 保留 Run() 入口，精简实现
│   ├── vnode.go            # VNode 接口定义
│   ├── component.go        # 组件节点
│   ├── element.go          # 元素节点
│   ├── fragment.go         # 片段节点
│   ├── text.go             # 文本节点
│   ├── hooks.go            # Hooks API
│   ├── layout.go           # 布局组件
│   ├── absolute.go         # 绝对布局
│   ├── grid.go             # 网格布局
│   ├── button.go           # 按钮组件
│   ├── input.go            # 输入组件
│   ├── checkbox.go         # 复选框
│   ├── select.go           # 选择器
│   ├── textarea.go         # 多行输入
│   ├── progress.go         # 进度条
│   ├── modal.go            # 模态框
│   ├── tooltip.go          # 提示框
│   ├── virtuallist.go      # 虚拟列表
│   ├── memory_safety.go    # 内存安全 API
│   └── validator.go        # 验证器
│
├── framework/              # 框架层 (保持不变)
│   ├── app.go              # 应用框架
│   ├── component/          # 组件框架
│   ├── binding/            # 数据绑定
│   ├── event/              # 事件系统
│   ├── layout/             # 布局框架
│   ├── theme/              # 主题系统
│   └── testing/            # 测试工具
│
├── runtime/                # 运行时层 (保持不变)
│   ├── paint/              # 渲染管线
│   ├── layout/             # 布局引擎
│   ├── event/              # 事件系统
│   ├── scheduler/          # 底层调度器
│   ├── platform/           # 平台适配
│   ├── style/              # 样式
│   └── input/              # 输入处理
│
├── devtools/               # 开发工具 (保持不变)
├── examples/               # 示例应用
└── docs/                   # 文档
```

---

## 三、文件迁移清单

### 3.1 迁移到 internal/reconciler/

| 源文件 | 目标文件 | 包名 | 说明 |
|--------|----------|------|------|
| ui/fiber.go | internal/reconciler/fiber.go | reconciler | Fiber 数据结构和算法 |
| ui/reconciler.go | internal/reconciler/reconciler.go | reconciler | 协调器核心逻辑 |
| ui/begin_work.go | internal/reconciler/begin_work.go | reconciler | BeginWork 阶段 |
| ui/complete_work.go | internal/reconciler/complete_work.go | reconciler | CompleteWork 阶段 |
| ui/diff.go | internal/reconciler/diff.go | reconciler | Diff 算法 |

### 3.2 迁移到 internal/scheduler/

| 源文件 | 目标文件 | 包名 | 说明 |
|--------|----------|------|------|
| ui/scheduler.go | internal/scheduler/ui_scheduler.go | scheduler | UI 调度器适配器 |

### 3.3 迁移到 internal/state/

| 源文件 | 目标文件 | 包名 | 说明 |
|--------|----------|------|------|
| ui/instance.go | internal/state/instance.go | state | 组件实例 |
| ui/instance_manager.go | internal/state/instance_manager.go | state | 实例管理器 |
| ui/interaction_state.go | internal/state/interaction_state.go | state | 交互状态管理器 |

### 3.4 保留在 ui/ 的文件

| 文件 | 操作 | 说明 |
|------|------|------|
| ui/vnode.go | 保留 | VNode 接口定义 |
| ui/component.go | 保留 | 组件节点 |
| ui/element.go | 保留 | 元素节点 |
| ui/fragment.go | 保留 | 片段节点 |
| ui/text.go | 保留 | 文本节点 |
| ui/hooks.go | 保留 | Hooks API |
| ui/layout.go | 保留 | 布局容器组件 |
| ui/absolute.go | 保留 | 绝对布局组件 |
| ui/grid.go | 保留 | 网格布局组件 |
| ui/button.go | 保留 | 按钮组件 |
| ui/input.go | 保留 | 输入组件 |
| ui/checkbox.go | 保留 | 复选框组件 |
| ui/select.go | 保留 | 选择器组件 |
| ui/textarea.go | 保留 | 多行输入组件 |
| ui/progress.go | 保留 | 进度条组件 |
| ui/modal.go | 保留 | 模态框组件 |
| ui/tooltip.go | 保留 | 提示框组件 |
| ui/virtuallist.go | 保留 | 虚拟列表组件 |
| ui/memory_safety.go | 保留 | 内存安全 API |
| ui/validator.go | 保留 | 验证器 |
| ui/app.go | 精简 | 只保留 Run() 和 Options |

### 3.5 新增文件

| 文件 | 说明 |
|------|------|
| internal/reconciler/public.go | 定义公开接口，供 ui/ 使用 |
| internal/state/public.go | 状态系统公开接口 |
| internal/README.md | internal 包说明文档 |

---

## 四、解耦方案

### 4.1 新的依赖关系

```
ui/ (公开 API)
  ↓
internal/reconciler, internal/scheduler, internal/state
  ↓
framework/ (组件框架)
  ↓
runtime/ (底层运行时)
```

### 4.2 接口隔离

新增 `internal/reconciler/public.go`:

```go
package reconciler

// Reconciler 接口 - 供 ui/ 使用
type Reconciler interface {
    Render(ctx PaintContext, buffer *Buffer, renderFunc func() VNode)
    ScheduleUpdate(lane Lane)
    MarkDirty()
    GetInstanceManager() *state.InstanceManager
    GetInteractionStateManager() *state.InteractionStateManager
}

// NewReconciler 创建协调器
func NewReconciler(app *framework.App, rootComponent ui.ComponentFunc, config ReconcilerConfig) Reconciler
```

### 4.3 ui/app.go 精简

精简后的 ui/app.go 结构:

```go
package ui

import (
    "github.com/wwsheng009/mint/framework"
    "github.com/wwsheng009/mint/internal/reconciler"
    "github.com/wwsheng009/mint/internal/state"
)

// Run 启动声明式 UI 应用
func Run(app ComponentFunc, opts ...Option) error {
    // 创建框架应用
    fwApp := framework.NewApp()

    // 创建内部协调器
    r := reconciler.NewReconciler(fwApp, app, ...)

    // 设置根组件...
    return fwApp.Run()
}
```

---

## 五、实施步骤

### 阶段 1: 创建目录结构

```bash
mkdir -p internal/reconciler
mkdir -p internal/scheduler
mkdir -p internal/state
```

### 阶段 2: 迁移核心文件

按以下顺序迁移文件：

1. **迁移状态系统** (最底层，无依赖)
   - ui/instance.go → internal/state/instance.go
   - ui/instance_manager.go → internal/state/instance_manager.go
   - ui/interaction_state.go → internal/state/interaction_state.go

2. **迁移协调器** (依赖状态系统)
   - ui/fiber.go → internal/reconciler/fiber.go
   - ui/diff.go → internal/reconciler/diff.go
   - ui/begin_work.go → internal/reconciler/begin_work.go
   - ui/complete_work.go → internal/reconciler/complete_work.go
   - ui/reconciler.go → internal/reconciler/reconciler.go

3. **迁移调度器**
   - ui/scheduler.go → internal/scheduler/ui_scheduler.go

### 阶段 3: 更新导入路径

更新所有受影响文件的 import 语句：

```go
// 旧的导入
import "github.com/wwsheng009/mint/ui"

// 新的导入
import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/internal/reconciler"
    "github.com/wwsheng009/mint/internal/state"
)
```

### 阶段 4: 添加公开接口

创建 `internal/reconciler/public.go` 和 `internal/state/public.go`，定义公开接口。

### 阶段 5: 更新 ui/app.go

将 ui/app.go 精简，只保留 Run() 入口和 Options 定义。

### 阶段 6: 运行测试

```bash
# 运行所有测试
go test ./...

# 运行示例验证
cd examples/counter && go run main.go
cd examples/demo && go run main.go
```

### 阶段 7: 删除旧文件

确认所有测试通过后，删除 ui/ 中已迁移的文件。

---

## 六、验证步骤

1. **编译验证**: `go build ./...`
2. **单元测试**: `go test ./...`
3. **示例运行**: 运行 examples/ 中的多个示例
4. **导入检查**: 确保没有循环依赖
5. **API 兼容**: 确认公开 API (ui.Run, ui.Text 等) 仍然可用

---

## 七、风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| 破坏现有 API | 保留 ui/ 中的所有公开函数，只移动实现 |
| 循环依赖 | 使用接口隔离，让 internal/ 只依赖 ui/ 的接口 |
| 测试失败 | 先迁移测试文件，再迁移实现 |
| 文档过时 | 同步更新所有文档 |

---

## 八、关键文件清单

需要修改的关键文件：

1. **ui/app.go** - 精简为纯入口
2. **ui/reconciler.go** - 迁移到 internal/
3. **ui/fiber.go** - 迁移到 internal/
4. **go.mod** - 确认模块路径

---

## 九、预期效果

重构后的架构：

- **ui/** - 纯公开 API，约 2000 行代码（组件 + API）
- **internal/reconciler/** - Fiber 协调器实现
- **internal/scheduler/** - 调度器实现
- **internal/state/** - 状态管理实现
- **framework/** - 框架层
- **runtime/** - 运行时层

依赖方向：ui → internal → framework → runtime

清晰的职责边界，便于测试、扩展和维护。
