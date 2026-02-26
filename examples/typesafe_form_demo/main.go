// Package main demonstrates type-safe form handling with StateKey[T].
//
// This example shows the full MVP data flow with compile-time type safety:
//   1. Define StateKey[T] package-level keys (type-safe field identifiers)
//   2. Use ForField() to bind components to StateKey[T]
//   3. Register a FieldChangeIntent handler
//   4. Instances emit intents carrying values
//   5. State updates become the single source of truth
//
// Benefits:
//   - No string key typos
//   - IDE autocomplete support
//   - Refactor-safe field renames
//   - Type-checked values
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/input"
)

// =============================================================================
// Type-Safe State Keys (Package Level)
// =============================================================================

// Form state keys - these provide compile-time type safety.
var (
	keyUsername = intent.StateKey[string]("username")
	keyEmail    = intent.StateKey[string]("email")

	keyAcceptTerms = intent.StateKey[bool]("acceptTerms")
	keySubscribe   = intent.StateKey[bool]("subscribe")

	keyAge = intent.StateKey[int]("age")
)

// =============================================================================
// Submit Intent
// =============================================================================

type SubmitIntent struct{}

func (SubmitIntent) IntentType() string { return "Submit" }

// =============================================================================
// Form Component Builders
// =============================================================================

func App() ui.VNode {
	// Register a single handler for FieldChangeIntent
	ui.RegisterIntent(func(actx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
		fmt.Printf("Intent Received: Field='%s', Value='%s'\n", i.Field, i.Value)

		// State becomes the single source of truth
		actx.SetState(i.Field, i.Value)
		actx.ScheduleUpdate()
		return intent.HandledResult()
	})

	// Submit handler
	ui.RegisterIntent(func(actx *intent.ActionContext, i SubmitIntent) intent.IntentResult {
		username, _ := actx.GetState(keyUsername.String())
		email, _ := actx.GetState(keyEmail.String())
		age, _ := actx.GetState(keyAge.String())
		acceptTerms, _ := actx.GetState(keyAcceptTerms.String())
		subscribe, _ := actx.GetState(keySubscribe.String())

		fmt.Println("\n=== Form Submission ===")
		fmt.Printf("Username:   %v\n", username)
		fmt.Printf("Email:      %v\n", email)
		fmt.Printf("Age:        %v\n", age)
		fmt.Printf("Accept T&C: %v\n", acceptTerms)
		fmt.Printf("Subscribe:  %v\n", subscribe)
		fmt.Println("========================\n")

		return intent.HandledResult()
	})

	// Define form components using type-safe bindings
	usernameInput := input.NewBuilder().
		Placeholder("Enter your username").
		ForField(intent.ForField(keyUsername)).
		Width(30).
		Build()

	emailInput := input.NewBuilder().
		Placeholder("Enter your email").
		ForField(intent.ForField(keyEmail)).
		Width(30).
		Build()

	ageInput := input.NewBuilder().
		Placeholder("Enter your age").
		ForField(intent.ForField(keyAge)).
		Width(10).
		Build()

	termsCheckbox := checkbox.NewBuilder().
		Label("I accept the terms and conditions").
		ForField(intent.ForField(keyAcceptTerms)).
		Build()

	subscribeCheckbox := checkbox.NewBuilder().
		Label("Subscribe to newsletter").
		ForField(intent.ForField(keySubscribe)).
		Build()

	submitButton := button.NewBuilder("Submit Form").
		OnPress(SubmitIntent{}).
		Variant(button.VariantPrimary).
		Build()

	// Build the form layout
	form := ui.VStack(
		ui.Text("=== Type-Safe Registration Form ==="),
		ui.Text(""),

		ui.HStack(ui.Text("Username: "), ui.Text("  "), usernameInput),
		ui.HStack(ui.Text("Email:    "), ui.Text("  "), emailInput),
		ui.HStack(ui.Text("Age:      "), ui.Text("  "), ageInput),

		ui.Text(""),

		termsCheckbox,
		subscribeCheckbox,

		ui.Text(""),

		ui.HStack(ui.Text("  "), ui.Text("  "), submitButton),
	)

	return form
}

func main() {
	ui.Run(App,
		ui.WithWidth(50),
		ui.WithHeight(18),
		ui.WithTitle("Type-Safe Form Demo"),
		ui.WithInit(func() {
			// Initialize state with action context
			// State will be initialized automatically on first render
		}),
	)
}
