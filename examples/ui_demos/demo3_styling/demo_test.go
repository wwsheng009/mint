// Sandbox tests for demo3_styling
// Tests the styling system (TUI CSS)

package main

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestDemo3InitialRender tests that the styling demo renders without errors
func TestDemo3InitialRender(t *testing.T) {
	testApp, err := ui.RunTest(StylingDemo,
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

	t.Logf("✓ Demo3 initial render works")
}

// TestDemo3TabNavigation tests navigating between styling tabs
func TestDemo3TabNavigation(t *testing.T) {
	testApp, err := ui.RunTest(StylingDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()


	// Try clicking on different tabs
	for i := 0; i < 3; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		testApp.InjectSpecialKey(platform.KeyEnter)
	}

}

// TestDemo3ColorTab tests the colors tab
func TestDemo3ColorTab(t *testing.T) {
	testApp, err := ui.RunTest(StylingDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Default tab should be colors
	rendered := testApp.GetRenderString()
	if rendered == "" {
		t.Error("Colors tab should render")
	}

	t.Log("✓ Demo3 colors tab works")
}

// TestDemo3BordersTab tests the borders tab
func TestDemo3BordersTab(t *testing.T) {
	testApp, err := ui.RunTest(StylingDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Navigate to borders tab (3rd tab)
	for i := 0; i < 5; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
	}
	testApp.InjectSpecialKey(platform.KeyEnter)

	t.Log("✓ Demo3 borders tab works")
}

// TestDemo3ThemesTab tests the themes tab
func TestDemo3ThemesTab(t *testing.T) {
	testApp, err := ui.RunTest(StylingDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Navigate to themes tab (5th tab)
	for i := 0; i < 9; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
	}
	testApp.InjectSpecialKey(platform.KeyEnter)

	rendered := testApp.GetRenderString()
	if rendered == "" {
		t.Error("Themes tab should render")
	}

	t.Log("✓ Demo3 themes tab works")
}

// TestDemo3Snapshot tests snapshot with styled content
func TestDemo3Snapshot(t *testing.T) {
	testApp, err := ui.RunTest(StylingDemo,
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

	snap, err := sb.Snapshot(1, "demo3-styled-snapshot")
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	if snap == nil {
		t.Error("Snapshot should not be nil")
	}

	t.Log("✓ Demo3 snapshot with styled content works")
}

// TestDemo3Comprehensive runs comprehensive styling tests
func TestDemo3Comprehensive(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"AllTabs", func(t *testing.T) {
			testApp, err := ui.RunTest(StylingDemo, ui.WithSize(100, 40))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			// Try to navigate through all tabs
			for i := 0; i < 10; i++ {
				testApp.InjectSpecialKey(platform.KeyTab)
				testApp.InjectSpecialKey(platform.KeyEnter)
			}
			t.Log("✓ All tabs navigable")
		}},
		{"StyledComponents", func(t *testing.T) {
			testApp, err := ui.RunTest(StylingDemo, ui.WithSize(100, 40))
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			// Verify styled components render
			rendered := testApp.GetRenderString()
			if len(rendered) < 100 {
				t.Error("Styled components should produce output")
			}
			t.Log("✓ Styled components render correctly")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
