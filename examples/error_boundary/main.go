// examples/error_boundary/main.go - Error boundary demo application
package main

import (
	"flag"
	"log"

	"github.com/wwsheng009/mint/ui"
)

// shouldPanic controls whether the component will panic
var shouldPanic = false

var shouldPanicFlag = flag.Bool("panic", false, "Set to true to trigger a panic in the component")

// ErrorBoundaryApp creates the demo application
func ErrorBoundaryApp() ui.VNode {
	return ui.VStack(
		ui.Element("text").Prop("content", "Error Boundary Demo").Build(),
		ui.ErrorBoundary(
			"demoBoundary",
			PanicComponent,
			ui.FallbackBox("Error Occurred", "Something went wrong in this component"),
		),
	)
}

// PanicComponent renders text or panics based on shouldPanic flag
func PanicComponent() ui.VNode {
	if shouldPanic {
		panic("intentional panic for demo")
	}
	return ui.VStack(
		ui.Element("text").Prop("content", "Normal Content").Build(),
		ui.Element("text").Prop("content", "This component renders successfully").Build(),
		ui.Element("text").Prop("content", "").Build(),
		ui.Element("text").Prop("content", "Run with -panic flag to see error handling").Build(),
	)
}

func main() {
	flag.Parse()

	if *shouldPanicFlag {
		shouldPanic = true
	}

	if err := ui.Run(ErrorBoundaryApp); err != nil {
		log.Fatal(err)
	}
}
