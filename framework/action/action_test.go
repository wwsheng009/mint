package action

import (
	"testing"
)

// TestAction_TypeConstants 测试所有 Action 类型常量
func TestAction_TypeConstants(t *testing.T) {
	tests := []struct {
		name  string
		atype ActionType
	}{
		// 导航类
		{"NavigateNext", ActionNavigateNext},
		{"NavigatePrev", ActionNavigatePrev},
		{"NavigateUp", ActionNavigateUp},
		{"NavigateDown", ActionNavigateDown},
		{"NavigateLeft", ActionNavigateLeft},
		{"NavigateRight", ActionNavigateRight},
		{"NavigatePageUp", ActionNavigatePageUp},
		{"NavigatePageDown", ActionNavigatePageDown},
		{"NavigateHome", ActionNavigateHome},
		{"NavigateEnd", ActionNavigateEnd},

		// 选择类
		{"Select", ActionSelect},
		{"Toggle", ActionToggle},
		{"Expand", ActionExpand},
		{"Collapse", ActionCollapse},

		// 编辑类
		{"InputText", ActionInputText},
		{"DeleteChar", ActionDeleteChar},
		{"DeleteWord", ActionDeleteWord},
		{"DeleteLine", ActionDeleteLine},
		{"Backspace", ActionBackspace},
		{"Enter", ActionEnter},

		// 表单类
		{"Submit", ActionSubmit},
		{"Cancel", ActionCancel},
		{"Validate", ActionValidate},
		{"Reset", ActionReset},

		// 系统类
		{"Quit", ActionQuit},
		{"Focus", ActionFocus},
		{"Blur", ActionBlur},
		{"Inspect", ActionInspect},
		{"Refresh", ActionRefresh},

		// 鼠标类
		{"Click", ActionClick},
		{"DoubleClick", ActionDoubleClick},
		{"RightClick", ActionRightClick},
		{"MiddleClick", ActionMiddleClick},
		{"Scroll", ActionScroll},
		{"Hover", ActionHover},
		{"DragStart", ActionDragStart},
		{"DragMove", ActionDragMove},
		{"DragEnd", ActionDragEnd},

		// 剪贴板
		{"Copy", ActionCopy},
		{"Cut", ActionCut},
		{"Paste", ActionPaste},

		// 搜索类
		{"Search", ActionSearch},
		{"SearchNext", ActionSearchNext},
		{"SearchPrev", ActionSearchPrev},
		{"Replace", ActionReplace},
		{"ReplaceAll", ActionReplaceAll},

		// 视图类
		{"ZoomIn", ActionZoomIn},
		{"ZoomOut", ActionZoomOut},
		{"ZoomReset", ActionZoomReset},
		{"SplitView", ActionSplitView},
		{"CloseView", ActionCloseView},
		{"Maximize", ActionMaximize},
		{"Minimize", ActionMinimize},
		{"Fullscreen", ActionFullscreen},

		// 自定义
		{"Custom", ActionCustom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.atype == "" {
				t.Errorf("ActionType %s should not be empty", tt.name)
			}
		})
	}
}

// TestAction_IsNavigation 测试导航 Action 判断
func TestAction_IsNavigation(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected bool
	}{
		{"NavigateNext", &Action{Type: ActionNavigateNext}, true},
		{"NavigatePrev", &Action{Type: ActionNavigatePrev}, true},
		{"NavigateUp", &Action{Type: ActionNavigateUp}, true},
		{"NavigateDown", &Action{Type: ActionNavigateDown}, true},
		{"NavigateLeft", &Action{Type: ActionNavigateLeft}, true},
		{"NavigateRight", &Action{Type: ActionNavigateRight}, true},
		{"NavigatePageUp", &Action{Type: ActionNavigatePageUp}, true},
		{"NavigatePageDown", &Action{Type: ActionNavigatePageDown}, true},
		{"NavigateHome", &Action{Type: ActionNavigateHome}, true},
		{"NavigateEnd", &Action{Type: ActionNavigateEnd}, true},
		{"Click", &Action{Type: ActionClick}, false},
		{"InputText", &Action{Type: ActionInputText}, false},
		{"Quit", &Action{Type: ActionQuit}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.IsNavigation(); got != tt.expected {
				t.Errorf("IsNavigation() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAction_IsEditing 测试编辑 Action 判断
func TestAction_IsEditing(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected bool
	}{
		{"InputText", &Action{Type: ActionInputText}, true},
		{"DeleteChar", &Action{Type: ActionDeleteChar}, true},
		{"DeleteWord", &Action{Type: ActionDeleteWord}, true},
		{"DeleteLine", &Action{Type: ActionDeleteLine}, true},
		{"Backspace", &Action{Type: ActionBackspace}, true},
		{"Enter", &Action{Type: ActionEnter}, true},
		{"Click", &Action{Type: ActionClick}, false},
		{"NavigateDown", &Action{Type: ActionNavigateDown}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.IsEditing(); got != tt.expected {
				t.Errorf("IsEditing() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAction_IsSelection 测试选择类 Action 判断
func TestAction_IsSelection(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected bool
	}{
		{"Select", &Action{Type: ActionSelect}, true},
		{"Toggle", &Action{Type: ActionToggle}, true},
		{"Expand", &Action{Type: ActionExpand}, true},
		{"Collapse", &Action{Type: ActionCollapse}, true},
		{"Click", &Action{Type: ActionClick}, false},
		{"InputText", &Action{Type: ActionInputText}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.IsSelection(); got != tt.expected {
				t.Errorf("IsSelection() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAction_IsForm 测试表单类 Action 判断
func TestAction_IsForm(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected bool
	}{
		{"Submit", &Action{Type: ActionSubmit}, true},
		{"Cancel", &Action{Type: ActionCancel}, true},
		{"Validate", &Action{Type: ActionValidate}, true},
		{"Reset", &Action{Type: ActionReset}, true},
		{"Click", &Action{Type: ActionClick}, false},
		{"InputText", &Action{Type: ActionInputText}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.IsForm(); got != tt.expected {
				t.Errorf("IsForm() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAction_IsSystem 测试系统类 Action 判断
func TestAction_IsSystem(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected bool
	}{
		{"Quit", &Action{Type: ActionQuit}, true},
		{"Focus", &Action{Type: ActionFocus}, true},
		{"Blur", &Action{Type: ActionBlur}, true},
		{"Inspect", &Action{Type: ActionInspect}, true},
		{"Refresh", &Action{Type: ActionRefresh}, true},
		{"Click", &Action{Type: ActionClick}, false},
		{"InputText", &Action{Type: ActionInputText}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.IsSystem(); got != tt.expected {
				t.Errorf("IsSystem() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAction_IsMouse 测试鼠标 Action 判断
func TestAction_IsMouse(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected bool
	}{
		{"Click", &Action{Type: ActionClick}, true},
		{"DoubleClick", &Action{Type: ActionDoubleClick}, true},
		{"RightClick", &Action{Type: ActionRightClick}, true},
		{"MiddleClick", &Action{Type: ActionMiddleClick}, true},
		{"Scroll", &Action{Type: ActionScroll}, true},
		{"Hover", &Action{Type: ActionHover}, true},
		{"DragStart", &Action{Type: ActionDragStart}, true},
		{"DragMove", &Action{Type: ActionDragMove}, true},
		{"DragEnd", &Action{Type: ActionDragEnd}, true},
		{"NavigateDown", &Action{Type: ActionNavigateDown}, false},
		{"InputText", &Action{Type: ActionInputText}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.IsMouse(); got != tt.expected {
				t.Errorf("IsMouse() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAction_IsClipboard 测试剪贴板 Action 判断
func TestAction_IsClipboard(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected bool
	}{
		{"Copy", &Action{Type: ActionCopy}, true},
		{"Cut", &Action{Type: ActionCut}, true},
		{"Paste", &Action{Type: ActionPaste}, true},
		{"Click", &Action{Type: ActionClick}, false},
		{"InputText", &Action{Type: ActionInputText}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.IsClipboard(); got != tt.expected {
				t.Errorf("IsClipboard() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAction_IsSearch 测试搜索类 Action 判断
func TestAction_IsSearch(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected bool
	}{
		{"Search", &Action{Type: ActionSearch}, true},
		{"SearchNext", &Action{Type: ActionSearchNext}, true},
		{"SearchPrev", &Action{Type: ActionSearchPrev}, true},
		{"Replace", &Action{Type: ActionReplace}, true},
		{"ReplaceAll", &Action{Type: ActionReplaceAll}, true},
		{"Click", &Action{Type: ActionClick}, false},
		{"InputText", &Action{Type: ActionInputText}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.IsSearch(); got != tt.expected {
				t.Errorf("IsSearch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAction_IsView 测试视图类 Action 判断
func TestAction_IsView(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected bool
	}{
		{"ZoomIn", &Action{Type: ActionZoomIn}, true},
		{"ZoomOut", &Action{Type: ActionZoomOut}, true},
		{"ZoomReset", &Action{Type: ActionZoomReset}, true},
		{"SplitView", &Action{Type: ActionSplitView}, true},
		{"CloseView", &Action{Type: ActionCloseView}, true},
		{"Maximize", &Action{Type: ActionMaximize}, true},
		{"Minimize", &Action{Type: ActionMinimize}, true},
		{"Fullscreen", &Action{Type: ActionFullscreen}, true},
		{"Click", &Action{Type: ActionClick}, false},
		{"InputText", &Action{Type: ActionInputText}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.IsView(); got != tt.expected {
				t.Errorf("IsView() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAction_RequiresTarget 测试是否需要目标
func TestAction_RequiresTarget(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected bool
	}{
		{
			name:     "Click with target",
			action:   &Action{Type: ActionClick, TargetID: 12345},
			expected: true,
		},
		{
			name:     "Click without target",
			action:   &Action{Type: ActionClick, TargetID: 0},
			expected: false,
		},
		{
			name:     "Scroll with target",
			action:   &Action{Type: ActionScroll, TargetID: 67890},
			expected: true,
		},
		{
			name:     "NavigateDown (keyboard)",
			action:   &Action{Type: ActionNavigateDown},
			expected: false,
		},
		{
			name:     "InputText (keyboard)",
			action:   &Action{Type: ActionInputText},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.RequiresTarget(); got != tt.expected {
				t.Errorf("RequiresTarget() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAction_GetPayloadString 测试获取字符串 Payload
func TestAction_GetPayloadString(t *testing.T) {
	tests := []struct {
		name          string
		action        *Action
		expectedStr   string
		expectedOk    bool
	}{
		{
			name:        "String payload",
			action:      &Action{Type: ActionInputText, Payload: "hello"},
			expectedStr: "hello",
			expectedOk:  true,
		},
		{
			name:        "Int payload",
			action:      &Action{Type: ActionScroll, Payload: 1},
			expectedStr: "",
			expectedOk:  false,
		},
		{
			name:        "Nil payload",
			action:      &Action{Type: ActionNavigateDown},
			expectedStr: "",
			expectedOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, ok := tt.action.GetPayloadString()
			if ok != tt.expectedOk {
				t.Errorf("GetPayloadString() ok = %v, want %v", ok, tt.expectedOk)
			}
			if str != tt.expectedStr {
				t.Errorf("GetPayloadString() = %v, want %v", str, tt.expectedStr)
			}
		})
	}
}

// TestAction_GetPayloadInt 测试获取整数 Payload
func TestAction_GetPayloadInt(t *testing.T) {
	tests := []struct {
		name       string
		action     *Action
		expectedInt int
		expectedOk  bool
	}{
		{
			name:       "Int payload",
			action:     &Action{Type: ActionScroll, Payload: 1},
			expectedInt: 1,
			expectedOk:  true,
		},
		{
			name:       "Negative int",
			action:     &Action{Type: ActionScroll, Payload: -1},
			expectedInt: -1,
			expectedOk:  true,
		},
		{
			name:       "String payload",
			action:     &Action{Type: ActionInputText, Payload: "hello"},
			expectedInt: 0,
			expectedOk:  false,
		},
		{
			name:       "Nil payload",
			action:     &Action{Type: ActionNavigateDown},
			expectedInt: 0,
			expectedOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, ok := tt.action.GetPayloadInt()
			if ok != tt.expectedOk {
				t.Errorf("GetPayloadInt() ok = %v, want %v", ok, tt.expectedOk)
			}
			if i != tt.expectedInt {
				t.Errorf("GetPayloadInt() = %v, want %v", i, tt.expectedInt)
			}
		})
	}
}

// TestAction_GetPayloadPoint 测试获取点 Payload
func TestAction_GetPayloadPoint(t *testing.T) {
	tests := []struct {
		name          string
		action        *Action
		expectedX     int
		expectedY     int
		expectedOk    bool
	}{
		{
			name: "Struct payload",
			action: &Action{
				Type:    ActionClick,
				Payload: struct{ X, Y int }{X: 10, Y: 20},
			},
			expectedX:  10,
			expectedY:  20,
			expectedOk: true,
		},
		{
			name: "Map payload",
			action: &Action{
				Type:    ActionHover,
				Payload: map[string]int{"x": 5, "y": 15},
			},
			expectedX:  5,
			expectedY:  15,
			expectedOk: true,
		},
		{
			name: "Incomplete map payload",
			action: &Action{
				Type:    ActionHover,
				Payload: map[string]int{"x": 5},
			},
			expectedX:  0,
			expectedY:  0,
			expectedOk: false,
		},
		{
			name:       "String payload",
			action:     &Action{Type: ActionInputText, Payload: "hello"},
			expectedX:  0,
			expectedY:  0,
			expectedOk: false,
		},
		{
			name:       "Nil payload",
			action:     &Action{Type: ActionNavigateDown},
			expectedX:  0,
			expectedY:  0,
			expectedOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y, ok := tt.action.GetPayloadPoint()
			if ok != tt.expectedOk {
				t.Errorf("GetPayloadPoint() ok = %v, want %v", ok, tt.expectedOk)
			}
			if x != tt.expectedX || y != tt.expectedY {
				t.Errorf("GetPayloadPoint() = (%d,%d), want (%d,%d)", x, y, tt.expectedX, tt.expectedY)
			}
		})
	}
}

// TestAction_String 测试 String() 方法
func TestAction_String(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected string
	}{
		{
			name:     "Simple action",
			action:   &Action{Type: ActionNavigateDown},
			expected: "navigate_down",
		},
		{
			name:     "Action with target",
			action:   &Action{Type: ActionClick, TargetID: 12345},
			expected: "click@12345",
		},
		{
			name:     "Action with payload",
			action:   &Action{Type: ActionInputText, Payload: "hello"},
			expected: "input_text(hello)",
		},
		{
			name:     "Action with target and payload",
			action:   &Action{Type: ActionScroll, TargetID: 67890, Payload: 1},
			expected: "scroll@67890(1)",
		},
		{
			name:     "Action with source",
			action:   &Action{Type: ActionNavigateDown, Source: "keyboard"},
			expected: "navigate_down [keyboard]",
		},
		{
			name:     "Complete action",
			action:   &Action{Type: ActionClick, TargetID: 99999, Payload: struct{ X, Y int }{10, 20}, Source: "mouse"},
			expected: "click@99999({10 20}) [mouse]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.action.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestAction_NewAction 测试 Action 构造函数
func TestAction_NewAction(t *testing.T) {
	t.Run("NewAction", func(t *testing.T) {
		act := NewAction(ActionNavigateDown)
		if act.Type != ActionNavigateDown {
			t.Errorf("Type = %v, want %v", act.Type, ActionNavigateDown)
		}
		if act.Payload != nil {
			t.Errorf("Payload should be nil, got %v", act.Payload)
		}
		if act.Source != "" {
			t.Errorf("Source should be empty, got %q", act.Source)
		}
		if act.TargetID != 0 {
			t.Errorf("TargetID should be 0, got %d", act.TargetID)
		}
	})

	t.Run("NewActionWithPayload", func(t *testing.T) {
		act := NewActionWithPayload(ActionInputText, "hello")
		if act.Type != ActionInputText {
			t.Errorf("Type = %v, want %v", act.Type, ActionInputText)
		}
		if act.Payload != "hello" {
			t.Errorf("Payload = %v, want %v", act.Payload, "hello")
		}
	})

	t.Run("NewActionFromMouse", func(t *testing.T) {
		act := NewActionFromMouse(ActionClick, 12345, 10, 20)
		if act.Type != ActionClick {
			t.Errorf("Type = %v, want %v", act.Type, ActionClick)
		}
		if act.TargetID != 12345 {
			t.Errorf("TargetID = %d, want %d", act.TargetID, 12345)
		}
		if act.Source != "mouse" {
			t.Errorf("Source = %q, want %q", act.Source, "mouse")
		}
		x, y, ok := act.GetPayloadPoint()
		if !ok {
			t.Fatal("GetPayloadPoint() should return ok=true")
		}
		if x != 10 || y != 20 {
			t.Errorf("Payload = (%d,%d), want (10,20)", x, y)
		}
	})

	t.Run("NewActionFromKey", func(t *testing.T) {
		act := NewActionFromKey(ActionNavigateDown, "keyboard")
		if act.Type != ActionNavigateDown {
			t.Errorf("Type = %v, want %v", act.Type, ActionNavigateDown)
		}
		if act.Source != "keyboard" {
			t.Errorf("Source = %q, want %q", act.Source, "keyboard")
		}
	})
}

// TestAction_Clone 测试 Clone 方法
func TestAction_Clone(t *testing.T) {
	original := &Action{
		Type:     ActionClick,
		Payload:  struct{ X, Y int }{X: 10, Y: 20},
		Source:   "mouse",
		TargetID: 12345,
	}

	cloned := original.Clone()

	// 验证所有字段都相同
	if cloned.Type != original.Type {
		t.Errorf("Cloned Type = %v, want %v", cloned.Type, original.Type)
	}
	if cloned.TargetID != original.TargetID {
		t.Errorf("Cloned TargetID = %d, want %d", cloned.TargetID, original.TargetID)
	}
	if cloned.Source != original.Source {
		t.Errorf("Cloned Source = %q, want %q", cloned.Source, original.Source)
	}

	// 修改原始对象，确保克隆是深拷贝
	original.Type = ActionNavigateDown
	original.TargetID = 67890

	if cloned.Type != ActionClick {
		t.Errorf("Cloned Type should not be affected by original modification")
	}
	if cloned.TargetID != 12345 {
		t.Errorf("Cloned TargetID should not be affected by original modification")
	}
}

// TestAction_WithModifiers 测试链式修改方法
func TestAction_WithModifiers(t *testing.T) {
	base := NewAction(ActionClick)

	t.Run("WithTarget", func(t *testing.T) {
		modified := base.WithTarget(12345)
		if modified.TargetID != 12345 {
			t.Errorf("TargetID = %d, want %d", modified.TargetID, 12345)
		}
		// 原对象不应该被修改
		if base.TargetID != 0 {
			t.Errorf("Original TargetID should remain 0")
		}
	})

	t.Run("WithPayload", func(t *testing.T) {
		modified := base.WithPayload("test payload")
		if modified.Payload != "test payload" {
			t.Errorf("Payload = %v, want %v", modified.Payload, "test payload")
		}
		// 原对象不应该被修改
		if base.Payload != nil {
			t.Errorf("Original Payload should remain nil")
		}
	})

	t.Run("WithSource", func(t *testing.T) {
		modified := base.WithSource("mouse")
		if modified.Source != "mouse" {
			t.Errorf("Source = %q, want %q", modified.Source, "mouse")
		}
		// 原对象不应该被修改
		if base.Source != "" {
			t.Errorf("Original Source should remain empty")
		}
	})

	t.Run("Chained", func(t *testing.T) {
		modified := base.
			WithTarget(54321).
			WithPayload(struct{ X, Y int }{10, 20}).
			WithSource("mouse")

		if modified.TargetID != 54321 {
			t.Errorf("TargetID = %d, want %d", modified.TargetID, 54321)
		}
		if modified.Source != "mouse" {
			t.Errorf("Source = %q, want %q", modified.Source, "mouse")
		}
	})
}
