// Direct Render Modal Positioning Test
// Directly renders multiple modals at different positions to verify rendering
package main

import (
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui/components/modal"
)

func main() {
	err := ui.Run(DirectRenderTest,
		ui.WithWidth(80),
		ui.WithHeight(50),
		ui.WithTitle("Modal Positioning - Direct Render"),
	)
	if err != nil {
		panic(err)
	}
}

// DirectRenderTest shows all positioning methods at once
func DirectRenderTest() ui.VNode {
	return ui.VStack(
		// Header
		ui.VStack(
			app.NewTextBuilder("📍 Modal Positioning - Direct Render Test").
				Bold(true).
				FgColor("cyan").
				Build(),
			app.Text(""),
			app.NewTextBuilder("All modals rendered simultaneously for inspection").
				FgColor("gray").
				Build(),
			app.Text(""),
		),

		app.NewTextBuilder("═").FgColor("white").Build(),

		// Row 1: Centered, Left, Right
		ui.HStack(
			ui.Text("  "),
			ui.VStack(
				ui.Spacer().Build(),
				// Centered Modal
				modal.NewBuilder().
					Title("Center").
					Content(app.Text("Centered")).
					Width(20).
					Height(8).
					Center().
					Open(true).
					Single().
					Build(),
				ui.Spacer().Build(),
			),
			app.Text("  "),
			ui.VStack(
				ui.Spacer().Build(),
				// Left Aligned
				ui.HStack(
					modal.NewBuilder().
						Title("Left").
						Content(app.Text("Left")).
						Width(20).
						Height(8).
						Centered(false).
						Open(true).
						Single().
						Build(),
					ui.Spacer().Flex(1).Build(),
				),
				ui.Spacer().Build(),
			),
			app.Text("  "),
			ui.VStack(
				ui.Spacer().Build(),
				// Right Aligned
				ui.HStack(
					ui.Spacer().Flex(1).Build(),
					modal.NewBuilder().
						Title("Right").
						Content(app.Text("Right")).
						Width(20).
						Height(8).
						Centered(false).
						Open(true).
						Single().
						Build(),
				),
				ui.Spacer().Build(),
			),
			app.Text("  "),
		),

		app.NewTextBuilder("═").FgColor("white").Build(),

		// Row 2: Top and Bottom
		ui.VStack(
			// Top Aligned
			ui.HStack(
				ui.Text("  "),
				ui.VStack(
					modal.NewBuilder().
						Title("Top").
						Content(app.Text("Top Aligned")).
						Width(20).
						Height(8).
						Centered(false).
						Open(true).
						Single().
						Build(),
					ui.Spacer().Flex(2).Build(),
				),
				app.Text("  "),
				ui.VStack(
					ui.Spacer().Flex(2).Build(),
					modal.NewBuilder().
						Title("Bottom").
						Content(app.Text("Bottom")).
						Width(20).
						Height(8).
						Centered(false).
						Open(true).
						Single().
						Build(),
				),
				app.Text("  "),
			),
		),

		app.NewTextBuilder("═").FgColor("white").Build(),

		// Row 3: Custom Position (30%, 20%)
		ui.VStack(
			ui.Spacer().Flex(2).Build(),
			ui.HStack(
				ui.Spacer().Flex(3).Build(),
				modal.NewBuilder().
					Title("Custom").
					Content(app.Text("30%,20%")).
					Width(20).
					Height(8).
					Centered(false).
					Open(true).
					Single().
					Build(),
				ui.Spacer().Flex(7).Build(),
			),
			ui.Spacer().Flex(8).Build(),
		),

		app.NewTextBuilder("═").FgColor("white").Build(),

		// Reference grid
		ui.VStack(
			app.NewTextBuilder("Reference Position Grid:").
				FgColor("yellow").
				Build(),
			app.Text(""),
			printGrid(),
		),
	)
}

// printGrid shows expected positions
func printGrid() ui.VNode {
	return ui.HStack(
		ui.Text("  "),
		ui.VStack(
			app.NewTextBuilder("┌─────────┐").FgColor("gray").Build(),
			app.NewTextBuilder("│ Center  │").FgColor("gray").Build(),
			app.NewTextBuilder("│ should  │").FgColor("gray").Build(),
			app.NewTextBuilder("│ be in   │").FgColor("gray").Build(),
			app.NewTextBuilder("│ middle │").FgColor("gray").Build(),
			app.NewTextBuilder("└─────────┘").FgColor("gray").Build(),
		),
		app.Text("  "),
		ui.VStack(
			ui.Text(""),
			app.NewTextBuilder("Left: No left spacer").
				FgColor("cyan").
				Build(),
			app.NewTextBuilder("Right: Left spacer").
				FgColor("cyan").
				Build(),
			app.NewTextBuilder("Top: No top spacer").
				FgColor("cyan").
				Build(),
			app.NewTextBuilder("Bottom: Top spacer").
				FgColor("cyan").
				Build(),
			app.NewTextBuilder("Custom: Precise flex").
				FgColor("cyan").
				Build(),
		),
		app.Text("  "),
	)
}
