package main

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/button"
)

// TestFocusStyleVisibility tests that focus style changes are visible
func TestFocusStyleVisibility(t *testing.T) {
	// Note: Theme is already initialized by the test framework

	// Create test buttons
	primaryBtn := ui.NewButtonBuilder("[Open Modal]").
		Variant(ui.ButtonVariantPrimary).
		FocusStyle(ui.FocusStyleBracket).
		Build()

	dangerBtn := ui.NewButtonBuilder("Quit").
		Variant(ui.ButtonVariantDanger).
		FocusStyle(ui.FocusStyleBracket).
		Build()

	// Get VNode and create Instance
	primaryBtnVNode, ok := primaryBtn.(*button.VNode)
	if !ok {
		t.Fatal("Expected ButtonVNode")
	}
	primaryInst := primaryBtnVNode.CreateInstance().(*button.Instance)

	dangerBtnVNode, ok := dangerBtn.(*button.VNode)
	if !ok {
		t.Fatal("Expected ButtonVNode")
	}
	dangerInst := dangerBtnVNode.CreateInstance().(*button.Instance)

	fmt.Println("\n=== Testing Focus Style Visibility ===")

	// Test 1: Primary button without focus
	fmt.Println("Test 1: Primary button WITHOUT focus")
	primaryInst.SetFocus(false)
	commandsNoFocus := primaryInst.Paint(10, 10)
	if len(commandsNoFocus) > 0 {
		cmd := commandsNoFocus[0]
		style := cmd.Style
		fmt.Printf("  Style: FG=%s, BG=%s, Bold=%v\n", style.FG, style.BG, style.IsBold())
		fmt.Printf("  Text: %q\n", cmd.Text)
	}

	// Test 2: Primary button WITH focus
	fmt.Println("\nTest 2: Primary button WITH focus")
	primaryInst.SetFocus(true)
	commandsWithFocus := primaryInst.Paint(10, 10)
	if len(commandsWithFocus) > 0 {
		cmd := commandsWithFocus[0]
		style := cmd.Style
		fmt.Printf("  Style: FG=%s, BG=%s, Bold=%v\n", style.FG, style.BG, style.IsBold())
		fmt.Printf("  Text: %q\n", cmd.Text)

		// Verify focus style is different from non-focus
		if style.FG == theme.FocusBright() {
			fmt.Println("  ✓ Focus foreground is FocusBright (yellow)")
		} else {
			t.Errorf("  ✗ Expected FocusBright, got %s", style.FG)
		}

		if style.BG != "" {
			fmt.Printf("  ✓ Background preserved: %s\n", style.BG)
		} else {
			t.Error("  ✗ Background should be preserved")
		}
	}

	// Test 3: Danger button without focus
	fmt.Println("\nTest 3: Danger button WITHOUT focus")
	dangerInst.SetFocus(false)
	commandsNoFocus = dangerInst.Paint(10, 20)
	if len(commandsNoFocus) > 0 {
		cmd := commandsNoFocus[0]
		style := cmd.Style
		fmt.Printf("  Style: FG=%s, BG=%s, Bold=%v\n", style.FG, style.BG, style.IsBold())
	}

	// Test 4: Danger button WITH focus
	fmt.Println("\nTest 4: Danger button WITH focus")
	dangerInst.SetFocus(true)
	commandsWithFocus = dangerInst.Paint(10, 20)
	if len(commandsWithFocus) > 0 {
		cmd := commandsWithFocus[0]
		style := cmd.Style
		fmt.Printf("  Style: FG=%s, BG=%s, Bold=%v\n", style.FG, style.BG, style.IsBold())

		// Verify focus style is different from non-focus
		if style.FG == theme.FocusBright() {
			fmt.Println("  ✓ Focus foreground is FocusBright (yellow)")
		} else {
			t.Errorf("  ✗ Expected FocusBright, got %s", style.FG)
		}

		if style.BG != "" {
			fmt.Printf("  ✓ Background preserved: %s\n", style.BG)
		} else {
			t.Error("  ✗ Background should be preserved")
		}
	}

	// Test 5: ANSI output comparison
	fmt.Println("\n=== ANSI Output Comparison ===")

	// Generate ANSI for non-focus
	primaryInst.SetFocus(false)
	commandsNoFocus = primaryInst.Paint(0, 0)

	// Generate ANSI for focus
	primaryInst.SetFocus(true)
	commandsWithFocus = primaryInst.Paint(0, 0)

	if len(commandsNoFocus) > 0 && len(commandsWithFocus) > 0 {
		noFocusCmd := commandsNoFocus[0]
		focusCmd := commandsWithFocus[0]

		// Create a simple style state machine to generate ANSI
		ssm := paint.NewStyleStateMachine()

		noFocusANSI := ssm.Update(noFocusCmd.Style)
		focusANSI := ssm.Update(focusCmd.Style)

		fmt.Printf("Non-focus ANSI: %q\n", noFocusANSI)
		fmt.Printf("Focus ANSI:     %q\n", focusANSI)

		if noFocusANSI != focusANSI {
			fmt.Println("✓ ANSI output is DIFFERENT between focus states")
		} else {
			t.Error("✗ ANSI output is SAME - focus style not working!")
		}
	}

	fmt.Println("\n=== Test Complete ===")
}

// TestFocusIndicator verifies that focus indicator (>) is shown
func TestFocusIndicator(t *testing.T) {
	// Note: Theme is already initialized by the test framework

	btn := ui.NewButtonBuilder("[Test]").
		Variant(ui.ButtonVariantPrimary).
		FocusStyle(ui.FocusStyleBracket).
		Build()

	btnVNode, ok := btn.(*button.VNode)
	if !ok {
		t.Fatal("Expected ButtonVNode")
	}
	btnInst := btnVNode.CreateInstance().(*button.Instance)

	fmt.Println("\n=== Testing Focus Indicator ===")

	// Without focus
	btnInst.SetFocus(false)
	commandsNoFocus := btnInst.Paint(0, 0)
	if len(commandsNoFocus) > 0 {
		cmd := commandsNoFocus[0]
		fmt.Printf("Without focus: %q\n", cmd.Text)
		if cmd.Text[0] != '>' {
			fmt.Println("✓ No focus indicator")
		} else {
			t.Error("✗ Focus indicator shown when not focused")
		}
	}

	// With focus
	btnInst.SetFocus(true)
	commandsWithFocus := btnInst.Paint(0, 0)
	if len(commandsWithFocus) > 0 {
		cmd := commandsWithFocus[0]
		fmt.Printf("With focus:    %q\n", cmd.Text)
		if cmd.Text[0] == '>' {
			fmt.Println("✓ Focus indicator shown")
		} else {
			t.Error("✗ Focus indicator missing")
		}
	}
}
