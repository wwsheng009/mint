package layout

import (
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestBasicTextRender 测试基本的 Text 渲染
func TestBasicTextRender(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	testApp, err := ui.RunTest(func() ui.VNode {
		return ui.Text("Hello World")
	},
		ui.WithWidth(30),
		ui.WithHeight(5),
		ui.WithTitle("Basic Text Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("Render output: '%s'", rendered)

	if len(rendered) == 0 {
		t.Fatal("Render output is empty")
	}
}
