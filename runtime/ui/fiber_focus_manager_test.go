package ui

import "testing"

type mockFiberFocusableInstance struct {
	focused bool
	changes int
	key     string
	props   Props
	dirty   bool
}

func (m *mockFiberFocusableInstance) SetFocus(focused bool) {
	if m.focused != focused {
		m.focused = focused
		m.changes++
	}
}

func (m *mockFiberFocusableInstance) Key() string                   { return m.key }
func (m *mockFiberFocusableInstance) SetKey(key string)             { m.key = key }
func (m *mockFiberFocusableInstance) Init(props Props)              { m.props = props }
func (m *mockFiberFocusableInstance) Destroy()                      {}
func (m *mockFiberFocusableInstance) OnMount()                      {}
func (m *mockFiberFocusableInstance) OnUnmount()                    {}
func (m *mockFiberFocusableInstance) SetProps(props Props) bool     { m.props = props; return true }
func (m *mockFiberFocusableInstance) GetProps() Props               { return m.props }
func (m *mockFiberFocusableInstance) MarkDirty()                    { m.dirty = true }
func (m *mockFiberFocusableInstance) IsDirty() bool                 { return m.dirty }
func (m *mockFiberFocusableInstance) GetContext() *ComponentContext { return nil }
func (m *mockFiberFocusableInstance) HasFocus() bool                { return m.focused }
func (m *mockFiberFocusableInstance) IsDisabled() bool              { return false }

func TestFiberFocusManager_SetFocusByIndex_NoRetoggleOnSameIndex(t *testing.T) {
	manager := NewFiberFocusManager()

	instance := &mockFiberFocusableInstance{}
	fiber := &Fiber{
		NodeID:   1,
		Instance: instance,
	}

	manager.UpdateFocusableList([]*Fiber{fiber})

	if ok := manager.SetFocusByIndex(0); !ok {
		t.Fatal("SetFocusByIndex(0) should succeed")
	}
	if instance.changes != 1 || !instance.focused {
		t.Fatalf("after first focus: changes=%d focused=%v, want changes=1 focused=true", instance.changes, instance.focused)
	}

	// Setting focus to the same index should not clear and re-apply focus.
	if ok := manager.SetFocusByIndex(0); !ok {
		t.Fatal("SetFocusByIndex(0) should still succeed")
	}
	if instance.changes != 1 || !instance.focused {
		t.Fatalf("after same-index focus: changes=%d focused=%v, want changes=1 focused=true", instance.changes, instance.focused)
	}
}

func TestFiberFocusManager_SetActiveLayer_TrapsFocusToModalSubtree(t *testing.T) {
	manager := NewFiberFocusManager()

	baseInst := &mockFiberFocusableInstance{}
	modalPrimaryInst := &mockFiberFocusableInstance{}
	modalSecondaryInst := &mockFiberFocusableInstance{}

	root := &Fiber{Key: "root", Type: VNodeComponent}
	baseButton := &Fiber{
		NodeID:   1,
		Layer:    LayerBase,
		Instance: baseInst,
	}
	modalRoot := &Fiber{
		NodeID: 2,
		Layer:  LayerModal,
	}
	modalPrimary := &Fiber{
		NodeID:   3,
		Layer:    LayerBase,
		Instance: modalPrimaryInst,
	}
	modalSecondary := &Fiber{
		NodeID:   4,
		Layer:    LayerBase,
		Instance: modalSecondaryInst,
	}

	root.Child = baseButton
	baseButton.Sibling = modalRoot
	modalRoot.Child = modalPrimary
	modalPrimary.Sibling = modalSecondary

	manager.CollectFromFiber(root)
	manager.SetActiveLayer(LayerModal)

	if got := manager.GetCurrent(); got != modalPrimary {
		t.Fatalf("focused fiber = %#v, want modal primary", got)
	}
	if !modalPrimaryInst.focused {
		t.Fatal("modal primary should receive focus when modal layer becomes active")
	}
	if baseInst.focused {
		t.Fatal("base button should not remain focused while modal layer is active")
	}

	if !manager.FocusNext() {
		t.Fatal("FocusNext should cycle within modal layer")
	}
	if got := manager.GetCurrent(); got != modalSecondary {
		t.Fatalf("focused fiber after next = %#v, want modal secondary", got)
	}

	if !manager.FocusNext() {
		t.Fatal("FocusNext should wrap within modal layer")
	}
	if got := manager.GetCurrent(); got != modalPrimary {
		t.Fatalf("focused fiber after wrap = %#v, want modal primary", got)
	}
}
