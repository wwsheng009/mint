package main

import (
	"github.com/wwsheng009/mint/app"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(ElegantAPIDemo,
		ui.WithWidth(80),
		ui.WithHeight(25),
		ui.WithTitle("Elegant API Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func ElegantAPIDemo() ui.VNode {
	return ui.VStackBuilder(
		ui.Text("✨ Elegant VNode Builder API Demo"),
		ui.Text("────────────────────────────────"),

		// Example 1: Elegant chaining - no SetProp!
		ui.Text("1. Flex buttons (no SetProp needed):"),
		ui.HStackBuilder(
			app.ButtonBuilder("Left").
				PaddingH(1, 2).   // Right padding
				Flex(1).          // ✅ Elegant!
				SetTextAlign(rtui.AlignStart).
				Build(),
			app.ButtonBuilder("Center").
				PaddingH(1, 1).   // Even padding
				Flex(1).          // ✅ Elegant!
				SetTextAlign(rtui.AlignCenter).
				Build(),
			app.ButtonBuilder("Right").
				PaddingH(2, 1).   // Left padding
				Flex(1).          // ✅ Elegant!
				SetTextAlign(rtui.AlignEnd).
				Build(),
		).
			Gap(1).
			Build(),

		ui.Text(""),

		// Example 2: Universal padding works on Text too!
		ui.Text("2. Text with PaddingAll(2):"),
		ui.PaddingAll(ui.Text("Padded Text"), 2),

		ui.Text(""),

		// Example 3: Margin for spacing
		ui.Text("3. Buttons with MarginV(0, 1):"),
		app.ButtonBuilder("Btn1").
			MarginV(0, 1).  // ✅ Elegant!
			Build(),
		app.ButtonBuilder("Btn2").
			MarginV(0, 1).  // ✅ Elegant!
			Build(),
		app.ButtonBuilder("Btn3").
			MarginV(0, 1).  // ✅ Elegant!
			Build(),

		ui.Text(""),

		// Example 4: Combined padding + margin + flex
		ui.Text("4. Combined: Padding + Margin + Flex:"),
		app.ButtonBuilder("Spacious").
			PaddingAll(1).    // Inner padding
			MarginV(0, 0).    // No margin
			Flex(1).          // Fill width
			Build(),

		ui.Text(""),

		// Example 5: CSS-like API
		ui.Text("5. Just like CSS:"),
		ui.Text("   button {"),
		ui.Text("       padding: 2px;"),
		ui.Text("       margin: 1px;"),
		ui.Text("       flex: 1;"),
		ui.Text("   }"),
		ui.Text(""),
		ui.Text("   Mint TUI:"),
		app.ButtonBuilder("Click Me").
			PaddingAll(2).
			MarginAll(1).
			Flex(1).
			Build(),
	).
		Gap(0).
	Build()
}
