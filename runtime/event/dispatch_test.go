package event

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
)

// MockComponent 模拟组件
type MockComponent struct {
	id           string
	mouseHandler func(*MouseEvent, int, int) bool
}

func (m *MockComponent) ID() string {
	return m.id
}

func (m *MockComponent) HandleMouse(ev *MouseEvent, localX, intY int) bool {
	if m.mouseHandler != nil {
		return m.mouseHandler(ev, localX, intY)
	}
	return false
}

func TestLegacyHitTest(t *testing.T) {
	// 创建测试用的 LayoutBox
	boxes := []runtime.LayoutBox{
		{
			NodeID: "button1",
			X:      5,
			Y:      7,
			W:      12,
			H:      3,
		},
		{
			NodeID: "button2",
			X:      25,
			Y:      7,
			W:      12,
			H:      3,
		},
	}

	tests := []struct {
		name     string
		x, y     int
		expected string
		found    bool
	}{
		{"点击 button1 中心", 11, 8, "button1", true},
		{"点击 button2 中心", 31, 8, "button2", true},
		{"点击 button1 左上角", 5, 7, "button1", true},
		{"点击 button1 右下角", 16, 9, "button1", true},
		{"点击空白区域", 20, 5, "", false},
		{"点击空白区域2", 0, 0, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LegacyHitTest(tt.x, tt.y, boxes)
			if result.Found != tt.found {
				t.Errorf("期望找到=%v, 实际=%v", tt.found, result.Found)
			}
			if result.Found && result.NodeID != tt.expected {
				t.Errorf("期望=%s, 实际=%s", tt.expected, result.NodeID)
			}
		})
	}
}

func TestMouseEventBounds(t *testing.T) {
	// 测试边界检查
	boxes := []runtime.LayoutBox{
		{
			NodeID: "test",
			X:      10,
			Y:      10,
			W:      20,
			H:      5,
		},
	}

	tests := []struct {
		x, y    int
		inBounds bool
		localX  int
		localY  int
	}{
		{10, 10, true, 0, 0},    // 左上角
		{29, 14, true, 19, 4},   // 右下角
		{20, 12, true, 10, 2},   // 中心
		{9, 10, false, 0, 0},    // 左边界外
		{30, 12, false, 0, 0},   // 右边界外
		{20, 9, false, 0, 0},    // 上边界外
		{20, 15, false, 0, 0},   // 下边界外
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := LegacyHitTest(tt.x, tt.y, boxes)
			if result.Found != tt.inBounds {
				t.Errorf("pos=(%d,%d): 期望 inBounds=%v, 实际=%v", tt.x, tt.y, tt.inBounds, result.Found)
			}
			if result.Found && (result.X != tt.localX || result.Y != tt.localY) {
				t.Errorf("pos=(%d,%d): 期望 local=(%d,%d), 实际=(%d,%d)",
					tt.x, tt.y, tt.localX, tt.localY, result.X, result.Y)
			}
		})
	}
}

func TestDispatchMouseEvent(t *testing.T) {
	// 创建模拟组件
	clicked := false
	mockComp := &MockComponent{
		id: "test-btn",
		mouseHandler: func(ev *MouseEvent, localX, intY int) bool {
			clicked = true
			return true
		},
	}

	// 创建 LayoutNode
	node := &runtime.LayoutNode{
		ID: mockComp.id,
		Component: &runtime.ComponentRef{
			Instance: mockComp,
		},
		X:              10,
		Y:              10,
		MeasuredWidth:  20,
		MeasuredHeight: 5,
	}

	// 创建 LayoutBox
	boxes := []runtime.LayoutBox{
		{
			NodeID: mockComp.id,
			X:      10,
			Y:      10,
			W:      20,
			H:      5,
			Node:   node,
		},
	}

	// 创建鼠标事件
	mouseEv := &MouseEvent{
		X:     15, // 在盒子内
		Y:     12,
		Type:  MousePress,
		Button: MouseLeft,
	}

	ev := &EventStruct{
		TypeValue: EventMousePress,
		Mouse:     mouseEv,
	}

	// 分发事件
	result := DispatchEvent(ev, boxes)

	// 验证
	if !clicked {
		t.Error("组件的 HandleMouse 没有被调用")
	}
	if !result.Handled {
		t.Error("事件没有被标记为已处理")
	}
	if !result.Updated {
		t.Error("事件没有被标记为已更新")
	}
}
