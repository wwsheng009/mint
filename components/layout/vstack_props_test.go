package layout

import (
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestVStackWithProps 测试带有 Width/Height 的 VStack
func TestVStackWithProps(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(func() ui.VNode {
		return ui.VStackBuilder(
			ui.Text("Test Line 1"),
			ui.Text("Test Line 2"),
		).
			Width(30).
			Height(5).
			Build()
	},
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("VStack Props Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("=== VStack with Props Render ===\n%s\n=== End ===", rendered)

	if len(rendered) == 0 {
		t.Fatal("VStack with Width/Height renders nothing")
	}

	// 简单检查
	t.Logf("Render length: %d chars", len(rendered))
}
