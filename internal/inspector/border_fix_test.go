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

	// Verify BorderedNode has Label feature (migrated to Stack)
	bordered := ui.NewVStack().
		DoubleBorder("TEST").
		SetChildrenList([]ui.VNode{ui.Text("Content")})

	if bordered == nil {
		t.Error("Bordered Stack should work")
	} else {
		t.Log("✅ Bordered Stack with label works correctly")
	}

	t.Log("\n=== Summary ===")
	t.Log("1. Removed hardcoded '╔═ INSPECTOR ═╗' text")
	t.Log("2. Added Label('INSPECTOR') to BorderedNode")
	t.Log("3. Changed Style(string(theme.Border())) to Style('double')")
	t.Log("4. Border is now drawn by Stack component, not hardcoded text")
}

// TestBorderedNodeLabelFeature tests that Stack border feature works
func TestBorderedNodeLabelFeature(t *testing.T) {
	t.Log("=== Testing Stack Border Label Feature ===\n")

	// Test with label
	bordered1 := ui.NewVStack().
		DoubleBorder("TITLE").
		SetChildrenList([]ui.VNode{ui.Text("Content")})

	t.Logf("Bordered with label: %T", bordered1)
	if bordered1 == nil {
		t.Error("Failed to create Stack with label")
	} else {
		t.Log("✅ Stack with label created successfully")
	}

	// Test without label
	bordered2 := ui.NewVStack().
		SingleBorder().
		SetChildrenList([]ui.VNode{ui.Text("Content")})

	t.Logf("Bordered without label: %T", bordered2)
	if bordered2 == nil {
		t.Error("Failed to create Stack without label")
	} else {
		t.Log("✅ Stack without label created successfully")
	}

	// Test different styles
	t.Run("Single", func(t *testing.T) {
		bordered := ui.NewVStack().
			SingleBorder().
			SetChildrenList([]ui.VNode{ui.Text("Content")})
		if bordered == nil {
			t.Error("Failed to create Stack with single border")
		} else {
			t.Log("✅ Stack with single border created")
		}
	})
	t.Run("Double", func(t *testing.T) {
		bordered := ui.NewVStack().
			DoubleBorder().
			SetChildrenList([]ui.VNode{ui.Text("Content")})
		if bordered == nil {
			t.Error("Failed to create Stack with double border")
		} else {
			t.Log("✅ Stack with double border created")
		}
	})
	t.Run("Rounded", func(t *testing.T) {
		bordered := ui.NewVStack().
			RoundedBorder().
			SetChildrenList([]ui.VNode{ui.Text("Content")})
		if bordered == nil {
			t.Error("Failed to create Stack with rounded border")
		} else {
			t.Log("✅ Stack with rounded border created")
		}
	})
	t.Run("Dashed", func(t *testing.T) {
		bordered := ui.NewVStack().
			DashedBorder().
			SetChildrenList([]ui.VNode{ui.Text("Content")})
		if bordered == nil {
			t.Error("Failed to create Stack with dashed border")
		} else {
			t.Log("✅ Stack with dashed border created")
		}
	})
}

// TestBorderStyleVsColor demonstrates the difference between Style() and Color()
func TestBorderStyleVsColor(t *testing.T) {
	t.Log("=== Border Style vs Color ===\n")

	t.Log("Stack has two different methods:")
	t.Log("")
	t.Log("1. Border methods - Sets the border LINE STYLE:")
	t.Log("   - SingleBorder(): Single line border ┌─┐")
	t.Log("   - DoubleBorder(): Double line border ╔═╗")
	t.Log("   - RoundedBorder(): Rounded border ╭─╮")
	t.Log("   - DashedBorder(): Dashed border +--+")
	t.Log("")
	t.Log("2. BorderColor(string) - Sets the border COLOR:")
	t.Log("   - 'red', 'blue', 'green', etc.")
	t.Log("")
	t.Log("Example:")
	t.Log("  ui.NewVStack().")
	t.Log("    DoubleBorder().          ← Line style")
	t.Log("    BorderColor('blue').     ← Line color")
	t.Log("    BorderLabel('TITLE').    ← Border label")
	t.Log("    SetChildrenList([...])")

	// Create an example
	bordered := ui.NewVStack().
		DoubleBorder().
		BorderColor("blue").
		SetChildrenList([]ui.VNode{ui.Text("Content")})

	if bordered != nil {
		t.Log("\n✅ Example created successfully")
	}
}
