// RadioGroup Demo - Demonstrates Composable Components
//
// This example shows how to create a radio group using individual Checkbox components
// composed with VStack, demonstrating the flexibility of the component system.
//
// 运行: go run ./examples/radiogroup_demo/

package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/statemachine"
	mintui "github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/divider"
)

// =============================================================================
// Application State
// =============================================================================

type AppState struct {
	SelectedCity  string
	SelectedTier string
	Submitted     bool
	ErrorMsg      string
}

// =============================================================================
// Intent Types
// =============================================================================

type SubmitIntent struct{}

func (SubmitIntent) IntentType() string { return "Submit" }

type ResetIntent struct{}

func (ResetIntent) IntentType() string { return "Reset" }

type CitySelectionIntent struct {
	city string
}

func (CitySelectionIntent) IntentType() string { return "CitySelection" }

type TierSelectionIntent struct {
	tier string
}

func (TierSelectionIntent) IntentType() string { return "TierSelection" }

// =============================================================================
// Reducer - Centralized Logic
// =============================================================================

// appReducerBuilder - FieldMap pattern
var appReducerBuilder = reducer.NewBuilder[AppState]()

func init() {
	// Handle field changes using FieldMap
	fieldMapBuilder := reducer.BindField(appReducerBuilder)
	fieldMapBuilder.BindFieldMap(map[string]func(AppState, string) AppState{
		"City": func(s AppState, val string) AppState {
			// Radio group logic: unselect other cities
			if val != "" && val != s.SelectedCity {
				s.SelectedCity = val
			} else if val == s.SelectedCity {
				// Toggle: unselect if clicking same option
				s.SelectedCity = ""
			}
			return s
		},
		"Tier": func(s AppState, val string) AppState {
			// Radio group logic for tier selection
			if val != "" {
				s.SelectedTier = val
			} else if val == s.SelectedTier {
				s.SelectedTier = ""
			}
			return s
		},
	})

	// Extend from fieldMapBuilder to add action handlers
	appReducerBuilder = fieldMapBuilder.GetBuilder().
		On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.SelectedCity == "" {
				s.ErrorMsg = "Please select a city"
				return s
			}
			if s.SelectedTier == "" {
				s.ErrorMsg = "Please select a tier"
				return s
			}
			s.Submitted = true
			s.ErrorMsg = ""
			return s
		}).
		On(ResetIntent{}, func(s AppState, i intent.Intent) AppState {
			return AppState{
				SelectedCity:  "",
				SelectedTier: "",
				Submitted:    false,
				ErrorMsg:     "",
			}
		}).
		On(CitySelectionIntent{}, func(s AppState, i intent.Intent) AppState {
			if cityIntent, ok := i.(CitySelectionIntent); ok {
				// Single-select logic for cities
				if cityIntent.city != "" && cityIntent.city != s.SelectedCity {
					s.SelectedCity = cityIntent.city
				} else if cityIntent.city == s.SelectedCity {
					// Toggle: unselect if clicking same option
					s.SelectedCity = ""
				}
			}
			return s
		}).
		On(TierSelectionIntent{}, func(s AppState, i intent.Intent) AppState {
			if tierIntent, ok := i.(TierSelectionIntent); ok {
				// Single-select logic for tiers
				if tierIntent.tier != "" && tierIntent.tier != s.SelectedTier {
					s.SelectedTier = tierIntent.tier
				} else if tierIntent.tier == s.SelectedTier {
					s.SelectedTier = ""
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
	// Define city options
	cities := []struct {
		value string
		label string
	}{
		{"bj", "Beijing"},
		{"sh", "Shanghai"},
		{"gz", "Guangzhou"},
		{"sz", "Shenzhen"},
	}

	// Define tier options
	tiers := []struct {
		value string
		label string
	}{
		{"free", "Free Tier"},
		{"basic", "Basic (￥99/mo)"},
		{"pro", "Pro (￥299/mo)"},
	}

	// Build city radio buttons (checkboxes that act as radio group)
	var cityCheckboxes []mintui.VNode
	for _, city := range cities {
		cb := checkbox.NewBuilder().
			Label(city.label).
			OnToggle(CitySelectionIntent{city: city.value}).
			Checked(state.SelectedCity == city.value).
			Build()
		cityCheckboxes = append(cityCheckboxes, cb)
	}

	// Build tier radio buttons
	var tierCheckboxes []mintui.VNode
	for _, tier := range tiers {
		cb := checkbox.NewBuilder().
			Label(tier.label).
			OnToggle(TierSelectionIntent{tier: tier.value}).
			Checked(state.SelectedTier == tier.value).
			Build()
		tierCheckboxes = append(tierCheckboxes, cb)
	}

	// Build buttons
	submitButton := mintui.NewButtonBuilder("Submit").
		OnPress(SubmitIntent{}).
		Build()

	resetButton := mintui.NewButtonBuilder("Reset").
		OnPress(ResetIntent{}).
		Build()

	// Build layout
	var layout []mintui.VNode

	layout = append(layout,
		// Header
		mintui.NewTextBuilder("📋 RadioGroup Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		mintui.Text(""),
		mintui.NewTextBuilder("Composable Components (Checkbox + VStack)").
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

	// City selection section
	layout = append(layout,
		mintui.NewTextBuilder("Select City:").
			Bold(true).
			FgColor("white").
			Build(),
	)
	layout = append(layout, cityCheckboxes...)

	layout = append(layout, mintui.Text(""))

	// Tier selection section
	layout = append(layout,
		mintui.NewTextBuilder("Select Tier:").
			Bold(true).
			FgColor("white").
			Build(),
	)
	layout = append(layout, tierCheckboxes...)

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
			mintui.NewTextBuilder("✅ Form Submitted!").
				FgColor("green").
				Bold(true).
				Build(),
			divider.NewBuilder().
				FillWidth(true).
				Build(),
			mintui.NewTextBuilder("City:   "+state.SelectedCity).
				FgColor("white").
				Build(),
			mintui.NewTextBuilder("Tier:   "+state.SelectedTier).
				FgColor("white").
				Build(),
		)
	} else {
		// Show current state
		layout = append(layout,
			mintui.Text(""),
			divider.NewBuilder().
				FillWidth(true).
				Build(),
			mintui.NewTextBuilder("Current State:").
				FgColor("gray").
				Build(),
			mintui.HStack(
				mintui.NewTextBuilder("  City:  ").
					FgColor("gray").
					Build(),
				mintui.NewTextBuilder(state.SelectedCity).
					FgColor("white").
					Build(),
			),
			mintui.HStack(
				mintui.NewTextBuilder("  Tier:  ").
					FgColor("gray").
					Build(),
				mintui.NewTextBuilder(state.SelectedTier).
					FgColor("white").
					Build(),
			),
		)
	}

	layout = append(layout,
		mintui.Text(""),
		divider.NewBuilder().
			FillWidth(true).
			Build(),
		mintui.NewTextBuilder("Features:").
			FgColor("gray").
			Build(),
		mintui.NewTextBuilder("✓ Composable components - Checkbox + VStack").
			FgColor("gray").
			Build(),
		mintui.NewTextBuilder("✓ Radio-group behavior via reducer logic").
			FgColor("gray").
			Build(),
		mintui.NewTextBuilder("✓ Store + Reducer integration").
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
		SelectedCity:  "",
		SelectedTier: "",
		Submitted:    false,
		ErrorMsg:     "",
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
		mintui.WithHeight(40),
		mintui.WithTitle("RadioGroup Demo"),
		mintui.WithInit(func() {
			// Register intent handlers
			appReducerBuilder.RegisterToGlobal(rt.GetStore())
		}),
	)
}
