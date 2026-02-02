package ui

import (
	"testing"
)

func TestPropsBasicOperations(t *testing.T) {
	p := make(Props)

	// Test Set and Get
	p.Set("key1", "value1")
	if val := p.Get("key1"); val != "value1" {
		t.Errorf("Get() = %v, want %v", val, "value1")
	}

	// Test GetString
	if val := p.GetString("key1"); val != "value1" {
		t.Errorf("GetString() = %v, want %v", val, "value1")
	}

	// Test GetInt
	p.Set("number", 42)
	if val := p.GetInt("number"); val != 42 {
		t.Errorf("GetInt() = %v, want %v", val, 42)
	}

	// Test GetBool
	p.Set("flag", true)
	if val := p.GetBool("flag"); val != true {
		t.Errorf("GetBool() = %v, want %v", val, true)
	}

	// Test missing key
	if val := p.Get("missing"); val != nil {
		t.Errorf("Get() missing key = %v, want nil", val)
	}
}

func TestPropsMerge(t *testing.T) {
	p1 := make(Props)
	p1.Set("key1", "value1")

	p2 := make(Props)
	p2.Set("key2", "value2")

	merged := p1.Merge(p2)

	if val := merged.Get("key1"); val != "value1" {
		t.Errorf("Merge() key1 = %v, want %v", val, "value1")
	}
	if val := merged.Get("key2"); val != "value2" {
		t.Errorf("Merge() key2 = %v, want %v", val, "value2")
	}
}

func TestPropsClone(t *testing.T) {
	p1 := make(Props)
	p1.Set("key1", "value1")

	p2 := p1.Clone()
	p2.Set("key1", "value2")

	// Original should be unchanged
	if val := p1.Get("key1"); val != "value1" {
		t.Errorf("Clone() original changed = %v, want %v", val, "value1")
	}

	// Clone should have new value
	if val := p2.Get("key1"); val != "value2" {
		t.Errorf("Clone() copy = %v, want %v", val, "value2")
	}
}

func TestVNodeType(t *testing.T) {
	tests := []struct {
		name string
		typ  VNodeType
	}{
		{"VNodeComponent", VNodeComponent},
		{"VNodeText", VNodeText},
		{"VNodeElement", VNodeElement},
		{"VNodeFragment", VNodeFragment},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.typ.String() == "" {
				t.Errorf("VNodeType.String() returned empty for %v", tt.typ)
			}
		})
	}
}

func TestComponentContextBasics(t *testing.T) {
	ctx := NewComponentContext("test-comp")

	if ctx.ComponentID == "" {
		t.Error("ComponentID should not be empty")
	}

	// Test hook creation
	hook := Hook{
		Type:    HookState,
		Value:   42,
	}
	ctx.Hooks = append(ctx.Hooks, hook)

	// Test ResetContext
	ctx.ResetContext()
	if ctx.HookIndex != 0 {
		t.Errorf("ResetContext() HookIndex = %v, want 0", ctx.HookIndex)
	}

	// Test GetOrCreateHook
	retrieved := ctx.GetOrCreateHook(HookState)
	if retrieved.Type != HookState {
		t.Errorf("GetOrCreateHook() Type = %v, want %v", retrieved.Type, HookState)
	}
	if retrieved.Value != 42 {
		t.Errorf("GetOrCreateHook() Value = %v, want 42", retrieved.Value)
	}
}

func TestComponentContextFinishRender(t *testing.T) {
	ctx := NewComponentContext("test-comp")

	// First render with no hooks - should succeed
	if err := ctx.FinishRender(); err != nil {
		t.Errorf("FinishRender() first render with no hooks returned error: %v", err)
	}

	// Reset and render again with no hooks - should succeed
	ctx.ResetContext()
	if err := ctx.FinishRender(); err != nil {
		t.Errorf("FinishRender() second render with no hooks returned error: %v", err)
	}

	// Test that validator tracks hooks correctly
	ctx2 := NewComponentContext("test-comp-2")
	// First render with a hook
	ctx2.Hooks = append(ctx2.Hooks, Hook{Type: HookState})
	ctx2.HookIndex = 1
	if err := ctx2.FinishRender(); err != nil {
		t.Errorf("FinishRender() with hooks returned error: %v", err)
	}

	// Verify validator state changed
	if ctx2.Validator.isFirstRender {
		t.Error("Validator should not be first render after FinishRender")
	}
}

func TestLaneOperations(t *testing.T) {
	// Test MergeLanes
	lane1 := Lane(1)
	lane2 := Lane(2)
	merged := MergeLanes(lane1, lane2)

	if merged != Lane(3) {
		t.Errorf("MergeLanes() = %v, want %v", merged, Lane(3))
	}

	// Test NoLane
	if LaneNoLane != Lane(0) {
		t.Errorf("LaneNoLane = %v, want %v", LaneNoLane, Lane(0))
	}
}

func TestFiberFlags(t *testing.T) {
	fiber := &Fiber{}

	// Initially no flags
	if fiber.HasNoPendingWork() == false {
		t.Error("New fiber should have no pending work")
	}

	// Mark for update
	fiber.MarkUpdate(LaneSyncLane)
	if !fiber.HasEffect() {
		t.Error("MarkUpdate() should set effect flag")
	}

	// Test SubtreeFlags (manually set for testing)
	fiber.SubtreeFlags = EffectUpdate
	if !fiber.HasSubtreeEffect() {
		t.Error("HasSubtreeEffect() should return true when SubtreeFlags is set")
	}
}

func TestCloneFiber(t *testing.T) {
	// Create a simple fiber tree
	child := &Fiber{
		Key:  "child",
		Type: VNodeText,
	}
	parent := &Fiber{
		Key:   "parent",
		Type:  VNodeElement,
		Child: child,
	}

	// Clone
	cloned := CloneFiber(parent)

	if cloned.Key != parent.Key {
		t.Errorf("CloneFiber() Key = %v, want %v", cloned.Key, parent.Key)
	}
	if cloned.Type != parent.Type {
		t.Errorf("CloneFiber() Type = %v, want %v", cloned.Type, parent.Type)
	}

	// CloneFiber copies Child and Sibling (shallow copy of the tree structure)
	if cloned.Child == nil {
		t.Error("CloneFiber() should copy Child pointer")
	}
	// But it should be the same child reference (not a deep clone)
	if cloned.Child != child {
		t.Error("CloneFiber() Child should be same reference (shallow copy)")
	}
}

func TestCountFibers(t *testing.T) {
	// Create a simple fiber tree
	leaf1 := &Fiber{Key: "leaf1", Type: VNodeText}
	leaf2 := &Fiber{Key: "leaf2", Type: VNodeText}
	child := &Fiber{Key: "child", Type: VNodeElement, Child: leaf1}
	root := &Fiber{Key: "root", Type: VNodeElement, Child: child}

	// Link siblings
	leaf1.Sibling = leaf2

	count := CountFibers(root)
	if count != 4 {
		t.Errorf("CountFibers() = %v, want 4", count)
	}

	// Test nil fiber
	if CountFibers(nil) != 0 {
		t.Error("CountFibers(nil) should return 0")
	}
}

func TestElementBuilder(t *testing.T) {
	// Test basic element creation
	elem := Element("div").
		Prop("id", "test").
		Prop("class", "container").
		Build()

	if elem.Type() != VNodeElement {
		t.Errorf("Element() Type = %v, want %v", elem.Type(), VNodeElement)
	}

	props := elem.Props()
	if props.GetString("id") != "test" {
		t.Errorf("Element() Prop id = %v, want %v", props.GetString("id"), "test")
	}
	if props.GetString("class") != "container" {
		t.Errorf("Element() Prop class = %v, want %v", props.GetString("class"), "container")
	}
}

func TestComponentBuilder(t *testing.T) {
	fn := func() VNode {
		return Element("text").Prop("content", "test").Build()
	}

	// Test component creation
	comp := Component("TestComponent", fn).
		Key("test-key").
		Build()

	if comp.Type() != VNodeComponent {
		t.Errorf("Component() Type = %v, want %v", comp.Type(), VNodeComponent)
	}

	if comp.Key() != "test-key" {
		t.Errorf("Component() Key = %v, want %v", comp.Key(), "test-key")
	}
}

func TestFragment(t *testing.T) {
	children := []VNode{
		Element("text").Prop("content", "child1").Build(),
		Element("text").Prop("content", "child2").Build(),
	}

	frag := Fragment(children...)

	if frag.Type() != VNodeFragment {
		t.Errorf("Fragment() Type = %v, want %v", frag.Type(), VNodeFragment)
	}

	fragChildren := frag.Children()
	if len(fragChildren) != 2 {
		t.Errorf("Fragment() Children count = %v, want 2", len(fragChildren))
	}
}

func TestLayoutBuilders(t *testing.T) {
	// Test HStack
	hstack := HStack(
		Element("text").Prop("content", "a").Build(),
		Element("text").Prop("content", "b").Build(),
	)
	if hstack.Type() != VNodeElement {
		t.Error("HStack() should return element")
	}

	// Test VStack
	vstack := VStack(
		Element("text").Prop("content", "a").Build(),
		Element("text").Prop("content", "b").Build(),
	)
	if vstack.Type() != VNodeElement {
		t.Error("VStack() should return element")
	}

	// Test Box builder
	box := Box()
	if box == nil {
		t.Error("Box() should return builder")
	}

	// Test Spacer
	spacer := Spacer()
	if spacer == nil {
		t.Error("Spacer() should return builder")
	}
}

func TestTextElement(t *testing.T) {
	content := "Hello, World!"
	elem := Element("text").Prop("content", content).Build()

	if elem.Props().GetString("content") != content {
		t.Errorf("Text element content = %v, want %v", elem.Props().GetString("content"), content)
	}
}
