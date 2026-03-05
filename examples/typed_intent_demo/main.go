// Typed Intent Demo - Demonstrates type-safe StateKey[T] and TypedFieldChange[T]
//
// This example shows:
// 1. Defining type-safe state keys with StateKey[T]
// 2. Using TypedFieldChange[T] for type-safe updates
// 3. Handling typed intents in a Reducer
// 4. The benefits of compile-time type checking
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Type-Safe State Keys
// =============================================================================

// Define all state keys in one place for type safety and discoverability.
// This eliminates string keys and provides compile-time type checking.
var (
	// Form fields with type safety
	Username = intent.NewStateKey[string]("username")
	Email    = intent.NewStateKey[string]("email")
	Age      = intent.NewStateKey[int]("age")
	Active   = intent.NewStateKey[bool]("active")

	// Error states
	UsernameErr = intent.NewStateKey[string]("username_err")
	EmailErr    = intent.NewStateKey[string]("email_err")
)

// =============================================================================
// Application State
// =============================================================================

type FormState struct {
	Username    string
	Email       string
	Age         int
	Active      bool
	UsernameErr string
	EmailErr    string
}

// =============================================================================
// Intent Types
// =============================================================================

// SubmitIntent triggers form submission
type SubmitIntent struct{}

func (SubmitIntent) IntentType() string { return "Submit" }

// ResetIntent clears the form
type ResetIntent struct{}

func (ResetIntent) IntentType() string { return "Reset" }

// ToggleActiveIntent toggles the active checkbox
type ToggleActiveIntent struct{}

func (ToggleActiveIntent) IntentType() string { return "ToggleActive" }

// IncrementAgeIntent increments age
type IncrementAgeIntent struct{}

func (IncrementAgeIntent) IntentType() string { return "IncrementAge" }

// DecrementAgeIntent decrements age
type DecrementAgeIntent struct{}

func (DecrementAgeIntent) IntentType() string { return "DecrementAge" }

// =============================================================================
// Reducer - Centralized Logic
// =============================================================================

// FormReducer handles all state changes in one place
func FormReducer(state FormState, i intent.Intent) FormState {
	switch v := i.(type) {
	// Type-safe string field changes
	case intent.TypedFieldChange[string]:
		switch v.Key.String() {
		case Username.String():
			state.Username = v.Value
			// Real-time validation
			if len(state.Username) < 3 {
				state.UsernameErr = "Username must be at least 3 characters"
			} else {
				state.UsernameErr = ""
			}
		case Email.String():
			state.Email = v.Value
			// Email validation
			if state.Email != "" && !isValidEmail(state.Email) {
				state.EmailErr = "Invalid email format"
			} else {
				state.EmailErr = ""
			}
		}

	// Type-safe int field changes
	case intent.TypedFieldChange[int]:
		if v.Key.String() == Age.String() {
			state.Age = v.Value
			if state.Age < 0 {
				state.Age = 0
			}
		}

	// Type-safe bool field changes
	case intent.TypedFieldChange[bool]:
		if v.Key.String() == Active.String() {
			state.Active = v.Value
		}

	// Action intents
	case SubmitIntent:
		// Validate all fields on submit
		hasError := false
		if len(state.Username) < 3 {
			state.UsernameErr = "Username must be at least 3 characters"
			hasError = true
		}
		if state.Email != "" && !isValidEmail(state.Email) {
			state.EmailErr = "Invalid email format"
			hasError = true
		}
		if !hasError {
			fmt.Printf("\n✅ Form submitted successfully!\n")
			fmt.Printf("   Username: %s\n", state.Username)
			fmt.Printf("   Email: %s\n", state.Email)
			fmt.Printf("   Age: %d\n", state.Age)
			fmt.Printf("   Active: %v\n\n", state.Active)
		}

	case ResetIntent:
		return FormState{}

	case ToggleActiveIntent:
		state.Active = !state.Active

	case IncrementAgeIntent:
		state.Age++

	case DecrementAgeIntent:
		state.Age--
		if state.Age < 0 {
			state.Age = 0
		}
	}

	return state
}

// =============================================================================
// View Components
// =============================================================================

func TypedFormDemo() ui.VNode {
	// Get component state using hooks
	username, setUsername := ui.UseStateString("")
	email, setEmail := ui.UseStateString("")
	age, setAge, _ := ui.UseStateInt(0)
	active, setActive := ui.UseStateBool(false)
	usernameErr, setUsernameErr := ui.UseStateString("")
	emailErr, setEmailErr := ui.UseStateString("")

	// Save state setters to context for handler access
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.SetState("setUsername", setUsername)
		ctx.SetState("setEmail", setEmail)
		ctx.SetState("setAge", setAge)
		ctx.SetState("setActive", setActive)
		ctx.SetState("setUsernameErr", setUsernameErr)
		ctx.SetState("setEmailErr", setEmailErr)
	}

	// Register handlers with ActionContext
	ui.On(SubmitIntent{}, func(ctx *intent.ActionContext) {
		// Validate on submit
		hasError := false
		if len(username) < 3 {
			setUsernameErr("Username must be at least 3 characters")
			hasError = true
		} else {
			setUsernameErr("")
		}
		if email != "" && !isValidEmail(email) {
			setEmailErr("Invalid email format")
			hasError = true
		} else {
			setEmailErr("")
		}
		if !hasError {
			fmt.Printf("\n✅ Submitted: %+v\n", map[string]interface{}{
				"username": username,
				"email":    email,
				"age":      age,
				"active":   active,
			})
		}
	})

	ui.On(ResetIntent{}, func(ctx *intent.ActionContext) {
		setUsername("")
		setEmail("")
		setAge(0)
		setActive(false)
		setUsernameErr("")
		setEmailErr("")
	})

	ui.On(ToggleActiveIntent{}, func(ctx *intent.ActionContext) {
		setActive(!active)
	})

	ui.On(IncrementAgeIntent{}, func(ctx *intent.ActionContext) {
		setAge(age + 1)
	})

	ui.On(DecrementAgeIntent{}, func(ctx *intent.ActionContext) {
		newAge := age - 1
		if newAge < 0 {
			newAge = 0
		}
		setAge(newAge)
	})

	// Handle typed field changes
	ui.On(intent.TypedFieldChange[string]{}, func(ctx *intent.ActionContext) {
		// This would be handled by a more sophisticated system
		// For demo, we show the pattern
	})

	return ui.VStack(
		ui.NewTextBuilder("📝 Type-Safe Intent Demo").Build(),
		ui.Text(""),

		// Username field
		ui.HStack(
			ui.NewTextBuilder("Username:").Build(),
			ui.Text("  "),
			ui.NewInputBuilder().
				Placeholder("Enter username").
				Value(username).
				Build(),
		),
		ui.NewTextBuilder(usernameErr).FgColor("red").Build(),

		// Email field
		ui.HStack(
			ui.NewTextBuilder("Email:").Build(),
			ui.Text("  "),
			ui.NewInputBuilder().
				Placeholder("Enter email").
				Value(email).
				Build(),
		),
		ui.NewTextBuilder(emailErr).FgColor("red").Build(),

		// Age field with +/- buttons
		ui.HStack(
			ui.NewTextBuilder("Age:").Build(),
			ui.Text("  "),
			ui.NewButtonBuilder(" - ").
				OnPress(DecrementAgeIntent{}).
				Build(),
			ui.Text(fmt.Sprintf("  %d  ", age)),
			ui.NewButtonBuilder(" + ").
				OnPress(IncrementAgeIntent{}).
				Build(),
		),

		// Active checkbox
		ui.HStack(
			ui.NewTextBuilder("Active:").Build(),
			ui.Text("  "),
			ui.NewButtonBuilder(fmt.Sprintf("[%s]", boolToCheck(active))).
				OnPress(ToggleActiveIntent{}).
				Build(),
		),

		ui.Text(""),

		// Action buttons
		ui.HStack(
			ui.NewButtonBuilder("Reset").
				OnPress(ResetIntent{}).
				Build(),
			ui.NewButtonBuilder("Submit").
				OnPress(SubmitIntent{}).
				Build(),
		),

		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.NewTextBuilder("Type-Safe Keys: Username, Email, Age, Active").FgColor("gray").Build(),
		ui.NewTextBuilder("Compile-time checking prevents typos!").FgColor("gray").Build(),
	)
}

// =============================================================================
// Helper Functions
// =============================================================================

func isValidEmail(email string) bool {
	return len(email) > 3 && contains(email, "@")
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func boolToCheck(b bool) string {
	if b {
		return "✓"
	}
	return " "
}

// =============================================================================
// Main
// =============================================================================

func main() {
	fmt.Println(`
╔════════════════════════════════════════════════════════════╗
║           Type-Safe Intent Demo                            ║
║                                                            ║
║  This demo shows StateKey[T] and TypedFieldChange[T]       ║
║  providing compile-time type safety for state management.  ║
║                                                            ║
║  Benefits:                                                 ║
║  • No more string key typos                                ║
║  • IDE autocomplete support                                ║
║  • Compile-time type checking                              ║
║  • Easier refactoring                                      ║
╚════════════════════════════════════════════════════════════╝
`)

	// Show type-safe key usage
	fmt.Println("Type-Safe State Keys:")
	fmt.Printf("  Username: %s (string)\n", Username.String())
	fmt.Printf("  Email:    %s (string)\n", Email.String())
	fmt.Printf("  Age:      %s (int)\n", Age.String())
	fmt.Printf("  Active:   %s (bool)\n", Active.String())
	fmt.Println()

	// Show how to create typed intents
	fmt.Println("Creating Typed Intents:")
	fmt.Printf("  Username.Change(\"alice\"): %v\n", Username.Change("alice"))
	fmt.Printf("  Age.Change(25): %v\n", Age.Change(25))
	fmt.Printf("  Active.Change(true): %v\n", Active.Change(true))
	fmt.Println()

	// Run the UI
	ui.Run(TypedFormDemo)
}
