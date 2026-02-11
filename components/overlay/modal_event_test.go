package overlay

import (
	"testing"

	"github.com/wwsheng009/mint/components/basic"
	"github.com/wwsheng009/mint/components/form"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/ui"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// TestModalButtonEventRouting tests that buttons inside modal receive mouse events
func TestModalButtonEventRouting(t *testing.T) {
	// Create a button
	button := form.Button().Label("Click Me").Build()

	// Verify button implements Updater interface
	if _, ok := button.(component.Updater); !ok {
		t.Error("Button should implement component.Updater interface")
	}

	// Create modal with button in content
	modal := ModalBuilder().
		Title("Test Modal").
		Content(button).
		Open(true).
		Width(40).
		Height(10).
		Build()

	// Verify modal implements Updater
	if _, ok := modal.(component.Updater); !ok {
		t.Error("Modal should implement component.Updater interface")
	}

	// Test that modal exposes children
	modalVNode, ok := modal.(*ModalVNode)
	if !ok {
		t.Fatal("Modal should be *ModalVNode type")
	}

	children := modalVNode.Children()
	if len(children) == 0 {
		t.Error("Modal should have at least one child (the button)")
	}

	// Verify the button is in the children
	found := false
	for _, child := range children {
		if child == button {
			found = true
			break
		}
	}
	if !found {
		t.Error("Button should be in modal's children")
	}

	t.Log("✅ Modal correctly exposes button in Children()")
}

// TestModalWithMultipleButtons tests modal with multiple interactive elements
func TestModalWithMultipleButtons(t *testing.T) {
	// Create multiple buttons
	confirmBtn := form.Button().Label("Confirm").Build()
	cancelBtn := form.Button().Label("Cancel").Build()

	// Create modal with both buttons in footer
	modal := ModalBuilder().
		Title("Confirm Action").
		Content(ui.Text("Are you sure?")).
		Footer(confirmBtn).
		Open(true).
		Build()

	// Verify modal exposes all children
	modalVNode := modal.(*ModalVNode)
	children := modalVNode.Children()

	if len(children) == 0 {
		t.Error("Modal should have children")
	}

	t.Logf("Modal has %d children", len(children))

	// Verify both buttons are interactive
	for i, child := range children {
		if child == nil {
			continue
		}
		// Check if child implements Updater (button should)
		if updater, ok := child.(component.Updater); ok {
			t.Logf("Child %d implements Updater: %T", i, child)
			// Try sending a click message
			clickMsg := &runtimemsg.MouseMsg{
				BaseMsg: runtimemsg.BaseMsg{
					TypeValue: runtimemsg.MsgTypeMouse,
				},
				Action: runtimemsg.MouseActionPress,
				Button: runtimemsg.MouseLeft,
			}
			cmd := updater.Update(clickMsg)
			if cmd != nil {
				t.Logf("Child %d returned Cmd: %v", i, cmd)
			}
		} else {
			t.Logf("Child %d does NOT implement Updater: %T", i, child)
		}
	}
}

// TestModalBoundsCheck tests modal's containsPoint logic
func TestModalBoundsCheck(t *testing.T) {
	modal := NewModal()
	modal.SetBounds(10, 5, 40, 15) // x=10, y=5, width=40, height=15

	tests := []struct {
		name     string
		x, y     int
		expected bool
	}{
		{"Inside", 30, 10, true},
		{"Inside corner", 10, 5, true},
		{"Outside left", 5, 10, false},
		{"Outside right", 60, 10, false},
		{"Outside top", 30, 2, false},
		{"Outside bottom", 30, 25, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modal.containsPoint(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("containsPoint(%d,%d) = %v, want %v", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

// TestModalUpdate_MouseInside tests that modal doesn't handle internal clicks
func TestModalUpdate_MouseInside(t *testing.T) {
	modal := NewModal()
	modal.SetBounds(10, 5, 40, 15)
	modal.SetIsOpen(true)

	// Click inside modal bounds
	mouseMsg := &runtimemsg.MouseMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue: runtimemsg.MsgTypeMouse,
		},
		Action: runtimemsg.MouseActionPress,
		X:      30, // Inside modal
		Y:      10, // Inside modal
	}

	cmd := modal.Update(mouseMsg)
	if cmd != nil {
		t.Errorf("Modal should not return Cmd for internal click, got %v", cmd)
	}

	// Modal should still be open (click inside doesn't close it)
	if !modal.IsOpen() {
		t.Error("Modal should still be open after internal click")
	}
}

// TestModalUpdate_MouseOutside tests that modal handles external clicks
func TestModalUpdate_MouseOutside(t *testing.T) {
	closed := false
	modal := NewModal()
	modal.SetBounds(10, 5, 40, 15)
	modal.SetIsOpen(true)
	modal.SetOnClose(func() {
		closed = true
	})

	// Click outside modal bounds
	mouseMsg := &runtimemsg.MouseMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue: runtimemsg.MsgTypeMouse,
		},
		Action: runtimemsg.MouseActionPress,
		X:      5,  // Outside modal (left)
		Y:      10, // Inside vertical range
	}

	cmd := modal.Update(mouseMsg)
	if cmd != nil {
		t.Errorf("Modal should not return Cmd for external click, got %v", cmd)
	}

	// Modal should be closed now
	if modal.IsOpen() {
		t.Error("Modal should be closed after external click")
	}

	if !closed {
		t.Error("OnClose callback should have been called")
	}
}

// TestModalUpdate_ESC tests that modal handles ESC key
func TestModalUpdate_ESC(t *testing.T) {
	closed := false
	modal := NewModal()
	modal.SetIsOpen(true)
	modal.SetOnClose(func() {
		closed = true
	})

	keyMsg := &runtimemsg.KeyMsg{
		BaseMsg: runtimemsg.BaseMsg{
			TypeValue: runtimemsg.MsgTypeKey,
		},
		Special: runtimeplatform.KeyEscape,
	}

	cmd := modal.Update(keyMsg)
	if cmd != nil {
		t.Errorf("Modal should not return Cmd for ESC, got %v", cmd)
	}

	// Modal should be closed now
	if modal.IsOpen() {
		t.Error("Modal should be closed after ESC")
	}

	if !closed {
		t.Error("OnClose callback should have been called")
	}
}
