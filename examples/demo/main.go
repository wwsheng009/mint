package main

import (
	"reflect"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// Intent definitions for DemoApp
type SetTabIntent struct {
	Tab string
}

func (SetTabIntent) IntentType() string { return "SetTab" }
func (SetTabIntent) StayPressed() bool  { return true }

type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

type DecrementIntent struct{}
func (DecrementIntent) IntentType() string { return "Decrement" }
func (DecrementIntent) StayPressed() bool  { return true }

// DemoApp demonstrates all UI components
func DemoApp() ui.VNode {
	currentTab, _ := ui.UseStateString("counter")

	text, setText := ui.UseStateString("")
	textSetterKey := intent.StateKey[func(string)]("textSetter")

	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState[textSetterKey.String()] = setText
	}

	checked1Key := intent.StateKey[bool]("checked1")
	checked1SetterKey := intent.StateKey[func(bool)]("checked1Setter")
	checked1, setChecked1 := ui.UseStateBool(false)

	checked2Key := intent.StateKey[bool]("checked2")
	checked2SetterKey := intent.StateKey[func(bool)]("checked2Setter")
	checked2, setChecked2 := ui.UseStateBool(false)

	checked3Key := intent.StateKey[bool]("checked3")
	checked3SetterKey := intent.StateKey[func(bool)]("checked3Setter")
	checked3, setChecked3 := ui.UseStateBool(false)

	// 保存 setters
	if ctx != nil {
		ctx.GlobalState[checked1SetterKey.String()] = setChecked1
		ctx.GlobalState[checked2SetterKey.String()] = setChecked2
		ctx.GlobalState[checked3SetterKey.String()] = setChecked3
	}

	return ui.VStack(
		ui.NewTextBuilder("╔═══════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     Mint UI Declarative Framework     ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚═══════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		// Tab navigation (简化：不实现 Tab 切换功能)
		ui.HStack(
			ui.NewButtonBuilder(" [1] Counter ").
				OnPress(intent.ClickIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" [2] Input ").
				OnPress(intent.ClickIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" [3] Tasks ").
				OnPress(intent.ClickIntent{}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("───────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		// Content based on selected tab
		func() ui.VNode {
			if currentTab == "counter" {
				return ui.Fragment(
					ui.NewTextBuilder("📊 Counter Demo").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.NewTextBuilder("   Count:   ").
						FgColor("bright-black").
						Build(),
					ui.HStack(
						ui.NewTextBuilder("   ").
							FgColor("green").
							Bold(true).
							Build(),
						ui.NewTextBuilder("  ").
							FgColor("green").
							Bold(true).
							Build(),
						ui.NewButtonBuilder("  [ - ]  ").
							OnPress(intent.ClickIntent{}).
							Build(),
						ui.Text(" "),
						ui.NewButtonBuilder("  [ + ]  ").
							OnPress(intent.ClickIntent{}).
							Build(),
					),
				)
			} else if currentTab == "input" {
				return ui.Fragment(
					ui.NewTextBuilder("📝 Input Demo").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.NewInputBuilder().
						ForField(intent.ForField[string]("text")).
						Placeholder("Type here...").
						Build(),
					ui.Text(""),
					ui.NewTextBuilder("You typed: " + text).
						FgColor("magenta").
						Build(),
				)
			} else {
				return ui.Fragment(
					ui.NewTextBuilder("✓ Tasks Demo").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.NewCheckboxBuilder().
						ForField(intent.ForField(checked1Key)).
						Label("Review documentation").
						Checked(checked1).
						Build(),
					ui.NewCheckboxBuilder().
						ForField(intent.ForField(checked2Key)).
						Label("Write tests").
						Checked(checked2).
						Build(),
					ui.NewCheckboxBuilder().
						ForField(intent.ForField(checked3Key)).
						Label("Build release").
						Checked(checked3).
						Build(),
				)
			}
		}(),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Space/Enter: select | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

func callSetter(fn interface{}, arg interface{}) {
	if fn == nil {
		return
	}
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return
	}
	argV := reflect.ValueOf(arg)
	v.Call([]reflect.Value{argV})
}

func main() {
	err := ui.Run(DemoApp,
		ui.WithWidth(50),
		ui.WithHeight(24),
		ui.WithTitle("Mint UI Demo (MVP)"),
		ui.WithInit(func() {
			textSetterKey := intent.StateKey[func(string)]("textSetter")
			checked1Key := intent.StateKey[bool]("checked1")
			checked1SetterKey := intent.StateKey[func(bool)]("checked1Setter")
			checked2Key := intent.StateKey[bool]("checked2")
			checked2SetterKey := intent.StateKey[func(bool)]("checked2Setter")
			checked3Key := intent.StateKey[bool]("checked3")
			checked3SetterKey := intent.StateKey[func(bool)]("checked3Setter")

			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				switch i.Field {
				case "text":
					setter, _ := ctx.GetState(textSetterKey.String())
					callSetter(setter, i.Value)
				case checked1Key.String():
					setter, _ := ctx.GetState(checked1SetterKey.String())
					value := i.Value == "true"
					callSetter(setter, value)
				case checked2Key.String():
					setter, _ := ctx.GetState(checked2SetterKey.String())
					value := i.Value == "true"
					callSetter(setter, value)
				case checked3Key.String():
					setter, _ := ctx.GetState(checked3SetterKey.String())
					value := i.Value == "true"
					callSetter(setter, value)
				}
				return intent.HandledResult()
			})
		}),
	)
	if err != nil {
		panic(err)
	}
}
