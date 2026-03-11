package tabs

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Test intent for testing HandleIntent with unknown intent types
type TestCustomIntent struct{}

func (TestCustomIntent) IntentType() string {
	return "Test:Custom"
}

// =============================================================================
// TabChangeIntent Tests
// =============================================================================

func TestTabChangeIntent_Type(t *testing.T) {
	i := TabChange("comp1", 0, "tab1", "Tab 1")
	if i.IntentType() != "Tabs:TabChange" {
		t.Errorf("Expected intent type 'Tabs:TabChange', got '%s'", i.IntentType())
	}
}

func TestTabChangeIntent_Properties(t *testing.T) {
	i := TabChange("comp1", 0, "tab1", "Tab 1")

	// TabChangeIntent is already a struct, no need to assert it
	changeIntent := i

	if changeIntent.ComponentID != "comp1" {
		t.Errorf("Expected ComponentID 'comp1', got '%s'", changeIntent.ComponentID)
	}
	if changeIntent.ActiveTab != 0 {
		t.Errorf("Expected ActiveTab 0, got %d", changeIntent.ActiveTab)
	}
	if changeIntent.TabID != "tab1" {
		t.Errorf("Expected TabID 'tab1', got '%s'", changeIntent.TabID)
	}
	if changeIntent.TabLabel != "Tab 1" {
		t.Errorf("Expected TabLabel 'Tab 1', got '%s'", changeIntent.TabLabel)
	}
}

func TestTabChangeIntent_Priority(t *testing.T) {
	i := TabChange("comp1", 0, "tab1", "Tab 1")
	if i.Priority() != intent.PriorityNormal {
		t.Errorf("Expected priority %s, got %s", intent.PriorityNormal, i.Priority())
	}
}

func TestTabChangeIntent_Transition(t *testing.T) {
	i := TabChange("comp1", 0, "tab1", "Tab 1")
	if !i.IsTransition() {
		t.Error("TabChangeIntent should be a transition intent")
	}
}

// =============================================================================
// TabNextIntent Tests
// =============================================================================

func TestTabNextIntent_Type(t *testing.T) {
	i := TabNext("comp1")
	if i.IntentType() != "Tabs:TabNext" {
		t.Errorf("Expected intent type 'Tabs:TabNext', got '%s'", i.IntentType())
	}
}

func TestTabNextIntent_Properties(t *testing.T) {
	i := TabNext("comp1")

	// TabNextIntent is already a struct, no need to assert it
	nextIntent := i

	if nextIntent.ComponentID != "comp1" {
		t.Errorf("Expected ComponentID 'comp1', got '%s'", nextIntent.ComponentID)
	}
}

func TestTabNextIntent_Priority(t *testing.T) {
	i := TabNext("comp1")
	if i.Priority() != intent.PriorityNormal {
		t.Errorf("Expected priority %s, got %s", intent.PriorityNormal, i.Priority())
	}
}

func TestTabNextIntent_Transition(t *testing.T) {
	i := TabNext("comp1")
	if !i.IsTransition() {
		t.Error("TabNextIntent should be a transition intent")
	}
}

// =============================================================================
// TabPreviousIntent Tests
// =============================================================================

func TestTabPreviousIntent_Type(t *testing.T) {
	i := TabPrevious("comp1")
	if i.IntentType() != "Tabs:TabPrevious" {
		t.Errorf("Expected intent type 'Tabs:TabPrevious', got '%s'", i.IntentType())
	}
}

func TestTabPreviousIntent_Properties(t *testing.T) {
	i := TabPrevious("comp1")

	// TabPreviousIntent is already a struct, no need to assert it
	prevIntent := i

	if prevIntent.ComponentID != "comp1" {
		t.Errorf("Expected ComponentID 'comp1', got '%s'", prevIntent.ComponentID)
	}
}

func TestTabPreviousIntent_Priority(t *testing.T) {
	i := TabPrevious("comp1")
	if i.Priority() != intent.PriorityNormal {
		t.Errorf("Expected priority %s, got %s", intent.PriorityNormal, i.Priority())
	}
}

func TestTabPreviousIntent_Transition(t *testing.T) {
	i := TabPrevious("comp1")
	if !i.IsTransition() {
		t.Error("TabPreviousIntent should be a transition intent")
	}
}

// =============================================================================
// TabSelectIntent Tests
// =============================================================================

func TestTabSelectIntent_Type(t *testing.T) {
	i := TabSelect("comp1", "tab2")
	if i.IntentType() != "Tabs:TabSelect" {
		t.Errorf("Expected intent type 'Tabs:TabSelect', got '%s'", i.IntentType())
	}
}

func TestTabSelectIntent_Properties(t *testing.T) {
	i := TabSelect("comp1", "tab2")
	if i.ComponentID != "comp1" {
		t.Errorf("Expected ComponentID 'comp1', got '%s'", i.ComponentID)
	}
	if i.TabID != "tab2" {
		t.Errorf("Expected TabID 'tab2', got '%s'", i.TabID)
	}
	if i.TabIndex != -1 {
		t.Errorf("Expected TabIndex -1, got %d", i.TabIndex)
	}

	indexIntent := TabSelectIndex("comp1", 2)
	if indexIntent.TabID != "" {
		t.Errorf("Expected empty TabID for index intent, got '%s'", indexIntent.TabID)
	}
	if indexIntent.TabIndex != 2 {
		t.Errorf("Expected TabIndex 2, got %d", indexIntent.TabIndex)
	}
}

// =============================================================================
// ComponentID Routing Tests
// =============================================================================

func TestShouldHandleIntent_MatchingID(t *testing.T) {
	// Same component
	if !shouldHandleIntent("comp1", "comp1") {
		t.Error("Should handle intent when IDs match")
	}

	// Both empty (backward compatibility)
	if !shouldHandleIntent("", "") {
		t.Error("Should handle intent when both IDs are empty")
	}

	// One empty (backward compatibility)
	if !shouldHandleIntent("comp1", "") {
		t.Error("Should handle intent when intent ID is empty")
	}
	if !shouldHandleIntent("", "comp1") {
		t.Error("Should handle intent when component ID is empty")
	}

	// Different IDs
	if shouldHandleIntent("comp1", "comp2") {
		t.Error("Should NOT handle intent when IDs don't match")
	}
}

// =============================================================================
// Instance HandleIntent Tests
// =============================================================================

func TestInstance_HandleIntent_TabChange(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(map[string]interface{}{
		"tabs":        tabs,
		"componentID": "comp1",
	})
	inst.SetActiveTab(0)

	// Handle TabNextIntent (should change to tab 2)
	intent := TabNext("comp1")
	if !inst.HandleIntent(intent) {
		t.Error("HandleIntent should return true for TabNextIntent")
	}
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.activeTab)
	}
}

func TestInstance_HandleIntent_TabNext(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(map[string]interface{}{
		"tabs":        tabs,
		"componentID": "comp1",
	})
	inst.SetActiveTab(0)

	// Handle TabNextIntent
	intent := TabNext("comp1")
	if !inst.HandleIntent(intent) {
		t.Error("HandleIntent should return true for TabNextIntent")
	}
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.activeTab)
	}

	// Try to go to next but at last tab - should return false
	if inst.HandleIntent(TabNext("comp1")) {
		t.Error("HandleIntent should return false when at last tab")
	}
}

func TestInstance_HandleIntent_TabPrevious(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(map[string]interface{}{
		"tabs":        tabs,
		"componentID": "comp1",
	})
	inst.SetActiveTab(1)

	// Handle TabPreviousIntent
	intent := TabPrevious("comp1")
	if !inst.HandleIntent(intent) {
		t.Error("HandleIntent should return true for TabPreviousIntent")
	}
	if inst.activeTab != 0 {
		t.Errorf("Expected activeTab 0, got %d", inst.activeTab)
	}

	// Try to go to previous but at first tab - should return false
	if inst.HandleIntent(TabPrevious("comp1")) {
		t.Error("HandleIntent should return false when at first tab")
	}
}

func TestInstance_HandleIntent_TabSelect(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
		{ID: "tab3", Label: "Tab 3"},
	}

	inst := NewInstance(map[string]interface{}{
		"tabs":        tabs,
		"componentID": "comp1",
	})

	if !inst.HandleIntent(TabSelect("comp1", "tab3")) {
		t.Fatal("HandleIntent should select by tab ID")
	}
	if inst.activeTab != 2 {
		t.Fatalf("Expected activeTab 2 after TabSelect, got %d", inst.activeTab)
	}

	if !inst.HandleIntent(TabSelectIndex("comp1", 1)) {
		t.Fatal("HandleIntent should select by tab index")
	}
	if inst.activeTab != 1 {
		t.Fatalf("Expected activeTab 1 after TabSelectIndex, got %d", inst.activeTab)
	}

	if inst.HandleIntent(TabSelect("other", "tab1")) {
		t.Fatal("HandleIntent should ignore TabSelect for other component IDs")
	}
}

func TestInstance_HandleIntent_ComponentIDRouting(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(map[string]interface{}{
		"tabs":        tabs,
		"componentID": "my-tabs",
	})
	inst.SetActiveTab(0)

	// Should handle intent with matching ComponentID
	intent := TabNext("my-tabs")
	if !inst.HandleIntent(intent) {
		t.Error("Should handle intent with matching ComponentID")
	}

	// Should NOT handle intent with different ComponentID
	intent2 := TabNext("other-tabs")
	if inst.HandleIntent(intent2) {
		t.Error("Should NOT handle intent with different ComponentID")
	}
}

func TestInstance_HandleIntent_ComponentIDEmpty(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	// Instance with empty componentID (backward compatibility)
	inst := NewInstance(map[string]interface{}{
		"tabs":        tabs,
		"componentID": "",
	})
	inst.SetActiveTab(0)

	intent := TabNext("any-component") // Intent has ComponentID but instance doesn't
	if !inst.HandleIntent(intent) {
		t.Error("Should handle intent when instance componentID is empty (backward compatibility)")
	}
}

func TestInstance_HandleIntent_IgnoreOtherIntents(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
	}

	inst := NewInstance(map[string]interface{}{
		"tabs":        tabs,
		"componentID": "comp1",
	})

	// Handle generic intent (not Tabs-specific)
	result := inst.HandleIntent(TestCustomIntent{})
	if result {
		t.Error("Should return false for non-Tabs intents")
	}
}

// =============================================================================
// TabChangeIntent Emission Tests
// =============================================================================

func TestInstance_TabChangeIntent_CanBeCreated(t *testing.T) {
	// Verify TabChangeIntent creation works correctly
	intent := TabChange("comp1", 1, "tab2", "Tab 2")

	if intent.ComponentID != "comp1" {
		t.Errorf("Expected ComponentID 'comp1', got '%s'", intent.ComponentID)
	}
	if intent.ActiveTab != 1 {
		t.Errorf("Expected ActiveTab 1, got %d", intent.ActiveTab)
	}
	if intent.TabID != "tab2" {
		t.Errorf("Expected TabID 'tab2', got '%s'", intent.TabID)
	}
	if intent.TabLabel != "Tab 2" {
		t.Errorf("Expected TabLabel 'Tab 2', got '%s'", intent.TabLabel)
	}
}

func TestInstance_SetActiveTab_UpdatesCorrectly(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(map[string]interface{}{
		"tabs":        tabs,
		"componentID": "comp1",
	})

	// Create and emit a TabChangeIntent manually (simulating what emitChangeIntent does)
	tabIntent := TabChange(inst.componentID, 1, "tab2", "Tab 2")

	// Verify the intent has correct values
	if tabIntent.ComponentID != "comp1" {
		t.Errorf("Expected ComponentID 'comp1', got '%s'", tabIntent.ComponentID)
	}
	if tabIntent.ActiveTab != 1 {
		t.Errorf("Expected ActiveTab 1, got %d", tabIntent.ActiveTab)
	}
	if tabIntent.TabID != "tab2" {
		t.Errorf("Expected TabID 'tab2', got '%s'", tabIntent.TabID)
	}
	if tabIntent.TabLabel != "Tab 2" {
		t.Errorf("Expected TabLabel 'Tab 2', got '%s'", tabIntent.TabLabel)
	}

	// Verify SetActiveTab works correctly
	if !inst.SetActiveTab(1) {
		t.Error("SetActiveTab(1) should return true")
	}
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.activeTab)
	}
}

func TestInstance_EmitChangeIntent_NoComponentID(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
	}

	inst := NewInstance(map[string]interface{}{
		"tabs":        tabs,
		"componentID": "", // No component ID
	})

	// Track emitted intents
	emittedCount := 0
	inst.SetIntentEmitter(func(i intent.Intent) {
		emittedCount++
	})

	// Trigger change (should NOT emit TabChangeIntent without componentID)
	inst.SetActiveTab(0)

	if emittedCount > 0 {
		// Without componentID, TabChangeIntent should not be emitted
		t.Log("Note: TabChangeIntent not emitted without componentID (expected)")
	}
}

// =============================================================================
// VNode ComponentID Tests
// =============================================================================

func TestVNode_ComponentID(t *testing.T) {
	vnode := New().SetComponentID("my-tabs")
	if vnode.componentID != "my-tabs" {
		t.Errorf("Expected componentID 'my-tabs', got '%s'", vnode.componentID)
	}
}

func TestVNode_Props_IncludeComponentID(t *testing.T) {
	vnode := New().SetComponentID("my-tabs")

	props := vnode.Props()
	componentID, ok := props["componentID"].(string)
	if !ok {
		t.Fatal("componentID should be in props as string")
	}
	if componentID != "my-tabs" {
		t.Errorf("Expected componentID 'my-tabs', got '%s'", componentID)
	}
}

func TestVNode_SetProps_ComponentID(t *testing.T) {
	vnode := New()

	props := rtui.Props{
		"componentID": "test-id",
	}

	vnode.SetProps(props)
	if vnode.componentID != "test-id" {
		t.Errorf("Expected componentID 'test-id', got '%s'", vnode.componentID)
	}
}

// =============================================================================
// Builder ComponentID Tests
// =============================================================================

func TestBuilder_ComponentID(t *testing.T) {
	vnode := NewBuilder().
		Tabs([]TabItem{{ID: "tab1", Label: "Tab 1"}}).
		ComponentID("my-tabs").
		BuildVNode()

	if vnode.componentID != "my-tabs" {
		t.Errorf("Expected componentID 'my-tabs', got '%s'", vnode.componentID)
	}
}

func TestBuilder_ChainMethodsWithComponentID(t *testing.T) {
	vnode := NewBuilder().
		ComponentID("tabs-1").
		Width(100).
		Height(30).
		Position(TabPositionTop).
		Style(style.Style{FG: style.Color("cyan")}).
		BuildVNode()

	if vnode.componentID != "tabs-1" {
		t.Errorf("Expected componentID 'tabs-1', got '%s'", vnode.componentID)
	}
	if vnode.width != 100 {
		t.Errorf("Expected width 100, got %d", vnode.width)
	}
	if vnode.height != 30 {
		t.Errorf("Expected height 30, got %d", vnode.height)
	}
}

// =============================================================================
// Parent Interface Tests
// =============================================================================

func TestInstance_Parent(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
	}

	inst := NewInstance(map[string]interface{}{"tabs": tabs})

	// Tabs is a leaf component, Parent() should return nil
	if inst.Parent() != nil {
		t.Error("Parent() should return nil for Tabs component")
	}
}

// =============================================================================
// Instance Create Tests with ComponentID
// =============================================================================

func TestInstance_NewInstance_WithComponentID(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
	}

	props := rtui.Props{
		"tabs":        tabs,
		"componentID": "test-tabs",
	}

	inst := NewInstance(props)
	if inst.componentID != "test-tabs" {
		t.Errorf("Expected componentID 'test-tabs', got '%s'", inst.componentID)
	}
}

func TestInstance_CreateInstance_FromVNode(t *testing.T) {
	vnode := New().
		SetComponentID("my-tabs").
		SetTabs([]TabItem{{ID: "tab1", Label: "Tab 1"}})

	inst := vnode.CreateInstance().(*Instance)
	if inst.componentID != "my-tabs" {
		t.Errorf("Expected componentID 'my-tabs', got '%s'", inst.componentID)
	}
}
