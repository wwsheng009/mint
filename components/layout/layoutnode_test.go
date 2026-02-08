package layout

import (
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestLayoutNodeDirect 测试 LayoutNode 是否能渲染 children
func TestLayoutNodeDirect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建一个简单的 LayoutNode
	layoutNode := ui.VStack(
		ui.Text("Line A"),
		ui.Text("Line B"),
		ui.Text("Line C"),
	)

	t.Logf("LayoutNode type: %T", layoutNode)

	testApp, err := ui.RunTest(func() ui.VNode {
		return layoutNode
	},
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("LayoutNode Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("=== LayoutNode Render ===\n%s\n=== End ===", rendered)

	if len(rendered) == 0 {
		t.Fatal("LayoutNode render is empty")
	}
}
