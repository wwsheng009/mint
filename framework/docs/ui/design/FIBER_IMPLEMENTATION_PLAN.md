# Fiber 架构实施计划

## 📋 文档概述

**文档类型**: 实施计划
**项目**: Mint UI Fiber Reconciler
**预计工期**: 4 周
**开始日期**: 2026-02-01
**版本**: v1.0

---

## 🎯 总体目标

实现完整的 Fiber 架构，使 Mint UI 具备：
1. **可中断渲染** - 时间切片，不阻塞 UI
2. **优先级调度** - 用户输入优先于数据更新
3. **增量协调** - 只更新变化的部分
4. **Effect 系统** - 正确的副作用处理

---

## 📊 实施阶段

### Phase 1: 基础 Reconciler (Week 1)

**目标**: 创建可工作的基本协调器

#### 任务清单

```markdown
- [ ] 1.1 创建 `ui/reconciler.go`
  - [ ] Reconciler 结构体
  - [ ] ScheduleUpdate 方法
  - [ ] 基础 WorkLoop（同步版本）
  - [ ] CommitRoot 方法

- [ ] 1.2 创建 `ui/begin_work.go`
  - [ ] BeginWork 入口
  - [ ] beginWorkComponent
  - [ ] beginWorkElement
  - [ ] beginWorkText

- [ ] 1.3 创建 `ui/complete_work.go`
  - [ ] CompleteWork 入口
  - [ ] completeWorkComponent
  - [ ] completeWorkElement

- [ ] 1.4 集成到 `ui/app.go`
  - [ ] 添加 reconciler 字段
  - [ ] 添加 paintWithFiber 方法
  - [ ] 添加环境变量控制

- [ ] 1.5 基础测试
  - [ ] 单元测试
  - [ ] 集成测试
```

#### 验收标准

```go
// 以下代码应该能运行
func main() {
    ui.Run(func() ui.VNode {
        count, setCount := ui.UseStateInt(0)
        return ui.VStack(
            ui.Text(fmt.Sprintf("Count: %d", count)),
            ui.Button("+").OnClick(func() {
                setCount(count + 1)
            }),
        )
    }, ui.WithEnv("MINT_USE_FIBER=true"))
}
```

---

### Phase 2: Commit 阶段 (Week 2)

**目标**: 实现变更提交到 buffer

#### 任务清单

```markdown
- [ ] 2.1 创建 `ui/commit.go`
  - [ ] CommitRoot 主方法
  - [ ] commitBeforeMutationEffects
  - [ ] commitMutationEffects
  - [ ] commitLayoutEffects

- [ ] 2.2 创建 `ui/effects.go`
  - [ ] Effect 结构体
  - [ ] collectEffects
  - [ ] runEffects
  - [ ] flushPassiveEffects

- [ ] 2.3 Buffer 渲染
  - [ ] renderFiberToBuffer 实现
  - [ ] 处理不同 VNode 类型

- [ ] 2.4 测试
  - [ ] Effect 测试
  - [ ] Commit 测试
```

#### 验收标准

- [ ] 状态更新正确反映到 buffer
- [ ] useEffect 正确执行
- [ ] Cleanup 函数正确调用

---

### Phase 3: 时间切片 (Week 3)

**目标**: 实现可中断渲染

#### 任务清单

```markdown
- [ ] 3.1 增强 reconciler.go
  - [ ] WorkLoopWithTimeSlice
  - [ ] deadline 检查
  - [ ] 请求下一帧继续

- [ ] 3.2 优先级调度
  - [ ] Lane 系统集成
  - [ ] 优先级队列
  - [ ] 同步更新优先处理

- [ ] 3.3 性能测试
  - [ ] 大型组件树测试
  - [ ] 时间切片验证
  - [ ] 内存泄漏检测
```

#### 验收标准

- [ ] 1000+ 节点组件树不阻塞
- [ ] 用户输入响应 < 16ms
- [ ] 时间切片正确中断

---

### Phase 4: Key 协调 (Week 4)

**目标**: 实现基于 key 的子节点协调

#### 任务清单

```markdown
- [ ] 4.1 创建 `ui/reconcile.go`
  - [ ] reconcileChildren 主方法
  - [ ] reconcileChildrenArray
  - [ ] mapRemainingChildren
  - [ ] updateFromMap

- [ ] 4.2 Key 匹配算法
  - [ ] Key 到 Fiber 映射
  - [ ] O(1) 查找
  - [ ] 移动/复用 Fiber

- [ ] 4.3 测试
  - [ ] Key 测试
  - [ ] 列表增删测试
  - [ ] 性能测试
```

#### 验收标准

- [ ] Key 匹配正确工作
- [ ] 列表重排不重建组件
- [ ] 性能满足预期

---

## 📂 文件创建顺序

```
Week 1:
├── ui/reconciler.go          # Day 1-2
├── ui/begin_work.go          # Day 3-4
├── ui/complete_work.go       # Day 4
└── ui/app.go (修改)          # Day 5

Week 2:
├── ui/commit.go              # Day 6-7
└── ui/effects.go             # Day 8-9

Week 3:
└── ui/reconciler.go (增强)   # Day 10-12

Week 4:
├── ui/reconcile.go           # Day 13-15
└── 测试与优化                 # Day 16-20
```

---

## 🔗 依赖关系

```
┌─────────────────────────────────────────────────────────────┐
│                        Phase 1                            │
│  reconciler.go ──► begin_work.go ──► complete_work.go   │
│       │                                                      │
│       └──► app.go (集成)                                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        Phase 2                            │
│  reconciler.go ──► commit.go ──► effects.go               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        Phase 3                            │
│  reconciler.go (时间切片 + 优先级)                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        Phase 4                            │
│  reconcile.go (Key 协调)                                    │
└─────────────────────────────────────────────────────────────┘
```

---

## 🧪 测试策略

### 单元测试

```go
// ui/reconciler_test.go
func TestReconciler_ScheduleUpdate(t *testing.T)
func TestReconciler_WorkLoop_Sync(t *testing.T)
func TestReconciler_CommitRoot(t *testing.T)

// ui/begin_work_test.go
func TestBeginWork_Component(t *testing.T)
func TestBeginWork_Element(t *testing.T)

// ui/commit_test.go
func TestCommit_RenderToBuffer(t *testing.T)
```

### 集成测试

```go
// ui/fiber_integration_test.go
func TestFiber_CounterApp(t *testing.T) {
    // 完整的计数器应用测试
}

func TestFiber_EffectTiming(t *testing.T) {
    // Effect 时序测试
}

func TestFiber_StateUpdate(t *testing.T) {
    // 状态更新测试
}
```

### 性能测试

```go
// ui/fiber_bench_test.go
func BenchmarkFiber_Reconcile(b *testing.B)
func BenchmarkFiber_LargeTree(b *testing.B)
func BenchmarkFiber_TimeSlicing(b *testing.B)
```

---

## 🚀 部署策略

### 阶段 1: 环境变量控制

```bash
# 默认使用传统渲染
go run examples/counter/main.go

# 启用 Fiber（开发/测试）
MINT_USE_FIBER=true go run examples/counter/main.go
```

### 阶段 2: API 控制

```go
// 明确选择使用 Fiber
ui.Run(app, ui.UseFiber(true))

// 或者
ui.EnableFiber()
ui.Run(app)
```

### 阶段 3: 默认启用

```go
// Fiber 成为默认方式
ui.Run(app) // 默认使用 Fiber
```

---

## 📈 进度跟踪

| 日期 | 任务 | 状态 | 备注 |
|------|------|------|------|
| Week 1 | Phase 1 实施 | ⏳ 待开始 | 基础 Reconciler |
| Week 2 | Phase 2 实施 | ⏳ 待开始 | Commit 阶段 |
| Week 3 | Phase 3 实施 | ⏳ 待开始 | 时间切片 |
| Week 4 | Phase 4 实施 | ⏳ 待开始 | Key 协调 |

---

## ✅ 最终验收

### 功能完整性

- [ ] 所有 Phase 任务完成
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试全部通过
- [ ] 性能测试达标

### 性能指标

| 指标 | 目标值 | 当前值 |
|------|--------|--------|
| 渲染帧率 | ≥ 60 FPS | - |
| 组件树规模 | 1000+ 节点 | - |
| 输入响应 | < 16ms | - |
| 时间切片 | 5ms 预算 | - |

### 稳定性

- [ ] 无内存泄漏
- [ ] 无 panic/crash
- [ ] 现有应用兼容

---

## 📚 相关文档

- [Fiber 架构设计](./FIBER_ARCHITECTURE.md)
- [Reconciler 实施方案](./RECONCILER_IMPLEMENTATION.md)
- [系统架构设计](./SYSTEM_ARCHITECTURE.md)
- [实施计划](./IMPLEMENTATION_PLAN.md)

---

**文档版本**: v1.0
**最后更新**: 2026-02-01
