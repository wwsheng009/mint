package navigation

import (
	"fmt"
	"strings"
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestTabsContentRendering 测试 Tab 组件是否能渲染内容
func TestTabsContentRendering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建一个带内容的 Tab 组件
	tabsBuilder := TabsBuilder()

	// 添加第一个标签页
	tabsBuilder.AddTab("tab1", "Tab 1")
	tabsBuilder.Content("tab1", ui.VStack(
		ui.Text("Content for Tab 1"),
		ui.Text("Line 2"),
		ui.Text("Line 3"),
	))

	// 添加第二个标签页
	tabsBuilder.AddTab("tab2", "Tab 2")
	tabsBuilder.Content("tab2", ui.VStack(
		ui.Text("Content for Tab 2"),
		ui.Text("Different content"),
	))

	// 设置活动标签为第一个
	tabsBuilder.ActiveTab(0)
	tabsComponent := tabsBuilder.Build()

	// 创建测试应用
	testApp, err := ui.RunTest(func() ui.VNode {
		return ui.VStack(
			ui.Text("Tab Test"),
			ui.Text("────────"),
			tabsComponent,
		)
	},
		ui.WithWidth(60),
		ui.WithHeight(15),
		ui.WithTitle("Tabs Content Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待渲染
	time.Sleep(200 * time.Millisecond)

	// 获取渲染输出
	rendered := testApp.GetRenderString()

	t.Logf("=== Tab Component Render ===\n%s\n=== End ===", rendered)

	// 验证内容存在
	if !strings.Contains(rendered, "Tab 1") {
		t.Error("Tab label not found in render")
	}

	if !strings.Contains(rendered, "Content for Tab 1") {
		t.Error("Tab content not found in render")
		t.Logf("Expected 'Content for Tab 1' but not found")
	} else {
		t.Log("✓ Tab content rendered successfully!")
	}
}

// TestTabSwitching 测试切换标签页
func TestTabSwitching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建 Tab 组件
	tabsBuilder := TabsBuilder()

	tabsBuilder.AddTab("tab1", "First")
	tabsBuilder.Content("tab1", ui.Text("First Tab Content"))

	tabsBuilder.AddTab("tab2", "Second")
	tabsBuilder.Content("tab2", ui.Text("Second Tab Content"))

	tabsBuilder.ActiveTab(0)
	tabsComponent := tabsBuilder.Build().(*TabsVNode)

	testApp, err := ui.RunTest(func() ui.VNode {
		return tabsComponent
	},
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("Tab Switching Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(150 * time.Millisecond)

	// 初始状态 - 应该显示第一个标签的内容
	render1 := testApp.GetRenderString()
	t.Logf("=== Initial render (Tab 1) ===\n%s\n=== End ===", render1)

	if !strings.Contains(render1, "First Tab Content") {
		t.Error("First tab content not found")
	} else {
		t.Log("✓ First tab content rendered")
	}

	// 切换到第二个标签
	tabsComponent.SetActiveTab(1)
	time.Sleep(150 * time.Millisecond)

	render2 := testApp.GetRenderString()
	t.Logf("=== After switching to Tab 2 ===\n%s\n=== End ===", render2)

	if !strings.Contains(render2, "Second Tab Content") {
		t.Error("Second tab content not found")
	} else {
		t.Log("✓ Second tab content rendered after switching")
	}
}

// TestTabsWithScrollView 测试带 ScrollView 的标签页
func TestTabsWithScrollView(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建长文本
	longText := ""
	for i := 1; i <= 30; i++ {
		longText += strings.TrimSpace(fmt.Sprintf("Line %d: ScrollView test content\n", i))
	}

	tabsBuilder := TabsBuilder()

	tabsBuilder.AddTab("scroll", "ScrollTab")
	tabsBuilder.Content("scroll", ui.VStack(
		ui.Text("ScrollView Demo"),
		ui.Text("─────────────"),
		ui.Text(longText),
	))

	tabsBuilder.ActiveTab(0)

	testApp, err := ui.RunTest(func() ui.VNode {
		return tabsBuilder.Build()
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Tabs with ScrollView"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("=== ScrollView Tab Render ===\n%s\n=== End ===", rendered)

	if strings.Contains(rendered, "Line 1") && strings.Contains(rendered, "Line 30") {
		t.Log("✓ ScrollView content visible")
	} else if strings.Contains(rendered, "Line 1") || strings.Contains(rendered, "Line 2") {
		t.Log("✓ At least some content is visible (virtual scrolling may be working)")
	} else {
		t.Error("ScrollView content not visible")
	}
}
