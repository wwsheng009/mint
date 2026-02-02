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

// FormApp 表单应用，用于演示 TestHelper
func FormApp() ui.VNode {
	username, setUsername := ui.UseStateString("")
	password, setPassword := ui.UseStateString("")
	submitted, setSubmitted := ui.UseStateBool(false)
	message, setMessage := ui.UseStateString("")

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
				MaxLength(15).
				OnChange(setUsername).
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.Text("Password: "),
			app.InputBuilder().
				Value(password).
				Placeholder("Enter password").
				MaxLength(20).
				Password().
				OnChange(setPassword).
				Build(),
		),
		ui.Text(""),
		app.ButtonBuilder("  [ Submit ]  ").
			OnClick(func() {
				setSubmitted(true)
				if username != "" && password != "" {
					setMessage(fmt.Sprintf("Welcome, %s!", username))
				} else {
					setMessage("Please fill all fields")
				}
			}).
			Build(),
		ui.Text(""),
		app.ButtonBuilder("  [ Clear ]  ").
			OnClick(func() {
				setUsername("")
				setPassword("")
				setSubmitted(false)
				setMessage("")
			}).
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
