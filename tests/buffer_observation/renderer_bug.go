//go:build ignore
// +build ignore

// Renderer Bug Test - Full Render Path
//
// Purpose: Reproduce bug using full Renderer.Render() path with diff
// This tests the actual bug scenario from store_mixed_demo
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/tests/buffer_observation/utils"
	"github.com/wwsheng009/mint/ui"
)

type AppState struct {
	Count int
}

var appStore = store.NewStore(AppState{Count: 0})

func ExpanderComponent() ui.VNode {
	count := appStore.Get().Count
	expanded := count%2 == 0

	items := []ui.VNode{
		ui.NewTextBuilder("=== 1. Store 订阅状态 ===").
			FgColor("green").
			Build(),
		ui.Text(fmt.Sprintf("  折叠状态: %s (基于 Count) ", map[bool]string{
			true:  "展开",
			false: "折叠",
		}[expanded])),
	}

	if expanded {
		items = append(items,
			ui.Text("  这是一个基于 Store 的状态"),
			ui.Text("  可以跨组件访问"),
			ui.Text("  数据流：Intent → Reducer → Store → UI"),
		)
	}

	return ui.VStack(items...)
}

func CounterComponent() ui.VNode {
	count := appStore.Get().Count
	return ui.VStack(
		ui.NewTextBuilder("=== 2. Store 计数器 ===").
			FgColor("green").
			Build(),
		ui.HStack(
			ui.Text("计数: "),
			ui.NewTextBuilder(fmt.Sprintf("%d", count)).
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" + ").
				Variant(ui.ButtonVariantPrimary).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" - ").
				Variant(ui.ButtonVariantSecondary).
				Build(),
		),
	)
}

func App() ui.VNode {
	return ui.VStack(
		ExpanderComponent(),
		ui.Text(""),
		CounterComponent(),
	)
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Renderer Bug Test (Full Render Path)                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()

	// 创建Renderer (真实路径)
	renderer := paint.NewRenderer(80, 25)
	renderer.ForceFullRender() // 首次渲染强制全量

	// Paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 25},
		AvailableWidth:  80,
		AvailableHeight: 25,
	}

	// Create declarative node
	node := render.NewDeclarativeNodeFromFuncWithFiber(App)
	node.SetApp(fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	var frame1Content, frame2Content []string

	// Frame 1: count=0
	{
		appStore.Set(AppState{Count: 0})
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Println("FRAME 1: count=0, Expander expanded")
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		// Get back buffer
		backBuf := renderer.GetBackBuffer()
		backBuf.Reset(80, 25)

		// Paint to back buffer (真实路径)
		node.Paint(ctx, backBuf)

		fmt.Println("Back Buffer content:")
		utils.PrintBuffer(backBuf, 80, 25)

		// Render (通过Diff输出)
		renderOutput := renderer.Render()
		fmt.Printf("\nRender output (first 500 chars):\n%s\n---\n", truncate(renderOutput, 500))
		utils.SaveBuffer(backBuf, 80, 25)
		os.Rename("output.txt", "frame1.txt")

		// 保存front buffer用于比较
		frame1Front := renderer.GetFrontBuffer()
		frame1Content = make([]string, 10)
		for y := 0; y < 10; y++ {
			frame1Content[y] = extractRow(frame1Front, y, 80)
		}
	}

	// Frame 2: count=1
	{
		appStore.Set(AppState{Count: 1})
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Println("FRAME 2: count=1, Expander collapsed")
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		// Get back buffer
		backBuf := renderer.GetBackBuffer()
		backBuf.Reset(80, 25)

		// Paint to back buffer
		node.Paint(ctx, backBuf)

		// Render (通过Diff输出)
		renderer.Render()

		fmt.Println("Back Buffer content:")
		utils.PrintBuffer(backBuf, 80, 25)

		fmt.Printf("\nRender completed (see output in frame2.txt)\n")
		utils.SaveBuffer(backBuf, 80, 25)
		os.Rename("output.txt", "frame2.txt")

		// 保存front buffer
		frame2Front := renderer.GetFrontBuffer()
		frame2Content = make([]string, 10)
		for y := 0; y < 10; y++ {
			frame2Content[y] = extractRow(frame2Front, y, 80)
		}

		// 比较
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Println("ANALYSIS: Comparing Front Buffers")
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		buttonLine1 := findCounterLine(frame1Content)
		buttonLine2 := findCounterLine(frame2Content)
		fmt.Printf("Button moved from line %d to %d\n", buttonLine1, buttonLine2)

		if buttonLine1 >= 0 && buttonLine2 >= 0 && buttonLine1 != buttonLine2 {
			fmt.Printf("\nChecking old position (line %d) in frame 2:\n", buttonLine1)
			fmt.Printf("  Frame 1: %s\n", frame1Content[buttonLine1])
			fmt.Printf("  Frame 2: %s\n", frame2Content[buttonLine1])

			// 检查是否残留按钮字符
			hasButton := strings.Contains(frame2Content[buttonLine1], "[") ||
				strings.Contains(frame2Content[buttonLine1], "]") ||
				strings.Contains(frame2Content[buttonLine1], "+") ||
				strings.Contains(frame2Content[buttonLine1], "*")

			if hasButton {
				fmt.Printf("\n❌ BUG DETECTED: Button characters remain at old position!\n")
			} else {
				fmt.Printf("\n✅ PASS: No button characters at old position\n")
			}
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("Check frame1.txt and frame2.txt for detailed content")
	fmt.Printf(strings.Repeat("=", 80))
}

func extractRow(buffer *paint.Buffer, row, width int) string {
	var sb strings.Builder
	for x := 0; x < width; x++ {
		cell := buffer.GetContent(x, row)
		if cell.IsContinuation {
			continue
		}
		if cell.Cluster == "" || cell.Cluster == " " {
			sb.WriteRune(' ')
		} else {
			sb.WriteString(cell.Cluster)
		}
	}
	return sb.String()
}

func findCounterLine(rows []string) int {
	for y, row := range rows {
		if strings.Contains(row, "计数:") {
			return y
		}
	}
	return -1
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
