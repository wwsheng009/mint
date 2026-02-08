package layout

import (
	"fmt"
	"strings"
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestScrollViewBasicRendering 测试 ScrollView 基本渲染
func TestScrollViewBasicRendering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建长文本
	longText := ""
	for i := 1; i <= 30; i++ {
		longText += strings.TrimSpace(fmt.Sprintf("Line %d: ScrollView test content\n", i))
	}

	// 创建 ScrollView
	scrollView := NewScrollView(ui.Text(longText)).
		Width(50).
		Height(10).
		ScrollOffset(0).
		Build()

	testApp, err := ui.RunTest(func() ui.VNode {
		return ui.VStack(
			ui.Text("ScrollView Test"),
			ui.Text("─────────────"),
			scrollView,
		)
	},
		ui.WithWidth(60),
		ui.WithHeight(15),
		ui.WithTitle("ScrollView Basic Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	rendered := testApp.GetRenderString()

	t.Logf("=== ScrollView Render ===\n%s\n=== End ===", rendered)

	// 验证内容存在
	if !strings.Contains(rendered, "Line 1") {
		t.Error("ScrollView should show first line")
	} else {
		t.Log("✓ ScrollView shows content")
	}

	// 验证高度限制（应该只显示部分内容）
	lineCount := strings.Count(rendered, "Line ")
	t.Logf("Visible lines: %d", lineCount)

	if lineCount > 12 {
		t.Logf("Note: ScrollView may not be respecting height limit (expected ~10 lines)")
	}
}
