# StayPressedIntent - Intent 视觉反馈控制

## 概述

`StayPressedIntent` 是一个接口，用于控制当用户通过键盘（Enter/Submit）触发按钮后，是否应该保持按下的视觉状态。

由于终端 UI **不会发送键盘释放事件**，因此需要在按下时决定是否重置视觉状态。

## 接口定义

```go
type StayPressedIntent interface {
    intent.Intent

    // StayPressed 返回 true 表示保持按下视觉效果，false 表示立即重置
    StayPressed() bool
}
```

## 使用场景

### 🟢 `StayPressed() = true` - 保持按下效果

适用于**导航、状态更新**等需要给用户视觉反馈的场景：

```go
type UpdateStepIntent struct {
    Step int
}

func (UpdateStepIntent) IntentType() string { return "UpdateStep" }
func (UpdateStepIntent) StayPressed() bool  { return true } // 用户可以看到按钮短暂地被按下
```

**效果：**
```
用户按 Enter
  ├─ 按钮显示为按下状态（Pressed = true）
  ├─ 发射 Intent，更新状态
  └─ 下一帧渲染（可能显示新的内容）
```

### 🔴 `StayPressed() = false` - 立即重置

适用于**取消性、破坏性操作**等不需要视觉反馈的场景：

```go
type QuitIntent struct{}

func (QuitIntent) IntentType() string { return "Quit" }
func (QuitIntent) StayPressed() bool  { return false } // 立即恢复
```

**效果：**
```
用户按 Enter
  ├─ 按钮瞬间显示为按下，立即恢复
  └─ 发射 Intent，退出应用
```

## 默认行为

如果 Intent **未实现** `StayPressedIntent` 接口，则默认行为是**立即重置**（向后兼容）：

```go
type SomeIntent struct{}

func (SomeIntent) IntentType() string { return "SomeType" }
// 未实现 StayPressed() -> 默认立即重置
```

## 完整示例

```go
package main

import (
    "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/ui/components/button"
    app "github.com/wwsheng009/mint/ui/components"
)

// 导航类 Intent - 保持视觉反馈
type NavigateIntent struct {
    Page string
}

func (NavigateIntent) IntentType() string { return "Navigate" }
func (NavigateIntent) StayPressed() bool  { return true }

// 警告类 Intent - 立即重置
type DeleteIntent struct {
    ID string
}

func (DeleteIntent) IntentType() string { return "Delete" }
func (DeleteIntent) StayPressed() bool  { return false }

func MyComponent() ui.VNode {
    return ui.VStack(
        app.ButtonBuilder("Next Page").
            OnPress(NavigateIntent{Page: "/settings"}).
            Build(),

        app.ButtonBuilder("Delete Item").
            OnPress(DeleteIntent{ID: "123"}).
            Build(),
    )
}
```

## 实现细节

### PressableBehavior 中的处理逻辑

```go
case action.ActionEnter, action.ActionSubmit:
    if b.pressIntent != nil {
        inst.EmitIntent(b.pressIntent)

        // 检查是否实现 StayPressedIntent
        var shouldResetPressed bool
        if sp, ok := b.pressIntent.(StayPressedIntent); ok {
            shouldResetPressed = !sp.StayPressed()
        } else {
            shouldResetPressed = true  // 默认立即重置
        }

        if shouldResetPressed {
            b.pressed = false
            state.Pressed = false
            inst.MarkDirty()
        }
        // 否则保持 pressed，下一帧自然更新
    }
```

## 迁移指南

### 现有代码保持兼容

不实现 `StayPressed()` 的现有代码会自动获得默认行为（立即重置），**无需修改**。

### 逐步迁移（推荐）

对于需要视觉反馈的 Intent，逐步添加 `StayPressed()` 方法：

```diff
  type UpdateStepIntent struct {
      Step int
  }

  func (UpdateStepIntent) IntentType() string { return "UpdateStep" }
+ func (UpdateStepIntent) StayPressed() bool  { return true }
```

## 最佳实践

| Intent 类型 | StayPressed() | 理由 |
|------------|---------------|------|
| 导航/分页 | `true` | 用户需要看到操作反馈 |
| 状态更新 | `true` | 明确表示状态已改变 |
| 打开 Modal/Dialog | `true` | 视觉确认操作 |
| 退出/取消 | `false` | 即将离开，反馈无意义 |
| 删除/销毁 | `false` | 破坏性操作，避免误触 |
| 切换开关 | `true` | 清晰显示状态切换 |

## 相关文件

- `ui/components/control/types.go` - 接口定义和实现
- `examples/ant_design_demo/main.go` - 完整示例
