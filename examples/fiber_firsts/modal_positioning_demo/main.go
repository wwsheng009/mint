// Fiber-First Modal Positioning Demo
// Demonstrates different ways to position Modal components interactively
package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/modal"
)

func main() {
	err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(45),
		ui.WithTitle("Modal Positioning Demo"),
		ui.WithPluginSetup(func(app *framework.App) {
			// Register modal support for ESC key and click-outside-to-close
			app.AddMiddleware(modal.NewModalMiddleware())
		}),
	)
	if err != nil {
		panic(err)
	}
}

// =============================================================================
// Custom Intent Types
// =============================================================================

type ShowPositioningIntent struct {
	Position string // "center", "left", "right", "top", "bottom", "custom"
}

func (ShowPositioningIntent) IntentType() string { return "ShowPositioning" }

type ClosePositioningIntent struct{}

func (ClosePositioningIntent) IntentType() string { return "ClosePositioning" }

// =============================================================================
// Main App
// =============================================================================

func App() ui.VNode {
	position, setPosition := ui.UseStateString("")

	// Register Intent handlers in App function
	rtui.RegisterIntent(func(ctx *intent.ActionContext, i ShowPositioningIntent) intent.IntentResult {
		setPosition(i.Position)
		return intent.HandledResult()
	})

	rtui.RegisterIntent(func(ctx *intent.ActionContext, i ClosePositioningIntent) intent.IntentResult {
		setPosition("")
		return intent.HandledResult()
	})

	return ui.VStack(
		// Header
		ui.VStack(
			app.NewTextBuilder("📍 Modal Positioning Demo").
				Bold(true).
				FgColor("cyan").
				Build(),
			app.Text(""),
			app.NewTextBuilder("Learn how to control modal display positions").
				FgColor("gray").
				Build(),
			app.NewTextBuilder("Press ESC or click outside to close").
				FgColor("gray").
				Build(),
			app.Text(""),
		),

		app.NewTextBuilder("─").FgColor("gray").Build(),
		app.Text(""),

		// Position Options Grid (Column 1)
		ui.HStack(
			app.Text("  "),
			// Left column
			ui.VStack(
				positionButton("Center", "center"),
				app.Text(""),
				positionButton("Left Aligned", "left"),
				app.Text(""),
				positionButton("Right Aligned", "right"),
			),
			app.Text("   "),
			// Right column
			ui.VStack(
				positionButton("Top Aligned", "top"),
				app.Text(""),
				positionButton("Bottom Aligned", "bottom"),
				app.Text(""),
				positionButton("Custom (30%, 20%)", "custom"),
			),
		),

		app.Text(""),
		app.NewTextBuilder("─").FgColor("gray").Build(),
		app.Text(""),

		// Status Display
		app.HStack(
			app.Text("  "),
			app.NewTextBuilder("Current: ").FgColor("blue").Build(),
			app.NewTextBuilder(getPositionName(position)).
				FgColor("cyan").
				Build(),
			app.Spacer().Build(),
			app.NewTextBuilder(getPositionDescription(position)).
				FgColor("gray").
				Build(),
			app.Text("  "),
		),

		app.Text(""),
		// Render the positioned modal
		getPositionedModal(position),
	)
}

func positionButton(label, pos string) ui.VNode {
	return app.ButtonBuilder(label+"   ").
		Variant(app.ButtonVariantPrimary).
		OnPress(ShowPositioningIntent{Position: pos}).
		Build()
}

func getPositionName(pos string) string {
	switch pos {
	case "":
		return "None"
	case "center":
		return "Centered (Default)"
	case "left":
		return "Left Aligned"
	case "right":
		return "Right Aligned"
	case "top":
		return "Top Aligned"
	case "bottom":
		return "Bottom Aligned"
	case "custom":
		return "Custom Position"
	default:
		return "Unknown"
	}
}

func getPositionDescription(pos string) string {
	switch pos {
	case "center":
		return "Modal centered in container"
	case "left":
		return "Aligned to left with no left spacer"
	case "right":
		return "Aligned to right with left spacer"
	case "top":
		return "Aligned to top with no top spacer"
	case "bottom":
		return "Aligned to bottom with top spacer"
	case "custom":
		return "Positioned at 30% left, 20% top using spacers"
	default:
		return ""
	}
}

// =============================================================================
// Positioned Modals
// =============================================================================

func getPositionedModal(position string) ui.VNode {
	switch position {
	case "center":
		return centeredModal()
	case "left":
		return leftAlignedModal()
	case "right":
		return rightAlignedModal()
	case "top":
		return topAlignedModal()
	case "bottom":
		return bottomAlignedModal()
	case "custom":
		return customPositionModal()
	default:
		return app.Text("")
	}
}

// Centered Modal (Default - most common)
func centeredModal() ui.VNode {
	return ui.VStack(
		// Spacer creates vertical space
		ui.Spacer().Flex(1).Build(),

		// Centered modal
		modal.NewBuilder().
			Key("modal-centered").
			Title("Centered Modal").
			Content(ui.VStack(
				app.NewTextBuilder("🎯 Centered Position").
					FgColor("yellow").
					Bold(true).
					Build(),
				app.Text(""),
				app.NewTextBuilder("This is the default and most common").
					FgColor("gray").
					Build(),
				app.NewTextBuilder("positioning for modals.").
					FgColor("gray").
					Build(),
				app.Text(""),
				app.NewTextBuilder("Built with:").
					FgColor("cyan").
					Build(),
				app.NewTextBuilder("  modal.NewBuilder().").
					FgColor("green").
					Build(),
				app.NewTextBuilder("    Title(\"...\").").
					FgColor("green").
					Build(),
				app.NewTextBuilder("    Center().").
					FgColor("yellow").
					Build(),
				app.NewTextBuilder("    Build()").
					FgColor("green").
					Build(),
			)).
			Width(50).
			Height(14).
			Center().
			Open(true).
			OnClose(ClosePositioningIntent{}).
			Rounded().
			Build(),

		ui.Spacer().Flex(1).Build(),
	)
}

// Left Aligned Modal
func leftAlignedModal() ui.VNode {
	return ui.HStack(
		// Modal left aligned (no left spacer)
		modal.NewBuilder().
			Key("modal-left").
			Title("Left Aligned").
			Content(ui.VStack(
				app.NewTextBuilder("⬅️ Left Aligned").
					FgColor("yellow").
					Bold(true).
					Build(),
				app.Text(""),
				app.NewTextBuilder("No spacer on the left side").
					FgColor("gray").
					Build(),
				app.NewTextBuilder("pushes modal to the left.").
					FgColor("gray").
					Build(),
				app.Text(""),
				app.NewTextBuilder("Built with:").
					FgColor("cyan").
					Build(),
				app.NewTextBuilder("  ui.HStack(").
					FgColor("green").
					Build(),
				app.NewTextBuilder("    modal.NewBuilder().").
					FgColor("green").
					Build(),
				app.NewTextBuilder("      Centered(false).").
					FgColor("yellow").
					Build(),
				app.NewTextBuilder("      Build(),").
					FgColor("green").
					Build(),
				app.NewTextBuilder("    ui.Spacer().Build(),").
					FgColor("green").
					Build(),
				app.NewTextBuilder("  )").
					FgColor("green").
					Build(),
			)).
			Width(45).
			Height(13).
			Centered(false).
			Open(true).
			OnClose(ClosePositioningIntent{}).
			Rounded().
			Build(),

		// Spacer pushes modal to left
		ui.Spacer().Flex(1).Build(),
	)
}

// Right Aligned Modal
func rightAlignedModal() ui.VNode {
	return ui.HStack(
		// Spacer pushes modal to right
		ui.Spacer().Flex(1).Build(),

		// Modal right aligned
		modal.NewBuilder().
			Key("modal-right").
			Title("Right Aligned").
			Content(ui.VStack(
				app.NewTextBuilder("➡️ Right Aligned").
					FgColor("yellow").
					Bold(true).
					Build(),
				app.Text(""),
				app.NewTextBuilder("Spacer on the left").
					FgColor("gray").
					Build(),
				app.NewTextBuilder("pushes modal to the right.").
					FgColor("gray").
					Build(),
				app.Text(""),
				app.NewTextBuilder("Built with:").
					FgColor("cyan").
					Build(),
				app.NewTextBuilder("  ui.HStack(").
					FgColor("green").
					Build(),
				app.NewTextBuilder("    ui.Spacer().Build(),").
					FgColor("green").
					Build(),
				app.NewTextBuilder("    modal.NewBuilder().").
					FgColor("green").
					Build(),
				app.NewTextBuilder("      Centered(false).").
					FgColor("yellow").
					Build(),
				app.NewTextBuilder("      Build(),").
					FgColor("green").
					Build(),
				app.NewTextBuilder("  )").
					FgColor("green").
					Build(),
			)).
			Width(45).
			Height(13).
			Centered(false).
			Open(true).
			OnClose(ClosePositioningIntent{}).
			Rounded().
			Build(),
	)
}

// Top Aligned Modal
func topAlignedModal() ui.VNode {
	return ui.VStack(
		// Modal top aligned (no top spacer)
		modal.NewBuilder().
			Key("modal-top").
			Title("Top Aligned").
			Content(ui.VStack(
				app.NewTextBuilder("⬆️ Top Aligned").
					FgColor("yellow").
					Bold(true).
					Build(),
				app.Text(""),
				app.NewTextBuilder("No spacer on the top").
					FgColor("gray").
					Build(),
				app.NewTextBuilder("pushes modal to the top.").
					FgColor("gray").
					Build(),
				app.Text(""),
				app.NewTextBuilder("Built with:").
					FgColor("cyan").
					Build(),
				app.NewTextBuilder("  ui.VStack(").
					FgColor("green").
					Build(),
				app.NewTextBuilder("    modal.NewBuilder().").
					FgColor("green").
					Build(),
				app.NewTextBuilder("      Centered(false).").
					FgColor("yellow").
					Build(),
				app.NewTextBuilder("      Build(),").
					FgColor("green").
					Build(),
				app.NewTextBuilder("    ui.Spacer().Build(),").
					FgColor("green").
					Build(),
				app.NewTextBuilder("  )").
					FgColor("green").
					Build(),
			)).
			Width(45).
			Height(13).
			Centered(false).
			Open(true).
			OnClose(ClosePositioningIntent{}).
			Rounded().
			Build(),

		// Spacer pushes modal to top
		ui.Spacer().Flex(1).Build(),
	)
}

// Bottom Aligned Modal
func bottomAlignedModal() ui.VNode {
	return ui.VStack(
		// Spacer pushes modal to bottom
		ui.Spacer().Flex(1).Build(),

		// Modal bottom aligned
		modal.NewBuilder().
			Key("modal-bottom").
			Title("Bottom Aligned").
			Content(ui.VStack(
				app.NewTextBuilder("⬇️ Bottom Aligned").
					FgColor("yellow").
					Bold(true).
					Build(),
				app.Text(""),
				app.NewTextBuilder("Spacer on the top").
					FgColor("gray").
					Build(),
				app.NewTextBuilder("pushes modal to the bottom.").
					FgColor("gray").
					Build(),
				app.Text(""),
				app.NewTextBuilder("Built with:").
					FgColor("cyan").
					Build(),
				app.NewTextBuilder("  ui.VStack(").
					FgColor("green").
					Build(),
				app.NewTextBuilder("    ui.Spacer().Build(),").
					FgColor("green").
					Build(),
				app.NewTextBuilder("    modal.NewBuilder().").
					FgColor("green").
					Build(),
				app.NewTextBuilder("      Centered(false).").
					FgColor("yellow").
					Build(),
				app.NewTextBuilder("      Build(),").
					FgColor("green").
					Build(),
				app.NewTextBuilder("  )").
					FgColor("green").
					Build(),
			)).
			Width(45).
			Height(13).
			Centered(false).
			Open(true).
			OnClose(ClosePositioningIntent{}).
			Rounded().
			Build(),
	)
}

// Custom Position Modal (30% left, 20% top)
func customPositionModal() ui.VNode {
	// To position at (20% top, 30% left):
	// - Top: 2 parts, Bottom: 8 parts = 20% top
	// - Left: 3 parts, Right: 7 parts = 30% left
	//
	// Formula:
	//   top%    = topSpacer / (top + bottom)
	//   left%   = leftSpacer / (left + right)
	return ui.VStack(
		// Top spacer: 20% (2 parts)
		ui.Spacer().Flex(2).Build(),

		ui.HStack(
			// Left spacer: 30% (3 parts)
			ui.Spacer().Flex(3).Build(),

			// Modal at (20% top, 30% left)
			modal.NewBuilder().
				Key("modal-custom").
				Title("Custom Position").
				Content(ui.VStack(
					app.NewTextBuilder("🎯 Custom: (30%, 20%)").
						FgColor("yellow").
						Bold(true).
						Build(),
					app.Text(""),
					app.NewTextBuilder("Using spacers to set").
						FgColor("gray").
						Build(),
					app.NewTextBuilder("precise percentage position.").
						FgColor("gray").
						Build(),
					app.Text(""),
					app.NewTextBuilder("Position Formula:").
						FgColor("cyan").
						Build(),
					app.NewTextBuilder("  left%  = left / (left + right)").
						FgColor("green").
						Build(),
					app.NewTextBuilder("  top%   = top / (top + bottom)").
						FgColor("green").
						Build(),
					app.Text(""),
					app.NewTextBuilder("This example:").
						FgColor("yellow").
						Build(),
					app.NewTextBuilder("  left  = 3, right  = 7 → 30%").
						FgColor("gray").
						Build(),
					app.NewTextBuilder("  top   = 2, bottom = 8 → 20%").
						FgColor("gray").
						Build(),
				)).
				Width(52).
				Height(16).
				Centered(false).
				Open(true).
				OnClose(ClosePositioningIntent{}).
				Rounded().
				Build(),

			// Right spacer: 70% (7 parts)
			ui.Spacer().Flex(7).Build(),
		),

		// Bottom spacer: 80% (8 parts)
		ui.Spacer().Flex(8).Build(),
	)
}
