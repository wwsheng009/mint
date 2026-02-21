package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/stack"
	"github.com/wwsheng009/mint/ui/components/border"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/layout"
	"os"
)

func main() {
	os.Setenv("MINT_DEBUG_STACK_MEASURE", "true")

	fmt.Println("========================================")
	fmt.Println("HStack.Measure with auto-height Panel")
	fmt.Println("========================================")

	// Create Panel 2 structure (Panel = Border(VStack(flex(Text))))
	textNode := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)
	wrapped := rtui.Flex(textNode, 1)
	vbox := stack.New(stack.Column)
	vbox.SetChildrenList([]rtui.VNode{wrapped})

	border := border.New()
	border.SetChild(vbox)
	border.SetBorderLabel(" Demo ")
	border.SetWidth(18)

	// Measure border directly
	fmt.Println("\nDirect Border Measure (MaxInt constraints):")
	borderInst := border.CreateInstance()
	if measurable, ok := borderInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		constraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  layout.MaxInt,
			MinHeight: 0,
			MaxHeight: layout.MaxInt,
		}
		size := measurable.Measure(constraints)
		fmt.Printf("  Border measured: %dx%d\n", size.Width, size.Height)
	}

	// Panel 1: fixed height
	panel1 := stack.New(stack.Row)
	panel1.SetChildrenList([]rtui.VNode{vbox})
	panel1Inst := panel1.CreateInstance()
	if h, ok := panel1Inst.(interface{ SetWidth(int) }); ok {
		h.SetWidth(20)
	}
	if h, ok := panel1Inst.(interface{ SetHeight(int) }); ok {
		h.SetHeight(3)
	}

	// Row with both panels
	row := stack.New(stack.Row)
	row.SetGap(3)
	row.SetChildrenList([]rtui.VNode{panel1, border})

	fmt.Println("\nRow.Measure (Row auto-height, Panel1 fixed=3, Panel2 auto):")
	rowInst := row.CreateInstance()
	if measurable, ok := rowInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		constraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  50,
			MinHeight: 0,
			MaxHeight: layout.MaxInt,  // Row auto-height
		}
		size := measurable.Measure(constraints)
		fmt.Printf("  Row measured: %dx%d\n", size.Width, size.Height)
	}
}
