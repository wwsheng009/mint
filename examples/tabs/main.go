package main

import (
	"github.com/wwsheng009/mint/ui"
)

func main() {
	ui.Run(func() ui.VNode {
		// Use int for tab state: 0 = Home, 1 = Profile, 2 = Settings
		activeTab, setActiveTab, _ := ui.UseStateInt(0)

		// Tab content based on active tab
		var content ui.VNode
		switch activeTab {
		case 0:
			content = ui.VStack(
				ui.NewTextBuilder("Welcome to the Home tab!").FgColor("green").Build(),
				ui.Text(""),
				ui.NewTextBuilder("This is the main content area.").FgColor("gray").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Navigate using the buttons below:").FgColor("bright-black").Build(),
			)
		case 1:
			content = ui.VStack(
				ui.NewTextBuilder("User Profile").FgColor("cyan").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Name:   John Doe").Build(),
				ui.NewTextBuilder("Email:  john@example.com").Build(),
				ui.NewTextBuilder("Role:   Administrator").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Member since: Jan 2025").FgColor("gray").Build(),
			)
		case 2:
			content = ui.VStack(
				ui.NewTextBuilder("System Settings").FgColor("yellow").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Theme:     Dark").Build(),
				ui.NewTextBuilder("Language:  English").Build(),
				ui.NewTextBuilder("Auto-save:  Enabled").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Notifications: On").FgColor("green").Build(),
			)
		}

		return ui.VStack(
			ui.NewTextBuilder("Tabs Demo").Bold(true).FgColor("cyan").Build(),
			ui.Text(""),
			ui.HStack(
				ui.ButtonBuilder(" Home ").
					OnClick(func() {
						setActiveTab(0)
					}).
					Build(),
				ui.ButtonBuilder(" Profile ").
					OnClick(func() {
						setActiveTab(1)
					}).
					Build(),
				ui.ButtonBuilder(" Settings ").
					OnClick(func() {
						setActiveTab(2)
					}).
					Build(),
			),
			ui.Text(""),
			ui.DividerBuilder().Style(ui.DividerDashed).Build(),
			ui.Text(""),
			content,
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Tabs Demo"),
	)
}
