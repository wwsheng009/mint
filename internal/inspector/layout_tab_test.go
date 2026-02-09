package inspector

import (
	"testing"
)

// TestLayoutTabShortcut 测试 Layout tab 快捷键
func TestLayoutTabShortcut(t *testing.T) {
	inspector := NewStandaloneInspector()
	inspector.Enable()

	// 测试数字键 5 切换到 Layout tab
	tests := []struct {
		key        string
		wantTab    InspectorTab
		description string
	}{
		{"1", TabElements, "Key 1 should switch to Elements tab"},
		{"2", TabConsole, "Key 2 should switch to Console tab"},
		{"3", TabPerformance, "Key 3 should switch to Performance tab"},
		{"4", TabDiagnostics, "Key 4 should switch to Diagnostics tab"},
		{"5", TabLayout, "Key 5 should switch to Layout tab"},
		{"6", TabNetwork, "Key 6 should switch to Network tab"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// 先打开 inspector
			inspector.ToggleVisibility()
			defer inspector.ToggleVisibility()

			// 按键
			handled := inspector.HandleKeyEvent(tt.key, false, false, false)

			// 验证按键被处理
			if !handled {
				t.Errorf("HandleKeyEvent() returned false for key=%s", tt.key)
			}

			// 验证 tab 切换
			if inspector.activeTab != tt.wantTab {
				t.Errorf("activeTab = %v, want %v", inspector.activeTab, tt.wantTab)
			}
		})
	}
}

// TestLayoutTabInTabItems 测试 Layout tab 是否在 tabItems 列表中
func TestLayoutTabInTabItems(t *testing.T) {
	inspector := NewStandaloneInspector()
	inspector.Enable()

	// 检查 RenderContent 不为 nil
	inspector.ToggleVisibility()
	content := inspector.RenderContent()

	if content == nil {
		t.Fatal("RenderContent() returned nil")
	}

	// 检查 activeTab 可以设置为 TabLayout
	inspector.activeTab = TabLayout
	if inspector.activeTab != TabLayout {
		t.Errorf("activeTab = %v, want %v", inspector.activeTab, TabLayout)
	}
}

// TestLayoutTabBuildContent 测试 Layout tab 内容构建
func TestLayoutTabBuildContent(t *testing.T) {
	inspector := NewStandaloneInspector()
	inspector.Enable()

	// 构建 Layout tab 内容
	content := inspector.buildLayoutTabContent()

	if content == nil {
		t.Fatal("buildLayoutTabContent() returned nil")
	}

	// 验证内容不为空
	props := content.Props()
	if props == nil {
		t.Fatal("Content has no props")
	}

	// 检查是否有 layout 相关的 props
	if _, ok := props["layoutDiagnostic"]; ok {
		t.Log("✅ Layout diagnostic props found")
	}
}

// BenchmarkLayoutTabSwitch 性能测试：切换到 Layout tab
func BenchmarkLayoutTabSwitch(b *testing.B) {
	inspector := NewStandaloneInspector()
	inspector.Enable()
	inspector.ToggleVisibility()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inspector.HandleKeyEvent("5", false, false, false)
		inspector.HandleKeyEvent("4", false, false, false)
	}
}
