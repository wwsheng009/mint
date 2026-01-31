package ui

import (
	"testing"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
)

// =============================================================================
// Universal Test Harness for Interactive Components
// =============================================================================

// ComponentTest wraps a component for automated testing
type ComponentTest struct {
	t          *testing.T
	root       *declarativeRoot
	rendered   bool
	eventQueue []frameworkevent.Event
}

// NewComponentTest creates a new test harness for a component
func NewComponentTest(t *testing.T, componentFunc ComponentFunc) *ComponentTest {
	root := newDeclarativeRoot(componentFunc, nil).(*declarativeRoot)
	return &ComponentTest{
		t:        t,
		root:     root,
		rendered: false,
	}
}

// Render renders the component
func (ct *ComponentTest) Render() *ComponentTest {
	ct.root.ctx.resetContext()
	setCurrentContext(ct.root.ctx)
	ct.root.appFn()
	setCurrentContext(nil)
	ct.rendered = true

	// Collect interactive elements
	ct.collectElements()
	return ct
}

// collectElements collects all interactive elements from the rendered tree
func (ct *ComponentTest) collectElements() {
	// Clear previous collections
	ct.root.buttons = ct.root.buttons[:0]
	ct.root.inputs = ct.root.inputs[:0]
	ct.root.textareas = ct.root.textareas[:0]
	ct.root.checkboxes = ct.root.checkboxes[:0]
	ct.root.selects = ct.root.selects[:0]

	// Collect from the rendered vnode (with context set)
	ct.root.ctx.resetContext()
	setCurrentContext(ct.root.ctx)
	vnode := ct.root.appFn()
	ct.root.resetInteractiveElements()
	ct.root.collectInteractiveElements(vnode)
	setCurrentContext(nil)
}

// =============================================================================
// Interaction Simulation
// =============================================================================

// ClickButton simulates clicking a button by index
func (ct *ComponentTest) ClickButton(index int) *ComponentTest {
	if index < 0 || index >= len(ct.root.buttons) {
		ct.t.Errorf("Button index %d out of range (have %d buttons)", index, len(ct.root.buttons))
		return ct
	}

	button := ct.root.buttons[index]
	if onClick := button.OnClick(); onClick != nil {
		onClick()
		// Re-render after click
		ct.collectElements()
	}
	return ct
}

// ClickButtonByLabel simulates clicking the first button with the given label
func (ct *ComponentTest) ClickButtonByLabel(label string) *ComponentTest {
	for i, btn := range ct.root.buttons {
		if btn.Label() == label {
			return ct.ClickButton(i)
		}
	}
	ct.t.Errorf("No button found with label '%s'", label)
	return ct
}

// PressKey simulates a key press
func (ct *ComponentTest) PressKey(key frameworkevent.Key) *ComponentTest {
	// Create a real KeyEvent
	keyEv := &frameworkevent.KeyEvent{
		BaseEvent: frameworkevent.NewBaseEvent(frameworkevent.EventKeyPress),
		Key:       key,
	}
	ct.root.HandleEvent(keyEv)
	// Re-render after event
	ct.collectElements()
	return ct
}

// PressEnter simulates pressing Enter key
func (ct *ComponentTest) PressEnter() *ComponentTest {
	// Create a KeyEvent with Enter key
	keyEv := &frameworkevent.KeyEvent{
		BaseEvent: frameworkevent.NewBaseEvent(frameworkevent.EventKeyPress),
		Special:   frameworkevent.KeyEnter,
	}
	ct.root.HandleEvent(keyEv)
	ct.collectElements()
	return ct
}

// PressTab simulates pressing Tab key
func (ct *ComponentTest) PressTab() *ComponentTest {
	// Create a KeyEvent with Tab key
	keyEv := &frameworkevent.KeyEvent{
		BaseEvent: frameworkevent.NewBaseEvent(frameworkevent.EventKeyPress),
		Special:   frameworkevent.KeyTab,
	}
	ct.root.HandleEvent(keyEv)
	ct.collectElements()
	return ct
}

// TypeText simulates typing text into the first input field
func (ct *ComponentTest) TypeText(text string) *ComponentTest {
	if len(ct.root.inputs) == 0 {
		ct.t.Error("No input field found")
		return ct
	}

	input := ct.root.inputs[0]
	if onChange := input.OnChange(); onChange != nil {
		onChange(text)
		ct.collectElements()
	}
	return ct
}

// ToggleCheckbox toggles a checkbox by index
func (ct *ComponentTest) ToggleCheckbox(index int) *ComponentTest {
	if index < 0 || index >= len(ct.root.checkboxes) {
		ct.t.Errorf("Checkbox index %d out of range (have %d checkboxes)", index, len(ct.root.checkboxes))
		return ct
	}

	checkbox := ct.root.checkboxes[index]
	newState := checkbox.Toggle()
	if onChange := checkbox.OnChange(); onChange != nil {
		onChange(newState)
	}
	ct.collectElements()
	return ct
}

// SelectOption selects an option by index in the first select
func (ct *ComponentTest) SelectOption(index int) *ComponentTest {
	if len(ct.root.selects) == 0 {
		ct.t.Error("No select found")
		return ct
	}

	selectEl := ct.root.selects[0]
	if index < 0 || index >= len(selectEl.Options()) {
		ct.t.Errorf("Option index %d out of range (have %d options)", index, len(selectEl.Options()))
		return ct
	}

	selectEl.SetSelected(index)
	if onChange := selectEl.OnChange(); onChange != nil {
		onChange(selectEl.Options()[index].Value)
	}
	ct.collectElements()
	return ct
}

// =============================================================================
// Assertion Methods
// =============================================================================

// AssertButtonCount verifies the number of buttons
func (ct *ComponentTest) AssertButtonCount(expected int) *ComponentTest {
	actual := len(ct.root.buttons)
	if actual != expected {
		ct.t.Errorf("Expected %d buttons, got %d", expected, actual)
	}
	return ct
}

// AssertButtonLabel verifies a button's label
func (ct *ComponentTest) AssertButtonLabel(index int, expectedLabel string) *ComponentTest {
	if index < 0 || index >= len(ct.root.buttons) {
		ct.t.Errorf("Button index %d out of range (have %d buttons)", index, len(ct.root.buttons))
		return ct
	}
	actual := ct.root.buttons[index].Label()
	if actual != expectedLabel {
		ct.t.Errorf("Expected button label '%s', got '%s'", expectedLabel, actual)
	}
	return ct
}

// AssertInputCount verifies the number of inputs
func (ct *ComponentTest) AssertInputCount(expected int) *ComponentTest {
	actual := len(ct.root.inputs)
	if actual != expected {
		ct.t.Errorf("Expected %d inputs, got %d", expected, actual)
	}
	return ct
}

// AssertCheckboxCount verifies the number of checkboxes
func (ct *ComponentTest) AssertCheckboxCount(expected int) *ComponentTest {
	actual := len(ct.root.checkboxes)
	if actual != expected {
		ct.t.Errorf("Expected %d checkboxes, got %d", expected, actual)
	}
	return ct
}

// AssertInputValue verifies the value of the first input
func (ct *ComponentTest) AssertInputValue(expected string) *ComponentTest {
	if len(ct.root.inputs) == 0 {
		ct.t.Error("No input field found")
		return ct
	}
	actual := ct.root.inputs[0].Value()
	if actual != expected {
		ct.t.Errorf("Expected input value '%s', got '%s'", expected, actual)
	}
	return ct
}

// AssertCheckboxChecked verifies a checkbox's checked state
func (ct *ComponentTest) AssertCheckboxChecked(index int, expected bool) *ComponentTest {
	if index < 0 || index >= len(ct.root.checkboxes) {
		ct.t.Errorf("Checkbox index %d out of range", index)
		return ct
	}
	actual := ct.root.checkboxes[index].Checked()
	if actual != expected {
		ct.t.Errorf("Expected checkbox checked=%v, got %v", expected, actual)
	}
	return ct
}

// GetButtonCount returns the current button count
func (ct *ComponentTest) GetButtonCount() int {
	return len(ct.root.buttons)
}

// GetInputCount returns the current input count
func (ct *ComponentTest) GetInputCount() int {
	return len(ct.root.inputs)
}

// GetCheckboxCount returns the current checkbox count
func (ct *ComponentTest) GetCheckboxCount() int {
	return len(ct.root.checkboxes)
}

// =============================================================================
// Component Tests Using the Harness
// =============================================================================

// TestCounterComponent tests the counter example using the harness
func TestCounterComponent(t *testing.T) {
	// Define the counter component
	Counter := func() VNode {
		_, setCount, _ := UseStateInt(0)

		return VStack(
			NewTextBuilder("Counter Demo").
				FgColor("cyan").
				Bold(true).
				Build(),
			Text(""),
			HStack(
				ButtonBuilder("  -  ").
					OnClick(func() {
						setCount(func(c int) int { return c - 1 })
					}).
					Build(),
				Text("   "),
				ButtonBuilder("  +  ").
					OnClick(func() {
						setCount(func(c int) int { return c + 1 })
					}).
					Build(),
			),
		)
	}

	// Test initial state
	NewComponentTest(t, Counter).
		Render().
		AssertButtonCount(2)
}

// TestButtonInteraction tests button click interactions
func TestButtonInteraction(t *testing.T) {
	clickCount := 0

	component := func() VNode {
		return VStack(
			ButtonBuilder("Click Me").
				OnClick(func() {
					clickCount++
				}).
				Build(),
		)
	}

	test := NewComponentTest(t, component).
		Render()

	// Initial state
	if clickCount != 0 {
		t.Errorf("Expected clickCount=0, got %d", clickCount)
	}

	// Click the button
	test.ClickButton(0)

	// After click, handler should have been called
	if clickCount != 1 {
		t.Errorf("Expected clickCount=1 after click, got %d", clickCount)
	}

	// Click again
	test.ClickButton(0)

	if clickCount != 2 {
		t.Errorf("Expected clickCount=2 after second click, got %d", clickCount)
	}
}

// TestButtonInteractionByLabel tests clicking by label
func TestButtonInteractionByLabel(t *testing.T) {
	clicked := ""

	component := func() VNode {
		return VStack(
			ButtonBuilder("Save").
				OnClick(func() {
					clicked = "save"
				}).
				Build(),
			ButtonBuilder("Cancel").
				OnClick(func() {
					clicked = "cancel"
				}).
				Build(),
		)
	}

	test := NewComponentTest(t, component).
		Render()

	// Click Save button
	test.ClickButtonByLabel("Save")

	if clicked != "save" {
		t.Errorf("Expected clicked='save', got '%s'", clicked)
	}

	// Click Cancel button
	test.ClickButtonByLabel("Cancel")

	if clicked != "cancel" {
		t.Errorf("Expected clicked='cancel', got '%s'", clicked)
	}
}

// TestStateChangeWithInteraction tests state changes triggered by button clicks
func TestStateChangeWithInteraction(t *testing.T) {
	var displayText string

	component := func() VNode {
		text, setText := UseStateString("Hello")

		// Capture current text for testing
		displayText = text

		return VStack(
			NewTextBuilder(text).Build(),
			ButtonBuilder("Change").
				OnClick(func() {
					setText("World")
				}).
				Build(),
		)
	}

	test := NewComponentTest(t, component).
		Render()

	// Initial state
	if displayText != "Hello" {
		t.Errorf("Expected text='Hello', got '%s'", displayText)
	}

	// Click button to change state
	test.ClickButton(0)

	// After click, re-render to see new state
	test.Render()

	if displayText != "World" {
		t.Errorf("Expected text='World' after click, got '%s'", displayText)
	}
}

// TestCheckboxInteraction tests checkbox toggle
func TestCheckboxInteraction(t *testing.T) {
	component := func() VNode {
		checked, setChecked := UseStateBool(false)

		return VStack(
			CheckboxBuilder().
				Label("Accept Terms").
				Checked(checked).
				OnChange(setChecked).
				Build(),
		)
	}

	test := NewComponentTest(t, component).
		Render()

	// Initial state: unchecked
	test.AssertCheckboxChecked(0, false)

	// Toggle checkbox
	test.ToggleCheckbox(0)

	// Should now be checked
	test.AssertCheckboxChecked(0, true)

	// Toggle again
	test.ToggleCheckbox(0)

	// Should be unchecked again
	test.AssertCheckboxChecked(0, false)
}

// TestInputInteraction tests typing into input fields
func TestInputInteraction(t *testing.T) {
	component := func() VNode {
		text, setText := UseStateString("")

		return VStack(
			InputBuilder().
				Value(text).
				Placeholder("Type here").
				OnChange(setText).
				Build(),
		)
	}

	test := NewComponentTest(t, component).
		Render()

	// Initial state
	test.AssertInputValue("")

	// Type some text
	test.TypeText("Hello World")

	// Input should have the new value
	test.AssertInputValue("Hello World")
}

// TestConditionalModal tests a modal that opens/closes based on state
func TestConditionalModal(t *testing.T) {
	var isModalOpen bool

	component := func() VNode {
		isOpen, setIsOpen, _ := UseStateInt(0)
		isModalOpen = isOpen == 1

		if isOpen == 1 {
			// Modal is open - show close button
			return VStack(
				NewTextBuilder("Modal Content").Build(),
				ButtonBuilder("Close").
					OnClick(func() {
						setIsOpen(0)
					}).
					Build(),
			)
		}

		// Modal is closed - show open button
		return VStack(
			NewTextBuilder("Main Content").Build(),
			ButtonBuilder("Open").
				OnClick(func() {
					setIsOpen(1)
				}).
				Build(),
		)
	}

	test := NewComponentTest(t, component).
		Render()

	// Initial state: modal closed, one button ("Open")
	test.AssertButtonCount(1)
	test.AssertButtonLabel(0, "Open")
	if isModalOpen {
		t.Error("Expected modal to be closed")
	}

	// Click "Open" button
	test.ClickButton(0)

	// Modal should now be open, one button ("Close")
	test.AssertButtonCount(1)
	test.AssertButtonLabel(0, "Close")
	if !isModalOpen {
		t.Error("Expected modal to be open")
	}

	// Click "Close" button
	test.ClickButton(0)

	// Modal should be closed again
	test.AssertButtonCount(1)
	test.AssertButtonLabel(0, "Open")
	if isModalOpen {
		t.Error("Expected modal to be closed")
	}
}

// TestMultiModalFlow tests a complete modal interaction flow
func TestMultiModalFlow(t *testing.T) {
	component := func() VNode {
		state, setState, _ := UseStateInt(0)

		switch state {
		case 0:
			// Initial screen
			return VStack(
				NewTextBuilder("Welcome").Build(),
				ButtonBuilder("Start").
					OnClick(func() {
						setState(1)
					}).
					Build(),
			)
		case 1:
			// Confirmation modal
			return VStack(
				NewTextBuilder("Are you sure?").Build(),
				HStack(
					ButtonBuilder("Yes").
						OnClick(func() {
							setState(2)
						}).
						Build(),
					ButtonBuilder("No").
						OnClick(func() {
							setState(0)
						}).
						Build(),
				),
			)
		case 2:
			// Success screen
			return VStack(
				NewTextBuilder("Done!").Build(),
				ButtonBuilder("Back").
					OnClick(func() {
						setState(0)
					}).
					Build(),
			)
		default:
			return Text("")
		}
	}

	test := NewComponentTest(t, component).
		Render()

	// Step 0: Welcome screen with "Start" button
	test.AssertButtonCount(1)
	test.AssertButtonLabel(0, "Start")

	// Click Start -> go to confirmation
	test.ClickButton(0)

	// Step 1: Confirmation with "Yes" and "No" buttons
	test.AssertButtonCount(2)
	test.AssertButtonLabel(0, "Yes")
	test.AssertButtonLabel(1, "No")

	// Click Yes -> go to success
	test.ClickButton(0) // Click Yes

	// Step 2: Success screen with "Back" button
	test.AssertButtonCount(1)
	test.AssertButtonLabel(0, "Back")

	// Click Back -> return to welcome
	test.ClickButton(0)

	// Back to welcome
	test.AssertButtonCount(1)
	test.AssertButtonLabel(0, "Start")
}

// TestSelectInteraction tests dropdown selection
func TestSelectInteraction(t *testing.T) {
	var selectedValue string

	component := func() VNode {
		value, setValue := UseStateString("")

		selectedValue = value

		return VStack(
			SelectBuilder().
				AddOption("opt1", "Option 1").
				AddOption("opt2", "Option 2").
				AddOption("opt3", "Option 3").
				Selected(0).
				OnChange(setValue).
				Build(),
		)
	}

	test := NewComponentTest(t, component).
		Render()

	// Initial selection (index 0 = "opt1")
	if selectedValue != "" {
		t.Logf("Initial selection: '%s'", selectedValue)
	}

	// Select option 2 (index 1)
	test.SelectOption(1)

	if selectedValue != "opt2" {
		t.Errorf("Expected selectedValue='opt2', got '%s'", selectedValue)
	}
}

// =============================================================================
// Automated Component Tests (Table-Driven)
// =============================================================================

// TestAllComponents runs basic smoke tests on all components
func TestAllComponents(t *testing.T) {
	components := map[string]ComponentFunc{
		"Button": func() VNode {
			return ButtonBuilder("Click Me").Build()
		},
		"Text": func() VNode {
			return NewTextBuilder("Hello").Build()
		},
		"VStack": func() VNode {
			return VStack(
				NewTextBuilder("A").Build(),
				NewTextBuilder("B").Build(),
			)
		},
		"HStack": func() VNode {
			return HStack(
				NewTextBuilder("A").Build(),
				NewTextBuilder("B").Build(),
			)
		},
		"Fragment": func() VNode {
			return Fragment(
				NewTextBuilder("A").Build(),
				NewTextBuilder("B").Build(),
			)
		},
		"Checkbox": func() VNode {
			return CheckboxBuilder().
				Label("Check me").
				Build()
		},
		"Progress": func() VNode {
			return ProgressBuilder().
				Value(50).
				Max(100).
				Build()
		},
		"Spinner": func() VNode {
			return SpinnerBuilder().
				Message("Loading...").
				Build()
		},
		"Input": func() VNode {
			return InputBuilder().
				Placeholder("Type here").
				Build()
		},
		"Select": func() VNode {
			return SelectBuilder().
				AddOption("a", "Option A").
				AddOption("b", "Option B").
				Build()
		},
		"Table": func() VNode {
			return TableBuilder().
				Columns([]TableColumn{
					{Title: "ID", Width: 5},
					{Title: "Name", Width: 10},
				}).
				AddRow("1", "Item").
				Build()
		},
	}

	for name, componentFunc := range components {
		t.Run(name, func(t *testing.T) {
			// Test that component renders without panic
			NewComponentTest(t, componentFunc).Render()
		})
	}
}
