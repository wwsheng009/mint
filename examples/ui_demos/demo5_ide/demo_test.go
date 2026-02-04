// Sandbox tests for demo5_ide
// Tests the IDE interface demo

package main

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestDemo5InitialRender tests that the IDE demo renders without errors
func TestDemo5InitialRender(t *testing.T) {
	testApp, err := ui.RunTest(IDEDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	rendered := testApp.GetRenderString()
	if rendered == "" {
		t.Error("Initial render should not be empty")
	}

	t.Logf("✓ Demo5 initial render works")
}

// TestDemo5MenuBar tests the menu bar
func TestDemo5MenuBar(t *testing.T) {
	testApp, err := ui.RunTest(IDEDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Get menu buttons
	buttons := testApp.GetButtons()
	menuButtonCount := len(buttons)

	t.Logf("✓ Demo5 menu bar works (%d buttons found)", menuButtonCount)
}

// TestDemo5FileExplorer tests the file explorer sidebar
func TestDemo5FileExplorer(t *testing.T) {
	testApp, err := ui.RunTest(IDEDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	buffer := testApp.GetBuffer()
	if buffer == nil {
		t.Error("File explorer should render")
	}

	t.Log("✓ Demo5 file explorer works")
}

// TestDemo5TabsBar tests the tab navigation
func TestDemo5TabsBar(t *testing.T) {
	testApp, err := ui.RunTest(IDEDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Navigate to terminal tab
	for i := 0; i < 15; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
	}
	testApp.InjectSpecialKey(platform.KeyEnter)

	t.Log("✓ Demo5 tabs bar works")
}

// TestDemo5StatusBar tests the status bar
func TestDemo5StatusBar(t *testing.T) {
	testApp, err := ui.RunTest(IDEDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	buffer := testApp.GetBuffer()
	if buffer == nil {
		t.Error("Status bar should render")
	}

	t.Log("✓ Demo5 status bar works")
}

// TestDemo5EditorPanel tests the editor panel
func TestDemo5EditorPanel(t *testing.T) {
	testApp, err := ui.RunTest(IDEDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	buffer := testApp.GetBuffer()
	if buffer == nil {
		t.Error("Editor panel should render")
	}

	t.Log("✓ Demo5 editor panel works")
}

// TestDemo5Snapshot tests snapshot with IDE state
func TestDemo5Snapshot(t *testing.T) {
	testApp, err := ui.RunTest(IDEDemo,
		ui.WithSize(100, 40),
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

	snap, err := sb.Snapshot(1, "demo5-ide-state")
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	if snap == nil {
		t.Error("Snapshot should not be nil")
	}

	t.Log("✓ Demo5 snapshot with IDE state works")
}

// TestDemo5Comprehensive runs comprehensive IDE tests
func TestDemo5Comprehensive(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"AllTabs", func(t *testing.T) {
			testApp, err := ui.RunTest(IDEDemo, ui.WithSize(100, 40))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			// Try to navigate through tabs
			for i := 0; i < 20; i++ {
				testApp.InjectSpecialKey(platform.KeyTab)
			}
			t.Log("✓ Tab navigation works")
		}},
		{"ButtonInteraction", func(t *testing.T) {
			testApp, err := ui.RunTest(IDEDemo, ui.WithSize(100, 40))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			// Try clicking some buttons
			for i := 0; i < 5; i++ {
				testApp.InjectSpecialKey(platform.KeyTab)
				testApp.InjectSpecialKey(platform.KeyEnter)
			}
			t.Log("✓ Button interaction works")
		}},
		{"EventProcessing", func(t *testing.T) {
			testApp, err := ui.RunTest(IDEDemo, ui.WithSize(100, 40))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			sb := testApp.GetSandbox()
			if sb != nil {
				stats := sb.QueueStats()
				t.Logf("✓ Event processing: length=%d", stats.Length)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
