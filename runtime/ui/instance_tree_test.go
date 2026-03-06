package ui_test

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/optiongroup"
)

// TestTreeNode verifies TreeNode interface functionality
func TestTreeNode(t *testing.T) {
	t.Run("BaseComponentInstance implements TreeNode", func(t *testing.T) {
		inst := &rtui.BaseComponentInstance{}
		inst.SetKey("root")

		// Verify TreeNode interface is satisfied
		var _ rtui.TreeNode = inst

		// Test Parent returns nil initially
		if inst.Parent() != nil {
			t.Error("Expected Parent() to return nil for root instance")
		}

		// Test Children returns empty initially
		if children := inst.Children(); len(children) != 0 {
			t.Errorf("Expected Children() to return empty slice, got %d items", len(children))
		}
	})
}

// TestTreeContainer verifies TreeContainer interface functionality
func TestTreeContainer(t *testing.T) {
	t.Run("BaseComponentInstance implements TreeContainer", func(t *testing.T) {
		parent := &rtui.BaseComponentInstance{}
		parent.SetKey("parent")
		child1 := &rtui.BaseComponentInstance{}
		child1.SetKey("child1")
		child2 := &rtui.BaseComponentInstance{}
		child2.SetKey("child2")

		// Verify TreeContainer interface is satisfied
		var _ rtui.TreeContainer = parent

		// Verify TreeNode interface is satisfied for children
		var _ rtui.TreeNode = child1
		var _ rtui.TreeNode = child2

		// Test AddChild
		parent.AddChild(child1)
		if children := parent.Children(); len(children) != 1 {
			t.Errorf("Expected 1 child after AddChild, got %d", len(children))
		}

		// Verify parent reference is set
		if child1.Parent() != parent {
			t.Error("Expected child's Parent() to return the parent instance")
		}

		// Test duplicate detection
		parent.AddChild(child1) // Try adding same child again
		if children := parent.Children(); len(children) != 1 {
			t.Errorf("Expected 1 child after duplicate AddChild, got %d", len(children))
		}

		// Test AddChild with nil
		parent.AddChild(nil)
		if children := parent.Children(); len(children) != 1 {
			t.Errorf("Expected still 1 child after nil AddChild, got %d", len(children))
		}

		// Test adding second child
		parent.AddChild(child2)
		if children := parent.Children(); len(children) != 2 {
			t.Errorf("Expected 2 children, got %d", len(children))
		}

		if child2.Parent() != parent {
			t.Error("Expected child2's Parent() to return the parent instance")
		}
	})

	t.Run("RemoveChild", func(t *testing.T) {
		parent := &rtui.BaseComponentInstance{}
		parent.SetKey("parent")
		child1 := &rtui.BaseComponentInstance{}
		child1.SetKey("child1")
		child2 := &rtui.BaseComponentInstance{}
		child2.SetKey("child2")

		parent.AddChild(child1)
		parent.AddChild(child2)

		// Remove child1
		parent.RemoveChild(child1)

		// Verify child1 is removed
		if children := parent.Children(); len(children) != 1 {
			t.Errorf("Expected 1 child after RemoveChild, got %d", len(children))
		}

		// Verify child1's parent is cleared
		if child1.Parent() != nil {
			t.Error("Expected child1's Parent() to be nil after removal")
		}

		// Verify child2 is still present
		if children := parent.Children(); len(children) != 1 || children[0] != child2 {
			t.Error("Expected child2 to still be a child of parent")
		}

		// Test RemoveChild with non-existent instance (should be safe)
		child3 := &rtui.BaseComponentInstance{}
		child3.SetKey("child3")
		parent.RemoveChild(child3)
		if len(parent.Children()) != 1 {
			t.Error("RemoveChild should be idempotent for non-existent children")
		}

		// Test RemoveChild with nil (should be safe)
		parent.RemoveChild(nil)
		if len(parent.Children()) != 1 {
			t.Error("RemoveChild should handle nil gracefully")
		}
	})

	t.Run("ClearChildren", func(t *testing.T) {
		parent := &rtui.BaseComponentInstance{}
		parent.SetKey("parent")
		child1 := &rtui.BaseComponentInstance{}
		child1.SetKey("child1")
		child2 := &rtui.BaseComponentInstance{}
		child2.SetKey("child2")
		child3 := &rtui.BaseComponentInstance{}
		child3.SetKey("child3")

		parent.AddChild(child1)
		parent.AddChild(child2)
		parent.AddChild(child3)

		// Clear all children
		parent.ClearChildren()

		// Verify all children are removed
		if children := parent.Children(); len(children) != 0 {
			t.Errorf("Expected 0 children after ClearChildren, got %d", len(children))
		}

		// Verify all children's parent references are cleared
		if child1.Parent() != nil || child2.Parent() != nil || child3.Parent() != nil {
			t.Error("Expected all children's Parent() to be nil after ClearChildren")
		}
	})
}

// TestOptionGroupInstanceTree verifies OptionGroup/Option Instance Tree integration
func TestOptionGroupInstanceTree(t *testing.T) {
	t.Run("OptionGroup implements TreeContainer", func(t *testing.T) {
		group := &optiongroup.Instance{}

		// Verify TreeContainer interface is satisfied
		var _ rtui.TreeContainer = group
		var _ rtui.TreeNode = group

		// Test initial state
		if group.Parent() != nil {
			t.Error("Expected OptionGroup.Parent() to return nil")
		}

		if children := group.Children(); len(children) != 0 {
			t.Errorf("Expected 0 children initially, got %d", len(children))
		}
	})

	t.Run("Option implements TreeNode", func(t *testing.T) {
		props := rtui.Props{
			"key":   "option1",
			"value": "value1",
			"label": "Label 1",
			"idx":   0,
		}
		option := optiongroup.NewOptionInstance(props)

		// Verify TreeNode interface is satisfied
		var _ rtui.TreeNode = option

		// Option is a leaf component, so Children should return nil
		if children := option.Children(); children != nil {
			t.Error("Expected Option.Children() to return nil for leaf components")
		}

		// Initially, parent should be nil
		if option.Parent() != nil {
			t.Error("Expected Option.Parent() to return nil initially")
		}
	})

	t.Run("OptionGroup can manage Option children", func(t *testing.T) {
		group := &optiongroup.Instance{}

		option1Props := rtui.Props{
			"key":   "option1",
			"value": "value1",
			"label": "Label 1",
			"idx":   0,
		}
		option2Props := rtui.Props{
			"key":   "option2",
			"value": "value2",
			"label": "Label 2",
			"idx":   1,
		}

		option1 := optiongroup.NewOptionInstance(option1Props)
		option2 := optiongroup.NewOptionInstance(option2Props)

		// Add children
		group.AddChild(option1)
		group.AddChild(option2)

		// Verify children are tracked
		if children := group.Children(); len(children) != 2 {
			t.Errorf("Expected 2 children, got %d", len(children))
		}

		// Verify parent references are set via Instance Tree
		if option1.Parent() != group {
			t.Error("Expected option1.Parent() to return the group instance")
		}
		if option2.Parent() != group {
			t.Error("Expected option2.Parent() to return the group instance")
		}

		// Remove one child
		group.RemoveChild(option1)

		// Verify removal
		if children := group.Children(); len(children) != 1 {
			t.Errorf("Expected 1 child after removal, got %d", len(children))
		}

		if option1.Parent() != nil {
			t.Error("Expected option1.Parent() to be nil after removal")
		}

		if option2.Parent() != group {
			t.Error("Expected option2.Parent() to still be the group instance")
		}

		// Clear all children
		group.ClearChildren()

		if children := group.Children(); len(children) != 0 {
			t.Errorf("Expected 0 children after ClearChildren, got %d", len(children))
		}

		if option2.Parent() != nil {
			t.Error("Expected option2.Parent() to be nil after ClearChildren")
		}
	})
}
