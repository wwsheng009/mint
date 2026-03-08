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
