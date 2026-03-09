package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type focusLayerInstance struct {
	key     string
	props   rtui.Props
	dirty   bool
	focused bool
}

func (m *focusLayerInstance) Key() string                        { return m.key }
func (m *focusLayerInstance) SetKey(key string)                  { m.key = key }
func (m *focusLayerInstance) Init(props rtui.Props)              { m.props = props }
func (m *focusLayerInstance) Destroy()                           {}
func (m *focusLayerInstance) OnMount()                           {}
func (m *focusLayerInstance) OnUnmount()                         {}
func (m *focusLayerInstance) SetProps(props rtui.Props) bool     { m.props = props; return true }
func (m *focusLayerInstance) GetProps() rtui.Props               { return m.props }
func (m *focusLayerInstance) MarkDirty()                         { m.dirty = true }
func (m *focusLayerInstance) IsDirty() bool                      { return m.dirty }
func (m *focusLayerInstance) GetContext() *rtui.ComponentContext { return nil }
func (m *focusLayerInstance) SetFocus(focused bool)              { m.focused = focused }
func (m *focusLayerInstance) HasFocus() bool                     { return m.focused }
func (m *focusLayerInstance) IsDisabled() bool                   { return false }

func TestReconciler_UpdateFocusManagerFromFiber_SwitchesToOverlayFocus(t *testing.T) {
	reconciler := NewReconciler(nil, nil, ReconcilerConfig{EnableFiber: true})
	reconciler.focusMgr = rtui.NewFiberFocusManager()

	baseInst := &focusLayerInstance{}
	overlayInst := &focusLayerInstance{}

	root := &rtui.Fiber{NodeID: 1, Tag: "root", Layer: rtui.LayerBase}
	baseFiber := &rtui.Fiber{
		NodeID:   2,
		Tag:      "button",
		Layer:    rtui.LayerBase,
		Instance: baseInst,
		Return:   root,
	}
	overlayFiber := &rtui.Fiber{
		NodeID:   3,
		Tag:      "menu-popup",
		Layer:    rtui.LayerOverlay,
		Instance: overlayInst,
		Return:   root,
	}
	root.Child = baseFiber
	baseFiber.Sibling = overlayFiber

	reconciler.focusMgr.UpdateFocusableList([]*rtui.Fiber{baseFiber})
	if ok := reconciler.focusMgr.SetFocusByIndex(0); !ok {
		t.Fatal("SetFocusByIndex(0) should succeed")
	}

	reconciler.updateFocusManagerFromFiber(root)

	if got := reconciler.focusMgr.GetCurrent(); got != overlayFiber {
		t.Fatalf("focused fiber = %#v, want overlay fiber", got)
	}
	if !overlayInst.focused {
		t.Fatal("overlay instance should receive focus after overlay layer becomes active")
	}
}
