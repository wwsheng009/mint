package main

import (
	"fmt"
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/components/layout"
)

func main() {
	btn := app.ButtonBuilder("[1] Event").Build()
	
	fmt.Printf("Button type: %T\n", btn)
	fmt.Printf("Button Type(): %v\n", btn.Type())
	if tagger, ok := btn.(interface{ Tag() string }); ok {
		fmt.Printf("Button Tag(): %s\n", tagger.Tag())
	}
	fmt.Println()
	
	// Wrap it
	wrapped := layout.NewWrapBuilder(btn).
		Gap(1).
		ScreenWidth(78).
		FillWidth().
		Build()
		
	fmt.Printf("Wrapped type: %T\n", wrapped)
	
	children := wrapped.Children()
	fmt.Printf("Wrapped children count: %d\n", len(children))
	
	if len(children) > 0 {
		row1 := children[0]
		fmt.Printf("Row1 type: %T\n", row1)
		fmt.Printf("Row1 Type(): %v\n", row1.Type())
		if tagger, ok := row1.(interface{ Tag() string }); ok {
			fmt.Printf("Row1 Tag(): %s\n", tagger.Tag())
		}
		
		row1Children := row1.Children()
		fmt.Printf("Row1 children count: %d\n", len(row1Children))
		
		if len(row1Children) > 0 {
			child := row1Children[0]
			fmt.Printf("Child type: %T\n", child)
			fmt.Printf("Child Type(): %v\n", child.Type())
			if tagger, ok := child.(interface{ Tag() string }); ok {
				fmt.Printf("Child Tag(): %s\n", tagger.Tag())
			}
			
			props := child.Props()
			if props != nil {
				if flex, ok := props["flex"].(int); ok {
					fmt.Printf("Child flex prop: %d ✅\n", flex)
				} else {
					fmt.Printf("Child flex prop: NOT SET ❌\n")
				}
			}
		}
	}
}
