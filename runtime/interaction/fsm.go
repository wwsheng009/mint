package interaction

import (
	"github.com/wwsheng009/mint/runtime/input"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// InteractionState 交互状态枚举
type InteractionState int

const (
	StateIdle     InteractionState = iota // 空闲
	StateHover                             // 悬停
	StatePressed                           // 按下
	StateDragging                          // 拖拽中
	StateSelecting                         // 选择中
)

// InteractionContext 交互上下文
//
// 跟踪全局的交互状态：
// - HotID: 当前鼠标悬停的组件
// - ActiveID: 当前按下的组件
type InteractionContext struct {
	HotID    int  // 当前 hover 的组件 ID
	ActiveID int  // 当前按下的组件 ID

	// 按下时的起始位置（用于拖拽判断）
	StartX, StartY int

	// 当前状态
	State InteractionState

	// 组件注册表（ID → Instance）
	Instances map[int]interface{}
}

// NewInteractionContext 创建新的交互上下文
func NewInteractionContext() *InteractionContext {
	return &InteractionContext{
		Instances: make(map[int]interface{}),
	}
}

// RegisterInstance 注册交互组件
func (c *InteractionContext) RegisterInstance(id int, inst interface{}) {
	if c.Instances == nil {
		c.Instances = make(map[int]interface{})
	}
	c.Instances[id] = inst
}

// UnregisterInstance 注销交互组件
func (c *InteractionContext) UnregisterInstance(id int) {
	delete(c.Instances, id)
}

// Update 更新交互状态
//
// 核心逻辑：
// - InputMove: 更新 HotID
// - InputPress: 设置 ActiveID
// - InputRelease: 处理 Click/Cancel
// - InputKeyboard: 重置所有 pressed 状态
func (c *InteractionContext) Update(intents []input.InputIntent, hitTest func(int, int) int) {
	for _, intent := range intents {
		switch e := intent.(type) {
		case input.InputMoveIntent:
			c.handleMove(e.X, e.Y, hitTest)

		case input.InputPressIntent:
			c.handlePress(e.X, e.Y, hitTest, e.Source)

		case input.InputReleaseIntent:
			c.handleRelease(e.X, e.Y, hitTest, e.Source)

		case input.InputKeyboardIntent:
			c.handleKeyboard(e.Key, e.Special, e.Mod)
		}
	}
}

// handleMove 处理鼠标移动
func (c *InteractionContext) handleMove(x, y int, hitTest func(int, int) int) {
	id := hitTest(x, y)
	c.HotID = id

	// 检查是否进入拖拽状态
	if c.ActiveID != 0 {
		if abs(x-c.StartX) > DragThreshold || abs(y-c.StartY) > DragThreshold {
			c.State = StateDragging
		}
	}
}

// handlePress 处理按下事件
func (c *InteractionContext) handlePress(x, y int, hitTest func(int, int) int, source string) {
	id := hitTest(x, y)

	if id != 0 {
		c.ActiveID = id
		c.StartX = x
		c.StartY = y

		if source == "mouse" {
			c.State = StatePressed
		}
		// Keyboard press uses different logic (handled elsewhere)
	}
}

// handleRelease 处理释放事件
func (c *InteractionContext) handleRelease(x, y int, hitTest func(int, int) int, source string) {
	if c.ActiveID != 0 {
		targetID := hitTest(x, y)

		// 鼠标释放
		if source == "mouse" {
			if c.ActiveID == targetID {
				// Click：在同一组件内按下并释放
				c.dispatchClick(c.ActiveID)
			} else {
				// Cancel：拖出后释放
				c.dispatchCancel(c.ActiveID)
			}
		}

		c.ActiveID = 0
		if c.State == StateDragging {
			c.State = StateIdle
		}
	}
}

// handleKeyboard 处理键盘输入
//
// 根据 docs/event/PRESSED_STATE_COMPLETE_SOLUTION.md 的设计原则：
// 新的键盘输入应该重置所有交互状态
func (c *InteractionContext) handleKeyboard(key rune, special runtimeplatform.SpecialKey, mod runtimemsg.Modifiers) {
	c.resetAllPressedStates()
}

// dispatchClick 分发 click 事件
func (c *InteractionContext) dispatchClick(id int) {
	if inst, ok := c.Instances[id].(ClickHandler); ok {
		inst.OnClick()
	}
}

// dispatchCancel 分发 cancel 事件
func (c *InteractionContext) dispatchCancel(id int) {
	if inst, ok := c.Instances[id].(CancelHandler); ok {
		inst.OnCancel()
	}
}

// resetAllPressedStates 重置所有组件的 pressed 状态
func (c *InteractionContext) resetAllPressedStates() {
	for _, inst := range c.Instances {
		if handler, ok := inst.(PressedResetHandler); ok {
			handler.ResetPressed()
		}
	}
}

// 交互处理接口
type ClickHandler interface {
	OnClick()
}

type CancelHandler interface {
	OnCancel()
}

// PressedResetHandler 是需要能够重置 pressed 状态的组件接口
//
// 注意：这个接口设计需要组件能够访问 Instance（为了更新 InteractionState.Pressed）。
// 在实际使用中，通常组件会实现这个接口并在 ResetPressed 中通过闭包或其他方式访问 Instance。
//
// 如果组件行为（Behavior）需要访问 Instance，设计如下：
//
//   type Button struct {
//       PressableBehavior *control.PressableBehavior
//       instance          control.Instance
//   }
//
//   func (b *Button) ResetPressed() {
//       b.PressableBehavior.ResetPressedWithInstance(b.instance)
//   }
type PressedResetHandler interface {
	ResetPressed()
}

// PressedResetWithInstance 是带 Instance 参数的重置接口
// 由 Behavior 实现内部逻辑
type PressedResetWithInstance interface {
	ResetPressedWithInstance(inst interface{})
}

// 常量
const DragThreshold = 3 // 拖拽判定阈值（像素）

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
