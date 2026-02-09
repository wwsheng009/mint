package inspector

import (
	"testing"
)

// TestDigitKeyResponse 测试数字键是否正确响应
func TestDigitKeyResponse(t *testing.T) {
	inspector := NewStandaloneInspector()
	inspector.Enable()

	// 打开 inspector
	inspector.ToggleVisibility()

	// 测试每个数字键
	digitKeys := []struct {
		key        string
		wantTab    InspectorTab
	}{
		{"1", TabElements},
		{"2", TabConsole},
		{"3", TabPerformance},
		{"4", TabDiagnostics},
		{"5", TabLayout},
		{"6", TabNetwork},
	}

	for _, tt := range digitKeys {
		t.Run(tt.key, func(t *testing.T) {
			// 记录切换前的 tab
			beforeTab := inspector.activeTab

			// 按键
			handled := inspector.HandleKeyEvent(tt.key, false, false, false)

			// 验证按键被处理
			if !handled {
				t.Errorf("HandleKeyEvent(%s) returned false, expected true", tt.key)
			}

			// 验证 tab 切换了
			if inspector.activeTab != tt.wantTab {
				t.Errorf("After key %s: activeTab = %v, want %v (before: %v)",
					tt.key, inspector.activeTab, tt.wantTab, beforeTab)
			}

			// 验证 RenderContent 不为 nil
			content := inspector.RenderContent()
			if content == nil {
				t.Errorf("RenderContent() returned nil after key %s", tt.key)
			}
		})
	}
}

// TestLayoutTabKey5 详细测试数字键 5
func TestLayoutTabKey5(t *testing.T) {
	inspector := NewStandaloneInspector()
	inspector.Enable()
	inspector.ToggleVisibility()

	// 初始状态应该是 Elements tab
	if inspector.activeTab != TabElements {
		t.Errorf("Initial activeTab = %v, want TabElements", inspector.activeTab)
	}

	// 按键 5
	handled := inspector.HandleKeyEvent("5", false, false, false)

	if !handled {
		t.Fatal("HandleKeyEvent('5') returned false")
	}

	// 验证切换到 Layout tab
	if inspector.activeTab != TabLayout {
		t.Errorf("After key '5': activeTab = %v, want TabLayout", inspector.activeTab)
	}

	// 获取 Layout tab 的内容
	content := inspector.buildLayoutTabContent()

	if content == nil {
		t.Fatal("buildLayoutTabContent() returned nil")
	}

	t.Logf("✅ Key '5' successfully switched to Layout tab")
	t.Logf("✅ Layout tab content is not nil")
	t.Logf("✅ activeTab = %v", inspector.activeTab)
}

// TestTabSwitchingSequence 测试连续切换 tab
func TestTabSwitchingSequence(t *testing.T) {
	inspector := NewStandaloneInspector()
	inspector.Enable()
	inspector.ToggleVisibility()

	// 测试序列：1 -> 5 -> 2 -> 5
	sequence := []struct {
		key     string
		wantTab InspectorTab
	}{
		{"1", TabElements},
		{"5", TabLayout},
		{"2", TabConsole},
		{"5", TabLayout},
	}

	for _, tt := range sequence {
		t.Run(tt.key, func(t *testing.T) {
			handled := inspector.HandleKeyEvent(tt.key, false, false, false)

			if !handled {
				t.Errorf("HandleKeyEvent(%s) returned false", tt.key)
			}

			if inspector.activeTab != tt.wantTab {
				t.Errorf("activeTab = %v, want %v", inspector.activeTab, tt.wantTab)
			}

			// 每次 tab 切换后，内容都应该能渲染
			content := inspector.RenderContent()
			if content == nil {
				t.Errorf("RenderContent() returned nil after switching to %v", tt.wantTab)
			}
		})
	}

	t.Logf("✅ Tab switching sequence works correctly")
}
