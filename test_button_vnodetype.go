package main

import (
	"fmt"
	"github.com/wwsheng009/mint/app"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
	btn := app.ButtonBuilder("[1] Event").Build()
	
	fmt.Printf("Button type: %T\n", btn)
	fmt.Printf("Button Type(): %v\n", btn.Type())
	fmt.Printf("VNodeTypeComponent: %v\n", rtui.VNodeTypeComponent)
	fmt.Printf("VNodeElement: %v\n", rtui.VNodeElement)
	fmt.Printf("Is Component? %v\n", btn.Type() == rtui.VNodeTypeComponent)
	
	// Check if it implements the ComponentVNode interface
	if _, ok := btn.(*rtui.ComponentVNode); ok {
		fmt.Println("Button IS a ComponentVNode ❌")
	} else {
		fmt.Println("Button is NOT a ComponentVNode ✅")
	}
}
