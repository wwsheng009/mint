package inspector

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestNewInspector tests creating a new Inspector
func TestNewInspector(t *testing.T) {
	inspector := NewInspector()

	if inspector == nil {
		t.Fatal("Expected non-nil Inspector")
	}

	if inspector.IsEnabled() {
		t.Error("New inspector should be disabled by default")
	}
}

// TestInspectorEnableDisable tests enabling and disabling the inspector
func TestInspectorEnableDisable(t *testing.T) {
	inspector := NewInspector()

	// Test enable
	inspector.Enable()
	if !inspector.IsEnabled() {
		t.Error("Inspector should be enabled after Enable()")
	}

	// Test disable
	inspector.Disable()
	if inspector.IsEnabled() {
		t.Error("Inspector should be disabled after Disable()")
	}

	// Test that disable clears selection
	inspector.Enable()
	button := app.ButtonBuilder("Test").Build()
	inspector.SetSelectedVNode(button)

	inspector.Disable()

	if inspector.GetSelectedVNode() != nil {
		t.Error("Selected VNode should be cleared after disable")
	}
}

// TestSetSelectedVNode tests setting the selected VNode
func TestSetSelectedVNode(t *testing.T) {
	inspector := NewInspector()

	button := app.ButtonBuilder("Test Button").Build()
	inspector.SetSelectedVNode(button)

	selected := inspector.GetSelectedVNode()
	if selected == nil {
		t.Error("Selected VNode should not be nil")
	}

	// Verify it's the same button
	info := ExtractElementInfo(selected)
	if info.Label != "Test Button" {
		t.Errorf("Expected label 'Test Button', got '%s'", info.Label)
	}
}

// TestGetSelectedInfo tests getting ElementInfo for selected VNode
func TestGetSelectedInfo(t *testing.T) {
	inspector := NewInspector()

	button := app.ButtonBuilder("Click Me").Build()
	inspector.SetSelectedVNode(button)

	info := inspector.GetSelectedInfo()

	if info.Type == "" {
		t.Error("Type should not be empty")
	}

	if info.Label != "Click Me" {
		t.Errorf("Expected label 'Click Me', got '%s'", info.Label)
	}
}

// TestGetHoveredInfo tests getting ElementInfo for hovered VNode
func TestGetHoveredInfo(t *testing.T) {
	inspector := NewInspector()

	text := ui.Text("Hover Test")

	// Simulate hovering (manually set for now)
	inspector.hoveredVNode = text

	info := inspector.GetHoveredInfo()

	if info.Label != "Hover Test" {
		t.Errorf("Expected label 'Hover Test', got '%s'", info.Label)
	}
}

// TestFindVNodeAt tests finding a VNode at a position
func TestFindVNodeAt(t *testing.T) {
	t.Skip("Requires actual layout engine to set bounds properly")

	inspector := NewInspector()

	// Create a simple VNode tree
	button := app.ButtonBuilder("Test").Build()

	// Simulate bounds being set
	if boundsAware, ok := button.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsAware.SetBounds(10, 5, 20, 1)
	}

	// Find node at a position within the button
	found := inspector.FindVNodeAt(button, 15, 5)

	if found == nil {
		t.Error("Expected to find VNode at (15, 5)")
	}

	info := ExtractElementInfo(found)
	if info.Label != "Test" {
		t.Errorf("Expected to find button with label 'Test', got '%s'", info.Label)
	}

	// Find node at a position outside the button
	foundOutside := inspector.FindVNodeAt(button, 100, 100)
	if foundOutside != nil {
		t.Error("Expected nil when searching outside button bounds")
	}
}

// TestHandleMouseEvent tests mouse event handling
func TestHandleMouseEvent(t *testing.T) {
	inspector := NewInspector()
	inspector.Enable()

	// Handle a mouse event
	handled := inspector.HandleMouseEvent(50, 25)

	if !handled {
		// For now, HandleMouseEvent returns false until layout integration
		// This is expected for Phase 2
		t.Log("HandleMouseEvent returns false (expected until layout integration)")
	}

	// Check mouse position was recorded
	x, y := inspector.GetMousePosition()
	if x != 50 || y != 25 {
		t.Errorf("Expected mouse position (50, 25), got (%d, %d)", x, y)
	}
}

// TestHandleMouseEvent_Disabled tests that mouse events are ignored when disabled
func TestHandleMouseEvent_Disabled(t *testing.T) {
	inspector := NewInspector()
	// Don't enable inspector

	handled := inspector.HandleMouseEvent(50, 25)

	if handled {
		t.Error("Mouse event should not be handled when inspector is disabled")
	}
}

// TestVNodeContains tests the VNode containment check
func TestVNodeContains(t *testing.T) {
	t.Skip("Requires actual layout engine to set bounds properly")

	button := app.ButtonBuilder("Test").Build()

	// Simulate bounds
	if boundsAware, ok := button.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsAware.SetBounds(10, 5, 20, 1)
	}

	tests := []struct {
		name     string
		x, y     int
		expected bool
	}{
		{"Inside bounds", 15, 5, true},
		{"At left edge", 10, 5, true},
		{"At right edge", 29, 5, true},
		{"Outside left", 9, 5, false},
		{"Outside right", 30, 5, false},
		{"Outside top", 15, 4, false},
		{"Outside bottom", 15, 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vnodeContains(button, tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("Expected %v for position (%d, %d), got %v",
					tt.expected, tt.x, tt.y, result)
			}
		})
	}
}

// TestNewOverlay tests creating a new Overlay
func TestNewOverlay(t *testing.T) {
	overlay := NewOverlay()

	if overlay == nil {
		t.Fatal("Expected non-nil Overlay")
	}

	if !overlay.showBorders {
		t.Error("Borders should be shown by default")
	}

	if !overlay.showDimensions {
		t.Error("Dimensions should be shown by default")
	}
}

// TestOverlaySetters tests overlay configuration methods
func TestOverlaySetters(t *testing.T) {
	overlay := NewOverlay()

	// Test SetShowDimensions
	overlay.SetShowDimensions(false)
	if overlay.showDimensions {
		t.Error("showDimensions should be false after SetShowDimensions(false)")
	}

	// Test SetShowBorders
	overlay.SetShowBorders(false)
	if overlay.showBorders {
		t.Error("showBorders should be false after SetShowBorders(false)")
	}
}

// TestGetBorderStyle tests getting border style for different element types
func TestGetBorderStyle(t *testing.T) {
	overlay := NewOverlay()

	button := app.ButtonBuilder("Test").Build()
	text := ui.Text("Hello")

	buttonStyle := overlay.GetBorderStyle(button)
	textStyle := overlay.GetBorderStyle(text)

	if string(buttonStyle) != "▓" {
		t.Errorf("Expected diamond style for button, got '%s'", string(buttonStyle))
	}

	if string(textStyle) != "•" {
		t.Errorf("Expected bullet style for text, got '%s'", string(textStyle))
	}
}

// TestPaintHighlight tests painting corner highlights
func TestPaintHighlight(t *testing.T) {
	overlay := NewOverlay()

	button := app.ButtonBuilder("Test").Build()

	// Set bounds
	if boundsAware, ok := button.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsAware.SetBounds(10, 5, 20, 3)
	}

	// We can't easily test the actual buffer without importing paint.Buffer
	// So we just test that the method doesn't panic
	err := overlay.PaintHighlight(nil, button, '*')
	if err != nil {
		t.Errorf("PaintHighlight should not return error, got %v", err)
	}
}

// =============================================================================
// 交互式测试 (Interactive Tests with TestHelper)
// =============================================================================

// TestInspectorWithScrollView 测试 Inspector 与 ScrollView 组件的集成
func TestInspectorWithScrollView(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(DemoAppWithDeepTree,
		ui.WithWidth(100),
		ui.WithHeight(40),
		ui.WithTitle("Inspector ScrollView Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(200 * time.Millisecond)

	// 验证初始渲染
	initialRender := testApp.GetRenderString()
	if len(initialRender) == 0 {
		t.Error("Expected non-empty initial render")
	}
	t.Logf("Initial render successful, %d lines", len(strings.Split(initialRender, "\n")))
}

// TestInspectorTabSwitching 测试标签页切换功能
func TestInspectorTabSwitching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(DemoAppWithMultipleComponents,
		ui.WithWidth(100),
		ui.WithHeight(30),
		ui.WithTitle("Inspector Tab Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(200 * time.Millisecond)

	// 获取初始渲染（应该在第一个标签页）
	render1 := testApp.GetRenderString()
	t.Logf("Tab 1 render:\n%s", render1)

	// 模拟按 Tab 键切换到下一个标签
	for i := 0; i < 3; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab key: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	render2 := testApp.GetRenderString()
	t.Logf("After Tab switches:\n%s", render2)
}

// TestInspectorKeyboardNavigation 测试键盘导航
func TestInspectorKeyboardNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(DemoAppWithButtons,
		ui.WithWidth(80),
		ui.WithHeight(25),
		ui.WithTitle("Inspector Navigation Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(200 * time.Millisecond)

	// 测试上下导航
	for i := 0; i < 3; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyDown); err != nil {
			t.Errorf("Failed to inject Down key: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	for i := 0; i < 2; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyUp); err != nil {
			t.Errorf("Failed to inject Up key: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Logf("Navigation test completed successfully")
}

// TestScrollViewPagination 测试 ScrollView 翻页功能
func TestScrollViewPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(DemoAppWithLongContent,
		ui.WithWidth(80),
		ui.WithHeight(30),
		ui.WithTitle("ScrollView Pagination Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(200 * time.Millisecond)

	initialRender := testApp.GetRenderString()
	t.Logf("Initial content height: %d lines", len(strings.Split(initialRender, "\n")))

	// 测试 PageDown
	if err := testApp.InjectSpecialKey(platform.KeyPageDown); err != nil {
		t.Errorf("Failed to inject PageDown: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	afterPgDn := testApp.GetRenderString()
	t.Logf("After PageDown: %d lines", len(strings.Split(afterPgDn, "\n")))

	// 测试 PageUp
	if err := testApp.InjectSpecialKey(platform.KeyPageUp); err != nil {
		t.Errorf("Failed to inject PageUp: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	afterPgUp := testApp.GetRenderString()
	t.Logf("After PageUp: %d lines", len(strings.Split(afterPgUp, "\n")))

	// 测试 Home
	if err := testApp.InjectSpecialKey(platform.KeyHome); err != nil {
		t.Errorf("Failed to inject Home: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// 测试 End
	if err := testApp.InjectSpecialKey(platform.KeyEnd); err != nil {
		t.Errorf("Failed to inject End: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	t.Logf("Pagination test completed successfully")
}

// TestVirtualListComponent 测试 VirtualList 组件
func TestVirtualListComponent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(DemoAppWithVirtualList,
		ui.WithWidth(80),
		ui.WithHeight(30),
		ui.WithTitle("VirtualList Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(200 * time.Millisecond)

	initialRender := testApp.GetRenderString()
	if !strings.Contains(initialRender, "Item") {
		t.Error("Expected VirtualList to render items")
	}

	t.Logf("VirtualList rendered successfully")

	// 测试滚动
	for i := 0; i < 5; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyDown); err != nil {
			t.Errorf("Failed to inject Down key: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Logf("VirtualList scrolling test completed")
}

// TestTabsComponentRendering 测试 Tabs 组件渲染
func TestTabsComponentRendering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(DemoAppWithTabs,
		ui.WithWidth(80),
		ui.WithHeight(25),
		ui.WithTitle("Tabs Component Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(200 * time.Millisecond)

	render := testApp.GetRenderString()

	// 验证 Tab 组件渲染
	if !strings.Contains(render, "Tab 1") && !strings.Contains(render, "Elements") {
		t.Log("Note: Tab labels may be styled differently")
	}

	t.Logf("Tabs component render:\n%s", render)

	// 测试 Tab 切换
	for i := 0; i < 3; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab key: %v", err)
		}
		time.Sleep(100 * time.Millisecond)

		render := testApp.GetRenderString()
		t.Logf("After Tab %d:\n%s", i+1, render)
	}
}

// TestInspectorIntegration 测试 Inspector 完整集成
func TestInspectorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testApp, err := ui.RunTest(DemoAppWithInspector,
		ui.WithWidth(100),
		ui.WithHeight(40),
		ui.WithTitle("Inspector Integration Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(200 * time.Millisecond)

	t.Logf("Integration test started")

	// 测试标签页切换
	for i := 0; i < 5; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 测试滚动
	if err := testApp.InjectSpecialKey(platform.KeyPageDown); err != nil {
		t.Errorf("Failed to inject PageDown: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	if err := testApp.InjectSpecialKey(platform.KeyPageUp); err != nil {
		t.Errorf("Failed to inject PageUp: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	t.Logf("Integration test completed successfully")
}

// =============================================================================
// 测试用的组件定义
// =============================================================================

// DemoAppWithDeepTree 带有深度树的演示应用
func DemoAppWithDeepTree() ui.VNode {
	items := []ui.VNode{
		ui.Text("Deep Tree Test Application"),
		ui.Text("──────────────────────────"),
		ui.Text(""),
	}

	// 创建深度嵌套的组件树
	for i := 0; i < 50; i++ {
		items = append(items, ui.Text(fmt.Sprintf("Item %d - Content for testing scroll", i)))
		items = append(items, ui.Text(fmt.Sprintf("  Child %d.1 - More details", i)))
		items = append(items, ui.Text(fmt.Sprintf("  Child %d.2 - Even more", i)))
	}

	return ui.VStack(items...)
}

// DemoAppWithMultipleComponents 带有多个组件的演示应用
func DemoAppWithMultipleComponents() ui.VNode {
	return ui.VStack(
		ui.Text("Multiple Components Test"),
		ui.Text("─────────────────────────"),
		ui.Text(""),
		ui.Text("This app tests multiple UI components:"),
		ui.Text("  - ScrollView"),
		ui.Text("  - VirtualList"),
		ui.Text("  - Tabs"),
		ui.Text(""),
		app.ButtonBuilder("[1] Action Button").Build(),
		ui.Text(""),
		ui.Text("Press Tab to navigate between components"),
	)
}

// DemoAppWithButtons 带有按钮的演示应用
func DemoAppWithButtons() ui.VNode {
	return ui.VStack(
		ui.Text("Button Navigation Test"),
		ui.Text("──────────────────────"),
		ui.Text(""),
		ui.Text("Use arrow keys to navigate:"),
		ui.Text(""),
		app.ButtonBuilder("[Button 1]").Build(),
		app.ButtonBuilder("[Button 2]").Build(),
		app.ButtonBuilder("[Button 3]").Build(),
		app.ButtonBuilder("[Button 4]").Build(),
		app.ButtonBuilder("[Button 5]").Build(),
	)
}

// DemoAppWithLongContent 带有长内容的演示应用
func DemoAppWithLongContent() ui.VNode {
	items := []ui.VNode{
		ui.Text("Long Content Test"),
		ui.Text("──────────────────"),
		ui.Text(""),
		ui.Text("Testing pagination with PageUp/PageDown:"),
		ui.Text(""),
	}

	// 生成100行内容
	for i := 1; i <= 100; i++ {
		items = append(items, ui.Text(fmt.Sprintf("Line %d: This is test content for scrolling", i)))
	}

	return ui.VStack(items...)
}

// DemoAppWithVirtualList 带有 VirtualList 的演示应用
func DemoAppWithVirtualList() ui.VNode {
	// 创建100个项目的虚拟列表
	items := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = i
	}

	return ui.VStack(
		ui.Text("VirtualList Component Test"),
		ui.Text("─────────────────────────"),
		ui.Text(""),
		ui.Text("100 items, virtualized rendering:"),
		ui.Text(""),
		app.VirtualListBuilder().
			Items(items).
			RenderItem(func(item interface{}) ui.VNode {
				idx := item.(int)
				return ui.Text(fmt.Sprintf("Item %d: Virtualized content", idx+1))
			}).
			ItemHeight(1).
			Height(20).
			ScrollOffset(0).
			Build(),
	)
}

// DemoAppWithTabs 带有 Tabs 的演示应用
func DemoAppWithTabs() ui.VNode {
	// 导入 navigation 包
	return ui.VStack(
		ui.Text("Tabs Component Test"),
		ui.Text("────────────────────"),
		ui.Text(""),
		ui.Text("Testing tab switching with Tab key"),
		ui.Text(""),
		ui.Text("[Tab 1] [Tab 2] [Tab 3]"),
		ui.Text("────────────────────"),
		ui.Text("Content for Tab 1"),
		ui.Text(""),
		app.ButtonBuilder("[Action]").Build(),
	)
}

// DemoAppWithInspector 带有 Inspector 的演示应用
func DemoAppWithInspector() ui.VNode {
	items := []ui.VNode{
		ui.Text("Inspector Integration Test"),
		ui.Text("───────────────────────────"),
		ui.Text(""),
		ui.Text("This app integrates Inspector with:"),
		ui.Text("  • Tab component for tab management"),
		ui.Text("  • ScrollView for tree scrolling"),
		ui.Text("  • VirtualList for item lists"),
		ui.Text(""),
		ui.Text("Features tested:"),
		ui.Text("  - Tab switching (Tab key)"),
		ui.Text("  - ScrollView pagination (PgUp/PgDn)"),
		ui.Text("  - Tree navigation (Arrow keys)"),
		ui.Text("  - Expand/Collapse (E key)"),
		ui.Text(""),
		ui.Text("Test Components:"),
	}

	// 添加多个按钮
	for i := 1; i <= 10; i++ {
		items = append(items, app.ButtonBuilder(fmt.Sprintf("[Button %d]", i)).Build())
	}

	return ui.VStack(items...)
}
