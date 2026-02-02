package main

import (
	"github.com/wwsheng009/mint/app"
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
			content = app.VStack(
				app.NewTextBuilder("Welcome to the Home tab!").FgColor("green").Build(),
				app.Text(""),
				app.NewTextBuilder("This is the main content area.").FgColor("gray").Build(),
				app.Text(""),
				app.NewTextBuilder("Navigate using the buttons below:").FgColor("bright-black").Build(),
			)
		case 1:
			content = app.VStack(
				app.NewTextBuilder("User Profile").FgColor("cyan").Build(),
				app.Text(""),
				app.NewTextBuilder("Name:   John Doe").Build(),
				app.NewTextBuilder("Email:  john@example.com").Build(),
				app.NewTextBuilder("Role:   Administrator").Build(),
				app.Text(""),
				app.NewTextBuilder("Member since: Jan 2025").FgColor("gray").Build(),
			)
		case 2:
			content = app.VStack(
				app.NewTextBuilder("System Settings").FgColor("yellow").Build(),
				app.Text(""),
				app.NewTextBuilder("Theme:     Dark").Build(),
				app.NewTextBuilder("Language:  English").Build(),
				app.NewTextBuilder("Auto-save:  Enabled").Build(),
				app.Text(""),
				app.NewTextBuilder("Notifications: On").FgColor("green").Build(),
			)
		}

		return app.VStack(
			app.NewTextBuilder("Tabs Demo").Bold(true).FgColor("cyan").Build(),
			app.Text(""),
			app.HStack(
				app.ButtonBuilder(" Home ").
					OnClick(func() {
						setActiveTab(0)
					}).
					Build(),
				app.ButtonBuilder(" Profile ").
					OnClick(func() {
						setActiveTab(1)
					}).
					Build(),
				app.ButtonBuilder(" Settings ").
					OnClick(func() {
						setActiveTab(2)
					}).
					Build(),
			),
			app.Text(""),
			app.DividerBuilder().Style(app.DividerDashed).Build(),
			app.Text(""),
			content,
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Tabs Demo"),
	)
}
