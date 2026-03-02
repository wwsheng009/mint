package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// Hello demonstrates controlled input with real-time updates
func Hello() ui.VNode {

	return app.VStack(
		ui.Text("Hello World"),
	)
}

func main() {
	err := ui.Run(Hello,
		ui.WithWidth(50),
		ui.WithHeight(18),
		ui.WithTitle("Input Demo"),
	)
	if err != nil {
		panic(err)
	}
}
