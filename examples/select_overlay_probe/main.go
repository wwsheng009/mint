package main

import (
	"fmt"
	"strconv"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
)

type probeState struct {
	Country string
}

var probeStore = store.NewStore(probeState{
	Country: "0",
})

func init() {
	reducer.NewBuilder[probeState]().
		On(intent.FieldChangeIntent{}, func(s probeState, i intent.Intent) probeState {
			fieldChange, ok := i.(intent.FieldChangeIntent)
			if !ok {
				return s
			}
			if fieldChange.Field == "country" {
				s.Country = fieldChange.Value
			}
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), probeStore)
}

func resetProbeState() {
	probeStore.Update(func(s probeState) probeState {
		s.Country = "0"
		return s
	})
}

func probeOptions() []selectcomp.Option {
	return []selectcomp.Option{
		{Value: "us", Label: "United States"},
		{Value: "cn", Label: "China"},
		{Value: "jp", Label: "Japan"},
	}
}

func probeSelectedIndex() int {
	options := probeOptions()
	raw := probeStore.Get().Country
	idx, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	if idx < 0 || idx >= len(options) {
		return 0
	}
	return idx
}

func probeSelectedLabel() string {
	options := probeOptions()
	idx := probeSelectedIndex()
	if idx >= 0 && idx < len(options) {
		return options[idx].Label
	}
	return "Unknown"
}

func SelectOverlayProbe() ui.VNode {
	return ui.VStack(
		ui.TextBold("Select Overlay Probe"),
		ui.Text("Standalone overlay select using the real app runtime."),
		ui.Text("Use Enter/Down/Enter or mouse click to choose an option."),
		ui.Text(""),
		ui.Text("Country:"),
		ui.NewSelectBuilder().
			SetID("probe.country").
			OverlayPopup(true).
			ForField(intent.BindField("country")).
			Options(probeOptions()).
			Selected(probeSelectedIndex()).
			Width(24).
			Build(),
		ui.Text(""),
		ui.Text(fmt.Sprintf("Selected: %s", probeSelectedLabel())),
	)
}

func main() {
	resetProbeState()
	err := ui.Run(SelectOverlayProbe,
		ui.WithWidth(50),
		ui.WithHeight(14),
		ui.WithTitle("Select Overlay Probe"),
		ui.WithPluginSetup(func(app *framework.App) {
			selectcomp.Install(app)
		}),
	)
	if err != nil {
		panic(err)
	}
}
