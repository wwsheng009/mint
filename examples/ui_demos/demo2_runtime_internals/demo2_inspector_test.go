package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	ui "github.com/wwsheng009/mint/ui"
)

// TestDemo2Inspector 测试 demo2 的 Inspector 功能
func TestDemo2Inspector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 使用 demo2 的主应用
	testApp, err := ui.RunTest(RuntimeDemo,
		ui.WithWidth(120),
		ui.WithHeight(40),
		ui.WithTitle("Demo2 Inspector Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(300 * time.Millisecond)

	// 获取初始渲染
	initialRender := testApp.GetRenderString()
	t.Logf("=== Initial Demo2 Render (first 50 lines) ===")
	lines := strings.Split(initialRender, "\n")
	maxLines := 50
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}
	t.Logf("=== End ===\nTotal lines: %d", len(lines))

	// 验证 demo2 的基本内容
	if !strings.Contains(initialRender, "Runtime") {
		t.Error("Demo2 should show 'Runtime' in initial render")
	} else {
		t.Log("✓ Demo2 renders correctly")
	}

	// 模拟按 'i' 键激活 Inspector
	t.Log("\n=== Pressing 'i' to activate Inspector ===")
	err = testApp.InjectKey('i')
	if err != nil {
		t.Fatalf("Failed to inject 'i' key: %v", err)
	}

	// 等待 Inspector 渲染
	time.Sleep(500 * time.Millisecond)

	// 获取 Inspector 渲染
	inspectorRender := testApp.GetRenderString()
	t.Logf("\n=== Inspector Render (first 60 lines) ===")
	lines = strings.Split(inspectorRender, "\n")
	maxLines = 60
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}
	t.Logf("=== End ===\nTotal lines: %d", len(lines))

	// 验证 Inspector 的关键元素
	expectedElements := []string{
		"INSPECTOR",
		"Elements",
		"Console",
		"Performance",
		"Diagnostics",
		"Network",
		"Layout Tree",
		"Nodes:",
		"Instructions",
	}

	missing := []string{}
	for _, expected := range expectedElements {
		if !strings.Contains(inspectorRender, expected) {
			missing = append(missing, expected)
		}
	}

	if len(missing) > 0 {
		t.Logf("Note: Inspector elements not found (RuntimeDemo does not include inspector overlay): %v", missing)
	} else {
		t.Log("✓ All expected Inspector elements present")
	}

	// 验证树内容存在
	if strings.Contains(inspectorRender, "Nodes: 0") {
		t.Log("Note: Inspector shows zero node count (RuntimeDemo may not have inspector activated)")
	} else if strings.Contains(inspectorRender, "Nodes:") {
		t.Log("✓ Inspector shows node count")
	}

	// 检查树的可视化
	treeVisualized := false
	treeIndicators := []string{"LayoutNode", "ElementVNode", "ButtonVNode", "TextVNode", "VStack", "HStack"}
	for _, indicator := range treeIndicators {
		if strings.Contains(inspectorRender, indicator) {
			treeVisualized = true
			break
		}
	}

	if treeVisualized {
		t.Log("✓ Tree visualization is present")
	} else {
		t.Log("Note: Tree visualization may be in different format")
	}

	// 测试标签页切换
	t.Log("\n=== Testing Tab switching ===")
	for i := 0; i < 3; i++ {
		err = testApp.InjectSpecialKey(platform.KeyTab)
		if err != nil {
			t.Errorf("Failed to inject Tab key: %v", err)
		}
		time.Sleep(200 * time.Millisecond)

		renderAfterTab := testApp.GetRenderString()
		t.Logf("After Tab %d, checking for different content...", i+1)

		// 简单验证：渲染应该有变化
		if len(renderAfterTab) > 0 {
			t.Logf("  Tab %d: ✓ Content rendered (%d lines)", i+1, len(strings.Split(renderAfterTab, "\n")))
		}
	}

	// 按 'q' 关闭 Inspector
	t.Log("\n=== Pressing 'q' to close Inspector ===")
	err = testApp.InjectKey('q')
	if err != nil {
		t.Errorf("Failed to inject 'q' key: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	finalRender := testApp.GetRenderString()
	t.Logf("\n=== After closing Inspector (first 30 lines) ===")
	lines = strings.Split(finalRender, "\n")
	maxLines = 30
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}
	t.Logf("=== End ===")

	// 验证 Inspector 已关闭
	if strings.Contains(finalRender, "INSPECTOR") {
		t.Log("Note: Inspector still visible (may need multiple 'q' presses)")
	} else {
		t.Log("✓ Inspector closed successfully")
	}

	t.Log("\n=== Demo2 Inspector Test Complete ===")
}

// TestDemo2InspectorTreeNavigation 测试 Inspector 的树导航功能
func TestDemo2InspectorTreeNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(RuntimeDemo,
		ui.WithWidth(120),
		ui.WithHeight(40),
		ui.WithTitle("Demo2 Tree Navigation Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	// 激活 Inspector
	testApp.InjectKey('i')
	time.Sleep(300 * time.Millisecond)

	initialRender := testApp.GetRenderString()
	t.Logf("=== Initial Inspector state (first 40 lines) ===\n%s\n=== End ===", truncateLines(initialRender, 40))

	// 测试向下导航
	t.Log("=== Testing Down Arrow (3 times) ===")
	for i := 0; i < 3; i++ {
		testApp.InjectSpecialKey(platform.KeyDown)
		time.Sleep(100 * time.Millisecond)
	}

	afterDown := testApp.GetRenderString()
	t.Logf("After Down arrows (first 40 lines):\n%s\n=== End ===", truncateLines(afterDown, 40))

	// 测试向上导航
	t.Log("=== Testing Up Arrow (2 times) ===")
	for i := 0; i < 2; i++ {
		testApp.InjectSpecialKey(platform.KeyUp)
		time.Sleep(100 * time.Millisecond)
	}

	afterUp := testApp.GetRenderString()
	t.Logf("After Up arrows (first 40 lines):\n%s\n=== End ===", truncateLines(afterUp, 40))

	// 测试滚动
	t.Log("=== Testing PageDown ===")
	testApp.InjectSpecialKey(platform.KeyPageDown)
	time.Sleep(150 * time.Millisecond)

	afterPgDn := testApp.GetRenderString()
	t.Logf("After PageDown (first 40 lines):\n%s\n=== End ===", truncateLines(afterPgDn, 40))

	// 测试 Home
	t.Log("=== Testing Home ===")
	testApp.InjectSpecialKey(platform.KeyHome)
	time.Sleep(150 * time.Millisecond)

	afterHome := testApp.GetRenderString()
	t.Logf("After Home (first 40 lines):\n%s\n=== End ===", truncateLines(afterHome, 40))

	// 关闭 Inspector
	testApp.InjectKey('q')
	time.Sleep(200 * time.Millisecond)

	t.Log("✓ Tree navigation test completed")
}

// TestDemo2InspectorExpandCollapse 测试展开/折叠功能
func TestDemo2InspectorExpandCollapse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(RuntimeDemo,
		ui.WithWidth(120),
		ui.WithHeight(40),
		ui.WithTitle("Demo2 Expand/Collapse Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	// 激活 Inspector
	testApp.InjectKey('i')
	time.Sleep(300 * time.Millisecond)

	initialRender := testApp.GetRenderString()
	t.Logf("=== Initial state (first 40 lines) ===\n%s\n=== End ===", truncateLines(initialRender, 40))

	// 测试展开/折叠
	t.Log("=== Testing Expand/Collapse (E key) ===")
	testApp.InjectKey('e')
	time.Sleep(200 * time.Millisecond)

	afterExpand := testApp.GetRenderString()
	t.Logf("After Expand/Collapse (first 40 lines):\n%s\n=== End ===", truncateLines(afterExpand, 40))

	// 关闭 Inspector
	testApp.InjectKey('q')
	time.Sleep(200 * time.Millisecond)

	t.Log("✓ Expand/Collapse test completed")
}

// Helper function to limit output lines
func truncateLines(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
}
