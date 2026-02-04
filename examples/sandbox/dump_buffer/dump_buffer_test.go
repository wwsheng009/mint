// dump_buffer - A helper test to dump TUI app buffer to text for verification
//
// This is a sandbox test that captures the app output and saves it to a file.
// Run with: go test -v ./examples/sandbox/dump_buffer
//
// Output files:
//   - buffer_output.txt  - Plain text render (no colors)

package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// DemoApp is a simple app for testing buffer dumping
func DemoApp() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	input, setInput := ui.UseStateString("")

	countText := fmt.Sprintf("%d", count)

	return ui.VStack(
		app.NewTextBuilder("=== Buffer Dump Test ===").
			Bold(true).
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("").
			Build(),
		ui.HStack(
			app.NewTextBuilder("Count: ").
				FgColor("green").
				Build(),
			app.NewTextBuilder(countText).
				Bold(true).
				Build(),
		),
		app.NewTextBuilder("").
			Build(),
		ui.HStack(
			app.NewTextBuilder("Input: ").
				FgColor("green").
				Build(),
			app.InputBuilder().
				Value(input).
				Placeholder("Type here...").
				OnChange(setInput).
				Build(),
		),
		app.NewTextBuilder("").
			Build(),
		app.ButtonBuilder("[Increment]").
			OnClick(func() {
				setCount(func(c int) int { return c + 1 })
			}).
			Build(),
	)
}

func TestDumpOutput(t *testing.T) {
	app, err := ui.RunTestWithSandbox(DemoApp,
		ui.WithWidth(40),
		ui.WithHeight(15),
		ui.WithTitle("Buffer Dump Test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Dump to stdout for visual verification
	t.Log("=== Initial Buffer Dump ===")
	app.DumpBuffer()

	// Save to file
	if err := app.SaveBufferToFile("buffer_output.txt"); err != nil {
		t.Fatal(err)
	}
	t.Log("Buffer saved to buffer_output.txt")

	// Test interaction - click the increment button
	app.GetSandbox().Helper().
		Tab().      // Navigate to button
		Press('\n'). // Press Enter
		Process()

	time.Sleep(100 * time.Millisecond)

	t.Log("=== After Click Dump ===")
	app.DumpBuffer()

	if err := app.SaveBufferToFile("buffer_output_after_click.txt"); err != nil {
		t.Fatal(err)
	}
	t.Log("Buffer saved to buffer_output_after_click.txt")
}
