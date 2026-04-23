package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

type AppState struct {
	Date string
	Time string
}

var appStore = store.NewStore(AppState{
	Date: "2026-04-05",
	Time: "09:30",
})

var appReducer = reducer.NewBuilder[AppState]().
	On(intent.FieldChangeIntent{}, func(state AppState, i intent.Intent) AppState {
		change, ok := i.(intent.FieldChangeIntent)
		if !ok {
			return state
		}
		switch change.Field {
		case "schedule.date":
			state.Date = change.Value
		case "schedule.time":
			state.Time = change.Value
		}
		return state
	})

func DateTimePickerDemo() ui.VNode {
	date := ui.UseStoreSelector(appStore, func(state AppState) string { return state.Date })
	timeValue := ui.UseStoreSelector(appStore, func(state AppState) string { return state.Time })

	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Date + Time Picker Demo").
				Bold(true).
				FgColor("cyan").
				Build(),
			ui.NewTextBuilder("Use Tab to focus, arrow keys / Enter to pick, or type directly.").
				FgColor("bright-black").
				Build(),
			ui.Text(""),
			ui.NewTextBuilder("Ship Date").Bold(true).Build(),
			ui.NewDatePickerBuilder().
				SetID("demo-date").
				ComponentID("demo.datepicker").
				Value(date).
				ForField(intent.BindField("schedule.date")).
				Width(18).
				Build(),
			ui.NewTextBuilder("Ship Time").Bold(true).Build(),
			ui.NewTimePickerBuilder().
				SetID("demo-time").
				ComponentID("demo.timepicker").
				Value(timeValue).
				ForField(intent.BindField("schedule.time")).
				Width(10).
				Build(),
			ui.Text(""),
			ui.NewTextBuilder(fmt.Sprintf("Selected: %s %s", date, timeValue)).
				FgColor("yellow").
				Build(),
		})
}

func main() {
	appReducer.RegisterToGlobal(appStore)

	err := ui.Run(DateTimePickerDemo,
		ui.WithWidth(60),
		ui.WithHeight(16),
		ui.WithTitle("Date + Time Picker Demo"),
	)
	if err != nil {
		panic(err)
	}
}
