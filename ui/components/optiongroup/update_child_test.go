package optiongroup

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestUpdateChildInstances tests that child Option instances are updated
// when OptionGroup selection changes.
func TestUpdateChildInstances(t *testing.T) {
	options := []Option{
		{Value: "bj", Label: "Beijing"},
		{Value: "sh", Label: "Shanghai"},
		{Value: "gz", Label: "Guangzhou"},
	}

	// Create OptionGroup Instance
	props := rtui.Props{
		"key":   "city-group",
		"mode":  ModeSingle,
		"options": options,
	}
	inst := NewInstance(props)
	inst.SetProps(props)

	// Initial state: nothing selected
	if inst.selected != "" {
		t.Errorf("Expected no initial selection, got %q", inst.selected)
	}

	// Check initial child states
	for _, child := range inst.childInstances {
		if child.selected {
			t.Errorf("Option %q should not be selected initially", child.value)
		}
	}

	// Select "bj"
	inst.SelectOption("bj")

	// Verify group state
	if inst.selected != "bj" {
		t.Errorf("Expected selected = 'bj', got %q", inst.selected)
	}

	// Verify child states - Beijing should be selected, others not
	for _, child := range inst.childInstances {
		expectedSelected := (child.value == "bj")
		if child.selected != expectedSelected {
			t.Errorf("Option %q: expected selected=%v, got %v", child.value, expectedSelected, child.selected)
		}
	}

	// Select "sh"
	inst.SelectOption("sh")

	// Verify group state
	if inst.selected != "sh" {
		t.Errorf("Expected selected = 'sh', got %q", inst.selected)
	}

	// Verify child states - Shanghai should be selected, others not
	for _, child := range inst.childInstances {
		expectedSelected := (child.value == "sh")
		if child.selected != expectedSelected {
			t.Errorf("Option %q: expected selected=%v, got %v", child.value, expectedSelected, child.selected)
		}
	}
}

// TestMultiSelectUpdateChildInstances tests multi-select child updates.
func TestMultiSelectUpdateChildInstances(t *testing.T) {
	options := []Option{
		{Value: "dev", Label: "Development"},
		{Value: "design", Label: "Design"},
	}

	// Create multi-select OptionGroup Instance
	props := rtui.Props{
		"key":   "interests-group",
		"mode":  ModeMultiple,
		"options": options,
	}
	inst := NewInstance(props)
	inst.SetProps(props)

	// Select "dev"
	inst.SelectOption("dev")

	// Verify - dev selected, design not
	for _, child := range inst.childInstances {
		expectedSelected := (child.value == "dev")
		if child.selected != expectedSelected {
			t.Errorf("Option %q: expected selected=%v, got %v", child.value, expectedSelected, child.selected)
		}
	}

	// Select "design"
	inst.SelectOption("design")

	// Verify - both selected
	for _, child := range inst.childInstances {
		if !child.selected {
			t.Errorf("Option %q should be selected", child.value)
		}
	}

	// Deselect "dev"
	inst.DeselectOption("dev")

	// Verify - only design selected
	for _, child := range inst.childInstances {
		expectedSelected := (child.value == "design")
		if child.selected != expectedSelected {
			t.Errorf("Option %q: expected selected=%v, got %v", child.value, expectedSelected, child.selected)
		}
	}
}
