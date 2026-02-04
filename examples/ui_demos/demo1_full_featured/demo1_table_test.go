// demo1_table_test.go - Test demo1 with Table layout
package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

// TestDemo1TableLayout verifies the Table layout renders correctly
func TestDemo1TableLayout(t *testing.T) {
	app, err := ui.RunTestWithSandbox(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
		ui.WithTitle("Demo1 Table Test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Get rendered output
	output := app.GetRenderString()

	t.Log("=== Demo1 Table Layout Output ===")
	t.Log(output)
	t.Log("=== End of Output ===")

	// Verify key elements are present
	checks := []struct {
		name     string
		expected string
	}{
		{"Header border", "+"},
		{"Menu label", "Menu"},
		{"Input label", "Input:"},
		// Note: Buttons may be truncated due to column width, so we check for partial text
		{"Add Count button (partial)", "Add Cou"},
		{"Subtract Count button (partial)", "Subtrac"},
		{"Quit button (partial)", "Quit ["},
		{"Log Output", "Log Output"},
		{"VirtualList", "Log line #"},
	}

	passed := 0
	failed := 0
	for _, check := range checks {
		if !contains(output, check.expected) {
			t.Errorf("Expected to find %q in output", check.name)
			failed++
		} else {
			t.Logf("✓ Found %q", check.name)
			passed++
		}
	}

	t.Logf("Test summary: %d passed, %d failed", passed, failed)

	// Output already displayed above via t.Log
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
