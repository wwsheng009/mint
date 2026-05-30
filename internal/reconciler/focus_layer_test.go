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
	reconciler := NewReconciler(nil, nil, ReconcilerConfig{})
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

func TestReconciler_UpdateFocusManagerFromFiber_TrapsFocusInModalSubtree(t *testing.T) {
	reconciler := NewReconciler(nil, nil, ReconcilerConfig{})
	reconciler.focusMgr = rtui.NewFiberFocusManager()

	baseInst := &focusLayerInstance{}
	modalPrimaryInst := &focusLayerInstance{}
	modalSecondaryInst := &focusLayerInstance{}

	root := &rtui.Fiber{NodeID: 1, Tag: "root", Layer: rtui.LayerBase}
	baseFiber := &rtui.Fiber{
		NodeID:   2,
		Tag:      "button",
		Layer:    rtui.LayerBase,
		Instance: baseInst,
		Return:   root,
	}
	modalFiber := &rtui.Fiber{
		NodeID: 3,
		Tag:    "modal",
		Layer:  rtui.LayerModal,
		Return: root,
	}
	modalPrimaryFiber := &rtui.Fiber{
		NodeID:   4,
		Tag:      "button",
		Layer:    rtui.LayerBase,
		Instance: modalPrimaryInst,
		Return:   modalFiber,
	}
	modalSecondaryFiber := &rtui.Fiber{
		NodeID:   5,
		Tag:      "button",
		Layer:    rtui.LayerBase,
		Instance: modalSecondaryInst,
		Return:   modalFiber,
	}

	root.Child = baseFiber
	baseFiber.Sibling = modalFiber
	modalFiber.Child = modalPrimaryFiber
	modalPrimaryFiber.Sibling = modalSecondaryFiber

	reconciler.focusMgr.UpdateFocusableList([]*rtui.Fiber{baseFiber})
	if ok := reconciler.focusMgr.SetFocusByIndex(0); !ok {
		t.Fatal("SetFocusByIndex(0) should succeed")
	}

	reconciler.updateFocusManagerFromFiber(root)

	if got := reconciler.focusMgr.GetActiveLayer(); got != rtui.LayerModal {
		t.Fatalf("active layer = %v, want modal", got)
	}
	if got := reconciler.focusMgr.GetCurrent(); got != modalPrimaryFiber {
		t.Fatalf("focused fiber = %#v, want modal primary fiber", got)
	}
	if !modalPrimaryInst.focused {
		t.Fatal("modal primary instance should receive focus when modal layer is active")
	}
	if baseInst.focused {
		t.Fatal("base instance should not remain focused while modal is active")
	}

	if !reconciler.focusMgr.FocusNext() {
		t.Fatal("FocusNext should stay within modal subtree")
	}
	if got := reconciler.focusMgr.GetCurrent(); got != modalSecondaryFiber {
		t.Fatalf("focused fiber after next = %#v, want modal secondary fiber", got)
	}
}

func TestReconciler_UpdateFocusManagerFromFiber_IgnoresEmptyPortalHosts(t *testing.T) {
	reconciler := NewReconciler(nil, nil, ReconcilerConfig{})
	reconciler.focusMgr = rtui.NewFiberFocusManager()

	baseInst := &focusLayerInstance{}

	root := &rtui.Fiber{NodeID: 1, Tag: "root", Layer: rtui.LayerBase}
	overlayHost := &rtui.Fiber{
		NodeID: 2,
		Tag:    "box",
		Layer:  rtui.LayerOverlay,
		Props: rtui.Props{
			"portalRootId": rtui.DefaultOverlayPortalRootID,
		},
		Return: root,
	}
	modalHost := &rtui.Fiber{
		NodeID: 3,
		Tag:    "box",
		Layer:  rtui.LayerModal,
		Props: rtui.Props{
			"portalRootId": rtui.DefaultModalPortalRootID,
		},
		Return: root,
	}
	baseFiber := &rtui.Fiber{
		NodeID:   4,
		Tag:      "button",
		Layer:    rtui.LayerBase,
		Instance: baseInst,
		Return:   root,
	}

	root.Child = overlayHost
	overlayHost.Sibling = modalHost
	modalHost.Sibling = baseFiber

	reconciler.updateFocusManagerFromFiber(root)

	if got := reconciler.focusMgr.GetActiveLayer(); got != rtui.LayerBase {
		t.Fatalf("active layer = %v, want base", got)
	}
	if got := reconciler.focusMgr.GetCurrent(); got != baseFiber {
		t.Fatalf("focused fiber = %#v, want base fiber", got)
	}
	if !baseInst.focused {
		t.Fatal("base instance should receive focus when only empty portal hosts exist")
	}
}
