// Typed Intent Demo - Demonstrates Store + Reducer architecture
//
// This example shows type-safe form handling with FieldMap and ForField
// MIGRATED from hooks-based API to Store + Reducer
package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/statemachine"
	mintui "github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/divider"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/optiongroup"
)

// =============================================================================
// Application State
// =============================================================================

type FormState struct {
	Username string
	Email    string
	Age      int
	Active   bool
	City     string       // OptionGroup single-select
	Interests string       // OptionGroup multi-select (comma-separated)
	ErrorMsg string
	Submitted bool
	SubmitTime time.Time
}

// =============================================================================
// Intent Types
// =============================================================================

type SubmitIntent struct{}

func (SubmitIntent) IntentType() string { return "Submit" }

type ResetIntent struct{}

func (ResetIntent) IntentType() string { return "Reset" }

type IncrementAgeIntent struct{}

func (IncrementAgeIntent) IntentType() string { return "IncrementAge" }

type DecrementAgeIntent struct{}

func (DecrementAgeIntent) IntentType() string { return "DecrementAge" }

// =============================================================================
// Reducer - Centralized Logic
// =============================================================================

// appReducerBuilder - FieldMap + pattern
var appReducerBuilder = reducer.NewBuilder[FormState]()

func init() {
	// Handle field changes using FieldMap
	fieldMapBuilder := reducer.BindField(appReducerBuilder)
	fieldMapBuilder.BindFieldMap(map[string]func(FormState, string) FormState{
		"Username": func(s FormState, val string) FormState {
			s.Username = val
			// Real-time validation
			if len(s.Username) < 3 && len(s.Username) > 0 {
				s.ErrorMsg = "Username must be at least 3 characters"
			} else {
				// Clear username error if it exists
				if strings.HasPrefix(s.ErrorMsg, "Username") {
					s.ErrorMsg = ""
				}
			}
			return s
		},
		"Email": func(s FormState, val string) FormState {
			s.Email = val
			// Email validation
			if s.Email != "" && !strings.Contains(s.Email, "@") {
				s.ErrorMsg = "Invalid email format"
			} else if len(s.ErrorMsg) == 0 || strings.HasPrefix(s.ErrorMsg, "Username") {
				s.ErrorMsg = ""
			}
			return s
		},
		"Age": func(s FormState, val string) FormState {
			age, err := reducer.ParseInt(val)
			if err == nil {
				s.Age = age
				if s.Age < 0 {
					s.Age = 0
				}
			}
			return s
		},
		"Active": func(s FormState, val string) FormState {
			s.Active = val == "true" || (strings.ToLower(val) == "on")
			return s
		},
		"City": func(s FormState, val string) FormState {
			s.City = val
			return s
		},
		"Interests": func(s FormState, val string) FormState {
			s.Interests = val
			return s
		},
	})

	// Extend from fieldMapBuilder to add action handlers
	appReducerBuilder = fieldMapBuilder.GetBuilder().
		On(SubmitIntent{}, func(s FormState, i intent.Intent) FormState {
			// Validate all fields on submit
			if len(s.Username) < 3 {
				s.ErrorMsg = "Username must be at least 3 characters"
				return s
			}
			if len(s.Email) == 0 || !strings.Contains(s.Email, "@") {
				s.ErrorMsg = "Invalid email format"
				return s
			}
			s.Submitted = true
			s.SubmitTime = time.Now()
			return s
		}).
		On(ResetIntent{}, func(s FormState, i intent.Intent) FormState {
			return FormState{}
		}).
		On(IncrementAgeIntent{}, func(s FormState, i intent.Intent) FormState {
			s.Age++
			return s
		}).
		On(DecrementAgeIntent{}, func(s FormState, i intent.Intent) FormState {
			s.Age--
			if s.Age < 0 {
				s.Age = 0
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
func AppView(state FormState) any {
	return renderFormView(state)
}

// renderFormView is the actual view function
func renderFormView(state FormState) mintui.VNode {
	// Build form components with ForField binding
	usernameInput := input.NewBuilder().
		Placeholder("Type username").
		ForField(intent.BindField("Username")).
		Value(state.Username).
		Width(30).
		Build()

	emailInput := input.NewBuilder().
		Placeholder("Enter email").
		ForField(intent.BindField("Email")).
		Value(state.Email).
		Width(30).
		Build()

	ageInput := input.NewBuilder().
		Placeholder("Enter age").
		ForField(intent.BindField("Age")).
		Value(reducer.FormatInt(state.Age)).
		Width(10).
		Build()

	activeCheckbox := checkbox.NewBuilder().
		Label("Active").
		ForField(intent.BindField("Active")).
		Checked(state.Active).
		Build()

	submitButton := button.NewBuilder("Submit").
		OnPress(SubmitIntent{}).
		Build()

	resetButton := button.NewBuilder("Reset").
		OnPress(ResetIntent{}).
		Build()

	incAgeButton := button.NewBuilder(" + ").
		OnPress(IncrementAgeIntent{}).
		Build()

	decAgeButton := button.NewBuilder(" - ").
		OnPress(DecrementAgeIntent{}).
		Build()

	// Build the form layout
	var layout []mintui.VNode

	layout = append(layout,
		// Header
		mintui.NewTextBuilder("📝 Typed Intent Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		mintui.Text(""),
		mintui.NewTextBuilder("Store + Reducer Architecture").
			FgColor("gray").
			Build(),
		mintui.Text(""),
		divider.NewBuilder().
			FillWidth(true).
			Build(),
	)

	// Show error message if any
		if state.ErrorMsg != "" {
			layout = append(layout,
				mintui.NewTextBuilder("⚠ "+state.ErrorMsg).
					FgColor("red").
					Build(),
				mintui.Text(""),
			)
		}

		// Form fields using ForField binding
		layout = append(layout,
			mintui.HStack(
				mintui.Text("Username: "),
				mintui.Text("  "),
				usernameInput,
			),

			mintui.HStack(
				mintui.Text("Email:    "),
				mintui.Text("  "),
				emailInput,
			),

			mintui.HStack(
				mintui.Text("Age:      "),
				mintui.Text("  "),
				ageInput,
				mintui.Text(" "),
				decAgeButton,
				mintui.Text(" "),
				incAgeButton,
			),

			mintui.HStack(
				mintui.Text("  "),
				mintui.Text("  "),
				activeCheckbox,
			),

			mintui.Text(""),

			// OptionGroup - Single Select (City)
			optiongroup.NewBuilder([]optiongroup.Option{
				{Value: "bj", Label: "Beijing"},
				{Value: "sh", Label: "Shanghai"},
				{Value: "gz", Label: "Guangzhou"},
				{Value: "sz", Label: "Shenzhen"},
			}).
				Key("city-group").
				Label("City:").
				Mode(optiongroup.ModeSingle).
				ForField(intent.BindField("City")).
				Selected(state.City).
				Vertical().
				Build(),

			mintui.Text(""),

			// OptionGroup - Multiple Select (Interests)
			optiongroup.NewBuilder([]optiongroup.Option{
				{Value: "dev", Label: "Development"},
				{Value: "design", Label: "Design"},
				{Value: "test", Label: "Testing"},
				{Value: "pm", Label: "Product Management"},
			}).
				Key("interests-group").
				Label("Interests:").
				Mode(optiongroup.ModeMultiple).
				ForField(intent.BindField("Interests")).
				Selecteds(strings.Split(state.Interests, ",")).
				Vertical().
				Build(),

			mintui.Text(""),

			// Action buttons
			mintui.HStack(
				mintui.Text("  "),
				resetButton,
				mintui.Text(" "),
				submitButton,
			),
		)

		// State display
		if !state.Submitted {
			layout = append(layout,
				mintui.Text(""),
				divider.NewBuilder().
					FillWidth(true).
					Build(),
				mintui.NewTextBuilder("Current State:").
					FgColor("gray").
					Build(),
				mintui.HStack(
					mintui.NewTextBuilder("  Username: ").
						FgColor("gray").
						Build(),
					mintui.NewTextBuilder(fmt.Sprintf("%q", state.Username)).
						FgColor("white").
						Build(),
				),
				mintui.HStack(
					mintui.NewTextBuilder("  Email:    ").
						FgColor("gray").
						Build(),
					mintui.NewTextBuilder(fmt.Sprintf("%q", state.Email)).
						FgColor("white").
						Build(),
				),
				mintui.HStack(
					mintui.NewTextBuilder("  Age:      ").
						FgColor("gray").
						Build(),
					mintui.NewTextBuilder(fmt.Sprintf("%d", state.Age)).
						FgColor("white").
						Build(),
				),
				mintui.HStack(
					mintui.NewTextBuilder("  Active:   ").
						FgColor("gray").
						Build(),
					mintui.NewTextBuilder(fmt.Sprintf("%v", state.Active)).
						FgColor("white").
						Build(),
				),
				mintui.HStack(
					mintui.NewTextBuilder("  City:     ").
						FgColor("gray").
						Build(),
					mintui.NewTextBuilder(fmt.Sprintf("%q", state.City)).
						FgColor("white").
						Build(),
				),
				mintui.HStack(
					mintui.NewTextBuilder("  Interests:").
						FgColor("gray").
						Build(),
					mintui.NewTextBuilder(fmt.Sprintf("%q", state.Interests)).
						FgColor("white").
						Build(),
				),
			)
		} else {
			layout = append(layout,
				mintui.Text(""),
				divider.NewBuilder().
					FillWidth(true).
					Build(),
				mintui.NewTextBuilder("✅ Form Submitted Successfully!").
					FgColor("green").
					Build(),
				divider.NewBuilder().
					FillWidth(true).
					Build(),
				mintui.NewTextBuilder("Username:       "+state.Username).
					FgColor("white").
					Build(),
				mintui.NewTextBuilder("Email:          "+state.Email).
					FgColor("white").
					Build(),
				mintui.NewTextBuilder("Age:             "+fmt.Sprintf("%d", state.Age)).
					FgColor("white").
					Build(),
				mintui.NewTextBuilder("Active:          "+fmt.Sprintf("%v", state.Active)).
					FgColor("white").
					Build(),
				mintui.NewTextBuilder("City:           "+state.City).
					FgColor("white").
					Build(),
				mintui.NewTextBuilder("Interests:      "+state.Interests).
					FgColor("white").
					Build(),
				mintui.NewTextBuilder("Submitted at:     "+state.SubmitTime.Format("15:04:05")).
					FgColor("gray").
					Build(),
			)
		}

		layout = append(layout,
			mintui.Text(""),
			divider.NewBuilder().
				FillWidth(true).
				Build(),
			mintui.NewTextBuilder("Type-Safe Features:").
				FgColor("gray").
				Build(),
			mintui.NewTextBuilder("✓ FieldMap for automatic field updates").
				FgColor("gray").
				Build(),
			mintui.NewTextBuilder("✓ Centralized logic in Reducer").
				FgColor("gray").
				Build(),
			mintui.NewTextBuilder("✓ Compile-time type checking").
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
	// ============================================================
	// CPU Profiling Setup (用于性能分析)
	// ============================================================
	// 检查是否启用 CPU profiling
	if os.Getenv("MINT_CPU_PROFILE") == "true" {
		// 启动 pprof HTTP 服务器 (用于在线分析)
		// 访问 http://localhost:6060/debug/pprof/ 查看
		go func() {
			fmt.Println("pprof server listening on :6060")
			fmt.Println("Access http://localhost:6060/debug/pprof/")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				fmt.Printf("pprof server error: %v\n", err)
			}
		}()

		// 如果设置了 CPU profile 输出文件，自动采样
		profileFile := os.Getenv("MINT_CPU_PROFILE_FILE")
		if profileFile != "" {
			duration := 30 // 默认采样30秒
			if d := os.Getenv("MINT_CPU_PROFILE_DURATION"); d != "" {
				fmt.Sscanf(d, "%d", &duration)
			}

			f, err := os.Create(profileFile)
			if err != nil {
				fmt.Printf("Could not create CPU profile: %v\n", err)
				return
			}
			defer f.Close()

			if err := pprof.StartCPUProfile(f); err != nil {
				fmt.Printf("Could not start CPU profile: %v\n", err)
				return
			}
			defer pprof.StopCPUProfile()

			fmt.Printf("CPU profiling enabled: sampling for %d seconds to %s\n", duration, profileFile)

			// 在指定时长后自动退出
			go func() {
				time.Sleep(time.Duration(duration) * time.Second)
				fmt.Println("CPU profiling completed, exiting...")
				pprof.StopCPUProfile()
				os.Exit(0)
			}()
		}
	}

	// Create initial state
	initialState := FormState{
		Username:   "",
		Email:      "",
		Age:        0,
		Active:     false,
		City:       "bj", // Default select Beijing
		Interests:  "",
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
		mintui.WithHeight(35),
		mintui.WithTitle("Typed Intent Demo (Store+Reducer)"),
		mintui.WithInit(func() {
			// Register intent handlers
			appReducerBuilder.RegisterToGlobal(rt.GetStore())
		}),
	)
}
