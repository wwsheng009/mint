package main

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	clockcomp "github.com/wwsheng009/mint/ui/components/clock"
)

type ZonePreset struct {
	Label    string
	Name     string
	Location *time.Location
	Accent   style.Color
}

type AppState struct {
	ZoneIndex      int
	Shape          clockcomp.DialShape
	RadiusX        int
	RadiusY        int
	CellAspectX    float64
	ShowSecondHand bool
	SmoothSecond   bool
	ShowDigital    bool
	Preset         clockcomp.Preset
	HandStyle      clockcomp.HandRenderStyle
	LastAction     string
}

type ToggleSecondHandIntent struct{}
type ToggleSmoothSecondIntent struct{}
type ToggleDigitalIntent struct{}
type ToggleHandStyleIntent struct{}
type ResetDemoIntent struct{}
type NextZoneIntent struct{}
type NextPresetIntent struct{}
type NextShapeIntent struct{}

type AdjustWidthIntent struct {
	Delta int
}

type AdjustHeightIntent struct {
	Delta int
}

type AdjustCellAspectIntent struct {
	Delta float64
}

type SetZoneIntent struct {
	Index int
}

type SetPresetIntent struct {
	Preset clockcomp.Preset
}

type SetShapeIntent struct {
	Shape clockcomp.DialShape
}

func (ToggleSecondHandIntent) IntentType() string   { return "ClockDemoToggleSecondHand" }
func (ToggleSmoothSecondIntent) IntentType() string { return "ClockDemoToggleSmoothSecond" }
func (ToggleDigitalIntent) IntentType() string      { return "ClockDemoToggleDigital" }
func (ToggleHandStyleIntent) IntentType() string    { return "ClockDemoToggleHandStyle" }
func (ResetDemoIntent) IntentType() string          { return "ClockDemoReset" }
func (NextZoneIntent) IntentType() string           { return "ClockDemoNextZone" }
func (NextPresetIntent) IntentType() string         { return "ClockDemoNextPreset" }
func (NextShapeIntent) IntentType() string          { return "ClockDemoNextShape" }
func (AdjustWidthIntent) IntentType() string        { return "ClockDemoAdjustWidth" }
func (AdjustHeightIntent) IntentType() string       { return "ClockDemoAdjustHeight" }
func (AdjustCellAspectIntent) IntentType() string   { return "ClockDemoAdjustCellAspect" }
func (SetZoneIntent) IntentType() string            { return "ClockDemoSetZone" }
func (SetPresetIntent) IntentType() string          { return "ClockDemoSetPreset" }
func (SetShapeIntent) IntentType() string           { return "ClockDemoSetShape" }
func (ToggleSecondHandIntent) StayPressed() bool    { return true }
func (ToggleSmoothSecondIntent) StayPressed() bool  { return true }
func (ToggleDigitalIntent) StayPressed() bool       { return true }
func (ToggleHandStyleIntent) StayPressed() bool     { return true }
func (ResetDemoIntent) StayPressed() bool           { return true }
func (NextZoneIntent) StayPressed() bool            { return true }
func (NextPresetIntent) StayPressed() bool          { return true }
func (NextShapeIntent) StayPressed() bool           { return true }
func (AdjustWidthIntent) StayPressed() bool         { return true }
func (AdjustHeightIntent) StayPressed() bool        { return true }
func (AdjustCellAspectIntent) StayPressed() bool    { return true }
func (SetZoneIntent) StayPressed() bool             { return true }
func (SetPresetIntent) StayPressed() bool           { return true }
func (SetShapeIntent) StayPressed() bool            { return true }

var (
	zonePresets = []ZonePreset{
		{
			Label:    "UTC",
			Name:     "UTC",
			Location: time.UTC,
			Accent:   style.BrightCyan,
		},
		{
			Label:    "Shanghai",
			Name:     "Asia/Shanghai",
			Location: mustLoadLocation("Asia/Shanghai"),
			Accent:   style.BrightYellow,
		},
		{
			Label:    "NewYork",
			Name:     "America/New_York",
			Location: mustLoadLocation("America/New_York"),
			Accent:   style.BrightGreen,
		},
		{
			Label:    "London",
			Name:     "Europe/London",
			Location: mustLoadLocation("Europe/London"),
			Accent:   style.BrightMagenta,
		},
	}
	clockPresets = []clockcomp.Preset{
		clockcomp.PresetNone,
		clockcomp.PresetClassic,
		clockcomp.PresetNeon,
		clockcomp.PresetMinimal,
		clockcomp.PresetAlert,
	}
	dialShapes = []clockcomp.DialShape{
		clockcomp.DialShapeCircle,
		clockcomp.DialShapeEllipse,
	}
	snapshotInstant = time.Date(2026, 3, 29, 9, 15, 30, 0, time.UTC)
	demoStore       = store.NewStore(newInitialState())
)

func init() {
	reducer.NewBuilder[AppState]().
		On(ToggleSecondHandIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowSecondHand = !s.ShowSecondHand
			s.LastAction = fmt.Sprintf("ShowSecondHand = %t", s.ShowSecondHand)
			return s
		}).
		On(ToggleSmoothSecondIntent{}, func(s AppState, i intent.Intent) AppState {
			s.SmoothSecond = !s.SmoothSecond
			s.LastAction = fmt.Sprintf("SmoothSecond = %t", s.SmoothSecond)
			return s
		}).
		On(ToggleDigitalIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowDigital = !s.ShowDigital
			s.LastAction = fmt.Sprintf("ShowDigital = %t", s.ShowDigital)
			return s
		}).
		On(ToggleHandStyleIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.HandStyle == clockcomp.HandRenderStyleUnicode {
				s.HandStyle = clockcomp.HandRenderStyleASCII
			} else {
				s.HandStyle = clockcomp.HandRenderStyleUnicode
			}
			s.LastAction = fmt.Sprintf("HandStyle = %s", handStyleLabel(s.HandStyle))
			return s
		}).
		On(NextPresetIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Preset = nextPreset(s.Preset)
			s.LastAction = fmt.Sprintf("Preset = %s", clockcomp.PresetName(s.Preset))
			return s
		}).
		On(NextShapeIntent{}, func(s AppState, i intent.Intent) AppState {
			s = withGeometry(s, nextShape(s.Shape), s.RadiusX, s.RadiusY)
			s.LastAction = fmt.Sprintf("Shape = %s", clockcomp.DialShapeName(s.Shape))
			return s
		}).
		On(AdjustWidthIntent{}, func(s AppState, i intent.Intent) AppState {
			adjust, ok := i.(AdjustWidthIntent)
			if !ok {
				return s
			}
			s = adjustWidth(s, adjust.Delta)
			s.LastAction = fmt.Sprintf("WidthRadius = %d", s.RadiusX)
			return s
		}).
		On(AdjustHeightIntent{}, func(s AppState, i intent.Intent) AppState {
			adjust, ok := i.(AdjustHeightIntent)
			if !ok {
				return s
			}
			s = adjustHeight(s, adjust.Delta)
			s.LastAction = fmt.Sprintf("HeightRadius = %d", s.RadiusY)
			return s
		}).
		On(AdjustCellAspectIntent{}, func(s AppState, i intent.Intent) AppState {
			adjust, ok := i.(AdjustCellAspectIntent)
			if !ok {
				return s
			}
			s.CellAspectX = clampCellAspectX(s.CellAspectX + adjust.Delta)
			s.LastAction = fmt.Sprintf("CellAspectX = %.2f", s.CellAspectX)
			return s
		}).
		On(SetZoneIntent{}, func(s AppState, i intent.Intent) AppState {
			setZone, ok := i.(SetZoneIntent)
			if !ok {
				return s
			}
			s.ZoneIndex = normalizeZoneIndex(setZone.Index)
			s.LastAction = fmt.Sprintf("Zone = %s", selectedZone(s).Name)
			return s
		}).
		On(NextZoneIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ZoneIndex = normalizeZoneIndex(s.ZoneIndex + 1)
			s.LastAction = fmt.Sprintf("Zone = %s", selectedZone(s).Name)
			return s
		}).
		On(SetPresetIntent{}, func(s AppState, i intent.Intent) AppState {
			setPreset, ok := i.(SetPresetIntent)
			if !ok {
				return s
			}
			s.Preset = normalizePreset(setPreset.Preset)
			s.LastAction = fmt.Sprintf("Preset = %s", clockcomp.PresetName(s.Preset))
			return s
		}).
		On(SetShapeIntent{}, func(s AppState, i intent.Intent) AppState {
			setShape, ok := i.(SetShapeIntent)
			if !ok {
				return s
			}
			s = withGeometry(s, normalizeShape(setShape.Shape), s.RadiusX, s.RadiusY)
			s.LastAction = fmt.Sprintf("Shape = %s", clockcomp.DialShapeName(s.Shape))
			return s
		}).
		On(ResetDemoIntent{}, func(s AppState, i intent.Intent) AppState {
			return newInitialState()
		}).
		BuildAndRegister(intent.DefaultRegistry(), demoStore)
}

func main() {
	err := ui.Run(App,
		ui.WithWidth(128),
		ui.WithHeight(42),
		ui.WithTitle("Clock Demo"),
		ui.WithPluginSetup(func(app *framework.App) {
			app.OnKeyCombo("f1", func() { ui.EmitIntentGlobal(NextShapeIntent{}) })
			app.OnKeyCombo("f2", func() { ui.EmitIntentGlobal(ToggleSecondHandIntent{}) })
			app.OnKeyCombo("f3", func() { ui.EmitIntentGlobal(ToggleSmoothSecondIntent{}) })
			app.OnKeyCombo("f4", func() { ui.EmitIntentGlobal(ToggleDigitalIntent{}) })
			app.OnKeyCombo("f5", func() { ui.EmitIntentGlobal(AdjustWidthIntent{Delta: -1}) })
			app.OnKeyCombo("f6", func() { ui.EmitIntentGlobal(AdjustWidthIntent{Delta: 1}) })
			app.OnKeyCombo("f7", func() { ui.EmitIntentGlobal(NextZoneIntent{}) })
			app.OnKeyCombo("f8", func() { ui.EmitIntentGlobal(ResetDemoIntent{}) })
			app.OnKeyCombo("f9", func() { ui.EmitIntentGlobal(ToggleHandStyleIntent{}) })
			app.OnKeyCombo("f10", func() { ui.EmitIntentGlobal(NextPresetIntent{}) })
			app.OnKeyCombo("f11", func() { ui.EmitIntentGlobal(AdjustHeightIntent{Delta: -1}) })
			app.OnKeyCombo("f12", func() { ui.EmitIntentGlobal(AdjustHeightIntent{Delta: 1}) })
		}),
	)
	if err != nil {
		panic(err)
	}
}

func App() ui.VNode {
	state := demoStore.Get()

	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			headerPanel(state),
			controlsPanel(state),
			ui.HStackBuilder(
				ui.Flex(leftPane(state), 3),
				ui.Flex(rightPane(state), 2),
			).Gap(1).Stretch().Build(),
		})
}

func headerPanel(state AppState) ui.VNode {
	zone := selectedZone(state)
	widthCells, heightCells := dialCells(state.RadiusX, state.RadiusY, state.CellAspectX)
	return ui.NewVStack().
		SingleBorder("Clock Demo").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Live refresh, fixed snapshots, timezone-aware rendering, and circle/ellipse geometry.").Bold(true).FgColor("bright-cyan").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Zone=%s  Shape=%s  Radii=%dx%d  Cells=%dx%d",
				zone.Name,
				clockcomp.DialShapeName(state.Shape),
				state.RadiusX,
				state.RadiusY,
				widthCells,
				heightCells,
			)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("CellAspectX=%.2f  Circle uses matched logical radii and lets aspect compensation control the visible width.", state.CellAspectX)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Seconds=%s  Smooth=%s  Digital=%s  Preset=%s  HandStyle=%s",
				onOff(state.ShowSecondHand),
				onOff(state.SmoothSecond),
				onOff(state.ShowDigital),
				clockcomp.PresetName(state.Preset),
				handStyleLabel(state.HandStyle),
			)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Snapshot instant: %s", snapshotInstant.Format(time.RFC3339))).FgColor("yellow").Build(),
			ui.NewTextBuilder(fmt.Sprintf("LastAction=%s", state.LastAction)).FgColor("bright-black").Build(),
		})
}

func controlsPanel(state AppState) ui.VNode {
	zoneButtons := make([]ui.VNode, 0, len(zonePresets))
	for index, zone := range zonePresets {
		zoneButtons = append(zoneButtons, zoneButton(zone.Label, state.ZoneIndex == index, SetZoneIntent{Index: index}))
	}

	presetButtons := make([]ui.VNode, 0, len(clockPresets))
	for _, preset := range clockPresets {
		presetButtons = append(presetButtons, presetButton(clockcomp.PresetName(preset), state.Preset == preset, SetPresetIntent{Preset: preset}))
	}

	shapeButtons := make([]ui.VNode, 0, len(dialShapes))
	for _, shape := range dialShapes {
		shapeButtons = append(shapeButtons, shapeButton(clockcomp.DialShapeName(shape), state.Shape == shape, SetShapeIntent{Shape: shape}))
	}

	return ui.NewVStack().
		SingleBorder("Controls").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.HStackBuilder(
				toggleButton("Seconds", state.ShowSecondHand, ToggleSecondHandIntent{}),
				toggleButton("Smooth", state.SmoothSecond, ToggleSmoothSecondIntent{}),
				toggleButton("Digital", state.ShowDigital, ToggleDigitalIntent{}),
				toggleButton("Hands", state.HandStyle == clockcomp.HandRenderStyleUnicode, ToggleHandStyleIntent{}),
				ui.NewButtonBuilder("Reset").Secondary().OnPress(ResetDemoIntent{}).Build(),
				ui.NewTextBuilder("Buttons and F-keys both update the same store-backed state.").FgColor("bright-black").Build(),
			).Gap(1).Build(),
			ui.HStackBuilder(
				shapeButtons...,
			).Gap(1).Build(),
			ui.NewTextBuilder(shapeHint(state.Shape)).FgColor("bright-black").Build(),
			ui.HStackBuilder(
				ui.NewTextBuilder("Width").FgColor("cyan").Build(),
				ui.NewButtonBuilder("-").Secondary().OnPress(AdjustWidthIntent{Delta: -1}).Build(),
				ui.NewTextBuilder(fmt.Sprintf("%d", state.RadiusX)).Bold(true).FgColor("bright-white").Build(),
				ui.NewButtonBuilder("+").Primary().OnPress(AdjustWidthIntent{Delta: 1}).Build(),
				ui.NewTextBuilder("Horizontal radius range: 3..8").FgColor("bright-black").Build(),
			).Gap(1).Build(),
			ui.HStackBuilder(
				ui.NewTextBuilder("Height").FgColor("cyan").Build(),
				ui.NewButtonBuilder("-").Secondary().OnPress(AdjustHeightIntent{Delta: -1}).Build(),
				ui.NewTextBuilder(fmt.Sprintf("%d", state.RadiusY)).Bold(true).FgColor("bright-white").Build(),
				ui.NewButtonBuilder("+").Primary().OnPress(AdjustHeightIntent{Delta: 1}).Build(),
				ui.NewTextBuilder(heightHint(state.Shape)).FgColor("bright-black").Build(),
			).Gap(1).Build(),
			ui.HStackBuilder(
				ui.NewTextBuilder("Aspect").FgColor("cyan").Build(),
				ui.NewButtonBuilder("-").Secondary().OnPress(AdjustCellAspectIntent{Delta: -0.25}).Build(),
				ui.NewTextBuilder(fmt.Sprintf("%.2f", state.CellAspectX)).Bold(true).FgColor("bright-white").Build(),
				ui.NewButtonBuilder("+").Primary().OnPress(AdjustCellAspectIntent{Delta: 0.25}).Build(),
				ui.NewTextBuilder("Horizontal cell compensation range: 1.00..3.00").FgColor("bright-black").Build(),
			).Gap(1).Build(),
			ui.HStackBuilder(zoneButtons...).Gap(1).Build(),
			ui.HStackBuilder(presetButtons...).Gap(1).Build(),
			ui.NewTextBuilder("Shortcuts: F1 shape  F2 seconds  F3 smooth  F4 digital  F5/F6 width  F7 next zone  F8 reset  F9 hand style  F10 preset  F11/F12 height  Aspect uses buttons").FgColor("bright-black").Build(),
		})
}

func leftPane(state AppState) ui.VNode {
	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			livePreviewPanel(state),
			snapshotPanel(state),
		})
}

func rightPane(state AppState) ui.VNode {
	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			timezonePanel(state),
			recipePanel(),
		})
}

func livePreviewPanel(state AppState) ui.VNode {
	zone := selectedZone(state)
	widthCells, heightCells := dialCells(state.RadiusX, state.RadiusY, state.CellAspectX)

	clockNode := ui.NewClockBuilder().
		Shape(state.Shape).
		Radii(state.RadiusX, state.RadiusY).
		CellAspectX(state.CellAspectX).
		Realtime().
		Location(zone.Location).
		ShowSecondHand(state.ShowSecondHand).
		SmoothSecond(state.SmoothSecond).
		ShowDigital(state.ShowDigital).
		Preset(state.Preset).
		HandStyle(state.HandStyle).
		Build()

	return ui.NewVStack().
		SingleBorder("Live Preview").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.HStackBuilder(
				clockNode,
				ui.NewVStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder(fmt.Sprintf("Current zone: %s", zone.Name)).FgColor(zone.Accent).Bold(true).Build(),
						ui.NewTextBuilder("This clock runs in realtime and keeps ticking while the app is idle.").FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Shape: %s", clockcomp.DialShapeName(state.Shape))).FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Radii: %d x %d", state.RadiusX, state.RadiusY)).FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("CellAspectX: %.2f", state.CellAspectX)).FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Face cells: %d x %d", widthCells, heightCells)).FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Second hand: %s", onOff(state.ShowSecondHand))).FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Smooth mode: %s", onOff(state.SmoothSecond))).FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Digital label: %s", onOff(state.ShowDigital))).FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Preset: %s", clockcomp.PresetName(state.Preset))).FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Hand style: %s", handStyleLabel(state.HandStyle))).FgColor("bright-white").Build(),
						ui.NewTextBuilder("Circle mode keeps width and height locked. Ellipse mode lets each axis vary independently.").FgColor("bright-black").Build(),
					}),
			).Gap(3).Build(),
		})
}

func snapshotPanel(state AppState) ui.VNode {
	zone := selectedZone(state)
	localSnapshot := snapshotInstant.In(zone.Location)
	_, snapshotRadiusX, snapshotRadiusY := normalizeGeometry(state.Shape, maxInt(3, state.RadiusX-1), maxInt(3, state.RadiusY-1))
	snapshotTheme := clockcomp.ThemePreset(state.Preset).Merge(
		clockcomp.Theme{}.
			WithDigitalStyle(style.Style{}.Foreground(zone.Accent).Bold(true)).
			WithMinuteHandStyle(style.Style{}.Foreground(zone.Accent).Bold(true)).
			WithCenterStyle(style.Style{}.Foreground(zone.Accent)),
	)

	clockNode := ui.NewClockBuilder().
		Shape(state.Shape).
		Radii(snapshotRadiusX, snapshotRadiusY).
		CellAspectX(state.CellAspectX).
		StaticTime(snapshotInstant).
		Location(zone.Location).
		ShowSecondHand(true).
		SmoothSecond(false).
		ShowDigital(true).
		Theme(snapshotTheme).
		HandStyle(state.HandStyle).
		Build()

	return ui.NewVStack().
		SingleBorder("Static Snapshot").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.HStackBuilder(
				clockNode,
				ui.NewVStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder("The snapshot never ticks. It always renders the same instant.").FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Base instant:  %s", snapshotInstant.Format("2006-01-02 15:04:05 MST"))).FgColor("bright-black").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Rendered as:   %s", localSnapshot.Format("2006-01-02 15:04:05 MST"))).FgColor("yellow").Build(),
						ui.NewTextBuilder(fmt.Sprintf("Shape: %s  Radii: %d x %d", clockcomp.DialShapeName(state.Shape), snapshotRadiusX, snapshotRadiusY)).FgColor("bright-white").Build(),
						ui.NewTextBuilder(fmt.Sprintf("CellAspectX: %.2f", state.CellAspectX)).FgColor("bright-white").Build(),
						ui.NewTextBuilder("This panel uses ThemePreset(...).Merge(...) to layer a small zone-accent theme.").FgColor("bright-white").Build(),
						ui.NewTextBuilder("This is useful for audit markers, historical events, and fixed schedule cards.").FgColor("bright-white").Build(),
					}),
			).Gap(3).Build(),
		})
}

func timezonePanel(state AppState) ui.VNode {
	children := []ui.VNode{
		ui.NewTextBuilder("Same instant rendered in four timezones:").Bold(true).FgColor("bright-cyan").Build(),
		ui.NewTextBuilder(snapshotInstant.Format("2006-01-02 15:04:05 MST")).FgColor("bright-black").Build(),
	}

	for _, zone := range zonePresets {
		localTime := snapshotInstant.In(zone.Location)
		children = append(children,
			ui.NewTextBuilder(fmt.Sprintf("%-10s %s", zone.Label, localTime.Format("2006-01-02 15:04:05 MST"))).
				FgColor(colorName(zone.Accent)).
				Build(),
		)
	}

	children = append(children,
		ui.NewTextBuilder(fmt.Sprintf("Selected zone for live preview: %s", selectedZone(state).Name)).FgColor("bright-white").Build(),
	)

	return ui.NewVStack().
		SingleBorder("Timezone View").
		SetGap(0).
		SetChildrenList(children)
}

func recipePanel() ui.VNode {
	return ui.NewVStack().
		SingleBorder("Usage Notes").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Wall clock").FgColor("bright-cyan").Bold(true).Build(),
			ui.NewTextBuilder("Use live mode with a visible second hand for dashboards and ops screens.").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Dial shape").FgColor("bright-cyan").Bold(true).Build(),
			ui.NewTextBuilder("Circle keeps width and height matched. Ellipse lets each axis be adjusted independently.").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Cell aspect").FgColor("bright-cyan").Bold(true).Build(),
			ui.NewTextBuilder("CellAspectX(...) controls horizontal compensation for different terminal font grids.").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Preset mode").FgColor("bright-cyan").Bold(true).Build(),
			ui.NewTextBuilder("Default / Classic / Neon / Minimal / Alert are built in now.").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Theme preset").FgColor("bright-cyan").Bold(true).Build(),
			ui.NewTextBuilder("ThemePreset(...) and Theme.WithPreset(...) are shorthand constructors for preset-based themes.").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Theme merge").FgColor("bright-cyan").Bold(true).Build(),
			ui.NewTextBuilder("ThemePreset(...).Merge(...) lets you layer a preset with a small accent theme.").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Snapshot card").FgColor("bright-cyan").Bold(true).Build(),
			ui.NewTextBuilder("Use StaticTime(...) plus Location(...) for historical or scheduled events.").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Face styles").FgColor("bright-cyan").Bold(true).Build(),
			ui.NewTextBuilder("Dial, tick marks, center hub, and digital label can all be themed separately.").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Per-hand styles").FgColor("bright-cyan").Bold(true).Build(),
			ui.NewTextBuilder("Explicit hand/face styles can still override any preset.").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Focus").FgColor("bright-cyan").Bold(true).Build(),
			ui.NewTextBuilder("This demo is intentionally control-heavy so every supported prop is visible.").FgColor("bright-black").Build(),
		})
}

func toggleButton(label string, active bool, action intent.Intent) ui.VNode {
	builder := ui.NewButtonBuilder(fmt.Sprintf("%s: %s", label, onOff(active))).OnPress(action)
	if active {
		return builder.Primary().Build()
	}
	return builder.Secondary().Build()
}

func zoneButton(label string, active bool, action intent.Intent) ui.VNode {
	builder := ui.NewButtonBuilder(label).OnPress(action)
	if active {
		return builder.Primary().Build()
	}
	return builder.Secondary().Build()
}

func presetButton(label string, active bool, action intent.Intent) ui.VNode {
	builder := ui.NewButtonBuilder(label).OnPress(action)
	if active {
		return builder.Primary().Build()
	}
	return builder.Secondary().Build()
}

func shapeButton(label string, active bool, action intent.Intent) ui.VNode {
	builder := ui.NewButtonBuilder(label).OnPress(action)
	if active {
		return builder.Primary().Build()
	}
	return builder.Secondary().Build()
}

func newInitialState() AppState {
	state := AppState{
		ZoneIndex:      1,
		Shape:          clockcomp.DialShapeCircle,
		RadiusX:        6,
		RadiusY:        6,
		CellAspectX:    clockcomp.DefaultCellAspectX,
		ShowSecondHand: true,
		SmoothSecond:   true,
		ShowDigital:    true,
		Preset:         clockcomp.PresetClassic,
		HandStyle:      clockcomp.HandRenderStyleASCII,
		LastAction:     "Ready",
	}
	return withGeometry(state, state.Shape, state.RadiusX, state.RadiusY)
}

func selectedZone(state AppState) ZonePreset {
	return zonePresets[normalizeZoneIndex(state.ZoneIndex)]
}

func normalizeZoneIndex(index int) int {
	if len(zonePresets) == 0 {
		return 0
	}
	if index < 0 {
		return len(zonePresets) - 1
	}
	if index >= len(zonePresets) {
		return index % len(zonePresets)
	}
	return index
}

func normalizeShape(shape clockcomp.DialShape) clockcomp.DialShape {
	switch shape {
	case clockcomp.DialShapeEllipse:
		return clockcomp.DialShapeEllipse
	default:
		return clockcomp.DialShapeCircle
	}
}

func normalizeGeometry(shape clockcomp.DialShape, radiusX, radiusY int) (clockcomp.DialShape, int, int) {
	shape = normalizeShape(shape)
	radiusX = clamp(radiusX, 3, 8)
	radiusY = clamp(radiusY, 3, 8)
	if shape == clockcomp.DialShapeCircle {
		radiusY = radiusX
	}
	return shape, radiusX, radiusY
}

func withGeometry(state AppState, shape clockcomp.DialShape, radiusX, radiusY int) AppState {
	state.Shape, state.RadiusX, state.RadiusY = normalizeGeometry(shape, radiusX, radiusY)
	return state
}

func adjustWidth(state AppState, delta int) AppState {
	if state.Shape == clockcomp.DialShapeCircle {
		next := clamp(state.RadiusX+delta, 3, 8)
		return withGeometry(state, state.Shape, next, next)
	}
	return withGeometry(state, state.Shape, state.RadiusX+delta, state.RadiusY)
}

func adjustHeight(state AppState, delta int) AppState {
	if state.Shape == clockcomp.DialShapeCircle {
		next := clamp(state.RadiusY+delta, 3, 8)
		return withGeometry(state, state.Shape, next, next)
	}
	return withGeometry(state, state.Shape, state.RadiusX, state.RadiusY+delta)
}

func dialCells(radiusX, radiusY int, cellAspectX float64) (int, int) {
	renderRadiusX := maxInt(1, int(float64(radiusX)*clampCellAspectX(cellAspectX)+0.5))
	return renderRadiusX*2 + 1, radiusY*2 + 1
}

func shapeHint(shape clockcomp.DialShape) string {
	if shape == clockcomp.DialShapeCircle {
		return "Circle mode locks width and height together and applies terminal cell-aspect compensation."
	}
	return "Ellipse mode allows width and height to change independently with the same horizontal compensation."
}

func heightHint(shape clockcomp.DialShape) string {
	if shape == clockcomp.DialShapeCircle {
		return "Locked to width in circle mode."
	}
	return "Vertical radius range: 3..8"
}

func clampCellAspectX(cellAspectX float64) float64 {
	if cellAspectX < 1.0 {
		return 1.0
	}
	if cellAspectX > 3.0 {
		return 3.0
	}
	return cellAspectX
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func mustLoadLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return location
}

func onOff(value bool) string {
	if value {
		return "On"
	}
	return "Off"
}

func colorName(c style.Color) string {
	return string(c)
}

func normalizePreset(preset clockcomp.Preset) clockcomp.Preset {
	for _, candidate := range clockPresets {
		if candidate == preset {
			return preset
		}
	}
	return clockcomp.PresetNone
}

func nextPreset(current clockcomp.Preset) clockcomp.Preset {
	for index, preset := range clockPresets {
		if preset == current {
			return clockPresets[(index+1)%len(clockPresets)]
		}
	}
	return clockPresets[0]
}

func nextShape(current clockcomp.DialShape) clockcomp.DialShape {
	for index, shape := range dialShapes {
		if shape == current {
			return dialShapes[(index+1)%len(dialShapes)]
		}
	}
	return dialShapes[0]
}

func handStyleLabel(handStyle clockcomp.HandRenderStyle) string {
	switch handStyle {
	case clockcomp.HandRenderStyleUnicode:
		return "Unicode"
	default:
		return "ASCII"
	}
}
