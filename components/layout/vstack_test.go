package layout

import (
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestVStackWithChildren 测试 VStack 是否渲染 children
func TestVStackWithChildren(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(func() ui.VNode {
		return ui.VStack(
			ui.Text("Line 1"),
			ui.Text("Line 2"),
			ui.Text("Line 3"),
		)
	},
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("VStack Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("=== VStack Render ===\n%s\n=== End ===", rendered)

	if len(rendered) == 0 {
		t.Fatal("VStack render output is empty")
	}
}
