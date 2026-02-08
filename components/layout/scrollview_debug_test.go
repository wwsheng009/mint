package layout

import (
	"testing"
	"time"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
)

// TestScrollViewContentExtraction 测试 ScrollView 的内容提取
func TestScrollViewContentExtraction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// 创建测试内容
	testContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	textNode := ui.Text(testContent)

	// 创建 ScrollView
	builder := NewScrollView(textNode).
		Width(30).
		Height(3).
		ScrollOffset(0)

	t.Logf("ScrollView builder: %+v", builder)

	scrollView := builder.Build()

	t.Logf("ScrollView type: %T", scrollView)
	t.Logf("ScrollView props: %+v", scrollView.Props())

	// 检查 children
	children := scrollView.Children()
	t.Logf("Number of children: %d", len(children))

	for i, child := range children {
		t.Logf("Child %d: type=%T", i, child)
		if textNode, ok := child.(*rtui.TextVNode); ok {
			t.Logf("  Content length: %d", len(textNode.Content()))
			if len(textNode.Content()) < 100 {
				t.Logf("  Content: '%s'", textNode.Content())
			}
		}
	}

	testApp, err := ui.RunTest(func() ui.VNode {
		return scrollView
	},
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("ScrollView Debug"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("=== Render ===\n%s\n=== End ===", rendered)

	if len(rendered) == 0 {
		t.Fatal("Render is empty")
	}
}
