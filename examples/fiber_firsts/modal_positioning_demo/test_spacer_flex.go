// Direct Flex Spacer Test
// Tests if Spacer flex attribute works correctly in Fiber-First mode
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Direct Flex Spacer Test                                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	tests := []struct {
		name  string
		node  rtui.VNode
		desc  string
	}{
		{
			name: "Test 1",
			node: rtui.VStack(
				rtui.Spacer().Flex(1).Build(),
				newtext.New("Centered Text"),
				rtui.Spacer().Flex(1).Build(),
			),
			desc: "VStack(Spacer(1), Text, Spacer(1)) - Text should be centered",
		},
		{
			name: "Test 2",
			node: rtui.HStack(
				newtext.New("Left"),
				rtui.Spacer().Flex(1).Build(),
				newtext.New("Right"),
			),
			desc: "HStack(Text, Spacer(1), Text) - Spacer pushes Text to edges",
		},
		{
			name: "Test 3",
			node: rtui.HStack(
				rtui.Spacer().Flex(1).Build(),
				newtext.New("Centered"),
				rtui.Spacer().Flex(1).Build(),
			),
			desc: "HStack(Spacer(1), Text, Spacer(1)) - Text should be centered",
		},
	}

	for _, test := range tests {
		fmt.Printf("\n=== %s ===\n", test.name)
		fmt.Printf("%s\n\n", test.desc)

		fwApp := framework.NewApp()
		node := render.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode { return test.node }, fwApp)
		node.SetRenderMode(render.RenderModeFiberFirst)

		buf := paint.NewBuffer(60, 15)
		ctx := component.PaintContext{
			Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 15},
			AvailableWidth:  60,
			AvailableHeight: 15,
		}

		node.Paint(ctx, buf)

		// Show output
		for y := 0; y < 15; y++ {
			line := "|"
			for x := 0; x < 60; x++ {
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

		// Check layout boxes
		fmt.Println("\nLayout Boxes:")
		boxes := node.GetLayoutBoxes()
		if boxes != nil {
			// Find spacer fibers
			for _, box := range boxes {
				fmt.Printf("  [%s] Pos:(%d,%d) Size:%dx%d\n", box.ID, box.X, box.Y, box.Width, box.Height)
			}
		}
		fmt.Println()
	}

	fmt.Println("\n=== Expected Results ===")
	fmt.Println("  Test 1: 'Centered Text' should be vertically centered")
	fmt.Println("  Test 2: 'Left' on left, 'Right' on right (spacer between them)")
	fmt.Println("  Test 3: 'Centered' should be horizontally centered")
	fmt.Println("\n=== Test Complete ===")
}
