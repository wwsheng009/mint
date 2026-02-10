# 事件系统重构任务列表

> **版本**: v1.0
> **日期**: 2025-02-10
> **预计工期**: 7 周

---

## 📊 任务概览

| 阶段 | 任务数 | 预计工期 | 状态 |
|------|--------|---------|------|
| Phase 0 | 5 | 1-2 天 | 🟡 进行中 |
| Phase 1 | 8 | 2 周 | ⏳ 待开始 |
| Phase 2 | 10 | 2 周 | ⏳ 待开始 |
| Phase 3 | 6 | 1 周 | ⏳ 待开始 |
| Phase 4 | 8 | 1 周 | ⏳ 待开始 |
| Phase 5 | 6 | 持续 | ⏳ 待开始 |
| Phase 6 | 4 | 1 周 | ⏳ 待开始 |

**总计**: 47 个任务

---

## Phase 0: 前置修复（立即）

### P0-1: ElementVNode 添加 bounds 字段
**文件**: `runtime/ui/element.go`
**优先级**: P0
**估时**: 30 分钟
**状态**: ✅ 已完成

```go
type ElementVNode struct {
    // ... 现有字段
    bounds [4]int
}
```

### P0-2: ElementVNode 实现 SetBounds/GetBounds
**文件**: `runtime/ui/element.go`
**优先级**: P0
**估时**: 15 分钟
**状态**: ✅ 已完成

### P0-3: vnodeContains 支持两种签名
**文件**: `internal/inspector/inspector.go`
**优先级**: P0
**估时**: 30 分钟
**状态**: ✅ 已完成

### P0-4: 修复 handleOverlayMouse 边界计算
**文件**: `internal/inspector/standalone_inspector.go`
**优先级**: P0
**估时**: 1 小时
**状态**: ✅ 已完成

### P0-5: 验证 Inspector 基本功能
**优先级**: P0
**估时**: 1 小时
**状态**: ⏳ 待测试

**验收**:
- [ ] Inspector tab 点击正常切换
- [ ] Hover 显示正确组件路径
- [ ] 鼠标坐标显示正确

---

## Phase 1: HitMap 系统（Week 1-2）

### P1-1: 定义 HitMap 核心结构
**文件**: `runtime/event/hitmap.go` (新建)
**优先级**: P0
**估时**: 2 小时
**依赖**: 无
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `Rect` 结构
- [ ] 定义 `HitMapEntry` 结构
- [ ] 定义 `HitMap` 结构
- [ ] 实现 `BuildHitMap(root)`

### P1-2: 实现 HitMap 递归构建
**文件**: `runtime/event/hitmap.go`
**优先级**: P0
**估时**: 2 小时
**依赖**: P1-1
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `walkAndBuild()` 递归函数
- [ ] 实现 `sortByZOrder()` 排序
- [ ] 处理边界情况（nil 节点）

### P1-3: 实现 HitTest 方法
**文件**: `runtime/event/hitmap.go`
**优先级**: P0
**估时**: 1 小时
**依赖**: P1-2
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `HitTest(x, y)` 方法
- [ ] 正确处理 Z-index（上层优先）
- [ ] 返回 `LocalXY` 计算函数

### P1-4: 扩展 MouseEvent 结构
**文件**: `framework/event/mouse_event.go`
**优先级**: P0
**估时**: 1 小时
**依赖**: 无
**状态**: ⏳ 待开始

**任务**:
- [ ] 添加 `TargetID` 字段
- [ ] 添加 `LocalX, LocalY` 字段
- [ ] 添加 `Delta` 字段（滚轮）
- [ ] 添加 `MouseAction` 枚举

### P1-5: App 集成 HitMap 构建
**文件**: `framework/app.go`
**优先级**: P0
**估时**: 2 小时
**依赖**: P1-2
**状态**: ⏳ 待开始

**任务**:
- [ ] 添加 `hitMap` 字段
- [ ] 在 `render()` 后构建 HitMap
- [ ] 实现 `GetHitMap()` 方法

### P1-6: Pump 填充命中信息
**文件**: `runtime/pump.go`
**优先级**: P0
**估时**: 2 小时
**依赖**: P1-4, P1-5
**状态**: ⏳ 待开始

**任务**:
- [ ] 修改 `handleMouseInput()` 方法
- [ ] 从 App 获取 HitMap
- [ ] 调用 `HitTest()` 填充 TargetID/LocalX/LocalY

### P1-7: 单元测试 - HitMap
**文件**: `runtime/event/hitmap_test.go` (新建)
**优先级**: P1
**估时**: 2 小时
**依赖**: P1-3
**状态**: ⏳ 待开始

**任务**:
- [ ] `TestHitMap_Build()` - 测试构建
- [ ] `TestHitMap_HitTest()` - 测试命中
- [ ] `TestHitMap_FindByID()` - 测试 ID 查找
- [ ] `TestHitMap_ZOrder()` - 测试 Z-index 排序
- [ ] 覆盖率 > 90%

### P1-8: 集成测试 - Inspector 鼠标事件
**文件**: `internal/inspector/inspector_mouse_test.go`
**优先级**: P0
**估时**: 2 小时
**依赖**: P1-6
**状态**: ⏳ 待开始

**任务**:
- [ ] 测试 tab 点击切换
- [ ] 测试 hover 检测
- [ ] 测试坐标显示正确性
- [ ] 验证 Inspector 无需手动计算坐标

**验收标准（Phase 1）**:
- [ ] HitMap 正确构建所有节点
- [ ] `HitTest(x,y)` 返回正确目标
- [ ] MouseEvent 包含 TargetID/LocalX/LocalY
- [ ] Inspector tab 点击工作正常
- [ ] Hover 检测显示正确组件
- [ ] 单元测试覆盖率 > 85%

---

## Phase 2: Action 系统（Week 3-4）

### P2-1: 定义 Action 类型
**文件**: `framework/action/action.go` (新建)
**优先级**: P0
**估时**: 1 小时
**依赖**: 无
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `Action` 结构
- [ ] 定义 `ActionType` 常量
- [ ] 实现 `IsNavigation()` 方法
- [ ] 实现 `IsEditing()` 方法

### P2-2: 实现 InputProcessor
**文件**: `framework/action/processor.go` (新建)
**优先级**: P0
**估时**: 2 小时
**依赖**: P2-1
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `Process(Event)` 方法
- [ ] 实现 `processKeyEvent()` 方法
- [ ] 实现 `processMouseEvent()` 方法
- [ ] 定义默认转换规则

### P2-3: 实现 KeyMap
**文件**: `framework/action/keymap.go` (新建)
**优先级**: P0
**估时**: 2 小时
**依赖**: P2-1
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `KeySignature` 结构
- [ ] 实现 `Lookup()` 方法
- [ ] 实现 `PushContext()` / `PopContext()`
- [ ] 实现 `Bind()` 方法

### P2-4: 定义 ActionTarget 接口
**文件**: `framework/component/action_target.go` (新建)
**优先级**: P0
**估时**: 30 分钟
**依赖**: P2-1
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `ActionTarget` 接口
- [ ] 提供示例实现

### P2-5: TreeView 实现 ActionTarget
**文件**: `components/display/treeview.go`
**优先级**: P1
**估时**: 2 小时
**依赖**: P2-4
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `HandleAction()` 方法
- [ ] 支持 Navigate* Action
- [ ] 支持 Select/Toggle/Expand/Collapse Action
- [ ] 支持 Scroll Action

### P2-6: Tabs 实现 ActionTarget
**文件**: `components/navigation/tabs.go`
**优先级**: P1
**估时**: 1 小时
**依赖**: P2-4
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `HandleAction()` 方法
- [ ] 支持 Navigate* Action（切换 tab）
- [ ] 支持 Select Action

### P2-7: Button 实现 ActionTarget
**文件**: `components/button/button.go`
**优先级**: P1
**估时**: 1 小时
**依赖**: P2-4
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `HandleAction()` 方法
- [ ] 支持 Select Action
- [ ] 支持 Hover Action

### P2-8: 输入组件实现 ActionTarget
**文件**: `components/form/input.go`, `textarea.go`, `select.go`
**优先级**: P1
**估时**: 3 小时
**依赖**: P2-4
**状态**: ⏳ 待开始

**任务**:
- [ ] Input 支持 InputText/DeleteChar Action
- [ ] Textarea 支持 InputText/DeleteLine Action
- [ ] Select 支持 Navigate* / Select Action

### P2-9: 单元测试 - Action 转换
**文件**: `framework/action/processor_test.go` (新建)
**优先级**: P1
**估时**: 2 小时
**依赖**: P2-2
**状态**: ⏳ 待开始

**任务**:
- [ ] `TestKeyToAction()` - 测试键盘转换
- [ ] `TestMouseToAction()` - 测试鼠标转换
- [ ] `TestKeyMapContext()` - 测试上下文切换
- [ ] 覆盖率 > 85%

### P2-10: 集成测试 - Action 处理
**文件**: `framework/action/integration_test.go` (新建)
**优先级**: P1
**估时**: 3 小时
**依赖**: P2-5, P2-6, P2-7
**状态**: ⏳ 待开始

**任务**:
- [ ] 测试 TreeView Action 处理
- [ ] 测试 Tabs Action 处理
- [ ] 测试 Button Action 处理
- [ ] 验证 Action → State Update 流程

**验收标准（Phase 2）**:
- [ ] Action 类型定义完整
- [ ] KeyEvent 正确转换为 Navigate/Input/Quit
- [ ] MouseEvent 正确转换为 Click/Scroll/Hover
- [ ] KeyMap 支持上下文切换
- [ ] 至少 3 个组件实现 HandleAction
- [ ] 单元测试覆盖率 > 85%

---

## Phase 3: Router 三阶段分发（Week 5）

### P3-1: 实现 Router 核心结构
**文件**: `runtime/event/router.go` (新建/修改)
**优先级**: P0
**估时**: 2 小时
**依赖**: P2-2
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `Router` 结构
- [ ] 实现 `NewRouter()` 构造函数
- [ ] 实现 `AddCaptureHandler()` 方法

### P3-2: 实现 Capture Phase
**文件**: `runtime/event/router.go`
**优先级**: P0
**估时**: 2 小时
**依赖**: P3-1
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `capturePhase()` 方法
- [ ] 按优先级排序处理器
- [ ] 支持 `StopPropagation()`

### P3-3: 实现 Target Phase
**文件**: `runtime/event/router.go`
**优先级**: P0
**估时**: 2 小时
**依赖**: P3-1
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `targetPhase()` 方法
- [ ] 根据 TargetID 查找节点
- [ ] 调用 `HandleAction()` 接口

### P3-4: 实现 Bubble Phase
**文件**: `runtime/event/router.go`
**优先级**: P0
**估时**: 2 小时
**依赖**: P3-1
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `bubblePhase()` 方法
- [ ] 获取父链
- [ ] 向上冒泡 Action

### P3-5: Inspector 实现捕获监听器
**文件**: `internal/inspector/standalone_inspector.go`
**优先级**: P0
**估时**: 2 小时
**依赖**: P3-2
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `CaptureHandler` 接口
- [ ] 实现 `HandleCapture()` 方法
- [ ] 实现 `Priority()` 方法
- [ ] 注册到 Router

### P3-6: 单元测试 - 三阶段分发
**文件**: `runtime/event/router_test.go` (新建)
**优先级**: P0
**估时**: 3 小时
**依赖**: P3-2, P3-3, P3-4
**状态**: ⏳ 待开始

**任务**:
- [ ] `TestRouter_CapturePhase()` - 测试捕获阶段
- [ ] `TestRouter_TargetPhase()` - 测试目标阶段
- [ ] `TestRouter_BubblePhase()` - 测试冒泡阶段
- [ ] `TestRouter_StopPropagation()` - 测试停止传播
- [ ] `TestRouter_InspectorPriority()` - 测试 Inspector 优先级
- [ ] 覆盖率 > 90%

**验收标准（Phase 3）**:
- [ ] Router 实现三阶段分发
- [ ] Inspector 正确捕获 overlay 事件
- [ ] 全局快捷键在 Capture Phase 处理
- [ ] 事件可停止传播
- [ ] 单元测试覆盖率 > 85%

---

## Phase 4: Msg/Cmd 系统（Week 6）

### P4-1: 定义 Msg 核心接口
**文件**: `framework/msg/msg.go` (新建)
**优先级**: P0
**估时**: 1 小时
**依赖**: 无
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `Msg` 接口
- [ ] 定义 `MsgType` 常量
- [ ] 提供调试方法

### P4-2: 实现 KeyMsg
**文件**: `framework/msg/key_msg.go` (新建)
**优先级**: P0
**估时**: 1 小时
**依赖**: P4-1
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `KeyMsg` 结构
- [ ] 实现 `IsPrintable()` 方法
- [ ] 实现 `HasModifier()` 方法

### P4-3: 实现 MouseMsg
**文件**: `framework/msg/mouse_msg.go` (新建)
**优先级**: P0
**估时**: 1 小时
**依赖**: P4-1
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `MouseMsg` 结构
- [ ] 添加 `TargetID`/`LocalX`/`LocalY` 字段
- [ ] 实现 `IsClick()` / `IsScroll()` 方法

### P4-4: 实现 SandboxMsg
**文件**: `framework/msg/sandbox_msg.go` (新建)
**优先级**: P0
**估时**: 2 小时
**依赖**: P4-1
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `SandboxMsg` 结构
- [ ] 定义 `SandboxInjectType` 常量
- [ ] 支持键盘/鼠标/Action/状态注入
- [ ] 实现分类方法（`IsInput`/`IsDirectAction`/`IsStateMutation`）

### P4-5: 实现 Event → Msg 适配器
**文件**: `framework/msg/adapter.go` (新建)
**优先级**: P0
**估时**: 2 小时
**依赖**: P4-2, P4-3
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `ToMsg(Event)` 函数
- [ ] 转换 KeyEvent → KeyMsg
- [ ] 转换 MouseEvent → MouseMsg
- [ ] 转换 ResizeEvent → ResizeMsg

### P4-6: 定义 Cmd 接口
**文件**: `framework/cmd/cmd.go` (新建)
**优先级**: P0
**估时**: 1 小时
**依赖**: 无
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `Cmd` 接口
- [ ] 定义标准 Cmd: `Batch`, `After`, `Tick`, `IO`

### P4-7: 实现 Updater 接口
**文件**: `framework/component/updater.go` (新建)
**优先级**: P1
**估时**: 1 小时
**依赖**: P4-1, P4-6
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `Updater` 接口
- [ ] App 实现 `processMsg()` 方法
- [ ] 支持回退到旧 Event 系统

### P4-8: 单元测试 - Msg/Cmd
**文件**: `framework/msg/msg_test.go` (新建)
**优先级**: P1
**估时**: 2 小时
**依赖**: P4-5, P4-6
**状态**: ⏳ 待开始

**任务**:
- [ ] `TestMsgConversion()` - 测试 Event → Msg
- [ ] `TestCmdBatch()` - 测试 Cmd 组合
- [ ] `TestCmdAfter()` - 测试延迟 Cmd
- [ ] `TestCmdTick()` - 测试定时器 Cmd
- [ ] 覆盖率 > 85%

**验收标准（Phase 4）**:
- [ ] Msg 类型定义完整
- [ ] Event → Msg 转换正确
- [ ] Cmd 可组合、可执行
- [ ] 组件可选择性实现 Update
- [ ] 单元测试覆盖率 > 85%

---

## Phase 5: 测试与工具（持续）

### P5-1: 实现 TestableApp
**文件**: `framework/testing/testable_app.go` (新建)
**优先级**: P0
**估时**: 2 小时
**依赖**: P1-7
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `TestableApp` 结构
- [ ] 实现 `InjectMouseByID()` 方法
- [ ] 实现 `InjectKeys()` 方法
- [ ] 实现断言助手（`AssertHovered`, `AssertFocused`）

### P5-2: 实现 Sandbox Injector
**文件**: `framework/sandbox/injector.go` (新建)
**优先级**: P0
**估时**: 3 小时
**依赖**: P4-4
**状态**: ⏳ 待开始

**任务**:
- [ ] 定义 `Injector` 结构
- [ ] 实现 `InjectKey()` / `InjectKeySequence()`
- [ ] 实现 `InjectMouseClick()`
- [ ] 实现 `InjectAction()`
- [ ] 实现 `InjectSetState()`

### P5-3: HitMap 可视化工具
**文件**: `runtime/event/hitmap_debug.go` (新建)
**优先级**: P2
**估时**: 2 小时
**依赖**: P1-3
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `Dump()` 方法
- [ ] 实现 `Visualize()` 方法（TUI 叠加显示）
- [ ] 支持 `TUI_DEBUG_HITMAP=1` 环境变量

### P5-4: 事件流可视化工具
**文件**: `framework/event/debug.go` (新建)
**优先级**: P2
**估时**: 2 小时
**依赖**: P3-6
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现事件日志记录
- [ ] 实现事件轨迹可视化
- [ ] 支持 `TUI_DEBUG_EVENT=1` 环境变量

### P5-5: 集成测试 - 按 ID 注入
**文件**: `framework/testing/integration_test.go` (新建)
**优先级**: P0
**估时**: 3 小时
**依赖**: P5-1, P5-2
**状态**: ⏳ 待开始

**任务**:
- [ ] 测试 `InjectMouseByID()`
- [ ] 测试 `InjectKeys()`
- [ ] 测试 `InjectAction()`
- [ ] 验证测试可读性提升

### P5-6: 文档和示例
**文件**: `docs/`, `examples/`
**优先级**: P1
**估时**: 4 小时
**依赖**: P5-1, P5-2
**状态**: ⏳ 待开始

**任务**:
- [ ] 更新 API 文档
- [ ] 编写迁移指南
- [ ] 提供测试示例
- [ ] 更新 README

**验收标准（Phase 5）**:
- [ ] 测试可按 ID 注入事件
- [ ] HitMap 可视化可用
- [ ] 事件流可调试
- [ ] 文档完整

---

## Phase 6: 性能优化（Week 7）

### P6-1: 鼠标事件节流
**文件**: `runtime/scheduler/mouse_throttle.go` (新建)
**优先级**: P0
**估时**: 2 小时
**依赖**: P4-3
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `MouseThrottler` 结构
- [ ] 实现 `Throttle()` 方法
- [ ] 支持时间间隔和像素阈值配置
- [ ] 节流到 60Hz

### P6-2: HitMap 增量更新
**文件**: `runtime/event/hitmap_incremental.go` (新建)
**优先级**: P1
**估时**: 3 小时
**依赖**: P1-2
**状态**: ⏳ 待开始

**任务**:
- [ ] 实现 `IncrementalUpdate()` 方法
- [ ] 只重新构建变化的子树
- [ ] 性能测试对比

### P6-3: 性能基准测试
**文件**: `runtime/event/hitmap_bench_test.go` (新建)
**优先级**: P1
**估时**: 2 小时
**依赖**: P6-2
**状态**: ⏳ 待开始

**任务**:
- [ ] 测试 HitMap 构建性能（1000 节点）
- [ ] 测试 HitTest 性能
- [ ] 对比全量 vs 增量更新
- [ ] 确保目标：<10ms (1000 节点)

### P6-4: 内存优化
**文件**: `runtime/event/hitmap.go`
**优先级**: P2
**估时**: 2 小时
**依赖**: P1-2
**状态**: ⏳ 待开始

**任务**:
- [ ] 优化 HitMapEntry 内存布局
- [ ] 减少不必要的拷贝
- [ ] 内存泄漏检查

**验收标准（Phase 6）**:
- [ ] MouseMove 节流到 60Hz
- [ ] HitMap 构建 <10ms (1000 节点)
- [ ] 增量更新可用
- [ ] 事件延迟 <16ms
- [ ] 无内存泄漏

---

## 📅 总体时间表

```
Week 1-2: Phase 1 (HitMap) + Phase 0 验收
Week 3-4: Phase 2 (Action 系统)
Week 5:   Phase 3 (Router 三阶段)
Week 6:   Phase 4 (Msg/Cmd)
Week 7:   Phase 5 (测试工具) + Phase 6 (性能优化)
```

---

## 🎯 里程碑

| 里程碑 | 验收标准 | 目标日期 |
|--------|---------|---------|
| M1: HitMap 可用 | Inspector 鼠标事件正常 | Week 2 |
| M2: Action 转换 | 3+ 组件实现 HandleAction | Week 4 |
| M3: 三阶段分发 | Inspector 作为捕获监听器 | Week 5 |
| M4: Msg 统一 | Event → Msg → Action 流通 | Week 6 |
| M5: 测试工具 | 按 ID 注入可用 | Week 7 |
| M6: 性能达标 | 所有性能指标达标 | Week 7 |

---

## ✅ 总验收标准

### 功能完整性
- [ ] Inspector tab 点击工作
- [ ] Hover 检测显示正确组件
- [ ] 所有事件流经三阶段
- [ ] Action 语义化完整
- [ ] Msg/Cmd 可组合
- [ ] Sandbox 注入所有类型可用

### 性能指标
- [ ] HitMap 构建 <10ms (1000 节点)
- [ ] MouseMove 节流到 60Hz
- [ ] 事件延迟 <16ms
- [ ] 无内存泄漏

### 测试覆盖率
- [ ] 单元测试 >80%
- [ ] 集成测试覆盖主要流程
- [ ] 测试可按 ID 注入事件

### 文档完整性
- [ ] API 文档更新
- [ ] 迁移指南完整
- [ ] 示例代码同步
- [ ] 设计文档完整

---

## 🚀 立即开始

**下一步**: 从 Phase 0 的 P0-5 开始，验证 Inspector 基本功能。

```bash
# 运行 demo
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
go run main.go

# 测试功能
# 1. 按 F12 打开 Inspector
# 2. 点击 tab 栏（应该切换）
# 3. 移动鼠标（应该显示 hover 信息）
```

**当前任务**: 等待用户确认开始实施
