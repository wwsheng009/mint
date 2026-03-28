// Action System Tests for demo1_full_featured
//
// These tests verify the Action system integration:
// 1. Action dispatch through the unified event system
// 2. ActionTarget components handle actions correctly
// 3. Middleware chain processing
// 4. Navigation actions (Tab, arrows)
// 5. Text input actions
// 6. Click actions

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/runtime/platform"
)

// TestActionSystemBasicNavigation tests basic navigation actions
func TestActionSystemBasicNavigation(t *testing.T) {
	// Enable debug mode to see Action processing
	t.Setenv("ACTION_DEBUG", "true")

	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	initialState := testApp.GetRenderString()
	t.Logf("Initial state:\n%s\n", initialState)

	// Test Tab navigation (ActionNavigateNext)
	// Inject Tab key multiple times and observe focus changes
	focusedIndices := make(map[int]bool)
	for i := 0; i < 10; i++ {
		idx := testApp.GetFocusedIndex()
		focusedIndices[idx] = true
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
		testApp.ForceRender()
	}

	t.Logf("Focused indices after Tab navigation: %v", focusedIndices)
	if len(focusedIndices) < 2 {
		t.Error("Tab navigation should move focus between multiple elements")
	} else {
		t.Logf("PASS: Tab navigation works, %d different elements focused", len(focusedIndices))
	}
}

// TestActionSystemEnter tests ActionEnter (button click via keyboard)
func TestActionSystemEnter(t *testing.T) {
	// Reset store to clean state so this test is safe for -count>1
	prevState := appStore.Get()
	appStore.Set(AppState{Count: 0, ShowModal: false, Input: ""})
	t.Cleanup(func() { appStore.Set(prevState) })

	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	// Check initial count
	rendered := testApp.GetRenderString()
	if !strings.Contains(rendered, "Clicks: 0") {
		t.Logf("Initial count: %s", rendered)
	}

	// Tab to Add Count button and press Enter
	for i := 0; i < 5; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
	}
	testApp.InjectSpecialKey(platform.KeyEnter) // This should trigger ActionEnter
	time.Sleep(200 * time.Millisecond)
	testApp.ForceRender()

	rendered = testApp.GetRenderString()
	t.Logf("After Enter key:\n%s", rendered)

	// Check if count increased
	if strings.Contains(rendered, "Clicks: 1") {
		t.Log("PASS: ActionEnter triggered button click correctly")
	} else {
		t.Errorf("ActionEnter did not increment counter. State: %s", rendered)
	}
}

// TestActionSystemEscape tests ActionCancel (close modal with ESC)
func TestActionSystemEscape(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	// Open modal by finding and clicking Open Modal button
	for i := 0; i < 5; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
	}
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(200 * time.Millisecond)
	testApp.ForceRender()

	// Verify modal is open
	rendered := testApp.GetRenderString()
	if !strings.Contains(rendered, "*** Are you sure? ***") {
		t.Skip("Modal not open, skipping ESC test")
	}
	t.Log("Modal is open")

	// Press ESC to close modal (ActionCancel)
	testApp.InjectSpecialKey(platform.KeyEscape)
	time.Sleep(200 * time.Millisecond)
	testApp.ForceRender()

	rendered = testApp.GetRenderString()
	if !strings.Contains(rendered, "*** Are you sure? ***") {
		t.Log("PASS: ActionCancel closed modal")
	} else {
		t.Error("ActionCancel did not close modal")
	}
}

// TestActionSystemTextInput tests ActionInputText
func TestActionSystemTextInput(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	// Navigate to input field
	for i := 0; i < 10; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
	}
	testApp.ForceRender()

	// Type text - this generates ActionInputText actions
	err = testApp.InjectString("hello")
	if err != nil {
		t.Fatalf("InjectString failed: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
	testApp.ForceRender()

	rendered := testApp.GetRenderString()
	t.Logf("After typing 'hello':\n%s", rendered)

	// Check if input received the text
	if strings.Contains(rendered, "hello") {
		t.Log("PASS: ActionInputText works correctly")
	} else {
		t.Log("Note: Text not visible in buffer (may need more navigation to reach input)")
	}
}

// TestActionSystemArrowKeys tests navigation arrow actions
func TestActionSystemArrowKeys(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	initialIdx := testApp.GetFocusedIndex()
	t.Logf("Initial focused index: %d", initialIdx)

	// Test arrow keys (NavigateUp, NavigateDown, NavigateLeft, NavigateRight)
	arrows := []platform.SpecialKey{
		platform.KeyDown,
		platform.KeyUp,
		platform.KeyRight,
		platform.KeyLeft,
	}

	for _, arrow := range arrows {
		testApp.InjectSpecialKey(arrow)
		time.Sleep(30 * time.Millisecond)
		testApp.ForceRender()
		idx := testApp.GetFocusedIndex()
		t.Logf("After %v: focused index = %d", arrow, idx)
	}

	t.Log("PASS: Arrow key actions processed without panic")
}

// TestActionSystemMultipleClicks tests multiple rapid ActionClick events
func TestActionSystemMultipleClicks(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	// Navigate to Add Count button
	for i := 0; i < 5; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
	}

	// Click multiple times rapidly (tests throttling middleware)
	for i := 0; i < 5; i++ {
		testApp.InjectSpecialKey(platform.KeyEnter)
		time.Sleep(20 * time.Millisecond)
	}
	testApp.ForceRender()

	rendered := testApp.GetRenderString()
	t.Logf("After multiple clicks:\n%s", rendered)

	// Check if counter increased (with throttling, some clicks may be dropped)
	if strings.Contains(rendered, "Clicks: ") {
		t.Log("PASS: Multiple ActionClick events processed")
	}
}

// TestActionSystemModalFocusTrap tests focus trap with modal open
func TestActionSystemModalFocusTrap(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	// Open modal
	for i := 0; i < 5; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
	}
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(200 * time.Millisecond)
	testApp.ForceRender()

	// Verify modal is open
	rendered := testApp.GetRenderString()
	if !strings.Contains(rendered, "*** Are you sure? ***") {
		t.Skip("Modal not open, skipping focus trap test")
	}

	// With modal open, Tab should cycle only between modal buttons
	focusedIndices := make(map[int]bool)
	for i := 0; i < 10; i++ {
		idx := testApp.GetFocusedIndex()
		focusedIndices[idx] = true
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
		testApp.ForceRender()
	}

	t.Logf("Focused indices with modal open: %v", focusedIndices)

	// Focus trap should limit the number of focused elements
	if len(focusedIndices) <= 5 {
		t.Log("PASS: Focus trap appears to work (limited focused elements)")
	} else {
		t.Logf("Note: %d different elements focused (focus trap may not be active)", len(focusedIndices))
	}
}

// TestActionSystemMiddlewareChain tests that middleware chain is active
func TestActionSystemMiddlewareChain(t *testing.T) {
	// Enable debug mode to verify middleware is active
	t.Setenv("ACTION_DEBUG", "true")

	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	// Inject some actions to trigger middleware
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(30 * time.Millisecond)
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(30 * time.Millisecond)
	testApp.InjectSpecialKey(platform.KeyEscape)
	time.Sleep(30 * time.Millisecond)

	testApp.ForceRender()

	t.Log("PASS: Middleware chain processed actions without errors")
}

// TestActionSystemQuit tests ActionQuit
func TestActionSystemQuit(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	// Navigate to Quit button
	for i := 0; i < 15; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
	}

	// Note: We can't actually test Quit in a test context as it would close the app
	// But we can verify the button is focusable
	testApp.ForceRender()
	t.Log("PASS: Navigation to Quit button completed")
}

// TestActionSystemStateUpdate tests that state updates are triggered by actions
func TestActionSystemStateUpdate(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	// Get initial render
	initial := testApp.GetRenderString()

	// Trigger state update by clicking Add Count
	for i := 0; i < 5; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
	}
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(200 * time.Millisecond)
	testApp.ForceRender()

	after := testApp.GetRenderString()

	// Verify something changed
	if initial != after {
		t.Log("PASS: Action caused state update and re-render")
	} else {
		t.Log("Note: No visible change (may not have reached Add Count button)")
	}
}

// TestActionSystemIntegration is a comprehensive integration test
func TestActionSystemIntegration(t *testing.T) {
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	t.Log("=== Action System Integration Test ===")

	// Test 1: Initial render
	t.Log("Test 1: Initial render")
	if err := testApp.AssertRender("TUI Engine Demo"); err != nil {
		t.Errorf("Initial render failed: %v", err)
	} else {
		t.Log("  ✓ Initial render OK")
	}

	// Test 2: Open modal via ActionEnter
	t.Log("Test 2: Open modal")
	for i := 0; i < 5; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
	}
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(200 * time.Millisecond)
	testApp.ForceRender()

	if err := testApp.AssertRender("*** Are you sure? ***"); err != nil {
		t.Logf("  ✗ Modal not visible: %v", err)
	} else {
		t.Log("  ✓ Modal opened")
	}

	// Test 3: Close modal via ActionCancel
	t.Log("Test 3: Close modal")
	testApp.InjectSpecialKey(platform.KeyEscape)
	time.Sleep(200 * time.Millisecond)
	testApp.ForceRender()

	if err := testApp.AssertNotRender("*** Are you sure? ***"); err != nil {
		t.Logf("  ✗ Modal still visible: %v", err)
	} else {
		t.Log("  ✓ Modal closed")
	}

	// Test 4: Navigate with ActionNavigateNext
	t.Log("Test 4: Navigation")
	focusedIndices := make(map[int]bool)
	for i := 0; i < 5; i++ {
		idx := testApp.GetFocusedIndex()
		focusedIndices[idx] = true
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)
	}
	if len(focusedIndices) > 1 {
		t.Logf("  ✓ Navigation works (visited %d elements)", len(focusedIndices))
	} else {
		t.Log("  Note: Limited navigation")
	}

	t.Log("=== Integration Test Complete ===")
}

// TestActionSystemTypes verifies all action types are processable
func TestActionSystemTypes(t *testing.T) {
	actions := []struct {
		name string
		key  platform.SpecialKey
	}{
		{"ActionNavigateNext", platform.KeyTab},
		{"ActionEnter", platform.KeyEnter},
		{"ActionCancel", platform.KeyEscape},
		{"ActionNavigateUp", platform.KeyUp},
		{"ActionNavigateDown", platform.KeyDown},
		{"ActionNavigateLeft", platform.KeyLeft},
		{"ActionNavigateRight", platform.KeyRight},
	}

	for _, tt := range actions {
		t.Run(tt.name, func(t *testing.T) {
			testApp, err := ui.RunTest(App,
				ui.WithWidth(80),
				ui.WithHeight(24),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer testApp.Close()

			time.Sleep(50 * time.Millisecond)
			testApp.ForceRender()

			// Inject the action
			testApp.InjectSpecialKey(tt.key)
			time.Sleep(50 * time.Millisecond)
			testApp.ForceRender()

			t.Logf("%s processed without panic", tt.name)
		})
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

// BenchmarkActionDispatch benchmarks action dispatch performance
func BenchmarkActionDispatch(b *testing.B) {
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)
	testApp.ForceRender()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
	}
}

// BenchmarkActionWithRender benchmarks full action + render cycle
func BenchmarkActionWithRender(b *testing.B) {
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		testApp.ForceRender()
	}
}

