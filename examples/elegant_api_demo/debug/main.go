package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
)

// ElegantAPIDemo - demo component for testing margin
func ElegantAPIDemo() rtui.VNode {
	return rtui.VStackBuilder(
		ui.Text("✨ Margin Debug - All Directions"),
		ui.Text("───────────────────────────────────────────────"),

		// 测试 1: MarginV - 垂直方向 (top/bottom)
		ui.Text("Test 1: MarginV(top, bottom)"),
		ui.Text("  Btn1: MarginV(1, 0) [top=1, bottom=0]"),
		ui.NewButtonBuilder("Btn1").SetID("Btn1").MarginV(1, 0).Build(),
		ui.Text("  Btn2: MarginV(0, 1) [top=0, bottom=1]"),
		ui.NewButtonBuilder("Btn2").SetID("Btn2").MarginV(0, 1).Build(),
		ui.Text("  Btn3: MarginV(1, 1) [top=1, bottom=1]"),
		ui.NewButtonBuilder("Btn3").SetID("Btn3").MarginV(1, 1).Build(),

		// 测试 2: MarginH - 水平方向 (left/right)
		ui.Text("Test 2: MarginH(left, right) [in HStack]"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("L").SetID("L").MarginH(1, 0).Build(),  // left=1, right=0
			ui.NewButtonBuilder("C").SetID("C").MarginH(0, 0).Build(),  // left=0, right=0
			ui.NewButtonBuilder("R").SetID("R").MarginH(0, 1).Build(),  // left=0, right=1
		).Gap(1).Build(),

		// 测试 3: Margin - 四个方向
		ui.Text("Test 3: Margin(top, right, bottom, left)"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("T").SetID("T").Margin(1, 0, 0, 0).Build(),  // top=1
			ui.NewButtonBuilder("R").SetID("R").Margin(0, 1, 0, 0).Build(),  // right=1
			ui.NewButtonBuilder("B").SetID("B").Margin(0, 0, 1, 0).Build(),  // bottom=1
			ui.NewButtonBuilder("L").SetID("L").Margin(0, 0, 0, 1).Build(),  // left=1
			ui.NewButtonBuilder("A").SetID("A").Margin(1, 1, 1, 1).Build(),  // all=1
		).Gap(1).Build(),

		// 测试 4: MarginAll - 所有方向
		ui.Text("Test 4: MarginAll(value)"),
		ui.NewButtonBuilder("Margin1").SetID("Margin1").MarginAll(1).Build(),
		ui.NewButtonBuilder("Margin2").SetID("Margin2").MarginAll(2).Build(),
	).Gap(0).Build()
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Margin Debug Test                                            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(ElegantAPIDemo)
    node.SetApp(fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())
	fmt.Printf("  Fiber-First Enabled: %v\n", node.IsFiberFirstEnabled())

	// Create buffer
	buf := paint.NewBuffer(80, 25)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 25},
		AvailableWidth:  80,
		AvailableHeight: 25,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("Rendering with margin support...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	// Render
	node.Paint(ctx, buf)

	fmt.Println("\n=== Render Output ===")
	printBuffer(buf, 80, 25)

		// Get layout boxes (用于详细分析)
	var boxes []*layout.LayoutBox
	nodeBoxes := node.GetLayoutBoxes()
	if nodeBoxes != nil {
		boxes = nodeBoxes
	}

	fmt.Println("\n=== Expected Margin Results ===")
	fmt.Println("Test 1 - MarginV (Vertical):")
	fmt.Println("  Btn1: MarginV(1,0) → top=1, bottom=0")
	fmt.Println("  Btn2: MarginV(0,1) → top=0, bottom=1")
	fmt.Println("  Btn3: MarginV(1,1) → top=1, bottom=1")
	fmt.Println("  Expected: Btn1 top margin, Btn2 bottom margin, Btn3 both margins")

	fmt.Println("\nTest 2 - MarginH (Horizontal in HStack):")
	fmt.Println("  L: MarginH(1,0) → left=1, right=0")
	fmt.Println("  C: MarginH(0,0) → left=0, right=0")
	fmt.Println("  R: MarginH(0,1) → left=0, right=1")
	fmt.Println("  Expected: L has left spacing, R has right spacing")

	fmt.Println("\nTest 3 - Margin (Four directions):")
	fmt.Println("  T: Margin(1,0,0,0) → top=1, others=0")
	fmt.Println("  R: Margin(0,1,0,0) → right=1, others=0")
	fmt.Println("  B: Margin(0,0,1,0) → bottom=1, others=0")
	fmt.Println("  L: Margin(0,0,0,1) → left=1, others=0")
	fmt.Println("  A: Margin(1,1,1,1) → all=1")

	fmt.Println("\nTest 4 - MarginAll:")
	fmt.Println("  Margin1: MarginAll(1) → all directions = 1")
	fmt.Println("  Margin2: MarginAll(2) → all directions = 2")

	// 打印详细的 layout box 信息
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Detailed Layout Box Analysis")
	fmt.Println(strings.Repeat("=", 80))
	printDetailedLayoutBoxes(boxes)

	// 打印期望的 margin 值到实际位置的对照
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Margin Validation Report")
	fmt.Println(strings.Repeat("=", 80))
	printMarginValidation(boxes)

	fmt.Println("\n=== Test Complete ===")
}

// 打印详细的 layout box 信息
func printDetailedLayoutBoxes(boxes []*layout.LayoutBox) {
	if boxes == nil {
		fmt.Println("  No layout boxes found!")
		return
	}

	// 定义每个测试组的按钮名称
	buttonGroups := map[string][]string{
		"Test 1 - MarginV (Btn1, Btn2, Btn3)": {"Btn1", "Btn2", "Btn3"},
		"Test 2 - MarginH (L, C, R)":         {"L", "C", "R"},
		"Test 3 - Margin (T, R, B, L, A)":     {"T", "B", "A"},
		"Test 4 - MarginAll (Margin1, Margin2)": {"Margin1", "Margin2"},
	}

	for groupName, names := range buttonGroups {
		if len(names) == 0 {
			continue
		}

		// 查找对应的 boxes
		var groupBoxes []*layout.LayoutBox
		for _, box := range boxes {
			for _, name := range names {
				if strings.Contains(box.ID, name) ||
				   strings.Contains(strings.ToLower(box.ID), strings.ToLower(name)) {
					groupBoxes = append(groupBoxes, box)
					break
				}
			}
		}

		if len(groupBoxes) == 0 {
			continue
		}

		fmt.Printf("\n[ %s ]\n", groupName)
		for _, box := range groupBoxes {
			fmt.Printf("  Node: %-12s | Pos: (%3d,%3d) | Size: %dx%d | Layer: %d\n",
				box.ID, box.X, box.Y, box.Width, box.Height, box.Layer)
		}
	}
}

// 打印 margin 验证报告
func printMarginValidation(boxes []*layout.LayoutBox) {
	if boxes == nil {
		fmt.Println("  No layout boxes to validate!")
		return
	}

	// 定义每个按钮期望的 margin
	expectedMargins := map[string][4]int{ // top, right, bottom, left
		"Btn1":    {1, 0, 0, 0},
		"Btn2":    {0, 0, 1, 0},
		"Btn3":    {1, 0, 1, 0},
		"L":       {0, 0, 0, 1},
		"C":       {0, 0, 0, 0},
		"R":       {0, 1, 0, 0},
		"T":       {1, 0, 0, 0},
		"B":       {0, 0, 1, 0},
		"A":       {1, 1, 1, 1},
		"Margin1": {1, 1, 1, 1},
		"Margin2": {2, 2, 2, 2},
	}

	// 按类型分组查找按钮
	buttonMap := make(map[string][]*layout.LayoutBox)

	for _, box := range boxes {
		// 使用 PropsID（如果存在）作为按钮标识
		id := box.PropsID
		if id == "" {
			id = box.ID  // 回退到数字 ID
		}
		lowerID := strings.ToLower(id)

		// 检查是否是按钮
		if strings.Contains(lowerID, "button") || strings.Contains(lowerID, "btn") ||
		   strings.Contains(id, "L") || strings.Contains(id, "C") || strings.Contains(id, "R") ||
		   strings.Contains(id, "T") || strings.Contains(id, "B") || strings.Contains(id, "A") ||
		   strings.Contains(id, "Margin") {

			// 使用 PropsID 作为 name（如果设置的话）
			if box.PropsID != "" {
				buttonMap[box.PropsID] = append(buttonMap[box.PropsID], box)
			} else {
				// 回退到提取名称
				name := extractButtonName(id)
				buttonMap[name] = append(buttonMap[name], box)
			}
		}
	}

	// 验证每个按钮
	fmt.Println("\nMargin Settings Validation:")
	fmt.Println("  Note: Actual layout positions incorporate margins")
	fmt.Println("")

	buttonNames := []string{"Btn1", "Btn2", "Btn3", "L", "C", "R", "T", "B", "A", "Margin1", "Margin2"}

	for _, name := range buttonNames {
		boxes, exists := buttonMap[name]
		if !exists || len(boxes) == 0 {
			continue
		}

		expected, hasExpected := expectedMargins[name]

		for i, box := range boxes {
			suffix := ""
			if i > 0 {
				suffix = fmt.Sprintf(" (#%d)", i+1)
			}
			fmt.Printf("  %-20s | Pos: (%3d,%3d) | Size: %dx%d",
				name+suffix, box.X, box.Y, box.Width, box.Height)

			if hasExpected {
				fmt.Printf(" | Exp margin: [T:%d R:%d B:%d L:%d]",
					expected[0], expected[1], expected[2], expected[3])
			}
			fmt.Println()
		}
	}

	// 检查间距是否合理
	fmt.Println("\nSpacing Analysis:")
	checkSpacing(buttonMap, "Btn1", "Btn2", "Vertical")
	checkSpacing(buttonMap, "Btn2", "Btn3", "Vertical")
}

func extractButtonName(id string) string {
	id = strings.TrimSpace(id)

	// 移除常见前缀并保留数字
	replacements := []struct {
		prefix string
		name   string
	}{
		{"Button-1", "Btn1"},
		{"button-1", "Btn1"},
		{"Button-2", "Btn2"},
		{"button-2", "Btn2"},
		{"Button-3", "Btn3"},
		{"button-3", "Btn3"},
		{"Btn", "Btn"},
	}

	for _, repl := range replacements {
		if id == repl.prefix {
			return repl.name
		}
		if strings.HasPrefix(id, repl.prefix) {
			return id // 保留完整的名称（如 "Button-1-xxx"）
		}
	}

	// 特殊处理
	if id == "1" || strings.HasPrefix(id, "1-") {
		return "Btn1"
	}
	if id == "2" || strings.HasPrefix(id, "2-") {
		return "Btn2"
	}
	if id == "3" || strings.HasPrefix(id, "3-") {
		return "Btn3"
	}
	if strings.Contains(id, "Margin") {
		return id
	}

	return id
}

func checkSpacing(buttonMap map[string][]*layout.LayoutBox, name1, name2, direction string) {
	boxes1, ok1 := buttonMap[name1]
	boxes2, ok2 := buttonMap[name2]
	if !ok1 || !ok2 || len(boxes1) == 0 || len(boxes2) == 0 {
		fmt.Printf("  %s → %s:  ⚠ Buttons not found\n", name1, name2)
		return
	}

	box1 := boxes1[0]
	box2 := boxes2[0]

	actualGap := 0
	if direction == "Vertical" {
		gapStart := box1.Y + box1.Height
		gapEnd := box2.Y
		actualGap = gapEnd - gapStart
	}

	var expectedMsg string
	if actualGap >= 2 {
		expectedMsg = "✓ Spacing detected"
	} else if actualGap >= 1 {
		expectedMsg = "~ Minimal spacing"
	} else {
		expectedMsg = "✗ No spacing"
	}

	fmt.Printf("  %s → %s: gap = %2d cells  %s\n", name1, name2, actualGap, expectedMsg)
}

func printBuffer(buf *paint.Buffer, w, h int) {
	fmt.Println("┌" + strings.Repeat("─", w) + "┐")
	for y := 0; y < h; y++ {
		line := "|"
		for x := 0; x < w; x++ {
			if y < len(buf.Cells) && x < len(buf.Cells[y]) {
				cell := buf.Cells[y][x]
				if len(cell.Cluster) == 0 || cell.Cluster == " " {
					line += " "
				} else {
					for _, r := range cell.Cluster {
						line += string(r)
						break
					}
				}
			} else {
				line += " "
			}
		}
		line += "|"
		fmt.Println(line)
	}
	fmt.Println("└" + strings.Repeat("─", w) + "┘")
}

