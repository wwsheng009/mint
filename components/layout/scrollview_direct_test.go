package layout

import (
	"strings"
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestScrollViewDirect 测试 ScrollView 直接返回的 VNode
func TestScrollViewDirect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建长文本
	content := strings.Repeat("Line of text\n", 20)

	// 创建 ScrollView
	scrollView := NewScrollView(ui.Text(content)).
		Width(40).
		Height(10).
		Build()

	t.Logf("ScrollView type: %T", scrollView)

	// 测试 ScrollView 本身
	testApp, err := ui.RunTest(func() ui.VNode {
		return scrollView
	},
		ui.WithWidth(50),
		ui.WithHeight(15),
		ui.WithTitle("ScrollView Direct"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("=== ScrollView Direct Render ===\n%s\n=== End ===", rendered)

	if len(rendered) == 0 {
		t.Fatal("ScrollView render is completely empty")
	}

	// 检查是否包含任何内容
	if strings.Contains(rendered, "Line of text") {
		t.Log("✓ ScrollView renders content")
	} else {
		t.Error("ScrollView should render content")
		t.Logf("Got: %s", rendered)
	}
}
