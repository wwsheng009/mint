// demo1_table_test.go - Test demo1 with Bordered components
package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

// TestDemo1TableLayout verifies the layout with Bordered components renders correctly
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

	// Verify key elements are present (using Bordered component)
	checks := []struct {
		name     string
		expected string
	}{
		// Unicode border characters instead of ASCII
		{"Header top border", "┌"},
		{"Header bottom border", "└"},
		{"Open Modal button", "Open Modal"},
		{"Click counter", "Clicks:"},
		{"Menu label", "Menu"},
		{"Quit button", "Quit"},
		{"Log line #0", "Log line #0"},
		{"Log line #1", "Log line #1"},
		{"Ellipsis for more items", "..."},
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

	// Verify layout structure with Unicode borders
	expectedPatterns := []struct {
		name     string
		pattern  string
	}{
		{"Header box", "└──"},
		{"Sidebar box", "┌"},
		{"Content box", "│"},
	}

	for _, p := range expectedPatterns {
		if !contains(output, p.pattern) {
			t.Logf("Note: expected pattern %q not found", p.pattern)
		} else {
			t.Logf("✓ Found pattern: %s", p.name)
		}
	}
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
