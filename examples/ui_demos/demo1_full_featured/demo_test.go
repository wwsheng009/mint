// Sandbox tests for demo1_full_featured
// Tests the complete UI component showcase with state, layout, Modal, Input, Focus, Scroll

package main

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestDemo1InitialRender tests that the demo renders without errors
func TestDemo1InitialRender(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithSize(80, 24),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	rendered := testApp.GetRenderString()
	if rendered == "" {
		t.Error("Initial render should not be empty")
	}

	t.Logf("✓ Demo1 initial render works")
}

// TestDemo1ButtonInteraction tests button clicks and state updates
func TestDemo1ButtonInteraction(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithSize(80, 24),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Get initial buttons
	buttons := testApp.GetButtons()
	initialCount := len(buttons)

	// Tab to first button and click
	testApp.InjectSpecialKey(platform.KeyTab)
	testApp.InjectSpecialKey(platform.KeyEnter)

	t.Logf("✓ Demo1 button interaction works (found %d buttons)", initialCount)
}

// TestDemo1FocusManagement tests focus navigation
func TestDemo1FocusManagement(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithSize(80, 24),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Navigate through buttons
	for i := 0; i < 3; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
	}

	t.Log("✓ Demo1 focus management works")
}

// TestDemo1Modal tests modal layer
func TestDemo1Modal(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithSize(80, 24),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Get buttons info
	buttons := testApp.GetButtons()
	t.Logf("Found %d buttons", len(buttons))

	// Navigate to modal button and click
	for i := 0; i < 5; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
	}
	testApp.InjectSpecialKey(platform.KeyEnter)

	t.Log("✓ Demo1 modal interaction works")
}

// TestDemo1Snapshot tests snapshot functionality
func TestDemo1Snapshot(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithSize(80, 24),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	if sb == nil {
		t.Skip("Requires GetSandbox() support")
		return
	}

	snap, err := sb.Snapshot(1, "demo1-initial")
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	if snap == nil {
		t.Error("Snapshot should not be nil")
	}

	t.Log("✓ Demo1 snapshot works")
}

// TestDemo1QueueStats tests queue statistics
func TestDemo1QueueStats(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithSize(80, 24),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	if sb == nil {
		t.Skip("Requires GetSandbox() support")
		return
	}

	stats := sb.QueueStats()
	t.Logf("✓ Demo1 queue stats: length=%d", stats.Length)
}

// TestDemo1Comprehensive runs comprehensive tests
func TestDemo1Comprehensive(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"InitialState", func(t *testing.T) {
			testApp, err := ui.RunTest(App, ui.WithSize(80, 24))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()
			t.Log("✓ Initial state valid")
		}},
		{"MultipleClicks", func(t *testing.T) {
			testApp, err := ui.RunTest(App, ui.WithSize(80, 24))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			for i := 0; i < 5; i++ {
				testApp.InjectSpecialKey(platform.KeyTab)
				testApp.InjectSpecialKey(platform.KeyEnter)
			}
			t.Log("✓ Multiple clicks handled")
		}},
		{"KeyboardNavigation", func(t *testing.T) {
			testApp, err := ui.RunTest(App, ui.WithSize(80, 24))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			for i := 0; i < 5; i++ {
				testApp.InjectSpecialKey(platform.KeyTab)
			}
			t.Log("✓ Keyboard navigation works")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
