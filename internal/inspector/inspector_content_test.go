package inspector

import (
	"strings"
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestInspectorWithRealContent 测试 Inspector 与实际内容的集成
func TestInspectorWithRealContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建 Inspector
	inspector := NewStandaloneInspector()
	inspector.Enable()
	inspector.ToggleVisibility()
	inspector.SetOverlaySize(100, 40)

	// 创建一个复杂的应用树
	appRoot := ui.VStack(
		ui.Text("Root Container"),
		ui.Text("───────────────"),
		ui.HStack(
			ui.Text("Left"),
			ui.NewButtonBuilder("[Button A]").Build(),
			ui.Text("Right"),
		),
		ui.VStack(
			ui.Text("Nested VStack"),
			ui.NewButtonBuilder("[Button B]").Build(),
			ui.NewButtonBuilder("[Button C]").Build(),
		),
		ui.Text("───────────────"),
		ui.Text("End of content"),
	)

	// 附加应用到 Inspector
	inspector.AttachToApp(appRoot)

	testApp, err := ui.RunTest(func() ui.VNode {
		return ui.VStack(
			appRoot,
			ui.Text(""),
			ui.Text("Press 'i' to toggle Inspector"),
		)
	},
		ui.WithWidth(100),
		ui.WithHeight(40),
		ui.WithTitle("Inspector with Real Content"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(300 * time.Millisecond)

	// 获取树视图内容
	treeView := inspector.GetTreeView()
	lines, totalLines := treeView.GetTreeLines()

	t.Logf("Tree has %d lines, %d total nodes", len(lines), totalLines)
	t.Logf("=== Tree Content (first 30 lines) ===")
	for i := 0; i < len(lines) && i < 30; i++ {
		t.Logf("  %s", lines[i])
	}
	t.Logf("=== End ===")

	// 验证树有内容
	if totalLines == 0 {
		t.Error("Tree should have content after AttachToApp")
	}

	if len(lines) == 0 {
		t.Error("Tree lines should not be empty")
	}

	// 验证包含当前实现下稳定可见的节点类型和内容标签。
	// HStack/VStack 在树视图中已经归一显示为 LayoutNode。
	expectedTypes := []string{
		"LayoutNode",
		"TextVNode",
		"ButtonVNode",
		"Root Container",
		"Nested VStack",
	}
	missingTypes := []string{}
	treeContent := strings.Join(lines, " ")

	for _, expectedType := range expectedTypes {
		if !strings.Contains(treeContent, expectedType) {
			missingTypes = append(missingTypes, expectedType)
		}
	}

	if len(missingTypes) > 0 {
		t.Errorf("Tree is missing expected node types: %v", missingTypes)
	} else {
		t.Log("✓ Tree contains all expected node types")
	}

	// 现在测试 Inspector 的渲染
	overlay := inspector.RenderOverlay()
	if overlay == nil {
		t.Fatal("Inspector overlay is nil")
	}

	testApp2, err := ui.RunTest(func() ui.VNode {
		return overlay
	},
		ui.WithWidth(100),
		ui.WithHeight(40),
		ui.WithTitle("Inspector Overlay"),
	)
	if err != nil {
		t.Fatalf("Failed to create overlay test app: %v", err)
	}
	defer testApp2.Close()

	time.Sleep(300 * time.Millisecond)

	// 获取渲染输出
	rendered := testApp2.GetRenderString()

	t.Logf("=== Inspector Overlay Render ===\n%s\n=== End ===", rendered)

	// 验证 Inspector 显示了树内容
	if !strings.Contains(rendered, "LayoutNode") && !strings.Contains(rendered, "ButtonVNode") {
		t.Error("Inspector should display tree nodes")
	} else {
		t.Log("✓ Inspector displays tree nodes")
	}

	// 验证显示节点数量
	if strings.Contains(rendered, "Nodes: 0") {
		t.Error("Inspector should show non-zero node count")
	} else if strings.Contains(rendered, "Nodes:") {
		t.Log("✓ Inspector shows node count")
	}
}

// TestInspectorWithAttachedApp 测试完整的 Inspector 功能
func TestInspectorWithAttachedApp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建 Inspector
	inspector := NewStandaloneInspector()
	inspector.Enable()
	inspector.ToggleVisibility()
	inspector.SetOverlaySize(100, 40)

	// 创建应用
	createApp := func() ui.VNode {
		return ui.VStack(
			ui.Text("Demo Application"),
			ui.Text("────────────────"),
			ui.HStack(
				ui.Text("Column 1:"),
				ui.VStack(
					ui.NewButtonBuilder("[Btn 1]").Build(),
					ui.NewButtonBuilder("[Btn 2]").Build(),
					ui.NewButtonBuilder("[Btn 3]").Build(),
				),
				ui.Text("  "),
				ui.Text("Column 2:"),
				ui.VStack(
					ui.NewButtonBuilder("[Btn 4]").Build(),
					ui.NewButtonBuilder("[Btn 5]").Build(),
				),
			),
			ui.Text("────────────────"),
			ui.Text("Footer content"),
		)
	}

	// 附加应用到 Inspector
	appRoot := createApp()
	inspector.AttachToApp(appRoot)

	// 获取树信息
	treeView := inspector.GetTreeView()
	lines, _ := treeView.GetTreeLines()

	t.Logf("Application tree has %d lines", len(lines))

	// 验证树的结构。当前 Inspector 会把布局容器显示为 LayoutNode。
	layoutNodeCount := 0
	hasButton := false
	hasDemoLabel := false
	hasColumnLabel := false

	for _, line := range lines {
		if strings.Contains(line, "LayoutNode") {
			layoutNodeCount++
		}
		if strings.Contains(line, "Button") {
			hasButton = true
		}
		if strings.Contains(line, "Demo Application") {
			hasDemoLabel = true
		}
		if strings.Contains(line, "Column 1:") {
			hasColumnLabel = true
		}
	}

	if layoutNodeCount < 2 {
		t.Errorf("Tree should contain multiple LayoutNode containers, got %d", layoutNodeCount)
	} else {
		t.Logf("✓ Tree contains %d LayoutNode containers", layoutNodeCount)
	}

	if !hasButton {
		t.Error("Tree should contain Button nodes")
	} else {
		t.Log("✓ Tree contains Buttons")
	}

	if !hasDemoLabel {
		t.Error("Tree should contain the Demo Application label")
	}

	if !hasColumnLabel {
		t.Error("Tree should contain the Column 1 label")
	}

	// 渲染 Inspector 查看实际显示
	overlay := inspector.RenderOverlay()

	testApp, err := ui.RunTest(func() ui.VNode {
		return overlay
	},
		ui.WithWidth(100),
		ui.WithHeight(40),
		ui.WithTitle("Full Inspector Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(300 * time.Millisecond)

	rendered := testApp.GetRenderString()

	t.Logf("=== Full Inspector Render ===\n%s\n=== End ===", rendered)

	// 验证关键内容存在
	expectedContent := []string{
		"Elements",
		"Console(2)",
		"Layout Tree",
		"Nodes:",
	}

	missing := []string{}
	for _, expected := range expectedContent {
		if !strings.Contains(rendered, expected) {
			missing = append(missing, expected)
		}
	}

	if len(missing) > 0 {
		t.Errorf("Inspector missing content: %v", missing)
	} else {
		t.Log("✓ All expected content present")
	}

	// 检查节点数是否大于0
	if strings.Contains(rendered, "Nodes: 0") {
		t.Error("Node count should be greater than 0")
		t.Log("This suggests AttachToApp didn't work properly")
	} else if strings.Contains(rendered, "Nodes: ") {
		// 提取节点数
		idx := strings.Index(rendered, "Nodes: ")
		if idx >= 0 {
			nodeCountStr := rendered[idx+7:]
			endIdx := strings.IndexAny(nodeCountStr, " |\n")
			if endIdx > 0 {
				nodeCountStr = nodeCountStr[:endIdx]
				t.Logf("✓ Inspector shows: Nodes: %s", nodeCountStr)
			}
		}
	}
}
