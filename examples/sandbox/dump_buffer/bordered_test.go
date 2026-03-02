// bordered_full_test.go - Full test for Bordered component
package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/stack"
)

func TestBorderedSimple(t *testing.T) {
	app, err := ui.RunTestWithSandbox(SimpleBorderedApp,
		ui.WithWidth(40),
		ui.WithHeight(15),
		ui.WithTitle("Bordered Simple Test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Simple Bordered Output ===")
	t.Log(output)
	t.Log("=== End ===")

	// Verify all border characters are present
	checks := []string{"┌", "─", "┐", "│", "└", "┘", "Hello"}
	for _, expected := range checks {
		if !contains(output, expected) {
			t.Errorf("Missing: %q", expected)
		} else {
			t.Logf("✓ Found: %q", expected)
		}
	}
}

func TestBorderedWithLabel(t *testing.T) {
	app, err := ui.RunTestWithSandbox(BorderedWithLabelApp,
		ui.WithWidth(40),
		ui.WithHeight(15),
		ui.WithTitle("Bordered Label Test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== With Label Output ===")
	t.Log(output)
	t.Log("=== End ===")

	checks := []string{"┌", "─", "┐", "│", "└", "┘", "Title", "Hello"}
	for _, expected := range checks {
		if !contains(output, expected) {
			t.Errorf("Missing: %q", expected)
		} else {
			t.Logf("✓ Found: %q", expected)
		}
	}
}

func TestBorderedMultiLine(t *testing.T) {
	app, err := ui.RunTestWithSandbox(BorderedMultiLineApp,
		ui.WithWidth(40),
		ui.WithHeight(15),
		ui.WithTitle("Bordered MultiLine Test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== MultiLine Output ===")
	t.Log(output)
	t.Log("=== End ===")

	checks := []string{"┌", "─", "┐", "│", "└", "┘", "Line 1", "Line 2"}
	for _, expected := range checks {
		if !contains(output, expected) {
			t.Errorf("Missing: %q", expected)
		} else {
			t.Logf("✓ Found: %q", expected)
		}
	}
}

func TestBorderedDoubleStyle(t *testing.T) {
	app, err := ui.RunTestWithSandbox(BorderedDoubleStyleApp,
		ui.WithWidth(40),
		ui.WithHeight(15),
		ui.WithTitle("Bordered Double Style Test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Double Style Output ===")
	t.Log(output)
	t.Log("=== End ===")

	checks := []string{"╔", "═", "╗", "║", "╚", "╝", "Hello"}
	for _, expected := range checks {
		if !contains(output, expected) {
			t.Errorf("Missing: %q", expected)
		} else {
			t.Logf("✓ Found: %q", expected)
		}
	}
}

func TestBorderedDashedStyle(t *testing.T) {
	app, err := ui.RunTestWithSandbox(BorderedDashedStyleApp,
		ui.WithWidth(40),
		ui.WithHeight(15),
		ui.WithTitle("Bordered Dashed Style Test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Dashed Style Output ===")
	t.Log(output)
	t.Log("=== End ===")

	checks := []string{"+", "-", "|", "Hello"}
	for _, expected := range checks {
		if !contains(output, expected) {
			t.Errorf("Missing: %q", expected)
		} else {
			t.Logf("✓ Found: %q", expected)
		}
	}
}

func SimpleBorderedApp() ui.VNode {
	return stack.NewVStack().SingleBorder().SetChildrenList([]ui.VNode{
		ui.Text("Hello"),
	})
}

func BorderedWithLabelApp() ui.VNode {
	return stack.NewVStack().SingleBorder("Title").SetChildrenList([]ui.VNode{
		ui.Text("Hello"),
	})
}

func BorderedMultiLineApp() ui.VNode {
	return stack.NewVStack().SingleBorder().SetChildrenList([]ui.VNode{
		ui.VStack(
			ui.Text("Line 1"),
			ui.Text("Line 2"),
			ui.Text("Line 3"),
		),
	})
}

func BorderedDoubleStyleApp() ui.VNode {
	return stack.NewVStack().DoubleBorder().SetChildrenList([]ui.VNode{
		ui.Text("Hello"),
	})
}

func BorderedDashedStyleApp() ui.VNode {
	return stack.NewVStack().DashedBorder().SetChildrenList([]ui.VNode{
		ui.Text("Hello"),
	})
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
