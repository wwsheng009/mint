package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	progresscomp "github.com/wwsheng009/mint/ui/components/progress"
)

// =============================================================================
// AppState - Progress Demo State
// =============================================================================

type AppState struct {
	Progress       int
	AccentIndex    int
	UseCustomColor bool
}

var (
	accentColors = []style.Color{
		style.Cyan,
		style.Yellow,
		style.Magenta,
		style.BrightGreen,
		style.BrightBlue,
	}
	accentNames = []string{
		"cyan",
		"yellow",
		"magenta",
		"bright-green",
		"bright-blue",
	}
)

// =============================================================================
// Intent Types
// =============================================================================

type IncrementProgressIntent struct{}

func (IncrementProgressIntent) IntentType() string { return "IncrementProgress" }
func (IncrementProgressIntent) StayPressed() bool  { return true }

type DecrementProgressIntent struct{}

func (DecrementProgressIntent) IntentType() string { return "DecrementProgress" }
func (DecrementProgressIntent) StayPressed() bool  { return true }

type ResetProgressIntent struct{}

func (ResetProgressIntent) IntentType() string { return "ResetProgress" }
func (ResetProgressIntent) StayPressed() bool  { return true }

type CycleAccentIntent struct{}

func (CycleAccentIntent) IntentType() string { return "CycleProgressAccent" }
func (CycleAccentIntent) StayPressed() bool  { return true }

type ToggleCustomColorIntent struct{}

func (ToggleCustomColorIntent) IntentType() string { return "ToggleProgressCustomColor" }
func (ToggleCustomColorIntent) StayPressed() bool  { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var progressStore = store.NewStore(AppState{
	Progress:       35,
	AccentIndex:    0,
	UseCustomColor: false,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(IncrementProgressIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.Progress < 100 {
				s.Progress += 10
				if s.Progress > 100 {
					s.Progress = 100
				}
			}
			return s
		}).
		On(DecrementProgressIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.Progress > 0 {
				s.Progress -= 10
				if s.Progress < 0 {
					s.Progress = 0
				}
			}
			return s
		}).
		On(ResetProgressIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Progress = 0
			return s
		}).
		On(CycleAccentIntent{}, func(s AppState, i intent.Intent) AppState {
			s.AccentIndex = (s.AccentIndex + 1) % len(accentColors)
			s.UseCustomColor = true
			return s
		}).
		On(ToggleCustomColorIntent{}, func(s AppState, i intent.Intent) AppState {
			s.UseCustomColor = !s.UseCustomColor
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), progressStore)
}

// =============================================================================
// Progress Demo Component
// =============================================================================

func ProgressDemo() ui.VNode {
	state := ui.UseStoreSelector(progressStore, func(s AppState) AppState { return s })
	customStyle := currentProgressStyle(state)

	return ui.VStack(
		ui.NewTextBuilder("Progress Animation Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.NewTextBuilder("pure progress response above, active sweep preview below, plus color switching").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("circle/dashboard are segmented views, so they show the same value with lower visual resolution than line/block").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("active sweep uses a highlighted head on top of the filled portion; it should not make segmented progress look shorter than the line/block view").
			FgColor("bright-black").
			Build(),
		ui.Text(""),

		ui.NewTextBuilder("Interactive Progress Response").Bold(true).Build(),
		ui.NewProgressBuilder().
			Label("Line").
			Value(state.Progress).
			Max(100).
			Width(34).
			Status(progresscomp.StatusNormal).
			Style(customStyle).
			Build(),
		ui.NewProgressBuilder().
			Label("Block").
			Value(state.Progress).
			Max(100).
			Width(34).
			Block().
			Status(progresscomp.StatusNormal).
			Style(customStyle).
			Build(),
		ui.NewProgressBuilder().
			Label("Circle").
			Value(state.Progress).
			Max(100).
			Circle().
			Status(progresscomp.StatusNormal).
			Style(customStyle).
			Build(),
		ui.NewProgressBuilder().
			Label("Dashboard").
			Value(state.Progress).
			Max(100).
			Dashboard().
			Status(progresscomp.StatusNormal).
			Style(customStyle).
			Build(),
		ui.Text(""),

		ui.NewTextBuilder("Interactive Active Sweep Preview").Bold(true).Build(),
		ui.NewProgressBuilder().
			Label("Line Active").
			Value(state.Progress).
			Max(100).
			Width(34).
			Status(progresscomp.StatusActive).
			Style(customStyle).
			Build(),
		ui.NewProgressBuilder().
			Label("Block Active").
			Value(state.Progress).
			Max(100).
			Width(34).
			Block().
			Status(progresscomp.StatusActive).
			Style(customStyle).
			Build(),
		ui.NewProgressBuilder().
			Label("Circle Active").
			Value(state.Progress).
			Max(100).
			Circle().
			Status(progresscomp.StatusActive).
			Style(customStyle).
			Build(),
		ui.NewProgressBuilder().
			Label("Dashboard Active").
			Value(state.Progress).
			Max(100).
			Dashboard().
			Status(progresscomp.StatusActive).
			Style(customStyle).
			Build(),
		ui.Text(""),

		ui.NewTextBuilder(progressSummary(state)).FgColor("bright-black").Build(),
		ui.NewTextBuilder(colorSummary(state)).FgColor("bright-black").Build(),
		ui.Text(""),

		ui.HStack(
			ui.NewButtonBuilder(" -10 ").OnPress(DecrementProgressIntent{}).Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" +10 ").OnPress(IncrementProgressIntent{}).Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" Reset ").OnPress(ResetProgressIntent{}).Build(),
		),
		ui.HStack(
			ui.NewButtonBuilder(" Cycle Color ").OnPress(CycleAccentIntent{}).Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(toggleColorLabel(state)).OnPress(ToggleCustomColorIntent{}).Build(),
		),
		ui.Text(""),

		ui.NewTextBuilder("Semantic Colors").Bold(true).Build(),
		ui.NewProgressBuilder().
			Label("Normal").
			Value(55).
			Max(100).
			Width(28).
			Status(progresscomp.StatusNormal).
			Build(),
		ui.NewProgressBuilder().
			Label("Success").
			Value(75).
			Max(100).
			Width(28).
			Status(progresscomp.StatusSuccess).
			Build(),
		ui.NewProgressBuilder().
			Label("Exception").
			Value(65).
			Max(100).
			Width(28).
			Status(progresscomp.StatusException).
			Build(),
		ui.NewProgressBuilder().
			Label("Active").
			Value(45).
			Max(100).
			Width(28).
			Status(progresscomp.StatusActive).
			Build(),
		ui.Text(""),

		ui.NewTextBuilder("Custom Color Override").Bold(true).Build(),
		ui.NewProgressBuilder().
			Label("Accent").
			Value(60).
			Max(100).
			Width(30).
			Status(progresscomp.StatusNormal).
			Style(style.Style{}.Foreground(currentAccentColor(state)).Bold(true)).
			Build(),
		ui.NewTextBuilder("q: quit").FgColor("bright-black").Build(),
	)
}

// =============================================================================
// Helpers
// =============================================================================

func currentAccentColor(state AppState) style.Color {
	return accentColors[state.AccentIndex%len(accentColors)]
}

func currentProgressStyle(state AppState) style.Style {
	if !state.UseCustomColor {
		return style.Style{}
	}
	return style.Style{}.Foreground(currentAccentColor(state)).Bold(true)
}

func progressSummary(state AppState) string {
	return fmt.Sprintf("Current progress: %d%%", state.Progress)
}

func colorSummary(state AppState) string {
	mode := "semantic status color"
	if state.UseCustomColor {
		mode = "custom style color"
	}
	return fmt.Sprintf("Color mode: %s | accent: %s", mode, accentNames[state.AccentIndex%len(accentNames)])
}

func toggleColorLabel(state AppState) string {
	if state.UseCustomColor {
		return " Custom Color: ON "
	}
	return " Custom Color: OFF "
}

// =============================================================================
// Main
// =============================================================================

func main() {
	err := ui.Run(ProgressDemo,
		ui.WithWidth(72),
		ui.WithHeight(38),
		ui.WithTitle("Progress Demo (Animation + Colors + Blocks)"),
	)
	if err != nil {
		panic(err)
	}
}
