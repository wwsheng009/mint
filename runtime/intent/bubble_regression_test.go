package intent_test

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/form"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
)

// =============================================================================
// Intent Bubble Regression Tests
// These tests verify the problems found in INTENT_BUBBLE_AUDIT_REPORT.md
// =============================================================================

// -----------------------------------------------------------------------------
// Test 1: Verify FiberUtil IntentEmitter Wiring Issue
// Problem: FiberUtil connects IntentEmitter to Global Intent Runtime instead of Bubble System
// -----------------------------------------------------------------------------

// TestFiberUtil_IntentEmitterWiring verifies that IntentEmitter is not wired correctly.
//
// EXPECTED BEHAVIOR (after fix):
// - IntentEmitter should call intent.Emit(component, intent) to bubble
// - Intent should propagate up the Instance Tree through Parent()
//
// ACTUAL BEHAVIOR (before fix):
// - IntentEmitter calls Global Intent Runtime
// - Intent does NOT bubble through Parent() chain
// - This is P0 blocking issue
func TestFiberUtil_IntentEmitterWiring(t *testing.T) {
	t.Skip("Waiting for fix - see P0-1 in INTENT_BUBBLE_AUDIT_REPORT.md")

	// Setup: Create a parent and child with proper parent reference
	parent := createMockComponent("parent")
	child := createMockComponent("child")

	// Set parent reference (simulating what should happen in AddChild)
	parent.AddChild(child)

	// Verify parent reference is set
	if child.Parent() == nil {
		t.Fatal("Child should have parent reference set")
	}

	// Track intent emission at parent level
	parentReceivedIntent := false
	intent.SetBubbleTestHook(func(comp interface{}, i intent.Intent) bool {
		// This hook would be called if intent.Emit(component, intent) is working
		if comp == parent {
			parentReceivedIntent = true
		}
		return false
	})
	defer intent.SetBubbleTestHook(nil)

	// Emit intent from child (after fix, this should bubble)
	intent.Emit(child, selectcomp.SelectChange(0, "value1", "Label 1"))

	// ASSERT: Parent should receive the intent via bubble
	if !parentReceivedIntent {
		t.Error("FAIL: Parent should have received bubbled intent from child")
		t.Error("This indicates FiberUtil IntentEmitter is NOT calling intent.Emit(component, intent)")
		t.Error("Currently it's calling global runtime.Emit() instead")
	}

	t.Log("PASS: IntentEmitter correctly bubbles intentions to parent")
}

// -----------------------------------------------------------------------------
// Test 2: Verify Form.AddChild Does Not Set Parent Reference
// Problem: Form.AddChild only adds to childInstances, doesn't set child.parent
// -----------------------------------------------------------------------------

// TestForm_AddChild_ParentReference verifies that Form.AddChild sets parent reference.
//
// EXPECTED BEHAVIOR (after fix):
// - Form.AddChild should call child.SetParent(form) or set childBase.parent = form
// - Child's Parent() should return the Form instance
//
// ACTUAL BEHAVIOR (before fix):
// - Form.AddChild only appends to childInstances slice
// - Child's Parent() returns nil (parent reference is not set)
// - This is P0 blocking issue
func TestForm_AddChild_ParentReference(t *testing.T) {
	t.Run("Form should set parent reference on child", func(t *testing.T) {
		// Setup: Create Form and Select instances
		formProps := rtui.Props{
			"key":  "test-form",
			"label": "Test Form",
		}
		formInst := form.NewInstance(formProps)

		selectProps := rtui.Props{
			"key":      "test-select",
			"options": []selectcomp.Option{
				{Value: "1", Label: "Option 1"},
			},
			"selectedIndex": 0,
		}
		selectInst := selectcomp.NewInstance(selectProps)

		// Add select to form
		formInst.AddChild(selectInst)

		// ACTUAL BEHAVIOR (before fix): Parent() returns nil
		// EXPECTED BEHAVIOR (after fix): Parent() returns formInst

		if selectInst.Parent() == nil {
			t.Error("FAIL: Form.AddChild does NOT set parent reference")
			t.Error("Expected: selectInst.Parent() should return formInst")
			t.Error("Actual: selectInst.Parent() returns nil")
			t.Error("This prevents intent bubbles from working - see P0-2 in INTENT_BUBBLE_AUDIT_REPORT.md")

			// Additional verification: Check if formInst is in form's childInstances
			children := formInst.Children()
			if len(children) != 1 {
				t.Error("Form should have 1 child")
			}
			if len(children) > 0 && children[0] != selectInst {
				t.Error("Form's child should be selectInst")
			}
		} else {
			t.Log("PASS: Form.AddChild sets parent reference correctly")
		}
	})

	t.Run("Verify reference is two-way", func(t *testing.T) {
		// Setup
		formProps := rtui.Props{"key": "form-2"}
		formInst := form.NewInstance(formProps)

		selectProps := rtui.Props{
			"key":      "select-2",
			"options": []selectcomp.Option{{Value: "a", Label: "A"}},
			"selectedIndex": 0,
		}
		selectInst := selectcomp.NewInstance(selectProps)

		formInst.AddChild(selectInst)

		// After fix, both directions should work
		parentOfChild := selectInst.Parent()
		childrenOfParent := formInst.Children()

		if len(childrenOfParent) == 0 {
			t.Error("Form should have children")
		}

		if len(childrenOfParent) > 0 && parentOfChild != formInst {
			t.Error("Child's parent should point to form")
			t.Errorf("Expected parent: formInst, got: %v", parentOfChild)
		}
	})
}

// -----------------------------------------------------------------------------
// Test 3: Verify Real Intent Bubble Flow
// Problem: Current tests use manual intent collection, not real bubble
// -----------------------------------------------------------------------------

// TestRealIntentBubble_Flow verifies the actual bubble flow through Parent() chain.
//
// EXPECTED BEHAVIOR (after fix):
// - Component calls intent.Emit(self, intent)
// - Bubbling propagates through Parent() chain
// - Each parent's HandleIntent is called (if implemented)
// - Stops when HandleIntent returns true or reaches root
//
// ACTUAL BEHAVIOR (before fix):
// - Intent goes to global runtime, not parent chain
// - Parent components never see the intent
// - intent.Emit() doesn't actually bubble
func TestRealIntentBubble_Flow(t *testing.T) {
	t.Skip("Waiting for fix - see P1-3 in INTENT_BUBBLE_AUDIT_REPORT.md")

	t.Run("Intent bubbles through Parent() chain", func(t *testing.T) {
		// Create chain: Grandparent -> Parent -> Child
		grandparent := createMockComponent("grandparent")
		parent := createMockComponent("parent")
		child := createMockComponent("child")

		// Set up parent references
		parent.AddChild(child)
		grandparent.AddChild(parent)

		// Verify parent references are set
		if child.Parent() != grandparent && child.Parent() != parent {
			t.Fatalf("Child should have grandparent or parent as parent, got %v", child.Parent())
		}

		// Track HandleIntent calls
		var callOrder []string
		grandparent.handleFunc = func(intent.Intent) bool {
			callOrder = append(callOrder, "grandparent")
			return true // Stop bubbling
		}
		parent.handleFunc = func(intent.Intent) bool {
			callOrder = append(callOrder, "parent")
			return false // Continue bubbling
		}
		child.handleFunc = func(intent.Intent) bool {
			callOrder = append(callOrder, "child")
			return false // Continue bubbling
		}

		// Emit from child
		testIntent := selectcomp.SelectChange(0, "val", "lbl")
		intent.Emit(child, testIntent)

		// After fix, bubbling should occur in order: child -> parent -> grandparent
		if len(callOrder) == 0 {
			t.Fatal("FAIL: No HandleIntent was called")
			t.Error("This means intent.Emit() is not bubbling through Parent() chain")
			t.Error("Intent is likely going to global runtime instead")
		}

		expectedOrder := []string{"child", "parent", "grandparent"}
		if !sliceEqual(callOrder, expectedOrder) {
			t.Errorf("FAIL: Bubble order incorrect")
			t.Errorf("Expected: %v", expectedOrder)
			t.Errorf("Actual: %v", callOrder)
		}

		t.Log("PASS: Intent bubbles correctly through Parent() chain")
	})
}

// -----------------------------------------------------------------------------
// Test 4: Verify ComponentID Routing
// Problem: ComponentID routing is implemented separately in each component
// -----------------------------------------------------------------------------

// TestComponentID_Routing verifies Component-based routing works consistently.
//
// EXPECTED BEHAVIOR:
// - Components with componentID should only handle intents with matching ID
// - Components without componentID should handle all intents
// - Routing logic should be consistent across all components
//
// ACTUAL BEHAVIOR:
// - Each component implements routing separately
// - Inconsistent implementations possible
// - This is P1 issue (see P1-4 in INTENT_BUBBLE_AUDIT_REPORT.md)
func TestComponentID_Routing(t *testing.T) {
	t.Run("Component with ID filters intents", func(t *testing.T) {
		// Setup: Create Select with componentID
		selectProps := rtui.Props{
			"key":         "select-1",
			"componentID": "field-username",
			"options": []selectcomp.Option{
				{Value: "user1", Label: "User 1"},
				{Value: "user2", Label: "User 2"},
			},
			"selectedIndex": 0,
		}
		selectInst := selectcomp.NewInstance(selectProps)

		// Test 1: Intent with matching componentID should be handled
		intent1 := selectcomp.SelectByIndexWithID("field-username", 1)
		handled1 := selectInst.HandleIntent(intent1)

		if !handled1 {
			t.Error("Select should handle intent with matching componentID")
		}

		t.Log("PASS: Matching componentID intent is handled")

		// Test 2: Intent with non-matching componentID should NOT be handled
		intent2 := selectcomp.SelectByIndexWithID("field-email", 0)
		handled2 := selectInst.HandleIntent(intent2)

		if handled2 {
			t.Error("Select should NOT handle intent with non-matching componentID")
		}

		// Verify selection didn't change
		if selectInst.SelectedIndex() != 1 {
			t.Error("Selection should remain at index 1")
		}

		t.Log("PASS: Non-matching componentID intent is ignored")
	})

	t.Run("Component without ID handles all intents", func(t *testing.T) {
		// Setup: Create Select without componentID
		selectProps := rtui.Props{
			"key":   "select-2",
			"options": []selectcomp.Option{
				{Value: "a", Label: "A"},
				{Value: "b", Label: "B"},
			},
			"selectedIndex": 0,
		}
		selectInst := selectcomp.NewInstance(selectProps)

		// All intents should be handled regardless of componentID
		intent1 := selectcomp.SelectByIndexWithID("field-1", 1)
		handled1 := selectInst.HandleIntent(intent1)

		intent2 := selectcomp.SelectByIndexWithID("field-2", 0)
		handled2 := selectInst.HandleIntent(intent2)

		if !handled1 || !handled2 {
			t.Error("Select without componentID should handle all intents")
		}

		t.Log("PASS: Component without ID handles all intents")
	})
}

// =============================================================================
// Mock Components for Testing
// =============================================================================

// mockComponent is a test component that tracks HandleIntent calls
type mockComponent struct {
	rtui.BaseComponentInstance
	key       string
	parent    interface{}
	handleFunc func(intent.Intent) bool
}

var (
	_ intent.TreeComponent    = (*mockComponent)(nil)
	_ intent.IntentHandler   = (*mockComponent)(nil)
	_ rtui.TreeContainer     = (*mockComponent)(nil)
	_ rtui.ComponentInstance = (*mockComponent)(nil)
)

func createMockComponent(key string) *mockComponent {
	return &mockComponent{
		key: key,
	}
}

func (m *mockComponent) Key() string {
	return m.key
}

func (m *mockComponent) SetKey(key string) {
	m.key = key
}

// Parent implements TreeComponent
func (m *mockComponent) Parent() interface{} {
	return m.parent
}

// Children implements TreeNode
func (m *mockComponent) Children() []rtui.ComponentInstance {
	return []rtui.ComponentInstance{} // No children
}

// HandleIntent implements IntentHandler
func (m *mockComponent) HandleIntent(i intent.Intent) bool {
	if m.handleFunc != nil {
		return m.handleFunc(i)
	}
	return false
}

// AddChild implements TreeContainer
func (m *mockComponent) AddChild(child rtui.ComponentInstance) {
	if childBase, ok := child.(*mockComponent); ok {
		childBase.parent = m
	}
}

// RemoveChild implements TreeContainer
func (m *mockComponent) RemoveChild(child rtui.ComponentInstance) {
	// Not implemented for mock
}

// Optional interfaces
func (m *mockComponent) Init(props rtui.Props)               {}
func (m *mockComponent) Destroy()                            {}
func (m *mockComponent) OnMount()                            {}
func (m *mockComponent) OnUnmount()                          {}
func (m *mockComponent) SetProps(props rtui.Props) bool     { return false }
func (m *mockComponent) GetProps() rtui.Props               { return rtui.Props{} }
func (m *mockComponent) MarkDirty()                         {}
func (m *mockComponent) IsDirty() bool                      { return false }
func (m *mockComponent) GetContext() *rtui.ComponentContext { return nil }
func (m *mockComponent) ClearChildren()                     {}

// =============================================================================
// Test Helpers
// =============================================================================

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
