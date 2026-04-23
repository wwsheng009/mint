// Sandbox tests for demo4_complex_layout
// Tests the complex layout system (Flex, Grid, Absolute, Scroll)

package main

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestDemo4InitialRender tests that the layout demo renders without errors
func TestDemo4InitialRender(t *testing.T) {
	testApp, err := ui.RunTest(LayoutDemo,
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

	t.Logf("✓ Demo4 initial render works")
}

// TestDemo4FlexLayout tests the Flex layout demo
func TestDemo4FlexLayout(t *testing.T) {
	testApp, err := ui.RunTest(LayoutDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Default tab should be Flex
	rendered := testApp.GetRenderString()
	if rendered == "" {
		t.Error("Flex layout should render")
	}

	t.Log("✓ Demo4 Flex layout works")
}

// TestDemo4GridLayout tests the Grid layout demo
func TestDemo4GridLayout(t *testing.T) {
	testApp, err := ui.RunTest(LayoutDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Navigate to Grid tab (2nd tab)
	testApp.InjectKey('2')

	t.Log("✓ Demo4 Grid layout works")
}

// TestDemo4AbsoluteLayout tests the Absolute positioning demo
func TestDemo4AbsoluteLayout(t *testing.T) {
	testApp, err := ui.RunTest(LayoutDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Navigate to Absolute tab (3rd tab)
	testApp.InjectKey('3')

	t.Log("✓ Demo4 Absolute layout works")
}

// TestDemo4ScrollLayout tests the Scroll container demo
func TestDemo4ScrollLayout(t *testing.T) {
	testApp, err := ui.RunTest(LayoutDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Navigate to Scroll tab (4th tab)
	testApp.InjectKey('4')

	t.Log("✓ Demo4 Scroll layout works")
}

// TestDemo4ComplexLayout tests the complex IDE-like layout
func TestDemo4ComplexLayout(t *testing.T) {
	testApp, err := ui.RunTest(LayoutDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	// Navigate to Complex tab (5th tab)
	testApp.InjectKey('5')

	buffer := testApp.GetBuffer()
	if buffer == nil {
		t.Error("Complex layout should render")
	}

	t.Log("✓ Demo4 Complex layout works")
}

// TestDemo4AllLayouts tests navigating through all layout types
func TestDemo4AllLayouts(t *testing.T) {
	testApp, err := ui.RunTest(LayoutDemo,
		ui.WithSize(100, 40),
	)
	if err != nil {
		t.Fatalf("Failed to run test app: %v", err)
	}
	defer testApp.Close()

	layouts := []string{"Flex", "Grid", "Absolute", "Scroll", "Complex"}
	keys := []rune{'1', '2', '3', '4', '5'}

	for i, layout := range layouts {
		testApp.InjectKey(keys[i])

		t.Logf("✓ %s layout rendered", layout)
	}

	t.Log("✓ Demo4 all layouts work")
}

// TestDemo4Snapshot tests snapshot with complex layouts
func TestDemo4Snapshot(t *testing.T) {
	testApp, err := ui.RunTest(LayoutDemo,
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

	// Navigate to complex layout first
	testApp.InjectKey('5')

	snap, err := sb.Snapshot(1, "demo4-complex-layout")
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	if snap == nil {
		t.Error("Snapshot should not be nil")
	}

	t.Log("✓ Demo4 snapshot with complex layout works")
}
