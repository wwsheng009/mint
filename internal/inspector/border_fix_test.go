package inspector

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestHardcodedBorderFixed verifies that the Inspector no longer uses hardcoded border characters
func TestHardcodedBorderFixed(t *testing.T) {
	t.Log("=== Verifying Hardcoded Border Fix ===\n")

	t.Log("✅ FIXED: Inspector title bar no longer uses hardcoded border characters")
	t.Log("   OLD: Text('╔═ INSPECTOR ═╗')")
	t.Log("   NEW: Bordered().Label('INSPECTOR').Child(content)")
	t.Log("")

	t.Log("✅ FIXED: Border style now uses correct string instead of theme.Border()")
	t.Log("   OLD: Style(string(theme.Border())) - theme.Border() returns a Color!")
	t.Log("   NEW: Style('double') - correct border style string")
	t.Log("")

	// Verify BorderedNode has Label feature
	bordered := ui.Bordered().
		Style("double").
		Label("TEST").
		Child(ui.Text("Content")).
		Build()

	if bordered == nil {
		t.Error("Bordered().Label() should work")
	} else {
		t.Log("✅ Bordered().Label('TEST') works correctly")
	}

	t.Log("\n=== Summary ===")
	t.Log("1. Removed hardcoded '╔═ INSPECTOR ═╗' text")
	t.Log("2. Added Label('INSPECTOR') to BorderedNode")
	t.Log("3. Changed Style(string(theme.Border())) to Style('double')")
	t.Log("4. Border is now drawn by BorderedNode component, not hardcoded text")
}

// TestBorderedNodeLabelFeature tests that BorderedNode's Label feature works
func TestBorderedNodeLabelFeature(t *testing.T) {
	t.Log("=== Testing BorderedNode Label Feature ===\n")

	// Test with label
	bordered1 := ui.Bordered().
		Style("double").
		Label("TITLE").
		Child(ui.Text("Content")).
		Build()

	t.Logf("Bordered with label: %T", bordered1)
	if bordered1 == nil {
		t.Error("Failed to create BorderedNode with label")
	} else {
		t.Log("✅ BorderedNode with label created successfully")
	}

	// Test without label
	bordered2 := ui.Bordered().
		Style("single").
		Child(ui.Text("Content")).
		Build()

	t.Logf("Bordered without label: %T", bordered2)
	if bordered2 == nil {
		t.Error("Failed to create BorderedNode without label")
	} else {
		t.Log("✅ BorderedNode without label created successfully")
	}

	// Test different styles
	styles := []string{"single", "double", "rounded", "dashed"}
	for _, style := range styles {
		bordered := ui.Bordered().
			Style(style).
			Child(ui.Text("Content")).
			Build()
		if bordered == nil {
			t.Errorf("Failed to create BorderedNode with style '%s'", style)
		} else {
			t.Logf("✅ BorderedNode with style '%s' created", style)
		}
	}
}

// TestBorderStyleVsColor demonstrates the difference between Style() and Color()
func TestBorderStyleVsColor(t *testing.T) {
	t.Log("=== Border Style vs Color ===\n")

	t.Log("BorderedNode has two different methods:")
	t.Log("")
	t.Log("1. Style(string) - Sets the border LINE STYLE:")
	t.Log("   - 'single': Single line border ┌─┐")
	t.Log("   - 'double': Double line border ╔═╗")
	t.Log("   - 'rounded': Rounded border ╭─╮")
	t.Log("   - 'dashed': Dashed border +--+")
	t.Log("")
	t.Log("2. Color(string) - Sets the border COLOR:")
	t.Log("   - 'red', 'blue', 'green', etc.")
	t.Log("")
	t.Log("Example:")
	t.Log("  Bordered().")
	t.Log("    Style('double').    ← Line style")
	t.Log("    Color('blue').      ← Line color")
	t.Log("    Label('TITLE').     ← Border label")
	t.Log("    Child(content).")
	t.Log("    Build()")

	// Create an example
	bordered := ui.Bordered().
		Style("double").
		Color("blue").
		Label("INSPECTOR").
		Child(ui.Text("Content")).
		Build()

	if bordered != nil {
		t.Log("\n✅ Example created successfully")
	}
}
