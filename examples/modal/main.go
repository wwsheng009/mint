package main

import (
	"reflect"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

type OpenModalIntent struct{}
func (OpenModalIntent) IntentType() string { return "OpenModal" }
func (OpenModalIntent) StayPressed() bool  { return true }

type CloseModalIntent struct{}
func (CloseModalIntent) IntentType() string { return "CloseModal" }
func (CloseModalIntent) StayPressed() bool  { return true }

func main() {
	ui.Run(
		func() ui.VNode {
			// Use int state: 0 = closed, 1 = open
			state, setState, _ := ui.UseStateInt(0)
			stateSetterKey := intent.StateKey[func(int)]("modalStateSetter")

			// Save setter
			ctx := ui.GetCurrentContext()
			if ctx != nil {
				ctx.GlobalState[stateSetterKey.String()] = setState
			}

			// If modal is open, show modal content
			if state == 1 {
				return app.VStack(
					ui.NewTextBuilder("┌───────────────────────────────────────┐").FgColor("cyan").Build(),
					ui.NewTextBuilder("│           MODAL IS OPEN               │").FgColor("cyan").Build(),
					ui.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
					ui.NewTextBuilder("│  Do you want to proceed?              │").FgColor("white").Build(),
					ui.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
					ui.HStack(
						ui.NewTextBuilder("│  ").FgColor("cyan").Build(),
						ui.NewButtonBuilder(" Yes ").
							OnPress(CloseModalIntent{}).
							Build(),
						ui.NewTextBuilder("  ").FgColor("cyan").Build(),
						ui.NewButtonBuilder(" No ").
							OnPress(CloseModalIntent{}).
							Build(),
						ui.NewTextBuilder("               │").FgColor("cyan").Build(),
					),
					ui.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
					ui.NewTextBuilder("└───────────────────────────────────────┘").FgColor("cyan").Build(),
					ui.Text(""),
					ui.NewTextBuilder("Press Tab to focus, Enter to close").FgColor("gray").Build(),
				)
			}

			// Modal is closed - show main content
			return app.VStack(
				ui.NewTextBuilder("Modal Demo").Bold(true).FgColor("cyan").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Click the button below to open a modal dialog").FgColor("gray").Build(),
				ui.Text(""),
				ui.NewButtonBuilder("  Show Modal  ").
					OnPress(OpenModalIntent{}).
					Build(),
				ui.Text(""),
				ui.NewTextBuilder("Tab/Arrows: focus | Enter/Space: click").FgColor("gray").Build(),
			)
		},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Modal Demo (MVP)"),
		ui.WithInit(func() {
			stateSetterKey := intent.StateKey[func(int)]("modalStateSetter")

			ui.RegisterIntent(func(ctx *intent.ActionContext, i OpenModalIntent) intent.IntentResult {
				setter, _ := ctx.GetState(stateSetterKey.String())
				callSetter(setter, 1)
				return intent.HandledResult()
			})

			ui.RegisterIntent(func(ctx *intent.ActionContext, i CloseModalIntent) intent.IntentResult {
				setter, _ := ctx.GetState(stateSetterKey.String())
				callSetter(setter, 0)
				return intent.HandledResult()
			})
		}),
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
