package main

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/panel"
)

// TestSimplePanelRendering tests basic Panel rendering
func TestSimplePanelRendering(t *testing.T) {
	// Create a simple panel
	vnode := panel.New().
		SetTitle("Test").
		SetWidth(20).
		SetHeight(5).
		SetContent(newtext.New("Hello")).
		Rounded()

	// Create instance
	inst := vnode.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	// Get props
	props := inst.GetProps()
	fmt.Printf("Instance Props: %+v\n", props)

	// Check borderLabel
	if props["borderLabel"] != " Test " {
		t.Errorf("Expected borderLabel ' Test ', got '%v'", props["borderLabel"])
	}

	// Paint
	if paintable, ok := inst.(interface{ Paint(int, int) []paint.DrawCmd }); ok {
		cmds := paintable.Paint(0, 0)
		fmt.Printf("Paint commands: %d\n", len(cmds))
		for i, cmd := range cmds {
			fmt.Printf("  [%d] x=%d y=%d text=%q\n", i, cmd.X, cmd.Y, cmd.Text)
		}

		// Check that top border contains label
		if len(cmds) > 0 {
			if !contains(cmds[0].Text, "Test") {
				t.Errorf("Top border should contain 'Test', got: %s", cmds[0].Text)
			}
		}
	} else {
		t.Error("Instance is not paintable")
	}
}

// TestPanelFiberCreation tests Panel through Fiber creation
func TestPanelFiberCreation(t *testing.T) {
	vnode := panel.New().
		SetTitle("Fiber Test").
		SetWidth(20).
		SetHeight(5).
		SetContent(newtext.New("Content")).
		Rounded()

	// Create Fiber
	fiber := ui.CreateFiber(vnode)
	if fiber == nil {
		t.Fatal("CreateFiber returned nil")
	}

	fmt.Printf("Fiber Tag: %s\n", fiber.Tag)
	fmt.Printf("Fiber Instance: %v\n", fiber.Instance != nil)

	if fiber.Instance != nil {
		props := fiber.Instance.GetProps()
		fmt.Printf("Fiber Instance Props: %+v\n", props)

		if props["borderLabel"] != " Fiber Test " {
			t.Errorf("Expected borderLabel ' Fiber Test ', got '%v'", props["borderLabel"])
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
