package intent_test

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/form"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
)

// =============================================================================
// Nested Intent Bubble Integration Tests (Phase 10 Extension)
// =============================================================================

// TestNestedIntentBubble_FormTo tests intent bubbling from Select to Form.
func TestNestedIntentBubble_FormTo(t *testing.T) {
	t.Run("Select intent bubbles to Form parent", func(t *testing.T) {
		// Create parent Form component
		formProps := rtui.Props{
			"key":  "test-form-1",
			"name": "testForm",
		}
		formInst := form.NewInstance(formProps)

		// Create child Select component with formID
		selectProps := rtui.Props{
			"key":      "test-select-1",
			"formID":   "test-form",
			"disabled": false,
			"options": []selectcomp.Option{
				{Value: "1", Label: "Option 1"},
				{Value: "2", Label: "Option 2"},
			},
			"selectedIndex": 0,
		}
		selectInst := selectcomp.NewInstance(selectProps)

		// Add select to form
		formInst.AddChild(selectInst)

		// Verify form has select as child
		if children := formInst.Children(); len(children) != 1 || children[0] != selectInst {
			t.Error("Form should have Select as a child")
		}

		// Set intent emitter on select
		intentEmitted := false
		var emittedIntent intent.Intent
		selectInst.SetIntentEmitter(func(i intent.Intent) {
			intentEmitted = true
			emittedIntent = i
		})

		// Trigger selection change
		selectInst.SelectNext()

		// Verify intent was emitted
		if !intentEmitted {
			t.Error("Intent should be emitted on selection change")
		}

		// Verify intent type
		if _, ok := emittedIntent.(selectcomp.SelectChangeIntent); !ok {
			t.Error("Emitted intent should be SelectChangeIntent")
		}

		// Verify actual selection changed
		if selectInst.SelectedIndex() != 1 {
			t.Errorf("Expected selectedIndex 1, got %d", selectInst.SelectedIndex())
		}
	})

	t.Run("Select handles intents via componentID routing", func(t *testing.T) {
		// Create parent Form component
		formProps := rtui.Props{
			"key":  "test-form-2",
			"name": "testForm2",
		}
		formInst := form.NewInstance(formProps)

		// Create child Select component with componentID
		selectProps := rtui.Props{
			"key":         "test-select-2",
			"componentID": "select-field",
			"options": []selectcomp.Option{
				{Value: "a", Label: "Alpha"},
				{Value: "b", Label: "Beta"},
				{Value: "c", Label: "Gamma"},
			},
			"selectedIndex": 0,
		}
		selectInst := selectcomp.NewInstance(selectProps)

		// Add select to form
		formInst.AddChild(selectInst)

		// Set intent emitter on select for bubbling
		bubbleIntents := make([]intent.Intent, 0)
		selectInst.SetIntentEmitter(func(i intent.Intent) {
			bubbleIntents = append(bubbleIntents, i)
		})

		// Handle intents with componentID routing
		intent1 := selectcomp.SelectByIndexWithID("select-field", 2)
		handled := selectInst.HandleIntent(intent1)

		if !handled {
			t.Error("Select should handle SelectByIndexIntent with matching componentID")
		}

		if selectInst.SelectedIndex() != 2 {
			t.Errorf("Expected selectedIndex 2, got %d", selectInst.SelectedIndex())
		}

		// Test with non-matching componentID
		intent2 := selectcomp.SelectByIndexWithID("other-select", 1)
		handled2 := selectInst.HandleIntent(intent2)

		if handled2 {
			t.Error("Select should not handle SelectByIndexIntent with non-matching componentID")
		}

		// Selection should not have changed
		if selectInst.SelectedIndex() != 2 {
			t.Errorf("Expected selectedIndex to remain 2, got %d", selectInst.SelectedIndex())
		}
	})
}

// TestMultiChildIntentBubbling tests intent bubbling with multiple children in one parent.
func TestMultiChildIntentBubbling(t *testing.T) {
	t.Run("Multiple Selects in one Form", func(t *testing.T) {
		// Parent Form
		formProps := rtui.Props{
			"key":  "test-form-3",
			"name": "multiSelectForm",
		}
		formInst := form.NewInstance(formProps)

		// Child Select 1
		select1Props := rtui.Props{
			"key":         "select-1",
			"componentID": "field1",
			"options": []selectcomp.Option{
				{Value: "a", Label: "A"},
				{Value: "b", Label: "B"},
			},
			"selectedIndex": 0,
		}
		selectInst1 := selectcomp.NewInstance(select1Props)

		// Child Select 2
		select2Props := rtui.Props{
			"key":         "select-2",
			"componentID": "field2",
			"options": []selectcomp.Option{
				{Value: "1", Label: "One"},
				{Value: "2", Label: "Two"},
			},
			"selectedIndex": 0,
		}
		selectInst2 := selectcomp.NewInstance(select2Props)

		formInst.AddChild(selectInst1)
		formInst.AddChild(selectInst2)

		// Track intents per child
		var select1Intents, select2Intents []intent.Intent

		selectInst1.SetIntentEmitter(func(i intent.Intent) {
			select1Intents = append(select1Intents, i)
		})
		selectInst2.SetIntentEmitter(func(i intent.Intent) {
			select2Intents = append(select2Intents, i)
		})

		// Trigger selection change on select1
		selectInst1.SelectNext()

		// Only select1 should emit, not select2
		if len(select1Intents) != 1 {
			t.Error("Select1 should emit intent")
		}
		if len(select2Intents) != 0 {
			t.Error("Select2 should not emit intent")
		}

		// Trigger selection change on select2
		selectInst2.SelectNext()

		// Now select2 should also emit
		if len(select2Intents) != 1 {
			t.Error("Select2 should emit intent")
		}

		// Verify both selections changed
		if selectInst1.SelectedIndex() != 1 {
			t.Errorf("Expected selectInst1 selectedIndex 1, got %d", selectInst1.SelectedIndex())
		}
		if selectInst2.SelectedIndex() != 1 {
			t.Errorf("Expected selectInst2 selectedIndex 1, got %d", selectInst2.SelectedIndex())
		}
	})
}
