package intent_test

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
)

// =============================================================================
// Example: Basic Intent Usage
// =============================================================================

func Example_basicIntent() {
	// Create a runtime
	rt := intent.NewRuntime()

	// Register a handler for OpenModalIntent
	intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i intent.OpenModalIntent) intent.IntentResult {
		fmt.Printf("Opening modal: %s\n", i.ModalID)
		ctx.SetState("showModal", true)
		return intent.HandledResult()
	})

	// Emit the intent
	result := rt.Emit(intent.OpenModal("settings-modal"))
	fmt.Printf("Handled: %v\n", result.Handled)

	// Output:
	// Opening modal: settings-modal
	// Handled: true
}

// =============================================================================
// Example: Intent with Priority
// =============================================================================

func Example_intentWithPriority() {
	// Create a runtime
	rt := intent.NewRuntime()

	// Register a handler for FocusIntent
	intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i intent.FocusIntent) intent.IntentResult {
		fmt.Printf("Focusing: %s (Priority: %s)\n", i.TargetID, i.Priority())
		return intent.HandledResult()
	})

	// Emit with automatic priority (Immediate for FocusIntent)
	result := rt.Emit(intent.Focus("username-input"))
	fmt.Printf("Handled: %v\n", result.Handled)

	// Output:
	// Focusing: username-input (Priority: Immediate)
	// Handled: true
}

// =============================================================================
// Example: Transition Intent (Async)
// =============================================================================

func Example_transitionIntent() {
	// Create a runtime
	rt := intent.NewRuntime()

	// Register a handler for LoadDataIntent (transition intent)
	// Note: Transition intents are queued and processed asynchronously
	intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i intent.LoadDataIntent) intent.IntentResult {
		return intent.HandledResult()
	})

	// Emit the transition intent
	// Transition intents are queued for async processing
	result := rt.Emit(intent.LoadData("/api/users", "users"))
	fmt.Printf("Handled: %v, Async: %v\n", result.Handled, result.Async)

	// Process the queue (simulating scheduler behavior)
	rt.Dispatcher.ProcessQueue(10 * time.Millisecond)

	// Output:
	// Handled: true, Async: true
}

// =============================================================================
// Example: Custom Intent Type
// =============================================================================

// Custom intent type
type IncrementCounterIntent struct {
	Step int
}

func (IncrementCounterIntent) IntentType() string {
	return "IncrementCounter"
}

func (IncrementCounterIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

func Example_customIntent() {
	// Create a runtime
	rt := intent.NewRuntime()

	// Register handler for custom intent
	intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i IncrementCounterIntent) intent.IntentResult {
		current, _ := ctx.GetState("counter")
		counter := 0
		if v, ok := current.(int); ok {
			counter = v
		}
		counter += i.Step
		ctx.SetState("counter", counter)
		fmt.Printf("Counter incremented by %d to %d\n", i.Step, counter)
		return intent.HandledResult()
	})

	// Emit custom intent
	rt.Emit(IncrementCounterIntent{Step: 5})
	rt.Emit(IncrementCounterIntent{Step: 3})

	// Output:
	// Counter incremented by 5 to 5
	// Counter incremented by 3 to 8
}

// =============================================================================
// Example: Intent Builder Pattern
// =============================================================================

func Example_intentBuilder() {
	// Create a runtime
	rt := intent.NewRuntime()

	intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i intent.NavigateIntent) intent.IntentResult {
		fmt.Printf("Navigating to: %s\n", i.Path)
		return intent.HandledResult()
	})

	// Use builder pattern
	result := intent.NewBuilder(intent.Navigate("/dashboard")).
		WithPriority(intent.PriorityUserBlocking).
		WithSource("main-menu").
		Dispatch(rt.Dispatcher)

	fmt.Printf("Handled: %v\n", result.Handled)

	// Output:
	// Navigating to: /dashboard
	// Handled: true
}

// =============================================================================
// Example: Transition Wrapper
// =============================================================================

// SaveDataIntent is a custom intent that can be wrapped as transition
type SaveDataIntent struct {
	Data string
}

func (SaveDataIntent) IntentType() string {
	return "SaveData"
}

func Example_transitionWrapper() {
	// Create an intent
	original := SaveDataIntent{Data: "test"}

	// Wrap it as a transition intent
	wrapped := intent.Transition(original)

	fmt.Printf("Is transition: %v\n", wrapped.IsTransition())
	fmt.Printf("Intent type: %s\n", wrapped.IntentType())

	// Output:
	// Is transition: true
	// Intent type: SaveData
}

// =============================================================================
// Example: Priority Wrapper
// =============================================================================

func Example_priorityWrapper() {
	// Create an intent with default priority
	original := intent.Navigate("/settings")

	// Override priority
	wrapped := intent.WithPriority(original, intent.PriorityIdle)

	fmt.Printf("New priority: %s\n", wrapped.Priority())

	// Output:
	// New priority: Idle
}

// =============================================================================
// Example: Middleware
// =============================================================================

func Example_middleware() {
	// Create a runtime with isolated registry for testing
	rt := intent.NewRuntimeWithNewRegistry()

	// Add logging middleware
	rt.Registry.Use(func(next intent.Handler) intent.Handler {
		return intent.HandlerFunc(func(ctx *intent.ActionContext, i intent.Intent) intent.IntentResult {
			fmt.Printf("[LOG] Before: %s\n", i.IntentType())
			result := next.Handle(ctx, i)
			fmt.Printf("[LOG] After: %s (handled: %v)\n", i.IntentType(), result.Handled)
			return result
		})
	})

	intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i intent.OpenModalIntent) intent.IntentResult {
		fmt.Println("Handler executed")
		return intent.HandledResult()
	})

	rt.Emit(intent.OpenModal("test"))

	// Output:
	// [LOG] Before: OpenModal
	// Handler executed
	// [LOG] After: OpenModal (handled: true)
}

// =============================================================================
// Example: Emitter for Components
// =============================================================================

func Example_emitter() {
	// Create a runtime with isolated registry for testing
	rt := intent.NewRuntimeWithNewRegistry()

	intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i intent.ClickIntent) intent.IntentResult {
		fmt.Printf("Click handled: %s\n", i.TargetID)
		return intent.HandledResult()
	})

	// Create an emitter (typically used by components)
	emitter := intent.NewEmitter(rt.Dispatcher, "button-1")

	// Emit intent with source tracking
	result := emitter.Emit(intent.Click("button-1"))
	fmt.Printf("Handled: %v\n", result.Handled)

	// Output:
	// Click handled: button-1
	// Handled: true
}
