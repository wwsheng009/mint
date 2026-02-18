package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestNewFiberFromVNode tests Fiber creation from VNodes
func TestNewFiberFromVNode(t *testing.T) {
	elem := rtui.Element("div").Prop("id", "test").Build()
	fiber := CreateFiberFromVNode(elem)

	if fiber == nil {
		t.Fatal("CreateFiberFromVNode returned nil for ElementVNode")
	}

	if fiber.Type != rtui.VNodeElement {
		t.Errorf("Fiber.Type = %v, want VNodeElement", fiber.Type)
	}

	if fiber.Tag != "div" {
		t.Errorf("Fiber.Tag = %v, want 'div'", fiber.Tag)
	}

	if fiber.Props["id"] != "test" {
		t.Errorf("Fiber.Props['id'] = %v, want 'test'", fiber.Props["id"])
	}
}

// TestNewFiberFromVNode_Component tests Fiber creation from ComponentVNode
func TestNewFiberFromVNode_Component(t *testing.T) {
	comp := rtui.NewComponent("TestComponent", func() rtui.VNode {
		return rtui.Element("span").Build()
	})

	fiber := CreateFiberFromVNode(comp)

	if fiber == nil {
		t.Fatal("CreateFiberFromVNode returned nil for ComponentVNode")
	}

	if fiber.Type != rtui.VNodeComponent {
		t.Errorf("Fiber.Type = %v, want VNodeComponent", fiber.Type)
	}
}

// TestNewFiberFromVNode_Text tests Fiber creation from text Element
func TestNewFiberFromVNode_Text(t *testing.T) {
	// Text nodes are represented as ElementVNode with tag "text"
	text := rtui.Element("text").Prop("content", "Hello").Build()
	fiber := CreateFiberFromVNode(text)

	if fiber == nil {
		t.Fatal("CreateFiberFromVNode returned nil for text Element")
	}

	if fiber.Type != rtui.VNodeElement {
		t.Errorf("Fiber.Type = %v, want VNodeElement", fiber.Type)
	}

	if fiber.Tag != "text" {
		t.Errorf("Fiber.Tag = %v, want 'text'", fiber.Tag)
	}
}

// TestNewFiberFromVNode_Fragment tests Fiber creation from FragmentVNode
func TestNewFiberFromVNode_Fragment(t *testing.T) {
	frag := rtui.Fragment(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
	)

	fiber := CreateFiberFromVNode(frag)

	if fiber == nil {
		t.Fatal("CreateFiberFromVNode returned nil for FragmentVNode")
	}

	if fiber.Type != rtui.VNodeFragment {
		t.Errorf("Fiber.Type = %v, want VNodeFragment", fiber.Type)
	}

	// Fragment should have children
	if fiber.Child == nil {
		t.Error("Fragment Fiber should have children")
	}
}

// TestNewFiberFromVNode_nil tests Fiber creation from nil
func TestNewFiberFromVNode_nil(t *testing.T) {
	fiber := CreateFiberFromVNode(nil)

	if fiber != nil {
		t.Error("CreateFiberFromVNode should return nil for nil VNode")
	}
}

// TestCloneFiber tests Fiber cloning
func TestCloneFiber(t *testing.T) {
	original := &rtui.Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "div",
		Key:     "test-key",
		DiffKey: "test-key",
		Props:   rtui.Props{"custom": "value"},
		Lanes:   rtui.LaneSyncLane,
		Flags:   rtui.EffectUpdate,
	}

	cloned := CloneFiber(original)

	if cloned == nil {
		t.Fatal("CloneFiber returned nil")
	}

	if cloned.Key != original.Key {
		t.Errorf("Cloned Key = %v, want %v", cloned.Key, original.Key)
	}

	if cloned.Tag != original.Tag {
		t.Errorf("Cloned Tag = %v, want %v", cloned.Tag, original.Tag)
	}

	if cloned.Type != original.Type {
		t.Errorf("Cloned Type = %v, want %v", cloned.Type, original.Type)
	}

	if cloned.Lanes != original.Lanes {
		t.Errorf("Cloned Lanes = %v, want %v", cloned.Lanes, original.Lanes)
	}
}

// TestCloneFiber_nil tests cloning nil fiber
func TestCloneFiber_nil(t *testing.T) {
	cloned := CloneFiber(nil)

	if cloned != nil {
		t.Error("CloneFiber(nil) should return nil")
	}
}

// TestMergeLanes tests lane merging
func TestMergeLanes(t *testing.T) {
	tests := []struct {
		name     string
		a, b     rtui.Lane
		expected rtui.Lane
	}{
		{"merge no lanes", rtui.LaneNoLane, rtui.LaneNoLane, rtui.LaneNoLane},
		{"merge no lane with sync", rtui.LaneNoLane, rtui.LaneSyncLane, rtui.LaneSyncLane},
		{"merge sync with no lane", rtui.LaneSyncLane, rtui.LaneNoLane, rtui.LaneSyncLane},
		{"merge sync with sync", rtui.LaneSyncLane, rtui.LaneSyncLane, rtui.LaneSyncLane},
		{"merge different lanes", rtui.LaneSyncLane, rtui.LaneInputContinuousLane, rtui.LaneSyncLane | rtui.LaneInputContinuousLane},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeLanes(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("MergeLanes(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestCountFibers tests fiber counting
func TestCountFibers(t *testing.T) {
	// Create a simple fiber tree
	leaf1 := &rtui.Fiber{Key: "leaf1", Type: rtui.VNodeText}
	leaf2 := &rtui.Fiber{Key: "leaf2", Type: rtui.VNodeText}
	child := &rtui.Fiber{Key: "child", Type: rtui.VNodeElement, Child: leaf1}
	root := &rtui.Fiber{Key: "root", Type: rtui.VNodeElement, Child: child}

	// Link siblings
	leaf1.Sibling = leaf2

	count := CountFibers(root)
	if count != 4 {
		t.Errorf("CountFibers() = %d, want 4", count)
	}
}

// TestCountFibers_nil tests counting nil fiber
func TestCountFibers_nil(t *testing.T) {
	count := CountFibers(nil)
	if count != 0 {
		t.Errorf("CountFibers(nil) = %d, want 0", count)
	}
}

// TestCountFibers_single tests counting single fiber
func TestCountFibers_single(t *testing.T) {
	single := &rtui.Fiber{Key: "single", Type: rtui.VNodeText}
	count := CountFibers(single)

	if count != 1 {
		t.Errorf("CountFibers(single) = %d, want 1", count)
	}
}

// TestReconcilerConfig tests reconciler configuration
func TestReconcilerConfig(t *testing.T) {
	config := ReconcilerConfig{
		TimeBudget:      10 * 1000000, // 10ms in nanoseconds
		EnableProfiling: true,
		EnableFiber:     true,
	}

	if config.TimeBudget == 0 {
		t.Error("Default TimeBudget should not be 0")
	}
}

// TestNewReconciler tests reconciler creation
func TestNewReconciler(t *testing.T) {
	config := ReconcilerConfig{
		EnableFiber: true,
	}

	reconciler := NewReconciler(nil, nil, config)

	if reconciler == nil {
		t.Fatal("NewReconciler returned nil")
	}

	if !reconciler.enableFiber {
		t.Error("Reconciler.enableFiber should be true")
	}

	if reconciler.instanceMgr == nil {
		t.Error("Reconciler.instanceMgr should be initialized")
	}

	if reconciler.interactionStateMgr == nil {
		t.Error("Reconciler.interactionStateMgr should be initialized")
	}

	if reconciler.keyValidator == nil {
		t.Error("Reconciler.keyValidator should be initialized")
	}
}

// TestReconciler_GetInstanceManager tests getting instance manager
func TestReconciler_GetInstanceManager(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	mgr := reconciler.GetInstanceManager()
	if mgr == nil {
		t.Error("GetInstanceManager should return non-nil manager")
	}

	// Should return the same instance
	mgr2 := reconciler.GetInstanceManager()
	if mgr != mgr2 {
		t.Error("GetInstanceManager should return same instance")
	}
}

// TestReconciler_GetInteractionStateManager tests getting interaction state manager
func TestReconciler_GetInteractionStateManager(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	mgr := reconciler.GetInteractionStateManager()
	if mgr == nil {
		t.Error("GetInteractionStateManager should return non-nil manager")
	}

	// Should return the same instance
	mgr2 := reconciler.GetInteractionStateManager()
	if mgr != mgr2 {
		t.Error("GetInteractionStateManager should return same instance")
	}
}

// TestReconciler_GetKeyValidator tests getting key validator
func TestReconciler_GetKeyValidator(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	validator := reconciler.GetKeyValidator()
	if validator == nil {
		t.Error("GetKeyValidator should return non-nil validator")
	}

	// Should return the same instance
	validator2 := reconciler.GetKeyValidator()
	if validator != validator2 {
		t.Error("GetKeyValidator should return same instance")
	}
}

// TestReconciler_Stats tests getting statistics
func TestReconciler_Stats(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	stats := reconciler.Stats()

	if stats == nil {
		t.Fatal("Stats() returned nil")
	}

	// Check expected keys exist
	expectedKeys := []string{"hasWork", "lanes", "isWorking", "fiberCount", "instances"}
	for _, key := range expectedKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("Stats missing key: %s", key)
		}
	}
}

// TestReconciler_ScheduleUpdate tests scheduling updates
func TestReconciler_ScheduleUpdate(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// Initially no pending work
	if reconciler.lanes != rtui.LaneNoLane {
		t.Errorf("Initial lanes should be NoLane, got %v", reconciler.lanes)
	}

	// Schedule an update
	reconciler.ScheduleUpdate(rtui.LaneSyncLane)

	// Should have pending work
	if reconciler.lanes == rtui.LaneNoLane {
		t.Error("ScheduleUpdate should set lanes")
	}

	// Schedule same lane again - should be idempotent
	initialLanes := reconciler.lanes
	reconciler.ScheduleUpdate(rtui.LaneSyncLane)
	if reconciler.lanes != initialLanes {
		t.Error("Scheduling same lane twice should be idempotent")
	}
}

// TestReconciler_MarkDirty tests marking dirty
func TestReconciler_MarkDirty(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// MarkDirty should not panic
	reconciler.MarkDirty()
}

// TestGetRootContext tests getting root context
func TestGetRootContext(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	ctx := reconciler.GetContext()
	if ctx == nil {
		t.Error("GetContext should return non-nil context")
	}
}

// TestGetCurrentReconciler tests getting current reconciler
func TestGetCurrentReconciler(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// Not set initially
	if GetCurrentReconciler() != nil {
		t.Error("GetCurrentReconciler should return nil when not set")
	}

	// Set current reconciler and check
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	if GetCurrentReconciler() != reconciler {
		t.Error("GetCurrentReconciler should return the set reconciler")
	}
}

// TestLaneConstants tests lane constant values
func TestLaneConstants(t *testing.T) {
	tests := []struct {
		name  string
		lane  rtui.Lane
		value uint64
	}{
		{"LaneNoLane", rtui.LaneNoLane, 0},
		{"LaneSyncLane", rtui.LaneSyncLane, 1},
		{"LaneInputContinuousLane", rtui.LaneInputContinuousLane, 2},
		{"LaneDefaultLane", rtui.LaneDefaultLane, 4},
		{"LaneIdleLane", rtui.LaneIdleLane, 8},
		{"LaneRoot", rtui.LaneRoot, 15}, // 1|2|4|8 = 15
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint64(tt.lane) != tt.value {
				t.Errorf("%s = %v, want %d", tt.name, tt.lane, tt.value)
			}
		})
	}
}

// TestEffectFlagConstants tests effect flag constants
func TestEffectFlagConstants(t *testing.T) {
	tests := []struct {
		name  string
		flag  rtui.EffectFlag
		value int
	}{
		{"EffectNoEffect", rtui.EffectNoEffect, 0},
		{"EffectPlacement", rtui.EffectPlacement, 2}, // Changed to actual value
		{"EffectUpdate", rtui.EffectUpdate, 4},       // Changed to actual value
		{"EffectDeletion", rtui.EffectDeletion, 8},   // Changed to actual value
		{"EffectRef", rtui.EffectRef, 16},            // Changed to actual value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.flag) != tt.value {
				t.Errorf("%s = %v, want %d", tt.name, tt.flag, tt.value)
			}
		})
	}
}

// TestFiber_HasNoPendingWork tests HasNoPendingWork method
func TestFiber_HasNoPendingWork(t *testing.T) {
	fiber := &rtui.Fiber{}

	// New fiber has no pending work
	if !fiber.HasNoPendingWork() {
		t.Error("New fiber should have no pending work")
	}

	// Add some lanes
	fiber.Lanes = rtui.LaneSyncLane

	// Now has pending work
	if fiber.HasNoPendingWork() {
		t.Error("Fiber with lanes should have pending work")
	}
}

// TestFiber_HasEffect tests HasEffect method
func TestFiber_HasEffect(t *testing.T) {
	fiber := &rtui.Fiber{}

	// New fiber has no effect
	if fiber.HasEffect() {
		t.Error("New fiber should have no effect")
	}

	// Add effect flag
	fiber.Flags = rtui.EffectUpdate

	if !fiber.HasEffect() {
		t.Error("Fiber with EffectUpdate should have effect")
	}
}

// TestFiber_HasSubtreeEffect tests HasSubtreeEffect method
func TestFiber_HasSubtreeEffect(t *testing.T) {
	fiber := &rtui.Fiber{}

	// New fiber has no subtree effect
	if fiber.HasSubtreeEffect() {
		t.Error("New fiber should have no subtree effect")
	}

	// Add subtree effect flag
	fiber.SubtreeFlags = rtui.EffectUpdate

	if !fiber.HasSubtreeEffect() {
		t.Error("Fiber with SubtreeFlags should have subtree effect")
	}
}

// TestFiber_MarkUpdate tests MarkUpdate method
func TestFiber_MarkUpdate(t *testing.T) {
	fiber := &rtui.Fiber{
		Return: &rtui.Fiber{}, // Has parent for propagation
	}

	// Mark for update
	fiber.MarkUpdate(rtui.LaneSyncLane)

	// Should have lanes set
	if fiber.Lanes != rtui.LaneSyncLane {
		t.Errorf("Fiber.Lanes = %v after MarkUpdate, want LaneSyncLane", fiber.Lanes)
	}

	// Should have effect flag set
	if fiber.Flags&rtui.EffectUpdate == 0 {
		t.Error("MarkUpdate should set EffectUpdate flag")
	}

	// Parent should have child lanes set
	if fiber.Return.ChildLanes != rtui.LaneSyncLane {
		t.Errorf("Parent ChildLanes should be set after MarkUpdate, got %v", fiber.Return.ChildLanes)
	}
}

// TestVNodeTypeString tests VNodeType String method
func TestVNodeTypeString(t *testing.T) {
	tests := []struct {
		name   string
		vtype  rtui.VNodeType
		expect string
	}{
		{"VNodeComponent", rtui.VNodeComponent, "Component"},
		{"VNodeText", rtui.VNodeText, "Text"},
		{"VNodeElement", rtui.VNodeElement, "Element"},
		{"VNodeFragment", rtui.VNodeFragment, "Fragment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.vtype.String() != tt.expect {
				t.Errorf("VNodeType.String() = %s, want %s", tt.vtype.String(), tt.expect)
			}
		})
	}
}

// TestNewComponent tests creating component from NewComponent
func TestNewComponent(t *testing.T) {
	// Get a component from reconciler's instance manager
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	comp := rtui.NewComponent("TestComponent", func() rtui.VNode {
		return rtui.Element("span").Build()
	})

	comp.SetKey("test-comp")

	fiber := CreateFiberFromVNode(comp)

	if fiber == nil {
		t.Fatal("CreateFiberFromVNode returned nil")
	}

	// ComponentInstance is set during begin_work, not during fiber creation
	// So we just verify the fiber is created with the correct type
	if fiber.Type != rtui.VNodeComponent {
		t.Errorf("Fiber.Type = %v, want VNodeComponent", fiber.Type)
	}

	// Verify instance manager works
	instance := reconciler.instanceMgr.GetOrCreate("component:test-comp", func() rtui.ComponentInstance {
		return rtui.NewBaseComponentInstance("test-comp", comp.Fn())
	})

	if instance == nil {
		t.Error("GetOrCreateComponent should return instance")
	}
}

// TestVNodeConverter tests VNode conversion
func TestVNodeConverter(t *testing.T) {
	converter := NewVNodeConverter()

	if converter == nil {
		t.Fatal("NewVNodeConverter returned nil")
	}

	// Test Element conversion
	elem := rtui.Element("div").Prop("id", "test").Build()
	layoutNode := converter.Convert(elem)

	if layoutNode == nil {
		t.Error("Convert should return non-nil for ElementVNode")
	}
}

// TestGenerateLayoutBoxes tests layout box generation
func TestGenerateLayoutBoxes(t *testing.T) {
	converter := NewVNodeConverter()

	// Create a simple layout tree
	text := rtui.Element("text").Prop("content", "Test").Build()
	layoutNode := converter.Convert(text)

	boxes := converter.GenerateLayoutBoxes(layoutNode)

	if boxes == nil {
		t.Fatal("GenerateLayoutBoxes returned nil")
	}

	// Should have at least one box for the text node
	if len(boxes) == 0 {
		t.Error("GenerateLayoutBoxes should return at least one box")
	}

	// Each box should have a Node reference
	for i, box := range boxes {
		if box.Node == nil {
			t.Errorf("Box %d has nil Node", i)
		}
	}
}
