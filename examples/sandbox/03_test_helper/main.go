// 03_test_helper/main.go
// TestHelper 链式 API 示例
//
// 演示如何使用 TestHelper 的流畅链式 API
// 简化测试代码编写。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// FormApp 表单应用，用于演示 TestHelper
func FormApp() ui.VNode {
	username, setUsername := ui.UseStateString("")
	password, setPassword := ui.UseStateString("")
	submitted, setSubmitted := ui.UseStateBool(false)
	message, setMessage := ui.UseStateString("")

	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     TestHelper Demo           ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Username: "),
			ui.InputBuilder().
				Value(username).
				Placeholder("Enter username").
				MaxLength(15).
				OnChange(setUsername).
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.Text("Password: "),
			ui.InputBuilder().
				Value(password).
				Placeholder("Enter password").
				MaxLength(20).
				Password().
				OnChange(setPassword).
				Build(),
		),
		ui.Text(""),
		ui.ButtonBuilder("  [ Submit ]  ").
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
		ui.ButtonBuilder("  [ Clear ]  ").
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
				return ui.NewTextBuilder(message).
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
