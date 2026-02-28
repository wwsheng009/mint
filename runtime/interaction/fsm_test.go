package interaction

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/input"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// 测试用组件
type TestInteractionComponent struct {
	id              int
	clicked         bool
	cancelled       bool
	pressedReset    bool
	bounds          Bounds
}

type Bounds struct {
	X, Y, Width, Height int
}

func (c *TestInteractionComponent) OnClick() {
	c.clicked = true
}

func (c *TestInteractionComponent) OnCancel() {
	c.cancelled = true
}

func (c *TestInteractionComponent) ResetPressed() {
	c.pressedReset = true
}

func (c *TestInteractionComponent) Contains(x, y int) bool {
	return x >= c.bounds.X && x < c.bounds.X+c.bounds.Width &&
		y >= c.bounds.Y && y < c.bounds.Y+c.bounds.Height
}

func TestInteractionContext_Click(t *testing.T) {
	ctx := NewInteractionContext()

	// 注册测试组件：位于 (10,10), 大小 20x20
	comp := &TestInteractionComponent{
		id:     1,
		bounds: Bounds{X: 10, Y: 10, Width: 20, Height: 20},
	}
	ctx.RegisterInstance(1, comp)

	// 命中测试函数
	hitTest := func(x, y int) int {
		if comp.Contains(x, y) {
			return 1
		}
		return 0
	}

	// 按下
	intents := []input.InputIntent{
		input.InputPressIntent{
			X:      15,
			Y:      15,
			Button: runtimemsg.MouseLeft,
			Source: "mouse",
		},
	}
	ctx.Update(intents, hitTest)

	if ctx.ActiveID != 1 {
		t.Errorf("Expected ActiveID=1, got %d", ctx.ActiveID)
	}
	if ctx.State != StatePressed {
		t.Errorf("Expected State=Pressed, got %d", ctx.State)
	}
	if ctx.StartX != 15 || ctx.StartY != 15 {
		t.Errorf("Expected start (15,15), got (%d,%d)", ctx.StartX, ctx.StartY)
	}

	// 在同一位置释放
	intents = []input.InputIntent{
		input.InputReleaseIntent{
			X:      15,
			Y:      15,
			Button: runtimemsg.MouseLeft,
			Source: "mouse",
		},
	}
	ctx.Update(intents, hitTest)

	if !comp.clicked {
		t.Error("Expected clicked=true")
	}
	if ctx.ActiveID != 0 {
		t.Errorf("Expected ActiveID=0, got %d", ctx.ActiveID)
	}
}

func TestInteractionContext_DragAndCancel(t *testing.T) {
	ctx := NewInteractionContext()

	// 注册测试组件：位于 (10,10), 大小 20x20
	comp := &TestInteractionComponent{
		id:     1,
		bounds: Bounds{X: 10, Y: 10, Width: 20, Height: 20},
	}
	ctx.RegisterInstance(1, comp)

	// 命中测试函数
	hitTest := func(x, y int) int {
		if comp.Contains(x, y) {
			return 1
		}
		return 0
	}

	// 按下
	intents := []input.InputIntent{
		input.InputPressIntent{
			X:      15,
			Y:      15,
			Button: runtimemsg.MouseLeft,
			Source: "mouse",
		},
	}
	ctx.Update(intents, hitTest)

	// 移动到组件外部（超出拖拽阈值）
	intents = []input.InputIntent{
		input.InputMoveIntent{X: 50, Y: 50},
	}
	ctx.Update(intents, hitTest)

	// 应该进入拖拽状态
	if ctx.State != StateDragging {
		t.Errorf("Expected State=Dragging, got %d", ctx.State)
	}

	// 在外部释放
	intents = []input.InputIntent{
		input.InputReleaseIntent{
			X:      50,
			Y:      50,
			Button: runtimemsg.MouseLeft,
			Source: "mouse",
		},
	}
	ctx.Update(intents, hitTest)

	// 应该触发 cancel
	if !comp.cancelled {
		t.Error("Expected cancelled=true")
	}
	if comp.clicked {
		t.Error("Expected clicked=false (should be cancel)")
	}
}

func TestInteractionContext_KeyboardReset(t *testing.T) {
	ctx := NewInteractionContext()

	// 注册测试组件
	comp := &TestInteractionComponent{
		id:     1,
		bounds: Bounds{X: 10, Y: 10, Width: 20, Height: 20},
	}
	ctx.RegisterInstance(1, comp)

	// 注册另一个测试组件
	comp2 := &TestInteractionComponent{
		id:     2,
		bounds: Bounds{X: 100, Y: 100, Width: 20, Height: 20},
	}
	ctx.RegisterInstance(2, comp2)

	// 新键盘输入
	intents := []input.InputIntent{
		input.InputKeyboardIntent{
			Key: 'a',
		},
	}
	ctx.Update(intents, nil)

	// 验证所有组件的 pressed 状态都被重置
	if !comp.pressedReset {
		t.Error("Expected comp.pressedReset=true")
	}
	if !comp2.pressedReset {
		t.Error("Expected comp2.pressedReset=true")
	}
}

func TestInteractionContext_SmallDrag(t *testing.T) {
	ctx := NewInteractionContext()

	// 注册测试组件：位于 (10,10), 大小 20x20
	comp := &TestInteractionComponent{
		id:     1,
		bounds: Bounds{X: 10, Y: 10, Width: 20, Height: 20},
	}
	ctx.RegisterInstance(1, comp)

	// 命中测试函数
	hitTest := func(x, y int) int {
		if comp.Contains(x, y) {
			return 1
		}
		return 0
	}

	// 按下
	intents := []input.InputIntent{
		input.InputPressIntent{
			X:      15,
			Y:      15,
			Button: runtimemsg.MouseLeft,
			Source: "mouse",
		},
	}
	ctx.Update(intents, hitTest)

	// 小幅度移动（在拖拽阈值内）
	// DragThreshold = 3，移动 2 像素不应该触发拖拽
	intents = []input.InputIntent{
		input.InputMoveIntent{X: 17, Y: 17},
	}
	ctx.Update(intents, hitTest)

	// 应该保持 Pressed 状态
	if ctx.State != StatePressed {
		t.Errorf("Expected State=Pressed, got %d", ctx.State)
	}

	// 在组件内释放
	intents = []input.InputIntent{
		input.InputReleaseIntent{
			X:      17,
			Y:      17,
			Button: runtimemsg.MouseLeft,
			Source: "mouse",
		},
	}
	ctx.Update(intents, hitTest)

	// 应该触发 click（没有触发拖拽）
	if !comp.clicked {
		t.Error("Expected clicked=true")
	}
	if comp.cancelled {
		t.Error("Expected cancelled=false")
	}
}
