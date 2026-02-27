package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type SetHomeTabIntent struct{}
func (SetHomeTabIntent) IntentType() string { return "SetHomeTab" }
func (SetHomeTabIntent) StayPressed() bool  { return true }

type SetProfileTabIntent struct{}
func (SetProfileTabIntent) IntentType() string { return "SetProfileTab" }
func (SetProfileTabIntent) StayPressed() bool  { return true }

type SetSettingsTabIntent struct{}
func (SetSettingsTabIntent) IntentType() string { return "SetSettingsTab" }
func (SetSettingsTabIntent) StayPressed() bool  { return true }

func main() {
	ui.Run(func() ui.VNode {
		// Use int for tab state: 0 = Home, 1 = Profile, 2 = Settings
		activeTab, setActiveTab, _ := ui.UseStateInt(0)

		// Register intent handlers
		ui.On(SetHomeTabIntent{}, func() {
			setActiveTab(0)
		})
		ui.On(SetProfileTabIntent{}, func() {
			setActiveTab(1)
		})
		ui.On(SetSettingsTabIntent{}, func() {
			setActiveTab(2)
		})

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
					OnPress(SetHomeTabIntent{}).
					Build(),
				app.ButtonBuilder(" Profile ").
					OnPress(SetProfileTabIntent{}).
					Build(),
				app.ButtonBuilder(" Settings ").
					OnPress(SetSettingsTabIntent{}).
					Build(),
			),
			app.Text(""),
			ui.Text("─────────────────────────────────────"),
			app.Text(""),
			content,
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Tabs Demo"),
	)
}
