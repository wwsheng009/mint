# Animation System

动画系统核心实现。

## 职责

- **动画管理（Manager）**：管理多个 `Animation` 实例，可由内部 ticker 或外部时间戳驱动
- **外部驱动器（Drivers）**：`TweenDriver` / `LoopDriver` 供 `framework.App` 这类宿主统一驱动
- **构建器（Builders）**：`FadeIn`、`Pulse`、`Sequence` 等便捷构造器
- **缓动函数**：提供多种缓动算法（Linear、Quad、Cubic、Elastic、Bounce 等）
- **插值计算**：支持数值、字符串等多种类型的插值
- **动画状态机**：`idle`、`running`、`paused`、`completed`、`cancelled`

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：

- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 推荐架构

当前推荐的接线方式是：

1. 组件内部持有 `TweenDriver` 或 `LoopDriver`
2. 组件实现 `rtui.TickableInstance`
3. `framework.App` 在主循环里统一调用 `Tick(now time.Time)`
4. 只有 `WantsTick()` 为 `true` 的实例才会被驱动

这条链路已经用于 `cursor`、`spin`、`progress`、`toast`、`notification`、`tooltip` 等组件。

如果宿主已经有稳定的时间源或渲染循环，优先使用 driver；只有在确实需要“一个独立的动画注册表 + 生命周期管理器”时，再使用 `Manager`。

## 核心概念

### Animation

`Animation` 表示一个由 `Manager` 驱动的动画定义，适合需要统一注册、暂停、恢复、串并行编排的场景。

```go
type Animation struct {
    ID          string
    Type        AnimationType
    From        interface{}
    To          interface{}
    Current     interface{}
    Duration    time.Duration
    Elapsed     time.Duration
    Delay       time.Duration
    State       AnimationState
    Easing      EasingFunction
    Repeat      int           // 0=单次，>0=总播放次数，-1=无限
    Alternate   bool
    RepeatDelay time.Duration
    OnProgress  func(float64)
    OnComplete  func()
}
```

### TweenDriver

`TweenDriver` 是一个不自带 ticker 的标量插值器：

- 适合透明度、缩放、位移、百分比等连续值
- 使用 `Tick(now)` 推进时间
- 使用 `Value()` / `Progress()` 读取当前状态

### LoopDriver

`LoopDriver` 是一个不自带 ticker 的循环计时器：

- 适合帧动画、延迟显示、自动消失、光标闪烁等离散步骤场景
- 使用 `Tick(now)` 推进时间
- 使用 `Progress()`、`StepIndex(steps)`、`Cycle()` 读取循环状态

### AnimationState

```go
const (
    AnimationStateIdle
    AnimationStateRunning
    AnimationStatePaused
    AnimationStateCompleted
    AnimationStateCancelled
)
```

### 缓动函数

| 函数 | 说明 |
|------|------|
| `Linear` | 线性（匀速） |
| `EaseInQuad`, `EaseOutQuad`, `EaseInOutQuad` | 二次方缓动 |
| `EaseInCubic`, `EaseOutCubic`, `EaseInOutCubic` | 三次方缓动 |
| `EaseInElastic`, `EaseOutElastic`, `EaseInOutElastic` | 弹性效果 |
| `EaseInBounce`, `EaseOutBounce`, `EaseInOutBounce` | 弹跳效果 |

## 使用示例

### 推荐：在组件里使用 TweenDriver

```go
type FadeInstance struct {
    opacity float64
    dirty   bool
    fade    *animation.TweenDriver
}

func (inst *FadeInstance) StartFadeIn() {
    inst.fade = animation.NewTweenDriver(animation.TweenDriverConfig{
        From:      0,
        To:        1,
        Duration:  180 * time.Millisecond,
        Easing:    animation.EaseOutQuad,
        AutoStart: true,
    })
    inst.dirty = true
}

func (inst *FadeInstance) WantsTick() bool {
    return inst.fade != nil && inst.fade.WantsTick()
}

func (inst *FadeInstance) Tick(now time.Time) bool {
    if inst.fade == nil || !inst.fade.Tick(now) {
        return false
    }
    inst.opacity = inst.fade.Value()
    inst.dirty = true
    return true
}
```

上面这类实例不需要自己开 ticker。只要它挂在 `framework.App` 的 fiber 树中，`App.handleTick()` 就会在主循环里统一驱动它。

### 推荐：在组件里使用 LoopDriver

```go
type SpinnerInstance struct {
    frame int
    dirty bool
    loop  *animation.LoopDriver
}

func (inst *SpinnerInstance) Start() {
    inst.loop = animation.NewLoopDriver(animation.LoopDriverConfig{
        Duration:  10 * 80 * time.Millisecond,
        Delay:     120 * time.Millisecond,
        Cycles:    0, // 0 = 无限循环
        AutoStart: true,
    })
}

func (inst *SpinnerInstance) WantsTick() bool {
    return inst.loop != nil && inst.loop.WantsTick()
}

func (inst *SpinnerInstance) Tick(now time.Time) bool {
    if inst.loop == nil || !inst.loop.Tick(now) {
        return false
    }
    if !inst.loop.Started() {
        inst.frame = 0
    } else {
        inst.frame = inst.loop.StepIndex(10)
    }
    inst.dirty = true
    return true
}
```

这类写法是 `spin`、`tooltip delay`、`toast auto dismiss`、`notification auto dismiss` 等组件当前采用的模式。

### framework.App 集成方式

`framework.App` 会遍历 fiber 树，找到实现了 `rtui.TickableInstance` 的实例，然后把当前时间传给 `Tick(now)`：

```go
type TickableInstance interface {
    ComponentInstance
    WantsTick() bool
    Tick(now time.Time) bool
}
```

因此组件侧只需要关心：

- 何时返回 `WantsTick() == true`
- `Tick(now)` 是否真的导致了可见状态变化
- 发生变化时记得把实例标记为 dirty

### 使用 Animation Manager

`Manager` 仍然适合以下场景：

- 需要按 ID 管理多条动画
- 需要暂停、恢复、停止、查询运行状态
- 需要 `Animation` builder 和 `Sequence(...)` 这类组合能力
- 宿主不是 `framework.App`，但仍希望集中管理动画

```go
manager := animation.NewManager()

anim := animation.FadeIn("fade-in", 200*time.Millisecond).
    WithRepeat(1).
    WithOnProgress(func(progress float64) {
        fmt.Printf("progress=%.2f\n", progress)
    })

manager.Add(anim)
manager.StartAnimation(anim.ID)

for manager.HasRunning() {
    manager.Tick(time.Now())
}
```

如果你没有外部主循环，也可以让 `Manager` 自己起 ticker：

```go
manager := animation.NewManager()
manager.Add(animation.Pulse("pulse", 0.9, 1.0, time.Second))
manager.StartAnimation("pulse")
manager.Start(60)
defer manager.Stop()
```

已有宿主循环时，优先使用 `Tick(now)`；只有在独立工具、原型或测试场景下，才建议使用 `Start(fps)`。

### 创建简单 Animation

```go
anim := &animation.Animation{
    ID:       "fade-in",
    From:     0.0,
    To:       1.0,
    Duration: 500 * time.Millisecond,
    Easing:   animation.EaseOutQuad,
    OnProgress: func(progress float64) {
        fmt.Printf("进度: %.2f\n", progress)
    },
    OnComplete: func() {
        fmt.Println("动画完成")
    },
}
```

### 重复、交替与延迟

```go
// 无限循环
infiniteAnim := &animation.Animation{
    ID:       "pulse",
    From:     0.5,
    To:       1.0,
    Duration: time.Second,
    Easing:   animation.EaseInOutSine,
    Repeat:   -1,
}

// 交替播放（来回循环）
alternateAnim := &animation.Animation{
    ID:        "breathing",
    From:      0.7,
    To:        1.0,
    Duration:  2 * time.Second,
    Repeat:    -1,
    Alternate: true,
}

// 延迟启动
deferredAnim := &animation.Animation{
    ID:       "delayed-fade",
    From:     0.0,
    To:       1.0,
    Duration: 500 * time.Millisecond,
    Delay:    200 * time.Millisecond,
}
```

`Repeat` 的语义是：

- `0`：单次播放
- `> 0`：总共播放 N 次
- `-1`：无限循环

### 字符串动画

```go
typewriterAnim := &animation.Animation{
    ID:       "typewriter",
    From:     "",
    To:       "Hello, World!",
    Duration: 2 * time.Second,
    Easing:   animation.Linear,
    OnProgress: func(progress float64) {
        n := int(float64(len("Hello, World!")) * progress)
        displayPartialText(n)
    },
}
```

### 构建器

```go
fadeIn := animation.FadeIn("fade-in", 200*time.Millisecond)
progress := animation.Progress("progress", 500*time.Millisecond)
pulse := animation.Pulse("pulse", 0.95, 1.0, time.Second)
```

### 串行动画

```go
seq := animation.Sequence(
    "toast-enter-exit",
    animation.FadeIn("fade-in", 120*time.Millisecond),
    animation.Wait("hold", 2*time.Second),
    animation.FadeOut("fade-out", 120*time.Millisecond),
)

manager.Add(seq)
manager.StartAnimation(seq.ID)
```

`Sequence(...)` 会把子动画拍平成一个有限时长的串行动画，支持子动画上的：

- `Delay`
- 有限 `Repeat`
- `RepeatDelay`
- `Alternate`

不支持把 `Repeat=-1` 的无限循环动画放进 `Sequence(...)`。这类情况会直接 panic，因为整个序列不再具有有限总时长。

## 核心类型

| 类型 | 说明 |
|------|------|
| `TweenDriver` | 外部驱动的连续值插值器 |
| `LoopDriver` | 外部驱动的循环计时器 |
| `Manager` | 动画管理器，管理多个 `Animation` 实例 |
| `Animation` | 适合由 `Manager` 统一调度的动画定义 |
| `AnimationState` | 动画状态枚举 |
| `EasingFunction` | 缓动函数类型：`func(float64) float64` |

## 文件结构

| 文件 | 说明 |
|------|------|
| `drivers.go` | `TweenDriver` / `LoopDriver` 实现 |
| `manager.go` | `Manager` 实现与 `Animation` 调度 |
| `types.go` | `Animation` 及相关类型定义 |
| `builders.go` | Builder、`Sequence`、`Parallel` 等辅助方法 |
| `easing.go` | 缓动函数集合 |

## 最佳实践

### 1. 优先复用宿主时间源

如果宿主已经有主循环、渲染循环或统一 tick 机制，优先使用 driver + `Tick(now)`，不要再额外启动一个动画 ticker。

### 2. 为 Manager 动画设置唯一 ID

```go
anim := &animation.Animation{
    ID: fmt.Sprintf("fade-in-%d", time.Now().UnixNano()),
}
```

不要在同一个 `Manager` 里重复使用固定 ID。

### 3. 区分配置态和运行态

`Animation` 在运行过程中会维护内部播放状态。完成后的实例如果要再次使用，应重新 `Add` 后再 `StartAnimation`，或者先 `Clone()` 再复用。

### 4. 清理完成的动画

`Manager` 会自动清理已完成动画。如果你需要在外部保留结果，请把最终状态写回组件字段或业务状态，而不是依赖 `Manager` 长期持有已完成实例。

### 5. 选择合适的驱动模型

- 组件内部的小型时间逻辑：优先 `TweenDriver` / `LoopDriver`
- 需要按 ID 控制、暂停恢复、统一编排：使用 `Manager`
- 帧序列、闪烁、延迟显示：优先 `LoopDriver`
- 连续数值插值：优先 `TweenDriver`

### 6. 使用合适的缓动函数

```go
FadeIn:     animation.EaseOutQuad
FadeOut:    animation.EaseInQuad
SlideRight: animation.EaseInOutCubic
BounceIn:   animation.EaseOutBounce
Pulse:      animation.EaseInOutSine
```

### 7. 只在状态真的变化时返回 true

无论是 driver 还是 `Manager` 回调，只有在可见状态发生变化时才应标记 dirty 并触发后续渲染。

## 与其他模块集成

### 与 framework.App 集成

推荐通过 `TickableInstance` 接口接入，让 `framework.App` 统一推进时间。

### 与 Paint 集成

动画通常只更新组件的运行态字段，例如透明度、帧索引、位移，再由 `Paint()` 使用这些字段生成 draw commands。

### 与 Focus / Action 集成

如果动画由交互触发，建议在 action 或组件事件里只做“启动动画”这一步，不要在事件处理器内部自行起 goroutine/ticker。

### 与测试集成

- driver 测试优先使用显式 `Tick(fixedTime)`
- `Manager` 测试优先使用 `Tick(now)`，避免依赖真实时间
- 对 delay、repeat、alternate 这类边界行为写回归测试
