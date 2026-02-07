package main

import (
	ui "github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(Test,
		ui.WithWidth(80),
		ui.WithHeight(10),
		ui.WithTitle("Test Trailing Spaces"),
	)
	if err != nil {
		panic(err)
	}
}

func Test() ui.VNode {
	return ui.VStack(
		ui.Text("Testing trailing spaces:"),
		ui.Text(""),
		ui.Text("Short text..................................."),  // 51 chars
		ui.Text("Short text....................."),           // 37 chars
		ui.Text("[ Btn ]                                    "),  // Button with trailing spaces
		ui.Text(""),
		ui.Text("Count the spaces above!"),
	)
}
