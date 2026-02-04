package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// TextVNode Tests
// =============================================================================

func TestTextVNode(t *testing.T) {
	text := &TextVNode{
		ElementVNode: &ElementVNode{tag: "text"},
		content:     "hello",
	}

	if text.Content() != "hello" {
		t.Errorf("Content() = %s, want 'hello'", text.Content())
	}

	if text.Tag() != "text" {
		t.Errorf("Tag() = %s, want 'text'", text.Tag())
	}

	// Nil receiver
	var nilText *TextVNode
	if nilText.Content() != "" {
		t.Error("nil TextVNode Content() should return empty string")
	}
	if nilText.Tag() != "text" {
		t.Error("nil TextVNode Tag() should return 'text'")
	}
}

// =============================================================================
// ButtonVNode Tests
// =============================================================================

func TestButtonVNode(t *testing.T) {
	onClick := func() {}
	button := &ButtonVNode{
		ElementVNode: &ElementVNode{tag: "button"},
		label:       "Click Me",
		onClick:     onClick,
		disabled:    false,
	}

	if button.Label() != "Click Me" {
		t.Errorf("Label() = %s, want 'Click Me'", button.Label())
	}

	if button.OnClick() == nil {
		t.Error("OnClick() should not return nil")
	}

	if button.Disabled() {
		t.Error("Disabled() should return false")
	}

	if button.ContainsPoint(0, 0) {
		t.Error("ContainsPoint should return false")
	}

	// Nil receiver
	var nilButton *ButtonVNode
	if nilButton.Label() != "" {
		t.Error("nil ButtonVNode Label() should return empty string")
	}
	if nilButton.OnClick() != nil {
		t.Error("nil ButtonVNode OnClick() should return nil")
	}
	if nilButton.Disabled() {
		t.Error("nil ButtonVNode Disabled() should return false")
	}
}

// =============================================================================
// InputVNode Tests
// =============================================================================

func TestInputVNode(t *testing.T) {
	input := &InputVNode{
		ElementVNode: &ElementVNode{tag: "input"},
		value:       "test value",
		placeholder: "enter text",
	}

	if input.Value() != "test value" {
		t.Errorf("Value() = %s, want 'test value'", input.Value())
	}

	if input.Placeholder() != "enter text" {
		t.Errorf("Placeholder() = %s, want 'enter text'", input.Placeholder())
	}

	// Nil receiver
	var nilInput *InputVNode
	if nilInput.Value() != "" {
		t.Error("nil InputVNode Value() should return empty string")
	}
	if nilInput.Placeholder() != "" {
		t.Error("nil InputVNode Placeholder() should return empty string")
	}
}

// =============================================================================
// TextareaVNode Tests
// =============================================================================

func TestTextareaVNode(t *testing.T) {
	textarea := &TextareaVNode{
		ElementVNode: &ElementVNode{tag: "textarea"},
		value:       "long text",
		placeholder: "enter long text",
		rows:        5,
	}

	if textarea.Value() != "long text" {
		t.Errorf("Value() = %s, want 'long text'", textarea.Value())
	}

	if textarea.Placeholder() != "enter long text" {
		t.Errorf("Placeholder() = %s, want 'enter long text'", textarea.Placeholder())
	}

	// Nil receiver
	var nilTextarea *TextareaVNode
	if nilTextarea.Value() != "" {
		t.Error("nil TextareaVNode Value() should return empty string")
	}
	if nilTextarea.Placeholder() != "" {
		t.Error("nil TextareaVNode Placeholder() should return empty string")
	}
}

// =============================================================================
// CheckboxVNode Tests
// =============================================================================

func TestCheckboxVNode(t *testing.T) {
	checkbox := &CheckboxVNode{
		ElementVNode: &ElementVNode{tag: "checkbox"},
		label:       "Agree",
		checked:     true,
	}

	if checkbox.Label() != "Agree" {
		t.Errorf("Label() = %s, want 'Agree'", checkbox.Label())
	}

	if !checkbox.Checked() {
		t.Error("Checked() should return true")
	}

	// Nil receiver
	var nilCheckbox *CheckboxVNode
	if nilCheckbox.Label() != "" {
		t.Error("nil CheckboxVNode Label() should return empty string")
	}
	if nilCheckbox.Checked() {
		t.Error("nil CheckboxVNode Checked() should return false")
	}
}

// =============================================================================
// SelectVNode Tests
// =============================================================================

func TestSelectVNode(t *testing.T) {
	options := []SelectOption{
		{Value: "1", Label: "Option 1"},
		{Value: "2", Label: "Option 2"},
	}
	sel := &SelectVNode{
		ElementVNode: &ElementVNode{tag: "select"},
		selected:    1,
		options:     options,
	}

	if sel.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1", sel.Selected())
	}

	selOptions := sel.Options()
	if len(selOptions) != 2 {
		t.Errorf("Options() length = %d, want 2", len(selOptions))
	}

	// Nil receiver
	var nilSelect *SelectVNode
	if nilSelect.Selected() != 0 {
		t.Error("nil SelectVNode Selected() should return 0")
	}
	if nilSelect.Options() != nil {
		t.Error("nil SelectVNode Options() should return nil")
	}
}

// =============================================================================
// ProgressVNode Tests
// =============================================================================

func TestProgressVNode(t *testing.T) {
	progress := &ProgressVNode{
		ElementVNode: &ElementVNode{tag: "progress"},
		value:       50,
		max:         100,
		percent:     50,
		width:       200,
	}

	if progress.Value() != 50 {
		t.Errorf("Value() = %d, want 50", progress.Value())
	}

	if progress.Max() != 100 {
		t.Errorf("Max() = %d, want 100", progress.Max())
	}

	if progress.Percent() != 50 {
		t.Errorf("Percent() = %d, want 50", progress.Percent())
	}

	if progress.Width() != 200 {
		t.Errorf("Width() = %d, want 200", progress.Width())
	}

	// Nil receiver
	var nilProgress *ProgressVNode
	if nilProgress.Value() != 0 {
		t.Error("nil ProgressVNode Value() should return 0")
	}
	if nilProgress.Max() != 0 {
		t.Error("nil ProgressVNode Max() should return 0")
	}
	if nilProgress.Percent() != 0 {
		t.Error("nil ProgressVNode Percent() should return 0")
	}
	if nilProgress.Width() != 0 {
		t.Error("nil ProgressVNode Width() should return 0")
	}
}

// =============================================================================
// TableVNode Tests
// =============================================================================

func TestTableVNode(t *testing.T) {
	table := &TableVNode{
		ElementVNode: &ElementVNode{tag: "table"},
		columns:     []string{"A", "B", "C"},
		rows:        [][]string{
			{"1", "2", "3"},
			{"4", "5", "6"},
		},
	}

	cols := table.Columns()
	if len(cols) != 3 {
		t.Errorf("Columns() length = %d, want 3", len(cols))
	}

	rows := table.Rows()
	if len(rows) != 2 {
		t.Errorf("Rows() length = %d, want 2", len(rows))
	}

	// Nil receiver
	var nilTable *TableVNode
	if nilTable.Columns() != nil {
		t.Error("nil TableVNode Columns() should return nil")
	}
	if nilTable.Rows() != nil {
		t.Error("nil TableVNode Rows() should return nil")
	}
}

// =============================================================================
// VirtualListVNode Tests
// =============================================================================

func TestVirtualListVNode(t *testing.T) {
	vlist := &VirtualListVNode{
		ElementVNode: &ElementVNode{tag: "virtual-list"},
		listHeight:  400,
		itemCount:  100,
		itemHeight: 40,
	}

	if vlist.ListHeight() != 400 {
		t.Errorf("ListHeight() = %d, want 400", vlist.ListHeight())
	}

	if vlist.ItemCount() != 100 {
		t.Errorf("ItemCount() = %d, want 100", vlist.ItemCount())
	}

	if vlist.ItemHeight() != 40 {
		t.Errorf("ItemHeight() = %d, want 40", vlist.ItemHeight())
	}

	// Nil receiver
	var nilVList *VirtualListVNode
	if nilVList.ListHeight() != 0 {
		t.Error("nil VirtualListVNode ListHeight() should return 0")
	}
	if nilVList.ItemCount() != 0 {
		t.Error("nil VirtualListVNode ItemCount() should return 0")
	}
	if nilVList.ItemHeight() != 0 {
		t.Error("nil VirtualListVNode ItemHeight() should return 0")
	}
}

// =============================================================================
// TabsVNode Tests
// =============================================================================

func TestTabsVNode(t *testing.T) {
	tabs := &TabsVNode{
		ElementVNode: &ElementVNode{tag: "tabs"},
		activeTab:   "tab2",
	}

	if tabs.ActiveTab() != "tab2" {
		t.Errorf("ActiveTab() = %s, want 'tab2'", tabs.ActiveTab())
	}

	// Nil receiver
	var nilTabs *TabsVNode
	if nilTabs.ActiveTab() != "" {
		t.Error("nil TabsVNode ActiveTab() should return empty string")
	}
}

// =============================================================================
// ModalVNode Tests
// =============================================================================

func TestModalVNode(t *testing.T) {
	content := Element("div").Build()
	footer := Element("div").Build()
	modal := &ModalVNode{
		ElementVNode: &ElementVNode{tag: "modal"},
		width:       400,
		height:      300,
		title:       "Dialog",
		isOpen:      true,
		content:     content,
		footer:      footer,
	}

	if modal.Width() != 400 {
		t.Errorf("Width() = %d, want 400", modal.Width())
	}

	if modal.Height() != 300 {
		t.Errorf("Height() = %d, want 300", modal.Height())
	}

	if modal.Title() != "Dialog" {
		t.Errorf("Title() = %s, want 'Dialog'", modal.Title())
	}

	if !modal.IsOpen() {
		t.Error("IsOpen() should return true")
	}

	if modal.Content() != content {
		t.Error("Content() should return the content VNode")
	}

	if modal.Footer() != footer {
		t.Error("Footer() should return the footer VNode")
	}

	// Nil receiver
	var nilModal *ModalVNode
	if nilModal.Width() != 0 {
		t.Error("nil ModalVNode Width() should return 0")
	}
	if nilModal.Height() != 0 {
		t.Error("nil ModalVNode Height() should return 0")
	}
	if nilModal.Title() != "" {
		t.Error("nil ModalVNode Title() should return empty string")
	}
	if nilModal.IsOpen() {
		t.Error("nil ModalVNode IsOpen() should return false")
	}
}

// =============================================================================
// ElementVNode Props/Key/Style/Children Tests
// =============================================================================

func TestElementVNodeProps(t *testing.T) {
	elem := &ElementVNode{
		tag:   "div",
		props: Props{"class": "container"},
	}

	props := elem.Props()
	if props.GetString("class") != "container" {
		t.Errorf("Props() should return element props")
	}

	// Empty props - nil or empty map are both acceptable
	elem2 := &ElementVNode{tag: "div"}
	props2 := elem2.Props()
	if props2 != nil && len(props2) != 0 {
		t.Error("Props() should return nil or empty props map for new element")
	}
}

func TestElementVNodeKey(t *testing.T) {
	elem := &ElementVNode{tag: "div", key: "test-key"}

	if elem.Key() != "test-key" {
		t.Errorf("Key() = %s, want 'test-key'", elem.Key())
	}

	// Empty key
	elem2 := &ElementVNode{tag: "div"}
	if elem2.Key() != "" {
		t.Errorf("Key() should return empty string for element without key")
	}
}

func TestElementVNodeStyle(t *testing.T) {
	s := style.Style{}.Foreground("red").Background("blue")
	elem := &ElementVNode{tag: "div", style: s}

	style := elem.Style()
	if style.FG != "red" {
		t.Errorf("Style FG = %s, want red", style.FG)
	}
	if style.BG != "blue" {
		t.Errorf("Style BG = %s, want blue", style.BG)
	}

	// Empty style
	elem2 := &ElementVNode{tag: "div"}
	style2 := elem2.Style()
	if style2.FG != "" || style2.BG != "" {
		t.Error("Style() should return empty style for element without style")
	}
}

func TestElementVNodeChildren(t *testing.T) {
	child1 := Element("span").Build()
	child2 := Element("span").Build()
	elem := &ElementVNode{
		tag:      "div",
		children: []VNode{child1, child2},
	}

	children := elem.Children()
	if len(children) != 2 {
		t.Errorf("Children() length = %d, want 2", len(children))
	}

	// Nil children
	elem2 := &ElementVNode{tag: "div"}
	children2 := elem2.Children()
	if children2 != nil {
		t.Error("Children() should return nil for element without children")
	}
}

// =============================================================================
// ComponentVNode Name/Props/SetChildren/SetProps Tests
// =============================================================================

func TestComponentVNodeName(t *testing.T) {
	comp := &ComponentVNode{name: "TestComp"}

	if comp.Name() != "TestComp" {
		t.Errorf("Name() = %s, want 'TestComp'", comp.Name())
	}
}

func TestComponentVNodeProps(t *testing.T) {
	comp := &ComponentVNode{
		name:  "TestComp",
		props: Props{"prop1": "value1"},
	}

	props := comp.Props()
	if props.GetString("prop1") != "value1" {
		t.Errorf("Props() should return component props")
	}
}

func TestComponentVNodeSetChildren(t *testing.T) {
	comp := &ComponentVNode{name: "TestComp"}

	// SetChildren on ComponentVNode is a no-op (components don't have static children)
	// This should not panic
	comp.SetChildren([]VNode{Element("div").Build()})
}

func TestComponentVNodeSetProps(t *testing.T) {
	comp := &ComponentVNode{name: "TestComp"}

	newProps := Props{"new": "value"}
	comp.SetProps(newProps)

	if comp.Props().GetString("new") != "value" {
		t.Error("SetProps should set new props")
	}
}

// =============================================================================
// ErrorBoundaryVNode SetProps Tests
// =============================================================================

func TestErrorBoundaryVNodeSetProps(t *testing.T) {
	eb := &ErrorBoundaryVNode{
		name: "TestBoundary",
	}

	newProps := Props{"fallback": Element("div").Build()}
	eb.SetProps(newProps)

	// Just verify it doesn't panic
}
