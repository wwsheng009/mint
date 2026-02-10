package event

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/event"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/platform"
)

// mockLayoutNode 是一个简单的 layout.Node 模拟实现
type mockLayoutNode struct {
	id       string
	nodeType string
	children []layout.Node
	x, y     int
	width    int
	height   int
}

func (m *mockLayoutNode) ID() string                 { return m.id }
func (m *mockLayoutNode) Type() string               { return m.nodeType }
func (m *mockLayoutNode) Children() []layout.Node  { return m.children }
func (m *mockLayoutNode) GetPosition() (int, int) { return m.x, m.y }
func (m *mockLayoutNode) SetPosition(x, y int)     { m.x, m.y = x, y }
func (m *mockLayoutNode) GetSize() (int, int)      { return m.width, m.height }
func (m *mockLayoutNode) SetSize(w, h int)          { m.width, m.height = w, h }
func (m *mockLayoutNode) GetWidth() int             { return m.width }
func (m *mockLayoutNode) GetHeight() int            { return m.height }

// TestPump_HitMapIntegration 测试 Pump 与 HitMap 的集成
func TestPump_HitMapIntegration(t *testing.T) {
	// 创建一个简单的布局树
	child1 := &mockLayoutNode{
		id:       "button-1",
		nodeType: "button",
		x:        10,
		y:        10,
		width:    20,
		height:   5,
		children: nil,
	}

	child2 := &mockLayoutNode{
		id:       "button-2",
		nodeType: "button",
		x:        40,
		y:        10,
		width:    15,
		height:   5,
		children: nil,
	}

	root := &mockLayoutNode{
		id:       "container",
		nodeType: "container",
		x:        0,
		y:        0,
		width:    80,
		height:   25,
		children: []layout.Node{child1, child2},
	}

	// 构建 HitMap
	hitMap := runtimeevent.BuildHitMap(root)
	if hitMap == nil {
		t.Fatal("Failed to build HitMap")
	}

	// 创建 Pump
	pump := NewPumpWithSource(NewChannelEventSource(make(chan platform.RawInput)))

	// 设置 HitMap
	pump.SetHitMap(hitMap)

	// 测试用例
	tests := []struct {
		name            string
		mouseAction     platform.MouseAction
		mouseX, mouseY  int
		expectedTarget  string
		expectedLocalX  int
		expectedLocalY  int
	}{
		{
			name:           "点击 button-1 中心",
			mouseAction:    platform.MousePress,
			mouseX:         20, // button-1 的 x=10, width=20, 中心在 20
			mouseY:         12, // button-1 的 y=10, height=5, 中心在 12
			expectedTarget: "button-1",
			expectedLocalX: 10, // 20 - 10 = 10
			expectedLocalY: 2,  // 12 - 10 = 2
		},
		{
			name:           "点击 button-2 左上角",
			mouseAction:    platform.MousePress,
			mouseX:         40, // button-2 的左边界
			mouseY:         10, // button-2 的上边界
			expectedTarget: "button-2",
			expectedLocalX: 0,  // 40 - 40 = 0
			expectedLocalY: 0,  // 10 - 10 = 0
		},
		{
			name:           "点击空白区域",
			mouseAction:    platform.MousePress,
			mouseX:         100,
			mouseY:         100,
			expectedTarget: "", // 空字符串表示未命中
			expectedLocalX: 0,
			expectedLocalY: 0,
		},
		{
			name:           "鼠标移动到 button-1",
			mouseAction:    platform.MouseMotion,
			mouseX:         15,
			mouseY:         13,
			expectedTarget: "button-1",
			expectedLocalX: 5, // 15 - 10 = 5
			expectedLocalY: 3, // 13 - 10 = 3
		},
		{
			name:           "鼠标滚轮滚动",
			mouseAction:    platform.MouseWheelUp,
			mouseX:         50, // button-2 内
			mouseY:         12,
			expectedTarget: "button-2",
			expectedLocalX: 10, // 50 - 40 = 10
			expectedLocalY: 2,  // 12 - 10 = 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建原始输入
			raw := platform.RawInput{
				Type:       platform.InputMouse,
				MouseAction: tt.mouseAction,
				MouseX:      tt.mouseX,
				MouseY:      tt.mouseY,
			}

			// 转换为消息（Phase 1: Pump now outputs Msg）
			msg := pump.convertToMsg(raw)
			if msg == nil {
				t.Fatal("convertToMsg returned nil")
			}

			// 转换Msg为Event（临时适配器，用于测试验证）
			ev := MsgToEvent(msg)
			if ev == nil {
				t.Fatal("MsgToEvent returned nil")
			}

			// 类型断言为 MouseEvent
			mouseEv, ok := ev.(*MouseEvent)
			if !ok {
				t.Fatalf("Event is not *MouseEvent, got %T", ev)
			}

			// 验证 TargetID
			if mouseEv.TargetID != tt.expectedTarget {
				t.Errorf("TargetID mismatch: got %s, want %s",
					mouseEv.TargetID, tt.expectedTarget)
			}

			// 如果预期命中了目标，验证局部坐标
			if tt.expectedTarget != "" {
				if mouseEv.LocalX != tt.expectedLocalX {
					t.Errorf("LocalX mismatch: got %d, want %d",
						mouseEv.LocalX, tt.expectedLocalX)
				}
				if mouseEv.LocalY != tt.expectedLocalY {
					t.Errorf("LocalY mismatch: got %d, want %d",
						mouseEv.LocalY, tt.expectedLocalY)
				}
			}

			// 验证 Action 字段
			var expectedAction event.MouseAction
			switch tt.mouseAction {
			case platform.MousePress:
				expectedAction = event.MouseActionPress
			case platform.MouseRelease:
				expectedAction = event.MouseActionRelease
			case platform.MouseMotion:
				expectedAction = event.MouseActionMove
			case platform.MouseWheelUp, platform.MouseWheelDown:
				expectedAction = event.MouseActionWheel
			}

			if mouseEv.Action != expectedAction {
				t.Errorf("Action mismatch: got %v, want %v",
					mouseEv.Action, expectedAction)
			}
		})
	}
}

// TestPump_WheelDelta 测试鼠标滚轮增量
func TestPump_WheelDelta(t *testing.T) {
	pump := NewPumpWithSource(NewChannelEventSource(make(chan platform.RawInput)))

	tests := []struct {
		name        string
		mouseAction platform.MouseAction
		expectedDelta int
	}{
		{"WheelUp", platform.MouseWheelUp, 1},
		{"WheelDown", platform.MouseWheelDown, -1},
		{"Press", platform.MousePress, 0},
		{"Release", platform.MouseRelease, 0},
		{"Motion", platform.MouseMotion, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := platform.RawInput{
				Type:        platform.InputMouse,
				MouseAction: tt.mouseAction,
				MouseX:      10,
				MouseY:      10,
			}

			msg := pump.convertToMsg(raw)
			if msg == nil {
				t.Fatal("convertToMsg returned nil")
			}

			ev := MsgToEvent(msg)
			if ev == nil {
				t.Fatal("MsgToEvent returned nil")
			}

			mouseEv, ok := ev.(*MouseEvent)
			if !ok {
				t.Fatalf("Event is not *MouseEvent, got %T", ev)
			}

			if mouseEv.Delta != tt.expectedDelta {
				t.Errorf("Delta mismatch: got %d, want %d",
					mouseEv.Delta, tt.expectedDelta)
			}
		})
	}
}

// TestPump_NilHitMap 测试没有 HitMap 时的行为
func TestPump_NilHitMap(t *testing.T) {
	pump := NewPumpWithSource(NewChannelEventSource(make(chan platform.RawInput)))

	// 不设置 HitMap（保持为 nil）

	raw := platform.RawInput{
		Type:        platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseX:      10,
		MouseY:      10,
	}

	msg := pump.convertToMsg(raw)
	if msg == nil {
		t.Fatal("convertToMsg returned nil")
	}

	ev := MsgToEvent(msg)
	if ev == nil {
		t.Fatal("MsgToEvent returned nil")
	}

	mouseEv, ok := ev.(*MouseEvent)
	if !ok {
		t.Fatalf("Event is not *MouseEvent, got %T", ev)
	}

	// HitMap 为 nil 时，TargetID 应该为空
	if mouseEv.TargetID != "" {
		t.Errorf("TargetID should be empty when HitMap is nil, got %s",
			mouseEv.TargetID)
	}

	// 局部坐标应该为 0
	if mouseEv.LocalX != 0 || mouseEv.LocalY != 0 {
		t.Errorf("Local coordinates should be 0 when HitMap is nil, got (%d,%d)",
			mouseEv.LocalX, mouseEv.LocalY)
	}
}

// TestPump_ConcurrentHitMapAccess 测试并发访问 HitMap
func TestPump_ConcurrentHitMapAccess(t *testing.T) {
	// 创建简单的 HitMap
	node := &mockLayoutNode{
		id:       "test",
		nodeType: "container",
		x:        0,
		y:        0,
		width:    100,
		height:   100,
		children: nil,
	}

	hitMap := runtimeevent.BuildHitMap(node)

	pump := NewPumpWithSource(NewChannelEventSource(make(chan platform.RawInput)))

	// 并发设置 HitMap 和转换事件
	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			pump.SetHitMap(hitMap)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			raw := platform.RawInput{
				Type:        platform.InputMouse,
				MouseAction: platform.MousePress,
				MouseX:      10,
				MouseY:      10,
			}
			pump.convertToMsg(raw)
		}
		done <- true
	}()

	<-done
	<-done

	// 验证最终状态
	raw := platform.RawInput{
		Type:        platform.InputMouse,
		MouseAction: platform.MousePress,
		MouseX:      10,
		MouseY:      10,
	}
	msg := pump.convertToMsg(raw)
	if msg == nil {
		t.Fatal("convertToMsg returned nil")
	}

	ev := MsgToEvent(msg)
	if ev == nil {
		t.Fatal("MsgToEvent returned nil")
	}

	mouseEv, ok := ev.(*MouseEvent)
	if !ok {
		t.Fatalf("Event is not *MouseEvent, got %T", ev)
	}

	// 应该成功命中
	if mouseEv.TargetID != "test" {
		t.Errorf("TargetID mismatch: got %s, want test", mouseEv.TargetID)
	}
}
