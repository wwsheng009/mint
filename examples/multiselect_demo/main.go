// Multi-Select Demo - Two approaches for implementing multi-select
//
// 演示两种多选实现方式：
// 1. OptionGroup 的 ModeMultiple (每个选项是独立的Fiber节点)
// 2. Checkbox 组合 (组合式组件)
//
// Architecture Note:
// - OptionGroup 选项现在是独立的 FocusableInstance
// - 每个选项可以通过 Tab 键单独获取焦点
// - 鼠标点击可以精确命中单个选项（通过 HitMap）
// - 导航动作由全局 FocusManager 处理，不再有焦点陷阱
//
// 运行: go run ./examples/multiselect_demo/

package main

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/statemachine"
	mintui "github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/divider"
	"github.com/wwsheng009/mint/ui/components/optiongroup"
)

// =============================================================================
// Application State
// =============================================================================

type AppState struct {
	// OptionGroup multi-select (comma-separated string)
	SelectedKills string

	// Checkbox combination (comma-separated for simplicity too)
	SelectedFeatures string

	Submitted bool
	ErrorMsg  string
}

// =============================================================================
// Intent Types
// =============================================================================

type SubmitIntent struct{}

func (SubmitIntent) IntentType() string { return "Submit" }

type ResetIntent struct{}

func (ResetIntent) IntentType() string { return "Reset" }

// FeatureSelectionIntent for individual checkboxes
type FeatureSelectionIntent struct {
	feature string
}

func (FeatureSelectionIntent) IntentType() string { return "FeatureSelection" }

// =============================================================================
// Reducer - Centralized Logic
// =============================================================================

// appReducerBuilder - FieldMap pattern
var appReducerBuilder = reducer.NewBuilder[AppState]()

func init() {
	// Handle field changes using FieldMap
	fieldMapBuilder := reducer.BindField(appReducerBuilder)
		fieldMapBuilder.BindFieldMap(map[string]func(AppState, string) AppState{
		"Kills": func(s AppState, val string) AppState {
			// Store as comma-separated string for OptionGroup
			s.SelectedKills = val
			return s
		},
	})

	// Extend from fieldMapBuilder to add action handlers
	appReducerBuilder = fieldMapBuilder.GetBuilder().
		On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
			if len(s.SelectedKills) == 0 && len(s.SelectedFeatures) == 0 {
				s.ErrorMsg = "Please select at least one option"
				return s
			}
			s.Submitted = true
			s.ErrorMsg = ""
			return s
		}).
		On(ResetIntent{}, func(s AppState, i intent.Intent) AppState {
			return AppState{
				SelectedKills:   "",
				SelectedFeatures: "",
				Submitted:          false,
				ErrorMsg:          "",
			}
		}).
		On(FeatureSelectionIntent{}, func(s AppState, i intent.Intent) AppState {
			if featureIntent, ok := i.(FeatureSelectionIntent); ok {
				feature := featureIntent.feature
				// Parse current features
				featureList := strings.Split(s.SelectedFeatures, ",")
				found := false
				for idx, f := range featureList {
					if f == feature {
						found = true
						// Toggle: remove if found
						if len(featureList) == 1 {
							s.SelectedFeatures = ""
						} else {
							featureList = append(featureList[:idx], featureList[idx+1:]...)
							s.SelectedFeatures = strings.Join(featureList, ",")
						}
						break
					}
				}
				if !found {
					// Add if not found
					if s.SelectedFeatures != "" {
						s.SelectedFeatures += ","
					}
					s.SelectedFeatures += feature
				}
			}
			return s
		})
}

// AppReducer is the built Reducer
var AppReducer = appReducerBuilder.Build()

// =============================================================================
// View Components
// =============================================================================

// AppView renders the application state
func AppView(state AppState) any {
	return renderAppView(state)
}

// renderAppView is the actual view function
func renderAppView(state AppState) mintui.VNode {
	// ==================== Method 1: OptionGroup (Component-based) ====================
	killsOptions := []optiongroup.Option{
		{Value: "fire", Label: "Fire 🔥"},
		{Value: "ice", Label: "Ice ❄️"},
		{Value: "thunder", Label: "Thunder ⚡"},
		{Value: "poison", Label: "Poison ☠️"},
		{Value: "light", Label: "Light 💡"},
	}

	// OptionGroup multi-select
	killsOptionGroup := optiongroup.NewBuilder(killsOptions).
		Key("kills-group").
		Label("Kills (Multi-select):").
		Mode(optiongroup.ModeMultiple).
		ForField(intent.BindField("Kills")).
		Vertical().
		Build()

	// ==================== Method 2: Checkbox Composition (Composable) ====================
	features := []struct {
		value  string
		label  string
	}{
		{"auto-save", "Auto-save 💾"},
		{"cloud-sync", "Cloud Sync ☁️"},
		{"dark-mode", "Dark Mode 🌙"},
		{"notifications", "Notifications 🔔"},
		{"sounds", "Sounds 🔊"},
		{"vibration", "Vibration 📳"},
	}

	// Checkbox composition (multi-select)
	var featureCheckboxes []mintui.VNode
	for _, feat := range features {
		featIntent := FeatureSelectionIntent{feature: feat.value}

		// Check if this feature is selected
		isSelected := false
		featureList := strings.Split(state.SelectedFeatures, ",")
		for _, f := range featureList {
			if f == feat.value {
				isSelected = true
				break
			}
		}

		cb := checkbox.NewBuilder().
			Label(feat.label).
			OnToggle(featIntent).
			Checked(isSelected).
			Build()
		featureCheckboxes = append(featureCheckboxes, cb)
	}

	// Build buttons
	submitButton := button.NewBuilder("Submit").
		OnPress(SubmitIntent{}).
		Build()

	resetButton := button.NewBuilder("Reset").
		OnPress(ResetIntent{}).
		Build()

	// Build layout
	var layout []mintui.VNode

	layout = append(layout,
		// Header
		mintui.NewTextBuilder("✅ Multi-Select Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		mintui.Text(""),
		mintui.NewTextBuilder("Two Implementation Approaches").
			FgColor("gray").
			Build(),
		mintui.Text(""),
		divider.NewBuilder().
			FillWidth(true).
			Build(),
	)

	// Show error message
	if state.ErrorMsg != "" {
		layout = append(layout,
			mintui.NewTextBuilder("⚠ "+state.ErrorMsg).
				FgColor("red").
				Build(),
			mintui.Text(""),
		)
	}

	// Method 1: OptionGroup multi-select
	layout = append(layout,
		mintui.NewTextBuilder("Method 1: OptionGroup (Fiber-first)").
			Bold(true).
			FgColor("yellow").
			Build(),
		mintui.Text("  • Each option is an independent Fiber node"),
		mintui.Text("  • Precise mouse targeting via HitMap"),
		mintui.Text("  • Tab navigation between options"),
		mintui.Text("  • Built-in multi-select mode"),
		mintui.Text(""),
		)

	layout = append(layout, killsOptionGroup)

	layout = append(layout, mintui.Text(""))

	// Method 2: Checkbox composition
	layout = append(layout,
		mintui.NewTextBuilder("Method 2: Checkbox Composition (Composable)").
			Bold(true).
			FgColor("green").
			Build(),
		mintui.Text("  • Each option is independent"),
		mintui.Text("  • Radio-group vs Multi-select based on reducer"),
		mintui.Text(""),
	)

	layout = append(layout, featureCheckboxes...)

	layout = append(layout, mintui.Text(""))

	// Buttons
	layout = append(layout,
		mintui.HStack(
			mintui.Text("  "),
			resetButton,
			mintui.Text(" "),
			submitButton,
		),
	)

	// Show result if submitted
	if state.Submitted {
		layout = append(layout,
			mintui.Text(""),
			divider.NewBuilder().
				FillWidth(true).
				Build(),
			mintui.NewTextBuilder("✅ Selection Submitted!").
				FgColor("green").
				Bold(true).
				Build(),
			divider.NewBuilder().
				FillWidth(true).
				Build(),

			// OptionGroup results
			mintui.NewTextBuilder("Kills:").
				FgColor("white").
				Build(),
			mintui.NewTextBuilder(state.SelectedKills).
				FgColor("cyan").
				Build(),
			mintui.Text(""),

			// Checkbox composition results
			mintui.NewTextBuilder("Features:").
				FgColor("white").
				Build(),
			mintui.NewTextBuilder(state.SelectedFeatures).
				FgColor("cyan").
				Build(),
		)
	} else {
		// Show current state
		layout = append(layout,
			mintui.Text(""),
			divider.NewBuilder().
				FillWidth(true).
				Build(),
			mintui.NewTextBuilder("Current Selection:").
				FgColor("gray").
				Build(),
			mintui.HStack(
				mintui.NewTextBuilder("  Kills:  ").
					FgColor("gray").
					Build(),
				mintui.NewTextBuilder(state.SelectedKills).
					FgColor("cyan").
					Build(),
			),
			mintui.HStack(
				mintui.NewTextBuilder("  Features: ").
					FgColor("gray").
					Build(),
				mintui.NewTextBuilder(state.SelectedFeatures).
					FgColor("cyan").
					Build(),
			),
		)
	}

	layout = append(layout,
		mintui.Text(""),
		divider.NewBuilder().
			FillWidth(true).
			Build(),
		mintui.NewTextBuilder("Comparison:").
			FgColor("gray").
			Build(),
		mintui.NewTextBuilder("┌─ OptionGroup (Fiber-first)").
			FgColor("white").
			Build(),
		mintui.NewTextBuilder("│  • State: comma-separated string (\"fire,ice,thunder\")").
			FgColor("gray").
			Build(),
		mintui.NewTextBuilder("│  • Architecture: each option = independent Fiber").
			FgColor("gray").
			Build(),
		mintui.NewTextBuilder("│  • Navigation: Tab between options (no focus trap)").
			FgColor("gray").
			Build(),
		mintui.NewTextBuilder("├─ Checkbox (Multiple)  ").
			FgColor("white").
			Build(),
		mintui.NewTextBuilder("│  • State: []string slice").
			FgColor("gray").
			Build(),
		mintui.NewTextBuilder("│  • Reducer: maintain slice directly").
			FgColor("gray").
			Build(),
		mintui.NewTextBuilder("└─ Choice depends on complexity").
			FgColor("gray").
			Build(),
		mintui.Text(""),
	)

	return mintui.VStack(layout...)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// Create initial state
	initialState := AppState{
		SelectedKills:   "",
		SelectedFeatures: "",
		Submitted:       false,
		ErrorMsg:        "",
	}

	// Create AppRuntime
	rt := statemachine.NewAppRuntime(
		initialState,
		AppView,
		AppReducer,
	)

	// Run the UI
	mintui.RunApp(rt,
		mintui.WithWidth(60),
		mintui.WithHeight(50),
		mintui.WithTitle("Multi-Select Demo"),
		mintui.WithInit(func() {
			// Register intent handlers
			appReducerBuilder.RegisterToGlobal(rt.GetStore())
		}),
	)
}
