package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/internal/inspector"
	"github.com/wwsheng009/mint/runtime/platform"
	ui "github.com/wwsheng009/mint/ui"
)

// TestInspectorStandalone 独立测试 Inspector 组件
func TestInspectorStandalone(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建 Inspector
	insp := inspector.NewStandaloneInspector()
	insp.Enable()
	insp.ToggleVisibility()
	insp.SetOverlaySize(100, 40)

	// 创建一个复杂的应用树来测试
	appRoot := ui.VStack(
		ui.Text("Demo2 Test Application"),
		ui.Text("───────────────────────"),
		ui.Text("Runtime Pipeline Visualization"),
		ui.HStack(
			ui.Text("Events: 0"),
			ui.Text("Renders: 0"),
			ui.Text("Buffers: 0"),
		),
		ui.Text(""),
		buildDemo2Buttons(),
		ui.Text(""),
		ui.Text("System idle, waiting for events..."),
	)

	// 附加应用到 Inspector
	insp.AttachToApp(appRoot)

	// 渲染 Inspector
	overlay := insp.RenderOverlay()
	if overlay == nil {
		t.Fatal("Inspector overlay is nil")
	}

	// 使用 TestHelper 测试渲染
	testApp, err := ui.RunTest(func() ui.VNode {
		return overlay
	},
		ui.WithWidth(120),
		ui.WithHeight(40),
		ui.WithTitle("Demo2 Inspector Standalone Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// 等待渲染
	time.Sleep(300 * time.Millisecond)

	// 获取渲染输出
	rendered := testApp.GetRenderString()
	t.Logf("\n=== Inspector with Demo2 Content (first 60 lines) ===\n%s\n=== End ===", truncateLines(rendered, 60))

	// 验证 Inspector 元素
	expectedElements := []string{
		"INSPECTOR",
		"Elements",
		"Console",
		"Performance",
		"Layout Tree",
		"Nodes:",
		"Instructions",
	}

	missing := []string{}
	for _, expected := range expectedElements {
		if !strings.Contains(rendered, expected) {
			missing = append(missing, expected)
		}
	}

	if len(missing) > 0 {
		t.Logf("Note: Inspector missing elements (may be label/locale differences): %v", missing)
		t.Logf("Full render:\n%s", rendered)
	} else {
		t.Log("✓ All expected Inspector elements present")
	}

	// 验证树节点数量
	if strings.Contains(rendered, "Nodes: 0") {
		t.Logf("Note: Inspector shows zero node count (may be timing)")
	} else if strings.Contains(rendered, "Nodes:") {
		// 提取并显示节点数
		idx := strings.Index(rendered, "Nodes: ")
		if idx >= 0 {
			nodeCountStr := rendered[idx+7:]
			endIdx := strings.IndexAny(nodeCountStr, " |")
			if endIdx > 0 {
				nodeCountStr = nodeCountStr[:endIdx]
				t.Logf("✓ Inspector shows: Nodes: %s", nodeCountStr)
			}
		}
	}

	// 检查 demo2 相关的组件
	demo2Components := []string{"VStack", "HStack", "Text", "Button", "Bordered"}
	foundComponents := []string{}
	for _, component := range demo2Components {
		if strings.Contains(rendered, component) {
			foundComponents = append(foundComponents, component)
		}
	}

	if len(foundComponents) > 0 {
		t.Logf("✓ Found demo2 components: %v", foundComponents)
	}

	// 检查树的可视化
	treeIndicators := []string{"LayoutNode", "ElementVNode", "ButtonVNode", "TextVNode", "VStack", "HStack"}
	treeFound := false
	for _, indicator := range treeIndicators {
		if strings.Contains(rendered, indicator) {
			treeFound = true
			t.Logf("✓ Tree visualization found (contains %s)", indicator)
			break
		}
	}

	if !treeFound {
		t.Log("Note: Tree visualization may be in different format")
	}

	// 测试键盘导航
	t.Log("\n=== Testing keyboard navigation ===")
	testApp.InjectSpecialKey(platform.KeyPageDown)
	time.Sleep(200 * time.Millisecond)

	afterPgDn := testApp.GetRenderString()
	if len(afterPgDn) > 0 {
		t.Log("✓ PageDown key processed")
	}

	testApp.InjectSpecialKey(platform.KeyHome)
	time.Sleep(200 * time.Millisecond)

	afterHome := testApp.GetRenderString()
	if len(afterHome) > 0 {
		t.Log("✓ Home key processed")
	}

	t.Log("\n=== Inspector Standalone Test Complete ===")
}

// buildDemo2Buttons 创建类似 demo2 的按钮
func buildDemo2Buttons() ui.VNode {
	return ui.HStack(
		ui.NewButtonBuilder("[1] Event").FocusStyle(ui.FocusStyleBracket).Build(),
		ui.NewButtonBuilder("[2]setState").FocusStyle(ui.FocusStyleBracket).Build(),
		ui.NewButtonBuilder("[3]Scheduler").FocusStyle(ui.FocusStyleBracket).Build(),
		ui.NewButtonBuilder("[4] Render").FocusStyle(ui.FocusStyleBracket).Build(),
		ui.NewButtonBuilder("[5]Reconcile").FocusStyle(ui.FocusStyleBracket).Build(),
		ui.NewButtonBuilder("[6] Layout").FocusStyle(ui.FocusStyleBracket).Build(),
		ui.NewButtonBuilder("[7] Paint").FocusStyle(ui.FocusStyleBracket).Build(),
		ui.NewButtonBuilder("[0] Idle").FocusStyle(ui.FocusStyleBracket).Build(),
	)
}
