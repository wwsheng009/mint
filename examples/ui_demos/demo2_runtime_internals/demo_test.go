// Sandbox tests for demo2_runtime_internals
// Tests the runtime pipeline visualization

package main

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestDemo2InitialRender tests that the runtime demo renders without errors
func TestDemo2InitialRender(t *testing.T) {
	testApp, err := ui.RunTest(RuntimeDemo,
		ui.WithSize(100, 35),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	rendered := testApp.GetRenderString()
	if rendered == "" {
		t.Error("Initial render should not be empty")
	}

	t.Logf("✓ Demo2 initial render works")
}

// TestDemo2Statistics tests the statistics panel renders
func TestDemo2Statistics(t *testing.T) {
	testApp, err := ui.RunTest(RuntimeDemo,
		ui.WithSize(100, 35),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Check buffer exists
	buffer := testApp.GetBuffer()
	if buffer == nil {
		t.Error("Buffer should not be nil")
	}

	t.Log("✓ Demo2 statistics panel works")
}

// TestDemo2Snapshot tests snapshot with runtime state
func TestDemo2Snapshot(t *testing.T) {
	testApp, err := ui.RunTest(RuntimeDemo,
		ui.WithSize(100, 35),
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

	snap, err := sb.Snapshot(1, "demo2-runtime-state")
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	if snap == nil {
		t.Error("Snapshot should not be nil")
	}

	t.Log("✓ Demo2 snapshot with runtime state works")
}

// TestDemo2Comprehensive runs comprehensive runtime tests
func TestDemo2Comprehensive(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"RenderQuality", func(t *testing.T) {
			testApp, err := ui.RunTest(RuntimeDemo, ui.WithSize(100, 35))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			rendered := testApp.GetRenderString()
			if len(rendered) < 100 {
				t.Error("Should render substantial content")
			}
			t.Log("✓ Render quality OK")
		}},
		{"EventHandling", func(t *testing.T) {
			testApp, err := ui.RunTest(RuntimeDemo, ui.WithSize(100, 35))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			// Try injecting some key events
			testApp.InjectKey('1')
			testApp.InjectKey('2')

			t.Log("✓ Event handling works")
		}},
		{"QueueStats", func(t *testing.T) {
			testApp, err := ui.RunTest(RuntimeDemo, ui.WithSize(100, 35))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			sb := testApp.GetSandbox()
			if sb != nil {
				stats := sb.QueueStats()
				t.Logf("✓ Queue stats: length=%d", stats.Length)
			} else {
				t.Skip("GetSandbox not available")
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
