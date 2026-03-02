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
				ui.NewTextBuilder("Welcome to the Home tab!").FgColor("green").Build(),
				ui.Text(""),
				ui.NewTextBuilder("This is the main content area.").FgColor("gray").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Navigate using the buttons below:").FgColor("bright-black").Build(),
			)
		case 1:
			content = app.VStack(
				ui.NewTextBuilder("User Profile").FgColor("cyan").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Name:   John Doe").Build(),
				ui.NewTextBuilder("Email:  john@example.com").Build(),
				ui.NewTextBuilder("Role:   Administrator").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Member since: Jan 2025").FgColor("gray").Build(),
			)
		case 2:
			content = app.VStack(
				ui.NewTextBuilder("System Settings").FgColor("yellow").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Theme:     Dark").Build(),
				ui.NewTextBuilder("Language:  English").Build(),
				ui.NewTextBuilder("Auto-save:  Enabled").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Notifications: On").FgColor("green").Build(),
			)
		}

		return app.VStack(
			ui.NewTextBuilder("Tabs Demo").Bold(true).FgColor("cyan").Build(),
			ui.Text(""),
			ui.HStack(
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
			ui.Text(""),
			ui.Text("─────────────────────────────────────"),
			ui.Text(""),
			content,
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Tabs Demo"),
	)
}
