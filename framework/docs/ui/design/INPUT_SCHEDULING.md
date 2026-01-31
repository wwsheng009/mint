# 输入优先级调度设计文档

**版本**: v1.0
**日期**: 2026-01-31
**来源**: idea/idea8_Concurrent.md
**状态**: 🟡 中优先级

---

## 一、概述

### 1.1 设计目标

实现**输入永远优先于渲染**的调度策略，确保 UI 对用户操作的即时响应。

### 1.2 核心原则

```
输入响应 > 交互事件 > UI 更新 > 后台任务
```

**问题场景**: 当正在进行复杂的渲染计算时，用户按下了按键。

```go
// ❌ 错误：渲染完成才处理输入
startRendering()  // 可能需要 50ms
// ... 渲染中 ...
handleKeyPress()  // 用户按键被延迟

// ✅ 正确：立即中断渲染，处理输入
startRendering()
if hasPendingInput() {
    interruptRendering()
    handleKeyPress()  // 立即响应
    resumeRendering()
}
```

---

## 二、优先级定义

### 2.1 优先级等级

```go
// scheduler/priority.go

package scheduler

// Priority 优先级
type Priority int

const (
    // PriorityImmediate 立即优先级（输入事件）
    // 用于: 按键、焦点变化、窗口调整
    PriorityImmediate Priority = 3

    // PriorityUserBlock 用户阻塞优先级（交互事件）
    // 用于: 点击、表单提交、模态框操作
    PriorityUserBlock Priority = 2

    // PriorityNormal 普通优先级（UI 更新）
    // 用于: 状态更新、组件重渲染
    PriorityNormal Priority = 1

    // PriorityLow 低优先级（后台任务）
    // 用于: 日志输出、统计收集、清理
    PriorityLow Priority = 0
)

// String 返回优先级名称
func (p Priority) String() string {
    switch p {
    case PriorityImmediate:
        return "immediate"
    case PriorityUserBlock:
        return "user-block"
    case PriorityNormal:
        return "normal"
    case PriorityLow:
        return "low"
    default:
        return "unknown"
    }
}
```

### 2.2 任务类型与优先级映射

| 事件类型 | 优先级 | 说明 |
|----------|--------|------|
| KeyPress / KeyRelease | Immediate | 按键输入 |
| MouseClick / DoubleClick | Immediate | 鼠标点击 |
| MouseMove | Normal | 鼠标移动（高频） |
| WindowResize | Immediate | 窗口调整 |
| FocusChange | Immediate | 焦点变化 |
| ButtonClick | UserBlock | 按钮点击 |
| FormSubmit | UserBlock | 表单提交 |
| ModalOpen / Close | UserBlock | 模态框操作 |
| StateUpdate | Normal | 状态更新 |
| Animation | Normal | 动画帧 |
| Log | Low | 日志输出 |
| Metrics | Low | 统计收集 |

---

## 三、输入事件处理

### 3.1 输入事件队列

```go
// scheduler/input_queue.go

package scheduler

import (
    "container/list"
    "sync"
)

// InputEvent 输入事件
type InputEvent struct {
    Type     string
    Data     interface{}
    Priority Priority
    Handler  func() error
}

// InputQueue 输入事件队列
type InputQueue struct {
    mu     sync.Mutex
    queue  *list.List
    notify chan struct{}
}

// NewInputQueue 创建输入队列
func NewInputQueue() *InputQueue {
    return &InputQueue{
        queue:  list.New(),
        notify: make(chan struct{}, 1),
    }
}

// Push 推入输入事件
func (q *InputQueue) Push(event *InputEvent) {
    q.mu.Lock()
    defer q.mu.Unlock()

    // 按优先级插入（高优先级在前）
    for e := q.queue.Front(); e != nil; e = e.Next() {
        existing := e.Value.(*InputEvent)
        if existing.Priority < event.Priority {
            q.queue.InsertBefore(e, event)
            goto notify
        }
    }

    // 最低优先级，放最后
    q.queue.PushBack(event)

notify:
    // 通知有新事件
    select {
    case q.notify <- struct{}{}:
    default:
    }
}

// Pop 弹出输入事件
func (q *InputQueue) Pop() *InputEvent {
    q.mu.Lock()
    defer q.mu.Unlock()

    if q.queue.Len() == 0 {
        return nil
    }

    front := q.queue.Front()
    q.queue.Remove(front)

    return front.Value.(*InputEvent)
}

// HasPending 检查是否有待处理事件
func (q *InputQueue) HasPending() bool {
    q.mu.Lock()
    defer q.mu.Unlock()
    return q.queue.Len() > 0
}

// Len 返回队列长度
func (q *InputQueue) Len() int {
    q.mu.Lock()
    defer q.mu.Unlock()
    return q.queue.Len()
}
```

### 3.2 输入处理器

```go
// scheduler/input_handler.go

package scheduler

import (
    "github.com/wwsheng009/mint/runtime/event"
)

// InputHandler 输入处理器
type InputHandler struct {
    queue    *InputQueue
    platform runtime.Platform
}

// NewInputHandler 创建输入处理器
func NewInputHandler(platform runtime.Platform) *InputHandler {
    return &InputHandler{
        queue:    NewInputQueue(),
        platform: platform,
    }
}

// ProcessInput 处理输入（立即执行）
func (h *InputHandler) ProcessInput() error {
    // 读取原始输入
    raw := h.platform.ReadInput()
    if raw == nil {
        return nil
    }

    // 转换为输入事件
    event := h.convertToEvent(raw)

    // 立即处理
    return event.Handler()
}

// convertToEvent 转换原始输入为事件
func (h *InputHandler) convertToEvent(raw *platform.RawInput) *InputEvent {
    switch raw.Type {
    case platform.InputKey:
        return &InputEvent{
            Type:     "keypress",
            Priority: PriorityImmediate,
            Handler:  h.handleKeyPress(raw),
        }

    case platform.InputMouse:
        return &InputEvent{
            Type:     "mouse",
            Priority: PriorityImmediate,
            Handler:  h.handleMouseEvent(raw),
        }

    case platform.InputResize:
        return &InputEvent{
            Type:     "resize",
            Priority: PriorityImmediate,
            Handler:  h.handleResize(raw),
        }

    default:
        return nil
    }
}
```

---

## 四、可中断渲染

### 4.1 可中断任务

```go
// scheduler/interruptible.go

package scheduler

import (
    "context"
    "time"
)

// InterruptibleTask 可中断任务
type InterruptibleTask struct {
    ID       string
    Func     func(ctx context.Context) error
    Priority Priority
    cancel   context.CancelFunc
}

// NewInterruptibleTask 创建可中断任务
func NewInterruptibleTask(id string, priority Priority, fn func(ctx context.Context) error) *InterruptibleTask {
    return &InterruptibleTask{
        ID:       id,
        Func:     fn,
        Priority: priority,
    }
}

// Execute 执行任务（可被中断）
func (t *InterruptibleTask) Execute(inputQueue *InputQueue) error {
    ctx, cancel := context.WithCancel(context.Background())
    t.cancel = cancel

    done := make(chan error, 1)

    go func() {
        done <- t.Func(ctx)
    }()

    for {
        select {
        case err := <-done:
            return err

        case <-time.After(5 * time.Millisecond):
            // 定期检查是否有输入
            if inputQueue.HasPending() {
                // 有输入事件，中断当前任务
                cancel()
                return ErrInterrupted
            }
        }
    }
}

// Resume 恢复执行被中断的任务
func (t *InterruptibleTask) Resume(inputQueue *InputQueue) error {
    return t.Execute(inputQueue)
}

// Cancel 取消任务
func (t *InterruptibleTask) Cancel() {
    if t.cancel != nil {
        t.cancel()
    }
}

// ErrInterrupted 中断错误
var ErrInterrupted = errors.New("task interrupted by input")
```

### 4.2 渲染任务中断

```go
// scheduler/render_task.go

package scheduler

// RenderTask 渲染任务
type RenderTask struct {
    Fiber     *reconciler.Fiber
    Lane      reconciler.Lane
}

// Execute 执行渲染（可中断）
func (t *RenderTask) Execute(inputQueue *InputQueue) error {
    // 分阶段渲染
    for {
        // 检查是否有输入
        if inputQueue.HasPending() {
            return ErrInterrupted
        }

        // 执行一个工作单元
        hasMore := t.workLoopStep()

        if !hasMore {
            return nil // 渲染完成
        }

        // 让出 CPU
        runtime.Gosched()
    }
}

// workLoopStep 执行一步工作循环
func (t *RenderTask) workLoopStep() bool {
    // 执行一个 Fiber 节点的工作
    // ...

    return false
}
```

---

## 五、调度器实现

### 5.1 核心调度器

```go
// scheduler/scheduler.go

package scheduler

import (
    "sync"
    "time"
)

// Scheduler 调度器
type Scheduler struct {
    mu           sync.RWMutex
    inputQueue   *InputQueue
    taskQueue    []*InterruptibleTask
    currentTask  *InterruptibleTask
    running      bool
    inputHandler *InputHandler
}

// NewScheduler 创建调度器
func NewScheduler(platform runtime.Platform) *Scheduler {
    return &Scheduler{
        inputQueue:   NewInputQueue(),
        taskQueue:    make([]*InterruptibleTask, 0),
        inputHandler: NewInputHandler(platform),
    }
}

// Start 启动调度器
func (s *Scheduler) Start() {
    s.mu.Lock()
    s.running = true
    s.mu.Unlock()

    go s.run()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.running = false
    if s.currentTask != nil {
        s.currentTask.Cancel()
    }
}

// run 主循环
func (s *Scheduler) run() {
    ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            s.processTasks()

        default:
            // 检查并处理输入（无阻塞）
            if s.inputQueue.HasPending() {
                s.processInput()
            }
        }
    }
}

// processInput 处理输入事件
func (s *Scheduler) processInput() {
    for s.inputQueue.HasPending() {
        // 中断当前任务
        s.interruptCurrentTask()

        // 处理输入
        event := s.inputQueue.Pop()
        if err := event.Handler(); err != nil {
            // 处理错误
        }

        // 输入处理后，重新调度被中断的任务
        s.rescheduleInterruptedTasks()
    }
}

// processTasks 处理任务队列
func (s *Scheduler) processTasks() {
    // 如果有输入事件，优先处理
    if s.inputQueue.HasPending() {
        s.processInput()
        return
    }

    // 处理普通任务
    s.mu.Lock()
    if len(s.taskQueue) == 0 {
        s.mu.Unlock()
        return
    }

    // 取出高优先级任务
    task := s.taskQueue[0]
    s.taskQueue = s.taskQueue[1:]
    s.currentTask = task
    s.mu.Unlock()

    // 执行任务（可中断）
    if err := task.Execute(s.inputQueue); err != nil {
        if err == ErrInterrupted {
            // 被输入中断，重新加入队列
            s.Schedule(task)
        }
    }

    s.currentTask = nil
}

// Schedule 调度任务
func (s *Scheduler) Schedule(task *InterruptibleTask) {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 按优先级插入
    s.taskQueue = insertByPriority(s.taskQueue, task)
}

// interruptCurrentTask 中断当前任务
func (s *Scheduler) interruptCurrentTask() {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.currentTask != nil {
        s.currentTask.Cancel()
    }
}

// rescheduleInterruptedTasks 重新调度被中断的任务
func (s *Scheduler) rescheduleInterruptedTasks() {
    s.mu.Lock()
    defer s.mu.Unlock()

    for _, task := range s.taskQueue {
        if task.Priority == PriorityNormal {
            // 降低被中断任务的重跑优先级
            // 避免饥饿
            task.Priority = PriorityLow
        }
    }
}

// insertByPriority 按优先级插入任务
func insertByPriority(tasks []*InterruptibleTask, task *InterruptibleTask) []*InterruptibleTask {
    for i, t := range tasks {
        if t.Priority < task.Priority {
            // 找到插入位置
            result := make([]*InterruptibleTask, 0, len(tasks)+1)
            result = append(result, tasks[:i]...)
            result = append(result, task)
            result = append(result, tasks[i:]...)
            return result
        }
    }

    // 最低优先级，放最后
    return append(tasks, task)
}
```

---

## 六、Runtime 集成

### 6.1 主循环集成

```go
// core/runtime.go

package core

import (
    "github.com/wwsheng009/mint/scheduler"
)

// Runtime 运行时
type Runtime struct {
    scheduler   *scheduler.Scheduler
    platform    runtime.Platform
    buffer      *paint.Buffer
    root        *reconciler.Fiber
}

// Run 主循环
func (r *Runtime) Run() error {
    // 启动调度器
    r.scheduler.Start()
    defer r.scheduler.Stop()

    for {
        // 检查是否应该退出
        if r.shouldQuit() {
            break
        }

        // 调度器自动处理输入和渲染
        runtime.Gosched()
    }

    return nil
}

// MarkDirty 标记需要重新渲染
func (r *Runtime) MarkDirty() {
    task := scheduler.NewInterruptibleTask(
        "render",
        scheduler.PriorityNormal,
        func(ctx context.Context) error {
            return r.render(ctx)
        },
    )

    r.scheduler.Schedule(task)
}

// render 渲染（可中断）
func (r *Runtime) render(ctx context.Context) error {
    // 检查是否被中断
    select {
    case <-ctx.Done():
        return scheduler.ErrInterrupted
    default:
    }

    // 执行渲染
    reconciler.Update(r.root)
    reconciler.Render(r.root, r.buffer)

    return nil
}
```

---

## 七、性能优化

### 7.1 输入节流

```go
// scheduler/throttle.go

package scheduler

import (
    "sync"
    "time"
)

// Throttler 节流器（用于高频事件）
type Throttler struct {
    mu        sync.Mutex
    lastEvent time.Time
    minInterval time.Duration
    pending   *InputEvent
    timer     *time.Timer
}

// NewThrottler 创建节流器
func NewThrottler(minInterval time.Duration) *Throttler {
    return &Throttler{
        minInterval: minInterval,
    }
}

// Process 处理事件（带节流）
func (t *Throttler) Process(event *InputEvent) {
    t.mu.Lock()
    defer t.mu.Unlock()

    now := time.Now()

    if now.Sub(t.lastEvent) < t.minInterval {
        // 节流中，保存待处理事件
        t.pending = event
        if t.timer == nil {
            t.timer = time.AfterFunc(t.minInterval, t.flush)
        }
        return
    }

    // 立即处理
    t.lastEvent = now
    event.Handler()

    // 处理待处理事件
    if t.pending != nil {
        t.pending.Handler()
        t.pending = nil
    }
}

// flush 刷新待处理事件
func (t *Throttler) flush() {
    t.mu.Lock()
    defer t.mu.Unlock()

    if t.pending != nil {
        t.pending.Handler()
        t.pending = nil
    }
    t.timer = nil
}
```

### 7.2 鼠标移动优化

```go
// scheduler/mouse.go

package scheduler

// MouseMoveHandler 鼠标移动处理器
type MouseMoveHandler struct {
    throttle  *Throttler
    lastX, lastY int
    moved      bool
}

// NewMouseMoveHandler 创建鼠标移动处理器
func NewMouseMoveHandler() *MouseMoveHandler {
    return &MouseMoveHandler{
        throttle: NewThrottler(16 * time.Millisecond), // 60 FPS
    }
}

// Handle 处理鼠标移动
func (h *MouseMoveHandler) Handle(x, y int) {
    // 检查是否真的移动了
    if x == h.lastX && y == h.lastY {
        return
    }

    h.lastX = x
    h.lastY = y
    h.moved = true

    // 节流处理
    event := &InputEvent{
        Type:     "mousemove",
        Priority: PriorityNormal, // 鼠标移动降低优先级
        Handler: func() error {
            return h.processMove(x, y)
        },
    }

    h.throttle.Process(event)
}
```

---

## 八、实施计划

### 阶段 1: 基础实现

- [ ] 实现优先级定义
- [ ] 实现输入队列
- [ ] 实现基础调度器

### 阶段 2: 可中断任务

- [ ] 实现可中断任务
- [ ] 实现任务中断和恢复
- [ ] 实现输入检测

### 阶段 3: 集成优化

- [ ] 集成到 Runtime
- [ ] 实现节流器
- [ ] 优化鼠标移动

### 阶段 4: 测试验证

- [ ] 编写单元测试
- [ ] 性能基准测试
- [ ] 响应性测试

---

## 九、测试策略

```go
// scheduler/scheduler_test.go

func TestInputPriority(t *testing.T) {
    s := NewScheduler(nil)

    // 添加不同优先级的任务
    lowTask := NewInterruptibleTask("low", PriorityLow, func(ctx context.Context) error {
        return nil
    })
    s.Schedule(lowTask)

    highTask := NewInterruptibleTask("high", PriorityImmediate, func(ctx context.Context) error {
        return nil
    })

    // 模拟输入事件
    s.inputQueue.Push(&InputEvent{
        Priority: PriorityImmediate,
        Handler:  func() error { return nil },
    })

    // 高优先级任务应该先执行
    s.processTasks()

    // 验证
    assert.True(t, highTask.Executed)
    assert.False(t, lowTask.Executed)
}

func TestInterruption(t *testing.T) {
    s := NewScheduler(nil)

    interrupted := false
    task := NewInterruptibleTask("test", PriorityNormal, func(ctx context.Context) error {
        for i := 0; i < 100; i++ {
            select {
            case <-ctx.Done():
                interrupted = true
                return ErrInterrupted
            default:
                time.Sleep(1 * time.Millisecond)
            }
        }
        return nil
    })

    go func() {
        time.Sleep(10 * time.Millisecond)
        s.inputQueue.Push(&InputEvent{
            Priority: PriorityImmediate,
            Handler:  func() error { return nil },
        })
    }()

    err := task.Execute(s.inputQueue)

    assert.True(t, interrupted)
    assert.Equal(t, ErrInterrupted, err)
}
```

---

**文档版本**: v1.0
**最后更新**: 2026-01-31
