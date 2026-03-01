// Test main.go centered modal behavior
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
	modal "github.com/wwsheng009/mint/ui/components/modal"
	text "github.com/wwsheng009/mint/ui/components/text"
	app "github.com/wwsheng009/mint/ui"
)

// Test 1: Centered modal from main.go (uses Stack+Spacer)
func Test1_CenteredModalFromMain() rtui.VNode {
	return app.VStack(
		app.Spacer().Flex(1).Build(),
		modal.NewBuilder().
			Key("modal-centered").
			Title("Centered Modal").
			Content(app.VStack(
				text.NewBuilder("🎯 Centered Position").
					FgColor("yellow").
					Bold(true).
					Build(),
				app.Text(""),
				text.NewBuilder("This is the default and most common").
					FgColor("gray").
					Build(),
				text.NewBuilder("positioning for modals.").
					FgColor("gray").
					Build(),
			)).
			Width(50).
			Height(14).
			Center().
			Open(true).
			Rounded().
			Build(),
		app.Spacer().Flex(1).Build(),
	)
}

// Test 2: Left aligned modal from main.go
func Test2_LeftAlignedModal() rtui.VNode {
	return app.HStack(
		modal.NewBuilder().
			Key("modal-left").
			Title("Left Aligned").
			Content(app.VStack(
				text.NewBuilder("⬅️ Left Aligned").
					FgColor("yellow").
					Bold(true).
					Build(),
				app.Text(""),
				text.NewBuilder("No spacer on the left side").
					FgColor("gray").
					Build(),
			)).
			Width(45).
			Height(13).
			Centered(false).
			Open(true).
			Rounded().
			Build(),
		app.Spacer().Flex(1).Build(),
	)
}

// Test 3: Modal without any wrapper (should flow naturally)
func Test3_ModalWithoutWrapper() rtui.VNode {
	return modal.NewBuilder().
		Key("modal-no-wrapper").
		Title("Modal Without Wrapper").
		Content(text.NewBuilder("Modal with no VStack/HStack wrapper").Build()).
		Width(30).
		Height(8).
		Single().
		Open(true).
		Build()
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Testing Modal Positioning from main.go                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	tests := []struct {
		name  string
		node  rtui.VNode
		desc  string
	}{
		{"Test 1", Test1_CenteredModalFromMain(), "VStack(Spacer, Modal, Spacer) + .Center()"},
		{"Test 2", Test2_LeftAlignedModal(), "HStack(Modal, Spacer) + .Centered(false)"},
		{"Test 3", Test3_ModalWithoutWrapper(), "Modal alone (no wrapper)"},
	}

	for _, test := range tests {
		fmt.Printf("\n%s\n%s\n", strings.Repeat("=", 80), strings.Repeat("=", 80))
		fmt.Printf("  %s\n", test.name)
		fmt.Printf("  %s\n", test.desc)
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		fwApp := framework.NewApp()
		node := render.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode { return test.node }, fwApp)
		node.SetRenderMode(render.RenderModeFiberFirst)

		buf := paint.NewBuffer(80, 30)
		ctx := component.PaintContext{
			Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 30},
			AvailableWidth:  80,
			AvailableHeight: 30,
		}

		node.Paint(ctx, buf)

		// Show render output
		fmt.Println("\nRender Output:")
		printBuffer(buf, 80, 20)

		// Show modal position
		fmt.Println("\nModal Position (Layout Phase):")
		boxes := node.GetLayoutBoxes()
		if boxes != nil {
			for _, box := range boxes {
				if box.Layer == 2 && box.Border.Style != layout.BorderNone {
					fmt.Printf("  Position: (%d, %d), Size: %dx%d\n", box.X, box.Y, box.Width, box.Height)
				}
			}
		}

		fmt.Printf("\n  ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑ ↑\n\n")
	}
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
