package render

import (
	"testing"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Test Intent types for testing
type TestUpdateIntent struct {
	Key   string
	Value int
}
func (TestUpdateIntent) IntentType() string { return "TestUpdate" }

type UpdateStringIntent struct {
	Key   string
	Value string
}
func (UpdateStringIntent) IntentType() string { return "UpdateString" }

type UpdateIntIntent struct {
	Key   string
	Value int
}
func (UpdateIntIntent) IntentType() string { return "UpdateInt" }

// TestIntentToStateSync tests that Intent handlers update the same state that components read
func TestIntentToStateSync(t *testing.T) {

	// Track state changes

	// Create a simple component that reads state
	renderFn := func() rtui.VNode {
		ctx := rtui.GetCurrentContext()
		_ = ctx.GetIntState("step", 0)
		return rtui.Element("text").Prop("content", "Step").Build()
	}

	// Create an Intent handler
	intentUpdate := &TestUpdateIntent{Key: "step", Value: 5}

	// Create the declarative node
	fwApp := framework.NewApp()
	fwApp.SetConfigSize(10, 10)
	fwApp.Resize(10, 10)

	// Create the node with Fiber reconciler
	declarativeRoot := NewDeclarativeNodeFromFuncWithFiber(renderFn)
	declarativeRoot.SetApp(fwApp)
	// Initialize Intent Runtime
	intentRuntime := intent.NewRuntime()
	rtui.SetGlobalIntentRuntime(intentRuntime)

	// Set Intent Runtime on declarative node (this connects StateSetter)
	SetDeclarativeNodeIntentRuntime(declarativeRoot, intentRuntime)

	// Register the Intent handler
	intentRuntime.Register("TestUpdate", intent.HandlerFunc(func(ctx *intent.ActionContext, i intent.Intent) intent.IntentResult {
		ti := i.(*TestUpdateIntent)
		ctx.SetState(ti.Key, ti.Value)
		return intent.HandledResult()
	}))

	// Emit the Intent (simulating a button click or user action)
	intentRuntime.Emit(intentUpdate)

	// Verify state was updated by directly checking the instance state
	if declarativeRoot.instance.GetIntState("step", 0) != 5 {
		t.Fatalf("Expected state value 5, got %d", declarativeRoot.instance.GetIntState("step", 0))
	}

	t.Log("✓ Intent Handler successfully updated ComponentContext state")
}

// TestMultipleStateUpdates tests multiple state updates
// TestMultipleStateUpdates tests multiple state updates
func TestMultipleStateUpdates(t *testing.T) {
	renderFn := func() rtui.VNode {
		ctx := rtui.GetCurrentContext()
		_ = ctx.GetStringState("name", "")
		_ = ctx.GetIntState("age", 0)
		return rtui.Element("text").Prop("content", "Test").Build()
	}

	fwApp := framework.NewApp()
	fwApp.SetConfigSize(10, 10)
	fwApp.Resize(10, 10)

	declarativeRoot := NewDeclarativeNodeFromFuncWithFiber(renderFn)
	declarativeRoot.SetApp(fwApp)
	intentRuntime := intent.NewRuntime()
	rtui.SetGlobalIntentRuntime(intentRuntime)
	SetDeclarativeNodeIntentRuntime(declarativeRoot, intentRuntime)

	// Register handlers
	intentRuntime.Register("UpdateString", intent.HandlerFunc(func(ctx *intent.ActionContext, i intent.Intent) intent.IntentResult {
		ui := i.(*UpdateStringIntent)
		ctx.SetState(ui.Key, ui.Value)
		return intent.HandledResult()
	}))
	intentRuntime.Register("UpdateInt", intent.HandlerFunc(func(ctx *intent.ActionContext, i intent.Intent) intent.IntentResult {
		ui := i.(*UpdateIntIntent)
		ctx.SetState(ui.Key, ui.Value)
		return intent.HandledResult()
	}))

	// Emit multiple Intents
	intentRuntime.Emit(&UpdateStringIntent{Key: "name", Value: "Alice"})
	intentRuntime.Emit(&UpdateIntIntent{Key: "age", Value: 30})

	// Verify state was updated by directly checking the instance state
	if declarativeRoot.instance.GetStringState("name", "") != "Alice" {
		t.Fatalf("Expected name 'Alice', got '%s'", declarativeRoot.instance.GetStringState("name", ""))
	}
	if declarativeRoot.instance.GetIntState("age", 0) != 30 {
		t.Fatalf("Expected age 30, got %d", declarativeRoot.instance.GetIntState("age", 0))
	}

	t.Log("✓ Multiple Intent updates successful")
}
