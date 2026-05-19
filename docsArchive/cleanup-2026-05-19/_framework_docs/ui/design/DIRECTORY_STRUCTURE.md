# Mint UI 目录结构设计方案

**版本**: v1.1
**日期**: 2026-02-01
**状态**: ⚠️ **文档包含未实施的重构计划**

---

## ⚠️ 重要说明

本文档包含两部分内容：
1. **当前实际目录结构**（已实现）
2. **计划中的重构方案**（未实施）

文档末尾的重构路径部分为规划内容，详见 `docs/plan/directory-refactor-plan.md`。

---

## 一、设计原则

### 1.1 分层原则

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│                   (用户代码 - examples/)                      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      SDK Layer                               │
│                   (对外 API - ui/)                            │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Framework Layer                            │
│              (声明式框架 - framework/)                        │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Runtime Layer                             │
│               (运行时核心 - runtime/)                         │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Platform Layer                             │
│              (平台抽象 - runtime/platform/)                   │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 可见性原则

| 目录 | 可见性 | 说明 |
|------|--------|------|
| `ui/` | 公开 | 对外 SDK，稳定的 API |
| `framework/` | 公开 | 框架层，可被用户直接使用 |
| `runtime/` | 内部 | 运行时核心，不直接暴露给用户 |
| `devtools/` | 内部 | 调试工具，开发时使用 |
| `internal/` | 私有 | 内部实现，禁止外部导入 |

### 1.3 过渡期原则

- 新旧代码可共存
- 通过 `_legacy` 后缀标记旧代码
- 通过 `_v2` 后缀标记新代码
- 逐步迁移，避免大爆炸式重构

---

## 二、最终目录结构

```
mint/
├── cmd/                            # 可执行程序入口
│   ├── mint/                       # 主程序
│   │   └── main.go
│   ├── mint-devtools/              # DevTools 服务器
│   │   └── main.go
│   └── examples/                   # 示例程序入口（可选）
│       └── ...
│
├── ui/                             # 🔴 对外 SDK (新建)
│   ├── app.go                      # Run(), RunWithOptions()
│   ├── vnode.go                    # VNode 接口和类型
│   ├── builder.go                  # Element(), Text(), Button()
│   ├── layout.go                   # HStack(), VStack(), Spacer()
│   ├── hooks.go                    # useState, useEffect 等
│   ├── context.go                  # useContext, createContext
│   ├── style.go                    # Style(), StyleProps
│   ├── key.go                      # Key() 用于列表 diff
│   └── options.go                  # 运行时选项
│
├── framework/                      # 🟡 声明式框架层 (改造)
│   ├── reconciler/                 # Reconciler 系统 (新建)
│   │   ├── diff.go                 # Diff 算法
│   │   ├── fiber.go                # Fiber 节点
│   │   ├── scheduler.go            # 调度器
│   │   ├── workloop.go             # 工作循环
│   │   ├── lanes.go                # 优先级定义
│   │   └── render_phase.go         # 渲染阶段
│   │
│   ├── hooks/                      # Hooks 实现 (新建)
│   │   ├── state.go                # useState
│   │   ├── effect.go               # useEffect
│   │   ├── context.go              # useContext, createContext
│   │   ├── memo.go                 # useMemo, useCallback
│   │   ├── ref.go                  # useRef, useImperativeHandle
│   │   ├── reducer.go              # useReducer
│   │   └── context.go              # Context 实现
│   │
│   ├── component/                  # 组件系统 (改造)
│   │   ├── vnode/                  # VNode 实现 (新建)
│   │   │   ├── element.go          # 元素节点
│   │   │   ├── text.go             # 文本节点
│   │   │   ├── fragment.go         # 片段节点
│   │   │   ├── component.go        # 组件节点
│   │   │   └── props.go            # Props 定义
│   │   │
│   │   ├── base.go                 # 基础组件 (保留)
│   │   ├── container.go            # 容器组件 (改造)
│   │   ├── capabilities.go         # 能力接口 (保留)
│   │   ├── context.go              # 组件上下文 (保留)
│   │   └── state_holder.go         # 状态持有者 (保留)
│   │
│   ├── layout/                     # 布局封装 (新建)
│   │   ├── hstack.go               # 水平布局
│   │   ├── vstack.go               # 垂直布局
│   │   ├── stack.go                # 通用堆叠
│   │   ├── spacer.go               # 弹性空间
│   │   ├── align.go                # 对齐方式
│   │   ├── virtual.go              # 虚拟化列表 (复用 runtime/layout)
│   │   ├── grid.go                 # Grid 布局 (新增)
│   │   ├── grid_algorithm.go       # Grid 算法 (新增)
│   │   ├── grid_cache.go           # Grid 缓存 (新增)
│   │   ├── dimension.go            # Dimension 类型 (新增)
│   │   ├── absolute.go             # Absolute 定位 (新增)
│   │   ├── absolute_algorithm.go   # Absolute 算法 (新增)
│   │   ├── absolute_builder.go     # Absolute API (新增)
│   │   └── position.go             # Position 类型 (新增)
│   │
│   ├── render/                     # 渲染管线 (新建)
│   │   ├── drawcmd.go              # DrawCmd 定义
│   │   ├── rasterize.go            # 光栅化
│   │   ├── buffer_adapter.go       # Buffer 适配器
│   │   ├── diff.go                 # Buffer Diff
│   │   ├── terminal_state.go       # 终端状态跟踪 (新增)
│   │   ├── style_change.go         # 样式变化检测 (新增)
│   │   ├── rle.go                  # RLE 编码 (新增)
│   │   └── optimizer.go            # 渲染优化器 (新增)
│   │
│   ├── terminal/                   # 终端相关 (新建)
│   │   ├── ansi.go                 # ANSI 代码 (复用 runtime)
│   │   ├── driver.go               # 驱动适配
│   │   └── screen.go               # 屏幕管理
│   │
│   ├── event/                      # 事件系统 (改造)
│   │   ├── dispatcher.go           # 声明式事件分发 (新建)
│   │   ├── handler.go              # 事件处理器 (新建)
│   │   └── adapter.go              # 与 runtime 事件桥接 (新建)
│   │   ├── pump.go                 # 事件泵 (保留)
│   │   ├── keyboard.go             # 键盘工具 (保留)
│   │   └── handler_legacy.go       # 旧处理器 (标记废弃)
│   │
│   ├── style/                      # 样式系统 (保留)
│   │   ├── theme.go                # 主题定义
│   │   ├── token.go                # Design Token
│   │   └── palette.go              # 调色板
│   │
│   ├── state/                      # 状态系统 (改造)
│   │   ├── hooks_adapter.go        # Hooks 与状态桥接 (新建)
│   │   ├── global.go               # 全局状态 (改造)
│   │   ├── derived.go              # 派生状态 (保留)
│   │   └── reactive.go             # 反应式状态 (保留)
│   │
│   ├── animation/                  # 动画系统 (扩展)
│   │   ├── hooks.go                # useAnimation (新建)
│   │   ├── animator.go             # 动画器 (保留)
│   │   └── easing.go               # 缓动函数 (保留)
│   │
│   ├── remote/                     # 远程渲染 (新建)
│   │   ├── protocol.go             # 协议定义
│   │   ├── server.go               # 服务器
│   │   └── stream.go               # DrawCmd 流式传输
│   │
│   ├── layer/                      # Layer 系统 (新增)
│   │   ├── layer.go                # Layer 类型定义
│   │   ├── tree.go                 # LayerTree 实现
│   │   ├── manager.go              # Manager 实现
│   │   ├── event.go                # 事件处理
│   │   ├── render.go               # 渲染集成
│   │   ├── layout.go               # 布局处理
│   │   └── focus.go                # 焦点管理
│   │
│   ├── input/                      # 输入缓冲系统 (新增)
│   │   ├── buffer.go               # UTF-32 文本缓冲
│   │   ├── selection.go            # 选择区域管理
│   │   ├── cursor.go               # 光标移动
│   │   ├── line.go                 # 行操作
│   │   ├── scroll.go               # 滚动控制
│   │   └── history.go              # 撤销/重做
│   │
│   ├── scheduler/                  # 输入优先级调度 (新增)
│   │   ├── priority.go             # 优先级定义
│   │   ├── input_queue.go          # 输入队列
│   │   ├── input_handler.go        # 输入处理器
│   │   ├── interruptible.go        # 可中断任务
│   │   └── scheduler.go            # 核心调度器
│   │
│   ├── editor/                     # 编辑器组件 (新增)
│   │   ├── buffer.go               # UTF-32 文本缓冲
│   │   ├── cursor.go               # 光标管理
│   │   ├── selection.go            # 选择区
│   │   ├── scroll.go               # 滚动控制
│   │   ├── history.go              # 撤销/重做
│   │   ├── lexer.go                # 词法分析器接口 (新增)
│   │   ├── go_lexer.go             # Go 语言 Lexer (新增)
│   │   ├── js_lexer.go             # JavaScript Lexer (新增)
│   │   ├── incremental_lexer.go    # 增量 Lexer (新增)
│   │   ├── token.go                # Token 类型 (新增)
│   │   ├── token_style.go          # Token 样式 (新增)
│   │   └── highlight_render.go     # 高亮渲染 (新增)
│   │
│   ├── components/                 # 内置组件 (新建)
│   │   ├── text.go                 # Text
│   │   ├── button.go               # Button
│   │   ├── input.go                # TextInput
│   │   ├── checkbox.go             # Checkbox
│   │   ├── list.go                 # List
│   │   ├── table.go                # Table
│   │   ├── modal.go                # Modal
│   │   ├── progress.go             # ProgressBar
│   │   └── tabs.go                 # Tabs
│   │
│   ├── binding/                    # 数据绑定 (保留，可能废弃)
│   ├── display/                    # 显示组件 (保留，兼容)
│   ├── form/                       # 表单组件 (保留，兼容)
│   ├── input/                      # 输入组件 (保留，兼容)
│   ├── interactive/                # 交互组件 (保留，兼容)
│   ├── overlay/                    # 覆盖层 (保留)
│   ├── widget/                     # 小部件 (保留)
│   ├── util/                       # 工具函数 (保留)
│   └── v8/                         # V8 集成 (保留，独立模块)
│
├── runtime/                        # 🟢 运行时核心 (保持)
│   ├── platform/                   # 平台抽象
│   │   ├── terminal.go             # 终端接口
│   │   ├── windows.go              # Windows 实现
│   │   ├── linux.go                # Linux 实现
│   │   └── darwin.go               # macOS 实现
│   │
│   ├── layout/                     # 布局引擎 (保持)
│   │   ├── flexbox.go              # Flex 布局
│   │   ├── constraint.go           # 约束系统
│   │   ├── cache.go                # 布局缓存
│   │   └── measure.go              # 测量接口
│   │
│   ├── event/                      # 事件系统 (保持)
│   │   ├── event.go                # 事件定义
│   │   ├── types.go                # 事件类型
│   │   ├── phase.go                # 传播阶段
│   │   ├── dispatch.go             # 分发器
│   │   ├── hittest.go              # 命中测试
│   │   ├── filter.go               # 过滤器
│   │   └── handler.go              # 处理器
│   │
│   ├── paint/                      # 渲染系统 (保持)
│   │   ├── buffer.go               # 缓冲区
│   │   ├── cell.go                 # 单元格
│   │   ├── style.go                # 样式应用
│   │   ├── dirty.go                # 脏区域
│   │   └── output.go               # 输出
│   │
│   ├── focus/                      # 焦点管理 (保持)
│   │   ├── manager.go              # 焦点管理器
│   │   ├── scope.go                # 焦点域
│   │   └── navigation.go           # 导航
│   │
│   ├── input/                      # 输入处理 (保持)
│   │   ├── reader.go               # 输入读取
│   │   ├── parser.go               # 解析器
│   │   └── keymap.go               # 按键映射
│   │
│   ├── style/                      # 样式实现 (保持)
│   │   ├── color.go                # 颜色
│   │   ├── style.go                # 样式
│   │   └── ansi.go                 # ANSI 代码
│   │
│   ├── action/                     # Action 系统 (保持)
│   │   ├── action.go               # Action 定义
│   │   ├── dispatcher.go           # 分发器
│   │   └── types.go                # 类型
│   │
│   ├── state/                      # 状态管理 (保持)
│   │   ├── tracker.go              # 状态追踪
│   │   ├── snapshot.go             # 快照
│   │   └── history.go              # 历史记录
│   │
│   ├── scheduler/                  # 调度器 (保持)
│   │   ├── scheduler.go            # 基础调度器
│   │   └── queue.go                # 任务队列
│   │
│   ├── core/                       # 核心运行时 (保持)
│   │   ├── runtime.go              # Runtime 主结构
│   │   ├── context.go              # 上下文管理
│   │   └── lifecycle.go            # 生命周期
│   │
│   ├── animation/                  # 动画实现 (保持)
│   │   ├── timeline.go             # 时间轴
│   │   ├── easing.go               # 缓动函数
│   │   └── manager.go              # 动画管理器
│   │
│   ├── selection/                  # 选择管理 (保持)
│   │   ├── manager.go              # 选择管理器
│   │   └── range.go                # 范围
│   │
│   ├── registry/                   # 注册表 (保持)
│   │   └── registry.go
│   │
│   └── exports/                    # 导出类型 (保持)
│       └── exports.go
│
├── devtools/                       # 🔵 DevTools (保持，扩展)
│   ├── core/                       # 核心功能
│   │   ├── devtools.go             # 主入口
│   │   ├── types.go                # 类型定义
│   │   ├── collector.go            # 收集器
│   │   ├── bus.go                  # 事件总线
│   │   └── adapter.go              # 运行时适配
│   │
│   ├── protocol/                   # 协议层
│   │   ├── server.go               # 服务器
│   │   ├── websocket.go            # WebSocket
│   │   ├── message.go              # 消息定义
│   │   └── api.go                  # HTTP API
│   │
│   ├── client/                     # 客户端
│   │   ├── panel.go                # TUI 面板
│   │   └── visualizer.go           # 可视化
│   │
│   ├── observation/                # 观察层
│   │   └── layer.go
│   │
│   ├── snapshot/                   # 快照系统
│   │   ├── manager.go
│   │   ├── snapshot.go
│   │   └── diff.go
│   │
│   ├── memory/                     # 内存监控
│   │   └── monitor.go
│   │
│   ├── ui/                         # Web UI (新建)
│   │   └── dashboard/              # 仪表盘
│   │       ├── index.html
│   │       ├── app.js
│   │       └── styles.css
│   │
│   └── bridge/                     # 桥接层 (新建)
│       ├── fiber.go                # Fiber 树导出
│       ├── profiler.go             # 性能分析
│       └── inspector.go            # 组件检查器
│
├── examples/                       # 🟣 示例程序
│   ├── hello/                      # Hello World
│   │   └── main.go
│   ├── counter/                    # 计数器
│   │   └── main.go
│   ├── todo/                       # Todo 列表
│   │   └── main.go
│   ├── form/                       # 表单示例
│   │   └── main.go
│   ├── layout/                     # 布局示例
│   │   └── main.go
│   ├── animation/                  # 动画示例
│   │   └── main.go
│   └── dashboard/                  # 完整仪表盘
│       └── main.go
│
├── docs/                           # 📚 文档 (根目录)
│   ├── getting-started.md          # 快速开始
│   ├── api-reference.md            # API 参考
│   ├── tutorials/                  # 教程
│   │   ├── basics.md
│   │   ├── components.md
│   │   ├── state.md
│   │   └── animation.md
│   └── architecture/               # 架构文档
│       └── ...
│
├── framework/docs/                 # 设计文档 (保留)
│   └── ui/
│       ├── design/                 # 设计文档
│       │   ├── SYSTEM_ARCHITECTURE.md
│       │   ├── IMPLEMENTATION_GAP_ANALYSIS.md
│       │   ├── DIRECTORY_STRUCTURE.md (本文档)
│       │   └── ...
│       └── idea/                   # 构思文档
│           └── ...
│
├── internal/                       # ⚫ 私有代码 (Go 约定)
│   └── (可选的私有实现)
│
├── pkg/                            # 可复用的独立包 (可选)
│   └── (如需要独立发布的子包)
│
├── tests/                          # 集成测试 (可选)
│   ├── integration/
│   └── e2e/
│
├── go.mod                          # Go 模块定义
├── go.sum                          # 依赖锁定
├── README.md                       # 项目说明
├── LICENSE                         # 许可证
└── .github/                        # GitHub 配置
    └── workflows/                  # CI/CD
```

---

## 二、当前实际目录结构（v0.1 已实现）

> **更新日期**: 2026-02-01
> **状态**: ✅ 运行中

### 2.1 实际目录结构

```
mint/
├── ui/                     # 🔵 声明式 UI 核心（所有代码在此包）
│   ├── 核心系统 (2169+443+435+256+134+384+545 = 4366 行)
│   │   ├── app.go              # 2169 行 - Run 入口、声明式根组件
│   │   ├── fiber.go            # 443 行 - Fiber 节点结构
│   │   ├── reconciler.go       # 435 行 - 协调器核心
│   │   ├── begin_work.go       # 256 行 - BeginWork 阶段
│   │   ├── complete_work.go    # 134 行 - CompleteWork 阶段
│   │   ├── diff.go             # 384 行 - Diff 算法
│   │   └── scheduler.go        # 545 行 - 调度器
│   │
│   ├── 状态管理
│   │   ├── instance.go
│   │   ├── instance_manager.go
│   │   └── interaction_state.go
│   │
│   ├── VNode 系统
│   │   ├── vnode.go
│   │   ├── component.go
│   │   ├── element.go
│   │   ├── fragment.go
│   │   └── text.go
│   │
│   ├── Hooks API
│   │   └── hooks.go
│   │
│   ├── 布局组件
│   │   ├── layout.go
│   │   ├── absolute.go
│   │   └── grid.go
│   │
│   ├── 输入组件
│   │   ├── button.go
│   │   ├── input.go
│   │   ├── checkbox.go
│   │   ├── select.go
│   │   └── textarea.go
│   │
│   ├── 其他组件
│   │   ├── progress.go
│   │   ├── modal.go
│   │   ├── tooltip.go
│   │   └── virtuallist.go
│   │
│   ├── 工具
│   │   ├── memory_safety.go
│   │   └── validator.go
│   │
│   └── 测试文件 (13 个 *_test.go)
│
├── framework/              # 🟢 框架层
│   ├── app.go
│   ├── component/
│   ├── event/
│   ├── binding/
│   └── ...
│
├── runtime/                # 🟣 运行时核心（稳定）
│   ├── paint/
│   ├── event/
│   ├── layout/
│   ├── focus/
│   ├── style/
│   ├── input/
│   └── scheduler/
│
├── devtools/               # 🔵 DevTools
│   ├── core/
│   ├── protocol/
│   └── observation/
│
└── docs/                   # 📚 文档
    └── plan/
        └── directory-refactor-plan.md
```

### 2.2 当前依赖关系

```
ui/ (包含所有实现)
  ↓ 依赖
framework/
  ↓ 依赖
runtime/ (基础层)

devtools/
  ↓ 依赖
ui/ 和 runtime/
```

---

## 三、计划中的重构方案（未实施）

> 以下内容为规划，详见 `docs/plan/directory-refactor-plan.md`

### 3.1 目标结构摘要

```
mint/
├── internal/               # ⚠️ 未创建
│   ├── reconciler/         # 从 ui/ 迁移
│   ├── scheduler/          # 从 ui/ 迁移
│   └── state/              # 从 ui/ 迁移
│
└── ui/                     # 精简为公开 API
    └── app.go (~200 行)
```

---

## 四、模块依赖关系（规划）

### 4.1 依赖图

```
ui/
  ↓ 依赖
framework/reconciler/
  ↓ 依赖
  ├── framework/hooks/
  ├── framework/component/vnode/
  ├── runtime/layout/
  ├── runtime/event/
  └── runtime/paint/

framework/components/
  ↓ 依赖
  ├── ui/
  ├── framework/hooks/
  └── framework/component/

runtime/
  ↓ 独立基础层
  (被所有上层依赖)

devtools/
  ↓ 依赖
  ├── framework/reconciler/
  └── runtime/
```

### 4.2 导入规则

```go
// ✅ 允许：ui → framework
import "mint/framework/reconciler"

// ✅ 允许：framework → runtime
import "mint/runtime/layout"

// ❌ 禁止：runtime → framework
// runtime 是基础层，不能依赖上层

// ❌ 禁止：外部 → internal
// internal 是私有包
```

---

## 四、迁移路径

### 阶段 1: 基础结构搭建 (Week 1)

```
mint/
├── ui/                    # 新建
│   ├── app.go
│   ├── vnode.go
│   └── builder.go
├── framework/
│   ├── reconciler/        # 新建
│   │   ├── diff.go
│   │   └── fiber.go
│   └── hooks/             # 新建
│       └── state.go
```

### 阶段 2: 组件系统迁移 (Week 2-3)

```
framework/
├── component/
│   ├── vnode/             # 新建：VNode 实现
│   ├── base.go            # 保留：兼容层
│   └── adapter.go         # 新建：新旧桥接
└── components/            # 新建：声明式组件
    ├── text.go
    └── button.go
```

### 阶段 3: 渲染管线改造 (Week 4)

```
framework/
├── render/                # 新建：DrawCmd 中间层
│   ├── drawcmd.go
│   └── buffer_adapter.go
└── layout/                # 新建：声明式布局 API
    ├── hstack.go
    └── vstack.go
```

### 阶段 4: DevTools 集成 (Week 5+)

```
devtools/
├── bridge/                # 新建：Fiber 树桥接
│   ├── fiber.go
│   └── profiler.go
└── ui/
    └── dashboard/         # 新建：Web UI
```

---

## 五、命名规范

### 5.1 文件命名

```
// 全小写 + 下划线
 reconciler/
 ├── diff.go           # Diff 算法
 ├── fiber.go          # Fiber 节点
 └── work_loop.go      # 工作循环

// 测试文件
 └── diff_test.go      # 对应 diff.go 的测试
```

### 5.2 包命名

```
// 简洁、描述性
package reconciler  // ✅ 好
package r           // ❌ 差

// 避免与标准库冲突
package event       // ✅ 好
package events      // ⚠️ 可能与标准库 events 冲突
```

### 5.3 接口命名

```go
// 动作 + er 后缀
type Measurer interface{}     // ✅
type Measure interface{}      // ❌

// 纯接口用 -er 后缀
type Painter interface{}      // ✅
type Paint interface{}        // ❌ (但 Paint 作为结构体名可以)

// 组合接口
type ComponentNode interface {
    Node
    Measurer
    Painter
}
```

### 5.4 后缀约定

```
_legacy.go     # 旧版实现，标记废弃
_v2.go         # 新版实现，共存期使用
_test.go       # 测试文件
_mock.go       # Mock 对象
_example.go    # 示例代码（包内示例）
```

---

## 六、Go 模块配置

### 6.1 推荐的 go.mod 结构

```go
module github.com/yao/mint

go 1.23

require (
    // 依赖...
)
```

### 6.2 子模块 (可选)

如果需要独立发布某些子包：

```
github.com/yao/mint          // 主模块
github.com/yao/mint-runtime  // 运行时（可选）
github.com/yao/mint-ui       // UI SDK（可选）
```

---

## 七、文档结构

```
docs/
├── README.md                    # 文档首页
├── getting-started.md           # 快速开始
├── installation.md              # 安装指南
├── migration.md                 # 迁移指南
│
├── api/                         # API 文档
│   ├── ui.md                    # ui 包 API
│   ├── components.md            # 组件 API
│   └── hooks.md                 # Hooks API
│
├── tutorials/                   # 教程
│   ├── basics.md                # 基础教程
│   ├── components.md            # 组件教程
│   ├── state-management.md      # 状态管理
│   └── advanced.md              # 高级主题
│
├── architecture/                # 架构文档
│   ├── overview.md              # 架构概览
│   ├── reconciler.md            # Reconciler
│   ├── layout.md                # 布局系统
│   ├── rendering.md             # 渲染管线
│   └── event-system.md          # 事件系统
│
└── reference/                   # 参考手册
    ├── glossary.md              # 术语表
    ├── troubleshooting.md       # 故障排除
    └── performance.md           # 性能优化
```

---

## 八、检查清单

### 新建目录时确认

- [ ] 目录名符合 snake_case 规范
- [ ] 包名简洁且描述性强
- [ ] 依赖方向正确（上层依赖下层）
- [ ] 添加对应的 README.md 说明
- [ ] 添加示例代码（如需要）

### 迁移代码时确认

- [ ] 标记旧代码为 _legacy
- [ ] 添加适配器层保持兼容
- [ ] 更新 import 路径
- [ ] 更新测试文件
- [ ] 更新文档引用

---

## 九、总结

### 关键决策

1. **保留 `runtime/` 作为独立基础层** - 成熟稳定，无需重构
2. **新建 `ui/` 作为对外 SDK** - 清晰的 API 边界
3. **在 `framework/` 下扩展新功能** - reconciler、hooks 等
4. **保持 `devtools/` 独立** - 调试工具可独立演进

### 迁移策略

- **渐进式迁移** - 新旧代码共存
- **适配器模式** - 无缝桥接新旧系统
- **标记废弃** - _legacy 后缀标识
- **文档先行** - 设计文档 → 实现 → 示例

### 下一步

1. 创建 `ui/` 目录和基础文件
2. 创建 `framework/reconciler/` 和 `framework/hooks/`
3. 建立新旧系统的适配器
4. 更新示例代码使用新 API
