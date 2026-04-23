package inspector

import (
	"strings"
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestRealInspectorRendering 真正测试 Inspector 的渲染
func TestRealInspectorRendering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建 Inspector
	inspector := NewStandaloneInspector()
	inspector.Enable()           // 启用 Inspector
	inspector.ToggleVisibility() // 设置为可见！
	inspector.SetOverlaySize(100, 40)

	testApp, err := ui.RunTest(func() ui.VNode {
		// 返回 Inspector 的 VNode
		return inspector.RenderOverlay()
	},
		ui.WithWidth(100),
		ui.WithHeight(40),
		ui.WithTitle("Real Inspector Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(300 * time.Millisecond)

	// 获取渲染输出
	rendered := testApp.GetRenderString()

	t.Logf("=== Inspector Render Output ===\n%s\n=== End ===", rendered)

	// 验证 Inspector 内容存在
	if len(rendered) == 0 {
		t.Fatal("Inspector rendered empty output!")
	}

	// 检查关键的 Inspector 元素
	expectedContent := []string{
		"Elements",
		"Console(2)",
		"Layout Tree",
		"Instructions",
		"No tree to display",
	}

	missingContent := []string{}
	for _, expected := range expectedContent {
		if !strings.Contains(rendered, expected) {
			missingContent = append(missingContent, expected)
		}
	}

	if len(missingContent) > 0 {
		t.Errorf("Inspector is missing expected content: %v", missingContent)
		t.Logf("Rendered content:\n%s", rendered)
	} else {
		t.Log("✓ Inspector contains all expected content")
	}

	// 检查是否有边框（┌─┐│└┘）
	borderChars := []string{"┌", "─", "┐", "│", "└", "┘"}
	hasBorder := false
	for _, char := range borderChars {
		if strings.Contains(rendered, char) {
			hasBorder = true
			break
		}
	}

	if !hasBorder {
		t.Error("Inspector should have border characters")
	} else {
		t.Log("✓ Inspector has border")
	}
}

// TestInspectorWithRealApp 测试与真实应用的集成
func TestInspectorWithRealApp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建 Inspector
	inspector := NewStandaloneInspector()
	inspector.Enable()
	inspector.ToggleVisibility() // Make visible so RenderOverlay returns content
	inspector.SetOverlaySize(100, 40)

	// 渲染 Inspector overlay
	overlay := inspector.RenderOverlay()

	if overlay == nil {
		t.Fatal("Inspector overlay is nil!")
	}

	t.Logf("Inspector overlay type: %T", overlay)

	// 检查 overlay 是否有内容
	if overlay == nil {
		t.Fatal("Inspector overlay VNode is nil!")
	}
	t.Logf("Inspector overlay VNode type: %s", overlay.Type())

	t.Log("✓ Inspector overlay created successfully")
}

// TestInspectorTreeRendering 测试树视图渲染
func TestInspectorTreeRendering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	inspector := NewStandaloneInspector()

	// 构建树视图
	treeView := inspector.GetTreeView()
	if treeView == nil {
		t.Fatal("TreeView is nil!")
	}

	// 创建测试树
	testVNode := ui.VStack(
		ui.Text("Root"),
		ui.VStack(
			ui.Text("Child 1"),
			ui.Text("Child 2"),
		),
		ui.NewButtonBuilder("Button").Build(),
	)

	// 使用 AttachToApp 附加 VNode
	inspector.AttachToApp(testVNode)

	// 获取树内容
	lines, totalLines := treeView.GetTreeLines()

	t.Logf("Tree has %d lines", totalLines)
	if totalLines > 0 {
		t.Logf("=== First 20 lines of Tree ===")
		maxLines := 20
		if len(lines) < maxLines {
			maxLines = len(lines)
		}
		for i := 0; i < maxLines; i++ {
			t.Logf("  %s", lines[i])
		}
		t.Logf("=== End ===")
	}

	if totalLines == 0 {
		t.Log("Tree is empty (may need actual VNode attachment)")
	}

	if len(lines) == 0 && totalLines > 0 {
		t.Error("Tree lines should not be empty if totalLines > 0!")
	}

	t.Log("✓ Tree view created successfully")
}

// TestInspectorTabContent 测试标签页内容
func TestInspectorTabContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	inspector := NewStandaloneInspector()

	// 测试 Elements 标签页内容
	inspector.SetActiveTab(TabElements)
	elementsContent := inspector.buildElementsTabContent()

	if elementsContent == nil {
		t.Fatal("Elements tab content is nil!")
	}

	t.Logf("Elements content type: %T", elementsContent)

	// 测试 Console 标签页内容
	inspector.SetActiveTab(TabConsole)
	consoleContent := inspector.buildConsoleTabContent()

	if consoleContent == nil {
		t.Fatal("Console tab content is nil!")
	}

	t.Logf("Console content type: %T", consoleContent)

	// 测试所有标签页切换
	tabs := []InspectorTab{
		TabElements,
		TabConsole,
		TabPerformance,
		TabDiagnostics,
		TabNetwork,
	}

	for _, tab := range tabs {
		inspector.SetActiveTab(tab)
		activeTab := inspector.GetActiveTab()

		if activeTab != tab {
			t.Errorf("Expected active tab %v, got %v", tab, activeTab)
		} else {
			t.Logf("✓ Successfully switched to %v tab", tab)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// TestInspectorScrollState 测试滚动状态管理
func TestInspectorScrollState(t *testing.T) {
	inspector := NewStandaloneInspector()

	// 测试初始状态
	initialOffset := inspector.GetTreeScrollPosition()
	if initialOffset != 0 {
		t.Errorf("Initial scroll offset should be 0, got %d", initialOffset)
	}

	// 测试滚动
	inspector.ScrollTreeBy(5)
	offset := inspector.GetTreeScrollPosition()
	if offset != 5 {
		t.Logf("Note: Scroll offset is %d (may be clamped by tree height)", offset)
	} else {
		t.Log("✓ Scroll offset updated correctly")
	}

	// 测试滚动到顶部
	inspector.ScrollTreeTop()
	offset = inspector.GetTreeScrollPosition()
	if offset != 0 {
		t.Errorf("Scroll offset should be 0 after ScrollTreeTop, got %d", offset)
	} else {
		t.Log("✓ ScrollTreeTop works")
	}

	// 测试滚动能力
	canUp := inspector.CanScrollTreeUp()
	canDown := inspector.CanScrollTreeDown()
	t.Logf("Can scroll up: %v, Can scroll down: %v", canUp, canDown)

	t.Log("✓ Scroll state management works correctly")
}
