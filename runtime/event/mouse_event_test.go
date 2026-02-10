package event

import "testing"

// TestMouseEvent_NewFields 测试 MouseEvent 的新字段
func TestMouseEvent_NewFields(t *testing.T) {
	// 创建一个鼠标按下事件
	ev := &MouseEvent{
		X:        100,
		Y:        50,
		Type:     MousePress,
		Action:   MouseActionPress,
		TargetID: "test-button",
		LocalX:   10,
		LocalY:   5,
		Button:   MouseLeft,
		Delta:    0,
		Mod:      ModNone,
	}

	// 验证基本字段
	if ev.X != 100 || ev.Y != 50 {
		t.Errorf("Coordinates incorrect, got (%d,%d), want (100,50)", ev.X, ev.Y)
	}

	// 验证新字段 - TargetID
	if ev.TargetID != "test-button" {
		t.Errorf("TargetID incorrect, got %s, want test-button", ev.TargetID)
	}

	// 验证新字段 - LocalX/LocalY
	if ev.LocalX != 10 || ev.LocalY != 5 {
		t.Errorf("Local coordinates incorrect, got (%d,%d), want (10,5)",
			ev.LocalX, ev.LocalY)
	}

	// 验证新字段 - Delta
	if ev.Delta != 0 {
		t.Errorf("Delta incorrect, got %d, want 0", ev.Delta)
	}

	// 验证 Action
	if ev.Action != MouseActionPress {
		t.Errorf("Action incorrect, got %v, want MouseActionPress", ev.Action)
	}
}

// TestMouseEvent_ScrollEvent 测试鼠标滚动事件
func TestMouseEvent_ScrollEvent(t *testing.T) {
	// 创建一个鼠标滚动向上事件
	ev := &MouseEvent{
		X:        200,
		Y:        150,
		Type:     MouseScroll,
		Action:   MouseActionWheel,
		TargetID: "scrollable-container",
		LocalX:   50,
		LocalY:   30,
		Button:   MouseNone,
		Delta:    1,  // 向上滚动
		Mod:      ModNone,
	}

	// 验证滚动事件
	if ev.Action != MouseActionWheel {
		t.Errorf("Action should be Wheel, got %v", ev.Action)
	}

	if ev.Delta != 1 {
		t.Errorf("Delta incorrect, got %d, want 1 (scroll up)", ev.Delta)
	}

	if ev.Button != MouseNone {
		t.Error("Scroll event should have no button pressed")
	}
}

// TestMouseEvent_MouseActionString 测试 MouseAction 的 String 方法
func TestMouseEvent_MouseActionString(t *testing.T) {
	tests := []struct {
		action    MouseAction
		wantStr   string
	}{
		{MouseActionPress, "Press"},
		{MouseActionRelease, "Release"},
		{MouseActionMove, "Move"},
		{MouseActionWheel, "Wheel"},
	}

	for _, tt := range tests {
		t.Run(tt.wantStr, func(t *testing.T) {
			if got := tt.action.String(); got != tt.wantStr {
				t.Errorf("String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

// TestMouseEvent_MouseClickTypeString 测试 MouseClickType 的 String 方法
func TestMouseEvent_MouseClickTypeString(t *testing.T) {
	tests := []struct {
		button   MouseClickType
		wantStr  string
	}{
		{MouseLeft, "Left"},
		{MouseMiddle, "Middle"},
		{MouseRight, "Right"},
		{MouseNone, "None"},
	}

	for _, tt := range tests {
		t.Run(tt.wantStr, func(t *testing.T) {
			if got := tt.button.String(); got != tt.wantStr {
				t.Errorf("String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

// TestMouseEvent_BackwardCompatibility 测试向后兼容性
func TestMouseEvent_BackwardCompatibility(t *testing.T) {
	// 使用旧的方式创建事件（仅设置旧字段）
	ev := &MouseEvent{
		X:      50,
		Y:      30,
		Type:   MousePress,
		Button: MouseLeft,
		Mod:    ModCtrl,
		Data:   "test data",
	}

	// 验证旧字段仍然可用
	if ev.Type != MousePress {
		t.Errorf("Type incorrect, got %v, want MousePress", ev.Type)
	}

	if ev.Button != MouseLeft {
		t.Errorf("Button incorrect, got %v, want MouseLeft", ev.Button)
	}

	if ev.Mod != ModCtrl {
		t.Errorf("Mod incorrect, got %v, want ModCtrl", ev.Mod)
	}

	if ev.Data != "test data" {
		t.Errorf("Data incorrect, got %v, want 'test data'", ev.Data)
	}

	// 新字段应该为零值
	if ev.TargetID != "" {
		t.Errorf("TargetID should be empty for old-style events, got %s", ev.TargetID)
	}

	if ev.LocalX != 0 || ev.LocalY != 0 {
		t.Errorf("Local coordinates should be 0 for old-style events, got (%d,%d)",
			ev.LocalX, ev.LocalY)
	}

	if ev.Delta != 0 {
		t.Errorf("Delta should be 0 for old-style events, got %d", ev.Delta)
	}
}

// TestMouseEvent_HitInfoComplete 测试命中信息完整性
func TestMouseEvent_HitInfoComplete(t *testing.T) {
	// 模拟 Pump 填充命中信息后的完整事件
	ev := &MouseEvent{
		X:        150,
		Y:        80,
		Type:     MousePress,
		Action:   MouseActionPress,
		TargetID: "button-1",
		LocalX:   25,
		LocalY:   15,
		Button:   MouseLeft,
		Delta:    0,
		Mod:      ModNone,
	}

	// 验证屏幕坐标
	t.Run("ScreenCoords", func(t *testing.T) {
		if ev.X != 150 || ev.Y != 80 {
			t.Errorf("Screen coords incorrect, got (%d,%d), want (150,80)",
				ev.X, ev.Y)
		}
	})

	// 验证命中信息完整
	t.Run("HitInfo", func(t *testing.T) {
		if ev.TargetID == "" {
			t.Error("TargetID should be set after hit testing")
		}

		if ev.LocalX < 0 || ev.LocalY < 0 {
			t.Errorf("Local coordinates should be non-negative, got (%d,%d)",
				ev.LocalX, ev.LocalY)
		}

		// 验证局部坐标在合理范围内（不应该大于屏幕坐标）
		if ev.LocalX > ev.X || ev.LocalY > ev.Y {
			t.Errorf("Local coords should be <= screen coords, got local(%d,%d) > screen(%d,%d)",
				ev.LocalX, ev.LocalY, ev.X, ev.Y)
		}
	})

	// 验证按钮
	t.Run("Button", func(t *testing.T) {
		if ev.Button != MouseLeft {
			t.Errorf("Button incorrect, got %v, want MouseLeft", ev.Button)
		}
	})
}

// TestMouseEvent_Modifiers 测试修饰键
func TestMouseEvent_Modifiers(t *testing.T) {
	tests := []struct {
		name    string
		mod     KeyModifier
		wantStr string
	}{
		{"None", ModNone, "None"},
		{"Shift", ModShift, "Shift"},
		{"Ctrl", ModCtrl, "Ctrl"},
		{"Alt", ModAlt, "Alt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &MouseEvent{
				Mod: tt.mod,
			}

			if ev.Mod != tt.mod {
				t.Errorf("Mod = %v, want %v", ev.Mod, tt.mod)
			}
		})
	}
}

// TestMouseEvent_ActionTypeConsistency 测试 Action 和 Type 的一致性
func TestMouseEvent_ActionTypeConsistency(t *testing.T) {
	// 新代码应该同时设置 Action 和 Type 以保持兼容性
	tests := []struct {
		name           string
		action         MouseAction
		expectedType   MouseEventType
	}{
		{"Press", MouseActionPress, MousePress},
		{"Release", MouseActionRelease, MouseRelease},
		{"Move", MouseActionMove, MouseMove},
		{"Wheel", MouseActionWheel, MouseScroll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &MouseEvent{
				Action: tt.action,
				Type:   tt.expectedType,
			}

			if ev.Action != tt.action {
				t.Errorf("Action = %v, want %v", ev.Action, tt.action)
			}

			if ev.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", ev.Type, tt.expectedType)
			}
		})
	}
}

// TestMouseEvent_DeltaScrolling 测试滚动增量
func TestMouseEvent_DeltaScrolling(t *testing.T) {
	tests := []struct {
		name        string
		delta       int
		description string
	}{
		{"ScrollUp", 1, "scrolling up (away from user)"},
		{"ScrollDown", -1, "scrolling down (toward user)"},
		{"NoScroll", 0, "no scroll"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &MouseEvent{
				Action: MouseActionWheel,
				Delta:  tt.delta,
			}

			if ev.Delta != tt.delta {
				t.Errorf("Delta = %d, want %d (%s)", ev.Delta, tt.delta, tt.description)
			}
		})
	}
}
