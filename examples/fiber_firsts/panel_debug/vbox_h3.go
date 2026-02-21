package main

import (
	"fmt"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/stack"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
	textNode := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)
	wrapped := rtui.Flex(textNode, 1)
	vbox := stack.New(stack.Column)
	vbox.SetChildrenList([]rtui.VNode{wrapped})

	constraints := layout.Constraints{MinWidth:0, MaxWidth:18, MinHeight:0, MaxHeight:3}
	vb := vbox.CreateInstance()

	if vbInst, ok := vb.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		size := vbInst.Measure(constraints)
		fmt.Printf("Constraints: MaxHeight=%d\n", constraints.MaxHeight)
		fmt.Printf("VStack.Measured: %dx%d\n", size.Width, size.Height)
	}
}
