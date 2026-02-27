// 03_test_helper/main.go
// TestHelper 链式 API 示例
//
// 演示如何使用 TestHelper 的流畅链式 API
// 简化测试代码编写。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type SubmitFormIntent struct{}
func (SubmitFormIntent) IntentType() string { return "SubmitForm" }
func (SubmitFormIntent) StayPressed() bool  { return true }

type ClearFormIntent struct{}
func (ClearFormIntent) IntentType() string { return "ClearForm" }
func (ClearFormIntent) StayPressed() bool  { return true }

// FormApp 表单应用，用于演示 TestHelper
func FormApp() ui.VNode {
	username, _ := ui.UseStateString("")
	password, _ := ui.UseStateString("")
	submitted, setSubmitted := ui.UseStateBool(false)
	message, setMessage := ui.UseStateString("")

	// Register intent handlers
	ui.On(SubmitFormIntent{}, func() {
		setSubmitted(true)
		if username != "" && password != "" {
			setMessage(fmt.Sprintf("Welcome, %s!", username))
		} else {
			setMessage("Please fill all fields")
		}
	})
	ui.On(ClearFormIntent{}, func() {
		setSubmitted(false)
		setMessage("")
	})

	return ui.VStack(
		app.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("║     TestHelper Demo           ║").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Username: "),
			app.InputBuilder().
				Value(username).
				Placeholder("Enter username").
				Build(), // TODO: integrate with FieldChangeIntent
		),
		ui.Text(""),
		ui.HStack(
			ui.Text("Password: "),
			app.InputBuilder().
				Value(password).
				Placeholder("Enter password").
				Password().
				Build(), // TODO: integrate with FieldChangeIntent
		),
		ui.Text(""),
		app.ButtonBuilder("  [ Submit ]  ").
			OnPress(SubmitFormIntent{}).
			Build(),
		ui.Text(""),
		app.ButtonBuilder("  [ Clear ]  ").
			OnPress(ClearFormIntent{}).
			Build(),
		ui.Text(""),
		func() ui.VNode {
			if submitted {
				return app.NewTextBuilder(message).
					FgColor(func() string {
						if username != "" && password != "" {
							return "green"
						}
						return "red"
					}()).
					Bold(true).
					Build()
			}
			return ui.Text("")
		}(),
	)
}

func main() {
	err := ui.Run(FormApp,
		ui.WithWidth(40),
		ui.WithHeight(16),
		ui.WithTitle("TestHelper Demo"),
	)
	if err != nil {
		panic(err)
	}
}
