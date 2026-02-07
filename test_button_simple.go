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
	
	// Check if it implements the ComponentVNode interface
	if comp, ok := btn.(*rtui.ComponentVNode); ok {
		fmt.Printf("Button IS a ComponentVNode ❌: %+v\n", comp)
	} else {
		fmt.Println("Button is NOT a ComponentVNode ✅")
	}
}
