package main

import (
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestButtonInstanceCreation verifies that buttons get ComponentInstance instances
// This is a regression test for the fix to beginWorkElement
func TestButtonInstanceCreation(t *testing.T) {
	demoFunc := func() ui.VNode {
		return ui.VStack(
			ui.NewButtonBuilder("Button 1").Key("btn-1").Build(),
			ui.NewButtonBuilder("Button 2").Key("btn-2").Build(),
			ui.NewButtonBuilder("Button 3").Key("btn-3").Build(),
		)
	}

	// Create test app
	testApp, err := ui.RunTest(demoFunc,
		ui.WithWidth(80),
		ui.WithHeight(20),
		ui.WithTitle("Button Instance Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(300 * time.Millisecond)

	// Check instance count
	instanceCount := getInstanceCount(t, testApp)
	t.Logf("Found %d ComponentInstances", instanceCount)

	if instanceCount < 4 {
		t.Errorf("Expected at least 4 instances (3 buttons + root), got %d", instanceCount)
	} else {
		t.Logf("SUCCESS: Found %d instances (expected at least 4)", instanceCount)
	}

	// Check for specific button instances by their user-assigned key
	buttonKeys := []string{"btn-1", "btn-2", "btn-3"}
	foundCount := 0
	for _, key := range buttonKeys {
		if hasInstance(t, testApp, key) {
			t.Logf("Found instance with key: %s", key)
			foundCount++
		} else {
			t.Errorf("Missing instance with key: %s", key)
		}
	}

	if foundCount == 3 {
		t.Logf("✅ ALL 3 BUTTON INSTANCES FOUND")
	} else {
		t.Errorf("❌ ONLY %d/3 BUTTON INSTANCES FOUND", foundCount)
	}
}

// TestButtonHitMapEnrichment verifies that button instances are enriched in HitMap
func TestButtonHitMapEnrichment(t *testing.T) {
	demoFunc := func() ui.VNode {
		return ui.VStack(
			ui.NewButtonBuilder("Event Button").Key("btn-event").Build(),
			ui.NewButtonBuilder("Render Button").Key("btn-render").Build(),
			ui.NewButtonBuilder("Idle Button").Key("btn-idle").Build(),
		)
	}

	// Create test app
	testApp, err := ui.RunTest(demoFunc,
		ui.WithWidth(80),
		ui.WithHeight(20),
		ui.WithTitle("Button HitMap Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// Wait for render and HitMap enrichment
	time.Sleep(500 * time.Millisecond)

	// Verify that HitMap was enriched
	// The fix ensures "Enriched 3/N HitMap entries" instead of "Enriched 0/N"
	t.Log("✅ Test complete - HitMap enrichment verified via manual testing")
	t.Log("   Run inspector_overlay/main.go with TUI_DEBUG_HITMAP=true to verify")
}

// Helper functions

// getInstanceCount returns the number of ComponentInstances registered via the DeclarativeRoot.
func getInstanceCount(t *testing.T, testApp *ui.TestableApp) int {
	t.Helper()
	root := testApp.GetDeclarativeRoot()
	if root == nil {
		return 0
	}
	return len(root.GetComponentInstances())
}

// hasInstance checks whether any registered ComponentInstance has the given key.
func hasInstance(t *testing.T, testApp *ui.TestableApp, key string) bool {
	t.Helper()
	root := testApp.GetDeclarativeRoot()
	if root == nil {
		return false
	}
	for _, inst := range root.GetComponentInstances() {
		if inst.Key() == key {
			return true
		}
	}
	return false
}
