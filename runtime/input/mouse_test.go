package input

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime/platform"
)

// TestMouseTracker_New 测试鼠标追踪器创建
func TestMouseTracker_New(t *testing.T) {
	tracker := NewMouseTracker()

	assert.NotNil(t, tracker)
	assert.Equal(t, 500*time.Millisecond, tracker.doubleClickTimeout)
	assert.Equal(t, 5, tracker.doubleClickDistance)
	assert.False(t, tracker.isDragging)
}

// TestMouseTracker_Move 测试鼠标移动
func TestMouseTracker_Move(t *testing.T) {
	tracker := NewMouseTracker()

	input := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MouseMotion,
		MouseX:     100,
		MouseY:     200,
	}

	result := tracker.ProcessInput(input)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Action)
	assert.Equal(t, "mouse_motion", string(result.Action.Type))
	assert.False(t, result.IsDragMove)

	// 验证鼠标坐标更新
	payload := result.Action.Payload.(MouseEventPayload)
	assert.Equal(t, 100, payload.X)
	assert.Equal(t, 200, payload.Y)
}

// TestMouseTracker_Click 测试鼠标单击
func TestMouseTracker_Click(t *testing.T) {
	tracker := NewMouseTracker()

	input := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}

	result := tracker.ProcessInput(input)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Action)
	assert.Equal(t, "mouse_click", string(result.Action.Type))
	assert.False(t, result.IsDoubleClick)
	assert.False(t, result.IsTripleClick)
	assert.True(t, result.IsDragStart)

	payload := result.Action.Payload.(MouseEventPayload)
	assert.Equal(t, 100, payload.X)
	assert.Equal(t, 200, payload.Y)
	assert.Equal(t, platform.MouseLeft, payload.Button)
}

// TestMouseTracker_RightClick 测试鼠标右键单击
func TestMouseTracker_RightClick(t *testing.T) {
	tracker := NewMouseTracker()

	input := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseRight,
		MouseX:     100,
		MouseY:     200,
	}

	result := tracker.ProcessInput(input)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Action)
	assert.Equal(t, "mouse_right_click", string(result.Action.Type))

	payload := result.Action.Payload.(MouseEventPayload)
	assert.Equal(t, platform.MouseRight, payload.Button)
}

// TestMouseTracker_MiddleClick 测试鼠标中键单击
func TestMouseTracker_MiddleClick(t *testing.T) {
	tracker := NewMouseTracker()

	input := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseMiddle,
		MouseX:     100,
		MouseY:     200,
	}

	result := tracker.ProcessInput(input)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Action)
	assert.Equal(t, "mouse_middle_click", string(result.Action.Type))

	payload := result.Action.Payload.(MouseEventPayload)
	assert.Equal(t, platform.MouseMiddle, payload.Button)
}

// TestMouseTracker_DoubleClick 测试鼠标双击
func TestMouseTracker_DoubleClick(t *testing.T) {
	tracker := NewMouseTracker()
	tracker.doubleClickTimeout = 100 * time.Millisecond // 缩短测试时间

	// 第一次点击
	input1 := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}
	result1 := tracker.ProcessInput(input1)
	assert.False(t, result1.IsDoubleClick)

	// 短暂等待
	time.Sleep(50 * time.Millisecond)

	// 第二次点击（相同位置，相同按钮，在超时时间内）
	input2 := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     102, // 小于 doubleClickDistance (5)
		MouseY:     201,
	}
	result2 := tracker.ProcessInput(input2)

	assert.NotNil(t, result2)
	assert.True(t, result2.IsDoubleClick)
	assert.Equal(t, "mouse_double_click", string(result2.Action.Type))
}

// TestMouseTracker_TripleClick 测试鼠标三击
func TestMouseTracker_TripleClick(t *testing.T) {
	tracker := NewMouseTracker()
	tracker.doubleClickTimeout = 100 * time.Millisecond

	// 第一次点击
	input1 := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}
	tracker.ProcessInput(input1)

	time.Sleep(50 * time.Millisecond)

	// 第二次点击
	input2 := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}
	result2 := tracker.ProcessInput(input2)
	assert.True(t, result2.IsDoubleClick)

	time.Sleep(50 * time.Millisecond)

	// 第三次点击
	input3 := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}
	result3 := tracker.ProcessInput(input3)

	assert.NotNil(t, result3)
	assert.True(t, result3.IsTripleClick)
	assert.Equal(t, "mouse_triple_click", string(result3.Action.Type))
}

// TestMouseTracker_DoubleClick_Timeout 测试双击超时
func TestMouseTracker_DoubleClick_Timeout(t *testing.T) {
	tracker := NewMouseTracker()
	tracker.doubleClickTimeout = 50 * time.Millisecond

	// 第一次点击
	input1 := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}
	result1 := tracker.ProcessInput(input1)
	assert.False(t, result1.IsDoubleClick)

	// 等待超过超时时间
	time.Sleep(100 * time.Millisecond)

	// 第二次点击（超时后）
	input2 := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}
	result2 := tracker.ProcessInput(input2)

	assert.False(t, result2.IsDoubleClick)
	assert.Equal(t, "mouse_click", string(result2.Action.Type))
}

// TestMouseTracker_DoubleClick_Distance 测试双击位置距离
func TestMouseTracker_DoubleClick_Distance(t *testing.T) {
	tracker := NewMouseTracker()
	tracker.doubleClickDistance = 5

	// 第一次点击
	input1 := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}
	tracker.ProcessInput(input1)

	time.Sleep(50 * time.Millisecond)

	// 第二次点击（超出距离限制）
	input2 := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     110, // 距离第一次点击 > 5
		MouseY:     200,
	}
	result2 := tracker.ProcessInput(input2)

	assert.False(t, result2.IsDoubleClick)
}

// TestMouseTracker_Drag 测试鼠标拖拽
func TestMouseTracker_Drag(t *testing.T) {
	tracker := NewMouseTracker()

	// 按下鼠标（开始拖拽）
	pressInput := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}
	pressResult := tracker.ProcessInput(pressInput)

	assert.True(t, pressResult.IsDragStart)
	assert.Equal(t, 100, pressResult.DragStartX)
	assert.Equal(t, 200, pressResult.DragStartY)

	// 移动鼠标（拖拽中）
	moveInput := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MouseMotion,
		MouseX:     150,
		MouseY:     250,
	}
	moveResult := tracker.ProcessInput(moveInput)

	assert.True(t, moveResult.IsDragMove)
	assert.Equal(t, 100, moveResult.DragStartX)
	assert.Equal(t, 200, moveResult.DragStartY)
	assert.Equal(t, 50, moveResult.DragDeltaX)
	assert.Equal(t, 50, moveResult.DragDeltaY)

	payload := moveResult.Action.Payload.(MouseEventPayload)
	assert.Equal(t, 150, payload.X)
	assert.Equal(t, 250, payload.Y)
	assert.Equal(t, 100, payload.StartX)
	assert.Equal(t, 200, payload.StartY)
	assert.Equal(t, 50, payload.DeltaX)
	assert.Equal(t, 50, payload.DeltaY)
}

// TestMouseTracker_DragEnd 测试拖拽结束
func TestMouseTracker_DragEnd(t *testing.T) {
	tracker := NewMouseTracker()

	// 开始拖拽
	tracker.ProcessInput(platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	})

	// 释放鼠标
	releaseInput := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MouseRelease,
		MouseButton: platform.MouseLeft,
		MouseX:     150,
		MouseY:     250,
	}
	result := tracker.ProcessInput(releaseInput)

	assert.True(t, result.IsDragEnd)
	assert.Equal(t, 100, result.DragStartX)
	assert.Equal(t, 200, result.DragStartY)
	assert.Equal(t, 50, result.DragDeltaX)
	assert.Equal(t, 50, result.DragDeltaY)
}

// TestMouseTracker_SmallDrag 测试小距离拖拽（视为点击）
func TestMouseTracker_SmallDrag(t *testing.T) {
	tracker := NewMouseTracker()
	tracker.doubleClickDistance = 5

	// 开始拖拽
	tracker.ProcessInput(platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	})

	// 释放鼠标（距离很小）
	releaseInput := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MouseRelease,
		MouseButton: platform.MouseLeft,
		MouseX:     103, // 偏移 3 < 5
		MouseY:     202, // 偏移 2 < 5
	}
	result := tracker.ProcessInput(releaseInput)

	assert.False(t, result.IsDragEnd) // 不视为拖拽结束
}

// TestMouseTracker_WheelUp 测试鼠标滚轮向上
func TestMouseTracker_WheelUp(t *testing.T) {
	tracker := NewMouseTracker()

	input := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MouseWheelUp,
		MouseX:     100,
		MouseY:     200,
	}

	result := tracker.ProcessInput(input)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Action)
	assert.Equal(t, "mouse_wheel_up", string(result.Action.Type))
}

// TestMouseTracker_WheelDown 测试鼠标滚轮向下
func TestMouseTracker_WheelDown(t *testing.T) {
	tracker := NewMouseTracker()

	input := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MouseWheelDown,
		MouseX:     100,
		MouseY:     200,
	}

	result := tracker.ProcessInput(input)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Action)
	assert.Equal(t, "mouse_wheel_down", string(result.Action.Type))
}

// TestMouseTracker_Release 测试鼠标释放
func TestMouseTracker_Release(t *testing.T) {
	tracker := NewMouseTracker()

	input := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MouseRelease,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}

	result := tracker.ProcessInput(input)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Action)
	assert.Equal(t, "mouse_release", string(result.Action.Type))

	payload := result.Action.Payload.(MouseEventPayload)
	assert.Equal(t, 100, payload.X)
	assert.Equal(t, 200, payload.Y)
	assert.Equal(t, platform.MouseLeft, payload.Button)
}

// TestMouseTracker_DifferentButtons 测试不同鼠标按钮
func TestMouseTracker_DifferentButtons(t *testing.T) {
	tracker := NewMouseTracker()
	tracker.doubleClickTimeout = 100 * time.Millisecond

	// 左键第一次点击
	tracker.ProcessInput(platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	})

	time.Sleep(50 * time.Millisecond)

	// 右键点击（不同按钮）
	rightClick := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseRight,
		MouseX:     100,
		MouseY:     200,
	}
	result := tracker.ProcessInput(rightClick)

	assert.False(t, result.IsDoubleClick)
	assert.Equal(t, "mouse_right_click", string(result.Action.Type))
}

// TestMouseTracker_NonMouseInput 测试非鼠标输入
func TestMouseTracker_NonMouseInput(t *testing.T) {
	tracker := NewMouseTracker()

	// 键盘输入
	input := platform.RawInput{
		Type:    platform.InputKeyPress,
		Key:     'a',
		Special: platform.KeyUnknown,
	}

	result := tracker.ProcessInput(input)

	assert.Nil(t, result)
}

// TestMouseTracker_ConsecutiveMotions 测试连续移动
func TestMouseTracker_ConsecutiveMotions(t *testing.T) {
	tracker := NewMouseTracker()

	positions := [][2]int{
		{100, 200},
		{110, 210},
		{120, 220},
		{130, 230},
	}

	for _, pos := range positions {
		input := platform.RawInput{
			Type:       platform.InputMouse,
			MouseAction: platform.MouseMotion,
			MouseX:     pos[0],
			MouseY:     pos[1],
		}
		result := tracker.ProcessInput(input)

		assert.NotNil(t, result)
		assert.Equal(t, pos[0], result.Action.Payload.(MouseEventPayload).X)
		assert.Equal(t, pos[1], result.Action.Payload.(MouseEventPayload).Y)
	}
}

// TestMouseTracker_Position 测试鼠标位置
func TestMouseTracker_Position(t *testing.T) {
	tests := []struct {
		name     string
		inputX   int
		inputY   int
		expectedX int
		expectedY int
	}{
		{"原点", 0, 0, 0, 0},
		{"正数", 100, 200, 100, 200},
		{"大坐标", 1000, 2000, 1000, 2000},
		{"零坐标", 50, 0, 50, 0},
		{"负坐标（边界）", -1, -1, -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewMouseTracker()

			input := platform.RawInput{
				Type:       platform.InputMouse,
				MouseAction: platform.MouseMotion,
				MouseX:     tt.inputX,
				MouseY:     tt.inputY,
			}

			result := tracker.ProcessInput(input)

			assert.NotNil(t, result)
			payload := result.Action.Payload.(MouseEventPayload)
			assert.Equal(t, tt.expectedX, payload.X)
			assert.Equal(t, tt.expectedY, payload.Y)
		})
	}
}

// TestMouseTracker_Throttle 测试鼠标移动节流
// 注意：MouseTracker 本身不实现节流，节流应该在调用方实现
func TestMouseTracker_Throttle(t *testing.T) {
	tracker := NewMouseTracker()

	// 模拟高频率移动
	for i := 0; i < 100; i++ {
		input := platform.RawInput{
			Type:       platform.InputMouse,
			MouseAction: platform.MouseMotion,
			MouseX:     100 + i,
			MouseY:     200,
		}
		result := tracker.ProcessInput(input)

		assert.NotNil(t, result)
		assert.Equal(t, 100+i, result.Action.Payload.(MouseEventPayload).X)
	}
}

// TestMouseTracker_ExtendedDrag 测试扩展拖拽场景
func TestMouseTracker_ExtendedDrag(t *testing.T) {
	tracker := NewMouseTracker()

	// 开始拖拽
	tracker.ProcessInput(platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	})

	// 多次移动
	positions := [][2]int{
		{110, 210},
		{120, 220},
		{130, 230},
		{140, 240},
	}

	for i, pos := range positions {
		input := platform.RawInput{
			Type:       platform.InputMouse,
			MouseAction: platform.MouseMotion,
			MouseX:     pos[0],
			MouseY:     pos[1],
		}
		result := tracker.ProcessInput(input)

		assert.True(t, result.IsDragMove)
		assert.Equal(t, 100, result.DragStartX)
		assert.Equal(t, 200, result.DragStartY)
		assert.Equal(t, 10*(i+1), result.DragDeltaX)
		assert.Equal(t, 10*(i+1), result.DragDeltaY)
	}

	// 结束拖拽
	releaseInput := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MouseRelease,
		MouseButton: platform.MouseLeft,
		MouseX:     140,
		MouseY:     240,
	}
	result := tracker.ProcessInput(releaseInput)

	assert.True(t, result.IsDragEnd)
	assert.Equal(t, 40, result.DragDeltaX)
	assert.Equal(t, 40, result.DragDeltaY)
}

// BenchmarkMouseTracker_Move 性能基准测试：鼠标移动
func BenchmarkMouseTracker_Move(b *testing.B) {
	tracker := NewMouseTracker()
	input := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MouseMotion,
		MouseX:     100,
		MouseY:     200,
	}

	for i := 0; i < b.N; i++ {
		input.MouseX = 100 + (i % 100)
		tracker.ProcessInput(input)
	}
}

// BenchmarkMouseTracker_Click 性能基准测试：鼠标点击
func BenchmarkMouseTracker_Click(b *testing.B) {
	tracker := NewMouseTracker()
	input := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}

	for i := 0; i < b.N; i++ {
		tracker.ProcessInput(input)
	}
}

// BenchmarkMouseTracker_Drag 性能基准测试：鼠标拖拽
func BenchmarkMouseTracker_Drag(b *testing.B) {
	tracker := NewMouseTracker()

	pressInput := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseButton: platform.MouseLeft,
		MouseX:     100,
		MouseY:     200,
	}

	motionInput := platform.RawInput{
		Type:       platform.InputMouse,
		MouseAction: platform.MouseMotion,
		MouseX:     150,
		MouseY:     250,
	}

	for i := 0; i < b.N; i++ {
		// 模拟拖拽循环
		tracker.ProcessInput(pressInput)
		motionInput.MouseX = 100 + (i % 100)
		tracker.ProcessInput(motionInput)
	}
}
