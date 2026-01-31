# Animation System

动画系统核心实现。

## 职责

- **动画管理（Manager）**：全局管理多个动画实例，自动更新帧
- **缓动函数**：提供多种缓动算法（Linear, Quad, Cubic, elastic 等）
- **插值计算**：支持数值、字符串等多种类型的插值
- **动画状态机**：Idle, Running, Paused, Completed, Cancelled

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 核心概念

### Animation（动画实例）

`Animation` 定义单个动画的行为：

```go
type Animation struct {
    ID          string              // 唯一标识
    From        interface{}         // 起始值
    To          interface{}         // 目标值
    Current     interface{}         // 当前值
    Duration    time.Duration       // 持续时间
    Elapsed     time.Duration       // 已过去时间
    Delay       time.Duration       // 延迟启动
    State       AnimationState      // 状态
    Easing      EasingFunction      // 缓动函数
    Repeat      int                 // 重复次数（0=无限，1=None）
    Alternate   bool                // 交替播放
    RepeatDelay time.Duration       // 重复延迟
    OnProgress  func(float64)       // 进度回调
    OnComplete  func()              // 完成回调
}
```

### 动画状态机

```go
type AnimationState int

const (
    AnimationStateIdle       // 空闲（未开始）
    AnimationStateRunning    // 运行中
    AnimationStatePaused     // 暂停
    AnimationStateCompleted  // 完成
    AnimationStateCancelled  // 取消
)
```

### 缓动函数

缓动函数控制动画的加速度/减速度：

| 函数 | 说明 |
|------|------|
| `Linear` | 线性（匀速） |
| `EaseInQuad`, `EaseOutQuad`, `EaseInOutQuad` | 二次方缓动 |
| `EaseInCubic`, `EaseOutCubic`, `EaseInOutCubic` | 三次方缓动 |
| `EaseInElastic`, `EaseOutElastic`, `EaseInOutElastic` | 弹性效果 |
| `EaseInBounce`, `EaseOutBounce`, `EaseInOutBounce` | 弹跳效果 |

## 使用示例

### 创建简单动画

```go
import "github.com/wwsheng009/mint/runtime/animation"

// 创建数值动画
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
        fmt.Println("动画完成！")
    },
}
```

### 使用 Animation Manager

```go
// 创建并启动管理器
manager := animation.NewManager()
manager.Start(60) // 60 FPS

// 添加动画
manager.Add(anim)

// 控制动画
manager.StartAnimation("fade-in")
//manager.PauseAnimation("fade-in")
//manager.StopAnimation("fade-in")

// 查询状态
if manager.HasRunning() {
    fmt.Println("有 %d 个动画正在运行", manager.GetRunningCount())
}

// 停止管理器
defer manager.Stop()
```

### 数值动画（透明度、尺寸等）

```go
func AnimateOpacity() {
    anim := &animation.Animation{
        ID:       "opacity",
        From:     0.0,
        To:       1.0,
        Duration: 300 * time.Millisecond,
        Easing:   animation.EaseInOutCubic,
    }
    manager.Add(anim)
    manager.StartAnimation("opacity")
}
```

### 重复动画

```go
// 无限循环
infiniteAnim := &animation.Animation{
    ID:       "pulse",
    From:     0.5,
    To:       1.0,
    Duration: 1000 * time.Millisecond,
    EaseType: "ease-in-out-sine",
    Repeat:   0, // 0 = 无限
}

// 交替播放（来回循环）
alternateAnim := &animation.Animation{
    ID:        "breathing",
    From:      0.7,
    To:        1.0,
    Duration:  2000 * time.Millisecond,
    Repeat:    0,
    Alternate: true,
}
```

### 弹跳效果

```go
bounceAnim := &animation.Animation{
    ID:       "bounce-in",
    From:     -50.0, // 从上方进入
    To:       0.0,
    Duration: 800 * time.Millisecond,
    Easing:   animation.EaseOutBounce,
}
```

### 带延迟的动画

```go
deferredAnim := &animation.Animation{
    ID:       "delayed-fade",
    From:     0.0,
    To:       1.0,
    Duration: 500 * time.Millisecond,
    Delay:    200 * time.Millisecond, // 延迟 200ms 开始
}
```

### 字符串动画（打字机效果）

```go
typewriterAnim := &animation.Animation{
    ID:       "typewriter",
    From:     "",
    To:       "Hello, World!",
    Duration: 2000 * time.Millisecond,
    Easing:   animation.Linear,
    OnProgress: func(progress float64) {
        len := int(float64(len("Hello, World!")) * progress)
        displayPartialText(toString, len)
    },
}
```

### 链式动画

```go
func ChainAnimations(manager *animation.Manager) {
    seq := action.Sequence(
        action.ActionFunc(func(ctx context.Context) action.ActionResult {
            // 动画 1：淡入
            anim1 := createFadeInAnimation()
            manager.Add(anim1)
            manager.StartAnimation(anim1.ID)
            return action.OKAction
        }),
        action.ActionFunc(func(ctx context.Context) action.ActionResult {
            // 动画 2：滑动
            anim2 := createSlideAnimation()
            manager.Add(anim2)
            manager.StartAnimation(anim2.ID)
            return action.OKAction
        }),
    )
    seq.Execute(context.Background())
}
```

### 串行动画示例

```go
// 动画完成后自动触发下一个动画
onComplete := func() {
    nextAnim := createNextAnimation()
    manager.Add(nextAnim)
    manager.StartAnimation(nextAnim.ID)
}

firstAnim := &animation.Animation{
    ID:         "first",
    From:       0,
    To:         100,
    Duration:   500 * time.Millisecond,
    Easing:     animation.EaseOutQuad,
    OnComplete: onComplete,
}
```

## 核心类型

| 类型 | 说明 |
|------|------|
| `Manager` | 动画管理器，管理多个动画实例 |
| `Animation` | 动画实例定义 |
| `AnimationState` | 动画状态枚举 |
| `EasingFunction` | 缓动函数类型：`func(float64) float64` |

## 文件结构

| 文件 | 说明 |
|------|------|
| `manager.go` | Manager 实现，全局动画调度 |
| `types.go` | Animation 类型定义（如果有） |
| `builders.go` | 动画构建器（如果有） |
| `easing.go` | 缓动函数集合（30+ 种缓动） |

## 最佳实践

### 1. 为动画设置唯一 ID

```go
// 推荐：使用前缀
anim := &animation.Animation{
    ID: fmt.Sprintf("fade-in-%d", time.Now().UnixNano()),
}

// 不推荐：固定 ID（会导致冲突）
anim := &animation.Animation{
    ID: "fade", // 多个动画会冲突
}
```

### 2. 清理完成的动画

Manager 会自动清理不重复的已完成动画。

### 3. 使用合适的缓动函数

```go
// 淡入/淡出：使用 Out
FadeIn:    animation.EaseOutQuad
FadeOut:   animation.EaseInQuad

// UI 过渡：使用 InOut
SlideRight: animation.EaseInOutCubic

// 特殊效果：使用特殊缓动
BounceIn:   animation.EaseOutBounce
Pulse:      animation.EaseInOutSine
```

### 4. 限制动画数量

```go
if manager.GetRunningCount() > 10 {
    // 限制：最多同时运行 10 个动画
    fmt.Println("动画数量过多，跳过新动画")
    return
}
```

### 5. 性能优化

```go
// 使用 requestAnimationFrame 模式
func UpdateUIWithAnimation() {
    if manager.HasRunning() {
        // 只有在动画运行时才更新 UI
        UpdateUI()
    }
}
```

## 与其他模块集成

### 与 Paint 集成

```go
// 动画控制透明度或颜色
anim.OnProgress = func(progress float64) {
    opacity := anim.From.(float64) + (anim.To.(float64)-anim.From.(float64))*progress
    buffer.SetAlpha(0, 0, opacity)
}
```

### 与 Focus 集成

```go
// 焦点切换动画
func AnimateFocusChange(fm *focus.Manager, fromID, toID string) {
    anim := &animation.Animation{
        ID:       "focus-transition",
        From:     fromID,
        To:       toID,
        Duration: 150 * time.Millisecond,
        Easing:   animation.EaseOutQuad,
    }
    manager.Add(anim)
}
```

### 与 DevTools 集成

用于时间旅行调试和动画回放。
