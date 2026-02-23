package render

import (
	"fmt"
	"os"
	"testing"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/ui"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// TestDeclarativeNode_Paint_Chinese tests that Chinese characters are correctly rendered to buffer
func TestDeclarativeNode_Paint_Chinese(t *testing.T) {
	// Enable Fiber-first mode
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	tests := []struct {
		name       string
		content    string
		wantWidth  int
		wantHeight int
	}{
		{"ASCII only", "Hello", 5, 1},
		{"Chinese only", "你好世界", 8, 1},
		{"Mixed content", "Hi你好World", 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create render function
			renderFn := func() ui.VNode {
				return newstack.NewVStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						newtext.New(tt.content),
					})
			}

			// Create framework App
			fwApp := framework.NewApp()
			fwApp.Resize(60, 10)

			// Create DeclarativeNode with Fiber
			declarativeNode := NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
			fwApp.SetRoot(declarativeNode)

			// Create buffer
			buf := paint.NewBuffer(60, 10)

			// Create paint context
			ctx := component.PaintContext{
				AvailableWidth:  60,
				AvailableHeight: 10,
				Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 10},
			}

			// Call Paint
			declarativeNode.Paint(ctx, buf)

			// Debug: Print buffer content
			t.Logf("=== Buffer Content for %q ===", tt.content)
			printBufferContent(t, buf, 10)

			// Check if content was written to buffer
			found := false
			for y := 0; y < buf.Height; y++ {
				for x := 0; x < buf.Width; x++ {
					cell := buf.GetContent(x, y)
					if cell.Cluster != "" && cell.Cluster != "\x00" && cell.Cluster != " " {
						t.Logf("Found content at (%d,%d): Cluster=%q, Width=%d", x, y, cell.Cluster, cell.Width)
						found = true
					}
				}
			}

			if !found {
				t.Errorf("No content found in buffer for %q", tt.content)
			}
		})
	}
}

// TestDeclarativeNode_Paint_BufferDebug tests the full paint pipeline and outputs buffer details
func TestDeclarativeNode_Paint_BufferDebug(t *testing.T) {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
	os.Setenv("MINT_DEBUG_TEST", "true")

	content := "测试中文"

	renderFn := func() ui.VNode {
		return newstack.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				newtext.New("ASCII: Hello"),
				newtext.New("中文: " + content),
				newtext.New("混合: Test测试"),
			})
	}

	fwApp := framework.NewApp()
	fwApp.Resize(60, 10)

	declarativeNode := NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
	fwApp.SetRoot(declarativeNode)

	buf := paint.NewBuffer(60, 10)
	ctx := component.PaintContext{
		AvailableWidth:  60,
		AvailableHeight: 10,
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 10},
	}

	declarativeNode.Paint(ctx, buf)

	t.Log("=== Full Buffer String Output ===")
	t.Log(buf.String())

	t.Log("=== Buffer Cells Detail ===")
	for y := 0; y < buf.Height; y++ {
		rowContent := ""
		for x := 0; x < buf.Width; x++ {
			cell := buf.GetContent(x, y)
			if cell.Cluster != "" && cell.Cluster != "\x00" {
				rowContent += fmt.Sprintf("[%q:%d]", cell.Cluster, cell.Width)
			}
		}
		if rowContent != "" {
			t.Logf("Row %d: %s", y, rowContent)
		}
	}
}

// printBufferContent prints the content of a buffer for debugging
func printBufferContent(t *testing.T, buf *paint.Buffer, maxRows int) {
	t.Helper()
	for y := 0; y < buf.Height && y < maxRows; y++ {
		var row string
		for x := 0; x < buf.Width; x++ {
			cell := buf.GetContent(x, y)
			if cell.IsContinuation {
				row += "·" // continuation marker
			} else if cell.Cluster != "" && cell.Cluster != "\x00" {
				row += cell.Cluster
			} else {
				row += " "
			}
		}
		t.Logf("Row %d: |%s|", y, row)
	}
}
