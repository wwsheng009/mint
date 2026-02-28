package button

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestVNode_FlexProps(t *testing.T) {
	// Test default flex value
	btn := New("Test")
	if btn.flex != 0 {
		t.Errorf("Expected default flex to be 0, got %d", btn.flex)
	}

	// Test Props() includes flex
	props := btn.Props()
	if props["flex"] != 0 {
		t.Errorf("Expected Props() flex to be 0, got %v", props["flex"])
	}
}

func TestVNode_SetFlex(t *testing.T) {
	btn := New("Test")

	// Test setting flex via SetProps - need to use concrete type
	btn.SetProps(rtui.Props{"flex": 2})
	if btn.flex != 2 {
		t.Errorf("Expected flex to be 2 after SetProps, got %d", btn.flex)
	}

	// Test Props() reflects the change
	props := btn.Props()
	if props["flex"] != 2 {
		t.Errorf("Expected Props() flex to be 2, got %v", props["flex"])
	}
}

func TestVNode_FlexIntegration(t *testing.T) {
	// Test flex propagation through to instance
	btn := New("Test")
	btn.flex = 3

	// Create instance should have flex in props
	instance := btn.CreateInstance()
	props := instance.GetProps()
	flexVal, ok := props["flex"].(int)
	if !ok {
		t.Error("Instance.GetProps() should return flex as int")
	}
	if flexVal != 3 {
		t.Errorf("Expected flex=3 in Instance.GetProps(), got %d", flexVal)
	}
}
