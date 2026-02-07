package main

import (
	"fmt"
	"github.com/wwsheng009/mint/app"
)

func main() {
	btn := app.ButtonBuilder("[1] Event").Build()
	
	fmt.Printf("Button type: %T\n", btn)
	fmt.Printf("Button Children count: %d\n", len(btn.Children()))
	
	children := btn.Children()
	for i, child := range children {
		fmt.Printf("Child %d: type=%T, Type()=%v\n", i, child, child.Type())
		if tagger, ok := child.(interface{ Tag() string }); ok {
			fmt.Printf("  Tag: %s\n", tagger.Tag())
		}
	}
}
